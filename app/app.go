package app

import (
	"context"
	"os"

	"github.com/ZeljkoBenovic/tpser/pkg/conf"
	"github.com/ZeljkoBenovic/tpser/pkg/eth"
	"github.com/ZeljkoBenovic/tpser/pkg/logger"
	"go.uber.org/fx"
)

type App interface {
	Run()
}

type app struct {
	fx  *fx.App
	eth eth.Eth
}

func New() App {
	a := &app{}
	a.fx = fx.New(
		fx.Provide(
			context.Background,
			logger.New,
			conf.New,
			eth.New,
		),
		fx.Invoke(a.newApp),
		fx.NopLogger,
	)

	return a
}

func (a *app) newApp(_ fx.Lifecycle, eth eth.Eth, log logger.Logger) {
	if err := eth.Run(); err != nil {
		log.Fatalln("Could not run application", "err", err.Error())
	}
	
	os.Exit(0)
}

func (a *app) Run() {
	a.fx.Run()
}
