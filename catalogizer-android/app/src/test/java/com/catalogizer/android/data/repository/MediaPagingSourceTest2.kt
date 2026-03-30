package com.catalogizer.android.data.repository

import androidx.paging.PagingSource
import com.catalogizer.android.MainDispatcherRule
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
import org.junit.Rule
import org.junit.Test
import retrofit2.Response

@OptIn(ExperimentalCoroutinesApi::class)
class MediaPagingSourceTest2 {

    @get:Rule
    val mainDispatcherRule = MainDispatcherRule()

    private val mockApi = mockk<CatalogizerApi>(relaxed = true)

    private val testItem1 = MediaItem(
        id = 1L, title = "Movie 1", mediaType = "movie",
        directoryPath = "/path", createdAt = "2024-01-01", updatedAt = "2024-01-01"
    )
    private val testItem2 = MediaItem(
        id = 2L, title = "Movie 2", mediaType = "movie",
        directoryPath = "/path", createdAt = "2024-01-01", updatedAt = "2024-01-01"
    )

    @After
    fun tearDown() {
        clearAllMocks()
    }

    @Test
    fun `load returns Page with items on success`() = runTest {
        val searchResponse = MediaSearchResponse(
            items = listOf(testItem1, testItem2),
            total = 100,
            limit = 20,
            offset = 0
        )
        coEvery { mockApi.searchMedia(any(), any(), any(), any(), any(), any(), any(), any(), any(), any()) } returns
            Response.success(searchResponse)

        val pagingSource = MediaPagingSource(mockApi, MediaSearchRequest())
        val result = pagingSource.load(
            PagingSource.LoadParams.Refresh(key = null, loadSize = 20, placeholdersEnabled = false)
        )

        assertTrue(result is PagingSource.LoadResult.Page)
        val page = result as PagingSource.LoadResult.Page
        assertEquals(2, page.data.size)
        assertEquals("Movie 1", page.data[0].title)
        assertNull(page.prevKey)
        assertEquals(20, page.nextKey)
    }

    @Test
    fun `load returns null nextKey when at end of data`() = runTest {
        val searchResponse = MediaSearchResponse(
            items = listOf(testItem1),
            total = 1,
            limit = 20,
            offset = 0
        )
        coEvery { mockApi.searchMedia(any(), any(), any(), any(), any(), any(), any(), any(), any(), any()) } returns
            Response.success(searchResponse)

        val pagingSource = MediaPagingSource(mockApi, MediaSearchRequest())
        val result = pagingSource.load(
            PagingSource.LoadParams.Refresh(key = null, loadSize = 20, placeholdersEnabled = false)
        )

        assertTrue(result is PagingSource.LoadResult.Page)
        val page = result as PagingSource.LoadResult.Page
        assertNull(page.nextKey)
    }

    @Test
    fun `load returns Error on API failure`() = runTest {
        coEvery { mockApi.searchMedia(any(), any(), any(), any(), any(), any(), any(), any(), any(), any()) } returns
            Response.error(500, okhttp3.ResponseBody.create(null, "Server error"))

        val pagingSource = MediaPagingSource(mockApi, MediaSearchRequest())
        val result = pagingSource.load(
            PagingSource.LoadParams.Refresh(key = null, loadSize = 20, placeholdersEnabled = false)
        )

        assertTrue(result is PagingSource.LoadResult.Error)
    }

    @Test
    fun `load returns Error on exception`() = runTest {
        coEvery { mockApi.searchMedia(any(), any(), any(), any(), any(), any(), any(), any(), any(), any()) } throws
            RuntimeException("Network error")

        val pagingSource = MediaPagingSource(mockApi, MediaSearchRequest())
        val result = pagingSource.load(
            PagingSource.LoadParams.Refresh(key = null, loadSize = 20, placeholdersEnabled = false)
        )

        assertTrue(result is PagingSource.LoadResult.Error)
    }

    @Test
    fun `load with offset returns correct prevKey`() = runTest {
        val searchResponse = MediaSearchResponse(
            items = listOf(testItem1),
            total = 100,
            limit = 20,
            offset = 40
        )
        coEvery { mockApi.searchMedia(any(), any(), any(), any(), any(), any(), any(), any(), any(), any()) } returns
            Response.success(searchResponse)

        val pagingSource = MediaPagingSource(mockApi, MediaSearchRequest())
        val result = pagingSource.load(
            PagingSource.LoadParams.Append(key = 40, loadSize = 20, placeholdersEnabled = false)
        )

        assertTrue(result is PagingSource.LoadResult.Page)
        val page = result as PagingSource.LoadResult.Page
        assertEquals(20, page.prevKey)
        assertEquals(60, page.nextKey)
    }

    @Test
    fun `load passes search request parameters to API`() = runTest {
        val searchRequest = MediaSearchRequest(
            query = "matrix",
            mediaType = "movie",
            sortBy = "rating",
            sortOrder = "desc"
        )
        val searchResponse = MediaSearchResponse(
            items = emptyList(), total = 0, limit = 20, offset = 0
        )
        coEvery { mockApi.searchMedia(any(), any(), any(), any(), any(), any(), any(), any(), any(), any()) } returns
            Response.success(searchResponse)

        val pagingSource = MediaPagingSource(mockApi, searchRequest)
        pagingSource.load(
            PagingSource.LoadParams.Refresh(key = null, loadSize = 20, placeholdersEnabled = false)
        )

        coVerify {
            mockApi.searchMedia(
                query = "matrix",
                mediaType = "movie",
                sortBy = "rating",
                sortOrder = "desc",
                limit = 20,
                offset = 0,
                yearMin = null,
                yearMax = null,
                ratingMin = null,
                quality = null
            )
        }
    }
}
