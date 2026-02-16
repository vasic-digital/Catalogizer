package com.catalogizer.android.data.local

import com.catalogizer.android.data.sync.SyncOperation
import com.catalogizer.android.data.sync.SyncOperationType
import org.junit.Assert.*
import org.junit.Test

class SyncOperationDaoTest {

    @Test
    fun `SyncOperation creates with correct defaults`() {
        val operation = SyncOperation(
            type = SyncOperationType.UPDATE_PROGRESS,
            mediaId = 1L,
            data = """{"progress": 0.5}""",
            timestamp = System.currentTimeMillis()
        )

        assertEquals(0L, operation.id)
        assertEquals(SyncOperationType.UPDATE_PROGRESS, operation.type)
        assertEquals(1L, operation.mediaId)
        assertEquals(0, operation.retryCount)
        assertEquals(3, operation.maxRetries)
    }

    @Test
    fun `SyncOperation with custom retry values`() {
        val operation = SyncOperation(
            type = SyncOperationType.TOGGLE_FAVORITE,
            mediaId = 42L,
            data = """{"isFavorite": true}""",
            timestamp = 1000L,
            retryCount = 2,
            maxRetries = 5
        )

        assertEquals(2, operation.retryCount)
        assertEquals(5, operation.maxRetries)
    }

    @Test
    fun `SyncOperation equality works correctly`() {
        val ts = 1000L
        val op1 = SyncOperation(id = 1, type = SyncOperationType.UPDATE_PROGRESS, mediaId = 1, data = "test", timestamp = ts)
        val op2 = SyncOperation(id = 1, type = SyncOperationType.UPDATE_PROGRESS, mediaId = 1, data = "test", timestamp = ts)
        val op3 = SyncOperation(id = 2, type = SyncOperationType.UPDATE_PROGRESS, mediaId = 1, data = "test", timestamp = ts)

        assertEquals(op1, op2)
        assertNotEquals(op1, op3)
    }

    @Test
    fun `SyncOperation copy updates correctly`() {
        val original = SyncOperation(
            type = SyncOperationType.UPLOAD_RATING,
            mediaId = 1L,
            data = """{"rating": 8.5}""",
            timestamp = 1000L
        )
        val retried = original.copy(retryCount = original.retryCount + 1)

        assertEquals(0, original.retryCount)
        assertEquals(1, retried.retryCount)
    }

    @Test
    fun `SyncOperation with null data`() {
        val operation = SyncOperation(
            type = SyncOperationType.DELETE_MEDIA,
            mediaId = 1L,
            data = null,
            timestamp = 1000L
        )

        assertNull(operation.data)
    }

    @Test
    fun `SyncOperationType has all expected values`() {
        val types = SyncOperationType.values()
        assertEquals(5, types.size)
        assertTrue(types.contains(SyncOperationType.UPDATE_PROGRESS))
        assertTrue(types.contains(SyncOperationType.TOGGLE_FAVORITE))
        assertTrue(types.contains(SyncOperationType.UPLOAD_RATING))
        assertTrue(types.contains(SyncOperationType.UPDATE_METADATA))
        assertTrue(types.contains(SyncOperationType.DELETE_MEDIA))
    }
}
