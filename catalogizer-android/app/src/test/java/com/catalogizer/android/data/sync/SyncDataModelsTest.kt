package com.catalogizer.android.data.sync

import kotlinx.serialization.encodeToString
import kotlinx.serialization.json.Json
import org.junit.Assert.*
import org.junit.Before
import org.junit.Test

class SyncDataModelsTest {

    private lateinit var json: Json

    @Before
    fun setup() {
        json = Json { ignoreUnknownKeys = true }
    }

    // --- SyncStatus ---

    @Test
    fun `SyncStatus has correct defaults`() {
        val status = SyncStatus()

        assertFalse(status.isRunning)
        assertNull(status.lastSyncTime)
        assertNull(status.lastSyncResult)
        assertEquals(0, status.pendingOperations)
    }

    @Test
    fun `SyncStatus with running state`() {
        val status = SyncStatus(isRunning = true, pendingOperations = 5)

        assertTrue(status.isRunning)
        assertEquals(5, status.pendingOperations)
    }

    @Test
    fun `SyncStatus copy updates correctly`() {
        val initial = SyncStatus()
        val running = initial.copy(isRunning = true)
        val completed = running.copy(
            isRunning = false,
            lastSyncTime = 1000L,
            lastSyncResult = SyncResult(success = true, timestamp = 1000L, syncedItems = 10)
        )

        assertFalse(initial.isRunning)
        assertTrue(running.isRunning)
        assertFalse(completed.isRunning)
        assertEquals(1000L, completed.lastSyncTime)
        assertTrue(completed.lastSyncResult!!.success)
    }

    @Test
    fun `SyncStatus serialization round-trip`() {
        val original = SyncStatus(
            isRunning = false,
            lastSyncTime = 1000L,
            lastSyncResult = SyncResult(success = true, timestamp = 1000L, syncedItems = 5),
            pendingOperations = 3
        )

        val serialized = json.encodeToString(original)
        val deserialized = json.decodeFromString<SyncStatus>(serialized)

        assertEquals(original.isRunning, deserialized.isRunning)
        assertEquals(original.lastSyncTime, deserialized.lastSyncTime)
        assertEquals(original.pendingOperations, deserialized.pendingOperations)
    }

    // --- SyncResult ---

    @Test
    fun `SyncResult successful result`() {
        val result = SyncResult(
            success = true,
            timestamp = 1000L,
            syncedItems = 10,
            failedItems = 0
        )

        assertTrue(result.success)
        assertEquals(1000L, result.timestamp)
        assertEquals(10, result.syncedItems)
        assertEquals(0, result.failedItems)
        assertNull(result.errorMessage)
    }

    @Test
    fun `SyncResult failed result`() {
        val result = SyncResult(
            success = false,
            timestamp = 1000L,
            syncedItems = 5,
            failedItems = 3,
            errorMessage = "Network timeout"
        )

        assertFalse(result.success)
        assertEquals(5, result.syncedItems)
        assertEquals(3, result.failedItems)
        assertEquals("Network timeout", result.errorMessage)
    }

    @Test
    fun `SyncResult serialization round-trip`() {
        val original = SyncResult(
            success = true,
            timestamp = 2000L,
            syncedItems = 15,
            failedItems = 2,
            errorMessage = null
        )

        val serialized = json.encodeToString(original)
        val deserialized = json.decodeFromString<SyncResult>(serialized)

        assertEquals(original, deserialized)
    }

    // --- WatchProgressData ---

    @Test
    fun `WatchProgressData holds correct values`() {
        val data = WatchProgressData(
            mediaId = 42L,
            progress = 0.75,
            timestamp = 1000L
        )

        assertEquals(42L, data.mediaId)
        assertEquals(0.75, data.progress, 0.01)
        assertEquals(1000L, data.timestamp)
    }

    @Test
    fun `WatchProgressData serialization round-trip`() {
        val original = WatchProgressData(mediaId = 1L, progress = 0.5, timestamp = 1000L)
        val serialized = json.encodeToString(original)
        val deserialized = json.decodeFromString<WatchProgressData>(serialized)

        assertEquals(original, deserialized)
    }

    // --- FavoriteData ---

    @Test
    fun `FavoriteData holds correct values`() {
        val data = FavoriteData(mediaId = 42L, isFavorite = true)

        assertEquals(42L, data.mediaId)
        assertTrue(data.isFavorite)
    }

    @Test
    fun `FavoriteData serialization round-trip`() {
        val original = FavoriteData(mediaId = 1L, isFavorite = false)
        val serialized = json.encodeToString(original)
        val deserialized = json.decodeFromString<FavoriteData>(serialized)

        assertEquals(original, deserialized)
    }

    // --- RatingData ---

    @Test
    fun `RatingData holds correct values`() {
        val data = RatingData(mediaId = 42L, rating = 8.5)

        assertEquals(42L, data.mediaId)
        assertEquals(8.5, data.rating, 0.01)
    }

    @Test
    fun `RatingData serialization round-trip`() {
        val original = RatingData(mediaId = 1L, rating = 9.0)
        val serialized = json.encodeToString(original)
        val deserialized = json.decodeFromString<RatingData>(serialized)

        assertEquals(original, deserialized)
    }

    // --- MetadataUpdateData ---

    @Test
    fun `MetadataUpdateData holds correct values`() {
        val metadataJson = """{"title":"Updated Title","year":"2024"}"""
        val data = MetadataUpdateData(mediaId = 42L, metadata = metadataJson)

        assertEquals(42L, data.mediaId)
        assertEquals(metadataJson, data.metadata)
    }

    @Test
    fun `MetadataUpdateData serialization round-trip`() {
        val original = MetadataUpdateData(mediaId = 1L, metadata = """{"key":"value"}""")
        val serialized = json.encodeToString(original)
        val deserialized = json.decodeFromString<MetadataUpdateData>(serialized)

        assertEquals(original, deserialized)
    }

    // --- MediaDeletionData ---

    @Test
    fun `MediaDeletionData with server deletion`() {
        val data = MediaDeletionData(mediaId = 42L, localOnly = false)

        assertEquals(42L, data.mediaId)
        assertFalse(data.localOnly)
    }

    @Test
    fun `MediaDeletionData with local-only deletion`() {
        val data = MediaDeletionData(mediaId = 42L, localOnly = true)

        assertTrue(data.localOnly)
    }

    @Test
    fun `MediaDeletionData default is not local-only`() {
        val data = MediaDeletionData(mediaId = 1L)

        assertFalse(data.localOnly)
    }

    @Test
    fun `MediaDeletionData serialization round-trip`() {
        val original = MediaDeletionData(mediaId = 1L, localOnly = true)
        val serialized = json.encodeToString(original)
        val deserialized = json.decodeFromString<MediaDeletionData>(serialized)

        assertEquals(original, deserialized)
    }
}
