package tests

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"catalogizer/internal/services"
	_ "github.com/mattn/go-sqlite3"
)

func TestVideoPlayerSubtitleHandling(t *testing.T) {
	// Setup test database and dependencies
	db := setupTestDB(t)
	defer db.Close()

	logger := zaptest.NewLogger(t)
	mockServer := NewMockServer()
	defer mockServer.Close()

	// Create services with correct signatures
	translationService := services.NewTranslationService(logger)
	cacheService := services.NewCacheService(db, logger)
	subtitleService := services.NewSubtitleService(db, logger, cacheService)
	coverArtService := services.NewCoverArtService(db, logger)
	positionService := services.NewPlaybackPositionService(db, logger)
	mediaPlayerService := services.NewMediaPlayerService(db, logger)

	videoPlayerService := services.NewVideoPlayerService(
		db, logger, mediaPlayerService, positionService,
		subtitleService, coverArtService, translationService,
	)

	t.Run("LoadVideoWithDefaultSubtitle", func(t *testing.T) {
		// Insert test video
		videoID := insertTestVideoForSubtitle(t, db, "Test Movie", "movie")

		// Insert test subtitle tracks
		_ = insertTestSubtitleTrack(t, db, videoID, "en", true, false)  // Default subtitle
		_ = insertTestSubtitleTrack(t, db, videoID, "es", false, false) // Non-default subtitle
		_ = insertTestSubtitleTrack(t, db, videoID, "fr", false, true)  // Forced subtitle

		// Play video
		req := &services.PlayVideoRequest{
			UserID:     1,
			VideoID:    videoID,
			PlayMode:   services.VideoPlayModeSingle,
			Quality:    services.Quality1080p,
			DeviceInfo: services.DeviceInfo{
				DeviceID:   "test-device-1",
				DeviceName: "Test Device",
			},
		}

		session, err := videoPlayerService.PlayVideo(context.Background(), req)
		require.NoError(t, err)
		require.NotNil(t, session)

		// Verify subtitle tracks are loaded
		require.NotEmpty(t, session.SubtitleTracks)
		assert.Len(t, session.SubtitleTracks, 3) // All 3 subtitles should be loaded

		// Verify default subtitle is set (using track index as identifier)
		require.NotNil(t, session.ActiveSubtitle)
		assert.Equal(t, int64(0), *session.ActiveSubtitle) // First track (default subtitle) should be active

		// Verify the active subtitle is the English one (default)
		activeTrack := session.SubtitleTracks[*session.ActiveSubtitle]
		assert.Equal(t, "en", activeTrack.Language)
		assert.True(t, activeTrack.IsDefault)
		assert.False(t, activeTrack.IsForced)
	})

	t.Run("LoadVideoWithNoDefaultSubtitle", func(t *testing.T) {
		// Insert test video
		videoID := insertTestVideoForSubtitle(t, db, "Test Movie 2", "movie")

		// Insert test subtitle tracks with no default
		_ = insertTestSubtitleTrack(t, db, videoID, "en", false, false)
		_ = insertTestSubtitleTrack(t, db, videoID, "es", false, false)

		// Play video
		req := &services.PlayVideoRequest{
			UserID:     1,
			VideoID:    videoID,
			PlayMode:   services.VideoPlayModeSingle,
			Quality:    services.Quality1080p,
			DeviceInfo: services.DeviceInfo{
				DeviceID:   "test-device-2",
				DeviceName: "Test Device 2",
			},
		}

		session, err := videoPlayerService.PlayVideo(context.Background(), req)
		require.NoError(t, err)
		require.NotNil(t, session)

		// Verify subtitle tracks are loaded
		require.NotEmpty(t, session.SubtitleTracks)
		assert.Len(t, session.SubtitleTracks, 2) // Both subtitles should be loaded

		// Verify no default subtitle is set (no track marked as default)
		assert.Nil(t, session.ActiveSubtitle)
	})

	t.Run("LoadVideoWithForcedSubtitle", func(t *testing.T) {
		// Insert test video
		videoID := insertTestVideoForSubtitle(t, db, "Test Movie 3", "movie")

		// Insert test subtitle tracks with forced subtitle as first track
		_ = insertTestSubtitleTrack(t, db, videoID, "fr", false, true)  // Forced subtitle
		_ = insertTestSubtitleTrack(t, db, videoID, "en", true, false)  // Default subtitle

		// Play video
		req := &services.PlayVideoRequest{
			UserID:     1,
			VideoID:    videoID,
			PlayMode:   services.VideoPlayModeSingle,
			Quality:    services.Quality1080p,
			DeviceInfo: services.DeviceInfo{
				DeviceID:   "test-device-3",
				DeviceName: "Test Device 3",
			},
		}

		session, err := videoPlayerService.PlayVideo(context.Background(), req)
		require.NoError(t, err)
		require.NotNil(t, session)

		// Verify subtitle tracks are loaded
		require.NotEmpty(t, session.SubtitleTracks)
		assert.Len(t, session.SubtitleTracks, 2) // Both subtitles should be loaded

		// Verify the default subtitle is set, not the forced one
		require.NotNil(t, session.ActiveSubtitle)
		assert.Equal(t, int64(1), *session.ActiveSubtitle) // Second track (default subtitle) should be active

		// Verify the active subtitle is the English one (default)
		activeTrack := session.SubtitleTracks[*session.ActiveSubtitle]
		assert.Equal(t, "en", activeTrack.Language)
		assert.True(t, activeTrack.IsDefault)
		assert.False(t, activeTrack.IsForced)
	})
}

// Helper function to insert test subtitle track
func insertTestSubtitleTrack(t *testing.T, db *sql.DB, videoID int64, language string, isDefault, isForced bool) int64 {
	query := `
		INSERT INTO subtitle_tracks (
			media_item_id, language, language_code, source, format, path, 
			is_default, is_forced, encoding, sync_offset, created_at, verified_sync
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	
	result, err := db.Exec(query,
		videoID,            // media_item_id
		language,            // language
		language,            // language_code
		"embedded",           // source
		"srt",               // format
		"/path/to/subtitle.srt", // path
		isDefault,           // is_default
		isForced,            // is_forced
		"utf-8",             // encoding
		0.0,                 // sync_offset
		time.Now(),           // created_at
		true,                 // verified_sync
	)
	require.NoError(t, err)

	id, err := result.LastInsertId()
	require.NoError(t, err)

	return id
}

// Helper function to insert test video for subtitle tests
func insertTestVideoForSubtitle(t *testing.T, db *sql.DB, title string, videoType string) int64 {
	query := `
		INSERT INTO media_items (
			title, type, duration, file_size, file_path, 
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	
	result, err := db.Exec(query,
		title,               // title
		videoType,           // type
		7200,                // duration (2 hours)
		1073741824,          // file_size (1GB)
		"/path/to/video.mp4", // file_path
		time.Now(),           // created_at
		time.Now(),           // updated_at
	)
	require.NoError(t, err)

	id, err := result.LastInsertId()
	require.NoError(t, err)

	return id
}