package main

import (
	"context"
	"fmt"
	"log"

	"github.com/kabi175/jpack/logger"
	"github.com/kabi175/jpack/mongo"
	"github.com/kabi175/jpack/schema"
)

// ImmutableSchemaExample demonstrates immutable schema usage
func ImmutableSchemaExample() {
	// Initialize logger
	logger.Initialize()
	// Create MongoDB client
	client, err := mongo.NewMongoClient("mongodb://localhost:27017", "jpack_immutable_example")
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}
	defer client.Close(context.Background())

	// Example 1: Using SchemaBuilder to create an immutable schema
	logger.App.Info().Msg("=== Example 1: SchemaBuilder ===")
	userSchema := schema.NewSchemaBuilder("User").
		AddRequiredUniqueField("id", schema.JString, nil).
		AddRequiredField("name", schema.JString, nil).
		AddRequiredUniqueField("email", schema.JString, nil).
		AddField("age", schema.JInt, 18).
		AddField("created_at", schema.JTime, nil).
		AddField("updated_at", schema.JTime, nil).
		SetIDField("id").
		AddValidation(func(ctx context.Context, rec schema.JRecord) error {
			age := rec.Get("age")
			if age != nil {
				if ageInt, ok := age.(int); ok && ageInt < 0 {
					return fmt.Errorf("age cannot be negative")
				}
			}
			return nil
		}).
		BuildImmutable() // This makes the schema immutable

	logger.App.Info().
		Bool("immutable", userSchema.IsImmutable()).
		Msg("Schema is immutable")

	// Register the immutable schema
	err = client.RegisterSchema(userSchema)
	if err != nil {
		log.Fatal("Failed to register schema:", err)
	}

	// Example 2: Using convenience functions
	logger.App.Info().Msg("=== Example 2: Convenience Functions ===")
	productSchema := schema.CreateProductSchema()
	logger.App.Info().
		Bool("immutable", productSchema.IsImmutable()).
		Msg("Product schema is immutable")

	err = client.RegisterSchema(productSchema)
	if err != nil {
		log.Fatal("Failed to register product schema:", err)
	}

	// Example 3: Schema replacement
	logger.App.Info().Msg("=== Example 3: Schema Replacement ===")

	// Create a new version of the user schema with additional fields
	updatedUserSchema := schema.NewSchemaBuilder("User").
		AddRequiredUniqueField("id", schema.JString, nil).
		AddRequiredField("name", schema.JString, nil).
		AddRequiredUniqueField("email", schema.JString, nil).
		AddField("age", schema.JInt, 18).
		AddField("phone", schema.JString, nil).   // New field
		AddField("address", schema.JString, nil). // New field
		AddField("created_at", schema.JTime, nil).
		AddField("updated_at", schema.JTime, nil).
		SetIDField("id").
		AddValidation(func(ctx context.Context, rec schema.JRecord) error {
			age := rec.Get("age")
			if age != nil {
				if ageInt, ok := age.(int); ok && ageInt < 0 {
					return fmt.Errorf("age cannot be negative")
				}
			}
			return nil
		}).
		BuildImmutable()

	// Replace the existing schema
	err = client.GetDatabase().Collection("User").Drop(context.Background())
	if err != nil {
		log.Printf("Warning: Failed to drop collection: %v", err)
	}

	// Register the new schema (this will replace the old one)
	err = client.RegisterSchema(updatedUserSchema)
	if err != nil {
		log.Fatal("Failed to register updated schema:", err)
	}

	logger.App.Info().Msg("Schema replaced successfully")

	// Example 4: Demonstrating immutability
	logger.App.Info().Msg("=== Example 4: Immutability Demonstration ===")

	// Try to modify an immutable schema (this will panic)
	defer func() {
		if r := recover(); r != nil {
			logger.App.Info().
				Interface("panic", r).
				Msg("Caught panic as expected")
		}
	}()

	// This will panic because the schema is immutable
	// userSchema.AddField("new_field", schema.JString, nil)

	// Example 5: Cloning for modifications
	logger.App.Info().Msg("=== Example 5: Cloning for Modifications ===")

	// Get the current schema
	currentSchema, err := client.GetSchema("User")
	if err != nil {
		log.Fatal("User schema not found:", err)
	}

	// Clone the schema to make modifications
	mutableClone := currentSchema.Clone()
	logger.App.Info().
		Bool("immutable", mutableClone.IsImmutable()).
		Msg("Clone is immutable")

	// Now we can modify the clone
	mutableClone.AddField("middle_name", schema.JString, nil)
	mutableClone.AddField("last_name", schema.JString, nil)

	// Freeze the clone to make it immutable
	immutableClone := mutableClone.Freeze()
	logger.App.Info().
		Bool("immutable", immutableClone.IsImmutable()).
		Msg("Frozen clone is immutable")

	// Example 6: Field immutability
	logger.App.Info().Msg("=== Example 6: Field Immutability ===")

	// Create a mutable schema
	mutableSchema := schema.NewSchemaBuilder("TestSchema").
		AddField("test_field", schema.JString, nil).
		Build()

	// Get the field and modify it
	testField, exists := mutableSchema.Field("test_field")
	if exists {
		testField.SetRequired(true)
		logger.App.Info().
			Bool("required", testField.IsRequired()).
			Msg("Field is required")

		// Freeze the field
		frozenField := testField.Freeze()
		logger.App.Info().
			Bool("immutable", frozenField.IsImmutable()).
			Msg("Frozen field is immutable")

		// Try to modify the frozen field (this will panic)
		defer func() {
			if r := recover(); r != nil {
				logger.App.Info().
					Interface("panic", r).
					Msg("Caught panic when modifying frozen field")
			}
		}()

		// This will panic
		// frozenField.SetUnique(true)
	}

	// Example 7: Using the repository with immutable schemas
	logger.App.Info().Msg("=== Example 7: Repository Usage ===")

	// Get repository
	userRepo, err := client.GetRepository("User")
	if err != nil {
		log.Fatal("Failed to get repository:", err)
	}

	// Create a user record
	user := schema.NewJRecord().
		Set("id", "user_001").
		Set("name", "John Doe").
		Set("email", "john.doe@example.com").
		Set("age", 30).
		Set("phone", "+1234567890").
		Set("address", "123 Main St")

	// Save user
	savedUser, err := userRepo.Save(context.Background(), user)
	if err != nil {
		log.Fatal("Failed to save user:", err)
	}

	logger.App.Info().
		Interface("user", savedUser.ToMap()).
		Msg("Saved user")

	// Find user by ID
	foundUser, err := userRepo.FindById(context.Background(), "user_001")
	if err != nil {
		log.Fatal("Failed to find user:", err)
	}

	logger.App.Info().
		Interface("user", foundUser.ToMap()).
		Msg("Found user")

	// Clean up
	err = userRepo.Delete(context.Background(), "user_001")
	if err != nil {
		log.Fatal("Failed to delete user:", err)
	}

	logger.App.Info().Msg("Immutable schema example completed successfully!")
}

// SchemaEvolutionExample demonstrates schema evolution patterns
func SchemaEvolutionExample() {
	// Initialize logger
	logger.Initialize()
	logger.App.Info().Msg("=== Schema Evolution Example ===")

	// Create MongoDB client
	client, err := mongo.NewMongoClient("mongodb://localhost:27017", "jpack_evolution_example")
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}
	defer client.Close(context.Background())

	// Version 1: Basic user schema
	logger.App.Info().Msg("Creating Version 1 schema...")
	v1Schema := schema.NewSchemaBuilder("User").
		AddRequiredUniqueField("id", schema.JString, nil).
		AddRequiredField("name", schema.JString, nil).
		AddRequiredUniqueField("email", schema.JString, nil).
		SetIDField("id").
		BuildImmutable()

	err = client.RegisterSchema(v1Schema)
	if err != nil {
		log.Fatal("Failed to register v1 schema:", err)
	}

	// Version 2: Add age field
	logger.App.Info().Msg("Creating Version 2 schema...")
	v2Schema := schema.NewSchemaBuilder("User").
		AddRequiredUniqueField("id", schema.JString, nil).
		AddRequiredField("name", schema.JString, nil).
		AddRequiredUniqueField("email", schema.JString, nil).
		AddField("age", schema.JInt, 18). // New field
		SetIDField("id").
		BuildImmutable()

	// Replace the schema
	err = client.GetDatabase().Collection("User").Drop(context.Background())
	if err != nil {
		log.Printf("Warning: Failed to drop collection: %v", err)
	}

	err = client.RegisterSchema(v2Schema)
	if err != nil {
		log.Fatal("Failed to register v2 schema:", err)
	}

	// Version 3: Add phone and address fields
	logger.App.Info().Msg("Creating Version 3 schema...")
	v3Schema := schema.NewSchemaBuilder("User").
		AddRequiredUniqueField("id", schema.JString, nil).
		AddRequiredField("name", schema.JString, nil).
		AddRequiredUniqueField("email", schema.JString, nil).
		AddField("age", schema.JInt, 18).
		AddField("phone", schema.JString, nil).   // New field
		AddField("address", schema.JString, nil). // New field
		SetIDField("id").
		BuildImmutable()

	// Replace the schema
	err = client.GetDatabase().Collection("User").Drop(context.Background())
	if err != nil {
		log.Printf("Warning: Failed to drop collection: %v", err)
	}

	err = client.RegisterSchema(v3Schema)
	if err != nil {
		log.Fatal("Failed to register v3 schema:", err)
	}

	logger.App.Info().Msg("Schema evolution completed successfully!")
}

// RunImmutableExamples runs all immutable schema examples
func RunImmutableExamples() {
	ImmutableSchemaExample()
	SchemaEvolutionExample()
}
