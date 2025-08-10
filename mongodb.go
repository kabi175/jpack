package jpack

import (
	"context"
	"errors"

	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type key string

var (
	Conn key = "jpack.conn.mongo"
)

const (
	// ConnKey is the key used to store the MongoDB connection in the context.
	defaultMongoPK = "_id"
)

func MustConn(ctx context.Context) *mongo.Database {
	conn, ok := ctx.Value(Conn).(*mongo.Database)
	if !ok || conn == nil {
		panic("jpack: mongo connection not found in context")
	}
	return conn
}

type mongoRecord struct {
	originalRecord map[string]any
	record         map[string]any

	schema JSchema
}

// DirtyKeys implements JRecord.
func (m *mongoRecord) DirtyKeys() []string {
	var dirtyKeys []string
	for key := range m.record {
		if _, exists := m.originalRecord[key]; !exists || m.record[key] != m.originalRecord[key] {
			dirtyKeys = append(dirtyKeys, key)
		}
	}
	return dirtyKeys
}

// Fields implements JRecord.
func (m *mongoRecord) Fields() []JField {
	var fields []JField
	for _, field := range m.Schema().Fields() {
		if _, ok := m.originalRecord[field.Name()]; ok {
			fields = append(fields, field)
		}
	}

	return fields
}

// IsModified implements JRecord.
func (m *mongoRecord) IsModified() bool {
	return len(m.DirtyKeys()) > 0
}

// IsNew implements JRecord.
func (m *mongoRecord) IsNew() bool {
	return len(m.originalRecord) == 0
}

// Save implements JRecord.
func (m *mongoRecord) Save(ctx context.Context) error {

	coll := MustConn(ctx).Collection(m.Schema().Name())
	pkField, _ := PK(m.schema)
	if m.IsNew() {
		convertToBSON, err := m.convertToBSON(ctx, m.record)
		if err != nil {
			log.Error().Err(err).Msg("jpack: failed to convert record to BSON")
			return err
		}
		res, err := coll.InsertOne(ctx, convertToBSON)
		if err != nil {
			return nil
		}

		// m.record[defaultMongoPK] = res.InsertedID
		objID, ok := res.InsertedID.(bson.ObjectID)
		if ok {
			m.record[pkField.Name()] = objID.Hex() // Store the ID as a string in the record
		}
		// After inserting, we can set the original record to the current record
		m.originalRecord = m.record
		// and clear the record to indicate that it has been saved.
		m.record = bson.M{}

		return nil
	} else {
		convertToBSON, err := m.convertToBSON(ctx, m.record)
		delete(convertToBSON, pkField.Name()) // Remove the id field from the update
		delete(convertToBSON, defaultMongoPK) // Remove the mongo id field from the update
		if err != nil {
			log.Error().Err(err).Msg("jpack: failed to convert record to BSON")
			return err
		}

		objID, err := m.objectID()
		if err != nil {
			return err
		}

		update := bson.M{"$set": convertToBSON}
		_, err = coll.UpdateByID(ctx, objID, update)

		if err != nil {
			return err
		}

		return nil
	}

}

func (m *mongoRecord) objectID() (bson.ObjectID, error) {
	pkField, _ := PK(m.schema)
	pkID, ok := m.record[pkField.Name()]
	if !ok {
		pkID, ok = m.originalRecord[pkField.Name()]
		if !ok {
			return bson.ObjectID{}, errors.New("record id can't be empty")
		}
	}

	pkStr, ok := pkID.(string)
	if !ok {
		return bson.ObjectID{}, errors.New("record id must be a string")
	}

	objID, err := bson.ObjectIDFromHex(pkStr)
	if err != nil {
		return bson.ObjectID{}, errors.Join(errors.New("failed to convert record id to ObjectID"), err)
	}

	return objID, nil

}

// Schema implements JRecord.
func (m *mongoRecord) Schema() JSchema {
	return m.schema
}

// SetValue implements JRecord.
func (m *mongoRecord) SetValue(field JField, value any) error {
	if m.record == nil {
		m.record = bson.M{}
	}

	if field == nil {
		return errors.New("field cannot be nil")
	}

	if field.Schema().Name() != m.Schema().Name() {
		return errors.New("field schema does not match record schema")
	}

	err := field.Type().Validate(value)
	if err != nil {
		return err
	}

	m.record[field.Name()] = value
	return nil
}

// Validate implements JRecord.
func (m *mongoRecord) Validate() error {
	return m.schema.Validate(m)
}

// Value implements JRecord.
func (m *mongoRecord) Value(field JField) (any, bool) {

	val, ok := m.record[field.Name()]
	if ok {
		return val, true
	}

	// If the value is not found in the record, check the original record
	val, ok = m.originalRecord[field.Name()]
	if ok {
		return val, true
	}

	// If the value is not found in either record, return nil and false
	return nil, false
}

func (m *mongoRecord) convertToBSON(ctx context.Context, record map[string]any) (bson.M, error) {
	bsonRecord := bson.M{}
	for _, field := range m.Schema().Fields() {
		val, ok := record[field.Name()]
		if ok {
			err := field.Type().SetValue(ctx, field, val, bsonRecord)
			if err != nil {
				log.Error().Err(err).Str("field", field.Name()).Msg("failed to set value in BSON record")
				return nil, err
			}

		}
	}
	return bsonRecord, nil
}

var _ JRecord = &mongoRecord{}

func NewMongoRecord(schema JSchema) *mongoRecord {
	return &mongoRecord{
		schema:         schema,
		originalRecord: make(map[string]any),
		record:         make(map[string]any),
	}
}

// mongoQuery implements the Query interface for MongoDB
type mongoQuery struct {
	schema     JSchema
	ctx        context.Context
	collection *mongo.Collection

	// Query building fields
	projection bson.M
	where      []bson.M
	orderBy    bson.D
	limit      *int64
	offset     *int64
	withRefs   map[string]func(JSchema, Query) Query
}

// NewMongoQuery creates a new MongoDB query for the given schema
func NewMongoQuery(ctx context.Context, schema JSchema) Query {
	db := MustConn(ctx)
	collection := db.Collection(schema.Name())

	return &mongoQuery{
		schema:     schema,
		ctx:        ctx,
		collection: collection,
		projection: bson.M{},
		where:      []bson.M{},
		orderBy:    bson.D{},
		withRefs:   make(map[string]func(JSchema, Query) Query),
	}
}

// Schema implements Query
func (q *mongoQuery) Schema() JSchema {
	return q.schema
}

// Select implements Query
func (q *mongoQuery) Select(fields ...JField) Query {
	projection := bson.M{}

	// Always include _id for MongoDB
	projection["_id"] = 1

	for _, field := range fields {
		if field.Schema().Name() == q.schema.Name() {
			projection[field.Name()] = 1
		}
	}

	q.projection = projection
	return q
}

// With implements Query for eager loading
func (q *mongoQuery) With(ref JRef, fn func(JSchema, Query) Query) Query {
	q.withRefs[ref.Name()] = fn
	return q
}

// Where implements Query
func (q *mongoQuery) Where(filter Filter) Query {
	// Convert the filter to MongoDB BSON format using the resolver
	mongoFilter := ResolveFilter(filter)
	if mongoFilter != nil {
		q.where = append(q.where, mongoFilter)
	}
	return q
}

// OrderBy implements Query
func (q *mongoQuery) OrderBy(fields ...JField) Query {
	orderBy := bson.D{}

	for _, field := range fields {
		if field.Schema().Name() == q.schema.Name() {
			// Default to ascending order
			orderBy = append(orderBy, bson.E{Key: field.Name(), Value: 1})
		}
	}

	q.orderBy = orderBy
	return q
}

// Limit implements Query
func (q *mongoQuery) Limit(limit int) Query {
	limit64 := int64(limit)
	q.limit = &limit64
	return q
}

// Offset implements Query
func (q *mongoQuery) Offset(offset int) Query {
	offset64 := int64(offset)
	q.offset = &offset64
	return q
}

// Execute implements Query
func (q *mongoQuery) Execute() ([]JRecord, error) {
	// Build the filter
	filter := bson.M{}
	if len(q.where) > 0 {
		filter = bson.M{"$and": q.where}
	}

	// Build options
	opts := options.Find()

	if len(q.projection) > 0 {
		opts.SetProjection(q.projection)
	}

	if len(q.orderBy) > 0 {
		opts.SetSort(q.orderBy)
	}

	if q.limit != nil {
		opts.SetLimit(*q.limit)
	}

	if q.offset != nil {
		opts.SetSkip(*q.offset)
	}

	// Execute the query
	cursor, err := q.collection.Find(q.ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(q.ctx)

	var records []JRecord

	for cursor.Next(q.ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}

		// Convert BSON document to mongoRecord
		record := NewMongoRecord(q.schema)

		// Convert ObjectID to string for the id field
		if id, ok := doc["_id"].(bson.ObjectID); ok {
			pkField, _ := PK(q.Schema())
			record.originalRecord[pkField.Name()] = id.Hex()
		}

		// Convert other fields
		for key, value := range doc {
			if key != "_id" {
				record.originalRecord[key] = value
			}
		}

		records = append(records, record)
	}

	// Handle eager loading
	if len(q.withRefs) > 0 {
		if err := q.loadReferences(records); err != nil {
			return nil, err
		}
	}

	return records, nil
}

// First implements Query
func (q *mongoQuery) First() (JRecord, error) {
	// Build the filter
	filter := bson.M{}
	if len(q.where) > 0 {
		filter = bson.M{"$and": q.where}
	}

	// Build options
	opts := options.FindOne()

	if len(q.projection) > 0 {
		opts.SetProjection(q.projection)
	}

	if len(q.orderBy) > 0 {
		opts.SetSort(q.orderBy)
	}

	if q.offset != nil {
		opts.SetSkip(*q.offset)
	}

	// Execute the query
	var doc bson.M
	err := q.collection.FindOne(q.ctx, filter, opts).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	// Convert BSON document to mongoRecord
	record := NewMongoRecord(q.schema)

	// Convert ObjectID to string for the id field
	if id, ok := doc["_id"].(bson.ObjectID); ok {
		pkField, _ := PK(q.Schema())
		record.originalRecord[pkField.Name()] = id.Hex()
	}

	// Convert other fields
	for key, value := range doc {
		if key != "_id" {
			record.originalRecord[key] = value
		}
	}

	// Handle eager loading
	if len(q.withRefs) > 0 {
		if err := q.loadReferences([]JRecord{record}); err != nil {
			return nil, err
		}
	}

	return record, nil
}

// Count implements Query
func (q *mongoQuery) Count() (int, error) {
	// Build the filter
	filter := bson.M{}
	if len(q.where) > 0 {
		filter = bson.M{"$and": q.where}
	}

	// Execute the count query
	count, err := q.collection.CountDocuments(q.ctx, filter)
	if err != nil {
		return 0, err
	}

	return int(count), nil
}

// loadReferences handles eager loading of referenced records
func (q *mongoQuery) loadReferences(records []JRecord) error {
	for refName, refFn := range q.withRefs {
		// Find the reference field
		refField, ok := q.schema.Field(refName)
		if !ok {
			continue
		}

		ref, ok := refField.(JRef)
		if !ok {
			continue
		}

		// Create a query for the referenced schema
		refQuery := NewMongoQuery(q.ctx, ref.RelSchema())

		// Apply the custom function to the reference query
		refQuery = refFn(ref.RelSchema(), refQuery)

		// Execute the reference query
		refRecords, err := refQuery.Execute()
		if err != nil {
			return err
		}

		// Create a map of reference records by ID for quick lookup
		refMap := make(map[string]JRecord)
		for _, refRecord := range refRecords {
			if id, ok := refRecord.Value(refField); ok {
				if idStr, ok := id.(string); ok {
					refMap[idStr] = refRecord
				}
			}
		}

		// Attach reference records to the main records
		for _, record := range records {
			if refID, ok := record.Value(refField); ok {
				if refIDStr, ok := refID.(string); ok {
					if refRecord, exists := refMap[refIDStr]; exists {
						// Set the reference record in the main record
						record.SetValue(refField, refRecord)
					}
				}
			}
		}
	}

	return nil
}

var _ Query = &mongoQuery{}
