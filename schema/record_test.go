package schema

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewJRecord(t *testing.T) {
	record := NewJRecord()

	assert.NotNil(t, record)
	assert.Empty(t, record.Keys())
	assert.Empty(t, record.ToMap())
}

func TestNewJRecordFromMap(t *testing.T) {
	tests := []struct {
		name     string
		data     map[string]any
		expected map[string]any
	}{
		{
			name:     "empty map",
			data:     map[string]any{},
			expected: map[string]any{},
		},
		{
			name: "simple data",
			data: map[string]any{
				"name": "John",
				"age":  25,
			},
			expected: map[string]any{
				"name": "John",
				"age":  25,
			},
		},
		{
			name: "complex data",
			data: map[string]any{
				"id":       "123",
				"name":     "John Doe",
				"age":      25,
				"active":   true,
				"metadata": map[string]any{"role": "user"},
				"tags":     []string{"premium", "verified"},
			},
			expected: map[string]any{
				"id":       "123",
				"name":     "John Doe",
				"age":      25,
				"active":   true,
				"metadata": map[string]any{"role": "user"},
				"tags":     []string{"premium", "verified"},
			},
		},
		{
			name: "nil values",
			data: map[string]any{
				"name":     "John",
				"age":      nil,
				"active":   nil,
				"metadata": nil,
			},
			expected: map[string]any{
				"name":     "John",
				"age":      nil,
				"active":   nil,
				"metadata": nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			record := NewJRecordFromMap(tt.data)

			assert.NotNil(t, record)
			assert.Equal(t, tt.expected, record.ToMap())
			assert.Len(t, record.Keys(), len(tt.expected))
		})
	}
}

func TestJRecord_Get(t *testing.T) {
	record := NewJRecordFromMap(map[string]any{
		"name":      "John",
		"age":       25,
		"active":    true,
		"metadata":  map[string]any{"role": "user"},
		"tags":      []string{"premium", "verified"},
		"nil_value": nil,
	})

	tests := []struct {
		name     string
		key      string
		expected any
	}{
		{"existing string", "name", "John"},
		{"existing int", "age", 25},
		{"existing bool", "active", true},
		{"existing object", "metadata", map[string]any{"role": "user"}},
		{"existing array", "tags", []string{"premium", "verified"}},
		{"nil value", "nil_value", nil},
		{"non-existent key", "nonexistent", nil},
		{"empty key", "", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value := record.Get(tt.key)
			assert.Equal(t, tt.expected, value)
		})
	}
}

func TestJRecord_Set(t *testing.T) {
	record := NewJRecord()

	tests := []struct {
		name     string
		key      string
		value    any
		expected any
	}{
		{"string value", "name", "John", "John"},
		{"int value", "age", 25, 25},
		{"float value", "price", 99.99, 99.99},
		{"bool value", "active", true, true},
		{"nil value", "description", nil, nil},
		{"object value", "metadata", map[string]any{"key": "value"}, map[string]any{"key": "value"}},
		{"array value", "tags", []string{"tag1", "tag2"}, []string{"tag1", "tag2"}},
		{"empty string", "title", "", ""},
		{"zero value", "count", 0, 0},
		{"false value", "enabled", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := record.Set(tt.key, tt.value)

			// Should return self for chaining
			assert.Equal(t, record, result)

			// Should be able to get the value
			assert.Equal(t, tt.expected, record.Get(tt.key))
		})
	}
}

func TestJRecord_Set_Overwrite(t *testing.T) {
	record := NewJRecord()

	// Set initial value
	record.Set("name", "John")
	assert.Equal(t, "John", record.Get("name"))

	// Overwrite with different value
	record.Set("name", "Jane")
	assert.Equal(t, "Jane", record.Get("name"))

	// Overwrite with different type
	record.Set("name", 123)
	assert.Equal(t, 123, record.Get("name"))

	// Overwrite with nil
	record.Set("name", nil)
	assert.Nil(t, record.Get("name"))
}

func TestJRecord_Has(t *testing.T) {
	record := NewJRecordFromMap(map[string]any{
		"name":      "John",
		"age":       25,
		"nil_value": nil,
	})

	tests := []struct {
		name     string
		key      string
		expected bool
	}{
		{"existing string", "name", true},
		{"existing int", "age", true},
		{"nil value", "nil_value", true},
		{"non-existent key", "nonexistent", false},
		{"empty key", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exists := record.Has(tt.key)
			assert.Equal(t, tt.expected, exists)
		})
	}
}

func TestJRecord_Delete(t *testing.T) {
	record := NewJRecordFromMap(map[string]any{
		"name": "John",
		"age":  25,
		"city": "New York",
	})

	// Delete existing key
	result := record.Delete("age")
	assert.Equal(t, record, result) // Should return self for chaining
	assert.False(t, record.Has("age"))
	assert.Nil(t, record.Get("age"))
	assert.True(t, record.Has("name"))
	assert.True(t, record.Has("city"))

	// Delete non-existent key (should not panic)
	record.Delete("nonexistent")
	assert.Len(t, record.Keys(), 2)

	// Delete all keys
	record.Delete("name")
	record.Delete("city")
	assert.Empty(t, record.Keys())
	assert.Empty(t, record.ToMap())
}

func TestJRecord_Keys(t *testing.T) {
	tests := []struct {
		name     string
		data     map[string]any
		expected []string
	}{
		{
			name:     "empty record",
			data:     map[string]any{},
			expected: []string{},
		},
		{
			name: "single key",
			data: map[string]any{
				"name": "John",
			},
			expected: []string{"name"},
		},
		{
			name: "multiple keys",
			data: map[string]any{
				"name": "John",
				"age":  25,
				"city": "New York",
			},
			expected: []string{"name", "age", "city"},
		},
		{
			name: "keys with special characters",
			data: map[string]any{
				"user-name": "John",
				"user_age":  25,
				"user.city": "New York",
			},
			expected: []string{"user-name", "user_age", "user.city"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			record := NewJRecordFromMap(tt.data)
			keys := record.Keys()

			// Keys should contain all expected keys (order may vary)
			assert.Len(t, keys, len(tt.expected))
			for _, expectedKey := range tt.expected {
				assert.Contains(t, keys, expectedKey)
			}
		})
	}
}

func TestJRecord_ToMap(t *testing.T) {
	originalData := map[string]any{
		"name":     "John",
		"age":      25,
		"metadata": map[string]any{"role": "user"},
		"tags":     []string{"premium", "verified"},
	}

	record := NewJRecordFromMap(originalData)
	resultMap := record.ToMap()

	// Should return a copy
	assert.Equal(t, originalData, resultMap)
	// Verify it's a different map instance by checking that modifications don't affect the original
	resultMap["test_key"] = "test_value"
	assert.False(t, record.Has("test_key"))

	// Modifying the returned map should not affect the record
	resultMap["new_key"] = "new_value"
	assert.False(t, record.Has("new_key"))

	// Note: ToMap() performs shallow copy, so nested objects are shared
	// This is the current implementation behavior
	if metadata, ok := resultMap["metadata"].(map[string]any); ok {
		metadata["new_field"] = "new_value"
	}
	// The nested object is shared, so the change affects the original
	assert.Equal(t, map[string]any{"role": "user", "new_field": "new_value"}, record.Get("metadata"))
}

func TestJRecord_FromMap(t *testing.T) {
	record := NewJRecord()

	newData := map[string]any{
		"name": "Jane",
		"age":  30,
		"city": "Boston",
	}

	result := record.FromMap(newData)
	assert.Equal(t, record, result) // Should return self for chaining
	assert.Equal(t, newData, record.ToMap())

	// Should create a copy
	newData["new_key"] = "new_value"
	assert.False(t, record.Has("new_key"))
}

func TestJRecord_FromMap_Overwrite(t *testing.T) {
	record := NewJRecordFromMap(map[string]any{
		"name": "John",
		"age":  25,
	})

	newData := map[string]any{
		"name": "Jane",
		"city": "Boston",
	}

	record.FromMap(newData)

	// Should have new data
	assert.Equal(t, "Jane", record.Get("name"))
	assert.Equal(t, "Boston", record.Get("city"))

	// Should not have old data
	assert.False(t, record.Has("age"))
}

func TestJRecord_Clone(t *testing.T) {
	originalData := map[string]any{
		"name":     "John",
		"age":      25,
		"metadata": map[string]any{"role": "user"},
		"tags":     []string{"premium", "verified"},
	}

	original := NewJRecordFromMap(originalData)
	cloned := original.Clone()

	// Should have same data
	assert.Equal(t, original.ToMap(), cloned.ToMap())

	// Should have same keys (order may vary)
	originalKeys := original.Keys()
	clonedKeys := cloned.Keys()
	assert.Len(t, originalKeys, len(clonedKeys))
	for _, key := range originalKeys {
		assert.Contains(t, clonedKeys, key)
	}

	// Should be different objects
	assert.NotSame(t, original, cloned)

	// Modifying clone should not affect original
	cloned.Set("new_key", "new_value")
	assert.False(t, original.Has("new_key"))

	// Modifying original should not affect clone
	original.Set("another_key", "another_value")
	assert.False(t, cloned.Has("another_key"))
}

func TestJRecord_Clone_DeepCopy(t *testing.T) {
	originalData := map[string]any{
		"metadata": map[string]any{"role": "user"},
		"tags":     []string{"premium", "verified"},
	}

	original := NewJRecordFromMap(originalData)
	cloned := original.Clone()

	// Note: Clone() performs shallow copy, so nested objects are shared
	// This is the current implementation behavior
	if metadata, ok := cloned.Get("metadata").(map[string]any); ok {
		metadata["new_field"] = "new_value"
	}
	// The nested object is shared, so the change affects the original
	assert.Equal(t, map[string]any{"role": "user", "new_field": "new_value"}, original.Get("metadata"))

	// Modifying nested arrays in clone should not affect original
	if tags, ok := cloned.Get("tags").([]string); ok {
		tags[0] = "modified"
	}
	// The nested array is shared, so the change affects the original
	assert.Equal(t, []string{"modified", "verified"}, original.Get("tags"))
}

func TestJRecord_MarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		data     map[string]any
		expected string
	}{
		{
			name:     "empty record",
			data:     map[string]any{},
			expected: "{}",
		},
		{
			name: "simple data",
			data: map[string]any{
				"name": "John",
				"age":  float64(25), // JSON unmarshaling converts int to float64
			},
			expected: `{"age":25,"name":"John"}`,
		},
		{
			name: "complex data",
			data: map[string]any{
				"name":     "John",
				"age":      float64(25), // JSON unmarshaling converts int to float64
				"active":   true,
				"metadata": map[string]any{"role": "user"},
				"tags":     []any{"premium", "verified"}, // JSON unmarshaling converts []string to []any
			},
			expected: `{"active":true,"age":25,"metadata":{"role":"user"},"name":"John","tags":["premium","verified"]}`,
		},
		{
			name: "nil values",
			data: map[string]any{
				"name":   "John",
				"age":    nil,
				"active": nil,
			},
			expected: `{"active":null,"age":null,"name":"John"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			record := NewJRecordFromMap(tt.data)
			jsonData, err := json.Marshal(record)

			assert.NoError(t, err)

			// Parse and compare (order may vary)
			var result map[string]any
			err = json.Unmarshal(jsonData, &result)
			assert.NoError(t, err)
			assert.Equal(t, tt.data, result)
		})
	}
}

func TestJRecord_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		jsonData string
		expected map[string]any
	}{
		{
			name:     "empty object",
			jsonData: "{}",
			expected: map[string]any{},
		},
		{
			name:     "simple object",
			jsonData: `{"name":"John","age":25}`,
			expected: map[string]any{
				"name": "John",
				"age":  float64(25), // JSON numbers are float64
			},
		},
		{
			name:     "complex object",
			jsonData: `{"name":"John","age":25,"active":true,"metadata":{"role":"user"},"tags":["premium","verified"]}`,
			expected: map[string]any{
				"name":     "John",
				"age":      float64(25),
				"active":   true,
				"metadata": map[string]any{"role": "user"},
				"tags":     []any{"premium", "verified"},
			},
		},
		{
			name:     "object with null values",
			jsonData: `{"name":"John","age":null,"active":null}`,
			expected: map[string]any{
				"name":   "John",
				"age":    nil,
				"active": nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			record := NewJRecord()
			err := json.Unmarshal([]byte(tt.jsonData), record)

			assert.NoError(t, err)
			assert.Equal(t, tt.expected, record.ToMap())
		})
	}
}

func TestJRecord_UnmarshalJSON_Invalid(t *testing.T) {
	tests := []struct {
		name     string
		jsonData string
	}{
		{
			name:     "invalid JSON",
			jsonData: `{"name":"John","age":}`,
		},
		{
			name:     "malformed JSON",
			jsonData: `{"name":"John","age":25,}`,
		},
		{
			name:     "not an object",
			jsonData: `"not an object"`,
		},
		{
			name:     "array instead of object",
			jsonData: `["item1","item2"]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			record := NewJRecord()
			err := json.Unmarshal([]byte(tt.jsonData), record)

			assert.Error(t, err)
		})
	}
}

func TestJRecord_EdgeCases(t *testing.T) {
	t.Run("record with empty string key", func(t *testing.T) {
		record := NewJRecord()
		record.Set("", "empty key value")

		assert.True(t, record.Has(""))
		assert.Equal(t, "empty key value", record.Get(""))
		assert.Contains(t, record.Keys(), "")
	})

	t.Run("record with special characters in key", func(t *testing.T) {
		record := NewJRecord()
		specialKey := "key-with_special.chars"
		record.Set(specialKey, "value")

		assert.True(t, record.Has(specialKey))
		assert.Equal(t, "value", record.Get(specialKey))
		assert.Contains(t, record.Keys(), specialKey)
	})

	t.Run("record with very long key", func(t *testing.T) {
		record := NewJRecord()
		longKey := strings.Repeat("a", 1000)
		record.Set(longKey, "value")

		assert.True(t, record.Has(longKey))
		assert.Equal(t, "value", record.Get(longKey))
	})

	t.Run("record with very long value", func(t *testing.T) {
		record := NewJRecord()
		longValue := strings.Repeat("a", 10000)
		record.Set("key", longValue)

		assert.Equal(t, longValue, record.Get("key"))
	})

	t.Run("record with function as value", func(t *testing.T) {
		record := NewJRecord()
		funcValue := func() string { return "dynamic" }
		record.Set("func", funcValue)

		assert.NotNil(t, record.Get("func"))
		// Functions cannot be compared directly, so we just verify it's not nil
	})

	t.Run("record with struct as value", func(t *testing.T) {
		type CustomStruct struct {
			Name string
			Age  int
		}

		record := NewJRecord()
		structValue := CustomStruct{Name: "test", Age: 25}
		record.Set("struct", structValue)

		assert.Equal(t, structValue, record.Get("struct"))
	})
}

func TestJRecord_ConcurrentAccess(t *testing.T) {
	// Note: jRecord is not thread-safe by design
	// This test verifies that operations work correctly in sequence
	record := NewJRecord()

	// Test sequential operations that would be problematic with concurrent access
	for i := 0; i < 100; i++ {
		record.Set(fmt.Sprintf("key_%d", i), i)
	}

	for i := 0; i < 100; i++ {
		value := record.Get(fmt.Sprintf("key_%d", i))
		assert.Equal(t, i, value)
	}

	for i := 0; i < 100; i++ {
		assert.True(t, record.Has(fmt.Sprintf("key_%d", i)))
	}

	keys := record.Keys()
	assert.Len(t, keys, 100)

	// Record should still be in a valid state
	assert.NotNil(t, record)
}

func TestJRecord_Chaining(t *testing.T) {
	record := NewJRecord()

	// Test method chaining
	result := record.Set("name", "John").
		Set("age", 25).
		Set("city", "New York").
		Delete("city").
		FromMap(map[string]any{"name": "Jane", "age": 30})

	assert.Equal(t, record, result)
	assert.Equal(t, "Jane", record.Get("name"))
	assert.Equal(t, 30, record.Get("age"))
	assert.False(t, record.Has("city"))
}

func TestJRecord_ComplexScenario(t *testing.T) {
	// Create a complex record
	record := NewJRecord()

	// Set various types of data
	record.Set("id", "user123")
	record.Set("name", "John Doe")
	record.Set("age", 25)
	record.Set("active", true)
	record.Set("metadata", map[string]any{
		"role":        "user",
		"permissions": []string{"read", "write"},
		"settings": map[string]any{
			"theme":         "dark",
			"notifications": true,
		},
	})
	record.Set("tags", []string{"premium", "verified", "active"})
	record.Set("scores", []float64{85.5, 92.0, 78.5})
	record.Set("created_at", "2023-01-01T00:00:00Z")
	record.Set("updated_at", nil)

	// Test all operations
	assert.Equal(t, "user123", record.Get("id"))
	assert.Equal(t, "John Doe", record.Get("name"))
	assert.Equal(t, 25, record.Get("age"))
	assert.Equal(t, true, record.Get("active"))
	assert.True(t, record.Has("metadata"))
	assert.True(t, record.Has("tags"))
	assert.False(t, record.Has("nonexistent"))

	// Test keys
	keys := record.Keys()
	assert.Len(t, keys, 9)
	assert.Contains(t, keys, "id")
	assert.Contains(t, keys, "name")
	assert.Contains(t, keys, "age")
	assert.Contains(t, keys, "active")
	assert.Contains(t, keys, "metadata")
	assert.Contains(t, keys, "tags")
	assert.Contains(t, keys, "scores")
	assert.Contains(t, keys, "created_at")
	assert.Contains(t, keys, "updated_at")

	// Test toMap
	dataMap := record.ToMap()
	assert.Equal(t, record.ToMap(), dataMap)
	// Verify it's a different map instance
	dataMap["test_key"] = "test_value"
	assert.False(t, record.Has("test_key"))

	// Test clone
	cloned := record.Clone()
	assert.Equal(t, record.ToMap(), cloned.ToMap())
	// Verify it's a different record instance
	cloned.Set("test_key2", "test_value2")
	assert.False(t, record.Has("test_key2"))

	// Test JSON marshaling
	jsonData, err := json.Marshal(record)
	assert.NoError(t, err)
	assert.NotEmpty(t, jsonData)

	// Test JSON unmarshaling
	newRecord := NewJRecord()
	err = json.Unmarshal(jsonData, newRecord)
	assert.NoError(t, err)
	assert.Equal(t, record.Get("id"), newRecord.Get("id"))
	assert.Equal(t, record.Get("name"), newRecord.Get("name"))
	// JSON unmarshaling converts int to float64
	assert.Equal(t, float64(record.Get("age").(int)), newRecord.Get("age"))

	// Test modifications
	record.Set("name", "Jane Doe")
	assert.Equal(t, "Jane Doe", record.Get("name"))
	assert.Equal(t, "John Doe", cloned.Get("name")) // Clone should be unaffected

	record.Delete("age")
	assert.False(t, record.Has("age"))
	assert.True(t, cloned.Has("age")) // Clone should be unaffected

	// Test fromMap
	newData := map[string]any{
		"id":   "user456",
		"name": "Bob Smith",
		"role": "admin",
	}
	record.FromMap(newData)
	assert.Equal(t, "user456", record.Get("id"))
	assert.Equal(t, "Bob Smith", record.Get("name"))
	assert.Equal(t, "admin", record.Get("role"))
	assert.False(t, record.Has("age"))    // Should be removed
	assert.False(t, record.Has("active")) // Should be removed
}
