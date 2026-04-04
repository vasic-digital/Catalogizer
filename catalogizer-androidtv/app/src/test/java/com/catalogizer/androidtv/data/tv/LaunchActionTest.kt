package com.catalogizer.androidtv.data.tv

import org.junit.Assert.*
import org.junit.Test

class LaunchActionTest {

    @Test
    fun `LaunchAction has DETAIL and IMMEDIATE_PLAY values`() {
        assertEquals(2, LaunchAction.values().size)
        assertNotNull(LaunchAction.DETAIL)
        assertNotNull(LaunchAction.IMMEDIATE_PLAY)
    }

    @Test
    fun `LaunchAction fromString returns correct value`() {
        assertEquals(LaunchAction.DETAIL, LaunchAction.fromString("DETAIL"))
        assertEquals(LaunchAction.IMMEDIATE_PLAY, LaunchAction.fromString("IMMEDIATE_PLAY"))
    }

    @Test
    fun `LaunchAction fromString returns DETAIL for unknown input`() {
        assertEquals(LaunchAction.DETAIL, LaunchAction.fromString("UNKNOWN"))
        assertEquals(LaunchAction.DETAIL, LaunchAction.fromString(""))
    }
}
