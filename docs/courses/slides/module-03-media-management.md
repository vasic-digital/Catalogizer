# Module 3: Media Management - Slide Deck Outline

**Total Slides**: 14
**Estimated Duration**: 75 minutes

---

## Slide 1: Title Slide (2 min)

**Title**: Advanced Media Features

- Scanning, browsing, search, favorites, collections, playlists, subtitles, media player
- Prerequisites: Module 2 completed with connected storage sources
- By the end: organize media with collections and play media from the browser

---

## Slide 2: Favorites (5 min)

**Title**: Marking and Managing Favorites

- Add/remove items from Favorites via the heart icon or Favorites page
- useFavorites hook manages client-side state
- Export favorites to JSON and CSV formats with full metadata
- Import favorites from previously exported files
- Demo: favorite a movie, export the list, clear, and re-import

---

## Slide 3: Favorites Statistics (4 min)

**Title**: Insights Into Your Favorites

- Statistics tab shows total count, media type breakdown
- Most Common Type, Recent Activity, Storage Impact cards
- Recently Added tab for chronological view
- Bulk Actions for batch operations on favorites
- Exercise reference: Exercise 3.1 -- organize favorites by type

---

## Slide 4: Collections Overview (5 min)

**Title**: Creating and Managing Collections

- Three collection types: Manual, Smart, Dynamic
- Manual: drag and drop media items into the collection
- Smart: automatic filter-based population rules
- Dynamic: updated in real time based on conditions
- Access permissions: Public, Private, Friends Only
- Demo: create a Manual and a Smart collection

---

## Slide 5: Collection Operations (5 min)

**Title**: Working With Collections

- Bulk selection for batch add/remove operations
- Edit collection metadata: name, description, cover image
- Delete collections with confirmation dialog
- Collection detail page shows items with sorting and filtering
- Exercise reference: Exercise 3.2 -- create a Smart collection with filters

---

## Slide 6: Playlists (5 min)

**Title**: Creating and Managing Playlists

- Create playlists through the Playlists page
- Add media items from browse or search results
- Reorder items using drag-and-drop (usePlaylistReorder hook)
- Backend: playlist_service.go handles CRUD and ordering
- Demo: create a playlist and reorder tracks

---

## Slide 7: Playlist Playback (4 min)

**Title**: Playing Playlists Sequentially

- Play entire playlists through the built-in media player
- Sequential playback with auto-advance to next item
- Shuffle and repeat modes available
- Current track indicator in the playlist view
- Exercise reference: Exercise 3.3 -- build and play a movie marathon playlist

---

## Slide 8: Subtitle Management (5 min)

**Title**: Managing Subtitles Across Your Library

- Access the Subtitle Manager page for centralized operations
- Upload and associate subtitle files (.srt, .ass, .vtt) with videos
- subtitle_service.go handles automatic subtitle matching
- Switch between multiple subtitle tracks during playback
- Demo: upload a subtitle file and attach it to a movie

---

## Slide 9: Format Conversion (5 min)

**Title**: Converting Media Between Formats

- Conversion Tools page for media format conversion
- PDF to images, text, or HTML conversion
- Image, video, audio format conversions
- Monitor conversion progress and manage the queue
- Exercise reference: Exercise 3.4 -- convert a PDF to images

---

## Slide 10: Media Player Basics (5 min)

**Title**: Built-In Media Player

- MediaPlayer component for video and audio playback
- usePlayerState hook: play, pause, seek, volume control
- Fullscreen mode with keyboard shortcuts
- Stream media from remote protocol sources via handlers
- Demo: play a video file from an SMB share

---

## Slide 11: Advanced Player Features (5 min)

**Title**: Playback Position, Lyrics, and Deep Links

- Resume playback from saved positions (playback_position_service.go)
- Lyrics display during music playback (lyrics_service.go)
- Video player and music player specialized services
- Share specific moments using deep linking (deep_linking_service.go)
- Demo: pause a video, close the tab, and resume from the same position

---

## Slide 12: Media Detection Pipeline (5 min)

**Title**: How Media Is Identified and Categorized

- UniversalScanner crawls storage sources
- Title parser uses regex patterns for movie, TV, music, game, software
- AggregationService creates MediaItem entries and builds hierarchy
- Duplicate detection: same title + type + year across sources
- 11 media types seeded in the media_types table

---

## Slide 13: Metadata Enrichment (5 min)

**Title**: External Metadata Providers

- TMDB and OMDB for movie and TV metadata
- MusicBrainz for music albums and artists
- OpenLibrary for books
- Graceful degradation: missing API keys do not block the pipeline
- Exercise reference: Exercise 3.5 -- verify metadata enrichment for scanned media

---

## Slide 14: Module Summary and Next Steps (4 min)

**Title**: What We Covered

- Favorites with export/import and statistics
- Collections: Manual, Smart, Dynamic with permissions
- Playlists with drag-and-drop reordering and sequential playback
- Subtitle management and format conversion
- Built-in media player with resume, lyrics, and deep linking
- Next module: Multi-Platform Experience (Android, Desktop, API client)
