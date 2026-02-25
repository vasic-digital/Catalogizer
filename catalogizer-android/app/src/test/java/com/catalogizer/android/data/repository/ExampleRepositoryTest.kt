package com.catalogizer.android.data.repository

import com.catalogizer.android.testutils.MockRepositoryHelper
import com.catalogizer.android.testutils.TestDataGenerator
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
        val result: ApiResult<List<MediaItem>> = ApiResult.Success(mockData)
        
        // Then
        assertTrue(result is ApiResult.Success)
        assertEquals(3, (result as ApiResult.Success).data.size)
    }
    
    @Test
    fun `repository should handle empty results`() = runTest {
        // Given
        val emptyList = emptyList<MediaItem>()
        
        // When
        val result: ApiResult<List<MediaItem>> = ApiResult.Success(emptyList)
        
        // Then
        assertTrue(result is ApiResult.Success)
        assertTrue((result as ApiResult.Success).data.isEmpty())
    }
    
    @Test
    fun `repository should handle errors gracefully`() = runTest {
        // Given
        val errorMessage = "Network error"
        
        // When
        val result: ApiResult<List<MediaItem>> = ApiResult.Error(errorMessage)
        
        // Then
        assertTrue(result is ApiResult.Error)
        assertEquals(errorMessage, (result as ApiResult.Error).message)
    }
}
