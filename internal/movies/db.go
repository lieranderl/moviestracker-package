package movies

import (
	"context"
	"fmt"
	"log"

	"cloud.google.com/go/firestore"
)

func (m *Short) WriteMovieToDb(ctx context.Context, firestoreClient *firestore.Client, collection string) {
	moviesListRef := firestoreClient.Collection(collection)
	_, err := moviesListRef.Doc(fmt.Sprint(m.ID)).Set(ctx, m)
	if err != nil {
		log.Println("Failed to write", m.ID, m.Title)
	}
}
