package repository

import (
	"fmt"
	"strings"

	"github.com/kabi175/jpack/logger"
	"go.mongodb.org/mongo-driver/bson"
)

// QueryBuilder helps build MongoDB queries from JCriteria
type QueryBuilder struct {
	criteria JCriteria
}

// NewQueryBuilder creates a new query builder
func NewQueryBuilder(criteria JCriteria) *QueryBuilder {
	return &QueryBuilder{
		criteria: criteria,
	}
}

// BuildFilter builds MongoDB filter from JCriteria
func (qb *QueryBuilder) BuildFilter() (bson.M, error) {
	filter := bson.M{}

	logger.Repository.Debug().
		Int("filters_count", len(qb.criteria.Filters)).
		Msg("building MongoDB filter from criteria")

	for _, f := range qb.criteria.Filters {
		op, exists := GetOperator(f.Op.Name())
		if !exists {
			logger.Repository.Error().
				Str("operator", f.Op.Name()).
				Msg("unknown operator")
			return nil, fmt.Errorf("unknown operator: %s", f.Op.Name())
		}

		mongoFilter, err := op.Apply(f.Field, f.Value)
		if err != nil {
			logger.Repository.Error().
				Str("operator", f.Op.Name()).
				Str("field", f.Field).
				Err(err).
				Msg("operator failed")
			return nil, fmt.Errorf("operator %s failed: %w", f.Op.Name(), err)
		}

		// Merge the filter
		if mongoMap, ok := mongoFilter.(bson.M); ok {
			for k, v := range mongoMap {
				filter[k] = v
			}
		}

		logger.Repository.Debug().
			Str("field", f.Field).
			Str("operator", f.Op.Name()).
			Interface("value", f.Value).
			Msg("applied filter")
	}

	logger.Repository.Debug().
		Interface("filter", filter).
		Msg("MongoDB filter built successfully")

	return filter, nil
}

// BuildSort builds MongoDB sort from JCriteria
func (qb *QueryBuilder) BuildSort() bson.D {
	sort := bson.D{}

	logger.Repository.Debug().
		Int("sort_count", len(qb.criteria.Sort)).
		Msg("building MongoDB sort from criteria")

	for _, s := range qb.criteria.Sort {
		direction := 1
		if s.Direction == SortDesc {
			direction = -1
		}
		sort = append(sort, bson.E{Key: s.Field, Value: direction})

		logger.Repository.Debug().
			Str("field", s.Field).
			Str("direction", string(s.Direction)).
			Int("mongo_direction", direction).
			Msg("added sort field")
	}

	logger.Repository.Debug().
		Interface("sort", sort).
		Msg("MongoDB sort built successfully")

	return sort
}

// BuildProjection builds MongoDB projection from JCriteria
func (qb *QueryBuilder) BuildProjection() bson.M {
	if len(qb.criteria.Projection) == 0 {
		logger.Repository.Debug().Msg("no projection fields specified")
		return nil
	}

	projection := bson.M{}
	for _, field := range qb.criteria.Projection {
		projection[field] = 1
	}

	logger.Repository.Debug().
		Int("fields_count", len(qb.criteria.Projection)).
		Strs("fields", qb.criteria.Projection).
		Interface("projection", projection).
		Msg("MongoDB projection built successfully")

	return projection
}

// BuildOptions builds MongoDB find options from JCriteria
func (qb *QueryBuilder) BuildOptions() (bson.M, error) {
	options := bson.M{}

	logger.Repository.Debug().
		Int64("limit", qb.criteria.Limit).
		Int64("offset", qb.criteria.Offset).
		Msg("building MongoDB find options from criteria")

	// Sort
	if sort := qb.BuildSort(); len(sort) > 0 {
		options["sort"] = sort
	}

	// Projection
	if projection := qb.BuildProjection(); projection != nil {
		options["projection"] = projection
	}

	// Limit
	if qb.criteria.Limit > 0 {
		options["limit"] = qb.criteria.Limit
	}

	// Skip (offset)
	if qb.criteria.Offset > 0 {
		options["skip"] = qb.criteria.Offset
	}

	logger.Repository.Debug().
		Interface("options", options).
		Msg("MongoDB find options built successfully")

	return options, nil
}

// JCriteriaBuilder helps build JCriteria
type JCriteriaBuilder struct {
	criteria JCriteria
}

// NewJCriteriaBuilder creates a new criteria builder
func NewJCriteriaBuilder() *JCriteriaBuilder {
	return &JCriteriaBuilder{
		criteria: JCriteria{
			Filters:    make([]JFilter, 0),
			Sort:       make([]JSort, 0),
			Projection: make([]string, 0),
		},
	}
}

// AddFilter adds a filter to the criteria
func (cb *JCriteriaBuilder) AddFilter(field string, op JOperator, value any) *JCriteriaBuilder {
	cb.criteria.Filters = append(cb.criteria.Filters, JFilter{
		Field: field,
		Op:    op,
		Value: value,
	})
	return cb
}

// AddSort adds a sort to the criteria
func (cb *JCriteriaBuilder) AddSort(field string, direction SortDirection) *JCriteriaBuilder {
	cb.criteria.Sort = append(cb.criteria.Sort, JSort{
		Field:     field,
		Direction: direction,
	})
	return cb
}

// SetLimit sets the limit
func (cb *JCriteriaBuilder) SetLimit(limit int64) *JCriteriaBuilder {
	cb.criteria.Limit = limit
	return cb
}

// SetOffset sets the offset
func (cb *JCriteriaBuilder) SetOffset(offset int64) *JCriteriaBuilder {
	cb.criteria.Offset = offset
	return cb
}

// AddProjection adds a field to projection
func (cb *JCriteriaBuilder) AddProjection(field string) *JCriteriaBuilder {
	cb.criteria.Projection = append(cb.criteria.Projection, field)
	return cb
}

// Build builds the JCriteria
func (cb *JCriteriaBuilder) Build() JCriteria {
	return cb.criteria
}

// Convenience methods for common filters
func (cb *JCriteriaBuilder) Where(field string, value any) *JCriteriaBuilder {
	return cb.AddFilter(field, OpEQ, value)
}

func (cb *JCriteriaBuilder) WhereNot(field string, value any) *JCriteriaBuilder {
	return cb.AddFilter(field, OpNE, value)
}

func (cb *JCriteriaBuilder) WhereGreaterThan(field string, value any) *JCriteriaBuilder {
	return cb.AddFilter(field, OpGT, value)
}

func (cb *JCriteriaBuilder) WhereGreaterThanOrEqual(field string, value any) *JCriteriaBuilder {
	return cb.AddFilter(field, OpGTE, value)
}

func (cb *JCriteriaBuilder) WhereLessThan(field string, value any) *JCriteriaBuilder {
	return cb.AddFilter(field, OpLT, value)
}

func (cb *JCriteriaBuilder) WhereLessThanOrEqual(field string, value any) *JCriteriaBuilder {
	return cb.AddFilter(field, OpLTE, value)
}

func (cb *JCriteriaBuilder) WhereIn(field string, values any) *JCriteriaBuilder {
	return cb.AddFilter(field, OpIN, values)
}

func (cb *JCriteriaBuilder) WhereNotIn(field string, values any) *JCriteriaBuilder {
	return cb.AddFilter(field, OpNIN, values)
}

func (cb *JCriteriaBuilder) WhereContains(field string, value any) *JCriteriaBuilder {
	return cb.AddFilter(field, OpCONTAINS, value)
}

func (cb *JCriteriaBuilder) WhereStartsWith(field string, value any) *JCriteriaBuilder {
	return cb.AddFilter(field, OpSTARTS_WITH, value)
}

func (cb *JCriteriaBuilder) WhereEndsWith(field string, value any) *JCriteriaBuilder {
	return cb.AddFilter(field, OpENDS_WITH, value)
}

func (cb *JCriteriaBuilder) WhereIsNull(field string) *JCriteriaBuilder {
	return cb.AddFilter(field, OpIS_NULL, nil)
}

func (cb *JCriteriaBuilder) WhereIsNotNull(field string) *JCriteriaBuilder {
	return cb.AddFilter(field, OpIS_NOT_NULL, nil)
}

func (cb *JCriteriaBuilder) WhereIsToday(field string) *JCriteriaBuilder {
	return cb.AddFilter(field, OpIS_TODAY, nil)
}

func (cb *JCriteriaBuilder) WhereIsWeekend(field string) *JCriteriaBuilder {
	return cb.AddFilter(field, OpIS_WEEKEND, nil)
}

func (cb *JCriteriaBuilder) WhereIsWithinBusinessHours(field string) *JCriteriaBuilder {
	return cb.AddFilter(field, OpIS_WITHIN_BUSINESS_HOURS, nil)
}

func (cb *JCriteriaBuilder) OrderBy(field string) *JCriteriaBuilder {
	return cb.AddSort(field, SortAsc)
}

func (cb *JCriteriaBuilder) OrderByDesc(field string) *JCriteriaBuilder {
	return cb.AddSort(field, SortDesc)
}

// String returns a string representation of the criteria
func (c JCriteria) String() string {
	var parts []string

	if len(c.Filters) > 0 {
		filterParts := make([]string, len(c.Filters))
		for i, f := range c.Filters {
			filterParts[i] = fmt.Sprintf("%s %s %v", f.Field, f.Op.Name(), f.Value)
		}
		parts = append(parts, "Filters: ["+strings.Join(filterParts, ", ")+"]")
	}

	if len(c.Sort) > 0 {
		sortParts := make([]string, len(c.Sort))
		for i, s := range c.Sort {
			sortParts[i] = fmt.Sprintf("%s %s", s.Field, s.Direction)
		}
		parts = append(parts, "Sort: ["+strings.Join(sortParts, ", ")+"]")
	}

	if c.Limit > 0 {
		parts = append(parts, fmt.Sprintf("Limit: %d", c.Limit))
	}

	if c.Offset > 0 {
		parts = append(parts, fmt.Sprintf("Offset: %d", c.Offset))
	}

	if len(c.Projection) > 0 {
		parts = append(parts, "Projection: ["+strings.Join(c.Projection, ", ")+"]")
	}

	return "JCriteria{" + strings.Join(parts, ", ") + "}"
}
