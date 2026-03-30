# Catalogizer Architecture Documentation

## Overview

Catalogizer is built using a modern microservices-inspired architecture with clear separation of concerns, robust error handling, and high availability in mind. The system is designed to handle media collection management at scale while maintaining resilience against various failure modes.

## System Architecture

### Mermaid System Architecture Diagram

```mermaid
graph TB
    subgraph Clients["Client Applications"]
        CW["catalog-web<br/>(React/TypeScript)"]
        CD["catalogizer-desktop<br/>(Tauri/Rust + React)"]
        IW["installer-wizard<br/>(Tauri/Rust + React)"]
        CA["catalogizer-android<br/>(Kotlin/Compose)"]
        CTV["catalogizer-androidtv<br/>(Kotlin/Compose)"]
        ACL["catalogizer-api-client<br/>(TypeScript Library)"]
    end

    subgraph Infrastructure["Infrastructure"]
        LB["Load Balancer"]
        RP["Reverse Proxy (Nginx)"]
    end

    subgraph Backend["catalog-api (Go/Gin)"]
        API["REST API /api/v1/*"]
        WS["WebSocket Server /ws/*"]
        AUTH["Auth Service (JWT + Sessions)"]
        MW["Middleware Stack<br/>CORS, Rate Limit,<br/>Input Validation, Metrics"]
    end

    subgraph CoreServices["Core Services"]
        CAT["Catalog Service"]
        SMB_SVC["SMB Service"]
        SMB_DISC["SMB Discovery"]
        MEDIA_REC["Media Recognition"]
        DUP["Duplicate Detection"]
        REC["Recommendation Service"]
        SUB["Subtitle Service"]
        CONV["Conversion Service"]
        ANALYTICS["Analytics Service"]
        CONFIG["Configuration Service"]
        ERR_SVC["Error/Crash Reporting"]
        LOG_SVC["Log Management"]
        FAV["Favorites Service"]
    end

    subgraph DataLayer["Data Layer"]
        SQLITE["SQLite/SQLCipher<br/>(Encrypted DB)"]
        REDIS["Redis<br/>(Rate Limiting / Cache)"]
    end

    subgraph FileWatch["Real-time File Monitoring"]
        WATCHER["SMBChangeWatcher"]
        ENH["EnhancedChangeWatcher<br/>(rename detection)"]
        QUEUE["Change Event Queue"]
    end

    subgraph Providers["External Metadata Providers"]
        TMDB["TMDB"]
        IMDB["IMDB"]
        TVDB["TVDB"]
        SPOTIFY["Spotify"]
        STEAM["Steam"]
    end

    subgraph Storage["Multi-Protocol Storage"]
        LOCAL["Local FS"]
        SMB_P["SMB/CIFS"]
        FTP["FTP"]
        NFS["NFS"]
        WEBDAV["WebDAV"]
    end

    CW & CD & CA & CTV & ACL & IW -->|HTTP/WS| LB
    LB --> RP
    RP --> API & WS

    API --> MW --> AUTH
    API --> CAT & SMB_SVC & SMB_DISC
    API --> MEDIA_REC & REC & SUB & CONV
    API --> ANALYTICS & CONFIG & ERR_SVC & LOG_SVC & FAV

    CAT & AUTH & CONV & ANALYTICS --> SQLITE
    AUTH --> REDIS
    MEDIA_REC --> TMDB & IMDB & TVDB & SPOTIFY & STEAM
    WATCHER & ENH --> QUEUE
    QUEUE --> MEDIA_REC
    WS --> QUEUE
    CAT --> LOCAL & SMB_P & FTP & NFS & WEBDAV

    style Backend fill:#e1f5fe
    style DataLayer fill:#fff3e0
    style Providers fill:#f3e5f5
    style Storage fill:#e8f5e9
    style Clients fill:#fce4ec
    style FileWatch fill:#fff9c4
```

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        Load Balancer                           │
└─────────────────┬───────────────────────────────────────────────┘
                  │
┌─────────────────▼───────────────────────────────────────────────┐
│                   Reverse Proxy (Nginx)                        │
└─────────────┬─────────────────────────────────┬─────────────────┘
              │                                 │
┌─────────────▼─────────────┐         ┌────────▼─────────────┐
│      Frontend (React)     │         │   Backend (Go API)   │
│                           │         │                      │
│  • TypeScript             │         │  • Gin Framework     │
│  • Tailwind CSS           │         │  • JWT Auth          │
│  • React Query            │         │  • SQLCipher DB      │
│  • WebSocket Client       │         │  • Circuit Breakers  │
│  • Real-time Updates      │         │  • Retry Logic       │
└───────────────────────────┘         └──────┬───────────────┘
                                             │
         ┌────────────────────────────────────┼────────────────┐
         │                                    │                │
┌───────▼────────┐  ┌─────────▼──────────┐  ┌▼──────────────┐
│ Media Detection│  │   SMB Resilience   │  │ External APIs │
│                │  │                    │  │               │
│ • Pattern      │  │ • Auto-reconnect   │  │ • TMDB        │
│   Recognition  │  │ • Offline Cache    │  │ • IMDB        │
│ • Quality      │  │ • Health Checks    │  │ • Spotify     │
│   Analysis     │  │ • Circuit Breaker  │  │ • Steam       │
│ • Metadata     │  │ • Retry Logic      │  │ • Rate Limit  │
│   Extraction   │  │ • Event Streaming  │  │ • Fallback    │
└────────────────┘  └────────────────────┘  └───────────────┘
         │                     │                      │
         └─────────────────────┼──────────────────────┘
                               │
          ┌─────────────────────▼─────────────────────┐
          │         Multi-Protocol Sources           │
          │                                           │
          │  ┌─────────────┐  ┌─────────────┐        │
          │  │   Server 1  │  │   Server 2  │  ...   │
          │  │   /media    │  │   /backup   │        │
          │  └─────────────┘  └─────────────┘        │
          └───────────────────────────────────────────┘
```

#### System Architecture Diagram (Drawio/PlantUML)

The following PlantUML diagram provides a visual representation of the high-level system architecture. This diagram can be imported into Drawio for editing.

```plantuml
@startuml System Architecture
skinparam componentStyle uml2
skinparam backgroundColor #FEFEFE
skinparam component {
    BackgroundColor #FFFFFF
    BorderColor #000000
    BorderThickness 2
}

[Load Balancer] as LB #LightBlue
[Reverse Proxy\n(Nginx)] as RP #LightGreen
[Frontend\n(React TypeScript)] as FE #LightYellow
[Backend\n(Go API)] as BE #LightCyan
[Media Detection] as MD #LightPink
[SMB Resilience] as SMB #LightGray
[External APIs] as EA #LightSalmon
[Multi-Protocol Sources] as MPS #LightCoral

LB --> RP : HTTP/HTTPS
RP --> FE : Static Assets
RP --> BE : API Calls
BE --> MD : Media Processing
BE --> SMB : SMB Operations
BE --> EA : Metadata Fetch
MD --> MPS : File Access
SMB --> MPS : Network Access
EA --> MPS : Indirect Access

note right of FE : TypeScript, Tailwind CSS,\nReact Query, WebSocket
note right of BE : Gin Framework, JWT Auth,\nSQLCipher DB, Circuit Breakers
note bottom of MD : Pattern Recognition,\nQuality Analysis, Metadata Extraction
note bottom of SMB : Auto-reconnect, Offline Cache,\nHealth Checks, Circuit Breaker
note bottom of EA : TMDB, IMDB, Spotify, Steam,\nRate Limiting, Fallback
note bottom of MPS : SMB, FTP, Local FS,\n/media, /backup directories
@enduml
```

### Component Details

#### 1. Frontend Layer (React TypeScript)

**Technology Stack:**
- React 18 with TypeScript
- Vite for build tooling
- Tailwind CSS for styling
- React Query for state management
- Framer Motion for animations
- WebSocket for real-time updates

**Key Components:**
```typescript
// Authentication & Security
AuthContext              // JWT token management
ProtectedRoute          // Route-level authorization
UserManagement          // Admin user controls

// Media Management
MediaBrowser            // Grid/list view of media
MediaCard              // Individual media display
MediaFilters           // Advanced filtering
MediaSearch            // Search functionality

// Real-time Features
WebSocketContext       // WebSocket connection management
ConnectionStatus       // Connection health indicator
RealTimeUpdates        // Live data synchronization

// Analytics & Monitoring
Dashboard              // Overview metrics
Analytics              // Detailed statistics
HealthMonitor          // System health display
```

**State Management:**
```typescript
// React Query Cache Structure
{
  'media-search': MediaSearchResponse,
  'media-stats': MediaStatistics,
  'auth-status': AuthenticationStatus,
  'smb-status': SMBSourceStatus,
  'health-check': SystemHealth
}
```

**Frontend Component Diagram (Drawio/PlantUML)**

```plantuml
@startuml Frontend Components
skinparam componentStyle uml2
skinparam backgroundColor #FEFEFE

package "Authentication & Security" as Auth {
    [AuthContext] as AC
    [ProtectedRoute] as PR
    [UserManagement] as UM
}

package "Media Management" as Media {
    [MediaBrowser] as MB
    [MediaCard] as MC
    [MediaFilters] as MF
    [MediaSearch] as MS
}

package "Real-time Features" as RT {
    [WebSocketContext] as WSC
    [ConnectionStatus] as CS
    [RealTimeUpdates] as RTU
}

package "Analytics & Monitoring" as AM {
    [Dashboard] as D
    [Analytics] as A
    [HealthMonitor] as HM
}

AC --> PR
PR --> UM
MB --> MC
MB --> MF
MB --> MS
WSC --> CS
WSC --> RTU
D --> A
D --> HM

note right of AC : JWT token management
note right of MB : Grid/list view of media
note right of WSC : WebSocket connection management
note right of D : Overview metrics
@enduml
```

#### 2. Backend Layer (Go API)

**Technology Stack:**
- Go 1.21+ with Gin framework
- SQLCipher for encrypted database
- JWT for authentication
- WebSocket for real-time communication
- Circuit breakers for fault tolerance
- Structured logging with Zap

**Service Architecture:**
```go
// Core Services
AuthService             // Authentication & authorization
MediaService           // Media detection & management
SMBManager             // SMB source management
AnalyticsService       // Statistics & reporting
WebSocketService       // Real-time communication

// Infrastructure Services
CircuitBreaker         // Fault tolerance
RetryManager           // Automatic retry logic
HealthChecker          // System health monitoring
CacheManager           // Offline caching
EventStreamer          // Event publishing
```

**Database Schema:**
```sql
-- Core Tables
users                  -- User accounts & permissions
media_items            -- Media metadata & file info
external_metadata      -- Third-party API data
media_versions         -- Multiple quality versions
smb_sources           -- SMB connection configurations

-- Monitoring Tables
health_checks         -- System health history
error_logs           -- Error tracking & analysis
performance_metrics  -- System performance data
audit_logs          -- User activity tracking
```

**Backend Service Diagram (Drawio/PlantUML)**

```plantuml
@startuml Backend Services
skinparam componentStyle uml2
skinparam backgroundColor #FEFEFE

package "Core Services" as Core {
    [AuthService] as AS
    [MediaService] as MS
    [SMBManager] as SM
    [AnalyticsService] as AnS
    [WebSocketService] as WSS
}

package "Infrastructure Services" as Infra {
    [CircuitBreaker] as CB
    [RetryManager] as RM
    [HealthChecker] as HC
    [CacheManager] as CM
    [EventStreamer] as ES
}

AS --> MS
MS --> SM
MS --> AnS
MS --> WSS
SM --> CB
SM --> RM
SM --> HC
SM --> CM
SM --> ES

note right of AS : Authentication & authorization
note right of MS : Media detection & management
note right of SM : SMB source management
note right of CB : Fault tolerance
note right of HC : System health monitoring
@enduml
```

#### 3. SMB Resilience Layer

The SMB resilience layer is a critical component that ensures the system remains functional even when SMB sources are temporarily unavailable.

**Key Features:**
- **Automatic Reconnection**: Exponential backoff retry strategy
- **Offline Caching**: Local cache for metadata when sources are down
- **Health Monitoring**: Continuous health checks with alerting
- **Circuit Breakers**: Prevent cascade failures
- **Event-Driven Architecture**: Real-time status updates

```go
// SMB Resilience Components
type ResilientSMBManager struct {
    sources      map[string]*SMBSource
    offlineCache *OfflineCache
    healthChecker *HealthChecker
    circuitBreaker *CircuitBreaker
    eventBus     *EventBus
}

// Connection States
StateClosed       // Normal operation
StateHalfOpen     // Testing after failure
StateOpen         // Circuit breaker activated
StateOffline      // Extended failure mode
```

**SMB Resilience Diagram (Drawio/PlantUML)**

```plantuml
@startuml SMB Resilience
skinparam componentStyle uml2
skinparam backgroundColor #FEFEFE

class ResilientSMBManager {
    +sources: map[string]*SMBSource
    +offlineCache: *OfflineCache
    +healthChecker: *HealthChecker
    +circuitBreaker: *CircuitBreaker
    +eventBus: *EventBus
}

class SMBSource {
    +host: string
    +share: string
    +credentials: Credentials
}

class OfflineCache {
    +entries: map[string]*CacheEntry
    +maxSize: int
    +eviction: EvictionPolicy
}

class HealthChecker {
    +checkInterval: time.Duration
    +timeout: time.Duration
}

class CircuitBreaker {
    +state: CircuitState
    +failures: int
    +maxFailures: int
    +resetTimeout: time.Duration
}

class EventBus {
    +subscribers: []Subscriber
}

ResilientSMBManager --> SMBSource : manages
ResilientSMBManager --> OfflineCache : uses
ResilientSMBManager --> HealthChecker : uses
ResilientSMBManager --> CircuitBreaker : uses
ResilientSMBManager --> EventBus : uses

note right of ResilientSMBManager : Main SMB management component
note bottom of CircuitBreaker : Closed/Open/Half-Open states
@enduml
```

#### 4. Media Detection Engine

**Detection Pipeline:**
```
File System Event → Pattern Analysis → Media Type Detection → Quality Analysis → Metadata Extraction → Database Storage
```

**Pattern Recognition:**
```go
// Movie Detection Patterns
moviePatterns := []Pattern{
    {Regex: `^(.+?)\s*[\(\[]?(\d{4})[\)\]]?.*\.(mkv|mp4|avi)$`, Weight: 0.9},
    {Regex: `^(.+?)\s*-\s*(\d{4}).*\.(mkv|mp4|avi)$`, Weight: 0.8},
}

// TV Show Patterns
tvPatterns := []Pattern{
    {Regex: `^(.+?)\s*[Ss](\d{1,2})[Ee](\d{1,2}).*$`, Weight: 0.95},
    {Regex: `^(.+?)\s*(\d{1,2})x(\d{1,2}).*$`, Weight: 0.85},
}
```

**Quality Analysis:**
```go
type QualityMetrics struct {
    Resolution    string  // 720p, 1080p, 4K, etc.
    Codec         string  // H.264, H.265, AV1
    Bitrate       int     // Video bitrate
    AudioCodec    string  // AAC, DTS, Dolby
    FileSize      int64   // File size in bytes
    OverallScore  float64 // Calculated quality score
}
```

**Media Detection Pipeline Diagram (Drawio/PlantUML)**

```plantuml
@startuml Media Detection Pipeline
skinparam activity {
    BackgroundColor #FEFEFE
    BorderColor #000000
    BorderThickness 2
}

start
:File System Event;
:Pattern Analysis;
if (Media Type Detected?) then (yes)
    :Quality Analysis;
    :Metadata Extraction;
    :Database Storage;
    :WebSocket Notification;
    :Frontend Update;
else (no)
    :Skip Processing;
endif
stop
@enduml
```

#### 5. External API Integration

**Provider Architecture:**
```go
type MetadataProvider interface {
    Search(query string) ([]Metadata, error)
    GetDetails(id string) (*DetailedMetadata, error)
    GetRateLimit() *RateLimit
}

// Implemented Providers
TMDBProvider     // The Movie Database
IMDBProvider     // Internet Movie Database
TVDBProvider     // TheTVDB
SpotifyProvider  // Music metadata
SteamProvider    // Game information
```

**Rate Limiting & Fallback:**
```go
type ProviderManager struct {
    providers    []MetadataProvider
    rateLimiter  *RateLimiter
    circuitBreaker *CircuitBreaker
    cache        *MetadataCache
}
```

**External API Integration Diagram (Drawio/PlantUML)**

```plantuml
@startuml External API Integration
skinparam componentStyle uml2
skinparam backgroundColor #FEFEFE

interface MetadataProvider {
    +Search(query: string): []Metadata
    +GetDetails(id: string): *DetailedMetadata
    +GetRateLimit(): *RateLimit
}

class TMDBProvider
class IMDBProvider
class TVDBProvider
class SpotifyProvider
class SteamProvider

class ProviderManager {
    +providers: []MetadataProvider
    +rateLimiter: *RateLimiter
    +circuitBreaker: *CircuitBreaker
    +cache: *MetadataCache
}

class RateLimiter
class CircuitBreaker
class MetadataCache

MetadataProvider <|.. TMDBProvider
MetadataProvider <|.. IMDBProvider
MetadataProvider <|.. TVDBProvider
MetadataProvider <|.. SpotifyProvider
MetadataProvider <|.. SteamProvider

ProviderManager --> MetadataProvider : uses
ProviderManager --> RateLimiter : uses
ProviderManager --> CircuitBreaker : uses
ProviderManager --> MetadataCache : uses

note right of TMDBProvider : The Movie Database
note right of SpotifyProvider : Music metadata
note bottom of ProviderManager : Manages multiple providers\nwith rate limiting and caching
@enduml
```

## Data Flow

### 1. Media Discovery Flow

```
SMB Source Change → File System Watcher → Event Queue → Media Detector →
Pattern Analysis → External API Lookup → Quality Analysis → Database Storage →
WebSocket Notification → Frontend Update
```

### 2. User Authentication Flow

```
Login Request → JWT Validation → Permission Check → Database Query →
Token Generation → Client Storage → Authenticated Session
```

### 3. Real-time Update Flow

```
System Event → Event Bus → WebSocket Server → Connected Clients →
React Query Cache Invalidation → UI Update
```

### 4. SMB Failure Recovery Flow

```
SMB Connection Lost → Circuit Breaker Opens → Offline Mode Activated →
Background Reconnection Attempts → Connection Restored →
Cache Synchronization → Normal Operation Resumed
```

## Resilience Patterns

### 1. Circuit Breaker Pattern

```go
type CircuitBreaker struct {
    state        CircuitState  // Closed, Open, Half-Open
    failures     int
    maxFailures  int
    resetTimeout time.Duration
}

// Usage in SMB operations
func (s *SMBService) ListDirectory(path string) error {
    return s.circuitBreaker.Execute(func() error {
        return s.actualListDirectory(path)
    })
}
```

### 2. Retry Pattern with Exponential Backoff

```go
type RetryConfig struct {
    MaxAttempts   int
    InitialDelay  time.Duration
    BackoffFactor float64
    Jitter        bool
}

// Automatic retry for transient failures
func (s *SMBService) ConnectWithRetry(source *SMBSource) error {
    return Retry(context.Background(), s.retryConfig, func() error {
        return s.connect(source)
    })
}
```

### 3. Bulkhead Pattern

```go
type Bulkhead struct {
    semaphore chan struct{}  // Limit concurrent operations
    timeout   time.Duration
}

// Isolate different resource pools
smbBulkhead := NewBulkhead(BulkheadConfig{
    MaxConcurrent: 10,
    Timeout:      30 * time.Second,
})
```

### 4. Cache-Aside Pattern

```go
type OfflineCache struct {
    entries   map[string]*CacheEntry
    maxSize   int
    eviction  EvictionPolicy
}

// Cache metadata when sources are unavailable
func (c *OfflineCache) GetOrFetch(key string, fetcher func() interface{}) interface{} {
    if value, exists := c.Get(key); exists {
        return value
    }

    value := fetcher()
    c.Set(key, value)
    return value
}
```

## Security Architecture

### 1. Authentication & Authorization

```go
// JWT Claims Structure
type Claims struct {
    UserID      int64    `json:"user_id"`
    Username    string   `json:"username"`
    Role        string   `json:"role"`
    Permissions []string `json:"permissions"`
    jwt.RegisteredClaims
}

// Permission-based Access Control
func RequirePermission(permission string) gin.HandlerFunc {
    return func(c *gin.Context) {
        user := GetCurrentUser(c)
        if !user.HasPermission(permission) {
            c.JSON(403, gin.H{"error": "Insufficient permissions"})
            c.Abort()
            return
        }
        c.Next()
    }
}
```

### 2. Database Security

```sql
-- SQLCipher Configuration
PRAGMA cipher_page_size = 4096;
PRAGMA kdf_iter = 64000;
PRAGMA cipher_hmac_algorithm = HMAC_SHA512;
PRAGMA cipher_kdf_algorithm = PBKDF2_HMAC_SHA512;
```

### 3. Input Validation & Sanitization

```go
// Request validation with binding tags
type MediaSearchRequest struct {
    Query    string `json:"query" binding:"max=255"`
    Type     string `json:"type" binding:"oneof=movie tv_show music game"`
    Year     int    `json:"year" binding:"min=1900,max=2030"`
    Rating   float64 `json:"rating" binding:"min=0,max=10"`
}
```

## Performance Optimizations

### 1. Database Indexing Strategy

```sql
-- High-performance indexes
CREATE INDEX idx_media_type_year ON media_items(media_type, year);
CREATE INDEX idx_media_rating_desc ON media_items(rating DESC);
CREATE INDEX idx_media_updated_desc ON media_items(updated_at DESC);

-- Full-text search index
CREATE VIRTUAL TABLE media_search USING fts5(
    title, description,
    content='media_items',
    content_rowid='id'
);
```

### 2. Caching Strategy

```go
// Multi-level caching
type CacheManager struct {
    l1Cache    *sync.Map         // In-memory cache
    l2Cache    *redis.Client     // Redis cache (if available)
    dbCache    *sql.DB           // Database cache
}

// Cache hierarchy
Memory Cache (L1) → Redis Cache (L2) → Database → External APIs
```

### 3. Connection Pooling

```go
// Database connection pool
db.SetMaxOpenConns(25)
db.SetMaxIdleConns(5)
db.SetConnMaxLifetime(5 * time.Minute)

// HTTP client pool for external APIs
client := &http.Client{
    Transport: &http.Transport{
        MaxIdleConns:        100,
        MaxIdleConnsPerHost: 10,
        IdleConnTimeout:     90 * time.Second,
    },
    Timeout: 30 * time.Second,
}
```

## Monitoring & Observability

### 1. Health Checks

```go
// Comprehensive health monitoring
healthChecker.AddCheck(HealthCheck{
    Name:     "database",
    Check:    func(ctx context.Context) error { return db.PingContext(ctx) },
    Critical: true,
})

healthChecker.AddCheck(HealthCheck{
    Name:     "smb_sources",
    Check:    checkSMBSources,
    Critical: false,
})
```

### 2. Metrics Collection

```go
// Performance metrics
type Metrics struct {
    RequestDuration   histogram
    RequestCount      counter
    ActiveConnections gauge
    ErrorRate         counter
}
```

### 3. Structured Logging

```go
// Contextual logging with Zap
logger.Info("Media item processed",
    zap.String("title", media.Title),
    zap.String("type", media.Type),
    zap.Duration("processing_time", duration),
    zap.Int64("file_size", media.FileSize))
```

## Scalability Considerations

### 1. Horizontal Scaling

- **Stateless Design**: All services are stateless for easy scaling
- **Load Balancing**: Round-robin distribution of requests
- **Database Replication**: Read replicas for query scaling
- **Microservice Architecture**: Independent service scaling

### 2. Vertical Scaling

- **Resource Optimization**: Efficient memory and CPU usage
- **Connection Pooling**: Optimal database connection management
- **Caching**: Reduced database load through intelligent caching
- **Async Processing**: Non-blocking operations where possible

### 3. Data Partitioning

```sql
-- Partition large tables by date
CREATE TABLE media_items_2024 PARTITION OF media_items
FOR VALUES FROM ('2024-01-01') TO ('2025-01-01');

-- Index partitioning for better query performance
CREATE INDEX idx_media_2024_type ON media_items_2024(media_type);
```

This architecture provides a robust, scalable, and resilient foundation for the Catalogizer system, with comprehensive error handling and recovery mechanisms to ensure high availability even in the face of various failure scenarios.

## Lazy Initialization Pattern

### LazyServiceRegistry

The `LazyServiceRegistry` in `internal/lifecycle/` provides deferred service initialization with dependency ordering. Services are registered with their constructors but not instantiated until first use. This reduces startup time and memory consumption by avoiding initialization of services that may never be called during a given request lifecycle.

```go
// internal/lifecycle/lazy_registry.go
type LazyServiceRegistry struct {
    mu        sync.RWMutex
    factories map[string]func() (interface{}, error)
    instances map[string]interface{}
}

func (r *LazyServiceRegistry) Register(name string, factory func() (interface{}, error)) {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.factories[name] = factory
}

func (r *LazyServiceRegistry) Get(name string) (interface{}, error) {
    r.mu.RLock()
    if inst, ok := r.instances[name]; ok {
        r.mu.RUnlock()
        return inst, nil
    }
    r.mu.RUnlock()

    r.mu.Lock()
    defer r.mu.Unlock()
    // Double-check after acquiring write lock
    if inst, ok := r.instances[name]; ok {
        return inst, nil
    }
    factory, ok := r.factories[name]
    if !ok {
        return nil, fmt.Errorf("service %q not registered", name)
    }
    inst, err := factory()
    if err != nil {
        return nil, fmt.Errorf("failed to initialize %q: %w", name, err)
    }
    r.instances[name] = inst
    return inst, nil
}
```

**Key properties:**
- Thread-safe double-checked locking pattern avoids redundant initialization.
- Services are only created when `Get()` is first called.
- Factory functions capture dependencies via closures, enabling natural dependency ordering.
- Used in `main.go` to register metadata providers, the aggregation service, and optional services like Redis caching.

### LazyProvider

The `LazyProvider` pattern extends lazy initialization to metadata providers (TMDB, OMDB, OpenLibrary, MusicBrainz). Each provider is wrapped in a `LazyProvider` that defers API client creation until the first metadata lookup. If a provider's API key is missing, the provider is never instantiated, saving memory and avoiding startup errors.

## BoundedSemaphore Concurrency Control

The `BoundedSemaphore` in `internal/concurrency/` limits the number of concurrent operations system-wide. Unlike a simple channel-based semaphore, `BoundedSemaphore` supports context-aware acquisition with timeout, preventing goroutine leaks when the system is under pressure.

```go
// internal/concurrency/semaphore.go
type BoundedSemaphore struct {
    sem chan struct{}
}

func NewBoundedSemaphore(limit int) *BoundedSemaphore {
    return &BoundedSemaphore{sem: make(chan struct{}, limit)}
}

func (s *BoundedSemaphore) Acquire(ctx context.Context) error {
    select {
    case s.sem <- struct{}{}:
        return nil
    case <-ctx.Done():
        return ctx.Err()
    }
}

func (s *BoundedSemaphore) Release() {
    <-s.sem
}
```

**Usage:** The `SearchAll` function in the media entity service uses `BoundedSemaphore` to limit concurrent provider searches to a configurable maximum (default: 5), preventing API rate limit exhaustion and excessive goroutine creation.

## Non-Blocking Health Check Pattern

The `/health/deep` endpoint performs a comprehensive health check of all subsystems (database, Redis, filesystem, metadata providers) with a 100ms timeout per check. This ensures the health endpoint never blocks the caller even when a subsystem is unresponsive.

```go
// Health deep check with per-component timeout
func (h *HealthHandler) DeepHealthCheck(c *gin.Context) {
    ctx, cancel := context.WithTimeout(c.Request.Context(), 100*time.Millisecond)
    defer cancel()

    checks := map[string]func(context.Context) error{
        "database": h.checkDatabase,
        "redis":    h.checkRedis,
        "storage":  h.checkStorage,
    }

    results := make(map[string]string)
    for name, check := range checks {
        if err := check(ctx); err != nil {
            results[name] = "degraded"
        } else {
            results[name] = "healthy"
        }
    }
    c.JSON(http.StatusOK, gin.H{"status": overallStatus(results), "checks": results})
}
```

**Design rationale:** Load balancers and orchestrators poll health endpoints frequently. A slow health check can cascade into false-positive failures. The 100ms timeout ensures the endpoint always responds promptly, marking slow subsystems as "degraded" rather than blocking.

## Admin Handler Architecture

The `AdminHandler` in `internal/handlers/admin_handler.go` provides privileged system administration operations. It is wired into the router under `/api/v1/admin` with admin-role middleware applied to the entire group.

**Capabilities:**
- **System info**: Runtime memory statistics, goroutine count, version, uptime, database pool stats.
- **User management**: List all users with roles, update user roles, activate/deactivate/lock accounts.
- **Storage overview**: Aggregated storage root statistics with file counts and total sizes.
- **Backup/restore**: Create database backups via SQLite `VACUUM INTO`, list available backups, restore from backup.
- **Scan triggering**: Initiate storage root scans from the admin panel without navigating to the storage management page.

**Security model:** All admin endpoints require the `admin` role (checked by `RequireRole("admin")` middleware). The handler validates that admin users cannot lock their own account to prevent self-lockout scenarios. Backup operations are audited in the `auth_audit_log` table.

## Safety Improvements

### Database Query Timeout (30s Default)

All database queries are executed with a default 30-second context timeout. This prevents runaway queries from holding database connections indefinitely, which could exhaust the connection pool.

```go
// Applied at the DB wrapper level
func (db *DB) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
    if _, hasDeadline := ctx.Deadline(); !hasDeadline {
        var cancel context.CancelFunc
        ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
        defer cancel()
    }
    // ... dialect rewriting and execution
}
```

### Redis Middleware Timeout (500ms)

The Redis rate limiter middleware uses a 500ms timeout for all Redis operations. If Redis is slow or unreachable, the middleware allows the request to proceed rather than blocking it, implementing a fail-open pattern for availability.

```go
// middleware/redis_rate_limiter.go
ctx, cancel := context.WithTimeout(c.Request.Context(), 500*time.Millisecond)
defer cancel()
// If Redis check fails, allow the request through
```

## Module Registry

### Overview

The Module Registry is a centralized initialization system that wires 12 independent Go modules into catalog-api at startup. Each module is a separate git submodule with its own repository, tests, and documentation. The registry ensures that all modules are initialized in the correct dependency order and that their services are available to handlers throughout the application lifecycle.

### How It Integrates

At application startup, `main.go` calls `RegisterModules()` which initializes each module's services and registers them with the `LazyServiceRegistry`. Modules that depend on other modules declare those dependencies through constructor injection -- the registry's factory closures capture references to already-initialized dependencies, ensuring natural ordering without a separate dependency graph resolver.

```go
// main.go (simplified)
func main() {
    db := database.Connect(cfg)
    registry := lifecycle.NewLazyServiceRegistry()

    RegisterModules(registry, db, cfg)

    // Handlers receive the registry and pull services on demand
    handler := handlers.NewMediaHandler(registry)
    // ...
}
```

### Registered Modules and Services

| Module | Submodule Path | Service Provided | Purpose |
|--------|---------------|------------------|---------|
| Memory | `Memory/` | Memory monitor | Heap tracking, leak detection, GC pressure alerts |
| Observability | `Observability/` | Health aggregator | Unified health status across all subsystems |
| Recovery | `Recovery/` | Resilience facade | Circuit breakers, retry policies, fallback chains |
| Security | `Security/` | Guardrail engine | Input validation, rate limiting rules, permission checks |
| Storage | `Storage/` | Storage resolver | Maps storage root configs to protocol-specific clients |
| Streaming | `Streaming/` | Transport factory | Creates streaming transports (HTTP range, WebSocket, HLS) |
| Media | `Media/` | Media detector | File pattern analysis, type detection, quality scoring |
| Database | `Database/` | Dialect manager | SQL rewriting, migration runner, connection pool stats |
| Cache | `Cache/` | Cache coordinator | Multi-level caching (memory, Redis, database) |
| Concurrency | `Concurrency/` | Semaphore pool | Bounded concurrency for parallel operations |
| Watcher | `Watcher/` | File watcher | Filesystem event monitoring with debounce |
| RateLimiter | `RateLimiter/` | Rate limiter | Token bucket rate limiting with configurable policies |

### Design Decisions

- **Lazy initialization**: Modules are not instantiated until first use. A module whose API key is missing or whose backing service is unavailable does not block startup.
- **Fail-open for optional modules**: If Redis is unavailable, the Cache module degrades to in-memory caching. If a metadata provider API key is missing, that provider is skipped.
- **Thread-safe access**: The `LazyServiceRegistry` uses double-checked locking (RWMutex) so concurrent handler goroutines can safely call `registry.Get("service-name")` without redundant initialization.
- **Testability**: In unit tests, modules can be replaced with mocks by registering alternative factory functions before the handler under test is created.

## Related Documentation

- [Database Schema](DATABASE_SCHEMA.md) - Complete database table and index reference
- [SQL Migrations](SQL_MIGRATIONS.md) - Migration versions, schema changes, and how to create new migrations
- [Auth Flow](AUTH_FLOW.md) - Authentication and authorization architecture
- [Go Backend Guide](GO_BACKEND_GUIDE.md) - Backend development patterns and conventions
- [React Frontend Guide](REACT_FRONTEND_GUIDE.md) - Frontend architecture and component patterns
- [Android Architecture](ANDROID_ARCHITECTURE.md) - MVVM architecture for mobile clients
- [Tauri IPC Guide](TAURI_IPC_GUIDE.md) - Desktop application IPC commands and events
- [API Documentation](../api/API_DOCUMENTATION.md) - REST API endpoint reference
- [WebSocket Events](../api/WEBSOCKET_EVENTS.md) - Real-time event bus documentation
- [Deployment Guide](../DEPLOYMENT_GUIDE.md) - Production deployment instructions
- [Monitoring Guide](../deployment/MONITORING_GUIDE.md) - Metrics and observability