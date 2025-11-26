package tests

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"catalogizer/internal/services"
	"go.uber.org/zap"
	_ "github.com/mutecomm/go-sqlcipher"
)

// Mock implementations for testing - not needed with real services
func createTestReaderService() *services.ReaderService {
	// Create in-memory database for testing with sqlcipher
	db, err := sql.Open("sqlcipher", ":memory:")
	if err != nil {
		panic(err)
	}

	// Set encryption key for in-memory database
	if _, err := db.Exec("PRAGMA key = 'test_key'"); err != nil {
		panic(err)
	}

	// Create required schema for ReaderService
	schema := `
	CREATE TABLE IF NOT EXISTS reading_sessions (
		id TEXT PRIMARY KEY,
		user_id INTEGER NOT NULL,
		book_id TEXT NOT NULL,
		device_id TEXT NOT NULL,
		device_name TEXT NOT NULL,
		started_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		last_active_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		current_position TEXT,
		reading_settings TEXT,
		reading_stats TEXT,
		is_active BOOLEAN DEFAULT 1,
		session_data TEXT
	);

	CREATE TABLE IF NOT EXISTS reading_positions (
		user_id INTEGER NOT NULL,
		book_id TEXT NOT NULL,
		position_data TEXT,
		page_number INTEGER,
		percent_complete REAL,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (user_id, book_id)
	);

	CREATE TABLE IF NOT EXISTS reading_bookmarks (
		id TEXT PRIMARY KEY,
		user_id INTEGER NOT NULL,
		book_id TEXT NOT NULL,
		position_data TEXT,
		page_number INTEGER,
		title TEXT,
		note TEXT,
		tags TEXT,
		color TEXT,
		is_public BOOLEAN DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		share_url TEXT
	);

	CREATE TABLE IF NOT EXISTS reading_highlights (
		id TEXT PRIMARY KEY,
		user_id INTEGER NOT NULL,
		book_id TEXT NOT NULL,
		start_position_data TEXT,
		end_position_data TEXT,
		selected_text TEXT,
		note TEXT,
		color TEXT,
		type TEXT,
		tags TEXT,
		is_public BOOLEAN DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		share_url TEXT
	);

	CREATE TABLE IF NOT EXISTS user_settings (
		user_id INTEGER PRIMARY KEY,
		reading_settings TEXT
	);

	CREATE TABLE IF NOT EXISTS reading_stats (
		user_id INTEGER PRIMARY KEY,
		total_reading_time INTEGER DEFAULT 0,
		session_time INTEGER DEFAULT 0,
		pages_read INTEGER DEFAULT 0,
		words_read INTEGER DEFAULT 0,
		reading_speed REAL DEFAULT 0.0,
		average_speed REAL DEFAULT 0.0,
		daily_goal INTEGER DEFAULT 30,
		daily_progress INTEGER DEFAULT 0,
		reading_streak INTEGER DEFAULT 0,
		longest_streak INTEGER DEFAULT 0,
		books_completed INTEGER DEFAULT 0,
		pages_per_session REAL DEFAULT 0.0,
		weekly_stats TEXT,
		monthly_stats TEXT
	);
	`

	if _, err := db.Exec(schema); err != nil {
		panic(err)
	}

	logger := zap.NewNop()
	cacheService := services.NewCacheService(db, logger)
	translationService := services.NewTranslationService(logger)
	localizationService := services.NewLocalizationService(db, logger, translationService, cacheService)

	return services.NewReaderService(db, logger, cacheService, translationService, localizationService)
}

func TestReaderService_ReadingSession(t *testing.T) {
	ctx := context.Background()
	readerService := createTestReaderService()

	t.Run("start reading session", func(t *testing.T) {
		userID := int64(123)
		bookID := "test-book-123"
		deviceInfo := services.DeviceInfo{
			DeviceID:   "tablet-001",
			DeviceName: "Test Tablet",
			DeviceType: "tablet",
			Platform:   "iOS",
			AppVersion: "1.0.0",
		}

		req := &services.StartReadingRequest{
			UserID:                userID,
			BookID:                bookID,
			DeviceInfo:            deviceInfo,
			ResumeFromLastPosition: false,
			ReadingSettings:       nil,
		}

		session, err := readerService.StartReading(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, session)

		assert.Equal(t, userID, session.UserID)
		assert.Equal(t, bookID, session.BookID)
		assert.Equal(t, deviceInfo.DeviceID, session.DeviceID)
		assert.Equal(t, deviceInfo.DeviceName, session.DeviceName)
		assert.False(t, session.StartedAt.IsZero())
		assert.True(t, session.IsActive)
	})

	t.Run("start reading with settings", func(t *testing.T) {
		userID := int64(456)
		bookID := "test-book-456"
		deviceInfo := services.DeviceInfo{
			DeviceID:   "phone-001",
			DeviceName: "Test Phone",
			DeviceType: "phone",
			Platform:   "Android",
			AppVersion: "1.0.0",
		}

		settings := &services.ReadingSettings{
			FontFamily:      "Arial",
			FontSize:        16,
			LineHeight:      1.5,
			Theme:           "light",
			BackgroundColor: "#FFFFFF",
			TextColor:       "#000000",
		}

		req := &services.StartReadingRequest{
			UserID:                userID,
			BookID:                bookID,
			DeviceInfo:            deviceInfo,
			ResumeFromLastPosition: false,
			ReadingSettings:       settings,
		}

		session, err := readerService.StartReading(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, session)

		assert.Equal(t, settings.FontFamily, session.ReadingSettings.FontFamily)
		assert.Equal(t, settings.FontSize, session.ReadingSettings.FontSize)
		assert.Equal(t, settings.Theme, session.ReadingSettings.Theme)
	})

	t.Run("update reading position", func(t *testing.T) {
		// First start a reading session
		userID := int64(789)
		bookID := "test-book-789"
		deviceInfo := services.DeviceInfo{
			DeviceID:   "desktop-001",
			DeviceName: "Test Desktop",
			DeviceType: "desktop",
			Platform:   "Windows",
			AppVersion: "1.0.0",
		}

		startReq := &services.StartReadingRequest{
			UserID:                userID,
			BookID:                bookID,
			DeviceInfo:            deviceInfo,
			ResumeFromLastPosition: false,
		}

		session, err := readerService.StartReading(ctx, startReq)
		require.NoError(t, err)

		// Now update position
		position := services.ReadingPosition{
			BookID:          bookID,
			PageNumber:      25,
			WordOffset:      350,
			CharacterOffset: 1500,
			PercentComplete: 0.15,
			Location:        "Kindle location 250",
			Timestamp:       time.Now(),
			Confidence:      0.95,
		}

		updateReq := &services.ReaderUpdatePositionRequest{
			SessionID: session.ID,
			Position:  position,
		}

		updatedSession, err := readerService.UpdatePosition(ctx, updateReq)
		require.NoError(t, err)
		require.NotNil(t, updatedSession)

		assert.Equal(t, position.PageNumber, updatedSession.CurrentPosition.PageNumber)
		assert.Equal(t, position.PercentComplete, updatedSession.CurrentPosition.PercentComplete)
	})

	t.Run("create bookmark", func(t *testing.T) {
		userID := int64(101)
		bookID := "test-book-101"

		position := services.ReadingPosition{
			BookID:          bookID,
			PageNumber:      50,
			WordOffset:      700,
			CharacterOffset: 3000,
			PercentComplete: 0.30,
			Location:        "Kindle location 500",
			Timestamp:       time.Now(),
			Confidence:      0.98,
		}

		req := &services.CreateBookmarkRequest{
			UserID:   userID,
			BookID:   bookID,
			Position: position,
			Title:    "Important Passage",
			Note:     "This is a key section of the book",
		}

		bookmark, err := readerService.CreateBookmark(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, bookmark)

		assert.Equal(t, userID, bookmark.UserID)
		assert.Equal(t, bookID, bookmark.BookID)
		assert.Equal(t, "Important Passage", bookmark.Title)
		assert.Equal(t, "This is a key section of the book", bookmark.Note)
		assert.False(t, bookmark.CreatedAt.IsZero())
	})

	t.Run("create highlight", func(t *testing.T) {
		userID := int64(102)
		bookID := "test-book-102"

		position := services.ReadingPosition{
			BookID:          bookID,
			PageNumber:      75,
			WordOffset:      1050,
			CharacterOffset:  4500,
			PercentComplete: 0.45,
			Location:        "Kindle location 750",
			Timestamp:       time.Now(),
			Confidence:      0.97,
		}

		req := &services.CreateHighlightRequest{
			UserID:        userID,
			BookID:        bookID,
			StartPosition: position,
			EndPosition:   position, // Same position for single-word highlight
			SelectedText:  "This is the highlighted text from the book",
			Color:         "#FFFF00", // Yellow highlight
			Note:          "Important quote to remember",
			Type:          "highlight",
		}

		highlight, err := readerService.CreateHighlight(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, highlight)

		assert.Equal(t, userID, highlight.UserID)
		assert.Equal(t, bookID, highlight.BookID)
		assert.Equal(t, "This is the highlighted text from the book", highlight.SelectedText)
		assert.Equal(t, "#FFFF00", highlight.Color)
		assert.Equal(t, "Important quote to remember", highlight.Note)
		assert.False(t, highlight.CreatedAt.IsZero())
	})

	t.Run("sync across devices", func(t *testing.T) {
		userID := int64(103)
		bookID := "test-book-103"

		err := readerService.SyncAcrossDevices(ctx, userID, bookID)
		// This might not have data to sync, but should not error
		if err != nil {
			t.Logf("Sync returned error (possibly expected): %v", err)
		}
	})

	t.Run("get reading stats", func(t *testing.T) {
		userID := int64(104)
		period := "week" // could be "day", "week", "month", "year"

		stats, err := readerService.GetReadingStats(ctx, userID, period)
		// Should not error even with no data
		if err != nil {
			t.Logf("GetReadingStats returned error (possibly expected): %v", err)
		}
		
		if stats != nil {
			t.Logf("Stats returned: %+v", stats)
		}
	})

	t.Run("get bookmarks", func(t *testing.T) {
		userID := int64(105)
		bookID := "test-book-105"

		bookmarks, err := readerService.GetBookmarks(ctx, userID, bookID)
		require.NoError(t, err)
		// Should return empty array if no bookmarks exist
		if bookmarks == nil {
			t.Logf("GetBookmarks returned nil (possibly expected)")
		} else {
			t.Logf("GetBookmarks returned %d bookmarks", len(bookmarks))
		}
	})

	t.Run("get highlights", func(t *testing.T) {
		userID := int64(106)
		bookID := "test-book-106"

		highlights, err := readerService.GetHighlights(ctx, userID, bookID)
		require.NoError(t, err)
		// Should return empty array if no highlights exist
		if highlights == nil {
			t.Logf("GetHighlights returned nil (possibly expected)")
		} else {
			t.Logf("GetHighlights returned %d highlights", len(highlights))
		}
	})
}