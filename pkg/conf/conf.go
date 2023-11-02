package conf

import (
	"errors"
	"flag"
	"fmt"

	"github.com/ZeljkoBenovic/tpser/pkg/logger"
)

type Mode string

func (m Mode) String() string {
	return string(m)
}

const (
	BlocksFetcher Mode = "blocks-fetcher"
	LongSender    Mode = "long-sender"
)

type Conf struct {
	Mode    Mode
	JsonRPC string
	Blocks  Blocks
}

type Blocks struct {
	Start int64
	End   int64
}

var (
	ErrJsonRPCNotDefined  = errors.New("json-rpc endpoint not defined")
	ErrEndBlockNotDefined = errors.New("end block not defined")
)

type rawConf struct {
	jsonRpc    string
	blockStart int64
	blockEnd   int64
	mode       string
}

func New(logger logger.Logger) (Conf, error) {
	raw := &rawConf{}
	conf, err := raw.getConfig()
	if err != nil {
		logger.Fatalln("Could not initialize config", "err", err.Error())
	}

	return conf, nil
}

func (c *rawConf) getConfig() (Conf, error) {
	flag.StringVar(&c.jsonRpc, "json-rpc", "", "JSON-RPC or WS endpoint")
	flag.Int64Var(&c.blockStart, "block-start", 1, "the start block range")
	flag.Int64Var(&c.blockEnd, "block-end", 0, "the end block range")
	flag.StringVar(
		&c.mode,
		"mode",
		BlocksFetcher.String(),
		fmt.Sprintf("mode of operation (%s, %s)", BlocksFetcher.String(), LongSender.String()),
	)
	flag.Parse()

	if c.jsonRpc == "" {
		return Conf{}, ErrJsonRPCNotDefined
	}
	if c.mode == BlocksFetcher.String() && c.blockEnd == 0 {
		return Conf{}, ErrEndBlockNotDefined
	}

	return Conf{
		JsonRPC: c.jsonRpc,
		Blocks: Blocks{
			Start: c.blockStart,
			End:   c.blockEnd,
		},
		Mode: Mode(c.mode),
	}, nil
}
