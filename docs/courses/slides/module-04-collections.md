# Module 4: Multi-Platform Experience - Slide Deck Outline

**Total Slides**: 12
**Estimated Duration**: 55 minutes

---

## Slide 1: Title Slide (2 min)

**Title**: Multi-Platform Experience

- Android, Android TV, Desktop, Installer Wizard, API Client
- Prerequisites: Module 2 completed
- By the end: understand how to build and use every client platform

---

## Slide 2: Android Mobile Architecture (5 min)

**Title**: Android App -- MVVM with Compose

- Kotlin + Jetpack Compose UI
- ViewModel with StateFlow for reactive state management
- Repository pattern: Room (local cache) + Retrofit (API calls)
- Hilt dependency injection throughout the app
- Requires jvmToolchain(17) and --add-opens JVM args for kapt

---

## Slide 3: Android Mobile Features (5 min)

**Title**: Browsing and Managing Media on Mobile

- Browse catalog with grid/list view optimized for touch
- Manage favorites and search media on the go
- Offline mode with Room database local caching
- Auto-sync when connection is restored
- Demo: install APK and browse the catalog
- Exercise reference: Exercise 4.1 -- build and install the Android app

---

## Slide 4: Android TV App (5 min)

**Title**: Leanback UI for the Big Screen

- Leanback UI optimized for large screens and D-pad remote control
- Browse collections, search, and play media directly on TV
- Shared Kotlin/Compose architecture with the mobile app
- ADB reverse proxy for development: adb reverse tcp:8080
- D-pad navigation: dpad_center before type, KEYCODE_TAB between fields

---

## Slide 5: Android TV Deployment (4 min)

**Title**: Building and Deploying to Android TV

- ./gradlew assembleDebug for debug builds
- adb connect <device-ip>:5555 for wireless ADB
- adb install to deploy the APK to the TV device
- Package: com.catalogizer.androidtv
- Exercise reference: Exercise 4.2 -- deploy to an Android TV device

---

## Slide 6: Desktop Application (5 min)

**Title**: Tauri Desktop App -- Rust + React

- Tauri architecture: React frontend with Rust backend via IPC
- Commands and events bridge JavaScript and Rust
- Native OS features: file system integration, system tray, notifications
- npm run tauri:dev for development, npm run tauri:build for release
- Demo: launch the desktop app and browse local media

---

## Slide 7: Installer Wizard (4 min)

**Title**: Guided First-Time Setup

- Tauri-based installer wizard for new users
- Step-by-step configuration: server URL, credentials, storage sources
- Validates connectivity and creates initial configuration
- Automatically launches the main desktop app on completion
- Exercise reference: Exercise 4.3 -- run the installer wizard

---

## Slide 8: API Client Library (5 min)

**Title**: TypeScript API Client for Custom Integrations

- Install: npm install @vasic-digital/catalogizer-api-client
- Authenticate with username/password, receive JWT token
- Client services mirror the API: media, collections, favorites, search
- Type-safe request and response interfaces
- Demo: authenticate and list media from a Node.js script

---

## Slide 9: API Client Advanced Usage (4 min)

**Title**: Building Automations With the API Client

- Pagination helpers for large result sets
- Error handling with typed error responses
- Token refresh handled automatically
- Build custom scripts: batch imports, automated scans, reports
- Exercise reference: Exercise 4.4 -- write a script that creates a collection

---

## Slide 10: Cross-Platform Sync (4 min)

**Title**: Real-Time Sync Across Devices

- WebSocket pushes events to all connected clients
- Favorite a movie on the phone, see it on the web immediately
- Collection updates propagate to all platforms
- Scan progress visible on every connected device
- Demo: make a change on one platform and observe it on another

---

## Slide 11: Building All Platforms (5 min)

**Title**: Release Build Pipeline

- scripts/release-build.sh --container builds all 7 components
- Per-component builders in scripts/lib/build-*.sh
- SHA256 change detection skips unchanged components
- Build framework: Build/ submodule with versioning and orchestration
- All builds run in containers with resource limits

---

## Slide 12: Module Summary and Next Steps (3 min)

**Title**: What We Covered

- Android mobile app with MVVM + Compose and offline support
- Android TV with leanback UI for remote-controlled browsing
- Desktop app via Tauri with native OS integration
- API client library for custom TypeScript integrations
- Cross-platform real-time sync via WebSocket
- Next module: Administration and Configuration
