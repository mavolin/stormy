// Package service provides the idea service.
package service

import (
	"context"
	"sync"
	"time"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/getsentry/sentry-go"
	lru "github.com/hashicorp/golang-lru"
	"github.com/mavolin/adam/pkg/plugin"
	"github.com/mavolin/disstate/v4/pkg/state"
	"github.com/mavolin/sentryadam/pkg/sentrystate"
	"go.uber.org/zap"

	"github.com/mavolin/stormy/internal/errhandler"
	"github.com/mavolin/stormy/modules/idea/repository"
	"github.com/mavolin/stormy/modules/idea/service/singleflight"
)

// Service is the service that reposts ideas and publishes voting results.
type Service struct {
	repo   repository.Repository
	selfID discord.UserID

	sf *singleflight.Manager

	timeoutWatcher *deadlineWatcher
	expiredIdeas   *lru.TwoQueueCache

	log        *zap.SugaredLogger
	errhandler *errhandler.Handler

	missingManageChannelPerm map[discord.ChannelID]struct{}
	mu                       sync.Mutex
}

type Options struct {
	State      *state.State
	Repository repository.Repository

	Logger *zap.SugaredLogger
	Hub    *sentry.Hub
}

func New(o Options) (*Service, error) {
	self, err := o.State.Me()
	if err != nil {
		return nil, err
	}

	expiredIdeas, err := lru.New2Q(10_000)
	if err != nil {
		return nil, err
	}

	service := &Service{
		repo:                     o.Repository,
		selfID:                   self.ID,
		sf:                       singleflight.NewManager(),
		expiredIdeas:             expiredIdeas,
		log:                      o.Logger.Named("idea"),
		errhandler:               errhandler.NewHandler("idea", o.Logger, o.Hub),
		missingManageChannelPerm: make(map[discord.ChannelID]struct{}),
	}

	now := time.Now()

	service.timeoutWatcher = newDeadlineWatcher(service, o.State, service.repo)
	if err = service.timeoutWatcher.start(now); err != nil {
		return nil, err
	}

	if err = service.countExpiredVotes(o.State, now); err != nil {
		return nil, err
	}

	o.State.AddHandler(service.onNewIdea, sentrystate.NewMiddleware(sentrystate.HandlerMeta{
		Hub:                o.Hub,
		PluginSource:       plugin.BuiltInSource,
		PluginID:           ".idea",
		Operation:          "onNewIdea",
		MonitorPerformance: true,
	}), errhandler.DeleteError)
	o.State.AddHandler(service.onReactionAdd, sentrystate.NewMiddleware(sentrystate.HandlerMeta{
		Hub:                o.Hub,
		PluginSource:       plugin.BuiltInSource,
		PluginID:           ".idea",
		Operation:          "onReactionAdd",
		MonitorPerformance: true,
	}))

	o.State.DeriveIntents()

	return service, nil
}

// idea returns the idea associated with the given message id, if there is any.
// Before querying the repository, it checks if the idea was recently checked
// and was expired then.
//
// Therefore, this method should be preferred over the repository's Idea
// method.
func (service *Service) idea(ctx context.Context, messageID discord.MessageID) (*repository.Idea, error) {
	// use Get over Contains to update frequency and recency
	if _, ok := service.expiredIdeas.Get(messageID); ok {
		return nil, nil
	}

	i, err := service.repo.Idea(ctx, messageID)
	if err == nil && i == nil {
		service.expiredIdeas.Add(messageID, nil)
	}

	return i, err
}

func (service *Service) saveIdea(ctx context.Context, i *repository.Idea) error {
	if i.VoteUntil != nil {
		service.timeoutWatcher.addIdea(i)
	}

	return service.repo.SaveIdea(ctx, i)
}

// deleteIdea stores the deletion of the message in the cache and then calls
// repo.DeleteIdea.
func (service *Service) deleteIdea(ctx context.Context, messageID discord.MessageID) error {
	service.expiredIdeas.Add(messageID, nil)
	return service.repo.DeleteIdea(ctx, messageID)
}
