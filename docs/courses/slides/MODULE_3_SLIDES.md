# Module 3: Media Management - Slide Outlines

---

## Slide 3.0.1: Title Slide

**Title**: Media Management

**Subtitle**: Organizing, Playing, and Converting Your Media Collection

**Speaker Notes**: This module covers the organizational and playback features that make Catalogizer more than just a file browser. We cover favorites, collections, playlists, subtitles, format conversion, and the built-in media player.

---

## Slide 3.1.1: Favorites

**Title**: Quick Access to What Matters

**Bullet Points**:
- Toggle the heart icon on any MediaCard to add/remove favorites
- Dedicated Favorites page with Grid and List views, full filter and sort
- `useFavorites` hook manages state on the frontend
- Backend: `favorites_service.go` + `favorites_repository.go`

**Visual**: Screenshot of a MediaCard with the heart icon highlighted, and the Favorites page

**Speaker Notes**: Favorites are the simplest organizational tool. One click to save, one click to remove. They sync across all platforms, so a favorite added on your phone appears on the web and desktop.

---

## Slide 3.1.2: Favorites Import and Export

**Title**: Portable Favorites

**Bullet Points**:
- Export to JSON format (full metadata included)
- Export to CSV format (spreadsheet compatible)
- Import from JSON or CSV files
- Matching uses metadata, not file paths -- works across different sources
- Use cases: backup, sharing recommendations, migrating between instances

**Speaker Notes**: The metadata-based matching is important. If you export favorites from one Catalogizer instance and import them on another with different file paths, the system still finds the right items by matching on title, type, and provider IDs.

---

## Slide 3.2.1: Collections

**Title**: Three Types of Collections

**Bullet Points**:
- **Manual**: Hand-picked items; add via drag-and-drop or bulk selection
- **Smart**: Rule-based auto-population (e.g., "all movies from 2024", "all 4K videos")
- **Dynamic**: Adaptive criteria that evolve with your library
- Access permissions: Public, Private, or shared with specific users
- Backend: `catalog.go` service; Frontend: `useCollections` hook

**Visual**: Diagram showing the three collection types with example rules

**Speaker Notes**: Manual collections are like playlists without playback order. Smart collections are the power feature -- define rules once and Catalogizer maintains them forever. Every new movie that matches your "4K Movies" rule is automatically added.

---

## Slide 3.2.2: Bulk Operations

**Title**: Managing Collections at Scale

**Bullet Points**:
- Select multiple items with checkboxes in the Media Browser
- Add all selected items to a collection in one action
- Move items between collections
- Drag-and-drop single items for quick organization
- Edit collection metadata: name, description, permissions

**Speaker Notes**: Bulk operations are essential when you have thousands of media items. Rather than adding items one at a time, select a group and assign them to a collection in a single action.

---

## Slide 3.3.1: Playlists

**Title**: Ordered Playback Sequences

**Bullet Points**:
- Playlists emphasize order and sequential playback; collections organize thematically
- Drag-and-drop reordering via `usePlaylistReorder` hook
- Auto-advancement: player moves to next item when current finishes
- Items can belong to multiple playlists simultaneously
- Backend: `playlist_service.go`; Frontend: `usePlaylists.tsx`

**Visual**: Screenshot of a playlist with numbered items and drag handles

**Speaker Notes**: Think of playlists as your movie night queue or your weekend music mix. The order matters. Create a playlist, arrange the items how you want, and hit play. The player handles the rest.

---

## Slide 3.4.1: Subtitle Management

**Title**: Centralized Subtitle Control

**Bullet Points**:
- Supported formats: SRT, ASS, SSA, VTT
- Automatic matching: subtitle files with same base name as video
- Manual association for unmatched files
- Upload new subtitle files with language specification
- Backend: `subtitle_service.go` + `subtitle_handler.go`
- Frontend: `SubtitleManager.tsx`

**Visual**: Screenshot of the Subtitle Manager showing matched and unmatched subtitles

**Speaker Notes**: Subtitle management is often overlooked in media managers. Catalogizer treats subtitles as first-class citizens. Auto-matching handles the common case. For edge cases, manual association is straightforward.

---

## Slide 3.4.2: Subtitles During Playback

**Title**: Subtitle Track Selection in the Player

**Bullet Points**:
- Track selector in the media player for switching subtitles
- Support for multiple language tracks per video
- Turn subtitles on or off during playback
- Naming convention tip: `movie.en.srt`, `movie.fr.srt` for best auto-matching

**Speaker Notes**: During playback, the subtitle selector appears in the player controls. If a video has English, French, and Spanish subtitle files associated, all three are available to switch between without leaving the player.

---

## Slide 3.5.1: Format Conversion

**Title**: Convert Media Between Formats

**Bullet Points**:
- **Video**: Convert between containers and codecs
- **Audio**: MP3, FLAC, WAV, AAC, and more
- **Documents**: Text format conversions
- **PDF**: To images (thumbnails), to text (search indexing), to HTML (web display)
- Batch conversion queue with real-time progress via WebSocket
- Backend: `conversion_service.go` + `conversion_repository.go`

**Visual**: Screenshot of the Conversion Tools page with format selection

**Speaker Notes**: The conversion system runs in the background. Queue up multiple files, set their target formats, and continue using Catalogizer while they process. Converted files automatically appear in the catalog.

---

## Slide 3.5.2: PDF Conversion Workflow

**Title**: PDF as a First-Class Media Type

**Bullet Points**:
- PDF to images: Generate thumbnails and preview images
- PDF to text: Extract content for full-text search indexing
- PDF to HTML: Render for web display within the catalog
- Conversion progress reported in real-time
- Output files indexed alongside other media

**Speaker Notes**: PDF to text conversion is particularly valuable for searchability. Once converted, the text content of PDF documents becomes searchable through Catalogizer's search system, just like movie titles or album names.

---

## Slide 3.6.1: Built-in Media Player

**Title**: Play Anything, Anywhere

**Bullet Points**:
- `MediaPlayer.tsx` component with `usePlayerState` hook
- Controls: play, pause, seek, volume, playback speed, fullscreen
- Three backend services: `media_player_service.go` (general), `video_player_service.go`, `music_player_service.go`
- Transparent streaming from all storage protocols
- Subtitle track selection during playback

**Visual**: Screenshot of the video player with controls visible

**Speaker Notes**: The media player streams content directly from the storage source, regardless of protocol. An SMB file plays just as smoothly as a local file. The backend handles all protocol negotiation transparently.

---

## Slide 3.6.2: Playback Position and Lyrics

**Title**: Resume Where You Left Off

**Bullet Points**:
- `playback_position_service.go` saves position on pause or close
- Resume from saved position when returning to the same item
- Position syncs across devices: start on desktop, continue on phone
- `lyrics_service.go` displays synchronized lyrics during music playback
- `cover_art_service.go` provides album artwork in the player

**Speaker Notes**: Cross-device position sync is seamless. Watch 45 minutes of a movie on your laptop, close it, and open the Android app -- it offers to resume from exactly where you stopped. Lyrics display adds a karaoke-like experience for music playback.

---

## Slide 3.6.3: Deep Linking and Playlist Integration

**Title**: Share Moments and Play Continuously

**Bullet Points**:
- `deep_linking_service.go` generates links to specific playback positions
- Share a link and the recipient jumps directly to the relevant moment
- Playlist auto-advancement: next item plays when current finishes
- Combined with position tracking: pause a playlist and resume days later

**Speaker Notes**: Deep linking is perfect for sharing a specific scene or moment with someone. The link includes the media item and the exact timestamp. Playlist integration means the player is not just for single items -- it handles queued playback with seamless transitions.

---

## Slide 3.6.4: Module 3 Summary

**Title**: What We Covered

**Bullet Points**:
- Favorites for quick access with JSON/CSV import and export
- Collections: Manual, Smart (rule-based), and Dynamic with access permissions
- Playlists for ordered sequential playback with drag-and-drop reordering
- Subtitle management with auto-matching and multi-language support
- Format conversion with batch queue and real-time progress
- Built-in media player with position tracking, lyrics, deep linking, and cross-device sync

**Next Steps**: Module 4 -- Multi-Platform Experience (Android, TV, Desktop, API Client)

**Speaker Notes**: These features transform Catalogizer from a simple file browser into a full media management platform. The combination of Smart collections, playback position sync, and deep linking creates a rich user experience across all your devices.
