package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"catalogizer/database"
	"catalogizer/internal/media/models"
)

// ExternalMetadataRepository handles external_metadata table operations.
type ExternalMetadataRepository struct {
	db *database.DB
}

// NewExternalMetadataRepository creates a new external metadata repository.
func NewExternalMetadataRepository(db *database.DB) *ExternalMetadataRepository {
	return &ExternalMetadataRepository{db: db}
}

// Create inserts external metadata and returns its ID.
func (r *ExternalMetadataRepository) Create(ctx context.Context, em *models.ExternalMetadata) (int64, error) {
	query := `INSERT INTO external_metadata (
		media_item_id, provider, external_id, data, rating,
		review_url, cover_url, trailer_url, last_fetched
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	now := time.Now()
	id, err := r.db.InsertReturningID(ctx, query,
		em.MediaItemID, em.Provider, em.ExternalID, em.Data, em.Rating,
		em.ReviewURL, em.CoverURL, em.TrailerURL, now,
	)
	if err != nil {
		return 0, fmt.Errorf("insert external metadata: %w", err)
	}
	em.ID = id
	em.LastFetched = now
	return id, nil
}

// GetByItem returns all external metadata for a media item.
func (r *ExternalMetadataRepository) GetByItem(ctx context.Context, mediaItemID int64) ([]*models.ExternalMetadata, error) {
	query := `SELECT id, media_item_id, provider, external_id, data, rating,
		review_url, cover_url, trailer_url, last_fetched
	FROM external_metadata WHERE media_item_id = ?
	ORDER BY provider`

	rows, err := r.db.QueryContext(ctx, query, mediaItemID)
	if err != nil {
		return nil, fmt.Errorf("get external metadata by item: %w", err)
	}
	defer rows.Close()

	var items []*models.ExternalMetadata
	for rows.Next() {
		em := &models.ExternalMetadata{}
		if err := rows.Scan(
			&em.ID, &em.MediaItemID, &em.Provider, &em.ExternalID, &em.Data,
			&em.Rating, &em.ReviewURL, &em.CoverURL, &em.TrailerURL, &em.LastFetched,
		); err != nil {
			return nil, err
		}
		items = append(items, em)
	}
	return items, rows.Err()
}

// GetByProvider returns external metadata for a specific provider and external ID.
func (r *ExternalMetadataRepository) GetByProvider(ctx context.Context, provider, externalID string) (*models.ExternalMetadata, error) {
	query := `SELECT id, media_item_id, provider, external_id, data, rating,
		review_url, cover_url, trailer_url, last_fetched
	FROM external_metadata WHERE provider = ? AND external_id = ? LIMIT 1`

	em := &models.ExternalMetadata{}
	err := r.db.QueryRowContext(ctx, query, provider, externalID).Scan(
		&em.ID, &em.MediaItemID, &em.Provider, &em.ExternalID, &em.Data,
		&em.Rating, &em.ReviewURL, &em.CoverURL, &em.TrailerURL, &em.LastFetched,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get by provider: %w", err)
	}
	return em, nil
}

// Upsert creates or updates external metadata by provider + media_item_id.
func (r *ExternalMetadataRepository) Upsert(ctx context.Context, em *models.ExternalMetadata) error {
	existing, err := r.findByItemAndProvider(ctx, em.MediaItemID, em.Provider)
	if err != nil {
		return err
	}

	if existing != nil {
		query := `UPDATE external_metadata SET
			external_id = ?, data = ?, rating = ?, review_url = ?,
			cover_url = ?, trailer_url = ?, last_fetched = ?
		WHERE id = ?`
		_, err := r.db.ExecContext(ctx, query,
			em.ExternalID, em.Data, em.Rating, em.ReviewURL,
			em.CoverURL, em.TrailerURL, time.Now(), existing.ID,
		)
		if err != nil {
			return fmt.Errorf("update external metadata: %w", err)
		}
		em.ID = existing.ID
		return nil
	}

	_, err = r.Create(ctx, em)
	return err
}

// Delete removes external metadata by ID.
func (r *ExternalMetadataRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM external_metadata WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("delete external metadata: %w", err)
	}
	return nil
}

func (r *ExternalMetadataRepository) findByItemAndProvider(ctx context.Context, mediaItemID int64, provider string) (*models.ExternalMetadata, error) {
	query := `SELECT id, media_item_id, provider, external_id, data, rating,
		review_url, cover_url, trailer_url, last_fetched
	FROM external_metadata WHERE media_item_id = ? AND provider = ? LIMIT 1`

	em := &models.ExternalMetadata{}
	err := r.db.QueryRowContext(ctx, query, mediaItemID, provider).Scan(
		&em.ID, &em.MediaItemID, &em.Provider, &em.ExternalID, &em.Data,
		&em.Rating, &em.ReviewURL, &em.CoverURL, &em.TrailerURL, &em.LastFetched,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("find by item and provider: %w", err)
	}
	return em, nil
}
