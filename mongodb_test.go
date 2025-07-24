package jpack

import (
	"context"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
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
