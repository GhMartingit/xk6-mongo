# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]

### Added

#### Connection Management
- **Connection verification with ping**: Clients now verify MongoDB connection with a ping during initialization
- **Connection timeout**: 10-second timeout for connection establishment
- **Operation timeout**: 30-second default timeout for all database operations
- **Graceful disconnection**: Proper timeout handling during disconnect

#### Input Validation
- Database and collection name validation (checks for invalid characters)
- Nil/empty input validation for all operations:
  - Documents cannot be nil
  - Filters cannot be nil for update/delete operations
  - Document arrays cannot be empty
  - Pipeline cannot be nil for aggregations
  - Limit cannot be negative for find operations
  - Field names cannot be empty for distinct operations

#### New Operations
- **BulkWrite**: Execute multiple write operations (inserts, updates, deletes) in a single call
- **FindWithOptions**: Advanced find with support for:
  - Batch size control
  - Projection (select specific fields)
  - Skip (pagination)
  - Limit and sort

#### Error Handling
- Consistent error message constants
- Structured error variables for common validation errors
- Better error logging throughout

#### Testing
- Comprehensive validation test suite (`validation_test.go`)
- Tests for all input validation scenarios
- Tests for utility functions (toPascalCase, validateDatabaseAndCollection)
- Unit tests covering error cases
- **Performance benchmarks** (`benchmark_test.go`):
  - 23 benchmark tests covering critical functions
  - Memory allocation tracking
  - Input validation: ~24 ns/op with 0 allocations
  - PascalCase conversion: ~230 ns/op
  - Update detection: <3 ns/op
  - Pipeline detection: <1 ns/op

#### CI/CD
- GitHub Actions workflow for automated testing
- Multi-version Go testing (1.22, 1.23, 1.24)
- MongoDB service container for integration tests
- golangci-lint integration
- Code coverage reporting with Codecov
- Automated build verification

#### Documentation
- Removed duplicate content from README
- Added configuration section with timeout and connection pooling examples
- Added API reference section
- Added performance tips section
- Added error handling examples
- Added advanced find and bulk write examples
- Added CHANGELOG file
- **CONTRIBUTING.md**: Complete contribution guidelines with:
  - Code of conduct
  - Development setup instructions
  - Testing guidelines
  - PR process documentation
  - Coding standards
- **SECURITY.md**: Security best practices including:
  - Connection security (TLS/SSL)
  - Authentication mechanisms
  - Data protection guidelines
  - Testing considerations
  - Compliance considerations (GDPR, PCI DSS, HIPAA)
- **Complex aggregation example**: Multi-stage pipeline with $lookup, $group, $project

### Changed

#### Code Quality
- Extracted repeated database/collection access into `getCollection` helper
- Extracted context creation into `getContext` helper
- Centralized error message constants
- All operations now use context with timeout (no more `context.Background()`)
- Proper cursor closure with deferred `Close()` calls

#### Client Structure
- Client now stores `defaultTimeout` for configurable operation timeouts
- Better separation of concerns in client initialization
- **Retry support**: Added `retryWrites` and `retryReads` fields for transient failure handling

### Fixed

- Connection failures now properly disconnect and clean up resources
- Cursor leaks prevented with deferred Close() calls
- Invalid database names are now rejected before MongoDB operations
- Race conditions eliminated with proper context timeout handling

## Performance Improvements

- Connection pooling configuration exposed via client options
- Batch size control for large result sets
- Projection support to reduce data transfer
- Bulk operations for reduced network round trips

## Breaking Changes

None - all changes are backward compatible additions and improvements.

## Migration Guide

### For Existing Users

All existing code will continue to work. New features are opt-in:

**Before:**
```js
const results = client.find("db", "col", {}, null, 100);
```

**After (with new options):**
```js
const results = client.findWithOptions("db", "col", {}, {
    limit: 100,
    batch_size: 50,
    projection: { field1: 1, field2: 1 }
});
```

**Connection pooling:**
```js
// New: Configure connection pool
const options = {
    max_pool_size: 100,
    min_pool_size: 10
};
const client = xk6_mongo.newClientWithOptions(uri, options);
```

## Contributors

- Martin Ghazaryan - Initial implementation and improvements
