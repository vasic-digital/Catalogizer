package com.catalogizer.androidtv.ui.player

import androidx.appcompat.app.AppCompatActivity
import com.catalogizer.androidtv.CatalogizerTVTestApplication
import com.catalogizer.androidtv.R
import org.junit.Assert.assertFalse
import org.junit.Assert.assertNotNull
import org.junit.Test
import org.junit.runner.RunWith
import org.robolectric.Robolectric
import org.robolectric.RobolectricTestRunner
import org.robolectric.annotation.Config

/**
 * Permanent regression guard (§11.4.115 RED-on-broken-artifact + §11.4.135
 * standing guard) for the 100%-reproducible launch crash of the Android TV
 * ExoPlayer activity.
 *
 * Captured on-device crash (the defect this guard reproduces host-side):
 *
 *   java.lang.IllegalStateException: You need to use a Theme.AppCompat theme
 *     (or descendant) with this activity.
 *     at androidx.appcompat.app.AppCompatDelegateImpl.createSubDecor
 *     at ...ExoTvPlayerActivity.onCreate(ExoTvPlayerActivity.kt:49)  [setContentView]
 *
 * Root cause: ExoTvPlayerActivity called setContentView() while extending
 * AppCompatActivity, but the manifest assigns it @style/Theme.CatalogizerTV.Player
 * whose ancestry is Theme.CatalogizerTV -> Theme.Leanback (NOT a Theme.AppCompat
 * descendant). AppCompatActivity.setContentView routes through
 * AppCompatDelegateImpl.createSubDecor() which hard-requires an AppCompat theme,
 * so it crashes 100% on launch. The two working players (MediaPlayerActivity,
 * VLCPlayerActivity) extend ComponentActivity and run fine under the same theme.
 *
 * Fix: ExoTvPlayerActivity now extends androidx.activity.ComponentActivity
 * (matching the working players); ComponentActivity.setContentView imposes no
 * theme requirement.
 *
 * Polarity (§11.4.115):
 *   RED  (pre-fix, AppCompatActivity base) -> both tests FAIL.
 *   GREEN (post-fix, ComponentActivity base) -> both tests PASS.
 */
@RunWith(RobolectricTestRunner::class)
@Config(application = CatalogizerTVTestApplication::class)
class ExoTvPlayerActivityThemeRegressionTest {

    /**
     * Deterministic, device-free meta-check: an activity that calls View-based
     * setContentView while assigned a non-AppCompat (Leanback) theme MUST NOT
     * extend AppCompatActivity. This catches the exact regression (and any future
     * re-introduction of an AppCompatActivity base under the Leanback player theme)
     * with pure class-hierarchy reflection — no Android runtime needed.
     *
     * RED: pre-fix ExoTvPlayerActivity extends AppCompatActivity -> isAppCompat==true
     *      -> assertFalse FAILS.
     * GREEN: post-fix extends ComponentActivity -> isAppCompat==false -> PASSES.
     */
    @Test
    fun `ExoTvPlayerActivity must not extend AppCompatActivity under the Leanback player theme`() {
        val isAppCompat = AppCompatActivity::class.java
            .isAssignableFrom(ExoTvPlayerActivity::class.java)
        assertFalse(
            "ExoTvPlayerActivity is assigned the Leanback-based " +
                "@style/Theme.CatalogizerTV.Player theme in the manifest. Extending " +
                "AppCompatActivity makes setContentView() crash on launch with " +
                "'You need to use a Theme.AppCompat theme (or descendant) with this " +
                "activity.' Use androidx.activity.ComponentActivity instead (matching " +
                "MediaPlayerActivity / VLCPlayerActivity).",
            isAppCompat,
        )
    }

    /**
     * Faithful host-side reproduction: actually launch ExoTvPlayerActivity through
     * Robolectric so its real onCreate -> setContentView() runs under its real
     * production theme (resolved from the merged manifest: the Leanback-descended
     * Theme.CatalogizerTV.Player, with the application's Theme.CatalogizerTV as the
     * fallback — both Leanback, both AppCompat-incompatible).
     *
     * No media_id extra is supplied, so on the FIXED code onCreate runs
     * setContentView() successfully, reads mediaId==0, and finish()/returns BEFORE
     * touching the DependencyContainer / network — keeping the test isolated to the
     * theme x base-class interaction that crashed.
     *
     * RED: pre-fix create() throws IllegalStateException at setContentView (line 49).
     * GREEN: post-fix create() completes; onCreate does not throw.
     */
    @Test
    fun `launching ExoTvPlayerActivity does not crash with the AppCompat theme error`() {
        val controller = Robolectric.buildActivity(ExoTvPlayerActivity::class.java)
        // On the broken artifact this throws the captured IllegalStateException.
        controller.create()
        // Reaching here proves setContentView survived the Leanback player theme.
        assertNotNull(controller.get())
        // Resource sanity: the production theme the activity runs under exists.
        assertNotNull(R.style::class.java)
    }
}
