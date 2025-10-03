package converter

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/kabi175/jpack/logger"
	"github.com/kabi175/jpack/schema"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// stringConverter implements JConverter for string type
type stringConverter struct{}

func NewStringConverter() JConverter {
	return &stringConverter{}
}

func (c *stringConverter) ToDB(value any) (any, error) {
	if value == nil {
		return nil, nil
	}

	logger.Converter.Debug().
		Interface("value", value).
		Str("type", "string").
		Msg("converting to database format")

	switch v := value.(type) {
	case string:
		return v, nil
	case []byte:
		return string(v), nil
	default:
		result := fmt.Sprintf("%v", v)
		logger.Converter.Debug().
			Interface("original", value).
			Str("converted", result).
			Msg("converted to string using fmt.Sprintf")
		return result, nil
	}
}

func (c *stringConverter) FromDB(raw any) (any, error) {
	if raw == nil {
		return "", nil
	}

	logger.Converter.Debug().
		Interface("raw", raw).
		Str("type", "string").
		Msg("converting from database format")

	switch v := raw.(type) {
	case string:
		return v, nil
	case []byte:
		return string(v), nil
	default:
		result := fmt.Sprintf("%v", v)
		logger.Converter.Debug().
			Interface("original", raw).
			Str("converted", result).
			Msg("converted from database using fmt.Sprintf")
		return result, nil
	}
}

func (c *stringConverter) Type() schema.JFieldType {
	return schema.JString
}

// intConverter implements JConverter for int type
type intConverter struct{}

func NewIntConverter() JConverter {
	return &intConverter{}
}

func (c *intConverter) ToDB(value any) (any, error) {
	if value == nil {
		return nil, nil
	}

	switch v := value.(type) {
	case int:
		return v, nil
	case int32:
		return int(v), nil
	case int64:
		return int(v), nil
	case float32:
		return int(v), nil
	case float64:
		return int(v), nil
	default:
		logger.Converter.Error().
			Interface("value", value).
			Str("type", "int").
			Msg("conversion failed")
		return nil, fmt.Errorf("cannot convert %T to int", value)
	}
}

func (c *intConverter) FromDB(raw any) (any, error) {
	if raw == nil {
		return 0, nil
	}

	switch v := raw.(type) {
	case int:
		return v, nil
	case int32:
		return int(v), nil
	case int64:
		return int(v), nil
	case float32:
		return int(v), nil
	case float64:
		return int(v), nil
	default:
		logger.Converter.Error().
			Interface("raw", raw).
			Str("type", "int").
			Msg("conversion from database failed")
		return nil, fmt.Errorf("cannot convert %T to int", raw)
	}
}

func (c *intConverter) Type() schema.JFieldType {
	return schema.JInt
}

// int64Converter implements JConverter for int64 type
type int64Converter struct{}

func NewInt64Converter() JConverter {
	return &int64Converter{}
}

func (c *int64Converter) ToDB(value any) (any, error) {
	if value == nil {
		return nil, nil
	}

	switch v := value.(type) {
	case int64:
		return v, nil
	case int:
		return int64(v), nil
	case int32:
		return int64(v), nil
	case float32:
		return int64(v), nil
	case float64:
		return int64(v), nil
	default:
		return nil, fmt.Errorf("cannot convert %T to int64", value)
	}
}

func (c *int64Converter) FromDB(raw any) (any, error) {
	if raw == nil {
		return int64(0), nil
	}

	switch v := raw.(type) {
	case int64:
		return v, nil
	case int:
		return int64(v), nil
	case int32:
		return int64(v), nil
	case float32:
		return int64(v), nil
	case float64:
		return int64(v), nil
	default:
		return nil, fmt.Errorf("cannot convert %T to int64", raw)
	}
}

func (c *int64Converter) Type() schema.JFieldType {
	return schema.JInt64
}

// floatConverter implements JConverter for float32 type
type floatConverter struct{}

func NewFloatConverter() JConverter {
	return &floatConverter{}
}

func (c *floatConverter) ToDB(value any) (any, error) {
	if value == nil {
		return nil, nil
	}

	switch v := value.(type) {
	case float32:
		return v, nil
	case float64:
		return float32(v), nil
	case int:
		return float32(v), nil
	case int32:
		return float32(v), nil
	case int64:
		return float32(v), nil
	default:
		return nil, fmt.Errorf("cannot convert %T to float32", value)
	}
}

func (c *floatConverter) FromDB(raw any) (any, error) {
	if raw == nil {
		return float32(0), nil
	}

	switch v := raw.(type) {
	case float32:
		return v, nil
	case float64:
		return float32(v), nil
	case int:
		return float32(v), nil
	case int32:
		return float32(v), nil
	case int64:
		return float32(v), nil
	default:
		return nil, fmt.Errorf("cannot convert %T to float32", raw)
	}
}

func (c *floatConverter) Type() schema.JFieldType {
	return schema.JFloat
}

// float64Converter implements JConverter for float64 type
type float64Converter struct{}

func NewFloat64Converter() JConverter {
	return &float64Converter{}
}

func (c *float64Converter) ToDB(value any) (any, error) {
	if value == nil {
		return nil, nil
	}

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
	default:
		return nil, fmt.Errorf("cannot convert %T to float64", value)
	}
}

func (c *float64Converter) FromDB(raw any) (any, error) {
	if raw == nil {
		return float64(0), nil
	}

	switch v := raw.(type) {
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
	default:
		return nil, fmt.Errorf("cannot convert %T to float64", raw)
	}
}

func (c *float64Converter) Type() schema.JFieldType {
	return schema.JFloat64
}

// boolConverter implements JConverter for bool type
type boolConverter struct{}

func NewBoolConverter() JConverter {
	return &boolConverter{}
}

func (c *boolConverter) ToDB(value any) (any, error) {
	if value == nil {
		return nil, nil
	}

	switch v := value.(type) {
	case bool:
		return v, nil
	case string:
		return v == "true" || v == "1" || v == "yes", nil
	case int:
		return v != 0, nil
	case int32:
		return v != 0, nil
	case int64:
		return v != 0, nil
	case float32:
		return v != 0, nil
	case float64:
		return v != 0, nil
	default:
		return nil, fmt.Errorf("cannot convert %T to bool", value)
	}
}

func (c *boolConverter) FromDB(raw any) (any, error) {
	if raw == nil {
		return false, nil
	}

	switch v := raw.(type) {
	case bool:
		return v, nil
	case string:
		return v == "true" || v == "1" || v == "yes", nil
	case int:
		return v != 0, nil
	case int32:
		return v != 0, nil
	case int64:
		return v != 0, nil
	case float32:
		return v != 0, nil
	case float64:
		return v != 0, nil
	default:
		return nil, fmt.Errorf("cannot convert %T to bool", raw)
	}
}

func (c *boolConverter) Type() schema.JFieldType {
	return schema.JBool
}

// timeConverter implements JConverter for time.Time type
type timeConverter struct{}

func NewTimeConverter() JConverter {
	return &timeConverter{}
}

func (c *timeConverter) ToDB(value any) (any, error) {
	if value == nil {
		return nil, nil
	}

	switch v := value.(type) {
	case time.Time:
		return v, nil
	case string:
		// Try to parse common time formats
		formats := []string{
			time.RFC3339,
			time.RFC3339Nano,
			"2006-01-02T15:04:05Z07:00",
			"2006-01-02 15:04:05",
			"2006-01-02",
		}

		for _, format := range formats {
			if t, err := time.Parse(format, v); err == nil {
				return t, nil
			}
		}
		logger.Converter.Error().
			Str("time_string", v).
			Msg("failed to parse time string")
		return nil, fmt.Errorf("cannot parse time string: %s", v)
	case int64:
		// Unix timestamp
		return time.Unix(v, 0), nil
	default:
		logger.Converter.Error().
			Interface("value", value).
			Str("type", "time.Time").
			Msg("conversion failed")
		return nil, fmt.Errorf("cannot convert %T to time.Time", value)
	}
}

func (c *timeConverter) FromDB(raw any) (any, error) {
	if raw == nil {
		return time.Time{}, nil
	}

	switch v := raw.(type) {
	case time.Time:
		return v, nil
	case primitive.DateTime:
		return v.Time(), nil
	case string:
		// Try to parse common time formats
		formats := []string{
			time.RFC3339,
			time.RFC3339Nano,
			"2006-01-02T15:04:05Z07:00",
			"2006-01-02 15:04:05",
			"2006-01-02",
		}

		for _, format := range formats {
			if t, err := time.Parse(format, v); err == nil {
				return t, nil
			}
		}
		logger.Converter.Error().
			Str("time_string", v).
			Msg("failed to parse time string")
		return nil, fmt.Errorf("cannot parse time string: %s", v)
	case int64:
		// Unix timestamp
		return time.Unix(v, 0), nil
	default:
		logger.Converter.Error().
			Interface("raw", raw).
			Str("type", "time.Time").
			Msg("conversion from database failed")
		return nil, fmt.Errorf("cannot convert %T to time.Time", raw)
	}
}

func (c *timeConverter) Type() schema.JFieldType {
	return schema.JTime
}

// objectConverter implements JConverter for object type (structs, maps)
type objectConverter struct{}

func NewObjectConverter() JConverter {
	return &objectConverter{}
}

func (c *objectConverter) ToDB(value any) (any, error) {
	if value == nil {
		return nil, nil
	}

	// If it's already a map, return as is
	if _, ok := value.(map[string]any); ok {
		return value, nil
	}

	// Convert struct to map using JSON marshaling
	jsonData, err := json.Marshal(value)
	if err != nil {
		logger.Converter.Error().
			Interface("value", value).
			Err(err).
			Msg("failed to marshal object to JSON")
		return nil, fmt.Errorf("cannot marshal object to JSON: %w", err)
	}

	var result map[string]any
	if err := json.Unmarshal(jsonData, &result); err != nil {
		logger.Converter.Error().
			Bytes("json_data", jsonData).
			Err(err).
			Msg("failed to unmarshal JSON to map")
		return nil, fmt.Errorf("cannot unmarshal JSON to map: %w", err)
	}

	return result, nil
}

func (c *objectConverter) FromDB(raw any) (any, error) {
	if raw == nil {
		return map[string]any{}, nil
	}

	// If it's already a map, return as is
	if result, ok := raw.(map[string]any); ok {
		return result, nil
	}

	// Convert to map using JSON marshaling
	jsonData, err := json.Marshal(raw)
	if err != nil {
		logger.Converter.Error().
			Interface("raw", raw).
			Err(err).
			Msg("failed to marshal raw data to JSON")
		return nil, fmt.Errorf("cannot marshal raw data to JSON: %w", err)
	}

	var result map[string]any
	if err := json.Unmarshal(jsonData, &result); err != nil {
		logger.Converter.Error().
			Bytes("json_data", jsonData).
			Err(err).
			Msg("failed to unmarshal JSON to map")
		return nil, fmt.Errorf("cannot unmarshal JSON to map: %w", err)
	}

	return result, nil
}

func (c *objectConverter) Type() schema.JFieldType {
	return schema.JObject
}

// arrayConverter implements JConverter for array type
type arrayConverter struct{}

func NewArrayConverter() JConverter {
	return &arrayConverter{}
}

func (c *arrayConverter) ToDB(value any) (any, error) {
	if value == nil {
		return nil, nil
	}

	// If it's already a slice, return as is
	if reflect.TypeOf(value).Kind() == reflect.Slice {
		return value, nil
	}

	// Convert to slice using JSON marshaling
	jsonData, err := json.Marshal(value)
	if err != nil {
		logger.Converter.Error().
			Interface("value", value).
			Err(err).
			Msg("failed to marshal array to JSON")
		return nil, fmt.Errorf("cannot marshal array to JSON: %w", err)
	}

	var result []any
	if err := json.Unmarshal(jsonData, &result); err != nil {
		logger.Converter.Error().
			Bytes("json_data", jsonData).
			Err(err).
			Msg("failed to unmarshal JSON to slice")
		return nil, fmt.Errorf("cannot unmarshal JSON to slice: %w", err)
	}

	return result, nil
}

func (c *arrayConverter) FromDB(raw any) (any, error) {
	if raw == nil {
		return []any{}, nil
	}

	// If it's already a slice, return as is
	if reflect.TypeOf(raw).Kind() == reflect.Slice {
		return raw, nil
	}

	// Convert to slice using JSON marshaling
	jsonData, err := json.Marshal(raw)
	if err != nil {
		logger.Converter.Error().
			Interface("raw", raw).
			Err(err).
			Msg("failed to marshal raw data to JSON")
		return nil, fmt.Errorf("cannot marshal raw data to JSON: %w", err)
	}

	var result []any
	if err := json.Unmarshal(jsonData, &result); err != nil {
		logger.Converter.Error().
			Bytes("json_data", jsonData).
			Err(err).
			Msg("failed to unmarshal JSON to slice")
		return nil, fmt.Errorf("cannot unmarshal JSON to slice: %w", err)
	}

	return result, nil
}

func (c *arrayConverter) Type() schema.JFieldType {
	return schema.JArray
}

// binaryConverter implements JConverter for binary type
type binaryConverter struct{}

func NewBinaryConverter() JConverter {
	return &binaryConverter{}
}

func (c *binaryConverter) ToDB(value any) (any, error) {
	if value == nil {
		return nil, nil
	}

	switch v := value.(type) {
	case []byte:
		return v, nil
	case string:
		return []byte(v), nil
	default:
		logger.Converter.Error().
			Interface("value", value).
			Str("type", "[]byte").
			Msg("conversion failed")
		return nil, fmt.Errorf("cannot convert %T to []byte", value)
	}
}

func (c *binaryConverter) FromDB(raw any) (any, error) {
	if raw == nil {
		return []byte{}, nil
	}

	switch v := raw.(type) {
	case []byte:
		return v, nil
	case string:
		return []byte(v), nil
	default:
		logger.Converter.Error().
			Interface("raw", raw).
			Str("type", "[]byte").
			Msg("conversion from database failed")
		return nil, fmt.Errorf("cannot convert %T to []byte", raw)
	}
}

func (c *binaryConverter) Type() schema.JFieldType {
	return schema.JBinary
}

// objectIDConverter implements JConverter for ObjectID type
type objectIDConverter struct{}

func NewObjectIDConverter() JConverter {
	return &objectIDConverter{}
}

func (c *objectIDConverter) ToDB(value any) (any, error) {
	if value == nil {
		return nil, nil
	}

	switch v := value.(type) {
	case primitive.ObjectID:
		return v, nil
	case string:
		if objectID, err := primitive.ObjectIDFromHex(v); err == nil {
			return objectID, nil
		}
		logger.Converter.Error().
			Str("object_id_string", v).
			Msg("invalid ObjectID string")
		return nil, fmt.Errorf("invalid ObjectID string: %s", v)
	case []byte:
		if len(v) == 12 {
			return primitive.ObjectID(v), nil
		}
		logger.Converter.Error().
			Int("bytes_length", len(v)).
			Msg("invalid ObjectID bytes length")
		return nil, fmt.Errorf("invalid ObjectID bytes length: %d", len(v))
	default:
		logger.Converter.Error().
			Interface("value", value).
			Str("type", "ObjectID").
			Msg("conversion failed")
		return nil, fmt.Errorf("cannot convert %T to ObjectID", value)
	}
}

func (c *objectIDConverter) FromDB(raw any) (any, error) {
	if raw == nil {
		return primitive.NilObjectID, nil
	}

	switch v := raw.(type) {
	case primitive.ObjectID:
		return v, nil
	case string:
		if objectID, err := primitive.ObjectIDFromHex(v); err == nil {
			return objectID, nil
		}
		logger.Converter.Error().
			Str("object_id_string", v).
			Msg("invalid ObjectID string")
		return nil, fmt.Errorf("invalid ObjectID string: %s", v)
	case []byte:
		if len(v) == 12 {
			return primitive.ObjectID(v), nil
		}
		logger.Converter.Error().
			Int("bytes_length", len(v)).
			Msg("invalid ObjectID bytes length")
		return nil, fmt.Errorf("invalid ObjectID bytes length: %d", len(v))
	default:
		logger.Converter.Error().
			Interface("raw", raw).
			Str("type", "ObjectID").
			Msg("conversion from database failed")
		return nil, fmt.Errorf("cannot convert %T to ObjectID", raw)
	}
}

func (c *objectIDConverter) Type() schema.JFieldType {
	return schema.JObjectID
}
