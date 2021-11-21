package memory

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/diamondburned/arikawa/v3/discord"

	idearepo "github.com/mavolin/stormy/modules/idea/repository"
)

type ideaRepo struct {
	ideas map[discord.MessageID]idearepo.Idea
	mu    sync.RWMutex
}

var _ idearepo.IdeaRepository = (*ideaRepo)(nil)

func newIdeaRepo() *ideaRepo {
	return &ideaRepo{ideas: make(map[discord.MessageID]idearepo.Idea)}
}

func (r *ideaRepo) Idea(messageID discord.MessageID) (*idearepo.Idea, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	i, ok := r.ideas[messageID]
	if !ok {
		return nil, nil
	}

	return &i, nil
}

func (r *ideaRepo) SaveIdea(idea *idearepo.Idea) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.ideas[idea.MessageID] = *idea
	return nil
}

func (r *ideaRepo) DeleteIdea(messageID discord.MessageID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.ideas, messageID)
	return nil
}

func (r *ideaRepo) ExpiringIdeas(afterT time.Time, afterID discord.MessageID, limit int) ([]*idearepo.Idea, error) {
	ideas := make([]*idearepo.Idea, 0, limit)

	r.mu.Lock()
	defer r.mu.Unlock()

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

func (r *ideaRepo) ExpiredIdeas(before time.Time) (idearepo.IdeaCursor, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

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

func (r *ideaRepo) DeleteExpiredIdeas(before time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()

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
