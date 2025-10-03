package projection

import (
	"context"
	"fmt"

	"github.com/kabi175/jpack/logger"
	"github.com/kabi175/jpack/repository"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// AggregationPipeline represents a MongoDB aggregation pipeline
type AggregationPipeline struct {
	stages []bson.M
}

// NewAggregationPipeline creates a new aggregation pipeline
func NewAggregationPipeline() *AggregationPipeline {
	return &AggregationPipeline{
		stages: make([]bson.M, 0),
	}
}

// Match adds a $match stage
func (ap *AggregationPipeline) Match(filter bson.M) *AggregationPipeline {
	ap.stages = append(ap.stages, bson.M{"$match": filter})
	return ap
}

// Group adds a $group stage
func (ap *AggregationPipeline) Group(group bson.M) *AggregationPipeline {
	ap.stages = append(ap.stages, bson.M{"$group": group})
	return ap
}

// Sort adds a $sort stage
func (ap *AggregationPipeline) Sort(sort bson.M) *AggregationPipeline {
	ap.stages = append(ap.stages, bson.M{"$sort": sort})
	return ap
}

// Limit adds a $limit stage
func (ap *AggregationPipeline) Limit(limit int64) *AggregationPipeline {
	ap.stages = append(ap.stages, bson.M{"$limit": limit})
	return ap
}

// Skip adds a $skip stage
func (ap *AggregationPipeline) Skip(skip int64) *AggregationPipeline {
	ap.stages = append(ap.stages, bson.M{"$skip": skip})
	return ap
}

// Project adds a $project stage
func (ap *AggregationPipeline) Project(projection bson.M) *AggregationPipeline {
	ap.stages = append(ap.stages, bson.M{"$project": projection})
	return ap
}

// Lookup adds a $lookup stage
func (ap *AggregationPipeline) Lookup(from, localField, foreignField, as string) *AggregationPipeline {
	ap.stages = append(ap.stages, bson.M{
		"$lookup": bson.M{
			"from":         from,
			"localField":   localField,
			"foreignField": foreignField,
			"as":           as,
		},
	})
	return ap
}

// Unwind adds an $unwind stage
func (ap *AggregationPipeline) Unwind(path string) *AggregationPipeline {
	ap.stages = append(ap.stages, bson.M{"$unwind": "$" + path})
	return ap
}

// AddFields adds an $addFields stage
func (ap *AggregationPipeline) AddFields(fields bson.M) *AggregationPipeline {
	ap.stages = append(ap.stages, bson.M{"$addFields": fields})
	return ap
}

// Count adds a $count stage
func (ap *AggregationPipeline) Count(alias string) *AggregationPipeline {
	ap.stages = append(ap.stages, bson.M{"$count": alias})
	return ap
}

// Sum adds a $sum stage
func (ap *AggregationPipeline) Sum(field, alias string) *AggregationPipeline {
	ap.stages = append(ap.stages, bson.M{
		"$group": bson.M{
			"_id": nil,
			alias: bson.M{"$sum": "$" + field},
		},
	})
	return ap
}

// Avg adds an $avg stage
func (ap *AggregationPipeline) Avg(field, alias string) *AggregationPipeline {
	ap.stages = append(ap.stages, bson.M{
		"$group": bson.M{
			"_id": nil,
			alias: bson.M{"$avg": "$" + field},
		},
	})
	return ap
}

// Min adds a $min stage
func (ap *AggregationPipeline) Min(field, alias string) *AggregationPipeline {
	ap.stages = append(ap.stages, bson.M{
		"$group": bson.M{
			"_id": nil,
			alias: bson.M{"$min": "$" + field},
		},
	})
	return ap
}

// Max adds a $max stage
func (ap *AggregationPipeline) Max(field, alias string) *AggregationPipeline {
	ap.stages = append(ap.stages, bson.M{
		"$group": bson.M{
			"_id": nil,
			alias: bson.M{"$max": "$" + field},
		},
	})
	return ap
}

// GroupBy adds a $group stage for grouping
func (ap *AggregationPipeline) GroupBy(field, alias string) *AggregationPipeline {
	ap.stages = append(ap.stages, bson.M{
		"$group": bson.M{
			"_id": "$" + field,
			alias: bson.M{"$push": "$$ROOT"},
		},
	})
	return ap
}

// AddStage adds a custom stage
func (ap *AggregationPipeline) AddStage(stage string, data bson.M) *AggregationPipeline {
	ap.stages = append(ap.stages, bson.M{stage: data})
	return ap
}

// Build builds the pipeline
func (ap *AggregationPipeline) Build() []bson.M {
	return ap.stages
}

// String returns a string representation of the pipeline
func (ap *AggregationPipeline) String() string {
	logger.Projection.Debug().
		Int("stages", len(ap.stages)).
		Msg("aggregation pipeline string representation")
	return fmt.Sprintf("AggregationPipeline{stages: %d}", len(ap.stages))
}

// AggregationExecutor executes aggregation pipelines
type AggregationExecutor interface {
	Execute(ctx context.Context, pipeline []bson.M) (*mongo.Cursor, error)
	ExecuteAndDecode(ctx context.Context, pipeline []bson.M, result interface{}) error
}

// jAggregationExecutor implements AggregationExecutor
type jAggregationExecutor struct {
	collection *mongo.Collection
}

// NewAggregationExecutor creates a new aggregation executor
func NewAggregationExecutor(collection *mongo.Collection) AggregationExecutor {
	return &jAggregationExecutor{
		collection: collection,
	}
}

// Execute executes the aggregation pipeline
func (ae *jAggregationExecutor) Execute(ctx context.Context, pipeline []bson.M) (*mongo.Cursor, error) {
	return ae.collection.Aggregate(ctx, pipeline)
}

// ExecuteAndDecode executes the aggregation pipeline and decodes results
func (ae *jAggregationExecutor) ExecuteAndDecode(ctx context.Context, pipeline []bson.M, result interface{}) error {
	cursor, err := ae.Execute(ctx, pipeline)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	return cursor.All(ctx, result)
}

// AggregationService provides aggregation services
type AggregationService interface {
	Count(ctx context.Context, criteria repository.JCriteria) (int64, error)
	Sum(ctx context.Context, criteria repository.JCriteria, field string) (float64, error)
	Avg(ctx context.Context, criteria repository.JCriteria, field string) (float64, error)
	Min(ctx context.Context, criteria repository.JCriteria, field string) (interface{}, error)
	Max(ctx context.Context, criteria repository.JCriteria, field string) (interface{}, error)
	GroupBy(ctx context.Context, criteria repository.JCriteria, groupField string) ([]bson.M, error)
	ExecutePipeline(ctx context.Context, pipeline []bson.M) ([]bson.M, error)
}

// jAggregationService implements AggregationService
type jAggregationService struct {
	executor AggregationExecutor
}

// NewAggregationService creates a new aggregation service
func NewAggregationService(executor AggregationExecutor) AggregationService {
	return &jAggregationService{
		executor: executor,
	}
}

// Count counts records
func (as *jAggregationService) Count(ctx context.Context, criteria repository.JCriteria) (int64, error) {
	// Build match stage from criteria
	queryBuilder := repository.NewQueryBuilder(criteria)
	filter, err := queryBuilder.BuildFilter()
	if err != nil {
		return 0, err
	}

	// Build aggregation pipeline
	pipeline := NewAggregationPipeline().
		Match(filter).
		Count("count")

	// Execute aggregation
	var result []bson.M
	err = as.executor.ExecuteAndDecode(ctx, pipeline.Build(), &result)
	if err != nil {
		return 0, err
	}

	if len(result) == 0 {
		return 0, nil
	}

	count, ok := result[0]["count"].(int32)
	if !ok {
		logger.Projection.Error().
			Interface("result", result[0]["count"]).
			Msg("unexpected count result type")
		return 0, fmt.Errorf("unexpected count result type")
	}

	return int64(count), nil
}

// Sum calculates sum of a field
func (as *jAggregationService) Sum(ctx context.Context, criteria repository.JCriteria, field string) (float64, error) {
	// Build match stage from criteria
	queryBuilder := repository.NewQueryBuilder(criteria)
	filter, err := queryBuilder.BuildFilter()
	if err != nil {
		return 0, err
	}

	// Build aggregation pipeline
	pipeline := NewAggregationPipeline().
		Match(filter).
		Sum(field, "sum")

	// Execute aggregation
	var result []bson.M
	err = as.executor.ExecuteAndDecode(ctx, pipeline.Build(), &result)
	if err != nil {
		return 0, err
	}

	if len(result) == 0 {
		return 0, nil
	}

	sum, ok := result[0]["sum"].(float64)
	if !ok {
		logger.Projection.Error().
			Interface("result", result[0]["sum"]).
			Msg("unexpected sum result type")
		return 0, fmt.Errorf("unexpected sum result type")
	}

	return sum, nil
}

// Avg calculates average of a field
func (as *jAggregationService) Avg(ctx context.Context, criteria repository.JCriteria, field string) (float64, error) {
	// Build match stage from criteria
	queryBuilder := repository.NewQueryBuilder(criteria)
	filter, err := queryBuilder.BuildFilter()
	if err != nil {
		return 0, err
	}

	// Build aggregation pipeline
	pipeline := NewAggregationPipeline().
		Match(filter).
		Avg(field, "avg")

	// Execute aggregation
	var result []bson.M
	err = as.executor.ExecuteAndDecode(ctx, pipeline.Build(), &result)
	if err != nil {
		return 0, err
	}

	if len(result) == 0 {
		return 0, nil
	}

	avg, ok := result[0]["avg"].(float64)
	if !ok {
		logger.Projection.Error().
			Interface("result", result[0]["avg"]).
			Msg("unexpected avg result type")
		return 0, fmt.Errorf("unexpected avg result type")
	}

	return avg, nil
}

// Min finds minimum value of a field
func (as *jAggregationService) Min(ctx context.Context, criteria repository.JCriteria, field string) (interface{}, error) {
	// Build match stage from criteria
	queryBuilder := repository.NewQueryBuilder(criteria)
	filter, err := queryBuilder.BuildFilter()
	if err != nil {
		return nil, err
	}

	// Build aggregation pipeline
	pipeline := NewAggregationPipeline().
		Match(filter).
		Min(field, "min")

	// Execute aggregation
	var result []bson.M
	err = as.executor.ExecuteAndDecode(ctx, pipeline.Build(), &result)
	if err != nil {
		return nil, err
	}

	if len(result) == 0 {
		return nil, nil
	}

	return result[0]["min"], nil
}

// Max finds maximum value of a field
func (as *jAggregationService) Max(ctx context.Context, criteria repository.JCriteria, field string) (interface{}, error) {
	// Build match stage from criteria
	queryBuilder := repository.NewQueryBuilder(criteria)
	filter, err := queryBuilder.BuildFilter()
	if err != nil {
		return nil, err
	}

	// Build aggregation pipeline
	pipeline := NewAggregationPipeline().
		Match(filter).
		Max(field, "max")

	// Execute aggregation
	var result []bson.M
	err = as.executor.ExecuteAndDecode(ctx, pipeline.Build(), &result)
	if err != nil {
		return nil, err
	}

	if len(result) == 0 {
		return nil, nil
	}

	return result[0]["max"], nil
}

// GroupBy groups records by a field
func (as *jAggregationService) GroupBy(ctx context.Context, criteria repository.JCriteria, groupField string) ([]bson.M, error) {
	// Build match stage from criteria
	queryBuilder := repository.NewQueryBuilder(criteria)
	filter, err := queryBuilder.BuildFilter()
	if err != nil {
		return nil, err
	}

	// Build aggregation pipeline
	pipeline := NewAggregationPipeline().
		Match(filter).
		GroupBy(groupField, "records")

	// Execute aggregation
	var result []bson.M
	err = as.executor.ExecuteAndDecode(ctx, pipeline.Build(), &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// ExecutePipeline executes a custom aggregation pipeline
func (as *jAggregationService) ExecutePipeline(ctx context.Context, pipeline []bson.M) ([]bson.M, error) {
	var result []bson.M
	err := as.executor.ExecuteAndDecode(ctx, pipeline, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// AggregationBuilder provides a fluent interface for building aggregations
type AggregationBuilder struct {
	pipeline *AggregationPipeline
}

// NewAggregationBuilder creates a new aggregation builder
func NewAggregationBuilder() *AggregationBuilder {
	return &AggregationBuilder{
		pipeline: NewAggregationPipeline(),
	}
}

// Match adds a $match stage
func (ab *AggregationBuilder) Match(filter bson.M) *AggregationBuilder {
	ab.pipeline.Match(filter)
	return ab
}

// Group adds a $group stage
func (ab *AggregationBuilder) Group(group bson.M) *AggregationBuilder {
	ab.pipeline.Group(group)
	return ab
}

// Sort adds a $sort stage
func (ab *AggregationBuilder) Sort(sort bson.M) *AggregationBuilder {
	ab.pipeline.Sort(sort)
	return ab
}

// Limit adds a $limit stage
func (ab *AggregationBuilder) Limit(limit int64) *AggregationBuilder {
	ab.pipeline.Limit(limit)
	return ab
}

// Skip adds a $skip stage
func (ab *AggregationBuilder) Skip(skip int64) *AggregationBuilder {
	ab.pipeline.Skip(skip)
	return ab
}

// Project adds a $project stage
func (ab *AggregationBuilder) Project(projection bson.M) *AggregationBuilder {
	ab.pipeline.Project(projection)
	return ab
}

// Lookup adds a $lookup stage
func (ab *AggregationBuilder) Lookup(from, localField, foreignField, as string) *AggregationBuilder {
	ab.pipeline.Lookup(from, localField, foreignField, as)
	return ab
}

// Unwind adds an $unwind stage
func (ab *AggregationBuilder) Unwind(path string) *AggregationBuilder {
	ab.pipeline.Unwind(path)
	return ab
}

// AddFields adds an $addFields stage
func (ab *AggregationBuilder) AddFields(fields bson.M) *AggregationBuilder {
	ab.pipeline.AddFields(fields)
	return ab
}

// Count adds a $count stage
func (ab *AggregationBuilder) Count(alias string) *AggregationBuilder {
	ab.pipeline.Count(alias)
	return ab
}

// Sum adds a $sum stage
func (ab *AggregationBuilder) Sum(field, alias string) *AggregationBuilder {
	ab.pipeline.Sum(field, alias)
	return ab
}

// Avg adds an $avg stage
func (ab *AggregationBuilder) Avg(field, alias string) *AggregationBuilder {
	ab.pipeline.Avg(field, alias)
	return ab
}

// Min adds a $min stage
func (ab *AggregationBuilder) Min(field, alias string) *AggregationBuilder {
	ab.pipeline.Min(field, alias)
	return ab
}

// Max adds a $max stage
func (ab *AggregationBuilder) Max(field, alias string) *AggregationBuilder {
	ab.pipeline.Max(field, alias)
	return ab
}

// GroupBy adds a $group stage for grouping
func (ab *AggregationBuilder) GroupBy(field, alias string) *AggregationBuilder {
	ab.pipeline.GroupBy(field, alias)
	return ab
}

// AddStage adds a custom stage
func (ab *AggregationBuilder) AddStage(stage string, data bson.M) *AggregationBuilder {
	ab.pipeline.AddStage(stage, data)
	return ab
}

// Build builds the pipeline
func (ab *AggregationBuilder) Build() []bson.M {
	return ab.pipeline.Build()
}
