// Package mongo provides MongoDB integration for JPack.
// It includes client management, repository creation, index management, and transaction support.
//
// The mongo package provides a high-level abstraction over the MongoDB Go driver,
// making it easy to work with schemas and perform CRUD operations.
//
// Example:
//
//	// Create MongoDB client
//	client, err := mongo.NewMongoClient("mongodb://localhost:27017", "mydb")
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer client.Close(context.Background())
//
//	// Create and register schema
//	userSchema := schema.NewSchemaBuilder("User").
//		AddRequiredField("name", schema.JString, nil).
//		AddRequiredUniqueField("email", schema.JString, nil).
//		AddField("age", schema.JInt, 18).
//		Build()
//
//	err = client.RegisterSchema(userSchema)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Create indexes
//	err = client.CreateIndexesForSchema(context.Background(), "User")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Get repository
//	userRepo, err := client.GetRepository("User")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Create and save user
//	user := schema.NewJRecord().
//		Set("id", "user_001").
//		Set("name", "John Doe").
//		Set("email", "john@example.com").
//		Set("age", 25)
//
//	savedUser, err := userRepo.Save(context.Background(), user)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Query users
//	criteria := repository.NewJCriteriaBuilder().
//		WhereGreaterThanOrEqual("age", 18).
//		OrderBy("name").
//		SetLimit(10).
//		Build()
//
//	users, err := userRepo.FindBy(context.Background(), criteria)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Use transactions
//	txManager := client.GetTransactionManager()
//	err = txManager.WithTransaction(context.Background(), func(ctx context.Context) error {
//		// Perform multiple operations within transaction
//		_, err := userRepo.Save(ctx, user)
//		return err
//	})
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Health check
//	healthChecker := client.GetHealthChecker()
//	healthy, err := healthChecker.CheckHealth(context.Background())
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Printf("Database healthy: %v\n", healthy)
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

// MongoClient wraps the MongoDB client with additional functionality.
// It provides schema management, repository creation, and database operations.
//
// Example:
//
//	client, err := mongo.NewMongoClient("mongodb://localhost:27017", "mydb")
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer client.Close(context.Background())
//
//	// Register schemas
//	userSchema := schema.NewSchemaBuilder("User").Build()
//	err = client.RegisterSchema(userSchema)
//
//	// Get repository
//	userRepo, err := client.GetRepository("User")
type MongoClient struct {
	client       *mongo.Client
	database     *mongo.Database
	schemas      map[string]schema.JSchema
	repositories map[string]*MongoRepository
}

// NewMongoClient creates a new MongoDB client.
// It connects to MongoDB using the provided URI and database name.
//
// Example:
//
//	client, err := mongo.NewMongoClient("mongodb://localhost:27017", "mydb")
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer client.Close(context.Background())
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

// NewMongoClientWithOptions creates a new MongoDB client with custom options.
// This allows you to configure connection pooling, timeouts, and other MongoDB client options.
//
// Example:
//
//	clientOptions := options.Client().
//		SetMaxPoolSize(100).
//		SetMinPoolSize(10).
//		SetMaxConnIdleTime(30 * time.Second)
//
//	client, err := mongo.NewMongoClientWithOptions(
//		"mongodb://localhost:27017",
//		"mydb",
//		clientOptions,
//	)
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

// GetClient returns the underlying MongoDB client.
// This provides direct access to the MongoDB Go driver client for advanced operations.
//
// Example:
//
//	client, err := mongo.NewMongoClient("mongodb://localhost:27017", "mydb")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	mongoClient := client.GetClient()
//	// Use mongoClient for advanced operations
func (mc *MongoClient) GetClient() *mongo.Client {
	return mc.client
}

// GetDatabase returns the database.
// This provides direct access to the MongoDB database for advanced operations.
//
// Example:
//
//	client, err := mongo.NewMongoClient("mongodb://localhost:27017", "mydb")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	database := client.GetDatabase()
//	// Use database for advanced operations
func (mc *MongoClient) GetDatabase() *mongo.Database {
	return mc.database
}

// RegisterSchema registers a schema and creates a repository.
// This makes the schema available for use and creates the corresponding MongoDB collection.
//
// Example:
//
//	client, err := mongo.NewMongoClient("mongodb://localhost:27017", "mydb")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	userSchema := schema.NewSchemaBuilder("User").
//		AddField("name", schema.JString, nil).
//		Build()
//
//	err = client.RegisterSchema(userSchema)
//	if err != nil {
//		log.Fatal(err)
//	}
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

// GetRepository returns a repository for a schema.
// The repository provides CRUD operations for the schema.
//
// Example:
//
//	client, err := mongo.NewMongoClient("mongodb://localhost:27017", "mydb")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Register schema first
//	userSchema := schema.NewSchemaBuilder("User").Build()
//	err = client.RegisterSchema(userSchema)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Get repository
//	userRepo, err := client.GetRepository("User")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Use repository
//	user := schema.NewJRecord().Set("id", "user_001").Set("name", "John")
//	savedUser, err := userRepo.Save(context.Background(), user)
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

// GetSchema returns a schema by name.
// This allows you to retrieve a previously registered schema.
//
// Example:
//
//	client, err := mongo.NewMongoClient("mongodb://localhost:27017", "mydb")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Register schema
//	userSchema := schema.NewSchemaBuilder("User").Build()
//	err = client.RegisterSchema(userSchema)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Get schema
//	retrievedSchema, err := client.GetSchema("User")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	fmt.Println(retrievedSchema.Name()) // "User"
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

// ListSchemas returns all registered schema names.
// This provides a list of all schemas that have been registered with the client.
//
// Example:
//
//	client, err := mongo.NewMongoClient("mongodb://localhost:27017", "mydb")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Register schemas
//	userSchema := schema.NewSchemaBuilder("User").Build()
//	orderSchema := schema.NewSchemaBuilder("Order").Build()
//	client.RegisterSchema(userSchema)
//	client.RegisterSchema(orderSchema)
//
//	// List schemas
//	schemaNames := client.ListSchemas()
//	fmt.Println(schemaNames) // ["User", "Order"]
func (mc *MongoClient) ListSchemas() []string {
	names := make([]string, 0, len(mc.schemas))
	for name := range mc.schemas {
		names = append(names, name)
	}
	return names
}

// CreateIndexes creates indexes for all registered schemas.
// This automatically creates indexes based on field constraints (unique, required, etc.).
//
// Example:
//
//	client, err := mongo.NewMongoClient("mongodb://localhost:27017", "mydb")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Register schema with unique field
//	userSchema := schema.NewSchemaBuilder("User").
//		AddRequiredUniqueField("email", schema.JString, nil).
//		Build()
//	client.RegisterSchema(userSchema)
//
//	// Create indexes for all schemas
//	err = client.CreateIndexes(context.Background())
//	if err != nil {
//		log.Fatal(err)
//	}
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

// CreateIndexesForSchema creates indexes for a specific schema.
// This creates indexes for a single schema instead of all registered schemas.
//
// Example:
//
//	client, err := mongo.NewMongoClient("mongodb://localhost:27017", "mydb")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Register schema
//	userSchema := schema.NewSchemaBuilder("User").
//		AddRequiredUniqueField("email", schema.JString, nil).
//		Build()
//	client.RegisterSchema(userSchema)
//
//	// Create indexes for specific schema
//	err = client.CreateIndexesForSchema(context.Background(), "User")
//	if err != nil {
//		log.Fatal(err)
//	}
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

// Close closes the MongoDB connection.
// This should be called when the client is no longer needed to free resources.
//
// Example:
//
//	client, err := mongo.NewMongoClient("mongodb://localhost:27017", "mydb")
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer client.Close(context.Background())
//
//	// Use client...
func (mc *MongoClient) Close(ctx context.Context) error {
	return mc.client.Disconnect(ctx)
}

// Ping tests the connection to MongoDB.
// This can be used to verify that the client is connected to the database.
//
// Example:
//
//	client, err := mongo.NewMongoClient("mongodb://localhost:27017", "mydb")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Test connection
//	err = client.Ping(context.Background())
//	if err != nil {
//		log.Fatal("Connection failed:", err)
//	}
//	fmt.Println("Connected to MongoDB")
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

// TransactionManager manages MongoDB transactions.
// It provides a high-level interface for executing operations within transactions.
//
// Example:
//
//	client, err := mongo.NewMongoClient("mongodb://localhost:27017", "mydb")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	txManager := client.GetTransactionManager()
//	err = txManager.ExecuteInTransaction(context.Background(), func(ctx context.Context) error {
//		// Perform multiple operations within transaction
//		userRepo, _ := client.GetRepository("User")
//		orderRepo, _ := client.GetRepository("Order")
//
//		user := schema.NewJRecord().Set("id", "user_001").Set("name", "John")
//		_, err := userRepo.Save(ctx, user)
//		if err != nil {
//			return err
//		}
//
//		order := schema.NewJRecord().Set("id", "order_001").Set("user_id", "user_001")
//		_, err = orderRepo.Save(ctx, order)
//		return err
//	})
type TransactionManager struct {
	client *MongoClient
}

// NewTransactionManager creates a new transaction manager.
// This is typically called internally by the MongoClient.
//
// Example:
//
//	client, err := mongo.NewMongoClient("mongodb://localhost:27017", "mydb")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	txManager := mongo.NewTransactionManager(client)
func NewTransactionManager(client *MongoClient) *TransactionManager {
	return &TransactionManager{
		client: client,
	}
}

// ExecuteInTransaction executes a function within a transaction.
// All operations within the function are executed atomically.
//
// Example:
//
//	txManager := client.GetTransactionManager()
//	err = txManager.ExecuteInTransaction(context.Background(), func(ctx context.Context) error {
//		// All operations here are atomic
//		userRepo, _ := client.GetRepository("User")
//		orderRepo, _ := client.GetRepository("Order")
//
//		user := schema.NewJRecord().Set("id", "user_001").Set("name", "John")
//		_, err := userRepo.Save(ctx, user)
//		if err != nil {
//			return err // Transaction will be rolled back
//		}
//
//		order := schema.NewJRecord().Set("id", "order_001").Set("user_id", "user_001")
//		_, err = orderRepo.Save(ctx, order)
//		return err
//	})
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

// GetTransactionManager returns a transaction manager for this client.
// This provides access to transaction functionality.
//
// Example:
//
//	client, err := mongo.NewMongoClient("mongodb://localhost:27017", "mydb")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	txManager := client.GetTransactionManager()
//	err = txManager.ExecuteInTransaction(context.Background(), func(ctx context.Context) error {
//		// Transaction operations
//		return nil
//	})
func (mc *MongoClient) GetTransactionManager() *TransactionManager {
	return NewTransactionManager(mc)
}

// HealthChecker checks the health of the MongoDB connection.
// It provides methods to verify that the database is accessible and responsive.
//
// Example:
//
//	client, err := mongo.NewMongoClient("mongodb://localhost:27017", "mydb")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	healthChecker := client.GetHealthChecker()
//	err = healthChecker.CheckHealth(context.Background())
//	if err != nil {
//		log.Fatal("Database health check failed:", err)
//	}
//	fmt.Println("Database is healthy")
type HealthChecker struct {
	client *MongoClient
}

// NewHealthChecker creates a new health checker.
// This is typically called internally by the MongoClient.
//
// Example:
//
//	client, err := mongo.NewMongoClient("mongodb://localhost:27017", "mydb")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	healthChecker := mongo.NewHealthChecker(client)
func NewHealthChecker(client *MongoClient) *HealthChecker {
	return &HealthChecker{
		client: client,
	}
}

// CheckHealth checks the health of the MongoDB connection.
// It performs a ping and verifies that collections can be listed.
//
// Example:
//
//	healthChecker := client.GetHealthChecker()
//	err = healthChecker.CheckHealth(context.Background())
//	if err != nil {
//		log.Fatal("Database health check failed:", err)
//	}
//	fmt.Println("Database is healthy")
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

// GetHealthChecker returns a health checker for this client.
// This provides access to health checking functionality.
//
// Example:
//
//	client, err := mongo.NewMongoClient("mongodb://localhost:27017", "mydb")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	healthChecker := client.GetHealthChecker()
//	err = healthChecker.CheckHealth(context.Background())
//	if err != nil {
//		log.Fatal("Database health check failed:", err)
//	}
func (mc *MongoClient) GetHealthChecker() *HealthChecker {
	return NewHealthChecker(mc)
}
