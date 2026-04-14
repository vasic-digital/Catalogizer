# Desktop App -- Tauri/Rust + React Course

**Component**: catalogizer-desktop
**Language**: Rust (backend) / TypeScript + React (frontend)
**Total Duration**: 2.5 hours (4 modules)
**Level**: Intermediate

---

## Course Overview

This course covers the complete architecture of the Catalogizer desktop application built with Tauri 2.0. You will learn how the Rust backend and React frontend communicate through IPC commands and events, how to implement native desktop features like system tray and notifications, how VLC integration provides media playback, and how the application is packaged for distribution as AppImage, .deb, and .rpm.

---

### Module 1: Tauri Architecture

**Duration**: 45 minutes
**Prerequisites**: Basic Rust concepts, React/TypeScript fundamentals

#### Learning Objectives
- Explain the Tauri 2.0 architecture: Rust backend, system WebView, React frontend, IPC bridge
- Trace the application startup from Rust `main.rs` through WebView initialization to React rendering
- Compare Tauri's resource usage against Electron and understand the tradeoffs
- Navigate the project structure: `src-tauri/src/` for Rust, `src/` for React

#### Topics Covered
1. **Tauri 2.0 fundamentals**
   - System WebView instead of bundled Chromium: WebView2 (Windows), WebKitGTK (Linux), WKWebView (macOS)
   - Dramatically smaller binary sizes compared to Electron (typically under 10 MB vs 150+ MB)
   - Lower memory footprint: no separate Chromium process
   - Security model: IPC allowlist restricts which Rust commands the frontend can invoke
2. **Rust backend (`src-tauri/src/main.rs`)**
   - Tauri builder configuration: window settings, security policies, IPC command registration
   - Application state management in Rust using Tauri's managed state
   - Async command handlers using Tokio runtime
   - VLC module (`src-tauri/src/vlc/`) providing native media playback capabilities
3. **React frontend (`src/`)**
   - Standard React application rendered in the system WebView
   - `SplashScreen` component displayed during initialization
   - `Layout` component providing the application shell with navigation
   - `@tauri-apps/api` package for invoking Rust backend commands
4. **Component structure**
   - `src/components/Layout.tsx`: application shell with sidebar, header, content area
   - `src/components/VLCPlayer.tsx`: media player UI wrapping the Rust VLC backend
   - `src/components/HistoryDrawer.tsx`: recently played media drawer
   - `src/components/ProgressBadge.tsx`: playback progress indicator on media cards
   - `src/components/SplashScreen.tsx`: branded loading screen during app startup
5. **Build and development workflow**
   - `npm run tauri:dev`: starts both the Vite dev server and the Rust backend with hot-reload
   - `npm run tauri:build`: produces platform-specific installers
   - Rust compilation happens alongside frontend build in a coordinated pipeline

#### Hands-On Exercise
Run the desktop app in development mode with `npm run tauri:dev`. Open the WebView DevTools (Ctrl+Shift+I on Linux) and inspect the React component tree. Examine the Rust logs in the terminal to see backend initialization. Modify a React component and observe hot-reload updating the UI without restarting the Rust backend.

#### Key Takeaways
- Tauri uses the system's native WebView, eliminating the 150+ MB Chromium overhead of Electron
- The Rust backend handles operations that require native OS access; the React frontend handles all UI rendering
- Development mode provides hot-reload for the frontend while keeping the Rust backend running
- The VLC module in the Rust backend provides media format support beyond what browser codecs offer

---

### Module 2: IPC Communication

**Duration**: 40 minutes
**Prerequisites**: Module 1

#### Learning Objectives
- Define Tauri IPC commands in Rust and invoke them from the React frontend
- Implement the event system for bidirectional communication between Rust and React
- Manage application state in Rust that persists across IPC calls
- Handle errors across the IPC boundary with typed Result types

#### Topics Covered
1. **IPC commands (`src-tauri/src/main.rs`)**
   - `#[tauri::command]` macro annotation on Rust async functions
   - Parameter passing: JSON serialization/deserialization via serde
   - Return types: `Result<T, String>` with typed success and error payloads
   - Command registration in the Tauri builder's `invoke_handler`
2. **Frontend invocation (`src/`)**
   - `invoke()` from `@tauri-apps/api/core`: async call to a named Rust command
   - Type-safe wrappers: TypeScript functions wrapping `invoke` with proper parameter and return types
   - Error handling: try/catch around `invoke` with user-facing error messages
3. **Event system**
   - Backend-to-frontend events: Rust emits events via `app_handle.emit()`, React listens with `listen()`
   - Frontend-to-backend events: React emits via `emit()`, Rust handles with event listeners
   - Use cases: scan progress notifications, playback state changes, system events
   - Event payloads serialized as JSON, matching TypeScript interfaces
4. **State management in Rust**
   - Tauri managed state: `app.manage(MyState::default())` for application-scoped singletons
   - State access in commands: `state: tauri::State<MyState>` parameter injection
   - Thread-safe state: `Mutex<T>` or `RwLock<T>` wrapping mutable state
   - Persistence: state serialized to disk on shutdown, restored on startup
5. **Security considerations**
   - IPC allowlist: only explicitly registered commands are callable from the frontend
   - Input validation on the Rust side: never trust data from the WebView
   - Path sanitization for file system commands to prevent directory traversal
   - No shell command execution from frontend-invocable commands

#### Hands-On Exercise
Create a new IPC command in Rust that reads system information (OS version, available disk space, memory usage) and returns it as a typed struct. Write a TypeScript wrapper function with proper types. Build a React component that calls the command on mount and displays the results. Add error handling for the case where system info is unavailable.

#### Key Takeaways
- IPC commands are the only way for the frontend to access native OS capabilities
- The `#[tauri::command]` macro handles JSON serialization automatically via serde
- Events enable push-based communication from Rust to React for real-time updates like scan progress
- The IPC allowlist is a security boundary: only registered commands are callable, preventing arbitrary code execution

---

### Module 3: VLC Integration

**Duration**: 35 minutes
**Prerequisites**: Module 1, Module 2

#### Learning Objectives
- Understand how the Rust VLC module wraps LibVLC for native media playback
- Implement playback controls via IPC commands: play, pause, seek, volume, stop
- Build the React VLC player component with transport controls and progress tracking
- Handle VLC library linking across Linux, macOS, and Windows

#### Topics Covered
1. **VLC module (`src-tauri/src/vlc/`)**
   - `mod.rs`: module structure and public API surface
   - `commands.rs`: Tauri IPC commands wrapping VLC operations
   - LibVLC instance management: creation, configuration, cleanup
   - Media parsing: format detection, duration, codec information, subtitle tracks
2. **Playback commands**
   - Play: load media URI (local file or streaming URL from catalog-api), start playback
   - Pause/Resume: toggle playback state
   - Seek: absolute position seek with millisecond precision
   - Volume: get/set volume level (0-100)
   - Stop: stop playback and release media resources
   - Track selection: audio track, subtitle track, video track switching
3. **React player component (`src/components/VLCPlayer.tsx`)**
   - Transport controls: play/pause button, seek bar, volume slider, fullscreen toggle
   - Progress polling: periodic `invoke` calls to get current position from VLC backend
   - Keyboard shortcuts: space (play/pause), left/right arrows (seek), up/down (volume)
   - Subtitle display configuration: font size, color, position
4. **Library linking**
   - Linux: system LibVLC from package manager (`libvlc-dev`), linked at build time
   - Feature gate: VLC support compiled conditionally based on build configuration
   - Fallback behavior when VLC is not available: graceful degradation with user notification
   - Build script (`build.rs`) detecting VLC library paths per platform

#### Hands-On Exercise
Play a local video file through the VLC player and exercise all transport controls. Switch audio and subtitle tracks during playback. Examine the IPC command flow in the Rust logs: observe the sequence of play, position polling, and seek commands. Modify the polling interval and observe the effect on seek bar smoothness vs CPU usage.

#### Key Takeaways
- VLC integration via Rust provides native codec support far beyond browser-embedded players
- Playback state lives in the Rust backend; the React frontend polls position via IPC for progress display
- The VLC feature is conditionally compiled, so the app builds and runs even on systems without LibVLC
- Library linking is platform-specific and handled by the build script, not hardcoded paths

---

### Module 4: System Integration

**Duration**: 30 minutes
**Prerequisites**: Modules 1-3

#### Learning Objectives
- Implement system tray integration with status display and quick actions
- Configure desktop notifications for scan completion and new media detection
- Build platform-specific installers: AppImage, .deb, and .rpm for Linux
- Handle application lifecycle: single-instance enforcement, auto-start, update checks

#### Topics Covered
1. **System tray**
   - Tray icon with Catalogizer branding
   - Context menu: open app, start scan, show recent media, settings, quit
   - Tray icon status indicators: idle, scanning, sync in progress
   - Click behavior: left-click opens/focuses the main window, right-click opens context menu
2. **Desktop notifications**
   - Native OS notification via Tauri notification API
   - Scan completion: "Scan complete: 1,247 files processed, 89 new entities"
   - New media detection: "New movie detected: The Matrix (1999)"
   - Notification click action: open the app and navigate to the relevant entity
3. **Packaging and distribution**
   - AppImage: single-file portable Linux distribution
   - `.deb`: Debian/Ubuntu package with dependency declarations (libvlc, libwebkit2gtk)
   - `.rpm`: Fedora/RHEL package with spec file
   - `APPIMAGE_EXTRACT_AND_RUN=1` required in container builds (no FUSE available)
   - Container builds use `podman build --network host` to resolve download issues
   - Build output in `releases/` directory (gitignored, not version-controlled)
4. **Application lifecycle**
   - Single-instance enforcement: second launch attempt focuses the existing window
   - Graceful shutdown: cleanup VLC resources, flush pending state, close database connections
   - Native file dialog integration: OS file picker for local storage root selection
   - Window state persistence: size, position, and maximized state restored on relaunch
5. **XDG integration (Linux)**
   - Desktop entry file for application launcher integration
   - MIME type associations for media files
   - `xdg-open` stub handling in containerized builds where the real `xdg-open` is unavailable

#### Hands-On Exercise
Build the desktop app with `npm run tauri:build` and locate the generated installer in the build output. Install the AppImage and verify the system tray icon appears. Trigger a scan and observe the desktop notification. Close the main window and verify the app continues running in the system tray. Right-click the tray icon and use the context menu to reopen the window.

#### Key Takeaways
- System tray keeps the app accessible without occupying taskbar space, with status indicators for ongoing operations
- Desktop notifications surface important events (scan complete, new media) without requiring the app to be focused
- Container builds require `APPIMAGE_EXTRACT_AND_RUN=1` and `--network host` to produce working packages
- Single-instance enforcement prevents resource conflicts from multiple app launches
