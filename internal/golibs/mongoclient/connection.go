package mongoclient

import (
	"context"
	"fmt"
	"sync"

	"github.com/manabie-com/backend/internal/golibs/configs"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

/*
	Used to create a singleton object of MongoDB client.

Initialized and exposed through  GetMongoClient().
*/
var (
	clientInstance *mongo.Client
	dbInstance     *mongo.Database
	errorMessage   error
)

// Used to execute client creation procedure only once.
var mongoOnce sync.Once

// GetMongoClient - Return mongodb connection to work with
func GetMongoClient(ctx context.Context, logger *zap.Logger, cfg configs.MongoConfig) (*mongo.Client, *mongo.Database) {
	// Perform connection creation operation only once.
	mongoOnce.Do(func() {
		// Set client options
		clientOptions := options.Client().ApplyURI(cfg.Connection)
		// Connect to MongoDB
		client, err := mongo.Connect(ctx, clientOptions)
		if err != nil {
			errorMessage = fmt.Errorf("failed to connect mongodb: %w", err)
		} else {
			// Check the connection
			err = client.Ping(ctx, nil)
			if err != nil {
				errorMessage = fmt.Errorf("failed to ping mongodb: %w", err)
			} else {
				clientInstance, dbInstance = client, client.Database(cfg.Database)
			}
		}
	})
	if errorMessage != nil {
		logger.Error("%s", zap.Error(errorMessage))
		return nil, nil
	}
	return clientInstance, dbInstance
}
