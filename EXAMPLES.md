# JPack Examples

This document provides comprehensive examples of using JPack for various use cases.

## Table of Contents

1. [Basic CRUD Operations](#basic-crud-operations)
2. [Custom Field Types](#custom-field-types)
3. [Schema Relationships](#schema-relationships)
4. [Advanced Record Management](#advanced-record-management)
5. [Error Handling](#error-handling)
6. [Testing Patterns](#testing-patterns)

## Basic CRUD Operations

### User Management System

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/kabi175/jpack"
    "go.mongodb.org/mongo-driver/v2/mongo"
    "go.mongodb.org/mongo-driver/v2/mongo/options"
)

func main() {
    // Define user schema
    userSchema := jpack.NewSchema("users").
        Field("id", &jpack.String{}).
        Field("username", &jpack.String{}).
        Field("email", &jpack.String{}).
        Field("age", &jpack.Number{}).
        FieldWithDefault("status", &jpack.String{}, "active").
        Build()

    // Setup MongoDB connection
    client, err := mongo.Connect(options.Client().ApplyURI("mongodb://localhost:27017"))
    if err != nil {
        log.Fatal(err)
    }
    defer client.Disconnect(context.TODO())

    ctx := context.WithValue(context.Background(), jpack.Conn, client.Database("example"))

    // Create a new user
    user := jpack.NewMongoRecord(userSchema)
    
    // Helper function to get fields
    getField := func(name string) jpack.JField {
        field, exists := userSchema.Field(name)
        if !exists {
            log.Fatalf("Field %s not found", name)
        }
        return field
    }

    // Set user data
    user.SetValue(getField("username"), "john_doe")
    user.SetValue(getField("email"), "john@example.com")
    user.SetValue(getField("age"), 25)

    // Save user
    if err := user.Save(ctx); err != nil {
        log.Fatal("Failed to save user:", err)
    }

    // Get the generated ID
    id, exists := user.Value(getField("id"))
    if exists {
        fmt.Printf("User created with ID: %s\n", id)
    }

    // Update user
    user.SetValue(getField("email"), "john.doe@example.com")
    user.SetValue(getField("age"), 26)

    if err := user.Save(ctx); err != nil {
        log.Fatal("Failed to update user:", err)
    }

    fmt.Printf("User updated. Modified fields: %v\n", user.DirtyKeys())
}
```

### Product Catalog

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/kabi175/jpack"
    "go.mongodb.org/mongo-driver/v2/mongo"
    "go.mongodb.org/mongo-driver/v2/mongo/options"
)

func main() {
    // Define product schema
    productSchema := jpack.NewSchema("products").
        Field("id", &jpack.String{}).
        Field("name", &jpack.String{}).
        Field("description", &jpack.String{}).
        Field("price", &jpack.Number{}).
        Field("stock_quantity", &jpack.Number{}).
        FieldWithDefault("status", &jpack.String{}, "available").
        Build()

    // Setup context
    client, err := mongo.Connect(options.Client().ApplyURI("mongodb://localhost:27017"))
    if err != nil {
        log.Fatal(err)
    }
    defer client.Disconnect(context.TODO())

    ctx := context.WithValue(context.Background(), jpack.Conn, client.Database("catalog"))

    // Create products
    products := []map[string]any{
        {
            "name":           "Laptop",
            "description":    "High-performance laptop",
            "price":          999,
            "stock_quantity": 10,
        },
        {
            "name":           "Mouse",
            "description":    "Wireless mouse",
            "price":          25,
            "stock_quantity": 100,
        },
        {
            "name":           "Keyboard",
            "description":    "Mechanical keyboard",
            "price":          75,
            "stock_quantity": 50,
        },
    }

    for _, productData := range products {
        product := jpack.NewMongoRecord(productSchema)
        
        for fieldName, value := range productData {
            field, exists := productSchema.Field(fieldName)
            if exists {
                product.SetValue(field, value)
            }
        }

        if err := product.Save(ctx); err != nil {
            log.Printf("Failed to save product %s: %v", productData["name"], err)
            continue
        }

        id, _ := product.Value(getField(productSchema, "id"))
        fmt.Printf("Product '%s' created with ID: %s\n", productData["name"], id)
    }
}

func getField(schema jpack.JSchema, name string) jpack.JField {
    field, exists := schema.Field(name)
    if !exists {
        log.Fatalf("Field %s not found", name)
    }
    return field
}
```

## Custom Field Types

### Email Field Type

```go
package main

import (
    "context"
    "errors"
    "fmt"
    "regexp"
    "strings"
    
    "github.com/kabi175/jpack"
)

// EmailType implements JFieldType for email validation
type EmailType struct{}

func (e *EmailType) Validate(value any) error {
    if value == nil {
        return nil // Allow nil values
    }
    
    str, ok := value.(string)
    if !ok {
        return errors.New("email must be a string")
    }
    
    // Basic email validation
    emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
    if !emailRegex.MatchString(str) {
        return errors.New("invalid email format")
    }
    
    return nil
}

func (e *EmailType) Scan(ctx context.Context, field jpack.JField, row map[string]any) (any, error) {
    v, ok := row[field.Name()]
    if !ok || v == nil {
        return nil, nil
    }
    
    str, ok := v.(string)
    if !ok {
        return nil, errors.New("email field must be a string")
    }
    
    return strings.ToLower(str), nil // Store emails in lowercase
}

func (e *EmailType) SetValue(ctx context.Context, field jpack.JField, value any, row map[string]any) error {
    if value == nil {
        row[field.Name()] = nil
        return nil
    }
    
    if err := e.Validate(value); err != nil {
        return err
    }
    
    str := value.(string)
    row[field.Name()] = strings.ToLower(str) // Store in lowercase
    return nil
}

// Usage example
func main() {
    userSchema := jpack.NewSchema("users").
        Field("id", &jpack.String{}).
        Field("name", &jpack.String{}).
        Field("email", &EmailType{}).
        Build()

    user := jpack.NewMongoRecord(userSchema)
    
    nameField, _ := userSchema.Field("name")
    emailField, _ := userSchema.Field("email")
    
    user.SetValue(nameField, "John Doe")
    
    // This will be stored in lowercase
    user.SetValue(emailField, "John.Doe@Example.COM")
    
    // Validate
    if err := user.Validate(); err != nil {
        fmt.Printf("Validation error: %v\n", err)
    } else {
        fmt.Println("User data is valid")
    }
}
```

### Enum Field Type

```go
package main

import (
    "context"
    "errors"
    "fmt"
    "slices"
    
    "github.com/kabi175/jpack"
)

// EnumType implements JFieldType for enum validation
type EnumType struct {
    AllowedValues []string
}

func (e *EnumType) Validate(value any) error {
    if value == nil {
        return nil // Allow nil values
    }
    
    str, ok := value.(string)
    if !ok {
        return errors.New("enum value must be a string")
    }
    
    if !slices.Contains(e.AllowedValues, str) {
        return fmt.Errorf("value '%s' is not allowed. Valid values: %v", str, e.AllowedValues)
    }
    
    return nil
}

func (e *EnumType) Scan(ctx context.Context, field jpack.JField, row map[string]any) (any, error) {
    v, ok := row[field.Name()]
    if !ok || v == nil {
        return nil, nil
    }
    
    str, ok := v.(string)
    if !ok {
        return nil, errors.New("enum field must be a string")
    }
    
    return str, nil
}

func (e *EnumType) SetValue(ctx context.Context, field jpack.JField, value any, row map[string]any) error {
    if err := e.Validate(value); err != nil {
        return err
    }
    
    row[field.Name()] = value
    return nil
}

// Usage example
func main() {
    statusEnum := &EnumType{
        AllowedValues: []string{"active", "inactive", "pending", "suspended"},
    }
    
    priorityEnum := &EnumType{
        AllowedValues: []string{"low", "medium", "high", "urgent"},
    }
    
    taskSchema := jpack.NewSchema("tasks").
        Field("id", &jpack.String{}).
        Field("title", &jpack.String{}).
        Field("status", statusEnum).
        Field("priority", priorityEnum).
        Build()

    task := jpack.NewMongoRecord(taskSchema)
    
    titleField, _ := taskSchema.Field("title")
    statusField, _ := taskSchema.Field("status")
    priorityField, _ := taskSchema.Field("priority")
    
    task.SetValue(titleField, "Complete documentation")
    task.SetValue(statusField, "active")
    task.SetValue(priorityField, "high")
    
    // This would cause a validation error
    // task.SetValue(statusField, "invalid_status")
    
    if err := task.Validate(); err != nil {
        fmt.Printf("Validation error: %v\n", err)
    } else {
        fmt.Println("Task data is valid")
    }
}
```

### Date Field Type

```go
package main

import (
    "context"
    "errors"
    "fmt"
    "time"
    
    "github.com/kabi175/jpack"
)

// DateType implements JFieldType for date handling
type DateType struct{}

func (d *DateType) Validate(value any) error {
    if value == nil {
        return nil
    }
    
    switch v := value.(type) {
    case time.Time:
        return nil
    case string:
        _, err := time.Parse(time.RFC3339, v)
        if err != nil {
            return fmt.Errorf("invalid date format, expected RFC3339: %v", err)
        }
        return nil
    default:
        return errors.New("date must be a time.Time or RFC3339 string")
    }
}

func (d *DateType) Scan(ctx context.Context, field jpack.JField, row map[string]any) (any, error) {
    v, ok := row[field.Name()]
    if !ok || v == nil {
        return nil, nil
    }
    
    switch val := v.(type) {
    case time.Time:
        return val, nil
    case string:
        return time.Parse(time.RFC3339, val)
    default:
        return nil, errors.New("date field must be a time.Time or string")
    }
}

func (d *DateType) SetValue(ctx context.Context, field jpack.JField, value any, row map[string]any) error {
    if value == nil {
        row[field.Name()] = nil
        return nil
    }
    
    if err := d.Validate(value); err != nil {
        return err
    }
    
    switch v := value.(type) {
    case time.Time:
        row[field.Name()] = v
    case string:
        t, err := time.Parse(time.RFC3339, v)
        if err != nil {
            return err
        }
        row[field.Name()] = t
    }
    
    return nil
}

// Usage example
func main() {
    eventSchema := jpack.NewSchema("events").
        Field("id", &jpack.String{}).
        Field("name", &jpack.String{}).
        Field("created_at", &DateType{}).
        Field("scheduled_at", &DateType{}).
        Build()

    event := jpack.NewMongoRecord(eventSchema)
    
    nameField, _ := eventSchema.Field("name")
    createdAtField, _ := eventSchema.Field("created_at")
    scheduledAtField, _ := eventSchema.Field("scheduled_at")
    
    event.SetValue(nameField, "Team Meeting")
    event.SetValue(createdAtField, time.Now())
    event.SetValue(scheduledAtField, "2024-12-25T10:00:00Z")
    
    if err := event.Validate(); err != nil {
        fmt.Printf("Validation error: %v\n", err)
    } else {
        fmt.Println("Event data is valid")
    }
}
```

## Advanced Record Management

### Batch Operations

```go
package main

import (
    "context"
    "fmt"
    "log"
    "sync"
    
    "github.com/kabi175/jpack"
    "go.mongodb.org/mongo-driver/v2/mongo"
    "go.mongodb.org/mongo-driver/v2/mongo/options"
)

func main() {
    // Setup
    productSchema := jpack.NewSchema("products").
        Field("id", &jpack.String{}).
        Field("name", &jpack.String{}).
        Field("price", &jpack.Number{}).
        Field("category", &jpack.String{}).
        Build()

    client, err := mongo.Connect(options.Client().ApplyURI("mongodb://localhost:27017"))
    if err != nil {
        log.Fatal(err)
    }
    defer client.Disconnect(context.TODO())

    ctx := context.WithValue(context.Background(), jpack.Conn, client.Database("inventory"))

    // Batch create products
    products := []map[string]any{
        {"name": "Laptop", "price": 999, "category": "Electronics"},
        {"name": "Mouse", "price": 25, "category": "Electronics"},
        {"name": "Desk", "price": 150, "category": "Furniture"},
        {"name": "Chair", "price": 200, "category": "Furniture"},
        {"name": "Monitor", "price": 300, "category": "Electronics"},
    }

    var wg sync.WaitGroup
    var mu sync.Mutex
    var createdIds []string

    for _, productData := range products {
        wg.Add(1)
        go func(data map[string]any) {
            defer wg.Done()
            
            product := jpack.NewMongoRecord(productSchema)
            for fieldName, value := range data {
                if field, exists := productSchema.Field(fieldName); exists {
                    product.SetValue(field, value)
                }
            }

            if err := product.Save(ctx); err != nil {
                log.Printf("Failed to save product %s: %v", data["name"], err)
                return
            }

            id, _ := product.Value(getField(productSchema, "id"))
            mu.Lock()
            createdIds = append(createdIds, id.(string))
            mu.Unlock()
            
            fmt.Printf("Product '%s' created with ID: %s\n", data["name"], id)
        }(productData)
    }

    wg.Wait()
    fmt.Printf("Created %d products\n", len(createdIds))
}

func getField(schema jpack.JSchema, name string) jpack.JField {
    field, exists := schema.Field(name)
    if !exists {
        log.Fatalf("Field %s not found", name)
    }
    return field
}
```

### Transaction-like Operations

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/kabi175/jpack"
    "go.mongodb.org/mongo-driver/v2/mongo"
    "go.mongodb.org/mongo-driver/v2/mongo/options"
)

func main() {
    // Setup schemas
    accountSchema := jpack.NewSchema("accounts").
        Field("id", &jpack.String{}).
        Field("name", &jpack.String{}).
        Field("balance", &jpack.Number{}).
        Build()

    transactionSchema := jpack.NewSchema("transactions").
        Field("id", &jpack.String{}).
        Field("from_account", &jpack.String{}).
        Field("to_account", &jpack.String{}).
        Field("amount", &jpack.Number{}).
        Field("description", &jpack.String{}).
        Build()

    client, err := mongo.Connect(options.Client().ApplyURI("mongodb://localhost:27017"))
    if err != nil {
        log.Fatal(err)
    }
    defer client.Disconnect(context.TODO())

    ctx := context.WithValue(context.Background(), jpack.Conn, client.Database("banking"))

    // Create accounts
    accounts := []map[string]any{
        {"name": "John Doe", "balance": 1000},
        {"name": "Jane Smith", "balance": 500},
    }

    var accountIds []string
    for _, accountData := range accounts {
        account := jpack.NewMongoRecord(accountSchema)
        account.SetValue(getField(accountSchema, "name"), accountData["name"])
        account.SetValue(getField(accountSchema, "balance"), accountData["balance"])
        
        if err := account.Save(ctx); err != nil {
            log.Fatal("Failed to save account:", err)
        }

        id, _ := account.Value(getField(accountSchema, "id"))
        accountIds = append(accountIds, id.(string))
        fmt.Printf("Account '%s' created with ID: %s\n", accountData["name"], id)
    }

    // Perform transfer operation
    transferAmount := 200
    fromAccountId := accountIds[0]
    toAccountId := accountIds[1]

    // Create transaction record
    transaction := jpack.NewMongoRecord(transactionSchema)
    transaction.SetValue(getField(transactionSchema, "from_account"), fromAccountId)
    transaction.SetValue(getField(transactionSchema, "to_account"), toAccountId)
    transaction.SetValue(getField(transactionSchema, "amount"), transferAmount)
    transaction.SetValue(getField(transactionSchema, "description"), "Transfer between accounts")

    if err := transaction.Save(ctx); err != nil {
        log.Fatal("Failed to save transaction:", err)
    }

    transactionId, _ := transaction.Value(getField(transactionSchema, "id"))
    fmt.Printf("Transaction created with ID: %s\n", transactionId)

    // Note: In a real application, you would use MongoDB transactions
    // to ensure atomicity of the balance updates
    fmt.Printf("Transfer of %d completed from %s to %s\n", 
        transferAmount, fromAccountId, toAccountId)
}

func getField(schema jpack.JSchema, name string) jpack.JField {
    field, exists := schema.Field(name)
    if !exists {
        log.Fatalf("Field %s not found", name)
    }
    return field
}
```

## Error Handling

### Comprehensive Error Handling

```go
package main

import (
    "context"
    "errors"
    "fmt"
    "log"
    
    "github.com/kabi175/jpack"
    "go.mongodb.org/mongo-driver/v2/mongo"
    "go.mongodb.org/mongo-driver/v2/mongo/options"
)

// Custom error types
var (
    ErrInvalidData = errors.New("invalid data")
    ErrSaveFailure = errors.New("save operation failed")
)

func main() {
    userSchema := jpack.NewSchema("users").
        Field("id", &jpack.String{}).
        Field("name", &jpack.String{}).
        Field("email", &jpack.String{}).
        Field("age", &jpack.Number{}).
        Build()

    client, err := mongo.Connect(options.Client().ApplyURI("mongodb://localhost:27017"))
    if err != nil {
        log.Fatal(err)
    }
    defer client.Disconnect(context.TODO())

    ctx := context.WithValue(context.Background(), jpack.Conn, client.Database("example"))

    // Example 1: Field validation errors
    user := jpack.NewMongoRecord(userSchema)
    
    if err := setUserData(user, userSchema, "John Doe", "invalid-email", "not-a-number"); err != nil {
        fmt.Printf("Data validation error: %v\n", err)
    }

    // Example 2: Successful operation with proper error handling
    user2 := jpack.NewMongoRecord(userSchema)
    if err := setUserData(user2, userSchema, "Jane Smith", "jane@example.com", 25); err != nil {
        fmt.Printf("Error setting user data: %v\n", err)
        return
    }

    if err := saveUserSafely(ctx, user2); err != nil {
        fmt.Printf("Error saving user: %v\n", err)
        return
    }

    fmt.Println("User saved successfully!")
}

func setUserData(record jpack.JRecord, schema jpack.JSchema, name string, email string, age any) error {
    nameField, exists := schema.Field("name")
    if !exists {
        return fmt.Errorf("name field not found in schema")
    }

    emailField, exists := schema.Field("email")
    if !exists {
        return fmt.Errorf("email field not found in schema")
    }

    ageField, exists := schema.Field("age")
    if !exists {
        return fmt.Errorf("age field not found in schema")
    }

    // Validate and set name
    if err := nameField.Type().Validate(name); err != nil {
        return fmt.Errorf("invalid name: %w", err)
    }
    if err := record.SetValue(nameField, name); err != nil {
        return fmt.Errorf("failed to set name: %w", err)
    }

    // Validate and set email
    if err := emailField.Type().Validate(email); err != nil {
        return fmt.Errorf("invalid email: %w", err)
    }
    if err := record.SetValue(emailField, email); err != nil {
        return fmt.Errorf("failed to set email: %w", err)
    }

    // Validate and set age
    if err := ageField.Type().Validate(age); err != nil {
        return fmt.Errorf("invalid age: %w", err)
    }
    if err := record.SetValue(ageField, age); err != nil {
        return fmt.Errorf("failed to set age: %w", err)
    }

    return nil
}

func saveUserSafely(ctx context.Context, record jpack.JRecord) error {
    // Validate before saving
    if err := record.Validate(); err != nil {
        return fmt.Errorf("record validation failed: %w", err)
    }

    // Check if record has required fields
    if !record.IsModified() && record.IsNew() {
        return fmt.Errorf("cannot save empty record")
    }

    // Attempt to save
    if err := record.Save(ctx); err != nil {
        // Check if it's a connection error
        if mongo.IsNetworkError(err) {
            return fmt.Errorf("network error while saving: %w", err)
        }
        
        // Check if it's a write concern error
        if mongo.IsTimeout(err) {
            return fmt.Errorf("timeout error while saving: %w", err)
        }
        
        return fmt.Errorf("failed to save record: %w", err)
    }

    return nil
}
```

## Testing Patterns

### Unit Testing with JPack

```go
package main

import (
    "context"
    "testing"
    
    "github.com/kabi175/jpack"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestUserSchema(t *testing.T) {
    userSchema := jpack.NewSchema("users").
        Field("id", &jpack.String{}).
        Field("name", &jpack.String{}).
        Field("email", &jpack.String{}).
        Field("age", &jpack.Number{}).
        Build()

    t.Run("schema creation", func(t *testing.T) {
        assert.Equal(t, "users", userSchema.Name())
        assert.Len(t, userSchema.Fields(), 4)
    })

    t.Run("field access", func(t *testing.T) {
        nameField, exists := userSchema.Field("name")
        assert.True(t, exists)
        assert.Equal(t, "name", nameField.Name())
        assert.IsType(t, &jpack.String{}, nameField.Type())
    })

    t.Run("non-existent field", func(t *testing.T) {
        _, exists := userSchema.Field("non_existent")
        assert.False(t, exists)
    })
}

func TestUserRecord(t *testing.T) {
    userSchema := jpack.NewSchema("users").
        Field("id", &jpack.String{}).
        Field("name", &jpack.String{}).
        Field("email", &jpack.String{}).
        Field("age", &jpack.Number{}).
        Build()

    t.Run("record creation", func(t *testing.T) {
        record := jpack.NewMongoRecord(userSchema)
        
        assert.True(t, record.IsNew())
        assert.False(t, record.IsModified())
        assert.Empty(t, record.DirtyKeys())
    })

    t.Run("setting values", func(t *testing.T) {
        record := jpack.NewMongoRecord(userSchema)
        
        nameField, _ := userSchema.Field("name")
        err := record.SetValue(nameField, "John Doe")
        require.NoError(t, err)
        
        assert.True(t, record.IsModified())
        assert.Contains(t, record.DirtyKeys(), "name")
        
        value, exists := record.Value(nameField)
        assert.True(t, exists)
        assert.Equal(t, "John Doe", value)
    })

    t.Run("field validation", func(t *testing.T) {
        record := jpack.NewMongoRecord(userSchema)
        
        ageField, _ := userSchema.Field("age")
        
        // Valid age
        err := record.SetValue(ageField, 25)
        assert.NoError(t, err)
        
        // Invalid age (string that can't be converted)
        err = record.SetValue(ageField, "not-a-number")
        assert.Error(t, err)
    })
}

func TestFieldTypes(t *testing.T) {
    t.Run("string field type", func(t *testing.T) {
        stringType := &jpack.String{}
        
        // Valid string
        err := stringType.Validate("hello")
        assert.NoError(t, err)
        
        // Invalid type
        err = stringType.Validate(123)
        assert.Error(t, err)
        
        // Nil value (should be allowed)
        err = stringType.Validate(nil)
        assert.NoError(t, err)
    })

    t.Run("number field type", func(t *testing.T) {
        numberType := &jpack.Number{}
        
        // Valid integer
        err := numberType.Validate(42)
        assert.NoError(t, err)
        
        // Valid string number
        err = numberType.Validate("42")
        assert.NoError(t, err)
        
        // Invalid string
        err = numberType.Validate("not-a-number")
        assert.Error(t, err)
        
        // Nil value (should be allowed)
        err = numberType.Validate(nil)
        assert.NoError(t, err)
    })
}

func TestSchemaBuilder(t *testing.T) {
    t.Run("fluent interface", func(t *testing.T) {
        schema := jpack.NewSchema("test").
            Field("field1", &jpack.String{}).
            Field("field2", &jpack.Number{}).
            FieldWithDefault("field3", &jpack.String{}, "default").
            Build()
        
        assert.Equal(t, "test", schema.Name())
        assert.Len(t, schema.Fields(), 3)
        
        field3, exists := schema.Field("field3")
        assert.True(t, exists)
        assert.Equal(t, "default", field3.Default())
    })

    t.Run("duplicate field names", func(t *testing.T) {
        schema := jpack.NewSchema("test").
            Field("name", &jpack.String{}).
            Field("name", &jpack.Number{}). // Should be ignored
            Build()
        
        assert.Len(t, schema.Fields(), 1)
        
        nameField, exists := schema.Field("name")
        assert.True(t, exists)
        assert.IsType(t, &jpack.String{}, nameField.Type()) // Should be the first one
    })
}

// Mock field type for testing
type MockFieldType struct {
    ValidateFunc func(value any) error
    ScanFunc     func(ctx context.Context, field jpack.JField, row map[string]any) (any, error)
    SetValueFunc func(ctx context.Context, field jpack.JField, value any, row map[string]any) error
}

func (m *MockFieldType) Validate(value any) error {
    if m.ValidateFunc != nil {
        return m.ValidateFunc(value)
    }
    return nil
}

func (m *MockFieldType) Scan(ctx context.Context, field jpack.JField, row map[string]any) (any, error) {
    if m.ScanFunc != nil {
        return m.ScanFunc(ctx, field, row)
    }
    return nil, nil
}

func (m *MockFieldType) SetValue(ctx context.Context, field jpack.JField, value any, row map[string]any) error {
    if m.SetValueFunc != nil {
        return m.SetValueFunc(ctx, field, value, row)
    }
    return nil
}

func TestCustomFieldType(t *testing.T) {
    t.Run("mock field type", func(t *testing.T) {
        mockType := &MockFieldType{
            ValidateFunc: func(value any) error {
                if value == "invalid" {
                    return errors.New("invalid value")
                }
                return nil
            },
        }
        
        err := mockType.Validate("valid")
        assert.NoError(t, err)
        
        err = mockType.Validate("invalid")
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "invalid value")
    })
}
```

### Integration Testing

```go
package main

import (
    "context"
    "testing"
    "time"
    
    "github.com/kabi175/jpack"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "go.mongodb.org/mongo-driver/v2/mongo"
    "go.mongodb.org/mongo-driver/v2/mongo/options"
)

func TestMongoIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }

    // Setup MongoDB connection
    client, err := mongo.Connect(options.Client().ApplyURI("mongodb://localhost:27017"))
    require.NoError(t, err)
    defer client.Disconnect(context.TODO())

    // Use test database
    dbName := "jpack_test_" + time.Now().Format("20060102_150405")
    db := client.Database(dbName)
    defer db.Drop(context.TODO())

    ctx := context.WithValue(context.Background(), jpack.Conn, db)

    userSchema := jpack.NewSchema("users").
        Field("id", &jpack.String{}).
        Field("name", &jpack.String{}).
        Field("email", &jpack.String{}).
        Field("age", &jpack.Number{}).
        Build()

    t.Run("create and save user", func(t *testing.T) {
        user := jpack.NewMongoRecord(userSchema)
        
        nameField, _ := userSchema.Field("name")
        emailField, _ := userSchema.Field("email")
        ageField, _ := userSchema.Field("age")
        
        user.SetValue(nameField, "John Doe")
        user.SetValue(emailField, "john@example.com")
        user.SetValue(ageField, 30)
        
        assert.True(t, user.IsNew())
        assert.True(t, user.IsModified())
        
        err := user.Save(ctx)
        require.NoError(t, err)
        
        assert.False(t, user.IsNew())
        assert.False(t, user.IsModified())
        
        // Check that ID was assigned
        idField, _ := userSchema.Field("id")
        id, exists := user.Value(idField)
        assert.True(t, exists)
        assert.NotEmpty(t, id)
    })

    t.Run("update user", func(t *testing.T) {
        user := jpack.NewMongoRecord(userSchema)
        
        nameField, _ := userSchema.Field("name")
        emailField, _ := userSchema.Field("email")
        ageField, _ := userSchema.Field("age")
        
        user.SetValue(nameField, "Jane Smith")
        user.SetValue(emailField, "jane@example.com")
        user.SetValue(ageField, 25)
        
        err := user.Save(ctx)
        require.NoError(t, err)
        
        // Update the user
        user.SetValue(emailField, "jane.smith@example.com")
        user.SetValue(ageField, 26)
        
        assert.True(t, user.IsModified())
        assert.ElementsMatch(t, []string{"email", "age"}, user.DirtyKeys())
        
        err = user.Save(ctx)
        require.NoError(t, err)
        
        assert.False(t, user.IsModified())
        assert.Empty(t, user.DirtyKeys())
    })
}

func TestBenchmarks(t *testing.T) {
    userSchema := jpack.NewSchema("users").
        Field("id", &jpack.String{}).
        Field("name", &jpack.String{}).
        Field("email", &jpack.String{}).
        Field("age", &jpack.Number{}).
        Build()

    t.Run("benchmark record creation", func(t *testing.T) {
        b := testing.Benchmark(func(b *testing.B) {
            for i := 0; i < b.N; i++ {
                record := jpack.NewMongoRecord(userSchema)
                _ = record
            }
        })
        
        t.Logf("Record creation: %v ns/op", b.NsPerOp())
    })

    t.Run("benchmark field access", func(t *testing.T) {
        record := jpack.NewMongoRecord(userSchema)
        nameField, _ := userSchema.Field("name")
        record.SetValue(nameField, "John Doe")
        
        b := testing.Benchmark(func(b *testing.B) {
            for i := 0; i < b.N; i++ {
                _, _ = record.Value(nameField)
            }
        })
        
        t.Logf("Field access: %v ns/op", b.NsPerOp())
    })
}
```

This comprehensive examples file demonstrates various use cases and patterns for using JPack effectively in real-world applications.
