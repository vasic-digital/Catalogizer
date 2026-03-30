package com.catalogizer.android.data.models

import org.junit.Assert.*
import org.junit.Test

class UserTest2 {

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
    fun `fullName returns combined first and last name`() {
        val user = createUser(firstName = "John", lastName = "Doe")
        assertEquals("John Doe", user.fullName)
    }

    @Test
    fun `fullName trims whitespace when lastName is empty`() {
        val user = createUser(firstName = "John", lastName = "")
        assertEquals("John", user.fullName)
    }

    @Test
    fun `fullName trims whitespace when firstName is empty`() {
        val user = createUser(firstName = "", lastName = "Doe")
        assertEquals("Doe", user.fullName)
    }

    @Test
    fun `fullName returns empty when both names are empty`() {
        val user = createUser(firstName = "", lastName = "")
        assertEquals("", user.fullName)
    }

    @Test
    fun `isAdmin returns true for admin role`() {
        val user = createUser(role = "admin")
        assertTrue(user.isAdmin)
    }

    @Test
    fun `isAdmin returns false for user role`() {
        val user = createUser(role = "user")
        assertFalse(user.isAdmin)
    }

    @Test
    fun `isAdmin returns false for moderator role`() {
        val user = createUser(role = "moderator")
        assertFalse(user.isAdmin)
    }

    @Test
    fun `isAdmin returns false for viewer role`() {
        val user = createUser(role = "viewer")
        assertFalse(user.isAdmin)
    }

    @Test
    fun `user with null lastLogin`() {
        val user = createUser(lastLogin = null)
        assertNull(user.lastLogin)
    }

    @Test
    fun `user with lastLogin set`() {
        val user = createUser(lastLogin = "2025-06-15T10:30:00Z")
        assertEquals("2025-06-15T10:30:00Z", user.lastLogin)
    }

    @Test
    fun `user with null permissions`() {
        val user = createUser(permissions = null)
        assertNull(user.permissions)
    }

    @Test
    fun `user with empty permissions`() {
        val user = createUser(permissions = emptyList())
        assertNotNull(user.permissions)
        assertTrue(user.permissions!!.isEmpty())
    }

    @Test
    fun `user with multiple permissions`() {
        val perms = listOf("read:media", "write:media", "read:catalog")
        val user = createUser(permissions = perms)
        assertEquals(3, user.permissions?.size)
        assertTrue(user.permissions!!.contains("read:media"))
    }

    @Test
    fun `user equality based on all fields`() {
        val user1 = createUser(id = 1)
        val user2 = createUser(id = 1)
        assertEquals(user1, user2)
    }

    @Test
    fun `user inequality based on id`() {
        val user1 = createUser(id = 1)
        val user2 = createUser(id = 2)
        assertNotEquals(user1, user2)
    }
}
