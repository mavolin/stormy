package logwrap

import (
	"context"
	"time"

	"github.com/diamondburned/arikawa/v3/discord"

	idearepo "github.com/mavolin/stormy/modules/idea/repository"
)

func (w *Wrapper) IdeaChannelSettings(ctx context.Context, channelID discord.ChannelID) (*idearepo.ChannelSettings,
	error) {
	s, err := w.r.IdeaChannelSettings(ctx, channelID)

	w.l.With(
		"channel_id", channelID,
		"ret_settings", s,
		"err", err,
	).Debug("IdeaChannelSettings")

	return s, err
}

func (w *Wrapper) SetIdeaChannelSettings(
	ctx context.Context, channelID discord.ChannelID,
	s idearepo.ChannelSettings,
) error {
	err := w.r.SetIdeaChannelSettings(ctx, channelID, s)

	w.l.With(
		"channel_id", channelID,
		"settings", s,
		"err", err,
	).Debug("SetIdeaChannelSettings")

	return err
}

func (w *Wrapper) DisableIdeaChannel(ctx context.Context, channelID discord.ChannelID) error {
	err := w.r.DisableIdeaChannel(ctx, channelID)

	w.l.With(
		"channel_id", channelID,
		"err", err,
	).Debug("DisableIdeaChannel")

	return err
}

func (w *Wrapper) Idea(ctx context.Context, messageID discord.MessageID) (*idearepo.Idea, error) {
	i, err := w.r.Idea(ctx, messageID)

	w.l.With(
		"message_id", messageID,
		"ret_idea", i,
		"err", err,
	).Debug("Idea")

	return i, err
}

func (w *Wrapper) SaveIdea(ctx context.Context, idea *idearepo.Idea) error {
	err := w.r.SaveIdea(ctx, idea)

	w.l.With(
		"idea", idea,
		"err", err,
	).Debug("SaveIdea")

	return err
}

func (w *Wrapper) DeleteIdea(ctx context.Context, messageID discord.MessageID) error {
	err := w.r.DeleteIdea(ctx, messageID)

	w.l.With(
		"message_id", messageID,
		"err", err,
	).Debug("DeleteIdea")

	return err
}

func (w *Wrapper) ExpiringIdeas(
	ctx context.Context, afterT time.Time, afterID discord.MessageID,
	limit int,
) ([]idearepo.Idea,
	error) {
	i, err := w.r.ExpiringIdeas(ctx, afterT, afterID, limit)

	w.l.With(
		"after_t", afterT,
		"after_id", afterID,
		"limit", limit,
		"ret_ideas", i,
		"err", err,
	).Debug("ExpiringIdeas")

	return i, err
}

func (w *Wrapper) ExpiredIdeas(ctx context.Context, t time.Time) (idearepo.IdeaCursor, error) {
	c, err := w.r.ExpiredIdeas(ctx, t)

	w.l.With(
		"t", t,
		"ret_cursor", c,
		"err", err,
	).Debug("ExpiredIdeas")

	return c, err
}

func (w *Wrapper) DeleteExpiredIdeas(ctx context.Context, t time.Time) error {
	err := w.r.DeleteExpiredIdeas(ctx, t)

	w.l.With(
		"t", t,
		"err", err,
	).Debug("DeleteExpiredIdeas")

	return err
}
