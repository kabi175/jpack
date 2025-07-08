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

type JEdge interface {
	Name() string
	Schema() JSchema
	Field() JField
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

type Field struct {
	name         string
	fType        JFieldType
	schema       JSchema
	defaultValue any
}

// Default implements JField.
func (f *Field) Default() any {
	return f.defaultValue
}

// Name implements JField.
func (f *Field) Name() string {
	return f.name
}

// Schema implements JField.
func (f *Field) Schema() JSchema {
	return f.schema
}

// Type implements JField.
func (f *Field) Type() JFieldType {
	return f.fType
}

var _ JField = &Field{}

type Edge struct {
	name   string
	schema JSchema
	field  JField
}

// Field implements JEdge.
func (e *Edge) Field() JField {
	return e.field
}

// Name implements JEdge.
func (e *Edge) Name() string {
	return e.name
}

// Schema implements JEdge.
func (e *Edge) Schema() JSchema {
	return e.schema
}

var _ JEdge = &Edge{}

type schemaImpl struct {
	name   string
	fields []JField
	edges  []JEdge
}

// AddEdge implements JSchema.
func (s *schemaImpl) AddEdge(edge JEdge) JSchema {
	for _, e := range s.edges {
		if edge.Name() == e.Name() {
			// If a field with the same name already exists, return the schema builder
			// without adding a new field.
			return s
		}
	}

	s.edges = append(s.edges, edge)
	return s
}

// AddField implements JSchema.
func (s *schemaImpl) AddField(field JField) JSchema {
	for _, f := range s.fields {
		if field.Name() == f.Name() {
			// If a field with the same name already exists, return the schema builder
			// without adding a new field.
			return s
		}
	}

	s.fields = append(s.fields, field)
	return s
}

// Field implements JSchema.
func (s *schemaImpl) Field(name string) (JField, bool) {
	for _, f := range s.fields {
		if f.Name() == name {
			return f, true
		}
	}
	return nil, false

}

// Edge implements JSchema.
func (s *schemaImpl) Edge() []JEdge {
	return s.edges
}

// Fields implements JSchema.
func (s *schemaImpl) Fields() []JField {
	return s.fields
}

// Name implements JSchema.
func (s *schemaImpl) Name() string {
	return s.name
}

// Validate implements JSchema.
func (s *schemaImpl) Validate(JRecord) error {
	panic("unimplemented")
}

var _ JSchema = &schemaImpl{}

type SchemaBuilder struct {
	name   string
	fields []JField
	edges  []JEdge

	schema *schemaImpl
}

func (s *SchemaBuilder) FieldWithDefault(name string, fType JFieldType, defaultValue any) *SchemaBuilder {

	for _, field := range s.fields {
		if field.Name() == name {
			// If a field with the same name already exists, return the schema builder
			// without adding a new field.
			return s
		}
	}

	field := &Field{
		name:         name,
		fType:        fType,
		schema:       s.schema,
		defaultValue: defaultValue,
	}

	s.fields = append(s.fields, field)
	return s
}

func (s *SchemaBuilder) Field(name string, fType JFieldType) *SchemaBuilder {
	return s.FieldWithDefault(name, fType, nil)
}

func (s *SchemaBuilder) Edge(name string, schema JSchema, field JField) *SchemaBuilder {
	for _, edge := range s.edges {
		if edge.Name() == name {
			// If an edge with the same name already exists, return the schema builder
			// without adding a new edge.
			return s
		}
	}

	s.edges = append(s.edges, &Edge{
		name:   name,
		schema: schema,
		field:  field,
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
