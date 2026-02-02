# Module 4: Multi-Platform Usage - Video Scripts

---

## Lesson 4.1: Android Mobile App

**Duration**: 15 minutes

### Narration

The Catalogizer Android app, found in the catalogizer-android directory, brings your entire media catalog to your phone. It is built with modern Android technologies: Kotlin, Jetpack Compose for the UI, and follows MVVM architecture.

Let us start with installation. You can build the app from source using Gradle. Navigate to the catalogizer-android directory and run ./gradlew assembleDebug. This produces an APK you can install on any Android 8.0 or later device.

Once installed, launch the app and configure the connection to your Catalogizer server. Enter the API URL -- this is the same URL your web frontend uses -- and log in with your credentials. The app uses the same JWT authentication as the web interface.

The architecture is clean and well-structured. At the top is the Compose UI layer with declarative composable functions. These observe state from ViewModels via Kotlin StateFlow. ViewModels contain the business logic and call into Repository classes for data access. Repositories use Retrofit for API calls to the Catalogizer backend and Room for local database storage.

Room is particularly important because it enables offline mode. When your phone loses connectivity, you can still browse media that has been previously loaded. Room stores the metadata locally, and when connectivity returns, it syncs with the server.

Hilt handles dependency injection throughout the app. This means ViewModels, Repositories, and services are all automatically provided with their dependencies.

The app supports the same core features as the web interface: browsing your catalog, searching, managing favorites, viewing collections, and playing media. The mobile-optimized UI makes it comfortable to use on smaller screens with touch navigation.

### On-Screen Actions

- [00:00] Open terminal at catalogizer-android directory
- [00:30] Run `./gradlew assembleDebug` -- show build progress
- [01:30] Show the generated APK file location
- [02:00] Install the APK on an Android device (or emulator)
- [02:30] Launch the app -- show the login screen
- [03:00] Enter the server URL and credentials
- [03:30] Log in and show the main screen
- [04:00] Browse the media catalog on mobile -- swipe through items
- [04:30] Tap on a media item to see details with metadata
- [05:00] Use the search feature on mobile
- [05:30] Navigate to favorites and show the mobile favorites view
- [06:00] Browse collections on the phone
- [06:30] Play a video file from the mobile app
- [07:00] Show offline mode: disconnect from network, browse cached data
- [07:30] Reconnect and show sync happening
- [08:00] Show the project structure in an IDE: app/src/main/java/com/
- [09:00] Open a Composable UI file and show the declarative layout
- [09:30] Open a ViewModel and show StateFlow usage
- [10:30] Open a Repository class showing Retrofit + Room integration
- [11:00] Show a Room entity and DAO
- [11:30] Show the Hilt module providing dependencies
- [12:30] Run `./gradlew test` to show unit tests passing
- [13:00] Show test files alongside source files
- [14:00] Final overview of the mobile app

### Key Points

- Build with `./gradlew assembleDebug` in the catalogizer-android directory
- MVVM architecture: Compose UI -> ViewModel (StateFlow) -> Repository -> Room + Retrofit
- Hilt for dependency injection throughout the app
- Room database enables offline mode with local caching
- Same JWT authentication as the web interface
- Supports browsing, search, favorites, collections, and media playback on mobile
- Unit tests via `./gradlew test`

### Tips

> **Tip**: Enable offline caching in the app settings before going somewhere with limited connectivity. This pre-loads your frequently accessed media metadata for offline browsing.

> **Tip**: The Android app shares the same user account and favorites as the web interface. Changes on one platform appear on the other.

---

## Lesson 4.2: Android TV App

**Duration**: 12 minutes

### Narration

The Catalogizer Android TV app, in the catalogizer-androidtv directory, is designed specifically for the living room experience. It uses the same Kotlin and Compose foundation as the mobile app but with a UI optimized for large screens and remote control navigation.

The TV app prioritizes visual browsing. Media items are displayed in large, easy-to-see cards arranged in horizontal rows by category. Navigation is designed around the D-pad -- up, down, left, right, and select -- making it natural to use with a TV remote.

Building the TV app follows the same process as the mobile app. Navigate to catalogizer-androidtv and run ./gradlew assembleDebug. Install the resulting APK on your Android TV device.

After configuring the server connection, the home screen presents your media library organized by category. You might see rows for "Recently Added", "Movies", "TV Shows", "Music", and your collections.

Search on TV uses a voice-friendly interface when available, or an on-screen keyboard. Results display in the same card-based layout.

Playing media on TV is the primary use case. Select a video and it plays in fullscreen, optimized for the TV display. The player supports remote control for play, pause, seek, and volume. Subtitle selection is available through a settings menu accessible via the remote.

The shared architecture with the mobile app means the same Repositories and data layer are reused. Only the UI layer differs, with TV-specific Composables replacing the mobile layouts. This shared foundation means bug fixes and features in the data layer benefit both apps simultaneously.

### On-Screen Actions

- [00:00] Show the catalogizer-androidtv project structure
- [00:30] Run `./gradlew assembleDebug`
- [01:00] Install on an Android TV device
- [01:30] Launch the app -- show the TV-optimized login screen
- [02:00] Configure server connection
- [02:30] Show the home screen with horizontal category rows
- [03:00] Navigate with D-pad: move between rows and items
- [03:30] Select a movie -- show the detail screen on TV
- [04:00] Show metadata and poster art on the large display
- [04:30] Press play -- show fullscreen video playback
- [05:00] Use remote control: pause, seek forward, seek back
- [05:30] Open subtitle settings via remote menu
- [06:00] Navigate to search and demonstrate
- [06:30] Browse collections on TV
- [07:00] Show favorites on the TV interface
- [07:30] Compare the TV UI structure to the mobile app in the code
- [08:30] Show shared Repository classes between mobile and TV
- [09:00] Show TV-specific Composable layouts
- [09:30] Show the build.gradle.kts configuration
- [10:30] Run `./gradlew test` for the TV app
- [11:00] Final demonstration of the TV experience

### Key Points

- Optimized for large screens and remote control (D-pad) navigation
- Card-based layout with horizontal category rows
- Same Kotlin/Compose/MVVM architecture as mobile app
- Shared data layer (Repositories, Room, Retrofit) with the mobile app
- Fullscreen video playback with remote control support
- Build with `./gradlew assembleDebug` in catalogizer-androidtv directory

### Tips

> **Tip**: Pair a Bluetooth keyboard with your Android TV for faster search input. The on-screen keyboard works but is slower with a remote.

> **Tip**: Organize your most-watched content into collections. They appear as dedicated rows on the TV home screen, making it easy to find your favorites.

---

## Lesson 4.3: Desktop Application

**Duration**: 15 minutes

### Narration

The Catalogizer desktop application is built with Tauri, a modern framework that combines a web frontend with a native Rust backend. This gives you the full web interface experience with native operating system integration.

The project lives in the catalogizer-desktop directory. The frontend is a React application -- similar to catalog-web -- while the backend is Rust code that handles native OS operations. Communication between the two happens through Tauri's IPC system using commands and events.

To develop the desktop app, you need Node.js for the frontend and the Rust toolchain for the backend. Start the development environment with npm run tauri:dev. This launches both the frontend dev server and the Rust backend, with hot reloading for the frontend.

For production builds, run npm run tauri:build. This compiles the Rust backend, bundles the frontend, and produces a native installer for your platform -- MSI for Windows, DMG for macOS, or AppImage/deb for Linux.

There is also the Installer Wizard, a separate Tauri application in the installer-wizard directory. This provides a guided setup experience for first-time users. It walks through server configuration, authentication, and initial settings with a step-by-step interface. The wizard uses the same Tauri architecture with its own React frontend and Rust backend.

The desktop app provides native features not available in the browser. System tray integration keeps Catalogizer accessible even when the window is minimized. Native file system access allows direct file operations. Desktop notifications alert you to new media discoveries or completed conversions.

The IPC layer is key to understanding how the desktop app works. React components call Tauri commands -- Rust functions exposed to the frontend. The Rust backend processes these commands and can access native APIs, perform file operations, or communicate with the system. Events flow in both directions, allowing the Rust backend to push updates to the React frontend.

### On-Screen Actions

- [00:00] Open the catalogizer-desktop project directory
- [00:30] Show the project structure: src/ for React, src-tauri/ for Rust
- [01:00] Open a React component file
- [01:30] Open a Rust command file showing IPC handlers
- [02:00] Run `npm run tauri:dev` -- show both servers starting
- [03:00] The desktop app window opens -- show the native frame
- [03:30] Browse the catalog in the desktop app -- looks like the web but with native chrome
- [04:00] Show system tray icon
- [04:30] Demonstrate native file operations
- [05:00] Show a desktop notification for a new media detection
- [05:30] Navigate the full UI: dashboard, media, collections, search
- [06:00] Play media using the built-in player
- [06:30] Close the app but show it persists in system tray
- [07:00] Click system tray to restore the window
- [07:30] Show the installer-wizard directory
- [08:00] Open the installer wizard project structure
- [08:30] Run `npm run tauri:dev` in the installer-wizard directory
- [09:00] Walk through the wizard steps: server config, auth, settings
- [10:00] Run `npm run tauri:build` for the desktop app
- [10:30] Show the build output: native installer file
- [11:00] Show the IPC command pattern in the Rust code
- [12:00] Show the IPC event pattern
- [12:30] Show how React calls Tauri commands
- [13:00] Discuss the contexts directory in the desktop app
- [13:30] Show the services and utilities
- [14:00] Final overview of the desktop experience

### Key Points

- Built with Tauri: React frontend + Rust backend with IPC communication
- Development: `npm run tauri:dev` for hot-reloading dev environment
- Production: `npm run tauri:build` creates native installers (MSI, DMG, AppImage/deb)
- Installer Wizard provides guided first-time setup (installer-wizard directory)
- Native features: system tray, file system access, desktop notifications
- IPC uses Tauri commands (React calls Rust) and events (bidirectional)

### Tips

> **Tip**: Use the Installer Wizard for first-time setup. It validates your server connection and configuration before the main app launches, preventing common setup issues.

> **Tip**: The desktop app can stay in the system tray and notify you of new media discoveries. This makes it a great background companion while you work.

---

## Lesson 4.4: API Client Library

**Duration**: 13 minutes

### Narration

The catalogizer-api-client is a TypeScript library that provides programmatic access to the Catalogizer API. It lives in the catalogizer-api-client directory and is designed for developers who want to build custom integrations or automations.

The library is organized into three main areas. The services directory contains client classes for each API endpoint group. The types directory defines TypeScript interfaces for all request and response objects. The utils directory provides helper functions for common operations.

To use the library, install it as a dependency in your project. Then import the client, configure it with your server URL and credentials, and start making API calls. Authentication is handled automatically -- the client manages JWT tokens, including refresh when they expire.

Let me walk through a typical workflow. First, create a client instance with the server URL. Then authenticate by calling the login method with your credentials. Now you can list media, search the catalog, manage favorites and collections, or trigger conversions -- all programmatically.

The type system is comprehensive. Every API response has a corresponding TypeScript interface, so you get full autocomplete and type checking in your editor. This makes it much harder to make mistakes when building integrations.

Testing the library is straightforward. Run npm run build to compile the TypeScript, then npm run test to execute the test suite. The tests directory contains test files that also serve as usage examples.

Common use cases include: building custom dashboards, automating media organization workflows, integrating Catalogizer with other services like home automation systems, creating batch processing scripts, and building mobile or desktop clients beyond the official ones.

### On-Screen Actions

- [00:00] Open the catalogizer-api-client directory in an editor
- [00:30] Show the project structure: src/index.ts, src/services/, src/types/, src/utils/
- [01:00] Open src/index.ts -- show the main exports
- [01:30] Open a service file -- show the API methods
- [02:30] Open a types file -- show TypeScript interfaces
- [03:30] Open a utility file
- [04:00] Create a new test script in a separate project
- [04:30] Import the client library
- [05:00] Configure with server URL
- [05:30] Authenticate: call login method
- [06:00] List media items: show the response with type hints
- [06:30] Search the catalog: show search parameters and results
- [07:00] Manage favorites: add an item, list favorites
- [07:30] Access collections programmatically
- [08:00] Show full TypeScript autocomplete in the editor
- [08:30] Run `npm run build` in the library directory
- [09:00] Run `npm run test` -- show tests passing
- [09:30] Open the test files as usage examples
- [10:00] Show a more complex example: automation script that organizes new media
- [11:00] Discuss integration possibilities: home automation, batch processing, custom UIs
- [12:00] Final overview of the API client capabilities

### Key Points

- TypeScript library for programmatic Catalogizer API access
- Organized into services (API methods), types (interfaces), and utils (helpers)
- Automatic JWT authentication with token refresh
- Full TypeScript type coverage for all API requests and responses
- Build with `npm run build`, test with `npm run test`
- Use cases: custom dashboards, automation scripts, third-party integrations, batch processing

### Tips

> **Tip**: Start by reading the test files in the library. They serve as practical usage examples for every API endpoint.

> **Tip**: The TypeScript types exported by the library are invaluable even if you are building a client in another language. They document the exact shape of every API request and response.
