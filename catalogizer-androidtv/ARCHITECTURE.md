# Architecture -- catalogizer-androidtv

## Purpose

Android TV application optimized for big-screen viewing and D-pad/remote navigation. Built with Kotlin, Jetpack Compose for TV, and Leanback components. Designed for 10-foot UI experience on devices like Xiaomi Mi Box, NVIDIA Shield, and Android TV emulators.

## Structure

```
app/src/main/java/com/catalogizer/androidtv/
  CatalogizerTVApplication.kt   Application class
  DependencyContainer.kt        Manual dependency injection container
  data/                          Repository implementations, Room DAOs, Retrofit API
  ui/                            TV-optimized Compose screens, ViewModels, navigation
  utils/                         Utility functions
app/src/test/                    Unit tests (JUnit 4 + MockK)
app/src/androidTest/             Instrumentation tests
```

## Key Components

- **Compose for TV** -- tv-foundation + tv-material for TV-optimized composables (rows, cards, carousels)
- **Leanback** -- Traditional TV navigation patterns and settings screens
- **ViewModels** -- Expose StateFlow<UiState> for reactive state
- **Room 2.6.1** -- Local database for offline cache
- **Retrofit 2.9.0 + Gson + OkHttp** -- REST API client
- **ExoPlayer (Media3 + Session)** -- Media playback with TV session support
- **TV Provider** -- Home screen channel/program integration
- **D-pad navigation** -- All interactive elements focusable with visible focus indicators

## Data Flow

```
D-pad input -> Compose TV Screen -> ViewModel.uiState (StateFlow)
    |
    ViewModel -> Repository -> Retrofit API (remote) + Room DAO (local cache)
    |
    Server discovery: localhost:8080 (adb reverse tcp:8080) -> LAN IP fallback
    |
    Media playback: ExoPlayer with media3-session for TV remote control integration
```

## Dependencies

Kotlin, Coroutines, Compose for TV (tv-foundation, tv-material), Leanback, Room, Retrofit + Gson, OkHttp, Coil, ExoPlayer + media3-session, Navigation Compose, DataStore, Paging 3, TV Provider.

## Testing Strategy

JUnit 4 + MockK/Mockito for unit tests with `unitTests.isReturnDefaultValues = true`. Coroutines test via kotlinx-coroutines-test. Robolectric for framework mocking. Espresso + Compose UI testing for instrumentation. JaCoCo for coverage. HelixQA autonomous LLM-driven testing on physical TV devices.
