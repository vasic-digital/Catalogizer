package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"time"

	"go.uber.org/zap"
)

type MusicPlayerService struct {
	db                 *sql.DB
	logger             *zap.Logger
	mediaPlayerService *MediaPlayerService
	playlistService    *PlaylistService
	positionService    *PlaybackPositionService
	lyricsService      *LyricsService
	coverArtService    *CoverArtService
	translationService *TranslationService
}

type MusicPlaybackSession struct {
	ID                string             `json:"id"`
	UserID            int64              `json:"user_id"`
	CurrentTrack      *MusicTrack        `json:"current_track"`
	Queue             []MusicTrack       `json:"queue"`
	QueueIndex        int                `json:"queue_index"`
	PlaylistID        *int64             `json:"playlist_id"`
	PlayMode          PlayMode           `json:"play_mode"`
	RepeatMode        RepeatMode         `json:"repeat_mode"`
	ShuffleEnabled    bool               `json:"shuffle_enabled"`
	ShuffleHistory    []int              `json:"shuffle_history"`
	Volume            float64            `json:"volume"`
	IsMuted           bool               `json:"is_muted"`
	Crossfade         bool               `json:"crossfade"`
	CrossfadeDuration int                `json:"crossfade_duration"`
	EqualizerPreset   string             `json:"equalizer_preset"`
	EqualizerBands    map[string]float64 `json:"equalizer_bands"`
	PlaybackState     PlaybackState      `json:"playback_state"`
	Position          int64              `json:"position"`
	Duration          int64              `json:"duration"`
	BufferedRanges    []BufferedRange    `json:"buffered_ranges"`
	PlaybackQuality   AudioQuality       `json:"playback_quality"`
	DeviceInfo        DeviceInfo         `json:"device_info"`
	LastActivity      time.Time          `json:"last_activity"`
	CreatedAt         time.Time          `json:"created_at"`
	UpdatedAt         time.Time          `json:"updated_at"`
}

type MusicTrack struct {
	ID            int64             `json:"id"`
	Title         string            `json:"title"`
	Artist        string            `json:"artist"`
	Album         string            `json:"album"`
	AlbumArtist   string            `json:"album_artist"`
	Genre         string            `json:"genre"`
	Year          int               `json:"year"`
	TrackNumber   int               `json:"track_number"`
	DiscNumber    int               `json:"disc_number"`
	Duration      int64             `json:"duration"`
	FilePath      string            `json:"file_path"`
	FileSize      int64             `json:"file_size"`
	Format        string            `json:"format"`
	Bitrate       int               `json:"bitrate"`
	SampleRate    int               `json:"sample_rate"`
	Channels      int               `json:"channels"`
	BPM           *int              `json:"bpm"`
	Key           *string           `json:"key"`
	Rating        *int              `json:"rating"`
	PlayCount     int64             `json:"play_count"`
	LastPlayed    *time.Time        `json:"last_played"`
	DateAdded     time.Time         `json:"date_added"`
	CoverArt      *CoverArt         `json:"cover_art"`
	Lyrics        *LyricsData       `json:"lyrics"`
	AudioFeatures *AudioFeatures    `json:"audio_features"`
	Waveform      *WaveformData     `json:"waveform"`
	Tags          map[string]string `json:"tags"`
	ReplayGain    *ReplayGainData   `json:"replay_gain"`
}

type AudioFeatures struct {
	Danceability     float64 `json:"danceability"`
	Energy           float64 `json:"energy"`
	Speechiness      float64 `json:"speechiness"`
	Acousticness     float64 `json:"acousticness"`
	Instrumentalness float64 `json:"instrumentalness"`
	Liveness         float64 `json:"liveness"`
	Valence          float64 `json:"valence"`
	Tempo            float64 `json:"tempo"`
	Loudness         float64 `json:"loudness"`
}

type WaveformData struct {
	ID       int64     `json:"id"`
	TrackID  int64     `json:"track_id"`
	Data     []float64 `json:"data"`
	Duration int64     `json:"duration"`
	Created  time.Time `json:"created"`
}

type ReplayGainData struct {
	TrackGain float64 `json:"track_gain"`
	TrackPeak float64 `json:"track_peak"`
	AlbumGain float64 `json:"album_gain"`
	AlbumPeak float64 `json:"album_peak"`
}

type BufferedRange struct {
	Start int64 `json:"start"`
	End   int64 `json:"end"`
}

type PlayMode string

const (
	PlayModeTrack    PlayMode = "track"
	PlayModeAlbum    PlayMode = "album"
	PlayModeArtist   PlayMode = "artist"
	PlayModePlaylist PlayMode = "playlist"
	PlayModeFolder   PlayMode = "folder"
	PlayModeGenre    PlayMode = "genre"
	PlayModeQueue    PlayMode = "queue"
)

type AudioQuality string

const (
	QualityLossless AudioQuality = "lossless" // FLAC/ALAC
)

type DeviceInfo struct {
	DeviceID   string `json:"device_id"`
	DeviceName string `json:"device_name"`
	DeviceType string `json:"device_type"`
	Platform   string `json:"platform"`
	AppVersion string `json:"app_version"`
}

type MusicLibraryStats struct {
	TotalTracks      int64            `json:"total_tracks"`
	TotalAlbums      int64            `json:"total_albums"`
	TotalArtists     int64            `json:"total_artists"`
	TotalGenres      int64            `json:"total_genres"`
	TotalDuration    int64            `json:"total_duration"`
	TotalSize        int64            `json:"total_size"`
	FormatBreakdown  map[string]int64 `json:"format_breakdown"`
	QualityBreakdown map[string]int64 `json:"quality_breakdown"`
	YearBreakdown    map[int]int64    `json:"year_breakdown"`
	TopGenres        []GenreStats     `json:"top_genres"`
	TopArtists       []ArtistStats    `json:"top_artists"`
	RecentlyAdded    []MusicTrack     `json:"recently_added"`
	MostPlayed       []MusicTrack     `json:"most_played"`
}

type Album struct {
	ID          int64        `json:"id"`
	Title       string       `json:"title"`
	Artist      string       `json:"artist"`
	AlbumArtist string       `json:"album_artist"`
	Year        int          `json:"year"`
	Genre       string       `json:"genre"`
	TrackCount  int          `json:"track_count"`
	Duration    int64        `json:"duration"`
	CoverArt    *CoverArt    `json:"cover_art"`
	Tracks      []MusicTrack `json:"tracks"`
	PlayCount   int64        `json:"play_count"`
	Rating      *int         `json:"rating"`
	DateAdded   time.Time    `json:"date_added"`
	LastPlayed  *time.Time   `json:"last_played"`
}

type Artist struct {
	ID         int64        `json:"id"`
	Name       string       `json:"name"`
	Biography  string       `json:"biography"`
	Country    string       `json:"country"`
	Genres     []string     `json:"genres"`
	Albums     []Album      `json:"albums"`
	TopTracks  []MusicTrack `json:"top_tracks"`
	TrackCount int          `json:"track_count"`
	AlbumCount int          `json:"album_count"`
	PlayCount  int64        `json:"play_count"`
	Followers  int64        `json:"followers"`
	CoverImage *CoverArt    `json:"cover_image"`
	DateAdded  time.Time    `json:"date_added"`
	LastPlayed *time.Time   `json:"last_played"`
}

type PlayTrackRequest struct {
	UserID     int64        `json:"user_id"`
	TrackID    int64        `json:"track_id"`
	PlayMode   PlayMode     `json:"play_mode"`
	StartTime  *int64       `json:"start_time"`
	Quality    AudioQuality `json:"quality"`
	DeviceInfo DeviceInfo   `json:"device_info"`
	PlaylistID *int64       `json:"playlist_id"`
	AlbumID    *int64       `json:"album_id"`
	ArtistID   *int64       `json:"artist_id"`
	FolderPath *string      `json:"folder_path"`
}

type PlayAlbumRequest struct {
	UserID     int64        `json:"user_id"`
	AlbumID    int64        `json:"album_id"`
	StartTrack *int         `json:"start_track"`
	Shuffle    bool         `json:"shuffle"`
	Quality    AudioQuality `json:"quality"`
	DeviceInfo DeviceInfo   `json:"device_info"`
}

type PlayArtistRequest struct {
	UserID     int64        `json:"user_id"`
	ArtistID   int64        `json:"artist_id"`
	Mode       string       `json:"mode"` // "top_tracks", "all_tracks", "albums"
	Shuffle    bool         `json:"shuffle"`
	Quality    AudioQuality `json:"quality"`
	DeviceInfo DeviceInfo   `json:"device_info"`
}

type UpdatePlaybackRequest struct {
	SessionID  string         `json:"session_id"`
	Position   *int64         `json:"position"`
	State      *PlaybackState `json:"state"`
	Volume     *float64       `json:"volume"`
	IsMuted    *bool          `json:"is_muted"`
	RepeatMode *RepeatMode    `json:"repeat_mode"`
	Shuffle    *bool          `json:"shuffle"`
}

type SeekRequest struct {
	SessionID string `json:"session_id"`
	Position  int64  `json:"position"`
}

type QueueRequest struct {
	SessionID string  `json:"session_id"`
	TrackIDs  []int64 `json:"track_ids"`
	Position  *int    `json:"position"`
}

func NewMusicPlayerService(
	db *sql.DB,
	logger *zap.Logger,
	mediaPlayerService *MediaPlayerService,
	playlistService *PlaylistService,
	positionService *PlaybackPositionService,
	lyricsService *LyricsService,
	coverArtService *CoverArtService,
	translationService *TranslationService,
) *MusicPlayerService {
	return &MusicPlayerService{
		db:                 db,
		logger:             logger,
		mediaPlayerService: mediaPlayerService,
		playlistService:    playlistService,
		positionService:    positionService,
		lyricsService:      lyricsService,
		coverArtService:    coverArtService,
		translationService: translationService,
	}
}

func (s *MusicPlayerService) PlayTrack(ctx context.Context, req *PlayTrackRequest) (*MusicPlaybackSession, error) {
	s.logger.Info("Starting track playback",
		zap.Int64("user_id", req.UserID),
		zap.Int64("track_id", req.TrackID),
		zap.String("play_mode", string(req.PlayMode)))

	track, err := s.getTrack(ctx, req.TrackID)
	if err != nil {
		return nil, fmt.Errorf("failed to get track: %w", err)
	}

	session := &MusicPlaybackSession{
		ID:                generateSessionID(),
		UserID:            req.UserID,
		CurrentTrack:      track,
		Queue:             []MusicTrack{*track},
		QueueIndex:        0,
		PlayMode:          req.PlayMode,
		RepeatMode:        RepeatModeOff,
		ShuffleEnabled:    false,
		Volume:            1.0,
		IsMuted:           false,
		Crossfade:         false,
		CrossfadeDuration: 3000,
		EqualizerPreset:   "flat",
		EqualizerBands:    make(map[string]float64),
		PlaybackState:     PlaybackStatePlaying,
		Position:          0,
		Duration:          track.Duration,
		PlaybackQuality:   req.Quality,
		DeviceInfo:        req.DeviceInfo,
		LastActivity:      time.Now(),
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	if req.StartTime != nil {
		session.Position = *req.StartTime
	}

	switch req.PlayMode {
	case PlayModeAlbum:
		if req.AlbumID != nil {
			if err := s.loadAlbumQueue(ctx, session, *req.AlbumID, track.ID); err != nil {
				s.logger.Warn("Failed to load album queue", zap.Error(err))
			}
		}
	case PlayModeArtist:
		if req.ArtistID != nil {
			if err := s.loadArtistQueue(ctx, session, *req.ArtistID, track.ID); err != nil {
				s.logger.Warn("Failed to load artist queue", zap.Error(err))
			}
		}
	case PlayModePlaylist:
		if req.PlaylistID != nil {
			session.PlaylistID = req.PlaylistID
			if err := s.loadPlaylistQueue(ctx, session, *req.PlaylistID, track.ID); err != nil {
				s.logger.Warn("Failed to load playlist queue", zap.Error(err))
			}
		}
	case PlayModeFolder:
		if req.FolderPath != nil {
			if err := s.loadFolderQueue(ctx, session, *req.FolderPath, track.ID); err != nil {
				s.logger.Warn("Failed to load folder queue", zap.Error(err))
			}
		}
	}

	if err := s.saveSession(ctx, session); err != nil {
		s.logger.Error("Failed to save session", zap.Error(err))
		return nil, fmt.Errorf("failed to save session: %w", err)
	}

	if err := s.recordPlayback(ctx, req.UserID, track.ID); err != nil {
		s.logger.Warn("Failed to record playback", zap.Error(err))
	}

	return session, nil
}

func (s *MusicPlayerService) PlayAlbum(ctx context.Context, req *PlayAlbumRequest) (*MusicPlaybackSession, error) {
	s.logger.Info("Starting album playback",
		zap.Int64("user_id", req.UserID),
		zap.Int64("album_id", req.AlbumID),
		zap.Bool("shuffle", req.Shuffle))

	album, err := s.getAlbumWithTracks(ctx, req.AlbumID)
	if err != nil {
		return nil, fmt.Errorf("failed to get album: %w", err)
	}

	if len(album.Tracks) == 0 {
		return nil, fmt.Errorf("album has no tracks")
	}

	startIndex := 0
	if req.StartTrack != nil && *req.StartTrack < len(album.Tracks) {
		startIndex = *req.StartTrack
	}

	session := &MusicPlaybackSession{
		ID:                generateSessionID(),
		UserID:            req.UserID,
		CurrentTrack:      &album.Tracks[startIndex],
		Queue:             album.Tracks,
		QueueIndex:        startIndex,
		PlayMode:          PlayModeAlbum,
		RepeatMode:        RepeatModeOff,
		ShuffleEnabled:    req.Shuffle,
		Volume:            1.0,
		IsMuted:           false,
		Crossfade:         true,
		CrossfadeDuration: 3000,
		EqualizerPreset:   "flat",
		EqualizerBands:    make(map[string]float64),
		PlaybackState:     PlaybackStatePlaying,
		Position:          0,
		Duration:          album.Tracks[startIndex].Duration,
		PlaybackQuality:   req.Quality,
		DeviceInfo:        req.DeviceInfo,
		LastActivity:      time.Now(),
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	if req.Shuffle {
		s.shuffleQueue(session)
	}

	if err := s.saveSession(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to save session: %w", err)
	}

	return session, nil
}

func (s *MusicPlayerService) PlayArtist(ctx context.Context, req *PlayArtistRequest) (*MusicPlaybackSession, error) {
	s.logger.Info("Starting artist playback",
		zap.Int64("user_id", req.UserID),
		zap.Int64("artist_id", req.ArtistID),
		zap.String("mode", req.Mode))

	var tracks []MusicTrack
	var err error

	switch req.Mode {
	case "top_tracks":
		tracks, err = s.getArtistTopTracks(ctx, req.ArtistID, 50)
	case "all_tracks":
		tracks, err = s.getArtistAllTracks(ctx, req.ArtistID)
	case "albums":
		albums, albumErr := s.getArtistAlbums(ctx, req.ArtistID)
		if albumErr != nil {
			return nil, fmt.Errorf("failed to get artist albums: %w", albumErr)
		}
		for _, album := range albums {
			albumTracks, trackErr := s.getAlbumTracks(ctx, album.ID)
			if trackErr == nil {
				tracks = append(tracks, albumTracks...)
			}
		}
	default:
		tracks, err = s.getArtistTopTracks(ctx, req.ArtistID, 50)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get artist tracks: %w", err)
	}

	if len(tracks) == 0 {
		return nil, fmt.Errorf("artist has no tracks")
	}

	session := &MusicPlaybackSession{
		ID:                generateSessionID(),
		UserID:            req.UserID,
		CurrentTrack:      &tracks[0],
		Queue:             tracks,
		QueueIndex:        0,
		PlayMode:          PlayModeArtist,
		RepeatMode:        RepeatModeOff,
		ShuffleEnabled:    req.Shuffle,
		Volume:            1.0,
		IsMuted:           false,
		Crossfade:         true,
		CrossfadeDuration: 3000,
		EqualizerPreset:   "flat",
		EqualizerBands:    make(map[string]float64),
		PlaybackState:     PlaybackStatePlaying,
		Position:          0,
		Duration:          tracks[0].Duration,
		PlaybackQuality:   req.Quality,
		DeviceInfo:        req.DeviceInfo,
		LastActivity:      time.Now(),
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	if req.Shuffle {
		s.shuffleQueue(session)
	}

	if err := s.saveSession(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to save session: %w", err)
	}

	return session, nil
}

func (s *MusicPlayerService) GetSession(ctx context.Context, sessionID string) (*MusicPlaybackSession, error) {
	s.logger.Debug("Getting playback session", zap.String("session_id", sessionID))

	query := `
		SELECT session_data, updated_at
		FROM music_playback_sessions
		WHERE id = $1 AND expires_at > NOW()
	`

	var sessionData string
	var updatedAt time.Time
	err := s.db.QueryRowContext(ctx, query, sessionID).Scan(&sessionData, &updatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("session not found or expired")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	var session MusicPlaybackSession
	if err := json.Unmarshal([]byte(sessionData), &session); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session: %w", err)
	}

	session.UpdatedAt = updatedAt
	return &session, nil
}

func (s *MusicPlayerService) UpdatePlayback(ctx context.Context, req *UpdatePlaybackRequest) (*MusicPlaybackSession, error) {
	s.logger.Debug("Updating playback", zap.String("session_id", req.SessionID))

	session, err := s.GetSession(ctx, req.SessionID)
	if err != nil {
		return nil, err
	}

	if req.Position != nil {
		session.Position = *req.Position
		if err := s.positionService.UpdatePosition(ctx, &UpdatePositionRequest{
			UserID:          session.UserID,
			MediaItemID:     session.CurrentTrack.ID,
			Position:        *req.Position,
			Duration:        session.Duration,
			DeviceInfo:      session.DeviceInfo.DeviceName,
			PlaybackQuality: string(session.PlaybackQuality),
		}); err != nil {
			s.logger.Warn("Failed to update position", zap.Error(err))
		}
	}

	if req.State != nil {
		session.PlaybackState = *req.State
	}

	if req.Volume != nil {
		session.Volume = *req.Volume
	}

	if req.IsMuted != nil {
		session.IsMuted = *req.IsMuted
	}

	if req.RepeatMode != nil {
		session.RepeatMode = *req.RepeatMode
	}

	if req.Shuffle != nil {
		session.ShuffleEnabled = *req.Shuffle
		if *req.Shuffle {
			s.shuffleQueue(session)
		} else {
			s.unshuffleQueue(session)
		}
	}

	session.LastActivity = time.Now()
	session.UpdatedAt = time.Now()

	if err := s.saveSession(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to save session: %w", err)
	}

	return session, nil
}

func (s *MusicPlayerService) NextTrack(ctx context.Context, sessionID string) (*MusicPlaybackSession, error) {
	s.logger.Debug("Skipping to next track", zap.String("session_id", sessionID))

	session, err := s.GetSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	nextIndex := s.getNextTrackIndex(session)
	if nextIndex == -1 {
		session.PlaybackState = PlaybackStateStopped
		return session, nil
	}

	session.QueueIndex = nextIndex
	session.CurrentTrack = &session.Queue[nextIndex]
	session.Position = 0
	session.Duration = session.CurrentTrack.Duration
	session.LastActivity = time.Now()
	session.UpdatedAt = time.Now()

	if err := s.saveSession(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to save session: %w", err)
	}

	if err := s.recordPlayback(ctx, session.UserID, session.CurrentTrack.ID); err != nil {
		s.logger.Warn("Failed to record playback", zap.Error(err))
	}

	return session, nil
}

func (s *MusicPlayerService) PreviousTrack(ctx context.Context, sessionID string) (*MusicPlaybackSession, error) {
	s.logger.Debug("Skipping to previous track", zap.String("session_id", sessionID))

	session, err := s.GetSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	if session.Position > 3000 {
		session.Position = 0
	} else {
		prevIndex := s.getPreviousTrackIndex(session)
		if prevIndex != -1 {
			session.QueueIndex = prevIndex
			session.CurrentTrack = &session.Queue[prevIndex]
			session.Duration = session.CurrentTrack.Duration
		}
		session.Position = 0
	}

	session.LastActivity = time.Now()
	session.UpdatedAt = time.Now()

	if err := s.saveSession(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to save session: %w", err)
	}

	return session, nil
}

func (s *MusicPlayerService) Seek(ctx context.Context, req *SeekRequest) (*MusicPlaybackSession, error) {
	s.logger.Debug("Seeking in track",
		zap.String("session_id", req.SessionID),
		zap.Int64("position", req.Position))

	session, err := s.GetSession(ctx, req.SessionID)
	if err != nil {
		return nil, err
	}

	if req.Position < 0 {
		req.Position = 0
	}
	if req.Position > session.Duration {
		req.Position = session.Duration
	}

	session.Position = req.Position
	session.LastActivity = time.Now()
	session.UpdatedAt = time.Now()

	if err := s.saveSession(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to save session: %w", err)
	}

	return session, nil
}

func (s *MusicPlayerService) AddToQueue(ctx context.Context, req *QueueRequest) (*MusicPlaybackSession, error) {
	s.logger.Info("Adding tracks to queue",
		zap.String("session_id", req.SessionID),
		zap.Int("track_count", len(req.TrackIDs)))

	session, err := s.GetSession(ctx, req.SessionID)
	if err != nil {
		return nil, err
	}

	tracks, err := s.getTracks(ctx, req.TrackIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get tracks: %w", err)
	}

	if req.Position != nil && *req.Position >= 0 && *req.Position <= len(session.Queue) {
		before := session.Queue[:*req.Position]
		after := session.Queue[*req.Position:]
		session.Queue = append(before, append(tracks, after...)...)

		if *req.Position <= session.QueueIndex {
			session.QueueIndex += len(tracks)
		}
	} else {
		session.Queue = append(session.Queue, tracks...)
	}

	session.LastActivity = time.Now()
	session.UpdatedAt = time.Now()

	if err := s.saveSession(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to save session: %w", err)
	}

	return session, nil
}

func (s *MusicPlayerService) GetLibraryStats(ctx context.Context, userID int64) (*MusicLibraryStats, error) {
	s.logger.Debug("Getting library statistics", zap.Int64("user_id", userID))

	stats := &MusicLibraryStats{
		FormatBreakdown:  make(map[string]int64),
		QualityBreakdown: make(map[string]int64),
		YearBreakdown:    make(map[int]int64),
		TopGenres:        make([]GenreStats, 0),
		TopArtists:       make([]ArtistStats, 0),
		RecentlyAdded:    make([]MusicTrack, 0),
		MostPlayed:       make([]MusicTrack, 0),
	}

	if err := s.getBasicStats(ctx, userID, stats); err != nil {
		return nil, err
	}

	if err := s.getFormatBreakdown(ctx, userID, stats); err != nil {
		s.logger.Warn("Failed to get format breakdown", zap.Error(err))
	}

	if err := s.getTopGenres(ctx, userID, stats); err != nil {
		s.logger.Warn("Failed to get top genres", zap.Error(err))
	}

	if err := s.getTopArtists(ctx, userID, stats); err != nil {
		s.logger.Warn("Failed to get top artists", zap.Error(err))
	}

	if err := s.getRecentlyAdded(ctx, userID, stats); err != nil {
		s.logger.Warn("Failed to get recently added", zap.Error(err))
	}

	if err := s.getMostPlayed(ctx, userID, stats); err != nil {
		s.logger.Warn("Failed to get most played", zap.Error(err))
	}

	return stats, nil
}

func (s *MusicPlayerService) SetEqualizer(ctx context.Context, sessionID string, preset string, bands map[string]float64) error {
	s.logger.Debug("Setting equalizer",
		zap.String("session_id", sessionID),
		zap.String("preset", preset))

	session, err := s.GetSession(ctx, sessionID)
	if err != nil {
		return err
	}

	session.EqualizerPreset = preset
	if bands != nil {
		session.EqualizerBands = bands
	}
	session.LastActivity = time.Now()
	session.UpdatedAt = time.Now()

	return s.saveSession(ctx, session)
}

func (s *MusicPlayerService) getTrack(ctx context.Context, trackID int64) (*MusicTrack, error) {
	query := `
		SELECT id, title, artist, album, album_artist, genre, year, track_number,
			   disc_number, duration, file_path, file_size, format, bitrate,
			   sample_rate, channels, bpm, key, rating, play_count, last_played,
			   date_added
		FROM media_items
		WHERE id = $1 AND type = 'audio'
	`

	var track MusicTrack
	var lastPlayed sql.NullTime
	var bpm sql.NullInt64
	var key sql.NullString
	var rating sql.NullInt64

	err := s.db.QueryRowContext(ctx, query, trackID).Scan(
		&track.ID, &track.Title, &track.Artist, &track.Album, &track.AlbumArtist,
		&track.Genre, &track.Year, &track.TrackNumber, &track.DiscNumber,
		&track.Duration, &track.FilePath, &track.FileSize, &track.Format,
		&track.Bitrate, &track.SampleRate, &track.Channels, &bpm, &key,
		&rating, &track.PlayCount, &lastPlayed, &track.DateAdded,
	)

	if err != nil {
		return nil, err
	}

	if bpm.Valid {
		bpmInt := int(bpm.Int64)
		track.BPM = &bpmInt
	}
	if key.Valid {
		track.Key = &key.String
	}
	if rating.Valid {
		ratingInt := int(rating.Int64)
		track.Rating = &ratingInt
	}
	if lastPlayed.Valid {
		track.LastPlayed = &lastPlayed.Time
	}

	return &track, nil
}

func (s *MusicPlayerService) getTracks(ctx context.Context, trackIDs []int64) ([]MusicTrack, error) {
	if len(trackIDs) == 0 {
		return []MusicTrack{}, nil
	}

	placeholders := make([]string, len(trackIDs))
	args := make([]interface{}, len(trackIDs))
	for i, id := range trackIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	query := fmt.Sprintf(`
		SELECT id, title, artist, album, album_artist, genre, year, track_number,
			   disc_number, duration, file_path, file_size, format, bitrate,
			   sample_rate, channels, bpm, key, rating, play_count, last_played,
			   date_added
		FROM media_items
		WHERE id IN (%s) AND type = 'audio'
		ORDER BY array_position(ARRAY[%s], id)
	`, strings.Join(placeholders, ","), strings.Join(placeholders, ","))

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tracks []MusicTrack
	for rows.Next() {
		var track MusicTrack
		var lastPlayed sql.NullTime
		var bpm sql.NullInt64
		var key sql.NullString
		var rating sql.NullInt64

		err := rows.Scan(
			&track.ID, &track.Title, &track.Artist, &track.Album, &track.AlbumArtist,
			&track.Genre, &track.Year, &track.TrackNumber, &track.DiscNumber,
			&track.Duration, &track.FilePath, &track.FileSize, &track.Format,
			&track.Bitrate, &track.SampleRate, &track.Channels, &bpm, &key,
			&rating, &track.PlayCount, &lastPlayed, &track.DateAdded,
		)

		if err != nil {
			continue
		}

		if bpm.Valid {
			bpmInt := int(bpm.Int64)
			track.BPM = &bpmInt
		}
		if key.Valid {
			track.Key = &key.String
		}
		if rating.Valid {
			ratingInt := int(rating.Int64)
			track.Rating = &ratingInt
		}
		if lastPlayed.Valid {
			track.LastPlayed = &lastPlayed.Time
		}

		tracks = append(tracks, track)
	}

	return tracks, nil
}

func (s *MusicPlayerService) getAlbumWithTracks(ctx context.Context, albumID int64) (*Album, error) {
	tracks, err := s.getAlbumTracks(ctx, albumID)
	if err != nil {
		return nil, err
	}

	if len(tracks) == 0 {
		return nil, fmt.Errorf("album not found")
	}

	album := &Album{
		ID:          albumID,
		Title:       tracks[0].Album,
		Artist:      tracks[0].Artist,
		AlbumArtist: tracks[0].AlbumArtist,
		Year:        tracks[0].Year,
		Genre:       tracks[0].Genre,
		TrackCount:  len(tracks),
		Tracks:      tracks,
		DateAdded:   tracks[0].DateAdded,
	}

	for _, track := range tracks {
		album.Duration += track.Duration
		album.PlayCount += track.PlayCount
		if track.LastPlayed != nil && (album.LastPlayed == nil || track.LastPlayed.After(*album.LastPlayed)) {
			album.LastPlayed = track.LastPlayed
		}
	}

	return album, nil
}

func (s *MusicPlayerService) getAlbumTracks(ctx context.Context, albumID int64) ([]MusicTrack, error) {
	query := `
		SELECT id, title, artist, album, album_artist, genre, year, track_number,
			   disc_number, duration, file_path, file_size, format, bitrate,
			   sample_rate, channels, bpm, key, rating, play_count, last_played,
			   date_added
		FROM media_items
		WHERE album_id = $1 AND type = 'audio'
		ORDER BY disc_number ASC, track_number ASC
	`

	rows, err := s.db.QueryContext(ctx, query, albumID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tracks []MusicTrack
	for rows.Next() {
		var track MusicTrack
		var lastPlayed sql.NullTime
		var bpm sql.NullInt64
		var key sql.NullString
		var rating sql.NullInt64

		err := rows.Scan(
			&track.ID, &track.Title, &track.Artist, &track.Album, &track.AlbumArtist,
			&track.Genre, &track.Year, &track.TrackNumber, &track.DiscNumber,
			&track.Duration, &track.FilePath, &track.FileSize, &track.Format,
			&track.Bitrate, &track.SampleRate, &track.Channels, &bpm, &key,
			&rating, &track.PlayCount, &lastPlayed, &track.DateAdded,
		)

		if err != nil {
			continue
		}

		if bpm.Valid {
			bpmInt := int(bpm.Int64)
			track.BPM = &bpmInt
		}
		if key.Valid {
			track.Key = &key.String
		}
		if rating.Valid {
			ratingInt := int(rating.Int64)
			track.Rating = &ratingInt
		}
		if lastPlayed.Valid {
			track.LastPlayed = &lastPlayed.Time
		}

		tracks = append(tracks, track)
	}

	return tracks, nil
}

func (s *MusicPlayerService) getArtistTopTracks(ctx context.Context, artistID int64, limit int) ([]MusicTrack, error) {
	query := `
		SELECT id, title, artist, album, album_artist, genre, year, track_number,
			   disc_number, duration, file_path, file_size, format, bitrate,
			   sample_rate, channels, bpm, key, rating, play_count, last_played,
			   date_added
		FROM media_items
		WHERE artist_id = $1 AND type = 'audio'
		ORDER BY play_count DESC, rating DESC
		LIMIT $2
	`

	rows, err := s.db.QueryContext(ctx, query, artistID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return s.scanTracks(rows)
}

func (s *MusicPlayerService) getArtistAllTracks(ctx context.Context, artistID int64) ([]MusicTrack, error) {
	query := `
		SELECT id, title, artist, album, album_artist, genre, year, track_number,
			   disc_number, duration, file_path, file_size, format, bitrate,
			   sample_rate, channels, bmp, key, rating, play_count, last_played,
			   date_added
		FROM media_items
		WHERE artist_id = $1 AND type = 'audio'
		ORDER BY year DESC, album ASC, disc_number ASC, track_number ASC
	`

	rows, err := s.db.QueryContext(ctx, query, artistID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return s.scanTracks(rows)
}

func (s *MusicPlayerService) getArtistAlbums(ctx context.Context, artistID int64) ([]Album, error) {
	return []Album{}, nil
}

func (s *MusicPlayerService) scanTracks(rows *sql.Rows) ([]MusicTrack, error) {
	var tracks []MusicTrack
	for rows.Next() {
		var track MusicTrack
		var lastPlayed sql.NullTime
		var bpm sql.NullInt64
		var key sql.NullString
		var rating sql.NullInt64

		err := rows.Scan(
			&track.ID, &track.Title, &track.Artist, &track.Album, &track.AlbumArtist,
			&track.Genre, &track.Year, &track.TrackNumber, &track.DiscNumber,
			&track.Duration, &track.FilePath, &track.FileSize, &track.Format,
			&track.Bitrate, &track.SampleRate, &track.Channels, &bpm, &key,
			&rating, &track.PlayCount, &lastPlayed, &track.DateAdded,
		)

		if err != nil {
			continue
		}

		if bpm.Valid {
			bpmInt := int(bpm.Int64)
			track.BPM = &bpmInt
		}
		if key.Valid {
			track.Key = &key.String
		}
		if rating.Valid {
			ratingInt := int(rating.Int64)
			track.Rating = &ratingInt
		}
		if lastPlayed.Valid {
			track.LastPlayed = &lastPlayed.Time
		}

		tracks = append(tracks, track)
	}

	return tracks, nil
}

func (s *MusicPlayerService) loadAlbumQueue(ctx context.Context, session *MusicPlaybackSession, albumID, currentTrackID int64) error {
	tracks, err := s.getAlbumTracks(ctx, albumID)
	if err != nil {
		return err
	}

	session.Queue = tracks
	for i, track := range tracks {
		if track.ID == currentTrackID {
			session.QueueIndex = i
			break
		}
	}

	return nil
}

func (s *MusicPlayerService) loadArtistQueue(ctx context.Context, session *MusicPlaybackSession, artistID, currentTrackID int64) error {
	tracks, err := s.getArtistTopTracks(ctx, artistID, 100)
	if err != nil {
		return err
	}

	session.Queue = tracks
	for i, track := range tracks {
		if track.ID == currentTrackID {
			session.QueueIndex = i
			break
		}
	}

	return nil
}

func (s *MusicPlayerService) loadPlaylistQueue(ctx context.Context, session *MusicPlaybackSession, playlistID, currentTrackID int64) error {
	playlistItems, err := s.playlistService.GetPlaylistItems(ctx, playlistID, session.UserID, 1000, 0)
	if err != nil {
		return err
	}

	var trackIDs []int64
	for _, item := range playlistItems {
		trackIDs = append(trackIDs, item.MediaItemID)
	}

	tracks, err := s.getTracks(ctx, trackIDs)
	if err != nil {
		return err
	}

	session.Queue = tracks
	for i, track := range tracks {
		if track.ID == currentTrackID {
			session.QueueIndex = i
			break
		}
	}

	return nil
}

func (s *MusicPlayerService) loadFolderQueue(ctx context.Context, session *MusicPlaybackSession, folderPath string, currentTrackID int64) error {
	query := `
		SELECT id, title, artist, album, album_artist, genre, year, track_number,
			   disc_number, duration, file_path, file_size, format, bitrate,
			   sample_rate, channels, bmp, key, rating, play_count, last_played,
			   date_added
		FROM media_items
		WHERE file_path LIKE $1 AND type = 'audio'
		ORDER BY file_path ASC
	`

	rows, err := s.db.QueryContext(ctx, query, folderPath+"%")
	if err != nil {
		return err
	}
	defer rows.Close()

	tracks, err := s.scanTracks(rows)
	if err != nil {
		return err
	}

	session.Queue = tracks
	for i, track := range tracks {
		if track.ID == currentTrackID {
			session.QueueIndex = i
			break
		}
	}

	return nil
}

func (s *MusicPlayerService) shuffleQueue(session *MusicPlaybackSession) {
	if len(session.Queue) <= 1 {
		return
	}

	currentTrack := session.Queue[session.QueueIndex]
	remainingTracks := make([]MusicTrack, 0, len(session.Queue)-1)

	for i, track := range session.Queue {
		if i != session.QueueIndex {
			remainingTracks = append(remainingTracks, track)
		}
	}

	rand.Shuffle(len(remainingTracks), func(i, j int) {
		remainingTracks[i], remainingTracks[j] = remainingTracks[j], remainingTracks[i]
	})

	session.Queue = append([]MusicTrack{currentTrack}, remainingTracks...)
	session.QueueIndex = 0
}

func (s *MusicPlayerService) unshuffleQueue(session *MusicPlaybackSession) {
	if len(session.ShuffleHistory) == 0 {
		return
	}

	sort.Slice(session.Queue, func(i, j int) bool {
		return session.Queue[i].ID < session.Queue[j].ID
	})

	for i, track := range session.Queue {
		if track.ID == session.CurrentTrack.ID {
			session.QueueIndex = i
			break
		}
	}

	session.ShuffleHistory = []int{}
}

func (s *MusicPlayerService) getNextTrackIndex(session *MusicPlaybackSession) int {
	if len(session.Queue) == 0 {
		return -1
	}

	switch session.RepeatMode {
	case RepeatModeTrack:
		return session.QueueIndex
	case RepeatModeAll, RepeatModeAlbum:
		if session.QueueIndex < len(session.Queue)-1 {
			return session.QueueIndex + 1
		}
		return 0
	default:
		if session.QueueIndex < len(session.Queue)-1 {
			return session.QueueIndex + 1
		}
		return -1
	}
}

func (s *MusicPlayerService) getPreviousTrackIndex(session *MusicPlaybackSession) int {
	if len(session.Queue) == 0 {
		return -1
	}

	if session.QueueIndex > 0 {
		return session.QueueIndex - 1
	}

	if session.RepeatMode == RepeatModeAll || session.RepeatMode == RepeatModeAlbum {
		return len(session.Queue) - 1
	}

	return -1
}

func (s *MusicPlayerService) saveSession(ctx context.Context, session *MusicPlaybackSession) error {
	sessionData, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	query := `
		INSERT INTO music_playback_sessions (id, user_id, session_data, expires_at, updated_at)
		VALUES ($1, $2, $3, NOW() + INTERVAL '24 hours', NOW())
		ON CONFLICT (id)
		DO UPDATE SET
			session_data = EXCLUDED.session_data,
			expires_at = NOW() + INTERVAL '24 hours',
			updated_at = NOW()
	`

	_, err = s.db.ExecContext(ctx, query, session.ID, session.UserID, string(sessionData))
	return err
}

func (s *MusicPlayerService) recordPlayback(ctx context.Context, userID, trackID int64) error {
	query := `
		UPDATE media_items
		SET play_count = play_count + 1, last_played = NOW()
		WHERE id = $1
	`

	_, err := s.db.ExecContext(ctx, query, trackID)
	return err
}

func (s *MusicPlayerService) getBasicStats(ctx context.Context, userID int64, stats *MusicLibraryStats) error {
	query := `
		SELECT
			COUNT(*) as total_tracks,
			COUNT(DISTINCT album) as total_albums,
			COUNT(DISTINCT artist) as total_artists,
			COUNT(DISTINCT genre) as total_genres,
			COALESCE(SUM(duration), 0) as total_duration,
			COALESCE(SUM(file_size), 0) as total_size
		FROM media_items
		WHERE type = 'audio' AND user_id = $1
	`

	return s.db.QueryRowContext(ctx, query, userID).Scan(
		&stats.TotalTracks, &stats.TotalAlbums, &stats.TotalArtists,
		&stats.TotalGenres, &stats.TotalDuration, &stats.TotalSize,
	)
}

func (s *MusicPlayerService) getFormatBreakdown(ctx context.Context, userID int64, stats *MusicLibraryStats) error {
	query := `
		SELECT format, COUNT(*)
		FROM media_items
		WHERE type = 'audio' AND user_id = $1
		GROUP BY format
		ORDER BY COUNT(*) DESC
	`

	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var format string
		var count int64
		if err := rows.Scan(&format, &count); err == nil {
			stats.FormatBreakdown[format] = count
		}
	}

	return nil
}

func (s *MusicPlayerService) getTopGenres(ctx context.Context, userID int64, stats *MusicLibraryStats) error {
	query := `
		SELECT genre, COUNT(*) as count, COALESCE(SUM(duration), 0) as duration
		FROM media_items
		WHERE type = 'audio' AND user_id = $1 AND genre != ''
		GROUP BY genre
		ORDER BY COUNT(*) DESC
		LIMIT 10
	`

	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var genre GenreStats
		if err := rows.Scan(&genre.Genre, &genre.Count, &genre.Duration); err == nil {
			stats.TopGenres = append(stats.TopGenres, genre)
		}
	}

	return nil
}

func (s *MusicPlayerService) getTopArtists(ctx context.Context, userID int64, stats *MusicLibraryStats) error {
	query := `
		SELECT artist, COUNT(*) as count, COALESCE(SUM(duration), 0) as duration
		FROM media_items
		WHERE type = 'audio' AND user_id = $1 AND artist != ''
		GROUP BY artist
		ORDER BY COUNT(*) DESC
		LIMIT 10
	`

	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var artist ArtistStats
		if err := rows.Scan(&artist.Artist, &artist.Count, &artist.Duration); err == nil {
			stats.TopArtists = append(stats.TopArtists, artist)
		}
	}

	return nil
}

func (s *MusicPlayerService) getRecentlyAdded(ctx context.Context, userID int64, stats *MusicLibraryStats) error {
	query := `
		SELECT id, title, artist, album, album_artist, genre, year, track_number,
			   disc_number, duration, file_path, file_size, format, bitrate,
			   sample_rate, channels, bmp, key, rating, play_count, last_played,
			   date_added
		FROM media_items
		WHERE type = 'audio' AND user_id = $1
		ORDER BY date_added DESC
		LIMIT 20
	`

	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return err
	}
	defer rows.Close()

	tracks, err := s.scanTracks(rows)
	if err != nil {
		return err
	}

	stats.RecentlyAdded = tracks
	return nil
}

func (s *MusicPlayerService) getMostPlayed(ctx context.Context, userID int64, stats *MusicLibraryStats) error {
	query := `
		SELECT id, title, artist, album, album_artist, genre, year, track_number,
			   disc_number, duration, file_path, file_size, format, bitrate,
			   sample_rate, channels, bmp, key, rating, play_count, last_played,
			   date_added
		FROM media_items
		WHERE type = 'audio' AND user_id = $1 AND play_count > 0
		ORDER BY play_count DESC
		LIMIT 20
	`

	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return err
	}
	defer rows.Close()

	tracks, err := s.scanTracks(rows)
	if err != nil {
		return err
	}

	stats.MostPlayed = tracks
	return nil
}
