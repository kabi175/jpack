# Contributing to JPack

Thank you for your interest in contributing to JPack! This document provides guidelines and information for contributors.

## Table of Contents

1. [Code of Conduct](#code-of-conduct)
2. [Getting Started](#getting-started)
3. [Development Setup](#development-setup)
4. [Making Changes](#making-changes)
5. [Testing](#testing)
6. [Documentation](#documentation)
7. [Submitting Changes](#submitting-changes)
8. [Code Style](#code-style)
9. [Architecture Guidelines](#architecture-guidelines)

## Code of Conduct

This project adheres to a code of conduct that we expect all contributors to follow. Please be respectful and constructive in all interactions.

### Our Standards

- Use welcoming and inclusive language
- Be respectful of differing viewpoints and experiences
- Gracefully accept constructive criticism
- Focus on what is best for the community
- Show empathy towards other community members

## Getting Started

### Prerequisites

- Go 1.24.2 or higher
- MongoDB 4.0 or higher (for integration tests)
- Git
- A GitHub account

### First Time Setup

1. **Fork the repository** on GitHub
2. **Clone your fork** locally:
   ```bash
   git clone https://github.com/yourusername/jpack.git
   cd jpack
   ```
3. **Add the original repository as upstream**:
   ```bash
   git remote add upstream https://github.com/kabi175/jpack.git
   ```
4. **Install dependencies**:
   ```bash
   go mod download
   ```
5. **Run tests to ensure everything works**:
   ```bash
   go test ./...
   ```

## Development Setup

### Environment Setup

1. **Go Environment**: Ensure Go is properly installed and `GOPATH` is set
2. **MongoDB**: Install MongoDB locally or use Docker:
   ```bash
   docker run -d -p 27017:27017 --name jpack-mongo mongo:latest
   ```
3. **Editor**: Use any editor with Go support (VS Code, GoLand, vim, etc.)

### Project Structure

```
jpack/
├── README.md              # Main documentation
├── EXAMPLES.md           # Usage examples
├── API_REFERENCE.md      # API documentation
├── CHANGELOG.md          # Version history
├── CONTRIBUTING.md       # This file
├── go.mod               # Go module definition
├── go.sum               # Go module checksums
├── jpack.go             # Core interfaces
├── jschema.go           # Schema implementation
├── primitive_types.go   # Built-in field types
├── mongodb.go           # MongoDB integration
├── field_types_test.go  # Field type tests
├── jschema_test.go      # Schema tests
└── mongodb_test.go      # MongoDB integration tests
```

### Code Organization

- **Core interfaces** in `jpack.go`
- **Schema implementation** in `jschema.go`
- **Field types** in `primitive_types.go`
- **Database adapters** in separate files (e.g., `mongodb.go`)
- **Tests** in `*_test.go` files

## Making Changes

### Branching Strategy

1. **Create a feature branch** from `main`:
   ```bash
   git checkout -b feature/your-feature-name
   ```
2. **Keep your branch updated** with upstream:
   ```bash
   git fetch upstream
   git rebase upstream/main
   ```

### Commit Messages

Use conventional commit messages:

```
type(scope): description

[optional body]

[optional footer]
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes
- `refactor`: Code refactoring
- `test`: Adding or modifying tests
- `chore`: Maintenance tasks

Examples:
```
feat(schema): add support for default field values
fix(mongodb): handle nil values in record conversion
docs(readme): update installation instructions
test(field-types): add validation tests for custom types
```

### Code Changes

1. **Follow Go conventions**: Use `gofmt`, `golint`, and `go vet`
2. **Add tests** for new functionality
3. **Update documentation** for API changes
4. **Ensure backwards compatibility** unless making a breaking change

## Testing

### Running Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific test file
go test -v jschema_test.go

# Run integration tests (requires MongoDB)
go test -v mongodb_test.go
```

### Test Categories

1. **Unit Tests**: Test individual components in isolation
2. **Integration Tests**: Test MongoDB integration
3. **Benchmark Tests**: Performance testing

### Writing Tests

#### Unit Tests

```go
func TestYourFunction(t *testing.T) {
    t.Run("positive case", func(t *testing.T) {
        // Arrange
        input := "test input"
        expected := "expected output"
        
        // Act
        result := YourFunction(input)
        
        // Assert
        assert.Equal(t, expected, result)
    })
    
    t.Run("error case", func(t *testing.T) {
        // Test error conditions
        result, err := YourFunction("invalid input")
        assert.Error(t, err)
        assert.Nil(t, result)
    })
}
```

#### Integration Tests

```go
func TestMongoIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }
    
    // Setup MongoDB connection
    client, err := mongo.Connect(options.Client().ApplyURI("mongodb://localhost:27017"))
    require.NoError(t, err)
    defer client.Disconnect(context.TODO())
    
    // Test implementation
}
```

### Test Guidelines

- Write tests for all public APIs
- Test both success and failure cases
- Use table-driven tests for multiple scenarios
- Mock external dependencies in unit tests
- Use descriptive test names
- Clean up resources in tests

## Documentation

### Documentation Types

1. **API Documentation**: Inline GoDoc comments
2. **User Documentation**: README, examples
3. **Developer Documentation**: Architecture, contributing

### GoDoc Guidelines

```go
// YourFunction performs a specific operation on the input.
// It returns the processed result and an error if the operation fails.
//
// Example:
//   result, err := YourFunction("input")
//   if err != nil {
//       return err
//   }
//   fmt.Println(result)
func YourFunction(input string) (string, error) {
    // Implementation
}
```

### Documentation Standards

- Document all public functions, types, and constants
- Include usage examples in documentation
- Keep documentation up to date with code changes
- Use clear, concise language
- Include error conditions and edge cases

## Submitting Changes

### Pull Request Process

1. **Ensure all tests pass**:
   ```bash
   go test ./...
   go vet ./...
   golint ./...
   ```

2. **Update documentation** if needed

3. **Create a pull request** with:
   - Clear title and description
   - Reference to related issues
   - List of changes made
   - Testing information

4. **Respond to review feedback** promptly

### Pull Request Template

```markdown
## Description
Brief description of the changes made.

## Related Issues
- Fixes #123
- Relates to #456

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
- [ ] Unit tests added/updated
- [ ] Integration tests added/updated
- [ ] All tests pass locally

## Checklist
- [ ] Code follows project style guidelines
- [ ] Self-review completed
- [ ] Documentation updated
- [ ] CHANGELOG updated (if applicable)
```

### Review Process

1. **Automated checks** must pass
2. **Code review** by maintainers
3. **Testing** in different environments
4. **Documentation** review if applicable
5. **Merge** once approved

## Code Style

### Go Style Guidelines

Follow the official Go style guide and these additional guidelines:

#### Formatting
- Use `gofmt` to format code
- Use `goimports` to organize imports
- Line length should be reasonable (prefer < 100 characters)

#### Naming
- Use descriptive names for variables and functions
- Follow Go naming conventions (camelCase, PascalCase)
- Use common abbreviations consistently

#### Error Handling
```go
// Good
if err != nil {
    return fmt.Errorf("failed to process data: %w", err)
}

// Bad
if err != nil {
    panic(err)
}
```

#### Interfaces
```go
// Good: Small, focused interfaces
type Reader interface {
    Read([]byte) (int, error)
}

// Bad: Large interfaces with many methods
type Everything interface {
    Read([]byte) (int, error)
    Write([]byte) (int, error)
    Close() error
    Seek(int64, int) (int64, error)
    // ... many more methods
}
```

### Project-Specific Guidelines

#### Interface Design
- Keep interfaces small and focused
- Use composition over inheritance
- Define interfaces at the point of use

#### Error Handling
- Use wrapped errors with context
- Provide meaningful error messages
- Don't ignore errors

#### Testing
- Use table-driven tests for multiple scenarios
- Test both success and failure cases
- Use meaningful test names

## Architecture Guidelines

### Design Principles

1. **Interface-Driven Design**: Define interfaces first, implementations second
2. **Composition over Inheritance**: Use embedding and composition
3. **Dependency Injection**: Use context for dependencies
4. **Immutability**: Prefer immutable data structures where possible

### Adding New Field Types

1. **Implement the `JFieldType` interface**:
   ```go
   type YourFieldType struct {
       // configuration fields
   }
   
   func (y *YourFieldType) Validate(value any) error {
       // validation logic
   }
   
   func (y *YourFieldType) Scan(ctx context.Context, field JField, row map[string]any) (any, error) {
       // scanning logic
   }
   
   func (y *YourFieldType) SetValue(ctx context.Context, field JField, value any, row map[string]any) error {
       // value setting logic
   }
   ```

2. **Add comprehensive tests**
3. **Update documentation**
4. **Add usage examples**

### Database Integration

1. **Create adapter interfaces** for database operations
2. **Implement context-based configuration**
3. **Handle connection lifecycle properly**
4. **Provide comprehensive error handling**

### Performance Considerations

- Use efficient data structures
- Minimize allocations in hot paths
- Consider memory usage for large datasets
- Profile performance-critical code

## Getting Help

### Resources

- **Documentation**: README, API reference, examples
- **Issues**: GitHub issues for bugs and feature requests
- **Discussions**: GitHub discussions for questions

### Contact

- **GitHub Issues**: For bugs and feature requests
- **Email**: For security issues or private concerns
- **Discussions**: For general questions and ideas

## Recognition

Contributors will be recognized in:
- GitHub contributors list
- Release notes for significant contributions
- Project documentation

Thank you for contributing to JPack!
