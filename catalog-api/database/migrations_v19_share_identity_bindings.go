package database

import (
	"context"
	"fmt"
)

// createShareIdentityBindingsTable persists the remembered working combination
// of a storage share and the identity that authenticated against it — the core
// of the identity-share-discovery epic ("we remember the combinations of shares
// and identity which work together"). A row lets the discovery layer skip
// re-probing every configured identity against a known-good (host, share,
// protocol) and go straight to the one that last worked.
//
// SECURITY (§11.4.10): the table stores only identity_index (the
// CATALOGIZER_IDENTITY_<N> slot, 0 = guest) and identity_label (username or
// "guest"). It MUST NEVER hold a password, token, or secret column — the index
// is the lookup key back into the gitignored credential configuration.
func (db *DB) createShareIdentityBindingsTable(ctx context.Context) error {
	if db.dialect.IsPostgres() {
		return db.createShareIdentityBindingsPostgres(ctx)
	}
	return db.createShareIdentityBindingsSQLite(ctx)
}

func (db *DB) createShareIdentityBindingsSQLite(ctx context.Context) error {
	schema := `
	CREATE TABLE IF NOT EXISTS share_identity_bindings (
		id              INTEGER PRIMARY KEY AUTOINCREMENT,
		host            TEXT NOT NULL,
		share_name      TEXT NOT NULL,
		protocol        TEXT NOT NULL DEFAULT 'smb',
		identity_index  INTEGER NOT NULL DEFAULT 0,
		identity_label  TEXT NOT NULL,
		last_ok_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	CREATE UNIQUE INDEX IF NOT EXISTS idx_share_identity_unique
		ON share_identity_bindings (host, share_name, protocol);
	CREATE INDEX IF NOT EXISTS idx_share_identity_host
		ON share_identity_bindings (host);
	CREATE INDEX IF NOT EXISTS idx_share_identity_last_ok
		ON share_identity_bindings (last_ok_at);
	`
	if _, err := db.ExecContext(ctx, schema); err != nil {
		return fmt.Errorf("sqlite: create share_identity_bindings: %w", err)
	}
	return nil
}

func (db *DB) createShareIdentityBindingsPostgres(ctx context.Context) error {
	schema := `
	CREATE TABLE IF NOT EXISTS share_identity_bindings (
		id              BIGSERIAL PRIMARY KEY,
		host            TEXT NOT NULL,
		share_name      TEXT NOT NULL,
		protocol        TEXT NOT NULL DEFAULT 'smb',
		identity_index  INTEGER NOT NULL DEFAULT 0,
		identity_label  TEXT NOT NULL,
		last_ok_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	CREATE UNIQUE INDEX IF NOT EXISTS idx_share_identity_unique
		ON share_identity_bindings (host, share_name, protocol);
	CREATE INDEX IF NOT EXISTS idx_share_identity_host
		ON share_identity_bindings (host);
	CREATE INDEX IF NOT EXISTS idx_share_identity_last_ok
		ON share_identity_bindings (last_ok_at);
	`
	if _, err := db.ExecContext(ctx, schema); err != nil {
		return fmt.Errorf("postgres: create share_identity_bindings: %w", err)
	}
	return nil
}
