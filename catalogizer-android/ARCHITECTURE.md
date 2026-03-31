# Architecture -- catalogizer-android

## Purpose

Native Android application for Catalogizer media management. Built with Kotlin and Jetpack Compose following MVVM architecture. Provides media browsing, search, playback, and collection management on Android phones and tablets.

## Structure

```
app/src/main/java/com/catalogizer/android/
  CatalogizerApplication.kt     Application class
  DependencyContainer.kt        Manual dependency injection container
  data/                          Repository implementations, Room DAOs, Retrofit API
  ui/                            Compose screens, ViewModels, navigation
app/src/test/                    Unit tests (JUnit 4 + MockK)
app/src/androidTest/             Instrumentation tests (Espresso + Compose)
```

## Key Components

- **Jetpack Compose** -- Declarative UI with Material 3 design system
- **ViewModels** -- Expose `StateFlow<UiState>` for reactive state management
- **Room 2.6.1** -- Local database for offline cache
- **Retrofit 2.9.0 + OkHttp** -- REST API client with Kotlinx Serialization
- **ExoPlayer (Media3)** -- Media playback
- **DependencyContainer** -- Manual DI (not Hilt) initialized in Application class
- **DataStore Preferences** -- Server URL and settings persistence
- **WorkManager** -- Background sync tasks
- **Paging 3** -- Paginated data loading

## Data Flow

```
Compose Screen -> ViewModel.uiState (StateFlow)
    |
    ViewModel -> Repository -> Retrofit API (remote) + Room DAO (local cache)
    |
    Server discovery: localhost:8080 (adb reverse) -> fallback to login screen with URL input
    |
    Media playback: ExoPlayer with stream URL from API
```

## Dependencies

Kotlin, Coroutines, Jetpack Compose (BOM 2023.10), Room, Retrofit, OkHttp, Kotlinx Serialization, Coil, ExoPlayer, Navigation Compose, DataStore, Paging 3, WorkManager.

## Testing Strategy

JUnit 4 + MockK/Mockito for unit tests. Coroutines test via kotlinx-coroutines-test. Robolectric for Android framework mocking. Espresso + Compose UI testing for instrumentation tests. JaCoCo for coverage reports.
