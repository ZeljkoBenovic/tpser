package fetcher

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/core/types"
)

// Config is the configuration for the get blocks mode
type Config struct {
	Start uint64 // Start block number
	End   uint64 // End block number
	Tail  uint64 // Number of blocks to fetch from the latest block
}

// client is the interface that wraps the basic ethclient methods
type client interface {
	// BlockNumber returns the latest block number
	BlockNumber(ctx context.Context) (uint64, error)

	// BlockByNumber returns the block by the specified number
	BlockByNumber(ctx context.Context, number *big.Int) (*types.Block, error)
}
