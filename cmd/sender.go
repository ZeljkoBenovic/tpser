package main

import (
	"context"
	"flag"

	"github.com/peterbourgon/ff/v3/ffcli"
)

type senderCfg struct {
	rootCfg *rootCfg

	toAddr string

	txPerSec          uint64
	txSendDurationMin uint64
	includeTpsReport  bool
}

func newSenderCmd(rootCfg *rootCfg) *ffcli.Command {
	cfg := &senderCfg{
		rootCfg: rootCfg,
	}

	fs := flag.NewFlagSet("sender", flag.ExitOnError)

	rootCfg.registerFlags(fs)
	cfg.registerFlags(fs)

	cmd := &ffcli.Command{
		Name:       "sender",
		ShortUsage: "sender [flags] <subcommand>",
		FlagSet:    fs,
		Exec: func(_ context.Context, _ []string) error {
			return flag.ErrHelp
		},
	}

	cmd.Subcommands = []*ffcli.Command{
		newSenderSingleCmd(cfg),
		newSendMultipleCmd(cfg),
	}

	return cmd
}

func (c *senderCfg) registerFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&c.toAddr,
		"to",
		"",
		"address to which the funds will be sent",
	)

	fs.Uint64Var(
		&c.txPerSec,
		"tps",
		100,
		"the number of transactions per second to send",
	)

	fs.Uint64Var(
		&c.txSendDurationMin,
		"duration",
		60,
		"the number of minutes after witch to stop the send",
	)

	fs.BoolVar(
		&c.includeTpsReport,
		"report",
		false,
		"set to true to include tps report after the long-sender node",
	)
}
