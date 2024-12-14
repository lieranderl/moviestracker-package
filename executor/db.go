package executor

import (
	"context"
	"fmt"
	"os"

	"github.com/lieranderl/moviestracker-package/internal/databases"
	"github.com/lieranderl/moviestracker-package/internal/movies"
	"google.golang.org/api/option"
)

func (p *trackersPipeline) SaveToDb() *trackersPipeline {
	if len(p.errors) > 0 {
		return p
	}

	ctx, cancel := context.WithCancel(p.config.ctx)
	defer cancel()

	var err error

	dbtype := os.Getenv("DBTYPE")
	collection := os.Getenv("DBNAME")

	switch dbtype {
	case "firebase":
		err = saveToFirebase(ctx, p.config.firebaseConfig, p.config.goption, collection, p.movies)
	case "mongo":
		err = saveToMongo(ctx, p.config.mongoURI, "movies", collection, p.movies)
	default:
		err = fmt.Errorf("unsupported database type: %s", dbtype)
	}

	if err != nil {
		p.errors = append(p.errors, err)
	}

	return p
}

// saveToFirebase saves movies to Firebase.
func saveToFirebase(ctx context.Context, project string, goption option.ClientOption, collection string, movies_short []*movies.Short) error {
	client, err := db.FirestoreClient(ctx, project, goption)
	if err != nil {
		return fmt.Errorf("failed to initialize Firestore client: %w", err)
	}
	defer client.Close()

	storage := movies.NewFirestoreStorage(client, collection)
	for _, movie := range movies_short {
		if err := storage.WriteMovie(ctx, movie); err != nil {
			return fmt.Errorf("failed to write movie to Firestore: %w", err)
		}
	}
	return nil
}

// saveToMongo saves movies to MongoDB.
func saveToMongo(ctx context.Context, mongoURI, dbName, collectionName string, movies_short []*movies.Short) error {
	collection, err := db.GetMongoCollection(mongoURI, dbName, collectionName)
	if err != nil {
		return fmt.Errorf("failed to initialize MongoDB collection: %w", err)
	}

	storage := movies.NewMongoStorage(collection)
	for _, movie := range movies_short {
		if err := storage.WriteMovie(ctx, movie); err != nil {
			return fmt.Errorf("failed to write movie to MongoDB: %w", err)
		}
	}
	return nil
}
