package schema

// SchemaBuilder provides a fluent interface for building schemas.
// SchemaBuilder works with a mutable internal schema that is made immutable when Build() is called.
//
// Example:
//
//	schema := schema.NewSchemaBuilder("User").
//		AddRequiredField("name", schema.JString, nil).
//		AddRequiredUniqueField("email", schema.JString, nil).
//		AddField("age", schema.JInt, 18).
//		SetIDField("id").
//		AddValidation(emailValidator).
//		Build()
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

// NewSchemaBuilder creates a new schema builder with automatic ID field.
// The builder automatically creates an "id" field as the primary key.
//
// Example:
//
//	builder := schema.NewSchemaBuilder("User")
//	// This automatically adds an "id" field of type JString
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

// AddField adds a field to the schema.
// The field is optional by default.
//
// Example:
//
//	schema := schema.NewSchemaBuilder("User").
//		AddField("name", schema.JString, nil).
//		AddField("age", schema.JInt, 18).
//		Build()
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

// AddRequiredField adds a required field to the schema.
// Required fields must be present when creating records.
//
// Example:
//
//	schema := schema.NewSchemaBuilder("User").
//		AddRequiredField("name", schema.JString, nil).
//		AddRequiredField("email", schema.JString, nil).
//		Build()
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

// AddUniqueField adds a unique field to the schema.
// Unique fields must have unique values across all records.
//
// Example:
//
//	schema := schema.NewSchemaBuilder("User").
//		AddUniqueField("username", schema.JString, nil).
//		AddUniqueField("email", schema.JString, nil).
//		Build()
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

// AddRequiredUniqueField adds a required and unique field to the schema.
// The field must be present and have unique values across all records.
//
// Example:
//
//	schema := schema.NewSchemaBuilder("User").
//		AddRequiredUniqueField("email", schema.JString, nil).
//		AddRequiredUniqueField("username", schema.JString, nil).
//		Build()
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

// AddFieldWithValidation adds a field with validation to the schema.
// The validation function will be called when validating records.
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
//	schema := schema.NewSchemaBuilder("User").
//		AddFieldWithValidation("email", schema.JString, nil, emailValidator).
//		Build()
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

// AddRef adds a reference to the schema.
// References establish relationships between schemas.
//
// Example:
//
//	profileSchema := schema.NewSchemaBuilder("Profile").Build()
//	userSchema := schema.NewSchemaBuilder("User").
//		AddRef("profile", profileSchema).
//		Build()
func (sb *SchemaBuilder) AddRef(name string, targetSchema JSchema) *SchemaBuilder {
	sb.refs[name] = &refConfig{
		name:   name,
		schema: targetSchema,
		array:  false,
	}
	return sb
}

// AddArrayRef adds an array reference to the schema.
// Array references allow multiple instances of the target schema.
//
// Example:
//
//	orderSchema := schema.NewSchemaBuilder("Order").Build()
//	userSchema := schema.NewSchemaBuilder("User").
//		AddArrayRef("orders", orderSchema).
//		Build()
func (sb *SchemaBuilder) AddArrayRef(name string, targetSchema JSchema) *SchemaBuilder {
	sb.refs[name] = &refConfig{
		name:   name,
		schema: targetSchema,
		array:  true,
	}
	return sb
}

// AddEdge adds an edge to the schema.
// Edges define relationships between schemas.
//
// Example:
//
//	userSchema := schema.NewSchemaBuilder("User").Build()
//	orderSchema := schema.NewSchemaBuilder("Order").Build()
//	edge := schema.NewJEdge("user_orders", userSchema, orderSchema, schema.EdgeOneToMany)
//
//	userSchema = userSchema.Update(func(sb *schema.SchemaBuilder) {
//		sb.AddEdge(edge)
//	})
func (sb *SchemaBuilder) AddEdge(edge JEdge) *SchemaBuilder {
	sb.edges = append(sb.edges, edge)
	return sb
}

// AddValidation adds a validation function to the schema.
// Schema-level validations are applied to the entire record.
//
// Example:
//
//	ageValidator := func(ctx context.Context, rec schema.JRecord) error {
//		age := rec.Get("age").(int)
//		if age < 18 {
//			return fmt.Errorf("age must be at least 18")
//		}
//		return nil
//	}
//
//	schema := schema.NewSchemaBuilder("User").
//		AddField("age", schema.JInt, 18).
//		AddValidation(ageValidator).
//		Build()
func (sb *SchemaBuilder) AddValidation(fn ValidationFunc) *SchemaBuilder {
	sb.validations = append(sb.validations, fn)
	return sb
}

// SetIDField sets the ID field for the schema.
// The ID field is used as the primary key for records.
//
// Example:
//
//	schema := schema.NewSchemaBuilder("User").
//		AddField("user_id", schema.JString, nil).
//		SetIDField("user_id").
//		Build()
func (sb *SchemaBuilder) SetIDField(fieldName string) *SchemaBuilder {
	if _, exists := sb.fields[fieldName]; exists {
		sb.idFieldName = fieldName
	}
	return sb
}

// Build builds and returns an immutable schema.
// Once built, the schema cannot be modified directly.
//
// Example:
//
//	schema := schema.NewSchemaBuilder("User").
//		AddField("name", schema.JString, nil).
//		Build()
//
//	// Schema is now immutable
//	fmt.Println(schema.IsImmutable()) // true
func (sb *SchemaBuilder) Build() JSchema {
	// Create the schema with all fields, refs, edges, and validations
	schema := newJSchemaFromBuilder(sb)
	return schema
}

// FieldBuilder provides a fluent interface for building fields.
// Fields are immutable once built.
//
// Example:
//
//	field := schema.NewFieldBuilder("email", schema.JString, nil).
//		Required().
//		Unique().
//		WithValidation(emailValidator).
//		Build()
type FieldBuilder struct {
	name         string
	fieldType    JFieldType
	defaultValue any
	required     bool
	unique       bool
	validation   ValidationFunc
}

// NewFieldBuilder creates a new field builder.
// The field is optional by default.
//
// Example:
//
//	field := schema.NewFieldBuilder("name", schema.JString, "default").
//		Required().
//		Build()
func NewFieldBuilder(name string, fieldType JFieldType, defaultValue any) *FieldBuilder {
	return &FieldBuilder{
		name:         name,
		fieldType:    fieldType,
		defaultValue: defaultValue,
		required:     false,
		unique:       false,
	}
}

// SetRequired makes the field required.
// Required fields must be present when creating records.
//
// Example:
//
//	field := schema.NewFieldBuilder("name", schema.JString, nil).
//		SetRequired(true).
//		Build()
func (fb *FieldBuilder) SetRequired(required bool) *FieldBuilder {
	fb.required = required
	return fb
}

// Required makes the field required.
// This is a convenience method that calls SetRequired(true).
//
// Example:
//
//	field := schema.NewFieldBuilder("name", schema.JString, nil).
//		Required().
//		Build()
func (fb *FieldBuilder) Required() *FieldBuilder {
	fb.required = true
	return fb
}

// SetUnique makes the field unique.
// Unique fields must have unique values across all records.
//
// Example:
//
//	field := schema.NewFieldBuilder("email", schema.JString, nil).
//		SetUnique(true).
//		Build()
func (fb *FieldBuilder) SetUnique(unique bool) *FieldBuilder {
	fb.unique = unique
	return fb
}

// Unique makes the field unique.
// This is a convenience method that calls SetUnique(true).
//
// Example:
//
//	field := schema.NewFieldBuilder("email", schema.JString, nil).
//		Unique().
//		Build()
func (fb *FieldBuilder) Unique() *FieldBuilder {
	fb.unique = true
	return fb
}

// WithValidation adds a validation function to the field.
// The validation function will be called when validating records.
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
//	field := schema.NewFieldBuilder("email", schema.JString, nil).
//		WithValidation(emailValidator).
//		Build()
func (fb *FieldBuilder) WithValidation(fn ValidationFunc) *FieldBuilder {
	fb.validation = fn
	return fb
}

// Build builds and returns an immutable field.
// Once built, the field cannot be modified directly.
//
// Example:
//
//	field := schema.NewFieldBuilder("name", schema.JString, nil).
//		Required().
//		Build()
//
//	// Field is now immutable
//	fmt.Println(field.IsImmutable()) // true
func (fb *FieldBuilder) Build() JField {
	return newJFieldFromBuilder(fb)
}

// SchemaRegistryBuilder provides a fluent interface for building and managing schemas in a registry.
// It allows for easy registration and management of multiple schemas.
//
// Example:
//
//	registry := schema.NewSchemaRegistryBuilder().
//		RegisterSchema(userSchema).
//		RegisterSchema(orderSchema).
//		Build()
type SchemaRegistryBuilder struct {
	registry JSchemaRegistry
}

// NewSchemaRegistryBuilder creates a new schema registry builder.
// It initializes a new registry for schema management.
//
// Example:
//
//	builder := schema.NewSchemaRegistryBuilder()
func NewSchemaRegistryBuilder() *SchemaRegistryBuilder {
	return &SchemaRegistryBuilder{
		registry: NewJSchemaRegistry(),
	}
}

// NewSchemaRegistryBuilderWithRegistry creates a new schema registry builder with an existing registry.
// This allows you to use an existing registry instead of creating a new one.
//
// Example:
//
//	existingRegistry := schema.NewJSchemaRegistry()
//	builder := schema.NewSchemaRegistryBuilderWithRegistry(existingRegistry)
func NewSchemaRegistryBuilderWithRegistry(registry JSchemaRegistry) *SchemaRegistryBuilder {
	return &SchemaRegistryBuilder{
		registry: registry,
	}
}

// RegisterSchema registers a schema in the registry.
// This allows the schema to be retrieved later by name.
//
// Example:
//
//	builder := schema.NewSchemaRegistryBuilder().
//		RegisterSchema(userSchema).
//		RegisterSchema(orderSchema)
func (srb *SchemaRegistryBuilder) RegisterSchema(schema JSchema) *SchemaRegistryBuilder {
	srb.registry.Register(schema)
	return srb
}

// RegisterImmutableSchema registers an immutable schema in the registry.
// Deprecated: All schemas are immutable now, use RegisterSchema instead.
//
// Example:
//
//	builder := schema.NewSchemaRegistryBuilder().
//		RegisterImmutableSchema(userSchema) // Deprecated
func (srb *SchemaRegistryBuilder) RegisterImmutableSchema(schema JSchema) *SchemaRegistryBuilder {
	srb.registry.Register(schema)
	return srb
}

// ReplaceSchema replaces an existing schema in the registry.
// This is useful for schema evolution and updates.
//
// Example:
//
//	builder := schema.NewSchemaRegistryBuilder().
//		RegisterSchema(userSchemaV1).
//		ReplaceSchema(userSchemaV2) // Replaces V1 with V2
func (srb *SchemaRegistryBuilder) ReplaceSchema(schema JSchema) *SchemaRegistryBuilder {
	srb.registry.Replace(schema)
	return srb
}

// Build builds and returns the registry.
// The registry contains all schemas that were registered during the build process.
//
// Example:
//
//	registry := schema.NewSchemaRegistryBuilder().
//		RegisterSchema(userSchema).
//		RegisterSchema(orderSchema).
//		Build()
//
//	// Use the registry
//	userSchema, ok := registry.Get("User")
func (srb *SchemaRegistryBuilder) Build() JSchemaRegistry {
	return srb.registry
}

// GetRegistry returns the current registry.
// This allows access to the registry before calling Build().
//
// Example:
//
//	builder := schema.NewSchemaRegistryBuilder().
//		RegisterSchema(userSchema)
//
//	registry := builder.GetRegistry()
//	userSchema, ok := registry.Get("User")
func (srb *SchemaRegistryBuilder) GetRegistry() JSchemaRegistry {
	return srb.registry
}
