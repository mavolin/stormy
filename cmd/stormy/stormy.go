package main

import (
	"flag"
	stdlog "log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/getsentry/sentry-go"
	"github.com/mavolin/adam/pkg/bot"
	"github.com/mavolin/adam/pkg/impl/command/help"
	"github.com/mavolin/sentryadam/pkg/sentryadam"
	"go.uber.org/zap"

	"github.com/mavolin/stormy/internal/config"
	"github.com/mavolin/stormy/internal/errhandler"
	"github.com/mavolin/stormy/internal/setup"
	"github.com/mavolin/stormy/internal/zapadam"
	"github.com/mavolin/stormy/internal/zapstate"
	"github.com/mavolin/stormy/pkg/repository"
	"github.com/mavolin/stormy/plugin/idea"
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

	repo := setup.Repository(setup.RepositoryOptions{})
	addPlugins(b, repo)

	b.AddIntents(b.State.DeriveIntents())
	b.AddIntents(gateway.IntentGuildMessageTyping)

	l.Info("starting bot")
	if err = b.Open(4 * time.Second); err != nil {
		return err
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	recSig := <-sig

	l.Infof("received %s, shutting down", recSig)

	return b.State.Close()
}

func addMiddlewares(b *bot.Bot, l *zap.SugaredLogger, hub *sentry.Hub) {
	sentryWrapper := sentryadam.New(sentryadam.Options{Hub: hub})

	b.AddMiddleware(sentryWrapper.PreRouteMiddleware)
	b.AddMiddleware(zapadam.NewFallbackMiddleware(l))
	b.AddMiddleware(bot.CheckMessageType)
	b.AddMiddleware(bot.CheckHuman) // if Options.AllowBot is true
	b.AddMiddleware(bot.NewSettingsRetriever(bot.StaticSettings()))
	b.AddMiddleware(bot.CheckPrefix)
	b.AddMiddleware(bot.FindCommand)
	b.AddMiddleware(bot.CheckChannelTypes)
	b.AddMiddleware(bot.CheckBotPermissions)
	b.AddMiddleware(bot.NewThrottlerChecker(bot.DefaultThrottlerErrorCheck))
	b.AddMiddleware(sentryWrapper.PostRouteMiddleware)

	b.AddMiddleware(zapadam.NewMiddleware(l))

	b.AddPostMiddleware(bot.CheckRestrictions)
	b.AddPostMiddleware(bot.ParseArgs)
	b.AddPostMiddleware(sentryWrapper.PreInvokeMiddleware)
	b.AddPostMiddleware(bot.InvokeCommand)
}

func addPlugins(b *bot.Bot, r repository.Repository) {
	b.AddCommand(help.New(help.Options{}))

	b.AddModule(idea.New(r))
}
