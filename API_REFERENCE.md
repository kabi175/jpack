# JPack API Reference

This document provides a comprehensive reference for all JPack interfaces, types, and functions.

## Table of Contents

1. [Core Interfaces](#core-interfaces)
2. [Schema Building](#schema-building)
3. [Field Types](#field-types)
4. [Record Operations](#record-operations)
5. [MongoDB Integration](#mongodb-integration)
6. [Constants and Variables](#constants-and-variables)

## Core Interfaces

### JSchema

The `JSchema` interface defines the contract for schema definitions.

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

#### Methods

- **`Name() string`** - Returns the schema name
- **`Fields() []JField`** - Returns all fields in the schema
- **`Field(name string) (JField, bool)`** - Returns a field by name and existence flag
- **`AddField(field JField) JSchema`** - Adds a field to the schema
- **`Edge() []JEdge`** - Returns all edges (relationships) in the schema
- **`AddEdge(edge JEdge) JSchema`** - Adds an edge to the schema
- **`Validate(JRecord) error`** - Validates a record against the schema

### JField

The `JField` interface represents a field in a schema.

```go
type JField interface {
    Name() string
    Type() JFieldType
    Schema() JSchema
    Default() any
}
```

#### Methods

- **`Name() string`** - Returns the field name
- **`Type() JFieldType`** - Returns the field type
- **`Schema() JSchema`** - Returns the schema this field belongs to
- **`Default() any`** - Returns the default value for the field

### JFieldType

The `JFieldType` interface defines the contract for field type implementations.

```go
type JFieldType interface {
    Validate(value any) error
    Scan(ctx context.Context, field JField, row map[string]any) (value any, err error)
    SetValue(ctx context.Context, field JField, value any, row map[string]any) error
}
```

#### Methods

- **`Validate(value any) error`** - Validates a value for this field type
- **`Scan(ctx context.Context, field JField, row map[string]any) (value any, err error)`** - Converts a database value to the field's type
- **`SetValue(ctx context.Context, field JField, value any, row map[string]any) error`** - Sets a value in the database row

### JRecord

The `JRecord` interface represents a record instance.

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

#### Methods

- **`Schema() JSchema`** - Returns the schema for this record
- **`Value(JField) (any, bool)`** - Gets a field value and existence flag
- **`SetValue(field JField, value any) error`** - Sets a field value
- **`Fields() []JField`** - Returns all fields that have values in this record
- **`IsModified() bool`** - Returns true if the record has been modified
- **`IsNew() bool`** - Returns true if this is a new record (not yet saved)
- **`DirtyKeys() []string`** - Returns field names that have been modified
- **`Save(ctx context.Context) error`** - Saves the record to the database
- **`Validate() error`** - Validates the record

### JEdge

The `JEdge` interface represents a relationship between schemas.

```go
type JEdge interface {
    Name() string
    Schema() JSchema
    Field() JField
}
```

#### Methods

- **`Name() string`** - Returns the edge name
- **`Schema() JSchema`** - Returns the target schema
- **`Field() JField`** - Returns the field that defines the relationship

### JPolicy

The `JPolicy` interface defines access control policies.

```go
type JPolicy interface {
    IsValid(ctx context.Context, record JRecord) error
}
```

#### Methods

- **`IsValid(ctx context.Context, record JRecord) error`** - Validates if a record is valid according to this policy

## Schema Building

### SchemaBuilder

The `SchemaBuilder` struct provides a fluent interface for building schemas.

```go
type SchemaBuilder struct {
    // fields are private
}
```

#### Methods

- **`Field(name string, fType JFieldType) *SchemaBuilder`** - Adds a field to the schema
- **`FieldWithDefault(name string, fType JFieldType, defaultValue any) *SchemaBuilder`** - Adds a field with a default value
- **`Edge(name string, schema JSchema, field JField) *SchemaBuilder`** - Adds an edge to the schema
- **`Build() JSchema`** - Builds and returns the final schema

### Functions

#### NewSchema

```go
func NewSchema(name string) *SchemaBuilder
```

Creates a new schema builder with the given name.

**Parameters:**
- `name` - The schema name

**Returns:**
- `*SchemaBuilder` - A new schema builder instance

**Example:**
```go
schema := jpack.NewSchema("users").
    Field("id", &jpack.String{}).
    Field("name", &jpack.String{}).
    Build()
```

## Field Types

### String

The `String` type handles string values.

```go
type String struct{}
```

#### Methods

- **`Validate(value any) error`** - Validates that the value is a string or nil
- **`Scan(ctx context.Context, field JField, row map[string]any) (any, error)`** - Reads a string value from the database
- **`SetValue(ctx context.Context, field JField, value any, row map[string]any) error`** - Sets a string value in the database

**Validation Rules:**
- Accepts `string` values
- Accepts `nil` values
- Accepts pointer to string (dereferenced)
- Rejects structs, arrays, slices, functions, channels, maps

**Usage:**
```go
stringField := &jpack.String{}
err := stringField.Validate("hello") // nil
err := stringField.Validate(123)     // error
```

### Number

The `Number` type handles integer values.

```go
type Number struct{}
```

#### Methods

- **`Validate(value any) error`** - Validates that the value is a number or can be converted to one
- **`Scan(ctx context.Context, field JField, row map[string]any) (any, error)`** - Reads a number value from the database
- **`SetValue(ctx context.Context, field JField, value any, row map[string]any) error`** - Sets a number value in the database

**Validation Rules:**
- Accepts `int`, `int8`, `int16`, `int32`, `int64` values
- Accepts string representations of integers
- Accepts pointer to number (dereferenced)
- Accepts `nil` values
- Converts all numeric types to `int` for storage

**Usage:**
```go
numberField := &jpack.Number{}
err := numberField.Validate(42)     // nil
err := numberField.Validate("123")  // nil (converts to int)
err := numberField.Validate("abc")  // error
```

### DateTime

The `DateTime` type handles datetime values with automatic GMT timezone conversion.

```go
type DateTime struct{}
```

#### Methods

- **`Validate(value any) error`** - Validates that the value is a valid datetime
- **`Scan(ctx context.Context, field JField, row map[string]any) (any, error)`** - Reads a datetime value from the database
- **`SetValue(ctx context.Context, field JField, value any, row map[string]any) error`** - Sets a datetime value in the database

**Validation Rules:**
- Accepts `time.Time` values
- Accepts RFC3339 formatted strings (e.g., "2024-12-25T10:00:00Z")
- Accepts RFC3339 strings with timezone offsets (e.g., "2024-12-25T10:00:00+05:30")
- Accepts pointer to time.Time (dereferenced)
- Accepts `nil` values
- Automatically converts all times to GMT (UTC) timezone for storage

**Usage:**
```go
dateTimeField := &jpack.DateTime{}
err := dateTimeField.Validate(time.Now())                    // nil
err := dateTimeField.Validate("2024-12-25T10:00:00Z")       // nil
err := dateTimeField.Validate("2024-12-25T10:00:00+05:30")  // nil (converts to GMT)
err := dateTimeField.Validate("2024-12-25 10:00:00")        // error (invalid format)
err := dateTimeField.Validate(123)                          // error (invalid type)
```

### Boolean

The `Boolean` type handles boolean values with flexible input conversion.

```go
type Boolean struct{}
```

#### Methods

- **`Validate(value any) error`** - Validates that the value is a valid boolean
- **`Scan(ctx context.Context, field JField, row map[string]any) (any, error)`** - Reads a boolean value from the database
- **`SetValue(ctx context.Context, field JField, value any, row map[string]any) error`** - Sets a boolean value in the database

**Validation Rules:**
- Accepts `bool` values
- Accepts string representations: "true"/"false", "1"/"0", "yes"/"no", "on"/"off", "enabled"/"disabled"
- Case-insensitive string parsing with whitespace trimming
- Converts numeric values (non-zero = true, zero = false)
- Accepts pointer to bool (dereferenced)
- Accepts `nil` values
- Rejects invalid string values and unsupported types

**Usage:**
```go
booleanField := &jpack.Boolean{}
err := booleanField.Validate(true)           // nil
err := booleanField.Validate("yes")          // nil (converts to true)
err := booleanField.Validate("TRUE")         // nil (converts to true)
err := booleanField.Validate(1)              // nil (converts to true)
err := booleanField.Validate(0)              // nil (converts to false)
err := booleanField.Validate("invalid")      // error (invalid boolean string)
err := booleanField.Validate([]string{})     // error (unsupported type)
```

### Options

The `Options` type handles enum values with dynamic options from a service.

```go
type Options struct {
    service OptionService
}
```

#### Option Struct

```go
type Option struct {
    UniqueName  string `json:"uniqueName"`
    DisplayName string `json:"displayName"`
}
```

#### OptionService Interface

```go
type OptionService interface {
    GetOptions(ctx context.Context) ([]Option, error)
}
```

#### Functions

#### NewOptions

```go
func NewOptions(service OptionService) *Options
```

Creates a new Options FieldType with the given service.

**Parameters:**
- `service` - The service that provides available options

**Returns:**
- `*Options` - A new Options FieldType instance

#### Methods

- **`Validate(value any) error`** - Validates that the value is in the list of available options (uses uniqueName)
- **`Scan(ctx context.Context, field JField, row map[string]any) (any, error)`** - Reads an options value from the database
- **`SetValue(ctx context.Context, field JField, value any, row map[string]any) error`** - Sets an options value in the database
- **`GetDisplayName(ctx context.Context, uniqueName string) (string, error)`** - Gets display name for a unique name
- **`GetUniqueName(ctx context.Context, displayName string) (string, error)`** - Gets unique name for a display name
- **`GetAllOptions(ctx context.Context) ([]Option, error)`** - Gets all available options

**Validation Rules:**
- Accepts `string` values that match the `uniqueName` in the service's options list
- Accepts `nil` values
- Accepts pointer to string (dereferenced)
- Rejects values not in the service's options list
- Rejects non-string values

**Usage:**
```go
// Define a service
type StatusService struct{}
func (s *StatusService) GetOptions(ctx context.Context) ([]jpack.Option, error) {
    return []jpack.Option{
        {UniqueName: "active", DisplayName: "Active"},
        {UniqueName: "inactive", DisplayName: "Inactive"},
        {UniqueName: "pending", DisplayName: "Pending"},
    }, nil
}

// Create options field type
optionsField := jpack.NewOptions(&StatusService{})

// Validate values (uses uniqueName)
err := optionsField.Validate("active")     // nil
err := optionsField.Validate("invalid")    // error (not in options)
err := optionsField.Validate(123)          // error (not a string)

// Get display name for client
displayName, err := optionsField.GetDisplayName(ctx, "active") // "Active"

// Get unique name from display name
uniqueName, err := optionsField.GetUniqueName(ctx, "Active") // "active"
```

### InMemoryOptionService

The `InMemoryOptionService` provides a thread-safe in-memory implementation of the `OptionService` interface.

```go
type InMemoryOptionService struct {
    options []Option
    mu      sync.RWMutex
}
```

#### Functions

#### NewInMemoryOptionService

```go
func NewInMemoryOptionService(options []Option) *InMemoryOptionService
```

Creates a new in-memory option service with the given options.

**Parameters:**
- `options` - Initial list of options (can be nil for empty service)

**Returns:**
- `*InMemoryOptionService` - A new in-memory option service instance

#### Methods

- **`GetOptions(ctx context.Context) ([]Option, error)`** - Returns all options (implements OptionService)
- **`AddOption(option Option)`** - Adds a new option (ignores duplicates)
- **`RemoveOption(uniqueName string) bool`** - Removes an option by uniqueName
- **`UpdateOption(option Option) bool`** - Updates an existing option
- **`GetOptionByUniqueName(uniqueName string) (Option, bool)`** - Gets option by uniqueName
- **`GetOptionByDisplayName(displayName string) (Option, bool)`** - Gets option by displayName
- **`Clear()`** - Removes all options
- **`Count() int`** - Returns the number of options
- **`HasOption(uniqueName string) bool`** - Checks if option exists by uniqueName
- **`HasDisplayName(displayName string) bool`** - Checks if option exists by displayName

**Features:**
- Thread-safe with read-write mutex
- Context cancellation support
- Duplicate prevention (by uniqueName)
- Defensive copying to prevent external modification
- Full CRUD operations

**Usage:**
```go
// Create with initial options
options := []jpack.Option{
    {UniqueName: "active", DisplayName: "Active"},
    {UniqueName: "inactive", DisplayName: "Inactive"},
}
service := jpack.NewInMemoryOptionService(options)

// Use with Options FieldType
optionsField := jpack.NewOptions(service)

// Dynamic operations
service.AddOption(jpack.Option{UniqueName: "pending", DisplayName: "Pending"})
service.RemoveOption("inactive")
service.UpdateOption(jpack.Option{UniqueName: "active", DisplayName: "Active Updated"})

// Query operations
count := service.Count()
hasOption := service.HasOption("active")
option, found := service.GetOptionByUniqueName("active")
```

## Record Operations

### MongoRecord

The `mongoRecord` struct implements `JRecord` for MongoDB operations.

```go
type mongoRecord struct {
    // fields are private
}
```

#### Functions

#### NewMongoRecord

```go
func NewMongoRecord(schema JSchema) *mongoRecord
```

Creates a new MongoDB record instance.

**Parameters:**
- `schema` - The schema for this record

**Returns:**
- `*mongoRecord` - A new record instance

**Example:**
```go
userSchema := jpack.NewSchema("users").
    Field("name", &jpack.String{}).
    Build()

record := jpack.NewMongoRecord(userSchema)
```

### Field Implementation

The `Field` struct implements the `JField` interface.

```go
type Field struct {
    // fields are private
}
```

This struct is used internally by the schema builder and is not typically instantiated directly.

### Edge Implementation

The `Edge` struct implements the `JEdge` interface.

```go
type Edge struct {
    // fields are private
}
```

This struct is used internally by the schema builder and is not typically instantiated directly.

## MongoDB Integration

### Context Keys

#### Conn

```go
var Conn key = "jpack.conn.mongo"
```

The context key used to store the MongoDB database connection.

**Usage:**
```go
ctx := context.WithValue(context.Background(), jpack.Conn, mongoDatabase)
```

### Constants

#### defaultMongoPK

```go
const defaultMongoPK = "_id"
```

The default MongoDB primary key field name.

### Functions

#### MustConn

```go
func MustConn(ctx context.Context) *mongo.Database
```

Retrieves the MongoDB database connection from the context. Panics if not found.

**Parameters:**
- `ctx` - Context containing the database connection

**Returns:**
- `*mongo.Database` - The MongoDB database instance

**Panics:**
- If the connection is not found in the context
- If the connection is nil

**Example:**
```go
ctx := context.WithValue(context.Background(), jpack.Conn, db)
database := jpack.MustConn(ctx)
```

## Future Interfaces

The following interfaces are defined but not yet implemented:

### Filter

```go
type Filter interface {
    Left() any
    Right() any
    Operator() string
    And() []Filter
    Or() []Filter
}
```

Represents query filters for database operations.

### SelectField

```go
type SelectField interface {
    // Methods to be defined
}
```

Represents fields to select in queries.

### JQuery

```go
type JQuery interface {
    Select() []SelectField
    Filter() Filter
}
```

Represents database queries.

### JRepository

```go
type JRepository interface {
    Create(context.Context, []JRecord) error
    Update(context.Context, []JRecord) error
    First(context.Context, JQuery) (mo.Option[JRecord], error)
    FindAll(context.Context, JQuery) ([]JRecord, error)
    Delete(context.Context, JRecord) error
}
```

Represents a repository for database operations.

## Error Handling

### Common Errors

JPack operations can return various types of errors:

#### Validation Errors

Returned when field validation fails:
```go
err := numberField.Validate("not-a-number")
// Returns: "value is not a valid integer"
```

#### Schema Errors

Returned when schema operations fail:
```go
field, exists := schema.Field("non_existent")
// exists will be false
```

#### Database Errors

Returned when database operations fail:
```go
err := record.Save(ctx)
// May return MongoDB-specific errors
```

### Best Practices

1. **Always check for field existence:**
   ```go
   field, exists := schema.Field("fieldname")
   if !exists {
       return errors.New("field not found")
   }
   ```

2. **Validate before saving:**
   ```go
   if err := record.Validate(); err != nil {
       return fmt.Errorf("validation failed: %w", err)
   }
   ```

3. **Handle MongoDB connection errors:**
   ```go
   if err := record.Save(ctx); err != nil {
       if mongo.IsNetworkError(err) {
           // Handle network errors
       }
       return err
   }
   ```

## Type Conversions

### Automatic Conversions

JPack automatically handles certain type conversions:

#### String to Number

```go
numberField := &jpack.Number{}
err := numberField.Validate("123") // nil - automatically converts
```

#### MongoDB ObjectID to String

```go
// MongoDB ObjectID is automatically converted to hex string
// when accessed through record.Value()
```

### Custom Type Conversions

You can implement custom type conversions by creating custom field types:

```go
type CustomType struct{}

func (c *CustomType) Validate(value any) error {
    // Custom validation logic
    return nil
}

func (c *CustomType) Scan(ctx context.Context, field JField, row map[string]any) (any, error) {
    // Custom scanning logic
    return nil, nil
}

func (c *CustomType) SetValue(ctx context.Context, field JField, value any, row map[string]any) error {
    // Custom value setting logic
    return nil
}
```

## Performance Considerations

### Field Access

Field access is O(n) where n is the number of fields in the schema. For better performance with many fields, consider:

1. Caching field references
2. Using field maps for frequent lookups

### Memory Usage

- `mongoRecord` maintains both original and modified field values
- Consider the memory impact for large records or many concurrent records

### Database Operations

- Use batch operations for multiple records
- Consider using MongoDB transactions for multi-document operations
- Implement proper indexing on frequently queried fields

## Version Compatibility

JPack is built with:
- Go 1.24.2+
- MongoDB Driver v2+
- Support for modern Go features like generics (where applicable)

## Thread Safety

JPack components have the following thread safety characteristics:

- **Schemas**: Thread-safe for read operations after building
- **Records**: NOT thread-safe - use mutex protection for concurrent access
- **Field Types**: Thread-safe for validation and type operations
- **Schema Builder**: NOT thread-safe - should be used from single goroutine

## Debugging

### Logging

JPack uses zerolog for internal logging. To enable debug logging:

```go
import "github.com/rs/zerolog/log"

log.Logger = log.With().Caller().Logger()
```

### Common Issues

1. **Field not found errors**: Check field names for typos
2. **Validation errors**: Verify data types match field type expectations
3. **MongoDB connection errors**: Ensure context contains valid connection
4. **Type conversion errors**: Check that values can be converted to expected types

### Troubleshooting

1. **Enable logging** to see internal operations
2. **Check field validation** before setting values
3. **Verify MongoDB connection** in context
4. **Use proper error handling** to catch and diagnose issues
