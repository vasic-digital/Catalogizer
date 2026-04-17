package database

import (
	"context"
	"fmt"
)

// createImageQualityAssessmentsTable persists the result of a quality gate
// check for a specific (entity_type, entity_id, variant) tuple. Writing the
// assessment lets the cover-art pipeline skip re-scoring on subsequent
// requests and gives operators a queryable record of which provider supplied
// each cover.
func (db *DB) createImageQualityAssessmentsTable(ctx context.Context) error {
	if db.dialect.IsPostgres() {
		return db.createImageQualityAssessmentsPostgres(ctx)
	}
	return db.createImageQualityAssessmentsSQLite(ctx)
}

func (db *DB) createImageQualityAssessmentsSQLite(ctx context.Context) error {
	schema := `
	CREATE TABLE IF NOT EXISTS image_quality_assessments (
		id              INTEGER PRIMARY KEY AUTOINCREMENT,
		entity_type     TEXT NOT NULL,
		entity_id       INTEGER NOT NULL,
		variant         TEXT NOT NULL,
		source          TEXT NOT NULL,
		provider_ref    TEXT,
		width           INTEGER NOT NULL,
		height          INTEGER NOT NULL,
		blur_var        REAL NOT NULL,
		bpp             REAL NOT NULL,
		aspect_ratio    REAL,
		verdict         TEXT NOT NULL,
		fail_reason     TEXT,
		format          TEXT,
		cache_path      TEXT,
		attempt_count   INTEGER NOT NULL DEFAULT 1,
		assessed_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		last_checked_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	CREATE UNIQUE INDEX IF NOT EXISTS idx_iqa_entity_variant
		ON image_quality_assessments (entity_type, entity_id, variant);
	CREATE INDEX IF NOT EXISTS idx_iqa_source ON image_quality_assessments (source);
	CREATE INDEX IF NOT EXISTS idx_iqa_verdict ON image_quality_assessments (verdict);
	CREATE INDEX IF NOT EXISTS idx_iqa_last_checked ON image_quality_assessments (last_checked_at);
	`
	if _, err := db.ExecContext(ctx, schema); err != nil {
		return fmt.Errorf("sqlite: create image_quality_assessments: %w", err)
	}
	return nil
}

func (db *DB) createImageQualityAssessmentsPostgres(ctx context.Context) error {
	schema := `
	CREATE TABLE IF NOT EXISTS image_quality_assessments (
		id              BIGSERIAL PRIMARY KEY,
		entity_type     TEXT NOT NULL,
		entity_id       BIGINT NOT NULL,
		variant         TEXT NOT NULL,
		source          TEXT NOT NULL,
		provider_ref    TEXT,
		width           INTEGER NOT NULL,
		height          INTEGER NOT NULL,
		blur_var        DOUBLE PRECISION NOT NULL,
		bpp             DOUBLE PRECISION NOT NULL,
		aspect_ratio    DOUBLE PRECISION,
		verdict         TEXT NOT NULL,
		fail_reason     TEXT,
		format          TEXT,
		cache_path      TEXT,
		attempt_count   INTEGER NOT NULL DEFAULT 1,
		assessed_at     TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		last_checked_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	CREATE UNIQUE INDEX IF NOT EXISTS idx_iqa_entity_variant
		ON image_quality_assessments (entity_type, entity_id, variant);
	CREATE INDEX IF NOT EXISTS idx_iqa_source ON image_quality_assessments (source);
	CREATE INDEX IF NOT EXISTS idx_iqa_verdict ON image_quality_assessments (verdict);
	CREATE INDEX IF NOT EXISTS idx_iqa_last_checked ON image_quality_assessments (last_checked_at);
	`
	if _, err := db.ExecContext(ctx, schema); err != nil {
		return fmt.Errorf("postgres: create image_quality_assessments: %w", err)
	}
	return nil
}
