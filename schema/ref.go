package schema

// jRef implements JRef interface
type jRef struct {
	name         string
	targetSchema JSchema
	isArray      bool
}

// NewJRef creates a new reference
func NewJRef(name string, targetSchema JSchema, isArray bool) JRef {
	return &jRef{
		name:         name,
		targetSchema: targetSchema,
		isArray:      isArray,
	}
}

func (r *jRef) Name() string {
	return r.name
}

func (r *jRef) TargetSchema() JSchema {
	return r.targetSchema
}

func (r *jRef) IsArray() bool {
	return r.isArray
}
