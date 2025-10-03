package schema

import (
	"fmt"
	"sync"

	"github.com/kabi175/jpack/logger"
)

// lazyEntry represents a lazy-loaded schema entry
type lazyEntry struct {
	schema     JSchema
	datasource ExternalSchemaSource
	once       sync.Once
	loaded     bool
	err        error
}

// jSchemaRegistry implements JSchemaRegistry interface
type jSchemaRegistry struct {
	schemas map[string]*lazyEntry
	mutex   sync.RWMutex
}

// NewJSchemaRegistry creates a new schema registry
func NewJSchemaRegistry() JSchemaRegistry {
	return &jSchemaRegistry{
		schemas: make(map[string]*lazyEntry),
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
	r.schemas[schema.Name()] = &lazyEntry{
		schema: immutableSchema,
		loaded: true,
	}

	logger.Schema.Debug().
		Str("schema", schema.Name()).
		Int("total_schemas", len(r.schemas)).
		Msg("schema registered successfully")

	return nil
}

func (r *jSchemaRegistry) RegisterLazy(schemaName string, datasource ExternalSchemaSource) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if schemaName == "" {
		logger.Schema.Error().Msg("attempted to register lazy schema with empty name")
		return fmt.Errorf("schema name cannot be empty")
	}

	if datasource == nil {
		logger.Schema.Error().Msg("attempted to register lazy schema with nil datasource")
		return fmt.Errorf("datasource cannot be nil")
	}

	if _, exists := r.schemas[schemaName]; exists {
		logger.Schema.Warn().
			Str("schema", schemaName).
			Msg("schema already registered")
		return fmt.Errorf("schema '%s' is already registered", schemaName)
	}

	logger.Schema.Info().
		Str("schema", schemaName).
		Msg("registering lazy schema")

	r.schemas[schemaName] = &lazyEntry{
		datasource: datasource,
		loaded:     false,
	}

	logger.Schema.Debug().
		Str("schema", schemaName).
		Int("total_schemas", len(r.schemas)).
		Msg("lazy schema registered successfully")

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
		return nil, false
	}

	// If already loaded, return immediately
	if entry.loaded {
		logger.Schema.Debug().
			Str("schema", name).
			Msg("schema retrieved from registry (already loaded)")
		return entry.schema, true
	}

	// Lazy load the schema using sync.Once for thread safety
	entry.once.Do(func() {
		logger.Schema.Info().
			Str("schema", name).
			Msg("lazy loading schema")

		schema, err := entry.datasource.Load(name)
		if err != nil {
			logger.Schema.Error().
				Str("schema", name).
				Err(err).
				Msg("failed to lazy load schema")
			entry.err = err
			return
		}

		// Make the schema immutable
		immutableSchema := schema.Clone().Freeze()
		entry.schema = immutableSchema
		entry.loaded = true

		logger.Schema.Debug().
			Str("schema", name).
			Msg("schema lazy loaded successfully")
	})

	if entry.err != nil {
		logger.Schema.Error().
			Str("schema", name).
			Err(entry.err).
			Msg("schema lazy loading failed")
		return nil, false
	}

	return entry.schema, true
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

func (r *jSchemaRegistry) List() []JSchema {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	schemas := make([]JSchema, 0, len(r.schemas))
	for name, entry := range r.schemas {
		// Only include schemas that have been loaded
		if entry.loaded && entry.schema != nil {
			schemas = append(schemas, entry.schema)
		} else {
			logger.Schema.Debug().
				Str("schema", name).
				Msg("skipping unloaded schema in list")
		}
	}

	logger.Schema.Debug().
		Int("count", len(schemas)).
		Msg("listing loaded schemas")

	return schemas
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

// RegisterLazySchema is a convenience function to register a lazy schema in the global registry
func RegisterLazySchema(schemaName string, datasource ExternalSchemaSource) error {
	return globalRegistry.RegisterLazy(schemaName, datasource)
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
	r.schemas[schema.Name()] = &lazyEntry{
		schema: immutableSchema,
		loaded: true,
	}

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

	r.schemas[schema.Name()] = &lazyEntry{
		schema: schema,
		loaded: true,
	}

	logger.Schema.Debug().
		Str("schema", schema.Name()).
		Int("total_schemas", len(r.schemas)).
		Msg("immutable schema registered successfully")

	return nil
}
