package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"catalogizer/database"
	"catalogizer/internal/models"

	"github.com/DATA-DOG/go-sqlmock"
	_ "github.com/mutecomm/go-sqlcipher"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newMockStatsRepo(t *testing.T) (*StatsRepository, sqlmock.Sqlmock) {
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	db := database.WrapDB(sqlDB, database.DialectSQLite)
	return NewStatsRepository(db), mock
}

// ---------------------------------------------------------------------------
// GetOverallStats
// ---------------------------------------------------------------------------

func TestStatsRepository_GetOverallStats(t *testing.T) {
	overallStatsCols := []string{
		"total_files", "total_directories", "total_size", "total_duplicates",
		"duplicate_groups", "storage_roots_count", "active_storage_roots", "last_scan_time",
	}

	tests := []struct {
		name    string
		setup   func(mock sqlmock.Sqlmock)
		wantErr bool
		check   func(t *testing.T, stats *models.OverallStats)
	}{
		{
			name: "success with data",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT").
					WillReturnRows(sqlmock.NewRows(overallStatsCols).
						AddRow(1500, 120, int64(1073741824), 50, 20, 3, 2, int64(1700000000)))
			},
			check: func(t *testing.T, stats *models.OverallStats) {
				assert.Equal(t, int64(1500), stats.TotalFiles)
				assert.Equal(t, int64(120), stats.TotalDirectories)
				assert.Equal(t, int64(1073741824), stats.TotalSize)
				assert.Equal(t, int64(50), stats.TotalDuplicates)
				assert.Equal(t, int64(20), stats.DuplicateGroups)
				assert.Equal(t, int64(3), stats.StorageRootsCount)
				assert.Equal(t, int64(2), stats.ActiveStorageRoots)
				assert.Equal(t, int64(1700000000), stats.LastScanTime)
			},
		},
		{
			name: "empty catalog",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT").
					WillReturnRows(sqlmock.NewRows(overallStatsCols).
						AddRow(0, 0, int64(0), 0, 0, 0, 0, int64(0)))
			},
			check: func(t *testing.T, stats *models.OverallStats) {
				assert.Equal(t, int64(0), stats.TotalFiles)
				assert.Equal(t, int64(0), stats.TotalSize)
			},
		},
		{
			name: "database error",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT").
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockStatsRepo(t)
			tt.setup(mock)

			stats, err := repo.GetOverallStats(context.Background())
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, stats)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, stats)
			tt.check(t, stats)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// ---------------------------------------------------------------------------
// GetStorageRootStats
// ---------------------------------------------------------------------------

func TestStatsRepository_GetStorageRootStats(t *testing.T) {
	storageStatsCols := []string{
		"name", "total_files", "total_directories", "total_size",
		"duplicate_files", "duplicate_groups", "last_scan_time", "is_online",
	}

	tests := []struct {
		name    string
		root    string
		setup   func(mock sqlmock.Sqlmock)
		wantErr bool
		errMsg  string
		check   func(t *testing.T, stats *models.StorageRootStats)
	}{
		{
			name: "success",
			root: "my-root",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT").
					WithArgs("my-root").
					WillReturnRows(sqlmock.NewRows(storageStatsCols).
						AddRow("my-root", int64(500), int64(40), int64(536870912), int64(10), int64(5), int64(1700000000), true))
			},
			check: func(t *testing.T, stats *models.StorageRootStats) {
				assert.Equal(t, "my-root", stats.Name)
				assert.Equal(t, int64(500), stats.TotalFiles)
				assert.Equal(t, int64(40), stats.TotalDirectories)
				assert.Equal(t, int64(536870912), stats.TotalSize)
				assert.True(t, stats.IsOnline)
			},
		},
		{
			name: "storage root not found",
			root: "nonexistent",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT").
					WithArgs("nonexistent").
					WillReturnError(sql.ErrNoRows)
			},
			wantErr: true,
			errMsg:  "storage root not found",
		},
		{
			name: "database error",
			root: "my-root",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT").
					WithArgs("my-root").
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
			errMsg:  "failed to get storage root stats",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockStatsRepo(t)
			tt.setup(mock)

			stats, err := repo.GetStorageRootStats(context.Background(), tt.root)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}
			require.NoError(t, err)
			require.NotNil(t, stats)
			tt.check(t, stats)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// ---------------------------------------------------------------------------
// GetFileTypeStats
// ---------------------------------------------------------------------------

func TestStatsRepository_GetFileTypeStats(t *testing.T) {
	fileTypeStatsCols := []string{
		"file_type", "extension", "count", "total_size", "average_size",
	}

	tests := []struct {
		name    string
		root    string
		limit   int
		setup   func(mock sqlmock.Sqlmock)
		wantErr bool
		check   func(t *testing.T, stats []models.FileTypeStats)
	}{
		{
			name:  "all storage roots",
			root:  "",
			limit: 10,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT").
					WithArgs(10).
					WillReturnRows(sqlmock.NewRows(fileTypeStatsCols).
						AddRow("video", "mp4", int64(100), int64(107374182400), 1073741824.0).
						AddRow("audio", "mp3", int64(500), int64(5368709120), 10737418.24).
						AddRow("text", "txt", int64(200), int64(1048576), 5242.88))
			},
			check: func(t *testing.T, stats []models.FileTypeStats) {
				require.Len(t, stats, 3)
				assert.Equal(t, "video", stats[0].FileType)
				assert.Equal(t, "mp4", stats[0].Extension)
				assert.Equal(t, int64(100), stats[0].Count)
				assert.Equal(t, "audio", stats[1].FileType)
				assert.Equal(t, "text", stats[2].FileType)
			},
		},
		{
			name:  "specific storage root",
			root:  "my-root",
			limit: 5,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT").
					WithArgs("my-root", 5).
					WillReturnRows(sqlmock.NewRows(fileTypeStatsCols).
						AddRow("image", "jpg", int64(300), int64(3221225472), 10737418.24))
			},
			check: func(t *testing.T, stats []models.FileTypeStats) {
				require.Len(t, stats, 1)
				assert.Equal(t, "image", stats[0].FileType)
				assert.Equal(t, "jpg", stats[0].Extension)
			},
		},
		{
			name:  "empty result",
			root:  "",
			limit: 10,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT").
					WithArgs(10).
					WillReturnRows(sqlmock.NewRows(fileTypeStatsCols))
			},
			check: func(t *testing.T, stats []models.FileTypeStats) {
				assert.Empty(t, stats)
			},
		},
		{
			name:  "database error",
			root:  "",
			limit: 10,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT").
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockStatsRepo(t)
			tt.setup(mock)

			stats, err := repo.GetFileTypeStats(context.Background(), tt.root, tt.limit)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			tt.check(t, stats)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// ---------------------------------------------------------------------------
// GetSizeDistribution
// ---------------------------------------------------------------------------

func TestStatsRepository_GetSizeDistribution(t *testing.T) {
	sizeDistCols := []string{
		"empty", "tiny", "small", "medium", "large", "huge", "massive",
	}

	tests := []struct {
		name    string
		root    string
		setup   func(mock sqlmock.Sqlmock)
		wantErr bool
		check   func(t *testing.T, dist *models.SizeDistribution)
	}{
		{
			name: "all storage roots",
			root: "",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT").
					WillReturnRows(sqlmock.NewRows(sizeDistCols).
						AddRow(int64(10), int64(50), int64(200), int64(300), int64(150), int64(80), int64(20)))
			},
			check: func(t *testing.T, dist *models.SizeDistribution) {
				// empty (10) is added to tiny (50), so tiny = 60
				assert.Equal(t, int64(60), dist.Tiny)
				assert.Equal(t, int64(200), dist.Small)
				assert.Equal(t, int64(300), dist.Medium)
				assert.Equal(t, int64(150), dist.Large)
				assert.Equal(t, int64(80), dist.Huge)
				assert.Equal(t, int64(20), dist.Massive)
			},
		},
		{
			name: "specific storage root",
			root: "my-root",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT").
					WithArgs("my-root").
					WillReturnRows(sqlmock.NewRows(sizeDistCols).
						AddRow(int64(0), int64(100), int64(50), int64(25), int64(10), int64(5), int64(1)))
			},
			check: func(t *testing.T, dist *models.SizeDistribution) {
				assert.Equal(t, int64(100), dist.Tiny) // 0 empty + 100 tiny
				assert.Equal(t, int64(50), dist.Small)
				assert.Equal(t, int64(1), dist.Massive)
			},
		},
		{
			name: "empty catalog",
			root: "",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT").
					WillReturnRows(sqlmock.NewRows(sizeDistCols).
						AddRow(int64(0), int64(0), int64(0), int64(0), int64(0), int64(0), int64(0)))
			},
			check: func(t *testing.T, dist *models.SizeDistribution) {
				assert.Equal(t, int64(0), dist.Tiny)
				assert.Equal(t, int64(0), dist.Small)
				assert.Equal(t, int64(0), dist.Medium)
				assert.Equal(t, int64(0), dist.Large)
				assert.Equal(t, int64(0), dist.Huge)
				assert.Equal(t, int64(0), dist.Massive)
			},
		},
		{
			name: "database error",
			root: "",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT").
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockStatsRepo(t)
			tt.setup(mock)

			dist, err := repo.GetSizeDistribution(context.Background(), tt.root)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, dist)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, dist)
			tt.check(t, dist)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// ---------------------------------------------------------------------------
// GetDuplicateStats
// ---------------------------------------------------------------------------

func TestStatsRepository_GetDuplicateStats(t *testing.T) {
	dupStatsCols := []string{
		"total_duplicates", "duplicate_groups", "wasted_space",
		"largest_group", "average_group_size",
	}

	tests := []struct {
		name    string
		root    string
		setup   func(mock sqlmock.Sqlmock)
		wantErr bool
		check   func(t *testing.T, stats *models.DuplicateStats)
	}{
		{
			name: "all storage roots",
			root: "",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT").
					WillReturnRows(sqlmock.NewRows(dupStatsCols).
						AddRow(int64(50), int64(20), int64(536870912), int64(10), 2.5))
			},
			check: func(t *testing.T, stats *models.DuplicateStats) {
				assert.Equal(t, int64(50), stats.TotalDuplicates)
				assert.Equal(t, int64(20), stats.DuplicateGroups)
				assert.Equal(t, int64(536870912), stats.WastedSpace)
				assert.Equal(t, int64(10), stats.LargestDuplicateGroup)
				assert.InDelta(t, 2.5, stats.AverageGroupSize, 0.01)
			},
		},
		{
			name: "no duplicates",
			root: "",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT").
					WillReturnRows(sqlmock.NewRows(dupStatsCols).
						AddRow(int64(0), int64(0), int64(0), int64(0), 0.0))
			},
			check: func(t *testing.T, stats *models.DuplicateStats) {
				assert.Equal(t, int64(0), stats.TotalDuplicates)
				assert.Equal(t, int64(0), stats.WastedSpace)
			},
		},
		{
			name: "database error",
			root: "",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT").
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockStatsRepo(t)
			tt.setup(mock)

			stats, err := repo.GetDuplicateStats(context.Background(), tt.root)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, stats)
			tt.check(t, stats)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// ===========================================================================
// Real SQLite-backed tests for uncovered functions
// ===========================================================================

func newRealStatsRepo(t *testing.T) *StatsRepository {
	t.Helper()
	sqlDB, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { sqlDB.Close() })

	db := database.WrapDB(sqlDB, database.DialectSQLite)

	_, err = sqlDB.Exec(`
		CREATE TABLE storage_roots (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			path TEXT NOT NULL,
			enabled INTEGER NOT NULL DEFAULT 1
		)
	`)
	require.NoError(t, err)

	_, err = sqlDB.Exec(`
		CREATE TABLE files (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			path TEXT NOT NULL,
			size INTEGER NOT NULL DEFAULT 0,
			file_type TEXT,
			extension TEXT,
			is_directory INTEGER NOT NULL DEFAULT 0,
			is_duplicate INTEGER NOT NULL DEFAULT 0,
			duplicate_group_id INTEGER,
			deleted INTEGER NOT NULL DEFAULT 0,
			storage_root_id INTEGER,
			created_at INTEGER NOT NULL DEFAULT 0,
			accessed_at INTEGER,
			last_scan_at DATETIME
		)
	`)
	require.NoError(t, err)

	_, err = sqlDB.Exec(`
		CREATE TABLE duplicate_groups (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			file_count INTEGER NOT NULL DEFAULT 0,
			total_size INTEGER NOT NULL DEFAULT 0
		)
	`)
	require.NoError(t, err)

	_, err = sqlDB.Exec(`
		CREATE TABLE scan_history (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			storage_root_id INTEGER NOT NULL,
			scan_type TEXT NOT NULL,
			status TEXT NOT NULL,
			start_time DATETIME NOT NULL,
			end_time DATETIME,
			files_processed INTEGER NOT NULL DEFAULT 0,
			files_added INTEGER NOT NULL DEFAULT 0,
			files_updated INTEGER NOT NULL DEFAULT 0,
			files_deleted INTEGER NOT NULL DEFAULT 0,
			error_count INTEGER NOT NULL DEFAULT 0,
			error_message TEXT
		)
	`)
	require.NoError(t, err)

	return NewStatsRepository(db)
}

func seedStatsData(t *testing.T, repo *StatsRepository) {
	t.Helper()
	now := time.Now().Unix()

	// Create storage roots
	_, err := repo.db.Exec(`INSERT INTO storage_roots (name, path, enabled) VALUES (?, ?, ?)`, "root1", "/data/root1", 1)
	require.NoError(t, err)
	_, err = repo.db.Exec(`INSERT INTO storage_roots (name, path, enabled) VALUES (?, ?, ?)`, "root2", "/data/root2", 1)
	require.NoError(t, err)

	// Create duplicate groups
	_, err = repo.db.Exec(`INSERT INTO duplicate_groups (file_count, total_size) VALUES (?, ?)`, 3, 9000)
	require.NoError(t, err)
	_, err = repo.db.Exec(`INSERT INTO duplicate_groups (file_count, total_size) VALUES (?, ?)`, 2, 4000)
	require.NoError(t, err)

	// Create files: regular files, directories, duplicates
	files := []struct {
		name, path, fileType, ext string
		size                      int64
		isDir, isDup              int
		dupGroupID                *int
		storageRootID             int
		createdAt                 int64
		accessedAt                *int64
	}{
		{"movie.mp4", "/data/root1/movie.mp4", "video", "mp4", 5000, 0, 1, intPtr(1), 1, now - 86400, int64Ptr(now)},
		{"movie_copy.mp4", "/data/root1/movie_copy.mp4", "video", "mp4", 5000, 0, 1, intPtr(1), 1, now - 86400, nil},
		{"movie_copy2.mp4", "/data/root2/movie_copy2.mp4", "video", "mp4", 5000, 0, 1, intPtr(1), 2, now - 86400, nil},
		{"song.mp3", "/data/root1/song.mp3", "audio", "mp3", 2000, 0, 1, intPtr(2), 1, now - 172800, nil},
		{"song_copy.mp3", "/data/root1/song_copy.mp3", "audio", "mp3", 2000, 0, 1, intPtr(2), 1, now - 172800, nil},
		{"photo.jpg", "/data/root1/photo.jpg", "image", "jpg", 1000, 0, 0, nil, 1, now - 2592000, int64Ptr(now - 3600)},
		{"docs", "/data/root1/docs", "", "", 0, 1, 0, nil, 1, now - 86400, nil},
	}

	for _, f := range files {
		_, err := repo.db.Exec(
			`INSERT INTO files (name, path, file_type, extension, size, is_directory, is_duplicate, duplicate_group_id, deleted, storage_root_id, created_at, accessed_at)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, 0, ?, ?, ?)`,
			f.name, f.path, f.fileType, f.ext, f.size, f.isDir, f.isDup, f.dupGroupID, f.storageRootID, f.createdAt, f.accessedAt,
		)
		require.NoError(t, err)
	}

	// Create scan history entries
	scanNow := time.Now().Truncate(time.Second)
	scanEnd := scanNow.Add(5 * time.Minute)
	_, err = repo.db.Exec(
		`INSERT INTO scan_history (storage_root_id, scan_type, status, start_time, end_time, files_processed, files_added, files_updated, files_deleted, error_count)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		1, "full", "completed", scanNow.Add(-time.Hour), scanEnd.Add(-time.Hour), 100, 50, 30, 5, 0,
	)
	require.NoError(t, err)
	_, err = repo.db.Exec(
		`INSERT INTO scan_history (storage_root_id, scan_type, status, start_time, end_time, files_processed, files_added, files_updated, files_deleted, error_count, error_message)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		2, "incremental", "failed", scanNow, scanEnd, 10, 2, 1, 0, 3, "timeout",
	)
	require.NoError(t, err)
}

func intPtr(i int) *int       { return &i }
func int64Ptr(i int64) *int64 { return &i }

// ---------------------------------------------------------------------------
// GetTopDuplicateGroups
// ---------------------------------------------------------------------------

func TestStatsRepository_GetTopDuplicateGroups_Real(t *testing.T) {
	repo := newRealStatsRepo(t)
	seedStatsData(t, repo)
	ctx := context.Background()

	t.Run("sort by count", func(t *testing.T) {
		groups, err := repo.GetTopDuplicateGroups(ctx, "count", 10, "")
		require.NoError(t, err)
		require.Len(t, groups, 2)
		// First group should have higher file_count (3)
		assert.Equal(t, int64(3), groups[0].FileCount)
		assert.Equal(t, int64(2), groups[1].FileCount)
	})

	t.Run("sort by size", func(t *testing.T) {
		groups, err := repo.GetTopDuplicateGroups(ctx, "size", 10, "")
		require.NoError(t, err)
		require.Len(t, groups, 2)
		// First group should have higher total_size (9000)
		assert.Equal(t, int64(9000), groups[0].TotalSize)
	})

	t.Run("limit results", func(t *testing.T) {
		groups, err := repo.GetTopDuplicateGroups(ctx, "count", 1, "")
		require.NoError(t, err)
		assert.Len(t, groups, 1)
	})

	t.Run("filter by storage root", func(t *testing.T) {
		groups, err := repo.GetTopDuplicateGroups(ctx, "count", 10, "root2")
		require.NoError(t, err)
		// Only group 1 has files in root2
		assert.Len(t, groups, 1)
		assert.Equal(t, int64(1), groups[0].GroupID)
	})

	t.Run("empty result for nonexistent root", func(t *testing.T) {
		groups, err := repo.GetTopDuplicateGroups(ctx, "count", 10, "nonexistent")
		require.NoError(t, err)
		assert.Empty(t, groups)
	})
}

// ---------------------------------------------------------------------------
// GetAccessPatterns
// ---------------------------------------------------------------------------

func TestStatsRepository_GetAccessPatterns_Real(t *testing.T) {
	repo := newRealStatsRepo(t)
	seedStatsData(t, repo)
	ctx := context.Background()

	t.Run("all storage roots", func(t *testing.T) {
		patterns, err := repo.GetAccessPatterns(ctx, "", 30)
		require.NoError(t, err)
		require.NotNil(t, patterns)
		// Files with accessed_at != nil and > threshold: movie.mp4, photo.jpg
		assert.Equal(t, int64(2), patterns.RecentlyAccessed)
		// Files with no accessed_at: movie_copy.mp4, movie_copy2.mp4, song.mp3, song_copy.mp3
		assert.Equal(t, int64(4), patterns.NeverAccessed)
		assert.Len(t, patterns.AccessFrequency, 30)
		assert.NotNil(t, patterns.PopularExtensions)
		assert.NotNil(t, patterns.PopularDirectories)
	})

	t.Run("filter by storage root", func(t *testing.T) {
		patterns, err := repo.GetAccessPatterns(ctx, "root1", 30)
		require.NoError(t, err)
		require.NotNil(t, patterns)
		// root1 has: movie.mp4 (accessed), movie_copy.mp4 (null), song.mp3 (null), song_copy.mp3 (null), photo.jpg (accessed)
		assert.Equal(t, int64(2), patterns.RecentlyAccessed)
		assert.Equal(t, int64(3), patterns.NeverAccessed)
	})
}

// ---------------------------------------------------------------------------
// GetGrowthTrends
// ---------------------------------------------------------------------------

func TestStatsRepository_GetGrowthTrends_Real(t *testing.T) {
	repo := newRealStatsRepo(t)
	seedStatsData(t, repo)
	ctx := context.Background()

	t.Run("all storage roots", func(t *testing.T) {
		trends, err := repo.GetGrowthTrends(ctx, "", 12)
		require.NoError(t, err)
		require.NotNil(t, trends)
		assert.NotNil(t, trends.MonthlyGrowth)
		assert.Equal(t, 0.0, trends.TotalGrowthRate) // simplified implementation
	})

	t.Run("filter by storage root", func(t *testing.T) {
		trends, err := repo.GetGrowthTrends(ctx, "root1", 12)
		require.NoError(t, err)
		require.NotNil(t, trends)
	})

	t.Run("narrow window", func(t *testing.T) {
		trends, err := repo.GetGrowthTrends(ctx, "", 1)
		require.NoError(t, err)
		require.NotNil(t, trends)
	})
}

// ---------------------------------------------------------------------------
// GetScanHistory
// ---------------------------------------------------------------------------

func TestStatsRepository_GetScanHistory_Real(t *testing.T) {
	repo := newRealStatsRepo(t)
	seedStatsData(t, repo)
	ctx := context.Background()

	t.Run("all scan history", func(t *testing.T) {
		history, totalCount, err := repo.GetScanHistory(ctx, "", 10, 0)
		require.NoError(t, err)
		assert.Equal(t, int64(2), totalCount)
		assert.Len(t, history, 2)
		// Most recent first
		assert.Equal(t, "incremental", history[0].ScanType)
		assert.Equal(t, "failed", history[0].Status)
		assert.Equal(t, int64(3), history[0].ErrorCount)
		assert.NotNil(t, history[0].ErrorMessage)
		assert.Equal(t, "timeout", *history[0].ErrorMessage)
	})

	t.Run("pagination", func(t *testing.T) {
		history, totalCount, err := repo.GetScanHistory(ctx, "", 1, 0)
		require.NoError(t, err)
		assert.Equal(t, int64(2), totalCount)
		assert.Len(t, history, 1)

		history2, _, err := repo.GetScanHistory(ctx, "", 1, 1)
		require.NoError(t, err)
		assert.Len(t, history2, 1)
		assert.NotEqual(t, history[0].ID, history2[0].ID)
	})
}
