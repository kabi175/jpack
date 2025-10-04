// Package validation provides validation functionality for JPack.
// It includes field-level, record-level, and schema-level validations.
//
// The validation system provides a comprehensive set of built-in validators
// and allows for custom validation functions to be registered.
//
// Example:
//
//	// Create validation registry and engine
//	registry := validation.NewValidationRegistry()
//	engine := validation.NewValidationEngine(registry)
//
//	// Register custom validation rule
//	rule := validation.ValidationRule{
//		Name:        "email_domain",
//		Description: "Email must be from allowed domain",
//		Validator: func(ctx context.Context, rec schema.JRecord) error {
//			email, ok := rec.Get("email").(string)
//			if !ok {
//				return fmt.Errorf("email must be a string")
//			}
//			if !strings.HasSuffix(email, "@company.com") {
//				return fmt.Errorf("email must be from company.com domain")
//			}
//			return nil
//		},
//	}
//	err := registry.RegisterRule(rule)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Create schema with built-in validators
//	userSchema := schema.NewSchemaBuilder("User").
//		AddRequiredField("name", schema.JString, nil).
//		AddRequiredField("email", schema.JString, nil).
//		AddField("age", schema.JInt, 18).
//		AddValidation(validation.ValidateMinLength("name", 2)).
//		AddValidation(validation.ValidateEmail("email")).
//		AddValidation(validation.ValidateRange("age", 18, 120)).
//		Build()
//
//	// Validate record against schema
//	user := schema.NewJRecord().
//		Set("name", "John Doe").
//		Set("email", "john@example.com").
//		Set("age", 25)
//
//	err = userSchema.Validate(context.Background(), user)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Cross-field validation
//	orderSchema := schema.NewSchemaBuilder("Order").
//		AddField("total", schema.JFloat64, 0.0).
//		AddField("discount", schema.JFloat64, 0.0).
//		AddValidation(func(ctx context.Context, rec schema.JRecord) error {
//			total, _ := rec.Get("total").(float64)
//			discount, _ := rec.Get("discount").(float64)
//			if discount > total {
//				return fmt.Errorf("discount cannot be greater than total")
//			}
//			return nil
//		}).
//		Build()
//
//	// Use validation engine for advanced validation
//	fieldCtx := validation.ValidationContext{
//		Schema:     userSchema,
//		Record:     user,
//		Field:      emailField,
//		Value:      "user@company.com",
//		Repository: userRepo,
//	}
//	err = engine.ValidateField(context.Background(), fieldCtx)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Validate all (field + record + schema)
//	recordCtx := validation.ValidationContext{
//		Schema:     userSchema,
//		Record:     user,
//		Repository: userRepo,
//	}
//	err = engine.ValidateAll(context.Background(), recordCtx)
//	if err != nil {
//		log.Fatal(err)
//	}
package validation

import (
	"context"
	"fmt"
	"strings"

	"github.com/kabi175/jpack/logger"
	"github.com/kabi175/jpack/schema"
)

// ValidationFunc represents a validation function.
// It validates a record and returns an error if validation fails.
//
// Example:
//
//	emailValidator := func(ctx context.Context, rec schema.JRecord) error {
//		email := rec.Get("email").(string)
//		if !strings.Contains(email, "@") {
//			return fmt.Errorf("invalid email format")
//		}
//		return nil
//	}
//
//	err := emailValidator(ctx, record)
type ValidationFunc func(ctx context.Context, rec schema.JRecord) error

// ValidationError represents a validation error.
// It provides detailed information about validation failures.
//
// Example:
//
//	err := &validation.ValidationError{
//		Field:   "email",
//		Message: "invalid email format",
//		Value:   "invalid-email",
//		Data:    map[string]any{"expected": "valid email format"},
//	}
type ValidationError struct {
	// Field is the field that failed validation
	Field string
	// Message is the error message
	Message string
	// Value is the value that failed validation
	Value any
	// Data contains additional error information
	Data map[string]any
}

func (e *ValidationError) Error() string {
	if e.Field != "" {
		logger.Validation.Debug().
			Str("field", e.Field).
			Str("message", e.Message).
			Msg("validation error occurred")
		return fmt.Sprintf("validation error for field '%s': %s", e.Field, e.Message)
	}

	logger.Validation.Debug().
		Str("message", e.Message).
		Msg("validation error occurred")

	return e.Message
}

// ValidationResult represents the result of validation.
// It contains the validation status and any errors that occurred.
//
// Example:
//
//	result := &validation.ValidationResult{
//		Valid:  false,
//		Errors: []validation.ValidationError{
//			{Field: "email", Message: "invalid email format"},
//			{Field: "age", Message: "age must be at least 18"},
//		},
//	}
//
//	if result.HasErrors() {
//		fmt.Println("Validation failed:", result.Error())
//	}
type ValidationResult struct {
	// Valid indicates whether validation passed
	Valid bool
	// Errors contains all validation errors
	Errors []ValidationError
}

func (r *ValidationResult) AddError(field, message string, value any) {
	r.Errors = append(r.Errors, ValidationError{
		Field:   field,
		Message: message,
		Value:   value,
	})
	r.Valid = false
}

func (r *ValidationResult) HasErrors() bool {
	return len(r.Errors) > 0
}

func (r *ValidationResult) Error() string {
	if len(r.Errors) == 0 {
		return ""
	}

	var messages []string
	for _, err := range r.Errors {
		messages = append(messages, err.Error())
	}
	return strings.Join(messages, "; ")
}

// FieldValidator represents a field-level validator.
// It validates individual fields within a record.
//
// Example:
//
//	type EmailValidator struct{}
//
//	func (v *EmailValidator) Validate(ctx context.Context, field schema.JField, value any) error {
//		email := value.(string)
//		if !strings.Contains(email, "@") {
//			return fmt.Errorf("invalid email format")
//		}
//		return nil
//	}
type FieldValidator interface {
	// Validate validates a field value
	Validate(ctx context.Context, field schema.JField, value any) error
}

// RecordValidator represents a record-level validator.
// It validates entire records against a schema.
//
// Example:
//
//	type UserRecordValidator struct{}
//
//	func (v *UserRecordValidator) Validate(ctx context.Context, schema schema.JSchema, rec schema.JRecord) error {
//		// Validate entire record
//		return nil
//	}
type RecordValidator interface {
	// Validate validates a record against a schema
	Validate(ctx context.Context, schema schema.JSchema, rec schema.JRecord) error
}

// SchemaValidator represents a schema-level validator.
// It validates schema definitions themselves.
//
// Example:
//
//	type SchemaStructureValidator struct{}
//
//	func (v *SchemaStructureValidator) Validate(ctx context.Context, schema schema.JSchema) error {
//		// Validate schema structure
//		return nil
//	}
type SchemaValidator interface {
	// Validate validates a schema definition
	Validate(ctx context.Context, schema schema.JSchema) error
}

// ValidationRule represents a validation rule.
// It defines a named validation with a description and validator function.
//
// Example:
//
//	rule := validation.ValidationRule{
//		Name:        "email_format",
//		Description: "Validates email format",
//		Validator: func(ctx context.Context, rec schema.JRecord) error {
//			email := rec.Get("email").(string)
//			if !strings.Contains(email, "@") {
//				return fmt.Errorf("invalid email format")
//			}
//			return nil
//		},
//	}
//
//	validation.RegisterValidationRule(rule)
type ValidationRule struct {
	// Name is the unique name for the validation rule
	Name string
	// Description describes what the validation rule does
	Description string
	// Validator is the validation function
	Validator ValidationFunc
}

// ValidationRegistry manages validation rules.
// It provides registration, retrieval, and management of validation rules.
//
// Example:
//
//	registry := validation.NewValidationRegistry()
//
//	// Register a rule
//	rule := validation.ValidationRule{
//		Name:        "email_format",
//		Description: "Validates email format",
//		Validator:   emailValidator,
//	}
//	err := registry.RegisterRule(rule)
//
//	// Get a rule
//	retrievedRule, ok := registry.GetRule("email_format")
//	if ok {
//		fmt.Println("Found rule:", retrievedRule.Description)
//	}
//
//	// List all rules
//	ruleNames := registry.ListRules()
//	fmt.Println("Available rules:", ruleNames)
type ValidationRegistry interface {
	// RegisterRule registers a validation rule
	RegisterRule(rule ValidationRule) error
	// GetRule retrieves a validation rule by name
	GetRule(name string) (ValidationRule, bool)
	// UnregisterRule removes a validation rule
	UnregisterRule(name string) error
	// ListRules returns all registered rule names
	ListRules() []string
}

// jValidationRegistry implements ValidationRegistry
type jValidationRegistry struct {
	rules map[string]ValidationRule
}

// NewValidationRegistry creates a new validation registry.
// It initializes a registry with built-in validation rules.
//
// Example:
//
//	registry := validation.NewValidationRegistry()
//
//	// Register custom rule
//	rule := validation.ValidationRule{
//		Name:        "custom_rule",
//		Description: "Custom validation rule",
//		Validator:   customValidator,
//	}
//	registry.RegisterRule(rule)
func NewValidationRegistry() ValidationRegistry {
	registry := &jValidationRegistry{
		rules: make(map[string]ValidationRule),
	}

	// Register built-in validation rules
	registry.RegisterRule(ValidationRule{
		Name:        "required",
		Description: "Field is required",
		Validator:   RequiredValidator(),
	})

	registry.RegisterRule(ValidationRule{
		Name:        "min_length",
		Description: "Minimum length validation",
		Validator:   MinLengthValidator(),
	})

	registry.RegisterRule(ValidationRule{
		Name:        "max_length",
		Description: "Maximum length validation",
		Validator:   MaxLengthValidator(),
	})

	registry.RegisterRule(ValidationRule{
		Name:        "regex",
		Description: "Regular expression validation",
		Validator:   RegexValidator(),
	})

	registry.RegisterRule(ValidationRule{
		Name:        "email",
		Description: "Email format validation",
		Validator:   EmailValidator(),
	})

	registry.RegisterRule(ValidationRule{
		Name:        "min_value",
		Description: "Minimum value validation",
		Validator:   MinValueValidator(),
	})

	registry.RegisterRule(ValidationRule{
		Name:        "max_value",
		Description: "Maximum value validation",
		Validator:   MaxValueValidator(),
	})

	registry.RegisterRule(ValidationRule{
		Name:        "range",
		Description: "Range validation",
		Validator:   RangeValidator(),
	})

	registry.RegisterRule(ValidationRule{
		Name:        "unique",
		Description: "Unique value validation",
		Validator:   UniqueValidator(),
	})

	return registry
}

// Global validation registry
var globalValidationRegistry ValidationRegistry = NewValidationRegistry()

// GetGlobalValidationRegistry returns the global validation registry
func GetGlobalValidationRegistry() ValidationRegistry {
	return globalValidationRegistry
}

func (r *jValidationRegistry) RegisterRule(rule ValidationRule) error {
	if rule.Name == "" {
		logger.Validation.Error().Msg("attempted to register rule with empty name")
		return fmt.Errorf("rule name cannot be empty")
	}

	if rule.Validator == nil {
		logger.Validation.Error().
			Str("rule", rule.Name).
			Msg("attempted to register rule with nil validator")
		return fmt.Errorf("rule validator cannot be nil")
	}

	logger.Validation.Debug().
		Str("rule", rule.Name).
		Str("description", rule.Description).
		Msg("registering validation rule")

	r.rules[rule.Name] = rule

	logger.Validation.Debug().
		Str("rule", rule.Name).
		Int("total_rules", len(r.rules)).
		Msg("validation rule registered successfully")

	return nil
}

func (r *jValidationRegistry) GetRule(name string) (ValidationRule, bool) {
	rule, exists := r.rules[name]
	if !exists {
		logger.Validation.Debug().
			Str("rule", name).
			Msg("validation rule not found")
	} else {
		logger.Validation.Debug().
			Str("rule", name).
			Msg("validation rule found")
	}
	return rule, exists
}

func (r *jValidationRegistry) UnregisterRule(name string) error {
	if _, exists := r.rules[name]; !exists {
		logger.Validation.Warn().
			Str("rule", name).
			Msg("attempted to unregister non-existent rule")
		return fmt.Errorf("rule '%s' is not registered", name)
	}

	logger.Validation.Info().
		Str("rule", name).
		Msg("unregistering validation rule")

	delete(r.rules, name)

	logger.Validation.Debug().
		Str("rule", name).
		Int("remaining_rules", len(r.rules)).
		Msg("validation rule unregistered successfully")

	return nil
}

func (r *jValidationRegistry) ListRules() []string {
	names := make([]string, 0, len(r.rules))
	for name := range r.rules {
		names = append(names, name)
	}

	logger.Validation.Debug().
		Int("count", len(names)).
		Strs("rules", names).
		Msg("listing validation rules")

	return names
}

// RegisterValidationRule is a convenience function to register a validation rule
func RegisterValidationRule(rule ValidationRule) error {
	return globalValidationRegistry.RegisterRule(rule)
}

// GetValidationRule is a convenience function to get a validation rule
func GetValidationRule(name string) (ValidationRule, bool) {
	return globalValidationRegistry.GetRule(name)
}

// ValidationContext provides context for validation.
// It contains all the information needed to perform validation.
//
// Example:
//
//	ctx := validation.NewValidationContext(userSchema, userRecord, userRepo)
//
//	// Validate field
//	fieldCtx := validation.ValidationContext{
//		Schema:     userSchema,
//		Record:     userRecord,
//		Field:      emailField,
//		Value:      "user@example.com",
//		Repository: userRepo,
//	}
//
//	err := validationEngine.ValidateField(context.Background(), fieldCtx)
type ValidationContext struct {
	// Schema is the schema being validated against
	Schema schema.JSchema
	// Record is the record being validated
	Record schema.JRecord
	// Field is the field being validated (for field-level validation)
	Field schema.JField
	// Value is the value being validated (for field-level validation)
	Value any
	// Repository is the repository for unique checks and other database operations
	Repository interface{}
}

// ValidationEngine orchestrates validation.
// It provides a centralized way to perform different types of validation.
//
// Example:
//
//	engine := validation.NewValidationEngine(registry)
//
//	// Validate field
//	fieldCtx := validation.ValidationContext{
//		Schema:     userSchema,
//		Record:     userRecord,
//		Field:      emailField,
//		Value:      "user@example.com",
//		Repository: userRepo,
//	}
//	err := engine.ValidateField(context.Background(), fieldCtx)
//
//	// Validate record
//	recordCtx := validation.ValidationContext{
//		Schema:     userSchema,
//		Record:     userRecord,
//		Repository: userRepo,
//	}
//	err = engine.ValidateRecord(context.Background(), recordCtx)
//
//	// Validate all
//	err = engine.ValidateAll(context.Background(), recordCtx)
type ValidationEngine interface {
	// ValidateField validates a single field
	ValidateField(ctx context.Context, validationCtx ValidationContext) error
	// ValidateRecord validates an entire record
	ValidateRecord(ctx context.Context, validationCtx ValidationContext) error
	// ValidateSchema validates schema-level rules
	ValidateSchema(ctx context.Context, validationCtx ValidationContext) error
	// ValidateAll performs all types of validation
	ValidateAll(ctx context.Context, validationCtx ValidationContext) error
}

// jValidationEngine implements ValidationEngine
type jValidationEngine struct {
	registry ValidationRegistry
}

// NewValidationEngine creates a new validation engine.
// It initializes an engine with the provided validation registry.
//
// Example:
//
//	registry := validation.NewValidationRegistry()
//	engine := validation.NewValidationEngine(registry)
//
//	// Use engine for validation
//	ctx := validation.NewValidationContext(userSchema, userRecord, userRepo)
//	err := engine.ValidateAll(context.Background(), ctx)
func NewValidationEngine(registry ValidationRegistry) ValidationEngine {
	return &jValidationEngine{
		registry: registry,
	}
}

// GetGlobalValidationEngine returns the global validation engine.
// This provides access to the default validation engine with built-in rules.
//
// Example:
//
//	engine := validation.GetGlobalValidationEngine()
//
//	ctx := validation.NewValidationContext(userSchema, userRecord, userRepo)
//	err := engine.ValidateAll(context.Background(), ctx)
func GetGlobalValidationEngine() ValidationEngine {
	return NewValidationEngine(globalValidationRegistry)
}

func (e *jValidationEngine) ValidateField(ctx context.Context, validationCtx ValidationContext) error {
	if validationCtx.Field == nil {
		logger.Validation.Error().Msg("field cannot be nil in validation context")
		return fmt.Errorf("field cannot be nil")
	}

	logger.Validation.Debug().
		Str("field", validationCtx.Field.Name()).
		Interface("value", validationCtx.Value).
		Msg("validating field")

	// Field-level validations
	if validationCtx.Field.Validation() != nil {
		if err := validationCtx.Field.Validation()(ctx, validationCtx.Record); err != nil {
			return err
		}
	}

	// Required validation
	if validationCtx.Field.IsRequired() && validationCtx.Value == nil {
		logger.Validation.Debug().
			Str("field", validationCtx.Field.Name()).
			Msg("required field validation failed")
		return &ValidationError{
			Field:   validationCtx.Field.Name(),
			Message: "field is required",
			Value:   validationCtx.Value,
		}
	}

	logger.Validation.Debug().
		Str("field", validationCtx.Field.Name()).
		Msg("field validation passed")

	return nil
}

func (e *jValidationEngine) ValidateRecord(ctx context.Context, validationCtx ValidationContext) error {
	if validationCtx.Schema == nil || validationCtx.Record == nil {
		logger.Validation.Error().Msg("schema and record cannot be nil in validation context")
		return fmt.Errorf("schema and record cannot be nil")
	}

	logger.Validation.Debug().
		Str("schema", validationCtx.Schema.Name()).
		Int("fields_count", len(validationCtx.Schema.Fields())).
		Msg("validating record")

	// Validate all fields
	for _, field := range validationCtx.Schema.Fields() {
		value := validationCtx.Record.Get(field.Name())
		fieldCtx := ValidationContext{
			Schema:     validationCtx.Schema,
			Record:     validationCtx.Record,
			Field:      field,
			Value:      value,
			Repository: validationCtx.Repository,
		}

		if err := e.ValidateField(ctx, fieldCtx); err != nil {
			logger.Validation.Debug().
				Str("field", field.Name()).
				Err(err).
				Msg("field validation failed")
			return err
		}
	}

	logger.Validation.Debug().
		Str("schema", validationCtx.Schema.Name()).
		Msg("record validation passed")

	return nil
}

func (e *jValidationEngine) ValidateSchema(ctx context.Context, validationCtx ValidationContext) error {
	if validationCtx.Schema == nil {
		logger.Validation.Error().Msg("schema cannot be nil in validation context")
		return fmt.Errorf("schema cannot be nil")
	}

	logger.Validation.Debug().
		Str("schema", validationCtx.Schema.Name()).
		Int("validations_count", len(validationCtx.Schema.Validations())).
		Msg("validating schema")

	// Schema-level validations
	for _, validation := range validationCtx.Schema.Validations() {
		if err := validation(ctx, validationCtx.Record); err != nil {
			logger.Validation.Debug().
				Str("schema", validationCtx.Schema.Name()).
				Err(err).
				Msg("schema validation failed")
			return err
		}
	}

	logger.Validation.Debug().
		Str("schema", validationCtx.Schema.Name()).
		Msg("schema validation passed")

	return nil
}

func (e *jValidationEngine) ValidateAll(ctx context.Context, validationCtx ValidationContext) error {
	logger.Validation.Debug().
		Str("schema", validationCtx.Schema.Name()).
		Msg("starting full validation")

	// Validate record first
	if err := e.ValidateRecord(ctx, validationCtx); err != nil {
		logger.Validation.Debug().
			Str("schema", validationCtx.Schema.Name()).
			Err(err).
			Msg("record validation failed during full validation")
		return err
	}

	// Validate schema
	if err := e.ValidateSchema(ctx, validationCtx); err != nil {
		logger.Validation.Debug().
			Str("schema", validationCtx.Schema.Name()).
			Err(err).
			Msg("schema validation failed during full validation")
		return err
	}

	logger.Validation.Debug().
		Str("schema", validationCtx.Schema.Name()).
		Msg("full validation passed")

	return nil
}

// NewValidationContext creates a validation context.
// This is a helper function to create validation contexts easily.
//
// Example:
//
//	ctx := validation.NewValidationContext(userSchema, userRecord, userRepo)
//
//	// Use context for validation
//	engine := validation.GetGlobalValidationEngine()
//	err := engine.ValidateAll(context.Background(), ctx)
func NewValidationContext(schema schema.JSchema, record schema.JRecord, repository interface{}) ValidationContext {
	return ValidationContext{
		Schema:     schema,
		Record:     record,
		Repository: repository,
	}
}
