// Package zaplog provides zap log wrappers for adam.
package zaplog

import (
	"log"

	"github.com/mavolin/adam/pkg/bot"
	"github.com/mavolin/adam/pkg/plugin"
	jww "github.com/spf13/jwalterweatherman"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Init initializes the global zap logger.
func Init(debug bool) {
	if debug {
		l, err := zap.NewDevelopment()
		if err != nil {
			panic(err)
		}

		zap.ReplaceGlobals(l)
	} else {
		l, err := zap.NewProduction()
		if err != nil {
			panic(err)
		}

		zap.ReplaceGlobals(l)
	}

	jww.TRACE = mustStdLogAt(zap.L(), zapcore.DebugLevel)
	jww.DEBUG = mustStdLogAt(zap.L(), zapcore.DebugLevel)
	jww.INFO = mustStdLogAt(zap.L(), zapcore.InfoLevel)
	jww.WARN = mustStdLogAt(zap.L(), zapcore.WarnLevel)
	jww.ERROR = mustStdLogAt(zap.L(), zapcore.ErrorLevel)
	jww.CRITICAL = mustStdLogAt(zap.L(), zapcore.ErrorLevel)
	jww.FATAL = mustStdLogAt(zap.L(), zapcore.FatalLevel)
	jww.LOG = mustStdLogAt(zap.L(), zapcore.InfoLevel)
}

func mustStdLogAt(l *zap.Logger, lvl zapcore.Level) *log.Logger {
	stdl, err := zap.NewStdLogAt(l, lvl)
	if err != nil {
		zap.S().Named("config").Fatal(err)
	}

	return stdl
}

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
