# CLAUDE.md — catalogizer-androidtv

## Overview

Android TV application for Catalogizer, optimized for big-screen viewing and D-pad/remote navigation. Built with Kotlin, Jetpack Compose for TV, and Leanback components. Designed for 10-foot UI experience on devices like Xiaomi Mi Box, NVIDIA Shield, and Android TV emulators.

- **Package**: `com.catalogizer.androidtv`
- **SDK**: compileSdk 34, minSdk 26, targetSdk 34
- **Version**: 1.1.0 (versionCode 3)

## Commands

```bash
./gradlew assembleDebug          # build debug APK
./gradlew assembleRelease        # build release APK (requires signing config)
./gradlew test                   # run unit tests
./gradlew testDebugUnitTest      # run debug unit tests only
./gradlew connectedAndroidTest   # run instrumentation tests on device/emulator
./gradlew installDebug           # install debug APK on connected TV device
./gradlew lint                   # run Android lint
./gradlew jacocoTestReport       # generate test coverage report (HTML + XML)
```

## Architecture

**MVVM: Compose UI -> ViewModel (StateFlow) -> Repository -> Room + Retrofit**

### Source Structure

```
app/src/main/java/com/catalogizer/androidtv/
  CatalogizerTVApplication.kt   # Application class
  DependencyContainer.kt        # Manual dependency injection container
  data/                          # Repository implementations, Room DAOs, Retrofit API
  ui/                            # TV-optimized Compose screens, ViewModels, navigation
  utils/                         # Utility functions
```

### Key Libraries

| Library | Purpose |
|---|---|
| Jetpack Compose (BOM 2023.10) | Declarative UI |
| Compose for TV (`tv-foundation`, `tv-material` 1.0.0-alpha10) | TV-specific composables |
| Leanback 1.0.0 | TV navigation and UI patterns |
| Material 3 | Design system |
| Navigation Compose | Screen navigation |
| Room 2.6.1 | Local database (offline cache) |
| Retrofit 2.9.0 + Gson + OkHttp 4.12 | REST API client |
| Kotlinx Serialization | JSON serialization |
| Coil | Image loading |
| ExoPlayer (Media3 + Session) | Media playback with TV session support |
| TV Provider | Home screen channel/program integration |
| DataStore Preferences | Key-value storage |
| Paging 3 | Paginated data loading |

### Dependency Injection

Uses a manual `DependencyContainer` (not Hilt) for service wiring. Initialized in `CatalogizerTVApplication`.

### Server Configuration

Same as the Android phone app. Debug builds use empty API base URL. On first launch:
1. Tries `http://localhost:8080` (works with `adb reverse tcp:8080 tcp:8080`)
2. Falls back to login screen with server URL input and Discover button
3. Server URL is persisted via DataStore

For physical TV devices on LAN, enter the server's IP (e.g., `http://192.168.0.100:8080`).

## TV-Specific Considerations

### D-pad and Remote Navigation

- **Focus handling**: All interactive elements must be focusable and have visible focus indicators for D-pad navigation.
- **Input sequence for text fields**: Press `dpad_center` BEFORE `type` command, use `KEYCODE_TAB` to move between fields. This is critical for ADB-driven QA testing.
- **No touch events**: All interactions are D-pad directional buttons, center/select, and back.

### Leanback Integration

- `tv-foundation` and `tv-material` provide TV-optimized composables (rows, cards, carousels).
- `leanback` and `leanback-preference` provide traditional TV navigation patterns and settings screens.
- `tvprovider` enables home screen channel and program recommendations.

### Media Playback

Uses Media3 ExoPlayer with `media3-session` for TV media session integration (play/pause from remote, background audio control).

## Testing

- **Unit tests**: JUnit 4 + MockK/Mockito. Coroutines test via `kotlinx-coroutines-test`. Robolectric for Android framework mocking. `unitTests.isReturnDefaultValues = true` is set for convenience.
- **Instrumentation tests**: Espresso + Compose UI testing. Run on TV device or emulator.
- **Test files**: `app/src/test/` (unit), `app/src/androidTest/` (instrumentation).
- **Coverage**: JaCoCo configured via `jacocoTestReport` task.
- **HelixQA**: Autonomous LLM-driven QA testing targets this app on physical devices (e.g., Mi Box 4 at 192.168.0.134:5555).

## JDK and Build Constraints

- **JDK 17 target**: Compile options target Java 17 (`sourceCompatibility`/`targetCompatibility = VERSION_17`, `jvmTarget = "17"`), unlike the phone app which targets JDK 21.
- **`--add-opens` JVM args**: Required in both `gradle.properties` (for Gradle JVM and Kotlin daemon) and in the `kapt` block of `build.gradle.kts`. These open `jdk.compiler` internal modules for Room annotation processing with JDK 21.
- **Kotlin daemon args**: Separate `kotlin.daemon.jvmargs` with `--add-opens` flags in `gradle.properties`.
- **JDK image transform disabled**: Multiple properties set `android.useNewJdkImageTransform=false`.
- **Signing**: Release builds read signing config from `../docker/signing/signing.properties`.
- **ProGuard**: Release builds enable minification with `proguard-rules.pro`.

## Conventions

- **Kotlin**: Coroutines for async, StateFlow for reactive state, `freeCompilerArgs = ["-Xjvm-default=all"]`.
- **Compose**: TV-optimized screens using `tv-foundation`/`tv-material` composables.
- **Offline-first**: Room database provides local cache.
- **Networking**: Uses Gson converter (unlike the phone app which uses Kotlinx Serialization converter).

## Constraints

- **Container builds**: Use Podman. Requires Android SDK 34 in the builder image.
- **Resource limits**: Gradle JVM limited to `-Xmx2048m -XX:MaxMetaspaceSize=512m`. Kotlin daemon limited to `-Xmx1024m`.
- **API keys**: Never commit `local.properties` or `.env` with real secrets.
- **ADB reverse proxy**: Must set up `adb reverse tcp:8080 tcp:8080` for each device before testing.
