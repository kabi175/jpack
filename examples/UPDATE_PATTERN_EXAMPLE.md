# Schema Update Pattern Examples

## Basic Update Pattern

```go
// Create initial schema
userSchema := NewSchemaBuilder("User").
    AddField("id", JString, nil).
    AddField("name", JString, nil).
    Build()

// Update schema - creates new instance
updatedSchema := userSchema.Update(func(sb *SchemaBuilder) {
    sb.AddField("email", JString, nil)
    sb.AddField("phone", JString, nil)
})

// Original schema unchanged
fmt.Println(len(userSchema.Fields()))         // 2
fmt.Println(len(updatedSchema.Fields()))      // 4
```

## Field Update Pattern

```go
// Create field
nameField := NewFieldBuilder("name", JString, "").Build()

// Update field - creates new instance
requiredNameField := nameField.Update(func(fb *FieldBuilder) {
    fb.Required()
    fb.Unique()
})

// Original field unchanged
fmt.Println(nameField.IsRequired())           // false
fmt.Println(requiredNameField.IsRequired())   // true
```

## Schema Versioning

```go
// Version 1: Basic user
v1Schema := NewSchemaBuilder("User").
    AddField("id", JString, nil).
    AddField("name", JString, nil).
    Build()

// Version 2: Add email
v2Schema := v1Schema.Update(func(sb *SchemaBuilder) {
    sb.AddField("email", JString, nil)
})

// Version 3: Add profile fields
v3Schema := v2Schema.Update(func(sb *SchemaBuilder) {
    sb.AddField("phone", JString, nil)
    sb.AddField("address", JString, nil)
    sb.AddField("age", JInt, 0)
})

// All versions are independent
registry.Register(v1Schema)  // User v1
registry.Register(v2Schema)  // User v2
registry.Register(v3Schema)  // User v3
```

## Adding Validation via Update

```go
baseSchema := NewSchemaBuilder("Product").
    AddField("id", JString, nil).
    AddField("price", JFloat64, 0.0).
    Build()

// Add validation
validatedSchema := baseSchema.Update(func(sb *SchemaBuilder) {
    sb.AddValidation(func(ctx context.Context, rec JRecord) error {
        price, ok := rec.Get("price").(float64)
        if !ok || price <= 0 {
            return fmt.Errorf("price must be positive")
        }
        return nil
    })
})
```

## Adding References via Update

```go
profileSchema := NewSchemaBuilder("Profile").
    AddField("bio", JString, nil).
    Build()

// Start with basic user schema
userSchema := NewSchemaBuilder("User").
    AddField("id", JString, nil).
    AddField("name", JString, nil).
    Build()

// Add profile reference
userWithProfileSchema := userSchema.Update(func(sb *SchemaBuilder) {
    sb.AddRef("profile", profileSchema)
})
```

## Adding Edges via Update

```go
userSchema := NewSchemaBuilder("User").
    AddField("id", JString, nil).
    Build()

orderSchema := NewSchemaBuilder("Order").
    AddField("id", JString, nil).
    Build()

// Add edge relationship
userWithOrdersSchema := userSchema.Update(func(sb *SchemaBuilder) {
    sb.AddEdge(NewJEdge(
        "user_orders",
        userSchema,
        orderSchema,
        EdgeOneToMany,
    ))
})
```

## Chaining Multiple Updates

```go
schema := NewSchemaBuilder("User").
    AddField("id", JString, nil).
    Build()

// Chain updates for progressive enhancement
finalSchema := schema.
    Update(func(sb *SchemaBuilder) {
        sb.AddField("name", JString, nil)
        sb.AddField("email", JString, nil)
    }).
    Update(func(sb *SchemaBuilder) {
        sb.AddValidation(emailValidation)
    }).
    Update(func(sb *SchemaBuilder) {
        sb.AddRef("profile", profileSchema)
    })
```

## Conditional Updates

```go
func addFieldsBasedOnFeatureFlags(schema JSchema, flags FeatureFlags) JSchema {
    return schema.Update(func(sb *SchemaBuilder) {
        if flags.EmailEnabled {
            sb.AddField("email", JString, nil)
        }
        
        if flags.PhoneEnabled {
            sb.AddField("phone", JString, nil)
        }
        
        if flags.AddressEnabled {
            sb.AddField("address", JString, nil)
            sb.AddField("city", JString, nil)
            sb.AddField("country", JString, nil)
        }
    })
}

// Usage
baseSchema := NewSchemaBuilder("User").
    AddField("id", JString, nil).
    AddField("name", JString, nil).
    Build()

// Create different versions based on feature flags
basicUserSchema := addFieldsBasedOnFeatureFlags(baseSchema, BasicFeatures)
premiumUserSchema := addFieldsBasedOnFeatureFlags(baseSchema, PremiumFeatures)
```

## Migration Pattern

```go
// Old schema (v1)
v1Schema := registry.Get("User")

// Migrate to v2 by adding new fields
v2Schema := v1Schema.Update(func(sb *SchemaBuilder) {
    // Add new required fields with defaults
    sb.AddField("created_at", JTime, time.Now())
    sb.AddField("updated_at", JTime, time.Now())
    sb.AddField("version", JInt, 2)
})

// Replace in registry
registry.Replace(v2Schema)

// Migrate records
for _, record := range oldRecords {
    newRecord := record.Clone().
        Set("created_at", time.Now()).
        Set("updated_at", time.Now()).
        Set("version", 2)
    
    if err := v2Schema.Validate(ctx, newRecord); err != nil {
        log.Error("Migration failed", err)
    }
}
```

## Benefits of Update Pattern

### 1. No Side Effects
```go
original := NewSchemaBuilder("User").AddField("id", JString, nil).Build()
modified := original.Update(func(sb *SchemaBuilder) {
    sb.AddField("name", JString, nil)
})

// original is unchanged - no side effects!
```

### 2. Thread-Safe
```go
var schema JSchema = /* ... */

// Safe to use in multiple goroutines
go func() {
    v1 := schema.Update(func(sb *SchemaBuilder) {
        sb.AddField("field1", JString, nil)
    })
}()

go func() {
    v2 := schema.Update(func(sb *SchemaBuilder) {
        sb.AddField("field2", JString, nil)
    })
}()
```

### 3. Composable
```go
func addAuditFields(schema JSchema) JSchema {
    return schema.Update(func(sb *SchemaBuilder) {
        sb.AddField("created_at", JTime, nil)
        sb.AddField("updated_at", JTime, nil)
        sb.AddField("created_by", JString, nil)
    })
}

func addSoftDelete(schema JSchema) JSchema {
    return schema.Update(func(sb *SchemaBuilder) {
        sb.AddField("deleted_at", JTime, nil)
        sb.AddField("deleted_by", JString, nil)
    })
}

// Compose transformations
userSchema := NewSchemaBuilder("User").
    AddField("id", JString, nil).
    AddField("name", JString, nil).
    Build()

auditedSchema := addAuditFields(userSchema)
softDeleteSchema := addSoftDelete(auditedSchema)
```

### 4. Testable
```go
func TestSchemaUpdate(t *testing.T) {
    original := NewSchemaBuilder("Test").
        AddField("id", JString, nil).
        Build()
    
    updated := original.Update(func(sb *SchemaBuilder) {
        sb.AddField("name", JString, nil)
    })
    
    // Easy to verify original unchanged
    assert.Equal(t, 1, len(original.Fields()))
    assert.Equal(t, 2, len(updated.Fields()))
    
    // Easy to verify new field exists
    field, exists := updated.Field("name")
    assert.True(t, exists)
    assert.Equal(t, "name", field.Name())
}
```

## Comparison with Old Pattern

### Old (Panic-based)
```go
schema := NewJSchema("User")
schema.AddField("id", JString, nil)
schema.Freeze()

// This panics!
schema.AddField("name", JString, nil)  // PANIC!

// Have to clone first
newSchema := schema.Clone()
newSchema.AddField("name", JString, nil)
newSchema.Freeze()
```

### New (Update-based)
```go
schema := NewSchemaBuilder("User").
    AddField("id", JString, nil).
    Build()

// This just works - no panic!
newSchema := schema.Update(func(sb *SchemaBuilder) {
    sb.AddField("name", JString, nil)
})

// Original unchanged, new schema created
```

