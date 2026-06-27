package com.catalogizer.androidtv.ui.components

import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.size
import androidx.compose.ui.Modifier
import androidx.compose.ui.test.assertIsDisplayed
import androidx.compose.ui.test.getUnclippedBoundsInRoot
import androidx.compose.ui.test.junit4.createComposeRule
import androidx.compose.ui.test.onNodeWithText
import androidx.compose.ui.unit.dp
import com.catalogizer.androidtv.data.models.MediaItem
import org.junit.Rule
import org.junit.Test

/**
 * DEFECT-G on-device regression guard — title-clipped-off-card layout bug.
 *
 * Renders a real [MediaCard] (placeholder cover path, no network) inside a
 * fixed-size Box and asserts the title node is BOTH displayed AND laid out
 * within the card bounds (its bottom edge does not exceed the card bottom).
 *
 * This is the TRUE pixel-bounds oracle for the fix (FINDING 3) and requires a
 * device/emulator — it is the conductor's on-device follow-up per §11.4.52
 * (classification AUTONOMOUS_DESIGNED). The autonomous device-free leg of the
 * guard is the JVM unit test [MediaCardTitleLayoutTest].
 *
 * RED_MODE polarity (§11.4.115):
 *  - RED_MODE=1 (run against the PRE-FIX source where the cover Box is
 *    fillMaxWidth().aspectRatio(2f/3f) of a 2:3 card): the title bottom
 *    exceeds the card bottom (clipped off-card) — assert the defect present.
 *  - RED_MODE=0 (post-fix): the title is displayed and its bounds fall within
 *    the card — standing GREEN guard (§11.4.135).
 *
 * Run: ./gradlew connectedDebugAndroidTest  (with a connected Android TV
 * device / emulator). The conductor owns MIBOX4; this file authors the test
 * but does NOT execute it here (no device available to this agent).
 */
class MediaCardLayoutInstrumentedTest {

    @get:Rule
    val composeTestRule = createComposeRule()

    private val redMode: Int = System.getenv("RED_MODE")?.toIntOrNull() ?: 0

    private val sample = MediaItem(
        id = 42L,
        title = "Inception",
        mediaType = "movie",
        year = 2010,
        description = "A test movie",
        directoryPath = "/media/movies/inception",
        createdAt = "2010-07-16T00:00:00Z",
        updatedAt = "2010-07-16T00:00:00Z"
    )

    @Test
    fun titleIsLaidOutWithinCardBounds() {
        composeTestRule.setContent {
            // Constrain the card to a realistic poster footprint so bounds are
            // deterministic. The card itself is aspectRatio(2f/3f) internally.
            Box(modifier = Modifier.size(width = 200.dp, height = 300.dp)) {
                MediaCard(
                    mediaItem = sample,
                    onClick = {},
                    onFocus = {},
                    modifier = Modifier
                )
            }
        }

        val titleNode = composeTestRule.onNodeWithText("Inception")
        val cardHeight = 300.dp

        if (redMode == 1) {
            // PRE-FIX: title is pushed below the card. Its bottom exceeds the
            // card bottom (off-card) — the defect signature.
            val titleBottom = titleNode.getUnclippedBoundsInRoot().bottom
            assert(titleBottom > cardHeight) {
                "PRE-FIX expected title clipped off-card: titleBottom=$titleBottom card=$cardHeight"
            }
        } else {
            // POST-FIX: title visible and within the card.
            titleNode.assertIsDisplayed()
            val titleBottom = titleNode.getUnclippedBoundsInRoot().bottom
            assert(titleBottom <= cardHeight) {
                "POST-FIX expected title within card: titleBottom=$titleBottom card=$cardHeight"
            }
        }
    }
}
