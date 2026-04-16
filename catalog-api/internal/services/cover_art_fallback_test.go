package services

import (
	"strings"
	"testing"
)

func TestGeneratePlaceholderSVG_KnownTypes(t *testing.T) {
	knownTypes := []struct {
		mediaType     string
		expectedLabel string
	}{
		{"movie", "MOVIE"},
		{"tv_show", "TV SHOW"},
		{"tv_season", "SEASON"},
		{"tv_episode", "EPISODE"},
		{"music_artist", "ARTIST"},
		{"music_album", "ALBUM"},
		{"song", "SONG"},
		{"game", "GAME"},
		{"software", "SOFTWARE"},
		{"book", "BOOK"},
		{"comic", "COMIC"},
	}

	for _, tc := range knownTypes {
		t.Run(tc.mediaType, func(t *testing.T) {
			svg := GeneratePlaceholderSVG(tc.mediaType)
			svgStr := string(svg)

			if len(svg) == 0 {
				t.Fatalf("GeneratePlaceholderSVG(%q) returned empty", tc.mediaType)
			}

			if !strings.Contains(svgStr, "<svg") {
				t.Errorf("expected SVG content, got: %s", svgStr[:100])
			}

			if !strings.Contains(svgStr, tc.expectedLabel) {
				t.Errorf("expected label %q for type %s, not found in SVG", tc.expectedLabel, tc.mediaType)
			}

			if !strings.Contains(svgStr, `width="300"`) {
				t.Errorf("expected width 300 in SVG for type %s", tc.mediaType)
			}

			if !strings.Contains(svgStr, `height="450"`) {
				t.Errorf("expected height 450 in SVG for type %s", tc.mediaType)
			}
		})
	}
}

func TestGeneratePlaceholderSVG_UnknownType(t *testing.T) {
	svg := GeneratePlaceholderSVG("unknown_type")
	svgStr := string(svg)

	if len(svg) == 0 {
		t.Fatal("GeneratePlaceholderSVG returned empty for unknown type")
	}

	if !strings.Contains(svgStr, "<svg") {
		t.Error("expected valid SVG for unknown type")
	}

	// Should use default accent color
	if !strings.Contains(svgStr, "#e94560") {
		t.Error("expected default accent color #e94560 for unknown type")
	}

	// Should use the type name as label (uppercase, underscores replaced)
	if !strings.Contains(svgStr, "UNKNOWN TYPE") {
		t.Error("expected label 'UNKNOWN TYPE' for unknown_type media type")
	}
}

func TestGeneratePlaceholderSVG_EmptyType(t *testing.T) {
	svg := GeneratePlaceholderSVG("")
	svgStr := string(svg)

	if len(svg) == 0 {
		t.Fatal("GeneratePlaceholderSVG returned empty for empty type")
	}

	if !strings.Contains(svgStr, "<svg") {
		t.Error("expected valid SVG for empty type")
	}

	// Should use default accent color
	if !strings.Contains(svgStr, "#e94560") {
		t.Error("expected default accent color #e94560 for empty type")
	}
}

func TestGeneratePlaceholderSVG_ValidSVGStructure(t *testing.T) {
	svg := GeneratePlaceholderSVG("movie")
	svgStr := string(svg)

	// Check required SVG elements
	requiredElements := []string{
		`xmlns="http://www.w3.org/2000/svg"`,
		`viewBox="0 0 300 450"`,
		"<rect",
		"<text",
		"<defs>",
		"<linearGradient",
		`fill="url(#bg)"`,
		`rx="8"`,
	}

	for _, elem := range requiredElements {
		if !strings.Contains(svgStr, elem) {
			t.Errorf("SVG missing required element: %s", elem)
		}
	}
}

func TestCoverURLRequest_Struct(t *testing.T) {
	req := CoverURLRequest{
		ID:            42,
		MediaTypeName: "movie",
	}

	if req.ID != 42 {
		t.Errorf("expected ID 42, got %d", req.ID)
	}
	if req.MediaTypeName != "movie" {
		t.Errorf("expected MediaTypeName 'movie', got %s", req.MediaTypeName)
	}
}

func TestGetCoverURL_NilDB(t *testing.T) {
	// CoverArtService with nil DB should return placeholder
	service := &CoverArtService{
		db:     nil,
		logger: nil,
	}

	url := service.GetCoverURL(nil, 1, "movie")
	expected := "/api/v1/cover/placeholder/movie"
	if url != expected {
		t.Errorf("expected %q, got %q", expected, url)
	}
}

func TestGetCoverURLsBatch_Empty(t *testing.T) {
	service := &CoverArtService{
		db:     nil,
		logger: nil,
	}

	result := service.GetCoverURLsBatch(nil, nil)
	if len(result) != 0 {
		t.Errorf("expected empty map, got %d entries", len(result))
	}

	result = service.GetCoverURLsBatch(nil, []CoverURLRequest{})
	if len(result) != 0 {
		t.Errorf("expected empty map, got %d entries", len(result))
	}
}

func TestGetCoverURLsBatch_NilDB_PlaceholderFallback(t *testing.T) {
	service := &CoverArtService{
		db:     nil,
		logger: nil,
	}

	items := []CoverURLRequest{
		{ID: 1, MediaTypeName: "movie"},
		{ID: 2, MediaTypeName: "tv_show"},
		{ID: 3, MediaTypeName: "song"},
	}

	result := service.GetCoverURLsBatch(nil, items)

	if len(result) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(result))
	}

	expected := map[int64]string{
		1: "/api/v1/cover/placeholder/movie",
		2: "/api/v1/cover/placeholder/tv_show",
		3: "/api/v1/cover/placeholder/song",
	}

	for id, want := range expected {
		if got := result[id]; got != want {
			t.Errorf("item %d: expected %q, got %q", id, want, got)
		}
	}
}
