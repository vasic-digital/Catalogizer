package services

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// =============================================================================
// RecognizeMedia — top-level movie/TV recognition
// =============================================================================

func TestMovieRecognitionProvider_RecognizeMedia_TMDbMovie(t *testing.T) {
	logger := zap.NewNop()
	provider := NewMovieRecognitionProvider(logger)

	// TMDb search returns a movie result
	searchResp := TMDbSearchResponse{
		Results: []TMDbResult{
			{ID: 550, Title: "Fight Club", MediaType: "movie", VoteAverage: 8.4, VoteCount: 20000},
		},
	}
	movieDetails := TMDbMovieDetails{
		ID:          550,
		Title:       "Fight Club",
		Overview:    "An insomniac office worker...",
		ReleaseDate: "1999-10-15",
		Runtime:     139,
		VoteAverage: 8.4,
		VoteCount:   20000,
		PosterPath:  "/poster.jpg",
		IMDbID:      "tt0137523",
		Genres:      []TMDbGenre{{Name: "Drama"}, {Name: "Thriller"}},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/search/") {
			json.NewEncoder(w).Encode(searchResp)
		} else if strings.Contains(r.URL.Path, "/movie/") {
			json.NewEncoder(w).Encode(movieDetails)
		}
	}))
	defer ts.Close()

	provider.baseURLs["tmdb"] = ts.URL

	req := &MediaRecognitionRequest{
		FilePath:  "/movies/Fight.Club.1999.1080p.BluRay.mkv",
		FileName:  "Fight.Club.1999.1080p.BluRay.mkv",
		MediaType: MediaTypeMovie,
	}

	result, err := provider.RecognizeMedia(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "Fight Club", result.Title)
	assert.Equal(t, "TMDb", result.APIProvider)
	assert.Equal(t, 1999, result.Year)
	assert.Contains(t, result.Genres, "Drama")
	assert.Equal(t, "tt0137523", result.IMDbID)
	assert.NotEmpty(t, result.CoverArt)
}

func TestMovieRecognitionProvider_RecognizeMedia_TMDbTV(t *testing.T) {
	logger := zap.NewNop()
	provider := NewMovieRecognitionProvider(logger)

	searchResp := TMDbSearchResponse{
		Results: []TMDbResult{
			{ID: 1399, Name: "Breaking Bad", MediaType: "tv", VoteAverage: 8.9, VoteCount: 10000},
		},
	}
	tvDetails := TMDbTVDetails{
		ID:           1399,
		Name:         "Breaking Bad",
		Overview:     "A high school chemistry teacher...",
		FirstAirDate: "2008-01-20",
		VoteAverage:  8.9,
		VoteCount:    10000,
		PosterPath:   "/bb_poster.jpg",
		Genres:       []TMDbGenre{{Name: "Drama"}, {Name: "Crime"}},
		ExternalIDs:  TMDbExternalIDs{IMDbID: "tt0903747"},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/search/") {
			json.NewEncoder(w).Encode(searchResp)
		} else if strings.Contains(r.URL.Path, "/tv/") {
			json.NewEncoder(w).Encode(tvDetails)
		}
	}))
	defer ts.Close()

	provider.baseURLs["tmdb"] = ts.URL

	req := &MediaRecognitionRequest{
		FilePath:  "/tv/Breaking.Bad.S01E01.720p.mkv",
		FileName:  "Breaking.Bad.S01E01.720p.mkv",
		MediaType: MediaTypeTVSeries,
	}

	result, err := provider.RecognizeMedia(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "TMDb", result.APIProvider)
}

func TestMovieRecognitionProvider_RecognizeMedia_FallbackToOMDb(t *testing.T) {
	logger := zap.NewNop()
	provider := NewMovieRecognitionProvider(logger)

	// TMDb returns no results
	tsTMDb := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(TMDbSearchResponse{Results: []TMDbResult{}})
	}))
	defer tsTMDb.Close()

	omdbResp := OMDbResponse{
		Title:      "The Matrix",
		Year:       "1999",
		Runtime:    "136 min",
		Genre:      "Action, Sci-Fi",
		Director:   "Lana Wachowski, Lilly Wachowski",
		Actors:     "Keanu Reeves, Laurence Fishburne",
		Plot:       "A computer hacker learns...",
		IMDbRating: "8.7",
		IMDbVotes:  "1,500,000",
		IMDbID:     "tt0133093",
		Type:       "movie",
		Poster:     "https://example.com/matrix.jpg",
		Response:   "True",
	}

	tsOMDb := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(omdbResp)
	}))
	defer tsOMDb.Close()

	provider.baseURLs["tmdb"] = tsTMDb.URL
	provider.baseURLs["omdb"] = tsOMDb.URL

	req := &MediaRecognitionRequest{
		FilePath:  "/movies/The.Matrix.1999.mkv",
		FileName:  "The.Matrix.1999.mkv",
		MediaType: MediaTypeMovie,
	}

	result, err := provider.RecognizeMedia(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "The Matrix", result.Title)
	assert.Equal(t, "OMDb", result.APIProvider)
	assert.Equal(t, "tt0133093", result.IMDbID)
	assert.Contains(t, result.Genres, "Action")
}

func TestMovieRecognitionProvider_RecognizeMedia_FallbackToBasic(t *testing.T) {
	logger := zap.NewNop()
	provider := NewMovieRecognitionProvider(logger)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	provider.baseURLs["tmdb"] = ts.URL
	provider.baseURLs["omdb"] = ts.URL

	req := &MediaRecognitionRequest{
		FilePath:  "/movies/Unknown.Movie.2024.mkv",
		FileName:  "Unknown.Movie.2024.mkv",
		MediaType: MediaTypeMovie,
	}

	result, err := provider.RecognizeMedia(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "filename_parsing", result.RecognitionMethod)
}

// =============================================================================
// searchTMDb — TMDb API interaction
// =============================================================================

func TestMovieRecognitionProvider_SearchTMDb_NoResults(t *testing.T) {
	logger := zap.NewNop()
	provider := NewMovieRecognitionProvider(logger)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(TMDbSearchResponse{Results: []TMDbResult{}})
	}))
	defer ts.Close()

	provider.baseURLs["tmdb"] = ts.URL

	_, err := provider.searchTMDb(context.Background(), "Nonexistent", 0, MediaTypeMovie, 0, 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no results found")
}

func TestMovieRecognitionProvider_SearchTMDb_HTTPError(t *testing.T) {
	logger := zap.NewNop()
	provider := NewMovieRecognitionProvider(logger)

	provider.baseURLs["tmdb"] = "http://localhost:1"

	_, err := provider.searchTMDb(context.Background(), "Test", 2024, MediaTypeMovie, 0, 0)
	assert.Error(t, err)
}

func TestMovieRecognitionProvider_SearchTMDb_UnsupportedMediaType(t *testing.T) {
	logger := zap.NewNop()
	provider := NewMovieRecognitionProvider(logger)

	// Return a result with no title and no name (neither movie nor TV)
	searchResp := TMDbSearchResponse{
		Results: []TMDbResult{
			{ID: 1, MediaType: "person", VoteAverage: 5.0},
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(searchResp)
	}))
	defer ts.Close()

	provider.baseURLs["tmdb"] = ts.URL

	_, err := provider.searchTMDb(context.Background(), "Test", 0, MediaTypeMovie, 0, 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported media type")
}

// =============================================================================
// searchOMDb — OMDb API interaction
// =============================================================================

func TestMovieRecognitionProvider_SearchOMDb_Success(t *testing.T) {
	logger := zap.NewNop()
	provider := NewMovieRecognitionProvider(logger)

	omdbResp := OMDbResponse{
		Title:      "Inception",
		Year:       "2010",
		Released:   "16 Jul 2010",
		Runtime:    "148 min",
		Genre:      "Action, Adventure, Sci-Fi",
		Director:   "Christopher Nolan",
		Actors:     "Leonardo DiCaprio, Joseph Gordon-Levitt",
		Plot:       "A thief who steals corporate secrets...",
		IMDbRating: "8.8",
		IMDbVotes:  "2,000,000",
		IMDbID:     "tt1375666",
		Type:       "movie",
		Poster:     "https://example.com/inception.jpg",
		Response:   "True",
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(omdbResp)
	}))
	defer ts.Close()

	provider.baseURLs["omdb"] = ts.URL

	result, err := provider.searchOMDb(context.Background(), "Inception", 2010, MediaTypeMovie)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "Inception", result.Title)
	assert.Equal(t, "OMDb", result.APIProvider)
	assert.Equal(t, "tt1375666", result.IMDbID)
	assert.Equal(t, "Christopher Nolan", result.Director)
	assert.NotEmpty(t, result.CoverArt)
	assert.Greater(t, len(result.Cast), 0)
	assert.Greater(t, result.Duration, int64(0))
}

func TestMovieRecognitionProvider_SearchOMDb_NotFound(t *testing.T) {
	logger := zap.NewNop()
	provider := NewMovieRecognitionProvider(logger)

	omdbResp := OMDbResponse{
		Response: "False",
		Error:    "Movie not found!",
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(omdbResp)
	}))
	defer ts.Close()

	provider.baseURLs["omdb"] = ts.URL

	_, err := provider.searchOMDb(context.Background(), "Nonexistent", 0, MediaTypeMovie)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Movie not found")
}

func TestMovieRecognitionProvider_SearchOMDb_TVSeries(t *testing.T) {
	logger := zap.NewNop()
	provider := NewMovieRecognitionProvider(logger)

	omdbResp := OMDbResponse{
		Title:      "Breaking Bad",
		Year:       "2008-2013",
		Genre:      "Crime, Drama, Thriller",
		IMDbRating: "9.5",
		IMDbID:     "tt0903747",
		Type:       "series",
		Response:   "True",
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify type=series is passed
		assert.Equal(t, "series", r.URL.Query().Get("type"))
		json.NewEncoder(w).Encode(omdbResp)
	}))
	defer ts.Close()

	provider.baseURLs["omdb"] = ts.URL

	result, err := provider.searchOMDb(context.Background(), "Breaking Bad", 0, MediaTypeTVSeries)
	require.NoError(t, err)
	assert.Equal(t, MediaTypeTVSeries, result.MediaType)
}

func TestMovieRecognitionProvider_SearchOMDb_HTTPError(t *testing.T) {
	logger := zap.NewNop()
	provider := NewMovieRecognitionProvider(logger)

	provider.baseURLs["omdb"] = "http://localhost:1"

	_, err := provider.searchOMDb(context.Background(), "Test", 0, MediaTypeMovie)
	assert.Error(t, err)
}

// =============================================================================
// getTMDbMovieDetails — detailed movie info
// =============================================================================

func TestMovieRecognitionProvider_GetTMDbMovieDetails_Success(t *testing.T) {
	logger := zap.NewNop()
	provider := NewMovieRecognitionProvider(logger)

	movieDetails := TMDbMovieDetails{
		ID:            550,
		Title:         "Fight Club",
		OriginalTitle: "Fight Club",
		Overview:      "An insomniac office worker...",
		ReleaseDate:   "1999-10-15",
		Runtime:       139,
		VoteAverage:   8.4,
		VoteCount:     20000,
		PosterPath:    "/poster.jpg",
		BackdropPath:  "/backdrop.jpg",
		IMDbID:        "tt0137523",
		Genres:        []TMDbGenre{{Name: "Drama"}},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(movieDetails)
	}))
	defer ts.Close()

	provider.baseURLs["tmdb"] = ts.URL

	result, err := provider.getTMDbMovieDetails(context.Background(), 550)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "Fight Club", result.Title)
	assert.Equal(t, MediaTypeMovie, result.MediaType)
	assert.Equal(t, 1999, result.Year)
	assert.NotNil(t, result.ReleaseDate)
	assert.Len(t, result.CoverArt, 3) // poster medium + poster original + backdrop
}

func TestMovieRecognitionProvider_GetTMDbMovieDetails_HTTPError(t *testing.T) {
	logger := zap.NewNop()
	provider := NewMovieRecognitionProvider(logger)

	provider.baseURLs["tmdb"] = "http://localhost:1"

	_, err := provider.getTMDbMovieDetails(context.Background(), 1)
	assert.Error(t, err)
}

// =============================================================================
// getTMDbTVDetails — detailed TV info
// =============================================================================

func TestMovieRecognitionProvider_GetTMDbTVDetails_Success(t *testing.T) {
	logger := zap.NewNop()
	provider := NewMovieRecognitionProvider(logger)

	tvDetails := TMDbTVDetails{
		ID:           1399,
		Name:         "Breaking Bad",
		OriginalName: "Breaking Bad",
		Overview:     "A high school teacher...",
		FirstAirDate: "2008-01-20",
		VoteAverage:  8.9,
		VoteCount:    10000,
		PosterPath:   "/bb.jpg",
		Genres:       []TMDbGenre{{Name: "Drama"}},
		ExternalIDs:  TMDbExternalIDs{IMDbID: "tt0903747", TVDBID: 81189},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(tvDetails)
	}))
	defer ts.Close()

	provider.baseURLs["tmdb"] = ts.URL

	result, err := provider.getTMDbTVDetails(context.Background(), 1399, 0, 0)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "Breaking Bad", result.Title)
	assert.Equal(t, MediaTypeTVSeries, result.MediaType)
	assert.Equal(t, "tt0903747", result.IMDbID)
}

func TestMovieRecognitionProvider_GetTMDbTVDetails_WithEpisode(t *testing.T) {
	logger := zap.NewNop()
	provider := NewMovieRecognitionProvider(logger)

	tvDetails := TMDbTVDetails{
		ID:           1399,
		Name:         "Breaking Bad",
		Overview:     "A teacher...",
		FirstAirDate: "2008-01-20",
		VoteAverage:  8.9,
		VoteCount:    10000,
		Genres:       []TMDbGenre{{Name: "Drama"}},
	}

	episodeDetails := TMDbEpisodeDetails{
		ID:            62085,
		Name:          "Pilot",
		SeasonNumber:  1,
		EpisodeNumber: 1,
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/season/") {
			json.NewEncoder(w).Encode(episodeDetails)
		} else {
			json.NewEncoder(w).Encode(tvDetails)
		}
	}))
	defer ts.Close()

	provider.baseURLs["tmdb"] = ts.URL

	result, err := provider.getTMDbTVDetails(context.Background(), 1399, 1, 1)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, MediaTypeTVEpisode, result.MediaType)
	assert.Equal(t, "Pilot", result.Title)
	assert.Equal(t, 1, result.Season)
	assert.Equal(t, 1, result.Episode)
}

// =============================================================================
// getTMDbEpisodeDetails
// =============================================================================

func TestMovieRecognitionProvider_GetTMDbEpisodeDetails_Success(t *testing.T) {
	logger := zap.NewNop()
	provider := NewMovieRecognitionProvider(logger)

	epDetails := TMDbEpisodeDetails{
		ID:            62085,
		Name:          "Pilot",
		Overview:      "First episode",
		SeasonNumber:  1,
		EpisodeNumber: 1,
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(epDetails)
	}))
	defer ts.Close()

	provider.baseURLs["tmdb"] = ts.URL

	result, err := provider.getTMDbEpisodeDetails(context.Background(), 1399, 1, 1)
	require.NoError(t, err)
	assert.Equal(t, "Pilot", result.Name)
}

func TestMovieRecognitionProvider_GetTMDbEpisodeDetails_HTTPError(t *testing.T) {
	logger := zap.NewNop()
	provider := NewMovieRecognitionProvider(logger)

	provider.baseURLs["tmdb"] = "http://localhost:1"

	_, err := provider.getTMDbEpisodeDetails(context.Background(), 1, 1, 1)
	assert.Error(t, err)
}
