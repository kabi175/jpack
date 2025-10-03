package schema

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileSchemaSource_Load_SingleSchema(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		content  string
		wantErr  bool
	}{
		{
			name:     "valid single YAML schema",
			filename: "test_single.yaml",
			content: `
name: User
fields:
  id:
    type: string
    required: true
    unique: true
  name:
    type: string
    required: true
  age:
    type: int
    defaultValue: 18
`,
			wantErr: false,
		},
		{
			name:     "valid single JSON schema",
			filename: "test_single.json",
			content: `{
  "name": "Product",
  "fields": {
    "id": {
      "type": "string",
      "required": true,
      "unique": true
    },
    "name": {
      "type": "string",
      "required": true
    },
    "price": {
      "type": "float64",
      "required": true
    }
  }
}`,
			wantErr: false,
		},
		{
			name:     "invalid YAML content",
			filename: "test_invalid.yaml",
			content: `
name: User
fields:
  id:
    type: string
    required: true
  name: # missing value
`,
			wantErr: true,
		},
		{
			name:     "invalid JSON content",
			filename: "test_invalid.json",
			content: `{
  "name": "Product",
  "fields": {
    "id": {
      "type": "string",
      "required": true
    }
    // missing comma
    "name": {
      "type": "string"
    }
  }
}`,
			wantErr: true,
		},
		{
			name:     "empty schema name",
			filename: "test_empty_name.yaml",
			content: `
name: ""
fields:
  id:
    type: string
    required: true
`,
			wantErr: true,
		},
		{
			name:     "missing schema name",
			filename: "test_no_name.yaml",
			content: `
fields:
  id:
    type: string
    required: true
`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary file
			tmpFile, err := os.CreateTemp("", tt.filename)
			require.NoError(t, err)
			defer os.Remove(tmpFile.Name())

			_, err = tmpFile.WriteString(tt.content)
			require.NoError(t, err)
			tmpFile.Close()

			// Test FileSchemaSource
			source := NewFileSchemaSource(tmpFile.Name())
			schema, err := source.Load("")

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, schema)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, schema)
				assert.NotEmpty(t, schema.Name())
				assert.Greater(t, len(schema.Fields()), 0)
			}
		})
	}
}

func TestFileSchemaSource_Load_BulkSchema(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		content  string
		wantErr  bool
	}{
		{
			name:     "valid bulk YAML schema",
			filename: "test_bulk.yaml",
			content: `
- name: User
  fields:
    id:
      type: string
      required: true
    name:
      type: string
      required: true
- name: Product
  fields:
    id:
      type: string
      required: true
    name:
      type: string
      required: true
`,
			wantErr: false,
		},
		{
			name:     "valid bulk JSON schema",
			filename: "test_bulk.json",
			content: `[
  {
    "name": "User",
    "fields": {
      "id": {
        "type": "string",
        "required": true
      },
      "name": {
        "type": "string",
        "required": true
      }
    }
  },
  {
    "name": "Product",
    "fields": {
      "id": {
        "type": "string",
        "required": true
      },
      "name": {
        "type": "string",
        "required": true
      }
    }
  }
]`,
			wantErr: false,
		},
		{
			name:     "empty bulk schema",
			filename: "test_empty_bulk.yaml",
			content:  `[]`,
			wantErr:  true,
		},
		{
			name:     "invalid bulk YAML",
			filename: "test_invalid_bulk.yaml",
			content: `
- name: User
  fields:
    id:
      type: string
      required: true
    # invalid YAML structure
- name: Product
  fields: [invalid yaml array]
`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary file
			tmpFile, err := os.CreateTemp("", tt.filename)
			require.NoError(t, err)
			defer os.Remove(tmpFile.Name())

			_, err = tmpFile.WriteString(tt.content)
			require.NoError(t, err)
			tmpFile.Close()

			// Test FileSchemaSource
			source := NewFileSchemaSource(tmpFile.Name())
			schema, err := source.Load("")

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, schema)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, schema)
				assert.NotEmpty(t, schema.Name())
			}
		})
	}
}

func TestFileSchemaSource_Load_SpecificSchema(t *testing.T) {
	content := `
- name: User
  fields:
    id:
      type: string
      required: true
    name:
      type: string
      required: true
- name: Product
  fields:
    id:
      type: string
      required: true
    name:
      type: string
      required: true
- name: Order
  fields:
    id:
      type: string
      required: true
    total:
      type: float64
      required: true
`

	// Create temporary file
	tmpFile, err := os.CreateTemp("", "test_specific.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(content)
	require.NoError(t, err)
	tmpFile.Close()

	source := NewFileSchemaSource(tmpFile.Name())

	tests := []struct {
		name         string
		schemaName   string
		expectedName string
		wantErr      bool
	}{
		{
			name:         "load existing User schema",
			schemaName:   "User",
			expectedName: "User",
			wantErr:      false,
		},
		{
			name:         "load existing Product schema",
			schemaName:   "Product",
			expectedName: "Product",
			wantErr:      false,
		},
		{
			name:         "load existing Order schema",
			schemaName:   "Order",
			expectedName: "Order",
			wantErr:      false,
		},
		{
			name:         "load non-existent schema",
			schemaName:   "NonExistent",
			expectedName: "",
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema, err := source.Load(tt.schemaName)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, schema)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, schema)
				assert.Equal(t, tt.expectedName, schema.Name())
			}
		})
	}
}

func TestFileSchemaSource_Load_FileNotFound(t *testing.T) {
	source := NewFileSchemaSource("non_existent_file.yaml")
	schema, err := source.Load("")

	assert.Error(t, err)
	assert.Nil(t, schema)
	assert.Contains(t, err.Error(), "failed to read schema file")
}

func TestFileSchemaSource_Load_NoExtension(t *testing.T) {
	content := `
name: User
fields:
  id:
    type: string
    required: true
  name:
    type: string
    required: true
`

	// Create temporary file without extension
	tmpFile, err := os.CreateTemp("", "test_no_ext")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(content)
	require.NoError(t, err)
	tmpFile.Close()

	source := NewFileSchemaSource(tmpFile.Name())
	schema, err := source.Load("")

	assert.NoError(t, err)
	assert.NotNil(t, schema)
	assert.Equal(t, "User", schema.Name())
}

func TestFileSchemaSource_tryUnmarshal(t *testing.T) {
	source := &FileSchemaSource{}

	tests := []struct {
		name     string
		filename string
		content  string
		target   interface{}
		wantErr  bool
	}{
		{
			name:     "valid JSON with .json extension",
			filename: "test.json",
			content:  `{"name": "test"}`,
			target:   &map[string]interface{}{},
			wantErr:  false,
		},
		{
			name:     "valid YAML with .yaml extension",
			filename: "test.yaml",
			content:  `name: test`,
			target:   &map[string]interface{}{},
			wantErr:  false,
		},
		{
			name:     "valid YAML with .yml extension",
			filename: "test.yml",
			content:  `name: test`,
			target:   &map[string]interface{}{},
			wantErr:  false,
		},
		{
			name:     "invalid JSON with .json extension",
			filename: "test.json",
			content:  `{invalid json}`,
			target:   &map[string]interface{}{},
			wantErr:  true,
		},
		{
			name:     "invalid YAML with .yaml extension",
			filename: "test.yaml",
			content:  `invalid: yaml: content: [`,
			target:   &map[string]interface{}{},
			wantErr:  true,
		},
		{
			name:     "no extension - try JSON first",
			filename: "test",
			content:  `{"name": "test"}`,
			target:   &map[string]interface{}{},
			wantErr:  false,
		},
		{
			name:     "no extension - fallback to YAML",
			filename: "test",
			content:  `name: test`,
			target:   &map[string]interface{}{},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source.Path = tt.filename
			err := source.tryUnmarshal([]byte(tt.content), tt.target)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConvertDefToSchema(t *testing.T) {
	tests := []struct {
		name    string
		def     SchemaDef
		wantErr bool
	}{
		{
			name: "valid schema with fields",
			def: SchemaDef{
				Name: "User",
				Fields: map[string]FieldDef{
					"id": {
						Type:     "string",
						Required: true,
						Unique:   true,
					},
					"name": {
						Type:     "string",
						Required: true,
					},
					"age": {
						Type:         "int",
						DefaultValue: 18,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid schema with refs and edges",
			def: SchemaDef{
				Name: "Order",
				Fields: map[string]FieldDef{
					"id": {
						Type:     "string",
						Required: true,
					},
				},
				Refs: map[string]RefDef{
					"user": {
						TargetSchema: "User",
						IsArray:      false,
					},
				},
				Edges: []EdgeDef{
					{
						Name:         "order_user",
						TargetSchema: "User",
						Type:         "many-to-one",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "empty schema name",
			def: SchemaDef{
				Name: "",
				Fields: map[string]FieldDef{
					"id": {
						Type: "string",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid field type",
			def: SchemaDef{
				Name: "Test",
				Fields: map[string]FieldDef{
					"id": {
						Type: "invalid_type",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid edge type",
			def: SchemaDef{
				Name: "Test",
				Fields: map[string]FieldDef{
					"id": {
						Type: "string",
					},
				},
				Edges: []EdgeDef{
					{
						Name:         "test_edge",
						TargetSchema: "Target",
						Type:         "invalid-edge-type",
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema, err := ConvertDefToSchema(tt.def)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, schema)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, schema)
				assert.Equal(t, tt.def.Name, schema.Name())
				assert.Equal(t, len(tt.def.Fields), len(schema.Fields()))
			}
		})
	}
}

func TestParseFieldType(t *testing.T) {
	tests := []struct {
		name     string
		typeStr  string
		expected JFieldType
		wantErr  bool
	}{
		{"string type", "string", JString, false},
		{"int type", "int", JInt, false},
		{"int64 type", "int64", JInt64, false},
		{"float type", "float", JFloat, false},
		{"float64 type", "float64", JFloat64, false},
		{"bool type", "bool", JBool, false},
		{"time type", "time", JTime, false},
		{"object type", "object", JObject, false},
		{"array type", "array", JArray, false},
		{"binary type", "binary", JBinary, false},
		{"objectid type", "objectid", JObjectID, false},
		{"invalid type", "invalid", "", true},
		{"empty type", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseFieldType(tt.typeStr)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, "", string(result))
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestParseEdgeType(t *testing.T) {
	tests := []struct {
		name     string
		typeStr  string
		expected EdgeType
		wantErr  bool
	}{
		{"one-to-one", "one-to-one", EdgeOneToOne, false},
		{"one-to-many", "one-to-many", EdgeOneToMany, false},
		{"many-to-one", "many-to-one", EdgeManyToOne, false},
		{"many-to-many", "many-to-many", EdgeManyToMany, false},
		{"invalid type", "invalid", "", true},
		{"empty type", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseEdgeType(tt.typeStr)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, "", string(result))
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestHTTPSchemaSource(t *testing.T) {
	source := NewHTTPSchemaSource("https://api.example.com/schemas")

	assert.Equal(t, "https://api.example.com/schemas", source.URL)
	assert.Nil(t, source.Client)

	// Test Load method (should return error as it's not implemented)
	schema, err := source.Load("test")
	assert.Error(t, err)
	assert.Nil(t, schema)
	assert.Contains(t, err.Error(), "HTTPSchemaSource not implemented yet")
}

func TestDatabaseSchemaSource(t *testing.T) {
	mockConnection := "mock_connection"
	table := "schema_definitions"

	source := NewDatabaseSchemaSource(mockConnection, table)

	assert.Equal(t, mockConnection, source.Connection)
	assert.Equal(t, table, source.Table)

	// Test Load method (should return error as it's not implemented)
	schema, err := source.Load("test")
	assert.Error(t, err)
	assert.Nil(t, schema)
	assert.Contains(t, err.Error(), "DatabaseSchemaSource not implemented yet")
}

func TestFileSchemaSource_Load_ComplexSchema(t *testing.T) {
	content := `
name: ECommerce
fields:
  id:
    type: string
    required: true
    unique: true
  name:
    type: string
    required: true
  description:
    type: string
  price:
    type: float64
    required: true
  category:
    type: string
    required: true
  tags:
    type: array
  in_stock:
    type: bool
    defaultValue: true
  metadata:
    type: object
  created_at:
    type: time
  updated_at:
    type: time
refs:
  orders:
    targetSchema: Order
    isArray: true
  reviews:
    targetSchema: Review
    isArray: true
edges:
  - name: product_orders
    targetSchema: Order
    type: one-to-many
  - name: product_reviews
    targetSchema: Review
    type: one-to-many
`

	// Create temporary file
	tmpFile, err := os.CreateTemp("", "test_complex.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(content)
	require.NoError(t, err)
	tmpFile.Close()

	source := NewFileSchemaSource(tmpFile.Name())
	schema, err := source.Load("")

	assert.NoError(t, err)
	assert.NotNil(t, schema)
	assert.Equal(t, "ECommerce", schema.Name())
	assert.Equal(t, 10, len(schema.Fields()))
	assert.Equal(t, 2, len(schema.Refs()))
	assert.Equal(t, 2, len(schema.Edges()))

	// Test specific fields
	idField, exists := schema.Field("id")
	assert.True(t, exists)
	assert.True(t, idField.IsRequired())
	assert.True(t, idField.IsUnique())

	priceField, exists := schema.Field("price")
	assert.True(t, exists)
	assert.True(t, priceField.IsRequired())
	assert.Equal(t, JFloat64, priceField.Type())

	inStockField, exists := schema.Field("in_stock")
	assert.True(t, exists)
	assert.False(t, inStockField.IsRequired())
	assert.Equal(t, true, inStockField.DefaultValue())
}

func TestFileSchemaSource_Load_BulkWithComplexSchemas(t *testing.T) {
	content := `
- name: User
  fields:
    id:
      type: string
      required: true
      unique: true
    name:
      type: string
      required: true
    email:
      type: string
      required: true
      unique: true
    age:
      type: int
      defaultValue: 18
    profile:
      type: object
    preferences:
      type: array
    created_at:
      type: time
  refs:
    orders:
      targetSchema: Order
      isArray: true
  edges:
    - name: user_orders
      targetSchema: Order
      type: one-to-many

- name: Product
  fields:
    id:
      type: string
      required: true
      unique: true
    name:
      type: string
      required: true
    price:
      type: float64
      required: true
    category:
      type: string
      required: true
    in_stock:
      type: bool
      defaultValue: true
  refs:
    orders:
      targetSchema: Order
      isArray: true
  edges:
    - name: product_orders
      targetSchema: Order
      type: many-to-many

- name: Order
  fields:
    id:
      type: string
      required: true
      unique: true
    user_id:
      type: string
      required: true
    total:
      type: float64
      required: true
    status:
      type: string
      defaultValue: pending
    items:
      type: array
    created_at:
      type: time
  refs:
    user:
      targetSchema: User
      isArray: false
    products:
      targetSchema: Product
      isArray: true
  edges:
    - name: order_user
      targetSchema: User
      type: many-to-one
    - name: order_products
      targetSchema: Product
      type: many-to-many
`

	// Create temporary file
	tmpFile, err := os.CreateTemp("", "test_bulk_complex.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(content)
	require.NoError(t, err)
	tmpFile.Close()

	source := NewFileSchemaSource(tmpFile.Name())

	// Test loading each schema
	schemas := []string{"User", "Product", "Order"}
	for _, schemaName := range schemas {
		t.Run("load_"+schemaName, func(t *testing.T) {
			schema, err := source.Load(schemaName)
			assert.NoError(t, err)
			assert.NotNil(t, schema)
			assert.Equal(t, schemaName, schema.Name())
			assert.Greater(t, len(schema.Fields()), 0)
		})
	}

	// Test loading non-existent schema
	t.Run("load_non_existent", func(t *testing.T) {
		schema, err := source.Load("NonExistent")
		assert.Error(t, err)
		assert.Nil(t, schema)
		assert.Contains(t, err.Error(), "not found in bulk file")
	})
}

func TestFileSchemaSource_Load_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr bool
	}{
		{
			name:    "empty file",
			content: "",
			wantErr: true,
		},
		{
			name:    "whitespace only",
			content: "   \n\t  ",
			wantErr: true,
		},
		{
			name:    "null JSON",
			content: "null",
			wantErr: true,
		},
		{
			name:    "empty object",
			content: "{}",
			wantErr: true,
		},
		{
			name:    "empty array",
			content: "[]",
			wantErr: true,
		},
		{
			name: "schema with no fields",
			content: `
name: EmptySchema
fields: {}
`,
			wantErr: false,
		},
		{
			name: "schema with empty refs and edges",
			content: `
name: MinimalSchema
fields:
  id:
    type: string
    required: true
refs: {}
edges: []
`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary file
			tmpFile, err := os.CreateTemp("", "test_edge_case.yaml")
			require.NoError(t, err)
			defer os.Remove(tmpFile.Name())

			_, err = tmpFile.WriteString(tt.content)
			require.NoError(t, err)
			tmpFile.Close()

			source := NewFileSchemaSource(tmpFile.Name())
			schema, err := source.Load("")

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, schema)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, schema)
			}
		})
	}
}

func TestFileSchemaSource_Load_PathHandling(t *testing.T) {
	content := `
name: TestSchema
fields:
  id:
    type: string
    required: true
`

	// Test with absolute path
	tmpFile, err := os.CreateTemp("", "test_absolute.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(content)
	require.NoError(t, err)
	tmpFile.Close()

	absPath, err := filepath.Abs(tmpFile.Name())
	require.NoError(t, err)

	source := NewFileSchemaSource(absPath)
	schema, err := source.Load("")

	assert.NoError(t, err)
	assert.NotNil(t, schema)
	assert.Equal(t, "TestSchema", schema.Name())

	// Test with relative path
	relPath := filepath.Base(tmpFile.Name())
	source = NewFileSchemaSource(relPath)
	schema, err = source.Load("")

	// This might fail depending on current working directory
	// but we're testing that the path is handled correctly
	if err != nil {
		assert.Contains(t, err.Error(), "failed to read schema file")
		assert.Nil(t, schema)
	} else {
		assert.NotNil(t, schema)
		assert.Equal(t, "TestSchema", schema.Name())
	}
}
