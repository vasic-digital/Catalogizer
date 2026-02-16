# Module 1: Installation & Setup - Slide Outlines

---

## Slide 1.1.1: Title Slide

**Title**: What is Catalogizer?

**Subtitle**: Multi-Protocol Media Collection Management

**Visual**: Catalogizer logo centered, subtitle below

**Speaker Notes**: Welcome the audience. Introduce yourself and the course structure. This first module covers what Catalogizer is, why it exists, and how to install it. By the end of the module, students will have a running instance.

---

## Slide 1.1.2: The Problem

**Title**: Media Is Everywhere

**Bullet Points**:
- Files scattered across SMB shares, FTP servers, NFS mounts, WebDAV endpoints, and local disks
- No single view of your entire collection
- Manual organization is tedious and error-prone
- Metadata is missing or inconsistent across sources
- No cross-device access to a unified catalog

**Speaker Notes**: Paint the picture of a typical media hoarder's setup. Multiple NAS devices, a few external drives, maybe some cloud storage. Everything is siloed. Finding a specific movie means checking three different places.

---

## Slide 1.1.3: The Solution

**Title**: Catalogizer Unifies Your Media

**Bullet Points**:
- Connects to 5 storage protocols: SMB/CIFS, FTP/FTPS, NFS, WebDAV, Local
- Detects 50+ media types automatically
- Enriches with metadata from TMDB, IMDB, TVDB, MusicBrainz, Spotify, Steam
- Real-time monitoring via WebSocket
- Access from web, desktop, Android phone, and Android TV

**Visual**: Diagram showing multiple storage sources converging into a single Catalogizer catalog

**Speaker Notes**: Explain that Catalogizer is the single pane of glass for all your media. Emphasize the automatic detection -- you do not manually tag anything.

---

## Slide 1.1.4: System Components

**Title**: Architecture at a Glance

**Bullet Points**:
- **catalog-api** (Go/Gin) -- Backend REST API and media detection engine
- **catalog-web** (React/TypeScript) -- Modern web frontend with real-time updates
- **catalogizer-desktop** (Tauri: Rust + React) -- Native desktop application
- **installer-wizard** (Tauri) -- Guided first-time setup
- **catalogizer-android** (Kotlin/Compose) -- Android mobile app
- **catalogizer-androidtv** (Kotlin/Compose) -- Android TV app
- **catalogizer-api-client** (TypeScript) -- API client library

**Visual**: Component diagram with arrows showing data flow between components

**Speaker Notes**: Emphasize that catalog-api and catalog-web are the core. Everything else extends the experience. You do not need all components to get started.

---

## Slide 1.1.5: Submodule Architecture

**Title**: Reusable Building Blocks

**Bullet Points**:
- Independent git submodules under the vasic-digital organization
- Go modules: Auth, Cache, Database, Concurrency, Storage, EventBus, Streaming, Security, Observability, Filesystem, RateLimiter, Config, Discovery, Media, Middleware, Watcher
- TypeScript modules: WebSocket-Client, UI-Components-React
- Kotlin module: Android-Toolkit
- Each has its own repo, tests, and documentation

**Speaker Notes**: The submodule architecture means common functionality is shared across projects. If you fix a bug in the Auth module, every project using it benefits.

---

## Slide 1.2.1: System Requirements

**Title**: What You Need

**Bullet Points**:
- **Backend**: Go 1.21+
- **Frontend**: Node.js 18+ with npm
- **Desktop**: Rust toolchain (rustup.rs)
- **Mobile**: Android Studio with SDK 26+, JDK 17+
- **Containers**: Podman 5+ or Docker 20.10+
- **Database**: SQLite (development) or PostgreSQL 15+ (production)
- **Browser**: Chrome 90+, Firefox 88+, Safari 14+, Edge 90+

**Speaker Notes**: Not everything is needed at once. For evaluation, Go and Node.js are sufficient to run the backend and frontend.

---

## Slide 1.2.2: Backend Architecture Layers

**Title**: Request Flow Through the Backend

**Visual**: Flowchart: Request -> Gin Router -> Auth Middleware -> Handler -> Service -> Repository -> SQLite/PostgreSQL

**Bullet Points**:
- Handlers parse requests and format responses (thin layer)
- Services contain business logic (core layer)
- Repositories manage database access (data layer)
- Two handler/service layers: top-level and internal/

**Speaker Notes**: Walk through a concrete example. A user searches for a movie. The request hits the search handler, which calls the catalog service, which queries the repository, which reads from the database.

---

## Slide 1.2.3: Media Detection Pipeline

**Title**: How Media Is Identified

**Visual**: Pipeline diagram: File Discovery -> Detector -> Analyzer -> Providers -> Catalog Entry

**Bullet Points**:
- **Universal Scanner**: Crawls connected storage sources
- **Detector**: Identifies media type from filename, path, extension
- **Analyzer**: Extracts technical metadata (resolution, codec, bitrate)
- **Providers**: Fetch external metadata (TMDB, IMDB, TVDB, MusicBrainz, Spotify, Steam)
- **Recognition**: Specialized providers for movies, music, books, games/software

**Speaker Notes**: This is the magic of Catalogizer. A file named "Inception.2010.1080p.BluRay.mkv" is identified as a movie, its resolution detected as 1080p, and TMDB provides the poster, synopsis, cast, and ratings.

---

## Slide 1.2.4: Real-Time Event System

**Title**: Live Updates Without Refreshing

**Visual**: Diagram: File Change -> Event Bus -> WebSocket Server -> Connected Clients (Browser, Desktop, Mobile)

**Bullet Points**:
- Event bus captures all changes: new files, deletions, metadata updates
- WebSocket server pushes events to all connected clients
- Frontend contexts (AuthContext, WebSocketContext) distribute events
- No polling -- instant updates across all connected devices

**Speaker Notes**: When someone drops a new file onto an SMB share, every connected browser, desktop app, and mobile app sees it appear within seconds. This is powered by the event bus and WebSocket infrastructure.

---

## Slide 1.3.1: Container Installation

**Title**: Getting Started with Containers

**Bullet Points**:
- Fastest path to a running instance
- Supports both Podman (preferred) and Docker
- Three compose configurations:
  - `docker-compose.dev.yml` -- Development with hot reloading
  - `docker-compose.yml` -- Production with resource limits
  - `docker-compose.build.yml` -- Reproducible cross-platform builds
- Services: PostgreSQL 15, Redis 7, Nginx, catalog-api, catalog-web

**Speaker Notes**: Containers handle all dependencies. No need to install Go, Node.js, or configure a database manually. Just compose up and go.

---

## Slide 1.3.2: Container Configuration

**Title**: Essential Environment Variables

**Bullet Points**:
- `POSTGRES_PASSWORD` -- Required; compose refuses to start without it
- `JWT_SECRET` -- Signs authentication tokens; use 32+ random characters
- `DB_ENCRYPTION_KEY` -- Encrypts the database at rest; exactly 32 characters
- `TMDB_API_KEY` -- Free from themoviedb.org; enables movie metadata
- `GIN_MODE` -- Set to "release" for production

**Visual**: Example .env file with annotated variables

**Speaker Notes**: Walk through creating the .env file. Stress that JWT_SECRET and DB_ENCRYPTION_KEY must be kept secure and never committed to version control.

---

## Slide 1.3.3: Development vs Production

**Title**: Choosing the Right Configuration

| Aspect | Development | Production |
|--------|-------------|------------|
| Compose file | docker-compose.dev.yml | docker-compose.yml |
| Hot reloading | Yes | No |
| Database | SQLite or PostgreSQL | PostgreSQL with health checks |
| Resource limits | None | CPU and memory caps |
| Debug logging | Enabled | Disabled |
| Security | Relaxed | Full enforcement |

**Speaker Notes**: Always use the dev configuration for learning and testing. Production configuration adds health checks, resource limits, and tighter security.

---

## Slide 1.3.4: Podman-Specific Notes

**Title**: Using Podman Instead of Docker

**Bullet Points**:
- Use fully qualified image names: `docker.io/library/postgres:15-alpine`
- Use `podman build --network host` to avoid SSL issues
- Set `GOTOOLCHAIN=local` to prevent Go auto-downloading toolchains
- All compose commands are identical: `podman-compose` replaces `docker-compose`

**Speaker Notes**: Podman is a rootless container runtime. It provides better security isolation than Docker but has some quirks with image name resolution and networking that we need to work around.

---

## Slide 1.4.1: Manual Installation -- Backend

**Title**: Setting Up catalog-api

**Bullet Points**:
- Clone the repository and initialize submodules
- `cd catalog-api && go mod tidy`
- Create `.env` with PORT, JWT_SECRET, DB_TYPE, API keys
- `go run main.go` -- Server starts on port 8080
- SQLite is the default database; file created automatically
- Verify: `curl http://localhost:8080/api/v1/health`

**Visual**: Terminal screenshot showing server startup output

**Speaker Notes**: Walk through each step live. Explain that go mod tidy downloads all dependencies. The .env file is the primary configuration mechanism.

---

## Slide 1.4.2: Manual Installation -- Frontend

**Title**: Setting Up catalog-web

**Bullet Points**:
- `cd catalog-web && npm install`
- Create `.env.local` with VITE_API_BASE_URL and VITE_WS_URL
- Feature flags: VITE_ENABLE_ANALYTICS, VITE_ENABLE_REALTIME, VITE_ENABLE_EXTERNAL_METADATA, VITE_ENABLE_OFFLINE_MODE
- `npm run dev` -- Frontend starts on port 5173 with HMR
- Production build: `npm run build`
- Quality checks: `npm run lint && npm run type-check`

**Visual**: Browser screenshot of the login page at localhost:5173

**Speaker Notes**: Enable all feature flags during development so you can explore the full functionality. The frontend connects to the backend automatically via the configured API URL.

---

## Slide 1.4.3: SMB Resilience Configuration

**Title**: Network Storage Resilience

**Bullet Points**:
- `SMB_RETRY_ATTEMPTS` (default 5) -- Maximum reconnection attempts
- `SMB_RETRY_DELAY_SECONDS` (default 30) -- Initial retry delay with exponential backoff
- `SMB_HEALTH_CHECK_INTERVAL` (default 60) -- Health check frequency in seconds
- `SMB_CONNECTION_TIMEOUT` (default 30) -- Connection timeout in seconds
- `SMB_OFFLINE_CACHE_SIZE` (default 1000) -- Number of items cached for offline access

**Speaker Notes**: These settings control how Catalogizer handles network interruptions. The circuit breaker pattern prevents hammering a downed server. Offline cache ensures users still see data during outages.

---

## Slide 1.4.4: Verification Checklist

**Title**: Confirming Your Installation

**Bullet Points**:
- [ ] Backend responds at http://localhost:8080
- [ ] Frontend loads at http://localhost:5173
- [ ] Login works with initial credentials
- [ ] Dashboard displays Quick Stats panel
- [ ] WebSocket connection established (check browser dev tools)
- [ ] At least one storage source can be connected

**Speaker Notes**: Run through this checklist to confirm everything is working. If any step fails, refer to the Troubleshooting Guide. The most common issue is incorrect .env configuration.

---

## Slide 1.4.5: Module 1 Summary

**Title**: What We Covered

**Bullet Points**:
- Catalogizer unifies media across 5 storage protocols into one catalog
- 7 components span web, desktop, mobile, and API access
- Submodule architecture provides reusable building blocks
- Container installation is the fastest path (Podman or Docker)
- Manual installation gives full control over each component
- SQLite for development; PostgreSQL for production

**Next Steps**: Module 2 -- Getting Started with Media Management

**Speaker Notes**: Recap the key points. Ensure everyone has a running installation before moving to Module 2. Offer to help troubleshoot during the break.
