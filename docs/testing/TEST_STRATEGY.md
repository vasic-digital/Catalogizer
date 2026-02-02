# Test Strategy

This document describes the testing approach across all Catalogizer modules, including test categories, infrastructure, conventions, and how to write tests for each part of the system.

## Test Coverage Overview

| Module | Test Files | Categories | Command |
|--------|-----------|------------|---------|
| catalog-api (Go) | ~77 | Unit, integration, benchmark | `cd catalog-api && go test ./...` |
| catalog-web (React) | ~38 | Unit, integration, snapshot, accessibility | `cd catalog-web && npm run test` |
| catalogizer-android (Kotlin) | ~5 | Unit (ViewModel) | `cd catalogizer-android && ./gradlew test` |
| catalogizer-api-client (TS) | ~1 | Unit | `cd catalogizer-api-client && npm run test` |
| Full system | N/A | All above + security | `./scripts/run-all-tests.sh` |

## Test Categories

### 1. Unit Tests

Isolated tests for individual functions, classes, or components. Dependencies are mocked.

**Go backend** - Files named `*_test.go` beside source files:

```go
// handlers/auth_handler_test.go
type AuthHandlerTestSuite struct {
    suite.Suite
    handler     *AuthHandler
    authService *services.AuthService
}

func (suite *AuthHandlerTestSuite) SetupTest() {
    suite.authService = services.NewAuthService(nil, "test-secret-key")
    suite.handler = NewAuthHandler(suite.authService)
}

func (suite *AuthHandlerTestSuite) TestLoginMethodNotAllowed() {
    req, _ := http.NewRequest("GET", "/login", nil)
    w := httptest.NewRecorder()
    suite.handler.Login(w, req)
    assert.Equal(suite.T(), http.StatusMethodNotAllowed, w.Code)
}

func TestAuthHandlerSuite(t *testing.T) {
    suite.Run(t, new(AuthHandlerTestSuite))
}
```

**React frontend** - Files in `__tests__/` directories:

```tsx
// components/auth/__tests__/LoginForm.test.tsx
jest.mock('@/contexts/AuthContext', () => ({
  useAuth: jest.fn(),
}))

describe('LoginForm', () => {
  beforeEach(() => {
    jest.clearAllMocks()
    mockUseAuth.mockReturnValue({
      login: jest.fn().mockResolvedValue(undefined),
    })
  })

  it('renders the login form with all elements', () => {
    render(<MemoryRouter><LoginForm /></MemoryRouter>)
    expect(screen.getByText('Welcome back')).toBeInTheDocument()
    expect(screen.getByPlaceholderText('Enter your username')).toBeInTheDocument()
  })
})
```

**Android** - Files in `src/test/`:

```kotlin
// ui/viewmodel/AuthViewModelTest.kt
@ExperimentalCoroutinesApi
class AuthViewModelTest {
    @get:Rule val instantExecutorRule = InstantTaskExecutorRule()
    @get:Rule val mainDispatcherRule = MainDispatcherRule()

    @Before
    fun setup() {
        Dispatchers.setMain(StandardTestDispatcher())
        mockAuthRepository = mockk(relaxed = true)
        every { mockAuthRepository.isAuthenticated } returns flowOf(true)
        viewModel = AuthViewModel(mockAuthRepository)
    }

    @Test
    fun `initial auth state should check authentication status`() = runTest {
        advanceUntilIdle()
        assertNotNull(viewModel.authState)
        verify { mockAuthRepository.isAuthenticated }
    }
}
```

### 2. Integration Tests

Tests that exercise multiple components together, often with a real or in-memory database.

**Go backend** - Named `*_integration_test.go`:

```go
// internal/auth/token_integration_test.go
// Tests the full token lifecycle: create -> validate -> refresh -> invalidate
```

**React frontend** - Named `*.integration.test.tsx`:

```tsx
// __tests__/AuthFlow.integration.test.tsx
// Tests the full auth flow: login -> store token -> access protected route -> logout

// components/auth/__tests__/ProtectedRoute.integration.test.tsx
// Tests ProtectedRoute with real AuthContext behavior
```

### 3. Benchmark Tests

Performance tests that measure execution speed under load. Go-specific.

**Go backend** - Named `*_bench_test.go`:

```go
// internal/media/detector/engine_bench_test.go
func BenchmarkAnalyzeDirectory(b *testing.B) {
    sizes := []int{5, 50, 500}
    for _, size := range sizes {
        b.Run(fmt.Sprintf("files=%d", size), func(b *testing.B) {
            engine := newBenchEngine()
            engine.LoadRules(rules, mediaTypes)
            files := generateFiles(size)
            b.ResetTimer()
            for i := 0; i < b.N; i++ {
                engine.AnalyzeDirectory("/test", files)
            }
        })
    }
}
```

Run benchmarks:
```bash
cd catalog-api
go test -bench=. ./internal/media/detector/
go test -bench=. ./internal/media/providers/
go test -bench=. ./services/ -run=^$ -benchmem
```

### 4. Snapshot Tests

Capture rendered component output and compare against saved snapshots. React-specific.

```tsx
// components/__tests__/snapshots.test.tsx
// Captures rendered HTML structure of UI components
// Stored in __snapshots__/snapshots.test.tsx.snap
```

Update snapshots when intentional UI changes are made:
```bash
cd catalog-web
npm run test -- --updateSnapshot
```

### 5. Accessibility Tests

Validate components meet WCAG accessibility standards using `jest-axe`.

```tsx
// components/__tests__/accessibility.test.tsx
import { axe, toHaveNoViolations } from 'jest-axe'
expect.extend(toHaveNoViolations)

describe('Accessibility Tests', () => {
  it('Button has no accessibility violations', async () => {
    const { container } = render(<Button>Click me</Button>)
    const results = await axe(container)
    expect(results).toHaveNoViolations()
  })

  it('Input has no accessibility violations', async () => {
    const { container } = render(<Input label="Name" />)
    const results = await axe(container)
    expect(results).toHaveNoViolations()
  })
})
```

Components tested for accessibility: Button, Input, Card, Badge, Select, Textarea, Switch, Progress.

## How to Write Tests for Each Module

### Go Backend (catalog-api)

**File placement**: `*_test.go` beside the source file (same package).

**Framework**: `testing` standard library + `testify/suite` + `testify/assert`.

**Pattern: Table-driven tests**

```go
func TestValidateConfig(t *testing.T) {
    tests := []struct {
        name    string
        config  *Config
        wantErr bool
    }{
        {
            name:    "valid config",
            config:  getDefaultConfig(),
            wantErr: false,
        },
        {
            name:    "invalid port",
            config:  &Config{Server: ServerConfig{Port: -1}},
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := validateConfig(tt.config)
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

**Pattern: Test suite with setup/teardown**

```go
type MyServiceTestSuite struct {
    suite.Suite
    service *MyService
    db      *sql.DB
}

func (suite *MyServiceTestSuite) SetupTest() {
    // Create in-memory DB or mock
    suite.service = NewMyService(...)
}

func (suite *MyServiceTestSuite) TearDownTest() {
    // Cleanup
}

func (suite *MyServiceTestSuite) TestSomeMethod() {
    result, err := suite.service.SomeMethod()
    assert.NoError(suite.T(), err)
    assert.NotNil(suite.T(), result)
}

func TestMyServiceSuite(t *testing.T) {
    suite.Run(t, new(MyServiceTestSuite))
}
```

**Run commands**:
```bash
go test ./...                              # All tests
go test -v -run TestName ./handlers/       # Single test
go test -v -count=1 ./...                  # No cache
go test -bench=. -benchmem ./internal/...  # Benchmarks
go test -cover ./...                       # With coverage
```

### React Frontend (catalog-web)

**File placement**: `__tests__/` directory inside the component's directory.

**Framework**: Jest + React Testing Library + jest-axe.

**Test setup** (`src/test/setup.ts`): Mocks for WebSocket, localStorage, IntersectionObserver, ResizeObserver, HTMLMediaElement, Canvas, and crypto.

**Pattern: Component rendering test**

```tsx
import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { QueryClientProvider, QueryClient } from '@tanstack/react-query'

const queryClient = new QueryClient({
  defaultOptions: { queries: { retry: false } }
})

const renderWithProviders = (ui: React.ReactElement) => {
  return render(
    <QueryClientProvider client={queryClient}>
      <MemoryRouter>
        {ui}
      </MemoryRouter>
    </QueryClientProvider>
  )
}

describe('MediaCard', () => {
  it('renders media title', () => {
    renderWithProviders(<MediaCard item={mockMediaItem} />)
    expect(screen.getByText('Test Movie')).toBeInTheDocument()
  })
})
```

**Pattern: User interaction test**

```tsx
import userEvent from '@testing-library/user-event'

it('calls login on form submit', async () => {
  const mockLogin = jest.fn().mockResolvedValue(undefined)
  mockUseAuth.mockReturnValue({ login: mockLogin })

  render(<MemoryRouter><LoginForm /></MemoryRouter>)

  await userEvent.type(screen.getByPlaceholderText('Enter your username'), 'admin')
  await userEvent.type(screen.getByPlaceholderText('Enter your password'), 'password')
  await userEvent.click(screen.getByRole('button', { name: /sign in/i }))

  expect(mockLogin).toHaveBeenCalledWith({
    username: 'admin',
    password: 'password',
  })
})
```

**Pattern: Mocking contexts**

```tsx
jest.mock('@/contexts/AuthContext', () => ({
  useAuth: jest.fn(),
}))

const mockUseAuth = require('@/contexts/AuthContext').useAuth
mockUseAuth.mockReturnValue({
  isAuthenticated: true,
  user: { role: 'admin', username: 'admin' },
  hasPermission: jest.fn().mockReturnValue(true),
})
```

**Pattern: Mocking framer-motion (required for animated components)**

```tsx
jest.mock('framer-motion', () => ({
  motion: {
    div: ({ children, ...props }: any) => <div {...props}>{children}</div>,
  },
}))
```

**Run commands**:
```bash
npm run test                  # All tests
npm run test -- --watch       # Watch mode
npm run test -- --coverage    # With coverage
npm run test -- --updateSnapshot  # Update snapshots
npm run test -- LoginForm     # Specific test file
```

### Android (catalogizer-android)

**File placement**: `src/test/java/com/catalogizer/android/`.

**Framework**: JUnit 4 + MockK + Kotlin coroutine test utilities.

**Required test rules**:

```kotlin
@get:Rule val instantExecutorRule = InstantTaskExecutorRule()  // Sync LiveData
@get:Rule val mainDispatcherRule = MainDispatcherRule()        // Replace Main dispatcher
```

**`MainDispatcherRule`** (custom rule in test sources):

```kotlin
class MainDispatcherRule(
    private val dispatcher: TestDispatcher = StandardTestDispatcher()
) : TestWatcher() {
    override fun starting(description: Description) = Dispatchers.setMain(dispatcher)
    override fun finished(description: Description) = Dispatchers.resetMain()
}
```

**Pattern: ViewModel test**

```kotlin
@Test
fun `login should update auth state`() = runTest {
    val loginResponse = LoginResponse(token = "test_token", ...)
    coEvery { mockAuthRepository.login(any(), any()) } returns ApiResult.success(loginResponse)

    viewModel.login("testuser", "password")
    advanceUntilIdle()

    val state = viewModel.authState.value
    assertTrue(state.isAuthenticated)
}
```

**Run commands**:
```bash
./gradlew test                              # All unit tests
./gradlew testDebugUnitTest                 # Debug variant only
./gradlew test --tests "*AuthViewModelTest" # Specific test class
```

### API Client (catalogizer-api-client)

```bash
cd catalogizer-api-client
npm run build && npm run test
```

## Test Infrastructure and Utilities

### Go: Test Helpers

- **`httptest.NewRecorder()`** - Captures HTTP responses for handler tests
- **`httptest.NewServer()`** - Creates real HTTP servers for integration tests
- **`t.Parallel()`** - Marks tests as safe to run concurrently
- **`testify/suite`** - Groups related tests with shared setup/teardown

### React: Test Setup (`src/test/setup.ts`)

The setup file provides mocks for browser APIs not available in jsdom:
- `window.matchMedia` - Media query matching
- `IntersectionObserver` - Visibility detection
- `ResizeObserver` - Element resize detection
- `WebSocket` - WebSocket connection simulation (auto-connects after 10ms)
- `localStorage` / `sessionStorage` - Storage APIs
- `HTMLMediaElement.prototype.play/pause` - Video/audio playback
- `HTMLCanvasElement.getContext` - Canvas 2D rendering context
- `crypto.randomUUID` - UUID generation

### Android: Test Utilities

- `MainDispatcherRule` - Custom JUnit rule that replaces `Dispatchers.Main` with `StandardTestDispatcher`
- `CatalogizerTestApplication` - Test application class for instrumented tests
- `MockK` configuration: Use `relaxed = true` for mock dependencies, `coEvery` for suspend functions

## CI/CD Test Pipeline

The comprehensive test script (`scripts/run-all-tests.sh`) runs all tests in sequence:

```bash
#!/bin/bash
# scripts/run-all-tests.sh
set -e

# 1. Go backend tests
cd catalog-api
go test ./...

# 2. React frontend tests
cd ../catalog-web
npm run test
npm run lint
npm run type-check

# 3. Android tests
cd ../catalogizer-android
./gradlew test

# 4. API client tests
cd ../catalogizer-api-client
npm run build && npm run test

# 5. Security scans (optional, requires tools)
# SonarQube scan
# Snyk dependency vulnerability scan
```

The script tracks pass/fail counts and generates reports to `reports/`.

### Docker-based testing

```bash
# Full development environment
docker-compose -f docker-compose.dev.yml up

# Security testing environment
docker-compose -f docker-compose.security.yml up
```

### Security testing

Security-specific tests are in the middleware layer:

```
catalog-api/middleware/
├── redis_rate_limiter_security_test.go       # Rate limiting bypass tests
├── redis_rate_limiter_security_fixed_test.go  # Regression tests for security fixes
├── input_validation_test.go                   # Injection prevention tests
└── auth_test.go                               # Auth middleware tests
```

Run security tests:
```bash
cd catalog-api
go test -v ./middleware/ -run Security
```

External security scanning:
```bash
./scripts/security-test.sh     # General security tests
./scripts/snyk-scan.sh         # Dependency vulnerability scanning
./scripts/sonarqube-scan.sh    # Code quality and security analysis
```

## Writing Tests: Quick Reference

| Module | Location | Framework | Mocking | Runner |
|--------|----------|-----------|---------|--------|
| Go backend | `*_test.go` beside source | testify/suite + assert | Interface-based / constructor injection | `go test` |
| React frontend | `__tests__/` inside component dir | Jest + RTL | `jest.mock()` | `npm run test` |
| Android | `src/test/` mirroring main structure | JUnit 4 + MockK | `mockk(relaxed=true)` | `./gradlew test` |
| API client | beside source | Jest | `jest.mock()` | `npm run test` |
| Benchmarks | `*_bench_test.go` | Go testing.B | N/A | `go test -bench=.` |
| Accessibility | `accessibility.test.tsx` | jest-axe | N/A | `npm run test` |
| Snapshots | `snapshots.test.tsx` | Jest snapshots | N/A | `npm run test` |

## Best Practices

1. **Test behavior, not implementation** - Assert on observable outputs, not internal state
2. **Keep tests independent** - Each test should set up its own state and not depend on test execution order
3. **Use descriptive test names** - Go: `TestLoginWithInvalidCredentials`, Kotlin: backtick names `\`login should update auth state\``, React: nested `describe`/`it` blocks
4. **Mock at boundaries** - Mock external services (APIs, databases) but not the code under test
5. **Run tests before committing** - All tests must pass; the CI pipeline enforces this
6. **Maintain the test setup file** - When adding new browser APIs to the frontend, add mocks to `src/test/setup.ts`
7. **Benchmark before optimizing** - Use Go benchmarks to measure performance impact of changes
