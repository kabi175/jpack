package schema

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewJEdge(t *testing.T) {
	tests := []struct {
		name         string
		edgeName     string
		fromSchema   JSchema
		toSchema     JSchema
		edgeType     EdgeType
		expectedName string
		expectedType EdgeType
	}{
		{
			name:         "one-to-one edge",
			edgeName:     "user_profile",
			fromSchema:   NewJSchema("User"),
			toSchema:     NewJSchema("Profile"),
			edgeType:     EdgeOneToOne,
			expectedName: "user_profile",
			expectedType: EdgeOneToOne,
		},
		{
			name:         "one-to-many edge",
			edgeName:     "user_orders",
			fromSchema:   NewJSchema("User"),
			toSchema:     NewJSchema("Order"),
			edgeType:     EdgeOneToMany,
			expectedName: "user_orders",
			expectedType: EdgeOneToMany,
		},
		{
			name:         "many-to-one edge",
			edgeName:     "order_user",
			fromSchema:   NewJSchema("Order"),
			toSchema:     NewJSchema("User"),
			edgeType:     EdgeManyToOne,
			expectedName: "order_user",
			expectedType: EdgeManyToOne,
		},
		{
			name:         "many-to-many edge",
			edgeName:     "user_roles",
			fromSchema:   NewJSchema("User"),
			toSchema:     NewJSchema("Role"),
			edgeType:     EdgeManyToMany,
			expectedName: "user_roles",
			expectedType: EdgeManyToMany,
		},
		{
			name:         "empty edge name",
			edgeName:     "",
			fromSchema:   NewJSchema("From"),
			toSchema:     NewJSchema("To"),
			edgeType:     EdgeOneToOne,
			expectedName: "",
			expectedType: EdgeOneToOne,
		},
		{
			name:         "edge with special characters",
			edgeName:     "user-profile_v2",
			fromSchema:   NewJSchema("User"),
			toSchema:     NewJSchema("Profile"),
			edgeType:     EdgeOneToOne,
			expectedName: "user-profile_v2",
			expectedType: EdgeOneToOne,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			edge := NewJEdge(tt.edgeName, tt.fromSchema, tt.toSchema, tt.edgeType)

			assert.Equal(t, tt.expectedName, edge.Name())
			assert.Equal(t, tt.fromSchema, edge.FromSchema())
			assert.Equal(t, tt.toSchema, edge.ToSchema())
			assert.Equal(t, tt.expectedType, edge.Type())
		})
	}
}

func TestJEdge_EdgeCases(t *testing.T) {
	t.Run("nil from schema", func(t *testing.T) {
		toSchema := NewJSchema("To")
		edge := NewJEdge("test", nil, toSchema, EdgeOneToOne)

		assert.Equal(t, "test", edge.Name())
		assert.Nil(t, edge.FromSchema())
		assert.Equal(t, toSchema, edge.ToSchema())
		assert.Equal(t, EdgeOneToOne, edge.Type())
	})

	t.Run("nil to schema", func(t *testing.T) {
		fromSchema := NewJSchema("From")
		edge := NewJEdge("test", fromSchema, nil, EdgeOneToOne)

		assert.Equal(t, "test", edge.Name())
		assert.Equal(t, fromSchema, edge.FromSchema())
		assert.Nil(t, edge.ToSchema())
		assert.Equal(t, EdgeOneToOne, edge.Type())
	})

	t.Run("both schemas nil", func(t *testing.T) {
		edge := NewJEdge("test", nil, nil, EdgeOneToOne)

		assert.Equal(t, "test", edge.Name())
		assert.Nil(t, edge.FromSchema())
		assert.Nil(t, edge.ToSchema())
		assert.Equal(t, EdgeOneToOne, edge.Type())
	})

	t.Run("self-referencing edge", func(t *testing.T) {
		schema := NewJSchema("SelfRef")
		edge := NewJEdge("self", schema, schema, EdgeOneToOne)

		assert.Equal(t, "self", edge.Name())
		assert.Equal(t, schema, edge.FromSchema())
		assert.Equal(t, schema, edge.ToSchema())
		assert.Equal(t, EdgeOneToOne, edge.Type())
	})

	t.Run("circular edge", func(t *testing.T) {
		userSchema := NewJSchema("User")
		orderSchema := NewJSchema("Order")

		// User -> Order
		userOrderEdge := NewJEdge("user_orders", userSchema, orderSchema, EdgeOneToMany)

		// Order -> User
		orderUserEdge := NewJEdge("order_user", orderSchema, userSchema, EdgeManyToOne)

		assert.Equal(t, "user_orders", userOrderEdge.Name())
		assert.Equal(t, userSchema, userOrderEdge.FromSchema())
		assert.Equal(t, orderSchema, userOrderEdge.ToSchema())
		assert.Equal(t, EdgeOneToMany, userOrderEdge.Type())

		assert.Equal(t, "order_user", orderUserEdge.Name())
		assert.Equal(t, orderSchema, orderUserEdge.FromSchema())
		assert.Equal(t, userSchema, orderUserEdge.ToSchema())
		assert.Equal(t, EdgeManyToOne, orderUserEdge.Type())
	})
}

func TestJEdge_AllEdgeTypes(t *testing.T) {
	userSchema := NewJSchema("User")
	orderSchema := NewJSchema("Order")

	edgeTypes := []EdgeType{EdgeOneToOne, EdgeOneToMany, EdgeManyToOne, EdgeManyToMany}

	for _, edgeType := range edgeTypes {
		t.Run(string(edgeType), func(t *testing.T) {
			edge := NewJEdge("test_edge", userSchema, orderSchema, edgeType)

			assert.Equal(t, "test_edge", edge.Name())
			assert.Equal(t, userSchema, edge.FromSchema())
			assert.Equal(t, orderSchema, edge.ToSchema())
			assert.Equal(t, edgeType, edge.Type())
		})
	}
}

func TestJEdge_ComplexSchemas(t *testing.T) {
	t.Run("edge between complex schemas", func(t *testing.T) {
		userSchema := NewJSchema("User")
		userSchema.AddField("id", JString, "").SetRequired(true)
		userSchema.AddField("name", JString, "").SetRequired(true)
		userSchema.AddField("email", JString, "").SetRequired(true)

		orderSchema := NewJSchema("Order")
		orderSchema.AddField("id", JString, "").SetRequired(true)
		orderSchema.AddField("total", JFloat64, 0.0)
		orderSchema.AddField("status", JString, "pending")

		edge := NewJEdge("user_orders", userSchema, orderSchema, EdgeOneToMany)

		assert.Equal(t, "user_orders", edge.Name())
		assert.Equal(t, userSchema, edge.FromSchema())
		assert.Equal(t, orderSchema, edge.ToSchema())
		assert.Equal(t, EdgeOneToMany, edge.Type())

		// Verify schemas have fields
		assert.Len(t, userSchema.Fields(), 3)
		assert.Len(t, orderSchema.Fields(), 3)
	})

	t.Run("edge between schemas with refs", func(t *testing.T) {
		userSchema := NewJSchema("User")
		profileSchema := NewJSchema("Profile")
		orderSchema := NewJSchema("Order")

		// User has profile
		userSchema.AddRef("profile", profileSchema)

		// Order references user
		orderSchema.AddRef("user", userSchema)

		edge := NewJEdge("user_orders", userSchema, orderSchema, EdgeOneToMany)

		assert.Equal(t, "user_orders", edge.Name())
		assert.Equal(t, userSchema, edge.FromSchema())
		assert.Equal(t, orderSchema, edge.ToSchema())
		assert.Equal(t, EdgeOneToMany, edge.Type())

		// Verify schemas have refs
		assert.Len(t, userSchema.Refs(), 1)
		assert.Len(t, orderSchema.Refs(), 1)
	})

	t.Run("edge between schemas with edges", func(t *testing.T) {
		userSchema := NewJSchema("User")
		orderSchema := NewJSchema("Order")
		productSchema := NewJSchema("Product")

		// User -> Order
		userOrderEdge := NewJEdge("user_orders", userSchema, orderSchema, EdgeOneToMany)
		userSchema.AddEdge(userOrderEdge)

		// Order -> Product
		orderProductEdge := NewJEdge("order_products", orderSchema, productSchema, EdgeManyToMany)
		orderSchema.AddEdge(orderProductEdge)

		// User -> Product (through orders)
		userProductEdge := NewJEdge("user_products", userSchema, productSchema, EdgeOneToMany)

		assert.Equal(t, "user_products", userProductEdge.Name())
		assert.Equal(t, userSchema, userProductEdge.FromSchema())
		assert.Equal(t, productSchema, userProductEdge.ToSchema())
		assert.Equal(t, EdgeOneToMany, userProductEdge.Type())

		// Verify schemas have edges
		assert.Len(t, userSchema.Edges(), 1)
		assert.Len(t, orderSchema.Edges(), 1)
	})
}

func TestJEdge_ImmutableSchemas(t *testing.T) {
	t.Run("edge between immutable schemas", func(t *testing.T) {
		userSchema := NewImmutableJSchema("User")
		orderSchema := NewImmutableJSchema("Order")

		edge := NewJEdge("user_orders", userSchema, orderSchema, EdgeOneToMany)

		assert.Equal(t, "user_orders", edge.Name())
		assert.Equal(t, userSchema, edge.FromSchema())
		assert.Equal(t, orderSchema, edge.ToSchema())
		assert.Equal(t, EdgeOneToMany, edge.Type())

		// Verify schemas are immutable
		assert.True(t, userSchema.IsImmutable())
		assert.True(t, orderSchema.IsImmutable())
	})

	t.Run("edge between mutable and immutable schemas", func(t *testing.T) {
		userSchema := NewJSchema("User")
		orderSchema := NewImmutableJSchema("Order")

		edge := NewJEdge("user_orders", userSchema, orderSchema, EdgeOneToMany)

		assert.Equal(t, "user_orders", edge.Name())
		assert.Equal(t, userSchema, edge.FromSchema())
		assert.Equal(t, orderSchema, edge.ToSchema())
		assert.Equal(t, EdgeOneToMany, edge.Type())

		// Verify schema immutability
		assert.False(t, userSchema.IsImmutable())
		assert.True(t, orderSchema.IsImmutable())
	})
}

func TestJEdge_StringRepresentation(t *testing.T) {
	userSchema := NewJSchema("User")
	orderSchema := NewJSchema("Order")

	t.Run("one-to-one edge string", func(t *testing.T) {
		edge := NewJEdge("user_profile", userSchema, orderSchema, EdgeOneToOne)

		str := fmt.Sprintf("Edge: %s (%s) -> %s",
			edge.Name(),
			edge.Type(),
			edge.ToSchema().Name())
		assert.Equal(t, "Edge: user_profile (one-to-one) -> Order", str)
	})

	t.Run("one-to-many edge string", func(t *testing.T) {
		edge := NewJEdge("user_orders", userSchema, orderSchema, EdgeOneToMany)

		str := fmt.Sprintf("Edge: %s (%s) -> %s",
			edge.Name(),
			edge.Type(),
			edge.ToSchema().Name())
		assert.Equal(t, "Edge: user_orders (one-to-many) -> Order", str)
	})

	t.Run("many-to-one edge string", func(t *testing.T) {
		edge := NewJEdge("order_user", orderSchema, userSchema, EdgeManyToOne)

		str := fmt.Sprintf("Edge: %s (%s) -> %s",
			edge.Name(),
			edge.Type(),
			edge.ToSchema().Name())
		assert.Equal(t, "Edge: order_user (many-to-one) -> User", str)
	})

	t.Run("many-to-many edge string", func(t *testing.T) {
		edge := NewJEdge("user_roles", userSchema, orderSchema, EdgeManyToMany)

		str := fmt.Sprintf("Edge: %s (%s) -> %s",
			edge.Name(),
			edge.Type(),
			edge.ToSchema().Name())
		assert.Equal(t, "Edge: user_roles (many-to-many) -> Order", str)
	})
}

func TestJEdge_Comparison(t *testing.T) {
	userSchema1 := NewJSchema("User1")
	userSchema2 := NewJSchema("User2")
	orderSchema := NewJSchema("Order")

	edge1 := NewJEdge("user_orders", userSchema1, orderSchema, EdgeOneToMany)
	edge2 := NewJEdge("user_orders", userSchema1, orderSchema, EdgeOneToMany)
	edge3 := NewJEdge("user_orders", userSchema2, orderSchema, EdgeOneToMany)
	edge4 := NewJEdge("user_orders", userSchema1, orderSchema, EdgeOneToOne)

	// Edges with same properties should have equal content
	assert.Equal(t, edge1.Name(), edge2.Name())
	assert.Equal(t, edge1.FromSchema(), edge2.FromSchema())
	assert.Equal(t, edge1.ToSchema(), edge2.ToSchema())
	assert.Equal(t, edge1.Type(), edge2.Type())

	// Different from schemas
	assert.NotEqual(t, edge1.FromSchema(), edge3.FromSchema())

	// Different edge types
	assert.NotEqual(t, edge1.Type(), edge4.Type())
}

func TestJEdge_ConcurrentAccess(t *testing.T) {
	userSchema := NewJSchema("User")
	orderSchema := NewJSchema("Order")
	edge := NewJEdge("user_orders", userSchema, orderSchema, EdgeOneToMany)

	// Test concurrent access to edge properties
	done := make(chan bool, 3)

	go func() {
		defer func() { done <- true }()
		for i := 0; i < 100; i++ {
			_ = edge.Name()
		}
	}()

	go func() {
		defer func() { done <- true }()
		for i := 0; i < 100; i++ {
			_ = edge.FromSchema()
		}
	}()

	go func() {
		defer func() { done <- true }()
		for i := 0; i < 100; i++ {
			_ = edge.ToSchema()
		}
	}()

	// Wait for all goroutines to complete
	for i := 0; i < 3; i++ {
		<-done
	}

	// Edge should still be in a valid state
	assert.Equal(t, "user_orders", edge.Name())
	assert.Equal(t, userSchema, edge.FromSchema())
	assert.Equal(t, orderSchema, edge.ToSchema())
	assert.Equal(t, EdgeOneToMany, edge.Type())
}

func TestJEdge_DeepNesting(t *testing.T) {
	// Create deeply nested schema structure
	level1 := NewJSchema("Level1")
	level2 := NewJSchema("Level2")
	level3 := NewJSchema("Level3")
	level4 := NewJSchema("Level4")

	// Create edges between levels
	edge1 := NewJEdge("level1_to_level2", level1, level2, EdgeOneToMany)
	edge2 := NewJEdge("level2_to_level3", level2, level3, EdgeOneToMany)
	edge3 := NewJEdge("level3_to_level4", level3, level4, EdgeOneToMany)
	edge4 := NewJEdge("level4_to_level1", level4, level1, EdgeManyToOne)

	assert.Equal(t, "level1_to_level2", edge1.Name())
	assert.Equal(t, level1, edge1.FromSchema())
	assert.Equal(t, level2, edge1.ToSchema())
	assert.Equal(t, EdgeOneToMany, edge1.Type())

	assert.Equal(t, "level2_to_level3", edge2.Name())
	assert.Equal(t, level2, edge2.FromSchema())
	assert.Equal(t, level3, edge2.ToSchema())
	assert.Equal(t, EdgeOneToMany, edge2.Type())

	assert.Equal(t, "level3_to_level4", edge3.Name())
	assert.Equal(t, level3, edge3.FromSchema())
	assert.Equal(t, level4, edge3.ToSchema())
	assert.Equal(t, EdgeOneToMany, edge3.Type())

	assert.Equal(t, "level4_to_level1", edge4.Name())
	assert.Equal(t, level4, edge4.FromSchema())
	assert.Equal(t, level1, edge4.ToSchema())
	assert.Equal(t, EdgeManyToOne, edge4.Type())
}

func TestJEdge_ComplexRelationship(t *testing.T) {
	// Create a complex relationship structure
	userSchema := NewJSchema("User")
	orderSchema := NewJSchema("Order")
	productSchema := NewJSchema("Product")
	categorySchema := NewJSchema("Category")

	// User -> Order (one-to-many)
	userOrderEdge := NewJEdge("user_orders", userSchema, orderSchema, EdgeOneToMany)

	// Order -> Product (many-to-many)
	orderProductEdge := NewJEdge("order_products", orderSchema, productSchema, EdgeManyToMany)

	// Product -> Category (many-to-one)
	productCategoryEdge := NewJEdge("product_category", productSchema, categorySchema, EdgeManyToOne)

	// User -> Product (through orders, one-to-many)
	userProductEdge := NewJEdge("user_products", userSchema, productSchema, EdgeOneToMany)

	// Category -> Product (one-to-many)
	categoryProductEdge := NewJEdge("category_products", categorySchema, productSchema, EdgeOneToMany)

	// Verify all edges
	assert.Equal(t, "user_orders", userOrderEdge.Name())
	assert.Equal(t, EdgeOneToMany, userOrderEdge.Type())

	assert.Equal(t, "order_products", orderProductEdge.Name())
	assert.Equal(t, EdgeManyToMany, orderProductEdge.Type())

	assert.Equal(t, "product_category", productCategoryEdge.Name())
	assert.Equal(t, EdgeManyToOne, productCategoryEdge.Type())

	assert.Equal(t, "user_products", userProductEdge.Name())
	assert.Equal(t, EdgeOneToMany, userProductEdge.Type())

	assert.Equal(t, "category_products", categoryProductEdge.Name())
	assert.Equal(t, EdgeOneToMany, categoryProductEdge.Type())
}

func TestJEdge_Validation(t *testing.T) {
	userSchema := NewJSchema("User")
	orderSchema := NewJSchema("Order")

	// Add validations to schemas
	userSchema.AddValidation(func(ctx context.Context, rec JRecord) error {
		return nil
	})

	orderSchema.AddValidation(func(ctx context.Context, rec JRecord) error {
		return nil
	})

	// Create edge
	edge := NewJEdge("user_orders", userSchema, orderSchema, EdgeOneToMany)

	assert.Equal(t, "user_orders", edge.Name())
	assert.Equal(t, userSchema, edge.FromSchema())
	assert.Equal(t, orderSchema, edge.ToSchema())
	assert.Equal(t, EdgeOneToMany, edge.Type())

	// Verify schemas have validations
	assert.Len(t, userSchema.Validations(), 1)
	assert.Len(t, orderSchema.Validations(), 1)
}
