package services

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewMusicPlayerService(t *testing.T) {
	var mockDB *sql.DB
	mockLogger := zap.NewNop()

	service := NewMusicPlayerService(mockDB, mockLogger, nil, nil, nil, nil, nil, nil)

	assert.NotNil(t, service)
}

func TestMusicPlayerService_GetNextTrackIndex(t *testing.T) {
	var mockDB *sql.DB
	mockLogger := zap.NewNop()
	service := NewMusicPlayerService(mockDB, mockLogger, nil, nil, nil, nil, nil, nil)

	tests := []struct {
		name       string
		session    *MusicPlaybackSession
		expected   int
	}{
		{
			name: "next track in queue",
			session: &MusicPlaybackSession{
				Queue:      []MusicTrack{{ID: 1}, {ID: 2}, {ID: 3}},
				QueueIndex: 0,
				RepeatMode: RepeatModeOff,
			},
			expected: 1,
		},
		{
			name: "end of queue no repeat",
			session: &MusicPlaybackSession{
				Queue:      []MusicTrack{{ID: 1}, {ID: 2}, {ID: 3}},
				QueueIndex: 2,
				RepeatMode: RepeatModeOff,
			},
			expected: -1,
		},
		{
			name: "repeat track mode",
			session: &MusicPlaybackSession{
				Queue:      []MusicTrack{{ID: 1}, {ID: 2}, {ID: 3}},
				QueueIndex: 1,
				RepeatMode: RepeatModeTrack,
			},
			expected: 1,
		},
		{
			name: "repeat all wraps around",
			session: &MusicPlaybackSession{
				Queue:      []MusicTrack{{ID: 1}, {ID: 2}, {ID: 3}},
				QueueIndex: 2,
				RepeatMode: RepeatModeAll,
			},
			expected: 0,
		},
		{
			name: "repeat all mid queue",
			session: &MusicPlaybackSession{
				Queue:      []MusicTrack{{ID: 1}, {ID: 2}, {ID: 3}},
				QueueIndex: 1,
				RepeatMode: RepeatModeAll,
			},
			expected: 2,
		},
		{
			name: "empty queue",
			session: &MusicPlaybackSession{
				Queue:      []MusicTrack{},
				QueueIndex: 0,
				RepeatMode: RepeatModeOff,
			},
			expected: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.getNextTrackIndex(tt.session)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMusicPlayerService_GetPreviousTrackIndex(t *testing.T) {
	var mockDB *sql.DB
	mockLogger := zap.NewNop()
	service := NewMusicPlayerService(mockDB, mockLogger, nil, nil, nil, nil, nil, nil)

	tests := []struct {
		name       string
		session    *MusicPlaybackSession
		expected   int
	}{
		{
			name: "previous track in queue",
			session: &MusicPlaybackSession{
				Queue:      []MusicTrack{{ID: 1}, {ID: 2}, {ID: 3}},
				QueueIndex: 2,
				RepeatMode: RepeatModeOff,
			},
			expected: 1,
		},
		{
			name: "beginning of queue no repeat",
			session: &MusicPlaybackSession{
				Queue:      []MusicTrack{{ID: 1}, {ID: 2}, {ID: 3}},
				QueueIndex: 0,
				RepeatMode: RepeatModeOff,
			},
			expected: -1,
		},
		{
			name: "repeat all wraps to end",
			session: &MusicPlaybackSession{
				Queue:      []MusicTrack{{ID: 1}, {ID: 2}, {ID: 3}},
				QueueIndex: 0,
				RepeatMode: RepeatModeAll,
			},
			expected: 2,
		},
		{
			name: "empty queue",
			session: &MusicPlaybackSession{
				Queue:      []MusicTrack{},
				QueueIndex: 0,
				RepeatMode: RepeatModeOff,
			},
			expected: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.getPreviousTrackIndex(tt.session)
			assert.Equal(t, tt.expected, result)
		})
	}
}
