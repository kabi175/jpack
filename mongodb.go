package jpack

import (
	"context"
	"errors"

	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
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
			m.record["id"] = objID.Hex() // Store the ID as a string in the record
		}
		// After inserting, we can set the original record to the current record
		m.originalRecord = m.record
		// and clear the record to indicate that it has been saved.
		m.record = bson.M{}

		return nil
	} else {
		convertToBSON, err := m.convertToBSON(ctx, m.record)
		delete(convertToBSON, "id")           // Remove the id field from the update
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
	pkID, ok := m.record["id"]
	if !ok {
		pkID, ok = m.originalRecord["id"]
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
