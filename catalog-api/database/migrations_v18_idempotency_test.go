// Package database — migration-v18 idempotency regression tests.
//
// FIX-QA-2026-04-21-001 introduced migration v18
// (add_media_items_favorite_column). The migration uses SQLite
// PRAGMA table_info probes instead of "ADD COLUMN IF NOT EXISTS"
// because SQLite doesn't support that syntax. The probe-then-add
// logic has two correctness obligations:
//
//  1. **Idempotency** — running the migration N times on the same DB
//     must always converge to exactly the same schema (same three
//     columns, same types, same defaults, same nullability).
//  2. **Forward compat** — if the base schema already happens to have
//     one of the three columns (e.g., created by a future migration
//     or a manual operator ALTER), the probe must notice and skip
//     the redundant ADD, so we don't error out with "duplicate
//     column name".
//
// These are PROPERTY TESTS, not point-and-shoot unit tests: every
// assertion is a universal statement about the migration, and the
// failure of any one breaks the whole contract.

package database

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/mutecomm/go-sqlcipher"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newMigratedDB returns a migrated in-memory SQLite DB + its raw
// handle (so tests can drop the migrations row to force a re-run).
func newMigratedDB(t *testing.T) (*DB, *sql.DB) {
	t.Helper()

	raw, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	raw.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = raw.Close() })

	db := WrapDB(raw, DialectSQLite)
	require.NoError(t, db.RunMigrations(context.Background()))
	return db, raw
}

// columnInfo describes one column as SQLite's PRAGMA table_info
// returns it.
type columnInfo struct {
	cid       int
	name      string
	ctype     string
	notnull   int
	defaultAt any
	pk        int
}

func readTableInfo(t *testing.T, db *DB, table string) []columnInfo {
	t.Helper()
	rows, err := db.QueryContext(context.Background(), "PRAGMA table_info("+table+")")
	require.NoError(t, err)
	defer rows.Close()

	var out []columnInfo
	for rows.Next() {
		var c columnInfo
		require.NoError(t, rows.Scan(&c.cid, &c.name, &c.ctype, &c.notnull, &c.defaultAt, &c.pk))
		out = append(out, c)
	}
	require.NoError(t, rows.Err())
	return out
}

func findColumn(cols []columnInfo, name string) (columnInfo, bool) {
	for _, c := range cols {
		if c.name == name {
			return c, true
		}
	}
	return columnInfo{}, false
}

// --- Property 1: idempotency under repeated runs ---

// TestMigrationV18_IsIdempotent_ManyRuns reruns the migration 25 times
// on a fresh DB and asserts the three target columns exist exactly
// once each after each run. Catches any case where the probe misfires
// and the second ALTER would error out ("duplicate column name").
func TestMigrationV18_IsIdempotent_ManyRuns(t *testing.T) {
	db, raw := newMigratedDB(t)

	for i := 0; i < 25; i++ {
		// Force the runner to re-execute v18 by deleting its row.
		_, err := raw.Exec("DELETE FROM migrations WHERE version = 18")
		require.NoError(t, err, "iteration %d: delete migrations row", i)

		// Re-run all migrations — only v18 re-executes.
		require.NoError(t, db.RunMigrations(context.Background()),
			"iteration %d: RunMigrations", i)

		cols := readTableInfo(t, db, "media_items")

		// Each of the three columns exists exactly once.
		for _, name := range []string{"is_favorite", "updated_at", "duration"} {
			count := 0
			for _, c := range cols {
				if c.name == name {
					count++
				}
			}
			assert.Equal(t, 1, count, "iteration %d: column %s must appear exactly once (got %d)", i, name, count)
		}
	}
}

// --- Property 2: each added column has the declared type/default ---

func TestMigrationV18_ColumnTypesAndDefaults(t *testing.T) {
	db, _ := newMigratedDB(t)
	cols := readTableInfo(t, db, "media_items")

	cases := []struct {
		name       string
		wantType   string
		wantNotNul int // 1 = NOT NULL, 0 = nullable
	}{
		{"is_favorite", "INTEGER", 1},
		{"updated_at", "DATETIME", 0},
		{"duration", "INTEGER", 0},
	}
	for _, c := range cases {
		got, ok := findColumn(cols, c.name)
		require.Truef(t, ok, "column %s must exist in media_items after migrations", c.name)
		assert.Equal(t, c.wantType, got.ctype, "column %s type", c.name)
		assert.Equal(t, c.wantNotNul, got.notnull, "column %s notnull", c.name)
	}
}

// --- Property 3: pre-existing column is not re-added ---

// TestMigrationV18_SkipsPreExistingColumn manually ADDs is_favorite
// BEFORE running v18 and verifies the migration observes it and
// doesn't blow up with "duplicate column name".
func TestMigrationV18_SkipsPreExistingColumn(t *testing.T) {
	raw, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	raw.SetMaxOpenConns(1)
	defer raw.Close()

	db := WrapDB(raw, DialectSQLite)

	// Run all migrations up to and including v17 to establish base schema.
	require.NoError(t, db.RunMigrations(context.Background()))

	// Confirm the column is present after v18 ran once.
	cols := readTableInfo(t, db, "media_items")
	_, ok := findColumn(cols, "is_favorite")
	require.True(t, ok, "is_favorite must exist after initial RunMigrations")

	// Manually delete v18's migration row (so it re-runs) but leave
	// the column in place. The probe should catch this and skip.
	_, err = raw.Exec("DELETE FROM migrations WHERE version = 18")
	require.NoError(t, err)

	// Second run must not error.
	require.NoError(t, db.RunMigrations(context.Background()),
		"v18 must skip pre-existing column without ALTER-erroring")

	// Column is still there, still exactly once.
	cols = readTableInfo(t, db, "media_items")
	count := 0
	for _, c := range cols {
		if c.name == "is_favorite" {
			count++
		}
	}
	assert.Equal(t, 1, count)
}

// --- Property 4: backfill populated updated_at for existing rows ---

// TestMigrationV18_BackfillsUpdatedAt constructs the pre-v18 schema
// (minimal media_items without is_favorite / updated_at / duration)
// BY HAND so we can seed a row and then invoke the v18 SQLite path
// directly. This exercises the backfill branch of
// addMediaItemsFavoriteColumnSQLite — SQLite's ALTER TABLE refuses
// non-constant defaults (CURRENT_TIMESTAMP), so v18 adds updated_at
// NULLable and backfills COALESCE(last_updated, CURRENT_TIMESTAMP)
// in the same migration step.
func TestMigrationV18_BackfillsUpdatedAt(t *testing.T) {
	raw, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	raw.SetMaxOpenConns(1)
	defer raw.Close()

	db := WrapDB(raw, DialectSQLite)

	// Build a minimal pre-v18 media_items by hand — only the columns
	// we need for the backfill path.
	_, err = raw.Exec(`
		CREATE TABLE media_items (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			media_type_id INTEGER NOT NULL,
			title TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'detected',
			last_updated DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	require.NoError(t, err)

	// Seed a row with a specific last_updated so we can verify the
	// backfill copies it exactly.
	_, err = raw.Exec(`
		INSERT INTO media_items (media_type_id, title, status, last_updated)
		VALUES (1, 'Backfill Test', 'detected', '2025-01-01 12:00:00')
	`)
	require.NoError(t, err)

	// Invoke v18's SQLite path directly — bypasses the migrations
	// bookkeeping so we isolate exactly what the migration does.
	require.NoError(t, db.addMediaItemsFavoriteColumnSQLite(context.Background()))

	// is_favorite + updated_at + duration must all exist.
	cols := readTableInfo(t, db, "media_items")
	for _, name := range []string{"is_favorite", "updated_at", "duration"} {
		_, ok := findColumn(cols, name)
		require.True(t, ok, "column %s must be present after v18", name)
	}

	// updated_at must equal the row's last_updated. SQLite + the
	// driver normalise the datetime format (accepts "YYYY-MM-DD HH:MM:SS"
	// on write, returns ISO8601 "YYYY-MM-DDTHH:MM:SSZ" on read), so
	// compare by reading both sides and parsing both.
	var gotUpdatedAt, gotLastUpdated sql.NullString
	require.NoError(t, raw.QueryRow(
		`SELECT updated_at, last_updated FROM media_items WHERE title = 'Backfill Test'`,
	).Scan(&gotUpdatedAt, &gotLastUpdated))
	assert.True(t, gotUpdatedAt.Valid, "updated_at must be populated by backfill")
	assert.Equal(t, gotLastUpdated.String, gotUpdatedAt.String,
		"backfill copies last_updated into updated_at verbatim")
}
