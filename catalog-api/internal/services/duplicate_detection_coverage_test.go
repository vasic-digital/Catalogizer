package services

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// =============================================================================
// detectDuplicatesForMediaType — the core per-type detection logic
// =============================================================================

func TestDuplicateDetectionService_DetectDuplicatesForMediaType_LessThanTwoItems(t *testing.T) {
	logger := zap.NewNop()
	service := NewDuplicateDetectionService(nil, logger, nil)

	// getMediaItems returns empty by default (stub), so < 2 items
	req := &DuplicateDetectionRequest{
		MinSimilarity: 0.5,
		BatchSize:     100,
	}

	groups, err := service.detectDuplicatesForMediaType(context.Background(), MediaTypeMovie, req)
	require.NoError(t, err)
	assert.Empty(t, groups)
}

func TestDuplicateDetectionService_DetectDuplicatesForMediaType_NoItems(t *testing.T) {
	logger := zap.NewNop()
	service := NewDuplicateDetectionService(nil, logger, nil)

	req := &DuplicateDetectionRequest{
		MinSimilarity: 0.8,
	}

	groups, err := service.detectDuplicatesForMediaType(context.Background(), MediaTypeMusic, req)
	require.NoError(t, err)
	assert.Empty(t, groups)
}

// =============================================================================
// determinePrimaryItem — stub but still callable
// =============================================================================

func TestDuplicateDetectionService_DeterminePrimaryItem_Ext(t *testing.T) {
	logger := zap.NewNop()
	service := NewDuplicateDetectionService(nil, logger, nil)

	group := &DuplicateGroup{
		ID:        "test-group",
		MediaType: MediaTypeMovie,
		PrimaryItem: DuplicateItem{
			MediaID:  "item1",
			Title:    "Primary Item",
			FileSize: 1000,
		},
		DuplicateItems: []DuplicateItem{
			{MediaID: "item2", Title: "Dup Item 1", FileSize: 500},
			{MediaID: "item3", Title: "Dup Item 2", FileSize: 2000},
		},
	}

	// Should not panic
	assert.NotPanics(t, func() {
		service.determinePrimaryItem(group)
	})
}

func TestDuplicateDetectionService_DeterminePrimaryItem_EmptyGroup(t *testing.T) {
	logger := zap.NewNop()
	service := NewDuplicateDetectionService(nil, logger, nil)

	group := &DuplicateGroup{
		ID:             "empty-group",
		DuplicateItems: []DuplicateItem{},
	}

	assert.NotPanics(t, func() {
		service.determinePrimaryItem(group)
	})
}

// =============================================================================
// analyzeFieldMatches — stub but still callable
// =============================================================================

func TestDuplicateDetectionService_AnalyzeFieldMatches_Ext(t *testing.T) {
	logger := zap.NewNop()
	service := NewDuplicateDetectionService(nil, logger, nil)

	item1 := &DuplicateItem{
		MediaID:  "a",
		Title:    "Item A",
		Artist:   "Artist A",
		Year:     2020,
		FileSize: 1000,
	}
	item2 := &DuplicateItem{
		MediaID:  "b",
		Title:    "Item B",
		Artist:   "Artist B",
		Year:     2021,
		FileSize: 2000,
	}

	analysis := &SimilarityAnalysis{
		Components:       make(map[string]float64),
		MatchingFields:   []string{},
		DifferencesFound: []string{},
	}

	// Should not panic
	assert.NotPanics(t, func() {
		service.analyzeFieldMatches(item1, item2, analysis)
	})
}

func TestDuplicateDetectionService_AnalyzeFieldMatches_IdenticalItems(t *testing.T) {
	logger := zap.NewNop()
	service := NewDuplicateDetectionService(nil, logger, nil)

	item := &DuplicateItem{
		MediaID:  "same",
		Title:    "Same Item",
		Artist:   "Same Artist",
		Year:     2020,
		FileSize: 1000,
	}

	analysis := &SimilarityAnalysis{
		Components:       make(map[string]float64),
		MatchingFields:   []string{},
		DifferencesFound: []string{},
	}

	assert.NotPanics(t, func() {
		service.analyzeFieldMatches(item, item, analysis)
	})
}

// =============================================================================
// DetectDuplicates — top-level detection with empty DB
// =============================================================================

func TestDuplicateDetectionService_DetectDuplicates_AllMediaTypes(t *testing.T) {
	logger := zap.NewNop()
	service := NewDuplicateDetectionService(nil, logger, nil)

	req := &DuplicateDetectionRequest{
		MediaTypes:    []MediaType{}, // Empty = all types
		MinSimilarity: 0.8,
		UserID:        1,
	}

	groups, err := service.DetectDuplicates(context.Background(), req)
	require.NoError(t, err)
	assert.Empty(t, groups) // No items in stub DB
}

func TestDuplicateDetectionService_DetectDuplicates_SpecificMediaTypes(t *testing.T) {
	logger := zap.NewNop()
	service := NewDuplicateDetectionService(nil, logger, nil)

	req := &DuplicateDetectionRequest{
		MediaTypes:    []MediaType{MediaTypeMovie, MediaTypeMusic},
		MinSimilarity: 0.7,
		UserID:        1,
	}

	groups, err := service.DetectDuplicates(context.Background(), req)
	require.NoError(t, err)
	assert.Empty(t, groups)
}

// =============================================================================
// calculateSimilarity — comprehensive similarity calculation tests
// =============================================================================

func TestDuplicateDetectionService_CalculateSimilarity_ComprehensiveComparison(t *testing.T) {
	logger := zap.NewNop()
	service := NewDuplicateDetectionService(nil, logger, nil)

	tests := []struct {
		name      string
		item1     *DuplicateItem
		item2     *DuplicateItem
		mediaType MediaType
		minScore  float64
		maxScore  float64
	}{
		{
			name: "identical movies",
			item1: &DuplicateItem{
				Title:    "The Matrix",
				Director: "Wachowski Sisters",
				Year:     1999,
				Duration: 8160000,
				Quality:  "1080p",
			},
			item2: &DuplicateItem{
				Title:    "The Matrix",
				Director: "Wachowski Sisters",
				Year:     1999,
				Duration: 8160000,
				Quality:  "1080p",
			},
			mediaType: MediaTypeMovie,
			minScore:  0.5,
			maxScore:  1.0,
		},
		{
			name: "different movies",
			item1: &DuplicateItem{
				Title:    "The Matrix",
				Director: "Wachowski Sisters",
				Year:     1999,
			},
			item2: &DuplicateItem{
				Title:    "Inception",
				Director: "Christopher Nolan",
				Year:     2010,
			},
			mediaType: MediaTypeMovie,
			minScore:  0.0,
			maxScore:  0.5,
		},
		{
			name: "identical music",
			item1: &DuplicateItem{
				Title:    "Bohemian Rhapsody",
				Artist:   "Queen",
				Album:    "A Night at the Opera",
				Year:     1975,
				Duration: 354000,
				Bitrate:  320,
			},
			item2: &DuplicateItem{
				Title:    "Bohemian Rhapsody",
				Artist:   "Queen",
				Album:    "A Night at the Opera",
				Year:     1975,
				Duration: 354000,
				Bitrate:  320,
			},
			mediaType: MediaTypeMusic,
			minScore:  0.5,
			maxScore:  1.0,
		},
		{
			name: "identical books",
			item1: &DuplicateItem{
				Title:  "Clean Code",
				Author: "Robert C. Martin",
				Year:   2008,
			},
			item2: &DuplicateItem{
				Title:  "Clean Code",
				Author: "Robert C. Martin",
				Year:   2008,
			},
			mediaType: MediaTypeBook,
			minScore:  0.5,
			maxScore:  1.0,
		},
		{
			name: "identical software",
			item1: &DuplicateItem{
				Title:    "Visual Studio Code",
				FileSize: 100000000,
				Metadata: map[string]interface{}{"version": "1.85", "platform": "windows"},
			},
			item2: &DuplicateItem{
				Title:    "Visual Studio Code",
				FileSize: 100000000,
				Metadata: map[string]interface{}{"version": "1.85", "platform": "windows"},
			},
			mediaType: MediaTypeSoftware,
			minScore:  0.3,
			maxScore:  1.0,
		},
		{
			name: "generic items",
			item1: &DuplicateItem{
				Title:    "Some File",
				FileName: "file1.dat",
				FileSize: 5000,
				Format:   "dat",
			},
			item2: &DuplicateItem{
				Title:    "Some File",
				FileName: "file1.dat",
				FileSize: 5000,
				Format:   "dat",
			},
			mediaType: MediaType("unknown"),
			minScore:  0.3,
			maxScore:  1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.item1.ExternalIDs == nil {
				tt.item1.ExternalIDs = map[string]string{}
			}
			if tt.item2.ExternalIDs == nil {
				tt.item2.ExternalIDs = map[string]string{}
			}
			if tt.item1.Fingerprints == nil {
				tt.item1.Fingerprints = map[string]string{}
			}
			if tt.item2.Fingerprints == nil {
				tt.item2.Fingerprints = map[string]string{}
			}

			analysis := service.calculateSimilarity(context.Background(), tt.item1, tt.item2, tt.mediaType)
			assert.GreaterOrEqual(t, analysis.OverallScore, tt.minScore,
				"score should be >= %f for %s", tt.minScore, tt.name)
			assert.LessOrEqual(t, analysis.OverallScore, tt.maxScore,
				"score should be <= %f for %s", tt.maxScore, tt.name)
		})
	}
}

func TestDuplicateDetectionService_CalculateSimilarity_WithFingerprints(t *testing.T) {
	logger := zap.NewNop()
	service := NewDuplicateDetectionService(nil, logger, nil)

	item1 := &DuplicateItem{
		Title:        "Test",
		ExternalIDs:  map[string]string{},
		Fingerprints: map[string]string{"audio": "abc123", "video": "def456"},
	}
	item2 := &DuplicateItem{
		Title:        "Test",
		ExternalIDs:  map[string]string{},
		Fingerprints: map[string]string{"audio": "abc123", "video": "ghi789"},
	}

	analysis := service.calculateSimilarity(context.Background(), item1, item2, MediaTypeMusic)
	assert.NotNil(t, analysis)
	assert.GreaterOrEqual(t, analysis.OverallScore, 0.0)
}

// =============================================================================
// calculateBookMetadataSimilarity — detailed book metadata comparison
// =============================================================================

func TestDuplicateDetectionService_CalculateBookMetadataSimilarity_Coverage(t *testing.T) {
	logger := zap.NewNop()
	service := NewDuplicateDetectionService(nil, logger, nil)

	tests := []struct {
		name     string
		item1    *DuplicateItem
		item2    *DuplicateItem
		minScore float64
	}{
		{
			name: "same author and year",
			item1: &DuplicateItem{
				Author:      "Stephen King",
				Year:        1977,
				ExternalIDs: map[string]string{},
			},
			item2: &DuplicateItem{
				Author:      "Stephen King",
				Year:        1977,
				ExternalIDs: map[string]string{},
			},
			minScore: 0.9,
		},
		{
			name: "same ISBN",
			item1: &DuplicateItem{
				ExternalIDs: map[string]string{"isbn": "9780134685991"},
			},
			item2: &DuplicateItem{
				ExternalIDs: map[string]string{"isbn": "9780134685991"},
			},
			minScore: 0.0, // extractISBN is a stub returning ""
		},
		{
			name: "empty metadata",
			item1: &DuplicateItem{
				ExternalIDs: map[string]string{},
			},
			item2: &DuplicateItem{
				ExternalIDs: map[string]string{},
			},
			minScore: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := service.calculateBookMetadataSimilarity(tt.item1, tt.item2)
			assert.GreaterOrEqual(t, score, tt.minScore)
			assert.LessOrEqual(t, score, 1.0)
		})
	}
}

// =============================================================================
// calculateSoftwareMetadataSimilarity — software metadata comparison
// =============================================================================

func TestDuplicateDetectionService_CalculateSoftwareMetadataSimilarity_Coverage(t *testing.T) {
	logger := zap.NewNop()
	service := NewDuplicateDetectionService(nil, logger, nil)

	tests := []struct {
		name     string
		item1    *DuplicateItem
		item2    *DuplicateItem
		minScore float64
	}{
		{
			name: "same file size",
			item1: &DuplicateItem{
				FileSize: 100000000,
				Metadata: map[string]interface{}{},
			},
			item2: &DuplicateItem{
				FileSize: 100000000,
				Metadata: map[string]interface{}{},
			},
			minScore: 0.0,
		},
		{
			name: "different file sizes",
			item1: &DuplicateItem{
				FileSize: 100000000,
				Metadata: map[string]interface{}{},
			},
			item2: &DuplicateItem{
				FileSize: 200000000,
				Metadata: map[string]interface{}{},
			},
			minScore: 0.0,
		},
		{
			name: "empty metadata",
			item1: &DuplicateItem{
				Metadata: map[string]interface{}{},
			},
			item2: &DuplicateItem{
				Metadata: map[string]interface{}{},
			},
			minScore: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := service.calculateSoftwareMetadataSimilarity(tt.item1, tt.item2)
			assert.GreaterOrEqual(t, score, tt.minScore)
			assert.LessOrEqual(t, score, 1.0)
		})
	}
}

// =============================================================================
// calculateGenericMetadataSimilarity — generic metadata comparison
// =============================================================================

func TestDuplicateDetectionService_CalculateGenericMetadataSimilarity_Coverage(t *testing.T) {
	logger := zap.NewNop()
	service := NewDuplicateDetectionService(nil, logger, nil)

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
				Format:   "mkv",
				FileSize: 5000000,
			},
			item2: &DuplicateItem{
				FileName: "movie.mkv",
				Format:   "mkv",
				FileSize: 5000000,
			},
			minScore: 0.9,
		},
		{
			name: "different format same name",
			item1: &DuplicateItem{
				FileName: "movie.avi",
				Format:   "avi",
				FileSize: 5000000,
			},
			item2: &DuplicateItem{
				FileName: "movie.mkv",
				Format:   "mkv",
				FileSize: 5000000,
			},
			minScore: 0.0,
		},
		{
			name:     "empty items",
			item1:    &DuplicateItem{},
			item2:    &DuplicateItem{},
			minScore: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := service.calculateGenericMetadataSimilarity(tt.item1, tt.item2)
			assert.GreaterOrEqual(t, score, tt.minScore)
			assert.LessOrEqual(t, score, 1.0)
		})
	}
}

// =============================================================================
// Stub helper methods — ensure they don't panic
// =============================================================================

func TestDuplicateDetectionService_StubMethods(t *testing.T) {
	logger := zap.NewNop()
	service := NewDuplicateDetectionService(nil, logger, nil)

	t.Run("storeDuplicateGroup", func(t *testing.T) {
		err := service.storeDuplicateGroup(context.Background(), &DuplicateGroup{ID: "test"})
		assert.NoError(t, err)
	})

	t.Run("getMediaItems", func(t *testing.T) {
		items, err := service.getMediaItems(context.Background(), MediaTypeMovie, &DuplicateDetectionRequest{})
		assert.NoError(t, err)
		assert.Empty(t, items)
	})

	t.Run("createOrAddToDuplicateGroup", func(t *testing.T) {
		group := service.createOrAddToDuplicateGroup(
			nil,
			DuplicateItem{MediaID: "a"},
			DuplicateItem{MediaID: "b"},
			&SimilarityAnalysis{OverallScore: 0.9},
			MediaTypeMovie,
		)
		assert.Nil(t, group) // stub returns nil
	})

	t.Run("mergeDuplicateGroup", func(t *testing.T) {
		groups := service.mergeDuplicateGroup([]DuplicateGroup{}, DuplicateGroup{ID: "new"})
		assert.NotNil(t, groups)
	})

	t.Run("calculateHashSimilarity", func(t *testing.T) {
		score := service.calculateHashSimilarity("abc", "abc", "audio")
		assert.Equal(t, 0.0, score)
	})

	t.Run("areQualitiesSimilar", func(t *testing.T) {
		assert.False(t, service.areQualitiesSimilar("1080p", "720p"))
	})

	t.Run("extractISBN", func(t *testing.T) {
		isbn := service.extractISBN(map[string]string{"isbn": "123"})
		assert.Empty(t, isbn)
	})

	t.Run("areISBNsRelated", func(t *testing.T) {
		assert.False(t, service.areISBNsRelated("isbn1", "isbn2"))
	})

	t.Run("extractVersion", func(t *testing.T) {
		v := service.extractVersion(map[string]interface{}{"version": "1.0"})
		assert.Empty(t, v)
	})

	t.Run("extractPlatform", func(t *testing.T) {
		p := service.extractPlatform(map[string]interface{}{"platform": "windows"})
		assert.Empty(t, p)
	})

	t.Run("calculateVersionSimilarity", func(t *testing.T) {
		score := service.calculateVersionSimilarity("1.0", "1.0")
		assert.Equal(t, 0.0, score)
	})

	t.Run("arePlatformsSimilar", func(t *testing.T) {
		assert.False(t, service.arePlatformsSimilar("windows", "linux"))
	})

	t.Run("areFormatsSimilar", func(t *testing.T) {
		assert.False(t, service.areFormatsSimilar("mkv", "avi"))
	})
}

// =============================================================================
// DuplicateGroup / DuplicateItem struct field coverage
// =============================================================================

func TestDuplicateGroup_StructFields(t *testing.T) {
	now := time.Now()
	group := DuplicateGroup{
		ID:        "grp-001",
		MediaType: MediaTypeMovie,
		PrimaryItem: DuplicateItem{
			MediaID:  "primary",
			FilePath: "/movies/movie.mkv",
			FileName: "movie.mkv",
			FileSize: 5000000000,
			FileHash: "sha256abc",
			Title:    "The Movie",
		},
		DuplicateItems:  []DuplicateItem{{MediaID: "dup1"}, {MediaID: "dup2"}},
		Confidence:      0.95,
		DetectionMethod: "hash_match",
		MatchTypes:      []string{"file_hash", "title"},
		CreatedAt:       now,
		UpdatedAt:       now,
		Status:          "pending",
		AutoResolved:    false,
	}

	assert.Equal(t, "grp-001", group.ID)
	assert.Equal(t, 2, len(group.DuplicateItems))
	assert.Equal(t, "pending", group.Status)
}

func TestDeduplicationAction_StructFields(t *testing.T) {
	action := DeduplicationAction{
		GroupID:       "grp-001",
		Action:        "keep_primary",
		PrimaryItemID: "primary",
		ItemsToRemove: []string{"dup1"},
		ItemsToKeep:   []string{"primary"},
		MergeStrategy: "",
		UserID:        42,
		Reason:        "User selected primary",
		ExecutedAt:    time.Now(),
	}

	assert.Equal(t, "grp-001", action.GroupID)
	assert.Equal(t, "keep_primary", action.Action)
	assert.Equal(t, int64(42), action.UserID)
}

func TestSimilarityAnalysis_StructFields(t *testing.T) {
	analysis := SimilarityAnalysis{
		OverallScore:          0.85,
		TitleSimilarity:       0.9,
		MetadataSimilarity:    0.8,
		FingerprintSimilarity: 0.7,
		FileSimilarity:        0.6,
		ExternalIDMatch:       true,
		HashMatch:             false,
		Components:            map[string]float64{"title": 0.45, "metadata": 0.4},
		MatchingFields:        []string{"title", "year"},
		DifferencesFound:      []string{"quality"},
	}

	assert.Equal(t, 0.85, analysis.OverallScore)
	assert.True(t, analysis.ExternalIDMatch)
	assert.False(t, analysis.HashMatch)
}

func TestTextSimilarityMetrics_StructFields(t *testing.T) {
	metrics := TextSimilarityMetrics{
		LevenshteinDistance: 3,
		JaroWinklerScore:   0.85,
		CosineSimilarity:   0.9,
		JaccardIndex:       0.75,
		LCSRatio:           0.8,
		SoundexMatch:       true,
		MetaphoneMatch:     false,
	}

	assert.Equal(t, 3, metrics.LevenshteinDistance)
	assert.True(t, metrics.SoundexMatch)
}
