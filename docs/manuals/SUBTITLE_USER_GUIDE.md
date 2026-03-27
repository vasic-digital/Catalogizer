# Catalogizer -- Subtitle User Guide

## Table of Contents

1. [Overview](#overview)
2. [Subtitle Search](#subtitle-search)
3. [Downloading Subtitles](#downloading-subtitles)
4. [Uploading Custom Subtitles](#uploading-custom-subtitles)
5. [Subtitle Translation](#subtitle-translation)
6. [Subtitle Sync Verification](#subtitle-sync-verification)
7. [Supported Languages](#supported-languages)
8. [Supported Providers](#supported-providers)
9. [Subtitle File Formats](#subtitle-file-formats)
10. [Configuration](#configuration)
11. [Troubleshooting](#troubleshooting)

---

## Overview

Catalogizer includes a comprehensive subtitle management system that allows you to search, download, upload, translate, and verify subtitles for your media files. Subtitles are managed through the web application at `/subtitles`, through the entity detail page for individual media items, or programmatically via the REST API.

The subtitle system integrates with multiple external providers to give you the widest possible selection of subtitles across languages and sources. All subtitle operations respect the semaphore-based concurrency control (default: 3 concurrent downloads) to avoid overwhelming provider APIs.

---

## Subtitle Search

### Searching by Media File

The most common way to find subtitles is by selecting a media file and searching for matching subtitles across all configured providers.

1. Navigate to the entity detail page for a movie or TV episode.
2. In the file section, click the subtitle icon next to a media file.
3. The subtitle search panel opens, pre-populated with the file's metadata (title, year, season, episode).
4. Select your preferred language from the language dropdown.
5. Click **Search**.

The search queries all enabled providers simultaneously and presents results in a unified list, sorted by relevance. Each result shows:

- **Provider name** and source
- **Language** (with country variant if applicable, e.g., Portuguese vs. Brazilian Portuguese)
- **Release name** -- the subtitle release group or upload name
- **Format** (SRT, ASS, VTT, SUB, etc.)
- **Downloads count** -- how many times this subtitle has been downloaded from the provider (a rough quality indicator)
- **Rating** -- user rating from the provider, if available
- **Hearing impaired** flag -- indicates subtitles that include sound descriptions (e.g., "[door creaks]", "[music playing]")

### Searching from the Subtitle Manager

For bulk subtitle management, navigate to `/subtitles` or **Media > Subtitles** in the navigation. The subtitle manager page provides:

- A file browser where you can select any scanned media file
- Batch search across multiple files (select several files and search for all at once)
- A history of previous subtitle downloads

### Search via API

Programmatic subtitle search is available via the REST API:

```
GET /api/v1/subtitles/search?file_id=123&language=en
GET /api/v1/subtitles/search?title=Breaking+Bad&season=3&episode=7&language=es
```

The response includes results from all providers, deduplicated by content hash when possible.

---

## Downloading Subtitles

### Single Download

From the search results list, click **Download** on your preferred subtitle. The subtitle file is downloaded from the provider and stored alongside the media file with a matching name:

```
/media/Movies/The Matrix (1999)/The.Matrix.1999.1080p.BluRay.mkv
/media/Movies/The Matrix (1999)/The.Matrix.1999.1080p.BluRay.en.srt    <- downloaded subtitle
```

The naming convention is `<media-filename>.<language-code>.<format>`. If a subtitle in the same language already exists, a confirmation dialog asks whether to replace it.

### Batch Download

From the subtitle manager, select multiple search results and click **Download Selected** to download all of them at once. Batch downloads are processed sequentially to respect provider rate limits.

### Download Locations

Downloaded subtitles are placed in the same directory as the media file by default. This ensures most media players detect them automatically. The download location can be changed in settings to use a dedicated subtitles directory:

```
Default:    /media/Movies/Film (2024)/Film.en.srt       (same directory)
Alternate:  /media/Movies/Film (2024)/Subs/Film.en.srt  (subdirectory)
```

### Download via API

```
POST /api/v1/subtitles/download
{
    "file_id": 123,
    "provider": "opensubtitles",
    "subtitle_id": "abc-123-def",
    "language": "en"
}
```

---

## Uploading Custom Subtitles

If you have subtitle files that are not available from any provider -- personal translations, corrections, or custom timing adjustments -- you can upload them directly.

### Upload from Entity Detail Page

1. Navigate to the entity detail page for a movie or TV episode.
2. In the file section, click the subtitle icon next to a media file.
3. Click the **Upload** tab in the subtitle panel.
4. Select the subtitle file from your local machine (drag-and-drop is supported).
5. Select the language for the subtitle.
6. Optionally check the **Hearing Impaired** flag if the subtitle includes sound descriptions.
7. Click **Upload**.

The uploaded subtitle is stored alongside the media file using the standard naming convention.

### Upload via API

```
POST /api/v1/subtitles/upload
Content-Type: multipart/form-data

file_id: 123
language: en
hearing_impaired: false
file: <subtitle-file>
```

### Supported Upload Formats

Uploaded subtitles can be in any of the supported formats (SRT, ASS/SSA, VTT, SUB/IDX, SMI). The system validates the file structure and rejects malformed files with a descriptive error message.

---

## Subtitle Translation

Catalogizer can translate existing subtitles from one language to another. Translation is useful when subtitles are available in a language you understand but not in your preferred language.

### How to Translate

1. From the subtitle search results or the existing subtitles list, find a subtitle in the source language.
2. Click the **Translate** button on the subtitle entry.
3. Select the target language from the dropdown.
4. Click **Start Translation**.

The translation is processed server-side. Depending on the subtitle length and provider load, translation typically takes 5-30 seconds for a full-length movie subtitle. Progress is shown in real-time via WebSocket updates.

### Translation Output

The translated subtitle is saved as a new file alongside the original:

```
The.Matrix.1999.1080p.BluRay.en.srt      <- original English
The.Matrix.1999.1080p.BluRay.fr.srt      <- translated French
```

Timing information from the source subtitle is preserved exactly. Only the text content is translated.

### Translation via API

```
POST /api/v1/subtitles/translate
{
    "file_id": 123,
    "source_language": "en",
    "target_language": "fr",
    "subtitle_path": "/media/Movies/The Matrix (1999)/The.Matrix.1999.1080p.BluRay.en.srt"
}
```

### Translation Limitations

- Translation quality depends on the translation service and language pair. Common language pairs (English to Spanish, French, German) typically produce good results. Less common pairs may have lower quality.
- Subtitle styling (colors, positioning, fonts in ASS/SSA format) is preserved, but text layout may change due to different word lengths in the target language.
- Hearing-impaired annotations (e.g., "[music playing]") are translated along with dialogue text.

---

## Subtitle Sync Verification

Subtitle synchronization is critical for a good viewing experience. Even small timing offsets (200-500ms) are noticeable and distracting. Catalogizer provides tools to verify that downloaded or uploaded subtitles are properly synchronized with the media file.

### Automatic Sync Check

When a subtitle is downloaded or uploaded, the system performs an automatic sync verification by comparing the subtitle timing against the media file duration:

- **Duration match**: The subtitle's last timestamp should be within 2 minutes of the media file's total duration. A subtitle that ends 30 minutes before the movie does likely belongs to a different cut or release.
- **Start time check**: The first subtitle entry should appear within the first 10 minutes of the media file. A subtitle starting at the 45-minute mark likely has a global timing offset.

### Sync Verification Results

Each subtitle displays a sync status indicator:

| Status | Icon | Meaning |
|--------|------|---------|
| **Synced** | Green checkmark | Duration and start time match within acceptable tolerances |
| **Warning** | Yellow triangle | Minor timing discrepancy detected (subtitle ends slightly before or after the media) |
| **Out of sync** | Red cross | Significant timing mismatch (wrong cut, major offset, or wrong file entirely) |
| **Unknown** | Gray question mark | Sync verification could not be performed (media file unavailable or unsupported format) |

### Manual Sync Adjustment

If a subtitle is slightly out of sync, you can apply a global time offset:

1. On the entity detail page, find the subtitle in the file section.
2. Click the **Sync** button.
3. Enter a time offset in milliseconds (positive to delay subtitles, negative to advance them).
4. Click **Apply**.

The offset is applied to all timing entries in the subtitle file. A backup of the original file is created before modification.

### Sync Verification via API

```
GET /api/v1/subtitles/sync-check?file_id=123&subtitle_path=/path/to/subtitle.srt
```

Response:

```json
{
    "status": "warning",
    "media_duration_ms": 8160000,
    "subtitle_end_ms": 7920000,
    "offset_ms": -240000,
    "message": "Subtitle ends 4 minutes before media file ends"
}
```

---

## Supported Languages

Catalogizer supports subtitles in all languages provided by the configured subtitle providers. Language selection uses ISO 639-1 two-letter codes with optional country variants. The most commonly used languages include:

| Code | Language | Code | Language |
|------|----------|------|----------|
| `en` | English | `ko` | Korean |
| `es` | Spanish | `ar` | Arabic |
| `fr` | French | `hi` | Hindi |
| `de` | German | `th` | Thai |
| `it` | Italian | `tr` | Turkish |
| `pt` | Portuguese | `pl` | Polish |
| `pt-BR` | Brazilian Portuguese | `nl` | Dutch |
| `ja` | Japanese | `sv` | Swedish |
| `zh` | Chinese (Simplified) | `cs` | Czech |
| `zh-TW` | Chinese (Traditional) | `ro` | Romanian |
| `ru` | Russian | `el` | Greek |

The full list of available languages depends on the provider. The search interface populates the language dropdown dynamically based on provider capabilities.

---

## Supported Providers

### OpenSubtitles

The primary subtitle provider, offering the largest database of subtitles worldwide. Supports subtitle search by file hash (most accurate matching), movie/episode metadata, and free-text search.

- **Coverage**: Movies, TV shows, documentaries
- **Languages**: 70+ languages
- **Features**: Hash-based matching, user ratings, hearing-impaired flag
- **Rate limits**: Varies by subscription tier

### Other Providers

Additional subtitle providers may be configured depending on your deployment. The subtitle system is designed with a provider interface that allows adding new sources. Providers that are unavailable or missing API keys are silently skipped during search -- they do not block results from other providers.

---

## Subtitle File Formats

Catalogizer supports the following subtitle formats for download, upload, and display:

| Format | Extension | Description |
|--------|-----------|-------------|
| **SubRip** | `.srt` | The most common format. Plain text with sequential numbering and timestamps. Widely supported by all media players. |
| **Advanced SubStation** | `.ass`, `.ssa` | Rich formatting support including fonts, colors, positioning, and animation effects. Common for anime fansubs. |
| **WebVTT** | `.vtt` | Web-native format used by HTML5 video players. Similar to SRT with additional styling capabilities. |
| **MicroDVD** | `.sub` | Frame-based timing (requires known frame rate). Often paired with `.idx` index files. |
| **SAMI** | `.smi` | Microsoft format supporting multiple languages in a single file. HTML-like markup. |

When downloading from providers, the subtitle format is determined by the provider. SRT is the most commonly available format and is recommended for maximum compatibility across media players and devices.

---

## Configuration

### Provider API Keys

Subtitle providers may require API keys. Configure them in the `catalog-api` environment or `config.json`:

```env
OPENSUBTITLES_API_KEY=your_api_key_here
OPENSUBTITLES_USERNAME=your_username
OPENSUBTITLES_PASSWORD=your_password
```

### Concurrency Settings

Subtitle download concurrency is controlled by the semaphore system. The default limit of 3 concurrent downloads can be adjusted:

```json
{
  "concurrency": {
    "subtitle_download": 3
  }
}
```

Or via environment variable: `CONCURRENCY_SUBTITLE_DOWNLOAD=5`

### Default Language

Set a default subtitle language preference in your user settings to avoid selecting it on every search:

1. Navigate to **Settings > Preferences**.
2. Under **Subtitles**, select your preferred language.
3. Click **Save**.

The default language pre-populates the language dropdown in all subtitle search interfaces.

---

## Troubleshooting

### No Search Results

- Verify that at least one subtitle provider is configured with valid API keys.
- Check the application logs (**Admin > Logs**, filter by subtitle component) for provider errors.
- Try searching with simplified metadata (title and year only, without release group or resolution).
- Some older or obscure titles may not have subtitles available from any provider.

### Download Fails

- Provider rate limits may be active. Wait a few minutes and retry.
- Check that the storage root where the media file resides is writable. Read-only storage roots (e.g., NFS exports without write access) cannot store downloaded subtitles.
- Verify network connectivity from the catalog-api server to the subtitle provider.

### Subtitles Out of Sync

- Ensure the subtitle matches the correct release of the media (theatrical cut vs. extended edition, Blu-ray vs. web release). Different cuts have different timing.
- Use the manual sync adjustment to apply a global offset if the subtitle is consistently early or late.
- Try downloading a different subtitle from the search results -- multiple uploads for the same media may have different timing accuracy.

### Translation Not Available

- Translation requires a configured translation service. Check that the necessary API keys or service endpoints are configured.
- Not all language pairs may be supported. Try an intermediate translation (e.g., source language to English, then English to target language).

### Subtitle Not Detected by Media Player

- Verify the subtitle file is in the same directory as the media file and has a matching base name.
- Some media players require subtitles to have the exact same filename as the media file (excluding the language code and format extension).
- Ensure the subtitle format is supported by your media player. SRT has the broadest compatibility.
- On Android and Android TV, the Catalogizer apps detect subtitles automatically and present them in the player's subtitle selection menu.
