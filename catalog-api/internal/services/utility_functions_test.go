package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// ============================================================================
// DuplicateDetectionService — Text Similarity Algorithm Tests
// ============================================================================

func newTestDuplicateDetectionService() *DuplicateDetectionService {
	return NewDuplicateDetectionService(nil, zap.NewNop(), nil)
}

func TestDuplicateDetection_NormalizeText(t *testing.T) {
	svc := newTestDuplicateDetectionService()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty", "", ""},
		{"lowercase", "Hello World", "hello world"},
		{"removes stop words", "The Lord of the Rings", "lord rings"},
		{"removes special chars", "Hello, World! 2024", "hello world 2024"},
		{"removes articles", "A Tale of Two Cities", "tale two cities"},
		{"normalizes spaces", "  hello   world  ", "hello world"},
		{"preserves numbers", "2001 Space Odyssey", "2001 space odyssey"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := svc.normalizeText(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestDuplicateDetection_LevenshteinDistance(t *testing.T) {
	svc := newTestDuplicateDetectionService()

	tests := []struct {
		name     string
		s1, s2   string
		expected int
	}{
		{"identical", "hello", "hello", 0},
		{"empty first", "", "hello", 5},
		{"empty second", "hello", "", 5},
		{"both empty", "", "", 0},
		{"one char diff", "hello", "hallo", 1},
		{"one insertion", "hello", "helloo", 1},
		{"one deletion", "hello", "helo", 1},
		{"completely different", "abc", "xyz", 3},
		{"kitten-sitting", "kitten", "sitting", 3},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := svc.levenshteinDistance(tc.s1, tc.s2)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestDuplicateDetection_JaroWinklerSimilarity(t *testing.T) {
	svc := newTestDuplicateDetectionService()

	tests := []struct {
		name string
		s1   string
		s2   string
		min  float64
		max  float64
	}{
		{"identical", "hello", "hello", 1.0, 1.0},
		{"empty first", "", "hello", 0.0, 0.0},
		{"empty second", "hello", "", 0.0, 0.0},
		{"both empty strings", "", "", 1.0, 1.0}, // jaro-winkler returns 1.0 for identical strings including both empty
		{"similar", "martha", "marhta", 0.9, 1.0},
		{"completely different", "abc", "xyz", 0.0, 0.01},
		{"common prefix", "prefix_a", "prefix_b", 0.8, 1.0},
		{"single char match", "a", "a", 1.0, 1.0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := svc.jaroWinklerSimilarity(tc.s1, tc.s2)
			assert.GreaterOrEqual(t, result, tc.min, "should be >= min")
			assert.LessOrEqual(t, result, tc.max, "should be <= max")
		})
	}
}

func TestDuplicateDetection_CosineSimilarity(t *testing.T) {
	svc := newTestDuplicateDetectionService()

	tests := []struct {
		name string
		t1   string
		t2   string
		min  float64
		max  float64
	}{
		{"identical", "hello world", "hello world", 0.99, 1.01}, // floating point tolerance
		{"empty first", "", "hello", 0.0, 0.0},
		{"empty second", "hello", "", 0.0, 0.0},
		{"partial overlap", "hello world foo", "hello world bar", 0.5, 0.9},
		{"no overlap", "abc def", "xyz uvw", 0.0, 0.0},
		{"repeated words", "hello hello", "hello", 1.0, 1.0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := svc.cosineSimilarity(tc.t1, tc.t2)
			assert.GreaterOrEqual(t, result, tc.min)
			assert.LessOrEqual(t, result, tc.max)
		})
	}
}

func TestDuplicateDetection_JaccardIndex(t *testing.T) {
	svc := newTestDuplicateDetectionService()

	tests := []struct {
		name string
		t1   string
		t2   string
		min  float64
		max  float64
	}{
		{"identical", "hello world", "hello world", 1.0, 1.0},
		{"no overlap", "abc def", "xyz uvw", 0.0, 0.0},
		{"half overlap", "a b c d", "a b e f", 0.3, 0.4},
		{"subset", "a b", "a b c d", 0.4, 0.6},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := svc.jaccardIndex(tc.t1, tc.t2)
			assert.GreaterOrEqual(t, result, tc.min)
			assert.LessOrEqual(t, result, tc.max)
		})
	}
}

func TestDuplicateDetection_LCSRatio(t *testing.T) {
	svc := newTestDuplicateDetectionService()

	tests := []struct {
		name string
		s1   string
		s2   string
		min  float64
		max  float64
	}{
		{"identical", "hello", "hello", 1.0, 1.0},
		{"both empty", "", "", 1.0, 1.0},
		{"no common", "abc", "xyz", 0.0, 0.01},
		{"subsequence", "abc", "aXbYc", 0.5, 0.7},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := svc.lcsRatio(tc.s1, tc.s2)
			assert.GreaterOrEqual(t, result, tc.min)
			assert.LessOrEqual(t, result, tc.max)
		})
	}
}

func TestDuplicateDetection_LCSLength(t *testing.T) {
	svc := newTestDuplicateDetectionService()

	assert.Equal(t, 5, svc.lcsLength("hello", "hello"))
	assert.Equal(t, 0, svc.lcsLength("abc", "xyz"))
	assert.Equal(t, 3, svc.lcsLength("abc", "aXbYc"))
	assert.Equal(t, 0, svc.lcsLength("", "abc"))
}

func TestDuplicateDetection_Soundex(t *testing.T) {
	svc := newTestDuplicateDetectionService()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty", "", "0000"},
		{"Robert", "Robert", "R163"},
		{"Rupert", "Rupert", "R163"},
		{"Ashcraft", "Ashcraft", "A226"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, svc.soundex(tc.input))
		})
	}

	assert.True(t, svc.soundexMatch("Robert", "Rupert"))
	assert.False(t, svc.soundexMatch("Robert", "John"))
}

func TestDuplicateDetection_Metaphone(t *testing.T) {
	svc := newTestDuplicateDetectionService()

	assert.True(t, svc.metaphoneMatch("hello", "hello"))
	assert.False(t, svc.metaphoneMatch("abc", "xyz"))
	assert.Equal(t, "", svc.metaphone(""))
}

func TestDuplicateDetection_CalculateTextSimilarity(t *testing.T) {
	svc := newTestDuplicateDetectionService()

	score := svc.calculateTextSimilarity("The Lord of the Rings", "The Lord of the Rings")
	assert.Equal(t, 1.0, score)

	score = svc.calculateTextSimilarity("", "hello")
	assert.Equal(t, 0.0, score)

	score = svc.calculateTextSimilarity("Star Wars Episode IV", "Star Wars Episode V")
	assert.Greater(t, score, 0.5)

	score = svc.calculateTextSimilarity("abc", "xyz")
	assert.Less(t, score, 0.5)
}

func TestDuplicateDetection_CalculateTextMetrics(t *testing.T) {
	svc := newTestDuplicateDetectionService()

	metrics := svc.calculateTextMetrics("hello", "hello")
	assert.Equal(t, 0, metrics.LevenshteinDistance)
	assert.Equal(t, 1.0, metrics.JaroWinklerScore)
	assert.Equal(t, 1.0, metrics.CosineSimilarity)
	assert.Equal(t, 1.0, metrics.JaccardIndex)
	assert.Equal(t, 1.0, metrics.LCSRatio)
	assert.True(t, metrics.SoundexMatch)
	assert.True(t, metrics.MetaphoneMatch)

	metrics = svc.calculateTextMetrics("abc", "xyz")
	assert.Equal(t, 3, metrics.LevenshteinDistance)
	assert.Less(t, metrics.JaroWinklerScore, 0.5)
}

func TestDuplicateDetection_CalculateVideoMetadataSimilarity(t *testing.T) {
	svc := newTestDuplicateDetectionService()

	tests := []struct {
		name string
		i1   *DuplicateItem
		i2   *DuplicateItem
		min  float64
		max  float64
	}{
		{
			"same director and year",
			&DuplicateItem{Director: "Christopher Nolan", Year: 2010, Duration: 148000, Quality: "1080p"},
			&DuplicateItem{Director: "Christopher Nolan", Year: 2010, Duration: 148000, Quality: "720p"},
			0.7, 1.0,
		},
		{
			"different everything",
			&DuplicateItem{Director: "Nolan", Year: 2010, Duration: 100000},
			&DuplicateItem{Director: "Spielberg", Year: 1993, Duration: 200000},
			0.0, 0.5,
		},
		{
			"empty metadata",
			&DuplicateItem{},
			&DuplicateItem{},
			0.0, 0.1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := svc.calculateVideoMetadataSimilarity(tc.i1, tc.i2)
			assert.GreaterOrEqual(t, result, tc.min)
			assert.LessOrEqual(t, result, tc.max)
		})
	}
}

func TestDuplicateDetection_CalculateAudioMetadataSimilarity(t *testing.T) {
	svc := newTestDuplicateDetectionService()

	item1 := &DuplicateItem{Artist: "Beatles", Album: "Abbey Road", Year: 1969, Bitrate: 320}
	item2 := &DuplicateItem{Artist: "Beatles", Album: "Abbey Road", Year: 1969, Bitrate: 256}
	result := svc.calculateAudioMetadataSimilarity(item1, item2)
	assert.Greater(t, result, 0.7)

	item3 := &DuplicateItem{Artist: "Madonna", Album: "Ray of Light", Year: 1998}
	result = svc.calculateAudioMetadataSimilarity(item1, item3)
	assert.Less(t, result, 0.5)
}

func TestDuplicateDetection_CalculateBookMetadataSimilarity(t *testing.T) {
	svc := newTestDuplicateDetectionService()

	item1 := &DuplicateItem{Author: "Tolkien", Year: 1954}
	item2 := &DuplicateItem{Author: "Tolkien", Year: 1954}
	result := svc.calculateBookMetadataSimilarity(item1, item2)
	assert.Greater(t, result, 0.7)

	item3 := &DuplicateItem{Author: "Rowling", Year: 1997}
	result = svc.calculateBookMetadataSimilarity(item1, item3)
	assert.Less(t, result, 0.5)
}

func TestDuplicateDetection_CalculateSoftwareMetadataSimilarity(t *testing.T) {
	svc := newTestDuplicateDetectionService()

	// Same year: year similarity contributes
	item1 := &DuplicateItem{Year: 2024}
	item2 := &DuplicateItem{Year: 2024}
	result := svc.calculateSoftwareMetadataSimilarity(item1, item2)
	assert.GreaterOrEqual(t, result, 0.0)

	// Different year: lower score
	item3 := &DuplicateItem{Year: 2020}
	result2 := svc.calculateSoftwareMetadataSimilarity(item1, item3)
	assert.GreaterOrEqual(t, result2, 0.0)
	assert.LessOrEqual(t, result2, result) // same year should score >= different year
}

func TestDuplicateDetection_CalculateGenericMetadataSimilarity(t *testing.T) {
	svc := newTestDuplicateDetectionService()

	// Same year and duration: should produce some positive score
	item1 := &DuplicateItem{Year: 2024, Duration: 120000}
	item2 := &DuplicateItem{Year: 2024, Duration: 120000}
	result := svc.calculateGenericMetadataSimilarity(item1, item2)
	assert.GreaterOrEqual(t, result, 0.0)

	// Different year: lower score
	item3 := &DuplicateItem{Year: 2010, Duration: 60000}
	result2 := svc.calculateGenericMetadataSimilarity(item1, item3)
	assert.GreaterOrEqual(t, result2, 0.0)
	assert.LessOrEqual(t, result2, result)
}

func TestDuplicateDetection_CalculateFingerprintSimilarity(t *testing.T) {
	svc := newTestDuplicateDetectionService()

	// calculateHashSimilarity is a stub that always returns 0.0,
	// so calculateFingerprintSimilarity always returns 0.0
	fp1 := map[string]string{"audio": "abc123", "video": "def456"}
	fp2 := map[string]string{"audio": "abc123", "video": "xyz789"}
	result := svc.calculateFingerprintSimilarity(fp1, fp2)
	assert.Equal(t, 0.0, result) // stub always returns 0

	// Empty fingerprints
	result = svc.calculateFingerprintSimilarity(map[string]string{}, fp1)
	assert.Equal(t, 0.0, result)

	result = svc.calculateFingerprintSimilarity(fp1, map[string]string{})
	assert.Equal(t, 0.0, result)
}

func TestDuplicateDetection_CalculateFileSimilarity(t *testing.T) {
	svc := newTestDuplicateDetectionService()

	// With matching file names, sizes, and formats: high score
	item1 := &DuplicateItem{FileName: "movie.mkv", FileSize: 1000000, Format: "mkv"}
	item2 := &DuplicateItem{FileName: "movie.mkv", FileSize: 1000000, Format: "mkv"}
	result := svc.calculateFileSimilarity(item1, item2)
	assert.Greater(t, result, 0.8)

	// Different file names and sizes: lower score
	item3 := &DuplicateItem{FileName: "other.avi", FileSize: 5000000, Format: "avi"}
	result = svc.calculateFileSimilarity(item1, item3)
	assert.Less(t, result, 0.5)

	// Empty items: FileName text similarity returns 0 for empty
	item4 := &DuplicateItem{}
	item5 := &DuplicateItem{}
	result = svc.calculateFileSimilarity(item4, item5)
	assert.GreaterOrEqual(t, result, 0.0)
}

func TestDuplicateDetection_CalculateSimilarity_HashMatch(t *testing.T) {
	svc := newTestDuplicateDetectionService()
	ctx := context.Background()

	item1 := &DuplicateItem{FileHash: "abc123", Title: "Test"}
	item2 := &DuplicateItem{FileHash: "abc123", Title: "Test"}
	analysis := svc.calculateSimilarity(ctx, item1, item2, MediaTypeMovie)
	assert.Equal(t, 1.0, analysis.OverallScore)
	assert.True(t, analysis.HashMatch)
}

func TestDuplicateDetection_CalculateSimilarity_ExternalIDMatch(t *testing.T) {
	svc := newTestDuplicateDetectionService()
	ctx := context.Background()

	item1 := &DuplicateItem{
		Title:       "Test Movie",
		ExternalIDs: map[string]string{"tmdb": "12345"},
	}
	item2 := &DuplicateItem{
		Title:       "Test Movie Remaster",
		ExternalIDs: map[string]string{"tmdb": "12345"},
	}
	analysis := svc.calculateSimilarity(ctx, item1, item2, MediaTypeMovie)
	assert.Equal(t, 0.95, analysis.OverallScore)
	assert.True(t, analysis.ExternalIDMatch)
}

func TestDuplicateDetection_CalculateSimilarity_TitleBased(t *testing.T) {
	svc := newTestDuplicateDetectionService()
	ctx := context.Background()

	item1 := &DuplicateItem{Title: "Inception", Year: 2010}
	item2 := &DuplicateItem{Title: "Inception", Year: 2010}
	analysis := svc.calculateSimilarity(ctx, item1, item2, MediaTypeMovie)
	assert.Greater(t, analysis.OverallScore, 0.5)
	assert.Greater(t, analysis.TitleSimilarity, 0.8)
}

func TestDuplicateDetection_GetSimilarityWeights(t *testing.T) {
	svc := newTestDuplicateDetectionService()

	movieWeights := svc.getSimilarityWeights(MediaTypeMovie)
	assert.Contains(t, movieWeights, "title")
	assert.Contains(t, movieWeights, "metadata")
	assert.Contains(t, movieWeights, "file")

	musicWeights := svc.getSimilarityWeights(MediaTypeMusic)
	assert.Contains(t, musicWeights, "title")
}

// ============================================================================
// SubtitleService — Additional Parsing Tests (non-colliding with existing)
// ============================================================================

func newTestSubtitleServiceUtil() *SubtitleService {
	return NewSubtitleService(nil, zap.NewNop(), nil)
}

func TestSubtitleParseSRT_Empty(t *testing.T) {
	svc := newTestSubtitleServiceUtil()
	lines, err := svc.parseSRT("")
	require.NoError(t, err)
	assert.Empty(t, lines)
}

func TestSubtitleParseSRT_MultiLineText(t *testing.T) {
	svc := newTestSubtitleServiceUtil()
	content := "1\n00:00:01,000 --> 00:00:03,000\nLine one\nLine two\n\n"
	lines, err := svc.parseSRT(content)
	require.NoError(t, err)
	require.Len(t, lines, 1)
	assert.Contains(t, lines[0].Text, "Line one")
}

func TestSubtitleParseVTT_WithBOM(t *testing.T) {
	svc := newTestSubtitleServiceUtil()
	content := "\xef\xbb\xbfWEBVTT\n\n00:00:01.000 --> 00:00:03.000\nHello\n\n"
	lines, err := svc.parseVTT(content)
	require.NoError(t, err)
	require.Len(t, lines, 1)
	assert.Equal(t, "Hello", lines[0].Text)
}

func TestSubtitleParseVTT_ShortTimestamp(t *testing.T) {
	svc := newTestSubtitleServiceUtil()
	content := "WEBVTT\n\n01:23.456 --> 02:34.567\nShort timestamp\n\n"
	lines, err := svc.parseVTT(content)
	require.NoError(t, err)
	require.Len(t, lines, 1)
	assert.Equal(t, "00:01:23,456", lines[0].StartTime)
	assert.Equal(t, "00:02:34,567", lines[0].EndTime)
}

func TestSubtitleParseVTT_NoCues(t *testing.T) {
	svc := newTestSubtitleServiceUtil()
	_, err := svc.parseVTT("WEBVTT\n\n")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no valid subtitle cues")
}

func TestSubtitleParseASS_WithBOM(t *testing.T) {
	svc := newTestSubtitleServiceUtil()
	content := "\xef\xbb\xbf[Script Info]\nTitle: Test\n\n[Events]\nFormat: Layer, Start, End, Style, Name, MarginL, MarginR, MarginV, Effect, Text\nDialogue: 0,0:00:01.00,0:00:03.00,Default,,0,0,0,,Hello BOM\n"
	lines, err := svc.parseASS(content)
	require.NoError(t, err)
	require.Len(t, lines, 1)
	assert.Equal(t, "Hello BOM", lines[0].Text)
}

func TestParseASSTimestamp_Util(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{"normal", "0:00:01.50", "00:00:01,500", false},
		{"hours", "1:23:45.67", "01:23:45,670", false},
		{"zero", "0:00:00.00", "00:00:00,000", false},
		{"invalid", "invalid", "", true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := parseASSTimestamp(tc.input)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expected, result)
			}
		})
	}
}

func TestGetSubtitleStringValue_Util(t *testing.T) {
	s := "hello"
	assert.Equal(t, "hello", getSubtitleStringValue(&s))
	assert.Equal(t, "", getSubtitleStringValue(nil))

	empty := ""
	assert.Equal(t, "", getSubtitleStringValue(&empty))
}

func TestSubtitleSortResults(t *testing.T) {
	svc := newTestSubtitleServiceUtil()

	results := []SubtitleSearchResult{
		{Language: "en", MatchScore: 0.5, Rating: 3.0},
		{Language: "en", MatchScore: 0.9, Rating: 4.0},
		{Language: "fr", MatchScore: 0.7, Rating: 5.0},
	}

	svc.sortSubtitleResults(results)

	assert.Equal(t, 0.9, results[0].MatchScore)
	assert.Equal(t, 0.7, results[1].MatchScore)
	assert.Equal(t, 0.5, results[2].MatchScore)
}

func TestSubtitleSortResults_TieBreaker(t *testing.T) {
	svc := newTestSubtitleServiceUtil()

	results := []SubtitleSearchResult{
		{MatchScore: 0.8, Rating: 3.0},
		{MatchScore: 0.8, Rating: 5.0},
	}

	svc.sortSubtitleResults(results)

	assert.Equal(t, 5.0, results[0].Rating)
	assert.Equal(t, 3.0, results[1].Rating)
}

// ============================================================================
// BookRecognitionProvider — Utility Function Tests
// ============================================================================

func newTestBookRecognitionProvider() *BookRecognitionProvider {
	return NewBookRecognitionProvider(zap.NewNop())
}

func TestBookRecognition_ExtractISBNFromText(t *testing.T) {
	provider := newTestBookRecognitionProvider()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"ISBN-13 with prefix", "ISBN: 978-0-321-12521-7", "9780321125217"},
		{"ISBN-10", "ISBN 0-321-12521-6", "0321125216"},
		{"ISBN in text", "The book ISBN: 9780321125217 was published", "9780321125217"},
		{"no ISBN", "no isbn here", ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := provider.extractISBNFromText(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestBookRecognition_CleanISBN(t *testing.T) {
	provider := newTestBookRecognitionProvider()

	assert.Equal(t, "9780321125217", provider.cleanISBN("978-0-321-12521-7"))
	assert.Equal(t, "032112521X", provider.cleanISBN("0-321-12521-X"))
	assert.Equal(t, "032112521X", provider.cleanISBN("0-321-12521-x"))
}

func TestBookRecognition_ExtractMetadataFromContent(t *testing.T) {
	provider := newTestBookRecognitionProvider()

	text := "Short\nThis is a valid title line for testing extraction\nChapter 1: Introduction\nSome content here\nChapter 2: Methods\n"
	metadata := provider.extractMetadataFromContent(text)

	assert.NotEmpty(t, metadata.Title)
	assert.Len(t, metadata.ChapterTitles, 2)
	assert.Equal(t, "Introduction", metadata.ChapterTitles[0])
	assert.Equal(t, "Methods", metadata.ChapterTitles[1])
}

func TestBookRecognition_DetectLanguage(t *testing.T) {
	provider := newTestBookRecognitionProvider()

	assert.Equal(t, "en", provider.detectLanguage("The quick brown fox jumps over the lazy dog and the cat"))
	assert.Equal(t, "es", provider.detectLanguage("El gato negro de la casa que es grande y bonita"))
	assert.Equal(t, "fr", provider.detectLanguage("Le chat noir de la maison et le chien avoir un os"))
}

func TestBookRecognition_ExtractTopics(t *testing.T) {
	provider := newTestBookRecognitionProvider()

	text := "Machine Learning is growing. Machine Learning applications are everywhere. Data Science is also important. Data Science papers are published."
	topics := provider.extractTopics(text)
	assert.NotEmpty(t, topics)
	found := false
	for _, topic := range topics {
		if topic == "Machine Learning" || topic == "Data Science" {
			found = true
		}
	}
	assert.True(t, found, "Should extract at least one capitalized topic")
}

func TestBookRecognition_ExtractKeywords(t *testing.T) {
	provider := newTestBookRecognitionProvider()

	text := "machine learning algorithms machine learning models machine learning training data processing"
	keywords := provider.extractKeywords(text)
	assert.NotEmpty(t, keywords)
	foundMachine := false
	for _, kw := range keywords {
		if kw == "machine" || kw == "learning" {
			foundMachine = true
		}
	}
	assert.True(t, foundMachine)
}

func TestBookRecognition_CalculateReadabilityScore(t *testing.T) {
	provider := newTestBookRecognitionProvider()

	assert.Equal(t, 0.0, provider.calculateReadabilityScore(""))
	assert.Equal(t, 0.0, provider.calculateReadabilityScore("hello world"))

	simpleText := "The cat sat. The dog ran. The bird flew. It was nice."
	score := provider.calculateReadabilityScore(simpleText)
	assert.Greater(t, score, 0.0)
	assert.LessOrEqual(t, score, 1.0)
}

func TestBookRecognition_CountSyllables(t *testing.T) {
	provider := newTestBookRecognitionProvider()

	assert.Greater(t, provider.countSyllables("hello world"), 0)
	assert.Greater(t, provider.countSyllables("internationalization"), 0)
	assert.Equal(t, 0, provider.countSyllables(""))
}

func TestBookRecognition_DetermineBookType(t *testing.T) {
	provider := newTestBookRecognitionProvider()

	tests := []struct {
		name     string
		info     GoogleBookVolumeInfo
		expected MediaType
	}{
		{"magazine", GoogleBookVolumeInfo{PrintType: "MAGAZINE"}, MediaTypeMagazine},
		{"comic category", GoogleBookVolumeInfo{Categories: []string{"Comics & Graphic Novels"}}, MediaTypeComicBook},
		{"magazine category", GoogleBookVolumeInfo{Categories: []string{"Periodical Literature"}}, MediaTypeMagazine},
		{"reference category", GoogleBookVolumeInfo{Categories: []string{"Reference Manual"}}, MediaTypeManual},
		{"default book", GoogleBookVolumeInfo{PrintType: "BOOK"}, MediaTypeBook},
		{"no info", GoogleBookVolumeInfo{}, MediaTypeBook},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, provider.determineBookType(tc.info))
		})
	}
}

func TestBookRecognition_MapCrossrefType(t *testing.T) {
	provider := newTestBookRecognitionProvider()

	tests := []struct {
		name     string
		input    string
		expected MediaType
	}{
		{"journal-article", "journal-article", MediaTypeJournal},
		{"book-chapter", "book-chapter", MediaTypeBook},
		{"book", "book", MediaTypeBook},
		{"proceedings", "proceedings-article", MediaTypeJournal},
		{"reference", "reference-entry", MediaTypeManual},
		{"unknown", "unknown-type", MediaTypeBook},
		{"case insensitive", "JOURNAL-ARTICLE", MediaTypeJournal},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, provider.mapCrossrefType(tc.input))
		})
	}
}

func TestBookRecognition_GetFileExtension(t *testing.T) {
	provider := newTestBookRecognitionProvider()

	assert.Equal(t, "pdf", provider.getFileExtension("book.pdf"))
	assert.Equal(t, "epub", provider.getFileExtension("novel.epub"))
	assert.Equal(t, "gz", provider.getFileExtension("archive.tar.gz"))
	assert.Equal(t, "", provider.getFileExtension("noextension"))
}

func TestBookRecognition_ParseDate(t *testing.T) {
	provider := newTestBookRecognitionProvider()

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"ISO date", "2024-01-15", false},
		{"year-month", "2024-01", false},
		{"year only", "2024", false},
		{"long format", "January 15, 2024", false},
		{"short month format", "Jan 15, 2024", false},
		{"ISO datetime", "2024-01-15T10:30:00Z", false},
		{"invalid", "not-a-date", true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			date, err := provider.parseDate(tc.input)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.False(t, date.IsZero())
			}
		})
	}
}

func TestBookRecognition_CalculateGoogleBooksConfidence(t *testing.T) {
	provider := newTestBookRecognitionProvider()

	assert.Equal(t, 0.8, provider.calculateGoogleBooksConfidence(4.5, 200))
	assert.Equal(t, 0.7, provider.calculateGoogleBooksConfidence(3.8, 75))
	assert.Equal(t, 0.5, provider.calculateGoogleBooksConfidence(2.0, 10))
}

func TestBookRecognition_CalculateOpenLibraryConfidence(t *testing.T) {
	provider := newTestBookRecognitionProvider()

	assert.InDelta(t, 0.8, provider.calculateOpenLibraryConfidence(10, 5), 0.001)
	assert.InDelta(t, 0.7, provider.calculateOpenLibraryConfidence(10, 0), 0.001)
	assert.InDelta(t, 0.6, provider.calculateOpenLibraryConfidence(3, 1), 0.001)
	assert.InDelta(t, 0.5, provider.calculateOpenLibraryConfidence(2, 0), 0.001)
}

func TestBookRecognition_GenerateID(t *testing.T) {
	provider := newTestBookRecognitionProvider()

	id1 := provider.generateID("test-input")
	id2 := provider.generateID("test-input")
	id3 := provider.generateID("different-input")

	assert.Equal(t, id1, id2, "Same input should produce same ID")
	assert.NotEqual(t, id1, id3, "Different input should produce different ID")
	assert.Len(t, id1, 12)
}

// ============================================================================
// Standalone Utility Functions
// ============================================================================

func TestIsVowel(t *testing.T) {
	// isVowel only checks uppercase letters
	assert.True(t, isVowel('A'))
	assert.True(t, isVowel('E'))
	assert.True(t, isVowel('I'))
	assert.True(t, isVowel('O'))
	assert.True(t, isVowel('U'))
	assert.False(t, isVowel('a')) // lowercase not handled
	assert.False(t, isVowel('B'))
	assert.False(t, isVowel('X'))
	assert.False(t, isVowel('Z'))
}

func TestMin3(t *testing.T) {
	assert.Equal(t, 1, min3(1, 2, 3))
	assert.Equal(t, 1, min3(3, 1, 2))
	assert.Equal(t, 1, min3(2, 3, 1))
	assert.Equal(t, 0, min3(0, 0, 0))
	assert.Equal(t, -1, min3(-1, 0, 1))
}

// ============================================================================
// MusicPlayerService — Queue Management Tests (Pure Functions)
// ============================================================================

func newTestMusicPlayerServiceUtil() *MusicPlayerService {
	return &MusicPlayerService{logger: zap.NewNop()}
}

func TestMusicPlayer_ShuffleQueue(t *testing.T) {
	svc := newTestMusicPlayerServiceUtil()

	session := &MusicPlaybackSession{
		Queue: []MusicTrack{
			{ID: 1}, {ID: 2}, {ID: 3}, {ID: 4}, {ID: 5},
		},
		QueueIndex: 2,
	}

	svc.shuffleQueue(session)

	assert.Equal(t, 0, session.QueueIndex)
	assert.Equal(t, int64(3), session.Queue[0].ID)
	assert.Len(t, session.Queue, 5)
}

func TestMusicPlayer_ShuffleQueue_Empty(t *testing.T) {
	svc := newTestMusicPlayerServiceUtil()
	session := &MusicPlaybackSession{Queue: []MusicTrack{}}
	svc.shuffleQueue(session)
	assert.Empty(t, session.Queue)
}

func TestMusicPlayer_ShuffleQueue_SingleTrack(t *testing.T) {
	svc := newTestMusicPlayerServiceUtil()
	session := &MusicPlaybackSession{
		Queue:      []MusicTrack{{ID: 1}},
		QueueIndex: 0,
	}
	svc.shuffleQueue(session)
	assert.Len(t, session.Queue, 1)
	assert.Equal(t, int64(1), session.Queue[0].ID)
}

func TestMusicPlayer_UnshuffleQueue(t *testing.T) {
	svc := newTestMusicPlayerServiceUtil()
	currentTrack := MusicTrack{ID: 3}
	session := &MusicPlaybackSession{
		Queue: []MusicTrack{
			{ID: 3}, {ID: 1}, {ID: 5}, {ID: 2}, {ID: 4},
		},
		QueueIndex:     0,
		CurrentTrack:   &currentTrack,
		ShuffleHistory: []int{2, 0, 4, 1, 3},
	}

	svc.unshuffleQueue(session)

	for i := 0; i < len(session.Queue)-1; i++ {
		assert.Less(t, session.Queue[i].ID, session.Queue[i+1].ID)
	}
	assert.Equal(t, int64(3), session.Queue[session.QueueIndex].ID)
	assert.Empty(t, session.ShuffleHistory)
}

func TestMusicPlayer_UnshuffleQueue_EmptyHistory(t *testing.T) {
	svc := newTestMusicPlayerServiceUtil()
	session := &MusicPlaybackSession{
		Queue:          []MusicTrack{{ID: 3}, {ID: 1}, {ID: 2}},
		QueueIndex:     0,
		ShuffleHistory: []int{},
	}
	originalQueue := make([]MusicTrack, len(session.Queue))
	copy(originalQueue, session.Queue)

	svc.unshuffleQueue(session)

	assert.Equal(t, originalQueue, session.Queue)
}

func TestMusicPlayer_GetNextTrackIndex_RepeatAlbum(t *testing.T) {
	svc := newTestMusicPlayerServiceUtil()

	session := &MusicPlaybackSession{
		Queue:      []MusicTrack{{ID: 1}, {ID: 2}, {ID: 3}},
		QueueIndex: 2,
		RepeatMode: RepeatModeAlbum,
	}
	assert.Equal(t, 0, svc.getNextTrackIndex(session))

	session.QueueIndex = 1
	assert.Equal(t, 2, svc.getNextTrackIndex(session))
}

func TestMusicPlayer_GetPreviousTrackIndex_RepeatAlbum(t *testing.T) {
	svc := newTestMusicPlayerServiceUtil()

	session := &MusicPlaybackSession{
		Queue:      []MusicTrack{{ID: 1}, {ID: 2}, {ID: 3}},
		QueueIndex: 0,
		RepeatMode: RepeatModeAlbum,
	}
	assert.Equal(t, 2, svc.getPreviousTrackIndex(session))
}
