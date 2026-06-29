-- SQLite: enforce one external_metadata row per (media_item_id, provider).
--
-- FIX-CONCURRENCY-2026-06-29: the read-then-write Upsert allowed concurrent
-- enrichment to create duplicate (media_item_id, provider) rows (confirmed in
-- production: 1149 rows deduped). Nothing in the schema prevented it.
--
-- Self-healing + idempotent: delete older duplicates (keep MAX(id) per group)
-- BEFORE creating the unique index so the index can be created on dirty data;
-- both steps are no-ops once the data is clean and the index exists.
--
-- NOTE: the authoritative runtime migration is the Go migration v20
-- (database/migrations_v20_external_metadata_unique.go, registered in
-- database/migrations.go and applied by DB.RunMigrations). This SQL file
-- mirrors it for the golang-migrate CLI path documented in this directory's
-- README.

DELETE FROM external_metadata
WHERE id NOT IN (
    SELECT MAX(id) FROM external_metadata GROUP BY media_item_id, provider
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_external_metadata_item_provider
    ON external_metadata (media_item_id, provider);
