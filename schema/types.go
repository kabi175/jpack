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

// JField represents a field in a schema
type JField interface {
	Name() string
	Type() JFieldType
	DefaultValue() any
	IsRequired() bool
	SetRequired(required bool) JField
	IsUnique() bool
	SetUnique(unique bool) JField
	Validation() ValidationFunc
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

// JSchema represents a schema definition
type JSchema interface {
	Name() string
	Fields() []JField
	Field(name string) (JField, bool)
	AddField(name string, fType JFieldType, defaultValue any) JField
	AddRef(name string, schema JSchema) JRef
	Refs() []JRef
	Ref(name string) (JRef, bool)
	Edges() []JEdge
	AddEdge(edge JEdge) JSchema
	Validate(ctx context.Context, rec JRecord) error
	AddValidation(fn ValidationFunc)
	Validations() []ValidationFunc
	GetIDField() JField
	SetIDField(field JField) JSchema
	IsImmutable() bool
	Clone() JSchema
	Freeze() JSchema
}

// JSchemaRegistry manages schemas
type JSchemaRegistry interface {
	Register(schema JSchema) error
	Get(name string) (JSchema, bool)
	Unregister(name string) error
	List() []string
	Replace(schema JSchema) error
	RegisterImmutable(schema JSchema) error
}
