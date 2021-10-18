package setup

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/mavolin/stormy/pkg/repository"
	"github.com/mavolin/stormy/pkg/repository/logwrap"
	"github.com/mavolin/stormy/pkg/repository/memory"
)

type RepositoryOptions struct {
	Logger *zap.SugaredLogger
}

func Repository(o RepositoryOptions) repository.Repository {
	return logWrap(memory.New(), o.Logger.Named("repository"))
}

func logWrap(r repository.Repository, l *zap.SugaredLogger) repository.Repository {
	if !l.Desugar().Core().Enabled(zapcore.DebugLevel) {
		return r
	}

	return logwrap.Wrap(r, l)
}
