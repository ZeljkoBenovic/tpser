package getblocks

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/ZeljkoBenovic/tpser/pkg/conf"
	"github.com/ZeljkoBenovic/tpser/pkg/eth/types"
	"github.com/ZeljkoBenovic/tpser/pkg/logger"
	"github.com/briandowns/spinner"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/olekukonko/tablewriter"
)

type GetBlocks struct {
	ctx  context.Context
	log  logger.Logger
	eth  *ethclient.Client
	conf conf.Conf

	wg      sync.WaitGroup
	blocks  []types.BlockInfo
	limiter chan struct{}
}

func New(ctx context.Context, log logger.Logger, eth *ethclient.Client, conf conf.Conf) *GetBlocks {
	return &GetBlocks{
		ctx:     ctx,
		log:     log.Named("getblocks"),
		eth:     eth,
		conf:    conf,
		wg:      sync.WaitGroup{},
		blocks:  make([]types.BlockInfo, 0),
		limiter: make(chan struct{}, runtime.NumCPU()*50),
	}
}

func (g *GetBlocks) RunMode() error {
	return g.GetBlocksByNumbers(g.conf.Blocks.Start, g.conf.Blocks.End)
}

func (g *GetBlocks) GetBlocksByNumbers(startBlock, endBlock int64) error {
	s := spinner.New(spinner.CharSets[35], 500*time.Millisecond)
	s.Start()

	for i := startBlock; i <= endBlock; i++ {
		g.limiter <- struct{}{}
		g.wg.Add(1)

		go g.getBlockByNumberInfo(i)
	}

	g.wg.Wait()
	s.Stop()

	g.sortBlocks()
	g.outputStats()

	return nil
}

func (g *GetBlocks) getBlockByNumberInfo(blockNumber int64) {
	defer func() {
		<-g.limiter
		g.wg.Done()
	}()

	block, err := g.eth.BlockByNumber(g.ctx, big.NewInt(blockNumber))
	if err != nil {
		g.log.Error("Could not fetch block", "number", blockNumber, "err", err.Error())
	}

	g.blocks = append(g.blocks, types.BlockInfo{
		TransactionNum: block.Transactions().Len(),
		GasLimit:       block.GasLimit(),
		GasUsed:        block.GasUsed(),
		Hash:           block.Hash().String(),
		Number:         block.NumberU64(),
		Time:           block.Time(),
	})
}

func (g *GetBlocks) outputStats() {
	totalTxs := uint64(0)
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"TIME", "NUMBER", "TXS", "GAS_LIMIT", "GAS_USED"})

	for _, block := range g.blocks {
		blTime := time.Unix(int64(block.Time), 0)

		table.Append([]string{
			blTime.Format(time.DateTime),
			fmt.Sprintf("%d", block.Number),
			fmt.Sprintf("%d", block.TransactionNum),
			fmt.Sprintf("%d", block.GasLimit),
			fmt.Sprintf("%.2f%%", float64(block.GasUsed)/float64(block.GasLimit)*100),
		})

		totalTxs += uint64(block.TransactionNum)
	}

	g.calculateTPS(totalTxs, g.blocks, table)

	table.Render()
}

// simple bubble sort O(n^2)
func (g *GetBlocks) sortBlocks() {
	var done = false

	for !done {
		done = true

		for i := 0; i < len(g.blocks)-1; i++ {
			if g.blocks[i].Number > g.blocks[i+1].Number {
				first := &g.blocks[i]
				second := &g.blocks[i+1]
				*first, *second = g.blocks[i+1], g.blocks[i]

				done = false
			}
		}
	}
}

func (g *GetBlocks) calculateTPS(totalTxs uint64, blocks []types.BlockInfo, table *tablewriter.Table) {
	timeStart := time.Unix(int64(blocks[0].Time), 0)
	timeFinish := time.Unix(int64(blocks[len(blocks)-1].Time), 0)
	totalTimeToComplete := timeFinish.Sub(timeStart)

	tps := float64(totalTxs) / totalTimeToComplete.Seconds()

	table.SetFooter([]string{fmt.Sprintf("DURATION: %.2f s", totalTimeToComplete.Seconds()), "TOTAL TX", fmt.Sprintf("%d", totalTxs), "TPS", fmt.Sprintf("%.2f", tps)})

	table.SetFooterColor(tablewriter.Colors{},
		tablewriter.Colors{tablewriter.Bold}, tablewriter.Colors{tablewriter.BgBlueColor},
		tablewriter.Colors{tablewriter.Bold}, tablewriter.Colors{tablewriter.BgGreenColor},
	)
}
