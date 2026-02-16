package com.catalogizer.android.data.repository

import org.junit.Assert.*
import org.junit.Test

class OfflineStatsTest {

    @Test
    fun `OfflineStats holds correct values`() {
        val stats = OfflineStats(
            cachedItems = 100,
            pendingSyncOperations = 5,
            failedSyncOperations = 2,
            usedStorageBytes = 1_000_000_000L,
            totalStorageBytes = 5_000_000_000L,
            storagePercentageUsed = 20
        )

        assertEquals(100, stats.cachedItems)
        assertEquals(5, stats.pendingSyncOperations)
        assertEquals(2, stats.failedSyncOperations)
        assertEquals(1_000_000_000L, stats.usedStorageBytes)
        assertEquals(5_000_000_000L, stats.totalStorageBytes)
        assertEquals(20L, stats.storagePercentageUsed)
    }

    @Test
    fun `OfflineStats with zero values`() {
        val stats = OfflineStats(
            cachedItems = 0,
            pendingSyncOperations = 0,
            failedSyncOperations = 0,
            usedStorageBytes = 0L,
            totalStorageBytes = 5_000_000_000L,
            storagePercentageUsed = 0
        )

        assertEquals(0, stats.cachedItems)
        assertEquals(0, stats.pendingSyncOperations)
        assertEquals(0L, stats.usedStorageBytes)
        assertEquals(0L, stats.storagePercentageUsed)
    }

    @Test
    fun `OfflineStats with full storage`() {
        val stats = OfflineStats(
            cachedItems = 500,
            pendingSyncOperations = 0,
            failedSyncOperations = 0,
            usedStorageBytes = 5_000_000_000L,
            totalStorageBytes = 5_000_000_000L,
            storagePercentageUsed = 100
        )

        assertEquals(100L, stats.storagePercentageUsed)
        assertEquals(stats.usedStorageBytes, stats.totalStorageBytes)
    }

    @Test
    fun `OfflineStats equality works correctly`() {
        val stats1 = OfflineStats(100, 5, 2, 1000L, 5000L, 20)
        val stats2 = OfflineStats(100, 5, 2, 1000L, 5000L, 20)
        val stats3 = OfflineStats(200, 5, 2, 1000L, 5000L, 20)

        assertEquals(stats1, stats2)
        assertNotEquals(stats1, stats3)
    }

    @Test
    fun `OfflineStats copy updates correctly`() {
        val original = OfflineStats(100, 5, 2, 1000L, 5000L, 20)
        val updated = original.copy(cachedItems = 150, pendingSyncOperations = 3)

        assertEquals(100, original.cachedItems)
        assertEquals(150, updated.cachedItems)
        assertEquals(3, updated.pendingSyncOperations)
    }
}
