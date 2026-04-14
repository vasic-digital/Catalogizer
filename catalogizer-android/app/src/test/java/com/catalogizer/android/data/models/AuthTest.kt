package com.catalogizer.android.data.models

import org.junit.Assert.*
import org.junit.Test

class AuthTest {

    // --- User ---

    @Test
    fun `User fullName combines first and last name`() {
        val user = createUser(firstName = "John", lastName = "Doe")
        assertEquals("John Doe", user.fullName)
    }

    @Test
    fun `User fullName trims whitespace when lastName is empty`() {
        val user = createUser(firstName = "John", lastName = "")
        assertEquals("John", user.fullName)
    }

    @Test
    fun `User fullName trims whitespace when firstName is empty`() {
        val user = createUser(firstName = "", lastName = "Doe")
        assertEquals("Doe", user.fullName)
    }

    @Test
    fun `User fullName is empty when both names are empty`() {
        val user = createUser(firstName = "", lastName = "")
        assertEquals("", user.fullName)
    }

    @Test
    fun `User isAdmin returns true for admin role`() {
        val user = createUser(role = "admin")
        assertTrue(user.isAdmin)
    }

    @Test
    fun `User isAdmin returns false for user role`() {
        val user = createUser(role = "user")
        assertFalse(user.isAdmin)
    }

    @Test
    fun `User isAdmin returns false for moderator role`() {
        val user = createUser(role = "moderator")
        assertFalse(user.isAdmin)
    }

    @Test
    fun `User isAdmin returns false for viewer role`() {
        val user = createUser(role = "viewer")
        assertFalse(user.isAdmin)
    }

    @Test
    fun `User default optional fields are null`() {
        val user = createUser()
        assertNull(user.lastLogin)
        assertNull(user.permissions)
    }

    @Test
    fun `User with permissions list`() {
        val perms = listOf("read:media", "write:media", "delete:media")
        val user = createUser(permissions = perms)
        assertEquals(3, user.permissions?.size)
        assertEquals("read:media", user.permissions?.get(0))
        assertEquals("write:media", user.permissions?.get(1))
        assertEquals("delete:media", user.permissions?.get(2))
    }

    @Test
    fun `User with empty permissions list`() {
        val user = createUser(permissions = emptyList())
        assertNotNull(user.permissions)
        assertTrue(user.permissions!!.isEmpty())
    }

    @Test
    fun `User copy preserves all fields`() {
        val original = User(
            id = 1, username = "test", email = "test@test.com",
            firstName = "First", lastName = "Last", role = "admin",
            isActive = true, lastLogin = "2026-01-01",
            createdAt = "2025-01-01", updatedAt = "2026-01-01",
            permissions = listOf("read:media")
        )
        val copy = original.copy(username = "changed")

        assertEquals("changed", copy.username)
        assertEquals(original.id, copy.id)
        assertEquals(original.email, copy.email)
        assertEquals(original.firstName, copy.firstName)
        assertEquals(original.lastName, copy.lastName)
        assertEquals(original.role, copy.role)
        assertEquals(original.isActive, copy.isActive)
        assertEquals(original.lastLogin, copy.lastLogin)
        assertEquals(original.createdAt, copy.createdAt)
        assertEquals(original.updatedAt, copy.updatedAt)
        assertEquals(original.permissions, copy.permissions)
    }

    @Test
    fun `User equality based on all fields`() {
        val user1 = createUser(id = 1, username = "admin")
        val user2 = createUser(id = 1, username = "admin")
        assertEquals(user1, user2)
    }

    @Test
    fun `User inequality when fields differ`() {
        val user1 = createUser(id = 1)
        val user2 = createUser(id = 2)
        assertNotEquals(user1, user2)
    }

    // --- LoginRequest ---

    @Test
    fun `LoginRequest stores username and password`() {
        val request = LoginRequest(username = "admin", password = "pass123")
        assertEquals("admin", request.username)
        assertEquals("pass123", request.password)
    }

    @Test
    fun `LoginRequest with empty credentials`() {
        val request = LoginRequest(username = "", password = "")
        assertEquals("", request.username)
        assertEquals("", request.password)
    }

    @Test
    fun `LoginRequest equality`() {
        val r1 = LoginRequest("user", "pass")
        val r2 = LoginRequest("user", "pass")
        assertEquals(r1, r2)
    }

    @Test
    fun `LoginRequest copy changes password`() {
        val original = LoginRequest("admin", "old")
        val updated = original.copy(password = "new")
        assertEquals("admin", updated.username)
        assertEquals("new", updated.password)
    }

    // --- LoginResponse ---

    @Test
    fun `LoginResponse stores all fields`() {
        val user = createUser()
        val response = LoginResponse(
            user = user,
            token = "jwt-token",
            refreshToken = "refresh-token",
            expiresIn = 3600
        )
        assertEquals(user, response.user)
        assertEquals("jwt-token", response.token)
        assertEquals("refresh-token", response.refreshToken)
        assertEquals(3600L, response.expiresIn)
    }

    @Test
    fun `LoginResponse copy updates token`() {
        val user = createUser()
        val original = LoginResponse(user = user, token = "old", refreshToken = "ref", expiresIn = 3600)
        val refreshed = original.copy(token = "new-token", expiresIn = 7200)
        assertEquals("new-token", refreshed.token)
        assertEquals(7200L, refreshed.expiresIn)
        assertEquals("ref", refreshed.refreshToken)
    }

    // --- RegisterRequest ---

    @Test
    fun `RegisterRequest stores all fields`() {
        val request = RegisterRequest(
            username = "newuser",
            email = "new@example.com",
            password = "secret",
            firstName = "New",
            lastName = "User"
        )
        assertEquals("newuser", request.username)
        assertEquals("new@example.com", request.email)
        assertEquals("secret", request.password)
        assertEquals("New", request.firstName)
        assertEquals("User", request.lastName)
    }

    // --- AuthStatus ---

    @Test
    fun `AuthStatus authenticated with user`() {
        val user = createUser()
        val status = AuthStatus(
            authenticated = true,
            user = user,
            permissions = listOf("read:media")
        )
        assertTrue(status.authenticated)
        assertNotNull(status.user)
        assertEquals(listOf("read:media"), status.permissions)
        assertNull(status.error)
    }

    @Test
    fun `AuthStatus unauthenticated with error`() {
        val status = AuthStatus(
            authenticated = false,
            error = "Token expired"
        )
        assertFalse(status.authenticated)
        assertNull(status.user)
        assertNull(status.permissions)
        assertEquals("Token expired", status.error)
    }

    @Test
    fun `AuthStatus defaults`() {
        val status = AuthStatus(authenticated = false)
        assertFalse(status.authenticated)
        assertNull(status.user)
        assertNull(status.permissions)
        assertNull(status.error)
    }

    // --- AuthState ---

    @Test
    fun `AuthState defaults to unauthenticated not loading no error`() {
        val state = AuthState()
        assertFalse(state.isAuthenticated)
        assertFalse(state.isLoading)
        assertNull(state.error)
        assertNull(state.user)
    }

    @Test
    fun `AuthState loading state`() {
        val state = AuthState(isLoading = true)
        assertTrue(state.isLoading)
        assertFalse(state.isAuthenticated)
    }

    @Test
    fun `AuthState authenticated with user`() {
        val user = createUser()
        val state = AuthState(isAuthenticated = true, user = user)
        assertTrue(state.isAuthenticated)
        assertNotNull(state.user)
    }

    @Test
    fun `AuthState error state`() {
        val state = AuthState(error = "Invalid credentials")
        assertFalse(state.isAuthenticated)
        assertEquals("Invalid credentials", state.error)
    }

    @Test
    fun `AuthState copy transitions correctly`() {
        val idle = AuthState()
        val loading = idle.copy(isLoading = true)
        val authenticated = loading.copy(isLoading = false, isAuthenticated = true, user = createUser())
        val errored = idle.copy(isLoading = false, error = "Fail")

        assertFalse(idle.isLoading)
        assertTrue(loading.isLoading)
        assertTrue(authenticated.isAuthenticated)
        assertFalse(authenticated.isLoading)
        assertNotNull(authenticated.user)
        assertEquals("Fail", errored.error)
    }

    // --- ChangePasswordRequest ---

    @Test
    fun `ChangePasswordRequest stores passwords`() {
        val request = ChangePasswordRequest(currentPassword = "old", newPassword = "new")
        assertEquals("old", request.currentPassword)
        assertEquals("new", request.newPassword)
    }

    @Test
    fun `ChangePasswordRequest equality`() {
        val r1 = ChangePasswordRequest("old", "new")
        val r2 = ChangePasswordRequest("old", "new")
        assertEquals(r1, r2)
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
    fun `UpdateProfileRequest defaults to nulls`() {
        val request = UpdateProfileRequest()
        assertNull(request.firstName)
        assertNull(request.lastName)
        assertNull(request.email)
    }

    @Test
    fun `UpdateProfileRequest partial update`() {
        val request = UpdateProfileRequest(firstName = "Only First")
        assertEquals("Only First", request.firstName)
        assertNull(request.lastName)
        assertNull(request.email)
    }

    // --- ApiResponse ---

    @Test
    fun `ApiResponse with data and message`() {
        val response = ApiResponse(data = "payload", message = "OK")
        assertEquals("payload", response.data)
        assertEquals("OK", response.message)
        assertNull(response.error)
    }

    @Test
    fun `ApiResponse with error`() {
        val response = ApiResponse<String>(error = "Not found")
        assertNull(response.data)
        assertEquals("Not found", response.error)
    }

    @Test
    fun `ApiResponse all nulls by default`() {
        val response = ApiResponse<Int>()
        assertNull(response.data)
        assertNull(response.error)
        assertNull(response.message)
    }

    // --- ErrorResponse ---

    @Test
    fun `ErrorResponse with full data`() {
        val details = mapOf("field" to "email", "reason" to "invalid_format")
        val response = ErrorResponse(error = "Validation error", code = 422, details = details)
        assertEquals("Validation error", response.error)
        assertEquals(422, response.code)
        assertEquals("email", response.details?.get("field"))
    }

    @Test
    fun `ErrorResponse minimal`() {
        val response = ErrorResponse(error = "Internal server error")
        assertEquals("Internal server error", response.error)
        assertNull(response.code)
        assertNull(response.details)
    }

    @Test
    fun `ErrorResponse with empty details map`() {
        val response = ErrorResponse(error = "Error", details = emptyMap())
        assertNotNull(response.details)
        assertTrue(response.details!!.isEmpty())
    }

    // --- UserRole ---

    @Test
    fun `UserRole fromValue returns correct role`() {
        assertEquals(UserRole.ADMIN, UserRole.fromValue("admin"))
        assertEquals(UserRole.MODERATOR, UserRole.fromValue("moderator"))
        assertEquals(UserRole.USER, UserRole.fromValue("user"))
        assertEquals(UserRole.VIEWER, UserRole.fromValue("viewer"))
    }

    @Test
    fun `UserRole fromValue returns USER for unknown value`() {
        assertEquals(UserRole.USER, UserRole.fromValue("superadmin"))
        assertEquals(UserRole.USER, UserRole.fromValue(""))
        assertEquals(UserRole.USER, UserRole.fromValue("unknown"))
    }

    @Test
    fun `UserRole displayName values`() {
        assertEquals("Administrator", UserRole.ADMIN.displayName)
        assertEquals("Moderator", UserRole.MODERATOR.displayName)
        assertEquals("User", UserRole.USER.displayName)
        assertEquals("Viewer", UserRole.VIEWER.displayName)
    }

    @Test
    fun `UserRole value matches API string`() {
        assertEquals("admin", UserRole.ADMIN.value)
        assertEquals("moderator", UserRole.MODERATOR.value)
        assertEquals("user", UserRole.USER.value)
        assertEquals("viewer", UserRole.VIEWER.value)
    }

    // --- Permissions ---

    @Test
    fun `Permissions constants have correct values`() {
        assertEquals("read:media", Permissions.READ_MEDIA)
        assertEquals("write:media", Permissions.WRITE_MEDIA)
        assertEquals("delete:media", Permissions.DELETE_MEDIA)
        assertEquals("read:catalog", Permissions.READ_CATALOG)
        assertEquals("write:catalog", Permissions.WRITE_CATALOG)
        assertEquals("delete:catalog", Permissions.DELETE_CATALOG)
        assertEquals("trigger:analysis", Permissions.TRIGGER_ANALYSIS)
        assertEquals("view:analysis", Permissions.VIEW_ANALYSIS)
        assertEquals("manage:users", Permissions.MANAGE_USERS)
        assertEquals("manage:roles", Permissions.MANAGE_ROLES)
        assertEquals("view:logs", Permissions.VIEW_LOGS)
        assertEquals("admin:system", Permissions.SYSTEM_ADMIN)
        assertEquals("access:api", Permissions.API_ACCESS)
        assertEquals("write:api", Permissions.API_WRITE)
    }

    // --- Helpers ---

    private fun createUser(
        id: Long = 1L,
        username: String = "testuser",
        email: String = "test@example.com",
        firstName: String = "Test",
        lastName: String = "User",
        role: String = "user",
        isActive: Boolean = true,
        lastLogin: String? = null,
        createdAt: String = "2025-01-01T00:00:00Z",
        updatedAt: String = "2025-01-01T00:00:00Z",
        permissions: List<String>? = null
    ): User = User(
        id = id,
        username = username,
        email = email,
        firstName = firstName,
        lastName = lastName,
        role = role,
        isActive = isActive,
        lastLogin = lastLogin,
        createdAt = createdAt,
        updatedAt = updatedAt,
        permissions = permissions
    )
}
