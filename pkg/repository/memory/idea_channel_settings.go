package memory

import (
	"sync"

	"github.com/diamondburned/arikawa/v3/discord"

	idearepo "github.com/mavolin/stormy/modules/idea/repository"
)

type ideaChannelSettingsRepo struct {
	settings map[discord.ChannelID]idearepo.ChannelSettings
	mu       sync.RWMutex
}

var _ idearepo.ChannelSettingsRepository = (*ideaChannelSettingsRepo)(nil)

func newIdeaChannelSettingsRepo() *ideaChannelSettingsRepo {
	return &ideaChannelSettingsRepo{settings: make(map[discord.ChannelID]idearepo.ChannelSettings)}
}

func (r *ideaChannelSettingsRepo) IdeaChannelSettings(channelID discord.ChannelID) (*idearepo.ChannelSettings, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	s, ok := r.settings[channelID]
	if !ok {
		return nil, nil
	}

	return &s, nil
}

func (r *ideaChannelSettingsRepo) SetIdeaChannelSettings(
	channelID discord.ChannelID, s idearepo.ChannelSettings,
) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.settings[channelID] = s
	return nil
}

func (r *ideaChannelSettingsRepo) DisableIdeaChannel(channelID discord.ChannelID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.settings, channelID)
	return idearepo.ErrChannelAlreadyDisabled
}
