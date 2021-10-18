package setup

import (
	"fmt"
	"time"

	"github.com/diamondburned/arikawa/v3/discord"
)

type Repository interface {
	IdeaChannelSettings(discord.ChannelID) (*ChannelSettings, error)
	SetIdeaChannelSettings(discord.ChannelID, ChannelSettings) error
}

type ChannelSettings struct {
	VoteType     VoteType
	VoteDuration time.Duration
	Anonymous    bool
	Color        discord.Color
	Thumbnail    string
}

func (c ChannelSettings) Equals(other ChannelSettings) bool {
	return c.VoteType == other.VoteType && c.VoteDuration == other.VoteDuration &&
		c.Anonymous == other.Anonymous && c.Color == other.Color &&
		c.Thumbnail == other.Thumbnail
}

type VoteType uint8

const (
	Thumbs VoteType = iota
	TwoEmojis
	ThreeEmojis
	FiveEmojis
)

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