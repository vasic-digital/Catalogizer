# Module 1: Introduction and Installation - Slide Deck Outline

**Total Slides**: 12
**Estimated Duration**: 45 minutes

---

## Slide 1: Title Slide (2 min)

**Title**: Introduction and Installation

- Course overview: Mastering Catalogizer
- Target audience: end users, administrators, developers
- Prerequisites: basic familiarity with web applications
- What you will build by the end of this module

---

## Slide 2: What Is Catalogizer? (3 min)

**Title**: The Multi-Protocol Media Collection Manager

- Detects, categorizes, and organizes media across storage protocols
- Supported protocols: SMB, FTP, NFS, WebDAV, local filesystem
- Automatic metadata enrichment from TMDB, IMDB, MusicBrainz, OpenLibrary
- Real-time monitoring via WebSocket push events
- Demo: show a running instance with populated catalog

---

## Slide 3: System Components (4 min)

**Title**: Architecture at a Glance

- catalog-api (Go 1.24 / Gin) -- REST backend and detection engine
- catalog-web (React 18 / TypeScript / Vite) -- web frontend
- catalogizer-desktop and installer-wizard (Tauri: Rust + React)
- catalogizer-android and catalogizer-androidtv (Kotlin / Compose)
- catalogizer-api-client (TypeScript library)
- 29 independent git submodules for reusable components

---

## Slide 4: Request Flow (4 min)

**Title**: How a Request Travels Through the Backend

- Handler (parse request) -> Service (business logic) -> Repository (data access) -> Database
- Dual package layout: top-level for domain, internal/ for infrastructure
- Middleware chain: auth, rate limiting, metrics, compression
- Exercise reference: trace a /api/v1/health request through the code

---

## Slide 5: System Requirements (3 min)

**Title**: What You Need Before Installing

- Go 1.24+, Node.js 18+, Git
- Optional: Rust toolchain (desktop), Android SDK 34 (mobile), JDK 17
- Podman 5+ (container runtime, preferred over Docker)
- SQLite for development, PostgreSQL 15+ for production
- Minimum 4 GB RAM, 2 CPU cores for local development

---

## Slide 6: Container Installation (5 min)

**Title**: Fastest Path -- Containers with Podman

- Clone repo, initialize submodules
- podman-compose -f docker-compose.dev.yml up
- Services started: PostgreSQL, Redis, Nginx, catalog-api, catalog-web
- Critical: use fully qualified image names for Podman
- Critical: --network host for builds to avoid SSL issues
- Demo: run docker-compose.dev.yml and open the dashboard

---

## Slide 7: Manual Backend Setup (5 min)

**Title**: Setting Up catalog-api

- cd catalog-api && go mod tidy
- Create .env with PORT, JWT_SECRET, DB_TYPE, API keys
- go run main.go -- writes .service-port file for frontend discovery
- Dynamic port binding with HTTP/3 (QUIC) support
- Verify: curl http://localhost:8080/api/v1/health
- Exercise reference: Exercise 1.1 from EXERCISES.md

---

## Slide 8: Manual Frontend Setup (4 min)

**Title**: Setting Up catalog-web

- cd catalog-web && npm install && npm run dev
- Vite reads .service-port to auto-proxy /api to the backend
- Frontend runs on port 3000 with hot module replacement
- Path aliases: @/components, @/hooks, @/lib, @/types, @/services
- Verify: open http://localhost:3000 and log in

---

## Slide 9: Environment Configuration (4 min)

**Title**: Essential Environment Variables

- PORT, GIN_MODE, DB_TYPE, JWT_SECRET, ADMIN_PASSWORD
- TMDB_API_KEY and OMDB_API_KEY for metadata enrichment (optional)
- Config precedence: env vars > .env > config.json > defaults
- Never commit .env files with real secrets
- Demo: create .env file and start services

---

## Slide 10: Database Options (3 min)

**Title**: SQLite vs PostgreSQL

- SQLite: zero setup, single file, WAL mode enabled
- PostgreSQL: set DB_TYPE=postgres with host, port, name, user, password
- Dual-dialect abstraction auto-rewrites SQL for each database
- Migration system handles schema upgrades automatically
- Container default: PostgreSQL with port 5432 mapped to 5433

---

## Slide 11: Verification Checklist (4 min)

**Title**: Confirming Your Installation

- Backend responds at the configured port
- Frontend loads and displays the login page
- Login works with initial admin credentials
- Dashboard shows Quick Stats panel
- WebSocket connection established (check browser dev tools)
- Exercise reference: complete the verification checklist in Exercise 1.2

---

## Slide 12: Module Summary and Next Steps (4 min)

**Title**: What We Covered

- Catalogizer unifies media across 5 protocols into one catalog
- 7 components span web, desktop, mobile, and API access
- Container install is fastest; manual install gives full control
- SQLite for development; PostgreSQL for production
- Next module: Getting Started with Media Management
- Homework: connect at least one storage source before Module 2
