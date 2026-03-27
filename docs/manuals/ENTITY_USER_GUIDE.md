# Catalogizer -- Entity User Guide

## Table of Contents

1. [What Are Entities](#what-are-entities)
2. [Entity Types](#entity-types)
3. [Entity Hierarchy](#entity-hierarchy)
4. [Browsing Entities](#browsing-entities)
5. [Searching Entities](#searching-entities)
6. [Viewing Entity Details](#viewing-entity-details)
7. [Duplicate Detection](#duplicate-detection)
8. [Entity Metadata](#entity-metadata)
9. [Working with Collections](#working-with-collections)
10. [Troubleshooting](#troubleshooting)

---

## What Are Entities

Entities are the core organizational unit in Catalogizer. When you add storage roots and scan your media library, Catalogizer does not simply list raw files. Instead, it runs a post-scan aggregation pipeline that transforms scanned files into structured **media entities** -- richly categorized items with metadata, hierarchical relationships, and linked source files.

A single entity represents a logical media item: a movie, a TV show, an album, a game, or any other recognized media type. Each entity can be linked to one or more physical files on disk. For example, a movie entity might be linked to a 1080p MKV file on your NAS and a 4K remux on a local drive.

The aggregation pipeline performs the following steps automatically after every scan:

1. **Title parsing** -- Extracts title, year, season/episode numbers, and other metadata from file and directory names using pattern-matching rules.
2. **Entity creation** -- Creates or updates a `media_item` record in the database for each recognized title.
3. **File linking** -- Associates the scanned file with its corresponding entity via the `media_files` junction table.
4. **Hierarchy building** -- For multi-level media (TV shows, music), constructs parent-child relationships automatically.
5. **Duplicate detection** -- Identifies files that map to the same logical entity (same title, type, and year).

---

## Entity Types

Catalogizer recognizes 11 distinct media types. Each type is seeded in the `media_types` database table and determines how the entity is displayed, enriched with metadata, and organized in the hierarchy.

| Type | Description | Example |
|------|-------------|---------|
| `movie` | Feature films, documentaries, short films | The Matrix (1999) |
| `tv_show` | Television series (top-level container) | Breaking Bad |
| `tv_season` | A season within a TV show | Breaking Bad - Season 3 |
| `tv_episode` | An individual episode | S03E07 - One Minute |
| `music_artist` | A musical performer or band (top-level container) | Pink Floyd |
| `music_album` | An album by an artist | The Dark Side of the Moon |
| `song` | An individual track | Time |
| `game` | Video games for any platform | The Witcher 3 |
| `software` | Applications, utilities, development tools | Visual Studio Code |
| `book` | E-books, audiobooks, PDFs | Dune by Frank Herbert |
| `comic` | Comic books, graphic novels, manga | Watchmen |

The type is determined automatically by the title parser during aggregation. The parser examines file paths, directory structure, and naming conventions to assign the correct type. For example, a file at `/media/TV Shows/Breaking Bad/Season 3/S03E07.mkv` is recognized as a `tv_episode` based on the directory hierarchy and S##E## naming pattern.

---

## Entity Hierarchy

Some media types are organized in parent-child hierarchies. The hierarchy is represented by a `parent_id` self-referencing column in the `media_items` table.

### Television Hierarchy

```
tv_show (Breaking Bad)
  |
  +-- tv_season (Season 1)
  |     +-- tv_episode (S01E01 - Pilot)
  |     +-- tv_episode (S01E02 - Cat's in the Bag...)
  |     +-- ...
  |
  +-- tv_season (Season 2)
  |     +-- tv_episode (S02E01 - Seven Thirty-Seven)
  |     +-- ...
  |
  +-- tv_season (Season 3)
        +-- ...
```

When you browse a TV show entity, you see all its seasons listed as children. Selecting a season shows its episodes. Each episode links to the actual media file(s) on disk.

### Music Hierarchy

```
music_artist (Pink Floyd)
  |
  +-- music_album (The Dark Side of the Moon)
  |     +-- song (Speak to Me)
  |     +-- song (Breathe)
  |     +-- song (Time)
  |     +-- ...
  |
  +-- music_album (Wish You Were Here)
        +-- song (Shine On You Crazy Diamond, Parts I-V)
        +-- ...
```

Music follows the same pattern: artist at the top, albums as children, songs as grandchildren. Each song entity links to the audio file.

### Flat Types

Movies, games, software, books, and comics are flat -- they have no parent-child hierarchy. Each entity stands alone and links directly to its file(s).

---

## Browsing Entities

The entity browser is the primary interface for exploring your media library. Access it from the web application at `/browse` or via the "Browse" item in the main navigation.

### Entity Browser Page

The browser displays a grid or list of entity cards. Each card shows:

- **Poster or thumbnail** (loaded from metadata providers or extracted from the file)
- **Title** and **year**
- **Media type** badge (color-coded by type)
- **File count** (number of linked source files)
- **Rating** (if available from metadata providers)

### Filtering by Type

Use the type filter bar at the top of the browser to show only specific media types. Click a type badge to toggle its visibility. Multiple types can be active simultaneously. The filter state is preserved in the URL query string, so you can bookmark filtered views.

### Sorting

Entities can be sorted by:

- **Title** (alphabetical, A-Z or Z-A)
- **Year** (newest first or oldest first)
- **Date added** (most recently scanned first)
- **Rating** (highest rated first)
- **File count** (most files first -- useful for finding duplicates)

### Grid vs. List View

Toggle between grid view (poster cards) and list view (compact rows with more metadata visible) using the view toggle in the browser toolbar. Your preference is saved locally and restored on future visits.

### Pagination and Virtual Scrolling

For large libraries, the entity browser uses virtual scrolling to maintain smooth performance. Only the visible entities are rendered in the browser DOM, regardless of total library size. Scroll freely through tens of thousands of entities without performance degradation.

---

## Searching Entities

### Quick Search

The search bar at the top of the entity browser accepts free-text queries. Type a title or partial title and results appear as you type (debounced at 300ms to avoid excessive API calls).

### Search Parameters

For more precise searches, use the advanced search panel (click the filter icon next to the search bar). Available search parameters:

- **Title** -- Full or partial title match (case-insensitive)
- **Type** -- Restrict to one or more media types
- **Year** -- Exact year or year range (e.g., 2020-2025)
- **Rating** -- Minimum rating threshold
- **Has metadata** -- Filter to entities with or without external metadata

### Search API

Programmatic search is available via the REST API:

```
GET /api/v1/entities?q=matrix&type=movie&year=1999
GET /api/v1/entities?type=tv_show&sort=rating&order=desc&limit=20
```

---

## Viewing Entity Details

Click any entity card in the browser to open the entity detail page at `/entity/:id`. The detail page displays:

### Header Section

- Full title, original title (if different), and year
- Media type badge
- Poster image (full resolution)
- Rating from metadata providers (TMDB, IMDB scores when available)
- Genre tags

### File Section

A list of all physical files linked to this entity. Each file entry shows:

- File path and storage root
- File size and format (MKV, MP4, FLAC, etc.)
- Resolution and codec information (for video)
- Bitrate and sample rate (for audio)
- Last scanned timestamp

Click a file entry to open the media player (if supported) or navigate to the file's location in the storage browser.

### Hierarchy Section (TV and Music Only)

For hierarchical entities, this section shows the parent-child tree. On a TV show page, you see all seasons and their episode counts. On a season page, you see the episode list. On an artist page, you see all albums. Navigation between levels is provided via breadcrumbs and clickable child entries.

### Metadata Section

External metadata retrieved from providers:

- **Plot summary / description**
- **Cast and crew** (movies and TV)
- **Track listing** (music albums)
- **Publisher and platform** (games and software)
- **Author and ISBN** (books)
- **Genre, language, and country of origin**

### User Metadata

Your personal notes, tags, and rating for this entity. User metadata is stored separately from external metadata and is never overwritten by provider updates.

---

## Duplicate Detection

The aggregation pipeline automatically detects duplicates -- files that map to the same logical entity. Duplicates are identified when two or more files share the same parsed title, media type, and year.

### Viewing Duplicates

Entities with multiple linked files display a "Duplicates" badge on their card in the browser. On the entity detail page, the file section lists all linked files, making it easy to compare resolutions, codecs, and file sizes.

### Common Duplicate Scenarios

- Same movie in multiple resolutions (720p, 1080p, 4K)
- Same movie from different storage roots (local disk and NAS)
- Same album in different formats (FLAC and MP3)
- Re-encoded or remuxed copies of the same content

Duplicate detection does not automatically delete files. It surfaces the information so you can make informed decisions about which copies to keep.

---

## Entity Metadata

Catalogizer enriches entities with metadata from external providers. The enrichment process runs automatically after aggregation and can also be triggered manually from the entity detail page.

### Metadata Providers

| Provider | Media Types | Data Provided |
|----------|-------------|---------------|
| TMDB | Movies, TV shows | Poster, backdrop, plot, cast, crew, ratings, genres |
| OMDB | Movies, TV shows | IMDB rating, Rotten Tomatoes score, awards, box office |
| OpenLibrary | Books | Cover image, author, publisher, ISBN, page count, subjects |
| MusicBrainz | Music | Artist info, album track listing, release date, label |

Metadata providers are optional. If API keys are not configured or a provider is unavailable, the entity is still created with the information parsed from the file name. Missing metadata can be filled in later when the provider becomes available.

### Manual Metadata Refresh

On the entity detail page, click the refresh icon in the metadata section to re-query all providers for updated information. This is useful when:

- A newly released movie or album has metadata added to providers after your initial scan
- Provider data has been corrected (wrong poster, incorrect cast listing)
- You changed the entity title and want to re-match against providers

### User Metadata

In addition to provider metadata, you can add your own:

- **Personal rating** (1-5 stars)
- **Tags** (free-form labels for personal organization)
- **Notes** (free-text field for any purpose)

User metadata is stored in the `user_metadata` table, linked to both the entity and your user account. It is never overwritten by provider updates and is included in backup/export operations.

---

## Working with Collections

Entities can be organized into collections for custom grouping beyond the automatic type-based organization. See the main User Guide for full collection management documentation. Key entity-related collection operations:

- **Add to collection**: From the entity detail page or via the context menu on entity cards in the browser.
- **Remove from collection**: From the collection detail page or the entity detail page.
- **Smart collections**: Rule-based collections that automatically include entities matching specified criteria (e.g., "all movies from 2024 with rating above 7").

---

## Troubleshooting

### Entity Not Created After Scan

If a scanned file does not result in an entity, the title parser could not determine the media type from the file path. Check that your file naming follows common conventions:

- Movies: `Movie Title (2024).mkv` or `Movie.Title.2024.1080p.BluRay.mkv`
- TV episodes: `Show Name/Season 01/S01E01 - Episode Title.mkv`
- Music: `Artist/Album (Year)/01 - Track Title.flac`

Files with ambiguous names may require manual entity creation from the admin panel.

### Wrong Entity Type Assigned

If a file is categorized as the wrong type (e.g., a documentary detected as a movie), you can correct this from the entity detail page by changing the type in the metadata section. The correction is persistent and survives future re-scans.

### Missing Metadata

Ensure API keys are configured for the relevant metadata providers in `config.json` or via environment variables (`TMDB_API_KEY`, `OMDB_API_KEY`). OpenLibrary and MusicBrainz do not require API keys but may rate-limit requests. Try the manual metadata refresh from the entity detail page.

### Duplicate Entities for the Same Content

If the same content appears as two separate entities (rather than one entity with multiple files), the title parser extracted different titles from the file names. Standardize your file naming to ensure consistency. You can merge duplicate entities from the admin panel.
