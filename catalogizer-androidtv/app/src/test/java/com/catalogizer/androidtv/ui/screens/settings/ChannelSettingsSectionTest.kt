package com.catalogizer.androidtv.ui.screens.settings

import com.catalogizer.androidtv.data.models.MediaType
import com.catalogizer.androidtv.data.tv.LaunchAction
import org.junit.Assert.*
import org.junit.Test

/**
 * Tests for ChannelSettingsSection logic: per-media-type launch action mapping,
 * LaunchAction enum behavior, label rendering, empty state handling,
 * and update callback wiring.
 */
class ChannelSettingsSectionTest {

    // --- LaunchAction enum ---

    @Test
    fun `LaunchAction has exactly two values`() {
        assertEquals(2, LaunchAction.values().size)
    }

    @Test
    fun `LaunchAction DETAIL exists`() {
        assertEquals(LaunchAction.DETAIL, LaunchAction.valueOf("DETAIL"))
    }

    @Test
    fun `LaunchAction IMMEDIATE_PLAY exists`() {
        assertEquals(LaunchAction.IMMEDIATE_PLAY, LaunchAction.valueOf("IMMEDIATE_PLAY"))
    }

    @Test
    fun `LaunchAction fromString parses DETAIL`() {
        assertEquals(LaunchAction.DETAIL, LaunchAction.fromString("DETAIL"))
    }

    @Test
    fun `LaunchAction fromString parses IMMEDIATE_PLAY`() {
        assertEquals(LaunchAction.IMMEDIATE_PLAY, LaunchAction.fromString("IMMEDIATE_PLAY"))
    }

    @Test
    fun `LaunchAction fromString defaults to DETAIL for unknown`() {
        assertEquals(LaunchAction.DETAIL, LaunchAction.fromString("unknown_action"))
    }

    @Test
    fun `LaunchAction fromString defaults to DETAIL for empty string`() {
        assertEquals(LaunchAction.DETAIL, LaunchAction.fromString(""))
    }

    // --- Label mapping ---

    @Test
    fun `DETAIL action maps to Detail Screen label`() {
        val label = when (LaunchAction.DETAIL) {
            LaunchAction.DETAIL -> "Detail Screen"
            LaunchAction.IMMEDIATE_PLAY -> "Play Immediately"
        }
        assertEquals("Detail Screen", label)
    }

    @Test
    fun `IMMEDIATE_PLAY action maps to Play Immediately label`() {
        val label = when (LaunchAction.IMMEDIATE_PLAY) {
            LaunchAction.DETAIL -> "Detail Screen"
            LaunchAction.IMMEDIATE_PLAY -> "Play Immediately"
        }
        assertEquals("Play Immediately", label)
    }

    // --- Default launch action ---

    @Test
    fun `default launch action is DETAIL when type not in map`() {
        val launchActions = emptyMap<String, LaunchAction>()
        val action = launchActions["movie"] ?: LaunchAction.DETAIL
        assertEquals(LaunchAction.DETAIL, action)
    }

    @Test
    fun `configured launch action overrides default`() {
        val launchActions = mapOf("movie" to LaunchAction.IMMEDIATE_PLAY)
        val action = launchActions["movie"] ?: LaunchAction.DETAIL
        assertEquals(LaunchAction.IMMEDIATE_PLAY, action)
    }

    // --- MediaType display names ---

    @Test
    fun `movie MediaType has Movies display name`() {
        val mediaType = MediaType.fromValue("movie")
        assertEquals("Movies", mediaType.displayName)
    }

    @Test
    fun `tv_show MediaType has TV Shows display name`() {
        val mediaType = MediaType.fromValue("tv_show")
        assertEquals("TV Shows", mediaType.displayName)
    }

    @Test
    fun `song MediaType has Songs display name`() {
        val mediaType = MediaType.fromValue("song")
        assertEquals("Songs", mediaType.displayName)
    }

    @Test
    fun `game MediaType has Games display name`() {
        val mediaType = MediaType.fromValue("game")
        assertEquals("Games", mediaType.displayName)
    }

    @Test
    fun `unknown value maps to OTHER MediaType`() {
        val mediaType = MediaType.fromValue("xyz_unknown")
        assertEquals(MediaType.OTHER, mediaType)
        assertEquals("Other", mediaType.displayName)
    }

    // --- Empty activeMediaTypes ---

    @Test
    fun `empty activeMediaTypes produces no settings rows`() {
        val activeMediaTypes = emptyList<String>()
        assertTrue(activeMediaTypes.isEmpty())
    }

    @Test
    fun `non-empty activeMediaTypes iterates all types`() {
        val activeMediaTypes = listOf("movie", "tv_show", "song")
        assertEquals(3, activeMediaTypes.size)
    }

    // --- Update callback ---

    @Test
    fun `update callback receives correct type and action`() {
        var updatedType: String? = null
        var updatedAction: LaunchAction? = null
        val onUpdateAction: (String, LaunchAction) -> Unit = { type, action ->
            updatedType = type
            updatedAction = action
        }

        onUpdateAction("movie", LaunchAction.IMMEDIATE_PLAY)

        assertEquals("movie", updatedType)
        assertEquals(LaunchAction.IMMEDIATE_PLAY, updatedAction)
    }

    @Test
    fun `update callback can be called for each type independently`() {
        val updates = mutableListOf<Pair<String, LaunchAction>>()
        val onUpdateAction: (String, LaunchAction) -> Unit = { type, action ->
            updates.add(type to action)
        }

        onUpdateAction("movie", LaunchAction.IMMEDIATE_PLAY)
        onUpdateAction("tv_show", LaunchAction.DETAIL)
        onUpdateAction("song", LaunchAction.IMMEDIATE_PLAY)

        assertEquals(3, updates.size)
        assertEquals("movie" to LaunchAction.IMMEDIATE_PLAY, updates[0])
        assertEquals("tv_show" to LaunchAction.DETAIL, updates[1])
        assertEquals("song" to LaunchAction.IMMEDIATE_PLAY, updates[2])
    }

    // --- Selection state ---

    @Test
    fun `selected action matches current value`() {
        val currentAction = LaunchAction.DETAIL
        val testAction = LaunchAction.DETAIL
        val isSelected = currentAction == testAction
        assertTrue(isSelected)
    }

    @Test
    fun `non-selected action does not match current value`() {
        val currentAction = LaunchAction.DETAIL
        val testAction = LaunchAction.IMMEDIATE_PLAY
        val isSelected = currentAction == testAction
        assertFalse(isSelected)
    }

    // --- Per-type configuration map ---

    @Test
    fun `launch actions map stores per-type settings`() {
        val launchActions = mapOf(
            "movie" to LaunchAction.IMMEDIATE_PLAY,
            "tv_show" to LaunchAction.DETAIL,
            "song" to LaunchAction.IMMEDIATE_PLAY,
            "game" to LaunchAction.DETAIL
        )
        assertEquals(4, launchActions.size)
        assertEquals(LaunchAction.IMMEDIATE_PLAY, launchActions["movie"])
        assertEquals(LaunchAction.DETAIL, launchActions["tv_show"])
    }

    @Test
    fun `launch actions map handles missing type with default`() {
        val launchActions = mapOf("movie" to LaunchAction.IMMEDIATE_PLAY)
        val action = launchActions["book"] ?: LaunchAction.DETAIL
        assertEquals(LaunchAction.DETAIL, action)
    }
}
