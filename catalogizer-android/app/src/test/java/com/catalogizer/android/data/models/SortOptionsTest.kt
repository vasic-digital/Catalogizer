package com.catalogizer.android.data.models

import org.junit.Assert.*
import org.junit.Test

class SortOptionsTest {

    // --- SortOption ---

    @Test
    fun `SortOption fromValue returns TITLE for title`() {
        assertEquals(SortOption.TITLE, SortOption.fromValue("title"))
    }

    @Test
    fun `SortOption fromValue returns YEAR for year`() {
        assertEquals(SortOption.YEAR, SortOption.fromValue("year"))
    }

    @Test
    fun `SortOption fromValue returns RATING for rating`() {
        assertEquals(SortOption.RATING, SortOption.fromValue("rating"))
    }

    @Test
    fun `SortOption fromValue defaults to UPDATED_AT for unknown`() {
        assertEquals(SortOption.UPDATED_AT, SortOption.fromValue("unknown"))
    }

    @Test
    fun `SortOption fromValue defaults to UPDATED_AT for empty string`() {
        assertEquals(SortOption.UPDATED_AT, SortOption.fromValue(""))
    }

    @Test
    fun `SortOption has 6 values`() {
        assertEquals(6, SortOption.values().size)
    }

    @Test
    fun `SortOption display names are human readable`() {
        assertEquals("Title", SortOption.TITLE.displayName)
        assertEquals("Year", SortOption.YEAR.displayName)
        assertEquals("Rating", SortOption.RATING.displayName)
        assertEquals("Recently Updated", SortOption.UPDATED_AT.displayName)
        assertEquals("Recently Added", SortOption.CREATED_AT.displayName)
        assertEquals("File Size", SortOption.FILE_SIZE.displayName)
    }

    // --- SortOrder ---

    @Test
    fun `SortOrder fromValue returns ASC for asc`() {
        assertEquals(SortOrder.ASC, SortOrder.fromValue("asc"))
    }

    @Test
    fun `SortOrder fromValue returns DESC for desc`() {
        assertEquals(SortOrder.DESC, SortOrder.fromValue("desc"))
    }

    @Test
    fun `SortOrder fromValue defaults to DESC for unknown`() {
        assertEquals(SortOrder.DESC, SortOrder.fromValue("unknown"))
    }

    @Test
    fun `SortOrder display names are correct`() {
        assertEquals("Ascending", SortOrder.ASC.displayName)
        assertEquals("Descending", SortOrder.DESC.displayName)
    }

    @Test
    fun `SortOrder has exactly 2 values`() {
        assertEquals(2, SortOrder.values().size)
    }
}
