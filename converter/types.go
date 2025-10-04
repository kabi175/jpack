// Package converter provides type conversion functionality for JPack.
// It includes converters for various data types and a registry for managing them.
//
// The converter package provides a flexible system for converting data between
// Go types and database representations, making it easy to work with different
// data types in a consistent manner.
//
// Example:
//
//	// Register a custom converter
//	customConverter := &MyCustomConverter{}
//	err := converter.RegisterConverter(customConverter)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Get a converter
//	conv, exists := converter.GetConverter(schema.JString)
//	if !exists {
//		log.Fatal("String converter not found")
//	}
//
//	// Convert data
//	dbValue, err := conv.ToDB("hello")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Convert back
//	goValue, err := conv.FromDB(dbValue)
//	if err != nil {
//		log.Fatal(err)
//	}
package converter

import (
	"fmt"

	"github.com/kabi175/jpack/logger"
	"github.com/kabi175/jpack/schema"
)

// JConverter defines the interface for data type converters.
// Converters handle the conversion between Go types and database representations.
//
// Example:
//
//	type MyCustomConverter struct{}
//
//	func (c *MyCustomConverter) ToDB(value any) (any, error) {
//		// Convert Go value to database representation
//		return value, nil
//	}
//
//	func (c *MyCustomConverter) FromDB(raw any) (any, error) {
//		// Convert database representation to Go value
//		return raw, nil
//	}
//
//	func (c *MyCustomConverter) Type() schema.JFieldType {
//		return schema.JString
//	}
type JConverter interface {
	// ToDB converts a Go value to its database representation
	ToDB(value any) (any, error)
	// FromDB converts a database representation to a Go value
	FromDB(raw any) (any, error)
	// Type returns the field type this converter handles
	Type() schema.JFieldType
}

// ConverterRegistry manages converters.
// It provides registration, retrieval, and management of type converters.
//
// Example:
//
//	registry := converter.NewConverterRegistry()
//
//	// Register a converter
//	customConverter := &MyCustomConverter{}
//	err := registry.Register(customConverter)
//
//	// Get a converter
//	conv, exists := registry.Get(schema.JString)
//	if !exists {
//		log.Fatal("String converter not found")
//	}
//
//	// List all registered converters
//	types := registry.List()
//	for _, fieldType := range types {
//		fmt.Printf("Registered converter for: %s\n", fieldType)
//	}
//
//	// Unregister a converter
//	err = registry.Unregister(schema.JString)
type ConverterRegistry interface {
	// Register registers a converter
	Register(converter JConverter) error
	// Get retrieves a converter by field type
	Get(fieldType schema.JFieldType) (JConverter, bool)
	// Unregister removes a converter by field type
	Unregister(fieldType schema.JFieldType) error
	// List returns all registered field types
	List() []schema.JFieldType
}

// jConverterRegistry implements ConverterRegistry.
// It provides a thread-safe registry for managing type converters.
type jConverterRegistry struct {
	converters map[schema.JFieldType]JConverter
}

// NewConverterRegistry creates a new converter registry.
// It initializes the registry with default converters for common types.
//
// Example:
//
//	registry := converter.NewConverterRegistry()
//
//	// The registry is pre-populated with default converters
//	types := registry.List()
//	fmt.Printf("Available converters: %v\n", types)
func NewConverterRegistry() ConverterRegistry {
	registry := &jConverterRegistry{
		converters: make(map[schema.JFieldType]JConverter),
	}

	// Register default converters
	registry.Register(NewStringConverter())
	registry.Register(NewIntConverter())
	registry.Register(NewInt64Converter())
	registry.Register(NewFloatConverter())
	registry.Register(NewFloat64Converter())
	registry.Register(NewBoolConverter())
	registry.Register(NewTimeConverter())
	registry.Register(NewObjectConverter())
	registry.Register(NewArrayConverter())
	registry.Register(NewBinaryConverter())
	registry.Register(NewObjectIDConverter())

	return registry
}

// Global converter registry
var globalConverterRegistry ConverterRegistry = NewConverterRegistry()

// GetGlobalConverterRegistry returns the global converter registry.
// This is a convenience function for accessing the global registry instance.
//
// Example:
//
//	registry := converter.GetGlobalConverterRegistry()
//
//	// Use the global registry
//	conv, exists := registry.Get(schema.JString)
//	if !exists {
//		log.Fatal("String converter not found")
//	}
func GetGlobalConverterRegistry() ConverterRegistry {
	return globalConverterRegistry
}

func (r *jConverterRegistry) Register(converter JConverter) error {
	if converter == nil {
		logger.Converter.Error().Msg("attempted to register nil converter")
		return fmt.Errorf("converter cannot be nil")
	}

	logger.Converter.Debug().
		Str("type", string(converter.Type())).
		Msg("registering converter")

	r.converters[converter.Type()] = converter

	logger.Converter.Debug().
		Str("type", string(converter.Type())).
		Int("total_converters", len(r.converters)).
		Msg("converter registered successfully")

	return nil
}

func (r *jConverterRegistry) Get(fieldType schema.JFieldType) (JConverter, bool) {
	converter, exists := r.converters[fieldType]
	if !exists {
		logger.Converter.Debug().
			Str("type", string(fieldType)).
			Msg("converter not found")
	} else {
		logger.Converter.Debug().
			Str("type", string(fieldType)).
			Msg("converter found")
	}
	return converter, exists
}

func (r *jConverterRegistry) Unregister(fieldType schema.JFieldType) error {
	if _, exists := r.converters[fieldType]; !exists {
		logger.Converter.Warn().
			Str("type", string(fieldType)).
			Msg("attempted to unregister non-existent converter")
		return fmt.Errorf("converter for type %s is not registered", fieldType)
	}

	logger.Converter.Info().
		Str("type", string(fieldType)).
		Msg("unregistering converter")

	delete(r.converters, fieldType)

	logger.Converter.Debug().
		Str("type", string(fieldType)).
		Int("remaining_converters", len(r.converters)).
		Msg("converter unregistered successfully")

	return nil
}

func (r *jConverterRegistry) List() []schema.JFieldType {
	types := make([]schema.JFieldType, 0, len(r.converters))
	for fieldType := range r.converters {
		types = append(types, fieldType)
	}

	logger.Converter.Debug().
		Int("count", len(types)).
		Msg("listing registered converters")

	return types
}

// RegisterConverter is a convenience function to register a converter in the global registry.
// This is a shortcut for registering converters without creating a registry instance.
//
// Example:
//
//	customConverter := &MyCustomConverter{}
//	err := converter.RegisterConverter(customConverter)
//	if err != nil {
//		log.Fatal(err)
//	}
func RegisterConverter(converter JConverter) error {
	return globalConverterRegistry.Register(converter)
}

// GetConverter is a convenience function to get a converter from the global registry.
// This is a shortcut for retrieving converters without creating a registry instance.
//
// Example:
//
//	conv, exists := converter.GetConverter(schema.JString)
//	if !exists {
//		log.Fatal("String converter not found")
//	}
//
//	// Use the converter
//	dbValue, err := conv.ToDB("hello")
//	if err != nil {
//		log.Fatal(err)
//	}
func GetConverter(fieldType schema.JFieldType) (JConverter, bool) {
	return globalConverterRegistry.Get(fieldType)
}
