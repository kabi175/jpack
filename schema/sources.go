package schema

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kabi175/jpack/logger"
	"github.com/samber/lo"
	"gopkg.in/yaml.v3"
)

// SchemaDef represents a schema definition loaded from external sources
type SchemaDef struct {
	Name   string              `json:"name" yaml:"name"`
	Fields map[string]FieldDef `json:"fields" yaml:"fields"`
	Refs   map[string]RefDef   `json:"refs,omitempty" yaml:"refs,omitempty"`
	Edges  []EdgeDef           `json:"edges,omitempty" yaml:"edges,omitempty"`
}

// FieldDef represents a field definition
type FieldDef struct {
	Type         string      `json:"type" yaml:"type"`
	DefaultValue interface{} `json:"defaultValue,omitempty" yaml:"defaultValue,omitempty"`
	Required     bool        `json:"required,omitempty" yaml:"required,omitempty"`
	Unique       bool        `json:"unique,omitempty" yaml:"unique,omitempty"`
}

// RefDef represents a reference definition
type RefDef struct {
	TargetSchema string `json:"targetSchema" yaml:"targetSchema"`
	IsArray      bool   `json:"isArray,omitempty" yaml:"isArray,omitempty"`
}

// EdgeDef represents an edge definition
type EdgeDef struct {
	Name         string `json:"name" yaml:"name"`
	TargetSchema string `json:"targetSchema" yaml:"targetSchema"`
	Type         string `json:"type" yaml:"type"`
}

// FileSchemaSource loads schema definitions from JSON and YAML files
type FileSchemaSource struct {
	Path string
}

// NewFileSchemaSource creates a new file-based schema source
func NewFileSchemaSource(path string) *FileSchemaSource {
	return &FileSchemaSource{Path: path}
}

// Load loads a schema definition from a JSON or YAML file
func (f *FileSchemaSource) Load(schemaName string) (JSchema, error) {
	data, err := os.ReadFile(f.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to read schema file %s: %w", f.Path, err)
	}

	// First try to load as bulk schema ([]SchemaDef)
	var bulkDefs []SchemaDef
	bulkErr := f.tryUnmarshal(data, &bulkDefs)

	if bulkErr == nil && len(bulkDefs) > 0 {
		// Successfully loaded as bulk schema
		if schemaName != "" {
			// Find specific schema by name
			for _, def := range bulkDefs {
				if def.Name == schemaName {
					return ConvertDefToSchema(def)
				}
			}
			return nil, fmt.Errorf("schema '%s' not found in bulk file %s", schemaName, f.Path)
		}
		// Return the first schema if no specific name requested
		schemaDef, ok := lo.Find(bulkDefs, func(def SchemaDef) bool {
			return def.Name == schemaName
		})
		if !ok {
			logger.Schema.Error().
				Str("schema", schemaName).
				Str("file", f.Path).
				Msg("schema not found in bulk file")

			return nil, fmt.Errorf("schema '%s' not found in bulk file %s", schemaName, f.Path)
		}
		return ConvertDefToSchema(schemaDef)
	}

	// If bulk loading failed, try single schema (SchemaDef)
	var def SchemaDef
	singleErr := f.tryUnmarshal(data, &def)

	if singleErr != nil {
		return nil, fmt.Errorf("failed to unmarshal schema definition as bulk or single schema: bulk error: %v, single error: %w", bulkErr, singleErr)
	}

	// Override the name if provided
	if schemaName != "" {
		def.Name = schemaName
	}

	return ConvertDefToSchema(def)
}

// tryUnmarshal attempts to unmarshal data as JSON or YAML
func (f *FileSchemaSource) tryUnmarshal(data []byte, target interface{}) error {
	// Detect file format based on extension
	ext := strings.ToLower(filepath.Ext(f.Path))
	switch ext {
	case ".json":
		return json.Unmarshal(data, target)
	case ".yaml", ".yml":
		return yaml.Unmarshal(data, target)
	default:
		// Try JSON first, then YAML
		if err := json.Unmarshal(data, target); err != nil {
			return yaml.Unmarshal(data, target)
		}
		return nil
	}
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
