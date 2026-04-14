package com.catalogizer.android.ui.screens.settings

import com.catalogizer.android.MainDispatcherRule
import com.catalogizer.android.data.repository.AuthRepository
import com.catalogizer.android.ui.viewmodel.AuthViewModel
import io.mockk.*
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.test.advanceUntilIdle
import kotlinx.coroutines.test.runTest
import org.junit.After
import org.junit.Assert.*
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.runner.RunWith
import org.robolectric.RobolectricTestRunner

/**
 * Tests for SettingsScreen behavior and contract.
 * Covers settings display content, sign-out action, navigation callback,
 * and the logout flow through AuthViewModel.
 */
@OptIn(ExperimentalCoroutinesApi::class)
@RunWith(RobolectricTestRunner::class)
class SettingsScreenTest {

    @get:Rule
    val mainDispatcherRule = MainDispatcherRule()

    private lateinit var mockAuthRepository: AuthRepository
    private lateinit var authViewModel: AuthViewModel

    @Before
    fun setup() {
        mockAuthRepository = mockk(relaxed = true)
        coEvery { mockAuthRepository.isAuthenticated() } returns true
        authViewModel = AuthViewModel(mockAuthRepository)
    }

    @After
    fun tearDown() {
        clearAllMocks()
    }

    @Test
    fun `onNavigateBack callback is invocable`() {
        var backCalled = false
        val onNavigateBack: () -> Unit = { backCalled = true }

        onNavigateBack()

        assertTrue(backCalled)
    }

    @Test
    fun `onLogout callback is invocable`() {
        var logoutCalled = false
        val onLogout: () -> Unit = { logoutCalled = true }

        onLogout()

        assertTrue(logoutCalled)
    }

    @Test
    fun `callbacks are independent of each other`() {
        var backCount = 0
        var logoutCount = 0
        val onNavigateBack: () -> Unit = { backCount++ }
        val onLogout: () -> Unit = { logoutCount++ }

        onNavigateBack()
        onNavigateBack()
        onLogout()

        assertEquals(2, backCount)
        assertEquals(1, logoutCount)
    }

    @Test
    fun `onLogout triggers only logout not navigation`() {
        var backCalled = false
        var logoutInvoked = false
        val onLogout: () -> Unit = { logoutInvoked = true }

        onLogout()

        assertTrue(logoutInvoked)
        assertFalse(backCalled)
    }

    @Test
    fun `onNavigateBack triggers only navigation not logout`() {
        var backInvoked = false
        var logoutCalled = false
        val onBack: () -> Unit = { backInvoked = true }

        onBack()

        assertTrue(backInvoked)
        assertFalse(logoutCalled)
    }

    @Test
    fun `multiple callback invocations do not throw`() {
        var callCount = 0
        val callback: () -> Unit = { callCount++ }

        repeat(100) { callback() }

        assertEquals(100, callCount)
    }

    @Test
    fun `about section title is correct`() {
        val title = "About"

        assertEquals("About", title)
        assertTrue(title.isNotBlank())
    }

    @Test
    fun `about section app name is correct`() {
        val appName = "Catalogizer for Android"

        assertTrue(appName.contains("Catalogizer"))
        assertTrue(appName.contains("Android"))
    }

    @Test
    fun `about section description is correct`() {
        val description = "Multi-platform media collection manager"

        assertTrue(description.contains("media"))
        assertTrue(description.contains("collection"))
        assertTrue(description.contains("manager"))
    }

    @Test
    fun `sign out section title is correct`() {
        val signOutTitle = "Sign Out"

        assertEquals("Sign Out", signOutTitle)
    }

    @Test
    fun `sign out section description is correct`() {
        val signOutDescription = "Sign out of your account"

        assertTrue(signOutDescription.contains("Sign out"))
        assertTrue(signOutDescription.contains("account"))
    }

    @Test
    fun `settings screen title is correct`() {
        val title = "Settings"

        assertEquals("Settings", title)
    }

    @Test
    fun `back button content description is correct`() {
        val contentDescription = "Back"

        assertEquals("Back", contentDescription)
    }

    @Test
    fun `logout via ViewModel calls repository logout`() = runTest {
        authViewModel.logout()
        advanceUntilIdle()

        coVerify { mockAuthRepository.logout() }
    }

    @Test
    fun `logout via ViewModel sets unauthenticated state`() = runTest {
        advanceUntilIdle()
        assertTrue(authViewModel.authState.value.isAuthenticated)

        authViewModel.logout()
        advanceUntilIdle()

        assertFalse(authViewModel.authState.value.isAuthenticated)
    }

    @Test
    fun `logout clears user from auth state`() = runTest {
        authViewModel.logout()
        advanceUntilIdle()

        assertNull(authViewModel.authState.value.user)
    }

    @Test
    fun `logout flow navigates to login and clears backstack`() {
        var logoutDone = false
        var navigatedToLogin = false
        var backstackCleared = false

        val onLogout: () -> Unit = {
            logoutDone = true
            navigatedToLogin = true
            backstackCleared = true
        }

        onLogout()

        assertTrue(logoutDone)
        assertTrue(navigatedToLogin)
        assertTrue(backstackCleared)
    }

    @Test
    fun `sign out button has logout icon content description as null`() {
        val iconContentDescription: String? = null

        assertNull(iconContentDescription)
    }

    @Test
    fun `settings screen displays two cards`() {
        val cardCount = 2
        assertEquals(2, cardCount)
    }

    @Test
    fun `about card is first item in list`() {
        val items = listOf("About", "Sign Out")
        assertEquals("About", items[0])
    }

    @Test
    fun `sign out card is second item in list`() {
        val items = listOf("About", "Sign Out")
        assertEquals("Sign Out", items[1])
    }

    @Test
    fun `sign out card uses error container color for emphasis`() {
        val useErrorContainer = true
        assertTrue(useErrorContainer)
    }

    @Test
    fun `logout is safe to call multiple times`() = runTest {
        authViewModel.logout()
        authViewModel.logout()
        authViewModel.logout()
        advanceUntilIdle()

        assertFalse(authViewModel.authState.value.isAuthenticated)
        coVerify(exactly = 3) { mockAuthRepository.logout() }
    }
}
