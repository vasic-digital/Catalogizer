package database

import (
	"context"
	"os"
	"testing"

	"catalogizer/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestDB creates a temporary database for migration testing
func newTestDB(t *testing.T) (*DB, func()) {
	tmpFile, err := os.CreateTemp("", "migrations_test_*.db")
	require.NoError(t, err)
	tmpFile.Close()

	cfg := &config.DatabaseConfig{
		Path:               tmpFile.Name(),
		MaxOpenConnections: 1,
		MaxIdleConnections: 1,
		ConnMaxLifetime:    3600,
		ConnMaxIdleTime:    1800,
		EnableWAL:          true,
		CacheSize:          2000,
		BusyTimeout:        5000,
	}

	db, err := NewConnection(cfg)
	require.NoError(t, err)

	cleanup := func() {
		db.Close()
		os.Remove(tmpFile.Name())
	}

	return db, cleanup
}

// ---------------------------------------------------------------------------
// RunMigrations
// ---------------------------------------------------------------------------

func TestRunMigrations_Success(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()

	ctx := context.Background()
	err := db.RunMigrations(ctx)
	assert.NoError(t, err)
}

func TestRunMigrations_Idempotent(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Run migrations twice; second run should be a no-op
	err := db.RunMigrations(ctx)
	require.NoError(t, err)

	err = db.RunMigrations(ctx)
	assert.NoError(t, err)
}

// ---------------------------------------------------------------------------
// createMigrationsTable
// ---------------------------------------------------------------------------

func TestCreateMigrationsTable(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()

	ctx := context.Background()
	err := db.createMigrationsTable(ctx)
	require.NoError(t, err)

	// Verify table exists by querying it
	var count int
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM migrations").Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestCreateMigrationsTable_Idempotent(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()

	ctx := context.Background()

	err := db.createMigrationsTable(ctx)
	require.NoError(t, err)

	// Call again; should succeed due to IF NOT EXISTS
	err = db.createMigrationsTable(ctx)
	assert.NoError(t, err)
}

// ---------------------------------------------------------------------------
// runMigration
// ---------------------------------------------------------------------------

func TestRunMigration_AppliesOnce(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create migrations table first
	err := db.createMigrationsTable(ctx)
	require.NoError(t, err)

	callCount := 0
	migration := Migration{
		Version: 100,
		Name:    "test_migration",
		Up: func(ctx context.Context) error {
			callCount++
			return nil
		},
	}

	// First run should apply
	err = db.runMigration(ctx, migration)
	require.NoError(t, err)
	assert.Equal(t, 1, callCount)

	// Second run should skip (already applied)
	err = db.runMigration(ctx, migration)
	require.NoError(t, err)
	assert.Equal(t, 1, callCount)
}

func TestRunMigration_RecordsMigration(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()

	ctx := context.Background()
	err := db.createMigrationsTable(ctx)
	require.NoError(t, err)

	migration := Migration{
		Version: 200,
		Name:    "recorded_migration",
		Up: func(ctx context.Context) error {
			return nil
		},
	}

	err = db.runMigration(ctx, migration)
	require.NoError(t, err)

	// Verify migration was recorded
	var count int
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM migrations WHERE version = ?", 200).Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)

	var name string
	err = db.QueryRowContext(ctx, "SELECT name FROM migrations WHERE version = ?", 200).Scan(&name)
	assert.NoError(t, err)
	assert.Equal(t, "recorded_migration", name)
}

// ---------------------------------------------------------------------------
// createInitialTables
// ---------------------------------------------------------------------------

func TestCreateInitialTables(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()

	ctx := context.Background()
	err := db.createInitialTables(ctx)
	require.NoError(t, err)

	// Verify core tables exist
	tables := []string{"storage_roots", "files", "file_metadata", "duplicate_groups", "virtual_paths", "scan_history"}
	for _, table := range tables {
		var count int
		err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&count)
		assert.NoError(t, err, "checking table %s", table)
		assert.Equal(t, 1, count, "table %s should exist", table)
	}
}

func TestCreateInitialTables_Idempotent(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()

	ctx := context.Background()

	err := db.createInitialTables(ctx)
	require.NoError(t, err)

	// Call again; should succeed due to IF NOT EXISTS
	err = db.createInitialTables(ctx)
	assert.NoError(t, err)
}

// ---------------------------------------------------------------------------
// createConversionJobsTable
// ---------------------------------------------------------------------------

func TestCreateConversionJobsTable(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()

	// Need auth tables first for the foreign key reference (users table)
	ctx := context.Background()
	err := db.createAuthTables(ctx)
	require.NoError(t, err)

	err = db.createConversionJobsTable(ctx)
	require.NoError(t, err)

	// Verify table exists
	var count int
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='conversion_jobs'").Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)

	// Verify indexes exist
	var indexCount int
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM sqlite_master WHERE type='index' AND name LIKE 'idx_conversion_jobs_%'").Scan(&indexCount)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, indexCount, 3)
}

// ---------------------------------------------------------------------------
// createAuthTables
// ---------------------------------------------------------------------------

func TestCreateAuthTables(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()

	ctx := context.Background()
	err := db.createAuthTables(ctx)
	require.NoError(t, err)

	// Verify auth tables exist
	authTables := []string{"users", "roles", "user_sessions", "permissions", "user_permissions", "auth_audit_log"}
	for _, table := range authTables {
		var count int
		err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&count)
		assert.NoError(t, err, "checking table %s", table)
		assert.Equal(t, 1, count, "table %s should exist", table)
	}

	// Verify default roles were inserted
	var roleCount int
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM roles").Scan(&roleCount)
	assert.NoError(t, err)
	assert.Equal(t, 2, roleCount)
}

// ---------------------------------------------------------------------------
// MigrationSequence
// ---------------------------------------------------------------------------

func TestMigrationSequence_AllVersionsApplied(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()

	ctx := context.Background()
	err := db.RunMigrations(ctx)
	require.NoError(t, err)

	// Verify all 6 migrations were recorded
	var count int
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM migrations").Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 6, count)

	// Verify each version exists
	for v := 1; v <= 6; v++ {
		var exists int
		err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM migrations WHERE version = ?", v).Scan(&exists)
		assert.NoError(t, err)
		assert.Equal(t, 1, exists, "migration version %d should be recorded", v)
	}
}

// ---------------------------------------------------------------------------
// createSubtitleTables
// ---------------------------------------------------------------------------

func TestCreateSubtitleTables(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()

	// Subtitle tables need media_items reference which doesn't exist standalone,
	// but the tables use IF NOT EXISTS and the FK is deferred in SQLite by default.
	// We can still verify the tables get created.
	ctx := context.Background()
	err := db.createSubtitleTables(ctx)
	require.NoError(t, err)

	// Verify subtitle tables exist
	subtitleTables := []string{"subtitle_tracks", "subtitle_sync_status", "subtitle_cache", "subtitle_downloads", "media_subtitles"}
	for _, table := range subtitleTables {
		var count int
		err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&count)
		assert.NoError(t, err, "checking table %s", table)
		assert.Equal(t, 1, count, "table %s should exist", table)
	}
}

// ---------------------------------------------------------------------------
// MigrationStruct
// ---------------------------------------------------------------------------

func TestMigrationStruct(t *testing.T) {
	m := Migration{
		Version: 42,
		Name:    "test_migration",
		Up: func(ctx context.Context) error {
			return nil
		},
	}

	assert.Equal(t, 42, m.Version)
	assert.Equal(t, "test_migration", m.Name)
	assert.NotNil(t, m.Up)

	// Verify the Up function executes without error
	err := m.Up(context.Background())
	assert.NoError(t, err)
}
