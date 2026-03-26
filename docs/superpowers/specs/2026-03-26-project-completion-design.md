# Catalogizer Project Completion - Design Specification

> **Status:** APPROVED - Ready for Implementation Plan
> **Date:** 2026-03-26
> **Scope:** Complete all unfinished work, maximize test coverage, harden security, optimize performance, and finalize all documentation

---

## Executive Summary

The Catalogizer project is approximately 85% complete. This spec defines the remaining 15% — code fixes, security hardening, test coverage expansion, performance optimization, documentation completion, and content updates. All work must be non-breaking, backward-compatible, and verified by existing + new tests.

## Current State (Verified 2026-03-26)

### What's Working
- 23/23 Go modules wired via `replace` directives
- Analytics/Reporting/Favorites services connected to routes in main.go
- Go backend: 0 TODO/FIXME in production code, proper goroutine management
- Frontend: 93% file coverage, 14 lazy-loaded pages, 7 lazy sub-components
- All 22 Go submodules have go.mod, tests, CLAUDE.md, README.md, ARCHITECTURE.md
- Security infrastructure: SonarQube, Snyk, Trivy, GoSec, OWASP Dep Check configured
- k6 load/stress/soak tests implemented
- 239 challenges registered (50 original + 174 userflow + 15 module)
- 8 video course modules with scripts + 10 slide decks

### What Needs Work

**Code-Level (13 items):**
1. 13 metadata providers are stubs (only TMDB is implemented)
2. DetectionEngine.analyzeFileContent has placeholder content analysis
3. Hardcoded `localhost:3006` URL in collectionsApi.ts (lines 145, 328)
4. `any` return type in collectionsApi.ts:79 (getCollectionItems)
5. Android TV debug/release both point to `catalogizer.dev`
6. WebDAV GetQuota() returns placeholder
7. SMB Connect() is stub implementation
8. Lazy/Memory/Recovery submodules wired but NEVER imported in catalog-api
9. 67 services/repos/handlers eagerly initialized (0 lazy loading on backend)
10. No semaphore patterns in catalog-api
11. 3 vitest.config.ts missing from TS submodules
12. Redis client uses library defaults (no explicit pool config)
13. No explicit HTTP client connection pooling beyond SMB

**Test Coverage (6 areas):**
14. ReportingService: 50+ private methods lack dedicated tests
15. AnalyticsService: 20+ private helper functions untested
16. SyncService: 12+ private sync functions partially covered
17. challenge.go handler: no dedicated test file
18. Android TV: 37 test files vs 166 for Android (gap)
19. No spike test scenario in k6

**Documentation (5 items):**
20. 11 docs have TODO/WIP/planning-only content
21. MODULE_9 and MODULE_10 course scripts not created
22. SonarQube scanner execution not documented
23. Semgrep not configured in docker-compose.security.yml
24. Website content needs update for new features

## Architecture Decisions

1. **Stub providers**: Implement using free/keyless APIs where possible (OpenLibrary, MusicBrainz, IGDB). For paid APIs, implement with graceful degradation (return empty + log warning when no API key).
2. **Lazy loading**: Use `digital.vasic.lazy` module for heavyweight services (MediaAnalyzer, SMBDiscovery, ProviderManager). Keep fast services eager.
3. **Semaphores**: Use `digital.vasic.concurrency` semaphore for scan operations, file operations, and provider requests.
4. **Recovery**: Use `digital.vasic.recovery` circuit breaker for all external API calls (providers, SMB, WebDAV, FTP).
5. **Memory**: Use `digital.vasic.memory` leak detection in test mode only (build tag).
6. **Non-breaking**: All changes must pass existing 239 challenges + all existing unit/integration tests.

## Success Criteria

- All existing tests pass (0 regressions)
- Go test coverage >= 90% per package
- Frontend test coverage >= 95% of files
- 0 critical/high Snyk vulnerabilities
- 0 critical SonarQube issues
- All 11 incomplete docs completed
- All 10 course module scripts exist
- Website reflects current feature set
- k6 stress test: p95 < 500ms at 50 concurrent users
- No goroutine leaks under 30-minute soak test
- Semgrep scan: 0 high-severity findings
