# Report System Tests

This directory contains comprehensive tests for the unified report system, including unit tests, integration tests, and frontend service tests.

## Test Structure

### Backend Tests

1. **`unified_report_test.go`** - Unit tests for the unified report controller
   - API endpoint validation
   - Parameter validation and conversion
   - Error handling
   - Response structure validation
   - Different output formats
   - Authentication and authorization
   - Performance benchmarks

2. **`integration_report_test.go`** - Integration tests for the complete system
   - End-to-end report generation flow
   - Database integration
   - All report types with real data
   - Preview functionality
   - Error scenarios
   - Data integrity checks
   - Concurrent request handling
   - Performance testing

### Frontend Tests

3. **`frontend/src/services/__tests__/reportService.test.ts`** - Frontend service tests
   - ReportService functionality
   - API parameter conversion
   - Error handling
   - Response parsing
   - Download functionality
   - Mock API testing

## Running Tests

### Backend Tests

#### Prerequisites

```bash
# Install required Go packages
go mod tidy

# Install testify for testing framework
go get github.com/stretchr/testify
```

#### Run All Tests

```bash
# Run all tests
go test ./tests/...

# Run with verbose output
go test -v ./tests/...

# Run with coverage
go test -cover ./tests/...

# Generate coverage report
go test -coverprofile=coverage.out ./tests/...
go tool cover -html=coverage.out -o coverage.html
```

#### Run Specific Test Files

```bash
# Run unified report tests only
go test -v ./tests -run TestUnified

# Run integration tests only
go test -v ./tests -run TestReportIntegrationSuite

# Run specific test function
go test -v ./tests -run TestUnifiedReportEndpoints
```

#### Run Benchmarks

```bash
# Run all benchmarks
go test -bench=. ./tests/...

# Run specific benchmark
go test -bench=BenchmarkBalanceSheet ./tests/...

# Run benchmarks with memory profiling
go test -bench=. -benchmem ./tests/...
```

### Frontend Tests

#### Prerequisites

```bash
# Install dependencies
npm install

# Install Jest and testing utilities
npm install --save-dev @jest/globals jest @types/jest
```

#### Run Frontend Tests

```bash
# Run all tests
npm test

# Run tests in watch mode
npm test -- --watch

# Run with coverage
npm test -- --coverage

# Run specific test file
npm test reportService.test.ts
```

## Test Coverage

The test suite covers the following areas:

### API Endpoints
- ✅ All report type endpoints (`/api/reports/{type}`)
- ✅ Preview endpoints (`/api/reports/preview/{type}`)
- ✅ Available reports endpoint (`/api/reports/available`)
- ✅ Legacy compatibility routes
- ✅ Parameter validation for all endpoints

### Report Types
- ✅ Balance Sheet
- ✅ Profit & Loss Statement
- ✅ Cash Flow Statement
- ✅ Trial Balance
- ✅ General Ledger
- ✅ Sales Summary
- ✅ Vendor Analysis

### Output Formats
- ✅ JSON responses
- ✅ PDF file downloads
- ✅ Excel file downloads
- ✅ CSV file downloads

### Error Handling
- ✅ Invalid report types
- ✅ Missing required parameters
- ✅ Invalid date formats
- ✅ Invalid date ranges
- ✅ Authentication errors
- ✅ Authorization errors
- ✅ Network errors

### Response Structure
- ✅ Standard response format validation
- ✅ Metadata structure validation
- ✅ Error response format validation
- ✅ Custom headers validation

### Performance
- ✅ Response time validation
- ✅ Concurrent request handling
- ✅ Memory usage benchmarks
- ✅ Load testing

### Data Integrity
- ✅ Cross-report data consistency
- ✅ Database transaction integrity
- ✅ Parameter parsing accuracy

## Test Data

### Database Test Data
The integration tests use an in-memory SQLite database with the following test data:

- **Company**: Test Company Ltd
- **Users**: Test user with admin role
- **Accounts**: Complete chart of accounts (Assets, Liabilities, Equity, Revenue, Expenses)
- **Contacts**: Sample customers and suppliers
- **Products**: Test products with pricing
- **Sales**: Sample sales transactions
- **Purchases**: Sample purchase transactions
- **Journal Entries**: Corresponding accounting entries

### Mock Data
Unit tests use mocked data and repositories to isolate functionality testing.

## Test Configuration

### Environment Variables
```bash
# Test database configuration
TEST_DB_DRIVER=sqlite
TEST_DB_DSN=:memory:

# Test server configuration
TEST_SERVER_PORT=8080
TEST_JWT_SECRET=test-secret

# Test timeouts
TEST_TIMEOUT=30s
TEST_DB_TIMEOUT=5s
```

### Test Flags

```bash
# Skip integration tests (only unit tests)
go test -short ./tests/...

# Run only integration tests
go test -run Integration ./tests/...

# Enable race condition detection
go test -race ./tests/...

# Set test timeout
go test -timeout 30s ./tests/...
```

## Continuous Integration

### GitHub Actions Configuration

```yaml
name: Report System Tests

on: [push, pull_request]

jobs:
  backend-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v3
        with:
          go-version: '1.21'
      - run: go mod download
      - run: go test -v -cover ./tests/...
      - run: go test -bench=. ./tests/...

  frontend-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-node@v3
        with:
          node-version: '18'
      - run: npm ci
      - run: npm test -- --coverage
```

## Debugging Tests

### Common Issues

1. **Database Connection Errors**
   ```bash
   # Check if test database is properly initialized
   go test -v ./tests -run TestIntegration
   ```

2. **Mock Setup Issues**
   ```bash
   # Verify mock configurations
   go test -v ./tests -run TestMock
   ```

3. **Timing Issues**
   ```bash
   # Run with race detection
   go test -race ./tests/...
   ```

### Debug Mode
```bash
# Enable debug logging during tests
export DEBUG=true
go test -v ./tests/...

# Run single test with full output
go test -v ./tests -run TestSpecificFunction
```

## Performance Benchmarks

Expected performance metrics:

- **Report Generation**: < 500ms for typical reports
- **Concurrent Requests**: Handle 50+ simultaneous requests
- **Memory Usage**: < 50MB per report generation
- **Database Queries**: < 100ms for data retrieval

### Running Performance Tests
```bash
# Memory profiling
go test -bench=BenchmarkReportGeneration -memprofile=mem.prof ./tests/...
go tool pprof mem.prof

# CPU profiling
go test -bench=BenchmarkReportGeneration -cpuprofile=cpu.prof ./tests/...
go tool pprof cpu.prof

# Load testing
go test -bench=BenchmarkConcurrentReportGeneration -benchtime=10s ./tests/...
```

## Contributing

When adding new tests:

1. Follow the existing test structure and naming conventions
2. Include both unit and integration tests for new functionality
3. Add appropriate mocks for external dependencies
4. Ensure tests are deterministic and can run in any order
5. Update this README with any new test requirements

### Test Naming Conventions

- Test functions: `Test{ComponentName}{Functionality}`
- Benchmark functions: `Benchmark{ComponentName}{Operation}`
- Test suites: `{ComponentName}TestSuite`

### Test Organization

- Group related tests using subtests (`t.Run()`)
- Use table-driven tests for multiple similar scenarios
- Keep setup and teardown in appropriate lifecycle methods
- Use descriptive test names and error messages
