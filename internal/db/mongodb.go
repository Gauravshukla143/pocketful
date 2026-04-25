package db

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Client is the global MongoDB client instance.
var Client *mongo.Client

// Database is the global MongoDB database instance.
var Database *mongo.Database

// Connect establishes a connection to MongoDB.
func Connect(uri, dbName string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	// Ping to verify connection
	if err := client.Ping(ctx, nil); err != nil {
		log.Fatalf("Failed to ping MongoDB: %v", err)
	}

	Client = client
	Database = client.Database(dbName)
	log.Printf("✅ Connected to MongoDB | Database: %s", dbName)
}

// GetCollection returns a MongoDB collection by name.
func GetCollection(name string) *mongo.Collection {
	return Database.Collection(name)
}

// Disconnect closes the MongoDB connection gracefully.
func Disconnect() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := Client.Disconnect(ctx); err != nil {
		log.Printf("Error disconnecting MongoDB: %v", err)
	} else {
		log.Println("MongoDB connection closed.")
	}
}
