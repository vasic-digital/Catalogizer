package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewMovieRecognitionProvider(t *testing.T) {
	mockLogger := zap.NewNop()
	provider := NewMovieRecognitionProvider(mockLogger)

	assert.NotNil(t, provider)
}

func TestMovieRecognitionProvider_GetProviderName(t *testing.T) {
	mockLogger := zap.NewNop()
	provider := NewMovieRecognitionProvider(mockLogger)

	name := provider.GetProviderName()
	assert.Equal(t, "movie_recognition", name)
}

func TestMovieRecognitionProvider_SupportsMediaType(t *testing.T) {
	mockLogger := zap.NewNop()
	provider := NewMovieRecognitionProvider(mockLogger)

	tests := []struct {
		name      string
		mediaType MediaType
		expected  bool
	}{
		{
			name:      "supports movie",
			mediaType: MediaTypeMovie,
			expected:  true,
		},
		{
			name:      "supports tv series",
			mediaType: MediaTypeTVSeries,
			expected:  true,
		},
		{
			name:      "supports tv episode",
			mediaType: MediaTypeTVEpisode,
			expected:  true,
		},
		{
			name:      "supports concert",
			mediaType: MediaTypeConcert,
			expected:  true,
		},
		{
			name:      "supports documentary",
			mediaType: MediaTypeDocumentary,
			expected:  true,
		},
		{
			name:      "does not support music",
			mediaType: MediaTypeMusic,
			expected:  false,
		},
		{
			name:      "does not support ebook",
			mediaType: MediaTypeEbook,
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := provider.SupportsMediaType(tt.mediaType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMovieRecognitionProvider_GetConfidenceThreshold(t *testing.T) {
	mockLogger := zap.NewNop()
	provider := NewMovieRecognitionProvider(mockLogger)

	threshold := provider.GetConfidenceThreshold()
	assert.Equal(t, 0.4, threshold)
}

func TestMovieRecognitionProvider_ExtractTitleFromFilename(t *testing.T) {
	mockLogger := zap.NewNop()
	provider := NewMovieRecognitionProvider(mockLogger)

	tests := []struct {
		name     string
		filename string
	}{
		{
			name:     "movie with year and quality",
			filename: "The.Matrix.1999.1080p.BluRay.x264.mp4",
		},
		{
			name:     "movie with dots",
			filename: "Inception.2010.mkv",
		},
		{
			name:     "tv show with season episode",
			filename: "Breaking.Bad.S01E01.720p.mkv",
		},
		{
			name:     "simple filename",
			filename: "movie.avi",
		},
		{
			name:     "empty filename",
			filename: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := provider.extractTitleFromFilename(tt.filename)
			assert.NotNil(t, result)
		})
	}
}

func TestMovieRecognitionProvider_ExtractYearFromFilename(t *testing.T) {
	mockLogger := zap.NewNop()
	provider := NewMovieRecognitionProvider(mockLogger)

	tests := []struct {
		name     string
		filename string
		expected int
	}{
		{
			name:     "year in filename",
			filename: "The.Matrix.1999.1080p.mkv",
			expected: 1999,
		},
		{
			name:     "year 2024",
			filename: "Dune.Part.Two.2024.mp4",
			expected: 2024,
		},
		{
			name:     "no year in filename",
			filename: "movie.mkv",
			expected: 0,
		},
		{
			name:     "empty filename",
			filename: "",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := provider.extractYearFromFilename(tt.filename)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMovieRecognitionProvider_ExtractSeasonEpisode(t *testing.T) {
	mockLogger := zap.NewNop()
	provider := NewMovieRecognitionProvider(mockLogger)

	tests := []struct {
		name            string
		filename        string
		expectedSeason  int
		expectedEpisode int
	}{
		{
			name:            "S01E01 format",
			filename:        "Breaking.Bad.S01E01.mp4",
			expectedSeason:  1,
			expectedEpisode: 1,
		},
		{
			name:            "S05E16 format",
			filename:        "Breaking.Bad.S05E16.Felina.mp4",
			expectedSeason:  5,
			expectedEpisode: 16,
		},
		{
			name:            "1x01 format",
			filename:        "Show.1x01.avi",
			expectedSeason:  1,
			expectedEpisode: 1,
		},
		{
			name:            "no season episode",
			filename:        "Movie.2024.mp4",
			expectedSeason:  0,
			expectedEpisode: 0,
		},
		{
			name:            "empty filename",
			filename:        "",
			expectedSeason:  0,
			expectedEpisode: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			season, episode := provider.extractSeasonEpisode(tt.filename)
			assert.Equal(t, tt.expectedSeason, season)
			assert.Equal(t, tt.expectedEpisode, episode)
		})
	}
}

func TestMovieRecognitionProvider_GetFileExtension(t *testing.T) {
	mockLogger := zap.NewNop()
	provider := NewMovieRecognitionProvider(mockLogger)

	tests := []struct {
		name     string
		filename string
		expected string
	}{
		{
			name:     "mp4 extension",
			filename: "movie.mp4",
			expected: "mp4",
		},
		{
			name:     "mkv extension",
			filename: "video.mkv",
			expected: "mkv",
		},
		{
			name:     "no extension",
			filename: "noext",
			expected: "",
		},
		{
			name:     "multiple dots",
			filename: "my.movie.2024.avi",
			expected: "avi",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := provider.getFileExtension(tt.filename)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMovieRecognitionProvider_ParseYear(t *testing.T) {
	mockLogger := zap.NewNop()
	provider := NewMovieRecognitionProvider(mockLogger)

	tests := []struct {
		name     string
		dateStr  string
		expected int
	}{
		{
			name:     "full date",
			dateStr:  "2024-01-15",
			expected: 2024,
		},
		{
			name:     "year only",
			dateStr:  "1999",
			expected: 1999,
		},
		{
			name:     "empty string",
			dateStr:  "",
			expected: 0,
		},
		{
			name:     "short string",
			dateStr:  "20",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := provider.parseYear(tt.dateStr)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMovieRecognitionProvider_ParseRuntime(t *testing.T) {
	mockLogger := zap.NewNop()
	provider := NewMovieRecognitionProvider(mockLogger)

	tests := []struct {
		name     string
		runtime  string
		expected int64
	}{
		{
			name:     "standard runtime",
			runtime:  "120 min",
			expected: 7200,
		},
		{
			name:     "short runtime",
			runtime:  "30 min",
			expected: 1800,
		},
		{
			name:     "no match",
			runtime:  "N/A",
			expected: 0,
		},
		{
			name:     "empty string",
			runtime:  "",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := provider.parseRuntime(tt.runtime)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMovieRecognitionProvider_CalculateConfidence(t *testing.T) {
	mockLogger := zap.NewNop()
	provider := NewMovieRecognitionProvider(mockLogger)

	tests := []struct {
		name          string
		title         string
		rating        float64
		voteCount     int
		minConfidence float64
	}{
		{
			name:          "high rated popular movie",
			title:         "The Shawshank Redemption",
			rating:        9.3,
			voteCount:     2500000,
			minConfidence: 0.9,
		},
		{
			name:          "medium rated movie",
			title:         "Some Movie",
			rating:        6.5,
			voteCount:     500,
			minConfidence: 0.7,
		},
		{
			name:          "low rated with few votes",
			title:         "Unknown Movie",
			rating:        3.0,
			voteCount:     100,
			minConfidence: 0.5,
		},
		{
			name:          "empty title",
			title:         "",
			rating:        8.0,
			voteCount:     5000,
			minConfidence: 0.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			confidence := provider.calculateConfidence(tt.title, tt.rating, tt.voteCount)
			assert.GreaterOrEqual(t, confidence, tt.minConfidence)
			assert.LessOrEqual(t, confidence, 1.0)
		})
	}
}

func TestMovieRecognitionProvider_CalculateOMDbConfidence(t *testing.T) {
	mockLogger := zap.NewNop()
	provider := NewMovieRecognitionProvider(mockLogger)

	tests := []struct {
		name          string
		rating        string
		votes         string
		minConfidence float64
	}{
		{
			name:          "high rated popular",
			rating:        "8.5",
			votes:         "1,500,000",
			minConfidence: 0.7,
		},
		{
			name:          "low rated",
			rating:        "4.0",
			votes:         "100",
			minConfidence: 0.4,
		},
		{
			name:          "invalid rating",
			rating:        "N/A",
			votes:         "N/A",
			minConfidence: 0.4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			confidence := provider.calculateOMDbConfidence(tt.rating, tt.votes)
			assert.GreaterOrEqual(t, confidence, tt.minConfidence)
			assert.LessOrEqual(t, confidence, 1.0)
		})
	}
}

func TestMovieRecognitionProvider_MapOMDbType(t *testing.T) {
	mockLogger := zap.NewNop()
	provider := NewMovieRecognitionProvider(mockLogger)

	tests := []struct {
		name     string
		omdbType string
		expected MediaType
	}{
		{
			name:     "movie type",
			omdbType: "movie",
			expected: MediaTypeMovie,
		},
		{
			name:     "series type",
			omdbType: "series",
			expected: MediaTypeTVSeries,
		},
		{
			name:     "episode type",
			omdbType: "episode",
			expected: MediaTypeTVEpisode,
		},
		{
			name:     "unknown type defaults to movie",
			omdbType: "unknown",
			expected: MediaTypeMovie,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := provider.mapOMDbType(tt.omdbType)
			assert.Equal(t, tt.expected, result)
		})
	}
}
