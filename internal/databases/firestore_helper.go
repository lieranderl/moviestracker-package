package db

import (
	"context"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/option"
)

func FirestoreClient(ctx context.Context, projectID string, clientOption option.ClientOption) (*firestore.Client, error) {
	return firestore.NewClient(ctx, projectID, clientOption)
}

// func DeleteOldMovies(ctx context.Context, client *firestore.Client, collection string, cutoffDate time.Time) error {
// 	iter := client.Collection(collection).Where("LastTimeFound", "<", cutoffDate).Documents(ctx)
// 	batch := client.Batch()
// 	numDeleted := 0

// 	for {
// 		doc, err := iter.Next()
// 		if err == iterator.Done {
// 			break
// 		}
// 		if err != nil {
// 			return err
// 		}
// 		batch.Delete(doc.Ref)
// 		numDeleted++
// 	}

// 	if numDeleted > 0 {
// 		_, err := batch.Commit(ctx)
// 		return err
// 	}

// 	return nil
// }

// func SaveMovies(ctx context.Context, movies []*movies.Short, projectID string, clientOption option.ClientOption) error {
// 	client, err := NewClient(ctx, projectID, clientOption)
// 	if err != nil {
// 		return err
// 	}
// 	defer client.Close()

// 	collection := client.Collection("movies")
// 	for _, movie := range movies {
// 		_, err := collection.Doc(movie.ID).Set(ctx, movie)
// 		if err != nil {
// 			return err
// 		}
// 	}
// 	return nil
// }
