package com.catalogizer.androidtv

import android.content.Context
import io.mockk.*
import org.junit.After
import org.junit.Assert.*
import org.junit.Before
import org.junit.Test
import java.io.File

class DependencyContainerTest {

    private lateinit var mockContext: Context

    @Before
    fun setup() {
        mockContext = mockk(relaxed = true)
        every { mockContext.applicationContext } returns mockContext
        every { mockContext.filesDir } returns File("/tmp/test-files")

        // Reset the singleton between tests
        val field = DependencyContainer::class.java.getDeclaredField("instance")
        field.isAccessible = true
        field.set(null, null)
    }

    @After
    fun tearDown() {
        clearAllMocks()
        // Reset singleton
        val field = DependencyContainer::class.java.getDeclaredField("instance")
        field.isAccessible = true
        field.set(null, null)
    }

    @Test
    fun `getInstance should return same instance for same context`() {
        val instance1 = DependencyContainer.getInstance(mockContext)
        val instance2 = DependencyContainer.getInstance(mockContext)

        assertSame(instance1, instance2)
    }

    @Test
    fun `getInstance should use application context`() {
        val activityContext = mockk<Context>(relaxed = true)
        every { activityContext.applicationContext } returns mockContext

        DependencyContainer.getInstance(activityContext)

        verify { activityContext.applicationContext }
    }

    @Test
    fun `DependencyContainer should expose authRepository`() {
        val container = DependencyContainer(mockContext)

        val authRepository = container.authRepository

        assertNotNull(authRepository)
    }

    @Test
    fun `createAuthViewModel should return new instance each time`() {
        val container = DependencyContainer(mockContext)

        val vm1 = container.createAuthViewModel()
        val vm2 = container.createAuthViewModel()

        assertNotNull(vm1)
        assertNotNull(vm2)
        assertNotSame(vm1, vm2)
    }

    @Test
    fun `DependencyContainer constructor should accept context`() {
        val container = DependencyContainer(mockContext)

        assertNotNull(container)
    }
}
