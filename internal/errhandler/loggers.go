package errhandler

import (
	"github.com/getsentry/sentry-go"
	"github.com/mavolin/adam/pkg/bot"
	"github.com/mavolin/adam/pkg/errors"
	"github.com/mavolin/adam/pkg/plugin"
	"github.com/mavolin/adam/pkg/utils/discorderr"
	"github.com/mavolin/disstate/v4/pkg/state"
	"github.com/mavolin/sentryadam/pkg/sentryadam"
	"go.uber.org/zap"

	"github.com/mavolin/stormy/pkg/utils/zapadam"
)

// ErrorLog returns the logger function used for errors.Log.
// It logs the error using the passed *zap.SugaredLogger, and extracts the
// assigned *sentry.Hub and *zap.SugaredLogger from the context using
// sentryadam.Hub.
func ErrorLog(ctx *plugin.Context, err *errors.InternalError) {
	h := sentryadam.Hub(ctx)

	if derr := discorderr.As(err); derr != nil {
		h.WithScope(func(s *sentry.Scope) {
			s.SetExtra("discord_error_body", string(derr.Body))
			h.CaptureException(err)
		})
	} else {
		h.CaptureException(err)
	}

	zapadam.Get(ctx).
		With("err", err.Unwrap()).
		With("stack_trace", err.StackTrace().String()).
		Error("error during command execution")
}

// NewGatewayHandler returns the error handler function used for the
// bot.Options.GatewayErrorHandler.
func NewGatewayHandler(l *zap.SugaredLogger, h *sentry.Hub) func(error) {
	l = l.Named("gateway")

	h = h.Clone()
	h.Scope().SetTransaction("gateway")

	return func(err error) {
		if !bot.FilterGatewayError(err) {
			h.CaptureException(err)
			l.Error(err)
		}
	}
}

// NewStateErrorHandler returns the logger function used for the
// bot.Option.StateErrorHandler.
func NewStateErrorHandler(s *state.State, l *zap.SugaredLogger, h *sentry.Hub) func(error) {
	l = l.Named("state")

	h = h.Clone()
	h.Scope().SetTransaction("state")

	return func(err error) {
		if errors.Is(err, Abort) {
			return
		}

		var herr handler
		if errors.As(err, &herr) {
			newErr := herr.Handle(s)
			if newErr != nil {
				err = newErr
			} else if !errors.As(err, new(*InternalError)) {
				return
			}
		}

		if derr := discorderr.As(err); derr != nil {
			h.WithScope(func(s *sentry.Scope) {
				s.SetExtra("discord_error_body", string(derr.Body))
				h.CaptureException(err)
			})
		} else {
			h.CaptureException(err)
		}

		l.Error(err)
	}
}

// NewStatePanicHandler returns the logger function used for the
// bot.Option.StatePanicHandler.
func NewStatePanicHandler(l *zap.SugaredLogger, h *sentry.Hub) func(interface{}) {
	l = l.Named("state")
	h.Scope().SetTransaction("state")

	return func(rec interface{}) {
		h.Recover(rec)
		l.Errorf("recovered from panic: %+v", rec)
	}
}
