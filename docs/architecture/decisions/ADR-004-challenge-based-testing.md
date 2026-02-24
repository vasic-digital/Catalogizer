# ADR-004: Challenge-Based Testing Framework

## Status
Accepted (2026-02-23)

## Context

Catalogizer is a complex system that integrates multiple protocols (SMB, FTP, NFS, WebDAV, local), a dual-dialect database, real-time scanning, media entity aggregation, asset management, and a multi-platform frontend. Traditional unit tests verify individual functions in isolation, but they cannot validate:

- **End-to-end workflows**: A scan that discovers files on a NAS, creates database records, triggers entity aggregation, resolves cover art, and makes entities browsable via the API
- **Infrastructure prerequisites**: Database connectivity, storage root accessibility, protocol-specific authentication
- **Cross-component interactions**: The scanner writing to the database, the aggregation service creating entities, the asset manager resolving cover art, and the frontend querying all of it
- **Zero-warning compliance**: Every API endpoint must exist, return valid responses, and match expected shapes. No 404s, no console errors, no failed network requests.

Integration test frameworks (like Go's `testing` package with test servers) get closer but lack:
- A structured way to define dependencies between tests (scan must complete before browsing)
- Progress reporting for long-running operations (NAS scans can take 25+ minutes)
- A REST API interface for triggering tests from the frontend or CI/CD tooling
- Persistent result storage for historical analysis

## Decision

We implement a custom challenge framework as a Go submodule (`digital.vasic.challenges`) that defines structured test scenarios called "challenges." Challenges are Go structs that embed `challenge.BaseChallenge` and implement a custom `Execute()` method containing assertions. They are registered at startup, exposed via REST API, and can be run individually or as a full suite.

### Architecture

```
Challenges/ submodule (digital.vasic.challenges)
    |
    | Provides: BaseChallenge, Runner, Config, Result, ProgressReporter
    v
catalog-api/challenges/register.go
    |
    | RegisterAll() registers 20 challenges with the ChallengeService
    v
catalog-api/services/challenge_service.go
    |
    | Manages registration, execution, and result storage
    v
catalog-api/handlers/challenge.go
    |
    | Exposes REST API: /api/v1/challenges/*
    v
Clients (curl, frontend, test scripts)
```

### Challenge Definitions (CH-001 through CH-020)

| ID | Name | Category | Description |
|----|------|----------|-------------|
| CH-001 | first-catalog-smb-connect | infrastructure | Verify SMB/NAS connectivity |
| CH-002 | first-catalog-dir-discovery | infrastructure | Discover directories on storage root |
| CH-003 | first-catalog-populate | infrastructure | Full NAS scan with progress reporting |
| CH-004 | browsing-api-catalog | browsing | Verify catalog browsing API endpoints |
| CH-005 | browsing-web-app | browsing | Verify frontend serves with zero errors |
| CH-006 | asset-serving | assets | Verify asset serving pipeline |
| CH-007 | asset-lazy-loading | assets | Verify lazy asset resolution |
| CH-008 | auth-token-refresh | auth | Verify JWT auth and token refresh flow |
| CH-009 through CH-015 | various | scan/browse | Storage roots, scan operations, search, collections |
| CH-016 through CH-020 | various | entities | Entity aggregation, hierarchy, metadata, duplicates, user metadata |

### Execution Model

- **`RunAll`** is synchronous and blocking. When triggered, all 20 challenges execute sequentially in dependency order. No other challenge can run until it finishes. For a full NAS scan, this can take 25+ minutes.
- **Progress-based liveness detection**: A 5-minute stale threshold monitors each challenge. If no progress is reported within 5 minutes, the challenge is killed with status `stuck`.
- **`challenge.NewConfig()`** sets a default timeout of 5 minutes per challenge. Challenges that need longer (like the populate challenge) must zero out the timeout to inherit the runner's timeout.
- **ProgressReporter**: Challenges embedding `BaseChallenge` automatically receive a `ProgressReporter` from the runner, which they use to report progress at regular intervals (e.g., every 5 seconds during a scan poll loop).

### REST API

```
GET    /api/v1/challenges              List all challenges
GET    /api/v1/challenges/:id          Get challenge details
POST   /api/v1/challenges/:id/run      Run single challenge
POST   /api/v1/challenges/run          Run all challenges (blocking)
POST   /api/v1/challenges/run/category/:category  Run by category
GET    /api/v1/challenges/results      Get stored results
```

### Critical Constraint

All challenge operations are executed exclusively by system deliverables (compiled binaries). The catalog-api service and other Catalogizer applications are the only authorized executors. Custom scripts, curl commands, or third-party tools must never trigger API endpoints within challenge execution. Scanning, storage root creation, and all other operations go through the running services, exactly as an end user would.

## Consequences

### Positive

- **End-to-end validation**: Challenges verify the complete workflow from NAS connectivity through entity browsing, catching integration issues that unit tests miss.
- **Self-documenting**: Challenge definitions serve as executable specifications of what the system must do.
- **REST-accessible**: Challenges can be triggered from the frontend, CLI, or CI/CD pipelines through the same API.
- **Dependency ordering**: Challenges declare dependencies (e.g., `populate` depends on `smb-connect`), ensuring they run in a valid order.
- **Progress visibility**: Long-running operations report progress, preventing false timeout kills and providing user-facing feedback.
- **Zero-warning enforcement**: The challenge suite (especially `browsing-web-app`) validates that every API endpoint exists and returns valid responses, enforcing the zero-warning policy.
- **Historical results**: Results are stored and accessible via the API for trend analysis and regression detection.

### Negative

- **Long execution time**: Running all 20 challenges with a full NAS scan takes 25+ minutes, making it unsuitable for rapid iteration. Individual challenge runs mitigate this.
- **Blocking execution**: `RunAll` is synchronous and blocking, preventing any other challenge from running during execution. This is by design (to prevent resource conflicts during scanning) but limits parallelism.
- **External dependency**: Infrastructure challenges (CH-001, CH-002, CH-003) require a reachable NAS/storage system. They will fail in environments without network storage access.
- **Custom framework**: Using a custom challenge framework instead of a standard Go testing framework means contributors must learn the challenge API. The trade-off is acceptable because standard testing frameworks lack the REST API, progress reporting, and dependency management features required.
