package errhandler

import (
	"github.com/getsentry/sentry-go"
	"github.com/mavolin/adam/pkg/errors"
	"go.uber.org/zap"
)

type Handler struct {
	l *zap.SugaredLogger
	h *sentry.Hub
}

func NewHandler(name string, l *zap.SugaredLogger, h *sentry.Hub) *Handler {
	return &Handler{l: l.Named(name), h: h.Clone()}
}

func (h *Handler) Capture(err error) {
	if err == nil {
		return
	}

	err = errors.WithStack(err)

	h.l.Error(err)
	h.h.CaptureException(err)
}

func (h *Handler) Do(f func() error) {
	h.Capture(f())
}
