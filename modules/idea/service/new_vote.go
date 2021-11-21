package service

import (
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/mavolin/adam/pkg/errors"
	"github.com/mavolin/adam/pkg/i18n"
	"github.com/mavolin/adam/pkg/utils/discorderr"
	"github.com/mavolin/disstate/v4/pkg/event"
	"github.com/mavolin/disstate/v4/pkg/state"
	"github.com/mavolin/sentryadam/pkg/sentrystate"

	"github.com/mavolin/stormy/modules/idea/repository"
)

func (service *Service) onReactionAdd(s *state.State, e *event.MessageReactionAdd) error {
	if e.GuildID == 0 || e.Member.User.Bot {
		return nil
	}

	if !e.Emoji.IsUnicode() {
		return nil
	}

	return service.sf.DoAsync(e.MessageID, func() error {
		i, err := service.idea(e.MessageID)
		if err != nil || i == nil {
			return err
		}

		service.log.With(
			"guild_id", i.GuildID,
			"channel_id", i.ChannelID,
			"message_id", i.MessageID,
			"vote_until", i.VoteUntil,
		).Debug("new vote")

		defer sentrystate.Transaction(e.Base).Finish()

		if err = service.onMutexReaction(s, e, i); err != nil {
			return err
		}

		return service.onVoteReaction(s, e, i)
	})
}

func (service *Service) onMutexReaction(s *state.State, e *event.MessageReactionAdd, i *repository.Idea) error {
	var result *repository.SectionGroup

Groups:
	for _, g := range i.Groups {
		for _, emoji := range g.Emojis {
			if emoji == e.Emoji.Name {
				result = &g
				break Groups
			}
		}
	}

	// the emoji that was used does not belong to a mutex section, return
	if result == nil {
		return nil
	}

	perms, err := s.Permissions(e.ChannelID, service.selfID)
	if err != nil {
		return errors.WithStack(err)
	}

	// ensure that wwe can delete reactions
	if !perms.Has(discord.PermissionManageMessages) {
		service.mu.Lock()

		// make sure we don't spam this error, instead only send it once per
		// restart or if we got the permission and then got it taken away again
		_, ok := service.missingManageChannelPerm[e.ChannelID]
		if ok {
			service.mu.Unlock()
			return nil
		}

		service.missingManageChannelPerm[e.ChannelID] = struct{}{}

		service.mu.Unlock()

		embed := errors.NewInfoEmbed(i18n.NewFallbackLocalizer())
		embed.Description = "I need the 'Manage Messages' permission to enforce group votes.\n" +
			"Otherwise, voting results may be inaccurate."

		_, err = s.SendEmbeds(e.ChannelID, embed)
		return errors.WithStack(err)
	}

	delete(service.missingManageChannelPerm, e.ChannelID)

	// There is no (inexpensive) way to check which reactions were added by a
	// specific user.
	// The only thing we can do is don't check and simply call
	// DeleteUserReaction.

	for _, emoji := range result.Emojis {
		if e.Emoji.IsUnicode() && emoji == e.Emoji.Name {
			continue
		}

		err = s.DeleteUserReaction(e.ChannelID, e.MessageID, e.UserID, discord.APIEmoji(emoji))
		if err != nil {
			if !discorderr.Is(discorderr.As(err), discorderr.UnknownResource...) {
				return errors.WithStack(err)
			}
		}
	}

	return nil
}

func (service *Service) onVoteReaction(s *state.State, e *event.MessageReactionAdd, i *repository.Idea) error {
	voteEmojis := i.VoteType.Emojis()

	var ok bool
	for _, emoji := range voteEmojis {
		if e.Emoji.Name == emoji {
			ok = true
			break
		}
	}

	if !ok { // not a vote reaction
		return nil
	}

	for _, emoji := range voteEmojis {
		if e.Emoji.Name == emoji {
			continue
		}

		err := s.DeleteUserReaction(e.ChannelID, e.MessageID, e.UserID, discord.APIEmoji(emoji))
		if err != nil {
			if !discorderr.Is(discorderr.As(err), discorderr.UnknownResource...) {
				return errors.WithStack(err)
			}
		}
	}

	return nil
}
