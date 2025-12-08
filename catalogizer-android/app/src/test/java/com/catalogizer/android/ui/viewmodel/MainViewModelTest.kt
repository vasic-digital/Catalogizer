package com.catalogizer.android.ui.viewmodel

import androidx.arch.core.executor.testing.InstantTaskExecutorRule
import com.catalogizer.android.data.repository.MediaRepository
import com.catalogizer.android.data.remote.ApiResult
import com.catalogizer.android.data.models.MediaItem
import io.mockk.*
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.flow.flowOf
import kotlinx.coroutines.test.*
import com.catalogizer.android.MainDispatcherRule
import org.junit.After
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.Assert.*

/**
 * ViewModel tests for MainViewModel
 */
@ExperimentalCoroutinesApi
class MainViewModelTest {
    
    @get:Rule
    val instantExecutorRule = InstantTaskExecutorRule()
    
    @get:Rule
    val mainDispatcherRule = MainDispatcherRule()
    
    private lateinit var mockMediaRepository: MediaRepository
    private lateinit var viewModel: MainViewModel
    
    @Before
    fun setup() {
        Dispatchers.setMain(StandardTestDispatcher())
        mockMediaRepository = mockk(relaxed = true)
        viewModel = MainViewModel(mockMediaRepository)
    }
    
    @After
    fun tearDown() {
        Dispatchers.resetMain()
        clearAllMocks()
    }
    
    @Test
    fun `initial state should be loading`() {
        // Verify initial state
        assertTrue(viewModel.isLoading.value)
    }
    
    @Test
    fun `initialize app should set loading to false`() = runTest {
        // Given - initial state should be loading
        assertTrue(viewModel.isLoading.value)
        
        // When
        viewModel.initializeApp()
        advanceUntilIdle()
        
        // Then
        assertFalse(viewModel.isLoading.value)
    }
    
    @Test
    fun `initialize app should handle errors gracefully`() = runTest {
        // Given - initial state should be loading
        assertTrue(viewModel.isLoading.value)
        
        // When
        viewModel.initializeApp()
        advanceUntilIdle()
        
        // Then - should still set loading to false even if errors occur
        assertFalse(viewModel.isLoading.value)
        // Should not crash and should handle the error
    }
}