package schema

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewJSchema(t *testing.T) {
	tests := []struct {
		name       string
		schemaName string
		expected   string
	}{
		{"valid name", "User", "User"},
		{"empty name", "", ""},
		{"special characters", "User-Profile_v2", "User-Profile_v2"},
		{"numbers", "Schema123", "Schema123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := NewJSchema(tt.schemaName)
			assert.Equal(t, tt.expected, schema.Name())
			assert.False(t, schema.IsImmutable())
			assert.Empty(t, schema.Fields())
			assert.Empty(t, schema.Refs())
			assert.Empty(t, schema.Edges())
			assert.Empty(t, schema.Validations())
			assert.Nil(t, schema.GetIDField())
		})
	}
}

func TestNewImmutableJSchema(t *testing.T) {
	schema := NewImmutableJSchema("TestSchema")

	assert.Equal(t, "TestSchema", schema.Name())
	assert.True(t, schema.IsImmutable())
	assert.Empty(t, schema.Fields())
	assert.Empty(t, schema.Refs())
	assert.Empty(t, schema.Edges())
	assert.Empty(t, schema.Validations())
	assert.Nil(t, schema.GetIDField())
}

func TestJSchema_AddField(t *testing.T) {
	tests := []struct {
		name         string
		fieldName    string
		fieldType    JFieldType
		defaultValue any
		expectedType JFieldType
	}{
		{"string field", "name", JString, "default", JString},
		{"int field", "age", JInt, 25, JInt},
		{"float field", "price", JFloat64, 99.99, JFloat64},
		{"bool field", "active", JBool, true, JBool},
		{"nil default", "description", JString, nil, JString},
		{"empty string", "title", JString, "", JString},
		{"zero value", "count", JInt, 0, JInt},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := NewJSchema("TestSchema")
			field := schema.AddField(tt.fieldName, tt.fieldType, tt.defaultValue)

			assert.Equal(t, tt.fieldName, field.Name())
			assert.Equal(t, tt.expectedType, field.Type())
			assert.Equal(t, tt.defaultValue, field.DefaultValue())
			assert.False(t, field.IsRequired())
			assert.False(t, field.IsUnique())
			assert.False(t, field.IsImmutable())

			// Check that field was added to schema
			retrievedField, exists := schema.Field(tt.fieldName)
			assert.True(t, exists)
			assert.Equal(t, field, retrievedField)

			// Check that it's the first field and becomes ID field
			if len(schema.Fields()) == 1 {
				assert.Equal(t, field, schema.GetIDField())
			}
		})
	}
}

func TestJSchema_AddField_Immutable(t *testing.T) {
	schema := NewImmutableJSchema("TestSchema")

	assert.Panics(t, func() {
		schema.AddField("name", JString, "default")
	}, "should panic when adding field to immutable schema")
}

func TestJSchema_AddField_Duplicate(t *testing.T) {
	schema := NewJSchema("TestSchema")

	// Add first field
	field1 := schema.AddField("name", JString, "default")

	// Add field with same name (should overwrite)
	field2 := schema.AddField("name", JInt, 42)

	assert.Equal(t, "name", field2.Name())
	assert.Equal(t, JInt, field2.Type())
	assert.Equal(t, 42, field2.DefaultValue())

	// Should only have one field
	assert.Len(t, schema.Fields(), 1)

	// ID field should still be the first field added
	assert.Equal(t, field1, schema.GetIDField())
}

func TestJSchema_Field(t *testing.T) {
	schema := NewJSchema("TestSchema")

	// Test non-existent field
	field, exists := schema.Field("nonexistent")
	assert.False(t, exists)
	assert.Nil(t, field)

	// Add field and test retrieval
	addedField := schema.AddField("name", JString, "default")
	retrievedField, exists := schema.Field("name")

	assert.True(t, exists)
	assert.Equal(t, addedField, retrievedField)
}

func TestJSchema_AddRef(t *testing.T) {
	userSchema := NewJSchema("User")
	orderSchema := NewJSchema("Order")

	// Add reference
	ref := orderSchema.AddRef("user", userSchema)

	assert.Equal(t, "user", ref.Name())
	assert.Equal(t, userSchema, ref.TargetSchema())
	assert.False(t, ref.IsArray())

	// Check that ref was added to schema
	retrievedRef, exists := orderSchema.Ref("user")
	assert.True(t, exists)
	assert.Equal(t, ref, retrievedRef)
}

func TestJSchema_AddRef_Immutable(t *testing.T) {
	schema := NewImmutableJSchema("TestSchema")
	targetSchema := NewJSchema("Target")

	assert.Panics(t, func() {
		schema.AddRef("target", targetSchema)
	}, "should panic when adding ref to immutable schema")
}

func TestJSchema_AddRef_Duplicate(t *testing.T) {
	schema := NewJSchema("TestSchema")
	target1 := NewJSchema("Target1")
	target2 := NewJSchema("Target2")

	// Add first ref
	schema.AddRef("target", target1)

	// Add ref with same name (should overwrite)
	ref2 := schema.AddRef("target", target2)

	assert.Equal(t, "target", ref2.Name())
	assert.Equal(t, target2, ref2.TargetSchema())

	// Should only have one ref
	assert.Len(t, schema.Refs(), 1)
}

func TestJSchema_Ref(t *testing.T) {
	schema := NewJSchema("TestSchema")

	// Test non-existent ref
	ref, exists := schema.Ref("nonexistent")
	assert.False(t, exists)
	assert.Nil(t, ref)

	// Add ref and test retrieval
	targetSchema := NewJSchema("Target")
	addedRef := schema.AddRef("target", targetSchema)
	retrievedRef, exists := schema.Ref("target")

	assert.True(t, exists)
	assert.Equal(t, addedRef, retrievedRef)
}

func TestJSchema_AddEdge(t *testing.T) {
	userSchema := NewJSchema("User")
	orderSchema := NewJSchema("Order")

	edge := NewJEdge("user_orders", userSchema, orderSchema, EdgeOneToMany)

	// Add edge
	result := userSchema.AddEdge(edge)

	// Should return self for chaining
	assert.Equal(t, userSchema, result)

	// Check that edge was added
	edges := userSchema.Edges()
	assert.Len(t, edges, 1)
	assert.Equal(t, edge, edges[0])
}

func TestJSchema_AddEdge_Immutable(t *testing.T) {
	schema := NewImmutableJSchema("TestSchema")
	targetSchema := NewJSchema("Target")
	edge := NewJEdge("test_edge", schema, targetSchema, EdgeOneToOne)

	assert.Panics(t, func() {
		schema.AddEdge(edge)
	}, "should panic when adding edge to immutable schema")
}

func TestJSchema_AddEdge_Multiple(t *testing.T) {
	userSchema := NewJSchema("User")
	orderSchema := NewJSchema("Order")
	productSchema := NewJSchema("Product")

	// Add multiple edges
	edge1 := NewJEdge("user_orders", userSchema, orderSchema, EdgeOneToMany)
	edge2 := NewJEdge("user_products", userSchema, productSchema, EdgeManyToMany)

	userSchema.AddEdge(edge1).AddEdge(edge2)

	edges := userSchema.Edges()
	assert.Len(t, edges, 2)
	assert.Contains(t, edges, edge1)
	assert.Contains(t, edges, edge2)
}

func TestJSchema_SetIDField(t *testing.T) {
	schema := NewJSchema("TestSchema")

	// Add fields
	nameField := schema.AddField("name", JString, "default")
	ageField := schema.AddField("age", JInt, 0)

	// Initially, first field should be ID field
	assert.Equal(t, nameField, schema.GetIDField())

	// Set different field as ID
	result := schema.SetIDField(ageField)

	// Should return self for chaining
	assert.Equal(t, schema, result)

	// Check that ID field was changed
	assert.Equal(t, ageField, schema.GetIDField())
}

func TestJSchema_SetIDField_Immutable(t *testing.T) {
	// Create mutable schema first, then make it immutable
	schema := NewJSchema("TestSchema")
	field := schema.AddField("name", JString, "default")
	schema.Freeze() // Make it immutable

	assert.Panics(t, func() {
		schema.SetIDField(field)
	}, "should panic when setting ID field on immutable schema")
}

func TestJSchema_SetIDField_NonExistent(t *testing.T) {
	schema := NewJSchema("TestSchema")
	otherSchema := NewJSchema("Other")
	field := otherSchema.AddField("name", JString, "default")

	// This should work but the field won't be in this schema
	result := schema.SetIDField(field)
	assert.Equal(t, schema, result)
	assert.Equal(t, field, schema.GetIDField())
}

func TestJSchema_AddValidation(t *testing.T) {
	schema := NewJSchema("TestSchema")

	// Add validation function
	validationFunc := func(ctx context.Context, rec JRecord) error {
		return nil
	}

	schema.AddValidation(validationFunc)

	validations := schema.Validations()
	assert.Len(t, validations, 1)
	assert.NotNil(t, validations[0])
	// Functions cannot be compared directly, so we just verify it's not nil
}

func TestJSchema_AddValidation_Immutable(t *testing.T) {
	schema := NewImmutableJSchema("TestSchema")

	validationFunc := func(ctx context.Context, rec JRecord) error {
		return nil
	}

	assert.Panics(t, func() {
		schema.AddValidation(validationFunc)
	}, "should panic when adding validation to immutable schema")
}

func TestJSchema_AddValidation_Multiple(t *testing.T) {
	schema := NewJSchema("TestSchema")

	// Add multiple validations
	validation1 := func(ctx context.Context, rec JRecord) error { return nil }
	validation2 := func(ctx context.Context, rec JRecord) error { return nil }

	schema.AddValidation(validation1)
	schema.AddValidation(validation2)

	validations := schema.Validations()
	assert.Len(t, validations, 2)
	assert.NotNil(t, validations[0])
	assert.NotNil(t, validations[1])
	// Functions cannot be compared directly, so we just verify they're not nil
}

func TestJSchema_Validate(t *testing.T) {
	schema := NewJSchema("TestSchema")

	// Add required field
	schema.AddField("name", JString, "default").SetRequired(true)

	// Add optional field
	schema.AddField("age", JInt, 0)

	tests := []struct {
		name      string
		record    JRecord
		wantError bool
		errorMsg  string
	}{
		{
			name: "valid record with required field",
			record: NewJRecordFromMap(map[string]any{
				"name": "John",
				"age":  25,
			}),
			wantError: false,
		},
		{
			name: "valid record with only required field",
			record: NewJRecordFromMap(map[string]any{
				"name": "Jane",
			}),
			wantError: false,
		},
		{
			name: "invalid record missing required field",
			record: NewJRecordFromMap(map[string]any{
				"age": 25,
			}),
			wantError: true,
			errorMsg:  "required field 'name' is missing",
		},
		{
			name:      "empty record",
			record:    NewJRecord(),
			wantError: true,
			errorMsg:  "required field 'name' is missing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			err := schema.Validate(ctx, tt.record)

			if tt.wantError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestJSchema_Validate_WithFieldValidation(t *testing.T) {
	schema := NewJSchema("TestSchema")

	// Add field with validation
	field := schema.AddField("age", JInt, 0)
	field.SetValidation(func(ctx context.Context, rec JRecord) error {
		age := rec.Get("age")
		if age == nil {
			return nil // Let required validation handle this
		}
		if ageInt, ok := age.(int); ok && ageInt < 0 {
			return assert.AnError
		}
		return nil
	})

	tests := []struct {
		name      string
		record    JRecord
		wantError bool
	}{
		{
			name: "valid age",
			record: NewJRecordFromMap(map[string]any{
				"age": 25,
			}),
			wantError: false,
		},
		{
			name: "zero age",
			record: NewJRecordFromMap(map[string]any{
				"age": 0,
			}),
			wantError: false,
		},
		{
			name: "negative age",
			record: NewJRecordFromMap(map[string]any{
				"age": -5,
			}),
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			err := schema.Validate(ctx, tt.record)

			if tt.wantError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "field 'age' validation failed")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestJSchema_Validate_WithSchemaValidation(t *testing.T) {
	schema := NewJSchema("TestSchema")

	// Add schema-level validation
	schema.AddValidation(func(ctx context.Context, rec JRecord) error {
		name := rec.Get("name")
		age := rec.Get("age")

		if name != nil && age != nil {
			if nameStr, ok := name.(string); ok && nameStr == "admin" {
				if ageInt, ok := age.(int); ok && ageInt < 18 {
					return assert.AnError
				}
			}
		}
		return nil
	})

	tests := []struct {
		name      string
		record    JRecord
		wantError bool
	}{
		{
			name: "valid admin",
			record: NewJRecordFromMap(map[string]any{
				"name": "admin",
				"age":  25,
			}),
			wantError: false,
		},
		{
			name: "invalid admin age",
			record: NewJRecordFromMap(map[string]any{
				"name": "admin",
				"age":  16,
			}),
			wantError: true,
		},
		{
			name: "non-admin with low age",
			record: NewJRecordFromMap(map[string]any{
				"name": "user",
				"age":  16,
			}),
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			err := schema.Validate(ctx, tt.record)

			if tt.wantError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "schema validation failed")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestJSchema_Clone(t *testing.T) {
	// Create original schema
	original := NewJSchema("Original")

	// Add fields
	nameField := original.AddField("name", JString, "default").SetRequired(true)
	original.AddField("age", JInt, 0).SetUnique(true)

	// Add validation
	original.AddValidation(func(ctx context.Context, rec JRecord) error {
		return nil
	})

	// Add ref
	targetSchema := NewJSchema("Target")
	original.AddRef("target", targetSchema)

	// Add edge
	edge := NewJEdge("test_edge", original, targetSchema, EdgeOneToOne)
	original.AddEdge(edge)

	// Clone schema
	cloned := original.Clone()

	// Test basic properties
	assert.Equal(t, original.Name(), cloned.Name())
	assert.False(t, cloned.IsImmutable()) // Clone should be mutable

	// Test fields
	originalFields := original.Fields()
	clonedFields := cloned.Fields()
	assert.Len(t, clonedFields, len(originalFields))

	// Check that fields are copied but are different objects
	// Create a map of cloned fields by name for easier lookup
	clonedFieldsMap := make(map[string]JField)
	for _, field := range clonedFields {
		clonedFieldsMap[field.Name()] = field
	}

	for _, originalField := range originalFields {
		clonedField, exists := clonedFieldsMap[originalField.Name()]
		assert.True(t, exists, "Cloned field %s should exist", originalField.Name())
		assert.Equal(t, originalField.Name(), clonedField.Name())
		assert.Equal(t, originalField.Type(), clonedField.Type())
		assert.Equal(t, originalField.DefaultValue(), clonedField.DefaultValue())
		assert.Equal(t, originalField.IsRequired(), clonedField.IsRequired())
		assert.Equal(t, originalField.IsUnique(), clonedField.IsUnique())
		assert.False(t, clonedField.IsImmutable()) // Cloned fields should be mutable
	}

	// Test ID field
	assert.NotNil(t, cloned.GetIDField())
	assert.Equal(t, nameField.Name(), cloned.GetIDField().Name())

	// Test refs
	originalRefs := original.Refs()
	clonedRefs := cloned.Refs()
	assert.Len(t, clonedRefs, len(originalRefs))

	// Test edges
	originalEdges := original.Edges()
	clonedEdges := cloned.Edges()
	assert.Len(t, clonedEdges, len(originalEdges))

	// Test validations
	originalValidations := original.Validations()
	clonedValidations := cloned.Validations()
	assert.Len(t, clonedValidations, len(originalValidations))

	// Test that modifications to clone don't affect original
	cloned.AddField("new_field", JString, "new")
	assert.Len(t, original.Fields(), 2)
	assert.Len(t, cloned.Fields(), 3)
}

func TestJSchema_Clone_Immutable(t *testing.T) {
	// Create mutable schema first, then make it immutable
	original := NewJSchema("Original")
	original.AddField("name", JString, "default")
	original.Freeze() // Make it immutable

	// Clone should be mutable
	cloned := original.Clone()
	assert.False(t, cloned.IsImmutable())

	// Should be able to modify clone
	assert.NotPanics(t, func() {
		cloned.AddField("age", JInt, 0)
	})
}

func TestJSchema_Freeze(t *testing.T) {
	schema := NewJSchema("TestSchema")

	// Add field
	field := schema.AddField("name", JString, "default")

	// Freeze schema
	result := schema.Freeze()

	// Should return self for chaining
	assert.Equal(t, schema, result)

	// Schema should be immutable
	assert.True(t, schema.IsImmutable())

	// Field should also be immutable
	assert.True(t, field.IsImmutable())

	// Should panic when trying to modify
	assert.Panics(t, func() {
		schema.AddField("age", JInt, 0)
	})

	assert.Panics(t, func() {
		field.SetRequired(true)
	})
}

func TestJSchema_Freeze_AlreadyImmutable(t *testing.T) {
	schema := NewImmutableJSchema("TestSchema")

	// Freezing already immutable schema should work
	result := schema.Freeze()
	assert.Equal(t, schema, result)
	assert.True(t, schema.IsImmutable())
}

func TestJSchema_String(t *testing.T) {
	schema := NewJSchema("User")

	// Add fields
	schema.AddField("id", JString, "").SetRequired(true).SetUnique(true)
	schema.AddField("name", JString, "").SetRequired(true)
	schema.AddField("age", JInt, 0)

	// Add ref
	targetSchema := NewJSchema("Profile")
	schema.AddRef("profile", targetSchema)

	// Get string representation
	str := schema.String()

	// Check that it contains expected information
	assert.Contains(t, str, "Schema: User")
	assert.Contains(t, str, "Fields:")
	assert.Contains(t, str, "id: string (required) (unique)")
	assert.Contains(t, str, "name: string (required)")
	assert.Contains(t, str, "age: int")
	assert.Contains(t, str, "References:")
	assert.Contains(t, str, "profile -> Profile")
}

func TestJSchema_String_Empty(t *testing.T) {
	schema := NewJSchema("Empty")
	str := schema.String()

	assert.Contains(t, str, "Schema: Empty")
	assert.Contains(t, str, "Fields:")
	// Should not contain References section
	assert.NotContains(t, str, "References:")
}

func TestJSchema_EdgeCases(t *testing.T) {
	t.Run("empty schema name", func(t *testing.T) {
		schema := NewJSchema("")
		assert.Equal(t, "", schema.Name())
	})

	t.Run("schema with only edges", func(t *testing.T) {
		schema := NewJSchema("EdgeOnly")
		targetSchema := NewJSchema("Target")
		edge := NewJEdge("test_edge", schema, targetSchema, EdgeOneToOne)
		schema.AddEdge(edge)

		assert.Len(t, schema.Fields(), 0)
		assert.Len(t, schema.Edges(), 1)
		assert.Nil(t, schema.GetIDField())
	})

	t.Run("schema with only refs", func(t *testing.T) {
		schema := NewJSchema("RefOnly")
		targetSchema := NewJSchema("Target")
		schema.AddRef("target", targetSchema)

		assert.Len(t, schema.Fields(), 0)
		assert.Len(t, schema.Refs(), 1)
		assert.Nil(t, schema.GetIDField())
	})

	t.Run("schema with only validations", func(t *testing.T) {
		schema := NewJSchema("ValidationOnly")
		schema.AddValidation(func(ctx context.Context, rec JRecord) error {
			return nil
		})

		assert.Len(t, schema.Fields(), 0)
		assert.Len(t, schema.Validations(), 1)
	})

	t.Run("set nil ID field", func(t *testing.T) {
		schema := NewJSchema("Test")
		schema.AddField("name", JString, "default")

		// Set ID field to nil should panic or handle gracefully
		assert.Panics(t, func() {
			schema.SetIDField(nil)
		}, "should panic when setting nil ID field")
	})
}

func TestJSchema_ConcurrentAccess(t *testing.T) {
	schema := NewJSchema("Concurrent")

	// Test concurrent field additions
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(index int) {
			defer func() { done <- true }()

			fieldName := fmt.Sprintf("field_%d", index)
			schema.AddField(fieldName, JString, "default")
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Should have 10 fields
	assert.Len(t, schema.Fields(), 10)
}

func TestJSchema_Validation_ContextCancellation(t *testing.T) {
	schema := NewJSchema("TestSchema")
	schema.AddField("name", JString, "default").SetRequired(true)

	// Add validation that respects context
	schema.AddValidation(func(ctx context.Context, rec JRecord) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			return nil
		}
	})

	record := NewJRecordFromMap(map[string]any{
		"name": "test",
	})

	// Test with cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := schema.Validate(ctx, record)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context canceled")
}

func TestJSchema_Validation_ContextTimeout(t *testing.T) {
	schema := NewJSchema("TestSchema")
	schema.AddField("name", JString, "default").SetRequired(true)

	// Add validation that respects context
	schema.AddValidation(func(ctx context.Context, rec JRecord) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			return nil
		}
	})

	record := NewJRecordFromMap(map[string]any{
		"name": "test",
	})

	// Test with timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 0)
	defer cancel()

	err := schema.Validate(ctx, record)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context deadline exceeded")
}

func TestJSchema_ComplexScenario(t *testing.T) {
	// Create a complex schema with all features
	userSchema := NewJSchema("User")

	// Add fields with various properties
	idField := userSchema.AddField("id", JString, "").SetRequired(true).SetUnique(true)
	userSchema.AddField("name", JString, "").SetRequired(true)
	userSchema.AddField("email", JString, "").SetRequired(true).SetUnique(true)
	userSchema.AddField("age", JInt, 0)
	userSchema.AddField("active", JBool, true)
	userSchema.AddField("metadata", JObject, map[string]any{})
	userSchema.AddField("tags", JArray, []string{})

	// Set ID field
	userSchema.SetIDField(idField)

	// Add validations
	userSchema.AddValidation(func(ctx context.Context, rec JRecord) error {
		age := rec.Get("age")
		if age != nil {
			if ageInt, ok := age.(int); ok && ageInt < 0 {
				return assert.AnError
			}
		}
		return nil
	})

	// Create related schemas
	profileSchema := NewJSchema("Profile")
	profileSchema.AddField("bio", JString, "")

	orderSchema := NewJSchema("Order")
	orderSchema.AddField("total", JFloat64, 0.0)

	// Add references
	userSchema.AddRef("profile", profileSchema)
	userSchema.AddRef("orders", orderSchema)

	// Add edges
	userOrderEdge := NewJEdge("user_orders", userSchema, orderSchema, EdgeOneToMany)
	userProfileEdge := NewJEdge("user_profile", userSchema, profileSchema, EdgeOneToOne)
	userSchema.AddEdge(userOrderEdge).AddEdge(userProfileEdge)

	// Test the complete schema
	assert.Equal(t, "User", userSchema.Name())
	assert.Len(t, userSchema.Fields(), 7)
	assert.Len(t, userSchema.Refs(), 2)
	assert.Len(t, userSchema.Edges(), 2)
	assert.Len(t, userSchema.Validations(), 1)
	assert.Equal(t, idField, userSchema.GetIDField())
	assert.False(t, userSchema.IsImmutable())

	// Test validation with valid record
	validRecord := NewJRecordFromMap(map[string]any{
		"id":       "user123",
		"name":     "John Doe",
		"email":    "john@example.com",
		"age":      25,
		"active":   true,
		"metadata": map[string]any{"role": "user"},
		"tags":     []string{"premium", "verified"},
	})

	ctx := context.Background()
	err := userSchema.Validate(ctx, validRecord)
	assert.NoError(t, err)

	// Test validation with invalid record
	invalidRecord := NewJRecordFromMap(map[string]any{
		"id":    "user123",
		"name":  "John Doe",
		"email": "john@example.com",
		"age":   -5, // Invalid age
	})

	err = userSchema.Validate(ctx, invalidRecord)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "schema validation failed")

	// Test cloning
	clonedSchema := userSchema.Clone()
	assert.Equal(t, userSchema.Name(), clonedSchema.Name())
	assert.Len(t, clonedSchema.Fields(), 7)
	assert.Len(t, clonedSchema.Refs(), 2)
	assert.Len(t, clonedSchema.Edges(), 2)
	assert.Len(t, clonedSchema.Validations(), 1)
	assert.False(t, clonedSchema.IsImmutable())

	// Test freezing
	frozenSchema := userSchema.Freeze()
	assert.True(t, frozenSchema.IsImmutable())

	// Test string representation
	str := userSchema.String()
	assert.Contains(t, str, "Schema: User")
	assert.Contains(t, str, "id: string (required) (unique)")
	assert.Contains(t, str, "name: string (required)")
	assert.Contains(t, str, "email: string (required) (unique)")
	assert.Contains(t, str, "age: int")
	assert.Contains(t, str, "active: bool")
	assert.Contains(t, str, "metadata: object")
	assert.Contains(t, str, "tags: array")
	assert.Contains(t, str, "profile -> Profile")
	assert.Contains(t, str, "orders -> Order")
}
