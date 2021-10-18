// Package memory provides an in-memory repository.
package memory

import (
	"github.com/diamondburned/arikawa/v3/discord"
	"go.uber.org/zap"

	"github.com/mavolin/stormy/pkg/repository"
	"github.com/mavolin/stormy/plugin/idea/setup"
)

type Repository struct {
	log *zap.SugaredLogger

	ideaChannelSettings map[discord.ChannelID]setup.ChannelSettings
}

var _ repository.Repository = new(Repository)

func New(l *zap.SugaredLogger) *Repository {
	return &Repository{
		log:                 l.Named("db"),
		ideaChannelSettings: make(map[discord.ChannelID]setup.ChannelSettings),
	}
}

func (r *Repository) IdeaChannelSettings(channelID discord.ChannelID) (*setup.ChannelSettings, error) {
	s, ok := r.ideaChannelSettings[channelID]
	if !ok {
		r.log.With(
			"channel_id", channelID,
			"ret_settings", "nil",
			"err", "nil",
		).Debug("IdeaChannelSettings")

		return nil, nil
	}

	r.log.With(
		"channel_id", channelID,
		"ret_settings", &s,
		"err", "nil",
	).Debug("IdeaChannelSettings")

	return &s, nil
}

func (r *Repository) SetIdeaChannelSettings(channelID discord.ChannelID, s setup.ChannelSettings) error {
	r.log.With(
		"channel_id", channelID,
		"settings", s,
	).Debug("SetIdeaChannelSettings")

	r.ideaChannelSettings[channelID] = s
	return nil
}
