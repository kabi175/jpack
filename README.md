# JPack - Schema-Driven Persistence Framework for Go

JPack is a schema-driven persistence framework for Go that works like JPA but is tailored for MongoDB. It provides a comprehensive set of features including schema definition, repository pattern, validation, hooks, converters, and more.

## Features

- **Schema Definition**: Define schemas with fields, types, validations, and relationships
- **Immutable Schemas**: Schemas are immutable once registered, ensuring consistency
- **Schema Evolution**: Replace entire schemas for versioning and evolution
- **Repository Pattern**: CRUD operations with advanced querying capabilities
- **MongoDB Integration**: Seamless integration with MongoDB using the official driver
- **Validation Layer**: Field-level, record-level, and schema-level validations
- **Lifecycle Hooks**: Before/after hooks for save, delete, find, and update operations
- **Type Converters**: Automatic conversion between Go types and MongoDB documents
- **Query API**: Fluent query builder with pluggable operators
- **Projections**: Field projections and aggregations
- **Index Management**: Automatic index creation based on schema definitions

## Installation

```bash
go get github.com/kabi175/jpack
```

## Quick Start

### 1. Define a Schema

```go
package main

import (
    "context"
    "log"
    
    "github.com/kabi175/jpack/mongo"
    "github.com/kabi175/jpack/schema"
    "github.com/kabi175/jpack/validation"
)

func main() {
    // Create MongoDB client
    client, err := mongo.NewMongoClient("mongodb://localhost:27017", "mydb")
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close(context.Background())

    // Define user schema
    userSchema := schema.NewJSchema("User").
        AddField("id", schema.JString, nil).
        AddField("name", schema.JString, nil).
        AddField("email", schema.JString, nil).
        AddField("age", schema.JInt, 18)

    // Add validations
    userSchema.Field("name").(schema.JField).SetRequired(true)
    userSchema.Field("email").(schema.JField).SetRequired(true).SetUnique(true)
    userSchema.AddValidation(validation.ValidateEmail("email"))
    userSchema.AddValidation(validation.ValidateRange("age", 0, 150))

    // Register schema
    err = client.RegisterSchema(userSchema)
    if err != nil {
        log.Fatal(err)
    }
}
```

### 2. Create and Save Records

```go
// Get repository
userRepo, err := client.GetRepository("User")
if err != nil {
    log.Fatal(err)
}

// Create a user record
user := schema.NewJRecord().
    Set("id", "user_001").
    Set("name", "John Doe").
    Set("email", "john@example.com").
    Set("age", 30)

// Save user
savedUser, err := userRepo.Save(context.Background(), user)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Saved user: %+v\n", savedUser.ToMap())
```

### 3. Query Records

```go
// Find by ID
foundUser, err := userRepo.FindById(context.Background(), "user_001")
if err != nil {
    log.Fatal(err)
}

// Query with criteria
criteria := repository.NewJCriteriaBuilder().
    Where("age", repository.OpGTE, 18).
    OrderBy("name").
    SetLimit(10).
    Build()

users, err := userRepo.FindBy(context.Background(), criteria)
if err != nil {
    log.Fatal(err)
}

// Count records
count, err := userRepo.Count(context.Background(), repository.JCriteria{})
if err != nil {
    log.Fatal(err)
}
```

### 4. Update and Delete

```go
// Update records
updateCriteria := repository.NewJCriteriaBuilder().
    Where("id", repository.OpEQ, "user_001").
    Build()

updates := schema.NewJRecord().Set("age", 31)
err = userRepo.Update(context.Background(), updateCriteria, updates)

// Delete record
err = userRepo.Delete(context.Background(), "user_001")
```

## Advanced Features

### Custom Operators

JPack supports pluggable operators for custom query logic:

```go
// Register custom operator
customOp := repository.NewBasicOperator("IS_TODAY", func(field string, _ any) any {
    todayStart := time.Now().Truncate(24 * time.Hour)
    tomorrow := todayStart.Add(24 * time.Hour)
    return bson.M{field: bson.M{
        "$gte": todayStart,
        "$lt":  tomorrow,
    }}
})

repository.RegisterOperator(customOp)

// Use custom operator
criteria := repository.NewJCriteriaBuilder().
    WhereIsToday("created_at").
    Build()
```

### Lifecycle Hooks

Register hooks for various lifecycle events:

```go
// Register hooks
err := hooks.RegisterBeforeSaveHook("timestamp", func(ctx context.Context, rec schema.JRecord) error {
    rec.Set("updated_at", time.Now())
    return nil
}, 100, "Add timestamp")

err = hooks.RegisterAfterSaveHook("logging", func(ctx context.Context, rec schema.JRecord) error {
    log.Printf("Saved record: %+v", rec.ToMap())
    return nil
}, 50, "Log save operation")
```

### Custom Converters

Create custom type converters:

```go
// Money converter
type Money struct {
    Amount   float64 `json:"amount"`
    Currency string  `json:"currency"`
}

moneyConverter := converter.NewCustomConverter(
    "money",
    func(value any) (any, error) {
        if m, ok := value.(Money); ok {
            return map[string]any{
                "amount":   m.Amount,
                "currency": m.Currency,
            }, nil
        }
        return nil, fmt.Errorf("invalid money type")
    },
    func(raw any) (any, error) {
        if m, ok := raw.(map[string]any); ok {
            return Money{
                Amount:   m["amount"].(float64),
                Currency: m["currency"].(string),
            }, nil
        }
        return nil, fmt.Errorf("invalid money data")
    },
)

converter.RegisterConverter(moneyConverter)
```

### Projections and Aggregations

```go
// Field projections
projection := projection.NewIncludeProjection("name", "email")
users, err := userRepo.FindBy(context.Background(), repository.JCriteria{
    Projection: projection.Fields,
})

// Aggregations
collection := client.GetDatabase().Collection("User")
aggExecutor := projection.NewAggregationExecutor(collection)
aggService := projection.NewAggregationService(aggExecutor)

// Count users
count, err := aggService.Count(context.Background(), repository.JCriteria{})

// Average age
avgAge, err := aggService.Avg(context.Background(), repository.JCriteria{}, "age")
```

### Index Management

```go
// Create indexes based on schema
err = client.CreateIndexesForSchema(context.Background(), "User")

// Custom index creation
indexManager := mongo.NewIndexManager(collection, userSchema)
indexDef := mongo.IndexDefinition{
    Fields: []mongo.IndexField{
        {Name: "email", Type: mongo.IndexAscending},
    },
    Options: mongo.IndexOption{
        Unique: true,
        Sparse: true,
    },
}
err = indexManager.CreateCustomIndex(context.Background(), indexDef)
```

## Schema Definition

### Immutable Schemas

JPack uses immutable schemas to ensure consistency and prevent accidental modifications. Once a schema is registered, it cannot be modified directly. To make changes, you must create a new schema and replace the existing one.

```go
// Create a mutable schema
userSchema := schema.NewJSchema("User")
userSchema.AddField("id", schema.JString, nil)
userSchema.AddField("name", schema.JString, nil)

// Register the schema (automatically makes it immutable)
client.RegisterSchema(userSchema)

// To modify, create a new schema and replace
updatedSchema := schema.NewJSchema("User")
updatedSchema.AddField("id", schema.JString, nil)
updatedSchema.AddField("name", schema.JString, nil)
updatedSchema.AddField("email", schema.JString, nil) // New field

// Replace the existing schema
client.RegisterSchema(updatedSchema) // This will replace the old one
```

### Schema Builder

Use the fluent SchemaBuilder for cleaner schema creation:

```go
userSchema := schema.NewSchemaBuilder("User").
    AddRequiredUniqueField("id", schema.JString, nil).
    AddRequiredField("name", schema.JString, nil).
    AddRequiredUniqueField("email", schema.JString, nil).
    AddField("age", schema.JInt, 18).
    SetIDField("id").
    AddValidation(func(ctx context.Context, rec schema.JRecord) error {
        // Custom validation logic
        return nil
    }).
    BuildImmutable() // Makes the schema immutable
```

### Field Types

- `JString` - String
- `JInt` - Integer
- `JInt64` - 64-bit integer
- `JFloat` - Float32
- `JFloat64` - Float64
- `JBool` - Boolean
- `JTime` - Time.Time
- `JObject` - Object/Map
- `JArray` - Array/Slice
- `JBinary` - Binary data
- `JObjectID` - MongoDB ObjectID

### Field Options

```go
field := schema.NewJField("email", schema.JString, nil)
field.SetRequired(true)
field.SetUnique(true)
field.SetValidation(validation.ValidateEmail("email"))
```

### Schema Validations

```go
schema.AddValidation(validation.ValidateMinLength("name", 2))
schema.AddValidation(validation.ValidateMaxLength("name", 100))
schema.AddValidation(validation.ValidateRange("age", 0, 150))
schema.AddValidation(validation.ValidateFieldsEqual("password", "confirmPassword"))
```

### Schema Evolution

For schema evolution, use the replacement pattern:

```go
// Version 1
v1Schema := schema.NewSchemaBuilder("User").
    AddRequiredUniqueField("id", schema.JString, nil).
    AddRequiredField("name", schema.JString, nil).
    BuildImmutable()

client.RegisterSchema(v1Schema)

// Version 2 - Add email field
v2Schema := schema.NewSchemaBuilder("User").
    AddRequiredUniqueField("id", schema.JString, nil).
    AddRequiredField("name", schema.JString, nil).
    AddRequiredUniqueField("email", schema.JString, nil). // New field
    BuildImmutable()

// Replace the schema
client.RegisterSchema(v2Schema)
```

## Query API

### Built-in Operators

- `OpEQ` - Equals
- `OpNE` - Not equals
- `OpGT` - Greater than
- `OpGTE` - Greater than or equal
- `OpLT` - Less than
- `OpLTE` - Less than or equal
- `OpIN` - In array
- `OpNIN` - Not in array
- `OpEXISTS` - Field exists
- `OpREGEX` - Regular expression
- `OpCONTAINS` - Contains text
- `OpSTARTS_WITH` - Starts with
- `OpENDS_WITH` - Ends with
- `OpIS_NULL` - Is null
- `OpIS_NOT_NULL` - Is not null
- `OpIS_TODAY` - Is today
- `OpIS_WEEKEND` - Is weekend
- `OpIS_WITHIN_BUSINESS_HOURS` - Within business hours

### Query Builder

```go
criteria := repository.NewJCriteriaBuilder().
    Where("age", repository.OpGTE, 18).
    Where("status", repository.OpEQ, "active").
    WhereIn("category", []any{"electronics", "clothing"}).
    OrderBy("name").
    OrderByDesc("created_at").
    SetLimit(10).
    SetOffset(0).
    Build()
```

## Validation

### Field Validations

```go
// Required validation
validation.ValidateRequired("name")

// Length validations
validation.ValidateMinLength("name", 2)
validation.ValidateMaxLength("name", 100)

// Value validations
validation.ValidateMinValue("age", 0)
validation.ValidateMaxValue("age", 150)
validation.ValidateRange("age", 0, 150)

// Format validations
validation.ValidateEmail("email")
validation.ValidateRegex("phone", `^\d{10}$`)

// List validations
validation.ValidateFieldIn("status", []any{"active", "inactive"})
validation.ValidateFieldNotIn("status", []any{"deleted"})
```

### Cross-field Validations

```go
// Field equality
validation.ValidateFieldsEqual("password", "confirmPassword")

// Field inequality
validation.ValidateFieldsNotEqual("email", "username")

// Field comparisons
validation.ValidateFieldGreaterThan("endDate", "startDate")
validation.ValidateFieldLessThan("startDate", "endDate")
```

## Hooks

### Hook Types

- `HookBeforeSave` - Before saving
- `HookAfterSave` - After saving
- `HookBeforeDelete` - Before deleting
- `HookAfterDelete` - After deleting
- `HookBeforeFind` - Before finding
- `HookAfterFind` - After finding
- `HookBeforeUpdate` - Before updating
- `HookAfterUpdate` - After updating

### Built-in Hooks

```go
// Timestamp hook
timestampHook := hooks.NewTimestampHook("created_at", "updated_at")

// Validation hook
validationHook := hooks.NewValidationHook(schema)

// Logging hook
loggingHook := hooks.NewLoggingHook(logger)

// Audit hook
auditHook := hooks.NewAuditHook("audit_info")

// Soft delete hook
softDeleteHook := hooks.NewSoftDeleteHook("deleted_at", "deleted_by")

// Cache hook
cacheHook := hooks.NewCacheHook(cache)
```

## Examples

See the `examples/` directory for complete working examples:

- `user_example.go` - Basic user management
- `product_example.go` - Product catalog with complex queries
- `custom_converter_example.go` - Custom type converters

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Support

For questions, issues, or contributions, please open an issue on GitHub.