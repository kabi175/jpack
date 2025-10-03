package schema

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewJRef(t *testing.T) {
	tests := []struct {
		name          string
		refName       string
		targetSchema  JSchema
		isArray       bool
		expectedName  string
		expectedArray bool
	}{
		{
			name:          "single reference",
			refName:       "user",
			targetSchema:  NewJSchema("User"),
			isArray:       false,
			expectedName:  "user",
			expectedArray: false,
		},
		{
			name:          "array reference",
			refName:       "orders",
			targetSchema:  NewJSchema("Order"),
			isArray:       true,
			expectedName:  "orders",
			expectedArray: true,
		},
		{
			name:          "empty reference name",
			refName:       "",
			targetSchema:  NewJSchema("Target"),
			isArray:       false,
			expectedName:  "",
			expectedArray: false,
		},
		{
			name:          "reference with special characters",
			refName:       "user-profile_v2",
			targetSchema:  NewJSchema("UserProfile"),
			isArray:       false,
			expectedName:  "user-profile_v2",
			expectedArray: false,
		},
		{
			name:          "reference to immutable schema",
			refName:       "config",
			targetSchema:  NewImmutableJSchema("Config"),
			isArray:       false,
			expectedName:  "config",
			expectedArray: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ref := NewJRef(tt.refName, tt.targetSchema, tt.isArray)

			assert.Equal(t, tt.expectedName, ref.Name())
			assert.Equal(t, tt.targetSchema, ref.TargetSchema())
			assert.Equal(t, tt.expectedArray, ref.IsArray())
		})
	}
}

func TestJRef_EdgeCases(t *testing.T) {
	t.Run("nil target schema", func(t *testing.T) {
		ref := NewJRef("test", nil, false)

		assert.Equal(t, "test", ref.Name())
		assert.Nil(t, ref.TargetSchema())
		assert.False(t, ref.IsArray())
	})

	t.Run("reference to self", func(t *testing.T) {
		schema := NewJSchema("SelfRef")
		ref := NewJRef("self", schema, false)

		assert.Equal(t, "self", ref.Name())
		assert.Equal(t, schema, ref.TargetSchema())
		assert.False(t, ref.IsArray())
	})

	t.Run("circular reference", func(t *testing.T) {
		userSchema := NewJSchema("User")
		orderSchema := NewJSchema("Order")

		// User references Order
		userRef := NewJRef("orders", orderSchema, true)

		// Order references User
		orderRef := NewJRef("user", userSchema, false)

		assert.Equal(t, "orders", userRef.Name())
		assert.Equal(t, orderSchema, userRef.TargetSchema())
		assert.True(t, userRef.IsArray())

		assert.Equal(t, "user", orderRef.Name())
		assert.Equal(t, userSchema, orderRef.TargetSchema())
		assert.False(t, orderRef.IsArray())
	})

	t.Run("reference to schema with fields", func(t *testing.T) {
		targetSchema := NewSchemaBuilder("Target").
			AddRequiredField("name", JString, "").
			Build()

		ref := NewJRef("target", targetSchema, false)

		assert.Equal(t, "target", ref.Name())
		assert.Equal(t, targetSchema, ref.TargetSchema())
		assert.False(t, ref.IsArray())

		// Verify target schema has fields (including automatic ID field)
		assert.Len(t, targetSchema.Fields(), 2) // id, name
	})

	t.Run("reference to schema with refs", func(t *testing.T) {
		profileSchema := NewSchemaBuilder("Profile").Build()
		orderSchema := NewSchemaBuilder("Order").Build()
		userSchema := NewSchemaBuilder("User").
			AddRef("profile", profileSchema).
			AddRef("orders", orderSchema).
			Build()

		// Reference to user schema
		ref := NewJRef("user", userSchema, false)

		assert.Equal(t, "user", ref.Name())
		assert.Equal(t, userSchema, ref.TargetSchema())
		assert.False(t, ref.IsArray())

		// Verify user schema has refs
		assert.Len(t, userSchema.Refs(), 2)
	})

	t.Run("reference to schema with edges", func(t *testing.T) {
		userSchema := NewSchemaBuilder("User").Build()
		orderSchema := NewSchemaBuilder("Order").Build()

		// Add edge
		edge := NewJEdge("user_orders", userSchema, orderSchema, EdgeOneToMany)
		userSchema = userSchema.Update(func(sb *SchemaBuilder) {
			sb.AddEdge(edge)
		})

		// Reference to user schema
		ref := NewJRef("user", userSchema, false)

		assert.Equal(t, "user", ref.Name())
		assert.Equal(t, userSchema, ref.TargetSchema())
		assert.False(t, ref.IsArray())

		// Verify user schema has edges
		assert.Len(t, userSchema.Edges(), 1)
	})
}

func TestJRef_ArrayVsSingle(t *testing.T) {
	targetSchema := NewJSchema("Target")

	t.Run("single reference", func(t *testing.T) {
		ref := NewJRef("single", targetSchema, false)

		assert.Equal(t, "single", ref.Name())
		assert.Equal(t, targetSchema, ref.TargetSchema())
		assert.False(t, ref.IsArray())
	})

	t.Run("array reference", func(t *testing.T) {
		ref := NewJRef("array", targetSchema, true)

		assert.Equal(t, "array", ref.Name())
		assert.Equal(t, targetSchema, ref.TargetSchema())
		assert.True(t, ref.IsArray())
	})
}

func TestJRef_ComplexTargetSchema(t *testing.T) {
	// Create a complex target schema
	targetSchema := NewSchemaBuilder("ComplexTarget").
		AddRequiredField("name", JString, "").
		AddField("metadata", JObject, map[string]any{}).
		AddField("tags", JArray, []string{}).
		AddValidation(func(ctx context.Context, rec JRecord) error {
			return nil
		}).
		Build()

	// Create reference
	ref := NewJRef("complex", targetSchema, true)

	assert.Equal(t, "complex", ref.Name())
	assert.Equal(t, targetSchema, ref.TargetSchema())
	assert.True(t, ref.IsArray())

	// Verify target schema properties
	assert.Equal(t, "ComplexTarget", targetSchema.Name())
	assert.Len(t, targetSchema.Fields(), 4) // id, name, metadata, tags
	assert.Len(t, targetSchema.Validations(), 1)
	assert.True(t, targetSchema.IsImmutable()) // All schemas are now immutable
}

func TestJRef_ImmutableTargetSchema(t *testing.T) {
	// Create immutable schema directly
	targetSchema := NewSchemaBuilder("ImmutableTarget").
		AddRequiredField("name", JString, "").
		Build()

	// Create reference
	ref := NewJRef("immutable", targetSchema, false)

	assert.Equal(t, "immutable", ref.Name())
	assert.Equal(t, targetSchema, ref.TargetSchema())
	assert.False(t, ref.IsArray())

	// Verify target schema is immutable
	assert.True(t, targetSchema.IsImmutable())
}

func TestJRef_StringRepresentation(t *testing.T) {
	targetSchema := NewSchemaBuilder("Target").Build()

	t.Run("single reference string", func(t *testing.T) {
		ref := NewJRef("single", targetSchema, false)

		// Test that ref can be used in string operations
		str := fmt.Sprintf("Ref: %s -> %s", ref.Name(), ref.TargetSchema().Name())
		assert.Equal(t, "Ref: single -> Target", str)
	})

	t.Run("array reference string", func(t *testing.T) {
		ref := NewJRef("array", targetSchema, true)

		// Test that ref can be used in string operations
		str := fmt.Sprintf("Ref: %s -> %s (array)", ref.Name(), ref.TargetSchema().Name())
		assert.Equal(t, "Ref: array -> Target (array)", str)
	})
}

func TestJRef_Comparison(t *testing.T) {
	targetSchema1 := NewJSchema("Target1")
	targetSchema2 := NewJSchema("Target2")

	ref1 := NewJRef("ref", targetSchema1, false)
	ref2 := NewJRef("ref", targetSchema1, false)
	ref3 := NewJRef("ref", targetSchema2, false)
	ref4 := NewJRef("ref", targetSchema1, true)

	// Refs with same properties should have equal content
	assert.Equal(t, ref1.Name(), ref2.Name())
	assert.Equal(t, ref1.TargetSchema(), ref2.TargetSchema())
	assert.Equal(t, ref1.IsArray(), ref2.IsArray())

	// Different target schemas
	assert.NotEqual(t, ref1.TargetSchema(), ref3.TargetSchema())

	// Different array properties
	assert.NotEqual(t, ref1.IsArray(), ref4.IsArray())
}

func TestJRef_ConcurrentAccess(t *testing.T) {
	targetSchema := NewJSchema("ConcurrentTarget")
	ref := NewJRef("concurrent", targetSchema, false)

	// Test concurrent access to ref properties
	done := make(chan bool, 3)

	go func() {
		defer func() { done <- true }()
		for i := 0; i < 100; i++ {
			_ = ref.Name()
		}
	}()

	go func() {
		defer func() { done <- true }()
		for i := 0; i < 100; i++ {
			_ = ref.TargetSchema()
		}
	}()

	go func() {
		defer func() { done <- true }()
		for i := 0; i < 100; i++ {
			_ = ref.IsArray()
		}
	}()

	// Wait for all goroutines to complete
	for i := 0; i < 3; i++ {
		<-done
	}

	// Ref should still be in a valid state
	assert.Equal(t, "concurrent", ref.Name())
	assert.Equal(t, targetSchema, ref.TargetSchema())
	assert.False(t, ref.IsArray())
}

func TestJRef_DeepNesting(t *testing.T) {
	// Create deeply nested schema structure
	level1 := NewJSchema("Level1")
	level2 := NewJSchema("Level2")
	level3 := NewJSchema("Level3")

	// Level1 references Level2
	ref1 := NewJRef("level2", level2, false)

	// Level2 references Level3
	ref2 := NewJRef("level3", level3, true)

	// Level3 references Level1 (circular)
	ref3 := NewJRef("level1", level1, false)

	assert.Equal(t, "level2", ref1.Name())
	assert.Equal(t, level2, ref1.TargetSchema())
	assert.False(t, ref1.IsArray())

	assert.Equal(t, "level3", ref2.Name())
	assert.Equal(t, level3, ref2.TargetSchema())
	assert.True(t, ref2.IsArray())

	assert.Equal(t, "level1", ref3.Name())
	assert.Equal(t, level1, ref3.TargetSchema())
	assert.False(t, ref3.IsArray())
}

func TestJRef_AllEdgeTypes(t *testing.T) {
	// Create all possible edge types
	edgeTypes := []EdgeType{EdgeOneToOne, EdgeOneToMany, EdgeManyToOne, EdgeManyToMany}

	for _, edgeType := range edgeTypes {
		t.Run(string(edgeType), func(t *testing.T) {
			// Create fresh schemas for each test
			userSchema := NewSchemaBuilder("User").Build()
			orderSchema := NewSchemaBuilder("Order").Build()

			edge := NewJEdge("test_edge", userSchema, orderSchema, edgeType)
			userSchema = userSchema.Update(func(sb *SchemaBuilder) {
				sb.AddEdge(edge)
			})

			// Create reference to schema with this edge
			ref := NewJRef("user", userSchema, false)

			assert.Equal(t, "user", ref.Name())
			assert.Equal(t, userSchema, ref.TargetSchema())
			assert.False(t, ref.IsArray())

			// Verify edge was added
			edges := userSchema.Edges()
			assert.Len(t, edges, 1)
			assert.Equal(t, edgeType, edges[0].Type())
		})
	}
}

func TestJRef_Validation(t *testing.T) {
	targetSchema := NewSchemaBuilder("ValidatedTarget").
		AddFieldWithValidation("email", JString, "", func(ctx context.Context, rec JRecord) error {
			email := rec.Get("email")
			if email != nil {
				if emailStr, ok := email.(string); ok && !strings.Contains(emailStr, "@") {
					return assert.AnError
				}
			}
			return nil
		}).
		Build()

	// Create reference
	ref := NewJRef("validated", targetSchema, false)

	assert.Equal(t, "validated", ref.Name())
	assert.Equal(t, targetSchema, ref.TargetSchema())
	assert.False(t, ref.IsArray())

	// Verify target schema has validation
	fields := targetSchema.Fields()
	assert.Len(t, fields, 2) // id field + email field
	// Find the email field (not the id field)
	var emailField JField
	for _, field := range fields {
		if field.Name() == "email" {
			emailField = field
			break
		}
	}
	assert.NotNil(t, emailField)
	assert.NotNil(t, emailField.Validation())
}
