package com.catalogizer.androidtv.data.tv

import android.content.Context
import androidx.arch.core.executor.testing.InstantTaskExecutorRule
import com.catalogizer.androidtv.MainDispatcherRule
import com.catalogizer.androidtv.data.models.AuthState
import com.catalogizer.androidtv.data.repository.AuthRepository
import com.catalogizer.androidtv.data.repository.SettingsRepository
import com.catalogizer.androidtv.ui.viewmodel.AuthViewModel
import io.mockk.coEvery
import io.mockk.coVerify
import io.mockk.every
import io.mockk.mockk
import io.mockk.mockkObject
import io.mockk.unmockkObject
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.test.advanceUntilIdle
import kotlinx.coroutines.test.runTest
import org.junit.After
import org.junit.Before
import org.junit.Rule
import org.junit.Test

@OptIn(ExperimentalCoroutinesApi::class)
class LogoutCleanupTest {

    @get:Rule
    val instantTaskExecutorRule = InstantTaskExecutorRule()

    @get:Rule
    val mainDispatcherRule = MainDispatcherRule()

    private lateinit var authRepository: AuthRepository
    private lateinit var tvChannelRepository: TvChannelRepository
    private lateinit var watchNextManager: WatchNextManager
    private lateinit var settingsRepository: SettingsRepository
    private lateinit var context: Context
    private val mockAuthState = MutableStateFlow(AuthState.Unauthenticated)

    @Before
    fun setup() {
        authRepository = mockk()
        tvChannelRepository = mockk(relaxed = true)
        watchNextManager = mockk(relaxed = true)
        settingsRepository = mockk(relaxed = true)
        context = mockk(relaxed = true)

        every { authRepository.authState } returns mockAuthState
        coEvery { authRepository.logout() } returns Unit

        // Mock TvChannelSyncWorker.cancel so it does not require real WorkManager
        mockkObject(TvChannelSyncWorker)
        every { TvChannelSyncWorker.cancel(any()) } returns Unit
    }

    @After
    fun tearDown() {
        unmockkObject(TvChannelSyncWorker)
    }

    private fun createViewModel(
        tvRepo: TvChannelRepository? = tvChannelRepository,
        watchNext: WatchNextManager? = watchNextManager,
        ctx: Context? = context,
        settings: SettingsRepository? = settingsRepository
    ) = AuthViewModel(authRepository, tvRepo, watchNext, ctx, settings)

    @Test
    fun `AuthViewModel logout calls deleteAllChannels`() = runTest {
        val viewModel = createViewModel()
        viewModel.logout()
        advanceUntilIdle()
        coVerify { tvChannelRepository.deleteAllChannels() }
    }

    @Test
    fun `AuthViewModel logout calls removeAll on WatchNextManager`() = runTest {
        val viewModel = createViewModel()
        viewModel.logout()
        advanceUntilIdle()
        coVerify { watchNextManager.removeAll() }
    }

    @Test
    fun `AuthViewModel logout calls cancel on TvChannelSyncWorker`() = runTest {
        val viewModel = createViewModel()
        viewModel.logout()
        advanceUntilIdle()
        // TvChannelSyncWorker.cancel(context) should have been invoked
        io.mockk.verify { TvChannelSyncWorker.cancel(context) }
    }

    @Test
    fun `AuthViewModel logout calls clearAllChannelIds`() = runTest {
        val viewModel = createViewModel()
        viewModel.logout()
        advanceUntilIdle()
        coVerify { settingsRepository.clearAllChannelIds() }
    }

    @Test
    fun `AuthViewModel logout calls clearAllLaunchActions`() = runTest {
        val viewModel = createViewModel()
        viewModel.logout()
        advanceUntilIdle()
        coVerify { settingsRepository.clearAllLaunchActions() }
    }

    @Test
    fun `AuthViewModel logout still calls authRepository logout even if cleanup fails`() = runTest {
        coEvery { tvChannelRepository.deleteAllChannels() } throws RuntimeException("Cleanup exploded")

        val viewModel = createViewModel()
        viewModel.logout()
        advanceUntilIdle()

        // The try/catch in AuthViewModel should swallow the cleanup error and still call logout()
        coVerify { authRepository.logout() }
    }

    @Test
    fun `AuthViewModel logout with null dependencies still logs out`() = runTest {
        val viewModel = createViewModel(
            tvRepo = null,
            watchNext = null,
            ctx = null,
            settings = null
        )
        viewModel.logout()
        advanceUntilIdle()
        coVerify { authRepository.logout() }
    }
}
