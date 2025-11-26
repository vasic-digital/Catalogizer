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

func TestMusicPlayerService(t *testing.T) {
	// Setup test database and dependencies
	db := setupTestDB(t)
	defer db.Close()

	logger := zaptest.NewLogger(t)
	mockServer := NewMockServer()
	defer mockServer.Close()

	// Create services with mock dependencies
	translationService := services.NewTranslationService(db, logger, mockServer.URL())
	cacheService := services.NewCacheService(db, logger)
	coverArtService := services.NewCoverArtService(db, logger, mockServer.URL(), cacheService)
	lyricsService := services.NewLyricsService(db, logger, mockServer.URL(), cacheService, translationService)
	subtitleService := services.NewSubtitleService(db, logger, mockServer.URL(), cacheService, translationService)
	positionService := services.NewPlaybackPositionService(db, logger)
	playlistService := services.NewPlaylistService(db, logger)
	mediaPlayerService := services.NewMediaPlayerService(db, logger)

	musicPlayerService := services.NewMusicPlayerService(
		db, logger, mediaPlayerService, playlistService, positionService,
		lyricsService, coverArtService, translationService,
	)

	t.Run("PlayTrack", func(t *testing.T) {
		// Insert test track
		trackID := insertTestTrack(t, db, "Test Artist", "Test Song", "Test Album")

		req := &services.PlayTrackRequest{
			UserID:     1,
			TrackID:    trackID,
			PlayMode:   services.PlayModeTrack,
			Quality:    services.QualityHigh,
			DeviceInfo: services.DeviceInfo{
				DeviceID:   "test-device-1",
				DeviceName: "Test Device",
				DeviceType: "desktop",
				Platform:   "web",
			},
		}

		session, err := musicPlayerService.PlayTrack(context.Background(), req)
		require.NoError(t, err)
		require.NotNil(t, session)

		assert.Equal(t, req.UserID, session.UserID)
		assert.Equal(t, trackID, session.CurrentTrack.ID)
		assert.Equal(t, services.StatePlaying, session.PlaybackState)
		assert.Equal(t, int64(0), session.Position)
		assert.Equal(t, req.Quality, session.PlaybackQuality)
		assert.Len(t, session.Queue, 1)
		assert.Equal(t, 0, session.QueueIndex)
	})

	t.Run("PlayAlbum", func(t *testing.T) {
		// Insert test album with multiple tracks
		albumID := insertTestAlbum(t, db, "Test Artist", "Test Album")
		track1ID := insertTestTrackWithAlbum(t, db, "Test Artist", "Song 1", "Test Album", albumID, 1)
		track2ID := insertTestTrackWithAlbum(t, db, "Test Artist", "Song 2", "Test Album", albumID, 2)
		track3ID := insertTestTrackWithAlbum(t, db, "Test Artist", "Song 3", "Test Album", albumID, 3)

		req := &services.PlayAlbumRequest{
			UserID:     1,
			AlbumID:    albumID,
			Shuffle:    false,
			Quality:    services.QualityHigh,
			DeviceInfo: services.DeviceInfo{
				DeviceID:   "test-device-1",
				DeviceName: "Test Device",
			},
		}

		session, err := musicPlayerService.PlayAlbum(context.Background(), req)
		require.NoError(t, err)
		require.NotNil(t, session)

		assert.Equal(t, req.UserID, session.UserID)
		assert.Equal(t, services.PlayModeAlbum, session.PlayMode)
		assert.Equal(t, services.StatePlaying, session.PlaybackState)
		assert.Len(t, session.Queue, 3)
		assert.Equal(t, 0, session.QueueIndex)
		assert.Equal(t, track1ID, session.CurrentTrack.ID)
	})

	t.Run("UpdatePlayback", func(t *testing.T) {
		// Create a session first
		trackID := insertTestTrack(t, db, "Test Artist", "Update Test", "Test Album")

		playReq := &services.PlayTrackRequest{
			UserID:     1,
			TrackID:    trackID,
			PlayMode:   services.PlayModeTrack,
			Quality:    services.QualityHigh,
			DeviceInfo: services.DeviceInfo{DeviceID: "test-device-1"},
		}

		session, err := musicPlayerService.PlayTrack(context.Background(), playReq)
		require.NoError(t, err)

		// Update playback
		updateReq := &services.UpdatePlaybackRequest{
			SessionID: session.ID,
			Position:  int64Ptr(30000), // 30 seconds
			State:     &[]services.PlaybackState{services.StatePaused}[0],
			Volume:    float64Ptr(0.8),
			IsMuted:   boolPtr(false),
		}

		updatedSession, err := musicPlayerService.UpdatePlayback(context.Background(), updateReq)
		require.NoError(t, err)
		require.NotNil(t, updatedSession)

		assert.Equal(t, int64(30000), updatedSession.Position)
		assert.Equal(t, services.StatePaused, updatedSession.PlaybackState)
		assert.Equal(t, 0.8, updatedSession.Volume)
		assert.Equal(t, false, updatedSession.IsMuted)
	})

	t.Run("NextTrack", func(t *testing.T) {
		// Create album with multiple tracks
		albumID := insertTestAlbum(t, db, "Test Artist", "Next Test Album")
		track1ID := insertTestTrackWithAlbum(t, db, "Test Artist", "Song 1", "Next Test Album", albumID, 1)
		track2ID := insertTestTrackWithAlbum(t, db, "Test Artist", "Song 2", "Next Test Album", albumID, 2)

		req := &services.PlayAlbumRequest{
			UserID:     1,
			AlbumID:    albumID,
			Quality:    services.QualityHigh,
			DeviceInfo: services.DeviceInfo{DeviceID: "test-device-1"},
		}

		session, err := musicPlayerService.PlayAlbum(context.Background(), req)
		require.NoError(t, err)

		// Skip to next track
		nextSession, err := musicPlayerService.NextTrack(context.Background(), session.ID)
		require.NoError(t, err)
		require.NotNil(t, nextSession)

		assert.Equal(t, 1, nextSession.QueueIndex)
		assert.Equal(t, track2ID, nextSession.CurrentTrack.ID)
		assert.Equal(t, int64(0), nextSession.Position)
	})

	t.Run("AddToQueue", func(t *testing.T) {
		// Create initial session
		trackID := insertTestTrack(t, db, "Test Artist", "Queue Test", "Test Album")

		playReq := &services.PlayTrackRequest{
			UserID:     1,
			TrackID:    trackID,
			PlayMode:   services.PlayModeTrack,
			Quality:    services.QualityHigh,
			DeviceInfo: services.DeviceInfo{DeviceID: "test-device-1"},
		}

		session, err := musicPlayerService.PlayTrack(context.Background(), playReq)
		require.NoError(t, err)

		// Add more tracks to queue
		track2ID := insertTestTrack(t, db, "Test Artist 2", "Queue Test 2", "Test Album 2")
		track3ID := insertTestTrack(t, db, "Test Artist 3", "Queue Test 3", "Test Album 3")

		queueReq := &services.QueueRequest{
			SessionID: session.ID,
			TrackIDs:  []int64{track2ID, track3ID},
		}

		updatedSession, err := musicPlayerService.AddToQueue(context.Background(), queueReq)
		require.NoError(t, err)
		require.NotNil(t, updatedSession)

		assert.Len(t, updatedSession.Queue, 3)
		assert.Equal(t, trackID, updatedSession.Queue[0].ID)
		assert.Equal(t, track2ID, updatedSession.Queue[1].ID)
		assert.Equal(t, track3ID, updatedSession.Queue[2].ID)
	})

	t.Run("GetLibraryStats", func(t *testing.T) {
		// Insert test data
		insertTestTrack(t, db, "Artist 1", "Song 1", "Album 1")
		insertTestTrack(t, db, "Artist 1", "Song 2", "Album 1")
		insertTestTrack(t, db, "Artist 2", "Song 3", "Album 2")

		stats, err := musicPlayerService.GetLibraryStats(context.Background(), 1)
		require.NoError(t, err)
		require.NotNil(t, stats)

		assert.True(t, stats.TotalTracks >= 3)
		assert.True(t, stats.TotalAlbums >= 2)
		assert.True(t, stats.TotalArtists >= 2)
	})
}

func TestVideoPlayerService(t *testing.T) {
	// Setup test database and dependencies
	db := setupTestDB(t)
	defer db.Close()

	logger := zaptest.NewLogger(t)
	mockServer := NewMockServer()
	defer mockServer.Close()

	// Create services
	translationService := services.NewTranslationService(db, logger, mockServer.URL())
	cacheService := services.NewCacheService(db, logger)
	subtitleService := services.NewSubtitleService(db, logger, mockServer.URL(), cacheService, translationService)
	coverArtService := services.NewCoverArtService(db, logger, mockServer.URL(), cacheService)
	positionService := services.NewPlaybackPositionService(db, logger)
	mediaPlayerService := services.NewMediaPlayerService(db, logger)

	videoPlayerService := services.NewVideoPlayerService(
		db, logger, mediaPlayerService, positionService,
		subtitleService, coverArtService, translationService,
	)

	t.Run("PlayVideo", func(t *testing.T) {
		// Insert test video
		videoID := insertTestVideo(t, db, "Test Movie", services.VideoTypeMovie)

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

		assert.Equal(t, req.UserID, session.UserID)
		assert.Equal(t, videoID, session.CurrentVideo.ID)
		assert.Equal(t, services.StatePlaying, session.PlaybackState)
		assert.Equal(t, int64(0), session.Position)
		assert.Equal(t, req.Quality, session.VideoQuality)
	})

	t.Run("UpdateVideoPlayback", func(t *testing.T) {
		// Create video session
		videoID := insertTestVideo(t, db, "Update Test Movie", services.VideoTypeMovie)

		playReq := &services.PlayVideoRequest{
			UserID:     1,
			VideoID:    videoID,
			PlayMode:   services.VideoPlayModeSingle,
			Quality:    services.Quality1080p,
			DeviceInfo: services.DeviceInfo{DeviceID: "test-device-1"},
		}

		session, err := videoPlayerService.PlayVideo(context.Background(), playReq)
		require.NoError(t, err)

		// Update playback
		updateReq := &services.UpdateVideoPlaybackRequest{
			SessionID:     session.ID,
			Position:      int64Ptr(120000), // 2 minutes
			State:         &[]services.PlaybackState{services.StatePaused}[0],
			Volume:        float64Ptr(0.9),
			PlaybackSpeed: float64Ptr(1.25),
			Quality:       &[]services.VideoQuality{services.Quality720p}[0],
		}

		updatedSession, err := videoPlayerService.UpdateVideoPlayback(context.Background(), updateReq)
		require.NoError(t, err)
		require.NotNil(t, updatedSession)

		assert.Equal(t, int64(120000), updatedSession.Position)
		assert.Equal(t, services.StatePaused, updatedSession.PlaybackState)
		assert.Equal(t, 0.9, updatedSession.Volume)
		assert.Equal(t, 1.25, updatedSession.PlaybackSpeed)
		assert.Equal(t, services.Quality720p, updatedSession.VideoQuality)
	})

	t.Run("CreateVideoBookmark", func(t *testing.T) {
		// Create video session
		videoID := insertTestVideo(t, db, "Bookmark Test Movie", services.VideoTypeMovie)

		playReq := &services.PlayVideoRequest{
			UserID:     1,
			VideoID:    videoID,
			PlayMode:   services.VideoPlayModeSingle,
			Quality:    services.Quality1080p,
			DeviceInfo: services.DeviceInfo{DeviceID: "test-device-1"},
		}

		session, err := videoPlayerService.PlayVideo(context.Background(), playReq)
		require.NoError(t, err)

		// Seek to bookmark position
		seekReq := &services.VideoSeekRequest{
			SessionID: session.ID,
			Position:  180000, // 3 minutes
		}

		session, err = videoPlayerService.SeekVideo(context.Background(), seekReq)
		require.NoError(t, err)

		// Create bookmark
		bookmarkReq := &services.CreateVideoBookmarkRequest{
			SessionID:   session.ID,
			Title:       "Test Bookmark",
			Description: "This is a test bookmark",
		}

		bookmark, err := videoPlayerService.CreateVideoBookmark(context.Background(), bookmarkReq)
		require.NoError(t, err)
		require.NotNil(t, bookmark)

		assert.Equal(t, int64(1), bookmark.UserID)
		assert.Equal(t, videoID, bookmark.VideoID)
		assert.Equal(t, int64(180000), bookmark.Position)
		assert.Equal(t, "Test Bookmark", bookmark.Title)
		assert.Equal(t, "This is a test bookmark", bookmark.Description)
	})

	t.Run("GetContinueWatching", func(t *testing.T) {
		// Insert test videos and positions
		video1ID := insertTestVideo(t, db, "Continue Movie 1", services.VideoTypeMovie)
		video2ID := insertTestVideo(t, db, "Continue Movie 2", services.VideoTypeMovie)

		// Set partial watch positions
		positionService.UpdatePosition(context.Background(), &services.UpdatePositionRequest{
			UserID:      1,
			MediaItemID: video1ID,
			Position:    600000,  // 10 minutes
			Duration:    7200000, // 2 hours
			DeviceInfo:  "test-device",
		})

		positionService.UpdatePosition(context.Background(), &services.UpdatePositionRequest{
			UserID:      1,
			MediaItemID: video2ID,
			Position:    1800000, // 30 minutes
			Duration:    5400000, // 1.5 hours
			DeviceInfo:  "test-device",
		})

		videos, err := videoPlayerService.GetContinueWatching(context.Background(), 1, 10)
		require.NoError(t, err)

		assert.True(t, len(videos) >= 2)

		// Verify videos are in the continue watching list
		videoIDs := make([]int64, len(videos))
		for i, video := range videos {
			videoIDs[i] = video.ID
		}
		assert.Contains(t, videoIDs, video1ID)
		assert.Contains(t, videoIDs, video2ID)
	})
}

func TestPlaylistService(t *testing.T) {
	// Setup test database and dependencies
	db := setupTestDB(t)
	defer db.Close()

	logger := zaptest.NewLogger(t)
	playlistService := services.NewPlaylistService(db, logger)

	t.Run("CreatePlaylist", func(t *testing.T) {
		req := &services.CreatePlaylistRequest{
			UserID:      1,
			Name:        "Test Playlist",
			Description: "This is a test playlist",
			IsPublic:    false,
			Tags:        []string{"test", "rock", "favorites"},
		}

		playlist, err := playlistService.CreatePlaylist(context.Background(), req)
		require.NoError(t, err)
		require.NotNil(t, playlist)

		assert.Equal(t, req.UserID, playlist.UserID)
		assert.Equal(t, req.Name, playlist.Name)
		assert.Equal(t, req.Description, playlist.Description)
		assert.Equal(t, req.IsPublic, playlist.IsPublic)
		assert.Equal(t, 0, playlist.TrackCount)
		assert.Equal(t, int64(0), playlist.TotalDuration)
	})

	t.Run("AddToPlaylist", func(t *testing.T) {
		// Create playlist
		createReq := &services.CreatePlaylistRequest{
			UserID: 1,
			Name:   "Add Test Playlist",
		}

		playlist, err := playlistService.CreatePlaylist(context.Background(), createReq)
		require.NoError(t, err)

		// Create test tracks
		track1ID := insertTestTrack(t, db, "Artist 1", "Song 1", "Album 1")
		track2ID := insertTestTrack(t, db, "Artist 2", "Song 2", "Album 2")

		// Add tracks to playlist
		addReq := &services.AddToPlaylistRequest{
			PlaylistID:   playlist.ID,
			MediaItemIDs: []int64{track1ID, track2ID},
			UserID:       1,
		}

		err = playlistService.AddToPlaylist(context.Background(), addReq)
		require.NoError(t, err)

		// Verify items were added
		items, err := playlistService.GetPlaylistItems(context.Background(), playlist.ID, 1, 10, 0)
		require.NoError(t, err)

		assert.Len(t, items, 2)
		assert.Equal(t, track1ID, items[0].MediaItemID)
		assert.Equal(t, track2ID, items[1].MediaItemID)
		assert.Equal(t, 1, items[0].Position)
		assert.Equal(t, 2, items[1].Position)
	})

	t.Run("CreateSmartPlaylist", func(t *testing.T) {
		// Create smart playlist with criteria
		criteria := &services.SmartPlaylistCriteria{
			Rules: []services.SmartRule{
				{
					Field:    "genre",
					Operator: "equals",
					Value:    "rock",
				},
				{
					Field:    "year",
					Operator: "greater_than",
					Value:    2020,
				},
			},
			Logic: "AND",
			Limit: 50,
			Order: "play_count_desc",
		}

		req := &services.CreatePlaylistRequest{
			UserID:          1,
			Name:            "Smart Rock Playlist",
			Description:     "Automatically updated rock playlist",
			IsSmartPlaylist: true,
			SmartCriteria:   criteria,
		}

		playlist, err := playlistService.CreatePlaylist(context.Background(), req)
		require.NoError(t, err)
		require.NotNil(t, playlist)

		assert.Equal(t, true, playlist.IsSmartPlaylist)
		assert.NotEmpty(t, playlist.SmartCriteria)
	})

	t.Run("ReorderPlaylist", func(t *testing.T) {
		// Create playlist with tracks
		createReq := &services.CreatePlaylistRequest{
			UserID: 1,
			Name:   "Reorder Test Playlist",
		}

		playlist, err := playlistService.CreatePlaylist(context.Background(), createReq)
		require.NoError(t, err)

		// Add tracks
		track1ID := insertTestTrack(t, db, "Artist 1", "Song A", "Album 1")
		track2ID := insertTestTrack(t, db, "Artist 2", "Song B", "Album 2")
		track3ID := insertTestTrack(t, db, "Artist 3", "Song C", "Album 3")

		addReq := &services.AddToPlaylistRequest{
			PlaylistID:   playlist.ID,
			MediaItemIDs: []int64{track1ID, track2ID, track3ID},
			UserID:       1,
		}

		err = playlistService.AddToPlaylist(context.Background(), addReq)
		require.NoError(t, err)

		// Get items to find item IDs
		items, err := playlistService.GetPlaylistItems(context.Background(), playlist.ID, 1, 10, 0)
		require.NoError(t, err)
		require.Len(t, items, 3)

		// Move first item to last position
		reorderReq := &services.ReorderPlaylistRequest{
			PlaylistID:  playlist.ID,
			UserID:      1,
			ItemID:      items[0].ID,
			NewPosition: 3,
		}

		err = playlistService.ReorderPlaylist(context.Background(), reorderReq)
		require.NoError(t, err)

		// Verify new order
		reorderedItems, err := playlistService.GetPlaylistItems(context.Background(), playlist.ID, 1, 10, 0)
		require.NoError(t, err)

		assert.Equal(t, track2ID, reorderedItems[0].MediaItemID)
		assert.Equal(t, track3ID, reorderedItems[1].MediaItemID)
		assert.Equal(t, track1ID, reorderedItems[2].MediaItemID)
	})
}

func TestPositionService(t *testing.T) {
	// Setup test database
	db := setupTestDB(t)
	defer db.Close()

	logger := zaptest.NewLogger(t)
	positionService := services.NewPlaybackPositionService(db, logger)

	t.Run("UpdatePosition", func(t *testing.T) {
		trackID := insertTestTrack(t, db, "Position Artist", "Position Song", "Position Album")

		req := &services.UpdatePositionRequest{
			UserID:          1,
			MediaItemID:     trackID,
			Position:        45000,  // 45 seconds
			Duration:        180000, // 3 minutes
			DeviceInfo:      "test-device",
			PlaybackQuality: "high",
		}

		err := positionService.UpdatePosition(context.Background(), req)
		require.NoError(t, err)

		// Verify position was saved
		position, err := positionService.GetPosition(context.Background(), 1, trackID)
		require.NoError(t, err)
		require.NotNil(t, position)

		assert.Equal(t, int64(1), position.UserID)
		assert.Equal(t, trackID, position.MediaItemID)
		assert.Equal(t, int64(45000), position.Position)
		assert.Equal(t, int64(180000), position.Duration)
		assert.Equal(t, float64(25), position.PercentComplete)
	})

	t.Run("CreateBookmark", func(t *testing.T) {
		trackID := insertTestTrack(t, db, "Bookmark Artist", "Bookmark Song", "Bookmark Album")

		req := &services.BookmarkRequest{
			UserID:      1,
			MediaItemID: trackID,
			Position:    90000, // 1.5 minutes
			Name:        "Guitar Solo",
			Description: "Amazing guitar solo starts here",
		}

		bookmark, err := positionService.CreateBookmark(context.Background(), req)
		require.NoError(t, err)
		require.NotNil(t, bookmark)

		assert.Equal(t, req.UserID, bookmark.UserID)
		assert.Equal(t, trackID, bookmark.MediaItemID)
		assert.Equal(t, int64(90000), bookmark.Position)
		assert.Equal(t, "Guitar Solo", bookmark.Name)
		assert.Equal(t, "Amazing guitar solo starts here", bookmark.Description)
	})

	t.Run("GetContinueWatching", func(t *testing.T) {
		// Create multiple tracks with different completion percentages
		track1ID := insertTestTrack(t, db, "Artist 1", "Song 1", "Album 1")
		track2ID := insertTestTrack(t, db, "Artist 2", "Song 2", "Album 2")
		track3ID := insertTestTrack(t, db, "Artist 3", "Song 3", "Album 3")

		// Track 1: 25% complete (should appear in continue watching)
		positionService.UpdatePosition(context.Background(), &services.UpdatePositionRequest{
			UserID:      1,
			MediaItemID: track1ID,
			Position:    45000,
			Duration:    180000,
			DeviceInfo:  "test-device",
		})

		// Track 2: 50% complete (should appear in continue watching)
		positionService.UpdatePosition(context.Background(), &services.UpdatePositionRequest{
			UserID:      1,
			MediaItemID: track2ID,
			Position:    90000,
			Duration:    180000,
			DeviceInfo:  "test-device",
		})

		// Track 3: 95% complete (should NOT appear in continue watching)
		positionService.UpdatePosition(context.Background(), &services.UpdatePositionRequest{
			UserID:      1,
			MediaItemID: track3ID,
			Position:    171000,
			Duration:    180000,
			DeviceInfo:  "test-device",
		})

		positions, err := positionService.GetContinueWatching(context.Background(), 1, 10)
		require.NoError(t, err)

		// Should only get tracks 1 and 2 (track 3 is >90% complete)
		assert.Len(t, positions, 2)

		mediaIDs := make([]int64, len(positions))
		for i, pos := range positions {
			mediaIDs[i] = pos.MediaItemID
		}
		assert.Contains(t, mediaIDs, track1ID)
		assert.Contains(t, mediaIDs, track2ID)
		assert.NotContains(t, mediaIDs, track3ID)
	})

	t.Run("GetPlaybackStats", func(t *testing.T) {
		// Create test data
		track1ID := insertTestTrack(t, db, "Stats Artist 1", "Stats Song 1", "Stats Album 1")
		track2ID := insertTestTrack(t, db, "Stats Artist 2", "Stats Song 2", "Stats Album 2")

		// Record some playback
		positionService.UpdatePosition(context.Background(), &services.UpdatePositionRequest{
			UserID:      1,
			MediaItemID: track1ID,
			Position:    180000,
			Duration:    180000,
			DeviceInfo:  "test-device",
		})

		positionService.UpdatePosition(context.Background(), &services.UpdatePositionRequest{
			UserID:      1,
			MediaItemID: track2ID,
			Position:    90000,
			Duration:    180000,
			DeviceInfo:  "test-device",
		})

		req := &services.PlaybackStatsRequest{
			UserID: 1,
			Limit:  10,
		}

		stats, err := positionService.GetPlaybackStats(context.Background(), req)
		require.NoError(t, err)
		require.NotNil(t, stats)

		assert.True(t, stats.TotalMediaItems >= 2)
		assert.True(t, stats.CompletedItems >= 1)
	})
}

// Helper functions

func setupTestDB(t *testing.T) *sql.DB {
	// This would typically connect to a test database
	// For this example, we'll assume the database is already set up
	// In practice, you might use an in-memory SQLite database or Docker container
	db, err := sql.Open("postgres", "postgres://test:test@localhost/catalogizer_test?sslmode=disable")
	require.NoError(t, err)
	return db
}

func insertTestTrack(t *testing.T, db *sql.DB, artist, title, album string) int64 {
	query := `
		INSERT INTO media_items (type, title, artist, album, duration, file_path, date_added)
		VALUES ('audio', $1, $2, $3, 180000, '/test/path.mp3', NOW())
		RETURNING id
	`

	var id int64
	err := db.QueryRow(query, title, artist, album).Scan(&id)
	require.NoError(t, err)
	return id
}

func insertTestTrackWithAlbum(t *testing.T, db *sql.DB, artist, title, album string, albumID int64, trackNumber int) int64 {
	query := `
		INSERT INTO media_items (type, title, artist, album, album_id, track_number, duration, file_path, date_added)
		VALUES ('audio', $1, $2, $3, $4, $5, 180000, '/test/path.mp3', NOW())
		RETURNING id
	`

	var id int64
	err := db.QueryRow(query, title, artist, album, albumID, trackNumber).Scan(&id)
	require.NoError(t, err)
	return id
}

func insertTestAlbum(t *testing.T, db *sql.DB, artist, title string) int64 {
	query := `
		INSERT INTO albums (title, artist, created_at)
		VALUES ($1, $2, NOW())
		RETURNING id
	`

	var id int64
	err := db.QueryRow(query, title, artist).Scan(&id)
	require.NoError(t, err)
	return id
}

func insertTestVideo(t *testing.T, db *sql.DB, title string, videoType services.VideoType) int64 {
	query := `
		INSERT INTO media_items (type, title, duration, file_path, date_added)
		VALUES ('video', $1, 7200000, '/test/video.mp4', NOW())
		RETURNING id
	`

	var id int64
	err := db.QueryRow(query, title).Scan(&id)
	require.NoError(t, err)
	return id
}

// Helper function pointers
func int64Ptr(v int64) *int64 {
	return &v
}

func float64Ptr(v float64) *float64 {
	return &v
}

func boolPtr(v bool) *bool {
	return &v
}