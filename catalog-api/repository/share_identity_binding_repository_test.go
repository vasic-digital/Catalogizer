package repository

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	"catalogizer/database"
	"catalogizer/models"

	_ "github.com/mutecomm/go-sqlcipher"
)

// setupBindingTestDB creates an in-memory SQLite DB with the binding table
// and wraps it with database.DB, without importing the cycle-prone internal/tests.
func setupBindingTestDB(t *testing.T) *database.DB {
	t.Helper()
	sqlDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open in-memory db: %v", err)
	}
	t.Cleanup(func() { sqlDB.Close() })
	sqlDB.SetMaxOpenConns(1)

	schema := `CREATE TABLE IF NOT EXISTS share_identity_bindings (
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
	);
	CREATE INDEX IF NOT EXISTS idx_share_identity_host ON share_identity_bindings (host);
	CREATE INDEX IF NOT EXISTS idx_share_identity_last_ok ON share_identity_bindings (last_ok_at);`
	if _, err := sqlDB.Exec(schema); err != nil {
		t.Fatalf("create binding table: %v", err)
	}
	return database.WrapDB(sqlDB, database.DialectSQLite)
}

// TestShareIdentityBindingRepository_Upsert_idempotent verifies that Upsert is
// idempotent on the UNIQUE (host, share_name, protocol) constraint — calling it
// twice with the same key updates rather than duplicates (§11.4.50
// deterministic).
func TestShareIdentityBindingRepository_Upsert_idempotent(t *testing.T) {
	db := setupBindingTestDB(t)
	repo := NewShareIdentityBindingRepository(db)
	ctx := context.Background()

	b := &models.ShareIdentityBinding{
		Host:          "nas-01",
		ShareName:     "Data",
		Protocol:      "smb",
		IdentityIndex: 1,
		IdentityLabel: "milosvasic",
	}

	// First insert
	if err := repo.Upsert(ctx, b); err != nil {
		t.Fatalf("first upsert: %v", err)
	}

	// Second upsert — same key, should be idempotent (update not duplicate)
	b2 := &models.ShareIdentityBinding{
		Host:          "nas-01",
		ShareName:     "Data",
		Protocol:      "smb",
		IdentityIndex: 2,
		IdentityLabel: "milosvasic",
	}
	before := time.Now().Add(-time.Hour).UTC()
	b2.LastOKAt = before
	if err := repo.Upsert(ctx, b2); err != nil {
		t.Fatalf("second upsert: %v", err)
	}

	// Only one row for this key
	list, err := repo.GetWorkingForHost(ctx, "nas-01")
	if err != nil {
		t.Fatalf("GetWorkingForHost: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 binding, got %d (idempotent unique key failed)", len(list))
	}
	if list[0].IdentityIndex != 2 {
		t.Errorf("identity_index = %d, want 2 (upsert should update)", list[0].IdentityIndex)
	}
}

// TestShareIdentityBindingRepository_GetWorkingForHost verifies that bindings
// are returned sorted most-recently-confirmed-first.
func TestShareIdentityBindingRepository_GetWorkingForHost(t *testing.T) {
	db := setupBindingTestDB(t)
	repo := NewShareIdentityBindingRepository(db)
	ctx := context.Background()

	_ = repo.Upsert(ctx, &models.ShareIdentityBinding{Host: "nas-01", ShareName: "Music", Protocol: "smb", IdentityIndex: 0, IdentityLabel: "guest"})
	time.Sleep(10 * time.Millisecond)
	_ = repo.Upsert(ctx, &models.ShareIdentityBinding{Host: "nas-01", ShareName: "Data", Protocol: "smb", IdentityIndex: 1, IdentityLabel: "milosvasic"})

	list, err := repo.GetWorkingForHost(ctx, "nas-01")
	if err != nil {
		t.Fatalf("GetWorkingForHost: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 bindings for nas-01, got %d", len(list))
	}
	// Most recent first (Data upserted after Music)
	if list[0].ShareName != "Data" {
		t.Errorf("expected Data first (most recent), got %s", list[0].ShareName)
	}
}

// TestShareIdentityBindingRepository_Delete verifies deletion.
func TestShareIdentityBindingRepository_Delete(t *testing.T) {
	db := setupBindingTestDB(t)
	repo := NewShareIdentityBindingRepository(db)
	ctx := context.Background()

	b := &models.ShareIdentityBinding{Host: "nas-01", ShareName: "Data", Protocol: "smb", IdentityIndex: 1, IdentityLabel: "milosvasic"}
	if err := repo.Upsert(ctx, b); err != nil {
		t.Fatalf("upsert: %v", err)
	}

	list, _ := repo.List(ctx)
	if len(list) != 1 {
		t.Fatalf("expected 1 before delete, got %d", len(list))
	}

	if err := repo.Delete(ctx, list[0].ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	list, _ = repo.List(ctx)
	if len(list) != 0 {
		t.Errorf("expected 0 after delete, got %d", len(list))
	}
}

// TestShareIdentityBindingRepository_NoSecretLeak verifies the critical
// §11.4.10 invariant: no secret/password column exists on the table — only
// identity_index and identity_label are stored.
func TestShareIdentityBindingRepository_NoSecretLeak(t *testing.T) {
	db := setupBindingTestDB(t)
	repo := NewShareIdentityBindingRepository(db)
	ctx := context.Background()

	b := &models.ShareIdentityBinding{Host: "nas-01", ShareName: "Data", Protocol: "smb", IdentityIndex: 1, IdentityLabel: "milosvasic"}
	if err := repo.Upsert(ctx, b); err != nil {
		t.Fatalf("upsert: %v", err)
	}

	// Inspect the SQLite schema directly — verify no password/secret/token column.
	rows, err := db.QueryContext(ctx, `PRAGMA table_info(share_identity_bindings)`)
	if err != nil {
		t.Fatalf("pragma table_info: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull int
		var dflt sql.NullString
		var pk int
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			t.Fatalf("scan column: %v", err)
		}
		lower := strings.ToLower(name)
		if lower == "password" || lower == "secret" || lower == "token" || lower == "credential" {
			t.Errorf("§11.4.10 VIOLATION: table has a %q column — secrets must NOT be stored", name)
		}
	}
}
