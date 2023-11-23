package main

import (
	"context"
	"flag"

	"github.com/ZeljkoBenovic/tpser/pkg/modes"
	"github.com/peterbourgon/ff/v3/ffcli"
)

type blocksFetcherCfg struct {
	rootCfg *rootCfg

	blockStart uint64
	blockEnd   uint64
	blockTail  uint64
}

func newBlocksFetcherCmd(rootCfg *rootCfg) *ffcli.Command {
	cfg := &blocksFetcherCfg{
		rootCfg: rootCfg,
	}

	fs := flag.NewFlagSet("fetcher", flag.ExitOnError)

	rootCfg.registerFlags(fs)
	cfg.registerFlags(fs)

	return &ffcli.Command{
		Name:       "fetcher",
		ShortUsage: "fetcher [flags] [<arg>...]",
		FlagSet:    fs,
		Exec: func(ctx context.Context, _ []string) error {
			return cfg.exec(ctx)
		},
	}
}

func (c *blocksFetcherCfg) registerFlags(fs *flag.FlagSet) {
	fs.Uint64Var(
		&c.blockStart,
		"block-start",
		1,
		"the start block range",
	)

	fs.Uint64Var(
		&c.blockEnd,
		"block-end",
		0,
		"the end block range",
	)

	fs.Uint64Var(
		&c.blockTail,
		"block-tail",
		0,
		"the range of blocks to fetch from latest",
	)
}

func (c *blocksFetcherCfg) exec(ctx context.Context) error {
	cfg := modes.GetBlocksConfig{
		Start: c.blockStart,
		End:   c.blockEnd,
		Tail:  c.blockTail,
	}

	// TODO init client
	mode := modes.NewGetBlocksMode(nil, cfg)

	return mode.Run(ctx)
}
