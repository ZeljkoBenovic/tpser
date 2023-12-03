package longsender

import (
	"context"
	"crypto/ecdsa"
	"fmt"

	"github.com/ZeljkoBenovic/tpser/pkg/logger"
	"github.com/ethereum/go-ethereum/crypto"
)

type SingleMode struct {
	*base

	config SingleConfig
}

func NewSingle(
	log logger.Logger,
	client client,
	config SingleConfig,
) *SingleMode {
	return &SingleMode{
		base:   newBase(log, client, config.Config),
		config: config,
	}
}

func (s *SingleMode) Run(ctx context.Context) error {
	// Set up the context to stop after the timeout
	runCtx, cancel := context.WithTimeout(ctx, s.config.TxSendTimeout)
	defer cancel()

	// Initialize the account from the private key
	pk, err := crypto.HexToECDSA(s.config.PrivateKey)
	if err != nil {
		return fmt.Errorf("could not setup private key: %w", err)
	}

	return s.sendTransactions(runCtx, []*ecdsa.PrivateKey{pk})
}
