package longsender

import (
	"context"
	"sync"
	"time"

	"github.com/ZeljkoBenovic/tpser/pkg/conf"
	"github.com/ZeljkoBenovic/tpser/pkg/eth/modes/getblocks"
	"github.com/ZeljkoBenovic/tpser/pkg/eth/tools/txsender"
	"github.com/ZeljkoBenovic/tpser/pkg/eth/tools/txsigner"
	"github.com/ZeljkoBenovic/tpser/pkg/logger"
	"github.com/ethereum/go-ethereum/ethclient"
)

type longsender struct {
	ctx  context.Context
	log  logger.Logger
	eth  *ethclient.Client
	conf conf.Conf

	signer    *txsigner.TxSigner
	sender    *txsender.TxSender
	getblocks *getblocks.GetBlocks

	wg  sync.WaitGroup
	mux sync.Mutex
}

func New(ctx context.Context, log logger.Logger, eth *ethclient.Client, conf conf.Conf) *longsender {
	return &longsender{
		wg:        sync.WaitGroup{},
		ctx:       ctx,
		log:       log,
		eth:       eth,
		conf:      conf,
		signer:    txsigner.New(ctx, log, eth),
		sender:    txsender.New(ctx, log, eth),
		getblocks: getblocks.New(ctx, log, eth, conf),
	}
}

func (l *longsender) RunMode() error {
	var (
		firstBlock uint64
		lastBlock  uint64
		err        error
	)
	l.log.Info("Sending transactions", "tps", l.conf.TxPerSec, "duration_min", l.conf.TxSendTimeoutMin)

	if err := l.initSender(); err != nil {
		return err
	}

	txNum := make([]struct{}, l.conf.TxPerSec)
	nonce := l.signer.GetNonce()

	tick := time.Tick(time.Second)
	timeout := time.After(time.Minute * time.Duration(l.conf.TxSendTimeoutMin))

	if l.conf.IncludeTPSReport {
		firstBlock, err = l.eth.BlockNumber(l.ctx)
		if err != nil {
			return err
		}
	}

	for {
		select {
		case <-tick:
			for range txNum {
				l.wg.Add(1)

				tx, err := l.signer.GetNextSignedTx(nonce)
				if err != nil {
					return err
				}

				go func() {
					defer l.wg.Done()
					hash, txErr := l.sender.SendSignedTransaction(tx)
					if txErr != nil {
						l.log.Error("Transaction send error", "err", txErr)
						return
					}

					l.log.Debug("Transaction sent", "hash", hash.String())
				}()

				nonce++
			}

			l.wg.Wait()
			l.log.Debug("Transaction batch sent")
		case <-timeout:
			if l.conf.IncludeTPSReport {
				lastBlock, err = l.eth.BlockNumber(l.ctx)
				if err != nil {
					return err
				}

				l.log.Info("Transaction send timeout reached, generating report")

				return l.getblocks.GetBlocksByNumbers(int64(firstBlock), int64(lastBlock))
			} else {
				l.log.Info("Transaction send timeout reached, stopping send", "timeout_min", l.conf.TxSendTimeoutMin)
				return nil
			}

		}
	}
}

func (l *longsender) initSender() error {
	if err := l.signer.SetPrivateKey(l.conf.PrivateKey); err != nil {
		return err
	}

	if err := l.signer.SetToAddress(l.conf.ToAddress); err != nil {
		return err
	}

	return nil
}
