package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/kabi175/jpack"
)

func main() {
	// Create a user schema with various field types
	userSchema := jpack.NewSchema("users").
		Field("id", &jpack.String{}).
		Field("name", &jpack.String{}).
		Field("age", &jpack.Number{}).
		Field("is_active", &jpack.Boolean{}).
		Field("created_at", &jpack.DateTime{}).
		FieldWithDefault("status", &jpack.String{}, "active").
		Build()

	// Create a new record
	record := jpack.NewMongoRecord(userSchema)

	// Helper function to get fields
	getField := func(name string) jpack.JField {
		field, exists := userSchema.Field(name)
		if !exists {
			log.Fatalf("Field %s not found", name)
		}
		return field
	}

	fmt.Println("=== Schema Validation Example ===")

	// Example 1: Valid record
	fmt.Println("1. Valid record:")
	record.SetValue(getField("id"), "123")
	record.SetValue(getField("name"), "John Doe")
	record.SetValue(getField("age"), 25)
	record.SetValue(getField("is_active"), true)
	record.SetValue(getField("created_at"), "2024-12-25T10:00:00Z")

	err := userSchema.Validate(context.Background(), record)
	if err != nil {
		fmt.Printf("   Validation failed: %v\n", err)
	} else {
		fmt.Println("   ✅ Validation passed")
	}

	// Example 2: Missing required field
	fmt.Println("\n2. Missing required field:")
	record2 := jpack.NewMongoRecord(userSchema)
	record2.SetValue(getField("id"), "456")
	record2.SetValue(getField("name"), "Jane Smith")
	// Missing age field

	err = userSchema.Validate(context.Background(), record2)
	if err != nil {
		fmt.Printf("   ❌ Validation failed: %v\n", err)
	} else {
		fmt.Println("   ✅ Validation passed")
	}

	// Example 3: Invalid field value
	fmt.Println("\n3. Invalid field value:")
	record3 := jpack.NewMongoRecord(userSchema)
	record3.SetValue(getField("id"), "789")
	record3.SetValue(getField("name"), "Bob Wilson")

	// Try to set invalid age (this will fail at SetValue level)
	err = record3.SetValue(getField("age"), "not a number")
	if err != nil {
		fmt.Printf("   ❌ SetValue failed: %v\n", err)
	}

	// Now validate - age field will be missing
	err = userSchema.Validate(context.Background(), record3)
	if err != nil {
		fmt.Printf("   ❌ Validation failed: %v\n", err)
	} else {
		fmt.Println("   ✅ Validation passed")
	}

	// Example 4: Record with default values
	fmt.Println("\n4. Record with default values:")
	record4 := jpack.NewMongoRecord(userSchema)
	record4.SetValue(getField("id"), "101")
	record4.SetValue(getField("name"), "Alice Johnson")
	record4.SetValue(getField("age"), 30)
	record4.SetValue(getField("is_active"), "yes") // Boolean accepts string
	record4.SetValue(getField("created_at"), "2024-12-25T15:30:00Z")
	// status field not set, will use default "active"

	err = userSchema.Validate(context.Background(), record4)
	if err != nil {
		fmt.Printf("   ❌ Validation failed: %v\n", err)
	} else {
		fmt.Println("   ✅ Validation passed")
	}

	// Example 5: Wrong schema
	fmt.Println("\n5. Wrong schema:")
	wrongSchema := jpack.NewSchema("wrong").Field("id", &jpack.String{}).Build()
	record5 := jpack.NewMongoRecord(wrongSchema)
	record5.SetValue(getField("id"), "999")

	err = userSchema.Validate(context.Background(), record5)
	if err != nil {
		fmt.Printf("   ❌ Validation failed: %v\n", err)
	} else {
		fmt.Println("   ✅ Validation passed")
	}

	fmt.Println("\n=== End of Examples ===")

	// Example 6: Custom validation
	fmt.Println("\n6. Custom validation example:")
	userSchemaWithValidation := jpack.NewSchema("users").
		Field("id", &jpack.String{}).
		Field("name", &jpack.String{}).
		Field("email", &jpack.String{}).
		Field("age", &jpack.Number{}).
		Validation(func(ctx context.Context, record jpack.JRecord) error {
			// Custom validation: name must be at least 3 characters
			nameField, _ := record.Schema().Field("name")
			if name, exists := record.Value(nameField); exists {
				if nameStr, ok := name.(string); ok {
					if len(strings.TrimSpace(nameStr)) < 3 {
						return fmt.Errorf("name must be at least 3 characters long")
					}
				}
			}
			return nil
		}).
		Validation(func(ctx context.Context, record jpack.JRecord) error {
			// Custom validation: email must contain @
			emailField, _ := record.Schema().Field("email")
			if email, exists := record.Value(emailField); exists {
				if emailStr, ok := email.(string); ok {
					if !strings.Contains(emailStr, "@") {
						return fmt.Errorf("email must contain @ symbol")
					}
				}
			}
			return nil
		}).
		Build()

	// Helper function to get fields from userSchemaWithValidation
	getValidationField := func(name string) jpack.JField {
		field, exists := userSchemaWithValidation.Field(name)
		if !exists {
			log.Fatalf("Field %s not found", name)
		}
		return field
	}

	// Test valid user with custom validation
	validUserWithValidation := jpack.NewMongoRecord(userSchemaWithValidation)
	validUserWithValidation.SetValue(getValidationField("id"), "201")
	validUserWithValidation.SetValue(getValidationField("name"), "John Doe")
	validUserWithValidation.SetValue(getValidationField("email"), "john@example.com")
	validUserWithValidation.SetValue(getValidationField("age"), 25)

	err = userSchemaWithValidation.Validate(context.Background(), validUserWithValidation)
	if err != nil {
		fmt.Printf("   ❌ Validation failed: %v\n", err)
	} else {
		fmt.Println("   ✅ Custom validation passed")
	}

	// Test invalid user with custom validation (short name)
	invalidUserWithValidation := jpack.NewMongoRecord(userSchemaWithValidation)
	invalidUserWithValidation.SetValue(getValidationField("id"), "202")
	invalidUserWithValidation.SetValue(getValidationField("name"), "Jo") // Too short
	invalidUserWithValidation.SetValue(getValidationField("email"), "jo@example.com")
	invalidUserWithValidation.SetValue(getValidationField("age"), 30)

	err = userSchemaWithValidation.Validate(context.Background(), invalidUserWithValidation)
	if err != nil {
		fmt.Printf("   ❌ Custom validation failed: %v\n", err)
	} else {
		fmt.Println("   ✅ Custom validation passed")
	}

	fmt.Println("\n=== End of All Examples ===")
}
