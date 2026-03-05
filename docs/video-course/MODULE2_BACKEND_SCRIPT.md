# Module 2: Backend Development - Core Services - Script

**Duration**: 60 minutes
**Module**: 2 - Backend Development

---

## Scene 1: Project Structure and Main Entry Point (0:00 - 15:00)

**[Visual: Terminal showing catalog-api directory tree]**

**Narrator**: Welcome to Module 2. In this module, we build the Go backend from the ground up. Catalogizer's backend follows a clean, layered architecture: Handler, Service, Repository, and Database. Let us start by examining how the application boots up.

**[Visual: Open `catalog-api/main.go` in editor]**

**Narrator**: The `main.go` file is the orchestrator. It wires together every layer of the application using constructor injection. Notice how every service receives its dependencies explicitly -- there is no global state, no service locator. This is the `NewService` pattern used throughout the project.

```go
// catalog-api/main.go
import (
    "catalogizer/database"
    "catalogizer/handlers"
    "catalogizer/services"
    "catalogizer/repository"
    "catalogizer/internal/auth"
    "catalogizer/internal/services"
    // ...
)

var (
    Version     = "dev"
    BuildNumber = "0"
    BuildDate   = "unknown"
)
```

**[Visual: Highlight the version variables]**

**Narrator**: Version, BuildNumber, and BuildDate are injected at build time via Go's `-ldflags` mechanism. This allows every binary to report exactly which commit and build produced it.

**[Visual: Show the `findAvailablePort` function]**

**Narrator**: Catalogizer uses dynamic port binding. On startup, it probes for an available port starting from the configured value, then writes the chosen port to a `.service-port` file. The frontend reads this file to configure its API proxy.

```go
// catalog-api/main.go
func findAvailablePort(host string, startPort, maxAttempts int) (int, error) {
    discoverer := discovery.NewTCPDiscoverer()
    for i := 0; i < maxAttempts; i++ {
        port := startPort + i
        target := discovery.DiscoveryTarget{
            Name:    "catalog-api",
            Host:    host,
            Port:    strconv.Itoa(port),
            Method:  "tcp",
            Timeout: 100 * time.Millisecond,
        }
        reachable, err := discoverer.Discover(context.Background(), target)
        if err != nil || !reachable {
            return port, nil
        }
    }
    return 0, fmt.Errorf("no available port in range %d-%d", startPort, startPort+maxAttempts-1)
}
```

**[Visual: Show directory layout diagram]**

**Narrator**: The project uses a dual package layout. Top-level packages -- `handlers/`, `services/`, `repository/`, `middleware/` -- contain domain logic. The `internal/` directory mirrors this structure for infrastructure concerns: `internal/handlers/`, `internal/services/`, `internal/auth/`, `internal/media/`. This separation keeps domain code clean and infrastructure code hidden from external consumers.

**[Visual: Show graceful shutdown code]**

**Narrator**: Graceful shutdown is handled by listening for OS signals -- SIGINT and SIGTERM -- then draining active connections before exiting. This ensures no request is dropped during deployment.

```go
// catalog-api/main.go
quit := make(chan os.Signal, 1)
signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
<-quit
```

---

## Scene 2: Database Layer (15:00 - 35:00)

**[Visual: Open `catalog-api/database/dialect.go`]**

**Narrator**: The database layer is where Catalogizer's dual-dialect abstraction lives. This is one of the most elegant parts of the codebase. A single `Dialect` struct handles all SQL differences between SQLite and PostgreSQL.

```go
// catalog-api/database/dialect.go
type DialectType string

const (
    DialectSQLite   DialectType = "sqlite"
    DialectPostgres DialectType = "postgres"
)

type Dialect struct {
    Type DialectType
}
```

**[Visual: Show `RewritePlaceholders` function]**

**Narrator**: `RewritePlaceholders` converts SQLite's `?` placeholders into PostgreSQL's numbered `$1, $2, $3` format. This means every repository can write standard SQLite-style queries, and PostgreSQL support comes for free.

```go
// catalog-api/database/dialect.go
func (d *Dialect) RewritePlaceholders(query string) string {
    if d.Type != DialectPostgres {
        return query
    }
    var b strings.Builder
    n := 0
    inSingleQuote := false
    for i := 0; i < len(query); i++ {
        ch := query[i]
        if ch == '\'' {
            inSingleQuote = !inSingleQuote
        }
        if ch == '?' && !inSingleQuote {
            n++
            fmt.Fprintf(&b, "$%d", n)
        } else {
            b.WriteByte(ch)
        }
    }
    return b.String()
}
```

**[Visual: Show `RewriteInsertOrIgnore` function]**

**Narrator**: Similarly, `RewriteInsertOrIgnore` converts SQLite's `INSERT OR IGNORE INTO` to PostgreSQL's `INSERT INTO ... ON CONFLICT DO NOTHING`. And `BooleanLiterals` rewrites `= 0` and `= 1` to `= FALSE` and `= TRUE` for known boolean columns.

**[Visual: Open `catalog-api/database/connection.go`]**

**Narrator**: The `DB` struct wraps Go's standard `*sql.DB` with shadowed `Exec()`, `Query()`, and `QueryRow()` methods. Every SQL call passes through the dialect rewriter automatically. You never have to think about which database you are targeting.

```go
// catalog-api/database/connection.go
type DB struct {
    *sql.DB
    config  *config.DatabaseConfig
    dialect Dialect
}
```

**[Visual: Show connection setup with WAL mode]**

**Narrator**: For SQLite, we explicitly set WAL (Write-Ahead Logging) mode with a `PRAGMA` after opening the connection. This is necessary because `go-sqlcipher` ignores connection string pragmas. WAL mode dramatically improves concurrent read performance.

**[Visual: Open `catalog-api/database/migrations.go`]**

**Narrator**: The migration system is version-based. Each migration has a version number, a name, and an `Up` function. The system tracks which migrations have been applied in a `migrations` table, and only runs new ones.

```go
// catalog-api/database/migrations.go
migrations := []Migration{
    {Version: 1, Name: "create_initial_tables", Up: db.createInitialTables},
    {Version: 2, Name: "migrate_smb_to_storage_roots", Up: db.migrateSMBToStorageRoots},
    {Version: 3, Name: "create_auth_tables", Up: db.createAuthTables},
    // ... up to version 9
    {Version: 9, Name: "create_performance_indexes", Up: db.createPerformanceIndexes},
}
```

**[Visual: Show separate SQLite and PostgreSQL migration files]**

**Narrator**: Each migration has separate SQLite and PostgreSQL implementations in `migrations_sqlite.go` and `migrations_postgres.go`. This keeps dialect-specific DDL cleanly separated.

**[Visual: Show `WrapDB` usage in tests]**

**Narrator**: For unit tests, `database.WrapDB(sqlDB, DialectSQLite)` wraps an in-memory SQLite database with the same dialect-aware interface. Tests run fast and require no external dependencies.

---

## Scene 3: Service Layer Design (35:00 - 50:00)

**[Visual: Open `catalog-api/services/auth_service.go`]**

**Narrator**: The service layer contains all business logic. Every service follows the same pattern: a struct with private fields, a `NewService` constructor that accepts dependencies, and methods that implement business operations.

```go
// catalog-api/services/auth_service.go
type AuthService struct {
    userRepo   *repository.UserRepository
    jwtSecret  []byte
    jwtExpiry  time.Duration
    refreshExp time.Duration
}

func NewAuthService(userRepo *repository.UserRepository, jwtSecret string) *AuthService {
    return &AuthService{
        userRepo:   userRepo,
        jwtSecret:  []byte(jwtSecret),
        jwtExpiry:  24 * time.Hour,
        refreshExp: 7 * 24 * time.Hour,
    }
}
```

**[Visual: Show error wrapping pattern]**

**Narrator**: Error handling follows Go best practices: wrap errors with context using `fmt.Errorf` and the `%w` verb. This preserves the error chain for debugging while adding meaningful context at each layer.

```go
// catalog-api/services/auth_service.go
func (s *AuthService) Login(req models.LoginRequest, ipAddress string, userAgent string) (*AuthResult, error) {
    user, err := s.userRepo.GetByUsernameOrEmail(req.Username)
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, errors.New("invalid credentials")
        }
        return nil, fmt.Errorf("failed to get user: %w", err)
    }
    // ...
}
```

**[Visual: Show the repository layer pattern]**

**Narrator**: Services depend on repositories, never on the database directly. The repository layer encapsulates all SQL queries. This makes services testable -- you can mock the repository interface in tests.

```go
// catalog-api/repository/file_repository.go
type FileRepository struct {
    db *database.DB
}

func NewFileRepository(db *database.DB) *FileRepository {
    return &FileRepository{db: db}
}

func (r *FileRepository) GetFileByID(ctx context.Context, id int64) (*models.FileWithMetadata, error) {
    query := `SELECT f.id, f.storage_root_id, ... FROM files f WHERE f.id = ?`
    // The dialect-aware DB automatically rewrites ? for PostgreSQL
    err := r.db.QueryRowContext(ctx, query, id).Scan(...)
    // ...
}
```

**[Visual: Show context propagation through layers]**

**Narrator**: Every function that does I/O accepts a `context.Context` as its first parameter. This allows request cancellation and timeout propagation from the HTTP handler all the way to the database query.

---

## Scene 4: Handler Layer with Gin (50:00 - 60:00)

**[Visual: Open `catalog-api/handlers/auth_handler.go`]**

**Narrator**: The handler layer is thin. Handlers parse requests, call services, and format responses. They never contain business logic.

```go
// catalog-api/handlers/auth_handler.go
func (h *AuthHandler) LoginGin(c *gin.Context) {
    var req models.LoginRequest

    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
        return
    }

    ipAddress := c.ClientIP()
    userAgent := c.GetHeader("User-Agent")

    result, err := h.authService.Login(req, ipAddress, userAgent)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, result)
}
```

**[Visual: Show Gin router setup in main.go]**

**Narrator**: Routes are registered under `/api/v1` in `main.go`. The Gin router groups endpoints by feature -- auth, files, media, collections -- and applies middleware per group.

**[Visual: Show request validation pattern]**

**Narrator**: Request validation uses Gin's `ShouldBindJSON` for struct binding and the `go-playground/validator` package for field-level validation. Invalid requests get a 400 response immediately, before any business logic runs.

**[Visual: Show response formatting]**

**Narrator**: Responses use a consistent JSON structure. Success responses return the data directly. Error responses always include an `"error"` key with a human-readable message. This consistency makes the API predictable for frontend consumers.

**[Visual: Course title card]**

**Narrator**: That wraps up the backend core. You have seen the full request lifecycle: from the Gin handler, through the service layer, down to the dialect-aware database. In Module 3, we will build on this foundation to implement authentication and authorization.

---

## Key Code Examples

### Directory Layout
```
catalog-api/
  main.go                     # Entry point, wiring, graceful shutdown
  handlers/                   # Domain HTTP handlers
  services/                   # Domain business logic
  repository/                 # Domain data access
  middleware/                  # Domain middleware (auth, rate limiting)
  database/                   # Dialect abstraction, migrations
  filesystem/                 # Protocol client interface and factory
  models/                     # Shared data models
  internal/
    auth/                     # Infrastructure auth (JWT, sessions)
    handlers/                 # Infrastructure handlers
    services/                 # Infrastructure services (scanner, aggregation)
    media/                    # Media detection pipeline
    metrics/                  # Prometheus metrics, health checks
    smb/                      # SMB resilience (circuit breaker, offline cache)
```

### Dialect-Aware Database Wrapper
```go
// database/connection.go
type DB struct {
    *sql.DB
    config  *config.DatabaseConfig
    dialect Dialect
}

// Shadowed methods auto-rewrite SQL for the target dialect
func (db *DB) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
    rewritten := db.dialect.RewritePlaceholders(query)
    rewritten = db.dialect.RewriteInsertOrIgnore(rewritten)
    rewritten = db.dialect.BooleanLiterals(rewritten)
    return db.DB.QueryRowContext(ctx, rewritten, args...)
}
```

### Test Database Setup
```go
// internal/tests/test_helper.go
func SetupTestDB(t *testing.T) *database.DB {
    sqlDB, err := sql.Open("sqlite3", ":memory:")
    require.NoError(t, err)
    db := database.WrapDB(sqlDB, database.DialectSQLite)
    // Run migrations against in-memory DB
    return db
}
```

---

## Quiz Questions

1. What are the two SQL dialects supported by the database layer, and how does the `Dialect` struct handle differences between them?
   **Answer**: SQLite and PostgreSQL. The `Dialect` struct provides `RewritePlaceholders` (? to $N), `RewriteInsertOrIgnore` (INSERT OR IGNORE to ON CONFLICT DO NOTHING), and `BooleanLiterals` (0/1 to FALSE/TRUE) methods that automatically transform SQL at query time.

2. Why does Catalogizer write the server port to a `.service-port` file on startup?
   **Answer**: Because the backend uses dynamic port binding (it finds the first available port). The frontend dev server reads this file to configure its API proxy target, enabling automatic backend discovery without hardcoded ports.

3. What is the purpose of the dual package layout (top-level vs `internal/`) in catalog-api?
   **Answer**: Top-level packages (`handlers/`, `services/`, `repository/`) contain domain business logic. The `internal/` directory contains infrastructure concerns (auth, media detection, metrics, SMB resilience). This separation keeps domain code clean and prevents external packages from importing infrastructure internals.

4. How does the migration system ensure idempotency?
   **Answer**: Each migration has a unique version number. Before running a migration, the system checks the `migrations` table for that version. If it has already been applied, it is skipped. Each migration also has separate SQLite and PostgreSQL implementations.
