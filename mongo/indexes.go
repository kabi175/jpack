package mongo

import (
	"context"
	"fmt"

	"github.com/kabi175/jpack/logger"
	"github.com/kabi175/jpack/schema"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// IndexType represents the type of index
type IndexType string

const (
	IndexAscending   IndexType = "asc"
	IndexDescending  IndexType = "desc"
	IndexText        IndexType = "text"
	IndexGeo2D       IndexType = "2d"
	IndexGeo2DSphere IndexType = "2dsphere"
	IndexHashed      IndexType = "hashed"
)

// IndexOption represents index options
type IndexOption struct {
	Unique                  bool
	Sparse                  bool
	Background              bool
	ExpireAfterSeconds      *int32
	PartialFilterExpression bson.M
}

// IndexDefinition represents an index definition
type IndexDefinition struct {
	Fields  []IndexField
	Options IndexOption
}

// IndexField represents a field in an index
type IndexField struct {
	Name string
	Type IndexType
}

// IndexManager manages MongoDB indexes based on schema
type IndexManager struct {
	collection *mongo.Collection
	schema     schema.JSchema
}

// NewIndexManager creates a new index manager
func NewIndexManager(collection *mongo.Collection, schema schema.JSchema) *IndexManager {
	return &IndexManager{
		collection: collection,
		schema:     schema,
	}
}

// CreateIndexes creates indexes based on schema definition
func (im *IndexManager) CreateIndexes(ctx context.Context) error {
	// Create indexes for unique fields
	for _, field := range im.schema.Fields() {
		if field.IsUnique() {
			indexDef := IndexDefinition{
				Fields: []IndexField{
					{Name: field.Name(), Type: IndexAscending},
				},
				Options: IndexOption{
					Unique: true,
					Sparse: true,
				},
			}

			if err := im.createIndex(ctx, indexDef); err != nil {
				logger.Mongo.Error().
					Str("field", field.Name()).
					Err(err).
					Msg("failed to create unique index for field")
				return fmt.Errorf("failed to create unique index for field '%s': %w", field.Name(), err)
			}
		}
	}

	// Create compound indexes for references
	for _, ref := range im.schema.Refs() {
		indexDef := IndexDefinition{
			Fields: []IndexField{
				{Name: ref.Name(), Type: IndexAscending},
			},
			Options: IndexOption{
				Sparse: true,
			},
		}

		if err := im.createIndex(ctx, indexDef); err != nil {
			logger.Mongo.Error().
				Str("field", ref.Name()).
				Err(err).
				Msg("failed to create reference index for field")
			return fmt.Errorf("failed to create reference index for field '%s': %w", ref.Name(), err)
		}
	}

	return nil
}

// CreateCustomIndex creates a custom index
func (im *IndexManager) CreateCustomIndex(ctx context.Context, indexDef IndexDefinition) error {
	return im.createIndex(ctx, indexDef)
}

// DropIndex drops an index by name
func (im *IndexManager) DropIndex(ctx context.Context, indexName string) error {
	_, err := im.collection.Indexes().DropOne(ctx, indexName)
	if err != nil {
		logger.Mongo.Error().
			Str("index", indexName).
			Err(err).
			Msg("failed to drop index")
		return fmt.Errorf("failed to drop index '%s': %w", indexName, err)
	}
	return nil
}

// ListIndexes lists all indexes
func (im *IndexManager) ListIndexes(ctx context.Context) ([]bson.M, error) {
	cursor, err := im.collection.Indexes().List(ctx)
	if err != nil {
		logger.Mongo.Error().
			Err(err).
			Msg("failed to list indexes")
		return nil, fmt.Errorf("failed to list indexes: %w", err)
	}
	defer cursor.Close(ctx)

	var indexes []bson.M
	for cursor.Next(ctx) {
		var index bson.M
		if err := cursor.Decode(&index); err != nil {
			logger.Mongo.Error().
				Err(err).
				Msg("failed to decode index")
			return nil, fmt.Errorf("failed to decode index: %w", err)
		}
		indexes = append(indexes, index)
	}

	return indexes, nil
}

// createIndex creates an index with the given definition
func (im *IndexManager) createIndex(ctx context.Context, indexDef IndexDefinition) error {
	// Build index model
	indexModel := mongo.IndexModel{}

	// Build keys document
	keys := bson.D{}
	for _, field := range indexDef.Fields {
		var direction int
		switch field.Type {
		case IndexAscending:
			direction = 1
		case IndexDescending:
			direction = -1
		case IndexText:
			direction = 1
		case IndexGeo2D:
			direction = 1
		case IndexGeo2DSphere:
			direction = 1
		case IndexHashed:
			direction = 1
		default:
			direction = 1
		}

		keys = append(keys, bson.E{Key: field.Name, Value: direction})
	}

	indexModel.Keys = keys

	// Build options
	opts := options.Index()

	if indexDef.Options.Unique {
		opts.SetUnique(true)
	}

	if indexDef.Options.Sparse {
		opts.SetSparse(true)
	}

	if indexDef.Options.Background {
		opts.SetBackground(true)
	}

	if indexDef.Options.ExpireAfterSeconds != nil {
		opts.SetExpireAfterSeconds(*indexDef.Options.ExpireAfterSeconds)
	}

	if indexDef.Options.PartialFilterExpression != nil {
		opts.SetPartialFilterExpression(indexDef.Options.PartialFilterExpression)
	}

	// Set text index options if any field is text
	for _, field := range indexDef.Fields {
		if field.Type == IndexText {
			opts.SetDefaultLanguage("english")
			break
		}
	}

	indexModel.Options = opts

	// Create index
	_, err := im.collection.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		logger.Mongo.Error().
			Err(err).
			Msg("failed to create index")
		return fmt.Errorf("failed to create index: %w", err)
	}

	return nil
}

// SchemaIndexBuilder helps build indexes from schema
type SchemaIndexBuilder struct {
	schema schema.JSchema
}

// NewSchemaIndexBuilder creates a new schema index builder
func NewSchemaIndexBuilder(schema schema.JSchema) *SchemaIndexBuilder {
	return &SchemaIndexBuilder{
		schema: schema,
	}
}

// BuildIndexes builds index definitions from schema
func (sib *SchemaIndexBuilder) BuildIndexes() []IndexDefinition {
	var indexes []IndexDefinition

	// Build unique indexes
	for _, field := range sib.schema.Fields() {
		if field.IsUnique() {
			indexes = append(indexes, IndexDefinition{
				Fields: []IndexField{
					{Name: field.Name(), Type: IndexAscending},
				},
				Options: IndexOption{
					Unique: true,
					Sparse: true,
				},
			})
		}
	}

	// Build reference indexes
	for _, ref := range sib.schema.Refs() {
		indexes = append(indexes, IndexDefinition{
			Fields: []IndexField{
				{Name: ref.Name(), Type: IndexAscending},
			},
			Options: IndexOption{
				Sparse: true,
			},
		})
	}

	return indexes
}

// BuildCompoundIndexes builds compound index definitions
func (sib *SchemaIndexBuilder) BuildCompoundIndexes() []IndexDefinition {
	var indexes []IndexDefinition

	// Example: Build compound index for common query patterns
	// This would be customized based on actual usage patterns

	// Build index for created_at and updated_at if they exist
	createdAtField, hasCreatedAt := sib.schema.Field("created_at")
	updatedAtField, hasUpdatedAt := sib.schema.Field("updated_at")

	if hasCreatedAt && hasUpdatedAt {
		indexes = append(indexes, IndexDefinition{
			Fields: []IndexField{
				{Name: createdAtField.Name(), Type: IndexDescending},
				{Name: updatedAtField.Name(), Type: IndexDescending},
			},
			Options: IndexOption{
				Background: true,
			},
		})
	}

	return indexes
}

// BuildTextIndexes builds text index definitions
func (sib *SchemaIndexBuilder) BuildTextIndexes() []IndexDefinition {
	var indexes []IndexDefinition

	// Build text indexes for string fields that might be searched
	for _, field := range sib.schema.Fields() {
		if field.Type() == schema.JString {
			// Check if field name suggests it's searchable
			fieldName := field.Name()
			if isSearchableField(fieldName) {
				indexes = append(indexes, IndexDefinition{
					Fields: []IndexField{
						{Name: fieldName, Type: IndexText},
					},
					Options: IndexOption{
						Background: true,
					},
				})
			}
		}
	}

	return indexes
}

// BuildTTLIndexes builds TTL index definitions
func (sib *SchemaIndexBuilder) BuildTTLIndexes() []IndexDefinition {
	var indexes []IndexDefinition

	// Build TTL indexes for time fields that might expire
	for _, field := range sib.schema.Fields() {
		if field.Type() == schema.JTime {
			fieldName := field.Name()
			if isTTLField(fieldName) {
				// Default TTL of 30 days
				expireAfterSeconds := int32(30 * 24 * 60 * 60)
				indexes = append(indexes, IndexDefinition{
					Fields: []IndexField{
						{Name: fieldName, Type: IndexAscending},
					},
					Options: IndexOption{
						ExpireAfterSeconds: &expireAfterSeconds,
						Background:         true,
					},
				})
			}
		}
	}

	return indexes
}

// Helper functions

// isSearchableField checks if a field name suggests it's searchable
func isSearchableField(fieldName string) bool {
	searchableFields := []string{
		"title", "name", "description", "content", "text", "message",
		"subject", "body", "summary", "keywords", "tags", "category",
	}

	for _, searchable := range searchableFields {
		if fieldName == searchable {
			return true
		}
	}

	return false
}

// isTTLField checks if a field name suggests it's a TTL field
func isTTLField(fieldName string) bool {
	ttlFields := []string{
		"expires_at", "expired_at", "ttl", "expires", "expired",
		"deleted_at", "archived_at", "purged_at",
	}

	for _, ttl := range ttlFields {
		if fieldName == ttl {
			return true
		}
	}

	return false
}

// IndexBuilder provides a fluent interface for building indexes
type IndexBuilder struct {
	fields  []IndexField
	options IndexOption
}

// NewIndexBuilder creates a new index builder
func NewIndexBuilder() *IndexBuilder {
	return &IndexBuilder{
		fields:  make([]IndexField, 0),
		options: IndexOption{},
	}
}

// AddField adds a field to the index
func (ib *IndexBuilder) AddField(name string, indexType IndexType) *IndexBuilder {
	ib.fields = append(ib.fields, IndexField{
		Name: name,
		Type: indexType,
	})
	return ib
}

// SetUnique sets the unique option
func (ib *IndexBuilder) SetUnique(unique bool) *IndexBuilder {
	ib.options.Unique = unique
	return ib
}

// SetSparse sets the sparse option
func (ib *IndexBuilder) SetSparse(sparse bool) *IndexBuilder {
	ib.options.Sparse = sparse
	return ib
}

// SetBackground sets the background option
func (ib *IndexBuilder) SetBackground(background bool) *IndexBuilder {
	ib.options.Background = background
	return ib
}

// SetExpireAfterSeconds sets the TTL option
func (ib *IndexBuilder) SetExpireAfterSeconds(seconds int32) *IndexBuilder {
	ib.options.ExpireAfterSeconds = &seconds
	return ib
}

// SetPartialFilterExpression sets the partial filter expression
func (ib *IndexBuilder) SetPartialFilterExpression(filter bson.M) *IndexBuilder {
	ib.options.PartialFilterExpression = filter
	return ib
}

// Build builds the index definition
func (ib *IndexBuilder) Build() IndexDefinition {
	return IndexDefinition{
		Fields:  ib.fields,
		Options: ib.options,
	}
}
