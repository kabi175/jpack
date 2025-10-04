// Package schema provides schema definition and management for JPack.
// It includes immutable schemas, field definitions, relationships, and validation.
//
// The schema package provides a flexible system for defining data structures
// with validation, relationships, and type safety. All schemas are immutable
// by default, ensuring thread safety and preventing accidental modifications.
//
// Example:
//
//	// Create a user schema
//	userSchema := schema.NewSchemaBuilder("User").
//		AddRequiredField("name", schema.JString, nil).
//		AddRequiredUniqueField("email", schema.JString, nil).
//		AddField("age", schema.JInt, 18).
//		AddValidation(func(ctx context.Context, rec schema.JRecord) error {
//			age, ok := rec.Get("age").(int)
//			if ok && age < 18 {
//				return fmt.Errorf("age must be at least 18")
//			}
//			return nil
//		}).
//		Build()
//
//	// Validate a record
//	user := schema.NewJRecord().
//		Set("name", "John Doe").
//		Set("email", "john@example.com").
//		Set("age", 25)
//
//	err := userSchema.Validate(context.Background(), user)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Update schema (creates new instance)
//	updatedSchema := userSchema.Update(func(sb *schema.SchemaBuilder) {
//		sb.AddField("phone", schema.JString, nil)
//	})
//
//	// Create relationships
//	orderSchema := schema.NewSchemaBuilder("Order").
//		AddField("total", schema.JFloat64, 0.0).
//		AddRef("user", userSchema).
//		AddEdge(schema.NewJEdge("user_orders", userSchema, orderSchema, schema.EdgeOneToMany)).
//		Build()
package schema

import (
	"context"
	"fmt"
	"strings"

	"github.com/kabi175/jpack/logger"
)

// jSchema implements JSchema interface (always immutable)
type jSchema struct {
	name        string
	fields      map[string]JField
	refs        map[string]JRef
	edges       []JEdge
	validations []ValidationFunc
	idField     JField
}

// newJSchemaFromBuilder creates a new immutable schema from a builder
func newJSchemaFromBuilder(sb *SchemaBuilder) JSchema {
	s := &jSchema{
		name:        sb.name,
		fields:      make(map[string]JField),
		refs:        make(map[string]JRef),
		edges:       make([]JEdge, len(sb.edges)),
		validations: make([]ValidationFunc, len(sb.validations)),
	}

	// Build fields
	for name, fb := range sb.fields {
		field := newJFieldFromBuilder(&FieldBuilder{
			name:         fb.name,
			fieldType:    fb.fieldType,
			defaultValue: fb.defaultValue,
			required:     fb.required,
			unique:       fb.unique,
			validation:   fb.validation,
		})
		s.fields[name] = field

		// Set ID field
		if fb.name == sb.idFieldName {
			s.idField = field
		}
	}

	// Build refs
	for name, ref := range sb.refs {
		s.refs[name] = NewJRef(ref.name, ref.schema, ref.array)
	}

	// Copy edges
	copy(s.edges, sb.edges)

	// Copy validations
	copy(s.validations, sb.validations)

	logger.Schema.Debug().
		Str("schema", s.name).
		Int("fields_count", len(s.fields)).
		Int("refs_count", len(s.refs)).
		Int("edges_count", len(s.edges)).
		Msg("created immutable schema")

	return s
}

func (s *jSchema) Name() string {
	return s.name
}

func (s *jSchema) Fields() []JField {
	fields := make([]JField, 0, len(s.fields))
	for _, field := range s.fields {
		fields = append(fields, field)
	}
	return fields
}

func (s *jSchema) Field(name string) (JField, bool) {
	field, exists := s.fields[name]
	return field, exists
}

func (s *jSchema) Refs() []JRef {
	refs := make([]JRef, 0, len(s.refs))
	for _, ref := range s.refs {
		refs = append(refs, ref)
	}
	return refs
}

func (s *jSchema) Ref(name string) (JRef, bool) {
	ref, exists := s.refs[name]
	return ref, exists
}

func (s *jSchema) Edges() []JEdge {
	// Return a copy to prevent external modification
	edges := make([]JEdge, len(s.edges))
	copy(edges, s.edges)
	return edges
}

func (s *jSchema) Validate(ctx context.Context, rec JRecord) error {
	// Field-level validations
	for _, field := range s.fields {
		if field.IsRequired() && !rec.Has(field.Name()) {
			return fmt.Errorf("required field '%s' is missing", field.Name())
		}

		if field.Validation() != nil {
			if err := field.Validation()(ctx, rec); err != nil {
				return fmt.Errorf("field '%s' validation failed: %w", field.Name(), err)
			}
		}
	}

	// Schema-level validations
	for _, validation := range s.validations {
		if err := validation(ctx, rec); err != nil {
			return fmt.Errorf("schema validation failed: %w", err)
		}
	}

	return nil
}

func (s *jSchema) Validations() []ValidationFunc {
	// Return a copy to prevent external modification
	validations := make([]ValidationFunc, len(s.validations))
	copy(validations, s.validations)
	return validations
}

func (s *jSchema) GetIDField() JField {
	return s.idField
}

// String returns a string representation of the schema
func (s *jSchema) String() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Schema: %s\n", s.name))
	sb.WriteString("Fields:\n")
	for _, field := range s.fields {
		sb.WriteString(fmt.Sprintf("  - %s: %s", field.Name(), field.Type()))
		if field.IsRequired() {
			sb.WriteString(" (required)")
		}
		if field.IsUnique() {
			sb.WriteString(" (unique)")
		}
		sb.WriteString("\n")
	}

	if len(s.refs) > 0 {
		sb.WriteString("References:\n")
		for _, ref := range s.refs {
			sb.WriteString(fmt.Sprintf("  - %s -> %s", ref.Name(), ref.TargetSchema().Name()))
			if ref.IsArray() {
				sb.WriteString("[]")
			}
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

// Update creates a new schema with modifications applied via callback
// The callback receives a SchemaBuilder initialized with a clone of this schema
func (s *jSchema) Update(fn func(*SchemaBuilder)) JSchema {
	// Create a new builder with a copy of the current schema state
	builder := &SchemaBuilder{
		name:        s.name,
		fields:      make(map[string]*fieldBuilder),
		refs:        make(map[string]*refConfig),
		edges:       make([]JEdge, len(s.edges)),
		validations: make([]ValidationFunc, len(s.validations)),
		idFieldName: "",
	}

	// Copy fields
	for name, field := range s.fields {
		builder.fields[name] = &fieldBuilder{
			name:         field.Name(),
			fieldType:    field.Type(),
			defaultValue: field.DefaultValue(),
			required:     field.IsRequired(),
			unique:       field.IsUnique(),
			validation:   field.Validation(),
		}
	}

	// Set ID field name
	if s.idField != nil {
		builder.idFieldName = s.idField.Name()
	}

	// Copy refs
	for name, ref := range s.refs {
		builder.refs[name] = &refConfig{
			name:   ref.Name(),
			schema: ref.TargetSchema(),
			array:  ref.IsArray(),
		}
	}

	// Copy edges
	copy(builder.edges, s.edges)

	// Copy validations
	copy(builder.validations, s.validations)

	// Apply the modifications
	fn(builder)

	// Build and return the new immutable schema
	return builder.Build()
}

// Clone creates a copy of the schema (for backward compatibility)
func (s *jSchema) Clone() JSchema {
	return s.Update(func(sb *SchemaBuilder) {
		// No modifications, just return a copy
	})
}

// IsImmutable always returns true (for backward compatibility)
func (s *jSchema) IsImmutable() bool {
	return true
}

// Freeze returns the schema itself (for backward compatibility)
func (s *jSchema) Freeze() JSchema {
	logger.Schema.Info().
		Str("schema", s.name).
		Msg("freezing schema to make it immutable")
	return s
}

// Legacy constructor functions for backward compatibility
// These will be deprecated in favor of SchemaBuilder

// NewJSchema creates a new mutable schema (deprecated - use NewSchemaBuilder instead)
// This is kept for backward compatibility with existing code
func NewJSchema(name string) *mutableSchemaAdapter {
	return &mutableSchemaAdapter{
		builder: NewSchemaBuilder(name),
	}
}

// NewImmutableJSchema creates a new immutable schema (deprecated - use NewSchemaBuilder instead)
func NewImmutableJSchema(name string) JSchema {
	return NewSchemaBuilder(name).Build()
}

// mutableSchemaAdapter provides backward compatibility for code that uses the old mutable API
// This adapter allows mutation through method calls but builds an immutable schema when needed
type mutableSchemaAdapter struct {
	builder *SchemaBuilder
	built   JSchema
}

func (m *mutableSchemaAdapter) Name() string {
	if m.built != nil {
		return m.built.Name()
	}
	return m.builder.name
}

func (m *mutableSchemaAdapter) Fields() []JField {
	if m.built != nil {
		return m.built.Fields()
	}
	m.built = m.builder.Build()
	return m.built.Fields()
}

func (m *mutableSchemaAdapter) Field(name string) (JField, bool) {
	if m.built != nil {
		return m.built.Field(name)
	}
	m.built = m.builder.Build()
	return m.built.Field(name)
}

func (m *mutableSchemaAdapter) Refs() []JRef {
	if m.built != nil {
		return m.built.Refs()
	}
	m.built = m.builder.Build()
	return m.built.Refs()
}

func (m *mutableSchemaAdapter) Ref(name string) (JRef, bool) {
	if m.built != nil {
		return m.built.Ref(name)
	}
	m.built = m.builder.Build()
	return m.built.Ref(name)
}

func (m *mutableSchemaAdapter) Edges() []JEdge {
	if m.built != nil {
		return m.built.Edges()
	}
	m.built = m.builder.Build()
	return m.built.Edges()
}

func (m *mutableSchemaAdapter) Validate(ctx context.Context, rec JRecord) error {
	if m.built == nil {
		m.built = m.builder.Build()
	}
	return m.built.Validate(ctx, rec)
}

func (m *mutableSchemaAdapter) Validations() []ValidationFunc {
	if m.built != nil {
		return m.built.Validations()
	}
	m.built = m.builder.Build()
	return m.built.Validations()
}

func (m *mutableSchemaAdapter) GetIDField() JField {
	if m.built == nil {
		m.built = m.builder.Build()
	}
	return m.built.GetIDField()
}

func (m *mutableSchemaAdapter) String() string {
	if m.built == nil {
		m.built = m.builder.Build()
	}
	return m.built.String()
}

func (m *mutableSchemaAdapter) Update(fn func(*SchemaBuilder)) JSchema {
	if m.built == nil {
		m.built = m.builder.Build()
	}
	return m.built.Update(fn)
}

// Freeze returns the built immutable schema (for backward compatibility)
func (m *mutableSchemaAdapter) Freeze() JSchema {
	if m.built == nil {
		m.built = m.builder.Build()
	}
	logger.Schema.Info().
		Str("schema", m.built.Name()).
		Msg("freezing schema to make it immutable")
	return m.built
}

// Clone creates a new mutable adapter (for backward compatibility)
func (m *mutableSchemaAdapter) Clone() JSchema {
	if m.built == nil {
		m.built = m.builder.Build()
	}

	// Create a new builder from the built schema
	newBuilder := &SchemaBuilder{
		name:        m.built.Name(),
		fields:      make(map[string]*fieldBuilder),
		refs:        make(map[string]*refConfig),
		edges:       make([]JEdge, len(m.built.Edges())),
		validations: make([]ValidationFunc, len(m.built.Validations())),
	}

	// Copy fields
	for _, field := range m.built.Fields() {
		newBuilder.fields[field.Name()] = &fieldBuilder{
			name:         field.Name(),
			fieldType:    field.Type(),
			defaultValue: field.DefaultValue(),
			required:     field.IsRequired(),
			unique:       field.IsUnique(),
			validation:   field.Validation(),
		}
	}

	// Set ID field name
	if m.built.GetIDField() != nil {
		newBuilder.idFieldName = m.built.GetIDField().Name()
	}

	// Copy refs
	for _, ref := range m.built.Refs() {
		newBuilder.refs[ref.Name()] = &refConfig{
			name:   ref.Name(),
			schema: ref.TargetSchema(),
			array:  ref.IsArray(),
		}
	}

	// Copy edges
	copy(newBuilder.edges, m.built.Edges())

	// Copy validations
	copy(newBuilder.validations, m.built.Validations())

	return &mutableSchemaAdapter{
		builder: newBuilder,
	}
}

// IsImmutable returns false for adapter (backward compatibility)
func (m *mutableSchemaAdapter) IsImmutable() bool {
	return false
}
