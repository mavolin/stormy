// Package config reads the application's configuration.
package config

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/caarlos0/env/v6"
	"github.com/diamondburned/arikawa/v3/discord"
)

type (
	Config struct {
		BotToken string           `env:"STORMY_BOT_TOKEN,notEmpty"`
		Owners   []discord.UserID `env:"STORMY_OWNERS"`

		Status   discord.Status `env:"STORMY_STATUS"`
		Activity Activity

		Sentry Sentry
		Mongo  Mongo
	}

	Activity struct {
		Type discord.ActivityType `env:"STORMY_ACTIVITY_TYPE"`
		Name string               `env:"STORMY_ACTIVITY_NAME"`
		URL  discord.URL          `env:"STORMY_ACTIVITY_URL"`
	}

	Sentry struct {
		DSN              string  `env:"STORMY_SENTRY_DSN"`
		Server           string  `env:"STORMY_SENTRY_SERVER"`
		SampleRate       float64 `env:"STORMY_SENTRY_SAMPLE_RATE"`
		TracesSampleRate float64 `env:"STORMY_SENTRY_TRACES_SAMPLE_RATE"`
	}

	Mongo struct {
		URI    string `env:"SENTRY_MONGO_URI"`
		DBName string `env:"SENTRY_MONGO_DB_NAME"`
	}
)

var parseFuncs = map[reflect.Type]env.ParserFunc{
	reflect.TypeOf(discord.ActivityType(0)): parseActivityType,
}

func Read() (*Config, error) {
	c := &Config{Status: discord.OnlineStatus}

	return c, env.ParseWithFuncs(c, parseFuncs)
}

func parseActivityType(activityStr string) (interface{}, error) {
	switch strings.ToLower(activityStr) {
	case "game", "playing":
		return discord.GameActivity, nil
	case "streaming":
		return discord.StreamingActivity, nil
	case "listening":
		return discord.ListeningActivity, nil
	case "watching":
		return discord.WatchingActivity, nil
	default:
		return nil, fmt.Errorf("config: unknown activity type '%s'", activityStr)
	}
}
