package main

import (
	"fmt"
	"os"

	"github.com/kabi175/jpack/logger"
)

func main() {
	// Initialize logger
	logger.Initialize()

	// Check if specific example is requested
	if len(os.Args) > 1 {
		example := os.Args[1]
		switch example {
		case "user":
			UserExample()
		case "product":
			ProductExample()
		case "custom":
			CustomConverterExample()
		case "immutable":
			RunImmutableExamples()
		case "lazy":
			LazyRegistryExample()
		case "concurrent":
			ConcurrentExample()
		case "file":
			FileSchemaExample()
		default:
			fmt.Printf("Unknown example: %s\n", example)
			fmt.Println("Available examples: user, product, custom, immutable, lazy, concurrent, file")
			os.Exit(1)
		}
		return
	}

	// Run all examples by default
	fmt.Println("=== Running All Examples ===")

	// Run basic examples
	fmt.Println("\n1. User Example")
	UserExample()

	fmt.Println("\n2. Product Example")
	ProductExample()

	fmt.Println("\n3. Custom Converter Example")
	CustomConverterExample()

	// Run immutable schema examples
	fmt.Println("\n4. Immutable Schema Examples")
	RunImmutableExamples()

	// Run lazy registry examples
	fmt.Println("\n5. Lazy Registry Example")
	LazyRegistryExample()

	// Run concurrent example
	fmt.Println("\n6. Concurrent Example")
	ConcurrentExample()

	// Run file schema example
	fmt.Println("\n7. File Schema Example")
	FileSchemaExample()

	fmt.Println("\n=== All Examples Completed Successfully! ===")
}
