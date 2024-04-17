package app

import (
	"context"
	"os"
	"os/signal"

	"github.com/ZeljkoBenovic/tpser/pkg/conf"
	"github.com/ZeljkoBenovic/tpser/pkg/eth"
	"github.com/ZeljkoBenovic/tpser/pkg/logger"
	"go.uber.org/fx"
)

func Run() {
	newCtx, cancel := context.WithCancel(context.Background())

	go func(cancel context.CancelFunc) {
		sig := make(chan os.Signal)
		signal.Notify(sig, os.Interrupt, os.Kill)
		<-sig
		cancel()
	}(cancel)

	fx.New(
		fx.Provide(
			func() context.Context { return newCtx },
			logger.NewLogrusLogger,
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
