package com.catalogizer.androidtv.ui.viewmodel

import android.content.Context
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.catalogizer.androidtv.data.models.AuthState
import com.catalogizer.androidtv.data.repository.AuthRepository
import com.catalogizer.androidtv.data.repository.SettingsRepository
import com.catalogizer.androidtv.data.tv.TvChannelRepository
import com.catalogizer.androidtv.data.tv.TvChannelSyncWorker
import com.catalogizer.androidtv.data.tv.WatchNextManager
import kotlinx.coroutines.flow.SharingStarted
import kotlinx.coroutines.flow.stateIn
import kotlinx.coroutines.launch

/**
 * Manages authentication state for the TV app including login, logout, token refresh,
 * and error clearing. On logout, performs full TV channel cleanup before signing out.
 */
class AuthViewModel(
    private val authRepository: AuthRepository,
    private val tvChannelRepository: TvChannelRepository? = null,
    private val watchNextManager: WatchNextManager? = null,
    private val context: Context? = null,
    private val settingsRepository: SettingsRepository? = null
) : ViewModel() {

    val authState = authRepository.authState.stateIn(
        viewModelScope,
        SharingStarted.WhileSubscribed(5000),
        AuthState.Unauthenticated
    )

    fun login(username: String, password: String) {
        viewModelScope.launch {
            try {
                authRepository.login(username, password)
            } catch (e: Exception) {
                // Handle login error - the repository will update the authState with error
            }
        }
    }

    fun logout() {
        viewModelScope.launch {
            // Clean up TV channels before logout
            try {
                tvChannelRepository?.deleteAllChannels()
                watchNextManager?.removeAll()
                context?.let { TvChannelSyncWorker.cancel(it) }
                settingsRepository?.clearAllChannelIds()
                settingsRepository?.clearAllLaunchActions()
            } catch (e: Exception) {
                android.util.Log.w("AuthVM", "Channel cleanup failed: ${e.message}")
            }

            authRepository.logout()
        }
    }

    fun refreshToken() {
        viewModelScope.launch {
            authRepository.refreshToken()
        }
    }

    fun clearError() {
        viewModelScope.launch {
            authRepository.clearError()
        }
    }
}
