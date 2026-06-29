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
 * [ImageViewerContent] screen.
 *
 * Renders the REAL composable to a PNG ON THE HOST (Robolectric + Roborazzi — no
 * device, no emulator, no running app) for every visual state × {light, dark}:
 *  - Loading  : spinner + "Loading image…" + top overlay.
 *  - Error    : branded broken-image message (§11.4.1 — never a blank screen).
 *  - Ready    : the image surface (Coil falls into its error slot under
 *               Robolectric's no-network env, which is itself the §11.4.1
 *               graceful path) with the Fit vs Crop (zoomed) overlay state.
 *
 * Dual validation per §11.4.170:
 *  (i)  golden image-diff — `captureRoboImage(...)` writes a PNG the conductor
 *       records via `recordRoborazziDebug` / re-checks via `verifyRoborazziDebug`.
 *  (ii) rendered-label/bounds oracle (the host-render analogue of the OCR/vision
 *       layout check) — every capture is paired with assertions that the title
 *       label AND the focusable Back affordance are actually displayed (present,
 *       not clipped / overlapped / off-screen). VALUE/token-equality assertions
 *       are forbidden as the proof here (§11.4.170); these are rendered pixels +
 *       the rendered-label oracle.
 *
 * The conductor finalizes Roborazzi gradle wiring + records the goldens (this
 * subagent has no device and does not run gradle).
 */
@RunWith(RobolectricTestRunner::class)
@GraphicsMode(GraphicsMode.Mode.NATIVE)
@Config(sdk = [33])
class ImageViewerScreenshotTest {

    @get:Rule
    val composeRule = createComposeRule()

    private val sampleTitle = "Holiday Photo 2026"
    private val outDir = "src/test/screenshots/imageviewer"

    private fun render(
        darkTheme: Boolean,
        state: ImageViewerUiState,
        zoomed: Boolean = false,
        hasSiblings: Boolean = false,
        positionLabel: String? = null
    ) {
        composeRule.setContent {
            CatalogizerTVTheme(darkTheme = darkTheme) {
                Box(modifier = Modifier.fillMaxSize()) {
                    ImageViewerContent(
                        title = sampleTitle,
                        state = state,
                        zoomed = zoomed,
                        positionLabel = positionLabel,
                        hasSiblings = hasSiblings,
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
    fun imageViewer_loading_dark() {
        render(darkTheme = true, state = ImageViewerUiState.Loading)
        assertOverlayLabelled()
        composeRule.onNodeWithText("Loading image…").assertIsDisplayed()
        composeRule.onRoot().captureRoboImage("$outDir/imageviewer_loading_dark.png")
    }

    @Test
    fun imageViewer_loading_light() {
        render(darkTheme = false, state = ImageViewerUiState.Loading)
        assertOverlayLabelled()
        composeRule.onRoot().captureRoboImage("$outDir/imageviewer_loading_light.png")
    }

    @Test
    fun imageViewer_error_dark() {
        render(darkTheme = true, state = ImageViewerUiState.Error("No file linked to this image."))
        assertOverlayLabelled()
        composeRule.onNodeWithText("Unable to Display Image").assertIsDisplayed()
        composeRule.onNodeWithText("No file linked to this image.").assertIsDisplayed()
        composeRule.onRoot().captureRoboImage("$outDir/imageviewer_error_dark.png")
    }

    @Test
    fun imageViewer_error_light() {
        render(darkTheme = false, state = ImageViewerUiState.Error("No file linked to this image."))
        assertOverlayLabelled()
        composeRule.onRoot().captureRoboImage("$outDir/imageviewer_error_light.png")
    }

    @Test
    fun imageViewer_ready_fit_dark_with_siblings() {
        render(
            darkTheme = true,
            state = ImageViewerUiState.Ready("https://example.invalid/holiday.jpg"),
            zoomed = false,
            hasSiblings = true,
            positionLabel = "3 / 12"
        )
        assertOverlayLabelled()
        // Sibling navigation hint + position are rendered when a set exists.
        composeRule.onNodeWithText("Use ◀ ▶ to browse").assertIsDisplayed()
        composeRule.onNodeWithText("3 / 12").assertIsDisplayed()
        composeRule.onRoot().captureRoboImage("$outDir/imageviewer_ready_fit_dark.png")
    }

    @Test
    fun imageViewer_ready_zoomed_light_single() {
        render(
            darkTheme = false,
            state = ImageViewerUiState.Ready("https://example.invalid/holiday.jpg"),
            zoomed = true,
            hasSiblings = false,
            positionLabel = null
        )
        assertOverlayLabelled()
        // Single-image mode: the zoom affordance reflects the zoomed state.
        composeRule.onNodeWithContentDescription("Fit to screen").assertIsDisplayed()
        composeRule.onRoot().captureRoboImage("$outDir/imageviewer_ready_zoomed_light.png")
    }
}
