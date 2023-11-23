package modes

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"time"

	"github.com/ZeljkoBenovic/tpser/pkg/eth/types"
	"github.com/briandowns/spinner"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/olekukonko/tablewriter"
	"golang.org/x/sync/errgroup"
)

type getBlocks struct {
	eth  *ethclient.Client
	conf GetBlocksConfig
}

// newGetBlocksMode creates a new get blocks mode
func newGetBlocksMode(
	eth *ethclient.Client,
	conf GetBlocksConfig,
) *getBlocks {
	return &getBlocks{
		eth:  eth,
		conf: conf,
	}
}

func (g *getBlocks) Run(ctx context.Context) error {
	var (
		startBlock = g.conf.Start
		endBlock   = g.conf.End
	)

	if g.conf.Tail != 0 {
		latestBlock, err := g.eth.BlockNumber(ctx)
		if err != nil {
			return fmt.Errorf("could not get latest block: %w", err)
		}

		endBlock = latestBlock
		startBlock = endBlock - g.conf.Tail
	}

	return g.printBlockRange(ctx, startBlock, endBlock)
}

// printBlockRange outputs the block info for the specified range
func (g *getBlocks) printBlockRange(
	ctx context.Context,
	startBlock,
	endBlock uint64,
) error {
	s := spinner.New(spinner.CharSets[35], 500*time.Millisecond)

	s.Start()
	defer s.Stop()

	group, groupCtx := errgroup.WithContext(ctx)

	blocks := make([]*types.BlockInfo, endBlock-startBlock)
	for block := startBlock; block <= endBlock; block++ {
		block := block

		group.Go(func() error {
			// Fetch the block info
			info, err := g.getBlockInfo(groupCtx, block)
			if err != nil {
				return err
			}

			// Save the block info
			blocks[block] = info

			return nil
		})
	}

	if err := group.Wait(); err != nil {
		return err
	}

	// Output the stats
	outputStats(blocks)

	return nil
}

// getBlockInfo fetches the block info from the chain
func (g *getBlocks) getBlockInfo(ctx context.Context, blockNumber uint64) (*types.BlockInfo, error) {
	block, err := g.eth.BlockByNumber(ctx, big.NewInt(0).SetUint64(blockNumber))
	if err != nil {
		return nil, fmt.Errorf("could not fetch block %d, %w", blockNumber, err)
	}

	return &types.BlockInfo{
		TransactionNum: block.Transactions().Len(),
		GasLimit:       block.GasLimit(),
		GasUsed:        block.GasUsed(),
		Hash:           block.Hash().String(),
		Number:         block.NumberU64(),
		Time:           block.Time(),
	}, nil
}

// outputStats outputs the formatted block stats
func outputStats(blocks []*types.BlockInfo) {
	totalTxs := uint64(0)
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(
		[]string{
			"TIME",
			"NUMBER",
			"TXS",
			"GAS_LIMIT",
			"GAS_USED",
		})

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

	printTPS(totalTxs, blocks, table)

	table.Render()
}

// printTPS prints the TPS information for the block range
func printTPS(totalTxs uint64, blocks []*types.BlockInfo, table *tablewriter.Table) {
	timeStart := time.Unix(int64(blocks[0].Time), 0)
	timeFinish := time.Unix(int64(blocks[len(blocks)-1].Time), 0)
	totalTimeToComplete := timeFinish.Sub(timeStart)

	tps := float64(totalTxs) / totalTimeToComplete.Seconds()

	table.SetFooter(
		[]string{
			fmt.Sprintf("DURATION: %.2f s", totalTimeToComplete.Seconds()),
			"TOTAL TX",
			fmt.Sprintf("%d", totalTxs),
			"TPS",
			fmt.Sprintf("%.2f", tps),
		},
	)

	table.SetFooterColor(tablewriter.Colors{},
		tablewriter.Colors{tablewriter.Bold}, tablewriter.Colors{tablewriter.BgBlueColor},
		tablewriter.Colors{tablewriter.Bold}, tablewriter.Colors{tablewriter.BgGreenColor},
	)
}
