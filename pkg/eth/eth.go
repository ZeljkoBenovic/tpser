package eth

import (
	"context"
	"errors"
	"github.com/ZeljkoBenovic/tpser/pkg/prom"

	"github.com/ZeljkoBenovic/tpser/pkg/conf"
	"github.com/ZeljkoBenovic/tpser/pkg/eth/modes/getblocks"
	"github.com/ZeljkoBenovic/tpser/pkg/eth/modes/longsender"
	"github.com/ZeljkoBenovic/tpser/pkg/eth/modes/txinfo"
	"github.com/ZeljkoBenovic/tpser/pkg/logger"
	"github.com/ethereum/go-ethereum/ethclient"
)

// Eth is the interface for the main application
type Eth interface {
	// Run runs the `tpser` application with the specified mode of operation
	Run() error
}

// Common is the interface that all Eth modules must implement
type Common interface {
	RunMode() error
}

var (
	ErrModeNotSupported = errors.New("mode not supported")
)

// factoryFunc is the function which must return Common interface
type factoryFunc func(context.Context, logger.Logger, *ethclient.Client, conf.Conf, *prom.Prom) Common

// modesFactory is a map of functions, with conf.Mode as key, that returns a Common interface.
var modesFactory = map[conf.Mode]factoryFunc{
	conf.BlocksFetcher: func(ctx context.Context, log logger.Logger, eth *ethclient.Client, conf conf.Conf, _ *prom.Prom) Common {
		return getblocks.New(ctx, log, eth, conf)
	},
	conf.LongSender: func(ctx context.Context, log logger.Logger, eth *ethclient.Client, conf conf.Conf, prom *prom.Prom) Common {
		return longsender.New(ctx, log, eth, conf, prom)
	},
	conf.TxInfo: func(ctx context.Context, log logger.Logger, eth *ethclient.Client, conf conf.Conf, _ *prom.Prom) Common {
		return txinfo.New(ctx, log, eth, conf)
	},
}

type eth struct {
	ctx context.Context

	conf         conf.Conf
	log          logger.Logger
	ethClient    *ethclient.Client
	prom         *prom.Prom
	modesFactory map[conf.Mode]factoryFunc
}

func New(conf conf.Conf, log logger.Logger, ctx context.Context, prom *prom.Prom) (Eth, error) {
	e, err := ethclient.Dial(conf.JsonRPC)
	if err != nil {
		log.Error("Could not dial json-rpc", "json-rpc", conf.JsonRPC)
		return nil, err
	}
	return &eth{
		ethClient:    e,
		log:          log,
		ctx:          ctx,
		conf:         conf,
		prom:         prom,
		modesFactory: modesFactory,
	}, nil
}

func (e *eth) Run() error {
	modeConstructor, ok := e.modesFactory[e.conf.Mode]
	if !ok {
		return ErrModeNotSupported
	}

	mode := modeConstructor(e.ctx, e.log, e.ethClient, e.conf, e.prom)
	return mode.RunMode()
}
