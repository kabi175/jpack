package jpack

import (
	"context"
)

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
	AddField(field JField) JSchema

	Edge() []JEdge
	AddEdge(edge JEdge) JSchema

	Validate(JRecord) error
}

type JPolicy interface {
	IsValid(ctx context.Context, record JRecord) error
}

type SchemaBuilder struct {
	name   string
	fields []JField
	edges  []JEdge

	schema *schemaImpl
}

func (s *SchemaBuilder) appendFieldIfNotPresent(field JField) {
	for _, f := range s.fields {
		if f.Name() == field.Name() {
			// If a field with the same name already exists, return the schema builder
			// without adding a new field.
			return
		}
	}

	s.fields = append(s.fields, field)
}

func (s *SchemaBuilder) FieldWithDefault(name string, fType JFieldType, defaultValue any) *SchemaBuilder {

	field := &fieldImpl{
		name:         name,
		fType:        fType,
		schema:       s.schema,
		defaultValue: defaultValue,
	}

	s.appendFieldIfNotPresent(field)
	return s
}

func (s *SchemaBuilder) Field(name string, fType JFieldType) *SchemaBuilder {
	return s.FieldWithDefault(name, fType, nil)
}

func (s *SchemaBuilder) Ref(name string, schema JSchema) *SchemaBuilder {
	field := &refImpl{
		fieldImpl: fieldImpl{
			name:   name,
			fType:  &Ref{},
			schema: s.schema,
		},
		relSchema: schema,
	}

	s.appendFieldIfNotPresent(field)
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

func (s *SchemaBuilder) Build() JSchema {
	s.schema.fields = s.fields
	s.schema.edges = s.edges

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
