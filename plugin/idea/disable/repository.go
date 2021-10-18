package disable

import (
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/mavolin/adam/pkg/errors"
)

// ErrDisabled is the error returned if attempting to disable brainstorming
// in a channel, where brainstorming is already disabled.
var ErrDisabled = errors.NewUserInfo("Brainstorming is already disabled in this channel.")

type Repository interface {
	IdeaDisableChannel(discord.ChannelID) error
}
