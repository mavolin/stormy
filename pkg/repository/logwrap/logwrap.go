// Package logwrap provides a repository that logs what another repository
// does.
package logwrap

import (
	"go.uber.org/zap"

	"github.com/mavolin/stormy/pkg/repository"
)

type Wrapper struct {
	r repository.Repository
	l *zap.SugaredLogger
}

var _ repository.Repository = new(Wrapper)

func Wrap(r repository.Repository, l *zap.SugaredLogger) *Wrapper {
	return &Wrapper{r: r, l: l}
}
