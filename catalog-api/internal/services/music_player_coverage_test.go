package services

import (
	"catalogizer/database"
	"context"
	"database/sql"
	"encoding/json"
	"testing"

	_ "github.com/mutecomm/go-sqlcipher"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func setupMusicPlayerCoverageDB(t *testing.T) (*database.DB, *MusicPlayerService) {
	t.Helper()
	sqlDB, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)

	db := database.WrapDB(sqlDB, database.DialectSQLite)

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS music_playback_sessions (
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
	`)
	require.NoError(t, err)

	logger := zap.NewNop()
	svc := NewMusicPlayerService(db, logger, nil, nil, nil, nil, nil, nil)
	return db, svc
}

func insertMusicSession(t *testing.T, db *database.DB, session *MusicPlaybackSession) {
	t.Helper()
	sessionData, err := json.Marshal(session)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO music_playback_sessions (id, user_id, session_data, expires_at)
		VALUES (?, ?, ?, datetime('now', '+1 hour'))`, session.ID, session.UserID, string(sessionData))
	require.NoError(t, err)
}

// =============================================================================
// GetSession
// =============================================================================

func TestMusicPlayerCov_GetSession_NotFound(t *testing.T) {
	_, svc := setupMusicPlayerCoverageDB(t)
	_, err := svc.GetSession(context.Background(), "nonexistent")
	assert.Error(t, err)
}

func TestMusicPlayerCov_GetSession_Valid(t *testing.T) {
	db, svc := setupMusicPlayerCoverageDB(t)

	session := &MusicPlaybackSession{
		ID: "test-music-sess", UserID: 1,
		CurrentTrack:  &MusicTrack{ID: 1, Title: "Test"},
		PlaybackState: PlaybackStatePlaying, Position: 5000, Duration: 240000,
	}
	insertMusicSession(t, db, session)

	result, err := svc.GetSession(context.Background(), "test-music-sess")
	require.NoError(t, err)
	assert.Equal(t, int64(5000), result.Position)
}

func TestMusicPlayerCov_GetSession_Expired(t *testing.T) {
	db, svc := setupMusicPlayerCoverageDB(t)

	sessionData, _ := json.Marshal(&MusicPlaybackSession{ID: "expired", UserID: 1})
	_, err := db.Exec(`INSERT INTO music_playback_sessions (id, user_id, session_data, expires_at)
		VALUES (?, ?, ?, datetime('now', '-1 hour'))`, "expired", 1, string(sessionData))
	require.NoError(t, err)

	_, err = svc.GetSession(context.Background(), "expired")
	assert.Error(t, err)
}

// =============================================================================
// Seek
// =============================================================================

func TestMusicPlayerCov_Seek_NoSession(t *testing.T) {
	_, svc := setupMusicPlayerCoverageDB(t)
	_, err := svc.Seek(context.Background(), &SeekRequest{SessionID: "no-session", Position: 5000})
	assert.Error(t, err)
}

func TestMusicPlayerCov_Seek_Valid(t *testing.T) {
	db, svc := setupMusicPlayerCoverageDB(t)

	session := &MusicPlaybackSession{
		ID: "seek-sess", UserID: 1,
		CurrentTrack:  &MusicTrack{ID: 1, Title: "Track", Duration: 240000},
		PlaybackState: PlaybackStatePlaying, Duration: 240000,
	}
	insertMusicSession(t, db, session)

	result, err := svc.Seek(context.Background(), &SeekRequest{SessionID: "seek-sess", Position: 120000})
	require.NoError(t, err)
	assert.Equal(t, int64(120000), result.Position)
}

func TestMusicPlayerCov_Seek_Negative(t *testing.T) {
	db, svc := setupMusicPlayerCoverageDB(t)

	session := &MusicPlaybackSession{
		ID: "seek-neg", UserID: 1,
		CurrentTrack:  &MusicTrack{ID: 1, Duration: 240000},
		PlaybackState: PlaybackStatePlaying, Duration: 240000,
	}
	insertMusicSession(t, db, session)

	result, err := svc.Seek(context.Background(), &SeekRequest{SessionID: "seek-neg", Position: -100})
	require.NoError(t, err)
	assert.Equal(t, int64(0), result.Position)
}

func TestMusicPlayerCov_Seek_BeyondDuration(t *testing.T) {
	db, svc := setupMusicPlayerCoverageDB(t)

	session := &MusicPlaybackSession{
		ID: "seek-beyond", UserID: 1,
		CurrentTrack:  &MusicTrack{ID: 1, Duration: 240000},
		PlaybackState: PlaybackStatePlaying, Duration: 240000,
	}
	insertMusicSession(t, db, session)

	result, err := svc.Seek(context.Background(), &SeekRequest{SessionID: "seek-beyond", Position: 999999})
	require.NoError(t, err)
	assert.Equal(t, int64(240000), result.Position)
}

// =============================================================================
// SetEqualizer
// =============================================================================

func TestMusicPlayerCov_SetEqualizer_NoSession(t *testing.T) {
	_, svc := setupMusicPlayerCoverageDB(t)
	err := svc.SetEqualizer(context.Background(), "no-session", "rock", nil)
	assert.Error(t, err)
}

func TestMusicPlayerCov_SetEqualizer_Valid(t *testing.T) {
	db, svc := setupMusicPlayerCoverageDB(t)

	session := &MusicPlaybackSession{
		ID: "eq-sess", UserID: 1,
		CurrentTrack:    &MusicTrack{ID: 1, Title: "Track"},
		PlaybackState:   PlaybackStatePlaying,
		EqualizerPreset: "flat", EqualizerBands: make(map[string]float64),
	}
	insertMusicSession(t, db, session)

	bands := map[string]float64{"60Hz": 3.0, "230Hz": 1.5}
	err := svc.SetEqualizer(context.Background(), "eq-sess", "rock", bands)
	assert.NoError(t, err)
}

func TestMusicPlayerCov_SetEqualizer_NilBands(t *testing.T) {
	db, svc := setupMusicPlayerCoverageDB(t)

	session := &MusicPlaybackSession{
		ID: "eq-nil", UserID: 1,
		CurrentTrack:    &MusicTrack{ID: 1},
		PlaybackState:   PlaybackStatePlaying,
		EqualizerPreset: "flat", EqualizerBands: make(map[string]float64),
	}
	insertMusicSession(t, db, session)

	err := svc.SetEqualizer(context.Background(), "eq-nil", "jazz", nil)
	assert.NoError(t, err)
}

// =============================================================================
// UpdatePlayback — without position update (avoids nil positionService)
// =============================================================================

func TestMusicPlayerCov_UpdatePlayback_NoSession(t *testing.T) {
	_, svc := setupMusicPlayerCoverageDB(t)
	_, err := svc.UpdatePlayback(context.Background(), &UpdatePlaybackRequest{SessionID: "no-session"})
	assert.Error(t, err)
}

func TestMusicPlayerCov_UpdatePlayback_StateOnly(t *testing.T) {
	db, svc := setupMusicPlayerCoverageDB(t)

	session := &MusicPlaybackSession{
		ID: "upd-sess", UserID: 1,
		CurrentTrack:  &MusicTrack{ID: 1, Title: "Track", Duration: 240000},
		PlaybackState: PlaybackStatePlaying, Volume: 1.0, Duration: 240000,
	}
	insertMusicSession(t, db, session)

	vol := 0.5
	muted := true
	paused := PlaybackStatePaused
	repeat := RepeatModeAll
	shuffle := true

	req := &UpdatePlaybackRequest{
		SessionID:  "upd-sess",
		State:      &paused,
		Volume:     &vol,
		IsMuted:    &muted,
		RepeatMode: &repeat,
		Shuffle:    &shuffle,
	}
	result, err := svc.UpdatePlayback(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, PlaybackStatePaused, result.PlaybackState)
	assert.Equal(t, 0.5, result.Volume)
	assert.True(t, result.IsMuted)
}

// =============================================================================
// AddToQueue — error path only (no session)
// =============================================================================

func TestMusicPlayerCov_AddToQueue_NoSession(t *testing.T) {
	_, svc := setupMusicPlayerCoverageDB(t)
	_, err := svc.AddToQueue(context.Background(), &QueueRequest{SessionID: "nonexistent", TrackIDs: []int64{1}})
	assert.Error(t, err)
}

// =============================================================================
// NextTrack / PreviousTrack — session-based
// =============================================================================

func TestMusicPlayerCov_NextTrack_WithSession(t *testing.T) {
	db, svc := setupMusicPlayerCoverageDB(t)

	session := &MusicPlaybackSession{
		ID: "next-sess", UserID: 1,
		CurrentTrack: &MusicTrack{ID: 1, Title: "Track 1"},
		Queue: []MusicTrack{
			{ID: 1, Title: "Track 1", Duration: 200000},
			{ID: 2, Title: "Track 2", Duration: 180000},
		},
		QueueIndex: 0, PlaybackState: PlaybackStatePlaying, Duration: 200000,
	}
	insertMusicSession(t, db, session)

	result, err := svc.NextTrack(context.Background(), "next-sess")
	require.NoError(t, err)
	assert.Equal(t, 1, result.QueueIndex)
	assert.Equal(t, "Track 2", result.CurrentTrack.Title)
}

func TestMusicPlayerCov_NextTrack_AtEnd(t *testing.T) {
	db, svc := setupMusicPlayerCoverageDB(t)

	session := &MusicPlaybackSession{
		ID: "next-end", UserID: 1,
		CurrentTrack: &MusicTrack{ID: 1, Title: "Only"},
		Queue:        []MusicTrack{{ID: 1, Title: "Only", Duration: 200000}},
		QueueIndex:   0, PlaybackState: PlaybackStatePlaying,
	}
	insertMusicSession(t, db, session)

	result, err := svc.NextTrack(context.Background(), "next-end")
	require.NoError(t, err)
	assert.Equal(t, PlaybackStateStopped, result.PlaybackState)
}

func TestMusicPlayerCov_NextTrack_NoSession(t *testing.T) {
	_, svc := setupMusicPlayerCoverageDB(t)
	_, err := svc.NextTrack(context.Background(), "nope")
	assert.Error(t, err)
}

func TestMusicPlayerCov_PreviousTrack_WithSession(t *testing.T) {
	db, svc := setupMusicPlayerCoverageDB(t)

	session := &MusicPlaybackSession{
		ID: "prev-sess", UserID: 1,
		CurrentTrack: &MusicTrack{ID: 2, Title: "Track 2"},
		Queue: []MusicTrack{
			{ID: 1, Title: "Track 1", Duration: 200000},
			{ID: 2, Title: "Track 2", Duration: 180000},
		},
		QueueIndex: 1, PlaybackState: PlaybackStatePlaying, Duration: 180000,
	}
	insertMusicSession(t, db, session)

	result, err := svc.PreviousTrack(context.Background(), "prev-sess")
	require.NoError(t, err)
	assert.Equal(t, 0, result.QueueIndex)
}

func TestMusicPlayerCov_PreviousTrack_AtStart(t *testing.T) {
	db, svc := setupMusicPlayerCoverageDB(t)

	session := &MusicPlaybackSession{
		ID: "prev-start", UserID: 1,
		CurrentTrack: &MusicTrack{ID: 1, Title: "First"},
		Queue:        []MusicTrack{{ID: 1, Title: "First", Duration: 200000}},
		QueueIndex:   0, PlaybackState: PlaybackStatePlaying, Position: 5000,
	}
	insertMusicSession(t, db, session)

	result, err := svc.PreviousTrack(context.Background(), "prev-start")
	require.NoError(t, err)
	assert.Equal(t, int64(0), result.Position)
}

func TestMusicPlayerCov_PreviousTrack_NoSession(t *testing.T) {
	_, svc := setupMusicPlayerCoverageDB(t)
	_, err := svc.PreviousTrack(context.Background(), "nope")
	assert.Error(t, err)
}
