package mongo

import (
	"context"
	"time"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/mavolin/adam/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	idearepo "github.com/mavolin/stormy/modules/idea/repository"
)

type ideaChannelSettingsRepo struct {
	r *Repository
	c *mongo.Collection
}

var _ idearepo.ChannelSettingsRepository = (*ideaChannelSettingsRepo)(nil)

func newIdeaChannelSettingsRepo(r *Repository) *ideaChannelSettingsRepo {
	return &ideaChannelSettingsRepo{r, r.db.Collection("idea_channel_settings")}
}

// =============================================================================
// Structs
// =====================================================================================

type channelSettings struct {
	ChannelID    discord.ChannelID `bson:"channel_id"`
	VoteType     idearepo.VoteType `bson:"vote_type"`
	VoteDuration time.Duration     `bson:"vote_duration"`
	Anonymous    bool              `bson:"anonymous"`
	Color        discord.Color     `bson:"color"`
}

func newChannelSettings(channelID discord.ChannelID, src *idearepo.ChannelSettings) *channelSettings {
	return &channelSettings{
		ChannelID:    channelID,
		VoteType:     src.VoteType,
		VoteDuration: src.VoteDuration,
		Anonymous:    src.Anonymous,
		Color:        src.Color,
	}
}

func (c *channelSettings) asSrcType() *idearepo.ChannelSettings {
	return &idearepo.ChannelSettings{
		VoteType:     c.VoteType,
		VoteDuration: c.VoteDuration,
		Anonymous:    c.Anonymous,
		Color:        c.Color,
	}
}

// =============================================================================
// DB Operations
// =====================================================================================

func (r *ideaChannelSettingsRepo) DisableIdeaChannel(ctx context.Context, channelID discord.ChannelID) error {
	res, err := r.c.DeleteOne(ctx, bson.M{"channel_id": channelID})
	if err != nil {
		return errors.WithStack(err)
	}

	if res.DeletedCount == 0 {
		return idearepo.ErrChannelAlreadyDisabled
	}

	return nil
}

func (r *ideaChannelSettingsRepo) IdeaChannelSettings(
	ctx context.Context, channelID discord.ChannelID,
) (*idearepo.ChannelSettings, error) {
	var settings channelSettings

	err := r.c.FindOne(ctx, bson.M{"channel_id": channelID}).Decode(&settings)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}

		return nil, errors.WithStack(err)
	}

	return settings.asSrcType(), nil
}

func (r *ideaChannelSettingsRepo) SetIdeaChannelSettings(
	ctx context.Context, channelID discord.ChannelID, srcSettings idearepo.ChannelSettings,
) error {
	settings := newChannelSettings(channelID, &srcSettings)

	_, err := r.c.ReplaceOne(ctx, bson.M{"channel_id": channelID}, settings,
		options.Replace().SetUpsert(true))
	return errors.WithStack(err)
}
