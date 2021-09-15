package zapstate

import (
	"github.com/mavolin/disstate/v4/pkg/event"
	"github.com/mavolin/disstate/v4/pkg/state"
	"go.uber.org/zap"
)

type key struct{}

// Get returns the *zap.SugaredLogger previously stored in the base using
// the middleware returned by NewMiddleware.
func Get(b *event.Base) *zap.SugaredLogger {
	if l := b.Get(key{}); l != nil {
		if l, ok := l.(*zap.SugaredLogger); ok && l != nil {
			return l
		}
	}

	return nil
}

// NewMiddleware returns a bot.Middleware that adds the passed logger to the
// event's base.
func NewMiddleware(l *zap.SugaredLogger) func(*state.State, *event.Base) {
	return func(_ *state.State, base *event.Base) {
		base.Set(key{}, l)
	}
}
