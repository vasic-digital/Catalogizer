package services

import (
	"testing"
)

func TestMediaTypeVideoConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant MediaType
		expected string
	}{
		{"MediaTypeMovie", MediaTypeMovie, "movie"},
		{"MediaTypeTV", MediaTypeTV, "tv"},
		{"MediaTypeTVSeries", MediaTypeTVSeries, "tv_series"},
		{"MediaTypeTVEpisode", MediaTypeTVEpisode, "tv_episode"},
		{"MediaTypeConcert", MediaTypeConcert, "concert"},
		{"MediaTypeDocumentary", MediaTypeDocumentary, "documentary"},
		{"MediaTypeCourse", MediaTypeCourse, "course"},
		{"MediaTypeTraining", MediaTypeTraining, "training"},
		{"MediaTypeVideo", MediaTypeVideo, "video"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.constant) != tt.expected {
				t.Errorf("expected %s = '%s', got '%s'", tt.name, tt.expected, string(tt.constant))
			}
		})
	}
}

func TestMediaTypeAudioConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant MediaType
		expected string
	}{
		{"MediaTypeMusic", MediaTypeMusic, "music"},
		{"MediaTypeAlbum", MediaTypeAlbum, "album"},
		{"MediaTypeAudiobook", MediaTypeAudiobook, "audiobook"},
		{"MediaTypePodcast", MediaTypePodcast, "podcast"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.constant) != tt.expected {
				t.Errorf("expected %s = '%s', got '%s'", tt.name, tt.expected, string(tt.constant))
			}
		})
	}
}

func TestMediaTypeGamesSoftwareConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant MediaType
		expected string
	}{
		{"MediaTypeGame", MediaTypeGame, "game"},
		{"MediaTypeGameOS", MediaTypeGameOS, "game_os"},
		{"MediaTypeSoftware", MediaTypeSoftware, "software"},
		{"MediaTypeSoftwareOS", MediaTypeSoftwareOS, "software_os"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.constant) != tt.expected {
				t.Errorf("expected %s = '%s', got '%s'", tt.name, tt.expected, string(tt.constant))
			}
		})
	}
}

func TestMediaTypeBooksDocumentsConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant MediaType
		expected string
	}{
		{"MediaTypeBook", MediaTypeBook, "book"},
		{"MediaTypeEbook", MediaTypeEbook, "ebook"},
		{"MediaTypeComicBook", MediaTypeComicBook, "comic_book"},
		{"MediaTypeMagazine", MediaTypeMagazine, "magazine"},
		{"MediaTypeNewspaper", MediaTypeNewspaper, "newspaper"},
		{"MediaTypeJournal", MediaTypeJournal, "journal"},
		{"MediaTypeManual", MediaTypeManual, "manual"},
		{"MediaTypeDocument", MediaTypeDocument, "document"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.constant) != tt.expected {
				t.Errorf("expected %s = '%s', got '%s'", tt.name, tt.expected, string(tt.constant))
			}
		})
	}
}

func TestMediaTypeAdditionalConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant MediaType
		expected string
	}{
		{"MediaTypeImage", MediaTypeImage, "image"},
		{"MediaTypeUnknown", MediaTypeUnknown, "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.constant) != tt.expected {
				t.Errorf("expected %s = '%s', got '%s'", tt.name, tt.expected, string(tt.constant))
			}
		})
	}
}

func TestMediaTypeIsStringType(t *testing.T) {
	var mt MediaType = "custom_type"
	if string(mt) != "custom_type" {
		t.Errorf("expected MediaType to hold arbitrary string 'custom_type', got '%s'", string(mt))
	}

	// Verify type conversion works
	s := string(MediaTypeMovie)
	if s != "movie" {
		t.Errorf("expected string(MediaTypeMovie) = 'movie', got '%s'", s)
	}

	// Verify assignment from string
	mt = MediaType("movie")
	if mt != MediaTypeMovie {
		t.Errorf("expected MediaType('movie') == MediaTypeMovie")
	}
}

func TestMediaTypeComparison(t *testing.T) {
	// Same values should be equal
	if MediaTypeTV == MediaTypeTVSeries {
		t.Error("MediaTypeTV and MediaTypeTVSeries should not be equal (tv vs tv_series)")
	}

	// Identical values
	var a MediaType = "movie"
	if a != MediaTypeMovie {
		t.Error("expected MediaType('movie') to equal MediaTypeMovie")
	}
}

func TestPlaybackStateConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant PlaybackState
		expected string
	}{
		{"PlaybackStatePlaying", PlaybackStatePlaying, "playing"},
		{"PlaybackStatePaused", PlaybackStatePaused, "paused"},
		{"PlaybackStateStopped", PlaybackStateStopped, "stopped"},
		{"PlaybackStateLoading", PlaybackStateLoading, "loading"},
		{"PlaybackStateError", PlaybackStateError, "error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.constant) != tt.expected {
				t.Errorf("expected %s = '%s', got '%s'", tt.name, tt.expected, string(tt.constant))
			}
		})
	}
}

func TestPlaybackStateIsStringType(t *testing.T) {
	var ps PlaybackState = "buffering"
	if string(ps) != "buffering" {
		t.Errorf("expected PlaybackState to hold arbitrary string 'buffering', got '%s'", string(ps))
	}

	ps = PlaybackState("playing")
	if ps != PlaybackStatePlaying {
		t.Error("expected PlaybackState('playing') == PlaybackStatePlaying")
	}
}

func TestPlaybackStateComparison(t *testing.T) {
	if PlaybackStatePlaying == PlaybackStatePaused {
		t.Error("PlaybackStatePlaying and PlaybackStatePaused should not be equal")
	}
	if PlaybackStateStopped == PlaybackStateError {
		t.Error("PlaybackStateStopped and PlaybackStateError should not be equal")
	}
}

func TestRepeatModeConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant RepeatMode
		expected string
	}{
		{"RepeatModeOff", RepeatModeOff, "off"},
		{"RepeatModeTrack", RepeatModeTrack, "track"},
		{"RepeatModeAlbum", RepeatModeAlbum, "album"},
		{"RepeatModeAll", RepeatModeAll, "all"},
		{"RepeatModeShuffle", RepeatModeShuffle, "shuffle"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.constant) != tt.expected {
				t.Errorf("expected %s = '%s', got '%s'", tt.name, tt.expected, string(tt.constant))
			}
		})
	}
}

func TestRepeatModeIsStringType(t *testing.T) {
	var rm RepeatMode = "playlist"
	if string(rm) != "playlist" {
		t.Errorf("expected RepeatMode to hold arbitrary string 'playlist', got '%s'", string(rm))
	}

	rm = RepeatMode("off")
	if rm != RepeatModeOff {
		t.Error("expected RepeatMode('off') == RepeatModeOff")
	}
}

func TestRepeatModeComparison(t *testing.T) {
	if RepeatModeOff == RepeatModeAll {
		t.Error("RepeatModeOff and RepeatModeAll should not be equal")
	}
	if RepeatModeTrack == RepeatModeAlbum {
		t.Error("RepeatModeTrack and RepeatModeAlbum should not be equal")
	}
}

func TestQualityConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant Quality
		expected string
	}{
		{"QualityLow", QualityLow, "low"},
		{"QualityUltra", QualityUltra, "ultra"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.constant) != tt.expected {
				t.Errorf("expected %s = '%s', got '%s'", tt.name, tt.expected, string(tt.constant))
			}
		})
	}
}

func TestQualityIsStringType(t *testing.T) {
	var q Quality = "medium"
	if string(q) != "medium" {
		t.Errorf("expected Quality to hold arbitrary string 'medium', got '%s'", string(q))
	}

	q = Quality("low")
	if q != QualityLow {
		t.Error("expected Quality('low') == QualityLow")
	}
}

func TestQualityComparison(t *testing.T) {
	if QualityLow == QualityUltra {
		t.Error("QualityLow and QualityUltra should not be equal")
	}
}

func TestAllMediaTypeConstantsAreUnique(t *testing.T) {
	allTypes := []MediaType{
		MediaTypeMovie, MediaTypeTV, MediaTypeTVSeries, MediaTypeTVEpisode,
		MediaTypeConcert, MediaTypeDocumentary, MediaTypeCourse, MediaTypeTraining,
		MediaTypeVideo, MediaTypeMusic, MediaTypeAlbum, MediaTypeAudiobook,
		MediaTypePodcast, MediaTypeGame, MediaTypeGameOS, MediaTypeSoftware,
		MediaTypeSoftwareOS, MediaTypeBook, MediaTypeEbook, MediaTypeComicBook,
		MediaTypeMagazine, MediaTypeNewspaper, MediaTypeJournal, MediaTypeManual,
		MediaTypeDocument, MediaTypeImage, MediaTypeUnknown,
	}

	seen := make(map[MediaType]string)
	for _, mt := range allTypes {
		if prev, exists := seen[mt]; exists {
			// MediaTypeTV and MediaTypeTVSeries are intentionally different ("tv" vs "tv_series")
			t.Errorf("duplicate MediaType value '%s': found in both '%s' and current", string(mt), prev)
		}
		seen[mt] = string(mt)
	}
}

func TestAllPlaybackStateConstantsAreUnique(t *testing.T) {
	allStates := []PlaybackState{
		PlaybackStatePlaying, PlaybackStatePaused, PlaybackStateStopped,
		PlaybackStateLoading, PlaybackStateError,
	}

	seen := make(map[PlaybackState]bool)
	for _, ps := range allStates {
		if seen[ps] {
			t.Errorf("duplicate PlaybackState value '%s'", string(ps))
		}
		seen[ps] = true
	}
}

func TestAllRepeatModeConstantsAreUnique(t *testing.T) {
	allModes := []RepeatMode{
		RepeatModeOff, RepeatModeTrack, RepeatModeAlbum,
		RepeatModeAll, RepeatModeShuffle,
	}

	seen := make(map[RepeatMode]bool)
	for _, rm := range allModes {
		if seen[rm] {
			t.Errorf("duplicate RepeatMode value '%s'", string(rm))
		}
		seen[rm] = true
	}
}
