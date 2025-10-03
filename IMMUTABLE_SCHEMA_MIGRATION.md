# Immutable Schema Migration Guide

## Overview

The schema system has been redesigned to use an immutable-by-default approach with a builder pattern. This eliminates panic-based immutability and provides a cleaner, more functional API.

## Key Changes

### 1. All Schemas are Immutable

**Before:**
```go
schema := NewJSchema("User")
schema.AddField("name", JString, nil)  // Modifies schema in place
schema.Freeze() // Makes it immutable, panics on further modification
```

**After:**
```go
// Using SchemaBuilder (recommended)
schema := NewSchemaBuilder("User").
    AddField("name", JString, nil).
    Build()  // Returns immutable schema

// Updating an immutable schema
updatedSchema := schema.Update(func(sb *SchemaBuilder) {
    sb.AddField("email", JString, nil)
    sb.AddField("age", JInt, nil)
})
```

### 2. All Fields are Immutable

**Before:**
```go
field := NewJField("name", JString, nil)
field.SetRequired(true)  // Modifies field in place
field.SetUnique(true)
```

**After:**
```go
// Using FieldBuilder (recommended)
field := NewFieldBuilder("name", JString, nil).
    Required().
    Unique().
    Build()

// Updating a field
updatedField := field.Update(func(fb *FieldBuilder) {
    fb.Required()
    fb.WithValidation(myValidationFunc)
})
```

### 3. Schema Update Pattern

The `Update` method provides a callback-based approach for modifying immutable schemas:

```go
// Get original schema
userSchema := NewSchemaBuilder("User").
    AddField("id", JString, nil).
    AddField("name", JString, nil).
    Build()

// Create updated version
updatedSchema := userSchema.Update(func(sb *SchemaBuilder) {
    // Add new fields
    sb.AddField("email", JString, nil)
    sb.AddField("age", JInt, nil)
    
    // Add validation
    sb.AddValidation(func(ctx context.Context, rec JRecord) error {
        // Custom validation logic
        return nil
    })
    
    // Add references
    sb.AddRef("profile", profileSchema)
})

// Original schema is unchanged
fmt.Println(len(userSchema.Fields()))  // 2
fmt.Println(len(updatedSchema.Fields()))  // 4
```

### 4. Field Update Pattern

```go
// Create original field
nameField := NewFieldBuilder("name", JString, nil).
    Required().
    Build()

// Create updated version
uniqueNameField := nameField.Update(func(fb *FieldBuilder) {
    fb.Unique()
})

// Original field is unchanged
fmt.Println(nameField.IsUnique())      // false
fmt.Println(uniqueNameField.IsUnique()) // true
```

## Backward Compatibility

The old API is still supported but deprecated:

```go
// Old mutable API (deprecated but still works)
schema := NewJSchema("User")
schema.AddField("name", JString, nil)
schema.AddField("email", JString, nil)

// This returns a mutable adapter that logs warnings
// when mutation methods are called
```

**Note:** The old API uses `mutableSchemaAdapter` internally, which provides backward compatibility but logs warnings when mutation methods are called.

## Benefits

1. **No Panics**: Eliminates panic-based immutability errors
2. **Thread-Safe**: Immutable schemas can be safely shared across goroutines
3. **Functional Style**: Update pattern encourages functional programming style
4. **Clear Intent**: Builder pattern makes schema construction explicit
5. **Versioning**: Easy to create modified versions of schemas without affecting originals

## Migration Steps

### Step 1: Update Schema Creation

**Before:**
```go
userSchema := NewJSchema("User")
userSchema.AddField("id", JString, nil)
userSchema.AddField("name", JString, nil)
userSchema.Freeze()
```

**After:**
```go
userSchema := NewSchemaBuilder("User").
    AddField("id", JString, nil).
    AddField("name", JString, nil).
    Build()
```

### Step 2: Update Schema Modifications

**Before:**
```go
// Clone and modify
newSchema := schema.Clone()
newSchema.AddField("email", JString, nil)
newSchema.Freeze()
```

**After:**
```go
newSchema := schema.Update(func(sb *SchemaBuilder) {
    sb.AddField("email", JString, nil)
})
```

### Step 3: Update Field Creation

**Before:**
```go
field := NewJField("name", JString, nil)
field.SetRequired(true)
field.SetUnique(true)
field.Freeze()
```

**After:**
```go
field := NewFieldBuilder("name", JString, nil).
    Required().
    Unique().
    Build()
```

### Step 4: Update Tests

**Before:**
```go
func TestSchema(t *testing.T) {
    schema := NewJSchema("Test")
    assert.False(t, schema.IsImmutable())
    
    schema.AddField("name", JString, nil)
    schema.Freeze()
    
    assert.True(t, schema.IsImmutable())
    assert.Panics(t, func() {
        schema.AddField("email", JString, nil)
    })
}
```

**After:**
```go
func TestSchema(t *testing.T) {
    schema := NewSchemaBuilder("Test").
        AddField("name", JString, nil).
        Build()
    
    assert.True(t, schema.IsImmutable())  // Always true now
    
    // No panics, just creates new schema
    updatedSchema := schema.Update(func(sb *SchemaBuilder) {
        sb.AddField("email", JString, nil)
    })
    
    assert.Equal(t, 1, len(schema.Fields()))
    assert.Equal(t, 2, len(updatedSchema.Fields()))
}
```

## API Reference

### SchemaBuilder

```go
type SchemaBuilder struct { /* ... */ }

func NewSchemaBuilder(name string) *SchemaBuilder
func (sb *SchemaBuilder) AddField(name string, fieldType JFieldType, defaultValue any) *SchemaBuilder
func (sb *SchemaBuilder) AddRequiredField(name string, fieldType JFieldType, defaultValue any) *SchemaBuilder
func (sb *SchemaBuilder) AddUniqueField(name string, fieldType JFieldType, defaultValue any) *SchemaBuilder
func (sb *SchemaBuilder) AddRequiredUniqueField(name string, fieldType JFieldType, defaultValue any) *SchemaBuilder
func (sb *SchemaBuilder) AddFieldWithValidation(name string, fieldType JFieldType, defaultValue any, validation ValidationFunc) *SchemaBuilder
func (sb *SchemaBuilder) AddRef(name string, targetSchema JSchema) *SchemaBuilder
func (sb *SchemaBuilder) AddArrayRef(name string, targetSchema JSchema) *SchemaBuilder
func (sb *SchemaBuilder) AddEdge(edge JEdge) *SchemaBuilder
func (sb *SchemaBuilder) AddValidation(fn ValidationFunc) *SchemaBuilder
func (sb *SchemaBuilder) SetIDField(fieldName string) *SchemaBuilder
func (sb *SchemaBuilder) Build() JSchema
```

### FieldBuilder

```go
type FieldBuilder struct { /* ... */ }

func NewFieldBuilder(name string, fieldType JFieldType, defaultValue any) *FieldBuilder
func (fb *FieldBuilder) SetRequired(required bool) *FieldBuilder
func (fb *FieldBuilder) Required() *FieldBuilder
func (fb *FieldBuilder) SetUnique(unique bool) *FieldBuilder
func (fb *FieldBuilder) Unique() *FieldBuilder
func (fb *FieldBuilder) WithValidation(fn ValidationFunc) *FieldBuilder
func (fb *FieldBuilder) Build() JField
```

### JSchema Interface (Immutable)

```go
type JSchema interface {
    // Read-only accessors
    Name() string
    Fields() []JField
    Field(name string) (JField, bool)
    Refs() []JRef
    Ref(name string) (JRef, bool)
    Edges() []JEdge
    Validations() []ValidationFunc
    GetIDField() JField
    String() string
    
    // Validation
    Validate(ctx context.Context, rec JRecord) error
    
    // Update creates a new schema with modifications
    Update(fn func(*SchemaBuilder)) JSchema
    
    // Backward compatibility (deprecated)
    Clone() JSchema
    IsImmutable() bool  // Always returns true
    Freeze() JSchema    // Returns self
    AddField(name string, fType JFieldType, defaultValue any) JField  // Logs warning
    AddRef(name string, schema JSchema) JRef                           // Logs warning
    AddEdge(edge JEdge) JSchema                                        // Logs warning
    AddValidation(fn ValidationFunc)                                   // Logs warning
    SetIDField(field JField) JSchema                                   // Logs warning
}
```

### JField Interface (Immutable)

```go
type JField interface {
    // Read-only accessors
    Name() string
    Type() JFieldType
    DefaultValue() any
    IsRequired() bool
    IsUnique() bool
    Validation() ValidationFunc
    String() string
    
    // Update creates a new field with modifications
    Update(fn func(*FieldBuilder)) JField
    
    // Backward compatibility (returns new instances)
    SetRequired(required bool) JField
    SetUnique(unique bool) JField
    SetValidation(fn ValidationFunc) JField
    IsImmutable() bool  // Always returns true
    Freeze() JField     // Returns self
}
```

## Testing

Run tests with:
```bash
go test ./schema -v
```

Note: Some tests may need to be updated to work with the new immutable design. Tests that expect panics or mutable behavior will need to be refactored.

## Examples

See the `/examples` directory for updated examples using the new builder pattern.

## Questions?

For questions or issues, please refer to the main project documentation or open an issue.

