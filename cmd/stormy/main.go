package main

import (
	"flag"
	"os"
	"os/signal"
	"strings"

	"github.com/diamondburned/arikawa/v2/discord"
	"github.com/mavolin/adam/pkg/bot"
	"github.com/mavolin/adam/pkg/errors"
	"github.com/mavolin/adam/pkg/impl/command/help"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/mavolin/stormy/internal/zaplog"
)

var (
	debugMode  = flag.Bool("debug", false, "enables debug mode")
	configPath = flag.String("config", "stormy.json", "The path to the config file, if using one.")
)

func main() {
	initLogger()
	readConfig()

	activityType, activityName := parseActivity()

	b, err := bot.New(bot.Options{
		Token:               viper.GetString("token"),
		Owners:              parseOwners(),
		ActivityType:        activityType,
		ActivityName:        activityName,
		ActivityURL:         viper.GetString("activity_url"),
		GatewayErrorHandler: zaplog.GatewayLog(zap.S().Named("gateway")),
		StateErrorHandler:   zaplog.StateError(zap.S().Named("state")),
		StatePanicHandler:   zaplog.StatePanic(zap.S().Named("state")),
	})
	if err != nil {
		zap.S().Named("startup").Fatal(err)
	}

	addPlugins(b)

	if err = b.Open(); err != nil {
		zap.S().Named("startup").
			With("err", err).
			Fatal("unable to open gateway connection")
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)

	<-sig

	zap.S().Named("startup").Info("received SIGINT, shutting down")

	if err = b.Close(); err != nil {
		zap.S().Named("startup").
			With("err", err).
			Error("unable to close gracefully")
	}
}

func initLogger() {
	if *debugMode {
		l, err := zap.NewDevelopment()
		if err != nil {
			panic(err)
		}

		zap.ReplaceGlobals(l)
	} else {
		l, err := zap.NewProduction()
		if err != nil {
			panic(err)
		}

		zap.ReplaceGlobals(l)
	}

	errors.Log = zaplog.CommandError(zap.S().Named("bot"))
}

func readConfig() {
	viper.AutomaticEnv()
	viper.AddConfigPath(*configPath)

	if err := viper.ReadInConfig(); err != nil {
		zap.S().Named("startup").Fatal(err)
	}
}

func addPlugins(b *bot.Bot) {
	b.AddCommand(help.New(help.Options{}))
}

func parseOwners() []discord.UserID {
	ownersstr := viper.GetStringSlice("owners")
	owners := make([]discord.UserID, len(ownersstr))

	for i, o := range ownersstr {
		s, err := discord.ParseSnowflake(o)
		if err != nil {
			zap.S().Named("startup").
				With("err", err).
				Fatal("invalid owner")
		}

		owners[i] = discord.UserID(s)
	}

	return owners
}

func parseActivity() (t discord.ActivityType, name string) {
	activity := viper.GetString("activity")

	switch {
	case strings.HasPrefix(activity, "Playing"):
		return discord.GameActivity, activityName(activity, "Playing")
	case strings.HasPrefix(activity, "Streaming"):
		return discord.StreamingActivity, activityName(activity, "Streaming")
	case strings.HasPrefix(activity, "Listening to"):
		return discord.ListeningActivity, activityName(activity, "Listening to")
	case strings.HasPrefix(activity, "Watching"):
		return discord.WatchingActivity, activityName(activity, "Watching")
	default:
		zap.S().Named("config").Warn("invalid activity \"", activity, "\", using none")
		return 0, ""
	}
}

func activityName(activity string, activityType string) string {
	if len(activity) <= len(activityType)+1 {
		zap.S().Named("config").Warn("no activity name specified, using no activity")
		return ""
	}

	return activity[len(activityType)+1:]
}
