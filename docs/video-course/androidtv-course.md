# Android TV -- Kotlin/Compose Course

**Component**: catalogizer-androidtv
**Language**: Kotlin / Jetpack Compose / Leanback
**Total Duration**: 3.5 hours (5 modules)
**Level**: Intermediate

---

## Course Overview

This course covers the complete architecture of the Catalogizer Android TV application. You will learn TV-optimized Compose UI with focus management, D-PAD and remote control navigation, Home Screen Channel integration with Watch Next and deep linking, VLC-based media playback with transport controls, and background synchronization with WorkManager. The TV app targets Android 9+ devices including the Xiaomi Mi Box 4.

---

### Module 1: Leanback Architecture

**Duration**: 50 minutes
**Prerequisites**: Kotlin fundamentals, basic Android development, Jetpack Compose basics

#### Learning Objectives
- Understand the TV-optimized MVVM architecture and how it differs from the phone app
- Navigate the package structure: `ui/`, `data/`, `player/`, `utils/`
- Explain how `DependencyContainer` provides manual dependency injection for the TV app
- Trace the application lifecycle from `CatalogizerTVApplication` through screen rendering

#### Topics Covered
1. **Application entry point (`CatalogizerTVApplication.kt`)**
   - Custom `Application` class initializing `DependencyContainer` with TV-specific configuration
   - TV provider registration for Home Screen Channel support
   - Network configuration: ADB reverse proxy (`tcp:8080`) connecting to catalog-api on development host
2. **Dependency container (`DependencyContainer.kt`)**
   - Manual DI providing: `AuthRepository`, `MediaRepository`, `SettingsRepository`
   - Retrofit client with base URL configuration for server discovery
   - TV-specific singletons: `TvChannelRepository`, `WatchNextManager`
3. **ViewModel layer (`ui/viewmodel/`)**
   - `AuthViewModel`: login state with D-PAD-friendly credential entry
   - `HomeViewModel`: media rows organized by category for the TV browse layout
   - `MainViewModel`: global navigation state, connection status
   - `SettingsViewModel`: server URL, channel preferences, playback settings
   - All ViewModels expose `StateFlow` collected by Compose TV UI
4. **Screen structure (`ui/screens/`)**
   - `home/`: horizontal rows of media cards, category headers, hero banner
   - `login/`: large-text credential fields optimized for 10-foot UI viewing distance
   - `media/`: entity detail with metadata, related items, play button
   - `player/`: full-screen video playback with overlay transport controls
   - `search/`: voice and text search with D-PAD keyboard navigation
   - `settings/`: server configuration, channel management, playback preferences
5. **TV design principles**
   - 10-foot UI: larger text, higher contrast, simplified layouts
   - Overscan-safe margins: content stays within safe area on all TV displays
   - Focus-driven interaction model: no touch, no mouse, only D-PAD directional input
   - Reduced information density compared to phone: fewer items per row, larger cards

#### Hands-On Exercise
Deploy the app to an Android TV device or emulator via ADB. Set up ADB reverse proxy (`adb reverse tcp:8080 tcp:8080`) so the TV can reach catalog-api on the development machine. Navigate the Home screen and observe how focus moves between rows and cards. Inspect the ViewModel StateFlow values in the debugger while navigating.

#### Key Takeaways
- The TV app shares the MVVM pattern and repository layer with the phone app but has entirely different UI screens
- Manual DI via `DependencyContainer` avoids Hilt annotation processing issues with the TV SDK
- ADB reverse proxy is essential for development: it lets the TV device reach the API on the development host
- TV UI is designed for 10-foot viewing distance with focus-driven navigation instead of touch

---

### Module 2: D-PAD Navigation

**Duration**: 40 minutes
**Prerequisites**: Module 1

#### Learning Objectives
- Implement focus management for TV Compose components with directional navigation
- Handle keyboard and remote control input events for all D-PAD directions
- Build focusable card components that visually indicate focus state
- Solve common focus traps and navigation dead-ends in complex layouts

#### Topics Covered
1. **Focus system fundamentals**
   - Compose `Modifier.focusable()` and `FocusRequester` for programmatic focus control
   - Focus traversal order: left/right within a row, up/down between rows
   - Initial focus assignment: first item in the first row receives focus on screen entry
   - Focus restoration: returning to a screen places focus on the previously focused item
2. **D-PAD input handling**
   - `Modifier.onKeyEvent()` intercepting DPAD_UP, DPAD_DOWN, DPAD_LEFT, DPAD_RIGHT, DPAD_CENTER
   - DPAD_CENTER triggers selection (equivalent to click on touch devices)
   - Long-press handling for context menus
   - Back button navigation with stack-based screen history
3. **Remote control integration**
   - Standard Android TV remote: D-PAD, Select, Back, Home, Play/Pause, Fast Forward, Rewind
   - Media key handling: KEYCODE_MEDIA_PLAY_PAUSE, KEYCODE_MEDIA_NEXT, KEYCODE_MEDIA_PREVIOUS
   - Voice search button triggering the search screen
4. **Focus-aware card components (`ui/components/`)**
   - Card scale animation on focus: focused card grows slightly to indicate selection
   - Border highlight color change on focus matching the Material theme
   - Card content: cover art, title, progress overlay, type badge
   - Row scrolling: focused card stays centered, adjacent cards scroll into view
5. **Credential entry with D-PAD**
   - Critical sequence: `dpad_center` to focus the text field BEFORE `input text` for typing
   - `KEYCODE_TAB` to move between username and password fields
   - On-screen keyboard interaction via D-PAD
   - Login button focus after password entry
6. **Common pitfalls and solutions**
   - Focus traps: ensure every focusable element has valid neighbors in all four directions
   - Focus loss on data refresh: save and restore FocusRequester state across recompositions
   - Invisible focusable elements: off-screen items in lazy lists must not steal focus
   - Performance: avoid recomposition of entire rows when only focus state changes

#### Hands-On Exercise
Build a TV card component with focus-aware animations (scale up on focus, highlight border). Create a horizontal row of cards and verify D-PAD left/right navigation works correctly. Add a second row and verify D-PAD up/down moves focus between rows. Implement the credential entry sequence: navigate to the login screen, focus the username field with DPAD_CENTER, type text, TAB to password, type, and submit.

#### Key Takeaways
- D-PAD navigation requires explicit focus management that touch-based apps never need
- The credential entry sequence (`dpad_center` before `input text`, `KEYCODE_TAB` between fields) is critical for TV login flows
- Focus-aware animations provide essential visual feedback since there is no cursor or touch indicator
- Focus restoration must be handled explicitly to prevent disorienting focus jumps on screen return

---

### Module 3: Home Screen Channels

**Duration**: 45 minutes
**Prerequisites**: Module 1, Module 2, understanding of Android TV Home Screen

#### Learning Objectives
- Implement Home Screen Channels using `androidx.tvprovider` for content surfacing outside the app
- Build the Watch Next row integration for partially-watched content and auto-next-episode
- Configure deep linking from Home Screen programs to specific media entities within the app
- Set up background sync to keep channel content fresh

#### Topics Covered
1. **TV Channel Repository (`data/tv/TvChannelRepository.kt`)**
   - Default "Catalogizer Picks" channel auto-created on first app launch
   - Per-category dynamic channels: Movies, TV Shows, Music, Games, Books
   - Channel creation using `androidx.tvprovider` content provider
   - Program insertion with cover art, title, description, intent URI
   - `CatalogizerTvProvider.kt` and `CatalogizerTvProviderImpl.kt` implementing the provider interface
2. **Channel Program Mapper (`data/tv/ChannelProgramMapper.kt`)**
   - Converting media entities to TV provider `PreviewProgram` objects
   - Media type mapping: movie -> TYPE_MOVIE, tv_episode -> TYPE_TV_EPISODE, song -> TYPE_TRACK, software -> TYPE_CLIP (TYPE_APP does not exist in tvprovider 1.0.0)
   - Cover art URI construction from backend asset endpoints
   - Duration, rating, and genre metadata population
3. **Watch Next integration (`data/tv/WatchNextManager.kt`)**
   - Adding partially-watched items to the system Watch Next row
   - Progress tracking: items appear with a progress bar showing completion percentage
   - Auto-next-episode: when an episode finishes, the next episode appears in Watch Next
   - Removal of completed items (90%+ progress) from Watch Next
4. **Deep linking (`ui/ChannelDeepLinkActivity.kt`)**
   - URI scheme: `catalogizer://media/{id}?type={type}`
   - `ChannelDeepLinkActivity` resolving intent URIs to specific screens
   - Per-category launch behavior configurable in Settings: open detail page, start playback, or browse category
   - `LaunchAction.kt` defining the action enum: VIEW_DETAIL, PLAY, BROWSE_CATEGORY
5. **Channel content in Settings**
   - User-facing toggle to enable/disable per-category channels
   - Default channel selection
   - Deep link behavior configuration per category

#### Hands-On Exercise
Launch the app and verify the "Catalogizer Picks" channel appears on the Android TV Home Screen. Add media to the library and observe dynamic category channels populating. Play a video partially (50%), exit the app, and verify it appears in the Watch Next row with a progress bar. Click the Watch Next item and verify deep linking opens the correct media entity.

#### Key Takeaways
- Home Screen Channels surface Catalogizer content directly on the TV launcher without opening the app
- `TYPE_APP` does not exist in tvprovider 1.0.0; software media maps to `TYPE_CLIP` as a workaround
- Watch Next integration provides continuity: partially-watched items appear with progress and auto-advance to next episode
- Deep linking via `catalogizer://media/{id}?type={type}` connects Home Screen programs to specific in-app content

---

### Module 4: Media Playback

**Duration**: 40 minutes
**Prerequisites**: Module 1, Module 2

#### Learning Objectives
- Integrate VLC for media playback on Android TV with full transport control support
- Build episode navigation for sequential TV show viewing
- Handle playback lifecycle: foreground service, audio focus, HDMI-CEC integration
- Implement progress tracking synchronized with Watch Next and the backend API

#### Topics Covered
1. **VLC player (`player/VLCPlayer.kt`)**
   - LibVLC integration for broad format support (MKV, AVI, FLAC, DTS, Dolby)
   - Streaming from catalog-api media endpoints with authentication headers
   - Hardware-accelerated decoding with software fallback
   - Subtitle track selection and external subtitle file loading
2. **Player screen (`ui/screens/player/`)**
   - Full-screen video surface with overlay transport controls
   - Controls auto-hide after 5 seconds of inactivity, reappear on any D-PAD input
   - Progress bar with seek (left/right D-PAD for 10s skip, long-press for fast seek)
   - Media info overlay: title, episode number, duration, remaining time
3. **Episode navigation**
   - Previous/next episode buttons in transport controls
   - Auto-advance: playback of next episode starts automatically after current episode completes
   - Season boundary handling: prompt user when advancing from last episode of a season to next season
   - Episode list accessible via D-PAD up from transport controls
4. **Remote control media keys**
   - KEYCODE_MEDIA_PLAY_PAUSE: toggle play/pause
   - KEYCODE_MEDIA_FAST_FORWARD / KEYCODE_MEDIA_REWIND: 30s skip forward/backward
   - KEYCODE_MEDIA_NEXT / KEYCODE_MEDIA_PREVIOUS: next/previous episode
   - KEYCODE_MEDIA_STOP: stop playback and return to detail screen
5. **Progress synchronization**
   - Position reported to backend every 10 seconds during playback
   - Watch Next row updated with current progress on pause and stop
   - Cross-device resume: progress synced via API, resumable from phone or web
   - Completion threshold: 90% marks item as watched, updates Watch Next accordingly

#### Hands-On Exercise
Play a video file and exercise all transport controls via the remote: play/pause, seek, skip forward/backward. Navigate to the next episode using the media next key. Pause playback, exit the app, and verify the Watch Next row shows the correct progress. Resume from the Watch Next item and verify playback starts at the saved position.

#### Key Takeaways
- VLC provides broader format support than ExoPlayer, which is critical for TV users with diverse media libraries
- Transport controls must respond to both D-PAD and dedicated media keys on the remote
- Episode auto-advance with season boundary handling provides a seamless binge-watching experience
- Progress tracking is synchronized to Watch Next, the backend API, and other devices for cross-platform continuity

---

### Module 5: Background Sync

**Duration**: 25 minutes
**Prerequisites**: Modules 1-4

#### Learning Objectives
- Configure WorkManager for periodic Home Screen Channel updates
- Implement the sync trigger chain: app launch, SyncService events, and periodic worker
- Handle cleanup on logout: channel deletion, Watch Next removal, worker cancellation
- Monitor sync health and diagnose stale channel content

#### Topics Covered
1. **TvChannelSyncWorker (`data/tv/TvChannelSyncWorker.kt`)**
   - `PeriodicWorkRequest` running every 6 hours to refresh channel content
   - Network connectivity constraint: sync deferred when offline
   - Fetches latest media entities from API and updates channel programs
   - Removes programs for deleted or unavailable media
2. **Sync triggers**
   - App launch: immediate sync on `CatalogizerTVApplication.onCreate()`
   - `SyncService.kt` events: sync triggered when new media is detected or scan completes
   - Manual refresh: user-initiated sync from Settings screen
   - Each trigger is idempotent: concurrent sync requests are coalesced
3. **Channel update strategy**
   - Differential update: compare current channel programs against latest API data
   - Add new programs for recently scanned media
   - Update existing programs with refreshed metadata (new cover art, updated progress)
   - Remove programs for media that no longer exists or is no longer accessible
4. **Logout cleanup**
   - Delete all Catalogizer channels from the TV provider
   - Remove all Watch Next entries created by the app
   - Cancel the periodic `TvChannelSyncWorker`
   - Clear DataStore keys (auth tokens, preferences)
   - Clear Room database cache
   - Full cleanup ensures no residual data remains after user signs out
5. **Diagnostics and monitoring**
   - WorkManager task status inspection via App Inspection in Android Studio
   - Last sync timestamp stored in DataStore, displayed in Settings
   - Sync failure logging with error categorization (network, auth expired, API error)
   - Retry logic: failed sync retries with exponential backoff up to 3 attempts

#### Hands-On Exercise
Force a channel sync from Settings and observe the channel content updating on the Home Screen. Inspect the WorkManager periodic task status using Android Studio App Inspection. Log out and verify that all channels, Watch Next entries, and cached data are removed. Log back in and verify channels are recreated on app launch.

#### Key Takeaways
- The 6-hour periodic sync keeps Home Screen Channels fresh without excessive battery or network usage
- Three sync triggers (app launch, SyncService events, periodic worker) ensure content stays current across usage patterns
- Logout cleanup is comprehensive: channels, Watch Next, workers, DataStore, and Room are all cleared
- Each sync trigger is idempotent, so overlapping triggers do not cause duplicate work or race conditions
