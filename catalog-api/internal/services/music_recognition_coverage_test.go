package services

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// =============================================================================
// RecognizeMedia — top-level music recognition
// =============================================================================

func TestMusicRecognitionProvider_RecognizeMedia_ByMetadata(t *testing.T) {
	logger := zap.NewNop()
	provider := NewMusicRecognitionProvider(logger)

	// Last.fm search returns results
	searchResp := LastFMSearchResponse{
		Results: LastFMResults{
			TrackMatches: LastFMTrackMatches{
				Track: []LastFMTrack{
					{Name: "Bohemian Rhapsody", Artist: "Queen", MBID: "mb123"},
				},
			},
		},
	}

	trackInfo := LastFMTrackInfo{
		Track: LastFMTrackDetail{
			Name:      "Bohemian Rhapsody",
			Duration:  "354000",
			Listeners: "5000000",
			Playcount: "20000000",
			Artist:    LastFMArtistDetail{Name: "Queen", MBID: "queen_mbid"},
			Album:     LastFMAlbumDetail{Title: "A Night at the Opera", Artist: "Queen"},
			TopTags:   LastFMTopTags{Tag: []LastFMTag{{Name: "rock"}, {Name: "classic rock"}}},
			Wiki:      LastFMWiki{Summary: "A legendary song"},
			URL:       "https://last.fm/track/queen/bohemian-rhapsody",
		},
	}

	callCount := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method := r.URL.Query().Get("method")
		if method == "track.search" || callCount == 0 {
			callCount++
			json.NewEncoder(w).Encode(searchResp)
		} else {
			json.NewEncoder(w).Encode(trackInfo)
		}
	}))
	defer ts.Close()

	provider.baseURLs["lastfm"] = ts.URL

	req := &MediaRecognitionRequest{
		FilePath:  "/music/Queen - Bohemian Rhapsody.mp3",
		FileName:  "Queen - Bohemian Rhapsody.mp3",
		MediaType: MediaTypeMusic,
	}

	result, err := provider.RecognizeMedia(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "Bohemian Rhapsody", result.Title)
	assert.Equal(t, "Queen", result.Artist)
	assert.Equal(t, "Last.fm", result.APIProvider)
}

func TestMusicRecognitionProvider_RecognizeMedia_FallbackToBasic(t *testing.T) {
	logger := zap.NewNop()
	provider := NewMusicRecognitionProvider(logger)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	provider.baseURLs["lastfm"] = ts.URL
	provider.baseURLs["musicbrainz"] = ts.URL
	provider.baseURLs["acoustid"] = ts.URL

	req := &MediaRecognitionRequest{
		FilePath:  "/music/unknown_track.mp3",
		FileName:  "unknown_track.mp3",
		MediaType: MediaTypeMusic,
	}

	result, err := provider.RecognizeMedia(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "filename_parsing", result.RecognitionMethod)
}

// =============================================================================
// searchLastFM — Last.fm API interaction
// =============================================================================

func TestMusicRecognitionProvider_SearchLastFM_Success(t *testing.T) {
	logger := zap.NewNop()
	provider := NewMusicRecognitionProvider(logger)

	searchResp := LastFMSearchResponse{
		Results: LastFMResults{
			TrackMatches: LastFMTrackMatches{
				Track: []LastFMTrack{
					{Name: "Stairway to Heaven", Artist: "Led Zeppelin", MBID: "stairway_mb"},
				},
			},
		},
	}

	trackInfo := LastFMTrackInfo{
		Track: LastFMTrackDetail{
			Name:      "Stairway to Heaven",
			Duration:  "482000",
			Listeners: "3000000",
			Playcount: "10000000",
			Artist:    LastFMArtistDetail{Name: "Led Zeppelin"},
			Album: LastFMAlbumDetail{
				Title: "Led Zeppelin IV",
				Image: []LastFMImage{{Text: "https://example.com/art.jpg", Size: "extralarge"}},
			},
			TopTags: LastFMTopTags{Tag: []LastFMTag{{Name: "rock"}}},
			MBID:    "stairway_mb",
			URL:     "https://last.fm/track",
		},
	}

	callCount := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if callCount == 0 {
			callCount++
			json.NewEncoder(w).Encode(searchResp)
		} else {
			json.NewEncoder(w).Encode(trackInfo)
		}
	}))
	defer ts.Close()

	provider.baseURLs["lastfm"] = ts.URL

	result, err := provider.searchLastFM(context.Background(), "Stairway to Heaven", "Led Zeppelin", "")
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "Stairway to Heaven", result.Title)
	assert.Equal(t, "Led Zeppelin", result.Artist)
	assert.NotEmpty(t, result.CoverArt)
}

func TestMusicRecognitionProvider_SearchLastFM_NoResults(t *testing.T) {
	logger := zap.NewNop()
	provider := NewMusicRecognitionProvider(logger)

	searchResp := LastFMSearchResponse{
		Results: LastFMResults{
			TrackMatches: LastFMTrackMatches{Track: []LastFMTrack{}},
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(searchResp)
	}))
	defer ts.Close()

	provider.baseURLs["lastfm"] = ts.URL

	_, err := provider.searchLastFM(context.Background(), "Nonexistent", "", "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no tracks found")
}

func TestMusicRecognitionProvider_SearchLastFM_HTTPError(t *testing.T) {
	logger := zap.NewNop()
	provider := NewMusicRecognitionProvider(logger)

	provider.baseURLs["lastfm"] = "http://localhost:1"

	_, err := provider.searchLastFM(context.Background(), "Test", "Artist", "")
	assert.Error(t, err)
}

// =============================================================================
// searchMusicBrainz — MusicBrainz API interaction
// =============================================================================

func TestMusicRecognitionProvider_SearchMusicBrainz_Success(t *testing.T) {
	logger := zap.NewNop()
	provider := NewMusicRecognitionProvider(logger)

	mbResp := MusicBrainzSearchResponse{
		Recordings: []MusicBrainzRecording{
			{
				ID:    "mb_rec_001",
				Score: 95,
				Title: "Yesterday",
				ArtistCredit: []MusicBrainzArtistCredit{
					{Artist: MusicBrainzArtist{ID: "beatles_id", Name: "The Beatles"}},
				},
				Releases: []MusicBrainzReleaseBasic{
					{ID: "rel_001", Title: "Help!", Date: "1965"},
				},
				Genres: []MusicBrainzGenre{{Name: "pop"}},
				Tags:   []MusicBrainzTag{{Name: "60s"}},
				ISRCs:  []string{"GBAYE0601234"},
			},
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(mbResp)
	}))
	defer ts.Close()

	provider.baseURLs["musicbrainz"] = ts.URL

	result, err := provider.searchMusicBrainz(context.Background(), "Yesterday", "The Beatles", "Help!")
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "Yesterday", result.Title)
	assert.Equal(t, "The Beatles", result.Artist)
	assert.Equal(t, "Help!", result.Album)
	assert.Equal(t, 1965, result.Year)
	assert.Equal(t, "MusicBrainz", result.APIProvider)
	assert.Contains(t, result.Genres, "pop")
	assert.Contains(t, result.Tags, "60s")
}

func TestMusicRecognitionProvider_SearchMusicBrainz_NoResults(t *testing.T) {
	logger := zap.NewNop()
	provider := NewMusicRecognitionProvider(logger)

	mbResp := MusicBrainzSearchResponse{
		Recordings: []MusicBrainzRecording{},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(mbResp)
	}))
	defer ts.Close()

	provider.baseURLs["musicbrainz"] = ts.URL

	_, err := provider.searchMusicBrainz(context.Background(), "Nonexistent", "", "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no recordings found")
}

// =============================================================================
// getMusicBrainzDetails — detailed recording info
// =============================================================================

func TestMusicRecognitionProvider_GetMusicBrainzDetails_Success(t *testing.T) {
	logger := zap.NewNop()
	provider := NewMusicRecognitionProvider(logger)

	recording := MusicBrainzRecording{
		ID:    "mb_detail_001",
		Title: "Detailed Track",
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(recording)
	}))
	defer ts.Close()

	provider.baseURLs["musicbrainz"] = ts.URL

	result, err := provider.getMusicBrainzDetails(context.Background(), "mb_detail_001")
	require.NoError(t, err)
	assert.Equal(t, "Detailed Track", result.Title)
}

func TestMusicRecognitionProvider_GetMusicBrainzDetails_HTTPError(t *testing.T) {
	logger := zap.NewNop()
	provider := NewMusicRecognitionProvider(logger)

	provider.baseURLs["musicbrainz"] = "http://localhost:1"

	_, err := provider.getMusicBrainzDetails(context.Background(), "nonexistent")
	assert.Error(t, err)
}

// =============================================================================
// convertMusicBrainzRecording — data conversion
// =============================================================================

func TestMusicRecognitionProvider_ConvertMusicBrainzRecording_Full(t *testing.T) {
	logger := zap.NewNop()
	provider := NewMusicRecognitionProvider(logger)

	recording := MusicBrainzRecording{
		ID:    "mb_full",
		Score: 90,
		Title: "Full Track",
		ArtistCredit: []MusicBrainzArtistCredit{
			{Artist: MusicBrainzArtist{Name: "Full Artist"}},
		},
		Releases: []MusicBrainzReleaseBasic{
			{Title: "Full Album", Date: "2020"},
		},
		Genres: []MusicBrainzGenre{{Name: "electronic"}},
		Tags:   []MusicBrainzTag{{Name: "ambient"}},
		ISRCs:  []string{"US1234567890"},
	}

	result := provider.convertMusicBrainzRecording(recording)
	require.NotNil(t, result)
	assert.Equal(t, "Full Track", result.Title)
	assert.Equal(t, "Full Artist", result.Artist)
	assert.Equal(t, "Full Album", result.Album)
	assert.Equal(t, 2020, result.Year)
	assert.Contains(t, result.Genres, "electronic")
	assert.Contains(t, result.Tags, "ambient")
	assert.Equal(t, "US1234567890", result.ExternalIDs["isrc"])
}

func TestMusicRecognitionProvider_ConvertMusicBrainzRecording_Minimal(t *testing.T) {
	logger := zap.NewNop()
	provider := NewMusicRecognitionProvider(logger)

	recording := MusicBrainzRecording{
		ID:    "mb_min",
		Score: 50,
		Title: "Minimal Track",
	}

	result := provider.convertMusicBrainzRecording(recording)
	require.NotNil(t, result)
	assert.Equal(t, "Minimal Track", result.Title)
	assert.Empty(t, result.Artist)
}

// =============================================================================
// enhanceMusicResult — enrichment from MusicBrainz data
// =============================================================================

func TestMusicRecognitionProvider_EnhanceMusicResult(t *testing.T) {
	logger := zap.NewNop()
	provider := NewMusicRecognitionProvider(logger)

	result := &MediaRecognitionResult{
		Title:       "",
		ExternalIDs: map[string]string{},
	}

	mbRecording := &MusicBrainzRecording{
		ID:    "mb_enhance",
		Title: "Enhanced Title",
		Genres: []MusicBrainzGenre{
			{Name: "rock"},
			{Name: "alternative"},
		},
		Tags: []MusicBrainzTag{
			{Name: "90s"},
		},
		ISRCs: []string{"ISRC001"},
		ArtistCredit: []MusicBrainzArtistCredit{
			{Artist: MusicBrainzArtist{Name: "Enhanced Artist"}},
		},
	}

	provider.enhanceMusicResult(result, mbRecording)

	assert.Equal(t, "Enhanced Title", result.Title)
	assert.Contains(t, result.Genres, "rock")
	assert.Contains(t, result.Genres, "alternative")
}

func TestMusicRecognitionProvider_EnhanceMusicResult_PreservesExistingTitle(t *testing.T) {
	logger := zap.NewNop()
	provider := NewMusicRecognitionProvider(logger)

	result := &MediaRecognitionResult{
		Title:       "Original Title",
		ExternalIDs: map[string]string{},
		Genres:      []string{"pop"},
	}

	mbRecording := &MusicBrainzRecording{
		Title: "Different Title",
	}

	provider.enhanceMusicResult(result, mbRecording)

	// Original title should be preserved
	assert.Equal(t, "Original Title", result.Title)
}

// =============================================================================
// recognizeByFingerprint — audio fingerprint recognition
// =============================================================================

func TestMusicRecognitionProvider_RecognizeByFingerprint_Success(t *testing.T) {
	logger := zap.NewNop()
	provider := NewMusicRecognitionProvider(logger)

	acoustIDResp := AcoustIDResponse{
		Status: "ok",
		Results: []AcoustIDResult{
			{
				ID:    "acoust_001",
				Score: 0.95,
				Recordings: []AcoustIDRecording{
					{
						ID:       "mb_rec_fp",
						Title:    "Fingerprinted Song",
						Duration: 240.5,
						Artists:  []AcoustIDArtist{{ID: "artist_fp", Name: "FP Artist"}},
						Releases: []AcoustIDRelease{{ID: "rel_fp", Title: "FP Album", Date: "2022"}},
					},
				},
			},
		},
	}

	mbRecording := MusicBrainzRecording{
		ID:    "mb_rec_fp",
		Title: "Fingerprinted Song",
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			json.NewEncoder(w).Encode(acoustIDResp)
		} else {
			json.NewEncoder(w).Encode(mbRecording)
		}
	}))
	defer ts.Close()

	provider.baseURLs["acoustid"] = ts.URL
	provider.baseURLs["musicbrainz"] = ts.URL

	result, err := provider.recognizeByFingerprint(context.Background(), []byte("fake audio data for fingerprinting"))
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "Fingerprinted Song", result.Title)
	assert.Equal(t, "FP Artist", result.Artist)
	assert.Equal(t, "FP Album", result.Album)
	assert.Equal(t, 0.95, result.Confidence)
	assert.Equal(t, "audio_fingerprint", result.RecognitionMethod)
}

func TestMusicRecognitionProvider_RecognizeByFingerprint_NoResults(t *testing.T) {
	logger := zap.NewNop()
	provider := NewMusicRecognitionProvider(logger)

	acoustIDResp := AcoustIDResponse{
		Status:  "ok",
		Results: []AcoustIDResult{},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(acoustIDResp)
	}))
	defer ts.Close()

	provider.baseURLs["acoustid"] = ts.URL

	_, err := provider.recognizeByFingerprint(context.Background(), []byte("fake audio"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no results from AcoustID")
}

func TestMusicRecognitionProvider_RecognizeByFingerprint_NoRecordings(t *testing.T) {
	logger := zap.NewNop()
	provider := NewMusicRecognitionProvider(logger)

	acoustIDResp := AcoustIDResponse{
		Status: "ok",
		Results: []AcoustIDResult{
			{ID: "no_rec", Score: 0.5, Recordings: []AcoustIDRecording{}},
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(acoustIDResp)
	}))
	defer ts.Close()

	provider.baseURLs["acoustid"] = ts.URL

	_, err := provider.recognizeByFingerprint(context.Background(), []byte("fake audio"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no recordings found")
}

func TestMusicRecognitionProvider_RecognizeByFingerprint_HTTPError(t *testing.T) {
	logger := zap.NewNop()
	provider := NewMusicRecognitionProvider(logger)

	provider.baseURLs["acoustid"] = "http://localhost:1"

	_, err := provider.recognizeByFingerprint(context.Background(), []byte("fake audio"))
	assert.Error(t, err)
}

// =============================================================================
// RecognizeMedia with audio sample (fingerprint path)
// =============================================================================

func TestMusicRecognitionProvider_RecognizeMedia_WithAudioSample(t *testing.T) {
	logger := zap.NewNop()
	provider := NewMusicRecognitionProvider(logger)

	acoustIDResp := AcoustIDResponse{
		Status: "ok",
		Results: []AcoustIDResult{
			{
				ID:    "acoust_full",
				Score: 0.92,
				Recordings: []AcoustIDRecording{
					{
						ID:       "mb_full_fp",
						Title:    "Audio Sample Song",
						Duration: 180.0,
						Artists:  []AcoustIDArtist{{Name: "Sample Artist"}},
						Releases: []AcoustIDRelease{{Title: "Sample Album", Date: "2023"}},
					},
				},
			},
		},
	}

	mbRecording := MusicBrainzRecording{ID: "mb_full_fp", Title: "Audio Sample Song"}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			json.NewEncoder(w).Encode(acoustIDResp)
		} else {
			json.NewEncoder(w).Encode(mbRecording)
		}
	}))
	defer ts.Close()

	provider.baseURLs["acoustid"] = ts.URL
	provider.baseURLs["musicbrainz"] = ts.URL

	req := &MediaRecognitionRequest{
		FilePath:    "/music/song.mp3",
		FileName:    "song.mp3",
		MediaType:   MediaTypeMusic,
		AudioSample: []byte("fake audio sample data for fingerprinting test"),
	}

	result, err := provider.RecognizeMedia(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "Audio Sample Song", result.Title)
	assert.Equal(t, "audio_fingerprint", result.RecognitionMethod)
}

// =============================================================================
// recognizeByMetadata — metadata search pipeline
// =============================================================================

func TestMusicRecognitionProvider_RecognizeByMetadata_LastFMSuccess(t *testing.T) {
	logger := zap.NewNop()
	provider := NewMusicRecognitionProvider(logger)

	searchResp := LastFMSearchResponse{
		Results: LastFMResults{
			TrackMatches: LastFMTrackMatches{
				Track: []LastFMTrack{{Name: "Track", Artist: "Artist"}},
			},
		},
	}

	trackInfo := LastFMTrackInfo{
		Track: LastFMTrackDetail{
			Name:   "Track",
			Artist: LastFMArtistDetail{Name: "Artist"},
			Album:  LastFMAlbumDetail{Title: "Album"},
		},
	}

	callCount := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if callCount == 0 {
			callCount++
			json.NewEncoder(w).Encode(searchResp)
		} else {
			json.NewEncoder(w).Encode(trackInfo)
		}
	}))
	defer ts.Close()

	provider.baseURLs["lastfm"] = ts.URL

	result, err := provider.recognizeByMetadata(context.Background(), "Track", "Artist", "Album")
	require.NoError(t, err)
	assert.Equal(t, "Track", result.Title)
}

func TestMusicRecognitionProvider_RecognizeByMetadata_AllFail(t *testing.T) {
	logger := zap.NewNop()
	provider := NewMusicRecognitionProvider(logger)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	provider.baseURLs["lastfm"] = ts.URL
	provider.baseURLs["musicbrainz"] = ts.URL

	_, err := provider.recognizeByMetadata(context.Background(), "Test", "Artist", "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no results")
}
