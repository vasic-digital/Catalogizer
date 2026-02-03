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

Debug builds connect to `http://10.0.2.2:8080` (localhost from emulator).

For physical devices, update `API_BASE_URL` in `app/build.gradle.kts` or use network configuration.

## Network Security

The app allows cleartext traffic only to local networks (10.x, 192.168.x, 127.x) required for local SMB/FTP/NFS protocols. See `network_security_config.xml`.

## Related Documentation

- [Android Architecture](/docs/architecture/ANDROID_ARCHITECTURE.md)
- [Android Guide](/docs/guides/ANDROID_GUIDE.md)
