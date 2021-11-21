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

type ideaRepo struct {
	r *Repository
	c *mongo.Collection
}

var _ idearepo.IdeaRepository = (*ideaRepo)(nil)

func newIdeaRepo(r *Repository) *ideaRepo {
	return &ideaRepo{r, r.db.Collection("ideas")}
}

// =============================================================================
// Structs
// =====================================================================================

// ================================ idea ================================

type idea struct {
	GuildID   discord.GuildID   `bson:"guild_id"`
	ChannelID discord.ChannelID `bson:"channel_id"`
	MessageID discord.MessageID `bson:"message_id"`

	GlobalSectionEmojis []string       `bson:"global_section_emojis,omitempty"`
	Groups              []sectionGroup `bson:"groups,omitempty"`

	VoteType  idearepo.VoteType `bson:"vote_type"`
	VoteUntil *time.Time        `bson:"vote_until"`
}

func newIdea(src *idearepo.Idea) *idea {
	groups := make([]sectionGroup, len(src.Groups))

	for i, g := range src.Groups {
		g := g
		groups[i] = *newSectionGroup(&g)
	}

	return &idea{
		GuildID:             src.GuildID,
		ChannelID:           src.ChannelID,
		MessageID:           src.MessageID,
		GlobalSectionEmojis: src.GlobalSectionEmojis,
		Groups:              groups,
		VoteType:            src.VoteType,
		VoteUntil:           src.VoteUntil,
	}
}

func (i *idea) asSrcType() *idearepo.Idea {
	groups := make([]idearepo.SectionGroup, len(i.Groups))

	for i, g := range i.Groups {
		groups[i] = *g.asSrcType()
	}

	return &idearepo.Idea{
		GuildID:             i.GuildID,
		ChannelID:           i.ChannelID,
		MessageID:           i.MessageID,
		GlobalSectionEmojis: i.GlobalSectionEmojis,
		Groups:              groups,
		VoteType:            i.VoteType,
		VoteUntil:           i.VoteUntil,
	}
}

// ================================ sectionGroup ================================

type sectionGroup struct {
	Title  string
	Emojis []string
}

func newSectionGroup(src *idearepo.SectionGroup) *sectionGroup {
	return &sectionGroup{Title: src.Title, Emojis: src.Emojis}
}

func (g *sectionGroup) asSrcType() *idearepo.SectionGroup {
	return &idearepo.SectionGroup{Title: g.Title, Emojis: g.Emojis}
}

// ================================ ideaCursor ================================

type ideaCursor struct {
	c *mongo.Cursor
}

var _ idearepo.IdeaCursor = (*ideaCursor)(nil)

func newIdeaCursor(c *mongo.Cursor) *ideaCursor {
	return &ideaCursor{c: c}
}

func (c *ideaCursor) BatchLength() int {
	return c.c.RemainingBatchLength()
}

func (c *ideaCursor) Next(ctx context.Context) (*idearepo.Idea, error) {
	if !c.c.Next(ctx) {
		return nil, nil
	}

	var i idea

	if err := c.c.Decode(&i); err != nil {
		return nil, errors.WithStack(err)
	}

	return i.asSrcType(), nil
}

func (c *ideaCursor) Close(ctx context.Context) error {
	return c.c.Close(ctx)
}

// =============================================================================
// DB Operations
// =====================================================================================

func (r *ideaRepo) Idea(ctx context.Context, messageID discord.MessageID) (*idearepo.Idea, error) {
	var i idea

	err := r.c.FindOne(ctx, bson.M{"message_id": messageID}).Decode(&i)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}

		return nil, err
	}

	return i.asSrcType(), nil
}

func (r *ideaRepo) SaveIdea(ctx context.Context, i *idearepo.Idea) error {
	_, err := r.c.InsertOne(ctx, newIdea(i))
	return errors.WithStack(err)
}

func (r *ideaRepo) DeleteIdea(ctx context.Context, messageID discord.MessageID) error {
	_, err := r.c.DeleteOne(ctx, bson.M{"message_id": messageID})
	return errors.WithStack(err)
}

func (r *ideaRepo) ExpiringIdeas(
	ctx context.Context, afterT time.Time, afterID discord.MessageID, limit int,
) ([]idearepo.Idea, error) {
	cursor, err := r.c.Find(ctx, bson.M{
		"$and": []bson.M{
			hasShardIDs("guildID", r.r.shardIDs, r.r.numShards),
			{
				"$or": []bson.M{
					{"vote_until": bson.M{"$gt": afterT}},
					{
						"$and": []bson.M{
							{"vote_until": afterID},
							{"message_id": bson.M{"$gt": afterID}},
						},
					},
				},
			},
		},
	}, options.Find().
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "vote_until", Value: 1}, {Key: "message_id", Value: 1}}))
	if err != nil {
		return nil, errors.WithStack(err)
	}

	ideas := make([]idea, 0, limit)

	if err := cursor.All(ctx, &ideas); err != nil {
		return nil, errors.WithStack(err)
	}

	srcIdeas := make([]idearepo.Idea, len(ideas))
	for i, idea := range ideas {
		srcIdeas[i] = *idea.asSrcType()
	}

	return srcIdeas, nil
}

func (r *ideaRepo) ExpiredIdeas(ctx context.Context, before time.Time) (idearepo.IdeaCursor, error) {
	cursor, err := r.c.Find(ctx, bson.M{"vote_until": bson.M{"$lt": before}})
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return newIdeaCursor(cursor), nil
}

func (r *ideaRepo) DeleteExpiredIdeas(ctx context.Context, before time.Time) error {
	_, err := r.c.DeleteMany(ctx, bson.M{"vote_until": bson.M{"$lt": before}})
	return errors.WithStack(err)
}
