package main

import (
	"context"
	"fmt"

	"github.com/kabi175/jpack/schema"
)

func runImmutableSchemaExample() {
	fmt.Println("=== Immutable Schema Example ===")

	// 1. Create an immutable schema using SchemaBuilder
	fmt.Println("1. Creating immutable schema...")
	userSchema := schema.NewSchemaBuilder("User").
		AddField("name", schema.JString, nil).
		AddField("age", schema.JInt, nil).
		Build()

	fmt.Printf("Created schema '%s' with %d fields\n", userSchema.Name(), len(userSchema.Fields()))
	fmt.Printf("Is immutable: %v\n\n", userSchema.IsImmutable())

	// 2. Update schema using callback pattern (creates new instance)
	fmt.Println("2. Updating schema with new fields...")
	updatedSchema := userSchema.Update(func(sb *schema.SchemaBuilder) {
		sb.AddField("email", schema.JString, nil)
		sb.AddRequiredField("username", schema.JString, nil)
		sb.AddValidation(func(ctx context.Context, rec schema.JRecord) error {
			// Example validation: username must be present
			if !rec.Has("username") {
				return fmt.Errorf("username is required")
			}
			return nil
		})
	})

	fmt.Printf("Original schema fields: %d\n", len(userSchema.Fields()))
	fmt.Printf("Updated schema fields: %d\n\n", len(updatedSchema.Fields()))

	// 3. Multiple updates create multiple versions
	fmt.Println("3. Creating multiple schema versions...")
	v1Schema := userSchema // Version 1: id, name, age

	v2Schema := v1Schema.Update(func(sb *schema.SchemaBuilder) {
		sb.AddField("email", schema.JString, nil)
	})

	v3Schema := v2Schema.Update(func(sb *schema.SchemaBuilder) {
		sb.AddField("phone", schema.JString, nil)
		sb.AddField("address", schema.JString, nil)
	})

	fmt.Printf("V1 fields: %d\n", len(v1Schema.Fields()))
	fmt.Printf("V2 fields: %d\n", len(v2Schema.Fields()))
	fmt.Printf("V3 fields: %d\n\n", len(v3Schema.Fields()))

	// 4. Field immutability
	fmt.Println("4. Working with immutable fields...")
	nameField := schema.NewFieldBuilder("name", schema.JString, "").
		Required().
		Build()

	fmt.Printf("Name field required: %v\n", nameField.IsRequired())
	fmt.Printf("Name field unique: %v\n", nameField.IsUnique())

	// Update field to add unique constraint
	uniqueNameField := nameField.Update(func(fb *schema.FieldBuilder) {
		fb.Unique()
	})

	fmt.Printf("Original field unique: %v\n", nameField.IsUnique())
	fmt.Printf("Updated field unique: %v\n\n", uniqueNameField.IsUnique())

	// 5. Creating fields with SchemaBuilder
	fmt.Println("5. Creating schema with required and unique fields...")
	productSchema := schema.NewSchemaBuilder("Product").
		AddRequiredUniqueField("sku", schema.JString, nil).
		AddRequiredField("name", schema.JString, nil).
		AddField("price", schema.JFloat64, 0.0).
		AddField("in_stock", schema.JBool, true).
		Build()

	fmt.Printf("Product schema fields:\n")
	for _, field := range productSchema.Fields() {
		fmt.Printf("  - %s: %s", field.Name(), field.Type())
		if field.IsRequired() {
			fmt.Printf(" (required)")
		}
		if field.IsUnique() {
			fmt.Printf(" (unique)")
		}
		fmt.Println()
	}
	fmt.Println()

	// 6. Validation with immutable schemas
	fmt.Println("6. Validating records...")
	validationSchema := schema.NewSchemaBuilder("User").
		AddRequiredField("email", schema.JString, nil).
		AddRequiredField("age", schema.JInt, nil).
		AddValidation(func(ctx context.Context, rec schema.JRecord) error {
			age, ok := rec.Get("age").(int)
			if !ok {
				return fmt.Errorf("age must be an integer")
			}
			if age < 18 {
				return fmt.Errorf("age must be at least 18")
			}
			return nil
		}).
		Build()

	// Valid record
	validRecord := schema.NewJRecord().
		Set("id", "user123").
		Set("email", "user@example.com").
		Set("age", 25)

	if err := validationSchema.Validate(context.Background(), validRecord); err != nil {
		fmt.Printf("Validation failed: %v\n", err)
	} else {
		fmt.Println("✓ Valid record passed validation")
	}

	// Invalid record (age < 18)
	invalidRecord := schema.NewJRecord().
		Set("id", "teen123").
		Set("email", "teen@example.com").
		Set("age", 16)

	if err := validationSchema.Validate(context.Background(), invalidRecord); err != nil {
		fmt.Printf("✓ Invalid record failed validation: %v\n", err)
	}
	fmt.Println()

	// 7. Schema with references and edges
	fmt.Println("7. Creating schema with relationships...")

	profileSchema := schema.NewSchemaBuilder("Profile").
		AddField("bio", schema.JString, nil).
		AddField("avatar", schema.JString, nil).
		Build()

	orderSchema := schema.NewSchemaBuilder("Order").
		AddField("total", schema.JFloat64, 0.0).
		Build()

	userWithRefsSchema := schema.NewSchemaBuilder("User").
		AddField("name", schema.JString, nil).
		AddRef("profile", profileSchema).
		AddArrayRef("orders", orderSchema).
		AddEdge(schema.NewJEdge("user_orders", userSchema, orderSchema, schema.EdgeOneToMany)).
		Build()

	fmt.Printf("User schema with relationships:\n")
	fmt.Printf("  Fields: %d\n", len(userWithRefsSchema.Fields()))
	fmt.Printf("  Refs: %d\n", len(userWithRefsSchema.Refs()))
	fmt.Printf("  Edges: %d\n", len(userWithRefsSchema.Edges()))
	fmt.Println()

	fmt.Println("=== Example Complete ===")
}
