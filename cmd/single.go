package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/ZeljkoBenovic/tpser/pkg/modes/longsender"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/peterbourgon/ff/v3/ffcli"
)

type senderSingleCfg struct {
	rootCfg *senderCfg

	privateKey string
}

func newSenderSingleCmd(rootCfg *senderCfg) *ffcli.Command {
	cfg := &senderSingleCfg{
		rootCfg: rootCfg,
	}

	fs := flag.NewFlagSet("single", flag.ExitOnError)

	rootCfg.rootCfg.registerFlags(fs)
	rootCfg.registerFlags(fs)
	cfg.registerFlags(fs)

	return &ffcli.Command{
		Name:       "single",
		ShortUsage: "single [flags]",
		FlagSet:    fs,
		Exec: func(ctx context.Context, _ []string) error {
			return cfg.exec(ctx)
		},
	}
}

func (c *senderSingleCfg) registerFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&c.privateKey,
		"pk",
		"",
		"the private key for the sender account",
	)
}

func (c *senderSingleCfg) exec(ctx context.Context) error {
	cfg := longsender.SingleConfig{
		PrivateKey: c.privateKey,

		Config: longsender.Config{
			TxSendTimeout: time.Duration(c.rootCfg.txSendDurationMin) * time.Minute,
			ToSendAddress: c.rootCfg.toAddr,
			TPS:           c.rootCfg.txPerSec,
		},
	}

	// Init the client
	client, err := ethclient.Dial(c.rootCfg.rootCfg.JSONRPC)
	if err != nil {
		return fmt.Errorf("unable to dial JSON-RPC, %w", err)
	}

	// TODO add logger
	mode := longsender.NewSingle(nil, client, cfg)

	// Run the mode
	return mode.Run(ctx)
}
