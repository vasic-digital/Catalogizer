package com.catalogizer.android.data.sync

import org.junit.Assert.*
import org.junit.Test

class SyncOperationTest {

    @Test
    fun `SyncOperation should have correct default values`() {
        val operation = SyncOperation(
            type = SyncOperationType.UPDATE_PROGRESS,
            mediaId = 1L,
            data = """{"mediaId":1,"progress":0.5,"timestamp":1234567890}""",
            timestamp = 1234567890L
        )

        assertEquals(0L, operation.id)
        assertEquals(SyncOperationType.UPDATE_PROGRESS, operation.type)
        assertEquals(1L, operation.mediaId)
        assertNotNull(operation.data)
        assertEquals(0, operation.retryCount)
        assertEquals(3, operation.maxRetries)
    }

    @Test
    fun `SyncOperation with custom retry values should retain them`() {
        val operation = SyncOperation(
            id = 42L,
            type = SyncOperationType.TOGGLE_FAVORITE,
            mediaId = 5L,
            data = """{"mediaId":5,"isFavorite":true}""",
            timestamp = System.currentTimeMillis(),
            retryCount = 2,
            maxRetries = 5
        )

        assertEquals(42L, operation.id)
        assertEquals(2, operation.retryCount)
        assertEquals(5, operation.maxRetries)
    }

    @Test
    fun `SyncOperationType should have all expected values`() {
        val types = SyncOperationType.values()

        assertEquals(5, types.size)
        assertTrue(types.contains(SyncOperationType.UPDATE_PROGRESS))
        assertTrue(types.contains(SyncOperationType.TOGGLE_FAVORITE))
        assertTrue(types.contains(SyncOperationType.UPLOAD_RATING))
        assertTrue(types.contains(SyncOperationType.UPDATE_METADATA))
        assertTrue(types.contains(SyncOperationType.DELETE_MEDIA))
    }

    @Test
    fun `SyncOperation with null data should be valid`() {
        val operation = SyncOperation(
            type = SyncOperationType.DELETE_MEDIA,
            mediaId = 10L,
            data = null,
            timestamp = System.currentTimeMillis()
        )

        assertNull(operation.data)
        assertEquals(SyncOperationType.DELETE_MEDIA, operation.type)
    }

    @Test
    fun `SyncOperation copy should create independent copy`() {
        val original = SyncOperation(
            id = 1L,
            type = SyncOperationType.UPLOAD_RATING,
            mediaId = 3L,
            data = """{"mediaId":3,"rating":4.5}""",
            timestamp = 1000L,
            retryCount = 0,
            maxRetries = 3
        )

        val copy = original.copy(retryCount = 1)

        assertEquals(0, original.retryCount)
        assertEquals(1, copy.retryCount)
        assertEquals(original.id, copy.id)
        assertEquals(original.type, copy.type)
        assertEquals(original.mediaId, copy.mediaId)
    }
}
