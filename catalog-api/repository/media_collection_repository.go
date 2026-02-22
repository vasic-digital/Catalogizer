package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"catalogizer/database"
	"catalogizer/internal/media/models"
)

// MediaCollectionRepository handles media_collections database operations.
type MediaCollectionRepository struct {
	db *database.DB
}

// NewMediaCollectionRepository creates a new media collection repository.
func NewMediaCollectionRepository(db *database.DB) *MediaCollectionRepository {
	return &MediaCollectionRepository{db: db}
}

// marshalJSONFieldString is a helper to marshal JSON fields to string.
func marshalJSONFieldString(v interface{}) (string, error) {
	s, err := marshalJSONField(v)
	if err != nil {
		return "", err
	}
	if s == nil {
		return "null", nil
	}
	return *s, nil
}

// unmarshalJSONFieldString is a helper to unmarshal JSON fields from string.
func unmarshalJSONFieldString(data string, v interface{}) error {
	if data == "" || data == "null" {
		return nil
	}
	return json.Unmarshal([]byte(data), v)
}

// Create inserts a new media collection and returns the generated ID.
func (r *MediaCollectionRepository) Create(ctx context.Context, coll *models.MediaCollection) (int64, error) {
	externalIDsJSON, err := marshalJSONFieldString(coll.ExternalIDs)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal external_ids: %w", err)
	}

	now := time.Now()
	if coll.CreatedAt.IsZero() {
		coll.CreatedAt = now
	}
	if coll.UpdatedAt.IsZero() {
		coll.UpdatedAt = now
	}

	query := `INSERT INTO media_collections (
		name, collection_type, description, total_items,
		external_ids, cover_url, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	id, err := r.db.InsertReturningID(ctx, query,
		coll.Name, coll.CollectionType, coll.Description, coll.TotalItems,
		externalIDsJSON, coll.CoverURL, coll.CreatedAt, coll.UpdatedAt,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to create media collection: %w", err)
	}

	coll.ID = id
	return id, nil
}

// GetByID retrieves a media collection by its ID.
func (r *MediaCollectionRepository) GetByID(ctx context.Context, id int64) (*models.MediaCollection, error) {
	query := `SELECT id, name, collection_type, description, total_items,
		external_ids, cover_url, created_at, updated_at
		FROM media_collections WHERE id = ?`

	row := r.db.QueryRowContext(ctx, query, id)
	coll, err := r.scanCollection(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("media collection not found")
		}
		return nil, fmt.Errorf("failed to get media collection: %w", err)
	}
	return coll, nil
}

// List returns all media collections with optional pagination.
func (r *MediaCollectionRepository) List(ctx context.Context, limit, offset int) ([]*models.MediaCollection, int, error) {
	// Count total
	var total int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM media_collections").Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count media collections: %w", err)
	}

	query := `SELECT id, name, collection_type, description, total_items,
		external_ids, cover_url, created_at, updated_at
		FROM media_collections ORDER BY id LIMIT ? OFFSET ?`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list media collections: %w", err)
	}
	defer rows.Close()

	var collections []*models.MediaCollection
	for rows.Next() {
		coll, err := r.scanCollection(rows)
		if err != nil {
			continue
		}
		collections = append(collections, coll)
	}

	return collections, total, nil
}

// Update updates an existing media collection.
func (r *MediaCollectionRepository) Update(ctx context.Context, coll *models.MediaCollection) error {
	externalIDsJSON, err := marshalJSONFieldString(coll.ExternalIDs)
	if err != nil {
		return fmt.Errorf("failed to marshal external_ids: %w", err)
	}

	coll.UpdatedAt = time.Now()

	query := `UPDATE media_collections SET
		name = ?, collection_type = ?, description = ?, total_items = ?,
		external_ids = ?, cover_url = ?, updated_at = ?
		WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query,
		coll.Name, coll.CollectionType, coll.Description, coll.TotalItems,
		externalIDsJSON, coll.CoverURL, coll.UpdatedAt, coll.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update media collection: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("media collection not found")
	}
	return nil
}

// Delete removes a media collection by ID.
func (r *MediaCollectionRepository) Delete(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM media_collections WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete media collection: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("media collection not found")
	}
	return nil
}

// scanCollection scans a database row into a MediaCollection struct.
func (r *MediaCollectionRepository) scanCollection(row interface{ Scan(...interface{}) error }) (*models.MediaCollection, error) {
	var (
		id                   int64
		name, collType       string
		description          *string
		totalItems           int
		externalIDsJSON      string
		coverURL             *string
		createdAt, updatedAt time.Time
	)
	err := row.Scan(&id, &name, &collType, &description, &totalItems,
		&externalIDsJSON, &coverURL, &createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}

	var externalIDs map[string]string
	if err := unmarshalJSONFieldString(externalIDsJSON, &externalIDs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal external_ids: %w", err)
	}

	coll := &models.MediaCollection{
		ID:             id,
		Name:           name,
		CollectionType: collType,
		Description:    description,
		TotalItems:     totalItems,
		ExternalIDs:    externalIDs,
		CoverURL:       coverURL,
		CreatedAt:      createdAt,
		UpdatedAt:      updatedAt,
	}
	return coll, nil
}
