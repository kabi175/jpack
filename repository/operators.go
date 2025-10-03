package repository

import (
	"fmt"
	"sync"
	"time"

	"github.com/kabi175/jpack/logger"
	"go.mongodb.org/mongo-driver/bson"
)

// basicOperator implements JOperator interface
type basicOperator struct {
	name string
	fn   func(field string, value any) any
}

// NewBasicOperator creates a new basic operator
func NewBasicOperator(name string, fn func(field string, value any) any) JOperator {
	return &basicOperator{
		name: name,
		fn:   fn,
	}
}

func (b *basicOperator) Name() string {
	return b.name
}

func (b *basicOperator) Apply(field string, value any) (any, error) {
	return b.fn(field, value), nil
}

// Built-in operators
var (
	OpEQ = NewBasicOperator("EQ", func(field string, value any) any {
		return bson.M{field: value}
	})

	OpNE = NewBasicOperator("NE", func(field string, value any) any {
		return bson.M{field: bson.M{"$ne": value}}
	})

	OpGT = NewBasicOperator("GT", func(field string, value any) any {
		return bson.M{field: bson.M{"$gt": value}}
	})

	OpGTE = NewBasicOperator("GTE", func(field string, value any) any {
		return bson.M{field: bson.M{"$gte": value}}
	})

	OpLT = NewBasicOperator("LT", func(field string, value any) any {
		return bson.M{field: bson.M{"$lt": value}}
	})

	OpLTE = NewBasicOperator("LTE", func(field string, value any) any {
		return bson.M{field: bson.M{"$lte": value}}
	})

	OpIN = NewBasicOperator("IN", func(field string, value any) any {
		return bson.M{field: bson.M{"$in": value}}
	})

	OpNIN = NewBasicOperator("NIN", func(field string, value any) any {
		return bson.M{field: bson.M{"$nin": value}}
	})

	OpEXISTS = NewBasicOperator("EXISTS", func(field string, value any) any {
		return bson.M{field: bson.M{"$exists": value}}
	})

	OpREGEX = NewBasicOperator("REGEX", func(field string, value any) any {
		return bson.M{field: bson.M{"$regex": value}}
	})

	OpCONTAINS = NewBasicOperator("CONTAINS", func(field string, value any) any {
		return bson.M{field: bson.M{"$regex": fmt.Sprintf(".*%v.*", value), "$options": "i"}}
	})

	OpSTARTS_WITH = NewBasicOperator("STARTS_WITH", func(field string, value any) any {
		return bson.M{field: bson.M{"$regex": fmt.Sprintf("^%v", value), "$options": "i"}}
	})

	OpENDS_WITH = NewBasicOperator("ENDS_WITH", func(field string, value any) any {
		return bson.M{field: bson.M{"$regex": fmt.Sprintf("%v$", value), "$options": "i"}}
	})

	OpIS_NULL = NewBasicOperator("IS_NULL", func(field string, _ any) any {
		return bson.M{field: bson.M{"$exists": false}}
	})

	OpIS_NOT_NULL = NewBasicOperator("IS_NOT_NULL", func(field string, _ any) any {
		return bson.M{field: bson.M{"$exists": true, "$ne": nil}}
	})

	// Date/time operators
	OpIS_TODAY = NewBasicOperator("IS_TODAY", func(field string, _ any) any {
		now := time.Now()
		todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		tomorrow := todayStart.Add(24 * time.Hour)
		return bson.M{field: bson.M{
			"$gte": todayStart,
			"$lt":  tomorrow,
		}}
	})

	OpIS_WEEKEND = NewBasicOperator("IS_WEEKEND", func(field string, _ any) any {
		now := time.Now()
		weekday := now.Weekday()
		return bson.M{field: bson.M{
			"$gte": time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()),
			"$lt":  time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 999999999, now.Location()),
			"$expr": bson.M{
				"$in": []any{weekday, time.Saturday, time.Sunday},
			},
		}}
	})

	OpIS_WITHIN_BUSINESS_HOURS = NewBasicOperator("IS_WITHIN_BUSINESS_HOURS", func(field string, _ any) any {
		now := time.Now()
		hour := now.Hour()
		weekday := now.Weekday()

		// Business hours: Monday-Friday, 9 AM - 5 PM
		return bson.M{field: bson.M{
			"$gte": time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()),
			"$lt":  time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 999999999, now.Location()),
			"$expr": bson.M{
				"$and": []any{
					bson.M{"$gte": []any{hour, 9}},
					bson.M{"$lt": []any{hour, 17}},
					bson.M{"$gte": []any{weekday, time.Monday}},
					bson.M{"$lte": []any{weekday, time.Friday}},
				},
			},
		}}
	})
)

// OperatorRegistry manages custom operators
type OperatorRegistry struct {
	operators map[string]JOperator
	mutex     sync.RWMutex
}

// Global operator registry
var globalOperatorRegistry = &OperatorRegistry{
	operators: make(map[string]JOperator),
}

// RegisterOperator registers a custom operator
func RegisterOperator(op JOperator) {
	globalOperatorRegistry.mutex.Lock()
	defer globalOperatorRegistry.mutex.Unlock()

	logger.Repository.Debug().
		Str("operator", op.Name()).
		Msg("registering custom operator")

	globalOperatorRegistry.operators[op.Name()] = op

	logger.Repository.Debug().
		Str("operator", op.Name()).
		Int("total_operators", len(globalOperatorRegistry.operators)).
		Msg("custom operator registered successfully")
}

// GetOperator retrieves an operator by name
func GetOperator(name string) (JOperator, bool) {
	globalOperatorRegistry.mutex.RLock()
	defer globalOperatorRegistry.mutex.RUnlock()

	// Check built-in operators first
	switch name {
	case "EQ":
		return OpEQ, true
	case "NE":
		return OpNE, true
	case "GT":
		return OpGT, true
	case "GTE":
		return OpGTE, true
	case "LT":
		return OpLT, true
	case "LTE":
		return OpLTE, true
	case "IN":
		return OpIN, true
	case "NIN":
		return OpNIN, true
	case "EXISTS":
		return OpEXISTS, true
	case "REGEX":
		return OpREGEX, true
	case "CONTAINS":
		return OpCONTAINS, true
	case "STARTS_WITH":
		return OpSTARTS_WITH, true
	case "ENDS_WITH":
		return OpENDS_WITH, true
	case "IS_NULL":
		return OpIS_NULL, true
	case "IS_NOT_NULL":
		return OpIS_NOT_NULL, true
	case "IS_TODAY":
		return OpIS_TODAY, true
	case "IS_WEEKEND":
		return OpIS_WEEKEND, true
	case "IS_WITHIN_BUSINESS_HOURS":
		return OpIS_WITHIN_BUSINESS_HOURS, true
	}

	// Check custom operators
	op, exists := globalOperatorRegistry.operators[name]
	if !exists {
		logger.Repository.Debug().
			Str("operator", name).
			Msg("operator not found")
	} else {
		logger.Repository.Debug().
			Str("operator", name).
			Msg("operator found")
	}

	return op, exists
}

// ListOperators returns all registered operator names
func ListOperators() []string {
	globalOperatorRegistry.mutex.RLock()
	defer globalOperatorRegistry.mutex.RUnlock()

	operators := []string{
		"EQ", "NE", "GT", "GTE", "LT", "LTE",
		"IN", "NIN", "EXISTS", "REGEX", "CONTAINS",
		"STARTS_WITH", "ENDS_WITH", "IS_NULL", "IS_NOT_NULL",
		"IS_TODAY", "IS_WEEKEND", "IS_WITHIN_BUSINESS_HOURS",
	}

	for name := range globalOperatorRegistry.operators {
		operators = append(operators, name)
	}

	logger.Repository.Debug().
		Int("count", len(operators)).
		Strs("operators", operators).
		Msg("listing registered operators")

	return operators
}
