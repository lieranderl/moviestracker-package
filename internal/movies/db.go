package movies

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func (m *Short) WriteMovieToMongo(ctx context.Context, collection *mongo.Collection) error {
	_, err := collection.UpdateOne(ctx, bson.M{"id": m.ID}, bson.M{"$set": m}, options.UpdateOne().SetUpsert(true))
	if err != nil {
		return fmt.Errorf("write mongo movie %s (%s): %w", m.ID, m.Title, err)
	}
	return nil
}
