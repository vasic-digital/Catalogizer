package com.catalogizer.androidtv.ui.screens.viewer

import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.ui.Modifier
import androidx.compose.ui.test.assertIsDisplayed
import androidx.compose.ui.test.junit4.createComposeRule
import androidx.compose.ui.test.onNodeWithContentDescription
import androidx.compose.ui.test.onNodeWithText
import androidx.compose.ui.test.onRoot
import com.catalogizer.androidtv.ui.theme.CatalogizerTVTheme
import com.github.takahirom.roborazzi.captureRoboImage
import org.junit.Rule
import org.junit.Test
import org.junit.runner.RunWith
import org.robolectric.RobolectricTestRunner
import org.robolectric.annotation.Config
import org.robolectric.annotation.GraphicsMode

/**
 * §11.4.170 device-independent host-side rendered-UI visual proof for the new
 * [BookReaderContent] screen.
 *
 * Renders the REAL composable to a PNG ON THE HOST (Robolectric + Roborazzi — no
 * device, no emulator, no running app) for every visual state × {light, dark}:
 *  - Loading      : spinner + "Loading book…" + top overlay.
 *  - Error (.pdf) : the honest "not a PDF document" message (§11.4.1 — never a
 *                   blank screen for an unreadable file).
 *  - Ready        : the page surface (Coil falls into its error slot under
 *                   Robolectric's no-network env, the §11.4.1 graceful path) WITH
 *                   the "Page X / N" indicator + the page-turn hint.
 *
 * Dual validation per §11.4.170:
 *  (i)  golden image-diff — `captureRoboImage(...)` writes a PNG the conductor
 *       records via `recordRoborazziDebug` / re-checks via `verifyRoborazziDebug`.
 *  (ii) rendered-label/bounds oracle — every capture asserts the title label, the
 *       focusable Back affordance, and (for the Ready state) the rendered page
 *       indicator are actually DISPLAYED (present, not clipped / overlapped /
 *       off-screen). VALUE/token-equality assertions are forbidden as the proof
 *       here (§11.4.170); these are rendered pixels + the rendered-label oracle.
 *
 * The conductor records the goldens (this subagent records them via
 * `recordRoborazziDebug`; the screen has no device and no running app).
 */
@RunWith(RobolectricTestRunner::class)
@GraphicsMode(GraphicsMode.Mode.NATIVE)
@Config(sdk = [33])
class BookReaderScreenshotTest {

    @get:Rule
    val composeRule = createComposeRule()

    private val sampleTitle = "Clean Code"
    private val outDir = "src/test/screenshots/bookreader"

    private fun render(
        darkTheme: Boolean,
        state: BookReaderUiState,
        pageImageUrl: String? = null,
        pageLabel: String? = null,
        zoomed: Boolean = false,
        hasMultiplePages: Boolean = false
    ) {
        composeRule.setContent {
            CatalogizerTVTheme(darkTheme = darkTheme) {
                Box(modifier = Modifier.fillMaxSize()) {
                    BookReaderContent(
                        title = sampleTitle,
                        state = state,
                        pageImageUrl = pageImageUrl,
                        pageLabel = pageLabel,
                        zoomed = zoomed,
                        hasMultiplePages = hasMultiplePages,
                        authToken = null,
                        onBack = {},
                        onToggleZoom = {},
                        modifier = Modifier.fillMaxSize()
                    )
                }
            }
        }
    }

    /** The §11.4.170 rendered-label oracle shared by every capture. */
    private fun assertOverlayLabelled() {
        // Title is rendered + on-screen (not clipped / overlapped / off-screen).
        composeRule.onNodeWithText(sampleTitle).assertIsDisplayed()
        // The focusable Back affordance is present so the user always has a way out.
        composeRule.onNodeWithContentDescription("Back").assertIsDisplayed()
    }

    @Test
    fun bookReader_loading_dark() {
        render(darkTheme = true, state = BookReaderUiState.Loading)
        assertOverlayLabelled()
        composeRule.onNodeWithText("Loading book…").assertIsDisplayed()
        composeRule.onRoot().captureRoboImage("$outDir/bookreader_loading_dark.png")
    }

    @Test
    fun bookReader_loading_light() {
        render(darkTheme = false, state = BookReaderUiState.Loading)
        assertOverlayLabelled()
        composeRule.onRoot().captureRoboImage("$outDir/bookreader_loading_light.png")
    }

    @Test
    fun bookReader_error_dark() {
        render(
            darkTheme = true,
            state = BookReaderUiState.Error("This file is not a PDF document (.pdf).")
        )
        assertOverlayLabelled()
        composeRule.onNodeWithText("Unable to Open Book").assertIsDisplayed()
        // §11.4.1 — the not-a-PDF message is rendered, never a blank screen.
        composeRule.onNodeWithText("This file is not a PDF document (.pdf).").assertIsDisplayed()
        composeRule.onRoot().captureRoboImage("$outDir/bookreader_error_dark.png")
    }

    @Test
    fun bookReader_error_light() {
        render(
            darkTheme = false,
            state = BookReaderUiState.Error("This file is not a PDF document (.pdf).")
        )
        assertOverlayLabelled()
        composeRule.onRoot().captureRoboImage("$outDir/bookreader_error_light.png")
    }

    @Test
    fun bookReader_ready_page_dark_with_indicator() {
        render(
            darkTheme = true,
            state = BookReaderUiState.Ready(totalPages = 58),
            pageImageUrl = "https://example.invalid/pdf-pages/2",
            pageLabel = "Page 3 / 58",
            zoomed = false,
            hasMultiplePages = true
        )
        assertOverlayLabelled()
        // The page indicator + page-turn hint are rendered for a multi-page book.
        composeRule.onNodeWithText("Page 3 / 58").assertIsDisplayed()
        composeRule.onNodeWithText("Use ◀ ▶ to turn pages").assertIsDisplayed()
        composeRule.onRoot().captureRoboImage("$outDir/bookreader_ready_dark.png")
    }

    @Test
    fun bookReader_ready_zoomed_light_single_page() {
        render(
            darkTheme = false,
            state = BookReaderUiState.Ready(totalPages = 1),
            pageImageUrl = "https://example.invalid/pdf-pages/0",
            pageLabel = "Page 1 / 1",
            zoomed = true,
            hasMultiplePages = false
        )
        assertOverlayLabelled()
        // Single-page book: the zoom affordance reflects the zoomed state.
        composeRule.onNodeWithContentDescription("Fit to screen").assertIsDisplayed()
        composeRule.onRoot().captureRoboImage("$outDir/bookreader_ready_zoomed_light.png")
    }
}
