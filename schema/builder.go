package schema

// SchemaBuilder provides a fluent interface for building schemas
// SchemaBuilder works with a mutable internal schema that is made immutable when Build() is called
type SchemaBuilder struct {
	name        string
	fields      map[string]*fieldBuilder
	refs        map[string]*refConfig
	edges       []JEdge
	validations []ValidationFunc
	idFieldName string
}

type fieldBuilder struct {
	name         string
	fieldType    JFieldType
	defaultValue any
	required     bool
	unique       bool
	validation   ValidationFunc
}

type refConfig struct {
	name   string
	schema JSchema
	array  bool
}

// NewSchemaBuilder creates a new schema builder with automatic ID field
func NewSchemaBuilder(name string) *SchemaBuilder {
	sb := &SchemaBuilder{
		name:        name,
		fields:      make(map[string]*fieldBuilder),
		refs:        make(map[string]*refConfig),
		edges:       make([]JEdge, 0),
		validations: make([]ValidationFunc, 0),
		idFieldName: "id", // Default ID field name
	}

	// Automatically add ID field as primary key
	sb.fields["id"] = &fieldBuilder{
		name:         "id",
		fieldType:    JString,
		defaultValue: nil,
		required:     true,
		unique:       true,
	}

	return sb
}

// AddField adds a field to the schema
func (sb *SchemaBuilder) AddField(name string, fieldType JFieldType, defaultValue any) *SchemaBuilder {
	// Allow overriding the default ID field if explicitly adding an "id" field
	sb.fields[name] = &fieldBuilder{
		name:         name,
		fieldType:    fieldType,
		defaultValue: defaultValue,
		required:     false,
		unique:       false,
	}
	return sb
}

// AddRequiredField adds a required field to the schema
func (sb *SchemaBuilder) AddRequiredField(name string, fieldType JFieldType, defaultValue any) *SchemaBuilder {
	sb.fields[name] = &fieldBuilder{
		name:         name,
		fieldType:    fieldType,
		defaultValue: defaultValue,
		required:     true,
		unique:       false,
	}
	return sb
}

// AddUniqueField adds a unique field to the schema
func (sb *SchemaBuilder) AddUniqueField(name string, fieldType JFieldType, defaultValue any) *SchemaBuilder {
	sb.fields[name] = &fieldBuilder{
		name:         name,
		fieldType:    fieldType,
		defaultValue: defaultValue,
		required:     false,
		unique:       true,
	}
	return sb
}

// AddRequiredUniqueField adds a required and unique field to the schema
func (sb *SchemaBuilder) AddRequiredUniqueField(name string, fieldType JFieldType, defaultValue any) *SchemaBuilder {
	sb.fields[name] = &fieldBuilder{
		name:         name,
		fieldType:    fieldType,
		defaultValue: defaultValue,
		required:     true,
		unique:       true,
	}
	return sb
}

// AddFieldWithValidation adds a field with validation to the schema
func (sb *SchemaBuilder) AddFieldWithValidation(name string, fieldType JFieldType, defaultValue any, validation ValidationFunc) *SchemaBuilder {
	sb.fields[name] = &fieldBuilder{
		name:         name,
		fieldType:    fieldType,
		defaultValue: defaultValue,
		required:     false,
		unique:       false,
		validation:   validation,
	}
	return sb
}

// AddRef adds a reference to the schema
func (sb *SchemaBuilder) AddRef(name string, targetSchema JSchema) *SchemaBuilder {
	sb.refs[name] = &refConfig{
		name:   name,
		schema: targetSchema,
		array:  false,
	}
	return sb
}

// AddArrayRef adds an array reference to the schema
func (sb *SchemaBuilder) AddArrayRef(name string, targetSchema JSchema) *SchemaBuilder {
	sb.refs[name] = &refConfig{
		name:   name,
		schema: targetSchema,
		array:  true,
	}
	return sb
}

// AddEdge adds an edge to the schema
func (sb *SchemaBuilder) AddEdge(edge JEdge) *SchemaBuilder {
	sb.edges = append(sb.edges, edge)
	return sb
}

// AddValidation adds a validation function to the schema
func (sb *SchemaBuilder) AddValidation(fn ValidationFunc) *SchemaBuilder {
	sb.validations = append(sb.validations, fn)
	return sb
}

// SetIDField sets the ID field for the schema
func (sb *SchemaBuilder) SetIDField(fieldName string) *SchemaBuilder {
	if _, exists := sb.fields[fieldName]; exists {
		sb.idFieldName = fieldName
	}
	return sb
}

// Build builds and returns an immutable schema
func (sb *SchemaBuilder) Build() JSchema {
	// Create the schema with all fields, refs, edges, and validations
	schema := newJSchemaFromBuilder(sb)
	return schema
}

// FieldBuilder provides a fluent interface for building fields
type FieldBuilder struct {
	name         string
	fieldType    JFieldType
	defaultValue any
	required     bool
	unique       bool
	validation   ValidationFunc
}

// NewFieldBuilder creates a new field builder
func NewFieldBuilder(name string, fieldType JFieldType, defaultValue any) *FieldBuilder {
	return &FieldBuilder{
		name:         name,
		fieldType:    fieldType,
		defaultValue: defaultValue,
		required:     false,
		unique:       false,
	}
}

// SetRequired makes the field required
func (fb *FieldBuilder) SetRequired(required bool) *FieldBuilder {
	fb.required = required
	return fb
}

// Required makes the field required
func (fb *FieldBuilder) Required() *FieldBuilder {
	fb.required = true
	return fb
}

// SetUnique makes the field unique
func (fb *FieldBuilder) SetUnique(unique bool) *FieldBuilder {
	fb.unique = unique
	return fb
}

// Unique makes the field unique
func (fb *FieldBuilder) Unique() *FieldBuilder {
	fb.unique = true
	return fb
}

// WithValidation adds a validation function to the field
func (fb *FieldBuilder) WithValidation(fn ValidationFunc) *FieldBuilder {
	fb.validation = fn
	return fb
}

// Build builds and returns an immutable field
func (fb *FieldBuilder) Build() JField {
	return newJFieldFromBuilder(fb)
}

// SchemaRegistryBuilder provides a fluent interface for building and managing schemas in a registry
type SchemaRegistryBuilder struct {
	registry JSchemaRegistry
}

// NewSchemaRegistryBuilder creates a new schema registry builder
func NewSchemaRegistryBuilder() *SchemaRegistryBuilder {
	return &SchemaRegistryBuilder{
		registry: NewJSchemaRegistry(),
	}
}

// NewSchemaRegistryBuilderWithRegistry creates a new schema registry builder with an existing registry
func NewSchemaRegistryBuilderWithRegistry(registry JSchemaRegistry) *SchemaRegistryBuilder {
	return &SchemaRegistryBuilder{
		registry: registry,
	}
}

// RegisterSchema registers a schema in the registry
func (srb *SchemaRegistryBuilder) RegisterSchema(schema JSchema) *SchemaRegistryBuilder {
	srb.registry.Register(schema)
	return srb
}

// RegisterImmutableSchema registers an immutable schema in the registry (deprecated - all schemas are immutable now)
func (srb *SchemaRegistryBuilder) RegisterImmutableSchema(schema JSchema) *SchemaRegistryBuilder {
	srb.registry.Register(schema)
	return srb
}

// ReplaceSchema replaces an existing schema in the registry
func (srb *SchemaRegistryBuilder) ReplaceSchema(schema JSchema) *SchemaRegistryBuilder {
	srb.registry.Replace(schema)
	return srb
}

// Build builds and returns the registry
func (srb *SchemaRegistryBuilder) Build() JSchemaRegistry {
	return srb.registry
}

// GetRegistry returns the current registry
func (srb *SchemaRegistryBuilder) GetRegistry() JSchemaRegistry {
	return srb.registry
}
