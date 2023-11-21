package movies

import (
	"context"
	"log"

	"cloud.google.com/go/firestore"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (m *Short) WriteMovieToFirestore(ctx context.Context, firestoreClient *firestore.Client, collection string) {
	moviesListRef := firestoreClient.Collection(collection)
	_, err := moviesListRef.Doc(m.ID).Set(ctx, m)
	if err != nil {
		log.Println("Failed to write", m.ID, m.Title)
	}
}

func (m *Short) WriteMovieToMongo(ctx context.Context, collection *mongo.Collection) {
	_, err := collection.UpdateOne(ctx, bson.M{"id": m.ID}, bson.M{"$set": m}, options.Update().SetUpsert(true))
	if err != nil {
		log.Println(err)
		log.Println("Failed to write", m.ID, m.Title)
	}
}
