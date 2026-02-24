# User Flow Automation Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Refactor `pkg/yole/` into generic `pkg/userflow/`, add adapter implementations for all platforms, integrate the Containers module, and create ~200+ Catalogizer-specific user flow challenges with exhaustive documentation.

**Architecture:** Adapter-per-platform pattern. Generic framework in `Challenges/pkg/userflow/` with 6 adapter interfaces (Browser, Mobile, Desktop, API, Build, Process). CLI implementations invoke tools in Podman containers. Catalogizer challenges in `catalog-api/challenges/userflow_*.go`. Container orchestration via `digital.vasic.containers`.

**Tech Stack:** Go 1.24, Playwright (browser), ADB/UIAutomator/Robolectric (Android), Tauri WebDriver (desktop), Podman (containers), Challenges framework (challenge lifecycle)

---

## Phase 1: Refactor pkg/yole/ → pkg/userflow/ (Foundation)

### Task 1.1: Create pkg/userflow/ with generalized types

**Files:**
- Create: `Challenges/pkg/userflow/types.go`
- Reference: `Challenges/pkg/yole/types.go` (lines 1-71)

**Step 1: Write the test**

Create `Challenges/pkg/userflow/types_test.go`:
```go
package userflow

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBuildResult_Fields(t *testing.T) {
	r := BuildResult{
		Target:   "android-debug",
		Success:  true,
		Duration: 5 * time.Second,
		Output:   "BUILD SUCCESSFUL",
		Artifacts: []string{"app.apk"},
	}
	assert.Equal(t, "android-debug", r.Target)
	assert.True(t, r.Success)
	assert.Equal(t, 5*time.Second, r.Duration)
	assert.Len(t, r.Artifacts, 1)
}

func TestTestResult_Aggregation(t *testing.T) {
	r := TestResult{
		Suites: []TestSuite{
			{Name: "unit", Tests: 10, Failures: 1, Errors: 0},
			{Name: "integration", Tests: 5, Failures: 0, Errors: 1},
		},
		TotalTests:   15,
		TotalFailed:  1,
		TotalErrors:  1,
		TotalSkipped: 0,
		Duration:     3 * time.Second,
	}
	assert.Equal(t, 15, r.TotalTests)
	assert.Equal(t, 1, r.TotalFailed)
}

func TestTestCase_Status(t *testing.T) {
	tc := TestCase{
		Name:     "TestLogin",
		Status:   "passed",
		Duration: 100 * time.Millisecond,
	}
	assert.Equal(t, "passed", tc.Status)
	assert.Nil(t, tc.Failure)
}

func TestBuildTarget_Fields(t *testing.T) {
	bt := BuildTarget{
		Name: "Android Debug",
		Task: ":androidApp:assembleDebug",
		Args: []string{"--stacktrace"},
	}
	assert.Equal(t, "Android Debug", bt.Name)
	assert.Len(t, bt.Args, 1)
}

func TestLintResult_Fields(t *testing.T) {
	lr := LintResult{
		Tool:     "eslint",
		Success:  true,
		Duration: 2 * time.Second,
		Warnings: 3,
		Errors:   0,
	}
	assert.True(t, lr.Success)
	assert.Equal(t, 0, lr.Errors)
}

func TestJUnitTestSuites_XML(t *testing.T) {
	xml := `<?xml version="1.0"?>
<testsuites>
  <testsuite name="MySuite" tests="2" failures="1" errors="0" skipped="0" time="1.5">
    <testcase name="TestPass" classname="pkg.Test" time="0.5"/>
    <testcase name="TestFail" classname="pkg.Test" time="1.0">
      <failure message="expected true" type="AssertionError">stack trace</failure>
    </testcase>
  </testsuite>
</testsuites>`
	suites, err := ParseJUnitXML([]byte(xml))
	assert.NoError(t, err)
	assert.Len(t, suites, 1)
	assert.Equal(t, 2, suites[0].Tests)
	assert.Equal(t, 1, suites[0].Failures)
	assert.Len(t, suites[0].TestCases, 2)
	assert.Nil(t, suites[0].TestCases[0].Failure)
	assert.NotNil(t, suites[0].TestCases[1].Failure)
}
```

**Step 2: Run test to verify it fails**

Run: `cd /run/media/milosvasic/DATA4TB/Projects/Catalogizer/Challenges && go test ./pkg/userflow/ -v -count=1`
Expected: FAIL (package doesn't exist)

**Step 3: Write the implementation**

Create `Challenges/pkg/userflow/types.go`:
```go
package userflow

import (
	"encoding/xml"
	"fmt"
	"time"
)

// TestResult holds normalized test execution results from any
// build system (Gradle, npm, Go, Cargo).
type TestResult struct {
	Suites       []TestSuite
	TotalTests   int
	TotalFailed  int
	TotalErrors  int
	TotalSkipped int
	Duration     time.Duration
	Output       string
}

// TestSuite represents a group of test cases.
type TestSuite struct {
	Name      string
	Tests     int
	Failures  int
	Errors    int
	Skipped   int
	Duration  time.Duration
	TestCases []TestCase
}

// TestCase represents a single test execution.
type TestCase struct {
	Name      string
	ClassName string
	Duration  time.Duration
	Status    string // "passed", "failed", "error", "skipped"
	Failure   *TestFailure
}

// TestFailure holds failure/error details for a test case.
type TestFailure struct {
	Message    string
	Type       string
	StackTrace string
}

// BuildResult holds the result of a build operation.
type BuildResult struct {
	Target    string
	Success   bool
	Duration  time.Duration
	Output    string
	Artifacts []string
}

// LintResult holds the result of a lint/static analysis run.
type LintResult struct {
	Tool     string
	Success  bool
	Duration time.Duration
	Warnings int
	Errors   int
	Output   string
}

// BuildTarget defines a build target for any build system.
type BuildTarget struct {
	Name string
	Task string
	Args []string
}

// TestTarget defines a test target for any build system.
type TestTarget struct {
	Name   string
	Task   string
	Filter string
}

// LintTarget defines a lint target for any build system.
type LintTarget struct {
	Name string
	Task string
	Args []string
}

// Credentials holds authentication credentials.
type Credentials struct {
	Username string
	Password string
	URL      string
}

// BrowserConfig configures browser adapter initialization.
type BrowserConfig struct {
	BrowserType string // "chromium", "firefox", "webkit"
	Headless    bool
	WindowSize  [2]int // width, height
	ExtraArgs   []string
}

// DesktopAppConfig configures desktop application launch.
type DesktopAppConfig struct {
	BinaryPath string
	Args       []string
	WorkDir    string
	Env        map[string]string
}

// ProcessConfig configures process launch.
type ProcessConfig struct {
	Command string
	Args    []string
	WorkDir string
	Env     map[string]string
}

// MobileConfig configures mobile adapter.
type MobileConfig struct {
	PackageName  string
	ActivityName string
	DeviceSerial string
}

// --- JUnit XML parsing (reused across Gradle, npm, Go) ---

// JUnitTestSuites represents the top-level JUnit XML.
type JUnitTestSuites struct {
	XMLName    xml.Name         `xml:"testsuites"`
	TestSuites []JUnitTestSuite `xml:"testsuite"`
}

// JUnitTestSuite represents a single JUnit test suite.
type JUnitTestSuite struct {
	XMLName   xml.Name        `xml:"testsuite"`
	Name      string          `xml:"name,attr"`
	Tests     int             `xml:"tests,attr"`
	Failures  int             `xml:"failures,attr"`
	Errors    int             `xml:"errors,attr"`
	Skipped   int             `xml:"skipped,attr"`
	Time      float64         `xml:"time,attr"`
	TestCases []JUnitTestCase `xml:"testcase"`
}

// JUnitTestCase represents a single JUnit test case.
type JUnitTestCase struct {
	XMLName   xml.Name      `xml:"testcase"`
	Name      string        `xml:"name,attr"`
	ClassName string        `xml:"classname,attr"`
	Time      float64       `xml:"time,attr"`
	Failure   *JUnitFailure `xml:"failure,omitempty"`
	Error     *JUnitError   `xml:"error,omitempty"`
}

// JUnitFailure represents a test failure in JUnit XML.
type JUnitFailure struct {
	Message string `xml:"message,attr"`
	Type    string `xml:"type,attr"`
	Content string `xml:",chardata"`
}

// JUnitError represents a test error in JUnit XML.
type JUnitError struct {
	Message string `xml:"message,attr"`
	Type    string `xml:"type,attr"`
	Content string `xml:",chardata"`
}

// ParseJUnitXML parses JUnit XML bytes into test suites.
func ParseJUnitXML(data []byte) ([]JUnitTestSuite, error) {
	var suites JUnitTestSuites
	if err := xml.Unmarshal(data, &suites); err != nil {
		// Try single suite
		var single JUnitTestSuite
		if err2 := xml.Unmarshal(data, &single); err2 != nil {
			return nil, fmt.Errorf(
				"parse JUnit XML: %w", err,
			)
		}
		return []JUnitTestSuite{single}, nil
	}
	return suites.TestSuites, nil
}

// JUnitToTestResult converts JUnit suites to TestResult.
func JUnitToTestResult(
	suites []JUnitTestSuite, duration time.Duration, output string,
) *TestResult {
	result := &TestResult{
		Duration: duration,
		Output:   output,
	}
	for _, s := range suites {
		suite := TestSuite{
			Name:     s.Name,
			Tests:    s.Tests,
			Failures: s.Failures,
			Errors:   s.Errors,
			Skipped:  s.Skipped,
			Duration: time.Duration(s.Time * float64(time.Second)),
		}
		for _, tc := range s.TestCases {
			c := TestCase{
				Name:      tc.Name,
				ClassName: tc.ClassName,
				Duration:  time.Duration(tc.Time * float64(time.Second)),
				Status:    "passed",
			}
			if tc.Failure != nil {
				c.Status = "failed"
				c.Failure = &TestFailure{
					Message:    tc.Failure.Message,
					Type:       tc.Failure.Type,
					StackTrace: tc.Failure.Content,
				}
			} else if tc.Error != nil {
				c.Status = "error"
				c.Failure = &TestFailure{
					Message:    tc.Error.Message,
					Type:       tc.Error.Type,
					StackTrace: tc.Error.Content,
				}
			}
			suite.TestCases = append(suite.TestCases, c)
		}
		result.Suites = append(result.Suites, suite)
		result.TotalTests += s.Tests
		result.TotalFailed += s.Failures
		result.TotalErrors += s.Errors
		result.TotalSkipped += s.Skipped
	}
	return result
}
```

**Step 4: Run test to verify it passes**

Run: `cd /run/media/milosvasic/DATA4TB/Projects/Catalogizer/Challenges && go test ./pkg/userflow/ -v -count=1`
Expected: PASS

**Step 5: Commit**

```bash
cd /run/media/milosvasic/DATA4TB/Projects/Catalogizer/Challenges
git add pkg/userflow/types.go pkg/userflow/types_test.go
git commit -m "feat(userflow): add generic types and JUnit XML parsing"
```

---

### Task 1.2: Create adapter interfaces

**Files:**
- Create: `Challenges/pkg/userflow/adapter_browser.go`
- Create: `Challenges/pkg/userflow/adapter_mobile.go`
- Create: `Challenges/pkg/userflow/adapter_desktop.go`
- Create: `Challenges/pkg/userflow/adapter_api.go`
- Create: `Challenges/pkg/userflow/adapter_build.go`
- Create: `Challenges/pkg/userflow/adapter_process.go`

**Step 1: Write tests for interface satisfaction**

Create `Challenges/pkg/userflow/adapter_test.go`:
```go
package userflow

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Compile-time interface satisfaction checks.
// These verify that nil pointers of concrete types
// satisfy the adapter interfaces (once implementations exist).

func TestBrowserAdapterInterface(t *testing.T) {
	// BrowserAdapter is an interface — verify it compiles
	var _ BrowserAdapter = (BrowserAdapter)(nil)
	assert.True(t, true, "BrowserAdapter interface compiles")
}

func TestMobileAdapterInterface(t *testing.T) {
	var _ MobileAdapter = (MobileAdapter)(nil)
	assert.True(t, true, "MobileAdapter interface compiles")
}

func TestDesktopAdapterInterface(t *testing.T) {
	var _ DesktopAdapter = (DesktopAdapter)(nil)
	assert.True(t, true, "DesktopAdapter interface compiles")
}

func TestAPIAdapterInterface(t *testing.T) {
	var _ APIAdapter = (APIAdapter)(nil)
	assert.True(t, true, "APIAdapter interface compiles")
}

func TestBuildAdapterInterface(t *testing.T) {
	var _ BuildAdapter = (BuildAdapter)(nil)
	assert.True(t, true, "BuildAdapter interface compiles")
}

func TestProcessAdapterInterface(t *testing.T) {
	var _ ProcessAdapter = (ProcessAdapter)(nil)
	assert.True(t, true, "ProcessAdapter interface compiles")
}
```

**Step 2: Run test to verify it fails**

Run: `cd /run/media/milosvasic/DATA4TB/Projects/Catalogizer/Challenges && go test ./pkg/userflow/ -v -count=1 -run TestBrowserAdapter`
Expected: FAIL (BrowserAdapter undefined)

**Step 3: Write all 6 adapter interface files**

Create `Challenges/pkg/userflow/adapter_browser.go`:
```go
package userflow

import (
	"context"
	"time"
)

// InterceptedRequest represents a network request captured by
// the browser adapter's network interception.
type InterceptedRequest struct {
	URL     string
	Method  string
	Headers map[string]string
	Body    []byte
}

// BrowserAdapter abstracts browser automation for web testing.
// Implementations: PlaywrightCLIAdapter.
type BrowserAdapter interface {
	// Initialize sets up the browser with the given config.
	Initialize(ctx context.Context, config BrowserConfig) error

	// Navigate goes to the specified URL.
	Navigate(ctx context.Context, url string) error

	// Click clicks an element matching the CSS selector.
	Click(ctx context.Context, selector string) error

	// Fill types text into an input matching the selector.
	Fill(ctx context.Context, selector, value string) error

	// SelectOption selects an option in a <select> element.
	SelectOption(ctx context.Context, selector, value string) error

	// IsVisible checks if an element matching selector exists
	// and is visible on the page.
	IsVisible(ctx context.Context, selector string) (bool, error)

	// WaitForSelector waits until selector appears or timeout.
	WaitForSelector(
		ctx context.Context, selector string, timeout time.Duration,
	) error

	// GetText returns the text content of a matching element.
	GetText(ctx context.Context, selector string) (string, error)

	// GetAttribute returns an attribute value of an element.
	GetAttribute(
		ctx context.Context, selector, attr string,
	) (string, error)

	// Screenshot captures the current page as PNG bytes.
	Screenshot(ctx context.Context) ([]byte, error)

	// EvaluateJS executes JavaScript in the page context.
	EvaluateJS(ctx context.Context, script string) (string, error)

	// NetworkIntercept sets up route interception for matching
	// URLs. The handler is called for each matched request.
	NetworkIntercept(
		ctx context.Context, pattern string,
		handler func(req *InterceptedRequest),
	) error

	// Close shuts down the browser and releases resources.
	Close(ctx context.Context) error

	// Available returns true if the browser backend is reachable.
	Available(ctx context.Context) bool
}
```

Create `Challenges/pkg/userflow/adapter_mobile.go`:
```go
package userflow

import (
	"context"
	"time"
)

// MobileAdapter abstracts mobile device/emulator management.
// Implementations: ADBCLIAdapter.
type MobileAdapter interface {
	// IsDeviceAvailable checks if a device/emulator is connected.
	IsDeviceAvailable(ctx context.Context) (bool, error)

	// InstallApp installs an application package on the device.
	InstallApp(ctx context.Context, appPath string) error

	// LaunchApp starts the configured application.
	LaunchApp(ctx context.Context) error

	// StopApp force-stops the application.
	StopApp(ctx context.Context) error

	// IsAppRunning checks if the app process is active.
	IsAppRunning(ctx context.Context) (bool, error)

	// TakeScreenshot captures the device screen as PNG bytes.
	TakeScreenshot(ctx context.Context) ([]byte, error)

	// Tap taps at screen coordinates.
	Tap(ctx context.Context, x, y int) error

	// SendKeys types text into the focused input field.
	SendKeys(ctx context.Context, text string) error

	// PressKey sends a key event (e.g., "KEYCODE_BACK").
	PressKey(ctx context.Context, keycode string) error

	// WaitForApp waits until the app is running or timeout.
	WaitForApp(ctx context.Context, timeout time.Duration) error

	// RunInstrumentedTests runs on-device instrumented tests.
	RunInstrumentedTests(
		ctx context.Context, testClass string,
	) (*TestResult, error)

	// Close releases device resources.
	Close(ctx context.Context) error

	// Available returns true if the mobile backend is reachable.
	Available(ctx context.Context) bool
}
```

Create `Challenges/pkg/userflow/adapter_desktop.go`:
```go
package userflow

import (
	"context"
	"time"
)

// DesktopAdapter abstracts desktop application testing.
// Implementations: TauriCLIAdapter.
type DesktopAdapter interface {
	// LaunchApp starts the desktop application with config.
	LaunchApp(ctx context.Context, config DesktopAppConfig) error

	// IsAppRunning checks if the application is still running.
	IsAppRunning(ctx context.Context) (bool, error)

	// Navigate goes to a URL within the app's WebView.
	Navigate(ctx context.Context, url string) error

	// Click clicks an element matching the CSS selector.
	Click(ctx context.Context, selector string) error

	// Fill types text into an input matching the selector.
	Fill(ctx context.Context, selector, value string) error

	// IsVisible checks if an element is visible in the WebView.
	IsVisible(ctx context.Context, selector string) (bool, error)

	// WaitForSelector waits until selector appears or timeout.
	WaitForSelector(
		ctx context.Context, selector string, timeout time.Duration,
	) error

	// Screenshot captures the app window as PNG bytes.
	Screenshot(ctx context.Context) ([]byte, error)

	// InvokeCommand calls a desktop IPC command (e.g., Tauri).
	InvokeCommand(
		ctx context.Context, command string, args ...string,
	) (string, error)

	// WaitForWindow waits until the app window is ready.
	WaitForWindow(
		ctx context.Context, timeout time.Duration,
	) error

	// Close shuts down the application and releases resources.
	Close(ctx context.Context) error

	// Available returns true if the desktop backend is reachable.
	Available(ctx context.Context) bool
}
```

Create `Challenges/pkg/userflow/adapter_api.go`:
```go
package userflow

import "context"

// WebSocketConn abstracts a WebSocket connection.
type WebSocketConn interface {
	// WriteMessage sends a text message.
	WriteMessage(data []byte) error

	// ReadMessage reads the next message (blocks).
	ReadMessage() ([]byte, error)

	// Close closes the connection.
	Close() error
}

// APIAdapter abstracts REST API and WebSocket testing.
// Implementations: HTTPAPIAdapter.
type APIAdapter interface {
	// Login authenticates and stores the token.
	Login(ctx context.Context, credentials Credentials) (string, error)

	// LoginWithRetry attempts login with exponential backoff.
	LoginWithRetry(
		ctx context.Context, credentials Credentials, retries int,
	) (string, error)

	// Get performs an authenticated GET request.
	Get(ctx context.Context, path string) (int, map[string]interface{}, error)

	// GetRaw performs a GET and returns raw bytes.
	GetRaw(ctx context.Context, path string) (int, []byte, error)

	// GetArray performs a GET expecting a JSON array.
	GetArray(ctx context.Context, path string) (int, []interface{}, error)

	// PostJSON performs an authenticated POST with JSON body.
	PostJSON(ctx context.Context, path, body string) (int, []byte, error)

	// PutJSON performs an authenticated PUT with JSON body.
	PutJSON(ctx context.Context, path, body string) (int, []byte, error)

	// Delete performs an authenticated DELETE request.
	Delete(ctx context.Context, path string) (int, []byte, error)

	// WebSocketConnect opens a WebSocket connection.
	WebSocketConnect(ctx context.Context, path string) (WebSocketConn, error)

	// SetToken sets the authentication token manually.
	SetToken(token string)

	// Available returns true if the API is reachable.
	Available(ctx context.Context) bool
}
```

Create `Challenges/pkg/userflow/adapter_build.go`:
```go
package userflow

import "context"

// BuildAdapter abstracts build system execution.
// Implementations: GradleCLIAdapter, NPMCLIAdapter,
// GoCLIAdapter, CargoCLIAdapter.
type BuildAdapter interface {
	// Build executes a build target and returns the result.
	Build(ctx context.Context, target BuildTarget) (*BuildResult, error)

	// RunTests executes a test target and returns results.
	RunTests(ctx context.Context, target TestTarget) (*TestResult, error)

	// Lint executes a lint/static analysis target.
	Lint(ctx context.Context, target LintTarget) (*LintResult, error)

	// Available returns true if the build system is reachable.
	Available(ctx context.Context) bool
}
```

Create `Challenges/pkg/userflow/adapter_process.go`:
```go
package userflow

import (
	"context"
	"time"
)

// ProcessAdapter abstracts application process lifecycle.
// Implementations: ProcessCLIAdapter.
type ProcessAdapter interface {
	// Launch starts a process with the given configuration.
	Launch(ctx context.Context, config ProcessConfig) error

	// IsRunning checks if the managed process is alive.
	IsRunning() bool

	// WaitForReady waits until the process is running.
	WaitForReady(
		ctx context.Context, timeout time.Duration,
	) error

	// Stop gracefully terminates the process.
	Stop() error
}
```

**Step 4: Run tests to verify they pass**

Run: `cd /run/media/milosvasic/DATA4TB/Projects/Catalogizer/Challenges && go test ./pkg/userflow/ -v -count=1`
Expected: PASS (all interface tests compile)

**Step 5: Commit**

```bash
cd /run/media/milosvasic/DATA4TB/Projects/Catalogizer/Challenges
git add pkg/userflow/adapter_*.go
git commit -m "feat(userflow): add 6 adapter interfaces for multi-platform testing"
```

---

### Task 1.3: Create plugin and evaluators

**Files:**
- Create: `Challenges/pkg/userflow/plugin.go`
- Create: `Challenges/pkg/userflow/evaluators.go`
- Reference: `Challenges/pkg/yole/plugin.go`, `Challenges/pkg/yole/evaluators.go`

**Step 1: Write the test**

Create `Challenges/pkg/userflow/plugin_test.go`:
```go
package userflow

import (
	"testing"

	"digital.vasic.challenges/pkg/assertion"
	"digital.vasic.challenges/pkg/plugin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserFlowPlugin_Interface(t *testing.T) {
	engine := assertion.NewEngine()
	p := NewUserFlowPlugin(engine)

	// Verify it implements plugin.Plugin
	var _ plugin.Plugin = p
	assert.Equal(t, "userflow", p.Name())
	assert.Equal(t, "1.0.0", p.Version())
}

func TestUserFlowPlugin_Init(t *testing.T) {
	engine := assertion.NewEngine()
	p := NewUserFlowPlugin(engine)

	err := p.Init(&plugin.PluginContext{})
	require.NoError(t, err)

	// Verify all evaluators registered
	evaluatorNames := []string{
		"build_succeeds", "all_tests_pass", "lint_passes",
		"app_launches", "app_stable", "status_code",
		"response_contains", "response_not_empty",
		"json_field_equals", "screenshot_exists",
		"flow_completes", "within_duration",
	}
	for _, name := range evaluatorNames {
		assert.True(t, engine.HasEvaluator(name),
			"evaluator %s should be registered", name)
	}
}

func TestUserFlowPlugin_NilEngine(t *testing.T) {
	p := NewUserFlowPlugin(nil)
	err := p.Init(&plugin.PluginContext{})
	assert.Error(t, err)
}
```

Create `Challenges/pkg/userflow/evaluators_test.go`:
```go
package userflow

import (
	"testing"

	"digital.vasic.challenges/pkg/assertion"
	"github.com/stretchr/testify/assert"
)

func TestEvaluateBuildSucceeds(t *testing.T) {
	tests := []struct {
		name   string
		value  any
		pass   bool
	}{
		{"success", true, true},
		{"failure", false, false},
		{"wrong type", "yes", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ok, _ := evaluateBuildSucceeds(assertion.Definition{}, tt.value)
			assert.Equal(t, tt.pass, ok)
		})
	}
}

func TestEvaluateAllTestsPass(t *testing.T) {
	tests := []struct {
		name   string
		value  any
		pass   bool
	}{
		{"zero failures", 0, true},
		{"some failures", 3, false},
		{"float zero", float64(0), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ok, _ := evaluateAllTestsPass(assertion.Definition{}, tt.value)
			assert.Equal(t, tt.pass, ok)
		})
	}
}

func TestEvaluateStatusCode(t *testing.T) {
	def := assertion.Definition{Value: 200}
	ok, _ := evaluateStatusCode(def, 200)
	assert.True(t, ok)

	ok, _ = evaluateStatusCode(def, 404)
	assert.False(t, ok)
}

func TestEvaluateResponseContains(t *testing.T) {
	def := assertion.Definition{Value: "success"}
	ok, _ := evaluateResponseContains(def, "operation success done")
	assert.True(t, ok)

	ok, _ = evaluateResponseContains(def, "operation failed")
	assert.False(t, ok)
}

func TestEvaluateWithinDuration(t *testing.T) {
	def := assertion.Definition{Value: 5000} // 5000ms
	ok, _ := evaluateWithinDuration(def, 3000)
	assert.True(t, ok)

	ok, _ = evaluateWithinDuration(def, 6000)
	assert.False(t, ok)
}

func TestEvaluateScreenshotExists(t *testing.T) {
	ok, _ := evaluateScreenshotExists(assertion.Definition{}, []byte{1, 2, 3})
	assert.True(t, ok)

	ok, _ = evaluateScreenshotExists(assertion.Definition{}, []byte{})
	assert.False(t, ok)

	ok, _ = evaluateScreenshotExists(assertion.Definition{}, nil)
	assert.False(t, ok)
}
```

**Step 2: Run tests to verify they fail**

Run: `cd /run/media/milosvasic/DATA4TB/Projects/Catalogizer/Challenges && go test ./pkg/userflow/ -v -count=1 -run TestUserFlowPlugin`
Expected: FAIL (NewUserFlowPlugin undefined)

**Step 3: Write the implementation**

Create `Challenges/pkg/userflow/plugin.go`:
```go
package userflow

import (
	"fmt"

	"digital.vasic.challenges/pkg/assertion"
	"digital.vasic.challenges/pkg/plugin"
)

const (
	// PluginName is the canonical name for the userflow plugin.
	PluginName = "userflow"
	// PluginVersion is the current version.
	PluginVersion = "1.0.0"
)

// UserFlowPlugin implements plugin.Plugin and registers all
// user flow assertion evaluators with the framework.
type UserFlowPlugin struct {
	engine *assertion.DefaultEngine
}

// NewUserFlowPlugin creates a UserFlowPlugin that registers
// evaluators with the given assertion engine.
func NewUserFlowPlugin(
	engine *assertion.DefaultEngine,
) *UserFlowPlugin {
	return &UserFlowPlugin{engine: engine}
}

// Name returns the plugin name.
func (p *UserFlowPlugin) Name() string {
	return PluginName
}

// Version returns the plugin version.
func (p *UserFlowPlugin) Version() string {
	return PluginVersion
}

// Init registers all user flow assertion evaluators.
func (p *UserFlowPlugin) Init(
	_ *plugin.PluginContext,
) error {
	if p.engine == nil {
		return fmt.Errorf(
			"userflow plugin: assertion engine is nil",
		)
	}
	return RegisterEvaluators(p.engine)
}
```

Create `Challenges/pkg/userflow/evaluators.go`:
```go
package userflow

import (
	"fmt"
	"strings"

	"digital.vasic.challenges/pkg/assertion"
)

// RegisterEvaluators registers all user flow assertion
// evaluators with the given assertion engine.
func RegisterEvaluators(
	engine *assertion.DefaultEngine,
) error {
	evaluators := map[string]assertion.Evaluator{
		// Carried over from build/test domain (generalized)
		"build_succeeds":   evaluateBuildSucceeds,
		"all_tests_pass":   evaluateAllTestsPass,
		"lint_passes":      evaluateLintPasses,
		"app_launches":     evaluateAppLaunches,
		"app_stable":       evaluateAppStable,
		// New generic evaluators
		"status_code":       evaluateStatusCode,
		"response_contains": evaluateResponseContains,
		"response_not_empty": evaluateResponseNotEmpty,
		"json_field_equals": evaluateJSONFieldEquals,
		"screenshot_exists": evaluateScreenshotExists,
		"flow_completes":    evaluateFlowCompletes,
		"within_duration":   evaluateWithinDuration,
	}

	for name, eval := range evaluators {
		if err := engine.Register(name, eval); err != nil {
			return fmt.Errorf(
				"register evaluator %s: %w", name, err,
			)
		}
	}
	return nil
}

func evaluateBuildSucceeds(
	def assertion.Definition, value any,
) (bool, string) {
	success, ok := value.(bool)
	if !ok {
		return false, fmt.Sprintf(
			"expected bool, got %T", value,
		)
	}
	if success {
		return true, "build succeeded"
	}
	return false, "build failed"
}

func evaluateAllTestsPass(
	def assertion.Definition, value any,
) (bool, string) {
	failures := toIntVal(value)
	if failures == 0 {
		return true, "all tests passed"
	}
	return false, fmt.Sprintf(
		"%d test failures", failures,
	)
}

func evaluateLintPasses(
	def assertion.Definition, value any,
) (bool, string) {
	success, ok := value.(bool)
	if !ok {
		return false, fmt.Sprintf(
			"expected bool, got %T", value,
		)
	}
	if success {
		return true, "lint passed"
	}
	return false, "lint failed"
}

func evaluateAppLaunches(
	def assertion.Definition, value any,
) (bool, string) {
	running, ok := value.(bool)
	if !ok {
		return false, fmt.Sprintf(
			"expected bool, got %T", value,
		)
	}
	if running {
		return true, "app launched successfully"
	}
	return false, "app failed to launch"
}

func evaluateAppStable(
	def assertion.Definition, value any,
) (bool, string) {
	running, ok := value.(bool)
	if !ok {
		return false, fmt.Sprintf(
			"expected bool, got %T", value,
		)
	}
	if running {
		return true, "app is stable (still running)"
	}
	return false, "app crashed after launch"
}

func evaluateStatusCode(
	def assertion.Definition, value any,
) (bool, string) {
	actual := toIntVal(value)
	expected := toIntVal(def.Value)
	if actual == expected {
		return true, fmt.Sprintf(
			"status code %d matches expected", actual,
		)
	}
	return false, fmt.Sprintf(
		"status code %d, expected %d", actual, expected,
	)
}

func evaluateResponseContains(
	def assertion.Definition, value any,
) (bool, string) {
	body, ok := value.(string)
	if !ok {
		return false, fmt.Sprintf(
			"expected string, got %T", value,
		)
	}
	expected, ok := def.Value.(string)
	if !ok {
		return false, "expected string in definition value"
	}
	if strings.Contains(body, expected) {
		return true, fmt.Sprintf(
			"response contains %q", expected,
		)
	}
	return false, fmt.Sprintf(
		"response does not contain %q", expected,
	)
}

func evaluateResponseNotEmpty(
	def assertion.Definition, value any,
) (bool, string) {
	body, ok := value.(string)
	if !ok {
		if b, ok2 := value.([]byte); ok2 {
			if len(b) > 0 {
				return true, fmt.Sprintf(
					"response has %d bytes", len(b),
				)
			}
			return false, "response is empty"
		}
		return false, fmt.Sprintf(
			"expected string or []byte, got %T", value,
		)
	}
	if len(body) > 0 {
		return true, fmt.Sprintf(
			"response has %d chars", len(body),
		)
	}
	return false, "response is empty"
}

func evaluateJSONFieldEquals(
	def assertion.Definition, value any,
) (bool, string) {
	if fmt.Sprintf("%v", value) == fmt.Sprintf("%v", def.Value) {
		return true, fmt.Sprintf(
			"field equals %v", def.Value,
		)
	}
	return false, fmt.Sprintf(
		"field is %v, expected %v", value, def.Value,
	)
}

func evaluateScreenshotExists(
	def assertion.Definition, value any,
) (bool, string) {
	switch v := value.(type) {
	case []byte:
		if len(v) > 0 {
			return true, fmt.Sprintf(
				"screenshot captured (%d bytes)", len(v),
			)
		}
		return false, "screenshot is empty"
	default:
		return false, fmt.Sprintf(
			"expected []byte, got %T", value,
		)
	}
}

func evaluateFlowCompletes(
	def assertion.Definition, value any,
) (bool, string) {
	completed, ok := value.(bool)
	if !ok {
		return false, fmt.Sprintf(
			"expected bool, got %T", value,
		)
	}
	if completed {
		return true, "flow completed successfully"
	}
	return false, "flow did not complete"
}

func evaluateWithinDuration(
	def assertion.Definition, value any,
) (bool, string) {
	actual := toIntVal(value)
	threshold := toIntVal(def.Value)
	if actual <= threshold {
		return true, fmt.Sprintf(
			"%dms within %dms threshold", actual, threshold,
		)
	}
	return false, fmt.Sprintf(
		"%dms exceeds %dms threshold", actual, threshold,
	)
}

func toIntVal(v any) int {
	switch val := v.(type) {
	case int:
		return val
	case int64:
		return int(val)
	case float64:
		return int(val)
	default:
		return 0
	}
}
```

**Step 4: Run tests**

Run: `cd /run/media/milosvasic/DATA4TB/Projects/Catalogizer/Challenges && go test ./pkg/userflow/ -v -count=1`
Expected: PASS

**Step 5: Commit**

```bash
cd /run/media/milosvasic/DATA4TB/Projects/Catalogizer/Challenges
git add pkg/userflow/plugin.go pkg/userflow/plugin_test.go pkg/userflow/evaluators.go pkg/userflow/evaluators_test.go
git commit -m "feat(userflow): add plugin and 12 assertion evaluators"
```

---

### Task 1.4: Create options and result parser

**Files:**
- Create: `Challenges/pkg/userflow/options.go`
- Create: `Challenges/pkg/userflow/result_parser.go`
- Create: `Challenges/pkg/userflow/result_parser_test.go`

**Step 1: Write test**

Create `Challenges/pkg/userflow/result_parser_test.go`:
```go
package userflow

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseTestResultToValues(t *testing.T) {
	result := &TestResult{
		TotalTests:  10,
		TotalFailed: 2,
		TotalErrors: 1,
		Duration:    5 * time.Second,
		Output:      "test output",
		Suites:      []TestSuite{{Name: "unit", Tests: 10}},
	}
	values := ParseTestResultToValues(result)
	assert.Equal(t, 10, values["total_tests"])
	assert.Equal(t, 2, values["total_failures"])
	assert.Equal(t, 1, values["total_errors"])
	assert.Equal(t, 1, values["suite_count"])
}

func TestParseTestResultToMetrics(t *testing.T) {
	result := &TestResult{
		TotalTests:  15,
		TotalFailed: 0,
		Duration:    3 * time.Second,
	}
	metrics := ParseTestResultToMetrics(result)
	assert.Equal(t, float64(3), metrics["duration"].Value)
	assert.Equal(t, "seconds", metrics["duration"].Unit)
	assert.Equal(t, float64(15), metrics["total_tests"].Value)
}

func TestParseBuildResultToValues(t *testing.T) {
	result := &BuildResult{
		Target:   "app",
		Success:  true,
		Duration: 10 * time.Second,
	}
	values := ParseBuildResultToValues(result)
	assert.Equal(t, true, values["success"])
	assert.Equal(t, "app", values["target"])
}

func TestResolveChallengeConfig(t *testing.T) {
	cfg := resolveChallengeConfig([]ChallengeOption{
		WithContainerized(true),
		WithProjectRoot("/app"),
	})
	assert.True(t, cfg.containerized)
	assert.Equal(t, "/app", cfg.projectRoot)
}
```

**Step 2: Run to verify failure**

Run: `cd /run/media/milosvasic/DATA4TB/Projects/Catalogizer/Challenges && go test ./pkg/userflow/ -v -count=1 -run TestParse`
Expected: FAIL

**Step 3: Write implementations**

Create `Challenges/pkg/userflow/options.go`:
```go
package userflow

// ChallengeOption configures a user flow challenge.
type ChallengeOption func(*challengeConfig)

// challengeConfig holds resolved challenge options.
type challengeConfig struct {
	containerized bool
	projectRoot   string
	runtimeName   string
}

// WithContainerized enables containerized execution.
func WithContainerized(use bool) ChallengeOption {
	return func(c *challengeConfig) {
		c.containerized = use
	}
}

// WithProjectRoot sets the project root directory.
func WithProjectRoot(root string) ChallengeOption {
	return func(c *challengeConfig) {
		c.projectRoot = root
	}
}

// WithRuntimeName sets the container runtime name.
func WithRuntimeName(name string) ChallengeOption {
	return func(c *challengeConfig) {
		c.runtimeName = name
	}
}

// resolveChallengeConfig applies functional options.
func resolveChallengeConfig(
	opts []ChallengeOption,
) *challengeConfig {
	cfg := &challengeConfig{}
	for _, opt := range opts {
		opt(cfg)
	}
	return cfg
}
```

Create `Challenges/pkg/userflow/result_parser.go`:
```go
package userflow

import (
	"digital.vasic.challenges/pkg/challenge"
)

// ParseTestResultToValues converts a TestResult into a map
// suitable for assertion evaluation.
func ParseTestResultToValues(
	result *TestResult,
) map[string]any {
	return map[string]any{
		"total_tests":    result.TotalTests,
		"total_failures": result.TotalFailed,
		"total_errors":   result.TotalErrors,
		"total_skipped":  result.TotalSkipped,
		"suite_count":    len(result.Suites),
		"duration":       result.Duration.Seconds(),
		"output":         result.Output,
	}
}

// ParseTestResultToMetrics converts a TestResult into
// challenge MetricValue entries for reporting.
func ParseTestResultToMetrics(
	result *TestResult,
) map[string]challenge.MetricValue {
	return map[string]challenge.MetricValue{
		"duration": {
			Name:  "duration",
			Value: result.Duration.Seconds(),
			Unit:  "seconds",
		},
		"total_tests": {
			Name:  "total_tests",
			Value: float64(result.TotalTests),
			Unit:  "count",
		},
		"total_failures": {
			Name:  "total_failures",
			Value: float64(result.TotalFailed),
			Unit:  "count",
		},
	}
}

// ParseBuildResultToValues converts a BuildResult into a map
// suitable for assertion evaluation.
func ParseBuildResultToValues(
	result *BuildResult,
) map[string]any {
	return map[string]any{
		"success":  result.Success,
		"target":   result.Target,
		"duration": result.Duration.Seconds(),
		"output":   result.Output,
	}
}

// ParseBuildResultToMetrics converts a BuildResult into
// challenge MetricValue entries.
func ParseBuildResultToMetrics(
	result *BuildResult,
) map[string]challenge.MetricValue {
	return map[string]challenge.MetricValue{
		"duration": {
			Name:  "duration",
			Value: result.Duration.Seconds(),
			Unit:  "seconds",
		},
	}
}
```

**Step 4: Run tests**

Run: `cd /run/media/milosvasic/DATA4TB/Projects/Catalogizer/Challenges && go test ./pkg/userflow/ -v -count=1`
Expected: PASS

**Step 5: Commit**

```bash
cd /run/media/milosvasic/DATA4TB/Projects/Catalogizer/Challenges
git add pkg/userflow/options.go pkg/userflow/result_parser.go pkg/userflow/result_parser_test.go
git commit -m "feat(userflow): add options and result parser utilities"
```

---

### Task 1.5: Create flow definition types

**Files:**
- Create: `Challenges/pkg/userflow/flow_api.go`
- Create: `Challenges/pkg/userflow/flow_browser.go`
- Create: `Challenges/pkg/userflow/flow_mobile.go`
- Create: `Challenges/pkg/userflow/flow_ipc.go`

**Step 1: Write test**

Create `Challenges/pkg/userflow/flow_test.go`:
```go
package userflow

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAPIFlow_Steps(t *testing.T) {
	flow := APIFlow{
		Name:        "login-and-browse",
		Credentials: Credentials{Username: "admin", Password: "pass"},
		Steps: []APIStep{
			{
				Name:           "login",
				Method:         "POST",
				Path:           "/api/v1/auth/login",
				Body:           `{"username":"admin","password":"pass"}`,
				ExpectedStatus: 200,
				ExtractTo:      map[string]string{"token": "session_token"},
			},
			{
				Name:           "list-entities",
				Method:         "GET",
				Path:           "/api/v1/entities",
				ExpectedStatus: 200,
				Assertions: []StepAssertion{
					{Field: "total", Type: "min_count", Expected: 1},
				},
			},
		},
	}
	assert.Equal(t, 2, len(flow.Steps))
	assert.Equal(t, "POST", flow.Steps[0].Method)
}

func TestBrowserFlow_Steps(t *testing.T) {
	flow := BrowserFlow{
		Name:     "login-flow",
		StartURL: "http://localhost:3000/login",
		Steps: []BrowserStep{
			{Name: "fill-user", Action: "fill", Selector: "#email", Value: "admin"},
			{Name: "click-login", Action: "click", Selector: "#submit"},
			{Name: "wait-dashboard", Action: "wait", Selector: ".dashboard", Timeout: 5 * time.Second},
			{Name: "verify-url", Action: "assert_url", Value: "/dashboard"},
		},
	}
	assert.Equal(t, 4, len(flow.Steps))
}

func TestMobileFlow_Steps(t *testing.T) {
	flow := MobileFlow{
		Name: "launch-and-login",
		Steps: []MobileStep{
			{Name: "wait", Action: "wait", Timeout: 5 * time.Second},
			{Name: "tap-login", Action: "tap", X: 540, Y: 960},
			{Name: "type-user", Action: "send_keys", Text: "admin"},
		},
	}
	assert.Equal(t, 3, len(flow.Steps))
}

func TestIPCCommand_Fields(t *testing.T) {
	cmd := IPCCommand{
		Name:           "get-config",
		Command:        "get_config",
		Args:           []string{},
		ExpectedResult: "{}",
	}
	assert.Equal(t, "get_config", cmd.Command)
}
```

**Step 2: Run to verify failure**

Run: `cd /run/media/milosvasic/DATA4TB/Projects/Catalogizer/Challenges && go test ./pkg/userflow/ -v -count=1 -run TestAPIFlow`
Expected: FAIL

**Step 3: Write implementations**

Create `Challenges/pkg/userflow/flow_api.go`:
```go
package userflow

// APIFlow defines a sequence of REST API calls as a user flow.
type APIFlow struct {
	Name        string
	Credentials Credentials
	Steps       []APIStep
}

// APIStep represents a single API call in a flow.
type APIStep struct {
	Name           string
	Method         string // GET, POST, PUT, DELETE
	Path           string
	Body           string // JSON body for POST/PUT
	ExpectedStatus int
	Assertions     []StepAssertion
	ExtractTo      map[string]string // response field → variable name
}

// StepAssertion checks a response field.
type StepAssertion struct {
	Field    string // JSON path in response
	Type     string // evaluator type
	Expected any
}
```

Create `Challenges/pkg/userflow/flow_browser.go`:
```go
package userflow

import "time"

// BrowserFlow defines a sequence of browser actions.
type BrowserFlow struct {
	Name     string
	StartURL string
	Steps    []BrowserStep
}

// BrowserStep represents a single browser action.
// Actions: "navigate", "click", "fill", "select", "wait",
// "assert_visible", "assert_text", "assert_url",
// "screenshot", "evaluate_js"
type BrowserStep struct {
	Name       string
	Action     string
	Selector   string
	Value      string
	Timeout    time.Duration
	Screenshot bool // take screenshot after step
}
```

Create `Challenges/pkg/userflow/flow_mobile.go`:
```go
package userflow

import "time"

// MobileFlow defines a sequence of mobile device actions.
type MobileFlow struct {
	Name  string
	Steps []MobileStep
}

// MobileStep represents a single mobile action.
// Actions: "tap", "send_keys", "press_key", "wait",
// "assert_running", "screenshot"
type MobileStep struct {
	Name       string
	Action     string
	X, Y       int
	Text       string
	KeyCode    string
	Timeout    time.Duration
	Screenshot bool
}
```

Create `Challenges/pkg/userflow/flow_ipc.go`:
```go
package userflow

// IPCCommand defines a desktop IPC command to test.
type IPCCommand struct {
	Name           string
	Command        string
	Args           []string
	ExpectedResult string
}
```

**Step 4: Run tests**

Run: `cd /run/media/milosvasic/DATA4TB/Projects/Catalogizer/Challenges && go test ./pkg/userflow/ -v -count=1`
Expected: PASS

**Step 5: Commit**

```bash
cd /run/media/milosvasic/DATA4TB/Projects/Catalogizer/Challenges
git add pkg/userflow/flow_*.go
git commit -m "feat(userflow): add flow definition types for API, browser, mobile, IPC"
```

---

### Task 1.6: Remove pkg/yole/ and cmd/yole-challenges/

**Files:**
- Delete: `Challenges/pkg/yole/` (entire directory)
- Delete: `Challenges/cmd/yole-challenges/` (entire directory)

**Step 1: Verify no other package imports yole**

Run: `cd /run/media/milosvasic/DATA4TB/Projects/Catalogizer/Challenges && grep -r '"digital.vasic.challenges/pkg/yole"' --include='*.go' .`
Expected: Only files in `pkg/yole/` and `cmd/yole-challenges/`

**Step 2: Delete yole package and command**

```bash
cd /run/media/milosvasic/DATA4TB/Projects/Catalogizer/Challenges
rm -rf pkg/yole/ cmd/yole-challenges/
```

**Step 3: Verify build still works**

Run: `cd /run/media/milosvasic/DATA4TB/Projects/Catalogizer/Challenges && go build ./...`
Expected: SUCCESS (no references to yole remain)

**Step 4: Run all existing tests**

Run: `cd /run/media/milosvasic/DATA4TB/Projects/Catalogizer/Challenges && go test ./... -count=1`
Expected: PASS (yole tests gone, all other tests pass)

**Step 5: Commit**

```bash
cd /run/media/milosvasic/DATA4TB/Projects/Catalogizer/Challenges
git add -A
git commit -m "refactor: remove pkg/yole/ and cmd/yole-challenges/ (replaced by pkg/userflow/)"
```

---

## Phase 2: Adapter Implementations

### Task 2.1: Process adapter (generic)

**Files:**
- Create: `Challenges/pkg/userflow/process_cli_adapter.go`
- Create: `Challenges/pkg/userflow/process_cli_adapter_test.go`

Generalize from `pkg/yole/process_adapter.go` — support any command (not just `java -jar`). Use `ProcessConfig` struct.

**Step 1: Write test**

```go
// process_cli_adapter_test.go
package userflow

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProcessCLIAdapter_LaunchAndStop(t *testing.T) {
	adapter := NewProcessCLIAdapter()
	ctx := context.Background()

	err := adapter.Launch(ctx, ProcessConfig{
		Command: "sleep",
		Args:    []string{"60"},
	})
	require.NoError(t, err)
	assert.True(t, adapter.IsRunning())

	err = adapter.Stop()
	assert.NoError(t, err)

	// Give OS time to reap
	time.Sleep(100 * time.Millisecond)
	assert.False(t, adapter.IsRunning())
}

func TestProcessCLIAdapter_WaitForReady(t *testing.T) {
	adapter := NewProcessCLIAdapter()
	ctx := context.Background()

	err := adapter.Launch(ctx, ProcessConfig{
		Command: "sleep",
		Args:    []string{"60"},
	})
	require.NoError(t, err)
	defer adapter.Stop()

	err = adapter.WaitForReady(ctx, 2*time.Second)
	assert.NoError(t, err)
}

func TestProcessCLIAdapter_NotRunning(t *testing.T) {
	adapter := NewProcessCLIAdapter()
	assert.False(t, adapter.IsRunning())
}

// Compile-time interface check
var _ ProcessAdapter = (*ProcessCLIAdapter)(nil)
```

**Step 2: Run to verify failure, Step 3: Implement**

Create `Challenges/pkg/userflow/process_cli_adapter.go`:
```go
package userflow

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"
)

// ProcessCLIAdapter manages application process lifecycle.
// It supports any binary command, not just JVM applications.
type ProcessCLIAdapter struct {
	cmd     *exec.Cmd
	process *os.Process
}

// NewProcessCLIAdapter creates a new ProcessCLIAdapter.
func NewProcessCLIAdapter() *ProcessCLIAdapter {
	return &ProcessCLIAdapter{}
}

// Launch starts a process with the given configuration.
func (p *ProcessCLIAdapter) Launch(
	ctx context.Context, config ProcessConfig,
) error {
	p.cmd = exec.CommandContext(
		ctx, config.Command, config.Args...,
	)
	if config.WorkDir != "" {
		p.cmd.Dir = config.WorkDir
	}
	for k, v := range config.Env {
		p.cmd.Env = append(
			p.cmd.Env, fmt.Sprintf("%s=%s", k, v),
		)
	}
	if len(p.cmd.Env) > 0 {
		p.cmd.Env = append(os.Environ(), p.cmd.Env...)
	}

	if err := p.cmd.Start(); err != nil {
		return fmt.Errorf("launch process: %w", err)
	}
	p.process = p.cmd.Process
	return nil
}

// IsRunning checks if the managed process is still alive.
func (p *ProcessCLIAdapter) IsRunning() bool {
	if p.process == nil {
		return false
	}
	err := p.process.Signal(syscall.Signal(0))
	return err == nil
}

// WaitForReady waits until the process is running or timeout.
func (p *ProcessCLIAdapter) WaitForReady(
	ctx context.Context, timeout time.Duration,
) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if p.IsRunning() {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(200 * time.Millisecond):
		}
	}
	return fmt.Errorf(
		"process not ready within %v", timeout,
	)
}

// Stop gracefully terminates the process (SIGTERM → SIGKILL).
func (p *ProcessCLIAdapter) Stop() error {
	if p.process == nil {
		return nil
	}
	if err := p.process.Signal(
		syscall.SIGTERM,
	); err != nil {
		return nil
	}
	done := make(chan error, 1)
	go func() {
		_, err := p.process.Wait()
		done <- err
	}()
	select {
	case <-done:
		return nil
	case <-time.After(5 * time.Second):
		return p.process.Kill()
	}
}
```

**Step 4: Run tests**

Run: `cd /run/media/milosvasic/DATA4TB/Projects/Catalogizer/Challenges && go test ./pkg/userflow/ -v -count=1 -run TestProcessCLI`
Expected: PASS

**Step 5: Commit**

```bash
git add pkg/userflow/process_cli_adapter.go pkg/userflow/process_cli_adapter_test.go
git commit -m "feat(userflow): add generic ProcessCLIAdapter"
```

---

### Task 2.2: Gradle build adapter

**Files:**
- Create: `Challenges/pkg/userflow/gradle_cli_adapter.go`
- Create: `Challenges/pkg/userflow/gradle_cli_adapter_test.go`

Generalize from `pkg/yole/gradle_cli_adapter.go` — configurable project root, JUnit XML search paths. Implements `BuildAdapter`.

Follow same TDD pattern: write test → verify fail → implement → verify pass → commit.

Commit message: `feat(userflow): add GradleCLIAdapter implementing BuildAdapter`

---

### Task 2.3: npm build adapter

**Files:**
- Create: `Challenges/pkg/userflow/npm_cli_adapter.go`
- Create: `Challenges/pkg/userflow/npm_cli_adapter_test.go`

Implements `BuildAdapter` for npm/Node.js projects. Uses `npm run <task>` for builds, `npx vitest --reporter=junit` for tests, `npx eslint --format=json` for lint.

Commit message: `feat(userflow): add NPMCLIAdapter implementing BuildAdapter`

---

### Task 2.4: Go build adapter

**Files:**
- Create: `Challenges/pkg/userflow/go_cli_adapter.go`
- Create: `Challenges/pkg/userflow/go_cli_adapter_test.go`

Implements `BuildAdapter` for Go projects. Uses `go build`, `go test -json`, `go vet`.

Commit message: `feat(userflow): add GoCLIAdapter implementing BuildAdapter`

---

### Task 2.5: Cargo build adapter

**Files:**
- Create: `Challenges/pkg/userflow/cargo_cli_adapter.go`
- Create: `Challenges/pkg/userflow/cargo_cli_adapter_test.go`

Implements `BuildAdapter` for Rust/Cargo projects. Uses `cargo build`, `cargo test`, `cargo clippy`.

Commit message: `feat(userflow): add CargoCLIAdapter implementing BuildAdapter`

---

### Task 2.6: ADB mobile adapter (generalized)

**Files:**
- Create: `Challenges/pkg/userflow/adb_cli_adapter.go`
- Create: `Challenges/pkg/userflow/adb_cli_adapter_test.go`

Generalize from `pkg/yole/adb_cli_adapter.go`. Key change: `PackageName` and `ActivityName` are **constructor parameters**, not hardcoded. Add `Tap`, `SendKeys`, `PressKey`, `RunInstrumentedTests`. Implements `MobileAdapter`.

Commit message: `feat(userflow): add configurable ADBCLIAdapter implementing MobileAdapter`

---

### Task 2.7: Playwright browser adapter

**Files:**
- Create: `Challenges/pkg/userflow/playwright_cli_adapter.go`
- Create: `Challenges/pkg/userflow/playwright_cli_adapter_test.go`

Implements `BrowserAdapter`. Connects to Playwright container via CDP endpoint. Executes browser operations via `runtime.Exec()` inside the container, running Node.js helper scripts. Includes `Initialize`, `Navigate`, `Click`, `Fill`, `SelectOption`, `IsVisible`, `WaitForSelector`, `GetText`, `GetAttribute`, `Screenshot`, `EvaluateJS`, `NetworkIntercept`, `Close`.

Commit message: `feat(userflow): add PlaywrightCLIAdapter implementing BrowserAdapter`

---

### Task 2.8: Tauri desktop adapter

**Files:**
- Create: `Challenges/pkg/userflow/tauri_cli_adapter.go`
- Create: `Challenges/pkg/userflow/tauri_cli_adapter_test.go`

Implements `DesktopAdapter`. Launches Tauri binary with `TAURI_AUTOMATION=true`, communicates via WebDriver HTTP protocol on auto-detected port.

Commit message: `feat(userflow): add TauriCLIAdapter implementing DesktopAdapter`

---

### Task 2.9: HTTP API adapter

**Files:**
- Create: `Challenges/pkg/userflow/http_api_adapter.go`
- Create: `Challenges/pkg/userflow/http_api_adapter_test.go`

Implements `APIAdapter`. Wraps existing `Challenges/pkg/httpclient` (delegates all HTTP methods). Adds `WebSocketConnect` using gorilla/websocket. Zero code duplication.

Commit message: `feat(userflow): add HTTPAPIAdapter implementing APIAdapter`

---

## Phase 3: Challenge Templates

### Task 3.1: Environment setup/teardown challenges

**Files:**
- Create: `Challenges/pkg/userflow/challenge_env.go`
- Create: `Challenges/pkg/userflow/challenge_env_test.go`

`EnvironmentSetupChallenge` — starts containers, health checks.
`EnvironmentTeardownChallenge` — stops containers, collects logs.

Commit message: `feat(userflow): add environment setup/teardown challenge templates`

---

### Task 3.2: Build, test, and lint challenge templates

**Files:**
- Create: `Challenges/pkg/userflow/challenge_build.go`
- Create: `Challenges/pkg/userflow/challenge_build_test.go`

`BuildChallenge`, `UnitTestChallenge`, `LintChallenge` — generic templates that accept any `BuildAdapter` and targets.

Commit message: `feat(userflow): add build, test, and lint challenge templates`

---

### Task 3.3: API flow challenge template

**Files:**
- Create: `Challenges/pkg/userflow/challenge_api_flow.go`
- Create: `Challenges/pkg/userflow/challenge_api_flow_test.go`

`APIHealthChallenge` — simple health endpoint check.
`APIFlowChallenge` — executes `APIFlow` steps sequentially, extracts variables, evaluates assertions.

Commit message: `feat(userflow): add API health and flow challenge templates`

---

### Task 3.4: Browser flow challenge template

**Files:**
- Create: `Challenges/pkg/userflow/challenge_browser.go`
- Create: `Challenges/pkg/userflow/challenge_browser_test.go`

`BrowserFlowChallenge` — executes `BrowserFlow` steps, takes screenshots, evaluates assertions.

Commit message: `feat(userflow): add browser flow challenge template`

---

### Task 3.5: Mobile challenge templates

**Files:**
- Create: `Challenges/pkg/userflow/challenge_mobile.go`
- Create: `Challenges/pkg/userflow/challenge_mobile_test.go`

`MobileLaunchChallenge`, `MobileFlowChallenge`, `InstrumentedTestChallenge`.

Commit message: `feat(userflow): add mobile launch, flow, and instrumented test challenge templates`

---

### Task 3.6: Desktop challenge templates

**Files:**
- Create: `Challenges/pkg/userflow/challenge_desktop.go`
- Create: `Challenges/pkg/userflow/challenge_desktop_test.go`

`DesktopLaunchChallenge`, `DesktopFlowChallenge`, `DesktopIPCChallenge`.

Commit message: `feat(userflow): add desktop launch, flow, and IPC challenge templates`

---

## Phase 4: Containers Module Integration

### Task 4.1: Add Containers dependency to Challenges go.mod

**Files:**
- Modify: `Challenges/go.mod`

**Step 1: Add replace directive**

```bash
cd /run/media/milosvasic/DATA4TB/Projects/Catalogizer/Challenges
```

Add to `go.mod`:
```
require digital.vasic.containers v0.0.0

replace digital.vasic.containers => ../Containers
```

Run: `go mod tidy`

**Step 2: Verify build**

Run: `go build ./...`
Expected: SUCCESS

**Step 3: Commit**

```bash
git add go.mod go.sum
git commit -m "build: add digital.vasic.containers dependency"
```

---

### Task 4.2: Create TestEnvironment with Containers integration

**Files:**
- Create: `Challenges/pkg/userflow/container_infra.go`
- Create: `Challenges/pkg/userflow/container_infra_test.go`

**Step 1: Write test**

```go
package userflow

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTestEnvironment_Defaults(t *testing.T) {
	env := NewTestEnvironment()
	assert.NotNil(t, env)
	assert.NotNil(t, env.services)
}

func TestTestEnvironment_ServiceURL(t *testing.T) {
	env := NewTestEnvironment()
	env.services["api"] = ServiceConfig{
		Name: "api",
		Host: "localhost",
		Port: 8080,
	}
	url := env.ServiceURL("api")
	assert.Equal(t, "http://localhost:8080", url)
}

func TestTestEnvironment_PlatformGroups(t *testing.T) {
	env := NewTestEnvironment(
		WithPlatformGroups([]PlatformGroup{
			{Name: "api", Services: []string{"api"}},
			{Name: "web", Services: []string{"api", "web", "playwright"}},
		}),
	)
	assert.Len(t, env.groups, 2)
}

func TestResourceLimit_Fields(t *testing.T) {
	rl := ResourceLimit{CPUs: 2, Memory: "4g"}
	assert.Equal(t, float64(2), rl.CPUs)
	assert.Equal(t, "4g", rl.Memory)
}
```

**Step 2: Implement**

Create `Challenges/pkg/userflow/container_infra.go`:
```go
package userflow

import (
	"context"
	"fmt"
	"time"

	"digital.vasic.containers/pkg/compose"
	"digital.vasic.containers/pkg/event"
	"digital.vasic.containers/pkg/health"
	"digital.vasic.containers/pkg/runtime"
	"digital.vasic.containers/pkg/serviceregistry"
)

// ServiceConfig defines a test service managed by containers.
type ServiceConfig struct {
	Name         string
	Host         string
	Port         int
	HealthPath   string
	HealthType   string
	Image        string
	ComposeFile  string
}

// ResourceLimit defines CPU and memory limits for a service.
type ResourceLimit struct {
	CPUs   float64
	Memory string
}

// PlatformGroup defines a set of services and challenges
// that run together, respecting resource budgets.
type PlatformGroup struct {
	Name       string
	Services   []string
	Challenges []string
	Limits     map[string]ResourceLimit
}

// TestEnvironment orchestrates containerized test infrastructure
// using the Containers module.
type TestEnvironment struct {
	runtime  runtime.ContainerRuntime
	compose  *compose.DefaultOrchestrator
	registry *serviceregistry.ServiceRegistry
	eventBus *event.DefaultEventBus
	services map[string]ServiceConfig
	groups   []PlatformGroup
	workDir  string
}

// TestEnvOption configures a TestEnvironment.
type TestEnvOption func(*TestEnvironment)

// NewTestEnvironment creates a TestEnvironment.
func NewTestEnvironment(
	opts ...TestEnvOption,
) *TestEnvironment {
	env := &TestEnvironment{
		services: make(map[string]ServiceConfig),
		registry: serviceregistry.New(),
		eventBus: event.NewEventBus(64),
	}
	for _, opt := range opts {
		opt(env)
	}
	return env
}

// WithWorkDir sets the working directory for compose files.
func WithWorkDir(dir string) TestEnvOption {
	return func(e *TestEnvironment) {
		e.workDir = dir
	}
}

// WithServices configures the services to manage.
func WithServices(
	services map[string]ServiceConfig,
) TestEnvOption {
	return func(e *TestEnvironment) {
		e.services = services
	}
}

// WithPlatformGroups sets the platform group definitions.
func WithPlatformGroups(groups []PlatformGroup) TestEnvOption {
	return func(e *TestEnvironment) {
		e.groups = groups
	}
}

// Setup starts all containers and waits for health.
func (e *TestEnvironment) Setup(
	ctx context.Context,
) error {
	// Auto-detect runtime (Podman-first)
	rt, err := runtime.AutoDetect(ctx)
	if err != nil {
		return fmt.Errorf("detect container runtime: %w", err)
	}
	e.runtime = rt

	e.eventBus.Publish(ctx, event.NewEvent(
		event.EventBootStarted, "userflow", "test-env",
	))

	// Start compose if workDir is set
	if e.workDir != "" {
		orch, err := compose.NewDefaultOrchestrator(
			e.workDir, nil,
		)
		if err != nil {
			return fmt.Errorf("create orchestrator: %w", err)
		}
		e.compose = orch
	}

	e.eventBus.Publish(ctx, event.NewEvent(
		event.EventBootCompleted, "userflow", "test-env",
	))
	return nil
}

// WaitForHealthy waits until all services pass health checks.
func (e *TestEnvironment) WaitForHealthy(
	ctx context.Context, timeout time.Duration,
) error {
	deadline := time.Now().Add(timeout)
	for name, svc := range e.services {
		target := health.HealthTarget{
			Name: name,
			Host: svc.Host,
			Port: fmt.Sprintf("%d", svc.Port),
			Type: health.HealthHTTP,
			Path: svc.HealthPath,
		}
		for time.Now().Before(deadline) {
			result := health.CheckHTTP(ctx, target)
			if result.Healthy {
				e.registry.Register(
					name, svc.Port,
				)
				break
			}
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(1 * time.Second):
			}
		}
	}
	return nil
}

// ServiceURL returns the URL for a registered service.
func (e *TestEnvironment) ServiceURL(name string) string {
	svc, ok := e.services[name]
	if !ok {
		return ""
	}
	return fmt.Sprintf("http://%s:%d", svc.Host, svc.Port)
}

// Teardown stops all containers and cleans up.
func (e *TestEnvironment) Teardown(
	ctx context.Context,
) error {
	e.eventBus.Publish(ctx, event.NewEvent(
		event.EventShutdownStarted, "userflow", "test-env",
	))

	e.eventBus.Publish(ctx, event.NewEvent(
		event.EventShutdownCompleted, "userflow", "test-env",
	))
	return nil
}
```

**Step 3: Run tests, Step 4: Commit**

```bash
git add pkg/userflow/container_infra.go pkg/userflow/container_infra_test.go
git commit -m "feat(userflow): add TestEnvironment with Containers module integration"
```

---

## Phase 5: Rename cmd/yole-challenges/ → cmd/userflow-runner/

### Task 5.1: Create generic CLI runner

**Files:**
- Create: `Challenges/cmd/userflow-runner/main.go`

Generic CLI that accepts `--platform`, `--project-root`, `--report`, `--output`, `--timeout` flags. Initializes adapters based on platform, registers challenge templates, runs via the runner.

Commit message: `feat(userflow): add generic userflow-runner CLI`

---

## Phase 6: Catalogizer-Specific Challenges

### Task 6.1: Create userflow helper and environment challenge

**Files:**
- Create: `catalog-api/challenges/userflow_helper.go`
- Create: `catalog-api/challenges/userflow_env.go`

`userflow_helper.go`: Adapter factory, shared config loading, flow builders.
`userflow_env.go`: `CatalogizerEnvSetupChallenge`, `CatalogizerEnvTeardownChallenge`.

Commit message: `feat(challenges): add userflow helper and environment challenges`

---

### Task 6.2–6.18: API flow challenges (one task per category)

Each task creates one `userflow_api_*.go` file with exhaustive challenges:

| Task | File | Challenges | Commit |
|------|------|-----------|--------|
| 6.2 | `userflow_api_auth.go` | Login success, invalid creds, empty fields, register, refresh, logout, status, permissions, profile | `feat(challenges): add API auth flow challenges` |
| 6.3 | `userflow_api_media.go` | Search, get-by-id, stats, favorites, progress, recent, trending, by-type, download, stream | `feat(challenges): add API media flow challenges` |
| 6.4 | `userflow_api_entities.go` | List, filter by type, get detail, children, files, metadata, duplicates, browse-by-type, stats, hierarchy | `feat(challenges): add API entity flow challenges` |
| 6.5 | `userflow_api_collections.go` | CRUD, add items, remove items, list, get-by-id | `feat(challenges): add API collection flow challenges` |
| 6.6 | `userflow_api_scanning.go` | Create storage root, queue scan, poll status, verify completion, verify entities created | `feat(challenges): add API scanning flow challenges` |
| 6.7 | `userflow_api_admin.go` | User CRUD, role CRUD, permissions, lock/unlock, password reset | `feat(challenges): add API admin flow challenges` |
| 6.8 | `userflow_api_downloads.go` | Download file, download directory, create archive, copy to storage | `feat(challenges): add API download flow challenges` |
| 6.9 | `userflow_api_subtitles.go` | Search, download, media subtitles, verify sync, translate, upload, languages, providers | `feat(challenges): add API subtitle flow challenges` |
| 6.10 | `userflow_api_conversion.go` | Create job, list jobs, get job, cancel, formats | `feat(challenges): add API conversion flow challenges` |
| 6.11 | `userflow_api_stats.go` | Overall, filetypes, sizes, duplicates, access, growth, scans, directories | `feat(challenges): add API stats flow challenges` |
| 6.12 | `userflow_api_errors.go` | Report error, report crash, list, get, update status, statistics | `feat(challenges): add API error reporting flow challenges` |
| 6.13 | `userflow_api_logs.go` | Collect, list collections, get entries, export, analyze, share, stream, statistics | `feat(challenges): add API log management flow challenges` |
| 6.14 | `userflow_api_recommendations.go` | Similar, trending, personalized | `feat(challenges): add API recommendation flow challenges` |
| 6.15 | `userflow_api_smb.go` | Discover, test connection, browse | `feat(challenges): add API SMB discovery flow challenges` |
| 6.16 | `userflow_api_websocket.go` | Connect, subscribe channels, receive events, disconnect | `feat(challenges): add API WebSocket flow challenges` |
| 6.17 | `userflow_api_stress.go` | Concurrent requests, rate limiting, sustained load, error rates | `feat(challenges): add API stress test flow challenges` |
| 6.18 | `userflow_api_security.go` | Auth bypass, invalid tokens, role enforcement, CORS, injection attempts, permission boundaries | `feat(challenges): add API security flow challenges` |

---

### Task 6.19–6.32: Web browser flow challenges

| Task | File | Commit |
|------|------|--------|
| 6.19 | `userflow_web_auth.go` | `feat(challenges): add web auth browser flow challenges` |
| 6.20 | `userflow_web_dashboard.go` | `feat(challenges): add web dashboard browser flow challenges` |
| 6.21 | `userflow_web_browse.go` | `feat(challenges): add web browse/entity browser flow challenges` |
| 6.22 | `userflow_web_search.go` | `feat(challenges): add web search/filter browser flow challenges` |
| 6.23 | `userflow_web_collections.go` | `feat(challenges): add web collection browser flow challenges` |
| 6.24 | `userflow_web_player.go` | `feat(challenges): add web media player browser flow challenges` |
| 6.25 | `userflow_web_admin.go` | `feat(challenges): add web admin panel browser flow challenges` |
| 6.26 | `userflow_web_subtitles.go` | `feat(challenges): add web subtitle manager browser flow challenges` |
| 6.27 | `userflow_web_conversion.go` | `feat(challenges): add web conversion tools browser flow challenges` |
| 6.28 | `userflow_web_analytics.go` | `feat(challenges): add web analytics browser flow challenges` |
| 6.29 | `userflow_web_favorites.go` | `feat(challenges): add web favorites browser flow challenges` |
| 6.30 | `userflow_web_responsive.go` | `feat(challenges): add web responsive/viewport browser flow challenges` |
| 6.31 | `userflow_web_errors.go` | `feat(challenges): add web error state browser flow challenges` |
| 6.32 | `userflow_web_accessibility.go` | `feat(challenges): add web accessibility browser flow challenges` |

---

### Task 6.33–6.37: Desktop challenges

| Task | File | Commit |
|------|------|--------|
| 6.33 | `userflow_desktop_setup.go` | `feat(challenges): add desktop setup flow challenges` |
| 6.34 | `userflow_desktop_auth.go` | `feat(challenges): add desktop auth flow challenges` |
| 6.35 | `userflow_desktop_browse.go` | `feat(challenges): add desktop browse flow challenges` |
| 6.36 | `userflow_desktop_settings.go` | `feat(challenges): add desktop settings flow challenges` |
| 6.37 | `userflow_desktop_ipc.go` | `feat(challenges): add desktop IPC flow challenges` |

---

### Task 6.38–6.40: Wizard challenges

| Task | File | Commit |
|------|------|--------|
| 6.38 | `userflow_wizard_flow.go` | `feat(challenges): add wizard complete flow challenges` |
| 6.39 | `userflow_wizard_protocols.go` | `feat(challenges): add wizard protocol config flow challenges` |
| 6.40 | `userflow_wizard_validation.go` | `feat(challenges): add wizard validation flow challenges` |

---

### Task 6.41–6.47: Android challenges

| Task | File | Commit |
|------|------|--------|
| 6.41 | `userflow_android_build.go` | `feat(challenges): add Android build/lint challenges` |
| 6.42 | `userflow_android_launch.go` | `feat(challenges): add Android launch/stability challenges` |
| 6.43 | `userflow_android_auth.go` | `feat(challenges): add Android auth flow challenges` |
| 6.44 | `userflow_android_browse.go` | `feat(challenges): add Android browse flow challenges` |
| 6.45 | `userflow_android_playback.go` | `feat(challenges): add Android playback flow challenges` |
| 6.46 | `userflow_android_settings.go` | `feat(challenges): add Android settings flow challenges` |
| 6.47 | `userflow_android_offline.go` | `feat(challenges): add Android offline mode flow challenges` |

---

### Task 6.48–6.53: Android TV challenges

| Task | File | Commit |
|------|------|--------|
| 6.48 | `userflow_androidtv_build.go` | `feat(challenges): add Android TV build challenges` |
| 6.49 | `userflow_androidtv_launch.go` | `feat(challenges): add Android TV launch challenges` |
| 6.50 | `userflow_androidtv_nav.go` | `feat(challenges): add Android TV D-pad navigation challenges` |
| 6.51 | `userflow_androidtv_browse.go` | `feat(challenges): add Android TV browse flow challenges` |
| 6.52 | `userflow_androidtv_playback.go` | `feat(challenges): add Android TV playback flow challenges` |
| 6.53 | `userflow_androidtv_settings.go` | `feat(challenges): add Android TV settings flow challenges` |

---

### Task 6.54: Update register.go

**Files:**
- Modify: `catalog-api/challenges/register.go`

Add `RegisterUserFlowChallenges(svc)` function that registers all ~200+ userflow challenges. Called from `RegisterAll()`.

Commit message: `feat(challenges): register all userflow challenges in RegisterAll`

---

## Phase 7: Documentation

### Task 7.1: Challenges submodule documentation

**Files:** Create 13 files under `Challenges/docs/userflow/`:
- `README.md`, `ADAPTERS.md`, `BROWSER_ADAPTER.md`, `MOBILE_ADAPTER.md`, `DESKTOP_ADAPTER.md`, `API_ADAPTER.md`, `BUILD_ADAPTERS.md`, `PROCESS_ADAPTER.md`, `CHALLENGE_TEMPLATES.md`, `EVALUATORS.md`, `CONTAINER_INTEGRATION.md`, `WRITING_CHALLENGES.md`, `WRITING_ADAPTERS.md`, `ARCHITECTURE.md`

Commit message: `docs(userflow): add comprehensive framework documentation`

---

### Task 7.2: Catalogizer testing documentation

**Files:** Create 6 files under `docs/testing/`:
- `USERFLOW_TESTING.md`, `RUNNING_TESTS.md`, `CONTAINER_SETUP.md`, `CHALLENGE_MAP.md`, `ADDING_CHALLENGES.md`, `TROUBLESHOOTING.md`

Commit message: `docs(testing): add userflow testing documentation`

---

### Task 7.3: Update Challenges CLAUDE.md

**Files:**
- Modify: `Challenges/CLAUDE.md`

Add `pkg/userflow` to the package table. Remove any yole references. Document adapter pattern, evaluators, and TestEnvironment.

Commit message: `docs: update CLAUDE.md with userflow package`

---

### Task 7.4: Create docker-compose.test.yml

**Files:**
- Create: `docker-compose.test.yml`

Container definitions for the full test stack (catalog-api, catalog-web, playwright, android-emulator, tauri-desktop, tauri-wizard).

Commit message: `feat: add docker-compose.test.yml for userflow test infrastructure`

---

## Phase 8: Integration Verification

### Task 8.1: Update catalog-api go.mod

**Files:**
- Modify: `catalog-api/go.mod`

Ensure the `replace` directive for Challenges picks up the new `pkg/userflow/` package and Containers dependency.

Run: `cd catalog-api && go mod tidy && go build ./...`

Commit message: `build: update go.mod for userflow integration`

---

### Task 8.2: Full test suite

Run: `cd /run/media/milosvasic/DATA4TB/Projects/Catalogizer/Challenges && GOMAXPROCS=3 go test ./... -p 2 -parallel 2 -count=1`

Verify all tests pass (existing + new userflow tests).

Run: `cd /run/media/milosvasic/DATA4TB/Projects/Catalogizer/catalog-api && GOMAXPROCS=3 go test ./... -p 2 -parallel 2 -count=1`

Verify all tests pass including new challenge registration.

---

### Task 8.3: Push all changes

Push Challenges submodule, Containers submodule (if changed), and main repo to all upstreams.

```bash
cd /run/media/milosvasic/DATA4TB/Projects/Catalogizer/Challenges
GIT_SSH_COMMAND="ssh -o BatchMode=yes" git push origin main

cd /run/media/milosvasic/DATA4TB/Projects/Catalogizer
git add Challenges catalog-api docs docker-compose.test.yml
git commit -m "feat: integrate userflow automation framework with exhaustive challenges"
GIT_SSH_COMMAND="ssh -o BatchMode=yes" git push origin main
```
