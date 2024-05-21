package longsender

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ZeljkoBenovic/tpser/pkg/conf"
	"github.com/ZeljkoBenovic/tpser/pkg/eth/modes/getblocks"
	"github.com/ZeljkoBenovic/tpser/pkg/eth/tools/txreceipts"
	"github.com/ZeljkoBenovic/tpser/pkg/eth/tools/txsender"
	"github.com/ZeljkoBenovic/tpser/pkg/eth/tools/txsigner"
	"github.com/ZeljkoBenovic/tpser/pkg/logger"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var ErrPrivKeyOrMnemonicNotProvided = errors.New("longsender requires mnemonic or private key")

type longsender struct {
	ctx    context.Context
	cancel context.CancelFunc
	log    logger.Logger
	eth    *ethclient.Client
	conf   conf.Conf

	signer    *txsigner.TxSigner
	sender    *txsender.TxSender
	getblocks *getblocks.GetBlocks
	receipts  *txreceipts.TxReceipts

	wg        sync.WaitGroup
	nonce     *atomic.Uint64
	noncesMap safeNonce

	txRequestDuration txRequestDuration
}

func New(ctx context.Context, log logger.Logger, eth *ethclient.Client, conf conf.Conf) *longsender {
	newCtx, cancel := context.WithCancel(ctx)

	// enable indefinite runs
	if conf.TxSendTimeoutMin > 0 {
		newCtx, cancel = context.WithTimeout(ctx, time.Duration(conf.TxSendTimeoutMin)*time.Minute)
	}

	return &longsender{
		wg:     sync.WaitGroup{},
		ctx:    newCtx,
		cancel: cancel,
		log:    log,
		eth:    eth,
		conf:   conf,
		nonce:  &atomic.Uint64{},
		noncesMap: safeNonce{
			nonces: map[int]*atomic.Uint64{},
		},
		signer:    txsigner.New(ctx, log, eth, conf),
		sender:    txsender.New(ctx, log, eth),
		getblocks: getblocks.New(ctx, log, eth, conf),
		receipts:  txreceipts.New(ctx, log, eth, conf),

		txRequestDuration: txRequestDuration{
			mut: sync.Mutex{},
			duration: promauto.NewHistogram(prometheus.HistogramOpts{
				Namespace: "tpser",
				Name:      "tpser_tx_duration",
				Help:      "Transaction sending request duration",
			}),
		},
	}
}

type safeNonce struct {
	sync.RWMutex
	nonces map[int]*atomic.Uint64
}

type txRequestDuration struct {
	mut      sync.Mutex
	duration prometheus.Histogram
}

func (t *txRequestDuration) Observe(observable float64) {
	t.mut.Lock()
	t.duration.Observe(observable)
	t.mut.Unlock()
}

func (s *safeNonce) Load(index int) uint64 {
	s.RLock()
	defer s.RUnlock()

	return s.nonces[index].Load()
}

func (s *safeNonce) Increment(index int) {
	s.Lock()
	defer s.Unlock()

	s.nonces[index].Add(1)
}

func (s *safeNonce) Store(index int, value uint64) {
	s.Lock()
	defer s.Unlock()

	s.nonces[index].Store(value)
}

func (l *longsender) RunMode() error {
	if l.conf.Mnemonic != "" {
		return l.sendTxFromMnemonics()
	} else if l.conf.PrivateKey != "" {
		return l.sendTxWithPrivateKey()
	} else {
		return ErrPrivKeyOrMnemonicNotProvided
	}
}

func (l *longsender) sendTxFromMnemonics() error {
	l.log.Info("Sending transactions using mnemonics", "tps", l.conf.TxPerSec, "duration_min", l.conf.TxSendTimeoutMin)
	var (
		firstBlock uint64
		lastBlock  uint64
		err        error
		// split number of transactions evenly
		txNum = make([]struct{}, l.conf.TxPerSec/int64(l.conf.TotalAccounts))
		tick  = time.Tick(time.Second * time.Duration(l.conf.TxSendInterval))
	)

	if l.conf.IncludeTPSReport {
		firstBlock, err = l.eth.BlockNumber(l.ctx)
		if err != nil {
			return err
		}
	}

	signers, err := l.initMnemonicAccounts()
	if err != nil {
		return err
	}

	for {
		select {
		case <-tick:
			// each signer should send its own batch
			for ind, signer := range signers {
				for range txNum {
					l.wg.Add(1)
					ind := ind

					tx, err := signer.GetNextSignedTx(l.noncesMap.Load(ind))
					if err != nil {
						return err
					}

					go func(
						tx *types.Transaction,
						signer *txsigner.TxSigner,
					) {
						defer l.wg.Done()

						currentNonce := l.noncesMap.Load(ind)

						sendStart := time.Now()

						hash, txErr := l.sender.SendSignedTransaction(tx)
						if txErr != nil {
							l.log.Error("Transaction send error",
								"err", txErr,
								"hash", tx.Hash(),
								"from", signer.GetFromAddress(),
								"nonce", currentNonce,
							)

							return
						}

						sendEnd := time.Since(sendStart)

						l.txRequestDuration.Observe(float64(sendEnd.Milliseconds()))

						l.receipts.StoreTxHash(hash)

						l.log.Debug("Transaction sent",
							"hash", hash.String(),
							"from", signer.GetFromAddress(),
							"nonce", currentNonce,
						)

					}(tx, signer)

					l.noncesMap.Increment(ind)
				}
			}

			l.wg.Wait()
			l.log.Debug("Transaction batch sent")
		case <-l.ctx.Done():
			if l.conf.IncludeTPSReport {
				lastBlock, err = l.eth.BlockNumber(l.ctx)
				if err != nil {
					return err
				}

				l.log.Info("Transaction send timeout reached, generating report")

				return l.getblocks.GetBlocksByNumbers(int64(firstBlock), int64(lastBlock))
			} else if l.conf.WaitForConfirm {
				l.log.Info("Waiting for transactions verification...")

				l.receipts.ConfirmTransactions()
			} else {
				l.log.Info("Transaction send timeout reached, stopping send", "timeout_min", l.conf.TxSendTimeoutMin)
				return nil
			}

		}
	}
}

func (l *longsender) initMnemonicAccounts() ([]*txsigner.TxSigner, error) {
	var (
		signers  = make([]*txsigner.TxSigner, 0)
		nonces   = map[int]*atomic.Uint64{}
		numOfAcc = make([]struct{}, l.conf.TotalAccounts)
	)

	for range numOfAcc {
		signers = append(signers, txsigner.New(l.ctx, l.log, l.eth, l.conf))
	}

	for ind, signer := range signers {
		if err := signer.SetPrivateKey(txsigner.WithNumberOfAccounts(ind)); err != nil {
			l.log.Error("Could not set private key", "err", err.Error())
			return nil, err
		}

		if err := signer.SetToAddress(l.conf.ToAddress); err != nil {
			l.log.Error("Could not set to address", "err", err.Error())
			return nil, err
		}

		nonces[ind] = &atomic.Uint64{}
		nonces[ind].Store(signer.GetNonce())

		l.noncesMap.nonces = nonces
	}

	return signers, nil
}

func (l *longsender) sendTxWithPrivateKey() error {
	var (
		firstBlock uint64
		lastBlock  uint64
		err        error
	)
	l.log.Info("Sending transactions using private key", "tps", l.conf.TxPerSec, "duration_min", l.conf.TxSendTimeoutMin)

	if err := l.initSender(); err != nil {
		return err
	}

	txNum := make([]struct{}, l.conf.TxPerSec)
	l.nonce.Store(l.signer.GetNonce())

	tick := time.Tick(time.Second * time.Duration(l.conf.TxSendInterval))

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

				tx, err := l.signer.GetNextSignedTx(l.nonce.Load())
				if err != nil {
					return err
				}

				go func(tx *types.Transaction) {
					defer l.wg.Done()

					sendStart := time.Now()

					hash, txErr := l.sender.SendSignedTransaction(tx)
					if txErr != nil {
						l.log.Error("Transaction send error",
							"err", txErr,
							"from", l.signer.GetFromAddress(),
							"nonce", l.nonce.Load(),
							"hash", tx.Hash(),
						)
						return
					}

					sendEnd := time.Since(sendStart)

					l.txRequestDuration.Observe(float64(sendEnd.Milliseconds()))

					l.receipts.StoreTxHash(hash)

					l.log.Debug("Transaction sent",
						"hash", hash.String(),
						"from", l.signer.GetFromAddress(),
						"nonce", l.nonce.Load(),
					)
				}(tx)

				l.nonce.Add(1)
			}

			l.wg.Wait()
			l.log.Debug("Transaction batch sent")
		case <-l.ctx.Done():
			if l.conf.IncludeTPSReport {
				lastBlock, err = l.eth.BlockNumber(l.ctx)
				if err != nil {
					return err
				}

				l.log.Info("Transaction send timeout reached, generating report")

				return l.getblocks.GetBlocksByNumbers(int64(firstBlock), int64(lastBlock))
			} else if l.conf.WaitForConfirm {
				l.log.Info("Waiting for transactions verification...")

				l.receipts.ConfirmTransactions()
				return nil
			} else {
				l.log.Info("Transaction send timeout reached, stopping send", "timeout_min", l.conf.TxSendTimeoutMin)
				return nil
			}
		}
	}
}

func (l *longsender) initSender() error {
	if err := l.signer.SetPrivateKey(); err != nil {
		return err
	}

	if err := l.signer.SetToAddress(l.conf.ToAddress); err != nil {
		return err
	}

	return nil
}
