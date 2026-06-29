package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"catalogizer/config"
	"catalogizer/database"
	"catalogizer/internal/media/models"
	"catalogizer/repository"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupEntityTestDB(t *testing.T) (*database.DB, func()) {
	t.Helper()
	tmpFile, err := os.CreateTemp("", "entity_handler_test_*.db")
	require.NoError(t, err)
	tmpFile.Close()

	cfg := &config.DatabaseConfig{
		Type:               "sqlite",
		Path:               tmpFile.Name(),
		MaxOpenConnections: 1,
		MaxIdleConnections: 1,
		ConnMaxLifetime:    3600,
		EnableWAL:          true,
		CacheSize:          2000,
		BusyTimeout:        5000,
	}

	db, err := database.NewConnection(cfg)
	require.NoError(t, err)

	err = db.RunMigrations(context.Background())
	require.NoError(t, err)

	cleanup := func() {
		db.Close()
		os.Remove(tmpFile.Name())
	}
	return db, cleanup
}

func setupEntityHandler(t *testing.T, db *database.DB) (*MediaEntityHandler, *repository.MediaItemRepository) {
	t.Helper()
	itemRepo := repository.NewMediaItemRepository(db)
	fileRepo := repository.NewMediaFileRepository(db)
	extMetaRepo := repository.NewExternalMetadataRepository(db)
	userMetaRepo := repository.NewUserMetadataRepository(db)

	handler := NewMediaEntityHandler(itemRepo, fileRepo, extMetaRepo, userMetaRepo)
	return handler, itemRepo
}

func TestMediaEntityHandler_GetEntityTypes(t *testing.T) {
	db, cleanup := setupEntityTestDB(t)
	defer cleanup()

	handler, _ := setupEntityHandler(t, db)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/entities/types", nil)

	handler.GetEntityTypes(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	types, ok := resp["types"].([]interface{})
	require.True(t, ok)
	assert.GreaterOrEqual(t, len(types), 11) // 11 seeded types
}

func TestMediaEntityHandler_GetEntityStats(t *testing.T) {
	db, cleanup := setupEntityTestDB(t)
	defer cleanup()

	handler, itemRepo := setupEntityHandler(t, db)
	ctx := context.Background()

	_, typeID, err := itemRepo.GetMediaTypeByName(ctx, "movie")
	require.NoError(t, err)

	_, _ = itemRepo.Create(ctx, &models.MediaItem{MediaTypeID: typeID, Title: "Movie 1", Status: "detected"})
	_, _ = itemRepo.Create(ctx, &models.MediaItem{MediaTypeID: typeID, Title: "Movie 2", Status: "detected"})

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/entities/stats", nil)

	handler.GetEntityStats(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(2), resp["total_entities"])
}

func TestMediaEntityHandler_ListEntities(t *testing.T) {
	db, cleanup := setupEntityTestDB(t)
	defer cleanup()

	handler, itemRepo := setupEntityHandler(t, db)
	ctx := context.Background()

	_, typeID, _ := itemRepo.GetMediaTypeByName(ctx, "movie")
	_, _ = itemRepo.Create(ctx, &models.MediaItem{MediaTypeID: typeID, Title: "The Matrix", Status: "detected"})
	_, _ = itemRepo.Create(ctx, &models.MediaItem{MediaTypeID: typeID, Title: "Inception", Status: "detected"})

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/entities?query=Matrix", nil)

	handler.ListEntities(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(1), resp["total"])
}

func TestMediaEntityHandler_GetEntity(t *testing.T) {
	db, cleanup := setupEntityTestDB(t)
	defer cleanup()

	handler, itemRepo := setupEntityHandler(t, db)
	ctx := context.Background()

	_, typeID, _ := itemRepo.GetMediaTypeByName(ctx, "movie")
	year := 1999
	id, _ := itemRepo.Create(ctx, &models.MediaItem{
		MediaTypeID: typeID,
		Title:       "The Matrix",
		Year:        &year,
		Status:      "detected",
	})

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/entities/1", nil)
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	_ = id

	handler.GetEntity(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "The Matrix", resp["title"])
	assert.Equal(t, "movie", resp["media_type"])
}

func TestMediaEntityHandler_EnrichAllEntities_NoEntities(t *testing.T) {
	db, cleanup := setupEntityTestDB(t)
	defer cleanup()

	handler, _ := setupEntityHandler(t, db)
	handler.db = db

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/v1/entities/enrich", nil)

	handler.EnrichAllEntities(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "No entities need enrichment", resp["message"])
	assert.Equal(t, false, resp["accepted"])
}

func TestMediaEntityHandler_EnrichAllEntities_QueuesBackgroundJob(t *testing.T) {
	db, cleanup := setupEntityTestDB(t)
	defer cleanup()

	handler, itemRepo := setupEntityHandler(t, db)
	handler.db = db
	ctx := context.Background()

	// Seed a media type and an entity
	_, typeID, _ := itemRepo.GetMediaTypeByName(ctx, "movie")
	id, _ := itemRepo.Create(ctx, &models.MediaItem{MediaTypeID: typeID, Title: "Test Movie", Status: "detected"})

	// Seed directory_analyses so the JOIN succeeds
	_, err := db.ExecContext(ctx,
		`INSERT INTO directory_analyses (directory_path, media_item_id) VALUES (?, ?)`,
		"/media/movies/Test Movie", id)
	require.NoError(t, err)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/v1/entities/enrich", nil)

	handler.EnrichAllEntities(c)

	// Should return 200 OK immediately (async processing)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "Batch enrichment queued", resp["message"])
	assert.Equal(t, true, resp["accepted"])

	// Wait for background goroutine to finish
	handler.Close()

	// Verify the entity now has external metadata (or at least the handler didn't crash)
	// Since TMDB is not configured, the enrichment may not add metadata,
	// but the background job should complete without panic.
}

func TestMediaEntityHandler_EnrichAllEntities_RespectsLimitParam(t *testing.T) {
	db, cleanup := setupEntityTestDB(t)
	defer cleanup()

	handler, itemRepo := setupEntityHandler(t, db)
	handler.db = db
	ctx := context.Background()

	// Seed multiple entities
	_, typeID, _ := itemRepo.GetMediaTypeByName(ctx, "movie")
	for i := 1; i <= 10; i++ {
		id, _ := itemRepo.Create(ctx, &models.MediaItem{MediaTypeID: typeID, Title: fmt.Sprintf("Movie %d", i), Status: "detected"})
		_, err := db.ExecContext(ctx,
			`INSERT INTO directory_analyses (directory_path, media_item_id) VALUES (?, ?)`,
			fmt.Sprintf("/media/movies/Movie %d", i), id)
		require.NoError(t, err)
	}

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/v1/entities/enrich?limit=3", nil)

	handler.EnrichAllEntities(c)

	// Should return 200 OK immediately (async processing)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, true, resp["accepted"])

	handler.Close()
}

func TestMediaEntityHandler_EnrichAllEntities_LimitBounds(t *testing.T) {
	db, cleanup := setupEntityTestDB(t)
	defer cleanup()

	handler, itemRepo := setupEntityHandler(t, db)
	handler.db = db
	ctx := context.Background()

	_, typeID, _ := itemRepo.GetMediaTypeByName(ctx, "movie")
	id, _ := itemRepo.Create(ctx, &models.MediaItem{MediaTypeID: typeID, Title: "Solo", Status: "detected"})
	_, err := db.ExecContext(ctx,
		`INSERT INTO directory_analyses (directory_path, media_item_id) VALUES (?, ?)`,
		"/media/movies/Solo", id)
	require.NoError(t, err)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/v1/entities/enrich?limit=999", nil)

	handler.EnrichAllEntities(c)

	// Should return 200 OK immediately (async processing)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, true, resp["accepted"])

	handler.Close()
}

func TestMediaEntityHandler_EnrichAllEntities_NoDB(t *testing.T) {
	db, cleanup := setupEntityTestDB(t)
	defer cleanup()

	handler, _ := setupEntityHandler(t, db)
	// handler.db is nil by default

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/v1/entities/enrich", nil)

	handler.EnrichAllEntities(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "Database not available", resp["error"])
}

// TestMediaEntityHandler_EnrichSelectsItemsWithoutDirectoryAnalyses proves
// that the enrichment selection reaches EVERY media item lacking external
// metadata — including items that have NO directory_analyses row. The
// original query inner-JOINed directory_analyses, so an item without such a
// row (23704 of 27750 items in production) could never be enriched. The
// LEFT JOIN makes those items reachable with an empty DirPath, so the
// background job falls straight through to the TMDB/LLM path for them.
func TestMediaEntityHandler_EnrichSelectsItemsWithoutDirectoryAnalyses(t *testing.T) {
	db, cleanup := setupEntityTestDB(t)
	defer cleanup()

	handler, itemRepo := setupEntityHandler(t, db)
	handler.db = db
	ctx := context.Background()

	_, typeID, err := itemRepo.GetMediaTypeByName(ctx, "movie")
	require.NoError(t, err)

	// Item WITH a directory_analyses row — was always reachable.
	withDA, err := itemRepo.Create(ctx, &models.MediaItem{MediaTypeID: typeID, Title: "Has Directory", Status: "detected"})
	require.NoError(t, err)
	_, err = db.ExecContext(ctx,
		`INSERT INTO directory_analyses (directory_path, media_item_id) VALUES (?, ?)`,
		"/media/movies/Has Directory", withDA)
	require.NoError(t, err)

	// Item WITHOUT any directory_analyses row — the previously-unreachable
	// case the inner JOIN silently dropped.
	withoutDA, err := itemRepo.Create(ctx, &models.MediaItem{MediaTypeID: typeID, Title: "No Directory", Status: "detected"})
	require.NoError(t, err)

	entities, err := handler.selectEntitiesNeedingEnrichment(ctx, 50)
	require.NoError(t, err)

	got := make(map[int64]string, len(entities))
	for _, e := range entities {
		got[e.ID] = e.DirPath
	}

	// The item WITH directory_analyses must still be selected, carrying its path.
	dirPath, ok := got[withDA]
	assert.True(t, ok, "item WITH directory_analyses must be selected for enrichment")
	assert.Equal(t, "/media/movies/Has Directory", dirPath)

	// The item WITHOUT directory_analyses MUST now be selected (LEFT JOIN),
	// with an empty DirPath so the local-cover lookup is skipped.
	dirPath, ok = got[withoutDA]
	assert.True(t, ok, "item WITHOUT directory_analyses MUST be selected (LEFT JOIN); the inner JOIN dropped it")
	assert.Equal(t, "", dirPath, "item without directory_analyses must have empty DirPath so the local-cover lookup is skipped")
}

// TestMediaEntityHandler_EnrichExcludesItemsWithExternalMetadata proves the
// LEFT JOIN change did NOT loosen the `WHERE em.id IS NULL` filter: items
// that already carry external metadata are still excluded from enrichment.
func TestMediaEntityHandler_EnrichExcludesItemsWithExternalMetadata(t *testing.T) {
	db, cleanup := setupEntityTestDB(t)
	defer cleanup()

	handler, itemRepo := setupEntityHandler(t, db)
	handler.db = db
	ctx := context.Background()

	_, typeID, err := itemRepo.GetMediaTypeByName(ctx, "movie")
	require.NoError(t, err)

	needs, err := itemRepo.Create(ctx, &models.MediaItem{MediaTypeID: typeID, Title: "Needs Cover", Status: "detected"})
	require.NoError(t, err)

	hasMeta, err := itemRepo.Create(ctx, &models.MediaItem{MediaTypeID: typeID, Title: "Already Enriched", Status: "detected"})
	require.NoError(t, err)
	coverURL := "/api/v1/download/file/1"
	require.NoError(t, handler.extMetaRepo.Upsert(ctx, &models.ExternalMetadata{
		MediaItemID: hasMeta,
		Provider:    "local_scan",
		ExternalID:  fmt.Sprintf("local:%d", hasMeta),
		CoverURL:    &coverURL,
	}))

	entities, err := handler.selectEntitiesNeedingEnrichment(ctx, 50)
	require.NoError(t, err)

	ids := make(map[int64]bool, len(entities))
	for _, e := range entities {
		ids[e.ID] = true
	}
	assert.True(t, ids[needs], "item without external_metadata must be selected")
	assert.False(t, ids[hasMeta], "item WITH external_metadata must be excluded from enrichment")
}

// TestMediaEntityHandler_MarkEnrichmentAttemptedAdvancesQueue proves the
// progress-marker fix for the stuck-queue bug: when an item is processed but
// no cover/metadata can be found (TMDB no-result AND LLM fallback empty), a
// sentinel external_metadata row is written so the item is no longer selected
// by `WHERE em.id IS NULL` on the next batch. Without the sentinel,
// selectEntitiesNeedingEnrichment returns the SAME leading unmatchable items
// on every call and the queue never advances to the real movies deeper in the
// catalog (operator symptom: only ~5 covers ever appear).
func TestMediaEntityHandler_MarkEnrichmentAttemptedAdvancesQueue(t *testing.T) {
	db, cleanup := setupEntityTestDB(t)
	defer cleanup()

	handler, itemRepo := setupEntityHandler(t, db)
	handler.db = db
	ctx := context.Background()

	_, typeID, err := itemRepo.GetMediaTypeByName(ctx, "movie")
	require.NoError(t, err)

	// An item whose enrichment yields nothing (noisy title, no TMDB match,
	// no LLM fallback) — the leading unmatchable item that jammed the queue.
	stuck, err := itemRepo.Create(ctx, &models.MediaItem{MediaTypeID: typeID, Title: "Noisy Album Title 2049 [FLAC]", Status: "detected"})
	require.NoError(t, err)

	// Precondition: with no external_metadata row, the item is selected.
	before, err := handler.selectEntitiesNeedingEnrichment(ctx, 50)
	require.NoError(t, err)
	selectedBefore := false
	for _, e := range before {
		if e.ID == stuck {
			selectedBefore = true
		}
	}
	require.True(t, selectedBefore, "precondition: an item without external_metadata must be selected for enrichment")

	// Process it: enrichment found nothing, so mark it attempted.
	require.NoError(t, handler.markEnrichmentAttempted(ctx, stuck))

	// After the sentinel, the item MUST no longer be selected — the queue
	// advances instead of returning the same stuck item forever.
	after, err := handler.selectEntitiesNeedingEnrichment(ctx, 50)
	require.NoError(t, err)
	for _, e := range after {
		assert.NotEqual(t, stuck, e.ID, "item marked enrichment-attempted MUST NOT be re-selected; the queue must advance")
	}
}

// TestMediaEntityHandler_EnrichmentAttemptedSentinelHasNoCover proves the
// sentinel does NOT masquerade as a real cover. cover_art_service.GetCoverURL
// selects `WHERE cover_url IS NOT NULL AND cover_url != ''`; the sentinel row
// carries a NULL cover_url, so that lookup finds nothing and falls through to
// the placeholder — no false cover is ever shown for an unmatchable item.
func TestMediaEntityHandler_EnrichmentAttemptedSentinelHasNoCover(t *testing.T) {
	db, cleanup := setupEntityTestDB(t)
	defer cleanup()

	handler, itemRepo := setupEntityHandler(t, db)
	handler.db = db
	ctx := context.Background()

	_, typeID, err := itemRepo.GetMediaTypeByName(ctx, "movie")
	require.NoError(t, err)

	item, err := itemRepo.Create(ctx, &models.MediaItem{MediaTypeID: typeID, Title: "Unmatchable Comic Scan", Status: "detected"})
	require.NoError(t, err)

	require.NoError(t, handler.markEnrichmentAttempted(ctx, item))

	// Exactly one (sentinel) external_metadata row exists for the item.
	var total int
	require.NoError(t, db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM external_metadata WHERE media_item_id = ?`, item).Scan(&total))
	assert.Equal(t, 1, total, "exactly one sentinel external_metadata row must be written")

	// Mirror GetCoverURL's filter: the sentinel must NOT present a usable cover.
	var usableCovers int
	require.NoError(t, db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM external_metadata
		 WHERE media_item_id = ? AND cover_url IS NOT NULL AND cover_url != ''`, item).Scan(&usableCovers))
	assert.Equal(t, 0, usableCovers, "sentinel row must carry a NULL/empty cover_url so GetCoverURL falls through to the placeholder")
}
