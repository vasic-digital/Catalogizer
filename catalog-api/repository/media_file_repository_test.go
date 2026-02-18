package repository

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"testing"
	"time"

	"catalogizer/database"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newMockMediaFileRepo creates a MediaFileRepository backed by sqlmock.
func newMockMediaFileRepo(t *testing.T) (*MediaFileRepository, sqlmock.Sqlmock) {
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	db := database.WrapDB(sqlDB, database.DialectSQLite)
	return NewMediaFileRepository(db), mock
}

// mediaFileColumns is the standard column set for media_files queries.
var mediaFileColumns = []string{
	"id", "media_item_id", "file_id", "quality_info", "language", "is_primary", "created_at",
}

func sampleMediaFileRow(now time.Time) []driver.Value {
	return []driver.Value{
		int64(1), int64(10), int64(100), nil, nil, true, now,
	}
}

// ---------------------------------------------------------------------------
// LinkFileToItem
// ---------------------------------------------------------------------------

func TestMediaFileRepository_LinkFileToItem(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(mock sqlmock.Sqlmock)
		wantErr bool
		wantID  int64
	}{
		{
			name: "success",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO media_files").
					WillReturnResult(sqlmock.NewResult(7, 1))
			},
			wantID: 7,
		},
		{
			name: "database error",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO media_files").
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockMediaFileRepo(t)
			tt.setup(mock)

			id, err := repo.LinkFileToItem(context.Background(), 10, 100, nil, nil, true)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantID, id)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// ---------------------------------------------------------------------------
// GetFilesByItem
// ---------------------------------------------------------------------------

func TestMediaFileRepository_GetFilesByItem(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name      string
		itemID    int64
		setup     func(mock sqlmock.Sqlmock)
		wantErr   bool
		wantCount int
		check     func(t *testing.T, records []MediaFileRecord)
	}{
		{
			name:   "returns records",
			itemID: 10,
			setup: func(mock sqlmock.Sqlmock) {
				row1 := sampleMediaFileRow(now)
				row2 := sampleMediaFileRow(now)
				row2[0] = int64(2)
				row2[2] = int64(200)
				row2[5] = false
				mock.ExpectQuery("SELECT .+ FROM media_files").
					WithArgs(int64(10)).
					WillReturnRows(sqlmock.NewRows(mediaFileColumns).
						AddRow(row1...).
						AddRow(row2...))
			},
			wantCount: 2,
			check: func(t *testing.T, records []MediaFileRecord) {
				assert.Equal(t, int64(100), records[0].FileID)
				assert.True(t, records[0].IsPrimary)
				assert.Equal(t, int64(200), records[1].FileID)
				assert.False(t, records[1].IsPrimary)
			},
		},
		{
			name:   "empty result",
			itemID: 99,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT .+ FROM media_files").
					WithArgs(int64(99)).
					WillReturnRows(sqlmock.NewRows(mediaFileColumns))
			},
			wantCount: 0,
		},
		{
			name:   "database error",
			itemID: 10,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT .+ FROM media_files").
					WithArgs(int64(10)).
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockMediaFileRepo(t)
			tt.setup(mock)

			records, err := repo.GetFilesByItem(context.Background(), tt.itemID)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Len(t, records, tt.wantCount)
			if tt.check != nil {
				tt.check(t, records)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// ---------------------------------------------------------------------------
// GetItemByFile
// ---------------------------------------------------------------------------

func TestMediaFileRepository_GetItemByFile(t *testing.T) {
	tests := []struct {
		name      string
		fileID    int64
		setup     func(mock sqlmock.Sqlmock)
		wantErr   bool
		wantCount int
	}{
		{
			name:   "returns item IDs",
			fileID: 100,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT media_item_id").
					WithArgs(int64(100)).
					WillReturnRows(sqlmock.NewRows([]string{"media_item_id"}).
						AddRow(int64(10)).
						AddRow(int64(20)))
			},
			wantCount: 2,
		},
		{
			name:   "empty result",
			fileID: 999,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT media_item_id").
					WithArgs(int64(999)).
					WillReturnRows(sqlmock.NewRows([]string{"media_item_id"}))
			},
			wantCount: 0,
		},
		{
			name:   "database error",
			fileID: 100,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT media_item_id").
					WithArgs(int64(100)).
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockMediaFileRepo(t)
			tt.setup(mock)

			ids, err := repo.GetItemByFile(context.Background(), tt.fileID)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Len(t, ids, tt.wantCount)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// ---------------------------------------------------------------------------
// UnlinkFile
// ---------------------------------------------------------------------------

func TestMediaFileRepository_UnlinkFile(t *testing.T) {
	tests := []struct {
		name    string
		itemID  int64
		fileID  int64
		setup   func(mock sqlmock.Sqlmock)
		wantErr bool
	}{
		{
			name:   "success",
			itemID: 10,
			fileID: 100,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE FROM media_files WHERE media_item_id").
					WithArgs(int64(10), int64(100)).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
		},
		{
			name:   "database error",
			itemID: 10,
			fileID: 100,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE FROM media_files WHERE media_item_id").
					WithArgs(int64(10), int64(100)).
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockMediaFileRepo(t)
			tt.setup(mock)

			err := repo.UnlinkFile(context.Background(), tt.itemID, tt.fileID)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// ---------------------------------------------------------------------------
// CountByItem
// ---------------------------------------------------------------------------

func TestMediaFileRepository_CountByItem(t *testing.T) {
	tests := []struct {
		name    string
		itemID  int64
		setup   func(mock sqlmock.Sqlmock)
		wantErr bool
		want    int64
	}{
		{
			name:   "returns count",
			itemID: 10,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT COUNT").
					WithArgs(int64(10)).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))
			},
			want: 3,
		},
		{
			name:   "database error",
			itemID: 10,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT COUNT").
					WithArgs(int64(10)).
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockMediaFileRepo(t)
			tt.setup(mock)

			count, err := repo.CountByItem(context.Background(), tt.itemID)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, count)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// ---------------------------------------------------------------------------
// SetPrimary
// ---------------------------------------------------------------------------

func TestMediaFileRepository_SetPrimary(t *testing.T) {
	tests := []struct {
		name    string
		itemID  int64
		fileID  int64
		setup   func(mock sqlmock.Sqlmock)
		wantErr bool
	}{
		{
			name:   "success",
			itemID: 10,
			fileID: 100,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("UPDATE media_files SET is_primary").
					WithArgs(false, int64(10)).
					WillReturnResult(sqlmock.NewResult(0, 2))
				mock.ExpectExec("UPDATE media_files SET is_primary").
					WithArgs(true, int64(10), int64(100)).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
		},
		{
			name:   "no matching file",
			itemID: 10,
			fileID: 999,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("UPDATE media_files SET is_primary").
					WithArgs(false, int64(10)).
					WillReturnResult(sqlmock.NewResult(0, 0))
				mock.ExpectExec("UPDATE media_files SET is_primary").
					WithArgs(true, int64(10), int64(999)).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			wantErr: true,
		},
		{
			name:   "clear error",
			itemID: 10,
			fileID: 100,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("UPDATE media_files SET is_primary").
					WithArgs(false, int64(10)).
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockMediaFileRepo(t)
			tt.setup(mock)

			err := repo.SetPrimary(context.Background(), tt.itemID, tt.fileID)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
