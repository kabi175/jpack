package schema

// jField implements JField interface
type jField struct {
	name         string
	fieldType    JFieldType
	defaultValue any
	required     bool
	unique       bool
	validation   ValidationFunc
	immutable    bool
}

// NewJField creates a new field
func NewJField(name string, fieldType JFieldType, defaultValue any) JField {
	return &jField{
		name:         name,
		fieldType:    fieldType,
		defaultValue: defaultValue,
		required:     false,
		unique:       false,
		immutable:    false,
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

func (f *jField) SetRequired(required bool) JField {
	if f.immutable {
		panic("cannot modify immutable field")
	}
	f.required = required
	return f
}

func (f *jField) IsUnique() bool {
	return f.unique
}

func (f *jField) SetUnique(unique bool) JField {
	if f.immutable {
		panic("cannot modify immutable field")
	}
	f.unique = unique
	return f
}

func (f *jField) Validation() ValidationFunc {
	return f.validation
}

func (f *jField) SetValidation(fn ValidationFunc) JField {
	if f.immutable {
		panic("cannot modify immutable field")
	}
	f.validation = fn
	return f
}

// IsImmutable returns whether the field is immutable
func (f *jField) IsImmutable() bool {
	return f.immutable
}

// Freeze makes the field immutable
func (f *jField) Freeze() JField {
	f.immutable = true
	return f
}
