# Catalogizer Android

Native Android application for Catalogizer media management.

## Tech Stack

- **Kotlin** with Coroutines
- **Jetpack Compose** for UI
- **MVVM Architecture**
- **Room** for local database
- **Retrofit** for networking
- **Hilt** for dependency injection
- **Coil** for image loading

## Requirements

- **Android Studio** Hedgehog (2023.1.1) or newer
- **JDK 11**
- **Android SDK** 34 (target), 26 (minimum)

## Quick Start

```bash
# Build debug APK
./gradlew assembleDebug

# Run unit tests
./gradlew test

# Install on connected device
./gradlew installDebug
```

## Available Gradle Tasks

| Task | Description |
|------|-------------|
| `./gradlew assembleDebug` | Build debug APK |
| `./gradlew assembleRelease` | Build release APK |
| `./gradlew test` | Run unit tests |
| `./gradlew connectedAndroidTest` | Run instrumentation tests |
| `./gradlew installDebug` | Install debug on device |
| `./gradlew lint` | Run Android lint |

## Project Structure

```
app/
├── src/main/
│   ├── java/com/catalogizer/android/
│   │   ├── data/           # Repository, Room, Retrofit
│   │   ├── di/             # Hilt modules
│   │   ├── domain/         # Use cases, models
│   │   ├── ui/             # Compose screens, ViewModels
│   │   └── util/           # Utilities
│   └── res/                # Android resources
└── build.gradle.kts        # App-level build config
```

## Configuration

Debug builds start with no server URL configured. On first launch:
1. The app tries `http://localhost:8080` (works with `adb reverse tcp:8080 tcp:8080`)
2. If that fails, the login screen shows server URL input and a Discover button
3. Once a server URL is entered and connected, it is persisted across app restarts

For emulators, use `http://10.0.2.2:8080` as the server URL.
For physical devices, enter the server's LAN IP (e.g., `http://192.168.0.100:8080`).

## Network Security

The app allows cleartext traffic only to local networks (10.x, 192.168.x, 127.x) required for local SMB/FTP/NFS protocols. See `network_security_config.xml`.

## Related Documentation

- [Android Architecture](/docs/architecture/ANDROID_ARCHITECTURE.md)
- [Android Guide](/docs/guides/ANDROID_GUIDE.md)
