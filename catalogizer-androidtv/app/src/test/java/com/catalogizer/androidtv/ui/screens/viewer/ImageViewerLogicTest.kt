package com.catalogizer.androidtv.ui.screens.viewer

import org.junit.Assert.assertEquals
import org.junit.Assert.assertNull
import org.junit.Test

/**
 * Pure-logic tests for the image viewer — URL resolution, sibling index math,
 * and the position label. No device, no Robolectric, no Compose: these exercise
 * the deterministic helpers that drive the screen's D-pad state transitions and
 * stream-URL resolution (§11.4.52 autonomous, §11.4.6 proven-not-guessed).
 */
class ImageViewerLogicTest {

    // --- resolveStreamUrl: mirrors the player's relative/absolute handling ---

    @Test
    fun `relative stream path is prefixed with base url`() {
        assertEquals(
            "http://host:8080/api/v1/stream/42",
            resolveStreamUrl("/api/v1/stream/42", "http://host:8080")
        )
    }

    @Test
    fun `base url trailing slash is not doubled`() {
        assertEquals(
            "http://host:8080/api/v1/stream/42",
            resolveStreamUrl("/api/v1/stream/42", "http://host:8080/")
        )
    }

    @Test
    fun `absolute stream url is used as-is`() {
        assertEquals(
            "https://cdn.example.com/img/42.jpg",
            resolveStreamUrl("https://cdn.example.com/img/42.jpg", "http://host:8080")
        )
    }

    @Test
    fun `null stream path resolves to null`() {
        assertNull(resolveStreamUrl(null, "http://host:8080"))
    }

    @Test
    fun `blank stream path resolves to null`() {
        assertNull(resolveStreamUrl("   ", "http://host:8080"))
    }

    // --- sibling index math: wraparound both directions ---

    @Test
    fun `next index advances`() {
        assertEquals(1, nextSiblingIndex(current = 0, size = 3))
        assertEquals(2, nextSiblingIndex(current = 1, size = 3))
    }

    @Test
    fun `next index wraps from last to first`() {
        assertEquals(0, nextSiblingIndex(current = 2, size = 3))
    }

    @Test
    fun `prev index retreats`() {
        assertEquals(1, prevSiblingIndex(current = 2, size = 3))
        assertEquals(0, prevSiblingIndex(current = 1, size = 3))
    }

    @Test
    fun `prev index wraps from first to last`() {
        assertEquals(2, prevSiblingIndex(current = 0, size = 3))
    }

    @Test
    fun `index math is safe for empty and single-element sets`() {
        assertEquals(0, nextSiblingIndex(current = 0, size = 0))
        assertEquals(0, prevSiblingIndex(current = 0, size = 0))
        assertEquals(0, nextSiblingIndex(current = 0, size = 1))
        assertEquals(0, prevSiblingIndex(current = 0, size = 1))
    }

    // --- position label ---

    @Test
    fun `position label shows one-based index over total`() {
        assertEquals("1 / 3", positionLabel(currentIndex = 0, total = 3))
        assertEquals("3 / 3", positionLabel(currentIndex = 2, total = 3))
    }

    @Test
    fun `position label is null for single-image or no set`() {
        assertNull(positionLabel(currentIndex = 0, total = 1))
        assertNull(positionLabel(currentIndex = 0, total = 0))
        assertNull(positionLabel(currentIndex = -1, total = 3))
    }
}
