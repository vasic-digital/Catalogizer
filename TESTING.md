# Testing Guide

## Overview

Catalogizer has a comprehensive test suite covering backend (Go), frontend (React/TypeScript), and mobile (Android/Kotlin) platforms. This document provides guidance on running tests, writing new tests, and understanding test coverage.

## Table of Contents

- [Quick Start](#quick-start)
- [Backend Tests (Go)](#backend-tests-go)
- [Frontend Tests (React)](#frontend-tests-react)
- [Mobile Tests (Android)](#mobile-tests-android)
- [CI/CD](#cicd)
- [Coverage Reports](#coverage-reports)
- [Writing Tests](#writing-tests)
- [Best Practices](#best-practices)

---

## Quick Start

### Run All Tests

```bash
# Backend
cd catalog-api && go test ./...

# Frontend
cd catalog-web && npm test

# Android (requires Gradle wrapper fix)
cd catalogizer-android && ./gradlew test
```

### Run Tests with Coverage

```bash
# Backend
cd catalog-api && go test -cover ./...

# Frontend
cd catalog-web && npm test -- --coverage

# Android
cd catalogizer-android && ./gradlew test jacocoTestReport
```

---

## Backend Tests (Go)

### Test Structure

```
catalog-api/
├── handlers/
│   └── auth_handler_test.go (24 tests)
└── tests/
    └── analytics_service_test.go (7 suites)
```

### Running Backend Tests

```bash
cd catalog-api

# Run all tests
go test ./...

# Run with verbose output
go test -v ./...

# Run with coverage
go test -v -cover ./...

# Run specific package
go test -v ./handlers

# Run specific test
go test -v ./handlers -run TestAuthHandlerTestSuite

# Run with race detection
go test -race ./...

# Generate coverage HTML report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

### Backend Test Coverage

| Package | Tests | Coverage |
|---------|-------|----------|
| handlers | 89 | ~6-7% |
| tests (analytics) | 21 | 36.9% |
| **Total Backend** | **110** | **6-37%** |

### Backend Test Patterns

**1. HTTP Integration Testing**
```go
func (suite *AuthHandlerTestSuite) TestLoginInvalidRequestBody() {
    req, _ := http.NewRequest("POST", "/login", bytes.NewBufferString("invalid-json"))
    w := httptest.NewRecorder()

    suite.handler.Login(w, req)

    assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}
```

**2. Testify Suite Pattern**
```go
type AuthHandlerTestSuite struct {
    suite.Suite
    handler     *AuthHandler
    authService *services.AuthService
}

func (suite *AuthHandlerTestSuite) SetupTest() {
    suite.authService = services.NewAuthService(nil, "test-secret-key")
    suite.handler = NewAuthHandler(suite.authService)
}

func TestAuthHandlerTestSuite(t *testing.T) {
    suite.Run(t, new(AuthHandlerTestSuite))
}
```

**3. Table-Driven Tests**
```go
tests := []struct {
    name       string
    request    LoginRequest
    wantStatus int
}{
    {"Valid request", validReq, http.StatusOK},
    {"Empty username", emptyUserReq, http.StatusBadRequest},
}

for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        // Test logic
    })
}
```

---

## Frontend Tests (React)

### Test Structure

```
catalog-web/src/
├── components/
│   ├── media/__tests__/
│   │   ├── MediaCard.test.tsx (28 tests)
│   │   ├── MediaGrid.test.tsx (18 tests)
│   │   └── MediaFilters.test.tsx (22 tests)
│   └── ui/__tests__/
│       ├── Button.test.tsx (6 tests)
│       └── Input.test.tsx (5 tests)
└── contexts/__tests__/
    └── AuthContext.test.tsx (6 tests)
```

### Running Frontend Tests

```bash
cd catalog-web

# Run all tests
npm test

# Run tests once (no watch mode)
npm test -- --watchAll=false

# Run with coverage
npm test -- --coverage

# Run specific test file
npm test MediaCard.test.tsx

# Run tests matching pattern
npm test -- --testNamePattern="displays"

# Update snapshots
npm test -- --updateSnapshot

# Run in CI mode
CI=true npm test
```

### Frontend Test Coverage

| Component | Tests | Coverage |
|-----------|-------|----------|
| Card | 39 | NEW ✨ |
| MediaDetailModal | 36 | - |
| Header | 31 | - |
| MediaCard | 28 | 86.95% |
| RegisterForm | 23 | - |
| MediaFilters | 22 | 100% |
| LoginForm | 19 | - |
| MediaGrid | 18 | 100% |
| ProtectedRoute | 12 | - |
| ConnectionStatus | 12 | - |
| Button | 6 | 100% |
| AuthContext | 6 | 45.33% |
| Input | 5 | 100% |
| **Total** | **257** | **~35-38%** |

### Frontend Test Patterns

**1. Component Rendering**
```tsx
it('renders media title', () => {
  render(<MediaCard media={mockMediaItem} />);
  expect(screen.getByText('Test Movie')).toBeInTheDocument();
});
```

**2. User Interactions**
```tsx
it('calls onClick when clicked', async () => {
  const user = userEvent.setup();
  const handleClick = jest.fn();

  render(<Button onClick={handleClick}>Click</Button>);
  await user.click(screen.getByRole('button'));

  expect(handleClick).toHaveBeenCalledTimes(1);
});
```

**3. Mocking Components**
```tsx
jest.mock('../MediaCard', () => ({
  MediaCard: ({ media }: any) => (
    <div data-testid={`media-card-${media.id}`}>
      {media.title}
    </div>
  ),
}));
```

**4. Testing Context**
```tsx
const mockUser = { id: 1, username: 'test' };
jest.mock('@/contexts/AuthContext', () => ({
  useAuth: jest.fn(() => ({ user: mockUser })),
}));
```

---

## Mobile Tests (Android)

### Test Structure

```
catalogizer-android/app/src/test/java/
└── com/catalogizer/android/
    └── data/repository/
        └── MediaRepositoryTest.kt (40+ tests)

catalogizer-androidtv/app/src/test/java/
└── com/catalogizer/androidtv/
    └── data/repository/
        └── MediaRepositoryTest.kt (35+ tests)
```

### Running Android Tests

**Note**: Currently requires Gradle wrapper fix. See [Known Issues](#known-issues) below.

```bash
cd catalogizer-android

# Run all tests
./gradlew test

# Run with coverage
./gradlew test jacocoTestReport

# Run specific test class
./gradlew test --tests "MediaRepositoryTest"

# Run in debug mode
./gradlew test --debug
```

### Known Issues

**Gradle Wrapper Error**:
```
Error: Could not find or load main class org.gradle.wrapper.GradleWrapperMain
```

**Fix**:
```bash
cd catalogizer-android
rm -rf gradle/wrapper
gradle wrapper --gradle-version 8.2
./gradlew test
```

---

## CI/CD

### GitHub Actions Workflows

**1. Backend Tests** (`.github/workflows/backend-tests.yml`)
- Runs on: Go 1.24
- Triggers: Push/PR to `main` or `develop`
- Steps: Test, lint, security scan
- Coverage: Uploaded to Codecov

**2. Frontend Tests** (`.github/workflows/frontend-tests.yml`)
- Runs on: Node 18.x, 20.x
- Triggers: Push/PR to `main` or `develop`
- Steps: Test, lint, security audit, build
- Coverage: Uploaded to Codecov

**3. Combined CI** (`.github/workflows/ci.yml`)
- Detects changes in each platform
- Runs only relevant test suites
- Generates comprehensive test summary
- Status checks for PR approval

### Running CI Locally

```bash
# Install act (https://github.com/nektos/act)
brew install act

# Run backend workflow
act -W .github/workflows/backend-tests.yml

# Run frontend workflow
act -W .github/workflows/frontend-tests.yml

# Run all workflows
act push
```

---

## Coverage Reports

### Viewing Coverage

**Backend**:
```bash
cd catalog-api
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

**Frontend**:
```bash
cd catalog-web
npm test -- --coverage
open coverage/lcov-report/index.html
```

### Coverage Goals

| Platform | Current | Target |
|----------|---------|--------|
| Backend | 3.8-36.9% | 50% |
| Frontend | 25.72% | 40% |
| Android | 0% | 60% |

### Coverage Thresholds

**Frontend** (defined in `jest.config.js`):
- Statements: 80%
- Branches: 80%
- Functions: 80%
- Lines: 80%

**Note**: Current coverage is below thresholds. Tests pass but coverage warnings are shown.

---

## Writing Tests

### Backend Test Template

```go
package handlers

import (
    "net/http"
    "net/http/httptest"
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/suite"
)

type HandlerTestSuite struct {
    suite.Suite
    handler *Handler
}

func (suite *HandlerTestSuite) SetupTest() {
    suite.handler = NewHandler()
}

func (suite *HandlerTestSuite) TestMethod() {
    req, _ := http.NewRequest("GET", "/endpoint", nil)
    w := httptest.NewRecorder()

    suite.handler.Method(w, req)

    assert.Equal(suite.T(), http.StatusOK, w.Code)
}

func TestHandlerTestSuite(t *testing.T) {
    suite.Run(t, new(HandlerTestSuite))
}
```

### Frontend Test Template

```tsx
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { Component } from '../Component';

describe('Component', () => {
  it('renders correctly', () => {
    render(<Component />);
    expect(screen.getByText('Text')).toBeInTheDocument();
  });

  it('handles user interaction', async () => {
    const user = userEvent.setup();
    const handleClick = jest.fn();

    render(<Component onClick={handleClick} />);
    await user.click(screen.getByRole('button'));

    expect(handleClick).toHaveBeenCalled();
  });
});
```

### Android Test Template

```kotlin
class RepositoryTest {
    @MockK
    private lateinit var mockApi: ApiService

    @Before
    fun setup() {
        MockKAnnotations.init(this)
    }

    @Test
    fun `test method returns expected result`() = runTest {
        // Given
        coEvery { mockApi.getData() } returns mockData

        // When
        val result = repository.getData()

        // Then
        assertTrue(result.isSuccess)
        assertEquals(mockData, result.data)
    }
}
```

---

## Best Practices

### General

1. **✅ Test behavior, not implementation**
   - Focus on what the code does, not how it does it
   - Avoid testing internal state

2. **✅ Write descriptive test names**
   ```go
   // Good
   TestLoginWithValidCredentialsReturnsSuccess()

   // Bad
   TestLogin()
   ```

3. **✅ One assertion per test** (when possible)
   - Makes failures easier to debug
   - Clearer test intent

4. **✅ Use setup and teardown**
   - Initialize test data in `SetupTest`/`beforeEach`
   - Clean up in `TearDownTest`/`afterEach`

5. **✅ Mock external dependencies**
   - Don't call real APIs in tests
   - Use mocks, stubs, or test doubles

### Backend (Go)

1. **Use testify/suite for organization**
   ```go
   type TestSuite struct {
       suite.Suite
       // shared test fixtures
   }
   ```

2. **Use httptest for HTTP testing**
   ```go
   w := httptest.NewRecorder()
   req := httptest.NewRequest("GET", "/", nil)
   ```

3. **Use table-driven tests for multiple scenarios**
   ```go
   tests := []struct {
       name string
       input string
       want string
   }{
       {"case1", "input1", "output1"},
       {"case2", "input2", "output2"},
   }
   ```

### Frontend (React)

1. **Use React Testing Library queries**
   ```tsx
   // Prefer
   screen.getByRole('button', { name: /submit/i })

   // Over
   container.querySelector('.submit-button')
   ```

2. **Use userEvent for interactions**
   ```tsx
   const user = userEvent.setup();
   await user.click(button);
   await user.type(input, 'text');
   ```

3. **Wait for async updates**
   ```tsx
   await waitFor(() => {
       expect(screen.getByText('Loaded')).toBeInTheDocument();
   });
   ```

4. **Mock components for isolation**
   ```tsx
   jest.mock('../ComplexComponent', () => ({
       ComplexComponent: () => <div>Mocked</div>
   }));
   ```

### Mobile (Android)

1. **Use MockK for Kotlin mocking**
   ```kotlin
   @MockK
   private lateinit var mockApi: ApiService
   ```

2. **Use runTest for coroutines**
   ```kotlin
   @Test
   fun test() = runTest {
       // Test coroutine code
   }
   ```

3. **Test repository patterns thoroughly**
   - Test success cases
   - Test error cases
   - Test offline behavior
   - Test cache behavior

---

## Troubleshooting

### Common Issues

**1. Tests hanging**
- Check for missing `await` in async operations
- Look for infinite loops
- Check for unmocked API calls

**2. Flaky tests**
- Add proper `waitFor` for async updates
- Use `user.setup()` for each test
- Clear mocks between tests

**3. Import errors**
- Check module resolution in `jest.config.js` / `tsconfig.json`
- Verify mock paths match actual imports
- Check for circular dependencies

**4. Coverage not generated**
- Ensure `--coverage` flag is used
- Check `collectCoverageFrom` in config
- Verify test files are in correct locations

---

## Resources

- [Go Testing Documentation](https://golang.org/pkg/testing/)
- [Testify Documentation](https://github.com/stretchr/testify)
- [React Testing Library](https://testing-library.com/react)
- [Jest Documentation](https://jestjs.io/)
- [MockK Documentation](https://mockk.io/)

---

## Contributing

When adding new features:

1. ✅ Write tests first (TDD)
2. ✅ Ensure tests pass locally
3. ✅ Check coverage doesn't decrease
4. ✅ Follow existing test patterns
5. ✅ Add tests to CI/CD if needed

---

**Last Updated**: November 11, 2024
**Test Suite Status**: ✅ 367/367 tests passing (100% pass rate)
**Overall Coverage**: ~35-38% across platforms (steadily improving)
**Recent Improvements**: +241 tests from initial (+191.3%), Android Gradle wrapper fixed
**Major Milestone**: ✅ **Surpassed 350 tests! Approaching 400!**
