# HelixQA Autonomous QA Session - Comprehensive Implementation Plan

## Executive Summary

This document outlines the complete implementation plan for extending HelixQA with an advanced "Autonomous QA Session" test type. This extension will enable LLM-powered autonomous testing across all platforms with documentation-driven verification, curiosity-driven exploration, and comprehensive evidence collection.

## Project Overview

### Goals

1. **Extend HelixQA** with a new test type: "Autonomous QA Session"
2. **Extend LLMsVerifier** with Strategy pattern and Recipe builder for HelixQA-specific LLM selection
3. **Integrate OpenCode** (and other CLI agents) in headless mode for LLM orchestration
4. **Implement comprehensive documentation processing** for feature extraction and coverage tracking
5. **Create full evidence collection** (screenshots, video, logs) with timeline and ticketing
6. **Achieve 100% test coverage** across all test types
7. **Provide complete documentation** (user guides, API docs, video courses, diagrams)

### Scope

| Component | Changes | New Modules |
|-----------|---------|-------------|
| HelixQA | Enhanced autonomous session, new config options | `pkg/llmstrategy/`, `pkg/sessionstore/` |
| LLMsVerifier | Strategy pattern, Recipe builder | `pkg/strategy/`, `pkg/recipe/`, `pkg/helixqa/` |
| LLMOrchestrator | OpenCode adapter enhancements | OpenCode improvements |
| VisionEngine | Enhanced screen analysis | Navigation graph improvements |
| DocProcessor | Feature extraction, coverage | Already complete |
| New Module | QATicketing | Standalone ticketing system |

---

## Phase 1: Foundation & Infrastructure (Week 1-2)

### 1.1 LLMsVerifier Strategy Pattern Extension

#### 1.1.1 Core Strategy Interfaces

**File: `pkg/strategy/interface.go`**

```go
// VerificationStrategy defines how LLMs are scored and selected
type VerificationStrategy interface {
    // Name returns the strategy identifier
    Name() string
    
    // Score evaluates a model and returns a score (0-1)
    Score(ctx context.Context, model ModelInfo) (StrategyScore, error)
    
    // Validate checks if a model meets strategy requirements
    Validate(ctx context.Context, model ModelInfo) ValidationResult
    
    // Rank sorts models by strategy-specific criteria
    Rank(ctx context.Context, models []ModelInfo) ([]RankedModel, error)
    
    // Select chooses the best model from ranked list
    Select(ctx context.Context, ranked []RankedModel, req Requirements) (ModelInfo, error)
}

// StrategyScore contains detailed scoring breakdown
type StrategyScore struct {
    Overall      float64            `json:"overall"`
    Dimensions   map[string]float64 `json:"dimensions"`
    Confidence   float64            `json:"confidence"`
    Reasoning    string             `json:"reasoning"`
    Timestamp    time.Time          `json:"timestamp"`
}

// Requirements specifies what capabilities are needed
type Requirements struct {
    NeedsVision       bool              `json:"needs_vision"`
    NeedsStreaming    bool              `json:"needs_streaming"`
    MinContextWindow  int               `json:"min_context_window"`
    MaxLatencyMs      int               `json:"max_latency_ms"`
    MinQualityScore   float64           `json:"min_quality_score"`
    PreferredProvider string            `json:"preferred_provider"`
    CustomConstraints map[string]any    `json:"custom_constraints"`
}
```

#### 1.1.2 Recipe Builder Pattern

**File: `pkg/recipe/builder.go`**

```go
// RecipeBuilder constructs verification recipes
type RecipeBuilder struct {
    strategy    VerificationStrategy
    constraints []Constraint
    weights     map[string]float64
    fallbacks   []FallbackRule
}

// Recipe is a complete verification configuration
type Recipe struct {
    ID           string              `json:"id"`
    Name         string              `json:"name"`
    Strategy     VerificationStrategy `json:"-"`
    Constraints  []Constraint        `json:"constraints"`
    Weights      map[string]float64  `json:"weights"`
    Fallbacks    []FallbackRule      `json:"fallbacks"`
    Timeout      time.Duration       `json:"timeout"`
    MaxRetries   int                 `json:"max_retries"`
}

func NewRecipeBuilder() *RecipeBuilder {
    return &RecipeBuilder{
        weights:   make(map[string]float64),
        constraints: make([]Constraint, 0),
        fallbacks: make([]FallbackRule, 0),
    }
}

func (b *RecipeBuilder) WithStrategy(s VerificationStrategy) *RecipeBuilder {
    b.strategy = s
    return b
}

func (b *RecipeBuilder) WithWeight(dimension string, weight float64) *RecipeBuilder {
    b.weights[dimension] = weight
    return b
}

func (b *RecipeBuilder) WithConstraint(c Constraint) *RecipeBuilder {
    b.constraints = append(b.constraints, c)
    return b
}

func (b *RecipeBuilder) WithFallback(rule FallbackRule) *RecipeBuilder {
    b.fallbacks = append(b.fallbacks, rule)
    return b
}

func (b *RecipeBuilder) Build() (*Recipe, error) {
    if b.strategy == nil {
        return nil, ErrStrategyRequired
    }
    return &Recipe{
        ID:          uuid.New().String(),
        Strategy:    b.strategy,
        Constraints: b.constraints,
        Weights:     b.weights,
        Fallbacks:   b.fallbacks,
    }, nil
}
```

#### 1.1.3 HelixQA-Specific Strategy

**File: `pkg/helixqa/strategy.go`**

```go
// QAStrategy is optimized for autonomous QA testing
type QAStrategy struct {
    baseStrategy VerificationStrategy
    visionWeight float64
    speedWeight  float64
    costWeight   float64
}

func NewQAStrategy() *QAStrategy {
    return &QAStrategy{
        baseStrategy: NewDefaultStrategy(),
        visionWeight: 0.4,  // Vision is critical for QA
        speedWeight:  0.3,  // Need fast responses
        costWeight:   0.3,  // Cost efficiency matters
    }
}

func (s *QAStrategy) Score(ctx context.Context, model ModelInfo) (StrategyScore, error) {
    base, err := s.baseStrategy.Score(ctx, model)
    if err != nil {
        return StrategyScore{}, err
    }
    
    // QA-specific scoring adjustments
    qaScore := base.Overall
    
    // Bonus for vision capabilities
    if model.SupportsVision {
        qaScore += s.visionWeight * 0.2
    }
    
    // Penalty for high latency
    if model.AvgLatencyMs > 3000 {
        qaScore -= s.speedWeight * 0.1
    }
    
    return StrategyScore{
        Overall:    math.Min(1.0, qaScore),
        Dimensions: base.Dimensions,
        Confidence: base.Confidence,
        Reasoning:  fmt.Sprintf("QA-optimized score for %s", model.Name),
    }, nil
}
```

### 1.2 Tasks for Phase 1

| ID | Task | Priority | Est. Time | Dependencies |
|----|------|----------|-----------|--------------|
| P1-001 | Create `pkg/strategy/interface.go` with core interfaces | Critical | 4h | None |
| P1-002 | Create `pkg/strategy/default.go` with default strategy | Critical | 6h | P1-001 |
| P1-003 | Create `pkg/recipe/builder.go` with builder pattern | Critical | 6h | P1-001 |
| P1-004 | Create `pkg/recipe/validator.go` for recipe validation | High | 4h | P1-003 |
| P1-005 | Create `pkg/helixqa/strategy.go` with QA strategy | Critical | 8h | P1-001, P1-002 |
| P1-006 | Create `pkg/helixqa/recipe.go` with predefined QA recipes | High | 4h | P1-003, P1-005 |
| P1-007 | Write unit tests for strategy interfaces | Critical | 4h | P1-001 |
| P1-008 | Write unit tests for recipe builder | Critical | 4h | P1-003 |
| P1-009 | Write integration tests for QA strategy | High | 6h | P1-005 |
| P1-010 | Update LLMsVerifier API to expose strategy endpoints | High | 4h | P1-001-P1-006 |
| P1-011 | Create LLMsVerifier submodule in Catalogizer | High | 2h | None |
| P1-012 | Wire LLMsVerifier into HelixQA go.mod | High | 1h | P1-011 |

**Total Phase 1: ~53 hours**

---

## Phase 2: OpenCode Headless Integration (Week 2-3)

### 2.1 OpenCode Adapter Enhancements

#### 2.1.1 Enhanced OpenCode Configuration

**File: `LLMOrchestrator/pkg/adapter/opencode_config.go`**

```go
// OpenCodeConfig holds OpenCode-specific configuration
type OpenCodeConfig struct {
    BinaryPath       string            `json:"binary_path"`
    Provider         string            `json:"provider"`
    Model            string            `json:"model"`
    APIKey           string            `json:"-"` // Loaded from env
    WorkingDir       string            `json:"working_dir"`
    Headless         bool              `json:"headless"`
    NonInteractive   bool              `json:"non_interactive"`
    Timeout          time.Duration     `json:"timeout"`
    MaxTokens        int               `json:"max_tokens"`
    Temperature      float64           `json:"temperature"`
    SystemPrompt     string            `json:"system_prompt"`
    ExtraFlags       []string          `json:"extra_flags"`
    EnvVars          map[string]string `json:"env_vars"`
}

func DefaultOpenCodeConfig() *OpenCodeConfig {
    return &OpenCodeConfig{
        BinaryPath:     "opencode",
        Headless:       true,
        NonInteractive: true,
        Timeout:        120 * time.Second,
        MaxTokens:      4096,
        Temperature:    0.7,
        EnvVars:        make(map[string]string),
    }
}
```

#### 2.1.2 OpenCode Headless Mode Handler

**File: `LLMOrchestrator/pkg/adapter/opencode_headless.go`**

```go
// OpenCodeHeadless manages OpenCode in headless CLI mode
type OpenCodeHeadless struct {
    config   *OpenCodeConfig
    cmd      *exec.Cmd
    stdin    io.WriteCloser
    stdout   io.Reader
    stderr   io.Reader
    parser   *OpenCodeParser
    mu       sync.Mutex
    running  bool
}

func (h *OpenCodeHeadless) Start(ctx context.Context) error {
    args := h.buildArgs()
    
    h.cmd = exec.CommandContext(ctx, h.config.BinaryPath, args...)
    h.cmd.Dir = h.config.WorkingDir
    h.cmd.Env = h.buildEnv()
    
    stdin, err := h.cmd.StdinPipe()
    if err != nil {
        return fmt.Errorf("stdin pipe: %w", err)
    }
    h.stdin = stdin
    
    stdout, err := h.cmd.StdoutPipe()
    if err != nil {
        return fmt.Errorf("stdout pipe: %w", err)
    }
    h.stdout = stdout
    
    stderr, err := h.cmd.StderrPipe()
    if err != nil {
        return fmt.Errorf("stderr pipe: %w", err)
    }
    h.stderr = stderr
    
    if err := h.cmd.Start(); err != nil {
        return fmt.Errorf("start command: %w", err)
    }
    
    h.running = true
    return nil
}

func (h *OpenCodeHeadless) buildArgs() []string {
    args := []string{}
    
    if h.config.Headless {
        args = append(args, "--headless")
    }
    if h.config.NonInteractive {
        args = append(args, "--non-interactive")
    }
    if h.config.Provider != "" {
        args = append(args, "--provider", h.config.Provider)
    }
    if h.config.Model != "" {
        args = append(args, "--model", h.config.Model)
    }
    
    args = append(args, h.config.ExtraFlags...)
    
    return args
}
```

### 2.2 Multi-Agent Support

#### 2.2.1 Agent Pool with Multiple Provider Types

**File: `LLMOrchestrator/pkg/agent/multi_pool.go`**

```go
// MultiProviderPool manages agents from multiple CLI providers
type MultiProviderPool struct {
    pools    map[string]agent.AgentPool
    selector AgentSelector
    mu       sync.RWMutex
}

func NewMultiProviderPool(configs map[string]*PoolConfig) (*MultiProviderPool, error) {
    pools := make(map[string]agent.AgentPool)
    
    for provider, cfg := range configs {
        switch provider {
        case "opencode":
            pools[provider] = NewOpenCodePool(cfg)
        case "claude-code":
            pools[provider] = NewClaudeCodePool(cfg)
        case "gemini":
            pools[provider] = NewGeminiPool(cfg)
        case "junie":
            pools[provider] = NewJuniePool(cfg)
        case "qwen-code":
            pools[provider] = NewQwenCodePool(cfg)
        default:
            return nil, fmt.Errorf("unknown provider: %s", provider)
        }
    }
    
    return &MultiProviderPool{
        pools:    pools,
        selector: NewRoundRobinSelector(),
    }, nil
}

func (m *MultiProviderPool) Acquire(ctx context.Context, req agent.AgentRequirements) (agent.Agent, error) {
    m.mu.RLock()
    defer m.mu.RUnlock()
    
    // Find best pool based on requirements
    selected := m.selector.Select(m.pools, req)
    if selected == "" {
        return nil, ErrNoSuitableAgent
    }
    
    return m.pools[selected].Acquire(ctx, req)
}
```

### 2.3 Tasks for Phase 2

| ID | Task | Priority | Est. Time | Dependencies |
|----|------|----------|-----------|--------------|
| P2-001 | Create `opencode_config.go` with enhanced configuration | Critical | 4h | P1 |
| P2-002 | Create `opencode_headless.go` with headless mode handler | Critical | 8h | P2-001 |
| P2-003 | Enhance OpenCode parser for headless output | High | 6h | P2-002 |
| P2-004 | Create `multi_pool.go` for multi-provider support | Critical | 8h | P2-002 |
| P2-005 | Implement agent selector strategies | High | 4h | P2-004 |
| P2-006 | Add Claude Code headless enhancements | High | 4h | P2-002 |
| P2-007 | Add Gemini headless enhancements | High | 4h | P2-002 |
| P2-008 | Add Junie headless enhancements | Medium | 4h | P2-002 |
| P2-009 | Add Qwen Code headless enhancements | Medium | 4h | P2-002 |
| P2-010 | Write unit tests for headless adapters | Critical | 6h | P2-002-P2-009 |
| P2-011 | Write integration tests for multi-provider pool | Critical | 6h | P2-004 |
| P2-012 | Update LLMOrchestrator API for multi-provider | High | 4h | P2-004 |

**Total Phase 2: ~62 hours**

---

## Phase 3: Enhanced Autonomous Session (Week 3-5)

### 3.1 Documentation Loading Pipeline

#### 3.1.1 Enhanced DocProcessor Integration

The existing DocProcessor already provides:
- Markdown/YAML/HTML parsing
- Feature extraction via LLM
- Coverage tracking

Enhancements needed:
- Support for ADOC, RST, DITA formats
- Video course transcript parsing
- API documentation parsing (OpenAPI, GraphQL schemas)
- Code comment extraction

#### 3.1.2 Feature-to-Test Mapping

**File: `HelixQA/pkg/autonomous/mapper.go`**

```go
// FeatureTestMapper maps documentation features to executable tests
type FeatureTestMapper struct {
    docProcessor *docprocessor.Processor
    llmAgent     llm.LLMAgent
    cache        *FeatureCache
}

type FeatureTestMapping struct {
    FeatureID    string       `json:"feature_id"`
    FeatureName  string       `json:"feature_name"`
    Platforms    []string     `json:"platforms"`
    TestSteps    []TestStep   `json:"test_steps"`
    ExpectedUI   []UIElement  `json:"expected_ui"`
    Preconditions []string    `json:"preconditions"`
    Priority     string       `json:"priority"`
    Category     string       `json:"category"`
}

func (m *FeatureTestMapper) MapFeature(ctx context.Context, f feature.Feature) (*FeatureTestMapping, error) {
    // Check cache first
    if cached, ok := m.cache.Get(f.ID); ok {
        return cached, nil
    }
    
    // Use LLM to generate test steps
    steps, err := m.llmAgent.GenerateTestSteps(ctx, f)
    if err != nil {
        return nil, fmt.Errorf("generate steps: %w", err)
    }
    
    // Infer expected UI elements
    uiElements, err := m.llmAgent.InferUIElements(ctx, f)
    if err != nil {
        return nil, fmt.Errorf("infer ui: %w", err)
    }
    
    mapping := &FeatureTestMapping{
        FeatureID:     f.ID,
        FeatureName:   f.Name,
        Platforms:     f.Platforms,
        TestSteps:     steps,
        ExpectedUI:    uiElements,
        Preconditions: f.Preconditions,
        Priority:      f.Priority,
        Category:      string(f.Category),
    }
    
    m.cache.Set(f.ID, mapping)
    return mapping, nil
}
```

### 3.2 Enhanced Navigation Engine

#### 3.2.1 LLM-Powered Navigation

**File: `HelixQA/pkg/navigator/llm_navigator.go`**

```go
// LLMNavigator uses LLM reasoning for intelligent navigation
type LLMNavigator struct {
    agent      agent.Agent
    analyzer   analyzer.Analyzer
    graph      *NavigationGraph
    state      *StateTracker
    executor   ActionExecutor
    history    *NavigationHistory
}

func (n *LLMNavigator) NavigateToFeature(ctx context.Context, featureID string) error {
    // 1. Check if we know how to reach this screen
    targetScreen := n.graph.ScreenForFeature(featureID)
    if targetScreen == nil {
        // Ask LLM to infer the navigation path
        path, err := n.inferPath(ctx, featureID)
        if err != nil {
            return fmt.Errorf("infer path: %w", err)
        }
        targetScreen = path.Destination
    }
    
    // 2. Compute shortest path from current screen
    current := n.state.CurrentScreen()
    path := n.graph.ShortestPath(current.ID, targetScreen.ID)
    
    // 3. Execute navigation actions
    for _, action := range path.Actions {
        if err := n.executeAction(ctx, action); err != nil {
            return fmt.Errorf("execute %s: %w", action.Type, err)
        }
        
        // Verify screen after each action
        screen, err := n.verifyScreen(ctx)
        if err != nil {
            return fmt.Errorf("verify screen: %w", err)
        }
        
        // Update graph with discovered navigation
        n.graph.AddTransition(current.ID, screen.ID, action)
        current = screen
    }
    
    return nil
}

func (n *LLMNavigator) inferPath(ctx context.Context, featureID string) (*NavigationPath, error) {
    prompt := fmt.Sprintf(`
        I need to navigate to a screen that contains feature "%s".
        
        Current screen: %s
        Known screens: %v
        
        What navigation actions should I take? Return a JSON array of actions.
    `, featureID, n.state.CurrentScreen().Name, n.graph.KnownScreenNames())
    
    resp, err := n.agent.Send(ctx, prompt)
    if err != nil {
        return nil, err
    }
    
    return n.parseNavigationPath(resp.Content)
}
```

### 3.3 Enhanced Issue Detection

#### 3.3.1 Multi-Category Issue Detection

**File: `HelixQA/pkg/issuedetector/categories.go`**

```go
// IssueCategory represents different types of issues
type IssueCategory string

const (
    CategoryCrash        IssueCategory = "crash"
    CategoryANR          IssueCategory = "anr"
    CategoryVisualBug    IssueCategory = "visual_bug"
    CategoryUXIssue      IssueCategory = "ux_issue"
    CategoryAccessibility IssueCategory = "accessibility"
    CategoryPerformance  IssueCategory = "performance"
    CategorySecurity     IssueCategory = "security"
    CategoryFunctional   IssueCategory = "functional"
    CategoryDataLoss     IssueCategory = "data_loss"
)

// IssueSeverity levels
type IssueSeverity string

const (
    SeverityCritical IssueSeverity = "critical"
    SeverityHigh     IssueSeverity = "high"
    SeverityMedium   IssueSeverity = "medium"
    SeverityLow      IssueSeverity = "low"
    SeverityInfo     IssueSeverity = "info"
)

// Issue represents a detected problem
type Issue struct {
    ID              string            `json:"id"`
    Title           string            `json:"title"`
    Description     string            `json:"description"`
    Category        IssueCategory     `json:"category"`
    Severity        IssueSeverity     `json:"severity"`
    Platform        string            `json:"platform"`
    FeatureID       string            `json:"feature_id,omitempty"`
    ReproductionSteps []string        `json:"reproduction_steps"`
    ExpectedBehavior string           `json:"expected_behavior"`
    ActualBehavior   string           `json:"actual_behavior"`
    Evidence        *Evidence         `json:"evidence"`
    LLMAnalysis     string            `json:"llm_analysis,omitempty"`
    SuggestedFix    string            `json:"suggested_fix,omitempty"`
    Timestamp       time.Time         `json:"timestamp"`
    VideoTimestamp  time.Duration     `json:"video_timestamp"`
}
```

#### 3.3.2 LLM-Powered Issue Analysis

**File: `HelixQA/pkg/issuedetector/llm_analyzer.go`**

```go
// LLMIssueAnalyzer uses LLM to classify and analyze issues
type LLMIssueAnalyzer struct {
    agent    agent.Agent
    prompts  *PromptTemplates
}

func (a *LLMIssueAnalyzer) AnalyzeIssue(ctx context.Context, before, after []byte, action string) (*Issue, error) {
    prompt := a.prompts.BuildAnalysisPrompt(before, after, action)
    
    resp, err := a.agent.SendWithAttachments(ctx, prompt, []agent.Attachment{
        {Type: "image", Data: before, MIME: "image/png"},
        {Type: "image", Data: after, MIME: "image/png"},
    })
    if err != nil {
        return nil, fmt.Errorf("LLM analysis: %w", err)
    }
    
    var issue Issue
    if err := json.Unmarshal([]byte(resp.Content), &issue); err != nil {
        return nil, fmt.Errorf("parse issue: %w", err)
    }
    
    issue.ID = generateIssueID()
    issue.Timestamp = time.Now()
    
    return &issue, nil
}

func (a *LLMIssueAnalyzer) SuggestFix(ctx context.Context, issue *Issue) (string, error) {
    prompt := fmt.Sprintf(`
        Analyze this QA issue and suggest a fix:
        
        Title: %s
        Description: %s
        Category: %s
        Severity: %s
        Platform: %s
        Steps to reproduce: %v
        
        Provide a specific, actionable fix recommendation.
    `, issue.Title, issue.Description, issue.Category, issue.Severity, 
       issue.Platform, issue.ReproductionSteps)
    
    resp, err := a.agent.Send(ctx, prompt)
    if err != nil {
        return "", err
    }
    
    return resp.Content, nil
}
```

### 3.4 Enhanced Evidence Collection

#### 3.4.1 Session Recording with Timeline

**File: `HelixQA/pkg/session/recorder.go`**

```go
// SessionRecorder captures all evidence during a session
type SessionRecorder struct {
    sessionID    string
    outputDir    string
    videos       map[string]*VideoRecorder
    screenshots  *ScreenshotManager
    timeline     *Timeline
    logs         *LogCollector
    mu           sync.Mutex
}

type TimelineEvent struct {
    ID             string        `json:"id"`
    Type           EventType     `json:"type"`
    Platform       string        `json:"platform"`
    Timestamp      time.Time     `json:"timestamp"`
    VideoOffset    time.Duration `json:"video_offset"`
    ScreenshotPath string        `json:"screenshot_path,omitempty"`
    FeatureID      string        `json:"feature_id,omitempty"`
    IssueID        string        `json:"issue_id,omitempty"`
    Action         string        `json:"action,omitempty"`
    Result         string        `json:"result,omitempty"`
    Metadata       map[string]any `json:"metadata,omitempty"`
}

func (r *SessionRecorder) RecordAction(ctx context.Context, platform, action string, result error) (string, error) {
    r.mu.Lock()
    defer r.mu.Unlock()
    
    // Capture screenshot
    screenshot, err := r.screenshots.Capture(ctx, platform)
    if err != nil {
        return "", fmt.Errorf("capture: %w", err)
    }
    
    // Get video timestamp
    videoOffset := r.videos[platform].CurrentOffset()
    
    // Create timeline event
    event := TimelineEvent{
        ID:             uuid.New().String(),
        Type:           EventTypeAction,
        Platform:       platform,
        Timestamp:      time.Now(),
        VideoOffset:    videoOffset,
        ScreenshotPath: screenshot.Path,
        Action:         action,
        Result:         errorToString(result),
    }
    
    r.timeline.Add(event)
    
    return event.ID, nil
}

func (r *SessionRecorder) RecordIssue(ctx context.Context, platform string, issue *Issue) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    
    // Capture annotated screenshot
    annotated, err := r.screenshots.CaptureAnnotated(ctx, platform, issue)
    if err != nil {
        return fmt.Errorf("capture annotated: %w", err)
    }
    
    videoOffset := r.videos[platform].CurrentOffset()
    
    event := TimelineEvent{
        ID:             uuid.New().String(),
        Type:           EventTypeIssue,
        Platform:       platform,
        Timestamp:      time.Now(),
        VideoOffset:    videoOffset,
        ScreenshotPath: annotated.Path,
        IssueID:        issue.ID,
        Metadata: map[string]any{
            "severity": issue.Severity,
            "category": issue.Category,
            "title":    issue.Title,
        },
    }
    
    r.timeline.Add(event)
    issue.VideoTimestamp = videoOffset
    issue.Evidence.Screenshots = append(issue.Evidence.Screenshots, annotated.Path)
    
    return nil
}
```

### 3.5 Enhanced Ticket Generation

#### 3.5.1 Comprehensive Ticket Format

**File: `HelixQA/pkg/ticket/generator.go`**

```go
// TicketGenerator creates detailed markdown tickets
type TicketGenerator struct {
    outputDir string
    templates *TicketTemplates
}

func (g *TicketGenerator) Generate(issue *Issue, session *SessionInfo) (*Ticket, error) {
    ticket := &Ticket{
        ID:        issue.ID,
        Title:     issue.Title,
        CreatedAt: time.Now(),
        Status:    "open",
        Labels:    []string{string(issue.Category), string(issue.Severity), issue.Platform},
    }
    
    var buf bytes.Buffer
    if err := g.templates.Ticket.Execute(&buf, map[string]any{
        "Issue":   issue,
        "Session": session,
    }); err != nil {
        return nil, fmt.Errorf("execute template: %w", err)
    }
    
    ticket.Content = buf.String()
    ticket.Path = filepath.Join(g.outputDir, ticket.ID+".md")
    
    if err := os.WriteFile(ticket.Path, []byte(ticket.Content), 0644); err != nil {
        return nil, fmt.Errorf("write ticket: %w", err)
    }
    
    return ticket, nil
}
```

**Ticket Template:**

```markdown
# {{.Issue.ID}}: {{.Issue.Title}}

**Severity:** {{.Issue.Severity}} | **Platform:** {{.Issue.Platform}} | **Category:** {{.Issue.Category}}

## Session Information

- **Session ID:** {{.Session.ID}}
- **Started:** {{.Session.StartTime}}
- **Feature Under Test:** {{.Issue.FeatureID}}

## Steps to Reproduce

{{range $i, $step := .Issue.ReproductionSteps}}
{{add $i 1}}. {{$step}}
{{end}}

## Expected Behavior

{{.Issue.ExpectedBehavior}}

## Actual Behavior

{{.Issue.ActualBehavior}}

## Evidence

### Screenshots

{{range .Issue.Evidence.Screenshots}}
- [{{.}}]({{.}})
{{end}}

### Video Reference

- Video: `{{.Session.VideoPath}}` @ `{{formatDuration .Issue.VideoTimestamp}}`
  ```bash
  ffplay -ss {{formatSeconds .Issue.VideoTimestamp}} {{.Session.VideoPath}}
  ```

### Logs

```
{{.Issue.Evidence.Logs}}
```

## LLM Analysis

{{.Issue.LLMAnalysis}}

## Suggested Fix

{{.Issue.SuggestedFix}}

---

*Generated by HelixQA Autonomous Session on {{now}}*
```

### 3.6 Tasks for Phase 3

| ID | Task | Priority | Est. Time | Dependencies |
|----|------|----------|-----------|--------------|
| P3-001 | Create `mapper.go` for feature-to-test mapping | Critical | 6h | P1 |
| P3-002 | Add DocProcessor format support (ADOC, RST) | High | 4h | P1 |
| P3-003 | Create `llm_navigator.go` for LLM-powered navigation | Critical | 8h | P2 |
| P3-004 | Enhance NavigationGraph with LLM inference | High | 6h | P3-003 |
| P3-005 | Create `llm_analyzer.go` for issue analysis | Critical | 8h | P2 |
| P3-006 | Expand issue categories and severity levels | High | 4h | P3-005 |
| P3-007 | Create enhanced `recorder.go` with timeline | Critical | 8h | P1 |
| P3-008 | Create annotated screenshot capture | High | 4h | P3-007 |
| P3-009 | Create enhanced `generator.go` for tickets | Critical | 6h | P3-005 |
| P3-010 | Create ticket templates | High | 2h | P3-009 |
| P3-011 | Wire all components into SessionCoordinator | Critical | 8h | P3-001-P3-010 |
| P3-012 | Add configuration options to .env | High | 2h | P3-011 |
| P3-013 | Write unit tests for mapper | Critical | 4h | P3-001 |
| P3-014 | Write unit tests for navigator | Critical | 4h | P3-003 |
| P3-015 | Write unit tests for issue detector | Critical | 4h | P3-005 |
| P3-016 | Write integration tests for full session | Critical | 8h | P3-011 |

**Total Phase 3: ~76 hours**

---

## Phase 4: Configuration & Environment (Week 5)

### 4.1 Environment Configuration

#### 4.1.1 Comprehensive .env Support

**File: `HelixQA/.env.example`**

```bash
# =============================================================================
# HelixQA Autonomous QA Session Configuration
# =============================================================================

# -----------------------------------------------------------------------------
# LLM Provider Configuration
# -----------------------------------------------------------------------------

# OpenAI
OPENAI_API_KEY=sk-...
OPENAI_MODEL=gpt-4o

# Anthropic
ANTHROPIC_API_KEY=sk-ant-...
ANTHROPIC_MODEL=claude-3.5-sonnet

# Google Gemini
GOOGLE_API_KEY=...
GOOGLE_MODEL=gemini-2.0-flash

# Groq
GROQ_API_KEY=...
GROQ_MODEL=llama-3.3-70b-versatile

# DeepSeek
DEEPSEEK_API_KEY=...
DEEPSEEK_MODEL=deepseek-chat

# xAI
XAI_API_KEY=...
XAI_MODEL=grok-beta

# Qwen
QWEN_API_KEY=...
QWEN_MODEL=qwen-max

# -----------------------------------------------------------------------------
# CLI Agent Configuration
# -----------------------------------------------------------------------------

# Comma-separated list of enabled agents
HELIX_AGENTS_ENABLED=opencode,claude-code,gemini

# Agent binary paths
HELIX_AGENT_OPENCODE_PATH=/usr/local/bin/opencode
HELIX_AGENT_CLAUDE_PATH=/usr/local/bin/claude
HELIX_AGENT_GEMINI_PATH=/usr/local/bin/gemini
HELIX_AGENT_JUNIE_PATH=/usr/local/bin/junie
HELIX_AGENT_QWEN_PATH=/usr/local/bin/qwen-code

# Agent pool configuration
HELIX_AGENT_POOL_SIZE=3
HELIX_AGENT_TIMEOUT=120s
HELIX_AGENT_MAX_RETRIES=3

# Preferred agents per platform (optional)
HELIX_ANDROID_PREFERRED_AGENT=claude-code
HELIX_DESKTOP_PREFERRED_AGENT=opencode
HELIX_WEB_PREFERRED_AGENT=gemini

# -----------------------------------------------------------------------------
# LLMsVerifier Strategy Configuration
# -----------------------------------------------------------------------------

# Verification strategy: default, qa, speed, quality, cost
HELIX_VERIFIER_STRATEGY=qa

# Minimum score threshold (0-1)
HELIX_VERIFIER_MIN_SCORE=0.7

# Recipe configuration
HELIX_RECIPE_VISION_WEIGHT=0.4
HELIX_RECIPE_SPEED_WEIGHT=0.3
HELIX_RECIPE_COST_WEIGHT=0.3

# -----------------------------------------------------------------------------
# Vision Configuration
# -----------------------------------------------------------------------------

# OpenCV settings
HELIX_VISION_OPENCV_ENABLED=true
HELIX_VISION_SSIM_THRESHOLD=0.95

# LLM Vision provider (auto selects best available)
HELIX_VISION_LLM_PROVIDER=anthropic

# -----------------------------------------------------------------------------
# Autonomous Session Configuration
# -----------------------------------------------------------------------------

# Enable autonomous test type (default: true)
HELIX_AUTONOMOUS_ENABLED=true

# Target platforms (comma-separated): android,desktop,web,api
HELIX_AUTONOMOUS_PLATFORMS=desktop

# Session timeout
HELIX_AUTONOMOUS_TIMEOUT=2h

# Coverage target (0-1)
HELIX_AUTONOMOUS_COVERAGE_TARGET=0.90

# Curiosity-driven exploration
HELIX_AUTONOMOUS_CURIOSITY_ENABLED=true
HELIX_AUTONOMOUS_CURIOSITY_TIMEOUT=30m

# -----------------------------------------------------------------------------
# Documentation Processing
# -----------------------------------------------------------------------------

# Documentation root directory
HELIX_DOCS_ROOT=./docs

# Auto-discover documentation
HELIX_DOCS_AUTO_DISCOVER=true

# Supported formats (comma-separated)
HELIX_DOCS_FORMATS=md,yaml,html,adoc,rst

# -----------------------------------------------------------------------------
# Recording Configuration
# -----------------------------------------------------------------------------

# Video recording
HELIX_RECORDING_VIDEO=true
HELIX_RECORDING_FFMPEG_PATH=/usr/bin/ffmpeg

# Screenshot quality (1-100)
HELIX_RECORDING_SCREENSHOT_QUALITY=90

# Video format and codec
HELIX_RECORDING_VIDEO_FORMAT=mp4
HELIX_RECORDING_VIDEO_CODEC=libx264

# -----------------------------------------------------------------------------
# Ticket Configuration
# -----------------------------------------------------------------------------

# Enable ticket generation
HELIX_TICKETS_ENABLED=true

# Minimum severity to generate ticket: critical,high,medium,low,info
HELIX_TICKETS_MIN_SEVERITY=low

# Auto-assign tickets
HELIX_TICKETS_AUTO_ASSIGN=false

# -----------------------------------------------------------------------------
# Report Configuration
# -----------------------------------------------------------------------------

# Output directory
HELIX_OUTPUT_DIR=./qa-results

# Report formats (comma-separated): markdown,html,json
HELIX_REPORT_FORMATS=markdown,html,json

# Include LLM analysis in reports
HELIX_REPORT_INCLUDE_LLM_ANALYSIS=true

# -----------------------------------------------------------------------------
# Platform-Specific Configuration
# -----------------------------------------------------------------------------

# Android
HELIX_ANDROID_DEVICE=emulator-5554
HELIX_ANDROID_PACKAGE=com.example.app
HELIX_ANDROID_ACTIVITY=.MainActivity

# Desktop (Linux)
HELIX_DESKTOP_PROCESS=myapp
HELIX_DESKTOP_DISPLAY=:0

# Web
HELIX_WEB_URL=http://localhost:3000
HELIX_WEB_BROWSER=chromium

# API
HELIX_API_BASE_URL=http://localhost:8080
HELIX_API_TIMEOUT=30s

# -----------------------------------------------------------------------------
# Logging Configuration
# -----------------------------------------------------------------------------

# Log level: debug,info,warn,error
HELIX_LOG_LEVEL=info

# Log file (empty for stdout)
HELIX_LOG_FILE=

# -----------------------------------------------------------------------------
# Resource Limits
# -----------------------------------------------------------------------------

# Maximum memory per platform worker (MB)
HELIX_MAX_MEMORY_MB=512

# Maximum goroutines per worker
HELIX_MAX_GOROUTINES=10
```

### 4.2 Configuration Loader

**File: `HelixQA/pkg/config/loader.go`**

```go
// LoadFromEnv loads configuration from environment variables
func LoadFromEnv() (*Config, error) {
    cfg := DefaultConfig()
    
    // LLM Provider keys
    cfg.LLM.OpenAIKey = os.Getenv("OPENAI_API_KEY")
    cfg.LLM.AnthropicKey = os.Getenv("ANTHROPIC_API_KEY")
    cfg.LLM.GoogleKey = os.Getenv("GOOGLE_API_KEY")
    cfg.LLM.GroqKey = os.Getenv("GROQ_API_KEY")
    cfg.LLM.DeepSeekKey = os.Getenv("DEEPSEEK_API_KEY")
    cfg.LLM.XAIKey = os.Getenv("XAI_API_KEY")
    cfg.LLM.QwenKey = os.Getenv("QWEN_API_KEY")
    
    // CLI Agents
    cfg.Agents.Enabled = parseStringSlice(os.Getenv("HELIX_AGENTS_ENABLED"))
    cfg.Agents.OpenCodePath = getEnvOrDefault("HELIX_AGENT_OPENCODE_PATH", "opencode")
    cfg.Agents.ClaudePath = getEnvOrDefault("HELIX_AGENT_CLAUDE_PATH", "claude")
    cfg.Agents.GeminiPath = getEnvOrDefault("HELIX_AGENT_GEMINI_PATH", "gemini")
    cfg.Agents.PoolSize = parseInt(os.Getenv("HELIX_AGENT_POOL_SIZE"), 1)
    cfg.Agents.Timeout = parseDuration(os.Getenv("HELIX_AGENT_TIMEOUT"), 120*time.Second)
    
    // Verifier Strategy
    cfg.Verifier.Strategy = getEnvOrDefault("HELIX_VERIFIER_STRATEGY", "qa")
    cfg.Verifier.MinScore = parseFloat(os.Getenv("HELIX_VERIFIER_MIN_SCORE"), 0.7)
    
    // Autonomous Session
    cfg.Autonomous.Enabled = parseBool(os.Getenv("HELIX_AUTONOMOUS_ENABLED"), true)
    cfg.Autonomous.Platforms = parseStringSlice(os.Getenv("HELIX_AUTONOMOUS_PLATFORMS"))
    cfg.Autonomous.Timeout = parseDuration(os.Getenv("HELIX_AUTONOMOUS_TIMEOUT"), 2*time.Hour)
    cfg.Autonomous.CoverageTarget = parseFloat(os.Getenv("HELIX_AUTONOMOUS_COVERAGE_TARGET"), 0.9)
    cfg.Autonomous.CuriosityEnabled = parseBool(os.Getenv("HELIX_AUTONOMOUS_CURIOSITY_ENABLED"), true)
    
    // Recording
    cfg.Recording.VideoEnabled = parseBool(os.Getenv("HELIX_RECORDING_VIDEO"), true)
    cfg.Recording.FFmpegPath = getEnvOrDefault("HELIX_RECORDING_FFMPEG_PATH", "ffmpeg")
    
    // Tickets
    cfg.Tickets.Enabled = parseBool(os.Getenv("HELIX_TICKETS_ENABLED"), true)
    cfg.Tickets.MinSeverity = getEnvOrDefault("HELIX_TICKETS_MIN_SEVERITY", "low")
    
    // Output
    cfg.Output.Dir = getEnvOrDefault("HELIX_OUTPUT_DIR", "./qa-results")
    cfg.Output.Formats = parseStringSlice(os.Getenv("HELIX_REPORT_FORMATS"))
    
    return cfg, cfg.Validate()
}
```

### 4.3 Tasks for Phase 4

| ID | Task | Priority | Est. Time | Dependencies |
|----|------|----------|-----------|--------------|
| P4-001 | Create comprehensive .env.example | Critical | 4h | P3 |
| P4-002 | Create `loader.go` for env configuration | Critical | 6h | P4-001 |
| P4-003 | Create validation for all config options | High | 4h | P4-002 |
| P4-004 | Create config documentation | High | 2h | P4-001 |
| P4-005 | Write unit tests for config loader | Critical | 4h | P4-002 |

**Total Phase 4: ~20 hours**

---

## Phase 5: Comprehensive Testing (Week 5-6)

### 5.1 Test Categories

#### 5.1.1 Unit Tests

| Package | Target Coverage | Focus Areas |
|---------|-----------------|-------------|
| `pkg/strategy` | 95% | Interface contracts, scoring algorithms |
| `pkg/recipe` | 95% | Builder pattern, validation |
| `pkg/autonomous` | 90% | Coordinator phases, worker lifecycle |
| `pkg/navigator` | 90% | Path finding, action execution |
| `pkg/issuedetector` | 95% | Category detection, LLM analysis |
| `pkg/session` | 90% | Recording, timeline management |
| `pkg/ticket` | 90% | Generation, formatting |
| `pkg/config` | 95% | Loading, validation |

#### 5.1.2 Integration Tests

| Test Suite | Description |
|------------|-------------|
| `TestFullSessionLifecycle` | Complete 4-phase session with mock apps |
| `TestMultiPlatformParallel` | 3 platforms running simultaneously |
| `TestLLMIntegration` | Real LLM calls (with recording for replay) |
| `TestVisionIntegration` | OpenCV + LLM Vision analysis |
| `TestEvidenceCollection` | Screenshot, video, log collection |
| `TestTicketGeneration` | End-to-end ticket creation |

#### 5.1.3 E2E Tests

| Test Scenario | Platform | Duration |
|---------------|----------|----------|
| Catalogizer Desktop QA | Desktop | 30 min |
| Catalogizer Web QA | Web | 20 min |
| Catalogizer API QA | API | 10 min |
| Multi-Platform Combined | All | 1 hour |

#### 5.1.4 Security Tests

| Test | Description |
|------|-------------|
| `TestPromptInjectionSanitization` | Verify malicious prompts are blocked |
| `TestAPIKeyHandling` | Ensure keys are never logged or exposed |
| `TestFileAccessControl` | Verify output files have correct permissions |
| `TestCommandInjectionPrevention` | Ensure CLI commands are sanitized |

#### 5.1.5 Stress Tests

| Test | Description |
|------|-------------|
| `TestConcurrentWorkers` | 10 workers per platform for 1 hour |
| `TestLongRunningSession` | 8-hour session with memory monitoring |
| `TestHighVolumeScreenshots` | 1000 screenshots in 10 minutes |
| `TestLLMRateLimiting` | Handle rate limits gracefully |

#### 5.1.6 Automation Tests

| Test | Description |
|------|-------------|
| `TestAndroidAutomation` | Full Playwright/ADB automation |
| `TestDesktopAutomation` | xdotool/xdo automation |
| `TestWebAutomation` | Playwright browser automation |
| `TestAPIAutomation` | HTTP client automation |

### 5.2 Test Infrastructure

#### 5.2.1 Mock Implementations

```go
// Mock implementations for testing

type MockAgent struct {
    responses map[string]string
}

func (m *MockAgent) Send(ctx context.Context, prompt string) (Response, error) {
    if resp, ok := m.responses[prompt]; ok {
        return Response{Content: resp}, nil
    }
    return Response{Content: "default mock response"}, nil
}

type MockAnalyzer struct{}

func (m *MockAnalyzer) Analyze(ctx context.Context, img []byte) (*ScreenAnalysis, error) {
    return &ScreenAnalysis{
        ScreenID:   "mock-screen",
        Components: []UIComponent{},
    }, nil
}

type MockExecutor struct {
    clicks []ClickEvent
}

func (m *MockExecutor) Click(ctx context.Context, x, y int) error {
    m.clicks = append(m.clicks, ClickEvent{X: x, Y: y, Time: time.Now()})
    return nil
}
```

### 5.3 Tasks for Phase 5

| ID | Task | Priority | Est. Time | Dependencies |
|----|------|----------|-----------|--------------|
| P5-001 | Create mock implementations for all interfaces | Critical | 8h | P3 |
| P5-002 | Write unit tests for strategy package | Critical | 6h | P5-001 |
| P5-003 | Write unit tests for recipe package | Critical | 4h | P5-001 |
| P5-004 | Write unit tests for autonomous package | Critical | 8h | P5-001 |
| P5-005 | Write unit tests for navigator package | Critical | 6h | P5-001 |
| P5-006 | Write unit tests for issuedetector package | Critical | 6h | P5-001 |
| P5-007 | Write unit tests for session package | Critical | 4h | P5-001 |
| P5-008 | Write unit tests for ticket package | Critical | 4h | P5-001 |
| P5-009 | Write integration tests for full session | Critical | 8h | P5-001-P5-008 |
| P5-010 | Write integration tests for multi-platform | High | 6h | P5-009 |
| P5-011 | Write E2E tests for Catalogizer desktop | High | 8h | P5-009 |
| P5-012 | Write E2E tests for Catalogizer web | High | 6h | P5-009 |
| P5-013 | Write security tests | Critical | 6h | P5-001 |
| P5-014 | Write stress tests | High | 6h | P5-009 |
| P5-015 | Write automation tests | High | 8h | P5-009 |
| P5-016 | Set up CI test runner (local only) | High | 4h | P5-001-P5-015 |

**Total Phase 5: ~88 hours**

---

## Phase 6: Documentation (Week 6-7)

### 6.1 Documentation Structure

```
docs/
├── user-guides/
│   ├── getting-started.md
│   ├── configuration.md
│   ├── running-sessions.md
│   ├── understanding-reports.md
│   ├── troubleshooting.md
│   └── best-practices.md
├── developer-guides/
│   ├── architecture.md
│   ├── extending-helixqa.md
│   ├── adding-platforms.md
│   ├── custom-strategies.md
│   └── contributing.md
├── api-reference/
│   ├── strategy-api.md
│   ├── recipe-api.md
│   ├── session-api.md
│   ├── navigator-api.md
│   └── ticket-api.md
├── diagrams/
│   ├── architecture.svg
│   ├── sequence-diagrams.svg
│   ├── class-diagrams.svg
│   └── flowcharts.svg
└── video-course/
    ├── module-01-introduction/
    ├── module-02-configuration/
    ├── module-03-running-sessions/
    ├── module-04-advanced-features/
    └── module-05-troubleshooting/
```

### 6.2 User Guide Topics

1. **Getting Started**
   - Installation
   - Prerequisites
   - First session
   - Understanding output

2. **Configuration**
   - Environment variables
   - Strategy selection
   - Platform configuration
   - LLM provider setup

3. **Running Sessions**
   - Command-line options
   - Monitoring progress
   - Pausing/resuming
   - Handling failures

4. **Understanding Reports**
   - Report structure
   - Coverage metrics
   - Ticket format
   - Video evidence

5. **Troubleshooting**
   - Common issues
   - Debug mode
   - Log analysis
   - Recovery procedures

### 6.3 Video Course Outline

**Module 1: Introduction (15 min)**
- What is HelixQA Autonomous Session?
- Key features and benefits
- Architecture overview

**Module 2: Configuration (25 min)**
- Setting up environment
- Configuring LLM providers
- CLI agent installation
- Platform-specific setup

**Module 3: Running Sessions (30 min)**
- Starting a session
- Monitoring progress
- Understanding phases
- Working with results

**Module 4: Advanced Features (35 min)**
- Custom strategies
- Multi-platform testing
- Curiosity-driven exploration
- Evidence management

**Module 5: Troubleshooting (20 min)**
- Common issues
- Debug techniques
- Performance tuning
- Getting help

### 6.4 Tasks for Phase 6

| ID | Task | Priority | Est. Time | Dependencies |
|----|------|----------|-----------|--------------|
| P6-001 | Write getting-started.md | Critical | 4h | P5 |
| P6-002 | Write configuration.md | Critical | 4h | P4 |
| P6-003 | Write running-sessions.md | Critical | 4h | P5 |
| P6-004 | Write understanding-reports.md | High | 4h | P5 |
| P6-005 | Write troubleshooting.md | High | 4h | P5 |
| P6-006 | Write architecture.md | Critical | 6h | P5 |
| P6-007 | Write extending-helixqa.md | High | 4h | P5 |
| P6-008 | Write api-reference documents | High | 8h | P5 |
| P6-009 | Create architecture diagrams | High | 4h | P6-006 |
| P6-010 | Create sequence diagrams | High | 4h | P6-006 |
| P6-011 | Create flowcharts | High | 4h | P6-006 |
| P6-012 | Record Module 1 video | Medium | 4h | P6-001 |
| P6-013 | Record Module 2 video | Medium | 4h | P6-002 |
| P6-014 | Record Module 3 video | Medium | 4h | P6-003 |
| P6-015 | Record Module 4 video | Medium | 4h | P6-006-P6-008 |
| P6-016 | Record Module 5 video | Medium | 4h | P6-005 |

**Total Phase 6: ~66 hours**

---

## Phase 7: Final Integration & Deployment (Week 7)

### 7.1 Integration Checklist

- [ ] All modules compile without warnings
- [ ] All tests pass (unit, integration, e2e, security, stress)
- [ ] Documentation is complete
- [ ] Video course is published
- [ ] GitHub/GitLab repos are synchronized
- [ ] Submodules are properly configured
- [ ] CI/CD is disabled (per project requirements)

### 7.2 Final Deliverables

1. **LLMsVerifier submodule** with Strategy pattern
2. **Enhanced LLMOrchestrator** with multi-provider support
3. **Enhanced HelixQA** with Autonomous QA Session
4. **Complete test suite** (100% pass rate)
5. **Documentation** (user guides, API docs)
6. **Video course** (5 modules)
7. **Diagrams** (architecture, sequence, flowcharts)

### 7.3 Tasks for Phase 7

| ID | Task | Priority | Est. Time | Dependencies |
|----|------|----------|-----------|--------------|
| P7-001 | Final code review | Critical | 8h | P6 |
| P7-002 | Fix all remaining issues | Critical | 8h | P7-001 |
| P7-003 | Run full test suite | Critical | 4h | P7-002 |
| P7-004 | Create release notes | High | 2h | P7-003 |
| P7-005 | Tag releases in all repos | High | 2h | P7-004 |
| P7-006 | Sync GitHub/GitLab | High | 2h | P7-005 |
| P7-007 | Create project completion report | High | 4h | P7-001-P7-006 |

**Total Phase 7: ~30 hours**

---

## Summary

### Total Estimated Effort

| Phase | Hours | Duration |
|-------|-------|----------|
| Phase 1: Foundation | 53h | Week 1-2 |
| Phase 2: OpenCode Integration | 62h | Week 2-3 |
| Phase 3: Enhanced Session | 76h | Week 3-5 |
| Phase 4: Configuration | 20h | Week 5 |
| Phase 5: Testing | 88h | Week 5-6 |
| Phase 6: Documentation | 66h | Week 6-7 |
| Phase 7: Final Integration | 30h | Week 7 |
| **Total** | **395h** | **~7 weeks** |

### Key Milestones

1. **Week 2**: Strategy pattern and Recipe builder complete
2. **Week 3**: OpenCode headless integration complete
3. **Week 5**: Enhanced autonomous session functional
4. **Week 6**: All tests passing
5. **Week 7**: Documentation and video course complete

### Risk Mitigation

| Risk | Mitigation |
|------|------------|
| LLM API rate limits | Implement backoff, caching, and fallback providers |
| OpenCV installation issues | Provide LLM Vision fallback |
| Platform-specific bugs | Extensive mock testing before platform tests |
| LLM response parsing failures | Multiple parser strategies, re-prompt fallback |
| Memory exhaustion | Resource limits, cleanup routines |

---

## Appendix A: Module Dependency Graph

```
HelixQA
├── LLMsVerifier (submodule)
│   └── pkg/strategy
│   └── pkg/recipe
│   └── pkg/helixqa
├── LLMOrchestrator (submodule)
│   └── pkg/adapter
│   └── pkg/agent
│   └── pkg/protocol
├── VisionEngine (submodule)
│   └── pkg/analyzer
│   └── pkg/opencv
│   └── pkg/llmvision
├── DocProcessor (submodule)
│   └── pkg/loader
│   └── pkg/feature
│   └── pkg/llm
└── Challenges (submodule)
    └── pkg/runner
    └── pkg/bank
```

## Appendix B: Configuration Reference

See `.env.example` for complete configuration options.

## Appendix C: API Reference

Generated from GoDoc comments in source code.

## Appendix D: Glossary

- **Autonomous QA Session**: LLM-powered testing session that navigates and tests applications automatically
- **Strategy Pattern**: Pluggable algorithm for LLM selection and scoring
- **Recipe**: Complete configuration for LLM verification
- **Feature Map**: Documentation-derived features to verify
- **Navigation Graph**: Discovered screens and transitions
- **Timeline**: Chronological event log with video timestamps
