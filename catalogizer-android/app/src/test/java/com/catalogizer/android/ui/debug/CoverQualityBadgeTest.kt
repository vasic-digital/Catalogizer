package com.catalogizer.android.ui.debug

import org.junit.Assert.assertEquals
import org.junit.Assert.assertNotNull
import org.junit.Before
import org.junit.Test

/**
 * Unit coverage for the CoverQualityBadge module. We cannot instantiate
 * the Compose UI from a pure-JVM unit test, so these cases exercise the
 * data surface (CoverQualitySignal + cache).
 */
class CoverQualityBadgeTest {

    @Before
    fun setUp() {
        resetCoverQualityCache()
    }

    @Test
    fun signal_carriesQualityAndSource() {
        val s = CoverQualitySignal(quality = "pass", source = "tmdb")
        assertEquals("pass", s.quality)
        assertEquals("tmdb", s.source)
    }

    @Test
    fun resetCoverQualityCache_isIdempotent() {
        // Double reset must not throw.
        resetCoverQualityCache()
        resetCoverQualityCache()
        assertNotNull(CoverQualitySignal("unknown", ""))
    }
}
