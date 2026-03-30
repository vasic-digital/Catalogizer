package com.catalogizer.android.data.models

import org.junit.Assert.*
import org.junit.Test

class PermissionsTest {

    @Test
    fun `media permissions have correct values`() {
        assertEquals("read:media", Permissions.READ_MEDIA)
        assertEquals("write:media", Permissions.WRITE_MEDIA)
        assertEquals("delete:media", Permissions.DELETE_MEDIA)
    }

    @Test
    fun `catalog permissions have correct values`() {
        assertEquals("read:catalog", Permissions.READ_CATALOG)
        assertEquals("write:catalog", Permissions.WRITE_CATALOG)
        assertEquals("delete:catalog", Permissions.DELETE_CATALOG)
    }

    @Test
    fun `analysis permissions have correct values`() {
        assertEquals("trigger:analysis", Permissions.TRIGGER_ANALYSIS)
        assertEquals("view:analysis", Permissions.VIEW_ANALYSIS)
    }

    @Test
    fun `admin permissions have correct values`() {
        assertEquals("manage:users", Permissions.MANAGE_USERS)
        assertEquals("manage:roles", Permissions.MANAGE_ROLES)
        assertEquals("view:logs", Permissions.VIEW_LOGS)
        assertEquals("admin:system", Permissions.SYSTEM_ADMIN)
    }

    @Test
    fun `api permissions have correct values`() {
        assertEquals("access:api", Permissions.API_ACCESS)
        assertEquals("write:api", Permissions.API_WRITE)
    }

    @Test
    fun `all permissions follow action colon resource format`() {
        val allPermissions = listOf(
            Permissions.READ_MEDIA, Permissions.WRITE_MEDIA, Permissions.DELETE_MEDIA,
            Permissions.READ_CATALOG, Permissions.WRITE_CATALOG, Permissions.DELETE_CATALOG,
            Permissions.TRIGGER_ANALYSIS, Permissions.VIEW_ANALYSIS,
            Permissions.MANAGE_USERS, Permissions.MANAGE_ROLES, Permissions.VIEW_LOGS,
            Permissions.SYSTEM_ADMIN, Permissions.API_ACCESS, Permissions.API_WRITE
        )
        for (perm in allPermissions) {
            assertTrue("Permission '$perm' should contain ':'", perm.contains(":"))
            val parts = perm.split(":")
            assertEquals("Permission '$perm' should have exactly two parts", 2, parts.size)
            assertTrue("Permission '$perm' action part should not be empty", parts[0].isNotEmpty())
            assertTrue("Permission '$perm' resource part should not be empty", parts[1].isNotEmpty())
        }
    }
}
