package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestDetectMediaType_EpisodePatternsClassifiedAsTVEpisode is the §11.4.115
// RED-baseline-on-the-broken-artifact regression guard for the scanner
// mis-classifying TV episodes as movies.
//
// Captured DB evidence (live catalogizer DB): 399 of the 1192 items with
// media_type_id=1 (movie) are actually TV episodes whose directory name leads
// with an SxxExx / "Episode N" token (e.g. "S01E09 The Perfect Storm",
// "S2E02 - The Awful Truth", "Episode 5 - Kissed by Fire", "S03E12 The Lady
// Killer"). They pollute the movie shelf and can never match TMDB as movies.
//
// Root cause (FACT, proven by probe): AggregationService.detectMediaType
// classifies a directory via ParseTVShow, whose tvSxxExxRe requires a NON-EMPTY
// show-name prefix BEFORE the SxxExx token (`^(.+?)[\s._-]+S(\d{1,2})E(\d{1,2})`).
// A LEADING SxxExx ("S01E09 ...") therefore yields Season==nil && Episode==nil,
// so the title falls through to the fragile substring hack (only "s01"/"s02"
// recognised) and finally to the "hasVideo => movie" default.
//
// Conservative scope: only unambiguous episode tokens (SxxExx, leading
// "Episode N", NxNN) reclassify. Real movies bearing numbers ("Ocean's Eleven",
// "2012", "Se7en", "Apollo 13") MUST stay movie.
func TestDetectMediaType_EpisodePatternsClassifiedAsTVEpisode(t *testing.T) {
	s := &AggregationService{}

	cases := []struct {
		name      string // directory name as seen by the scanner
		wantType  string
		wantTitle string // empty => not asserted
	}{
		// Positives: leading / un-prefixed episode tokens that previously leaked
		// to "movie" (or were mis-bucketed as "tv_show" by the s01/s02 hack).
		{"S01E09 The Perfect Storm", "tv_episode", ""},
		{"S2E02 - The Awful Truth", "tv_episode", ""},
		{"Episode 5 - Kissed by Fire", "tv_episode", ""},
		{"S03E12 The Lady Killer", "tv_episode", ""},
		{"s1e9 Some Episode", "tv_episode", ""},
		{"S1.E09 Dotted Sep", "tv_episode", ""},
		{"1x09 Alternative Notation", "tv_episode", ""},

		// Existing show-prefixed SxxExx behaviour MUST be preserved (tv_show),
		// so the show-level aggregation and TestDetectMediaType_TVShow keep working.
		{"Breaking Bad S01E01 720p", "tv_show", "Breaking Bad"},
		{"Game of Thrones Season 3", "tv_show", ""},

		// Negatives: real movies with numbers MUST stay movie (§11.4.6).
		{"Ocean's Eleven", "movie", "Ocean's Eleven"},
		{"2012", "movie", "2012"},
		{"Se7en", "movie", "Se7en"},
		{"Apollo 13", "movie", "Apollo 13"},
		{"The Matrix (1999)", "movie", "The Matrix"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := directoryInfo{
				name:      tc.name,
				fileCount: 1,
				fileTypes: map[string]int{".mkv": 1},
			}
			gotType, parsed := s.detectMediaType(dir)
			assert.Equal(t, tc.wantType, gotType, "media type for %q", tc.name)
			// Every classified item MUST carry a non-empty title (the scanner
			// uses parsed.Title as the media_items primary key surface).
			assert.NotEmpty(t, parsed.Title, "title for %q must not be empty", tc.name)
			if tc.wantTitle != "" {
				assert.Equal(t, tc.wantTitle, parsed.Title, "title for %q", tc.name)
			}
		})
	}
}
