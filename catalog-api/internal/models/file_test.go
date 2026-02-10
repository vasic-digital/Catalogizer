package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestFileItem_JSON(t *testing.T) {
	item := FileItem{
		Name: "movie.mp4",
		Type: "video",
		Path: "/media/movies/movie.mp4",
	}

	data, err := json.Marshal(item)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var result FileItem
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if result.Name != "movie.mp4" {
		t.Errorf("expected name 'movie.mp4', got %q", result.Name)
	}
	if result.Type != "video" {
		t.Errorf("expected type 'video', got %q", result.Type)
	}
}

func TestSMBPath_JSON(t *testing.T) {
	path := SMBPath{
		Server: "nas01",
		Share:  "media",
		Path:   "/movies/action",
		Valid:  true,
	}

	data, err := json.Marshal(path)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var result SMBPath
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if result.Server != "nas01" || !result.Valid {
		t.Errorf("unexpected result: %+v", result)
	}
}

func TestFileInfo_JSON(t *testing.T) {
	hash := "abc123"
	ext := ".mp4"
	fi := FileInfo{
		ID:          1,
		Name:        "movie.mp4",
		Path:        "/media/movies/movie.mp4",
		IsDirectory: false,
		Type:        "file",
		Size:        1024000,
		Hash:        &hash,
		Extension:   &ext,
		SmbRoot:     "nas01",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	data, err := json.Marshal(fi)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var result FileInfo
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if result.Name != "movie.mp4" {
		t.Errorf("expected name 'movie.mp4', got %q", result.Name)
	}
	if result.Hash == nil || *result.Hash != "abc123" {
		t.Errorf("unexpected hash: %v", result.Hash)
	}
}

func TestMediaTypeConstants(t *testing.T) {
	types := []string{
		MediaTypeVideo,
		MediaTypeAudio,
		MediaTypeImage,
		MediaTypeText,
		MediaTypeBook,
		MediaTypeGame,
		MediaTypeOther,
	}

	seen := make(map[string]bool)
	for _, mt := range types {
		if mt == "" {
			t.Error("media type constant should not be empty")
		}
		if seen[mt] {
			t.Errorf("duplicate media type constant: %q", mt)
		}
		seen[mt] = true
	}

	if len(types) != 7 {
		t.Errorf("expected 7 media type constants, got %d", len(types))
	}
}

func TestOverallStats_JSON(t *testing.T) {
	stats := OverallStats{
		TotalFiles:         1000,
		TotalDirectories:   50,
		TotalSize:          1024 * 1024 * 1024,
		TotalDuplicates:    20,
		DuplicateGroups:    10,
		StorageRootsCount:  3,
		ActiveStorageRoots: 2,
	}

	data, err := json.Marshal(stats)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var result OverallStats
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if result.TotalFiles != 1000 {
		t.Errorf("expected 1000 total files, got %d", result.TotalFiles)
	}
}

func TestDirectoryStats_JSON(t *testing.T) {
	stats := DirectoryStats{
		Path:           "/media/movies",
		TotalSize:      500000000,
		FileCount:      100,
		DirectoryCount: 20,
		DuplicateCount: 5,
	}

	data, err := json.Marshal(stats)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var result DirectoryStats
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if result.Path != "/media/movies" {
		t.Errorf("expected path '/media/movies', got %q", result.Path)
	}
}

func TestDuplicateGroup_JSON(t *testing.T) {
	group := DuplicateGroup{
		Hash:      "sha256:abc123",
		Size:      1024000,
		Count:     3,
		TotalSize: 3072000,
		Files: []FileInfo{
			{Name: "copy1.mp4", Path: "/a/copy1.mp4"},
			{Name: "copy2.mp4", Path: "/b/copy2.mp4"},
		},
	}

	data, err := json.Marshal(group)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var result DuplicateGroup
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if result.Count != 3 {
		t.Errorf("expected count 3, got %d", result.Count)
	}
	if len(result.Files) != 2 {
		t.Errorf("expected 2 files, got %d", len(result.Files))
	}
}

func TestSearchRequest_JSON(t *testing.T) {
	minSize := int64(1024)
	req := SearchRequest{
		Query:   "movie",
		Path:    "/media",
		MinSize: &minSize,
		Limit:   50,
		SortBy:  "name",
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var result SearchRequest
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if result.Query != "movie" {
		t.Errorf("expected query 'movie', got %q", result.Query)
	}
	if result.MinSize == nil || *result.MinSize != 1024 {
		t.Errorf("unexpected min size: %v", result.MinSize)
	}
}

func TestSizeDistribution_JSON(t *testing.T) {
	dist := SizeDistribution{
		Tiny:    100,
		Small:   200,
		Medium:  150,
		Large:   80,
		Huge:    30,
		Massive: 10,
	}

	data, err := json.Marshal(dist)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var result SizeDistribution
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if result.Tiny != 100 || result.Massive != 10 {
		t.Errorf("unexpected distribution: %+v", result)
	}
}

func TestCopyRequest_JSON(t *testing.T) {
	req := CopyRequest{
		SourcePath:      "/media/src/file.mp4",
		DestinationPath: "/media/dst/file.mp4",
		SmbRoot:         "nas01",
		Overwrite:       true,
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var result CopyRequest
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if !result.Overwrite {
		t.Error("expected overwrite to be true")
	}
}

func TestDownloadRequest_JSON(t *testing.T) {
	req := DownloadRequest{
		Paths:   []string{"/media/file1.mp4", "/media/file2.mp4"},
		Format:  "zip",
		SmbRoot: "nas01",
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var result DownloadRequest
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if len(result.Paths) != 2 {
		t.Errorf("expected 2 paths, got %d", len(result.Paths))
	}
	if result.Format != "zip" {
		t.Errorf("expected format 'zip', got %q", result.Format)
	}
}

func TestFileWithMetadata_JSON(t *testing.T) {
	fwm := FileWithMetadata{
		File: &FileInfo{
			Name: "movie.mp4",
			Path: "/media/movie.mp4",
			Size: 1024000,
		},
		Metadata: &MediaMetadata{
			Title: "Test Movie",
			Genre: "Action",
		},
	}

	data, err := json.Marshal(fwm)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var result FileWithMetadata
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if result.File == nil || result.File.Name != "movie.mp4" {
		t.Errorf("unexpected file: %+v", result.File)
	}
	if result.Metadata == nil || result.Metadata.Title != "Test Movie" {
		t.Errorf("unexpected metadata: %+v", result.Metadata)
	}
}

func TestGrowthTrends_JSON(t *testing.T) {
	trends := GrowthTrends{
		MonthlyGrowth: []MonthlyGrowth{
			{Month: "2026-01", FilesAdded: 100, SizeAdded: 1024000},
			{Month: "2026-02", FilesAdded: 150, SizeAdded: 2048000},
		},
		TotalGrowthRate: 0.15,
		FileGrowthRate:  0.10,
		SizeGrowthRate:  0.20,
	}

	data, err := json.Marshal(trends)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var result GrowthTrends
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if len(result.MonthlyGrowth) != 2 {
		t.Errorf("expected 2 monthly entries, got %d", len(result.MonthlyGrowth))
	}
}
