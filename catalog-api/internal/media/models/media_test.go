package models

import (
	"encoding/json"
	"testing"
)

func TestMediaType_MarshalDetectionPatterns(t *testing.T) {
	mt := &MediaType{
		DetectionPatterns: []string{"*.mp4", "*.mkv", "*.avi"},
	}
	data, err := mt.MarshalDetectionPatterns()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var patterns []string
	if err := json.Unmarshal(data, &patterns); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}
	if len(patterns) != 3 {
		t.Errorf("expected 3 patterns, got %d", len(patterns))
	}
}

func TestMediaType_UnmarshalDetectionPatterns(t *testing.T) {
	mt := &MediaType{}
	data, _ := json.Marshal([]string{"*.mp4", "*.mkv"})
	if err := mt.UnmarshalDetectionPatterns(data); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mt.DetectionPatterns) != 2 {
		t.Errorf("expected 2 patterns, got %d", len(mt.DetectionPatterns))
	}
}

func TestMediaItem_MarshalGenre(t *testing.T) {
	mi := &MediaItem{
		Genre: []string{"Action", "Drama"},
	}
	data, err := mi.MarshalGenre()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var genres []string
	if err := json.Unmarshal(data, &genres); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if len(genres) != 2 || genres[0] != "Action" {
		t.Errorf("unexpected genres: %v", genres)
	}
}

func TestMediaItem_UnmarshalGenre(t *testing.T) {
	mi := &MediaItem{}
	data, _ := json.Marshal([]string{"Comedy"})
	if err := mi.UnmarshalGenre(data); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mi.Genre) != 1 || mi.Genre[0] != "Comedy" {
		t.Errorf("unexpected genre: %v", mi.Genre)
	}
}

func TestMediaItem_MarshalCastCrew(t *testing.T) {
	director := "Christopher Nolan"
	mi := &MediaItem{
		CastCrew: &CastCrew{
			Director: &director,
			Actors: []Actor{
				{Name: "Leonardo DiCaprio", Character: "Cobb", Order: 1},
			},
		},
	}
	data, err := mi.MarshalCastCrew()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var cc CastCrew
	if err := json.Unmarshal(data, &cc); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if cc.Director == nil || *cc.Director != "Christopher Nolan" {
		t.Errorf("unexpected director: %v", cc.Director)
	}
}

func TestMediaItem_UnmarshalCastCrew(t *testing.T) {
	mi := &MediaItem{}
	cc := CastCrew{Actors: []Actor{{Name: "Actor1"}}}
	data, _ := json.Marshal(cc)
	if err := mi.UnmarshalCastCrew(data); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mi.CastCrew == nil || len(mi.CastCrew.Actors) != 1 {
		t.Errorf("unexpected cast crew: %v", mi.CastCrew)
	}
}

func TestQualityInfo_IsBetterThan(t *testing.T) {
	tests := []struct {
		name     string
		qi       *QualityInfo
		other    *QualityInfo
		expected bool
	}{
		{
			name:     "higher score is better",
			qi:       &QualityInfo{QualityScore: 90},
			other:    &QualityInfo{QualityScore: 70},
			expected: true,
		},
		{
			name:     "lower score is not better",
			qi:       &QualityInfo{QualityScore: 50},
			other:    &QualityInfo{QualityScore: 80},
			expected: false,
		},
		{
			name:     "nil self returns false",
			qi:       nil,
			other:    &QualityInfo{QualityScore: 80},
			expected: false,
		},
		{
			name:     "nil other returns true",
			qi:       &QualityInfo{QualityScore: 80},
			other:    nil,
			expected: true,
		},
		{
			name:     "both nil returns false",
			qi:       nil,
			other:    nil,
			expected: false,
		},
		{
			name:     "equal scores",
			qi:       &QualityInfo{QualityScore: 80},
			other:    &QualityInfo{QualityScore: 80},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.qi.IsBetterThan(tt.other)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestQualityInfo_GetDisplayName(t *testing.T) {
	profile := "BluRay 1080p"
	tests := []struct {
		name     string
		qi       *QualityInfo
		expected string
	}{
		{
			name:     "with quality profile",
			qi:       &QualityInfo{QualityProfile: &profile},
			expected: "BluRay 1080p",
		},
		{
			name:     "with 4K resolution",
			qi:       &QualityInfo{Resolution: &Resolution{Width: 3840, Height: 2160}},
			expected: "4K/UHD",
		},
		{
			name:     "with 1080p resolution",
			qi:       &QualityInfo{Resolution: &Resolution{Width: 1920, Height: 1080}},
			expected: "1080p",
		},
		{
			name:     "without profile or resolution",
			qi:       &QualityInfo{},
			expected: "Unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.qi.GetDisplayName()
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestResolution_GetDisplayName(t *testing.T) {
	tests := []struct {
		name     string
		r        Resolution
		expected string
	}{
		{"4K", Resolution{Width: 3840, Height: 2160}, "4K/UHD"},
		{"1080p", Resolution{Width: 1920, Height: 1080}, "1080p"},
		{"720p", Resolution{Width: 1280, Height: 720}, "720p"},
		{"480p", Resolution{Width: 720, Height: 480}, "480p/DVD"},
		{"low quality", Resolution{Width: 320, Height: 240}, "Low Quality"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.r.GetDisplayName()
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestMediaItem_JSONRoundTrip(t *testing.T) {
	year := 2010
	desc := "A mind-bending thriller"
	mi := MediaItem{
		ID:          1,
		Title:       "Inception",
		Year:        &year,
		Description: &desc,
		Genre:       []string{"Action", "Sci-Fi"},
		Status:      "active",
	}

	data, err := json.Marshal(mi)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var result MediaItem
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if result.Title != "Inception" {
		t.Errorf("expected title 'Inception', got %q", result.Title)
	}
	if result.Year == nil || *result.Year != 2010 {
		t.Errorf("expected year 2010, got %v", result.Year)
	}
	if len(result.Genre) != 2 {
		t.Errorf("expected 2 genres, got %d", len(result.Genre))
	}
}

func TestMediaSearchRequest_JSONRoundTrip(t *testing.T) {
	minRating := 7.5
	req := MediaSearchRequest{
		Query:     "inception",
		MediaTypes: []string{"video"},
		MinRating: &minRating,
		Limit:     20,
		Offset:    0,
		SortBy:    "rating",
		SortOrder: "desc",
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var result MediaSearchRequest
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if result.Query != "inception" {
		t.Errorf("expected query 'inception', got %q", result.Query)
	}
	if result.MinRating == nil || *result.MinRating != 7.5 {
		t.Errorf("expected min rating 7.5, got %v", result.MinRating)
	}
}

func TestYearRange(t *testing.T) {
	yr := YearRange{From: 2000, To: 2025}
	data, err := json.Marshal(yr)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var result YearRange
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if result.From != 2000 || result.To != 2025 {
		t.Errorf("unexpected year range: %+v", result)
	}
}
