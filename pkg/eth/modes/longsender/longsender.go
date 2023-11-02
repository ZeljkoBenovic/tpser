package longsender

import (
	"context"

	"github.com/ZeljkoBenovic/tpser/pkg/conf"
	"github.com/ZeljkoBenovic/tpser/pkg/logger"
	"github.com/ethereum/go-ethereum/ethclient"
)

type longsender struct {
	ctx  context.Context
	log  logger.Logger
	eth  *ethclient.Client
	conf conf.Conf
}

func New(ctx context.Context, log logger.Logger, eth *ethclient.Client, conf conf.Conf) *longsender {
	return &longsender{
		ctx:  ctx,
		log:  log,
		eth:  eth,
		conf: conf,
	}
}

func (l *longsender) RunMode() error {
	l.log.Info("Running longsender...")
	return nil
}
