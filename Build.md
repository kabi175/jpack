Create a **schema-driven persistence framework in Go** that works like JPA but is tailored for MongoDB. The framework should support:

### 1. Schema Definition

* Define a `JSchema` interface:

  * `Name() string`
  * `Fields() []JField`
  * `Field(name string) (JField, bool)`
  * `AddField(name string, fType JFieldType, defaultValue any) JField`
  * `AddRef(name string, schema JSchema) JRef`
  * `Edge() []JEdge`
  * `AddEdge(edge JEdge) JSchema`
  * `Validate(ctx context.Context, rec JRecord) error`
  * `AddValidation(fn ValidationFunc)`
  * `Validations() []ValidationFunc`
* Define supporting types:

  * `JField`, `JFieldType`, `JRef`, `JEdge`
  * `JRecord` → map-like structure for data
  * `ValidationFunc` → `func(ctx context.Context, rec JRecord) error`
  * `JSchemaRegistry` to store schemas

### 2. Repository Layer

* Define `CrudRepository`:

  * `Save(ctx, rec JRecord) (JRecord, error)`
  * `FindById(ctx, id any) (JRecord, error)`
  * `Delete(ctx, id any) error`
  * `FindAll(ctx) ([]JRecord, error)`
* Define `JpaRepository` (extends CrudRepository):

  * `FindBy(ctx, criteria JCriteria) ([]JRecord, error)`
  * `Count(ctx, criteria JCriteria) (int64, error)`
  * `Update(ctx, criteria JCriteria, updates JRecord) error`
* Define `JCriteria`, `JFilter`, `JSort` for querying.
* Add support for **projections** (selecting subset of fields).

### 3. Custom DataType Converters

* Define `JConverter` interface:

  ```go
  type JConverter interface {
      ToDB(value any) (any, error)
      FromDB(raw any) (any, error)
      Type() JFieldType
  }
  ```
* Support converters for:

  * `time.Time <-> Mongo Date`
  * Enums <-> Strings
  * Structs <-> JSON blobs
  * Custom types (`Money`, `Email`, etc.)
* Global `ConverterRegistry` keyed by `JFieldType`.

### 4. Validation Layer

* Field-level (required, regex, min/max, unique).
* Record-level (cross-field validation).
* Schema-level (business rules).
* `JSchema.Validate()` should run all validations.

### 5. Hooks / Interceptors

* Define lifecycle hooks:

  * BeforeSave, AfterSave
  * BeforeDelete, AfterDelete
  * BeforeFind, AfterFind
* Hook signature:

  ```go
  type HookFunc func(ctx context.Context, rec JRecord) error
  ```

### 6. MongoDB Integration

* Implement `MongoRepository` that satisfies `CrudRepository` & `JpaRepository`.
* Map `JRecord` <-> `bson.M` automatically using registered converters.
* Support schema-based indexes (unique, compound).

### 7. Query API

If we want the **Query API** to support **custom operators** (like `IsToday`, `IsWeekend`, `IsWithinBusinessHours`, etc.), we need to extend the design to make operators **pluggable** instead of hardcoded.

---


### 7.1. **Operator Definition**

Instead of making `Op` just an enum, make it an **interface** or at least support custom implementations.

```go
type JOperator interface {
    Name() string
    Apply(field string, value any) (filter any, err error)
}
```

### 7.2. **Default Operators**

Ship built-in operators:

```go
var (
    OpEQ  = NewBasicOperator("EQ", func(field string, value any) any {
        return bson.M{field: value}
    })
    OpGTE = NewBasicOperator("GTE", func(field string, value any) any {
        return bson.M{field: bson.M{"$gte": value}}
    })
    // ... other common ones
)
```

`NewBasicOperator` could be:

```go
func NewBasicOperator(name string, fn func(field string, value any) any) JOperator {
    return &basicOperator{name: name, fn: fn}
}

type basicOperator struct {
    name string
    fn   func(field string, value any) any
}

func (b *basicOperator) Name() string { return b.name }
func (b *basicOperator) Apply(field string, value any) (any, error) {
    return b.fn(field, value), nil
}
```

---

### 7.3. **Custom Operator Example**

For `IsToday`:

```go
var OpIsToday = NewBasicOperator("IsToday", func(field string, _ any) any {
    todayStart := time.Now().Truncate(24 * time.Hour)
    tomorrow := todayStart.Add(24 * time.Hour)
    return bson.M{field: bson.M{
        "$gte": todayStart,
        "$lt":  tomorrow,
    }}
})
```

Then use it:

```go
criteria := JCriteria{
    Filters: []JFilter{
        {Field: "createdAt", Op: OpIsToday},
    },
}
```

---

### 7.4. **JFilter Updated**

Your `JFilter` can now be:

```go
type JFilter struct {
    Field string
    Op    JOperator
    Value any
}
```

When building the Mongo query:

```go
func buildMongoQuery(criteria JCriteria) (bson.M, error) {
    filters := bson.M{}
    for _, f := range criteria.Filters {
        mongoFilter, err := f.Op.Apply(f.Field, f.Value)
        if err != nil {
            return nil, err
        }
        for k, v := range mongoFilter.(bson.M) {
            filters[k] = v
        }
    }
    return filters, nil
}
```

---

### 7.5. **Operator Registry (Optional)**

Allow registering custom operators globally, so users can add new ones easily:

```go
var operatorRegistry = map[string]JOperator{}

func RegisterOperator(op JOperator) {
    operatorRegistry[op.Name()] = op
}

func GetOperator(name string) (JOperator, bool) {
    op, ok := operatorRegistry[name]
    return op, ok
}
```

---

* Built-in operators handle normal cases (`EQ`, `GTE`, `IN`, etc.).
* Developers can register **domain-specific operators** like `IsToday`, `IsWeekend`, or even complex ones like `WithinGeoFence`.


### 8. Projections & Aggregations

* Allow projection of partial fields (`only name, email`).
* Add simple aggregations (`Count`, `GroupBy`).

---

### Example Usage

```go
userSchema := NewJSchema("User").
    AddField("id", JString, nil).
    AddField("name", JString, nil).
    AddField("email", JString, nil).
    AddField("age", JInt, 18).
    AddValidation(func(ctx context.Context, rec JRecord) error {
        if rec.Get("age").(int) < 0 {
            return errors.New("age must be positive")
        }
        return nil
    })

repo := NewMongoRepository(userSchema, mongoClient)

user := NewJRecord().
    Set("name", "Alice").
    Set("email", "alice@example.com")

saved, _ := repo.Save(ctx, user)

results, _ := repo.FindBy(ctx, JCriteria{
    Filters: []JFilter{{Field: "age", Op: GTE, Value: 18}},
})
```


**Goal:** Generate a modular, extensible Go codebase with these interfaces, default implementations, and MongoDB backend support. Organize the code into packages like `schema`, `repository`, `converter`, `validation`, `hooks`, and `mongo`.

