package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/kabi175/jpack/schema"
)

// LazyRegistryExample demonstrates the lazy-loaded schema registry
func LazyRegistryExample() {
	// Create a new schema registry
	registry := schema.NewJSchemaRegistry()

	// Example 1: Eager registration
	fmt.Println("=== Eager Registration Example ===")
	userSchema := schema.NewJSchema("User")
	userSchema.AddField("id", schema.JString, nil)
	userSchema.AddField("email", schema.JString, nil)
	userSchema.AddField("name", schema.JString, nil)

	err := registry.Register(userSchema)
	if err != nil {
		log.Fatalf("Failed to register user schema: %v", err)
	}

	// Retrieve the eagerly registered schema
	userSchemaRetrieved, ok := registry.Get("User")
	if ok {
		fmt.Printf("Loaded schema: %s\n", userSchemaRetrieved.Name())
		fmt.Printf("Fields: %d\n", len(userSchemaRetrieved.Fields()))
	}

	// Example 2: Lazy registration with file source
	fmt.Println("\n=== Lazy Registration Example ===")

	// Create a sample schema file
	orderSchemaDef := schema.SchemaDef{
		Name: "Order",
		Fields: map[string]schema.FieldDef{
			"id": {
				Type:     "string",
				Required: true,
			},
			"userId": {
				Type:     "string",
				Required: true,
			},
			"total": {
				Type:     "float64",
				Required: true,
			},
			"status": {
				Type:         "string",
				DefaultValue: "pending",
			},
		},
		Refs: map[string]schema.RefDef{
			"user": {
				TargetSchema: "User",
				IsArray:      false,
			},
		},
	}

	// Write the schema definition to a temporary file
	schemaData, err := json.MarshalIndent(orderSchemaDef, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal schema definition: %v", err)
	}

	tempFile := "temp_order_schema.json"
	err = os.WriteFile(tempFile, schemaData, 0644)
	if err != nil {
		log.Fatalf("Failed to write schema file: %v", err)
	}
	defer os.Remove(tempFile) // Clean up

	// Register the schema lazily
	fileSource := schema.NewFileSchemaSource(tempFile)
	err = registry.RegisterLazy("Order", fileSource)
	if err != nil {
		log.Fatalf("Failed to register lazy schema: %v", err)
	}

	fmt.Println("Order schema registered lazily")

	// First Get will trigger lazy load
	fmt.Println("First Get() call - this will trigger lazy loading:")
	orderSchema, ok := registry.Get("Order")
	if ok {
		fmt.Printf("Loaded schema: %s\n", orderSchema.Name())
		fmt.Printf("Fields: %d\n", len(orderSchema.Fields()))
		for _, field := range orderSchema.Fields() {
			fmt.Printf("  - %s: %s", field.Name(), field.Type())
			if field.IsRequired() {
				fmt.Print(" (required)")
			}
			fmt.Println()
		}
	} else {
		fmt.Println("Failed to load Order schema")
	}

	// Second Get should return immediately (already loaded)
	fmt.Println("\nSecond Get() call - should return immediately:")
	orderSchema2, ok := registry.Get("Order")
	if ok {
		fmt.Printf("Retrieved schema: %s (already loaded)\n", orderSchema2.Name())
	} else {
		fmt.Println("Failed to retrieve Order schema")
	}

	// Example 3: List loaded schemas
	fmt.Println("\n=== List Loaded Schemas ===")
	loadedSchemas := registry.List()
	fmt.Printf("Loaded schemas count: %d\n", len(loadedSchemas))
	for _, schema := range loadedSchemas {
		fmt.Printf("  - %s\n", schema.Name())
	}

	// Example 4: HTTP and Database sources (placeholder)
	fmt.Println("\n=== Other Source Types ===")

	// HTTP source (placeholder)
	httpSource := schema.NewHTTPSchemaSource("https://api.example.com/schemas/Product")
	fmt.Printf("HTTP source created: %s\n", httpSource.URL)

	// Database source (placeholder)
	dbSource := schema.NewDatabaseSchemaSource(nil, "schema_definitions")
	fmt.Printf("Database source created for table: %s\n", dbSource.Table)

	fmt.Println("\n=== Lazy Registry Example Complete ===")
}
