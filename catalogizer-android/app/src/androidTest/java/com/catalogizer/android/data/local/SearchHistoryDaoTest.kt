package com.catalogizer.android.data.local

import android.content.Context
import androidx.room.Room
import androidx.test.core.app.ApplicationProvider
import androidx.test.ext.junit.runners.AndroidJUnit4
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.runBlocking
import org.junit.After
import org.junit.Assert.*
import org.junit.Before
import org.junit.Test
import org.junit.runner.RunWith

/**
 * Instrumentation tests for SearchHistoryDao
 */
@RunWith(AndroidJUnit4::class)
class SearchHistoryDaoTest {

    private lateinit var database: CatalogizerDatabase
    private lateinit var searchHistoryDao: SearchHistoryDao

    @Before
    fun setup() {
        val context = ApplicationProvider.getApplicationContext<Context>()
        database = Room.inMemoryDatabaseBuilder(
            context,
            CatalogizerDatabase::class.java
        ).allowMainThreadQueries().build()
        searchHistoryDao = database.searchHistoryDao()
    }

    @After
    fun teardown() {
        database.close()
    }

    private fun createSearchHistory(
        query: String,
        timestamp: Long = System.currentTimeMillis(),
        resultsCount: Int = 10
    ): SearchHistory {
        return SearchHistory(
            query = query,
            timestamp = timestamp,
            resultsCount = resultsCount
        )
    }

    @Test
    fun insertAndRetrieveSearchHistory() = runBlocking {
        // Given
        val search = createSearchHistory("matrix")

        // When
        searchHistoryDao.insertSearch(search)
        val recent = searchHistoryDao.getRecentSearches().first()

        // Then
        assertEquals(1, recent.size)
        assertEquals("matrix", recent[0].query)
    }

    @Test
    fun getRecentSearchesReturnsLimitedResults() = runBlocking {
        // Given
        val searches = (1..15).map { i ->
            createSearchHistory(
                query = "search $i",
                timestamp = System.currentTimeMillis() + i * 1000
            )
        }
        searches.forEach { searchHistoryDao.insertSearch(it) }

        // When
        val recent = searchHistoryDao.getRecentSearches(5).first()

        // Then
        assertEquals(5, recent.size)
        // Should be ordered by timestamp DESC
        assertEquals("search 15", recent[0].query)
    }

    @Test
    fun getRecentSearchesOrdersByTimestampDesc() = runBlocking {
        // Given
        val baseTime = System.currentTimeMillis()
        val searches = listOf(
            createSearchHistory("old search", timestamp = baseTime - 2000),
            createSearchHistory("new search", timestamp = baseTime),
            createSearchHistory("middle search", timestamp = baseTime - 1000)
        )
        searches.forEach { searchHistoryDao.insertSearch(it) }

        // When
        val recent = searchHistoryDao.getRecentSearches().first()

        // Then
        assertEquals(3, recent.size)
        assertEquals("new search", recent[0].query)
        assertEquals("middle search", recent[1].query)
        assertEquals("old search", recent[2].query)
    }

    @Test
    fun deleteSearchByQuery() = runBlocking {
        // Given
        val search1 = createSearchHistory("matrix")
        val search2 = createSearchHistory("inception")
        searchHistoryDao.insertSearch(search1)
        searchHistoryDao.insertSearch(search2)

        // When
        searchHistoryDao.deleteSearch("matrix")
        val recent = searchHistoryDao.getRecentSearches().first()

        // Then
        assertEquals(1, recent.size)
        assertEquals("inception", recent[0].query)
    }

    @Test
    fun clearHistory() = runBlocking {
        // Given
        val searches = listOf(
            createSearchHistory("search 1"),
            createSearchHistory("search 2"),
            createSearchHistory("search 3")
        )
        searches.forEach { searchHistoryDao.insertSearch(it) }

        // When
        searchHistoryDao.clearHistory()
        val recent = searchHistoryDao.getRecentSearches().first()

        // Then
        assertTrue(recent.isEmpty())
    }

    @Test
    fun deleteOldSearches() = runBlocking {
        // Given
        val baseTime = System.currentTimeMillis()
        val oldSearch = createSearchHistory("old", timestamp = baseTime - 100000)
        val newSearch = createSearchHistory("new", timestamp = baseTime)
        searchHistoryDao.insertSearch(oldSearch)
        searchHistoryDao.insertSearch(newSearch)

        // When
        searchHistoryDao.deleteOldSearches(baseTime - 50000)
        val recent = searchHistoryDao.getRecentSearches().first()

        // Then
        assertEquals(1, recent.size)
        assertEquals("new", recent[0].query)
    }

    @Test
    fun duplicateQueryReplacesExisting() = runBlocking {
        // Given
        val search1 = createSearchHistory("matrix", resultsCount = 5)
        searchHistoryDao.insertSearch(search1)

        // When
        val search2 = createSearchHistory("matrix", resultsCount = 15)
        searchHistoryDao.insertSearch(search2)
        val recent = searchHistoryDao.getRecentSearches().first()

        // Then
        // Due to REPLACE strategy, we expect updated values
        assertEquals(1, recent.size)
    }

    @Test
    fun searchHistoryStoresResultsCount() = runBlocking {
        // Given
        val search = createSearchHistory("test query", resultsCount = 42)

        // When
        searchHistoryDao.insertSearch(search)
        val recent = searchHistoryDao.getRecentSearches().first()

        // Then
        assertEquals(1, recent.size)
        assertEquals(42, recent[0].resultsCount)
    }
}
