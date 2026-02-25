package com.catalogizer.android.testutils

import androidx.arch.core.executor.testing.InstantTaskExecutorRule
import org.junit.Before
import org.junit.Rule
import org.junit.runner.RunWith
import org.mockito.junit.MockitoJUnitRunner

/**
 * Base class for ViewModel tests with common setup.
 */
@RunWith(MockitoJUnitRunner::class)
abstract class ViewModelTestBase {
    
    @get:Rule
    val instantExecutorRule = InstantTaskExecutorRule()
    
    @get:Rule
    val testDispatcherRule = TestDispatcherRule()
    
    @Before
    open fun setUp() {
        // Common setup for all ViewModel tests
    }
}
