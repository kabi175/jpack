Create a **lazy-loaded schema registry in Go** for a schema-driven persistence framework.

### Requirements

#### 1. `JSchemaRegistry` Interface

```go
type JSchemaRegistry interface {
    Register(schema JSchema) error
    RegisterLazy(schemaName string, datasource ExternalSchemaSource) error
    Get(name string) (JSchema, bool)
    List() []JSchema
}
```

#### 2. `ExternalSchemaSource`

* Represents a source for schema definitions.
* Should support loading schema from a file, HTTP service, DB, or dynamically generated.

```go
type ExternalSchemaSource interface {
    Load(schemaName string) (JSchema, error)
}
```

#### 3. Implementation Details

* Internally store schemas in a `map[string]*lazyEntry`.
* `lazyEntry` should use `sync.Once` to ensure schema is only loaded once (lazy load).
* `Register(schema JSchema)` → immediately stores the schema.
* `RegisterLazy(schemaName string, datasource ExternalSchemaSource)` → defers loading until first `Get(name)` call.
* `Get(name string)` →

  * If schema exists and is already loaded, return it.
  * If it was registered lazily, call `datasource.Load(schemaName)` once.
  * Thread-safe.
* `List()` → returns all loaded schemas (only those that have been actually loaded).

#### 4. Example Usage

```go
registry := NewSchemaRegistry()

// Eager registration
userSchema := NewJSchema("User").
    AddField("id", JString, nil).
    AddField("email", JString, nil)
registry.Register(userSchema)

// Lazy registration
registry.RegisterLazy("Order", &FileSchemaSource{Path: "order_schema.json"})

// First Get will trigger lazy load
orderSchema, ok := registry.Get("Order")
if ok {
    fmt.Println("Loaded schema:", orderSchema.Name())
}
```

#### 5. `FileSchemaSource` Example Implementation

* Loads schema definition from JSON file.

```go
type FileSchemaSource struct {
    Path string
}

func (f *FileSchemaSource) Load(schemaName string) (JSchema, error) {
    data, err := os.ReadFile(f.Path)
    if err != nil {
        return nil, err
    }
    var def SchemaDef
    if err := json.Unmarshal(data, &def); err != nil {
        return nil, err
    }
    return ConvertDefToSchema(def), nil
}
```

---

**Goal:**
Generate a **thread-safe, lazy-loaded schema registry** that supports both eager and lazy schema registration, with pluggable `ExternalSchemaSource` implementations (file, HTTP, DB, etc.).