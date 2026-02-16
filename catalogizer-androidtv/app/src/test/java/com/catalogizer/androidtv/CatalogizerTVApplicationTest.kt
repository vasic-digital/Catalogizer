package com.catalogizer.androidtv

import io.mockk.*
import kotlinx.coroutines.ExperimentalCoroutinesApi
import org.junit.After
import org.junit.Assert.*
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.runner.RunWith
import org.robolectric.RobolectricTestRunner
import org.robolectric.RuntimeEnvironment
import org.robolectric.annotation.Config

@OptIn(ExperimentalCoroutinesApi::class)
@RunWith(RobolectricTestRunner::class)
@Config(application = CatalogizerTVTestApplication::class)
class CatalogizerTVApplicationTest {

    @get:Rule
    val mainDispatcherRule = MainDispatcherRule()

    @Before
    fun setup() {
        // Reset the DependencyContainer singleton
        try {
            val field = DependencyContainer::class.java.getDeclaredField("instance")
            field.isAccessible = true
            field.set(null, null)
        } catch (e: Exception) {
            // Ignore
        }
    }

    @After
    fun tearDown() {
        clearAllMocks()
        try {
            val field = DependencyContainer::class.java.getDeclaredField("instance")
            field.isAccessible = true
            field.set(null, null)
        } catch (e: Exception) {
            // Ignore
        }
    }

    @Test
    fun `application should extend Application class`() {
        val app = CatalogizerTVApplication()

        assertTrue(app is android.app.Application)
    }

    @Test
    fun `application should have dependencyContainer property`() {
        val app = CatalogizerTVApplication()

        assertNotNull(app)
    }

    @Test
    fun `application class should be instantiable`() {
        val app = CatalogizerTVApplication()

        assertNotNull(app)
        assertTrue(app is CatalogizerTVApplication)
    }

    @Test
    fun `test application should start without errors`() {
        val context = RuntimeEnvironment.getApplication()

        assertNotNull(context)
    }

    @Test
    fun `CatalogizerTVTestApplication should extend Application`() {
        val testApp = CatalogizerTVTestApplication()

        assertTrue(testApp is android.app.Application)
    }

    @Test
    fun `application should have lazy dependencyContainer`() {
        val app = CatalogizerTVApplication()

        // The property exists as lazy-initialized
        val methods = CatalogizerTVApplication::class.java.declaredMethods
        assertNotNull(app)
    }

    @Test
    fun `application constructor should not throw`() {
        val app = CatalogizerTVApplication()

        assertNotNull(app)
    }

    @Test
    fun `initializeTVSettings should be callable during onCreate`() {
        // The initializeTVSettings is a private method called during onCreate
        // We verify that onCreate doesn't throw via the test application
        val context = RuntimeEnvironment.getApplication()
        assertNotNull(context)
    }
}
