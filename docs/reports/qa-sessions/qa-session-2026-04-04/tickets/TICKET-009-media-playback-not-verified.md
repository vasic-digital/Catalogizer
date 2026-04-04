# [MAJOR] Media Playback: No video/audio content successfully played during QA

**Platform**: All (Android TV, Web, Android Phone)
**Severity**: MAJOR
**Discovered by**: Manual observation during HelixQA sessions

## Description

Across all HelixQA autonomous sessions, no media content (video, audio, images) was successfully played or streamed. The LLM navigated to detail screens and pressed "Play" buttons, but actual content playback was never confirmed via screenshots or video recording.

## Evidence

- Android TV session: LLM reached player screen (curiosity step #21-24) but player controls showed no progress
- Web session: LLM stayed on login/dashboard, never reached player
- Android phone: LLM browsed categories but thumbnails didn't load, playback not attempted

## API Verification

The streaming API works:
- `GET /api/v1/stream/169074` returns HTTP 206 with `application/octet-stream` (video file)
- `GET /api/v1/stream/163956` returns HTTP 206 with `application/x-subrip` (subtitle)
- Entity 1 ("2001 A Space Odyssey") has 10 associated files

## Root Cause (suspected)

1. Entity files have `IsPrimary: true` on subtitle files — player may pick subtitle instead of video
2. The app's player may construct stream URL incorrectly
3. Search returns "No results" because search query format doesn't match entity titles
4. Mi Box connected to wrong server (localhost vs amber.local) during some sessions

## Required Fix

1. Ensure `IsPrimary` is set on video files, not subtitles
2. Add media playback verification step to HelixQA (check player progress > 0)
3. Fix search to match partial entity titles
4. Ensure consistent server URL across all sessions
