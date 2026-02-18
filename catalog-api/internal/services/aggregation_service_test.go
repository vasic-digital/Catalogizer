package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetectMediaType_VideoDirectory(t *testing.T) {
	s := &AggregationService{}

	dir := directoryInfo{
		name:      "The Matrix (1999)",
		fileCount: 3,
		fileTypes: map[string]int{".mkv": 1, ".srt": 2},
	}

	typeName, parsed := s.detectMediaType(dir)
	assert.Equal(t, "movie", typeName)
	assert.Equal(t, "The Matrix", parsed.Title)
	assert.NotNil(t, parsed.Year)
	assert.Equal(t, 1999, *parsed.Year)
}

func TestDetectMediaType_TVShow(t *testing.T) {
	s := &AggregationService{}

	dir := directoryInfo{
		name:      "Breaking Bad S01E01 720p",
		fileCount: 1,
		fileTypes: map[string]int{".mkv": 1},
	}

	typeName, parsed := s.detectMediaType(dir)
	assert.Equal(t, "tv_show", typeName)
	assert.Equal(t, "Breaking Bad", parsed.Title)
	assert.NotNil(t, parsed.Season)
	assert.Equal(t, 1, *parsed.Season)
}

func TestDetectMediaType_TVShowSeasonFolder(t *testing.T) {
	s := &AggregationService{}

	dir := directoryInfo{
		name:      "Game of Thrones Season 3",
		fileCount: 10,
		fileTypes: map[string]int{".mkv": 10},
	}

	typeName, _ := s.detectMediaType(dir)
	assert.Equal(t, "tv_show", typeName)
}

func TestDetectMediaType_MusicAlbum(t *testing.T) {
	s := &AggregationService{}

	dir := directoryInfo{
		name:      "Pink Floyd - The Wall",
		fileCount: 13,
		fileTypes: map[string]int{".flac": 13},
	}

	typeName, parsed := s.detectMediaType(dir)
	assert.Equal(t, "music_album", typeName)
	assert.Equal(t, "Pink Floyd", parsed.Artist)
	assert.Equal(t, "The Wall", parsed.Album)
}

func TestDetectMediaType_Software(t *testing.T) {
	s := &AggregationService{}

	dir := directoryInfo{
		name:      "Ubuntu 24.04",
		fileCount: 1,
		fileTypes: map[string]int{".iso": 1},
	}

	typeName, _ := s.detectMediaType(dir)
	assert.Equal(t, "software", typeName)
}

func TestDetectMediaType_Comic(t *testing.T) {
	s := &AggregationService{}

	dir := directoryInfo{
		name:      "Spider-Man #300",
		fileCount: 1,
		fileTypes: map[string]int{".cbr": 1},
	}

	typeName, _ := s.detectMediaType(dir)
	assert.Equal(t, "comic", typeName)
}

func TestDetectMediaType_Book(t *testing.T) {
	s := &AggregationService{}

	dir := directoryInfo{
		name:      "Dune - Frank Herbert",
		fileCount: 1,
		fileTypes: map[string]int{".epub": 1},
	}

	typeName, _ := s.detectMediaType(dir)
	assert.Equal(t, "book", typeName)
}

func TestDetectMediaTypeFromPath(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{"/movies/test.mkv", "video"},
		{"/music/song.mp3", "audio"},
		{"/software/ubuntu.iso", "disc_image"},
		{"/books/novel.epub", "ebook"},
		{"/comics/issue.cbr", "comic"},
		{"/apps/setup.exe", "software"},
		{"/docs/readme.txt", "other"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			assert.Equal(t, tt.expected, DetectMediaTypeFromPath(tt.path))
		})
	}
}
