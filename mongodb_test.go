package jpack

import (
	"context"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var userSchema JSchema

func init() {
	userSchema = NewSchema("test_user").
		Field("id", &String{}).
		Field("first_name", &String{}).
		Field("last_name", &String{}).
		Field("email", &String{}).
		Field("age", &Number{}).
		Build()
}

func Test_mongoRecord_DirtyKeys(t *testing.T) {

	t.Run("Check Dirty keys", func(t *testing.T) {
		m := NewMongoRecord(userSchema)

		f, _ := userSchema.Field("first_name")
		m.SetValue(f, "John")

		want := []string{"first_name"}

		if got := m.DirtyKeys(); !reflect.DeepEqual(got, want) {
			t.Errorf("mongoRecord.DirtyKeys() = %v, want %v", got, want)
		}
	})

}

func mustField(t *testing.T, schema JSchema, name string) JField {
	t.Helper()
	field, ok := schema.Field(name)
	assert.True(t, ok, "Field %s should exist in schema", name)
	return field
}

func Test_mongoRecord_Save(t *testing.T) {
	uri := "mongodb://localhost:27017"
	client, err := mongo.Connect(options.Client().
		ApplyURI(uri))

	assert.NoError(t, err, "Failed to connect to MongoDB")
	defer func() {
		err := client.Disconnect(context.TODO())
		assert.NoError(t, err, "Failed to connect to MongoDB")
	}()

	client.Database("jpack_test").Drop(context.TODO())
	ctx := context.WithValue(context.Background(), Conn, client.Database("jpack_test"))
	m := NewMongoRecord(userSchema)
	t.Run("Create Record", func(t *testing.T) {
		m.SetValue(mustField(t, userSchema, "first_name"), "Jhon")
		m.SetValue(mustField(t, userSchema, "last_name"), "Doe")
		m.SetValue(mustField(t, userSchema, "email"), "Jhon@example.com")
		m.SetValue(mustField(t, userSchema, "age"), "23")
		err = m.Save(ctx)
		assert.NoError(t, err, "Failed to save record to MongoDB")

		pkId, ok := m.Value(mustField(t, userSchema, "id"))
		assert.True(t, ok, "Primary key should be set after saving a new record")
		assert.NotEmpty(t, pkId, "Primary key should not be empty after saving a new record")
		assert.IsType(t, "", pkId, "Primary key should be a string")
	})

	t.Run("Update Record", func(t *testing.T) {
		m.SetValue(mustField(t, userSchema, "email"), "Jhon@gmail.com")
		err = m.Save(ctx)
		assert.NoError(t, err, "Failed to update record in MongoDB")
	})

	t.Run("Save record with ref", func(t *testing.T) {
		postSchema := NewSchema("test_post").
			Field("id", &String{}).
			Field("title", &String{}).
			Ref("author", userSchema).
			Build()

		postRecord := NewMongoRecord(postSchema)
		postRecord.SetValue(mustField(t, postSchema, "title"), "My First Post")
		postRecord.SetValue(mustField(t, postSchema, "author"), m)

		err = postRecord.Save(ctx)
		assert.NoError(t, err, "Failed to save post record with ref to user")
	})

}

func TestMongoQuery(t *testing.T) {
	uri := "mongodb://localhost:27017"
	client, err := mongo.Connect(options.Client().
		ApplyURI(uri))

	assert.NoError(t, err, "Failed to connect to MongoDB")
	defer func() {
		err := client.Disconnect(context.TODO())
		assert.NoError(t, err, "Failed to disconnect from MongoDB")
	}()

	client.Database("jpack_test").Drop(context.TODO())
	ctx := context.WithValue(context.Background(), Conn, client.Database("jpack_test"))

	// Create some test data
	t.Run("Setup Test Data", func(t *testing.T) {
		// Insert a test user
		userRecord := NewMongoRecord(userSchema)
		userRecord.SetValue(mustField(t, userSchema, "first_name"), "John")
		userRecord.SetValue(mustField(t, userSchema, "last_name"), "Doe")
		userRecord.SetValue(mustField(t, userSchema, "email"), "john@example.com")
		userRecord.SetValue(mustField(t, userSchema, "age"), "30")
		err = userRecord.Save(ctx)
		assert.NoError(t, err, "Failed to save test user")

		// Insert another test user
		userRecord2 := NewMongoRecord(userSchema)
		userRecord2.SetValue(mustField(t, userSchema, "first_name"), "Jane")
		userRecord2.SetValue(mustField(t, userSchema, "last_name"), "Smith")
		userRecord2.SetValue(mustField(t, userSchema, "email"), "jane@example.com")
		userRecord2.SetValue(mustField(t, userSchema, "age"), "25")
		err = userRecord2.Save(ctx)
		assert.NoError(t, err, "Failed to save second test user")
	})

	t.Run("Test Query Execute", func(t *testing.T) {
		query := NewMongoQuery(ctx, userSchema)
		records, err := query.Execute()
		assert.NoError(t, err, "Failed to execute query")
		assert.Len(t, records, 2, "Should return 2 records")
	})

	t.Run("Test Query First", func(t *testing.T) {
		query := NewMongoQuery(ctx, userSchema)
		record, err := query.First()
		assert.NoError(t, err, "Failed to execute First query")
		assert.NotNil(t, record, "Should return a record")

		// Check that the record has the expected fields
		firstName, ok := record.Value(mustField(t, userSchema, "first_name"))
		assert.True(t, ok, "Should have first_name field")
		assert.NotEmpty(t, firstName, "First name should not be empty")
	})

	t.Run("Test Query Count", func(t *testing.T) {
		query := NewMongoQuery(ctx, userSchema)
		count, err := query.Count()
		assert.NoError(t, err, "Failed to execute Count query")
		assert.Equal(t, 2, count, "Should count 2 records")
	})

	t.Run("Test Query with Limit", func(t *testing.T) {
		query := NewMongoQuery(ctx, userSchema)
		query = query.Limit(1)
		records, err := query.Execute()
		assert.NoError(t, err, "Failed to execute query with limit")
		assert.Len(t, records, 1, "Should return 1 record due to limit")
	})

	t.Run("Test Query with Select", func(t *testing.T) {
		query := NewMongoQuery(ctx, userSchema)
		query = query.Select(mustField(t, userSchema, "first_name"), mustField(t, userSchema, "email"))
		records, err := query.Execute()
		assert.NoError(t, err, "Failed to execute query with select")
		assert.Len(t, records, 2, "Should return 2 records")

		// Check that the first record has the selected fields
		firstName, ok := records[0].Value(mustField(t, userSchema, "first_name"))
		assert.True(t, ok, "Should have first_name field")
		assert.NotEmpty(t, firstName, "First name should not be empty")

		email, ok := records[0].Value(mustField(t, userSchema, "email"))
		assert.True(t, ok, "Should have email field")
		assert.NotEmpty(t, email, "Email should not be empty")
	})

	t.Run("Test Query with OrderBy", func(t *testing.T) {
		query := NewMongoQuery(ctx, userSchema)
		query = query.OrderBy(mustField(t, userSchema, "first_name"))
		records, err := query.Execute()
		assert.NoError(t, err, "Failed to execute query with orderBy")
		assert.Len(t, records, 2, "Should return 2 records")

		// Check that records are ordered by first_name
		firstName1, _ := records[0].Value(mustField(t, userSchema, "first_name"))
		firstName2, _ := records[1].Value(mustField(t, userSchema, "first_name"))
		assert.LessOrEqual(t, firstName1, firstName2, "Records should be ordered by first_name")
	})

	t.Run("Test Query with Filter", func(t *testing.T) {
		query := NewMongoQuery(ctx, userSchema)
		// Create a filter using the new comparator system
		filter := Eq(mustField(t, userSchema, "first_name"), "John")
		query = query.Where(filter)
		records, err := query.Execute()
		assert.NoError(t, err, "Failed to execute query with filter")
		assert.Len(t, records, 1, "Should return 1 record with first_name = 'John'")

		// Check that the record has the expected first name
		firstName, ok := records[0].Value(mustField(t, userSchema, "first_name"))
		assert.True(t, ok, "Should have first_name field")
		assert.Equal(t, "John", firstName, "First name should be 'John'")
	})

	t.Run("Test Query with Complex Filter", func(t *testing.T) {
		query := NewMongoQuery(ctx, userSchema)
		// Create a complex filter: (first_name = 'John' OR first_name = 'Jane') AND age >= 25
		complexFilter := And(
			Or(
				Eq(mustField(t, userSchema, "first_name"), "John"),
				Eq(mustField(t, userSchema, "first_name"), "Jane"),
			),
			Gte(mustField(t, userSchema, "age"), 25),
		)
		query = query.Where(complexFilter)
		records, err := query.Execute()
		assert.NoError(t, err, "Failed to execute query with complex filter")
		assert.Len(t, records, 2, "Should return 2 records matching the complex filter")
	})

	t.Run("Test Query with Custom Resolver", func(t *testing.T) {
		// Register a custom resolver for case-insensitive search
		RegisterFilterResolver("CASE_INSENSITIVE", func(filter Filter) bson.M {
			field := filter.Field()
			value := filter.Value()
			if field == nil {
				return nil
			}
			if pattern, ok := value.(string); ok {
				return bson.M{field.Name(): bson.M{"$regex": pattern, "$options": "i"}}
			}
			return nil
		})

		// Create a custom filter
		customFilter := &filterImpl{
			field:    mustField(t, userSchema, "first_name"),
			value:    "john",
			operator: "CASE_INSENSITIVE",
		}

		query := NewMongoQuery(ctx, userSchema)
		query = query.Where(customFilter)
		records, err := query.Execute()
		assert.NoError(t, err, "Failed to execute query with custom resolver")
		assert.Len(t, records, 1, "Should return 1 record with case-insensitive match")

		// Check that the record has the expected first name (case-insensitive)
		firstName, ok := records[0].Value(mustField(t, userSchema, "first_name"))
		assert.True(t, ok, "Should have first_name field")
		assert.Equal(t, "John", firstName, "First name should be 'John' (case-insensitive match)")
	})
}
