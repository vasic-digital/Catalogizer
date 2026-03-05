# Module 9: Desktop and Mobile Apps - Script

**Duration**: 60 minutes
**Module**: 9 - Desktop and Mobile Apps

---

## Scene 1: Tauri Desktop App (0:00 - 25:00)

**[Visual: Screenshot of the Catalogizer desktop application]**

**Narrator**: Welcome to Module 9. Catalogizer is a true multi-platform application. Beyond the web frontend, it includes desktop applications built with Tauri and mobile applications built with Kotlin and Jetpack Compose. Let us start with the desktop.

**[Visual: Show `catalogizer-desktop/` directory structure]**

**Narrator**: The desktop app uses Tauri 2.0, which pairs a Rust backend with a React frontend. Unlike Electron, Tauri uses the operating system's native WebView (WebView2 on Windows, WebKitGTK on Linux, WKWebView on macOS), resulting in dramatically smaller binary sizes and lower memory usage.

**[Visual: Show the Tauri architecture diagram: React UI <-> IPC Bridge <-> Rust Backend]**

**Narrator**: The architecture has three layers. The React frontend runs in the WebView and handles all UI rendering. The Rust backend handles native operations -- file system access, system tray, notifications, and auto-updates. IPC commands and events bridge the two layers.

**[Visual: Show Rust backend IPC commands]**

**Narrator**: IPC commands are Rust functions annotated with Tauri macros. The frontend invokes them by name, passing JSON arguments. The backend processes the request and returns a result. This is type-safe on both sides.

```rust
// catalogizer-desktop/src-tauri/src/main.rs
#[tauri::command]
async fn scan_local_directory(path: String) -> Result<ScanResult, String> {
    // Native filesystem operations
    // Return result to frontend
}

#[tauri::command]
async fn get_system_info() -> Result<SystemInfo, String> {
    // OS version, available storage, CPU info
}
```

**[Visual: Show React frontend calling IPC commands]**

**Narrator**: The React frontend uses Tauri's `invoke` function to call backend commands. The call is asynchronous and returns a typed result.

```typescript
// catalogizer-desktop/src/lib/tauri.ts
import { invoke } from '@tauri-apps/api/core';

async function scanDirectory(path: string): Promise<ScanResult> {
  return invoke('scan_local_directory', { path });
}

async function getSystemInfo(): Promise<SystemInfo> {
  return invoke('get_system_info');
}
```

**[Visual: Show native file dialogs]**

**Narrator**: Tauri provides native file dialog APIs. When users add a local storage root, the desktop app shows the operating system's native folder picker -- not a web-based file browser. This provides a familiar, accessible experience.

**[Visual: Show auto-update mechanism]**

**Narrator**: Auto-updates use Tauri's built-in updater. The app checks a configured endpoint for new versions, downloads the update, and applies it on restart. Update checks happen on launch and on a configurable interval.

**[Visual: Show the installer wizard (`installer-wizard/`)]**

**Narrator**: The installer wizard is a separate Tauri application that guides first-time users through setup. It detects the user's environment, helps configure storage roots, sets up the backend, and verifies connectivity -- all through a step-by-step UI.

**[Visual: Show build commands]**

**Narrator**: Building the desktop apps:

```bash
# Development
cd catalogizer-desktop
npm run tauri:dev

# Production build
npm run tauri:build
# Outputs: .deb, .AppImage (Linux), .msi (Windows), .dmg (macOS)
```

**[Visual: Show AppImage container build note]**

**Narrator**: When building in containers, the `APPIMAGE_EXTRACT_AND_RUN=1` environment variable must be set because containers lack FUSE support needed for AppImage self-extraction.

---

## Scene 2: Android Development (25:00 - 45:00)

**[Visual: Show `catalogizer-android/` directory structure]**

**Narrator**: The Android app uses Kotlin with Jetpack Compose for the UI, following the MVVM (Model-View-ViewModel) architecture pattern. Room handles local persistence, Retrofit manages API communication, and Hilt provides dependency injection.

**[Visual: Architecture diagram: Compose UI -> ViewModel (StateFlow) -> Repository -> Room + Retrofit]**

**Narrator**: The data flow is unidirectional. Compose UI observes ViewModels. ViewModels expose StateFlow for reactive state. Repositories abstract data sources -- Room for local cache and Retrofit for the API. Hilt wires everything together at compile time.

**[Visual: Show ViewModel with StateFlow]**

**Narrator**: ViewModels use `StateFlow` for observable state. The UI collects the flow and recomposes whenever the state changes. Sealed classes represent different states: Loading, Success, and Error.

```kotlin
// catalogizer-android/app/src/main/java/digital/vasic/catalogizer/viewmodel/
class MediaViewModel @Inject constructor(
    private val repository: MediaRepository
) : ViewModel() {

    private val _state = MutableStateFlow<MediaState>(MediaState.Loading)
    val state: StateFlow<MediaState> = _state.asStateFlow()

    fun loadMedia(storageRootId: Long) {
        viewModelScope.launch {
            _state.value = MediaState.Loading
            repository.getMedia(storageRootId)
                .onSuccess { _state.value = MediaState.Success(it) }
                .onFailure { _state.value = MediaState.Error(it.message) }
        }
    }
}

sealed class MediaState {
    object Loading : MediaState()
    data class Success(val items: List<MediaItem>) : MediaState()
    data class Error(val message: String?) : MediaState()
}
```

**[Visual: Show Compose UI consuming ViewModel state]**

**Narrator**: Compose UI functions collect the StateFlow and render based on state. When the ViewModel updates, only the affected composables recompose.

```kotlin
@Composable
fun MediaScreen(viewModel: MediaViewModel = hiltViewModel()) {
    val state by viewModel.state.collectAsStateWithLifecycle()

    when (state) {
        is MediaState.Loading -> CircularProgressIndicator()
        is MediaState.Success -> MediaList(items = (state as MediaState.Success).items)
        is MediaState.Error -> ErrorMessage(message = (state as MediaState.Error).message)
    }
}
```

**[Visual: Show Room database for offline support]**

**Narrator**: Room provides type-safe local persistence. Media items, files, and user preferences are cached locally. When the device is offline, the app serves cached data. When online, it syncs with the API and updates the local cache.

**[Visual: Show Retrofit API client]**

**Narrator**: Retrofit generates the HTTP client from interface definitions. Each API endpoint is a suspend function, making it coroutine-compatible. OkHttp handles HTTP/3 via Cronet integration and Brotli decompression.

**[Visual: Show Hilt dependency injection]**

**Narrator**: Hilt annotations (`@HiltViewModel`, `@Inject`, `@Module`, `@Provides`) wire dependencies at compile time. No runtime reflection, no service locator pattern. This makes the code testable and the dependency graph explicit.

**[Visual: Show build configuration]**

**Narrator**: The Android build requires `jvmToolchain(17)` in the Gradle configuration. When running on JDK 21, `--add-opens` flags are needed in `gradle.properties` for kapt compatibility.

```bash
# Build and test
cd catalogizer-android
./gradlew test              # Unit tests
./gradlew assembleDebug     # Debug APK
```

---

## Scene 3: Android TV (45:00 - 60:00)

**[Visual: Show `catalogizer-androidtv/` directory structure]**

**Narrator**: The Android TV app adapts Catalogizer for the living room. It uses the Leanback UI framework designed for TV screens and remote control navigation.

**[Visual: Show Leanback UI components]**

**Narrator**: Leanback provides TV-optimized components: `BrowseSupportFragment` for the main browsing grid, `DetailsSupportFragment` for media details, `PlaybackSupportFragment` for video playback, and `SearchSupportFragment` for voice and text search.

**[Visual: Show remote control navigation]**

**Narrator**: TV navigation uses the D-pad (directional pad) -- Up, Down, Left, Right, and Select. Every interactive element must be focusable and provide clear visual focus indicators. The Leanback framework handles focus management automatically for its built-in components.

**[Visual: Show media playback integration]**

**Narrator**: Media playback on Android TV uses ExoPlayer with Leanback's transport controls. The app supports streaming from the Catalogizer API with adaptive bitrate selection and subtitle support.

**[Visual: Show the shared code between Android and Android TV]**

**Narrator**: The Android and Android TV apps share their data layer -- repositories, API clients, and database schemas. Only the UI layer differs. This is achieved through shared Kotlin modules and consistent MVVM patterns.

**[Visual: Show Gradle configuration differences]**

**Narrator**: The Android TV Gradle configuration mirrors the mobile app but targets the `leanback` feature. It was upgraded to Gradle 8.11.1 for JDK 21 compatibility, with the same `jvmToolchain(17)` and `--add-opens` JVM arguments.

```bash
# Build and test
cd catalogizer-androidtv
./gradlew test              # Unit tests
./gradlew assembleDebug     # Debug APK
```

**[Visual: Course title card]**

**Narrator**: Catalogizer runs everywhere -- web, desktop, mobile, and TV. Tauri gives us native desktop performance with a web tech frontend. Kotlin Compose provides reactive Android UIs. Leanback adapts the experience for big screens. In Module 10, we ensure all of this works correctly through comprehensive testing.

---

## Key Code Examples

### Tauri IPC Communication
```rust
// Rust backend command
#[tauri::command]
async fn connect_to_server(host: String, port: u16) -> Result<ConnectionStatus, String> {
    // Attempt connection to catalog-api
}
```

```typescript
// React frontend invocation
const status = await invoke<ConnectionStatus>('connect_to_server', {
  host: 'localhost',
  port: 8080,
});
```

### Android Build Notes
```properties
# gradle.properties (required for JDK 21 + kapt)
org.gradle.jvmargs=-Xmx2g \
  --add-opens=jdk.compiler/com.sun.tools.javac.main=ALL-UNNAMED \
  --add-opens=jdk.compiler/com.sun.tools.javac.processing=ALL-UNNAMED
```

### Component Build Commands
```bash
# Desktop
cd catalogizer-desktop && npm run tauri:build

# Installer wizard
cd installer-wizard && npm run tauri:build

# Android
cd catalogizer-android && ./gradlew assembleDebug

# Android TV
cd catalogizer-androidtv && ./gradlew assembleDebug
```

---

## Quiz Questions

1. How does Tauri differ from Electron for desktop applications?
   **Answer**: Tauri uses the OS native WebView instead of bundling Chromium, resulting in dramatically smaller binary sizes (10-20 MB vs 150+ MB) and lower memory usage. The backend is Rust instead of Node.js, providing better performance and memory safety. IPC between frontend and backend is type-safe via Tauri commands.

2. What is the MVVM architecture pattern used in the Android app?
   **Answer**: Model-View-ViewModel. The View (Compose UI) observes the ViewModel via StateFlow. The ViewModel contains business logic and calls Repository methods. Repositories abstract data sources (Room for local cache, Retrofit for API). Data flows unidirectionally: Repository -> ViewModel -> View. User actions flow: View -> ViewModel -> Repository.

3. Why does the Android TV app use the Leanback framework instead of standard Compose?
   **Answer**: Leanback provides TV-optimized components designed for remote control navigation (D-pad), large screens, and the "10-foot UI" viewing distance. It handles focus management, provides browsing grids, detail views, and playback controls specifically designed for TV interaction patterns. Standard Compose lacks these TV-specific behaviors.

4. What build configuration is needed for Kotlin projects running on JDK 21?
   **Answer**: `jvmToolchain(17)` must be set in the Gradle build script to target JDK 17 bytecode. `--add-opens` flags must be added to `gradle.properties` for kapt (annotation processing) compatibility with JDK 21's module system, specifically opening `jdk.compiler` packages to `ALL-UNNAMED`.
