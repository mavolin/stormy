package errhandler

import (
	"github.com/getsentry/sentry-go"
	"github.com/mavolin/adam/pkg/bot"
	"github.com/mavolin/adam/pkg/errors"
	"github.com/mavolin/adam/pkg/plugin"
	"go.uber.org/zap"

	"github.com/mavolin/stormy/internal/sentryadam"
	"github.com/mavolin/stormy/internal/zapadam"
)

// ErrorLog returns the logger function used for errors.Log.
// It logs the error using the passed *zap.SugaredLogger, and extracts the
// assigned *sentry.Hub and *zap.SugaredLogger from the context using sentryadam.Get.
func ErrorLog(err error, ctx *plugin.Context) {
	sentryadam.Hub(ctx).CaptureException(err)

	l := zapadam.Get(ctx).With("err", err)

	if serr, ok := err.(interface{ StackTrace() errors.StackTrace }); ok {
		l.With("stack_trace", serr.StackTrace().String())
	}

	l.Error("error during command execution")
}

// NewGateway returns the error handler function used for the
// bot.Options.GatwayErrorHandler.
func NewGateway(l *zap.SugaredLogger, h *sentry.Hub) func(error) {
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

// NewStateError returns the logger function used for the
// bot.Option.StateErrorHandler.
func NewStateError(l *zap.SugaredLogger, h *sentry.Hub) func(error) {
	l = l.Named("state")

	h = h.Clone()
	h.Scope().SetTransaction("state")

	return func(err error) {
		h.CaptureException(err)
		l.Error(err)
	}
}

// NewStatePanic returns the logger function used for the
// bot.Option.StatePanicHandler.
func NewStatePanic(l *zap.SugaredLogger, h *sentry.Hub) func(interface{}) {
	l = l.Named("state")

	h = h.Clone()
	h.Scope().SetTransaction("state")

	return func(rec interface{}) {
		h.Recover(rec)
		l.Errorf("recovered from panic: %+v", rec)
	}
}
