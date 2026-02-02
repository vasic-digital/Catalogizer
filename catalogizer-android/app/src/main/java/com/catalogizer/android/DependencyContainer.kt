package com.catalogizer.android

import android.content.Context
import androidx.datastore.core.DataStore
import androidx.datastore.preferences.core.PreferenceDataStoreFactory
import androidx.datastore.preferences.core.Preferences
import androidx.datastore.preferences.preferencesDataStoreFile
import androidx.room.Room
import com.catalogizer.android.data.local.CatalogizerDatabase
import com.catalogizer.android.data.remote.CatalogizerApi
import com.catalogizer.android.data.repository.AuthRepository
import com.catalogizer.android.data.repository.MediaRepository
import com.catalogizer.android.data.sync.SyncManager
import com.catalogizer.android.ui.viewmodel.AuthViewModel
import com.catalogizer.android.ui.viewmodel.HomeViewModel
import com.catalogizer.android.ui.viewmodel.MainViewModel
import com.catalogizer.android.ui.viewmodel.SearchViewModel
import okhttp3.OkHttpClient
import okhttp3.logging.HttpLoggingInterceptor
import retrofit2.Retrofit
import retrofit2.converter.gson.GsonConverterFactory
import java.util.concurrent.TimeUnit

class DependencyContainer(private val context: Context) {

    // DataStore
    private val dataStore: DataStore<Preferences> by lazy {
        PreferenceDataStoreFactory.create {
            context.preferencesDataStoreFile("catalogizer_prefs")
        }
    }

    // Database
    private val database: CatalogizerDatabase by lazy {
        Room.databaseBuilder(
            context.applicationContext,
            CatalogizerDatabase::class.java,
            "catalogizer_database"
        )
            .build()
    }

    // API
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

    // Repositories
    val authRepository: AuthRepository by lazy {
        AuthRepository(api, dataStore)
    }

    val mediaRepository: MediaRepository by lazy {
        MediaRepository(api, database.mediaDao())
    }

    // Sync Manager
    val syncManager: SyncManager by lazy {
        SyncManager(database, api, authRepository, mediaRepository, context)
    }

    // ViewModels
    fun createAuthViewModel(): AuthViewModel {
        return AuthViewModel(authRepository)
    }

    fun createMainViewModel(): MainViewModel {
        return MainViewModel(mediaRepository)
    }

    fun createHomeViewModel(): HomeViewModel {
        return HomeViewModel(mediaRepository)
    }

    fun createSearchViewModel(): SearchViewModel {
        return SearchViewModel(mediaRepository)
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