package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"catalogizer/config"
	"catalogizer/database"
	"catalogizer/repository"

	"digital.vasic.assets/pkg/asset"
	"digital.vasic.assets/pkg/manager"
	"digital.vasic.assets/pkg/store"
	"github.com/gin-gonic/gin"
	_ "github.com/mutecomm/go-sqlcipher"
)

func setupAssetTest(t *testing.T) (*AssetHandler, *gin.Engine, *store.MemoryStore) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := database.NewConnection(&config.DatabaseConfig{Path: dbPath})
	if err != nil {
		t.Fatalf("failed to create test DB: %v", err)
	}
	if err := db.RunMigrations(context.Background()); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	memStore := store.NewMemoryStore()
	repo := repository.NewAssetRepository(db)
	mgr := manager.New(manager.WithStore(memStore))

	handler := NewAssetHandler(mgr, repo)

	router := gin.New()
	router.GET("/api/v1/assets/:id", handler.ServeAsset)
	router.POST("/api/v1/assets/request", handler.RequestAsset)
	router.GET("/api/v1/assets/by-entity/:type/:id", handler.GetByEntity)

	return handler, router, memStore
}

func TestServeAsset_DefaultPlaceholder(t *testing.T) {
	_, router, _ := setupAssetTest(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/assets/nonexistent-id", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if w.Header().Get("X-Asset-Status") != "pending" {
		t.Errorf("expected X-Asset-Status: pending, got %s", w.Header().Get("X-Asset-Status"))
	}
	if w.Header().Get("Content-Type") != "image/png" {
		t.Errorf("expected Content-Type: image/png, got %s", w.Header().Get("Content-Type"))
	}
	if w.Body.Len() == 0 {
		t.Error("expected non-empty response body")
	}
}

func TestServeAsset_Ready(t *testing.T) {
	_, router, memStore := setupAssetTest(t)

	id := asset.NewID()
	content := []byte("real-image-content")
	err := memStore.Put(context.Background(), id, bytes.NewReader(content),
		&store.Info{ContentType: "image/jpeg", Size: int64(len(content))})
	if err != nil {
		t.Fatalf("failed to put content: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/assets/"+id.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if w.Header().Get("X-Asset-Status") != "ready" {
		t.Errorf("expected X-Asset-Status: ready, got %s", w.Header().Get("X-Asset-Status"))
	}
	if w.Header().Get("Content-Type") != "image/jpeg" {
		t.Errorf("expected Content-Type: image/jpeg, got %s", w.Header().Get("Content-Type"))
	}
	if w.Body.String() != "real-image-content" {
		t.Errorf("expected real-image-content, got %s", w.Body.String())
	}
}

func TestRequestAsset(t *testing.T) {
	_, router, _ := setupAssetTest(t)

	body := `{"type":"image","source_hint":"https://example.com/cover.jpg","entity_type":"file","entity_id":"42"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/assets/request", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("parse response: %v", err)
	}
	if resp["asset_id"] == nil || resp["asset_id"] == "" {
		t.Error("expected non-empty asset_id in response")
	}
	if resp["status"] != "pending" {
		t.Errorf("expected status pending, got %v", resp["status"])
	}
}

func TestRequestAsset_BadRequest(t *testing.T) {
	_, router, _ := setupAssetTest(t)

	body := `{"type":"image"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/assets/request", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestGetByEntity(t *testing.T) {
	_, router, _ := setupAssetTest(t)

	// First create an asset via the request endpoint
	body := `{"type":"image","entity_type":"file","entity_id":"99"}`
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/assets/request", bytes.NewBufferString(body))
	createReq.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, createReq)

	if w1.Code != http.StatusOK {
		t.Fatalf("create failed: %d %s", w1.Code, w1.Body.String())
	}

	// Now query by entity
	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/assets/by-entity/file/99", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, getReq)

	if w2.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w2.Code)
	}

	var assets []interface{}
	if err := json.Unmarshal(w2.Body.Bytes(), &assets); err != nil {
		t.Fatalf("parse response: %v", err)
	}
	if len(assets) != 1 {
		t.Errorf("expected 1 asset, got %d", len(assets))
	}
}

func TestGetByEntity_Empty(t *testing.T) {
	_, router, _ := setupAssetTest(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/assets/by-entity/file/9999", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var assets []interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &assets); err != nil {
		t.Fatalf("parse response: %v", err)
	}
	if len(assets) != 0 {
		t.Errorf("expected 0 assets, got %d", len(assets))
	}
}
