package services

import (
	"context"
	"testing"
	"time"

	"catalogizer/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestQualityRevalidator_TouchesStaleRows(t *testing.T) {
	db := setupTestAssetDBForQG(t)
	repo := repository.NewImageQualityRepository(db)

	// Seed two rows: one fresh, one stale.
	require.NoError(t, repo.Upsert(context.Background(), &repository.ImageQualityAssessment{
		EntityType: "movie", EntityID: 1, Variant: "primary",
		Source: "tmdb", Width: 1000, Height: 1500, BlurVar: 120, BPP: 0.6, AspectRatio: 0.667,
		Verdict: "pass", AssessedAt: time.Now(), LastCheckedAt: time.Now(),
	}))
	require.NoError(t, repo.Upsert(context.Background(), &repository.ImageQualityAssessment{
		EntityType: "movie", EntityID: 2, Variant: "primary",
		Source: "tmdb", Width: 1000, Height: 1500, BlurVar: 120, BPP: 0.6, AspectRatio: 0.667,
		Verdict: "pass", AssessedAt: time.Now(), LastCheckedAt: time.Now(),
	}))
	// Backdate row 2 so it is stale.
	_, err := db.ExecContext(context.Background(),
		`UPDATE image_quality_assessments SET last_checked_at = ? WHERE entity_id = 2`,
		time.Now().Add(-30*24*time.Hour))
	require.NoError(t, err)

	rv := NewQualityRevalidator(repo, zap.NewNop(),
		WithRevalidationStaleAge(24*time.Hour),
		WithRevalidationBatch(10),
	)
	rv.runOnce(context.Background())

	row2, err := repo.Find(context.Background(), "movie", 2, "primary")
	require.NoError(t, err)
	assert.True(t, time.Since(row2.LastCheckedAt) < time.Minute, "stale row should be touched")

	row1, err := repo.Find(context.Background(), "movie", 1, "primary")
	require.NoError(t, err)
	assert.True(t, time.Since(row1.LastCheckedAt) < time.Hour, "fresh row remains untouched")
}

func TestQualityRevalidator_StartStopSafe(t *testing.T) {
	db := setupTestAssetDBForQG(t)
	repo := repository.NewImageQualityRepository(db)
	rv := NewQualityRevalidator(repo, zap.NewNop(),
		WithRevalidationInterval(10*time.Millisecond),
		WithRevalidationStaleAge(10*time.Millisecond),
		WithRevalidationBatch(5),
	)
	rv.Start(context.Background())
	time.Sleep(25 * time.Millisecond)
	rv.Stop()
	// Second stop must be a no-op.
	rv.Stop()
}

func TestQualityRevalidator_NilReceiverSafe(t *testing.T) {
	var rv *QualityRevalidator
	rv.Start(context.Background()) // must not panic
	rv.Stop()                      // must not panic
}

func TestQualityRevalidator_NilRepoNoCrash(t *testing.T) {
	rv := NewQualityRevalidator(nil, zap.NewNop())
	rv.Start(context.Background())
	rv.Stop()
}
