# JPack - Schema-Driven Persistence Framework for Go

[![Go Version](https://img.shields.io/badge/go-1.24.2-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/kabi175/jpack)](https://goreportcard.com/report/github.com/kabi175/jpack)

JPack is a comprehensive, schema-driven persistence framework for Go that provides JPA-like functionality tailored specifically for MongoDB. It offers a robust set of features including immutable schemas, repository pattern, validation, hooks, converters, and advanced querying capabilities.

## 🚀 Features

- **🔒 Immutable Schemas**: Schemas are immutable by default, ensuring consistency and thread safety
- **🏗️ Builder Pattern**: Fluent API for schema construction with callback-based updates
- **📊 Repository Pattern**: Full CRUD operations with advanced querying capabilities
- **🔍 Advanced Querying**: Fluent query builder with pluggable operators and projections
- **✅ Validation Layer**: Field-level, record-level, and schema-level validations
- **🪝 Lifecycle Hooks**: Before/after hooks for all CRUD operations
- **🔄 Type Converters**: Automatic conversion between Go types and MongoDB documents
- **📈 Aggregations**: Built-in aggregation support with MongoDB pipeline
- **🗂️ Index Management**: Automatic index creation based on schema definitions
- **⚡ Lazy Loading**: Lazy-loaded schema registry for better performance
- **🔧 MongoDB Integration**: Seamless integration with the official MongoDB driver
- **🧵 Thread Safety**: All components are thread-safe and goroutine-friendly

## 📦 Installation

```bash
go get github.com/kabi175/jpack
```

## 🎯 Quick Start

### 1. Basic Setup

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

    // Define user schema using builder pattern
    userSchema := schema.NewSchemaBuilder("User").
        AddField("name", schema.JString, nil).
        AddField("email", schema.JString, nil).
        AddField("age", schema.JInt, 18).
        Build()

    // Register schema
    err = client.RegisterSchema(userSchema)
    if err != nil {
        log.Fatal(err)
    }

    // Get repository
    userRepo, err := client.GetRepository("User")
    if err != nil {
        log.Fatal(err)
    }

    // Create and save a user
    user := schema.NewJRecord().
        Set("id", "user_001").
        Set("name", "John Doe").
        Set("email", "john@example.com").
        Set("age", 30)

    savedUser, err := userRepo.Save(context.Background(), user)
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("Saved user: %+v", savedUser.ToMap())
}
```

### 2. Advanced Schema with Validations

```go
// Create schema with validations and field constraints
userSchema := schema.NewSchemaBuilder("User").
    AddRequiredField("name", schema.JString, nil).
    AddRequiredUniqueField("email", schema.JString, nil).
    AddField("age", schema.JInt, 18).
    AddValidation(func(ctx context.Context, rec schema.JRecord) error {
        return validation.ValidateEmail("email")(ctx, rec)
    }).
    AddValidation(func(ctx context.Context, rec schema.JRecord) error {
        return validation.ValidateRange("age", 0, 150)(ctx, rec)
    }).
    Build()
```

### 3. Querying with Criteria

```go
// Build complex queries
criteria := repository.NewJCriteriaBuilder().
    Where("age", repository.OpGTE, 18).
    Where("status", repository.OpEQ, "active").
    WhereIn("category", []any{"electronics", "clothing"}).
    OrderBy("name").
    OrderByDesc("created_at").
    SetLimit(10).
    SetOffset(0).
    Build()

users, err := userRepo.FindBy(context.Background(), criteria)
```

## 🏗️ Architecture

JPack follows a layered architecture:

```
┌─────────────────────────────────────────────────────────────┐
│                    Application Layer                        │
├─────────────────────────────────────────────────────────────┤
│                    Repository Layer                         │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐          │
│  │   MongoDB   │  │   Query     │  │  Projection │          │
│  │ Repository  │  │   Builder   │  │   Service   │          │
│  └─────────────┘  └─────────────┘  └─────────────┘          │
├─────────────────────────────────────────────────────────────┤
│                    Schema Layer                             │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐          │
│  │   Schema    │  │   Field     │  │ Validation  │          │
│  │  Builder    │  │   Builder   │  │   Engine    │          │
│  └─────────────┘  └─────────────┘  └─────────────┘          │
├─────────────────────────────────────────────────────────────┤
│                    Core Layer                               │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐          │
│  │   Hooks     │  │ Converters  │  │   Logger    │          │
│  │  Manager    │  │  Registry   │  │   System    │          │
│  └─────────────┘  └─────────────┘  └─────────────┘          │
├─────────────────────────────────────────────────────────────┤
│                    MongoDB Driver                           │
└─────────────────────────────────────────────────────────────┘
```

## 📚 Core Components

### 1. Schema System

JPack uses immutable schemas with a builder pattern for construction and callback-based updates for modifications.

#### Schema Builder

```go
// Create immutable schema
userSchema := schema.NewSchemaBuilder("User").
    AddRequiredField("name", schema.JString, nil).
    AddRequiredUniqueField("email", schema.JString, nil).
    AddField("age", schema.JInt, 18).
    SetIDField("id").
    Build()

// Update schema (creates new instance)
updatedSchema := userSchema.Update(func(sb *schema.SchemaBuilder) {
    sb.AddField("phone", schema.JString, nil)
    sb.AddValidation(myValidationFunc)
})
```

#### Field Types

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

### 2. Repository Pattern

The repository provides a clean abstraction over MongoDB operations:

```go
type JpaRepository interface {
    // Basic CRUD
    Save(ctx context.Context, rec JRecord) (JRecord, error)
    FindById(ctx context.Context, id any) (JRecord, error)
    Delete(ctx context.Context, id any) error
    FindAll(ctx context.Context) ([]JRecord, error)
    
    // Advanced operations
    FindBy(ctx context.Context, criteria JCriteria) ([]JRecord, error)
    Count(ctx context.Context, criteria JCriteria) (int64, error)
    Update(ctx context.Context, criteria JCriteria, updates JRecord) error
    DeleteBy(ctx context.Context, criteria JCriteria) error
    FindOne(ctx context.Context, criteria JCriteria) (JRecord, error)
}
```

### 3. Query System

#### Built-in Operators

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

#### Custom Operators

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

### 4. Validation System

#### Field Validations

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

#### Cross-field Validations

```go
// Field equality
validation.ValidateFieldsEqual("password", "confirmPassword")

// Field inequality
validation.ValidateFieldsNotEqual("email", "username")

// Field comparisons
validation.ValidateFieldGreaterThan("endDate", "startDate")
validation.ValidateFieldLessThan("startDate", "endDate")
```

### 5. Hooks System

#### Hook Types

- `HookBeforeSave` - Before saving
- `HookAfterSave` - After saving
- `HookBeforeDelete` - Before deleting
- `HookAfterDelete` - After deleting
- `HookBeforeFind` - Before finding
- `HookAfterFind` - After finding
- `HookBeforeUpdate` - Before updating
- `HookAfterUpdate` - After updating

#### Registering Hooks

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

#### Built-in Hooks

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
```

### 6. Type Converters

#### Custom Converters

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

### 7. Projections and Aggregations

#### Field Projections

```go
// Include specific fields
projection := projection.NewIncludeProjection("name", "email")
users, err := userRepo.FindBy(context.Background(), repository.JCriteria{
    Projection: projection.Fields,
})

// Exclude specific fields
excludeProjection := projection.NewExcludeProjection("password", "salt")
users, err := userRepo.FindBy(context.Background(), repository.JCriteria{
    Projection: excludeProjection.Fields,
})
```

#### Aggregations

```go
collection := client.GetDatabase().Collection("User")
aggExecutor := projection.NewAggregationExecutor(collection)
aggService := projection.NewAggregationService(aggExecutor)

// Count users
count, err := aggService.Count(context.Background(), repository.JCriteria{})

// Average age
avgAge, err := aggService.Avg(context.Background(), repository.JCriteria{}, "age")

// Group by category
groups, err := aggService.GroupBy(context.Background(), repository.JCriteria{}, "category")

// Custom pipeline
pipeline := []bson.M{
    {"$match": bson.M{"age": bson.M{"$gte": 18}}},
    {"$group": bson.M{"_id": "$category", "count": bson.M{"$sum": 1}}},
    {"$sort": bson.M{"count": -1}},
}
result, err := aggService.Pipeline(context.Background(), pipeline)
```

### 8. Index Management

#### Automatic Index Creation

```go
// Create indexes for all registered schemas
err = client.CreateIndexes(context.Background())

// Create indexes for specific schema
err = client.CreateIndexesForSchema(context.Background(), "User")
```

#### Custom Index Creation

```go
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

### 9. Lazy Schema Registry

#### Lazy Loading

```go
registry := schema.NewJSchemaRegistry()

// Eager registration
userSchema := schema.NewSchemaBuilder("User").
    AddField("name", schema.JString, nil).
    Build()
registry.Register(userSchema)

// Lazy registration
registry.RegisterLazy("Order", &FileSchemaSource{Path: "order_schema.json"})

// First Get will trigger lazy load
orderSchema, ok := registry.Get("Order")
if ok {
    fmt.Println("Loaded schema:", orderSchema.Name())
}
```

#### Custom Schema Sources

```go
type FileSchemaSource struct {
    Path string
}

func (f *FileSchemaSource) Load(schemaName string) (schema.JSchema, error) {
    data, err := os.ReadFile(f.Path)
    if err != nil {
        return nil, err
    }
    var def SchemaDef
    if err := json.Unmarshal(data, &def); err != nil {
        return nil, err
    }
    return ConvertDefToSchema(def), nil
}
```

## 🔧 Configuration

### MongoDB Client Options

```go
// Basic client
client, err := mongo.NewMongoClient("mongodb://localhost:27017", "mydb")

// Client with custom options
clientOptions := options.Client().
    SetMaxPoolSize(100).
    SetMinPoolSize(10).
    SetMaxConnIdleTime(30 * time.Second)

client, err := mongo.NewMongoClientWithOptions(
    "mongodb://localhost:27017", 
    "mydb", 
    clientOptions,
)
```

### Logging Configuration

```go
import "github.com/kabi175/jpack/logger"

// Initialize logger
logger.Initialize()

// Configure log levels
logger.SetLevel("debug") // debug, info, warn, error
```

## 🧪 Testing

JPack provides comprehensive testing utilities:

```go
// Test schema creation
func TestUserSchema(t *testing.T) {
    schema := schema.NewSchemaBuilder("User").
        AddRequiredField("name", schema.JString, nil).
        Build()
    
    assert.Equal(t, "User", schema.Name())
    assert.True(t, schema.IsImmutable())
}

// Test repository operations
func TestUserRepository(t *testing.T) {
    client := setupTestClient(t)
    defer client.Close(context.Background())
    
    userRepo, err := client.GetRepository("User")
    assert.NoError(t, err)
    
    user := schema.NewJRecord().
        Set("id", "test_user").
        Set("name", "Test User")
    
    saved, err := userRepo.Save(context.Background(), user)
    assert.NoError(t, err)
    assert.Equal(t, "Test User", saved.Get("name"))
}
```

## 🚀 Performance

### Best Practices

1. **Use Projections**: Only fetch required fields
2. **Leverage Indexes**: Create appropriate indexes for query patterns
3. **Batch Operations**: Use bulk operations for multiple records
4. **Connection Pooling**: Configure appropriate pool sizes
5. **Lazy Loading**: Use lazy schema registry for better startup performance

### Benchmarks

```go
// Benchmark schema creation
func BenchmarkSchemaCreation(b *testing.B) {
    for i := 0; i < b.N; i++ {
        schema.NewSchemaBuilder("User").
            AddField("name", schema.JString, nil).
            AddField("email", schema.JString, nil).
            Build()
    }
}

// Benchmark repository operations
func BenchmarkRepositorySave(b *testing.B) {
    client := setupBenchmarkClient()
    defer client.Close(context.Background())
    
    userRepo, _ := client.GetRepository("User")
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        user := schema.NewJRecord().
            Set("id", fmt.Sprintf("user_%d", i)).
            Set("name", "Test User")
        userRepo.Save(context.Background(), user)
    }
}
```

## 🔒 Security

### Data Validation

```go
// Input sanitization
func sanitizeInput(input string) string {
    return strings.TrimSpace(input)
}

// SQL injection prevention (MongoDB uses parameterized queries)
criteria := repository.NewJCriteriaBuilder().
    Where("name", repository.OpEQ, userInput). // Safe
    Build()
```

### Access Control

```go
// Role-based access control
func (r *UserRepository) FindByRole(ctx context.Context, role string) ([]schema.JRecord, error) {
    // Verify user permissions
    if !hasPermission(ctx, "read_users") {
        return nil, errors.New("insufficient permissions")
    }
    
    criteria := repository.NewJCriteriaBuilder().
        Where("role", repository.OpEQ, role).
        Build()
    
    return r.FindBy(ctx, criteria)
}
```

## 📖 Examples

### Complete User Management

```go
package main

import (
    "context"
    "log"
    "time"
    
    "github.com/kabi175/jpack/mongo"
    "github.com/kabi175/jpack/schema"
    "github.com/kabi175/jpack/validation"
    "github.com/kabi175/jpack/hooks"
)

func main() {
    // Setup
    client, err := mongo.NewMongoClient("mongodb://localhost:27017", "userdb")
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close(context.Background())

    // Create user schema
    userSchema := schema.NewSchemaBuilder("User").
        AddRequiredField("name", schema.JString, nil).
        AddRequiredUniqueField("email", schema.JString, nil).
        AddField("age", schema.JInt, 18).
        AddField("created_at", schema.JTime, nil).
        AddField("updated_at", schema.JTime, nil).
        AddValidation(func(ctx context.Context, rec schema.JRecord) error {
            return validation.ValidateEmail("email")(ctx, rec)
        }).
        AddValidation(func(ctx context.Context, rec schema.JRecord) error {
            return validation.ValidateRange("age", 0, 150)(ctx, rec)
        }).
        Build()

    // Register schema
    err = client.RegisterSchema(userSchema)
    if err != nil {
        log.Fatal(err)
    }

    // Setup hooks
    err = hooks.NewHookManager(userSchema).
        AddTimestampHook("created_at", "updated_at").
        AddValidationHook().
        AddLoggingHook(nil).
        RegisterAll()
    if err != nil {
        log.Fatal(err)
    }

    // Get repository
    userRepo, err := client.GetRepository("User")
    if err != nil {
        log.Fatal(err)
    }

    // Create user
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

    log.Printf("Saved user: %+v", savedUser.ToMap())

    // Find user
    foundUser, err := userRepo.FindById(context.Background(), "user_001")
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("Found user: %+v", foundUser.ToMap())

    // Query users
    criteria := repository.NewJCriteriaBuilder().
        WhereGreaterThanOrEqual("age", 18).
        OrderBy("name").
        SetLimit(10).
        Build()

    users, err := userRepo.FindBy(context.Background(), criteria)
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("Found %d users", len(users))

    // Update user
    updateCriteria := repository.NewJCriteriaBuilder().
        Where("id", repository.OpEQ, "user_001").
        Build()

    updates := schema.NewJRecord().Set("age", 31)
    err = userRepo.Update(context.Background(), updateCriteria, updates)
    if err != nil {
        log.Fatal(err)
    }

    // Delete user
    err = userRepo.Delete(context.Background(), "user_001")
    if err != nil {
        log.Fatal(err)
    }

    log.Println("User deleted successfully")
}
```

### E-commerce Product Catalog

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/kabi175/jpack/converter"
    "github.com/kabi175/jpack/mongo"
    "github.com/kabi175/jpack/repository"
    "github.com/kabi175/jpack/schema"
    "github.com/kabi175/jpack/validation"
)

func main() {
    client, err := mongo.NewMongoClient("mongodb://localhost:27017", "ecommerce")
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close(context.Background())

    // Register custom Money converter
    err = converter.RegisterConverter(converter.NewMoneyConverter())
    if err != nil {
        log.Fatal(err)
    }

    // Product schema with custom Money type for price
    productSchema := schema.NewSchemaBuilder("Product").
        AddRequiredField("name", schema.JString, nil).
        AddRequiredField("price", "money", nil). // Custom Money type
        AddRequiredField("category", schema.JString, nil).
        AddField("description", schema.JString, nil).
        AddField("tags", schema.JArray, nil).
        AddField("in_stock", schema.JBool, true).
        AddField("created_at", schema.JTime, nil).
        AddValidation(func(ctx context.Context, rec schema.JRecord) error {
            // Custom validation for Money type
            if price, ok := rec.Get("price").(converter.Money); ok {
                if price.Amount < 0 {
                    return fmt.Errorf("price amount cannot be negative")
                }
                if price.Currency == "" {
                    return fmt.Errorf("currency is required")
                }
            }
            return nil
        }).
        AddValidation(func(ctx context.Context, rec schema.JRecord) error {
            return validation.ValidateFieldIn("category", []any{
                "electronics", "clothing", "books", "home",
            })(ctx, rec)
        }).
        Build()

    err = client.RegisterSchema(productSchema)
    if err != nil {
        log.Fatal(err)
    }

    productRepo, err := client.GetRepository("Product")
    if err != nil {
        log.Fatal(err)
    }

    // Create products with Money type prices
    products := []schema.JRecord{
        schema.NewJRecord().
            Set("id", "prod_001").
            Set("name", "Laptop").
            Set("price", converter.Money{
                Amount:   999.99,
                Currency: "USD",
            }).
            Set("category", "electronics").
            Set("tags", []string{"laptop", "computer"}).
            Set("in_stock", true),
        schema.NewJRecord().
            Set("id", "prod_002").
            Set("name", "T-Shirt").
            Set("price", converter.Money{
                Amount:   19.99,
                Currency: "USD",
            }).
            Set("category", "clothing").
            Set("tags", []string{"clothing", "t-shirt"}).
            Set("in_stock", true),
        schema.NewJRecord().
            Set("id", "prod_003").
            Set("name", "Smartphone").
            Set("price", converter.Money{
                Amount:   699.99,
                Currency: "EUR",
            }).
            Set("category", "electronics").
            Set("tags", []string{"phone", "mobile"}).
            Set("in_stock", true),
    }

    // Save products
    for _, product := range products {
        _, err := productRepo.Save(context.Background(), product)
        if err != nil {
            log.Fatal(err)
        }
    }

    // Query products by category
    criteria := repository.NewJCriteriaBuilder().
        Where("category", repository.OpEQ, "electronics").
        OrderBy("name").
        Build()

    electronics, err := productRepo.FindBy(context.Background(), criteria)
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("Found %d electronics products", len(electronics))
    for _, product := range electronics {
        if price, ok := product.Get("price").(converter.Money); ok {
            log.Printf("Product: %s, Price: %.2f %s", 
                product.Get("name"), price.Amount, price.Currency)
        }
    }

    // Query products by currency
    usdCriteria := repository.NewJCriteriaBuilder().
        Where("price.currency", repository.OpEQ, "USD").
        OrderBy("price.amount").
        Build()

    usdProducts, err := productRepo.FindBy(context.Background(), usdCriteria)
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("Found %d USD products", len(usdProducts))
    for _, product := range usdProducts {
        if price, ok := product.Get("price").(converter.Money); ok {
            log.Printf("USD Product: %s, Price: $%.2f", 
                product.Get("name"), price.Amount)
        }
    }

    // Query products under $100 USD
    under100Criteria := repository.NewJCriteriaBuilder().
        Where("price.currency", repository.OpEQ, "USD").
        Where("price.amount", repository.OpLT, 100.0).
        Where("in_stock", repository.OpEQ, true).
        OrderBy("price.amount").
        Build()

    affordableProducts, err := productRepo.FindBy(context.Background(), under100Criteria)
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("Found %d affordable USD products under $100", len(affordableProducts))
    for _, product := range affordableProducts {
        if price, ok := product.Get("price").(converter.Money); ok {
            log.Printf("Affordable: %s, Price: $%.2f", 
                product.Get("name"), price.Amount)
        }
    }

    // Update product price with different currency
    updateCriteria := repository.NewJCriteriaBuilder().
        Where("id", repository.OpEQ, "prod_001").
        Build()

    newPrice := converter.Money{
        Amount:   1099.99,
        Currency: "EUR",
    }

    updates := schema.NewJRecord().Set("price", newPrice)
    err = productRepo.Update(context.Background(), updateCriteria, updates)
    if err != nil {
        log.Fatal(err)
    }

    log.Println("Updated laptop price to EUR")

    // Verify the update
    updatedProduct, err := productRepo.FindById(context.Background(), "prod_001")
    if err != nil {
        log.Fatal(err)
    }

    if price, ok := updatedProduct.Get("price").(converter.Money); ok {
        log.Printf("Updated Product: %s, New Price: %.2f %s", 
            updatedProduct.Get("name"), price.Amount, price.Currency)
    }
}
```

## 🤝 Contributing

### Development Setup

1. Fork the repository
2. Clone your fork
3. Install dependencies: `go mod download`
4. Run tests: `go test ./...`
5. Make your changes
6. Add tests for new functionality
7. Submit a pull request

### Code Style

- Follow Go conventions
- Use `gofmt` for formatting
- Add comprehensive tests
- Update documentation
- Use meaningful commit messages

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- MongoDB team for the excellent Go driver
- Go community for best practices and patterns
- Contributors and users who provide feedback

---

**JPack** - Making MongoDB persistence in Go simple, powerful, and type-safe.