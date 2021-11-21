package repository

import (
	"context"
	"time"

	"github.com/diamondburned/arikawa/v3/discord"
)

type IdeaRepository interface {
	Idea(context.Context, discord.MessageID) (*Idea, error)
	SaveIdea(context.Context, *Idea) error
	DeleteIdea(ctx context.Context, id discord.MessageID) error

	// ExpiringIdeas returns ideas that expire after or at the given time.
	// If an idea expires at the given time, its idea must be higher than
	// afterID to be returned.
	//
	// The returned slice must be sorted by VoteUntil in ascending order.
	// If two deadlines match, those entries are sorted by id in ascending
	// order.
	ExpiringIdeas(ctx context.Context, afterT time.Time, afterID discord.MessageID, limit int) ([]Idea, error)
	// ExpiredIdeas returns ideas that expired before the given time.
	ExpiredIdeas(ctx context.Context, before time.Time) (IdeaCursor, error)
	// DeleteExpiredIdeas deletes ideas that expired before the given time.
	DeleteExpiredIdeas(ctx context.Context, before time.Time) error
}

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
