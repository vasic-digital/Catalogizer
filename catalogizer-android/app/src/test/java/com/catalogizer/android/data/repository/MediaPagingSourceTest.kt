package com.catalogizer.android.data.repository

import androidx.paging.PagingSource
import com.catalogizer.android.data.models.MediaItem
import com.catalogizer.android.data.models.MediaSearchRequest
import com.catalogizer.android.data.models.MediaSearchResponse
import com.catalogizer.android.data.remote.CatalogizerApi
import io.mockk.*
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.test.runTest
import org.junit.After
import org.junit.Assert.*
import org.junit.Before
import org.junit.Test
import retrofit2.Response

@OptIn(ExperimentalCoroutinesApi::class)
class MediaPagingSourceTest {

    private val mockApi = mockk<CatalogizerApi>(relaxed = true)

    private val testMediaItem = MediaItem(
        id = 1L,
        title = "Test Movie",
        mediaType = "movie",
        directoryPath = "/movies/test",
        createdAt = "2024-01-01",
        updatedAt = "2024-01-01"
    )

    @After
    fun tearDown() {
        clearAllMocks()
    }

    @Test
    fun `load returns Page on successful API response`() = runTest {
        val items = listOf(testMediaItem, testMediaItem.copy(id = 2, title = "Movie 2"))
        val searchResponse = MediaSearchResponse(items = items, total = 50, limit = 20, offset = 0)
        coEvery { mockApi.searchMedia(any(), any(), any(), any(), any(), any(), any(), any(), any(), any()) } returns Response.success(searchResponse)

        val pagingSource = MediaPagingSource(mockApi, MediaSearchRequest())
        val result = pagingSource.load(PagingSource.LoadParams.Refresh(null, 20, false))

        assertTrue(result is PagingSource.LoadResult.Page)
        val page = result as PagingSource.LoadResult.Page
        assertEquals(2, page.data.size)
        assertNull(page.prevKey)
        assertEquals(20, page.nextKey)
    }

    @Test
    fun `load returns null nextKey when all items loaded`() = runTest {
        val items = listOf(testMediaItem)
        val searchResponse = MediaSearchResponse(items = items, total = 1, limit = 20, offset = 0)
        coEvery { mockApi.searchMedia(any(), any(), any(), any(), any(), any(), any(), any(), any(), any()) } returns Response.success(searchResponse)

        val pagingSource = MediaPagingSource(mockApi, MediaSearchRequest())
        val result = pagingSource.load(PagingSource.LoadParams.Refresh(null, 20, false))

        assertTrue(result is PagingSource.LoadResult.Page)
        val page = result as PagingSource.LoadResult.Page
        assertNull(page.nextKey)
    }

    @Test
    fun `load returns Error on API failure`() = runTest {
        coEvery { mockApi.searchMedia(any(), any(), any(), any(), any(), any(), any(), any(), any(), any()) } returns Response.error(
            500,
            okhttp3.ResponseBody.create(null, "Server error")
        )

        val pagingSource = MediaPagingSource(mockApi, MediaSearchRequest())
        val result = pagingSource.load(PagingSource.LoadParams.Refresh(null, 20, false))

        assertTrue(result is PagingSource.LoadResult.Error)
    }

    @Test
    fun `load returns Error on exception`() = runTest {
        coEvery { mockApi.searchMedia(any(), any(), any(), any(), any(), any(), any(), any(), any(), any()) } throws RuntimeException("Network error")

        val pagingSource = MediaPagingSource(mockApi, MediaSearchRequest())
        val result = pagingSource.load(PagingSource.LoadParams.Refresh(null, 20, false))

        assertTrue(result is PagingSource.LoadResult.Error)
    }

    @Test
    fun `load with offset calculates correct prev and next keys`() = runTest {
        val items = listOf(testMediaItem)
        val searchResponse = MediaSearchResponse(items = items, total = 100, limit = 20, offset = 20)
        coEvery { mockApi.searchMedia(any(), any(), any(), any(), any(), any(), any(), any(), any(), any()) } returns Response.success(searchResponse)

        val pagingSource = MediaPagingSource(mockApi, MediaSearchRequest())
        val result = pagingSource.load(PagingSource.LoadParams.Refresh(20, 20, false))

        assertTrue(result is PagingSource.LoadResult.Page)
        val page = result as PagingSource.LoadResult.Page
        assertEquals(0, page.prevKey)
        assertEquals(40, page.nextKey)
    }

    @Test
    fun `load passes search request parameters to API`() = runTest {
        val searchResponse = MediaSearchResponse(items = emptyList(), total = 0, limit = 20, offset = 0)
        coEvery { mockApi.searchMedia(any(), any(), any(), any(), any(), any(), any(), any(), any(), any()) } returns Response.success(searchResponse)

        val request = MediaSearchRequest(
            query = "inception",
            mediaType = "movie",
            yearMin = 2010,
            ratingMin = 7.0,
            sortBy = "rating",
            sortOrder = "desc"
        )

        val pagingSource = MediaPagingSource(mockApi, request)
        pagingSource.load(PagingSource.LoadParams.Refresh(null, 20, false))

        coVerify {
            mockApi.searchMedia(
                query = "inception",
                mediaType = "movie",
                yearMin = 2010,
                yearMax = null,
                ratingMin = 7.0,
                quality = null,
                sortBy = "rating",
                sortOrder = "desc",
                limit = 20,
                offset = 0
            )
        }
    }
}
