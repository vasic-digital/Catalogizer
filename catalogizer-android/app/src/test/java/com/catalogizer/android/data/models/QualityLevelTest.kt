package com.catalogizer.android.data.models

import org.junit.Assert.*
import org.junit.Test

class QualityLevelTest {

    @Test
    fun `fromValue returns correct quality for 1080p`() {
        assertEquals(QualityLevel.HD_1080P, QualityLevel.fromValue("1080p"))
    }

    @Test
    fun `fromValue returns correct quality for 720p`() {
        assertEquals(QualityLevel.HD_720P, QualityLevel.fromValue("720p"))
    }

    @Test
    fun `fromValue returns correct quality for 4k`() {
        assertEquals(QualityLevel.UHD_4K, QualityLevel.fromValue("4k"))
    }

    @Test
    fun `fromValue returns null for unknown quality`() {
        assertNull(QualityLevel.fromValue("unknown"))
    }

    @Test
    fun `fromValue returns null for empty string`() {
        assertNull(QualityLevel.fromValue(""))
    }

    @Test
    fun `getAllQualities returns all 9 quality levels`() {
        assertEquals(9, QualityLevel.getAllQualities().size)
    }

    @Test
    fun `quality display names are human readable`() {
        assertEquals("CAM", QualityLevel.CAM.displayName)
        assertEquals("720p HD", QualityLevel.HD_720P.displayName)
        assertEquals("1080p HD", QualityLevel.HD_1080P.displayName)
        assertEquals("4K UHD", QualityLevel.UHD_4K.displayName)
        assertEquals("HDR", QualityLevel.HDR.displayName)
        assertEquals("Dolby Vision", QualityLevel.DOLBY_VISION.displayName)
    }
}
