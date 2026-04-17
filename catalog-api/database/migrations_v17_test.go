package database

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateImageQualityAssessments_SQLite(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()

	ctx := context.Background()
	require.NoError(t, db.createImageQualityAssessmentsTable(ctx))

	var count int
	err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='image_quality_assessments'").Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 1, count, "image_quality_assessments should exist")

	_, err = db.ExecContext(ctx, `
		INSERT INTO image_quality_assessments
		(entity_type, entity_id, variant, source, width, height, blur_var, bpp, aspect_ratio, verdict, format, assessed_at, last_checked_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		"movie", 1, "poster", "tmdb", 1000, 1500, 120.5, 0.6, 0.667, "pass", "jpeg", time.Now(), time.Now(),
	)
	assert.NoError(t, err)

	// Unique index on (entity_type, entity_id, variant)
	_, err = db.ExecContext(ctx, `
		INSERT INTO image_quality_assessments
		(entity_type, entity_id, variant, source, width, height, blur_var, bpp, aspect_ratio, verdict, format, assessed_at, last_checked_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		"movie", 1, "poster", "fanarttv", 1500, 2250, 200.0, 0.8, 0.667, "pass", "jpeg", time.Now(), time.Now(),
	)
	assert.Error(t, err, "duplicate (entity_type, entity_id, variant) must be rejected")
}

func TestCreateImageQualityAssessments_Idempotent(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()

	ctx := context.Background()
	require.NoError(t, db.createImageQualityAssessmentsTable(ctx))
	require.NoError(t, db.createImageQualityAssessmentsTable(ctx), "second run must be idempotent")
}

func TestCreateImageQualityAssessments_IndexesPresent(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()

	ctx := context.Background()
	require.NoError(t, db.createImageQualityAssessmentsTable(ctx))

	indexes := []string{"idx_iqa_entity_variant", "idx_iqa_source", "idx_iqa_verdict", "idx_iqa_last_checked"}
	for _, idx := range indexes {
		var count int
		err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM sqlite_master WHERE type='index' AND name=?", idx).Scan(&count)
		assert.NoError(t, err, "checking index %s", idx)
		assert.Equal(t, 1, count, "index %s should exist", idx)
	}
}

func TestMigrations_IncludesV17(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()

	ctx := context.Background()
	require.NoError(t, db.RunMigrations(ctx))

	var verdictCount int
	err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM image_quality_assessments WHERE verdict='pass'").Scan(&verdictCount)
	require.NoError(t, err)
	assert.Equal(t, 0, verdictCount, "table starts empty")
}
