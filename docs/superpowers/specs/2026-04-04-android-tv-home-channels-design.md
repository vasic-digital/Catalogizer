# Android TV Home Screen Channels — Design Spec

**Date:** 2026-04-04
**Component:** catalogizer-androidtv
**Version:** Target v2.3.0

## Overview

Integrate full Android TV Home Screen Channels support into the Catalogizer Android TV app. The home screen will display Catalogizer content in three surfaces: a default curated channel, dynamic per-category channels, and the system Watch Next row. Clicking any item opens the app's detail screen (or plays immediately, based on per-category user settings).

## Decisions

| Decision | Choice | Rationale |
|---|---|---|
| Channel types | Default + per-category + Watch Next | Full TV channels experience users expect from a mature media app |
| Click behavior | Per-category configurable (detail vs. play) | Movies might auto-play, TV shows go to detail. Audio without album context always plays immediately |
| Sync strategy | Periodic WorkManager (6h) + on content change + on app launch | Maximum freshness without excessive battery/network use |
| Items per channel | Adaptive up to 30 | Show what you have; sparse channels with few items still look fine |
| Default channel content | Mix: recently added + trending + continue watching | Best first impression, surfaces content the user cares about |
| Category channels | Dynamic — one per media type with content | Types with zero content get no channel; new types auto-discovered |
| Watch Next behavior | Partially watched (5%-90%) + next episode for TV series | Standard Netflix/YouTube behavior; next episode auto-surfaces |

## Architecture

### Three Integration Surfaces

1. **Default Channel ("Catalogizer Picks")** — Auto-added to home screen on first app launch. Curated mix of recently added, trending, and continue watching content. Up to 30 items.

2. **Dynamic Category Channels** — One channel per media type that has content in the user's library. Created/destroyed dynamically based on entity stats from the API. Users add them via "Customize channels". Each shows up to 30 items, adaptive to available content.

3. **Watch Next Row** — System-level row shared across all apps. Populated with partially watched content (5%-90% progress) and next episodes for TV series.

### Data Flow

```
catalog-api (REST)
  -> TvChannelRepository
    -> ChannelProgramMapper (MediaItem -> PreviewProgram)
      -> TvContractCompat content provider (system)
        -> Android TV Home Screen

Triggers:
  - WorkManager periodic job (every 6h)
  - SyncService on content change
  - App launch (non-blocking, after home data loads)
```

## Component Design

### TvChannelRepository

Central class managing all interactions with the Android TV TvProvider content provider.

**Responsibilities:**
- Channel lifecycle: create, update, delete `PreviewChannel` entries
- Program population: insert/update/remove `PreviewProgram` and `WatchNextProgram` entries
- Channel-to-API mapping: maps each channel to an API query (entity type, sort, limit)
- Logo management: downloads channel logos and sets them on channels
- Deduplication: tracks published programs to avoid duplicates on refresh

**Channel ID Persistence:**
Stored in `DataStore<Preferences>` via `SettingsRepository`:
- `channel_id_default` -> Long (the "Catalogizer Picks" channel)
- `channel_id_{mediaType}` -> Long (per-category channels)

**Key Methods:**
```kotlin
class TvChannelRepository(
    private val context: Context,
    private val mediaRepository: MediaRepository,
    private val settingsRepository: SettingsRepository
) {
    suspend fun initializeDefaultChannel()
    suspend fun refreshAllChannels()
    suspend fun refreshDefaultChannel()
    suspend fun refreshCategoryChannel(type: String)
    suspend fun refreshWatchNext()
    suspend fun createCategoryChannels()
    suspend fun removeStaleCategoryChannels()
    suspend fun deleteAllChannels()
    fun getChannelIdForType(type: String): Long?
}
```

### ChannelProgramMapper

Converts `MediaItem` to `PreviewProgram` / `WatchNextProgram`.

**Field mapping:**

| MediaItem field | PreviewProgram field |
|---|---|
| `title` | `setTitle()` |
| `description` | `setDescription()` |
| `posterUrl` | `setPosterArtUri()` |
| `backdropUrl` | `setThumbnailUri()` |
| `duration` (seconds) | `setDurationMillis()` |
| `mediaType` | `setType()` |
| `rating` | `setReviewRating()` |
| `year` | `setReleaseDate()` |
| `genres` | `setGenre()` |
| `id` | Encoded into `setIntentUri()` deep link |
| `watchProgress` | `setLastPlaybackPositionMillis()` / `setDurationMillis()` |

**PreviewProgram type mapping:**

| Media Type | PreviewProgram Type |
|---|---|
| movie | TYPE_MOVIE |
| tv_show | TYPE_TV_SERIES |
| tv_episode | TYPE_TV_EPISODE |
| music | TYPE_TRACK |
| anime | TYPE_MOVIE |
| documentary | TYPE_MOVIE |
| concert | TYPE_EVENT |
| game | TYPE_GAME |
| software | TYPE_CLIP (TYPE_APP does not exist in tvprovider:1.0.0) |
| ebook | TYPE_CLIP |
| audiobook | TYPE_ALBUM |
| podcast | TYPE_CHANNEL |
| training | TYPE_CLIP |
| sports | TYPE_EVENT |
| news | TYPE_CLIP |
| other | TYPE_CLIP |

### WatchNextManager

Manages the system Watch Next row independently from channels.

**Population rules:**

| Condition | Action |
|---|---|
| `watchProgress` 5%-90% | Add/update with `WATCH_NEXT_TYPE_CONTINUE` |
| `watchProgress` > 90% (completed) | Remove from Watch Next |
| TV episode completed + next episode exists | Add next episode with `WATCH_NEXT_TYPE_NEXT` |
| Not watched for 30+ days | Remove (stale cleanup) |
| User dismisses from home screen | Respect dismissal; don't re-add on next refresh |

**Next episode resolution:** For TV shows, when an episode hits >90% progress:
1. Query the API for the parent show's children (seasons/episodes via entity hierarchy)
2. Find the next sequential episode (same season next number, or first of next season)
3. If found, add as `WATCH_NEXT_TYPE_NEXT`
4. If no next episode (series complete), just remove the completed one

### Deep Linking

**URI scheme:**
```
catalogizer://media/{mediaId}?type={mediaType}&action={detail|play}
```

**ChannelDeepLinkActivity** — Transparent activity that receives deep link intents and routes:

```
Intent received with catalogizer://media/{id}?type={type}
  |
  +- Is user authenticated?
  |   +- No -> Launch MainActivity (LoginScreen) with pending deep link
  |
  +- Check per-category launch setting for this type
  |   +- "immediate_play" -> Launch player directly
  |   +- "detail" (default) -> Launch MediaDetailScreen
  |
  +- Special case: audio without album/artist context
  |   +- Override to immediate play regardless of setting
  |
  +- Special case: Watch Next items
      +- Partially watched -> Resume at saved position (player)
      +- Next episode -> Open detail for the new episode
```

**Manifest registration:**
```xml
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

### Background Sync

**TvChannelSyncWorker** — `CoroutineWorker` running every 6 hours via WorkManager.

Constraints:
- Requires network connectivity
- Not during low battery
- Metered network OK (metadata is small)

Lifecycle:
- Enqueued on first launch (after login)
- Enqueued after each successful login
- Cancelled on logout

```kotlin
class TvChannelSyncWorker(
    context: Context,
    params: WorkerParameters
) : CoroutineWorker(context, params) {
    override suspend fun doWork(): Result {
        // 1. Check authentication (token in DataStore)
        // 2. Recreate API client from saved server URL + token
        // 3. TvChannelRepository.refreshAllChannels()
        // 4. Return Result.success() or Result.retry()
    }
}
```

**SyncService integration:** One-line addition at end of sync cycle to call `TvChannelRepository.refreshAllChannels()`.

**App launch:** `HomeViewModel.loadHomeData()` triggers non-blocking channel refresh after home data loads.

### Per-Category Launch Settings

**Data model:**
```kotlin
enum class LaunchAction { DETAIL, IMMEDIATE_PLAY }
```

Stored in `DataStore` via `SettingsRepository` as a map: `mediaType -> LaunchAction`. All types default to `DETAIL`.

**Settings UI:** New "Channel Tap Behavior" section in SettingsScreen. A list of all media types the user has content for, each with a toggle between "Detail Screen" and "Play Immediately".

## File Layout

### New Files

```
app/src/main/java/com/catalogizer/androidtv/
  data/tv/
    TvChannelRepository.kt          # Core channel/program CRUD
    TvChannelSyncWorker.kt          # WorkManager periodic sync
    ChannelProgramMapper.kt         # MediaItem -> PreviewProgram/WatchNextProgram
    WatchNextManager.kt             # Watch Next row logic
  ui/
    ChannelDeepLinkActivity.kt      # Deep link intent router
    screens/settings/
      ChannelSettingsSection.kt     # Per-category launch behavior UI

app/src/test/java/com/catalogizer/androidtv/
  data/tv/
    TvChannelRepositoryTest.kt
    ChannelProgramMapperTest.kt
    WatchNextManagerTest.kt
    TvChannelSyncWorkerTest.kt
  ui/
    ChannelDeepLinkActivityTest.kt
    ChannelSettingsTest.kt
```

### Modified Files

```
AndroidManifest.xml                   # ChannelDeepLinkActivity + RECEIVE_BOOT_COMPLETED
DependencyContainer.kt               # Wire TvChannelRepository
CatalogizerTVApplication.kt          # Enqueue WorkManager, initial channel setup
data/repository/SettingsRepository.kt # Channel ID persistence, launch action storage
data/sync/SyncService.kt             # Trigger channel refresh after sync
data/models/Settings.kt              # LaunchAction enum, channelLaunchActions map
ui/viewmodel/HomeViewModel.kt        # Trigger channel refresh after home data loads
ui/screens/settings/SettingsScreen.kt # "Channel Tap Behavior" section
app/build.gradle.kts                 # WorkManager dependency
```

### New Dependencies

```kotlin
implementation("androidx.work:work-runtime-ktx:2.9.0")
```

`androidx.tvprovider:tvprovider:1.0.0` already exists.

## Testing

### Unit Tests

| Test file | Covers |
|---|---|
| `TvChannelRepositoryTest.kt` | Channel creation, program insertion, dedup, stale removal |
| `ChannelProgramMapperTest.kt` | MediaItem -> PreviewProgram for all 16 media types |
| `WatchNextManagerTest.kt` | Progress thresholds (5%/90%), next episode, 30-day stale cleanup |
| `ChannelDeepLinkActivityTest.kt` | Intent parsing, auth gate, per-category routing, audio special case |
| `TvChannelSyncWorkerTest.kt` | Worker execution, retry, auth check |
| `ChannelSettingsTest.kt` | Per-category launch action persistence and retrieval |

### Key Test Scenarios

- Channel created only for types with content (mock stats returning subset)
- Programs capped at 30 per channel
- Watch Next removes items at >90% progress
- Next episode found via parent hierarchy
- Deep link routes to player when category set to IMMEDIATE_PLAY
- Deep link routes to detail when set to DETAIL (default)
- Audio without album context always plays immediately
- Unauthenticated deep link redirects to login, then resumes
- Worker skips refresh when no saved auth token
- Logout deletes all channels and cancels worker

## Logout Cleanup

On logout, the app **must** perform a complete cleanup of all TV channel state before navigating to the login screen:

1. **Delete all channels** — Remove the default channel and all category channels from the system TvProvider (`TvChannelRepository.deleteAllChannels()`)
2. **Remove all Watch Next entries** — Clear all Catalogizer entries from the system Watch Next row (`WatchNextManager.removeAll()`)
3. **Cancel WorkManager sync** — Cancel the periodic `TvChannelSyncWorker` so no background sync runs while logged out
4. **Clear persisted channel IDs** — Remove all `channel_id_*` keys from DataStore
5. **Clear per-category launch settings** — Reset channel tap behavior preferences
6. **Navigate to LoginScreen** — Pop entire back stack and navigate to login

This is triggered from `SettingsScreen` logout action and `AuthViewModel.logout()`. The cleanup must complete before navigation to prevent stale channels appearing on the home screen for a different (or no) user.

On next login, the full channel setup runs fresh: default channel created, category channels discovered, WorkManager re-enqueued.

## Out of Scope

- No changes to catalog-api (backend) — all existing endpoints suffice
- No new Room tables — channel IDs in DataStore, programs in system TvProvider
- Existing `CatalogizerTvProviderImpl` ContentProvider is not used — Android TV channels use `TvContractCompat`, not a custom provider. Existing provider remains for potential future use.
