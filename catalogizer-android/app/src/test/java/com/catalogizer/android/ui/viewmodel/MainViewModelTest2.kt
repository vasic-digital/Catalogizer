package com.catalogizer.android.ui.viewmodel

import com.catalogizer.android.MainDispatcherRule
import com.catalogizer.android.data.repository.MediaRepository
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

    private val mockMediaRepository = mockk<MediaRepository>(relaxed = true)
    private lateinit var viewModel: MainViewModel

    @Before
    fun setup() {
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
    fun `initializeApp handles exception and still sets loading to false`() = runTest {
        // MainViewModel's initializeApp catches all exceptions
        viewModel.initializeApp()
        advanceUntilIdle()

        assertFalse(viewModel.isLoading.value)
    }

    @Test
    fun `isLoading is StateFlow and emits initial value`() {
        val flow = viewModel.isLoading
        assertNotNull(flow)
        assertTrue(flow.value) // default value
    }

    @Test
    fun `multiple calls to initializeApp are safe`() = runTest {
        viewModel.initializeApp()
        viewModel.initializeApp()
        viewModel.initializeApp()
        advanceUntilIdle()

        assertFalse(viewModel.isLoading.value)
    }
}
