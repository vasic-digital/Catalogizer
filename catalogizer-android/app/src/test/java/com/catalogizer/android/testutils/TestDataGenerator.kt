package com.catalogizer.android.testutils

import com.catalogizer.android.data.models.*

/**
 * Generates test data for Android tests.
 */
object TestDataGenerator {
    
    // Generate test media items
    fun generateMediaItems(count: Int = 10): List<MediaItem> {
        val items = mutableListOf<MediaItem>()
        val types = MediaType.values()
        val genres = listOf(
            "Action", "Adventure", "Comedy", "Drama", "Horror",
            "Sci-Fi", "Fantasy", "Romance", "Thriller", "Documentary"
        )
        
        for (i in 1..count) {
            val type = types[i % types.size]
            val title = when (type) {
                MediaType.MOVIE -> "Test Movie $i"
                MediaType.TV_SHOW -> "Test TV Show $i"
                MediaType.MUSIC_ALBUM -> "Test Album $i"
                MediaType.GAME -> "Test Game $i"
                MediaType.BOOK -> "Test Book $i"
                else -> "Test Item $i"
            }
            
            val itemGenres = genres.shuffled().take(3)
            
            items.add(
                MediaItem(
                    id = i.toLong(),
                    title = title,
                    type = type,
                    year = 2010 + (i % 15),
                    posterPath = "/posters/poster_$i.jpg",
                    backdropPath = "/backdrops/backdrop_$i.jpg",
                    overview = "This is a test overview for $title. It's a great piece of media that everyone should experience.",
                    rating = 5.0 + (i % 5).toDouble(),
                    runtime = 90 + (i % 60),
                    genres = itemGenres,
                    createdAt = java.util.Date(System.currentTimeMillis() - i * 86400000L),
                    updatedAt = java.util.Date()
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
                    createdAt = java.util.Date(System.currentTimeMillis() - i * 86400000L),
                    updatedAt = java.util.Date()
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
