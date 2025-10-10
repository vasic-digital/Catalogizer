package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"catalog-api/internal/services"
)

type MediaPlayerHandlers struct {
	logger              *zap.Logger
	musicPlayerService  *services.MusicPlayerService
	videoPlayerService  *services.VideoPlayerService
	playlistService     *services.PlaylistService
	positionService     *services.PlaybackPositionService
	subtitleService     *services.SubtitleService
	lyricsService       *services.LyricsService
	coverArtService     *services.CoverArtService
	translationService  *services.TranslationService
}

type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}

type PlaybackSessionResponse struct {
	SessionID       string                      `json:"session_id"`
	CurrentTrack    *services.MusicTrack        `json:"current_track,omitempty"`
	CurrentVideo    *services.VideoContent      `json:"current_video,omitempty"`
	PlaybackState   services.PlaybackState      `json:"playback_state"`
	Position        int64                       `json:"position"`
	Duration        int64                       `json:"duration"`
	Volume          float64                     `json:"volume"`
	IsMuted         bool                        `json:"is_muted"`
	PlaybackSpeed   float64                     `json:"playback_speed,omitempty"`
	Queue           interface{}                 `json:"queue,omitempty"`
	QueueIndex      int                         `json:"queue_index,omitempty"`
	RepeatMode      string                      `json:"repeat_mode,omitempty"`
	ShuffleEnabled  bool                        `json:"shuffle_enabled,omitempty"`
	Subtitles       []services.SubtitleTrack    `json:"subtitles,omitempty"`
	AudioTracks     []services.AudioTrack       `json:"audio_tracks,omitempty"`
	Chapters        []services.Chapter          `json:"chapters,omitempty"`
	Lyrics          *services.LyricsData        `json:"lyrics,omitempty"`
	CoverArt        *services.CoverArt          `json:"cover_art,omitempty"`
	LastActivity    time.Time                   `json:"last_activity"`
}

func NewMediaPlayerHandlers(
	logger *zap.Logger,
	musicPlayerService *services.MusicPlayerService,
	videoPlayerService *services.VideoPlayerService,
	playlistService *services.PlaylistService,
	positionService *services.PlaybackPositionService,
	subtitleService *services.SubtitleService,
	lyricsService *services.LyricsService,
	coverArtService *services.CoverArtService,
	translationService *services.TranslationService,
) *MediaPlayerHandlers {
	return &MediaPlayerHandlers{
		logger:              logger,
		musicPlayerService:  musicPlayerService,
		videoPlayerService:  videoPlayerService,
		playlistService:     playlistService,
		positionService:     positionService,
		subtitleService:     subtitleService,
		lyricsService:       lyricsService,
		coverArtService:     coverArtService,
		translationService:  translationService,
	}
}

func (h *MediaPlayerHandlers) RegisterRoutes(router *mux.Router) {
	api := router.PathPrefix("/api/v1").Subrouter()

	// Music Player Routes
	api.HandleFunc("/music/play", h.PlayMusic).Methods("POST", "OPTIONS")
	api.HandleFunc("/music/play/album", h.PlayAlbum).Methods("POST", "OPTIONS")
	api.HandleFunc("/music/play/artist", h.PlayArtist).Methods("POST", "OPTIONS")
	api.HandleFunc("/music/session/{sessionId}", h.GetMusicSession).Methods("GET", "OPTIONS")
	api.HandleFunc("/music/session/{sessionId}/update", h.UpdateMusicPlayback).Methods("POST", "OPTIONS")
	api.HandleFunc("/music/session/{sessionId}/next", h.NextTrack).Methods("POST", "OPTIONS")
	api.HandleFunc("/music/session/{sessionId}/previous", h.PreviousTrack).Methods("POST", "OPTIONS")
	api.HandleFunc("/music/session/{sessionId}/seek", h.SeekMusic).Methods("POST", "OPTIONS")
	api.HandleFunc("/music/session/{sessionId}/queue", h.AddToMusicQueue).Methods("POST", "OPTIONS")
	api.HandleFunc("/music/session/{sessionId}/equalizer", h.SetEqualizer).Methods("POST", "OPTIONS")
	api.HandleFunc("/music/library/stats", h.GetMusicLibraryStats).Methods("GET", "OPTIONS")

	// Video Player Routes
	api.HandleFunc("/video/play", h.PlayVideo).Methods("POST", "OPTIONS")
	api.HandleFunc("/video/play/series", h.PlaySeries).Methods("POST", "OPTIONS")
	api.HandleFunc("/video/session/{sessionId}", h.GetVideoSession).Methods("GET", "OPTIONS")
	api.HandleFunc("/video/session/{sessionId}/update", h.UpdateVideoPlayback).Methods("POST", "OPTIONS")
	api.HandleFunc("/video/session/{sessionId}/next", h.NextVideo).Methods("POST", "OPTIONS")
	api.HandleFunc("/video/session/{sessionId}/previous", h.PreviousVideo).Methods("POST", "OPTIONS")
	api.HandleFunc("/video/session/{sessionId}/seek", h.SeekVideo).Methods("POST", "OPTIONS")
	api.HandleFunc("/video/session/{sessionId}/bookmark", h.CreateVideoBookmark).Methods("POST", "OPTIONS")
	api.HandleFunc("/video/continue-watching", h.GetContinueWatching).Methods("GET", "OPTIONS")
	api.HandleFunc("/video/watch-history", h.GetWatchHistory).Methods("GET", "OPTIONS")

	// Playlist Routes
	api.HandleFunc("/playlists", h.CreatePlaylist).Methods("POST", "OPTIONS")
	api.HandleFunc("/playlists", h.GetUserPlaylists).Methods("GET", "OPTIONS")
	api.HandleFunc("/playlists/{playlistId}", h.GetPlaylist).Methods("GET", "OPTIONS")
	api.HandleFunc("/playlists/{playlistId}", h.UpdatePlaylist).Methods("PUT", "OPTIONS")
	api.HandleFunc("/playlists/{playlistId}/items", h.GetPlaylistItems).Methods("GET", "OPTIONS")
	api.HandleFunc("/playlists/{playlistId}/items", h.AddToPlaylist).Methods("POST", "OPTIONS")
	api.HandleFunc("/playlists/{playlistId}/items/{itemId}", h.RemoveFromPlaylist).Methods("DELETE", "OPTIONS")
	api.HandleFunc("/playlists/{playlistId}/items/{itemId}/reorder", h.ReorderPlaylist).Methods("POST", "OPTIONS")
	api.HandleFunc("/playlists/{playlistId}/refresh", h.RefreshSmartPlaylist).Methods("POST", "OPTIONS")

	// Subtitle Routes
	api.HandleFunc("/subtitles/search", h.SearchSubtitles).Methods("POST", "OPTIONS")
	api.HandleFunc("/subtitles/download", h.DownloadSubtitle).Methods("POST", "OPTIONS")
	api.HandleFunc("/subtitles/translate", h.TranslateSubtitle).Methods("POST", "OPTIONS")

	// Lyrics Routes
	api.HandleFunc("/lyrics/search", h.SearchLyrics).Methods("POST", "OPTIONS")
	api.HandleFunc("/lyrics/sync", h.SynchronizeLyrics).Methods("POST", "OPTIONS")
	api.HandleFunc("/lyrics/concert", h.GetConcertLyrics).Methods("POST", "OPTIONS")

	// Cover Art Routes
	api.HandleFunc("/cover-art/search", h.SearchCoverArt).Methods("POST", "OPTIONS")
	api.HandleFunc("/cover-art/scan", h.ScanLocalCoverArt).Methods("POST", "OPTIONS")

	// Translation Routes
	api.HandleFunc("/translate", h.TranslateText).Methods("POST", "OPTIONS")
	api.HandleFunc("/translate/detect", h.DetectLanguage).Methods("POST", "OPTIONS")

	// Position Tracking Routes
	api.HandleFunc("/playback/position", h.UpdatePlaybackPosition).Methods("POST", "OPTIONS")
	api.HandleFunc("/playback/position/{mediaId}", h.GetPlaybackPosition).Methods("GET", "OPTIONS")
	api.HandleFunc("/playback/continue-watching", h.GetContinueWatchingList).Methods("GET", "OPTIONS")
	api.HandleFunc("/playback/bookmarks", h.CreateBookmark).Methods("POST", "OPTIONS")
	api.HandleFunc("/playback/bookmarks/{mediaId}", h.GetBookmarks).Methods("GET", "OPTIONS")
	api.HandleFunc("/playback/stats", h.GetPlaybackStats).Methods("GET", "OPTIONS")

	// Add CORS middleware
	api.Use(h.corsMiddleware)
}

// Music Player Handlers

func (h *MediaPlayerHandlers) PlayMusic(w http.ResponseWriter, r *http.Request) {
	var req services.PlayTrackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	session, err := h.musicPlayerService.PlayTrack(r.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to play music", zap.Error(err))
		h.sendError(w, "Failed to start playback", http.StatusInternalServerError)
		return
	}

	response := h.buildMusicSessionResponse(session)
	h.sendSuccess(w, response)
}

func (h *MediaPlayerHandlers) PlayAlbum(w http.ResponseWriter, r *http.Request) {
	var req services.PlayAlbumRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	session, err := h.musicPlayerService.PlayAlbum(r.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to play album", zap.Error(err))
		h.sendError(w, "Failed to start album playback", http.StatusInternalServerError)
		return
	}

	response := h.buildMusicSessionResponse(session)
	h.sendSuccess(w, response)
}

func (h *MediaPlayerHandlers) PlayArtist(w http.ResponseWriter, r *http.Request) {
	var req services.PlayArtistRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	session, err := h.musicPlayerService.PlayArtist(r.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to play artist", zap.Error(err))
		h.sendError(w, "Failed to start artist playback", http.StatusInternalServerError)
		return
	}

	response := h.buildMusicSessionResponse(session)
	h.sendSuccess(w, response)
}

func (h *MediaPlayerHandlers) GetMusicSession(w http.ResponseWriter, r *http.Request) {
	sessionID := mux.Vars(r)["sessionId"]

	session, err := h.musicPlayerService.GetSession(r.Context(), sessionID)
	if err != nil {
		h.logger.Error("Failed to get music session", zap.Error(err))
		h.sendError(w, "Session not found", http.StatusNotFound)
		return
	}

	response := h.buildMusicSessionResponse(session)
	h.sendSuccess(w, response)
}

func (h *MediaPlayerHandlers) UpdateMusicPlayback(w http.ResponseWriter, r *http.Request) {
	sessionID := mux.Vars(r)["sessionId"]

	var req services.UpdatePlaybackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	req.SessionID = sessionID

	session, err := h.musicPlayerService.UpdatePlayback(r.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to update music playback", zap.Error(err))
		h.sendError(w, "Failed to update playback", http.StatusInternalServerError)
		return
	}

	response := h.buildMusicSessionResponse(session)
	h.sendSuccess(w, response)
}

func (h *MediaPlayerHandlers) NextTrack(w http.ResponseWriter, r *http.Request) {
	sessionID := mux.Vars(r)["sessionId"]

	session, err := h.musicPlayerService.NextTrack(r.Context(), sessionID)
	if err != nil {
		h.logger.Error("Failed to skip to next track", zap.Error(err))
		h.sendError(w, "Failed to skip track", http.StatusInternalServerError)
		return
	}

	response := h.buildMusicSessionResponse(session)
	h.sendSuccess(w, response)
}

func (h *MediaPlayerHandlers) PreviousTrack(w http.ResponseWriter, r *http.Request) {
	sessionID := mux.Vars(r)["sessionId"]

	session, err := h.musicPlayerService.PreviousTrack(r.Context(), sessionID)
	if err != nil {
		h.logger.Error("Failed to skip to previous track", zap.Error(err))
		h.sendError(w, "Failed to skip track", http.StatusInternalServerError)
		return
	}

	response := h.buildMusicSessionResponse(session)
	h.sendSuccess(w, response)
}

func (h *MediaPlayerHandlers) SeekMusic(w http.ResponseWriter, r *http.Request) {
	sessionID := mux.Vars(r)["sessionId"]

	var req services.SeekRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	req.SessionID = sessionID

	session, err := h.musicPlayerService.Seek(r.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to seek in track", zap.Error(err))
		h.sendError(w, "Failed to seek", http.StatusInternalServerError)
		return
	}

	response := h.buildMusicSessionResponse(session)
	h.sendSuccess(w, response)
}

func (h *MediaPlayerHandlers) AddToMusicQueue(w http.ResponseWriter, r *http.Request) {
	sessionID := mux.Vars(r)["sessionId"]

	var req services.QueueRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	req.SessionID = sessionID

	session, err := h.musicPlayerService.AddToQueue(r.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to add to queue", zap.Error(err))
		h.sendError(w, "Failed to add to queue", http.StatusInternalServerError)
		return
	}

	response := h.buildMusicSessionResponse(session)
	h.sendSuccess(w, response)
}

func (h *MediaPlayerHandlers) SetEqualizer(w http.ResponseWriter, r *http.Request) {
	sessionID := mux.Vars(r)["sessionId"]

	var req struct {
		Preset string             `json:"preset"`
		Bands  map[string]float64 `json:"bands"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := h.musicPlayerService.SetEqualizer(r.Context(), sessionID, req.Preset, req.Bands)
	if err != nil {
		h.logger.Error("Failed to set equalizer", zap.Error(err))
		h.sendError(w, "Failed to set equalizer", http.StatusInternalServerError)
		return
	}

	h.sendSuccess(w, map[string]string{"message": "Equalizer updated successfully"})
}

func (h *MediaPlayerHandlers) GetMusicLibraryStats(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserID(r)
	if userID == 0 {
		h.sendError(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	stats, err := h.musicPlayerService.GetLibraryStats(r.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get library stats", zap.Error(err))
		h.sendError(w, "Failed to get library statistics", http.StatusInternalServerError)
		return
	}

	h.sendSuccess(w, stats)
}

// Video Player Handlers

func (h *MediaPlayerHandlers) PlayVideo(w http.ResponseWriter, r *http.Request) {
	var req services.PlayVideoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	session, err := h.videoPlayerService.PlayVideo(r.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to play video", zap.Error(err))
		h.sendError(w, "Failed to start video playback", http.StatusInternalServerError)
		return
	}

	response := h.buildVideoSessionResponse(session)
	h.sendSuccess(w, response)
}

func (h *MediaPlayerHandlers) PlaySeries(w http.ResponseWriter, r *http.Request) {
	var req services.PlaySeriesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	session, err := h.videoPlayerService.PlaySeries(r.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to play series", zap.Error(err))
		h.sendError(w, "Failed to start series playback", http.StatusInternalServerError)
		return
	}

	response := h.buildVideoSessionResponse(session)
	h.sendSuccess(w, response)
}

func (h *MediaPlayerHandlers) GetVideoSession(w http.ResponseWriter, r *http.Request) {
	sessionID := mux.Vars(r)["sessionId"]

	session, err := h.videoPlayerService.GetVideoSession(r.Context(), sessionID)
	if err != nil {
		h.logger.Error("Failed to get video session", zap.Error(err))
		h.sendError(w, "Session not found", http.StatusNotFound)
		return
	}

	response := h.buildVideoSessionResponse(session)
	h.sendSuccess(w, response)
}

func (h *MediaPlayerHandlers) UpdateVideoPlayback(w http.ResponseWriter, r *http.Request) {
	sessionID := mux.Vars(r)["sessionId"]

	var req services.UpdateVideoPlaybackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	req.SessionID = sessionID

	session, err := h.videoPlayerService.UpdateVideoPlayback(r.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to update video playback", zap.Error(err))
		h.sendError(w, "Failed to update playback", http.StatusInternalServerError)
		return
	}

	response := h.buildVideoSessionResponse(session)
	h.sendSuccess(w, response)
}

func (h *MediaPlayerHandlers) NextVideo(w http.ResponseWriter, r *http.Request) {
	sessionID := mux.Vars(r)["sessionId"]

	session, err := h.videoPlayerService.NextVideo(r.Context(), sessionID)
	if err != nil {
		h.logger.Error("Failed to skip to next video", zap.Error(err))
		h.sendError(w, "Failed to skip video", http.StatusInternalServerError)
		return
	}

	response := h.buildVideoSessionResponse(session)
	h.sendSuccess(w, response)
}

func (h *MediaPlayerHandlers) PreviousVideo(w http.ResponseWriter, r *http.Request) {
	sessionID := mux.Vars(r)["sessionId"]

	session, err := h.videoPlayerService.PreviousVideo(r.Context(), sessionID)
	if err != nil {
		h.logger.Error("Failed to skip to previous video", zap.Error(err))
		h.sendError(w, "Failed to skip video", http.StatusInternalServerError)
		return
	}

	response := h.buildVideoSessionResponse(session)
	h.sendSuccess(w, response)
}

func (h *MediaPlayerHandlers) SeekVideo(w http.ResponseWriter, r *http.Request) {
	sessionID := mux.Vars(r)["sessionId"]

	var req services.VideoSeekRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	req.SessionID = sessionID

	session, err := h.videoPlayerService.SeekVideo(r.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to seek in video", zap.Error(err))
		h.sendError(w, "Failed to seek", http.StatusInternalServerError)
		return
	}

	response := h.buildVideoSessionResponse(session)
	h.sendSuccess(w, response)
}

func (h *MediaPlayerHandlers) CreateVideoBookmark(w http.ResponseWriter, r *http.Request) {
	sessionID := mux.Vars(r)["sessionId"]

	var req services.CreateVideoBookmarkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	req.SessionID = sessionID

	bookmark, err := h.videoPlayerService.CreateVideoBookmark(r.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to create bookmark", zap.Error(err))
		h.sendError(w, "Failed to create bookmark", http.StatusInternalServerError)
		return
	}

	h.sendSuccess(w, bookmark)
}

func (h *MediaPlayerHandlers) GetContinueWatching(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserID(r)
	if userID == 0 {
		h.sendError(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	limit := 20
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	videos, err := h.videoPlayerService.GetContinueWatching(r.Context(), userID, limit)
	if err != nil {
		h.logger.Error("Failed to get continue watching", zap.Error(err))
		h.sendError(w, "Failed to get continue watching", http.StatusInternalServerError)
		return
	}

	h.sendSuccess(w, videos)
}

func (h *MediaPlayerHandlers) GetWatchHistory(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserID(r)
	if userID == 0 {
		h.sendError(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	req := services.WatchHistoryRequest{
		UserID: userID,
		Limit:  50,
		Offset: 0,
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			req.Limit = l
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			req.Offset = o
		}
	}

	if videoType := r.URL.Query().Get("type"); videoType != "" {
		vt := services.VideoType(videoType)
		req.VideoType = &vt
	}

	history, err := h.videoPlayerService.GetWatchHistory(r.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to get watch history", zap.Error(err))
		h.sendError(w, "Failed to get watch history", http.StatusInternalServerError)
		return
	}

	h.sendSuccess(w, history)
}

// Playlist Handlers

func (h *MediaPlayerHandlers) CreatePlaylist(w http.ResponseWriter, r *http.Request) {
	var req services.CreatePlaylistRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	playlist, err := h.playlistService.CreatePlaylist(r.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to create playlist", zap.Error(err))
		h.sendError(w, "Failed to create playlist", http.StatusInternalServerError)
		return
	}

	h.sendSuccess(w, playlist)
}

func (h *MediaPlayerHandlers) GetUserPlaylists(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserID(r)
	if userID == 0 {
		h.sendError(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	includePublic := r.URL.Query().Get("include_public") == "true"

	playlists, err := h.playlistService.GetUserPlaylists(r.Context(), userID, includePublic)
	if err != nil {
		h.logger.Error("Failed to get user playlists", zap.Error(err))
		h.sendError(w, "Failed to get playlists", http.StatusInternalServerError)
		return
	}

	h.sendSuccess(w, playlists)
}

func (h *MediaPlayerHandlers) GetPlaylist(w http.ResponseWriter, r *http.Request) {
	playlistID, err := h.getIDFromPath(r, "playlistId")
	if err != nil {
		h.sendError(w, "Invalid playlist ID", http.StatusBadRequest)
		return
	}

	userID := h.getUserID(r)

	playlist, err := h.playlistService.GetPlaylist(r.Context(), playlistID, userID)
	if err != nil {
		h.logger.Error("Failed to get playlist", zap.Error(err))
		h.sendError(w, "Playlist not found", http.StatusNotFound)
		return
	}

	h.sendSuccess(w, playlist)
}

func (h *MediaPlayerHandlers) UpdatePlaylist(w http.ResponseWriter, r *http.Request) {
	playlistID, err := h.getIDFromPath(r, "playlistId")
	if err != nil {
		h.sendError(w, "Invalid playlist ID", http.StatusBadRequest)
		return
	}

	var req services.UpdatePlaylistRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	req.ID = playlistID

	h.sendSuccess(w, map[string]string{"message": "Playlist updated successfully"})
}

func (h *MediaPlayerHandlers) GetPlaylistItems(w http.ResponseWriter, r *http.Request) {
	playlistID, err := h.getIDFromPath(r, "playlistId")
	if err != nil {
		h.sendError(w, "Invalid playlist ID", http.StatusBadRequest)
		return
	}

	userID := h.getUserID(r)
	limit := 100
	offset := 0

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	items, err := h.playlistService.GetPlaylistItems(r.Context(), playlistID, userID, limit, offset)
	if err != nil {
		h.logger.Error("Failed to get playlist items", zap.Error(err))
		h.sendError(w, "Failed to get playlist items", http.StatusInternalServerError)
		return
	}

	h.sendSuccess(w, items)
}

func (h *MediaPlayerHandlers) AddToPlaylist(w http.ResponseWriter, r *http.Request) {
	playlistID, err := h.getIDFromPath(r, "playlistId")
	if err != nil {
		h.sendError(w, "Invalid playlist ID", http.StatusBadRequest)
		return
	}

	var req services.AddToPlaylistRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	req.PlaylistID = playlistID

	err = h.playlistService.AddToPlaylist(r.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to add to playlist", zap.Error(err))
		h.sendError(w, "Failed to add items to playlist", http.StatusInternalServerError)
		return
	}

	h.sendSuccess(w, map[string]string{"message": "Items added to playlist successfully"})
}

func (h *MediaPlayerHandlers) RemoveFromPlaylist(w http.ResponseWriter, r *http.Request) {
	playlistID, err := h.getIDFromPath(r, "playlistId")
	if err != nil {
		h.sendError(w, "Invalid playlist ID", http.StatusBadRequest)
		return
	}

	itemID, err := h.getIDFromPath(r, "itemId")
	if err != nil {
		h.sendError(w, "Invalid item ID", http.StatusBadRequest)
		return
	}

	userID := h.getUserID(r)

	err = h.playlistService.RemoveFromPlaylist(r.Context(), playlistID, itemID, userID)
	if err != nil {
		h.logger.Error("Failed to remove from playlist", zap.Error(err))
		h.sendError(w, "Failed to remove item from playlist", http.StatusInternalServerError)
		return
	}

	h.sendSuccess(w, map[string]string{"message": "Item removed from playlist successfully"})
}

func (h *MediaPlayerHandlers) ReorderPlaylist(w http.ResponseWriter, r *http.Request) {
	playlistID, err := h.getIDFromPath(r, "playlistId")
	if err != nil {
		h.sendError(w, "Invalid playlist ID", http.StatusBadRequest)
		return
	}

	itemID, err := h.getIDFromPath(r, "itemId")
	if err != nil {
		h.sendError(w, "Invalid item ID", http.StatusBadRequest)
		return
	}

	var req services.ReorderPlaylistRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	req.PlaylistID = playlistID
	req.ItemID = itemID

	err = h.playlistService.ReorderPlaylist(r.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to reorder playlist", zap.Error(err))
		h.sendError(w, "Failed to reorder playlist", http.StatusInternalServerError)
		return
	}

	h.sendSuccess(w, map[string]string{"message": "Playlist reordered successfully"})
}

func (h *MediaPlayerHandlers) RefreshSmartPlaylist(w http.ResponseWriter, r *http.Request) {
	playlistID, err := h.getIDFromPath(r, "playlistId")
	if err != nil {
		h.sendError(w, "Invalid playlist ID", http.StatusBadRequest)
		return
	}

	err = h.playlistService.RefreshSmartPlaylist(r.Context(), playlistID)
	if err != nil {
		h.logger.Error("Failed to refresh smart playlist", zap.Error(err))
		h.sendError(w, "Failed to refresh smart playlist", http.StatusInternalServerError)
		return
	}

	h.sendSuccess(w, map[string]string{"message": "Smart playlist refreshed successfully"})
}

// Subtitle Handlers

func (h *MediaPlayerHandlers) SearchSubtitles(w http.ResponseWriter, r *http.Request) {
	var req services.SubtitleSearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	results, err := h.subtitleService.SearchSubtitles(r.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to search subtitles", zap.Error(err))
		h.sendError(w, "Failed to search subtitles", http.StatusInternalServerError)
		return
	}

	h.sendSuccess(w, results)
}

func (h *MediaPlayerHandlers) DownloadSubtitle(w http.ResponseWriter, r *http.Request) {
	var req services.SubtitleDownloadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	subtitle, err := h.subtitleService.DownloadSubtitle(r.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to download subtitle", zap.Error(err))
		h.sendError(w, "Failed to download subtitle", http.StatusInternalServerError)
		return
	}

	h.sendSuccess(w, subtitle)
}

func (h *MediaPlayerHandlers) TranslateSubtitle(w http.ResponseWriter, r *http.Request) {
	var req services.SubtitleTranslationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	subtitle, err := h.subtitleService.TranslateSubtitle(r.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to translate subtitle", zap.Error(err))
		h.sendError(w, "Failed to translate subtitle", http.StatusInternalServerError)
		return
	}

	h.sendSuccess(w, subtitle)
}

// Lyrics Handlers

func (h *MediaPlayerHandlers) SearchLyrics(w http.ResponseWriter, r *http.Request) {
	var req services.LyricsSearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	results, err := h.lyricsService.SearchLyrics(r.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to search lyrics", zap.Error(err))
		h.sendError(w, "Failed to search lyrics", http.StatusInternalServerError)
		return
	}

	h.sendSuccess(w, results)
}

func (h *MediaPlayerHandlers) SynchronizeLyrics(w http.ResponseWriter, r *http.Request) {
	var req services.LyricsSyncRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	lyrics, err := h.lyricsService.SynchronizeLyrics(r.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to synchronize lyrics", zap.Error(err))
		h.sendError(w, "Failed to synchronize lyrics", http.StatusInternalServerError)
		return
	}

	h.sendSuccess(w, lyrics)
}

func (h *MediaPlayerHandlers) GetConcertLyrics(w http.ResponseWriter, r *http.Request) {
	var req services.ConcertLyricsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	lyrics, err := h.lyricsService.GetConcertLyrics(r.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to get concert lyrics", zap.Error(err))
		h.sendError(w, "Failed to get concert lyrics", http.StatusInternalServerError)
		return
	}

	h.sendSuccess(w, lyrics)
}

// Cover Art Handlers

func (h *MediaPlayerHandlers) SearchCoverArt(w http.ResponseWriter, r *http.Request) {
	var req services.CoverArtSearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	results, err := h.coverArtService.SearchCoverArt(r.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to search cover art", zap.Error(err))
		h.sendError(w, "Failed to search cover art", http.StatusInternalServerError)
		return
	}

	h.sendSuccess(w, results)
}

func (h *MediaPlayerHandlers) ScanLocalCoverArt(w http.ResponseWriter, r *http.Request) {
	var req services.LocalCoverArtScanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	results, err := h.coverArtService.ScanLocalCoverArt(r.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to scan local cover art", zap.Error(err))
		h.sendError(w, "Failed to scan local cover art", http.StatusInternalServerError)
		return
	}

	h.sendSuccess(w, results)
}

// Translation Handlers

func (h *MediaPlayerHandlers) TranslateText(w http.ResponseWriter, r *http.Request) {
	var req services.TranslationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	result, err := h.translationService.TranslateText(r.Context(), req)
	if err != nil {
		h.logger.Error("Failed to translate text", zap.Error(err))
		h.sendError(w, "Failed to translate text", http.StatusInternalServerError)
		return
	}

	h.sendSuccess(w, result)
}

func (h *MediaPlayerHandlers) DetectLanguage(w http.ResponseWriter, r *http.Request) {
	var req services.LanguageDetectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	result, err := h.translationService.DetectLanguage(r.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to detect language", zap.Error(err))
		h.sendError(w, "Failed to detect language", http.StatusInternalServerError)
		return
	}

	h.sendSuccess(w, result)
}

// Position Tracking Handlers

func (h *MediaPlayerHandlers) UpdatePlaybackPosition(w http.ResponseWriter, r *http.Request) {
	var req services.UpdatePositionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := h.positionService.UpdatePosition(r.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to update position", zap.Error(err))
		h.sendError(w, "Failed to update playback position", http.StatusInternalServerError)
		return
	}

	h.sendSuccess(w, map[string]string{"message": "Position updated successfully"})
}

func (h *MediaPlayerHandlers) GetPlaybackPosition(w http.ResponseWriter, r *http.Request) {
	mediaID, err := h.getIDFromPath(r, "mediaId")
	if err != nil {
		h.sendError(w, "Invalid media ID", http.StatusBadRequest)
		return
	}

	userID := h.getUserID(r)
	if userID == 0 {
		h.sendError(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	position, err := h.positionService.GetPosition(r.Context(), userID, mediaID)
	if err != nil {
		h.logger.Error("Failed to get position", zap.Error(err))
		h.sendError(w, "Failed to get playback position", http.StatusInternalServerError)
		return
	}

	h.sendSuccess(w, position)
}

func (h *MediaPlayerHandlers) GetContinueWatchingList(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserID(r)
	if userID == 0 {
		h.sendError(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	limit := 20
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	positions, err := h.positionService.GetContinueWatching(r.Context(), userID, limit)
	if err != nil {
		h.logger.Error("Failed to get continue watching", zap.Error(err))
		h.sendError(w, "Failed to get continue watching list", http.StatusInternalServerError)
		return
	}

	h.sendSuccess(w, positions)
}

func (h *MediaPlayerHandlers) CreateBookmark(w http.ResponseWriter, r *http.Request) {
	var req services.BookmarkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	bookmark, err := h.positionService.CreateBookmark(r.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to create bookmark", zap.Error(err))
		h.sendError(w, "Failed to create bookmark", http.StatusInternalServerError)
		return
	}

	h.sendSuccess(w, bookmark)
}

func (h *MediaPlayerHandlers) GetBookmarks(w http.ResponseWriter, r *http.Request) {
	mediaID, err := h.getIDFromPath(r, "mediaId")
	if err != nil {
		h.sendError(w, "Invalid media ID", http.StatusBadRequest)
		return
	}

	userID := h.getUserID(r)
	if userID == 0 {
		h.sendError(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	bookmarks, err := h.positionService.GetBookmarks(r.Context(), userID, mediaID)
	if err != nil {
		h.logger.Error("Failed to get bookmarks", zap.Error(err))
		h.sendError(w, "Failed to get bookmarks", http.StatusInternalServerError)
		return
	}

	h.sendSuccess(w, bookmarks)
}

func (h *MediaPlayerHandlers) GetPlaybackStats(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserID(r)
	if userID == 0 {
		h.sendError(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	req := services.PlaybackStatsRequest{
		UserID: userID,
		Limit:  20,
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			req.Limit = l
		}
	}

	if mediaType := r.URL.Query().Get("media_type"); mediaType != "" {
		req.MediaType = mediaType
	}

	stats, err := h.positionService.GetPlaybackStats(r.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to get playback stats", zap.Error(err))
		h.sendError(w, "Failed to get playback statistics", http.StatusInternalServerError)
		return
	}

	h.sendSuccess(w, stats)
}

// Helper Methods

func (h *MediaPlayerHandlers) buildMusicSessionResponse(session *services.MusicPlaybackSession) *PlaybackSessionResponse {
	return &PlaybackSessionResponse{
		SessionID:      session.ID,
		CurrentTrack:   session.CurrentTrack,
		PlaybackState:  session.PlaybackState,
		Position:       session.Position,
		Duration:       session.Duration,
		Volume:         session.Volume,
		IsMuted:        session.IsMuted,
		Queue:          session.Queue,
		QueueIndex:     session.QueueIndex,
		RepeatMode:     string(session.RepeatMode),
		ShuffleEnabled: session.ShuffleEnabled,
		LastActivity:   session.LastActivity,
	}
}

func (h *MediaPlayerHandlers) buildVideoSessionResponse(session *services.VideoPlaybackSession) *PlaybackSessionResponse {
	return &PlaybackSessionResponse{
		SessionID:     session.ID,
		CurrentVideo:  session.CurrentVideo,
		PlaybackState: session.PlaybackState,
		Position:      session.Position,
		Duration:      session.Duration,
		Volume:        session.Volume,
		IsMuted:       session.IsMuted,
		PlaybackSpeed: session.PlaybackSpeed,
		Queue:         session.Playlist,
		QueueIndex:    session.PlaylistIndex,
		Subtitles:     session.SubtitleTracks,
		AudioTracks:   session.AudioTracks,
		Chapters:      session.Chapters,
		LastActivity:  session.LastActivity,
	}
}

func (h *MediaPlayerHandlers) sendSuccess(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(APIResponse{
		Success: true,
		Data:    data,
	})
}

func (h *MediaPlayerHandlers) sendError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(APIResponse{
		Success: false,
		Error:   message,
	})
}

func (h *MediaPlayerHandlers) getUserID(r *http.Request) int64 {
	userIDStr := r.Header.Get("X-User-ID")
	if userIDStr == "" {
		return 0
	}

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		return 0
	}

	return userID
}

func (h *MediaPlayerHandlers) getIDFromPath(r *http.Request, key string) (int64, error) {
	idStr := mux.Vars(r)[key]
	return strconv.ParseInt(idStr, 10, 64)
}

func (h *MediaPlayerHandlers) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-User-ID")
		w.Header().Set("Access-Control-Max-Age", "86400")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}