package modes

import (
	"context"

	"github.com/ethereum/go-ethereum/ethclient"
)

// Mode is a mode of operation
type Mode interface {
	Run(ctx context.Context) error
}

// NewGetBlocksMode creates a new get blocks mode
func NewGetBlocksMode(
	client *ethclient.Client,
	cfg GetBlocksConfig,
) Mode {
	return newGetBlocksMode(client, cfg)
}
