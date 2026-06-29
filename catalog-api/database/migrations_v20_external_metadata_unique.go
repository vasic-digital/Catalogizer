package database

import (
	"context"
	"fmt"
)

// createExternalMetadataUniqueIndex enforces at most one external_metadata row
// per (media_item_id, provider) pair.
//
// FIX-CONCURRENCY-2026-06-29: ExternalMetadataRepository.Upsert used a
// read-then-write sequence (findByItemAndProvider -> if exists UPDATE else
// Create). Under concurrent enrichment (multiple goroutines/batches processing
// the same media item at once) two callers could both find no existing row and
// both Create, producing DUPLICATE (media_item_id, provider) rows. This was
// confirmed in production — `tmdb`/`tmdb` duplicate pairs existed and 1149 rows
// were manually deduped. The pre-existing
// idx_external_metadata_provider(provider, external_id) index is NOT unique and
// does NOT cover (media_item_id, provider), so nothing prevented the duplicate.
//
// This migration is the real guarantee: a UNIQUE index on
// (media_item_id, provider). The repository's Upsert is updated in lockstep to
// recover the losing concurrent INSERT as an UPDATE so the race never surfaces
// as an error.
//
// Self-healing: existing environments may already carry duplicate rows, which
// would make CREATE UNIQUE INDEX fail. The migration first deletes the older
// duplicates (keeping the highest id per group — the most recently enriched
// row), then creates the index. Both steps are idempotent: dedup is a no-op
// once the data is clean, and CREATE UNIQUE INDEX IF NOT EXISTS is a no-op once
// the index exists.
func (db *DB) createExternalMetadataUniqueIndex(ctx context.Context) error {
	if db.dialect.IsPostgres() {
		return db.createExternalMetadataUniqueIndexPostgres(ctx)
	}
	return db.createExternalMetadataUniqueIndexSQLite(ctx)
}

// dedupExternalMetadataSQL keeps the highest id per (media_item_id, provider)
// group and removes the rest. id is INTEGER PRIMARY KEY AUTOINCREMENT / SERIAL
// (never NULL), so the MAX(id) subquery contains no NULLs and NOT IN is safe.
// Portable across SQLite and PostgreSQL.
const dedupExternalMetadataSQL = `DELETE FROM external_metadata
WHERE id NOT IN (
	SELECT MAX(id) FROM external_metadata GROUP BY media_item_id, provider
)`

const createExternalMetadataUniqueIndexSQL = `CREATE UNIQUE INDEX IF NOT EXISTS idx_external_metadata_item_provider
	ON external_metadata (media_item_id, provider)`

func (db *DB) createExternalMetadataUniqueIndexSQLite(ctx context.Context) error {
	if _, err := db.ExecContext(ctx, dedupExternalMetadataSQL); err != nil {
		return fmt.Errorf("sqlite: dedup external_metadata before unique index: %w", err)
	}
	if _, err := db.ExecContext(ctx, createExternalMetadataUniqueIndexSQL); err != nil {
		return fmt.Errorf("sqlite: create unique index on external_metadata(media_item_id, provider): %w", err)
	}
	return nil
}

func (db *DB) createExternalMetadataUniqueIndexPostgres(ctx context.Context) error {
	if _, err := db.ExecContext(ctx, dedupExternalMetadataSQL); err != nil {
		return fmt.Errorf("postgres: dedup external_metadata before unique index: %w", err)
	}
	if _, err := db.ExecContext(ctx, createExternalMetadataUniqueIndexSQL); err != nil {
		return fmt.Errorf("postgres: create unique index on external_metadata(media_item_id, provider): %w", err)
	}
	return nil
}
