// Package config reads the application's configuration.
package config

import (
	"strings"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/structs"
)

type (
	Config struct {
		BotToken string           `koanf:"bottoken"`
		Owners   []discord.UserID `koanf:"owners"`

		Status   discord.Status `koanf:"status"`
		Activity Activity       `koanf:"activity"`

		Sentry Sentry `koanf:"sentry"`
		Mongo  Mongo  `koanf:"mongo"`
	}

	Activity struct {
		Type discord.ActivityType `koanf:"type"`
		Name string               `koanf:"name"`
		URL  discord.URL          `koanf:"url"`
	}

	Sentry struct {
		DSN        string  `koanf:"dsn"`
		Server     string  `koanf:"server"`
		SampleRate float64 `koanf:"samplerate"`
	}

	Mongo struct {
		URI    string `koanf:"uri"`
		DBName string `koanf:"dbname"`
	}
)

var defaultConfig = &Config{
	Status: discord.OnlineStatus,
}

func Read(configFilePath string) (c *Config, err error) {
	k := koanf.New(".")

	if err := k.Load(structs.Provider(defaultConfig, "koanf"), nil); err != nil {
		return nil, err
	}

	// try to load
	_ = k.Load(file.Provider(configFilePath), json.Parser())

	err = k.Load(env.Provider("STORMY_", ".", envReplacer("STORMY_")), nil)
	if err != nil {
		return nil, err
	}

	return c, k.Unmarshal("", &c)
}

func envReplacer(prefix string) func(string) string {
	return func(s string) string {
		return strings.ReplaceAll(strings.ToLower(strings.TrimPrefix(prefix, s)), "_", ".")
	}
}
