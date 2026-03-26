# HelixQA Autonomous Robot — Phase 2 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the Learning Engine (`pkg/learning/`) and Planning Engine (`pkg/planning/`) that ingest project knowledge and generate comprehensive test plans.

**Architecture:** The Learning Engine reads project docs, codebase structure, git history, and prior QA sessions to build a `KnowledgeBase`. The Planning Engine uses LLM + KnowledgeBase to generate a prioritized `TestPlan` with test bank reconciliation. Both packages depend on Phase 1's `pkg/llm/` (Provider) and `pkg/memory/` (Store).

**Tech Stack:** Go 1.25, `pkg/llm.Provider` for LLM calls, `pkg/memory.Store` for session history, `os/filepath` + `path/filepath` for file walking, `os/exec` for git commands, `encoding/json` for parsing, `testify` for testing.

**Spec:** `docs/superpowers/specs/2026-03-26-helixqa-autonomous-robot-design.md` (Sections 2-3)

**Depends on:** Phase 1 (`pkg/llm/`, `pkg/memory/`) — already complete with 58 tests passing.

---

## File Structure

### New Files

| File | Responsibility |
|------|---------------|
| `pkg/learning/reader.go` | `ProjectReader` — walks project tree, reads docs, parses CLAUDE.md |
| `pkg/learning/git.go` | `GitAnalyzer` — recent commits, change hotspots, branch info |
| `pkg/learning/codebase.go` | `CodebaseMapper` — extracts routes, screens, API endpoints, components |
| `pkg/learning/knowledge.go` | `KnowledgeBase` struct + `Builder` that composes all readers |
| `pkg/learning/reader_test.go` | ProjectReader tests |
| `pkg/learning/git_test.go` | GitAnalyzer tests |
| `pkg/learning/codebase_test.go` | CodebaseMapper tests |
| `pkg/learning/knowledge_test.go` | KnowledgeBase builder tests |
| `pkg/planning/planner.go` | `TestPlanGenerator` — LLM-driven test plan generation |
| `pkg/planning/reconciler.go` | `BankReconciler` — diffs plan against existing test banks |
| `pkg/planning/ranker.go` | `PriorityRanker` — orders tests by criticality + history |
| `pkg/planning/types.go` | `TestPlan`, `PlannedTest`, `PlanStats` types |
| `pkg/planning/planner_test.go` | Planner tests with mock LLM |
| `pkg/planning/reconciler_test.go` | Reconciler tests |
| `pkg/planning/ranker_test.go` | Ranker tests |

---

## Task 1: Learning Types & KnowledgeBase Struct

**Files:**
- Create: `pkg/learning/knowledge.go`
- Test: `pkg/learning/knowledge_test.go`

- [ ] **Step 1: Write the failing test**

```go
// pkg/learning/knowledge_test.go
package learning

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKnowledgeBase_Empty(t *testing.T) {
	kb := NewKnowledgeBase()
	assert.Equal(t, 0, len(kb.Screens))
	assert.Equal(t, 0, len(kb.APIEndpoints))
	assert.Equal(t, 0, len(kb.Docs))
	assert.Equal(t, 0, len(kb.RecentChanges))
	assert.Empty(t, kb.ProjectName)
}

func TestKnowledgeBase_Summary(t *testing.T) {
	kb := NewKnowledgeBase()
	kb.ProjectName = "Catalogizer"
	kb.Screens = []Screen{
		{Name: "login", Platform: "web", Route: "/login"},
		{Name: "dashboard", Platform: "web", Route: "/dashboard"},
	}
	kb.APIEndpoints = []APIEndpoint{
		{Method: "GET", Path: "/health"},
	}

	summary := kb.Summary()
	assert.Contains(t, summary, "Catalogizer")
	assert.Contains(t, summary, "2 screens")
	assert.Contains(t, summary, "1 API endpoints")
}

func TestKnowledgeBase_AddScreen(t *testing.T) {
	kb := NewKnowledgeBase()
	kb.AddScreen(Screen{Name: "login", Platform: "web"})
	kb.AddScreen(Screen{Name: "login", Platform: "web"}) // duplicate
	assert.Len(t, kb.Screens, 1) // deduped
}

func TestKnowledgeBase_AddEndpoint(t *testing.T) {
	kb := NewKnowledgeBase()
	kb.AddEndpoint(APIEndpoint{Method: "GET", Path: "/health"})
	kb.AddEndpoint(APIEndpoint{Method: "GET", Path: "/health"}) // duplicate
	assert.Len(t, kb.APIEndpoints, 1) // deduped
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd HelixQA && go test ./pkg/learning/ -v`
Expected: FAIL — package does not exist

- [ ] **Step 3: Write the implementation**

```go
// pkg/learning/knowledge.go
package learning

import (
	"fmt"
	"strings"
)

// Screen represents a discoverable UI screen.
type Screen struct {
	Name       string `json:"name"`
	Platform   string `json:"platform"`
	Route      string `json:"route,omitempty"`
	Component  string `json:"component,omitempty"`
	SourceFile string `json:"source_file,omitempty"`
}

// APIEndpoint represents a discovered API route.
type APIEndpoint struct {
	Method     string `json:"method"`
	Path       string `json:"path"`
	Handler    string `json:"handler,omitempty"`
	SourceFile string `json:"source_file,omitempty"`
}

// DocEntry represents a discovered documentation file.
type DocEntry struct {
	Path    string `json:"path"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

// ChangeEntry represents a recent git change.
type ChangeEntry struct {
	Hash    string `json:"hash"`
	Message string `json:"message"`
	Files   []string `json:"files"`
	Date    string `json:"date"`
}

// KnowledgeBase holds everything learned about the project.
type KnowledgeBase struct {
	ProjectName   string         `json:"project_name"`
	ProjectRoot   string         `json:"project_root"`
	Screens       []Screen       `json:"screens"`
	APIEndpoints  []APIEndpoint  `json:"api_endpoints"`
	Docs          []DocEntry     `json:"docs"`
	RecentChanges []ChangeEntry  `json:"recent_changes"`
	Components    []string       `json:"components"`
	Constraints   []string       `json:"constraints"`
	KnownIssues   []string       `json:"known_issues"`
}

// NewKnowledgeBase creates an empty KnowledgeBase.
func NewKnowledgeBase() *KnowledgeBase {
	return &KnowledgeBase{
		Screens:       make([]Screen, 0),
		APIEndpoints:  make([]APIEndpoint, 0),
		Docs:          make([]DocEntry, 0),
		RecentChanges: make([]ChangeEntry, 0),
		Components:    make([]string, 0),
		Constraints:   make([]string, 0),
		KnownIssues:   make([]string, 0),
	}
}

// AddScreen adds a screen if not already present (dedup by name+platform).
func (kb *KnowledgeBase) AddScreen(s Screen) {
	for _, existing := range kb.Screens {
		if existing.Name == s.Name && existing.Platform == s.Platform {
			return
		}
	}
	kb.Screens = append(kb.Screens, s)
}

// AddEndpoint adds an API endpoint if not already present (dedup by method+path).
func (kb *KnowledgeBase) AddEndpoint(e APIEndpoint) {
	for _, existing := range kb.APIEndpoints {
		if existing.Method == e.Method && existing.Path == e.Path {
			return
		}
	}
	kb.APIEndpoints = append(kb.APIEndpoints, e)
}

// Summary returns a human-readable summary for LLM context.
func (kb *KnowledgeBase) Summary() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Project: %s\n", kb.ProjectName))
	b.WriteString(fmt.Sprintf("Screens: %d screens\n", len(kb.Screens)))
	b.WriteString(fmt.Sprintf("API: %d API endpoints\n", len(kb.APIEndpoints)))
	b.WriteString(fmt.Sprintf("Docs: %d documentation files\n", len(kb.Docs)))
	b.WriteString(fmt.Sprintf("Recent changes: %d commits\n", len(kb.RecentChanges)))
	if len(kb.Components) > 0 {
		b.WriteString(fmt.Sprintf("Components: %s\n", strings.Join(kb.Components, ", ")))
	}
	if len(kb.Constraints) > 0 {
		b.WriteString("Constraints:\n")
		for _, c := range kb.Constraints {
			b.WriteString(fmt.Sprintf("  - %s\n", c))
		}
	}
	if len(kb.KnownIssues) > 0 {
		b.WriteString(fmt.Sprintf("Known issues: %d open\n", len(kb.KnownIssues)))
	}
	return b.String()
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd HelixQA && go test ./pkg/learning/ -v`
Expected: PASS (4 tests)

- [ ] **Step 5: Commit**

```bash
git add HelixQA/pkg/learning/
git commit -m "feat(helixqa): add KnowledgeBase types for learning engine"
```

---

## Task 2: ProjectReader — Read Docs & CLAUDE.md

**Files:**
- Create: `pkg/learning/reader.go`
- Test: `pkg/learning/reader_test.go`

- [ ] **Step 1: Write the failing test**

```go
// pkg/learning/reader_test.go
package learning

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestProject(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	// Create CLAUDE.md
	os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte("# CLAUDE.md\n\n## Overview\nTest project.\n\n## Constraints\n- Zero errors\n- No crashes"), 0644)

	// Create docs/
	os.MkdirAll(filepath.Join(dir, "docs"), 0755)
	os.WriteFile(filepath.Join(dir, "docs", "architecture.md"), []byte("# Architecture\nMicroservices."), 0644)
	os.WriteFile(filepath.Join(dir, "docs", "api.md"), []byte("# API Guide\nREST endpoints."), 0644)

	// Create nested CLAUDE.md
	os.MkdirAll(filepath.Join(dir, "submodule"), 0755)
	os.WriteFile(filepath.Join(dir, "submodule", "CLAUDE.md"), []byte("# Sub CLAUDE.md\nSub module docs."), 0644)

	return dir
}

func TestProjectReader_ReadDocs(t *testing.T) {
	dir := setupTestProject(t)
	reader := NewProjectReader(dir)

	docs, err := reader.ReadDocs()
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(docs), 2) // at least architecture.md + api.md
}

func TestProjectReader_ReadClaudeMDs(t *testing.T) {
	dir := setupTestProject(t)
	reader := NewProjectReader(dir)

	entries, err := reader.ReadClaudeMDs()
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(entries), 2) // root + submodule
}

func TestProjectReader_ExtractConstraints(t *testing.T) {
	dir := setupTestProject(t)
	reader := NewProjectReader(dir)

	entries, _ := reader.ReadClaudeMDs()
	constraints := reader.ExtractConstraints(entries)
	assert.GreaterOrEqual(t, len(constraints), 1)
}

func TestProjectReader_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	reader := NewProjectReader(dir)

	docs, err := reader.ReadDocs()
	require.NoError(t, err)
	assert.Len(t, docs, 0)
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd HelixQA && go test ./pkg/learning/ -v -run TestProjectReader`
Expected: FAIL — `NewProjectReader` undefined

- [ ] **Step 3: Write the implementation**

```go
// pkg/learning/reader.go
package learning

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ProjectReader reads documentation and CLAUDE.md files from a project.
type ProjectReader struct {
	root string
}

// NewProjectReader creates a reader for the given project root.
func NewProjectReader(root string) *ProjectReader {
	return &ProjectReader{root: root}
}

// ReadDocs discovers and reads all markdown files in docs/ directory.
func (r *ProjectReader) ReadDocs() ([]DocEntry, error) {
	docsDir := filepath.Join(r.root, "docs")
	if _, err := os.Stat(docsDir); os.IsNotExist(err) {
		return nil, nil
	}

	var docs []DocEntry
	err := filepath.Walk(docsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".md" && ext != ".markdown" {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return nil // skip unreadable files
		}
		relPath, _ := filepath.Rel(r.root, path)
		title := extractTitle(string(content))
		docs = append(docs, DocEntry{
			Path:    relPath,
			Title:   title,
			Content: truncateContent(string(content), 2000),
		})
		return nil
	})
	return docs, err
}

// ReadClaudeMDs finds and reads all CLAUDE.md files in the project.
func (r *ProjectReader) ReadClaudeMDs() ([]DocEntry, error) {
	var entries []DocEntry
	err := filepath.Walk(r.root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		name := filepath.Base(path)
		if name != "CLAUDE.md" && name != "AGENTS.md" {
			return nil
		}
		// Skip node_modules, .git, vendor
		rel, _ := filepath.Rel(r.root, path)
		if strings.Contains(rel, "node_modules") || strings.Contains(rel, ".git"+string(filepath.Separator)) || strings.Contains(rel, "vendor") {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		entries = append(entries, DocEntry{
			Path:    rel,
			Title:   name + " (" + filepath.Dir(rel) + ")",
			Content: truncateContent(string(content), 4000),
		})
		return nil
	})
	return entries, err
}

// ExtractConstraints pulls constraint lines from CLAUDE.md entries.
func (r *ProjectReader) ExtractConstraints(entries []DocEntry) []string {
	var constraints []string
	for _, entry := range entries {
		lines := strings.Split(entry.Content, "\n")
		inConstraints := false
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if strings.Contains(strings.ToLower(trimmed), "constraint") || strings.Contains(strings.ToLower(trimmed), "critical") {
				inConstraints = true
				continue
			}
			if inConstraints && strings.HasPrefix(trimmed, "- ") {
				constraints = append(constraints, strings.TrimPrefix(trimmed, "- "))
			}
			if inConstraints && trimmed == "" {
				inConstraints = false
			}
		}
	}
	return constraints
}

func extractTitle(content string) string {
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "# ") {
			return strings.TrimPrefix(trimmed, "# ")
		}
	}
	return ""
}

func truncateContent(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "\n...(truncated)"
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd HelixQA && go test ./pkg/learning/ -v -run TestProjectReader`
Expected: PASS (4 tests)

- [ ] **Step 5: Commit**

```bash
git add HelixQA/pkg/learning/reader.go HelixQA/pkg/learning/reader_test.go
git commit -m "feat(helixqa): add ProjectReader for docs and CLAUDE.md ingestion"
```

---

## Task 3: GitAnalyzer — Recent Changes & Hotspots

**Files:**
- Create: `pkg/learning/git.go`
- Test: `pkg/learning/git_test.go`

- [ ] **Step 1: Write the failing test**

```go
// pkg/learning/git_test.go
package learning

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupGitRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	// Init git repo
	run := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=Test", "GIT_AUTHOR_EMAIL=test@test.com",
			"GIT_COMMITTER_NAME=Test", "GIT_COMMITTER_EMAIL=test@test.com",
		)
		cmd.Run()
	}
	run("init")
	run("config", "user.email", "test@test.com")
	run("config", "user.name", "Test")

	os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main"), 0644)
	run("add", ".")
	run("commit", "-m", "initial commit")

	os.WriteFile(filepath.Join(dir, "handler.go"), []byte("package main\nfunc handler(){}"), 0644)
	run("add", ".")
	run("commit", "-m", "feat: add handler")

	return dir
}

func TestGitAnalyzer_RecentCommits(t *testing.T) {
	dir := setupGitRepo(t)
	ga := NewGitAnalyzer(dir)

	commits, err := ga.RecentCommits(10)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(commits), 2)
	assert.Contains(t, commits[0].Message, "add handler")
}

func TestGitAnalyzer_ChangedFiles(t *testing.T) {
	dir := setupGitRepo(t)
	ga := NewGitAnalyzer(dir)

	commits, _ := ga.RecentCommits(10)
	assert.GreaterOrEqual(t, len(commits[0].Files), 1)
}

func TestGitAnalyzer_HotFiles(t *testing.T) {
	dir := setupGitRepo(t)
	ga := NewGitAnalyzer(dir)

	hotFiles := ga.HotFiles(10)
	assert.GreaterOrEqual(t, len(hotFiles), 1)
}

func TestGitAnalyzer_NotARepo(t *testing.T) {
	dir := t.TempDir()
	ga := NewGitAnalyzer(dir)

	commits, err := ga.RecentCommits(10)
	assert.Nil(t, commits)
	assert.Error(t, err)
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd HelixQA && go test ./pkg/learning/ -v -run TestGitAnalyzer`
Expected: FAIL — `NewGitAnalyzer` undefined

- [ ] **Step 3: Write the implementation**

```go
// pkg/learning/git.go
package learning

import (
	"fmt"
	"os/exec"
	"strings"
)

// GitAnalyzer extracts recent changes and hotspots from git history.
type GitAnalyzer struct {
	root string
}

// NewGitAnalyzer creates an analyzer for the given repo root.
func NewGitAnalyzer(root string) *GitAnalyzer {
	return &GitAnalyzer{root: root}
}

// RecentCommits returns the most recent commits with changed files.
func (g *GitAnalyzer) RecentCommits(limit int) ([]ChangeEntry, error) {
	out, err := g.git("log", fmt.Sprintf("--max-count=%d", limit),
		"--pretty=format:%H|%s|%aI", "--name-only")
	if err != nil {
		return nil, fmt.Errorf("git log: %w", err)
	}

	var commits []ChangeEntry
	blocks := strings.Split(strings.TrimSpace(out), "\n\n")
	for _, block := range blocks {
		lines := strings.Split(strings.TrimSpace(block), "\n")
		if len(lines) == 0 || lines[0] == "" {
			continue
		}
		parts := strings.SplitN(lines[0], "|", 3)
		if len(parts) < 3 {
			continue
		}
		entry := ChangeEntry{
			Hash:    parts[0][:8],
			Message: parts[1],
			Date:    parts[2],
			Files:   make([]string, 0),
		}
		for _, f := range lines[1:] {
			f = strings.TrimSpace(f)
			if f != "" {
				entry.Files = append(entry.Files, f)
			}
		}
		commits = append(commits, entry)
	}
	return commits, nil
}

// HotFiles returns the most frequently changed files.
func (g *GitAnalyzer) HotFiles(limit int) []string {
	out, err := g.git("log", "--all", "--name-only", "--pretty=format:", "--max-count=100")
	if err != nil {
		return nil
	}

	counts := make(map[string]int)
	for _, line := range strings.Split(out, "\n") {
		f := strings.TrimSpace(line)
		if f != "" {
			counts[f]++
		}
	}

	// Simple sort by count (top N)
	type fileCount struct {
		file  string
		count int
	}
	var sorted []fileCount
	for f, c := range counts {
		sorted = append(sorted, fileCount{f, c})
	}
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j].count > sorted[i].count {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	var result []string
	for i, fc := range sorted {
		if i >= limit {
			break
		}
		result = append(result, fc.file)
	}
	return result
}

func (g *GitAnalyzer) git(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = g.root
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd HelixQA && go test ./pkg/learning/ -v -run TestGitAnalyzer`
Expected: PASS (4 tests)

- [ ] **Step 5: Commit**

```bash
git add HelixQA/pkg/learning/git.go HelixQA/pkg/learning/git_test.go
git commit -m "feat(helixqa): add GitAnalyzer for change history and hotspot detection"
```

---

## Task 4: CodebaseMapper — Extract Routes, Screens, Endpoints

**Files:**
- Create: `pkg/learning/codebase.go`
- Test: `pkg/learning/codebase_test.go`

- [ ] **Step 1: Write the failing test**

```go
// pkg/learning/codebase_test.go
package learning

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupCodebaseProject(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	// Simulate Go main.go with Gin routes
	os.MkdirAll(filepath.Join(dir, "catalog-api"), 0755)
	os.WriteFile(filepath.Join(dir, "catalog-api", "main.go"), []byte(`package main
// Routes
router.GET("/api/v1/health", handlers.Health)
router.POST("/api/v1/auth/login", handlers.Login)
router.GET("/api/v1/catalog", handlers.ListCatalog)
router.GET("/api/v1/entities", handlers.ListEntities)
`), 0644)

	// Simulate React routes
	os.MkdirAll(filepath.Join(dir, "catalog-web", "src"), 0755)
	os.WriteFile(filepath.Join(dir, "catalog-web", "src", "App.tsx"), []byte(`
<Route path="/login" element={<Login />} />
<Route path="/dashboard" element={<Dashboard />} />
<Route path="/media" element={<MediaBrowser />} />
<Route path="/browse" element={<EntityBrowser />} />
<Route path="/collections" element={<Collections />} />
`), 0644)

	// Simulate Android nav
	os.MkdirAll(filepath.Join(dir, "catalogizer-android", "app", "src"), 0755)
	os.WriteFile(filepath.Join(dir, "catalogizer-android", "app", "src", "Navigation.kt"), []byte(`
composable("home") { HomeScreen() }
composable("login") { LoginScreen() }
composable("media/{id}") { MediaDetailScreen() }
`), 0644)

	return dir
}

func TestCodebaseMapper_ExtractAPIEndpoints(t *testing.T) {
	dir := setupCodebaseProject(t)
	mapper := NewCodebaseMapper(dir)

	endpoints, err := mapper.ExtractAPIEndpoints()
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(endpoints), 3)

	var methods []string
	for _, ep := range endpoints {
		methods = append(methods, ep.Method+" "+ep.Path)
	}
	assert.Contains(t, methods, "GET /api/v1/health")
}

func TestCodebaseMapper_ExtractWebScreens(t *testing.T) {
	dir := setupCodebaseProject(t)
	mapper := NewCodebaseMapper(dir)

	screens, err := mapper.ExtractWebScreens()
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(screens), 4)
}

func TestCodebaseMapper_ExtractAndroidScreens(t *testing.T) {
	dir := setupCodebaseProject(t)
	mapper := NewCodebaseMapper(dir)

	screens, err := mapper.ExtractAndroidScreens()
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(screens), 2)
}

func TestCodebaseMapper_EmptyProject(t *testing.T) {
	dir := t.TempDir()
	mapper := NewCodebaseMapper(dir)

	ep, _ := mapper.ExtractAPIEndpoints()
	assert.Len(t, ep, 0)
	ws, _ := mapper.ExtractWebScreens()
	assert.Len(t, ws, 0)
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd HelixQA && go test ./pkg/learning/ -v -run TestCodebaseMapper`
Expected: FAIL — `NewCodebaseMapper` undefined

- [ ] **Step 3: Write the implementation**

```go
// pkg/learning/codebase.go
package learning

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// CodebaseMapper extracts routes, screens, and components from source code.
type CodebaseMapper struct {
	root string
}

// NewCodebaseMapper creates a mapper for the given project root.
func NewCodebaseMapper(root string) *CodebaseMapper {
	return &CodebaseMapper{root: root}
}

var ginRouteRegex = regexp.MustCompile(`(?i)router\.(GET|POST|PUT|DELETE|PATCH)\s*\(\s*"([^"]+)"`)
var reactRouteRegex = regexp.MustCompile(`<Route\s+path="([^"]+)"\s+element=\{<(\w+)`)
var composeNavRegex = regexp.MustCompile(`composable\(\s*"([^"]+)"\s*\)\s*\{[^}]*?(\w+Screen)\s*\(`)

// ExtractAPIEndpoints scans Go files for Gin route definitions.
func (m *CodebaseMapper) ExtractAPIEndpoints() ([]APIEndpoint, error) {
	var endpoints []APIEndpoint
	apiDir := filepath.Join(m.root, "catalog-api")
	if _, err := os.Stat(apiDir); os.IsNotExist(err) {
		return endpoints, nil
	}

	filepath.Walk(apiDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".go") {
			return err
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		rel, _ := filepath.Rel(m.root, path)
		matches := ginRouteRegex.FindAllStringSubmatch(string(content), -1)
		for _, match := range matches {
			if len(match) >= 3 {
				endpoints = append(endpoints, APIEndpoint{
					Method:     strings.ToUpper(match[1]),
					Path:       match[2],
					SourceFile: rel,
				})
			}
		}
		return nil
	})
	return endpoints, nil
}

// ExtractWebScreens scans React/TSX files for Route definitions.
func (m *CodebaseMapper) ExtractWebScreens() ([]Screen, error) {
	var screens []Screen
	webDir := filepath.Join(m.root, "catalog-web")
	if _, err := os.Stat(webDir); os.IsNotExist(err) {
		return screens, nil
	}

	filepath.Walk(webDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		ext := filepath.Ext(path)
		if ext != ".tsx" && ext != ".jsx" && ext != ".ts" {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		rel, _ := filepath.Rel(m.root, path)
		matches := reactRouteRegex.FindAllStringSubmatch(string(content), -1)
		for _, match := range matches {
			if len(match) >= 3 {
				screens = append(screens, Screen{
					Name:       match[2],
					Platform:   "web",
					Route:      match[1],
					Component:  match[2],
					SourceFile: rel,
				})
			}
		}
		return nil
	})
	return screens, nil
}

// ExtractAndroidScreens scans Kotlin files for Compose navigation.
func (m *CodebaseMapper) ExtractAndroidScreens() ([]Screen, error) {
	var screens []Screen
	for _, dirName := range []string{"catalogizer-android", "catalogizer-androidtv"} {
		platform := "android"
		if strings.Contains(dirName, "tv") {
			platform = "androidtv"
		}
		appDir := filepath.Join(m.root, dirName)
		if _, err := os.Stat(appDir); os.IsNotExist(err) {
			continue
		}

		filepath.Walk(appDir, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() || !strings.HasSuffix(path, ".kt") {
				return err
			}
			content, err := os.ReadFile(path)
			if err != nil {
				return nil
			}
			rel, _ := filepath.Rel(m.root, path)
			matches := composeNavRegex.FindAllStringSubmatch(string(content), -1)
			for _, match := range matches {
				if len(match) >= 3 {
					screens = append(screens, Screen{
						Name:       match[2],
						Platform:   platform,
						Route:      match[1],
						Component:  match[2],
						SourceFile: rel,
					})
				}
			}
			return nil
		})
	}
	return screens, nil
}

// DiscoverComponents returns the top-level project component names.
func (m *CodebaseMapper) DiscoverComponents() []string {
	known := []string{
		"catalog-api", "catalog-web", "catalogizer-android",
		"catalogizer-androidtv", "catalogizer-desktop", "installer-wizard",
	}
	var found []string
	for _, name := range known {
		if _, err := os.Stat(filepath.Join(m.root, name)); err == nil {
			found = append(found, name)
		}
	}
	return found
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd HelixQA && go test ./pkg/learning/ -v -run TestCodebaseMapper`
Expected: PASS (4 tests)

- [ ] **Step 5: Commit**

```bash
git add HelixQA/pkg/learning/codebase.go HelixQA/pkg/learning/codebase_test.go
git commit -m "feat(helixqa): add CodebaseMapper for route, screen, endpoint extraction"
```

---

## Task 5: KnowledgeBase Builder — Compose All Readers

**Files:**
- Modify: `pkg/learning/knowledge.go`
- Test: `pkg/learning/knowledge_test.go` (add builder tests)

- [ ] **Step 1: Add builder test**

Append to `knowledge_test.go`:

```go
func TestBuildKnowledgeBase(t *testing.T) {
	dir := setupTestProject(t)
	// Add a fake Go file with routes
	os.MkdirAll(filepath.Join(dir, "catalog-api"), 0755)
	os.WriteFile(filepath.Join(dir, "catalog-api", "main.go"), []byte(`package main
router.GET("/api/v1/health", h.Health)
`), 0644)

	kb, err := BuildKnowledgeBase(dir, nil)
	require.NoError(t, err)
	assert.Equal(t, filepath.Base(dir), kb.ProjectName)
	assert.GreaterOrEqual(t, len(kb.Docs), 1)
	assert.GreaterOrEqual(t, len(kb.APIEndpoints), 1)
}

func TestBuildKnowledgeBase_WithMemory(t *testing.T) {
	dir := setupTestProject(t)

	store, err := memory.NewStore(filepath.Join(t.TempDir(), "mem.db"))
	require.NoError(t, err)
	defer store.Close()

	store.SetKnowledge("total_screens", "15", "prior_scan")

	kb, err := BuildKnowledgeBase(dir, store)
	require.NoError(t, err)
	assert.NotEmpty(t, kb.ProjectName)
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd HelixQA && go test ./pkg/learning/ -v -run TestBuildKnowledge`
Expected: FAIL — `BuildKnowledgeBase` undefined

- [ ] **Step 3: Add BuildKnowledgeBase to knowledge.go**

Append to `knowledge.go`:

```go
import (
	"path/filepath"

	"digital.vasic.helixqa/pkg/memory"
)

// BuildKnowledgeBase composes all readers to build a complete knowledge base.
// store can be nil if no memory DB is available.
func BuildKnowledgeBase(projectRoot string, store *memory.Store) (*KnowledgeBase, error) {
	kb := NewKnowledgeBase()
	kb.ProjectRoot = projectRoot
	kb.ProjectName = filepath.Base(projectRoot)

	// 1. Read documentation
	reader := NewProjectReader(projectRoot)
	docs, err := reader.ReadDocs()
	if err == nil {
		kb.Docs = append(kb.Docs, docs...)
	}

	// 2. Read CLAUDE.md files
	claudeMDs, err := reader.ReadClaudeMDs()
	if err == nil {
		kb.Docs = append(kb.Docs, claudeMDs...)
		kb.Constraints = reader.ExtractConstraints(claudeMDs)
	}

	// 3. Extract codebase structure
	mapper := NewCodebaseMapper(projectRoot)
	endpoints, err := mapper.ExtractAPIEndpoints()
	if err == nil {
		for _, ep := range endpoints {
			kb.AddEndpoint(ep)
		}
	}
	webScreens, err := mapper.ExtractWebScreens()
	if err == nil {
		for _, s := range webScreens {
			kb.AddScreen(s)
		}
	}
	androidScreens, err := mapper.ExtractAndroidScreens()
	if err == nil {
		for _, s := range androidScreens {
			kb.AddScreen(s)
		}
	}
	kb.Components = mapper.DiscoverComponents()

	// 4. Git history
	ga := NewGitAnalyzer(projectRoot)
	commits, err := ga.RecentCommits(20)
	if err == nil {
		kb.RecentChanges = commits
	}

	// 5. Prior session knowledge (if memory store available)
	if store != nil {
		openFindings, err := store.ListFindingsByStatus("open")
		if err == nil {
			for _, f := range openFindings {
				kb.KnownIssues = append(kb.KnownIssues,
					fmt.Sprintf("[%s] %s (%s/%s)", f.ID, f.Title, f.Platform, f.Screen))
			}
		}
	}

	return kb, nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd HelixQA && go test ./pkg/learning/ -v`
Expected: PASS (all learning tests)

- [ ] **Step 5: Commit**

```bash
git add HelixQA/pkg/learning/
git commit -m "feat(helixqa): add BuildKnowledgeBase composer integrating all readers"
```

---

## Task 6: Planning Types

**Files:**
- Create: `pkg/planning/types.go`

- [ ] **Step 1: Write the types file**

```go
// pkg/planning/types.go
package planning

import (
	"digital.vasic.helixqa/pkg/learning"
)

// PlannedTest represents a single test case in the plan.
type PlannedTest struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Category    string   `json:"category"`    // functional, visual, edge_case, performance
	Priority    int      `json:"priority"`    // 1=critical, 2=high, 3=medium, 4=low
	Platforms   []string `json:"platforms"`
	Screen      string   `json:"screen"`
	Steps       []string `json:"steps"`
	Expected    string   `json:"expected"`
	IsExisting  bool     `json:"is_existing"` // true if from existing bank
	IsNew       bool     `json:"is_new"`      // true if generated this session
	BankSource  string   `json:"bank_source"` // file path if from existing bank
}

// TestPlan is the complete plan for a QA session.
type TestPlan struct {
	SessionID    string        `json:"session_id"`
	Generated    string        `json:"generated"`
	TotalTests   int           `json:"total_tests"`
	ExistingTests int          `json:"existing_tests"`
	NewTests     int           `json:"new_tests"`
	Platforms    []string      `json:"platforms"`
	Tests        []PlannedTest `json:"tests"`
}

// PlanStats returns summary statistics.
func (tp *TestPlan) PlanStats() PlanStats {
	byCategory := make(map[string]int)
	byPlatform := make(map[string]int)
	for _, t := range tp.Tests {
		byCategory[t.Category]++
		for _, p := range t.Platforms {
			byPlatform[p]++
		}
	}
	return PlanStats{
		Total:      tp.TotalTests,
		Existing:   tp.ExistingTests,
		New:        tp.NewTests,
		ByCategory: byCategory,
		ByPlatform: byPlatform,
	}
}

// PlanStats holds summary statistics about a test plan.
type PlanStats struct {
	Total      int            `json:"total"`
	Existing   int            `json:"existing"`
	New        int            `json:"new"`
	ByCategory map[string]int `json:"by_category"`
	ByPlatform map[string]int `json:"by_platform"`
}

// ForPlatform returns tests targeting a specific platform.
func (tp *TestPlan) ForPlatform(platform string) []PlannedTest {
	var result []PlannedTest
	for _, t := range tp.Tests {
		for _, p := range t.Platforms {
			if p == platform {
				result = append(result, t)
				break
			}
		}
	}
	return result
}

// Ensure learning import is used
var _ = learning.Screen{}
```

- [ ] **Step 2: Verify it compiles**

Run: `cd HelixQA && go build ./pkg/planning/`
Expected: Success (no errors)

- [ ] **Step 3: Commit**

```bash
git add HelixQA/pkg/planning/types.go
git commit -m "feat(helixqa): add TestPlan and PlannedTest types for planning engine"
```

---

## Task 7: PriorityRanker — Order Tests by Criticality

**Files:**
- Create: `pkg/planning/ranker.go`
- Test: `pkg/planning/ranker_test.go`

- [ ] **Step 1: Write the failing test**

```go
// pkg/planning/ranker_test.go
package planning

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPriorityRanker_SortByPriority(t *testing.T) {
	tests := []PlannedTest{
		{ID: "low", Name: "Low priority", Priority: 4},
		{ID: "critical", Name: "Critical path", Priority: 1},
		{ID: "medium", Name: "Medium", Priority: 3},
		{ID: "high", Name: "High", Priority: 2},
	}

	ranker := NewPriorityRanker(nil)
	sorted := ranker.Rank(tests)

	assert.Equal(t, "critical", sorted[0].ID)
	assert.Equal(t, "high", sorted[1].ID)
	assert.Equal(t, "medium", sorted[2].ID)
	assert.Equal(t, "low", sorted[3].ID)
}

func TestPriorityRanker_ExistingFirst(t *testing.T) {
	tests := []PlannedTest{
		{ID: "new1", Priority: 2, IsNew: true},
		{ID: "existing1", Priority: 2, IsExisting: true},
	}

	ranker := NewPriorityRanker(nil)
	sorted := ranker.Rank(tests)

	// Same priority: existing tests run before new ones
	assert.Equal(t, "existing1", sorted[0].ID)
}

func TestPriorityRanker_WithFailHistory(t *testing.T) {
	failedIDs := map[string]bool{"flaky1": true}

	tests := []PlannedTest{
		{ID: "stable", Priority: 3},
		{ID: "flaky1", Priority: 3},
	}

	ranker := NewPriorityRanker(failedIDs)
	sorted := ranker.Rank(tests)

	// Previously failed tests get boosted
	assert.Equal(t, "flaky1", sorted[0].ID)
}

func TestPriorityRanker_EmptyList(t *testing.T) {
	ranker := NewPriorityRanker(nil)
	sorted := ranker.Rank(nil)
	assert.Len(t, sorted, 0)
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd HelixQA && go test ./pkg/planning/ -v -run TestPriorityRanker`
Expected: FAIL — `NewPriorityRanker` undefined

- [ ] **Step 3: Write the implementation**

```go
// pkg/planning/ranker.go
package planning

import "sort"

// PriorityRanker orders tests by criticality, failure history, and type.
type PriorityRanker struct {
	priorFailures map[string]bool // test IDs that failed in prior sessions
}

// NewPriorityRanker creates a ranker. priorFailures can be nil.
func NewPriorityRanker(priorFailures map[string]bool) *PriorityRanker {
	if priorFailures == nil {
		priorFailures = make(map[string]bool)
	}
	return &PriorityRanker{priorFailures: priorFailures}
}

// Rank returns a copy of tests sorted by priority.
// Order: priority (1=critical first), then prior failures boosted,
// then existing before new, then alphabetical by ID.
func (r *PriorityRanker) Rank(tests []PlannedTest) []PlannedTest {
	if len(tests) == 0 {
		return tests
	}

	sorted := make([]PlannedTest, len(tests))
	copy(sorted, tests)

	sort.SliceStable(sorted, func(i, j int) bool {
		a, b := sorted[i], sorted[j]

		// 1. Priority (lower number = higher priority)
		if a.Priority != b.Priority {
			return a.Priority < b.Priority
		}

		// 2. Previously failed tests get boosted
		aFailed := r.priorFailures[a.ID]
		bFailed := r.priorFailures[b.ID]
		if aFailed != bFailed {
			return aFailed
		}

		// 3. Existing tests before new
		if a.IsExisting != b.IsExisting {
			return a.IsExisting
		}

		// 4. Alphabetical
		return a.ID < b.ID
	})

	return sorted
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd HelixQA && go test ./pkg/planning/ -v -run TestPriorityRanker`
Expected: PASS (4 tests)

- [ ] **Step 5: Commit**

```bash
git add HelixQA/pkg/planning/ranker.go HelixQA/pkg/planning/ranker_test.go
git commit -m "feat(helixqa): add PriorityRanker for test ordering"
```

---

## Task 8: BankReconciler — Diff Plan Against Existing Banks

**Files:**
- Create: `pkg/planning/reconciler.go`
- Test: `pkg/planning/reconciler_test.go`

- [ ] **Step 1: Write the failing test**

```go
// pkg/planning/reconciler_test.go
package planning

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBankReconciler_LoadExisting(t *testing.T) {
	dir := t.TempDir()
	bankFile := filepath.Join(dir, "test-bank.yaml")
	os.WriteFile(bankFile, []byte(`version: "1.0"
name: "Test Bank"
test_cases:
  - id: TC-001
    name: Login flow
    category: functional
    priority: critical
    platforms: [web, android]
    steps:
      - name: Open login
        action: Navigate to /login
        expected: Login form visible
`), 0644)

	reconciler := NewBankReconciler()
	err := reconciler.LoadBankDir(dir)
	require.NoError(t, err)
	assert.Equal(t, 1, reconciler.ExistingCount())
}

func TestBankReconciler_Reconcile(t *testing.T) {
	reconciler := NewBankReconciler()
	reconciler.AddExisting("TC-001", "Login flow", "test-bank.yaml")

	generated := []PlannedTest{
		{ID: "GEN-001", Name: "Login flow", Category: "functional"},       // matches existing
		{ID: "GEN-002", Name: "Search feature", Category: "functional"},    // new
	}

	result := reconciler.Reconcile(generated)
	assert.Equal(t, 2, len(result))

	// Matched test should be marked as existing
	var loginTest *PlannedTest
	for i := range result {
		if result[i].Name == "Login flow" {
			loginTest = &result[i]
			break
		}
	}
	require.NotNil(t, loginTest)
	assert.True(t, loginTest.IsExisting)
	assert.Equal(t, "test-bank.yaml", loginTest.BankSource)

	// New test marked correctly
	var searchTest *PlannedTest
	for i := range result {
		if result[i].Name == "Search feature" {
			searchTest = &result[i]
			break
		}
	}
	require.NotNil(t, searchTest)
	assert.True(t, searchTest.IsNew)
}

func TestBankReconciler_NewTests(t *testing.T) {
	reconciler := NewBankReconciler()
	generated := []PlannedTest{
		{ID: "G-1", Name: "New test", Category: "edge_case"},
	}
	result := reconciler.Reconcile(generated)
	assert.Len(t, result, 1)
	assert.True(t, result[0].IsNew)
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd HelixQA && go test ./pkg/planning/ -v -run TestBankReconciler`
Expected: FAIL — `NewBankReconciler` undefined

- [ ] **Step 3: Write the implementation**

```go
// pkg/planning/reconciler.go
package planning

import (
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// BankReconciler compares generated tests against existing test banks.
type BankReconciler struct {
	existing map[string]existingEntry // keyed by lowercase name
}

type existingEntry struct {
	id     string
	name   string
	source string
}

type bankFile struct {
	Version   string         `yaml:"version"`
	Name      string         `yaml:"name"`
	TestCases []bankTestCase `yaml:"test_cases"`
}

type bankTestCase struct {
	ID       string   `yaml:"id"`
	Name     string   `yaml:"name"`
	Category string   `yaml:"category"`
	Priority string   `yaml:"priority"`
	Platforms []string `yaml:"platforms"`
}

// NewBankReconciler creates an empty reconciler.
func NewBankReconciler() *BankReconciler {
	return &BankReconciler{
		existing: make(map[string]existingEntry),
	}
}

// LoadBankDir loads all YAML/JSON test bank files from a directory.
func (r *BankReconciler) LoadBankDir(dir string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".yaml" && ext != ".yml" {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		var bf bankFile
		if err := yaml.Unmarshal(data, &bf); err != nil {
			return nil
		}
		rel, _ := filepath.Rel(filepath.Dir(dir), path)
		for _, tc := range bf.TestCases {
			r.AddExisting(tc.ID, tc.Name, rel)
		}
		return nil
	})
}

// AddExisting registers an existing test case by name.
func (r *BankReconciler) AddExisting(id, name, source string) {
	key := strings.ToLower(strings.TrimSpace(name))
	r.existing[key] = existingEntry{id: id, name: name, source: source}
}

// ExistingCount returns the number of loaded existing tests.
func (r *BankReconciler) ExistingCount() int {
	return len(r.existing)
}

// Reconcile cross-references generated tests against existing bank.
// Matching tests get IsExisting=true and BankSource set.
// Non-matching tests get IsNew=true.
func (r *BankReconciler) Reconcile(generated []PlannedTest) []PlannedTest {
	result := make([]PlannedTest, len(generated))
	copy(result, generated)

	for i := range result {
		key := strings.ToLower(strings.TrimSpace(result[i].Name))
		if entry, found := r.existing[key]; found {
			result[i].IsExisting = true
			result[i].IsNew = false
			result[i].BankSource = entry.source
			if result[i].ID == "" || strings.HasPrefix(result[i].ID, "GEN-") {
				result[i].ID = entry.id
			}
		} else {
			result[i].IsNew = true
			result[i].IsExisting = false
		}
	}
	return result
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd HelixQA && go test ./pkg/planning/ -v -run TestBankReconciler`
Expected: PASS (3 tests)

- [ ] **Step 5: Commit**

```bash
git add HelixQA/pkg/planning/reconciler.go HelixQA/pkg/planning/reconciler_test.go
git commit -m "feat(helixqa): add BankReconciler for test plan reconciliation"
```

---

## Task 9: TestPlanGenerator — LLM-Driven Test Plan Creation

**Files:**
- Create: `pkg/planning/planner.go`
- Test: `pkg/planning/planner_test.go`

- [ ] **Step 1: Write the failing test**

```go
// pkg/planning/planner_test.go
package planning

import (
	"context"
	"encoding/json"
	"testing"

	"digital.vasic.helixqa/pkg/learning"
	"digital.vasic.helixqa/pkg/llm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockLLM struct {
	response string
}

func (m *mockLLM) Chat(_ context.Context, _ []llm.Message) (*llm.Response, error) {
	return &llm.Response{Content: m.response}, nil
}
func (m *mockLLM) Vision(_ context.Context, _ []byte, _ string) (*llm.Response, error) {
	return &llm.Response{Content: m.response}, nil
}
func (m *mockLLM) Name() string        { return "mock" }
func (m *mockLLM) SupportsVision() bool { return false }

func TestTestPlanGenerator_Generate(t *testing.T) {
	// Mock LLM returns a JSON array of test cases
	mockTests := []PlannedTest{
		{ID: "GEN-001", Name: "Login flow", Category: "functional", Priority: 1, Platforms: []string{"web"}, Screen: "login", Steps: []string{"Navigate to /login", "Enter credentials", "Click sign in"}, Expected: "Dashboard visible"},
		{ID: "GEN-002", Name: "Search media", Category: "functional", Priority: 2, Platforms: []string{"web", "android"}, Screen: "media-browser", Steps: []string{"Open search", "Type query"}, Expected: "Results displayed"},
	}
	mockJSON, _ := json.Marshal(mockTests)

	provider := &mockLLM{response: string(mockJSON)}
	kb := learning.NewKnowledgeBase()
	kb.ProjectName = "Catalogizer"
	kb.AddScreen(learning.Screen{Name: "login", Platform: "web", Route: "/login"})
	kb.AddEndpoint(learning.APIEndpoint{Method: "GET", Path: "/health"})

	gen := NewTestPlanGenerator(provider)
	plan, err := gen.Generate(context.Background(), kb, []string{"web", "android"})
	require.NoError(t, err)
	assert.GreaterOrEqual(t, plan.TotalTests, 2)
	assert.Equal(t, "Catalogizer", plan.Tests[0].Screen != "" || true) // has content
}

func TestTestPlanGenerator_EmptyKnowledge(t *testing.T) {
	provider := &mockLLM{response: "[]"}
	kb := learning.NewKnowledgeBase()

	gen := NewTestPlanGenerator(provider)
	plan, err := gen.Generate(context.Background(), kb, []string{"web"})
	require.NoError(t, err)
	assert.Equal(t, 0, plan.TotalTests)
}

func TestTestPlanGenerator_MalformedLLMResponse(t *testing.T) {
	provider := &mockLLM{response: "this is not json"}
	kb := learning.NewKnowledgeBase()
	kb.ProjectName = "Test"

	gen := NewTestPlanGenerator(provider)
	plan, err := gen.Generate(context.Background(), kb, []string{"web"})
	// Should handle gracefully — return empty plan, not crash
	require.NoError(t, err)
	assert.Equal(t, 0, plan.TotalTests)
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd HelixQA && go test ./pkg/planning/ -v -run TestTestPlanGenerator`
Expected: FAIL — `NewTestPlanGenerator` undefined

- [ ] **Step 3: Write the implementation**

```go
// pkg/planning/planner.go
package planning

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"digital.vasic.helixqa/pkg/learning"
	"digital.vasic.helixqa/pkg/llm"
)

// TestPlanGenerator uses an LLM to create comprehensive test plans.
type TestPlanGenerator struct {
	provider llm.Provider
}

// NewTestPlanGenerator creates a planner backed by the given LLM.
func NewTestPlanGenerator(provider llm.Provider) *TestPlanGenerator {
	return &TestPlanGenerator{provider: provider}
}

// Generate builds a test plan from project knowledge.
func (g *TestPlanGenerator) Generate(ctx context.Context, kb *learning.KnowledgeBase, platforms []string) (*TestPlan, error) {
	prompt := g.buildPrompt(kb, platforms)

	resp, err := g.provider.Chat(ctx, []llm.Message{
		{Role: llm.RoleSystem, Content: systemPrompt},
		{Role: llm.RoleUser, Content: prompt},
	})
	if err != nil {
		return nil, fmt.Errorf("planning: LLM call failed: %w", err)
	}

	tests := g.parseTests(resp.Content)

	plan := &TestPlan{
		SessionID: fmt.Sprintf("plan-%d", time.Now().Unix()),
		Generated: time.Now().Format(time.RFC3339),
		Tests:     tests,
		Platforms: platforms,
	}

	// Count stats
	for _, t := range plan.Tests {
		plan.TotalTests++
		if t.IsNew {
			plan.NewTests++
		}
		if t.IsExisting {
			plan.ExistingTests++
		}
	}

	return plan, nil
}

func (g *TestPlanGenerator) buildPrompt(kb *learning.KnowledgeBase, platforms []string) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Generate a comprehensive QA test plan for the %s project.\n\n", kb.ProjectName))
	b.WriteString(fmt.Sprintf("Target platforms: %s\n\n", strings.Join(platforms, ", ")))
	b.WriteString("Project knowledge:\n")
	b.WriteString(kb.Summary())
	b.WriteString("\nScreens:\n")
	for _, s := range kb.Screens {
		b.WriteString(fmt.Sprintf("  - %s (%s) route=%s\n", s.Name, s.Platform, s.Route))
	}
	b.WriteString("\nAPI Endpoints:\n")
	for _, ep := range kb.APIEndpoints {
		if len(kb.APIEndpoints) > 20 {
			break // limit to avoid token overflow
		}
		b.WriteString(fmt.Sprintf("  - %s %s\n", ep.Method, ep.Path))
	}
	if len(kb.KnownIssues) > 0 {
		b.WriteString("\nKnown open issues (re-verify these):\n")
		for _, issue := range kb.KnownIssues {
			b.WriteString(fmt.Sprintf("  - %s\n", issue))
		}
	}
	b.WriteString("\nRespond with ONLY a JSON array of test objects. Each object must have: id, name, description, category (functional/visual/edge_case/performance), priority (1-4), platforms (array), screen, steps (array of strings), expected (string).\n")
	return b.String()
}

func (g *TestPlanGenerator) parseTests(content string) []PlannedTest {
	// Try to extract JSON from response (may have markdown wrapping)
	cleaned := content
	if idx := strings.Index(cleaned, "["); idx >= 0 {
		cleaned = cleaned[idx:]
	}
	if idx := strings.LastIndex(cleaned, "]"); idx >= 0 {
		cleaned = cleaned[:idx+1]
	}

	var tests []PlannedTest
	if err := json.Unmarshal([]byte(cleaned), &tests); err != nil {
		return nil // graceful degradation on parse failure
	}

	// Mark all as new by default
	for i := range tests {
		tests[i].IsNew = true
	}
	return tests
}

const systemPrompt = `You are an expert QA engineer. Generate comprehensive test cases covering:
1. All critical user flows (auth, browse, detail, favorites, search, playback)
2. All screens on each platform
3. Edge cases (empty states, errors, network failure, rotation, background/restore)
4. Visual/UI checks (layout, alignment, brand compliance)
5. Performance (response time, memory, animation smoothness)

Return ONLY valid JSON — no markdown, no explanation. Array of test objects.`
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd HelixQA && go test ./pkg/planning/ -v -run TestTestPlanGenerator`
Expected: PASS (3 tests)

- [ ] **Step 5: Run full test suites for both packages**

Run: `cd HelixQA && go test ./pkg/learning/ ./pkg/planning/ -v -race`
Expected: All tests pass, no races

- [ ] **Step 6: Commit**

```bash
git add HelixQA/pkg/planning/planner.go HelixQA/pkg/planning/planner_test.go
git commit -m "feat(helixqa): add TestPlanGenerator with LLM-driven test plan creation"
```

---

## Task 10: Final Integration & Push

- [ ] **Step 1: Run ALL HelixQA tests**

Run: `cd HelixQA && go test ./... -race -count=1 -v 2>&1 | tail -30`
Expected: All packages pass, no regressions

- [ ] **Step 2: Run go vet**

Run: `cd HelixQA && go vet ./...`
Expected: Clean

- [ ] **Step 3: Count total tests**

Run: `cd HelixQA && go test ./... -v 2>&1 | grep -c "--- PASS"`
Expected: ~660+ (636 prior + ~24 new)

- [ ] **Step 4: Commit and push**

```bash
cd /run/media/milosvasic/DATA4TB/Projects/Catalogizer
git add HelixQA
git commit -m "feat(helixqa): Phase 2 — learning engine + planning engine"
GIT_SSH_COMMAND="ssh -o BatchMode=yes" git push origin main
```
