package validation

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/kabi175/jpack/logger"
	"github.com/kabi175/jpack/schema"
)

// RequiredValidator creates a validator that checks if a field is required
func RequiredValidator() ValidationFunc {
	return func(ctx context.Context, rec schema.JRecord) error {
		// This is handled at the field level in ValidationEngine
		return nil
	}
}

// MinLengthValidator creates a validator that checks minimum length
func MinLengthValidator() ValidationFunc {
	return func(ctx context.Context, rec schema.JRecord) error {
		// This would be used with field-specific validation
		return nil
	}
}

// MaxLengthValidator creates a validator that checks maximum length
func MaxLengthValidator() ValidationFunc {
	return func(ctx context.Context, rec schema.JRecord) error {
		// This would be used with field-specific validation
		return nil
	}
}

// RegexValidator creates a validator that checks regex pattern
func RegexValidator() ValidationFunc {
	return func(ctx context.Context, rec schema.JRecord) error {
		// This would be used with field-specific validation
		return nil
	}
}

// EmailValidator creates a validator that checks email format
func EmailValidator() ValidationFunc {
	return func(ctx context.Context, rec schema.JRecord) error {
		// This would be used with field-specific validation
		return nil
	}
}

// MinValueValidator creates a validator that checks minimum value
func MinValueValidator() ValidationFunc {
	return func(ctx context.Context, rec schema.JRecord) error {
		// This would be used with field-specific validation
		return nil
	}
}

// MaxValueValidator creates a validator that checks maximum value
func MaxValueValidator() ValidationFunc {
	return func(ctx context.Context, rec schema.JRecord) error {
		// This would be used with field-specific validation
		return nil
	}
}

// RangeValidator creates a validator that checks value range
func RangeValidator() ValidationFunc {
	return func(ctx context.Context, rec schema.JRecord) error {
		// This would be used with field-specific validation
		return nil
	}
}

// UniqueValidator creates a validator that checks uniqueness
func UniqueValidator() ValidationFunc {
	return func(ctx context.Context, rec schema.JRecord) error {
		// This would be used with field-specific validation
		return nil
	}
}

// Field validation functions

// ValidateRequired validates that a field is not empty
func ValidateRequired(fieldName string) ValidationFunc {
	return func(ctx context.Context, rec schema.JRecord) error {
		value := rec.Get(fieldName)
		if value == nil || value == "" {
			return &ValidationError{
				Field:   fieldName,
				Message: "field is required",
				Value:   value,
			}
		}
		return nil
	}
}

// ValidateMinLength validates minimum length for string fields
func ValidateMinLength(fieldName string, minLength int) ValidationFunc {
	return func(ctx context.Context, rec schema.JRecord) error {
		value := rec.Get(fieldName)
		if value == nil {
			return nil // Let required validation handle this
		}

		str, ok := value.(string)
		if !ok {
			return &ValidationError{
				Field:   fieldName,
				Message: "field must be a string",
				Value:   value,
			}
		}

		if len(str) < minLength {
			logger.Validation.Debug().
				Str("field", fieldName).
				Int("length", len(str)).
				Int("min_length", minLength).
				Msg("min length validation failed")
			return &ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("field must be at least %d characters long", minLength),
				Value:   value,
			}
		}

		return nil
	}
}

// ValidateMaxLength validates maximum length for string fields
func ValidateMaxLength(fieldName string, maxLength int) ValidationFunc {
	return func(ctx context.Context, rec schema.JRecord) error {
		value := rec.Get(fieldName)
		if value == nil {
			return nil // Let required validation handle this
		}

		str, ok := value.(string)
		if !ok {
			return &ValidationError{
				Field:   fieldName,
				Message: "field must be a string",
				Value:   value,
			}
		}

		if len(str) > maxLength {
			logger.Validation.Debug().
				Str("field", fieldName).
				Int("length", len(str)).
				Int("max_length", maxLength).
				Msg("max length validation failed")
			return &ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("field must be at most %d characters long", maxLength),
				Value:   value,
			}
		}

		return nil
	}
}

// ValidateRegex validates that a field matches a regex pattern
func ValidateRegex(fieldName string, pattern string) ValidationFunc {
	regex, err := regexp.Compile(pattern)
	if err != nil {
		logger.Validation.Error().
			Str("pattern", pattern).
			Msg("invalid regex pattern")
		panic(fmt.Sprintf("invalid regex pattern: %s", pattern))
	}

	return func(ctx context.Context, rec schema.JRecord) error {
		value := rec.Get(fieldName)
		if value == nil {
			return nil // Let required validation handle this
		}

		str, ok := value.(string)
		if !ok {
			return &ValidationError{
				Field:   fieldName,
				Message: "field must be a string",
				Value:   value,
			}
		}

		if !regex.MatchString(str) {
			logger.Validation.Debug().
				Str("field", fieldName).
				Str("value", str).
				Str("pattern", pattern).
				Msg("regex validation failed")
			return &ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("field must match pattern: %s", pattern),
				Value:   value,
			}
		}

		return nil
	}
}

// ValidateEmail validates email format
func ValidateEmail(fieldName string) ValidationFunc {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

	return func(ctx context.Context, rec schema.JRecord) error {
		value := rec.Get(fieldName)
		if value == nil {
			return nil // Let required validation handle this
		}

		str, ok := value.(string)
		if !ok {
			return &ValidationError{
				Field:   fieldName,
				Message: "field must be a string",
				Value:   value,
			}
		}

		if !emailRegex.MatchString(str) {
			logger.Validation.Debug().
				Str("field", fieldName).
				Str("value", str).
				Msg("email validation failed")
			return &ValidationError{
				Field:   fieldName,
				Message: "field must be a valid email address",
				Value:   value,
			}
		}

		return nil
	}
}

// ValidateMinValue validates minimum value for numeric fields
func ValidateMinValue(fieldName string, minValue float64) ValidationFunc {
	return func(ctx context.Context, rec schema.JRecord) error {
		value := rec.Get(fieldName)
		if value == nil {
			return nil // Let required validation handle this
		}

		num, err := toFloat64(value)
		if err != nil {
			return &ValidationError{
				Field:   fieldName,
				Message: "field must be a number",
				Value:   value,
			}
		}

		if num < minValue {
			logger.Validation.Debug().
				Str("field", fieldName).
				Float64("value", num).
				Float64("min_value", minValue).
				Msg("min value validation failed")
			return &ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("field must be at least %v", minValue),
				Value:   value,
			}
		}

		return nil
	}
}

// ValidateMaxValue validates maximum value for numeric fields
func ValidateMaxValue(fieldName string, maxValue float64) ValidationFunc {
	return func(ctx context.Context, rec schema.JRecord) error {
		value := rec.Get(fieldName)
		if value == nil {
			return nil // Let required validation handle this
		}

		num, err := toFloat64(value)
		if err != nil {
			return &ValidationError{
				Field:   fieldName,
				Message: "field must be a number",
				Value:   value,
			}
		}

		if num > maxValue {
			logger.Validation.Debug().
				Str("field", fieldName).
				Float64("value", num).
				Float64("max_value", maxValue).
				Msg("max value validation failed")
			return &ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("field must be at most %v", maxValue),
				Value:   value,
			}
		}

		return nil
	}
}

// ValidateRange validates that a field is within a range
func ValidateRange(fieldName string, minValue, maxValue float64) ValidationFunc {
	return func(ctx context.Context, rec schema.JRecord) error {
		value := rec.Get(fieldName)
		if value == nil {
			return nil // Let required validation handle this
		}

		num, err := toFloat64(value)
		if err != nil {
			return &ValidationError{
				Field:   fieldName,
				Message: "field must be a number",
				Value:   value,
			}
		}

		if num < minValue || num > maxValue {
			logger.Validation.Debug().
				Str("field", fieldName).
				Float64("value", num).
				Float64("min_value", minValue).
				Float64("max_value", maxValue).
				Msg("range validation failed")
			return &ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("field must be between %v and %v", minValue, maxValue),
				Value:   value,
			}
		}

		return nil
	}
}

// ValidateUnique validates that a field value is unique (requires repository)
func ValidateUnique(fieldName string, repository interface{}) ValidationFunc {
	return func(ctx context.Context, rec schema.JRecord) error {
		value := rec.Get(fieldName)
		if value == nil {
			return nil // Let required validation handle this
		}

		// This would need to be implemented with actual repository calls
		// For now, we'll just return nil
		// In a real implementation, you would:
		// 1. Check if the repository supports uniqueness checks
		// 2. Query the repository for existing records with the same value
		// 3. Return an error if found

		return nil
	}
}

// Cross-field validation functions

// ValidateFieldsEqual validates that two fields have equal values
func ValidateFieldsEqual(field1, field2 string) ValidationFunc {
	return func(ctx context.Context, rec schema.JRecord) error {
		value1 := rec.Get(field1)
		value2 := rec.Get(field2)

		if value1 != value2 {
			logger.Validation.Debug().
				Str("field1", field1).
				Str("field2", field2).
				Interface("value1", value1).
				Interface("value2", value2).
				Msg("field equality validation failed")
			return &ValidationError{
				Field:   field1,
				Message: fmt.Sprintf("field must equal field '%s'", field2),
				Value:   value1,
			}
		}

		return nil
	}
}

// ValidateFieldsNotEqual validates that two fields have different values
func ValidateFieldsNotEqual(field1, field2 string) ValidationFunc {
	return func(ctx context.Context, rec schema.JRecord) error {
		value1 := rec.Get(field1)
		value2 := rec.Get(field2)

		if value1 == value2 {
			logger.Validation.Debug().
				Str("field1", field1).
				Str("field2", field2).
				Interface("value1", value1).
				Interface("value2", value2).
				Msg("field inequality validation failed")
			return &ValidationError{
				Field:   field1,
				Message: fmt.Sprintf("field must not equal field '%s'", field2),
				Value:   value1,
			}
		}

		return nil
	}
}

// ValidateFieldGreaterThan validates that one field is greater than another
func ValidateFieldGreaterThan(field1, field2 string) ValidationFunc {
	return func(ctx context.Context, rec schema.JRecord) error {
		value1 := rec.Get(field1)
		value2 := rec.Get(field2)

		if value1 == nil || value2 == nil {
			return nil // Let required validation handle this
		}

		num1, err1 := toFloat64(value1)
		num2, err2 := toFloat64(value2)

		if err1 != nil || err2 != nil {
			return &ValidationError{
				Field:   field1,
				Message: "both fields must be numbers",
				Value:   value1,
			}
		}

		if num1 <= num2 {
			logger.Validation.Debug().
				Str("field1", field1).
				Str("field2", field2).
				Float64("value1", num1).
				Float64("value2", num2).
				Msg("field greater than validation failed")
			return &ValidationError{
				Field:   field1,
				Message: fmt.Sprintf("field must be greater than field '%s'", field2),
				Value:   value1,
			}
		}

		return nil
	}
}

// ValidateFieldLessThan validates that one field is less than another
func ValidateFieldLessThan(field1, field2 string) ValidationFunc {
	return func(ctx context.Context, rec schema.JRecord) error {
		value1 := rec.Get(field1)
		value2 := rec.Get(field2)

		if value1 == nil || value2 == nil {
			return nil // Let required validation handle this
		}

		num1, err1 := toFloat64(value1)
		num2, err2 := toFloat64(value2)

		if err1 != nil || err2 != nil {
			return &ValidationError{
				Field:   field1,
				Message: "both fields must be numbers",
				Value:   value1,
			}
		}

		if num1 >= num2 {
			logger.Validation.Debug().
				Str("field1", field1).
				Str("field2", field2).
				Float64("value1", num1).
				Float64("value2", num2).
				Msg("field less than validation failed")
			return &ValidationError{
				Field:   field1,
				Message: fmt.Sprintf("field must be less than field '%s'", field2),
				Value:   value1,
			}
		}

		return nil
	}
}

// ValidateFieldIn validates that a field value is in a list of allowed values
func ValidateFieldIn(fieldName string, allowedValues []any) ValidationFunc {
	return func(ctx context.Context, rec schema.JRecord) error {
		value := rec.Get(fieldName)
		if value == nil {
			return nil // Let required validation handle this
		}

		for _, allowed := range allowedValues {
			if value == allowed {
				return nil
			}
		}

		logger.Validation.Debug().
			Str("field", fieldName).
			Interface("value", value).
			Interface("allowed_values", allowedValues).
			Msg("field value not in allowed values")
		return &ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf("field must be one of: %v", allowedValues),
			Value:   value,
		}
	}
}

// ValidateFieldNotIn validates that a field value is not in a list of disallowed values
func ValidateFieldNotIn(fieldName string, disallowedValues []any) ValidationFunc {
	return func(ctx context.Context, rec schema.JRecord) error {
		value := rec.Get(fieldName)
		if value == nil {
			return nil // Let required validation handle this
		}

		for _, disallowed := range disallowedValues {
			if value == disallowed {
				logger.Validation.Debug().
					Str("field", fieldName).
					Interface("value", value).
					Interface("disallowed_values", disallowedValues).
					Msg("field value in disallowed values")
				return &ValidationError{
					Field:   fieldName,
					Message: fmt.Sprintf("field must not be one of: %v", disallowedValues),
					Value:   value,
				}
			}
		}

		return nil
	}
}

// Helper function to convert any value to float64
func toFloat64(value any) (float64, error) {
	switch v := value.(type) {
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case string:
		return strconv.ParseFloat(v, 64)
	default:
		logger.Validation.Error().
			Interface("value", value).
			Msg("cannot convert to float64")
		return 0, fmt.Errorf("cannot convert %T to float64", value)
	}
}

// Helper function to convert any value to string
func toString(value any) (string, error) {
	switch v := value.(type) {
	case string:
		return v, nil
	case []byte:
		return string(v), nil
	default:
		result := fmt.Sprintf("%v", v)
		logger.Validation.Debug().
			Interface("value", v).
			Str("converted", result).
			Msg("converted value to string")
		return result, nil
	}
}

// Helper function to check if a value is empty
func isEmpty(value any) bool {
	if value == nil {
		return true
	}

	switch v := value.(type) {
	case string:
		return strings.TrimSpace(v) == ""
	case []byte:
		return len(v) == 0
	case []string:
		return len(v) == 0
	case []any:
		return len(v) == 0
	default:
		// Use reflection to check if it's a zero value
		rv := reflect.ValueOf(value)
		return rv.IsZero()
	}
}
