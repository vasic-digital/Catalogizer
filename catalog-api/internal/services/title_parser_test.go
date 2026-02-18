package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseMovieTitle(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		title        string
		year         *int
		qualityHints []string
	}{
		{
			name:  "parenthesized year",
			input: "The Matrix (1999)",
			title: "The Matrix",
			year:  intPtr(1999),
		},
		{
			name:         "dotted format with quality",
			input:        "The.Matrix.1999.1080p.BluRay",
			title:        "The Matrix",
			year:         intPtr(1999),
			qualityHints: []string{"1080p", "BluRay"},
		},
		{
			name:         "4K HDR REMUX",
			input:        "Inception.2010.2160p.UHD.HDR.REMUX",
			title:        "Inception",
			year:         intPtr(2010),
			qualityHints: []string{"2160p", "HDR", "REMUX"},
		},
		{
			name:  "no year",
			input: "Some Movie",
			title: "Some Movie",
			year:  nil,
		},
		{
			name:         "web-dl",
			input:        "Movie.Name.2023.WEB-DL.720p",
			title:        "Movie Name",
			year:         intPtr(2023),
			qualityHints: []string{"720p", "WEB-DL"},
		},
		{
			name:  "bracketed year",
			input: "The Matrix [1999]",
			title: "The Matrix",
			year:  intPtr(1999),
		},
		{
			name:         "DTS and Atmos",
			input:        "Dune.2021.1080p.BluRay.DTS.Atmos",
			title:        "Dune",
			year:         intPtr(2021),
			qualityHints: []string{"1080p", "BluRay", "DTS", "Atmos"},
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
			if tt.qualityHints != nil {
				for _, q := range tt.qualityHints {
					assert.Contains(t, result.QualityHints, q)
				}
			}
		})
	}
}

func TestParseTVShow(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		title   string
		season  *int
		episode *int
	}{
		{
			name:    "S01E02 format",
			input:   "Breaking.Bad.S01E02.720p",
			title:   "Breaking Bad",
			season:  intPtr(1),
			episode: intPtr(2),
		},
		{
			name:    "lowercase s01e05",
			input:   "the.office.s03e05.hdtv",
			title:   "the office",
			season:  intPtr(3),
			episode: intPtr(5),
		},
		{
			name:   "Season folder",
			input:  "Breaking Bad Season 1",
			title:  "Breaking Bad",
			season: intPtr(1),
		},
		{
			name:   "Season folder S format",
			input:  "Game of Thrones S04",
			title:  "Game of Thrones",
			season: intPtr(4),
		},
		{
			name:  "Complete series",
			input: "Friends Complete",
			title: "Friends",
		},
		{
			name:    "1x02 format",
			input:   "Seinfeld 3x15 The Boyfriend",
			title:   "Seinfeld",
			season:  intPtr(3),
			episode: intPtr(15),
		},
		{
			name:    "Season Episode spelled out",
			input:   "The Sopranos Season 2 Episode 5",
			title:   "The Sopranos",
			season:  intPtr(2),
			episode: intPtr(5),
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

func TestParseMusicAlbum(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		artist string
		album  string
		year   *int
	}{
		{
			name:   "artist dash album",
			input:  "Pink Floyd - The Wall",
			artist: "Pink Floyd",
			album:  "The Wall",
		},
		{
			name:   "with year",
			input:  "Nirvana - Nevermind (1991)",
			artist: "Nirvana",
			album:  "Nevermind",
			year:   intPtr(1991),
		},
		{
			name:   "slash separated",
			input:  "Pink Floyd/The Wall",
			artist: "Pink Floyd",
			album:  "The Wall",
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

func TestParseGameTitle(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		title    string
		platform string
		year     *int
	}{
		{
			name:     "with platform in parens",
			input:    "Half-Life 2 (PC)",
			title:    "Half-Life 2",
			platform: "PC",
		},
		{
			name:     "with platform and year",
			input:    "Half-Life 2 (2004) PC",
			title:    "Half-Life 2",
			platform: "PC",
			year:     intPtr(2004),
		},
		{
			name:     "bracketed platform",
			input:    "The Legend of Zelda [Switch]",
			title:    "The Legend of Zelda",
			platform: "Switch",
		},
		{
			name:  "simple name",
			input: "Minecraft",
			title: "Minecraft",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseGameTitle(tt.input)
			assert.Equal(t, tt.title, result.Title)
			if tt.platform != "" {
				assert.Equal(t, tt.platform, result.Platform)
			}
			if tt.year != nil {
				assert.NotNil(t, result.Year)
				assert.Equal(t, *tt.year, *result.Year)
			}
		})
	}
}

func TestParseSoftwareTitle(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		title   string
		version string
	}{
		{
			name:    "with version",
			input:   "Ubuntu 24.04",
			title:   "Ubuntu",
			version: "24.04",
		},
		{
			name:    "versioned software",
			input:   "VLC.Media.Player.v3.0.18",
			title:   "VLC Media Player",
			version: "3.0.18",
		},
		{
			name:    "dotted version",
			input:   "VLC 3.0.20",
			title:   "VLC",
			version: "3.0.20",
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

func TestCleanTitle(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"The.Matrix.1999", "The Matrix 1999"},
		{"some_movie_name", "some movie name"},
		{"  spaced  out  ", "spaced out"},
		{"Movie.Name.1080p.BluRay", "Movie Name"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.expected, CleanTitle(tt.input))
		})
	}
}

func TestExtractYear(t *testing.T) {
	tests := []struct {
		input    string
		expected *int
	}{
		{"Movie (2023)", intPtr(2023)},
		{"Movie [1999]", intPtr(1999)},
		{"Movie.2010.BluRay", intPtr(2010)},
		{"no year here", nil},
		{"year 1899 too old", nil},
		{"year 2100 too new", nil},
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

// intPtr is declared in subtitle_service_test.go in this package.
