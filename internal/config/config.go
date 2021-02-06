// Package config provides utilities to interact with the configuration
package config

import (
	"strings"

	"github.com/diamondburned/arikawa/v2/discord"
	"github.com/mavolin/adam/pkg/errors"
	"github.com/spf13/viper"
)

// C is the global config
var C config

type config struct {
	Token        string `mapstructure:"bot_token"`
	Owners       []discord.UserID
	ActivityType discord.ActivityType
	ActivityName string
	ActivityURL  string `mapstructure:"activity_url"`
}

// Load loads the config.
// If configPath is not empty, the config at that path will be loaded, instead
// of searching in the current directory, ./config and $CONFIG_DIR/
func Load(configPath string) error {
	v := viper.New()

	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.SetEnvPrefix("stormy")
	viper.AutomaticEnv()

	if len(configPath) > 0 {
		viper.SetConfigFile(configPath)
	} else {
		v.AddConfigPath(".")
		v.AddConfigPath("config")
		v.AddConfigPath("$CONFIG_DIR/")
		viper.SetConfigName("stormy")
	}

	err := v.ReadInConfig()
	if !errors.As(err, new(viper.ConfigFileNotFoundError)) {
		return err
	}

	if err = v.Unmarshal(&C); err != nil {
		return err
	}

	specialUnmarshal(v)
	return nil
}

func specialUnmarshal(v *viper.Viper) {
	C.ActivityType, C.ActivityName = parseActivity(v)
	C.Owners = userIDs(v.GetStringSlice("owners"))
}

var activityTypes = map[discord.ActivityType]string{
	discord.GameActivity:      "Playing",
	discord.StreamingActivity: "Streaming",
	discord.ListeningActivity: "Listening to",
	discord.WatchingActivity:  "Watching",
}

func parseActivity(v *viper.Viper) (t discord.ActivityType, name string) {
	activity := v.GetString("activity")

	for t, tstr := range activityTypes {
		if strings.HasPrefix(activity, tstr) {
			if len(activity) > len(tstr)+1 {
				return t, activity[len(tstr)+1:]
			}

			return 0, ""
		}
	}

	return 0, ""
}

// userIDs is a helper used to retrieve a slice of discord.UserIDs.
func userIDs(idsStr []string) []discord.UserID {
	ids := make([]discord.UserID, len(idsStr))

	for i, idStr := range idsStr {
		s, err := discord.ParseSnowflake(idStr)
		if err != nil {
			return nil
		}

		ids[i] = discord.UserID(s)
	}

	return ids
}
