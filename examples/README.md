# JPack Examples

This directory contains examples demonstrating how to use JPack features.

## Schema Validation Example

The `validation_example.go` file demonstrates how to use the schema validation functionality in JPack.

### Features Demonstrated

1. **Valid Record Validation**: Shows how to validate a record with all required fields and valid values
2. **Missing Required Fields**: Demonstrates validation failure when required fields are missing
3. **Invalid Field Values**: Shows how validation works when invalid values are provided
4. **Default Values**: Demonstrates how fields with default values are handled during validation
5. **Schema Mismatch**: Shows validation failure when a record's schema doesn't match the validation schema
6. **Custom Validation**: Demonstrates how to add custom validation functions to schemas

### Running the Example

```bash
go run validation_example.go
```

### Expected Output

```
=== Schema Validation Example ===

1. Valid record:
   ✅ Validation passed

2. Missing required field:
   ❌ Validation failed: required field 'age' is missing

3. Invalid field value:
   ❌ SetValue failed: value is not a valid integer
   ❌ Validation failed: required field 'age' is missing

4. Record with default values:
   ✅ Validation passed

5. Wrong schema:
   ❌ Validation failed: record schema does not match validation schema

6. Custom validation example:
   ✅ Custom validation passed
   ❌ Custom validation failed: name must be at least 3 characters long

=== End of All Examples ===
```

### Key Concepts

- **Schema Validation**: The `Validate(JRecord) error` method validates a record against its schema
- **Field Type Validation**: Each field type has its own validation logic (String, Number, Boolean, DateTime, etc.)
- **Required Fields**: Fields without default values are considered required
- **Default Values**: Fields with default values use the default when the field is missing from the record
- **Schema Matching**: Records can only be validated against their own schema
- **Custom Validation**: You can add custom validation functions using the `Validation()` method

## Validation Functions

JPack provides a set of common validation functions in the `validations` package that you can use to add validation rules to your schemas.

### Available Validation Functions

#### RequiredField(fieldName string)
Ensures a field is not null or empty string.

```go
schema := jpack.NewSchema("users").
    Field("name", &jpack.String{}).
    Validation(validations.RequiredField("name")).
    Build()
```

#### ReadOnlyField(fieldName string)
Prevents editing of a field after creation (allows setting during creation).

```go
schema := jpack.NewSchema("users").
    Field("created_at", &jpack.String{}).
    Validation(validations.ReadOnlyField("created_at")).
    Build()
```

#### StringLength(fieldName string, minLength, maxLength int)
Validates string length with minimum and maximum constraints.

```go
schema := jpack.NewSchema("users").
    Field("username", &jpack.String{}).
    Validation(validations.StringLength("username", 3, 20)).
    Build()
```

#### NumberRange(fieldName string, min, max float64)
Validates numeric ranges with minimum and maximum constraints.

```go
schema := jpack.NewSchema("users").
    Field("age", &jpack.Number{}).
    Validation(validations.NumberRange("age", 18, 65)).
    Build()
```

#### EmailFormat(fieldName string)
Validates email format (basic validation).

```go
schema := jpack.NewSchema("users").
    Field("email", &jpack.String{}).
    Validation(validations.EmailFormat("email")).
    Build()
```

#### PositiveNumber(fieldName string)
Ensures a number is positive (greater than 0).

```go
schema := jpack.NewSchema("products").
    Field("price", &jpack.Number{}).
    Validation(validations.PositiveNumber("price")).
    Build()
```

#### NonNegativeNumber(fieldName string)
Ensures a number is non-negative (greater than or equal to 0).

```go
schema := jpack.NewSchema("products").
    Field("quantity", &jpack.Number{}).
    Validation(validations.NonNegativeNumber("quantity")).
    Build()
```

#### StringNotEmpty(fieldName string)
Ensures a string is not empty (but allows nil).

```go
schema := jpack.NewSchema("products").
    Field("sku", &jpack.String{}).
    Validation(validations.StringNotEmpty("sku")).
    Build()
```

### Using Multiple Validations

You can combine multiple validation functions on a single schema:

```go
userSchema := jpack.NewSchema("users").
    Field("username", &jpack.String{}).
    Field("email", &jpack.String{}).
    Field("age", &jpack.Number{}).
    Validation(validations.RequiredField("username")).
    Validation(validations.RequiredField("email")).
    Validation(validations.StringLength("username", 3, 20)).
    Validation(validations.EmailFormat("email")).
    Validation(validations.NumberRange("age", 13, 120)).
    Build()
```

### Custom Validation Functions

You can also create your own validation functions:

```go
schema := jpack.NewSchema("users").
    Field("password", &jpack.String{}).
    Validation(func(record jpack.JRecord) error {
        passwordField, _ := record.Schema().Field("password")
        if password, exists := record.Value(passwordField); exists {
            if passwordStr, ok := password.(string); ok {
                if len(passwordStr) < 8 {
                    return fmt.Errorf("password must be at least 8 characters long")
                }
            }
        }
        return nil
    }).
    Build()
``` 