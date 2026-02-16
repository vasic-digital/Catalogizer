package com.catalogizer.android.data.local

import org.junit.Assert.*
import org.junit.Test

class DownloadItemTest {

    @Test
    fun `DownloadItem has correct defaults`() {
        val item = DownloadItem(
            mediaId = 1L,
            title = "Test Movie",
            coverImage = null,
            downloadUrl = "http://example.com/download/1",
            localPath = null
        )

        assertEquals(1L, item.mediaId)
        assertEquals("Test Movie", item.title)
        assertNull(item.coverImage)
        assertEquals("http://example.com/download/1", item.downloadUrl)
        assertNull(item.localPath)
        assertEquals(0f, item.progress, 0.01f)
        assertEquals(DownloadStatus.PENDING, item.status)
        assertTrue(item.createdAt > 0)
        assertTrue(item.updatedAt > 0)
    }

    @Test
    fun `DownloadItem with all fields populated`() {
        val item = DownloadItem(
            mediaId = 42L,
            title = "Complete Movie",
            coverImage = "http://example.com/cover.jpg",
            downloadUrl = "http://example.com/download/42",
            localPath = "/data/downloads/movie.mkv",
            progress = 1.0f,
            status = DownloadStatus.COMPLETED,
            createdAt = 1000L,
            updatedAt = 2000L
        )

        assertEquals(42L, item.mediaId)
        assertEquals("Complete Movie", item.title)
        assertEquals("http://example.com/cover.jpg", item.coverImage)
        assertEquals("/data/downloads/movie.mkv", item.localPath)
        assertEquals(1.0f, item.progress, 0.01f)
        assertEquals(DownloadStatus.COMPLETED, item.status)
        assertEquals(1000L, item.createdAt)
        assertEquals(2000L, item.updatedAt)
    }

    @Test
    fun `DownloadItem copy updates correctly`() {
        val original = DownloadItem(
            mediaId = 1L,
            title = "Test",
            coverImage = null,
            downloadUrl = "http://example.com/1",
            localPath = null,
            progress = 0.0f,
            status = DownloadStatus.PENDING
        )

        val downloading = original.copy(progress = 0.5f, status = DownloadStatus.DOWNLOADING)
        val completed = downloading.copy(progress = 1.0f, status = DownloadStatus.COMPLETED, localPath = "/downloaded/file.mkv")

        assertEquals(DownloadStatus.PENDING, original.status)
        assertEquals(0.5f, downloading.progress, 0.01f)
        assertEquals(DownloadStatus.DOWNLOADING, downloading.status)
        assertEquals(1.0f, completed.progress, 0.01f)
        assertEquals(DownloadStatus.COMPLETED, completed.status)
        assertEquals("/downloaded/file.mkv", completed.localPath)
    }

    @Test
    fun `DownloadStatus has all expected values`() {
        val statuses = DownloadStatus.values()
        assertEquals(6, statuses.size)
        assertTrue(statuses.contains(DownloadStatus.PENDING))
        assertTrue(statuses.contains(DownloadStatus.DOWNLOADING))
        assertTrue(statuses.contains(DownloadStatus.COMPLETED))
        assertTrue(statuses.contains(DownloadStatus.FAILED))
        assertTrue(statuses.contains(DownloadStatus.PAUSED))
        assertTrue(statuses.contains(DownloadStatus.CANCELLED))
    }

    @Test
    fun `DownloadItem equality works correctly`() {
        val ts = System.currentTimeMillis()
        val item1 = DownloadItem(mediaId = 1L, title = "Test", coverImage = null, downloadUrl = "url", localPath = null, createdAt = ts, updatedAt = ts)
        val item2 = DownloadItem(mediaId = 1L, title = "Test", coverImage = null, downloadUrl = "url", localPath = null, createdAt = ts, updatedAt = ts)
        val item3 = DownloadItem(mediaId = 2L, title = "Test", coverImage = null, downloadUrl = "url", localPath = null, createdAt = ts, updatedAt = ts)

        assertEquals(item1, item2)
        assertNotEquals(item1, item3)
    }
}
