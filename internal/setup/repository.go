package setup

import (
	"context"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/mavolin/stormy/pkg/repository"
	"github.com/mavolin/stormy/pkg/repository/logwrap"
	"github.com/mavolin/stormy/pkg/repository/memory"
	"github.com/mavolin/stormy/pkg/repository/mongo"
)

type RepositoryOptions struct {
	MongoURI    string
	MongoDBName string

	ShardIDs  []int
	NumShards int

	Logger *zap.SugaredLogger
}

func Repository(o RepositoryOptions) (repository.Repository, error) {
	o.Logger = o.Logger.Named("repository")

	if o.MongoURI != "" {
		r, err := newMongoRepository(o)
		if err != nil {
			return nil, err
		}

		return logWrap(r, o.Logger), nil
	}

	o.Logger.Warn("no mongodb uri was provided, running in memory db mode; changes are not permanent!")

	return logWrap(memory.New(), o.Logger), nil
}

func newMongoRepository(o RepositoryOptions) (repository.Repository, error) {
	r, err := mongo.NewRepository(mongo.Options{
		URI:       o.MongoURI,
		DBName:    o.MongoDBName,
		ShardIDs:  o.ShardIDs,
		NumShards: o.NumShards,
	})
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	o.Logger.Info("connecting to mongo instance")
	if err := r.Connect(ctx); err != nil {
		return nil, err
	}

	return r, nil
}

func logWrap(r repository.Repository, l *zap.SugaredLogger) repository.Repository {
	if !l.Desugar().Core().Enabled(zapcore.DebugLevel) {
		return r
	}

	return logwrap.Wrap(r, l)
}
