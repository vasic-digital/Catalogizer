package com.catalogizer.androidtv

import android.content.Context
import androidx.datastore.core.DataStore
import androidx.datastore.preferences.core.PreferenceDataStoreFactory
import androidx.datastore.preferences.core.Preferences
import androidx.datastore.preferences.preferencesDataStoreFile
import com.catalogizer.androidtv.data.discovery.NetworkDiscoveryService
import com.catalogizer.androidtv.data.models.Settings
import com.catalogizer.androidtv.data.remote.CatalogizerApi
import com.catalogizer.androidtv.data.repository.AuthRepository
import com.catalogizer.androidtv.data.repository.MediaRepository
import com.catalogizer.androidtv.data.repository.SettingsRepository
import com.catalogizer.androidtv.ui.viewmodel.AuthViewModel
import com.catalogizer.androidtv.ui.viewmodel.HomeViewModel
import com.catalogizer.androidtv.ui.viewmodel.MainViewModel
import com.catalogizer.androidtv.ui.viewmodel.SettingsViewModel
import com.catalogizer.androidtv.ui.screens.search.SearchViewModel
import com.catalogizer.androidtv.data.remote.AuthInterceptor
import okhttp3.OkHttpClient
import okhttp3.logging.HttpLoggingInterceptor
import com.jakewharton.retrofit2.converter.kotlinx.serialization.asConverterFactory
import kotlinx.serialization.json.Json
import okhttp3.MediaType.Companion.toMediaType
import retrofit2.Retrofit
import java.util.concurrent.TimeUnit

/**
 * Manual dependency injection container for the Android TV app. Provides singleton
 * access to repositories, API clients, network discovery, and ViewModel factories.
 * Manages runtime server URL switching and async initialization from [DataStore].
 */
class DependencyContainer(private val context: Context) {

    private val dataStore: DataStore<Preferences> by lazy {
        PreferenceDataStoreFactory.create {
            context.preferencesDataStoreFile("catalogizer_tv_prefs")
        }
    }

    val settingsRepository: SettingsRepository by lazy {
        SettingsRepository(dataStore)
    }

    val discoveryService: NetworkDiscoveryService by lazy {
        NetworkDiscoveryService(context)
    }

    val authRepository: AuthRepository by lazy {
        AuthRepository(context, null)
    }

    // Current active base URL — read from settings or discovered at runtime
    private var currentBaseUrl: String = BuildConfig.API_BASE_URL.ifBlank { "" }

    /**
     * Build OkHttpClient with auth interceptor and logging.
     */
    private fun buildOkHttpClient(): OkHttpClient {
        val logging = HttpLoggingInterceptor().apply {
            level = if (BuildConfig.DEBUG) HttpLoggingInterceptor.Level.BODY
                    else HttpLoggingInterceptor.Level.NONE
        }
        val authInterceptor = AuthInterceptor(authRepository)

        return OkHttpClient.Builder()
            .addInterceptor(authInterceptor)
            .addInterceptor(logging)
            .connectTimeout(30, TimeUnit.SECONDS)
            .readTimeout(30, TimeUnit.SECONDS)
            .writeTimeout(30, TimeUnit.SECONDS)
            .build()
    }

    /**
     * Create a Retrofit API instance pointing to the given base URL.
     */
    private fun buildApi(baseUrl: String): CatalogizerApi {
        // Retrofit requires a non-empty base URL. Use a placeholder when unconfigured;
        // all calls will fail with a connection error, which is the expected behavior
        // until the user configures a real server URL.
        val effectiveUrl = baseUrl.ifBlank { "http://localhost:8080" }
        return Retrofit.Builder()
            .baseUrl(effectiveUrl.trimEnd('/') + "/")
            .client(buildOkHttpClient())
            .addConverterFactory(Json { ignoreUnknownKeys = true; coerceInputValues = true }.asConverterFactory("application/json".toMediaType()))
            .build()
            .create(CatalogizerApi::class.java)
    }

    private var _api: CatalogizerApi? = null

    val api: CatalogizerApi
        get() {
            if (_api == null) {
                _api = buildApi(currentBaseUrl)
                authRepository.setApi(_api!!)
            }
            return _api!!
        }

    val mediaRepository: MediaRepository
        get() = MediaRepository(context, api)

    // ViewModels
    fun createAuthViewModel(): AuthViewModel = AuthViewModel(authRepository)
    fun createMainViewModel(): MainViewModel = MainViewModel(authRepository)
    fun createHomeViewModel(): HomeViewModel = HomeViewModel(mediaRepository)
    fun createSettingsViewModel(): SettingsViewModel = SettingsViewModel(settingsRepository)
    fun createSearchViewModel(): SearchViewModel = SearchViewModel(mediaRepository)

    /**
     * Change the API base URL at runtime. Recreates the Retrofit client.
     * Call this when the user selects a discovered or manually entered server.
     */
    fun switchServer(newBaseUrl: String) {
        currentBaseUrl = newBaseUrl.trimEnd('/')
        _api = buildApi(currentBaseUrl)
        authRepository.setApi(_api!!)
    }

    /**
     * Get the currently configured API base URL.
     */
    fun getServerUrl(): String = currentBaseUrl

    /**
     * Initialize the container: load saved server URL, create API client.
     * Call from Application.onCreate().
     *
     * Priority:
     * 1. Persisted server URL from DataStore (previous user selection)
     * 2. Probe localhost:8080 (works when ADB reverse is active)
     * 3. Leave empty — LoginScreen will prompt for manual entry or discovery
     */
    suspend fun initializeAsync() {
        // Load server URL from persisted settings
        try {
            val settings = settingsRepository.getSettingsAsync()
            if (settings.serverUrl.isNotBlank()) {
                currentBaseUrl = settings.serverUrl
            } else {
                // No saved URL — try localhost (ADB reverse proxy)
                val localhostUrl = "http://localhost:8080"
                val probe = discoveryService.probeServer(localhostUrl)
                if (probe != null) {
                    // Use the resolved URL (real IP from /discovery), not localhost
                    val resolvedUrl = probe.url
                    currentBaseUrl = resolvedUrl
                    settingsRepository.updateServerUrl(resolvedUrl)
                    settingsRepository.addServer(probe)
                }
            }
        } catch (_: Exception) {
            // Leave empty — LoginScreen handles unconfigured state
        }
        // Trigger API creation with the loaded URL (if any)
        _api = null // Force recreation with correct URL
        if (currentBaseUrl.isNotBlank()) {
            api // Only create API if we have a URL
        }
    }

    /**
     * Synchronous initialize (uses BuildConfig URL, settings loaded later).
     */
    fun initialize() {
        api
    }

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
