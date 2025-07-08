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

func mustField(t *testing.T, name string) JField {
	t.Helper()
	field, ok := userSchema.Field(name)
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
		m.SetValue(mustField(t, "first_name"), "Jhon")
		m.SetValue(mustField(t, "last_name"), "Doe")
		m.SetValue(mustField(t, "email"), "Jhon@example.com")
		m.SetValue(mustField(t, "age"), "23")
		err = m.Save(ctx)
		assert.NoError(t, err, "Failed to save record to MongoDB")

		pkId, ok := m.Value(mustField(t, "id"))
		assert.True(t, ok, "Primary key should be set after saving a new record")
		assert.NotEmpty(t, pkId, "Primary key should not be empty after saving a new record")
		assert.IsType(t, "", pkId, "Primary key should be a string")
	})

	t.Run("Update Record", func(t *testing.T) {
		m.SetValue(mustField(t, "email"), "Jhon@gmail.com")
		err = m.Save(ctx)
		assert.NoError(t, err, "Failed to update record in MongoDB")
	})

}
