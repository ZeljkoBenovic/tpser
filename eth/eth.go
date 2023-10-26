package eth

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/ZeljkoBenovic/tpser/conf"
	"github.com/ZeljkoBenovic/tpser/logger"
	"github.com/briandowns/spinner"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/olekukonko/tablewriter"
)

type Eth interface {
	GetBlockByNumberStats()
}

type eth struct {
	ctx context.Context

	conf      conf.Conf
	log       logger.Logger
	ethClient *ethclient.Client
}

func New(conf conf.Conf, log logger.Logger, ctx context.Context) (Eth, error) {
	e, err := ethclient.Dial(conf.JsonRPC)
	if err != nil {
		log.Error("Could not dial json-rpc", "json-rpc", conf.JsonRPC)
		return nil, err
	}

	return &eth{
		ethClient: e,
		log:       log,
		ctx:       ctx,
		conf:      conf,
	}, nil
}

func (e *eth) GetBlockByNumberStats() {
	s := spinner.New(spinner.CharSets[35], 500*time.Millisecond)
	s.Start()

	blockMap := make([]blockInfo, (e.conf.Blocks.End-e.conf.Blocks.Start)+1)
	wg := sync.WaitGroup{}
	limiter := make(chan struct{}, runtime.NumCPU()*10)
	startingBlock := e.conf.Blocks.Start - 1

	for i := int64(0); i <= e.conf.Blocks.End-e.conf.Blocks.Start; i++ {
		limiter <- struct{}{}
		startingBlock++
		wg.Add(1)

		go e.getBlockByNumberInfo(startingBlock, &blockMap[i], &wg, limiter)
	}

	wg.Wait()

	s.Stop()
	e.outputStats(blockMap)

}

func (e *eth) getBlockByNumberInfo(blockNumber int64, dest *blockInfo, wg *sync.WaitGroup, limiter <-chan struct{}) {
	defer func() {
		<-limiter
		wg.Done()
	}()

	block, err := e.ethClient.BlockByNumber(e.ctx, big.NewInt(blockNumber))
	if err != nil {
		e.log.Error("Could not fetch block", "number", blockNumber, "err", err.Error())
	}

	dest.number = block.NumberU64()
	dest.time = block.Time()
	dest.gasLimit = block.GasLimit()
	dest.gasUsed = block.GasUsed()
	dest.hash = block.Hash().String()
	dest.transactionNum = block.Transactions().Len()
}

func (e *eth) outputStats(blocks []blockInfo) {
	totalTxs := uint64(0)
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"TIME", "NUMBER", "TXS", "GAS_LIMIT", "GAS_USED"})

	for _, block := range blocks {
		blTime := time.Unix(int64(block.time), 0)

		table.Append([]string{
			blTime.Format(time.DateTime),
			fmt.Sprintf("%d", block.number),
			fmt.Sprintf("%d", block.transactionNum),
			fmt.Sprintf("%d", block.gasLimit),
			fmt.Sprintf("%.2f%%", float64(block.gasUsed)/float64(block.gasLimit)*100),
		})

		totalTxs += uint64(block.transactionNum)
	}

	e.calculateTPS(totalTxs, blocks, table)

	table.Render()
}

func (e *eth) calculateTPS(totalTxs uint64, blocks []blockInfo, table *tablewriter.Table) {
	timeStart := time.Unix(int64(blocks[0].time), 0)
	timeFinish := time.Unix(int64(blocks[len(blocks)-1].time), 0)
	totalTimeToComplete := timeFinish.Sub(timeStart)

	tps := float64(totalTxs) / totalTimeToComplete.Seconds()

	table.SetFooter([]string{fmt.Sprintf("DURATION: %.2f s", totalTimeToComplete.Seconds()), "TOTAL TX", fmt.Sprintf("%d", totalTxs), "TPS", fmt.Sprintf("%.2f", tps)})

	table.SetFooterColor(tablewriter.Colors{},
		tablewriter.Colors{tablewriter.Bold}, tablewriter.Colors{tablewriter.BgBlueColor},
		tablewriter.Colors{tablewriter.Bold}, tablewriter.Colors{tablewriter.BgGreenColor},
	)
}
