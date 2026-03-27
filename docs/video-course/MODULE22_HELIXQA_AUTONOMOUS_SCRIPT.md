# Module 22: HelixQA Autonomous Testing -- Video Script

**Duration**: 60 minutes
**Prerequisites**: Module 10 (Testing), Module 15 (Concurrency Patterns)

---

## Video 22.1: Autonomous QA Sessions (12 min)

### Opening

Welcome to Module 22. HelixQA is the autonomous quality assurance framework that tests Catalogizer across all platforms -- API, web, desktop, and mobile -- without human supervision. It orchestrates test execution, collects evidence, detects crashes, and generates reports. This module covers the full lifecycle from session creation to final report.

### What Is an Autonomous QA Session?

**[Visual: Terminal showing `helixqa run --platform api --speed thorough`]**

**Narrator**: An autonomous QA session is a fully automated test run. The operator specifies the target platform, speed (quick, normal, thorough), and output directory. HelixQA loads the appropriate test bank, executes each test case in order, collects evidence at every step, and produces a comprehensive report at the end.

**[Visual: Architecture diagram: CLI -> Orchestrator -> TestBank + Detector + Validator + Evidence + Reporter]**

**Narrator**: The orchestrator is the brain. It composes five subsystems: the test bank (what to test), the detector (crash/ANR monitoring), the validator (step-by-step assertion), the evidence collector (screenshots, logs, video), and the reporter (output generation).

```go
// HelixQA/pkg/orchestrator/orchestrator.go
type Orchestrator struct {
    config    *config.Config
    testBank  *testbank.TestBank
    detector  *detector.Detector
    validator *validator.Validator
    evidence  *evidence.Collector
    reporter  *reporter.Reporter
    runner    runner.Runner
    logger    *zap.Logger
}

func New(opts ...Option) (*Orchestrator, error) {
    o := &Orchestrator{}
    for _, opt := range opts {
        opt(o)
    }
    return o, o.validate()
}
```

### Session Lifecycle

**[Visual: State diagram: Init -> Loading -> Running -> Collecting -> Reporting -> Done]**

**Narrator**: A session moves through six states. During Init, the orchestrator validates configuration and connects to target services. Loading reads the test bank and filters test cases by platform and priority. Running executes tests one at a time. Collecting gathers final evidence after all tests complete. Reporting generates the output. Done signals completion.

```go
// HelixQA/pkg/orchestrator/orchestrator.go
func (o *Orchestrator) Run(ctx context.Context) (*SessionResult, error) {
    session := o.initSession()

    // Load and filter test cases
    tests, err := o.testBank.Load(o.config.Platform, o.config.Priority)
    if err != nil {
        return nil, fmt.Errorf("load test bank: %w", err)
    }

    // Start crash/ANR detector in background
    o.detector.Start(ctx)
    defer o.detector.Stop()

    // Execute each test case
    for _, tc := range tests {
        select {
        case <-ctx.Done():
            session.Status = "cancelled"
            break
        default:
            result := o.executeTest(ctx, tc)
            session.Results = append(session.Results, result)
        }
    }

    // Generate report
    report, err := o.reporter.Generate(session)
    if err != nil {
        return nil, fmt.Errorf("generate report: %w", err)
    }
    session.ReportPath = report.Path

    return session, nil
}
```

**Narrator**: Notice that the detector runs in the background for the entire session. If a crash or ANR occurs during any test, the detector captures it immediately and associates it with the currently executing test case.

---

## Video 22.2: Test Bank Authoring in YAML (12 min)

### Test Bank Structure

**[Visual: File tree showing `testbank/` directory with YAML files organized by platform]**

**Narrator**: The test bank is a collection of YAML files that define test cases. Each file groups related tests. The directory structure organizes tests by platform and feature area.

```
testbank/
  api/
    auth.yaml
    scanning.yaml
    entities.yaml
    collections.yaml
  web/
    navigation.yaml
    entity_browser.yaml
    media_player.yaml
  desktop/
    installation.yaml
    file_management.yaml
  mobile/
    launch.yaml
    media_playback.yaml
```

### YAML Test Case Format

**[Visual: Open a YAML test case file]**

**Narrator**: Each test case has an ID, title, platform, priority, preconditions, steps, and expected outcomes. The format is designed to be human-readable and machine-executable.

```yaml
# testbank/api/entities.yaml
- id: ENT-001
  title: "Entity creation after scan"
  platform: api
  priority: critical
  tags: [entity, scan, aggregation]
  preconditions:
    - "Storage root configured with test media"
    - "Scan completed successfully"
  steps:
    - action: "GET /api/v1/entities?media_type=movie"
      expect:
        status: 200
        body:
          - "$.items | length > 0"
          - "$.items[0].title | not empty"
          - "$.items[0].media_type_id | equals 1"
    - action: "GET /api/v1/entities/{id}/children"
      expect:
        status: 200
    - action: "GET /api/v1/entities/{id}/files"
      expect:
        status: 200
        body:
          - "$.files | length > 0"
  on_failure:
    evidence: [api_response, server_logs]
    severity: blocker

- id: ENT-002
  title: "TV show hierarchy is three levels deep"
  platform: api
  priority: high
  tags: [entity, hierarchy, tv]
  steps:
    - action: "GET /api/v1/entities?media_type=tv_show"
      expect:
        status: 200
        body:
          - "$.items | length > 0"
      save:
        show_id: "$.items[0].id"
    - action: "GET /api/v1/entities/{show_id}/children"
      expect:
        status: 200
        body:
          - "$.items[0].media_type | equals tv_season"
      save:
        season_id: "$.items[0].id"
    - action: "GET /api/v1/entities/{season_id}/children"
      expect:
        status: 200
        body:
          - "$.items[0].media_type | equals tv_episode"
  on_failure:
    evidence: [api_response, database_dump]
    severity: critical
```

### Platform-Specific Extensions

**[Visual: Show a web test case with Playwright actions]**

**Narrator**: Web test cases use Playwright-specific actions like `navigate`, `click`, `fill`, `wait_for`, and `screenshot`. Mobile test cases use ADB commands. Desktop test cases use Tauri IPC.

```yaml
# testbank/web/entity_browser.yaml
- id: WEB-ENT-001
  title: "Entity browser loads and displays media grid"
  platform: web
  priority: critical
  steps:
    - action: navigate
      url: "/browse"
      expect:
        selector: "[data-testid='entity-grid']"
        visible: true
    - action: wait_for
      selector: "[data-testid='entity-card']"
      timeout: 10000
    - action: screenshot
      name: "entity_browser_loaded"
    - action: click
      selector: "[data-testid='entity-card']:first-child"
      expect:
        url_contains: "/entity/"
    - action: screenshot
      name: "entity_detail_page"
  on_failure:
    evidence: [screenshot, browser_console, network_log]
    severity: blocker
```

### Filtering and Priority

**[Visual: Terminal showing test bank loading with filters applied]**

**Narrator**: The test bank supports filtering by platform, priority, and tags. The `--priority` flag controls which tests run: `critical` runs only critical tests, `high` includes critical and high, and `all` runs everything.

```go
// HelixQA/pkg/testbank/testbank.go
func (tb *TestBank) Load(
    platform string,
    minPriority string,
) ([]TestCase, error) {
    var filtered []TestCase
    for _, tc := range tb.cases {
        if tc.Platform != platform {
            continue
        }
        if priorityRank(tc.Priority) < priorityRank(minPriority) {
            continue
        }
        filtered = append(filtered, tc)
    }
    sort.Slice(filtered, func(i, j int) bool {
        return priorityRank(filtered[i].Priority) >
               priorityRank(filtered[j].Priority)
    })
    return filtered, nil
}
```

---

## Video 22.3: Evidence Collection (10 min)

### Evidence Types

**[Visual: Evidence collection diagram showing screenshots, video, logs, and API responses flowing into the evidence store]**

**Narrator**: HelixQA collects four types of evidence: screenshots, video recordings, log files, and API response captures. Every test step can trigger evidence collection, and failures always capture evidence automatically.

```go
// HelixQA/pkg/evidence/collector.go
type Collector struct {
    outputDir string
    runner    CommandRunner
    logger    *zap.Logger
    mu        sync.Mutex
    artifacts []Artifact
}

type Artifact struct {
    TestCaseID string    `json:"test_case_id"`
    StepIndex  int       `json:"step_index"`
    Type       string    `json:"type"` // screenshot, video, log, response
    Path       string    `json:"path"`
    Timestamp  time.Time `json:"timestamp"`
    SizeBytes  int64     `json:"size_bytes"`
}
```

### Screenshot Collection

**[Visual: Show screenshots being captured during a web test]**

**Narrator**: Screenshots are captured via platform-specific commands. For web, Playwright's screenshot API. For Android, `adb shell screencap`. For desktop, `ffmpeg` with x11grab or platform-native capture.

```go
// HelixQA/pkg/evidence/collector.go
func (c *Collector) CaptureScreenshot(
    ctx context.Context,
    testCaseID string,
    stepIndex int,
    name string,
) (*Artifact, error) {
    filename := fmt.Sprintf("%s_step%d_%s.png", testCaseID, stepIndex, name)
    path := filepath.Join(c.outputDir, "screenshots", filename)

    if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
        return nil, fmt.Errorf("create screenshot dir: %w", err)
    }

    // Platform-specific capture (delegated to CommandRunner)
    if err := c.runner.Run(ctx, c.screenshotCommand(path)); err != nil {
        return nil, fmt.Errorf("capture screenshot: %w", err)
    }

    return c.recordArtifact(testCaseID, stepIndex, "screenshot", path)
}
```

### Video Recording

**[Visual: Show video recording being started and stopped around a test case]**

**Narrator**: Video recording captures the full visual flow of a test case. On Android 9 and below, `adb shell screenrecord` is used with `--bit-rate 4000000`. On Android 10+, rapid screenshot capture is assembled into video via ffmpeg. For web sessions, Playwright's video recording or ffmpeg x11grab captures the browser.

```go
// HelixQA/pkg/evidence/collector.go
func (c *Collector) StartVideoRecording(
    ctx context.Context,
    testCaseID string,
) error {
    filename := fmt.Sprintf("%s_session.mp4", testCaseID)
    path := filepath.Join(c.outputDir, "video-sessions", filename)

    cmd := c.videoRecordCommand(path)
    return c.runner.RunBackground(ctx, cmd)
}

func (c *Collector) StopVideoRecording(
    ctx context.Context,
    testCaseID string,
) (*Artifact, error) {
    if err := c.runner.Stop(ctx, "video-record"); err != nil {
        return nil, fmt.Errorf("stop video recording: %w", err)
    }

    path := filepath.Join(c.outputDir, "video-sessions",
        fmt.Sprintf("%s_session.mp4", testCaseID))
    return c.recordArtifact(testCaseID, 0, "video", path)
}
```

### Log Collection

**[Visual: Show server logs being captured during a test failure]**

**Narrator**: Log collection captures server-side output during test execution. For the API, this means the Gin request log and application log. For Android, `adb logcat` captures device logs. Logs are timestamped and correlated with the test case that was running when they were captured.

---

## Video 22.4: LLM-Driven Exploration and Curiosity Mode (8 min)

### Beyond Scripted Tests

**[Visual: Terminal showing HelixQA in curiosity mode, generating and executing novel test actions]**

**Narrator**: Scripted test cases cover known scenarios. But what about unknown bugs? HelixQA's curiosity mode uses an LLM to explore the application autonomously, generating test actions that no human wrote.

### How Curiosity Mode Works

**[Visual: Diagram showing LLM receiving app state -> generating action -> executing -> observing result -> repeating]**

**Narrator**: In curiosity mode, the LLM receives the current application state -- a screenshot, the DOM snapshot, or the API schema -- and generates the next action. After execution, it observes the result and decides whether to continue exploring the current path or try something different.

```go
// HelixQA/pkg/orchestrator/curiosity.go
type CuriosityEngine struct {
    provider   AIProvider
    evidence   *evidence.Collector
    detector   *detector.Detector
    maxSteps   int
    logger     *zap.Logger
}

func (e *CuriosityEngine) Explore(
    ctx context.Context,
    platform string,
) ([]ExplorationResult, error) {
    var results []ExplorationResult

    state := e.captureInitialState(ctx, platform)

    for step := 0; step < e.maxSteps; step++ {
        // Ask the LLM what to do next
        action, err := e.generateNextAction(ctx, state, results)
        if err != nil {
            e.logger.Warn("LLM action generation failed",
                zap.Int("step", step), zap.Error(err))
            continue
        }

        // Execute the action
        outcome, err := e.executeAction(ctx, action)
        if err != nil {
            results = append(results, ExplorationResult{
                Step:    step,
                Action:  action,
                Error:   err.Error(),
                Crashed: e.detector.HasNewCrash(),
            })
            continue
        }

        // Capture new state
        state = e.captureState(ctx, platform)

        results = append(results, ExplorationResult{
            Step:     step,
            Action:   action,
            Outcome:  outcome,
            Evidence: e.evidence.GetLatest(),
        })
    }

    return results, nil
}
```

**Narrator**: The LLM prompt includes the history of previous actions and outcomes, so the model learns from the session as it progresses. If it finds an error, it can probe deeper. If a path is exhausted, it backtracks to try a different feature.

### Safety Boundaries

**[Visual: Show safety rules preventing destructive actions]**

**Narrator**: Curiosity mode has safety boundaries. The LLM cannot delete data, modify system settings, or execute shell commands outside the application. A whitelist of allowed action types (navigate, click, fill, API GET/POST) constrains what the model can do.

---

## Video 22.5: Crash and ANR Detection (8 min)

### Real-Time Monitoring

**[Visual: Show the detector catching a crash during a test]**

**Narrator**: The detector runs as a background goroutine for the entire QA session. It monitors for crashes and Application Not Responding (ANR) events in real time, across all supported platforms.

```go
// HelixQA/pkg/detector/detector.go
type Detector struct {
    runners   map[string]PlatformDetector
    crashes   []CrashEvent
    mu        sync.RWMutex
    stopCh    chan struct{}
    wg        sync.WaitGroup
    logger    *zap.Logger
}

type CrashEvent struct {
    Platform    string    `json:"platform"`
    Type        string    `json:"type"` // crash, anr, oom
    Message     string    `json:"message"`
    StackTrace  string    `json:"stack_trace,omitempty"`
    TestCaseID  string    `json:"test_case_id"`
    Timestamp   time.Time `json:"timestamp"`
}
```

### Platform-Specific Detection

**[Visual: Show Android logcat being monitored for crash patterns]**

**Narrator**: On Android, the detector monitors `adb logcat` for fatal exceptions, ANR dialogs, and out-of-memory kills. It filters for package-specific crashes using the application's package name.

```go
// HelixQA/pkg/detector/android.go
func (d *AndroidDetector) Monitor(ctx context.Context) {
    cmd := d.runner.Command("adb", "logcat", "-v", "time",
        "--pid", d.pid)

    scanner := bufio.NewScanner(cmd.Stdout)
    for scanner.Scan() {
        line := scanner.Text()
        if d.isCrash(line) {
            d.reportCrash(ctx, "crash", line)
        } else if d.isANR(line) {
            d.reportCrash(ctx, "anr", line)
        }
    }
}
```

**Narrator**: For web, the detector monitors browser console errors and unhandled promise rejections via Playwright's console event listener. For desktop Tauri apps, it monitors the process exit code and stderr output.

### ANR Thresholds

**[Visual: Show ANR detection configuration]**

**Narrator**: ANR detection uses configurable thresholds. On Android, the standard threshold is 5 seconds of UI thread blockage. For web, HelixQA monitors long tasks exceeding 200 milliseconds and flags sequences of long tasks as potential ANRs.

---

## Video 22.6: Report Generation (10 min)

### Report Format

**[Visual: Show a generated HTML report with test results, evidence links, and crash details]**

**Narrator**: After all tests complete, the reporter generates a comprehensive QA report. It reuses the `challenges/pkg/report` package for consistent formatting across HelixQA and the challenge system.

```go
// HelixQA/pkg/reporter/reporter.go
type Reporter struct {
    format   string // html, markdown, json
    template *template.Template
    logger   *zap.Logger
}

type QAReport struct {
    SessionID    string            `json:"session_id"`
    Platform     string            `json:"platform"`
    StartTime    time.Time         `json:"start_time"`
    EndTime      time.Time         `json:"end_time"`
    Duration     time.Duration     `json:"duration"`
    TotalTests   int               `json:"total_tests"`
    Passed       int               `json:"passed"`
    Failed       int               `json:"failed"`
    Skipped      int               `json:"skipped"`
    CrashEvents  []CrashEvent      `json:"crash_events"`
    TestResults  []TestResult      `json:"test_results"`
    Artifacts    []Artifact        `json:"artifacts"`
    Summary      string            `json:"summary"`
}
```

### Report Contents

**[Visual: Walk through each section of the report]**

**Narrator**: The report has six sections:

1. **Executive Summary** -- pass rate, crash count, session duration, and a one-paragraph assessment.
2. **Test Results Table** -- each test case with status (pass/fail/skip), duration, and failure message.
3. **Crash Report** -- detailed crash events with stack traces and associated test cases.
4. **Evidence Gallery** -- links to screenshots, video recordings, and log files organized by test case.
5. **Trend Comparison** -- if previous reports exist, a comparison showing regressions and improvements.
6. **Ticket Suggestions** -- for each failure, a pre-formatted markdown ticket ready for the fix pipeline.

### Ticket Generation

**[Visual: Show a generated ticket with title, description, reproduction steps, and evidence links]**

**Narrator**: The ticket generator in `pkg/ticket` creates markdown-formatted issue descriptions for each failure. Each ticket includes a title, severity, reproduction steps extracted from the test case, the actual vs. expected outcome, and links to evidence files.

```go
// HelixQA/pkg/ticket/generator.go
type Ticket struct {
    Title       string   `json:"title"`
    Severity    string   `json:"severity"`
    Platform    string   `json:"platform"`
    TestCaseID  string   `json:"test_case_id"`
    Description string   `json:"description"`
    ReproSteps  []string `json:"repro_steps"`
    Expected    string   `json:"expected"`
    Actual      string   `json:"actual"`
    Evidence    []string `json:"evidence"`
}
```

**Narrator**: These tickets can be fed directly into an AI fix pipeline, where a language model reads the ticket, locates the relevant code, and proposes a fix. This closes the loop from detection to resolution.

---

## Key Code Examples

### Running a HelixQA Session
```bash
# Full API test suite, thorough speed, HTML report
helixqa run --platform api --speed thorough --report html \
    --output qa-results/

# Web tests, critical priority only, with video recording
helixqa run --platform web --priority critical --video \
    --output qa-results/

# Curiosity mode: LLM-driven exploration of the web UI
helixqa run --platform web --curiosity --max-steps 100 \
    --output qa-results/

# List available test cases
helixqa list --platform api --priority all
```

### CLI Entry Point
```go
// HelixQA/cmd/helixqa/main.go
func main() {
    app := &cli.App{
        Name:  "helixqa",
        Usage: "Autonomous QA orchestration framework",
        Commands: []*cli.Command{
            runCommand(),
            listCommand(),
            reportCommand(),
            versionCommand(),
        },
    }
    if err := app.Run(os.Args); err != nil {
        log.Fatal(err)
    }
}
```

---

## Key Files Referenced

- `HelixQA/pkg/orchestrator/orchestrator.go` -- Main QA brain
- `HelixQA/pkg/orchestrator/curiosity.go` -- LLM-driven exploration
- `HelixQA/pkg/testbank/testbank.go` -- YAML test bank loading and filtering
- `HelixQA/pkg/detector/detector.go` -- Crash/ANR detection coordinator
- `HelixQA/pkg/detector/android.go` -- Android-specific crash detection
- `HelixQA/pkg/evidence/collector.go` -- Screenshot, video, log collection
- `HelixQA/pkg/validator/validator.go` -- Step-by-step assertion engine
- `HelixQA/pkg/reporter/reporter.go` -- Report generation
- `HelixQA/pkg/ticket/generator.go` -- Failure ticket generation
- `HelixQA/cmd/helixqa/main.go` -- CLI entry point

---

## Exercises

1. Write a YAML test case that verifies the duplicate detection API returns results after scanning two storage roots with overlapping content.
2. Implement a `WebDetector` that monitors Playwright browser console for unhandled promise rejections and JavaScript errors.
3. Extend the curiosity engine to maintain a "coverage map" of visited pages and API endpoints, prioritizing unexplored areas.
4. Write a table-driven test for the test bank priority filtering: given a set of test cases with mixed priorities, verify that `Load("api", "high")` returns only critical and high tests in the correct order.

---

## Quiz Questions

1. What are the five subsystems that the orchestrator composes?
   **Answer**: Test bank (what to test), detector (crash/ANR monitoring), validator (step-by-step assertions), evidence collector (screenshots, video, logs), and reporter (output generation).

2. How does curiosity mode differ from scripted test execution?
   **Answer**: Scripted tests follow predefined YAML steps. Curiosity mode uses an LLM to observe the current application state and generate the next action autonomously. It can discover bugs that no human anticipated. Safety boundaries constrain it to non-destructive actions.

3. What types of evidence does HelixQA collect, and when?
   **Answer**: Four types: screenshots, video recordings, log files, and API response captures. Evidence is collected at explicit step directives in the YAML, automatically on every failure, and continuously for video and log monitoring.

4. How does the Android crash detector work?
   **Answer**: It monitors `adb logcat` output filtered by the application's process ID. It pattern-matches log lines for fatal exceptions, ANR dialog messages, and out-of-memory kills, reporting each as a `CrashEvent` with the associated test case ID and stack trace.
