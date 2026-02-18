package repository

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	"catalogizer/config"
	"catalogizer/database"

	"digital.vasic.assets/pkg/asset"
	_ "github.com/mutecomm/go-sqlcipher"
)

func setupTestAssetDB(t *testing.T) *database.DB {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := database.NewConnection(&config.DatabaseConfig{
		Path: dbPath,
	})
	if err != nil {
		t.Fatalf("failed to create test DB: %v", err)
	}

	ctx := context.Background()
	if err := db.RunMigrations(ctx); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	t.Cleanup(func() { db.Close() })
	return db
}

func TestAssetRepository_CreateAndGet(t *testing.T) {
	db := setupTestAssetDB(t)
	repo := NewAssetRepository(db)
	ctx := context.Background()

	a := asset.New(asset.TypeImage, "file", "42")
	a.SourceHint = "https://example.com/cover.jpg"

	err := repo.CreateAsset(ctx, a)
	if err != nil {
		t.Fatalf("CreateAsset failed: %v", err)
	}

	got, err := repo.GetAsset(ctx, a.ID)
	if err != nil {
		t.Fatalf("GetAsset failed: %v", err)
	}
	if got == nil {
		t.Fatal("expected asset, got nil")
	}
	if got.ID != a.ID {
		t.Errorf("expected ID %s, got %s", a.ID, got.ID)
	}
	if got.Type != asset.TypeImage {
		t.Errorf("expected type image, got %s", got.Type)
	}
	if got.Status != asset.StatusPending {
		t.Errorf("expected status pending, got %s", got.Status)
	}
	if got.EntityType != "file" {
		t.Errorf("expected entity_type file, got %s", got.EntityType)
	}
	if got.EntityID != "42" {
		t.Errorf("expected entity_id 42, got %s", got.EntityID)
	}
}

func TestAssetRepository_Update(t *testing.T) {
	db := setupTestAssetDB(t)
	repo := NewAssetRepository(db)
	ctx := context.Background()

	a := asset.New(asset.TypeImage, "file", "1")
	if err := repo.CreateAsset(ctx, a); err != nil {
		t.Fatalf("CreateAsset failed: %v", err)
	}

	a.MarkReady("image/jpeg", 1024)
	if err := repo.UpdateAsset(ctx, a); err != nil {
		t.Fatalf("UpdateAsset failed: %v", err)
	}

	got, err := repo.GetAsset(ctx, a.ID)
	if err != nil {
		t.Fatalf("GetAsset failed: %v", err)
	}
	if got.Status != asset.StatusReady {
		t.Errorf("expected status ready, got %s", got.Status)
	}
	if got.ContentType != "image/jpeg" {
		t.Errorf("expected content_type image/jpeg, got %s", got.ContentType)
	}
	if got.Size != 1024 {
		t.Errorf("expected size 1024, got %d", got.Size)
	}
}

func TestAssetRepository_FindByEntity(t *testing.T) {
	db := setupTestAssetDB(t)
	repo := NewAssetRepository(db)
	ctx := context.Background()

	// Create two assets for the same entity
	a1 := asset.New(asset.TypeImage, "file", "100")
	a2 := asset.New(asset.TypeAudioCover, "file", "100")
	a3 := asset.New(asset.TypeImage, "file", "200")

	for _, a := range []*asset.Asset{a1, a2, a3} {
		if err := repo.CreateAsset(ctx, a); err != nil {
			t.Fatalf("CreateAsset failed: %v", err)
		}
	}

	assets, err := repo.FindByEntity(ctx, "file", "100")
	if err != nil {
		t.Fatalf("FindByEntity failed: %v", err)
	}
	if len(assets) != 2 {
		t.Errorf("expected 2 assets for file/100, got %d", len(assets))
	}
}

func TestAssetRepository_FindPending(t *testing.T) {
	db := setupTestAssetDB(t)
	repo := NewAssetRepository(db)
	ctx := context.Background()

	pending := asset.New(asset.TypeImage, "file", "1")
	ready := asset.New(asset.TypeImage, "file", "2")
	ready.MarkReady("image/png", 100)

	for _, a := range []*asset.Asset{pending, ready} {
		if err := repo.CreateAsset(ctx, a); err != nil {
			t.Fatalf("CreateAsset failed: %v", err)
		}
	}

	assets, err := repo.FindPending(ctx, 10)
	if err != nil {
		t.Fatalf("FindPending failed: %v", err)
	}
	if len(assets) != 1 {
		t.Errorf("expected 1 pending asset, got %d", len(assets))
	}
	if len(assets) > 0 && assets[0].ID != pending.ID {
		t.Errorf("expected pending asset ID %s, got %s", pending.ID, assets[0].ID)
	}
}

func TestAssetRepository_GetNotFound(t *testing.T) {
	db := setupTestAssetDB(t)
	repo := NewAssetRepository(db)

	got, err := repo.GetAsset(context.Background(), "nonexistent-id")
	if err != nil {
		t.Fatalf("GetAsset should not error for missing: %v", err)
	}
	if got != nil {
		t.Error("expected nil for nonexistent asset")
	}
}

// Suppress unused import
var _ = sql.ErrNoRows
