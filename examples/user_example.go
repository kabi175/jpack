package main

import (
	"context"
	"log"
	"time"

	"github.com/kabi175/jpack/converter"
	"github.com/kabi175/jpack/hooks"
	"github.com/kabi175/jpack/logger"
	"github.com/kabi175/jpack/mongo"
	"github.com/kabi175/jpack/projection"
	"github.com/kabi175/jpack/repository"
	"github.com/kabi175/jpack/schema"
	"github.com/kabi175/jpack/validation"
)

// User represents a user entity
type User struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Age       int       `json:"age"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UserExample demonstrates the framework usage
func UserExample() {
	// Initialize logger
	logger.Initialize()
	// Create MongoDB client
	client, err := mongo.NewMongoClient("mongodb://localhost:27017", "jpack_example")
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}
	defer client.Close(context.Background())

	// Create user schema using builder pattern (ID field is automatically created)
	userSchema := schema.NewSchemaBuilder("User").
		AddField("name", schema.JString, nil).
		AddField("email", schema.JString, nil).
		AddField("age", schema.JInt, 18).
		AddField("created_at", schema.JTime, nil).
		AddField("updated_at", schema.JTime, nil).
		Build()

	// Update schema with field validations and schema-level validations
	userSchema = userSchema.Update(func(sb *schema.SchemaBuilder) {
		// Add schema-level validations
		sb.AddValidation(func(ctx context.Context, rec schema.JRecord) error {
			return validation.ValidateMinLength("name", 2)(ctx, rec)
		})
		sb.AddValidation(func(ctx context.Context, rec schema.JRecord) error {
			return validation.ValidateMaxLength("name", 100)(ctx, rec)
		})
		sb.AddValidation(func(ctx context.Context, rec schema.JRecord) error {
			return validation.ValidateEmail("email")(ctx, rec)
		})
		sb.AddValidation(func(ctx context.Context, rec schema.JRecord) error {
			return validation.ValidateRange("age", 0, 150)(ctx, rec)
		})
	})

	// Update fields to be required/unique
	nameField, _ := userSchema.Field("name")
	nameField = nameField.Update(func(fb *schema.FieldBuilder) {
		fb.Required()
	})

	emailField, _ := userSchema.Field("email")
	emailField = emailField.Update(func(fb *schema.FieldBuilder) {
		fb.Required().Unique()
	})

	ageField, _ := userSchema.Field("age")
	ageField = ageField.Update(func(fb *schema.FieldBuilder) {
		fb.Required()
	})

	// Register schema
	err = client.RegisterSchema(userSchema)
	if err != nil {
		log.Fatal("Failed to register schema:", err)
	}

	// Create indexes
	err = client.CreateIndexesForSchema(context.Background(), "User")
	if err != nil {
		log.Fatal("Failed to create indexes:", err)
	}

	// Get repository
	userRepo, err := client.GetRepository("User")
	if err != nil {
		log.Fatal("Failed to get repository:", err)
	}

	// Set up hooks
	err = hooks.NewHookManager(userSchema).
		AddTimestampHook("created_at", "updated_at").
		AddValidationHook().
		AddLoggingHook(nil).RegisterAll()

	if err != nil {
		log.Fatal("Failed to register hooks:", err)
	}

	// Create a user record
	user := schema.NewJRecord().
		Set("id", "user_001").
		Set("name", "John Doe").
		Set("email", "john.doe@example.com").
		Set("age", 30)

	// Save user
	savedUser, err := userRepo.Save(context.Background(), user)
	if err != nil {
		log.Fatal("Failed to save user:", err)
	}

	logger.App.Info().
		Interface("user", savedUser.ToMap()).
		Msg("Saved user")

	// Find user by ID
	foundUser, err := userRepo.FindById(context.Background(), "user_001")
	if err != nil {
		log.Fatal("Failed to find user:", err)
	}

	logger.App.Info().
		Interface("user", foundUser.ToMap()).
		Msg("Found user")

	// Query users with criteria
	criteria := repository.NewJCriteriaBuilder().
		WhereGreaterThanOrEqual("age", 18).
		OrderBy("name").
		SetLimit(10).
		Build()

	users, err := userRepo.FindBy(context.Background(), criteria)
	if err != nil {
		log.Fatal("Failed to find users:", err)
	}

	logger.App.Info().
		Int("count", len(users)).
		Msg("Found users")
	for _, u := range users {
		logger.App.Info().
			Str("name", u.Get("name").(string)).
			Str("email", u.Get("email").(string)).
			Msg("User details")
	}

	// Update user
	updateCriteria := repository.NewJCriteriaBuilder().
		Where("id", "user_001").
		Build()

	updates := schema.NewJRecord().Set("age", 31)
	err = userRepo.Update(context.Background(), updateCriteria, updates)
	if err != nil {
		log.Fatal("Failed to update user:", err)
	}

	// Count users
	count, err := userRepo.Count(context.Background(), repository.JCriteria{})
	if err != nil {
		log.Fatal("Failed to count users:", err)
	}

	logger.App.Info().
		Int64("count", count).
		Msg("Total users")

	// Use projections
	proj := projection.NewIncludeProjection("name", "email")
	projectedUsers, err := userRepo.FindBy(context.Background(), repository.JCriteria{
		Projection: proj.Fields,
	})
	if err != nil {
		log.Fatal("Failed to project users:", err)
	}

	logger.App.Info().
		Interface("users", projectedUsers).
		Msg("Projected users")

	// Use aggregations
	collection := client.GetDatabase().Collection("User")
	aggExecutor := projection.NewAggregationExecutor(collection)
	aggService := projection.NewAggregationService(aggExecutor)

	// Count users
	userCount, err := aggService.Count(context.Background(), repository.JCriteria{})
	if err != nil {
		log.Fatal("Failed to aggregate count:", err)
	}

	logger.App.Info().
		Int64("count", userCount).
		Msg("Aggregated count")

	// Average age
	avgAge, err := aggService.Avg(context.Background(), repository.JCriteria{}, "age")
	if err != nil {
		log.Fatal("Failed to aggregate average age:", err)
	}

	logger.App.Info().
		Float64("avg_age", avgAge).
		Msg("Average age")

	// Delete user
	err = userRepo.Delete(context.Background(), "user_001")
	if err != nil {
		log.Fatal("Failed to delete user:", err)
	}

	logger.App.Info().Msg("User deleted successfully")
}

// ProductExample demonstrates more complex usage
func ProductExample() {
	// Create MongoDB client
	client, err := mongo.NewMongoClient("mongodb://localhost:27017", "jpack_example")
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}
	defer client.Close(context.Background())

	// Create product schema using builder pattern (ID field is automatically created)
	productSchema := schema.NewSchemaBuilder("Product").
		AddField("name", schema.JString, nil).
		AddField("description", schema.JString, nil).
		AddField("price", schema.JFloat64, 0.0).
		AddField("category", schema.JString, nil).
		AddField("tags", schema.JArray, nil).
		AddField("in_stock", schema.JBool, true).
		AddField("created_at", schema.JTime, nil).
		AddField("updated_at", schema.JTime, nil).
		Build()

	// Update product schema with validations
	productSchema = productSchema.Update(func(sb *schema.SchemaBuilder) {
		// Add schema-level validations
		sb.AddValidation(func(ctx context.Context, rec schema.JRecord) error {
			return validation.ValidateMinLength("name", 3)(ctx, rec)
		})
		sb.AddValidation(func(ctx context.Context, rec schema.JRecord) error {
			return validation.ValidateMinValue("price", 0.0)(ctx, rec)
		})
		sb.AddValidation(func(ctx context.Context, rec schema.JRecord) error {
			return validation.ValidateFieldIn("category", []any{"electronics", "clothing", "books", "home"})(ctx, rec)
		})
	})

	// Update fields to be required
	prodNameField, _ := productSchema.Field("name")
	prodNameField = prodNameField.Update(func(fb *schema.FieldBuilder) {
		fb.Required()
	})

	prodPriceField, _ := productSchema.Field("price")
	prodPriceField = prodPriceField.Update(func(fb *schema.FieldBuilder) {
		fb.Required()
	})

	prodCategoryField, _ := productSchema.Field("category")
	prodCategoryField = prodCategoryField.Update(func(fb *schema.FieldBuilder) {
		fb.Required()
	})

	// Register schema
	err = client.RegisterSchema(productSchema)
	if err != nil {
		log.Fatal("Failed to register schema:", err)
	}

	// Get repository
	productRepo, err := client.GetRepository("Product")
	if err != nil {
		log.Fatal("Failed to get repository:", err)
	}

	// Set up hooks
	hookManager := hooks.NewHookManager(productSchema)
	hookManager.AddTimestampHook("created_at", "updated_at").
		AddValidationHook().
		AddLoggingHook(nil)

	err = hookManager.RegisterAll()
	if err != nil {
		log.Fatal("Failed to register hooks:", err)
	}

	// Create products
	products := []schema.JRecord{
		schema.NewJRecord().
			Set("id", "prod_001").
			Set("name", "Laptop").
			Set("description", "High-performance laptop").
			Set("price", 999.99).
			Set("category", "electronics").
			Set("tags", []string{"laptop", "computer", "electronics"}).
			Set("in_stock", true),
		schema.NewJRecord().
			Set("id", "prod_002").
			Set("name", "T-Shirt").
			Set("description", "Comfortable cotton t-shirt").
			Set("price", 19.99).
			Set("category", "clothing").
			Set("tags", []string{"clothing", "t-shirt", "cotton"}).
			Set("in_stock", true),
		schema.NewJRecord().
			Set("id", "prod_003").
			Set("name", "Programming Book").
			Set("description", "Learn Go programming").
			Set("price", 49.99).
			Set("category", "books").
			Set("tags", []string{"book", "programming", "go"}).
			Set("in_stock", false),
	}

	// Save products
	for _, product := range products {
		_, err := productRepo.Save(context.Background(), product)
		if err != nil {
			log.Fatal("Failed to save product:", err)
		}
	}

	// Query products by category
	criteria := repository.NewJCriteriaBuilder().
		Where("category", "electronics").
		OrderBy("price").
		Build()

	electronics, err := productRepo.FindBy(context.Background(), criteria)
	if err != nil {
		log.Fatal("Failed to find electronics:", err)
	}

	logger.App.Info().
		Int("count", len(electronics)).
		Msg("Found electronics products")
	for _, p := range electronics {
		logger.App.Info().
			Str("name", p.Get("name").(string)).
			Float64("price", p.Get("price").(float64)).
			Msg("Product details")
	}

	// Query products in stock
	inStockCriteria := repository.NewJCriteriaBuilder().
		Where("in_stock", true).
		WhereLessThan("price", 100.0).
		OrderBy("name").
		Build()

	inStockProducts, err := productRepo.FindBy(context.Background(), inStockCriteria)
	if err != nil {
		log.Fatal("Failed to find in-stock products:", err)
	}

	logger.App.Info().
		Int("count", len(inStockProducts)).
		Msg("Found in-stock products under $100")
	for _, p := range inStockProducts {
		logger.App.Info().
			Str("name", p.Get("name").(string)).
			Float64("price", p.Get("price").(float64)).
			Msg("Product details")
	}

	// Use custom operators
	todayCriteria := repository.NewJCriteriaBuilder().
		WhereIsToday("created_at").
		Build()

	todayProducts, err := productRepo.FindBy(context.Background(), todayCriteria)
	if err != nil {
		log.Fatal("Failed to find today's products:", err)
	}

	logger.App.Info().
		Int("count", len(todayProducts)).
		Msg("Found products created today")

	// Use aggregations
	productCollection := client.GetDatabase().Collection("Product")
	aggExecutor := projection.NewAggregationExecutor(productCollection)
	aggService := projection.NewAggregationService(aggExecutor)

	// Average price by category
	avgPrice, err := aggService.Avg(context.Background(), repository.JCriteria{}, "price")
	if err != nil {
		log.Fatal("Failed to get average price:", err)
	}

	logger.App.Info().
		Float64("avg_price", avgPrice).
		Msg("Average product price")

	// Group by category
	categoryGroups, err := aggService.GroupBy(context.Background(), repository.JCriteria{}, "category")
	if err != nil {
		log.Fatal("Failed to group by category:", err)
	}

	logger.App.Info().
		Interface("groups", categoryGroups).
		Msg("Products grouped by category")

	// Clean up
	for _, product := range products {
		err = productRepo.Delete(context.Background(), product.Get("id"))
		if err != nil {
			log.Printf("Failed to delete product %s: %v", product.Get("id"), err)
		}
	}
}

// CustomConverterExample demonstrates custom converters
func CustomConverterExample() {
	// Register custom converters
	err := converter.RegisterConverter(converter.NewMoneyConverter())
	if err != nil {
		log.Fatal("Failed to register money converter:", err)
	}

	err = converter.RegisterConverter(converter.NewEmailConverter())
	if err != nil {
		log.Fatal("Failed to register email converter:", err)
	}

	// Create order schema with custom types using builder pattern (ID field is automatically created)
	orderSchema := schema.NewSchemaBuilder("Order").
		AddField("customer_email", "email", nil).
		AddField("total_amount", "money", nil).
		AddField("status", schema.JString, "pending").
		AddField("created_at", schema.JTime, nil).
		Build()

	// Update order schema with validations
	orderSchema = orderSchema.Update(func(sb *schema.SchemaBuilder) {
		// Add schema-level validations
		sb.AddValidation(func(ctx context.Context, rec schema.JRecord) error {
			return validation.ValidateFieldIn("status", []any{"pending", "processing", "shipped", "delivered", "cancelled"})(ctx, rec)
		})
	})

	// Update fields to be required
	orderEmailField, _ := orderSchema.Field("customer_email")
	orderEmailField = orderEmailField.Update(func(fb *schema.FieldBuilder) {
		fb.Required()
	})

	orderAmountField, _ := orderSchema.Field("total_amount")
	orderAmountField = orderAmountField.Update(func(fb *schema.FieldBuilder) {
		fb.Required()
	})

	orderStatusField, _ := orderSchema.Field("status")
	orderStatusField = orderStatusField.Update(func(fb *schema.FieldBuilder) {
		fb.Required()
	})

	// Create MongoDB client
	client, err := mongo.NewMongoClient("mongodb://localhost:27017", "jpack_example")
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}
	defer client.Close(context.Background())

	// Register schema
	err = client.RegisterSchema(orderSchema)
	if err != nil {
		log.Fatal("Failed to register schema:", err)
	}

	// Get repository
	orderRepo, err := client.GetRepository("Order")
	if err != nil {
		log.Fatal("Failed to get repository:", err)
	}

	// Create order with custom types
	order := schema.NewJRecord().
		Set("id", "order_001").
		Set("customer_email", converter.Email{
			Address:  "customer@example.com",
			Verified: true,
		}).
		Set("total_amount", converter.Money{
			Amount:   99.99,
			Currency: "USD",
		}).
		Set("status", "pending")

	// Save order
	savedOrder, err := orderRepo.Save(context.Background(), order)
	if err != nil {
		log.Fatal("Failed to save order:", err)
	}

	logger.App.Info().
		Interface("order", savedOrder.ToMap()).
		Msg("Saved order")

	// Find order
	foundOrder, err := orderRepo.FindById(context.Background(), "order_001")
	if err != nil {
		log.Fatal("Failed to find order:", err)
	}

	logger.App.Info().
		Interface("order", foundOrder.ToMap()).
		Msg("Found order")

	// Clean up
	err = orderRepo.Delete(context.Background(), "order_001")
	if err != nil {
		log.Fatal("Failed to delete order:", err)
	}
}

// RunAllExamples runs all examples
func RunAllExamples() {
	logger.App.Info().Msg("=== User Example ===")
	UserExample()

	logger.App.Info().Msg("=== Product Example ===")
	ProductExample()

	logger.App.Info().Msg("=== Custom Converter Example ===")
	CustomConverterExample()

	logger.App.Info().Msg("All examples completed successfully!")
}
