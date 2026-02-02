# Go Backend Architecture Guide (catalog-api)

This guide covers the architecture, patterns, and conventions used in the `catalog-api` Go backend. It is written for developers joining the project who need to understand how the code is organized and how to extend it.

## Package Structure Overview

```
catalog-api/
├── main.go                      # Entry point: config loading, DI wiring, route registration, server lifecycle
├── config/
│   └── config.go                # Root-level configuration types (Config, ServerConfig, DatabaseConfig, etc.)
├── database/
│   ├── connection.go            # Database connection wrapper with migration support
│   └── migrations.go            # Schema migration definitions
├── models/
│   ├── user.go                  # User, Role, LoginRequest, Session models
│   ├── file.go                  # File-related models
│   └── media.go                 # Media-related models
├── repository/                  # Data access layer (raw SQL against SQLite)
│   ├── user_repository.go       # CRUD for users
│   ├── file_repository.go       # File index queries
│   ├── stats_repository.go      # Aggregate statistics queries
│   ├── conversion_repository.go # Conversion job persistence
│   ├── analytics_repository.go  # Analytics data access
│   ├── configuration_repository.go
│   ├── error_reporting_repository.go
│   ├── crash_reporting_repository.go
│   ├── log_management_repository.go
│   ├── favorites_repository.go
│   ├── sync_repository.go
│   └── stress_test_repository.go
├── services/                    # Business logic layer
│   ├── auth_service.go          # JWT auth, login, password hashing
│   ├── conversion_service.go    # Media format conversion jobs
│   ├── analytics_service.go     # Usage analytics
│   ├── reporting_service.go     # Reporting aggregation
│   ├── configuration_service.go # Runtime configuration management
│   ├── error_reporting_service.go
│   ├── log_management_service.go
│   ├── favorites_service.go
│   └── sync_service.go
├── handlers/                    # HTTP request handlers (Gin + standard http.Handler)
│   ├── auth_handler.go          # Auth endpoints (login, register, logout, etc.)
│   ├── conversion_handler.go    # Conversion job endpoints
│   ├── user_handler.go          # User CRUD endpoints
│   ├── role_handler.go          # Role management endpoints
│   ├── stats.go                 # Statistics endpoints
│   ├── browse.go                # File browsing endpoints
│   ├── search.go                # Search endpoints
│   ├── media_handler.go         # Media-related endpoints
│   ├── subtitle_handler.go      # Subtitle search/download/sync
│   ├── recommendation_handler.go
│   ├── configuration_handler.go
│   ├── error_reporting_handler.go
│   ├── log_management_handler.go
│   ├── service_adapters.go      # Adapter structs bridging handler<->service interfaces
│   └── gin_adapter.go           # WrapHTTPHandler helper for using http.Handler with Gin
├── middleware/                   # Gin middleware
│   ├── auth.go                  # JWTMiddleware - token parsing and validation
│   ├── input_validation.go      # Input sanitization and validation
│   ├── request.go               # RequestID middleware
│   ├── advanced_rate_limiter.go # Per-endpoint rate limiting
│   └── redis_rate_limiter.go    # Distributed rate limiting via Redis
├── filesystem/                  # Multi-protocol filesystem abstraction
│   ├── interface.go             # FileSystemClient interface + StorageConfig types
│   ├── factory.go               # DefaultClientFactory - creates per-protocol clients
│   ├── smb_client.go            # SMB/CIFS implementation
│   ├── ftp_client.go            # FTP implementation
│   ├── nfs_client.go            # NFS implementation (platform-specific builds)
│   ├── webdav_client.go         # WebDAV implementation
│   └── local_client.go          # Local filesystem implementation
├── internal/                    # Internal packages (not importable from outside module)
│   ├── auth/
│   │   ├── service.go           # Internal auth service (session-backed, for rate limiting)
│   │   ├── middleware.go         # Auth middleware with RBAC, rate limiting
│   │   └── models.go            # Internal User, Claims, LoginResponse types
│   ├── config/
│   │   └── config.go            # Internal configuration structs
│   ├── handlers/
│   │   ├── catalog.go           # Catalog browsing handler
│   │   ├── copy.go              # File copy handler
│   │   ├── download.go          # File download handler
│   │   ├── media.go             # Media operations handler
│   │   ├── smb.go               # SMB operations handler
│   │   ├── smb_discovery.go     # SMB network discovery
│   │   ├── localization_handlers.go
│   │   └── media_player_handlers.go
│   ├── media/
│   │   ├── detector/
│   │   │   └── engine.go        # Content type detection engine (rule-based)
│   │   ├── analyzer/
│   │   │   └── analyzer.go      # Media metadata analyzer
│   │   ├── providers/
│   │   │   └── providers.go     # External metadata providers (TMDB, IMDB, etc.)
│   │   ├── realtime/
│   │   │   └── enhanced_watcher.go  # Filesystem change watcher (fsnotify + debounce)
│   │   ├── database/
│   │   │   └── database.go      # Media-specific database operations
│   │   └── models/
│   │       └── media.go         # Media detection models
│   ├── services/
│   │   ├── catalog.go           # Internal catalog service (multi-protocol scanning)
│   │   ├── smb.go               # SMB-specific service logic
│   │   ├── smb_discovery.go     # SMB share discovery
│   │   ├── recommendation_service.go
│   │   ├── media_recognition_service.go
│   │   ├── subtitle_service.go
│   │   ├── cache_service.go
│   │   └── rename_tracker.go    # Intelligent file rename detection
│   ├── middleware/               # Internal middleware (logger, error handler)
│   └── metrics/                  # Prometheus metrics collection
└── utils/                       # Shared utility functions
```

## Service Layer Patterns

### Constructor Injection

Every service follows the constructor injection pattern. Dependencies are passed into a `NewXxxService` function that returns a pointer to the service struct:

```go
// From services/auth_service.go
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

All wiring happens in `main.go`:

```go
// main.go initialization sequence
userRepo := root_repository.NewUserRepository(db)
authService := root_services.NewAuthService(userRepo, jwtSecret)
authHandler := root_handlers.NewAuthHandler(authService)
```

### Error Wrapping

Errors are wrapped with context using `fmt.Errorf("...: %w", err)` to preserve the error chain:

```go
// From repository/user_repository.go
func (r *UserRepository) Create(user *models.User) (int, error) {
    result, err := r.db.Exec(query, ...)
    if err != nil {
        return 0, fmt.Errorf("failed to create user: %w", err)
    }
    id, err := result.LastInsertId()
    if err != nil {
        return 0, fmt.Errorf("failed to get user ID: %w", err)
    }
    return int(id), nil
}
```

Sentinel errors are used in the models package for common cases like `ErrUnauthorized`.

## Repository Layer Patterns

Repositories wrap `*sql.DB` and execute raw SQL queries against SQLite. Each repository focuses on one entity or domain:

```go
// From repository/user_repository.go
type UserRepository struct {
    db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
    return &UserRepository{db: db}
}

func (r *UserRepository) GetByID(id int) (*models.User, error) {
    query := `
        SELECT id, username, email, password_hash, salt, role_id, ...
        FROM users WHERE id = ?
    `
    user := &models.User{}
    err := r.db.QueryRow(query, id).Scan(&user.ID, &user.Username, ...)
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, errors.New("user not found")
        }
        return nil, fmt.Errorf("failed to get user: %w", err)
    }
    return user, nil
}
```

Key conventions:
- `*sql.DB` is the only direct dependency
- Queries use `?` placeholders (SQLite parameterized queries)
- `sql.ErrNoRows` is checked explicitly for "not found" cases
- `sql.NullString` and related types handle nullable columns
- `time.Now()` timestamps are set in the repository layer

## Handler Layer Patterns

The codebase uses two handler styles that coexist:

### 1. Gin-native handlers (preferred for new code)

```go
// From handlers/auth_handler.go
func (h *AuthHandler) LoginGin(c *gin.Context) {
    var req models.LoginRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
        return
    }
    result, err := h.authService.Login(req, c.ClientIP(), c.GetHeader("User-Agent"))
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, result)
}
```

### 2. Standard http.Handler wrapped with adapter

Older handlers use `http.ResponseWriter`/`*http.Request`, wrapped via `WrapHTTPHandler`:

```go
// handlers/gin_adapter.go provides the bridge:
wrap := root_handlers.WrapHTTPHandler
usersGroup.POST("", wrap(userHandler.CreateUser))
```

### Service Adapters

When a handler expects a different interface than the concrete service provides, adapter structs bridge the gap:

```go
// handlers/service_adapters.go
type AuthServiceAdapter struct {
    Inner *services.AuthService
}
// Implements the interface the handler expects by delegating to Inner
```

## How to Add a New Endpoint

### Step 1: Define the model (if needed)

Add request/response structs to `models/`:

```go
// models/media.go
type CreatePlaylistRequest struct {
    Name        string `json:"name" binding:"required"`
    Description string `json:"description"`
    MediaIDs    []int  `json:"media_ids"`
}
```

### Step 2: Add repository methods (if data access needed)

```go
// repository/playlist_repository.go
type PlaylistRepository struct {
    db *sql.DB
}

func NewPlaylistRepository(db *sql.DB) *PlaylistRepository {
    return &PlaylistRepository{db: db}
}

func (r *PlaylistRepository) Create(playlist *models.Playlist) (int, error) {
    // SQL INSERT
}
```

### Step 3: Add service logic

```go
// services/playlist_service.go
type PlaylistService struct {
    repo     *repository.PlaylistRepository
    authSvc  *AuthService
}

func NewPlaylistService(repo *repository.PlaylistRepository, authSvc *AuthService) *PlaylistService {
    return &PlaylistService{repo: repo, authSvc: authSvc}
}

func (s *PlaylistService) CreatePlaylist(userID int, req models.CreatePlaylistRequest) (*models.Playlist, error) {
    // Business logic + validation
}
```

### Step 4: Add handler

```go
// handlers/playlist_handler.go
type PlaylistHandler struct {
    service *services.PlaylistService
}

func NewPlaylistHandler(service *services.PlaylistService) *PlaylistHandler {
    return &PlaylistHandler{service: service}
}

func (h *PlaylistHandler) CreatePlaylist(c *gin.Context) {
    var req models.CreatePlaylistRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    // Extract user from context (set by auth middleware)
    userID := c.GetInt("user_id")
    result, err := h.service.CreatePlaylist(userID, req)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusCreated, result)
}
```

### Step 5: Wire up in main.go

```go
// main.go
playlistRepo := root_repository.NewPlaylistRepository(db)
playlistService := root_services.NewPlaylistService(playlistRepo, authService)
playlistHandler := root_handlers.NewPlaylistHandler(playlistService)

// Register routes (inside the authenticated api group)
playlistGroup := api.Group("/playlists")
{
    playlistGroup.POST("", playlistHandler.CreatePlaylist)
    playlistGroup.GET("", playlistHandler.ListPlaylists)
}
```

### Step 6: Add tests

Create `handlers/playlist_handler_test.go` beside the handler file:

```go
type PlaylistHandlerTestSuite struct {
    suite.Suite
    handler *PlaylistHandler
}

func (suite *PlaylistHandlerTestSuite) SetupTest() {
    // Setup test dependencies
}

func TestPlaylistHandlerSuite(t *testing.T) {
    suite.Run(t, new(PlaylistHandlerTestSuite))
}
```

## Configuration Patterns

Configuration follows a layered priority: **environment variables > config.json > defaults**.

### Config structure (`config/config.go`)

```go
type Config struct {
    Server   ServerConfig   `json:"server"`
    Database DatabaseConfig `json:"database"`
    Auth     AuthConfig     `json:"auth"`
    Catalog  CatalogConfig  `json:"catalog"`
    Storage  StorageConfig  `json:"storage"`
    Logging  LoggingConfig  `json:"logging"`
}
```

### Loading sequence in main.go

1. `LoadConfig("config.json")` reads file or creates with defaults
2. Environment variables override sensitive values:
   ```go
   if jwtSecret := os.Getenv("JWT_SECRET"); jwtSecret != "" {
       cfg.Auth.JWTSecret = jwtSecret
   }
   ```
3. Validation runs (`validateConfig`) to catch missing required fields

### Key environment variables

| Variable | Purpose | Required |
|----------|---------|----------|
| `JWT_SECRET` | JWT signing secret (min 32 chars) | Yes (production) |
| `ADMIN_USERNAME` | Initial admin username | Yes (production) |
| `ADMIN_PASSWORD` | Initial admin password | Yes (production) |
| `PORT` | Server port | No (default: 8080) |
| `GIN_MODE` | Gin framework mode (`release`/`debug`) | No |
| `REDIS_ADDR` | Redis address for distributed rate limiting | No |
| `REDIS_PASSWORD` | Redis password | No |

## Filesystem Abstraction

The `filesystem/` package provides a unified interface for all storage protocols:

```go
// filesystem/interface.go
type FileSystemClient interface {
    Connect(ctx context.Context) error
    Disconnect(ctx context.Context) error
    ReadFile(ctx context.Context, path string) (io.ReadCloser, error)
    WriteFile(ctx context.Context, path string, data io.Reader) error
    ListDirectory(ctx context.Context, path string) ([]*FileInfo, error)
    // ... more operations
    GetProtocol() string
}
```

To add a new protocol:

1. Create `filesystem/myprotocol_client.go` implementing `FileSystemClient`
2. Add a case in `filesystem/factory.go`:
   ```go
   case "myprotocol":
       config := &MyProtocolConfig{...}
       return NewMyProtocolClient(config), nil
   ```
3. Update `SupportedProtocols()` to include `"myprotocol"`

## Middleware Stack

Middleware is applied in `main.go` in this order:

1. **CORS** - Cross-origin resource sharing headers
2. **Metrics** - Prometheus request metrics
3. **Logger** - Structured request logging (zap)
4. **ErrorHandler** - Panic recovery and error normalization
5. **RequestID** - Adds `X-Request-ID` header
6. **InputValidation** - Sanitizes inputs against injection attacks

Route-level middleware:
- **RequireAuth** - JWT token validation (applied to `/api/v1/*`)
- **RateLimitByUser** - Per-user sliding window rate limiting
- **RequirePermission** / **RequireRole** - RBAC authorization

## Database

- **Engine**: SQLite with `go-sqlcipher` for encryption support
- **Migrations**: Run at startup via `databaseDB.RunMigrations(ctx)`
- **WAL mode**: Enabled by default for concurrent read performance
- **Connection**: `database.NewConnection()` wraps `*sql.DB` with migration support

## Testing Conventions

- Test files live beside source: `auth_handler.go` / `auth_handler_test.go`
- Use `testify/suite` for test suites and `testify/assert` for assertions
- Table-driven tests for multiple input scenarios
- Benchmark tests use `*_bench_test.go` naming
- Run all backend tests: `cd catalog-api && go test ./...`
- Run a single test: `go test -v -run TestName ./pkg/`
