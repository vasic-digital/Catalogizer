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
 * Instrumentation tests for DownloadDao
 */
@RunWith(AndroidJUnit4::class)
class DownloadDaoTest {

    private lateinit var database: CatalogizerDatabase
    private lateinit var downloadDao: DownloadDao

    @Before
    fun setup() {
        val context = ApplicationProvider.getApplicationContext<Context>()
        database = Room.inMemoryDatabaseBuilder(
            context,
            CatalogizerDatabase::class.java
        ).allowMainThreadQueries().build()
        downloadDao = database.downloadDao()
    }

    @After
    fun teardown() {
        database.close()
    }

    private fun createDownloadItem(
        mediaId: Long = 1,
        title: String = "Test Download",
        status: DownloadStatus = DownloadStatus.PENDING,
        progress: Float = 0f
    ): DownloadItem {
        return DownloadItem(
            mediaId = mediaId,
            title = title,
            coverImage = "https://example.com/cover.jpg",
            downloadUrl = "https://example.com/download/file.mp4",
            localPath = null,
            progress = progress,
            status = status
        )
    }

    @Test
    fun insertAndRetrieveDownload() = runBlocking {
        // Given
        val download = createDownloadItem()

        // When
        downloadDao.insertDownload(download)
        val retrieved = downloadDao.getDownloadByMediaId(1)

        // Then
        assertNotNull(retrieved)
        assertEquals(download.title, retrieved?.title)
        assertEquals(download.status, retrieved?.status)
    }

    @Test
    fun getAllDownloads() = runBlocking {
        // Given
        val downloads = listOf(
            createDownloadItem(mediaId = 1, title = "Download 1"),
            createDownloadItem(mediaId = 2, title = "Download 2"),
            createDownloadItem(mediaId = 3, title = "Download 3")
        )
        downloads.forEach { downloadDao.insertDownload(it) }

        // When
        val allDownloads = downloadDao.getAllDownloads().first()

        // Then
        assertEquals(3, allDownloads.size)
    }

    @Test
    fun getDownloadsByStatus() = runBlocking {
        // Given
        val downloads = listOf(
            createDownloadItem(mediaId = 1, status = DownloadStatus.PENDING),
            createDownloadItem(mediaId = 2, status = DownloadStatus.DOWNLOADING),
            createDownloadItem(mediaId = 3, status = DownloadStatus.COMPLETED),
            createDownloadItem(mediaId = 4, status = DownloadStatus.DOWNLOADING)
        )
        downloads.forEach { downloadDao.insertDownload(it) }

        // When
        val downloading = downloadDao.getDownloadsByStatus(DownloadStatus.DOWNLOADING).first()

        // Then
        assertEquals(2, downloading.size)
        assertTrue(downloading.all { it.status == DownloadStatus.DOWNLOADING })
    }

    @Test
    fun updateDownloadProgress() = runBlocking {
        // Given
        val download = createDownloadItem(status = DownloadStatus.PENDING, progress = 0f)
        downloadDao.insertDownload(download)

        // When
        downloadDao.updateDownloadProgress(
            mediaId = 1,
            progress = 0.5f,
            status = DownloadStatus.DOWNLOADING
        )
        val retrieved = downloadDao.getDownloadByMediaId(1)

        // Then
        assertNotNull(retrieved)
        assertEquals(0.5f, retrieved!!.progress)
        assertEquals(DownloadStatus.DOWNLOADING, retrieved.status)
    }

    @Test
    fun updateDownload() = runBlocking {
        // Given
        val download = createDownloadItem()
        downloadDao.insertDownload(download)

        // When
        val updated = download.copy(
            localPath = "/storage/downloads/file.mp4",
            progress = 1f,
            status = DownloadStatus.COMPLETED
        )
        downloadDao.updateDownload(updated)
        val retrieved = downloadDao.getDownloadByMediaId(1)

        // Then
        assertNotNull(retrieved)
        assertEquals("/storage/downloads/file.mp4", retrieved!!.localPath)
        assertEquals(1f, retrieved.progress)
        assertEquals(DownloadStatus.COMPLETED, retrieved.status)
    }

    @Test
    fun deleteDownload() = runBlocking {
        // Given
        val download = createDownloadItem()
        downloadDao.insertDownload(download)

        // When
        downloadDao.deleteDownload(download)
        val retrieved = downloadDao.getDownloadByMediaId(1)

        // Then
        assertNull(retrieved)
    }

    @Test
    fun deleteDownloadByMediaId() = runBlocking {
        // Given
        val downloads = listOf(
            createDownloadItem(mediaId = 1),
            createDownloadItem(mediaId = 2)
        )
        downloads.forEach { downloadDao.insertDownload(it) }

        // When
        downloadDao.deleteDownloadByMediaId(1)

        // Then
        assertNull(downloadDao.getDownloadByMediaId(1))
        assertNotNull(downloadDao.getDownloadByMediaId(2))
    }

    @Test
    fun deleteDownloadsByStatus() = runBlocking {
        // Given
        val downloads = listOf(
            createDownloadItem(mediaId = 1, status = DownloadStatus.FAILED),
            createDownloadItem(mediaId = 2, status = DownloadStatus.COMPLETED),
            createDownloadItem(mediaId = 3, status = DownloadStatus.FAILED)
        )
        downloads.forEach { downloadDao.insertDownload(it) }

        // When
        downloadDao.deleteDownloadsByStatus(DownloadStatus.FAILED)
        val remaining = downloadDao.getAllDownloads().first()

        // Then
        assertEquals(1, remaining.size)
        assertEquals(DownloadStatus.COMPLETED, remaining[0].status)
    }

    @Test
    fun downloadStatusTransitions() = runBlocking {
        // Given
        val download = createDownloadItem(status = DownloadStatus.PENDING)
        downloadDao.insertDownload(download)

        // Test PENDING -> DOWNLOADING
        downloadDao.updateDownloadProgress(1, 0.1f, DownloadStatus.DOWNLOADING)
        var retrieved = downloadDao.getDownloadByMediaId(1)
        assertEquals(DownloadStatus.DOWNLOADING, retrieved?.status)

        // Test DOWNLOADING -> PAUSED
        downloadDao.updateDownloadProgress(1, 0.5f, DownloadStatus.PAUSED)
        retrieved = downloadDao.getDownloadByMediaId(1)
        assertEquals(DownloadStatus.PAUSED, retrieved?.status)

        // Test PAUSED -> DOWNLOADING
        downloadDao.updateDownloadProgress(1, 0.5f, DownloadStatus.DOWNLOADING)
        retrieved = downloadDao.getDownloadByMediaId(1)
        assertEquals(DownloadStatus.DOWNLOADING, retrieved?.status)

        // Test DOWNLOADING -> COMPLETED
        downloadDao.updateDownloadProgress(1, 1f, DownloadStatus.COMPLETED)
        retrieved = downloadDao.getDownloadByMediaId(1)
        assertEquals(DownloadStatus.COMPLETED, retrieved?.status)
        assertEquals(1f, retrieved?.progress)
    }

    @Test
    fun downloadFailureHandling() = runBlocking {
        // Given
        val download = createDownloadItem(status = DownloadStatus.DOWNLOADING, progress = 0.3f)
        downloadDao.insertDownload(download)

        // When - simulate failure
        downloadDao.updateDownloadProgress(1, 0.3f, DownloadStatus.FAILED)
        val retrieved = downloadDao.getDownloadByMediaId(1)

        // Then
        assertNotNull(retrieved)
        assertEquals(DownloadStatus.FAILED, retrieved!!.status)
        assertEquals(0.3f, retrieved.progress) // Progress preserved
    }

    @Test
    fun cancelledDownload() = runBlocking {
        // Given
        val download = createDownloadItem(status = DownloadStatus.DOWNLOADING, progress = 0.5f)
        downloadDao.insertDownload(download)

        // When
        downloadDao.updateDownloadProgress(1, 0f, DownloadStatus.CANCELLED)
        val retrieved = downloadDao.getDownloadByMediaId(1)

        // Then
        assertNotNull(retrieved)
        assertEquals(DownloadStatus.CANCELLED, retrieved!!.status)
    }

    @Test
    fun getAllDownloadsOrderedByCreatedAt() = runBlocking {
        // Given
        val baseTime = System.currentTimeMillis()
        val downloads = listOf(
            DownloadItem(
                mediaId = 1,
                title = "Old Download",
                coverImage = null,
                downloadUrl = "url1",
                localPath = null,
                createdAt = baseTime - 2000
            ),
            DownloadItem(
                mediaId = 2,
                title = "New Download",
                coverImage = null,
                downloadUrl = "url2",
                localPath = null,
                createdAt = baseTime
            ),
            DownloadItem(
                mediaId = 3,
                title = "Middle Download",
                coverImage = null,
                downloadUrl = "url3",
                localPath = null,
                createdAt = baseTime - 1000
            )
        )
        downloads.forEach { downloadDao.insertDownload(it) }

        // When
        val allDownloads = downloadDao.getAllDownloads().first()

        // Then - should be ordered by created_at DESC
        assertEquals(3, allDownloads.size)
        assertEquals("New Download", allDownloads[0].title)
        assertEquals("Middle Download", allDownloads[1].title)
        assertEquals("Old Download", allDownloads[2].title)
    }
}
