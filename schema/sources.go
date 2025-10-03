package schema

import (
	"encoding/json"
	"fmt"
	"os"
)

// SchemaDef represents a schema definition loaded from external sources
type SchemaDef struct {
	Name   string              `json:"name"`
	Fields map[string]FieldDef `json:"fields"`
	Refs   map[string]RefDef   `json:"refs,omitempty"`
	Edges  []EdgeDef           `json:"edges,omitempty"`
}

// FieldDef represents a field definition
type FieldDef struct {
	Type         string      `json:"type"`
	DefaultValue interface{} `json:"defaultValue,omitempty"`
	Required     bool        `json:"required,omitempty"`
	Unique       bool        `json:"unique,omitempty"`
}

// RefDef represents a reference definition
type RefDef struct {
	TargetSchema string `json:"targetSchema"`
	IsArray      bool   `json:"isArray,omitempty"`
}

// EdgeDef represents an edge definition
type EdgeDef struct {
	Name         string `json:"name"`
	TargetSchema string `json:"targetSchema"`
	Type         string `json:"type"`
}

// FileSchemaSource loads schema definitions from JSON files
type FileSchemaSource struct {
	Path string
}

// NewFileSchemaSource creates a new file-based schema source
func NewFileSchemaSource(path string) *FileSchemaSource {
	return &FileSchemaSource{Path: path}
}

// Load loads a schema definition from a JSON file
func (f *FileSchemaSource) Load(schemaName string) (JSchema, error) {
	data, err := os.ReadFile(f.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to read schema file %s: %w", f.Path, err)
	}

	var def SchemaDef
	if err := json.Unmarshal(data, &def); err != nil {
		return nil, fmt.Errorf("failed to unmarshal schema definition: %w", err)
	}

	// Override the name if provided
	if schemaName != "" {
		def.Name = schemaName
	}

	return ConvertDefToSchema(def)
}

// ConvertDefToSchema converts a SchemaDef to a JSchema
func ConvertDefToSchema(def SchemaDef) (JSchema, error) {
	if def.Name == "" {
		return nil, fmt.Errorf("schema name cannot be empty")
	}

	schema := NewJSchema(def.Name)

	// Add fields
	for fieldName, fieldDef := range def.Fields {
		fieldType, err := parseFieldType(fieldDef.Type)
		if err != nil {
			return nil, fmt.Errorf("invalid field type '%s' for field '%s': %w", fieldDef.Type, fieldName, err)
		}

		field := schema.AddField(fieldName, fieldType, fieldDef.DefaultValue)
		if fieldDef.Required {
			field.SetRequired(true)
		}
		if fieldDef.Unique {
			field.SetUnique(true)
		}
	}

	// Add references (note: target schemas must be registered separately)
	for refName, refDef := range def.Refs {
		// For now, we'll create a placeholder schema for the reference
		// In a real implementation, you might want to resolve this differently
		targetSchema := NewJSchema(refDef.TargetSchema)
		schema.AddRef(refName, targetSchema)
		if refDef.IsArray {
			// Note: JRef interface doesn't have SetIsArray method in current implementation
			// This would need to be added to support array references
		}
	}

	// Add edges
	for _, edgeDef := range def.Edges {
		targetSchema := NewJSchema(edgeDef.TargetSchema)
		edgeType, err := parseEdgeType(edgeDef.Type)
		if err != nil {
			return nil, fmt.Errorf("invalid edge type '%s' for edge '%s': %w", edgeDef.Type, edgeDef.Name, err)
		}

		edge := NewJEdge(edgeDef.Name, schema, targetSchema, edgeType)
		schema.AddEdge(edge)
	}

	return schema, nil
}

// parseFieldType converts string to JFieldType
func parseFieldType(typeStr string) (JFieldType, error) {
	switch typeStr {
	case "string":
		return JString, nil
	case "int":
		return JInt, nil
	case "int64":
		return JInt64, nil
	case "float":
		return JFloat, nil
	case "float64":
		return JFloat64, nil
	case "bool":
		return JBool, nil
	case "time":
		return JTime, nil
	case "object":
		return JObject, nil
	case "array":
		return JArray, nil
	case "binary":
		return JBinary, nil
	case "objectid":
		return JObjectID, nil
	default:
		return "", fmt.Errorf("unknown field type: %s", typeStr)
	}
}

// parseEdgeType converts string to EdgeType
func parseEdgeType(typeStr string) (EdgeType, error) {
	switch typeStr {
	case "one-to-one":
		return EdgeOneToOne, nil
	case "one-to-many":
		return EdgeOneToMany, nil
	case "many-to-one":
		return EdgeManyToOne, nil
	case "many-to-many":
		return EdgeManyToMany, nil
	default:
		return "", fmt.Errorf("unknown edge type: %s", typeStr)
	}
}

// HTTPSchemaSource loads schema definitions from HTTP endpoints
type HTTPSchemaSource struct {
	URL    string
	Client interface{} // HTTP client interface - can be customized
}

// NewHTTPSchemaSource creates a new HTTP-based schema source
func NewHTTPSchemaSource(url string) *HTTPSchemaSource {
	return &HTTPSchemaSource{URL: url}
}

// Load loads a schema definition from an HTTP endpoint
func (h *HTTPSchemaSource) Load(schemaName string) (JSchema, error) {
	// This is a placeholder implementation
	// In a real implementation, you would make an HTTP request to h.URL
	// and parse the response as JSON, similar to FileSchemaSource
	return nil, fmt.Errorf("HTTPSchemaSource not implemented yet")
}

// DatabaseSchemaSource loads schema definitions from a database
type DatabaseSchemaSource struct {
	Connection interface{} // Database connection interface
	Table      string      // Table containing schema definitions
}

// NewDatabaseSchemaSource creates a new database-based schema source
func NewDatabaseSchemaSource(connection interface{}, table string) *DatabaseSchemaSource {
	return &DatabaseSchemaSource{
		Connection: connection,
		Table:      table,
	}
}

// Load loads a schema definition from a database
func (d *DatabaseSchemaSource) Load(schemaName string) (JSchema, error) {
	// This is a placeholder implementation
	// In a real implementation, you would query the database
	// and parse the result as JSON, similar to FileSchemaSource
	return nil, fmt.Errorf("DatabaseSchemaSource not implemented yet")
}
