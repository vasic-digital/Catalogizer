# Module 4: Multi-Platform - Slide Outlines

---

## Slide 4.0.1: Title Slide

**Title**: Multi-Platform Experience

**Subtitle**: Access Your Catalog from Android, TV, Desktop, and Custom Integrations

**Speaker Notes**: Catalogizer is not limited to the web browser. This module covers four additional ways to access your media catalog: Android phone, Android TV, desktop application, and the API client library for custom integrations.

---

## Slide 4.1.1: Android Mobile App

**Title**: Your Catalog in Your Pocket

**Bullet Points**:
- Built with Kotlin and Jetpack Compose
- MVVM architecture: Compose UI -> ViewModel (StateFlow) -> Repository -> Room + Retrofit
- Hilt dependency injection throughout
- Offline mode via Room database local caching
- Same JWT authentication as the web interface
- Build: `./gradlew assembleDebug` in `catalogizer-android/`

**Visual**: Phone mockup showing the Android app media browser

**Speaker Notes**: The Android app shares the same user accounts, favorites, and collections as the web interface. Changes on one platform appear on the other. Offline mode means you can browse your catalog metadata even without network access.

---

## Slide 4.1.2: Android MVVM Architecture

**Title**: Clean Architecture on Mobile

**Visual**: Layer diagram: Compose UI -> ViewModel (StateFlow) -> Repository -> Room (local) + Retrofit (remote)

**Bullet Points**:
- Compose UI: Declarative layouts with material design components
- ViewModel: Business logic, exposes state via Kotlin StateFlow
- Repository: Single source of truth, coordinates Room and Retrofit
- Room: Local SQLite database for offline caching
- Retrofit: HTTP client for Catalogizer API calls
- Hilt: Automatic dependency injection across all layers

**Speaker Notes**: This is a standard modern Android architecture. If you are familiar with Android development, you will feel right at home. The Repository pattern means the ViewModel does not care whether data comes from the local database or the remote API.

---

## Slide 4.1.3: Offline Mode

**Title**: Works Without Network

**Bullet Points**:
- Room database stores previously loaded metadata locally
- Browse catalog, view details, and access favorites offline
- When connectivity returns, automatic sync with the server
- Configure offline caching in app settings
- Pre-load frequently accessed content before going offline

**Speaker Notes**: Offline mode is not just a fallback. It makes the app feel fast even on slow networks, because frequently accessed data is served from the local database. The sync happens transparently when the network is available.

---

## Slide 4.2.1: Android TV App

**Title**: Media Browsing on the Big Screen

**Bullet Points**:
- Built with Kotlin/Compose, optimized for large screens
- D-pad navigation for TV remote control
- Horizontal category rows: Recently Added, Movies, TV Shows, Music, Collections
- Card-based layout with large thumbnails
- Fullscreen video playback with remote control support
- Build: `./gradlew assembleDebug` in `catalogizer-androidtv/`

**Visual**: TV screen mockup showing horizontal category rows

**Speaker Notes**: The TV app prioritizes visual browsing. Everything is designed to look great from across the room and be navigable with a simple remote control. No touch screen needed.

---

## Slide 4.2.2: Shared Architecture

**Title**: Two Apps, One Foundation

**Bullet Points**:
- Mobile and TV apps share the same data layer (Repositories, Room, Retrofit)
- Only the UI layer differs: TV uses large-screen Composables
- Bug fixes in shared code benefit both apps simultaneously
- Same Gradle build system and dependency management
- Unit tests: `./gradlew test` in either project directory

**Visual**: Venn diagram showing shared (Repository, Room, Retrofit, Hilt) and unique (Mobile UI, TV UI) layers

**Speaker Notes**: This shared architecture is a significant engineering advantage. When a bug is fixed in the Repository layer, both apps get the fix. When a new API endpoint is added, both apps can access it through the shared Retrofit client.

---

## Slide 4.3.1: Desktop Application

**Title**: Native Desktop Experience with Tauri

**Bullet Points**:
- Tauri framework: React frontend + Rust backend
- IPC communication: React calls Rust commands; Rust pushes events to React
- Native features: system tray, file system access, desktop notifications
- Development: `npm run tauri:dev` (hot reloading for frontend)
- Production: `npm run tauri:build` (MSI for Windows, DMG for macOS, AppImage/deb for Linux)

**Visual**: Desktop app screenshot showing native window chrome with the catalog UI

**Speaker Notes**: Tauri gives us the best of both worlds. The familiar web UI runs inside a native window with access to OS features. The Rust backend handles native operations that JavaScript cannot do, like direct filesystem access and system tray integration.

---

## Slide 4.3.2: Installer Wizard

**Title**: Guided First-Time Setup

**Bullet Points**:
- Separate Tauri app in the `installer-wizard/` directory
- Step-by-step configuration: server URL, authentication, initial settings
- Specialized Rust modules: `network.rs` (connectivity testing), `smb.rs` (SMB configuration)
- Validates connections before saving configuration
- Exports configuration files for the main application

**Speaker Notes**: The Installer Wizard removes the friction of first-time setup. Instead of manually editing configuration files, users follow a guided process that validates each step. It even tests network connectivity and SMB share accessibility before proceeding.

---

## Slide 4.3.3: Tauri IPC Pattern

**Title**: How React Talks to Rust

**Bullet Points**:
- **Commands**: React calls Rust functions via `invoke()` (request-response)
- **Events**: Bidirectional; Rust pushes updates to React, React sends events to Rust
- Rust backend has full access to native OS APIs
- Frontend gets web development ergonomics with native capabilities
- Type-safe communication between the two layers

**Visual**: Diagram: React Component -> invoke("command_name") -> Rust Handler -> Return Value -> React Component

**Speaker Notes**: The IPC pattern is similar to a REST API but local. React calls a Rust function by name, passes arguments, and gets a return value. Events work like WebSocket messages but between the frontend and backend within the same application.

---

## Slide 4.4.1: API Client Library

**Title**: Programmatic Access for Custom Integrations

**Bullet Points**:
- TypeScript library in the `catalogizer-api-client/` directory
- Three main areas: `services/` (API methods), `types/` (interfaces), `utils/` (helpers)
- Automatic JWT token management with refresh
- Full TypeScript type coverage for all requests and responses
- Build: `npm run build`; Test: `npm run test`

**Visual**: Code snippet showing client initialization, authentication, and a media query

**Speaker Notes**: The API client is for developers who want to build their own integrations. Home automation scripts, custom dashboards, batch processing tools -- anything that needs to talk to Catalogizer programmatically.

---

## Slide 4.4.2: API Client Usage Examples

**Title**: What You Can Build

**Bullet Points**:
- Custom dashboards tailored to specific use cases
- Automation scripts: auto-organize new media into collections
- Third-party integrations: connect Catalogizer to home automation systems
- Batch processing: bulk metadata updates, format conversions
- Custom clients: build for platforms not officially supported
- Test files in `src/services/__tests__/` serve as usage examples

**Speaker Notes**: The test files are the best starting point for learning the API client. They demonstrate every endpoint with working code. Import them, modify them, and build your own integrations.

---

## Slide 4.4.3: Module 4 Summary

**Title**: What We Covered

**Bullet Points**:
- Android app: Kotlin/Compose with MVVM, offline mode via Room, Hilt DI
- Android TV: Large-screen optimized with D-pad navigation, shared data layer
- Desktop: Tauri (Rust + React) with native OS integration and IPC
- Installer Wizard: Guided setup with connection validation
- API Client: TypeScript library for custom integrations with full type coverage

**Next Steps**: Module 5 -- Administration (User Management, Security, Monitoring, Backup, Troubleshooting)

**Speaker Notes**: Students should now understand all the ways to access Catalogizer. The web interface is the primary tool, but mobile, TV, desktop, and API access extend the experience to every context. Module 5 shifts focus to running and maintaining a Catalogizer instance.
