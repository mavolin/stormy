// Package deleteall provides the deleteall middleware
package deleteall

import (
	"sync"
	"time"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/mavolin/adam/pkg/bot"
	"github.com/mavolin/adam/pkg/errors"
	"github.com/mavolin/adam/pkg/impl/replier"
	"github.com/mavolin/adam/pkg/plugin"
	"github.com/mavolin/disstate/v4/pkg/event"
	"github.com/mavolin/disstate/v4/pkg/state"
)

func NewMiddleware(delay time.Duration) bot.Middleware {
	return func(next bot.CommandFunc) bot.CommandFunc {
		return func(s *state.State, ctx *plugin.Context) error {
			userMsgIDs := []discord.MessageID{ctx.ID}
			var userMsgIDsMutex sync.Mutex

			rm := s.AddHandler(func(_ *state.State, e *event.MessageCreate) {
				if e.ChannelID != ctx.ChannelID || e.Author.ID != ctx.Author.ID {
					return
				}

				userMsgIDsMutex.Lock()
				userMsgIDs = append(userMsgIDs, e.ID)
				userMsgIDsMutex.Unlock()
			})
			defer rm()

			t := replier.NewTracker(ctx.Replier)
			ctx.Replier = t

			if err := next(s, ctx); err != nil {
				return err
			}

			rm()

			time.Sleep(delay)

			userMsgIDsMutex.Lock()
			defer userMsgIDsMutex.Unlock()

			guildMessages := t.GuildMessages()

			ids := make([]discord.MessageID, 0, len(guildMessages)+len(userMsgIDs))
			for _, msg := range guildMessages {
				ids = append(ids, msg.ID)
			}

			ids = append(ids, userMsgIDs...)

			return errors.WithStack(s.DeleteMessages(ctx.ChannelID, ids, ""))
		}
	}
}
