package longsender

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ZeljkoBenovic/tpser/pkg/prom"
	"golang.org/x/sync/errgroup"

	"github.com/ZeljkoBenovic/tpser/pkg/conf"
	"github.com/ZeljkoBenovic/tpser/pkg/eth/modes/getblocks"
	"github.com/ZeljkoBenovic/tpser/pkg/eth/transactions/txreceipts"
	"github.com/ZeljkoBenovic/tpser/pkg/eth/transactions/txsender"
	"github.com/ZeljkoBenovic/tpser/pkg/eth/transactions/txsigner"
	"github.com/ZeljkoBenovic/tpser/pkg/logger"
	"github.com/ethereum/go-ethereum/ethclient"
)

var ErrPrivKeyOrMnemonicNotProvided = errors.New("longsender requires mnemonic, public or private key")

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

	prom *prom.Prom
}

func New(ctx context.Context, log logger.Logger, eth *ethclient.Client, conf conf.Conf, prom *prom.Prom) *longsender {
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
		prom:      prom,
	}
}

type safeNonce struct {
	sync.RWMutex
	nonces map[int]*atomic.Uint64
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
	l.prom.SetTxSendInterval(float64(l.conf.TxSendInterval))
	l.prom.SetTxNumberPerInterval(float64(l.conf.TxPerSec))

	if l.conf.Mnemonic != "" {
		l.log.Info("Sending transactions using mnemonics")
		return l.sendTxFromMnemonics()
	} else if l.conf.PrivateKey != "" {
		l.log.Info("Sending transactions using private key")
		return l.sendTxUsingSingleWallet()
	} else if l.conf.Web3SignerURL != "" {
		l.log.Info("Sending transactions using web3signer service")
		l.signer.UseWeb3Signer()
		return l.sendTxUsingSingleWallet()
	} else {
		return ErrPrivKeyOrMnemonicNotProvided
	}
}

func (l *longsender) sendTxFromMnemonics() error {
	l.log.Info("Sending transactions using mnemonics", "tps", l.conf.TxPerSec, "duration_min", l.conf.TxSendTimeoutMin)
	var (
		firstBlock        uint64
		lastBlock         uint64
		err               error
		fetchPendingNonce bool
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
				if fetchPendingNonce {
					newNonce, err := signer.GetFreshNonce()
					if err != nil {
						l.log.Error("Could not fetch new nonce", "err", err.Error())
						l.prom.IncreaseTxErrorCount()
						continue
					}

					l.noncesMap.Store(ind, newNonce)
					fetchPendingNonce = false
					l.log.Debug("New nonce fetched", "nonce", l.noncesMap.Load(ind))
				}

				for range txNum {
					ind := ind
					signer := signer

					tx, err := signer.GetNextSignedTx(l.noncesMap.Load(ind))
					if err != nil {
						return err
					}

					errGr, _ := errgroup.WithContext(l.ctx)

					errGr.Go(func() error {
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
							return txErr
						}

						l.prom.ObserveTxRequestDuration(float64(time.Since(sendStart).Milliseconds()))

						l.receipts.StoreTxHash(hash)

						l.log.Info("Transaction sent",
							"hash", hash.String(),
							"from", signer.GetFromAddress(),
							"nonce", currentNonce,
						)

						return nil
					})

					if txErr := errGr.Wait(); txErr != nil {
						fetchPendingNonce = true
						l.prom.IncreaseTxErrorCount()
						continue
					}

					l.noncesMap.Increment(ind)
				}
			}
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

func (l *longsender) sendTxUsingSingleWallet() error {
	var (
		firstBlock        uint64
		lastBlock         uint64
		err               error
		fetchPendingNonce bool
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
			if fetchPendingNonce {
				newNonce, err := l.signer.GetFreshNonce()
				if err != nil {
					l.log.Error("Could not fetch new nonce", "err", err.Error())
					l.prom.IncreaseTxErrorCount()
					continue
				}

				l.nonce.Store(newNonce)
				fetchPendingNonce = false
				l.log.Debug("New nonce fetched", "nonce", l.nonce.Load())
			}

			for range txNum {
				tx, err := l.signer.GetNextSignedTx(l.nonce.Load())
				if err != nil {
					return err
				}

				errGr, _ := errgroup.WithContext(l.ctx)

				errGr.Go(func() error {
					sendStart := time.Now()

					hash, txErr := l.sender.SendSignedTransaction(tx)
					if txErr != nil {
						return txErr
					}

					l.prom.ObserveTxRequestDuration(float64(time.Since(sendStart).Milliseconds()))
					l.receipts.StoreTxHash(hash)

					l.log.Info("Transaction sent",
						"hash", hash.String(),
						"from", l.signer.GetFromAddress(),
						"nonce", l.nonce.Load(),
					)

					return nil
				})

				if txErr := errGr.Wait(); txErr != nil {
					l.log.Error("Transaction send error",
						"err", txErr,
						"from", l.signer.GetFromAddress(),
						"nonce", l.nonce.Load(),
						"hash", tx.Hash(),
					)
					fetchPendingNonce = true
					l.prom.IncreaseTxErrorCount()
					continue
				}

				l.nonce.Add(1)
			}
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
