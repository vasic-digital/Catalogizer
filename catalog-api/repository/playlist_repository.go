package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"catalogizer/database"
	"catalogizer/models"
)

// PlaylistRepository handles CRUD operations for playlists and playlist items.
type PlaylistRepository struct {
	db *database.DB
}

// NewPlaylistRepository creates a new playlist repository.
func NewPlaylistRepository(db *database.DB) *PlaylistRepository {
	return &PlaylistRepository{db: db}
}

// CreatePlaylist inserts a new playlist and returns its ID.
func (r *PlaylistRepository) CreatePlaylist(playlist *models.Playlist) (int, error) {
	query := `
		INSERT INTO playlists (user_id, name, description, is_public, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	id, err := r.db.InsertReturningID(context.Background(), query,
		playlist.UserID, playlist.Name, playlist.Description,
		playlist.IsPublic, playlist.CreatedAt, playlist.UpdatedAt)
	if err != nil {
		return 0, fmt.Errorf("failed to create playlist: %w", err)
	}

	return int(id), nil
}

// GetPlaylistByID returns a playlist by its ID.
func (r *PlaylistRepository) GetPlaylistByID(playlistID int) (*models.Playlist, error) {
	query := `
		SELECT id, user_id, name, description, is_public, created_at, updated_at
		FROM playlists
		WHERE id = ?
	`

	playlist := &models.Playlist{}
	err := r.db.QueryRow(query, playlistID).Scan(
		&playlist.ID, &playlist.UserID, &playlist.Name, &playlist.Description,
		&playlist.IsPublic, &playlist.CreatedAt, &playlist.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("playlist not found")
		}
		return nil, fmt.Errorf("failed to get playlist: %w", err)
	}

	return playlist, nil
}

// GetUserPlaylists returns all playlists for a given user.
func (r *PlaylistRepository) GetUserPlaylists(userID int, limit, offset int) ([]models.Playlist, error) {
	query := `
		SELECT id, user_id, name, description, is_public, created_at, updated_at
		FROM playlists
		WHERE user_id = ?
		ORDER BY updated_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.Query(query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get user playlists: %w", err)
	}
	defer rows.Close()

	return r.scanPlaylists(rows)
}

// UpdatePlaylist updates an existing playlist.
func (r *PlaylistRepository) UpdatePlaylist(playlist *models.Playlist) error {
	query := `
		UPDATE playlists
		SET name = ?, description = ?, is_public = ?, updated_at = ?
		WHERE id = ?
	`

	_, err := r.db.Exec(query,
		playlist.Name, playlist.Description, playlist.IsPublic,
		playlist.UpdatedAt, playlist.ID)
	return err
}

// DeletePlaylist removes a playlist and its items (cascade).
func (r *PlaylistRepository) DeletePlaylist(playlistID int) error {
	// Delete items first (in case foreign key cascades are not enforced)
	_, _ = r.db.Exec("DELETE FROM playlist_items WHERE playlist_id = ?", playlistID)

	query := `DELETE FROM playlists WHERE id = ?`
	_, err := r.db.Exec(query, playlistID)
	return err
}

// AddPlaylistItem inserts a new item into a playlist.
func (r *PlaylistRepository) AddPlaylistItem(item *models.PlaylistItem) (int, error) {
	// Get next position
	var maxPosition sql.NullInt64
	err := r.db.QueryRow(
		"SELECT MAX(position) FROM playlist_items WHERE playlist_id = ?",
		item.PlaylistID).Scan(&maxPosition)
	if err != nil {
		return 0, fmt.Errorf("failed to get max position: %w", err)
	}

	nextPosition := 0
	if maxPosition.Valid {
		nextPosition = int(maxPosition.Int64) + 1
	}
	item.Position = nextPosition

	query := `
		INSERT INTO playlist_items (playlist_id, entity_id, entity_type, position, added_at)
		VALUES (?, ?, ?, ?, ?)
	`

	id, err := r.db.InsertReturningID(context.Background(), query,
		item.PlaylistID, item.EntityID, item.EntityType,
		item.Position, item.AddedAt)
	if err != nil {
		return 0, fmt.Errorf("failed to add playlist item: %w", err)
	}

	return int(id), nil
}

// GetPlaylistItems returns all items for a given playlist ordered by position.
func (r *PlaylistRepository) GetPlaylistItems(playlistID int) ([]models.PlaylistItem, error) {
	query := `
		SELECT id, playlist_id, entity_id, entity_type, position, added_at
		FROM playlist_items
		WHERE playlist_id = ?
		ORDER BY position ASC
	`

	rows, err := r.db.Query(query, playlistID)
	if err != nil {
		return nil, fmt.Errorf("failed to get playlist items: %w", err)
	}
	defer rows.Close()

	var items []models.PlaylistItem
	for rows.Next() {
		var item models.PlaylistItem
		err := rows.Scan(
			&item.ID, &item.PlaylistID, &item.EntityID, &item.EntityType,
			&item.Position, &item.AddedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan playlist item: %w", err)
		}
		items = append(items, item)
	}

	return items, nil
}

// GetPlaylistItemByID returns a single playlist item by its ID.
func (r *PlaylistRepository) GetPlaylistItemByID(itemID int) (*models.PlaylistItem, error) {
	query := `
		SELECT id, playlist_id, entity_id, entity_type, position, added_at
		FROM playlist_items
		WHERE id = ?
	`

	item := &models.PlaylistItem{}
	err := r.db.QueryRow(query, itemID).Scan(
		&item.ID, &item.PlaylistID, &item.EntityID, &item.EntityType,
		&item.Position, &item.AddedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("playlist item not found")
		}
		return nil, fmt.Errorf("failed to get playlist item: %w", err)
	}

	return item, nil
}

// DeletePlaylistItem removes an item from a playlist.
func (r *PlaylistRepository) DeletePlaylistItem(itemID int) error {
	query := `DELETE FROM playlist_items WHERE id = ?`
	_, err := r.db.Exec(query, itemID)
	return err
}

// CountUserPlaylists returns the total number of playlists for a user.
func (r *PlaylistRepository) CountUserPlaylists(userID int) (int, error) {
	query := `SELECT COUNT(*) FROM playlists WHERE user_id = ?`
	var count int
	err := r.db.QueryRow(query, userID).Scan(&count)
	return count, err
}

// TouchPlaylistUpdatedAt updates the updated_at timestamp for a playlist.
func (r *PlaylistRepository) TouchPlaylistUpdatedAt(playlistID int) error {
	query := `UPDATE playlists SET updated_at = ? WHERE id = ?`
	_, err := r.db.Exec(query, time.Now(), playlistID)
	return err
}

func (r *PlaylistRepository) scanPlaylists(rows *sql.Rows) ([]models.Playlist, error) {
	var playlists []models.Playlist

	for rows.Next() {
		var playlist models.Playlist
		err := rows.Scan(
			&playlist.ID, &playlist.UserID, &playlist.Name, &playlist.Description,
			&playlist.IsPublic, &playlist.CreatedAt, &playlist.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan playlist: %w", err)
		}
		playlists = append(playlists, playlist)
	}

	return playlists, nil
}
