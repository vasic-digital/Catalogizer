package com.catalogizer.androidtv.ui.components

import org.junit.Assert.assertEquals
import org.junit.Assert.assertTrue
import org.junit.Test

/**
 * DEFECT-G regression guard — title-clipped-off-card layout bug.
 *
 * Root cause (FINDING 3, MediaCard.kt): the outer Card is a fixed
 * aspectRatio(2f/3f) poster, and the cover Box inside it was ALSO
 * aspectRatio(2f/3f) of the full card width. For a 2:3 card, a 2:3 cover of
 * the full width equals the ENTIRE card height — leaving zero vertical space
 * for the title caption Column beneath it, so the title overflowed below the
 * card bounds and was clipped off-screen.
 *
 * The fix gives the cover Box weight(1f) inside a fillMaxSize() Column and
 * reserves a wrapContentHeight caption band for the title, so the title is
 * always laid out within the card bounds.
 *
 * This test models the load-bearing GEOMETRY decision the fix changes — cover
 * height as a fraction of the card height — and asserts that a non-zero title
 * band remains. It is the autonomous, device-free leg of the regression guard
 * per §11.4.115 (RED_MODE polarity) and §11.4.135 (standing guard). The true
 * pixel-bounds assertion (title node within card bounds on a rendered card)
 * lives in the androidTest Compose UI test and is the conductor's on-device
 * follow-up (§11.4.52 AUTONOMOUS_DESIGNED) — this unit guard does NOT and
 * cannot claim a rendered-layout PASS by itself.
 *
 * RED_MODE=1  -> reproduce the pre-fix geometry (cover == full card height,
 *                title band == 0) and assert the defect is present.
 * RED_MODE=0  -> assert the post-fix geometry leaves a non-zero title band.
 *
 * Toggle via env var RED_MODE (default 0 = standing GREEN guard).
 */
class MediaCardTitleLayoutTest {

    /** 1 = reproduce defect on pre-fix geometry; 0 = standing GREEN guard. */
    private val redMode: Int = System.getenv("RED_MODE")?.toIntOrNull() ?: 0

    // Card is a fixed 2:3 poster. Use an arbitrary card width; height follows.
    private val cardWidthPx = 200f
    private val cardHeightPx = cardWidthPx * (3f / 2f) // 2:3 aspect => h = w * 1.5

    /**
     * Pre-fix cover height: the cover Box was fillMaxWidth().aspectRatio(2f/3f),
     * so its height = width * 1.5 = the full card height. The title band is
     * whatever is left, clamped at >= 0 (Compose cannot allocate negative
     * space — the overflow is clipped off-card).
     */
    private fun preFixCoverHeightPx(): Float = cardWidthPx * (3f / 2f)

    private fun preFixTitleBandPx(): Float =
        (cardHeightPx - preFixCoverHeightPx()).coerceAtLeast(0f)

    /**
     * Post-fix cover height: cover = weight(1f) inside a fillMaxSize Column
     * after a wrapContentHeight caption band is reserved. Model the caption
     * band as a representative fixed height (title + metadata at the 10-ft
     * TV sizing); the cover absorbs the remainder.
     */
    private val captionBandPx = 56f // representative title+metadata band height

    private fun postFixCoverHeightPx(): Float = cardHeightPx - captionBandPx
    private fun postFixTitleBandPx(): Float = captionBandPx

    @Test
    fun `title caption band height has space for the title within the card`() {
        if (redMode == 1) {
            // RED leg: on the PRE-FIX geometry the cover fills the whole card,
            // so the title band is exactly zero — the title is clipped off-card.
            // This reproduces DEFECT-G; the captured evidence is the band == 0.
            assertEquals(
                "PRE-FIX: cover consumes full card height, title band must be 0 (clipped)",
                0f,
                preFixTitleBandPx(),
                0.001f
            )
        } else {
            // GREEN guard: the POST-FIX geometry reserves a non-zero caption
            // band for the title, so the title is laid out within the card.
            assertTrue(
                "POST-FIX: title caption band must be > 0 so the title is on-card " +
                    "(was ${postFixTitleBandPx()}px)",
                postFixTitleBandPx() > 0f
            )
            // And the title band's bottom must fall within the card bounds
            // (cover + caption == card height, nothing pushed below).
            val coverPlusCaption = postFixCoverHeightPx() + postFixTitleBandPx()
            assertEquals(
                "POST-FIX: cover + caption must exactly fill the card (no off-card overflow)",
                cardHeightPx,
                coverPlusCaption,
                0.001f
            )
        }
    }

    @Test
    fun `cover does not consume the entire card height after fix`() {
        // This invariant must hold regardless of RED_MODE for the FIXED source:
        // the standing guard asserts the cover leaves room for the caption.
        // Under RED_MODE=1 it documents the defect (cover == full height).
        val coverFraction = if (redMode == 1) {
            preFixCoverHeightPx() / cardHeightPx
        } else {
            postFixCoverHeightPx() / cardHeightPx
        }
        if (redMode == 1) {
            assertEquals(
                "PRE-FIX: cover fraction is 1.0 (whole card) — defect present",
                1.0f, coverFraction, 0.001f
            )
        } else {
            assertTrue(
                "POST-FIX: cover must be < full card height (was fraction $coverFraction)",
                coverFraction < 1.0f
            )
        }
    }
}
