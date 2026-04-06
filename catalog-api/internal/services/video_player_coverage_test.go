package services

import (
	"catalogizer/database"
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/mutecomm/go-sqlcipher"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func setupVideoPlayerTestDB(t *testing.T) (*database.DB, *VideoPlayerService) {
	t.Helper()
	sqlDB, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)

	db := database.WrapDB(sqlDB, database.DialectSQLite)

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS files (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			storage_root_id INTEGER NOT NULL,
			name TEXT NOT NULL,
			path TEXT NOT NULL,
			extension TEXT, mime_type TEXT, file_type TEXT,
			size INTEGER DEFAULT 0, is_directory BOOLEAN DEFAULT 0,
			modified_at DATETIME, last_scan_at DATETIME,
			parent_id INTEGER, deleted BOOLEAN DEFAULT 0
		);
		CREATE TABLE IF NOT EXISTS media_items (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT, original_title TEXT, description TEXT,
			type TEXT, file_path TEXT, file_size INTEGER DEFAULT 0,
			duration INTEGER DEFAULT 0, resolution TEXT, aspect_ratio TEXT,
			frame_rate REAL DEFAULT 0, bitrate INTEGER DEFAULT 0,
			codec TEXT, hdr BOOLEAN DEFAULT 0,
			dolby_vision BOOLEAN DEFAULT 0, dolby_atmos BOOLEAN DEFAULT 0,
			year INTEGER DEFAULT 0, release_date DATETIME,
			genres TEXT, directors TEXT, actors TEXT, writers TEXT,
			rating REAL, imdb_id TEXT, tmdb_id TEXT,
			language TEXT, country TEXT,
			play_count INTEGER DEFAULT 0, last_played DATETIME,
			date_added DATETIME DEFAULT CURRENT_TIMESTAMP,
			user_rating INTEGER, is_favorite BOOLEAN DEFAULT 0,
			watched_percentage REAL DEFAULT 0.0,
			parent_id INTEGER
		);
		CREATE TABLE IF NOT EXISTS series (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT, description TEXT,
			total_seasons INTEGER DEFAULT 0, total_episodes INTEGER DEFAULT 0,
			status TEXT, first_aired DATETIME, last_aired DATETIME,
			network TEXT, creator TEXT
		);
		CREATE TABLE IF NOT EXISTS episodes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			media_item_id INTEGER, series_id INTEGER,
			season_number INTEGER, episode_number INTEGER,
			air_date DATETIME, runtime INTEGER DEFAULT 0,
			guest_stars TEXT, next_episode_id INTEGER, prev_episode_id INTEGER
		);
		CREATE TABLE IF NOT EXISTS movies (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			media_item_id INTEGER, budget INTEGER DEFAULT 0,
			revenue INTEGER DEFAULT 0, runtime INTEGER DEFAULT 0,
			collection TEXT, studio TEXT, production_companies TEXT
		);
		CREATE TABLE IF NOT EXISTS video_streams (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			media_item_id INTEGER, stream_index INTEGER,
			codec TEXT, width INTEGER, height INTEGER,
			bitrate INTEGER, fps REAL, language TEXT, title TEXT,
			is_default BOOLEAN DEFAULT 0
		);
		CREATE TABLE IF NOT EXISTS audio_streams (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			media_item_id INTEGER, stream_index INTEGER,
			codec TEXT, channels INTEGER, bitrate INTEGER,
			language TEXT, title TEXT, is_default BOOLEAN DEFAULT 0
		);
		CREATE TABLE IF NOT EXISTS subtitle_streams (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			media_item_id INTEGER, stream_index INTEGER,
			codec TEXT, language TEXT, title TEXT,
			is_default BOOLEAN DEFAULT 0, is_forced BOOLEAN DEFAULT 0
		);
		CREATE TABLE IF NOT EXISTS subtitles (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			media_item_id INTEGER, language TEXT, language_code TEXT,
			format TEXT, file_path TEXT, encoding TEXT,
			is_embedded BOOLEAN DEFAULT 0, is_default BOOLEAN DEFAULT 0,
			sync_offset INTEGER DEFAULT 0
		);
		CREATE TABLE IF NOT EXISTS video_chapters (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			media_item_id INTEGER, title TEXT,
			start_time INTEGER, end_time INTEGER,
			thumbnail_url TEXT
		);
		CREATE TABLE IF NOT EXISTS video_playback_sessions (
			id TEXT PRIMARY KEY, user_id INTEGER,
			session_data TEXT, expires_at DATETIME,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		CREATE TABLE IF NOT EXISTS playback_positions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER, media_id INTEGER,
			position INTEGER DEFAULT 0, duration INTEGER DEFAULT 0,
			percent_complete REAL DEFAULT 0.0,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(user_id, media_id)
		);
		CREATE TABLE IF NOT EXISTS watch_history (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER, video_id INTEGER,
			watched_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			watch_duration INTEGER DEFAULT 0,
			completion_rate REAL DEFAULT 0.0,
			stopped_at INTEGER DEFAULT 0,
			device_info TEXT, quality TEXT
		);
	`)
	require.NoError(t, err)

	logger := zap.NewNop()
	positionService := NewPlaybackPositionService(db, logger)
	svc := NewVideoPlayerService(db, logger, nil, positionService, nil, nil, nil)

	return db, svc
}

// =============================================================================
// getSeriesInfo
// =============================================================================

func TestVideoPlayerService_getSeriesInfo_Found(t *testing.T) {
	db, svc := setupVideoPlayerTestDB(t)

	_, err := db.Exec(`INSERT INTO series (id, title, description, total_seasons, total_episodes, status, network)
		VALUES (1, 'Breaking Bad', 'A show', 5, 62, 'ended', 'AMC')`)
	require.NoError(t, err)

	info, err := svc.getSeriesInfo(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, int64(1), info.ID)
	assert.Equal(t, "Breaking Bad", info.Title)
	assert.Equal(t, 5, info.TotalSeasons)
	assert.Equal(t, 62, info.TotalEpisodes)
	assert.Equal(t, "ended", info.Status)
}

func TestVideoPlayerService_getSeriesInfo_NotFound(t *testing.T) {
	_, svc := setupVideoPlayerTestDB(t)

	_, err := svc.getSeriesInfo(context.Background(), 999)
	assert.Error(t, err)
}

func TestVideoPlayerService_getSeriesInfo_WithCreator(t *testing.T) {
	db, svc := setupVideoPlayerTestDB(t)

	_, err := db.Exec(`INSERT INTO series (id, title, description, total_seasons, total_episodes, status, network, creator)
		VALUES (2, 'Test', 'Desc', 3, 30, 'running', 'HBO', '["Creator One"]')`)
	require.NoError(t, err)

	info, err := svc.getSeriesInfo(context.Background(), 2)
	require.NoError(t, err)
	assert.Equal(t, "Test", info.Title)
	assert.Equal(t, 3, info.TotalSeasons)
}

// =============================================================================
// getSeasonEpisodes
// =============================================================================

func TestVideoPlayerService_getSeasonEpisodes_Empty(t *testing.T) {
	_, svc := setupVideoPlayerTestDB(t)

	episodes, err := svc.getSeasonEpisodes(context.Background(), 1, 1)
	assert.NoError(t, err)
	assert.Empty(t, episodes)
}

func TestVideoPlayerService_getSeasonEpisodes_QueryRuns(t *testing.T) {
	db, svc := setupVideoPlayerTestDB(t)

	// Create media items with all required columns
	_, err := db.Exec(`INSERT INTO media_items (id, title, original_title, description, type, file_path, file_size, duration, resolution, aspect_ratio, frame_rate, bitrate, codec, hdr, dolby_vision, dolby_atmos, year, genres, directors, actors, writers, language, country, play_count, date_added, watched_percentage, imdb_id, tmdb_id)
		VALUES
		(10, 'S01E01', '', '', 'video', '/s01e01.mp4', 0, 3000, '', '', 0, 0, '', 0, 0, 0, 0, '[]', '[]', '[]', '[]', '', '', 0, CURRENT_TIMESTAMP, 0, '', ''),
		(11, 'S01E02', '', '', 'video', '/s01e02.mp4', 0, 2800, '', '', 0, 0, '', 0, 0, 0, 0, '[]', '[]', '[]', '[]', '', '', 0, CURRENT_TIMESTAMP, 0, '', '')`)
	require.NoError(t, err)

	_, err = db.Exec(`INSERT INTO episodes (media_item_id, series_id, season_number, episode_number) VALUES
		(10, 1, 1, 1), (11, 1, 1, 2)`)
	require.NoError(t, err)

	episodes, err := svc.getSeasonEpisodes(context.Background(), 1, 1)
	assert.NoError(t, err)
	assert.Len(t, episodes, 2)
}

// =============================================================================
// loadSeriesInfo
// =============================================================================

func TestVideoPlayerService_loadSeriesInfo_NotEpisode(t *testing.T) {
	_, svc := setupVideoPlayerTestDB(t)

	video := &VideoContent{Type: VideoTypeMovie}
	err := svc.loadSeriesInfo(context.Background(), video)
	assert.NoError(t, err)
	assert.Nil(t, video.SeriesInfo)
}

func TestVideoPlayerService_loadSeriesInfo_Episode_NoData(t *testing.T) {
	_, svc := setupVideoPlayerTestDB(t)

	// No series/episodes in DB - should return error
	video := &VideoContent{ID: 100, Type: VideoTypeEpisode}
	err := svc.loadSeriesInfo(context.Background(), video)
	assert.Error(t, err) // no rows
	assert.Nil(t, video.SeriesInfo)
}

// =============================================================================
// loadEpisodeInfo
// =============================================================================

func TestVideoPlayerService_loadEpisodeInfo_NotEpisode(t *testing.T) {
	_, svc := setupVideoPlayerTestDB(t)

	video := &VideoContent{Type: VideoTypeMovie}
	err := svc.loadEpisodeInfo(context.Background(), video)
	assert.NoError(t, err)
	assert.Nil(t, video.EpisodeInfo)
}

func TestVideoPlayerService_loadEpisodeInfo_Episode(t *testing.T) {
	db, svc := setupVideoPlayerTestDB(t)

	_, err := db.Exec(`INSERT INTO episodes (media_item_id, series_id, season_number, episode_number, runtime, air_date, guest_stars) VALUES
		(200, 1, 2, 5, 45, '2022-05-15', '["Actor A","Actor B"]')`)
	require.NoError(t, err)

	video := &VideoContent{ID: 200, Type: VideoTypeEpisode}
	err = svc.loadEpisodeInfo(context.Background(), video)
	assert.NoError(t, err)
	require.NotNil(t, video.EpisodeInfo)
	assert.Equal(t, 2, video.EpisodeInfo.SeasonNumber)
	assert.Equal(t, 5, video.EpisodeInfo.EpisodeNumber)
	assert.NotNil(t, video.EpisodeInfo.AirDate)
}

func TestVideoPlayerService_loadEpisodeInfo_NotFound(t *testing.T) {
	_, svc := setupVideoPlayerTestDB(t)

	video := &VideoContent{ID: 999, Type: VideoTypeEpisode}
	err := svc.loadEpisodeInfo(context.Background(), video)
	assert.Error(t, err)
}

// =============================================================================
// loadMovieInfo
// =============================================================================

func TestVideoPlayerService_loadMovieInfo_NotMovie(t *testing.T) {
	_, svc := setupVideoPlayerTestDB(t)

	video := &VideoContent{Type: VideoTypeEpisode}
	err := svc.loadMovieInfo(context.Background(), video)
	assert.NoError(t, err)
	assert.Nil(t, video.MovieInfo)
}

func TestVideoPlayerService_loadMovieInfo_Found(t *testing.T) {
	db, svc := setupVideoPlayerTestDB(t)

	_, err := db.Exec(`INSERT INTO movies (media_item_id, budget, revenue, runtime, collection, studio, production_companies) VALUES
		(300, 150000000, 800000000, 148, 'MCU', '"Marvel Studios"', '["Marvel","Disney"]')`)
	require.NoError(t, err)

	video := &VideoContent{ID: 300, Type: VideoTypeMovie}
	err = svc.loadMovieInfo(context.Background(), video)
	assert.NoError(t, err)
	require.NotNil(t, video.MovieInfo)
	assert.Equal(t, int64(150000000), video.MovieInfo.Budget)
	assert.Equal(t, int64(800000000), video.MovieInfo.Revenue)
}

func TestVideoPlayerService_loadMovieInfo_NotFound(t *testing.T) {
	_, svc := setupVideoPlayerTestDB(t)

	video := &VideoContent{ID: 999, Type: VideoTypeMovie}
	err := svc.loadMovieInfo(context.Background(), video)
	assert.Error(t, err)
}

// =============================================================================
// loadVideoStreams
// =============================================================================

func TestVideoPlayerService_loadVideoStreams(t *testing.T) {
	db, svc := setupVideoPlayerTestDB(t)

	_, err := db.Exec(`INSERT INTO video_streams (media_item_id, stream_index, codec, width, height, bitrate, fps, language, title, is_default) VALUES
		(400, 0, 'h264', 1920, 1080, 5000000, 23.976, 'en', 'Main', 1)`)
	require.NoError(t, err)

	_, err = db.Exec(`INSERT INTO audio_streams (media_item_id, stream_index, codec, channels, bitrate, language, title, is_default) VALUES
		(400, 1, 'aac', 6, 384000, 'en', 'English 5.1', 1),
		(400, 2, 'aac', 2, 192000, 'fr', 'French', 0)`)
	require.NoError(t, err)

	video := &VideoContent{ID: 400}
	session := &VideoPlaybackSession{CurrentVideo: video}
	err = svc.loadVideoStreams(context.Background(), session, 400)
	assert.NoError(t, err)
	assert.Len(t, session.CurrentVideo.VideoStreams, 1)
	assert.Len(t, session.CurrentVideo.AudioStreams, 2)
	assert.Len(t, session.AudioTracks, 2)
	assert.NotNil(t, session.ActiveAudioTrack)
}

func TestVideoPlayerService_loadVideoStreams_Empty(t *testing.T) {
	_, svc := setupVideoPlayerTestDB(t)

	video := &VideoContent{ID: 999}
	session := &VideoPlaybackSession{CurrentVideo: video}
	err := svc.loadVideoStreams(context.Background(), session, 999)
	assert.NoError(t, err)
	assert.Empty(t, session.CurrentVideo.VideoStreams)
}

// =============================================================================
// loadSubtitles
// =============================================================================

func TestVideoPlayerService_loadSubtitles(t *testing.T) {
	db, svc := setupVideoPlayerTestDB(t)

	_, err := db.Exec(`INSERT INTO subtitle_streams (media_item_id, stream_index, codec, language, title, is_default, is_forced) VALUES
		(500, 0, 'srt', 'en', 'English', 1, 0),
		(500, 1, 'srt', 'fr', 'French', 0, 0)`)
	require.NoError(t, err)

	_, err = db.Exec(`INSERT INTO subtitles (media_item_id, language, language_code, format, file_path, encoding, is_embedded, is_default) VALUES
		(500, 'English', 'en', 'srt', '/subs/en.srt', 'utf-8', 0, 1)`)
	require.NoError(t, err)

	video := &VideoContent{ID: 500}
	session := &VideoPlaybackSession{CurrentVideo: video}
	err = svc.loadSubtitles(context.Background(), session, 500)
	assert.NoError(t, err)
}

// =============================================================================
// loadChapters
// =============================================================================

func TestVideoPlayerService_loadChapters(t *testing.T) {
	db, svc := setupVideoPlayerTestDB(t)

	_, err := db.Exec(`INSERT INTO video_chapters (media_item_id, title, start_time, end_time) VALUES
		(600, 'Opening', 0, 120000),
		(600, 'Act 1', 120000, 1800000),
		(600, 'Ending', 1800000, 2400000)`)
	require.NoError(t, err)

	video := &VideoContent{ID: 600}
	session := &VideoPlaybackSession{CurrentVideo: video}
	err = svc.loadChapters(context.Background(), session, 600)
	assert.NoError(t, err)
	assert.Len(t, session.Chapters, 3)
	assert.Equal(t, "Opening", session.Chapters[0].Title)
}

func TestVideoPlayerService_loadChapters_Empty(t *testing.T) {
	_, svc := setupVideoPlayerTestDB(t)

	video := &VideoContent{ID: 999}
	session := &VideoPlaybackSession{CurrentVideo: video}
	err := svc.loadChapters(context.Background(), session, 999)
	assert.NoError(t, err)
	assert.Empty(t, session.Chapters)
}

// =============================================================================
// loadVideoMetadata
// =============================================================================

func TestVideoPlayerService_loadVideoMetadata_Movie(t *testing.T) {
	db, svc := setupVideoPlayerTestDB(t)

	_, err := db.Exec(`INSERT INTO movies (media_item_id, budget, revenue, runtime, collection) VALUES
		(700, 100000, 500000, 120, 'Test Collection')`)
	require.NoError(t, err)

	video := &VideoContent{ID: 700, Type: VideoTypeMovie}
	err = svc.loadVideoMetadata(context.Background(), video)
	assert.NoError(t, err)
	require.NotNil(t, video.MovieInfo)
}

func TestVideoPlayerService_loadVideoMetadata_NoData(t *testing.T) {
	_, svc := setupVideoPlayerTestDB(t)

	// No matching data - should not error (graceful degradation)
	video := &VideoContent{ID: 800, Type: VideoTypeEpisode}
	err := svc.loadVideoMetadata(context.Background(), video)
	assert.NoError(t, err)
}

func TestVideoPlayerService_loadVideoMetadata_OtherType(t *testing.T) {
	_, svc := setupVideoPlayerTestDB(t)

	video := &VideoContent{ID: 800, Type: VideoTypeClip}
	err := svc.loadVideoMetadata(context.Background(), video)
	assert.NoError(t, err)
}

// =============================================================================
// saveVideoSession + recordVideoPlayback
// =============================================================================

func TestVideoPlayerService_saveVideoSession(t *testing.T) {
	_, svc := setupVideoPlayerTestDB(t)

	session := &VideoPlaybackSession{
		ID:            "test-session-1",
		UserID:        1,
		CurrentVideo:  &VideoContent{ID: 100, Title: "Test Video"},
		PlaybackState: PlaybackStatePlaying,
		Position:      5000,
		Duration:      120000,
		LastActivity:  time.Now(),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	err := svc.saveVideoSession(context.Background(), session)
	assert.NoError(t, err)
}

// =============================================================================
// getVideoContent
// =============================================================================

// getVideoContent tested via error path - schema for full column set
// is complex and tested by existing integration tests

func TestVideoPlayerService_getVideoContent_NotFound(t *testing.T) {
	_, svc := setupVideoPlayerTestDB(t)

	_, err := svc.getVideoContent(context.Background(), 999999)
	assert.Error(t, err)
}

// =============================================================================
// NextVideo / PreviousVideo edge cases
// =============================================================================

func TestVideoPlayerService_NextVideo_NoSession(t *testing.T) {
	_, svc := setupVideoPlayerTestDB(t)

	_, err := svc.NextVideo(context.Background(), "nonexistent-session")
	assert.Error(t, err)
}

func TestVideoPlayerService_PreviousVideo_NoSession(t *testing.T) {
	_, svc := setupVideoPlayerTestDB(t)

	_, err := svc.PreviousVideo(context.Background(), "nonexistent-session")
	assert.Error(t, err)
}

// =============================================================================
// generateThumbnail — requires coverArtService, so test error path
// =============================================================================

func TestVideoPlayerService_generateThumbnail_NoVideo(t *testing.T) {
	_, svc := setupVideoPlayerTestDB(t)

	// Video doesn't exist, so getVideoContent will fail
	_, err := svc.generateThumbnail(context.Background(), 99999, 30000)
	assert.Error(t, err)
}
