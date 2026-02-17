# AGENTS.md - Catalogizer Development Guide

This guide helps AI agents work effectively in the Catalogizer codebase by documenting essential commands, patterns, and conventions.

## Project Overview

Catalogizer is a comprehensive, multi-platform media collection management system that automatically detects, categorizes, and organizes media files across multiple storage protocols (SMB, FTP, NFS, WebDAV, local filesystem). It follows a modern distributed architecture with clear separation of concerns across multiple components.

## Main Components

- **catalog-api**: Go backend REST API server for media cataloging
- **catalog-web**: React frontend with TypeScript and Tailwind CSS
- **catalogizer-desktop**: Tauri desktop application (Rust + React)
- **catalogizer-android**: Native Android app (Kotlin + Jetpack Compose)
- **catalogizer-androidtv**: Android TV app with Leanback UI
- **catalogizer-api-client**: TypeScript API client library
- **installer-wizard**: Tauri-based installation wizard for SMB configuration

## Essential Commands

### Backend (catalog-api)

```bash
cd catalog-api

# Install dependencies
go mod tidy

# Run development server
go run main.go

# Run with test mode
go run main.go -test-mode

# Build release binary
go build -o catalog-api

# Run tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific test file
go test -v ./handlers

# Run comprehensive tests
./scripts/run_comprehensive_tests.sh
```

### Frontend (catalog-web)

```bash
cd catalog-web

# Install dependencies
npm install

# Development server (hot reload)
npm run dev  # Runs on http://localhost:5173

# Build for production
npm run build

# Preview production build
npm run preview

# Run tests
npm run test

# Run tests with coverage
npm run test:coverage

# Watch mode for development
npm run test:watch

# Linting
npm run lint
npm run lint:fix

# Format code
npm run format

# Type checking
npm run type-check
```

### Desktop Application (catalogizer-desktop)

```bash
cd catalogizer-desktop

# Install dependencies
npm install

# Development (hot reload)
npm run tauri:dev

# Build for all platforms
npm run tauri:build

# Build for specific platform
npm run tauri:build -- --target x86_64-pc-windows-gnu  # Windows
npm run tauri:build -- --target x86_64-apple-darwin    # macOS
npm run tauri:build -- --target x86_64-unknown-linux-gnu # Linux
```

### Installer Wizard (installer-wizard)

```bash
cd installer-wizard

# Install dependencies
npm install

# Development server
npm run tauri:dev

# Build release
npm run tauri:build

# Run tests
npm run test

# Test with UI
npm run test:ui

# Coverage report
npm run test:coverage

# Health check (tests + build)
npm run health:check
```

### Android Apps (catalogizer-android & catalogizer-androidtv)

```bash
cd catalogizer-android  # or catalogizer-androidtv

# Build debug APK
./gradlew assembleDebug

# Build release APK
./gradlew assembleRelease

# Run tests
./gradlew test

# Run instrumented tests
./gradlew connectedAndroidTest

# Run linting
./gradlew lintKotlin

# Build and run on emulator
./gradlew installDebug
```

### API Client Library (catalogizer-api-client)

```bash
cd catalogizer-api-client

# Install dependencies
npm install

# Build TypeScript to JavaScript
npm run build

# Development (watch mode)
npm run dev

# Run tests
npm run test

# Lint code
npm run lint

# Publish to npm (after build)
npm publish
```

### Docker Operations

```bash
# Development setup
docker-compose -f docker-compose.dev.yml up

# Start with monitoring tools
docker-compose -f docker-compose.dev.yml --profile tools up

# Production setup
docker-compose up -d

# View logs
docker-compose logs -f api
docker-compose logs -f postgres
docker-compose logs -f redis

# Rebuild containers
docker-compose build --no-cache

# Stop all services
docker-compose down
```

### Build Scripts

```bash
# Build backend for multiple platforms
cd catalog-api && ./scripts/build.sh

# Build all client applications
./build-scripts/build-all.sh

# Run comprehensive protocol tests
cd catalog-api && ./scripts/run_comprehensive_tests.sh
```

### Full Test Suite with Security Scanning

```bash
# Run all tests including security
./scripts/run-all-tests.sh

# Run security tests only
./scripts/security-test.sh

# SonarQube scan (requires token)
SONAR_TOKEN=xxx ./scripts/sonarqube-scan.sh

# Snyk scan (requires token)
SNYK_TOKEN=xxx ./scripts/snyk-scan.sh
```

## Architecture Patterns

### Multi-Protocol Abstraction Layer (catalog-api/filesystem)

The codebase uses a Factory + Strategy pattern for unified file system operations across protocols. The key interface is in `catalog-api/filesystem/interface.go`:

```go
type FileSystemClient interface {
    Connect(ctx context.Context) error
    Disconnect(ctx context.Context) error
    ListDirectory(ctx context.Context, path string) ([]*FileInfo, error)
    // ... other methods
}
```

Protocol implementations are in:
- `smb_client.go` - SMB/CIFS implementation
- `ftp_client.go` - FTP/FTPS implementation
- `nfs_client.go` - NFS implementation
- `webdav_client.go` - WebDAV implementation
- `local_client.go` - Local filesystem

### Service Layer Architecture

```
HTTP Handler (handlers/)
       ↓
   Service Layer (services/, internal/services/)
       ↓
 Repository Layer (repository/)
       ↓
  Database Layer (database/)
```

### SMB Resilience Layer

Located in `catalog-api/internal/smb/`, implements:
- Circuit breaker pattern
- Cache-aside strategy
- Retry with exponential backoff
- Health checker for continuous monitoring

### Context-Based State Management (catalog-web)

React Contexts used:
- `AuthContext` - User authentication and permissions
- `WebSocketContext` - Real-time WebSocket connections

### MVVM Architecture (Android/AndroidTV)

```
UI Layer (Compose)
       ↓
ViewModel Layer
       ↓
Data Layer (Repository)
```

## Code Organization & Conventions

### Go Backend

- PascalCase for exported (public), camelCase for unexported (private)
- Interface names: `Reader`, `Writer`, `Service`
- Receiver names: short, often single letter
- Constructor injection (NewService pattern)
- `*_test.go` files alongside source for tests

### TypeScript/React

- PascalCase for classes, interfaces, components
- camelCase for functions, variables
- SCREAMING_SNAKE_CASE for constants
- Type/Interface suffixes: `IService`, `Props`, `State`
- Component structure: `components/`, `pages/`, `contexts/`, `hooks/`, `services/`, `types/`, `utils/`

### Kotlin/Android

- PascalCase for classes, interfaces
- camelCase for functions, variables
- ViewModel suffix: `MediaViewModel`
- MVVM with Hilt dependency injection
- Material Design 3 components

## Testing Approach

### Go

- Unit tests for each package
- Mock interfaces for dependencies
- Table-driven tests for multiple scenarios
- Coverage threshold: 80%

### TypeScript/React

- Jest for unit testing
- React Testing Library for component tests
- Mock API responses with MSW
- Coverage threshold: 80%

### Kotlin/Android

- JUnit for unit tests
- Espresso for UI tests
- Mockito for mocking

## Constraints

**All builds and services MUST use containers.** Never build or run services directly on the host machine. Always use the containerized approach: `podman run --network host` for single-container builds, `podman-compose` for multi-service environments. Nothing — builds, tests, service execution — should be executed directly on the host. The builder container has all required toolchains (Go, Node, Rust, JDK, Android SDK). Use Podman as the container runtime (Docker is not available).

**GitHub Actions are PERMANENTLY DISABLED.** Do NOT create any GitHub Actions workflow files. CI/CD, security scanning, and automated builds must be run locally in containers.

## Important Gotchas

### Protocol Implementation

- When adding new file protocols, implement the `FileSystemClient` interface
- Add protocol to the factory in `filesystem/factory.go`
- Write unit tests following the pattern in `filesystem/*_test.go`

### SMB Connections

- The SMB resilience layer handles connection failures automatically
- Connection status is broadcast via WebSocket to connected clients
- Offline cache provides metadata during network outages

### Configuration System

Configuration hierarchy (highest to lowest priority):
1. Environment Variables
2. .env File
3. config.json File
4. Code Defaults

#### Backend Configuration (catalog-api/.env)
```env
PORT=8080
HOST=0.0.0.0
GIN_MODE=debug

DB_PATH=./data/catalogizer.db
LOG_LEVEL=debug

SMB_SOURCES=smb://server/share
SMB_USERNAME=user
SMB_PASSWORD=pass
```

#### Frontend Configuration (catalog-web/.env.local)
```env
VITE_API_BASE_URL=http://localhost:8080
VITE_WS_URL=ws://localhost:8080/ws
VITE_ENABLE_ANALYTICS=true
VITE_ENABLE_REALTIME=true
```

#### Docker Configuration (.env)
```env
POSTGRES_USER=catalogizer
POSTGRES_PASSWORD=dev_password
POSTGRES_DB=catalogizer_dev

JWT_SECRET=your-jwt-secret-here
```

### Database Migrations

- Database migrations are in `catalog-api/database/migrations/`
- Follow the naming convention: `000001_description.up.sql` and `000001_description.down.sql`
- Use `catalog-api/database/migrations/README.md` for migration guidelines
- Apply migrations with: `go run main.go -migrate`

### WebSocket Integration

- Real-time updates are broadcast when file system changes occur
- Frontend uses `WebSocketContext` to maintain connection
- Messages trigger React Query invalidations for automatic UI updates

### Development Environment

- Use `docker-compose.dev.yml` for local development with hot reload
- API runs on http://localhost:8080
- Web app runs on http://localhost:5173
- pgAdmin available on http://localhost:5050
- Redis Commander available on http://localhost:8081

### Security Requirements

- All tests must pass with 100% success before deployment
- Security scans (SonarQube, Snyk, Trivy) are mandatory
- JWT tokens used for authentication
- Input validation and parameterized queries required

### Tauri Applications

- Frontend → Backend communication via Tauri Commands
- Backend → Frontend communication via Tauri Events
- Type-safe communication with serde serialization

#### Tauri Configuration Notes
- Desktop app runs on http://localhost:1420 in development
- File system access requires explicit scope configuration
- Shell access is limited to specific commands for security
- HTTP requests allowed to all domains (configurable in allowlist)

## Key Entry Points

### Backend

- `catalog-api/main.go:43-172` - API server bootstrap
- `catalog-api/filesystem/interface.go` - Core abstraction interface
- `catalog-api/filesystem/factory.go` - Protocol factory
- `catalog-api/internal/services/catalog_service.go` - Core catalog service

### Frontend

- `catalog-web/src/App.tsx:14-111` - Root component with routing
- `catalog-web/src/contexts/AuthContext.tsx` - Authentication state
- `catalog-web/src/contexts/WebSocketContext.tsx` - WebSocket management

### Mobile

- `catalogizer-android/app/src/main/java/com/catalogizer/android/CatalogizerApplication.kt` - Android app entry
- `catalogizer-androidtv/app/src/main/java/com/catalogizer/androidtv/CatalogizerTVApplication.kt` - TV app entry

## Project Dependencies

### Backend (Go)
- Gin Framework 1.9.1 (HTTP routing)
- JWT 5.3.0 (Authentication)
- Zap 1.26.0 (Logging)
- SQLite3 1.14.18 (Database)
- Protocol clients for SMB, FTP, NFS, WebDAV

### Frontend (TypeScript/React)
- React 18.2.0 (UI framework)
- TanStack React Query 4.24.6 (Server state)
- Tailwind CSS 3.2.7 (Styling)
- Vite 4.1.0 (Build tool)
- Axios 1.6.0 (HTTP client)
- React Router DOM 6.8.0 (Routing)
- Zustand 4.3.6 (State management)
- Framer Motion 10.0.1 (Animations)

### Mobile (Kotlin)
- Jetpack Compose (UI)
- Room (Database)
- Retrofit + OkHttp (Networking)
- Hilt (Dependency injection)
- Android SDK 34 (Target)
- Min SDK 26 (Android 8.0)

This documentation should help AI agents understand the codebase structure and work effectively with all components of the Catalogizer project.

## Quick Development Setup

For new agents setting up the development environment:

1. Clone the repository and navigate to the project root
2. Run `./scripts/install.sh --mode=development` to set up all services
3. This will configure:
   - PostgreSQL with test data
   - Redis cache
   - catalog-api with hot reload
   - catalog-web with hot reload
   - pgAdmin on port 5050
   - Redis Commander on port 8081

Access points:
- API: http://localhost:8080
- Web: http://localhost:5173
- pgAdmin: http://localhost:5050
- Redis Commander: http://localhost:8081

For individual component development, see the specific commands in each section above.