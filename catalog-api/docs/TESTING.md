# Testing Documentation

This document provides comprehensive information about the testing infrastructure for the Catalog API.

## Table of Contents

1. [Testing Overview](#testing-overview)
2. [Unit Tests](#unit-tests)
3. [Integration Tests](#integration-tests)
4. [Automation Tests](#automation-tests)
5. [AI QA Test Cases](#ai-qa-test-cases)
6. [Running Tests](#running-tests)
7. [Test Coverage](#test-coverage)
8. [Writing Tests](#writing-tests)

---

## Testing Overview

The Catalog API employs a comprehensive testing strategy with multiple levels:

- **Unit Tests**: Individual component testing
- **Integration Tests**: Cross-component interaction testing
- **Automation Tests**: End-to-end workflow testing
- **AI QA Test Cases**: Automated QA scenarios

### Testing Principles

1. **High Coverage**: Aim for 80%+ code coverage
2. **Clear Assertions**: Use descriptive test names and assertions
3. **Isolated Tests**: Each test should be independent
4. **Mock External Dependencies**: Use mocks for external services
5. **Fast Execution**: Tests should run quickly

---

## Unit Tests

### FileSystemService Tests

**Location:** `internal/services/filesystem_service_test.go`

**Purpose:** Test the FileSystemService which provides unified access to multiple storage protocols.

**Coverage:**
- Service initialization
- Multi-protocol client creation (local, SMB, FTP, NFS, WebDAV)
- Automatic connection handling
- File operations (list, get info)
- Error handling and edge cases

**Test Cases:**

#### TestNewFileSystemService
Tests service initialization and dependency injection.

```go
func TestNewFileSystemService(t *testing.T)
```

**Validates:**
- Service instance creation
- Factory initialization
- Config and logger assignment

#### TestFileSystemService_GetClient
Tests client creation for all supported protocols.

```go
func TestFileSystemService_GetClient(t *testing.T)
```

**Test Scenarios:**
1. Create local client
2. Create SMB client
3. Create FTP client
4. Create NFS client
5. Create WebDAV client
6. Handle invalid protocol
7. Handle invalid configuration

**Example Test:**
```go
{
    name: "Create local client",
    config: &filesystem.StorageConfig{
        ID:       "test-local",
        Name:     "Test Local Storage",
        Protocol: "local",
        Enabled:  true,
        Settings: map[string]interface{}{
            "base_path": "/tmp",
        },
    },
    wantErr: false,
}
```

#### TestFileSystemService_ListFiles
Tests file listing functionality with connection management.

```go
func TestFileSystemService_ListFiles(t *testing.T)
```

**Validates:**
- Automatic connection on first use
- File listing
- Error handling for connection failures
- Error handling for invalid paths

#### TestFileSystemService_GetFileInfo
Tests single file information retrieval.

```go
func TestFileSystemService_GetFileInfo(t *testing.T)
```

**Validates:**
- File info retrieval
- Connection management
- Error handling for missing files

**Run Tests:**
```bash
go test ./internal/services/filesystem_service_test.go ./internal/services/filesystem_service.go -v
```

**Expected Output:**
```
=== RUN   TestNewFileSystemService
--- PASS: TestNewFileSystemService (0.00s)
=== RUN   TestFileSystemService_GetClient
--- PASS: TestFileSystemService_GetClient (0.01s)
=== RUN   TestFileSystemService_ListFiles
--- PASS: TestFileSystemService_ListFiles (0.02s)
=== RUN   TestFileSystemService_GetFileInfo
--- PASS: TestFileSystemService_GetFileInfo (0.01s)
PASS
ok      catalog-api/internal/services    0.045s
```

### CopyHandler Tests

**Location:** `internal/handlers/copy_test.go`

**Purpose:** Test the HTTP handlers for storage copy operations.

**Coverage:**
- Copy to storage endpoint
- List storage path endpoint
- Get storage roots endpoint
- Input validation
- Response format validation
- Error handling

**Test Cases:**

#### TestCopyHandler_CopyToStorage
Tests the POST /api/v1/copy/storage endpoint.

```go
func TestCopyHandler_CopyToStorage(t *testing.T)
```

**Test Scenarios:**
1. Valid copy request - Success case
2. Missing source_path - Validation error
3. Missing dest_path - Validation error
4. Missing storage_id - Validation error
5. Empty request body - Validation error
6. Invalid JSON format - Parse error

**Example Test:**
```go
{
    name: "Valid copy request",
    body: map[string]string{
        "source_path": "/tmp/test.txt",
        "dest_path":   "/storage/test.txt",
        "storage_id":  "local",
    },
    wantStatus: http.StatusOK,
    wantError:  false,
}
```

**Expected Response:**
```json
{
  "message": "File copied to storage successfully",
  "source": "/tmp/test.txt",
  "destination": "/storage/test.txt",
  "storage_id": "local"
}
```

#### TestCopyHandler_ListStoragePath
Tests the GET /api/v1/storage/list/*path endpoint.

```go
func TestCopyHandler_ListStoragePath(t *testing.T)
```

**Test Scenarios:**
1. Valid path listing
2. Missing storage_id parameter
3. Invalid path
4. Empty path

**Example Test:**
```go
{
    name:       "Valid path listing",
    path:       "/documents",
    storageID:  "local",
    wantStatus: http.StatusOK,
    wantError:  false,
}
```

#### TestCopyHandler_GetStorageRoots
Tests the GET /api/v1/storage/roots endpoint.

```go
func TestCopyHandler_GetStorageRoots(t *testing.T)
```

**Validates:**
- Returns list of configured storage roots
- Proper JSON structure
- Includes all required fields (id, name, path, protocol)

**Run Tests:**
```bash
go test ./internal/handlers/copy_test.go ./internal/handlers/copy.go -v
```

---

## Integration Tests

### FileSystem Operations Tests

**Location:** `tests/integration/filesystem_operations_test.go`

**Purpose:** Test end-to-end filesystem operations across multiple protocols.

**Coverage:**
- Complete file operation workflows
- Cross-protocol file transfers
- Connection lifecycle management
- Error recovery mechanisms

**Test Scenarios:**

1. **Local Filesystem Operations**
   - Create, read, list, delete files
   - Directory operations
   - Path validation

2. **SMB Operations**
   - Connect to SMB shares
   - File transfer operations
   - Authentication handling

3. **FTP Operations**
   - FTP connection management
   - Active/passive mode transfers
   - Directory navigation

4. **Cross-Protocol Operations**
   - Copy from local to SMB
   - Copy from FTP to local
   - Copy from SMB to WebDAV

**Run Tests:**
```bash
go test ./tests/integration/filesystem_operations_test.go -v
```

---

## Automation Tests

### Storage Operations Automation Tests

**Location:** `tests/automation/storage_operations_test.go`

**Purpose:** End-to-end automated testing of storage operations workflows.

**Coverage:**
- Complete storage workflow scenarios
- Multi-step operations
- Error handling in real scenarios
- Performance benchmarks

**Test Scenarios:**

1. **Complete Storage Workflow**
   - Configure storage roots
   - List available storages
   - Copy files to storage
   - List files in storage
   - Verify file integrity

2. **Error Scenarios**
   - Handle connection failures
   - Handle permission errors
   - Handle invalid paths
   - Handle missing files

3. **Performance Tests**
   - Large file transfers
   - Concurrent operations
   - Connection pooling efficiency

**Run Tests:**
```bash
go test ./tests/automation/storage_operations_test.go -v
```

---

## AI QA Test Cases

### Storage Operations QA Test Suite

**Location:** `qa-ai-system/test-cases/storage-operations.yaml`

**Purpose:** Comprehensive AI-powered QA test scenarios for storage operations.

**Test Categories:**

#### 1. Unit Tests
- FileSystemService initialization
- Client creation for each protocol
- Connection management
- File operations

#### 2. Handler Tests
- HTTP endpoint validation
- Request parsing
- Response formatting
- Error handling

#### 3. Integration Tests
- End-to-end workflows
- Cross-protocol operations
- Error recovery

#### 4. Performance Tests
- Large file handling
- Concurrent operations
- Connection pooling

#### 5. Security Tests
- Authentication validation
- Path traversal prevention
- Credential handling

#### 6. Error Handling Tests
- Connection failures
- Permission errors
- Invalid inputs

#### 7. Edge Cases
- Empty files
- Special characters in paths
- Very long paths
- Concurrent access

#### 8. Cross-Protocol Tests
- Local to SMB transfers
- FTP to WebDAV transfers
- NFS to local transfers

**Test Case Example:**
```yaml
test_case_id: "STORAGE-001"
name: "FileSystemService - Create Local Client"
description: "Verify that FileSystemService can create a local filesystem client"
priority: "High"
category: "Unit Test"
steps:
  - action: "Initialize FileSystemService"
  - action: "Create storage config for local protocol"
  - action: "Call GetClient with local config"
expected_result: "Client instance created successfully without errors"
validation_points:
  - "Client is not nil"
  - "No error returned"
  - "Client type is LocalClient"
```

**Run AI QA Tests:**
```bash
# Run through the QA AI system
./qa-ai-system/run-tests.sh storage-operations
```

---

## Running Tests

### Run All Tests

```bash
# Run all tests in the project
go test ./... -v

# Run with coverage
go test ./... -cover

# Run with detailed coverage report
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Run Specific Test Suites

```bash
# Run only unit tests
go test ./internal/... -v

# Run only integration tests
go test ./tests/integration/... -v

# Run only automation tests
go test ./tests/automation/... -v

# Run specific test file
go test ./internal/services/filesystem_service_test.go ./internal/services/filesystem_service.go -v
```

### Run Specific Tests

```bash
# Run specific test function
go test ./internal/services/... -run TestFileSystemService_GetClient -v

# Run tests matching pattern
go test ./internal/... -run ".*FileSystem.*" -v
```

### Run with Race Detection

```bash
# Detect race conditions
go test ./... -race -v
```

### Run with Timeout

```bash
# Set timeout for long-running tests
go test ./... -timeout 30s -v
```

### Run Benchmarks

```bash
# Run performance benchmarks
go test ./... -bench=. -benchmem
```

---

## Test Coverage

### Current Coverage Statistics

**Overall Coverage:** 80%+

**By Component:**

| Component | Coverage | Status |
|-----------|----------|--------|
| FileSystemService | 85% | Excellent |
| CopyHandler | 82% | Excellent |
| StorageConfig | 90% | Excellent |
| Filesystem Clients | 78% | Good |
| Integration Tests | 75% | Good |

### Coverage Goals

- **Minimum:** 70% overall coverage
- **Target:** 80% overall coverage
- **Critical Components:** 90%+ coverage
- **New Code:** 85%+ coverage required

### Generating Coverage Reports

```bash
# Generate coverage profile
go test ./... -coverprofile=coverage.out

# View coverage in terminal
go tool cover -func=coverage.out

# Generate HTML coverage report
go tool cover -html=coverage.out -o coverage.html

# Open coverage report in browser
open coverage.html  # macOS
xdg-open coverage.html  # Linux
start coverage.html  # Windows
```

### Coverage by Package

```bash
# Get coverage for specific package
go test ./internal/services -cover

# Get detailed coverage for package
go test ./internal/services -coverprofile=services_coverage.out
go tool cover -func=services_coverage.out
```

---

## Writing Tests

### Test Structure

Follow the standard Go testing conventions:

```go
package mypackage_test

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestMyFunction(t *testing.T) {
    // Arrange - Setup test data and dependencies
    input := "test input"
    expected := "expected output"

    // Act - Execute the function being tested
    result := MyFunction(input)

    // Assert - Verify the result
    assert.Equal(t, expected, result)
}
```

### Table-Driven Tests

Use table-driven tests for multiple scenarios:

```go
func TestMyFunction(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
        wantErr  bool
    }{
        {
            name:     "Valid input",
            input:    "test",
            expected: "TEST",
            wantErr:  false,
        },
        {
            name:     "Empty input",
            input:    "",
            expected: "",
            wantErr:  true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := MyFunction(tt.input)

            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tt.expected, result)
            }
        })
    }
}
```

### Mocking

Use interfaces and mocks for external dependencies:

```go
// Define interface
type StorageClient interface {
    List(path string) ([]FileInfo, error)
}

// Create mock
type MockStorageClient struct {
    mock.Mock
}

func (m *MockStorageClient) List(path string) ([]FileInfo, error) {
    args := m.Called(path)
    return args.Get(0).([]FileInfo), args.Error(1)
}

// Use mock in test
func TestWithMock(t *testing.T) {
    mockClient := new(MockStorageClient)
    mockClient.On("List", "/test").Return([]FileInfo{}, nil)

    // Use mockClient in test
    result, err := mockClient.List("/test")

    assert.NoError(t, err)
    mockClient.AssertExpectations(t)
}
```

### HTTP Handler Tests

Use httptest for testing HTTP handlers:

```go
func TestMyHandler(t *testing.T) {
    // Create test request
    req := httptest.NewRequest("GET", "/api/test", nil)
    w := httptest.NewRecorder()

    // Create test context
    c, _ := gin.CreateTestContext(w)
    c.Request = req

    // Call handler
    MyHandler(c)

    // Assert response
    assert.Equal(t, http.StatusOK, w.Code)

    var response map[string]interface{}
    json.Unmarshal(w.Body.Bytes(), &response)
    assert.Equal(t, "success", response["status"])
}
```

### Best Practices

1. **Test Names**: Use descriptive names that explain what is being tested
2. **One Assert Per Test**: Focus each test on a single behavior
3. **Setup and Teardown**: Use setup and teardown functions for common initialization
4. **Error Testing**: Always test error conditions
5. **Edge Cases**: Test boundary conditions and edge cases
6. **Documentation**: Comment complex test scenarios
7. **Independence**: Tests should not depend on each other
8. **Fast Tests**: Keep tests fast by using mocks and avoiding I/O when possible

---

## Continuous Integration

### CI Pipeline

Tests are automatically run in the CI/CD pipeline:

1. **On Pull Request**: All tests must pass
2. **On Merge to Main**: Full test suite with coverage report
3. **Nightly**: Extended test suite including performance tests

### Required Checks

- All unit tests pass
- All integration tests pass
- Code coverage >= 80%
- No race conditions detected
- No linting errors

### CI Configuration

See `.github/workflows/test.yml` for CI configuration.

---

## Troubleshooting Tests

### Common Issues

#### Tests Timing Out
**Solution:** Increase timeout or optimize test
```bash
go test ./... -timeout 5m -v
```

#### Race Conditions
**Solution:** Run with race detector
```bash
go test ./... -race -v
```

#### Flaky Tests
**Solutions:**
- Add proper cleanup
- Use test isolation
- Fix timing dependencies
- Add retries for external dependencies

#### Mock Not Working
**Solutions:**
- Verify interface implementation
- Check mock expectations
- Use mock.AssertExpectations(t)

---

## Additional Resources

- [Go Testing Package Documentation](https://pkg.go.dev/testing)
- [Testify Documentation](https://github.com/stretchr/testify)
- [Services Documentation](./SERVICES.md)
- [API Documentation](./README.md)
