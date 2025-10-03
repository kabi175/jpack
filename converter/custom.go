package converter

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/kabi175/jpack/logger"
	"github.com/kabi175/jpack/schema"
)

// CustomConverter allows creating converters for custom types
type CustomConverter struct {
	fieldType  schema.JFieldType
	toDBFunc   func(any) (any, error)
	fromDBFunc func(any) (any, error)
}

// NewCustomConverter creates a new custom converter
func NewCustomConverter(fieldType schema.JFieldType, toDBFunc, fromDBFunc func(any) (any, error)) JConverter {
	return &CustomConverter{
		fieldType:  fieldType,
		toDBFunc:   toDBFunc,
		fromDBFunc: fromDBFunc,
	}
}

func (c *CustomConverter) ToDB(value any) (any, error) {
	return c.toDBFunc(value)
}

func (c *CustomConverter) FromDB(raw any) (any, error) {
	return c.fromDBFunc(raw)
}

func (c *CustomConverter) Type() schema.JFieldType {
	return c.fieldType
}

// EnumConverter converts between enum types and strings
type EnumConverter struct {
	fieldType schema.JFieldType
	enumMap   map[string]any
	valueMap  map[any]string
}

// NewEnumConverter creates a new enum converter
func NewEnumConverter(fieldType schema.JFieldType, enumValues map[string]any) JConverter {
	valueMap := make(map[any]string)
	for k, v := range enumValues {
		valueMap[v] = k
	}

	return &EnumConverter{
		fieldType: fieldType,
		enumMap:   enumValues,
		valueMap:  valueMap,
	}
}

func (c *EnumConverter) ToDB(value any) (any, error) {
	if value == nil {
		return nil, nil
	}

	// If it's already a string, check if it's a valid enum key
	if str, ok := value.(string); ok {
		if _, exists := c.enumMap[str]; exists {
			return str, nil
		}
		logger.Converter.Error().
			Str("enum_value", str).
			Msg("invalid enum value")
		return nil, fmt.Errorf("invalid enum value: %s", str)
	}

	// Try to find the string representation
	if str, exists := c.valueMap[value]; exists {
		return str, nil
	}

	logger.Converter.Error().
		Interface("value", value).
		Msg("cannot convert to enum string")
	return nil, fmt.Errorf("cannot convert %T to enum string", value)
}

func (c *EnumConverter) FromDB(raw any) (any, error) {
	if raw == nil {
		return nil, nil
	}

	if str, ok := raw.(string); ok {
		if value, exists := c.enumMap[str]; exists {
			return value, nil
		}
		logger.Converter.Error().
			Str("enum_string", str).
			Msg("invalid enum string")
		return nil, fmt.Errorf("invalid enum string: %s", str)
	}

	logger.Converter.Error().
		Interface("raw", raw).
		Msg("cannot convert from enum string")
	return nil, fmt.Errorf("cannot convert %T from enum string", raw)
}

func (c *EnumConverter) Type() schema.JFieldType {
	return c.fieldType
}

// MoneyConverter converts between Money struct and database representation
type Money struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

type MoneyConverter struct{}

func NewMoneyConverter() JConverter {
	return &MoneyConverter{}
}

func (c *MoneyConverter) ToDB(value any) (any, error) {
	if value == nil {
		return nil, nil
	}

	switch v := value.(type) {
	case Money:
		return map[string]any{
			"amount":   v.Amount,
			"currency": v.Currency,
		}, nil
	case map[string]any:
		return v, nil
	default:
		// Try to convert using JSON
		jsonData, err := json.Marshal(value)
		if err != nil {
			logger.Converter.Error().
				Interface("money", value).
				Err(err).
				Msg("failed to marshal Money to JSON")
			return nil, fmt.Errorf("cannot marshal Money to JSON: %w", err)
		}

		var result map[string]any
		if err := json.Unmarshal(jsonData, &result); err != nil {
			logger.Converter.Error().
				Bytes("json_data", jsonData).
				Err(err).
				Msg("failed to unmarshal JSON to Money")
			return nil, fmt.Errorf("cannot unmarshal JSON to Money: %w", err)
		}

		return result, nil
	}
}

func (c *MoneyConverter) FromDB(raw any) (any, error) {
	if raw == nil {
		return Money{}, nil
	}

	switch v := raw.(type) {
	case map[string]any:
		money := Money{}
		if amount, ok := v["amount"]; ok {
			if f, ok := amount.(float64); ok {
				money.Amount = f
			}
		}
		if currency, ok := v["currency"]; ok {
			if s, ok := currency.(string); ok {
				money.Currency = s
			}
		}
		return money, nil
	default:
		// Try to convert using JSON
		jsonData, err := json.Marshal(raw)
		if err != nil {
			return nil, fmt.Errorf("cannot marshal raw data to JSON: %w", err)
		}

		var result Money
		if err := json.Unmarshal(jsonData, &result); err != nil {
			logger.Converter.Error().
				Bytes("json_data", jsonData).
				Err(err).
				Msg("failed to unmarshal JSON to Money")
			return nil, fmt.Errorf("cannot unmarshal JSON to Money: %w", err)
		}

		return result, nil
	}
}

func (c *MoneyConverter) Type() schema.JFieldType {
	return "money" // Custom type
}

// EmailConverter converts between Email struct and database representation
type Email struct {
	Address  string `json:"address"`
	Verified bool   `json:"verified"`
}

type EmailConverter struct{}

func NewEmailConverter() JConverter {
	return &EmailConverter{}
}

func (c *EmailConverter) ToDB(value any) (any, error) {
	if value == nil {
		return nil, nil
	}

	switch v := value.(type) {
	case Email:
		return map[string]any{
			"address":  v.Address,
			"verified": v.Verified,
		}, nil
	case string:
		// Simple string email
		return map[string]any{
			"address":  v,
			"verified": false,
		}, nil
	case map[string]any:
		return v, nil
	default:
		// Try to convert using JSON
		jsonData, err := json.Marshal(value)
		if err != nil {
			logger.Converter.Error().
				Interface("email", value).
				Err(err).
				Msg("failed to marshal Email to JSON")
			return nil, fmt.Errorf("cannot marshal Email to JSON: %w", err)
		}

		var result map[string]any
		if err := json.Unmarshal(jsonData, &result); err != nil {
			logger.Converter.Error().
				Bytes("json_data", jsonData).
				Err(err).
				Msg("failed to unmarshal JSON to Email")
			return nil, fmt.Errorf("cannot unmarshal JSON to Email: %w", err)
		}

		return result, nil
	}
}

func (c *EmailConverter) FromDB(raw any) (any, error) {
	if raw == nil {
		return Email{}, nil
	}

	switch v := raw.(type) {
	case map[string]any:
		email := Email{}
		if address, ok := v["address"]; ok {
			if s, ok := address.(string); ok {
				email.Address = s
			}
		}
		if verified, ok := v["verified"]; ok {
			if b, ok := verified.(bool); ok {
				email.Verified = b
			}
		}
		return email, nil
	case string:
		// Simple string email
		return Email{
			Address:  v,
			Verified: false,
		}, nil
	default:
		// Try to convert using JSON
		jsonData, err := json.Marshal(raw)
		if err != nil {
			return nil, fmt.Errorf("cannot marshal raw data to JSON: %w", err)
		}

		var result Email
		if err := json.Unmarshal(jsonData, &result); err != nil {
			logger.Converter.Error().
				Bytes("json_data", jsonData).
				Err(err).
				Msg("failed to unmarshal JSON to Email")
			return nil, fmt.Errorf("cannot unmarshal JSON to Email: %w", err)
		}

		return result, nil
	}
}

func (c *EmailConverter) Type() schema.JFieldType {
	return "email" // Custom type
}

// JSONConverter converts between any Go type and JSON representation
type JSONConverter struct {
	fieldType  schema.JFieldType
	targetType reflect.Type
}

// NewJSONConverter creates a new JSON converter for a specific type
func NewJSONConverter(fieldType schema.JFieldType, targetType reflect.Type) JConverter {
	return &JSONConverter{
		fieldType:  fieldType,
		targetType: targetType,
	}
}

func (c *JSONConverter) ToDB(value any) (any, error) {
	if value == nil {
		return nil, nil
	}

	// If it's already a map or slice, return as is
	if reflect.TypeOf(value).Kind() == reflect.Map || reflect.TypeOf(value).Kind() == reflect.Slice {
		return value, nil
	}

	// Convert to JSON
	jsonData, err := json.Marshal(value)
	if err != nil {
		logger.Converter.Error().
			Interface("value", value).
			Err(err).
			Msg("failed to marshal to JSON")
		return nil, fmt.Errorf("cannot marshal to JSON: %w", err)
	}

	// Unmarshal to generic interface{}
	var result any
	if err := json.Unmarshal(jsonData, &result); err != nil {
		logger.Converter.Error().
			Bytes("json_data", jsonData).
			Err(err).
			Msg("failed to unmarshal JSON")
		return nil, fmt.Errorf("cannot unmarshal JSON: %w", err)
	}

	return result, nil
}

func (c *JSONConverter) FromDB(raw any) (any, error) {
	if raw == nil {
		return nil, nil
	}

	// If it's already the target type, return as is
	if c.targetType != nil && reflect.TypeOf(raw).AssignableTo(c.targetType) {
		return raw, nil
	}

	// Convert to JSON
	jsonData, err := json.Marshal(raw)
	if err != nil {
		logger.Converter.Error().
			Interface("value", raw).
			Err(err).
			Msg("failed to marshal to JSON")
		return nil, fmt.Errorf("cannot marshal to JSON: %w", err)
	}

	// Create new instance of target type
	if c.targetType != nil {
		target := reflect.New(c.targetType).Interface()
		if err := json.Unmarshal(jsonData, target); err != nil {
			logger.Converter.Error().
				Bytes("json_data", jsonData).
				Err(err).
				Msg("failed to unmarshal JSON to target type")
			return nil, fmt.Errorf("cannot unmarshal JSON to target type: %w", err)
		}
		return reflect.ValueOf(target).Elem().Interface(), nil
	}

	// Return as generic interface{}
	var result any
	if err := json.Unmarshal(jsonData, &result); err != nil {
		logger.Converter.Error().
			Bytes("json_data", jsonData).
			Err(err).
			Msg("failed to unmarshal JSON")
		return nil, fmt.Errorf("cannot unmarshal JSON: %w", err)
	}

	return result, nil
}

func (c *JSONConverter) Type() schema.JFieldType {
	return c.fieldType
}

// StringSliceConverter converts between []string and database representation
type StringSliceConverter struct{}

func NewStringSliceConverter() JConverter {
	return &StringSliceConverter{}
}

func (c *StringSliceConverter) ToDB(value any) (any, error) {
	if value == nil {
		return nil, nil
	}

	switch v := value.(type) {
	case []string:
		return v, nil
	case []any:
		result := make([]string, len(v))
		for i, item := range v {
			if str, ok := item.(string); ok {
				result[i] = str
			} else {
				result[i] = fmt.Sprintf("%v", item)
			}
		}
		return result, nil
	case string:
		// Split by comma
		return strings.Split(v, ","), nil
	default:
		logger.Converter.Error().
			Interface("value", value).
			Str("type", "[]string").
			Msg("conversion failed")
		return nil, fmt.Errorf("cannot convert %T to []string", value)
	}
}

func (c *StringSliceConverter) FromDB(raw any) (any, error) {
	if raw == nil {
		return []string{}, nil
	}

	switch v := raw.(type) {
	case []string:
		return v, nil
	case []any:
		result := make([]string, len(v))
		for i, item := range v {
			if str, ok := item.(string); ok {
				result[i] = str
			} else {
				result[i] = fmt.Sprintf("%v", item)
			}
		}
		return result, nil
	case string:
		// Split by comma
		return strings.Split(v, ","), nil
	default:
		logger.Converter.Error().
			Interface("raw", raw).
			Str("type", "[]string").
			Msg("conversion from database failed")
		return nil, fmt.Errorf("cannot convert %T to []string", raw)
	}
}

func (c *StringSliceConverter) Type() schema.JFieldType {
	return "string_slice" // Custom type
}
