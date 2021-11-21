// Package format provides means to parse and format ideas.
package format

import (
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/mavolin/disstate/v4/pkg/event"

	"github.com/mavolin/stormy/modules/idea/repository"
)

// Idea parses and formats an idea.
func Idea(e *event.MessageCreate, set *repository.ChannelSettings) (*IdeaData, error) {
	p := newParser(e)
	i, err := p.Parse()
	if err != nil {
		return nil, err
	}

	f := newIdeaFormatter(i, e, set)
	return f.format()
}

// Votes formats the votes of an idea.
// It returns up to two embed fields that display the votes.
func Votes(i *repository.Idea, msg *discord.Message) VoteData {
	return newVoteFormatter(i, msg).format()
}
