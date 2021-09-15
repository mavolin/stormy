package setup

import (
	"github.com/getsentry/sentry-go"

	"github.com/mavolin/stormy/internal/meta"
)

type SentryOptions struct {
	DSN              string
	SampleRate       float64
	TracesSampleRate float64
	Server           string
}

// Sentry sets up a *sentry.Hub.
func Sentry(o SentryOptions) (*sentry.Hub, error) {
	if o.DSN == "" {
		return sentry.NewHub(nil, sentry.NewScope()), nil
	}

	c, err := sentry.NewClient(sentry.ClientOptions{
		Dsn:              o.DSN,
		SampleRate:       o.SampleRate,
		TracesSampleRate: o.TracesSampleRate,
		ServerName:       o.Server,
		Release:          meta.Version,
	})
	if err != nil {
		return nil, err
	}

	return sentry.NewHub(c, sentry.NewScope()), nil
}
