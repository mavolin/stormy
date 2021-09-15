package setup

import (
	"github.com/getsentry/sentry-go"

	"github.com/mavolin/stormy/internal/meta"
)

type SentryOptions struct {
	DSN        string
	SampleRate float64
	Server     string
}

// Sentry sets up a *sentry.Hub.
func Sentry(o SentryOptions) (*sentry.Hub, error) {
	c, err := sentry.NewClient(sentry.ClientOptions{
		Dsn:        o.DSN,
		SampleRate: o.SampleRate,
		ServerName: o.Server,
		Release:    meta.Version,
	})
	if err != nil {
		return nil, err
	}

	return sentry.NewHub(c, sentry.NewScope()), nil
}