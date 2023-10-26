package app

import (
	"context"
	"os"

	"github.com/ZeljkoBenovic/tpser/conf"
	"github.com/ZeljkoBenovic/tpser/eth"
	"github.com/ZeljkoBenovic/tpser/logger"
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

func (a *app) newApp(_ fx.Lifecycle, eth eth.Eth) {
	//lc.Append(fx.Hook{
	//	OnStart: func(ctx context.Context) error {
	//		go func() {
	//			eth.RunStressTest()
	//
	//		}()
	//
	//		return nil
	//	},
	//	OnStop: nil,
	//})

	eth.GetBlockByNumberStats()
	os.Exit(0)
}

func (a *app) Run() {
	a.fx.Run()
}
