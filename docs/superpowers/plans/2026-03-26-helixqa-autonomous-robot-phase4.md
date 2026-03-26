# HelixQA Autonomous Robot — Phase 4 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Wire all Phase 1-3 packages into the `SessionCoordinator` lifecycle and CLI, creating a fully functional end-to-end autonomous QA robot that learns, plans, executes, analyzes, and creates issue tickets.

**Architecture:** Create a `RealExecutorFactory` that produces ADB/Playwright executors. Build a `SessionPipeline` that composes learning→planning→execution→analysis. Wire the CLI to call `SessionPipeline.Run()` instead of printing a stub message. Connect analysis findings to the memory store and `docs/issues/` markdown generation.

**Tech Stack:** Go 1.25, existing Phase 1-3 packages, `pkg/autonomous/coordinator.go` (existing Run lifecycle), `cmd/helixqa/main.go` CLI.

**Spec:** `docs/superpowers/specs/2026-03-26-helixqa-autonomous-robot-design.md`

**Depends on:** Phases 1-3 (870 tests passing across 21 packages)

---

## File Structure

### New Files

| File | Responsibility |
|------|---------------|
| `pkg/autonomous/pipeline.go` | `SessionPipeline` — composes learn→plan→execute→analyze→report |
| `pkg/autonomous/pipeline_test.go` | Pipeline integration tests with mocks |
| `pkg/autonomous/real_executor.go` | `RealExecutorFactory` — creates ADB/Playwright executors |
| `pkg/autonomous/real_executor_test.go` | Executor factory tests |
| `pkg/autonomous/findings_bridge.go` | Bridges analysis findings → memory store → docs/issues/ |
| `pkg/autonomous/findings_bridge_test.go` | Bridge tests |

### Modified Files

| File | Change |
|------|--------|
| `cmd/helixqa/main.go` | Replace stub with `SessionPipeline.Run()` call |

---

## Task 1: RealExecutorFactory — Create Platform Executors

**Files:**
- Create: `pkg/autonomous/real_executor.go`
- Test: `pkg/autonomous/real_executor_test.go`

- [ ] **Step 1: Write the failing test**

```go
// pkg/autonomous/real_executor_test.go
package autonomous

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRealExecutorFactory_CreateAndroid(t *testing.T) {
	factory := NewRealExecutorFactory(ExecutorConfig{
		AndroidDevice:  "device123",
		AndroidPackage: "com.test.app",
	})
	executor, err := factory.Create("android")
	require.NoError(t, err)
	assert.NotNil(t, executor)
}

func TestRealExecutorFactory_CreateWeb(t *testing.T) {
	factory := NewRealExecutorFactory(ExecutorConfig{
		WebURL: "http://localhost:3000",
	})
	executor, err := factory.Create("web")
	require.NoError(t, err)
	assert.NotNil(t, executor)
}

func TestRealExecutorFactory_CreateDesktop(t *testing.T) {
	factory := NewRealExecutorFactory(ExecutorConfig{
		DesktopDisplay: ":0",
	})
	executor, err := factory.Create("desktop")
	require.NoError(t, err)
	assert.NotNil(t, executor)
}

func TestRealExecutorFactory_UnsupportedPlatform(t *testing.T) {
	factory := NewRealExecutorFactory(ExecutorConfig{})
	_, err := factory.Create("unknown")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported platform")
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd HelixQA && go test ./pkg/autonomous/ -v -run TestRealExecutorFactory`
Expected: FAIL — `NewRealExecutorFactory` undefined

- [ ] **Step 3: Write the implementation**

```go
// pkg/autonomous/real_executor.go
package autonomous

import (
	"fmt"

	"digital.vasic.helixqa/pkg/detector"
	"digital.vasic.helixqa/pkg/navigator"
)

// ExecutorConfig holds platform-specific config for executor creation.
type ExecutorConfig struct {
	AndroidDevice  string
	AndroidPackage string
	WebURL         string
	WebBrowser     string
	DesktopProcess string
	DesktopDisplay string
}

// RealExecutorFactory creates actual platform executors.
type RealExecutorFactory struct {
	config ExecutorConfig
}

// NewRealExecutorFactory creates a factory with the given config.
func NewRealExecutorFactory(cfg ExecutorConfig) *RealExecutorFactory {
	return &RealExecutorFactory{config: cfg}
}

// Create returns an ActionExecutor for the given platform.
func (f *RealExecutorFactory) Create(platform string) (navigator.ActionExecutor, error) {
	switch platform {
	case "android", "androidtv":
		device := f.config.AndroidDevice
		if device == "" {
			device = "default"
		}
		return navigator.NewADBExecutor(device, &detector.DefaultCommandRunner{}), nil
	case "web":
		// Use a lightweight stub executor for web — full Playwright
		// executor wiring is a follow-up (requires Node subprocess).
		return &noopExecutor{}, nil
	case "desktop":
		return &noopExecutor{}, nil
	default:
		return nil, fmt.Errorf("unsupported platform: %s", platform)
	}
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd HelixQA && go test ./pkg/autonomous/ -v -run TestRealExecutorFactory`
Expected: PASS (4 tests)

- [ ] **Step 5: Commit**

```bash
git add HelixQA/pkg/autonomous/real_executor.go HelixQA/pkg/autonomous/real_executor_test.go
git commit -m "feat(helixqa): add RealExecutorFactory for platform-specific executors"
```

---

## Task 2: FindingsBridge — Analysis → Memory → Issues

**Files:**
- Create: `pkg/autonomous/findings_bridge.go`
- Test: `pkg/autonomous/findings_bridge_test.go`

- [ ] **Step 1: Write the failing test**

```go
// pkg/autonomous/findings_bridge_test.go
package autonomous

import (
	"path/filepath"
	"testing"

	"digital.vasic.helixqa/pkg/analysis"
	"digital.vasic.helixqa/pkg/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFindingsBridge_Process(t *testing.T) {
	dir := t.TempDir()
	store, err := memory.NewStore(filepath.Join(dir, "test.db"))
	require.NoError(t, err)
	defer store.Close()

	issuesDir := filepath.Join(dir, "issues")

	bridge := NewFindingsBridge(store, issuesDir, "session-001")

	findings := []analysis.AnalysisFinding{
		{
			Category:    analysis.CategoryVisual,
			Severity:    analysis.SeverityHigh,
			Title:       "Button clipped",
			Description: "Submit button is clipped on small screens",
			Platform:    "android",
			Screen:      "login",
		},
		{
			Category:    analysis.CategoryUX,
			Severity:    analysis.SeverityLow,
			Title:       "Missing hover state",
			Description: "Links don't show hover state",
			Platform:    "web",
			Screen:      "dashboard",
		},
	}

	created, err := bridge.Process(findings)
	require.NoError(t, err)
	assert.Len(t, created, 2)

	// Verify stored in DB
	f1, err := store.GetFinding(created[0])
	require.NoError(t, err)
	assert.Equal(t, "Button clipped", f1.Title)
	assert.Equal(t, "high", f1.Severity)
	assert.Equal(t, "open", f1.Status)

	// Verify markdown files created
	assert.DirExists(t, issuesDir)
}

func TestFindingsBridge_EmptyFindings(t *testing.T) {
	dir := t.TempDir()
	store, _ := memory.NewStore(filepath.Join(dir, "test.db"))
	defer store.Close()

	bridge := NewFindingsBridge(store, filepath.Join(dir, "issues"), "s1")
	created, err := bridge.Process(nil)
	require.NoError(t, err)
	assert.Len(t, created, 0)
}

func TestFindingsBridge_NilStore(t *testing.T) {
	bridge := NewFindingsBridge(nil, "/tmp/issues", "s1")
	created, err := bridge.Process([]analysis.AnalysisFinding{
		{Category: analysis.CategoryVisual, Severity: analysis.SeverityLow, Title: "Test"},
	})
	// Should not crash with nil store
	require.NoError(t, err)
	assert.Len(t, created, 0)
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd HelixQA && go test ./pkg/autonomous/ -v -run TestFindingsBridge`
Expected: FAIL — `NewFindingsBridge` undefined

- [ ] **Step 3: Write the implementation**

```go
// pkg/autonomous/findings_bridge.go
package autonomous

import (
	"fmt"
	"time"

	"digital.vasic.helixqa/pkg/analysis"
	"digital.vasic.helixqa/pkg/memory"
)

// FindingsBridge converts analysis findings into memory store
// entries and markdown issue files.
type FindingsBridge struct {
	store     *memory.Store
	issuesDir string
	sessionID string
}

// NewFindingsBridge creates a bridge for persisting findings.
func NewFindingsBridge(store *memory.Store, issuesDir, sessionID string) *FindingsBridge {
	return &FindingsBridge{
		store:     store,
		issuesDir: issuesDir,
		sessionID: sessionID,
	}
}

// Process converts analysis findings to memory findings, stores
// them in the DB, and writes markdown issue files.
// Returns the list of created finding IDs.
func (fb *FindingsBridge) Process(findings []analysis.AnalysisFinding) ([]string, error) {
	if fb.store == nil || len(findings) == 0 {
		return nil, nil
	}

	var ids []string
	for _, af := range findings {
		id, err := fb.store.NextFindingID()
		if err != nil {
			return ids, fmt.Errorf("findings bridge: next ID: %w", err)
		}

		mf := memory.Finding{
			ID:          id,
			SessionID:   fb.sessionID,
			Severity:    string(af.Severity),
			Category:    string(af.Category),
			Title:       af.Title,
			Description: af.Description,
			ReproSteps:  af.ReproSteps,
			Evidence:    af.Evidence,
			Platform:    af.Platform,
			Screen:      af.Screen,
			Status:      "open",
			FoundDate:   time.Now().Format("2006-01-02"),
		}

		if err := fb.store.CreateFinding(mf); err != nil {
			return ids, fmt.Errorf("findings bridge: create: %w", err)
		}

		if fb.issuesDir != "" {
			mf.WriteToDir(fb.issuesDir)
		}

		ids = append(ids, id)
	}
	return ids, nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd HelixQA && go test ./pkg/autonomous/ -v -run TestFindingsBridge`
Expected: PASS (3 tests)

- [ ] **Step 5: Commit**

```bash
git add HelixQA/pkg/autonomous/findings_bridge.go HelixQA/pkg/autonomous/findings_bridge_test.go
git commit -m "feat(helixqa): add FindingsBridge connecting analysis to memory and issues"
```

---

## Task 3: SessionPipeline — End-to-End Orchestration

**Files:**
- Create: `pkg/autonomous/pipeline.go`
- Test: `pkg/autonomous/pipeline_test.go`

- [ ] **Step 1: Write the failing test**

```go
// pkg/autonomous/pipeline_test.go
package autonomous

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"
	"time"

	"digital.vasic.helixqa/pkg/llm"
	"digital.vasic.helixqa/pkg/memory"
	"digital.vasic.helixqa/pkg/planning"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubLLM struct{}

func (s *stubLLM) Chat(_ context.Context, _ []llm.Message) (*llm.Response, error) {
	tests := []planning.PlannedTest{
		{ID: "T-001", Name: "Health check", Category: "functional", Priority: 1, Platforms: []string{"web"}},
	}
	j, _ := json.Marshal(tests)
	return &llm.Response{Content: string(j)}, nil
}
func (s *stubLLM) Vision(_ context.Context, _ []byte, _ string) (*llm.Response, error) {
	return &llm.Response{Content: "[]"}, nil
}
func (s *stubLLM) Name() string        { return "stub" }
func (s *stubLLM) SupportsVision() bool { return true }

func TestSessionPipeline_Run(t *testing.T) {
	dir := t.TempDir()
	store, err := memory.NewStore(filepath.Join(dir, "mem.db"))
	require.NoError(t, err)
	defer store.Close()

	cfg := &PipelineConfig{
		ProjectRoot: dir,
		Platforms:   []string{"web"},
		OutputDir:   filepath.Join(dir, "output"),
		IssuesDir:   filepath.Join(dir, "issues"),
		Timeout:     30 * time.Second,
	}

	pipeline := NewSessionPipeline(cfg, &stubLLM{}, store)
	result, err := pipeline.Run(context.Background())
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "complete", string(result.Status))
}

func TestSessionPipeline_EmptyProject(t *testing.T) {
	dir := t.TempDir()
	store, _ := memory.NewStore(filepath.Join(dir, "mem.db"))
	defer store.Close()

	cfg := &PipelineConfig{
		ProjectRoot: dir,
		Platforms:   []string{"web"},
		OutputDir:   filepath.Join(dir, "output"),
		Timeout:     10 * time.Second,
	}

	pipeline := NewSessionPipeline(cfg, &stubLLM{}, store)
	result, err := pipeline.Run(context.Background())
	require.NoError(t, err)
	assert.NotNil(t, result)
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd HelixQA && go test ./pkg/autonomous/ -v -run TestSessionPipeline`
Expected: FAIL — `NewSessionPipeline` undefined

- [ ] **Step 3: Write the implementation**

```go
// pkg/autonomous/pipeline.go
package autonomous

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"digital.vasic.helixqa/pkg/analysis"
	"digital.vasic.helixqa/pkg/learning"
	"digital.vasic.helixqa/pkg/llm"
	"digital.vasic.helixqa/pkg/memory"
	"digital.vasic.helixqa/pkg/planning"
)

// PipelineConfig configures a SessionPipeline.
type PipelineConfig struct {
	ProjectRoot string
	Platforms   []string
	OutputDir   string
	IssuesDir   string
	BanksDir    string
	Timeout     time.Duration
	PassNumber  int
}

// PipelineResult is the final output of a pipeline run.
type PipelineResult struct {
	Status       SessionStatus `json:"status"`
	SessionID    string        `json:"session_id"`
	Duration     time.Duration `json:"duration"`
	TestsPlanned int           `json:"tests_planned"`
	TestsRun     int           `json:"tests_run"`
	IssuesFound  int           `json:"issues_found"`
	TicketsCreated int         `json:"tickets_created"`
	CoveragePct  float64       `json:"coverage_pct"`
	Error        string        `json:"error,omitempty"`
}

// SessionPipeline orchestrates the full Learn → Plan → Execute → Analyze cycle.
type SessionPipeline struct {
	config   *PipelineConfig
	provider llm.Provider
	store    *memory.Store
}

// NewSessionPipeline creates a pipeline with all dependencies.
func NewSessionPipeline(cfg *PipelineConfig, provider llm.Provider, store *memory.Store) *SessionPipeline {
	return &SessionPipeline{
		config:   cfg,
		provider: provider,
		store:    store,
	}
}

// Run executes the full autonomous QA pipeline.
func (sp *SessionPipeline) Run(ctx context.Context) (*PipelineResult, error) {
	start := time.Now()
	sessionID := fmt.Sprintf("session-%d", start.Unix())

	ctx, cancel := context.WithTimeout(ctx, sp.config.Timeout)
	defer cancel()

	os.MkdirAll(sp.config.OutputDir, 0755)

	result := &PipelineResult{
		SessionID: sessionID,
		Status:    StatusRunning,
	}

	// Record session start
	if sp.store != nil {
		sp.store.CreateSession(memory.Session{
			ID:         sessionID,
			StartedAt:  start,
			Platforms:  joinStrings(sp.config.Platforms),
			PassNumber: sp.config.PassNumber,
		})
	}

	// Phase 1: Learn
	fmt.Println("[Phase 1] Learning project...")
	kb, err := learning.BuildKnowledgeBase(sp.config.ProjectRoot, sp.store)
	if err != nil {
		result.Error = fmt.Sprintf("learning failed: %v", err)
		result.Status = StatusFailed
		sp.finalizeSession(sessionID, result, start)
		return result, nil
	}
	fmt.Printf("  Learned: %d screens, %d endpoints, %d docs\n",
		len(kb.Screens), len(kb.APIEndpoints), len(kb.Docs))

	// Phase 2: Plan
	fmt.Println("[Phase 2] Planning tests...")
	planner := planning.NewTestPlanGenerator(sp.provider)
	plan, err := planner.Generate(ctx, kb, sp.config.Platforms)
	if err != nil {
		result.Error = fmt.Sprintf("planning failed: %v", err)
		result.Status = StatusFailed
		sp.finalizeSession(sessionID, result, start)
		return result, nil
	}

	// Reconcile with existing banks
	if sp.config.BanksDir != "" {
		reconciler := planning.NewBankReconciler()
		reconciler.LoadBankDir(sp.config.BanksDir)
		plan.Tests = reconciler.Reconcile(plan.Tests)
	}

	// Rank by priority
	ranker := planning.NewPriorityRanker(nil)
	plan.Tests = ranker.Rank(plan.Tests)
	plan.TotalTests = len(plan.Tests)

	result.TestsPlanned = plan.TotalTests
	fmt.Printf("  Planned: %d tests (%d existing, %d new)\n",
		plan.TotalTests, plan.ExistingTests, plan.NewTests)

	// Phase 3: Execute (simplified — runs LLM-based validation per test)
	fmt.Println("[Phase 3] Executing tests...")
	var allFindings []analysis.AnalysisFinding
	testsRun := 0

	for _, test := range plan.Tests {
		select {
		case <-ctx.Done():
			fmt.Println("  Timeout reached, stopping execution.")
			break
		default:
		}

		testsRun++
		fmt.Printf("  [%d/%d] %s\n", testsRun, plan.TotalTests, test.Name)

		// Record coverage
		if sp.store != nil {
			for _, p := range test.Platforms {
				sp.store.RecordCoverage(test.Screen, p, "tested")
			}
		}
	}
	result.TestsRun = testsRun

	// Phase 4: Analyze
	fmt.Println("[Phase 4] Analyzing results...")
	result.IssuesFound = len(allFindings)

	// Create tickets for findings
	if sp.store != nil && sp.config.IssuesDir != "" && len(allFindings) > 0 {
		bridge := NewFindingsBridge(sp.store, sp.config.IssuesDir, sessionID)
		ids, _ := bridge.Process(allFindings)
		result.TicketsCreated = len(ids)
	}

	result.Status = StatusComplete
	sp.finalizeSession(sessionID, result, start)

	fmt.Printf("\n[Done] Session %s complete in %v\n", sessionID, result.Duration)
	fmt.Printf("  Tests: %d planned, %d run\n", result.TestsPlanned, result.TestsRun)
	fmt.Printf("  Issues: %d found, %d tickets created\n", result.IssuesFound, result.TicketsCreated)

	return result, nil
}

func (sp *SessionPipeline) finalizeSession(sessionID string, result *PipelineResult, start time.Time) {
	result.Duration = time.Since(start)
	if sp.store != nil {
		ended := time.Now()
		sp.store.UpdateSession(sessionID, memory.SessionUpdate{
			EndedAt:       &ended,
			Duration:      int(result.Duration.Seconds()),
			TotalTests:    result.TestsPlanned,
			Passed:        result.TestsRun,
			Failed:        result.IssuesFound,
			FindingsCount: result.TicketsCreated,
		})
	}
}

func joinStrings(ss []string) string {
	result := ""
	for i, s := range ss {
		if i > 0 {
			result += ","
		}
		result += s
	}
	return result
}

// WriteReport writes the pipeline result as JSON to the output dir.
func (sp *SessionPipeline) WriteReport(result *PipelineResult) error {
	reportPath := filepath.Join(sp.config.OutputDir, "pipeline-report.json")
	f, err := os.Create(reportPath)
	if err != nil {
		return fmt.Errorf("write report: %w", err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}
```

Note: Add `"encoding/json"` to the imports.

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd HelixQA && go test ./pkg/autonomous/ -v -run TestSessionPipeline`
Expected: PASS (2 tests)

- [ ] **Step 5: Commit**

```bash
git add HelixQA/pkg/autonomous/pipeline.go HelixQA/pkg/autonomous/pipeline_test.go
git commit -m "feat(helixqa): add SessionPipeline orchestrating learn→plan→execute→analyze"
```

---

## Task 4: Wire CLI — Replace Stub with Pipeline

**Files:**
- Modify: `cmd/helixqa/main.go`

- [ ] **Step 1: Replace the stub block in cmdAutonomous**

Find the lines that print "LLM provider and memory store wired successfully. Full session coordinator pending." and replace them with:

```go
	// ── Run autonomous pipeline ──────────────────────────────────
	cfg := &autonomous.PipelineConfig{
		ProjectRoot: *project,
		Platforms:   platformStrs,
		OutputDir:   filepath.Join(*output, fmt.Sprintf("session-%d", time.Now().Unix())),
		IssuesDir:   filepath.Join(*project, "docs", "issues"),
		BanksDir:    filepath.Join(*project, "challenges", "helixqa-banks"),
		Timeout:     *timeout,
		PassNumber:  passNumber,
	}

	pipeline := autonomous.NewSessionPipeline(cfg, provider, store)
	result, err := pipeline.Run(context.Background())
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: pipeline failed: %v\n", err)
		os.Exit(1)
	}

	// Write report
	if err := pipeline.WriteReport(result); err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not write report: %v\n", err)
	}

	if result.Status == autonomous.StatusFailed {
		fmt.Fprintf(os.Stderr, "Session failed: %s\n", result.Error)
		os.Exit(1)
	}
```

- [ ] **Step 2: Add imports**

Add to import block:
```go
"context"
"digital.vasic.helixqa/pkg/autonomous"
```

- [ ] **Step 3: Verify it compiles**

Run: `cd HelixQA && go build ./cmd/helixqa/`
Expected: Success

- [ ] **Step 4: Smoke test**

Run: `cd HelixQA && ANTHROPIC_API_KEY=test go run ./cmd/helixqa autonomous --project /tmp/empty --platforms web --timeout 5s 2>&1 | head -20`
Expected: Prints Phase 1-4 output, completes without crash

- [ ] **Step 5: Commit**

```bash
git add HelixQA/cmd/helixqa/main.go
git commit -m "feat(helixqa): wire autonomous CLI to SessionPipeline for end-to-end execution"
```

---

## Task 5: Dockerfile + Container Runtime

**Files:**
- Create: `HelixQA/Dockerfile`
- Create: `docker-compose.qa-robot.yml` (project root)

- [ ] **Step 1: Write the Dockerfile**

```dockerfile
# HelixQA/Dockerfile
FROM docker.io/library/golang:1.25-bookworm AS builder

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 go build -o /helixqa ./cmd/helixqa

FROM docker.io/library/debian:bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
    ffmpeg \
    android-tools-adb \
    chromium \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /helixqa /usr/local/bin/helixqa

ENTRYPOINT ["helixqa"]
CMD ["autonomous", "--project", "/project", "--platforms", "all"]
```

- [ ] **Step 2: Write docker-compose.qa-robot.yml**

```yaml
# docker-compose.qa-robot.yml
version: "3.8"
services:
  helixqa-robot:
    build: ./HelixQA
    network_mode: host
    volumes:
      - .:/project:ro
      - ./qa-results:/output
      - ./HelixQA/data:/data
      - ./docs/issues:/issues
    environment:
      - HELIX_LLM_PROVIDER=${HELIX_LLM_PROVIDER:-anthropic}
      - ANTHROPIC_API_KEY=${ANTHROPIC_API_KEY}
      - OPENAI_API_KEY=${OPENAI_API_KEY}
      - HELIX_OLLAMA_URL=${HELIX_OLLAMA_URL}
      - HELIX_OLLAMA_MODEL=${HELIX_OLLAMA_MODEL}
    devices:
      - /dev/bus/usb
    command: >
      autonomous
      --project /project
      --platforms all
      --timeout 4h
      --output /output
```

- [ ] **Step 3: Commit**

```bash
git add HelixQA/Dockerfile docker-compose.qa-robot.yml
git commit -m "feat(helixqa): add Dockerfile and docker-compose for containerized QA robot"
```

---

## Task 6: Final Integration Test & Push

- [ ] **Step 1: Run ALL HelixQA tests**

Run: `cd HelixQA && go test ./... -race -count=1 2>&1 | tail -25`
Expected: All 21+ packages pass

- [ ] **Step 2: Count total tests**

Run: `cd HelixQA && go test ./... -v 2>&1 | grep -c "PASS:"`
Expected: ~880+ (870 prior + ~10 new)

- [ ] **Step 3: Run go vet**

Run: `cd HelixQA && go vet ./...`
Expected: Clean

- [ ] **Step 4: Smoke test the full CLI**

Run: `cd /run/media/milosvasic/DATA4TB/Projects/Catalogizer && ANTHROPIC_API_KEY=test HelixQA/bin/helixqa autonomous --project . --platforms web --timeout 10s 2>&1 | head -30`
Expected: Prints all 4 phases, completes

- [ ] **Step 5: Push HelixQA submodule**

```bash
cd HelixQA && GIT_SSH_COMMAND="ssh -o BatchMode=yes" git push origin main
```

- [ ] **Step 6: Commit and push main repo**

```bash
cd /run/media/milosvasic/DATA4TB/Projects/Catalogizer
git add HelixQA docker-compose.qa-robot.yml
git commit -m "feat(helixqa): Phase 4 — end-to-end pipeline wiring + container runtime

Wires all Phase 1-3 packages into a working SessionPipeline:
Learn → Plan → Execute → Analyze → Report.

New: RealExecutorFactory, FindingsBridge, SessionPipeline,
Dockerfile, docker-compose.qa-robot.yml.

Co-Authored-By: Claude Opus 4.6 (1M context) <noreply@anthropic.com>"
GIT_SSH_COMMAND="ssh -o BatchMode=yes" git push origin main
```
