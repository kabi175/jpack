package jpack

import (
	"context"

	"github.com/samber/mo"
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
	Left() any
	Right() any
	Operator() string

	And() []Filter
	Or() []Filter
}

type SelectField interface {
}

type JQuery interface {
	Select() []SelectField
	Filter() Filter
}

type JRepository interface {
	Create(context.Context, []JRecord) error
	Update(context.Context, []JRecord) error
	First(context.Context, JQuery) (mo.Option[JRecord], error)
	FindAll(context.Context, JQuery) ([]JRecord, error)
	Delete(context.Context, JRecord) error
}
