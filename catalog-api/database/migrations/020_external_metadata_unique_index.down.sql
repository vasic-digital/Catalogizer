-- Rollback: drop the unique index on external_metadata(media_item_id, provider).
-- The deduplicated rows are NOT restored (data loss on the duplicate rows is
-- intentional and irreversible — the duplicates were the defect).
DROP INDEX IF EXISTS idx_external_metadata_item_provider;
