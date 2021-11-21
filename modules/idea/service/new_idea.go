package service

import (
	"context"
	"time"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/mavolin/disstate/v4/pkg/event"
	"github.com/mavolin/disstate/v4/pkg/state"
	"github.com/mavolin/sentryadam/pkg/sentrystate"

	"github.com/mavolin/stormy/internal/errhandler"
	"github.com/mavolin/stormy/modules/idea/service/format"
)

func (service *Service) onNewIdea(s *state.State, e *event.MessageCreate) error {
	if e.Author.Bot || e.GuildID == 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	set, err := service.repo.IdeaChannelSettings(ctx, e.ChannelID)
	if err != nil || set == nil {
		return err
	}

	service.log.With(
		"guild_id", e.GuildID,
		"channel_id", e.ChannelID,
		"message_id", e.ID,
	).Debug("new idea")

	defer sentrystate.Transaction(e.Base).Finish()

	if err := service.ensurePermissions(s, e); err != nil {
		return err
	}

	i, err := format.Idea(e, set)
	if err != nil {
		return err
	}

	msg, err := s.SendEmbeds(e.ChannelID, *i.Embed)
	if err != nil {
		return errhandler.NewInternalError(e, err)
	}

	i.RepoIdea.GuildID = msg.GuildID
	i.RepoIdea.MessageID = msg.ID
	i.RepoIdea.ChannelID = msg.ChannelID

	ctx, cancel = context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err = service.saveIdea(ctx, i.RepoIdea); err != nil {
		if set.VoteDuration > 0 {
			return errhandler.NewInternalErrorWithDescription(e, err, "I couldn't save this idea to my database. "+
				"This means I can't enforce mutually exclusive votes in groups, "+
				"and I won't be able to publish voting results.")
		}

		return errhandler.NewInternalErrorWithDescription(e, err, "I couldn't save this idea to my database. "+
			"This means I can't enforce mutually exclusive votes in groups.")
	}

	// Deleting the original message is not crucial, so don't return on error
	err = s.DeleteMessage(e.ChannelID, e.ID, "Created new idea")
	service.errhandler.Capture(err)

	for _, emoji := range set.VoteType.Emojis() {
		if err := s.React(msg.ChannelID, msg.ID, discord.APIEmoji(emoji)); err != nil {
			return errhandler.NewInternalError(e, err)
		}
	}

	for _, emoji := range i.RepoIdea.GlobalSectionEmojis {
		if err := s.React(msg.ChannelID, msg.ID, discord.APIEmoji(emoji)); err != nil {
			return errhandler.NewInternalError(e, err)
		}
	}

	for _, g := range i.RepoIdea.Groups {
		for _, emoji := range g.Emojis {
			if err := s.React(msg.ChannelID, msg.ID, discord.APIEmoji(emoji)); err != nil {
				return errhandler.NewInternalError(e, err)
			}
		}
	}

	return nil
}

func (service *Service) ensurePermissions(s *state.State, e *event.MessageCreate) error {
	perms, err := s.Permissions(e.ChannelID, service.selfID)
	if err != nil {
		return errhandler.NewInternalError(e, err)
	}

	switch {
	case perms.Has(discord.PermissionAdministrator):
		return nil
	case !perms.Has(discord.PermissionSendMessages):
		return errhandler.Abort
	case !perms.Has(discord.PermissionAddReactions):
		return errhandler.NewInfo(e, "I need permissions to add reactions to post ideas in this channel.")
	case !perms.Has(discord.PermissionManageMessages):
		return errhandler.NewInfo(e, "I need permissions to manage messages to post ideas in this channel.")
	}

	return nil
}
