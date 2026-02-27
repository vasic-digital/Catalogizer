package com.catalogizer.android.testutils

import com.catalogizer.android.data.models.MediaItem
import com.catalogizer.android.data.models.User
import com.catalogizer.android.data.remote.ApiResult
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.flowOf

/**
 * Helper class for creating mock data in tests.
 */
object MockRepositoryHelper {
    
    // Mock Media Items
    fun createMockMediaItem(
        id: Long = 1L,
        title: String = "Test Movie",
        mediaType: String = "movie",
        year: Int = 2023
    ): MediaItem {
        return MediaItem(
            id = id,
            title = title,
            mediaType = mediaType,
            year = year,
            description = "Test description",
            coverImage = "/test/cover.jpg",
            rating = 8.5,
            quality = "1080p",
            fileSize = 1024L * 1024 * 1024,
            duration = 120,
            directoryPath = "/test/media",
            smbPath = "smb://server/test",
            createdAt = "2024-01-15T10:00:00Z",
            updatedAt = "2024-02-15T10:00:00Z",
            isFavorite = false,
            watchProgress = 0.0,
            lastWatched = null,
            isDownloaded = false
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
            firstName = "Test",
            lastName = "User",
            role = "user",
            isActive = true,
            lastLogin = "2024-02-15T10:00:00Z",
            createdAt = "2024-01-15T10:00:00Z",
            updatedAt = "2024-02-15T10:00:00Z",
            permissions = listOf("read:media")
        )
    }
    
    // Mock API Results
    fun <T> createSuccessApiResult(data: T): ApiResult<T> {
        return ApiResult.success(data)
    }
    
    fun <T> createErrorApiResult(message: String = "Test error"): ApiResult<T> {
        return ApiResult.error(message)
    }
    
    // Mock Flows
    fun <T> createMockFlow(data: T): Flow<T> {
        return flowOf(data)
    }
    
    fun <T> createMockFlowSequence(vararg items: T): Flow<T> {
        return kotlinx.coroutines.flow.flow {
            items.forEach { emit(it) }
        }
    }
}
