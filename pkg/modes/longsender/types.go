package longsender

import (
	"context"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type Config struct {
	TxSendTimeout time.Duration // timeout for when to stop sending transactions
	ToSendAddress string        // address to send to in EOA transfers
	TPS           uint64        // number of transactions to send, per second
}

type SingleConfig struct {
	Config

	PrivateKey string // private key for EOA
}

type MultipleConfig struct {
	Config

	Mnemonic    string // mnemonic for HD wallet
	NumAccounts uint64 // number of accounts to generate from the mnemonic
}

// client is the interface that wraps the basic ethclient methods.
// NOTE: This API directly mirrors the ethclient API, which is not really optimal.
// Consider refactoring this to a more friendly API, and wrap ethclient instead
type client interface {
	// BlockNumber returns the latest block number
	BlockNumber(ctx context.Context) (uint64, error)

	// SendTransaction sends a transaction to the network
	SendTransaction(ctx context.Context, tx *types.Transaction) error

	// ChainID returns the chain ID
	ChainID(ctx context.Context) (*big.Int, error)

	// PendingNonceAt returns the pending nonce for an account
	PendingNonceAt(ctx context.Context, account common.Address) (uint64, error)

	// SuggestGasPrice returns the suggested gas price from the network
	SuggestGasPrice(ctx context.Context) (*big.Int, error)
}
