// Package memory provides an in-memory repository.
package memory

import (
	"github.com/diamondburned/arikawa/v3/discord"

	"github.com/mavolin/stormy/pkg/repository"
	"github.com/mavolin/stormy/plugin/idea/setup"
)

type Repository struct {
	ideaChannelSettings map[discord.ChannelID]setup.ChannelSettings
}

var _ repository.Repository = new(Repository)

func New() *Repository {
	return &Repository{
		ideaChannelSettings: make(map[discord.ChannelID]setup.ChannelSettings),
	}
}

func (r *Repository) IdeaChannelSettings(channelID discord.ChannelID) (*setup.ChannelSettings, error) {
	s, ok := r.ideaChannelSettings[channelID]
	if !ok {
		return nil, nil
	}

	return &s, nil
}

func (r *Repository) IdeaSetChannelSettings(channelID discord.ChannelID, s setup.ChannelSettings) error {
	r.ideaChannelSettings[channelID] = s
	return nil
}

func (r *Repository) IdeaDisableChannel(channelID discord.ChannelID) error {
	delete(r.ideaChannelSettings, channelID)

	return nil
}
