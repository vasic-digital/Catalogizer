package detector

import (
	"catalogizer/internal/media/models"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func newTestEngine(t *testing.T) *DetectionEngine {
	t.Helper()
	logger, _ := zap.NewDevelopment()
	return NewDetectionEngine(logger)
}

func TestNewDetectionEngine(t *testing.T) {
	engine := newTestEngine(t)
	require.NotNil(t, engine)
	assert.NotNil(t, engine.logger)
	assert.Empty(t, engine.rules)
	assert.Empty(t, engine.mediaTypes)
}

func TestLoadRules(t *testing.T) {
	engine := newTestEngine(t)

	rules := []models.DetectionRule{
		{ID: 1, MediaTypeID: 1, RuleName: "low_priority", RuleType: "filename_pattern", Pattern: "*.mp4", ConfidenceWeight: 1.0, Enabled: true, Priority: 1},
		{ID: 2, MediaTypeID: 1, RuleName: "high_priority", RuleType: "filename_pattern", Pattern: "*.mkv", ConfidenceWeight: 1.0, Enabled: true, Priority: 10},
	}
	mediaTypes := []models.MediaType{
		{ID: 1, Name: "movie", Description: "Movies"},
	}

	engine.LoadRules(rules, mediaTypes)

	assert.Len(t, engine.rules, 2)
	assert.Len(t, engine.mediaTypes, 1)
	// Rules should be sorted by priority descending
	assert.Equal(t, 10, engine.rules[0].Priority)
	assert.Equal(t, 1, engine.rules[1].Priority)
}

func TestAnalyzeDirectory_FilenamePattern(t *testing.T) {
	tests := []struct {
		name           string
		dirPath        string
		files          []FileInfo
		rules          []models.DetectionRule
		mediaTypes     []models.MediaType
		expectNil      bool
		expectType     string
		minConfidence  float64
	}{
		{
			name:    "detects movie by video file extensions",
			dirPath: "/media/movies/Inception (2010)",
			files: []FileInfo{
				{Name: "inception.mkv", Path: "/media/movies/Inception (2010)/inception.mkv", Size: 5 * 1024 * 1024 * 1024, Extension: ".mkv"},
				{Name: "inception.srt", Path: "/media/movies/Inception (2010)/inception.srt", Size: 50000, Extension: ".srt"},
			},
			rules: []models.DetectionRule{
				{ID: 1, MediaTypeID: 1, RuleType: "filename_pattern", Pattern: `["*.mkv", "*.mp4", "*.avi"]`, ConfidenceWeight: 1.0, Enabled: true, Priority: 10},
			},
			mediaTypes:    []models.MediaType{{ID: 1, Name: "movie", Description: "Movies"}},
			expectNil:     false,
			expectType:    "movie",
			minConfidence: 0.1,
		},
		{
			name:    "no match when no files match pattern",
			dirPath: "/media/docs",
			files: []FileInfo{
				{Name: "readme.txt", Path: "/media/docs/readme.txt", Size: 1000, Extension: ".txt"},
			},
			rules: []models.DetectionRule{
				{ID: 1, MediaTypeID: 1, RuleType: "filename_pattern", Pattern: `["*.mkv", "*.mp4"]`, ConfidenceWeight: 1.0, Enabled: true, Priority: 10},
			},
			mediaTypes: []models.MediaType{{ID: 1, Name: "movie", Description: "Movies"}},
			expectNil:  true,
		},
		{
			name:    "no files yields nil result",
			dirPath: "/media/empty",
			files:   []FileInfo{},
			rules: []models.DetectionRule{
				{ID: 1, MediaTypeID: 1, RuleType: "filename_pattern", Pattern: `["*.mkv"]`, ConfidenceWeight: 1.0, Enabled: true, Priority: 10},
			},
			mediaTypes: []models.MediaType{{ID: 1, Name: "movie", Description: "Movies"}},
			expectNil:  true,
		},
		{
			name:    "disabled rules are skipped",
			dirPath: "/media/movies/Test",
			files: []FileInfo{
				{Name: "test.mkv", Path: "/media/movies/Test/test.mkv", Size: 1000, Extension: ".mkv"},
			},
			rules: []models.DetectionRule{
				{ID: 1, MediaTypeID: 1, RuleType: "filename_pattern", Pattern: `["*.mkv"]`, ConfidenceWeight: 1.0, Enabled: false, Priority: 10},
			},
			mediaTypes: []models.MediaType{{ID: 1, Name: "movie", Description: "Movies"}},
			expectNil:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			engine := newTestEngine(t)
			engine.LoadRules(tc.rules, tc.mediaTypes)

			result, err := engine.AnalyzeDirectory(tc.dirPath, tc.files)
			require.NoError(t, err)

			if tc.expectNil {
				assert.Nil(t, result)
				return
			}

			require.NotNil(t, result)
			assert.Equal(t, tc.expectType, result.MediaType.Name)
			assert.GreaterOrEqual(t, result.Confidence, tc.minConfidence)
			assert.NotEmpty(t, result.MatchedPatterns)
		})
	}
}

func TestAnalyzeDirectory_DirectoryStructure(t *testing.T) {
	structurePattern, _ := json.Marshal(map[string]interface{}{
		"required_dirs": []string{"Season"},
		"file_types":    map[string]interface{}{".mkv": 2},
	})

	tests := []struct {
		name      string
		dirPath   string
		files     []FileInfo
		expectNil bool
	}{
		{
			name:    "matches TV show directory structure",
			dirPath: "/media/tvshows/Breaking Bad",
			files: []FileInfo{
				{Name: "Season 1", Path: "/media/tvshows/Breaking Bad/Season 1", IsDir: true},
				{Name: "s01e01.mkv", Size: 1_000_000_000, Extension: ".mkv"},
				{Name: "s01e02.mkv", Size: 1_000_000_000, Extension: ".mkv"},
				{Name: "s01e03.mkv", Size: 1_000_000_000, Extension: ".mkv"},
			},
			expectNil: false,
		},
		{
			name:    "no match without required structure",
			dirPath: "/media/random",
			files: []FileInfo{
				{Name: "file.txt", Size: 100, Extension: ".txt"},
			},
			expectNil: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			engine := newTestEngine(t)
			engine.LoadRules(
				[]models.DetectionRule{
					{ID: 1, MediaTypeID: 2, RuleType: "directory_structure", Pattern: string(structurePattern), ConfidenceWeight: 1.0, Enabled: true, Priority: 10},
				},
				[]models.MediaType{{ID: 2, Name: "tv_show", Description: "TV Shows"}},
			)

			result, err := engine.AnalyzeDirectory(tc.dirPath, tc.files)
			require.NoError(t, err)

			if tc.expectNil {
				assert.Nil(t, result)
			} else {
				require.NotNil(t, result)
				assert.Equal(t, "tv_show", result.MediaType.Name)
				assert.Greater(t, result.Confidence, 0.0)
			}
		})
	}
}

func TestAnalyzeDirectory_FileAnalysis(t *testing.T) {
	contentPattern, _ := json.Marshal(map[string]interface{}{
		"size_patterns": map[string]interface{}{
			"large_video": map[string]interface{}{
				"min_size": 500_000_000.0,
				"max_size": 50_000_000_000.0,
			},
		},
	})

	engine := newTestEngine(t)
	engine.LoadRules(
		[]models.DetectionRule{
			{ID: 1, MediaTypeID: 1, RuleType: "file_analysis", Pattern: string(contentPattern), ConfidenceWeight: 1.0, Enabled: true, Priority: 10},
		},
		[]models.MediaType{{ID: 1, Name: "movie", Description: "Movies"}},
	)

	t.Run("matches large video files by size", func(t *testing.T) {
		result, err := engine.AnalyzeDirectory("/media/movies/Test", []FileInfo{
			{Name: "movie.mkv", Size: 4_000_000_000, Extension: ".mkv"},
		})
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Greater(t, result.Confidence, 0.0)
	})

	t.Run("no match for small files", func(t *testing.T) {
		result, err := engine.AnalyzeDirectory("/media/docs", []FileInfo{
			{Name: "doc.txt", Size: 1000, Extension: ".txt"},
		})
		require.NoError(t, err)
		assert.Nil(t, result)
	})
}

func TestAnalyzeDirectory_Hybrid(t *testing.T) {
	filenamePatternJSON := `["*.mkv", "*.mp4"]`
	structureJSON, _ := json.Marshal(map[string]interface{}{
		"required_dirs": []string{"Season"},
	})
	contentJSON, _ := json.Marshal(map[string]interface{}{
		"size_patterns": map[string]interface{}{
			"large_video": map[string]interface{}{
				"min_size": 500_000_000.0,
				"max_size": 50_000_000_000.0,
			},
		},
	})

	hybridPattern, _ := json.Marshal(map[string]interface{}{
		"filename":  filenamePatternJSON,
		"structure": string(structureJSON),
		"content":   string(contentJSON),
	})

	engine := newTestEngine(t)
	engine.LoadRules(
		[]models.DetectionRule{
			{ID: 1, MediaTypeID: 2, RuleType: "hybrid", Pattern: string(hybridPattern), ConfidenceWeight: 1.0, Enabled: true, Priority: 10},
		},
		[]models.MediaType{{ID: 2, Name: "tv_show", Description: "TV Shows"}},
	)

	result, err := engine.AnalyzeDirectory("/media/tvshows/Show", []FileInfo{
		{Name: "Season 1", IsDir: true},
		{Name: "s01e01.mkv", Size: 1_500_000_000, Extension: ".mkv"},
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "tv_show", result.MediaType.Name)
	assert.Equal(t, "hybrid", result.Method)
}

func TestAnalyzeDirectory_HighestConfidenceWins(t *testing.T) {
	engine := newTestEngine(t)
	engine.LoadRules(
		[]models.DetectionRule{
			{ID: 1, MediaTypeID: 1, RuleType: "filename_pattern", Pattern: `["*.mkv"]`, ConfidenceWeight: 0.5, Enabled: true, Priority: 5},
			{ID: 2, MediaTypeID: 2, RuleType: "filename_pattern", Pattern: `["*.mkv", "*.srt"]`, ConfidenceWeight: 1.0, Enabled: true, Priority: 10},
		},
		[]models.MediaType{
			{ID: 1, Name: "movie", Description: "Movies"},
			{ID: 2, Name: "tv_show", Description: "TV Shows"},
		},
	)

	result, err := engine.AnalyzeDirectory("/media/test", []FileInfo{
		{Name: "video.mkv", Size: 1_000_000_000, Extension: ".mkv"},
		{Name: "video.srt", Size: 50000, Extension: ".srt"},
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	// tv_show rule matches both patterns with higher weight, so it should win
	assert.Equal(t, "tv_show", result.MediaType.Name)
}

func TestExtractTitleAndYear(t *testing.T) {
	engine := newTestEngine(t)

	tests := []struct {
		name          string
		dirPath       string
		expectedTitle string
		expectYear    bool
		expectedYear  int
	}{
		{
			name:          "extracts year in parentheses",
			dirPath:       "/media/movies/Inception (2010)",
			expectedTitle: "Inception",
			expectYear:    true,
			expectedYear:  2010,
		},
		{
			name:          "extracts year in brackets",
			dirPath:       "/media/movies/Blade Runner [1982]",
			expectedTitle: "Blade Runner",
			expectYear:    true,
			expectedYear:  1982,
		},
		{
			name:          "cleans release info from title",
			dirPath:       "/media/movies/The.Matrix.1999.BluRay.1080p.x264",
			expectedTitle: "The Matrix p",
			expectYear:    true,
			expectedYear:  1999,
		},
		{
			name:          "handles title with no year",
			dirPath:       "/media/movies/SomeTitle",
			expectedTitle: "SomeTitle",
			expectYear:    false,
		},
		{
			name:          "cleans dots and underscores",
			dirPath:       "/media/movies/My_Movie.Name",
			expectedTitle: "My Movie Name",
			expectYear:    false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			title, year := engine.extractTitleAndYear(tc.dirPath, nil, "movie")
			assert.Equal(t, tc.expectedTitle, title)
			if tc.expectYear {
				require.NotNil(t, year)
				assert.Equal(t, tc.expectedYear, *year)
			} else {
				assert.Nil(t, year)
			}
		})
	}
}

func TestExtractQualityHints(t *testing.T) {
	engine := newTestEngine(t)

	tests := []struct {
		name     string
		dirPath  string
		files    []FileInfo
		expected []string // hints that should be present
	}{
		{
			name:    "detects 4K quality",
			dirPath: "/media/movies/Movie.4K.UHD",
			files:   nil,
			expected: []string{"4K"},
		},
		{
			name:    "detects 1080p from filename",
			dirPath: "/media/movies/Movie",
			files:   []FileInfo{{Name: "movie.1080p.bluray.mkv"}},
			expected: []string{"1080p", "BluRay"},
		},
		{
			name:    "detects HDR",
			dirPath: "/media/movies/Movie.HDR10",
			files:   nil,
			expected: []string{"HDR"},
		},
		{
			name:    "detects lossless audio",
			dirPath: "/media/music/Album",
			files:   []FileInfo{{Name: "track01.flac"}},
			expected: []string{"Lossless"},
		},
		{
			name:     "no hints for plain directory",
			dirPath:  "/media/docs/notes",
			files:    []FileInfo{{Name: "readme.txt"}},
			expected: []string{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			hints := engine.extractQualityHints(tc.dirPath, tc.files)
			for _, exp := range tc.expected {
				assert.Contains(t, hints, exp)
			}
			if len(tc.expected) == 0 {
				assert.Empty(t, hints)
			}
		})
	}
}

func TestBuildAnalysisData(t *testing.T) {
	engine := newTestEngine(t)

	files := []FileInfo{
		{Name: "movie.mkv", Size: 5 * 1024 * 1024 * 1024, Extension: ".mkv"},  // large
		{Name: "subtitle.srt", Size: 50000, Extension: ".srt"},                  // tiny
		{Name: "poster.jpg", Size: 500 * 1024, Extension: ".jpg"},               // tiny
	}

	data := engine.buildAnalysisData("/media/movies/Test", files, []string{"*.mkv"}, 0.8)

	require.NotNil(t, data)
	assert.Equal(t, []string{"*.mkv"}, data.MatchedPatterns)
	assert.Equal(t, 1, data.FileTypes[".mkv"])
	assert.Equal(t, 1, data.FileTypes[".srt"])
	assert.Equal(t, 1, data.FileTypes[".jpg"])
	assert.Contains(t, data.SizeDistribution, "large")
	assert.Contains(t, data.SizeDistribution, "tiny")
	assert.InDelta(t, 0.32, data.FilenameScore, 0.01)
	assert.InDelta(t, 0.24, data.StructureScore, 0.01)
	assert.InDelta(t, 0.24, data.MetadataScore, 0.01)
}

func TestValidateDetection(t *testing.T) {
	engine := newTestEngine(t)

	tests := []struct {
		name     string
		result   *DetectionResult
		expected bool
	}{
		{
			name:     "nil result is invalid",
			result:   nil,
			expected: false,
		},
		{
			name:     "low confidence is invalid",
			result:   &DetectionResult{Confidence: 0.05, MatchedPatterns: []string{"*.mkv"}},
			expected: false,
		},
		{
			name:     "no matched patterns is invalid",
			result:   &DetectionResult{Confidence: 0.5, MatchedPatterns: []string{}},
			expected: false,
		},
		{
			name:     "valid result",
			result:   &DetectionResult{Confidence: 0.5, MatchedPatterns: []string{"*.mkv"}},
			expected: true,
		},
		{
			name:     "minimum threshold passes",
			result:   &DetectionResult{Confidence: 0.1, MatchedPatterns: []string{"p"}},
			expected: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, engine.ValidateDetection(tc.result))
		})
	}
}

func TestGetSupportedMediaTypes(t *testing.T) {
	engine := newTestEngine(t)

	t.Run("empty when no types loaded", func(t *testing.T) {
		types := engine.GetSupportedMediaTypes()
		assert.Empty(t, types)
	})

	t.Run("returns loaded types", func(t *testing.T) {
		engine.LoadRules(nil, []models.MediaType{
			{ID: 1, Name: "movie"},
			{ID: 2, Name: "tv_show"},
			{ID: 3, Name: "music"},
		})
		types := engine.GetSupportedMediaTypes()
		assert.Len(t, types, 3)
	})
}

func TestGlobToRegex(t *testing.T) {
	engine := newTestEngine(t)

	tests := []struct {
		name    string
		glob    string
		input   string
		matches bool
	}{
		{"star matches any extension", "*.mkv", "movie.mkv", true},
		{"star matches any prefix", "*.mkv", "inception.mkv", true},
		{"no match different extension", "*.mkv", "movie.avi", false},
		{"question mark matches single char", "s01e0?.mkv", "s01e01.mkv", true},
		{"case insensitive matching", "*.MKV", "movie.mkv", true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			re, err := engine.globToRegex(tc.glob)
			require.NoError(t, err)
			assert.Equal(t, tc.matches, re.MatchString(tc.input))
		})
	}
}

func TestParseInt(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"2010", 2010},
		{"0", 0},
		{"1234567890", 1234567890},
		{"abc", 0},
		{"12ab34", 1234},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			assert.Equal(t, tc.expected, parseInt(tc.input))
		})
	}
}
