package com.catalogizer.android.data.repository

import org.junit.Assert.*
import org.junit.Test

class OfflineStatsTest2 {

    @Test
    fun `OfflineStats holds correct values`() {
        val stats = OfflineStats(
            cachedItems = 100,
            pendingSyncOperations = 5,
            failedSyncOperations = 2,
            usedStorageBytes = 1073741824L,
            totalStorageBytes = 5368709120L,
            storagePercentageUsed = 20L
        )
        assertEquals(100, stats.cachedItems)
        assertEquals(5, stats.pendingSyncOperations)
        assertEquals(2, stats.failedSyncOperations)
        assertEquals(1073741824L, stats.usedStorageBytes)
        assertEquals(5368709120L, stats.totalStorageBytes)
        assertEquals(20L, stats.storagePercentageUsed)
    }

    @Test
    fun `OfflineStats with zero values`() {
        val stats = OfflineStats(
            cachedItems = 0,
            pendingSyncOperations = 0,
            failedSyncOperations = 0,
            usedStorageBytes = 0L,
            totalStorageBytes = 0L,
            storagePercentageUsed = 0L
        )
        assertEquals(0, stats.cachedItems)
        assertEquals(0, stats.pendingSyncOperations)
        assertEquals(0L, stats.usedStorageBytes)
        assertEquals(0L, stats.storagePercentageUsed)
    }

    @Test
    fun `OfflineStats equality`() {
        val stats1 = OfflineStats(10, 1, 0, 500L, 1000L, 50L)
        val stats2 = OfflineStats(10, 1, 0, 500L, 1000L, 50L)
        assertEquals(stats1, stats2)
    }

    @Test
    fun `OfflineStats inequality`() {
        val stats1 = OfflineStats(10, 1, 0, 500L, 1000L, 50L)
        val stats2 = OfflineStats(20, 1, 0, 500L, 1000L, 50L)
        assertNotEquals(stats1, stats2)
    }

    // --- OfflineDataExport ---

    @Test
    fun `OfflineDataExport holds correct values`() {
        val export = OfflineDataExport(
            mediaItems = emptyList(),
            syncOperations = emptyList(),
            searchQueries = listOf("matrix", "star wars"),
            exportTimestamp = 1000L
        )
        assertTrue(export.mediaItems.isEmpty())
        assertTrue(export.syncOperations.isEmpty())
        assertEquals(2, export.searchQueries.size)
        assertEquals(1000L, export.exportTimestamp)
    }

    @Test
    fun `OfflineDataExport with empty data`() {
        val export = OfflineDataExport(
            mediaItems = emptyList(),
            syncOperations = emptyList(),
            searchQueries = emptyList(),
            exportTimestamp = 0L
        )
        assertTrue(export.mediaItems.isEmpty())
        assertTrue(export.syncOperations.isEmpty())
        assertTrue(export.searchQueries.isEmpty())
    }
}
