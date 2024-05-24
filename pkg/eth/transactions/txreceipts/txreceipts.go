package txreceipts

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/ZeljkoBenovic/tpser/pkg/conf"
	"github.com/ZeljkoBenovic/tpser/pkg/logger"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

type TxReceipts struct {
	ctx  context.Context
	eth  *ethclient.Client
	log  logger.Logger
	conf conf.Conf

	wg      *sync.WaitGroup
	limiter chan struct{}

	safeReceipts safeReceipts
}

type safeReceipts struct {
	sync.Mutex
	receipts  map[common.Hash]*types.Receipt
	confirmed uint64
}

func New(ctx context.Context, log logger.Logger, eth *ethclient.Client, cfg conf.Conf) *TxReceipts {
	return &TxReceipts{
		ctx:     ctx,
		eth:     eth,
		log:     log.Named("txreceipts"),
		conf:    cfg,
		wg:      &sync.WaitGroup{},
		limiter: make(chan struct{}, runtime.NumCPU()*100),
		safeReceipts: safeReceipts{
			receipts: make(map[common.Hash]*types.Receipt, 0),
		},
	}
}

func (r *TxReceipts) StoreTxHash(hash common.Hash) {
	r.safeReceipts.storeTxHash(hash)
}

func (r *TxReceipts) ConfirmTransactions() {
	var (
		txHashes = make([]common.Hash, 0)
	)

	ctx, cancel := context.WithTimeout(r.ctx, time.Minute*time.Duration(r.conf.WaitForConfirmTimeout))
	defer cancel()

	// extract tx hashes to prevent data race
	for hash := range r.safeReceipts.receipts {
		txHashes = append(txHashes, hash)
	}

	for _, hash := range txHashes {
		r.limiter <- struct{}{}
		r.wg.Add(1)

		go r.tryFetchReceiptsWithDeadline(ctx, hash)
	}

	r.wg.Wait()

	if r.safeReceipts.confirmed == uint64(len(txHashes)) {
		r.log.Info("All transactions successfully confirmed", "sent_tx", len(txHashes), "receipts", r.safeReceipts.confirmed)
	} else {
		r.log.Error("Transactions not confirmed", "sent_tx", len(txHashes), "receipts", r.safeReceipts.confirmed)
	}

}

func (r *TxReceipts) tryFetchReceiptsWithDeadline(ctx context.Context, hash common.Hash) {
	defer func() {
		r.wg.Done()
		<-r.limiter
	}()

	for {
		select {
		case <-ctx.Done():
			r.log.Error("Could not verify transaction",
				"verification_duration", r.conf.WaitForConfirmTimeout,
				"hash", hash.String(),
			)

			return

		default:
			receipt, _ := r.eth.TransactionReceipt(ctx, hash)
			if receipt != nil {
				if err := r.safeReceipts.storeTxReceipt(receipt); err != nil {
					r.log.Error("Could not store receipt", "err", err.Error())
				}

				r.log.Debug("Receipt stored", "hash", hash)

				return
			}

			r.log.Debug("Transaction receipt not available, retrying", "hash", hash)
			time.Sleep(time.Second * 5)
		}
	}
}

func (s *safeReceipts) storeTxHash(hash common.Hash) {
	s.Lock()
	defer s.Unlock()

	s.receipts[hash] = nil
}

func (s *safeReceipts) storeTxReceipt(receipt *types.Receipt) error {
	s.Lock()
	defer s.Unlock()

	_, ok := s.receipts[receipt.TxHash]
	if !ok {
		return fmt.Errorf("tx hash for the receipt not found: %s", receipt.TxHash)
	}

	s.receipts[receipt.TxHash] = receipt
	s.confirmed++

	return nil
}
