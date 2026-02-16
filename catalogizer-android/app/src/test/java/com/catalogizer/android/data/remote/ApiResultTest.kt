package com.catalogizer.android.data.remote

import org.junit.Assert.*
import org.junit.Test

class ApiResultTest {

    @Test
    fun `success creates result with data and no error`() {
        val result = ApiResult.success("test data")

        assertTrue(result.isSuccess)
        assertEquals("test data", result.data)
        assertNull(result.error)
    }

    @Test
    fun `error creates result with error and no data`() {
        val result = ApiResult.error<String>("Something went wrong")

        assertFalse(result.isSuccess)
        assertNull(result.data)
        assertEquals("Something went wrong", result.error)
    }

    @Test
    fun `isSuccess is true when data is present and no error`() {
        val result = ApiResult(data = 42)
        assertTrue(result.isSuccess)
    }

    @Test
    fun `isSuccess is false when data is null`() {
        val result = ApiResult<Int>(data = null)
        assertFalse(result.isSuccess)
    }

    @Test
    fun `isSuccess is false when error is present`() {
        val result = ApiResult(data = 42, error = "error")
        assertFalse(result.isSuccess)
    }

    @Test
    fun `isSuccess is false when both data and error are null`() {
        val result = ApiResult<String>()
        assertFalse(result.isSuccess)
    }

    @Test
    fun `success with complex type`() {
        val list = listOf("a", "b", "c")
        val result = ApiResult.success(list)

        assertTrue(result.isSuccess)
        assertEquals(3, result.data?.size)
        assertEquals("a", result.data?.get(0))
    }

    @Test
    fun `error with complex type parameter`() {
        val result = ApiResult.error<List<String>>("Network error")

        assertFalse(result.isSuccess)
        assertNull(result.data)
        assertEquals("Network error", result.error)
    }

    @Test
    fun `success with Unit type`() {
        val result = ApiResult.success(Unit)
        assertTrue(result.isSuccess)
        assertNotNull(result.data)
    }
}
