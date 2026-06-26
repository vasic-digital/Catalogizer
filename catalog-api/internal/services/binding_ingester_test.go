package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"testing"

	"catalogizer/database"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// testIngesterMigrations creates the tables the ingester needs.
func testIngesterMigrations(db *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS storage_roots (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			protocol TEXT NOT NULL,
			host TEXT,
			port INTEGER,
			path TEXT,
			username TEXT,
			password TEXT,
			domain TEXT,
			mount_point TEXT,
			options TEXT,
			url TEXT,
			enabled BOOLEAN DEFAULT 1,
			max_depth INTEGER DEFAULT 10,
			enable_duplicate_detection BOOLEAN DEFAULT 1,
			enable_metadata_extraction BOOLEAN DEFAULT 1,
			include_patterns TEXT,
			exclude_patterns TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			last_scan_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS share_identity_bindings (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			host TEXT NOT NULL,
			share_name TEXT NOT NULL,
			protocol TEXT NOT NULL DEFAULT 'smb',
			identity_index INTEGER NOT NULL DEFAULT 0,
			identity_label TEXT NOT NULL,
			last_ok_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(host, share_name, protocol)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_share_identity_host ON share_identity_bindings (host)`,
		`CREATE INDEX IF NOT EXISTS idx_share_identity_last_ok ON share_identity_bindings (last_ok_at)`,
	}
	for _, s := range stmts {
		if _, err := db.Exec(s); err != nil {
			return fmt.Errorf("test migration: %w\nSQL: %s", err, s)
		}
	}
	return nil
}

// TestBindingIngester_IngestProbeResult verifies the core contract:
//
//  1. An authenticated probe result creates storage_root entries + binding rows.
//  2. SECURITY (§11.4.10): the storage_root stores identity_index in the Options
//     JSON column — NEVER a password, username, or secret. The Username and
//     Password columns MUST be NULL.
//  3. Idempotency: a second call with the same result does not create duplicates
//     (NewRoots == 0 on the repeat call).
//  4. An unauthenticated probe is rejected.
//  5. A nil probe is rejected.
func TestBindingIngester_IngestProbeResult(t *testing.T) {
	logger := zap.NewNop()
	sqlDB, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { sqlDB.Close() })
	require.NoError(t, testIngesterMigrations(sqlDB))

	db := database.WrapDB(sqlDB, database.DialectSQLite)

	ingester := NewBindingIngester(db, logger)

	t.Run("rejects nil probe", func(t *testing.T) {
		_, err := ingester.IngestProbeResult(context.Background(), nil)
		require.Error(t, err)
		require.Contains(t, err.Error(), "nil probe result")
	})

	t.Run("rejects unauthenticated probe", func(t *testing.T) {
		result := &SMBProbeResult{
			Host:          "nas.example.com",
			Authenticated: false,
			IdentityIndex: 0,
			IdentityLabel: "guest",
			Shares:        []SMBShareInfo{{Host: "nas.example.com", ShareName: "Data", Path: "\\\\nas.example.com\\Data"}},
		}
		_, err := ingester.IngestProbeResult(context.Background(), result)
		require.Error(t, err)
		require.Contains(t, err.Error(), "not authenticated")
	})

	t.Run("creates storage roots and bindings from probe result", func(t *testing.T) {
		result := &SMBProbeResult{
			Host:          "192.168.1.100",
			Authenticated: true,
			IdentityIndex: 1,
			IdentityLabel: "media_user",
			Shares: []SMBShareInfo{
				{Host: "192.168.1.100", ShareName: "Data", Path: "\\\\192.168.1.100\\Data"},
				{Host: "192.168.1.100", ShareName: "Movies", Path: "\\\\192.168.1.100\\Movies"},
			},
		}

		out, err := ingester.IngestProbeResult(context.Background(), result)
		require.NoError(t, err)
		require.Equal(t, 2, out.BoundShares, "should upsert 2 binding rows")
		require.Equal(t, 2, out.NewRoots, "should insert 2 storage roots")

		// Verify storage_roots were created with the correct shape.
		verifyStorageRoot(t, sqlDB, "192.168.1.100:Data", "smb", "192.168.1.100", "Data", 1)
		verifyStorageRoot(t, sqlDB, "192.168.1.100:Movies", "smb", "192.168.1.100", "Movies", 1)

		// Verify share_identity_binding rows.
		verifyBinding(t, sqlDB, "192.168.1.100", "Data", 1, "media_user")
		verifyBinding(t, sqlDB, "192.168.1.100", "Movies", 1, "media_user")

		// SECURITY (§11.4.10): storage_root rows MUST have NULL Username and
		// NULL Password columns — no secret is stored.
		verifyNoSecretsInStorageRoot(t, sqlDB, "192.168.1.100:Data")
		verifyNoSecretsInStorageRoot(t, sqlDB, "192.168.1.100:Movies")

		// Idempotency: second call with the same result.
		out2, err := ingester.IngestProbeResult(context.Background(), result)
		require.NoError(t, err)
		require.Equal(t, 2, out2.BoundShares, "should upsert 2 bindings again (idempotent)")
		require.Equal(t, 0, out2.NewRoots, "should NOT insert new storage roots (idempotent)")
	})

	t.Run("guest identity (index 0) also creates clean storage roots", func(t *testing.T) {
		result := &SMBProbeResult{
			Host:          "files.local",
			Authenticated: true,
			IdentityIndex: 0,
			IdentityLabel: "guest",
			Shares: []SMBShareInfo{
				{Host: "files.local", ShareName: "public", Path: "\\\\files.local\\public"},
			},
		}

		out, err := ingester.IngestProbeResult(context.Background(), result)
		require.NoError(t, err)
		require.Equal(t, 1, out.BoundShares)
		require.Equal(t, 1, out.NewRoots)

		verifyStorageRoot(t, sqlDB, "files.local:public", "smb", "files.local", "public", 0)
		verifyNoSecretsInStorageRoot(t, sqlDB, "files.local:public")
		verifyBinding(t, sqlDB, "files.local", "public", 0, "guest")

		// Idempotency.
		out2, err := ingester.IngestProbeResult(context.Background(), result)
		require.NoError(t, err)
		require.Equal(t, 0, out2.NewRoots)
		require.Equal(t, 1, out2.BoundShares)
	})

	t.Run("same host different identity upserts binding but does not duplicate root", func(t *testing.T) {
		// First ingest with identity 1.
		r1 := &SMBProbeResult{
			Host:          "nas2.local",
			Authenticated: true,
			IdentityIndex: 1,
			IdentityLabel: "user1",
			Shares:        []SMBShareInfo{{Host: "nas2.local", ShareName: "Shared", Path: "\\\\nas2.local\\Shared"}},
		}
		out1, err := ingester.IngestProbeResult(context.Background(), r1)
		require.NoError(t, err)
		require.Equal(t, 1, out1.NewRoots)
		require.Equal(t, 1, out1.BoundShares)
		verifyStorageRoot(t, sqlDB, "nas2.local:Shared", "smb", "nas2.local", "Shared", 1)

		// Second ingest with identity 2 — binding updates (id->2) but root stays.
		r2 := &SMBProbeResult{
			Host:          "nas2.local",
			Authenticated: true,
			IdentityIndex: 2,
			IdentityLabel: "user2",
			Shares:        []SMBShareInfo{{Host: "nas2.local", ShareName: "Shared", Path: "\\\\nas2.local\\Shared"}},
		}
		out2, err := ingester.IngestProbeResult(context.Background(), r2)
		require.NoError(t, err)
		require.Equal(t, 0, out2.NewRoots, "root already exists for this (host, share)")
		require.Equal(t, 1, out2.BoundShares, "binding upserted with new identity")

		// Root still has identity_index=1 (first write wins on storage_root).
		verifyStorageRoot(t, sqlDB, "nas2.local:Shared", "smb", "nas2.local", "Shared", 1)
		// Binding now references identity 2.
		verifyBinding(t, sqlDB, "nas2.local", "Shared", 2, "user2")
	})

	// TestZeroSharesAuthenticatedProbe is the §11.4.115 polarity guard: an
	// authenticated probe result with ZERO non-IPC$ shares must NOT insert any
	// storage_root — the probe succeeded on the host but there are no data
	// shares to ingest. If a future change accidentally inserts a storage_root
	// for an empty share list, this test FAILs (RED), proving the guard catches
	// the regression.
	t.Run("zero shares authenticated probe does not insert any root", func(t *testing.T) {
		result := &SMBProbeResult{
			Host:          "nas.empty.example.com",
			Authenticated: true,
			IdentityIndex: 1,
			IdentityLabel: "browser",
			Shares:        []SMBShareInfo{},
		}

		out, err := ingester.IngestProbeResult(context.Background(), result)
		require.NoError(t, err)
		require.Equal(t, 0, out.BoundShares, "0 shares → 0 bound shares")
		require.Equal(t, 0, out.NewRoots, "0 shares → 0 new storage roots")

		// Verify no storage_root was created.
		var count int
		err = sqlDB.QueryRow(
			`SELECT COUNT(*) FROM storage_roots WHERE host = 'nas.empty.example.com'`,
		).Scan(&count)
		require.NoError(t, err)
		require.Equal(t, 0, count, "no storage_root should exist for a host with zero shares")

		// Verify no binding was created.
		err = sqlDB.QueryRow(
			`SELECT COUNT(*) FROM share_identity_bindings WHERE host = 'nas.empty.example.com'`,
		).Scan(&count)
		require.NoError(t, err)
		require.Equal(t, 0, count, "no binding should exist for a host with zero shares")
	})

	// TestIdentityLabelSecretLeakGuard is the §11.4.10 secret-leak guard: an
	// identity_label containing a SQL injection attempt or a tricky string is
	// stored AS-IS (it's a label, not a credential), but MUST NOT leak a
	// password or be interpreted as one. The label goes into the
	// share_identity_binding.identity_label column verbatim — it's never
	// executed as SQL, never interpreted as a password, and never copied into
	// storage_root.username/password.
	t.Run("identity label with tricky string is stored as-is, never interpreted as credential", func(t *testing.T) {
		trickyLabel := "admin' OR '1'='1"
		result := &SMBProbeResult{
			Host:          "tricky.example.com",
			Authenticated: true,
			IdentityIndex: 2,
			IdentityLabel: trickyLabel,
			Shares: []SMBShareInfo{
				{Host: "tricky.example.com", ShareName: "Data", Path: "\\\\tricky.example.com\\Data"},
			},
		}

		out, err := ingester.IngestProbeResult(context.Background(), result)
		require.NoError(t, err)
		require.Equal(t, 1, out.BoundShares)
		require.Equal(t, 1, out.NewRoots)

		// The label must be stored verbatim — not sanitised, not escaped, just as-is.
		verifyBinding(t, sqlDB, "tricky.example.com", "Data", 2, trickyLabel)

		// SECURITY (§11.4.10): the storage_root must have NULL username and
		// password — the tricky label must NOT end up in those columns.
		verifyNoSecretsInStorageRoot(t, sqlDB, "tricky.example.com:Data")

		// The storage_root.identity_index is the numeric index (2), not the label.
		verifyStorageRoot(t, sqlDB, "tricky.example.com:Data", "smb", "tricky.example.com", "Data", 2)
	})
}

// verifyStorageRoot checks a storage_roots row has the expected shape.
func verifyStorageRoot(t *testing.T, db *sql.DB, name, protocol, host, path string, wantIdentityIdx int) {
	t.Helper()

	var (
		gotID       int64
		gotName     string
		gotProtocol string
		gotHost     *string
		gotPath     *string
		gotOptions  *string
	)

	err := db.QueryRow(
		`SELECT id, name, protocol, host, path, options FROM storage_roots WHERE name = ?`, name,
	).Scan(&gotID, &gotName, &gotProtocol, &gotHost, &gotPath, &gotOptions)
	require.NoError(t, err, "storage_root %s should exist", name)

	require.Equal(t, name, gotName)
	require.Equal(t, protocol, gotProtocol)
	require.NotNil(t, gotHost)
	require.Equal(t, host, *gotHost)
	require.NotNil(t, gotPath)
	require.Equal(t, path, *gotPath)

	// Options must contain identity_index.
	require.NotNil(t, gotOptions, "options column must not be nil")
	var meta map[string]int
	require.NoError(t, json.Unmarshal([]byte(*gotOptions), &meta))
	require.Equal(t, wantIdentityIdx, meta["identity_index"],
		"options JSON must contain identity_index=%d", wantIdentityIdx)
}

// verifyNoSecretsInStorageRoot asserts that Username and Password columns are
// BOTH NULL for the given storage root — the fundamental §11.4.10 invariant.
func verifyNoSecretsInStorageRoot(t *testing.T, db *sql.DB, name string) {
	t.Helper()

	var username, password *string
	err := db.QueryRow(
		`SELECT username, password FROM storage_roots WHERE name = ?`, name,
	).Scan(&username, &password)
	require.NoError(t, err, "storage_root %s should exist", name)
	require.Nil(t, username, "username MUST be NULL (no secret in storage_root): %s", name)
	require.Nil(t, password, "password MUST be NULL (no secret in storage_root): %s", name)
}

// verifyBinding checks a share_identity_binding row has the expected fields.
func verifyBinding(t *testing.T, db *sql.DB, host, shareName string, wantIdx int, wantLabel string) {
	t.Helper()

	var (
		gotHost  string
		gotShare string
		gotIdx   int
		gotLabel string
	)

	err := db.QueryRow(
		`SELECT host, share_name, identity_index, identity_label
		   FROM share_identity_bindings
		  WHERE host = ? AND share_name = ?`,
		host, shareName,
	).Scan(&gotHost, &gotShare, &gotIdx, &gotLabel)
	require.NoError(t, err, "binding for %s/%s should exist", host, shareName)

	require.Equal(t, host, gotHost)
	require.Equal(t, shareName, gotShare)
	require.Equal(t, wantIdx, gotIdx)
	require.Equal(t, wantLabel, gotLabel)
}
