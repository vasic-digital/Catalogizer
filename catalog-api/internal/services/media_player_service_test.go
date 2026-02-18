package services

import (
	"catalogizer/database"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewMediaPlayerService(t *testing.T) {
	mockDB := database.WrapDB(nil, database.DialectSQLite)
	mockLogger := zap.NewNop()

	service := NewMediaPlayerService(mockDB, mockLogger)

	assert.NotNil(t, service)
}

func TestMediaPlayerService_GetSupportedMediaTypes(t *testing.T) {
	mockDB := database.WrapDB(nil, database.DialectSQLite)
	mockLogger := zap.NewNop()
	service := NewMediaPlayerService(mockDB, mockLogger)

	types := service.GetSupportedMediaTypes()
	assert.NotEmpty(t, types)
	assert.Contains(t, types, MediaTypeMusic)
	assert.Contains(t, types, MediaTypeVideo)
}

func TestGetMediaTypeFromExtension(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		expected MediaType
	}{
		{
			name:     "mp3 file",
			filename: "song.mp3",
			expected: MediaTypeMusic,
		},
		{
			name:     "flac file",
			filename: "album.flac",
			expected: MediaTypeMusic,
		},
		{
			name:     "mp4 video",
			filename: "movie.mp4",
			expected: MediaTypeVideo,
		},
		{
			name:     "mkv video",
			filename: "show.mkv",
			expected: MediaTypeVideo,
		},
		{
			name:     "exe game",
			filename: "game.exe",
			expected: MediaTypeGame,
		},
		{
			name:     "epub ebook",
			filename: "book.epub",
			expected: MediaTypeEbook,
		},
		{
			name:     "pdf document",
			filename: "report.pdf",
			expected: MediaTypeDocument,
		},
		{
			name:     "unknown extension fallback",
			filename: "file.xyz",
			expected: MediaTypeSoftware,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetMediaTypeFromExtension(tt.filename)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMediaPlayerService_FindSubtitleByLanguage(t *testing.T) {
	mockDB := database.WrapDB(nil, database.DialectSQLite)
	mockLogger := zap.NewNop()
	service := NewMediaPlayerService(mockDB, mockLogger)

	subtitles := []SubtitleTrack{
		{ID: "1", Language: "English", LanguageCode: "en"},
		{ID: "2", Language: "French", LanguageCode: "fr"},
		{ID: "3", Language: "Spanish", LanguageCode: "es"},
	}

	tests := []struct {
		name     string
		lang     string
		expected *string
	}{
		{
			name:     "find by language name",
			lang:     "English",
			expected: strPtr("1"),
		},
		{
			name:     "find by language code",
			lang:     "fr",
			expected: strPtr("2"),
		},
		{
			name:     "not found",
			lang:     "German",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.findSubtitleByLanguage(subtitles, tt.lang)
			if tt.expected != nil {
				assert.NotNil(t, result)
				assert.Equal(t, *tt.expected, result.ID)
			} else {
				assert.Nil(t, result)
			}
		})
	}
}

func TestMediaPlayerService_FindSubtitleByID(t *testing.T) {
	mockDB := database.WrapDB(nil, database.DialectSQLite)
	mockLogger := zap.NewNop()
	service := NewMediaPlayerService(mockDB, mockLogger)

	subtitles := []SubtitleTrack{
		{ID: "sub_1", Language: "English"},
		{ID: "sub_2", Language: "French"},
	}

	tests := []struct {
		name     string
		id       string
		wantNil  bool
	}{
		{
			name:    "existing subtitle",
			id:      "sub_1",
			wantNil: false,
		},
		{
			name:    "non-existing subtitle",
			id:      "sub_999",
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.findSubtitleByID(subtitles, tt.id)
			if tt.wantNil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.Equal(t, tt.id, result.ID)
			}
		})
	}
}

func TestGetFloatValue(t *testing.T) {
	val1 := 1.5
	val2 := 2.5
	val3 := 0.0

	tests := []struct {
		name     string
		values   []*float64
		expected float64
	}{
		{
			name:     "first non-nil",
			values:   []*float64{&val1, &val2},
			expected: 1.5,
		},
		{
			name:     "skip nil to second",
			values:   []*float64{nil, &val2},
			expected: 2.5,
		},
		{
			name:     "all nil returns zero",
			values:   []*float64{nil, nil},
			expected: 0.0,
		},
		{
			name:     "zero value still returned",
			values:   []*float64{&val3, &val1},
			expected: 0.0,
		},
		{
			name:     "empty values",
			values:   []*float64{},
			expected: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getFloatValue(tt.values...)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGenerateSessionID(t *testing.T) {
	id1 := generateSessionID()
	id2 := generateSessionID()

	assert.NotEmpty(t, id1)
	assert.NotEmpty(t, id2)
	assert.Contains(t, id1, "session_")
}

func strPtr(s string) *string {
	return &s
}
