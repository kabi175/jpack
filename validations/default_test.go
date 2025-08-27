package validations

import (
	"context"
	"testing"

	"github.com/kabi175/jpack"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func TestRequiredField(t *testing.T) {
	schema := jpack.NewSchema("test").
		Field("id", &jpack.String{}).
		Field("name", &jpack.String{}).
		Field("age", &jpack.Number{}).
		Build()

	t.Run("valid required field", func(t *testing.T) {
		record := jpack.NewMongoRecord(schema)
		idField, _ := schema.Field("id")
		record.SetValue(idField, "123")

		validation := RequiredField("id")
		err := validation(context.Background(), record)
		assert.NoError(t, err)
	})

	t.Run("missing required field", func(t *testing.T) {
		record := jpack.NewMongoRecord(schema)

		validation := RequiredField("id")
		err := validation(context.Background(), record)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "is required but not provided")
	})

	t.Run("empty string", func(t *testing.T) {
		record := jpack.NewMongoRecord(schema)
		nameField, _ := schema.Field("name")
		record.SetValue(nameField, "")

		validation := RequiredField("name")
		err := validation(context.Background(), record)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "is required but is empty")
	})

	t.Run("whitespace only string", func(t *testing.T) {
		record := jpack.NewMongoRecord(schema)
		nameField, _ := schema.Field("name")
		record.SetValue(nameField, "   ")

		validation := RequiredField("name")
		err := validation(context.Background(), record)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "is required but is empty")
	})

	t.Run("zero number", func(t *testing.T) {
		record := jpack.NewMongoRecord(schema)
		ageField, _ := schema.Field("age")
		record.SetValue(ageField, 0)

		validation := RequiredField("age")
		err := validation(context.Background(), record)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "is required but is zero")
	})

	t.Run("non-zero number", func(t *testing.T) {
		record := jpack.NewMongoRecord(schema)
		ageField, _ := schema.Field("age")
		record.SetValue(ageField, 25)

		validation := RequiredField("age")
		err := validation(context.Background(), record)
		assert.NoError(t, err)
	})
}

func TestReadOnlyField(t *testing.T) {
	schema := jpack.NewSchema("test").
		Field("id", &jpack.String{}).
		Field("created_at", &jpack.String{}).
		Build()

	t.Run("new record allows setting read-only field", func(t *testing.T) {
		record := jpack.NewMongoRecord(schema)
		createdAtField, _ := schema.Field("created_at")
		record.SetValue(createdAtField, "2024-01-01")

		validation := ReadOnlyField("created_at")
		err := validation(context.Background(), record)
		assert.NoError(t, err)
	})

	t.Run("existing record prevents editing read-only field", func(t *testing.T) {
		record := jpack.NewMongoRecord(schema)
		// Simulate an existing record by setting initial values
		idField, _ := schema.Field("id")
		createdAtField, _ := schema.Field("created_at")
		record.SetValue(idField, "123")
		record.SetValue(createdAtField, "2024-01-01")

		uri := "mongodb://localhost:27017"
		client, err := mongo.Connect(options.Client().
			ApplyURI(uri))

		assert.NoError(t, err, "Failed to connect to MongoDB")
		defer func() {
			err := client.Disconnect(context.TODO())
			assert.NoError(t, err, "Failed to disconnect from MongoDB")
		}()

		client.Database("jpack_test").Drop(context.TODO())
		ctx := context.WithValue(context.Background(), jpack.Conn, client.Database("jpack_test"))
		err = record.Save(ctx)
		assert.NoError(t, err, "Save should not return error")

		// Now try to modify the read-only field
		record.SetValue(createdAtField, "2024-01-02")

		validation := ReadOnlyField("created_at")
		err = validation(context.Background(), record)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "is read-only and cannot be modified")
	})
}

func TestStringLength(t *testing.T) {
	schema := jpack.NewSchema("test").
		Field("name", &jpack.String{}).
		Field("description", &jpack.String{}).
		Build()

	t.Run("valid string length", func(t *testing.T) {
		record := jpack.NewMongoRecord(schema)
		nameField, _ := schema.Field("name")
		record.SetValue(nameField, "John Doe")

		validation := StringLength("name", 3, 20)
		err := validation(context.Background(), record)
		assert.NoError(t, err)
	})

	t.Run("string too short", func(t *testing.T) {
		record := jpack.NewMongoRecord(schema)
		nameField, _ := schema.Field("name")
		record.SetValue(nameField, "Jo")

		validation := StringLength("name", 3, 20)
		err := validation(context.Background(), record)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must be at least 3 characters long")
	})

	t.Run("string too long", func(t *testing.T) {
		record := jpack.NewMongoRecord(schema)
		nameField, _ := schema.Field("name")
		record.SetValue(nameField, "This is a very long name that exceeds the limit")

		validation := StringLength("name", 3, 20)
		err := validation(context.Background(), record)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must be at most 20 characters long")
	})

	t.Run("nil value allowed", func(t *testing.T) {
		record := jpack.NewMongoRecord(schema)

		validation := StringLength("name", 3, 20)
		err := validation(context.Background(), record)
		assert.NoError(t, err)
	})

	t.Run("whitespace trimming", func(t *testing.T) {
		record := jpack.NewMongoRecord(schema)
		nameField, _ := schema.Field("name")
		record.SetValue(nameField, "  Jo  ") // Trimmed length is 2

		validation := StringLength("name", 3, 20)
		err := validation(context.Background(), record)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must be at least 3 characters long")
	})
}

func TestNumberRange(t *testing.T) {
	schema := jpack.NewSchema("test").
		Field("age", &jpack.Number{}).
		Field("score", &jpack.Number{}).
		Field("price", &jpack.Number{}).
		Build()

	t.Run("valid number in range", func(t *testing.T) {
		record := jpack.NewMongoRecord(schema)
		ageField, _ := schema.Field("age")
		record.SetValue(ageField, 25)

		validation := NumberRange("age", 18, 65)
		err := validation(context.Background(), record)
		assert.NoError(t, err)
	})

	t.Run("number below minimum", func(t *testing.T) {
		record := jpack.NewMongoRecord(schema)
		ageField, _ := schema.Field("age")
		record.SetValue(ageField, 15)

		validation := NumberRange("age", 18, 65)
		err := validation(context.Background(), record)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must be at least 18")
	})

	t.Run("number above maximum", func(t *testing.T) {
		record := jpack.NewMongoRecord(schema)
		ageField, _ := schema.Field("age")
		record.SetValue(ageField, 70)

		validation := NumberRange("age", 18, 65)
		err := validation(context.Background(), record)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must be at most 65")
	})

	t.Run("nil value allowed", func(t *testing.T) {
		record := jpack.NewMongoRecord(schema)

		validation := NumberRange("age", 18, 65)
		err := validation(context.Background(), record)
		assert.NoError(t, err)
	})

	t.Run("zero minimum", func(t *testing.T) {
		record := jpack.NewMongoRecord(schema)
		scoreField, _ := schema.Field("score")
		record.SetValue(scoreField, 85)

		validation := NumberRange("score", 0, 100)
		err := validation(context.Background(), record)
		assert.NoError(t, err)
	})

	t.Run("zero maximum (no upper limit)", func(t *testing.T) {
		record := jpack.NewMongoRecord(schema)
		priceField, _ := schema.Field("price")
		record.SetValue(priceField, 1000)

		validation := NumberRange("price", 0, 0) // No upper limit
		err := validation(context.Background(), record)
		assert.NoError(t, err)
	})
}

func TestEmailFormat(t *testing.T) {
	schema := jpack.NewSchema("test").
		Field("email", &jpack.String{}).
		Build()

	t.Run("valid email", func(t *testing.T) {
		record := jpack.NewMongoRecord(schema)
		emailField, _ := schema.Field("email")
		record.SetValue(emailField, "test@example.com")

		validation := EmailFormat("email")
		err := validation(context.Background(), record)
		assert.NoError(t, err)
	})

	t.Run("missing @ symbol", func(t *testing.T) {
		record := jpack.NewMongoRecord(schema)
		emailField, _ := schema.Field("email")
		record.SetValue(emailField, "testexample.com")

		validation := EmailFormat("email")
		err := validation(context.Background(), record)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing @")
	})

	t.Run("missing domain", func(t *testing.T) {
		record := jpack.NewMongoRecord(schema)
		emailField, _ := schema.Field("email")
		record.SetValue(emailField, "test@")

		validation := EmailFormat("email")
		err := validation(context.Background(), record)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing domain")
	})

	t.Run("missing local part", func(t *testing.T) {
		record := jpack.NewMongoRecord(schema)
		emailField, _ := schema.Field("email")
		record.SetValue(emailField, "@example.com")

		validation := EmailFormat("email")
		err := validation(context.Background(), record)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing local part")
	})

	t.Run("invalid format", func(t *testing.T) {
		record := jpack.NewMongoRecord(schema)
		emailField, _ := schema.Field("email")
		record.SetValue(emailField, "test@@example.com")

		validation := EmailFormat("email")
		err := validation(context.Background(), record)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid format")
	})

	t.Run("nil value allowed", func(t *testing.T) {
		record := jpack.NewMongoRecord(schema)

		validation := EmailFormat("email")
		err := validation(context.Background(), record)
		assert.NoError(t, err)
	})

	t.Run("empty string allowed", func(t *testing.T) {
		record := jpack.NewMongoRecord(schema)
		emailField, _ := schema.Field("email")
		record.SetValue(emailField, "")

		validation := EmailFormat("email")
		err := validation(context.Background(), record)
		assert.NoError(t, err)
	})
}

func TestPositiveNumber(t *testing.T) {
	schema := jpack.NewSchema("test").
		Field("quantity", &jpack.Number{}).
		Build()

	t.Run("positive number", func(t *testing.T) {
		record := jpack.NewMongoRecord(schema)
		quantityField, _ := schema.Field("quantity")
		record.SetValue(quantityField, 5)

		validation := PositiveNumber("quantity")
		err := validation(context.Background(), record)
		assert.NoError(t, err)
	})

	t.Run("zero number", func(t *testing.T) {
		record := jpack.NewMongoRecord(schema)
		quantityField, _ := schema.Field("quantity")
		record.SetValue(quantityField, 0)

		validation := PositiveNumber("quantity")
		err := validation(context.Background(), record)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must be at least")
	})

	t.Run("negative number", func(t *testing.T) {
		record := jpack.NewMongoRecord(schema)
		quantityField, _ := schema.Field("quantity")
		record.SetValue(quantityField, -5)

		validation := PositiveNumber("quantity")
		err := validation(context.Background(), record)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must be at least")
	})
}

func TestNonNegativeNumber(t *testing.T) {
	schema := jpack.NewSchema("test").
		Field("score", &jpack.Number{}).
		Build()

	t.Run("positive number", func(t *testing.T) {
		record := jpack.NewMongoRecord(schema)
		scoreField, _ := schema.Field("score")
		record.SetValue(scoreField, 85)

		validation := NonNegativeNumber("score")
		err := validation(context.Background(), record)
		assert.NoError(t, err)
	})

	t.Run("zero number", func(t *testing.T) {
		record := jpack.NewMongoRecord(schema)
		scoreField, _ := schema.Field("score")
		record.SetValue(scoreField, 0)

		validation := NonNegativeNumber("score")
		err := validation(context.Background(), record)
		assert.NoError(t, err)
	})

	t.Run("negative number", func(t *testing.T) {
		record := jpack.NewMongoRecord(schema)
		scoreField, _ := schema.Field("score")
		record.SetValue(scoreField, -10)

		validation := NonNegativeNumber("score")
		err := validation(context.Background(), record)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must be at least 0")
	})
}

func TestStringNotEmpty(t *testing.T) {
	schema := jpack.NewSchema("test").
		Field("title", &jpack.String{}).
		Build()

	t.Run("non-empty string", func(t *testing.T) {
		record := jpack.NewMongoRecord(schema)
		titleField, _ := schema.Field("title")
		record.SetValue(titleField, "Valid Title")

		validation := StringNotEmpty("title")
		err := validation(context.Background(), record)
		assert.NoError(t, err)
	})

	t.Run("empty string", func(t *testing.T) {
		record := jpack.NewMongoRecord(schema)
		titleField, _ := schema.Field("title")
		record.SetValue(titleField, "")

		validation := StringNotEmpty("title")
		err := validation(context.Background(), record)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must be at least 1 characters long")
	})

	t.Run("whitespace only", func(t *testing.T) {
		record := jpack.NewMongoRecord(schema)
		titleField, _ := schema.Field("title")
		record.SetValue(titleField, "   ")

		validation := StringNotEmpty("title")
		err := validation(context.Background(), record)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must be at least 1 characters long")
	})

	t.Run("nil value allowed", func(t *testing.T) {
		record := jpack.NewMongoRecord(schema)

		validation := StringNotEmpty("title")
		err := validation(context.Background(), record)
		assert.NoError(t, err)
	})
}
