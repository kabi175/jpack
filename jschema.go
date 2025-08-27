package jpack

import (
	"context"
)

// ValidationFunc is a function type for custom validation rules
type ValidationFunc func(ctx context.Context, record JRecord) error

type JFieldType interface {
	Validate(value any) error

	// Scan converts a database value to the field's type.
	// It is used when reading from the database.
	// It should return the value in the field's type or an error if the conversion fails.
	Scan(ctx context.Context, field JField, row map[string]any) (value any, err error)

	// Sets the value in the database row
	SetValue(ctx context.Context, field JField, value any, row map[string]any) error
}

type JField interface {
	Name() string
	Type() JFieldType
	Schema() JSchema
	Default() any
}

type JRef interface {
	JField
	RelSchema() JSchema
}

type JEdge interface {
	Name() string
	Schema() JSchema
	Ref() JRef
}

type JSchema interface {
	Name() string
	Fields() []JField
	Field(name string) (JField, bool)
	AddField(name string, fType JFieldType, defaultValue any) JField
	AddRef(name string, schema JSchema) JRef

	Edge() []JEdge
	AddEdge(edge JEdge) JSchema

	Validate(context.Context, JRecord) error
	AddValidation(ValidationFunc)
	Validations() []ValidationFunc
}

type JPolicy interface {
	IsValid(ctx context.Context, record JRecord) error
}

type SchemaBuilder struct {
	name        string
	fields      []JField
	edges       []JEdge
	validations []ValidationFunc

	schema *schemaImpl
}

func (s *SchemaBuilder) FieldWithDefault(name string, fType JFieldType, defaultValue any) *SchemaBuilder {
	s.schema.AddField(name, fType, defaultValue)
	return s
}

func (s *SchemaBuilder) Field(name string, fType JFieldType) *SchemaBuilder {
	return s.FieldWithDefault(name, fType, nil)
}

func (s *SchemaBuilder) Ref(name string, schema JSchema) *SchemaBuilder {
	s.schema.AddRef(name, schema)
	return s
}

func (s *SchemaBuilder) RefWithEdge(name string, schema JSchema, edgeName string) *SchemaBuilder {
	ref := s.schema.AddRef(name, schema)

	schema.AddEdge(&edgeImpl{
		name:   edgeName,
		schema: s.schema,
		field:  ref,
	})

	return s
}

func (s *SchemaBuilder) Edge(name string, schema JSchema, ref JRef) *SchemaBuilder {
	for _, edge := range s.edges {
		if edge.Name() == name {
			// If an edge with the same name already exists, return the schema builder
			// without adding a new edge.
			return s
		}
	}

	s.edges = append(s.edges, &edgeImpl{
		name:   name,
		schema: schema,
		field:  ref,
	})

	return s
}

// Validation adds a custom validation function to the schema
func (s *SchemaBuilder) Validation(validation ValidationFunc) *SchemaBuilder {
	s.validations = append(s.validations, validation)
	return s
}

func (s *SchemaBuilder) Build() JSchema {
	s.schema.validations = s.validations

	return s.schema
}

func NewSchema(name string) *SchemaBuilder {
	return &SchemaBuilder{
		name: name,
		schema: &schemaImpl{
			name: name,
		},
	}
}
