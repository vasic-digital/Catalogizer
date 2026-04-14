# Distributed Build System Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Extend the Catalogizer build system to distribute containerized component builds across multiple hosts (local, thinker.local, amber.local) based on real-time resource availability, using the existing Containers submodule's scheduler, SSH executor, and remote execution infrastructure.

**Architecture:** A new Go binary `cmd/distributed-build/main.go` in the Containers submodule orchestrates builds. It probes all configured hosts, uses the resource-aware scheduler to assign components to hosts, synchronizes source code via rsync, launches builder containers on remote hosts via SSH, and collects artifacts back to the local `releases/` directory. The existing `release-build.sh` shell pipeline remains untouched — the Go binary wraps and dispatches it.

**Tech Stack:** Go 1.25, Containers submodule (`pkg/scheduler`, `pkg/remote`, `pkg/distribution`, `pkg/envconfig`, `pkg/volume`), SSH/SCP/rsync, Podman containers.

---

## File Structure

| Action | Path | Responsibility |
|--------|------|---------------|
| Create | `Containers/cmd/distributed-build/main.go` | CLI entry point — flags, config loading, host registration, orchestration |
| Create | `Containers/internal/buildpkg/types.go` | Build-specific types: BuildComponent, BuildHost, BuildResult, BuildPlan |
| Create | `Containers/internal/buildpkg/planner.go` | Maps components → scheduler requirements, produces BuildPlan |
| Create | `Containers/internal/buildpkg/executor.go` | Source sync, remote container launch, artifact collection |
| Create | `Containers/internal/buildpkg/executor_test.go` | Unit tests for executor |
| Create | `Containers/internal/buildpkg/planner_test.go` | Unit tests for planner |
| Create | `Containers/internal/buildpkg/types_test.go` | Unit tests for type methods |
| Create | `Containers/internal/buildpkg/artifacts.go` | Artifact discovery and collection from remote hosts |
| Create | `Containers/internal/buildpkg/artifacts_test.go` | Unit tests for artifact collection |
| Modify | `Containers/go.mod` | Add `digital.vasic.containers/internal/buildpkg` if needed (implicit) |
| Modify | `.env.example` | Add `BUILD_DISTRIBUTED_*` and `BUILD_HOST_*` env vars |
| Modify | `scripts/release-build.sh` | Add `--distributed` flag passthrough (minimal change) |

---

## Task 1: Build Types

**Files:**
- Create: `Containers/internal/buildpkg/types.go`
- Test: `Containers/internal/buildpkg/types_test.go`

- [ ] **Step 1: Write the failing test**

```go
// Containers/internal/buildpkg/types_test.go
package buildpkg

import (
	"testing"
	"time"
)

func TestBuildComponent_ResourceRequirements(t *testing.T) {
	tests := []struct {
		name       string
		component  BuildComponent
		wantCPU    float64
		wantMemMB  uint64
		wantDiskMB uint64
		wantLabels map[string]string
	}{
		{
			name: "catalog-api gets moderate resources",
			component: BuildComponent{
				Name:    "catalog-api",
				HasGo:   true,
				HasNPM:  false,
				HasJDK:  false,
				HasRust: false,
			},
			wantCPU:    2.0,
			wantMemMB:  2048,
			wantDiskMB: 1024,
			wantLabels: map[string]string{"go": "true"},
		},
		{
			name: "catalogizer-androidtv gets heavy resources with jdk label",
			component: BuildComponent{
				Name:    "catalogizer-androidtv",
				HasGo:   false,
				HasNPM:  false,
				HasJDK:  true,
				HasRust: false,
			},
			wantCPU:    3.0,
			wantMemMB:  4096,
			wantDiskMB: 2048,
			wantLabels: map[string]string{"jdk": "true"},
		},
		{
			name: "catalogizer-desktop gets heavy resources with rust label",
			component: BuildComponent{
				Name:    "catalogizer-desktop",
				HasGo:   false,
				HasNPM:  true,
				HasJDK:  false,
				HasRust: true,
			},
			wantCPU:    3.0,
			wantMemMB:  4096,
			wantDiskMB: 2048,
			wantLabels: map[string]string{"npm": "true", "rust": "true"},
		},
		{
			name: "catalog-web gets light resources",
			component: BuildComponent{
				Name:    "catalog-web",
				HasGo:   false,
				HasNPM:  true,
				HasJDK:  false,
				HasRust: false,
			},
			wantCPU:    1.0,
			wantMemMB:  1024,
			wantDiskMB: 512,
			wantLabels: map[string]string{"npm": "true"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := tt.component.ResourceRequirements()
			if req.CPUCores != tt.wantCPU {
				t.Errorf("CPUCores = %v, want %v", req.CPUCores, tt.wantCPU)
			}
			if req.MemoryMB != tt.wantMemMB {
				t.Errorf("MemoryMB = %v, want %v", req.MemoryMB, tt.wantMemMB)
			}
			if req.DiskMB != tt.wantDiskMB {
				t.Errorf("DiskMB = %v, want %v", req.DiskMB, tt.wantDiskMB)
			}
			for k, v := range tt.wantLabels {
				if req.Labels[k] != v {
					t.Errorf("Labels[%q] = %q, want %q", k, req.Labels[k], v)
				}
			}
		})
	}
}

func TestBuildResult_Success(t *testing.T) {
	r := BuildResult{
		Component:  "catalog-api",
		Host:       "thinker.local",
		Status:     BuildStatusSuccess,
		Duration:   45 * time.Second,
		ArtifactPath: "releases/catalog-api/linux-amd64/v2.2.0-build.19",
	}
	if !r.IsSuccess() {
		t.Error("expected IsSuccess() = true")
	}
	if r.IsFailure() {
		t.Error("expected IsFailure() = false")
	}
}

func TestBuildResult_Failure(t *testing.T) {
	r := BuildResult{
		Component: "catalog-web",
		Host:      "amber.local",
		Status:    BuildStatusFailed,
		Error:     "npm ci exit code 1",
	}
	if r.IsSuccess() {
		t.Error("expected IsSuccess() = false")
	}
	if !r.IsFailure() {
		t.Error("expected IsFailure() = true")
	}
}

func TestBuildPlan_Assignments(t *testing.T) {
	plan := &BuildPlan{
		Assignments: []BuildAssignment{
			{Component: BuildComponent{Name: "catalog-api"}, Host: "local"},
			{Component: BuildComponent{Name: "catalog-web"}, Host: "thinker.local"},
			{Component: BuildComponent{Name: "catalogizer-androidtv"}, Host: "amber.local"},
		},
	}

	local := plan.LocalAssignments()
	if len(local) != 1 || local[0].Component.Name != "catalog-api" {
		t.Errorf("LocalAssignments() = %v, want 1 assignment for catalog-api", local)
	}

	remote := plan.RemoteAssignments()
	if len(remote) != 2 {
		t.Errorf("RemoteAssignments() returned %d, want 2", len(remote))
	}

	byHost := plan.ByHost()
	if len(byHost["thinker.local"]) != 1 {
		t.Errorf("ByHost()[thinker.local] = %v, want 1", byHost["thinker.local"])
	}
}

func TestAllComponents(t *testing.T) {
	comps := AllComponents()
	names := make(map[string]bool)
	for _, c := range comps {
		names[c.Name] = true
	}
	expected := []string{
		"catalog-api", "catalog-web", "catalogizer-api-client",
		"catalogizer-desktop", "installer-wizard",
		"catalogizer-android", "catalogizer-androidtv",
	}
	for _, n := range expected {
		if !names[n] {
			t.Errorf("AllComponents() missing %q", n)
		}
	}
	if len(comps) != len(expected) {
		t.Errorf("AllComponents() returned %d, want %d", len(comps), len(expected))
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd Containers && go test ./internal/buildpkg/ -run "TestBuild" -v`
Expected: FAIL — package does not exist

- [ ] **Step 3: Write minimal implementation**

```go
// Containers/internal/buildpkg/types.go
package buildpkg

import (
	"fmt"
	"time"

	"digital.vasic.containers/pkg/scheduler"
)

type BuildStatus string

const (
	BuildStatusPending BuildStatus = "pending"
	BuildStatusRunning BuildStatus = "running"
	BuildStatusSuccess BuildStatus = "success"
	BuildStatusFailed  BuildStatus = "failed"
	BuildStatusSkipped BuildStatus = "skipped"
)

type BuildComponent struct {
	Name    string
	HasGo   bool
	HasNPM  bool
	HasJDK  bool
	HasRust bool
}

func (bc BuildComponent) ResourceRequirements() scheduler.ContainerRequirements {
	req := scheduler.ContainerRequirements{
		Name:   bc.Name,
		Image:  "localhost/catalogizer-builder:latest",
		Labels: make(map[string]string),
	}

	if bc.HasJDK {
		req.CPUCores = 3.0
		req.MemoryMB = 4096
		req.DiskMB = 2048
		req.Labels["jdk"] = "true"
	} else if bc.HasRust {
		req.CPUCores = 3.0
		req.MemoryMB = 4096
		req.DiskMB = 2048
		req.Labels["rust"] = "true"
	} else if bc.HasGo {
		req.CPUCores = 2.0
		req.MemoryMB = 2048
		req.DiskMB = 1024
		req.Labels["go"] = "true"
	} else {
		req.CPUCores = 1.0
		req.MemoryMB = 1024
		req.DiskMB = 512
	}

	if bc.HasNPM {
		req.Labels["npm"] = "true"
	}

	return req
}

type BuildResult struct {
	Component    string
	Host         string
	Status       BuildStatus
	Duration     time.Duration
	ArtifactPath string
	Error        string
}

func (r BuildResult) IsSuccess() bool  { return r.Status == BuildStatusSuccess }
func (r BuildResult) IsFailure() bool  { return r.Status == BuildStatusFailed }

type BuildAssignment struct {
	Component BuildComponent
	Host      string
}

type BuildPlan struct {
	Assignments []BuildAssignment
}

func (bp *BuildPlan) LocalAssignments() []BuildAssignment {
	var result []BuildAssignment
	for _, a := range bp.Assignments {
		if a.Host == "" || a.Host == "local" {
			result = append(result, a)
		}
	}
	return result
}

func (bp *BuildPlan) RemoteAssignments() []BuildAssignment {
	var result []BuildAssignment
	for _, a := range bp.Assignments {
		if a.Host != "" && a.Host != "local" {
			result = append(result, a)
		}
	}
	return result
}

func (bp *BuildPlan) ByHost() map[string][]BuildAssignment {
	result := make(map[string][]BuildAssignment)
	for _, a := range bp.Assignments {
		result[a.Host] = append(result[a.Host], a)
	}
	return result
}

func AllComponents() []BuildComponent {
	return []BuildComponent{
		{Name: "catalog-api", HasGo: true},
		{Name: "catalog-web", HasNPM: true},
		{Name: "catalogizer-api-client", HasNPM: true},
		{Name: "catalogizer-desktop", HasNPM: true, HasRust: true},
		{Name: "installer-wizard", HasNPM: true, HasRust: true},
		{Name: "catalogizer-android", HasNPM: true, HasJDK: true},
		{Name: "catalogizer-androidtv", HasNPM: true, HasJDK: true},
	}
}

func FindComponent(name string) (BuildComponent, error) {
	for _, c := range AllComponents() {
		if c.Name == name {
			return c, nil
		}
	}
	return BuildComponent{}, fmt.Errorf("component %q not found", name)
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd Containers && go test ./internal/buildpkg/ -run "TestBuild" -v`
Expected: PASS (all 4 tests)

- [ ] **Step 5: Commit**

```bash
cd Containers && git add internal/buildpkg/types.go internal/buildpkg/types_test.go && git commit -m "feat(build): add build component types with resource requirements"
```

---

## Task 2: Build Planner

**Files:**
- Create: `Containers/internal/buildpkg/planner.go`
- Test: `Containers/internal/buildpkg/planner_test.go`

- [ ] **Step 1: Write the failing test**

```go
// Containers/internal/buildpkg/planner_test.go
package buildpkg

import (
	"context"
	"testing"

	"digital.vasic.containers/pkg/remote"
)

type stubHostManager struct {
	hosts     []remote.RemoteHost
	resources map[string]*remote.HostResources
}

func (s *stubHostManager) AddHost(host remote.RemoteHost) error {
	s.hosts = append(s.hosts, host)
	return nil
}
func (s *stubHostManager) RemoveHost(name string) error {
	return nil
}
func (s *stubHostManager) GetHost(name string) (*remote.RemoteHost, error) {
	for i := range s.hosts {
		if s.hosts[i].Name == name {
			return &s.hosts[i], nil
		}
	}
	return nil, fmt.Errorf("not found")
}
func (s *stubHostManager) ListHosts() []remote.RemoteHost { return s.hosts }
func (s *stubHostManager) ProbeHost(ctx context.Context, name string) (*remote.HostResources, error) {
	r, ok := s.resources[name]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return r, nil
}
func (s *stubHostManager) ProbeAll(ctx context.Context) map[string]*remote.HostResources {
	return s.resources
}
func (s *stubHostManager) HostState(name string) remote.HostState {
	return remote.HostOnline
}

func TestPlanner_PlanAll(t *testing.T) {
	hm := &stubHostManager{
		hosts: []remote.RemoteHost{
			{Name: "local", Address: "127.0.0.1"},
			{Name: "thinker", Address: "thinker.local", Labels: map[string]string{"go": "true", "npm": "true", "jdk": "true", "rust": "true"}},
			{Name: "amber", Address: "amber.local", Labels: map[string]string{"go": "true", "npm": "true", "jdk": "true", "rust": "true"}},
		},
		resources: map[string]*remote.HostResources{
			"local":   {Host: "local", CPUCores: 4, CPUPercent: 20, MemoryPercent: 30, MemoryTotalMB: 32000},
			"thinker": {Host: "thinker", CPUCores: 12, CPUPercent: 15, MemoryPercent: 20, MemoryTotalMB: 64000},
			"amber":   {Host: "amber", CPUCores: 8, CPUPercent: 10, MemoryPercent: 15, MemoryTotalMB: 32000},
		},
	}

	planner := NewPlanner(hm)
	plan, err := planner.PlanAll(context.Background())
	if err != nil {
		t.Fatalf("PlanAll() error: %v", err)
	}

	if len(plan.Assignments) != 7 {
		t.Fatalf("expected 7 assignments, got %d", len(plan.Assignments))
	}

	assigned := make(map[string]string)
	for _, a := range plan.Assignments {
		assigned[a.Component.Name] = a.Host
	}

	for _, comp := range AllComponents() {
		host, ok := assigned[comp.Name]
		if !ok {
			t.Errorf("component %q not assigned", comp.Name)
			continue
		}
		if host == "" {
			t.Errorf("component %q assigned to empty host", comp.Name)
		}
	}
}

func TestPlanner_PlanSingle(t *testing.T) {
	hm := &stubHostManager{
		hosts: []remote.RemoteHost{
			{Name: "local", Address: "127.0.0.1"},
		},
		resources: map[string]*remote.HostResources{
			"local": {Host: "local", CPUCores: 4, CPUPercent: 20, MemoryPercent: 30, MemoryTotalMB: 32000},
		},
	}

	planner := NewPlanner(hm)
	plan, err := planner.PlanSingle(context.Background(), "catalog-api")
	if err != nil {
		t.Fatalf("PlanSingle() error: %v", err)
	}

	if len(plan.Assignments) != 1 {
		t.Fatalf("expected 1 assignment, got %d", len(plan.Assignments))
	}
	if plan.Assignments[0].Component.Name != "catalog-api" {
		t.Errorf("assigned component = %q, want catalog-api", plan.Assignments[0].Component.Name)
	}
}

func TestPlanner_PlanSingleUnknownComponent(t *testing.T) {
	hm := &stubHostManager{
		hosts:     []remote.RemoteHost{{Name: "local", Address: "127.0.0.1"}},
		resources: map[string]*remote.HostResources{"local": {Host: "local"}},
	}

	planner := NewPlanner(hm)
	_, err := planner.PlanSingle(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for unknown component")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd Containers && go test ./internal/buildpkg/ -run "TestPlanner" -v`
Expected: FAIL — `NewPlanner` undefined

- [ ] **Step 3: Write minimal implementation**

```go
// Containers/internal/buildpkg/planner.go
package buildpkg

import (
	"context"
	"fmt"

	"digital.vasic.containers/pkg/remote"
	"digital.vasic.containers/pkg/scheduler"
)

type Planner struct {
	hostManager remote.HostManager
	scheduler   scheduler.Scheduler
}

func NewPlanner(hostManager remote.HostManager) *Planner {
	sched := scheduler.NewScheduler(
		hostManager,
		nil,
		scheduler.WithStrategy(scheduler.StrategyResourceAware),
	)
	return &Planner{
		hostManager: hostManager,
		scheduler:   sched,
	}
}

func NewPlannerWithScheduler(hostManager remote.HostManager, sched scheduler.Scheduler) *Planner {
	return &Planner{
		hostManager: hostManager,
		scheduler:   sched,
	}
}

func (p *Planner) PlanAll(ctx context.Context) (*BuildPlan, error) {
	components := AllComponents()
	return p.plan(ctx, components)
}

func (p *Planner) PlanSingle(ctx context.Context, componentName string) (*BuildPlan, error) {
	comp, err := FindComponent(componentName)
	if err != nil {
		return nil, err
	}
	return p.plan(ctx, []BuildComponent{comp})
}

func (p *Planner) plan(ctx context.Context, components []BuildComponent) (*BuildPlan, error) {
	var reqs []scheduler.ContainerRequirements
	for _, comp := range components {
		reqs = append(reqs, comp.ResourceRequirements())
	}

	placementPlan, err := p.scheduler.ScheduleBatch(ctx, reqs)
	if err != nil {
		return nil, fmt.Errorf("scheduling failed: %w", err)
	}

	plan := &BuildPlan{}
	for i, decision := range placementPlan.Decisions {
		if i >= len(components) {
			break
		}
		hostName := decision.HostName
		if hostName == "" {
			hostName = "local"
		}
		plan.Assignments = append(plan.Assignments, BuildAssignment{
			Component: components[i],
			Host:      hostName,
		})
	}

	return plan, nil
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd Containers && go test ./internal/buildpkg/ -run "TestPlanner" -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
cd Containers && git add internal/buildpkg/planner.go internal/buildpkg/planner_test.go && git commit -m "feat(build): add build planner with resource-aware scheduling"
```

---

## Task 3: Build Executor — Source Sync and Remote Build Launch

**Files:**
- Create: `Containers/internal/buildpkg/executor.go`
- Test: `Containers/internal/buildpkg/executor_test.go`

- [ ] **Step 1: Write the failing test**

```go
// Containers/internal/buildpkg/executor_test.go
package buildpkg

import (
	"context"
	"strings"
	"testing"
	"time"

	"digital.vasic.containers/pkg/remote"
)

type mockExecutor struct {
	executedCommands []execRecord
	copiedFiles      []copyRecord
	reachable        map[string]bool
}

type execRecord struct {
	host    remote.RemoteHost
	command string
}

type copyRecord struct {
	host       remote.RemoteHost
	localPath  string
	remotePath string
}

func (m *mockExecutor) Execute(ctx context.Context, host remote.RemoteHost, command string) (*remote.CommandResult, error) {
	m.executedCommands = append(m.executedCommands, execRecord{host: host, command: command})
	return &remote.CommandResult{Stdout: "ok", ExitCode: 0}, nil
}

func (m *mockExecutor) ExecuteStream(ctx context.Context, host remote.RemoteHost, command string) (interface{}, error) {
	return nil, nil
}

func (m *mockExecutor) CopyFile(ctx context.Context, host remote.RemoteHost, localPath, remotePath string) error {
	m.copiedFiles = append(m.copiedFiles, copyRecord{host: host, localPath: localPath, remotePath: remotePath})
	return nil
}

func (m *mockExecutor) CopyDir(ctx context.Context, host remote.RemoteHost, localDir, remoteDir string) error {
	m.copiedFiles = append(m.copiedFiles, copyRecord{host: host, localPath: localDir, remotePath: remoteDir})
	return nil
}

func (m *mockExecutor) IsReachable(ctx context.Context, host remote.RemoteHost) bool {
	return m.reachable[host.Name]
}

func TestBuildExecutor_SyncSource(t *testing.T) {
	mock := &mockExecutor{reachable: map[string]bool{"thinker": true}}
	exec := NewBuildExecutor(mock, "/project", "/tmp/catalogizer-build")

	host := remote.RemoteHost{Name: "thinker", Address: "thinker.local", User: "milosvasic"}
	err := exec.SyncSource(context.Background(), host)
	if err != nil {
		t.Fatalf("SyncSource() error: %v", err)
	}

	if len(mock.copiedFiles) == 0 {
		t.Fatal("expected CopyDir to be called")
	}
	cf := mock.copiedFiles[0]
	if cf.localPath != "/project" {
		t.Errorf("CopyDir localPath = %q, want /project", cf.localPath)
	}
	if cf.remotePath != "/tmp/catalogizer-build" {
		t.Errorf("CopyDir remotePath = %q, want /tmp/catalogizer-build", cf.remotePath)
	}
}

func TestBuildExecutor_SyncSourceUnreachable(t *testing.T) {
	mock := &mockExecutor{reachable: map[string]bool{"thinker": false}}
	exec := NewBuildExecutor(mock, "/project", "/tmp/catalogizer-build")

	host := remote.RemoteHost{Name: "thinker", Address: "thinker.local", User: "milosvasic"}
	err := exec.SyncSource(context.Background(), host)
	if err == nil {
		t.Fatal("expected error for unreachable host")
	}
}

func TestBuildExecutor_LaunchRemoteBuild(t *testing.T) {
	mock := &mockExecutor{reachable: map[string]bool{"thinker": true}}
	exec := NewBuildExecutor(mock, "/project", "/tmp/catalogizer-build")

	host := remote.RemoteHost{Name: "thinker", Address: "thinker.local", User: "milosvasic"}
	result, err := exec.LaunchRemoteBuild(context.Background(), host, "catalog-api", "v2.2.0-build.19", true)
	if err != nil {
		t.Fatalf("LaunchRemoteBuild() error: %v", err)
	}

	if result.Host != "thinker" {
		t.Errorf("result.Host = %q, want thinker", result.Host)
	}
	if result.Component != "catalog-api" {
		t.Errorf("result.Component = %q, want catalog-api", result.Component)
	}

	found := false
	for _, cmd := range mock.executedCommands {
		if strings.Contains(cmd.command, "catalog-api") && strings.Contains(cmd.command, "v2.2.0-build.19") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected build command containing component name and version, got commands: %v", mock.executedCommands)
	}
}

func TestBuildExecutor_BuildTimeout(t *testing.T) {
	mock := &mockExecutor{reachable: map[string]bool{"thinker": true}}
	exec := NewBuildExecutor(mock, "/project", "/tmp/catalogizer-build")
	exec = exec.WithBuildTimeout(1 * time.Nanosecond)

	host := remote.RemoteHost{Name: "thinker", Address: "thinker.local", User: "milosvasic"}
	ctx := context.Background()
	_, err := exec.LaunchRemoteBuild(ctx, host, "catalog-api", "v2.2.0-build.19", true)
	if err == nil {
		t.Fatal("expected timeout error")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd Containers && go test ./internal/buildpkg/ -run "TestBuildExecutor" -v`
Expected: FAIL — `NewBuildExecutor` undefined

- [ ] **Step 3: Write minimal implementation**

```go
// Containers/internal/buildpkg/executor.go
package buildpkg

import (
	"context"
	"fmt"
	"strings"
	"time"

	"digital.vasic.containers/pkg/remote"
)

type RemoteExecutor interface {
	Execute(ctx context.Context, host remote.RemoteHost, command string) (*remote.CommandResult, error)
	CopyDir(ctx context.Context, host remote.RemoteHost, localDir, remoteDir string) error
	IsReachable(ctx context.Context, host remote.RemoteHost) bool
}

type BuildExecutor struct {
	executor    RemoteExecutor
	projectDir  string
	remoteDir   string
	buildTimeout time.Duration
}

func NewBuildExecutor(executor RemoteExecutor, projectDir, remoteDir string) *BuildExecutor {
	return &BuildExecutor{
		executor:     executor,
		projectDir:   projectDir,
		remoteDir:    remoteDir,
		buildTimeout: 30 * time.Minute,
	}
}

func (be *BuildExecutor) WithBuildTimeout(d time.Duration) *BuildExecutor {
	return &BuildExecutor{
		executor:     be.executor,
		projectDir:   be.projectDir,
		remoteDir:    be.remoteDir,
		buildTimeout: d,
	}
}

func (be *BuildExecutor) SyncSource(ctx context.Context, host remote.RemoteHost) error {
	if !be.executor.IsReachable(ctx, host) {
		return fmt.Errorf("host %s (%s) is not reachable", host.Name, host.Address)
	}

	mkdirCmd := fmt.Sprintf("mkdir -p %s", be.remoteDir)
	_, err := be.executor.Execute(ctx, host, mkdirCmd)
	if err != nil {
		return fmt.Errorf("creating remote directory: %w", err)
	}

	err = be.executor.CopyDir(ctx, host, be.projectDir, be.remoteDir)
	if err != nil {
		return fmt.Errorf("copying source to %s: %w", host.Name, err)
	}

	return nil
}

func (be *BuildExecutor) LaunchRemoteBuild(ctx context.Context, host remote.RemoteHost, component, versionString string, skipTests bool) (*BuildResult, error) {
	if !be.executor.IsReachable(ctx, host) {
		return nil, fmt.Errorf("host %s (%s) is not reachable", host.Name, host.Address)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("cd %s && ", be.remoteDir))
	sb.WriteString("git submodule update --init --recursive 2>/dev/null; ")
	sb.WriteString(fmt.Sprintf("/project/scripts/release-build.sh --local --component %s --force", component))

	if skipTests {
		sb.WriteString(" --skip-tests")
	}

	cmd := sb.String()

	start := time.Now()
	result, err := be.executor.Execute(ctx, host, cmd)
	duration := time.Since(start)

	br := &BuildResult{
		Component: component,
		Host:      host.Name,
		Duration:  duration,
	}

	if err != nil {
		br.Status = BuildStatusFailed
		br.Error = fmt.Sprintf("execution error: %v", err)
		return br, err
	}

	if result.ExitCode != 0 {
		br.Status = BuildStatusFailed
		br.Error = fmt.Sprintf("build exited with code %d: %s", result.ExitCode, truncateString(result.Stderr, 500))
		return br, fmt.Errorf("build failed on %s: exit code %d", host.Name, result.ExitCode)
	}

	br.Status = BuildStatusSuccess
	return br, nil
}

func (be *BuildExecutor) LaunchLocalBuild(ctx context.Context, component, versionString string, skipTests bool) (*BuildResult, error) {
	return nil, fmt.Errorf("local builds are handled by the shell pipeline, not the Go executor")
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd Containers && go test ./internal/buildpkg/ -run "TestBuildExecutor" -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
cd Containers && git add internal/buildpkg/executor.go internal/buildpkg/executor_test.go && git commit -m "feat(build): add build executor with source sync and remote build launch"
```

---

## Task 4: Artifact Collection

**Files:**
- Create: `Containers/internal/buildpkg/artifacts.go`
- Test: `Containers/internal/buildpkg/artifacts_test.go`

- [ ] **Step 1: Write the failing test**

```go
// Containers/internal/buildpkg/artifacts_test.go
package buildpkg

import (
	"context"
	"testing"

	"digital.vasic.containers/pkg/remote"
)

func TestArtifactCollector_DiscoverArtifacts(t *testing.T) {
	mock := &mockExecutor{reachable: map[string]bool{"thinker": true}}
	collector := NewArtifactCollector(mock, "/project", "/tmp/catalogizer-build")

	host := remote.RemoteHost{Name: "thinker", Address: "thinker.local", User: "milosvasic"}
	paths, err := collector.DiscoverArtifacts(context.Background(), host, "catalog-api", "v2.2.0-build.19")
	if err != nil {
		t.Fatalf("DiscoverArtifacts() error: %v", err)
	}

	if len(paths) == 0 {
		t.Fatal("expected at least one artifact path")
	}
}

func TestArtifactCollector_CollectArtifacts(t *testing.T) {
	mock := &mockExecutor{reachable: map[string]bool{"thinker": true}}
	collector := NewArtifactCollector(mock, "/project", "/tmp/catalogizer-build")

	host := remote.RemoteHost{Name: "thinker", Address: "thinker.local", User: "milosvasic"}
	remotePaths := []string{
		"/tmp/catalogizer-build/releases/catalog-api/linux-amd64/v2.2.0-build.19/catalog-api",
	}

	err := collector.CollectArtifacts(context.Background(), host, remotePaths)
	if err != nil {
		t.Fatalf("CollectArtifacts() error: %v", err)
	}

	if len(mock.copiedFiles) == 0 {
		t.Fatal("expected CopyDir to be called for artifact collection")
	}
}

func TestArtifactCollector_CollectFromUnreachableHost(t *testing.T) {
	mock := &mockExecutor{reachable: map[string]bool{"thinker": false}}
	collector := NewArtifactCollector(mock, "/project", "/tmp/catalogizer-build")

	host := remote.RemoteHost{Name: "thinker", Address: "thinker.local", User: "milosvasic"}
	_, err := collector.DiscoverArtifacts(context.Background(), host, "catalog-api", "v2.2.0-build.19")
	if err == nil {
		t.Fatal("expected error for unreachable host")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd Containers && go test ./internal/buildpkg/ -run "TestArtifactCollector" -v`
Expected: FAIL — `NewArtifactCollector` undefined

- [ ] **Step 3: Write minimal implementation**

```go
// Containers/internal/buildpkg/artifacts.go
package buildpkg

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"digital.vasic.containers/pkg/remote"
)

type ArtifactCollector struct {
	executor   RemoteExecutor
	projectDir string
	remoteDir  string
}

func NewArtifactCollector(executor RemoteExecutor, projectDir, remoteDir string) *ArtifactCollector {
	return &ArtifactCollector{
		executor:   executor,
		projectDir: projectDir,
		remoteDir:  remoteDir,
	}
}

func (ac *ArtifactCollector) DiscoverArtifacts(ctx context.Context, host remote.RemoteHost, component, versionString string) ([]string, error) {
	if !ac.executor.IsReachable(ctx, host) {
		return nil, fmt.Errorf("host %s is not reachable", host.Name)
	}

	releaseBase := filepath.Join(ac.remoteDir, "releases", component)
	findCmd := fmt.Sprintf("find %s -name 'BUILD_INFO.json' -path '*%s*'", releaseBase, versionString)

	result, err := ac.executor.Execute(ctx, host, findCmd)
	if err != nil {
		return nil, fmt.Errorf("discovering artifacts on %s: %w", host.Name, err)
	}

	if result.ExitCode != 0 {
		return nil, fmt.Errorf("find command failed: %s", result.Stderr)
	}

	var paths []string
	for _, line := range strings.Split(strings.TrimSpace(result.Stdout), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			dir := filepath.Dir(line)
			paths = append(paths, dir)
		}
	}

	return paths, nil
}

func (ac *ArtifactCollector) CollectArtifacts(ctx context.Context, host remote.RemoteHost, remotePaths []string) error {
	if !ac.executor.IsReachable(ctx, host) {
		return fmt.Errorf("host %s is not reachable", host.Name)
	}

	for _, remotePath := range remotePaths {
		relPath := ""
		if strings.HasPrefix(remotePath, ac.remoteDir+"/") {
			relPath = strings.TrimPrefix(remotePath, ac.remoteDir+"/")
		} else {
			relPath = filepath.Base(remotePath)
		}

		localPath := filepath.Join(ac.projectDir, relPath)
		parentDir := filepath.Dir(localPath)

		mkdirCmd := fmt.Sprintf("mkdir -p %s", parentDir)
		result, err := ac.executor.Execute(ctx, host, mkdirCmd)
		if err != nil {
			return fmt.Errorf("creating local directory %s: %w", parentDir, err)
		}
		_ = result

		err = ac.executor.CopyDir(ctx, host, remotePath, localPath)
		if err != nil {
			return fmt.Errorf("collecting artifacts from %s:%s: %w", host.Name, remotePath, err)
		}
	}

	return nil
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd Containers && go test ./internal/buildpkg/ -run "TestArtifactCollector" -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
cd Containers && git add internal/buildpkg/artifacts.go internal/buildpkg/artifacts_test.go && git commit -m "feat(build): add artifact discovery and collection from remote hosts"
```

---

## Task 5: CLI Entry Point — distributed-build Command

**Files:**
- Create: `Containers/cmd/distributed-build/main.go`

- [ ] **Step 1: Write the implementation**

```go
// Containers/cmd/distributed-build/main.go
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"digital.vasic.containers/internal/buildpkg"
	"digital.vasic.containers/pkg/envconfig"
	"digital.vasic.containers/pkg/remote"
	"digital.vasic.containers/pkg/scheduler"
)

func main() {
	var (
		projectDir  string
		envFile     string
		component   string
		versionStr  string
		skipTests   bool
		dryRun      bool
		timeout     int
		strategy    string
	)
	flag.StringVar(&projectDir, "project", ".", "Path to Catalogizer project root")
	flag.StringVar(&envFile, "env", ".env", "Path to .env file with host configuration")
	flag.StringVar(&component, "component", "", "Build single component (default: all)")
	flag.StringVar(&versionStr, "version", "", "Version string (default: auto-detect)")
	flag.BoolVar(&skipTests, "skip-tests", false, "Skip test execution")
	flag.BoolVar(&dryRun, "dry-run", false, "Show plan without executing")
	flag.IntVar(&timeout, "timeout", 30, "Build timeout in minutes")
	flag.StringVar(&strategy, "strategy", "resource_aware", "Scheduling strategy")
	flag.Parse()

	ctx := context.Background()

	absProject, err := filepath.Abs(projectDir)
	if err != nil {
		log.Fatalf("resolving project path: %v", err)
	}

	remoteDir := fmt.Sprintf("/tmp/catalogizer-build-%d", time.Now().UnixMilli())

	cfg, err := loadConfig(envFile)
	if err != nil {
		log.Fatalf("loading config: %v", err)
	}

	sshExec, err := remote.NewSSHExecutor(nil)
	if err != nil {
		log.Fatalf("creating SSH executor: %v", err)
	}
	defer sshExec.Close()

	hostMgr := remote.NewHostManager(sshExec, nil)
	for _, h := range cfg.ToRemoteHosts() {
		if err := hostMgr.AddHost(h); err != nil {
			log.Printf("warning: could not add host %s: %v", h.Name, err)
		}
	}

	schedStrategy := scheduler.PlacementStrategy(strategy)
	sched := scheduler.NewScheduler(hostMgr, nil, scheduler.WithStrategy(schedStrategy))

	planner := buildpkg.NewPlannerWithScheduler(hostMgr, sched)

	var plan *buildpkg.BuildPlan
	if component != "" {
		plan, err = planner.PlanSingle(ctx, component)
		if err != nil {
			log.Fatalf("planning single component: %v", err)
		}
	} else {
		plan, err = planner.PlanAll(ctx)
		if err != nil {
			log.Fatalf("planning all components: %v", err)
		}
	}

	fmt.Println("=== Distributed Build Plan ===")
	for _, a := range plan.Assignments {
		fmt.Printf("  %-30s -> %s\n", a.Component.Name, a.Host)
	}

	if dryRun {
		fmt.Println("\n(dry run — no builds executed)")
		return
	}

	buildExec := buildpkg.NewBuildExecutor(sshExec, absProject, remoteDir)
	buildExec = buildExec.WithBuildTimeout(time.Duration(timeout) * time.Minute)

	artifactCollector := buildpkg.NewArtifactCollector(sshExec, absProject, remoteDir)

	fmt.Println("\n=== Syncing Source ===")
	remoteAssignments := plan.RemoteAssignments()
	hostSet := make(map[string]string)
	for _, a := range remoteAssignments {
		hostSet[a.Host] = a.Component.Name
	}

	for hostName := range hostSet {
		host, err := hostMgr.GetHost(hostName)
		if err != nil {
			log.Printf("warning: could not get host %s: %v", hostName, err)
			continue
		}
		fmt.Printf("  Syncing source to %s (%s)...\n", host.Name, host.Address)
		if err := buildExec.SyncSource(ctx, *host); err != nil {
			log.Fatalf("sync failed for %s: %v", host.Name, err)
		}
	}

	fmt.Println("\n=== Executing Builds ===")
	var results []buildpkg.BuildResult

	for _, a := range plan.LocalAssignments() {
		fmt.Printf("  [local] %s...\n", a.Component.Name)
		fmt.Printf("    (local builds are handled by shell pipeline, skipping in Go executor)\n")
	}

	for _, a := range remoteAssignments {
		host, err := hostMgr.GetHost(a.Host)
		if err != nil {
			log.Printf("error: could not get host %s: %v", a.Host, err)
			results = append(results, buildpkg.BuildResult{
				Component: a.Component.Name,
				Host:      a.Host,
				Status:    buildpkg.BuildStatusFailed,
				Error:     fmt.Sprintf("host lookup failed: %v", err),
			})
			continue
		}

		fmt.Printf("  [%s] %s...\n", a.Host, a.Component.Name)
		result, err := buildExec.LaunchRemoteBuild(ctx, *host, a.Component.Name, versionStr, skipTests)
		if err != nil {
			fmt.Printf("    FAILED: %v\n", err)
		} else {
			fmt.Printf("    SUCCESS (%s)\n", result.Duration.Round(time.Second))
		}
		if result != nil {
			results = append(results, *result)
		}
	}

	fmt.Println("\n=== Collecting Artifacts ===")
	for _, a := range remoteAssignments {
		host, err := hostMgr.GetHost(a.Host)
		if err != nil {
			continue
		}

		for _, r := range results {
			if r.Component == a.Component.Name && r.IsSuccess() {
				paths, err := artifactCollector.DiscoverArtifacts(ctx, *host, a.Component.Name, versionStr)
				if err != nil {
					log.Printf("  artifact discovery failed for %s on %s: %v", a.Component.Name, a.Host, err)
					continue
				}
				if len(paths) > 0 {
					fmt.Printf("  Collecting %d artifact(s) from %s:%s\n", len(paths), a.Host, a.Component.Name)
					if err := artifactCollector.CollectArtifacts(ctx, *host, paths); err != nil {
						log.Printf("  artifact collection failed: %v", err)
					}
				}
			}
		}
	}

	fmt.Println("\n=== Build Results ===")
	successCount := 0
	failCount := 0
	for _, r := range results {
		status := string(r.Status)
		if r.IsSuccess() {
			status = "OK"
			successCount++
		} else {
			status = "FAIL"
			failCount++
		}
		fmt.Printf("  %-30s [%-4s] on %-20s (%s)\n", r.Component, status, r.Host, r.Duration.Round(time.Second))
		if r.Error != "" {
			fmt.Printf("    Error: %s\n", r.Error)
		}
	}

	fmt.Printf("\nTotal: %d succeeded, %d failed\n", successCount, failCount)
	if failCount > 0 {
		os.Exit(1)
	}

	for hostName := range hostSet {
		host, err := hostMgr.GetHost(hostName)
		if err != nil {
			continue
		}
		cleanupCmd := fmt.Sprintf("rm -rf %s", remoteDir)
		_, _ = sshExec.Execute(ctx, *host, cleanupCmd)
	}
}

func loadConfig(envFile string) (*envconfig.DistributionConfig, error) {
	if _, err := os.Stat(envFile); err == nil {
		return envconfig.LoadFromFile(envFile)
	}
	cfg := envconfig.LoadFromEnv()
	if len(cfg.Hosts) == 0 {
		return nil, fmt.Errorf("no remote hosts configured — set CONTAINERS_REMOTE_HOST_* env vars or use --env flag")
	}
	return cfg, nil
}
```

- [ ] **Step 2: Run build to verify compilation**

Run: `cd Containers && go build ./cmd/distributed-build/`
Expected: Builds successfully with no errors

- [ ] **Step 3: Commit**

```bash
cd Containers && git add cmd/distributed-build/main.go && git commit -m "feat(build): add distributed-build CLI command"
```

---

## Task 6: Environment Configuration — .env.example Update

**Files:**
- Modify: `.env.example`

- [ ] **Step 1: Add distributed build env vars to .env.example**

Append the following block to `.env.example` after the existing `HELIX_VISION_*` section:

```env
# Distributed Build Configuration
# When enabled, builds are distributed across multiple hosts based on resource availability.
# Hosts must have passwordless SSH access and a container runtime (Podman/Docker).

BUILD_DISTRIBUTED=false
BUILD_HOST_1_NAME=thinker
BUILD_HOST_1_ADDRESS=thinker.local
BUILD_HOST_1_USER=milosvasic
BUILD_HOST_1_KEY_PATH=~/.ssh/id_ed25519
BUILD_HOST_1_RUNTIME=podman
BUILD_HOST_1_LABELS=go=true,npm=true,jdk=true,rust=true

BUILD_HOST_2_NAME=amber
BUILD_HOST_2_ADDRESS=amber.local
BUILD_HOST_2_USER=milosvasic
BUILD_HOST_2_KEY_PATH=~/.ssh/id_ed25519
BUILD_HOST_2_RUNTIME=podman
BUILD_HOST_2_LABELS=go=true,npm=true,jdk=true

BUILD_SCHEDULER_STRATEGY=resource_aware
BUILD_REMOTE_DIR=/tmp/catalogizer-build
BUILD_TIMEOUT_MINUTES=30
```

- [ ] **Step 2: Verify the file is valid**

Run: `grep -c "BUILD_HOST" .env.example`
Expected: 8 or more (the new host lines)

- [ ] **Step 3: Commit**

```bash
git add .env.example && git commit -m "docs: add distributed build host configuration to .env.example"
```

---

## Task 7: Shell Integration — --distributed Flag Passthrough

**Files:**
- Modify: `scripts/release-build.sh`
- Modify: `Build/lib/orchestrator.sh`

This task adds a `--distributed` flag that delegates to the Go binary instead of the local shell pipeline.

- [ ] **Step 1: Add --distributed flag to orchestrator argument parser**

In `Build/lib/orchestrator.sh`, add a new variable `BUILD_DISTRIBUTED` and parse the `--distributed` flag in `parse_build_args()`. Find the existing flag parsing section (the `case` block in `parse_build_args()`) and add:

```bash
--distributed)
    BUILD_DISTRIBUTED=true
    shift
    ;;
```

Also add `local BUILD_DISTRIBUTED=false` initialization alongside the other `local` variable declarations in `parse_build_args()`.

- [ ] **Step 2: Add distributed build dispatch in build_main()**

In `Build/lib/orchestrator.sh`, in the `build_main()` function, add a distributed gate **before** the existing container gate. Find the line `if [[ "$BUILD_USE_CONTAINER" == "true" ]] && ! is_container; then` and insert the following block before it:

```bash
if [[ "$BUILD_DISTRIBUTED" == "true" ]]; then
    log_info "Distributed build mode enabled"
    local dist_bin="$BUILD_PROJECT_ROOT/Containers/cmd/distributed-build"
    if ! command -v go &>/dev/null; then
        log_error "Go is required for distributed builds"
        return 1
    fi
    go run "$dist_bin" \
        --project "$BUILD_PROJECT_ROOT" \
        ${BUILD_SKIP_TESTS:+--skip-tests} \
        ${BUILD_SINGLE_COMPONENT:+--component "$BUILD_SINGLE_COMPONENT"} \
        ${BUILD_FORCE:+--force} \
        ${BUILD_DRY_RUN:+--dry-run}
    return $?
fi
```

- [ ] **Step 3: Run existing tests to verify no regression**

Run: `cd catalog-api && go test ./... -count=1 -timeout 60s 2>&1 | tail -5`
Expected: No new failures (this is a purely additive change that doesn't affect existing flows)

- [ ] **Step 4: Commit**

```bash
git add Build/lib/orchestrator.sh && git commit -m "feat(build): add --distributed flag that delegates to Go distributed-build binary"
```

---

## Task 8: Full Integration Test — Local Dry Run

**Files:** None (verification only)

- [ ] **Step 1: Build the distributed-build binary**

Run: `cd Containers && go build -o /tmp/distributed-build ./cmd/distributed-build/`
Expected: Binary compiles successfully

- [ ] **Step 2: Run dry-run mode with no hosts configured**

Run: `cd /run/media/milosvasic/DATA4TB/Projects/Catalogizer && /tmp/distributed-build --project . --dry-run 2>&1`
Expected: Error message about no remote hosts configured (expected — no env vars set)

- [ ] **Step 3: Run dry-run mode with mock host via env vars**

Run: `CONTAINERS_REMOTE_HOST_1_NAME=thinker CONTAINERS_REMOTE_HOST_1_ADDRESS=thinker.local CONTAINERS_REMOTE_HOST_1_USER=milosvasic CONTAINERS_REMOTE_HOST_1_RUNTIME=podman /tmp/distributed-build --project . --dry-run 2>&1`
Expected: Prints build plan showing component assignments (may show all as "local" if host is unreachable, which is correct behavior)

- [ ] **Step 4: Run all unit tests**

Run: `cd Containers && GOMAXPROCS=3 go test ./internal/buildpkg/ -v -count=1`
Expected: ALL tests PASS

- [ ] **Step 5: Commit any fixes if needed**

If any fixes were required during integration testing, commit them:
```bash
cd Containers && git add -A && git commit -m "fix(build): integration test fixes for distributed build system"
```

---

## Task 9: Remote Host Prerequisites Script

**Files:**
- Create: `scripts/prepare-build-host.sh`

This script prepares a remote host for distributed builds by checking prerequisites and optionally installing missing tooling.

- [ ] **Step 1: Write the script**

```bash
#!/usr/bin/env bash
set -euo pipefail

HOST="${1:?Usage: $0 <user@host>}"
shift

REMOTE_SCRIPT=$(cat <<'REMOTE_EOF'
set -e

echo "=== Checking build host prerequisites ==="

errors=0

check_cmd() {
    if command -v "$1" &>/dev/null; then
        echo "  [OK] $1 found: $(command -v "$1")"
    else
        echo "  [MISSING] $1 not found"
        errors=$((errors + 1))
    fi
}

check_cmd podman
check_cmd go
check_cmd node
check_cmd npm
check_cmd java
check_cmd rsync

echo ""
echo "=== Checking storage ==="
df -h /tmp | tail -1 | awk '{print "  /tmp: " $4 " available"}'

echo ""
echo "=== Checking memory ==="
free -h | grep Mem | awk '{print "  RAM: " $2 " total, " $4 " available"}'

echo ""
echo "=== Checking CPU ==="
nproc | awk '{print "  Cores: " $1}'

if [ "$errors" -gt 0 ]; then
    echo ""
    echo "ERROR: $errors prerequisite(s) missing"
    exit 1
fi

echo ""
echo "All prerequisites met."
REMOTE_EOF
)

echo "Checking host: $HOST"
ssh "$HOST" "$REMOTE_SCRIPT"
```

- [ ] **Step 2: Make executable**

Run: `chmod +x scripts/prepare-build-host.sh`

- [ ] **Step 3: Test against local machine**

Run: `./scripts/prepare-build-host.sh milosvasic@localhost`
Expected: Prints prerequisite check results (may show some missing tools, which is fine)

- [ ] **Step 4: Commit**

```bash
git add scripts/prepare-build-host.sh && git commit -m "feat(build): add remote host prerequisites checker script"
```

---

## Task 10: Usage Documentation

**Files:**
- Create: `docs/guides/DISTRIBUTED_BUILD.md`

- [ ] **Step 1: Write the documentation**

```markdown
# Distributed Build Guide

## Overview

The distributed build system extends Catalogizer's build pipeline to distribute component builds across multiple hosts based on real-time resource availability. It uses the Containers submodule's scheduler, SSH executor, and remote execution infrastructure.

## Architecture

```
release-build.sh --distributed
       |
       v
Go binary (Containers/cmd/distributed-build)
  1. Load host config from .env or env vars
  2. Probe all hosts for CPU/Memory/Disk usage
  3. Schedule components via resource-aware algorithm
  4. Sync source code to remote hosts via rsync/SCP
  5. Launch builder containers on each assigned host
  6. Collect build artifacts back to local releases/
```

## Host Requirements

Each build host must have:
- Passwordless SSH access from the build orchestrator
- Podman or Docker installed
- Go 1.25+, Node.js 18+, JDK 21 (for Android builds)
- At least 4GB free disk space in `/tmp`
- Network access to the project directory (for source sync)

## Configuration

### Environment Variables

Add to `.env` or set as environment variables:

```env
CONTAINERS_REMOTE_HOST_1_NAME=thinker
CONTAINERS_REMOTE_HOST_1_ADDRESS=thinker.local
CONTAINERS_REMOTE_HOST_1_USER=milosvasic
CONTAINERS_REMOTE_HOST_1_KEY_PATH=~/.ssh/id_ed25519
CONTAINERS_REMOTE_HOST_1_RUNTIME=podman
CONTAINERS_REMOTE_HOST_1_LABELS=go=true,npm=true,jdk=true,rust=true
```

Hosts are numbered 1-100. Discovery stops at the first gap.

### Labels

Labels control which hosts can build which components:
- `go=true` — Required for catalog-api
- `npm=true` — Required for catalog-web, desktop, installer, API client
- `jdk=true` — Required for Android and Android TV
- `rust=true` — Required for desktop (Tauri) and installer

### Scheduling Strategies

| Strategy | Description |
|----------|-------------|
| `resource_aware` (default) | Picks host with best CPU/Memory score |
| `round_robin` | Rotates through hosts evenly |
| `spread` | Picks host with fewest active builds |
| `bin_pack` | Fills most-used host first |

## Usage

### Check host prerequisites

```bash
./scripts/prepare-build-host.sh milosvasic@thinker.local
```

### Dry run (show plan only)

```bash
./scripts/release-build.sh --distributed --dry-run
```

### Build all components distributed

```bash
./scripts/release-build.sh --distributed --force --skip-tests
```

### Build single component on best available host

```bash
./scripts/release-build.sh --distributed --component catalog-api --force
```

### Use specific strategy

```bash
./scripts/release-build.sh --distributed --strategy spread --force
```

## Resource Allocation

The scheduler respects a 30-40% resource limit per host:

| Component | CPU | Memory | Disk |
|-----------|-----|--------|------|
| catalog-api | 2 cores | 2 GB | 1 GB |
| catalog-web | 1 core | 1 GB | 512 MB |
| catalogizer-android | 3 cores | 4 GB | 2 GB |
| catalogizer-androidtv | 3 cores | 4 GB | 2 GB |
| catalogizer-desktop | 3 cores | 4 GB | 2 GB |
| installer-wizard | 3 cores | 4 GB | 2 GB |
| catalogizer-api-client | 1 core | 1 GB | 512 MB |

## Troubleshooting

### Host unreachable

Check SSH connectivity:
```bash
ssh thinker.local echo ok
```

### Build timeout

Increase with `--timeout` (minutes):
```bash
./scripts/release-build.sh --distributed --timeout 60 --force
```

### Source sync failures

Ensure rsync is installed on both hosts:
```bash
rsync --version
```
```

- [ ] **Step 2: Commit**

```bash
git add docs/guides/DISTRIBUTED_BUILD.md && git commit -m "docs: add distributed build usage guide"
```
