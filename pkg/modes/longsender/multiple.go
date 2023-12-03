package longsender

import (
	"context"
	"fmt"

	localCrypto "github.com/ZeljkoBenovic/tpser/pkg/crypto"
	"github.com/ZeljkoBenovic/tpser/pkg/logger"
)

type MultipleMode struct {
	*base

	config MultipleConfig
}

func NewMultiple(
	log logger.Logger,
	client client,
	config MultipleConfig,
) *MultipleMode {
	return &MultipleMode{
		base:   newBase(log, client, config.Config),
		config: config,
	}
}

func (m *MultipleMode) Run(ctx context.Context) error {
	// Set up the context to stop after the timeout
	runCtx, cancel := context.WithTimeout(ctx, m.config.TxSendTimeout)
	defer cancel()

	pks, err := localCrypto.GenerateFromMnemonic(
		m.config.Mnemonic,
		m.config.NumAccounts,
	)
	if err != nil {
		return fmt.Errorf("could not setup private keys: %w", err)
	}

	return m.sendTransactions(runCtx, pks)
}
