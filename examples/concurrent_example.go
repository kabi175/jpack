package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/kabi175/jpack/schema"
)

// ConcurrentExample demonstrates thread safety of the lazy-loaded registry
func ConcurrentExample() {
	// Create a new schema registry
	registry := schema.NewJSchemaRegistry()

	// Create a simple file source for testing
	fileSource := &TestSchemaSource{}

	// Register a lazy schema
	err := registry.RegisterLazy("TestSchema", fileSource)
	if err != nil {
		panic(err)
	}

	// Test concurrent access
	var wg sync.WaitGroup
	numGoroutines := 10

	fmt.Printf("Testing concurrent access with %d goroutines...\n", numGoroutines)

	start := time.Now()

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Each goroutine tries to get the schema
			schema, ok := registry.Get("TestSchema")
			if ok {
				fmt.Printf("Goroutine %d: Got schema %s\n", id, schema.Name())
			} else {
				fmt.Printf("Goroutine %d: Failed to get schema\n", id)
			}
		}(i)
	}

	wg.Wait()
	elapsed := time.Since(start)

	fmt.Printf("Concurrent test completed in %v\n", elapsed)
	fmt.Printf("Schema was loaded %d times (should be 1)\n", fileSource.loadCount)
}

// TestSchemaSource is a simple test implementation
type TestSchemaSource struct {
	loadCount int
}

func (t *TestSchemaSource) Load(schemaName string) (schema.JSchema, error) {
	t.loadCount++
	fmt.Printf("Loading schema %s (count: %d)\n", schemaName, t.loadCount)

	// Simulate some loading time
	time.Sleep(10 * time.Millisecond)

	// Create a simple schema
	s := schema.NewJSchema(schemaName)
	s.AddField("id", schema.JString, nil)
	s.AddField("name", schema.JString, nil)

	return s, nil
}
