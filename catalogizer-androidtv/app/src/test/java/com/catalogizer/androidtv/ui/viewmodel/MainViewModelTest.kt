package com.catalogizer.androidtv.ui.viewmodel

import androidx.arch.core.executor.testing.InstantTaskExecutorRule
import com.catalogizer.androidtv.MainDispatcherRule
import com.catalogizer.androidtv.data.repository.AuthRepository
import io.mockk.mockk
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.launch
import kotlinx.coroutines.test.advanceUntilIdle
import kotlinx.coroutines.test.runTest
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Before
import org.junit.Rule
import org.junit.Test

@ExperimentalCoroutinesApi
class MainViewModelTest {

    @get:Rule
    val instantTaskExecutorRule = InstantTaskExecutorRule()

    @get:Rule
    val mainDispatcherRule = MainDispatcherRule()

    private lateinit var authRepository: AuthRepository
    private lateinit var viewModel: MainViewModel

    @Before
    fun setup() {
        authRepository = mockk()
        viewModel = MainViewModel(authRepository)
    }

    @Test
    fun `initial loading state should be true`() = runTest {
        val initialState = viewModel.isLoading.first()
        assertTrue(initialState)
    }

    @Test
    fun `initializeApp should set loading to false`() = runTest {
        // Initial state should be loading
        var loadingState = viewModel.isLoading.first()
        assertTrue(loadingState)

        // Call initializeApp
        viewModel.initializeApp()
        advanceUntilIdle()

        // Loading should now be false
        loadingState = viewModel.isLoading.first()
        assertFalse(loadingState)
    }

    @Test
    fun `isLoading state flow should emit correct values`() = runTest {
        // Collect all emissions
        val emissions = mutableListOf<Boolean>()

        val job = launch {
            viewModel.isLoading.collect { emissions.add(it) }
        }

        // Initial emission should be true
        assertTrue(emissions.first())

        // Call initializeApp
        viewModel.initializeApp()
        advanceUntilIdle()

        // Should have emitted false
        assertFalse(emissions.last())

        job.cancel()
    }

    @Test
    fun `multiple initializeApp calls should work correctly`() = runTest {
        // First call
        viewModel.initializeApp()
        advanceUntilIdle()

        var loadingState = viewModel.isLoading.first()
        assertFalse(loadingState)

        // Second call - should remain false
        viewModel.initializeApp()
        advanceUntilIdle()

        loadingState = viewModel.isLoading.first()
        assertFalse(loadingState)
    }

    @Test
    fun `initializeApp completes successfully without exceptions`() = runTest {
        // Should not throw any exceptions
        viewModel.initializeApp()
        advanceUntilIdle()

        // If we reach here, the test passes
        assertTrue(true)
    }
}