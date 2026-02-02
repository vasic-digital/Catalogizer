package detector

import (
	"catalogizer/internal/media/models"
	"encoding/json"
	"fmt"
	"testing"

	"go.uber.org/zap"
)

func newBenchEngine() *DetectionEngine {
	logger, _ := zap.NewProduction()
	return NewDetectionEngine(logger)
}

func generateFiles(n int) []FileInfo {
	files := make([]FileInfo, n)
	exts := []string{".mkv", ".mp4", ".avi", ".srt", ".nfo", ".jpg", ".txt"}
	for i := 0; i < n; i++ {
		ext := exts[i%len(exts)]
		files[i] = FileInfo{
			Name:      fmt.Sprintf("file_%04d%s", i, ext),
			Path:      fmt.Sprintf("/media/collection/file_%04d%s", i, ext),
			Size:      int64((i + 1)) * 100 * 1024 * 1024,
			Extension: ext,
			IsDir:     false,
		}
	}
	return files
}

// --- AnalyzeDirectory benchmarks ---

func BenchmarkAnalyzeDirectory(b *testing.B) {
	sizes := []int{5, 50, 500}
	for _, size := range sizes {
		b.Run(fmt.Sprintf("files=%d", size), func(b *testing.B) {
			engine := newBenchEngine()
			engine.LoadRules(
				[]models.DetectionRule{
					{ID: 1, MediaTypeID: 1, RuleType: "filename_pattern", Pattern: `["*.mkv", "*.mp4", "*.avi"]`, ConfidenceWeight: 1.0, Enabled: true, Priority: 10},
					{ID: 2, MediaTypeID: 2, RuleType: "filename_pattern", Pattern: `["*.srt", "*.nfo"]`, ConfidenceWeight: 0.8, Enabled: true, Priority: 5},
				},
				[]models.MediaType{
					{ID: 1, Name: "movie", Description: "Movies"},
					{ID: 2, Name: "tv_show", Description: "TV Shows"},
				},
			)
			files := generateFiles(size)
			dirPath := "/media/movies/Inception (2010) 1080p BluRay x264"

			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_, _ = engine.AnalyzeDirectory(dirPath, files)
			}
		})
	}
}

func BenchmarkAnalyzeDirectory_MultipleRuleTypes(b *testing.B) {
	engine := newBenchEngine()

	structurePattern, _ := json.Marshal(map[string]interface{}{
		"required_dirs": []string{"Season", "Extras"},
		"file_types":    map[string]interface{}{".mkv": 2, ".srt": 1},
	})
	contentPattern, _ := json.Marshal(map[string]interface{}{
		"size_patterns": map[string]interface{}{
			"large_video": map[string]interface{}{
				"min_size": 500_000_000.0,
				"max_size": 50_000_000_000.0,
			},
		},
	})

	engine.LoadRules(
		[]models.DetectionRule{
			{ID: 1, MediaTypeID: 1, RuleType: "filename_pattern", Pattern: `["*.mkv", "*.mp4"]`, ConfidenceWeight: 1.0, Enabled: true, Priority: 10},
			{ID: 2, MediaTypeID: 2, RuleType: "directory_structure", Pattern: string(structurePattern), ConfidenceWeight: 1.0, Enabled: true, Priority: 8},
			{ID: 3, MediaTypeID: 1, RuleType: "file_analysis", Pattern: string(contentPattern), ConfidenceWeight: 0.9, Enabled: true, Priority: 6},
		},
		[]models.MediaType{
			{ID: 1, Name: "movie", Description: "Movies"},
			{ID: 2, Name: "tv_show", Description: "TV Shows"},
		},
	)

	files := []FileInfo{
		{Name: "Season 1", IsDir: true},
		{Name: "s01e01.mkv", Size: 1_500_000_000, Extension: ".mkv"},
		{Name: "s01e02.mkv", Size: 1_500_000_000, Extension: ".mkv"},
		{Name: "s01e01.srt", Size: 50_000, Extension: ".srt"},
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = engine.AnalyzeDirectory("/media/tvshows/Breaking Bad", files)
	}
}

// --- ExtractTitleAndYear benchmarks ---

func BenchmarkExtractTitleAndYear(b *testing.B) {
	engine := newBenchEngine()

	benchmarks := []struct {
		name    string
		dirPath string
	}{
		{"simple_year_parens", "/media/movies/Inception (2010)"},
		{"year_brackets", "/media/movies/Blade Runner [1982]"},
		{"dotted_with_release_info", "/media/movies/The.Matrix.1999.BluRay.1080p.x264.DTS"},
		{"no_year", "/media/movies/SomeTitle"},
		{"complex_name", "/media/movies/The.Lord.of.the.Rings.The.Fellowship.of.the.Ring.2001.Extended.BluRay.4K.HDR.x265.DTS-HD"},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_, _ = engine.extractTitleAndYear(bm.dirPath, nil, "movie")
			}
		})
	}
}

// --- ExtractQualityHints benchmarks ---

func BenchmarkExtractQualityHints(b *testing.B) {
	engine := newBenchEngine()

	benchmarks := []struct {
		name    string
		dirPath string
		files   []FileInfo
	}{
		{
			"no_hints",
			"/media/docs/notes",
			[]FileInfo{{Name: "readme.txt"}},
		},
		{
			"single_hint",
			"/media/movies/Movie.1080p",
			nil,
		},
		{
			"many_hints",
			"/media/movies/Movie.4K.HDR10.BluRay.Remux",
			[]FileInfo{
				{Name: "movie.2160p.bluray.remux.mkv"},
				{Name: "movie.dts-hd.flac.mkv"},
			},
		},
		{
			"large_file_list",
			"/media/movies/Collection",
			generateFiles(100),
		},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = engine.extractQualityHints(bm.dirPath, bm.files)
			}
		})
	}
}
