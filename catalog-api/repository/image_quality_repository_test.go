package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func seedAssessment(t *testing.T, repo *ImageQualityRepository, entityType string, entityID int64, variant, verdict string) *ImageQualityAssessment {
	t.Helper()
	a := &ImageQualityAssessment{
		EntityType:  entityType,
		EntityID:    entityID,
		Variant:     variant,
		Source:      "tmdb",
		Width:       1000,
		Height:      1500,
		BlurVar:     120,
		BPP:         0.6,
		AspectRatio: 0.667,
		Verdict:     verdict,
		Format:      "jpeg",
		CachePath:   "/cache/x.jpg",
	}
	require.NoError(t, repo.Upsert(context.Background(), a))
	return a
}

func TestImageQualityRepo_Upsert_InsertThenUpdate(t *testing.T) {
	db := setupTestAssetDB(t)
	repo := NewImageQualityRepository(db)

	seedAssessment(t, repo, "movie", 42, "poster", "pass")
	first, err := repo.Find(context.Background(), "movie", 42, "poster")
	require.NoError(t, err)
	assert.Equal(t, 1, first.AttemptCount)
	assert.Equal(t, "pass", first.Verdict)

	seedAssessment(t, repo, "movie", 42, "poster", "fail_blurry")
	second, err := repo.Find(context.Background(), "movie", 42, "poster")
	require.NoError(t, err)
	assert.Equal(t, 2, second.AttemptCount, "attempt_count should increment on replace")
	assert.Equal(t, "fail_blurry", second.Verdict)
}

func TestImageQualityRepo_Find_NotFound(t *testing.T) {
	db := setupTestAssetDB(t)
	repo := NewImageQualityRepository(db)

	_, err := repo.Find(context.Background(), "movie", 99, "poster")
	assert.True(t, errors.Is(err, ErrNotFound))
}

func TestImageQualityRepo_TouchLastChecked(t *testing.T) {
	db := setupTestAssetDB(t)
	repo := NewImageQualityRepository(db)

	seedAssessment(t, repo, "movie", 1, "poster", "pass")
	before, err := repo.Find(context.Background(), "movie", 1, "poster")
	require.NoError(t, err)

	time.Sleep(5 * time.Millisecond)
	require.NoError(t, repo.TouchLastChecked(context.Background(), before.ID))

	after, err := repo.Find(context.Background(), "movie", 1, "poster")
	require.NoError(t, err)
	assert.True(t, after.LastCheckedAt.After(before.LastCheckedAt) || after.LastCheckedAt.Equal(before.LastCheckedAt))
}

func TestImageQualityRepo_SampleForRevalidation(t *testing.T) {
	db := setupTestAssetDB(t)
	repo := NewImageQualityRepository(db)

	seedAssessment(t, repo, "movie", 1, "poster", "pass")
	seedAssessment(t, repo, "movie", 2, "poster", "pass")
	seedAssessment(t, repo, "movie", 3, "poster", "pass")

	// Force an older last_checked_at on row 1 to guarantee ordering.
	_, err := db.ExecContext(context.Background(),
		`UPDATE image_quality_assessments SET last_checked_at = ? WHERE entity_id = 1`,
		time.Now().Add(-48*time.Hour),
	)
	require.NoError(t, err)

	sample, err := repo.SampleForRevalidation(context.Background(), time.Now().Add(-1*time.Hour), 5)
	require.NoError(t, err)
	require.NotEmpty(t, sample)
	assert.Equal(t, int64(1), sample[0].EntityID, "oldest row first")
}

func TestImageQualityRepo_NilReceiverGuards(t *testing.T) {
	var repo *ImageQualityRepository

	_, err := repo.Find(context.Background(), "movie", 1, "poster")
	assert.True(t, errors.Is(err, ErrNotFound))

	err = repo.Upsert(context.Background(), &ImageQualityAssessment{})
	assert.Error(t, err)

	err = repo.TouchLastChecked(context.Background(), 1)
	assert.Error(t, err)

	_, err = repo.SampleForRevalidation(context.Background(), time.Now(), 5)
	assert.Error(t, err)
}
