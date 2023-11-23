package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/peterbourgon/ff/v3/ffcli"
)

type rootCfg struct {
	mode     string
	jsonRPC  string
	logLevel string

	privateKey    string
	mnemonic      string
	totalAccounts uint64

	toAddr string

	txPerSec         uint64
	txSendTimeoutMin uint64
	includeTpsReport bool
}

func main() {
	fs := flag.NewFlagSet("root", flag.ExitOnError)

	// Initialize the config
	cfg := &rootCfg{}
	cfg.registerFlags(fs)

	// Create the root command
	cmd := &ffcli.Command{
		ShortUsage: "<subcommand> [flags]",
		LongHelp:   "Suite of tools for Ethereum TPS metrics",
		FlagSet:    fs,
		Exec: func(_ context.Context, _ []string) error {
			return flag.ErrHelp
		},
	}

	cmd.Subcommands = []*ffcli.Command{
		newBlocksFetcherCmd(cfg),
	}

	// Run the command
	if err := cmd.ParseAndRun(context.Background(), os.Args[1:]); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%+v", err)

		os.Exit(1)
	}
}

func (c *rootCfg) registerFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&c.mode,
		"mode",
		"",
		"mode of operation",
	)

	fs.StringVar(
		&c.jsonRPC,
		"json-rpc",
		"",
		"JSON-RPC or WS endpoint of the chain",
	)

	fs.StringVar(
		&c.logLevel,
		"log-level",
		"info",
		"log output level",
	)
}
