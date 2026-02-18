package repository

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"testing"
	"time"

	"catalogizer/database"
	"catalogizer/internal/media/models"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newMockDirectoryAnalysisRepo creates a DirectoryAnalysisRepository backed by sqlmock.
func newMockDirectoryAnalysisRepo(t *testing.T) (*DirectoryAnalysisRepository, sqlmock.Sqlmock) {
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	db := database.WrapDB(sqlDB, database.DialectSQLite)
	return NewDirectoryAnalysisRepository(db), mock
}

// directoryAnalysisColumns is the standard column set for directory_analyses queries.
var directoryAnalysisColumns = []string{
	"id", "directory_path", "smb_root", "media_item_id", "confidence_score",
	"detection_method", "analysis_data", "last_analyzed", "files_count", "total_size",
}

func sampleDirectoryAnalysisRow(now time.Time) []driver.Value {
	return []driver.Value{
		int64(1), "/media/movies/The Matrix (1999)", "nas-share", nil, 0.95,
		"hybrid", `{"matched_patterns":["movie_title_year"],"quality_indicators":["1080p"]}`, now, 5, int64(4200000000),
	}
}

// ---------------------------------------------------------------------------
// Create
// ---------------------------------------------------------------------------

func TestDirectoryAnalysisRepository_Create(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(mock sqlmock.Sqlmock)
		wantErr bool
		wantID  int64
	}{
		{
			name: "success",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO directory_analyses").
					WillReturnResult(sqlmock.NewResult(8, 1))
			},
			wantID: 8,
		},
		{
			name: "database error",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO directory_analyses").
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockDirectoryAnalysisRepo(t)
			tt.setup(mock)

			da := &models.DirectoryAnalysis{
				DirectoryPath:   "/media/movies/The Matrix (1999)",
				SmbRoot:         "nas-share",
				ConfidenceScore: 0.95,
				DetectionMethod: "hybrid",
				FilesCount:      5,
				TotalSize:       4200000000,
				AnalysisData: &models.AnalysisData{
					MatchedPatterns:   []string{"movie_title_year"},
					QualityIndicators: []string{"1080p"},
				},
			}
			id, err := repo.Create(context.Background(), da)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantID, id)
			assert.Equal(t, tt.wantID, da.ID)
			assert.False(t, da.LastAnalyzed.IsZero())
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// ---------------------------------------------------------------------------
// GetByPath
// ---------------------------------------------------------------------------

func TestDirectoryAnalysisRepository_GetByPath(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		path    string
		setup   func(mock sqlmock.Sqlmock)
		wantNil bool
		wantErr bool
		check   func(t *testing.T, da *models.DirectoryAnalysis)
	}{
		{
			name: "found",
			path: "/media/movies/The Matrix (1999)",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT .+ FROM directory_analyses WHERE directory_path").
					WithArgs("/media/movies/The Matrix (1999)").
					WillReturnRows(sqlmock.NewRows(directoryAnalysisColumns).
						AddRow(sampleDirectoryAnalysisRow(now)...))
			},
			check: func(t *testing.T, da *models.DirectoryAnalysis) {
				assert.Equal(t, int64(1), da.ID)
				assert.Equal(t, "/media/movies/The Matrix (1999)", da.DirectoryPath)
				assert.Equal(t, "nas-share", da.SmbRoot)
				assert.Equal(t, 0.95, da.ConfidenceScore)
				assert.Equal(t, "hybrid", da.DetectionMethod)
				assert.Equal(t, 5, da.FilesCount)
				assert.Equal(t, int64(4200000000), da.TotalSize)
				assert.Nil(t, da.MediaItemID)
				require.NotNil(t, da.AnalysisData)
				assert.Contains(t, da.AnalysisData.MatchedPatterns, "movie_title_year")
				assert.Contains(t, da.AnalysisData.QualityIndicators, "1080p")
			},
		},
		{
			name: "not found returns nil",
			path: "/nonexistent",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT .+ FROM directory_analyses WHERE directory_path").
					WithArgs("/nonexistent").
					WillReturnError(sql.ErrNoRows)
			},
			wantNil: true,
		},
		{
			name: "database error",
			path: "/test",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT .+ FROM directory_analyses WHERE directory_path").
					WithArgs("/test").
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockDirectoryAnalysisRepo(t)
			tt.setup(mock)

			da, err := repo.GetByPath(context.Background(), tt.path)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			if tt.wantNil {
				assert.Nil(t, da)
			} else {
				require.NotNil(t, da)
				tt.check(t, da)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// ---------------------------------------------------------------------------
// Update
// ---------------------------------------------------------------------------

func TestDirectoryAnalysisRepository_Update(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(mock sqlmock.Sqlmock)
		wantErr bool
	}{
		{
			name: "success",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("UPDATE directory_analyses SET").
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
		},
		{
			name: "database error",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("UPDATE directory_analyses SET").
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockDirectoryAnalysisRepo(t)
			tt.setup(mock)

			itemID := int64(42)
			da := &models.DirectoryAnalysis{
				ID:              1,
				SmbRoot:         "nas-share",
				MediaItemID:     &itemID,
				ConfidenceScore: 0.9,
				DetectionMethod: "hybrid",
				FilesCount:      5,
				TotalSize:       4200000000,
			}
			err := repo.Update(context.Background(), da)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.False(t, da.LastAnalyzed.IsZero())
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// ---------------------------------------------------------------------------
// GetUnprocessed
// ---------------------------------------------------------------------------

func TestDirectoryAnalysisRepository_GetUnprocessed(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name      string
		limit     int
		setup     func(mock sqlmock.Sqlmock)
		wantErr   bool
		wantCount int
		check     func(t *testing.T, items []*models.DirectoryAnalysis)
	}{
		{
			name:  "returns unprocessed ordered by confidence",
			limit: 10,
			setup: func(mock sqlmock.Sqlmock) {
				row1 := sampleDirectoryAnalysisRow(now)
				// row1 has confidence 0.95
				row2 := sampleDirectoryAnalysisRow(now)
				row2[0] = int64(2)
				row2[1] = "/media/movies/Another"
				row2[4] = 0.7
				mock.ExpectQuery("SELECT .+ FROM directory_analyses WHERE media_item_id IS NULL").
					WithArgs(10).
					WillReturnRows(sqlmock.NewRows(directoryAnalysisColumns).
						AddRow(row1...).
						AddRow(row2...))
			},
			wantCount: 2,
			check: func(t *testing.T, items []*models.DirectoryAnalysis) {
				assert.Equal(t, 0.95, items[0].ConfidenceScore)
				assert.Equal(t, 0.7, items[1].ConfidenceScore)
			},
		},
		{
			name:  "empty result",
			limit: 10,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT .+ FROM directory_analyses WHERE media_item_id IS NULL").
					WithArgs(10).
					WillReturnRows(sqlmock.NewRows(directoryAnalysisColumns))
			},
			wantCount: 0,
		},
		{
			name:  "database error",
			limit: 10,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT .+ FROM directory_analyses WHERE media_item_id IS NULL").
					WithArgs(10).
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockDirectoryAnalysisRepo(t)
			tt.setup(mock)

			items, err := repo.GetUnprocessed(context.Background(), tt.limit)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Len(t, items, tt.wantCount)
			if tt.check != nil {
				tt.check(t, items)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
