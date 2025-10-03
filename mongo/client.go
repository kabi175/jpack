package mongo

import (
	"context"
	"fmt"
	"time"

	"github.com/kabi175/jpack/logger"
	"github.com/kabi175/jpack/schema"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoClient wraps the MongoDB client with additional functionality
type MongoClient struct {
	client       *mongo.Client
	database     *mongo.Database
	schemas      map[string]schema.JSchema
	repositories map[string]*MongoRepository
}

// NewMongoClient creates a new MongoDB client
func NewMongoClient(uri, databaseName string) (*MongoClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		logger.Mongo.Error().
			Str("uri", uri).
			Err(err).
			Msg("failed to connect to MongoDB")
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Test the connection
	err = client.Ping(ctx, nil)
	if err != nil {
		logger.Mongo.Error().
			Str("uri", uri).
			Err(err).
			Msg("failed to ping MongoDB")
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	database := client.Database(databaseName)

	return &MongoClient{
		client:       client,
		database:     database,
		schemas:      make(map[string]schema.JSchema),
		repositories: make(map[string]*MongoRepository),
	}, nil
}

// NewMongoClientWithOptions creates a new MongoDB client with custom options
func NewMongoClientWithOptions(uri, databaseName string, clientOptions *options.ClientOptions) (*MongoClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions.ApplyURI(uri))
	if err != nil {
		logger.Mongo.Error().
			Str("uri", uri).
			Err(err).
			Msg("failed to connect to MongoDB with options")
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Test the connection
	err = client.Ping(ctx, nil)
	if err != nil {
		logger.Mongo.Error().
			Str("uri", uri).
			Err(err).
			Msg("failed to ping MongoDB with options")
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	database := client.Database(databaseName)

	return &MongoClient{
		client:       client,
		database:     database,
		schemas:      make(map[string]schema.JSchema),
		repositories: make(map[string]*MongoRepository),
	}, nil
}

// GetClient returns the underlying MongoDB client
func (mc *MongoClient) GetClient() *mongo.Client {
	return mc.client
}

// GetDatabase returns the database
func (mc *MongoClient) GetDatabase() *mongo.Database {
	return mc.database
}

// RegisterSchema registers a schema and creates a repository
func (mc *MongoClient) RegisterSchema(schema schema.JSchema) error {
	if schema == nil {
		logger.Mongo.Error().Msg("schema cannot be nil")
		return fmt.Errorf("schema cannot be nil")
	}

	if schema.Name() == "" {
		logger.Mongo.Error().Msg("schema name cannot be empty")
		return fmt.Errorf("schema name cannot be empty")
	}

	// Check if schema is already registered
	if _, exists := mc.schemas[schema.Name()]; exists {
		logger.Mongo.Error().
			Str("schema", schema.Name()).
			Msg("schema is already registered")
		return fmt.Errorf("schema '%s' is already registered", schema.Name())
	}

	// Register schema
	mc.schemas[schema.Name()] = schema

	// Create collection
	collection := mc.database.Collection(schema.Name())

	// Create repository
	repository := NewMongoRepository(schema, collection)
	mc.repositories[schema.Name()] = repository

	logger.Mongo.Info().
		Str("schema", schema.Name()).
		Msg("schema registered successfully")

	return nil
}

// GetRepository returns a repository for a schema
func (mc *MongoClient) GetRepository(schemaName string) (*MongoRepository, error) {
	repository, exists := mc.repositories[schemaName]
	if !exists {
		logger.Mongo.Error().
			Str("schema", schemaName).
			Msg("repository for schema not found")
		return nil, fmt.Errorf("repository for schema '%s' not found", schemaName)
	}
	return repository, nil
}

// GetSchema returns a schema by name
func (mc *MongoClient) GetSchema(schemaName string) (schema.JSchema, error) {
	schema, exists := mc.schemas[schemaName]
	if !exists {
		logger.Mongo.Error().
			Str("schema", schemaName).
			Msg("schema not found")
		return nil, fmt.Errorf("schema '%s' not found", schemaName)
	}
	return schema, nil
}

// ListSchemas returns all registered schema names
func (mc *MongoClient) ListSchemas() []string {
	names := make([]string, 0, len(mc.schemas))
	for name := range mc.schemas {
		names = append(names, name)
	}
	return names
}

// CreateIndexes creates indexes for all registered schemas
func (mc *MongoClient) CreateIndexes(ctx context.Context) error {
	for schemaName, schema := range mc.schemas {
		collection := mc.database.Collection(schemaName)
		indexManager := NewIndexManager(collection, schema)

		if err := indexManager.CreateIndexes(ctx); err != nil {
			logger.Mongo.Error().
				Str("schema", schemaName).
				Err(err).
				Msg("failed to create indexes for schema")
			return fmt.Errorf("failed to create indexes for schema '%s': %w", schemaName, err)
		}
	}
	return nil
}

// CreateIndexesForSchema creates indexes for a specific schema
func (mc *MongoClient) CreateIndexesForSchema(ctx context.Context, schemaName string) error {
	schema, exists := mc.schemas[schemaName]
	if !exists {
		logger.Mongo.Error().
			Str("schema", schemaName).
			Msg("schema not found for index creation")
		return fmt.Errorf("schema '%s' not found", schemaName)
	}

	collection := mc.database.Collection(schemaName)
	indexManager := NewIndexManager(collection, schema)

	return indexManager.CreateIndexes(ctx)
}

// DropCollection drops a collection
func (mc *MongoClient) DropCollection(ctx context.Context, collectionName string) error {
	collection := mc.database.Collection(collectionName)
	return collection.Drop(ctx)
}

// ListCollections lists all collections in the database
func (mc *MongoClient) ListCollections(ctx context.Context) ([]string, error) {
	cursor, err := mc.database.ListCollectionNames(ctx, nil)
	if err != nil {
		logger.Mongo.Error().
			Err(err).
			Msg("failed to list collections")
		return nil, fmt.Errorf("failed to list collections: %w", err)
	}
	return cursor, nil
}

// Close closes the MongoDB connection
func (mc *MongoClient) Close(ctx context.Context) error {
	return mc.client.Disconnect(ctx)
}

// Ping tests the connection to MongoDB
func (mc *MongoClient) Ping(ctx context.Context) error {
	return mc.client.Ping(ctx, nil)
}

// GetStats returns database statistics
func (mc *MongoClient) GetStats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Get database stats
	var dbStats map[string]interface{}
	err := mc.database.RunCommand(ctx, map[string]interface{}{"dbStats": 1}).Decode(&dbStats)
	if err != nil {
		logger.Mongo.Error().
			Err(err).
			Msg("failed to get database stats")
		return nil, fmt.Errorf("failed to get database stats: %w", err)
	}
	stats["database"] = dbStats

	// Get collection stats for registered schemas
	collectionStats := make(map[string]interface{})
	for schemaName := range mc.schemas {
		collection := mc.database.Collection(schemaName)
		var collStats map[string]interface{}
		err := collection.Database().RunCommand(ctx, map[string]interface{}{
			"collStats": schemaName,
		}).Decode(&collStats)
		if err != nil {
			// Continue with other collections if one fails
			continue
		}
		collectionStats[schemaName] = collStats
	}
	stats["collections"] = collectionStats

	return stats, nil
}

// TransactionManager manages MongoDB transactions
type TransactionManager struct {
	client *MongoClient
}

// NewTransactionManager creates a new transaction manager
func NewTransactionManager(client *MongoClient) *TransactionManager {
	return &TransactionManager{
		client: client,
	}
}

// ExecuteInTransaction executes a function within a transaction
func (tm *TransactionManager) ExecuteInTransaction(ctx context.Context, fn func(context.Context) error) error {
	session, err := tm.client.client.StartSession()
	if err != nil {
		logger.Mongo.Error().
			Err(err).
			Msg("failed to start session")
		return fmt.Errorf("failed to start session: %w", err)
	}
	defer session.EndSession(ctx)

	_, err = session.WithTransaction(ctx, func(sc mongo.SessionContext) (interface{}, error) {
		return nil, fn(sc)
	})

	if err != nil {
		logger.Mongo.Error().
			Err(err).
			Msg("transaction failed")
		return fmt.Errorf("transaction failed: %w", err)
	}

	return nil
}

// GetTransactionManager returns a transaction manager for this client
func (mc *MongoClient) GetTransactionManager() *TransactionManager {
	return NewTransactionManager(mc)
}

// HealthChecker checks the health of the MongoDB connection
type HealthChecker struct {
	client *MongoClient
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(client *MongoClient) *HealthChecker {
	return &HealthChecker{
		client: client,
	}
}

// CheckHealth checks the health of the MongoDB connection
func (hc *HealthChecker) CheckHealth(ctx context.Context) error {
	// Ping the database
	if err := hc.client.Ping(ctx); err != nil {
		logger.Mongo.Error().
			Err(err).
			Msg("database ping failed")
		return fmt.Errorf("database ping failed: %w", err)
	}

	// Check if we can list collections
	_, err := hc.client.ListCollections(ctx)
	if err != nil {
		logger.Mongo.Error().
			Err(err).
			Msg("failed to list collections in health check")
		return fmt.Errorf("failed to list collections: %w", err)
	}

	return nil
}

// GetHealthChecker returns a health checker for this client
func (mc *MongoClient) GetHealthChecker() *HealthChecker {
	return NewHealthChecker(mc)
}
