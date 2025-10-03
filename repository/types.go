package repository

import (
	"context"
)

// CrudRepository defines basic CRUD operations
type CrudRepository interface {
	Save(ctx context.Context, rec JRecord) (JRecord, error)
	FindById(ctx context.Context, id any) (JRecord, error)
	Delete(ctx context.Context, id any) error
	FindAll(ctx context.Context) ([]JRecord, error)
	Exists(ctx context.Context, id any) (bool, error)
}

// JpaRepository extends CrudRepository with advanced querying capabilities
type JpaRepository interface {
	CrudRepository
	FindBy(ctx context.Context, criteria JCriteria) ([]JRecord, error)
	Count(ctx context.Context, criteria JCriteria) (int64, error)
	Update(ctx context.Context, criteria JCriteria, updates JRecord) error
	DeleteBy(ctx context.Context, criteria JCriteria) error
	FindOne(ctx context.Context, criteria JCriteria) (JRecord, error)
}

// JRecord is imported from schema package
type JRecord interface {
	Get(key string) any
	Set(key string, value any) JRecord
	Has(key string) bool
	Delete(key string) JRecord
	Keys() []string
	ToMap() map[string]any
	FromMap(data map[string]any) JRecord
	Clone() JRecord
}

// JCriteria represents query criteria
type JCriteria struct {
	Filters    []JFilter
	Sort       []JSort
	Limit      int64
	Offset     int64
	Projection []string // Fields to include/exclude
}

// JFilter represents a query filter
type JFilter struct {
	Field string
	Op    JOperator
	Value any
}

// JSort represents sorting criteria
type JSort struct {
	Field     string
	Direction SortDirection
}

// SortDirection represents sort direction
type SortDirection string

const (
	SortAsc  SortDirection = "asc"
	SortDesc SortDirection = "desc"
)

// JOperator represents a query operator
type JOperator interface {
	Name() string
	Apply(field string, value any) (any, error)
}

// JProjection represents field projection
type JProjection struct {
	Fields []string
	Mode   ProjectionMode
}

// ProjectionMode represents projection mode
type ProjectionMode string

const (
	ProjectionInclude ProjectionMode = "include"
	ProjectionExclude ProjectionMode = "exclude"
)

// JAggregation represents aggregation operations
type JAggregation struct {
	Type     AggregationType
	Field    string
	Alias    string
	Pipeline []JAggregationStage
}

// AggregationType represents aggregation operation type
type AggregationType string

const (
	AggCount    AggregationType = "count"
	AggSum      AggregationType = "sum"
	AggAvg      AggregationType = "avg"
	AggMin      AggregationType = "min"
	AggMax      AggregationType = "max"
	AggGroupBy  AggregationType = "groupBy"
	AggPipeline AggregationType = "pipeline"
)

// JAggregationStage represents a stage in aggregation pipeline
type JAggregationStage struct {
	Stage string
	Data  map[string]any
}
