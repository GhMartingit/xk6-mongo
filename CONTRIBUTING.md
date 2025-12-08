# Contributing to xk6-mongo

Thank you for your interest in contributing to xk6-mongo! This document provides guidelines and instructions for contributing.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Making Changes](#making-changes)
- [Testing](#testing)
- [Submitting Changes](#submitting-changes)
- [Coding Standards](#coding-standards)
- [Performance Considerations](#performance-considerations)

## Code of Conduct

This project adheres to a code of conduct. By participating, you are expected to uphold this code:

- Be respectful and inclusive
- Welcome newcomers and help them get started
- Focus on what is best for the community
- Show empathy towards other community members

## Getting Started

1. **Fork the repository** on GitHub
2. **Clone your fork** locally:
   ```bash
   git clone https://github.com/YOUR-USERNAME/xk6-mongo.git
   cd xk6-mongo
   ```
3. **Add upstream remote**:
   ```bash
   git remote add upstream https://github.com/GhMartingit/xk6-mongo.git
   ```

## Development Setup

### Prerequisites

- Go 1.22 or later
- Docker (for running MongoDB during tests)
- xk6 CLI tool
- golangci-lint (for linting)

### Install Dependencies

```bash
go mod download
```

### Start MongoDB for Testing

```bash
docker run -d --rm --name xk6-mongo-dev -p 27017:27017 mongo:7
```

### Build the Extension

```bash
make build
# or
xk6 build --with github.com/GhMartingit/xk6-mongo=.
```

## Making Changes

### Branch Naming

Use descriptive branch names:
- `feature/add-transaction-support`
- `fix/connection-leak`
- `docs/update-readme`
- `test/add-benchmark-tests`

### Commit Messages

Follow the conventional commits specification:

```
type(scope): subject

body

footer
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `test`: Adding or updating tests
- `refactor`: Code refactoring
- `perf`: Performance improvements
- `chore`: Maintenance tasks

**Examples:**
```
feat(client): add retry logic for transient failures

Implements configurable retry logic with exponential backoff
for network errors and transient MongoDB failures.

Closes #123
```

```
fix(validation): prevent panic on nil filter

Added nil check before filter validation to prevent
panic when nil filter is passed to update operations.
```

## Testing

### Running Tests

```bash
# Run all tests
go test -v ./...

# Run tests with coverage
go test -v -coverprofile=coverage.txt ./...

# Run specific tests
go test -v -run TestValidation ./...

# Run benchmarks
go test -bench=. -benchmem ./...
```

### Writing Tests

All new features must include tests:

1. **Unit Tests**: Test individual functions
2. **Integration Tests**: Test with real MongoDB
3. **Validation Tests**: Test error cases
4. **Benchmarks**: For performance-critical code

**Example Test:**
```go
func TestNewFeature(t *testing.T) {
    t.Run("successful case", func(t *testing.T) {
        // Arrange
        client := &Client{}

        // Act
        result, err := client.NewMethod("db", "col")

        // Assert
        if err != nil {
            t.Fatalf("unexpected error: %v", err)
        }
        if result == nil {
            t.Error("expected non-nil result")
        }
    })

    t.Run("error case", func(t *testing.T) {
        // Test error handling
    })
}
```

### Integration Tests

Integration tests require MongoDB:

```bash
# Set MongoDB URI
export MONGODB_URI=mongodb://localhost:27017

# Run integration tests
go test -v ./...
```

## Submitting Changes

### Before Submitting

1. **Run tests**: Ensure all tests pass
   ```bash
   go test -v ./...
   ```

2. **Run linter**: Fix any linting issues
   ```bash
   golangci-lint run
   ```

3. **Format code**: Ensure code is properly formatted
   ```bash
   go fmt ./...
   ```

4. **Update documentation**: Update README if needed

5. **Add tests**: Ensure new code has test coverage

### Pull Request Process

1. **Update your fork**:
   ```bash
   git fetch upstream
   git rebase upstream/master
   ```

2. **Push to your fork**:
   ```bash
   git push origin your-branch-name
   ```

3. **Create Pull Request** on GitHub with:
   - Clear title describing the change
   - Description of what changed and why
   - Link to related issues
   - Screenshots (if UI-related)

4. **PR Description Template**:
   ```markdown
   ## Summary
   Brief description of the changes

   ## Changes
   - Added X
   - Fixed Y
   - Updated Z

   ## Testing
   - [ ] Unit tests added/updated
   - [ ] Integration tests added/updated
   - [ ] Benchmarks added (if performance-related)
   - [ ] Manual testing completed

   ## Related Issues
   Closes #123
   Related to #456
   ```

5. **Respond to feedback**: Address review comments promptly

## Coding Standards

### Go Style Guide

Follow the official [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments):

1. **Formatting**: Use `gofmt`
2. **Naming**: Use idiomatic Go names
   - Exported names: `ClientOptions`, `NewClient`
   - Unexported names: `validateInput`, `getCollection`
3. **Error handling**: Return errors, don't panic
4. **Documentation**: Document all exported functions

### Code Structure

```go
// Good: Clear function with validation and error handling
func (c *Client) Insert(database, collection string, doc any) error {
    if doc == nil {
        return errDocumentNil
    }

    col, err := c.getCollection(database, collection)
    if err != nil {
        log.Printf(errValidatingCollection, err)
        return err
    }

    ctx, cancel := c.getContext()
    defer cancel()

    _, err = col.InsertOne(ctx, doc)
    if err != nil {
        log.Printf(errInsertingDocument, err)
        return err
    }

    return nil
}
```

### Documentation

- Document all exported types, functions, and constants
- Use complete sentences
- Include examples for complex functionality

```go
// FindWithOptions provides advanced find options including batch size control.
// It accepts a map of options that can include:
//   - limit (int64): Maximum number of documents to return
//   - skip (int64): Number of documents to skip
//   - batch_size (int32): Number of documents per batch
//   - sort: Sort order specification
//   - projection: Field projection specification
//
// Example:
//   options := map[string]any{
//       "limit": 100,
//       "batch_size": 50,
//       "sort": bson.M{"createdAt": -1},
//   }
//   results, err := client.FindWithOptions("db", "col", filter, options)
func (c *Client) FindWithOptions(database, collection string, filter any, options map[string]any) ([]bson.M, error) {
    // Implementation
}
```

## Performance Considerations

When contributing performance-critical code:

1. **Add benchmarks**: Use `go test -bench`
2. **Profile memory**: Check allocations with `-benchmem`
3. **Avoid allocations**: Reuse objects when possible
4. **Use efficient data structures**: Choose appropriate types

**Example Benchmark:**
```go
func BenchmarkNewFeature(b *testing.B) {
    b.Run("common_case", func(b *testing.B) {
        b.ReportAllocs()
        for i := 0; i < b.N; i++ {
            // Code to benchmark
        }
    })
}
```

## Questions?

- Open an issue for bugs or feature requests
- Start a discussion for questions
- Check existing issues and PRs first

## License

By contributing, you agree that your contributions will be licensed under the same license as the project.

---

Thank you for contributing to xk6-mongo! 🎉
