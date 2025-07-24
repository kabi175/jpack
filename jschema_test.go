package jpack

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSchemaBuilder(t *testing.T) {
	t.Run("create schema", func(t *testing.T) {
		assert := assert.New(t)
		schemaBuilder := NewSchema("test_schema")
		assert.NotNil(schemaBuilder, "Schema Builder should not be nil")

		schema := schemaBuilder.Field("id", &Number{}).Build()
		assert.NotNil(schema, "Schema should not be nil")

		f, ok := schema.Field("id")
		assert.True(ok, "Field 'id' should exist in schema")
		assert.NotNil(f, "Field 'id' should not be nil")
		assert.Equal("id", f.Name(), "Field name should be 'id'")
	})

	t.Run("add ref field", func(t *testing.T) {
		assert := assert.New(t)

		userSchema := NewSchema("test_user").Field("id", &Number{}).
			Field("name", &String{}).
			Build()

		postSchema := NewSchema("test_post").Field("id", &Number{}).
			Field("title", &String{}).
			Ref("author", userSchema).
			Build()

		authorField, ok := postSchema.Field("author")
		assert.True(ok, "Field 'author' should exist in post schema")
		assert.NotNil(authorField, "Field 'author' should not be nil")
		assert.Equal("author", authorField.Name(), "Field name should be 'author'")
		assert.IsType(&Ref{}, authorField.Type(), "Field type should be Ref")

		authorRef, ok := authorField.(JRef)
		assert.True(ok, "Field 'author' should implement JRef")
		assert.NotNil(authorRef.RelSchema(), "Related schema should not be nil")
		assert.Equal(userSchema, authorRef.RelSchema(), "Field schema should match user schema")

	})

}
