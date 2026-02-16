package com.catalogizer.android.data.models

import org.junit.Assert.*
import org.junit.Test

class UserTest {

    private fun createTestUser(
        id: Long = 1L,
        username: String = "testuser",
        email: String = "test@example.com",
        firstName: String = "John",
        lastName: String = "Doe",
        role: String = "user",
        isActive: Boolean = true,
        lastLogin: String? = null,
        createdAt: String = "2025-01-01T00:00:00Z",
        updatedAt: String = "2025-06-01T00:00:00Z",
        permissions: List<String>? = null
    ) = User(
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

    @Test
    fun `fullName returns trimmed first and last name`() {
        val user = createTestUser(firstName = "John", lastName = "Doe")
        assertEquals("John Doe", user.fullName)
    }

    @Test
    fun `fullName with only first name trims correctly`() {
        val user = createTestUser(firstName = "John", lastName = "")
        assertEquals("John", user.fullName)
    }

    @Test
    fun `fullName with only last name trims correctly`() {
        val user = createTestUser(firstName = "", lastName = "Doe")
        assertEquals("Doe", user.fullName)
    }

    @Test
    fun `fullName with empty names returns empty string`() {
        val user = createTestUser(firstName = "", lastName = "")
        assertEquals("", user.fullName)
    }

    @Test
    fun `isAdmin returns true for admin role`() {
        val admin = createTestUser(role = "admin")
        assertTrue(admin.isAdmin)
    }

    @Test
    fun `isAdmin returns false for non-admin roles`() {
        val user = createTestUser(role = "user")
        assertFalse(user.isAdmin)

        val moderator = createTestUser(role = "moderator")
        assertFalse(moderator.isAdmin)

        val viewer = createTestUser(role = "viewer")
        assertFalse(viewer.isAdmin)
    }

    @Test
    fun `User equality works correctly`() {
        val user1 = createTestUser(id = 1)
        val user2 = createTestUser(id = 1)
        val user3 = createTestUser(id = 2)

        assertEquals(user1, user2)
        assertNotEquals(user1, user3)
    }

    @Test
    fun `User copy creates independent instance`() {
        val original = createTestUser()
        val copy = original.copy(username = "newuser", role = "admin")

        assertEquals("newuser", copy.username)
        assertEquals("admin", copy.role)
        assertTrue(copy.isAdmin)
        assertEquals("testuser", original.username)
    }

    @Test
    fun `User optional fields default to null`() {
        val user = createTestUser()
        assertNull(user.lastLogin)
        assertNull(user.permissions)
    }

    @Test
    fun `User with permissions returns correct list`() {
        val permissions = listOf("read:media", "write:media", "read:catalog")
        val user = createTestUser(permissions = permissions)

        assertNotNull(user.permissions)
        assertEquals(3, user.permissions?.size)
        assertTrue(user.permissions?.contains("read:media") == true)
        assertTrue(user.permissions?.contains("write:media") == true)
    }
}
