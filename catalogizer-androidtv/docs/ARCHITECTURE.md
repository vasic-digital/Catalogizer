# catalogizer-androidtv Architecture

## Overview

Android TV application optimized for the 10-foot UI experience. Built with Kotlin, Jetpack Compose for TV, and Leanback components. Designed for D-pad/remote navigation on devices like Xiaomi Mi Box, NVIDIA Shield, and Android TV emulators.

## Layers

```
Compose TV UI (tv-foundation, tv-material)
       |
  ViewModels (StateFlow<UiState>)
       |
  Repositories (AuthRepository, MediaRepository, SettingsRepository)
       |
  Data Sources
    +-- Remote: Retrofit + OkHttp + Gson (CatalogizerApi)
    +-- Prefs:  DataStore<Preferences>
    +-- Discovery: NetworkDiscoveryService
```

## TV-Specific Architecture

### D-pad Navigation Model

All interaction is via D-pad directional buttons, center/select, and back. There are no touch events. Every interactive element must be:
- Focusable with a visible focus indicator
- Reachable via directional navigation
- Ordered logically for left-right and up-down traversal

For text input (login, search), the sequence is: `dpad_center` to activate the field, then `type` to enter text, then `KEYCODE_TAB` to move to the next field.

### Focus Management

Compose TV components from `tv-foundation` and `tv-material` handle focus automatically. Custom components use `Modifier.focusable()` and focus requesters. The `MediaCarousel` and `MediaCard` components implement custom focus behavior for row-based browsing.

### Leanback Integration

- `tv-foundation` / `tv-material` (1.0.0-alpha10) -- TV-optimized composables (rows, cards, carousels)
- `leanback` / `leanback-preference` -- Traditional TV navigation and settings patterns
- `tvprovider` -- Home screen channel and program recommendations via `CatalogizerTvProvider`

### Media Playback

Uses Media3 ExoPlayer with `media3-session` for TV media session integration. `MediaPlaybackService` handles background playback and remote control events (play/pause from the TV remote).

## Dependency Injection

Manual `DependencyContainer` (same pattern as the phone app). Provides:
- `SettingsRepository`, `NetworkDiscoveryService`
- `AuthRepository` with `AuthInterceptor` for automatic token injection
- `MediaRepository`, ViewModel factories
- Runtime server switching via `switchServer(url)`

## Server Discovery

On first launch:
1. Load persisted URL from DataStore via `SettingsRepository`
2. Probe `localhost:8080` (works with `adb reverse tcp:8080 tcp:8080`)
3. Fall back to login screen with manual URL entry or `NetworkDiscoveryService`

## Key Differences from Phone App

| Aspect | Phone | TV |
|--------|-------|----|
| Input | Touch | D-pad / remote |
| UI framework | Compose + Material 3 | Compose TV + Leanback |
| JSON converter | Kotlinx Serialization | Gson |
| JDK target | 21 | 17 |
| Home screen integration | None | TV Provider channels |
| Media session | Basic | Full TV session support |
| Network discovery | Manual only | `NetworkDiscoveryService` |

## Key Design Decisions

- **Compose for TV** over pure Leanback: Modern declarative UI with TV focus handling built in.
- **Gson** instead of Kotlinx Serialization: Historical choice; works reliably with OkHttp interceptors.
- **JDK 17 target**: Better compatibility with TV device runtimes than JDK 21.
- **AuthInterceptor**: Automatically attaches JWT tokens to all API requests, unlike the phone app which handles auth at the repository level.
