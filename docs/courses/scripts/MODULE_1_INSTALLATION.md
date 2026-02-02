# Module 1: Introduction & Installation - Video Scripts

---

## Lesson 1.1: What is Catalogizer?

**Duration**: 8 minutes

### Narration

Welcome to the Catalogizer video course. In this first lesson, we are going to explore what Catalogizer is and why it exists.

Catalogizer is an advanced multi-protocol media collection management system. It automatically detects, categorizes, and organizes your media files across multiple storage protocols. Whether your media lives on SMB network shares, FTP servers, NFS mounts, WebDAV endpoints, or local disks, Catalogizer brings it all together into a single, searchable catalog.

The system identifies over 50 media types, including movies, TV shows, music, games, software, documentaries, and more. It does not just list files -- it enriches them with metadata from external providers like TMDB, IMDB, TVDB, MusicBrainz, Spotify, and Steam.

Let us take a look at the main components of Catalogizer:

First, the **catalog-api**. This is the Go backend built with the Gin framework. It handles authentication, media detection, file system monitoring, and exposes a REST API under /api/v1.

Second, **catalog-web**. This is the React TypeScript frontend with Tailwind CSS. It provides a modern, responsive interface with real-time updates via WebSocket.

Third, the **desktop applications**. These are built with Tauri, combining a React frontend with a Rust backend. There is also an installer wizard for guided setup.

Fourth, the **mobile applications**. There are separate apps for Android phones and Android TV, both built with Kotlin and Jetpack Compose using MVVM architecture.

Finally, the **API client library**. This is a TypeScript library that lets you integrate Catalogizer into your own applications and automations.

### On-Screen Actions

- [00:00] Show title card: "What is Catalogizer?"
- [00:15] Display the architecture diagram from README.md showing React Web App, Go REST API, SQLCipher DB, WebSocket, Media Detection Engine, External APIs, and Multi-Protocol File System Clients
- [01:30] Highlight each protocol: SMB, FTP, NFS, WebDAV, Local
- [02:30] Show the component list with logos: catalog-api (Go gopher), catalog-web (React logo), Tauri logo, Kotlin logo, TypeScript logo
- [04:00] Demonstrate a quick walkthrough of the web UI -- dashboard with media stats
- [05:30] Show the desktop app running alongside the web version
- [06:30] Brief clip of the Android app browsing media
- [07:15] Show the API client code in an editor, making a simple call

### Key Points

- Catalogizer unifies media scattered across multiple storage protocols into one catalog
- Automatic detection of 50+ media types with external metadata enrichment
- Six main components: backend API, web frontend, desktop app, installer wizard, Android apps, API client
- Five supported protocols: SMB/CIFS, FTP/FTPS, NFS, WebDAV, local filesystem
- Real-time monitoring with WebSocket-based live updates

### Tips

> **Tip**: You do not need to use all components. The catalog-api and catalog-web are the core -- everything else extends the experience to additional platforms.

---

## Lesson 1.2: System Requirements & Architecture Overview

**Duration**: 10 minutes

### Narration

Before we install Catalogizer, let us understand what you need and how the system is structured.

For the backend, you need Go version 1.21 or later and SQLCipher for encrypted database support. For the frontend, Node.js 18 or later with npm. For Docker deployment, Docker and Docker Compose are required. The web frontend supports Chrome 90+, Firefox 88+, Safari 14+, and Edge 90+.

Now let us walk through the architecture. The system follows a layered pattern. In the backend, requests flow through Handlers, then to Services, then to Repositories, and finally to the SQLite database encrypted with SQLCipher.

The media detection pipeline is a key component. When a file is discovered, it goes through the detector, which identifies the media type. Then the analyzer processes it further. Finally, providers like TMDB and IMDB fetch enriched metadata.

For real-time updates, an event bus captures changes and pushes them through WebSocket connections to all connected clients. This means when a new file appears on an SMB share, your web browser updates automatically.

The filesystem layer is particularly elegant. There is a UnifiedClient interface defined in filesystem/interface.go. Each protocol -- SMB, FTP, NFS, WebDAV, local -- has its own client implementation. The factory in filesystem/factory.go creates the right client based on the protocol. If you ever want to add a new protocol, you just implement this interface.

On the frontend, the React application uses AuthProvider and WebSocketProvider as context wrappers around the router. Server state is managed with React Query, and auth-gated routes use the ProtectedRoute component.

The Android apps follow MVVM architecture with Compose UI at the top, ViewModels managing StateFlow, Repositories for data access, Room for offline storage, and Retrofit for API calls. Hilt handles dependency injection.

### On-Screen Actions

- [00:00] Show system requirements checklist on screen
- [01:00] Draw the layered architecture: Handler -> Service -> Repository -> SQLite
- [02:30] Animate the media detection pipeline: file -> detector -> analyzer -> providers (show TMDB, IMDB logos)
- [04:00] Show the event bus -> WebSocket -> client browser diagram
- [05:00] Open filesystem/interface.go and highlight the UnifiedClient interface
- [06:00] Open filesystem/factory.go and show client creation logic
- [07:00] Show the frontend context chain: AuthProvider -> WebSocketProvider -> Router
- [08:00] Diagram the Android MVVM stack: Compose UI -> ViewModel -> Repository -> Room + Retrofit
- [09:00] Recap with full system diagram

### Key Points

- Backend: Go 1.21+, SQLCipher; Frontend: Node.js 18+; Docker for containerized deployment
- Layered backend: Handler -> Service -> Repository -> SQLCipher-encrypted SQLite
- Media pipeline: detector -> analyzer -> providers (TMDB, IMDB, TVDB, MusicBrainz, Spotify, Steam)
- Real-time: event bus -> WebSocket pushes updates to all clients
- UnifiedClient interface abstracts all storage protocols behind a common API

### Tips

> **Tip**: The architecture is designed for extension. Adding a new storage protocol means implementing a single interface. Adding a new metadata provider follows the same pattern in the providers package.

---

## Lesson 1.3: Docker Installation

**Duration**: 12 minutes

### Narration

Docker is the fastest way to get Catalogizer running. The project includes both production and development Docker Compose configurations.

Let us start with the development environment, which is what you will use for evaluation and testing. Open your terminal and navigate to the Catalogizer project root.

Run docker-compose with the dev configuration file. This starts all required services: the PostgreSQL database, Redis cache, the catalog-api backend, and the catalog-web frontend.

The production docker-compose.yml configures PostgreSQL 15 Alpine with a dedicated user, password, and database. It sets up health checks that run pg_isready every 10 seconds. Resource limits are configured -- 2 CPUs and 2GB RAM for the database.

Redis 7 Alpine is configured with a custom redis.conf from the config directory. It also has health checks via redis-cli ping.

Nginx acts as a reverse proxy, routing requests to the appropriate service. Its configuration lives in config/nginx.conf.

There are important environment variables you must set. At minimum, you need POSTGRES_PASSWORD -- the compose file will refuse to start without it. Other variables like POSTGRES_USER and POSTGRES_DB have sensible defaults.

Let me show you how to create a .env file at the project root with the required values. You will want to set the database credentials, JWT secret, and any external API keys you plan to use -- like TMDB_API_KEY for movie metadata.

After running docker-compose up, you can verify everything is healthy by checking container status. All services should show as healthy within about 30 seconds.

### On-Screen Actions

- [00:00] Open terminal at project root
- [00:30] Show the docker-compose.dev.yml file structure
- [01:30] Run: `docker-compose -f docker-compose.dev.yml up`
- [02:30] Show containers starting in the terminal output
- [03:30] Open docker-compose.yml and walk through the PostgreSQL service configuration
- [05:00] Highlight the health check configuration and resource limits
- [06:00] Show the Redis service with config/redis.conf volume mount
- [07:00] Show the Nginx service and config/nginx.conf
- [08:00] Create a .env file with required variables: POSTGRES_PASSWORD, JWT_SECRET, TMDB_API_KEY
- [09:30] Run `docker-compose up` for production
- [10:30] Run `docker ps` to show all healthy containers
- [11:00] Open browser to http://localhost to verify the application loads
- [11:30] Show the API health endpoint

### Key Points

- Development: `docker-compose -f docker-compose.dev.yml up`
- Production: `docker-compose up` (requires .env with POSTGRES_PASSWORD)
- Services: PostgreSQL 15, Redis 7, Nginx reverse proxy, catalog-api, catalog-web
- Health checks configured for all database and cache services
- Config files for nginx and redis live in the config/ directory -- do not move them without updating volume mounts

### Tips

> **Tip**: For development, use the dev compose file. It has faster startup, hot reloading, and relaxed security settings. Never use the dev configuration in production.

> **Tip**: If you change config/nginx.conf or config/redis.conf, you need to restart the respective containers for changes to take effect.

---

## Lesson 1.4: Manual Installation (Backend & Frontend)

**Duration**: 15 minutes

### Narration

If you prefer running Catalogizer without Docker, or need to set up a development environment for contributing, let us walk through the manual installation.

Start by cloning the repository and navigating to the catalog-api directory. Run go mod tidy to install all Go dependencies. This will pull in the Gin framework, SQLCipher bindings, WebSocket libraries, and all other backend dependencies.

Next, set up your environment variables. Copy the .env.example file to .env and edit it. The critical variables are: DB_PATH for your SQLite database location, DB_ENCRYPTION_KEY which must be a 32-character key for SQLCipher encryption, PORT for the API server (default 8080), and JWT_SECRET for authentication tokens.

Initialize the database by running the migration command. Then start the API server with go run main.go. You should see the Gin server start on port 8080. Verify it works by visiting the Swagger documentation at http://localhost:8080/swagger/index.html.

Now for the frontend. Navigate to the catalog-web directory and run npm install. Create a .env.local file from .env.example. The key settings are VITE_API_BASE_URL pointing to your backend and VITE_WS_URL for WebSocket connections. There are also feature flags: VITE_ENABLE_ANALYTICS, VITE_ENABLE_REALTIME, VITE_ENABLE_EXTERNAL_METADATA, and VITE_ENABLE_OFFLINE_MODE.

Start the development server with npm run dev. The frontend will start on port 5173 with hot module replacement enabled. Open your browser and navigate to http://localhost:5173.

You can also build for production using npm run build, which generates optimized static files. Before deploying, run npm run lint and npm run type-check to ensure code quality.

### On-Screen Actions

- [00:00] Open terminal and clone the repository
- [00:30] `cd catalog-api && go mod tidy`
- [01:30] Show the .env.example file contents
- [02:30] Create and edit the .env file with database, JWT, and SMB configuration
- [04:00] Run the database migration
- [04:30] `go run main.go` -- show server startup output
- [05:30] Open browser to http://localhost:8080/swagger/index.html -- show API docs
- [06:30] `cd ../catalog-web && npm install`
- [07:30] Show .env.example and create .env.local
- [08:30] Highlight the feature flags: analytics, realtime, external metadata, offline mode
- [09:30] `npm run dev` -- show Vite startup on port 5173
- [10:30] Open browser to http://localhost:5173 -- show the login page
- [11:30] Run `npm run lint && npm run type-check` to demonstrate quality checks
- [12:30] Run `npm run build` for production output
- [13:30] Show the build output directory
- [14:00] Quick verification: log in and see the dashboard

### Key Points

- Backend: `go mod tidy` -> configure .env -> migrate database -> `go run main.go` (port 8080)
- Frontend: `npm install` -> configure .env.local -> `npm run dev` (port 5173)
- Database encryption requires a 32-character DB_ENCRYPTION_KEY
- Feature flags in .env.local control analytics, real-time updates, external metadata, and offline mode
- API documentation available at /swagger/index.html

### Tips

> **Tip**: Keep your DB_ENCRYPTION_KEY safe. If you lose it, you will not be able to access your encrypted database.

> **Tip**: During development, enable all feature flags in .env.local so you can test the full functionality of the application.

---

## Lesson 1.5: First-Time Configuration

**Duration**: 10 minutes

### Narration

With Catalogizer installed, let us configure it properly for your environment.

Starting with SMB configuration. The SMB_SOURCES variable accepts a comma-separated list of SMB URLs. For example, smb://server1/media,smb://server2/videos. You also need SMB_USERNAME, SMB_PASSWORD, and optionally SMB_DOMAIN for Active Directory environments.

Catalogizer is built to handle SMB disconnections gracefully. The resilience configuration includes SMB_RETRY_ATTEMPTS (default 5), which is how many times it will try to reconnect. SMB_RETRY_DELAY_SECONDS (default 30) controls the base delay between retries with exponential backoff. SMB_HEALTH_CHECK_INTERVAL (default 60) sets how often it pings the connection. SMB_CONNECTION_TIMEOUT (default 30) is the maximum wait for initial connections. And SMB_OFFLINE_CACHE_SIZE (default 1000) determines how many items to cache when a source goes offline.

For external metadata enrichment, add your API keys. TMDB_API_KEY is the most important one -- it provides movie and TV show metadata. SPOTIFY_CLIENT_ID and SPOTIFY_CLIENT_SECRET enable music metadata. STEAM_API_KEY helps identify game media.

The monitoring configuration controls how Catalogizer watches your sources. WATCH_INTERVAL_SECONDS (default 30) sets the scan frequency. MAX_CONCURRENT_ANALYSIS (default 5) limits parallel processing to prevent overloading. ANALYSIS_TIMEOUT_MINUTES (default 10) sets the maximum time for analyzing a single item.

Finally, set your logging preferences. LOG_LEVEL can be debug, info, warn, or error. LOG_FILE points to where logs are written.

Once everything is configured, restart the services and verify. You should see your storage sources being scanned and media items appearing in the catalog.

### On-Screen Actions

- [00:00] Open the .env file in an editor
- [00:30] Configure SMB_SOURCES with example network paths
- [01:30] Set SMB credentials: username, password, domain
- [02:30] Walk through SMB resilience settings one by one, explaining each value
- [04:00] Add external API keys: TMDB, Spotify, Steam
- [05:30] Configure monitoring: watch interval, concurrent analysis, timeout
- [06:30] Set logging configuration
- [07:00] Restart the backend service
- [07:30] Show the logs as sources are scanned
- [08:00] Open the web UI and watch media items populate in real-time
- [08:30] Navigate to a detected movie and show TMDB metadata
- [09:00] Show a music item with Spotify metadata
- [09:30] Recap all configuration sections

### Key Points

- SMB_SOURCES takes comma-separated SMB URLs with associated credentials
- Resilience settings ensure Catalogizer handles network interruptions without data loss
- External API keys (TMDB, Spotify, Steam) enable rich metadata enrichment
- Monitoring settings control scan frequency and analysis parallelism
- Log level should be "info" for production; use "debug" only for troubleshooting

### Tips

> **Tip**: Start with a single storage source and verify it works before adding more. This makes troubleshooting easier if something goes wrong.

> **Tip**: The TMDB API key is free to obtain at themoviedb.org. It is the single most valuable external integration for movie and TV show metadata.
