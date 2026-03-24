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
import retrofit2.Retrofit
import retrofit2.converter.gson.GsonConverterFactory
import java.util.concurrent.TimeUnit

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

    // Current active base URL — read from settings or default
    private var currentBaseUrl: String = BuildConfig.API_BASE_URL

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
        return Retrofit.Builder()
            .baseUrl(baseUrl.trimEnd('/') + "/")
            .client(buildOkHttpClient())
            .addConverterFactory(GsonConverterFactory.create())
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

    val mediaRepository: MediaRepository by lazy {
        MediaRepository(context, api)
    }

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
     */
    suspend fun initializeAsync() {
        // Load server URL from persisted settings
        try {
            val settings = settingsRepository.getSettingsAsync()
            if (settings.serverUrl.isNotBlank() && settings.serverUrl != Settings.DEFAULT_SERVER_URL) {
                currentBaseUrl = settings.serverUrl
            }
            // Ensure DataStore file exists by writing defaults on first launch
            if (settings.serverUrl == Settings.DEFAULT_SERVER_URL) {
                settingsRepository.saveSettings(settings)
            }
        } catch (_: Exception) {
            // Use default/BuildConfig URL on first launch
        }
        // Trigger API creation with the loaded URL
        _api = null // Force recreation with correct URL
        api
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
