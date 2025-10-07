package com.catalogizer.android.ui.viewmodel

import androidx.arch.core.executor.testing.InstantTaskExecutorRule
import androidx.lifecycle.Observer
import com.catalogizer.android.data.models.MediaItem
import com.catalogizer.android.data.repository.MediaRepository
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.test.runTest
import org.junit.Assert.*
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.mockito.Mock
import org.mockito.Mockito.*
import org.mockito.MockitoAnnotations

@ExperimentalCoroutinesApi
class HomeViewModelTest {

    @get:Rule
    val instantTaskExecutorRule = InstantTaskExecutorRule()

    @Mock
    private lateinit var mediaRepository: MediaRepository

    @Mock
    private lateinit var mediaListObserver: Observer<List<MediaItem>>

    @Mock
    private lateinit var loadingObserver: Observer<Boolean>

    @Mock
    private lateinit var errorObserver: Observer<String>

    private lateinit var viewModel: HomeViewModel

    @Before
    fun setup() {
        MockitoAnnotations.openMocks(this)
        viewModel = HomeViewModel(mediaRepository)
        viewModel.mediaList.observeForever(mediaListObserver)
        viewModel.isLoading.observeForever(loadingObserver)
        viewModel.error.observeForever(errorObserver)
    }

    @Test
    fun `loadMedia success updates media list`() = runTest {
        // Given
        val mediaItems = listOf(
            MediaItem(id = 1, name = "Movie 1", path = "/movies/movie1.mp4", mediaType = "movie"),
            MediaItem(id = 2, name = "Song 1", path = "/music/song1.mp3", mediaType = "music")
        )
        `when`(mediaRepository.getAllMedia()).thenReturn(mediaItems)

        // When
        viewModel.loadMedia()

        // Then
        verify(mediaRepository).getAllMedia()
        verify(mediaListObserver).onChanged(mediaItems)
        verify(loadingObserver).onChanged(true)
        verify(loadingObserver).onChanged(false)
        verify(errorObserver, never()).onChanged(anyString())
    }

    @Test
    fun `loadMedia failure shows error`() = runTest {
        // Given
        val errorMessage = "Failed to load media"
        `when`(mediaRepository.getAllMedia()).thenThrow(RuntimeException(errorMessage))

        // When
        viewModel.loadMedia()

        // Then
        verify(mediaRepository).getAllMedia()
        verify(mediaListObserver).onChanged(emptyList())
        verify(loadingObserver).onChanged(true)
        verify(loadingObserver).onChanged(false)
        verify(errorObserver).onChanged("Failed to load media: $errorMessage")
    }

    @Test
    fun `searchMedia success updates media list`() = runTest {
        // Given
        val query = "movie"
        val searchResults = listOf(
            MediaItem(id = 1, name = "Test Movie", path = "/movies/test.mp4", mediaType = "movie")
        )
        `when`(mediaRepository.searchMedia(query)).thenReturn(searchResults)

        // When
        viewModel.searchMedia(query)

        // Then
        verify(mediaRepository).searchMedia(query)
        verify(mediaListObserver).onChanged(searchResults)
        verify(loadingObserver).onChanged(true)
        verify(loadingObserver).onChanged(false)
        verify(errorObserver, never()).onChanged(anyString())
    }

    @Test
    fun `searchMedia with empty query loads all media`() = runTest {
        // Given
        val allMedia = listOf(
            MediaItem(id = 1, name = "Movie 1", path = "/movies/movie1.mp4", mediaType = "movie"),
            MediaItem(id = 2, name = "Song 1", path = "/music/song1.mp3", mediaType = "music")
        )
        `when`(mediaRepository.getAllMedia()).thenReturn(allMedia)

        // When
        viewModel.searchMedia("")

        // Then
        verify(mediaRepository).getAllMedia()
        verify(mediaRepository, never()).searchMedia(anyString())
        verify(mediaListObserver).onChanged(allMedia)
    }

    @Test
    fun `filterByType success updates media list`() = runTest {
        // Given
        val mediaType = "movie"
        val filteredMedia = listOf(
            MediaItem(id = 1, name = "Movie 1", path = "/movies/movie1.mp4", mediaType = "movie"),
            MediaItem(id = 2, name = "Movie 2", path = "/movies/movie2.mp4", mediaType = "movie")
        )
        `when`(mediaRepository.getMediaByType(mediaType)).thenReturn(filteredMedia)

        // When
        viewModel.filterByType(mediaType)

        // Then
        verify(mediaRepository).getMediaByType(mediaType)
        verify(mediaListObserver).onChanged(filteredMedia)
        verify(loadingObserver).onChanged(true)
        verify(loadingObserver).onChanged(false)
        verify(errorObserver, never()).onChanged(anyString())
    }

    @Test
    fun `filterByType with null type loads all media`() = runTest {
        // Given
        val allMedia = listOf(
            MediaItem(id = 1, name = "Movie 1", path = "/movies/movie1.mp4", mediaType = "movie"),
            MediaItem(id = 2, name = "Song 1", path = "/music/song1.mp3", mediaType = "music")
        )
        `when`(mediaRepository.getAllMedia()).thenReturn(allMedia)

        // When
        viewModel.filterByType(null)

        // Then
        verify(mediaRepository).getAllMedia()
        verify(mediaRepository, never()).getMediaByType(anyString())
        verify(mediaListObserver).onChanged(allMedia)
    }

    @Test
    fun `refreshData reloads media`() = runTest {
        // Given
        val mediaItems = listOf(
            MediaItem(id = 1, name = "Movie 1", path = "/movies/movie1.mp4", mediaType = "movie")
        )
        `when`(mediaRepository.getAllMedia()).thenReturn(mediaItems)

        // When
        viewModel.refreshData()

        // Then
        verify(mediaRepository).getAllMedia()
        verify(mediaListObserver).onChanged(mediaItems)
    }

    @Test
    fun `getMediaStats returns correct statistics`() = runTest {
        // Given
        val mediaItems = listOf(
            MediaItem(id = 1, name = "Movie 1", path = "/movies/movie1.mp4", mediaType = "movie", size = 1000000),
            MediaItem(id = 2, name = "Movie 2", path = "/movies/movie2.mp4", mediaType = "movie", size = 2000000),
            MediaItem(id = 3, name = "Song 1", path = "/music/song1.mp3", mediaType = "music", size = 5000000)
        )
        `when`(mediaRepository.getAllMedia()).thenReturn(mediaItems)

        // When
        viewModel.loadMedia()
        val stats = viewModel.getMediaStats()

        // Then
        assertEquals(3, stats.totalItems)
        assertEquals(2, stats.moviesCount)
        assertEquals(1, stats.musicCount)
        assertEquals(8000000L, stats.totalSize)
    }

    @Test
    fun `clearError resets error state`() {
        // Given
        viewModel.error.value = "Test error"

        // When
        viewModel.clearError()

        // Then
        assertNull(viewModel.error.value)
    }

    @Test
    fun `initial state is correct`() {
        assertNotNull(viewModel.mediaList)
        assertNotNull(viewModel.isLoading)
        assertNotNull(viewModel.error)
        assertEquals(emptyList<MediaItem>(), viewModel.mediaList.value)
        assertEquals(false, viewModel.isLoading.value)
        assertNull(viewModel.error.value)
    }
}