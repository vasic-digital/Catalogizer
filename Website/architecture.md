---
title: System Architecture
description: Multi-platform architecture overview with component relationships, layered design, and data flow diagrams
---

# System Architecture

Catalogizer is composed of seven application components that share a common Go backend. This page provides a high-level view of how the components connect, how data flows through the system, and how the modular architecture enables multi-protocol media management.

---

## Component Topology

Seven components form the Catalogizer ecosystem. All clients communicate with the backend over HTTP/3 (QUIC) with Brotli compression, falling back to HTTP/2 + gzip when HTTP/3 is unavailable.

```mermaid
graph TB
    subgraph Clients
        WEB[catalog-web<br/>React 18 / TypeScript / Vite<br/>Port 3000]
        DESKTOP[catalogizer-desktop<br/>Tauri / Rust + React]
        WIZARD[installer-wizard<br/>Tauri / Rust + React]
        ANDROID[catalogizer-android<br/>Kotlin / Compose]
        ANDROIDTV[catalogizer-androidtv<br/>Kotlin / Leanback]
        APICLIENT[catalogizer-api-client<br/>TypeScript Library]
    end

    subgraph Backend
        API[catalog-api<br/>Go 1.25 / Gin<br/>Port 8080]
    end

    subgraph Data
        SQLITE[(SQLite / SQLCipher)]
        POSTGRES[(PostgreSQL)]
        REDIS[(Redis Cache)]
        PROM[Prometheus]
    end

    subgraph Storage
        SMB[SMB/CIFS Shares]
        FTP[FTP/FTPS Servers]
        NFS[NFS Exports]
        WEBDAV[WebDAV Endpoints]
        LOCAL[Local Filesystem]
    end

    WEB -->|HTTP/3 + WebSocket| API
    DESKTOP -->|HTTP/3| API
    WIZARD -->|HTTP/3| API
    ANDROID -->|HTTP/3 via Cronet| API
    ANDROIDTV -->|HTTP/3 via Cronet| API
    APICLIENT -->|HTTP/3| API

    API --> SQLITE
    API --> POSTGRES
    API --> REDIS
    API --> PROM

    API --> SMB
    API --> FTP
    API --> NFS
    API --> WEBDAV
    API --> LOCAL
```

### Component Summary

| Component | Technology | Purpose |
|-----------|-----------|---------|
| **catalog-api** | Go 1.25, Gin | REST API backend, media detection, storage protocol abstraction, challenge runner |
| **catalog-web** | React 18, TypeScript, Vite | Web frontend with real-time updates via WebSocket |
| **catalogizer-desktop** | Tauri (Rust + React) | Cross-platform desktop app for Windows, macOS, Linux |
| **installer-wizard** | Tauri (Rust + React) | First-time setup wizard with network discovery |
| **catalogizer-android** | Kotlin, Jetpack Compose | Android phone/tablet app with offline-first architecture |
| **catalogizer-androidtv** | Kotlin, Leanback | Android TV app with D-PAD navigation and home screen channels |
| **catalogizer-api-client** | TypeScript | Reusable typed API client library |

---

## Backend Layered Architecture

The backend follows a strict Handler-Service-Repository layered architecture with dependency injection throughout.

```mermaid
graph TB
    REQ[HTTP Request] --> MW

    subgraph Middleware Stack
        MW[CORS] --> AUTH[JWT Auth]
        AUTH --> RL[Rate Limiter]
        RL --> COMP[Brotli/Gzip Compression]
        COMP --> MET[Metrics Collection]
        MET --> IV[Input Validation]
    end

    IV --> H

    subgraph Handler Layer
        H[Handler<br/>Request parsing, response formatting]
    end

    H --> S

    subgraph Service Layer
        S[Service<br/>Business logic, orchestration]
    end

    S --> R

    subgraph Repository Layer
        R[Repository<br/>Data access, SQL queries]
    end

    R --> DB

    subgraph Database
        DB[database.DB Wrapper<br/>Automatic SQL dialect rewriting]
        DB --> SQLite[(SQLite)]
        DB --> PG[(PostgreSQL)]
    end
```

### Dual Package Layout

The backend uses two package trees to separate domain logic from infrastructure:

| Package Tree | Scope | Examples |
|---|---|---|
| **Top-level** (`handlers/`, `services/`, `repository/`, `middleware/`) | Domain logic | Media entities, collections, favorites, playback, search |
| **Internal** (`internal/handlers/`, `internal/services/`, `internal/middleware/`) | Infrastructure | Auth, metrics, WebSocket, cache, media detection pipeline, SMB circuit breaker |

This separation keeps infrastructure concerns isolated from business rules, making each layer independently testable.

---

## Multi-Protocol Filesystem Abstraction

All storage access goes through a `UnifiedClient` interface defined in `filesystem/interface.go`. A factory in `filesystem/factory.go` creates the appropriate client based on protocol type. The rest of the application is protocol-agnostic.

```mermaid
graph LR
    APP[Application Code] --> UC[UnifiedClient Interface]

    UC --> SMB_C[SMB Client<br/>Circuit Breaker<br/>Offline Cache<br/>Exponential Backoff]
    UC --> FTP_C[FTP Client<br/>TLS Support]
    UC --> NFS_C[NFS Client<br/>Auto Mount]
    UC --> WEBDAV_C[WebDAV Client<br/>HTTP-based]
    UC --> LOCAL_C[Local Client<br/>Direct I/O]

    SMB_C --> SHARES[Network Shares]
    FTP_C --> FTPSERV[FTP Servers]
    NFS_C --> NFSEXP[NFS Exports]
    WEBDAV_C --> WEBSERV[WebDAV Services]
    LOCAL_C --> DISK[Local Disks]
```

The interface supports standard filesystem operations: list directories, read files, get file info, copy, and delete. The `SeekableClient` extension enables HTTP Range requests for video seeking on protocols that support random-access reads (SMB, local).

Adding a new protocol requires implementing the `UnifiedClient` interface and registering the protocol in the factory. No application-level changes are needed.

---

## Media Detection Pipeline

Scanned files pass through a multi-stage detection pipeline that identifies, analyzes, enriches, and structures media entities.

```mermaid
graph TB
    SCAN[Universal Scanner<br/>Traverse storage roots] --> DET[Detector<br/>Identify media type from<br/>filename, path, extension]

    DET --> ANA[Analyzer<br/>Extract quality metadata<br/>Resolution, codec, bitrate]

    ANA --> PROV[Metadata Providers<br/>TMDB, OMDB, MusicBrainz,<br/>OpenLibrary, Spotify, Steam]

    PROV --> AGG[Aggregation Service<br/>Post-scan entity creation]

    subgraph Entity Construction
        AGG --> TITLE[Title Parser<br/>Regex patterns per media type]
        TITLE --> ITEM[MediaItem Creation<br/>11 media types]
        ITEM --> LINK[MediaFile Linking<br/>Junction table]
        LINK --> HIER[Hierarchy Builder<br/>Show → Season → Episode<br/>Artist → Album → Song]
        HIER --> DUP[Duplicate Detection<br/>Title + type + year matching]
    end

    DUP --> ENT[Structured Media Entities<br/>/api/v1/entities]
```

### Supported Media Types

The system recognizes 11 media types, seeded in the `media_types` table:

| Type | Hierarchy | Example |
|------|-----------|---------|
| movie | Standalone | "Inception (2010)" |
| tv_show | Parent of tv_season | "Breaking Bad" |
| tv_season | Child of tv_show, parent of tv_episode | "Season 1" |
| tv_episode | Child of tv_season | "S01E01 - Pilot" |
| music_artist | Parent of music_album | "Pink Floyd" |
| music_album | Child of music_artist, parent of song | "The Dark Side of the Moon" |
| song | Child of music_album | "Time" |
| game | Standalone | "The Witcher 3" |
| software | Standalone | "Blender 4.0" |
| book | Standalone | "Dune" |
| comic | Standalone | "Saga Vol. 1" |

---

## Real-Time Event System

The backend publishes events through an internal event bus that routes them to WebSocket connections. Clients receive live updates without polling.

```mermaid
sequenceDiagram
    participant Scanner
    participant EventBus
    participant WebSocket
    participant Client

    Scanner->>EventBus: scan.started
    EventBus->>WebSocket: broadcast
    WebSocket->>Client: {"type": "scan.started"}

    loop For each file
        Scanner->>EventBus: scan.progress
        EventBus->>WebSocket: broadcast
        WebSocket->>Client: {"type": "scan.progress", "payload": {...}}
    end

    Scanner->>EventBus: media.new
    EventBus->>WebSocket: broadcast
    WebSocket->>Client: {"type": "media.new", "payload": {...}}

    Scanner->>EventBus: scan.completed
    EventBus->>WebSocket: broadcast
    WebSocket->>Client: {"type": "scan.completed"}
```

Events include scan progress, new media detection, source connection/disconnection, format conversion progress, and collection changes.

---

## Submodule Architecture

The project uses 41 independent git submodules under the vasic-digital organization. Each submodule is a standalone repository with its own tests, documentation, and ARCHITECTURE.md.

### Go Modules (22 total)

Go submodules are wired via `replace` directives in `catalog-api/go.mod`. This allows the backend to import them by their canonical module paths while resolving to local filesystem paths during development.

Key modules include Auth (JWT authentication), Database (dual-dialect abstraction), Filesystem (protocol clients), Media (detection pipeline), Cache (Redis/in-memory), EventBus (pub/sub), RateLimiter (sliding window), Security (input validation), Observability (metrics/logging), and Challenges (test framework).

### TypeScript Modules (9 total)

TypeScript submodules are linked via `file:../` references in `package.json`. They include WebSocket-Client (React hooks), UI-Components (shared component library), Media-Types (type definitions), Auth-Context (authentication provider), Media-Browser, Media-Player, Collection-Manager, Dashboard-Analytics, and the API Client.

---

## Database Architecture

The dual-dialect abstraction in `database/dialect.go` allows the same application code to run against SQLite (development) and PostgreSQL (production) without query changes.

```mermaid
graph TB
    APP[Application Code<br/>Uses ? placeholders<br/>INSERT OR IGNORE syntax] --> WRAP[database.DB Wrapper<br/>Shadowed Exec/Query/QueryRow]

    WRAP --> RW[Dialect Rewriter]

    RW -->|SQLite| SQLITE_Q[Pass through unchanged]
    RW -->|PostgreSQL| PG_Q["? → $1,$2,...<br/>INSERT OR IGNORE → ON CONFLICT DO NOTHING<br/>0/1 → FALSE/TRUE"]

    SQLITE_Q --> SQLITE[(SQLite + SQLCipher)]
    PG_Q --> PG[(PostgreSQL)]
```

Migrations live in `database/migrations/` with separate implementations for each dialect. The `InsertReturningID()` helper abstracts the difference between SQLite's `LastInsertId()` and PostgreSQL's `RETURNING id`.

---

## Deployment Topology

```mermaid
graph TB
    subgraph Production
        NGINX[Nginx Reverse Proxy<br/>TLS Termination] --> API_PROD[catalog-api Container<br/>--cpus=2 --memory=4g]
        NGINX --> WEB_PROD[catalog-web Static Files]
        API_PROD --> PG_PROD[(PostgreSQL<br/>--cpus=1 --memory=2g)]
        API_PROD --> REDIS_PROD[(Redis<br/>--cpus=1 --memory=2g)]
        API_PROD --> PROM_PROD[Prometheus + Grafana]
    end

    subgraph Development
        DEV_API[go run main.go<br/>SQLite, dynamic port] --> DEV_DB[(SQLite File)]
        DEV_WEB[npm run dev<br/>Port 3000, reads .service-port]
        DEV_WEB -->|API Proxy| DEV_API
    end
```

All containers run with resource limits enforced (max 4 CPUs, 8 GB RAM total). The container runtime is Podman (rootless), with dynamic detection supporting Docker or nerdctl as alternatives.
