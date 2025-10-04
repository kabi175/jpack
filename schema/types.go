// Package schema provides the core schema definition and management functionality for JPack.
// It includes immutable schemas, field definitions, validation, and relationship management.
//
// The schema system is designed to be immutable by default, ensuring thread safety
// and preventing accidental modifications. Schemas are created using the builder pattern
// and can be updated using callback-based modifications that create new instances.
//
// Example:
//
//	userSchema := schema.NewSchemaBuilder("User").
//		AddRequiredField("name", schema.JString, nil).
//		AddRequiredUniqueField("email", schema.JString, nil).
//		AddField("age", schema.JInt, 18).
//		Build()
//
//	// Update schema (creates new instance)
//	updatedSchema := userSchema.Update(func(sb *schema.SchemaBuilder) {
//		sb.AddField("phone", schema.JString, nil)
//	})
package schema

import (
	"context"
)

// JFieldType represents the type of a field in a schema.
// It defines the data type that can be stored in a field.
type JFieldType string

const (
	// JString represents a string field type
	JString JFieldType = "string"
	// JInt represents an integer field type
	JInt JFieldType = "int"
	// JInt64 represents a 64-bit integer field type
	JInt64 JFieldType = "int64"
	// JFloat represents a 32-bit floating point field type
	JFloat JFieldType = "float"
	// JFloat64 represents a 64-bit floating point field type
	JFloat64 JFieldType = "float64"
	// JBool represents a boolean field type
	JBool JFieldType = "bool"
	// JTime represents a time.Time field type
	JTime JFieldType = "time"
	// JObject represents an object/map field type
	JObject JFieldType = "object"
	// JArray represents an array/slice field type
	JArray JFieldType = "array"
	// JBinary represents binary data field type
	JBinary JFieldType = "binary"
	// JObjectID represents a MongoDB ObjectID field type
	JObjectID JFieldType = "objectid"
)

// JField represents a field in a schema (always immutable).
// Fields define the structure and constraints for data stored in records.
//
// Example:
//
//	field := schema.NewFieldBuilder("email", schema.JString, nil).
//		Required().
//		Unique().
//		Build()
//
//	// Update field (creates new instance)
//	updatedField := field.Update(func(fb *schema.FieldBuilder) {
//		fb.WithValidation(emailValidator)
//	})
type JField interface {
	// Name returns the field name
	Name() string
	// Type returns the field type
	Type() JFieldType
	// DefaultValue returns the default value for the field
	DefaultValue() any
	// IsRequired returns true if the field is required
	IsRequired() bool
	// IsUnique returns true if the field must be unique
	IsUnique() bool
	// Validation returns the validation function for the field
	Validation() ValidationFunc
	// String returns a string representation of the field
	String() string

	// Update creates a new field with modifications applied via callback.
	// The callback receives a FieldBuilder initialized with a clone of this field.
	Update(fn func(*FieldBuilder)) JField

	// Backward compatibility methods (return new instances)
	SetRequired(required bool) JField
	SetUnique(unique bool) JField
	SetValidation(fn ValidationFunc) JField
	IsImmutable() bool
	Freeze() JField
}

// JRef represents a reference to another schema.
// References establish relationships between schemas and can be single or array references.
//
// Example:
//
//	userSchema := schema.NewSchemaBuilder("User").
//		AddRef("profile", profileSchema).
//		AddArrayRef("orders", orderSchema).
//		Build()
type JRef interface {
	// Name returns the reference name
	Name() string
	// TargetSchema returns the target schema
	TargetSchema() JSchema
	// IsArray returns true if this is an array reference
	IsArray() bool
}

// JEdge represents a relationship edge between schemas.
// Edges define the type and direction of relationships between schemas.
//
// Example:
//
//	edge := schema.NewJEdge("user_orders", userSchema, orderSchema, schema.EdgeOneToMany)
type JEdge interface {
	// Name returns the edge name
	Name() string
	// FromSchema returns the source schema
	FromSchema() JSchema
	// ToSchema returns the target schema
	ToSchema() JSchema
	// Type returns the relationship type
	Type() EdgeType
}

// EdgeType represents the type of relationship between schemas.
type EdgeType string

const (
	// EdgeOneToOne represents a one-to-one relationship
	EdgeOneToOne EdgeType = "one-to-one"
	// EdgeOneToMany represents a one-to-many relationship
	EdgeOneToMany EdgeType = "one-to-many"
	// EdgeManyToOne represents a many-to-one relationship
	EdgeManyToOne EdgeType = "many-to-one"
	// EdgeManyToMany represents a many-to-many relationship
	EdgeManyToMany EdgeType = "many-to-many"
)

// JRecord represents a data record.
// Records are the actual data instances that conform to a schema.
//
// Example:
//
//	record := schema.NewJRecord().
//		Set("id", "user_001").
//		Set("name", "John Doe").
//		Set("email", "john@example.com")
//
//	name := record.Get("name").(string)
//	hasEmail := record.Has("email")
//	recordMap := record.ToMap()
type JRecord interface {
	// Get retrieves a value by key
	Get(key string) any
	// Set sets a value for a key and returns a new record
	Set(key string, value any) JRecord
	// Has checks if a key exists in the record
	Has(key string) bool
	// Delete removes a key and returns a new record
	Delete(key string) JRecord
	// Keys returns all keys in the record
	Keys() []string
	// ToMap converts the record to a map
	ToMap() map[string]any
	// FromMap creates a record from a map
	FromMap(data map[string]any) JRecord
	// Clone creates a copy of the record
	Clone() JRecord
}

// ValidationFunc represents a validation function.
// Validation functions are used to validate records against schema rules.
//
// Example:
//
//	emailValidator := func(ctx context.Context, rec schema.JRecord) error {
//		email := rec.Get("email").(string)
//		if !strings.Contains(email, "@") {
//			return fmt.Errorf("invalid email format")
//		}
//		return nil
//	}
type ValidationFunc func(ctx context.Context, rec JRecord) error

// JSchema represents a schema definition (always immutable).
// Schemas define the structure, constraints, and relationships for data records.
//
// Example:
//
//	userSchema := schema.NewSchemaBuilder("User").
//		AddRequiredField("name", schema.JString, nil).
//		AddRequiredUniqueField("email", schema.JString, nil).
//		AddField("age", schema.JInt, 18).
//		Build()
//
//	// Validate a record against the schema
//	err := userSchema.Validate(ctx, userRecord)
//
//	// Update schema (creates new instance)
//	updatedSchema := userSchema.Update(func(sb *schema.SchemaBuilder) {
//		sb.AddField("phone", schema.JString, nil)
//	})
type JSchema interface {
	// Read-only accessors
	// Name returns the schema name
	Name() string
	// Fields returns all fields in the schema
	Fields() []JField
	// Field returns a field by name
	Field(name string) (JField, bool)
	// Refs returns all references in the schema
	Refs() []JRef
	// Ref returns a reference by name
	Ref(name string) (JRef, bool)
	// Edges returns all edges in the schema
	Edges() []JEdge
	// Validations returns all validation functions
	Validations() []ValidationFunc
	// GetIDField returns the ID field
	GetIDField() JField
	// String returns a string representation of the schema
	String() string

	// Validate validates a record against the schema
	Validate(ctx context.Context, rec JRecord) error

	// Update creates a new schema with modifications applied via callback.
	// The callback receives a SchemaBuilder initialized with a clone of this schema.
	Update(fn func(*SchemaBuilder)) JSchema

	// Backward compatibility methods
	Clone() JSchema
	IsImmutable() bool
	Freeze() JSchema
}

// ExternalSchemaSource represents a source for schema definitions.
// It provides a way to load schemas from external sources like files, databases, or HTTP services.
//
// Example:
//
//	type FileSchemaSource struct {
//		Path string
//	}
//
//	func (f *FileSchemaSource) Load(schemaName string) (schema.JSchema, error) {
//		data, err := os.ReadFile(f.Path)
//		if err != nil {
//			return nil, err
//		}
//		// Parse and return schema
//		return parseSchema(data), nil
//	}
type ExternalSchemaSource interface {
	// Load loads a schema by name from the external source
	Load(schemaName string) (JSchema, error)
}

// JSchemaRegistry manages schemas.
// It provides registration, retrieval, and management of schemas with support for lazy loading.
//
// Example:
//
//	registry := schema.NewJSchemaRegistry()
//
//	// Register schema immediately
//	userSchema := schema.NewSchemaBuilder("User").Build()
//	registry.Register(userSchema)
//
//	// Register schema lazily
//	registry.RegisterLazy("Order", &FileSchemaSource{Path: "order_schema.json"})
//
//	// Get schema (triggers lazy load if needed)
//	orderSchema, ok := registry.Get("Order")
type JSchemaRegistry interface {
	// Register registers a schema immediately
	Register(schema JSchema) error
	// RegisterLazy registers a schema for lazy loading
	RegisterLazy(schemaName string, datasource ExternalSchemaSource) error
	// Get retrieves a schema by name
	Get(name string) (JSchema, bool)
	// Unregister removes a schema from the registry
	Unregister(name string) error
	// List returns all registered schemas
	List() []JSchema
	// Replace replaces an existing schema
	Replace(schema JSchema) error
	// RegisterImmutable registers an immutable schema (deprecated - all schemas are immutable)
	RegisterImmutable(schema JSchema) error
}
