package com.catalogizer.android.ui.viewmodel

import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.test.runTest
import org.junit.Test
import org.junit.Assert.*

class ExampleStateViewModelTest {
    
    // Example ViewModel for testing
    class ExampleViewModel {
        private val _count = MutableStateFlow(0)
        val count: StateFlow<Int> = _count.asStateFlow()
        
        private val _text = MutableStateFlow("")
        val text: StateFlow<String> = _text.asStateFlow()
        
        fun increment() {
            _count.value++
        }
        
        fun updateText(newText: String) {
            _text.value = newText
        }
        
        fun reset() {
            _count.value = 0
            _text.value = ""
        }
    }
    
    @Test
    fun `viewmodel should initialize with default values`() = runTest {
        // Given
        val viewModel = ExampleViewModel()
        
        // Then
        assertEquals(0, viewModel.count.value)
        assertEquals("", viewModel.text.value)
    }
    
    @Test
    fun `viewmodel should update count when incremented`() = runTest {
        // Given
        val viewModel = ExampleViewModel()
        
        // When
        viewModel.increment()
        
        // Then
        assertEquals(1, viewModel.count.value)
    }
    
    @Test
    fun `viewmodel should update text`() = runTest {
        // Given
        val viewModel = ExampleViewModel()
        val testText = "Hello, World!"
        
        // When
        viewModel.updateText(testText)
        
        // Then
        assertEquals(testText, viewModel.text.value)
    }
    
    @Test
    fun `viewmodel should reset to initial state`() = runTest {
        // Given
        val viewModel = ExampleViewModel()
        viewModel.increment()
        viewModel.updateText("Test")
        
        // When
        viewModel.reset()
        
        // Then
        assertEquals(0, viewModel.count.value)
        assertEquals("", viewModel.text.value)
    }
}
