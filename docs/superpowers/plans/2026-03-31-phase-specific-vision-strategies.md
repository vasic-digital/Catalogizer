# Phase-Specific Vision Strategies Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace single-strategy vision model selection with per-phase optimized strategies, add bridged CLI model support, and run LLMsVerifier scoring rounds per phase.

**Architecture:** Each HelixQA pipeline phase gets its own strategy that selects the best model for that phase's specific needs. LLMsVerifier runs scoring rounds before each phase. Bridged CLI providers (Claude Code, Qwen Code, OpenCode/Zen) are discovered and scored alongside API and local providers.

**Tech Stack:** Go 1.24+, LLMsVerifier Strategy pattern, HelixQA pipeline phases, CLI process IPC for bridged models.

---

## Phase Architecture

```
HelixQA Pipeline Phase → LLMsVerifier Strategy → Best Model Selected
─────────────────────────────────────────────────────────────────────
Phase 1 (Learn)     → ChatStrategy      → Best reasoning model (Claude/GPT-4)
Phase 2 (Plan)      → PlanningStrategy  → Best structured output model
Phase 3 (Execute)   → NavigationStrategy → Best JSON-action vision model
Phase 3.5 (Curiosity) → NavigationStrategy → Same as Execute
Phase 4 (Analyze)   → AnalysisStrategy  → Best image description model (Astica)
```

## File Structure

### LLMsVerifier (generic strategies — reusable by any project)

| File | Responsibility |
|------|---------------|
| `pkg/strategy/interface.go` | Existing — VerificationStrategy interface |
| `pkg/vision/strategy.go` | Existing — VisionStrategy (generic vision scoring) |
| `pkg/vision/navigation.go` | NEW — NavigationStrategy (JSON-action producing vision models) |
| `pkg/vision/analysis.go` | NEW — AnalysisStrategy (image description/captioning models) |
| `pkg/chat/planning.go` | NEW — PlanningStrategy (structured output for test plans) |
| `pkg/bridge/discovery.go` | NEW — Discover bridged CLI models (claude, qwen-coder, opencode) |
| `pkg/bridge/provider.go` | NEW — BridgedModelInfo for scoring CLI models |
| `pkg/bridge/discovery_test.go` | NEW — Tests for CLI discovery |
| `pkg/vision/navigation_test.go` | NEW — Tests for NavigationStrategy |
| `pkg/vision/analysis_test.go` | NEW — Tests for AnalysisStrategy |
| `pkg/chat/planning_test.go` | NEW — Tests for PlanningStrategy |

### HelixQA (phase-specific wiring — HelixQA-specific)

| File | Responsibility |
|------|---------------|
| `pkg/llm/phase_selector.go` | NEW — PhaseModelSelector: runs LLMsVerifier per phase |
| `pkg/llm/phase_selector_test.go` | NEW — Tests |
| `pkg/llm/bridge_provider.go` | NEW — BridgedCLIProvider adapter for HelixQA |
| `pkg/llm/bridge_provider_test.go` | NEW — Tests |
| `pkg/autonomous/pipeline.go` | MODIFY — Use PhaseModelSelector instead of single provider |
| `cmd/helixqa/main.go` | MODIFY — Discover bridged models, wire phase selector |

---

## Task 1: NavigationStrategy in LLMsVerifier

Scores models specifically for UI navigation (producing JSON action arrays from screenshots).

**Files:**
- Create: `LLMsVerifier/pkg/vision/navigation.go`
- Create: `LLMsVerifier/pkg/vision/navigation_test.go`

**Scoring dimensions (NavigationStrategy):**
- JSON compliance (40%) — ability to produce valid JSON arrays, not descriptions
- GUI understanding (25%) — UI element detection, focus tracking
- Speed (20%) — fast response for interactive navigation (< 5s ideal)
- Cost (15%) — prefer free/cheap for high-volume navigation calls

**Key distinction from VisionStrategy:** NavigationStrategy penalizes models that return descriptions instead of JSON (like Astica). Astica gets LOW navigation score because it returns `caption_GPTS` text, not JSON arrays.

- [ ] **Step 1: Write failing test**
```go
// navigation_test.go
func TestNavigationStrategy_AsticaScoresLow(t *testing.T) {
    s := NewNavigationStrategy()
    astica := strategy.ModelInfo{
        ID: "astica-vision", Provider: "astica",
        SupportsVision: true,
        Capabilities: []string{"vision", "ocr", "object_detection"},
        // No "json_output" or "structured_output" capability
    }
    score, err := s.Score(context.Background(), astica)
    require.NoError(t, err)
    assert.Less(t, score.Overall, 0.5, "Astica should score low for navigation")
}

func TestNavigationStrategy_GeminiScoresHigh(t *testing.T) {
    s := NewNavigationStrategy()
    gemini := strategy.ModelInfo{
        ID: "gemini-flash", Provider: "google",
        SupportsVision: true,
        Capabilities: []string{"vision", "json_output", "structured_output", "gui_analysis"},
        AvgLatencyMs: 800,
    }
    score, err := s.Score(context.Background(), gemini)
    require.NoError(t, err)
    assert.Greater(t, score.Overall, 0.7, "Gemini should score high for navigation")
}
```

- [ ] **Step 2: Run test — verify FAIL**
```bash
cd LLMsVerifier && go test -run TestNavigationStrategy ./pkg/vision/ -v
```

- [ ] **Step 3: Implement NavigationStrategy**
```go
// navigation.go
type NavigationStrategy struct {
    jsonComplianceWeight float64  // 0.40
    guiUnderstandWeight  float64  // 0.25
    speedWeight          float64  // 0.20
    costWeight           float64  // 0.15
}

func NewNavigationStrategy() *NavigationStrategy
func (s *NavigationStrategy) Name() string { return "navigation" }
func (s *NavigationStrategy) Description() string
func (s *NavigationStrategy) Score(ctx, model) (StrategyScore, error)
func (s *NavigationStrategy) Validate(ctx, model) ValidationResult
func (s *NavigationStrategy) Rank(ctx, models) ([]RankedModel, error)
func (s *NavigationStrategy) Select(ctx, ranked, req) (ModelInfo, error)
```

JSON compliance scoring: models with `json_output` or `structured_output` capability get 1.0. Models without get 0.2 (Astica, descriptions-only models).

- [ ] **Step 4: Run tests — verify PASS**
- [ ] **Step 5: Commit**

---

## Task 2: AnalysisStrategy in LLMsVerifier

Scores models for screenshot analysis (descriptive captions, issue detection).

**Files:**
- Create: `LLMsVerifier/pkg/vision/analysis.go`
- Create: `LLMsVerifier/pkg/vision/analysis_test.go`

**Scoring dimensions (AnalysisStrategy):**
- Description quality (35%) — rich, detailed image descriptions
- OCR accuracy (20%) — text recognition in screenshots
- Object detection (20%) — UI element identification
- Comprehensiveness (15%) — covers all visual aspects
- Cost (10%)

**Key distinction:** Astica scores HIGHEST here because it's specialized for rich image descriptions, OCR, object detection. Models that produce JSON (Gemini, Kimi) score lower because their descriptions are less detailed.

- [ ] Steps similar to Task 1 — TDD cycle
- [ ] Astica scores > 0.9 for analysis
- [ ] Gemini scores ~0.7 for analysis
- [ ] Commit

---

## Task 3: PlanningStrategy in LLMsVerifier

Scores models for test plan generation (structured reasoning, long context).

**Files:**
- Create: `LLMsVerifier/pkg/chat/planning.go`
- Create: `LLMsVerifier/pkg/chat/planning_test.go`

**Scoring dimensions (PlanningStrategy):**
- Reasoning quality (35%) — complex multi-step planning
- Context window (25%) — needs large context for KB + test plans
- Structured output (20%) — produces organized test cases
- Speed (10%) — acceptable latency for planning (one-time per session)
- Cost (10%)

- [ ] TDD cycle
- [ ] Claude/GPT-4 score highest (strong reasoning)
- [ ] Commit

---

## Task 4: Bridged CLI Model Discovery

Discover models available via CLI tools (Claude Code, Qwen Code, OpenCode/Zen).

**Files:**
- Create: `LLMsVerifier/pkg/bridge/discovery.go`
- Create: `LLMsVerifier/pkg/bridge/provider.go`
- Create: `LLMsVerifier/pkg/bridge/discovery_test.go`

**Discovery mechanism:**
```go
type BridgedModel struct {
    CLIName   string // "claude", "qwen-coder", "opencode"
    CLIPath   string // resolved binary path
    Model     string // discovered model name
    Available bool
    SupportsVision bool
}

func DiscoverBridgedModels() []BridgedModel {
    // Check: which claude | claude --version
    // Check: which qwen-coder | qwen-coder --version
    // Check: which opencode | opencode --version
    // For each available: query model info
}
```

- [ ] TDD cycle
- [ ] Test with mock binaries
- [ ] Commit

---

## Task 5: BridgedCLIProvider in HelixQA

Adapter that wraps CLI tools as HelixQA LLM providers.

**Files:**
- Create: `HelixQA/pkg/llm/bridge_provider.go`
- Create: `HelixQA/pkg/llm/bridge_provider_test.go`

Based on HelixAgent's `ClaudeCLIProvider` pattern:
- Spawn CLI process with `--json` output
- Parse JSON response
- Support vision via file paths (save screenshot to temp, pass path)
- Support streaming via line-by-line reading

- [ ] TDD cycle
- [ ] Commit

---

## Task 6: PhaseModelSelector in HelixQA

Runs LLMsVerifier scoring round per pipeline phase and selects the best model.

**Files:**
- Create: `HelixQA/pkg/llm/phase_selector.go`
- Create: `HelixQA/pkg/llm/phase_selector_test.go`

```go
type PhaseModelSelector struct {
    providers       []Provider
    bridgedModels   []BridgedCLIProvider
    strategies      map[string]strategy.VerificationStrategy
}

func (s *PhaseModelSelector) SelectForPhase(phase string) Provider {
    // Get the strategy for this phase
    strat := s.strategies[phase]
    // Score all available providers
    // Return the highest-scoring one
}
```

Phase → Strategy mapping:
```go
"learn"     → PlanningStrategy (reasoning)
"plan"      → PlanningStrategy (structured output)
"execute"   → NavigationStrategy (JSON vision)
"curiosity" → NavigationStrategy (JSON vision)
"analyze"   → AnalysisStrategy (description vision)
```

- [ ] TDD cycle
- [ ] Commit

---

## Task 7: Wire PhaseModelSelector into Pipeline

**Files:**
- Modify: `HelixQA/pkg/autonomous/pipeline.go`
- Modify: `HelixQA/cmd/helixqa/main.go`

Before each phase, run the selector:
```go
// Phase 2: Plan
sp.setPhase("plan")
planProvider := sp.phaseSelector.SelectForPhase("plan")
gen := planning.NewTestPlanGenerator(planProvider)

// Phase 3.5: Curiosity
sp.setPhase("curiosity")
navProvider := sp.phaseSelector.SelectForPhase("curiosity")
// Use navProvider for vision calls (NOT Astica for navigation)
```

- [ ] Wire selector
- [ ] Test end-to-end
- [ ] Commit

---

## Task 8: Register Strategies as Presets

**Files:**
- Modify: `LLMsVerifier/pkg/recipe/presets.go`

```go
func NavigationPreset() *vision.NavigationStrategy
func AnalysisPreset() *vision.AnalysisStrategy
func PlanningPreset() *chat.PlanningStrategy
```

- [ ] Add presets
- [ ] Commit

---

## Task 9: Update Documentation

**Files:**
- Modify: `LLMsVerifier/CLAUDE.md`
- Modify: `HelixQA/CLAUDE.md`
- Modify: `LLMsVerifier/docs/ARCHITECTURE.md`
- Modify: `HelixQA/docs/architecture.md`
- Modify: Root `CLAUDE.md`

Document:
- 6 strategies total (Default, Vision, Navigation, Analysis, Planning, Catalogizer)
- Phase → Strategy mapping
- Bridged CLI model support
- How scoring rounds work per phase

- [ ] Update all docs
- [ ] Commit

---

## Task 10: Rebuild, Test, Deploy, HelixQA Session

- [ ] Rebuild all binaries
- [ ] Run all Go tests (catalog-api, HelixQA, LLMsVerifier, VisionEngine)
- [ ] Run all frontend tests
- [ ] Deploy APK to devices
- [ ] Launch HelixQA with phase-specific strategies
- [ ] Verify: Execute phase uses NavigationStrategy (NOT Astica)
- [ ] Verify: Analyze phase uses AnalysisStrategy (Astica first)
- [ ] Verify: Plan phase uses PlanningStrategy (Claude/Gemini)
- [ ] Commit and push all submodules + main repo

---

## Expected Results After Implementation

| Phase | Strategy | Expected Top Model | Why |
|-------|----------|-------------------|-----|
| Learn | PlanningStrategy | Gemini Flash | Large context, fast, free |
| Plan | PlanningStrategy | Claude/Gemini | Strong reasoning |
| Execute | NavigationStrategy | Gemini/Kimi | JSON output, fast |
| Curiosity | NavigationStrategy | Gemini/Kimi | JSON output, fast |
| Analyze | AnalysisStrategy | Astica | Best image descriptions |

Astica will score #1 for Analyze but LOW for Navigate — exactly what we need.
