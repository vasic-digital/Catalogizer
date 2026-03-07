package database

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"catalogizer/config"

	_ "github.com/mutecomm/go-sqlcipher"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// InsertReturningID — multiple inserts, edge cases
// ---------------------------------------------------------------------------

func TestInsertReturningID_MultipleInserts(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()

	ctx := context.Background()
	err := db.RunMigrations(ctx)
	require.NoError(t, err)

	// Insert multiple rows and verify IDs increase monotonically
	var ids []int64
	for i := 0; i < 5; i++ {
		id, err := db.InsertReturningID(ctx,
			`INSERT INTO users (username, email, password_hash, salt, role_id, is_active, created_at, updated_at)
			 VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
			"user_"+string(rune('a'+i)), "user"+string(rune('a'+i))+"@example.com",
			"hash", "salt", 1, true)
		require.NoError(t, err)
		assert.Greater(t, id, int64(0))
		ids = append(ids, id)
	}

	// Verify IDs are strictly increasing
	for i := 1; i < len(ids); i++ {
		assert.Greater(t, ids[i], ids[i-1], "ID %d should be greater than ID %d", ids[i], ids[i-1])
	}

	// Verify all 5 rows exist
	var count int
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users WHERE username LIKE 'user_%'").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 5, count)
}

func TestInsertReturningID_ErrorOnBadSQL(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()

	ctx := context.Background()
	err := db.RunMigrations(ctx)
	require.NoError(t, err)

	// Insert into non-existent table
	_, err = db.InsertReturningID(ctx, "INSERT INTO nonexistent_table (col) VALUES (?)", "val")
	assert.Error(t, err)
}

func TestInsertReturningID_WithNullableFields(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()

	ctx := context.Background()
	err := db.RunMigrations(ctx)
	require.NoError(t, err)

	// Insert a storage root — different table to diversify
	id, err := db.InsertReturningID(ctx,
		`INSERT INTO storage_roots (name, path, protocol, enabled, created_at, updated_at)
		 VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
		"test_root", "/tmp/test", "local", true)
	require.NoError(t, err)
	assert.Greater(t, id, int64(0))

	// Verify
	var name string
	err = db.QueryRowContext(ctx, "SELECT name FROM storage_roots WHERE id = ?", id).Scan(&name)
	require.NoError(t, err)
	assert.Equal(t, "test_root", name)
}

// ---------------------------------------------------------------------------
// TxInsertReturningID — multiple inserts in one transaction
// ---------------------------------------------------------------------------

func TestTxInsertReturningID_MultipleInSameTransaction(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()

	ctx := context.Background()
	err := db.RunMigrations(ctx)
	require.NoError(t, err)

	tx, err := db.Begin()
	require.NoError(t, err)

	id1, err := db.TxInsertReturningID(ctx, tx,
		`INSERT INTO users (username, email, password_hash, salt, role_id, is_active, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
		"txuser1", "tx1@example.com", "hash1", "salt1", 1, true)
	require.NoError(t, err)

	id2, err := db.TxInsertReturningID(ctx, tx,
		`INSERT INTO users (username, email, password_hash, salt, role_id, is_active, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
		"txuser2", "tx2@example.com", "hash2", "salt2", 1, true)
	require.NoError(t, err)

	assert.Greater(t, id2, id1)

	err = tx.Commit()
	require.NoError(t, err)

	// Verify both persisted
	var count int
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users WHERE username IN ('txuser1', 'txuser2')").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 2, count)
}

func TestTxInsertReturningID_ErrorOnBadSQL(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()

	ctx := context.Background()
	err := db.RunMigrations(ctx)
	require.NoError(t, err)

	tx, err := db.Begin()
	require.NoError(t, err)

	_, err = db.TxInsertReturningID(ctx, tx,
		"INSERT INTO nonexistent_table (col) VALUES (?)", "val")
	assert.Error(t, err)

	// Rollback after error
	err = tx.Rollback()
	assert.NoError(t, err)
}

// ---------------------------------------------------------------------------
// ExecContext / QueryContext / QueryRowContext
// ---------------------------------------------------------------------------

func TestExecContext_InsertAndUpdate(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()

	ctx := context.Background()
	err := db.RunMigrations(ctx)
	require.NoError(t, err)

	// Insert via ExecContext
	result, err := db.ExecContext(ctx,
		"INSERT INTO storage_roots (name, path, protocol, enabled, created_at, updated_at) VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)",
		"exec_root", "/tmp/exec", "local", true)
	require.NoError(t, err)

	rowsAffected, err := result.RowsAffected()
	require.NoError(t, err)
	assert.Equal(t, int64(1), rowsAffected)

	// Update via ExecContext
	result, err = db.ExecContext(ctx,
		"UPDATE storage_roots SET name = ? WHERE path = ?",
		"updated_root", "/tmp/exec")
	require.NoError(t, err)

	rowsAffected, err = result.RowsAffected()
	require.NoError(t, err)
	assert.Equal(t, int64(1), rowsAffected)

	// Verify via QueryRowContext
	var name string
	err = db.QueryRowContext(ctx, "SELECT name FROM storage_roots WHERE path = ?", "/tmp/exec").Scan(&name)
	require.NoError(t, err)
	assert.Equal(t, "updated_root", name)
}

func TestQueryContext_MultipleRows(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()

	ctx := context.Background()
	err := db.RunMigrations(ctx)
	require.NoError(t, err)

	// Insert multiple storage roots
	for i := 0; i < 3; i++ {
		_, err := db.ExecContext(ctx,
			"INSERT INTO storage_roots (name, path, protocol, enabled, created_at, updated_at) VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)",
			"root_"+string(rune('0'+i)), "/tmp/root_"+string(rune('0'+i)), "local", true)
		require.NoError(t, err)
	}

	rows, err := db.QueryContext(ctx, "SELECT name FROM storage_roots WHERE name LIKE 'root_%' ORDER BY name")
	require.NoError(t, err)
	defer rows.Close()

	var names []string
	for rows.Next() {
		var name string
		require.NoError(t, rows.Scan(&name))
		names = append(names, name)
	}
	require.NoError(t, rows.Err())
	assert.Equal(t, 3, len(names))
}

// ---------------------------------------------------------------------------
// rewriteQuery — comprehensive Postgres rewriting
// ---------------------------------------------------------------------------

func TestRewriteQuery_PostgresMultipleBooleans(t *testing.T) {
	pgDB := &DB{dialect: Dialect{Type: DialectPostgres}, config: &config.DatabaseConfig{}}

	// Multiple booleans in one query
	got := pgDB.rewriteQuery("SELECT * FROM t WHERE is_active = 1 AND deleted = 0 AND is_locked = 1 AND id = ?")
	assert.Equal(t, "SELECT * FROM t WHERE is_active = TRUE AND deleted = FALSE AND is_locked = TRUE AND id = $1", got)
}

func TestRewriteQuery_SQLiteNoRewriting(t *testing.T) {
	sqliteDB := &DB{dialect: Dialect{Type: DialectSQLite}, config: &config.DatabaseConfig{}}

	input := "INSERT OR IGNORE INTO t (a) VALUES (?) WHERE is_active = 1"
	got := sqliteDB.rewriteQuery(input)
	assert.Equal(t, input, got, "SQLite should not rewrite queries")
}

func TestRewriteQuery_PostgresCombinedTransforms(t *testing.T) {
	pgDB := &DB{dialect: Dialect{Type: DialectPostgres}, config: &config.DatabaseConfig{}}

	// INSERT OR IGNORE + placeholder rewriting + boolean rewriting
	got := pgDB.rewriteQuery("INSERT OR IGNORE INTO users (name, is_active) VALUES (?, ?)")
	assert.Contains(t, got, "INSERT INTO")
	assert.Contains(t, got, "$1")
	assert.Contains(t, got, "$2")
	assert.Contains(t, got, "ON CONFLICT DO NOTHING")
	assert.NotContains(t, got, "INSERT OR IGNORE")
}

func TestRewriteQuery_PostgresReplaceTransform(t *testing.T) {
	pgDB := &DB{dialect: Dialect{Type: DialectPostgres}, config: &config.DatabaseConfig{}}

	got := pgDB.rewriteQuery("INSERT OR REPLACE INTO cache (key, value) VALUES (?, ?)")
	assert.Contains(t, got, "INSERT INTO")
	assert.Contains(t, got, "$1")
	assert.Contains(t, got, "$2")
	assert.NotContains(t, got, "INSERT OR REPLACE")
}

// ---------------------------------------------------------------------------
// TableExists — edge cases
// ---------------------------------------------------------------------------

func TestTableExists_EmptyTableName(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()

	ctx := context.Background()
	err := db.RunMigrations(ctx)
	require.NoError(t, err)

	exists, err := db.TableExists(ctx, "")
	assert.NoError(t, err)
	assert.False(t, exists)
}

func TestTableExists_AfterDroppingTable(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create a temporary table
	_, err := db.ExecContext(ctx, "CREATE TABLE temp_test_table (id INTEGER PRIMARY KEY, name TEXT)")
	require.NoError(t, err)

	exists, err := db.TableExists(ctx, "temp_test_table")
	require.NoError(t, err)
	assert.True(t, exists)

	// Drop it
	_, err = db.ExecContext(ctx, "DROP TABLE temp_test_table")
	require.NoError(t, err)

	exists, err = db.TableExists(ctx, "temp_test_table")
	require.NoError(t, err)
	assert.False(t, exists)
}

// ---------------------------------------------------------------------------
// WrapDB — additional scenarios
// ---------------------------------------------------------------------------

func TestWrapDB_ConfigIsNonNil(t *testing.T) {
	rawDB, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer rawDB.Close()

	wrapped := WrapDB(rawDB, DialectSQLite)
	require.NotNil(t, wrapped)
	assert.NotNil(t, wrapped.config, "WrapDB should set a non-nil config")
}

func TestWrapDB_PostgresDialect(t *testing.T) {
	rawDB, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer rawDB.Close()

	wrapped := WrapDB(rawDB, DialectPostgres)
	require.NotNil(t, wrapped)
	assert.True(t, wrapped.Dialect().IsPostgres())
	assert.Equal(t, "postgres", wrapped.DatabaseType())
}

// ---------------------------------------------------------------------------
// HealthCheck / GetStats / DatabaseType
// ---------------------------------------------------------------------------

func TestHealthCheck_OnLiveDB(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()

	err := db.HealthCheck()
	assert.NoError(t, err)
}

func TestGetStats_ReturnsValidStats(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()

	stats := db.GetStats()
	assert.GreaterOrEqual(t, stats.OpenConnections, 0)
}

func TestDatabaseType_SQLiteAndPostgres(t *testing.T) {
	sqliteDB := &DB{dialect: Dialect{Type: DialectSQLite}, config: &config.DatabaseConfig{}}
	assert.Equal(t, "sqlite", sqliteDB.DatabaseType())

	pgDB := &DB{dialect: Dialect{Type: DialectPostgres}, config: &config.DatabaseConfig{}}
	assert.Equal(t, "postgres", pgDB.DatabaseType())
}

// ---------------------------------------------------------------------------
// NewConnection — connection pool settings
// ---------------------------------------------------------------------------

func TestNewConnection_WithCustomPoolSettings(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "pool_test_*.db")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	cfg := &config.DatabaseConfig{
		Path:               tmpFile.Name(),
		MaxOpenConnections: 20,
		MaxIdleConnections: 10,
		ConnMaxLifetime:    7200,
		ConnMaxIdleTime:    3600,
		EnableWAL:          true,
		CacheSize:          5000,
		BusyTimeout:        10000,
	}

	db, err := NewConnection(cfg)
	require.NoError(t, err)
	defer db.Close()

	err = db.Ping()
	assert.NoError(t, err)
	assert.True(t, db.Dialect().IsSQLite())
}

func TestNewConnection_WithoutWALOrCacheSize(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "nowal_test_*.db")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	cfg := &config.DatabaseConfig{
		Path:        tmpFile.Name(),
		EnableWAL:   false,
		CacheSize:   0,
		BusyTimeout: 5000,
	}

	db, err := NewConnection(cfg)
	require.NoError(t, err)
	defer db.Close()

	err = db.Ping()
	assert.NoError(t, err)
}

// ---------------------------------------------------------------------------
// Dialect helper: RewriteInsertOrReplace — additional edge cases
// ---------------------------------------------------------------------------

func TestDialect_RewriteInsertOrReplace_WithPrefix(t *testing.T) {
	pg := &Dialect{Type: DialectPostgres}

	// Leading whitespace
	got := pg.RewriteInsertOrReplace("  INSERT OR REPLACE INTO t (a) VALUES (1)")
	assert.Equal(t, "  INSERT INTO t (a) VALUES (1)", got)
}

func TestDialect_RewriteInsertOrIgnore_MixedCase(t *testing.T) {
	pg := &Dialect{Type: DialectPostgres}

	got := pg.RewriteInsertOrIgnore("Insert Or Ignore Into t (a) VALUES (1)")
	assert.Contains(t, got, "INSERT INTO")
	assert.Contains(t, got, "ON CONFLICT DO NOTHING")
}

// ---------------------------------------------------------------------------
// Dialect helper: BooleanDefault edge cases
// ---------------------------------------------------------------------------

func TestDialect_BooleanDefault_AllDialects(t *testing.T) {
	for _, dt := range []DialectType{DialectSQLite, DialectPostgres} {
		d := &Dialect{Type: dt}
		trueVal := d.BooleanDefault(true)
		falseVal := d.BooleanDefault(false)
		assert.NotEqual(t, trueVal, falseVal, "true and false defaults must differ for %s", dt)
	}
}

// ---------------------------------------------------------------------------
// Dialect helper: AutoIncrement
// ---------------------------------------------------------------------------

func TestDialect_AutoIncrement_SQLitePrimaryKey(t *testing.T) {
	d := &Dialect{Type: DialectSQLite}
	ai := d.AutoIncrement()
	assert.Contains(t, ai, "INTEGER PRIMARY KEY")
	assert.Contains(t, ai, "AUTOINCREMENT")
}

func TestDialect_AutoIncrement_PostgresSerial(t *testing.T) {
	d := &Dialect{Type: DialectPostgres}
	ai := d.AutoIncrement()
	assert.Contains(t, ai, "SERIAL PRIMARY KEY")
}

// ---------------------------------------------------------------------------
// Dialect helper: RewriteBooleanLiterals — multiple booleans in one query
// ---------------------------------------------------------------------------

func TestDialect_RewriteBooleanLiterals_MultipleBooleans(t *testing.T) {
	pg := &Dialect{Type: DialectPostgres}

	input := "SELECT * FROM t WHERE is_active = 1 AND is_locked = 0 AND enabled = 1 AND deleted = 0"
	got := pg.RewriteBooleanLiterals(input)

	assert.Contains(t, got, "is_active = TRUE")
	assert.Contains(t, got, "is_locked = FALSE")
	assert.Contains(t, got, "enabled = TRUE")
	assert.Contains(t, got, "deleted = FALSE")
}

func TestDialect_RewriteBooleanLiterals_SQLitePassthrough(t *testing.T) {
	sq := &Dialect{Type: DialectSQLite}

	input := "SELECT * FROM t WHERE is_active = 1 AND deleted = 0"
	got := sq.RewriteBooleanLiterals(input)
	assert.Equal(t, input, got, "SQLite should return query unchanged")
}

// ---------------------------------------------------------------------------
// Exec / Query / QueryRow (non-context) with INSERT OR IGNORE on SQLite
// ---------------------------------------------------------------------------

func TestShadowedExec_InsertOrIgnore(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()

	ctx := context.Background()
	err := db.RunMigrations(ctx)
	require.NoError(t, err)

	// Insert a role
	_, err = db.Exec("INSERT OR IGNORE INTO roles (name, description, permissions, is_system) VALUES (?, ?, ?, ?)",
		"shadow_test_role", "shadow test", "[]", false)
	assert.NoError(t, err)

	// Duplicate insert should be ignored
	_, err = db.Exec("INSERT OR IGNORE INTO roles (name, description, permissions, is_system) VALUES (?, ?, ?, ?)",
		"shadow_test_role", "shadow test", "[]", false)
	assert.NoError(t, err)

	// Verify only 1
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM roles WHERE name = ?", "shadow_test_role").Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestShadowedQuery_NoRows(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()

	ctx := context.Background()
	err := db.RunMigrations(ctx)
	require.NoError(t, err)

	rows, err := db.Query("SELECT name FROM roles WHERE name = ?", "nonexistent_role_xyz")
	require.NoError(t, err)
	defer rows.Close()

	found := false
	for rows.Next() {
		found = true
	}
	assert.False(t, found)
}

func TestShadowedQueryRow_NoRows(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()

	ctx := context.Background()
	err := db.RunMigrations(ctx)
	require.NoError(t, err)

	var name string
	err = db.QueryRow("SELECT name FROM roles WHERE name = ?", "nonexistent_role_xyz").Scan(&name)
	assert.Error(t, err)
	assert.Equal(t, sql.ErrNoRows, err)
}

// ---------------------------------------------------------------------------
// createContext edge cases
// ---------------------------------------------------------------------------

func TestCreateContext_NegativeBusyTimeout(t *testing.T) {
	db := &DB{
		config:  &config.DatabaseConfig{BusyTimeout: -100},
		dialect: Dialect{Type: DialectSQLite},
	}

	ctx, cancel := db.createContext()
	defer cancel()

	// Should default to 5s when BusyTimeout is negative
	deadline, ok := ctx.Deadline()
	assert.True(t, ok)
	assert.False(t, deadline.IsZero())
}
