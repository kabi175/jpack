# JPack

A Go library for building type-safe, schema-driven data access layers with MongoDB support.

## Overview

JPack is a lightweight Object-Relational Mapping (ORM) library for Go that provides:

- **Schema-first approach**: Define your data structures using a fluent schema builder
- **Type safety**: Strongly typed field validation and data conversion
- **MongoDB integration**: Built-in support for MongoDB with automatic BSON conversion
- **Flexible field types**: Extensible field type system with built-in primitives
- **Record lifecycle management**: Track modifications, validate data, and manage persistence

## Installation

```bash
go get github.com/kabi175/jpack
```

## Quick Start

### 1. Define a Schema

```go
package main

import (
    "github.com/kabi175/jpack"
)

func main() {
    // Create a user schema
    userSchema := jpack.NewSchema("users").
        Field("id", &jpack.String{}).
        Field("first_name", &jpack.String{}).
        Field("last_name", &jpack.String{}).
        Field("email", &jpack.String{}).
        Field("age", &jpack.Number{}).
        Build()
}
```

### 2. Create and Manipulate Records

```go
// Create a new MongoDB record
record := jpack.NewMongoRecord(userSchema)

// Set field values
firstNameField, _ := userSchema.Field("first_name")
lastNameField, _ := userSchema.Field("last_name")
emailField, _ := userSchema.Field("email")
ageField, _ := userSchema.Field("age")

record.SetValue(firstNameField, "John")
record.SetValue(lastNameField, "Doe")
record.SetValue(emailField, "john@example.com")
record.SetValue(ageField, 30)

// Check if record is modified
if record.IsModified() {
    fmt.Println("Record has been modified")
}

// Get dirty keys (fields that have been changed)
dirtyKeys := record.DirtyKeys()
fmt.Printf("Modified fields: %v\n", dirtyKeys)
```

### 3. Persist to MongoDB

```go
import (
    "context"
    "go.mongodb.org/mongo-driver/v2/mongo"
    "go.mongodb.org/mongo-driver/v2/mongo/options"
)

// Connect to MongoDB
client, err := mongo.Connect(options.Client().ApplyURI("mongodb://localhost:27017"))
if err != nil {
    panic(err)
}

// Create context with database connection
ctx := context.WithValue(context.Background(), jpack.Conn, client.Database("myapp"))

// Save the record
err = record.Save(ctx)
if err != nil {
    panic(err)
}

// The record now has an ID assigned
idField, _ := userSchema.Field("id")
id, exists := record.Value(idField)
if exists {
    fmt.Printf("Record saved with ID: %s\n", id)
}
```

## Core Concepts

### Schemas

Schemas define the structure and validation rules for your data:

```go
// Create a schema builder
builder := jpack.NewSchema("products")

// Add fields with types
schema := builder.
    Field("id", &jpack.String{}).
    Field("name", &jpack.String{}).
    Field("price", &jpack.Number{}).
    FieldWithDefault("status", &jpack.String{}, "active").
    Build()

// Access fields
nameField, exists := schema.Field("name")
if exists {
    fmt.Printf("Field name: %s\n", nameField.Name())
}

// List all fields
for _, field := range schema.Fields() {
    fmt.Printf("Field: %s, Type: %T\n", field.Name(), field.Type())
}
```

### Field Types

JPack includes built-in field types and allows custom implementations:

#### Built-in Types

**String Type**
- Validates string values
- Handles nil values and pointers
- Automatic type conversion for database operations

**Number Type**
- Validates integer values (int, int8, int16, int32, int64)
- Supports string representations of numbers
- Handles nil values and pointers
- Automatic conversion to int for storage

**DateTime Type**
- Validates `time.Time` values and RFC3339 formatted strings
- Automatically converts all times to GMT (UTC) timezone
- Supports timezone-aware strings (e.g., "2024-12-25T10:00:00+05:30")
- Handles nil values and pointers
- Stores all datetime values in GMT for consistency

**Boolean Type**
- Validates boolean values with flexible input conversion
- Supports string representations: "true"/"false", "1"/"0", "yes"/"no", "on"/"off", "enabled"/"disabled"
- Case-insensitive string parsing with whitespace trimming
- Converts numeric values (non-zero = true, zero = false)
- Handles nil values and pointers
- Automatic type conversion for database operations

**Options Type**
- Validates string values against a dynamic list of options from a service
- Uses `OptionService` interface to get available options with `uniqueName` and `displayName`
- `uniqueName` is used for database storage and validation
- `displayName` is used for client display purposes
- Supports dynamic option lists that can change at runtime
- Handles nil values and pointers
- Provides clear error messages for invalid options

```go
// Using built-in types
stringField := &jpack.String{}
numberField := &jpack.Number{}
dateTimeField := &jpack.DateTime{}
booleanField := &jpack.Boolean{}

// Create an options service
type StatusService struct{}
func (s *StatusService) GetOptions(ctx context.Context) ([]jpack.Option, error) {
    return []jpack.Option{
        {UniqueName: "active", DisplayName: "Active"},
        {UniqueName: "inactive", DisplayName: "Inactive"},
        {UniqueName: "pending", DisplayName: "Pending"},
    }, nil
}

// Or use the built-in in-memory service
statusOptions := []jpack.Option{
    {UniqueName: "active", DisplayName: "Active"},
    {UniqueName: "inactive", DisplayName: "Inactive"},
    {UniqueName: "pending", DisplayName: "Pending"},
}
inMemoryService := jpack.NewInMemoryOptionService(statusOptions)

optionsField := jpack.NewOptions(&StatusService{})
// or
optionsField := jpack.NewOptions(inMemoryService)

// Validate values (uses uniqueName for validation)
err := stringField.Validate("hello")        // nil - valid
err = stringField.Validate(123)             // error - not a string
err = numberField.Validate(42)              // nil - valid
err = numberField.Validate("123")           // nil - valid (converts to int)
err = dateTimeField.Validate(time.Now())    // nil - valid
err = dateTimeField.Validate("2024-12-25T10:00:00Z") // nil - valid
err = dateTimeField.Validate("2024-12-25T10:00:00+05:30") // nil - valid (converts to GMT)
err = booleanField.Validate(true)           // nil - valid
err = booleanField.Validate("yes")          // nil - valid (converts to true)
err = booleanField.Validate(1)              // nil - valid (converts to true)
err = booleanField.Validate("invalid")      // error - invalid boolean string
err = optionsField.Validate("active")       // nil - valid (uses uniqueName)
err = optionsField.Validate("invalid")      // error - not in options list
```

#### Custom Field Types

Implement the `JFieldType` interface to create custom field types:

```go
type EmailType struct{}

func (e *EmailType) Validate(value any) error {
    str, ok := value.(string)
    if !ok {
        return errors.New("email must be a string")
    }
    
    // Simple email validation
    if !strings.Contains(str, "@") {
        return errors.New("invalid email format")
    }
    
    return nil
}

func (e *EmailType) Scan(ctx context.Context, field jpack.JField, row map[string]any) (any, error) {
    v, ok := row[field.Name()]
    if !ok || v == nil {
        return nil, nil
    }
    return v.(string), nil
}

func (e *EmailType) SetValue(ctx context.Context, field jpack.JField, value any, row map[string]any) error {
    if err := e.Validate(value); err != nil {
        return err
    }
    row[field.Name()] = value
    return nil
}

// Use custom type in schema
userSchema := jpack.NewSchema("users").
    Field("email", &EmailType{}).
    Build()
```

### Records

Records represent individual data instances that conform to a schema:

```go
record := jpack.NewMongoRecord(userSchema)

// Set values
emailField, _ := userSchema.Field("email")
record.SetValue(emailField, "user@example.com")

// Get values
value, exists := record.Value(emailField)
if exists {
    fmt.Printf("Email: %s\n", value)
}

// Check record state
fmt.Printf("Is new: %v\n", record.IsNew())
fmt.Printf("Is modified: %v\n", record.IsModified())
fmt.Printf("Dirty keys: %v\n", record.DirtyKeys())

// Validate record
if err := record.Validate(); err != nil {
    fmt.Printf("Validation error: %v\n", err)
}
```

### Edges (Relationships)

Define relationships between schemas:

```go
// Define related schemas
userSchema := jpack.NewSchema("users").
    Field("id", &jpack.String{}).
    Field("name", &jpack.String{}).
    Build()

postSchema := jpack.NewSchema("posts").
    Field("id", &jpack.String{}).
    Field("title", &jpack.String{}).
    Field("user_id", &jpack.String{}).
    Build()

// Add edge to represent relationship
userIdField, _ := postSchema.Field("user_id")
postSchemaWithEdge := jpack.NewSchema("posts").
    Field("id", &jpack.String{}).
    Field("title", &jpack.String{}).
    Field("user_id", &jpack.String{}).
    Edge("user", userSchema, userIdField).
    Build()

// Access edges
edges := postSchemaWithEdge.Edge()
for _, edge := range edges {
    fmt.Printf("Edge: %s -> %s\n", edge.Name(), edge.Schema().Name())
}
```

## MongoDB Integration

JPack provides seamless MongoDB integration with automatic BSON conversion:

### Connection Setup

```go
import (
    "context"
    "go.mongodb.org/mongo-driver/v2/mongo"
    "go.mongodb.org/mongo-driver/v2/mongo/options"
)

// Connect to MongoDB
client, err := mongo.Connect(options.Client().ApplyURI("mongodb://localhost:27017"))
if err != nil {
    panic(err)
}
defer client.Disconnect(context.TODO())

// Create context with database connection
ctx := context.WithValue(context.Background(), jpack.Conn, client.Database("myapp"))
```

### Record Operations

```go
// Create new record
record := jpack.NewMongoRecord(userSchema)

// Insert new record
nameField, _ := userSchema.Field("name")
record.SetValue(nameField, "John Doe")

err := record.Save(ctx) // Inserts new document
if err != nil {
    panic(err)
}

// Update existing record
record.SetValue(nameField, "Jane Doe")
err = record.Save(ctx) // Updates existing document
if err != nil {
    panic(err)
}
```

### Automatic ID Management

JPack automatically manages MongoDB ObjectIDs:

```go
// After saving a new record
idField, _ := userSchema.Field("id")
id, exists := record.Value(idField)
if exists {
    fmt.Printf("Generated ID: %s\n", id) // Hex string representation
}

// The ID is automatically used for updates
record.SetValue(nameField, "Updated Name")
record.Save(ctx) // Uses the ID for UpdateByID operation
```

## Advanced Features

### Field Defaults

Set default values for fields:

```go
schema := jpack.NewSchema("products").
    Field("name", &jpack.String{}).
    FieldWithDefault("status", &jpack.String{}, "active").
    FieldWithDefault("quantity", &jpack.Number{}, 0).
    Build()

// Access default values
statusField, _ := schema.Field("status")
defaultStatus := statusField.Default()
fmt.Printf("Default status: %v\n", defaultStatus) // "active"
```

### Record State Tracking

Track changes to records:

```go
record := jpack.NewMongoRecord(userSchema)

// Initially, record is new and not modified
fmt.Printf("Is new: %v\n", record.IsNew())         // true
fmt.Printf("Is modified: %v\n", record.IsModified()) // false

// Set a value
nameField, _ := userSchema.Field("name")
record.SetValue(nameField, "John")

// Now record is modified
fmt.Printf("Is modified: %v\n", record.IsModified()) // true
fmt.Printf("Dirty keys: %v\n", record.DirtyKeys())   // ["name"]

// After saving
record.Save(ctx)
fmt.Printf("Is new: %v\n", record.IsNew())         // false
fmt.Printf("Is modified: %v\n", record.IsModified()) // false
```

### Error Handling

JPack provides detailed error information:

```go
// Field validation errors
numberField := &jpack.Number{}
err := numberField.Validate("not a number")
if err != nil {
    fmt.Printf("Validation error: %v\n", err)
}

// Record validation errors
record := jpack.NewMongoRecord(userSchema)
err = record.Validate()
if err != nil {
    fmt.Printf("Record validation error: %v\n", err)
}

// Save errors
err = record.Save(ctx)
if err != nil {
    fmt.Printf("Save error: %v\n", err)
}
```

## Testing

Run the test suite:

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific test files
go test -v jschema_test.go
go test -v field_types_test.go
go test -v mongodb_test.go
```

### MongoDB Testing

For MongoDB tests, ensure you have a MongoDB instance running:

```bash
# Start MongoDB (using Docker)
docker run -d -p 27017:27017 --name mongodb mongo:latest

# Run MongoDB tests
go test -v mongodb_test.go
```

## API Reference

### Core Interfaces

#### JSchema
```go
type JSchema interface {
    Name() string
    Fields() []JField
    Field(name string) (JField, bool)
    AddField(field JField) JSchema
    Edge() []JEdge
    AddEdge(edge JEdge) JSchema
    Validate(JRecord) error
}
```

#### JField
```go
type JField interface {
    Name() string
    Type() JFieldType
    Schema() JSchema
    Default() any
}
```

#### JFieldType
```go
type JFieldType interface {
    Validate(value any) error
    Scan(ctx context.Context, field JField, row map[string]any) (value any, err error)
    SetValue(ctx context.Context, field JField, value any, row map[string]any) error
}
```

#### JRecord
```go
type JRecord interface {
    Schema() JSchema
    Value(JField) (any, bool)
    SetValue(field JField, value any) error
    Fields() []JField
    IsModified() bool
    IsNew() bool
    DirtyKeys() []string
    Save(ctx context.Context) error
    Validate() error
}
```

### Schema Builder Methods

```go
func NewSchema(name string) *SchemaBuilder
func (s *SchemaBuilder) Field(name string, fType JFieldType) *SchemaBuilder
func (s *SchemaBuilder) FieldWithDefault(name string, fType JFieldType, defaultValue any) *SchemaBuilder
func (s *SchemaBuilder) Edge(name string, schema JSchema, field JField) *SchemaBuilder
func (s *SchemaBuilder) Build() JSchema
```

### MongoDB Functions

```go
func NewMongoRecord(schema JSchema) *mongoRecord
func MustConn(ctx context.Context) *mongo.Database
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass
5. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Dependencies

- [go.mongodb.org/mongo-driver/v2](https://github.com/mongodb/mongo-go-driver) - MongoDB driver
- [github.com/rs/zerolog](https://github.com/rs/zerolog) - Logging
- [github.com/samber/mo](https://github.com/samber/mo) - Monadic operations
- [github.com/stretchr/testify](https://github.com/stretchr/testify) - Testing framework

## Examples

See the `*_test.go` files for more comprehensive examples of usage patterns and best practices.