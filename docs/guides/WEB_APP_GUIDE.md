# Catalogizer Web App User Guide

This guide covers the Catalogizer web application in detail, including all features accessible through the browser-based interface.

## Table of Contents

1. [Getting Started](#getting-started)
2. [Dashboard](#dashboard)
3. [Media Browser](#media-browser)
4. [Collections](#collections)
5. [Favorites](#favorites)
6. [Playlists](#playlists)
7. [Subtitle Manager](#subtitle-manager)
8. [Format Converter](#format-converter)
9. [Analytics](#analytics)
10. [AI Dashboard](#ai-dashboard)
11. [Administration](#administration)
12. [Keyboard Shortcuts](#keyboard-shortcuts)

---

## Getting Started

### Login and Registration

1. Open your browser and navigate to your Catalogizer server URL (e.g. `http://localhost:5173`).
2. On the login screen, enter your username and password.
3. If your administrator has enabled registration, a "Register" link will be visible below the login form.
4. After successful login, you are redirected to the Dashboard. Your session is maintained via a JWT token stored in the browser.

### First-Time Setup

If this is a fresh installation:

1. Log in with the default admin credentials provided by your administrator.
2. Navigate to the Admin panel to configure storage sources.
3. Trigger a library scan from the Dashboard quick actions to discover media.
4. Change your password in your profile settings.

### Navigation

The main navigation bar appears at the top of the screen and includes:

- **Dashboard** -- overview and quick actions
- **Media Browser** -- browse and search the full media catalog
- **Collections** -- organize media into collections
- **Favorites** -- your bookmarked media items
- **Playlists** -- create and manage playlists
- **Subtitles** -- search, download, and manage subtitles
- **Conversion** -- convert media between formats
- **Analytics** -- charts and statistics about your library
- **Admin** -- administration panel (admin users only)
- **Profile Menu** -- account settings and logout (top-right corner)

---

## Dashboard

The Dashboard is the landing page after login. It displays a personalized greeting and provides an at-a-glance view of your media library.

### Statistics Overview

At the top of the Dashboard, you will see key metrics cards:

- **Total Media Items** -- the total number of items in your library
- **Active Users** -- number of users currently online
- **Sessions Today** -- total login sessions for the day
- **Average Session Duration** -- how long users typically stay logged in

### Media Distribution Chart

A pie chart shows the breakdown of your library by media type (movies, TV shows, music, documents, etc.), giving you a visual sense of your collection composition.

### System Status Panel

Real-time system health indicators:

- **CPU Usage** -- current server CPU utilization with color-coded bar (green/yellow/red)
- **Memory Usage** -- RAM utilization
- **Disk Usage** -- storage space consumed
- **Network Status** -- online/offline indicator
- **Uptime** -- how long the server has been running

### Activity Feed

A chronological list of recent events in your library, such as:

- New media detected
- Collections created or updated
- Files shared or downloaded

### Quick Actions

Four action buttons for common tasks:

- **Upload Media** -- opens the upload interface
- **Scan Library** -- triggers a re-scan of all configured storage sources to detect new files
- **Search** -- jumps to the Media Browser with the search bar focused
- **Settings** -- opens system settings

---

## Media Browser

The Media Browser is the primary interface for exploring your media catalog.

### Stats Cards

Four summary cards appear at the top:

- **Total Items** -- number of items in the library
- **Media Types** -- count of distinct media type categories
- **Total Size** -- combined storage size (displayed in GB)
- **Recent Additions** -- number of items added recently

### Searching

Type in the search bar to filter media by title, description, or other metadata. Search is debounced (300ms delay) so results update as you type without excessive API calls.

### Filtering

Click the **Filters** button to open a sidebar filter panel. Available filters include:

- Media type (movie, TV show, music, documentary, anime, etc.)
- Quality level
- Year range
- File size range
- Sort order (by title, date updated, date added, year, rating, file size)
- Sort direction (ascending or descending)

Click **Reset** to clear all active filters.

### View Modes

Toggle between two view modes using the buttons next to the filter controls:

- **Grid View** -- displays media as visual cards in a responsive grid layout
- **List View** -- displays media in a compact list format with more details per row

### Uploading Media

Click the **Upload** button (arrow-up icon) to expand the Upload Manager panel. You can drag and drop files or click to select files from your filesystem. The upload manager shows progress for each file being uploaded and refreshes the media list on completion.

### Viewing Media Details

Click on any media item to open the **Media Detail Modal**, which shows:

- Title, year, and quality information
- Media type and file size
- Description and metadata from external providers (TMDB, IMDB)
- Action buttons: Play, Download, Close

### Playing Media

Click the **Play** button on a media card or from the detail modal to open the fullscreen **Media Player**. The player supports video and audio playback with standard controls (play/pause, seek, volume). Click **Close** to exit the player.

### Downloading Media

Click the **Download** button on a media card or in the detail modal. A toast notification confirms when the download starts and completes.

### Pagination

When your library contains more items than the page limit (24 items by default), pagination controls appear at the top and bottom of the results:

- **Previous/Next** buttons navigate between pages
- A page indicator shows your current position (e.g. "Page 2 of 15")

### Refreshing

Click the **Refresh** button (circular arrows icon) to reload the current results from the server.

---

## Collections

Collections let you organize media items into themed groups.

### Collection Types

- **Manual Collections** -- you manually add and remove items
- **Smart Collections** -- automatically populated based on rules you define (e.g. "all movies from 2024 with rating > 7")
- **Favorites** -- a built-in collection of your favorite items

### Creating Collections

1. Click the **Smart Collection** button to open the Smart Collection Builder.
2. Enter a name and description.
3. Define rules (conditions) that determine which media items are automatically included.
4. Click **Save** to create the collection.

### Browsing Collections

- Use the tab bar to filter by: All Collections, Smart Collections, Manual Collections, Favorites, Templates, Automation, Integrations, AI Features.
- Search collections by name using the search bar.
- Filter by media type (All, Music, Video, Images, Documents).
- Sort by name, date created, date updated, or item count.
- Switch between grid and list views.

### Collection Actions

For each collection, you can:

- **Preview** -- see collection contents
- **Share** -- share with other users (set view, comment, download permissions)
- **Duplicate** -- create a copy of the collection
- **Export** -- export as JSON, CSV, or M3U format
- **Settings** -- edit collection name, description, and rules
- **Analytics** -- view statistics for the collection
- **Real-Time Collaboration** -- enable live collaboration with other users
- **Delete** -- permanently remove the collection

### Bulk Operations

1. Select multiple collections using the checkboxes on each card.
2. Or use "Select all" to select every visible collection.
3. Click **Bulk Actions** to perform operations on all selected collections at once: delete, share, export, or duplicate.

### AI Features

The Collections page includes AI-powered features accessible through the "AI Features" tab:

- **AI Collection Suggestions** -- automated recommendations for new collections
- **Natural Language Search** -- search using conversational queries like "show me action movies"
- **Content Categorizer** -- AI-powered categorization of media items
- **Behavior Analytics** -- understand usage patterns
- **Smart Organization** -- suggestions for reorganizing collections
- **Metadata Extraction** -- extract metadata from content using AI
- **Automation Rules** -- create AI-driven automation rules

---

## Favorites

The Favorites page provides a dedicated view for managing your bookmarked media items.

### Tabs

- **Favorites** -- your current favorites displayed in a filterable grid
- **Recently Added** -- favorites sorted by when you added them
- **Statistics** -- insights about your favorites

### Statistics Tab

Displays:

- Total favorite count
- Breakdown by media type (e.g. 45 movies, 12 TV shows, 8 music albums)
- Most common type
- Recent activity count (items added in the last week)

### Bulk Actions

Toggle **Bulk Actions** mode to select multiple favorites for batch operations.

### Import/Export

- **Import** -- load favorites from a previously exported file
- **Export** -- save your favorites list to a file for backup or sharing

---

## Playlists

Playlists allow you to create ordered sequences of media items for playback.

### Features

- Create, edit, and delete playlists
- Search and filter playlists
- **Smart Playlist Builder** -- define rules to automatically populate playlists
- **Playlist Player** -- play through playlist items in sequence
- Shuffle mode for randomized playback
- Drag and drop to reorder items within a playlist

---

## Subtitle Manager

The Subtitle Manager lets you search, download, upload, and manage subtitles for your media files.

### Selecting Media

1. Click **Select Media** to open the media search interface.
2. Type a title to search your library.
3. Click on a result to select it as the active media item.

### Searching Subtitles

1. With a media item selected, type a search query or use the pre-filled title.
2. Click **Filters** to refine your search:
   - **Language** -- select from a list of supported languages
   - **Providers** -- toggle which subtitle providers to search (e.g. OpenSubtitles, Subscene)
3. Click **Search** to fetch results.

### Search Results

Each result shows:

- Subtitle title
- Language
- Provider source
- Rating (if available)
- Release group
- Tags: HI (hearing impaired), Foreign Parts Only, Machine Translated

Click **Download** to download the subtitle and associate it with your selected media.

### Managing Existing Subtitles

When a media item is selected, the bottom section shows all subtitles currently associated with it, including:

- Language name
- Provider and format (e.g. SRT, SUB)
- Encoding information
- Sync offset (in milliseconds)
- Verification status

For each subtitle, you can:

- **Verify Sync** -- opens the Subtitle Sync Modal to check and adjust timing
- **Delete** -- remove the subtitle

### Uploading Subtitles

Click **Upload Subtitle** to open a modal where you can upload a subtitle file from your computer and associate it with the selected media item.

---

## Format Converter

The Format Converter page allows you to convert media files between different formats.

### Supported Formats

- **Video**: MP4, MKV, AVI, MOV, WebM
- **Audio**: MP3, WAV, FLAC

### Starting a Conversion

1. Select a source file from your media library.
2. Choose the target output format.
3. Configure quality settings.
4. Click **Start Conversion**.

### Managing Conversion Jobs

The page displays a list of all conversion jobs with:

- Job status (pending, in progress, completed, failed, cancelled)
- Progress percentage for active jobs
- Source and target format information

Available actions per job:

- **Cancel** -- stop an in-progress conversion
- **Retry** -- restart a failed conversion
- **Download** -- download the converted file when complete

The job list auto-refreshes every 30 seconds to show updated progress.

---

## Analytics

The Analytics page provides visual insights into your media collection through interactive charts.

### Key Metrics

Four stat cards with mini trend charts:

- **Total Media Items** -- with month-over-month growth percentage
- **Storage Used** -- total disk usage in GB with weekly change
- **Recent Additions** -- count of newly detected items
- **Media Types** -- number of distinct media categories

### Charts

- **Media Types Distribution** -- pie chart showing the percentage breakdown by type (movies, TV shows, music, etc.)
- **Quality Distribution** -- bar chart showing counts of media by quality level (720p, 1080p, 4K, etc.)
- **Collection Growth** -- area chart showing how your library has grown over time
- **Weekly Activity** -- bar chart showing new additions for each day of the current week

### Recently Added Media

A scrollable list of the 10 most recently added items, showing:

- Media type icon
- Title and year
- Type badge, quality, and file size
- Date added

---

## AI Dashboard

The AI Dashboard provides AI-powered insights and tools for your media collection. Features include:

- AI-generated collection suggestions
- Natural language search capabilities
- Content quality analysis
- Automated categorization
- Predictive analytics

---

## Administration

The Admin panel is available to users with administrator privileges.

### System Information

View server health metrics, version information, and uptime statistics.

### User Management

- View all registered users
- Create new user accounts
- Edit user roles and permissions
- Deactivate or delete accounts

### Storage Management

- View configured storage sources and their status
- Add new storage sources (SMB, FTP, NFS, WebDAV, Local)
- Test storage connections
- Remove storage sources

### Backup Management

- View available backups
- Create new backups
- Restore from a backup

---

## Keyboard Shortcuts

The web interface supports the following keyboard interactions:

- **Enter** in search fields -- triggers search
- **Escape** -- closes modals and overlays
- Standard browser shortcuts apply (Ctrl+F for browser search, etc.)

---

## Common Workflows

### Adding Media to Your Library

1. Go to Dashboard and click **Scan Library** to detect media from configured storage sources.
2. Alternatively, go to Media Browser and click the **Upload** button to manually upload files.
3. Browse newly added items in the Media Browser.

### Organizing Your Collection

1. Browse your media in the Media Browser.
2. Create a Smart Collection with rules matching a theme (e.g. "Horror movies from 2020-2024").
3. Share the collection with other users.

### Finding and Adding Subtitles

1. Go to the Subtitle Manager.
2. Select the media item you want subtitles for.
3. Search for subtitles in your preferred language.
4. Download the best matching result.
5. Verify sync timing if needed.

### Converting Media Formats

1. Go to Format Converter.
2. Select the source file.
3. Choose the desired output format and quality.
4. Start the conversion and monitor progress.
5. Download the converted file when complete.
