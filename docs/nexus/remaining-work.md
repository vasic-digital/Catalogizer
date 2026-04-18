---
title: Helix Nexus + Cover-Quality — Remaining Work & Known Issues
date: 2026-04-18
status: living document (update after every closure)
supersedes: docs/nexus/unfinished-and-issues.md (which is now summarised here)
---

# Remaining Work & Known Issues

This is the single authoritative list of what is still open across the
Helix Nexus + cover-quality program after the current session. Update
this file in the same commit that closes any item so the list stays
honest.

Severity legend:

- 🔴 **Blocker** — cannot ship the feature without closing.
- 🟠 **Significant** — library works but is not reachable.
- 🟡 **Polish** — nice to have; does not gate the release.
- ⚫ **External** — depends on hardware, a human, or a product decision.

---

## Status snapshot (2026-04-18)

| Tracker | Items | Done | Open |
|---|---|---|---|
| Wiring gaps (W1–W10) | 10 | **10** | 0 |
| Suspected bugs (B1–B10) | 10 | **10** | 0 |
| Polish (P1–P10) | 10 | **8** | 2 (env-gated: P7 front-end tests, P10 real-MinIO container test) |
| OpenClawing2 plan phases | 8 | **8 code + infra** | 1 (live section-9 campaign, env-gated) |
| External / human-gated (E1–E7) | 7 | 0 | 7 |

Every code-level gate is closed. The remaining open items fall into
two categories only: front-end test infrastructure (P7), container
integration harness (P10), and environment-gated operational tiers
(E1–E7 + the two-consecutive-green section-9 gate).

---

## 🟡 Remaining polish items

- **P7** — `useCoverQuality` hook has no Vitest coverage;
  `CoverQualityBadge` has no Storybook story. *Owner:* `catalog-web`
  PR — needs Vitest + Storybook infra spin-up.
- **P10** — MinIO test never spins up a real MinIO container;
  `PutObject` + `ListObjects` happy paths remain mock-only. *Owner:*
  Phase 8 campaign task — requires a podman-compose MinIO service
  under `docker-compose.qa.yml`.

Neither blocks HelixQA v3.0's library-level deliverables; both are
flagged so the quarterly refresh picks them up.

---

## ⚫ External / human-gated

- **E1** — Live two-consecutive-green section-9 production-readiness
  campaign with the full HelixQA services stack, Grafana panels
  populated, zero critical alerts — see
  `docs/plans/2026-04-18-openclawing2-integration-plan.md` §8. Dry-run
  via `scripts/openclaw-full-campaign.sh --skip-benchmarks` is already
  green against the unit + integration tiers.
- **E2** — Real WinAppDriver harness (Windows Pro or YWinAppDriver) +
  real macOS harness (XCUITest on a Mac runner).
- **E3** — Video course MP4s 01–08 (shot lists, VO scripts, exercises
  all shipped; filming + editing outstanding).
- **E4** — Public `helixqa.vasic.digital/nexus` DNS + VitePress deploy.
- ~~**E5**~~ — **CLOSED 2026-04-17** — Android (`ui/debug/CoverQualityBadge.kt`),
  Android TV (`ui/debug/CoverQualityBadge.kt` + MediaCard overlay),
  and Tauri desktop (`hooks/useCoverQuality.ts` + `components/debug/
  CoverQualityBadge.tsx`) all surface the X-Cover-Quality / X-Cover-
  Source debug pill. Release builds pay zero network cost (gated on
  BuildConfig.DEBUG / import.meta.env.DEV). 7 Vitest cases green.
- **E6** — Predictor training against a real historical flake dataset
  (to verify the "AUC > 0.75" success criterion end-to-end).
- **E7** — Fanart.tv / IGDB / Twitch credentials acquired + rotated
  via the operator vault.

---

## Closed — code-level tracker

### Wiring gaps (W1–W10) — all closed

| # | Where | Evidence |
|---|---|---|
| W1 | `catalog-api/main.go:882` mounts `SecurityHeaders()` before route registration | `FIX-OBS-001` in `banks/fixes-validation-obs.yaml` |
| W2 | `/api/v1/admin` group wrapped with `middleware.NewCSRF` | `TestCSRF_W2_AdminGroupRejectsCrossOriginMutations` |
| W3 | Every provider-originated string sanitized through `SanitizeMetadataString` | `TestQualityGate_W3_SanitizesProviderStringsInPersist` |
| W4 | `router.GET("/metrics", promhttp.Handler())` at `catalog-api/main.go:897` | `FIX-OBS-002` |
| W5 | `browser.NewInstrumentedEngine` factory guarantees metrics wiring | `TestNewInstrumentedEngine_W5_RecommendedRuntimeFactory` |
| W6 | `OrchestratorClient` implements `LLMClient`, used by Navigator / Healer / Generator | verified at import site |
| W7 | `AuditPersister.AsSink` + `AccessControl.SetSink` bridge RBAC to SQL | 7 tests in `rbac_sink_test.go` |
| W8 | `K6Runner.RunScenario` driven end-to-end via `NX-PERF-011..014` | `banks/nexus-perf.yaml` |
| W9 | `Auditor.RunAndAssert` driven by `NX-A11Y-011..013` | `banks/nexus-a11y.yaml` |
| W10 | `BuildNexusAdapterStack` exposed from `cmd/helixqa/nexus_adapters.go` | 3 regression tests |

### Bugs (B1–B10) — all closed

| # | Fix | Test |
|---|---|---|
| B1 | Fanart.tv `CanResolve` requires `imdb_id` for movies + `tvdb_id` for TV | `TestFanartTVResolver_B1_RequiresIMDBForMovies` |
| B2 | IGDB v4 client sends `Client-ID` + `Authorization: Bearer` headers | `TestIGDBResolver_B2_HeadersSentOnV4` |
| B3 | chromedp `Hover` dispatches real mouse events via JS | covered in `FIX-BROWSER-001` |
| B4 | rod `SavePDF` uses `io.ReadAll` | covered in `FIX-BROWSER-002` |
| B5 | CSRF `WithCSRFInsecureDev` toggle drops `__Host-` + clears Secure on dev | `TestCSRF_B5_InsecureDevDropsHostPrefixAndSecureFlag` |
| B6 | SAML verifier rejects empty `possibleRequestIDs` slice + empty-string entries | `TestSAMLVerifierFromCrewjam_B6_RejectsEmptyRequestIDSlice` |
| B7 | OTel `semconv/v1.26.0` compile-time pin assertions | `var _ = semconv.SchemaURL` block in `otel_tracer.go` |
| B8 | `Element.Selector` decoded through `decodeHTMLEntities` | `TestSnapshot_B8_DecodesSelectorHTMLEntities` |
| B9 | Linux `Click(Element{})` refuses with actionable error | `TestLinuxEngine_B9_RefusesEmptyElement` |
| B10 | `sortAsc` switches to `sort.Slice` above 500 samples | `TestPredictor_AUC_B10_LargeHoldoutSwitchesToSortSlice` |

### Polish (P1–P10) — 8 of 10 closed

| # | Closure | Test |
|---|---|---|
| P1 | `RateLimiter(6/min)` on `/admin/image-quality/revalidate` | inline in `catalog-api/main.go` |
| P2 | `CircuitBreakerResolver` decorator (trip + cooldown + reset) | 3 regression tests |
| P3 | `Predictor.SaveWeights` / `LoadWeights` / `NewPredictorFromFile` with atomic write + schema version | 5 regression tests |
| P4 | `FileEvidenceStore.Sweep(RetentionPolicy)` with MaxAge + MaxItems + MaxBytes | 3 regression tests |
| P5 | `observability.Handler()` registers `collectors.NewGoCollector` + `NewProcessCollector` by default | registry-test coverage |
| P6 | CSRF test uses stdlib `Cookie.Secure` / `HttpOnly` struct fields | `TestCSRF_B5_InsecureDevDropsHostPrefixAndSecureFlag` |
| P8 | Three fuzz targets: `FuzzDecodeHTMLEntities`, `FuzzParseAttrs`, `FuzzContentSecurityPolicy` | 3-second smoke runs green |
| P9 | `testbank.LoadFile` + `LoadDir` reject duplicate test case ids; JSON twins skipped | 3 regression tests |
| **P7** | still **open** — needs Vitest + Storybook setup in `catalog-web` | — |
| **P10** | still **open** — needs real MinIO container in QA compose | — |

### OpenClawing2 phases — 7 shipped, 1 environment-gated

| Phase | Tip | Deliverable |
|---|---|---|
| 1 | `70e9fa5` | OSS vendoring + licence audit + `pkg/opensource` registry |
| 2 | `a6064a8` | `pkg/nexus/agent` four-phase state machine |
| 3 | `552e32e` | MessageManager + context compaction |
| 4 | `ad70867` | RetryWithBackoff + LoopDetector + SelfHealer |
| 5 | `1b1b8d6` | Stagehand primitives (act/extract/observe/agent + PromptCache) |
| 6 | `451d7cb` | Coordinate scaling + Coord* action builders |
| 7 | `fd3f62c` | Rich ticketing (VideoTimestamp / BeforeAfter / LLMReasoning / …) |
| 8 (infra) | current | Cross-package integration test + campaign runner + Grafana dashboard + release notes. **Live run gated to E1.** |

---

## Recommended next actions

1. **Schedule E1** — a dedicated 2 h window with the full services
   stack; run `scripts/openclaw-full-campaign.sh` twice in a row and
   attach the FINAL-REPORT + Grafana panel snapshots to the v3.0
   release evidence.
2. **Close P7 + P10** alongside E1 so the release covers every code-
   level gate and the env-gated tiers reach parity.
3. **Work through E3–E7** in the order that maximises downstream
   value — E6 (real flake dataset) unlocks the Predictor AUC claim;
   E7 (provider credentials) unblocks live Fanart / IGDB runs.

Every item above is scoped + has clear acceptance criteria. Nothing
requires a spike or redesign.

---

## Closed so far (pre-2026-04-18 reference)

- All library-level R-phase deliverables.
- Cover-quality gate + 14 `CH-IQ-*` challenges + 3 provider resolvers.
- 10 Nexus Go packages with 260+ tests.
- 14 HelixQA banks (123 cases) + 10 fixes-validation banks + 8
  OC-* OpenClawing2 phase banks.
- Grafana dashboard JSON, 8 video module outlines, VitePress site
  source, SQL schemas, operator runbooks.
- Submodule decoupling sweep: Discovery, Config, Containers,
  Challenges, HelixQA — project-agnostic across the whole graph.
