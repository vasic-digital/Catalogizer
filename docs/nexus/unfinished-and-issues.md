---
title: Helix Nexus — Unfinished Items & Open Issues
date: 2026-04-17
status: honest-self-audit
---

# Unfinished Items & Open Issues

A deliberately honest audit of the Helix Nexus + cover-quality work
shipped in this session. Every item here is either missing, partial,
or carries a real risk. The list is triaged by severity so the next
session can pick the most impactful gap first.

## Legend

- 🔴 **Blocker** — the feature cannot be used in production until this
  is fixed.
- 🟠 **Significant** — the feature partially works but has a clear gap.
- 🟡 **Minor / hygiene** — nice to close but not blocking.

---

## Cover-Image Quality Gate (earlier in this session)

- 🟠 **Only 5 `CH-IQ-*` challenges registered** (`CH-IQ-PLACEHOLDER-
  FALLBACK`, `BLOCKS-LOW-RES`, `HEADER-ALWAYS-PRESENT`, `PLACEHOLDER-
  SVG-VALID`, `LLM-DISABLED-BY-DEFAULT`). The spec called for 14
  (`CH-IQ-001..014` including per-media-type threshold enforcement
  across all 11 hints, cache-hit-skips-rescore, concurrent dedup,
  revalidation).
- 🟠 **No dedicated Fanart.tv / IGDB / Cover Art Archive provider
  resolvers**. The gate decorator wraps the existing
  `ExternalMetadataResolver` / `LocalScanResolver` / `CachedFileResolver`
  but does not add new provider sources. When the existing providers
  return a low-quality candidate, the chain currently jumps straight
  to the LLM fallback (gated by env var) or the placeholder.
- 🟠 **`pkg/quality` test surface is ~20 cases, not the 60+** the spec
  committed to. Decompression-bomb, polyglot content-type, and
  concurrent-stress at scale are not yet exercised.
- 🔴 **Full container QA campaign never executed**. No
  `./scripts/release-build.sh --container --force`, no
  `./scripts/services-up.sh`, no `./scripts/helixqa-orchestrator.sh`
  has run since the cover-quality work landed. The feature is
  **regression-clean against the unit/integration test suite** but
  has not been exercised end-to-end on real devices or against the
  running Catalogizer web client.
- 🟡 The background `QualityRevalidator` samples 5% of rows every
  7 days, but there is no operator-visible admin endpoint to trigger
  it manually for validation. `CH-NX-CROSS-002` in the design
  references a revalidator trigger that does not exist yet.
- 🟡 `X-Cover-Quality` / `X-Cover-Source` headers are emitted but no
  client (web, Android, Android TV, Tauri desktop) currently reads
  them for UX attribution.

## Nexus Phase 1 — Browser engine

- 🔴 **`chromedp_driver.go` and `rod_driver.go` do not implement
  `ExtendedHandle`.** Calling `Engine.DoExtended(..., Action{Kind:
  "hover"})` against a real chromedp/rod session today returns
  `ErrActionUnsupported`. The mock driver in tests implements it, so
  the test suite is green, but the real drivers need
  `Hover`/`Drag`/`SelectOption`/`WaitFor`/`OpenTab`/`CloseTab`/
  `SavePDF`/`ConsoleMessages` implementations before the extended
  action surface is usable in production.
- 🟠 **Real-browser integration tests missing**. The default build
  avoids Chromium so CI-less hosts work, but there is no explicit
  `nexus_chromedp_integration` tag-gated harness that actually boots
  Chromium, captures a real snapshot, and asserts on `e1..eN` refs.
  `SnapshotFromHTML` is exercised on synthetic fixtures only.
- 🟡 `snapshot.go`'s state machine does not handle HTML entities
  inside attribute values (`&amp;` comes out literally). Minor, but
  some real sites exercise it.
- 🟡 `engine.go` allowlist comparison is case-sensitive on hostnames.
  Documented as intentional but some operators will trip over it.

## Nexus Phase 2 — Mobile engine

- 🔴 **No integration harness against a real Appium hub.** The whole
  package is unit-tested against `httptest.Server`; it has never
  talked to an actual UiAutomator2 or XCUITest driver. Parameter
  names (`mobile: clickGesture`, `mobile: touchAndHold`,
  `mobile: swipeGesture` etc.) are faithful to the Appium docs but
  unverified against a running hub.
- 🟠 **iOS real-device WDA flow unverified.** The runbook is written
  but the `WebDriverAgentURL` reuse path has not been exercised even
  against a simulator.
- 🟡 `Engine.Navigate` calls `mobile: deepLink` which is supported by
  UiAutomator2 but needs `mobile: openUrl` on XCUITest. The current
  dispatch does not branch on Platform for this command.
- 🟡 `StartRecording` returns no video when `mobile: stopScreenRecording`
  emits an empty string; callers need to decide whether that is a
  failure or a silent skip. Currently returns `nil, nil`.

## Nexus Phase 3 — Desktop engine

- 🔴 **The `atspi-find`, `atspi-type`, `atspi-action` helpers do not
  exist.** `pkg/nexus/desktop/linux.go` shells out to these commands;
  if they are not on `PATH` the Linux engine fails with a cryptic
  `exec: "atspi-find": executable file not found in $PATH`. We need
  either a vendored Go implementation using
  `github.com/godbus/dbus/v5`, or a small shell-script package under
  `tools/atspi-helpers/`.
- 🔴 **No integration test against a real WinAppDriver** or real
  macOS. All tests drive the engines through `httptest.Server` or an
  injected command runner.
- 🟠 **macOS screenshot pipeline is untested.** `screencapture -x -t
  png -` is the intended command; the test runner returns empty
  bytes by default so we have not proven the pipe actually produces
  a PNG.
- 🟡 `WindowsEngine.Close` is fine but `MacOSEngine.Close` uses
  `tell application id ... to quit` which silently fails when the
  app ignores quit events. No retry / force-quit path.
- 🟡 Installer-flow scenarios in the banks describe the shape but do
  not include golden fixtures for the wizard.

## Nexus Phase 4 — AI layer

- 🔴 **No real bridge to `LLMProvider` / `LLMOrchestrator`.** The
  `LLMClient` interface is defined but no concrete implementation
  wraps the existing submodules. Production deployments cannot yet
  drive `Navigator.Decide()` without someone writing a Chat adapter.
- 🟠 **Predictor is simplistic.** The logistic weights were hand-set;
  the spec's "AUC > 0.75 on historical set" success criterion is
  unverified because no historical sample set has been imported.
- 🟠 **Generator YAML validator is minimal.** It checks for `id`,
  `name`, `steps` fields but does not validate step shapes
  (action/expected/name strings), so a model hallucinating a broken
  bank could still slip through.
- 🟡 `CostTracker` entries are in-memory only; nothing persists them
  to the `helixqa_ai_decisions` table yet. The schema exists; the
  repository code does not.

## Nexus Phase 5 — a11y / perf / orchestrator / observability

- 🔴 **axe-core audit never actually runs.** `pkg/nexus/a11y` ships
  the parser and assertion but the `InjectionScript` is not yet
  called anywhere. The browser `Engine.Do` has no
  `Kind: "a11y_audit"` action; a caller would have to hand-roll the
  JavaScript evaluation.
- 🔴 **No SSO integration.** The plan's P5-07 listed SAML + OIDC.
  Shipped: `User` struct, `Role` enum, `AccessControl` + `AuditLog`.
  Not shipped: any IdP-facing code. Operators must hand a
  pre-populated `User` to the orchestrator themselves.
- 🟠 **Evidence vault only has a file backend.** The `EvidenceStore`
  interface exists but S3 / MinIO adapters are not written. A single
  implementation claim of "pluggable" is currently aspirational.
- 🟠 **`pkg/nexus/perf/k6.go` never runs k6.** It generates scripts
  and parses JSON output; calling `k6 run` is left to the caller.
- 🟠 **Grafana dashboard metrics are not emitted.** The dashboard
  references `helix_nexus_browser_active_sessions`,
  `helix_nexus_snapshot_duration_ms_bucket`,
  `helix_nexus_ai_cost_cents`, `helix_nexus_a11y_violations_total`,
  `helix_nexus_cwv_lcp_ms_bucket`, `helix_nexus_flow_duration_seconds`,
  `helix_nexus_evidence_bytes`, `helix_nexus_rbac_denials_total`.
  None of these are registered by any Go code yet. The dashboard
  renders but every panel shows "No data".
- 🟠 **OpenTelemetry exporter not wired.** `pkg/nexus/observability`
  ships a Tracer interface and an in-memory recorder; there is no
  OTel gRPC/HTTP exporter, no Jaeger integration. `Instrument` is
  only connected to the in-memory tracer during tests.
- 🟡 `CrossFlow` SQL schema exists but the orchestrator never writes
  to it; you can drive flows but no run metadata is persisted.
- 🟡 `audit.go`'s entries are in-memory only; no persistence to the
  `helixqa_audit_log` SQL schema.

## Runtime wiring

- 🔴 **None of the new Nexus code is wired into the `helixqa` binary
  or the `helixqa-orchestrator.sh` script.** The packages exist and
  tests pass but the actual CLI user interface does not expose them.
  Operators cannot run `helixqa autonomous --platforms web-nexus`
  today.
- 🔴 **Catalog-api continues to use the pre-Nexus Playwright bridge.**
  The cover-quality work did touch `catalog-api` but the Nexus
  layer was built entirely inside HelixQA. Integration back into
  `catalog-api/challenges` has not happened.
- 🟠 **`pkg/nexus/userflow.NexusBrowserAdapter` is not registered
  with the existing `digital.vasic.challenges` adapter registry.**
  Banks referencing `platform: web-nexus-chromedp` would not find
  the adapter today.

## Hooks + environment

- 🟡 The Semgrep post-tool-use hook has been erroring on every write
  in this session because the workstation lacks
  `SEMGREP_APP_TOKEN`. Writes succeed but the scan never runs.
  Running `/semgrep:setup-semgrep-plugin` once unblocks the pipeline.

## Content / delivery items that are intentionally deferred

- 🟡 Video module MP4 recordings (shot lists + VO + exercises are
  written, but literal filming requires a camera and narrator).
- 🟡 Public `helixqa.vasic.digital/nexus` deployment — VitePress
  source lives in the repo; nobody has pointed a DNS record at it
  yet.

---

## Recommended order of operations next session

1. **Wire `NexusBrowserAdapter` into the challenges adapter registry**
   and one Catalogizer bank case so an end-to-end flow actually runs
   through the new stack. (Unblocks every other Nexus gap.)
2. **Implement `ExtendedHandle` in `chromedp_driver.go` + `rod_driver.go`**
   (copy the mock-driver implementations, switch to real CDP calls).
3. **Write the LLMClient adapter** over `LLMProvider`/`LLMOrchestrator`;
   persist `CostTracker.Entries()` + `AuditLog.Entries()` to SQL.
4. **Build the `atspi-helpers` vendor tools** (Go binary or tiny
   shell scripts) so the Linux desktop engine is actually runnable.
5. **Run one real QA campaign** (container build, services up,
   helixqa orchestrator, all banks green) to flush out the
   integration gaps that unit tests cannot catch.
6. **Expose Nexus metrics** through a Prometheus collector so the
   Grafana dashboard shows live data.

Until items 1–5 land, Nexus is "library-complete" but not yet
"runtime-live". The distinction is important for anyone reading the
status tracker and assuming the autonomous orchestrator currently
drives the new stack end-to-end.
