# Catalogizer Codebase Architecture Guide

## Overview

**Catalogizer** is a comprehensive, multi-platform media collection management system that automatically detects, categorizes, and organizes media files across multiple storage protocols (SMB, FTP, NFS, WebDAV, local filesystem). The project follows a modern distributed architecture with clear separation of concerns across multiple components.

**Project Status:** Mature multi-component system with comprehensive testing, security scanning, and deployment automation.

---

## Table of Contents

1. [Main Project Components](#main-project-components)
2. [Technology Stack](#technology-stack)
3. [Architecture Patterns](#architecture-patterns)
4. [Component Interactions](#component-interactions)
5. [Build & Test Commands](#build--test-commands)
6. [Development Setup](#development-setup)
7. [Key Files & Entry Points](#key-files--entry-points)
8. [Configuration System](#configuration-system)
9. [Special Patterns & Conventions](#special-patterns--conventions)

---

## Main Project Components

### 1. **catalog-api** (Go Backend)
**Location:** `/Volumes/T7/Projects/Catalogizer/catalog-api/`
**Technology:** Go 1.24.0, Gin Framework, SQLite3
**Purpose:** High-performance REST API server for media cataloging, browsing, and file operations

**Key Directories:**
```
catalog-api/
├── main.go                    # Entry point (172 lines)
├── go.mod / go.sum          # Go dependencies
├── config/                   # Configuration management
├── database/                 # Database connections & migrations
│   ├── connection.go
│   └── migrations.go
├── filesystem/               # Multi-protocol file system abstraction
│   ├── interface.go          # UnifiedClient interface
│   ├── factory.go            # Protocol factory pattern
│   ├── smb_client.go         # SMB/CIFS implementation
│   ├── ftp_client.go         # FTP/FTPS implementation
│   ├── nfs_client.go         # NFS implementation
│   ├── webdav_client.go      # WebDAV implementation
│   ├── local_client.go       # Local filesystem
│   └── *_test.go             # Unit tests for each protocol
├── handlers/                 # HTTP request handlers
│   ├── auth_handler.go
│   ├── browse.go
│   ├── configuration_handler.go
│   ├── conversion_handler.go
│   ├── copy.go
│   ├── download.go
│   └── ... (more handlers)
├── internal/
│   ├── auth/                 # JWT authentication logic
│   ├── config/               # Config structs and loading
│   ├── handlers/             # Additional handlers
│   ├── media/
│   │   ├── analyzer/         # Media quality analysis
│   │   ├── database/         # Media data access layer
│   │   ├── detector/         # Media type detection patterns
│   │   ├── models/           # Media data models
│   │   ├── providers/        # External API integrations
│   │   │   ├── tmdb.go
│   │   │   ├── imdb.go
│   │   │   └── ... (other providers)
│   │   └── realtime/         # WebSocket real-time updates
│   ├── middleware/           # HTTP middleware (CORS, logging, auth)
│   ├── models/               # Shared data models
│   ├── recovery/             # Error recovery logic
│   ├── services/             # Business logic services
│   │   ├── catalog_service.go
│   │   ├── smb_service.go
│   │   └── smb_discovery_service.go
│   ├── smb/                  # SMB resilience layer
│   └── tests/                # Test utilities and fixtures
├── middleware/               # Gin middleware
├── models/                   # Data models
├── repository/               # Database access layer
├── services/                 # Business logic services
├── tests/                    # Integration and e2e tests
├── utils/                    # Utility functions
├── migrations/               # Database migration scripts
└── scripts/                  # Build and test scripts
```

**Key Patterns:**
- **Gin Framework:** HTTP routing with middleware support
- **Multi-Protocol Abstraction:** `filesystem/interface.go` defines `UnifiedClient` interface
- **Factory Pattern:** `filesystem/factory.go` creates protocol-specific clients
- **Service Layer:** Business logic separated in `services/` and `internal/services/`
- **Repository Pattern:** Data access via `repository/` layer
- **Middleware Pipeline:** Custom CORS, logging, error handling, JWT validation

**API Entry Point:** `main.go:43-172`
- Initializes Zap logger
- Loads configuration
- Opens SQLite database
- Creates Gin router with middleware
- Registers API routes under `/api/v1`
- Graceful shutdown handling

**Main API Route Groups:**
```go
/api/v1/
├── /catalog              # File browsing
├── /catalog-info         # File information
├── /search               # File search
├── /search/duplicates    # Duplicate detection
├── /download             # File downloads
├── /copy                 # File operations
├── /storage              # Storage management
├── /stats                # Statistics
└── /smb                  # SMB discovery
```

**Dependencies:**
```
github.com/gin-gonic/gin                  # HTTP framework
github.com/golang-jwt/jwt/v5              # JWT authentication
github.com/hirochachacha/go-smb2          # SMB client
github.com/jlaffaye/ftp                   # FTP client
github.com/studio-b12/gowebdav            # WebDAV client
github.com/mattn/go-sqlite3               # SQLite driver
go.uber.org/zap                           # Structured logging
golang.org/x/crypto                       # Cryptography
golang.org/x/image                        # Image processing
```

---

### 2. **catalog-web** (React Frontend)
**Location:** `/Volumes/T7/Projects/Catalogizer/catalog-web/`
**Technology:** React 18, TypeScript, Vite, Tailwind CSS, React Query
**Purpose:** Modern web UI for browsing, searching, and managing media collections

**Key Directories:**
```
catalog-web/
├── package.json           # npm dependencies & scripts
├── vite.config.ts        # Vite build config
├── tsconfig.json         # TypeScript config
├── tailwind.config.js    # Tailwind CSS configuration
├── postcss.config.js     # PostCSS config for Tailwind
├── index.html            # HTML entry point
├── src/
│   ├── main.tsx          # React app bootstrap
│   ├── App.tsx           # Root component (111 lines)
│   ├── index.css         # Global styles
│   ├── components/
│   │   ├── auth/         # Authentication components
│   │   │   ├── LoginForm.tsx
│   │   │   ├── RegisterForm.tsx
│   │   │   └── ProtectedRoute.tsx
│   │   ├── layout/       # Layout components
│   │   │   └── Layout.tsx
│   │   ├── media/        # Media display components
│   │   │   ├── MediaBrowser.tsx
│   │   │   ├── MediaCard.tsx
│   │   │   └── MediaFilters.tsx
│   │   └── ui/           # Reusable UI components
│   │       └── ConnectionStatus.tsx
│   ├── contexts/         # React contexts (state management)
│   │   ├── AuthContext.tsx
│   │   └── WebSocketContext.tsx
│   ├── pages/            # Page components
│   │   ├── Dashboard.tsx
│   │   ├── MediaBrowser.tsx
│   │   └── Analytics.tsx
│   ├── lib/              # Utility functions
│   ├── types/            # TypeScript type definitions
│   └── test/             # Test utilities
```

**Key Architecture Patterns:**
- **Context API:** AuthContext and WebSocketContext for global state
- **React Query:** Server state management and caching (TanStack React Query v4)
- **Component Composition:** Nested component hierarchy with clear responsibility
- **Custom Hooks:** Reusable logic (implied, used with React Query)
- **Protected Routes:** `ProtectedRoute` component for access control
- **Real-time Updates:** WebSocket integration via WebSocketContext

**Entry Point:** `src/App.tsx:14-111`
- Provider setup: AuthProvider → WebSocketProvider → Router
- Public routes: `/login`, `/register`
- Protected routes: `/dashboard`, `/media`, `/analytics`, `/admin`, `/profile`, `/settings`
- Permission-based access control on routes

**UI Framework Stack:**
```
React 18.2.0              # UI library
TypeScript 4.9.3          # Type safety
Vite 4.1.0                # Build tool
Tailwind CSS 3.2.7        # Styling
React Router DOM 6.8.0    # Routing
TanStack React Query 4    # Server state
Zustand 4.3.6             # State management
Framer Motion 10.0.1      # Animations
Headless UI               # Accessible components
Heroicons 2.0.16          # Icon library
```

**Key Dependencies:**
```json
{
  "@tanstack/react-query": "^4.24.6",     // Server state
  "@headlessui/react": "^1.7.13",         // Accessible components
  "@heroicons/react": "^2.0.16",          // Icons
  "react-router-dom": "^6.8.0",           // Routing
  "socket.io-client": "^4.6.1",           // Real-time updates
  "framer-motion": "^10.0.1",             // Animations
  "recharts": "^2.5.0",                   // Charts
  "tailwindcss": "^3.2.7",                // CSS framework
  "zod": "^3.20.6",                       // Schema validation
  "react-hook-form": "^7.43.5",           // Form handling
  "axios": "^1.6.0"                       // HTTP client
}
```

---

### 3. **catalogizer-api-client** (TypeScript/JavaScript Library)
**Location:** `/Volumes/T7/Projects/Catalogizer/catalogizer-api-client/`
**Technology:** TypeScript, Axios, WebSocket
**Purpose:** Cross-platform API client library (npm package: `@catalogizer/api-client`)

**Key Directories:**
```
catalogizer-api-client/
├── package.json              # npm package config
├── tsconfig.json            # TypeScript configuration
├── src/
│   ├── index.ts             # Main export (296 lines)
│   ├── services/
│   │   ├── catalog.ts       # Catalog/browsing service
│   │   ├── media.ts         # Media management service
│   │   ├── auth.ts          # Authentication service
│   │   └── ... (other services)
│   ├── types/               # TypeScript type definitions
│   │   ├── media.ts
│   │   ├── catalog.ts
│   │   └── ... (other types)
│   └── utils/               # Utility functions
│       ├── http.ts          # HTTP client setup
│       ├── websocket.ts     # WebSocket utilities
│       └── error.ts         # Error handling
├── dist/                    # Compiled JavaScript output
│   ├── index.js
│   └── index.d.ts          # TypeScript definitions
└── __tests__/               # Unit tests
```

**Purpose & Usage:**
- Published as npm package for use in web, desktop, and mobile clients
- Provides TypeScript-safe API client with type definitions
- Handles authentication, WebSocket connections, and API calls
- Can be used with Axios for HTTP or fetch API

**Dependencies:**
```json
{
  "axios": "^1.4.0",        // HTTP client
  "ws": "^8.13.0"           // WebSocket client
}
```

**Entry Point:** `src/index.ts` exports:
- API service classes (CatalogService, MediaService, AuthService)
- TypeScript type definitions for all API responses
- WebSocket client configuration
- HTTP client configuration and interceptors

---

### 4. **catalogizer-desktop** (Tauri Desktop App)
**Location:** `/Volumes/T7/Projects/Catalogizer/catalogizer-desktop/`
**Technology:** Tauri 2.0, React, TypeScript, Rust backend
**Purpose:** Cross-platform desktop application (Windows, macOS, Linux)

**Key Directories:**
```
catalogizer-desktop/
├── package.json             # npm dependencies
├── src/                     # React frontend source
│   ├── main.tsx
│   ├── App.tsx
│   └── ... (React components)
├── src-tauri/               # Tauri Rust backend
│   ├── Cargo.toml          # Rust dependencies
│   ├── src/                # Rust source code
│   │   └── main.rs         # Tauri app entry point
│   ├── tauri.conf.json     # Tauri configuration
│   └── icons/              # App icons
└── vite.config.ts          # Vite build configuration
```

**Tauri Configuration** (`src-tauri/tauri.conf.json`):
```json
{
  "build": {
    "beforeDevCommand": "npm run dev",
    "beforeBuildCommand": "npm run build",
    "devPath": "http://localhost:1420",
    "distDir": "../dist"
  },
  "app": {
    "windows": [{
      "title": "Catalogizer",
      "width": 1200,
      "height": 800,
      "minWidth": 800,
      "minHeight": 600,
      "resizable": true
    }]
  },
  "bundle": {
    "active": true,
    "targets": "all",  // Windows, macOS, Linux
    "identifier": "com.catalogizer.desktop"
  }
}
```

**Rust Dependencies** (`src-tauri/Cargo.toml`):
```toml
tauri = { version = "2.0", features = ["shell-open"] }
tokio = { version = "1", features = ["full"] }
reqwest = { version = "0.11", features = ["json"] }
serde_json = "1.0"
```

**Build Targets:** Windows (MSI/NSIS), macOS (DMG/Bundle), Linux (AppImage/Deb)

---

### 5. **installer-wizard** (Tauri Installation Helper)
**Location:** `/Volumes/T7/Projects/Catalogizer/installer-wizard/`
**Technology:** Tauri 2.0, React, TypeScript, Vite
**Purpose:** User-friendly SMB configuration and network discovery wizard

**Key Directories:**
```
installer-wizard/
├── package.json             # npm dependencies
├── src/                     # React wizard UI
│   ├── main.tsx
│   ├── App.tsx
│   ├── components/          # Wizard step components
│   ├── pages/              # Wizard pages
│   ├── services/           # Wizard services
│   └── contexts/           # State management
├── src-tauri/              # Tauri backend
│   ├── Cargo.toml
│   ├── tauri.conf.json     # Wizard-specific config
│   └── src/main.rs
└── scripts/                # Badge generation, status reports
```

**Key Features:**
- 🔍 Network discovery (automatic SMB device scanning)
- ⚙️ Visual configuration wizard (step-by-step setup)
- 🧪 Connection testing (real-time SMB validation)
- 💾 File management (save/load configuration)
- 📊 Test coverage: 93%, 30/30 tests passing

**Tauri Configuration** (`src-tauri/tauri.conf.json`):
```json
{
  "productName": "Catalogizer Installation Wizard",
  "app": {
    "windows": [{
      "title": "Catalogizer Installation Wizard",
      "width": 1000,
      "height": 700,
      "resizable": true,
      "center": true
    }]
  },
  "plugins": {
    "shell": { "open": true },
    "dialog": null,
    "fs": null
  }
}
```

---

### 6. **catalogizer-android** (Android Mobile App)
**Location:** `/Volumes/T7/Projects/Catalogizer/catalogizer-android/`
**Technology:** Kotlin, Jetpack Compose, Android SDK 34, Gradle
**Purpose:** Native Android application with MVVM architecture

**Build Configuration** (`build.gradle.kts`):
```kotlin
android {
    namespace = "com.catalogizer.android"
    compileSdk = 34
    defaultConfig {
        applicationId = "com.catalogizer.android"
        minSdk = 26          // Android 8.0
        targetSdk = 34       // Android 14
    }
}

plugins {
    id("com.android.application")
    id("org.jetbrains.kotlin.android")
    id("kotlin-kapt")
    id("kotlin-parcelize")
    id("org.jetbrains.kotlin.plugin.serialization")
    id("com.google.dagger.hilt.android")  // Dependency injection
}
```

**Key Directories:**
```
catalogizer-android/
├── build.gradle.kts         # App-level build config
├── app/
│   ├── src/main/
│   │   ├── java/com/catalogizer/android/
│   │   │   ├── data/
│   │   │   │   ├── local/    # Room database
│   │   │   │   ├── remote/   # API client
│   │   │   │   ├── models/   # Data models
│   │   │   │   ├── repository/ # Data access
│   │   │   │   └── sync/     # Offline sync
│   │   │   ├── ui/
│   │   │   │   ├── theme/    # Material Design 3
│   │   │   │   ├── navigation/ # Navigation graph
│   │   │   │   └── viewmodel/ # MVVM ViewModels
│   │   │   └── ...
│   │   └── res/              # Resources (layouts, strings)
│   └── build.gradle.kts      # App module build config
└── gradle/wrapper/           # Gradle wrapper
```

**Architecture:**
- **Pattern:** MVVM with manual dependency injection
- **UI Framework:** Jetpack Compose (Material Design 3)
- **Database:** Room (SQLite abstraction)
- **Networking:** Retrofit + OkHttp
- **DI:** Hilt (Android dependency injection)
- **Min SDK:** Android 8.0 (API 26)
- **Target SDK:** Android 14 (API 34)

**Key Dependencies:**
```
androidx.compose.ui:ui                      // Compose UI
androidx.compose.material3:material3         // Material Design 3
androidx.room:room-runtime                  // Database
com.google.dagger:hilt-android              // Dependency injection
androidx.lifecycle:lifecycle-runtime        // Lifecycle management
androidx.media3:media3-exoplayer            // ExoPlayer for media playback
androidx.work:work-runtime                  // Background tasks
```

---

### 7. **catalogizer-androidtv** (Android TV App)
**Location:** `/Volumes/T7/Projects/Catalogizer/catalogizer-androidtv/`
**Technology:** Kotlin, Jetpack Compose for TV, Leanback UI
**Purpose:** Android TV application optimized for large screens and D-pad navigation

**Similar structure to Android with TV-specific optimizations:**
- Leanback UI components for 10-foot interface
- D-pad navigation support
- Large touch targets for remote control
- ExoPlayer integration for media playback

---

### 8. **Catalogizer** (Legacy Desktop - Gradle-based)
**Location:** `/Volumes/T7/Projects/Catalogizer/Catalogizer/`
**Technology:** Gradle, Kotlin (older desktop implementation)
**Status:** Legacy component - newer implementation uses Tauri (catalogizer-desktop)

---

## Technology Stack

### Backend Services
| Component | Technology | Version | Purpose |
|-----------|-----------|---------|---------|
| **API Server** | Go | 1.24.0 | REST API, file browsing, media detection |
| **Framework** | Gin | 1.9.1 | HTTP routing and middleware |
| **Database** | SQLite3 | 1.14.18 | Local data persistence |
| **Auth** | JWT | 5.3.0 | Token-based authentication |
| **Logging** | Zap | 1.26.0 | Structured logging |
| **File Protocols** | SMB2, FTP, NFS, WebDAV | Latest | Multi-protocol file access |

### Frontend (Web)
| Component | Technology | Version | Purpose |
|-----------|-----------|---------|---------|
| **UI Framework** | React | 18.2.0 | Component-based UI |
| **Language** | TypeScript | 4.9.3 | Type-safe JavaScript |
| **Build Tool** | Vite | 4.1.0 | Fast bundling |
| **Styling** | Tailwind CSS | 3.2.7 | Utility-first CSS |
| **State** | React Query | 4.24.6 | Server state management |
| **Router** | React Router | 6.8.0 | Client-side routing |
| **Animations** | Framer Motion | 10.0.1 | UI animations |

### Desktop/Installer
| Component | Technology | Version | Purpose |
|-----------|-----------|---------|---------|
| **Framework** | Tauri | 2.0+ | Cross-platform desktop apps |
| **Frontend** | React + TypeScript | 18 + 5.0 | Desktop UI |
| **Backend** | Rust | 1.60+ | Native performance |
| **Build System** | Cargo | Latest | Rust package management |

### Mobile (Android)
| Component | Technology | Version | Purpose |
|-----------|-----------|---------|---------|
| **Language** | Kotlin | 1.9.20+ | Modern Android development |
| **Build** | Gradle | 8.2+ | Android build automation |
| **UI** | Jetpack Compose | Latest | Modern declarative UI |
| **Database** | Room | Latest | SQLite abstraction layer |
| **DI** | Hilt | 2.48 | Dependency injection |
| **HTTP** | Retrofit + OkHttp | Latest | Networking |
| **Target** | Android 8.0-14 | API 26-34 | Device support |

### Cross-Platform API Client
| Component | Technology | Version | Purpose |
|-----------|-----------|---------|---------|
| **Language** | TypeScript | 5.0.0 | Type-safe client library |
| **HTTP** | Axios | 1.4.0 | HTTP requests |
| **WebSocket** | WS | 8.13.0 | Real-time communication |
| **Package** | npm | Latest | Distribution |

### DevOps & Testing
| Component | Technology | Version | Purpose |
|-----------|-----------|---------|---------|
| **Containerization** | Docker | Latest | Application containerization |
| **Orchestration** | Docker Compose | 3.8+ | Multi-container orchestration |
| **Database (Docker)** | PostgreSQL | 15-alpine | Production database |
| **Cache (Docker)** | Redis | 7-alpine | In-memory caching |
| **Reverse Proxy** | Nginx | alpine | Load balancing, SSL termination |
| **Security Scanning** | SonarQube | Community | Code quality analysis |
| **Dependency Check** | Snyk | Free tier | Vulnerability scanning |
| **Container Scanning** | Trivy | Latest | Container image scanning |

---

## Architecture Patterns

### 1. **Multi-Protocol Abstraction Layer** (catalog-api/filesystem)

**Pattern:** Factory + Strategy Pattern
**Purpose:** Unified interface for multiple file system protocols

```
┌─────────────────────────────────┐
│   UnifiedClient Interface       │  (filesystem/interface.go)
│ - ListDirectory()               │
│ - GetFileInfo()                 │
│ - Connect()                     │
│ - Disconnect()                  │
└──────────────┬──────────────────┘
               │
    ┌──────────┼──────────┬──────────┬──────────┐
    │          │          │          │          │
┌───▼──┐  ┌───▼──┐  ┌───▼──┐  ┌───▼──┐  ┌───▼──┐
│ SMB  │  │ FTP  │  │ NFS  │  │WebDAV│  │Local │
│Client│  │Client│  │Client│  │Client│  │Client│
└──────┘  └──────┘  └──────┘  └──────┘  └──────┘

Factory (filesystem/factory.go):
  NewClient(protocol, config) → UnifiedClient
```

**Key Files:**
- `filesystem/interface.go` - Interface definition
- `filesystem/factory.go` - Protocol factory
- `filesystem/*_client.go` - Protocol implementations
- `filesystem/*_test.go` - Protocol-specific tests

**Benefit:** Easy to add new protocols without changing business logic

---

### 2. **Service Layer Architecture**

**Pattern:** Service + Repository Pattern

```
HTTP Handler (handlers/)
       ↓
   Service Layer (services/, internal/services/)
       ↓
 Repository Layer (repository/)
       ↓
  Database Layer (database/)
```

**Service Responsibilities:**
- `CatalogService` - File browsing and catalog operations
- `SMBService` - SMB-specific operations and resilience
- `SMBDiscoveryService` - Network device discovery
- `AnalyticsService` - Statistics and analytics
- Media Services - Media detection, analysis, metadata

**Benefits:**
- Separation of concerns
- Testability (mock services)
- Reusability across handlers

---

### 3. **SMB Resilience Layer** (catalog-api/internal/smb)

**Pattern:** Circuit Breaker + Cache-Aside + Retry with Exponential Backoff

```
┌─────────────────────────────────────┐
│      API Request                    │
└────────────┬────────────────────────┘
             │
    ┌────────▼─────────┐
    │ Circuit Breaker  │ (State: Closed/Open/Half-Open)
    └────────┬─────────┘
             │
    ┌────────▼──────────────┐
    │  Connection Check     │
    └─┬────────────┬────────┘
      │            │
   Success    Failure
      │            │
      │       ┌────▼─────────────┐
      │       │ Offline Cache    │
      │       │ (Return cached)  │
      │       └──────────────────┘
      │
  ┌───▼──────────────┐
  │ Normal Operation │
  └──────────────────┘

Retry Strategy:
  Attempt 1: Wait 30s
  Attempt 2: Wait 60s
  Attempt 3: Wait 120s
  Attempt 4: Wait 300s
  Attempt 5: Wait 600s
```

**Key Components:**
- Health checker for continuous monitoring
- Offline cache for metadata
- Circuit breaker to prevent cascade failures
- Event bus for status updates
- Background reconnection logic

---

### 4. **Context-Based State Management** (catalog-web)

**React Contexts:**
- `AuthContext` - User authentication and permissions
- `WebSocketContext` - Real-time WebSocket connections

```
App
├── AuthProvider
│   └── useAuth() - Access auth state and methods
├── WebSocketProvider
│   └── useWebSocket() - Access WebSocket and messages
└── Router
    ├── Public Routes (Login, Register)
    └── Protected Routes
        ├── ProtectedRoute Component
        └── Permission-based access control
```

**State Management Hierarchy:**
```
Global (Context API)
  ├── Authentication State
  ├── Authorization (Permissions)
  └── WebSocket Connection

Server State (React Query)
  ├── Media Lists
  ├── Search Results
  └── File Metadata

Component State (useState)
  ├── Form inputs
  ├── UI toggles
  └── Local component state
```

---

### 5. **MVVM Architecture** (Android/AndroidTV)

```
┌──────────────────────────────────┐
│      UI Layer (Compose)          │
│  - Screens                       │
│  - Components                    │
│  - Navigation                    │
└────────────┬─────────────────────┘
             │
    ┌────────▼──────────┐
    │  ViewModel Layer  │
    │  - Data binding   │
    │  - State mgmt     │
    │  - Business logic │
    └────────┬──────────┘
             │
    ┌────────▼──────────┐
    │   Data Layer      │
    │ - Repository      │
    │ - Local (Room)    │
    │ - Remote (API)    │
    └───────────────────┘
```

**Key Components:**
- **Views:** Composable UI functions
- **ViewModels:** State holders with LiveData/StateFlow
- **Repositories:** Abstract data sources (local/remote)
- **Data Models:** Serializable data classes

---

### 6. **Tauri Desktop Application Pattern**

```
┌─────────────────────────────────┐
│   React Frontend (TypeScript)   │
│  - User Interface               │
│  - Component Logic              │
│  - Local State Management       │
└────────────┬────────────────────┘
             │ (Tauri IPC)
    ┌────────▼────────────┐
    │  Tauri Bridge       │
    │  (Commands/Events)  │
    └────────┬────────────┘
             │
┌────────────▼────────────────────┐
│   Rust Backend                  │
│  - File System Operations       │
│  - System Integration           │
│  - Native Functionality         │
│  - HTTP Requests                │
└─────────────────────────────────┘
```

**IPC Communication:**
- Frontend → Backend: Tauri Commands
- Backend → Frontend: Tauri Events
- Type-safe communication with serde serialization

---

## Component Interactions

### Data Flow Diagram

```
User Interactions
    │
    ├─→ Web UI (catalog-web)
    │      │
    │      └─→ API Client (@catalogizer/api-client)
    │            │
    │            ├─→ HTTP Requests → REST API (catalog-api)
    │            └─→ WebSocket → WebSocket Server
    │
    ├─→ Desktop App (catalogizer-desktop)
    │      │
    │      ├─→ React Frontend
    │      │      └─→ API Client
    │      │            └─→ HTTP/WebSocket to catalog-api
    │      └─→ Rust Backend
    │            └─→ File System Operations
    │
    └─→ Mobile Apps (catalogizer-android/androidtv)
           │
           ├─→ Jetpack Compose UI
           │      └─→ API Client (Retrofit)
           │            └─→ HTTP/WebSocket to catalog-api
           └─→ Room Database
                  └─→ Offline Sync
```

### API Layer Interactions

```
catalog-api/main.go
    │
    ├─→ Config (config/config.go)
    ├─→ Database (database/connection.go)
    ├─→ Router (handlers/)
    │    ├─→ Catalog Handlers
    │    ├─→ Auth Handlers
    │    ├─→ Download Handlers
    │    ├─→ SMB Discovery Handlers
    │    └─→ Copy/Upload Handlers
    │
    ├─→ Services (services/)
    │    ├─→ CatalogService
    │    │    └─→ Filesystem Layer
    │    ├─→ SMBService
    │    │    └─→ SMB Resilience
    │    └─→ SMBDiscoveryService
    │
    ├─→ Middleware (middleware/)
    │    ├─→ CORS
    │    ├─→ Logger
    │    ├─→ ErrorHandler
    │    ├─→ RequestID
    │    └─→ JWT Auth
    │
    ├─→ Filesystem (filesystem/)
    │    ├─→ Protocol Factory
    │    └─→ Protocol Clients (SMB, FTP, NFS, WebDAV, Local)
    │
    └─→ Internal (internal/)
         ├─→ Auth (JWT validation)
         ├─→ Media (Detection, Analysis, Providers)
         ├─→ Recovery (Error handling)
         └─→ Tests (Test utilities)
```

### Real-time Update Flow

```
File System Change
    │
    ├─→ FSNotify (catalog-api/filesystem)
    │
    ├─→ Event Queue
    │
    ├─→ Media Detector
    │    ├─→ Pattern Analysis
    │    ├─→ Quality Analysis
    │    └─→ Database Storage
    │
    ├─→ Event Bus (internal/media/realtime)
    │
    ├─→ WebSocket Server
    │
    └─→ Connected Clients
         ├─→ Web UI (React Query invalidation)
         ├─→ Desktop App (UI update)
         └─→ Mobile Apps (LiveData notification)
```

---

## Build & Test Commands

### Backend (catalog-api)

```bash
# Navigate to backend
cd /Volumes/T7/Projects/Catalogizer/catalog-api

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

# Generate test coverage HTML
go test -coverprofile=coverage.html ./...
```

### Frontend (catalog-web)

```bash
# Navigate to frontend
cd /Volumes/T7/Projects/Catalogizer/catalog-web

# Install dependencies
npm install
# or
yarn install

# Development server (hot reload)
npm run dev
# Runs on http://localhost:5173 (Vite)

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

### API Client Library (catalogizer-api-client)

```bash
# Navigate to API client
cd /Volumes/T7/Projects/Catalogizer/catalogizer-api-client

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

### Desktop Application (catalogizer-desktop)

```bash
# Navigate to desktop
cd /Volumes/T7/Projects/Catalogizer/catalogizer-desktop

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

# Preview production build
npm run preview
```

### Installer Wizard (installer-wizard)

```bash
# Navigate to installer
cd /Volumes/T7/Projects/Catalogizer/installer-wizard

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

# Generate status badges
npm run badges

# Generate status report
npm run status:report

# Health check (tests + build)
npm run health:check
```

### Android App (catalogizer-android)

```bash
# Navigate to Android project
cd /Volumes/T7/Projects/Catalogizer/catalogizer-android

# Build debug APK
./gradlew assembleDebug

# Build release APK
./gradlew assembleRelease

# Run tests
./gradlew test

# Run instrumented tests
./gradlew connectedAndroidTest

# Run linting (ktlint)
./gradlew lintKotlin

# Build and run on emulator
./gradlew installDebug

# Gradle clean
./gradlew clean

# Gradle properties
cat gradle.properties
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

# Clean up volumes
docker-compose down -v
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

# Verify security setup
./scripts/verify-freemium-setup.sh

# Setup freemium tokens
./scripts/setup-freemium-tokens.sh
```

### Complete Build from Scratch

```bash
# Install with full mode
./install.sh

# Install server-only
./install.sh --mode server-only

# Build all client applications
./build-scripts/build-all.sh

# Deploy server
./deployment/scripts/deploy-server.sh --env=production

# Deploy Android
./deployment/scripts/deploy-android.sh

# Deploy desktop
./deployment/scripts/deploy-desktop.sh --target=windows
```

---

## Development Setup

### Prerequisites

```bash
# System requirements
- macOS/Linux/Windows
- Docker & Docker Compose
- Node.js 18+ (for web/desktop/installer/api-client)
- Go 1.24+ (for backend)
- Android Studio (for Android development)
- Rust & Cargo (for Tauri apps)
- Git

# Optional but recommended
- pgAdmin (PostgreSQL management)
- Redis Commander (Redis inspection)
- SonarQube (code quality)
- Snyk CLI (security scanning)
```

### Quick Start - Development Environment

```bash
# Clone repository
git clone <repo-url>
cd Catalogizer

# Run development setup
./install.sh --mode=development

# This sets up:
# - PostgreSQL with test data
# - Redis cache
# - catalog-api with hot reload
# - catalog-web with hot reload
# - pgAdmin on port 5050
# - Redis Commander on port 8081

# Access services:
# - API: http://localhost:8080
# - Web: http://localhost:5173
# - pgAdmin: http://localhost:5050
# - Redis Commander: http://localhost:8081
```

### Individual Component Development

```bash
# Terminal 1: Database services
docker-compose -f docker-compose.dev.yml up postgres redis

# Terminal 2: API development
cd catalog-api
go run main.go

# Terminal 3: Web frontend
cd catalog-web
npm run dev

# Terminal 4: Desktop app
cd catalogizer-desktop
npm run tauri:dev

# Terminal 5: Installer wizard
cd installer-wizard
npm run tauri:dev
```

### Environment Configuration

**Backend** (catalog-api/.env):
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

**Frontend** (catalog-web/.env.local):
```env
VITE_API_BASE_URL=http://localhost:8080
VITE_WS_URL=ws://localhost:8080/ws
VITE_ENABLE_ANALYTICS=true
VITE_ENABLE_REALTIME=true
```

**Docker Compose** (.env):
```env
POSTGRES_USER=catalogizer
POSTGRES_PASSWORD=dev_password
POSTGRES_DB=catalogizer_dev

JWT_SECRET=your-jwt-secret-here
```

---

## Key Files & Entry Points

### Backend Entry Points

| File | Lines | Purpose |
|------|-------|---------|
| **catalog-api/main.go** | 172 | API server bootstrap, router setup, service initialization |
| **catalog-api/config/config.go** | ~100 | Configuration loading and validation |
| **catalog-api/database/connection.go** | ~50 | Database connection and pool setup |
| **catalog-api/filesystem/factory.go** | ~100 | Protocol client factory |
| **catalog-api/services/catalog_service.go** | ~200 | Core catalog service |
| **catalog-api/services/smb_service.go** | ~200 | SMB-specific operations |
| **catalog-api/internal/auth/** | ~100 | JWT authentication |

### Frontend Entry Points

| File | Lines | Purpose |
|------|-------|---------|
| **catalog-web/src/main.tsx** | ~20 | React app bootstrap |
| **catalog-web/src/App.tsx** | 111 | Root component with routing |
| **catalog-web/src/contexts/AuthContext.tsx** | ~100 | Authentication state |
| **catalog-web/src/contexts/WebSocketContext.tsx** | ~100 | WebSocket management |
| **catalog-web/src/pages/Dashboard.tsx** | ~150 | Dashboard page |
| **catalog-web/src/pages/MediaBrowser.tsx** | ~150 | Media browsing |

### API Client Entry Point

| File | Lines | Purpose |
|------|-------|---------|
| **catalogizer-api-client/src/index.ts** | 296 | Client library exports |
| **catalogizer-api-client/src/services/** | ~100 each | Service implementations |
| **catalogizer-api-client/src/types/** | ~50 each | TypeScript type definitions |

### Desktop Entry Points

| File | Lines | Purpose |
|------|-------|---------|
| **catalogizer-desktop/src/main.tsx** | ~20 | Tauri app bootstrap |
| **catalogizer-desktop/src-tauri/src/main.rs** | ~50 | Rust backend entry |
| **catalogizer-desktop/src-tauri/tauri.conf.json** | ~65 | Tauri configuration |

---

## Configuration System

### Configuration Hierarchy

```
1. Environment Variables (highest priority)
2. .env File
3. config.json File
4. Code Defaults (lowest priority)
```

### Backend Configuration (catalog-api/config/config.go)

**Typical Structure:**
```go
type Config struct {
    Server struct {
        Port            int
        Host            string
        ReadTimeout     int
        WriteTimeout    int
    }
    Database struct {
        Database string
        MaxOpen  int
        MaxIdle  int
    }
    JWT struct {
        Secret    string
        ExpiryHrs int
    }
    SMB struct {
        Sources              []string
        Username             string
        Password             string
        RetryAttempts        int
        HealthCheckInterval  int
        OfflineCacheSize     int
    }
    Catalog struct {
        TempDir          string
        MaxArchiveSize   int64
        DownloadChunkSize int
    }
}
```

### Frontend Configuration

**Environment Variables** (catalog-web/.env.local):
```
VITE_API_BASE_URL    - Backend API URL
VITE_WS_URL          - WebSocket URL
VITE_ENABLE_*        - Feature flags
```

### Docker Configuration

**compose File Variable Substitution:**
```yaml
environment:
  POSTGRES_USER: ${POSTGRES_USER:-catalogizer}
  POSTGRES_PASSWORD: ${POSTGRES_PASSWORD:?Password required}
  API_PORT: ${API_PORT:-8080}
```

---

## Special Patterns & Conventions

### 1. **Dependency Injection**

**Go Backend:**
- Constructor injection (NewService pattern)
- Interface-based design for testability
- Dependency passing through function parameters

**Android:**
- Hilt framework for DI
- Module-based configuration
- Singleton scopes for services

**Tauri Desktop:**
- State management in Rust backend
- Dependency passing through commands

### 2. **Error Handling**

**Go:**
- Explicit error returns
- Custom error types in internal packages
- Error wrapping with context
- Middleware error handler for HTTP responses

**React:**
- Error boundaries for component failures
- Try-catch in async operations
- Query error handling with React Query
- Toast notifications for user feedback

**Android:**
- Result sealed classes for success/failure
- Repository pattern for error abstraction
- ViewModel error state management

### 3. **Testing Conventions**

**Go:** `*_test.go` files alongside source
- Unit tests for each package
- Mock interfaces for dependencies
- Table-driven tests for multiple scenarios

**TypeScript/React:**
- Jest for unit testing
- React Testing Library for component tests
- Mock API responses with MSW

**Kotlin/Android:**
- JUnit for unit tests
- Espresso for UI tests
- Mockito for mocking

### 4. **Code Organization**

**Go:**
```
package/
├── interface.go      # Defines contracts
├── implementation.go # Implementation
├── service.go       # Business logic
├── handler.go       # HTTP handling
└── *_test.go        # Tests
```

**React:**
```
src/
├── components/      # Reusable components
├── pages/          # Route pages
├── contexts/       # Global state
├── hooks/          # Custom hooks
├── services/       # API calls
├── types/          # TypeScript types
└── utils/          # Utilities
```

### 5. **Naming Conventions**

**Go:**
- PascalCase for exported (public)
- camelCase for unexported (private)
- Interface names: `Reader`, `Writer`, `Service`
- Receiver names: short, often single letter

**TypeScript:**
- PascalCase for classes, interfaces, components
- camelCase for functions, variables
- SCREAMING_SNAKE_CASE for constants
- Type/Interface suffixes: `IService`, `Props`, `State`

**Kotlin:**
- PascalCase for classes, interfaces
- camelCase for functions, variables
- ViewModel suffix: `MediaViewModel`
- Extension functions: descriptive names

### 6. **Git Workflow**

**Branch Protection:**
- Main branch requires PR reviews
- CI/CD checks must pass
- Security scans mandatory

**Commit Conventions:**
- Descriptive commit messages
- Auto-commit pattern: `Catalogizer - Auto-commit.` (current state)
- Conventional commits for releases

### 7. **Environment Management**

**Development:**
- `.env` files for local config
- Docker Compose for services
- Hot reload enabled
- Debug logging

**Testing:**
- Test databases and mocks
- Isolated test environments
- Test data fixtures

**Production:**
- Environment variables from deployment
- Encrypted secrets management
- Monitoring and alerting enabled

### 8. **Cross-Platform API Client Usage**

**Web:**
```typescript
import { CatalogService } from '@catalogizer/api-client'
const api = new CatalogService(baseURL)
const files = await api.listDirectory(path)
```

**Desktop (Tauri):**
```typescript
// Same API client
// Plus Tauri commands for native operations
const { invoke } = await import('@tauri-apps/api/tauri')
```

**Mobile (Android):**
```kotlin
// Retrofit-based API client
// Plus Room database for offline support
val api = retrofitInstance.create(CatalogService::class.java)
```

---

## Development Tips & Best Practices

### 1. **Effective Debugging**

```bash
# Backend debugging
go run main.go              # Direct execution
dlv debug ./cmd/server      # Delve debugger

# Frontend debugging
npm run dev                 # Dev server with source maps
# Browser DevTools F12

# Docker debugging
docker-compose logs -f api  # Follow logs
docker exec -it api sh      # Shell into container
```

### 2. **Performance Optimization Areas**

- **Database:** Indexed queries, connection pooling
- **Frontend:** Code splitting, lazy loading, React.memo
- **Mobile:** Background sync, efficient Room queries
- **Desktop:** Rust backend optimization

### 3. **Security Considerations**

- JWT tokens: Secure storage, refresh tokens
- Database: Encrypted connections, parameterized queries
- Frontend: Input validation, XSS prevention
- Docker: Non-root users, image scanning

### 4. **Testing Strategy**

- **Unit:** Fast, isolated tests for each function
- **Integration:** Service layer testing with dependencies
- **E2E:** Critical user flows in real environment
- **Security:** Regular scans with SonarQube/Snyk

### 5. **Scalability Patterns**

- **Stateless design:** Easy to scale horizontally
- **Caching:** Redis for shared cache
- **Database replication:** Read replicas for queries
- **Async processing:** Queue long-running tasks

---

## Summary

Catalogizer is a well-architected, modern media management system demonstrating:

✅ **Clear Separation of Concerns** - Backend, frontend, and mobile apps have distinct responsibilities
✅ **Multi-Protocol Support** - Abstracted file system layer supports SMB, FTP, NFS, WebDAV, local
✅ **Real-time Capabilities** - WebSocket integration for live updates
✅ **Cross-Platform** - Web, desktop, Android, Android TV
✅ **Resilient Design** - SMB resilience layer, circuit breakers, offline caching
✅ **Production Ready** - Docker setup, security scanning, comprehensive testing
✅ **Developer Experience** - Hot reload, type safety, clear project structure

The codebase is suitable for learning modern software architecture, contributing new features, or building similar multi-platform applications.

---

**Last Updated:** November 2024
**Repository:** Catalogizer
**Version:** 1.0.0+

For detailed documentation, see:
- `/Volumes/T7/Projects/Catalogizer/ARCHITECTURE.md` - System architecture
- `/Volumes/T7/Projects/Catalogizer/README.md` - Quick start guide
- `/Volumes/T7/Projects/Catalogizer/DEPLOYMENT.md` - Production deployment
- Individual component README files

