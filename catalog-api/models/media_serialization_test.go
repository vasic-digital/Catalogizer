package models

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// ExternalMetadataList Value/Scan Tests
// =============================================================================

func TestExternalMetadataList_Value(t *testing.T) {
	list := ExternalMetadataList{
		{ID: 1, Provider: "tmdb", ExternalID: "123", Title: "Test Movie"},
		{ID: 2, Provider: "imdb", ExternalID: "tt123", Title: "Test Movie"},
	}

	val, err := list.Value()
	require.NoError(t, err)
	require.NotNil(t, val)

	// Should be valid JSON bytes
	data, ok := val.([]byte)
	require.True(t, ok)

	var unmarshaled []ExternalMetadata
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)
	assert.Len(t, unmarshaled, 2)
	assert.Equal(t, "tmdb", unmarshaled[0].Provider)
	assert.Equal(t, "imdb", unmarshaled[1].Provider)
}

func TestExternalMetadataList_Value_Empty(t *testing.T) {
	list := ExternalMetadataList{}

	val, err := list.Value()
	require.NoError(t, err)
	require.NotNil(t, val)

	data, ok := val.([]byte)
	require.True(t, ok)
	assert.Equal(t, "[]", string(data))
}

func TestExternalMetadataList_Scan_Nil(t *testing.T) {
	var list ExternalMetadataList
	err := list.Scan(nil)

	require.NoError(t, err)
	assert.Equal(t, ExternalMetadataList{}, list)
}

func TestExternalMetadataList_Scan_Bytes(t *testing.T) {
	jsonData := []byte(`[{"id":1,"provider":"tmdb","external_id":"123","title":"Test"}]`)

	var list ExternalMetadataList
	err := list.Scan(jsonData)

	require.NoError(t, err)
	assert.Len(t, list, 1)
	assert.Equal(t, "tmdb", list[0].Provider)
	assert.Equal(t, "123", list[0].ExternalID)
}

func TestExternalMetadataList_Scan_String(t *testing.T) {
	jsonStr := `[{"id":1,"provider":"imdb","external_id":"tt456","title":"Movie"}]`

	var list ExternalMetadataList
	err := list.Scan(jsonStr)

	require.NoError(t, err)
	assert.Len(t, list, 1)
	assert.Equal(t, "imdb", list[0].Provider)
	assert.Equal(t, "tt456", list[0].ExternalID)
}

func TestExternalMetadataList_Scan_InvalidJSON(t *testing.T) {
	var list ExternalMetadataList
	err := list.Scan([]byte(`not valid json`))

	assert.Error(t, err)
}

func TestExternalMetadataList_Scan_EmptyArray(t *testing.T) {
	var list ExternalMetadataList
	err := list.Scan([]byte(`[]`))

	require.NoError(t, err)
	assert.Len(t, list, 0)
}

func TestExternalMetadataList_Scan_UnsupportedType(t *testing.T) {
	var list ExternalMetadataList
	err := list.Scan(12345)

	// Should return nil for unsupported types (no error, no panic)
	assert.NoError(t, err)
}

// =============================================================================
// MediaVersionsJSON Value/Scan Tests
// =============================================================================

func TestMediaVersionsJSON_Value(t *testing.T) {
	codec := "h264"
	resolution := "1920x1080"

	versions := MediaVersionsJSON{
		{ID: 1, Quality: "1080p", FileSize: 2048000, Codec: &codec, Resolution: &resolution},
		{ID: 2, Quality: "720p", FileSize: 1024000},
	}

	val, err := versions.Value()
	require.NoError(t, err)
	require.NotNil(t, val)

	data, ok := val.([]byte)
	require.True(t, ok)

	var unmarshaled []MediaVersion
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)
	assert.Len(t, unmarshaled, 2)
	assert.Equal(t, "1080p", unmarshaled[0].Quality)
	assert.Equal(t, "720p", unmarshaled[1].Quality)
}

func TestMediaVersionsJSON_Value_Empty(t *testing.T) {
	versions := MediaVersionsJSON{}

	val, err := versions.Value()
	require.NoError(t, err)
	require.NotNil(t, val)

	data, ok := val.([]byte)
	require.True(t, ok)
	assert.Equal(t, "[]", string(data))
}

func TestMediaVersionsJSON_Scan_Nil(t *testing.T) {
	var versions MediaVersionsJSON
	err := versions.Scan(nil)

	require.NoError(t, err)
	assert.Equal(t, MediaVersionsJSON{}, versions)
}

func TestMediaVersionsJSON_Scan_Bytes(t *testing.T) {
	jsonData := []byte(`[{"id":1,"quality":"4K","file_size":10485760}]`)

	var versions MediaVersionsJSON
	err := versions.Scan(jsonData)

	require.NoError(t, err)
	assert.Len(t, versions, 1)
	assert.Equal(t, "4K", versions[0].Quality)
	assert.Equal(t, int64(10485760), versions[0].FileSize)
}

func TestMediaVersionsJSON_Scan_String(t *testing.T) {
	jsonStr := `[{"id":2,"quality":"720p","file_size":1024000}]`

	var versions MediaVersionsJSON
	err := versions.Scan(jsonStr)

	require.NoError(t, err)
	assert.Len(t, versions, 1)
	assert.Equal(t, "720p", versions[0].Quality)
}

func TestMediaVersionsJSON_Scan_InvalidJSON(t *testing.T) {
	var versions MediaVersionsJSON
	err := versions.Scan([]byte(`{invalid}`))

	assert.Error(t, err)
}

func TestMediaVersionsJSON_Scan_EmptyArray(t *testing.T) {
	var versions MediaVersionsJSON
	err := versions.Scan([]byte(`[]`))

	require.NoError(t, err)
	assert.Len(t, versions, 0)
}

func TestMediaVersionsJSON_Scan_UnsupportedType(t *testing.T) {
	var versions MediaVersionsJSON
	err := versions.Scan(12345)

	assert.NoError(t, err)
}

// =============================================================================
// Roundtrip Tests (Value -> Scan)
// =============================================================================

func TestExternalMetadataList_Roundtrip(t *testing.T) {
	desc := "A great movie"
	year := 2024
	rating := 8.5

	original := ExternalMetadataList{
		{
			ID:          1,
			MediaID:     10,
			Provider:    "tmdb",
			ExternalID:  "550",
			Title:       "Fight Club",
			Description: &desc,
			Year:        &year,
			Rating:      &rating,
			Genres:      []string{"Drama", "Thriller"},
			Cast:        []string{"Brad Pitt", "Edward Norton"},
			Metadata:    map[string]string{"runtime": "139"},
			LastUpdated: "2024-01-01T00:00:00Z",
		},
	}

	val, err := original.Value()
	require.NoError(t, err)

	var restored ExternalMetadataList
	err = restored.Scan(val)
	require.NoError(t, err)

	assert.Len(t, restored, 1)
	assert.Equal(t, original[0].Provider, restored[0].Provider)
	assert.Equal(t, original[0].Title, restored[0].Title)
	assert.Equal(t, *original[0].Year, *restored[0].Year)
	assert.Equal(t, original[0].Genres, restored[0].Genres)
}

func TestMediaVersionsJSON_Roundtrip(t *testing.T) {
	codec := "hevc"
	resolution := "3840x2160"
	bitrate := int64(15000000)

	original := MediaVersionsJSON{
		{
			ID:         1,
			MediaID:    10,
			Version:    "uhd",
			Quality:    "2160p",
			FilePath:   "/movies/test_uhd.mkv",
			FileSize:   20971520000,
			Codec:      &codec,
			Resolution: &resolution,
			Bitrate:    &bitrate,
		},
	}

	val, err := original.Value()
	require.NoError(t, err)

	var restored MediaVersionsJSON
	err = restored.Scan(val)
	require.NoError(t, err)

	assert.Len(t, restored, 1)
	assert.Equal(t, original[0].Quality, restored[0].Quality)
	assert.Equal(t, original[0].FileSize, restored[0].FileSize)
	assert.Equal(t, *original[0].Codec, *restored[0].Codec)
	assert.Equal(t, *original[0].Bitrate, *restored[0].Bitrate)
}

// =============================================================================
// MediaStats Tests
// =============================================================================

func TestMediaStats_JSONMarshaling(t *testing.T) {
	stats := MediaStats{
		TotalItems: 1500,
		ByType: map[string]int64{
			"movie":   500,
			"tv_show": 300,
			"music":   700,
		},
		ByQuality: map[string]int64{
			"1080p": 800,
			"720p":  500,
			"4K":    200,
		},
		TotalSize:       1099511627776,
		RecentAdditions: 42,
	}

	data, err := json.Marshal(stats)
	require.NoError(t, err)

	var restored MediaStats
	err = json.Unmarshal(data, &restored)
	require.NoError(t, err)

	assert.Equal(t, stats.TotalItems, restored.TotalItems)
	assert.Equal(t, stats.TotalSize, restored.TotalSize)
	assert.Equal(t, stats.RecentAdditions, restored.RecentAdditions)
	assert.Len(t, restored.ByType, 3)
	assert.Equal(t, int64(500), restored.ByType["movie"])
}
