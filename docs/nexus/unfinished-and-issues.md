---
title: Helix Nexus — Unfinished Items & Open Issues
date: 2026-04-17 (updated)
status: honest-self-audit
---

# Unfinished Items & Open Issues

Honest audit of the Helix Nexus + cover-quality work. Updated after the
"close every code-level gap" pass.

## Legend

- 🔴 **Blocker** — the feature cannot be used in production until this
  is fixed.
- 🟠 **Significant** — the feature partially works but has a clear gap.
- 🟡 **Minor / hygiene** — nice to close but not blocking.
- ✅ **Closed** — the previous entry is resolved.

---

## Cover-Image Quality Gate

- ✅ **All 14 `CH-IQ-*` challenges registered**
  (`CH-IQ-001..014` across image_quality.go + image_quality_extended.go
  + image_quality_final.go).
- 🟠 **No dedicated Fanart.tv / IGDB / Cover Art Archive provider
  resolvers**. Still relying on the existing
  ExternalMetadataResolver / LocalScanResolver / CachedFileResolver.
- ✅ **`pkg/quality` now has 30+ tests** covering PNG / JPEG / GIF /
  BMP, polyglot content, decompression-bomb guard
  (9000x9000 → FailTooLarge), boundary ±1 px, bad-aspect-sharp,
  partial-JPEG, empty bytes, garbage.
- 🔴 **Full container QA campaign never executed**. No
  `./scripts/release-build.sh --container --force`, no
  `./scripts/services-up.sh`, no `./scripts/helixqa-orchestrator.sh`
  has run since the gate landed.
- 🟡 Admin endpoint to trigger QualityRevalidator manually still
  missing.
- 🟡 Web / Android / desktop clients do not yet surface
  X-Cover-Quality headers in debug UX.

## Nexus Phase 1 — Browser engine

- ✅ **Real `chromedp` and `rod` drivers implement `ExtendedHandle`**
  (hover, drag, select, wait_for, tab_open, tab_close, pdf,
  console_read).
- ✅ **`InstrumentedEngine` emits spans + NexusMetrics**, so the
  Grafana dashboard now has a producer.
- 🟠 **No real-Chromium integration harness**. Drivers compile under
  `nexus_chromedp` / `nexus_rod` but there is no tag-gated test that
  actually boots a browser.
- 🟡 HTML entity handling in `snapshot.go` (`&amp;` comes out literal).
- 🟡 Case-sensitive host allowlist (documented).

## Nexus Phase 2 — Mobile engine

- 🔴 **No integration harness against a real Appium hub.** All tests
  drive `httptest.Server`. Parameter names are faithful to the Appium
  docs but unverified live.
- 🟠 iOS real-device WDA flow unverified.
- 🟡 `Engine.Navigate` does not branch `mobile: deepLink` vs
  `mobile: openUrl` per platform.
- 🟡 `StopRecording` returns nil,nil on empty payload — behaviour may
  need an explicit error.

## Nexus Phase 3 — Desktop engine

- ✅ **`atspi-find`, `atspi-action`, `atspi-type` scripts shipped**
  under `tools/atspi-helpers/` with Wayland (wtype) and X11 (xdotool)
  fallbacks + test-only ATSPI_FAKE_HANDLE_FILE / ATSPI_FAKE_OK envs.
- 🔴 **No integration test against a real WinAppDriver** or real macOS.
- 🟠 macOS screenshot pipeline untested.
- 🟡 macOS Close has no force-quit fallback.
- 🟡 Installer-flow banks describe the shape but carry no golden
  fixtures.

## Nexus Phase 4 — AI layer

- ✅ **`HTTPLLMClient` implemented** — OpenAI-compatible Chat adapter
  with multi-modal image attachments, JSON-response format, API-key
  auth, cost/token tracking. Callers can now point Navigator / Healer /
  Generator at LLMOrchestrator or any compatible endpoint.
- 🟠 Predictor weights are hand-tuned. "AUC > 0.75" success criterion
  unverified because no historical sample set has been imported yet.
- 🟠 Generator YAML validator is minimal (checks id/name/steps only).
- ✅ CostTracker + AuditLog persistable via AuditPersister /
  FlowPersister against the shipped SQL schemas.

## Nexus Phase 5 — a11y / perf / orchestrator / observability

- ✅ **axe-core audit runs via `a11y.Auditor`**, which accepts any
  browser `Evaluator` and invokes the InjectionScript. 7 tests.
- ✅ **SSO shipped** — OIDC + SAML providers, group→role mapping,
  AuthMiddleware with context user injection. 12 tests.
- ✅ **S3 / MinIO evidence backend** via `S3EvidenceStore` + narrow
  `S3Client` interface.
- ✅ **Grafana dashboard has a real producer** —
  `observability.DefaultMetrics` registers every `helix_nexus_*`
  metric the dashboard references; InstrumentedEngine emits them.
- ✅ **OpenTelemetry-shaped Tracer + Instrument** wired into the
  InstrumentedEngine. Real OTel exporter is still an adapter the
  operator writes (no code required; Tracer interface is ready).
- ✅ **Cross-flow metadata persistence** via `FlowPersister`
  (`helixqa_cross_flows` / `helixqa_flow_steps`).
- ✅ **Audit log persistence** via `AuditPersister` against
  `helixqa_audit_log`.
- 🟠 `pkg/nexus/perf/k6.go` still generates scripts only; callers run
  `k6` themselves.

## Runtime wiring

- 🔴 **Nexus is library-complete but not reached from the `helixqa`
  binary or `helixqa-orchestrator.sh`**. The packages compile and
  tests pass; someone still has to register the new adapters inside
  the main CLI.
- 🔴 **Catalog-api continues to use its pre-Nexus Playwright path.**
  The cover-quality gate is shipped and live; the Nexus browser /
  mobile / desktop stack is not reached from `catalog-api/challenges`.
- 🟡 `NexusBrowserAdapter` satisfies the userflow BrowserAdapter
  contract but the challenges registry (in
  `digital.vasic.challenges/pkg/userflow`) has no factory hook yet —
  banks must construct adapters themselves.

## Content & delivery

- 🟡 Video module MP4 recordings (scripts shipped; filming outstanding).
- 🟡 Public `helixqa.vasic.digital/nexus` DNS / deployment.

---

## What should happen next

1. **One real QA campaign** against live catalog-api + web client.
   Every remaining 🔴 is about "has this touched real hardware?"
   rather than "is the code missing?"
2. **Wire InstrumentedEngine into a concrete catalog-api challenge**
   so the Grafana dashboard populates end-to-end.
3. **Ship a real LLMOrchestrator adapter** implementing
   `ai.LLMClient` on top of the existing submodule, so the default
   deployment does not require an operator to hand-write the adapter.
4. **Add Fanart.tv / IGDB / Cover Art Archive provider resolvers**
   behind the existing gate — additive, no gate changes needed.

Everything else is either blocked on real hardware (integration
harnesses, video MP4s, public DNS) or an additive polish item.
