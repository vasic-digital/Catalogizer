package com.catalogizer.androidtv.data.models

import org.junit.Assert.*
import org.junit.Test

class SortOptionsTest {

    // --- SortOption ---

    @Test
    fun `SortOption fromValue returns TITLE`() {
        assertEquals(SortOption.TITLE, SortOption.fromValue("title"))
    }

    @Test
    fun `SortOption fromValue returns DURATION`() {
        assertEquals(SortOption.DURATION, SortOption.fromValue("duration"))
    }

    @Test
    fun `SortOption fromValue defaults to UPDATED_AT`() {
        assertEquals(SortOption.UPDATED_AT, SortOption.fromValue("unknown"))
    }

    @Test
    fun `SortOption has 7 values including DURATION`() {
        assertEquals(7, SortOption.values().size)
    }

    @Test
    fun `SortOption display names`() {
        assertEquals("Title", SortOption.TITLE.displayName)
        assertEquals("Year", SortOption.YEAR.displayName)
        assertEquals("Rating", SortOption.RATING.displayName)
        assertEquals("Duration", SortOption.DURATION.displayName)
        assertEquals("Recently Updated", SortOption.UPDATED_AT.displayName)
        assertEquals("Recently Added", SortOption.CREATED_AT.displayName)
        assertEquals("File Size", SortOption.FILE_SIZE.displayName)
    }

    // --- SortOrder ---

    @Test
    fun `SortOrder fromValue returns ASC`() {
        assertEquals(SortOrder.ASC, SortOrder.fromValue("asc"))
    }

    @Test
    fun `SortOrder fromValue returns DESC`() {
        assertEquals(SortOrder.DESC, SortOrder.fromValue("desc"))
    }

    @Test
    fun `SortOrder fromValue defaults to DESC`() {
        assertEquals(SortOrder.DESC, SortOrder.fromValue("unknown"))
    }

    @Test
    fun `SortOrder display names`() {
        assertEquals("Ascending", SortOrder.ASC.displayName)
        assertEquals("Descending", SortOrder.DESC.displayName)
    }

    // --- QualityLevel ---

    @Test
    fun `QualityLevel fromValue returns HD_1080P`() {
        assertEquals(QualityLevel.HD_1080P, QualityLevel.fromValue("1080p"))
    }

    @Test
    fun `QualityLevel fromValue returns UHD_4K`() {
        assertEquals(QualityLevel.UHD_4K, QualityLevel.fromValue("4k"))
    }

    @Test
    fun `QualityLevel fromValue returns null for unknown`() {
        assertNull(QualityLevel.fromValue("unknown"))
    }

    @Test
    fun `QualityLevel getAllQualities returns 9 levels`() {
        assertEquals(9, QualityLevel.getAllQualities().size)
    }
}
