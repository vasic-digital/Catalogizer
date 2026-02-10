package models

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// MediaCatalogItem Model Tests
// =============================================================================

func TestMediaCatalogItem_JSONMarshaling(t *testing.T) {
	year := 2024
	desc := "Test description"
	cover := "https://example.com/cover.jpg"
	rating := 8.5
	quality := "1080p"
	fileSize := int64(1024000)
	duration := int64(7200)
	smbPath := "smb://server/share/movie.mkv"
	lastWatched := "2024-01-15T10:00:00Z"

	item := MediaCatalogItem{
		ID:            1,
		Title:         "Test Movie",
		MediaType:     "movie",
		Year:          &year,
		Description:   &desc,
		CoverImage:    &cover,
		Rating:        &rating,
		Quality:       &quality,
		FileSize:      &fileSize,
		Duration:      &duration,
		DirectoryPath: "/movies/test",
		SMBPath:       &smbPath,
		CreatedAt:     "2024-01-01T00:00:00Z",
		UpdatedAt:     "2024-01-01T00:00:00Z",
		IsFavorite:    true,
		WatchProgress: 0.75,
		LastWatched:   &lastWatched,
		IsDownloaded:  false,
	}

	data, err := json.Marshal(item)
	require.NoError(t, err)
	assert.Contains(t, string(data), "Test Movie")

	var unmarshaled MediaCatalogItem
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)
	assert.Equal(t, item.ID, unmarshaled.ID)
	assert.Equal(t, item.Title, unmarshaled.Title)
	assert.Equal(t, item.MediaType, unmarshaled.MediaType)
}

func TestMediaCatalogItem_MinimalFields(t *testing.T) {
	item := MediaCatalogItem{
		ID:            1,
		Title:         "Minimal",
		MediaType:     "video",
		DirectoryPath: "/test",
		CreatedAt:     "2024-01-01T00:00:00Z",
		UpdatedAt:     "2024-01-01T00:00:00Z",
		WatchProgress: 0.0,
		IsFavorite:    false,
		IsDownloaded:  false,
	}

	assert.Equal(t, "Minimal", item.Title)
	assert.Nil(t, item.Year)
	assert.Nil(t, item.Description)
	assert.Equal(t, 0.0, item.WatchProgress)
}

func TestMediaCatalogItem_WithMetadata(t *testing.T) {
	item := MediaCatalogItem{
		ID:            1,
		Title:         "Movie with Metadata",
		MediaType:     "movie",
		DirectoryPath: "/movies",
		CreatedAt:     "2024-01-01T00:00:00Z",
		UpdatedAt:     "2024-01-01T00:00:00Z",
		ExternalMetadata: []ExternalMetadata{
			{ID: 1, Provider: "tmdb", ExternalID: "12345"},
			{ID: 2, Provider: "imdb", ExternalID: "tt12345"},
		},
	}

	assert.Len(t, item.ExternalMetadata, 2)
	assert.Equal(t, "tmdb", item.ExternalMetadata[0].Provider)
}

func TestMediaCatalogItem_WithVersions(t *testing.T) {
	item := MediaCatalogItem{
		ID:            1,
		Title:         "Movie with Versions",
		MediaType:     "movie",
		DirectoryPath: "/movies",
		CreatedAt:     "2024-01-01T00:00:00Z",
		UpdatedAt:     "2024-01-01T00:00:00Z",
		Versions: []MediaVersion{
			{ID: 1, Quality: "1080p", FileSize: 2048000},
			{ID: 2, Quality: "720p", FileSize: 1024000},
		},
	}

	assert.Len(t, item.Versions, 2)
	assert.Equal(t, "1080p", item.Versions[0].Quality)
}

func TestMediaCatalogItem_FavoriteAndProgress(t *testing.T) {
	lastWatched := "2024-01-15T10:00:00Z"

	item := MediaCatalogItem{
		ID:            1,
		Title:         "In Progress",
		MediaType:     "movie",
		DirectoryPath: "/movies",
		CreatedAt:     "2024-01-01T00:00:00Z",
		UpdatedAt:     "2024-01-01T00:00:00Z",
		IsFavorite:    true,
		WatchProgress: 0.65,
		LastWatched:   &lastWatched,
		IsDownloaded:  true,
	}

	assert.True(t, item.IsFavorite)
	assert.Equal(t, 0.65, item.WatchProgress)
	assert.NotNil(t, item.LastWatched)
	assert.True(t, item.IsDownloaded)
}

// =============================================================================
// ExternalMetadata Model Tests
// =============================================================================

func TestExternalMetadata_JSONMarshaling(t *testing.T) {
	desc := "External description"
	year := 2024
	rating := 8.5
	poster := "https://image.tmdb.org/poster.jpg"
	backdrop := "https://image.tmdb.org/backdrop.jpg"

	metadata := ExternalMetadata{
		ID:          1,
		MediaID:     10,
		Provider:    "tmdb",
		ExternalID:  "12345",
		Title:       "External Title",
		Description: &desc,
		Year:        &year,
		Rating:      &rating,
		PosterURL:   &poster,
		BackdropURL: &backdrop,
		Genres:      []string{"Action", "Adventure"},
		Cast:        []string{"Actor 1", "Actor 2"},
		Crew:        []string{"Director 1"},
		Metadata:    map[string]string{"key": "value"},
		LastUpdated: "2024-01-01T00:00:00Z",
	}

	data, err := json.Marshal(metadata)
	require.NoError(t, err)
	assert.Contains(t, string(data), "tmdb")

	var unmarshaled ExternalMetadata
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)
	assert.Equal(t, metadata.Provider, unmarshaled.Provider)
	assert.Equal(t, metadata.ExternalID, unmarshaled.ExternalID)
}

func TestExternalMetadata_DifferentProviders(t *testing.T) {
	providers := []struct {
		name       string
		provider   string
		externalID string
	}{
		{"TMDB", "tmdb", "12345"},
		{"IMDB", "imdb", "tt12345"},
		{"OMDB", "omdb", "tt12345"},
		{"TVDB", "tvdb", "98765"},
	}

	for _, p := range providers {
		t.Run(p.name, func(t *testing.T) {
			metadata := ExternalMetadata{
				ID:          1,
				MediaID:     10,
				Provider:    p.provider,
				ExternalID:  p.externalID,
				Title:       "Test",
				LastUpdated: "2024-01-01T00:00:00Z",
			}

			assert.Equal(t, p.provider, metadata.Provider)
			assert.Equal(t, p.externalID, metadata.ExternalID)
		})
	}
}

func TestExternalMetadata_WithArrays(t *testing.T) {
	metadata := ExternalMetadata{
		ID:          1,
		MediaID:     10,
		Provider:    "tmdb",
		ExternalID:  "12345",
		Title:       "Movie with Arrays",
		Genres:      []string{"Action", "Thriller", "Sci-Fi"},
		Cast:        []string{"Lead Actor", "Supporting Actor", "Villain"},
		Crew:        []string{"Director", "Writer", "Producer"},
		LastUpdated: "2024-01-01T00:00:00Z",
	}

	assert.Len(t, metadata.Genres, 3)
	assert.Len(t, metadata.Cast, 3)
	assert.Len(t, metadata.Crew, 3)
	assert.Contains(t, metadata.Genres, "Action")
}

func TestExternalMetadata_WithMetadataMap(t *testing.T) {
	metadata := ExternalMetadata{
		ID:         1,
		MediaID:    10,
		Provider:   "tmdb",
		ExternalID: "12345",
		Title:      "Movie with Metadata",
		Metadata: map[string]string{
			"runtime":       "120",
			"budget":        "100000000",
			"revenue":       "500000000",
			"original_lang": "en",
		},
		LastUpdated: "2024-01-01T00:00:00Z",
	}

	assert.Len(t, metadata.Metadata, 4)
	assert.Equal(t, "120", metadata.Metadata["runtime"])
	assert.Equal(t, "en", metadata.Metadata["original_lang"])
}

func TestExternalMetadata_MinimalFields(t *testing.T) {
	metadata := ExternalMetadata{
		ID:          1,
		MediaID:     10,
		Provider:    "tmdb",
		ExternalID:  "12345",
		Title:       "Minimal Metadata",
		LastUpdated: "2024-01-01T00:00:00Z",
	}

	assert.Nil(t, metadata.Description)
	assert.Nil(t, metadata.Year)
	assert.Nil(t, metadata.Rating)
	assert.Nil(t, metadata.PosterURL)
	assert.Nil(t, metadata.BackdropURL)
}

// =============================================================================
// MediaVersion Model Tests
// =============================================================================

func TestMediaVersion_BasicFields(t *testing.T) {
	version := MediaVersion{
		ID:       1,
		Quality:  "1080p",
		FileSize: 2048000,
	}

	assert.Equal(t, int64(1), version.ID)
	assert.Equal(t, "1080p", version.Quality)
	assert.Equal(t, int64(2048000), version.FileSize)
}

func TestMediaVersion_DifferentQualities(t *testing.T) {
	qualities := []struct {
		name    string
		quality string
		size    int64
	}{
		{"4K", "2160p", 10485760000},
		{"Full HD", "1080p", 4194304000},
		{"HD", "720p", 2097152000},
		{"SD", "480p", 1048576000},
	}

	for _, q := range qualities {
		t.Run(q.name, func(t *testing.T) {
			version := MediaVersion{
				ID:       1,
				Quality:  q.quality,
				FileSize: q.size,
			}

			assert.Equal(t, q.quality, version.Quality)
			assert.Equal(t, q.size, version.FileSize)
		})
	}
}

func TestMediaVersion_JSONMarshaling(t *testing.T) {
	version := MediaVersion{
		ID:       1,
		Quality:  "1080p",
		FileSize: 2048000,
	}

	data, err := json.Marshal(version)
	require.NoError(t, err)
	assert.Contains(t, string(data), "1080p")

	var unmarshaled MediaVersion
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)
	assert.Equal(t, version.Quality, unmarshaled.Quality)
	assert.Equal(t, version.FileSize, unmarshaled.FileSize)
}

// =============================================================================
// Media Type Tests
// =============================================================================

func TestMediaCatalogItem_MediaTypes(t *testing.T) {
	mediaTypes := []string{"movie", "tv_show", "episode", "music", "podcast", "audiobook"}

	for _, mediaType := range mediaTypes {
		item := MediaCatalogItem{
			ID:            1,
			Title:         "Test " + mediaType,
			MediaType:     mediaType,
			DirectoryPath: "/test",
			CreatedAt:     "2024-01-01T00:00:00Z",
			UpdatedAt:     "2024-01-01T00:00:00Z",
		}

		assert.Equal(t, mediaType, item.MediaType)
	}
}

func TestMediaCatalogItem_QualityOptions(t *testing.T) {
	qualities := []string{"2160p", "1080p", "720p", "480p", "360p"}

	for _, quality := range qualities {
		q := quality // capture loop variable
		item := MediaCatalogItem{
			ID:            1,
			Title:         "Test",
			MediaType:     "movie",
			Quality:       &q,
			DirectoryPath: "/test",
			CreatedAt:     "2024-01-01T00:00:00Z",
			UpdatedAt:     "2024-01-01T00:00:00Z",
		}

		assert.NotNil(t, item.Quality)
		assert.Equal(t, quality, *item.Quality)
	}
}

func TestMediaCatalogItem_RatingRange(t *testing.T) {
	ratings := []float64{0.0, 5.5, 7.5, 8.5, 10.0}

	for _, rating := range ratings {
		r := rating // capture loop variable
		item := MediaCatalogItem{
			ID:            1,
			Title:         "Test",
			MediaType:     "movie",
			Rating:        &r,
			DirectoryPath: "/test",
			CreatedAt:     "2024-01-01T00:00:00Z",
			UpdatedAt:     "2024-01-01T00:00:00Z",
		}

		assert.NotNil(t, item.Rating)
		assert.Equal(t, rating, *item.Rating)
	}
}

func TestMediaCatalogItem_WatchProgressRange(t *testing.T) {
	progressValues := []float64{0.0, 0.25, 0.5, 0.75, 1.0}

	for _, progress := range progressValues {
		item := MediaCatalogItem{
			ID:            1,
			Title:         "Test",
			MediaType:     "movie",
			DirectoryPath: "/test",
			CreatedAt:     "2024-01-01T00:00:00Z",
			UpdatedAt:     "2024-01-01T00:00:00Z",
			WatchProgress: progress,
		}

		assert.Equal(t, progress, item.WatchProgress)
		assert.GreaterOrEqual(t, item.WatchProgress, 0.0)
		assert.LessOrEqual(t, item.WatchProgress, 1.0)
	}
}
