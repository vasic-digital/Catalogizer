package com.catalogizer.android.data.models

import org.junit.Assert.*
import org.junit.Test

class MediaEnumsTest {

    // --- MediaType ---

    @Test
    fun `MediaType fromValue returns correct type for known values`() {
        assertEquals(MediaType.MOVIE, MediaType.fromValue("movie"))
        assertEquals(MediaType.TV_SHOW, MediaType.fromValue("tv_show"))
        assertEquals(MediaType.DOCUMENTARY, MediaType.fromValue("documentary"))
        assertEquals(MediaType.ANIME, MediaType.fromValue("anime"))
        assertEquals(MediaType.MUSIC, MediaType.fromValue("music"))
        assertEquals(MediaType.AUDIOBOOK, MediaType.fromValue("audiobook"))
        assertEquals(MediaType.PODCAST, MediaType.fromValue("podcast"))
        assertEquals(MediaType.GAME, MediaType.fromValue("game"))
        assertEquals(MediaType.SOFTWARE, MediaType.fromValue("software"))
        assertEquals(MediaType.EBOOK, MediaType.fromValue("ebook"))
    }

    @Test
    fun `MediaType fromValue returns OTHER for unknown values`() {
        assertEquals(MediaType.OTHER, MediaType.fromValue("unknown"))
        assertEquals(MediaType.OTHER, MediaType.fromValue(""))
        assertEquals(MediaType.OTHER, MediaType.fromValue("xyz"))
    }

    @Test
    fun `MediaType getAllTypes returns all types`() {
        val allTypes = MediaType.getAllTypes()
        assertEquals(MediaType.values().size, allTypes.size)
        assertTrue(allTypes.contains(MediaType.MOVIE))
        assertTrue(allTypes.contains(MediaType.OTHER))
    }

    @Test
    fun `MediaType has correct display names`() {
        assertEquals("Movies", MediaType.MOVIE.displayName)
        assertEquals("TV Shows", MediaType.TV_SHOW.displayName)
        assertEquals("Music", MediaType.MUSIC.displayName)
        assertEquals("E-books", MediaType.EBOOK.displayName)
        assertEquals("Other", MediaType.OTHER.displayName)
    }

    @Test
    fun `MediaType has correct values`() {
        assertEquals("movie", MediaType.MOVIE.value)
        assertEquals("tv_show", MediaType.TV_SHOW.value)
        assertEquals("youtube_video", MediaType.YOUTUBE_VIDEO.value)
    }

    // --- QualityLevel ---

    @Test
    fun `QualityLevel fromValue returns correct quality for known values`() {
        assertEquals(QualityLevel.CAM, QualityLevel.fromValue("cam"))
        assertEquals(QualityLevel.HD_720P, QualityLevel.fromValue("720p"))
        assertEquals(QualityLevel.HD_1080P, QualityLevel.fromValue("1080p"))
        assertEquals(QualityLevel.UHD_4K, QualityLevel.fromValue("4k"))
        assertEquals(QualityLevel.HDR, QualityLevel.fromValue("hdr"))
        assertEquals(QualityLevel.DOLBY_VISION, QualityLevel.fromValue("dolby_vision"))
    }

    @Test
    fun `QualityLevel fromValue returns null for unknown values`() {
        assertNull(QualityLevel.fromValue("unknown"))
        assertNull(QualityLevel.fromValue(""))
        assertNull(QualityLevel.fromValue("8k"))
    }

    @Test
    fun `QualityLevel getAllQualities returns all quality levels`() {
        val allQualities = QualityLevel.getAllQualities()
        assertEquals(QualityLevel.values().size, allQualities.size)
    }

    @Test
    fun `QualityLevel has correct display names`() {
        assertEquals("CAM", QualityLevel.CAM.displayName)
        assertEquals("720p HD", QualityLevel.HD_720P.displayName)
        assertEquals("1080p HD", QualityLevel.HD_1080P.displayName)
        assertEquals("4K UHD", QualityLevel.UHD_4K.displayName)
        assertEquals("Dolby Vision", QualityLevel.DOLBY_VISION.displayName)
    }

    // --- SortOption ---

    @Test
    fun `SortOption fromValue returns correct option for known values`() {
        assertEquals(SortOption.TITLE, SortOption.fromValue("title"))
        assertEquals(SortOption.YEAR, SortOption.fromValue("year"))
        assertEquals(SortOption.RATING, SortOption.fromValue("rating"))
        assertEquals(SortOption.UPDATED_AT, SortOption.fromValue("updated_at"))
        assertEquals(SortOption.CREATED_AT, SortOption.fromValue("created_at"))
        assertEquals(SortOption.FILE_SIZE, SortOption.fromValue("file_size"))
    }

    @Test
    fun `SortOption fromValue returns UPDATED_AT for unknown values`() {
        assertEquals(SortOption.UPDATED_AT, SortOption.fromValue("unknown"))
        assertEquals(SortOption.UPDATED_AT, SortOption.fromValue(""))
    }

    @Test
    fun `SortOption has correct display names`() {
        assertEquals("Title", SortOption.TITLE.displayName)
        assertEquals("Rating", SortOption.RATING.displayName)
        assertEquals("Recently Updated", SortOption.UPDATED_AT.displayName)
        assertEquals("Recently Added", SortOption.CREATED_AT.displayName)
        assertEquals("File Size", SortOption.FILE_SIZE.displayName)
    }

    // --- SortOrder ---

    @Test
    fun `SortOrder fromValue returns correct order for known values`() {
        assertEquals(SortOrder.ASC, SortOrder.fromValue("asc"))
        assertEquals(SortOrder.DESC, SortOrder.fromValue("desc"))
    }

    @Test
    fun `SortOrder fromValue returns DESC for unknown values`() {
        assertEquals(SortOrder.DESC, SortOrder.fromValue("unknown"))
        assertEquals(SortOrder.DESC, SortOrder.fromValue(""))
    }

    @Test
    fun `SortOrder has correct display names`() {
        assertEquals("Ascending", SortOrder.ASC.displayName)
        assertEquals("Descending", SortOrder.DESC.displayName)
    }

    // --- UserRole ---

    @Test
    fun `UserRole fromValue returns correct role for known values`() {
        assertEquals(UserRole.ADMIN, UserRole.fromValue("admin"))
        assertEquals(UserRole.MODERATOR, UserRole.fromValue("moderator"))
        assertEquals(UserRole.USER, UserRole.fromValue("user"))
        assertEquals(UserRole.VIEWER, UserRole.fromValue("viewer"))
    }

    @Test
    fun `UserRole fromValue returns USER for unknown values`() {
        assertEquals(UserRole.USER, UserRole.fromValue("unknown"))
        assertEquals(UserRole.USER, UserRole.fromValue(""))
    }

    @Test
    fun `UserRole has correct display names`() {
        assertEquals("Administrator", UserRole.ADMIN.displayName)
        assertEquals("Moderator", UserRole.MODERATOR.displayName)
        assertEquals("User", UserRole.USER.displayName)
        assertEquals("Viewer", UserRole.VIEWER.displayName)
    }

    // --- Permissions ---

    @Test
    fun `Permissions constants are correct`() {
        assertEquals("read:media", Permissions.READ_MEDIA)
        assertEquals("write:media", Permissions.WRITE_MEDIA)
        assertEquals("delete:media", Permissions.DELETE_MEDIA)
        assertEquals("read:catalog", Permissions.READ_CATALOG)
        assertEquals("write:catalog", Permissions.WRITE_CATALOG)
        assertEquals("delete:catalog", Permissions.DELETE_CATALOG)
        assertEquals("trigger:analysis", Permissions.TRIGGER_ANALYSIS)
        assertEquals("view:analysis", Permissions.VIEW_ANALYSIS)
        assertEquals("manage:users", Permissions.MANAGE_USERS)
        assertEquals("manage:roles", Permissions.MANAGE_ROLES)
        assertEquals("view:logs", Permissions.VIEW_LOGS)
        assertEquals("admin:system", Permissions.SYSTEM_ADMIN)
        assertEquals("access:api", Permissions.API_ACCESS)
        assertEquals("write:api", Permissions.API_WRITE)
    }
}
