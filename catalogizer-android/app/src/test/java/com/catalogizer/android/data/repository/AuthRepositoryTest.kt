package com.catalogizer.android.data.repository

import androidx.datastore.core.DataStore
import androidx.datastore.preferences.core.*
import com.catalogizer.android.MainDispatcherRule
import com.catalogizer.android.data.models.*
import com.catalogizer.android.data.remote.ApiResult
import com.catalogizer.android.data.remote.CatalogizerApi
import io.mockk.*
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.flow.*
import kotlinx.coroutines.test.runTest
import org.junit.After
import org.junit.Assert.*
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import retrofit2.Response

@OptIn(ExperimentalCoroutinesApi::class)
class AuthRepositoryTest {

    @get:Rule
    val mainDispatcherRule = MainDispatcherRule()

    private lateinit var repository: AuthRepository
    private val mockApi = mockk<CatalogizerApi>(relaxed = true)
    private val mockDataStore = mockk<DataStore<Preferences>>(relaxed = true)

    private val testUser = User(
        id = 1L,
        username = "testuser",
        email = "test@example.com",
        firstName = "Test",
        lastName = "User",
        role = "user",
        isActive = true,
        createdAt = "2024-01-01T00:00:00Z",
        updatedAt = "2024-01-01T00:00:00Z",
        permissions = listOf("read:media", "write:media")
    )

    private val testLoginResponse = LoginResponse(
        user = testUser,
        token = "test-token-123",
        refreshToken = "refresh-token-456",
        expiresIn = 3600
    )

    private val emptyPreferences = mockk<Preferences> {
        every { get(any<Preferences.Key<String>>()) } returns null
        every { get(any<Preferences.Key<Long>>()) } returns null
        every { get(any<Preferences.Key<Boolean>>()) } returns null
    }

    @Before
    fun setup() {
        every { mockDataStore.data } returns flowOf(emptyPreferences)
        repository = AuthRepository(mockApi, mockDataStore)
    }

    @After
    fun tearDown() {
        clearAllMocks()
    }

    // --- Login Tests ---

    @Test
    fun `login success should save auth data and return success`() = runTest {
        val response = Response.success(testLoginResponse)
        coEvery { mockApi.login(any()) } returns response
        coEvery { mockDataStore.updateData(any()) } returns emptyPreferences

        val result = repository.login("testuser", "password123")

        assertTrue(result.isSuccess)
        assertNotNull(result.data)
        assertEquals("test-token-123", result.data?.token)
        coVerify { mockApi.login(LoginRequest("testuser", "password123")) }
    }

    @Test
    fun `login failure from API should return error`() = runTest {
        val errorResponse = Response.error<LoginResponse>(
            401,
            okhttp3.ResponseBody.create(null, "Unauthorized")
        )
        coEvery { mockApi.login(any()) } returns errorResponse

        val result = repository.login("testuser", "wrongpassword")

        assertFalse(result.isSuccess)
        assertNotNull(result.error)
    }

    @Test
    fun `login with exception should return error`() = runTest {
        coEvery { mockApi.login(any()) } throws RuntimeException("Network failure")

        val result = repository.login("testuser", "password")

        assertFalse(result.isSuccess)
        assertEquals("Network failure", result.error)
    }

    @Test
    fun `login with rememberMe saves remember preference`() = runTest {
        val response = Response.success(testLoginResponse)
        coEvery { mockApi.login(any()) } returns response
        coEvery { mockDataStore.updateData(any()) } returns emptyPreferences

        val result = repository.login("testuser", "password", rememberMe = true)

        assertTrue(result.isSuccess)
        coVerify { mockDataStore.updateData(any()) }
    }

    // --- Register Tests ---

    @Test
    fun `register success should return user`() = runTest {
        val registeredUser = User(
            id = 2L,
            username = "newuser",
            email = "new@email.com",
            firstName = "New",
            lastName = "User",
            role = "user",
            isActive = true,
            createdAt = "2024-01-01T00:00:00Z",
            updatedAt = "2024-01-01T00:00:00Z",
            permissions = null
        )
        val response = Response.success(registeredUser)
        coEvery { mockApi.register(any()) } returns response

        val result = repository.register(
            "newuser", "new@email.com", "password", "New", "User"
        )

        assertTrue(result.isSuccess)
        assertEquals("newuser", result.data?.username)
        coVerify {
            mockApi.register(RegisterRequest("newuser", "new@email.com", "password", "New", "User"))
        }
    }

    @Test
    fun `register with exception should return error`() = runTest {
        coEvery { mockApi.register(any()) } throws RuntimeException("Server error")

        val result = repository.register(
            "newuser", "new@email.com", "password", "New", "User"
        )

        assertFalse(result.isSuccess)
        assertEquals("Server error", result.error)
    }

    // --- Logout Tests ---

    @Test
    fun `logout should clear auth data even if API call fails`() = runTest {
        coEvery { mockApi.logout() } throws RuntimeException("Network error")
        coEvery { mockDataStore.updateData(any()) } returns emptyPreferences

        val result = repository.logout()

        assertTrue(result.isSuccess)
        coVerify { mockDataStore.updateData(any()) }
    }

    @Test
    fun `logout should call API logout and clear local data`() = runTest {
        val response = Response.success(Unit)
        coEvery { mockApi.logout() } returns response
        coEvery { mockDataStore.updateData(any()) } returns emptyPreferences

        val result = repository.logout()

        assertTrue(result.isSuccess)
        coVerify { mockApi.logout() }
        coVerify { mockDataStore.updateData(any()) }
    }

    // --- Token Refresh Tests ---

    @Test
    fun `refreshAuthToken with no refresh token should return error`() = runTest {
        val prefsWithNoToken = mockk<Preferences> {
            every { get(any<Preferences.Key<String>>()) } returns null
            every { get(any<Preferences.Key<Long>>()) } returns null
            every { get(any<Preferences.Key<Boolean>>()) } returns null
        }
        every { mockDataStore.data } returns flowOf(prefsWithNoToken)
        repository = AuthRepository(mockApi, mockDataStore)

        val result = repository.refreshAuthToken()

        assertFalse(result.isSuccess)
        assertEquals("No refresh token available", result.error)
    }

    // --- Profile Tests ---

    @Test
    fun `getProfile success should return user and save`() = runTest {
        val response = Response.success(testUser)
        coEvery { mockApi.getProfile() } returns response
        coEvery { mockDataStore.updateData(any()) } returns emptyPreferences

        val result = repository.getProfile()

        assertTrue(result.isSuccess)
        assertEquals("testuser", result.data?.username)
        coVerify { mockDataStore.updateData(any()) }
    }

    @Test
    fun `getProfile with exception should return error`() = runTest {
        coEvery { mockApi.getProfile() } throws RuntimeException("Connection refused")

        val result = repository.getProfile()

        assertFalse(result.isSuccess)
        assertEquals("Connection refused", result.error)
    }

    // --- Settings Tests ---

    @Test
    fun `setServerUrl should update datastore`() = runTest {
        coEvery { mockDataStore.updateData(any()) } returns emptyPreferences

        repository.setServerUrl("http://192.168.1.100:8080")

        coVerify { mockDataStore.updateData(any()) }
    }

    @Test
    fun `setBiometricEnabled should update datastore`() = runTest {
        coEvery { mockDataStore.updateData(any()) } returns emptyPreferences

        repository.setBiometricEnabled(true)

        coVerify { mockDataStore.updateData(any()) }
    }

    // --- Permission Tests ---

    @Test
    fun `hasPermission returns false when no user`() = runTest {
        val result = repository.hasPermission("read:media")

        assertFalse(result)
    }

    // --- Utility Tests ---

    @Test
    fun `getUserDisplayName returns Unknown User when no user`() = runTest {
        val displayName = repository.getUserDisplayName()

        assertEquals("Unknown User", displayName)
    }

    @Test
    fun `getUserInitials returns question mark when no user`() = runTest {
        val initials = repository.getUserInitials()

        assertEquals("?", initials)
    }
}
