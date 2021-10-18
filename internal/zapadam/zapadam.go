package zapadam

import (
	"github.com/mavolin/adam/pkg/bot"
	"github.com/mavolin/adam/pkg/plugin"
	"github.com/mavolin/disstate/v4/pkg/state"
	"go.uber.org/zap"
)

type key struct{}

// Get returns the *zap.SugaredLogger previously stored in the context using
// the middleware returned by NewMiddleware.
func Get(ctx *plugin.Context) *zap.SugaredLogger {
	if l := ctx.Get(key{}); l != nil {
		if l, ok := l.(*zap.SugaredLogger); ok && l != nil {
			return l
		}
	}

	return zap.NewNop().Sugar()
}

// NewMiddleware returns a bot.Middleware that logs the invocation of the command
// and then saves a logger named "cmd:"+ctx.InvokedCommand.SourceName()+"/"+
// ctx.InvokedCommand.ID()[1:] in the context.
func NewMiddleware(l *zap.SugaredLogger) bot.Middleware {
	return func(next bot.CommandFunc) bot.CommandFunc {
		return func(s *state.State, ctx *plugin.Context) error {
			l := l.With(
				"source", ctx.InvokedCommand.SourceName(),
				"command_id", ctx.InvokedCommand.ID(),
				"guild_id", ctx.GuildID,
				"channel_id", ctx.ChannelID,
				"message_id", ctx.Message.ID,
				"author_id", ctx.Author.ID,
			)
			l.Named("router").Info("command was invoked")

			ctx.Set(key{}, l.Named("cmd:"+ctx.InvokedCommand.SourceName()+"/"+string(ctx.InvokedCommand.ID()[1:])))

			return next(s, ctx)
		}
	}
}

// NewFallbackMiddleware creates a new middleware that adds a fallback logger
// to be used to log errors that occur during routing.
func NewFallbackMiddleware(l *zap.SugaredLogger) bot.Middleware {
	return func(next bot.CommandFunc) bot.CommandFunc {
		return func(s *state.State, ctx *plugin.Context) error {
			l := l.Named("router").With(
				"guild_id", ctx.GuildID,
				"channel_id", ctx.ChannelID,
				"message_id", ctx.Message.ID,
				"author_id", ctx.Author.ID,
			)
			ctx.Set(key{}, l)

			return next(s, ctx)
		}
	}
}
