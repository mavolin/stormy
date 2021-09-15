// Package sentrystate provides utilities to use sentry with disstate.
package sentrystate

import (
	"github.com/getsentry/sentry-go"
	"github.com/mavolin/disstate/v4/pkg/event"
	"github.com/mavolin/disstate/v4/pkg/state"
)

type key struct{}

// Key is the key used to set and retrieve *sentry.Hubs stored by middlewares
// in an event's event.Base.
var Key = new(key)

// Get retrieves the *sentry.Hub from the passed *event.Base.
// In order to return with a non-nil *sentry.Hub, a hub must have previously
// been added by one of the middlewares returned by MessageCreateMiddleware
// and MessageUpdateMiddleware.
func Get(b *event.Base) *sentry.Hub {
	if hub := b.Get(Key); hub != nil {
		if hub, ok := hub.(*sentry.Hub); ok && hub != nil {
			return hub
		}
	}

	return nil
}

// MessageCreateMiddleware creates a new middleware for the MessageCreate
// event.
func MessageCreateMiddleware(h *sentry.Hub, source string) func(*state.State, *event.MessageCreate) {
	return func(_ *state.State, e *event.MessageCreate) {
		h := h.Clone()
		h.Scope().SetTag("source", source)
		h.Scope().SetTags(map[string]string{
			"channel_id": e.ChannelID.String(),
			"author_id":  e.Author.ID.String(),
			"guild_id":   e.GuildID.String(),
		})
		h.Scope().SetExtra("message", e.Message)

		e.Set(Key, h)
	}
}

// MessageUpdateMiddleware creates a new middleware for the MessageUpdate
// event.
func MessageUpdateMiddleware(h *sentry.Hub, source string) func(*state.State, *event.MessageUpdate) {
	return func(_ *state.State, e *event.MessageUpdate) {
		h := h.Clone()
		h.Scope().SetTag("source", source)
		h.Scope().SetTags(map[string]string{
			"channel_id": e.ChannelID.String(),
			"author_id":  e.Author.ID.String(),
			"guild_id":   e.GuildID.String(),
		})
		h.Scope().SetExtra("message", e.Message)

		e.Set(Key, h)
	}
}
