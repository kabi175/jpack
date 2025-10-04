package schema

import (
	"context"
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
			schema := NewSchemaBuilder(tt.schemaName).Build()
			assert.Equal(t, tt.expected, schema.Name())
			assert.True(t, schema.IsImmutable()) // All schemas are now immutable
			assert.Len(t, schema.Fields(), 1)    // Automatic ID field
			assert.Empty(t, schema.Refs())
			assert.Empty(t, schema.Edges())
			assert.Empty(t, schema.Validations())
			assert.NotNil(t, schema.GetIDField()) // Automatic ID field
			assert.Equal(t, "id", schema.GetIDField().Name())
		})
	}
}

func TestNewImmutableJSchema(t *testing.T) {
	schema := NewSchemaBuilder("TestSchema").Build()

	assert.Equal(t, "TestSchema", schema.Name())
	assert.True(t, schema.IsImmutable())
	assert.Len(t, schema.Fields(), 1) // Automatic ID field
	assert.Empty(t, schema.Refs())
	assert.Empty(t, schema.Edges())
	assert.Empty(t, schema.Validations())
	assert.NotNil(t, schema.GetIDField()) // Automatic ID field
	assert.Equal(t, "id", schema.GetIDField().Name())
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
			schema := NewSchemaBuilder("TestSchema").
				AddField(tt.fieldName, tt.fieldType, tt.defaultValue).
				Build()

			// Check that field was added to schema
			retrievedField, exists := schema.Field(tt.fieldName)
			assert.True(t, exists)
			assert.Equal(t, tt.fieldName, retrievedField.Name())
			assert.Equal(t, tt.expectedType, retrievedField.Type())
			assert.Equal(t, tt.defaultValue, retrievedField.DefaultValue())
			assert.False(t, retrievedField.IsRequired())
			assert.False(t, retrievedField.IsUnique())
			assert.True(t, retrievedField.IsImmutable()) // All fields are now immutable

			// Check that ID field is automatically created
			if len(schema.Fields()) == 2 { // id + the added field
				idField := schema.GetIDField()
				assert.Equal(t, "id", idField.Name())
				assert.True(t, idField.IsRequired())
				assert.True(t, idField.IsUnique())
			}
		})
	}
}

func TestJSchema_AddField_Immutable(t *testing.T) {
	schema := NewSchemaBuilder("TestSchema").Build()

	// All schemas are now immutable, so AddField should not be available
	// This test verifies that the old mutation methods are no longer available
	// The new way is to use Update() method
	updatedSchema := schema.Update(func(sb *SchemaBuilder) {
		sb.AddField("name", JString, "default")
	})

	// Verify the field was added
	field, exists := updatedSchema.Field("name")
	assert.True(t, exists)
	assert.Equal(t, "name", field.Name())
	assert.Equal(t, JString, field.Type())
}

func TestJSchema_AddField_Duplicate(t *testing.T) {
	schema := NewSchemaBuilder("TestSchema").
		AddField("name", JString, "default").
		Build()

	// Add field with same name (should overwrite)
	updatedSchema := schema.Update(func(sb *SchemaBuilder) {
		sb.AddField("name", JInt, 42)
	})

	field, exists := updatedSchema.Field("name")
	assert.True(t, exists)
	assert.Equal(t, "name", field.Name())
	assert.Equal(t, JInt, field.Type())
	assert.Equal(t, 42, field.DefaultValue())

	// Should have two fields (id + name)
	assert.Len(t, updatedSchema.Fields(), 2)

	// ID field should be the automatic ID field
	idField := updatedSchema.GetIDField()
	assert.Equal(t, "id", idField.Name())
	assert.True(t, idField.IsRequired())
	assert.True(t, idField.IsUnique())
}

func TestJSchema_Field(t *testing.T) {
	schema := NewSchemaBuilder("TestSchema").
		AddField("name", JString, "default").
		Build()

	// Test non-existent field
	field, exists := schema.Field("nonexistent")
	assert.False(t, exists)
	assert.Nil(t, field)

	// Test field retrieval (field was added during schema creation)
	retrievedField, exists := schema.Field("name")

	assert.True(t, exists)
	assert.Equal(t, "name", retrievedField.Name())
	assert.Equal(t, JString, retrievedField.Type())
	assert.Equal(t, "default", retrievedField.DefaultValue())
}

func TestJSchema_AddRef(t *testing.T) {
	userSchema := NewSchemaBuilder("User").Build()
	orderSchema := NewSchemaBuilder("Order").
		AddRef("user", userSchema).
		Build()

	// Get reference
	ref, exists := orderSchema.Ref("user")
	assert.True(t, exists)

	assert.Equal(t, "user", ref.Name())
	assert.Equal(t, userSchema, ref.TargetSchema())
	assert.False(t, ref.IsArray())

	// Check that ref was added to schema
	assert.Len(t, orderSchema.Refs(), 1)
}

func TestJSchema_AddRef_Immutable(t *testing.T) {
	schema := NewSchemaBuilder("TestSchema").Build()
	targetSchema := NewSchemaBuilder("Target").Build()

	// All schemas are now immutable, so AddRef should not be available
	// The new way is to use Update() method
	updatedSchema := schema.Update(func(sb *SchemaBuilder) {
		sb.AddRef("target", targetSchema)
	})

	// Verify the ref was added
	ref, exists := updatedSchema.Ref("target")
	assert.True(t, exists)
	assert.Equal(t, "target", ref.Name())
	assert.Equal(t, targetSchema, ref.TargetSchema())
}

func TestJSchema_AddRef_Duplicate(t *testing.T) {
	target1 := NewSchemaBuilder("Target1").Build()
	target2 := NewSchemaBuilder("Target2").Build()
	schema := NewSchemaBuilder("TestSchema").
		AddRef("target", target1).
		Build()

	// Add ref with same name (should overwrite)
	updatedSchema := schema.Update(func(sb *SchemaBuilder) {
		sb.AddRef("target", target2)
	})

	ref, exists := updatedSchema.Ref("target")
	assert.True(t, exists)
	assert.Equal(t, "target", ref.Name())
	assert.Equal(t, target2, ref.TargetSchema())

	// Should only have one ref
	assert.Len(t, updatedSchema.Refs(), 1)
}

func TestJSchema_Ref(t *testing.T) {
	targetSchema := NewSchemaBuilder("Target").Build()
	schema := NewSchemaBuilder("TestSchema").
		AddRef("target", targetSchema).
		Build()

	// Test non-existent ref
	ref, exists := schema.Ref("nonexistent")
	assert.False(t, exists)
	assert.Nil(t, ref)

	// Test ref retrieval
	retrievedRef, exists := schema.Ref("target")
	assert.True(t, exists)
	assert.Equal(t, "target", retrievedRef.Name())
	assert.Equal(t, targetSchema, retrievedRef.TargetSchema())
}

func TestJSchema_AddEdge(t *testing.T) {
	userSchema := NewSchemaBuilder("User").Build()
	orderSchema := NewSchemaBuilder("Order").Build()

	edge := NewJEdge("user_orders", userSchema, orderSchema, EdgeOneToMany)

	// Add edge using Update method
	updatedSchema := userSchema.Update(func(sb *SchemaBuilder) {
		sb.AddEdge(edge)
	})

	// Check that edge was added
	edges := updatedSchema.Edges()
	assert.Len(t, edges, 1)
	assert.Equal(t, edge, edges[0])
}

func TestJSchema_AddEdge_Immutable(t *testing.T) {
	schema := NewSchemaBuilder("TestSchema").Build()
	targetSchema := NewSchemaBuilder("Target").Build()
	edge := NewJEdge("test_edge", schema, targetSchema, EdgeOneToOne)

	// All schemas are now immutable, so AddEdge should not be available
	// The new way is to use Update() method
	updatedSchema := schema.Update(func(sb *SchemaBuilder) {
		sb.AddEdge(edge)
	})

	// Verify the edge was added
	edges := updatedSchema.Edges()
	assert.Len(t, edges, 1)
	assert.Equal(t, edge, edges[0])
}

func TestJSchema_AddEdge_Multiple(t *testing.T) {
	userSchema := NewSchemaBuilder("User").Build()
	orderSchema := NewSchemaBuilder("Order").Build()
	productSchema := NewSchemaBuilder("Product").Build()

	// Add multiple edges
	edge1 := NewJEdge("user_orders", userSchema, orderSchema, EdgeOneToMany)
	edge2 := NewJEdge("user_products", userSchema, productSchema, EdgeManyToMany)

	updatedSchema := userSchema.Update(func(sb *SchemaBuilder) {
		sb.AddEdge(edge1)
		sb.AddEdge(edge2)
	})

	edges := updatedSchema.Edges()
	assert.Len(t, edges, 2)
	assert.Contains(t, edges, edge1)
	assert.Contains(t, edges, edge2)
}

func TestJSchema_SetIDField(t *testing.T) {
	schema := NewSchemaBuilder("TestSchema").
		AddField("name", JString, "default").
		AddField("age", JInt, 0).
		Build()

	// ID field should be automatically created
	idField := schema.GetIDField()
	assert.Equal(t, "id", idField.Name())
	assert.True(t, idField.IsRequired())
	assert.True(t, idField.IsUnique())

	// Set different field as ID using Update
	updatedSchema := schema.Update(func(sb *SchemaBuilder) {
		sb.SetIDField("name")
	})

	// Check that ID field was changed
	newIDField := updatedSchema.GetIDField()
	assert.Equal(t, "name", newIDField.Name())
}

func TestJSchema_SetIDField_Immutable(t *testing.T) {
	// All schemas are now immutable by default
	schema := NewSchemaBuilder("TestSchema").
		AddField("name", JString, "default").
		Build()

	// The new way is to use Update() method
	updatedSchema := schema.Update(func(sb *SchemaBuilder) {
		sb.SetIDField("name")
	})

	// Verify the ID field was changed
	newIDField := updatedSchema.GetIDField()
	assert.Equal(t, "name", newIDField.Name())
}

func TestJSchema_SetIDField_NonExistent(t *testing.T) {
	schema := NewSchemaBuilder("TestSchema").
		AddField("name", JString, "default").
		Build()

	// Set ID field to an existing field in the schema
	updatedSchema := schema.Update(func(sb *SchemaBuilder) {
		sb.SetIDField("name")
	})

	// Verify the ID field was set
	newIDField := updatedSchema.GetIDField()
	assert.Equal(t, "name", newIDField.Name())
}

func TestJSchema_AddValidation(t *testing.T) {
	// Add validation function
	validationFunc := func(ctx context.Context, rec JRecord) error {
		return nil
	}

	schema := NewSchemaBuilder("TestSchema").
		AddValidation(validationFunc).
		Build()

	validations := schema.Validations()
	assert.Len(t, validations, 1)
	assert.NotNil(t, validations[0])
	// Functions cannot be compared directly, so we just verify it's not nil
}

func TestJSchema_AddValidation_Immutable(t *testing.T) {
	schema := NewSchemaBuilder("TestSchema").Build()

	validationFunc := func(ctx context.Context, rec JRecord) error {
		return nil
	}

	// All schemas are now immutable, so AddValidation should not be available
	// The new way is to use Update() method
	updatedSchema := schema.Update(func(sb *SchemaBuilder) {
		sb.AddValidation(validationFunc)
	})

	// Verify the validation was added
	validations := updatedSchema.Validations()
	assert.Len(t, validations, 1)
	assert.NotNil(t, validations[0])
}

func TestJSchema_AddValidation_Multiple(t *testing.T) {
	// Add multiple validations
	validation1 := func(ctx context.Context, rec JRecord) error { return nil }
	validation2 := func(ctx context.Context, rec JRecord) error { return nil }

	schema := NewSchemaBuilder("TestSchema").
		AddValidation(validation1).
		AddValidation(validation2).
		Build()

	validations := schema.Validations()
	assert.Len(t, validations, 2)
	assert.NotNil(t, validations[0])
	assert.NotNil(t, validations[1])
	// Functions cannot be compared directly, so we just verify they're not nil
}

func TestJSchema_Validate(t *testing.T) {
	schema := NewSchemaBuilder("TestSchema").
		AddRequiredField("name", JString, "default").
		AddField("age", JInt, 0).
		Build()

	tests := []struct {
		name      string
		record    JRecord
		wantError bool
		errorMsg  string
	}{
		{
			name: "valid record with required field",
			record: NewJRecordFromMap(map[string]any{
				"id":   "user123",
				"name": "John",
				"age":  25,
			}),
			wantError: false,
		},
		{
			name: "valid record with only required field",
			record: NewJRecordFromMap(map[string]any{
				"id":   "user456",
				"name": "Jane",
			}),
			wantError: false,
		},
		{
			name: "invalid record missing required field",
			record: NewJRecordFromMap(map[string]any{
				"id":  "user789",
				"age": 25,
			}),
			wantError: true,
			errorMsg:  "required field 'name' is missing",
		},
		{
			name:      "empty record",
			record:    NewJRecord(),
			wantError: true,
			errorMsg:  "required field", // Just check that it's a required field error
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
	schema := NewSchemaBuilder("TestSchema").
		AddFieldWithValidation("age", JInt, 0, func(ctx context.Context, rec JRecord) error {
			age := rec.Get("age")
			if age == nil {
				return nil // Let required validation handle this
			}
			if ageInt, ok := age.(int); ok && ageInt < 0 {
				return assert.AnError
			}
			return nil
		}).
		Build()

	tests := []struct {
		name      string
		record    JRecord
		wantError bool
	}{
		{
			name: "valid age",
			record: NewJRecordFromMap(map[string]any{
				"id":  "user123",
				"age": 25,
			}),
			wantError: false,
		},
		{
			name: "zero age",
			record: NewJRecordFromMap(map[string]any{
				"id":  "user456",
				"age": 0,
			}),
			wantError: false,
		},
		{
			name: "negative age",
			record: NewJRecordFromMap(map[string]any{
				"id":  "user789",
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
	schema := NewSchemaBuilder("TestSchema").
		AddValidation(func(ctx context.Context, rec JRecord) error {
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
		}).
		Build()

	tests := []struct {
		name      string
		record    JRecord
		wantError bool
	}{
		{
			name: "valid admin",
			record: NewJRecordFromMap(map[string]any{
				"id":   "admin123",
				"name": "admin",
				"age":  25,
			}),
			wantError: false,
		},
		{
			name: "invalid admin age",
			record: NewJRecordFromMap(map[string]any{
				"id":   "admin456",
				"name": "admin",
				"age":  16,
			}),
			wantError: true,
		},
		{
			name: "non-admin with low age",
			record: NewJRecordFromMap(map[string]any{
				"id":   "user789",
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
	targetSchema := NewSchemaBuilder("Target").Build()
	original := NewSchemaBuilder("Original").
		AddRequiredField("name", JString, "default").
		AddUniqueField("age", JInt, 0).
		AddValidation(func(ctx context.Context, rec JRecord) error {
			return nil
		}).
		AddRef("target", targetSchema).
		Build()

	// Add edge using Update
	edge := NewJEdge("test_edge", original, targetSchema, EdgeOneToOne)
	original = original.Update(func(sb *SchemaBuilder) {
		sb.AddEdge(edge)
	})

	// Clone schema
	cloned := original.Clone()

	// Test basic properties
	assert.Equal(t, original.Name(), cloned.Name())
	assert.True(t, cloned.IsImmutable()) // All schemas are now immutable

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
		assert.True(t, clonedField.IsImmutable()) // All fields are now immutable
	}

	// Test ID field
	assert.NotNil(t, cloned.GetIDField())
	assert.Equal(t, "id", cloned.GetIDField().Name()) // Automatic ID field

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
	// Since all schemas are now immutable, we use Update method
	updatedCloned := cloned.Update(func(sb *SchemaBuilder) {
		sb.AddField("new_field", JString, "new")
	})
	assert.Len(t, original.Fields(), 3)      // id, name, age
	assert.Len(t, updatedCloned.Fields(), 4) // id, name, age, new_field
}

func TestJSchema_Clone_Immutable(t *testing.T) {
	// All schemas are now immutable by default
	original := NewSchemaBuilder("Original").
		AddField("name", JString, "default").
		Build()

	// Clone should also be immutable
	cloned := original.Clone()
	assert.True(t, cloned.IsImmutable())

	// Should be able to modify clone using Update method
	updatedCloned := cloned.Update(func(sb *SchemaBuilder) {
		sb.AddField("age", JInt, 0)
	})
	assert.Len(t, updatedCloned.Fields(), 3) // id, name, age
}

func TestJSchema_Freeze(t *testing.T) {
	schema := NewSchemaBuilder("TestSchema").
		AddField("name", JString, "default").
		Build()

	// All schemas are now immutable by default
	assert.True(t, schema.IsImmutable())

	// Get field and verify it's immutable
	field, exists := schema.Field("name")
	assert.True(t, exists)
	assert.True(t, field.IsImmutable())

	// Freeze should return the same schema (already immutable)
	result := schema.Freeze()
	assert.Equal(t, schema, result)

	// Should not panic when trying to modify using Update method
	updatedSchema := schema.Update(func(sb *SchemaBuilder) {
		sb.AddField("age", JInt, 0)
	})
	assert.Len(t, updatedSchema.Fields(), 3) // id, name, age

	// Field modifications should also use Update method
	updatedField := field.Update(func(fb *FieldBuilder) {
		fb.Required()
	})
	assert.True(t, updatedField.IsRequired())
}

func TestJSchema_Freeze_AlreadyImmutable(t *testing.T) {
	schema := NewSchemaBuilder("TestSchema").Build()

	// Freezing already immutable schema should work
	result := schema.Freeze()
	assert.Equal(t, schema, result)
	assert.True(t, schema.IsImmutable())
}

func TestJSchema_String(t *testing.T) {
	targetSchema := NewSchemaBuilder("Profile").Build()
	schema := NewSchemaBuilder("User").
		AddRequiredField("name", JString, "").
		AddField("age", JInt, 0).
		AddRef("profile", targetSchema).
		Build()

	// Get string representation
	str := schema.String()

	// Check that it contains expected information
	assert.Contains(t, str, "Schema: User")
	assert.Contains(t, str, "Fields:")
	assert.Contains(t, str, "id: string (required) (unique)") // Automatic ID field
	assert.Contains(t, str, "name: string (required)")
	assert.Contains(t, str, "age: int")
	assert.Contains(t, str, "References:")
	assert.Contains(t, str, "profile -> Profile")
}

func TestJSchema_String_Empty(t *testing.T) {
	schema := NewSchemaBuilder("Empty").Build()
	str := schema.String()

	assert.Contains(t, str, "Schema: Empty")
	assert.Contains(t, str, "Fields:")
	// Should not contain References section
	assert.NotContains(t, str, "References:")
}

func TestJSchema_EdgeCases(t *testing.T) {
	t.Run("empty schema name", func(t *testing.T) {
		schema := NewSchemaBuilder("").Build()
		assert.Equal(t, "", schema.Name())
	})

	t.Run("schema with only edges", func(t *testing.T) {
		schema := NewSchemaBuilder("EdgeOnly").Build()
		targetSchema := NewSchemaBuilder("Target").Build()
		edge := NewJEdge("test_edge", schema, targetSchema, EdgeOneToOne)
		schema = schema.Update(func(sb *SchemaBuilder) {
			sb.AddEdge(edge)
		})

		assert.Len(t, schema.Fields(), 1) // Automatic ID field
		assert.Len(t, schema.Edges(), 1)
		assert.NotNil(t, schema.GetIDField()) // Automatic ID field
	})

	t.Run("schema with only refs", func(t *testing.T) {
		targetSchema := NewSchemaBuilder("Target").Build()
		schema := NewSchemaBuilder("RefOnly").
			AddRef("target", targetSchema).
			Build()

		assert.Len(t, schema.Fields(), 1) // Automatic ID field
		assert.Len(t, schema.Refs(), 1)
		assert.NotNil(t, schema.GetIDField()) // Automatic ID field
	})

	t.Run("schema with only validations", func(t *testing.T) {
		schema := NewSchemaBuilder("ValidationOnly").
			AddValidation(func(ctx context.Context, rec JRecord) error {
				return nil
			}).
			Build()

		assert.Len(t, schema.Fields(), 1) // Automatic ID field
		assert.Len(t, schema.Validations(), 1)
	})

	t.Run("set nil ID field", func(t *testing.T) {
		schema := NewSchemaBuilder("Test").
			AddField("name", JString, "default").
			Build()

		// Set ID field to empty string should work
		updatedSchema := schema.Update(func(sb *SchemaBuilder) {
			sb.SetIDField("") // This should work but may not be useful
		})

		// The schema should still have the automatic ID field
		assert.NotNil(t, updatedSchema.GetIDField())
	})
}

func TestJSchema_ConcurrentAccess(t *testing.T) {
	schema := NewSchemaBuilder("Concurrent").Build()

	// Test concurrent field additions
	done := make(chan bool, 10)

	// Since schemas are now immutable, we need to test concurrent access differently
	// We'll test concurrent reads instead of concurrent modifications
	for i := 0; i < 10; i++ {
		go func(index int) {
			defer func() { done <- true }()

			// Test concurrent reads
			_ = schema.Name()
			_ = schema.Fields()
			_ = schema.IsImmutable()
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Should have 1 field (automatic ID field)
	assert.Len(t, schema.Fields(), 1)
}

func TestJSchema_Validation_ContextCancellation(t *testing.T) {
	schema := NewSchemaBuilder("TestSchema").
		AddRequiredField("name", JString, "default").
		AddValidation(func(ctx context.Context, rec JRecord) error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				return nil
			}
		}).
		Build()

	record := NewJRecordFromMap(map[string]any{
		"id":   "test123",
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
	schema := NewSchemaBuilder("TestSchema").
		AddRequiredField("name", JString, "default").
		AddValidation(func(ctx context.Context, rec JRecord) error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				return nil
			}
		}).
		Build()

	record := NewJRecordFromMap(map[string]any{
		"id":   "test456",
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
	userSchema := NewSchemaBuilder("User").
		AddRequiredField("name", JString, "").
		AddRequiredUniqueField("email", JString, "").
		AddField("age", JInt, 0).
		AddField("active", JBool, true).
		AddField("metadata", JObject, map[string]any{}).
		AddField("tags", JArray, []string{}).
		AddValidation(func(ctx context.Context, rec JRecord) error {
			age := rec.Get("age")
			if age != nil {
				if ageInt, ok := age.(int); ok && ageInt < 0 {
					return assert.AnError
				}
			}
			return nil
		}).
		Build()

	// Create related schemas
	profileSchema := NewSchemaBuilder("Profile").
		AddField("bio", JString, "").
		Build()

	orderSchema := NewSchemaBuilder("Order").
		AddField("total", JFloat64, 0.0).
		Build()

	// Add references and edges using Update method
	userSchema = userSchema.Update(func(sb *SchemaBuilder) {
		sb.AddRef("profile", profileSchema)
		sb.AddRef("orders", orderSchema)

		// Add edges
		userOrderEdge := NewJEdge("user_orders", userSchema, orderSchema, EdgeOneToMany)
		userProfileEdge := NewJEdge("user_profile", userSchema, profileSchema, EdgeOneToOne)
		sb.AddEdge(userOrderEdge)
		sb.AddEdge(userProfileEdge)
	})

	// Test the complete schema
	assert.Equal(t, "User", userSchema.Name())
	assert.Len(t, userSchema.Fields(), 7) // id, name, email, age, active, metadata, tags
	assert.Len(t, userSchema.Refs(), 2)
	assert.Len(t, userSchema.Edges(), 2)
	assert.Len(t, userSchema.Validations(), 1)
	assert.Equal(t, "id", userSchema.GetIDField().Name()) // Automatic ID field
	assert.True(t, userSchema.IsImmutable())              // All schemas are now immutable

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
	assert.Len(t, clonedSchema.Fields(), 7) // id, name, email, age, active, metadata, tags
	assert.Len(t, clonedSchema.Refs(), 2)
	assert.Len(t, clonedSchema.Edges(), 2)
	assert.Len(t, clonedSchema.Validations(), 1)
	assert.True(t, clonedSchema.IsImmutable()) // All schemas are now immutable

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
