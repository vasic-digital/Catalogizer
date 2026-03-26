# HelixQA Autonomous Robot — Design Specification

**Date:** 2026-03-26
**Status:** Approved
**Approach:** Extend existing HelixQA module (Approach A)
**LLM Strategy:** Adaptive (cloud / hybrid / self-hosted, auto-selected by config)

## 1. System Overview

HelixQA Autonomous Robot extends the existing 19.3K-line HelixQA Go module into a fully autonomous, fire-and-forget QA system that operates like a living human QA engineer. It learns the project, plans comprehensive testing, executes across all platforms with video recording and performance monitoring, performs deep post-analysis using LLM vision, and creates detailed issue tickets — all with photographic memory across sessions.

### Architecture

```
+-----------------------------------------------------------+
|                    CLI / Container Entry                    |
|  helixqa autonomous --project /path --platforms all        |
+-----------------------------------------------------------+
|                  Session Coordinator                       |
|  pkg/autonomous/coordinator.go (EXISTING, wire up)        |
+-----------+-----------+-----------+-----------------------+
| Learning  | Planning  | Execution | Post-Analysis          |
| Phase     | Phase     | Phase     | Phase                  |
+-----------+-----------+-----------+-----------------------+
|                  Platform Workers                          |
|  Web (Playwright) | Android (Maestro+ADB) | Desktop        |
+-------------------+-----------------------+----------------+
|                  Shared Services                           |
| LLM Provider | Vision  | Evidence | Memory  | Performance |
| (adaptive)   | Engine  | Collect  | Store   | Monitor     |
+-----------------------------------------------------------+
```

### External Tool Integrations

| Tool | Purpose | Integration Method |
|------|---------|-------------------|
| Playwright | Web automation + video + screenshots | Go subprocess (npx playwright) |
| Maestro | Mobile YAML flow execution | Go subprocess (maestro test) |
| scrcpy | Android video recording (all SDK versions) | Go subprocess (scrcpy --record) |
| ADB | Android device control, screenshots, logcat | Go subprocess (existing ADBExecutor) |
| ffmpeg | Video frame extraction, assembly, analysis | Go subprocess (existing VideoManager) |
| Perfetto | Deep Android system profiling | Go subprocess (perfetto trace config push) |
| LeakCanary | Android memory leak detection | Logcat parser (LeakCanary outputs to logcat) |
| pixelmatch | Screenshot visual comparison | Node subprocess or Go port |
| Allure | Report generation (HTML) | JSON output adapter |
| docker-android | Containerized Android emulators | Podman compose service |
| UI-TARS | Self-hosted vision model for UI analysis | HTTP API client |
| Ollama | Self-hosted LLM for reasoning | HTTP API client |

## 2. Phase 1 — Learning Engine

When the robot starts, it ingests everything about the project before testing.

### Sources (in order)

1. `CLAUDE.md`, `AGENTS.md`, all submodule `CLAUDE.md` files
2. `docs/` directory — all markdown documentation
3. Git history — recent commits, changelog patterns, change hotspots
4. Codebase structure — file tree, routes, screens, components, API endpoints
5. Existing test banks — `challenges/helixqa-banks/*.json`, `HelixQA/banks/*.yaml`
6. Previous QA sessions — memory DB + `qa-results/` reports
7. `docs/issues/` — open/in-progress tickets from prior sessions
8. UI assets — screenshots from prior sessions, design mockups

### New Package: `pkg/learning/`

| Component | Responsibility |
|-----------|---------------|
| `ProjectReader` | Walk project tree, read docs, parse CLAUDE.md files |
| `GitAnalyzer` | Extract recent changes, identify frequently-changed files |
| `CodebaseMapper` | Build map of routes, screens, components, API endpoints |
| `SessionHistorian` | Load prior QA session data from memory DB |
| `KnowledgeBase` | Unified struct holding everything learned, fed to LLM |

### Output

`KnowledgeBase` struct containing: project map, screen inventory, API endpoint list, known issues, prior coverage data, component relationships.

## 3. Phase 2 — Planning Engine

Builds a comprehensive test plan before touching any app.

### Planning Steps

1. **Screen inventory** — LLM enumerates every screen/route across all platforms from codebase map
2. **Flow enumeration** — Identifies all user flows: auth, CRUD, media browsing, search, favorites, collections, playback, admin, settings, sync, conversion
3. **Edge case generation** — LLM generates edge cases per screen: empty states, error states, network failure, invalid input, rotation, background/restore, rapid navigation
4. **Test bank reconciliation** — Cross-references plan against existing banks; new tests appended, existing tests marked for execution
5. **Priority ordering** — Critical paths first, then secondary flows, then edge cases, then curiosity exploration
6. **Coverage mapping** — Tracks coverage percentage vs prior sessions

### New Package: `pkg/planning/`

| Component | Responsibility |
|-----------|---------------|
| `TestPlanGenerator` | LLM + KnowledgeBase produces structured `TestPlan` |
| `BankReconciler` | Diffs plan against existing YAML/JSON banks, appends new cases |
| `PriorityRanker` | Orders by criticality, prior failure history, change recency |
| `CoverageEstimator` | Estimates coverage vs total known surface area |

### Output

`TestPlan` struct with ordered test cases (existing + newly generated), each with: platform targets, expected screenshots, estimated duration, priority, category.

## 4. Phase 3 — Execution Engine

Runs the test plan across all platforms in parallel with continuous recording and monitoring.

### Per-Platform Stack

| Capability | Web | Android Phone | Android TV | Desktop |
|------------|-----|---------------|------------|---------|
| UI Automation | Playwright | Maestro + ADB | Maestro + ADB | Playwright (Tauri WebView) |
| Video Recording | Playwright `--video on` | scrcpy (H.265) | `adb screenrecord` (SDK<=33) / scrcpy | ffmpeg x11grab |
| Screenshots | Playwright | `adb screencap` | `adb screencap` | Playwright |
| Crash Detection | Console error listener | `adb logcat` FATAL/ANR | `adb logcat` FATAL/ANR | stderr + process monitor |
| Performance | Web Vitals | `adb dumpsys meminfo/cpuinfo` + Perfetto | `adb dumpsys` | process stats |
| Memory Leaks | JS heap snapshots | LeakCanary + dumpsys deltas | dumpsys deltas | RSS tracking |
| Network | Request interception | `adb dumpsys netstats` | `adb dumpsys netstats` | network stats |
| Login | Playwright fill + click | Maestro `inputText` / ADB | ADB `input text` + keyevents | Playwright fill |

### Execution Flow Per Test Case

1. Pre-step screenshot (baseline state)
2. Start step timer
3. Execute action (navigate, click, type, scroll)
4. Wait for stability (no pending requests, no animations)
5. Post-step screenshot
6. Crash/ANR check via logcat parser
7. Performance metrics snapshot (memory, CPU, network)
8. LLM vision comparison: before/after screenshots for visual issues
9. Validate expected outcome
10. Record result + evidence to timeline

### Curiosity Phase

After planned tests complete, the robot enters curiosity-driven exploration. The LLM sees current screen state, decides what to try next (existing `NavigationEngine.ExploreUnknown()`), executes, analyzes. Continues until coverage target met or curiosity timeout expires.

### New/Extended Packages

| Package | Status | Changes |
|---------|--------|---------|
| `pkg/autonomous/coordinator.go` | Existing | Wire CLI entry (the TODO at line 411) |
| `pkg/performance/` | **New** | `MetricsCollector` with ADB dumpsys, Perfetto, Web Vitals collectors |
| `pkg/video/` | **New** | `ScrcpyRecorder` subprocess bridge, extends existing `VideoManager` |
| `pkg/maestro/` | **New** | `FlowRunner` subprocess bridge to Maestro CLI |
| `pkg/navigator/executor.go` | Existing | Fully implement `PlaywrightExecutor`, wire `ADBExecutor` to real devices |
| `pkg/detector/` | Existing | Add memory leak delta tracking, network error detection |

## 5. Phase 4 — Post-Analysis Engine

Deep analysis on all captured evidence after execution completes.

### Analysis Pipeline

1. **Video frame extraction** — ffmpeg extracts key frames (scene change detection, 1fps minimum)
2. **LLM Vision analysis** — Every screenshot and key frame analyzed for:
   - Visual defects (misalignment, clipping, overflow, wrong colors, missing assets)
   - Brand compliance (logo styling, colors, fonts, spacing)
   - UX issues (confusing navigation, missing feedback, dead-end screens)
   - Accessibility (low contrast, small touch targets, missing labels)
   - Responsiveness (layout adaptation across screen sizes)
   - Content quality (placeholder text, broken images, raw data in empty states)
3. **Performance analysis** — Time-series anomaly detection:
   - Memory growth trends (leak indicators)
   - CPU spikes correlated with actions
   - Network failures and slow responses
   - Frame drops during transitions
4. **Log analysis** — Pattern matching + LLM summarization:
   - Warnings indicating real problems
   - Repeated error patterns
   - Framework performance warnings
5. **Cross-reference with prior sessions** — Regression detection:
   - Regressions (passed before, fails now)
   - Fixed issues (filed tickets that now pass)
   - New surfaces not seen before
   - Performance trends across sessions

### New Package: `pkg/analysis/`

| Component | Responsibility |
|-----------|---------------|
| `VideoAnalyzer` | ffmpeg frame extraction + LLM vision per frame |
| `PerformanceAnalyzer` | Time-series anomaly detection on metrics |
| `LogAnalyzer` | Pattern matching + LLM summarization of logs |
| `RegressionDetector` | Diffs current results against memory DB |
| `AnalysisReport` | Aggregated findings fed to ticket generator |

### Output

List of `Finding` structs: severity, category, description, reproduction steps, evidence paths, affected platform, recommended fix. Each becomes an issue ticket.

## 6. Photographic Memory & Issue Lifecycle

### Memory Store — SQLite (`HelixQA/data/memory.db`)

| Table | Purpose |
|-------|---------|
| `sessions` | Every QA session: timestamp, duration, platforms, coverage, pass/fail, findings |
| `test_results` | Every test execution: session, test_case, platform, status, duration, evidence |
| `findings` | Every issue discovered: severity, category, title, repro steps, evidence, status |
| `screenshots` | Every screenshot: screen name, platform, path, dimensions, hash for dedup |
| `metrics` | Performance measurements: type, value, timestamp per session/platform |
| `knowledge` | Learned project facts: key/value with source and last_verified timestamp |
| `coverage` | Screen/flow coverage: screen name, platform, last tested, times tested, status |

### Issue Lifecycle in `docs/issues/`

Each finding produces: `docs/issues/HELIX-{NNN}-{slug}.md` with YAML frontmatter:

```yaml
id: HELIX-042
status: open          # open | in_progress | fixed | verified | reopened | wontfix
severity: high        # critical | high | medium | low | cosmetic
category: visual      # visual | ux | functional | performance | crash | accessibility
platform: android
screen: media-detail
found_session: session-20260326-204518
found_date: 2026-03-26
```

Body contains: description, reproduction steps, evidence references (screenshot paths, video timestamps), expected behavior, environment details.

### Session-to-Session Awareness

- Prior findings loaded before each session; fixed issues re-verified; regressions reopen tickets
- Coverage gaps from prior sessions prioritized in current plan
- Performance baselines enable trend detection
- LLM receives summary of prior sessions as context

### New Package: `pkg/memory/`

| Component | Responsibility |
|-----------|---------------|
| `Store` | SQLite wrapper with migrations (using `digital.vasic.database` patterns) |
| `SessionRecorder` | Extends existing `pkg/session/` to persist to DB |
| `IssueManager` | CRUD for `docs/issues/` markdown + DB sync |
| `CoverageTracker` | Extends existing tracking with persistence |
| `HistorySummarizer` | Generates LLM-friendly summaries of prior sessions |

## 7. Adaptive LLM Provider

Auto-selects based on available configuration:

1. Check env for API keys -> cloud provider (fastest to start)
2. Check for local Ollama/vLLM endpoint -> self-hosted
3. Check for UI-TARS model path -> dedicated vision model
4. Fallback: refuse to start (LLM required)

| Mode | Reasoning LLM | Vision LLM | Config |
|------|---------------|------------|--------|
| Cloud | Claude/GPT-4o/Gemini | Same (multimodal) | `HELIX_LLM_PROVIDER=anthropic` |
| Hybrid | Cloud API | Self-hosted UI-TARS | `HELIX_VISION_PROVIDER=ui-tars` |
| Self-hosted | Ollama (Qwen 2.5) | Ollama (Qwen 2.5 VL) | `HELIX_LLM_PROVIDER=ollama` |

### New Package: `pkg/llm/`

| Component | Responsibility |
|-----------|---------------|
| `Provider` interface | `Chat(ctx, messages)`, `Vision(ctx, image, prompt)` |
| `AnthropicProvider` | Claude API implementation |
| `OpenAIProvider` | GPT-4o API implementation |
| `GoogleProvider` | Gemini API implementation |
| `OllamaProvider` | Local Ollama HTTP API |
| `UITarsProvider` | UI-TARS HTTP API for vision |
| `AdaptiveProvider` | Wraps multiple, selects by config, falls back on failure |
| `PromptBuilder` | Versioned prompts for each analysis type |

## 8. Container Runtime

### Compose File: `docker-compose.qa-robot.yml`

```yaml
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
      - HELIX_LLM_PROVIDER=${HELIX_LLM_PROVIDER}
      - ANTHROPIC_API_KEY=${ANTHROPIC_API_KEY}
    devices:
      - /dev/bus/usb

  android-emulator:
    image: budtmo/docker-android:emulator_14.0
    privileged: true
    ports: ["6080:6080", "5555:5555"]
    environment:
      EMULATOR_DEVICE: "pixel_6"
    devices: ["/dev/kvm"]

  ui-tars:
    image: ghcr.io/bytedance/ui-tars:7b
    deploy:
      resources:
        reservations:
          devices: [{driver: nvidia, count: 1, capabilities: [gpu]}]
    ports: ["8000:8000"]
```

### Dockerfile (`HelixQA/Dockerfile`)

Base: Ubuntu 22.04 with Go 1.24, Node 18 (Playwright), Maestro, scrcpy, ffmpeg, ADB platform-tools.

### CLI

```bash
# Direct
helixqa autonomous --project . --platforms all --timeout 4h

# Containerized
podman-compose -f docker-compose.qa-robot.yml up helixqa-robot

# Multi-pass
helixqa autonomous --project . --platforms all --pass 2
```

## 9. Testing Strategy

| Layer | What | How |
|-------|------|-----|
| Unit | Each new package | Table-driven Go tests with CommandRunner mocks |
| Integration | LLM round-trips, ADB commands, Maestro parsing | Real subprocess behind `//go:build integration` |
| Contract | LLM prompt/response contracts | Recorded fixtures |
| Smoke | Full autonomous run | 5-minute containerized run, all 4 phases |

## 10. Quality Standards (Enforced by Robot)

- Zero crashes, zero ANRs (critical ticket)
- Zero console errors (high ticket)
- API response <500ms p95 (performance ticket)
- Memory growth <10% over 5-min session (leak ticket)
- All screens render within 3s (performance ticket)
- No visual clipping, overflow, misalignment (visual ticket)
- Brand compliance: VD logo in rounded square with red border
- Responsive layouts across all form factors
- All API endpoints return 2xx on valid requests
- Widget reusability: flag duplicated UI patterns

## 11. Report Output Structure

```
qa-results/session-{timestamp}/
+-- report.md                     # Executive summary
+-- report.html                   # Interactive HTML
+-- report.json                   # Machine-readable
+-- videos/
|   +-- web-session.mp4
|   +-- phone1-flow.mp4
|   +-- mibox1-flow.mp4
|   +-- desktop-flow.mp4
+-- screenshots/
|   +-- web/{screen}-{state}.png
|   +-- android/{screen}-{state}.png
|   +-- androidtv/{screen}-{state}.png
+-- evidence/
|   +-- logcat-phone1.txt
|   +-- console-web.json
|   +-- perfetto-traces/
+-- metrics/
|   +-- memory-timeline.json
|   +-- cpu-timeline.json
|   +-- network-requests.json
+-- tickets/
|   +-- HELIX-043.md
|   +-- HELIX-044.md
+-- test-bank-additions.yaml      # New test cases discovered
```

## 12. New Code Estimate

| Package | Lines (est.) | Status |
|---------|-------------|--------|
| `pkg/learning/` | ~1,200 | New |
| `pkg/planning/` | ~800 | New |
| `pkg/llm/` | ~1,500 | New |
| `pkg/performance/` | ~600 | New |
| `pkg/video/` (scrcpy) | ~400 | New |
| `pkg/maestro/` | ~500 | New |
| `pkg/analysis/` | ~1,000 | New |
| `pkg/memory/` | ~800 | New |
| CLI wiring | ~300 | Extend existing |
| Container files | ~200 | New |
| Tests | ~1,500 | New |
| **Total new** | **~8,800** | |
| **Existing HelixQA** | **19,300** | Preserved |
| **Final total** | **~28,100** | |

## 13. Open-Source Tools Reference

### Tier 1 — Direct Integration

| Tool | Stars | Purpose |
|------|-------|---------|
| Playwright | 84K | Web automation + video + screenshots |
| Maestro | 7K+ | YAML mobile flow execution |
| scrcpy | 115K+ | Android video recording |
| LeakCanary | 29K+ | Android memory leak detection |
| docker-android | 14K+ | Containerized Android emulators |
| Allure Report | 4K+ | Unified test reporting |
| k6 | 29K | Load testing (already integrated) |
| ffmpeg | 50K+ | Video processing + frame extraction |
| pixelmatch | 6.6K | Pixel-level screenshot comparison |

### Tier 2 — LLM-Driven Autonomy

| Tool | Stars | Purpose |
|------|-------|---------|
| UI-TARS | 10K+ | Self-hosted vision model for UI automation |
| browser-use | 84K | LLM-driven web exploration |
| Midscene.js | 5K+ | Vision-driven multi-platform automation |
| Stagehand | 12K+ | Self-healing browser automation |
| EvoMaster | 657 | Evolutionary API test generation |
| ReportPortal | 2K+ | ML-powered failure categorization |
| ACRA | 6.2K | Android crash reporting |
| Perfetto | 3K+ | System-level Android profiling |

### Tier 3 — Evaluate Later

| Tool | Purpose |
|------|---------|
| Skyvern | Planner-actor-validator browser agent |
| Magnitude | Dual-agent test planner/executor |
| Agent-S / Open Computer Use | Computer-use agents for desktop |
| BackstopJS | Visual regression scenarios |
| Kiwi TCMS | Test case management with API |
| Qwen 2.5 VL | Self-hosted multimodal LLM |
