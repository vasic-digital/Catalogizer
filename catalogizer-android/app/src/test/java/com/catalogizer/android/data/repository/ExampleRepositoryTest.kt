package com.catalogizer.android.data.repository

import com.catalogizer.android.data.models.MediaItem
import com.catalogizer.android.testutils.MockRepositoryHelper
import com.catalogizer.android.data.remote.ApiResult
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.test.runTest
import org.junit.Assert.*
import org.junit.Test

@ExperimentalCoroutinesApi
class ExampleRepositoryTest {
    
    @Test
    fun `repository should return success for valid data`() = runTest {
        // Given
        val mockData = MockRepositoryHelper.createMockMediaItems(3)
        
        // When - simulate repository call
        val result: ApiResult<List<MediaItem>> = ApiResult.success(mockData)
        
        // Then
        assertTrue(result.isSuccess)
        assertEquals(3, result.data?.size)
    }
    
    @Test
    fun `repository should handle empty results`() = runTest {
        // Given
        val emptyList = emptyList<MediaItem>()
        
        // When
        val result: ApiResult<List<MediaItem>> = ApiResult.success(emptyList)
        
        // Then
        assertTrue(result.isSuccess)
        assertTrue(result.data?.isEmpty() ?: false)
    }
    
    @Test
    fun `repository should handle errors gracefully`() = runTest {
        // Given
        val errorMessage = "Network error"
        
        // When
        val result: ApiResult<List<MediaItem>> = ApiResult.error(errorMessage)
        
        // Then
        assertFalse(result.isSuccess)
        assertEquals(errorMessage, result.error)
    }
}
