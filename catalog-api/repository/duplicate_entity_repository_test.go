package repository

import (
	"context"
	"fmt"
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
		{
			name: "db returns error",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT COUNT").
					WillReturnError(fmt.Errorf("connection refused"))
			},
			want:    0,
			wantErr: true,
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

func TestDuplicateEntityRepository_GetDuplicateGroups_ZeroGroups(t *testing.T) {
	repo, mock := newMockDuplicateEntityRepo(t)
	ctx := context.Background()

	// Count query returns 0
	mock.ExpectQuery("SELECT COUNT").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	// Groups query returns empty result set
	mock.ExpectQuery("SELECT mi.title").
		WithArgs(10, 0).
		WillReturnRows(sqlmock.NewRows([]string{"title", "name", "year", "cnt"}))

	groups, total, err := repo.GetDuplicateGroups(ctx, 10, 0)
	require.NoError(t, err)
	assert.Equal(t, int64(0), total)
	assert.Empty(t, groups)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDuplicateEntityRepository_GetDuplicateGroups_MultipleGroups(t *testing.T) {
	repo, mock := newMockDuplicateEntityRepo(t)
	ctx := context.Background()

	year1 := 1999
	year2 := 2008

	// Count query
	mock.ExpectQuery("SELECT COUNT").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	// Groups query returns two groups
	mock.ExpectQuery("SELECT mi.title").
		WithArgs(10, 0).
		WillReturnRows(sqlmock.NewRows([]string{"title", "name", "year", "cnt"}).
			AddRow("The Matrix", "movie", year1, 3).
			AddRow("The Dark Knight", "movie", year2, 2))

	// IDs query for first group
	mock.ExpectQuery("SELECT mi.id FROM media_items").
		WithArgs("The Matrix", "movie", year1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1).AddRow(2).AddRow(3))

	// IDs query for second group
	mock.ExpectQuery("SELECT mi.id FROM media_items").
		WithArgs("The Dark Knight", "movie", year2).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(4).AddRow(5))

	groups, total, err := repo.GetDuplicateGroups(ctx, 10, 0)
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, groups, 2)

	assert.Equal(t, "The Matrix", groups[0].Title)
	assert.Equal(t, 3, groups[0].Count)
	assert.Equal(t, []int64{1, 2, 3}, groups[0].EntityIDs)

	assert.Equal(t, "The Dark Knight", groups[1].Title)
	assert.Equal(t, 2, groups[1].Count)
	assert.Equal(t, []int64{4, 5}, groups[1].EntityIDs)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDuplicateEntityRepository_GetDuplicateGroups_NullYear(t *testing.T) {
	repo, mock := newMockDuplicateEntityRepo(t)
	ctx := context.Background()

	// Count query
	mock.ExpectQuery("SELECT COUNT").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	// Groups query with nil year
	mock.ExpectQuery("SELECT mi.title").
		WithArgs(10, 0).
		WillReturnRows(sqlmock.NewRows([]string{"title", "name", "year", "cnt"}).
			AddRow("Unknown Movie", "movie", nil, 2))

	// IDs query: when year is nil, the query appends "AND mi.year IS NULL"
	// so only title + type are passed as args
	mock.ExpectQuery("SELECT mi.id FROM media_items").
		WithArgs("Unknown Movie", "movie").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(10).AddRow(11))

	groups, total, err := repo.GetDuplicateGroups(ctx, 10, 0)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, groups, 1)
	assert.Nil(t, groups[0].Year)
	assert.Equal(t, "Unknown Movie", groups[0].Title)
	assert.Equal(t, []int64{10, 11}, groups[0].EntityIDs)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDuplicateEntityRepository_GetDuplicateGroups_CountQueryError(t *testing.T) {
	repo, mock := newMockDuplicateEntityRepo(t)
	ctx := context.Background()

	// Count query fails
	mock.ExpectQuery("SELECT COUNT").
		WillReturnError(fmt.Errorf("database locked"))

	groups, total, err := repo.GetDuplicateGroups(ctx, 10, 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to count duplicate groups")
	assert.Equal(t, int64(0), total)
	assert.Nil(t, groups)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDuplicateEntityRepository_GetDuplicateGroups_GroupsQueryError(t *testing.T) {
	repo, mock := newMockDuplicateEntityRepo(t)
	ctx := context.Background()

	// Count query succeeds
	mock.ExpectQuery("SELECT COUNT").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))

	// Groups query fails
	mock.ExpectQuery("SELECT mi.title").
		WithArgs(10, 0).
		WillReturnError(fmt.Errorf("syntax error"))

	groups, total, err := repo.GetDuplicateGroups(ctx, 10, 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to query duplicate groups")
	assert.Equal(t, int64(0), total)
	assert.Nil(t, groups)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDuplicateEntityRepository_CountVsGetDuplicateGroupsConsistency(t *testing.T) {
	// Verify that CountDuplicates and the total from GetDuplicateGroups
	// both use the same underlying count query pattern and return the
	// same total when the database state is identical.
	ctx := context.Background()

	// Setup CountDuplicates mock
	repoCount, mockCount := newMockDuplicateEntityRepo(t)
	mockCount.ExpectQuery("SELECT COUNT").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(7))

	count, err := repoCount.CountDuplicates(ctx)
	require.NoError(t, err)

	// Setup GetDuplicateGroups mock with same count
	repoGroups, mockGroups := newMockDuplicateEntityRepo(t)
	mockGroups.ExpectQuery("SELECT COUNT").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(7))
	mockGroups.ExpectQuery("SELECT mi.title").
		WithArgs(100, 0).
		WillReturnRows(sqlmock.NewRows([]string{"title", "name", "year", "cnt"}))

	_, total, err := repoGroups.GetDuplicateGroups(ctx, 100, 0)
	require.NoError(t, err)

	// Both should report the same total
	assert.Equal(t, count, total, "CountDuplicates and GetDuplicateGroups total should be consistent")
	assert.NoError(t, mockCount.ExpectationsWereMet())
	assert.NoError(t, mockGroups.ExpectationsWereMet())
}
