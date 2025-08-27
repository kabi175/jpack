package validations

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/kabi175/jpack"
)

// RequiredField creates a validation function that ensures a field is not null or empty string
func RequiredField(fieldName string) jpack.ValidationFunc {
	return func(ctx context.Context, record jpack.JRecord) error {
		field, exists := record.Schema().Field(fieldName)
		if !exists {
			return fmt.Errorf("field '%s' not found in schema", fieldName)
		}

		value, exists := record.Value(field)
		if !exists {
			return fmt.Errorf("field '%s' is required but not provided", fieldName)
		}

		// Check for nil values
		if value == nil {
			return fmt.Errorf("field '%s' is required but is null", fieldName)
		}

		// Check for empty strings
		if str, ok := value.(string); ok {
			if strings.TrimSpace(str) == "" {
				return fmt.Errorf("field '%s' is required but is empty", fieldName)
			}
		}

		// Check for zero values in numeric types
		reflectValue := reflect.ValueOf(value)
		switch reflectValue.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if reflectValue.Int() == 0 {
				return fmt.Errorf("field '%s' is required but is zero", fieldName)
			}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			if reflectValue.Uint() == 0 {
				return fmt.Errorf("field '%s' is required but is zero", fieldName)
			}
		case reflect.Float32, reflect.Float64:
			if reflectValue.Float() == 0 {
				return fmt.Errorf("field '%s' is required but is zero", fieldName)
			}
		}

		return nil
	}
}

// ReadOnlyField creates a validation function that prevents editing of a field after creation
func ReadOnlyField(fieldName string) jpack.ValidationFunc {
	return func(ctx context.Context, record jpack.JRecord) error {
		// If the record is new, allow setting the field
		if record.IsNew() {
			return nil
		}

		// If the record is modified and the field is in dirty keys, it's being edited
		if record.IsModified() {
			dirtyKeys := record.DirtyKeys()
			for _, key := range dirtyKeys {
				if key == fieldName {
					return fmt.Errorf("field '%s' is read-only and cannot be modified", fieldName)
				}
			}
		}

		return nil
	}
}

// StringLength creates a validation function that validates string length
func StringLength(fieldName string, minLength, maxLength int) jpack.ValidationFunc {
	return func(ctx context.Context, record jpack.JRecord) error {
		field, exists := record.Schema().Field(fieldName)
		if !exists {
			return fmt.Errorf("field '%s' not found in schema", fieldName)
		}

		value, exists := record.Value(field)
		if !exists {
			return nil // Field not set, skip validation
		}

		// Handle nil values
		if value == nil {
			if minLength > 0 {
				return fmt.Errorf("field '%s' can not be empty", fieldName)
			}
			return nil // Allow nil values when minLength is 0
		}

		// Convert to string
		var strValue string
		switch v := value.(type) {
		case string:
			strValue = v
		case *string:
			if v == nil {
				return nil
			}
			strValue = *v
		default:
			return fmt.Errorf("field '%s' must be a string, got %T", fieldName, value)
		}

		length := len(strings.TrimSpace(strValue))

		if minLength > 0 && length < minLength {
			return fmt.Errorf("field '%s' must be at least %d characters long, got %d", fieldName, minLength, length)
		}

		if maxLength > 0 && length > maxLength {
			return fmt.Errorf("field '%s' must be at most %d characters long, got %d", fieldName, maxLength, length)
		}

		return nil
	}
}

// NumberRange creates a validation function that validates numeric ranges
func NumberRange(fieldName string, min, max float64) jpack.ValidationFunc {
	return func(ctx context.Context, record jpack.JRecord) error {
		field, exists := record.Schema().Field(fieldName)
		if !exists {
			return fmt.Errorf("field '%s' not found in schema", fieldName)
		}

		value, exists := record.Value(field)
		if !exists {
			return nil // Field not set, skip validation
		}

		// Handle nil values
		if value == nil {
			return nil // Allow nil values, use RequiredField if you want to prevent them
		}

		// Convert to float64
		var numValue float64
		reflectValue := reflect.ValueOf(value)

		switch reflectValue.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			numValue = float64(reflectValue.Int())
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			numValue = float64(reflectValue.Uint())
		case reflect.Float32, reflect.Float64:
			numValue = reflectValue.Float()
		default:
			return fmt.Errorf("field '%s' must be a number, got %T", fieldName, value)
		}

		if min > 0 && numValue < min {
			return fmt.Errorf("field '%s' must be at least %g, got %g", fieldName, min, numValue)
		}

		if max > 0 && numValue > max {
			return fmt.Errorf("field '%s' must be at most %g, got %g", fieldName, max, numValue)
		}

		return nil
	}
}

// EmailFormat creates a validation function that validates email format
func EmailFormat(fieldName string) jpack.ValidationFunc {
	return func(ctx context.Context, record jpack.JRecord) error {
		field, exists := record.Schema().Field(fieldName)
		if !exists {
			return fmt.Errorf("field '%s' not found in schema", fieldName)
		}

		value, exists := record.Value(field)
		if !exists {
			return nil // Field not set, skip validation
		}

		// Handle nil values
		if value == nil {
			return nil // Allow nil values, use RequiredField if you want to prevent them
		}

		// Convert to string
		var strValue string
		switch v := value.(type) {
		case string:
			strValue = v
		case *string:
			if v == nil {
				return nil
			}
			strValue = *v
		default:
			return fmt.Errorf("field '%s' must be a string, got %T", fieldName, value)
		}

		strValue = strings.TrimSpace(strValue)
		if strValue == "" {
			return nil // Allow empty strings, use RequiredField if you want to prevent them
		}

		// Basic email validation
		if !strings.Contains(strValue, "@") {
			return fmt.Errorf("field '%s' must be a valid email address (missing @)", fieldName)
		}

		parts := strings.Split(strValue, "@")
		if len(parts) != 2 {
			return fmt.Errorf("field '%s' must be a valid email address (invalid format)", fieldName)
		}

		localPart := parts[0]
		domainPart := parts[1]

		if len(localPart) == 0 {
			return fmt.Errorf("field '%s' must be a valid email address (missing local part)", fieldName)
		}

		if len(domainPart) == 0 {
			return fmt.Errorf("field '%s' must be a valid email address (missing domain)", fieldName)
		}

		if !strings.Contains(domainPart, ".") {
			return fmt.Errorf("field '%s' must be a valid email address (invalid domain)", fieldName)
		}

		return nil
	}
}

// PositiveNumber creates a validation function that ensures a number is positive
func PositiveNumber(fieldName string) jpack.ValidationFunc {
	return NumberRange(fieldName, 0.000001, 0) // min > 0, max = 0 means no upper limit
}

// NonNegativeNumber creates a validation function that ensures a number is non-negative
func NonNegativeNumber(fieldName string) jpack.ValidationFunc {
	return func(ctx context.Context, record jpack.JRecord) error {
		field, exists := record.Schema().Field(fieldName)
		if !exists {
			return fmt.Errorf("field '%s' not found in schema", fieldName)
		}

		value, exists := record.Value(field)
		if !exists {
			return nil // Field not set, skip validation
		}

		// Handle nil values
		if value == nil {
			return nil // Allow nil values
		}

		// Convert to float64
		var numValue float64
		reflectValue := reflect.ValueOf(value)

		switch reflectValue.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			numValue = float64(reflectValue.Int())
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			numValue = float64(reflectValue.Uint())
		case reflect.Float32, reflect.Float64:
			numValue = reflectValue.Float()
		default:
			return fmt.Errorf("field '%s' must be a number, got %T", fieldName, value)
		}

		if numValue < 0 {
			return fmt.Errorf("field '%s' must be at least 0, got %g", fieldName, numValue)
		}

		return nil
	}
}

// StringNotEmpty creates a validation function that ensures a string is not empty (but allows nil)
func StringNotEmpty(fieldName string) jpack.ValidationFunc {
	return StringLength(fieldName, 1, 0) // min = 1, max = 0 means no upper limit
}
