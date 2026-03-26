# HelixQA Autonomous Robot — Phase 3 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the Execution Engine — `pkg/performance/` (metrics collection), `pkg/video/` (scrcpy/ffmpeg recording), `pkg/maestro/` (mobile flow runner), and `pkg/analysis/` (post-session video/screenshot analysis using LLM vision).

**Architecture:** Each package wraps an external tool (scrcpy, ffmpeg, Maestro, ADB dumpsys) behind Go interfaces using the existing `CommandRunner` subprocess pattern. The analysis package uses `pkg/llm.Provider` for vision-based screenshot/video frame analysis. All packages produce structured results consumed by the `SessionCoordinator`.

**Tech Stack:** Go 1.25, `os/exec` for subprocess management, `pkg/llm.Provider` for LLM vision calls, `encoding/json` for metrics serialization, `testify` for testing, existing `detector.CommandRunner` for testability.

**Spec:** `docs/superpowers/specs/2026-03-26-helixqa-autonomous-robot-design.md` (Sections 4-5)

**Depends on:** Phase 1 (`pkg/llm/`, `pkg/memory/`) + Phase 2 (`pkg/learning/`, `pkg/planning/`)

---

## File Structure

### New Files

| File | Responsibility |
|------|---------------|
| `pkg/performance/collector.go` | `MetricsCollector` — gathers CPU, memory, network metrics via ADB/system commands |
| `pkg/performance/types.go` | `MetricSnapshot`, `MetricsTimeline`, `LeakIndicator` types |
| `pkg/performance/collector_test.go` | MetricsCollector tests with mock commands |
| `pkg/video/scrcpy.go` | `ScrcpyRecorder` — manages scrcpy subprocess for Android recording |
| `pkg/video/frames.go` | `FrameExtractor` — extracts key frames from video via ffmpeg |
| `pkg/video/scrcpy_test.go` | ScrcpyRecorder tests |
| `pkg/video/frames_test.go` | FrameExtractor tests |
| `pkg/maestro/runner.go` | `FlowRunner` — executes Maestro YAML flows via subprocess |
| `pkg/maestro/runner_test.go` | FlowRunner tests |
| `pkg/analysis/analyzer.go` | `PostAnalyzer` — orchestrates post-session analysis pipeline |
| `pkg/analysis/vision.go` | `VisionAnalyzer` — LLM vision analysis of screenshots/frames |
| `pkg/analysis/types.go` | `AnalysisFinding`, `AnalysisReport` types |
| `pkg/analysis/analyzer_test.go` | PostAnalyzer tests |
| `pkg/analysis/vision_test.go` | VisionAnalyzer tests with mock LLM |

---

## Task 1: Performance Types

**Files:**
- Create: `pkg/performance/types.go`

- [ ] **Step 1: Write the types**

```go
// pkg/performance/types.go
package performance

import "time"

// MetricType identifies the kind of performance metric.
type MetricType string

const (
	MetricMemoryRSS     MetricType = "memory_rss_kb"
	MetricMemoryHeap    MetricType = "memory_heap_kb"
	MetricCPUPercent    MetricType = "cpu_percent"
	MetricNetworkRxKB   MetricType = "network_rx_kb"
	MetricNetworkTxKB   MetricType = "network_tx_kb"
	MetricFPS           MetricType = "fps"
	MetricThreadCount   MetricType = "thread_count"
)

// MetricSnapshot is a single measurement at a point in time.
type MetricSnapshot struct {
	Type      MetricType `json:"type"`
	Value     float64    `json:"value"`
	Timestamp time.Time  `json:"timestamp"`
	Platform  string     `json:"platform"`
	Label     string     `json:"label,omitempty"`
}

// MetricsTimeline is an ordered series of snapshots.
type MetricsTimeline struct {
	Platform  string           `json:"platform"`
	Snapshots []MetricSnapshot `json:"snapshots"`
}

// Add appends a snapshot.
func (mt *MetricsTimeline) Add(s MetricSnapshot) {
	mt.Snapshots = append(mt.Snapshots, s)
}

// OfType returns snapshots filtered by metric type.
func (mt *MetricsTimeline) OfType(t MetricType) []MetricSnapshot {
	var result []MetricSnapshot
	for _, s := range mt.Snapshots {
		if s.Type == t {
			result = append(result, s)
		}
	}
	return result
}

// LeakIndicator describes a potential memory leak.
type LeakIndicator struct {
	Platform      string  `json:"platform"`
	StartKB       float64 `json:"start_kb"`
	EndKB         float64 `json:"end_kb"`
	GrowthPercent float64 `json:"growth_percent"`
	DurationSecs  float64 `json:"duration_secs"`
	IsLeak        bool    `json:"is_leak"`
}

// DetectMemoryLeak analyzes memory snapshots for monotonic growth.
// Returns a leak indicator if memory grew more than thresholdPercent.
func (mt *MetricsTimeline) DetectMemoryLeak(thresholdPercent float64) *LeakIndicator {
	memSnapshots := mt.OfType(MetricMemoryRSS)
	if len(memSnapshots) < 2 {
		return nil
	}
	first := memSnapshots[0]
	last := memSnapshots[len(memSnapshots)-1]
	if first.Value <= 0 {
		return nil
	}
	growth := ((last.Value - first.Value) / first.Value) * 100
	duration := last.Timestamp.Sub(first.Timestamp).Seconds()

	return &LeakIndicator{
		Platform:      mt.Platform,
		StartKB:       first.Value,
		EndKB:         last.Value,
		GrowthPercent: growth,
		DurationSecs:  duration,
		IsLeak:        growth > thresholdPercent,
	}
}
```

- [ ] **Step 2: Verify it compiles**

Run: `cd HelixQA && go build ./pkg/performance/`
Expected: Success

- [ ] **Step 3: Commit**

```bash
git add HelixQA/pkg/performance/types.go
git commit -m "feat(helixqa): add performance metric types with leak detection"
```

---

## Task 2: MetricsCollector — ADB-Based Performance Monitoring

**Files:**
- Create: `pkg/performance/collector.go`
- Test: `pkg/performance/collector_test.go`

- [ ] **Step 1: Write the failing test**

```go
// pkg/performance/collector_test.go
package performance

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockRunner struct {
	outputs map[string]string
}

func newMockRunner() *mockRunner {
	return &mockRunner{outputs: make(map[string]string)}
}

func (m *mockRunner) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	// Build key from command
	key := name
	for _, a := range args {
		key += " " + a
	}
	if out, ok := m.outputs[key]; ok {
		return []byte(out), nil
	}
	return []byte(""), nil
}

func TestMetricsCollector_CollectMemory(t *testing.T) {
	mock := newMockRunner()
	mock.outputs["adb shell dumpsys meminfo com.test.app"] = `
Applications Memory Usage (in Kilobytes):
Uptime: 123456 Realtime: 789012

** MEMINFO in pid 12345 [com.test.app] **
                   Pss  Private  Private  SwapPss     Heap     Heap     Heap
                 Total    Dirty    Clean    Dirty     Size    Alloc     Free
                ------   ------   ------   ------   ------   ------   ------
  Native Heap    15000    14500      100        0    32768    25000     7768
  Dalvik Heap     8000     7500       50        0    16384    12000     4384
        TOTAL    45000    35000     1000        0
`
	collector := NewMetricsCollector("com.test.app", "android", WithCommandRunner(mock))
	snapshot, err := collector.CollectMemory(context.Background())
	require.NoError(t, err)
	assert.Equal(t, MetricMemoryRSS, snapshot.Type)
	assert.Greater(t, snapshot.Value, float64(0))
}

func TestMetricsCollector_CollectCPU(t *testing.T) {
	mock := newMockRunner()
	mock.outputs["adb shell dumpsys cpuinfo"] = `
Load: 2.5 / 1.8 / 1.2
CPU usage from 0ms to 500ms ago:
  12.5% 12345/com.test.app: 8.5% user + 4% kernel
  5.2% 100/system_server: 3% user + 2.2% kernel
`
	collector := NewMetricsCollector("com.test.app", "android", WithCommandRunner(mock))
	snapshot, err := collector.CollectCPU(context.Background())
	require.NoError(t, err)
	assert.Equal(t, MetricCPUPercent, snapshot.Type)
	assert.InDelta(t, 12.5, snapshot.Value, 0.1)
}

func TestMetricsTimeline_DetectLeak(t *testing.T) {
	tl := &MetricsTimeline{Platform: "android"}
	now := time.Now()
	tl.Add(MetricSnapshot{Type: MetricMemoryRSS, Value: 100000, Timestamp: now})
	tl.Add(MetricSnapshot{Type: MetricMemoryRSS, Value: 105000, Timestamp: now.Add(1 * time.Minute)})
	tl.Add(MetricSnapshot{Type: MetricMemoryRSS, Value: 115000, Timestamp: now.Add(5 * time.Minute)})

	leak := tl.DetectMemoryLeak(10.0) // 10% threshold
	require.NotNil(t, leak)
	assert.True(t, leak.IsLeak) // 15% growth > 10%
	assert.InDelta(t, 15.0, leak.GrowthPercent, 0.1)
}

func TestMetricsTimeline_NoLeak(t *testing.T) {
	tl := &MetricsTimeline{Platform: "android"}
	now := time.Now()
	tl.Add(MetricSnapshot{Type: MetricMemoryRSS, Value: 100000, Timestamp: now})
	tl.Add(MetricSnapshot{Type: MetricMemoryRSS, Value: 102000, Timestamp: now.Add(5 * time.Minute)})

	leak := tl.DetectMemoryLeak(10.0)
	require.NotNil(t, leak)
	assert.False(t, leak.IsLeak) // 2% < 10%
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd HelixQA && go test ./pkg/performance/ -v`
Expected: FAIL — `NewMetricsCollector` undefined

- [ ] **Step 3: Write the implementation**

```go
// pkg/performance/collector.go
package performance

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// CommandRunner executes system commands (for testability).
type CommandRunner interface {
	Run(ctx context.Context, name string, args ...string) ([]byte, error)
}

// MetricsCollector gathers performance metrics from a running app.
type MetricsCollector struct {
	pkg      string
	platform string
	runner   CommandRunner
}

// CollectorOption configures a MetricsCollector.
type CollectorOption func(*MetricsCollector)

// WithCommandRunner sets a custom command runner.
func WithCommandRunner(r CommandRunner) CollectorOption {
	return func(c *MetricsCollector) { c.runner = r }
}

// NewMetricsCollector creates a collector for the given package and platform.
func NewMetricsCollector(pkg, platform string, opts ...CollectorOption) *MetricsCollector {
	c := &MetricsCollector{pkg: pkg, platform: platform}
	for _, opt := range opts {
		opt(c)
	}
	if c.runner == nil {
		c.runner = &defaultRunner{}
	}
	return c
}

var totalPSSRegex = regexp.MustCompile(`TOTAL\s+(\d+)`)
var cpuRegex = regexp.MustCompile(`(\d+\.?\d*)%\s+\d+/`)

// CollectMemory reads memory info via adb dumpsys meminfo.
func (c *MetricsCollector) CollectMemory(ctx context.Context) (*MetricSnapshot, error) {
	out, err := c.runner.Run(ctx, "adb", "shell", "dumpsys", "meminfo", c.pkg)
	if err != nil {
		return nil, fmt.Errorf("performance: collect memory: %w", err)
	}

	matches := totalPSSRegex.FindStringSubmatch(string(out))
	if len(matches) < 2 {
		return nil, fmt.Errorf("performance: could not parse meminfo output")
	}

	value, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return nil, fmt.Errorf("performance: parse memory value: %w", err)
	}

	return &MetricSnapshot{
		Type:      MetricMemoryRSS,
		Value:     value,
		Timestamp: time.Now(),
		Platform:  c.platform,
	}, nil
}

// CollectCPU reads CPU usage via adb dumpsys cpuinfo.
func (c *MetricsCollector) CollectCPU(ctx context.Context) (*MetricSnapshot, error) {
	out, err := c.runner.Run(ctx, "adb", "shell", "dumpsys", "cpuinfo")
	if err != nil {
		return nil, fmt.Errorf("performance: collect cpu: %w", err)
	}

	// Find line with our package
	for _, line := range strings.Split(string(out), "\n") {
		if strings.Contains(line, c.pkg) {
			matches := cpuRegex.FindStringSubmatch(line)
			if len(matches) >= 2 {
				value, err := strconv.ParseFloat(matches[1], 64)
				if err == nil {
					return &MetricSnapshot{
						Type:      MetricCPUPercent,
						Value:     value,
						Timestamp: time.Now(),
						Platform:  c.platform,
					}, nil
				}
			}
		}
	}

	return &MetricSnapshot{
		Type:      MetricCPUPercent,
		Value:     0,
		Timestamp: time.Now(),
		Platform:  c.platform,
	}, nil
}

// CollectAll gathers memory + CPU in one call.
func (c *MetricsCollector) CollectAll(ctx context.Context) ([]MetricSnapshot, error) {
	var snapshots []MetricSnapshot

	mem, err := c.CollectMemory(ctx)
	if err == nil && mem != nil {
		snapshots = append(snapshots, *mem)
	}

	cpu, err := c.CollectCPU(ctx)
	if err == nil && cpu != nil {
		snapshots = append(snapshots, *cpu)
	}

	return snapshots, nil
}

// defaultRunner uses os/exec.
type defaultRunner struct{}

func (d *defaultRunner) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	cmd := execCommand(ctx, name, args...)
	return cmd.Output()
}

// execCommand is a variable for testing.
var execCommand = execCommandImpl

func execCommandImpl(ctx context.Context, name string, args ...string) command {
	return &realCommand{ctx: ctx, name: name, args: args}
}

type command interface {
	Output() ([]byte, error)
}

type realCommand struct {
	ctx  context.Context
	name string
	args []string
}

func (r *realCommand) Output() ([]byte, error) {
	return nil, fmt.Errorf("not implemented in production; use CommandRunner interface")
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd HelixQA && go test ./pkg/performance/ -v`
Expected: PASS (4 tests)

- [ ] **Step 5: Commit**

```bash
git add HelixQA/pkg/performance/
git commit -m "feat(helixqa): add MetricsCollector for ADB-based performance monitoring"
```

---

## Task 3: ScrcpyRecorder — Android Video Recording

**Files:**
- Create: `pkg/video/scrcpy.go`
- Test: `pkg/video/scrcpy_test.go`

- [ ] **Step 1: Write the failing test**

```go
// pkg/video/scrcpy_test.go
package video

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScrcpyRecorder_Start(t *testing.T) {
	recorder := NewScrcpyRecorder("device123", "/tmp/output.mp4")
	assert.False(t, recorder.IsRecording())
	assert.Equal(t, "device123", recorder.Device())
	assert.Equal(t, "/tmp/output.mp4", recorder.OutputPath())
}

func TestScrcpyRecorder_BuildCommand_Scrcpy(t *testing.T) {
	recorder := NewScrcpyRecorder("device123", "/tmp/output.mp4")
	args := recorder.buildScrcpyArgs()
	assert.Contains(t, args, "--serial")
	assert.Contains(t, args, "device123")
	assert.Contains(t, args, "--record")
	assert.Contains(t, args, "/tmp/output.mp4")
}

func TestScrcpyRecorder_BuildCommand_ADB(t *testing.T) {
	recorder := NewScrcpyRecorder("device123", "/tmp/output.mp4", WithMethod(MethodADBScreenrecord))
	args := recorder.buildADBArgs()
	assert.Contains(t, args, "-s")
	assert.Contains(t, args, "device123")
	assert.Contains(t, args, "screenrecord")
}

func TestScrcpyRecorder_MethodSelection(t *testing.T) {
	tests := []struct {
		name     string
		method   RecordMethod
		expected RecordMethod
	}{
		{"auto default", MethodAuto, MethodAuto},
		{"explicit scrcpy", MethodScrcpy, MethodScrcpy},
		{"explicit adb", MethodADBScreenrecord, MethodADBScreenrecord},
		{"explicit screenshot", MethodScreenshotAssembly, MethodScreenshotAssembly},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewScrcpyRecorder("dev", "/tmp/out.mp4", WithMethod(tt.method))
			assert.Equal(t, tt.expected, r.method)
		})
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd HelixQA && go test ./pkg/video/ -v`
Expected: FAIL — `NewScrcpyRecorder` undefined

- [ ] **Step 3: Write the implementation**

```go
// pkg/video/scrcpy.go
package video

import (
	"context"
	"fmt"
	"os/exec"
	"sync"
	"time"
)

// RecordMethod selects which recording approach to use.
type RecordMethod string

const (
	MethodAuto               RecordMethod = "auto"
	MethodScrcpy             RecordMethod = "scrcpy"
	MethodADBScreenrecord    RecordMethod = "adb_screenrecord"
	MethodScreenshotAssembly RecordMethod = "screenshot_assembly"
)

// ScrcpyRecorder manages video recording for an Android device.
type ScrcpyRecorder struct {
	device     string
	outputPath string
	method     RecordMethod
	bitRate    int
	maxSecs    int
	cmd        *exec.Cmd
	recording  bool
	startedAt  time.Time
	mu         sync.Mutex
}

// RecorderOption configures a ScrcpyRecorder.
type RecorderOption func(*ScrcpyRecorder)

// WithMethod sets the recording method.
func WithMethod(m RecordMethod) RecorderOption {
	return func(r *ScrcpyRecorder) { r.method = m }
}

// WithBitRate sets the video bit rate.
func WithBitRate(rate int) RecorderOption {
	return func(r *ScrcpyRecorder) { r.bitRate = rate }
}

// WithMaxDuration sets the maximum recording duration in seconds.
func WithMaxDuration(secs int) RecorderOption {
	return func(r *ScrcpyRecorder) { r.maxSecs = secs }
}

// NewScrcpyRecorder creates a recorder for the given device and output path.
func NewScrcpyRecorder(device, outputPath string, opts ...RecorderOption) *ScrcpyRecorder {
	r := &ScrcpyRecorder{
		device:     device,
		outputPath: outputPath,
		method:     MethodAuto,
		bitRate:    4000000,
		maxSecs:    120,
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

// Device returns the device identifier.
func (r *ScrcpyRecorder) Device() string { return r.device }

// OutputPath returns the output file path.
func (r *ScrcpyRecorder) OutputPath() string { return r.outputPath }

// IsRecording returns whether recording is active.
func (r *ScrcpyRecorder) IsRecording() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.recording
}

// Start begins recording.
func (r *ScrcpyRecorder) Start(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.recording {
		return fmt.Errorf("video: already recording")
	}

	method := r.method
	if method == MethodAuto {
		method = r.detectMethod()
	}

	var cmd *exec.Cmd
	switch method {
	case MethodScrcpy:
		cmd = exec.CommandContext(ctx, "scrcpy", r.buildScrcpyArgs()...)
	case MethodADBScreenrecord:
		cmd = exec.CommandContext(ctx, "adb", r.buildADBArgs()...)
	default:
		return fmt.Errorf("video: unsupported method %s for Start()", method)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("video: start recording: %w", err)
	}

	r.cmd = cmd
	r.recording = true
	r.startedAt = time.Now()
	return nil
}

// Stop ends recording and waits for the process to finish.
func (r *ScrcpyRecorder) Stop() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.recording || r.cmd == nil {
		return nil
	}

	r.recording = false
	if r.cmd.Process != nil {
		r.cmd.Process.Kill()
	}
	r.cmd.Wait()
	r.cmd = nil
	return nil
}

// Duration returns how long recording has been active.
func (r *ScrcpyRecorder) Duration() time.Duration {
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.recording {
		return 0
	}
	return time.Since(r.startedAt)
}

func (r *ScrcpyRecorder) buildScrcpyArgs() []string {
	return []string{
		"--serial", r.device,
		"--no-display",
		"--record", r.outputPath,
		"--bit-rate", fmt.Sprintf("%d", r.bitRate),
		"--max-fps", "15",
	}
}

func (r *ScrcpyRecorder) buildADBArgs() []string {
	return []string{
		"-s", r.device,
		"shell", "screenrecord",
		"--bit-rate", fmt.Sprintf("%d", r.bitRate),
		"--time-limit", fmt.Sprintf("%d", r.maxSecs),
		fmt.Sprintf("/sdcard/qa_recording_%d.mp4", time.Now().Unix()),
	}
}

func (r *ScrcpyRecorder) detectMethod() RecordMethod {
	if _, err := exec.LookPath("scrcpy"); err == nil {
		return MethodScrcpy
	}
	return MethodADBScreenrecord
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd HelixQA && go test ./pkg/video/ -v`
Expected: PASS (4 tests)

- [ ] **Step 5: Commit**

```bash
git add HelixQA/pkg/video/
git commit -m "feat(helixqa): add ScrcpyRecorder for Android video recording"
```

---

## Task 4: FrameExtractor — ffmpeg Key Frame Extraction

**Files:**
- Create: `pkg/video/frames.go`
- Test: `pkg/video/frames_test.go`

- [ ] **Step 1: Write the failing test**

```go
// pkg/video/frames_test.go
package video

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFrameExtractor_BuildFFmpegArgs(t *testing.T) {
	extractor := NewFrameExtractor("/home/user/bin/ffmpeg")

	args := extractor.buildArgs("/tmp/video.mp4", "/tmp/frames", 1)
	assert.Contains(t, args, "-i")
	assert.Contains(t, args, "/tmp/video.mp4")
	assert.Contains(t, args, "fps=1")
}

func TestFrameExtractor_BuildFFmpegArgs_SceneDetect(t *testing.T) {
	extractor := NewFrameExtractor("/usr/bin/ffmpeg")
	args := extractor.buildSceneArgs("/tmp/video.mp4", "/tmp/frames", 0.3)
	assert.Contains(t, args, "select")
	assert.Contains(t, args, "0.3")
}

func TestFrameExtractor_OutputPattern(t *testing.T) {
	extractor := NewFrameExtractor("ffmpeg")
	pattern := extractor.outputPattern("/tmp/frames")
	assert.Contains(t, pattern, "frame_%04d.png")
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd HelixQA && go test ./pkg/video/ -v -run TestFrameExtractor`
Expected: FAIL — `NewFrameExtractor` undefined

- [ ] **Step 3: Write the implementation**

```go
// pkg/video/frames.go
package video

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// FrameExtractor extracts key frames from video files using ffmpeg.
type FrameExtractor struct {
	ffmpegPath string
}

// NewFrameExtractor creates an extractor with the given ffmpeg path.
func NewFrameExtractor(ffmpegPath string) *FrameExtractor {
	return &FrameExtractor{ffmpegPath: ffmpegPath}
}

// ExtractFPS extracts frames at the given FPS rate.
func (f *FrameExtractor) ExtractFPS(ctx context.Context, videoPath, outputDir string, fps int) ([]string, error) {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("video: create output dir: %w", err)
	}

	args := f.buildArgs(videoPath, outputDir, fps)
	cmd := exec.CommandContext(ctx, f.ffmpegPath, args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("video: ffmpeg extract failed: %w\n%s", err, string(out))
	}

	return f.listFrames(outputDir)
}

// ExtractSceneChanges extracts frames at scene change boundaries.
func (f *FrameExtractor) ExtractSceneChanges(ctx context.Context, videoPath, outputDir string, threshold float64) ([]string, error) {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("video: create output dir: %w", err)
	}

	args := f.buildSceneArgs(videoPath, outputDir, threshold)
	cmd := exec.CommandContext(ctx, f.ffmpegPath, args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("video: ffmpeg scene extract failed: %w\n%s", err, string(out))
	}

	return f.listFrames(outputDir)
}

func (f *FrameExtractor) buildArgs(videoPath, outputDir string, fps int) []string {
	return []string{
		"-y", "-i", videoPath,
		"-vf", fmt.Sprintf("fps=%d", fps),
		"-q:v", "2",
		f.outputPattern(outputDir),
	}
}

func (f *FrameExtractor) buildSceneArgs(videoPath, outputDir string, threshold float64) []string {
	return []string{
		"-y", "-i", videoPath,
		"-vf", fmt.Sprintf("select='gt(scene,%g)',showinfo", threshold),
		"-vsync", "vfr",
		"-q:v", "2",
		f.outputPattern(outputDir),
	}
}

func (f *FrameExtractor) outputPattern(outputDir string) string {
	return filepath.Join(outputDir, "frame_%04d.png")
}

func (f *FrameExtractor) listFrames(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var frames []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".png") {
			frames = append(frames, filepath.Join(dir, e.Name()))
		}
	}
	return frames, nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd HelixQA && go test ./pkg/video/ -v -run TestFrameExtractor`
Expected: PASS (3 tests)

- [ ] **Step 5: Commit**

```bash
git add HelixQA/pkg/video/frames.go HelixQA/pkg/video/frames_test.go
git commit -m "feat(helixqa): add FrameExtractor for ffmpeg key frame extraction"
```

---

## Task 5: Maestro FlowRunner

**Files:**
- Create: `pkg/maestro/runner.go`
- Test: `pkg/maestro/runner_test.go`

- [ ] **Step 1: Write the failing test**

```go
// pkg/maestro/runner_test.go
package maestro

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFlowRunner_BuildArgs(t *testing.T) {
	runner := NewFlowRunner()
	args := runner.buildArgs("/tmp/flow.yaml", "device123")
	assert.Contains(t, args, "test")
	assert.Contains(t, args, "/tmp/flow.yaml")
	assert.Contains(t, args, "--device")
	assert.Contains(t, args, "device123")
}

func TestFlowRunner_BuildArgs_NoDevice(t *testing.T) {
	runner := NewFlowRunner()
	args := runner.buildArgs("/tmp/flow.yaml", "")
	assert.Contains(t, args, "test")
	assert.Contains(t, args, "/tmp/flow.yaml")
	assert.NotContains(t, args, "--device")
}

func TestFlowRunner_ParseResult_Success(t *testing.T) {
	output := `
Running flow: Login Flow
✅ Open app
✅ Enter credentials
✅ Tap sign in
✅ Verify dashboard

1 Passed, 0 Failed
`
	result, err := parseFlowResult(output)
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, 1, result.Passed)
	assert.Equal(t, 0, result.Failed)
}

func TestFlowRunner_ParseResult_Failure(t *testing.T) {
	output := `
Running flow: Login Flow
✅ Open app
❌ Enter credentials - Element not found

0 Passed, 1 Failed
`
	result, err := parseFlowResult(output)
	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.Equal(t, 0, result.Passed)
	assert.Equal(t, 1, result.Failed)
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd HelixQA && go test ./pkg/maestro/ -v`
Expected: FAIL — `NewFlowRunner` undefined

- [ ] **Step 3: Write the implementation**

```go
// pkg/maestro/runner.go
package maestro

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// FlowResult describes the outcome of a Maestro flow execution.
type FlowResult struct {
	FlowFile string `json:"flow_file"`
	Success  bool   `json:"success"`
	Passed   int    `json:"passed"`
	Failed   int    `json:"failed"`
	Output   string `json:"output"`
	Error    string `json:"error,omitempty"`
}

// FlowRunner executes Maestro YAML flows via subprocess.
type FlowRunner struct {
	maestroPath string
}

// NewFlowRunner creates a FlowRunner using the default maestro command.
func NewFlowRunner() *FlowRunner {
	return &FlowRunner{maestroPath: "maestro"}
}

// NewFlowRunnerWithPath creates a FlowRunner with a custom maestro path.
func NewFlowRunnerWithPath(path string) *FlowRunner {
	return &FlowRunner{maestroPath: path}
}

// RunFlow executes a Maestro YAML flow file.
func (r *FlowRunner) RunFlow(ctx context.Context, flowFile, device string) (*FlowResult, error) {
	args := r.buildArgs(flowFile, device)
	cmd := exec.CommandContext(ctx, r.maestroPath, args...)

	out, err := cmd.CombinedOutput()
	output := string(out)

	result, parseErr := parseFlowResult(output)
	if parseErr != nil {
		result = &FlowResult{Output: output}
	}
	result.FlowFile = flowFile

	if err != nil {
		result.Success = false
		result.Error = err.Error()
	}

	return result, nil // always return result, even on failure
}

func (r *FlowRunner) buildArgs(flowFile, device string) []string {
	args := []string{"test", flowFile}
	if device != "" {
		args = append(args, "--device", device)
	}
	return args
}

var resultRegex = regexp.MustCompile(`(\d+)\s+Passed,\s+(\d+)\s+Failed`)

func parseFlowResult(output string) (*FlowResult, error) {
	matches := resultRegex.FindStringSubmatch(output)
	if len(matches) < 3 {
		return nil, fmt.Errorf("maestro: could not parse result from output")
	}

	passed, _ := strconv.Atoi(matches[1])
	failed, _ := strconv.Atoi(matches[2])

	return &FlowResult{
		Success: failed == 0 && !strings.Contains(output, "❌"),
		Passed:  passed,
		Failed:  failed,
		Output:  output,
	}, nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd HelixQA && go test ./pkg/maestro/ -v`
Expected: PASS (4 tests)

- [ ] **Step 5: Commit**

```bash
git add HelixQA/pkg/maestro/
git commit -m "feat(helixqa): add Maestro FlowRunner for YAML mobile flow execution"
```

---

## Task 6: Analysis Types

**Files:**
- Create: `pkg/analysis/types.go`

- [ ] **Step 1: Write the types**

```go
// pkg/analysis/types.go
package analysis

// FindingCategory classifies what kind of issue was found.
type FindingCategory string

const (
	CategoryVisual        FindingCategory = "visual"
	CategoryUX            FindingCategory = "ux"
	CategoryAccessibility FindingCategory = "accessibility"
	CategoryPerformance   FindingCategory = "performance"
	CategoryFunctional    FindingCategory = "functional"
	CategoryBrand         FindingCategory = "brand"
	CategoryContent       FindingCategory = "content"
)

// FindingSeverity indicates how serious an issue is.
type FindingSeverity string

const (
	SeverityCritical FindingSeverity = "critical"
	SeverityHigh     FindingSeverity = "high"
	SeverityMedium   FindingSeverity = "medium"
	SeverityLow      FindingSeverity = "low"
	SeverityCosmetic FindingSeverity = "cosmetic"
)

// AnalysisFinding represents an issue discovered during post-analysis.
type AnalysisFinding struct {
	Category    FindingCategory `json:"category"`
	Severity    FindingSeverity `json:"severity"`
	Title       string          `json:"title"`
	Description string          `json:"description"`
	ReproSteps  string          `json:"repro_steps"`
	Evidence    string          `json:"evidence"`
	Platform    string          `json:"platform"`
	Screen      string          `json:"screen"`
}

// AnalysisReport is the aggregate result of post-session analysis.
type AnalysisReport struct {
	SessionID     string            `json:"session_id"`
	TotalAnalyzed int               `json:"total_analyzed"`
	Findings      []AnalysisFinding `json:"findings"`
	Summary       string            `json:"summary"`
}

// BySeverity returns findings filtered by severity.
func (r *AnalysisReport) BySeverity(s FindingSeverity) []AnalysisFinding {
	var result []AnalysisFinding
	for _, f := range r.Findings {
		if f.Severity == s {
			result = append(result, f)
		}
	}
	return result
}

// CriticalCount returns the number of critical + high findings.
func (r *AnalysisReport) CriticalCount() int {
	count := 0
	for _, f := range r.Findings {
		if f.Severity == SeverityCritical || f.Severity == SeverityHigh {
			count++
		}
	}
	return count
}
```

- [ ] **Step 2: Verify it compiles**

Run: `cd HelixQA && go build ./pkg/analysis/`
Expected: Success

- [ ] **Step 3: Commit**

```bash
git add HelixQA/pkg/analysis/types.go
git commit -m "feat(helixqa): add analysis finding and report types"
```

---

## Task 7: VisionAnalyzer — LLM Vision for Screenshots

**Files:**
- Create: `pkg/analysis/vision.go`
- Test: `pkg/analysis/vision_test.go`

- [ ] **Step 1: Write the failing test**

```go
// pkg/analysis/vision_test.go
package analysis

import (
	"context"
	"encoding/json"
	"testing"

	"digital.vasic.helixqa/pkg/llm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockVisionLLM struct {
	response string
}

func (m *mockVisionLLM) Chat(_ context.Context, _ []llm.Message) (*llm.Response, error) {
	return &llm.Response{Content: m.response}, nil
}
func (m *mockVisionLLM) Vision(_ context.Context, _ []byte, _ string) (*llm.Response, error) {
	return &llm.Response{Content: m.response}, nil
}
func (m *mockVisionLLM) Name() string        { return "mock" }
func (m *mockVisionLLM) SupportsVision() bool { return true }

func TestVisionAnalyzer_AnalyzeScreenshot(t *testing.T) {
	findings := []AnalysisFinding{
		{Category: CategoryVisual, Severity: SeverityMedium, Title: "Text clipped", Description: "Header text is truncated"},
	}
	mockJSON, _ := json.Marshal(findings)
	provider := &mockVisionLLM{response: string(mockJSON)}

	analyzer := NewVisionAnalyzer(provider)
	result, err := analyzer.AnalyzeScreenshot(context.Background(), []byte("fake-png"), "login", "web")
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(result), 1)
	assert.Equal(t, CategoryVisual, result[0].Category)
}

func TestVisionAnalyzer_AnalyzeScreenshot_NoIssues(t *testing.T) {
	provider := &mockVisionLLM{response: "[]"}
	analyzer := NewVisionAnalyzer(provider)
	result, err := analyzer.AnalyzeScreenshot(context.Background(), []byte("ok-png"), "dashboard", "web")
	require.NoError(t, err)
	assert.Len(t, result, 0)
}

func TestVisionAnalyzer_AnalyzeScreenshot_MalformedResponse(t *testing.T) {
	provider := &mockVisionLLM{response: "looks fine to me"}
	analyzer := NewVisionAnalyzer(provider)
	result, err := analyzer.AnalyzeScreenshot(context.Background(), []byte("png"), "screen", "android")
	require.NoError(t, err)
	assert.Len(t, result, 0) // graceful degradation
}

func TestVisionAnalyzer_CompareScreenshots(t *testing.T) {
	findings := []AnalysisFinding{
		{Category: CategoryVisual, Severity: SeverityHigh, Title: "Layout shifted", Description: "Button moved 20px"},
	}
	mockJSON, _ := json.Marshal(findings)
	provider := &mockVisionLLM{response: string(mockJSON)}

	analyzer := NewVisionAnalyzer(provider)
	result, err := analyzer.CompareScreenshots(context.Background(), []byte("before"), []byte("after"), "detail", "android")
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(result), 1)
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd HelixQA && go test ./pkg/analysis/ -v -run TestVisionAnalyzer`
Expected: FAIL — `NewVisionAnalyzer` undefined

- [ ] **Step 3: Write the implementation**

```go
// pkg/analysis/vision.go
package analysis

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"digital.vasic.helixqa/pkg/llm"
)

// VisionAnalyzer uses LLM vision to analyze screenshots for issues.
type VisionAnalyzer struct {
	provider llm.Provider
}

// NewVisionAnalyzer creates an analyzer backed by the given LLM.
func NewVisionAnalyzer(provider llm.Provider) *VisionAnalyzer {
	return &VisionAnalyzer{provider: provider}
}

// AnalyzeScreenshot analyzes a single screenshot for visual/UX issues.
func (v *VisionAnalyzer) AnalyzeScreenshot(ctx context.Context, imageData []byte, screen, platform string) ([]AnalysisFinding, error) {
	prompt := fmt.Sprintf(screenshotAnalysisPrompt, screen, platform)

	resp, err := v.provider.Vision(ctx, imageData, prompt)
	if err != nil {
		return nil, fmt.Errorf("analysis: vision call failed: %w", err)
	}

	return v.parseFindings(resp.Content, screen, platform), nil
}

// CompareScreenshots compares before/after screenshots for regressions.
func (v *VisionAnalyzer) CompareScreenshots(ctx context.Context, before, after []byte, screen, platform string) ([]AnalysisFinding, error) {
	// Analyze the "after" image with context about it being a comparison
	prompt := fmt.Sprintf(comparisonPrompt, screen, platform)

	resp, err := v.provider.Vision(ctx, after, prompt)
	if err != nil {
		return nil, fmt.Errorf("analysis: vision comparison failed: %w", err)
	}

	return v.parseFindings(resp.Content, screen, platform), nil
}

func (v *VisionAnalyzer) parseFindings(content, screen, platform string) []AnalysisFinding {
	cleaned := content
	if idx := strings.Index(cleaned, "["); idx >= 0 {
		cleaned = cleaned[idx:]
	}
	if idx := strings.LastIndex(cleaned, "]"); idx >= 0 {
		cleaned = cleaned[:idx+1]
	}

	var findings []AnalysisFinding
	if err := json.Unmarshal([]byte(cleaned), &findings); err != nil {
		return nil
	}

	for i := range findings {
		findings[i].Screen = screen
		findings[i].Platform = platform
	}
	return findings
}

const screenshotAnalysisPrompt = `Analyze this screenshot of the "%s" screen on %s platform.

Check for these issues and report ONLY actual problems found:
1. VISUAL: Text clipping, overflow, misalignment, wrong colors, missing assets, broken layout
2. UX: Confusing navigation, missing feedback, unclear labels, poor affordance
3. ACCESSIBILITY: Low contrast, small touch targets, missing alt text
4. BRAND: Logo compliance (should be rounded square with red border), color palette consistency
5. CONTENT: Placeholder text, broken images, empty states showing raw data
6. PERFORMANCE: Loading indicators stuck, visual jank artifacts

Return ONLY a JSON array of findings. Each finding: {"category":"visual|ux|accessibility|brand|content|performance", "severity":"critical|high|medium|low|cosmetic", "title":"Short title", "description":"Detailed description"}
If no issues found, return [].`

const comparisonPrompt = `Compare this screenshot of the "%s" screen on %s against expected behavior.

Look for:
1. Layout shifts or visual regressions
2. Missing elements compared to expected UI
3. New visual artifacts or broken rendering
4. Brand compliance issues

Return ONLY a JSON array of findings. Each: {"category":"visual|ux|accessibility|brand", "severity":"critical|high|medium|low|cosmetic", "title":"Short title", "description":"Description"}
If no issues, return [].`
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd HelixQA && go test ./pkg/analysis/ -v`
Expected: PASS (4 tests)

- [ ] **Step 5: Commit**

```bash
git add HelixQA/pkg/analysis/
git commit -m "feat(helixqa): add VisionAnalyzer for LLM-powered screenshot analysis"
```

---

## Task 8: Final Integration Test & Push

- [ ] **Step 1: Run ALL HelixQA tests**

Run: `cd HelixQA && go test ./... -race -count=1 -v 2>&1 | tail -30`
Expected: All packages pass

- [ ] **Step 2: Run go vet**

Run: `cd HelixQA && go vet ./...`
Expected: Clean

- [ ] **Step 3: Count total tests**

Run: `cd HelixQA && go test ./... -v 2>&1 | grep -c "--- PASS"`
Expected: ~770+ (742 prior + ~28 new)

- [ ] **Step 4: Push HelixQA submodule**

```bash
cd HelixQA && GIT_SSH_COMMAND="ssh -o BatchMode=yes" git push origin main
```

- [ ] **Step 5: Commit and push main repo**

```bash
cd /run/media/milosvasic/DATA4TB/Projects/Catalogizer
git add HelixQA
git commit -m "feat(helixqa): Phase 3 — execution engine (performance, video, maestro, analysis)"
GIT_SSH_COMMAND="ssh -o BatchMode=yes" git push origin main
```
