package schema

import (
	"context"
	"fmt"
	"strings"

	"github.com/kabi175/jpack/logger"
)

// jSchema implements JSchema interface
type jSchema struct {
	name        string
	fields      map[string]JField
	refs        map[string]JRef
	edges       []JEdge
	validations []ValidationFunc
	idField     JField
	immutable   bool
}

// NewJSchema creates a new schema
func NewJSchema(name string) JSchema {
	return &jSchema{
		name:        name,
		fields:      make(map[string]JField),
		refs:        make(map[string]JRef),
		edges:       make([]JEdge, 0),
		validations: make([]ValidationFunc, 0),
		immutable:   false,
	}
}

// NewImmutableJSchema creates a new immutable schema
func NewImmutableJSchema(name string) JSchema {
	return &jSchema{
		name:        name,
		fields:      make(map[string]JField),
		refs:        make(map[string]JRef),
		edges:       make([]JEdge, 0),
		validations: make([]ValidationFunc, 0),
		immutable:   true,
	}
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

func (s *jSchema) AddField(name string, fType JFieldType, defaultValue any) JField {
	if s.immutable {
		logger.Schema.Error().
			Str("schema", s.name).
			Str("field", name).
			Msg("attempted to modify immutable schema")
		panic("cannot modify immutable schema: use ReplaceSchema to create a new schema")
	}

	logger.Schema.Debug().
		Str("schema", s.name).
		Str("field", name).
		Str("type", string(fType)).
		Msg("adding field to schema")

	field := NewJField(name, fType, defaultValue)
	s.fields[name] = field

	// If this is the first field and no ID field is set, make it the ID field
	if s.idField == nil {
		s.idField = field
		logger.Schema.Debug().
			Str("schema", s.name).
			Str("field", name).
			Msg("set as ID field")
	}

	return field
}

func (s *jSchema) AddRef(name string, schema JSchema) JRef {
	if s.immutable {
		logger.Schema.Error().
			Str("schema", s.name).
			Str("ref", name).
			Msg("attempted to modify immutable schema")
		panic("cannot modify immutable schema: use ReplaceSchema to create a new schema")
	}

	logger.Schema.Debug().
		Str("schema", s.name).
		Str("ref", name).
		Str("target_schema", schema.Name()).
		Msg("adding reference to schema")

	ref := NewJRef(name, schema, false)
	s.refs[name] = ref
	return ref
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
	return s.edges
}

func (s *jSchema) AddEdge(edge JEdge) JSchema {
	if s.immutable {
		logger.Schema.Error().
			Str("schema", s.name).
			Str("edge", edge.Name()).
			Msg("attempted to modify immutable schema")
		panic("cannot modify immutable schema: use ReplaceSchema to create a new schema")
	}

	logger.Schema.Debug().
		Str("schema", s.name).
		Str("edge", edge.Name()).
		Msg("adding edge to schema")

	s.edges = append(s.edges, edge)
	return s
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

func (s *jSchema) AddValidation(fn ValidationFunc) {
	if s.immutable {
		logger.Schema.Error().
			Str("schema", s.name).
			Msg("attempted to modify immutable schema")
		panic("cannot modify immutable schema: use ReplaceSchema to create a new schema")
	}

	logger.Schema.Debug().
		Str("schema", s.name).
		Msg("adding validation to schema")

	s.validations = append(s.validations, fn)
}

func (s *jSchema) Validations() []ValidationFunc {
	return s.validations
}

func (s *jSchema) GetIDField() JField {
	return s.idField
}

func (s *jSchema) SetIDField(field JField) JSchema {
	if s.immutable {
		logger.Schema.Error().
			Str("schema", s.name).
			Str("field", field.Name()).
			Msg("attempted to modify immutable schema")
		panic("cannot modify immutable schema: use ReplaceSchema to create a new schema")
	}

	logger.Schema.Debug().
		Str("schema", s.name).
		Str("field", field.Name()).
		Msg("setting ID field")

	s.idField = field
	return s
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

// IsImmutable returns whether the schema is immutable
func (s *jSchema) IsImmutable() bool {
	return s.immutable
}

// Clone creates a mutable copy of the schema
func (s *jSchema) Clone() JSchema {
	clone := &jSchema{
		name:        s.name,
		fields:      make(map[string]JField),
		refs:        make(map[string]JRef),
		edges:       make([]JEdge, len(s.edges)),
		validations: make([]ValidationFunc, len(s.validations)),
		idField:     nil,   // Will be set during field copying
		immutable:   false, // Clone is always mutable
	}

	// Copy fields (create new mutable copies)
	for name, field := range s.fields {
		// Create a new field with the same properties but mutable
		newField := NewJField(field.Name(), field.Type(), field.DefaultValue())
		if field.IsRequired() {
			newField.SetRequired(true)
		}
		if field.IsUnique() {
			newField.SetUnique(true)
		}
		if field.Validation() != nil {
			newField.SetValidation(field.Validation())
		}
		clone.fields[name] = newField

		// Set the ID field if this was the original ID field
		if s.idField != nil && field.Name() == s.idField.Name() {
			clone.idField = newField
		}
	}

	// Copy refs
	for name, ref := range s.refs {
		clone.refs[name] = ref
	}

	// Copy edges
	copy(clone.edges, s.edges)

	// Copy validations
	copy(clone.validations, s.validations)

	return clone
}

// Freeze makes the schema immutable
func (s *jSchema) Freeze() JSchema {
	logger.Schema.Info().
		Str("schema", s.name).
		Msg("freezing schema to make it immutable")

	s.immutable = true

	// Also freeze all fields
	for _, field := range s.fields {
		field.Freeze()
	}

	logger.Schema.Debug().
		Str("schema", s.name).
		Int("fields_count", len(s.fields)).
		Msg("schema frozen successfully")

	return s
}
