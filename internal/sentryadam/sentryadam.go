// Package sentryadam provides utilities to use sentry with adam.
package sentryadam

import (
	"context"

	"github.com/getsentry/sentry-go"
	"github.com/mavolin/adam/pkg/bot"
	"github.com/mavolin/adam/pkg/plugin"
	"github.com/mavolin/disstate/v4/pkg/state"

	"github.com/mavolin/stormy/internal/sentrystate"
)

// Hub retrieves the *sentry.Hub from the passed *event.Base.
// In order for Hub to return with a non-nil *sentry.Hub, a hub must have
// previously been stored under the HubKey by a middleware, such as the
// three middlewares provided by sentryadam.
func Hub(ctx *plugin.Context) *sentry.Hub {
	return sentrystate.Hub(ctx.Base)
}

// Transaction retrieves the *sentry.Span from the passed *event.Base.
// In order for Transaction to return with a non-nil *sentry.Span, a span must
// have previously been stored under the SpanKey by a middleware, such as the
// three middlewares provided by sentryadam.
func Transaction(ctx *plugin.Context) *sentry.Span {
	return sentrystate.Transaction(ctx.Base)
}

type spanKey uint8

const (
	parentSpanKey spanKey = iota
	routeSpanKey
	middlewaresSpanKey
)

// PreRouteMiddleware is the first middleware to be added to the bot.
// It adds a *sentry.Hub to the *plugin.Context and starts a transaction.
//
// The bot's NoDefaultMiddlewares option must be enabled, and the default
// middleware must manually be added after the middleware returned by
// PreRouteMiddleware.
func PreRouteMiddleware(h *sentry.Hub) bot.Middleware {
	return func(next bot.CommandFunc) bot.CommandFunc {
		return func(s *state.State, ctx *plugin.Context) error {
			h := h.Clone()

			h.Scope().SetTags(map[string]string{
				"channel_id": ctx.ChannelID.String(),
				"author_id":  ctx.Author.ID.String(),
				"guild_id":   ctx.GuildID.String(),
			})
			h.Scope().SetExtra("message", ctx.Message)

			ctx.Set(sentrystate.HubKey, h)

			spanCtx := sentry.SetHubOnContext(context.Background(), h)

			span := sentry.StartSpan(spanCtx, "cmd")
			defer span.Finish()

			ctx.Set(parentSpanKey, span)

			routeSpan := span.StartChild("route")
			ctx.Set(routeSpanKey, routeSpan)

			return next(s, ctx)
		}
	}
}

// PostRouteMiddleware is the middleware that must be added immediately after
// the bot's default middlewares.
func PostRouteMiddleware() bot.Middleware {
	return func(next bot.CommandFunc) bot.CommandFunc {
		return func(s *state.State, ctx *plugin.Context) error {
			h := Hub(ctx)
			h.Scope().SetTags(map[string]string{
				"plugin_source": ctx.InvokedCommand.SourceName(),
				"command_id":    string(ctx.InvokedCommand.ID()),
			})

			routeSpan := ctx.Get(routeSpanKey).(*sentry.Span)
			routeSpan.Finish()

			parentSpan := ctx.Get(parentSpanKey).(*sentry.Span)

			middlewaresSpan := parentSpan.StartChild("middlewares")
			ctx.Set(middlewaresSpanKey, middlewaresSpan)

			return next(s, ctx)
		}
	}
}

// PreInvokeMiddleware is the post middleware that must be added last.
func PreInvokeMiddleware() bot.Middleware {
	return func(next bot.CommandFunc) bot.CommandFunc {
		return func(s *state.State, ctx *plugin.Context) error {
			middlewaresSpan := ctx.Get(middlewaresSpanKey).(*sentry.Span)
			middlewaresSpan.Finish()

			parentSpan := ctx.Get(parentSpanKey).(*sentry.Span)

			invokeSpan := parentSpan.StartChild("invoke")
			defer invokeSpan.Finish()

			ctx.Set(sentrystate.SpanKey, invokeSpan)

			return next(s, ctx)
		}
	}
}
