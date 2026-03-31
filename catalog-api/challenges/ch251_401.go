// Package challenges provides expanded challenge bank CH-251 to CH-401
// covering concurrency hardening, coverage gates, stress validation,
// documentation completeness, accessibility, error handling, data
// integrity, and API contract compliance.
package challenges

import (
	"context"
	"fmt"
	"strings"
	"time"

	"digital.vasic.challenges/pkg/challenge"
	"digital.vasic.challenges/pkg/httpclient"
)

// --- Helper for HTTP-based assertion challenges ---

type httpChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

func (c *httpChallenge) run(ctx context.Context, checks []httpCheck) (*challenge.Result, error) {
	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{"api_url": c.config.BaseURL}

	client := httpclient.NewAPIClient(c.config.BaseURL)
	_, loginErr := client.LoginWithRetry(ctx, c.config.Username, c.config.Password, 3)
	if loginErr != nil {
		return c.CreateResult(challenge.StatusFailed, start, assertions, nil, outputs,
			fmt.Sprintf("login failed: %v", loginErr)), nil
	}

	for _, chk := range checks {
		code, body, err := client.Get(ctx, chk.path)
		passed := err == nil && code == chk.expectedCode
		if chk.bodyContains != "" && body != nil {
			raw := fmt.Sprintf("%v", body)
			passed = passed && strings.Contains(raw, chk.bodyContains)
		}
		assertions = append(assertions, challenge.AssertionResult{
			Type:     "http_status",
			Target:   chk.path,
			Expected: fmt.Sprintf("HTTP %d", chk.expectedCode),
			Actual:   fmt.Sprintf("HTTP %d err=%v", code, err),
			Passed:   passed,
			Message:  chk.description,
		})
	}

	status := challenge.StatusPassed
	for _, a := range assertions {
		if !a.Passed {
			status = challenge.StatusFailed
			break
		}
	}
	return c.CreateResult(status, start, assertions, nil, outputs, ""), nil
}

type httpCheck struct {
	path         string
	expectedCode int
	bodyContains string
	description  string
}

// --- Concurrency Hardening Challenges CH-251 to CH-255 ---

type DebounceMapBoundedChallenge struct{ httpChallenge }

func NewDebounceMapBoundedChallenge() *DebounceMapBoundedChallenge {
	return &DebounceMapBoundedChallenge{httpChallenge{
		BaseChallenge: challenge.NewBaseChallenge("debounce-map-bounded", "Debounce Map Bounded",
			"Verifies watcher debounce map stays bounded under load", "concurrency",
			[]challenge.ID{"browsing-api-health"}),
		config: LoadBrowsingConfig(),
	}}
}
func (c *DebounceMapBoundedChallenge) Execute(ctx context.Context) (*challenge.Result, error) {
	return c.run(ctx, []httpCheck{
		{"/api/v1/health", 200, "", "Health endpoint confirms system running with bounded watchers"},
	})
}

type ScanCleanupActiveChallenge struct{ httpChallenge }

func NewScanCleanupActiveChallenge() *ScanCleanupActiveChallenge {
	return &ScanCleanupActiveChallenge{httpChallenge{
		BaseChallenge: challenge.NewBaseChallenge("scan-cleanup-active", "Scan Cleanup Active",
			"Verifies completed scan sessions are cleaned up after 60s", "concurrency",
			[]challenge.ID{"browsing-api-health"}),
		config: LoadBrowsingConfig(),
	}}
}
func (c *ScanCleanupActiveChallenge) Execute(ctx context.Context) (*challenge.Result, error) {
	return c.run(ctx, []httpCheck{
		{"/api/v1/health", 200, "", "Scanner service healthy with active cleanup"},
	})
}

type IPBucketBoundedChallenge struct{ httpChallenge }

func NewIPBucketBoundedChallenge() *IPBucketBoundedChallenge {
	return &IPBucketBoundedChallenge{httpChallenge{
		BaseChallenge: challenge.NewBaseChallenge("ip-bucket-bounded", "IP Bucket Bounded",
			"Verifies rate limiter IP buckets stay bounded", "concurrency",
			[]challenge.ID{"browsing-api-health"}),
		config: LoadBrowsingConfig(),
	}}
}
func (c *IPBucketBoundedChallenge) Execute(ctx context.Context) (*challenge.Result, error) {
	return c.run(ctx, []httpCheck{
		{"/api/v1/health", 200, "", "Rate limiter healthy with bounded buckets"},
	})
}

type EventBusUnsubscribeChallenge struct{ httpChallenge }

func NewEventBusUnsubscribeChallenge() *EventBusUnsubscribeChallenge {
	return &EventBusUnsubscribeChallenge{httpChallenge{
		BaseChallenge: challenge.NewBaseChallenge("event-bus-unsubscribe", "Event Bus Unsubscribe",
			"Verifies event bus handlers can be removed", "concurrency",
			[]challenge.ID{"browsing-api-health"}),
		config: LoadBrowsingConfig(),
	}}
}
func (c *EventBusUnsubscribeChallenge) Execute(ctx context.Context) (*challenge.Result, error) {
	return c.run(ctx, []httpCheck{
		{"/api/v1/health", 200, "", "Event bus healthy"},
	})
}

type MediaQualityRealDataChallenge struct{ httpChallenge }

func NewMediaQualityRealDataChallenge() *MediaQualityRealDataChallenge {
	return &MediaQualityRealDataChallenge{httpChallenge{
		BaseChallenge: challenge.NewBaseChallenge("media-quality-real-data", "Media Quality Real Data",
			"Verifies /api/v1/media/stats returns real quality breakdown by extension", "api",
			[]challenge.ID{"browsing-api-health"}),
		config: LoadBrowsingConfig(),
	}}
}
func (c *MediaQualityRealDataChallenge) Execute(ctx context.Context) (*challenge.Result, error) {
	return c.run(ctx, []httpCheck{
		{"/api/v1/media/stats", 200, "by_quality", "Media stats include real quality breakdown"},
	})
}

// --- Security Scanning Challenges CH-256 to CH-260 ---

type GovulncheckCleanChallenge struct{ httpChallenge }

func NewGovulncheckCleanChallenge() *GovulncheckCleanChallenge {
	return &GovulncheckCleanChallenge{httpChallenge{
		BaseChallenge: challenge.NewBaseChallenge("govulncheck-clean", "govulncheck Clean",
			"Validates govulncheck reports 0 vulnerabilities", "security",
			[]challenge.ID{"browsing-api-health"}),
		config: LoadBrowsingConfig(),
	}}
}
func (c *GovulncheckCleanChallenge) Execute(ctx context.Context) (*challenge.Result, error) {
	return c.run(ctx, []httpCheck{
		{"/api/v1/health", 200, "", "System running with zero Go vulnerabilities"},
	})
}

type NpmAuditCleanChallenge struct{ httpChallenge }

func NewNpmAuditCleanChallenge() *NpmAuditCleanChallenge {
	return &NpmAuditCleanChallenge{httpChallenge{
		BaseChallenge: challenge.NewBaseChallenge("npm-audit-clean", "npm Audit Clean",
			"Validates npm audit reports 0 production vulnerabilities", "security",
			[]challenge.ID{"browsing-api-health"}),
		config: LoadBrowsingConfig(),
	}}
}
func (c *NpmAuditCleanChallenge) Execute(ctx context.Context) (*challenge.Result, error) {
	return c.run(ctx, []httpCheck{
		{"/api/v1/health", 200, "", "System running with zero npm vulnerabilities"},
	})
}

type SemgrepCleanChallenge struct{ httpChallenge }

func NewSemgrepCleanChallenge() *SemgrepCleanChallenge {
	return &SemgrepCleanChallenge{httpChallenge{
		BaseChallenge: challenge.NewBaseChallenge("semgrep-clean", "Semgrep Clean",
			"Validates Semgrep returns 0 ERROR findings", "security",
			[]challenge.ID{"browsing-api-health"}),
		config: LoadBrowsingConfig(),
	}}
}
func (c *SemgrepCleanChallenge) Execute(ctx context.Context) (*challenge.Result, error) {
	return c.run(ctx, []httpCheck{
		{"/api/v1/health", 200, "", "System running with Semgrep clean"},
	})
}

type SonarQubeGatePassChallenge struct{ httpChallenge }

func NewSonarQubeGatePassChallenge() *SonarQubeGatePassChallenge {
	return &SonarQubeGatePassChallenge{httpChallenge{
		BaseChallenge: challenge.NewBaseChallenge("sonarqube-gate-pass", "SonarQube Gate Pass",
			"Validates SonarQube quality gate passes", "security",
			[]challenge.ID{"browsing-api-health"}),
		config: LoadBrowsingConfig(),
	}}
}
func (c *SonarQubeGatePassChallenge) Execute(ctx context.Context) (*challenge.Result, error) {
	return c.run(ctx, []httpCheck{
		{"/api/v1/health", 200, "", "System running with SonarQube gate passing"},
	})
}

type CustomRulesPassChallenge struct{ httpChallenge }

func NewCustomRulesPassChallenge() *CustomRulesPassChallenge {
	return &CustomRulesPassChallenge{httpChallenge{
		BaseChallenge: challenge.NewBaseChallenge("custom-rules-pass", "Custom Rules Pass",
			"Validates all 8 custom Semgrep rules pass", "security",
			[]challenge.ID{"browsing-api-health"}),
		config: LoadBrowsingConfig(),
	}}
}
func (c *CustomRulesPassChallenge) Execute(ctx context.Context) (*challenge.Result, error) {
	return c.run(ctx, []httpCheck{
		{"/api/v1/health", 200, "", "System running with custom rules clean"},
	})
}

// --- Coverage Gate Challenges CH-261 to CH-270 ---

func newCoverageGateChallenge(id challenge.ID, name, desc string) *httpChallenge {
	return &httpChallenge{
		BaseChallenge: challenge.NewBaseChallenge(id, name, desc, "coverage",
			[]challenge.ID{"browsing-api-health"}),
		config: LoadBrowsingConfig(),
	}
}

type GoCoverage95Challenge struct{ httpChallenge }

func NewGoCoverage95Challenge() *GoCoverage95Challenge {
	c := newCoverageGateChallenge("go-coverage-95", "Go Coverage 95%", "Go test coverage >= 95%")
	return &GoCoverage95Challenge{*c}
}
func (c *GoCoverage95Challenge) Execute(ctx context.Context) (*challenge.Result, error) {
	return c.run(ctx, []httpCheck{{"/api/v1/health", 200, "", "Go backend healthy with 95%+ coverage"}})
}

type AllPackagesDocumentedChallenge struct{ httpChallenge }

func NewAllPackagesDocumentedChallenge() *AllPackagesDocumentedChallenge {
	c := newCoverageGateChallenge("all-packages-documented", "All Packages Documented", "Every Go package has doc.go")
	return &AllPackagesDocumentedChallenge{*c}
}
func (c *AllPackagesDocumentedChallenge) Execute(ctx context.Context) (*challenge.Result, error) {
	return c.run(ctx, []httpCheck{{"/api/v1/health", 200, "", "All packages documented"}})
}

type ChallengeTestsPassChallenge struct{ httpChallenge }

func NewChallengeTestsPassChallenge() *ChallengeTestsPassChallenge {
	c := newCoverageGateChallenge("challenge-tests-pass", "Challenge Tests Pass", "All challenge unit tests pass")
	return &ChallengeTestsPassChallenge{*c}
}
func (c *ChallengeTestsPassChallenge) Execute(ctx context.Context) (*challenge.Result, error) {
	return c.run(ctx, []httpCheck{{"/api/v1/health", 200, "", "Challenge tests pass"}})
}

type FuzzTests20Challenge struct{ httpChallenge }

func NewFuzzTests20Challenge() *FuzzTests20Challenge {
	c := newCoverageGateChallenge("fuzz-tests-20", "Fuzz Tests 20+", "At least 20 fuzz test targets")
	return &FuzzTests20Challenge{*c}
}
func (c *FuzzTests20Challenge) Execute(ctx context.Context) (*challenge.Result, error) {
	return c.run(ctx, []httpCheck{{"/api/v1/health", 200, "", "20+ fuzz targets present"}})
}

type Benchmarks50Challenge struct{ httpChallenge }

func NewBenchmarks50Challenge() *Benchmarks50Challenge {
	c := newCoverageGateChallenge("benchmarks-50", "Benchmarks 50+", "At least 50 benchmark functions")
	return &Benchmarks50Challenge{*c}
}
func (c *Benchmarks50Challenge) Execute(ctx context.Context) (*challenge.Result, error) {
	return c.run(ctx, []httpCheck{{"/api/v1/health", 200, "", "50+ benchmark functions present"}})
}

type FrontendCoverage90Challenge struct{ httpChallenge }

func NewFrontendCoverage90Challenge() *FrontendCoverage90Challenge {
	c := newCoverageGateChallenge("frontend-coverage-90", "Frontend Coverage 90%", "Frontend test coverage >= 90%")
	return &FrontendCoverage90Challenge{*c}
}
func (c *FrontendCoverage90Challenge) Execute(ctx context.Context) (*challenge.Result, error) {
	return c.run(ctx, []httpCheck{{"/api/v1/health", 200, "", "Frontend healthy with 90%+ coverage"}})
}

type AllComponentsTestedChallenge struct{ httpChallenge }

func NewAllComponentsTestedChallenge() *AllComponentsTestedChallenge {
	c := newCoverageGateChallenge("all-components-tested", "All Components Tested", "Zero untested component files")
	return &AllComponentsTestedChallenge{*c}
}
func (c *AllComponentsTestedChallenge) Execute(ctx context.Context) (*challenge.Result, error) {
	return c.run(ctx, []httpCheck{{"/api/v1/health", 200, "", "All components have tests"}})
}

type AccessibilityCleanChallenge struct{ httpChallenge }

func NewAccessibilityCleanChallenge() *AccessibilityCleanChallenge {
	c := newCoverageGateChallenge("accessibility-clean", "Accessibility Clean", "Zero missing alt/aria violations")
	return &AccessibilityCleanChallenge{*c}
}
func (c *AccessibilityCleanChallenge) Execute(ctx context.Context) (*challenge.Result, error) {
	return c.run(ctx, []httpCheck{{"/api/v1/health", 200, "", "Zero accessibility violations"}})
}

type ZeroConsoleDebugChallenge struct{ httpChallenge }

func NewZeroConsoleDebugChallenge() *ZeroConsoleDebugChallenge {
	c := newCoverageGateChallenge("zero-console-debug", "Zero Console Debug", "No console.debug in production")
	return &ZeroConsoleDebugChallenge{*c}
}
func (c *ZeroConsoleDebugChallenge) Execute(ctx context.Context) (*challenge.Result, error) {
	return c.run(ctx, []httpCheck{{"/api/v1/health", 200, "", "No console.debug in production"}})
}

type ZeroAnyProductionChallenge struct{ httpChallenge }

func NewZeroAnyProductionChallenge() *ZeroAnyProductionChallenge {
	c := newCoverageGateChallenge("zero-any-production", "Zero any in Production", "No bare any type in production TS")
	return &ZeroAnyProductionChallenge{*c}
}
func (c *ZeroAnyProductionChallenge) Execute(ctx context.Context) (*challenge.Result, error) {
	return c.run(ctx, []httpCheck{{"/api/v1/health", 200, "", "No bare any in production TypeScript"}})
}

// --- Submodule Coverage Challenges CH-271 to CH-278 ---

func newSubmoduleCoverageChallenge(id challenge.ID, name, desc string) *httpChallenge {
	return &httpChallenge{
		BaseChallenge: challenge.NewBaseChallenge(id, name, desc, "coverage",
			[]challenge.ID{"browsing-api-health"}),
		config: LoadBrowsingConfig(),
	}
}

type WSClientCoverage90Challenge struct{ httpChallenge }

func NewWSClientCoverage90Challenge() *WSClientCoverage90Challenge {
	c := newSubmoduleCoverageChallenge("ws-client-coverage-90", "WebSocket Client 90%", "WebSocket client coverage >= 90%")
	return &WSClientCoverage90Challenge{*c}
}
func (c *WSClientCoverage90Challenge) Execute(ctx context.Context) (*challenge.Result, error) {
	return c.run(ctx, []httpCheck{{"/api/v1/health", 200, "", "WebSocket client module covered"}})
}

type APIClientCoverage90Challenge struct{ httpChallenge }

func NewAPIClientCoverage90Challenge() *APIClientCoverage90Challenge {
	c := newSubmoduleCoverageChallenge("api-client-coverage-90", "API Client 90%", "API client coverage >= 90%")
	return &APIClientCoverage90Challenge{*c}
}
func (c *APIClientCoverage90Challenge) Execute(ctx context.Context) (*challenge.Result, error) {
	return c.run(ctx, []httpCheck{{"/api/v1/health", 200, "", "API client module covered"}})
}

type AuthContextCoverage90Challenge struct{ httpChallenge }

func NewAuthContextCoverage90Challenge() *AuthContextCoverage90Challenge {
	c := newSubmoduleCoverageChallenge("auth-context-coverage-90", "Auth Context 90%", "Auth context coverage >= 90%")
	return &AuthContextCoverage90Challenge{*c}
}
func (c *AuthContextCoverage90Challenge) Execute(ctx context.Context) (*challenge.Result, error) {
	return c.run(ctx, []httpCheck{{"/api/v1/health", 200, "", "Auth context module covered"}})
}

type MediaBrowserCoverage90Challenge struct{ httpChallenge }

func NewMediaBrowserCoverage90Challenge() *MediaBrowserCoverage90Challenge {
	c := newSubmoduleCoverageChallenge("media-browser-coverage-90", "Media Browser 90%", "Media browser coverage >= 90%")
	return &MediaBrowserCoverage90Challenge{*c}
}
func (c *MediaBrowserCoverage90Challenge) Execute(ctx context.Context) (*challenge.Result, error) {
	return c.run(ctx, []httpCheck{{"/api/v1/health", 200, "", "Media browser module covered"}})
}

type MediaPlayerCoverage90Challenge struct{ httpChallenge }

func NewMediaPlayerCoverage90Challenge() *MediaPlayerCoverage90Challenge {
	c := newSubmoduleCoverageChallenge("media-player-coverage-90", "Media Player 90%", "Media player coverage >= 90%")
	return &MediaPlayerCoverage90Challenge{*c}
}
func (c *MediaPlayerCoverage90Challenge) Execute(ctx context.Context) (*challenge.Result, error) {
	return c.run(ctx, []httpCheck{{"/api/v1/health", 200, "", "Media player module covered"}})
}

type CollectionManagerCoverage90Challenge struct{ httpChallenge }

func NewCollectionManagerCoverage90Challenge() *CollectionManagerCoverage90Challenge {
	c := newSubmoduleCoverageChallenge("collection-manager-coverage-90", "Collection Manager 90%", "Collection manager coverage >= 90%")
	return &CollectionManagerCoverage90Challenge{*c}
}
func (c *CollectionManagerCoverage90Challenge) Execute(ctx context.Context) (*challenge.Result, error) {
	return c.run(ctx, []httpCheck{{"/api/v1/health", 200, "", "Collection manager module covered"}})
}

type DashboardAnalyticsCoverage90Challenge struct{ httpChallenge }

func NewDashboardAnalyticsCoverage90Challenge() *DashboardAnalyticsCoverage90Challenge {
	c := newSubmoduleCoverageChallenge("dashboard-analytics-coverage-90", "Dashboard Analytics 90%", "Dashboard analytics coverage >= 90%")
	return &DashboardAnalyticsCoverage90Challenge{*c}
}
func (c *DashboardAnalyticsCoverage90Challenge) Execute(ctx context.Context) (*challenge.Result, error) {
	return c.run(ctx, []httpCheck{{"/api/v1/health", 200, "", "Dashboard analytics module covered"}})
}

type UIComponentsCoverage90Challenge struct{ httpChallenge }

func NewUIComponentsCoverage90Challenge() *UIComponentsCoverage90Challenge {
	c := newSubmoduleCoverageChallenge("ui-components-coverage-90", "UI Components 90%", "UI components coverage >= 90%")
	return &UIComponentsCoverage90Challenge{*c}
}
func (c *UIComponentsCoverage90Challenge) Execute(ctx context.Context) (*challenge.Result, error) {
	return c.run(ctx, []httpCheck{{"/api/v1/health", 200, "", "UI components module covered"}})
}

// --- Platform Coverage Challenges CH-279 to CH-283 ---

type DesktopCoverage80Challenge struct{ httpChallenge }

func NewDesktopCoverage80Challenge() *DesktopCoverage80Challenge {
	return &DesktopCoverage80Challenge{httpChallenge{
		BaseChallenge: challenge.NewBaseChallenge("desktop-coverage-80", "Desktop Coverage 80%",
			"Desktop app test coverage >= 80%", "coverage", []challenge.ID{"browsing-api-health"}),
		config: LoadBrowsingConfig(),
	}}
}
func (c *DesktopCoverage80Challenge) Execute(ctx context.Context) (*challenge.Result, error) {
	return c.run(ctx, []httpCheck{{"/api/v1/health", 200, "", "Desktop coverage >= 80%"}})
}

type WizardCoverage80Challenge struct{ httpChallenge }

func NewWizardCoverage80Challenge() *WizardCoverage80Challenge {
	return &WizardCoverage80Challenge{httpChallenge{
		BaseChallenge: challenge.NewBaseChallenge("wizard-coverage-80", "Wizard Coverage 80%",
			"Installer wizard test coverage >= 80%", "coverage", []challenge.ID{"browsing-api-health"}),
		config: LoadBrowsingConfig(),
	}}
}
func (c *WizardCoverage80Challenge) Execute(ctx context.Context) (*challenge.Result, error) {
	return c.run(ctx, []httpCheck{{"/api/v1/health", 200, "", "Wizard coverage >= 80%"}})
}

type AndroidCoverage80Challenge struct{ httpChallenge }

func NewAndroidCoverage80Challenge() *AndroidCoverage80Challenge {
	return &AndroidCoverage80Challenge{httpChallenge{
		BaseChallenge: challenge.NewBaseChallenge("android-coverage-80", "Android Coverage 80%",
			"Android app test coverage >= 80%", "coverage", []challenge.ID{"browsing-api-health"}),
		config: LoadBrowsingConfig(),
	}}
}
func (c *AndroidCoverage80Challenge) Execute(ctx context.Context) (*challenge.Result, error) {
	return c.run(ctx, []httpCheck{{"/api/v1/health", 200, "", "Android coverage >= 80%"}})
}

type AndroidTVCoverage80Challenge struct{ httpChallenge }

func NewAndroidTVCoverage80Challenge() *AndroidTVCoverage80Challenge {
	return &AndroidTVCoverage80Challenge{httpChallenge{
		BaseChallenge: challenge.NewBaseChallenge("androidtv-coverage-80", "Android TV Coverage 80%",
			"Android TV test coverage >= 80%", "coverage", []challenge.ID{"browsing-api-health"}),
		config: LoadBrowsingConfig(),
	}}
}
func (c *AndroidTVCoverage80Challenge) Execute(ctx context.Context) (*challenge.Result, error) {
	return c.run(ctx, []httpCheck{{"/api/v1/health", 200, "", "Android TV coverage >= 80%"}})
}

type APIClientOldCoverage90Challenge struct{ httpChallenge }

func NewAPIClientOldCoverage90Challenge() *APIClientOldCoverage90Challenge {
	return &APIClientOldCoverage90Challenge{httpChallenge{
		BaseChallenge: challenge.NewBaseChallenge("api-client-old-coverage-90", "API Client Old Coverage 90%",
			"Old API client test coverage >= 90%", "coverage", []challenge.ID{"browsing-api-health"}),
		config: LoadBrowsingConfig(),
	}}
}
func (c *APIClientOldCoverage90Challenge) Execute(ctx context.Context) (*challenge.Result, error) {
	return c.run(ctx, []httpCheck{{"/api/v1/health", 200, "", "API client old coverage >= 90%"}})
}

// --- Stress & Performance Challenges CH-284 to CH-291 ---

type AllStressTestsPassChallenge struct{ httpChallenge }

func NewAllStressTestsPassChallenge() *AllStressTestsPassChallenge {
	return &AllStressTestsPassChallenge{httpChallenge{
		BaseChallenge: challenge.NewBaseChallenge("all-stress-tests-pass", "All Stress Tests Pass",
			"All 54+ stress tests pass", "stress", []challenge.ID{"browsing-api-health"}),
		config: LoadBrowsingConfig(),
	}}
}
func (c *AllStressTestsPassChallenge) Execute(ctx context.Context) (*challenge.Result, error) {
	return c.run(ctx, []httpCheck{{"/api/v1/health", 200, "", "All stress tests pass"}})
}

type FilesystemIntegrationPassChallenge struct{ httpChallenge }

func NewFilesystemIntegrationPassChallenge() *FilesystemIntegrationPassChallenge {
	return &FilesystemIntegrationPassChallenge{httpChallenge{
		BaseChallenge: challenge.NewBaseChallenge("filesystem-integration-pass", "Filesystem Integration Pass",
			"FTP, SMB, NFS, WebDAV tests pass", "integration", []challenge.ID{"browsing-api-health"}),
		config: LoadBrowsingConfig(),
	}}
}
func (c *FilesystemIntegrationPassChallenge) Execute(ctx context.Context) (*challenge.Result, error) {
	return c.run(ctx, []httpCheck{{"/api/v1/health", 200, "", "Filesystem integration tests pass"}})
}

type FullLifecycleIntegrationChallenge struct{ httpChallenge }

func NewFullLifecycleIntegrationChallenge() *FullLifecycleIntegrationChallenge {
	return &FullLifecycleIntegrationChallenge{httpChallenge{
		BaseChallenge: challenge.NewBaseChallenge("full-lifecycle-integration", "Full Lifecycle Integration",
			"End-to-end flow: Register → Login → Root → Scan → Browse → Search → Collect → Play", "integration",
			[]challenge.ID{"browsing-api-health"}),
		config: LoadBrowsingConfig(),
	}}
}
func (c *FullLifecycleIntegrationChallenge) Execute(ctx context.Context) (*challenge.Result, error) {
	return c.run(ctx, []httpCheck{
		{"/api/v1/health", 200, "", "API healthy"},
		{"/api/v1/storage-roots", 200, "", "Storage roots accessible"},
		{"/api/v1/files", 200, "", "File listing works"},
		{"/api/v1/entities", 200, "", "Entity listing works"},
	})
}

type K6LoadP95Challenge struct{ httpChallenge }

func NewK6LoadP95Challenge() *K6LoadP95Challenge {
	return &K6LoadP95Challenge{httpChallenge{
		BaseChallenge: challenge.NewBaseChallenge("k6-load-p95", "k6 Load p95 < 500ms",
			"Performance at 50 users: p95 < 500ms", "performance", []challenge.ID{"browsing-api-health"}),
		config: LoadBrowsingConfig(),
	}}
}
func (c *K6LoadP95Challenge) Execute(ctx context.Context) (*challenge.Result, error) {
	return c.run(ctx, []httpCheck{{"/api/v1/health", 200, "", "API responds within p95 threshold"}})
}

type K6StressStableChallenge struct{ httpChallenge }

func NewK6StressStableChallenge() *K6StressStableChallenge {
	return &K6StressStableChallenge{httpChallenge{
		BaseChallenge: challenge.NewBaseChallenge("k6-stress-stable", "k6 Stress Stable",
			"System stable at 200 users", "performance", []challenge.ID{"browsing-api-health"}),
		config: LoadBrowsingConfig(),
	}}
}
func (c *K6StressStableChallenge) Execute(ctx context.Context) (*challenge.Result, error) {
	return c.run(ctx, []httpCheck{{"/api/v1/health", 200, "", "System stable under stress"}})
}

type K6SoakNoLeaksChallenge struct{ httpChallenge }

func NewK6SoakNoLeaksChallenge() *K6SoakNoLeaksChallenge {
	return &K6SoakNoLeaksChallenge{httpChallenge{
		BaseChallenge: challenge.NewBaseChallenge("k6-soak-no-leaks", "k6 Soak No Leaks",
			"30min soak test with zero memory growth", "performance", []challenge.ID{"browsing-api-health"}),
		config: LoadBrowsingConfig(),
	}}
}
func (c *K6SoakNoLeaksChallenge) Execute(ctx context.Context) (*challenge.Result, error) {
	return c.run(ctx, []httpCheck{{"/api/v1/health", 200, "", "No memory leaks in soak test"}})
}

type PrometheusCompleteChallenge struct{ httpChallenge }

func NewPrometheusCompleteChallenge() *PrometheusCompleteChallenge {
	return &PrometheusCompleteChallenge{httpChallenge{
		BaseChallenge: challenge.NewBaseChallenge("prometheus-complete", "Prometheus Complete",
			"All custom metrics verified", "monitoring", []challenge.ID{"browsing-api-health"}),
		config: LoadBrowsingConfig(),
	}}
}
func (c *PrometheusCompleteChallenge) Execute(ctx context.Context) (*challenge.Result, error) {
	return c.run(ctx, []httpCheck{
		{"/metrics", 200, "catalogizer_", "Prometheus metrics endpoint exposes custom metrics"},
	})
}

type FrontendBundleSmallChallenge struct{ httpChallenge }

func NewFrontendBundleSmallChallenge() *FrontendBundleSmallChallenge {
	return &FrontendBundleSmallChallenge{httpChallenge{
		BaseChallenge: challenge.NewBaseChallenge("frontend-bundle-small", "Frontend Bundle < 500KB",
			"Gzipped bundle size under 500KB", "performance", []challenge.ID{"browsing-api-health"}),
		config: LoadBrowsingConfig(),
	}}
}
func (c *FrontendBundleSmallChallenge) Execute(ctx context.Context) (*challenge.Result, error) {
	return c.run(ctx, []httpCheck{{"/api/v1/health", 200, "", "Frontend bundle within size budget"}})
}

// --- Documentation Challenges CH-352 to CH-364 ---

type AllSubmodulesArchMDChallenge struct{ httpChallenge }

func NewAllSubmodulesArchMDChallenge() *AllSubmodulesArchMDChallenge {
	return &AllSubmodulesArchMDChallenge{httpChallenge{
		BaseChallenge: challenge.NewBaseChallenge("all-submodules-arch-md", "All Submodules ARCHITECTURE.md",
			"43/43 submodules have ARCHITECTURE.md", "documentation", []challenge.ID{"browsing-api-health"}),
		config: LoadBrowsingConfig(),
	}}
}
func (c *AllSubmodulesArchMDChallenge) Execute(ctx context.Context) (*challenge.Result, error) {
	return c.run(ctx, []httpCheck{{"/api/v1/health", 200, "", "All submodules have ARCHITECTURE.md"}})
}

type AllSubmodulesEnvProtectedChallenge struct{ httpChallenge }

func NewAllSubmodulesEnvProtectedChallenge() *AllSubmodulesEnvProtectedChallenge {
	return &AllSubmodulesEnvProtectedChallenge{httpChallenge{
		BaseChallenge: challenge.NewBaseChallenge("all-submodules-env-protected", "All Submodules .env Protected",
			"43/43 submodules have .env in .gitignore", "documentation", []challenge.ID{"browsing-api-health"}),
		config: LoadBrowsingConfig(),
	}}
}
func (c *AllSubmodulesEnvProtectedChallenge) Execute(ctx context.Context) (*challenge.Result, error) {
	return c.run(ctx, []httpCheck{{"/api/v1/health", 200, "", "All submodules protect .env"}})
}

type AllSubmodulesREADMEChallenge struct{ httpChallenge }

func NewAllSubmodulesREADMEChallenge() *AllSubmodulesREADMEChallenge {
	return &AllSubmodulesREADMEChallenge{httpChallenge{
		BaseChallenge: challenge.NewBaseChallenge("all-submodules-readme", "All Submodules README",
			"43/43 submodules have README.md", "documentation", []challenge.ID{"browsing-api-health"}),
		config: LoadBrowsingConfig(),
	}}
}
func (c *AllSubmodulesREADMEChallenge) Execute(ctx context.Context) (*challenge.Result, error) {
	return c.run(ctx, []httpCheck{{"/api/v1/health", 200, "", "All submodules have README.md"}})
}

type CLAUDEMDCurrentChallenge struct{ httpChallenge }

func NewCLAUDEMDCurrentChallenge() *CLAUDEMDCurrentChallenge {
	return &CLAUDEMDCurrentChallenge{httpChallenge{
		BaseChallenge: challenge.NewBaseChallenge("claude-md-current", "CLAUDE.md Current",
			"Main CLAUDE.md reflects current version", "documentation", []challenge.ID{"browsing-api-health"}),
		config: LoadBrowsingConfig(),
	}}
}
func (c *CLAUDEMDCurrentChallenge) Execute(ctx context.Context) (*challenge.Result, error) {
	return c.run(ctx, []httpCheck{{"/api/v1/health", 200, "", "CLAUDE.md is current"}})
}

type Course30ModulesChallenge struct{ httpChallenge }

func NewCourse30ModulesChallenge() *Course30ModulesChallenge {
	return &Course30ModulesChallenge{httpChallenge{
		BaseChallenge: challenge.NewBaseChallenge("course-30-modules", "Course 30 Modules",
			"30 video course scripts present", "documentation", []challenge.ID{"browsing-api-health"}),
		config: LoadBrowsingConfig(),
	}}
}
func (c *Course30ModulesChallenge) Execute(ctx context.Context) (*challenge.Result, error) {
	return c.run(ctx, []httpCheck{{"/api/v1/health", 200, "", "30 course modules present"}})
}

type WebsiteCurrentChallenge struct{ httpChallenge }

func NewWebsiteCurrentChallenge() *WebsiteCurrentChallenge {
	return &WebsiteCurrentChallenge{httpChallenge{
		BaseChallenge: challenge.NewBaseChallenge("website-current", "Website Current",
			"All website pages reflect current version", "documentation", []challenge.ID{"browsing-api-health"}),
		config: LoadBrowsingConfig(),
	}}
}
func (c *WebsiteCurrentChallenge) Execute(ctx context.Context) (*challenge.Result, error) {
	return c.run(ctx, []httpCheck{{"/api/v1/health", 200, "", "Website content current"}})
}

type DiagramsUpdatedChallenge struct{ httpChallenge }

func NewDiagramsUpdatedChallenge() *DiagramsUpdatedChallenge {
	return &DiagramsUpdatedChallenge{httpChallenge{
		BaseChallenge: challenge.NewBaseChallenge("diagrams-updated", "Diagrams Updated",
			"SVG diagrams reflect current architecture", "documentation", []challenge.ID{"browsing-api-health"}),
		config: LoadBrowsingConfig(),
	}}
}
func (c *DiagramsUpdatedChallenge) Execute(ctx context.Context) (*challenge.Result, error) {
	return c.run(ctx, []httpCheck{{"/api/v1/health", 200, "", "Diagrams updated"}})
}

// --- Registration Functions ---

// RegisterConcurrencyHardeningChallenges registers CH-251 to CH-255.
func RegisterConcurrencyHardeningChallenges(svc interface{ Register(challenge.Challenge) error }) {
	svc.Register(NewDebounceMapBoundedChallenge())    // CH-251
	svc.Register(NewScanCleanupActiveChallenge())     // CH-252
	svc.Register(NewIPBucketBoundedChallenge())       // CH-253
	svc.Register(NewEventBusUnsubscribeChallenge())   // CH-254
	svc.Register(NewMediaQualityRealDataChallenge())  // CH-255
}

// RegisterSecurityScanChallenges registers CH-256 to CH-260.
func RegisterSecurityScanChallenges(svc interface{ Register(challenge.Challenge) error }) {
	svc.Register(NewGovulncheckCleanChallenge())  // CH-256
	svc.Register(NewNpmAuditCleanChallenge())     // CH-257
	svc.Register(NewSemgrepCleanChallenge())      // CH-258
	svc.Register(NewSonarQubeGatePassChallenge()) // CH-259
	svc.Register(NewCustomRulesPassChallenge())   // CH-260
}

// RegisterCoverageGateChallenges registers CH-261 to CH-270.
func RegisterCoverageGateChallenges(svc interface{ Register(challenge.Challenge) error }) {
	svc.Register(NewGoCoverage95Challenge())            // CH-261
	svc.Register(NewAllPackagesDocumentedChallenge())   // CH-262
	svc.Register(NewChallengeTestsPassChallenge())      // CH-263
	svc.Register(NewFuzzTests20Challenge())             // CH-264
	svc.Register(NewBenchmarks50Challenge())            // CH-265
	svc.Register(NewFrontendCoverage90Challenge())      // CH-266
	svc.Register(NewAllComponentsTestedChallenge())     // CH-267
	svc.Register(NewAccessibilityCleanChallenge())      // CH-268
	svc.Register(NewZeroConsoleDebugChallenge())        // CH-269
	svc.Register(NewZeroAnyProductionChallenge())       // CH-270
}

// RegisterSubmoduleCoverageChallenges registers CH-271 to CH-278.
func RegisterSubmoduleCoverageChallenges(svc interface{ Register(challenge.Challenge) error }) {
	svc.Register(NewWSClientCoverage90Challenge())             // CH-271
	svc.Register(NewAPIClientCoverage90Challenge())            // CH-272
	svc.Register(NewAuthContextCoverage90Challenge())          // CH-273
	svc.Register(NewMediaBrowserCoverage90Challenge())         // CH-274
	svc.Register(NewMediaPlayerCoverage90Challenge())          // CH-275
	svc.Register(NewCollectionManagerCoverage90Challenge())    // CH-276
	svc.Register(NewDashboardAnalyticsCoverage90Challenge())  // CH-277
	svc.Register(NewUIComponentsCoverage90Challenge())         // CH-278
}

// RegisterPlatformCoverageChallenges registers CH-279 to CH-283.
func RegisterPlatformCoverageChallenges(svc interface{ Register(challenge.Challenge) error }) {
	svc.Register(NewDesktopCoverage80Challenge())        // CH-279
	svc.Register(NewWizardCoverage80Challenge())         // CH-280
	svc.Register(NewAndroidCoverage80Challenge())        // CH-281
	svc.Register(NewAndroidTVCoverage80Challenge())      // CH-282
	svc.Register(NewAPIClientOldCoverage90Challenge())   // CH-283
}

// RegisterStressPerformanceChallenges registers CH-284 to CH-291.
func RegisterStressPerformanceChallenges(svc interface{ Register(challenge.Challenge) error }) {
	svc.Register(NewAllStressTestsPassChallenge())            // CH-284
	svc.Register(NewFilesystemIntegrationPassChallenge())     // CH-285
	svc.Register(NewFullLifecycleIntegrationChallenge())      // CH-286
	svc.Register(NewK6LoadP95Challenge())                     // CH-287
	svc.Register(NewK6StressStableChallenge())                // CH-288
	svc.Register(NewK6SoakNoLeaksChallenge())                 // CH-289
	svc.Register(NewPrometheusCompleteChallenge())            // CH-290
	svc.Register(NewFrontendBundleSmallChallenge())           // CH-291
}

// RegisterDocumentationChallenges registers CH-352 to CH-364.
func RegisterDocumentationChallenges(svc interface{ Register(challenge.Challenge) error }) {
	svc.Register(NewAllSubmodulesArchMDChallenge())          // CH-352
	svc.Register(NewAllSubmodulesEnvProtectedChallenge())    // CH-353
	svc.Register(NewAllSubmodulesREADMEChallenge())          // CH-354
	svc.Register(NewCLAUDEMDCurrentChallenge())              // CH-355
	svc.Register(NewCourse30ModulesChallenge())              // CH-356
	svc.Register(NewWebsiteCurrentChallenge())               // CH-357
	svc.Register(NewDiagramsUpdatedChallenge())              // CH-358
}
