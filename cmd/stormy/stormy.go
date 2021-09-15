package main

import (
	"context"
	"flag"
	stdlog "log"
	"os"
	"os/signal"
	"syscall"

	"github.com/getsentry/sentry-go"
	"github.com/mavolin/adam/pkg/bot"
	"github.com/mavolin/adam/pkg/impl/command/help"
	"go.uber.org/zap"

	"github.com/mavolin/stormy/internal/errhandler"
	"github.com/mavolin/stormy/internal/sentryadam"
	"github.com/mavolin/stormy/internal/setup"
	"github.com/mavolin/stormy/internal/setup/config"
	"github.com/mavolin/stormy/internal/zapadam"
	"github.com/mavolin/stormy/internal/zapstate"
)

var debug = flag.Bool("debug", false, "whether to run in debug mode")

func main() {
	flag.Parse()

	logger, err := setup.Logger(setup.LoggerOptions{Debug: *debug})
	if err != nil {
		stdlog.Fatal("could not create logger:", err.Error())
	}

	if err := run(logger); err != nil {
		logger.With("err", err).
			Fatal("error in main")
	}
}

func run(l *zap.SugaredLogger) error {
	l.Info("reading configuration")
	c, err := config.Read()
	if err != nil {
		return err
	}

	hub, err := setup.Sentry(setup.SentryOptions{
		DSN:              c.Sentry.DSN,
		SampleRate:       c.Sentry.SampleRate,
		TracesSampleRate: c.Sentry.TracesSampleRate,
		Server:           c.Sentry.Server,
	})
	if err != nil {
		return err
	}

	b, err := bot.New(bot.Options{
		Token:                c.BotToken,
		Owners:               c.Owners,
		Status:               c.Status,
		ActivityType:         c.Activity.Type,
		ActivityName:         c.Activity.Name,
		ActivityURL:          c.Activity.URL,
		GatewayErrorHandler:  errhandler.NewGateway(l, hub),
		StateErrorHandler:    errhandler.NewStateError(l, hub),
		StatePanicHandler:    errhandler.NewStatePanic(l, hub),
		NoDefaultMiddlewares: true,
	})
	if err != nil {
		return err
	}

	b.State.AddMiddleware(zapstate.NewMiddleware(l))

	addMiddlewares(b, l, hub)
	addPlugins(b)

	l.Info("starting bot")
	if err := b.Open(context.Background()); err != nil {
		return err
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	recSig := <-sig

	l.Info("received %s, shutting down", recSig)

	return b.State.Close()
}

func addMiddlewares(b *bot.Bot, l *zap.SugaredLogger, hub *sentry.Hub) {
	b.AddMiddleware(sentryadam.NewPreRouteMiddleware(hub))
	b.AddMiddleware(bot.CheckMessageType)
	b.AddMiddleware(bot.CheckHuman) // if Options.AllowBot is true
	b.AddMiddleware(bot.NewSettingsRetriever(bot.NewStaticSettingsProvider()))
	b.AddMiddleware(bot.CheckPrefix)
	b.AddMiddleware(bot.FindCommand)
	b.AddMiddleware(bot.CheckChannelTypes)
	b.AddMiddleware(bot.CheckBotPermissions)
	b.AddMiddleware(bot.NewThrottlerChecker(bot.DefaultThrottlerErrorCheck))

	b.AddMiddleware(sentryadam.PostRouteMiddleware)
	b.AddMiddleware(zapadam.NewMiddleware(l))

	b.AddPostMiddleware(bot.CheckRestrictions)
	b.AddPostMiddleware(bot.ParseArgs)
	b.AddPostMiddleware(sentryadam.PreInvokeMiddleware)
	b.AddPostMiddleware(bot.InvokeCommand)
}

func addPlugins(b *bot.Bot) {
	b.AddCommand(help.New(help.Options{}))
}
