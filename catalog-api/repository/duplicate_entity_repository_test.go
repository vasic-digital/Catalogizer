package repository

import (
	"context"
	"testing"

	"catalogizer/database"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newMockDuplicateEntityRepo(t *testing.T) (*DuplicateEntityRepository, sqlmock.Sqlmock) {
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	db := database.WrapDB(sqlDB, database.DialectSQLite)
	return NewDuplicateEntityRepository(db), mock
}

func TestDuplicateEntityRepository_CountDuplicates(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(mock sqlmock.Sqlmock)
		want    int64
		wantErr bool
	}{
		{
			name: "returns count",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT COUNT").
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))
			},
			want: 5,
		},
		{
			name: "zero duplicates",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT COUNT").
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
			},
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockDuplicateEntityRepo(t)
			tt.setup(mock)

			got, err := repo.CountDuplicates(context.Background())
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestDuplicateEntityRepository_GetDuplicateGroups(t *testing.T) {
	repo, mock := newMockDuplicateEntityRepo(t)
	ctx := context.Background()

	// Count query
	mock.ExpectQuery("SELECT COUNT").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	// Groups query
	year := 1999
	mock.ExpectQuery("SELECT mi.title").
		WithArgs(10, 0).
		WillReturnRows(sqlmock.NewRows([]string{"title", "name", "year", "cnt"}).
			AddRow("The Matrix", "movie", year, 2))

	// IDs query for the group
	mock.ExpectQuery("SELECT mi.id FROM media_items").
		WithArgs("The Matrix", "movie", year).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1).AddRow(2))

	groups, total, err := repo.GetDuplicateGroups(ctx, 10, 0)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, groups, 1)
	assert.Equal(t, "The Matrix", groups[0].Title)
	assert.Equal(t, "movie", groups[0].MediaType)
	assert.Equal(t, 2, groups[0].Count)
	assert.Equal(t, []int64{1, 2}, groups[0].EntityIDs)
	assert.NoError(t, mock.ExpectationsWereMet())
}
