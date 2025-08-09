package jpack

import (
	"go.mongodb.org/mongo-driver/v2/bson"
)

type Query interface {
	// base schema that this query is built upon
	Schema() JSchema

	// selects fields from the schema
	Select(...JField) Query

	// uses eager loading to load the referenced schema
	With(JRef, func(JSchema, Query) Query) Query

	// where clause
	Where(Filter) Query

	// order by clause
	OrderBy(...JField) Query

	// limit clause
	Limit(int) Query

	// offset clause
	Offset(int) Query

	// execute the query
	Execute() ([]JRecord, error)

	// execute the query and return the first record
	First() (JRecord, error)

	// execute the query and return the count of records
	Count() (int, error)
}

// FilterResolver converts a Filter to MongoDB BSON format
type FilterResolver func(Filter) bson.M

// Global registry for filter resolvers
var filterResolvers = make(map[string]FilterResolver)

// RegisterFilterResolver registers a resolver for a specific operator
func RegisterFilterResolver(operator string, resolver FilterResolver) {
	filterResolvers[operator] = resolver
}

// GetFilterResolver retrieves a resolver for a specific operator
func GetFilterResolver(operator string) (FilterResolver, bool) {
	resolver, exists := filterResolvers[operator]
	return resolver, exists
}

// ResolveFilter converts a Filter to MongoDB BSON format using registered resolvers
func ResolveFilter(filter Filter) bson.M {
	if filter == nil {
		return nil
	}

	operator := filter.Operator()

	// Check if we have a custom resolver for this operator
	if resolver, exists := GetFilterResolver(operator); exists {
		return resolver(filter)
	}

	// Handle logical operators
	switch operator {
	case "AND":
		left := ResolveFilter(filter.Left())
		right := ResolveFilter(filter.Right())
		if left != nil && right != nil {
			return bson.M{"$and": []bson.M{left, right}}
		} else if left != nil {
			return left
		} else if right != nil {
			return right
		}
		return nil
	case "OR":
		left := ResolveFilter(filter.Left())
		right := ResolveFilter(filter.Right())
		if left != nil && right != nil {
			return bson.M{"$or": []bson.M{left, right}}
		} else if left != nil {
			return left
		} else if right != nil {
			return right
		}
		return nil
	case "NOT":
		right := ResolveFilter(filter.Right())
		if right != nil {
			return bson.M{"$not": right}
		}
		return nil
	}

	// Default behavior for field-based filters
	field := filter.Field()
	value := filter.Value()

	if field == nil {
		return nil
	}

	fieldName := field.Name()

	// Handle different operators
	switch operator {
	case "=":
		return bson.M{fieldName: value}
	case "!=":
		return bson.M{fieldName: bson.M{"$ne": value}}
	case "<":
		return bson.M{fieldName: bson.M{"$lt": value}}
	case "<=":
		return bson.M{fieldName: bson.M{"$lte": value}}
	case ">":
		return bson.M{fieldName: bson.M{"$gt": value}}
	case ">=":
		return bson.M{fieldName: bson.M{"$gte": value}}
	case "IN":
		if values, ok := value.([]any); ok {
			return bson.M{fieldName: bson.M{"$in": values}}
		}
	case "NOT IN":
		if values, ok := value.([]any); ok {
			return bson.M{fieldName: bson.M{"$nin": values}}
		}
	case "LIKE":
		if pattern, ok := value.(string); ok {
			return bson.M{fieldName: bson.M{"$regex": pattern}}
		}
	case "NOT LIKE":
		if pattern, ok := value.(string); ok {
			return bson.M{fieldName: bson.M{"$not": bson.M{"$regex": pattern}}}
		}
	case "BETWEEN":
		if values, ok := value.([]any); ok && len(values) == 2 {
			return bson.M{fieldName: bson.M{"$gte": values[0], "$lte": values[1]}}
		}
	case "NOT BETWEEN":
		if values, ok := value.([]any); ok && len(values) == 2 {
			return bson.M{fieldName: bson.M{"$not": bson.M{"$gte": values[0], "$lte": values[1]}}}
		}
	case "EXISTS":
		return bson.M{fieldName: bson.M{"$exists": true}}
	case "NOT EXISTS":
		return bson.M{fieldName: bson.M{"$exists": false}}
	}

	return nil
}

// Initialize default resolvers
func init() {
	// Register default resolvers for built-in operators
	RegisterFilterResolver("=", func(filter Filter) bson.M {
		field := filter.Field()
		value := filter.Value()
		if field == nil {
			return nil
		}
		return bson.M{field.Name(): value}
	})

	RegisterFilterResolver("!=", func(filter Filter) bson.M {
		field := filter.Field()
		value := filter.Value()
		if field == nil {
			return nil
		}
		return bson.M{field.Name(): bson.M{"$ne": value}}
	})

	RegisterFilterResolver("<", func(filter Filter) bson.M {
		field := filter.Field()
		value := filter.Value()
		if field == nil {
			return nil
		}
		return bson.M{field.Name(): bson.M{"$lt": value}}
	})

	RegisterFilterResolver("<=", func(filter Filter) bson.M {
		field := filter.Field()
		value := filter.Value()
		if field == nil {
			return nil
		}
		return bson.M{field.Name(): bson.M{"$lte": value}}
	})

	RegisterFilterResolver(">", func(filter Filter) bson.M {
		field := filter.Field()
		value := filter.Value()
		if field == nil {
			return nil
		}
		return bson.M{field.Name(): bson.M{"$gt": value}}
	})

	RegisterFilterResolver(">=", func(filter Filter) bson.M {
		field := filter.Field()
		value := filter.Value()
		if field == nil {
			return nil
		}
		return bson.M{field.Name(): bson.M{"$gte": value}}
	})

	RegisterFilterResolver("IN", func(filter Filter) bson.M {
		field := filter.Field()
		value := filter.Value()
		if field == nil {
			return nil
		}
		if values, ok := value.([]any); ok {
			return bson.M{field.Name(): bson.M{"$in": values}}
		}
		return nil
	})

	RegisterFilterResolver("NOT IN", func(filter Filter) bson.M {
		field := filter.Field()
		value := filter.Value()
		if field == nil {
			return nil
		}
		if values, ok := value.([]any); ok {
			return bson.M{field.Name(): bson.M{"$nin": values}}
		}
		return nil
	})

	RegisterFilterResolver("LIKE", func(filter Filter) bson.M {
		field := filter.Field()
		value := filter.Value()
		if field == nil {
			return nil
		}
		if pattern, ok := value.(string); ok {
			return bson.M{field.Name(): bson.M{"$regex": pattern}}
		}
		return nil
	})

	RegisterFilterResolver("NOT LIKE", func(filter Filter) bson.M {
		field := filter.Field()
		value := filter.Value()
		if field == nil {
			return nil
		}
		if pattern, ok := value.(string); ok {
			return bson.M{field.Name(): bson.M{"$not": bson.M{"$regex": pattern}}}
		}
		return nil
	})

	RegisterFilterResolver("BETWEEN", func(filter Filter) bson.M {
		field := filter.Field()
		value := filter.Value()
		if field == nil {
			return nil
		}
		if values, ok := value.([]any); ok && len(values) == 2 {
			return bson.M{field.Name(): bson.M{"$gte": values[0], "$lte": values[1]}}
		}
		return nil
	})

	RegisterFilterResolver("NOT BETWEEN", func(filter Filter) bson.M {
		field := filter.Field()
		value := filter.Value()
		if field == nil {
			return nil
		}
		if values, ok := value.([]any); ok && len(values) == 2 {
			return bson.M{field.Name(): bson.M{"$not": bson.M{"$gte": values[0], "$lte": values[1]}}}
		}
		return nil
	})

	RegisterFilterResolver("EXISTS", func(filter Filter) bson.M {
		field := filter.Field()
		if field == nil {
			return nil
		}
		return bson.M{field.Name(): bson.M{"$exists": true}}
	})

	RegisterFilterResolver("NOT EXISTS", func(filter Filter) bson.M {
		field := filter.Field()
		if field == nil {
			return nil
		}
		return bson.M{field.Name(): bson.M{"$exists": false}}
	})
}

type Comparator func(JField, any) Filter
type UnaryOperator func(JField) Filter
type RangeComparator func(JField, any, any) Filter
type LogicalOperator func(Filter, Filter) Filter
type UnaryLogicalOperator func(Filter) Filter

func NewComparator(operator string) Comparator {
	return func(field JField, value any) Filter {
		return &filterImpl{
			field:    field,
			value:    value,
			operator: operator,
		}
	}
}

func NewUnaryComparator(operator string) UnaryOperator {
	return func(field JField) Filter {
		return &filterImpl{
			field:    field,
			operator: operator,
		}
	}
}

func NewRangeComparator(operator string) RangeComparator {
	return func(field JField, min any, max any) Filter {
		return &filterImpl{
			field:    field,
			value:    []any{min, max},
			operator: operator,
		}
	}
}

var (
	Eq      Comparator = NewComparator("=")
	Ne      Comparator = NewComparator("!=")
	Lt      Comparator = NewComparator("<")
	Lte     Comparator = NewComparator("<=")
	Gt      Comparator = NewComparator(">")
	Gte     Comparator = NewComparator(">=")
	In      Comparator = NewComparator("IN")
	NotIn   Comparator = NewComparator("NOT IN")
	Like    Comparator = NewComparator("LIKE")
	NotLike Comparator = NewComparator("NOT LIKE")

	Between    RangeComparator = NewRangeComparator("BETWEEN")
	NotBetween RangeComparator = NewRangeComparator("NOT BETWEEN")

	Exists    UnaryOperator = NewUnaryComparator("EXISTS")
	NotExists UnaryOperator = NewUnaryComparator("NOT EXISTS")

	And LogicalOperator      = func(f1, f2 Filter) Filter { return f1.And(f2) }
	Or  LogicalOperator      = func(f1, f2 Filter) Filter { return f1.Or(f2) }
	Not UnaryLogicalOperator = func(f1 Filter) Filter { return f1.Not() }
)

// filterImpl implements the Filter interface
type filterImpl struct {
	field JField
	value any

	left     Filter
	right    Filter
	operator string
}

// Not implements Filter.
func (f *filterImpl) Not() Filter {
	return &filterImpl{
		right:    f,
		operator: "NOT",
	}
}

// Field implements Filter.
func (f *filterImpl) Field() JField {
	return f.field
}

// Value implements Filter.
func (f *filterImpl) Value() any {
	return f.value
}

// And implements Filter.
func (f *filterImpl) And(filter Filter) Filter {
	return &filterImpl{
		left:     f,
		right:    filter,
		operator: "AND",
	}
}

// Or implements Filter.
func (f *filterImpl) Or(filter Filter) Filter {
	return &filterImpl{
		left:     f,
		right:    filter,
		operator: "OR",
	}
}

// Left implements Filter.
func (f *filterImpl) Left() Filter {
	return f.left
}

// Operator implements Filter.
func (f *filterImpl) Operator() string {
	return f.operator
}

// Right implements Filter.
func (f *filterImpl) Right() Filter {
	return f.right
}

var _ Filter = &filterImpl{}
