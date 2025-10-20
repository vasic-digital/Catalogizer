-- Rollback initial schema migration

-- Drop indexes
DROP INDEX IF EXISTS idx_files_file_type;
DROP INDEX IF EXISTS idx_files_extension;
DROP INDEX IF EXISTS idx_files_name;
DROP INDEX IF EXISTS idx_scan_history_storage_root;
DROP INDEX IF EXISTS idx_file_metadata_file_id;
DROP INDEX IF EXISTS idx_files_deleted;
DROP INDEX IF EXISTS idx_files_duplicate_group;
DROP INDEX IF EXISTS idx_files_parent_id;
DROP INDEX IF EXISTS idx_files_storage_root_path;

-- Drop tables in reverse order of dependencies
DROP TABLE IF EXISTS scan_history;
DROP TABLE IF EXISTS virtual_paths;
DROP TABLE IF EXISTS file_metadata;
DROP TABLE IF EXISTS files;
DROP TABLE IF EXISTS duplicate_groups;
DROP TABLE IF EXISTS storage_roots;
