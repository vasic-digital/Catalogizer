package tests

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"catalog-api/internal/models"
	"catalog-api/internal/services"
)

func TestReaderService_BookSession(t *testing.T) {
	ctx := context.Background()
	readerService := services.NewReaderService()

	t.Run("open and close book session", func(t *testing.T) {
		userID := "test-user-123"
		bookPath := "/books/Test Book.pdf"

		// Open book
		session, err := readerService.OpenBook(ctx, userID, bookPath)
		require.NoError(t, err)
		require.NotNil(t, session)

		assert.Equal(t, userID, session.UserID)
		assert.Equal(t, bookPath, session.BookPath)
		assert.NotEmpty(t, session.ID)
		assert.False(t, session.StartTime.IsZero())

		// Close book
		err = readerService.CloseBook(ctx, session.ID)
		require.NoError(t, err)

		// Try to get closed session
		closedSession, err := readerService.GetSession(ctx, session.ID)
		require.NoError(t, err)
		assert.False(t, closedSession.EndTime.IsZero())
	})

	t.Run("open book on specific device", func(t *testing.T) {
		userID := "test-user-456"
		bookPath := "/books/Device Test.epub"
		deviceID := "tablet-001"

		session, err := readerService.OpenBookOnDevice(ctx, userID, bookPath, deviceID)
		require.NoError(t, err)
		require.NotNil(t, session)

		assert.Equal(t, deviceID, session.DeviceID)
		assert.Equal(t, userID, session.UserID)
		assert.Equal(t, bookPath, session.BookPath)
	})

	t.Run("get user's active sessions", func(t *testing.T) {
		userID := "test-user-789"

		// Open multiple books
		session1, err := readerService.OpenBook(ctx, userID, "/books/Book1.pdf")
		require.NoError(t, err)

		session2, err := readerService.OpenBook(ctx, userID, "/books/Book2.epub")
		require.NoError(t, err)

		// Get all active sessions
		sessions, err := readerService.GetUserSessions(ctx, userID)
		require.NoError(t, err)
		assert.Len(t, sessions, 2)

		// Verify sessions are in the list
		sessionIDs := make(map[string]bool)
		for _, s := range sessions {
			sessionIDs[s.ID] = true
		}
		assert.True(t, sessionIDs[session1.ID])
		assert.True(t, sessionIDs[session2.ID])
	})
}

func TestReaderService_ReadingPosition(t *testing.T) {
	ctx := context.Background()
	readerService := services.NewReaderService()

	userID := "test-user-position"
	bookPath := "/books/Position Test.pdf"

	session, err := readerService.OpenBook(ctx, userID, bookPath)
	require.NoError(t, err)

	t.Run("update and retrieve basic position", func(t *testing.T) {
		position := &models.ReadingPosition{
			Page:      25,
			Word:      1250,
			Character: 18750,
			Timestamp: time.Now(),
		}

		err := readerService.UpdatePosition(ctx, session.ID, position)
		require.NoError(t, err)

		retrievedPosition, err := readerService.GetPosition(ctx, session.ID)
		require.NoError(t, err)
		require.NotNil(t, retrievedPosition)

		assert.Equal(t, position.Page, retrievedPosition.Page)
		assert.Equal(t, position.Word, retrievedPosition.Word)
		assert.Equal(t, position.Character, retrievedPosition.Character)
	})

	t.Run("update position with CFI (EPUB)", func(t *testing.T) {
		position := &models.ReadingPosition{
			Page:      10,
			Word:      500,
			Character: 7500,
			CFI:       "epubcfi(/6/4[chapter01]!/4/2/1:245)",
			Timestamp: time.Now(),
		}

		err := readerService.UpdatePosition(ctx, session.ID, position)
		require.NoError(t, err)

		retrievedPosition, err := readerService.GetPosition(ctx, session.ID)
		require.NoError(t, err)

		assert.Equal(t, position.CFI, retrievedPosition.CFI)
	})

	t.Run("position history", func(t *testing.T) {
		// Update position multiple times
		positions := []*models.ReadingPosition{
			{Page: 1, Word: 50, Character: 750, Timestamp: time.Now().Add(-3 * time.Hour)},
			{Page: 5, Word: 250, Character: 3750, Timestamp: time.Now().Add(-2 * time.Hour)},
			{Page: 12, Word: 600, Character: 9000, Timestamp: time.Now().Add(-1 * time.Hour)},
			{Page: 18, Word: 900, Character: 13500, Timestamp: time.Now()},
		}

		for _, pos := range positions {
			err := readerService.UpdatePosition(ctx, session.ID, pos)
			require.NoError(t, err)
		}

		// Get position history
		history, err := readerService.GetPositionHistory(ctx, session.ID, 10)
		require.NoError(t, err)
		assert.Len(t, history, 4)

		// Should be in chronological order (newest first)
		assert.Equal(t, 18, history[0].Page)
		assert.Equal(t, 1, history[3].Page)
	})
}

func TestReaderService_DeviceSynchronization(t *testing.T) {
	ctx := context.Background()
	readerService := services.NewReaderService()

	userID := "test-user-sync"
	bookPath := "/books/Sync Test.pdf"

	t.Run("sync position across devices", func(t *testing.T) {
		// Open book on phone
		phoneSession, err := readerService.OpenBookOnDevice(ctx, userID, bookPath, "phone-001")
		require.NoError(t, err)

		// Open same book on tablet
		tabletSession, err := readerService.OpenBookOnDevice(ctx, userID, bookPath, "tablet-001")
		require.NoError(t, err)

		// Update position on phone
		phonePosition := &models.ReadingPosition{
			Page:      30,
			Word:      1500,
			Character: 22500,
			Timestamp: time.Now(),
		}
		err = readerService.UpdatePosition(ctx, phoneSession.ID, phonePosition)
		require.NoError(t, err)

		// Sync to tablet
		err = readerService.SyncPosition(ctx, tabletSession.ID)
		require.NoError(t, err)

		// Check position on tablet
		tabletPosition, err := readerService.GetPosition(ctx, tabletSession.ID)
		require.NoError(t, err)

		assert.Equal(t, phonePosition.Page, tabletPosition.Page)
		assert.Equal(t, phonePosition.Word, tabletPosition.Word)
		assert.Equal(t, phonePosition.Character, tabletPosition.Character)
	})

	t.Run("handle sync conflicts", func(t *testing.T) {
		// Open book on two devices
		device1Session, err := readerService.OpenBookOnDevice(ctx, userID, bookPath, "device-001")
		require.NoError(t, err)

		device2Session, err := readerService.OpenBookOnDevice(ctx, userID, bookPath, "device-002")
		require.NoError(t, err)

		// Update position on both devices with different positions
		pos1 := &models.ReadingPosition{
			Page:      40,
			Word:      2000,
			Character: 30000,
			Timestamp: time.Now().Add(-1 * time.Hour),
		}
		err = readerService.UpdatePosition(ctx, device1Session.ID, pos1)
		require.NoError(t, err)

		pos2 := &models.ReadingPosition{
			Page:      45,
			Word:      2250,
			Character: 33750,
			Timestamp: time.Now(), // More recent
		}
		err = readerService.UpdatePosition(ctx, device2Session.ID, pos2)
		require.NoError(t, err)

		// Sync device1 (should get the more recent position from device2)
		err = readerService.SyncPosition(ctx, device1Session.ID)
		require.NoError(t, err)

		// Check that device1 now has the more recent position
		syncedPosition, err := readerService.GetPosition(ctx, device1Session.ID)
		require.NoError(t, err)

		assert.Equal(t, pos2.Page, syncedPosition.Page)
		assert.Equal(t, pos2.Word, syncedPosition.Word)
	})
}

func TestReaderService_ReadingSettings(t *testing.T) {
	ctx := context.Background()
	readerService := services.NewReaderService()

	userID := "test-user-settings"
	deviceID := "device-settings"

	t.Run("get default reading settings", func(t *testing.T) {
		settings, err := readerService.GetReadingSettings(ctx, userID, deviceID)
		require.NoError(t, err)
		require.NotNil(t, settings)

		// Check default values
		assert.Equal(t, "Sepia", settings.Theme)
		assert.Equal(t, 16, settings.FontSize)
		assert.Equal(t, "Georgia", settings.FontFamily)
		assert.Equal(t, 1.5, settings.LineSpacing)
		assert.Equal(t, 50, settings.Margin)
		assert.Equal(t, 50, settings.Brightness)
		assert.True(t, settings.AutoSync)
	})

	t.Run("update reading settings", func(t *testing.T) {
		newSettings := &models.ReadingSettings{
			Theme:           "Dark",
			FontSize:        20,
			FontFamily:      "Arial",
			LineSpacing:     2.0,
			Margin:          30,
			Brightness:      75,
			AutoSync:        false,
			PageAnimation:   "Slide",
			ReadingMode:     "Scroll",
			TapZones:        true,
			VolumeKeys:      true,
			FullScreen:      true,
			StatusBar:       false,
			Navigation:      true,
			Bookmarks:       true,
			Highlights:      true,
			Notes:           true,
			Dictionary:      true,
			Translation:     "Spanish",
			SpeechRate:      1.0,
			SpeechPitch:     1.0,
		}

		err := readerService.UpdateReadingSettings(ctx, userID, deviceID, newSettings)
		require.NoError(t, err)

		// Retrieve and verify
		settings, err := readerService.GetReadingSettings(ctx, userID, deviceID)
		require.NoError(t, err)

		assert.Equal(t, newSettings.Theme, settings.Theme)
		assert.Equal(t, newSettings.FontSize, settings.FontSize)
		assert.Equal(t, newSettings.FontFamily, settings.FontFamily)
		assert.Equal(t, newSettings.LineSpacing, settings.LineSpacing)
		assert.Equal(t, newSettings.AutoSync, settings.AutoSync)
		assert.Equal(t, newSettings.Translation, settings.Translation)
	})
}

func TestReaderService_BookmarksAndHighlights(t *testing.T) {
	ctx := context.Background()
	readerService := services.NewReaderService()

	userID := "test-user-bookmarks"
	bookPath := "/books/Bookmark Test.epub"

	session, err := readerService.OpenBook(ctx, userID, bookPath)
	require.NoError(t, err)

	t.Run("add and retrieve bookmarks", func(t *testing.T) {
		bookmark := &models.Bookmark{
			Page:        15,
			Position:    750,
			Title:       "Important Chapter",
			Note:        "Key concept explained here",
			Timestamp:   time.Now(),
		}

		bookmarkID, err := readerService.AddBookmark(ctx, session.ID, bookmark)
		require.NoError(t, err)
		assert.NotEmpty(t, bookmarkID)

		bookmarks, err := readerService.GetBookmarks(ctx, session.ID)
		require.NoError(t, err)
		assert.Len(t, bookmarks, 1)

		assert.Equal(t, bookmark.Page, bookmarks[0].Page)
		assert.Equal(t, bookmark.Title, bookmarks[0].Title)
		assert.Equal(t, bookmark.Note, bookmarks[0].Note)
	})

	t.Run("add and retrieve highlights", func(t *testing.T) {
		highlight := &models.Highlight{
			StartPage:     20,
			EndPage:       20,
			StartPosition: 1000,
			EndPosition:   1250,
			Text:          "This is the highlighted text that was selected",
			Color:         "Yellow",
			Note:          "Important quote",
			Timestamp:     time.Now(),
		}

		highlightID, err := readerService.AddHighlight(ctx, session.ID, highlight)
		require.NoError(t, err)
		assert.NotEmpty(t, highlightID)

		highlights, err := readerService.GetHighlights(ctx, session.ID)
		require.NoError(t, err)
		assert.Len(t, highlights, 1)

		assert.Equal(t, highlight.Text, highlights[0].Text)
		assert.Equal(t, highlight.Color, highlights[0].Color)
		assert.Equal(t, highlight.Note, highlights[0].Note)
	})

	t.Run("search bookmarks and highlights", func(t *testing.T) {
		// Add multiple bookmarks with different content
		bookmarks := []*models.Bookmark{
			{Page: 10, Title: "Introduction", Note: "Book overview"},
			{Page: 25, Title: "Chapter 2", Note: "Character development"},
			{Page: 40, Title: "Plot Twist", Note: "Unexpected turn of events"},
		}

		for _, b := range bookmarks {
			_, err := readerService.AddBookmark(ctx, session.ID, b)
			require.NoError(t, err)
		}

		// Search bookmarks
		results, err := readerService.SearchBookmarks(ctx, session.ID, "character")
		require.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, "Chapter 2", results[0].Title)
	})
}

func TestReaderService_ReadingAnalytics(t *testing.T) {
	ctx := context.Background()
	readerService := services.NewReaderService()

	userID := "test-user-analytics"
	bookPath := "/books/Analytics Test.pdf"

	session, err := readerService.OpenBook(ctx, userID, bookPath)
	require.NoError(t, err)

	t.Run("track reading time", func(t *testing.T) {
		// Simulate reading session
		startTime := time.Now().Add(-2 * time.Hour)
		endTime := time.Now()

		err := readerService.RecordReadingTime(ctx, session.ID, startTime, endTime)
		require.NoError(t, err)

		// Get reading analytics
		analytics, err := readerService.GetReadingAnalytics(ctx, userID, 7) // Last 7 days
		require.NoError(t, err)
		require.NotNil(t, analytics)

		assert.True(t, analytics.TotalReadingTime > 0)
		assert.True(t, analytics.AverageSessionTime > 0)
		assert.Equal(t, 1, analytics.SessionsCount)
	})

	t.Run("calculate reading speed", func(t *testing.T) {
		// Simulate reading progress
		positions := []*models.ReadingPosition{
			{Page: 1, Word: 0, Timestamp: time.Now().Add(-30 * time.Minute)},
			{Page: 5, Word: 1000, Timestamp: time.Now().Add(-20 * time.Minute)},
			{Page: 10, Word: 2000, Timestamp: time.Now().Add(-10 * time.Minute)},
			{Page: 15, Word: 3000, Timestamp: time.Now()},
		}

		for _, pos := range positions {
			err := readerService.UpdatePosition(ctx, session.ID, pos)
			require.NoError(t, err)
		}

		// Calculate reading speed
		speed, err := readerService.CalculateReadingSpeed(ctx, session.ID)
		require.NoError(t, err)

		assert.True(t, speed > 0, "Reading speed should be positive")
		assert.True(t, speed < 1000, "Reading speed should be reasonable (WPM)")
	})

	t.Run("track reading streaks", func(t *testing.T) {
		// Record reading sessions for consecutive days
		for i := 0; i < 5; i++ {
			sessionDate := time.Now().AddDate(0, 0, -i)
			err := readerService.RecordDailyReading(ctx, userID, sessionDate, 30*time.Minute)
			require.NoError(t, err)
		}

		// Get current streak
		streak, err := readerService.GetReadingStreak(ctx, userID)
		require.NoError(t, err)

		assert.Equal(t, 5, streak.CurrentStreak)
		assert.True(t, streak.LongestStreak >= 5)
	})
}

func TestReaderService_FileHandling(t *testing.T) {
	ctx := context.Background()
	readerService := services.NewReaderService()

	t.Run("extract text from PDF", func(t *testing.T) {
		// Create a temporary PDF file for testing
		tmpFile := createTempPDFFile(t)
		defer os.Remove(tmpFile)

		text, err := readerService.ExtractText(ctx, tmpFile, 1, 5) // Extract pages 1-5
		require.NoError(t, err)
		assert.NotEmpty(t, text)
	})

	t.Run("extract text from EPUB", func(t *testing.T) {
		// Create a temporary EPUB file for testing
		tmpFile := createTempEPUBFile(t)
		defer os.Remove(tmpFile)

		text, err := readerService.ExtractText(ctx, tmpFile, 0, 0) // Extract all
		require.NoError(t, err)
		assert.NotEmpty(t, text)
	})

	t.Run("get book metadata", func(t *testing.T) {
		tmpFile := createTempPDFFile(t)
		defer os.Remove(tmpFile)

		metadata, err := readerService.GetBookMetadata(ctx, tmpFile)
		require.NoError(t, err)
		require.NotNil(t, metadata)

		assert.NotEmpty(t, metadata.Title)
		assert.True(t, metadata.PageCount > 0)
	})
}

func TestReaderService_OfflineMode(t *testing.T) {
	ctx := context.Background()
	readerService := services.NewReaderService()

	userID := "test-user-offline"
	bookPath := "/books/Offline Test.epub"

	t.Run("download book for offline reading", func(t *testing.T) {
		downloadPath, err := readerService.DownloadForOffline(ctx, userID, bookPath)
		require.NoError(t, err)
		assert.NotEmpty(t, downloadPath)

		// Verify file exists
		_, err = os.Stat(downloadPath)
		assert.NoError(t, err)
	})

	t.Run("sync offline changes", func(t *testing.T) {
		session, err := readerService.OpenBook(ctx, userID, bookPath)
		require.NoError(t, err)

		// Simulate offline position updates
		offlinePositions := []*models.ReadingPosition{
			{Page: 10, Timestamp: time.Now().Add(-2 * time.Hour)},
			{Page: 15, Timestamp: time.Now().Add(-1 * time.Hour)},
			{Page: 20, Timestamp: time.Now()},
		}

		for _, pos := range offlinePositions {
			err := readerService.QueueOfflineUpdate(ctx, session.ID, pos)
			require.NoError(t, err)
		}

		// Sync offline changes
		err = readerService.SyncOfflineChanges(ctx, session.ID)
		require.NoError(t, err)

		// Verify final position
		finalPosition, err := readerService.GetPosition(ctx, session.ID)
		require.NoError(t, err)
		assert.Equal(t, 20, finalPosition.Page)
	})
}

// Helper functions
func createTempPDFFile(t *testing.T) string {
	tmpDir := t.TempDir()
	pdfFile := filepath.Join(tmpDir, "test.pdf")

	// Create a minimal PDF content
	pdfContent := `%PDF-1.4
1 0 obj
<<
/Type /Catalog
/Pages 2 0 R
>>
endobj

2 0 obj
<<
/Type /Pages
/Kids [3 0 R]
/Count 1
>>
endobj

3 0 obj
<<
/Type /Page
/Parent 2 0 R
/MediaBox [0 0 612 792]
/Contents 4 0 R
>>
endobj

4 0 obj
<<
/Length 44
>>
stream
BT
/F1 12 Tf
100 700 Td
(Test PDF Content) Tj
ET
endstream
endobj

xref
0 5
0000000000 65535 f
0000000009 00000 n
0000000058 00000 n
0000000115 00000 n
0000000189 00000 n
trailer
<<
/Size 5
/Root 1 0 R
>>
startxref
284
%%EOF`

	err := os.WriteFile(pdfFile, []byte(pdfContent), 0644)
	require.NoError(t, err)

	return pdfFile
}

func createTempEPUBFile(t *testing.T) string {
	tmpDir := t.TempDir()
	epubFile := filepath.Join(tmpDir, "test.epub")

	// Create a minimal EPUB structure (simplified)
	epubContent := []byte("PK\x03\x04") // ZIP file signature
	err := os.WriteFile(epubFile, epubContent, 0644)
	require.NoError(t, err)

	return epubFile
}

func BenchmarkReaderService(b *testing.B) {
	ctx := context.Background()
	readerService := services.NewReaderService()

	userID := "benchmark-user"
	bookPath := "/books/benchmark.pdf"

	session, err := readerService.OpenBook(ctx, userID, bookPath)
	if err != nil {
		b.Fatalf("Failed to open book: %v", err)
	}

	b.Run("position updates", func(b *testing.B) {
		position := &models.ReadingPosition{
			Page:      1,
			Word:      100,
			Character: 1500,
			Timestamp: time.Now(),
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			position.Page = i % 100 + 1
			position.Word = (i % 100 + 1) * 100
			position.Timestamp = time.Now()

			err := readerService.UpdatePosition(ctx, session.ID, position)
			if err != nil {
				b.Fatalf("Position update failed: %v", err)
			}
		}
	})

	b.Run("bookmark additions", func(b *testing.B) {
		bookmark := &models.Bookmark{
			Page:      1,
			Position:  100,
			Title:     "Benchmark Bookmark",
			Note:      "Performance testing",
			Timestamp: time.Now(),
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			bookmark.Page = i % 50 + 1
			bookmark.Position = (i % 50 + 1) * 100
			bookmark.Title = fmt.Sprintf("Bookmark %d", i)
			bookmark.Timestamp = time.Now()

			_, err := readerService.AddBookmark(ctx, session.ID, bookmark)
			if err != nil {
				b.Fatalf("Bookmark addition failed: %v", err)
			}
		}
	})
}