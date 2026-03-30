# catalogizer-androidtv Testing

## Unit Tests

### Framework

- **JUnit 4** as the test runner
- **MockK** and **Mockito** for mocking
- **kotlinx-coroutines-test** for coroutine testing (`runTest`, `UnconfinedTestDispatcher`)
- **Robolectric** for Android framework mocking
- `testOptions.unitTests.isReturnDefaultValues = true` for convenience stubs

### Conventions

- Test files: `app/src/test/java/com/catalogizer/androidtv/`
- Naming: `<ClassName>Test.kt` with supplementary `<ClassName>Test2.kt` files for extended coverage
- `MainDispatcherRule` replaces `Dispatchers.Main` with `UnconfinedTestDispatcher`
- `CatalogizerTVTestApplication` provides a lightweight Application for Robolectric tests

### Test Structure

| Directory | Contents |
|-----------|----------|
| `data/local/` | Room type converters |
| `data/models/` | AuthState, MediaItem, Settings serialization |
| `data/remote/` | AuthInterceptor, LoginResponse parsing |
| `data/repository/` | Auth, Media, Settings repository tests |
| `data/sync/` | SyncService tests |
| `data/tv/` | TvProvider contract tests |
| `ui/viewmodel/` | Auth, Home, Main, Settings, Search ViewModel tests |
| `ui/screens/` | Home, Login, Search, Settings screen tests |
| `ui/navigation/` | TV screen route definitions |
| `utils/` | FormatUtils tests |

### Running

```bash
./gradlew test                   # all unit tests
./gradlew testDebugUnitTest      # debug variant only
./gradlew jacocoTestReport       # coverage (HTML + XML)
```

## TV-Specific Testing Considerations

### Focus and Navigation

TV tests must validate that:
- All interactive elements are focusable
- D-pad navigation reaches every control in the expected order
- Focus indicators are visible and correctly styled
- Back button navigates to the expected screen

### HelixQA Autonomous Testing

The Android TV app is a primary target for HelixQA LLM-driven QA sessions. HelixQA connects to physical devices (e.g., Mi Box 4 at `192.168.0.134:5555`) and uses vision models to autonomously navigate screens, validate UI state, and discover bugs.

Key HelixQA constraints for TV testing:
- `adb reverse tcp:8080 tcp:8080` must be set up before each session
- Text input requires `dpad_center` before `type`, `KEYCODE_TAB` between fields
- All navigation is D-pad based (no tap coordinates)

## Instrumented Tests

Instrumented tests run on a physical TV device or emulator via:

```bash
./gradlew connectedAndroidTest
```

Currently focused on unit tests; no instrumented test files exist yet.

## JDK Constraints

The project targets JDK 17. Both `gradle.properties` and the `kapt` block in `build.gradle.kts` include `--add-opens` flags for Room annotation processing compatibility with the host JDK 21.

## Mocking Patterns

- **Repositories**: `mockk<MediaRepository>()` with `coEvery { }` for suspend functions
- **API**: Mock `CatalogizerApi` interface directly
- **Settings**: Mock `SettingsRepository` with fake DataStore values
- **Dispatchers**: `MainDispatcherRule` in every ViewModel test
