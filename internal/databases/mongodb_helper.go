package db

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// GetMongoCollection initializes a MongoDB connection and returns a specific collection.
func GetMongoCollection(uri, dbName, collectionName string) (*mongo.Collection, error) {
	client, err := connectMongo(uri)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	collection := client.Database(dbName).Collection(collectionName)
	return collection, nil
}

// connectMongo initializes and verifies a MongoDB client connection.
func connectMongo(uri string) (*mongo.Client, error) {
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("error connecting to MongoDB: %w", err)
	}

	// Ping the database to ensure the connection is established
	if err := client.Ping(context.Background(), nil); err != nil {
		return nil, fmt.Errorf("error pinging MongoDB: %w", err)
	}

	return client, nil
}
