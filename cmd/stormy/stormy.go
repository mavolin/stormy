package main

import (
	"context"
	"flag"
	stdlog "log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mavolin/disstate/v4/pkg/event"
	"github.com/mavolin/disstate/v4/pkg/state"
	"go.uber.org/zap"

	"github.com/mavolin/stormy/internal/config"
	"github.com/mavolin/stormy/internal/setup"
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
	ml := l.Named("main")

	ml.Info("reading configuration")
	c, err := config.Read()
	if err != nil {
		return err
	}

	ml.With("data", c).Debug("config read")

	hub, err := setup.Sentry(setup.SentryOptions{
		DSN:              c.Sentry.DSN,
		SampleRate:       c.Sentry.SampleRate,
		TracesSampleRate: c.Sentry.TracesSampleRate,
		Server:           c.Sentry.Server,
	})
	if err != nil {
		return err
	}

	repo := setup.Repository(setup.RepositoryOptions{Logger: l})

	b, err := setup.Bot(setup.BotOptions{
		Token:        c.BotToken,
		Owners:       c.Owners,
		Status:       c.Status,
		ActivityType: c.Activity.Type,
		ActivityName: c.Activity.Name,
		ActivityURL:  c.Activity.URL,

		Logger:     l,
		Hub:        hub,
		Repository: repo,
	})
	if err != nil {
		return err
	}

	err = setup.Services(setup.ServiceOptions{
		State:      b.State,
		Repository: repo,
		Logger:     l,
		Hub:        hub,
	})
	if err != nil {
		return err
	}

	b.State.AddHandlerOnce(func(_ *state.State, e *event.Ready) {
		ml.Infof("received first ready event, accepting commands as @%s#%s",
			e.User.Username, e.User.Discriminator)
	})

	b.AddIntents(b.State.DeriveIntents())

	ml.Info("starting bot")
	if err = b.Open(4 * time.Second); err != nil {
		return err
	}

	ml.Info("started")

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	recSig := <-sig

	l.Infof("received %s, waiting for all commands to finish", recSig)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	defer func() {
		select {
		case <-ctx.Done():
			l.Info("timeout: I won't wait any longer for event handlers to finish")
		default:
			l.Info("done")
		}
	}()

	return b.State.Close(ctx)
}
