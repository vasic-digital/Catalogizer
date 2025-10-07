package com.catalogizer.android.data.repository

import com.catalogizer.android.data.local.MediaDao
import com.catalogizer.android.data.models.MediaItem
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.test.runTest
import org.junit.Assert.*
import org.junit.Before
import org.junit.Test
import org.mockito.Mock
import org.mockito.Mockito.*
import org.mockito.MockitoAnnotations

@ExperimentalCoroutinesApi
class MediaRepositoryTest {

    @Mock
    private lateinit var mediaDao: MediaDao

    @Mock
    private lateinit var remoteDataSource: Any // Would be your API service

    private lateinit var repository: MediaRepository

    @Before
    fun setup() {
        MockitoAnnotations.openMocks(this)
        repository = MediaRepository(mediaDao, remoteDataSource)
    }

    @Test
    fun `getAllMedia returns list from dao`() = runTest {
        // Given
        val expectedMedia = listOf(
            MediaItem(id = 1, name = "Test Movie", path = "/movies/test.mp4", mediaType = "movie"),
            MediaItem(id = 2, name = "Test Song", path = "/music/test.mp3", mediaType = "music")
        )
        `when`(mediaDao.getAllMedia()).thenReturn(expectedMedia)

        // When
        val result = repository.getAllMedia()

        // Then
        assertEquals(expectedMedia, result)
        verify(mediaDao).getAllMedia()
    }

    @Test
    fun `getMediaById returns correct item`() = runTest {
        // Given
        val expectedMedia = MediaItem(id = 1, name = "Test Movie", path = "/movies/test.mp4", mediaType = "movie")
        `when`(mediaDao.getMediaById(1)).thenReturn(expectedMedia)

        // When
        val result = repository.getMediaById(1)

        // Then
        assertEquals(expectedMedia, result)
        verify(mediaDao).getMediaById(1)
    }

    @Test
    fun `searchMedia returns filtered results`() = runTest {
        // Given
        val query = "test"
        val expectedResults = listOf(
            MediaItem(id = 1, name = "Test Movie", path = "/movies/test.mp4", mediaType = "movie")
        )
        `when`(mediaDao.searchMedia("%$query%")).thenReturn(expectedResults)

        // When
        val result = repository.searchMedia(query)

        // Then
        assertEquals(expectedResults, result)
        verify(mediaDao).searchMedia("%$query%")
    }

    @Test
    fun `getMediaByType returns filtered results`() = runTest {
        // Given
        val mediaType = "movie"
        val expectedResults = listOf(
            MediaItem(id = 1, name = "Test Movie", path = "/movies/test.mp4", mediaType = "movie")
        )
        `when`(mediaDao.getMediaByType(mediaType)).thenReturn(expectedResults)

        // When
        val result = repository.getMediaByType(mediaType)

        // Then
        assertEquals(expectedResults, result)
        verify(mediaDao).getMediaByType(mediaType)
    }

    @Test
    fun `insertMedia calls dao insert`() = runTest {
        // Given
        val mediaItem = MediaItem(id = 1, name = "New Movie", path = "/movies/new.mp4", mediaType = "movie")

        // When
        repository.insertMedia(mediaItem)

        // Then
        verify(mediaDao).insertMedia(mediaItem)
    }

    @Test
    fun `updateMedia calls dao update`() = runTest {
        // Given
        val mediaItem = MediaItem(id = 1, name = "Updated Movie", path = "/movies/updated.mp4", mediaType = "movie")

        // When
        repository.updateMedia(mediaItem)

        // Then
        verify(mediaDao).updateMedia(mediaItem)
    }

    @Test
    fun `deleteMedia calls dao delete`() = runTest {
        // Given
        val mediaItem = MediaItem(id = 1, name = "Movie to Delete", path = "/movies/delete.mp4", mediaType = "movie")

        // When
        repository.deleteMedia(mediaItem)

        // Then
        verify(mediaDao).deleteMedia(mediaItem)
    }

    @Test
    fun `getMediaCount returns correct count`() = runTest {
        // Given
        val expectedCount = 5
        `when`(mediaDao.getMediaCount()).thenReturn(expectedCount)

        // When
        val result = repository.getMediaCount()

        // Then
        assertEquals(expectedCount, result)
        verify(mediaDao).getMediaCount()
    }

    @Test
    fun `getTotalSize returns correct size`() = runTest {
        // Given
        val expectedSize = 1000000L
        `when`(mediaDao.getTotalSize()).thenReturn(expectedSize)

        // When
        val result = repository.getTotalSize()

        // Then
        assertEquals(expectedSize, result)
        verify(mediaDao).getTotalSize()
    }
}