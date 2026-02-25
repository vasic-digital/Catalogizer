package com.catalogizer.android.testutils

import com.catalogizer.android.data.models.MediaItem
import com.catalogizer.android.data.models.MediaType
import com.catalogizer.android.data.models.User
import com.catalogizer.android.data.remote.ApiResult
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.flow
import kotlinx.coroutines.flow.flowOf
import java.util.*

/**
 * Helper class for creating mock data in tests.
 */
object MockRepositoryHelper {
    
    // Mock Media Items
    fun createMockMediaItem(
        id: Long = 1L,
        title: String = "Test Movie",
        type: MediaType = MediaType.MOVIE,
        year: Int = 2023
    ): MediaItem {
        return MediaItem(
            id = id,
            title = title,
            type = type,
            year = year,
            posterPath = "/test/poster.jpg",
            backdropPath = "/test/backdrop.jpg",
            overview = "Test overview",
            rating = 8.5,
            runtime = 120,
            genres = listOf("Action", "Adventure"),
            createdAt = Date(),
            updatedAt = Date()
        )
    }
    
    fun createMockMediaItems(count: Int = 5): List<MediaItem> {
        return (1..count).map { index ->
            createMockMediaItem(
                id = index.toLong(),
                title = "Test Movie $index",
                year = 2020 + index
            )
        }
    }
    
    // Mock User
    fun createMockUser(
        id: Long = 1L,
        username: String = "testuser",
        email: String = "test@example.com"
    ): User {
        return User(
            id = id,
            username = username,
            email = email,
            createdAt = Date(),
            updatedAt = Date()
        )
    }
    
    // Mock API Results
    fun <T> createSuccessApiResult(data: T): ApiResult.Success<T> {
        return ApiResult.Success(data)
    }
    
    fun <T> createErrorApiResult(message: String = "Test error"): ApiResult.Error<T> {
        return ApiResult.Error(message)
    }
    
    fun <T> createLoadingApiResult(): ApiResult.Loading<T> {
        return ApiResult.Loading()
    }
    
    // Mock Flows
    fun <T> createMockFlow(data: T): Flow<T> {
        return flowOf(data)
    }
    
    fun <T> createMockFlowSequence(vararg items: T): Flow<T> {
        return flow {
            items.forEach { emit(it) }
        }
    }
}
