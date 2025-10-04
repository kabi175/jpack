package schema

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

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
	name string
	Path string

	schemas map[string]JSchema
	mutex   sync.RWMutex
	loaded  bool
}

var _ ExternalSchemaSource = &FileSchemaSource{}

// NewFileSchemaSource creates a new file-based schema source
func NewFileSchemaSource(name string, path string) ExternalSchemaSource {
	return &FileSchemaSource{
		name:    name,
		Path:    path,
		schemas: make(map[string]JSchema),
		loaded:  false,
		mutex:   sync.RWMutex{},
	}
}

// Has implements ExternalSchemaSource.
func (f *FileSchemaSource) Has(schemaName string) (bool, error) {
	f.mutex.RLock()
	_, exists := f.schemas[schemaName]
	f.mutex.RUnlock()
	return exists, nil
}

// Name implements ExternalSchemaSource.
func (f *FileSchemaSource) Name() string {
	return f.name
}

// ReplaceSchema implements ExternalSchemaSource.
func (f *FileSchemaSource) ReplaceSchema(schemaName string, newSchema JSchema) error {
	panic("unimplemented")
}

// UnRegisterSchema implements ExternalSchemaSource.
func (f *FileSchemaSource) UnRegisterSchema(schemaName string) error {
	panic("unimplemented")
}

// Load loads a schema definition from a JSON or YAML file
func (f *FileSchemaSource) Load(schemaName string) (JSchema, error) {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	if !f.loaded {
		if err := f.loadAll(); err != nil {
			return nil, err
		}
	}

	schema, exists := f.schemas[schemaName]
	if !exists {
		return nil, fmt.Errorf("schema '%s' not found in file '%s'", schemaName, f.Path)
	}

	return schema, nil
}

func (f *FileSchemaSource) loadAll() error {
	data, err := os.ReadFile(f.Path)
	if err != nil {
		return fmt.Errorf("failed to read schema file %s: %w", f.Path, err)
	}

	var singleDef SchemaDef
	singleErr := f.tryUnmarshal(data, &singleDef)
	if singleErr == nil {
		schema, err := ConvertDefToSchema(singleDef)
		if err != nil {
			return fmt.Errorf("failed to convert schema definition to schema: %w", err)
		}
		f.schemas[singleDef.Name] = schema
		return nil
	}
	log.Printf("failed to unmarshal schema definition as single schema: %v", singleErr)

	var bulkDefs []SchemaDef
	bulkErr := f.tryUnmarshal(data, &bulkDefs)
	if bulkErr != nil {
		return fmt.Errorf("failed to unmarshal schema definition as bulk or single schema: bulk error: %v", bulkErr)
	}

	for _, def := range bulkDefs {
		schema, err := ConvertDefToSchema(def)
		if err != nil {
			return fmt.Errorf("failed to convert schema definition to schema: %w", err)
		}
		f.schemas[def.Name] = schema
	}
	return nil
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

	// Use SchemaBuilder to create the schema
	builder := NewSchemaBuilder(def.Name)

	// Add fields
	for fieldName, fieldDef := range def.Fields {
		fieldType, err := parseFieldType(fieldDef.Type)
		if err != nil {
			return nil, fmt.Errorf("invalid field type '%s' for field '%s': %w", fieldDef.Type, fieldName, err)
		}

		// Create field with properties
		if fieldDef.Required && fieldDef.Unique {
			builder.AddRequiredUniqueField(fieldName, fieldType, fieldDef.DefaultValue)
		} else if fieldDef.Required {
			builder.AddRequiredField(fieldName, fieldType, fieldDef.DefaultValue)
		} else if fieldDef.Unique {
			builder.AddUniqueField(fieldName, fieldType, fieldDef.DefaultValue)
		} else {
			builder.AddField(fieldName, fieldType, fieldDef.DefaultValue)
		}
	}

	// Add references (note: target schemas must be registered separately)
	for refName, refDef := range def.Refs {
		// For now, we'll create a placeholder schema for the reference
		// In a real implementation, you might want to resolve this differently
		targetSchema := NewSchemaBuilder(refDef.TargetSchema).Build()
		if refDef.IsArray {
			builder.AddArrayRef(refName, targetSchema)
		} else {
			builder.AddRef(refName, targetSchema)
		}
	}

	// Add edges
	for _, edgeDef := range def.Edges {
		targetSchema := NewSchemaBuilder(edgeDef.TargetSchema).Build()
		edgeType, err := parseEdgeType(edgeDef.Type)
		if err != nil {
			return nil, fmt.Errorf("invalid edge type '%s' for edge '%s': %w", edgeDef.Type, edgeDef.Name, err)
		}

		// Create a temporary schema for the edge (this is a limitation of the current design)
		tempSchema := builder.Build()
		edge := NewJEdge(edgeDef.Name, tempSchema, targetSchema, edgeType)
		builder.AddEdge(edge)
	}

	return builder.Build(), nil
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
