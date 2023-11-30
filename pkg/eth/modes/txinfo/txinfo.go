package txinfo

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"strings"
	"sync"

	"github.com/ZeljkoBenovic/tpser/pkg/conf"
	"github.com/ZeljkoBenovic/tpser/pkg/logger"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/olekukonko/tablewriter"
)

type TxInfo struct {
	ctx  context.Context
	log  logger.Logger
	eth  *ethclient.Client
	conf conf.Conf

	txHashes []*types.Transaction
	wg       *sync.WaitGroup
	sync.Mutex
}

func New(ctx context.Context, log logger.Logger, eth *ethclient.Client, conf conf.Conf) *TxInfo {
	return &TxInfo{
		ctx:  ctx,
		log:  log.Named("TxInfo"),
		eth:  eth,
		conf: conf,
		wg:   &sync.WaitGroup{},
	}
}

func (t *TxInfo) RunMode() error {
	t.log.Info("Fetching transaction information")

	for _, hash := range t.conf.TxHashes {
		t.wg.Add(1)
		hash := strings.TrimSpace(hash) // closure capture

		go func() {
			defer t.wg.Done()

			tx, _, err := t.eth.TransactionByHash(t.ctx, common.HexToHash(hash))
			if err != nil {
				t.log.Error("Could not fetch transaction by hash", "err", err.Error(), "tx_hash", hash)
				return
			}

			t.Lock()
			t.txHashes = append(t.txHashes, tx)
			t.Unlock()
		}()
	}

	t.wg.Wait()
	t.displayOutput()
	return nil
}

func (t *TxInfo) displayOutput() {
	var (
		base, pow    = big.NewInt(10), big.NewInt(18)
		totalCostWei = big.NewInt(0)
		oneEthInWei  = base.Exp(base, pow, nil)
		wei          = big.NewFloat(float64(oneEthInWei.Int64()))
		table        = tablewriter.NewWriter(os.Stdout)
		totalsTable  = tablewriter.NewWriter(os.Stdout)
	)

	table.SetHeader([]string{"TX_HASH", "TO", "GAS_LIMIT(WEI)", "GAS_PRICE (WEI)", "COST (ETH)", "VALUE"})

	totalsTable.SetHeader([]string{"TOTAL_TXS", "TOTAL_COST (ETH)"})
	totalsTable.SetAlignment(tablewriter.ALIGN_CENTER)

	for _, tx := range t.txHashes {
		cost := big.NewFloat(float64(tx.Cost().Int64()))
		ethCost := new(big.Float).Quo(cost, wei)

		totalCostWei = totalCostWei.Add(totalCostWei, tx.Cost())

		table.Append([]string{
			tx.Hash().Hex(),
			tx.To().Hex(),
			fmt.Sprintf("%d", tx.Gas()),
			tx.GasPrice().String(),
			ethCost.Text('f', 18),
			tx.Value().String(),
		})
	}

	ethTotalCost := new(big.Float).Quo(big.NewFloat(float64(totalCostWei.Int64())), wei)

	totalsTable.Append([]string{fmt.Sprintf("%d", len(t.txHashes)), ethTotalCost.Text('f', 18)})

	table.Render()
	totalsTable.Render()
}
