package setup

import (
	"github.com/getsentry/sentry-go"
	"github.com/mavolin/disstate/v4/pkg/state"
	"go.uber.org/zap"

	ideaservice "github.com/mavolin/stormy/modules/idea/service"
	"github.com/mavolin/stormy/pkg/repository"
)

type ServiceOptions struct {
	State      *state.State
	Repository repository.Repository

	Logger *zap.SugaredLogger
	Hub    *sentry.Hub
}

func Services(o ServiceOptions) error {
	o.Logger = o.Logger.Named("service")

	_, err := ideaservice.New(ideaservice.Options{
		State:      o.State,
		Repository: o.Repository,
		Logger:     o.Logger,
		Hub:        o.Hub,
	})
	return err
}
