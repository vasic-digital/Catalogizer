package com.catalogizer.androidtv.ui.viewmodel

import com.catalogizer.androidtv.MainDispatcherRule
import com.catalogizer.androidtv.data.repository.AuthRepository
import io.mockk.*
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.test.advanceUntilIdle
import kotlinx.coroutines.test.runTest
import org.junit.After
import org.junit.Assert.*
import org.junit.Before
import org.junit.Rule
import org.junit.Test

@OptIn(ExperimentalCoroutinesApi::class)
class MainViewModelTest2 {

    @get:Rule
    val mainDispatcherRule = MainDispatcherRule()

    private val mockAuthRepository = mockk<AuthRepository>(relaxed = true)
    private lateinit var viewModel: MainViewModel

    @Before
    fun setup() {
        viewModel = MainViewModel(mockAuthRepository)
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
    fun `isLoading is StateFlow with initial true value`() {
        val flow = viewModel.isLoading
        assertNotNull(flow)
        assertTrue(flow.value)
    }

    @Test
    fun `multiple initializeApp calls are safe`() = runTest {
        viewModel.initializeApp()
        viewModel.initializeApp()
        advanceUntilIdle()

        assertFalse(viewModel.isLoading.value)
    }

    @Test
    fun `initializeApp completes successfully`() = runTest {
        viewModel.initializeApp()
        advanceUntilIdle()

        // After initialization, app should no longer be loading
        assertFalse(viewModel.isLoading.value)
    }
}
