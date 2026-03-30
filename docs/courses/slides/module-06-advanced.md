# Module 6: Developer Guide - Slide Deck Outline

**Total Slides**: 14
**Estimated Duration**: 70 minutes

---

## Slide 1: Title Slide (2 min)

**Title**: Developer Guide and API

- Architecture deep dive, adding features, submodule system, password change, media quality
- Prerequisites: familiarity with Go and TypeScript
- By the end: trace request flow, add features, and use advanced API endpoints

---

## Slide 2: Architecture Deep Dive (5 min)

**Title**: Tracing a Request Through the Backend

- Handler -> Service -> Repository -> Database (SQLite or PostgreSQL)
- Dual package layout: top-level (domain) and internal/ (infrastructure)
- Middleware chain: auth verification, rate limiting, metrics, Brotli compression
- HTTP/3 (QUIC) with self-signed TLS certs generated at startup
- Demo: trace /api/v1/entities from handler to database query

---

## Slide 3: Media Detection Pipeline (5 min)

**Title**: From File to Structured Entity

- UniversalScanner crawls storage sources
- detector/ identifies media type from filename and path patterns
- analyzer/ extracts technical metadata (resolution, codec, bitrate)
- providers/ fetch external metadata (TMDB, OMDB, MusicBrainz, OpenLibrary)
- AggregationService creates MediaItems with parent-child hierarchy

---

## Slide 4: Frontend Architecture (5 min)

**Title**: React Application Structure

- AuthProvider -> WebSocketProvider -> Router -> ProtectedRoute -> Pages
- React Query for server state, Zustand for client state
- Tailwind CSS, React Hook Form + Zod, framer-motion
- Path aliases: @/components, @/hooks, @/lib, @/types, @/services
- API proxy: reads .service-port for backend discovery
- Exercise reference: Exercise 6.1 -- add a new page with protected routing

---

## Slide 5: Development Environment Setup (5 min)

**Title**: Setting Up for Development

- Clone, initialize submodules, run go mod tidy
- Backend: go run main.go (writes .service-port)
- Frontend: npm install && npm run dev (port 3000, proxies /api)
- Container dev: podman-compose -f docker-compose.dev.yml up
- Kill port 3000 conflicts: ss -tlnp | grep :3000

---

## Slide 6: Adding a New API Endpoint (5 min)

**Title**: Handler, Service, Route Pattern

- Create handler function in handlers/ with Gin context
- Create service method in services/ with business logic
- Create repository method in repository/ for data access
- Register route in main.go under /api/v1
- Exercise reference: Exercise 6.2 -- add a custom endpoint

---

## Slide 7: Adding a New Storage Protocol (5 min)

**Title**: Implementing the UnifiedClient Interface

- filesystem/interface.go defines the UnifiedClient contract
- filesystem/factory.go creates protocol-specific clients
- Implement: Connect, Disconnect, List, Read, Write, Delete, Stat
- Register the new protocol in the factory
- Demo: walk through the SMB client implementation as a reference

---

## Slide 8: Submodule Architecture (5 min)

**Title**: Reusable Go Modules via Replace Directives

- 29 independent submodules under vasic-digital organization
- catalog-api/go.mod uses replace directives to map local paths
- Each module: own repo, tests, ARCHITECTURE.md, Upstreams
- scripts/setup-submodule.sh to scaffold new modules
- Commit and push: cd SubmoduleName && commit "message"

---

## Slide 9: Password Change API (4 min)

**Title**: Updating User Credentials

- POST /api/v1/auth/change-password with current and new password
- Password validation: minimum length, complexity rules
- Error responses: 400 (validation), 401 (wrong current password)
- Tokens remain valid after password change
- Exercise reference: Exercise 6.3 -- change password via API and UI

---

## Slide 10: Media Quality Analysis (5 min)

**Title**: Inspecting Resolution, Codec, and Bitrate

- Quality analysis endpoint returns technical metadata per file
- Resolution, codec, bitrate, frame rate, audio channels
- Trigger analysis and metadata refresh for individual items or entire roots
- Quality distribution statistics across the library
- Demo: query quality data for a scanned movie file

---

## Slide 11: Popular and Recent Media (4 min)

**Title**: Browsing by Popularity and Recency

- Popular media endpoint: sorted by favorite count
- Recent media endpoint: sorted by addition date
- Both support pagination and type filtering
- Search by file path for direct file lookup
- Exercise reference: Exercise 6.4 -- fetch popular media via API client

---

## Slide 12: Build Pipeline (5 min)

**Title**: Building All Components

- scripts/release-build.sh --container --force --skip-tests
- Build/ submodule: versioning, SHA256 change detection, orchestration
- Per-component builders in scripts/lib/build-*.sh
- Version injection via -ldflags at build time
- All 7 components build in ~17 minutes containerized

---

## Slide 13: Database Dialect Abstraction (4 min)

**Title**: Writing SQL That Works on Both Databases

- database/dialect.go: DialectSQLite and DialectPostgres
- Auto-rewrite: ? -> $1,$2 for PostgreSQL placeholders
- INSERT OR IGNORE -> ON CONFLICT DO NOTHING
- Boolean literals: = 0/1 -> = FALSE/TRUE
- InsertReturningID() replaces LastInsertId() for PostgreSQL

---

## Slide 14: Module Summary and Next Steps (3 min)

**Title**: What We Covered

- Request flow through handler, service, repository layers
- Media detection pipeline and aggregation
- Adding endpoints, protocols, and submodules
- Password change, media quality, popular/recent APIs
- Build pipeline and database dialect abstraction
- Advanced modules next: Testing, Deployment, Architecture, Monitoring
