package com.catalogizer.android.testutils

import com.catalogizer.android.data.models.*

/**
 * Generates test data for Android tests.
 */
object TestDataGenerator {
    
    // Generate test media items
    fun generateMediaItems(count: Int = 10): List<MediaItem> {
        val items = mutableListOf<MediaItem>()
        val types = listOf("movie", "tv_show", "music", "game", "book")
        val genres = listOf(
            "Action", "Adventure", "Comedy", "Drama", "Horror",
            "Sci-Fi", "Fantasy", "Romance", "Thriller", "Documentary"
        )
        
        for (i in 1..count) {
            val type = types[i % types.size]
            val title = when (type) {
                "movie" -> "Test Movie $i"
                "tv_show" -> "Test TV Show $i"
                "music" -> "Test Album $i"
                "game" -> "Test Game $i"
                "book" -> "Test Book $i"
                else -> "Test Item $i"
            }
            
            items.add(
                MediaItem(
                    id = i.toLong(),
                    title = title,
                    mediaType = type,
                    year = 2010 + (i % 15),
                    description = "This is a test description for $title. It's a great piece of media that everyone should experience.",
                    coverImage = "/covers/cover_$i.jpg",
                    rating = 5.0 + (i % 5).toDouble(),
                    quality = listOf("720p", "1080p", "4k")[i % 3],
                    fileSize = 1024L * 1024 * 1024 * (i % 10 + 1),
                    duration = 90 + (i % 60),
                    directoryPath = "/test/media/$type",
                    smbPath = "smb://server/media/$type/$title",
                    createdAt = "2024-01-${(i % 28 + 1).toString().padStart(2, '0')}T10:00:00Z",
                    updatedAt = "2024-02-${(i % 28 + 1).toString().padStart(2, '0')}T10:00:00Z",
                    isFavorite = i % 3 == 0,
                    watchProgress = (i % 100).toDouble(),
                    lastWatched = if (i % 2 == 0) "2024-02-${(i % 28 + 1).toString().padStart(2, '0')}T10:00:00Z" else null,
                    isDownloaded = i % 4 == 0
                )
            )
        }
        
        return items
    }
    
    // Generate test users
    fun generateUsers(count: Int = 5): List<User> {
        val users = mutableListOf<User>()
        
        for (i in 1..count) {
            users.add(
                User(
                    id = i.toLong(),
                    username = "user$i",
                    email = "user$i@example.com",
                    firstName = "First$i",
                    lastName = "Last$i",
                    role = if (i == 1) "admin" else "user",
                    isActive = true,
                    createdAt = "2024-01-${(i % 28 + 1).toString().padStart(2, '0')}T10:00:00Z",
                    updatedAt = "2024-02-${(i % 28 + 1).toString().padStart(2, '0')}T10:00:00Z"
                )
            )
        }
        
        return users
    }
    
    // Generate test search results
    fun generateSearchResults(query: String, count: Int = 5): List<MediaItem> {
        return generateMediaItems(count).map { item ->
            item.copy(title = "$query ${item.title}")
        }
    }
}
