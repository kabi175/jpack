package schema

import (
	"context"
)

// JFieldType represents the type of a field
type JFieldType string

const (
	JString   JFieldType = "string"
	JInt      JFieldType = "int"
	JInt64    JFieldType = "int64"
	JFloat    JFieldType = "float"
	JFloat64  JFieldType = "float64"
	JBool     JFieldType = "bool"
	JTime     JFieldType = "time"
	JObject   JFieldType = "object"
	JArray    JFieldType = "array"
	JBinary   JFieldType = "binary"
	JObjectID JFieldType = "objectid"
)

// JField represents a field in a schema (always immutable)
type JField interface {
	Name() string
	Type() JFieldType
	DefaultValue() any
	IsRequired() bool
	IsUnique() bool
	Validation() ValidationFunc
	String() string

	// Update creates a new field with modifications applied via callback
	Update(fn func(*FieldBuilder)) JField

	// Backward compatibility methods (return new instances)
	SetRequired(required bool) JField
	SetUnique(unique bool) JField
	SetValidation(fn ValidationFunc) JField
	IsImmutable() bool
	Freeze() JField
}

// JRef represents a reference to another schema
type JRef interface {
	Name() string
	TargetSchema() JSchema
	IsArray() bool
}

// JEdge represents a relationship edge between schemas
type JEdge interface {
	Name() string
	FromSchema() JSchema
	ToSchema() JSchema
	Type() EdgeType
}

// EdgeType represents the type of relationship
type EdgeType string

const (
	EdgeOneToOne   EdgeType = "one-to-one"
	EdgeOneToMany  EdgeType = "one-to-many"
	EdgeManyToOne  EdgeType = "many-to-one"
	EdgeManyToMany EdgeType = "many-to-many"
)

// JRecord represents a data record
type JRecord interface {
	Get(key string) any
	Set(key string, value any) JRecord
	Has(key string) bool
	Delete(key string) JRecord
	Keys() []string
	ToMap() map[string]any
	FromMap(data map[string]any) JRecord
	Clone() JRecord
}

// ValidationFunc represents a validation function
type ValidationFunc func(ctx context.Context, rec JRecord) error

// JSchema represents a schema definition (always immutable)
type JSchema interface {
	// Read-only accessors
	Name() string
	Fields() []JField
	Field(name string) (JField, bool)
	Refs() []JRef
	Ref(name string) (JRef, bool)
	Edges() []JEdge
	Validations() []ValidationFunc
	GetIDField() JField
	String() string

	// Validation
	Validate(ctx context.Context, rec JRecord) error

	// Update creates a new schema with modifications applied via callback
	// The callback receives a SchemaBuilder initialized with a clone of this schema
	Update(fn func(*SchemaBuilder)) JSchema

	// Backward compatibility methods
	Clone() JSchema
	IsImmutable() bool
	Freeze() JSchema
}

// ExternalSchemaSource represents a source for schema definitions
type ExternalSchemaSource interface {
	Load(schemaName string) (JSchema, error)
}

// JSchemaRegistry manages schemas
type JSchemaRegistry interface {
	Register(schema JSchema) error
	RegisterLazy(schemaName string, datasource ExternalSchemaSource) error
	Get(name string) (JSchema, bool)
	Unregister(name string) error
	List() []JSchema
	Replace(schema JSchema) error
	RegisterImmutable(schema JSchema) error
}
