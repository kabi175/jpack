package jpack

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
