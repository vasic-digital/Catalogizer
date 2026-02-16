package com.catalogizer.android.data.models

import org.junit.Assert.*
import org.junit.Test

class AuthModelsTest {

    // --- AuthState ---

    @Test
    fun `AuthState has correct defaults`() {
        val state = AuthState()

        assertFalse(state.isAuthenticated)
        assertFalse(state.isLoading)
        assertNull(state.error)
        assertNull(state.user)
    }

    @Test
    fun `AuthState authenticated state`() {
        val user = User(
            id = 1, username = "admin", email = "admin@test.com",
            firstName = "A", lastName = "B", role = "admin",
            isActive = true, createdAt = "2025-01-01", updatedAt = "2025-01-01"
        )
        val state = AuthState(isAuthenticated = true, user = user)

        assertTrue(state.isAuthenticated)
        assertNotNull(state.user)
        assertEquals("admin", state.user?.username)
    }

    @Test
    fun `AuthState loading state`() {
        val state = AuthState(isLoading = true)

        assertTrue(state.isLoading)
        assertFalse(state.isAuthenticated)
    }

    @Test
    fun `AuthState error state`() {
        val state = AuthState(error = "Invalid credentials")

        assertFalse(state.isAuthenticated)
        assertEquals("Invalid credentials", state.error)
    }

    @Test
    fun `AuthState copy updates correctly`() {
        val initial = AuthState()
        val loading = initial.copy(isLoading = true)
        val error = loading.copy(isLoading = false, error = "Failed")

        assertFalse(initial.isLoading)
        assertTrue(loading.isLoading)
        assertFalse(error.isLoading)
        assertEquals("Failed", error.error)
    }

    // --- ApiResponse ---

    @Test
    fun `ApiResponse with data`() {
        val response = ApiResponse(data = "test data", message = "Success")

        assertEquals("test data", response.data)
        assertEquals("Success", response.message)
        assertNull(response.error)
    }

    @Test
    fun `ApiResponse with error`() {
        val response = ApiResponse<String>(error = "Not found")

        assertNull(response.data)
        assertEquals("Not found", response.error)
    }

    @Test
    fun `ApiResponse with null data`() {
        val response = ApiResponse<String>()

        assertNull(response.data)
        assertNull(response.error)
        assertNull(response.message)
    }

    // --- ErrorResponse ---

    @Test
    fun `ErrorResponse with full data`() {
        val details = mapOf("field" to "username", "reason" to "already_taken")
        val response = ErrorResponse(
            error = "Validation error",
            code = 422,
            details = details
        )

        assertEquals("Validation error", response.error)
        assertEquals(422, response.code)
        assertEquals("username", response.details?.get("field"))
        assertEquals("already_taken", response.details?.get("reason"))
    }

    @Test
    fun `ErrorResponse with minimal data`() {
        val response = ErrorResponse(error = "Server error")

        assertEquals("Server error", response.error)
        assertNull(response.code)
        assertNull(response.details)
    }

    // --- LoginRequest ---

    @Test
    fun `LoginRequest holds credentials`() {
        val request = LoginRequest(username = "admin", password = "secret123")

        assertEquals("admin", request.username)
        assertEquals("secret123", request.password)
    }

    // --- RegisterRequest ---

    @Test
    fun `RegisterRequest holds all fields`() {
        val request = RegisterRequest(
            username = "newuser",
            email = "new@test.com",
            password = "pass123",
            firstName = "New",
            lastName = "User"
        )

        assertEquals("newuser", request.username)
        assertEquals("new@test.com", request.email)
        assertEquals("pass123", request.password)
        assertEquals("New", request.firstName)
        assertEquals("User", request.lastName)
    }

    // --- ChangePasswordRequest ---

    @Test
    fun `ChangePasswordRequest holds passwords`() {
        val request = ChangePasswordRequest(
            currentPassword = "old123",
            newPassword = "new456"
        )

        assertEquals("old123", request.currentPassword)
        assertEquals("new456", request.newPassword)
    }

    // --- UpdateProfileRequest ---

    @Test
    fun `UpdateProfileRequest with all fields`() {
        val request = UpdateProfileRequest(
            firstName = "Updated",
            lastName = "Name",
            email = "updated@test.com"
        )

        assertEquals("Updated", request.firstName)
        assertEquals("Name", request.lastName)
        assertEquals("updated@test.com", request.email)
    }

    @Test
    fun `UpdateProfileRequest with null fields`() {
        val request = UpdateProfileRequest(firstName = "Only First")

        assertEquals("Only First", request.firstName)
        assertNull(request.lastName)
        assertNull(request.email)
    }
}
