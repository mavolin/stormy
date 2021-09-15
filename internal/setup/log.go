package setup

import (
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type LoggerOptions struct {
	Debug bool
}

// Logger sets up the logger.
func Logger(o LoggerOptions) (*zap.SugaredLogger, error) {
	core, err := newCore(o)
	if err != nil {
		return nil, err
	}

	return zap.New(core, zapOptions(o)...).Sugar(), nil
}

func newCore(o LoggerOptions) (zapcore.Core, error) {
	lvl := zap.InfoLevel
	if o.Debug {
		lvl = zap.DebugLevel
	}

	enc := newJSONEncoder()
	if o.Debug {
		enc = newConsoleEncoder()
	}

	sink, _, err := zap.Open("stderr")
	if err != nil {
		return nil, err
	}

	core := zapcore.NewCore(enc, sink, zap.NewAtomicLevelAt(lvl))
	if !o.Debug {
		core = zapcore.NewSamplerWithOptions(core, time.Second, 100, 100)
	}

	return core, nil
}

func zapOptions(o LoggerOptions) (zopts []zap.Option) {
	if o.Debug {
		zopts = append(zopts, zap.Development(), zap.AddStacktrace(zapcore.WarnLevel))
	} else {
		zopts = append(zopts, zap.AddStacktrace(zapcore.ErrorLevel))
	}

	return append(zopts, zap.AddCaller())
}

func newConsoleEncoder() zapcore.Encoder {
	return zapcore.NewConsoleEncoder(zapcore.EncoderConfig{
		// console ignores the keys and simply checks if they are non-empty
		MessageKey:     ".",
		LevelKey:       ".",
		TimeKey:        ".",
		NameKey:        ".",
		CallerKey:      ".",
		FunctionKey:    ".",
		StacktraceKey:  ".",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseColorLevelEncoder,
		EncodeTime:     zapcore.RFC3339TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	})
}

func newJSONEncoder() zapcore.Encoder {
	return zapcore.NewJSONEncoder(zapcore.EncoderConfig{
		MessageKey:     "msg",
		LevelKey:       "lvl",
		TimeKey:        "time",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		StacktraceKey:  "stack_trace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.RFC3339TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	})
}
