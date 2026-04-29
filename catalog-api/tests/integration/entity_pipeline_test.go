package integration

import (
	"fmt"
	"testing"
	"time"

	"catalogizer/internal/tests"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// INTEGRATION TEST: Entity creation pipeline
// =============================================================================

// setupEntityTestData inserts a storage root and file records into the test DB
// so that entity pipeline tests can verify aggregation and hierarchy building.
func setupEntityTestData(t *testing.T, db interface {
	Exec(string, ...interface{}) (interface {
		RowsAffected() (int64, error)
		LastInsertId() (int64, error)
	}, error)
}) {
	t.Helper()
}

func TestEntityPipeline_FileInsertionAndLookup(t *testing.T) {
	db := tests.SetupTestDB(t)
	defer db.Close()

	// Create storage root
	_, err := db.Exec(
		`INSERT INTO storage_roots (name, protocol, path, enabled)
		 VALUES ('test-media', 'local', '/media/test', 1)`,
	)
	require.NoError(t, err)

	var rootID int64
	err = db.QueryRow(
		`SELECT id FROM storage_roots WHERE name = 'test-media'`,
	).Scan(&rootID)
	require.NoError(t, err)

	t.Run("InsertFileRecords", func(t *testing.T) {
		files := []struct {
			path string
			name string
			size int64
		}{
			{"/media/test/movies/inception.mkv", "inception.mkv", 4500000000},
			{"/media/test/movies/matrix.mkv", "matrix.mkv", 3200000000},
			{"/media/test/music/album/track01.flac", "track01.flac", 45000000},
			{"/media/test/tv/show_s01e01.mkv", "show_s01e01.mkv", 1200000000},
			{"/media/test/tv/show_s01e02.mkv", "show_s01e02.mkv", 1100000000},
		}

		for _, f := range files {
			_, err := db.Exec(
				`INSERT INTO files (storage_root_id, path, name, size, modified_at)
				 VALUES (?, ?, ?, ?, datetime('now'))`,
				rootID, f.path, f.name, f.size,
			)
			require.NoError(t, err)
		}

		var count int
		err := db.QueryRow(
			`SELECT COUNT(*) FROM files WHERE storage_root_id = ?`, rootID,
		).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 5, count, "All file records should be inserted")
	})

	t.Run("QueryFilesByStorageRoot", func(t *testing.T) {
		rows, err := db.Query(
			`SELECT name, size FROM files
			 WHERE storage_root_id = ? ORDER BY name`,
			rootID,
		)
		require.NoError(t, err)
		defer rows.Close()

		var names []string
		for rows.Next() {
			var name string
			var size int64
			require.NoError(t, rows.Scan(&name, &size))
			names = append(names, name)
			assert.Greater(t, size, int64(0),
				"File size should be positive")
		}
		require.NoError(t, rows.Err())
		assert.Equal(t, 5, len(names),
			"Should query all files for the storage root")
	})
}

func TestEntityPipeline_MediaItemCreation(t *testing.T) {
	db := tests.SetupTestDB(t)
	defer db.Close()

	// Create a storage root for the test
	_, err := db.Exec(
		`INSERT INTO storage_roots (name, protocol, path, enabled)
		 VALUES ('entity-test', 'local', '/media', 1)`,
	)
	require.NoError(t, err)

	var rootID int64
	err = db.QueryRow(
		`SELECT id FROM storage_roots WHERE name = 'entity-test'`,
	).Scan(&rootID)
	require.NoError(t, err)

	t.Run("CreateMovieEntity", func(t *testing.T) {
		result, err := db.Exec(
			`INSERT INTO media_items
			 (path, title, type, media_type, year, storage_root_id)
			 VALUES (?, ?, ?, ?, ?, ?)`,
			"/media/movies/inception.mkv", "Inception",
			"movie", "movie", 2010, rootID,
		)
		require.NoError(t, err)

		id, err := result.LastInsertId()
		require.NoError(t, err)
		assert.Greater(t, id, int64(0))

		var title, mediaType string
		var year int
		err = db.QueryRow(
			`SELECT title, type, year FROM media_items WHERE id = ?`, id,
		).Scan(&title, &mediaType, &year)
		require.NoError(t, err)

		assert.Equal(t, "Inception", title)
		assert.Equal(t, "movie", mediaType)
		assert.Equal(t, 2010, year)
	})

	t.Run("CreateTVShowHierarchy", func(t *testing.T) {
		// Create TV show (parent)
		showResult, err := db.Exec(
			`INSERT INTO media_items
			 (path, title, type, media_type, storage_root_id)
			 VALUES (?, ?, ?, ?, ?)`,
			"/media/tv/breaking_bad", "Breaking Bad",
			"tv_show", "tv_show", rootID,
		)
		require.NoError(t, err)
		showID, _ := showResult.LastInsertId()

		// Create season (child of show)
		seasonResult, err := db.Exec(
			`INSERT INTO media_items
			 (path, title, type, media_type, season, storage_root_id,
			  user_id)
			 VALUES (?, ?, ?, ?, ?, ?, NULL)`,
			"/media/tv/breaking_bad/s01", "Season 1",
			"tv_season", "tv_season", 1, rootID,
		)
		require.NoError(t, err)
		seasonID, _ := seasonResult.LastInsertId()

		// Create episodes (children of season)
		for ep := 1; ep <= 3; ep++ {
			_, err := db.Exec(
				`INSERT INTO media_items
				 (path, title, type, media_type, season, episode,
				  storage_root_id)
				 VALUES (?, ?, ?, ?, ?, ?, ?)`,
				fmt.Sprintf("/media/tv/breaking_bad/s01/e%02d.mkv", ep),
				fmt.Sprintf("Episode %d", ep),
				"tv_episode", "tv_episode", 1, ep, rootID,
			)
			require.NoError(t, err)
		}

		// Verify hierarchy IDs are valid
		assert.Greater(t, showID, int64(0))
		assert.Greater(t, seasonID, int64(0))

		// Verify episode count
		var epCount int
		err = db.QueryRow(
			`SELECT COUNT(*) FROM media_items
			 WHERE type = 'tv_episode' AND season = 1
			 AND storage_root_id = ?`,
			rootID,
		).Scan(&epCount)
		require.NoError(t, err)
		assert.Equal(t, 3, epCount,
			"Should have 3 episodes in season 1")
	})

	t.Run("CreateMusicHierarchy", func(t *testing.T) {
		// Create album
		_, err := db.Exec(
			`INSERT INTO albums (title, artist, genre, year)
			 VALUES (?, ?, ?, ?)`,
			"OK Computer", "Radiohead", "Alternative", 1997,
		)
		require.NoError(t, err)

		var albumID int64
		err = db.QueryRow(
			`SELECT id FROM albums WHERE title = 'OK Computer'`,
		).Scan(&albumID)
		require.NoError(t, err)

		// Create tracks linked to album
		tracks := []string{"Airbag", "Paranoid Android", "Subterranean"}
		for i, track := range tracks {
			_, err := db.Exec(
				`INSERT INTO media_items
				 (path, title, type, media_type, album_id,
				  track_number, artist, storage_root_id)
				 VALUES (?, ?, 'song', 'song', ?, ?, 'Radiohead', ?)`,
				fmt.Sprintf("/media/music/radiohead/%s.flac", track),
				track, albumID, i+1, rootID,
			)
			require.NoError(t, err)
		}

		var trackCount int
		err = db.QueryRow(
			`SELECT COUNT(*) FROM media_items WHERE album_id = ?`,
			albumID,
		).Scan(&trackCount)
		require.NoError(t, err)
		assert.Equal(t, 3, trackCount,
			"Album should have 3 tracks")
	})
}

func TestEntityPipeline_DuplicateDetection(t *testing.T) {
	db := tests.SetupTestDB(t)
	defer db.Close()

	_, err := db.Exec(
		`INSERT INTO storage_roots (name, protocol, path, enabled)
		 VALUES ('dup-test', 'local', '/dup', 1)`,
	)
	require.NoError(t, err)

	var rootID int64
	err = db.QueryRow(
		`SELECT id FROM storage_roots WHERE name = 'dup-test'`,
	).Scan(&rootID)
	require.NoError(t, err)

	t.Run("SameTitleTypeYearIsDuplicate", func(t *testing.T) {
		// Insert two media items with same title, type, year
		for i := 1; i <= 2; i++ {
			_, err := db.Exec(
				`INSERT INTO media_items
				 (path, title, type, media_type, year, storage_root_id)
				 VALUES (?, 'The Matrix', 'movie', 'movie', 1999, ?)`,
				fmt.Sprintf("/dup/matrix_copy%d.mkv", i), rootID,
			)
			require.NoError(t, err)
		}

		var count int
		err := db.QueryRow(
			`SELECT COUNT(*) FROM media_items
			 WHERE title = 'The Matrix' AND type = 'movie'
			 AND year = 1999`,
		).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 2, count,
			"Should detect two copies of the same title")
	})

	t.Run("DuplicateFilesByHash", func(t *testing.T) {
		// Create a duplicate group
		groupResult, err := db.Exec(
			`INSERT INTO duplicate_groups (file_count, total_size)
			 VALUES (2, 9000000)`,
		)
		require.NoError(t, err)
		groupID, _ := groupResult.LastInsertId()

		// Insert files with same hash in duplicate group
		hash := "abc123def456"
		for i := 1; i <= 2; i++ {
			_, err := db.Exec(
				`INSERT INTO files
				 (storage_root_id, path, name, size, modified_at,
				  md5, is_duplicate, duplicate_group_id)
				 VALUES (?, ?, ?, 4500000, datetime('now'), ?, 1, ?)`,
				rootID,
				fmt.Sprintf("/dup/file_copy%d.mkv", i),
				fmt.Sprintf("file_copy%d.mkv", i),
				hash, groupID,
			)
			require.NoError(t, err)
		}

		var dupCount int
		err = db.QueryRow(
			`SELECT COUNT(*) FROM files
			 WHERE duplicate_group_id = ? AND is_duplicate = 1`,
			groupID,
		).Scan(&dupCount)
		require.NoError(t, err)
		assert.Equal(t, 2, dupCount,
			"Both files should be in the duplicate group")
	})
}

func TestEntityPipeline_MediaTypeDistribution(t *testing.T) {
	db := tests.SetupTestDB(t)
	defer db.Close()

	_, err := db.Exec(
		`INSERT INTO storage_roots (name, protocol, path, enabled)
		 VALUES ('dist-test', 'local', '/dist', 1)`,
	)
	require.NoError(t, err)

	var rootID int64
	err = db.QueryRow(
		`SELECT id FROM storage_roots WHERE name = 'dist-test'`,
	).Scan(&rootID)
	require.NoError(t, err)

	// Insert entities of various types
	types := map[string]int{
		"movie":    5,
		"tv_show":  2,
		"song":     10,
		"game":     3,
		"software": 1,
		"book":     4,
	}

	for mediaType, count := range types {
		for i := 0; i < count; i++ {
			_, err := db.Exec(
				`INSERT INTO media_items
				 (path, title, type, media_type, storage_root_id)
				 VALUES (?, ?, ?, ?, ?)`,
				fmt.Sprintf("/dist/%s_%d", mediaType, i),
				fmt.Sprintf("%s_%d", mediaType, i),
				mediaType, mediaType, rootID,
			)
			require.NoError(t, err)
		}
	}

	t.Run("CountByType", func(t *testing.T) {
		rows, err := db.Query(
			`SELECT type, COUNT(*) as cnt FROM media_items
			 WHERE storage_root_id = ?
			 GROUP BY type ORDER BY cnt DESC`,
			rootID,
		)
		require.NoError(t, err)
		defer rows.Close()

		results := make(map[string]int)
		for rows.Next() {
			var typeName string
			var cnt int
			require.NoError(t, rows.Scan(&typeName, &cnt))
			results[typeName] = cnt
		}
		require.NoError(t, rows.Err())

		assert.Equal(t, 10, results["song"])
		assert.Equal(t, 5, results["movie"])
		assert.Equal(t, 4, results["book"])
		assert.Equal(t, 3, results["game"])
		assert.Equal(t, 2, results["tv_show"])
		assert.Equal(t, 1, results["software"])
	})

	t.Run("TotalEntityCount", func(t *testing.T) {
		var total int
		err := db.QueryRow(
			`SELECT COUNT(*) FROM media_items
			 WHERE storage_root_id = ?`, rootID,
		).Scan(&total)
		require.NoError(t, err)

		expectedTotal := 0
		for _, c := range types {
			expectedTotal += c
		}
		assert.Equal(t, expectedTotal, total,
			"Total entity count should match")
	})
}

func TestEntityPipeline_CoverArtAndMetadata(t *testing.T) {
	db := tests.SetupTestDB(t)
	defer db.Close()

	_, err := db.Exec(
		`INSERT INTO storage_roots (name, protocol, path, enabled)
		 VALUES ('meta-test', 'local', '/meta', 1)`,
	)
	require.NoError(t, err)

	var rootID int64
	err = db.QueryRow(
		`SELECT id FROM storage_roots WHERE name = 'meta-test'`,
	).Scan(&rootID)
	require.NoError(t, err)

	// Create a media item
	result, err := db.Exec(
		`INSERT INTO media_items
		 (path, title, type, media_type, storage_root_id)
		 VALUES ('/meta/movie.mkv', 'Test Movie', 'movie', 'movie', ?)`,
		rootID,
	)
	require.NoError(t, err)
	mediaID, _ := result.LastInsertId()

	t.Run("AttachCoverArt", func(t *testing.T) {
		_, err := db.Exec(
			`INSERT INTO cover_art
			 (media_item_id, source, url, width, height, format,
			  is_default)
			 VALUES (?, 'tmdb', 'https://image.tmdb.org/poster.jpg',
			  300, 450, 'jpeg', 1)`,
			mediaID,
		)
		require.NoError(t, err)

		var source, url string
		var width, height int
		err = db.QueryRow(
			`SELECT source, url, width, height FROM cover_art
			 WHERE media_item_id = ? AND is_default = 1`,
			mediaID,
		).Scan(&source, &url, &width, &height)
		require.NoError(t, err)
		assert.Equal(t, "tmdb", source)
		assert.Equal(t, 300, width)
		assert.Equal(t, 450, height)
	})

	t.Run("AttachSubtitleTracks", func(t *testing.T) {
		langs := []struct {
			lang string
			code string
		}{
			{"English", "en"},
			{"Spanish", "es"},
			{"French", "fr"},
		}
		for _, l := range langs {
			_, err := db.Exec(
				`INSERT INTO subtitle_tracks
				 (media_item_id, language, language_code, source, format)
				 VALUES (?, ?, ?, 'downloaded', 'srt')`,
				mediaID, l.lang, l.code,
			)
			require.NoError(t, err)
		}

		var subCount int
		err := db.QueryRow(
			`SELECT COUNT(*) FROM subtitle_tracks
			 WHERE media_item_id = ?`, mediaID,
		).Scan(&subCount)
		require.NoError(t, err)
		assert.Equal(t, 3, subCount,
			"Should have 3 subtitle tracks")
	})

	t.Run("AttachAudioTracks", func(t *testing.T) {
		_, err := db.Exec(
			`INSERT INTO audio_tracks
			 (media_item_id, language, language_code, codec,
			  channels, bitrate, is_default)
			 VALUES (?, 'English', 'en', 'aac', 6, 640000, 1)`,
			mediaID,
		)
		require.NoError(t, err)

		var codec string
		var channels int
		err = db.QueryRow(
			`SELECT codec, channels FROM audio_tracks
			 WHERE media_item_id = ? AND is_default = 1`,
			mediaID,
		).Scan(&codec, &channels)
		require.NoError(t, err)
		assert.Equal(t, "aac", codec)
		assert.Equal(t, 6, channels,
			"5.1 surround should have 6 channels")
	})
}

func TestEntityPipeline_CascadeDelete(t *testing.T) {
	db := tests.SetupTestDB(t)
	defer db.Close()

	_, err := db.Exec(
		`INSERT INTO storage_roots (name, protocol, path, enabled)
		 VALUES ('cascade-test', 'local', '/cascade', 1)`,
	)
	require.NoError(t, err)

	var rootID int64
	err = db.QueryRow(
		`SELECT id FROM storage_roots WHERE name = 'cascade-test'`,
	).Scan(&rootID)
	require.NoError(t, err)

	// Create media item with related records
	result, err := db.Exec(
		`INSERT INTO media_items
		 (path, title, type, media_type, storage_root_id)
		 VALUES ('/cascade/movie.mkv', 'Cascade Movie', 'movie',
		  'movie', ?)`,
		rootID,
	)
	require.NoError(t, err)
	mediaID, _ := result.LastInsertId()

	// Add subtitle and cover art
	_, err = db.Exec(
		`INSERT INTO subtitle_tracks
		 (media_item_id, language, language_code, source, format)
		 VALUES (?, 'English', 'en', 'embedded', 'srt')`,
		mediaID,
	)
	require.NoError(t, err)

	_, err = db.Exec(
		`INSERT INTO cover_art
		 (media_item_id, source, format, is_default)
		 VALUES (?, 'local', 'jpeg', 1)`,
		mediaID,
	)
	require.NoError(t, err)

	// Delete the media item - should cascade
	_, err = db.Exec(
		`DELETE FROM media_items WHERE id = ?`, mediaID,
	)
	require.NoError(t, err)

	// Verify cascaded deletes
	var subCount, artCount int
	err = db.QueryRow(
		`SELECT COUNT(*) FROM subtitle_tracks
		 WHERE media_item_id = ?`, mediaID,
	).Scan(&subCount)
	require.NoError(t, err)
	assert.Equal(t, 0, subCount,
		"Subtitle tracks should be cascade-deleted")

	err = db.QueryRow(
		`SELECT COUNT(*) FROM cover_art
		 WHERE media_item_id = ?`, mediaID,
	).Scan(&artCount)
	require.NoError(t, err)
	assert.Equal(t, 0, artCount,
		"Cover art should be cascade-deleted")
}

func TestEntityPipeline_PlaybackTracking(t *testing.T) {
	db := tests.SetupTestDB(t)
	defer db.Close()

	_, err := db.Exec(
		`INSERT INTO storage_roots (name, protocol, path, enabled)
		 VALUES ('playback-test', 'local', '/playback', 1)`,
	)
	require.NoError(t, err)

	var rootID int64
	err = db.QueryRow(
		`SELECT id FROM storage_roots WHERE name = 'playback-test'`,
	).Scan(&rootID)
	require.NoError(t, err)

	// Create a media item
	result, err := db.Exec(
		`INSERT INTO media_items
		 (path, title, type, media_type, duration, storage_root_id)
		 VALUES ('/playback/movie.mkv', 'Playback Test', 'movie',
		  'movie', 7200, ?)`,
		rootID,
	)
	require.NoError(t, err)
	mediaID, _ := result.LastInsertId()

	userID := 1 // Default test user

	t.Run("TrackPlaybackPosition", func(t *testing.T) {
		_, err := db.Exec(
			`INSERT INTO playback_positions
			 (user_id, media_item_id, position, duration,
			  percent_complete, last_played)
			 VALUES (?, ?, 3600, 7200, 50.0, datetime('now'))`,
			userID, mediaID,
		)
		require.NoError(t, err)

		var position int
		var pct float64
		err = db.QueryRow(
			`SELECT position, percent_complete
			 FROM playback_positions
			 WHERE user_id = ? AND media_item_id = ?`,
			userID, mediaID,
		).Scan(&position, &pct)
		require.NoError(t, err)
		assert.Equal(t, 3600, position,
			"Position should be at 1 hour")
		assert.InDelta(t, 50.0, pct, 0.01,
			"Should be 50%% complete")
	})

	t.Run("UpdatePlaybackPosition", func(t *testing.T) {
		_, err := db.Exec(
			`UPDATE playback_positions
			 SET position = 7000, percent_complete = 97.2,
			 last_played = datetime('now')
			 WHERE user_id = ? AND media_item_id = ?`,
			userID, mediaID,
		)
		require.NoError(t, err)

		var position int
		err = db.QueryRow(
			`SELECT position FROM playback_positions
			 WHERE user_id = ? AND media_item_id = ?`,
			userID, mediaID,
		).Scan(&position)
		require.NoError(t, err)
		assert.Equal(t, 7000, position)
	})
}

func TestEntityPipeline_BulkInsertPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping bulk insert performance test in short mode")  // SKIP-OK: #legacy-skip-untriaged-2026-04-29
	}

	db := tests.SetupTestDB(t)
	defer db.Close()

	_, err := db.Exec(
		`INSERT INTO storage_roots (name, protocol, path, enabled)
		 VALUES ('bulk-test', 'local', '/bulk', 1)`,
	)
	require.NoError(t, err)

	var rootID int64
	err = db.QueryRow(
		`SELECT id FROM storage_roots WHERE name = 'bulk-test'`,
	).Scan(&rootID)
	require.NoError(t, err)

	entityCount := 500
	start := time.Now()

	tx, err := db.Begin()
	require.NoError(t, err)

	for i := 0; i < entityCount; i++ {
		_, err := tx.Exec(
			`INSERT INTO media_items
			 (path, title, type, media_type, year, storage_root_id)
			 VALUES (?, ?, 'movie', 'movie', ?, ?)`,
			fmt.Sprintf("/bulk/movie_%d.mkv", i),
			fmt.Sprintf("Movie %d", i),
			2000+(i%25), rootID,
		)
		require.NoError(t, err)
	}

	require.NoError(t, tx.Commit())
	elapsed := time.Since(start)

	t.Logf("Inserted %d entities in %v (%.0f entities/sec)",
		entityCount, elapsed,
		float64(entityCount)/elapsed.Seconds())

	var count int
	err = db.QueryRow(
		`SELECT COUNT(*) FROM media_items
		 WHERE storage_root_id = ?`, rootID,
	).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, entityCount, count,
		"All entities should be inserted")

	assert.Less(t, elapsed, 5*time.Second,
		"Bulk insert of 500 entities should complete within 5s")
}
