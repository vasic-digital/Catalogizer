# Catalogizer Video Course - Complete Outline

## Course Overview

**Title**: Mastering Catalogizer - Building a Multi-Platform Media Collection Manager

**Duration**: 8-10 hours (across 12 modules)

**Audience**: Intermediate to advanced Go and TypeScript developers

**Prerequisites**:
- Go 1.24+ experience
- React 18+ and TypeScript knowledge
- Basic understanding of media processing concepts
- Familiarity with containerization (Podman/Docker)

---

## Module 1: Introduction and Architecture (45 min)

### Video 1.1: Course Overview (10 min)
- What is Catalogizer
- Course objectives and learning outcomes
- Repository structure overview
- Development environment setup

### Video 1.2: System Architecture (20 min)
- High-level architecture diagram
- Microservices vs monolithic decisions
- Protocol abstraction layer (SMB, FTP, NFS, WebDAV, Local)
- Database dual-dialect pattern (SQLite/PostgreSQL)

### Video 1.3: Submodule Architecture (15 min)
- 29 independent git submodules
- Go module replace directives
- TypeScript package linking
- Multi-remote push strategy

---

## Module 2: Backend Development - Core Services (60 min)

### Video 2.1: Project Structure and Main Entry Point (15 min)
- Directory layout conventions
- `main.go` structure and initialization
- Dependency injection patterns
- Graceful shutdown handling

### Video 2.2: Database Layer (20 min)
- Dialect abstraction (`database.DB` wrapper)
- Migration system
- Repository pattern implementation
- Connection pooling and resource limits

### Video 2.3: Service Layer Design (15 min)
- Service interfaces and implementations
- Constructor injection with `NewService` pattern
- Error handling and wrapping
- Context propagation

### Video 2.4: Handler Layer with Gin (10 min)
- Gin router setup
- Handler struct pattern
- Request validation
- Response formatting

---

## Module 3: Authentication and Authorization (45 min)

### Video 3.1: JWT Authentication (20 min)
- Token generation and validation
- Session management
- Refresh token rotation
- Security best practices

### Video 3.2: Role-Based Access Control (15 min)
- Permission system design
- Role definitions
- Middleware implementation
- Resource ownership checks

### Video 3.3: Multi-Device Sessions (10 min)
- Active session tracking
- Session deactivation
- Logout all devices

---

## Module 4: Media Detection and Processing (75 min)

### Video 4.1: Universal Scanner Design (20 min)
- Protocol abstraction interface
- Concurrent scanning strategies
- Progress reporting via WebSocket
- Error recovery and retry logic

### Video 4.2: Media Detection Pipeline (25 min)
- File type detection
- Metadata extraction
- Content analysis
- Duplicate detection algorithms

### Video 4.3: Entity Aggregation (20 min)
- Media entity model (11 types)
- Hierarchy building (TV shows, music)
- Title parsing with regex
- External metadata integration (TMDB, IMDB)

### Video 4.4: Thumbnail and Preview Generation (10 min)
- Image processing
- Video frame extraction
- Caching strategies

---

## Module 5: Frontend Development (60 min)

### Video 5.1: React + TypeScript Setup (15 min)
- Vite configuration
- Path aliases
- Environment variables
- API proxy setup

### Video 5.2: State Management (15 min)
- React Query for server state
- Zustand for client state
- Context providers pattern
- WebSocket integration

### Video 5.3: Component Architecture (20 min)
- Atomic design principles
- Shared UI components (`@vasic-digital/ui-components`)
- Form handling with React Hook Form + Zod
- Animation with Framer Motion

### Video 5.4: Testing Frontend (10 min)
- Vitest unit tests
- Playwright E2E tests
- MSW for API mocking

---

## Module 6: Real-Time Features (30 min)

### Video 6.1: WebSocket Server (15 min)
- Connection management
- Message protocol design
- Broadcast patterns
- Reconnection handling

### Video 6.2: Live Updates (15 min)
- Scan progress streaming
- Real-time notifications
- Presence indicators
- Event bus integration

---

## Module 7: Protocol Implementations (60 min)

### Video 7.1: SMB/CIFS Client (15 min)
- Connection handling
- Authentication
- File operations
- Offline caching

### Video 7.2: WebDAV Client (15 min)
- HTTP methods mapping
- PROPFIND parsing
- Lock management
- Batch operations

### Video 7.3: FTP Client (10 min)
- Active vs passive mode
- Binary vs ASCII transfer
- Directory listing

### Video 7.4: NFS and Local Filesystems (10 min)
- Mount point detection
- Permission handling
- Symbolic link resolution

### Video 7.5: Protocol Factory Pattern (10 min)
- Unified client interface
- Protocol auto-detection
- Configuration-driven creation

---

## Module 8: HTTP/3 and Performance (45 min)

### Video 8.1: HTTP/3 with QUIC (20 min)
- quic-go integration
- TLS certificate handling
- ALPN negotiation
- Fallback to HTTP/2

### Video 8.2: Brotli Compression (10 min)
- Content negotiation
- Compression levels
- Static asset optimization

### Video 8.3: Caching Strategies (15 min)
- Redis integration
- Cache invalidation
- Stale-while-revalidate
- Memory-mapped files

---

## Module 9: Desktop and Mobile Apps (60 min)

### Video 9.1: Tauri Desktop App (25 min)
- Rust backend integration
- IPC commands and events
- Native file dialogs
- Auto-updates

### Video 9.2: Android Development (20 min)
- Kotlin + Jetpack Compose
- MVVM architecture
- Room database
- Retrofit API client

### Video 9.3: Android TV (15 min)
- Leanback UI framework
- Remote control navigation
- Media playback
- Live channels integration

---

## Module 10: Testing and Quality (45 min)

### Video 10.1: Unit Testing Patterns (15 min)
- Table-driven tests
- Mock interfaces vs sqlmock
- Test helpers and fixtures
- Coverage measurement

### Video 10.2: Integration Testing (15 min)
- Container-based tests
- Database fixtures
- API contract testing
- End-to-end flows

### Video 10.3: Challenge System (15 min)
- Challenge framework architecture
- Writing custom challenges
- User flow automation
- CI/CD integration

---

## Module 11: Security and Monitoring (30 min)

### Video 11.1: Security Best Practices (15 min)
- gosec and staticcheck
- SQL injection prevention
- XSS protection
- CSRF tokens

### Video 11.2: Observability (15 min)
- Prometheus metrics
- Structured logging with Zap
- Health check endpoints
- Error reporting service

---

## Module 12: Deployment and DevOps (45 min)

### Video 12.1: Container Strategy (20 min)
- Multi-stage builds
- Podman compose configuration
- Resource limits
- Network isolation

### Video 12.2: Installation Wizard (15 min)
- Guided setup flow
- Environment detection
- Configuration generation
- First-time user experience

### Video 12.3: Production Considerations (10 min)
- PostgreSQL migration
- Horizontal scaling
- Backup strategies
- Monitoring alerts

---

## Bonus Content

### B1: Troubleshooting Guide (20 min)
- Common issues and solutions
- Debug logging techniques
- Performance profiling
- Memory leak detection

### B2: Contributing Guide (15 min)
- Code style conventions
- PR requirements
- Commit message format
- Documentation standards

### B3: API Reference Deep Dive (30 min)
- OpenAPI/Swagger documentation
- Authentication flows
- Rate limiting
- Versioning strategy

---

## Course Materials

- **Source Code**: Complete repository with tagged versions per module
- **Slides**: PDF and web-based presentations
- **Exercises**: Hands-on coding challenges
- **Cheat Sheets**: Quick reference guides
- **Sample Media**: Test files for scanning exercises

## Certification

- Module quizzes (passing score: 80%)
- Final project: Build a custom media scanner plugin
- Certificate of completion
