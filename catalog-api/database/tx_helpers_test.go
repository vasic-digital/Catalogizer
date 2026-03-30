package database

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/mutecomm/go-sqlcipher"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func newTxTestDB(t *testing.T) *DB {
	t.Helper()
	sqlDB, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)

	db := WrapDB(sqlDB, DialectSQLite)
	require.NotNil(t, db)

	_, err = db.ExecContext(context.Background(), `
		CREATE TABLE tx_test (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL
		)
	`)
	require.NoError(t, err)

	return db
}

// ---------------------------------------------------------------------------
// TxInsertReturningID — SQLite
// ---------------------------------------------------------------------------

func TestTxInsertReturningID_SQLite_Success(t *testing.T) {
	db := newTxTestDB(t)
	ctx := context.Background()

	tx, err := db.Begin()
	require.NoError(t, err)

	id, err := db.TxInsertReturningID(ctx, tx, "INSERT INTO tx_test (name) VALUES (?)", "first")
	require.NoError(t, err)
	assert.Equal(t, int64(1), id)

	id2, err := db.TxInsertReturningID(ctx, tx, "INSERT INTO tx_test (name) VALUES (?)", "second")
	require.NoError(t, err)
	assert.Equal(t, int64(2), id2)

	err = tx.Commit()
	require.NoError(t, err)

	// Verify rows committed
	var count int
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM tx_test").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 2, count)
}

func TestTxInsertReturningID_SQLite_Rollback(t *testing.T) {
	db := newTxTestDB(t)
	ctx := context.Background()

	tx, err := db.Begin()
	require.NoError(t, err)

	_, err = db.TxInsertReturningID(ctx, tx, "INSERT INTO tx_test (name) VALUES (?)", "rollback-me")
	require.NoError(t, err)

	err = tx.Rollback()
	require.NoError(t, err)

	// Verify row was NOT committed
	var count int
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM tx_test").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestTxInsertReturningID_SQLite_InvalidSQL(t *testing.T) {
	db := newTxTestDB(t)
	ctx := context.Background()

	tx, err := db.Begin()
	require.NoError(t, err)
	defer tx.Rollback()

	_, err = db.TxInsertReturningID(ctx, tx, "INSERT INTO nonexistent_table (name) VALUES (?)", "fail")
	assert.Error(t, err)
}

func TestTxInsertReturningID_SQLite_MultipleInserts(t *testing.T) {
	db := newTxTestDB(t)
	ctx := context.Background()

	tx, err := db.Begin()
	require.NoError(t, err)

	ids := make([]int64, 0, 5)
	for i := 0; i < 5; i++ {
		id, err := db.TxInsertReturningID(ctx, tx, "INSERT INTO tx_test (name) VALUES (?)", "item")
		require.NoError(t, err)
		ids = append(ids, id)
	}

	err = tx.Commit()
	require.NoError(t, err)

	// IDs should be sequential
	for i := 1; i < len(ids); i++ {
		assert.Equal(t, ids[i-1]+1, ids[i])
	}
}
