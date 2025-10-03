package projection

import (
	"context"
	"fmt"
	"strings"

	"github.com/kabi175/jpack/logger"
	"github.com/kabi175/jpack/repository"
	"github.com/kabi175/jpack/schema"
	"go.mongodb.org/mongo-driver/bson"
)

// ProjectionMode represents the projection mode
type ProjectionMode string

const (
	ProjectionInclude ProjectionMode = "include"
	ProjectionExclude ProjectionMode = "exclude"
)

// JProjection represents a field projection
type JProjection struct {
	Fields []string
	Mode   ProjectionMode
}

// NewProjection creates a new projection
func NewProjection(fields []string, mode ProjectionMode) *JProjection {
	return &JProjection{
		Fields: fields,
		Mode:   mode,
	}
}

// NewIncludeProjection creates an include projection
func NewIncludeProjection(fields ...string) *JProjection {
	return &JProjection{
		Fields: fields,
		Mode:   ProjectionInclude,
	}
}

// NewExcludeProjection creates an exclude projection
func NewExcludeProjection(fields ...string) *JProjection {
	return &JProjection{
		Fields: fields,
		Mode:   ProjectionExclude,
	}
}

// AddField adds a field to the projection
func (p *JProjection) AddField(field string) *JProjection {
	p.Fields = append(p.Fields, field)
	return p
}

// ToMongoProjection converts to MongoDB projection
func (p *JProjection) ToMongoProjection() bson.M {
	projection := bson.M{}

	for _, field := range p.Fields {
		if p.Mode == ProjectionInclude {
			projection[field] = 1
		} else {
			projection[field] = 0
		}
	}

	return projection
}

// String returns a string representation of the projection
func (p *JProjection) String() string {
	mode := "include"
	if p.Mode == ProjectionExclude {
		mode = "exclude"
	}
	logger.Projection.Debug().
		Str("mode", mode).
		Strs("fields", p.Fields).
		Msg("projection string representation")
	return fmt.Sprintf("Projection{%s: [%s]}", mode, strings.Join(p.Fields, ", "))
}

// AggregationType represents the type of aggregation
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

// JAggregation represents an aggregation operation
type JAggregation struct {
	Type     AggregationType
	Field    string
	Alias    string
	Pipeline []JAggregationStage
}

// JAggregationStage represents a stage in aggregation pipeline
type JAggregationStage struct {
	Stage string
	Data  bson.M
}

// NewAggregation creates a new aggregation
func NewAggregation(aggType AggregationType, field, alias string) *JAggregation {
	return &JAggregation{
		Type:     aggType,
		Field:    field,
		Alias:    alias,
		Pipeline: make([]JAggregationStage, 0),
	}
}

// NewCountAggregation creates a count aggregation
func NewCountAggregation(alias string) *JAggregation {
	return &JAggregation{
		Type:  AggCount,
		Alias: alias,
	}
}

// NewSumAggregation creates a sum aggregation
func NewSumAggregation(field, alias string) *JAggregation {
	return &JAggregation{
		Type:  AggSum,
		Field: field,
		Alias: alias,
	}
}

// NewAvgAggregation creates an average aggregation
func NewAvgAggregation(field, alias string) *JAggregation {
	return &JAggregation{
		Type:  AggAvg,
		Field: field,
		Alias: alias,
	}
}

// NewMinAggregation creates a minimum aggregation
func NewMinAggregation(field, alias string) *JAggregation {
	return &JAggregation{
		Type:  AggMin,
		Field: field,
		Alias: alias,
	}
}

// NewMaxAggregation creates a maximum aggregation
func NewMaxAggregation(field, alias string) *JAggregation {
	return &JAggregation{
		Type:  AggMax,
		Field: field,
		Alias: alias,
	}
}

// NewGroupByAggregation creates a group by aggregation
func NewGroupByAggregation(field, alias string) *JAggregation {
	return &JAggregation{
		Type:  AggGroupBy,
		Field: field,
		Alias: alias,
	}
}

// NewPipelineAggregation creates a pipeline aggregation
func NewPipelineAggregation(pipeline []JAggregationStage) *JAggregation {
	return &JAggregation{
		Type:     AggPipeline,
		Pipeline: pipeline,
	}
}

// AddStage adds a stage to the pipeline
func (a *JAggregation) AddStage(stage string, data bson.M) *JAggregation {
	a.Pipeline = append(a.Pipeline, JAggregationStage{
		Stage: stage,
		Data:  data,
	})
	return a
}

// ToMongoAggregation converts to MongoDB aggregation
func (a *JAggregation) ToMongoAggregation() bson.M {
	switch a.Type {
	case AggCount:
		return bson.M{"$count": a.Alias}
	case AggSum:
		return bson.M{"$sum": "$" + a.Field}
	case AggAvg:
		return bson.M{"$avg": "$" + a.Field}
	case AggMin:
		return bson.M{"$min": "$" + a.Field}
	case AggMax:
		return bson.M{"$max": "$" + a.Field}
	case AggGroupBy:
		return bson.M{"$group": bson.M{
			"_id":   "$" + a.Field,
			a.Alias: bson.M{"$push": "$$ROOT"},
		}}
	case AggPipeline:
		// For pipeline, we return the first stage or empty
		if len(a.Pipeline) > 0 {
			return a.Pipeline[0].Data
		}
		return bson.M{}
	default:
		return bson.M{}
	}
}

// ToMongoPipeline converts to MongoDB aggregation pipeline
func (a *JAggregation) ToMongoPipeline() []bson.M {
	if a.Type != AggPipeline {
		return []bson.M{a.ToMongoAggregation()}
	}

	pipeline := make([]bson.M, len(a.Pipeline))
	for i, stage := range a.Pipeline {
		pipeline[i] = bson.M{stage.Stage: stage.Data}
	}

	return pipeline
}

// String returns a string representation of the aggregation
func (a *JAggregation) String() string {
	var result string
	switch a.Type {
	case AggCount:
		result = fmt.Sprintf("CountAggregation{alias: %s}", a.Alias)
	case AggSum:
		result = fmt.Sprintf("SumAggregation{field: %s, alias: %s}", a.Field, a.Alias)
	case AggAvg:
		result = fmt.Sprintf("AvgAggregation{field: %s, alias: %s}", a.Field, a.Alias)
	case AggMin:
		result = fmt.Sprintf("MinAggregation{field: %s, alias: %s}", a.Field, a.Alias)
	case AggMax:
		result = fmt.Sprintf("MaxAggregation{field: %s, alias: %s}", a.Field, a.Alias)
	case AggGroupBy:
		result = fmt.Sprintf("GroupByAggregation{field: %s, alias: %s}", a.Field, a.Alias)
	case AggPipeline:
		result = fmt.Sprintf("PipelineAggregation{stages: %d}", len(a.Pipeline))
	default:
		result = "UnknownAggregation"
	}

	logger.Projection.Debug().
		Str("type", string(a.Type)).
		Str("field", a.Field).
		Str("alias", a.Alias).
		Msg("aggregation string representation")

	return result
}

// ProjectionBuilder provides a fluent interface for building projections
type ProjectionBuilder struct {
	fields []string
	mode   ProjectionMode
}

// NewProjectionBuilder creates a new projection builder
func NewProjectionBuilder() *ProjectionBuilder {
	return &ProjectionBuilder{
		fields: make([]string, 0),
		mode:   ProjectionInclude,
	}
}

// Include sets the projection mode to include
func (pb *ProjectionBuilder) Include() *ProjectionBuilder {
	pb.mode = ProjectionInclude
	return pb
}

// Exclude sets the projection mode to exclude
func (pb *ProjectionBuilder) Exclude() *ProjectionBuilder {
	pb.mode = ProjectionExclude
	return pb
}

// AddField adds a field to the projection
func (pb *ProjectionBuilder) AddField(field string) *ProjectionBuilder {
	pb.fields = append(pb.fields, field)
	return pb
}

// AddFields adds multiple fields to the projection
func (pb *ProjectionBuilder) AddFields(fields ...string) *ProjectionBuilder {
	pb.fields = append(pb.fields, fields...)
	return pb
}

// Build builds the projection
func (pb *ProjectionBuilder) Build() *JProjection {
	return &JProjection{
		Fields: pb.fields,
		Mode:   pb.mode,
	}
}

// ProjectionService provides projection and aggregation services
type ProjectionService interface {
	Project(ctx context.Context, criteria repository.JCriteria, projection *JProjection) ([]schema.JRecord, error)
	Aggregate(ctx context.Context, criteria repository.JCriteria, aggregations []*JAggregation) ([]bson.M, error)
	Count(ctx context.Context, criteria repository.JCriteria) (int64, error)
	GroupBy(ctx context.Context, criteria repository.JCriteria, groupField string) ([]bson.M, error)
}

// jProjectionService implements ProjectionService
type jProjectionService struct {
	repository repository.JpaRepository
}

// NewProjectionService creates a new projection service
func NewProjectionService(repo repository.JpaRepository) ProjectionService {
	return &jProjectionService{
		repository: repo,
	}
}

// Project projects fields from records
func (ps *jProjectionService) Project(ctx context.Context, criteria repository.JCriteria, projection *JProjection) ([]schema.JRecord, error) {
	// Add projection to criteria
	criteria.Projection = projection.Fields

	// Execute query with projection
	records, err := ps.repository.FindBy(ctx, criteria)
	if err != nil {
		return nil, err
	}

	// Convert to schema.JRecord
	result := make([]schema.JRecord, len(records))
	for i, record := range records {
		// Create a new schema record from the repository record
		schemaRec := schema.NewJRecord()
		for _, key := range record.Keys() {
			schemaRec.Set(key, record.Get(key))
		}
		result[i] = schemaRec
	}

	return result, nil
}

// Aggregate performs aggregation operations
func (ps *jProjectionService) Aggregate(ctx context.Context, criteria repository.JCriteria, aggregations []*JAggregation) ([]bson.M, error) {
	// This would need to be implemented in the MongoDB repository
	// For now, return empty result
	return []bson.M{}, nil
}

// Count counts records
func (ps *jProjectionService) Count(ctx context.Context, criteria repository.JCriteria) (int64, error) {
	return ps.repository.Count(ctx, criteria)
}

// GroupBy groups records by a field
func (ps *jProjectionService) GroupBy(ctx context.Context, criteria repository.JCriteria, groupField string) ([]bson.M, error) {
	// This would need to be implemented in the MongoDB repository
	// For now, return empty result
	return []bson.M{}, nil
}
