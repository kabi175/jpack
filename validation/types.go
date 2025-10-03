package validation

import (
	"context"
	"fmt"
	"strings"

	"github.com/kabi175/jpack/logger"
	"github.com/kabi175/jpack/schema"
)

// ValidationFunc represents a validation function
type ValidationFunc func(ctx context.Context, rec schema.JRecord) error

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
	Value   any
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

// ValidationResult represents the result of validation
type ValidationResult struct {
	Valid  bool
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

// FieldValidator represents a field-level validator
type FieldValidator interface {
	Validate(ctx context.Context, field schema.JField, value any) error
}

// RecordValidator represents a record-level validator
type RecordValidator interface {
	Validate(ctx context.Context, schema schema.JSchema, rec schema.JRecord) error
}

// SchemaValidator represents a schema-level validator
type SchemaValidator interface {
	Validate(ctx context.Context, schema schema.JSchema) error
}

// ValidationRule represents a validation rule
type ValidationRule struct {
	Name        string
	Description string
	Validator   ValidationFunc
}

// ValidationRegistry manages validation rules
type ValidationRegistry interface {
	RegisterRule(rule ValidationRule) error
	GetRule(name string) (ValidationRule, bool)
	UnregisterRule(name string) error
	ListRules() []string
}

// jValidationRegistry implements ValidationRegistry
type jValidationRegistry struct {
	rules map[string]ValidationRule
}

// NewValidationRegistry creates a new validation registry
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

// ValidationContext provides context for validation
type ValidationContext struct {
	Schema     schema.JSchema
	Record     schema.JRecord
	Field      schema.JField
	Value      any
	Repository interface{} // Optional repository for unique checks
}

// ValidationEngine orchestrates validation
type ValidationEngine interface {
	ValidateField(ctx context.Context, validationCtx ValidationContext) error
	ValidateRecord(ctx context.Context, validationCtx ValidationContext) error
	ValidateSchema(ctx context.Context, validationCtx ValidationContext) error
	ValidateAll(ctx context.Context, validationCtx ValidationContext) error
}

// jValidationEngine implements ValidationEngine
type jValidationEngine struct {
	registry ValidationRegistry
}

// NewValidationEngine creates a new validation engine
func NewValidationEngine(registry ValidationRegistry) ValidationEngine {
	return &jValidationEngine{
		registry: registry,
	}
}

// GetGlobalValidationEngine returns the global validation engine
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

// Helper function to create validation context
func NewValidationContext(schema schema.JSchema, record schema.JRecord, repository interface{}) ValidationContext {
	return ValidationContext{
		Schema:     schema,
		Record:     record,
		Repository: repository,
	}
}
