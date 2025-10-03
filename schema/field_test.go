package schema

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewJField(t *testing.T) {
	tests := []struct {
		name         string
		fieldName    string
		fieldType    JFieldType
		defaultValue any
		expectedType JFieldType
	}{
		{"string field", "name", JString, "default", JString},
		{"int field", "age", JInt, 25, JInt},
		{"int64 field", "timestamp", JInt64, int64(1234567890), JInt64},
		{"float field", "price", JFloat, 99.99, JFloat},
		{"float64 field", "rate", JFloat64, 99.99, JFloat64},
		{"bool field", "active", JBool, true, JBool},
		{"time field", "created_at", JTime, "2023-01-01T00:00:00Z", JTime},
		{"object field", "metadata", JObject, map[string]any{"key": "value"}, JObject},
		{"array field", "tags", JArray, []string{"tag1", "tag2"}, JArray},
		{"binary field", "data", JBinary, []byte("binary data"), JBinary},
		{"objectid field", "id", JObjectID, "507f1f77bcf86cd799439011", JObjectID},
		{"nil default", "description", JString, nil, JString},
		{"empty string", "title", JString, "", JString},
		{"zero value", "count", JInt, 0, JInt},
		{"false bool", "enabled", JBool, false, JBool},
		{"empty array", "items", JArray, []string{}, JArray},
		{"empty object", "config", JObject, map[string]any{}, JObject},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := NewJField(tt.fieldName, tt.fieldType, tt.defaultValue)

			assert.Equal(t, tt.fieldName, field.Name())
			assert.Equal(t, tt.expectedType, field.Type())
			assert.Equal(t, tt.defaultValue, field.DefaultValue())
			assert.False(t, field.IsRequired())
			assert.False(t, field.IsUnique())
			assert.False(t, field.IsImmutable())
			assert.Nil(t, field.Validation())
		})
	}
}

func TestJField_SetRequired(t *testing.T) {
	field := NewJField("name", JString, "default")

	// Test setting required to true
	result := field.SetRequired(true)
	assert.Equal(t, field, result) // Should return self for chaining
	assert.True(t, field.IsRequired())

	// Test setting required to false
	field.SetRequired(false)
	assert.False(t, field.IsRequired())

	// Test chaining
	field.SetRequired(true).SetRequired(false)
	assert.False(t, field.IsRequired())
}

func TestJField_SetRequired_Immutable(t *testing.T) {
	field := NewJField("name", JString, "default")
	field.Freeze()

	assert.Panics(t, func() {
		field.SetRequired(true)
	}, "should panic when modifying immutable field")
}

func TestJField_SetUnique(t *testing.T) {
	field := NewJField("email", JString, "")

	// Test setting unique to true
	result := field.SetUnique(true)
	assert.Equal(t, field, result) // Should return self for chaining
	assert.True(t, field.IsUnique())

	// Test setting unique to false
	field.SetUnique(false)
	assert.False(t, field.IsUnique())

	// Test chaining
	field.SetUnique(true).SetUnique(false)
	assert.False(t, field.IsUnique())
}

func TestJField_SetUnique_Immutable(t *testing.T) {
	field := NewJField("email", JString, "")
	field.Freeze()

	assert.Panics(t, func() {
		field.SetUnique(true)
	}, "should panic when modifying immutable field")
}

func TestJField_SetValidation(t *testing.T) {
	field := NewJField("age", JInt, 0)

	// Test setting validation function
	validationFunc := func(ctx context.Context, rec JRecord) error {
		age := rec.Get("age")
		if age != nil {
			if ageInt, ok := age.(int); ok && ageInt < 0 {
				return assert.AnError
			}
		}
		return nil
	}

	result := field.SetValidation(validationFunc)
	assert.Equal(t, field, result) // Should return self for chaining
	assert.NotNil(t, field.Validation())

	// Test setting validation to nil
	field.SetValidation(nil)
	assert.Nil(t, field.Validation())
}

func TestJField_SetValidation_Immutable(t *testing.T) {
	field := NewJField("age", JInt, 0)
	field.Freeze()

	validationFunc := func(ctx context.Context, rec JRecord) error {
		return nil
	}

	assert.Panics(t, func() {
		field.SetValidation(validationFunc)
	}, "should panic when modifying immutable field")
}

func TestJField_Freeze(t *testing.T) {
	field := NewJField("name", JString, "default")
	field.SetRequired(true).SetUnique(true)

	// Test freezing
	result := field.Freeze()
	assert.Equal(t, field, result) // Should return self for chaining
	assert.True(t, field.IsImmutable())

	// Test that modifications are now blocked
	assert.Panics(t, func() {
		field.SetRequired(false)
	})

	assert.Panics(t, func() {
		field.SetUnique(false)
	})

	assert.Panics(t, func() {
		field.SetValidation(nil)
	})
}

func TestJField_Freeze_AlreadyImmutable(t *testing.T) {
	field := NewJField("name", JString, "default")
	field.Freeze()

	// Freezing already immutable field should work
	result := field.Freeze()
	assert.Equal(t, field, result)
	assert.True(t, field.IsImmutable())
}

func TestJField_EdgeCases(t *testing.T) {
	t.Run("empty field name", func(t *testing.T) {
		field := NewJField("", JString, "default")
		assert.Equal(t, "", field.Name())
	})

	t.Run("field with special characters in name", func(t *testing.T) {
		field := NewJField("field-name_123", JString, "default")
		assert.Equal(t, "field-name_123", field.Name())
	})

	t.Run("field with complex default value", func(t *testing.T) {
		complexValue := map[string]any{
			"nested": map[string]any{
				"array":  []string{"item1", "item2"},
				"number": 42,
			},
		}
		field := NewJField("complex", JObject, complexValue)
		assert.Equal(t, complexValue, field.DefaultValue())
	})

	t.Run("field with function as default value", func(t *testing.T) {
		funcValue := func() string { return "dynamic" }
		field := NewJField("dynamic", JString, funcValue)
		assert.NotNil(t, field.DefaultValue())
		// Functions cannot be compared directly, so we just verify it's not nil
	})

	t.Run("field with struct as default value", func(t *testing.T) {
		type CustomStruct struct {
			Name string
			Age  int
		}
		structValue := CustomStruct{Name: "test", Age: 25}
		field := NewJField("custom", JObject, structValue)
		assert.Equal(t, structValue, field.DefaultValue())
	})
}

func TestJField_Validation_EdgeCases(t *testing.T) {
	t.Run("validation with nil context", func(t *testing.T) {
		field := NewJField("test", JString, "default")
		field.SetValidation(func(ctx context.Context, rec JRecord) error {
			if ctx == nil {
				return assert.AnError
			}
			return nil
		})

		// This test is more about ensuring the validation function can handle nil context
		// In practice, context should never be nil when called from schema validation
		assert.NotNil(t, field.Validation())
	})

	t.Run("validation with nil record", func(t *testing.T) {
		field := NewJField("test", JString, "default")
		field.SetValidation(func(ctx context.Context, rec JRecord) error {
			if rec == nil {
				return assert.AnError
			}
			return nil
		})

		assert.NotNil(t, field.Validation())
	})

	t.Run("validation that panics", func(t *testing.T) {
		field := NewJField("test", JString, "default")
		field.SetValidation(func(ctx context.Context, rec JRecord) error {
			panic("validation panic")
		})

		// Test that the validation function is stored correctly
		assert.NotNil(t, field.Validation())
	})
}

func TestJField_AllFieldTypes(t *testing.T) {
	fieldTypes := []JFieldType{
		JString, JInt, JInt64, JFloat, JFloat64, JBool, JTime, JObject, JArray, JBinary, JObjectID,
	}

	for _, fieldType := range fieldTypes {
		t.Run(string(fieldType), func(t *testing.T) {
			field := NewJField("test", fieldType, nil)
			assert.Equal(t, fieldType, field.Type())
			assert.Equal(t, "test", field.Name())
			assert.Nil(t, field.DefaultValue())
		})
	}
}

func TestJField_Chaining(t *testing.T) {
	field := NewJField("email", JString, "")

	// Test method chaining
	result := field.SetRequired(true).SetUnique(true).SetValidation(func(ctx context.Context, rec JRecord) error {
		return nil
	})

	assert.Equal(t, field, result)
	assert.True(t, field.IsRequired())
	assert.True(t, field.IsUnique())
	assert.NotNil(t, field.Validation())
}

func TestJField_ConcurrentAccess(t *testing.T) {
	field := NewJField("concurrent", JString, "default")

	// Test concurrent modifications
	done := make(chan bool, 3)

	go func() {
		defer func() { done <- true }()
		for i := 0; i < 100; i++ {
			field.SetRequired(i%2 == 0)
		}
	}()

	go func() {
		defer func() { done <- true }()
		for i := 0; i < 100; i++ {
			field.SetUnique(i%2 == 0)
		}
	}()

	go func() {
		defer func() { done <- true }()
		for i := 0; i < 100; i++ {
			field.SetValidation(func(ctx context.Context, rec JRecord) error {
				return nil
			})
		}
	}()

	// Wait for all goroutines to complete
	for i := 0; i < 3; i++ {
		<-done
	}

	// Field should still be in a valid state
	assert.Equal(t, "concurrent", field.Name())
	assert.Equal(t, JString, field.Type())
	assert.Equal(t, "default", field.DefaultValue())
}

func TestJField_ComplexValidation(t *testing.T) {
	field := NewJField("email", JString, "")

	// Add complex validation
	field.SetValidation(func(ctx context.Context, rec JRecord) error {
		email := rec.Get("email")
		if email == nil {
			return nil // Let required validation handle this
		}

		emailStr, ok := email.(string)
		if !ok {
			return assert.AnError
		}

		// Simple email validation
		if len(emailStr) == 0 {
			return assert.AnError
		}

		if !strings.Contains(emailStr, "@") {
			return assert.AnError
		}

		return nil
	})

	// Test validation function
	ctx := context.Background()

	// Valid email
	validRecord := NewJRecordFromMap(map[string]any{
		"email": "test@example.com",
	})
	err := field.Validation()(ctx, validRecord)
	assert.NoError(t, err)

	// Invalid email - no @
	invalidRecord1 := NewJRecordFromMap(map[string]any{
		"email": "invalid-email",
	})
	err = field.Validation()(ctx, invalidRecord1)
	assert.Error(t, err)

	// Invalid email - empty string
	invalidRecord2 := NewJRecordFromMap(map[string]any{
		"email": "",
	})
	err = field.Validation()(ctx, invalidRecord2)
	assert.Error(t, err)

	// Invalid email - wrong type
	invalidRecord3 := NewJRecordFromMap(map[string]any{
		"email": 123,
	})
	err = field.Validation()(ctx, invalidRecord3)
	assert.Error(t, err)
}

func TestJField_DefaultValueTypes(t *testing.T) {
	tests := []struct {
		name         string
		fieldType    JFieldType
		defaultValue any
		expectError  bool
	}{
		{"string with string default", JString, "default", false},
		{"string with nil default", JString, nil, false},
		{"int with int default", JInt, 42, false},
		{"int with nil default", JInt, nil, false},
		{"float with float default", JFloat64, 3.14, false},
		{"bool with bool default", JBool, true, false},
		{"array with array default", JArray, []string{"item1", "item2"}, false},
		{"object with object default", JObject, map[string]any{"key": "value"}, false},
		{"binary with binary default", JBinary, []byte("data"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := NewJField("test", tt.fieldType, tt.defaultValue)
			assert.Equal(t, tt.defaultValue, field.DefaultValue())
		})
	}
}

func TestJField_StringRepresentation(t *testing.T) {
	field := NewJField("name", JString, "default")
	field.SetRequired(true).SetUnique(true)

	// Test that field can be used in string operations
	str := fmt.Sprintf("Field: %s (%s)", field.Name(), field.Type())
	assert.Equal(t, "Field: name (string)", str)
}

func TestJField_Comparison(t *testing.T) {
	field1 := NewJField("name", JString, "default")
	field2 := NewJField("name", JString, "default")
	field3 := NewJField("age", JInt, 0)

	// Fields with same properties should be equal in terms of content
	assert.Equal(t, field1.Name(), field2.Name())
	assert.Equal(t, field1.Type(), field2.Type())
	assert.Equal(t, field1.DefaultValue(), field2.DefaultValue())

	// Different fields should have different properties
	assert.NotEqual(t, field1.Name(), field3.Name())
	assert.NotEqual(t, field1.Type(), field3.Type())
}
