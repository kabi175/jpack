// Package repository provides the repository pattern implementation for JPack.
// It includes CRUD operations, query building, and advanced querying capabilities.
//
// The repository pattern provides a clean abstraction over data access,
// making it easy to perform database operations without dealing with low-level details.
//
// Example:
//
//	// Get repository from client
//	userRepo, err := client.GetRepository("User")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Create and save record
//	user := schema.NewJRecord().
//		Set("id", "user_001").
//		Set("name", "John Doe").
//		Set("email", "john@example.com").
//		Set("age", 25)
//
//	savedUser, err := userRepo.Save(context.Background(), user)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Find by ID
//	foundUser, err := userRepo.FindById(context.Background(), "user_001")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Advanced querying with criteria
//	criteria := repository.NewJCriteriaBuilder().
//		Where("age", repository.OpGTE, 18).
//		Where("status", repository.OpEQ, "active").
//		OrderBy("name").
//		SetLimit(10).
//		Build()
//
//	users, err := userRepo.FindBy(context.Background(), criteria)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Count records
//	count, err := userRepo.Count(context.Background(), criteria)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Update multiple records
//	updateCriteria := repository.NewJCriteriaBuilder().
//		Where("status", repository.OpEQ, "inactive").
//		Build()
//
//	updates := schema.NewJRecord().Set("status", "active")
//	err = userRepo.Update(context.Background(), updateCriteria, updates)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Delete multiple records
//	err = userRepo.DeleteBy(context.Background(), updateCriteria)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Use projections to select specific fields
//	proj := repository.NewIncludeProjection("name", "email")
//	projectedUsers, err := userRepo.FindBy(context.Background(), repository.JCriteria{
//		Projection: proj.Fields,
//	})
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Use aggregations for analytics
//	agg := repository.NewJAggregation(repository.AggTypeCount, "age")
//	results, err := userRepo.Aggregate(context.Background(), repository.JCriteria{
//		Aggregations: []repository.JAggregation{agg},
//	})
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Complex queries with multiple operators
//	complexCriteria := repository.NewJCriteriaBuilder().
//		Where("age", repository.OpGTE, 18).
//		Where("age", repository.OpLT, 65).
//		Where("status", repository.OpIN, []any{"active", "pending"}).
//		Where("email", repository.OpNE, nil).
//		OrderBy("created_at", repository.SortDesc).
//		SetLimit(50).
//		SetOffset(10).
//		Build()
//
//	complexUsers, err := userRepo.FindBy(context.Background(), complexCriteria)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Find single record
//	singleUser, err := userRepo.FindOne(context.Background(), criteria)
//	if err != nil {
//		log.Fatal(err)
//	}
package repository

import (
	"context"
)

// CrudRepository defines basic CRUD operations.
// It provides the fundamental operations for creating, reading, updating, and deleting records.
//
// Example:
//
//	userRepo, err := client.GetRepository("User")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Create
//	user := schema.NewJRecord().Set("id", "user_001").Set("name", "John")
//	savedUser, err := userRepo.Save(context.Background(), user)
//
//	// Read
//	foundUser, err := userRepo.FindById(context.Background(), "user_001")
//	allUsers, err := userRepo.FindAll(context.Background())
//
//	// Check existence
//	exists, err := userRepo.Exists(context.Background(), "user_001")
//
//	// Delete
//	err = userRepo.Delete(context.Background(), "user_001")
type CrudRepository interface {
	// Save creates or updates a record
	Save(ctx context.Context, rec JRecord) (JRecord, error)
	// FindById retrieves a record by its ID
	FindById(ctx context.Context, id any) (JRecord, error)
	// Delete removes a record by its ID
	Delete(ctx context.Context, id any) error
	// FindAll retrieves all records
	FindAll(ctx context.Context) ([]JRecord, error)
	// Exists checks if a record exists by its ID
	Exists(ctx context.Context, id any) (bool, error)
}

// JpaRepository extends CrudRepository with advanced querying capabilities.
// It provides sophisticated querying, counting, updating, and deletion operations.
//
// Example:
//
//	userRepo, err := client.GetRepository("User")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Advanced querying
//	criteria := repository.NewJCriteriaBuilder().
//		Where("age", repository.OpGTE, 18).
//		Where("status", repository.OpEQ, "active").
//		OrderBy("name").
//		SetLimit(10).
//		Build()
//
//	users, err := userRepo.FindBy(context.Background(), criteria)
//	count, err := userRepo.Count(context.Background(), criteria)
//
//	// Update multiple records
//	updateCriteria := repository.NewJCriteriaBuilder().
//		Where("status", repository.OpEQ, "inactive").
//		Build()
//
//	updates := schema.NewJRecord().Set("status", "active")
//	err = userRepo.Update(context.Background(), updateCriteria, updates)
//
//	// Delete multiple records
//	err = userRepo.DeleteBy(context.Background(), updateCriteria)
//
//	// Find single record
//	user, err := userRepo.FindOne(context.Background(), criteria)
type JpaRepository interface {
	CrudRepository
	// FindBy retrieves records matching the criteria
	FindBy(ctx context.Context, criteria JCriteria) ([]JRecord, error)
	// Count returns the number of records matching the criteria
	Count(ctx context.Context, criteria JCriteria) (int64, error)
	// Update updates records matching the criteria
	Update(ctx context.Context, criteria JCriteria, updates JRecord) error
	// DeleteBy deletes records matching the criteria
	DeleteBy(ctx context.Context, criteria JCriteria) error
	// FindOne retrieves a single record matching the criteria
	FindOne(ctx context.Context, criteria JCriteria) (JRecord, error)
}

// JRecord represents a data record.
// It provides methods for accessing and manipulating record data.
//
// Example:
//
//	record := schema.NewJRecord().
//		Set("id", "user_001").
//		Set("name", "John Doe").
//		Set("email", "john@example.com")
//
//	name := record.Get("name").(string)
//	hasEmail := record.Has("email")
//	recordMap := record.ToMap()
//	clonedRecord := record.Clone()
type JRecord interface {
	// Get retrieves a value by key
	Get(key string) any
	// Set sets a value for a key and returns a new record
	Set(key string, value any) JRecord
	// Has checks if a key exists in the record
	Has(key string) bool
	// Delete removes a key and returns a new record
	Delete(key string) JRecord
	// Keys returns all keys in the record
	Keys() []string
	// ToMap converts the record to a map
	ToMap() map[string]any
	// FromMap creates a record from a map
	FromMap(data map[string]any) JRecord
	// Clone creates a copy of the record
	Clone() JRecord
}

// JCriteria represents query criteria.
// It defines the conditions, sorting, pagination, and projection for queries.
//
// Example:
//
//	criteria := repository.JCriteria{
//		Filters: []repository.JFilter{
//			{Field: "age", Op: repository.OpGTE, Value: 18},
//			{Field: "status", Op: repository.OpEQ, Value: "active"},
//		},
//		Sort: []repository.JSort{
//			{Field: "name", Direction: repository.SortAsc},
//			{Field: "created_at", Direction: repository.SortDesc},
//		},
//		Limit:  10,
//		Offset: 0,
//		Projection: []string{"name", "email"},
//	}
type JCriteria struct {
	// Filters defines the query conditions
	Filters []JFilter
	// Sort defines the sorting order
	Sort []JSort
	// Limit defines the maximum number of results
	Limit int64
	// Offset defines the number of results to skip
	Offset int64
	// Projection defines fields to include/exclude
	Projection []string
}

// JFilter represents a query filter.
// It defines a single condition for filtering records.
//
// Example:
//
//	filter := repository.JFilter{
//		Field: "age",
//		Op:    repository.OpGTE,
//		Value: 18,
//	}
type JFilter struct {
	// Field is the field name to filter on
	Field string
	// Op is the operator to use for comparison
	Op JOperator
	// Value is the value to compare against
	Value any
}

// JSort represents sorting criteria.
// It defines how to sort query results.
//
// Example:
//
//	sort := repository.JSort{
//		Field:     "name",
//		Direction: repository.SortAsc,
//	}
type JSort struct {
	// Field is the field name to sort by
	Field string
	// Direction is the sort direction
	Direction SortDirection
}

// SortDirection represents sort direction.
type SortDirection string

const (
	// SortAsc sorts in ascending order
	SortAsc SortDirection = "asc"
	// SortDesc sorts in descending order
	SortDesc SortDirection = "desc"
)

// JOperator represents a query operator.
// It defines how to apply a filter condition to a field.
//
// Example:
//
//	operator := repository.OpGTE
//	result, err := operator.Apply("age", 18)
type JOperator interface {
	// Name returns the operator name
	Name() string
	// Apply applies the operator to a field and value
	Apply(field string, value any) (any, error)
}

// JProjection represents field projection.
// It defines which fields to include or exclude in query results.
//
// Example:
//
//	projection := repository.JProjection{
//		Fields: []string{"name", "email"},
//		Mode:   repository.ProjectionInclude,
//	}
type JProjection struct {
	// Fields is the list of fields to project
	Fields []string
	// Mode determines whether to include or exclude fields
	Mode ProjectionMode
}

// ProjectionMode represents projection mode.
type ProjectionMode string

const (
	// ProjectionInclude includes only the specified fields
	ProjectionInclude ProjectionMode = "include"
	// ProjectionExclude excludes the specified fields
	ProjectionExclude ProjectionMode = "exclude"
)

// JAggregation represents aggregation operations.
// It defines how to aggregate data across multiple records.
//
// Example:
//
//	aggregation := repository.JAggregation{
//		Type:  repository.AggCount,
//		Field: "id",
//		Alias: "total_users",
//	}
type JAggregation struct {
	// Type is the aggregation operation type
	Type AggregationType
	// Field is the field to aggregate on
	Field string
	// Alias is the name for the aggregated result
	Alias string
	// Pipeline is the aggregation pipeline stages
	Pipeline []JAggregationStage
}

// AggregationType represents aggregation operation type.
type AggregationType string

const (
	// AggCount counts the number of records
	AggCount AggregationType = "count"
	// AggSum sums the values of a field
	AggSum AggregationType = "sum"
	// AggAvg calculates the average of a field
	AggAvg AggregationType = "avg"
	// AggMin finds the minimum value of a field
	AggMin AggregationType = "min"
	// AggMax finds the maximum value of a field
	AggMax AggregationType = "max"
	// AggGroupBy groups records by a field
	AggGroupBy AggregationType = "groupBy"
	// AggPipeline executes a custom aggregation pipeline
	AggPipeline AggregationType = "pipeline"
)

// JAggregationStage represents a stage in aggregation pipeline.
// It defines a single stage in a MongoDB aggregation pipeline.
//
// Example:
//
//	stage := repository.JAggregationStage{
//		Stage: "$match",
//		Data: map[string]any{
//			"age": map[string]any{"$gte": 18},
//		},
//	}
type JAggregationStage struct {
	// Stage is the aggregation stage name
	Stage string
	// Data is the stage configuration
	Data map[string]any
}
