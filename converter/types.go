package converter

import (
	"fmt"

	"github.com/kabi175/jpack/logger"
	"github.com/kabi175/jpack/schema"
)

// JConverter defines the interface for data type converters
type JConverter interface {
	ToDB(value any) (any, error)
	FromDB(raw any) (any, error)
	Type() schema.JFieldType
}

// ConverterRegistry manages converters
type ConverterRegistry interface {
	Register(converter JConverter) error
	Get(fieldType schema.JFieldType) (JConverter, bool)
	Unregister(fieldType schema.JFieldType) error
	List() []schema.JFieldType
}

// jConverterRegistry implements ConverterRegistry
type jConverterRegistry struct {
	converters map[schema.JFieldType]JConverter
}

// NewConverterRegistry creates a new converter registry
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

// GetGlobalConverterRegistry returns the global converter registry
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

// RegisterConverter is a convenience function to register a converter in the global registry
func RegisterConverter(converter JConverter) error {
	return globalConverterRegistry.Register(converter)
}

// GetConverter is a convenience function to get a converter from the global registry
func GetConverter(fieldType schema.JFieldType) (JConverter, bool) {
	return globalConverterRegistry.Get(fieldType)
}
