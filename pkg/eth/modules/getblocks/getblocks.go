package getblocks

import (
	"context"
	"errors"
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

type RunConfig struct {
	StartBlock int64
	EndBlock   int64
}

var (
	ErrConfigTypeNotSupported = errors.New("specified config type not supported")
)

type getblocks struct {
	ctx  context.Context
	log  logger.Logger
	eth  *ethclient.Client
	conf conf.Conf
}

func New(ctx context.Context, log logger.Logger, eth *ethclient.Client, conf conf.Conf) *getblocks {
	return &getblocks{
		ctx:  ctx,
		log:  log.Named("getblocks"),
		eth:  eth,
		conf: conf,
	}
}

func (g *getblocks) RunMode() error {
	fmt.Printf("START: %d END: %d", g.conf.Blocks.Start, g.conf.Blocks.End)
	return g.GetBlocksByNumbers(g.conf.Blocks.Start, g.conf.Blocks.End)
}

func (g *getblocks) GetBlocksByNumbers(startBlock, endBlock int64) error {
	s := spinner.New(spinner.CharSets[35], 500*time.Millisecond)
	s.Start()

	blockMap := make([]types.BlockInfo, (endBlock-startBlock)+1)
	wg := sync.WaitGroup{}
	limiter := make(chan struct{}, runtime.NumCPU()*10)
	startingBlock := startBlock - 1

	for i := int64(0); i <= endBlock-startingBlock; i++ {
		limiter <- struct{}{}
		startingBlock++
		wg.Add(1)

		go g.getBlockByNumberInfo(startingBlock, &blockMap[i], &wg, limiter)
	}

	wg.Wait()
	s.Stop()

	g.outputStats(blockMap)

	return nil
}

func (g *getblocks) getBlockByNumberInfo(blockNumber int64, dest *types.BlockInfo, wg *sync.WaitGroup, limiter <-chan struct{}) {
	defer func() {
		<-limiter
		wg.Done()
	}()

	block, err := g.eth.BlockByNumber(g.ctx, big.NewInt(blockNumber))
	if err != nil {
		g.log.Error("Could not fetch block", "number", blockNumber, "err", err.Error())
	}

	dest.Number = block.NumberU64()
	dest.Time = block.Time()
	dest.GasLimit = block.GasLimit()
	dest.GasUsed = block.GasUsed()
	dest.Hash = block.Hash().String()
	dest.TransactionNum = block.Transactions().Len()
}

func (g *getblocks) outputStats(blocks []types.BlockInfo) {
	totalTxs := uint64(0)
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"TIME", "NUMBER", "TXS", "GAS_LIMIT", "GAS_USED"})

	for _, block := range blocks {
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

	g.calculateTPS(totalTxs, blocks, table)

	table.Render()
}

func (g *getblocks) calculateTPS(totalTxs uint64, blocks []types.BlockInfo, table *tablewriter.Table) {
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
