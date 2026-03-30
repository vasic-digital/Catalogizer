# catalogizer-android Testing

## Unit Tests

### Framework

- **JUnit 4** as the test runner
- **MockK** and **Mockito** for mocking
- **kotlinx-coroutines-test** for coroutine testing (`runTest`, `UnconfinedTestDispatcher`)
- **Robolectric** for Android framework classes without a device
- **Turbine** for Flow testing (where used)

### Conventions

- Test files: `app/src/test/java/com/catalogizer/android/`
- Naming: `<ClassName>Test.kt` (e.g., `AuthViewModelTest.kt`)
- Some classes have supplementary test files: `<ClassName>Test2.kt` for additional coverage
- `MainDispatcherRule` -- JUnit rule that replaces `Dispatchers.Main` with `UnconfinedTestDispatcher`
- `TestDataGenerator` -- factory for test fixtures (media items, auth tokens)
- `ViewModelTestBase` -- base class for ViewModel tests with common setup
- `MockRepositoryHelper` -- pre-configured mock repositories

### Test Structure

| Directory | Contents |
|-----------|----------|
| `data/local/` | DAO tests, entity serialization, Room converters |
| `data/models/` | Model serialization, enum coverage |
| `data/remote/` | API result handling, Retrofit serialization, WebSocket events |
| `data/repository/` | Repository tests with mocked API and DAO |
| `data/sync/` | SyncManager, SyncWorker, SyncOperation tests |
| `ui/viewmodel/` | ViewModel tests with mocked repositories |
| `ui/screens/` | Compose screen snapshot/behavior tests |
| `ui/navigation/` | Route definition tests |

### Running

```bash
./gradlew test                   # all unit tests
./gradlew testDebugUnitTest      # debug variant only
./gradlew jacocoTestReport       # coverage (HTML + XML)
```

## Instrumented Tests

### Framework

- **Espresso** for view interactions
- **Compose UI Testing** (`ui-test-junit4`) for Compose screens
- Run on a physical device or emulator

### Test Files

Located in `app/src/androidTest/java/com/catalogizer/android/`:

| Test | Target |
|------|--------|
| `DownloadDaoTest` | Download DAO operations |
| `FavoriteDaoTest` | Favorite CRUD via Room |
| `MediaDaoTest` | Media item persistence |
| `SearchHistoryDaoTest` | Search history DAO |
| `WatchProgressDaoTest` | Watch progress tracking |
| `ComposeTestRule` | Shared Compose test rule utility |

### Running

```bash
./gradlew connectedAndroidTest           # all instrumented tests
./gradlew connectedDebugAndroidTest      # debug variant only
```

## JDK Constraints

The project targets JDK 21 (`jvmTarget = "21"`). Room's kapt annotation processor requires `--add-opens` JVM flags in `gradle.properties` to work with JDK 21. These flags open `jdk.compiler` internal modules to ALL-UNNAMED.

## Mocking Patterns

- **Repository layer**: MockK `mockk<MediaRepository>()` with `coEvery { }` for suspend functions
- **API responses**: Mock `CatalogizerApi` interface methods directly
- **Database**: In-memory Room database via `Room.inMemoryDatabaseBuilder()` for DAO tests
- **Dispatchers**: Replace `Dispatchers.Main` via `MainDispatcherRule` in every ViewModel test
