package db

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	mongoURI = "mongodb://lucy:password@localhost:27017,localhost:27018,localhost:27019/?replicaSet=lucy-mongo&authSource=admin"
	timeout  = 10 * time.Second
)

// ConnectMongo establishes a connection to MongoDB and returns the client instance.
// It uses a short-lived context for the connection attempt, and returns a
// long-lived context (background) suitable for listeners.
func ConnectMongo() (*mongo.Client, context.Context, context.CancelFunc, error) {
	// --- 1. Short-lived context for CONNECTION and PING ---
	// This ensures we don't wait forever if the database is down.
	connectCtx, connectCancel := context.WithTimeout(context.Background(), timeout)
	defer connectCancel() // Important: Ensure this context is cancelled after the connect/ping finishes

	client, err := mongo.Connect(connectCtx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Ping using the short-lived context
	if err := client.Ping(connectCtx, nil); err != nil {
		// The deferred connectCancel() will run here
		return nil, nil, nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	fmt.Println("✅ Connected to MongoDB")

	// --- 2. Long-lived context for the LISTENER/Application ---
	// Use context.Background() for the main application context.
	// Use context.WithCancel(context.Background()) if you want to be able to
	// gracefully shut down the listener later (highly recommended).

	appCtx, appCancel := context.WithCancel(context.Background())

	return client, appCtx, appCancel, nil
}
