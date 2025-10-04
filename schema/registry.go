package schema

import (
	"fmt"
	"sync"

	"github.com/kabi175/jpack/logger"
)

// jSchemaRegistry implements JSchemaRegistry interface
type jSchemaRegistry struct {
	schemas         map[string]JSchema
	externalSources map[string]ExternalSchemaSource
	mutex           sync.RWMutex
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

func (r *jSchemaRegistry) RegisterExternalSource(datasource ExternalSchemaSource) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if datasource == nil {
		logger.Schema.Error().Msg("attempted to register lazy schema with nil datasource")
		return fmt.Errorf("datasource cannot be nil")
	}

	if datasource.Name() == "" {
		logger.Schema.Error().Msg("attempted to register lazy schema with empty name")
		return fmt.Errorf("schema name cannot be empty")
	}

	if _, exists := r.externalSources[datasource.Name()]; exists {
		logger.Schema.Warn().
			Str("datasource", datasource.Name()).
			Msg("datasource already registered")
		return fmt.Errorf("datasource '%s' is already registered", datasource.Name())
	}

	r.externalSources[datasource.Name()] = datasource

	logger.Schema.Debug().
		Str("datasource", datasource.Name()).
		Int("total_external_sources", len(r.externalSources)).
		Msg("datasource registered successfully")

	return nil
}

func (r *jSchemaRegistry) Get(name string) (JSchema, bool) {
	r.mutex.RLock()
	entry, exists := r.schemas[name]
	r.mutex.RUnlock()

	if !exists {
		logger.Schema.Debug().
			Str("schema", name).
			Msg("schema not found in registry")

		for _, datasource := range r.externalSources {
			if ok, err := datasource.Has(name); ok {
				schema, err := datasource.Load(name)
				if err != nil {
					logger.Schema.Error().
						Str("datasource", datasource.Name()).
						Err(err).
						Msg("failed to load schema from external source")
					return nil, false
				}
				return schema, true
			} else if err != nil {
				logger.Schema.Error().
					Str("datasource", datasource.Name()).
					Err(err).
					Msg("failed to check if schema exists in external source")
				return nil, false
			}
		}

		logger.Schema.Debug().
			Str("schema", name).
			Msg("schema not found in registry or external sources")

		return nil, false
	}

	// If already loaded, return immediately
	if entry != nil {
		logger.Schema.Debug().
			Str("schema", name).
			Msg("schema retrieved from registry (already loaded)")
		return entry, true
	}

	return nil, false
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

// RegisterExternalSource is a convenience function to register a external source in the global registry
func RegisterLazySchema(datasource ExternalSchemaSource) error {
	return globalRegistry.RegisterExternalSource(datasource)
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
