package conf

import (
	"errors"
	"flag"

	"github.com/ZeljkoBenovic/tpser/logger"
)

type Conf struct {
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
	flag.Parse()

	if c.jsonRpc == "" {
		return Conf{}, ErrJsonRPCNotDefined
	}
	if c.blockEnd == 0 {
		return Conf{}, ErrEndBlockNotDefined
	}

	return Conf{
		JsonRPC: c.jsonRpc,
		Blocks: Blocks{
			Start: c.blockStart,
			End:   c.blockEnd,
		},
	}, nil
}
