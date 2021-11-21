package mongo

import "go.mongodb.org/mongo-driver/bson"

func hasShardIDs(guildIDField string, shardIDs []int, numShards int) bson.M {
	return bson.M{
		"$expr": bson.M{
			"$in": bson.A{
				bson.M{
					// shardID = (guildID >> 22) % numShards
					"$mod": bson.A{
						// a bit shift x positions to the right is the same as
						// dividing by 2^x
						bson.M{"$divide": bson.A{"$" + guildIDField, bson.M{"$pow": []int{2, 22}}}},
						numShards,
					},
				},
				shardIDs,
			},
		},
	}
}
