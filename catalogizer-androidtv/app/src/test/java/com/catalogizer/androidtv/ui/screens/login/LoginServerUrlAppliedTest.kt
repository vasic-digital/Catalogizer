package com.catalogizer.androidtv.ui.screens.login

import android.content.Context
import com.catalogizer.androidtv.DependencyContainer
import io.mockk.clearAllMocks
import io.mockk.every
import io.mockk.mockk
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import org.junit.After
import org.junit.Assert.assertEquals
import org.junit.Assert.assertTrue
import org.junit.Before
import org.junit.Test
import java.io.File

/**
 * RED-baseline regression test for the MIBOX4 login defect (DEFECT-C, §11.4.115 polarity
 * switch): the user typed a Server URL into the login field and pressed Sign In, but the
 * login request still targeted the previously-persisted server URL because the Sign In
 * path never applied the entered URL to the API client.
 *
 * §11.4.115 polarity discipline — the SAME assertion runs in both modes; only whether the
 * REAL production apply path runs flips:
 *   RED_MODE=1  -> reproduce the PRE-FIX Sign In path: do NOT call applyServerUrlBeforeLogin
 *                  (the old handler never applied the field), then assert the active URL is
 *                  the ENTERED one. This GENUINELY FAILS because the URL is still STALE —
 *                  proving the defect is real and the assertion catches it.
 *   RED_MODE=0  -> GREEN guard: drive the FIXED path by calling the actual production helper
 *                  applyServerUrlBeforeLogin(...), then assert the active URL is the ENTERED
 *                  one. Passes because the fix applies the entered URL before login.
 *
 * Critically, the GREEN branch exercises the REAL applyServerUrlBeforeLogin() that the
 * fixed LoginScreen Sign In handlers call (LoginScreen.kt:381/472/489) — it does NOT
 * re-implement the fix by calling switchServer() directly. That is what makes this a
 * genuine regression guard and not a tautology (the prior version's bluff).
 *
 * RED_MODE is read from the environment / system property (mirrors MediaRepositoryTest),
 * so the same compiled test reproduces the defect under `RED_MODE=1` and guards under the
 * default `RED_MODE=0`.
 */
class LoginServerUrlAppliedTest {

    private companion object {
        // Env/system-property driven (NOT a compile-time const) so one compiled test can be
        // run in either polarity: `RED_MODE=1 ./gradlew ...` reproduces; default guards.
        val RED_MODE: Boolean =
            (System.getenv("RED_MODE") ?: System.getProperty("RED_MODE") ?: "0") == "1"

        const val STALE_URL = "http://10.89.10.9:8080"
        const val ENTERED_URL = "http://192.168.0.132:8080"
    }

    private lateinit var mockContext: Context
    private lateinit var container: DependencyContainer

    @Before
    fun setup() {
        mockContext = mockk(relaxed = true)
        every { mockContext.applicationContext } returns mockContext
        every { mockContext.filesDir } returns File("/tmp/test-files")

        val field = DependencyContainer::class.java.getDeclaredField("instance")
        field.isAccessible = true
        field.set(null, null)

        container = DependencyContainer(mockContext)
        // Establish the stale, previously-persisted server URL as the active base URL,
        // exactly the state that produced the MIBOX4 "/10.89.10.9 (port 8080)" error.
        container.switchServer(STALE_URL)
        assertEquals(STALE_URL, container.getServerUrl())
    }

    @After
    fun tearDown() {
        clearAllMocks()
        val field = DependencyContainer::class.java.getDeclaredField("instance")
        field.isAccessible = true
        field.set(null, null)
    }

    @Test
    fun `entered server url becomes the active base url before login`() {
        if (!RED_MODE) {
            // FIXED Sign In path: invoke the ACTUAL production helper the handlers call.
            // Dispatchers.Unconfined runs the synchronous switchServer() inline; the async
            // DataStore persist inside the helper is fire-and-forget (try/caught), so the
            // mock context cannot break the assertion.
            var serverUrlError: String? = null
            applyServerUrlBeforeLogin(
                serverUrl = ENTERED_URL,
                container = container,
                coroutineScope = CoroutineScope(Dispatchers.Unconfined)
            ) { serverUrlError = it }
            assertEquals(
                "applyServerUrlBeforeLogin must not reject a valid entered URL",
                null,
                serverUrlError
            )
        }
        // else RED_MODE: pre-fix path — the entered field is NEVER applied (no helper call),
        // so the active URL stays STALE and the assertion below genuinely FAILS.

        val active = container.getServerUrl()
        assertEquals(
            "Login must target the host the user entered in the Server URL field " +
                "(RED_MODE=1 reproduces the defect: active stays the stale persisted host)",
            ENTERED_URL,
            active
        )
        assertTrue(
            "Active base URL must contain the entered host, never the stale 10.89.10.9",
            active.contains("192.168.0.132") && !active.contains("10.89.10.9")
        )
    }
}
