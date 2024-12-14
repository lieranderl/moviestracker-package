package movies

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MovieStorage defines the interface for movie storage operations
type MovieStorage interface {
	WriteMovie(ctx context.Context, movie *Short) error
}

// FirestoreStorage implements MovieStorage for Firestore
type FirestoreStorage struct {
	client     *firestore.Client
	collection string
}

// NewFirestoreStorage creates a new FirestoreStorage instance
func NewFirestoreStorage(client *firestore.Client, collection string) *FirestoreStorage {
	return &FirestoreStorage{
		client:     client,
		collection: collection,
	}
}

// WriteMovie implements MovieStorage interface for Firestore
func (fs *FirestoreStorage) WriteMovie(ctx context.Context, movie *Short) error {
	_, err := fs.client.Collection(fs.collection).Doc(movie.ID).Set(ctx, movie)
	if err != nil {
		return fmt.Errorf("failed to write movie %s (%s) to Firestore: %w", movie.ID, movie.Title, err)
	}
	return nil
}

// MongoStorage implements MovieStorage for MongoDB
type MongoStorage struct {
	collection *mongo.Collection
}

// NewMongoStorage creates a new MongoStorage instance
func NewMongoStorage(collection *mongo.Collection) *MongoStorage {
	return &MongoStorage{
		collection: collection,
	}
}

// WriteMovie implements MovieStorage interface for MongoDB
func (ms *MongoStorage) WriteMovie(ctx context.Context, movie *Short) error {
	_, err := ms.collection.UpdateOne(
		ctx,
		bson.M{"id": movie.ID},
		bson.M{"$set": movie},
		options.Update().SetUpsert(true),
	)
	if err != nil {
		return fmt.Errorf("failed to write movie %s (%s) to MongoDB: %w", movie.ID, movie.Title, err)
	}
	return nil
}
