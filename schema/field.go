package schema

import "fmt"

// jField implements JField interface (always immutable)
type jField struct {
	name         string
	fieldType    JFieldType
	defaultValue any
	required     bool
	unique       bool
	validation   ValidationFunc
}

// newJFieldFromBuilder creates a new immutable field from a builder
func newJFieldFromBuilder(fb *FieldBuilder) JField {
	return &jField{
		name:         fb.name,
		fieldType:    fb.fieldType,
		defaultValue: fb.defaultValue,
		required:     fb.required,
		unique:       fb.unique,
		validation:   fb.validation,
	}
}

// NewJField creates a new immutable field
func NewJField(name string, fieldType JFieldType, defaultValue any) JField {
	return &jField{
		name:         name,
		fieldType:    fieldType,
		defaultValue: defaultValue,
		required:     false,
		unique:       false,
	}
}

func (f *jField) Name() string {
	return f.name
}

func (f *jField) Type() JFieldType {
	return f.fieldType
}

func (f *jField) DefaultValue() any {
	return f.defaultValue
}

func (f *jField) IsRequired() bool {
	return f.required
}

func (f *jField) IsUnique() bool {
	return f.unique
}

func (f *jField) Validation() ValidationFunc {
	return f.validation
}

func (f *jField) String() string {
	return fmt.Sprintf("%s:%s", f.name, f.fieldType)
}

// Update creates a new field with modifications applied via callback
func (f *jField) Update(fn func(*FieldBuilder)) JField {
	// Create a new builder with a copy of the current field state
	builder := &FieldBuilder{
		name:         f.name,
		fieldType:    f.fieldType,
		defaultValue: f.defaultValue,
		required:     f.required,
		unique:       f.unique,
		validation:   f.validation,
	}

	// Apply the modifications
	fn(builder)

	// Build and return the new immutable field
	return builder.Build()
}

// Legacy methods for backward compatibility

// SetRequired creates a new field with the required flag set (for backward compatibility)
func (f *jField) SetRequired(required bool) JField {
	return &jField{
		name:         f.name,
		fieldType:    f.fieldType,
		defaultValue: f.defaultValue,
		required:     required,
		unique:       f.unique,
		validation:   f.validation,
	}
}

// SetUnique creates a new field with the unique flag set (for backward compatibility)
func (f *jField) SetUnique(unique bool) JField {
	return &jField{
		name:         f.name,
		fieldType:    f.fieldType,
		defaultValue: f.defaultValue,
		required:     f.required,
		unique:       unique,
		validation:   f.validation,
	}
}

// SetValidation creates a new field with the validation function set (for backward compatibility)
func (f *jField) SetValidation(fn ValidationFunc) JField {
	return &jField{
		name:         f.name,
		fieldType:    f.fieldType,
		defaultValue: f.defaultValue,
		required:     f.required,
		unique:       f.unique,
		validation:   fn,
	}
}

// IsImmutable always returns true (for backward compatibility)
func (f *jField) IsImmutable() bool {
	return true
}

// Freeze returns the field itself (for backward compatibility)
func (f *jField) Freeze() JField {
	return f
}
