package mongo

import (
	"context"
	"fmt"

	"github.com/kabi175/jpack/converter"
	"github.com/kabi175/jpack/hooks"
	"github.com/kabi175/jpack/logger"
	"github.com/kabi175/jpack/repository"
	"github.com/kabi175/jpack/schema"
	"github.com/kabi175/jpack/validation"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoRepository implements repository.JpaRepository for MongoDB
type MongoRepository struct {
	schema     schema.JSchema
	collection *mongo.Collection
	converter  converter.ConverterRegistry
	validator  validation.ValidationEngine
	hookExec   hooks.HookExecutor
}

// NewMongoRepository creates a new MongoDB repository
func NewMongoRepository(schema schema.JSchema, collection *mongo.Collection) *MongoRepository {
	return &MongoRepository{
		schema:     schema,
		collection: collection,
		converter:  converter.GetGlobalConverterRegistry(),
		validator:  validation.GetGlobalValidationEngine(),
		hookExec:   hooks.GetGlobalHookExecutor(),
	}
}

// NewMongoRepositoryWithOptions creates a new MongoDB repository with custom options
func NewMongoRepositoryWithOptions(
	schema schema.JSchema,
	collection *mongo.Collection,
	converter converter.ConverterRegistry,
	validator validation.ValidationEngine,
	hookExec hooks.HookExecutor,
) *MongoRepository {
	return &MongoRepository{
		schema:     schema,
		collection: collection,
		converter:  converter,
		validator:  validator,
		hookExec:   hookExec,
	}
}

// Save saves a record to MongoDB
func (r *MongoRepository) Save(ctx context.Context, rec schema.JRecord) (schema.JRecord, error) {
	// Execute before save hooks
	if err := r.hookExec.ExecuteBeforeSave(ctx, rec); err != nil {
		logger.Mongo.Error().
			Err(err).
			Msg("before save hook failed")
		return nil, fmt.Errorf("before save hook failed: %w", err)
	}

	// Validate record
	validationCtx := validation.NewValidationContext(r.schema, rec, r)
	if err := r.validator.ValidateAll(ctx, validationCtx); err != nil {
		logger.Mongo.Error().
			Err(err).
			Msg("validation failed")
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Convert record to MongoDB document
	doc, err := r.recordToDocument(rec)
	if err != nil {
		logger.Mongo.Error().
			Err(err).
			Msg("failed to convert record to document")
		return nil, fmt.Errorf("failed to convert record to document: %w", err)
	}

	// Generate ID if not present
	if doc["_id"] == nil {
		doc["_id"] = primitive.NewObjectID()
	}

	// Insert document
	result, err := r.collection.InsertOne(ctx, doc)
	if err != nil {
		logger.Mongo.Error().
			Err(err).
			Msg("failed to insert document")
		return nil, fmt.Errorf("failed to insert document: %w", err)
	}

	// Update record with generated ID
	rec.Set("_id", result.InsertedID)

	// Execute after save hooks
	if err := r.hookExec.ExecuteAfterSave(ctx, rec); err != nil {
		logger.Mongo.Error().
			Err(err).
			Msg("after save hook failed")
		return nil, fmt.Errorf("after save hook failed: %w", err)
	}

	return rec, nil
}

// FindById finds a record by ID
func (r *MongoRepository) FindById(ctx context.Context, id any) (schema.JRecord, error) {
	// Convert ID to ObjectID
	objectID, err := r.convertToObjectID(id)
	if err != nil {
		logger.Mongo.Error().
			Err(err).
			Msg("invalid ID")
		return nil, fmt.Errorf("invalid ID: %w", err)
	}

	// Find document
	var doc bson.M
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			logger.Mongo.Debug().Msg("record not found")
			return nil, fmt.Errorf("record not found")
		}
		logger.Mongo.Error().
			Err(err).
			Msg("failed to find document")
		return nil, fmt.Errorf("failed to find document: %w", err)
	}

	// Convert document to record
	rec, err := r.documentToRecord(doc)
	if err != nil {
		return nil, fmt.Errorf("failed to convert document to record: %w", err)
	}

	// Execute after find hooks
	if err := r.hookExec.ExecuteAfterFind(ctx, rec); err != nil {
		return nil, fmt.Errorf("after find hook failed: %w", err)
	}

	return rec, nil
}

// Delete deletes a record by ID
func (r *MongoRepository) Delete(ctx context.Context, id any) error {
	// Convert ID to ObjectID
	objectID, err := r.convertToObjectID(id)
	if err != nil {
		return fmt.Errorf("invalid ID: %w", err)
	}

	// Find record first for hooks
	rec, err := r.FindById(ctx, id)
	if err != nil {
		return err
	}

	// Execute before delete hooks
	if err := r.hookExec.ExecuteBeforeDelete(ctx, rec); err != nil {
		return fmt.Errorf("before delete hook failed: %w", err)
	}

	// Delete document
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("record not found")
	}

	// Execute after delete hooks
	if err := r.hookExec.ExecuteAfterDelete(ctx, rec); err != nil {
		return fmt.Errorf("after delete hook failed: %w", err)
	}

	return nil
}

// FindAll finds all records
func (r *MongoRepository) FindAll(ctx context.Context) ([]schema.JRecord, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("failed to find documents: %w", err)
	}
	defer cursor.Close(ctx)

	var records []schema.JRecord
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("failed to decode document: %w", err)
		}

		rec, err := r.documentToRecord(doc)
		if err != nil {
			return nil, fmt.Errorf("failed to convert document to record: %w", err)
		}

		// Execute after find hooks
		if err := r.hookExec.ExecuteAfterFind(ctx, rec); err != nil {
			return nil, fmt.Errorf("after find hook failed: %w", err)
		}

		records = append(records, rec)
	}

	return records, nil
}

// Exists checks if a record exists by ID
func (r *MongoRepository) Exists(ctx context.Context, id any) (bool, error) {
	// Convert ID to ObjectID
	objectID, err := r.convertToObjectID(id)
	if err != nil {
		return false, fmt.Errorf("invalid ID: %w", err)
	}

	count, err := r.collection.CountDocuments(ctx, bson.M{"_id": objectID})
	if err != nil {
		return false, fmt.Errorf("failed to count documents: %w", err)
	}

	return count > 0, nil
}

// FindBy finds records by criteria
func (r *MongoRepository) FindBy(ctx context.Context, criteria repository.JCriteria) ([]schema.JRecord, error) {
	// Build MongoDB query
	queryBuilder := repository.NewQueryBuilder(criteria)
	filter, err := queryBuilder.BuildFilter()
	if err != nil {
		return nil, fmt.Errorf("failed to build filter: %w", err)
	}

	opts, err := queryBuilder.BuildOptions()
	if err != nil {
		return nil, fmt.Errorf("failed to build options: %w", err)
	}

	// Create find options
	findOptions := options.Find()
	if sort, ok := opts["sort"]; ok {
		findOptions.SetSort(sort)
	}
	if projection, ok := opts["projection"]; ok {
		findOptions.SetProjection(projection)
	}
	if limit, ok := opts["limit"]; ok {
		findOptions.SetLimit(limit.(int64))
	}
	if skip, ok := opts["skip"]; ok {
		findOptions.SetSkip(skip.(int64))
	}

	// Find documents
	cursor, err := r.collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to find documents: %w", err)
	}
	defer cursor.Close(ctx)

	var records []schema.JRecord
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("failed to decode document: %w", err)
		}

		rec, err := r.documentToRecord(doc)
		if err != nil {
			return nil, fmt.Errorf("failed to convert document to record: %w", err)
		}

		// Execute after find hooks
		if err := r.hookExec.ExecuteAfterFind(ctx, rec); err != nil {
			return nil, fmt.Errorf("after find hook failed: %w", err)
		}

		records = append(records, rec)
	}

	return records, nil
}

// Count counts records by criteria
func (r *MongoRepository) Count(ctx context.Context, criteria repository.JCriteria) (int64, error) {
	// Build MongoDB query
	queryBuilder := repository.NewQueryBuilder(criteria)
	filter, err := queryBuilder.BuildFilter()
	if err != nil {
		return 0, fmt.Errorf("failed to build filter: %w", err)
	}

	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("failed to count documents: %w", err)
	}

	return count, nil
}

// Update updates records by criteria
func (r *MongoRepository) Update(ctx context.Context, criteria repository.JCriteria, updates schema.JRecord) error {
	// Build MongoDB query
	queryBuilder := repository.NewQueryBuilder(criteria)
	filter, err := queryBuilder.BuildFilter()
	if err != nil {
		return fmt.Errorf("failed to build filter: %w", err)
	}

	// Convert updates to MongoDB document
	updateDoc, err := r.recordToDocument(updates)
	if err != nil {
		return fmt.Errorf("failed to convert updates to document: %w", err)
	}

	// Remove _id from updates if present
	delete(updateDoc, "_id")

	// Execute before update hooks for all matching records
	records, err := r.FindBy(ctx, criteria)
	if err != nil {
		return fmt.Errorf("failed to find records for update: %w", err)
	}

	for _, rec := range records {
		// Apply updates to record
		for key, value := range updateDoc {
			rec.Set(key, value)
		}

		// Execute before update hooks
		if err := r.hookExec.ExecuteBeforeUpdate(ctx, rec); err != nil {
			return fmt.Errorf("before update hook failed: %w", err)
		}
	}

	// Update documents
	result, err := r.collection.UpdateMany(ctx, filter, bson.M{"$set": updateDoc})
	if err != nil {
		return fmt.Errorf("failed to update documents: %w", err)
	}

	// Execute after update hooks
	for _, rec := range records {
		if err := r.hookExec.ExecuteAfterUpdate(ctx, rec); err != nil {
			return fmt.Errorf("after update hook failed: %w", err)
		}
	}

	if result.ModifiedCount == 0 {
		return fmt.Errorf("no records were updated")
	}

	return nil
}

// DeleteBy deletes records by criteria
func (r *MongoRepository) DeleteBy(ctx context.Context, criteria repository.JCriteria) error {
	// Build MongoDB query
	queryBuilder := repository.NewQueryBuilder(criteria)
	filter, err := queryBuilder.BuildFilter()
	if err != nil {
		return fmt.Errorf("failed to build filter: %w", err)
	}

	// Execute before delete hooks for all matching records
	records, err := r.FindBy(ctx, criteria)
	if err != nil {
		return fmt.Errorf("failed to find records for deletion: %w", err)
	}

	for _, rec := range records {
		// Execute before delete hooks
		if err := r.hookExec.ExecuteBeforeDelete(ctx, rec); err != nil {
			return fmt.Errorf("before delete hook failed: %w", err)
		}
	}

	// Delete documents
	result, err := r.collection.DeleteMany(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to delete documents: %w", err)
	}

	// Execute after delete hooks
	for _, rec := range records {
		if err := r.hookExec.ExecuteAfterDelete(ctx, rec); err != nil {
			return fmt.Errorf("after delete hook failed: %w", err)
		}
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("no records were deleted")
	}

	return nil
}

// FindOne finds one record by criteria
func (r *MongoRepository) FindOne(ctx context.Context, criteria repository.JCriteria) (schema.JRecord, error) {
	// Build MongoDB query
	queryBuilder := repository.NewQueryBuilder(criteria)
	filter, err := queryBuilder.BuildFilter()
	if err != nil {
		return nil, fmt.Errorf("failed to build filter: %w", err)
	}

	opts, err := queryBuilder.BuildOptions()
	if err != nil {
		return nil, fmt.Errorf("failed to build options: %w", err)
	}

	// Create find options
	findOptions := options.FindOne()
	if sort, ok := opts["sort"]; ok {
		findOptions.SetSort(sort)
	}
	if projection, ok := opts["projection"]; ok {
		findOptions.SetProjection(projection)
	}

	// Find document
	var doc bson.M
	err = r.collection.FindOne(ctx, filter, findOptions).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("record not found")
		}
		return nil, fmt.Errorf("failed to find document: %w", err)
	}

	// Convert document to record
	rec, err := r.documentToRecord(doc)
	if err != nil {
		return nil, fmt.Errorf("failed to convert document to record: %w", err)
	}

	// Execute after find hooks
	if err := r.hookExec.ExecuteAfterFind(ctx, rec); err != nil {
		return nil, fmt.Errorf("after find hook failed: %w", err)
	}

	return rec, nil
}

// Helper methods

// recordToDocument converts a JRecord to a MongoDB document
func (r *MongoRepository) recordToDocument(rec schema.JRecord) (bson.M, error) {
	doc := make(bson.M)

	for _, field := range r.schema.Fields() {
		value := rec.Get(field.Name())
		if value == nil {
			continue
		}

		// Convert value using appropriate converter
		converter, exists := r.converter.Get(field.Type())
		if !exists {
			// Use default conversion
			doc[field.Name()] = value
			continue
		}

		convertedValue, err := converter.ToDB(value)
		if err != nil {
			return nil, fmt.Errorf("failed to convert field '%s': %w", field.Name(), err)
		}

		doc[field.Name()] = convertedValue
	}

	// Add any additional fields from the record
	for _, key := range rec.Keys() {
		if _, exists := doc[key]; !exists {
			doc[key] = rec.Get(key)
		}
	}

	return doc, nil
}

// documentToRecord converts a MongoDB document to a JRecord
func (r *MongoRepository) documentToRecord(doc bson.M) (schema.JRecord, error) {
	rec := schema.NewJRecord()

	for key, value := range doc {
		// Skip MongoDB internal fields
		if key == "_id" {
			rec.Set(key, value)
			continue
		}

		// Find field in schema
		field, exists := r.schema.Field(key)
		if !exists {
			// Add as-is for fields not in schema
			rec.Set(key, value)
			continue
		}

		// Convert value using appropriate converter
		converter, exists := r.converter.Get(field.Type())
		if !exists {
			// Use default conversion
			rec.Set(key, value)
			continue
		}

		convertedValue, err := converter.FromDB(value)
		if err != nil {
			return nil, fmt.Errorf("failed to convert field '%s': %w", key, err)
		}

		rec.Set(key, convertedValue)
	}

	return rec, nil
}

// convertToObjectID converts various ID types to ObjectID
func (r *MongoRepository) convertToObjectID(id any) (primitive.ObjectID, error) {
	switch v := id.(type) {
	case primitive.ObjectID:
		return v, nil
	case string:
		return primitive.ObjectIDFromHex(v)
	case []byte:
		if len(v) == 12 {
			return primitive.ObjectID(v), nil
		}
		return primitive.NilObjectID, fmt.Errorf("invalid ObjectID bytes length: %d", len(v))
	default:
		return primitive.NilObjectID, fmt.Errorf("cannot convert %T to ObjectID", id)
	}
}
