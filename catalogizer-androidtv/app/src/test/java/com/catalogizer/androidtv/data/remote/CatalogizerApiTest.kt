package com.catalogizer.androidtv.data.remote

import org.junit.Assert.assertEquals
import org.junit.Assert.assertNotNull
import org.junit.Assert.assertTrue
import org.junit.Test
import retrofit2.http.Body
import retrofit2.http.DELETE
import retrofit2.http.GET
import retrofit2.http.POST
import retrofit2.http.PUT
import retrofit2.http.Path
import retrofit2.http.Query
import retrofit2.http.QueryMap
import java.lang.reflect.Method

/**
 * Validates that [CatalogizerApi] declares every expected endpoint
 * with correct HTTP method, path, and parameter annotations. This
 * catches accidental renames or annotation removals at compile-time
 * speed (no network, no mock server).
 */
class CatalogizerApiTest {

    private val apiClass = CatalogizerApi::class.java

    // ------------------------------------------------------------------
    // Helper: find a declared method by name (Kotlin suspend functions
    // compile to a method with an extra Continuation parameter, so we
    // match by name only).
    // ------------------------------------------------------------------
    private fun findMethod(name: String): Method {
        val found = apiClass.declaredMethods.filter { it.name == name }
        assertTrue("Expected at least one method named '$name'", found.isNotEmpty())
        return found.first()
    }

    // ------------------------------------------------------------------
    // Authentication endpoints
    // ------------------------------------------------------------------

    @Test
    fun `login endpoint is POST with correct path`() {
        val method = findMethod("login")
        val post = method.getAnnotation(POST::class.java)
        assertNotNull("login must be annotated with @POST", post)
        assertEquals("api/v1/auth/login", post!!.value)
    }

    @Test
    fun `login accepts Body parameter`() {
        val method = findMethod("login")
        val hasBody = method.parameterAnnotations.any { annotations ->
            annotations.any { it is Body }
        }
        assertTrue("login must have a @Body parameter", hasBody)
    }

    @Test
    fun `refreshToken endpoint is POST with correct path`() {
        val method = findMethod("refreshToken")
        val post = method.getAnnotation(POST::class.java)
        assertNotNull("refreshToken must be annotated with @POST", post)
        assertEquals("api/v1/auth/refresh", post!!.value)
    }

    @Test
    fun `logout endpoint is POST with correct path`() {
        val method = findMethod("logout")
        val post = method.getAnnotation(POST::class.java)
        assertNotNull("logout must be annotated with @POST", post)
        assertEquals("api/v1/auth/logout", post!!.value)
    }

    // ------------------------------------------------------------------
    // Catalog & search endpoints
    // ------------------------------------------------------------------

    @Test
    fun `getCatalog endpoint is GET with correct path`() {
        val method = findMethod("getCatalog")
        val get = method.getAnnotation(GET::class.java)
        assertNotNull("getCatalog must be annotated with @GET", get)
        assertEquals("api/v1/catalog", get!!.value)
    }

    @Test
    fun `searchMedia endpoint is GET with QueryMap`() {
        val method = findMethod("searchMedia")
        val get = method.getAnnotation(GET::class.java)
        assertNotNull("searchMedia must be annotated with @GET", get)
        assertEquals("api/v1/media/search", get!!.value)
        val hasQueryMap = method.parameterAnnotations.any { annotations ->
            annotations.any { it is QueryMap }
        }
        assertTrue("searchMedia must have a @QueryMap parameter", hasQueryMap)
    }

    @Test
    fun `searchEntities endpoint is GET with correct path`() {
        val method = findMethod("searchEntities")
        val get = method.getAnnotation(GET::class.java)
        assertNotNull(get)
        assertEquals("api/v1/entities", get!!.value)
    }

    @Test
    fun `browseEntities endpoint uses Path for type`() {
        val method = findMethod("browseEntities")
        val get = method.getAnnotation(GET::class.java)
        assertNotNull(get)
        assertEquals("api/v1/entities/browse/{type}", get!!.value)
        val hasPath = method.parameterAnnotations.any { annotations ->
            annotations.any { it is Path && it.value == "type" }
        }
        assertTrue("browseEntities must have @Path(\"type\")", hasPath)
    }

    @Test
    fun `getEntityById endpoint uses Path for id`() {
        val method = findMethod("getEntityById")
        val get = method.getAnnotation(GET::class.java)
        assertNotNull(get)
        assertEquals("api/v1/entities/{id}", get!!.value)
    }

    @Test
    fun `getEntityChildren endpoint is GET with path id`() {
        val method = findMethod("getEntityChildren")
        val get = method.getAnnotation(GET::class.java)
        assertNotNull(get)
        assertEquals("api/v1/entities/{id}/children", get!!.value)
    }

    @Test
    fun `getEntityStream endpoint is GET with path id`() {
        val method = findMethod("getEntityStream")
        val get = method.getAnnotation(GET::class.java)
        assertNotNull(get)
        assertEquals("api/v1/entities/{id}/stream", get!!.value)
    }

    // ------------------------------------------------------------------
    // Playback session endpoints
    // ------------------------------------------------------------------

    @Test
    fun `startPlaybackSession endpoint is POST`() {
        val method = findMethod("startPlaybackSession")
        val post = method.getAnnotation(POST::class.java)
        assertNotNull(post)
        assertEquals("api/v1/playback/sessions/start", post!!.value)
    }

    @Test
    fun `progressPlaybackSession endpoint is POST`() {
        val method = findMethod("progressPlaybackSession")
        val post = method.getAnnotation(POST::class.java)
        assertNotNull(post)
        assertEquals("api/v1/playback/sessions/progress", post!!.value)
    }

    @Test
    fun `endPlaybackSession endpoint is POST`() {
        val method = findMethod("endPlaybackSession")
        val post = method.getAnnotation(POST::class.java)
        assertNotNull(post)
        assertEquals("api/v1/playback/sessions/end", post!!.value)
    }

    @Test
    fun `getEntityProgress endpoint is GET with path id`() {
        val method = findMethod("getEntityProgress")
        val get = method.getAnnotation(GET::class.java)
        assertNotNull(get)
        assertEquals("api/v1/entities/{id}/progress", get!!.value)
    }

    @Test
    fun `getEntityHistory endpoint is GET with path id and limit query`() {
        val method = findMethod("getEntityHistory")
        val get = method.getAnnotation(GET::class.java)
        assertNotNull(get)
        assertEquals("api/v1/entities/{id}/history", get!!.value)
        val hasQuery = method.parameterAnnotations.any { annotations ->
            annotations.any { it is Query && it.value == "limit" }
        }
        assertTrue("getEntityHistory must have @Query(\"limit\")", hasQuery)
    }

    // ------------------------------------------------------------------
    // Legacy media endpoints
    // ------------------------------------------------------------------

    @Test
    fun `getMediaById endpoint is GET with path id`() {
        val method = findMethod("getMediaById")
        val get = method.getAnnotation(GET::class.java)
        assertNotNull(get)
        assertEquals("api/v1/media/{id}", get!!.value)
    }

    @Test
    fun `getMediaInfo endpoint is GET with path`() {
        val method = findMethod("getMediaInfo")
        val get = method.getAnnotation(GET::class.java)
        assertNotNull(get)
        assertEquals("api/v1/catalog-info/{path}", get!!.value)
    }

    @Test
    fun `recognizeMedia endpoint is POST`() {
        val method = findMethod("recognizeMedia")
        val post = method.getAnnotation(POST::class.java)
        assertNotNull(post)
        assertEquals("api/v1/media/recognize", post!!.value)
    }

    @Test
    fun `updateWatchProgress endpoint is PUT with path id`() {
        val method = findMethod("updateWatchProgress")
        val put = method.getAnnotation(PUT::class.java)
        assertNotNull(put)
        assertEquals("api/v1/media/{id}/progress", put!!.value)
    }

    // ------------------------------------------------------------------
    // Favorites endpoints
    // ------------------------------------------------------------------

    @Test
    fun `addFavorite endpoint is POST`() {
        val method = findMethod("addFavorite")
        val post = method.getAnnotation(POST::class.java)
        assertNotNull(post)
        assertEquals("api/v1/favorites", post!!.value)
    }

    @Test
    fun `removeFavorite endpoint is DELETE with entity_type and entity_id`() {
        val method = findMethod("removeFavorite")
        val delete = method.getAnnotation(DELETE::class.java)
        assertNotNull(delete)
        assertEquals("api/v1/favorites/{entity_type}/{entity_id}", delete!!.value)
    }

    @Test
    fun `checkFavorite endpoint is GET with entity_type and entity_id`() {
        val method = findMethod("checkFavorite")
        val get = method.getAnnotation(GET::class.java)
        assertNotNull(get)
        assertEquals("api/v1/favorites/check/{entity_type}/{entity_id}", get!!.value)
    }

    @Test
    fun `listFavorites endpoint is GET`() {
        val method = findMethod("listFavorites")
        val get = method.getAnnotation(GET::class.java)
        assertNotNull(get)
        assertEquals("api/v1/favorites", get!!.value)
    }

    // ------------------------------------------------------------------
    // Stats endpoint
    // ------------------------------------------------------------------

    @Test
    fun `getEntityStats endpoint is GET`() {
        val method = findMethod("getEntityStats")
        val get = method.getAnnotation(GET::class.java)
        assertNotNull(get)
        assertEquals("api/v1/entities/stats", get!!.value)
    }

    // ------------------------------------------------------------------
    // Collections endpoints
    // ------------------------------------------------------------------

    @Test
    fun `getCollections endpoint is GET`() {
        val method = findMethod("getCollections")
        val get = method.getAnnotation(GET::class.java)
        assertNotNull(get)
        assertEquals("api/v1/collections", get!!.value)
    }

    @Test
    fun `getCollection endpoint is GET with path id`() {
        val method = findMethod("getCollection")
        val get = method.getAnnotation(GET::class.java)
        assertNotNull(get)
        assertEquals("api/v1/collections/{id}", get!!.value)
    }

    // ------------------------------------------------------------------
    // Recommendations endpoints
    // ------------------------------------------------------------------

    @Test
    fun `getSimilarMedia endpoint is GET with path mediaId`() {
        val method = findMethod("getSimilarMedia")
        val get = method.getAnnotation(GET::class.java)
        assertNotNull(get)
        assertEquals("api/v1/recommendations/similar/{mediaId}", get!!.value)
    }

    @Test
    fun `getTrendingMedia endpoint is GET`() {
        val method = findMethod("getTrendingMedia")
        val get = method.getAnnotation(GET::class.java)
        assertNotNull(get)
        assertEquals("api/v1/recommendations/trending", get!!.value)
    }

    // ------------------------------------------------------------------
    // Scans, playlists, stats, subtitles endpoints
    // ------------------------------------------------------------------

    @Test
    fun `getScans endpoint is GET`() {
        val method = findMethod("getScans")
        val get = method.getAnnotation(GET::class.java)
        assertNotNull(get)
        assertEquals("api/v1/scans", get!!.value)
    }

    @Test
    fun `getPlaylists endpoint is GET`() {
        val method = findMethod("getPlaylists")
        val get = method.getAnnotation(GET::class.java)
        assertNotNull(get)
        assertEquals("api/v1/playlists", get!!.value)
    }

    @Test
    fun `createPlaylist endpoint is POST`() {
        val method = findMethod("createPlaylist")
        val post = method.getAnnotation(POST::class.java)
        assertNotNull(post)
        assertEquals("api/v1/playlists", post!!.value)
    }

    @Test
    fun `addPlaylistItem endpoint is POST with path id`() {
        val method = findMethod("addPlaylistItem")
        val post = method.getAnnotation(POST::class.java)
        assertNotNull(post)
        assertEquals("api/v1/playlists/{id}/items", post!!.value)
    }

    @Test
    fun `getOverallStats endpoint is GET`() {
        val method = findMethod("getOverallStats")
        val get = method.getAnnotation(GET::class.java)
        assertNotNull(get)
        assertEquals("api/v1/stats/overall", get!!.value)
    }

    @Test
    fun `getSubtitles endpoint is GET with path mediaId`() {
        val method = findMethod("getSubtitles")
        val get = method.getAnnotation(GET::class.java)
        assertNotNull(get)
        assertEquals("api/v1/subtitles/media/{mediaId}", get!!.value)
    }

    @Test
    fun `getSubtitleLanguages endpoint is GET`() {
        val method = findMethod("getSubtitleLanguages")
        val get = method.getAnnotation(GET::class.java)
        assertNotNull(get)
        assertEquals("api/v1/subtitles/languages", get!!.value)
    }

    // ------------------------------------------------------------------
    // Completeness: ensure the interface declares the expected count
    // ------------------------------------------------------------------

    @Test
    fun `CatalogizerApi is an interface`() {
        assertTrue("CatalogizerApi must be an interface", apiClass.isInterface)
    }

    @Test
    fun `all expected endpoints are declared`() {
        val expectedMethods = listOf(
            "login", "refreshToken", "logout",
            "getCatalog", "searchMedia", "searchEntities",
            "browseEntities", "getEntityById", "getEntityChildren", "getEntityStream",
            "startPlaybackSession", "progressPlaybackSession", "endPlaybackSession",
            "getEntityProgress", "getEntityHistory",
            "getMediaById", "getMediaInfo", "recognizeMedia", "updateWatchProgress",
            "addFavorite", "removeFavorite", "checkFavorite", "listFavorites",
            "getEntityStats",
            "getCollections", "getCollection",
            "getSimilarMedia", "getTrendingMedia",
            "getScans",
            "getPlaylists", "createPlaylist", "addPlaylistItem",
            "getOverallStats",
            "getSubtitles", "getSubtitleLanguages"
        )
        val actualMethodNames = apiClass.declaredMethods.map { it.name }.toSet()
        for (name in expectedMethods) {
            assertTrue(
                "CatalogizerApi should declare method '$name'",
                actualMethodNames.contains(name)
            )
        }
    }

    // ------------------------------------------------------------------
    // DTO model tests
    // ------------------------------------------------------------------

    @Test
    fun `EntityStatsResponse default values are zero and empty`() {
        val stats = EntityStatsResponse()
        assertEquals(0, stats.totalEntities)
        assertTrue(stats.byType.isEmpty())
    }

    @Test
    fun `EntityStatsResponse holds supplied values`() {
        val stats = EntityStatsResponse(totalEntities = 150, byType = mapOf("movie" to 100, "tv_show" to 50))
        assertEquals(150, stats.totalEntities)
        assertEquals(100, stats.byType["movie"])
        assertEquals(50, stats.byType["tv_show"])
    }

    @Test
    fun `LoginUser holds required fields`() {
        val user = LoginUser(id = 1L, username = "admin")
        assertEquals(1L, user.id)
        assertEquals("admin", user.username)
    }

    @Test
    fun `LoginUser optional fields default to null`() {
        val user = LoginUser(id = 1L, username = "admin")
        assertEquals(null, user.email)
        assertEquals(null, user.display_name)
    }

    @Test
    fun `LoginResponse convenience properties delegate correctly`() {
        val user = LoginUser(id = 42L, username = "testuser", email = "test@test.com")
        val response = LoginResponse(
            user = user,
            session_token = "tok_abc123",
            refresh_token = "ref_xyz789",
            expires_at = "2026-12-31T23:59:59Z"
        )
        assertEquals("tok_abc123", response.token)
        assertEquals(42L, response.userId)
        assertEquals("testuser", response.username)
        assertEquals("2026-12-31T23:59:59Z", response.expiresAt)
    }

    @Test
    fun `LoginResponse optional fields default to null`() {
        val user = LoginUser(id = 1L, username = "u")
        val response = LoginResponse(user = user, session_token = "tok")
        assertEquals(null, response.refresh_token)
        assertEquals(null, response.expires_at)
        assertEquals(null, response.expiresAt)
    }
}
