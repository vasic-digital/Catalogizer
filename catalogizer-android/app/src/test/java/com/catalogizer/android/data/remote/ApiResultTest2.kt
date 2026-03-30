package com.catalogizer.android.data.remote

import org.junit.Assert.*
import org.junit.Test

class ApiResultTest2 {

    @Test
    fun `success creates result with data and isSuccess true`() {
        val result = ApiResult.success("hello")
        assertEquals("hello", result.data)
        assertNull(result.error)
        assertTrue(result.isSuccess)
    }

    @Test
    fun `error creates result with error message and isSuccess false`() {
        val result = ApiResult.error<String>("Something went wrong")
        assertNull(result.data)
        assertEquals("Something went wrong", result.error)
        assertFalse(result.isSuccess)
    }

    @Test
    fun `success with complex type`() {
        val data = mapOf("key1" to "value1", "key2" to "value2")
        val result = ApiResult.success(data)
        assertEquals(2, result.data?.size)
        assertTrue(result.isSuccess)
    }

    @Test
    fun `success with list type`() {
        val data = listOf(1, 2, 3, 4, 5)
        val result = ApiResult.success(data)
        assertEquals(5, result.data?.size)
        assertTrue(result.isSuccess)
    }

    @Test
    fun `success with null data reports not successful`() {
        val result = ApiResult<String>(data = null, error = null)
        assertFalse(result.isSuccess)
    }

    @Test
    fun `result with both data and error reports not successful`() {
        val result = ApiResult(data = "data", error = "error")
        assertFalse(result.isSuccess)
    }

    @Test
    fun `error with empty message`() {
        val result = ApiResult.error<String>("")
        assertEquals("", result.error)
        assertFalse(result.isSuccess)
    }

    @Test
    fun `success with Unit type`() {
        val result = ApiResult.success(Unit)
        assertNotNull(result.data)
        assertTrue(result.isSuccess)
    }
}
