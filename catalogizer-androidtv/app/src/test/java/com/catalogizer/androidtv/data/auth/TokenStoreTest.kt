package com.catalogizer.androidtv.data.auth

import android.content.Context
import androidx.test.core.app.ApplicationProvider
import kotlinx.coroutines.runBlocking
import org.junit.After
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertNull
import org.junit.Assert.assertTrue
import org.junit.Before
import org.junit.Test
import org.junit.runner.RunWith
import org.robolectric.RobolectricTestRunner
import org.robolectric.annotation.Config

@RunWith(RobolectricTestRunner::class)
@Config(sdk = [33])
class TokenStoreTest {

    private lateinit var store: TokenStore

    @Before
    fun setUp() {
        val ctx = ApplicationProvider.getApplicationContext<Context>()
        val prefs = ctx.getSharedPreferences("token_store_test", Context.MODE_PRIVATE)
        store = TokenStore.inMemory(prefs)
        runBlocking { store.clear() }
    }

    @After
    fun tearDown() {
        runBlocking { store.clear() }
    }

    @Test
    fun `save and load persists token`() = runBlocking {
        val saved = TokenStore.Record(
            token = "abc.def.ghi",
            username = "admin",
            userId = 1,
            expiresAtMs = 1_800_000_000_000L,
        )
        store.save(saved)
        val loaded = store.load()
        assertEquals(saved, loaded)
    }

    @Test
    fun `load returns null when empty`() = runBlocking {
        assertNull(store.load())
    }

    @Test
    fun `clear wipes record`() = runBlocking {
        store.save(TokenStore.Record("t", "u", 1, 1_800_000_000_000L))
        store.clear()
        assertNull(store.load())
    }

    @Test
    fun `isAuthenticated reflects presence and expiry`() = runBlocking {
        assertFalse(store.isAuthenticated(nowMs = 0L))

        val future = System.currentTimeMillis() + 60_000L
        store.save(TokenStore.Record("t", "u", 1, future))
        assertTrue(store.isAuthenticated(nowMs = System.currentTimeMillis()))

        val past = System.currentTimeMillis() - 60_000L
        store.save(TokenStore.Record("t", "u", 1, past))
        assertFalse(store.isAuthenticated(nowMs = System.currentTimeMillis()))
    }
}
