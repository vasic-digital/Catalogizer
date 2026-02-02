# Module 3: Media Management - Video Scripts

---

## Lesson 3.1: Favorites & Bookmarks

**Duration**: 10 minutes

### Narration

Favorites are the quickest way to mark media items you access frequently or want to come back to later. The Favorites page, implemented in Favorites.tsx, provides a dedicated view for all your bookmarked items.

Adding an item to favorites is straightforward. From the Media Browser, each MediaCard has a favorite toggle. Click the heart icon and the item is instantly added. The useFavorites hook manages the favorite state on the frontend, communicating with the backend API to persist your selections.

The Favorites page itself shows all your favorited items in the same Grid and List views available in the Media Browser. You can filter and sort within your favorites just like you would in the main catalog.

One particularly useful feature is favorites export and import. You can export your entire favorites list to JSON or CSV format, including full metadata for each item. This is great for backup, for sharing recommendations with friends, or for migrating between Catalogizer instances.

Importing works the same way in reverse. Upload a JSON or CSV file, and Catalogizer will match items in your catalog and re-add them to your favorites. The matching is done by media metadata, not just file paths, so it works even if files have moved between sources.

The favorites components live in the catalog-web/src/components/favorites/ directory, with the main hook in hooks/useFavorites.tsx. On the backend, the catalog service handles the favorites endpoints.

### On-Screen Actions

- [00:00] Navigate to the Media Browser
- [00:30] Click the favorite icon on a movie item -- show it being added
- [01:00] Click the favorite icon on a music item
- [01:30] Navigate to the Favorites page
- [02:00] Show both favorited items in the list
- [02:30] Switch between Grid and List views on the Favorites page
- [03:30] Apply a filter to show only video favorites
- [04:00] Demonstrate the Export button -- export to JSON
- [05:00] Show the exported JSON file contents in a text editor
- [05:30] Export to CSV format -- show the CSV file
- [06:00] Delete a favorite item
- [06:30] Use Import to re-add from the exported JSON
- [07:00] Verify the item reappears in favorites
- [07:30] Show the useFavorites.tsx hook code
- [08:30] Show the favorites components directory
- [09:00] Demonstrate removing all favorites and re-importing from CSV
- [09:30] Final overview of the Favorites page

### Key Points

- Add favorites from any MediaCard using the heart icon toggle
- Favorites page shows all bookmarked items with full filter and sort capabilities
- Export favorites to JSON or CSV with complete metadata
- Import favorites from JSON/CSV files -- matches by metadata, not just file paths
- useFavorites hook in catalog-web manages favorite state and API communication

### Tips

> **Tip**: Export your favorites periodically as a backup. If you ever need to reset or migrate your Catalogizer instance, you can re-import them without losing your curated list.

---

## Lesson 3.2: Collections & Organization

**Duration**: 12 minutes

### Narration

Collections let you group media items into organized sets. Think of them as folders with superpowers -- they support drag and drop, bulk operations, and even automatic population based on rules.

The Collections page, implemented in Collections.tsx, is your hub for creating and managing collections. You can create three types of collections.

Manual collections are the simplest. You create one, give it a name and description, and then manually add items to it by dragging and dropping from the media browser or using bulk selection.

Smart collections are filter-based. You define rules -- for example, "all movies from 2024" or "all music files larger than 10MB" -- and Catalogizer automatically populates the collection. When new media matching the rules is detected, it is automatically added.

Dynamic collections update based on more complex criteria and can adapt over time as your library changes.

Each collection has access permissions. You can set a collection to Public so anyone can view it, Private so only you can see it, or shared with specific users.

The useCollections hook manages collection state on the frontend. It provides methods for creating, updating, deleting collections, and adding or removing items. The backend catalog service (services/catalog.go) handles the persistence and rule evaluation.

Bulk operations are essential for managing large libraries. Select multiple items using checkboxes, then add them all to a collection in one action. You can also use this to move items between collections.

### On-Screen Actions

- [00:00] Navigate to Collections in the main navigation
- [00:30] Click "Create New Collection"
- [01:00] Create a Manual collection: enter name "Action Movies", set description
- [02:00] Navigate to Media Browser alongside the collection
- [02:30] Drag a movie item into the collection
- [03:00] Use bulk select to choose five items, then add to collection
- [03:30] Return to Collections page and open the new collection
- [04:00] Show items in the collection with count
- [04:30] Create a Smart collection: define rule "all items of type Movie"
- [05:30] Show the collection auto-populating
- [06:00] Add another rule: "year is 2024"
- [06:30] Show the filtered results
- [07:00] Create a Dynamic collection
- [07:30] Set access permissions: switch from Private to Public
- [08:00] Show the useCollections.ts hook in the code editor
- [08:30] Show the Collections.tsx page component
- [09:30] Show the catalog.go service methods for collections
- [10:00] Demonstrate editing a collection: rename, change description
- [10:30] Remove an item from a collection
- [11:00] Delete a collection
- [11:30] Final overview of the Collections page

### Key Points

- Three collection types: Manual (hand-picked), Smart (rule-based auto-population), Dynamic (adaptive)
- Drag and drop items into collections or use bulk selection
- Access permissions: Public, Private, or shared with specific users
- Smart collections auto-update when new matching media is detected
- useCollections hook manages frontend state; catalog.go handles backend persistence

### Tips

> **Tip**: Use Smart collections for ongoing organization. Set up rules like "all 4K movies" or "all music from a specific genre" and let Catalogizer maintain them automatically as your library grows.

---

## Lesson 3.3: Playlists

**Duration**: 10 minutes

### Narration

Playlists are ordered sequences of media items designed for sequential playback. While collections organize media thematically, playlists arrange items in a specific order for continuous enjoyment.

The Playlists page, implemented in Playlists.tsx, lets you create, edit, and manage playlists. Creating a playlist is similar to creating a collection, but with an emphasis on order.

The usePlaylists hook manages playlist data on the frontend, while the playlist_service.go on the backend handles storage and retrieval. What makes playlists special is the ordering -- you can drag items to rearrange them, and the usePlaylistReorder hook handles the drag-and-drop reordering logic.

When you play a playlist, items are played in sequence. The media player tracks your position in the playlist and automatically advances to the next item when the current one finishes. We will explore the media player in detail in Lesson 3.6.

You can add items to playlists from the Media Browser, from within a collection, or from search results. Multiple playlists can contain the same item, so you are not limited to having a movie or song in just one place.

The playlist components live in catalog-web/src/components/playlists/. The backend service in services/playlist_service.go provides the API for CRUD operations on playlists and their items.

### On-Screen Actions

- [00:00] Navigate to Playlists in the web UI
- [00:30] Click "Create Playlist" -- enter name "Weekend Movie Marathon"
- [01:00] Add three movies to the playlist from the Media Browser
- [02:00] Show the playlist with items in order
- [02:30] Drag an item to reorder -- move the third movie to first position
- [03:30] Show the reorder animation and updated list
- [04:00] Add two more items from search results
- [04:30] Start playing the playlist -- show the player advancing through items
- [05:30] Create a music playlist
- [06:00] Add several songs and arrange them
- [06:30] Play the music playlist
- [07:00] Show the usePlaylists.tsx hook in the code editor
- [07:30] Show the usePlaylistReorder.tsx hook
- [08:00] Show playlist_service.go on the backend
- [08:30] Demonstrate editing a playlist: rename, remove items
- [09:00] Delete a playlist
- [09:30] Final overview

### Key Points

- Playlists are ordered sequences designed for sequential playback
- Drag-and-drop reordering via usePlaylistReorder hook
- Auto-advancement: media player moves to next item when current finishes
- Items can belong to multiple playlists
- Backend: playlist_service.go; Frontend: usePlaylists.tsx + usePlaylistReorder.tsx

### Tips

> **Tip**: Create playlists for different moods or occasions. A "Movie Night" playlist with films in your preferred order makes it easy to just hit play and relax.

---

## Lesson 3.4: Subtitle Management

**Duration**: 10 minutes

### Narration

The Subtitle Manager is a dedicated page for handling subtitle files across your video library. Implemented in SubtitleManager.tsx, it provides centralized control over subtitle association and management.

Subtitles are critical for accessibility and multilingual viewing. The subtitle_service.go on the backend handles the logic for finding, matching, and associating subtitle files with their corresponding videos.

When Catalogizer scans your storage sources, it identifies subtitle files like SRT, ASS, SSA, and VTT formats alongside video files. The subtitle service attempts to automatically match them based on naming conventions -- a file named "movie.srt" next to "movie.mp4" is an obvious match.

In the Subtitle Manager, you can see all subtitle files in your library, which videos they are associated with, and their language. If automatic matching did not work, you can manually associate a subtitle file with any video.

You can also upload new subtitle files directly through the interface. Select a video, upload the subtitle file, specify the language, and the association is created.

The subtitles components are in catalog-web/src/components/subtitles/. During video playback, the media player offers a subtitle track selector, letting you switch between available subtitle files or turn them off entirely.

### On-Screen Actions

- [00:00] Navigate to the Subtitle Manager page
- [00:30] Show the list of detected subtitle files
- [01:00] Show a subtitle file automatically matched with its video
- [01:30] Show the language column and file format column
- [02:30] Find an unmatched subtitle file
- [03:00] Manually associate it with the correct video
- [03:30] Upload a new subtitle file for a video that has none
- [04:30] Specify the language during upload
- [05:00] Navigate to the video and play it with subtitles enabled
- [05:30] Show the subtitle track selector in the media player
- [06:00] Switch between subtitle tracks
- [06:30] Turn subtitles off
- [07:00] Show the SubtitleManager.tsx component code
- [07:30] Show subtitle_service.go on the backend
- [08:00] Show the subtitle matching logic
- [08:30] Show the subtitles components directory
- [09:00] Demonstrate subtitle support for different formats (SRT, VTT)
- [09:30] Final overview

### Key Points

- Subtitle Manager provides centralized management of all subtitle files
- Automatic matching based on naming conventions (same base name as video)
- Manual association for unmatched files
- Upload new subtitle files with language specification
- Media player includes subtitle track selector for switching during playback
- Backend: subtitle_service.go; Frontend: SubtitleManager.tsx and components/subtitles/

### Tips

> **Tip**: Name your subtitle files with the same base name as the video file and include the language code -- for example, "movie.en.srt" and "movie.fr.srt". This gives the automatic matching the best chance of working correctly.

---

## Lesson 3.5: Format Conversion & PDF Tools

**Duration**: 12 minutes

### Narration

Catalogizer includes built-in format conversion capabilities accessible through the Conversion Tools page, implemented in ConversionTools.tsx.

The conversion system supports transforming media between formats. For video, you can convert between containers and codecs. For audio, convert between formats like MP3, FLAC, WAV, and AAC. For documents, convert between text formats.

One particularly useful feature is the PDF conversion service. This converts PDF documents to images -- useful for thumbnails and previews -- to plain text for search indexing, or to HTML for web display. This service is built into the backend and integrated with the media detection pipeline.

The conversion components live in catalog-web/src/components/conversion/. The interface shows you a source file picker, format selection, and quality settings. Once you initiate a conversion, you can monitor its progress in real-time.

The conversion queue lets you batch multiple conversions. You can add several files, choose their target formats, and let them process in the background. Progress is reported via WebSocket so the UI stays up to date.

Let me demonstrate a few common conversion workflows. First, converting a video to a different format for compatibility. Second, converting a PDF to images for easy viewing. Third, converting a FLAC audio file to MP3 for mobile playback.

### On-Screen Actions

- [00:00] Navigate to Conversion Tools page
- [00:30] Show the interface: source selector, format picker, quality settings
- [01:30] Select a video file and choose a target format
- [02:30] Start the conversion and show progress indicator
- [03:30] While that runs, select a PDF file
- [04:00] Convert PDF to images -- show resulting image files
- [04:30] Convert PDF to text -- show extracted text
- [05:00] Convert PDF to HTML -- show rendered HTML output
- [05:30] Select a FLAC audio file and convert to MP3
- [06:30] Show the conversion queue with multiple items
- [07:00] Show progress updates arriving via WebSocket
- [07:30] Show a completed conversion and the output file
- [08:00] Open ConversionTools.tsx in the code editor
- [09:00] Show the conversion components directory
- [09:30] Show how converted files appear in the catalog
- [10:00] Demonstrate batch conversion: select multiple files
- [10:30] Configure batch settings and start
- [11:00] Final overview of conversion capabilities

### Key Points

- Conversion Tools page supports video, audio, document, and PDF format conversion
- PDF conversion: to images (thumbnails/previews), text (searchable), or HTML (web display)
- Batch conversion queue with real-time progress via WebSocket
- Converted files automatically appear in the catalog
- Frontend: ConversionTools.tsx and components/conversion/

### Tips

> **Tip**: Use PDF to text conversion to make document content searchable in Catalogizer's search system. The extracted text is indexed alongside other media metadata.

---

## Lesson 3.6: Built-in Media Player

**Duration**: 16 minutes

### Narration

Catalogizer includes a built-in media player that handles both video and audio playback directly in the browser. The player component is implemented in MediaPlayer.tsx and is supported by several specialized services.

The usePlayerState hook manages the player state: play, pause, seek, volume, playback speed, and fullscreen mode. It provides a clean API for the UI controls.

On the backend, there are three specialized player services. The media_player_service.go provides general playback capabilities. The video_player_service.go handles video-specific features. The music_player_service.go handles audio playback with music-specific features.

One of the most convenient features is playback position tracking. The playback_position_service.go saves your position whenever you pause or close the player. When you return to the same item, playback resumes from where you left off. This works across devices -- start watching on your desktop and continue on your phone.

For music, the lyrics service (lyrics_service.go) can display song lyrics synchronized with playback. If lyrics are available for a track, they appear alongside the player controls.

The media player handles streaming from remote sources transparently. Whether a file is on an SMB share, FTP server, or local disk, the media player handlers (media_player_handlers.go) coordinate with the filesystem layer to stream content to the player.

Playlist integration means the player advances through playlist items automatically. When one item finishes, the next begins. Combined with playback position tracking, you can pause a playlist on Tuesday and pick up exactly where you left off on Friday.

The cover art service (cover_art_service.go) provides album artwork and movie posters in the player interface. The deep linking service (deep_linking_service.go) enables sharing direct links to specific media items with playback position.

### On-Screen Actions

- [00:00] Open a video from the Media Browser -- player opens
- [00:30] Show player controls: play, pause, seek bar, volume, fullscreen
- [01:30] Demonstrate seeking to a specific position
- [02:00] Toggle fullscreen mode
- [02:30] Show subtitle track selection in the player
- [03:00] Pause the video -- note the position
- [03:30] Navigate away, then return to the same video -- show position restored
- [04:30] Open a music file in the player
- [05:00] Show the music player interface with album art (cover art service)
- [05:30] Show lyrics display during music playback
- [06:30] Open a playlist and start playback
- [07:00] Let the first item finish -- show auto-advance to next item
- [07:30] Skip to the next item manually
- [08:00] Show the player playing from an SMB source
- [08:30] Show the player playing from a local source
- [09:00] Open MediaPlayer.tsx in the code editor
- [09:30] Show usePlayerState.tsx hook
- [10:00] Show media_player_service.go on the backend
- [10:30] Show video_player_service.go
- [11:00] Show music_player_service.go
- [11:30] Show playback_position_service.go -- position tracking logic
- [12:00] Show lyrics_service.go
- [12:30] Show cover_art_service.go
- [13:00] Show deep_linking_service.go
- [13:30] Show media_player_handlers.go -- streaming coordination
- [14:00] Demonstrate deep linking: share a link to a specific playback position
- [14:30] Open the link in a new browser tab -- show playback starting at the shared position
- [15:00] Final overview of all player capabilities

### Key Points

- MediaPlayer.tsx with usePlayerState hook handles all playback UI and state
- Three backend services: media_player_service, video_player_service, music_player_service
- Playback position tracking persists across sessions and devices (playback_position_service.go)
- Lyrics display for music tracks (lyrics_service.go)
- Cover art and album artwork display (cover_art_service.go)
- Deep linking enables sharing specific playback positions (deep_linking_service.go)
- Transparent streaming from all storage protocols via media_player_handlers.go
- Playlist auto-advancement for continuous playback

### Tips

> **Tip**: Use deep linking to share specific moments in videos with others. The link includes the playback position, so the recipient jumps directly to the relevant point.

> **Tip**: Playback position syncs across devices. Start a movie on your desktop, pause it, and pick up on your Android phone right where you left off.
