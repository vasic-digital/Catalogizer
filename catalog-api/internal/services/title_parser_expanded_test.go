package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// ParseMovieTitle — expanded edge cases
// ---------------------------------------------------------------------------

func TestParseMovieTitle_Expanded(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		title        string
		year         *int
		qualityHints []string
	}{
		{
			name:  "simple title no year",
			input: "Casablanca",
			title: "Casablanca",
			year:  nil,
		},
		{
			name:         "multiple quality tags",
			input:        "Avatar.2009.3D.1080p.BluRay.x264.DTS",
			title:        "Avatar",
			year:         intPtr(2009),
			qualityHints: []string{"1080p", "BluRay", "DTS"},
		},
		{
			name:  "title with dots only",
			input: "Some.Movie",
			title: "Some Movie",
			year:  nil,
		},
		{
			name:  "year 2024",
			input: "Movie Name (2024)",
			title: "Movie Name",
			year:  intPtr(2024),
		},
		{
			name:  "year 1900 boundary",
			input: "Old Movie (1900)",
			title: "Old Movie",
			year:  intPtr(1900),
		},
		{
			name:         "REMUX tag",
			input:        "Film.2022.REMUX.2160p",
			title:        "Film",
			year:         intPtr(2022),
			qualityHints: []string{"REMUX", "2160p"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseMovieTitle(tt.input)
			assert.Equal(t, tt.title, result.Title)
			if tt.year != nil {
				assert.NotNil(t, result.Year)
				assert.Equal(t, *tt.year, *result.Year)
			} else {
				assert.Nil(t, result.Year)
			}
			for _, q := range tt.qualityHints {
				assert.Contains(t, result.QualityHints, q)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// ParseTVShow — expanded
// ---------------------------------------------------------------------------

func TestParseTVShow_Expanded(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		title   string
		season  *int
		episode *int
	}{
		{
			name:    "uppercase S E",
			input:   "House.of.Cards.S03E12",
			title:   "House of Cards",
			season:  intPtr(3),
			episode: intPtr(12),
		},
		{
			name:    "double digit season",
			input:   "Supernatural.S15E20",
			title:   "Supernatural",
			season:  intPtr(15),
			episode: intPtr(20),
		},
		{
			name:   "Season keyword only",
			input:  "The Wire Season 4",
			title:  "The Wire",
			season: intPtr(4),
		},
		{
			name:  "just show name",
			input: "Stranger Things",
			title: "Stranger Things",
		},
		{
			name:    "mixed case",
			input:   "Grey.s.Anatomy.s19e15",
			title:   "Grey s Anatomy",
			season:  intPtr(19),
			episode: intPtr(15),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseTVShow(tt.input)
			assert.Equal(t, tt.title, result.Title)
			if tt.season != nil {
				assert.NotNil(t, result.Season)
				assert.Equal(t, *tt.season, *result.Season)
			}
			if tt.episode != nil {
				assert.NotNil(t, result.Episode)
				assert.Equal(t, *tt.episode, *result.Episode)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// ParseMusicAlbum — expanded
// ---------------------------------------------------------------------------

func TestParseMusicAlbum_Expanded(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		artist string
		album  string
		year   *int
	}{
		{
			name:   "em dash separator",
			input:  "Radiohead - OK Computer",
			artist: "Radiohead",
			album:  "OK Computer",
		},
		{
			name:   "with year in brackets",
			input:  "The Beatles - Abbey Road [1969]",
			artist: "The Beatles",
			album:  "Abbey Road",
			year:   intPtr(1969),
		},
		{
			name:   "single word artist",
			input:  "Adele - 25",
			artist: "Adele",
			album:  "25",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseMusicAlbum(tt.input)
			if tt.artist != "" {
				assert.Equal(t, tt.artist, result.Artist)
			}
			if tt.album != "" {
				assert.Equal(t, tt.album, result.Album)
			}
			if tt.year != nil {
				assert.NotNil(t, result.Year)
				assert.Equal(t, *tt.year, *result.Year)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// ParseGameTitle — expanded
// ---------------------------------------------------------------------------

func TestParseGameTitle_Expanded(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		title    string
		platform string
	}{
		{
			name:     "Xbox One",
			input:    "Halo Infinite (Xbox One)",
			title:    "Halo Infinite",
			platform: "Xbox One",
		},
		{
			name:  "no platform",
			input: "Tetris",
			title: "Tetris",
		},
		{
			name:     "PS5",
			input:    "God of War Ragnarok [PS5]",
			title:    "God of War Ragnarok",
			platform: "PS5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseGameTitle(tt.input)
			assert.Equal(t, tt.title, result.Title)
			if tt.platform != "" {
				assert.Equal(t, tt.platform, result.Platform)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// ParseSoftwareTitle — expanded
// ---------------------------------------------------------------------------

func TestParseSoftwareTitle_Expanded(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		title   string
		version string
	}{
		{
			name:    "simple version",
			input:   "Firefox 120.0",
			title:   "Firefox",
			version: "120.0",
		},
		{
			name:    "v prefix",
			input:   "Node.js.v20.11.0",
			title:   "Node js",
			version: "20.11.0",
		},
		{
			name:  "no version",
			input: "LibreOffice",
			title: "LibreOffice",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseSoftwareTitle(tt.input)
			assert.Equal(t, tt.title, result.Title)
			if tt.version != "" {
				assert.Equal(t, tt.version, result.Version)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// CleanTitle — expanded
// ---------------------------------------------------------------------------

func TestCleanTitle_Expanded(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Normal Title", "Normal Title"},
		{"Under_score_title", "Under score title"},
		{"Dot.separated.title", "Dot separated title"},
		{"   extra   spaces   ", "extra spaces"},
		{"Mixed.Under_score.And.Dots", "Mixed Under score And Dots"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := CleanTitle(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// ---------------------------------------------------------------------------
// ExtractYear — expanded
// ---------------------------------------------------------------------------

func TestExtractYear_Expanded(t *testing.T) {
	tests := []struct {
		input    string
		expected *int
	}{
		{"Film (2023)", intPtr(2023)},
		{"Film [2000]", intPtr(2000)},
		{"Film.1999.720p", intPtr(1999)},
		{"Film 2024 BluRay", intPtr(2024)},
		{"no year here at all", nil},
		{"1899 is too old", nil},
		{"2100 is too new", nil},
		{"2099 is valid", intPtr(2099)},
		{"1900 is valid", intPtr(1900)},
		{"", nil},
		{"12345", nil},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := ExtractYear(tt.input)
			if tt.expected != nil {
				assert.NotNil(t, result)
				assert.Equal(t, *tt.expected, *result)
			} else {
				assert.Nil(t, result)
			}
		})
	}
}
