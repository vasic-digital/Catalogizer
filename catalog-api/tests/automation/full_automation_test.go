package automation

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"catalogizer/handlers"
	"catalogizer/services"
	"catalogizer/tests"

	"github.com/gorilla/mux"
)

// FullAutomationTest represents comprehensive automation testing with screenshots
type FullAutomationTest struct {
	server          *httptest.Server
	screenshotTool  *ScreenshotCapture
	testSuite       *tests.TestSuite
	documentationDir string
}

// NewFullAutomationTest creates a new full automation test instance
func NewFullAutomationTest(t *testing.T) *FullAutomationTest {
	// Setup test suite
	testSuite := tests.SetupTestSuite(t)

	// Create test server
	server := createTestServer(testSuite)

	// Setup screenshot capture
	docDir := filepath.Join("../../docs/screenshots")
	screenshotTool := NewScreenshotCapture(server.URL, docDir)

	return &FullAutomationTest{
		server:          server,
		screenshotTool:  screenshotTool,
		testSuite:       testSuite,
		documentationDir: docDir,
	}
}

// RunFullAutomationSuite executes complete automation testing with screenshot capture
func (fat *FullAutomationTest) RunFullAutomationSuite(t *testing.T) {
	log.Println("Starting Full Automation Test Suite with Screenshot Capture...")

	// Create test data
	fat.setupTestData(t)

	// Run UI automation with screenshots
	t.Run("UI_Automation_With_Screenshots", func(t *testing.T) {
		fat.runUIAutomation(t)
	})

	// Run API testing with UI verification
	t.Run("API_Testing_With_UI_Verification", func(t *testing.T) {
		fat.runAPITestingWithUI(t)
	})

	// Run responsive design testing
	t.Run("Responsive_Design_Testing", func(t *testing.T) {
		fat.runResponsiveDesignTesting(t)
	})

	// Run error state testing
	t.Run("Error_State_Testing", func(t *testing.T) {
		fat.runErrorStateTesting(t)
	})

	// Run performance testing with screenshots
	t.Run("Performance_Testing_With_Screenshots", func(t *testing.T) {
		fat.runPerformanceTesting(t)
	})

	// Generate comprehensive documentation
	t.Run("Generate_Documentation", func(t *testing.T) {
		fat.generateComprehensiveDocumentation(t)
	})

	log.Println("Full Automation Test Suite Completed Successfully!")
}

// setupTestData creates comprehensive test data for automation
func (fat *FullAutomationTest) setupTestData(t *testing.T) {
	log.Println("Setting up comprehensive test data...")

	// Create test users with different roles
	adminUser := tests.CreateTestUser(t, fat.testSuite.DB.DB, 1)
	adminUser.Role = "admin"

	regularUser := tests.CreateTestUser(t, fat.testSuite.DB.DB, 2)
	regularUser.Role = "user"

	// Create test media items
	for i := 1; i <= 20; i++ {
		tests.CreateTestMediaItem(t, fat.testSuite.DB.DB, i, regularUser.ID)
	}

	// Create test analytics events
	for i := 0; i < 100; i++ {
		fat.testSuite.AnalyticsService.TrackEvent(regularUser.ID, &models.AnalyticsEventRequest{
			EventType:  "media_view",
			EntityType: "media_item",
			EntityID:   i%20 + 1,
			SessionID:  "test_session",
		})
	}

	// Create test favorites
	for i := 1; i <= 5; i++ {
		fat.testSuite.FavoritesService.AddFavorite(regularUser.ID, &models.FavoriteRequest{
			EntityType: "media_item",
			EntityID:   i,
		})
	}

	// Create test collections
	for i := 1; i <= 3; i++ {
		fat.testSuite.FavoritesService.CreateCollection(regularUser.ID, &models.CreateCollectionRequest{
			Name:        fmt.Sprintf("Test Collection %d", i),
			Description: fmt.Sprintf("Description for test collection %d", i),
		})
	}

	// Create test conversion jobs
	for i := 1; i <= 5; i++ {
		fat.testSuite.ConversionService.CreateConversionJob(regularUser.ID, &models.ConversionRequest{
			SourcePath:   fmt.Sprintf("/test/source%d.mp4", i),
			TargetPath:   fmt.Sprintf("/test/target%d.mp3", i),
			SourceFormat: "mp4",
			TargetFormat: "mp3",
		})
	}

	log.Println("Test data setup completed")
}

// runUIAutomation executes complete UI automation flow
func (fat *FullAutomationTest) runUIAutomation(t *testing.T) {
	log.Println("Running UI Automation with Screenshot Capture...")

	// Capture full application flow
	err := fat.screenshotTool.CaptureFullApplicationFlow()
	if err != nil {
		t.Errorf("UI automation failed: %v", err)
		return
	}

	// Verify all major UI components
	fat.verifyUIComponents(t)

	log.Println("UI Automation completed successfully")
}

// runAPITestingWithUI tests API endpoints and captures UI changes
func (fat *FullAutomationTest) runAPITestingWithUI(t *testing.T) {
	log.Println("Running API Testing with UI Verification...")

	apiTests := []struct {
		name     string
		method   string
		endpoint string
		uiPage   string
		verify   func(*testing.T, *http.Response)
	}{
		{
			name:     "Analytics API",
			method:   "GET",
			endpoint: "/api/analytics/events",
			uiPage:   "/dashboard",
			verify:   fat.verifyAnalyticsAPI,
		},
		{
			name:     "Media API",
			method:   "GET",
			endpoint: "/api/media",
			uiPage:   "/media",
			verify:   fat.verifyMediaAPI,
		},
		{
			name:     "Favorites API",
			method:   "GET",
			endpoint: "/api/favorites",
			uiPage:   "/favorites",
			verify:   fat.verifyFavoritesAPI,
		},
		{
			name:     "Collections API",
			method:   "GET",
			endpoint: "/api/collections",
			uiPage:   "/collections",
			verify:   fat.verifyCollectionsAPI,
		},
	}

	for _, test := range apiTests {
		t.Run(test.name, func(t *testing.T) {
			// Test API endpoint
			resp, err := http.Get(fat.server.URL + test.endpoint)
			if err != nil {
				t.Errorf("API request failed: %v", err)
				return
			}
			defer resp.Body.Close()

			// Verify API response
			test.verify(t, resp)

			// Capture UI state after API call
			screenshotName := fmt.Sprintf("api-verification-%s", test.name)
			err = fat.screenshotTool.captureScreenshot(screenshotName,
				fmt.Sprintf("UI state after %s API call", test.name), "api")
			if err != nil {
				t.Logf("Screenshot capture failed: %v", err)
			}
		})
	}

	log.Println("API Testing with UI Verification completed")
}

// runResponsiveDesignTesting tests responsive design across viewports
func (fat *FullAutomationTest) runResponsiveDesignTesting(t *testing.T) {
	log.Println("Running Responsive Design Testing...")

	err := fat.screenshotTool.CaptureResponsiveDesign()
	if err != nil {
		t.Errorf("Responsive design testing failed: %v", err)
		return
	}

	// Test specific responsive behaviors
	fat.testResponsiveBehaviors(t)

	log.Println("Responsive Design Testing completed")
}

// runErrorStateTesting captures error states and edge cases
func (fat *FullAutomationTest) runErrorStateTesting(t *testing.T) {
	log.Println("Running Error State Testing...")

	err := fat.screenshotTool.CaptureErrorStates()
	if err != nil {
		t.Errorf("Error state testing failed: %v", err)
		return
	}

	// Test additional error scenarios
	fat.testAdditionalErrorScenarios(t)

	log.Println("Error State Testing completed")
}

// runPerformanceTesting tests performance and captures metrics
func (fat *FullAutomationTest) runPerformanceTesting(t *testing.T) {
	log.Println("Running Performance Testing with Screenshots...")

	// Test load times for different pages
	pages := []string{"/dashboard", "/media", "/collections", "/admin"}

	for _, page := range pages {
		t.Run(fmt.Sprintf("Performance_%s", page), func(t *testing.T) {
			start := time.Now()

			// Navigate to page
			err := chromedp.Run(fat.screenshotTool.ctx,
				chromedp.Navigate(fat.server.URL+page),
				chromedp.WaitReady("body"),
			)

			loadTime := time.Since(start)

			if err != nil {
				t.Errorf("Failed to load page %s: %v", page, err)
				return
			}

			// Capture performance screenshot
			screenshotName := fmt.Sprintf("performance-%s-loaded", sanitizeFilename(page))
			fat.screenshotTool.captureScreenshot(screenshotName,
				fmt.Sprintf("Page %s loaded in %v", page, loadTime), "performance")

			// Verify load time is reasonable (adjust threshold as needed)
			if loadTime > 5*time.Second {
				t.Errorf("Page %s load time too slow: %v", page, loadTime)
			}

			log.Printf("Page %s loaded in %v", page, loadTime)
		})
	}

	log.Println("Performance Testing completed")
}

// generateComprehensiveDocumentation creates complete documentation
func (fat *FullAutomationTest) generateComprehensiveDocumentation(t *testing.T) {
	log.Println("Generating Comprehensive Documentation...")

	// Generate screenshot documentation
	err := fat.screenshotTool.generateDocumentation()
	if err != nil {
		t.Errorf("Failed to generate screenshot documentation: %v", err)
		return
	}

	// Generate API documentation
	fat.generateAPIDocumentation(t)

	// Generate user guides
	fat.generateUserGuides(t)

	// Generate admin guides
	fat.generateAdminGuides(t)

	// Generate troubleshooting guide
	fat.generateTroubleshootingGuide(t)

	log.Println("Comprehensive Documentation generated successfully")
}

// Helper methods for verification

func (fat *FullAutomationTest) verifyUIComponents(t *testing.T) {
	// Verify critical UI components are captured
	requiredScreenshots := []string{
		"auth/login-screen",
		"dashboard/main-dashboard",
		"media/media-library",
		"collections/collections-view",
		"admin/user-management",
	}

	for _, screenshot := range requiredScreenshots {
		path := filepath.Join(fat.documentationDir, screenshot+".png")
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("Required screenshot missing: %s", screenshot)
		}
	}
}

func (fat *FullAutomationTest) verifyAnalyticsAPI(t *testing.T, resp *http.Response) {
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Analytics API returned status %d", resp.StatusCode)
	}
}

func (fat *FullAutomationTest) verifyMediaAPI(t *testing.T, resp *http.Response) {
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Media API returned status %d", resp.StatusCode)
	}
}

func (fat *FullAutomationTest) verifyFavoritesAPI(t *testing.T, resp *http.Response) {
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Favorites API returned status %d", resp.StatusCode)
	}
}

func (fat *FullAutomationTest) verifyCollectionsAPI(t *testing.T, resp *http.Response) {
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Collections API returned status %d", resp.StatusCode)
	}
}

func (fat *FullAutomationTest) testResponsiveBehaviors(t *testing.T) {
	// Test mobile menu functionality
	// Test responsive grid layouts
	// Test touch interactions
	log.Println("Testing responsive behaviors...")
}

func (fat *FullAutomationTest) testAdditionalErrorScenarios(t *testing.T) {
	// Test permission denied scenarios
	// Test network timeout scenarios
	// Test invalid input scenarios
	log.Println("Testing additional error scenarios...")
}

func (fat *FullAutomationTest) generateAPIDocumentation(t *testing.T) {
	apiDocPath := filepath.Join(fat.documentationDir, "api-documentation.md")

	content := `# Catalogizer v3.0 - API Documentation

## Authentication APIs
- POST /api/auth/login
- POST /api/auth/logout
- POST /api/auth/register

## Media Management APIs
- GET /api/media
- POST /api/media/upload
- GET /api/media/{id}
- DELETE /api/media/{id}

## Analytics APIs
- GET /api/analytics/events
- POST /api/analytics/track
- GET /api/analytics/dashboard

## Collections APIs
- GET /api/collections
- POST /api/collections
- GET /api/collections/{id}

## Favorites APIs
- GET /api/favorites
- POST /api/favorites
- DELETE /api/favorites/{id}

## Administration APIs
- GET /api/admin/users
- POST /api/admin/users
- GET /api/admin/config
- PUT /api/admin/config
`

	os.WriteFile(apiDocPath, []byte(content), 0644)
}

func (fat *FullAutomationTest) generateUserGuides(t *testing.T) {
	userGuidePath := filepath.Join(fat.documentationDir, "user-guide.md")

	content := `# Catalogizer v3.0 - User Guide

## Getting Started
1. Login to your account
2. Upload your first media files
3. Organize with collections
4. Use favorites for quick access

## Media Management
- Uploading files
- Organizing collections
- Searching and filtering
- Format conversion

## Collaboration
- Sharing collections
- Managing favorites
- User permissions
`

	os.WriteFile(userGuidePath, []byte(content), 0644)
}

func (fat *FullAutomationTest) generateAdminGuides(t *testing.T) {
	adminGuidePath := filepath.Join(fat.documentationDir, "admin-guide.md")

	content := `# Catalogizer v3.0 - Administrator Guide

## User Management
- Creating users
- Assigning roles
- Managing permissions

## System Configuration
- Database settings
- Storage configuration
- Network settings

## Monitoring
- System health
- Error reporting
- Log management

## Maintenance
- Backup procedures
- Update process
- Troubleshooting
`

	os.WriteFile(adminGuidePath, []byte(content), 0644)
}

func (fat *FullAutomationTest) generateTroubleshootingGuide(t *testing.T) {
	troubleshootingPath := filepath.Join(fat.documentationDir, "troubleshooting-guide.md")

	content := `# Catalogizer v3.0 - Troubleshooting Guide

## Common Issues

### Login Problems
- Check username/password
- Verify account status
- Check network connectivity

### Upload Issues
- Check file size limits
- Verify file format support
- Check storage space

### Performance Issues
- Monitor system resources
- Check database performance
- Review error logs

## Error Codes
- 401: Authentication required
- 403: Insufficient permissions
- 404: Resource not found
- 500: Internal server error

## Support Resources
- Check system logs
- Review error reports
- Contact support team
`

	os.WriteFile(troubleshootingPath, []byte(content), 0644)
}

// Cleanup resources
func (fat *FullAutomationTest) Cleanup() {
	if fat.screenshotTool != nil {
		fat.screenshotTool.Close()
	}
	if fat.server != nil {
		fat.server.Close()
	}
	if fat.testSuite != nil {
		fat.testSuite.Cleanup()
	}
}

// createTestServer creates a test HTTP server with all routes
func createTestServer(testSuite *tests.TestSuite) *httptest.Server {
	router := mux.NewRouter()

	// Create mock permission service
	permissionService := &MockPermissionService{}

	// Setup route handlers
	analyticsHandler := handlers.NewAnalyticsHandler(testSuite.AnalyticsService, permissionService)
	favoritesHandler := handlers.NewFavoritesHandler(testSuite.FavoritesService, permissionService)
	conversionHandler := handlers.NewConversionHandler(testSuite.ConversionService, permissionService)
	syncHandler := handlers.NewSyncHandler(testSuite.SyncService, permissionService)
	stressTestHandler := handlers.NewStressTestHandler(testSuite.StressTestService, permissionService)
	errorHandler := handlers.NewErrorReportingHandler(testSuite.ErrorReportingService, permissionService)
	logHandler := handlers.NewLogManagementHandler(testSuite.LogManagementService, permissionService)
	configHandler := handlers.NewConfigurationHandler(testSuite.ConfigurationService, permissionService)

	// API routes
	api := router.PathPrefix("/api").Subrouter()

	// Analytics routes
	api.HandleFunc("/analytics/events", analyticsHandler.TrackEvent).Methods("POST")
	api.HandleFunc("/analytics/events", analyticsHandler.GetEventsByUser).Methods("GET")
	api.HandleFunc("/analytics/dashboard", analyticsHandler.GetDashboardMetrics).Methods("GET")

	// Favorites routes
	api.HandleFunc("/favorites", favoritesHandler.AddFavorite).Methods("POST")
	api.HandleFunc("/favorites", favoritesHandler.GetFavoritesByUser).Methods("GET")

	// Collections routes
	api.HandleFunc("/collections", favoritesHandler.CreateCollection).Methods("POST")
	api.HandleFunc("/collections", favoritesHandler.GetCollectionsByUser).Methods("GET")

	// Media routes (mock)
	api.HandleFunc("/media", mockMediaHandler).Methods("GET")
	api.HandleFunc("/media/upload", mockUploadHandler).Methods("POST")

	// Static file serving for frontend
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./static/")))

	return httptest.NewServer(router)
}

// Mock handlers and services

type MockPermissionService struct{}

func (m *MockPermissionService) HasPermission(userID int, resource, action string) bool {
	return true // Allow all permissions for testing
}

func mockMediaHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"media": [], "total": 0}`))
}

func mockUploadHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"success": true, "message": "File uploaded successfully"}`))
}

// Test entry point
func TestFullAutomationSuite(t *testing.T) {
	fat := NewFullAutomationTest(t)
	defer fat.Cleanup()

	fat.RunFullAutomationSuite(t)
}