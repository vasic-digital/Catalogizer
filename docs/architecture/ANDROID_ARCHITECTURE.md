# Android Architecture Guide (catalogizer-android)

This guide documents the architecture, patterns, and conventions used in the `catalogizer-android` Kotlin/Compose application.

## Technology Stack

- **Kotlin** with coroutines and Flow
- **Jetpack Compose** for UI (Material 3)
- **Room** for local database
- **Retrofit + OkHttp** for API communication
- **Jetpack Navigation Compose** for screen navigation
- **Paging 3** for paginated lists
- **DataStore** for preferences
- **WorkManager** for background sync
- **MockK** for test mocking

## Project Structure

```
catalogizer-android/app/src/
├── main/java/com/catalogizer/android/
│   ├── CatalogizerApplication.kt       # Application class (WorkManager config)
│   ├── DependencyContainer.kt          # Manual DI container (singleton)
│   ├── CatalogizerWorkerFactory.kt     # WorkManager worker factory
│   ├── data/
│   │   ├── local/
│   │   │   ├── CatalogizerDatabase.kt  # Room database definition + TypeConverters
│   │   │   ├── MediaDao.kt             # Media data access object
│   │   │   ├── FavoriteDao.kt          # Favorites DAO
│   │   │   ├── WatchProgressDao.kt     # Watch progress DAO
│   │   │   └── SyncOperationDao.kt     # Offline sync queue DAO
│   │   ├── remote/
│   │   │   └── CatalogizerApi.kt       # Retrofit API interface + ApiResult + WebSocketEvent
│   │   ├── models/
│   │   │   ├── Auth.kt                 # Auth-related data classes
│   │   │   └── MediaItem.kt            # Media data models (Room entities)
│   │   ├── repository/
│   │   │   ├── AuthRepository.kt       # Auth operations (API + DataStore)
│   │   │   ├── MediaRepository.kt      # Media operations (API + Room cache)
│   │   │   └── OfflineRepository.kt    # Offline-first data access
│   │   └── sync/
│   │       ├── SyncManager.kt          # Orchestrates background sync
│   │       ├── SyncOperation.kt        # Sync operation models
│   │       └── SyncWorker.kt           # WorkManager worker for sync
│   └── ui/
│       ├── MainActivity.kt             # Single activity (Compose host)
│       ├── navigation/
│       │   └── CatalogizerNavigation.kt  # NavHost + screen definitions
│       ├── viewmodel/
│       │   ├── AuthViewModel.kt        # Authentication state management
│       │   ├── MainViewModel.kt        # App-level state
│       │   ├── HomeViewModel.kt        # Home screen data loading
│       │   └── SearchViewModel.kt      # Search state + results
│       ├── screens/
│       │   ├── home/
│       │   │   └── HomeScreen.kt       # Home screen (recent, favorites)
│       │   ├── login/
│       │   │   └── LoginScreen.kt      # Login form
│       │   ├── search/
│       │   │   └── SearchScreen.kt     # Media search + filters
│       │   └── settings/
│       │       └── SettingsScreen.kt   # App settings
│       └── theme/
│           └── Theme.kt                # Material 3 theme definition
└── test/java/com/catalogizer/android/
    ├── MainDispatcherRule.kt            # Test rule for coroutine dispatcher
    ├── CatalogizerTestApplication.kt    # Test application class
    └── ui/viewmodel/
        ├── AuthViewModelTest.kt
        ├── MainViewModelTest.kt
        ├── HomeViewModelTest.kt
        └── SearchViewModelTest.kt
```

## MVVM Pattern

The app follows the Model-View-ViewModel pattern with unidirectional data flow:

```
Compose UI (Screen)
    │ observes StateFlow
    ▼
ViewModel (state + actions)
    │ calls suspend functions
    ▼
Repository (business logic + caching)
    │
    ├──► Room DAO (local cache)
    └──► Retrofit API (remote data)
```

### ViewModel Pattern

ViewModels expose state as `StateFlow` and actions as regular functions:

```kotlin
// From ui/viewmodel/HomeViewModel.kt
class HomeViewModel(
    private val mediaRepository: MediaRepository
) : ViewModel() {

    private val _recentMedia = MutableStateFlow<List<MediaItem>>(emptyList())
    val recentMedia: StateFlow<List<MediaItem>> = _recentMedia.asStateFlow()

    private val _isLoading = MutableStateFlow(true)
    val isLoading: StateFlow<Boolean> = _isLoading.asStateFlow()

    private val _error = MutableStateFlow<String?>(null)
    val error: StateFlow<String?> = _error.asStateFlow()

    fun loadHomeData() {
        viewModelScope.launch {
            _isLoading.value = true
            _error.value = null
            try {
                val recentResult = mediaRepository.getRecentMedia(20)
                if (recentResult.isSuccess) {
                    _recentMedia.value = recentResult.data ?: emptyList()
                }
            } catch (e: Exception) {
                _error.value = e.message ?: "Failed to load media"
            } finally {
                _isLoading.value = false
            }
        }
    }
}
```

### Compose UI observes ViewModel state

```kotlin
// From ui/screens/home/HomeScreen.kt
@Composable
fun HomeScreen(
    viewModel: HomeViewModel,
    onNavigateToSearch: () -> Unit,
    onNavigateToSettings: () -> Unit,
    onNavigateToMediaDetail: (Long) -> Unit
) {
    val recentMedia by viewModel.recentMedia.collectAsStateWithLifecycle()
    val isLoading by viewModel.isLoading.collectAsStateWithLifecycle()
    val error by viewModel.error.collectAsStateWithLifecycle()

    LaunchedEffect(Unit) {
        viewModel.loadHomeData()
    }

    // Render UI based on state...
}
```

Key conventions:
- Use `collectAsStateWithLifecycle()` to observe StateFlow in Compose (lifecycle-aware)
- Use `LaunchedEffect` for one-time data loading triggered by composition
- ViewModel actions are plain functions (not suspend), they launch coroutines internally

## Dependency Injection

The project uses a **manual DI container** (not Hilt) implemented as a singleton:

```kotlin
// DependencyContainer.kt
class DependencyContainer(private val context: Context) {

    // Database (lazy initialization)
    private val database: CatalogizerDatabase by lazy {
        Room.databaseBuilder(
            context.applicationContext,
            CatalogizerDatabase::class.java,
            "catalogizer_database"
        )
            .addMigrations(*CatalogizerDatabase.ALL_MIGRATIONS)
            .fallbackToDestructiveMigration()
            .build()
    }

    // API (lazy initialization with OkHttp logging)
    private val api: CatalogizerApi by lazy {
        val logging = HttpLoggingInterceptor().apply {
            level = HttpLoggingInterceptor.Level.BODY
        }
        val client = OkHttpClient.Builder()
            .addInterceptor(logging)
            .connectTimeout(30, TimeUnit.SECONDS)
            .readTimeout(30, TimeUnit.SECONDS)
            .writeTimeout(30, TimeUnit.SECONDS)
            .build()

        Retrofit.Builder()
            .baseUrl(BuildConfig.API_BASE_URL)
            .client(client)
            .addConverterFactory(GsonConverterFactory.create())
            .build()
            .create(CatalogizerApi::class.java)
    }

    // Repositories (lazy, shared instances)
    val authRepository: AuthRepository by lazy { AuthRepository(api, dataStore) }
    val mediaRepository: MediaRepository by lazy { MediaRepository(api, database.mediaDao()) }

    // ViewModels (factory methods - new instance each call)
    fun createAuthViewModel(): AuthViewModel = AuthViewModel(authRepository)
    fun createHomeViewModel(): HomeViewModel = HomeViewModel(mediaRepository)
    fun createSearchViewModel(): SearchViewModel = SearchViewModel(mediaRepository)

    companion object {
        @Volatile
        private var instance: DependencyContainer? = null

        fun getInstance(context: Context): DependencyContainer {
            return instance ?: synchronized(this) {
                instance ?: DependencyContainer(context.applicationContext).also { instance = it }
            }
        }
    }
}
```

Access from Application class:

```kotlin
class CatalogizerApplication : Application(), Configuration.Provider {
    val dependencyContainer by lazy { DependencyContainer.getInstance(this) }

    override val workManagerConfiguration: Configuration
        get() = Configuration.Builder()
            .setWorkerFactory(CatalogizerWorkerFactory(dependencyContainer))
            .build()
}
```

### Adding a new dependency

1. Add the dependency as a `lazy` property in `DependencyContainer`
2. For ViewModels, add a factory method (`fun createXxxViewModel()`)
3. Wire it in the screen or activity where it is needed

## Room Database

### Database definition

```kotlin
// data/local/CatalogizerDatabase.kt
@Database(
    entities = [
        MediaItem::class,
        SearchHistory::class,
        DownloadItem::class,
        SyncOperation::class,
        WatchProgress::class,
        Favorite::class
    ],
    version = 1,
    exportSchema = false
)
@TypeConverters(Converters::class)
abstract class CatalogizerDatabase : RoomDatabase() {
    abstract fun mediaDao(): MediaDao
    abstract fun searchHistoryDao(): SearchHistoryDao
    abstract fun downloadDao(): DownloadDao
    abstract fun syncOperationDao(): SyncOperationDao
    abstract fun watchProgressDao(): WatchProgressDao
    abstract fun favoriteDao(): FavoriteDao

    companion object {
        val ALL_MIGRATIONS: Array<Migration> = arrayOf()
    }
}
```

### Type converters

Complex types (lists, maps, enums) are serialized to JSON strings for SQLite storage:

```kotlin
class Converters {
    private val json = Json { ignoreUnknownKeys = true }

    @TypeConverter
    fun fromStringList(value: List<String>?): String? =
        value?.let { json.encodeToString(it) }

    @TypeConverter
    fun toStringList(value: String?): List<String>? =
        value?.let { json.decodeFromString(it) }

    @TypeConverter
    fun fromSyncOperationType(value: SyncOperationType?): String? = value?.name

    @TypeConverter
    fun toSyncOperationType(value: String?): SyncOperationType? =
        value?.let { SyncOperationType.valueOf(it) }
}
```

### DAO pattern

DAOs return `Flow` for observable queries and use `suspend` for one-shot operations:

```kotlin
@Dao
interface MediaDao {
    @Query("SELECT * FROM media_items ORDER BY created_at DESC LIMIT :limit")
    fun getRecentlyAdded(limit: Int): Flow<List<MediaItem>>

    @Query("SELECT * FROM media_items WHERE id = :id")
    suspend fun getMediaById(id: Long): MediaItem?

    @Insert(onConflict = OnConflictStrategy.REPLACE)
    suspend fun insertMedia(media: MediaItem)

    @Query("UPDATE media_items SET is_favorite = :isFavorite WHERE id = :mediaId")
    suspend fun updateFavoriteStatus(mediaId: Long, isFavorite: Boolean)
}
```

## Retrofit API Client

### API interface definition

```kotlin
// data/remote/CatalogizerApi.kt
interface CatalogizerApi {
    @POST("auth/login")
    suspend fun login(@Body loginRequest: LoginRequest): Response<LoginResponse>

    @GET("media/search")
    suspend fun searchMedia(
        @Query("query") query: String? = null,
        @Query("media_type") mediaType: String? = null,
        @Query("limit") limit: Int = 20,
        @Query("offset") offset: Int = 0
    ): Response<MediaSearchResponse>

    @GET("media/{id}")
    suspend fun getMediaById(@Path("id") id: Long): Response<MediaItem>

    @PUT("media/{id}/progress")
    suspend fun updateWatchProgress(
        @Path("id") id: Long,
        @Body progressData: Map<String, Any>
    ): Response<Unit>
}
```

### ApiResult wrapper

All API responses are wrapped in a result type for consistent error handling:

```kotlin
data class ApiResult<T>(
    val data: T? = null,
    val error: String? = null,
    val isSuccess: Boolean = data != null && error == null
) {
    companion object {
        fun <T> success(data: T): ApiResult<T> = ApiResult(data = data)
        fun <T> error(error: String): ApiResult<T> = ApiResult(error = error)
    }
}

// Extension function for easy conversion
suspend fun <T> Response<T>.toApiResult(): ApiResult<T> {
    return try {
        if (isSuccessful) {
            body()?.let { ApiResult.success(it) } ?: ApiResult.error("Empty response")
        } else {
            ApiResult.error(errorBody()?.string() ?: "Unknown error (${code()})")
        }
    } catch (e: Exception) {
        ApiResult.error(e.message ?: "Network error")
    }
}
```

### WebSocket events (sealed class)

```kotlin
sealed class WebSocketEvent {
    data class MediaUpdate(val action: String, val mediaId: Long, val media: MediaItem?) : WebSocketEvent()
    data class SystemUpdate(val action: String, val component: String, val status: String) : WebSocketEvent()
    data class AnalysisComplete(val analysisId: String, val itemsProcessed: Int) : WebSocketEvent()
    data class Notification(val type: String, val title: String, val message: String) : WebSocketEvent()
    object Connected : WebSocketEvent()
    object Disconnected : WebSocketEvent()
    data class Error(val message: String) : WebSocketEvent()
}
```

## StateFlow / Compose Patterns

### Standard ViewModel state pattern

```kotlin
class SearchViewModel(private val mediaRepository: MediaRepository) : ViewModel() {
    // Private mutable state
    private val _searchQuery = MutableStateFlow("")
    private val _searchResults = MutableStateFlow<List<MediaItem>>(emptyList())
    private val _isLoading = MutableStateFlow(false)

    // Public read-only state
    val searchQuery: StateFlow<String> = _searchQuery.asStateFlow()
    val searchResults: StateFlow<List<MediaItem>> = _searchResults.asStateFlow()
    val isLoading: StateFlow<Boolean> = _isLoading.asStateFlow()

    fun search(query: String) {
        _searchQuery.value = query
        viewModelScope.launch {
            _isLoading.value = true
            val result = mediaRepository.searchMedia(query)
            if (result.isSuccess) {
                _searchResults.value = result.data ?: emptyList()
            }
            _isLoading.value = false
        }
    }
}
```

### Compose observation

```kotlin
@Composable
fun SearchScreen(viewModel: SearchViewModel) {
    val query by viewModel.searchQuery.collectAsStateWithLifecycle()
    val results by viewModel.searchResults.collectAsStateWithLifecycle()
    val isLoading by viewModel.isLoading.collectAsStateWithLifecycle()

    Column {
        TextField(value = query, onValueChange = { viewModel.search(it) })
        if (isLoading) CircularProgressIndicator()
        LazyColumn {
            items(results) { item -> MediaCard(item) }
        }
    }
}
```

## Repository Layer: Offline-First Pattern

Repositories implement an offline-first strategy:

```kotlin
// data/repository/MediaRepository.kt
suspend fun getRecentMedia(limit: Int = 10): ApiResult<List<MediaItem>> {
    return try {
        // Try remote first
        val result = api.getRecentMedia(limit).toApiResult()
        if (result.isSuccess && result.data != null) {
            // Cache to local DB on success
            mediaDao.insertAllMedia(result.data)
        }
        result
    } catch (e: Exception) {
        // Fallback to local cache on network failure
        val localData = mediaDao.getRecentlyAdded(limit).first()
        ApiResult.success(localData)
    }
}
```

Optimistic updates for user interactions:

```kotlin
suspend fun toggleFavorite(mediaId: Long): ApiResult<Unit> {
    val currentMedia = mediaDao.getMediaById(mediaId)
    val newStatus = !(currentMedia?.isFavorite ?: false)

    // Update local state immediately (optimistic)
    mediaDao.updateFavoriteStatus(mediaId, newStatus)

    // Sync with server
    val result = if (newStatus) api.addToFavorites(mediaId).toApiResult()
                 else api.removeFromFavorites(mediaId).toApiResult()

    // Revert if server sync failed
    if (!result.isSuccess) {
        mediaDao.updateFavoriteStatus(mediaId, !newStatus)
    }
    return result
}
```

## Paging 3 Integration

For large lists, the app uses Paging 3 with both remote and local sources:

```kotlin
// Remote paging
fun getMediaPaging(searchRequest: MediaSearchRequest): Flow<PagingData<MediaItem>> {
    return Pager(
        config = PagingConfig(
            pageSize = 20,
            enablePlaceholders = false,
            prefetchDistance = 5
        ),
        pagingSourceFactory = { MediaPagingSource(api, searchRequest) }
    ).flow
}

// Local paging (from Room)
fun getMediaByTypePaging(mediaType: String): Flow<PagingData<MediaItem>> {
    return Pager(
        config = PagingConfig(pageSize = 20),
        pagingSourceFactory = { mediaDao.getMediaByTypePaging(mediaType) }
    ).flow
}
```

## Navigation

Navigation uses Jetpack Navigation Compose with a sealed class for type-safe routes:

```kotlin
sealed class Screen(val route: String) {
    object Login : Screen("login")
    object Home : Screen("home")
    object Search : Screen("search")
    object Settings : Screen("settings")
}

@Composable
fun CatalogizerNavigation(
    isAuthenticated: Boolean,
    authViewModel: AuthViewModel,
    homeViewModel: HomeViewModel,
    searchViewModel: SearchViewModel,
    navController: NavHostController = rememberNavController()
) {
    val startDestination = if (isAuthenticated) Screen.Home.route else Screen.Login.route

    NavHost(navController = navController, startDestination = startDestination) {
        composable(Screen.Login.route) {
            LoginScreen(
                authViewModel = authViewModel,
                onLoginSuccess = {
                    navController.navigate(Screen.Home.route) {
                        popUpTo(Screen.Login.route) { inclusive = true }
                    }
                }
            )
        }
        composable(Screen.Home.route) {
            HomeScreen(viewModel = homeViewModel, ...)
        }
        // ... more screens
    }
}
```

## Background Sync (WorkManager)

The `SyncManager` orchestrates background data synchronization using WorkManager:

- `SyncWorker` runs periodic sync tasks
- `CatalogizerWorkerFactory` injects dependencies into workers
- `SyncOperation` entities track pending offline operations

## Testing

### Test setup

```kotlin
@ExperimentalCoroutinesApi
class AuthViewModelTest {
    @get:Rule val instantExecutorRule = InstantTaskExecutorRule()
    @get:Rule val mainDispatcherRule = MainDispatcherRule()

    private lateinit var mockAuthRepository: AuthRepository
    private lateinit var viewModel: AuthViewModel

    @Before
    fun setup() {
        Dispatchers.setMain(StandardTestDispatcher())
        mockAuthRepository = mockk(relaxed = true)
        every { mockAuthRepository.isAuthenticated } returns flowOf(true)
        viewModel = AuthViewModel(mockAuthRepository)
    }

    @After
    fun tearDown() {
        Dispatchers.resetMain()
        clearAllMocks()
    }
}
```

### Test patterns

- **MockK** for mocking dependencies (`mockk(relaxed = true)`)
- **MainDispatcherRule** replaces `Dispatchers.Main` for testing coroutines
- **InstantTaskExecutorRule** for synchronous LiveData execution
- **`runTest`** + **`advanceUntilIdle()`** for testing coroutine-based ViewModels

```kotlin
@Test
fun `login should update auth state`() = runTest {
    coEvery { mockAuthRepository.login(any(), any()) } returns ApiResult.success(loginResponse)
    viewModel.login("testuser", "password")
    advanceUntilIdle()
    // Assert state changes
}
```

Run tests: `cd catalogizer-android && ./gradlew test`
