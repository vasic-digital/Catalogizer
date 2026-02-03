package com.catalogizer.androidtv

import android.content.Context
import androidx.datastore.core.DataStore
import androidx.datastore.preferences.core.PreferenceDataStoreFactory
import androidx.datastore.preferences.core.Preferences
import androidx.datastore.preferences.preferencesDataStoreFile
import com.catalogizer.androidtv.data.remote.CatalogizerApi
import com.catalogizer.androidtv.data.repository.AuthRepository
import com.catalogizer.androidtv.data.repository.MediaRepository
import com.catalogizer.androidtv.data.repository.SettingsRepository
import com.catalogizer.androidtv.ui.viewmodel.AuthViewModel
import com.catalogizer.androidtv.ui.viewmodel.HomeViewModel
import com.catalogizer.androidtv.ui.viewmodel.MainViewModel
import com.catalogizer.androidtv.ui.viewmodel.SettingsViewModel
import com.catalogizer.androidtv.ui.screens.search.SearchViewModel
import androidx.lifecycle.ViewModelProvider
import androidx.lifecycle.ViewModelStoreOwner
import com.catalogizer.androidtv.data.remote.AuthInterceptor
import okhttp3.OkHttpClient
import okhttp3.logging.HttpLoggingInterceptor
import retrofit2.Retrofit
import retrofit2.converter.gson.GsonConverterFactory
import java.util.concurrent.TimeUnit

class DependencyContainer(private val context: Context) {

    // DataStore
    private val dataStore: DataStore<Preferences> by lazy {
        PreferenceDataStoreFactory.create {
            context.preferencesDataStoreFile("catalogizer_tv_prefs")
        }
    }

    // Initialization order: authRepository is created with null API, then api lazy
    // creates AuthInterceptor (reads authState synchronously) and calls setApi().
    // The API is guaranteed to be set before any HTTP call because the interceptor
    // is part of the OkHttp client inside api, so by the time it executes, setApi()
    // has already been called at the end of the api lazy block.
    val authRepository: AuthRepository by lazy {
        AuthRepository(context, null)
    }

    private val api: CatalogizerApi by lazy {
        val logging = HttpLoggingInterceptor().apply {
            level = HttpLoggingInterceptor.Level.BODY
        }

        val authInterceptor = AuthInterceptor(authRepository)

        val client = OkHttpClient.Builder()
            .addInterceptor(authInterceptor)
            .addInterceptor(logging)
            .connectTimeout(30, TimeUnit.SECONDS)
            .readTimeout(30, TimeUnit.SECONDS)
            .writeTimeout(30, TimeUnit.SECONDS)
            .build()

        val apiInstance = Retrofit.Builder()
            .baseUrl(BuildConfig.API_BASE_URL)
            .client(client)
            .addConverterFactory(GsonConverterFactory.create())
            .build()
            .create(CatalogizerApi::class.java)

        // Set the API on the repository after creation
        authRepository.setApi(apiInstance)

        apiInstance
    }

    val mediaRepository: MediaRepository by lazy {
        MediaRepository(context, api)
    }

    val settingsRepository: SettingsRepository by lazy {
        SettingsRepository(dataStore)
    }

    // ViewModels
    fun createAuthViewModel(): AuthViewModel {
        return AuthViewModel(authRepository)
    }

    fun createMainViewModel(): MainViewModel {
        return MainViewModel(authRepository)
    }

    fun createHomeViewModel(): HomeViewModel {
        return HomeViewModel(mediaRepository)
    }

    fun createSettingsViewModel(): SettingsViewModel {
        return SettingsViewModel(settingsRepository)
    }

    fun createSearchViewModel(): SearchViewModel {
        return SearchViewModel(mediaRepository)
    }

    // Eagerly initialize the API to resolve the circular dependency early.
    // Call from Application.onCreate() to fail fast if configuration is wrong.
    fun initialize() {
        api // triggers lazy initialization, which also calls authRepository.setApi()
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