package com.catalogizer.android

import android.content.Context
import com.catalogizer.android.ui.viewmodel.AuthViewModel
import com.catalogizer.android.ui.viewmodel.HomeViewModel
import com.catalogizer.android.ui.viewmodel.MainViewModel
import com.catalogizer.android.ui.viewmodel.SearchViewModel
import io.mockk.*
import kotlinx.coroutines.ExperimentalCoroutinesApi
import org.junit.After
import org.junit.Assert.*
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.runner.RunWith
import org.robolectric.RobolectricTestRunner
import java.io.File

@OptIn(ExperimentalCoroutinesApi::class)
@RunWith(RobolectricTestRunner::class)
class DependencyContainerTest {

    @get:Rule
    val mainDispatcherRule = MainDispatcherRule()

    private lateinit var mockContext: Context

    @Before
    fun setup() {
        mockContext = mockk(relaxed = true)
        every { mockContext.applicationContext } returns mockContext
        every { mockContext.filesDir } returns File("/tmp/test-files")

        // Reset the singleton between tests
        resetSingleton()
    }

    @After
    fun tearDown() {
        clearAllMocks()
        resetSingleton()
    }

    private fun resetSingleton() {
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
    fun `constructor should create valid instance`() {
        val container = DependencyContainer(mockContext)

        assertNotNull(container)
    }

    @Test
    fun `authRepository should return non-null instance`() {
        val container = DependencyContainer(mockContext)

        val authRepository = container.authRepository

        assertNotNull(authRepository)
    }

    @Test
    fun `authRepository should return same instance on multiple accesses`() {
        val container = DependencyContainer(mockContext)

        val repo1 = container.authRepository
        val repo2 = container.authRepository

        assertSame(repo1, repo2)
    }

    @Test
    fun `mediaRepository should return non-null instance`() {
        val container = DependencyContainer(mockContext)

        val mediaRepository = container.mediaRepository

        assertNotNull(mediaRepository)
    }

    @Test
    fun `mediaRepository should return same instance on multiple accesses`() {
        val container = DependencyContainer(mockContext)

        val repo1 = container.mediaRepository
        val repo2 = container.mediaRepository

        assertSame(repo1, repo2)
    }

    @Test
    fun `syncManager should return non-null instance`() {
        val container = DependencyContainer(mockContext)

        val syncManager = container.syncManager

        assertNotNull(syncManager)
    }

    @Test
    fun `syncManager should return same instance on multiple accesses`() {
        val container = DependencyContainer(mockContext)

        val sm1 = container.syncManager
        val sm2 = container.syncManager

        assertSame(sm1, sm2)
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
    fun `createAuthViewModel should return AuthViewModel type`() {
        val container = DependencyContainer(mockContext)

        val vm = container.createAuthViewModel()

        assertTrue(vm is AuthViewModel)
    }

    @Test
    fun `createMainViewModel should return new instance each time`() {
        val container = DependencyContainer(mockContext)

        val vm1 = container.createMainViewModel()
        val vm2 = container.createMainViewModel()

        assertNotNull(vm1)
        assertNotNull(vm2)
        assertNotSame(vm1, vm2)
    }

    @Test
    fun `createMainViewModel should return MainViewModel type`() {
        val container = DependencyContainer(mockContext)

        val vm = container.createMainViewModel()

        assertTrue(vm is MainViewModel)
    }

    @Test
    fun `createHomeViewModel should return new instance each time`() {
        val container = DependencyContainer(mockContext)

        val vm1 = container.createHomeViewModel()
        val vm2 = container.createHomeViewModel()

        assertNotNull(vm1)
        assertNotNull(vm2)
        assertNotSame(vm1, vm2)
    }

    @Test
    fun `createHomeViewModel should return HomeViewModel type`() {
        val container = DependencyContainer(mockContext)

        val vm = container.createHomeViewModel()

        assertTrue(vm is HomeViewModel)
    }

    @Test
    fun `createSearchViewModel should return new instance each time`() {
        val container = DependencyContainer(mockContext)

        val vm1 = container.createSearchViewModel()
        val vm2 = container.createSearchViewModel()

        assertNotNull(vm1)
        assertNotNull(vm2)
        assertNotSame(vm1, vm2)
    }

    @Test
    fun `createSearchViewModel should return SearchViewModel type`() {
        val container = DependencyContainer(mockContext)

        val vm = container.createSearchViewModel()

        assertTrue(vm is SearchViewModel)
    }

    @Test
    fun `getInstance is thread-safe with synchronized block`() {
        val instances = mutableListOf<DependencyContainer>()

        // Simulate concurrent access
        val threads = (1..10).map {
            Thread {
                val instance = DependencyContainer.getInstance(mockContext)
                synchronized(instances) {
                    instances.add(instance)
                }
            }
        }

        threads.forEach { it.start() }
        threads.forEach { it.join() }

        // All instances should be the same
        val firstInstance = instances.first()
        instances.forEach { assertSame(firstInstance, it) }
    }
}
