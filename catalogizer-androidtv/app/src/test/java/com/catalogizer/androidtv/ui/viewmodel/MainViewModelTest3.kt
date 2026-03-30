package com.catalogizer.androidtv.ui.viewmodel

import com.catalogizer.androidtv.MainDispatcherRule
import com.catalogizer.androidtv.data.repository.AuthRepository
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

    private lateinit var mockAuthRepository: AuthRepository
    private lateinit var viewModel: MainViewModel

    @Before
    fun setup() {
        mockAuthRepository = mockk(relaxed = true)
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
    fun `multiple initializeApp calls are safe`() = runTest {
        viewModel.initializeApp()
        advanceUntilIdle()
        assertFalse(viewModel.isLoading.value)

        viewModel.initializeApp()
        advanceUntilIdle()
        assertFalse(viewModel.isLoading.value)
    }

    @Test
    fun `isLoading is a StateFlow`() {
        assertNotNull(viewModel.isLoading)
        assertTrue(viewModel.isLoading.value)
    }
}
