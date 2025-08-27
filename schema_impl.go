package jpack

import (
	"context"
	"fmt"
)

type schemaImpl struct {
	name        string
	fields      []JField
	edges       []JEdge
	validations []ValidationFunc
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
func (s *schemaImpl) AddField(name string, fType JFieldType, defaultValue any) JField {
	field := &fieldImpl{
		name:         name,
		fType:        fType,
		schema:       s,
		defaultValue: defaultValue,
	}

	return s.appendFieldIfNotPresent(field)
}

func (s *schemaImpl) AddRef(name string, schema JSchema) JRef {
	field := &refImpl{
		fieldImpl: fieldImpl{
			name:   name,
			fType:  &Ref{},
			schema: s,
		},
		relSchema: schema,
	}

	if field == s.appendFieldIfNotPresent(field) {
		return field
	}
	return nil
}

func (s *schemaImpl) appendFieldIfNotPresent(field JField) JField {
	for _, f := range s.fields {
		if f.Name() == field.Name() {
			// If a field with the same name already exists, return the schema builder
			// without adding a new field.
			return nil
		}
	}

	s.fields = append(s.fields, field)
	return field
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

// AddValidation adds a custom validation function to the schema
func (s *schemaImpl) AddValidation(validation ValidationFunc) {
	s.validations = append(s.validations, validation)
}

// Validations returns all custom validation functions
func (s *schemaImpl) Validations() []ValidationFunc {
	return s.validations
}

// Validate implements JSchema.
func (s *schemaImpl) Validate(ctx context.Context, record JRecord) error {
	// Validate that the record's schema matches this schema
	if record.Schema() != s {
		return fmt.Errorf("record schema does not match validation schema")
	}

	// Validate each field in the schema
	for _, field := range s.fields {
		value, exists := record.Value(field)

		// If field doesn't exist and has no default, it's required
		if !exists {
			if field.Default() == nil {
				return fmt.Errorf("required field '%s' is missing", field.Name())
			}
			// Use default value for validation
			value = field.Default()
		}

		// Validate the field value against its type
		if err := field.Type().Validate(value); err != nil {
			return fmt.Errorf("field '%s' validation failed: %w", field.Name(), err)
		}
	}

	// Run custom validations
	for i, validation := range s.validations {
		if err := validation(ctx, record); err != nil {
			return fmt.Errorf("custom validation %d failed: %w", i+1, err)
		}
	}

	return nil
}

var _ JSchema = &schemaImpl{}

type edgeImpl struct {
	name   string
	schema JSchema
	field  JRef
}

// Ref implements JEdge.
func (e *edgeImpl) Ref() JRef {
	return e.field
}

// Name implements JEdge.
func (e *edgeImpl) Name() string {
	return e.name
}

// Schema implements JEdge.
func (e *edgeImpl) Schema() JSchema {
	return e.schema
}

var _ JEdge = &edgeImpl{}

type fieldImpl struct {
	name         string
	fType        JFieldType
	schema       JSchema
	defaultValue any
}

// Default implements JField.
func (f *fieldImpl) Default() any {
	return f.defaultValue
}

// Name implements JField.
func (f *fieldImpl) Name() string {
	return f.name
}

// Schema implements JField.
func (f *fieldImpl) Schema() JSchema {
	return f.schema
}

// Type implements JField.
func (f *fieldImpl) Type() JFieldType {
	return f.fType
}

var _ JField = &fieldImpl{}

type refImpl struct {
	fieldImpl
	relSchema JSchema
}

func (r *refImpl) RelSchema() JSchema {
	return r.relSchema
}

var _ JRef = &refImpl{}
