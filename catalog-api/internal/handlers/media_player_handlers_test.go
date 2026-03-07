package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

type MediaPlayerHandlersTestSuite struct {
	suite.Suite
	handlers *MediaPlayerHandlers
	router   *mux.Router
	logger   *zap.Logger
}

func (suite *MediaPlayerHandlersTestSuite) SetupSuite() {
	suite.logger = zap.NewNop()
}

func (suite *MediaPlayerHandlersTestSuite) SetupTest() {
	// Initialize handlers with nil services to test validation paths
	suite.handlers = NewMediaPlayerHandlers(
		suite.logger,
		nil, // musicPlayerService
		nil, // videoPlayerService
		nil, // playlistService
		nil, // positionService
		nil, // subtitleService
		nil, // lyricsService
		nil, // coverArtService
		nil, // translationService
	)

	suite.router = mux.NewRouter()
	suite.handlers.RegisterRoutes(suite.router)
}

// Constructor tests

func (suite *MediaPlayerHandlersTestSuite) TestNewMediaPlayerHandlers() {
	handlers := NewMediaPlayerHandlers(suite.logger, nil, nil, nil, nil, nil, nil, nil, nil)
	assert.NotNil(suite.T(), handlers)
	assert.NotNil(suite.T(), handlers.logger)
	assert.Nil(suite.T(), handlers.musicPlayerService)
	assert.Nil(suite.T(), handlers.videoPlayerService)
	assert.Nil(suite.T(), handlers.playlistService)
	assert.Nil(suite.T(), handlers.positionService)
	assert.Nil(suite.T(), handlers.subtitleService)
	assert.Nil(suite.T(), handlers.lyricsService)
	assert.Nil(suite.T(), handlers.coverArtService)
	assert.Nil(suite.T(), handlers.translationService)
}

// PlayMusic tests

func (suite *MediaPlayerHandlersTestSuite) TestPlayMusic_InvalidBody() {
	req := httptest.NewRequest("POST", "/api/v1/music/play", bytes.NewBufferString("not-json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	var resp APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), resp.Success)
	assert.Equal(suite.T(), "Invalid request body", resp.Error)
}

// PlayAlbum tests

func (suite *MediaPlayerHandlersTestSuite) TestPlayAlbum_InvalidBody() {
	req := httptest.NewRequest("POST", "/api/v1/music/play/album", bytes.NewBufferString("invalid"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

// PlayArtist tests

func (suite *MediaPlayerHandlersTestSuite) TestPlayArtist_InvalidBody() {
	req := httptest.NewRequest("POST", "/api/v1/music/play/artist", bytes.NewBufferString("invalid"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

// PlayVideo tests

func (suite *MediaPlayerHandlersTestSuite) TestPlayVideo_InvalidBody() {
	req := httptest.NewRequest("POST", "/api/v1/video/play", bytes.NewBufferString("invalid"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

// PlaySeries tests

func (suite *MediaPlayerHandlersTestSuite) TestPlaySeries_InvalidBody() {
	req := httptest.NewRequest("POST", "/api/v1/video/play/series", bytes.NewBufferString("invalid"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

// CreatePlaylist tests

func (suite *MediaPlayerHandlersTestSuite) TestCreatePlaylist_InvalidBody() {
	req := httptest.NewRequest("POST", "/api/v1/playlists", bytes.NewBufferString("invalid"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

// GetUserPlaylists tests

func (suite *MediaPlayerHandlersTestSuite) TestGetUserPlaylists_Unauthenticated() {
	req := httptest.NewRequest("GET", "/api/v1/playlists", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
	var resp APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "User not authenticated", resp.Error)
}

// GetMusicLibraryStats tests

func (suite *MediaPlayerHandlersTestSuite) TestGetMusicLibraryStats_Unauthenticated() {
	req := httptest.NewRequest("GET", "/api/v1/music/library/stats", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
}

// GetContinueWatching tests

func (suite *MediaPlayerHandlersTestSuite) TestGetContinueWatching_Unauthenticated() {
	req := httptest.NewRequest("GET", "/api/v1/video/continue-watching", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
}

// GetWatchHistory tests

func (suite *MediaPlayerHandlersTestSuite) TestGetWatchHistory_Unauthenticated() {
	req := httptest.NewRequest("GET", "/api/v1/video/watch-history", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
}

// SearchSubtitles tests

func (suite *MediaPlayerHandlersTestSuite) TestSearchSubtitles_InvalidBody() {
	req := httptest.NewRequest("POST", "/api/v1/subtitles/search", bytes.NewBufferString("invalid"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

// SearchLyrics tests

func (suite *MediaPlayerHandlersTestSuite) TestSearchLyrics_InvalidBody() {
	req := httptest.NewRequest("POST", "/api/v1/lyrics/search", bytes.NewBufferString("invalid"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

// TranslateText tests

func (suite *MediaPlayerHandlersTestSuite) TestTranslateText_InvalidBody() {
	req := httptest.NewRequest("POST", "/api/v1/translate", bytes.NewBufferString("invalid"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

// DetectLanguage tests

func (suite *MediaPlayerHandlersTestSuite) TestDetectLanguage_InvalidBody() {
	req := httptest.NewRequest("POST", "/api/v1/translate/detect", bytes.NewBufferString("invalid"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

// UpdatePlaybackPosition tests

func (suite *MediaPlayerHandlersTestSuite) TestUpdatePlaybackPosition_InvalidBody() {
	req := httptest.NewRequest("POST", "/api/v1/playback/position", bytes.NewBufferString("invalid"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

// GetPlaybackStats tests

func (suite *MediaPlayerHandlersTestSuite) TestGetPlaybackStats_Unauthenticated() {
	req := httptest.NewRequest("GET", "/api/v1/playback/stats", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
}

// GetContinueWatchingList tests

func (suite *MediaPlayerHandlersTestSuite) TestGetContinueWatchingList_Unauthenticated() {
	req := httptest.NewRequest("GET", "/api/v1/playback/continue-watching", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
}

// CreateBookmark tests

func (suite *MediaPlayerHandlersTestSuite) TestCreateBookmark_InvalidBody() {
	req := httptest.NewRequest("POST", "/api/v1/playback/bookmarks", bytes.NewBufferString("invalid"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

// Helper method tests

func (suite *MediaPlayerHandlersTestSuite) TestGetUserID_NoContext() {
	req := httptest.NewRequest("GET", "/test", nil)
	userID := suite.handlers.getUserID(req)
	assert.Equal(suite.T(), int64(0), userID)
}

func (suite *MediaPlayerHandlersTestSuite) TestGetUserID_WithInt64Context() {
	req := httptest.NewRequest("GET", "/test", nil)
	ctx := context.WithValue(req.Context(), "user_id", int64(42))
	req = req.WithContext(ctx)
	userID := suite.handlers.getUserID(req)
	assert.Equal(suite.T(), int64(42), userID)
}

func (suite *MediaPlayerHandlersTestSuite) TestGetUserID_WithIntContext() {
	req := httptest.NewRequest("GET", "/test", nil)
	ctx := context.WithValue(req.Context(), "user_id", 42)
	req = req.WithContext(ctx)
	userID := suite.handlers.getUserID(req)
	assert.Equal(suite.T(), int64(42), userID)
}

func (suite *MediaPlayerHandlersTestSuite) TestGetUserID_WithStringContext() {
	req := httptest.NewRequest("GET", "/test", nil)
	ctx := context.WithValue(req.Context(), "user_id", "42")
	req = req.WithContext(ctx)
	userID := suite.handlers.getUserID(req)
	assert.Equal(suite.T(), int64(0), userID)
}

// sendSuccess and sendError tests

func (suite *MediaPlayerHandlersTestSuite) TestSendSuccess() {
	w := httptest.NewRecorder()
	suite.handlers.sendSuccess(w, map[string]string{"key": "value"})

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	assert.Equal(suite.T(), "application/json", w.Header().Get("Content-Type"))
	var resp APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), resp.Success)
}

func (suite *MediaPlayerHandlersTestSuite) TestSendError() {
	w := httptest.NewRecorder()
	suite.handlers.sendError(w, "test error", http.StatusBadRequest)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	assert.Equal(suite.T(), "application/json", w.Header().Get("Content-Type"))
	var resp APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), resp.Success)
	assert.Equal(suite.T(), "test error", resp.Error)
}

func (suite *MediaPlayerHandlersTestSuite) TestSendError_InternalServerError() {
	w := httptest.NewRecorder()
	suite.handlers.sendError(w, "internal error", http.StatusInternalServerError)

	assert.Equal(suite.T(), http.StatusInternalServerError, w.Code)
}

// UpdateMusicPlayback tests

func (suite *MediaPlayerHandlersTestSuite) TestUpdateMusicPlayback_InvalidBody() {
	req := httptest.NewRequest("POST", "/api/v1/music/session/test-session/update", bytes.NewBufferString("invalid"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	var resp APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), resp.Success)
	assert.Equal(suite.T(), "Invalid request body", resp.Error)
}

// SeekMusic tests

func (suite *MediaPlayerHandlersTestSuite) TestSeekMusic_InvalidBody() {
	req := httptest.NewRequest("POST", "/api/v1/music/session/test-session/seek", bytes.NewBufferString("invalid"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

// AddToMusicQueue tests

func (suite *MediaPlayerHandlersTestSuite) TestAddToMusicQueue_InvalidBody() {
	req := httptest.NewRequest("POST", "/api/v1/music/session/test-session/queue", bytes.NewBufferString("invalid"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	var resp APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Invalid request body", resp.Error)
}

// SetEqualizer tests

func (suite *MediaPlayerHandlersTestSuite) TestSetEqualizer_InvalidBody() {
	req := httptest.NewRequest("POST", "/api/v1/music/session/test-session/equalizer", bytes.NewBufferString("invalid"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

// UpdateVideoPlayback tests

func (suite *MediaPlayerHandlersTestSuite) TestUpdateVideoPlayback_InvalidBody() {
	req := httptest.NewRequest("POST", "/api/v1/video/session/test-session/update", bytes.NewBufferString("invalid"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

// SeekVideo tests

func (suite *MediaPlayerHandlersTestSuite) TestSeekVideo_InvalidBody() {
	req := httptest.NewRequest("POST", "/api/v1/video/session/test-session/seek", bytes.NewBufferString("invalid"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

// CreateVideoBookmark tests

func (suite *MediaPlayerHandlersTestSuite) TestCreateVideoBookmark_InvalidBody() {
	req := httptest.NewRequest("POST", "/api/v1/video/session/test-session/bookmark", bytes.NewBufferString("invalid"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

// UpdatePlaylist tests

func (suite *MediaPlayerHandlersTestSuite) TestUpdatePlaylist_InvalidBody() {
	req := httptest.NewRequest("PUT", "/api/v1/playlists/1", bytes.NewBufferString("invalid"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

func (suite *MediaPlayerHandlersTestSuite) TestUpdatePlaylist_InvalidPlaylistID() {
	req := httptest.NewRequest("PUT", "/api/v1/playlists/abc", bytes.NewBufferString(`{"name":"test"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	var resp APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Invalid playlist ID", resp.Error)
}

func (suite *MediaPlayerHandlersTestSuite) TestUpdatePlaylist_ValidBody() {
	body := `{"name": "My Playlist", "description": "A test playlist"}`
	req := httptest.NewRequest("PUT", "/api/v1/playlists/1", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	var resp APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), resp.Success)
}

// GetPlaylist tests

func (suite *MediaPlayerHandlersTestSuite) TestGetPlaylist_InvalidID() {
	req := httptest.NewRequest("GET", "/api/v1/playlists/abc", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	var resp APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Invalid playlist ID", resp.Error)
}

// GetPlaylistItems tests

func (suite *MediaPlayerHandlersTestSuite) TestGetPlaylistItems_InvalidID() {
	req := httptest.NewRequest("GET", "/api/v1/playlists/abc/items", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

// AddToPlaylist tests

func (suite *MediaPlayerHandlersTestSuite) TestAddToPlaylist_InvalidBody() {
	req := httptest.NewRequest("POST", "/api/v1/playlists/1/items", bytes.NewBufferString("invalid"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

func (suite *MediaPlayerHandlersTestSuite) TestAddToPlaylist_InvalidPlaylistID() {
	req := httptest.NewRequest("POST", "/api/v1/playlists/abc/items", bytes.NewBufferString(`{"items":[]}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

// RemoveFromPlaylist tests

func (suite *MediaPlayerHandlersTestSuite) TestRemoveFromPlaylist_InvalidPlaylistID() {
	req := httptest.NewRequest("DELETE", "/api/v1/playlists/abc/items/1", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

func (suite *MediaPlayerHandlersTestSuite) TestRemoveFromPlaylist_InvalidItemID() {
	req := httptest.NewRequest("DELETE", "/api/v1/playlists/1/items/abc", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

// ReorderPlaylist tests

func (suite *MediaPlayerHandlersTestSuite) TestReorderPlaylist_InvalidPlaylistID() {
	req := httptest.NewRequest("POST", "/api/v1/playlists/abc/items/1/reorder", bytes.NewBufferString(`{"new_position":2}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

func (suite *MediaPlayerHandlersTestSuite) TestReorderPlaylist_InvalidItemID() {
	req := httptest.NewRequest("POST", "/api/v1/playlists/1/items/abc/reorder", bytes.NewBufferString(`{"new_position":2}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

func (suite *MediaPlayerHandlersTestSuite) TestReorderPlaylist_InvalidBody() {
	req := httptest.NewRequest("POST", "/api/v1/playlists/1/items/1/reorder", bytes.NewBufferString("invalid"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

// RefreshSmartPlaylist tests

func (suite *MediaPlayerHandlersTestSuite) TestRefreshSmartPlaylist_InvalidPlaylistID() {
	req := httptest.NewRequest("POST", "/api/v1/playlists/abc/refresh", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

// DownloadSubtitle tests

func (suite *MediaPlayerHandlersTestSuite) TestDownloadSubtitle_InvalidBody() {
	req := httptest.NewRequest("POST", "/api/v1/subtitles/download", bytes.NewBufferString("invalid"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

// TranslateSubtitle tests

func (suite *MediaPlayerHandlersTestSuite) TestTranslateSubtitle_InvalidBody() {
	req := httptest.NewRequest("POST", "/api/v1/subtitles/translate", bytes.NewBufferString("invalid"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

// SynchronizeLyrics tests

func (suite *MediaPlayerHandlersTestSuite) TestSynchronizeLyrics_InvalidBody() {
	req := httptest.NewRequest("POST", "/api/v1/lyrics/sync", bytes.NewBufferString("invalid"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

// GetConcertLyrics tests

func (suite *MediaPlayerHandlersTestSuite) TestGetConcertLyrics_InvalidBody() {
	req := httptest.NewRequest("POST", "/api/v1/lyrics/concert", bytes.NewBufferString("invalid"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

// SearchCoverArt tests

func (suite *MediaPlayerHandlersTestSuite) TestSearchCoverArt_InvalidBody() {
	req := httptest.NewRequest("POST", "/api/v1/cover-art/search", bytes.NewBufferString("invalid"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

// ScanLocalCoverArt tests

func (suite *MediaPlayerHandlersTestSuite) TestScanLocalCoverArt_InvalidBody() {
	req := httptest.NewRequest("POST", "/api/v1/cover-art/scan", bytes.NewBufferString("invalid"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

// GetPlaybackPosition tests

func (suite *MediaPlayerHandlersTestSuite) TestGetPlaybackPosition_InvalidMediaID() {
	req := httptest.NewRequest("GET", "/api/v1/playback/position/abc", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	var resp APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Invalid media ID", resp.Error)
}

func (suite *MediaPlayerHandlersTestSuite) TestGetPlaybackPosition_Unauthenticated() {
	req := httptest.NewRequest("GET", "/api/v1/playback/position/1", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
}

func (suite *MediaPlayerHandlersTestSuite) TestGetPlaybackPosition_ValidIDAndUser_PanicsOnNilService() {
	// With a valid media ID and authenticated user, the handler will try to call
	// the nil positionService, which panics. This confirms validation passed.
	req := httptest.NewRequest("GET", "/api/v1/playback/position/1", nil)
	ctx := context.WithValue(req.Context(), "user_id", int64(1))
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	assert.Panics(suite.T(), func() {
		suite.router.ServeHTTP(w, req)
	})
}

// GetBookmarks tests

func (suite *MediaPlayerHandlersTestSuite) TestGetBookmarks_InvalidMediaID() {
	req := httptest.NewRequest("GET", "/api/v1/playback/bookmarks/abc", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

func (suite *MediaPlayerHandlersTestSuite) TestGetBookmarks_Unauthenticated() {
	req := httptest.NewRequest("GET", "/api/v1/playback/bookmarks/1", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
}

// CORS middleware tests

func (suite *MediaPlayerHandlersTestSuite) TestCORSMiddleware_OptionsRequest() {
	req := httptest.NewRequest("OPTIONS", "/api/v1/playlists", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	assert.Contains(suite.T(), w.Header().Get("Access-Control-Allow-Methods"), "GET")
	assert.Contains(suite.T(), w.Header().Get("Access-Control-Allow-Methods"), "POST")
}

func (suite *MediaPlayerHandlersTestSuite) TestCORSMiddleware_AllowedOrigin() {
	req := httptest.NewRequest("OPTIONS", "/api/v1/playlists", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), "http://localhost:5173", w.Header().Get("Access-Control-Allow-Origin"))
}

// RegisterRoutes test - verify route registration by checking that wrong methods get 405

func (suite *MediaPlayerHandlersTestSuite) TestRegisterRoutes_VerifyEndpointsExist() {
	// These POST endpoints expect invalid body → 400, confirming the route is registered
	postEndpoints := []string{
		"/api/v1/music/play",
		"/api/v1/music/play/album",
		"/api/v1/music/play/artist",
		"/api/v1/video/play",
		"/api/v1/video/play/series",
		"/api/v1/playlists",
		"/api/v1/subtitles/search",
		"/api/v1/subtitles/download",
		"/api/v1/lyrics/search",
		"/api/v1/cover-art/search",
		"/api/v1/translate",
	}

	for _, path := range postEndpoints {
		req := httptest.NewRequest("POST", path, bytes.NewBufferString("invalid-json"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)

		// 400 confirms the route is registered (handler ran, rejected body)
		assert.Equal(suite.T(), http.StatusBadRequest, w.Code, "Expected 400 for %s", path)
	}
}

func TestMediaPlayerHandlersTestSuite(t *testing.T) {
	suite.Run(t, new(MediaPlayerHandlersTestSuite))
}
