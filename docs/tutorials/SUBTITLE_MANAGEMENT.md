# Subtitle Management

This tutorial covers searching for subtitles, downloading them, verifying synchronization, translating subtitles, and uploading custom subtitle files.

## Prerequisites

- Catalogizer API running with media items in the catalog
- A valid authentication token (see [Quick Start](QUICK_START.md))
- Optional: API keys for subtitle providers (OpenSubtitles, SubDB, Yify Subtitles, Subscene, Addic7ed)

## Overview

Catalogizer provides a comprehensive subtitle management system accessible through the REST API at `/api/v1/subtitles`. Supported operations include:

- Searching across multiple subtitle providers
- Downloading subtitles and associating them with media items
- Verifying subtitle-to-video synchronization
- Translating subtitles to different languages
- Uploading custom subtitle files

## Step 1: Search for Subtitles

Search for subtitles by media path, title, or file hash:

```bash
curl -X POST http://localhost:8080/api/v1/subtitles/search \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <your-token>" \
  -d '{
    "media_path": "/media/Movies/Inception.2010.1080p.mkv",
    "title": "Inception",
    "year": 2010,
    "languages": ["en", "es", "fr"],
    "providers": ["opensubtitles", "subdb"]
  }'
```

Fields:
- **media_path**: Path to the media file on the storage source
- **title**: Movie or episode title (optional, helps narrow results)
- **year**: Release year (optional)
- **season** / **episode**: For TV shows (optional)
- **languages**: Array of ISO 639-1 language codes
- **providers**: Specific providers to search (optional, searches all by default)
- **file_hash** / **file_size**: For hash-based matching (optional, improves accuracy)

**Expected result:** A JSON response with a list of subtitle results:

```json
{
  "success": true,
  "results": [
    {
      "id": "os-12345",
      "provider": "opensubtitles",
      "language": "English",
      "language_code": "en",
      "title": "Inception.2010.1080p.BluRay",
      "format": "srt",
      "downloads": 52340,
      "rating": 9.2,
      "is_hearing_impaired": false,
      "match_score": 0.95
    }
  ],
  "count": 8
}
```

Results are ranked by `match_score`, which accounts for hash matching, title similarity, and provider rating.

## Step 2: Download a Subtitle

Download a subtitle from the search results:

```bash
curl -X POST http://localhost:8080/api/v1/subtitles/download \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <your-token>" \
  -d '{
    "subtitle_id": "os-12345",
    "provider": "opensubtitles",
    "media_item_id": 42,
    "language_code": "en"
  }'
```

**Expected result:** A JSON response confirming the download with subtitle track details:

```json
{
  "success": true,
  "message": "Subtitle downloaded successfully",
  "track": {
    "id": 1,
    "media_item_id": 42,
    "language": "English",
    "language_code": "en",
    "format": "srt",
    "provider": "opensubtitles",
    "is_default": false
  }
}
```

## Step 3: Verify Subtitle Synchronization

Check if the downloaded subtitle is properly synchronized with the video:

```bash
curl -X POST http://localhost:8080/api/v1/subtitles/<track-id>/sync \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <your-token>" \
  -d '{
    "media_item_id": 42
  }'
```

**Expected result:** A sync verification result:

```json
{
  "success": true,
  "sync_result": {
    "is_synced": true,
    "offset_ms": 120,
    "confidence": 0.92,
    "suggested_offset_ms": 0
  }
}
```

If `is_synced` is false, the `suggested_offset_ms` value indicates the recommended timing adjustment.

## Step 4: Translate Subtitles

Translate an existing subtitle track to another language:

```bash
curl -X POST http://localhost:8080/api/v1/subtitles/<track-id>/translate \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <your-token>" \
  -d '{
    "target_language": "es",
    "source_language": "en"
  }'
```

**Expected result:** A new subtitle track in the target language:

```json
{
  "success": true,
  "translated_track": {
    "id": 2,
    "media_item_id": 42,
    "language": "Spanish",
    "language_code": "es",
    "format": "srt",
    "provider": "translation",
    "is_default": false
  }
}
```

The translation service supports batch translation and preserves subtitle timing. Translation providers can be configured via the API.

## Step 5: Upload Custom Subtitles

Upload your own subtitle file:

```bash
curl -X POST http://localhost:8080/api/v1/subtitles/upload \
  -H "Authorization: Bearer <your-token>" \
  -F "file=@/path/to/subtitle.srt" \
  -F "media_item_id=42" \
  -F "language_code=en" \
  -F "format=srt"
```

Supported formats: SRT, ASS, SSA, VTT, SUB

**Expected result:** A confirmation with the created subtitle track:

```json
{
  "success": true,
  "message": "Subtitle uploaded successfully",
  "track": {
    "id": 3,
    "media_item_id": 42,
    "language": "English",
    "language_code": "en",
    "format": "srt",
    "provider": "custom"
  }
}
```

## Step 6: List Subtitles for a Media Item

View all subtitle tracks associated with a media item:

```bash
curl -H "Authorization: Bearer <your-token>" \
  http://localhost:8080/api/v1/subtitles/media/42
```

**Expected result:** A list of all subtitle tracks:

```json
{
  "success": true,
  "subtitles": [
    {
      "id": 1,
      "language": "English",
      "language_code": "en",
      "format": "srt",
      "provider": "opensubtitles",
      "is_default": true
    },
    {
      "id": 2,
      "language": "Spanish",
      "language_code": "es",
      "format": "srt",
      "provider": "translation"
    }
  ],
  "media_item_id": 42
}
```

## Supported Subtitle Providers

| Provider | Description | API Key Required |
|----------|-------------|-----------------|
| OpenSubtitles | Largest subtitle database | Yes |
| SubDB | Hash-based subtitle matching | No |
| Yify Subtitles | Subtitles for YIFY/YTS releases | No |
| Subscene | Community-driven subtitle site | No |
| Addic7ed | TV show focused subtitles | Yes |

Configure API keys in your `.env` file:

```env
OPENSUBTITLES_API_KEY=your_key
ADDIC7ED_USERNAME=your_username
ADDIC7ED_PASSWORD=your_password
```

## Troubleshooting

### No subtitle results found

- Verify the media title and year are correct
- Try searching with fewer language filters
- Use file hash matching for more accurate results (provide `file_hash` and `file_size`)
- Some providers may be unavailable; try specifying different providers

### Subtitle synchronization issues

- Use the sync verification endpoint to check offset
- Apply the `suggested_offset_ms` to adjust timing
- Try downloading from a different provider, as subtitle timing varies between releases

### Translation produces poor results

- Verify the source language is correctly identified
- Translation quality depends on the configured translation provider
- For subtitles with technical jargon or slang, results may need manual review

### Upload fails with format error

- Ensure the subtitle file encoding is UTF-8
- Verify the file extension matches the declared format
- Check that the file is not corrupted (open it in a text editor to verify)
