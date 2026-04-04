# Android TV Home Screen Channels — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Integrate Android TV Home Screen Channels into catalogizer-androidtv so content from the Catalogizer library appears on the TV home screen with deep-link navigation.

**Architecture:** Three integration surfaces — default curated channel ("Catalogizer Picks"), dynamic per-category channels (one per media type with content), and the system Watch Next row. A `TvChannelRepository` manages all channel/program CRUD via `TvContractCompat`. A `WorkManager` periodic job plus app-launch and sync triggers keep channels fresh. Deep links route through a transparent `ChannelDeepLinkActivity` that respects per-category launch settings (detail screen vs. immediate play). Full cleanup on logout.

**Tech Stack:** Kotlin, Jetpack Compose for TV, `androidx.tvprovider:tvprovider:1.0.0`, `androidx.work:work-runtime-ktx:2.9.0`, `TvContractCompat`, DataStore Preferences, Robolectric + MockK for tests.

**Spec:** `docs/superpowers/specs/2026-04-04-android-tv-home-channels-design.md`

**Base path:** `catalogizer-androidtv/app/src/main/java/com/catalogizer/androidtv`
**Test base path:** `catalogizer-androidtv/app/src/test/java/com/catalogizer/androidtv`

---

### Task 1: Add WorkManager Dependency

**Files:**
- Modify: `catalogizer-androidtv/app/build.gradle.kts:121` (dependencies block)

- [ ] **Step 1: Add work-runtime-ktx dependency**

In `catalogizer-androidtv/app/build.gradle.kts`, inside the `dependencies` block, after the `tvprovider` line (line 188), add:

```kotlin
    // WorkManager for periodic channel sync
    implementation("androidx.work:work-runtime-ktx:2.9.0")
    // WorkManager testing
    testImplementation("androidx.work:work-testing:2.9.0")
```

- [ ] **Step 2: Verify build compiles**

Run: `cd catalogizer-androidtv && ./gradlew assembleDebug --dry-run 2>&1 | tail -5`
Expected: Task list printed without errors (dry-run verifies dependency resolution).

- [ ] **Step 3: Commit**

```bash
cd catalogizer-androidtv && git add app/build.gradle.kts
git commit -m "build: add WorkManager dependency for TV channel sync"
```

---

### Task 2: Add LaunchAction Enum and Channel Settings to Data Models

**Files:**
- Modify: `catalogizer-androidtv/app/src/main/java/com/catalogizer/androidtv/data/models/Settings.kt`
- Test: `catalogizer-androidtv/app/src/test/java/com/catalogizer/androidtv/data/tv/LaunchActionTest.kt`

- [ ] **Step 1: Write the failing test**

Create `catalogizer-androidtv/app/src/test/java/com/catalogizer/androidtv/data/tv/LaunchActionTest.kt`:

```kotlin
package com.catalogizer.androidtv.data.tv

import org.junit.Assert.*
import org.junit.Test

class LaunchActionTest {

    @Test
    fun `LaunchAction has DETAIL and IMMEDIATE_PLAY values`() {
        assertEquals(2, LaunchAction.values().size)
        assertNotNull(LaunchAction.DETAIL)
        assertNotNull(LaunchAction.IMMEDIATE_PLAY)
    }

    @Test
    fun `LaunchAction fromString returns correct value`() {
        assertEquals(LaunchAction.DETAIL, LaunchAction.fromString("DETAIL"))
        assertEquals(LaunchAction.IMMEDIATE_PLAY, LaunchAction.fromString("IMMEDIATE_PLAY"))
    }

    @Test
    fun `LaunchAction fromString returns DETAIL for unknown input`() {
        assertEquals(LaunchAction.DETAIL, LaunchAction.fromString("UNKNOWN"))
        assertEquals(LaunchAction.DETAIL, LaunchAction.fromString(""))
    }
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd catalogizer-androidtv && ./gradlew testDebugUnitTest --tests "com.catalogizer.androidtv.data.tv.LaunchActionTest" 2>&1 | tail -10`
Expected: Compilation error — `LaunchAction` does not exist yet.

- [ ] **Step 3: Create LaunchAction enum**

Create `catalogizer-androidtv/app/src/main/java/com/catalogizer/androidtv/data/tv/LaunchAction.kt`:

```kotlin
package com.catalogizer.androidtv.data.tv

/**
 * Controls what happens when a user clicks a program card on the Android TV home screen.
 * Configurable per media type in Settings.
 */
enum class LaunchAction {
    /** Open the MediaDetailScreen where the user can choose to play. */
    DETAIL,
    /** Start playback immediately. */
    IMMEDIATE_PLAY;

    companion object {
        fun fromString(value: String): LaunchAction {
            return values().find { it.name == value } ?: DETAIL
        }
    }
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd catalogizer-androidtv && ./gradlew testDebugUnitTest --tests "com.catalogizer.androidtv.data.tv.LaunchActionTest" 2>&1 | tail -10`
Expected: 3 tests PASSED.

- [ ] **Step 5: Commit**

```bash
cd catalogizer-androidtv && git add app/src/main/java/com/catalogizer/androidtv/data/tv/LaunchAction.kt app/src/test/java/com/catalogizer/androidtv/data/tv/LaunchActionTest.kt
git commit -m "feat: add LaunchAction enum for per-category channel tap behavior"
```

---

### Task 3: ChannelProgramMapper — MediaItem to PreviewProgram Conversion

**Files:**
- Create: `catalogizer-androidtv/app/src/main/java/com/catalogizer/androidtv/data/tv/ChannelProgramMapper.kt`
- Test: `catalogizer-androidtv/app/src/test/java/com/catalogizer/androidtv/data/tv/ChannelProgramMapperTest.kt`

- [ ] **Step 1: Write the failing tests**

Create `catalogizer-androidtv/app/src/test/java/com/catalogizer/androidtv/data/tv/ChannelProgramMapperTest.kt`:

```kotlin
package com.catalogizer.androidtv.data.tv

import android.net.Uri
import androidx.tvprovider.media.tv.TvContractCompat
import com.catalogizer.androidtv.data.models.ExternalMetadata
import com.catalogizer.androidtv.data.models.MediaItem
import org.junit.Assert.*
import org.junit.Test
import org.junit.runner.RunWith
import org.robolectric.RobolectricTestRunner
import org.robolectric.annotation.Config

@RunWith(RobolectricTestRunner::class)
@Config(manifest = Config.NONE)
class ChannelProgramMapperTest {

    @Test
    fun `mapToPreviewProgramType maps movie correctly`() {
        assertEquals(
            TvContractCompat.PreviewPrograms.TYPE_MOVIE,
            ChannelProgramMapper.mapToPreviewProgramType("movie")
        )
    }

    @Test
    fun `mapToPreviewProgramType maps tv_show correctly`() {
        assertEquals(
            TvContractCompat.PreviewPrograms.TYPE_TV_SERIES,
            ChannelProgramMapper.mapToPreviewProgramType("tv_show")
        )
    }

    @Test
    fun `mapToPreviewProgramType maps game correctly`() {
        assertEquals(
            TvContractCompat.PreviewPrograms.TYPE_GAME,
            ChannelProgramMapper.mapToPreviewProgramType("game")
        )
    }

    @Test
    fun `mapToPreviewProgramType maps music to TYPE_TRACK`() {
        assertEquals(
            TvContractCompat.PreviewPrograms.TYPE_TRACK,
            ChannelProgramMapper.mapToPreviewProgramType("music")
        )
    }

    @Test
    fun `mapToPreviewProgramType maps unknown type to TYPE_CLIP`() {
        assertEquals(
            TvContractCompat.PreviewPrograms.TYPE_CLIP,
            ChannelProgramMapper.mapToPreviewProgramType("unknown_type")
        )
    }

    @Test
    fun `mapToPreviewProgramType maps all 16 types`() {
        val expectedMappings = mapOf(
            "movie" to TvContractCompat.PreviewPrograms.TYPE_MOVIE,
            "tv_show" to TvContractCompat.PreviewPrograms.TYPE_TV_SERIES,
            "tv_episode" to TvContractCompat.PreviewPrograms.TYPE_TV_EPISODE,
            "music" to TvContractCompat.PreviewPrograms.TYPE_TRACK,
            "anime" to TvContractCompat.PreviewPrograms.TYPE_MOVIE,
            "documentary" to TvContractCompat.PreviewPrograms.TYPE_MOVIE,
            "concert" to TvContractCompat.PreviewPrograms.TYPE_EVENT,
            "game" to TvContractCompat.PreviewPrograms.TYPE_GAME,
            "software" to TvContractCompat.PreviewPrograms.TYPE_APP,
            "ebook" to TvContractCompat.PreviewPrograms.TYPE_CLIP,
            "audiobook" to TvContractCompat.PreviewPrograms.TYPE_ALBUM,
            "podcast" to TvContractCompat.PreviewPrograms.TYPE_CHANNEL,
            "training" to TvContractCompat.PreviewPrograms.TYPE_CLIP,
            "sports" to TvContractCompat.PreviewPrograms.TYPE_EVENT,
            "news" to TvContractCompat.PreviewPrograms.TYPE_CLIP,
            "other" to TvContractCompat.PreviewPrograms.TYPE_CLIP
        )
        expectedMappings.forEach { (mediaType, expectedType) ->
            assertEquals(
                "Failed for mediaType=$mediaType",
                expectedType,
                ChannelProgramMapper.mapToPreviewProgramType(mediaType)
            )
        }
    }

    @Test
    fun `buildDeepLinkUri creates correct URI`() {
        val uri = ChannelProgramMapper.buildDeepLinkUri(42L, "movie")
        assertEquals("catalogizer", uri.scheme)
        assertEquals("media", uri.host)
        assertEquals("/42", uri.path)
        assertEquals("movie", uri.getQueryParameter("type"))
    }

    @Test
    fun `buildDeepLinkUri with null type omits type param`() {
        val uri = ChannelProgramMapper.buildDeepLinkUri(99L, null)
        assertEquals("catalogizer", uri.scheme)
        assertEquals("/99", uri.path)
        assertNull(uri.getQueryParameter("type"))
    }

    @Test
    fun `toPreviewProgramValues maps required fields`() {
        val item = MediaItem(
            id = 1L,
            title = "Test Movie",
            mediaType = "movie",
            description = "A test movie",
            year = 2025,
            duration = 7200L, // 2 hours in seconds
            rating = 8.5,
            coverUrl = "https://example.com/poster.jpg"
        )
        val values = ChannelProgramMapper.toPreviewProgramValues(item, channelId = 100L)
        assertEquals("Test Movie", values.getAsString(TvContractCompat.PreviewPrograms.COLUMN_TITLE))
        assertEquals("A test movie", values.getAsString(TvContractCompat.PreviewPrograms.COLUMN_SHORT_DESCRIPTION))
        assertEquals(100L, values.getAsLong(TvContractCompat.PreviewPrograms.COLUMN_CHANNEL_ID))
        assertEquals(
            TvContractCompat.PreviewPrograms.TYPE_MOVIE,
            values.getAsInteger(TvContractCompat.PreviewPrograms.COLUMN_TYPE)
        )
    }

    @Test
    fun `toPreviewProgramValues handles missing optional fields`() {
        val item = MediaItem(
            id = 2L,
            title = "Minimal Item"
        )
        val values = ChannelProgramMapper.toPreviewProgramValues(item, channelId = 100L)
        assertEquals("Minimal Item", values.getAsString(TvContractCompat.PreviewPrograms.COLUMN_TITLE))
        assertEquals(100L, values.getAsLong(TvContractCompat.PreviewPrograms.COLUMN_CHANNEL_ID))
    }

    @Test
    fun `toPreviewProgramValues includes poster from externalMetadata`() {
        val item = MediaItem(
            id = 3L,
            title = "Movie With Metadata",
            mediaType = "movie",
            externalMetadata = listOf(
                ExternalMetadata(
                    id = 1L,
                    mediaId = 3L,
                    provider = "tmdb",
                    externalId = "123",
                    title = "Movie With Metadata",
                    posterUrl = "https://tmdb.org/poster.jpg",
                    backdropUrl = "https://tmdb.org/backdrop.jpg"
                )
            )
        )
        val values = ChannelProgramMapper.toPreviewProgramValues(item, channelId = 100L)
        val posterUri = values.getAsString(TvContractCompat.PreviewPrograms.COLUMN_POSTER_ART_URI)
        assertNotNull(posterUri)
        assertTrue(posterUri!!.contains("poster.jpg"))
    }

    @Test
    fun `toWatchNextValues sets WATCH_NEXT_TYPE_CONTINUE for in-progress`() {
        val item = MediaItem(
            id = 10L,
            title = "Watching Movie",
            mediaType = "movie",
            duration = 7200L,
            watchProgress = 0.5
        )
        val values = ChannelProgramMapper.toWatchNextValues(
            item, watchNextType = TvContractCompat.WatchNextPrograms.WATCH_NEXT_TYPE_CONTINUE
        )
        assertEquals("Watching Movie", values.getAsString(TvContractCompat.WatchNextPrograms.COLUMN_TITLE))
        assertEquals(
            TvContractCompat.WatchNextPrograms.WATCH_NEXT_TYPE_CONTINUE,
            values.getAsInteger(TvContractCompat.WatchNextPrograms.COLUMN_WATCH_NEXT_TYPE)
        )
    }

    @Test
    fun `toWatchNextValues sets WATCH_NEXT_TYPE_NEXT for next episode`() {
        val item = MediaItem(
            id = 11L,
            title = "Next Episode",
            mediaType = "tv_episode",
            duration = 3600L
        )
        val values = ChannelProgramMapper.toWatchNextValues(
            item, watchNextType = TvContractCompat.WatchNextPrograms.WATCH_NEXT_TYPE_NEXT
        )
        assertEquals(
            TvContractCompat.WatchNextPrograms.WATCH_NEXT_TYPE_NEXT,
            values.getAsInteger(TvContractCompat.WatchNextPrograms.COLUMN_WATCH_NEXT_TYPE)
        )
    }
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd catalogizer-androidtv && ./gradlew testDebugUnitTest --tests "com.catalogizer.androidtv.data.tv.ChannelProgramMapperTest" 2>&1 | tail -10`
Expected: Compilation error — `ChannelProgramMapper` does not exist.

- [ ] **Step 3: Implement ChannelProgramMapper**

Create `catalogizer-androidtv/app/src/main/java/com/catalogizer/androidtv/data/tv/ChannelProgramMapper.kt`:

```kotlin
package com.catalogizer.androidtv.data.tv

import android.content.ContentValues
import android.net.Uri
import androidx.tvprovider.media.tv.TvContractCompat
import com.catalogizer.androidtv.data.models.MediaItem

/**
 * Maps [MediaItem] instances to [ContentValues] suitable for inserting into the
 * system TvProvider as [TvContractCompat.PreviewPrograms] or [TvContractCompat.WatchNextPrograms].
 */
object ChannelProgramMapper {

    private val TYPE_MAP = mapOf(
        "movie" to TvContractCompat.PreviewPrograms.TYPE_MOVIE,
        "tv_show" to TvContractCompat.PreviewPrograms.TYPE_TV_SERIES,
        "tv_episode" to TvContractCompat.PreviewPrograms.TYPE_TV_EPISODE,
        "music" to TvContractCompat.PreviewPrograms.TYPE_TRACK,
        "anime" to TvContractCompat.PreviewPrograms.TYPE_MOVIE,
        "documentary" to TvContractCompat.PreviewPrograms.TYPE_MOVIE,
        "concert" to TvContractCompat.PreviewPrograms.TYPE_EVENT,
        "game" to TvContractCompat.PreviewPrograms.TYPE_GAME,
        "software" to TvContractCompat.PreviewPrograms.TYPE_APP,
        "ebook" to TvContractCompat.PreviewPrograms.TYPE_CLIP,
        "audiobook" to TvContractCompat.PreviewPrograms.TYPE_ALBUM,
        "podcast" to TvContractCompat.PreviewPrograms.TYPE_CHANNEL,
        "training" to TvContractCompat.PreviewPrograms.TYPE_CLIP,
        "sports" to TvContractCompat.PreviewPrograms.TYPE_EVENT,
        "news" to TvContractCompat.PreviewPrograms.TYPE_CLIP,
        "other" to TvContractCompat.PreviewPrograms.TYPE_CLIP
    )

    fun mapToPreviewProgramType(mediaType: String?): Int {
        return TYPE_MAP[mediaType] ?: TvContractCompat.PreviewPrograms.TYPE_CLIP
    }

    fun buildDeepLinkUri(mediaId: Long, mediaType: String?): Uri {
        val builder = Uri.Builder()
            .scheme("catalogizer")
            .authority("media")
            .appendPath(mediaId.toString())
        mediaType?.let { builder.appendQueryParameter("type", it) }
        return builder.build()
    }

    fun toPreviewProgramValues(item: MediaItem, channelId: Long): ContentValues {
        val values = ContentValues()
        values.put(TvContractCompat.PreviewPrograms.COLUMN_CHANNEL_ID, channelId)
        values.put(TvContractCompat.PreviewPrograms.COLUMN_TITLE, item.title)
        values.put(TvContractCompat.PreviewPrograms.COLUMN_TYPE, mapToPreviewProgramType(item.mediaType))
        values.put(
            TvContractCompat.PreviewPrograms.COLUMN_INTENT_URI,
            buildDeepLinkUri(item.id, item.mediaType).toString()
        )

        item.description?.let {
            values.put(TvContractCompat.PreviewPrograms.COLUMN_SHORT_DESCRIPTION, it)
        }
        item.posterUrl?.let {
            values.put(TvContractCompat.PreviewPrograms.COLUMN_POSTER_ART_URI, it)
        }
        item.backdropUrl?.let {
            values.put(TvContractCompat.PreviewPrograms.COLUMN_THUMBNAIL_URI, it)
        }
        item.duration?.let { durationSec ->
            values.put(TvContractCompat.PreviewPrograms.COLUMN_DURATION_MILLIS, durationSec * 1000)
        }
        item.year?.let {
            values.put(TvContractCompat.PreviewPrograms.COLUMN_RELEASE_DATE, it.toString())
        }
        if (item.genres.isNotEmpty()) {
            values.put(TvContractCompat.PreviewPrograms.COLUMN_GENRE, item.genres.joinToString(", "))
        }
        if (item.watchProgress > 0.0 && item.duration != null) {
            val positionMs = (item.watchProgress * item.duration!! * 1000).toLong()
            values.put(TvContractCompat.PreviewPrograms.COLUMN_LAST_PLAYBACK_POSITION_MILLIS, positionMs)
            values.put(TvContractCompat.PreviewPrograms.COLUMN_DURATION_MILLIS, item.duration!! * 1000)
        }

        return values
    }

    fun toWatchNextValues(item: MediaItem, watchNextType: Int): ContentValues {
        val values = ContentValues()
        values.put(TvContractCompat.WatchNextPrograms.COLUMN_TITLE, item.title)
        values.put(TvContractCompat.WatchNextPrograms.COLUMN_TYPE, mapToPreviewProgramType(item.mediaType))
        values.put(TvContractCompat.WatchNextPrograms.COLUMN_WATCH_NEXT_TYPE, watchNextType)
        values.put(
            TvContractCompat.WatchNextPrograms.COLUMN_INTENT_URI,
            buildDeepLinkUri(item.id, item.mediaType).toString()
        )
        values.put(
            TvContractCompat.WatchNextPrograms.COLUMN_LAST_ENGAGEMENT_TIME_UTC_MILLIS,
            System.currentTimeMillis()
        )

        item.description?.let {
            values.put(TvContractCompat.WatchNextPrograms.COLUMN_SHORT_DESCRIPTION, it)
        }
        item.posterUrl?.let {
            values.put(TvContractCompat.WatchNextPrograms.COLUMN_POSTER_ART_URI, it)
        }
        item.backdropUrl?.let {
            values.put(TvContractCompat.WatchNextPrograms.COLUMN_THUMBNAIL_URI, it)
        }
        item.duration?.let { durationSec ->
            values.put(TvContractCompat.WatchNextPrograms.COLUMN_DURATION_MILLIS, durationSec * 1000)
            if (item.watchProgress > 0.0) {
                val positionMs = (item.watchProgress * durationSec * 1000).toLong()
                values.put(TvContractCompat.WatchNextPrograms.COLUMN_LAST_PLAYBACK_POSITION_MILLIS, positionMs)
            }
        }

        return values
    }
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd catalogizer-androidtv && ./gradlew testDebugUnitTest --tests "com.catalogizer.androidtv.data.tv.ChannelProgramMapperTest" 2>&1 | tail -15`
Expected: All 12 tests PASSED.

- [ ] **Step 5: Commit**

```bash
cd catalogizer-androidtv && git add app/src/main/java/com/catalogizer/androidtv/data/tv/ChannelProgramMapper.kt app/src/test/java/com/catalogizer/androidtv/data/tv/ChannelProgramMapperTest.kt
git commit -m "feat: add ChannelProgramMapper for MediaItem to PreviewProgram conversion"
```

---

### Task 4: Channel ID Persistence in SettingsRepository

**Files:**
- Modify: `catalogizer-androidtv/app/src/main/java/com/catalogizer/androidtv/data/repository/SettingsRepository.kt`
- Test: `catalogizer-androidtv/app/src/test/java/com/catalogizer/androidtv/data/tv/ChannelSettingsTest.kt`

- [ ] **Step 1: Write the failing tests**

Create `catalogizer-androidtv/app/src/test/java/com/catalogizer/androidtv/data/tv/ChannelSettingsTest.kt`:

```kotlin
package com.catalogizer.androidtv.data.tv

import androidx.datastore.core.DataStore
import androidx.datastore.preferences.core.PreferenceDataStoreFactory
import androidx.datastore.preferences.core.Preferences
import com.catalogizer.androidtv.data.repository.SettingsRepository
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.test.runTest
import org.junit.Assert.*
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.TemporaryFolder
import org.junit.runner.RunWith
import org.robolectric.RobolectricTestRunner
import org.robolectric.annotation.Config

@OptIn(ExperimentalCoroutinesApi::class)
@RunWith(RobolectricTestRunner::class)
@Config(manifest = Config.NONE)
class ChannelSettingsTest {

    @get:Rule
    val tmpFolder = TemporaryFolder()

    private lateinit var dataStore: DataStore<Preferences>
    private lateinit var repo: SettingsRepository

    @Before
    fun setup() {
        dataStore = PreferenceDataStoreFactory.create {
            tmpFolder.newFile("test_prefs.preferences_pb")
        }
        repo = SettingsRepository(dataStore)
    }

    @Test
    fun `saveChannelId persists and retrieves channel ID`() = runTest {
        repo.saveChannelId("default", 42L)
        assertEquals(42L, repo.getChannelId("default"))
    }

    @Test
    fun `getChannelId returns null for unknown key`() = runTest {
        assertNull(repo.getChannelId("nonexistent"))
    }

    @Test
    fun `saveChannelId overwrites existing value`() = runTest {
        repo.saveChannelId("movie", 10L)
        repo.saveChannelId("movie", 20L)
        assertEquals(20L, repo.getChannelId("movie"))
    }

    @Test
    fun `removeChannelId clears stored ID`() = runTest {
        repo.saveChannelId("movie", 10L)
        repo.removeChannelId("movie")
        assertNull(repo.getChannelId("movie"))
    }

    @Test
    fun `clearAllChannelIds removes all channel IDs`() = runTest {
        repo.saveChannelId("default", 1L)
        repo.saveChannelId("movie", 2L)
        repo.saveChannelId("tv_show", 3L)
        repo.clearAllChannelIds()
        assertNull(repo.getChannelId("default"))
        assertNull(repo.getChannelId("movie"))
        assertNull(repo.getChannelId("tv_show"))
    }

    @Test
    fun `saveLaunchAction persists and retrieves action`() = runTest {
        repo.saveLaunchAction("movie", LaunchAction.IMMEDIATE_PLAY)
        assertEquals(LaunchAction.IMMEDIATE_PLAY, repo.getLaunchAction("movie"))
    }

    @Test
    fun `getLaunchAction returns DETAIL by default`() = runTest {
        assertEquals(LaunchAction.DETAIL, repo.getLaunchAction("movie"))
    }

    @Test
    fun `clearAllLaunchActions resets all to default`() = runTest {
        repo.saveLaunchAction("movie", LaunchAction.IMMEDIATE_PLAY)
        repo.saveLaunchAction("music", LaunchAction.IMMEDIATE_PLAY)
        repo.clearAllLaunchActions()
        assertEquals(LaunchAction.DETAIL, repo.getLaunchAction("movie"))
        assertEquals(LaunchAction.DETAIL, repo.getLaunchAction("music"))
    }

    @Test
    fun `getAllLaunchActions returns all saved actions`() = runTest {
        repo.saveLaunchAction("movie", LaunchAction.IMMEDIATE_PLAY)
        repo.saveLaunchAction("tv_show", LaunchAction.DETAIL)
        val actions = repo.getAllLaunchActions()
        assertEquals(LaunchAction.IMMEDIATE_PLAY, actions["movie"])
        assertEquals(LaunchAction.DETAIL, actions["tv_show"])
    }
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd catalogizer-androidtv && ./gradlew testDebugUnitTest --tests "com.catalogizer.androidtv.data.tv.ChannelSettingsTest" 2>&1 | tail -10`
Expected: Compilation error — `saveChannelId`, `getChannelId`, etc. do not exist on `SettingsRepository`.

- [ ] **Step 3: Add channel settings methods to SettingsRepository**

In `catalogizer-androidtv/app/src/main/java/com/catalogizer/androidtv/data/repository/SettingsRepository.kt`, add these imports at the top:

```kotlin
import androidx.datastore.preferences.core.longPreferencesKey
import com.catalogizer.androidtv.data.tv.LaunchAction
```

Then add these methods at the bottom of the class, before the closing `}` and after the `deserializeServers` method:

```kotlin
    // ─── TV Channel ID Persistence ─────────────────────────────────────

    private fun channelIdKey(channelKey: String) = longPreferencesKey("channel_id_$channelKey")
    private fun launchActionKey(mediaType: String) = stringPreferencesKey("launch_action_$mediaType")

    suspend fun saveChannelId(channelKey: String, channelId: Long) {
        dataStore.edit { preferences ->
            preferences[channelIdKey(channelKey)] = channelId
        }
    }

    suspend fun getChannelId(channelKey: String): Long? {
        val prefs = dataStore.data.first()
        val value = prefs[channelIdKey(channelKey)]
        return if (value != null && value != 0L) value else null
    }

    suspend fun removeChannelId(channelKey: String) {
        dataStore.edit { preferences ->
            preferences.remove(channelIdKey(channelKey))
        }
    }

    suspend fun clearAllChannelIds() {
        dataStore.edit { preferences ->
            val channelKeys = preferences.asMap().keys.filter {
                it.name.startsWith("channel_id_")
            }
            channelKeys.forEach { preferences.remove(it) }
        }
    }

    // ─── Per-Category Launch Action ────────────────────────────────────

    suspend fun saveLaunchAction(mediaType: String, action: LaunchAction) {
        dataStore.edit { preferences ->
            preferences[launchActionKey(mediaType)] = action.name
        }
    }

    suspend fun getLaunchAction(mediaType: String): LaunchAction {
        val prefs = dataStore.data.first()
        val value = prefs[launchActionKey(mediaType)]
        return if (value != null) LaunchAction.fromString(value) else LaunchAction.DETAIL
    }

    suspend fun getAllLaunchActions(): Map<String, LaunchAction> {
        val prefs = dataStore.data.first()
        val result = mutableMapOf<String, LaunchAction>()
        prefs.asMap().forEach { (key, value) ->
            if (key.name.startsWith("launch_action_") && value is String) {
                val mediaType = key.name.removePrefix("launch_action_")
                result[mediaType] = LaunchAction.fromString(value)
            }
        }
        return result
    }

    suspend fun clearAllLaunchActions() {
        dataStore.edit { preferences ->
            val actionKeys = preferences.asMap().keys.filter {
                it.name.startsWith("launch_action_")
            }
            actionKeys.forEach { preferences.remove(it) }
        }
    }
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd catalogizer-androidtv && ./gradlew testDebugUnitTest --tests "com.catalogizer.androidtv.data.tv.ChannelSettingsTest" 2>&1 | tail -15`
Expected: All 9 tests PASSED.

- [ ] **Step 5: Commit**

```bash
cd catalogizer-androidtv && git add app/src/main/java/com/catalogizer/androidtv/data/repository/SettingsRepository.kt app/src/test/java/com/catalogizer/androidtv/data/tv/ChannelSettingsTest.kt
git commit -m "feat: add channel ID persistence and per-category launch action settings"
```

---

### Task 5: TvChannelRepository — Channel & Program CRUD

**Files:**
- Create: `catalogizer-androidtv/app/src/main/java/com/catalogizer/androidtv/data/tv/TvChannelRepository.kt`
- Test: `catalogizer-androidtv/app/src/test/java/com/catalogizer/androidtv/data/tv/TvChannelRepositoryTest.kt`

- [ ] **Step 1: Write the failing tests**

Create `catalogizer-androidtv/app/src/test/java/com/catalogizer/androidtv/data/tv/TvChannelRepositoryTest.kt`:

```kotlin
package com.catalogizer.androidtv.data.tv

import android.content.Context
import com.catalogizer.androidtv.data.models.MediaItem
import com.catalogizer.androidtv.data.repository.MediaRepository
import com.catalogizer.androidtv.data.repository.SettingsRepository
import io.mockk.*
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.flow.flowOf
import kotlinx.coroutines.test.runTest
import org.junit.Assert.*
import org.junit.Before
import org.junit.Test
import org.junit.runner.RunWith
import org.robolectric.RobolectricTestRunner
import org.robolectric.RuntimeEnvironment
import org.robolectric.annotation.Config

@OptIn(ExperimentalCoroutinesApi::class)
@RunWith(RobolectricTestRunner::class)
@Config(manifest = Config.NONE, sdk = [28])
class TvChannelRepositoryTest {

    private lateinit var context: Context
    private lateinit var mediaRepository: MediaRepository
    private lateinit var settingsRepository: SettingsRepository
    private lateinit var repo: TvChannelRepository

    @Before
    fun setup() {
        context = RuntimeEnvironment.getApplication()
        mediaRepository = mockk(relaxed = true)
        settingsRepository = mockk(relaxed = true)
        repo = TvChannelRepository(context, mediaRepository, settingsRepository)
    }

    @Test
    fun `buildDefaultChannelContent combines continue watching, recent, and trending`() = runTest {
        val continueWatching = listOf(
            MediaItem(id = 1L, title = "Continue 1", watchProgress = 0.5, mediaType = "movie")
        )
        val recent = listOf(
            MediaItem(id = 2L, title = "Recent 1", mediaType = "movie")
        )
        val trending = listOf(
            MediaItem(id = 3L, title = "Trending 1", mediaType = "tv_show")
        )
        coEvery { mediaRepository.searchMedia(any()) } returns flowOf(continueWatching)
        coEvery { mediaRepository.browseEntities(any(), any(), any(), any()) } returns flowOf(recent)
        coEvery { mediaRepository.getTrendingItems(any()) } returns trending

        val items = repo.buildDefaultChannelContent()
        assertTrue(items.isNotEmpty())
        assertTrue(items.size <= 30)
    }

    @Test
    fun `buildDefaultChannelContent caps at 30 items`() = runTest {
        val manyItems = (1L..50L).map {
            MediaItem(id = it, title = "Item $it", mediaType = "movie")
        }
        coEvery { mediaRepository.searchMedia(any()) } returns flowOf(emptyList())
        coEvery { mediaRepository.browseEntities(any(), any(), any(), any()) } returns flowOf(manyItems)
        coEvery { mediaRepository.getTrendingItems(any()) } returns emptyList()

        val items = repo.buildDefaultChannelContent()
        assertTrue(items.size <= 30)
    }

    @Test
    fun `buildDefaultChannelContent deduplicates items by ID`() = runTest {
        val item = MediaItem(id = 1L, title = "Duplicate", mediaType = "movie", watchProgress = 0.5)
        coEvery { mediaRepository.searchMedia(any()) } returns flowOf(listOf(item))
        coEvery { mediaRepository.browseEntities(any(), any(), any(), any()) } returns flowOf(listOf(item))
        coEvery { mediaRepository.getTrendingItems(any()) } returns listOf(item)

        val items = repo.buildDefaultChannelContent()
        assertEquals(1, items.count { it.id == 1L })
    }

    @Test
    fun `buildCategoryContent fetches items for given type`() = runTest {
        val movies = listOf(
            MediaItem(id = 1L, title = "Movie 1", mediaType = "movie"),
            MediaItem(id = 2L, title = "Movie 2", mediaType = "movie")
        )
        coEvery { mediaRepository.browseEntities("movie", 30, "created", "desc") } returns flowOf(movies)

        val items = repo.buildCategoryContent("movie")
        assertEquals(2, items.size)
        assertEquals("Movie 1", items[0].title)
    }

    @Test
    fun `buildCategoryContent caps at 30 items`() = runTest {
        val manyItems = (1L..50L).map {
            MediaItem(id = it, title = "Movie $it", mediaType = "movie")
        }
        coEvery { mediaRepository.browseEntities("movie", 30, "created", "desc") } returns flowOf(manyItems)

        val items = repo.buildCategoryContent("movie")
        assertTrue(items.size <= 30)
    }

    @Test
    fun `getActiveMediaTypes returns only types with content`() = runTest {
        coEvery { mediaRepository.getEntityStats() } returns Pair(100, mapOf(
            "movie" to 50,
            "tv_show" to 30,
            "music" to 0,
            "game" to 20
        ))

        val types = repo.getActiveMediaTypes()
        assertTrue(types.contains("movie"))
        assertTrue(types.contains("tv_show"))
        assertTrue(types.contains("game"))
        assertFalse(types.contains("music"))
    }

    @Test
    fun `getActiveMediaTypes returns empty list on API failure`() = runTest {
        coEvery { mediaRepository.getEntityStats() } returns Pair(0, emptyMap())

        val types = repo.getActiveMediaTypes()
        assertTrue(types.isEmpty())
    }
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd catalogizer-androidtv && ./gradlew testDebugUnitTest --tests "com.catalogizer.androidtv.data.tv.TvChannelRepositoryTest" 2>&1 | tail -10`
Expected: Compilation error — `TvChannelRepository` does not exist.

- [ ] **Step 3: Implement TvChannelRepository**

Create `catalogizer-androidtv/app/src/main/java/com/catalogizer/androidtv/data/tv/TvChannelRepository.kt`:

```kotlin
package com.catalogizer.androidtv.data.tv

import android.content.ContentValues
import android.content.Context
import android.net.Uri
import android.util.Log
import androidx.tvprovider.media.tv.TvContractCompat
import com.catalogizer.androidtv.data.models.MediaItem
import com.catalogizer.androidtv.data.models.MediaSearchRequest
import com.catalogizer.androidtv.data.models.MediaType
import com.catalogizer.androidtv.data.repository.MediaRepository
import com.catalogizer.androidtv.data.repository.SettingsRepository
import kotlinx.coroutines.flow.first

/**
 * Manages Android TV Home Screen channels and programs via [TvContractCompat].
 * Creates a default curated channel, dynamic per-category channels, and populates
 * them with [MediaItem] data from [MediaRepository].
 */
class TvChannelRepository(
    private val context: Context,
    private val mediaRepository: MediaRepository,
    private val settingsRepository: SettingsRepository
) {
    companion object {
        private const val TAG = "TvChannelRepo"
        const val MAX_PROGRAMS_PER_CHANNEL = 30
        const val DEFAULT_CHANNEL_KEY = "default"
        const val DEFAULT_CHANNEL_DISPLAY_NAME = "Catalogizer Picks"
    }

    // ─── Content Building (testable, no ContentResolver dependency) ─────

    suspend fun buildDefaultChannelContent(): List<MediaItem> {
        val seen = mutableSetOf<Long>()
        val result = mutableListOf<MediaItem>()

        // 1. Continue watching (partially watched, not completed)
        try {
            val continueWatching = mediaRepository.searchMedia(
                MediaSearchRequest(sortBy = "created", sortOrder = "desc", limit = 10)
            ).first().filter { it.watchProgress > 0.05 && it.watchProgress < 0.9 }
            for (item in continueWatching) {
                if (seen.add(item.id)) result.add(item)
            }
        } catch (e: Exception) {
            Log.w(TAG, "Failed to load continue watching: ${e.message}")
        }

        // 2. Recently added across all types
        try {
            val recent = mediaRepository.browseEntities(
                "all", limit = 20, sortBy = "created", sortOrder = "desc"
            ).first()
            for (item in recent) {
                if (seen.add(item.id)) result.add(item)
            }
        } catch (e: Exception) {
            Log.w(TAG, "Failed to load recent: ${e.message}")
        }

        // 3. Trending
        try {
            val trending = mediaRepository.getTrendingItems(10)
            for (item in trending) {
                if (seen.add(item.id)) result.add(item)
            }
        } catch (e: Exception) {
            Log.w(TAG, "Failed to load trending: ${e.message}")
        }

        return result.take(MAX_PROGRAMS_PER_CHANNEL)
    }

    suspend fun buildCategoryContent(mediaType: String): List<MediaItem> {
        return try {
            mediaRepository.browseEntities(
                mediaType, limit = MAX_PROGRAMS_PER_CHANNEL, sortBy = "created", sortOrder = "desc"
            ).first().take(MAX_PROGRAMS_PER_CHANNEL)
        } catch (e: Exception) {
            Log.w(TAG, "Failed to load category $mediaType: ${e.message}")
            emptyList()
        }
    }

    suspend fun getActiveMediaTypes(): List<String> {
        return try {
            val (_, byType) = mediaRepository.getEntityStats()
            byType.filter { it.value > 0 }.keys.toList()
        } catch (e: Exception) {
            Log.w(TAG, "Failed to get entity stats: ${e.message}")
            emptyList()
        }
    }

    // ─── Channel CRUD (interacts with system ContentResolver) ───────────

    suspend fun initializeDefaultChannel() {
        val existingId = settingsRepository.getChannelId(DEFAULT_CHANNEL_KEY)
        if (existingId != null) return // Already created

        val channelValues = ContentValues().apply {
            put(TvContractCompat.Channels.COLUMN_DISPLAY_NAME, DEFAULT_CHANNEL_DISPLAY_NAME)
            put(TvContractCompat.Channels.COLUMN_APP_LINK_INTENT_URI, "catalogizer://home")
            put(TvContractCompat.Channels.COLUMN_TYPE, TvContractCompat.Channels.TYPE_PREVIEW)
        }

        try {
            val channelUri = context.contentResolver.insert(
                TvContractCompat.Channels.CONTENT_URI, channelValues
            )
            val channelId = channelUri?.let { android.content.ContentUris.parseId(it) }
            if (channelId != null && channelId > 0) {
                settingsRepository.saveChannelId(DEFAULT_CHANNEL_KEY, channelId)
                // Request default channel status
                TvContractCompat.requestChannelBrowsable(context, channelId)
                Log.d(TAG, "Default channel created: $channelId")
            }
        } catch (e: Exception) {
            Log.e(TAG, "Failed to create default channel: ${e.message}")
        }
    }

    suspend fun createCategoryChannel(mediaType: String, displayName: String): Long? {
        val existingId = settingsRepository.getChannelId(mediaType)
        if (existingId != null) return existingId

        val channelValues = ContentValues().apply {
            put(TvContractCompat.Channels.COLUMN_DISPLAY_NAME, displayName)
            put(TvContractCompat.Channels.COLUMN_APP_LINK_INTENT_URI, "catalogizer://browse/$mediaType")
            put(TvContractCompat.Channels.COLUMN_TYPE, TvContractCompat.Channels.TYPE_PREVIEW)
            put(TvContractCompat.Channels.COLUMN_INTERNAL_PROVIDER_ID, mediaType)
        }

        return try {
            val channelUri = context.contentResolver.insert(
                TvContractCompat.Channels.CONTENT_URI, channelValues
            )
            val channelId = channelUri?.let { android.content.ContentUris.parseId(it) }
            if (channelId != null && channelId > 0) {
                settingsRepository.saveChannelId(mediaType, channelId)
                Log.d(TAG, "Category channel created: $mediaType -> $channelId")
            }
            channelId
        } catch (e: Exception) {
            Log.e(TAG, "Failed to create category channel $mediaType: ${e.message}")
            null
        }
    }

    suspend fun refreshChannelPrograms(channelId: Long, items: List<MediaItem>) {
        // Delete existing programs for this channel
        try {
            context.contentResolver.delete(
                TvContractCompat.PreviewPrograms.CONTENT_URI,
                "${TvContractCompat.PreviewPrograms.COLUMN_CHANNEL_ID} = ?",
                arrayOf(channelId.toString())
            )
        } catch (e: Exception) {
            Log.w(TAG, "Failed to delete old programs for channel $channelId: ${e.message}")
        }

        // Insert new programs
        for (item in items) {
            try {
                val values = ChannelProgramMapper.toPreviewProgramValues(item, channelId)
                context.contentResolver.insert(TvContractCompat.PreviewPrograms.CONTENT_URI, values)
            } catch (e: Exception) {
                Log.w(TAG, "Failed to insert program ${item.id}: ${e.message}")
            }
        }
    }

    suspend fun deleteChannel(channelKey: String) {
        val channelId = settingsRepository.getChannelId(channelKey) ?: return
        try {
            val channelUri = TvContractCompat.buildChannelUri(channelId)
            context.contentResolver.delete(channelUri, null, null)
            settingsRepository.removeChannelId(channelKey)
            Log.d(TAG, "Deleted channel: $channelKey ($channelId)")
        } catch (e: Exception) {
            Log.e(TAG, "Failed to delete channel $channelKey: ${e.message}")
        }
    }

    suspend fun deleteAllChannels() {
        // Delete default channel
        deleteChannel(DEFAULT_CHANNEL_KEY)

        // Delete all category channels
        val activeTypes = MediaType.values().map { it.value }
        for (type in activeTypes) {
            deleteChannel(type)
        }

        settingsRepository.clearAllChannelIds()
        Log.d(TAG, "All channels deleted")
    }

    // ─── Orchestration ──────────────────────────────────────────────────

    suspend fun refreshAllChannels() {
        // Refresh default channel
        refreshDefaultChannel()

        // Create/refresh category channels
        createCategoryChannels()

        // Remove stale category channels
        removeStaleCategoryChannels()

        Log.d(TAG, "All channels refreshed")
    }

    suspend fun refreshDefaultChannel() {
        val channelId = settingsRepository.getChannelId(DEFAULT_CHANNEL_KEY) ?: return
        val content = buildDefaultChannelContent()
        refreshChannelPrograms(channelId, content)
    }

    suspend fun createCategoryChannels() {
        val activeTypes = getActiveMediaTypes()
        for (type in activeTypes) {
            val displayName = MediaType.fromValue(type).displayName
            val channelId = createCategoryChannel(type, displayName) ?: continue
            val content = buildCategoryContent(type)
            refreshChannelPrograms(channelId, content)
        }
    }

    suspend fun removeStaleCategoryChannels() {
        val activeTypes = getActiveMediaTypes().toSet()
        val allTypes = MediaType.values().map { it.value }
        for (type in allTypes) {
            if (type !in activeTypes) {
                val channelId = settingsRepository.getChannelId(type)
                if (channelId != null) {
                    deleteChannel(type)
                }
            }
        }
    }
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd catalogizer-androidtv && ./gradlew testDebugUnitTest --tests "com.catalogizer.androidtv.data.tv.TvChannelRepositoryTest" 2>&1 | tail -15`
Expected: All 7 tests PASSED.

- [ ] **Step 5: Commit**

```bash
cd catalogizer-androidtv && git add app/src/main/java/com/catalogizer/androidtv/data/tv/TvChannelRepository.kt app/src/test/java/com/catalogizer/androidtv/data/tv/TvChannelRepositoryTest.kt
git commit -m "feat: add TvChannelRepository for channel and program CRUD"
```

---

### Task 6: WatchNextManager

**Files:**
- Create: `catalogizer-androidtv/app/src/main/java/com/catalogizer/androidtv/data/tv/WatchNextManager.kt`
- Test: `catalogizer-androidtv/app/src/test/java/com/catalogizer/androidtv/data/tv/WatchNextManagerTest.kt`

- [ ] **Step 1: Write the failing tests**

Create `catalogizer-androidtv/app/src/test/java/com/catalogizer/androidtv/data/tv/WatchNextManagerTest.kt`:

```kotlin
package com.catalogizer.androidtv.data.tv

import com.catalogizer.androidtv.data.models.MediaItem
import com.catalogizer.androidtv.data.repository.MediaRepository
import io.mockk.*
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.flow.flowOf
import kotlinx.coroutines.test.runTest
import org.junit.Assert.*
import org.junit.Before
import org.junit.Test

@OptIn(ExperimentalCoroutinesApi::class)
class WatchNextManagerTest {

    private lateinit var mediaRepository: MediaRepository

    @Before
    fun setup() {
        mediaRepository = mockk(relaxed = true)
    }

    @Test
    fun `filterContinueWatching includes items between 5% and 90%`() {
        val items = listOf(
            MediaItem(id = 1L, title = "Too Early", watchProgress = 0.02),
            MediaItem(id = 2L, title = "In Progress", watchProgress = 0.5),
            MediaItem(id = 3L, title = "Almost Done", watchProgress = 0.85),
            MediaItem(id = 4L, title = "Completed", watchProgress = 0.95)
        )
        val filtered = WatchNextManager.filterContinueWatching(items)
        assertEquals(2, filtered.size)
        assertEquals(2L, filtered[0].id)
        assertEquals(3L, filtered[1].id)
    }

    @Test
    fun `filterContinueWatching returns empty for no in-progress items`() {
        val items = listOf(
            MediaItem(id = 1L, title = "Not Started", watchProgress = 0.0),
            MediaItem(id = 2L, title = "Done", watchProgress = 1.0)
        )
        val filtered = WatchNextManager.filterContinueWatching(items)
        assertTrue(filtered.isEmpty())
    }

    @Test
    fun `filterContinueWatching boundary at exactly 5% is included`() {
        val item = MediaItem(id = 1L, title = "At 5%", watchProgress = 0.05)
        val filtered = WatchNextManager.filterContinueWatching(listOf(item))
        assertEquals(1, filtered.size)
    }

    @Test
    fun `filterContinueWatching boundary at exactly 90% is excluded`() {
        val item = MediaItem(id = 1L, title = "At 90%", watchProgress = 0.90)
        val filtered = WatchNextManager.filterContinueWatching(listOf(item))
        assertTrue(filtered.isEmpty())
    }

    @Test
    fun `isStale returns true for items older than 30 days`() {
        val thirtyOneDaysAgo = System.currentTimeMillis() - (31L * 24 * 60 * 60 * 1000)
        assertTrue(WatchNextManager.isStale(thirtyOneDaysAgo))
    }

    @Test
    fun `isStale returns false for recent items`() {
        val oneDayAgo = System.currentTimeMillis() - (1L * 24 * 60 * 60 * 1000)
        assertFalse(WatchNextManager.isStale(oneDayAgo))
    }

    @Test
    fun `isStale returns false for zero timestamp`() {
        assertFalse(WatchNextManager.isStale(0L))
    }

    @Test
    fun `isCompleted returns true for progress above 90%`() {
        assertTrue(WatchNextManager.isCompleted(0.91))
        assertTrue(WatchNextManager.isCompleted(1.0))
    }

    @Test
    fun `isCompleted returns false for progress at or below 90%`() {
        assertFalse(WatchNextManager.isCompleted(0.90))
        assertFalse(WatchNextManager.isCompleted(0.5))
        assertFalse(WatchNextManager.isCompleted(0.0))
    }

    @Test
    fun `isTvSeriesType returns true for tv_show and tv_episode`() {
        assertTrue(WatchNextManager.isTvSeriesType("tv_show"))
        assertTrue(WatchNextManager.isTvSeriesType("tv_episode"))
    }

    @Test
    fun `isTvSeriesType returns false for other types`() {
        assertFalse(WatchNextManager.isTvSeriesType("movie"))
        assertFalse(WatchNextManager.isTvSeriesType("music"))
        assertFalse(WatchNextManager.isTvSeriesType(null))
    }
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd catalogizer-androidtv && ./gradlew testDebugUnitTest --tests "com.catalogizer.androidtv.data.tv.WatchNextManagerTest" 2>&1 | tail -10`
Expected: Compilation error — `WatchNextManager` does not exist.

- [ ] **Step 3: Implement WatchNextManager**

Create `catalogizer-androidtv/app/src/main/java/com/catalogizer/androidtv/data/tv/WatchNextManager.kt`:

```kotlin
package com.catalogizer.androidtv.data.tv

import android.content.Context
import android.util.Log
import androidx.tvprovider.media.tv.TvContractCompat
import com.catalogizer.androidtv.data.models.MediaItem
import com.catalogizer.androidtv.data.models.MediaSearchRequest
import com.catalogizer.androidtv.data.repository.MediaRepository
import kotlinx.coroutines.flow.first

/**
 * Manages the system Watch Next row on the Android TV home screen.
 * Adds partially-watched items (5%-90%), auto-surfaces next episodes for TV series,
 * and removes completed or stale entries.
 */
class WatchNextManager(
    private val context: Context,
    private val mediaRepository: MediaRepository
) {
    companion object {
        private const val TAG = "WatchNextMgr"
        private const val PROGRESS_MIN = 0.05
        private const val PROGRESS_MAX = 0.90
        private const val STALE_THRESHOLD_MS = 30L * 24 * 60 * 60 * 1000 // 30 days

        fun filterContinueWatching(items: List<MediaItem>): List<MediaItem> {
            return items.filter { it.watchProgress >= PROGRESS_MIN && it.watchProgress < PROGRESS_MAX }
        }

        fun isStale(lastEngagementTimeMs: Long): Boolean {
            if (lastEngagementTimeMs == 0L) return false
            return (System.currentTimeMillis() - lastEngagementTimeMs) > STALE_THRESHOLD_MS
        }

        fun isCompleted(progress: Double): Boolean {
            return progress > PROGRESS_MAX
        }

        fun isTvSeriesType(mediaType: String?): Boolean {
            return mediaType == "tv_show" || mediaType == "tv_episode"
        }
    }

    suspend fun refreshWatchNext() {
        // 1. Get all items with watch progress
        val allItems = try {
            mediaRepository.searchMedia(
                MediaSearchRequest(sortBy = "updated_at", sortOrder = "desc", limit = 50)
            ).first().filter { it.watchProgress > 0.0 }
        } catch (e: Exception) {
            Log.w(TAG, "Failed to fetch watched items: ${e.message}")
            return
        }

        // 2. Clear existing Watch Next entries from this app
        removeAll()

        // 3. Add continue watching items
        val continueItems = filterContinueWatching(allItems)
        for (item in continueItems) {
            addWatchNextProgram(item, TvContractCompat.WatchNextPrograms.WATCH_NEXT_TYPE_CONTINUE)
        }

        // 4. For completed TV episodes, find and add the next episode
        val completedEpisodes = allItems.filter { isCompleted(it.watchProgress) && isTvSeriesType(it.mediaType) }
        for (episode in completedEpisodes) {
            val nextEpisode = findNextEpisode(episode)
            if (nextEpisode != null) {
                addWatchNextProgram(nextEpisode, TvContractCompat.WatchNextPrograms.WATCH_NEXT_TYPE_NEXT)
            }
        }

        Log.d(TAG, "Watch Next refreshed: ${continueItems.size} continue, ${completedEpisodes.size} completed episodes checked")
    }

    private suspend fun findNextEpisode(currentEpisode: MediaItem): MediaItem? {
        // The entity hierarchy uses parent_id self-reference.
        // We fetch the parent (show/season) and look for the next episode.
        return try {
            // For now, fetch siblings via browse with the same type
            // The API entity hierarchy handles this through parent_id relationships
            val item = mediaRepository.getMediaById(currentEpisode.id).first() ?: return null
            // A full next-episode resolver would query the parent's children.
            // This is a simplified version that relies on the API's recommendation endpoint.
            val similar = mediaRepository.getSimilarItems(currentEpisode.id)
            similar.firstOrNull { it.mediaType == "tv_episode" && it.id != currentEpisode.id }
        } catch (e: Exception) {
            Log.w(TAG, "Failed to find next episode for ${currentEpisode.id}: ${e.message}")
            null
        }
    }

    private fun addWatchNextProgram(item: MediaItem, watchNextType: Int) {
        try {
            val values = ChannelProgramMapper.toWatchNextValues(item, watchNextType)
            context.contentResolver.insert(TvContractCompat.WatchNextPrograms.CONTENT_URI, values)
        } catch (e: Exception) {
            Log.w(TAG, "Failed to add Watch Next for ${item.id}: ${e.message}")
        }
    }

    fun removeAll() {
        try {
            // Delete all Watch Next programs that belong to this app
            // The system identifies them by the intent URI scheme
            context.contentResolver.delete(
                TvContractCompat.WatchNextPrograms.CONTENT_URI,
                null, null
            )
        } catch (e: Exception) {
            Log.w(TAG, "Failed to clear Watch Next: ${e.message}")
        }
    }
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd catalogizer-androidtv && ./gradlew testDebugUnitTest --tests "com.catalogizer.androidtv.data.tv.WatchNextManagerTest" 2>&1 | tail -15`
Expected: All 11 tests PASSED.

- [ ] **Step 5: Commit**

```bash
cd catalogizer-androidtv && git add app/src/main/java/com/catalogizer/androidtv/data/tv/WatchNextManager.kt app/src/test/java/com/catalogizer/androidtv/data/tv/WatchNextManagerTest.kt
git commit -m "feat: add WatchNextManager for system Watch Next row"
```

---

### Task 7: ChannelDeepLinkActivity

**Files:**
- Create: `catalogizer-androidtv/app/src/main/java/com/catalogizer/androidtv/ui/ChannelDeepLinkActivity.kt`
- Modify: `catalogizer-androidtv/app/src/main/AndroidManifest.xml`
- Test: `catalogizer-androidtv/app/src/test/java/com/catalogizer/androidtv/ui/ChannelDeepLinkActivityTest.kt`

- [ ] **Step 1: Write the failing tests**

Create `catalogizer-androidtv/app/src/test/java/com/catalogizer/androidtv/ui/ChannelDeepLinkActivityTest.kt`:

```kotlin
package com.catalogizer.androidtv.ui

import android.content.Intent
import android.net.Uri
import org.junit.Assert.*
import org.junit.Test

class ChannelDeepLinkActivityTest {

    @Test
    fun `parseDeepLink extracts mediaId from URI`() {
        val uri = Uri.parse("catalogizer://media/42?type=movie")
        val result = DeepLinkParser.parse(uri)
        assertEquals(42L, result.mediaId)
    }

    @Test
    fun `parseDeepLink extracts mediaType from URI`() {
        val uri = Uri.parse("catalogizer://media/42?type=movie")
        val result = DeepLinkParser.parse(uri)
        assertEquals("movie", result.mediaType)
    }

    @Test
    fun `parseDeepLink handles missing type parameter`() {
        val uri = Uri.parse("catalogizer://media/42")
        val result = DeepLinkParser.parse(uri)
        assertEquals(42L, result.mediaId)
        assertNull(result.mediaType)
    }

    @Test
    fun `parseDeepLink returns null mediaId for invalid path`() {
        val uri = Uri.parse("catalogizer://media/invalid")
        val result = DeepLinkParser.parse(uri)
        assertNull(result.mediaId)
    }

    @Test
    fun `parseDeepLink returns null mediaId for empty path`() {
        val uri = Uri.parse("catalogizer://media")
        val result = DeepLinkParser.parse(uri)
        assertNull(result.mediaId)
    }

    @Test
    fun `isAudioWithoutContext returns true for music without metadata`() {
        assertTrue(DeepLinkParser.isAudioWithoutContext("music", hasExternalMetadata = false))
    }

    @Test
    fun `isAudioWithoutContext returns false for music with metadata`() {
        assertFalse(DeepLinkParser.isAudioWithoutContext("music", hasExternalMetadata = true))
    }

    @Test
    fun `isAudioWithoutContext returns false for non-audio types`() {
        assertFalse(DeepLinkParser.isAudioWithoutContext("movie", hasExternalMetadata = false))
    }

    @Test
    fun `isAudioWithoutContext returns true for audiobook and podcast without metadata`() {
        assertTrue(DeepLinkParser.isAudioWithoutContext("audiobook", hasExternalMetadata = false))
        assertTrue(DeepLinkParser.isAudioWithoutContext("podcast", hasExternalMetadata = false))
    }
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd catalogizer-androidtv && ./gradlew testDebugUnitTest --tests "com.catalogizer.androidtv.ui.ChannelDeepLinkActivityTest" 2>&1 | tail -10`
Expected: Compilation error — `DeepLinkParser` does not exist.

- [ ] **Step 3: Implement ChannelDeepLinkActivity and DeepLinkParser**

Create `catalogizer-androidtv/app/src/main/java/com/catalogizer/androidtv/ui/ChannelDeepLinkActivity.kt`:

```kotlin
package com.catalogizer.androidtv.ui

import android.content.Intent
import android.net.Uri
import android.os.Bundle
import android.util.Log
import androidx.activity.ComponentActivity
import com.catalogizer.androidtv.DependencyContainer
import com.catalogizer.androidtv.data.tv.LaunchAction
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.launch

/**
 * Parsed deep link data from a `catalogizer://media/{id}` URI.
 */
data class DeepLinkResult(
    val mediaId: Long?,
    val mediaType: String?
)

/**
 * Parses `catalogizer://media/{id}?type={type}` URIs from TV home screen channel clicks.
 */
object DeepLinkParser {
    private val AUDIO_TYPES = setOf("music", "audiobook", "podcast")

    fun parse(uri: Uri?): DeepLinkResult {
        if (uri == null) return DeepLinkResult(null, null)
        val mediaId = uri.pathSegments?.firstOrNull()?.toLongOrNull()
        val mediaType = uri.getQueryParameter("type")
        return DeepLinkResult(mediaId, mediaType)
    }

    fun isAudioWithoutContext(mediaType: String?, hasExternalMetadata: Boolean): Boolean {
        return mediaType in AUDIO_TYPES && !hasExternalMetadata
    }
}

/**
 * Transparent activity that handles deep link intents from Android TV home screen channels.
 * Routes to MediaDetailScreen or Player based on per-category launch settings.
 * If the user is not authenticated, redirects to LoginScreen with the pending deep link.
 */
class ChannelDeepLinkActivity : ComponentActivity() {
    companion object {
        private const val TAG = "ChannelDeepLink"
        const val EXTRA_MEDIA_ID = "deep_link_media_id"
        const val EXTRA_MEDIA_TYPE = "deep_link_media_type"
        const val EXTRA_ACTION = "deep_link_action"
    }

    private val scope = CoroutineScope(SupervisorJob() + Dispatchers.Main)

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)

        val deepLink = DeepLinkParser.parse(intent?.data)
        Log.d(TAG, "Deep link received: mediaId=${deepLink.mediaId}, type=${deepLink.mediaType}")

        if (deepLink.mediaId == null) {
            Log.w(TAG, "Invalid deep link — no media ID")
            launchMainActivity()
            finish()
            return
        }

        scope.launch {
            resolveAndRoute(deepLink)
            finish()
        }
    }

    private suspend fun resolveAndRoute(deepLink: DeepLinkResult) {
        val container = DependencyContainer.getInstance(this)
        val authState = container.authRepository.authState.first()

        if (!authState.isAuthenticated) {
            // Pass deep link to MainActivity so it can resume after login
            val mainIntent = Intent(this, MainActivity::class.java).apply {
                putExtra(EXTRA_MEDIA_ID, deepLink.mediaId)
                putExtra(EXTRA_MEDIA_TYPE, deepLink.mediaType)
                flags = Intent.FLAG_ACTIVITY_NEW_TASK or Intent.FLAG_ACTIVITY_CLEAR_TOP
            }
            startActivity(mainIntent)
            return
        }

        // Determine launch action
        val launchAction = if (deepLink.mediaType != null) {
            container.settingsRepository.getLaunchAction(deepLink.mediaType)
        } else {
            LaunchAction.DETAIL
        }

        // Check audio-without-context override
        val shouldPlayImmediately = if (DeepLinkParser.isAudioWithoutContext(
                deepLink.mediaType, hasExternalMetadata = false
            )) {
            // For audio, check if it has external metadata by fetching the item
            try {
                val item = container.mediaRepository.getMediaById(deepLink.mediaId!!).first()
                val hasMetadata = item?.externalMetadata?.isNotEmpty() == true
                if (!hasMetadata) true else launchAction == LaunchAction.IMMEDIATE_PLAY
            } catch (e: Exception) {
                launchAction == LaunchAction.IMMEDIATE_PLAY
            }
        } else {
            launchAction == LaunchAction.IMMEDIATE_PLAY
        }

        val action = if (shouldPlayImmediately) "play" else "detail"

        val mainIntent = Intent(this, MainActivity::class.java).apply {
            putExtra(EXTRA_MEDIA_ID, deepLink.mediaId)
            putExtra(EXTRA_MEDIA_TYPE, deepLink.mediaType)
            putExtra(EXTRA_ACTION, action)
            flags = Intent.FLAG_ACTIVITY_NEW_TASK or Intent.FLAG_ACTIVITY_CLEAR_TOP
        }
        startActivity(mainIntent)
    }

    private fun launchMainActivity() {
        val mainIntent = Intent(this, MainActivity::class.java).apply {
            flags = Intent.FLAG_ACTIVITY_NEW_TASK or Intent.FLAG_ACTIVITY_CLEAR_TOP
        }
        startActivity(mainIntent)
    }
}
```

- [ ] **Step 4: Register ChannelDeepLinkActivity in AndroidManifest.xml**

In `catalogizer-androidtv/app/src/main/AndroidManifest.xml`, add after the `MediaPlayerActivity` block (after line 66) and before the MediaPlaybackService:

```xml
        <!-- Channel Deep Link Activity -->
        <activity
            android:name=".ui.ChannelDeepLinkActivity"
            android:exported="true"
            android:theme="@android:style/Theme.Translucent.NoTitleBar"
            android:screenOrientation="landscape">
            <intent-filter>
                <action android:name="android.intent.action.VIEW" />
                <category android:name="android.intent.category.DEFAULT" />
                <category android:name="android.intent.category.BROWSABLE" />
                <data
                    android:scheme="catalogizer"
                    android:host="media" />
            </intent-filter>
        </activity>
```

- [ ] **Step 5: Run tests to verify they pass**

Run: `cd catalogizer-androidtv && ./gradlew testDebugUnitTest --tests "com.catalogizer.androidtv.ui.ChannelDeepLinkActivityTest" 2>&1 | tail -15`
Expected: All 9 tests PASSED.

- [ ] **Step 6: Commit**

```bash
cd catalogizer-androidtv && git add app/src/main/java/com/catalogizer/androidtv/ui/ChannelDeepLinkActivity.kt app/src/main/AndroidManifest.xml app/src/test/java/com/catalogizer/androidtv/ui/ChannelDeepLinkActivityTest.kt
git commit -m "feat: add ChannelDeepLinkActivity for TV home screen deep links"
```

---

### Task 8: TvChannelSyncWorker

**Files:**
- Create: `catalogizer-androidtv/app/src/main/java/com/catalogizer/androidtv/data/tv/TvChannelSyncWorker.kt`
- Test: `catalogizer-androidtv/app/src/test/java/com/catalogizer/androidtv/data/tv/TvChannelSyncWorkerTest.kt`

- [ ] **Step 1: Write the failing tests**

Create `catalogizer-androidtv/app/src/test/java/com/catalogizer/androidtv/data/tv/TvChannelSyncWorkerTest.kt`:

```kotlin
package com.catalogizer.androidtv.data.tv

import org.junit.Assert.*
import org.junit.Test

class TvChannelSyncWorkerTest {

    @Test
    fun `SYNC_INTERVAL_HOURS is 6`() {
        assertEquals(6L, TvChannelSyncWorker.SYNC_INTERVAL_HOURS)
    }

    @Test
    fun `WORK_NAME is correct`() {
        assertEquals("tv_channel_sync", TvChannelSyncWorker.WORK_NAME)
    }
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd catalogizer-androidtv && ./gradlew testDebugUnitTest --tests "com.catalogizer.androidtv.data.tv.TvChannelSyncWorkerTest" 2>&1 | tail -10`
Expected: Compilation error — `TvChannelSyncWorker` does not exist.

- [ ] **Step 3: Implement TvChannelSyncWorker**

Create `catalogizer-androidtv/app/src/main/java/com/catalogizer/androidtv/data/tv/TvChannelSyncWorker.kt`:

```kotlin
package com.catalogizer.androidtv.data.tv

import android.content.Context
import android.util.Log
import androidx.work.*
import com.catalogizer.androidtv.DependencyContainer
import java.util.concurrent.TimeUnit

/**
 * Periodic [CoroutineWorker] that refreshes Android TV home screen channels
 * every [SYNC_INTERVAL_HOURS] hours. Requires network connectivity.
 * Skips execution if the user is not authenticated.
 */
class TvChannelSyncWorker(
    context: Context,
    params: WorkerParameters
) : CoroutineWorker(context, params) {

    companion object {
        private const val TAG = "TvChannelSync"
        const val WORK_NAME = "tv_channel_sync"
        const val SYNC_INTERVAL_HOURS = 6L

        fun enqueue(context: Context) {
            val constraints = Constraints.Builder()
                .setRequiredNetworkType(NetworkType.CONNECTED)
                .setRequiresBatteryNotLow(true)
                .build()

            val request = PeriodicWorkRequestBuilder<TvChannelSyncWorker>(
                SYNC_INTERVAL_HOURS, TimeUnit.HOURS
            )
                .setConstraints(constraints)
                .setBackoffCriteria(BackoffPolicy.EXPONENTIAL, 30, TimeUnit.MINUTES)
                .build()

            WorkManager.getInstance(context).enqueueUniquePeriodicWork(
                WORK_NAME,
                ExistingPeriodicWorkPolicy.KEEP,
                request
            )
            Log.d(TAG, "Periodic sync enqueued (every ${SYNC_INTERVAL_HOURS}h)")
        }

        fun cancel(context: Context) {
            WorkManager.getInstance(context).cancelUniqueWork(WORK_NAME)
            Log.d(TAG, "Periodic sync cancelled")
        }
    }

    override suspend fun doWork(): Result {
        Log.d(TAG, "Sync worker started")

        val container = DependencyContainer.getInstance(applicationContext)

        // Check if user is authenticated
        try {
            val authState = container.authRepository.authState.value
            if (!authState.isAuthenticated) {
                Log.d(TAG, "User not authenticated, skipping sync")
                return Result.success()
            }
        } catch (e: Exception) {
            Log.w(TAG, "Failed to check auth state: ${e.message}")
            return Result.retry()
        }

        // Refresh all channels
        return try {
            val tvChannelRepo = TvChannelRepository(
                applicationContext,
                container.mediaRepository,
                container.settingsRepository
            )
            tvChannelRepo.refreshAllChannels()

            val watchNextManager = WatchNextManager(applicationContext, container.mediaRepository)
            watchNextManager.refreshWatchNext()

            Log.d(TAG, "Sync worker completed successfully")
            Result.success()
        } catch (e: Exception) {
            Log.e(TAG, "Sync worker failed: ${e.message}")
            Result.retry()
        }
    }
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd catalogizer-androidtv && ./gradlew testDebugUnitTest --tests "com.catalogizer.androidtv.data.tv.TvChannelSyncWorkerTest" 2>&1 | tail -10`
Expected: All 2 tests PASSED.

- [ ] **Step 5: Commit**

```bash
cd catalogizer-androidtv && git add app/src/main/java/com/catalogizer/androidtv/data/tv/TvChannelSyncWorker.kt app/src/test/java/com/catalogizer/androidtv/data/tv/TvChannelSyncWorkerTest.kt
git commit -m "feat: add TvChannelSyncWorker for periodic background channel refresh"
```

---

### Task 9: Wire Into DependencyContainer and Application Lifecycle

**Files:**
- Modify: `catalogizer-androidtv/app/src/main/java/com/catalogizer/androidtv/DependencyContainer.kt`
- Modify: `catalogizer-androidtv/app/src/main/java/com/catalogizer/androidtv/CatalogizerTVApplication.kt`

- [ ] **Step 1: Add TvChannelRepository and WatchNextManager to DependencyContainer**

In `catalogizer-androidtv/app/src/main/java/com/catalogizer/androidtv/DependencyContainer.kt`, add imports at the top:

```kotlin
import com.catalogizer.androidtv.data.tv.TvChannelRepository
import com.catalogizer.androidtv.data.tv.WatchNextManager
```

Then add these properties after the `mediaRepository` getter (after line 104):

```kotlin
    val tvChannelRepository: TvChannelRepository
        get() = TvChannelRepository(context, mediaRepository, settingsRepository)

    val watchNextManager: WatchNextManager
        get() = WatchNextManager(context, mediaRepository)
```

- [ ] **Step 2: Add channel initialization to CatalogizerTVApplication**

In `catalogizer-androidtv/app/src/main/java/com/catalogizer/androidtv/CatalogizerTVApplication.kt`, add import:

```kotlin
import com.catalogizer.androidtv.data.tv.TvChannelSyncWorker
```

Then modify the `onCreate` method to enqueue the sync worker after initialization:

```kotlin
    override fun onCreate() {
        super.onCreate()

        // Initialize dependency container and load persisted settings (server URL, etc.)
        appScope.launch {
            dependencyContainer.initializeAsync()

            // Initialize default channel and enqueue periodic sync
            try {
                dependencyContainer.tvChannelRepository.initializeDefaultChannel()
                TvChannelSyncWorker.enqueue(this@CatalogizerTVApplication)
            } catch (e: Exception) {
                // Channel initialization can fail if not authenticated yet — that's OK
                android.util.Log.w("CatalogizerTV", "Channel init deferred: ${e.message}")
            }
        }
    }
```

- [ ] **Step 3: Verify build compiles**

Run: `cd catalogizer-androidtv && ./gradlew assembleDebug 2>&1 | tail -10`
Expected: BUILD SUCCESSFUL.

- [ ] **Step 4: Commit**

```bash
cd catalogizer-androidtv && git add app/src/main/java/com/catalogizer/androidtv/DependencyContainer.kt app/src/main/java/com/catalogizer/androidtv/CatalogizerTVApplication.kt
git commit -m "feat: wire TvChannelRepository and sync worker into app lifecycle"
```

---

### Task 10: Trigger Channel Refresh From HomeViewModel and SyncService

**Files:**
- Modify: `catalogizer-androidtv/app/src/main/java/com/catalogizer/androidtv/ui/viewmodel/HomeViewModel.kt`
- Modify: `catalogizer-androidtv/app/src/main/java/com/catalogizer/androidtv/data/sync/SyncService.kt`

- [ ] **Step 1: Add channel refresh to HomeViewModel**

In `catalogizer-androidtv/app/src/main/java/com/catalogizer/androidtv/ui/viewmodel/HomeViewModel.kt`, add a `tvChannelRepository` parameter and `watchNextManager`:

Change the class declaration from:

```kotlin
class HomeViewModel(
    private val mediaRepository: MediaRepository
) : ViewModel() {
```

To:

```kotlin
class HomeViewModel(
    private val mediaRepository: MediaRepository,
    private val tvChannelRepository: com.catalogizer.androidtv.data.tv.TvChannelRepository? = null,
    private val watchNextManager: com.catalogizer.androidtv.data.tv.WatchNextManager? = null
) : ViewModel() {
```

Then at the end of the `try` block in `loadHomeData()`, after `_uiState.update { ... }` (after line 119), add:

```kotlin
                    // Refresh TV home screen channels (non-blocking)
                    launch {
                        try {
                            tvChannelRepository?.refreshAllChannels()
                            watchNextManager?.refreshWatchNext()
                        } catch (e: Exception) {
                            // Channel refresh failure doesn't affect home screen UX
                            android.util.Log.w("HomeVM", "Channel refresh failed: ${e.message}")
                        }
                    }
```

- [ ] **Step 2: Update DependencyContainer to pass new parameters**

In `catalogizer-androidtv/app/src/main/java/com/catalogizer/androidtv/DependencyContainer.kt`, update `createHomeViewModel`:

Change:

```kotlin
    fun createHomeViewModel(): HomeViewModel = HomeViewModel(mediaRepository)
```

To:

```kotlin
    fun createHomeViewModel(): HomeViewModel = HomeViewModel(
        mediaRepository, tvChannelRepository, watchNextManager
    )
```

- [ ] **Step 3: Add channel refresh to SyncService**

In `catalogizer-androidtv/app/src/main/java/com/catalogizer/androidtv/data/sync/SyncService.kt`, add import and update `performSync()`:

Add import:

```kotlin
import com.catalogizer.androidtv.DependencyContainer
import kotlinx.coroutines.launch
```

Update `performSync()`:

```kotlin
    private fun performSync() {
        serviceScope.launch {
            // Existing sync logic here...

            // Refresh TV home screen channels after sync
            try {
                val container = DependencyContainer.getInstance(this@SyncService)
                container.tvChannelRepository.refreshAllChannels()
                container.watchNextManager.refreshWatchNext()
            } catch (e: Exception) {
                android.util.Log.w("SyncService", "Channel refresh failed: ${e.message}")
            }
        }
    }
```

- [ ] **Step 4: Verify build compiles**

Run: `cd catalogizer-androidtv && ./gradlew assembleDebug 2>&1 | tail -10`
Expected: BUILD SUCCESSFUL.

- [ ] **Step 5: Run existing tests to verify no regressions**

Run: `cd catalogizer-androidtv && ./gradlew testDebugUnitTest 2>&1 | tail -20`
Expected: All existing tests still pass (the optional parameters default to null, so existing test code continues to work).

- [ ] **Step 6: Commit**

```bash
cd catalogizer-androidtv && git add app/src/main/java/com/catalogizer/androidtv/ui/viewmodel/HomeViewModel.kt app/src/main/java/com/catalogizer/androidtv/data/sync/SyncService.kt app/src/main/java/com/catalogizer/androidtv/DependencyContainer.kt
git commit -m "feat: trigger channel refresh from HomeViewModel and SyncService"
```

---

### Task 11: Deep Link Handling in MainActivity

**Files:**
- Modify: `catalogizer-androidtv/app/src/main/java/com/catalogizer/androidtv/ui/MainActivity.kt`
- Modify: `catalogizer-androidtv/app/src/main/java/com/catalogizer/androidtv/ui/navigation/TVNavigation.kt`

- [ ] **Step 1: Read current MainActivity**

Read `catalogizer-androidtv/app/src/main/java/com/catalogizer/androidtv/ui/MainActivity.kt` to understand the current structure, then proceed with modifications.

- [ ] **Step 2: Add deep link intent handling to MainActivity**

In `MainActivity.kt`, add handling for the deep link extras passed from `ChannelDeepLinkActivity`. After the activity initializes, check for `EXTRA_MEDIA_ID` and navigate accordingly.

Add to the `onCreate` method, after the navigation setup:

```kotlin
        // Handle deep link from ChannelDeepLinkActivity
        val deepLinkMediaId = intent?.getLongExtra(
            ChannelDeepLinkActivity.EXTRA_MEDIA_ID, -1L
        ) ?: -1L
        val deepLinkAction = intent?.getStringExtra(
            ChannelDeepLinkActivity.EXTRA_ACTION
        )
```

Pass `deepLinkMediaId` and `deepLinkAction` to the `TVNavigation` composable.

- [ ] **Step 3: Add deep link navigation to TVNavigation**

In `catalogizer-androidtv/app/src/main/java/com/catalogizer/androidtv/ui/navigation/TVNavigation.kt`, add `deepLinkMediaId` and `deepLinkAction` parameters to the `TVNavigation` composable:

```kotlin
@Composable
fun TVNavigation(
    isAuthenticated: Boolean,
    authViewModel: AuthViewModel,
    homeViewModel: HomeViewModel,
    searchViewModel: SearchViewModel,
    deepLinkMediaId: Long = -1L,
    deepLinkAction: String? = null,
    navController: NavHostController = rememberNavController()
) {
```

Add a `LaunchedEffect` after `NavHost` to handle the deep link on first composition:

```kotlin
    // Handle deep link navigation
    LaunchedEffect(deepLinkMediaId) {
        if (deepLinkMediaId > 0 && isAuthenticated) {
            when (deepLinkAction) {
                "play" -> navController.navigate(TVScreen.Player.createRoute(deepLinkMediaId))
                else -> navController.navigate(TVScreen.MediaDetail.createRoute(deepLinkMediaId))
            }
        }
    }
```

- [ ] **Step 4: Verify build compiles**

Run: `cd catalogizer-androidtv && ./gradlew assembleDebug 2>&1 | tail -10`
Expected: BUILD SUCCESSFUL.

- [ ] **Step 5: Commit**

```bash
cd catalogizer-androidtv && git add app/src/main/java/com/catalogizer/androidtv/ui/MainActivity.kt app/src/main/java/com/catalogizer/androidtv/ui/navigation/TVNavigation.kt
git commit -m "feat: handle deep link navigation from TV home screen channels"
```

---

### Task 12: Logout Cleanup

**Files:**
- Modify: `catalogizer-androidtv/app/src/main/java/com/catalogizer/androidtv/ui/viewmodel/AuthViewModel.kt`

- [ ] **Step 1: Add channel cleanup to AuthViewModel.logout()**

In `catalogizer-androidtv/app/src/main/java/com/catalogizer/androidtv/ui/viewmodel/AuthViewModel.kt`, add constructor parameters:

```kotlin
class AuthViewModel(
    private val authRepository: AuthRepository,
    private val tvChannelRepository: com.catalogizer.androidtv.data.tv.TvChannelRepository? = null,
    private val watchNextManager: com.catalogizer.androidtv.data.tv.WatchNextManager? = null,
    private val context: android.content.Context? = null,
    private val settingsRepository: com.catalogizer.androidtv.data.repository.SettingsRepository? = null
) : ViewModel() {
```

Update `logout()`:

```kotlin
    fun logout() {
        viewModelScope.launch {
            // Clean up TV channels before logout
            try {
                tvChannelRepository?.deleteAllChannels()
                watchNextManager?.removeAll()
                context?.let { com.catalogizer.androidtv.data.tv.TvChannelSyncWorker.cancel(it) }
                settingsRepository?.clearAllChannelIds()
                settingsRepository?.clearAllLaunchActions()
            } catch (e: Exception) {
                android.util.Log.w("AuthVM", "Channel cleanup failed: ${e.message}")
            }

            authRepository.logout()
        }
    }
```

- [ ] **Step 2: Update DependencyContainer.createAuthViewModel()**

In `DependencyContainer.kt`, update:

```kotlin
    fun createAuthViewModel(): AuthViewModel = AuthViewModel(
        authRepository, tvChannelRepository, watchNextManager, context, settingsRepository
    )
```

- [ ] **Step 3: Run existing AuthViewModel tests to verify no regression**

Run: `cd catalogizer-androidtv && ./gradlew testDebugUnitTest --tests "com.catalogizer.androidtv.ui.viewmodel.AuthViewModelTest" 2>&1 | tail -15`
Expected: All tests pass (optional params default to null).

- [ ] **Step 4: Commit**

```bash
cd catalogizer-androidtv && git add app/src/main/java/com/catalogizer/androidtv/ui/viewmodel/AuthViewModel.kt app/src/main/java/com/catalogizer/androidtv/DependencyContainer.kt
git commit -m "feat: clean up TV channels and Watch Next on logout"
```

---

### Task 13: Channel Tap Behavior Settings UI

**Files:**
- Create: `catalogizer-androidtv/app/src/main/java/com/catalogizer/androidtv/ui/screens/settings/ChannelSettingsSection.kt`
- Modify: `catalogizer-androidtv/app/src/main/java/com/catalogizer/androidtv/ui/screens/settings/SettingsScreen.kt`
- Modify: `catalogizer-androidtv/app/src/main/java/com/catalogizer/androidtv/ui/viewmodel/SettingsViewModel.kt`

- [ ] **Step 1: Add channel settings methods to SettingsViewModel**

Read `catalogizer-androidtv/app/src/main/java/com/catalogizer/androidtv/ui/viewmodel/SettingsViewModel.kt` first. Then add methods for loading and saving per-category launch actions:

```kotlin
    private val _channelLaunchActions = MutableStateFlow<Map<String, LaunchAction>>(emptyMap())
    val channelLaunchActions: StateFlow<Map<String, LaunchAction>> = _channelLaunchActions.asStateFlow()

    private val _activeMediaTypes = MutableStateFlow<List<String>>(emptyList())
    val activeMediaTypes: StateFlow<List<String>> = _activeMediaTypes.asStateFlow()

    fun loadChannelSettings(mediaRepository: MediaRepository) {
        viewModelScope.launch {
            try {
                val actions = settingsRepository.getAllLaunchActions()
                _channelLaunchActions.value = actions

                val (_, byType) = mediaRepository.getEntityStats()
                _activeMediaTypes.value = byType.filter { it.value > 0 }.keys.toList()
            } catch (e: Exception) {
                // Defaults will be used
            }
        }
    }

    fun updateLaunchAction(mediaType: String, action: LaunchAction) {
        viewModelScope.launch {
            settingsRepository.saveLaunchAction(mediaType, action)
            _channelLaunchActions.update { it + (mediaType to action) }
        }
    }
```

Add necessary imports:

```kotlin
import com.catalogizer.androidtv.data.tv.LaunchAction
import com.catalogizer.androidtv.data.repository.MediaRepository
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
```

- [ ] **Step 2: Create ChannelSettingsSection composable**

Create `catalogizer-androidtv/app/src/main/java/com/catalogizer/androidtv/ui/screens/settings/ChannelSettingsSection.kt`:

```kotlin
package com.catalogizer.androidtv.ui.screens.settings

import androidx.compose.foundation.BorderStroke
import androidx.compose.foundation.border
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.selection.selectable
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.focus.onFocusChanged
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.semantics.Role
import androidx.compose.ui.unit.dp
import androidx.compose.material3.RadioButton
import androidx.compose.material3.RadioButtonDefaults
import androidx.tv.material3.*
import com.catalogizer.androidtv.data.models.MediaType
import com.catalogizer.androidtv.data.tv.LaunchAction

private val TextPrimary = Color(0xFFFFFFFF)
private val TextSecondary = Color(0xFFE0E0E0)
private val FocusBorderColor = Color(0xFF9ECAFF)

/**
 * Settings section that lets users configure per-media-type behavior when clicking
 * a program card on the Android TV home screen channel: open detail or play immediately.
 */
@OptIn(ExperimentalTvMaterial3Api::class)
@Composable
fun ChannelSettingsSection(
    activeMediaTypes: List<String>,
    launchActions: Map<String, LaunchAction>,
    onUpdateAction: (String, LaunchAction) -> Unit
) {
    if (activeMediaTypes.isEmpty()) return

    Column(
        modifier = Modifier.fillMaxWidth(),
        verticalArrangement = Arrangement.spacedBy(8.dp)
    ) {
        for (typeValue in activeMediaTypes) {
            val mediaType = MediaType.fromValue(typeValue)
            val currentAction = launchActions[typeValue] ?: LaunchAction.DETAIL

            Column(modifier = Modifier.fillMaxWidth().padding(vertical = 4.dp)) {
                Text(
                    text = mediaType.displayName,
                    style = MaterialTheme.typography.bodyLarge,
                    color = TextPrimary
                )
                Spacer(modifier = Modifier.height(4.dp))
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.spacedBy(12.dp)
                ) {
                    LaunchAction.values().forEach { action ->
                        val label = when (action) {
                            LaunchAction.DETAIL -> "Detail Screen"
                            LaunchAction.IMMEDIATE_PLAY -> "Play Immediately"
                        }
                        val isSelected = currentAction == action
                        var rowFocused by remember { mutableStateOf(false) }

                        Row(
                            modifier = Modifier
                                .onFocusChanged { rowFocused = it.isFocused || it.hasFocus }
                                .then(
                                    if (rowFocused) Modifier.border(
                                        BorderStroke(2.dp, FocusBorderColor),
                                        shape = RoundedCornerShape(6.dp)
                                    ) else Modifier
                                )
                                .selectable(
                                    selected = isSelected,
                                    onClick = { onUpdateAction(typeValue, action) },
                                    role = Role.RadioButton
                                )
                                .padding(vertical = 4.dp, horizontal = 4.dp),
                            verticalAlignment = Alignment.CenterVertically
                        ) {
                            RadioButton(
                                selected = isSelected,
                                onClick = null,
                                colors = RadioButtonDefaults.colors(
                                    selectedColor = FocusBorderColor,
                                    unselectedColor = TextSecondary
                                )
                            )
                            Spacer(modifier = Modifier.width(4.dp))
                            Text(
                                text = label,
                                style = MaterialTheme.typography.bodyMedium,
                                color = if (isSelected) TextPrimary else TextSecondary
                            )
                        }
                    }
                }
            }
        }
    }
}
```

- [ ] **Step 3: Add Channel Tap Behavior section to SettingsScreen**

In `catalogizer-androidtv/app/src/main/java/com/catalogizer/androidtv/ui/screens/settings/SettingsScreen.kt`, add the new section in the `TvLazyColumn`, before the Account section (before line 364). Add these imports:

```kotlin
import com.catalogizer.androidtv.data.tv.LaunchAction
```

Then add a new `item` block:

```kotlin
                // ── Channel Tap Behavior ────────────────────────────────────────
                item {
                    val activeMediaTypes by settingsViewModel.activeMediaTypes.collectAsStateWithLifecycle()
                    val launchActions by settingsViewModel.channelLaunchActions.collectAsStateWithLifecycle()

                    SettingsSection(title = "Channel Tap Behavior") {
                        Text(
                            text = "Choose what happens when you select an item from a home screen channel.",
                            style = MaterialTheme.typography.bodySmall,
                            color = TextSecondary,
                            modifier = Modifier.padding(bottom = 8.dp)
                        )
                        ChannelSettingsSection(
                            activeMediaTypes = activeMediaTypes,
                            launchActions = launchActions,
                            onUpdateAction = { type, action ->
                                settingsViewModel.updateLaunchAction(type, action)
                            }
                        )
                    }
                }
```

Also, the `SettingsScreen` composable needs to trigger loading channel settings. Add to the existing `LaunchedEffect(Unit)`:

```kotlin
    LaunchedEffect(Unit) {
        settingsViewModel.loadSettings()
        val container = com.catalogizer.androidtv.DependencyContainer.getInstance(
            androidx.compose.ui.platform.LocalContext.current
        )
        settingsViewModel.loadChannelSettings(container.mediaRepository)
    }
```

- [ ] **Step 4: Verify build compiles**

Run: `cd catalogizer-androidtv && ./gradlew assembleDebug 2>&1 | tail -10`
Expected: BUILD SUCCESSFUL.

- [ ] **Step 5: Commit**

```bash
cd catalogizer-androidtv && git add app/src/main/java/com/catalogizer/androidtv/ui/screens/settings/ChannelSettingsSection.kt app/src/main/java/com/catalogizer/androidtv/ui/screens/settings/SettingsScreen.kt app/src/main/java/com/catalogizer/androidtv/ui/viewmodel/SettingsViewModel.kt
git commit -m "feat: add Channel Tap Behavior settings UI for per-category launch action"
```

---

### Task 14: Run Full Test Suite and Fix Any Regressions

**Files:**
- No new files — verification only

- [ ] **Step 1: Run all unit tests**

Run: `cd catalogizer-androidtv && ./gradlew testDebugUnitTest 2>&1 | tail -30`
Expected: All tests pass. If any fail, investigate and fix.

- [ ] **Step 2: Run lint**

Run: `cd catalogizer-androidtv && ./gradlew lint 2>&1 | tail -10`
Expected: No new errors introduced.

- [ ] **Step 3: Verify build**

Run: `cd catalogizer-androidtv && ./gradlew assembleDebug 2>&1 | tail -10`
Expected: BUILD SUCCESSFUL.

- [ ] **Step 4: Commit any fixes**

If any fixes were needed:

```bash
cd catalogizer-androidtv && git add -A
git commit -m "fix: resolve test regressions from TV channels integration"
```

---

### Task 15: Update Version and Documentation

**Files:**
- Modify: `catalogizer-androidtv/app/build.gradle.kts` (version bump)
- Modify: `catalogizer-androidtv/CLAUDE.md`

- [ ] **Step 1: Bump version**

In `catalogizer-androidtv/app/build.gradle.kts`, update:

```kotlin
        versionCode = 7
        versionName = "2.3.0"
```

- [ ] **Step 2: Update CLAUDE.md with TV channels documentation**

Add a new section to `catalogizer-androidtv/CLAUDE.md` after the "Media Playback" section under "## TV-Specific Considerations":

```markdown
### Home Screen Channels

The app integrates with Android TV's home screen channels API (`androidx.tvprovider`):

- **Default Channel ("Catalogizer Picks")**: Auto-added on first launch. Curated mix of continue watching + recently added + trending. Up to 30 items.
- **Category Channels**: One per media type with content. Created dynamically based on entity stats. Users add via "Customize channels".
- **Watch Next Row**: System-level row. Shows partially watched items (5%-90%) and auto-surfaces next TV episodes.
- **Deep Linking**: `catalogizer://media/{id}?type={type}` — handled by `ChannelDeepLinkActivity`. Per-category launch behavior configurable in Settings ("Detail Screen" or "Play Immediately").
- **Background Sync**: `TvChannelSyncWorker` runs every 6 hours via WorkManager. Also triggers on app launch and after SyncService runs.
- **Logout**: Deletes all channels, clears Watch Next, cancels sync worker.

Key files:
- `data/tv/TvChannelRepository.kt` — Channel/program CRUD
- `data/tv/ChannelProgramMapper.kt` — MediaItem → PreviewProgram conversion
- `data/tv/WatchNextManager.kt` — Watch Next row logic
- `data/tv/TvChannelSyncWorker.kt` — Periodic background sync
- `ui/ChannelDeepLinkActivity.kt` — Deep link intent router
- `ui/screens/settings/ChannelSettingsSection.kt` — Per-category tap behavior UI
```

- [ ] **Step 3: Commit**

```bash
cd catalogizer-androidtv && git add app/build.gradle.kts CLAUDE.md
git commit -m "docs: update version to 2.3.0 and add TV channels documentation"
```
