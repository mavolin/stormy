package main

import (
	"flag"
	"os"
	"os/signal"

	"github.com/mavolin/adam/pkg/bot"
	"github.com/mavolin/adam/pkg/errors"
	"github.com/mavolin/adam/pkg/impl/command/help"
	"go.uber.org/zap"

	"github.com/mavolin/stormy/internal/config"
	"github.com/mavolin/stormy/internal/zaplog"
)

var (
	debug      = flag.Bool("debug", false, "enables debug mode")
	configPath = flag.String("config", "", "The path to the config file, if using one.")
)

var log *zap.SugaredLogger

func init() {
	flag.Parse()

	zaplog.Init(*debug)
	log = zap.S().Named("startup")

	if err := config.Load(*configPath); err != nil {
		log.With("err", err).
			Fatal("unable to load config")
	}
}

func main() {
	b, err := bot.New(bot.Options{
		Token:               config.C.Token,
		Owners:              config.C.Owners,
		ActivityType:        config.C.ActivityType,
		ActivityName:        config.C.ActivityName,
		ActivityURL:         config.C.ActivityURL,
		GatewayErrorHandler: zaplog.GatewayLog(zap.S().Named("gateway")),
		StateErrorHandler:   zaplog.StateError(zap.S().Named("state")),
		StatePanicHandler:   zaplog.StatePanic(zap.S().Named("state")),
	})
	if err != nil {
		log.Fatal(err)
	}

	errors.Log = zaplog.CommandError(zap.S().Named("bot"))

	addPlugins(b)

	if err = b.Open(); err != nil {
		log.With("err", err).
			Fatal("unable to open gateway connection")
	}

	wait()
	log.Info("received SIGINT, shutting down")

	if err = b.Close(); err != nil {
		log.With("err", err).
			Error("unable to close gracefully")
	}
}

func wait() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)

	<-sig
}

func addPlugins(b *bot.Bot) {
	b.AddCommand(help.New(help.Options{}))
}
