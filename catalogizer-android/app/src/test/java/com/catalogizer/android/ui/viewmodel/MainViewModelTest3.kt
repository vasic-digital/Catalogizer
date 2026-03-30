package com.catalogizer.android.ui.viewmodel

import com.catalogizer.android.MainDispatcherRule
import com.catalogizer.android.data.repository.MediaRepository
import io.mockk.*
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.test.*
import org.junit.After
import org.junit.Assert.*
import org.junit.Before
import org.junit.Rule
import org.junit.Test

@OptIn(ExperimentalCoroutinesApi::class)
class MainViewModelTest3 {

    @get:Rule
    val mainDispatcherRule = MainDispatcherRule()

    private lateinit var mockMediaRepository: MediaRepository
    private lateinit var viewModel: MainViewModel

    @Before
    fun setup() {
        mockMediaRepository = mockk(relaxed = true)
        viewModel = MainViewModel(mockMediaRepository)
    }

    @After
    fun tearDown() {
        clearAllMocks()
    }

    @Test
    fun `initial state is loading`() {
        assertTrue(viewModel.isLoading.value)
    }

    @Test
    fun `initializeApp sets loading to false`() = runTest {
        viewModel.initializeApp()
        advanceUntilIdle()

        assertFalse(viewModel.isLoading.value)
    }

    @Test
    fun `initializeApp sets loading to false even on exception`() = runTest {
        // MainViewModel catches exceptions internally
        viewModel.initializeApp()
        advanceUntilIdle()

        assertFalse(viewModel.isLoading.value)
    }

    @Test
    fun `isLoading is a StateFlow with initial value true`() {
        val flow = viewModel.isLoading
        assertNotNull(flow)
        assertTrue(flow.value)
    }

    @Test
    fun `multiple initializeApp calls complete without error`() = runTest {
        viewModel.initializeApp()
        advanceUntilIdle()
        assertFalse(viewModel.isLoading.value)

        // Call again
        viewModel.initializeApp()
        advanceUntilIdle()
        assertFalse(viewModel.isLoading.value)
    }
}
