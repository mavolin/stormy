// Package mongo provides a mongodb repository
package mongo

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Repository struct {
	client *mongo.Client
	db     *mongo.Database

	shardIDs  []int
	numShards int

	*ideaChannelSettingsRepo
	*ideaRepo
}

type Options struct {
	// URI is the uri used to connect to the database.
	URI string

	// DBName is the name of the database that stored stormy's collections.
	DBName string

	ShardIDs  []int
	NumShards int
}

func NewRepository(o Options) (*Repository, error) {
	client, err := mongo.NewClient(options.Client().ApplyURI(o.URI))
	if err != nil {
		return nil, err
	}

	r := &Repository{
		client:    client,
		db:        client.Database(o.DBName),
		shardIDs:  o.ShardIDs,
		numShards: o.NumShards,
	}
	r.ideaChannelSettingsRepo = newIdeaChannelSettingsRepo(r)
	r.ideaRepo = newIdeaRepo(r)

	return r, nil
}

func (r *Repository) Connect(ctx context.Context) error {
	return r.client.Connect(ctx)
}
