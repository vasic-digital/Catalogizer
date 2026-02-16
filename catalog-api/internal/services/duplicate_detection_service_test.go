package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewDuplicateDetectionService(t *testing.T) {
	mockLogger := zap.NewNop()
	service := NewDuplicateDetectionService(nil, mockLogger, nil)

	assert.NotNil(t, service)
}

func TestNewDuplicateDetectionService_NilLogger(t *testing.T) {
	service := NewDuplicateDetectionService(nil, nil, nil)

	assert.NotNil(t, service)
}

func TestDuplicateDetectionService_NormalizeText(t *testing.T) {
	mockLogger := zap.NewNop()
	service := NewDuplicateDetectionService(nil, mockLogger, nil)

	tests := []struct {
		name     string
		text     string
		expected string
	}{
		{
			name:     "lowercase conversion",
			text:     "Hello World",
			expected: "hello world",
		},
		{
			name:     "removes stop words",
			text:     "The Lord of the Rings",
			expected: "lord rings",
		},
		{
			name:     "removes special characters",
			text:     "Hello! World? Test.",
			expected: "hello world test",
		},
		{
			name:     "normalizes whitespace",
			text:     "Hello   World",
			expected: "hello world",
		},
		{
			name:     "empty string",
			text:     "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.normalizeText(tt.text)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDuplicateDetectionService_CalculateTextSimilarity(t *testing.T) {
	mockLogger := zap.NewNop()
	service := NewDuplicateDetectionService(nil, mockLogger, nil)

	tests := []struct {
		name        string
		text1       string
		text2       string
		minScore    float64
		maxScore    float64
	}{
		{
			name:     "identical texts",
			text1:    "The Lord of the Rings",
			text2:    "The Lord of the Rings",
			minScore: 1.0,
			maxScore: 1.0,
		},
		{
			name:     "similar texts",
			text1:    "The Lord of the Rings",
			text2:    "Lord of the Rings",
			minScore: 0.5,
			maxScore: 1.0,
		},
		{
			name:     "completely different",
			text1:    "Hello World",
			text2:    "Quantum Physics",
			minScore: 0.0,
			maxScore: 0.5,
		},
		{
			name:     "empty first text",
			text1:    "",
			text2:    "Some text",
			minScore: 0.0,
			maxScore: 0.0,
		},
		{
			name:     "empty second text",
			text1:    "Some text",
			text2:    "",
			minScore: 0.0,
			maxScore: 0.0,
		},
		{
			name:     "both empty",
			text1:    "",
			text2:    "",
			minScore: 0.0,
			maxScore: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := service.calculateTextSimilarity(tt.text1, tt.text2)
			assert.GreaterOrEqual(t, score, tt.minScore)
			assert.LessOrEqual(t, score, tt.maxScore)
		})
	}
}

func TestDuplicateDetectionService_LevenshteinDistance(t *testing.T) {
	mockLogger := zap.NewNop()
	service := NewDuplicateDetectionService(nil, mockLogger, nil)

	tests := []struct {
		name     string
		s1       string
		s2       string
		expected int
	}{
		{
			name:     "identical strings",
			s1:       "hello",
			s2:       "hello",
			expected: 0,
		},
		{
			name:     "one char difference",
			s1:       "hello",
			s2:       "hallo",
			expected: 1,
		},
		{
			name:     "completely different",
			s1:       "abc",
			s2:       "xyz",
			expected: 3,
		},
		{
			name:     "empty first string",
			s1:       "",
			s2:       "hello",
			expected: 5,
		},
		{
			name:     "empty second string",
			s1:       "hello",
			s2:       "",
			expected: 5,
		},
		{
			name:     "both empty",
			s1:       "",
			s2:       "",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.levenshteinDistance(tt.s1, tt.s2)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDuplicateDetectionService_JaroWinklerSimilarity(t *testing.T) {
	mockLogger := zap.NewNop()
	service := NewDuplicateDetectionService(nil, mockLogger, nil)

	tests := []struct {
		name     string
		s1       string
		s2       string
		minScore float64
		maxScore float64
	}{
		{
			name:     "identical strings",
			s1:       "hello",
			s2:       "hello",
			minScore: 1.0,
			maxScore: 1.0,
		},
		{
			name:     "similar strings",
			s1:       "hello",
			s2:       "hallo",
			minScore: 0.7,
			maxScore: 1.0,
		},
		{
			name:     "empty first string",
			s1:       "",
			s2:       "hello",
			minScore: 0.0,
			maxScore: 0.0,
		},
		{
			name:     "both empty",
			s1:       "",
			s2:       "",
			minScore: 1.0,
			maxScore: 1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := service.jaroWinklerSimilarity(tt.s1, tt.s2)
			assert.GreaterOrEqual(t, score, tt.minScore)
			assert.LessOrEqual(t, score, tt.maxScore)
		})
	}
}

func TestDuplicateDetectionService_CosineSimilarity(t *testing.T) {
	mockLogger := zap.NewNop()
	service := NewDuplicateDetectionService(nil, mockLogger, nil)

	tests := []struct {
		name     string
		text1    string
		text2    string
		minScore float64
		maxScore float64
	}{
		{
			name:     "identical texts",
			text1:    "hello world",
			text2:    "hello world",
			minScore: 0.99, // floating-point precision: cosine similarity may return ~0.9999999999999998
			maxScore: 1.001,
		},
		{
			name:     "overlapping words",
			text1:    "hello world",
			text2:    "hello there",
			minScore: 0.3,
			maxScore: 0.8,
		},
		{
			name:     "no overlap",
			text1:    "abc def",
			text2:    "xyz ghi",
			minScore: 0.0,
			maxScore: 0.0,
		},
		{
			name:     "empty text",
			text1:    "",
			text2:    "hello",
			minScore: 0.0,
			maxScore: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := service.cosineSimilarity(tt.text1, tt.text2)
			assert.GreaterOrEqual(t, score, tt.minScore)
			assert.LessOrEqual(t, score, tt.maxScore)
		})
	}
}

func TestDuplicateDetectionService_JaccardIndex(t *testing.T) {
	mockLogger := zap.NewNop()
	service := NewDuplicateDetectionService(nil, mockLogger, nil)

	tests := []struct {
		name     string
		text1    string
		text2    string
		minScore float64
		maxScore float64
	}{
		{
			name:     "identical texts",
			text1:    "hello world",
			text2:    "hello world",
			minScore: 1.0,
			maxScore: 1.0,
		},
		{
			name:     "partial overlap",
			text1:    "hello world",
			text2:    "hello there",
			minScore: 0.2,
			maxScore: 0.5,
		},
		{
			name:     "no overlap",
			text1:    "abc",
			text2:    "xyz",
			minScore: 0.0,
			maxScore: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := service.jaccardIndex(tt.text1, tt.text2)
			assert.GreaterOrEqual(t, score, tt.minScore)
			assert.LessOrEqual(t, score, tt.maxScore)
		})
	}
}

func TestDuplicateDetectionService_LCSRatio(t *testing.T) {
	mockLogger := zap.NewNop()
	service := NewDuplicateDetectionService(nil, mockLogger, nil)

	tests := []struct {
		name     string
		s1       string
		s2       string
		minScore float64
	}{
		{
			name:     "identical strings",
			s1:       "hello",
			s2:       "hello",
			minScore: 1.0,
		},
		{
			name:     "similar strings",
			s1:       "hello",
			s2:       "hallo",
			minScore: 0.5,
		},
		{
			name:     "both empty",
			s1:       "",
			s2:       "",
			minScore: 1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := service.lcsRatio(tt.s1, tt.s2)
			assert.GreaterOrEqual(t, score, tt.minScore)
			assert.LessOrEqual(t, score, 1.0)
		})
	}
}

func TestDuplicateDetectionService_Soundex(t *testing.T) {
	mockLogger := zap.NewNop()
	service := NewDuplicateDetectionService(nil, mockLogger, nil)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "0000",
		},
		{
			name:     "Robert",
			input:    "Robert",
			expected: "R163",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.soundex(tt.input)
			assert.Equal(t, tt.expected, result)
			assert.Len(t, result, 4)
		})
	}
}

func TestDuplicateDetectionService_SoundexMatch(t *testing.T) {
	mockLogger := zap.NewNop()
	service := NewDuplicateDetectionService(nil, mockLogger, nil)

	// Similar sounding words should match
	assert.True(t, service.soundexMatch("Robert", "Rupert"))
	// Different sounding words should not match
	assert.False(t, service.soundexMatch("Robert", "Smith"))
}

func TestDuplicateDetectionService_Metaphone(t *testing.T) {
	mockLogger := zap.NewNop()
	service := NewDuplicateDetectionService(nil, mockLogger, nil)

	result := service.metaphone("Hello")
	assert.NotEmpty(t, result)

	empty := service.metaphone("")
	assert.Empty(t, empty)
}

func TestDuplicateDetectionService_CalculateSimilarity_HashMatch(t *testing.T) {
	mockLogger := zap.NewNop()
	service := NewDuplicateDetectionService(nil, mockLogger, nil)

	item1 := &DuplicateItem{
		FileHash: "abc123",
		Title:    "Item A",
	}
	item2 := &DuplicateItem{
		FileHash: "abc123",
		Title:    "Item B",
	}

	analysis := service.calculateSimilarity(nil, item1, item2, MediaTypeMovie)
	assert.Equal(t, 1.0, analysis.OverallScore)
	assert.True(t, analysis.HashMatch)
	assert.Contains(t, analysis.MatchingFields, "file_hash")
}

func TestDuplicateDetectionService_CalculateSimilarity_ExternalIDMatch(t *testing.T) {
	mockLogger := zap.NewNop()
	service := NewDuplicateDetectionService(nil, mockLogger, nil)

	item1 := &DuplicateItem{
		Title:       "Item A",
		ExternalIDs: map[string]string{"tmdb": "12345"},
	}
	item2 := &DuplicateItem{
		Title:       "Item B",
		ExternalIDs: map[string]string{"tmdb": "12345"},
	}

	analysis := service.calculateSimilarity(nil, item1, item2, MediaTypeMovie)
	assert.Equal(t, 0.95, analysis.OverallScore)
	assert.True(t, analysis.ExternalIDMatch)
}

func TestDuplicateDetectionService_CalculateVideoMetadataSimilarity(t *testing.T) {
	mockLogger := zap.NewNop()
	service := NewDuplicateDetectionService(nil, mockLogger, nil)

	tests := []struct {
		name     string
		item1    *DuplicateItem
		item2    *DuplicateItem
		minScore float64
	}{
		{
			name: "same director and year",
			item1: &DuplicateItem{
				Director: "Christopher Nolan",
				Year:     2014,
			},
			item2: &DuplicateItem{
				Director: "Christopher Nolan",
				Year:     2014,
			},
			minScore: 0.9,
		},
		{
			name: "different director and year",
			item1: &DuplicateItem{
				Director: "Christopher Nolan",
				Year:     2014,
			},
			item2: &DuplicateItem{
				Director: "Steven Spielberg",
				Year:     1990,
			},
			minScore: 0.0,
		},
		{
			name:     "empty metadata",
			item1:    &DuplicateItem{},
			item2:    &DuplicateItem{},
			minScore: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := service.calculateVideoMetadataSimilarity(tt.item1, tt.item2)
			assert.GreaterOrEqual(t, score, tt.minScore)
			assert.LessOrEqual(t, score, 1.0)
		})
	}
}

func TestDuplicateDetectionService_CalculateAudioMetadataSimilarity(t *testing.T) {
	mockLogger := zap.NewNop()
	service := NewDuplicateDetectionService(nil, mockLogger, nil)

	tests := []struct {
		name     string
		item1    *DuplicateItem
		item2    *DuplicateItem
		minScore float64
	}{
		{
			name: "same artist and album",
			item1: &DuplicateItem{
				Artist: "The Beatles",
				Album:  "Abbey Road",
				Year:   1969,
			},
			item2: &DuplicateItem{
				Artist: "The Beatles",
				Album:  "Abbey Road",
				Year:   1969,
			},
			minScore: 0.9,
		},
		{
			name:     "empty metadata",
			item1:    &DuplicateItem{},
			item2:    &DuplicateItem{},
			minScore: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := service.calculateAudioMetadataSimilarity(tt.item1, tt.item2)
			assert.GreaterOrEqual(t, score, tt.minScore)
			assert.LessOrEqual(t, score, 1.0)
		})
	}
}

func TestDuplicateDetectionService_CalculateFileSimilarity(t *testing.T) {
	mockLogger := zap.NewNop()
	service := NewDuplicateDetectionService(nil, mockLogger, nil)

	tests := []struct {
		name     string
		item1    *DuplicateItem
		item2    *DuplicateItem
		minScore float64
	}{
		{
			name: "identical files",
			item1: &DuplicateItem{
				FileName: "movie.mkv",
				FileSize: 1000000,
				Format:   "mkv",
			},
			item2: &DuplicateItem{
				FileName: "movie.mkv",
				FileSize: 1000000,
				Format:   "mkv",
			},
			minScore: 0.9,
		},
		{
			name: "similar files",
			item1: &DuplicateItem{
				FileName: "movie.mkv",
				FileSize: 1000000,
				Format:   "mkv",
			},
			item2: &DuplicateItem{
				FileName: "movie.avi",
				FileSize: 1050000,
				Format:   "avi",
			},
			minScore: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := service.calculateFileSimilarity(tt.item1, tt.item2)
			assert.GreaterOrEqual(t, score, tt.minScore)
			assert.LessOrEqual(t, score, 1.0)
		})
	}
}

func TestDuplicateDetectionService_CalculateFingerprintSimilarity(t *testing.T) {
	mockLogger := zap.NewNop()
	service := NewDuplicateDetectionService(nil, mockLogger, nil)

	// Empty fingerprints
	score := service.calculateFingerprintSimilarity(map[string]string{}, map[string]string{})
	assert.Equal(t, 0.0, score)

	// Non-empty fingerprints (actual similarity depends on hash comparison)
	fp1 := map[string]string{"audio": "abc123"}
	fp2 := map[string]string{"audio": "abc123"}
	score = service.calculateFingerprintSimilarity(fp1, fp2)
	assert.GreaterOrEqual(t, score, 0.0)
	assert.LessOrEqual(t, score, 1.0)
}

func TestDuplicateDetectionService_GetSimilarityWeights(t *testing.T) {
	mockLogger := zap.NewNop()
	service := NewDuplicateDetectionService(nil, mockLogger, nil)

	tests := []struct {
		name      string
		mediaType MediaType
	}{
		{name: "music weights", mediaType: MediaTypeMusic},
		{name: "movie weights", mediaType: MediaTypeMovie},
		{name: "default weights", mediaType: MediaTypeBook},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			weights := service.getSimilarityWeights(tt.mediaType)
			assert.NotEmpty(t, weights)
			assert.Contains(t, weights, "title")
			assert.Contains(t, weights, "metadata")
			assert.Contains(t, weights, "fingerprint")
			assert.Contains(t, weights, "file")
		})
	}
}

func TestDuplicateDetectionService_CalculateTextMetrics(t *testing.T) {
	mockLogger := zap.NewNop()
	service := NewDuplicateDetectionService(nil, mockLogger, nil)

	metrics := service.calculateTextMetrics("hello world", "hello world")
	assert.NotNil(t, metrics)
	assert.Equal(t, 0, metrics.LevenshteinDistance)
	assert.Equal(t, 1.0, metrics.JaroWinklerScore)
	assert.InDelta(t, 1.0, metrics.CosineSimilarity, 0.001) // floating-point precision
	assert.Equal(t, 1.0, metrics.JaccardIndex)
}

func TestDuplicateDetectionService_LCSLength(t *testing.T) {
	mockLogger := zap.NewNop()
	service := NewDuplicateDetectionService(nil, mockLogger, nil)

	tests := []struct {
		name     string
		s1       string
		s2       string
		expected int
	}{
		{
			name:     "identical",
			s1:       "hello",
			s2:       "hello",
			expected: 5,
		},
		{
			name:     "common subsequence",
			s1:       "abcde",
			s2:       "ace",
			expected: 3,
		},
		{
			name:     "no common subsequence",
			s1:       "abc",
			s2:       "xyz",
			expected: 0,
		},
		{
			name:     "empty string",
			s1:       "",
			s2:       "hello",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.lcsLength(tt.s1, tt.s2)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestUtilityFunctions(t *testing.T) {
	// Test min3
	assert.Equal(t, 1, min3(1, 2, 3))
	assert.Equal(t, 1, min3(3, 1, 2))
	assert.Equal(t, 1, min3(2, 3, 1))

	// Test isVowel
	assert.True(t, isVowel('A'))
	assert.True(t, isVowel('E'))
	assert.True(t, isVowel('I'))
	assert.True(t, isVowel('O'))
	assert.True(t, isVowel('U'))
	assert.False(t, isVowel('B'))
	assert.False(t, isVowel('Z'))
}
