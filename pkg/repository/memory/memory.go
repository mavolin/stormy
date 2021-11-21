// Package memory provides an in-memory repository.
package memory

import (
	"sync"

	"github.com/diamondburned/arikawa/v3/discord"

	idearepo "github.com/mavolin/stormy/modules/idea/repository"
	"github.com/mavolin/stormy/pkg/repository"
)

type Repository struct {
	ideaChannelSettings  map[discord.ChannelID]idearepo.ChannelSettings
	ideaChannelSettingMu sync.RWMutex

	ideas  map[discord.MessageID]idearepo.Idea
	ideaMu sync.RWMutex
}

var _ repository.Repository = new(Repository)

func New() *Repository {
	return &Repository{
		ideaChannelSettings: make(map[discord.ChannelID]idearepo.ChannelSettings),
		ideas:               make(map[discord.MessageID]idearepo.Idea),
	}
}
