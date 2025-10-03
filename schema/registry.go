package schema

import (
	"fmt"
	"sync"

	"github.com/kabi175/jpack/logger"
)

// jSchemaRegistry implements JSchemaRegistry interface
type jSchemaRegistry struct {
	schemas map[string]JSchema
	mutex   sync.RWMutex
}

// NewJSchemaRegistry creates a new schema registry
func NewJSchemaRegistry() JSchemaRegistry {
	return &jSchemaRegistry{
		schemas: make(map[string]JSchema),
	}
}

// Global registry instance
var globalRegistry JSchemaRegistry = NewJSchemaRegistry()

// GetGlobalRegistry returns the global schema registry
func GetGlobalRegistry() JSchemaRegistry {
	return globalRegistry
}

func (r *jSchemaRegistry) Register(schema JSchema) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if schema == nil {
		logger.Schema.Error().Msg("attempted to register nil schema")
		return fmt.Errorf("schema cannot be nil")
	}

	if schema.Name() == "" {
		logger.Schema.Error().Msg("attempted to register schema with empty name")
		return fmt.Errorf("schema name cannot be empty")
	}

	if _, exists := r.schemas[schema.Name()]; exists {
		logger.Schema.Warn().
			Str("schema", schema.Name()).
			Msg("schema already registered")
		return fmt.Errorf("schema '%s' is already registered", schema.Name())
	}

	logger.Schema.Info().
		Str("schema", schema.Name()).
		Msg("registering schema")

	// Make the schema immutable when registering
	immutableSchema := schema.Clone().Freeze()
	r.schemas[schema.Name()] = immutableSchema

	logger.Schema.Debug().
		Str("schema", schema.Name()).
		Int("total_schemas", len(r.schemas)).
		Msg("schema registered successfully")

	return nil
}

func (r *jSchemaRegistry) Get(name string) (JSchema, bool) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	schema, exists := r.schemas[name]
	if !exists {
		logger.Schema.Debug().
			Str("schema", name).
			Msg("schema not found in registry")
	} else {
		logger.Schema.Debug().
			Str("schema", name).
			Msg("schema retrieved from registry")
	}

	return schema, exists
}

func (r *jSchemaRegistry) Unregister(name string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.schemas[name]; !exists {
		logger.Schema.Warn().
			Str("schema", name).
			Msg("attempted to unregister non-existent schema")
		return fmt.Errorf("schema '%s' is not registered", name)
	}

	logger.Schema.Info().
		Str("schema", name).
		Msg("unregistering schema")

	delete(r.schemas, name)

	logger.Schema.Debug().
		Str("schema", name).
		Int("remaining_schemas", len(r.schemas)).
		Msg("schema unregistered successfully")

	return nil
}

func (r *jSchemaRegistry) List() []string {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	names := make([]string, 0, len(r.schemas))
	for name := range r.schemas {
		names = append(names, name)
	}

	logger.Schema.Debug().
		Int("count", len(names)).
		Strs("schemas", names).
		Msg("listing registered schemas")

	return names
}

// RegisterSchema is a convenience function to register a schema in the global registry
func RegisterSchema(schema JSchema) error {
	return globalRegistry.Register(schema)
}

// GetSchema is a convenience function to get a schema from the global registry
func GetSchema(name string) (JSchema, bool) {
	return globalRegistry.Get(name)
}

// ReplaceSchema is a convenience function to replace a schema in the global registry
func ReplaceSchema(schema JSchema) error {
	return globalRegistry.Replace(schema)
}

// RegisterImmutableSchema is a convenience function to register an immutable schema in the global registry
func RegisterImmutableSchema(schema JSchema) error {
	return globalRegistry.RegisterImmutable(schema)
}

// Replace replaces an existing schema with a new one
func (r *jSchemaRegistry) Replace(schema JSchema) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if schema == nil {
		logger.Schema.Error().Msg("attempted to replace with nil schema")
		return fmt.Errorf("schema cannot be nil")
	}

	if schema.Name() == "" {
		logger.Schema.Error().Msg("attempted to replace schema with empty name")
		return fmt.Errorf("schema name cannot be empty")
	}

	// Check if schema exists
	if _, exists := r.schemas[schema.Name()]; !exists {
		logger.Schema.Warn().
			Str("schema", schema.Name()).
			Msg("attempted to replace non-existent schema")
		return fmt.Errorf("schema '%s' is not registered", schema.Name())
	}

	logger.Schema.Info().
		Str("schema", schema.Name()).
		Msg("replacing schema")

	// Make the schema immutable when replacing
	immutableSchema := schema.Clone().Freeze()
	r.schemas[schema.Name()] = immutableSchema

	logger.Schema.Debug().
		Str("schema", schema.Name()).
		Msg("schema replaced successfully")

	return nil
}

// RegisterImmutable registers a schema as immutable without cloning
func (r *jSchemaRegistry) RegisterImmutable(schema JSchema) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if schema == nil {
		logger.Schema.Error().Msg("attempted to register nil schema as immutable")
		return fmt.Errorf("schema cannot be nil")
	}

	if schema.Name() == "" {
		logger.Schema.Error().Msg("attempted to register schema with empty name as immutable")
		return fmt.Errorf("schema name cannot be empty")
	}

	if _, exists := r.schemas[schema.Name()]; exists {
		logger.Schema.Warn().
			Str("schema", schema.Name()).
			Msg("attempted to register already registered schema as immutable")
		return fmt.Errorf("schema '%s' is already registered", schema.Name())
	}

	// Ensure the schema is immutable
	if !schema.IsImmutable() {
		logger.Schema.Error().
			Str("schema", schema.Name()).
			Msg("schema must be immutable to use RegisterImmutable")
		return fmt.Errorf("schema must be immutable to use RegisterImmutable")
	}

	logger.Schema.Info().
		Str("schema", schema.Name()).
		Msg("registering immutable schema")

	r.schemas[schema.Name()] = schema

	logger.Schema.Debug().
		Str("schema", schema.Name()).
		Int("total_schemas", len(r.schemas)).
		Msg("immutable schema registered successfully")

	return nil
}
