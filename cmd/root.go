package main

import (
	"context"
	"flag"

	"github.com/peterbourgon/ff/v3/ffcli"
)

type rootCfg struct {
	JSONRPC  string
	LogLevel string
}

func newRootCmd() *ffcli.Command {
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
		newFetcherCmd(cfg),
		newSenderCmd(cfg),
	}

	return cmd
}

func (c *rootCfg) registerFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&c.JSONRPC,
		"json-rpc",
		"",
		"JSON-RPC or WS endpoint of the chain",
	)

	fs.StringVar(
		&c.LogLevel,
		"log-level",
		"info",
		"log output level",
	)
}
