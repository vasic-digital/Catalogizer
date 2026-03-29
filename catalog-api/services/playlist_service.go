package services

import (
	"fmt"
	"time"

	"catalogizer/models"
	"catalogizer/repository"
)

// PlaylistService provides business logic for playlist operations.
type PlaylistService struct {
	playlistRepo *repository.PlaylistRepository
}

// NewPlaylistService creates a new playlist service.
func NewPlaylistService(playlistRepo *repository.PlaylistRepository) *PlaylistService {
	return &PlaylistService{
		playlistRepo: playlistRepo,
	}
}

// CreatePlaylist creates a new playlist for the given user.
func (s *PlaylistService) CreatePlaylist(userID int, req *models.CreatePlaylistRequest) (*models.Playlist, error) {
	if s.playlistRepo == nil {
		return nil, fmt.Errorf("playlist repository not configured")
	}

	if req.Name == "" {
		return nil, fmt.Errorf("playlist name is required")
	}

	now := time.Now()
	playlist := &models.Playlist{
		UserID:      userID,
		Name:        req.Name,
		Description: req.Description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	id, err := s.playlistRepo.CreatePlaylist(playlist)
	if err != nil {
		return nil, fmt.Errorf("failed to create playlist: %w", err)
	}

	playlist.ID = id
	return playlist, nil
}

// GetPlaylist returns a playlist by ID, verifying the user has access.
func (s *PlaylistService) GetPlaylist(userID int, playlistID int) (*models.Playlist, error) {
	if s.playlistRepo == nil {
		return nil, fmt.Errorf("playlist repository not configured")
	}

	playlist, err := s.playlistRepo.GetPlaylistByID(playlistID)
	if err != nil {
		return nil, fmt.Errorf("playlist not found: %w", err)
	}

	if playlist.UserID != userID && !playlist.IsPublic {
		return nil, fmt.Errorf("unauthorized to access this playlist")
	}

	// Load items
	items, err := s.playlistRepo.GetPlaylistItems(playlistID)
	if err != nil {
		return nil, fmt.Errorf("failed to load playlist items: %w", err)
	}
	playlist.Items = items

	return playlist, nil
}

// GetUserPlaylists returns all playlists owned by the user.
func (s *PlaylistService) GetUserPlaylists(userID int, limit, offset int) ([]models.Playlist, error) {
	if s.playlistRepo == nil {
		return nil, fmt.Errorf("playlist repository not configured")
	}

	return s.playlistRepo.GetUserPlaylists(userID, limit, offset)
}

// UpdatePlaylist updates an existing playlist.
func (s *PlaylistService) UpdatePlaylist(userID int, playlistID int, req *models.UpdatePlaylistRequest) (*models.Playlist, error) {
	if s.playlistRepo == nil {
		return nil, fmt.Errorf("playlist repository not configured")
	}

	playlist, err := s.playlistRepo.GetPlaylistByID(playlistID)
	if err != nil {
		return nil, fmt.Errorf("playlist not found: %w", err)
	}

	if playlist.UserID != userID {
		return nil, fmt.Errorf("unauthorized to update this playlist")
	}

	if req.Name != nil {
		playlist.Name = *req.Name
	}
	if req.Description != nil {
		playlist.Description = *req.Description
	}
	if req.IsPublic != nil {
		playlist.IsPublic = *req.IsPublic
	}
	playlist.UpdatedAt = time.Now()

	if err := s.playlistRepo.UpdatePlaylist(playlist); err != nil {
		return nil, fmt.Errorf("failed to update playlist: %w", err)
	}

	return playlist, nil
}

// DeletePlaylist removes a playlist owned by the user.
func (s *PlaylistService) DeletePlaylist(userID int, playlistID int) error {
	if s.playlistRepo == nil {
		return fmt.Errorf("playlist repository not configured")
	}

	playlist, err := s.playlistRepo.GetPlaylistByID(playlistID)
	if err != nil {
		return fmt.Errorf("playlist not found: %w", err)
	}

	if playlist.UserID != userID {
		return fmt.Errorf("unauthorized to delete this playlist")
	}

	return s.playlistRepo.DeletePlaylist(playlistID)
}

// AddItem adds an entity to a playlist.
func (s *PlaylistService) AddItem(userID int, playlistID int, req *models.AddPlaylistItemRequest) (*models.PlaylistItem, error) {
	if s.playlistRepo == nil {
		return nil, fmt.Errorf("playlist repository not configured")
	}

	playlist, err := s.playlistRepo.GetPlaylistByID(playlistID)
	if err != nil {
		return nil, fmt.Errorf("playlist not found: %w", err)
	}

	if playlist.UserID != userID {
		return nil, fmt.Errorf("unauthorized to modify this playlist")
	}

	item := &models.PlaylistItem{
		PlaylistID: playlistID,
		EntityID:   req.EntityID,
		EntityType: req.EntityType,
		AddedAt:    time.Now(),
	}

	id, err := s.playlistRepo.AddPlaylistItem(item)
	if err != nil {
		return nil, fmt.Errorf("failed to add item to playlist: %w", err)
	}

	item.ID = id

	// Update playlist timestamp
	_ = s.playlistRepo.TouchPlaylistUpdatedAt(playlistID)

	return item, nil
}

// RemoveItem removes an item from a playlist.
func (s *PlaylistService) RemoveItem(userID int, playlistID int, itemID int) error {
	if s.playlistRepo == nil {
		return fmt.Errorf("playlist repository not configured")
	}

	playlist, err := s.playlistRepo.GetPlaylistByID(playlistID)
	if err != nil {
		return fmt.Errorf("playlist not found: %w", err)
	}

	if playlist.UserID != userID {
		return fmt.Errorf("unauthorized to modify this playlist")
	}

	item, err := s.playlistRepo.GetPlaylistItemByID(itemID)
	if err != nil {
		return fmt.Errorf("playlist item not found: %w", err)
	}

	if item.PlaylistID != playlistID {
		return fmt.Errorf("item does not belong to this playlist")
	}

	if err := s.playlistRepo.DeletePlaylistItem(itemID); err != nil {
		return fmt.Errorf("failed to remove item from playlist: %w", err)
	}

	// Update playlist timestamp
	_ = s.playlistRepo.TouchPlaylistUpdatedAt(playlistID)

	return nil
}
