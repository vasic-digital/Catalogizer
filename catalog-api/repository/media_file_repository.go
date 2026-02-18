package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"catalogizer/database"
)

// MediaFileRecord represents a row in the media_files junction table
type MediaFileRecord struct {
	ID          int64
	MediaItemID int64
	FileID      int64
	QualityInfo *string
	Language    *string
	IsPrimary   bool
	CreatedAt   time.Time
}

// DuplicateFileGroup represents a file linked to multiple media items
type DuplicateFileGroup struct {
	FileID    int64
	ItemCount int64
	ItemIDs   []int64
}

// MediaFileRepository handles media_files junction table operations
type MediaFileRepository struct {
	db *database.DB
}

// NewMediaFileRepository creates a new media file repository
func NewMediaFileRepository(db *database.DB) *MediaFileRepository {
	return &MediaFileRepository{db: db}
}

// LinkFileToItem creates an association between a media item and a file
func (r *MediaFileRepository) LinkFileToItem(ctx context.Context, mediaItemID, fileID int64, qualityInfo, language *string, isPrimary bool) (int64, error) {
	query := `
		INSERT INTO media_files (media_item_id, file_id, quality_info, language, is_primary, created_at)
		VALUES (?, ?, ?, ?, ?, ?)`

	id, err := r.db.InsertReturningID(ctx, query,
		mediaItemID, fileID, qualityInfo, language, isPrimary, time.Now().UTC())

	if err != nil {
		return 0, fmt.Errorf("failed to link file to media item: %w", err)
	}

	return id, nil
}

// UnlinkFile removes the association between a media item and a file
func (r *MediaFileRepository) UnlinkFile(ctx context.Context, mediaItemID, fileID int64) error {
	query := `DELETE FROM media_files WHERE media_item_id = ? AND file_id = ?`

	_, err := r.db.ExecContext(ctx, query, mediaItemID, fileID)
	if err != nil {
		return fmt.Errorf("failed to unlink file from media item: %w", err)
	}

	return nil
}

// GetFilesByItem retrieves all media_file records for a given media item
func (r *MediaFileRepository) GetFilesByItem(ctx context.Context, mediaItemID int64) ([]MediaFileRecord, error) {
	query := `
		SELECT id, media_item_id, file_id, quality_info, language, is_primary, created_at
		FROM media_files
		WHERE media_item_id = ?
		ORDER BY is_primary DESC, created_at ASC`

	rows, err := r.db.QueryContext(ctx, query, mediaItemID)
	if err != nil {
		return nil, fmt.Errorf("failed to get files for media item: %w", err)
	}
	defer rows.Close()

	var records []MediaFileRecord
	for rows.Next() {
		var rec MediaFileRecord
		err := rows.Scan(
			&rec.ID, &rec.MediaItemID, &rec.FileID,
			&rec.QualityInfo, &rec.Language, &rec.IsPrimary, &rec.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan media file record: %w", err)
		}
		records = append(records, rec)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating media file rows: %w", err)
	}

	return records, nil
}

// GetItemByFile returns all media item IDs linked to a given file
func (r *MediaFileRepository) GetItemByFile(ctx context.Context, fileID int64) ([]int64, error) {
	query := `
		SELECT media_item_id
		FROM media_files
		WHERE file_id = ?
		ORDER BY created_at ASC`

	rows, err := r.db.QueryContext(ctx, query, fileID)
	if err != nil {
		return nil, fmt.Errorf("failed to get media items for file: %w", err)
	}
	defer rows.Close()

	var itemIDs []int64
	for rows.Next() {
		var itemID int64
		if err := rows.Scan(&itemID); err != nil {
			return nil, fmt.Errorf("failed to scan media item ID: %w", err)
		}
		itemIDs = append(itemIDs, itemID)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating media item rows: %w", err)
	}

	return itemIDs, nil
}

// GetDuplicateFiles finds files that are linked to multiple media items
func (r *MediaFileRepository) GetDuplicateFiles(ctx context.Context) ([]DuplicateFileGroup, error) {
	query := `
		SELECT file_id, COUNT(*) as item_count
		FROM media_files
		GROUP BY file_id
		HAVING COUNT(*) > 1`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query duplicate files: %w", err)
	}
	defer rows.Close()

	var groups []DuplicateFileGroup
	for rows.Next() {
		var group DuplicateFileGroup
		if err := rows.Scan(&group.FileID, &group.ItemCount); err != nil {
			return nil, fmt.Errorf("failed to scan duplicate file group: %w", err)
		}
		groups = append(groups, group)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating duplicate file rows: %w", err)
	}

	// For each duplicate file, fetch the linked item IDs
	itemQuery := `
		SELECT media_item_id
		FROM media_files
		WHERE file_id = ?
		ORDER BY created_at ASC`

	for i := range groups {
		itemRows, err := r.db.QueryContext(ctx, itemQuery, groups[i].FileID)
		if err != nil {
			return nil, fmt.Errorf("failed to query item IDs for file %d: %w", groups[i].FileID, err)
		}

		for itemRows.Next() {
			var itemID int64
			if err := itemRows.Scan(&itemID); err != nil {
				itemRows.Close()
				return nil, fmt.Errorf("failed to scan item ID for file %d: %w", groups[i].FileID, err)
			}
			groups[i].ItemIDs = append(groups[i].ItemIDs, itemID)
		}

		if err = itemRows.Err(); err != nil {
			itemRows.Close()
			return nil, fmt.Errorf("error iterating item IDs for file %d: %w", groups[i].FileID, err)
		}
		itemRows.Close()
	}

	return groups, nil
}

// CountByItem returns the number of files linked to a media item
func (r *MediaFileRepository) CountByItem(ctx context.Context, mediaItemID int64) (int64, error) {
	query := `SELECT COUNT(*) FROM media_files WHERE media_item_id = ?`

	var count int64
	err := r.db.QueryRowContext(ctx, query, mediaItemID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count files for media item: %w", err)
	}

	return count, nil
}

// SetPrimary sets a specific file as the primary file for a media item,
// clearing the primary flag on all other files for that item
func (r *MediaFileRepository) SetPrimary(ctx context.Context, mediaItemID, fileID int64) error {
	// First, clear is_primary for all files of this item
	clearQuery := `UPDATE media_files SET is_primary = ? WHERE media_item_id = ?`
	_, err := r.db.ExecContext(ctx, clearQuery, false, mediaItemID)
	if err != nil {
		return fmt.Errorf("failed to clear primary flag for media item: %w", err)
	}

	// Then, set is_primary for the specified file
	setQuery := `UPDATE media_files SET is_primary = ? WHERE media_item_id = ? AND file_id = ?`
	result, err := r.db.ExecContext(ctx, setQuery, true, mediaItemID, fileID)
	if err != nil {
		return fmt.Errorf("failed to set primary file: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}
