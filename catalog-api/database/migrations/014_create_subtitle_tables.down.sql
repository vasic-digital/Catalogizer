-- Drop subtitle-related tables in reverse order of creation
DROP TRIGGER IF EXISTS update_subtitle_sync_status_updated_at;
DROP TRIGGER IF EXISTS set_subtitle_sync_status_completed_at;
DROP TRIGGER IF EXISTS update_subtitle_tracks_updated_at;

DROP TABLE IF EXISTS media_subtitles;
DROP TABLE IF EXISTS subtitle_downloads;
DROP TABLE IF EXISTS subtitle_cache;
DROP TABLE IF EXISTS subtitle_sync_status;
DROP TABLE IF EXISTS subtitle_tracks;