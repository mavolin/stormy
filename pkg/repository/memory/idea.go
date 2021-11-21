package memory

import (
	"context"
	"sort"
	"time"

	"github.com/diamondburned/arikawa/v3/discord"

	idearepo "github.com/mavolin/stormy/modules/idea/repository"
)

func (r *Repository) IdeaChannelSettings(channelID discord.ChannelID) (*idearepo.ChannelSettings, error) {
	r.ideaChannelSettingMu.RLock()
	defer r.ideaChannelSettingMu.RUnlock()

	s, ok := r.ideaChannelSettings[channelID]
	if !ok {
		return nil, nil
	}

	return &s, nil
}

func (r *Repository) SetIdeaChannelSettings(channelID discord.ChannelID, s idearepo.ChannelSettings) error {
	r.ideaChannelSettingMu.Lock()
	defer r.ideaChannelSettingMu.Unlock()

	r.ideaChannelSettings[channelID] = s
	return nil
}

func (r *Repository) DisableIdeaChannel(channelID discord.ChannelID) error {
	r.ideaChannelSettingMu.Lock()
	defer r.ideaChannelSettingMu.Unlock()

	delete(r.ideaChannelSettings, channelID)
	return idearepo.ErrChannelAlreadyDisabled
}

func (r *Repository) Idea(messageID discord.MessageID) (*idearepo.Idea, error) {
	r.ideaMu.RLock()
	defer r.ideaMu.RUnlock()

	i, ok := r.ideas[messageID]
	if !ok {
		return nil, nil
	}

	return &i, nil
}

func (r *Repository) SaveIdea(idea *idearepo.Idea) error {
	r.ideaMu.Lock()
	defer r.ideaMu.Unlock()

	r.ideas[idea.MessageID] = *idea
	return nil
}

func (r *Repository) DeleteIdea(messageID discord.MessageID) error {
	r.ideaMu.Lock()
	defer r.ideaMu.Unlock()

	delete(r.ideas, messageID)
	return nil
}

func (r *Repository) ExpiringIdeas(afterT time.Time, afterID discord.MessageID, limit int) ([]*idearepo.Idea, error) {
	ideas := make([]*idearepo.Idea, 0, limit)

	r.ideaMu.Lock()
	defer r.ideaMu.Unlock()

	for _, idea := range r.ideas {
		if idea.VoteUntil == nil {
			continue
		} else if !idea.VoteUntil.After(afterT) && !(idea.VoteUntil.Equal(afterT) && idea.MessageID > afterID) {
			continue
		}

		idea := idea

		i := sort.Search(len(ideas), func(i int) bool {
			cmp := ideas[i]
			return cmp.VoteUntil.After(*idea.VoteUntil) ||
				(cmp.VoteUntil.Equal(*idea.VoteUntil) && cmp.MessageID >= idea.MessageID)
		})
		if i > len(ideas) {
			if i < limit {
				ideas = append(ideas, &idea)
			}
		} else {
			if len(ideas) < limit {
				ideas = append(ideas, &idea)
			}

			copy(ideas[i+1:], ideas[i:])
			ideas[i] = &idea
		}
	}

	if len(ideas) == 0 {
		return nil, nil
	}

	return ideas, nil
}

func (r *Repository) ExpiredIdeas(before time.Time) (idearepo.IdeaCursor, error) {
	r.ideaMu.RLock()
	defer r.ideaMu.RUnlock()

	var ideas []idearepo.Idea
	for _, idea := range r.ideas {
		if idea.VoteUntil != nil && idea.VoteUntil.Before(before) {
			ideas = append(ideas, idea)
		}
	}

	if len(ideas) == 0 {
		return nil, nil
	}

	return newIdeaCursor(ideas), nil
}

func (r *Repository) DeleteExpiredIdeas(before time.Time) error {
	r.ideaMu.Lock()
	defer r.ideaMu.Unlock()

	for messageID, idea := range r.ideas {
		if idea.VoteUntil != nil && idea.VoteUntil.Before(before) {
			delete(r.ideas, messageID)
		}
	}

	return nil
}

// =============================================================================
// Utils
// =====================================================================================

type ideaCursor struct {
	ideas []idearepo.Idea
	i     int
}

func newIdeaCursor(ideas []idearepo.Idea) *ideaCursor {
	return &ideaCursor{ideas: ideas}
}

var _ idearepo.IdeaCursor = (*ideaCursor)(nil)

func (i *ideaCursor) BatchLength() int {
	return len(i.ideas)
}

func (i *ideaCursor) Next(context.Context) (*idearepo.Idea, error) {
	if i.i >= len(i.ideas) {
		return nil, nil
	}

	i.i++
	return &i.ideas[i.i-1], nil
}

func (i *ideaCursor) Close(context.Context) error {
	return nil
}
