# Module 10: Advanced Features - Slide Deck Outline

**Total Slides**: 10
**Estimated Duration**: 45 minutes

---

## Slide 1: Title Slide (2 min)

**Title**: Advanced Features

- Challenge system, user flow automation, media entity aggregation pipeline
- Prerequisites: Module 6 completed
- By the end: run challenges, understand user flow automation, trace entity aggregation

---

## Slide 2: Challenge Framework Overview (5 min)

**Title**: Structured Test Scenarios

- Challenges are Go structs embedding challenge.BaseChallenge with custom Execute()
- RegisterAll() in catalog-api/challenges/register.go
- REST endpoints: GET/POST /api/v1/challenges for listing and execution
- 239+ registered challenges: CH-* (original), UF-* (user flow), MOD-* (module verification)
- Demo: list challenges via the API and run one

---

## Slide 3: Challenge Execution Model (5 min)

**Title**: Running and Monitoring Challenges

- RunAll is synchronous/blocking -- no other challenge runs until it finishes
- Progress-based liveness detection: 5-minute stale threshold kills stuck challenges
- challenge.NewConfig() sets Timeout=5min by default, zero it for runner's timeout
- config.json write_timeout must be 900 for long-running RunAll
- Exercise reference: Exercise 10.1 -- run a single challenge and interpret the report

---

## Slide 4: Challenge Bank (4 min)

**Title**: 239+ Challenges Across All Components

- CH-001 to CH-050: original challenges covering API, database, scanning, browsing
- MOD-001 to MOD-015: module verification challenges for each Go submodule
- UF-*: 174 user flow challenges across API, Web, Desktop, Mobile
- Challenge bank definitions in challenges/config/
- Results and reports in challenges/ directory

---

## Slide 5: User Flow Automation Framework (5 min)

**Title**: Multi-Platform Test Automation

- Generic framework in Challenges/pkg/userflow/ with zero project-specific references
- 6 adapter interfaces: Browser, Mobile, Desktop, API, Build, Process
- 9 CLI adapter implementations: Playwright, ADB, Tauri, HTTP, Gradle, npm, Go, Cargo, Process
- 13 challenge templates: Env, Build, UnitTest, Lint, APIHealth, BrowserFlow, MobileFlow, etc.
- 12 evaluators via UserFlowPlugin for automated pass/fail determination

---

## Slide 6: User Flow CLI Runner (4 min)

**Title**: Running User Flow Challenges

- Challenges/cmd/userflow-runner with CLI flags
- --platform: api, web, desktop, mobile
- --report: generate detailed execution report
- --compose: use container test stack (docker-compose.test.yml)
- --root, --timeout, --output, --verbose for customization
- Exercise reference: Exercise 10.2 -- run API user flow challenges

---

## Slide 7: Catalogizer User Flow Challenges (4 min)

**Title**: 174 Platform-Specific User Flow Tests

- userflow_api.go: 49 API challenges (HTTP-based)
- userflow_web.go: 59 Web challenges (Playwright-based)
- userflow_desktop.go: 28 Desktop challenges (Tauri + Wizard)
- userflow_mobile.go: 38 Mobile challenges (Android + Android TV)
- Registered via RegisterUserFlow*Challenges() in register.go

---

## Slide 8: Media Entity Aggregation Pipeline (5 min)

**Title**: From Scanned Files to Structured Entities

- UniversalScanner completes scan, triggers AggregationService.AggregateAfterScan()
- Title parser: regex patterns for movie, TV, music, game, software
- MediaItem creation/update in media_items table
- MediaFile linking via media_files junction table
- Hierarchy builder: TV Show -> Season -> Episode, Artist -> Album -> Song
- Demo: scan a directory and trace entity creation in the database

---

## Slide 9: Entity System Details (5 min)

**Title**: 11 Media Types and Entity Hierarchy

- Media types seeded in media_types table: movie, tv_show, tv_season, tv_episode, music_artist, music_album, song, game, software, book, comic
- parent_id self-reference for hierarchy (media_items table)
- Entity API: /api/v1/entities with type filtering, search, hierarchy browsing
- Entity Browser UI at /browse and /entity/:id
- Duplicate detection: same title + type + year across sources
- Exercise reference: Exercise 10.3 -- verify entity hierarchy after scan

---

## Slide 10: Module Summary (3 min)

**Title**: What We Covered

- Challenge framework: 239+ challenges with progress tracking and liveness detection
- User flow automation: 6 adapters, 9 CLI implementations, 174 Catalogizer challenges
- Media entity pipeline: scan -> parse -> create -> link -> build hierarchy
- 11 media types with parent-child hierarchy and duplicate detection
- Next module: Monitoring and Observability
