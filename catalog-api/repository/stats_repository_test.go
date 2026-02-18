package repository

import (
	"context"
	"database/sql"
	"testing"

	"catalogizer/database"
	"catalogizer/internal/models"

	"github.com/DATA-DOG/go-sqlmock"
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
