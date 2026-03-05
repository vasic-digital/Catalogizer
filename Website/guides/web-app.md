---
title: Web App Guide
description: Using the Catalogizer web application to browse, search, and manage your media collection
---

# Web App Guide

The Catalogizer web application is a React-based interface for browsing, searching, and managing your media collection. It connects to the backend API over WebSocket for real-time updates and provides a responsive experience across desktop and mobile browsers.

---

## Dashboard

The dashboard is the first screen after login. It provides an at-a-glance summary of your catalog.

- **Total files**: Count of all detected media files across storage sources
- **Storage usage**: Aggregate size of cataloged media
- **Quality distribution**: Breakdown by resolution (720p, 1080p, 4K)
- **Growth chart**: How your library has changed over time
- **Recent additions**: Newly detected media from the latest scan
- **Source status**: Connection health of each storage source with last scan time

The dashboard refreshes automatically via WebSocket when scans complete or new media is detected.

---

## Navigation

The sidebar provides access to all major sections:

- **Dashboard** -- Overview and analytics
- **Browse** -- Media browser with filters
- **Search** -- Full-text search across all metadata
- **Collections** -- Organized groups of media
- **Playlists** -- Ordered sequences for playback
- **Favorites** -- Bookmarked items
- **Settings** -- Storage sources, user preferences, and system configuration

On mobile viewports, the sidebar collapses into a hamburger menu.

---

## Browsing Media

The media browser at `/browse` displays your catalog with multiple view modes.

### View Modes

- **Grid**: Thumbnail cards showing poster art, title, year, and quality badge
- **List**: Compact rows with title, type, size, quality, and source
- **Detail**: Expanded cards with full metadata and description

Toggle between views using the icons in the top-right corner of the browser.

### Filters

Narrow results using the filter panel:

- **Media type**: Movies, TV shows, music, games, software, books, comics
- **Quality**: Resolution, codec, or bitrate ranges
- **Source**: Filter by storage source
- **Year**: Release year range
- **Sort**: By title, date added, year, size, or quality

### Entity Hierarchy

Media is organized into entities with parent-child relationships:

- **TV Shows**: Show > Season > Episode
- **Music**: Artist > Album > Song

Click into a TV show to see its seasons, then click a season to browse episodes. The same hierarchical navigation applies to music artists and albums.

---

## Search

The search page provides full-text search across titles, descriptions, and metadata.

- Type a query in the search bar and press Enter or click Search
- Results appear instantly with highlighted matching terms
- Filter results by media type using the tabs below the search bar
- Search works across translated titles when localization is configured

Search queries match against:
- Original and translated titles
- Descriptions and synopses
- File names and paths
- External metadata (actor names, directors, album artists)

---

## Media Detail Page

Click any media item to open its detail page at `/entity/:id`.

- **Poster and backdrop**: Cover art fetched from external providers
- **Metadata**: Title, year, runtime, genres, rating, quality information
- **Description**: Synopsis from TMDB, IMDB, or other providers
- **Files**: List of associated files with size, codec, and source location
- **Play button**: Stream the media directly in the built-in player
- **Actions**: Add to favorites, add to collection, add to playlist

---

## Built-in Player

The media player streams video and audio directly in the browser.

- **Transport controls**: Play, pause, seek, volume, fullscreen
- **Subtitle tracks**: Select from detected subtitle files (SRT, ASS, VTT)
- **Resume playback**: Automatically resumes from your last position
- **Deep links**: Share a link to a specific timestamp with other users
- **Playlist mode**: Auto-advances to the next item in a playlist

---

## Collections

Collections organize media thematically. Access them from the sidebar.

### Collection Types

- **Manual**: Hand-pick items to include
- **Smart**: Define rules (e.g., "all 4K movies from 2024") and items are added automatically
- **Dynamic**: Adaptive criteria that evolve based on usage patterns

### Creating a Collection

1. Navigate to **Collections** in the sidebar
2. Click **New Collection**
3. Choose a type (Manual, Smart, or Dynamic)
4. Set a name and optional description
5. For Smart collections, define filter rules
6. Save the collection

### Visibility

Collections support three visibility levels:
- **Public**: Visible to all users
- **Private**: Visible only to the creator
- **Shared**: Visible to specific users

---

## Playlists

Playlists are ordered sequences designed for sequential playback.

- Create a playlist from the sidebar or from a media detail page
- Drag and drop to reorder items
- Click Play to start playback from the first item
- The player auto-advances through the playlist
- Playlists can be shared with other users

---

## Favorites

Bookmark media items for quick access.

- Click the heart/star icon on any media item to add it to favorites
- Access all favorites from the sidebar
- Export favorites to JSON or CSV for backup or transfer
- Import favorites on another Catalogizer instance -- matching uses metadata, not file paths

---

## Settings

The settings page provides configuration for storage and preferences.

### Storage Sources

- View all connected storage sources with health status
- Add new sources (Local, SMB, FTP, NFS, WebDAV)
- Test connections before saving
- Trigger manual scans per source
- Edit credentials or remove sources

### User Preferences

- Display language and locale
- Default view mode (grid, list, detail)
- Theme preferences
- Notification settings for scan events

### System Configuration (Admin)

Administrators have access to additional settings:

- User management (create, edit, disable accounts)
- Role and permission assignment
- API key management for external metadata providers
- Cache and performance tuning

---

## Real-Time Updates

The web app maintains a WebSocket connection to the backend for live updates.

- **Scan progress**: Progress bar and file count during active scans
- **New media notifications**: Toast notifications when new items are detected
- **Source status changes**: Immediate feedback when a storage source goes offline or recovers
- **Multi-user sync**: Changes made by one user (collections, favorites) appear for others in real time

The connection indicator in the bottom corner shows WebSocket status. It reconnects automatically if the connection drops.
