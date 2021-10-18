// Package logwrap provides a repository that logs what another repository
// does.
package logwrap

import (
	"github.com/diamondburned/arikawa/v3/discord"
	"go.uber.org/zap"

	"github.com/mavolin/stormy/pkg/repository"
	"github.com/mavolin/stormy/plugin/idea/setup"
)

type Wrapper struct {
	r repository.Repository
	l *zap.SugaredLogger
}

var _ repository.Repository = new(Wrapper)

func Wrap(r repository.Repository, l *zap.SugaredLogger) *Wrapper {
	return &Wrapper{r: r, l: l.Named("db")}
}

func (w *Wrapper) IdeaDisableChannel(channelID discord.ChannelID) error {
	err := w.r.IdeaDisableChannel(channelID)

	w.l.With(
		"channel_id", channelID,
		"err", err,
	).Debug("IdeaDisableChannel")

	return err
}

func (w *Wrapper) IdeaChannelSettings(channelID discord.ChannelID) (*setup.ChannelSettings, error) {
	s, err := w.r.IdeaChannelSettings(channelID)

	w.l.With(
		"channel_id", channelID,
		"ret_settings", s,
		"err", err,
	).Debug("IdeaChannelSettings")

	return s, err
}

func (w *Wrapper) IdeaSetChannelSettings(channelID discord.ChannelID, s setup.ChannelSettings) error {
	err := w.r.IdeaSetChannelSettings(channelID, s)

	w.l.With(
		"channel_id", channelID,
		"settings", s,
		"err", err,
	).Debug("IdeaSetChannelSettings")

	return err
}
