package automation

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/chromedp/chromedp"
)

// ScreenshotCapture handles automated screenshot capture for documentation
type ScreenshotCapture struct {
	ctx         context.Context
	cancel      context.CancelFunc
	baseURL     string
	outputDir   string
	screenshots []Screenshot
}

// Screenshot represents a captured screenshot with metadata
type Screenshot struct {
	Name         string    `json:"name"`
	Path         string    `json:"path"`
	URL          string    `json:"url"`
	Description  string    `json:"description"`
	Section      string    `json:"section"`
	Timestamp    time.Time `json:"timestamp"`
	ViewportSize Viewport  `json:"viewport_size"`
}

// Viewport represents browser viewport dimensions
type Viewport struct {
	Width  int64 `json:"width"`
	Height int64 `json:"height"`
}

// TestScenario represents a complete UI testing scenario
type TestScenario struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Steps       []TestStep `json:"steps"`
	Section     string     `json:"section"`
}

// TestStep represents a single step in a test scenario
type TestStep struct {
	Name             string            `json:"name"`
	Action           string            `json:"action"`
	Selector         string            `json:"selector,omitempty"`
	Value            string            `json:"value,omitempty"`
	WaitFor          string            `json:"wait_for,omitempty"`
	Screenshot       bool              `json:"screenshot"`
	Annotations      []Annotation      `json:"annotations,omitempty"`
	ExpectedElements []string          `json:"expected_elements,omitempty"`
	Metadata         map[string]string `json:"metadata,omitempty"`
}

// Annotation represents UI annotations for screenshots
type Annotation struct {
	Type        string  `json:"type"` // arrow, circle, rectangle, text
	X           float64 `json:"x"`
	Y           float64 `json:"y"`
	Width       float64 `json:"width,omitempty"`
	Height      float64 `json:"height,omitempty"`
	Text        string  `json:"text,omitempty"`
	Color       string  `json:"color"`
	Description string  `json:"description"`
}

// NewScreenshotCapture creates a new screenshot capture instance
func NewScreenshotCapture(baseURL, outputDir string) *ScreenshotCapture {
	// Create Chrome context
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true), // Run in headless mode for automated testing
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-web-security", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.WindowSize(1920, 1080),
	)

	allocCtx, _ := chromedp.NewExecAllocator(context.Background(), opts...)
	ctx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))

	// Create output directory
	os.MkdirAll(outputDir, 0755)

	return &ScreenshotCapture{
		ctx:         ctx,
		cancel:      cancel,
		baseURL:     baseURL,
		outputDir:   outputDir,
		screenshots: make([]Screenshot, 0),
	}
}

// Close cleans up the screenshot capture resources
func (sc *ScreenshotCapture) Close() {
	sc.cancel()
}

// CaptureFullApplicationFlow captures screenshots for the entire application
func (sc *ScreenshotCapture) CaptureFullApplicationFlow() error {
	scenarios := sc.getAllTestScenarios()

	for _, scenario := range scenarios {
		log.Printf("Executing scenario: %s", scenario.Name)
		if err := sc.executeScenario(scenario); err != nil {
			log.Printf("Error in scenario %s: %v", scenario.Name, err)
			continue
		}
	}

	// Generate documentation with screenshots
	return sc.generateDocumentation()
}

// executeScenario runs a complete test scenario and captures screenshots
func (sc *ScreenshotCapture) executeScenario(scenario TestScenario) error {
	for i, step := range scenario.Steps {
		log.Printf("Executing step %d: %s", i+1, step.Name)

		if err := sc.executeStep(step, scenario.Section); err != nil {
			return fmt.Errorf("step %d failed: %w", i+1, err)
		}

		// Capture screenshot if requested
		if step.Screenshot {
			screenshotName := fmt.Sprintf("%s-%d-%s",
				scenario.Name, i+1, sanitizeFilename(step.Name))

			if err := sc.captureScreenshot(screenshotName, step.Name, scenario.Section); err != nil {
				log.Printf("Screenshot capture failed: %v", err)
			}
		}

		// Small delay between steps
		time.Sleep(500 * time.Millisecond)
	}

	return nil
}

// executeStep performs a single test step
func (sc *ScreenshotCapture) executeStep(step TestStep, section string) error {
	var tasks chromedp.Tasks

	switch step.Action {
	case "navigate":
		tasks = append(tasks, chromedp.Navigate(sc.baseURL+step.Value))

	case "click":
		tasks = append(tasks, chromedp.Click(step.Selector))

	case "type":
		tasks = append(tasks, chromedp.SendKeys(step.Selector, step.Value))

	case "wait":
		if step.WaitFor != "" {
			tasks = append(tasks, chromedp.WaitVisible(step.WaitFor))
		} else {
			tasks = append(tasks, chromedp.Sleep(time.Duration(2)*time.Second))
		}

	case "scroll":
		tasks = append(tasks, chromedp.ScrollIntoView(step.Selector))

	case "hover":
		// Hover action - mouse over element
		// chromedp doesn't have a direct MouseOver, using scroll into view instead
		tasks = append(tasks, chromedp.ScrollIntoView(step.Selector))

	case "select":
		tasks = append(tasks, chromedp.SetValue(step.Selector, step.Value))

	case "verify":
		tasks = append(tasks, chromedp.WaitVisible(step.Selector))

	default:
		return fmt.Errorf("unknown action: %s", step.Action)
	}

	// Add wait for element if specified
	if step.WaitFor != "" && step.Action != "wait" {
		tasks = append(tasks, chromedp.WaitVisible(step.WaitFor))
	}

	return chromedp.Run(sc.ctx, tasks)
}

// captureScreenshot takes a screenshot and saves it with metadata
func (sc *ScreenshotCapture) captureScreenshot(name, description, section string) error {
	var buf []byte

	// Get current URL
	var currentURL string

	err := chromedp.Run(sc.ctx,
		chromedp.Location(&currentURL),
		chromedp.FullScreenshot(&buf, 90),
	)

	if err != nil {
		return fmt.Errorf("failed to capture screenshot: %w", err)
	}

	// Create section directory
	sectionDir := filepath.Join(sc.outputDir, sanitizeFilename(section))
	os.MkdirAll(sectionDir, 0755)

	// Save screenshot
	filename := fmt.Sprintf("%s.png", sanitizeFilename(name))
	filepath := filepath.Join(sectionDir, filename)

	if err := os.WriteFile(filepath, buf, 0644); err != nil {
		return fmt.Errorf("failed to save screenshot: %w", err)
	}

	// Save metadata
	screenshot := Screenshot{
		Name:         name,
		Path:         filepath,
		URL:          currentURL,
		Description:  description,
		Section:      section,
		Timestamp:    time.Now(),
		ViewportSize: Viewport{Width: 1920, Height: 1080},
	}

	sc.screenshots = append(sc.screenshots, screenshot)

	log.Printf("Screenshot captured: %s", filepath)
	return nil
}

// getAllTestScenarios returns all test scenarios for the application
func (sc *ScreenshotCapture) getAllTestScenarios() []TestScenario {
	return []TestScenario{
		// Authentication Flow
		{
			Name:        "authentication-flow",
			Description: "Complete authentication and user onboarding flow",
			Section:     "auth",
			Steps: []TestStep{
				{
					Name:       "navigate-to-login",
					Action:     "navigate",
					Value:      "/login",
					Screenshot: true,
				},
				{
					Name:       "show-login-form",
					Action:     "wait",
					WaitFor:    "form[data-testid='login-form']",
					Screenshot: true,
				},
				{
					Name:     "enter-username",
					Action:   "type",
					Selector: "input[name='username']",
					Value:    "demo@catalogizer.com",
				},
				{
					Name:     "enter-password",
					Action:   "type",
					Selector: "input[name='password']",
					Value:    "demo123",
				},
				{
					Name:       "click-login-button",
					Action:     "click",
					Selector:   "button[type='submit']",
					Screenshot: true,
				},
				{
					Name:       "dashboard-loaded",
					Action:     "wait",
					WaitFor:    "[data-testid='main-dashboard']",
					Screenshot: true,
				},
			},
		},

		// Dashboard Overview
		{
			Name:        "dashboard-overview",
			Description: "Main dashboard with analytics and quick actions",
			Section:     "dashboard",
			Steps: []TestStep{
				{
					Name:       "main-dashboard",
					Action:     "wait",
					WaitFor:    "[data-testid='main-dashboard']",
					Screenshot: true,
				},
				{
					Name:       "analytics-panel",
					Action:     "click",
					Selector:   "[data-testid='analytics-tab']",
					Screenshot: true,
				},
				{
					Name:       "realtime-metrics",
					Action:     "click",
					Selector:   "[data-testid='realtime-metrics']",
					Screenshot: true,
				},
				{
					Name:       "reports-view",
					Action:     "click",
					Selector:   "[data-testid='reports-view']",
					Screenshot: true,
				},
			},
		},

		// Media Management
		{
			Name:        "media-management",
			Description: "Media library, upload, and management features",
			Section:     "media",
			Steps: []TestStep{
				{
					Name:       "navigate-to-media",
					Action:     "navigate",
					Value:      "/media",
					Screenshot: true,
				},
				{
					Name:       "media-library-grid",
					Action:     "wait",
					WaitFor:    "[data-testid='media-grid']",
					Screenshot: true,
				},
				{
					Name:       "switch-to-list-view",
					Action:     "click",
					Selector:   "[data-testid='list-view-toggle']",
					Screenshot: true,
				},
				{
					Name:       "open-upload-modal",
					Action:     "click",
					Selector:   "[data-testid='upload-button']",
					Screenshot: true,
				},
				{
					Name:       "upload-interface",
					Action:     "wait",
					WaitFor:    "[data-testid='upload-modal']",
					Screenshot: true,
				},
				{
					Name:     "close-upload-modal",
					Action:   "click",
					Selector: "[data-testid='close-modal']",
				},
				{
					Name:       "open-media-details",
					Action:     "click",
					Selector:   "[data-testid='media-item']:first-child",
					Screenshot: true,
				},
			},
		},

		// Collections and Favorites
		{
			Name:        "collections-favorites",
			Description: "Collections management and favorites system",
			Section:     "collections",
			Steps: []TestStep{
				{
					Name:       "navigate-to-collections",
					Action:     "navigate",
					Value:      "/collections",
					Screenshot: true,
				},
				{
					Name:       "collections-overview",
					Action:     "wait",
					WaitFor:    "[data-testid='collections-grid']",
					Screenshot: true,
				},
				{
					Name:       "create-collection-modal",
					Action:     "click",
					Selector:   "[data-testid='create-collection']",
					Screenshot: true,
				},
				{
					Name:     "close-create-modal",
					Action:   "click",
					Selector: "[data-testid='close-modal']",
				},
				{
					Name:       "navigate-to-favorites",
					Action:     "navigate",
					Value:      "/favorites",
					Screenshot: true,
				},
				{
					Name:       "favorites-management",
					Action:     "wait",
					WaitFor:    "[data-testid='favorites-view']",
					Screenshot: true,
				},
			},
		},

		// Advanced Features
		{
			Name:        "advanced-features",
			Description: "Format conversion, sync, and advanced tools",
			Section:     "features",
			Steps: []TestStep{
				{
					Name:       "navigate-to-conversion",
					Action:     "navigate",
					Value:      "/tools/conversion",
					Screenshot: true,
				},
				{
					Name:       "format-conversion-interface",
					Action:     "wait",
					WaitFor:    "[data-testid='conversion-queue']",
					Screenshot: true,
				},
				{
					Name:       "navigate-to-sync",
					Action:     "navigate",
					Value:      "/tools/sync",
					Screenshot: true,
				},
				{
					Name:       "sync-backup-settings",
					Action:     "wait",
					WaitFor:    "[data-testid='sync-settings']",
					Screenshot: true,
				},
				{
					Name:       "navigate-to-errors",
					Action:     "navigate",
					Value:      "/admin/errors",
					Screenshot: true,
				},
				{
					Name:       "error-reporting-dashboard",
					Action:     "wait",
					WaitFor:    "[data-testid='error-dashboard']",
					Screenshot: true,
				},
				{
					Name:       "navigate-to-logs",
					Action:     "navigate",
					Value:      "/admin/logs",
					Screenshot: true,
				},
				{
					Name:       "log-management-interface",
					Action:     "wait",
					WaitFor:    "[data-testid='log-viewer']",
					Screenshot: true,
				},
			},
		},

		// Administration
		{
			Name:        "administration",
			Description: "User management and system administration",
			Section:     "admin",
			Steps: []TestStep{
				{
					Name:       "navigate-to-users",
					Action:     "navigate",
					Value:      "/admin/users",
					Screenshot: true,
				},
				{
					Name:       "user-management-table",
					Action:     "wait",
					WaitFor:    "[data-testid='users-table']",
					Screenshot: true,
				},
				{
					Name:       "navigate-to-config",
					Action:     "navigate",
					Value:      "/admin/config",
					Screenshot: true,
				},
				{
					Name:       "system-configuration",
					Action:     "wait",
					WaitFor:    "[data-testid='config-panel']",
					Screenshot: true,
				},
				{
					Name:       "navigate-to-health",
					Action:     "navigate",
					Value:      "/admin/health",
					Screenshot: true,
				},
				{
					Name:       "system-health-monitor",
					Action:     "wait",
					WaitFor:    "[data-testid='health-dashboard']",
					Screenshot: true,
				},
			},
		},

		// Installation Wizard
		{
			Name:        "installation-wizard",
			Description: "Complete setup wizard walkthrough",
			Section:     "wizard",
			Steps: []TestStep{
				{
					Name:       "navigate-to-wizard",
					Action:     "navigate",
					Value:      "/setup/wizard",
					Screenshot: true,
				},
				{
					Name:       "welcome-step",
					Action:     "wait",
					WaitFor:    "[data-testid='wizard-welcome']",
					Screenshot: true,
				},
				{
					Name:     "next-to-database",
					Action:   "click",
					Selector: "[data-testid='wizard-next']",
				},
				{
					Name:       "database-configuration",
					Action:     "wait",
					WaitFor:    "[data-testid='wizard-database']",
					Screenshot: true,
				},
				{
					Name:     "next-to-storage",
					Action:   "click",
					Selector: "[data-testid='wizard-next']",
				},
				{
					Name:       "storage-configuration",
					Action:     "wait",
					WaitFor:    "[data-testid='wizard-storage']",
					Screenshot: true,
				},
				{
					Name:     "next-to-network",
					Action:   "click",
					Selector: "[data-testid='wizard-next']",
				},
				{
					Name:       "network-configuration",
					Action:     "wait",
					WaitFor:    "[data-testid='wizard-network']",
					Screenshot: true,
				},
			},
		},

		// Mobile Responsive Views
		{
			Name:        "mobile-responsive",
			Description: "Mobile and responsive design testing",
			Section:     "mobile",
			Steps: []TestStep{
				{
					Name:   "set-mobile-viewport",
					Action: "viewport",
					Value:  "375x667",
				},
				{
					Name:       "mobile-dashboard",
					Action:     "navigate",
					Value:      "/dashboard",
					Screenshot: true,
				},
				{
					Name:       "mobile-media-view",
					Action:     "navigate",
					Value:      "/media",
					Screenshot: true,
				},
				{
					Name:       "mobile-navigation",
					Action:     "click",
					Selector:   "[data-testid='mobile-menu-toggle']",
					Screenshot: true,
				},
			},
		},
	}
}

// generateDocumentation creates markdown documentation with screenshots
func (sc *ScreenshotCapture) generateDocumentation() error {
	docPath := filepath.Join(sc.outputDir, "generated-documentation.md")

	content := sc.buildDocumentationContent()

	return os.WriteFile(docPath, []byte(content), 0644)
}

// buildDocumentationContent creates the complete documentation content
func (sc *ScreenshotCapture) buildDocumentationContent() string {
	content := `# Catalogizer v3.0 - Automated Screenshot Documentation

Generated on: ` + time.Now().Format("2006-01-02 15:04:05") + `

This documentation contains automatically captured screenshots of all application interfaces.

`

	// Group screenshots by section
	sections := make(map[string][]Screenshot)
	for _, screenshot := range sc.screenshots {
		sections[screenshot.Section] = append(sections[screenshot.Section], screenshot)
	}

	// Generate content for each section
	for section, screenshots := range sections {
		content += fmt.Sprintf("## %s\n\n", formatSectionTitle(section))

		for _, screenshot := range screenshots {
			relativePath := filepath.Join("screenshots", screenshot.Section, filepath.Base(screenshot.Path))
			content += fmt.Sprintf("### %s\n\n", screenshot.Name)
			content += fmt.Sprintf("![%s](%s)\n", screenshot.Description, relativePath)
			content += fmt.Sprintf("*%s*\n\n", screenshot.Description)
			content += fmt.Sprintf("**URL**: %s  \n", screenshot.URL)
			content += fmt.Sprintf("**Timestamp**: %s  \n", screenshot.Timestamp.Format("2006-01-02 15:04:05"))
			content += fmt.Sprintf("**Viewport**: %dx%d\n\n", screenshot.ViewportSize.Width, screenshot.ViewportSize.Height)
			content += "---\n\n"
		}
	}

	return content
}

// Helper functions

func sanitizeFilename(name string) string {
	// Replace invalid filename characters
	replacements := map[rune]string{
		' ':  "-",
		'/':  "-",
		'\\': "-",
		':':  "",
		'*':  "",
		'?':  "",
		'"':  "",
		'<':  "",
		'>':  "",
		'|':  "",
	}

	result := ""
	for _, r := range name {
		if replacement, exists := replacements[r]; exists {
			result += replacement
		} else {
			result += string(r)
		}
	}

	return result
}

func formatSectionTitle(section string) string {
	titles := map[string]string{
		"auth":        "Authentication & Login",
		"dashboard":   "Dashboard & Analytics",
		"media":       "Media Management",
		"collections": "Collections & Favorites",
		"features":    "Advanced Features",
		"admin":       "Administration",
		"wizard":      "Installation Wizard",
		"mobile":      "Mobile Interface",
	}

	if title, exists := titles[section]; exists {
		return title
	}

	return section
}

// SetViewport changes browser viewport size
func (sc *ScreenshotCapture) SetViewport(width, height int64) error {
	return chromedp.Run(sc.ctx,
		chromedp.EmulateViewport(width, height),
	)
}

// CaptureErrorStates captures screenshots of error states and edge cases
func (sc *ScreenshotCapture) CaptureErrorStates() error {
	errorScenarios := []TestScenario{
		{
			Name:        "error-states",
			Description: "Error states and edge cases",
			Section:     "errors",
			Steps: []TestStep{
				{
					Name:       "404-not-found",
					Action:     "navigate",
					Value:      "/nonexistent-page",
					Screenshot: true,
				},
				{
					Name:       "network-error",
					Action:     "navigate",
					Value:      "/api/invalid-endpoint",
					Screenshot: true,
				},
				{
					Name:       "empty-media-library",
					Action:     "navigate",
					Value:      "/media?filter=empty",
					Screenshot: true,
				},
				{
					Name:       "loading-states",
					Action:     "navigate",
					Value:      "/media?loading=true",
					Screenshot: true,
				},
			},
		},
	}

	for _, scenario := range errorScenarios {
		if err := sc.executeScenario(scenario); err != nil {
			log.Printf("Error scenario failed: %v", err)
		}
	}

	return nil
}

// CaptureResponsiveDesign captures screenshots across different viewport sizes
func (sc *ScreenshotCapture) CaptureResponsiveDesign() error {
	viewports := []Viewport{
		{Width: 1920, Height: 1080}, // Desktop
		{Width: 1366, Height: 768},  // Laptop
		{Width: 768, Height: 1024},  // Tablet
		{Width: 375, Height: 667},   // Mobile
	}

	pages := []string{"/dashboard", "/media", "/collections"}

	for _, viewport := range viewports {
		sc.SetViewport(viewport.Width, viewport.Height)

		for _, page := range pages {
			err := chromedp.Run(sc.ctx, chromedp.Navigate(sc.baseURL+page))
			if err != nil {
				continue
			}

			screenshotName := fmt.Sprintf("responsive-%dx%d-%s",
				viewport.Width, viewport.Height, sanitizeFilename(page))

			sc.captureScreenshot(screenshotName,
				fmt.Sprintf("Page %s at %dx%d", page, viewport.Width, viewport.Height),
				"responsive")
		}
	}

	return nil
}
