package com.catalogizer.android.data.models

import org.junit.Assert.*
import org.junit.Test

class UserRoleTest {

    @Test
    fun `fromValue returns ADMIN for admin string`() {
        assertEquals(UserRole.ADMIN, UserRole.fromValue("admin"))
    }

    @Test
    fun `fromValue returns MODERATOR for moderator string`() {
        assertEquals(UserRole.MODERATOR, UserRole.fromValue("moderator"))
    }

    @Test
    fun `fromValue returns USER for user string`() {
        assertEquals(UserRole.USER, UserRole.fromValue("user"))
    }

    @Test
    fun `fromValue returns VIEWER for viewer string`() {
        assertEquals(UserRole.VIEWER, UserRole.fromValue("viewer"))
    }

    @Test
    fun `fromValue defaults to USER for unknown role`() {
        assertEquals(UserRole.USER, UserRole.fromValue("unknown"))
    }

    @Test
    fun `fromValue defaults to USER for empty string`() {
        assertEquals(UserRole.USER, UserRole.fromValue(""))
    }

    @Test
    fun `role display names are correct`() {
        assertEquals("Administrator", UserRole.ADMIN.displayName)
        assertEquals("Moderator", UserRole.MODERATOR.displayName)
        assertEquals("User", UserRole.USER.displayName)
        assertEquals("Viewer", UserRole.VIEWER.displayName)
    }

    @Test
    fun `role values are correct`() {
        assertEquals("admin", UserRole.ADMIN.value)
        assertEquals("moderator", UserRole.MODERATOR.value)
        assertEquals("user", UserRole.USER.value)
        assertEquals("viewer", UserRole.VIEWER.value)
    }

    @Test
    fun `all roles are present`() {
        val roles = UserRole.values()
        assertEquals(4, roles.size)
    }
}
