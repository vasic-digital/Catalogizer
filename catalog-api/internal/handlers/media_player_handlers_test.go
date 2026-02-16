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

func TestMediaPlayerHandlersTestSuite(t *testing.T) {
	suite.Run(t, new(MediaPlayerHandlersTestSuite))
}
