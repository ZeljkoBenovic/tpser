package app

import (
	"context"
	"os"

	"github.com/ZeljkoBenovic/tpser/pkg/conf"
	"github.com/ZeljkoBenovic/tpser/pkg/eth"
	"github.com/ZeljkoBenovic/tpser/pkg/logger"
	"go.uber.org/fx"
)

func Run() {
	fx.New(
		fx.Provide(
			context.Background,
			logger.New,
			conf.New,
			eth.New,
		),
		fx.Invoke(mainApp),
		fx.NopLogger,
	).Run()
}

func mainApp(eth eth.Eth, log logger.Logger) {
	if err := eth.Run(); err != nil {
		log.Fatalln("Could not run application", "err", err.Error())
	}

	os.Exit(0)
}
