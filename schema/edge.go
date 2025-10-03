package schema

// jEdge implements JEdge interface
type jEdge struct {
	name       string
	fromSchema JSchema
	toSchema   JSchema
	edgeType   EdgeType
}

// NewJEdge creates a new edge
func NewJEdge(name string, fromSchema, toSchema JSchema, edgeType EdgeType) JEdge {
	return &jEdge{
		name:       name,
		fromSchema: fromSchema,
		toSchema:   toSchema,
		edgeType:   edgeType,
	}
}

func (e *jEdge) Name() string {
	return e.name
}

func (e *jEdge) FromSchema() JSchema {
	return e.fromSchema
}

func (e *jEdge) ToSchema() JSchema {
	return e.toSchema
}

func (e *jEdge) Type() EdgeType {
	return e.edgeType
}
