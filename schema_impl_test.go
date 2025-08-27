package jpack

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSchemaImpl_Validate(t *testing.T) {
	// Create a test schema
	schema := NewSchema("test").
		Field("id", &String{}).
		Field("name", &String{}).
		Field("age", &Number{}).
		FieldWithDefault("status", &String{}, "active").
		Build()

	t.Run("valid record", func(t *testing.T) {
		record := NewMongoRecord(schema)

		idField, _ := schema.Field("id")
		nameField, _ := schema.Field("name")
		ageField, _ := schema.Field("age")

		record.SetValue(idField, "123")
		record.SetValue(nameField, "John Doe")
		record.SetValue(ageField, 25)

		err := schema.Validate(context.Background(), record)
		assert.NoError(t, err)
	})

	t.Run("record with missing required field", func(t *testing.T) {
		record := NewMongoRecord(schema)

		idField, _ := schema.Field("id")
		nameField, _ := schema.Field("name")
		// Missing age field

		record.SetValue(idField, "123")
		record.SetValue(nameField, "John Doe")

		err := schema.Validate(context.Background(), record)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "required field 'age' is missing")
	})

	t.Run("record with invalid field value", func(t *testing.T) {
		record := NewMongoRecord(schema)

		idField, _ := schema.Field("id")
		nameField, _ := schema.Field("name")
		ageField, _ := schema.Field("age")

		record.SetValue(idField, "123")
		record.SetValue(nameField, "John Doe")
		// Try to set invalid age value - this should fail at SetValue level
		err := record.SetValue(ageField, "not a number")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "value is not a valid integer")

		// Now validate the record - age field should be missing
		err = schema.Validate(context.Background(), record)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "required field 'age' is missing")
	})

	t.Run("record with default value", func(t *testing.T) {
		record := NewMongoRecord(schema)

		idField, _ := schema.Field("id")
		nameField, _ := schema.Field("name")
		ageField, _ := schema.Field("age")
		// status field not set, should use default

		record.SetValue(idField, "123")
		record.SetValue(nameField, "John Doe")
		record.SetValue(ageField, 25)

		err := schema.Validate(context.Background(), record)
		assert.NoError(t, err)
	})

	t.Run("record with wrong schema", func(t *testing.T) {
		wrongSchema := NewSchema("wrong").Field("id", &String{}).Build()
		record := NewMongoRecord(wrongSchema)

		idField, _ := wrongSchema.Field("id")
		record.SetValue(idField, "123")

		err := schema.Validate(context.Background(), record)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "record schema does not match validation schema")
	})

	t.Run("empty record with defaults", func(t *testing.T) {
		record := NewMongoRecord(schema)

		idField, _ := schema.Field("id")
		nameField, _ := schema.Field("name")
		ageField, _ := schema.Field("age")

		record.SetValue(idField, "123")
		record.SetValue(nameField, "John Doe")
		record.SetValue(ageField, 25)
		// status field not set, should use default "active"

		err := schema.Validate(context.Background(), record)
		assert.NoError(t, err)
	})
}

func TestSchemaImpl_ValidateWithBoolean(t *testing.T) {
	schema := NewSchema("test").
		Field("id", &String{}).
		Field("is_active", &Boolean{}).
		Build()

	t.Run("valid boolean values", func(t *testing.T) {
		record := NewMongoRecord(schema)

		idField, _ := schema.Field("id")
		isActiveField, _ := schema.Field("is_active")

		record.SetValue(idField, "123")
		record.SetValue(isActiveField, true)

		err := schema.Validate(context.Background(), record)
		assert.NoError(t, err)
	})

	t.Run("boolean with string value", func(t *testing.T) {
		record := NewMongoRecord(schema)

		idField, _ := schema.Field("id")
		isActiveField, _ := schema.Field("is_active")

		record.SetValue(idField, "123")
		record.SetValue(isActiveField, "yes") // Boolean accepts string representations

		err := schema.Validate(context.Background(), record)
		assert.NoError(t, err)
	})

	t.Run("invalid boolean value", func(t *testing.T) {
		record := NewMongoRecord(schema)

		idField, _ := schema.Field("id")
		isActiveField, _ := schema.Field("is_active")

		record.SetValue(idField, "123")
		// Try to set invalid boolean value - this should fail at SetValue level
		err := record.SetValue(isActiveField, "invalid_boolean")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid boolean string value")

		// Now validate the record - is_active field should be missing
		err = schema.Validate(context.Background(), record)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "required field 'is_active' is missing")
	})
}

func TestSchemaImpl_ValidateWithDateTime(t *testing.T) {
	schema := NewSchema("test").
		Field("id", &String{}).
		Field("created_at", &DateTime{}).
		Build()

	t.Run("valid datetime values", func(t *testing.T) {
		record := NewMongoRecord(schema)

		idField, _ := schema.Field("id")
		createdAtField, _ := schema.Field("created_at")

		record.SetValue(idField, "123")
		record.SetValue(createdAtField, "2024-12-25T10:00:00Z")

		err := schema.Validate(context.Background(), record)
		assert.NoError(t, err)
	})

	t.Run("invalid datetime format", func(t *testing.T) {
		record := NewMongoRecord(schema)

		idField, _ := schema.Field("id")
		createdAtField, _ := schema.Field("created_at")

		record.SetValue(idField, "123")
		// Try to set invalid datetime format - this should fail at SetValue level
		err := record.SetValue(createdAtField, "2024-12-25 10:00:00") // Invalid format
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "value is not a valid RFC3339 datetime string")

		// Now validate the record - created_at field should be missing
		err = schema.Validate(context.Background(), record)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "required field 'created_at' is missing")
	})
}

func TestSchemaImpl_ValidateWithOptions(t *testing.T) {
	// Create a mock options service for testing
	mockService := &mockOptionsService{
		options: []Option{
			{UniqueName: "active", DisplayName: "Active"},
			{UniqueName: "pending", DisplayName: "Pending"},
			{UniqueName: "completed", DisplayName: "Completed"},
		},
	}

	optionsField := NewOptions(mockService)

	schema := NewSchema("test").
		Field("id", &String{}).
		Field("status", optionsField).
		Build()

	t.Run("valid options value", func(t *testing.T) {
		record := NewMongoRecord(schema)

		idField, _ := schema.Field("id")
		statusField, _ := schema.Field("status")

		record.SetValue(idField, "123")
		record.SetValue(statusField, "active")

		err := schema.Validate(context.Background(), record)
		assert.NoError(t, err)
	})

	t.Run("invalid options value", func(t *testing.T) {
		record := NewMongoRecord(schema)

		idField, _ := schema.Field("id")
		statusField, _ := schema.Field("status")

		record.SetValue(idField, "123")
		// Try to set invalid option - this should fail at SetValue level
		err := record.SetValue(statusField, "invalid_status")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not in the list of available options")

		// Now validate the record - status field should be missing
		err = schema.Validate(context.Background(), record)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "required field 'status' is missing")
	})
}

func TestSchemaImpl_CustomValidation(t *testing.T) {
	schema := NewSchema("test").
		Field("id", &String{}).
		Field("name", &String{}).
		Field("age", &Number{}).
		Build()

	t.Run("schema with custom validation", func(t *testing.T) {
		// Add custom validation that requires name to be at least 3 characters
		schema.AddValidation(func(ctx context.Context, record JRecord) error {
			nameField, _ := schema.Field("name")
			if name, exists := record.Value(nameField); exists {
				if nameStr, ok := name.(string); ok {
					if len(nameStr) < 3 {
						return fmt.Errorf("name must be at least 3 characters long")
					}
				}
			}
			return nil
		})

		// Add another custom validation that requires age to be positive
		schema.AddValidation(func(ctx context.Context, record JRecord) error {
			ageField, _ := schema.Field("age")
			if age, exists := record.Value(ageField); exists {
				if ageNum, ok := age.(int); ok {
					if ageNum <= 0 {
						return fmt.Errorf("age must be positive")
					}
				}
			}
			return nil
		})

		// Test valid record
		record := NewMongoRecord(schema)
		idField, _ := schema.Field("id")
		nameField, _ := schema.Field("name")
		ageField, _ := schema.Field("age")

		record.SetValue(idField, "123")
		record.SetValue(nameField, "John Doe")
		record.SetValue(ageField, 25)

		err := schema.Validate(context.Background(), record)
		assert.NoError(t, err)

		// Test invalid name (too short)
		record2 := NewMongoRecord(schema)
		record2.SetValue(idField, "456")
		record2.SetValue(nameField, "Jo") // Too short
		record2.SetValue(ageField, 30)

		err = schema.Validate(context.Background(), record2)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "custom validation 1 failed")
		assert.Contains(t, err.Error(), "name must be at least 3 characters long")

		// Test invalid age (negative)
		record3 := NewMongoRecord(schema)
		record3.SetValue(idField, "789")
		record3.SetValue(nameField, "Jane Smith")
		record3.SetValue(ageField, -5)

		err = schema.Validate(context.Background(), record3)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "custom validation 2 failed")
		assert.Contains(t, err.Error(), "age must be positive")
	})

	t.Run("schema builder with validation", func(t *testing.T) {
		// Create schema with validation using builder pattern
		builder := NewSchema("test2").
			Field("id", &String{}).
			Field("email", &String{})

		schemaWithValidation := builder.
			Validation(func(ctx context.Context, record JRecord) error {
				emailField, _ := builder.schema.Field("email")
				if email, exists := record.Value(emailField); exists {
					if emailStr, ok := email.(string); ok {
						if !strings.Contains(emailStr, "@") {
							return fmt.Errorf("email must contain @ symbol")
						}
					}
				}
				return nil
			}).
			Validation(func(ctx context.Context, record JRecord) error {
				emailField, _ := builder.schema.Field("email")
				if email, exists := record.Value(emailField); exists {
					if emailStr, ok := email.(string); ok {
						if !strings.Contains(emailStr, ".") {
							return fmt.Errorf("email must contain domain")
						}
					}
				}
				return nil
			}).
			Build()

		// Test valid email
		record := NewMongoRecord(schemaWithValidation)
		idField, _ := schemaWithValidation.Field("id")
		emailField, _ := schemaWithValidation.Field("email")

		record.SetValue(idField, "123")
		record.SetValue(emailField, "test@example.com")

		err := schemaWithValidation.Validate(context.Background(), record)
		assert.NoError(t, err)

		// Test invalid email (no @)
		record2 := NewMongoRecord(schemaWithValidation)
		record2.SetValue(idField, "456")
		record2.SetValue(emailField, "invalid-email")

		err = schemaWithValidation.Validate(context.Background(), record2)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "custom validation 1 failed")
		assert.Contains(t, err.Error(), "email must contain @ symbol")

		// Test invalid email (no domain)
		record3 := NewMongoRecord(schemaWithValidation)
		record3.SetValue(idField, "789")
		record3.SetValue(emailField, "test@")

		err = schemaWithValidation.Validate(context.Background(), record3)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "custom validation 2 failed")
		assert.Contains(t, err.Error(), "email must contain domain")
	})

	t.Run("validations method", func(t *testing.T) {
		// Test that Validations() returns the correct number of validations
		validations := schema.Validations()
		assert.Len(t, validations, 2) // We added 2 validations in the first test

		// Test that we can add more validations after building
		schema.AddValidation(func(ctx context.Context, record JRecord) error {
			return fmt.Errorf("test validation")
		})

		validations = schema.Validations()
		assert.Len(t, validations, 3)
	})
}

// mockOptionsService implements the OptionService interface for testing
type mockOptionsService struct {
	options []Option
}

func (m *mockOptionsService) GetOptions(ctx context.Context) ([]Option, error) {
	return m.options, nil
}
