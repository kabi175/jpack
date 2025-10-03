package schema

// SchemaBuilder provides a fluent interface for building schemas
type SchemaBuilder struct {
	schema JSchema
}

// NewSchemaBuilder creates a new schema builder
func NewSchemaBuilder(name string) *SchemaBuilder {
	return &SchemaBuilder{
		schema: NewJSchema(name),
	}
}

// NewImmutableSchemaBuilder creates a new immutable schema builder
func NewImmutableSchemaBuilder(name string) *SchemaBuilder {
	return &SchemaBuilder{
		schema: NewImmutableJSchema(name),
	}
}

// AddField adds a field to the schema
func (sb *SchemaBuilder) AddField(name string, fieldType JFieldType, defaultValue any) *SchemaBuilder {
	sb.schema.AddField(name, fieldType, defaultValue)
	return sb
}

// AddRequiredField adds a required field to the schema
func (sb *SchemaBuilder) AddRequiredField(name string, fieldType JFieldType, defaultValue any) *SchemaBuilder {
	field := sb.schema.AddField(name, fieldType, defaultValue)
	field.SetRequired(true)
	return sb
}

// AddUniqueField adds a unique field to the schema
func (sb *SchemaBuilder) AddUniqueField(name string, fieldType JFieldType, defaultValue any) *SchemaBuilder {
	field := sb.schema.AddField(name, fieldType, defaultValue)
	field.SetUnique(true)
	return sb
}

// AddRequiredUniqueField adds a required and unique field to the schema
func (sb *SchemaBuilder) AddRequiredUniqueField(name string, fieldType JFieldType, defaultValue any) *SchemaBuilder {
	field := sb.schema.AddField(name, fieldType, defaultValue)
	field.SetRequired(true).SetUnique(true)
	return sb
}

// AddRef adds a reference to the schema
func (sb *SchemaBuilder) AddRef(name string, targetSchema JSchema) *SchemaBuilder {
	sb.schema.AddRef(name, targetSchema)
	return sb
}

// AddEdge adds an edge to the schema
func (sb *SchemaBuilder) AddEdge(edge JEdge) *SchemaBuilder {
	sb.schema.AddEdge(edge)
	return sb
}

// AddValidation adds a validation function to the schema
func (sb *SchemaBuilder) AddValidation(fn ValidationFunc) *SchemaBuilder {
	sb.schema.AddValidation(fn)
	return sb
}

// SetIDField sets the ID field for the schema
func (sb *SchemaBuilder) SetIDField(fieldName string) *SchemaBuilder {
	if field, exists := sb.schema.Field(fieldName); exists {
		sb.schema.SetIDField(field)
	}
	return sb
}

// Build builds and returns the schema
func (sb *SchemaBuilder) Build() JSchema {
	return sb.schema
}

// BuildImmutable builds and returns an immutable schema
func (sb *SchemaBuilder) BuildImmutable() JSchema {
	return sb.schema.Freeze()
}

// FieldBuilder provides a fluent interface for building fields
type FieldBuilder struct {
	field JField
}

// NewFieldBuilder creates a new field builder
func NewFieldBuilder(name string, fieldType JFieldType, defaultValue any) *FieldBuilder {
	return &FieldBuilder{
		field: NewJField(name, fieldType, defaultValue),
	}
}

// Required makes the field required
func (fb *FieldBuilder) Required() *FieldBuilder {
	fb.field.SetRequired(true)
	return fb
}

// Unique makes the field unique
func (fb *FieldBuilder) Unique() *FieldBuilder {
	fb.field.SetUnique(true)
	return fb
}

// WithValidation adds a validation function to the field
func (fb *FieldBuilder) WithValidation(fn ValidationFunc) *FieldBuilder {
	fb.field.SetValidation(fn)
	return fb
}

// Build builds and returns the field
func (fb *FieldBuilder) Build() JField {
	return fb.field
}

// BuildImmutable builds and returns an immutable field
func (fb *FieldBuilder) BuildImmutable() JField {
	return fb.field.Freeze()
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

// RegisterImmutableSchema registers an immutable schema in the registry
func (srb *SchemaRegistryBuilder) RegisterImmutableSchema(schema JSchema) *SchemaRegistryBuilder {
	srb.registry.RegisterImmutable(schema)
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
