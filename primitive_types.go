package jpack

import (
	"context"
	"errors"
	"math"
	"reflect"
	"strconv"

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
