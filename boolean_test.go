package jpack

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBoolean_Validate(t *testing.T) {
	booleanField := &Boolean{}

	t.Run("Valid boolean true", func(t *testing.T) {
		err := booleanField.Validate(true)
		assert.NoError(t, err)
	})

	t.Run("Valid boolean false", func(t *testing.T) {
		err := booleanField.Validate(false)
		assert.NoError(t, err)
	})

	t.Run("Valid string true", func(t *testing.T) {
		err := booleanField.Validate("true")
		assert.NoError(t, err)
	})

	t.Run("Valid string false", func(t *testing.T) {
		err := booleanField.Validate("false")
		assert.NoError(t, err)
	})

	t.Run("Valid string 1", func(t *testing.T) {
		err := booleanField.Validate("1")
		assert.NoError(t, err)
	})

	t.Run("Valid string 0", func(t *testing.T) {
		err := booleanField.Validate("0")
		assert.NoError(t, err)
	})

	t.Run("Valid string yes", func(t *testing.T) {
		err := booleanField.Validate("yes")
		assert.NoError(t, err)
	})

	t.Run("Valid string no", func(t *testing.T) {
		err := booleanField.Validate("no")
		assert.NoError(t, err)
	})

	t.Run("Valid string on", func(t *testing.T) {
		err := booleanField.Validate("on")
		assert.NoError(t, err)
	})

	t.Run("Valid string off", func(t *testing.T) {
		err := booleanField.Validate("off")
		assert.NoError(t, err)
	})

	t.Run("Valid string enabled", func(t *testing.T) {
		err := booleanField.Validate("enabled")
		assert.NoError(t, err)
	})

	t.Run("Valid string disabled", func(t *testing.T) {
		err := booleanField.Validate("disabled")
		assert.NoError(t, err)
	})

	t.Run("Valid string with spaces", func(t *testing.T) {
		err := booleanField.Validate("  true  ")
		assert.NoError(t, err)
	})

	t.Run("Valid string case insensitive", func(t *testing.T) {
		err := booleanField.Validate("TRUE")
		assert.NoError(t, err)
	})

	t.Run("Valid integer 1", func(t *testing.T) {
		err := booleanField.Validate(1)
		assert.NoError(t, err)
	})

	t.Run("Valid integer 0", func(t *testing.T) {
		err := booleanField.Validate(0)
		assert.NoError(t, err)
	})

	t.Run("Valid integer 42", func(t *testing.T) {
		err := booleanField.Validate(42)
		assert.NoError(t, err)
	})

	t.Run("Valid float 1.0", func(t *testing.T) {
		err := booleanField.Validate(1.0)
		assert.NoError(t, err)
	})

	t.Run("Valid float 0.0", func(t *testing.T) {
		err := booleanField.Validate(0.0)
		assert.NoError(t, err)
	})

	t.Run("Valid float 3.14", func(t *testing.T) {
		err := booleanField.Validate(3.14)
		assert.NoError(t, err)
	})

	t.Run("Valid pointer to boolean", func(t *testing.T) {
		value := true
		err := booleanField.Validate(&value)
		assert.NoError(t, err)
	})

	t.Run("Valid nil pointer", func(t *testing.T) {
		var value *bool
		err := booleanField.Validate(value)
		assert.NoError(t, err)
	})

	t.Run("Valid nil value", func(t *testing.T) {
		err := booleanField.Validate(nil)
		assert.NoError(t, err)
	})

	t.Run("Invalid string", func(t *testing.T) {
		err := booleanField.Validate("invalid")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid boolean string value")
	})

	t.Run("Invalid type", func(t *testing.T) {
		err := booleanField.Validate([]string{"test"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be converted to boolean")
	})
}

func TestBoolean_Scan(t *testing.T) {
	booleanField := &Boolean{}
	field := &mockField{name: "active"}

	t.Run("Scan boolean true from database", func(t *testing.T) {
		row := map[string]any{"active": true}
		value, err := booleanField.Scan(context.Background(), field, row)
		assert.NoError(t, err)
		assert.Equal(t, true, value)
	})

	t.Run("Scan boolean false from database", func(t *testing.T) {
		row := map[string]any{"active": false}
		value, err := booleanField.Scan(context.Background(), field, row)
		assert.NoError(t, err)
		assert.Equal(t, false, value)
	})

	t.Run("Scan string true from database", func(t *testing.T) {
		row := map[string]any{"active": "true"}
		value, err := booleanField.Scan(context.Background(), field, row)
		assert.NoError(t, err)
		assert.Equal(t, true, value)
	})

	t.Run("Scan string false from database", func(t *testing.T) {
		row := map[string]any{"active": "false"}
		value, err := booleanField.Scan(context.Background(), field, row)
		assert.NoError(t, err)
		assert.Equal(t, false, value)
	})

	t.Run("Scan integer 1 from database", func(t *testing.T) {
		row := map[string]any{"active": 1}
		value, err := booleanField.Scan(context.Background(), field, row)
		assert.NoError(t, err)
		assert.Equal(t, true, value)
	})

	t.Run("Scan integer 0 from database", func(t *testing.T) {
		row := map[string]any{"active": 0}
		value, err := booleanField.Scan(context.Background(), field, row)
		assert.NoError(t, err)
		assert.Equal(t, false, value)
	})

	t.Run("Scan float 1.0 from database", func(t *testing.T) {
		row := map[string]any{"active": 1.0}
		value, err := booleanField.Scan(context.Background(), field, row)
		assert.NoError(t, err)
		assert.Equal(t, true, value)
	})

	t.Run("Scan float 0.0 from database", func(t *testing.T) {
		row := map[string]any{"active": 0.0}
		value, err := booleanField.Scan(context.Background(), field, row)
		assert.NoError(t, err)
		assert.Equal(t, false, value)
	})

	t.Run("Scan nil value from database", func(t *testing.T) {
		row := map[string]any{"active": nil}
		value, err := booleanField.Scan(context.Background(), field, row)
		assert.NoError(t, err)
		assert.Nil(t, value)
	})

	t.Run("Scan missing field from database", func(t *testing.T) {
		row := map[string]any{"other": true}
		value, err := booleanField.Scan(context.Background(), field, row)
		assert.NoError(t, err)
		assert.Nil(t, value)
	})

	t.Run("Scan invalid string from database", func(t *testing.T) {
		row := map[string]any{"active": "invalid"}
		value, err := booleanField.Scan(context.Background(), field, row)
		assert.Error(t, err)
		assert.Nil(t, value)
		assert.Contains(t, err.Error(), "invalid boolean string value")
	})

	t.Run("Scan invalid type from database", func(t *testing.T) {
		row := map[string]any{"active": []string{"test"}}
		value, err := booleanField.Scan(context.Background(), field, row)
		assert.Error(t, err)
		assert.Nil(t, value)
		assert.Contains(t, err.Error(), "cannot be converted to boolean")
	})
}

func TestBoolean_SetValue(t *testing.T) {
	booleanField := &Boolean{}
	field := &mockField{name: "active"}

	t.Run("Set boolean true value", func(t *testing.T) {
		row := make(map[string]any)
		err := booleanField.SetValue(context.Background(), field, true, row)
		assert.NoError(t, err)
		assert.Equal(t, true, row["active"])
	})

	t.Run("Set boolean false value", func(t *testing.T) {
		row := make(map[string]any)
		err := booleanField.SetValue(context.Background(), field, false, row)
		assert.NoError(t, err)
		assert.Equal(t, false, row["active"])
	})

	t.Run("Set string true value", func(t *testing.T) {
		row := make(map[string]any)
		err := booleanField.SetValue(context.Background(), field, "true", row)
		assert.NoError(t, err)
		assert.Equal(t, true, row["active"])
	})

	t.Run("Set string false value", func(t *testing.T) {
		row := make(map[string]any)
		err := booleanField.SetValue(context.Background(), field, "false", row)
		assert.NoError(t, err)
		assert.Equal(t, false, row["active"])
	})

	t.Run("Set string 1 value", func(t *testing.T) {
		row := make(map[string]any)
		err := booleanField.SetValue(context.Background(), field, "1", row)
		assert.NoError(t, err)
		assert.Equal(t, true, row["active"])
	})

	t.Run("Set string 0 value", func(t *testing.T) {
		row := make(map[string]any)
		err := booleanField.SetValue(context.Background(), field, "0", row)
		assert.NoError(t, err)
		assert.Equal(t, false, row["active"])
	})

	t.Run("Set integer 1 value", func(t *testing.T) {
		row := make(map[string]any)
		err := booleanField.SetValue(context.Background(), field, 1, row)
		assert.NoError(t, err)
		assert.Equal(t, true, row["active"])
	})

	t.Run("Set integer 0 value", func(t *testing.T) {
		row := make(map[string]any)
		err := booleanField.SetValue(context.Background(), field, 0, row)
		assert.NoError(t, err)
		assert.Equal(t, false, row["active"])
	})

	t.Run("Set float 1.0 value", func(t *testing.T) {
		row := make(map[string]any)
		err := booleanField.SetValue(context.Background(), field, 1.0, row)
		assert.NoError(t, err)
		assert.Equal(t, true, row["active"])
	})

	t.Run("Set float 0.0 value", func(t *testing.T) {
		row := make(map[string]any)
		err := booleanField.SetValue(context.Background(), field, 0.0, row)
		assert.NoError(t, err)
		assert.Equal(t, false, row["active"])
	})

	t.Run("Set nil value", func(t *testing.T) {
		row := make(map[string]any)
		err := booleanField.SetValue(context.Background(), field, nil, row)
		assert.NoError(t, err)
		assert.Nil(t, row["active"])
	})

	t.Run("Set invalid string value", func(t *testing.T) {
		row := make(map[string]any)
		err := booleanField.SetValue(context.Background(), field, "invalid", row)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid boolean string value")
	})

	t.Run("Set invalid type value", func(t *testing.T) {
		row := make(map[string]any)
		err := booleanField.SetValue(context.Background(), field, []string{"test"}, row)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be converted to boolean")
	})
}

func TestBoolean_StringConversions(t *testing.T) {
	booleanField := &Boolean{}

	testCases := []struct {
		input    string
		expected bool
		valid    bool
	}{
		{"true", true, true},
		{"TRUE", true, true},
		{"True", true, true},
		{"false", false, true},
		{"FALSE", false, true},
		{"False", false, true},
		{"1", true, true},
		{"0", false, true},
		{"yes", true, true},
		{"YES", true, true},
		{"no", false, true},
		{"NO", false, true},
		{"on", true, true},
		{"ON", true, true},
		{"off", false, true},
		{"OFF", false, true},
		{"enabled", true, true},
		{"ENABLED", true, true},
		{"disabled", false, true},
		{"DISABLED", false, true},
		{"", false, true},
		{"  true  ", true, true},
		{"  false  ", false, true},
		{"invalid", false, false},
		{"maybe", false, false},
		{"2", false, false},
	}

	for _, tc := range testCases {
		t.Run("String_"+tc.input, func(t *testing.T) {
			if tc.valid {
				err := booleanField.Validate(tc.input)
				assert.NoError(t, err)
			} else {
				err := booleanField.Validate(tc.input)
				assert.Error(t, err)
			}
		})
	}
}

func TestBoolean_NumericConversions(t *testing.T) {
	booleanField := &Boolean{}

	testCases := []struct {
		input    any
		expected bool
	}{
		{1, true},
		{0, false},
		{42, true},
		{-1, true},
		{1.0, true},
		{0.0, false},
		{3.14, true},
		{-0.5, true},
		{int8(1), true},
		{int16(0), false},
		{int32(42), true},
		{int64(0), false},
		{uint(1), true},
		{uint8(0), false},
		{uint16(42), true},
		{uint32(0), false},
		{uint64(1), true},
		{float32(1.0), true},
		{float64(0.0), false},
	}

	for _, tc := range testCases {
		t.Run("Numeric_"+fmt.Sprintf("%v", tc.input), func(t *testing.T) {
			err := booleanField.Validate(tc.input)
			assert.NoError(t, err)
		})
	}
}

func TestBoolean_PointerHandling(t *testing.T) {
	booleanField := &Boolean{}

	t.Run("Pointer to true", func(t *testing.T) {
		value := true
		err := booleanField.Validate(&value)
		assert.NoError(t, err)
	})

	t.Run("Pointer to false", func(t *testing.T) {
		value := false
		err := booleanField.Validate(&value)
		assert.NoError(t, err)
	})

	t.Run("Nil pointer", func(t *testing.T) {
		var value *bool
		err := booleanField.Validate(value)
		assert.NoError(t, err)
	})

	t.Run("Pointer to string", func(t *testing.T) {
		value := "true"
		err := booleanField.Validate(&value)
		assert.Error(t, err)
	})
}

func TestBoolean_Integration(t *testing.T) {
	booleanField := &Boolean{}
	field := &mockField{name: "isActive"}

	t.Run("Full workflow with boolean", func(t *testing.T) {
		// Validate
		err := booleanField.Validate(true)
		assert.NoError(t, err)

		// Set value
		row := make(map[string]any)
		err = booleanField.SetValue(context.Background(), field, true, row)
		assert.NoError(t, err)
		assert.Equal(t, true, row["isActive"])

		// Scan value
		value, err := booleanField.Scan(context.Background(), field, row)
		assert.NoError(t, err)
		assert.Equal(t, true, value)
	})

	t.Run("Full workflow with string", func(t *testing.T) {
		// Validate
		err := booleanField.Validate("yes")
		assert.NoError(t, err)

		// Set value
		row := make(map[string]any)
		err = booleanField.SetValue(context.Background(), field, "yes", row)
		assert.NoError(t, err)
		assert.Equal(t, true, row["isActive"])

		// Scan value
		value, err := booleanField.Scan(context.Background(), field, row)
		assert.NoError(t, err)
		assert.Equal(t, true, value)
	})

	t.Run("Full workflow with integer", func(t *testing.T) {
		// Validate
		err := booleanField.Validate(1)
		assert.NoError(t, err)

		// Set value
		row := make(map[string]any)
		err = booleanField.SetValue(context.Background(), field, 1, row)
		assert.NoError(t, err)
		assert.Equal(t, true, row["isActive"])

		// Scan value
		value, err := booleanField.Scan(context.Background(), field, row)
		assert.NoError(t, err)
		assert.Equal(t, true, value)
	})
}
