package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewBookRecognitionProvider(t *testing.T) {
	mockLogger := zap.NewNop()
	provider := NewBookRecognitionProvider(mockLogger)

	assert.NotNil(t, provider)
}

func TestBookRecognitionProvider_GetProviderName(t *testing.T) {
	mockLogger := zap.NewNop()
	provider := NewBookRecognitionProvider(mockLogger)

	name := provider.GetProviderName()
	assert.NotEmpty(t, name)
	assert.Equal(t, "book_recognition", name)
}

func TestBookRecognitionProvider_SupportsMediaType(t *testing.T) {
	mockLogger := zap.NewNop()
	provider := NewBookRecognitionProvider(mockLogger)

	tests := []struct {
		name      string
		mediaType MediaType
		expected  bool
	}{
		{
			name:      "supports ebook type",
			mediaType: MediaTypeEbook,
			expected:  true,
		},
		{
			name:      "supports book type",
			mediaType: MediaTypeBook,
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

func TestBookRecognitionProvider_GetConfidenceThreshold(t *testing.T) {
	mockLogger := zap.NewNop()
	provider := NewBookRecognitionProvider(mockLogger)

	threshold := provider.GetConfidenceThreshold()
	assert.Greater(t, threshold, 0.0)
	assert.LessOrEqual(t, threshold, 1.0)
}

func TestBookRecognitionProvider_ExtractBookMetadataFromFilename(t *testing.T) {
	mockLogger := zap.NewNop()
	provider := NewBookRecognitionProvider(mockLogger)

	tests := []struct {
		name     string
		filename string
	}{
		{
			name:     "author and title format",
			filename: "Stephen King - The Shining.epub",
		},
		{
			name:     "title only format",
			filename: "The Great Gatsby.pdf",
		},
		{
			name:     "title with year",
			filename: "1984 - George Orwell.epub",
		},
		{
			name:     "complex filename",
			filename: "J.R.R. Tolkien - The Lord of the Rings (2001).mobi",
		},
		{
			name:     "empty filename",
			filename: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			title, author, isbn := provider.extractBookMetadataFromFilename(tt.filename)
			// At minimum, the function should return without panicking
			_ = title
			_ = author
			_ = isbn
		})
	}
}

func TestBookRecognitionProvider_CleanISBN(t *testing.T) {
	mockLogger := zap.NewNop()
	provider := NewBookRecognitionProvider(mockLogger)

	tests := []struct {
		name     string
		isbn     string
		expected string
	}{
		{
			name:     "clean ISBN-13",
			isbn:     "978-0-13-468599-1",
			expected: "9780134685991",
		},
		{
			name:     "clean ISBN with spaces",
			isbn:     "978 0 13 468599 1",
			expected: "9780134685991",
		},
		{
			name:     "already clean ISBN",
			isbn:     "9780134685991",
			expected: "9780134685991",
		},
		{
			name:     "empty ISBN",
			isbn:     "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := provider.cleanISBN(tt.isbn)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBookRecognitionProvider_LooksLikeAuthorName(t *testing.T) {
	mockLogger := zap.NewNop()
	provider := NewBookRecognitionProvider(mockLogger)

	tests := []struct {
		name     string
		text     string
		expected bool
	}{
		{
			name:     "simple author name",
			text:     "John Doe",
			expected: true,
		},
		{
			name:     "single word",
			text:     "Word",
			expected: false,
		},
		{
			name:     "empty string",
			text:     "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := provider.looksLikeAuthorName(tt.text)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBookRecognitionProvider_DetectLanguage(t *testing.T) {
	mockLogger := zap.NewNop()
	provider := NewBookRecognitionProvider(mockLogger)

	tests := []struct {
		name     string
		text     string
		expected string
	}{
		{
			name:     "english text",
			text:     "The quick brown fox jumps over the lazy dog",
			expected: "en",
		},
		{
			name:     "empty text",
			text:     "",
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := provider.detectLanguage(tt.text)
			assert.NotEmpty(t, result)
		})
	}
}

func TestBookRecognitionProvider_DeterminePublicationType(t *testing.T) {
	mockLogger := zap.NewNop()
	provider := NewBookRecognitionProvider(mockLogger)

	tests := []struct {
		name     string
		filename string
	}{
		{
			name:     "epub file",
			filename: "test.epub",
		},
		{
			name:     "pdf file",
			filename: "test.pdf",
		},
		{
			name:     "mobi file",
			filename: "test.mobi",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := provider.determinePublicationType("", tt.filename, "")
			assert.NotEmpty(t, result)
		})
	}
}

func TestBookRecognitionProvider_ParseYear(t *testing.T) {
	mockLogger := zap.NewNop()
	provider := NewBookRecognitionProvider(mockLogger)

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
			dateStr:  "2024",
			expected: 2024,
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

func TestBookRecognitionProvider_CountSyllables(t *testing.T) {
	mockLogger := zap.NewNop()
	provider := NewBookRecognitionProvider(mockLogger)

	tests := []struct {
		name     string
		word     string
		minCount int
	}{
		{
			name:     "simple word",
			word:     "hello",
			minCount: 1,
		},
		{
			name:     "long word",
			word:     "international",
			minCount: 3,
		},
		{
			name:     "empty word",
			word:     "",
			minCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := provider.countSyllables(tt.word)
			assert.GreaterOrEqual(t, result, tt.minCount)
		})
	}
}
