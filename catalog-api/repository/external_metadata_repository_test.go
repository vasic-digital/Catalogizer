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

// newMockExternalMetadataRepo creates an ExternalMetadataRepository backed by sqlmock.
func newMockExternalMetadataRepo(t *testing.T) (*ExternalMetadataRepository, sqlmock.Sqlmock) {
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	db := database.WrapDB(sqlDB, database.DialectSQLite)
	return NewExternalMetadataRepository(db), mock
}

// externalMetadataColumns is the standard column set for external_metadata queries.
var externalMetadataColumns = []string{
	"id", "media_item_id", "provider", "external_id", "data", "rating",
	"review_url", "cover_url", "trailer_url", "last_fetched",
}

func sampleExternalMetadataRow(now time.Time) []driver.Value {
	rating := 8.2
	coverURL := "https://image.tmdb.org/cover.jpg"
	return []driver.Value{
		int64(1), int64(10), "tmdb", "12345", `{"title":"Test"}`, &rating,
		nil, &coverURL, nil, now,
	}
}

// ---------------------------------------------------------------------------
// Create
// ---------------------------------------------------------------------------

func TestExternalMetadataRepository_Create(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(mock sqlmock.Sqlmock)
		wantErr bool
		wantID  int64
	}{
		{
			name: "success",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO external_metadata").
					WillReturnResult(sqlmock.NewResult(5, 1))
			},
			wantID: 5,
		},
		{
			name: "database error",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO external_metadata").
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockExternalMetadataRepo(t)
			tt.setup(mock)

			em := &models.ExternalMetadata{
				MediaItemID: 10,
				Provider:    "tmdb",
				ExternalID:  "12345",
				Data:        `{"title":"Test"}`,
			}
			id, err := repo.Create(context.Background(), em)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantID, id)
			assert.Equal(t, tt.wantID, em.ID)
			assert.False(t, em.LastFetched.IsZero())
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// ---------------------------------------------------------------------------
// GetByItem
// ---------------------------------------------------------------------------

func TestExternalMetadataRepository_GetByItem(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name      string
		itemID    int64
		setup     func(mock sqlmock.Sqlmock)
		wantErr   bool
		wantCount int
		check     func(t *testing.T, items []*models.ExternalMetadata)
	}{
		{
			name:   "returns metadata",
			itemID: 10,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT .+ FROM external_metadata WHERE media_item_id").
					WithArgs(int64(10)).
					WillReturnRows(sqlmock.NewRows(externalMetadataColumns).
						AddRow(sampleExternalMetadataRow(now)...))
			},
			wantCount: 1,
			check: func(t *testing.T, items []*models.ExternalMetadata) {
				assert.Equal(t, "tmdb", items[0].Provider)
				assert.Equal(t, "12345", items[0].ExternalID)
				assert.NotNil(t, items[0].Rating)
				assert.Equal(t, 8.2, *items[0].Rating)
			},
		},
		{
			name:   "empty result",
			itemID: 99,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT .+ FROM external_metadata WHERE media_item_id").
					WithArgs(int64(99)).
					WillReturnRows(sqlmock.NewRows(externalMetadataColumns))
			},
			wantCount: 0,
		},
		{
			name:   "database error",
			itemID: 10,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT .+ FROM external_metadata WHERE media_item_id").
					WithArgs(int64(10)).
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockExternalMetadataRepo(t)
			tt.setup(mock)

			items, err := repo.GetByItem(context.Background(), tt.itemID)
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

// ---------------------------------------------------------------------------
// GetByProvider
// ---------------------------------------------------------------------------

func TestExternalMetadataRepository_GetByProvider(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name       string
		provider   string
		externalID string
		setup      func(mock sqlmock.Sqlmock)
		wantNil    bool
		wantErr    bool
		check      func(t *testing.T, em *models.ExternalMetadata)
	}{
		{
			name:       "found",
			provider:   "tmdb",
			externalID: "12345",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT .+ FROM external_metadata WHERE provider").
					WithArgs("tmdb", "12345").
					WillReturnRows(sqlmock.NewRows(externalMetadataColumns).
						AddRow(sampleExternalMetadataRow(now)...))
			},
			check: func(t *testing.T, em *models.ExternalMetadata) {
				assert.Equal(t, "tmdb", em.Provider)
				assert.Equal(t, "12345", em.ExternalID)
			},
		},
		{
			name:       "not found returns nil",
			provider:   "imdb",
			externalID: "tt0000",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT .+ FROM external_metadata WHERE provider").
					WithArgs("imdb", "tt0000").
					WillReturnError(sql.ErrNoRows)
			},
			wantNil: true,
		},
		{
			name:       "database error",
			provider:   "tmdb",
			externalID: "12345",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT .+ FROM external_metadata WHERE provider").
					WithArgs("tmdb", "12345").
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockExternalMetadataRepo(t)
			tt.setup(mock)

			em, err := repo.GetByProvider(context.Background(), tt.provider, tt.externalID)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			if tt.wantNil {
				assert.Nil(t, em)
			} else {
				require.NotNil(t, em)
				tt.check(t, em)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// ---------------------------------------------------------------------------
// Delete
// ---------------------------------------------------------------------------

func TestExternalMetadataRepository_Delete(t *testing.T) {
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
				mock.ExpectExec("DELETE FROM external_metadata WHERE id").
					WithArgs(int64(1)).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
		},
		{
			name: "database error",
			id:   1,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE FROM external_metadata WHERE id").
					WithArgs(int64(1)).
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockExternalMetadataRepo(t)
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
// Upsert (insert path)
// ---------------------------------------------------------------------------

func TestExternalMetadataRepository_Upsert_Insert(t *testing.T) {
	repo, mock := newMockExternalMetadataRepo(t)

	// findByItemAndProvider returns no rows (nothing exists yet)
	mock.ExpectQuery("SELECT .+ FROM external_metadata WHERE media_item_id").
		WithArgs(int64(10), "tmdb").
		WillReturnError(sql.ErrNoRows)

	// Then Create is called
	mock.ExpectExec("INSERT INTO external_metadata").
		WillReturnResult(sqlmock.NewResult(1, 1))

	em := &models.ExternalMetadata{
		MediaItemID: 10,
		Provider:    "tmdb",
		ExternalID:  "12345",
		Data:        `{"title":"Test"}`,
	}
	err := repo.Upsert(context.Background(), em)
	require.NoError(t, err)
	assert.Equal(t, int64(1), em.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ---------------------------------------------------------------------------
// Upsert (update path)
// ---------------------------------------------------------------------------

func TestExternalMetadataRepository_Upsert_Update(t *testing.T) {
	now := time.Now()
	repo, mock := newMockExternalMetadataRepo(t)

	// findByItemAndProvider returns existing record
	mock.ExpectQuery("SELECT .+ FROM external_metadata WHERE media_item_id").
		WithArgs(int64(10), "tmdb").
		WillReturnRows(sqlmock.NewRows(externalMetadataColumns).
			AddRow(sampleExternalMetadataRow(now)...))

	// Then UPDATE is called
	mock.ExpectExec("UPDATE external_metadata SET").
		WillReturnResult(sqlmock.NewResult(0, 1))

	em := &models.ExternalMetadata{
		MediaItemID: 10,
		Provider:    "tmdb",
		ExternalID:  "67890",
		Data:        `{"title":"Updated"}`,
	}
	err := repo.Upsert(context.Background(), em)
	require.NoError(t, err)
	assert.Equal(t, int64(1), em.ID) // inherits existing ID
	assert.NoError(t, mock.ExpectationsWereMet())
}
