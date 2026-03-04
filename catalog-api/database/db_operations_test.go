package database

import (
	"context"
	"os"
	"testing"

	"catalogizer/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// WrapDB
// ---------------------------------------------------------------------------

func TestWrapDB_NilReturnsNil(t *testing.T) {
	db := WrapDB(nil, DialectSQLite)
	assert.Nil(t, db)
}

func TestWrapDB_SQLite(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "wrapdb_test_*.db")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	cfg := &config.DatabaseConfig{
		Path:        tmpFile.Name(),
		EnableWAL:   true,
		BusyTimeout: 5000,
	}

	raw, err := NewConnection(cfg)
	require.NoError(t, err)
	defer raw.Close()

	wrapped := WrapDB(raw.DB, DialectSQLite)
	require.NotNil(t, wrapped)
	assert.True(t, wrapped.Dialect().IsSQLite())
	assert.False(t, wrapped.Dialect().IsPostgres())
}

func TestWrapDB_Postgres(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "wrapdb_pg_test_*.db")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	cfg := &config.DatabaseConfig{
		Path:        tmpFile.Name(),
		EnableWAL:   true,
		BusyTimeout: 5000,
	}

	raw, err := NewConnection(cfg)
	require.NoError(t, err)
	defer raw.Close()

	// Wrap with postgres dialect type even though underlying is sqlite
	wrapped := WrapDB(raw.DB, DialectPostgres)
	require.NotNil(t, wrapped)
	assert.True(t, wrapped.Dialect().IsPostgres())
	assert.False(t, wrapped.Dialect().IsSQLite())
}

// ---------------------------------------------------------------------------
// Dialect accessor
// ---------------------------------------------------------------------------

func TestDB_Dialect(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "dialect_test_*.db")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	cfg := &config.DatabaseConfig{
		Path:        tmpFile.Name(),
		EnableWAL:   true,
		BusyTimeout: 5000,
	}

	db, err := NewConnection(cfg)
	require.NoError(t, err)
	defer db.Close()

	d := db.Dialect()
	require.NotNil(t, d)
	assert.True(t, d.IsSQLite())
}

// ---------------------------------------------------------------------------
// DatabaseType
// ---------------------------------------------------------------------------

func TestDB_DatabaseType(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "dbtype_test_*.db")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	cfg := &config.DatabaseConfig{
		Path:        tmpFile.Name(),
		EnableWAL:   true,
		BusyTimeout: 5000,
	}

	db, err := NewConnection(cfg)
	require.NoError(t, err)
	defer db.Close()

	assert.Equal(t, "sqlite", db.DatabaseType())
}

// ---------------------------------------------------------------------------
// TableExists
// ---------------------------------------------------------------------------

func TestDB_TableExists(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Before migrations, the migrations table should not exist
	exists, err := db.TableExists(ctx, "nonexistent_table")
	assert.NoError(t, err)
	assert.False(t, exists)

	// Run migrations so tables are created
	err = db.RunMigrations(ctx)
	require.NoError(t, err)

	// Now storage_roots should exist
	exists, err = db.TableExists(ctx, "storage_roots")
	assert.NoError(t, err)
	assert.True(t, exists)

	// users table should exist
	exists, err = db.TableExists(ctx, "users")
	assert.NoError(t, err)
	assert.True(t, exists)

	// A non-existent table should still be false
	exists, err = db.TableExists(ctx, "definitely_not_a_table")
	assert.NoError(t, err)
	assert.False(t, exists)
}

// ---------------------------------------------------------------------------
// Exec / Query / QueryRow (the shadowed dialect-aware methods)
// ---------------------------------------------------------------------------

func TestDB_ShadowedMethods(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()

	ctx := context.Background()
	err := db.RunMigrations(ctx)
	require.NoError(t, err)

	// Test Exec (non-context version)
	_, err = db.Exec("INSERT OR IGNORE INTO roles (name, description, permissions, is_system) VALUES (?, ?, ?, ?)",
		"test_role", "A test role", "[]", false)
	assert.NoError(t, err)

	// Test Query (non-context version)
	rows, err := db.Query("SELECT name FROM roles WHERE name = ?", "test_role")
	require.NoError(t, err)
	defer rows.Close()

	found := false
	for rows.Next() {
		var name string
		err := rows.Scan(&name)
		require.NoError(t, err)
		assert.Equal(t, "test_role", name)
		found = true
	}
	assert.True(t, found)

	// Test QueryRow (non-context version)
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM roles WHERE name = ?", "test_role").Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
}

// ---------------------------------------------------------------------------
// QueryContext
// ---------------------------------------------------------------------------

func TestDB_QueryContext(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()

	ctx := context.Background()
	err := db.RunMigrations(ctx)
	require.NoError(t, err)

	rows, err := db.QueryContext(ctx, "SELECT name FROM roles ORDER BY name")
	require.NoError(t, err)
	defer rows.Close()

	var names []string
	for rows.Next() {
		var name string
		require.NoError(t, rows.Scan(&name))
		names = append(names, name)
	}
	assert.GreaterOrEqual(t, len(names), 2, "should have at least admin and user roles")
}

// ---------------------------------------------------------------------------
// InsertReturningID (SQLite path)
// ---------------------------------------------------------------------------

func TestDB_InsertReturningID_SQLite(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()

	ctx := context.Background()
	err := db.RunMigrations(ctx)
	require.NoError(t, err)

	// Insert a user
	id, err := db.InsertReturningID(ctx,
		`INSERT INTO users (username, email, password_hash, salt, role_id, is_active, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
		"testuser", "test@example.com", "hash123", "salt123", 1, true)
	require.NoError(t, err)
	assert.Greater(t, id, int64(0))

	// Verify it was inserted
	var username string
	err = db.QueryRowContext(ctx, "SELECT username FROM users WHERE id = ?", id).Scan(&username)
	assert.NoError(t, err)
	assert.Equal(t, "testuser", username)

	// Insert another and verify auto-increment
	id2, err := db.InsertReturningID(ctx,
		`INSERT INTO users (username, email, password_hash, salt, role_id, is_active, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
		"testuser2", "test2@example.com", "hash456", "salt456", 1, true)
	require.NoError(t, err)
	assert.Greater(t, id2, id)
}

// ---------------------------------------------------------------------------
// TxInsertReturningID (SQLite path)
// ---------------------------------------------------------------------------

func TestDB_TxInsertReturningID_SQLite(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()

	ctx := context.Background()
	err := db.RunMigrations(ctx)
	require.NoError(t, err)

	tx, err := db.Begin()
	require.NoError(t, err)

	id, err := db.TxInsertReturningID(ctx, tx,
		`INSERT INTO users (username, email, password_hash, salt, role_id, is_active, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
		"txuser", "tx@example.com", "txhash", "txsalt", 1, true)
	require.NoError(t, err)
	assert.Greater(t, id, int64(0))

	err = tx.Commit()
	require.NoError(t, err)

	// Verify the insert persisted
	var username string
	err = db.QueryRowContext(ctx, "SELECT username FROM users WHERE id = ?", id).Scan(&username)
	assert.NoError(t, err)
	assert.Equal(t, "txuser", username)
}

func TestDB_TxInsertReturningID_Rollback(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()

	ctx := context.Background()
	err := db.RunMigrations(ctx)
	require.NoError(t, err)

	tx, err := db.Begin()
	require.NoError(t, err)

	id, err := db.TxInsertReturningID(ctx, tx,
		`INSERT INTO users (username, email, password_hash, salt, role_id, is_active, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
		"rollbackuser", "rollback@example.com", "rbhash", "rbsalt", 1, true)
	require.NoError(t, err)
	assert.Greater(t, id, int64(0))

	err = tx.Rollback()
	require.NoError(t, err)

	// Verify the insert was rolled back
	var count int
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users WHERE username = ?", "rollbackuser").Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 0, count)
}

// ---------------------------------------------------------------------------
// rewriteQuery
// ---------------------------------------------------------------------------

func TestDB_RewriteQuery(t *testing.T) {
	// SQLite: no rewriting
	sqliteDB := &DB{dialect: Dialect{Type: DialectSQLite}, config: &config.DatabaseConfig{}}

	got := sqliteDB.rewriteQuery("SELECT * FROM t WHERE id = ? AND deleted = 0")
	assert.Equal(t, "SELECT * FROM t WHERE id = ? AND deleted = 0", got)

	// Postgres: full rewriting pipeline
	pgDB := &DB{dialect: Dialect{Type: DialectPostgres}, config: &config.DatabaseConfig{}}

	got = pgDB.rewriteQuery("SELECT * FROM t WHERE id = ? AND deleted = 0")
	assert.Equal(t, "SELECT * FROM t WHERE id = $1 AND deleted = FALSE", got)

	got = pgDB.rewriteQuery("INSERT OR IGNORE INTO t (a) VALUES (?)")
	assert.Equal(t, "INSERT INTO t (a) VALUES ($1) ON CONFLICT DO NOTHING", got)

	got = pgDB.rewriteQuery("INSERT OR REPLACE INTO t (a) VALUES (?)")
	assert.Equal(t, "INSERT INTO t (a) VALUES ($1)", got)

	got = pgDB.rewriteQuery("SELECT * FROM t WHERE is_active = 1 AND id = ?")
	assert.Equal(t, "SELECT * FROM t WHERE is_active = TRUE AND id = $1", got)
}

// ---------------------------------------------------------------------------
// createContext
// ---------------------------------------------------------------------------

func TestDB_CreateContext_DefaultTimeout(t *testing.T) {
	db := &DB{
		config:  &config.DatabaseConfig{BusyTimeout: 0},
		dialect: Dialect{Type: DialectSQLite},
	}

	ctx, cancel := db.createContext()
	defer cancel()

	// Context should have a deadline (5s default when BusyTimeout <= 0)
	deadline, ok := ctx.Deadline()
	assert.True(t, ok)
	assert.False(t, deadline.IsZero())
}

func TestDB_CreateContext_CustomTimeout(t *testing.T) {
	db := &DB{
		config:  &config.DatabaseConfig{BusyTimeout: 10000}, // 10 seconds
		dialect: Dialect{Type: DialectSQLite},
	}

	ctx, cancel := db.createContext()
	defer cancel()

	deadline, ok := ctx.Deadline()
	assert.True(t, ok)
	assert.False(t, deadline.IsZero())
}

// ---------------------------------------------------------------------------
// NewConnection edge cases
// ---------------------------------------------------------------------------

func TestNewConnection_EmptyType_DefaultsToSQLite(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "deftype_test_*.db")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	cfg := &config.DatabaseConfig{
		Type:        "",
		Path:        tmpFile.Name(),
		BusyTimeout: 5000,
	}

	db, err := NewConnection(cfg)
	require.NoError(t, err)
	defer db.Close()

	assert.Equal(t, "sqlite", db.DatabaseType())
}

func TestNewConnection_ExplicitSQLiteType(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "explicit_sqlite_test_*.db")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	cfg := &config.DatabaseConfig{
		Type:        "sqlite",
		Path:        tmpFile.Name(),
		BusyTimeout: 5000,
	}

	db, err := NewConnection(cfg)
	require.NoError(t, err)
	defer db.Close()

	assert.Equal(t, "sqlite", db.DatabaseType())
	assert.True(t, db.Dialect().IsSQLite())
}

func TestNewConnection_ZeroPoolSettings(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "zeropool_test_*.db")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	cfg := &config.DatabaseConfig{
		Path:               tmpFile.Name(),
		MaxOpenConnections: 0, // should not be applied
		MaxIdleConnections: 0,
		ConnMaxLifetime:    0,
		ConnMaxIdleTime:    0,
		BusyTimeout:        5000,
	}

	db, err := NewConnection(cfg)
	require.NoError(t, err)
	defer db.Close()

	err = db.Ping()
	assert.NoError(t, err)
}
