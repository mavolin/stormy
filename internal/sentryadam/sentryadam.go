// Package sentryadam provides utilities to use sentry with adam.
package sentryadam

import (
	"context"
	"strconv"

	"github.com/getsentry/sentry-go"
	"github.com/mavolin/adam/pkg/bot"
	"github.com/mavolin/adam/pkg/errors"
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

// NewPreRouteMiddleware is the first middleware to be added to the bot.
// It adds a *sentry.Hub to the *plugin.Context and starts a transaction.
//
// The bot's NoDefaultMiddlewares option must be enabled, and the default
// middleware must manually be added after the middleware returned by
// NewPreRouteMiddleware.
func NewPreRouteMiddleware(h *sentry.Hub) bot.Middleware {
	return func(next bot.CommandFunc) bot.CommandFunc {
		return func(s *state.State, ctx *plugin.Context) error {
			h := h.Clone()

			shardID := 0
			if ctx.GuildID != 0 {
				shardID = s.FromGuildID(ctx.GuildID).Identifier.Shard.ShardID()
			}

			h.Scope().SetTags(map[string]string{
				"channel_id": ctx.ChannelID.String(),
				"guild_id":   ctx.GuildID.String(),
				"shard_id":   strconv.Itoa(shardID),
			})
			h.Scope().SetExtra("message", ctx.Message)
			h.Scope().SetUser(sentry.User{
				ID:       ctx.Author.ID.String(),
				Username: ctx.User.Username,
			})

			ctx.Set(sentry.HubContextKey, h)

			spanCtx := sentry.SetHubOnContext(context.Background(), h)

			span := sentry.StartSpan(spanCtx, "cmd")
			defer span.Finish()

			ctx.Set(parentSpanKey, span)

			routeSpan := span.StartChild("route")
			ctx.Set(routeSpanKey, routeSpan)

			err := next(s, ctx)

			deriveSpanStatus(span, err)

			return err
		}
	}
}

// PostRouteMiddleware is the middleware that must be added immediately after
// the bot's default middlewares.
func PostRouteMiddleware(next bot.CommandFunc) bot.CommandFunc {
	return func(s *state.State, ctx *plugin.Context) error {
		h := Hub(ctx)
		h.Scope().SetTags(map[string]string{
			"plugin_source": ctx.InvokedCommand.SourceName(),
			"command_id":    string(ctx.InvokedCommand.ID()),
		})
		h.Scope().SetTransaction(ctx.InvokedCommand.SourceName() + "/" + string(ctx.InvokedCommand.ID()[1:]))

		ctx.Get(routeSpanKey).(*sentry.Span).Finish()

		parentSpan := ctx.Get(parentSpanKey).(*sentry.Span)

		middlewaresSpan := parentSpan.StartChild("middlewares")
		ctx.Set(middlewaresSpanKey, middlewaresSpan)

		return next(s, ctx)
	}
}

// PreInvokeMiddleware is the post middleware that must be added last.
func PreInvokeMiddleware(next bot.CommandFunc) bot.CommandFunc {
	return func(s *state.State, ctx *plugin.Context) error {
		ctx.Get(middlewaresSpanKey).(*sentry.Span).Finish()

		parentSpan := ctx.Get(parentSpanKey).(*sentry.Span)

		invokeSpan := parentSpan.StartChild("invoke")
		defer invokeSpan.Finish()

		ctx.Set(sentrystate.SpanKey, invokeSpan)

		return next(s, ctx)
	}
}

func deriveSpanStatus(span *sentry.Span, err error) {
	switch {
	case errors.Is(err, errors.Abort):
		span.Status = sentry.SpanStatusAborted
	case errors.Is(err, bot.ErrUnknownCommand):
		span.Status = sentry.SpanStatusNotFound
	case errors.As(err, new(*errors.InformationalError)):
		span.Status = sentry.SpanStatusCanceled
	case errors.As(err, new(*errors.UserError)), errors.As(err, new(*errors.UserInfo)):
		span.Status = sentry.SpanStatusFailedPrecondition
	case errors.As(err, new(*plugin.ArgumentError)):
		span.Status = sentry.SpanStatusInvalidArgument
	case errors.As(err, new(*plugin.BotPermissionsError)):
		span.Status = sentry.SpanStatusFailedPrecondition
	case errors.As(err, new(*plugin.ChannelTypeError)):
		span.Status = sentry.SpanStatusFailedPrecondition
	case errors.As(err, new(*plugin.RestrictionError)):
		span.Status = sentry.SpanStatusPermissionDenied
	case errors.As(err, new(*plugin.ThrottlingError)):
		span.Status = sentry.SpanStatusResourceExhausted
	case errors.As(err, new(errors.Error)):
		span.Status = sentry.SpanStatusUndefined
	case errors.As(err, new(*errors.InternalError)), err != nil:
		span.Status = sentry.SpanStatusInternalError
	default:
		span.Status = sentry.SpanStatusOK
	}
}
