# Android Phone -- Kotlin/Compose Course

**Component**: catalogizer-android
**Language**: Kotlin / Jetpack Compose / Material 3
**Total Duration**: 3.5 hours (5 modules)
**Level**: Intermediate

---

## Course Overview

This course covers the complete architecture of the Catalogizer Android phone application. You will learn the MVVM architecture with Jetpack Compose, Material 3 UI design, offline-first data management with Room and DataStore, media playback with ExoPlayer/Media3, and background synchronization with WorkManager. The app connects to catalog-api for server-side data while maintaining a local cache for offline access.

---

### Module 1: MVVM Architecture

**Duration**: 50 minutes
**Prerequisites**: Kotlin fundamentals, basic Android development, Jetpack Compose basics

#### Learning Objectives
- Trace the data flow through the MVVM layers: Compose UI -> ViewModel (StateFlow) -> Repository -> Room + Retrofit
- Understand the dependency injection setup using the manual `DependencyContainer` pattern
- Explain how `StateFlow` drives reactive UI updates without lifecycle-unsafe observers
- Navigate the package structure: `ui/`, `data/`, `utils/`

#### Topics Covered
1. **Application entry point (`CatalogizerApplication.kt`)**
   - Custom `Application` class initializing the `DependencyContainer`
   - `CatalogizerWorkerFactory` wiring WorkManager with dependency injection
   - Application-scoped singletons: Retrofit instance, Room database, repositories
2. **Dependency container (`DependencyContainer.kt`)**
   - Manual dependency injection without Hilt annotation processing overhead
   - Lazy initialization of repositories: `AuthRepository`, `MediaRepository`, `OfflineRepository`, `WebSocketRepository`
   - Retrofit client configuration with base URL, interceptors, and timeout settings
   - Room database instance shared across repositories
3. **ViewModel layer (`ui/viewmodel/`)**
   - `AuthViewModel`: login/logout state, token management, session persistence
   - `HomeViewModel`: media library state, category lists, recently added items
   - `MainViewModel`: navigation state, connection status, global UI state
   - `SearchViewModel`: search query, results, filter state with debounced execution
   - `StateFlow` exposing immutable UI state; `MutableStateFlow` private to ViewModel
4. **Repository layer (`data/repository/`)**
   - `AuthRepository`: JWT token storage in DataStore, login/logout API calls, session refresh
   - `MediaRepository`: media entity CRUD via Retrofit, local cache via Room
   - `OfflineRepository`: offline queue for mutations made without network connectivity
   - `WebSocketRepository`: real-time event subscription and message handling
5. **Data flow pattern**
   - UI observes `StateFlow` from ViewModel via `collectAsState()` in Compose
   - ViewModel calls repository suspend functions from `viewModelScope`
   - Repository decides: fetch from Room cache or Retrofit remote, merge results
   - Error handling via sealed `Result` classes propagated to UI as state

#### Hands-On Exercise
Trace a media entity fetch from the Home screen: identify the `collectAsState()` call in the Compose UI, follow the `StateFlow` in `HomeViewModel`, step into the `MediaRepository` method, and examine both the Room DAO query and the Retrofit API call. Add a new ViewModel property exposing a filtered media list and observe it updating the UI reactively.

#### Key Takeaways
- `StateFlow` provides lifecycle-safe reactive updates without the risks of `LiveData` observer leaks
- The `DependencyContainer` provides manual DI that is simpler to debug than annotation-based frameworks
- Repositories abstract the data source decision (local vs remote) from ViewModels
- Sealed `Result` classes provide exhaustive error handling at the UI layer

---

### Module 2: Compose UI

**Duration**: 40 minutes
**Prerequisites**: Module 1, Jetpack Compose fundamentals

#### Learning Objectives
- Build screens using Material 3 components with Catalogizer's custom theme
- Implement navigation between screens using Compose Navigation
- Create responsive layouts that adapt to different phone screen sizes
- Apply state hoisting patterns for testable Compose components

#### Topics Covered
1. **Theme system (`ui/theme/`)**
   - Material 3 dynamic color with custom brand palette
   - Dark and light theme variants
   - Typography scale matching Catalogizer design language
   - Shape system for cards, buttons, and containers
2. **Screen implementations (`ui/screens/`)**
   - `home/`: media library grid with category tabs, pull-to-refresh
   - `login/`: credential form with validation, biometric authentication option
   - `search/`: search bar with live results, type filter chips, recent searches
   - `settings/`: server URL configuration, theme selection, cache management, account settings
3. **Navigation (`ui/navigation/`)**
   - Compose Navigation with typed routes
   - Bottom navigation bar with Home, Search, Settings destinations
   - Deep linking support for media entity URLs
   - Navigation state preservation across configuration changes
4. **Reusable components (`ui/components/`)**
   - Media card with cover art (Coil image loading), title, metadata badges
   - Loading skeletons for progressive content display
   - Error states with retry actions
   - Pull-to-refresh wrapper for list screens
5. **Splash screen (`ui/splash/`)**
   - Animated splash with Vasic Digital branding
   - Auth token validation during splash: valid token proceeds to Home, expired token redirects to Login

#### Hands-On Exercise
Create a new screen that displays a media entity detail view with Compose. Implement state hoisting by extracting the screen's state into a ViewModel. Add the screen to the navigation graph with a typed route parameter for the entity ID. Apply the Material 3 theme and verify it renders correctly in both light and dark modes.

#### Key Takeaways
- Material 3 dynamic color adapts the app's palette to the device's wallpaper colors while maintaining brand identity
- State hoisting moves state ownership to ViewModels, making Compose screens stateless and testable
- Compose Navigation with typed routes provides compile-time safety for screen arguments
- The splash screen doubles as an auth gate: it validates the stored token before allowing entry

---

### Module 3: Offline-First with Room

**Duration**: 40 minutes
**Prerequisites**: Module 1, SQLite/Room basics

#### Learning Objectives
- Design Room entities and DAOs that mirror the backend's media entity schema
- Implement the offline queue pattern for mutations made without network connectivity
- Configure DataStore for key-value persistence (auth tokens, user preferences)
- Apply cache invalidation strategies that balance freshness with offline availability

#### Topics Covered
1. **Room database (`data/local/`)**
   - Entity definitions matching backend media types: MediaItem, MediaFile, Collection
   - DAO interfaces with Flow-returning queries for reactive updates
   - Type converters for complex types: lists, enums, timestamps
   - Database migrations for schema evolution across app versions
2. **Offline repository (`data/repository/OfflineRepository.kt`)**
   - Queued mutations: add-to-collection, favorite toggle, playlist modification stored locally
   - Sync on reconnect: queued operations replayed against the API in order
   - Conflict detection: server-side changes during offline period resolved with last-write-wins
   - Operation deduplication: repeated toggles collapsed before sync
3. **DataStore for preferences**
   - JWT access and refresh tokens stored in encrypted DataStore
   - User preferences: theme, default media type filter, notification settings
   - Server URL and connection history
   - Migration from SharedPreferences for legacy support
4. **Cache strategy**
   - Network-first with Room fallback: attempt Retrofit call, fall back to Room cache on failure
   - Cache TTL: media metadata cached for 1 hour, entity lists cached for 5 minutes
   - Background refresh: SyncWorker updates Room cache periodically
   - Cache size management: automatic eviction of oldest entries when storage limits are reached
5. **Sync indicators in UI**
   - Badge showing number of pending offline operations
   - Sync progress indicator during reconnection
   - Last-synced timestamp displayed in settings

#### Hands-On Exercise
Put the device in airplane mode and browse previously cached media entities. Add an item to a collection while offline and verify it appears in the offline queue. Re-enable network and observe the sync process replaying the queued mutation. Inspect the Room database using Database Inspector to see the cached entities.

#### Key Takeaways
- Room provides reactive queries via Flow, so the UI updates automatically when the local cache changes
- The offline queue ensures no user action is lost during network outages
- DataStore replaces SharedPreferences for type-safe, coroutine-friendly key-value storage
- Cache invalidation must balance freshness (short TTL) with offline availability (keep stale data as fallback)

---

### Module 4: Media Playback

**Duration**: 35 minutes
**Prerequisites**: Module 1, Module 2

#### Learning Objectives
- Integrate ExoPlayer/Media3 for video and audio playback with streaming support
- Load and display cover art using Coil with placeholder and error states
- Implement playback progress tracking synchronized with the backend
- Handle audio focus and background playback for music content

#### Topics Covered
1. **Player integration (`ui/player/`)**
   - ExoPlayer/Media3 setup with streaming data source for catalog-api media endpoints
   - Player UI: transport controls, seek bar, volume, fullscreen toggle
   - Video surface rendering in Compose using `AndroidView` bridge
   - Audio-only mode with lock-screen controls and notification media session
2. **Cover art loading**
   - Coil image loader configured with disk cache and memory cache
   - Placeholder images by media type during loading
   - Error fallback images when cover art is unavailable
   - Image size optimization: request appropriate resolution based on display context (thumbnail vs detail)
3. **Progress tracking**
   - Periodic position reports sent to backend during playback (every 10 seconds)
   - Resume position stored locally in Room for immediate access
   - Cross-device resume: position synced via API, available on web, TV, and other devices
   - Completion detection: marking items as watched/listened when 90% progress is reached
4. **Background playback**
   - Foreground service for music playback when app is backgrounded
   - Audio focus management: pause on incoming call, duck volume for notifications
   - Media session integration for lock-screen and notification controls
   - Bluetooth metadata broadcasting for car displays and headphones

#### Hands-On Exercise
Play a video file from the media library and examine the streaming request in Android Studio's Network Profiler. Pause playback, close the app, reopen it, and verify the resume position. Play a music file, background the app, and control playback from the notification shade. Check the backend API to see the stored progress.

#### Key Takeaways
- ExoPlayer/Media3 handles format detection, buffering, and adaptive streaming automatically
- Coil provides efficient image loading with multi-level caching (memory + disk)
- Progress tracking enables cross-device resume by syncing position to the backend
- Background music playback requires a foreground service with proper audio focus management

---

### Module 5: Sync and Background Work

**Duration**: 25 minutes
**Prerequisites**: Modules 1-4

#### Learning Objectives
- Configure WorkManager for periodic background synchronization
- Implement the SyncManager orchestrating sync operations across repositories
- Handle sync conflicts between local offline changes and server-side updates
- Monitor sync health and diagnose failed operations

#### Topics Covered
1. **SyncManager (`data/sync/SyncManager.kt`)**
   - Orchestrates sync across all repositories: media, collections, playlists, favorites
   - Sync order: auth token refresh first, then entity sync, then mutation replay
   - Error aggregation: individual operation failures do not abort the entire sync
   - Sync result reporting to UI via StateFlow
2. **SyncWorker (`data/sync/SyncWorker.kt`)**
   - WorkManager `PeriodicWorkRequest` running every 6 hours
   - Network constraint: sync only when connected (WiFi or cellular)
   - Battery constraint: defer sync when battery is low
   - `CatalogizerWorkerFactory` injecting dependencies into workers
3. **SyncService (`data/sync/SyncService.kt`)**
   - Immediate sync triggered by app launch
   - Manual sync triggered by pull-to-refresh gesture
   - Incremental sync: only fetch entities modified since last sync timestamp
   - Full sync: complete re-download triggered by user or after extended offline period
4. **SyncOperation (`data/sync/SyncOperation.kt`)**
   - Sealed class representing pending operations: Create, Update, Delete, Reorder
   - Serialization to Room for persistence across app restarts
   - Retry logic with exponential backoff for transient network failures
   - Operation status tracking: pending, in-progress, completed, failed
5. **Conflict resolution**
   - Last-write-wins for simple fields (title edits, favorite toggles)
   - Merge for collection membership (union of local and remote additions)
   - User notification for unresolvable conflicts requiring manual choice

#### Hands-On Exercise
Trigger a manual sync and observe the SyncManager logs in Logcat. Make offline changes, wait for the periodic SyncWorker to execute, and verify the changes are replayed. Simulate a conflict by editing the same entity on web and mobile while offline, then observe the resolution strategy. Use Android Studio's App Inspection to monitor WorkManager task execution.

#### Key Takeaways
- WorkManager ensures sync runs reliably even after app restarts or device reboots
- The SyncManager orchestrates sync order: auth refresh, entity sync, then offline mutation replay
- Incremental sync minimizes data transfer by only fetching changes since the last sync timestamp
- Conflict resolution uses last-write-wins for simple fields and merge for collection membership
