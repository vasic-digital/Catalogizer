package services

import (
	"catalogizer/database"
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	_ "github.com/mutecomm/go-sqlcipher"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// setupReaderTestDB creates an in-memory SQLite database with all tables
// required by the ReaderService methods.
func setupReaderTestDB(t *testing.T) *database.DB {
	t.Helper()

	sqlDB, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)

	db := database.WrapDB(sqlDB, database.DialectSQLite)

	tables := []string{
		`CREATE TABLE IF NOT EXISTS reading_sessions (
			id TEXT PRIMARY KEY,
			user_id INTEGER NOT NULL,
			book_id TEXT NOT NULL,
			device_id TEXT NOT NULL,
			device_name TEXT,
			started_at DATETIME,
			last_active_at DATETIME,
			current_position TEXT,
			reading_settings TEXT,
			reading_stats TEXT,
			is_active INTEGER DEFAULT 1,
			session_data TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS reading_positions (
			user_id INTEGER NOT NULL,
			book_id TEXT NOT NULL,
			position_data TEXT,
			page_number INTEGER,
			percent_complete REAL,
			timestamp DATETIME,
			PRIMARY KEY (user_id, book_id)
		)`,
		`CREATE TABLE IF NOT EXISTS reading_bookmarks (
			id TEXT PRIMARY KEY,
			user_id INTEGER NOT NULL,
			book_id TEXT NOT NULL,
			position_data TEXT,
			title TEXT,
			note TEXT,
			tags TEXT,
			color TEXT,
			created_at DATETIME,
			updated_at DATETIME,
			is_public INTEGER DEFAULT 0,
			share_url TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS reading_highlights (
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
			created_at DATETIME,
			updated_at DATETIME,
			is_public INTEGER DEFAULT 0,
			share_url TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS reading_history (
			user_id INTEGER NOT NULL,
			book_id TEXT NOT NULL,
			last_read_at DATETIME,
			read_count INTEGER DEFAULT 0,
			PRIMARY KEY (user_id, book_id)
		)`,
		`CREATE TABLE IF NOT EXISTS reading_stats (
			user_id INTEGER PRIMARY KEY,
			total_reading_time INTEGER DEFAULT 0,
			pages_read INTEGER DEFAULT 0,
			words_read INTEGER DEFAULT 0,
			average_speed REAL DEFAULT 0,
			books_completed INTEGER DEFAULT 0
		)`,
		`CREATE TABLE IF NOT EXISTS user_reading_settings (
			user_id INTEGER PRIMARY KEY,
			settings_data TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS user_reading_goals (
			user_id INTEGER PRIMARY KEY,
			daily_goal_minutes INTEGER DEFAULT 30
		)`,
	}

	for _, ddl := range tables {
		_, err := db.Exec(ddl)
		require.NoError(t, err)
	}

	return db
}

func TestNewReaderService(t *testing.T) {
	mockLogger := zap.NewNop()
	service := NewReaderService(nil, mockLogger, nil, nil, nil)

	assert.NotNil(t, service)
}

func TestNewReaderService_AllFields(t *testing.T) {
	db := setupReaderTestDB(t)
	defer db.Close()

	logger := zap.NewNop()
	cache := &CacheService{}
	translation := &TranslationService{}
	localization := &LocalizationService{}

	service := NewReaderService(db, logger, cache, translation, localization)

	assert.NotNil(t, service)
	assert.Equal(t, db, service.db)
	assert.Equal(t, logger, service.logger)
	assert.Equal(t, cache, service.cacheService)
	assert.Equal(t, translation, service.translationService)
	assert.Equal(t, localization, service.localizationService)
}

func TestNewReaderService_NilDependencies(t *testing.T) {
	db := setupReaderTestDB(t)
	defer db.Close()

	logger := zap.NewNop()
	service := NewReaderService(db, logger, nil, nil, nil)

	assert.NotNil(t, service)
	assert.Equal(t, db, service.db)
	assert.Nil(t, service.cacheService)
	assert.Nil(t, service.translationService)
	assert.Nil(t, service.localizationService)
}

func TestReaderService_GenerateSessionID(t *testing.T) {
	mockLogger := zap.NewNop()
	service := NewReaderService(nil, mockLogger, nil, nil, nil)

	id := service.generateSessionID(1, "book_123", "device_456")
	assert.NotEmpty(t, id)
	assert.Contains(t, id, "session_")
	assert.Contains(t, id, "1")
	assert.Contains(t, id, "book_123")
	assert.Contains(t, id, "device_456")
}

func TestReaderService_GenerateSessionID_TableDriven(t *testing.T) {
	service := NewReaderService(nil, zap.NewNop(), nil, nil, nil)

	tests := []struct {
		name     string
		userID   int64
		bookID   string
		deviceID string
	}{
		{
			name:     "standard input",
			userID:   1,
			bookID:   "book-123",
			deviceID: "device-abc",
		},
		{
			name:     "large user ID",
			userID:   9999999,
			bookID:   "my-book",
			deviceID: "phone-01",
		},
		{
			name:     "empty strings",
			userID:   0,
			bookID:   "",
			deviceID: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id := service.generateSessionID(tt.userID, tt.bookID, tt.deviceID)

			assert.True(t, strings.HasPrefix(id, "session_"), "session ID should start with 'session_'")
			assert.Contains(t, id, tt.bookID)
			assert.Contains(t, id, tt.deviceID)
		})
	}
}

func TestReaderService_GenerateSessionID_Uniqueness(t *testing.T) {
	service := NewReaderService(nil, zap.NewNop(), nil, nil, nil)

	id1 := service.generateSessionID(1, "book-1", "device-1")
	// Sleep briefly to ensure different Unix timestamp
	time.Sleep(1100 * time.Millisecond)
	id2 := service.generateSessionID(1, "book-1", "device-1")

	assert.NotEqual(t, id1, id2, "IDs generated at different times should differ")
}

func TestReaderService_GenerateBookmarkID(t *testing.T) {
	mockLogger := zap.NewNop()
	service := NewReaderService(nil, mockLogger, nil, nil, nil)

	position := ReadingPosition{PageNumber: 42}
	id := service.generateBookmarkID(1, "book_123", position)
	assert.NotEmpty(t, id)
	assert.Contains(t, id, "bookmark_")
}

func TestReaderService_GenerateBookmarkID_TableDriven(t *testing.T) {
	service := NewReaderService(nil, zap.NewNop(), nil, nil, nil)

	tests := []struct {
		name       string
		userID     int64
		bookID     string
		pageNumber int
	}{
		{
			name:       "page 1",
			userID:     1,
			bookID:     "book-abc",
			pageNumber: 1,
		},
		{
			name:       "page 500",
			userID:     42,
			bookID:     "novel-xyz",
			pageNumber: 500,
		},
		{
			name:       "page zero",
			userID:     10,
			bookID:     "intro",
			pageNumber: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			position := ReadingPosition{PageNumber: tt.pageNumber}
			id := service.generateBookmarkID(tt.userID, tt.bookID, position)

			assert.True(t, strings.HasPrefix(id, "bookmark_"), "bookmark ID should start with 'bookmark_'")
			assert.Contains(t, id, tt.bookID)
		})
	}
}

func TestReaderService_GenerateHighlightID(t *testing.T) {
	mockLogger := zap.NewNop()
	service := NewReaderService(nil, mockLogger, nil, nil, nil)

	position := ReadingPosition{PageNumber: 42}
	id := service.generateHighlightID(1, "book_123", position)
	assert.NotEmpty(t, id)
	assert.Contains(t, id, "highlight_")
}

func TestReaderService_GenerateHighlightID_WithBookID(t *testing.T) {
	service := NewReaderService(nil, zap.NewNop(), nil, nil, nil)

	position := ReadingPosition{PageNumber: 42}
	id := service.generateHighlightID(7, "book-test", position)

	assert.True(t, strings.HasPrefix(id, "highlight_"), "highlight ID should start with 'highlight_'")
	assert.Contains(t, id, "book-test")
}

func TestReaderService_GenerateShareURL(t *testing.T) {
	mockLogger := zap.NewNop()
	service := NewReaderService(nil, mockLogger, nil, nil, nil)

	tests := []struct {
		name        string
		itemType    string
		itemID      string
		expectedURL string
	}{
		{
			name:        "bookmark share URL",
			itemType:    "bookmark",
			itemID:      "bm_123",
			expectedURL: "https://catalogizer.com/share/bookmark/bm_123",
		},
		{
			name:        "highlight share URL",
			itemType:    "highlight",
			itemID:      "hl_456",
			expectedURL: "https://catalogizer.com/share/highlight/hl_456",
		},
		{
			name:        "annotation share URL",
			itemType:    "annotation",
			itemID:      "anno-789",
			expectedURL: "https://catalogizer.com/share/annotation/anno-789",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := service.generateShareURL(tt.itemType, tt.itemID)

			assert.Equal(t, tt.expectedURL, url)
			assert.True(t, strings.HasPrefix(url, "https://catalogizer.com/share/"))
			assert.Contains(t, url, tt.itemType)
			assert.Contains(t, url, tt.itemID)
		})
	}
}

func TestReaderService_GetDefaultReadingSettings(t *testing.T) {
	mockLogger := zap.NewNop()
	service := NewReaderService(nil, mockLogger, nil, nil, nil)

	settings := service.getDefaultReadingSettings()

	assert.Equal(t, "serif", settings.FontFamily)
	assert.Equal(t, 16, settings.FontSize)
	assert.Equal(t, 1.5, settings.LineHeight)
	assert.Equal(t, "justify", settings.TextAlign)
	assert.Equal(t, "light", settings.Theme)
	assert.Equal(t, "#ffffff", settings.BackgroundColor)
	assert.Equal(t, "#000000", settings.TextColor)
	assert.Equal(t, 1, settings.ColumnsPerPage)
	assert.Equal(t, "slide", settings.PageTransition)
	assert.False(t, settings.AutoScroll)
	assert.Equal(t, 5, settings.AutoScrollSpeed)
	assert.Equal(t, "day", settings.ReadingMode)
	assert.Equal(t, 1.0, settings.Brightness)
	assert.True(t, settings.Hyphenation)
	assert.True(t, settings.Justification)

	// Page margins
	assert.Equal(t, 20, settings.PageMargins.Top)
	assert.Equal(t, 20, settings.PageMargins.Bottom)
	assert.Equal(t, 15, settings.PageMargins.Left)
	assert.Equal(t, 15, settings.PageMargins.Right)

	// Blue light filter
	assert.False(t, settings.BlueLight.Enabled)
	assert.Equal(t, 0.3, settings.BlueLight.Intensity)

	// Status bar
	assert.True(t, settings.StatusBar.Visible)
	assert.True(t, settings.StatusBar.ShowProgress)
	assert.Equal(t, "bottom", settings.StatusBar.Position)

	// Gestures
	assert.True(t, settings.Gestures.TapToTurn)
	assert.True(t, settings.Gestures.SwipeToTurn)
	assert.False(t, settings.Gestures.VolumeKeys)
}

func TestReaderService_UpdateReadingStats(t *testing.T) {
	mockLogger := zap.NewNop()
	service := NewReaderService(nil, mockLogger, nil, nil, nil)

	tests := []struct {
		name          string
		readingTime   int64
		pages         int
		words         int
		expectedWPM   float64
		expectNonZero bool
	}{
		{
			name:          "normal reading session",
			readingTime:   300, // 5 minutes
			pages:         5,
			words:         1500,
			expectedWPM:   300.0, // 1500 / (300/60) = 1500/5
			expectNonZero: true,
		},
		{
			name:          "zero reading time",
			readingTime:   0,
			pages:         0,
			words:         0,
			expectedWPM:   0.0,
			expectNonZero: false,
		},
		{
			name:          "fast reading",
			readingTime:   60, // 1 minute
			pages:         3,
			words:         600,
			expectedWPM:   600.0, // 600 / (60/60) = 600/1
			expectNonZero: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stats := &ReadingStats{}
			service.updateReadingStats(stats, tt.readingTime, tt.pages, tt.words)

			assert.Equal(t, tt.readingTime, stats.SessionTime)
			assert.Equal(t, tt.readingTime, stats.TotalReadingTime)
			assert.Equal(t, tt.pages, stats.PagesRead)
			assert.Equal(t, tt.words, stats.WordsRead)

			if tt.expectNonZero {
				assert.Greater(t, stats.ReadingSpeed, 0.0)
				assert.InDelta(t, tt.expectedWPM, stats.ReadingSpeed, 0.01)
			}
		})
	}
}

func TestReaderService_UpdateReadingStats_Accumulates(t *testing.T) {
	mockLogger := zap.NewNop()
	service := NewReaderService(nil, mockLogger, nil, nil, nil)

	stats := &ReadingStats{}

	// First update
	service.updateReadingStats(stats, 300, 5, 1500)
	assert.Equal(t, int64(300), stats.TotalReadingTime)
	assert.Equal(t, 5, stats.PagesRead)
	assert.Equal(t, 1500, stats.WordsRead)

	// Second update accumulates
	service.updateReadingStats(stats, 200, 3, 900)
	assert.Equal(t, int64(500), stats.TotalReadingTime)
	assert.Equal(t, 8, stats.PagesRead)
	assert.Equal(t, 2400, stats.WordsRead)
	assert.Equal(t, int64(500), stats.SessionTime) // Session time also accumulates
}

func TestReaderService_UpdateReadingStats_DailyProgress(t *testing.T) {
	service := NewReaderService(nil, zap.NewNop(), nil, nil, nil)

	tests := []struct {
		name              string
		initialProgress   int
		readingTime       int64
		expectedProgress  int
	}{
		{
			name:             "two minutes adds to zero",
			initialProgress:  0,
			readingTime:      120,
			expectedProgress: 2, // 120/60 = 2 minutes
		},
		{
			name:             "accumulates with existing progress",
			initialProgress:  10,
			readingTime:      300,
			expectedProgress: 15, // 10 + 300/60 = 10 + 5
		},
		{
			name:             "sub-minute reading does not add",
			initialProgress:  0,
			readingTime:      30,
			expectedProgress: 0, // 30/60 = 0 (integer division)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stats := &ReadingStats{DailyProgress: tt.initialProgress}
			service.updateReadingStats(stats, tt.readingTime, 1, 100)
			assert.Equal(t, tt.expectedProgress, stats.DailyProgress)
		})
	}
}

func TestReaderService_ReadingPosition_Struct(t *testing.T) {
	pos := ReadingPosition{
		BookID:          "book_1",
		ChapterID:       "ch_1",
		PageNumber:      42,
		WordOffset:      100,
		CharacterOffset: 500,
		PercentComplete: 0.35,
		Confidence:      0.99,
	}

	assert.Equal(t, "book_1", pos.BookID)
	assert.Equal(t, 42, pos.PageNumber)
	assert.Equal(t, 0.35, pos.PercentComplete)
}

func TestReaderService_ReadingPosition_FullStruct(t *testing.T) {
	pos := ReadingPosition{
		BookID:          "book-1",
		ChapterID:       "ch-3",
		PageNumber:      42,
		WordOffset:      1500,
		CharacterOffset: 8000,
		PercentComplete: 55.5,
		Location:        "loc-1234",
		CFI:             "epubcfi(/6/4!/4/2/1:0)",
		Timestamp:       time.Now(),
		Confidence:      0.95,
		PositionContext: PositionContext{
			SurroundingText: "some text around the cursor",
			ParagraphStart:  "The chapter begins",
			SentenceStart:   "This sentence",
			ChapterTitle:    "Chapter 3",
			SectionTitle:    "Section A",
		},
	}

	assert.Equal(t, "book-1", pos.BookID)
	assert.Equal(t, 42, pos.PageNumber)
	assert.Equal(t, 55.5, pos.PercentComplete)
	assert.Equal(t, 0.95, pos.Confidence)
	assert.Equal(t, "loc-1234", pos.Location)
	assert.Equal(t, "epubcfi(/6/4!/4/2/1:0)", pos.CFI)
	assert.Equal(t, "Chapter 3", pos.PositionContext.ChapterTitle)
	assert.Equal(t, "Section A", pos.PositionContext.SectionTitle)
}

func TestReaderService_ReadingSettings_Struct(t *testing.T) {
	settings := ReadingSettings{
		FontFamily:      "sans-serif",
		FontSize:        18,
		LineHeight:      1.8,
		TextAlign:       "left",
		Theme:           "dark",
		BackgroundColor: "#1a1a1a",
		TextColor:       "#eeeeee",
		ReadingMode:     "night",
	}

	assert.Equal(t, "sans-serif", settings.FontFamily)
	assert.Equal(t, 18, settings.FontSize)
	assert.Equal(t, "dark", settings.Theme)
	assert.Equal(t, "night", settings.ReadingMode)
}

func TestReaderService_Bookmark_Struct(t *testing.T) {
	bookmark := Bookmark{
		ID:       "bm_1",
		UserID:   1,
		BookID:   "book_1",
		Title:    "Important passage",
		Note:     "This is a note",
		Tags:     []string{"favorite", "chapter1"},
		Color:    "yellow",
		IsPublic: true,
		ShareURL: "https://example.com/share/bm_1",
	}

	assert.Equal(t, "bm_1", bookmark.ID)
	assert.Equal(t, "Important passage", bookmark.Title)
	assert.Equal(t, 2, len(bookmark.Tags))
	assert.True(t, bookmark.IsPublic)
}

func TestReaderService_Highlight_Struct(t *testing.T) {
	highlight := Highlight{
		ID:           "hl_1",
		UserID:       1,
		BookID:       "book_1",
		SelectedText: "This is highlighted text",
		Color:        "blue",
		Type:         "highlight",
	}

	assert.Equal(t, "hl_1", highlight.ID)
	assert.Equal(t, "This is highlighted text", highlight.SelectedText)
	assert.Equal(t, "highlight", highlight.Type)
}

func TestReaderService_ReadingStats_Struct(t *testing.T) {
	stats := ReadingStats{
		TotalReadingTime: 7200,
		SessionTime:      1800,
		PagesRead:        100,
		WordsRead:        25000,
		ReadingSpeed:     250.0,
		AverageSpeed:     230.0,
		DailyGoal:        30,
		DailyProgress:    20,
		ReadingStreak:    5,
		LongestStreak:    14,
		BooksCompleted:   3,
		PagesPerSession:  25.0,
		WeeklyStats: WeeklyReadingStats{
			Week:          "2026-W07",
			TotalTime:     10800,
			PagesRead:     150,
			SessionsCount: 7,
			DaysActive:    5,
		},
		MonthlyStats: MonthlyReadingStats{
			Month:          "2026-02",
			TotalTime:      43200,
			PagesRead:      600,
			BooksCompleted: 2,
			AverageDaily:   24.0,
		},
	}

	assert.Equal(t, int64(7200), stats.TotalReadingTime)
	assert.Equal(t, 5, stats.ReadingStreak)
	assert.Equal(t, "2026-W07", stats.WeeklyStats.Week)
	assert.Equal(t, "2026-02", stats.MonthlyStats.Month)
	assert.Equal(t, 2, stats.MonthlyStats.BooksCompleted)
}

func TestReaderService_SyncStatus_Struct(t *testing.T) {
	status := SyncStatus{
		LastSyncAt:     time.Now(),
		IsSynced:       true,
		ConflictExists: false,
		SyncVersion:    5,
	}

	assert.True(t, status.IsSynced)
	assert.False(t, status.ConflictExists)
	assert.Equal(t, int64(5), status.SyncVersion)
	assert.Empty(t, status.ConflictDetails)
}

func TestReaderService_SyncConflict_Struct(t *testing.T) {
	conflict := SyncConflict{
		DeviceID:   "device-1",
		DeviceName: "Phone",
		Position: ReadingPosition{
			PageNumber:      100,
			PercentComplete: 80.0,
		},
		Timestamp:    time.Now(),
		ConflictType: "position_mismatch",
	}

	assert.Equal(t, "device-1", conflict.DeviceID)
	assert.Equal(t, "position_mismatch", conflict.ConflictType)
	assert.Equal(t, 100, conflict.Position.PageNumber)
}

func TestMinInt(t *testing.T) {
	assert.Equal(t, 1, minInt(1, 2))
	assert.Equal(t, 1, minInt(2, 1))
	assert.Equal(t, 5, minInt(5, 5))
	assert.Equal(t, -1, minInt(-1, 0))
}

// ---------- Database-backed tests ----------

func TestReaderService_StartReading_DB(t *testing.T) {
	db := setupReaderTestDB(t)
	defer db.Close()

	service := NewReaderService(db, zap.NewNop(), nil, nil, nil)

	req := &StartReadingRequest{
		UserID: 1,
		BookID: "book-test-001",
		DeviceInfo: DeviceInfo{
			DeviceID:   "device-001",
			DeviceName: "Test Phone",
			DeviceType: "mobile",
			Platform:   "android",
			AppVersion: "1.0.0",
		},
		ResumeFromLastPosition: false,
	}

	ctx := context.Background()
	session, err := service.StartReading(ctx, req)

	require.NoError(t, err)
	require.NotNil(t, session)

	assert.True(t, strings.HasPrefix(session.ID, "session_"))
	assert.Equal(t, int64(1), session.UserID)
	assert.Equal(t, "book-test-001", session.BookID)
	assert.Equal(t, "device-001", session.DeviceID)
	assert.Equal(t, "Test Phone", session.DeviceName)
	assert.True(t, session.IsActive)
	assert.False(t, session.StartedAt.IsZero())
	assert.False(t, session.LastActiveAt.IsZero())

	// Should have default reading settings since none were provided
	assert.Equal(t, "serif", session.ReadingSettings.FontFamily)
	assert.Equal(t, 16, session.ReadingSettings.FontSize)

	// Sync status should be initialized
	assert.True(t, session.SyncStatus.IsSynced)
	assert.Equal(t, int64(1), session.SyncStatus.SyncVersion)
}

func TestReaderService_StartReading_WithCustomSettings(t *testing.T) {
	db := setupReaderTestDB(t)
	defer db.Close()

	service := NewReaderService(db, zap.NewNop(), nil, nil, nil)

	customSettings := &ReadingSettings{
		FontFamily:      "sans-serif",
		FontSize:        20,
		LineHeight:      1.8,
		TextAlign:       "left",
		Theme:           "dark",
		BackgroundColor: "#1a1a1a",
		TextColor:       "#eeeeee",
		ReadingMode:     "night",
		Brightness:      0.7,
	}

	req := &StartReadingRequest{
		UserID: 2,
		BookID: "book-custom",
		DeviceInfo: DeviceInfo{
			DeviceID:   "device-custom",
			DeviceName: "Tablet",
			DeviceType: "tablet",
		},
		ReadingSettings:        customSettings,
		ResumeFromLastPosition: false,
	}

	ctx := context.Background()
	session, err := service.StartReading(ctx, req)

	require.NoError(t, err)
	require.NotNil(t, session)

	assert.Equal(t, "sans-serif", session.ReadingSettings.FontFamily)
	assert.Equal(t, 20, session.ReadingSettings.FontSize)
	assert.Equal(t, 1.8, session.ReadingSettings.LineHeight)
	assert.Equal(t, "dark", session.ReadingSettings.Theme)
	assert.Equal(t, "night", session.ReadingSettings.ReadingMode)
	assert.Equal(t, 0.7, session.ReadingSettings.Brightness)
}

func TestReaderService_StartReading_ResumeNoSavedPosition(t *testing.T) {
	db := setupReaderTestDB(t)
	defer db.Close()

	service := NewReaderService(db, zap.NewNop(), nil, nil, nil)

	req := &StartReadingRequest{
		UserID: 3,
		BookID: "book-resume",
		DeviceInfo: DeviceInfo{
			DeviceID:   "device-resume",
			DeviceName: "E-Reader",
		},
		ResumeFromLastPosition: true,
	}

	ctx := context.Background()
	session, err := service.StartReading(ctx, req)

	require.NoError(t, err)
	require.NotNil(t, session)

	// No saved position means starting from zero
	assert.Equal(t, 0, session.CurrentPosition.PageNumber)
	assert.Equal(t, 0.0, session.CurrentPosition.PercentComplete)
}

func TestReaderService_CreateBookmark_DB(t *testing.T) {
	db := setupReaderTestDB(t)
	defer db.Close()

	service := NewReaderService(db, zap.NewNop(), nil, nil, nil)

	req := &CreateBookmarkRequest{
		UserID: 1,
		BookID: "book-bm-1",
		Position: ReadingPosition{
			PageNumber:      42,
			PercentComplete: 35.5,
		},
		Title:    "Important passage",
		Note:     "Remember this for later",
		Tags:     []string{"important", "review"},
		Color:    "#ff0000",
		IsPublic: true,
	}

	ctx := context.Background()
	bookmark, err := service.CreateBookmark(ctx, req)

	require.NoError(t, err)
	require.NotNil(t, bookmark)

	assert.True(t, strings.HasPrefix(bookmark.ID, "bookmark_"))
	assert.Equal(t, int64(1), bookmark.UserID)
	assert.Equal(t, "book-bm-1", bookmark.BookID)
	assert.Equal(t, "Important passage", bookmark.Title)
	assert.Equal(t, "Remember this for later", bookmark.Note)
	assert.Equal(t, []string{"important", "review"}, bookmark.Tags)
	assert.Equal(t, "#ff0000", bookmark.Color)
	assert.True(t, bookmark.IsPublic)
	assert.Contains(t, bookmark.ShareURL, "https://catalogizer.com/share/bookmark/")
	assert.False(t, bookmark.CreatedAt.IsZero())
	assert.False(t, bookmark.UpdatedAt.IsZero())
}

func TestReaderService_CreateBookmark_Private(t *testing.T) {
	db := setupReaderTestDB(t)
	defer db.Close()

	service := NewReaderService(db, zap.NewNop(), nil, nil, nil)

	req := &CreateBookmarkRequest{
		UserID: 1,
		BookID: "book-priv",
		Position: ReadingPosition{
			PageNumber: 10,
		},
		Title:    "Private note",
		IsPublic: false,
	}

	ctx := context.Background()
	bookmark, err := service.CreateBookmark(ctx, req)

	require.NoError(t, err)
	require.NotNil(t, bookmark)

	assert.False(t, bookmark.IsPublic)
	assert.Empty(t, bookmark.ShareURL, "private bookmark should not have a share URL")
}

func TestReaderService_CreateHighlight_DB(t *testing.T) {
	db := setupReaderTestDB(t)
	defer db.Close()

	service := NewReaderService(db, zap.NewNop(), nil, nil, nil)

	req := &CreateHighlightRequest{
		UserID: 1,
		BookID: "book-hl-1",
		StartPosition: ReadingPosition{
			PageNumber:      50,
			WordOffset:      100,
			PercentComplete: 40.0,
		},
		EndPosition: ReadingPosition{
			PageNumber:      50,
			WordOffset:      120,
			PercentComplete: 40.2,
		},
		SelectedText: "This is a highlighted passage that is long enough for the truncation logic to work properly.",
		Note:         "Interesting thought here",
		Color:        "#ffff00",
		Type:         "highlight",
		Tags:         []string{"philosophy"},
		IsPublic:     true,
	}

	ctx := context.Background()
	highlight, err := service.CreateHighlight(ctx, req)

	require.NoError(t, err)
	require.NotNil(t, highlight)

	assert.True(t, strings.HasPrefix(highlight.ID, "highlight_"))
	assert.Equal(t, int64(1), highlight.UserID)
	assert.Equal(t, "book-hl-1", highlight.BookID)
	assert.Equal(t, 50, highlight.StartPosition.PageNumber)
	assert.Equal(t, 50, highlight.EndPosition.PageNumber)
	assert.Contains(t, highlight.SelectedText, "highlighted passage")
	assert.Equal(t, "Interesting thought here", highlight.Note)
	assert.Equal(t, "#ffff00", highlight.Color)
	assert.Equal(t, "highlight", highlight.Type)
	assert.Equal(t, []string{"philosophy"}, highlight.Tags)
	assert.True(t, highlight.IsPublic)
	assert.Contains(t, highlight.ShareURL, "https://catalogizer.com/share/highlight/")
}

func TestReaderService_GetBookmarks_Empty(t *testing.T) {
	db := setupReaderTestDB(t)
	defer db.Close()

	service := NewReaderService(db, zap.NewNop(), nil, nil, nil)

	ctx := context.Background()
	bookmarks, err := service.GetBookmarks(ctx, 999, "nonexistent-book")

	require.NoError(t, err)
	assert.Empty(t, bookmarks)
}

func TestReaderService_GetHighlights_Empty(t *testing.T) {
	db := setupReaderTestDB(t)
	defer db.Close()

	service := NewReaderService(db, zap.NewNop(), nil, nil, nil)

	ctx := context.Background()
	highlights, err := service.GetHighlights(ctx, 999, "nonexistent-book")

	require.NoError(t, err)
	assert.Empty(t, highlights)
}

func TestReaderService_GetUserDailyGoal_Default(t *testing.T) {
	db := setupReaderTestDB(t)
	defer db.Close()

	service := NewReaderService(db, zap.NewNop(), nil, nil, nil)

	ctx := context.Background()
	goal := service.getUserDailyGoal(ctx, 999) // non-existent user

	assert.Equal(t, 30, goal, "default daily goal should be 30 minutes")
}

func TestReaderService_GetUserDailyGoal_Custom(t *testing.T) {
	db := setupReaderTestDB(t)
	defer db.Close()

	_, err := db.Exec("INSERT INTO user_reading_goals (user_id, daily_goal_minutes) VALUES (?, ?)", 1, 60)
	require.NoError(t, err)

	service := NewReaderService(db, zap.NewNop(), nil, nil, nil)

	ctx := context.Background()
	goal := service.getUserDailyGoal(ctx, 1)

	assert.Equal(t, 60, goal)
}

func TestReaderService_CalculateReadingStreak_NoHistory(t *testing.T) {
	db := setupReaderTestDB(t)
	defer db.Close()

	service := NewReaderService(db, zap.NewNop(), nil, nil, nil)

	ctx := context.Background()
	streak := service.calculateReadingStreak(ctx, 999)

	assert.Equal(t, 0, streak)
}

func TestReaderService_GetTodayReadingTime_NoSessions(t *testing.T) {
	db := setupReaderTestDB(t)
	defer db.Close()

	service := NewReaderService(db, zap.NewNop(), nil, nil, nil)

	ctx := context.Background()
	minutes := service.getTodayReadingTime(ctx, 999)

	assert.Equal(t, 0, minutes)
}

func TestReaderService_GetReadingStats_DB(t *testing.T) {
	db := setupReaderTestDB(t)
	defer db.Close()

	// Seed reading stats for user
	_, err := db.Exec(
		"INSERT INTO reading_stats (user_id, total_reading_time, pages_read, words_read, average_speed, books_completed) VALUES (?, ?, ?, ?, ?, ?)",
		1, 7200, 100, 25000, 230.0, 3,
	)
	require.NoError(t, err)

	service := NewReaderService(db, zap.NewNop(), nil, nil, nil)

	ctx := context.Background()
	stats, err := service.GetReadingStats(ctx, 1, "all")

	require.NoError(t, err)
	require.NotNil(t, stats)

	assert.Equal(t, int64(7200), stats.TotalReadingTime)
	assert.Equal(t, 100, stats.PagesRead)
	assert.Equal(t, 25000, stats.WordsRead)
	assert.Equal(t, 3, stats.BooksCompleted)
	assert.Equal(t, 30, stats.DailyGoal) // default since no custom goal set
}

func TestReaderService_GetReadingStats_NonExistentUser(t *testing.T) {
	db := setupReaderTestDB(t)
	defer db.Close()

	service := NewReaderService(db, zap.NewNop(), nil, nil, nil)

	ctx := context.Background()
	stats, err := service.GetReadingStats(ctx, 999, "all")

	assert.Error(t, err)
	assert.Nil(t, stats)
}

func TestReaderService_SyncAcrossDevices_SingleSession(t *testing.T) {
	db := setupReaderTestDB(t)
	defer db.Close()

	service := NewReaderService(db, zap.NewNop(), nil, nil, nil)

	// Start a single session
	req := &StartReadingRequest{
		UserID: 1,
		BookID: "book-sync-1",
		DeviceInfo: DeviceInfo{
			DeviceID:   "device-a",
			DeviceName: "Phone",
		},
	}

	ctx := context.Background()
	_, err := service.StartReading(ctx, req)
	require.NoError(t, err)

	// Sync should succeed with a single session (no-op)
	err = service.SyncAcrossDevices(ctx, 1, "book-sync-1")
	assert.NoError(t, err)
}
