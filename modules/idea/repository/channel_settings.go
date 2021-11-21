package repository

import (
	"fmt"
	"time"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/mavolin/adam/pkg/errors"
)

// ErrChannelAlreadyDisabled is the error returned if attempting to disable
// brainstorming in a channel, where brainstorming is already disabled.
var ErrChannelAlreadyDisabled = errors.NewUserInfo("Brainstorming is already disabled in this channel.")

type ChannelSettingsRepository interface {
	DisableIdeaChannel(discord.ChannelID) error
	IdeaChannelSettings(discord.ChannelID) (*ChannelSettings, error)
	SetIdeaChannelSettings(discord.ChannelID, ChannelSettings) error
}

type ChannelSettings struct {
	VoteType     VoteType
	VoteDuration time.Duration
	Anonymous    bool
	Color        discord.Color
}

type VoteType uint8

const (
	Thumbs VoteType = iota
	TwoEmojis
	ThreeEmojis
	FiveEmojis
)

// String returns the string representation of VoteType.
func (t VoteType) String() string {
	switch t {
	case Thumbs:
		return "thumbs"
	case TwoEmojis:
		return "two emojis"
	case ThreeEmojis:
		return "three emojis"
	case FiveEmojis:
		return "five emojis"
	default:
		return fmt.Sprintf("undefined VoteType (%d)", t)
	}
}

// Emojis returns the emojis corresponding to the VoteType.
func (t VoteType) Emojis() []string {
	switch t {
	case Thumbs:
		return []string{"👍", "👎"}
	case TwoEmojis:
		return []string{"😀", "☹"}
	case ThreeEmojis:
		return []string{"😀", "😐", "☹"}
	case FiveEmojis:
		return []string{"😀", "🙂", "😐", "🙁", "☹"}
	default:
		return nil
	}
}
