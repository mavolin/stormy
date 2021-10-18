package setup

import (
	"github.com/mavolin/stormy/pkg/repository"
	"github.com/mavolin/stormy/pkg/repository/memory"
)

type RepositoryOptions struct {
}

func Repository(_ RepositoryOptions) repository.Repository {
	return memory.New()
}
