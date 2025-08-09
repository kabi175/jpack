package jpack

import (
	"context"
	"errors"
	"math"
	"reflect"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Number struct{}

// Scan implements JFieldType.
func (n *Number) Scan(ctx context.Context, field JField, row map[string]any) (value any, err error) {
	v, ok := row[field.Name()]
	if !ok {
		return nil, nil // No value found, return nil
	}

	if v == nil {
		return nil, nil // If the value is nil, return nil
	}

	reflectValue := reflect.ValueOf(v)

	return convertToInt(reflectValue)
}

// Set implements JFieldType.
func (n *Number) SetValue(ctx context.Context, field JField, value any, row map[string]any) error {
	reflectValue := reflect.ValueOf(value)

	// If the value is nil, set the row field to nil
	// This is important for nullable fields
	// to ensure that the database can handle null values correctly.
	// If the value is not nil, convert it to int and set it in the row
	// This allows for flexibility in the types of values that can be set,

	// while still ensuring that the row field is always an integer.
	// This is particularly useful for fields that can accept different types
	// of numeric values, such as int, int8, int16, int32,
	// int64, and string representations of integers.
	// It also handles pointer types by dereferencing them to get the value.
	if value == nil || (reflectValue.Kind() == reflect.Pointer && reflectValue.IsNil()) {
		row[field.Name()] = nil // Set the field to nil if the value is nil
		return nil
	}

	columnName := field.Name()

	num, err := convertToInt(reflectValue)
	if err != nil {
		return err // Return error if conversion fails
	}

	row[columnName] = num
	return nil
}

// Validate implements JFieldType.
func (n *Number) Validate(value any) error {

	var validate func(reflect.Value) error

	validate = func(reflectValue reflect.Value) error {
		switch reflectValue.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return nil // No error for valid integer types
		case reflect.String:
			_, err := strconv.ParseInt(reflectValue.String(), 10, 32)
			if err != nil {
				return errors.Join(errors.New("value is not a valid integer"), err)
			}
			return nil // No error for valid integer types
		case reflect.Pointer:
			return validate(reflectValue.Elem())
		default:
			return errors.Join(errors.New("value is a struct, expected an integer"), nil)
		}
	}

	return validate(reflect.ValueOf(value))
}

func convertToInt(reflectValue reflect.Value) (int, error) {
	switch reflectValue.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		// Convert to int if it's a larger integer type
		if reflectValue.CanInt() {
			return int(reflectValue.Int()), nil
		} else {
			return 0, errors.New("value cannot be converted to integer")
		}
	case reflect.Float32, reflect.Float64:
		return int(math.Round(reflectValue.Float())), nil

	case reflect.String:
		// Attempt to parse the string as an integer
		num, err := strconv.ParseInt(reflectValue.String(), 10, 32)
		if err != nil {
			return 0, errors.New("value is not a valid integer string")
		}
		return int(num), nil

	case reflect.Pointer:
		if reflectValue.IsNil() {
			return 0, nil // If the pointer is nil, return 0
		}
		// Handle pointer types, dereferencing to get the value
		return convertToInt(reflectValue.Elem())

	}
	return 0, errors.New("value is not an integer type")
}

var _ JFieldType = &Number{}

type String struct{}

// Scan implements JFieldType.
func (s *String) Scan(ctx context.Context, field JField, row map[string]any) (value any, err error) {
	v, ok := row[field.Name()]
	if !ok {
		return nil, nil // No value found, return nil
	}

	if v == nil {
		return nil, nil // If the value is nil, return nil
	}

	reflectValue := reflect.ValueOf(v)

	if reflectValue.Kind() == reflect.String {
		return v.(string), nil
	}

	return "", errors.New("value is not a string")
}

// SetValue implements JFieldType.
func (s *String) SetValue(ctx context.Context, field JField, value any, row map[string]any) error {
	reflectValue := reflect.ValueOf(value)

	// If the value is nil, set the row field to nil
	if value == nil || (reflectValue.Kind() == reflect.Pointer && reflectValue.IsNil()) {
		row[field.Name()] = nil // Set the field to nil if the value is nil
		return nil
	}

	if reflectValue.Kind() != reflect.String {
		return errors.New("value is not a string")
	}

	row[field.Name()] = reflectValue.String()
	return nil
}

// Validate implements JFieldType.
func (s *String) Validate(value any) error {

	reflectValue := reflect.ValueOf(value)

	if reflectValue.Kind() == reflect.Pointer {
		if reflectValue.IsNil() {
			return nil // If the pointer is nil, return nil
		}
		// Handle pointer types, dereferencing to get the value
		reflectValue = reflectValue.Elem()
	}

	switch reflectValue.Kind() {
	case reflect.Struct, reflect.Array, reflect.Slice, reflect.Func, reflect.Chan, reflect.Map:
		return errors.New("value is a struct, expected a string")

	default:
		return nil // No error for valid string types
	}

}

var _ JFieldType = &String{}

type Ref struct{}

// Scan implements JFieldType.
func (r *Ref) Scan(ctx context.Context, field JField, row map[string]any) (value any, err error) {
	v, ok := row[field.Name()]
	if !ok {
		return nil, nil // No value found, return nil
	}

	if v == nil {
		return nil, nil // If the value is nil, return nil
	}

	ref, ok := field.(JRef)
	if !ok {
		return nil, errors.New("value is of type ref but field is not a ref")
	}

	objIDHex, ok := v.(string)
	if ok {
		rec := NewMongoRecord(ref.RelSchema())
		refPK, ok := PK(ref.RelSchema())
		if !ok {
			return nil, errors.New("no primary key found in referenced schema")
		}

		rec.SetValue(refPK, objIDHex)
		return objIDHex, nil // Return the hex representation of the ObjectID
	}

	return "", errors.New("value is not a object id")
}

// SetValue implements JFieldType.
func (r *Ref) SetValue(ctx context.Context, field JField, value any, row map[string]any) error {
	if err := r.Validate(value); err != nil {
		return err
	}

	reflectValue := reflect.ValueOf(value)

	// If the value is nil, set the row field to nil
	if value == nil || (reflectValue.Kind() == reflect.Pointer && reflectValue.IsNil()) {
		row[field.Name()] = nil // Set the field to nil if the value is nil
		return nil
	}

	if val, ok := tryCovertToString(reflectValue); ok {
		row[field.Name()] = val
		return nil
	}

	record, ok := value.(JRecord)
	if ok {
		pkField, ok := PK(record.Schema())
		if !ok {
			return errors.New("no primary key found in referenced schema")
		}

		if pkValue, ok := record.Value(pkField); ok {
			return r.SetValue(ctx, field, pkValue, row)
		}
	}

	return nil
}

func tryCovertToString(reflectValue reflect.Value) (string, bool) {

	if reflectValue.Kind() != reflect.String {
		return "", false
	}

	return reflectValue.String(), true
}

// Validate implements JFieldType.
func (r *Ref) Validate(value any) error {
	if value == nil {
		return nil // If the value is nil, return nil
	}

	_, ok := value.(JRecord)
	if ok {
		return nil // No error for valid JRecord types
	}

	reflectValue := reflect.ValueOf(value)
	if reflectValue.Kind() == reflect.String {
		if _, err := bson.ObjectIDFromHex(reflectValue.String()); err != nil {
			return errors.New("value is not a valid ObjectID hex string")
		}
		return nil
	}

	return errors.New("value is not a valid ref string or JRecord")

}

var _ JFieldType = &Ref{}

type DateTime struct{}

// Scan implements JFieldType.
func (dt *DateTime) Scan(ctx context.Context, field JField, row map[string]any) (value any, err error) {
	v, ok := row[field.Name()]
	if !ok {
		return nil, nil // No value found, return nil
	}

	if v == nil {
		return nil, nil // If the value is nil, return nil
	}

	reflectValue := reflect.ValueOf(v)

	switch reflectValue.Kind() {
	case reflect.String:
		// Parse RFC3339 format string and convert to GMT
		t, err := time.Parse(time.RFC3339, reflectValue.String())
		if err != nil {
			return nil, errors.New("value is not a valid RFC3339 datetime string")
		}
		// Convert to GMT timezone
		return t.UTC(), nil
	case reflect.Struct:
		// Check if it's a time.Time
		if t, ok := v.(time.Time); ok {
			// Convert to GMT timezone
			return t.UTC(), nil
		}
		return nil, errors.New("value is not a time.Time struct")
	default:
		return nil, errors.New("value is not a valid datetime type")
	}
}

// SetValue implements JFieldType.
func (dt *DateTime) SetValue(ctx context.Context, field JField, value any, row map[string]any) error {
	reflectValue := reflect.ValueOf(value)

	// If the value is nil, set the row field to nil
	if value == nil || (reflectValue.Kind() == reflect.Pointer && reflectValue.IsNil()) {
		row[field.Name()] = nil // Set the field to nil if the value is nil
		return nil
	}

	if err := dt.Validate(value); err != nil {
		return err
	}

	columnName := field.Name()

	switch v := value.(type) {
	case time.Time:
		// Store in GMT timezone
		row[columnName] = v.UTC()
	case string:
		// Parse RFC3339 format string and convert to GMT
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return errors.New("value is not a valid RFC3339 datetime string")
		}
		row[columnName] = t.UTC()
	default:
		return errors.New("value is not a valid datetime type")
	}

	return nil
}

// Validate implements JFieldType.
func (dt *DateTime) Validate(value any) error {
	if value == nil {
		return nil // If the value is nil, return nil
	}

	reflectValue := reflect.ValueOf(value)

	// Handle pointer types
	if reflectValue.Kind() == reflect.Pointer {
		if reflectValue.IsNil() {
			return nil // If the pointer is nil, return nil
		}
		// Handle pointer types, dereferencing to get the value
		reflectValue = reflectValue.Elem()
	}

	switch reflectValue.Kind() {
	case reflect.Struct:
		// Check if it's a time.Time
		if _, ok := reflectValue.Interface().(time.Time); ok {
			return nil // No error for valid time.Time types
		}
		return errors.New("value is a struct but not a time.Time")
	case reflect.String:
		// Validate RFC3339 format
		_, err := time.Parse(time.RFC3339, reflectValue.String())
		if err != nil {
			return errors.Join(errors.New("value is not a valid RFC3339 datetime string"), err)
		}
		return nil // No error for valid RFC3339 string types
	default:
		return errors.New("value is not a valid datetime type (expected time.Time or RFC3339 string)")
	}
}

var _ JFieldType = &DateTime{}

// Option represents a single option with unique name and display name
type Option struct {
	UniqueName  string `json:"uniqueName"`
	DisplayName string `json:"displayName"`
}

// OptionService defines the interface for getting available options
type OptionService interface {
	GetOptions(ctx context.Context) ([]Option, error)
}

// Options represents an enum field type that gets its allowed values from a service
type Options struct {
	service OptionService
}

// NewOptions creates a new Options FieldType with the given service
func NewOptions(service OptionService) *Options {
	return &Options{
		service: service,
	}
}

// Scan implements JFieldType.
func (o *Options) Scan(ctx context.Context, field JField, row map[string]any) (value any, err error) {
	v, ok := row[field.Name()]
	if !ok {
		return nil, nil // No value found, return nil
	}

	if v == nil {
		return nil, nil // If the value is nil, return nil
	}

	reflectValue := reflect.ValueOf(v)

	switch reflectValue.Kind() {
	case reflect.String:
		return reflectValue.String(), nil
	default:
		return nil, errors.New("options field must be a string")
	}
}

// SetValue implements JFieldType.
func (o *Options) SetValue(ctx context.Context, field JField, value any, row map[string]any) error {
	reflectValue := reflect.ValueOf(value)

	// If the value is nil, set the row field to nil
	if value == nil || (reflectValue.Kind() == reflect.Pointer && reflectValue.IsNil()) {
		row[field.Name()] = nil // Set the field to nil if the value is nil
		return nil
	}

	if err := o.Validate(value); err != nil {
		return err
	}

	columnName := field.Name()

	switch v := value.(type) {
	case string:
		row[columnName] = v
	default:
		return errors.New("options field must be a string")
	}

	return nil
}

// Validate implements JFieldType.
func (o *Options) Validate(value any) error {
	if value == nil {
		return nil // If the value is nil, return nil
	}

	reflectValue := reflect.ValueOf(value)

	// Handle pointer types
	if reflectValue.Kind() == reflect.Pointer {
		if reflectValue.IsNil() {
			return nil // If the pointer is nil, return nil
		}
		// Handle pointer types, dereferencing to get the value
		reflectValue = reflectValue.Elem()
	}

	if reflectValue.Kind() != reflect.String {
		return errors.New("options field must be a string")
	}

	// Get the string value (this should be the uniqueName)
	strValue := reflectValue.String()

	// Get available options from the service
	availableOptions, err := o.service.GetOptions(context.Background())
	if err != nil {
		return errors.Join(errors.New("failed to get available options"), err)
	}

	// Check if the value (uniqueName) is in the allowed options
	for _, option := range availableOptions {
		if option.UniqueName == strValue {
			return nil // Value is valid
		}
	}

	return errors.New("value is not in the list of available options")
}

// GetDisplayName returns the display name for a given unique name
func (o *Options) GetDisplayName(ctx context.Context, uniqueName string) (string, error) {
	availableOptions, err := o.service.GetOptions(ctx)
	if err != nil {
		return "", errors.Join(errors.New("failed to get available options"), err)
	}

	for _, option := range availableOptions {
		if option.UniqueName == uniqueName {
			return option.DisplayName, nil
		}
	}

	return "", errors.New("option not found")
}

// GetUniqueName returns the unique name for a given display name
func (o *Options) GetUniqueName(ctx context.Context, displayName string) (string, error) {
	availableOptions, err := o.service.GetOptions(ctx)
	if err != nil {
		return "", errors.Join(errors.New("failed to get available options"), err)
	}

	for _, option := range availableOptions {
		if option.DisplayName == displayName {
			return option.UniqueName, nil
		}
	}

	return "", errors.New("option not found")
}

// GetAllOptions returns all available options from the service
func (o *Options) GetAllOptions(ctx context.Context) ([]Option, error) {
	return o.service.GetOptions(ctx)
}

var _ JFieldType = &Options{}

// Boolean represents a boolean field type
type Boolean struct{}

// Scan implements JFieldType interface for boolean values
func (b *Boolean) Scan(ctx context.Context, field JField, row map[string]any) (value any, err error) {
	fieldName := field.Name()

	// Check if field exists in row
	if rawValue, exists := row[fieldName]; !exists {
		return nil, nil // Field doesn't exist, return nil
	} else if rawValue == nil {
		return nil, nil // Field is nil, return nil
	} else {
		// Convert to boolean
		boolValue, err := convertToBool(rawValue)
		if err != nil {
			return nil, err
		}
		return boolValue, nil
	}
}

// SetValue implements JFieldType interface for boolean values
func (b *Boolean) SetValue(ctx context.Context, field JField, value any, row map[string]any) error {
	fieldName := field.Name()

	// Handle nil values
	if value == nil {
		row[fieldName] = nil
		return nil
	}

	// Convert to boolean
	boolValue, err := convertToBool(value)
	if err != nil {
		return err
	}

	row[fieldName] = boolValue
	return nil
}

// Validate implements JFieldType interface for boolean values
func (b *Boolean) Validate(value any) error {
	if value == nil {
		return nil // Nil values are valid
	}

	// Try to convert to boolean to validate
	_, err := convertToBool(value)
	return err
}

// convertToBool converts various types to boolean
func convertToBool(value any) (bool, error) {
	reflectValue := reflect.ValueOf(value)

	// Handle nil
	if !reflectValue.IsValid() {
		return false, nil
	}

	// Handle direct boolean
	if reflectValue.Kind() == reflect.Bool {
		return reflectValue.Bool(), nil
	}

	// Handle pointer to boolean
	if reflectValue.Kind() == reflect.Ptr {
		if reflectValue.IsNil() {
			return false, nil
		}
		elem := reflectValue.Elem()
		if elem.Kind() == reflect.Bool {
			return elem.Bool(), nil
		}
	}

	// Handle string representations
	if reflectValue.Kind() == reflect.String {
		str := reflectValue.String()
		switch strings.ToLower(strings.TrimSpace(str)) {
		case "true", "1", "yes", "on", "enabled":
			return true, nil
		case "false", "0", "no", "off", "disabled", "":
			return false, nil
		default:
			return false, errors.New("invalid boolean string value")
		}
	}

	// Handle numeric values
	if reflectValue.Kind() == reflect.Int || reflectValue.Kind() == reflect.Int8 ||
		reflectValue.Kind() == reflect.Int16 || reflectValue.Kind() == reflect.Int32 ||
		reflectValue.Kind() == reflect.Int64 {
		return reflectValue.Int() != 0, nil
	}

	if reflectValue.Kind() == reflect.Uint || reflectValue.Kind() == reflect.Uint8 ||
		reflectValue.Kind() == reflect.Uint16 || reflectValue.Kind() == reflect.Uint32 ||
		reflectValue.Kind() == reflect.Uint64 {
		return reflectValue.Uint() != 0, nil
	}

	if reflectValue.Kind() == reflect.Float32 || reflectValue.Kind() == reflect.Float64 {
		return reflectValue.Float() != 0, nil
	}

	return false, errors.New("value cannot be converted to boolean")
}

var _ JFieldType = &Boolean{}
