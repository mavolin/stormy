// Package repository provides the idea repository.
package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/mavolin/adam/pkg/errors"
)

// ErrChannelAlreadyDisabled is the error returned if attempting to disable
// brainstorming in a channel, where brainstorming is already disabled.
var ErrChannelAlreadyDisabled = errors.NewUserInfo("Brainstorming is already disabled in this channel.")

type Repository interface {
	DisableIdeaChannel(discord.ChannelID) error
	IdeaChannelSettings(discord.ChannelID) (*ChannelSettings, error)
	SetIdeaChannelSettings(discord.ChannelID, ChannelSettings) error

	Idea(discord.MessageID) (*Idea, error)
	SaveIdea(*Idea) error
	DeleteIdea(id discord.MessageID) error

	// ExpiringIdeas returns ideas that expire after or at the given time.
	// If an idea expires at the given time, its idea must be higher than
	// afterID to be returned.
	//
	// The returned slice must be sorted by VoteUntil in ascending order.
	// If two deadlines match, those entries are sorted by id in ascending
	// order.
	ExpiringIdeas(afterT time.Time, afterID discord.MessageID, limit int) ([]*Idea, error)
	// ExpiredIdeas returns ideas that expired before the given time.
	ExpiredIdeas(before time.Time) (IdeaCursor, error)
	// DeleteExpiredIdeas deletes ideas that expired before the given time.
	DeleteExpiredIdeas(before time.Time) error
}

// =============================================================================
// ChannelSettings
// =====================================================================================

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

// VoteEmojis contains all the emojis used by different VoteTypes.
var VoteEmojis = []string{"👍", "👎", "😀", "🙂", "😐", "🙁", "☹"}

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

// =============================================================================
// Idea
// =====================================================================================

type Idea struct {
	GuildID   discord.GuildID
	ChannelID discord.ChannelID
	MessageID discord.MessageID

	GlobalSectionEmojis []string
	Groups              []SectionGroup

	VoteType  VoteType
	VoteUntil *time.Time
}

type SectionGroup struct {
	Title  string
	Emojis []string
}

type IdeaCursor interface {
	// BatchLength returns the length of the current batch.
	BatchLength() int
	Next(context.Context) (*Idea, error)
	Close(context.Context) error
}
