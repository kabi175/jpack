package jpack

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

type JRecord interface {
	Schema() JSchema

	Value(JField) (any, bool)
	SetValue(field JField, value any) error

	Fields() []JField

	IsModified() bool
	IsNew() bool
	DirtyKeys() []string

	Save(ctx context.Context) error
	Validate() error
}

type Filter interface {
	Value() any
	Field() JField

	Left() Filter
	Right() Filter
	Operator() string

	And(Filter) Filter
	Or(Filter) Filter
	Not() Filter
}

// NewQuery creates a new query for the given schema and context
// This is a convenience function that returns the appropriate query implementation
// based on the context (MongoDB connection)
func NewQuery(ctx context.Context, schema JSchema) Query {
	// Check if MongoDB connection is available in context
	if _, ok := ctx.Value(Conn).(*mongo.Database); ok {
		return NewMongoQuery(ctx, schema)
	}

	// For now, only MongoDB is supported
	// In the future, this could support other databases
	panic("jpack: no supported database connection found in context")
}
