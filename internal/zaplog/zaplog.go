// Package zaplog provides zap log wrappers for adam.
package zaplog

import (
	"github.com/mavolin/adam/pkg/bot"
	"github.com/mavolin/adam/pkg/plugin"
	"go.uber.org/zap"
)

// CommandError returns the logger function used for errors.Log.
func CommandError(l *zap.SugaredLogger) func(error, *plugin.Context) {
	return func(err error, ctx *plugin.Context) {
		l.With("ident", ctx.InvokedCommand.Identifier).
			Error(err)
	}
}

// GatewayLog returns the logger function used for the
// bot.Options.GatwayErrorHandler.
func GatewayLog(l *zap.SugaredLogger) func(error) {
	return func(err error) {
		if bot.FilterGatewayError(err) {
			l.Error(err)
		}
	}
}

// StateError returns the logger function used for the
// bot.Option.StateErrorHandler.
func StateError(l *zap.SugaredLogger) func(error) {
	return func(err error) { l.Error(err) }
}

// StatePanic returns the logger function used for the
// bot.Option.StatePanicHandler.
func StatePanic(l *zap.SugaredLogger) func(interface{}) {
	return func(rec interface{}) {
		l.Errorf("recovered from panic: %+v", rec)
	}
}
