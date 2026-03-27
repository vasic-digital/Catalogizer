# Catalogizer Android TV

Android TV application optimized for big screen and D-pad navigation.

## Tech Stack

- **Kotlin** with Coroutines
- **Jetpack Compose for TV** (Leanback)
- **MVVM Architecture**
- **Room** for local database
- **Retrofit** for networking
- **Hilt** for dependency injection
- **Coil** for image loading

## Requirements

- **Android Studio** Hedgehog (2023.1.1) or newer
- **JDK 11**
- **Android SDK** 34 (target), 26 (minimum)
- **Android TV Emulator** or physical device

## Quick Start

```bash
# Build debug APK
./gradlew assembleDebug

# Run unit tests
./gradlew test

# Install on connected TV/emulator
./gradlew installDebug
```

## Available Gradle Tasks

| Task | Description |
|------|-------------|
| `./gradlew assembleDebug` | Build debug APK |
| `./gradlew assembleRelease` | Build release APK |
| `./gradlew test` | Run unit tests |
| `./gradlew installDebug` | Install on device |
| `./gradlew lint` | Run Android lint |

## Project Structure

```
app/
├── src/main/
│   ├── java/com/catalogizer/androidtv/
│   │   ├── data/           # Repository, Room, Retrofit
│   │   ├── di/             # Hilt modules
│   │   ├── domain/         # Use cases, models
│   │   ├── ui/             # TV-optimized Compose screens
│   │   └── util/           # Utilities
│   └── res/                # Android resources
└── build.gradle.kts        # App-level build config
```

## TV-Specific Features

- **D-pad navigation** optimized focus handling
- **Leanback UI** components for TV experience
- **Big screen layouts** optimized for 10-foot viewing
- **Remote control** support
- **Voice search** integration

## Testing on Emulator

1. Create Android TV emulator in Android Studio
2. Select "Television" category
3. Choose API 34 system image
4. Run: `./gradlew installDebug`

## Configuration

Debug builds start with no server URL configured. On first launch:
1. The app tries `http://localhost:8080` (works with `adb reverse tcp:8080 tcp:8080`)
2. If that fails, the login screen shows server URL input and a Discover button
3. Once a server URL is entered and connected, it is persisted across app restarts

For emulators, use `http://10.0.2.2:8080` as the server URL.
For physical devices, enter the server's LAN IP (e.g., `http://192.168.0.100:8080`).

## Related Documentation

- [Android TV Guide](/docs/guides/ANDROID_TV_GUIDE.md)
- [Android Architecture](/docs/architecture/ANDROID_ARCHITECTURE.md)
