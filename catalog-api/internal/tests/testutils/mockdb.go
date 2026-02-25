package testutils

import (
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

// MockDB creates a sqlmock database connection for testing
func MockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	return db, mock
}

// MockTime returns a fixed time for testing
func MockTime() time.Time {
	return time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
}

// MockUser creates a test user struct
type MockUser struct {
	ID        int64
	Username  string
	Email     string
	Password  string
	RoleID    int
	IsActive  bool
	CreatedAt time.Time
}

// NewMockUser creates a new mock user
func NewMockUser() MockUser {
	return MockUser{
		ID:        1,
		Username:  "testuser",
		Email:     "test@example.com",
		Password:  "hashed_password",
		RoleID:    1,
		IsActive:  true,
		CreatedAt: MockTime(),
	}
}

// MockMediaItem creates a test media item struct
type MockMediaItem struct {
	ID          int64
	Title       string
	Type        string
	Path        string
	Duration    int
	CreatedAt   time.Time
	UserID      int64
	StorageRoot int64
}

// NewMockMediaItem creates a new mock media item
func NewMockMediaItem() MockMediaItem {
	return MockMediaItem{
		ID:          1,
		Title:       "Test Movie",
		Type:        "movie",
		Path:        "/movies/test.mp4",
		Duration:    7200,
		CreatedAt:   MockTime(),
		UserID:      1,
		StorageRoot: 1,
	}
}

// MockStorageRoot creates a test storage root struct
type MockStorageRoot struct {
	ID       int64
	Name     string
	Protocol string
	Host     string
	Path     string
	Enabled  bool
}

// NewMockStorageRoot creates a new mock storage root
func NewMockStorageRoot() MockStorageRoot {
	return MockStorageRoot{
		ID:       1,
		Name:     "Test Storage",
		Protocol: "local",
		Host:     "localhost",
		Path:     "/media",
		Enabled:  true,
	}
}
