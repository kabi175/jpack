package main

import (
	"fmt"
	"log"

	"github.com/kabi175/jpack/schema"
)

// FileSchemaExample demonstrates loading schemas from JSON and YAML files
func FileSchemaExample() {
	fmt.Println("=== File Schema Source Example ===")

	// Test YAML schema loading
	fmt.Println("\n1. Loading YAML Schema:")
	yamlSource := schema.NewFileSchemaSource("examples/schemas/user_schema.yaml")
	userSchema, err := yamlSource.Load("User")
	if err != nil {
		log.Fatalf("Failed to load YAML schema: %v", err)
	}

	fmt.Printf("Loaded schema: %s\n", userSchema.Name())
	fmt.Printf("Fields: %d\n", len(userSchema.Fields()))
	for _, field := range userSchema.Fields() {
		fmt.Printf("  - %s: %s", field.Name(), field.Type())
		if field.IsRequired() {
			fmt.Print(" (required)")
		}
		if field.IsUnique() {
			fmt.Print(" (unique)")
		}
		fmt.Println()
	}

	// Test JSON schema loading
	fmt.Println("\n2. Loading JSON Schema:")
	jsonSource := schema.NewFileSchemaSource("examples/schemas/category_schema.json")
	categorySchema, err := jsonSource.Load("Category")
	if err != nil {
		log.Fatalf("Failed to load JSON schema: %v", err)
	}

	fmt.Printf("Loaded schema: %s\n", categorySchema.Name())
	fmt.Printf("Fields: %d\n", len(categorySchema.Fields()))
	for _, field := range categorySchema.Fields() {
		fmt.Printf("  - %s: %s", field.Name(), field.Type())
		if field.IsRequired() {
			fmt.Print(" (required)")
		}
		if field.IsUnique() {
			fmt.Print(" (unique)")
		}
		fmt.Println()
	}

	// Test with registry
	fmt.Println("\n3. Using with Lazy Registry:")
	registry := schema.NewJSchemaRegistry()

	// Register YAML schema lazily
	err = registry.RegisterLazy("Product", schema.NewFileSchemaSource("examples/schemas/product_schema.yaml"))
	if err != nil {
		log.Fatalf("Failed to register YAML schema: %v", err)
	}

	// Register JSON schema lazily
	err = registry.RegisterLazy("Order", schema.NewFileSchemaSource("examples/schemas/order_schema.yaml"))
	if err != nil {
		log.Fatalf("Failed to register JSON schema: %v", err)
	}

	// Load schemas
	productSchema, ok := registry.Get("Product")
	if ok {
		fmt.Printf("Lazy loaded Product schema: %s with %d fields\n", productSchema.Name(), len(productSchema.Fields()))
	}

	orderSchema, ok := registry.Get("Order")
	if ok {
		fmt.Printf("Lazy loaded Order schema: %s with %d fields\n", orderSchema.Name(), len(orderSchema.Fields()))
	}

	// List all loaded schemas
	loadedSchemas := registry.List()
	fmt.Printf("\nTotal loaded schemas: %d\n", len(loadedSchemas))
	for _, schema := range loadedSchemas {
		fmt.Printf("  - %s\n", schema.Name())
	}

	fmt.Println("\n=== File Schema Example Complete ===")
}
