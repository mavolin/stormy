// Package sentrystate provides utilities to use sentry with disstate.
package sentrystate

import (
	"context"
	"reflect"
	"strings"

	"github.com/getsentry/sentry-go"
	"github.com/mavolin/adam/pkg/plugin"
	"github.com/mavolin/disstate/v4/pkg/event"
	"github.com/mavolin/disstate/v4/pkg/state"
)

type spanKey struct{}

var SpanKey = new(spanKey)

// Hub retrieves the *sentry.Hub from the passed *event.Base.
// In order for Hub to return with a non-nil *sentry.Hub, a hub must have
// previously been stored under the HubKey by a middleware such as the one
// returned by NewMiddleware.
func Hub(b *event.Base) *sentry.Hub {
	if hub := b.Get(sentry.HubContextKey); hub != nil {
		if hub, ok := hub.(*sentry.Hub); ok && hub != nil {
			return hub
		}
	}

	return nil
}

// Transaction retrieves the *sentry.Span from the passed *event.Base.
// In order for Transaction to return with a non-nil *sentry.Span, a span must
// have previously been stored under the SpanKey by a middleware such as the
// one returned by NewMiddleware.
func Transaction(b *event.Base) *sentry.Span {
	if span := b.Get(SpanKey); span != nil {
		if span, ok := span.(*sentry.Span); ok && span != nil {
			return span
		}
	}

	return nil
}

type HandlerMeta struct {
	Hub            *sentry.Hub
	PluginProvider string
	PluginID       plugin.ID
	Operation      string
	Trace          bool
}

// NewMiddleware creates a new middleware that attaches a *sentry.Hub to the
// event's base.
//
// Optionally, if Trace is set to true, NewMiddleware will also start a Span that
// must be finished by the handler by calling Transaction(event.Base).Finish at
// the end of the function.
func NewMiddleware(m HandlerMeta) func(*state.State, interface{}) {
	var transactionBuilder strings.Builder
	transactionBuilder.Grow(
		len(m.PluginProvider) + len("/") + len(m.PluginID) + len("/") + len(m.Operation),
	)

	if m.PluginProvider != "" {
		transactionBuilder.WriteString(m.PluginProvider)
		transactionBuilder.WriteRune('/')
	}

	if m.PluginID != "" {
		transactionBuilder.WriteString(string(m.PluginID))
		transactionBuilder.WriteRune('/')
	}

	transactionBuilder.WriteString(m.Operation)

	transactionName := transactionBuilder.String()

	return func(_ *state.State, e interface{}) {
		b := reflect.ValueOf(e).FieldByName("Base").Interface().(*event.Base)

		h := m.Hub.Clone()

		h.Scope().SetTransaction(transactionName)
		h.Scope().SetExtra("event", e)

		b.Set(sentry.HubContextKey, h)

		if m.Trace {
			ctx := sentry.SetHubOnContext(context.Background(), h)
			span := sentry.StartSpan(ctx, m.Operation)
			b.Set(SpanKey, span)
		}
	}
}
