package setup

import (
	"go.uber.org/zap"

	"github.com/mavolin/stormy/pkg/repository"
	"github.com/mavolin/stormy/pkg/repository/memory"
)

type RepositoryOptions struct {
	Logger *zap.SugaredLogger
}

func Repository(o RepositoryOptions) repository.Repository {
	logger := o.Logger.Named("repository")

	return memory.New(logger)
}
