package main

import (
	"context"

	"github.com/kabi175/jpack/schema"
)

// Convenience functions for common schema patterns

// CreateUserSchema creates a common user schema
func CreateUserSchema() schema.JSchema {
	return schema.NewSchemaBuilder("User").
		AddRequiredUniqueField("id", schema.JString, nil).
		AddRequiredField("name", schema.JString, nil).
		AddRequiredUniqueField("email", schema.JString, nil).
		AddField("age", schema.JInt, 18).
		AddField("created_at", schema.JTime, nil).
		AddField("updated_at", schema.JTime, nil).
		SetIDField("id").
		AddValidation(func(ctx context.Context, rec schema.JRecord) error {
			// Add common user validations here
			return nil
		}).
		BuildImmutable()
}

// CreateProductSchema creates a common product schema
func CreateProductSchema() schema.JSchema {
	return schema.NewSchemaBuilder("Product").
		AddRequiredUniqueField("id", schema.JString, nil).
		AddRequiredField("name", schema.JString, nil).
		AddField("description", schema.JString, nil).
		AddRequiredField("price", schema.JFloat64, 0.0).
		AddRequiredField("category", schema.JString, nil).
		AddField("tags", schema.JArray, nil).
		AddField("in_stock", schema.JBool, true).
		AddField("created_at", schema.JTime, nil).
		AddField("updated_at", schema.JTime, nil).
		SetIDField("id").
		AddValidation(func(ctx context.Context, rec schema.JRecord) error {
			// Add common product validations here
			return nil
		}).
		BuildImmutable()
}

// CreateOrderSchema creates a common order schema
func CreateOrderSchema() schema.JSchema {
	return schema.NewSchemaBuilder("Order").
		AddRequiredUniqueField("id", schema.JString, nil).
		AddRequiredField("customer_id", schema.JString, nil).
		AddRequiredField("total_amount", schema.JFloat64, 0.0).
		AddField("status", schema.JString, "pending").
		AddField("items", schema.JArray, nil).
		AddField("created_at", schema.JTime, nil).
		AddField("updated_at", schema.JTime, nil).
		SetIDField("id").
		AddValidation(func(ctx context.Context, rec schema.JRecord) error {
			// Add common order validations here
			return nil
		}).
		BuildImmutable()
}
