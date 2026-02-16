package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewGameSoftwareRecognitionProvider(t *testing.T) {
	mockLogger := zap.NewNop()
	provider := NewGameSoftwareRecognitionProvider(mockLogger)

	assert.NotNil(t, provider)
}

func TestGameSoftwareRecognitionProvider_GetProviderName(t *testing.T) {
	mockLogger := zap.NewNop()
	provider := NewGameSoftwareRecognitionProvider(mockLogger)

	name := provider.GetProviderName()
	assert.NotEmpty(t, name)
}

func TestGameSoftwareRecognitionProvider_SupportsMediaType(t *testing.T) {
	mockLogger := zap.NewNop()
	provider := NewGameSoftwareRecognitionProvider(mockLogger)

	tests := []struct {
		name      string
		mediaType MediaType
		expected  bool
	}{
		{
			name:      "supports game type",
			mediaType: MediaTypeGame,
			expected:  true,
		},
		{
			name:      "supports software type",
			mediaType: MediaTypeSoftware,
			expected:  true,
		},
		{
			name:      "does not support video",
			mediaType: MediaTypeVideo,
			expected:  false,
		},
		{
			name:      "does not support music",
			mediaType: MediaTypeMusic,
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

func TestGameSoftwareRecognitionProvider_GetConfidenceThreshold(t *testing.T) {
	mockLogger := zap.NewNop()
	provider := NewGameSoftwareRecognitionProvider(mockLogger)

	threshold := provider.GetConfidenceThreshold()
	assert.Greater(t, threshold, 0.0)
	assert.LessOrEqual(t, threshold, 1.0)
}

func TestGameSoftwareRecognitionProvider_ExtractSoftwareMetadataFromFilename(t *testing.T) {
	mockLogger := zap.NewNop()
	provider := NewGameSoftwareRecognitionProvider(mockLogger)

	tests := []struct {
		name     string
		filename string
	}{
		{
			name:     "game installer",
			filename: "The.Witcher.3.Wild.Hunt.v1.32.GOG.exe",
		},
		{
			name:     "software installer",
			filename: "Visual.Studio.Code.1.85.2.Setup.exe",
		},
		{
			name:     "linux package",
			filename: "gimp-2.10.36-x86_64.AppImage",
		},
		{
			name:     "mac application",
			filename: "Photoshop.2024.dmg",
		},
		{
			name:     "empty filename",
			filename: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			softwareName, version, platform := provider.extractSoftwareMetadataFromFilename(tt.filename)
			_ = softwareName
			_ = version
			_ = platform
		})
	}
}

func TestGameSoftwareRecognitionProvider_LooksLikeGame(t *testing.T) {
	mockLogger := zap.NewNop()
	provider := NewGameSoftwareRecognitionProvider(mockLogger)

	tests := []struct {
		name        string
		softName    string
		filename    string
		expected    bool
	}{
		{
			name:     "game with common keywords",
			softName: "The Witcher 3 Wild Hunt",
			filename: "The.Witcher.3.Wild.Hunt.GOG.exe",
			expected: true,
		},
		{
			name:     "software application",
			softName: "Visual Studio Code",
			filename: "Visual.Studio.Code.Setup.exe",
			expected: false,
		},
		{
			name:     "empty strings",
			softName: "",
			filename: "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := provider.looksLikeGame(tt.softName, tt.filename)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGameSoftwareRecognitionProvider_GetFileExtension(t *testing.T) {
	mockLogger := zap.NewNop()
	provider := NewGameSoftwareRecognitionProvider(mockLogger)

	tests := []struct {
		name     string
		filename string
		expected string
	}{
		{
			name:     "exe extension",
			filename: "setup.exe",
			expected: "exe",
		},
		{
			name:     "dmg extension",
			filename: "app.dmg",
			expected: "dmg",
		},
		{
			name:     "no extension",
			filename: "filename",
			expected: "",
		},
		{
			name:     "multiple dots",
			filename: "my.app.v2.tar.gz",
			expected: "gz",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := provider.getFileExtension(tt.filename)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGameSoftwareRecognitionProvider_CalculateIGDBConfidence(t *testing.T) {
	mockLogger := zap.NewNop()
	provider := NewGameSoftwareRecognitionProvider(mockLogger)

	tests := []struct {
		name          string
		rating        float64
		ratingCount   int
		popularity    float64
		minConfidence float64
	}{
		{
			name:          "high rated popular game",
			rating:        90.0,
			ratingCount:   1000,
			popularity:    80.0,
			minConfidence: 0.5,
		},
		{
			name:          "low rated game",
			rating:        30.0,
			ratingCount:   10,
			popularity:    10.0,
			minConfidence: 0.3,
		},
		{
			name:          "zero values",
			rating:        0.0,
			ratingCount:   0,
			popularity:    0.0,
			minConfidence: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			confidence := provider.calculateIGDBConfidence(tt.rating, tt.ratingCount, tt.popularity)
			assert.GreaterOrEqual(t, confidence, tt.minConfidence)
			assert.LessOrEqual(t, confidence, 1.0)
		})
	}
}

func TestGameSoftwareRecognitionProvider_GenerateID(t *testing.T) {
	mockLogger := zap.NewNop()
	provider := NewGameSoftwareRecognitionProvider(mockLogger)

	id1 := provider.generateID("Game A")
	id2 := provider.generateID("Game B")
	id3 := provider.generateID("Game A")

	assert.NotEmpty(t, id1)
	assert.NotEmpty(t, id2)
	assert.NotEqual(t, id1, id2)
	assert.Equal(t, id1, id3)
}
