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

}
