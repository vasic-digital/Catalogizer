# catalogizer-android Architecture

## Overview

Native Android application for Catalogizer media management. Built with Kotlin, Jetpack Compose, and MVVM architecture. Targets phones and tablets running Android 8+ (SDK 26).

## Layers

```
Compose UI (screens, components, theme)
       |
  ViewModels (StateFlow<UiState>)
       |
  Repositories (AuthRepository, MediaRepository, OfflineRepository)
       |
  Data Sources
    +-- Remote: Retrofit + OkHttp (CatalogizerApi)
    +-- Local:  Room (CatalogizerDatabase)
    +-- Prefs:  DataStore<Preferences>
```

## Dependency Injection

Manual `DependencyContainer` (singleton, thread-safe via `@Volatile` + double-checked locking). No Hilt or Dagger -- the container provides:

- `CatalogizerDatabase` (Room, lazy)
- `DataStore<Preferences>` (lazy)
- `CatalogizerApi` (Retrofit, recreatable for server switching)
- Repositories: `AuthRepository`, `MediaRepository`, `SyncManager`
- ViewModel factories: `createAuthViewModel()`, `createHomeViewModel()`, etc.

Initialized in `CatalogizerApplication.onCreate()` with async server URL loading from DataStore.

## Network Layer

- **Retrofit 2.9** with `kotlinx.serialization` JSON converter
- **OkHttp 4.12** with logging interceptor (BODY in debug, NONE in release)
- 30-second connect/read/write timeouts
- **Runtime server switching**: `DependencyContainer.switchServer(url)` recreates the Retrofit client
- **Server discovery**: On first launch, probes `localhost:8080` (ADB reverse proxy), then falls back to user input

## Local Database (Room)

`CatalogizerDatabase` with DAOs:
- `MediaDao` -- cached media items for offline browsing
- `FavoriteDao` -- user favorites
- `WatchProgressDao` -- playback progress tracking
- `SyncOperationDao` -- pending sync operations queue

Migrations are registered via `CatalogizerDatabase.ALL_MIGRATIONS`.

## Background Sync

`SyncManager` coordinates offline-first synchronization using WorkManager:
- `SyncWorker` runs periodic background sync
- `SyncOperation` tracks pending create/update/delete operations
- Conflict resolution on reconnection

## Navigation

Jetpack Navigation Compose with a `CatalogizerNavigation` composable defining the nav graph. Screens: Splash, Login, Home, Search, Settings.

## Key Design Decisions

- **Manual DI** over Hilt: Simpler dependency graph, no annotation processing overhead.
- **Offline-first**: Room database caches all browsed data. WorkManager ensures sync completes even if the app is killed.
- **Kotlinx Serialization** over Gson: Compile-time safety, smaller APK.
- **DataStore** over SharedPreferences: Coroutine-safe, no main-thread I/O.
- **Cleartext traffic restricted**: Only allowed to local networks (10.x, 192.168.x, 127.x) via `network_security_config.xml`.
