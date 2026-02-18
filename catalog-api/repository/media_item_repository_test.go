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

// newMockMediaItemRepo creates a MediaItemRepository backed by sqlmock.
func newMockMediaItemRepo(t *testing.T) (*MediaItemRepository, sqlmock.Sqlmock) {
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	db := database.WrapDB(sqlDB, database.DialectSQLite)
	return NewMediaItemRepository(db), mock
}

// mediaItemColumns is the standard column set returned by media_items queries.
var mediaItemColumns = []string{
	"id", "media_type_id", "title", "original_title", "year", "description",
	"genre", "director", "cast_crew", "rating", "runtime", "language", "country",
	"status", "parent_id", "season_number", "episode_number", "track_number",
	"first_detected", "last_updated",
}

func sampleMediaItemRow(now time.Time) []driver.Value {
	return []driver.Value{
		int64(1), int64(2), "Test Movie", nil, 2024, nil,
		`["Action","Drama"]`, nil, nil, 8.5, 120, nil, nil,
		"active", nil, nil, nil, nil,
		now, now,
	}
}

// ---------------------------------------------------------------------------
// Create
// ---------------------------------------------------------------------------

func TestMediaItemRepository_Create(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(mock sqlmock.Sqlmock)
		wantErr bool
		wantID  int64
	}{
		{
			name: "success",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO media_items").
					WillReturnResult(sqlmock.NewResult(42, 1))
			},
			wantID: 42,
		},
		{
			name: "database error",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO media_items").
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockMediaItemRepo(t)
			tt.setup(mock)

			item := &models.MediaItem{
				MediaTypeID: 2,
				Title:       "Test Movie",
				Status:      "active",
			}
			id, err := repo.Create(context.Background(), item)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantID, id)
			assert.Equal(t, tt.wantID, item.ID)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// ---------------------------------------------------------------------------
// GetByID
// ---------------------------------------------------------------------------

func TestMediaItemRepository_GetByID(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		id      int64
		setup   func(mock sqlmock.Sqlmock)
		wantErr bool
		errMsg  string
		check   func(t *testing.T, item *models.MediaItem)
	}{
		{
			name: "found",
			id:   1,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT .+ FROM media_items WHERE id").
					WithArgs(int64(1)).
					WillReturnRows(sqlmock.NewRows(mediaItemColumns).AddRow(sampleMediaItemRow(now)...))
			},
			check: func(t *testing.T, item *models.MediaItem) {
				assert.Equal(t, int64(1), item.ID)
				assert.Equal(t, "Test Movie", item.Title)
				assert.Equal(t, "active", item.Status)
				assert.Equal(t, int64(2), item.MediaTypeID)
				require.NotNil(t, item.Genre)
				assert.Contains(t, item.Genre, "Action")
			},
		},
		{
			name: "not found",
			id:   999,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT .+ FROM media_items WHERE id").
					WithArgs(int64(999)).
					WillReturnError(sql.ErrNoRows)
			},
			wantErr: true,
			errMsg:  "media item not found",
		},
		{
			name: "database error",
			id:   1,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT .+ FROM media_items WHERE id").
					WithArgs(int64(1)).
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
			errMsg:  "failed to get media item",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockMediaItemRepo(t)
			tt.setup(mock)

			item, err := repo.GetByID(context.Background(), tt.id)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}
			require.NoError(t, err)
			require.NotNil(t, item)
			tt.check(t, item)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// ---------------------------------------------------------------------------
// GetByType
// ---------------------------------------------------------------------------

func TestMediaItemRepository_GetByType(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name      string
		typeID    int64
		limit     int
		offset    int
		setup     func(mock sqlmock.Sqlmock)
		wantErr   bool
		wantCount int
		wantTotal int64
	}{
		{
			name:   "returns items",
			typeID: 2,
			limit:  10,
			offset: 0,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT COUNT").
					WithArgs(int64(2)).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
				mock.ExpectQuery("SELECT .+ FROM media_items WHERE media_type_id").
					WithArgs(int64(2), 10, 0).
					WillReturnRows(sqlmock.NewRows(mediaItemColumns).AddRow(sampleMediaItemRow(now)...))
			},
			wantCount: 1,
			wantTotal: 1,
		},
		{
			name:   "empty result",
			typeID: 99,
			limit:  10,
			offset: 0,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT COUNT").
					WithArgs(int64(99)).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
				mock.ExpectQuery("SELECT .+ FROM media_items WHERE media_type_id").
					WithArgs(int64(99), 10, 0).
					WillReturnRows(sqlmock.NewRows(mediaItemColumns))
			},
			wantCount: 0,
			wantTotal: 0,
		},
		{
			name:   "count error",
			typeID: 2,
			limit:  10,
			offset: 0,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT COUNT").
					WithArgs(int64(2)).
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockMediaItemRepo(t)
			tt.setup(mock)

			items, total, err := repo.GetByType(context.Background(), tt.typeID, tt.limit, tt.offset)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Len(t, items, tt.wantCount)
			assert.Equal(t, tt.wantTotal, total)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// ---------------------------------------------------------------------------
// CountByType
// ---------------------------------------------------------------------------

func TestMediaItemRepository_CountByType(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(mock sqlmock.Sqlmock)
		wantErr bool
		check   func(t *testing.T, counts map[string]int64)
	}{
		{
			name: "returns counts",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT mt.name, COUNT").
					WillReturnRows(sqlmock.NewRows([]string{"name", "count"}).
						AddRow("movies", 10).
						AddRow("music", 5))
			},
			check: func(t *testing.T, counts map[string]int64) {
				assert.Equal(t, int64(10), counts["movies"])
				assert.Equal(t, int64(5), counts["music"])
			},
		},
		{
			name: "empty result",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT mt.name, COUNT").
					WillReturnRows(sqlmock.NewRows([]string{"name", "count"}))
			},
			check: func(t *testing.T, counts map[string]int64) {
				assert.Empty(t, counts)
			},
		},
		{
			name: "database error",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT mt.name, COUNT").
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockMediaItemRepo(t)
			tt.setup(mock)

			counts, err := repo.CountByType(context.Background())
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			tt.check(t, counts)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// ---------------------------------------------------------------------------
// Count
// ---------------------------------------------------------------------------

func TestMediaItemRepository_Count(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(mock sqlmock.Sqlmock)
		wantErr bool
		want    int64
	}{
		{
			name: "returns count",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT COUNT").
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(42))
			},
			want: 42,
		},
		{
			name: "database error",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT COUNT").
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockMediaItemRepo(t)
			tt.setup(mock)

			count, err := repo.Count(context.Background())
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
// Delete
// ---------------------------------------------------------------------------

func TestMediaItemRepository_Delete(t *testing.T) {
	tests := []struct {
		name    string
		id      int64
		setup   func(mock sqlmock.Sqlmock)
		wantErr bool
	}{
		{
			name: "success",
			id:   1,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE FROM media_items WHERE id").
					WithArgs(int64(1)).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
		},
		{
			name: "database error",
			id:   1,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE FROM media_items WHERE id").
					WithArgs(int64(1)).
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockMediaItemRepo(t)
			tt.setup(mock)

			err := repo.Delete(context.Background(), tt.id)
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
// GetChildren
// ---------------------------------------------------------------------------

func TestMediaItemRepository_GetChildren(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name      string
		parentID  int64
		setup     func(mock sqlmock.Sqlmock)
		wantErr   bool
		wantCount int
	}{
		{
			name:     "returns children",
			parentID: 10,
			setup: func(mock sqlmock.Sqlmock) {
				row1 := sampleMediaItemRow(now)
				row1[0] = int64(11)
				row1[2] = "Season 1"
				row2 := sampleMediaItemRow(now)
				row2[0] = int64(12)
				row2[2] = "Season 2"
				mock.ExpectQuery("SELECT .+ FROM media_items WHERE parent_id").
					WithArgs(int64(10)).
					WillReturnRows(sqlmock.NewRows(mediaItemColumns).
						AddRow(row1...).
						AddRow(row2...))
			},
			wantCount: 2,
		},
		{
			name:     "no children",
			parentID: 99,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT .+ FROM media_items WHERE parent_id").
					WithArgs(int64(99)).
					WillReturnRows(sqlmock.NewRows(mediaItemColumns))
			},
			wantCount: 0,
		},
		{
			name:     "database error",
			parentID: 10,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT .+ FROM media_items WHERE parent_id").
					WithArgs(int64(10)).
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockMediaItemRepo(t)
			tt.setup(mock)

			children, err := repo.GetChildren(context.Background(), tt.parentID)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Len(t, children, tt.wantCount)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// ---------------------------------------------------------------------------
// GetByTitle
// ---------------------------------------------------------------------------

func TestMediaItemRepository_GetByTitle(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		title   string
		typeID  int64
		setup   func(mock sqlmock.Sqlmock)
		wantNil bool
		wantErr bool
		check   func(t *testing.T, item *models.MediaItem)
	}{
		{
			name:   "found",
			title:  "Test Movie",
			typeID: 2,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT .+ FROM media_items WHERE title").
					WithArgs("Test Movie", int64(2)).
					WillReturnRows(sqlmock.NewRows(mediaItemColumns).AddRow(sampleMediaItemRow(now)...))
			},
			check: func(t *testing.T, item *models.MediaItem) {
				assert.Equal(t, "Test Movie", item.Title)
			},
		},
		{
			name:   "not found returns nil",
			title:  "Nonexistent",
			typeID: 2,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT .+ FROM media_items WHERE title").
					WithArgs("Nonexistent", int64(2)).
					WillReturnError(sql.ErrNoRows)
			},
			wantNil: true,
		},
		{
			name:   "database error",
			title:  "Test",
			typeID: 2,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT .+ FROM media_items WHERE title").
					WithArgs("Test", int64(2)).
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockMediaItemRepo(t)
			tt.setup(mock)

			item, err := repo.GetByTitle(context.Background(), tt.title, tt.typeID)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			if tt.wantNil {
				assert.Nil(t, item)
			} else {
				require.NotNil(t, item)
				tt.check(t, item)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// ---------------------------------------------------------------------------
// GetMediaTypeByName
// ---------------------------------------------------------------------------

func TestMediaItemRepository_GetMediaTypeByName(t *testing.T) {
	now := time.Now()

	mediaTypeColumns := []string{
		"id", "name", "description", "detection_patterns", "metadata_providers",
		"created_at", "updated_at",
	}

	tests := []struct {
		name    string
		typName string
		setup   func(mock sqlmock.Sqlmock)
		wantErr bool
		errMsg  string
		check   func(t *testing.T, mt *models.MediaType, id int64)
	}{
		{
			name:    "found",
			typName: "movie",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT .+ FROM media_types WHERE name").
					WithArgs("movie").
					WillReturnRows(sqlmock.NewRows(mediaTypeColumns).
						AddRow(int64(1), "movie", "Movies", `["*.mkv","*.mp4"]`, `["tmdb","imdb"]`, now, now))
			},
			check: func(t *testing.T, mt *models.MediaType, id int64) {
				assert.Equal(t, int64(1), id)
				assert.Equal(t, "movie", mt.Name)
				assert.Contains(t, mt.DetectionPatterns, "*.mkv")
				assert.Contains(t, mt.MetadataProviders, "tmdb")
			},
		},
		{
			name:    "not found",
			typName: "nonexistent",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT .+ FROM media_types WHERE name").
					WithArgs("nonexistent").
					WillReturnError(sql.ErrNoRows)
			},
			wantErr: true,
			errMsg:  "media type not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockMediaItemRepo(t)
			tt.setup(mock)

			mt, id, err := repo.GetMediaTypeByName(context.Background(), tt.typName)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}
			require.NoError(t, err)
			require.NotNil(t, mt)
			tt.check(t, mt, id)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
