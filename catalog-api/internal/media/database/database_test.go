package database

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// newTestMediaDatabase creates a MediaDatabase with a sqlmock backend for testing.
func newTestMediaDatabase(t *testing.T) (*MediaDatabase, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	logger := zap.NewNop()
	mdb := &MediaDatabase{
		db:       db,
		dbPath:   "test.db",
		password: "test-password",
		logger:   logger,
	}
	return mdb, mock
}

// ---------------------------------------------------------------------------
// DatabaseConfig validation and NewMediaDatabase constructor
// ---------------------------------------------------------------------------

func TestDatabaseConfig_FieldsExported(t *testing.T) {
	cfg := DatabaseConfig{
		Path:     "/tmp/test.db",
		Password: "secret",
	}
	assert.Equal(t, "/tmp/test.db", cfg.Path)
	assert.Equal(t, "secret", cfg.Password)
}

func TestNewMediaDatabase_EmptyPasswordReturnsError(t *testing.T) {
	tests := []struct {
		name   string
		config DatabaseConfig
	}{
		{
			name:   "empty password with path",
			config: DatabaseConfig{Path: "/tmp/some.db", Password: ""},
		},
		{
			name:   "empty password without path",
			config: DatabaseConfig{Path: "", Password: ""},
		},
	}

	logger := zap.NewNop()
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mdb, err := NewMediaDatabase(tc.config, logger)
			assert.Nil(t, mdb)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "database password is required for encryption")
		})
	}
}

func TestNewMediaDatabase_DefaultPathWhenEmpty(t *testing.T) {
	// When Path is empty, the constructor defaults it to "media_catalog.db".
	// With SQLCipher available, connect() may succeed (creating a file), but
	// initialize() will use the embedded schema. Either way, we verify
	// the error does NOT mention "password is required" -- path defaulting works.
	logger := zap.NewNop()
	cfg := DatabaseConfig{Path: "", Password: "some-password"}
	mdb, err := NewMediaDatabase(cfg, logger)
	if err != nil {
		// Expected to fail at connect or initialize, but NOT at password validation.
		assert.NotContains(t, err.Error(), "database password is required")
	} else {
		// If it somehow succeeded, clean up.
		defer mdb.Close()
		defer os.Remove("media_catalog.db")
		assert.NotNil(t, mdb)
	}
}

func TestNewMediaDatabase_InvalidPathReturnsConnectError(t *testing.T) {
	logger := zap.NewNop()
	cfg := DatabaseConfig{
		Path:     "/nonexistent/deeply/nested/path/test.db",
		Password: "password123",
	}
	_, err := NewMediaDatabase(cfg, logger)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to connect to database")
}

// ---------------------------------------------------------------------------
// GetDB
// ---------------------------------------------------------------------------

func TestGetDB_ReturnsUnderlyingConnection(t *testing.T) {
	mdb, _ := newTestMediaDatabase(t)
	defer mdb.db.Close()

	db := mdb.GetDB()
	assert.NotNil(t, db)
	assert.Equal(t, mdb.db, db)
}

func TestGetDB_ReturnsNilWhenDBIsNil(t *testing.T) {
	mdb := &MediaDatabase{
		logger: zap.NewNop(),
	}
	assert.Nil(t, mdb.GetDB())
}

// ---------------------------------------------------------------------------
// Close
// ---------------------------------------------------------------------------

func TestClose_ClosesConnection(t *testing.T) {
	mdb, mock := newTestMediaDatabase(t)

	mock.ExpectClose()

	err := mdb.Close()
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestClose_NilDBReturnsNil(t *testing.T) {
	mdb := &MediaDatabase{
		logger: zap.NewNop(),
	}
	err := mdb.Close()
	assert.NoError(t, err)
}

func TestClose_CalledTwice_SecondCallStillNilCheck(t *testing.T) {
	mdb, mock := newTestMediaDatabase(t)

	mock.ExpectClose()
	err := mdb.Close()
	assert.NoError(t, err)

	// After Close, mdb.db is still non-nil (Close doesn't nil the field).
	// Manually set to nil to test the nil guard path.
	mdb.db = nil
	err = mdb.Close()
	assert.NoError(t, err, "closing with nil db should return nil")
}

// ---------------------------------------------------------------------------
// HealthCheck
// ---------------------------------------------------------------------------

func TestHealthCheck_Success(t *testing.T) {
	mdb, mock := newTestMediaDatabase(t)
	defer mdb.db.Close()

	mock.ExpectQuery("PRAGMA integrity_check").
		WillReturnRows(sqlmock.NewRows([]string{"integrity_check"}).AddRow("ok"))

	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM media_types").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(7))

	err := mdb.HealthCheck()
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHealthCheck_IntegrityCheckFails(t *testing.T) {
	mdb, mock := newTestMediaDatabase(t)
	defer mdb.db.Close()

	mock.ExpectQuery("PRAGMA integrity_check").
		WillReturnError(fmt.Errorf("disk I/O error"))

	err := mdb.HealthCheck()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "integrity check failed")
}

func TestHealthCheck_IntegrityNotOK(t *testing.T) {
	mdb, mock := newTestMediaDatabase(t)
	defer mdb.db.Close()

	mock.ExpectQuery("PRAGMA integrity_check").
		WillReturnRows(sqlmock.NewRows([]string{"integrity_check"}).AddRow("page 5: corrupt"))

	err := mdb.HealthCheck()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database integrity check failed")
	assert.Contains(t, err.Error(), "page 5: corrupt")
}

func TestHealthCheck_MediaTypesQueryFails(t *testing.T) {
	mdb, mock := newTestMediaDatabase(t)
	defer mdb.db.Close()

	mock.ExpectQuery("PRAGMA integrity_check").
		WillReturnRows(sqlmock.NewRows([]string{"integrity_check"}).AddRow("ok"))

	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM media_types").
		WillReturnError(fmt.Errorf("no such table: media_types"))

	err := mdb.HealthCheck()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to query media_types")
}

// ---------------------------------------------------------------------------
// GetStats
// ---------------------------------------------------------------------------

func TestGetStats_Success(t *testing.T) {
	mdb, mock := newTestMediaDatabase(t)
	defer mdb.db.Close()

	// Page count and page size
	mock.ExpectQuery("PRAGMA page_count").
		WillReturnRows(sqlmock.NewRows([]string{"page_count"}).AddRow(100))
	mock.ExpectQuery("PRAGMA page_size").
		WillReturnRows(sqlmock.NewRows([]string{"page_size"}).AddRow(4096))

	// Table counts -- the order of iteration over a map is non-deterministic,
	// so we use MatchExpectationsInOrder(false).
	mock.MatchExpectationsInOrder(false)

	tables := []string{
		"media_types", "media_items", "external_metadata",
		"directory_analysis", "media_files", "media_collections", "user_metadata",
	}
	for _, table := range tables {
		mock.ExpectQuery(fmt.Sprintf("SELECT COUNT\\(\\*\\) FROM %s", table)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(10))
	}

	// Recent activity
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM directory_analysis WHERE last_analyzed").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))

	stats, err := mdb.GetStats()
	assert.NoError(t, err)
	assert.NotNil(t, stats)

	// Verify size calculation
	assert.Equal(t, int64(100*4096), stats["size_bytes"])

	// Verify table counts map exists
	tableCounts, ok := stats["table_counts"].(map[string]int64)
	assert.True(t, ok)
	for _, table := range tables {
		assert.Equal(t, int64(10), tableCounts[table])
	}

	// Verify recent activity
	assert.Equal(t, int64(3), stats["recent_analysis_24h"])
}

func TestGetStats_PageCountFails(t *testing.T) {
	mdb, mock := newTestMediaDatabase(t)
	defer mdb.db.Close()

	mock.ExpectQuery("PRAGMA page_count").
		WillReturnError(fmt.Errorf("database locked"))

	stats, err := mdb.GetStats()
	assert.Error(t, err)
	assert.Nil(t, stats)
}

func TestGetStats_PageSizeFails(t *testing.T) {
	mdb, mock := newTestMediaDatabase(t)
	defer mdb.db.Close()

	mock.ExpectQuery("PRAGMA page_count").
		WillReturnRows(sqlmock.NewRows([]string{"page_count"}).AddRow(100))
	mock.ExpectQuery("PRAGMA page_size").
		WillReturnError(fmt.Errorf("database locked"))

	stats, err := mdb.GetStats()
	assert.Error(t, err)
	assert.Nil(t, stats)
}

func TestGetStats_TableCountErrorContinues(t *testing.T) {
	mdb, mock := newTestMediaDatabase(t)
	defer mdb.db.Close()

	mock.ExpectQuery("PRAGMA page_count").
		WillReturnRows(sqlmock.NewRows([]string{"page_count"}).AddRow(50))
	mock.ExpectQuery("PRAGMA page_size").
		WillReturnRows(sqlmock.NewRows([]string{"page_size"}).AddRow(4096))

	mock.MatchExpectationsInOrder(false)

	// All table counts fail
	tables := []string{
		"media_types", "media_items", "external_metadata",
		"directory_analysis", "media_files", "media_collections", "user_metadata",
	}
	for _, table := range tables {
		mock.ExpectQuery(fmt.Sprintf("SELECT COUNT\\(\\*\\) FROM %s", table)).
			WillReturnError(fmt.Errorf("no such table"))
	}

	// Recent activity also fails -- that's fine, it's handled with if err == nil
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM directory_analysis WHERE last_analyzed").
		WillReturnError(fmt.Errorf("no such table"))

	stats, err := mdb.GetStats()
	assert.NoError(t, err)
	assert.NotNil(t, stats)

	// All table counts should remain at 0 (the default)
	tableCounts, ok := stats["table_counts"].(map[string]int64)
	assert.True(t, ok)
	for _, table := range tables {
		assert.Equal(t, int64(0), tableCounts[table])
	}
}

// ---------------------------------------------------------------------------
// ExecuteInTransaction
// ---------------------------------------------------------------------------

func TestExecuteInTransaction_Success(t *testing.T) {
	mdb, mock := newTestMediaDatabase(t)
	defer mdb.db.Close()

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO media_types").
		WithArgs("test_type", "A test type").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := mdb.ExecuteInTransaction(func(tx *sql.Tx) error {
		_, err := tx.Exec("INSERT INTO media_types (name, description) VALUES (?, ?)", "test_type", "A test type")
		return err
	})
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestExecuteInTransaction_FnReturnsError(t *testing.T) {
	mdb, mock := newTestMediaDatabase(t)
	defer mdb.db.Close()

	mock.ExpectBegin()
	mock.ExpectRollback()

	expectedErr := errors.New("something went wrong")
	err := mdb.ExecuteInTransaction(func(tx *sql.Tx) error {
		return expectedErr
	})
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestExecuteInTransaction_BeginFails(t *testing.T) {
	mdb, mock := newTestMediaDatabase(t)
	defer mdb.db.Close()

	mock.ExpectBegin().WillReturnError(fmt.Errorf("connection refused"))

	err := mdb.ExecuteInTransaction(func(tx *sql.Tx) error {
		return nil
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to begin transaction")
}

func TestExecuteInTransaction_CommitFails(t *testing.T) {
	mdb, mock := newTestMediaDatabase(t)
	defer mdb.db.Close()

	mock.ExpectBegin()
	mock.ExpectCommit().WillReturnError(fmt.Errorf("disk full"))

	err := mdb.ExecuteInTransaction(func(tx *sql.Tx) error {
		return nil
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "disk full")
}

func TestExecuteInTransaction_MultipleStatements(t *testing.T) {
	mdb, mock := newTestMediaDatabase(t)
	defer mdb.db.Close()

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO media_types").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO media_items").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := mdb.ExecuteInTransaction(func(tx *sql.Tx) error {
		if _, err := tx.Exec("INSERT INTO media_types (name) VALUES (?)", "movie"); err != nil {
			return err
		}
		if _, err := tx.Exec("INSERT INTO media_items (title) VALUES (?)", "Test Movie"); err != nil {
			return err
		}
		return nil
	})
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ---------------------------------------------------------------------------
// createSchemaFromString
// ---------------------------------------------------------------------------

func TestCreateSchemaFromString_Success(t *testing.T) {
	mdb, mock := newTestMediaDatabase(t)
	defer mdb.db.Close()

	schema := "CREATE TABLE IF NOT EXISTS test_table (id INTEGER PRIMARY KEY)"

	mock.ExpectBegin()
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS test_table").
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	err := mdb.createSchemaFromString(schema)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateSchemaFromString_BeginFails(t *testing.T) {
	mdb, mock := newTestMediaDatabase(t)
	defer mdb.db.Close()

	mock.ExpectBegin().WillReturnError(fmt.Errorf("database locked"))

	err := mdb.createSchemaFromString("CREATE TABLE test (id INT)")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to begin transaction")
}

func TestCreateSchemaFromString_ExecFails(t *testing.T) {
	mdb, mock := newTestMediaDatabase(t)
	defer mdb.db.Close()

	mock.ExpectBegin()
	mock.ExpectExec("CREATE TABLE").
		WillReturnError(fmt.Errorf("syntax error"))
	mock.ExpectRollback()

	err := mdb.createSchemaFromString("CREATE TABLE bad syntax")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to execute schema")
}

func TestCreateSchemaFromString_CommitFails(t *testing.T) {
	mdb, mock := newTestMediaDatabase(t)
	defer mdb.db.Close()

	mock.ExpectBegin()
	mock.ExpectExec("CREATE TABLE").
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit().WillReturnError(fmt.Errorf("write error"))

	err := mdb.createSchemaFromString("CREATE TABLE test (id INT)")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to commit schema")
}

// ---------------------------------------------------------------------------
// getEmbeddedSchema
// ---------------------------------------------------------------------------

func TestGetEmbeddedSchema_ContainsRequiredTables(t *testing.T) {
	schema := getEmbeddedSchema()

	requiredTables := []string{
		"media_types",
		"media_items",
		"external_metadata",
		"directory_analysis",
		"media_files",
	}

	for _, table := range requiredTables {
		t.Run(table, func(t *testing.T) {
			assert.Contains(t, schema, table,
				"embedded schema should contain table %s", table)
		})
	}
}

func TestGetEmbeddedSchema_ContainsDefaultMediaTypes(t *testing.T) {
	schema := getEmbeddedSchema()

	expectedTypes := []string{
		"movie", "tv_show", "music", "game", "software", "training", "other",
	}

	for _, mt := range expectedTypes {
		t.Run(mt, func(t *testing.T) {
			assert.Contains(t, schema, fmt.Sprintf("'%s'", mt),
				"embedded schema should seed media type %s", mt)
		})
	}
}

func TestGetEmbeddedSchema_NotEmpty(t *testing.T) {
	schema := getEmbeddedSchema()
	assert.NotEmpty(t, schema)
	assert.Contains(t, schema, "CREATE TABLE")
	assert.Contains(t, schema, "INSERT OR IGNORE")
}

// ---------------------------------------------------------------------------
// initialize (schema loading from file vs embedded)
// ---------------------------------------------------------------------------

func TestInitialize_UsesSchemaFileWhenPresent(t *testing.T) {
	tmpDir := t.TempDir()
	schemaFile := filepath.Join(tmpDir, "schema.sql")

	// Write a minimal schema file
	schemaSQL := "CREATE TABLE IF NOT EXISTS test_init (id INTEGER PRIMARY KEY)"
	err := os.WriteFile(schemaFile, []byte(schemaSQL), 0644)
	require.NoError(t, err)

	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	mdb := &MediaDatabase{
		db:       db,
		dbPath:   filepath.Join(tmpDir, "test.db"),
		password: "test",
		logger:   zap.NewNop(),
	}

	mock.ExpectBegin()
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS test_init").
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	err = mdb.initialize()
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestInitialize_FallsBackToEmbeddedSchema(t *testing.T) {
	tmpDir := t.TempDir()
	// No schema.sql file in tmpDir

	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	mdb := &MediaDatabase{
		db:       db,
		dbPath:   filepath.Join(tmpDir, "test.db"),
		password: "test",
		logger:   zap.NewNop(),
	}

	mock.ExpectBegin()
	// Embedded schema contains CREATE TABLE statements
	mock.ExpectExec("CREATE TABLE").
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	err = mdb.initialize()
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ---------------------------------------------------------------------------
// getTables
// ---------------------------------------------------------------------------

func TestGetTables_Success(t *testing.T) {
	mdb, mock := newTestMediaDatabase(t)
	defer mdb.db.Close()

	rows := sqlmock.NewRows([]string{"name"}).
		AddRow("media_types").
		AddRow("media_items").
		AddRow("media_files")

	mock.ExpectQuery("SELECT name FROM sqlite_master WHERE type='table'").
		WillReturnRows(rows)

	tables, err := mdb.getTables()
	assert.NoError(t, err)
	assert.Equal(t, []string{"media_types", "media_items", "media_files"}, tables)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetTables_Empty(t *testing.T) {
	mdb, mock := newTestMediaDatabase(t)
	defer mdb.db.Close()

	rows := sqlmock.NewRows([]string{"name"})
	mock.ExpectQuery("SELECT name FROM sqlite_master WHERE type='table'").
		WillReturnRows(rows)

	tables, err := mdb.getTables()
	assert.NoError(t, err)
	assert.Nil(t, tables)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetTables_QueryError(t *testing.T) {
	mdb, mock := newTestMediaDatabase(t)
	defer mdb.db.Close()

	mock.ExpectQuery("SELECT name FROM sqlite_master WHERE type='table'").
		WillReturnError(fmt.Errorf("database is locked"))

	tables, err := mdb.getTables()
	assert.Error(t, err)
	assert.Nil(t, tables)
}

// ---------------------------------------------------------------------------
// Vacuum
// ---------------------------------------------------------------------------

func TestVacuum_Success(t *testing.T) {
	mdb, mock := newTestMediaDatabase(t)
	defer mdb.db.Close()

	mock.ExpectExec("VACUUM").
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := mdb.Vacuum()
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestVacuum_Error(t *testing.T) {
	mdb, mock := newTestMediaDatabase(t)
	defer mdb.db.Close()

	mock.ExpectExec("VACUUM").
		WillReturnError(fmt.Errorf("database is locked"))

	err := mdb.Vacuum()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "vacuum failed")
}

// ---------------------------------------------------------------------------
// ChangePassword
// ---------------------------------------------------------------------------

func TestChangePassword_EmptyPasswordReturnsError(t *testing.T) {
	mdb, _ := newTestMediaDatabase(t)
	defer mdb.db.Close()

	err := mdb.ChangePassword("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "new password cannot be empty")
}

func TestChangePassword_Success(t *testing.T) {
	mdb, mock := newTestMediaDatabase(t)
	defer mdb.db.Close()

	mock.ExpectExec("PRAGMA rekey").
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := mdb.ChangePassword("new-secret")
	assert.NoError(t, err)
	assert.Equal(t, "new-secret", mdb.password)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestChangePassword_ExecFails(t *testing.T) {
	mdb, mock := newTestMediaDatabase(t)
	defer mdb.db.Close()

	mock.ExpectExec("PRAGMA rekey").
		WillReturnError(fmt.Errorf("encryption error"))

	originalPassword := mdb.password
	err := mdb.ChangePassword("new-secret")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to change password")
	// Password should not have changed on error
	assert.Equal(t, originalPassword, mdb.password)
}

// ---------------------------------------------------------------------------
// Backup
// ---------------------------------------------------------------------------

func TestBackup_AttachFails(t *testing.T) {
	mdb, mock := newTestMediaDatabase(t)
	defer mdb.db.Close()

	mock.ExpectExec("ATTACH DATABASE").
		WillReturnError(fmt.Errorf("permission denied"))

	err := mdb.Backup("/tmp/backup.db")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "backup failed")
}

// ---------------------------------------------------------------------------
// MediaDatabase struct field integrity
// ---------------------------------------------------------------------------

func TestMediaDatabase_FieldsSetCorrectly(t *testing.T) {
	logger := zap.NewNop()
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mdb := &MediaDatabase{
		db:       db,
		dbPath:   "/data/media.db",
		password: "encrypted",
		logger:   logger,
	}

	assert.Equal(t, "/data/media.db", mdb.dbPath)
	assert.Equal(t, "encrypted", mdb.password)
	assert.Equal(t, logger, mdb.logger)
	assert.Equal(t, db, mdb.db)
}

// ---------------------------------------------------------------------------
// connect (DSN construction verification)
// ---------------------------------------------------------------------------

func TestConnect_FailsWithInvalidDriver(t *testing.T) {
	// The connect method uses the "sqlite3" driver registered by go-sqlcipher.
	// If the driver is available but the path is unwritable, Ping should fail.
	mdb := &MediaDatabase{
		dbPath:   "/nonexistent/directory/that/does/not/exist/test.db",
		password: "test-password",
		logger:   zap.NewNop(),
	}

	err := mdb.connect()
	assert.Error(t, err)
}
