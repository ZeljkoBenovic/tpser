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

type multipleCfg struct {
	rootCfg *senderCfg

	mnemonic      string
	totalAccounts uint64
}

func newSendMultipleCmd(rootCfg *senderCfg) *ffcli.Command {
	cfg := &multipleCfg{
		rootCfg: rootCfg,
	}

	fs := flag.NewFlagSet("multiple", flag.ExitOnError)

	rootCfg.rootCfg.registerFlags(fs)
	rootCfg.registerFlags(fs)
	cfg.registerFlags(fs)

	return &ffcli.Command{
		Name:       "multiple",
		ShortUsage: "multiple [flags]",
		FlagSet:    fs,
		Exec: func(ctx context.Context, _ []string) error {
			return cfg.exec(ctx)
		},
	}
}

func (c *multipleCfg) registerFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&c.mnemonic,
		"mnemonic",
		"",
		"mnemonic string to derive accounts from",
	)

	fs.Uint64Var(
		&c.totalAccounts,
		"mnemonic-addr",
		1,
		"total number of account to send transactions from",
	)
}

func (c *multipleCfg) exec(ctx context.Context) error {
	cfg := longsender.MultipleConfig{
		Mnemonic:    c.mnemonic,
		NumAccounts: c.totalAccounts,

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
	mode := longsender.NewMultiple(nil, client, cfg)

	// Run the mode
	return mode.Run(ctx)
}
