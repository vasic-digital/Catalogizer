---
title: Helix Nexus — Open-Clawed Integration Plan for HelixQA
date: 2026-04-17
status: proposed (approved design; execution to be scheduled across iterations)
owners: [HelixQA core, catalog-api team]
source: docs/research/open-clawed/Open-Clawed.md
---

# Helix Nexus — Open-Clawed Integration Plan for HelixQA

## 0. Document purpose

This plan translates the `Open-Clawed.md` research (OpenClaw decoupling +
HelixQA "Helix Nexus" vision) into an executable program of work. Every
phase enumerates **fine-grained tasks**, **acceptance criteria**, **test
obligations per constitution category**, **challenges** that validate the
task, and the documentation surface that must be updated. No task is
considered "done" without matching coverage across the ten test categories
listed in `CLAUDE.md` (Article V).

The goal, stated plainly: make HelixQA **bleeding-edge enterprise-grade at
controlling and navigating application UIs and UX flows across web, mobile
(Android phone, Android TV, iOS) and desktop (Windows, macOS, Linux),
driven autonomously by LLMs**, with zero-shot test generation, self-healing
selectors, cross-platform orchestration, accessibility and performance
verification, and real-time operator-facing observability.

## 1. Executive summary

1. **Absorb the extractable parts of OpenClaw (browser-tool, CDP snapshot
   engine, role-based refs) into a Go-native layer inside HelixQA** using
   `chromedp` and `go-rod`, instead of shipping a separate Node.js bridge.
2. **Expand mobile** beyond current ADB/Android-TV coverage with an Appium
   WebDriver adapter that handles iOS real devices and modern Android
   gestures.
3. **Add native desktop automation** for Windows (WinAppDriver / UIA),
   macOS (XCUITest + AppleScript / Automator), and a hardened Linux
   AT-SPI/X11/Wayland adapter.
4. **Wire AI navigation** as a first-class core capability: multimodal
   vision model picks the next action from screenshot + accessibility
   tree, and a self-healing layer recovers when selectors drift.
5. **Make accessibility and performance non-optional**: every browser flow
   runs an axe-core pass; every page-load flow produces Core Web Vitals and
   compares against baselines.
6. **Stand up cross-platform orchestration** so one HelixQA test can span
   web → mobile → desktop with shared state.
7. **Deliver 100% coverage across all ten constitution categories** for
   every new module, with real-world challenges registered in
   `digital.vasic.challenges` and a HelixQA bank case for every user-facing
   feature.
8. **Produce documentation, SQL schemas, diagrams, user guides, manuals,
   video-course modules, and a public website section for Helix Nexus**
   alongside the code.

## 2. Current state review

| Area | Today | Evidence |
|---|---|---|
| Browser | Playwright + bridged CLI models | `HelixQA/tools/opensource/stagehand`, `pkg/bridge/` |
| Android phone | ADB + UIA dump | `catalogizer-android` + ADB adapter |
| Android TV | ADB + DPAD + screencap + screenrecord | `catalogizer-androidtv` + full-qa-androidtv bank |
| iOS | none | no adapter exists |
| Windows desktop | none | no adapter exists |
| macOS desktop | none | no adapter exists |
| Linux desktop | X11 grab + Tauri WebDriver for Tauri apps | `catalogizer-desktop` |
| AI navigation | LLMsVerifier phase-specific strategies (Navigation, Analysis, Planning); Gemini 2.0/2.5 Flash, Astica.AI, OpenAI, llama.cpp RPC | `pkg/vision` + `LLMsVerifier` |
| Accessibility | none | not integrated |
| Performance | k6 load scripts (tests/k6/) | already in main repo |
| Cross-platform | each platform has its own orchestrator run | `scripts/helixqa-orchestrator.sh` |
| Observability | markdown + JSON reports per session | `docs/reports/qa-sessions/` |

## 3. Target architecture (Helix Nexus)

```
+--------------------------------------------------------------------------+
|                    Layer 1 — AI orchestration                            |
|  test-generator | visual-analyzer | llm-navigator | predictive-healing   |
+--------------------------------------------------------------------------+
|                Layer 2 — unified automation API (Go)                     |
|     Browser  |  Mobile  |  Desktop  |  API / gRPC / WebSocket            |
+--------------------------------------------------------------------------+
|                  Layer 3 — platform drivers                              |
|  chromedp + go-rod  |  Appium (UiAutomator2 / XCUITest)  |  WinAppDriver  |
|  XCUITest + osascript  |  AT-SPI / Wayland  |  k6 browser  |  axe-core   |
+--------------------------------------------------------------------------+
|         Layer 4 — evidence, observability, compliance                    |
|  Grafana dashboard  |  Jaeger tracing  |  S3/MinIO evidence vault        |
|  WCAG / Section 508 reports  |  OTel exporters  |  SLA / SLO tracking    |
+--------------------------------------------------------------------------+
```

Layer 1 drives Layer 2 through an **adapter-per-platform** interface. Layer
2 is the only boundary existing HelixQA callers need to retarget.
Layer 3 and 4 are replaceable without touching tests.

## 4. Scope guardrails (YAGNI)

- **Not** building a browser from scratch. chromedp and go-rod give us
  enough CDP control.
- **Not** running full Playwright inside Go. For evergreen cross-browser
  work (Firefox, WebKit) we keep a thin Playwright fallback.
- **Not** shipping a cloud offering. The target is on-prem / developer
  workstation.
- **Not** bringing the entire OpenClaw monorepo in. We absorb patterns
  (snapshot refs, AI-friendly errors, browser profile isolation) and
  **rebuild the 725-line `browser-tool.ts` as Go code**, avoiding the
  TypeScript / gRPC bridge.
- **Not** dropping existing HelixQA banks or adapters. They continue to
  pass; the new layer augments.

## 5. Phased execution

The program spans **five phases** covering roughly twelve weeks of focused
implementation effort. Each phase is a standalone deliverable that can
merge to `main`; each task carries its own tests, challenges, and docs.

### Phase 0 — Program kickoff (week 0)

| # | Task | Effort | Acceptance criteria |
|---|---|---|---|
| P0-01 | Create `HelixQA/pkg/nexus/` namespace + `doc.go` | 1h | Package exists, no build breakage |
| P0-02 | Add go.mod dependencies: `chromedp v0.9+`, `go-rod/rod v0.114+`, `PuerkitoBio/goquery`, `sclevine/agouti` | 2h | `go build ./...` green |
| P0-03 | Decide adapter naming convention + interface signature file `pkg/nexus/adapter.go` | 2h | Review sign-off |
| P0-04 | Write Helix Nexus **charter** in `docs/nexus/charter.md` | 2h | Charter approved in PR |
| P0-05 | Scaffold SQL migration namespace for Nexus-specific tables (`migrations_nexus_*.go`) | 2h | Migrations table accepts new entries |
| P0-06 | Add 5 real-world scenario **challenges** that will remain red until Phase 1 completes: `CH-NX-KICKOFF-*` | 4h | Registered, all red, serve as guiding tests |
| P0-07 | Publish **Helix Nexus Vision** page on the HelixQA website (Mermaid diagram + goals) | 2h | Visible on helixqa.vasic.digital/nexus |

**Test coverage (all 10 categories)**: Phase 0 is scaffolding; unit + docs
+ challenges suffice. Stress/security/DDoS/benchmark/E2E remain placeholder
cases that light up in Phase 1.

### Phase 1 — Browser engine (weeks 1–3)

Absorb OpenClaw's deterministic CDP control and role-based snapshot
pattern directly into Go, alongside (not replacing) the current Playwright
adapter.

| # | Task | Effort | Acceptance |
|---|---|---|---|
| P1-01 | `pkg/nexus/browser/chromedp_driver.go` — launcher with CDP flags, hardened (127.0.0.1 only, sandbox on, unique user-data-dir, `disable-features=IsolateOrigins,site-per-process` reviewed). | 1.5d | Unit tests green |
| P1-02 | `pkg/nexus/browser/rod_driver.go` — same surface, go-rod implementation. | 1d | Unit tests green |
| P1-03 | `pkg/nexus/browser/engine.go` — unified `BrowserEngine` interface selecting between chromedp, rod, and existing Playwright fallback via `EngineType`. | 1d | Contract tests cover all three |
| P1-04 | `pkg/nexus/browser/snapshot.go` — ARIA-role snapshot with OpenClaw-style `e1`, `e2` references (stable across reloads when aria-label matches). Use goquery for HTML parse. | 2d | Snapshot diff tests on 50 fixture pages |
| P1-05 | `pkg/nexus/browser/actions.go` — click, type, drag, scroll, hover, select, screenshot, pdf, tabs, console. Each accepts an `e{n}` ref. | 2d | Action tests green |
| P1-06 | `pkg/nexus/browser/errors.go` — `ToAIFriendlyError(err) string` mirroring OpenClaw. | 0.5d | Table-driven error translator tests |
| P1-07 | `pkg/nexus/browser/pool.go` — warm-pool with `Acquire(ctx)` / `Release(b)`, size configurable via `NEXUS_BROWSER_POOL_SIZE`. | 1d | Stress test verifies no leaks |
| P1-08 | Security hardening: URL allowlist, `file://`/`javascript:`/`data:` scheme blocks, response-size caps, header scrubbing, inline-script execution disabled via Content-Security-Policy injection. | 1d | Security tests green; Semgrep clean |
| P1-09 | Register challenges `CH-NX-BROWSER-001` through `CH-NX-BROWSER-020` covering navigation, refresh, tab open/close, snapshot stability, reload self-heal, CDP reconnection, slow-network, offline, memory-leak long soak, 500 concurrent pages via pool. | 2d | All green on CI-less local runs |
| P1-10 | Integration: wire `pkg/nexus/browser.Engine` as a new HelixQA `BrowserAdapter` option. | 1d | `pkg/userflow/browser` exposes `NexusBrowserAdapter` |
| P1-11 | Docs: `docs/nexus/browser.md` with architecture diagram, sequence diagram for snapshot+action loop, migration guide from Playwright adapter. | 1d | Docs merged |
| P1-12 | Video-course module 01 recorded: *"From Playwright to CDP: HelixQA's new browser core"* | 1d | Video published |
| P1-13 | SQL: `helixqa_browser_sessions` table (session_id, engine, started_at, ended_at, pool_slot) + migration. | 0.5d | Migration parity tests green |

#### Phase 1 coverage matrix (all ten categories)

| Category | How satisfied |
|---|---|
| Unit | table-driven tests for every driver method (chromedp, rod, Playwright); 200+ tests |
| Integration | golden-flow run against 20 fixture HTML pages served from an in-process test server |
| E2E | Playwright-style flow running the existing Catalogizer web app against `EngineChromedp` |
| Full automation | `helixqa autonomous --platforms web-nexus` exercises the new engine unattended |
| Stress | 500 concurrent `Acquire/Release` cycles; 4h soak measuring RSS |
| Security | URL allowlist violation, CDP non-loopback blocked, inline-script execution refuser, sandbox flag regression |
| DDoS | rate-limit on `browser_service` gRPC endpoint (if exposed); slowloris reject |
| Benchmark | `BenchmarkSnapshot` < 250 ms/page (target from research doc §1.1) |
| Challenges | `CH-NX-BROWSER-001..020` registered |
| HelixQA | bank cases added to `HelixQA/banks/nexus-browser.yaml` (≥15 cases) |

### Phase 2 — Mobile engine (weeks 4–5)

| # | Task | Effort | Acceptance |
|---|---|---|---|
| P2-01 | `pkg/nexus/mobile/appium.go` — Appium 2.0 WebDriver client in Go (no CGo). Uses agouti or raw HTTP. | 2d | Unit+contract tests green |
| P2-02 | iOS capability profile builder (bundleId, UDID, XCTest, WebDriverAgent derivation). | 1d | Works against simulator & real device lane |
| P2-03 | Android capability profile builder (UiAutomator2). | 0.5d | Works on Mi Box, Pixel emulator |
| P2-04 | Gestures: tap, longPress, swipe (velocity/accel), pinch, rotate, 3DTouch, hwButtons. | 2d | Gesture tests green on both platforms |
| P2-05 | Accessibility hierarchy dump → normalized tree consumable by AI navigator. | 1d | Tree conformance tests for 10 reference apps |
| P2-06 | Screen recording integration (already exists for Android TV; extend to iOS simulator via `xcrun simctl io booted recordVideo`). | 1d | Videos recorded per session |
| P2-07 | iOS real-device lane documented (WDA build instructions, team ID placeholders). | 0.5d | Docs merged |
| P2-08 | SQL: `helixqa_mobile_devices` table (platform, udid, name, os_version, last_seen, availability). | 0.5d | Migration green |
| P2-09 | Real-world scenarios: `CH-NX-MOBILE-*` — login, share sheet, push notification handling, deep link, permission dialog, battery-saver, airplane-mode. 20 cases. | 2d | Registered, executable |
| P2-10 | HelixQA banks: `banks/nexus-mobile-android.yaml` + `nexus-mobile-ios.yaml` (15 cases each). | 1d | YAML + JSON parity tests |
| P2-11 | Docs: `docs/nexus/mobile.md` with Appium setup, device farms (BrowserStack, Sauce Labs optional), SDK compatibility matrix. | 1d | Merged |
| P2-12 | Video-course module 02 recorded: *"Unified Appium from Go"*. | 1d | Published |

Coverage matrix mirrors Phase 1. Stress test = 10 parallel sessions across
devices (bounded by host budget), soak = 2h session on one device.

### Phase 3 — Desktop engine (weeks 6–7)

| # | Task | Effort | Acceptance |
|---|---|---|---|
| P3-01 | `pkg/nexus/desktop/windows.go` — WinAppDriver HTTP client; session, find element by `accessibility id` / `name` / `class name`, click, sendKeys, context-menu, window management. | 2d | Tests via mock WinAppDriver server |
| P3-02 | `pkg/nexus/desktop/macos.go` — XCUITest via WDA + `osascript` for menu bar / global shortcuts / Dock. | 2d | Tests via fake WDA server |
| P3-03 | `pkg/nexus/desktop/linux.go` — AT-SPI over DBus (primary) with X11 grab fallback; Wayland detection. | 2d | Tests on Linux CI-less host |
| P3-04 | Unified `DesktopEngine` interface — same surface as browser / mobile. | 1d | Contract tests |
| P3-05 | Installer / uninstaller flows as first-class scenarios (MSI, DMG, DEB). | 1d | Scenarios green on canonical installers |
| P3-06 | Tray-icon, system notification, global shortcut support. | 1d | Tests cover macOS menu bar + Windows tray |
| P3-07 | SQL: `helixqa_desktop_hosts` table (platform, hostname, role, last_probe, available). | 0.5d | Migration green |
| P3-08 | Challenges `CH-NX-DESKTOP-*` (20 cases covering launch, tray, installer, file-open dialog, print preview, crash-on-launch detection, multi-window). | 2d | Registered |
| P3-09 | HelixQA banks: `banks/nexus-desktop-windows.yaml`, `banks/nexus-desktop-macos.yaml`, `banks/nexus-desktop-linux.yaml` (12 cases each). | 1.5d | YAML+JSON parity |
| P3-10 | Docs: `docs/nexus/desktop.md` covering all three OSes. | 1d | Merged |
| P3-11 | Video-course module 03: *"Desktop UI under HelixQA"*. | 1d | Published |

### Phase 4 — AI navigation + self-healing (weeks 8–9)

| # | Task | Effort | Acceptance |
|---|---|---|---|
| P4-01 | `pkg/nexus/ai/navigator.go` — `DecideNextAction(goal, VisualContext) -> NavigationAction`. Prompts via existing `LLMOrchestrator` / `LLMProvider`. | 2d | Contract tests with mocked LLM |
| P4-02 | `pkg/nexus/ai/healer.go` — `SelfHealSelector(failed, screenshot, description) -> string`. | 1.5d | Contract tests |
| P4-03 | `pkg/nexus/ai/generator.go` — natural-language user-story → test-steps generator. Produces bank YAML. | 2d | Generator emits valid bank YAML for 10 story cards |
| P4-04 | `pkg/nexus/ai/predictor.go` — ML-lite classifier (logistic regression over historical pass/fail data) flags flaky tests for auto-retry. Stored model lives in `models/helixqa-flake-predictor.onnx` or a plain JSON weights file. | 2d | Tests verify AUC > 0.75 on historical set |
| P4-05 | Cost tracking — every LLM call logs provider, tokens, cost estimate; budget cap per session surfaced via environment (`NEXUS_LLM_BUDGET_USD`). | 1d | Budget exceeded → hard abort |
| P4-06 | SQL: `helixqa_ai_decisions` (session_id, step, action, target, reasoning, confidence, model, tokens_in, tokens_out, cost_usd, outcome). | 0.5d | Migration + repo tests green |
| P4-07 | SQL: `helixqa_flake_predictions` (test_id, platform, probability, features_json, predicted_at). | 0.5d | Migration + repo green |
| P4-08 | Challenges `CH-NX-AI-*` (14 cases: goal satisfied under clean conditions; goal satisfied after DOM mutation; self-heal after selector drift; refuses unsafe action; budget enforced; predictor blocks rerun of known-flaky test). | 2d | Registered |
| P4-09 | HelixQA banks: `banks/nexus-ai.yaml` (12 cases). | 1d | YAML+JSON parity |
| P4-10 | Docs: `docs/nexus/ai.md` with prompt templates, guardrails, cost strategy. | 1d | Merged |
| P4-11 | Video-course module 04: *"LLM-led navigation, safely"*. | 1d | Published |

### Phase 5 — Accessibility, performance, cross-platform, enterprise (weeks 10–12)

| # | Task | Effort | Acceptance |
|---|---|---|---|
| P5-01 | `pkg/nexus/a11y/axe.go` — injects axe-core from local vendor copy (no CDN), runs audit, returns typed violations. | 1.5d | Results match axe reference run within tolerance |
| P5-02 | `pkg/nexus/a11y/compliance.go` — WCAG 2.2 A / AA / AAA assertion + Section 508 mapping. | 1d | Assertion tests on golden runs |
| P5-03 | `pkg/nexus/perf/k6.go` — wraps `k6 run` with browser binding; parses output JSON; enforces thresholds (`LCP p95 < 2500`, `INP p95 < 200`, `CLS p95 < 0.1`). | 1d | Script generated from scenario; test verifies thresholds |
| P5-04 | `pkg/nexus/perf/core_web_vitals.go` — in-process CWV collector via chromedp performance domain. | 1d | Metrics match DevTools for reference page |
| P5-05 | `pkg/nexus/orchestrator/cross_platform.go` — shared `ExecutionContext`, typed `Step`, state propagation between web / mobile / desktop. | 2d | Flow passes: register on web → verify in mobile app → confirm on desktop |
| P5-06 | SQL: `helixqa_cross_flows` + `helixqa_flow_steps`. | 0.5d | Migration green |
| P5-07 | Enterprise: SSO (SAML/OIDC) on dashboard; RBAC; audit log; per-team quota. | 2d | Access-control tests green |
| P5-08 | Real-time Grafana dashboard (already partially in place) extended for Nexus metrics. | 1d | Dashboard JSON under `monitoring/grafana/` |
| P5-09 | Evidence vault (S3 / MinIO) — pluggable `EvidenceStore` interface; default file-store retained. | 1d | Contract tests across both backends |
| P5-10 | OpenTelemetry: all adapters emit spans; tracing exporter default jaeger. | 1d | Span graph visible on local jaeger |
| P5-11 | Challenges `CH-NX-A11Y-*`, `CH-NX-PERF-*`, `CH-NX-XFLOW-*`, `CH-NX-OBS-*` (total 40 cases). | 2d | Registered |
| P5-12 | HelixQA banks: `banks/nexus-a11y.yaml`, `banks/nexus-perf.yaml`, `banks/nexus-xflow.yaml` (10 cases each). | 1.5d | YAML + JSON parity |
| P5-13 | Docs: `docs/nexus/a11y.md`, `docs/nexus/perf.md`, `docs/nexus/cross-platform.md`, `docs/nexus/enterprise.md`. | 2d | Merged |
| P5-14 | Video-course modules 05–08 recorded (a11y, perf, cross-platform, enterprise). | 2d | Published |
| P5-15 | Website: helixqa.vasic.digital gets a top-level `/nexus` section with all phase docs linked + interactive demo. | 2d | Live |

## 6. Coverage obligations per task (summary)

Every Pn-m task carries the following checklist. A task merges to main only
when every checkbox passes:

- [ ] **Unit** tests for every public function, table-driven where sensible
- [ ] **Integration** tests against the adapter's native boundary
- [ ] **E2E** flow proving the feature works via a real HelixQA session
- [ ] **Full automation** — the feature is reachable from `helixqa
      autonomous` without human intervention
- [ ] **Stress** — worst-case load test bounded by host resource budget
- [ ] **Security** — threat-model entries closed; Semgrep / Gosec /
      `govulncheck` clean; negative tests for unsafe inputs
- [ ] **DDoS / rate-limit** — any exposed gRPC/HTTP surface rejects floods
      and recovers
- [ ] **Benchmarks** — micro-benchmarks for hot paths with baselines stored
      under `tests/benchmark/baselines/` and regression detector
- [ ] **Challenges** — at least one scenario challenge registered via
      `digital.vasic.challenges` for every user-visible feature
- [ ] **HelixQA bank** — at least one bank case (YAML + auto-generated
      JSON) exercising the feature from a user perspective

The helixqa-orchestrator run is the final gate per phase; it must produce a
green `FINAL-REPORT.md` before the phase is considered done.

## 7. Real-world scenario challenges (design catalogue)

These scenarios drive the challenge suite. Each appears in both
`digital.vasic.challenges` (as a Go challenge) and the matching HelixQA
bank (as a YAML case). Numbering follows `CH-NX-<domain>-<nnn>`.

1. **Shopping checkout on web** (P1) — from product page → add to cart →
   login → address → payment → confirmation, tolerating modal dialogs and
   captcha-like challenges (LLM escalation).
2. **Multi-tab research** (P1) — open N tabs, copy text between them,
   aggregate in final doc.
3. **File upload / download** (P1) — picker, drag-drop, download progress
   verification, file hash comparison.
4. **Subscription upgrade from mobile** (P2) — StoreKit / Google Play
   billing happy path under sandbox accounts.
5. **Deep link from email to mobile app** (P2) — open link on device,
   verify activity, screenshot attribution.
6. **Android TV channel browse + play** (P2) — channel row → details →
   playback start → playback controls.
7. **iOS biometric unlock** (P2) — Face ID / Touch ID simulator toggles.
8. **Windows installer MSI** (P3) — install → first-run → uninstall →
   registry clean.
9. **macOS app notarized bundle** (P3) — launch unquarantined, verify
   gatekeeper assessment.
10. **Linux AppImage Tauri desktop** (P3) — wizard flow across 5 screens
    with system-tray action.
11. **Vision-assisted CAPTCHA edge case** (P4) — LLM refuses rather than
    auto-solves; challenge asserts refusal.
12. **Self-heal after selector drift** (P4) — DOM mutation between runs;
    healer finds new `aria-label` / visual match.
13. **Test generation from user story card** (P4) — paragraph → valid bank
    YAML; runs successfully the first time.
14. **Flake predictor wins** (P4) — historical flaky test auto-retried
    only within budget.
15. **Accessibility blocker** (P5) — runtime component violates WCAG 2.2
    AA; build fails; report lists ARIA role fixes.
16. **Performance regression** (P5) — LCP degrades by 30%; baseline delta
    triggers alert.
17. **Cross-platform payment reconciliation** (P5) — web checkout → mobile
    receipt → desktop sync; shared token state verified.
18. **Enterprise audit trail** (P5) — SSO login; RBAC prevents forbidden
    scenario; audit log recorded.
19. **Chaos: flaky network** (every phase) — 5% packet loss, 200 ms jitter,
    DNS timeout; HelixQA surfaces rather than masks the defect.
20. **Chaos: device death mid-test** (every phase) — emulator crash,
    browser process SIGKILL, mobile reboot; orchestrator cleanly fails the
    affected step, preserves evidence, moves on.

Each scenario gets a minimum of three variants: happy path, unhappy path,
and adversarial (malformed input / hostile user).

## 8. Documentation & content deliverables

Every phase ships documentation alongside code. The full content surface
is:

| Surface | Location | Owner | Format |
|---|---|---|---|
| User guide | `docs/nexus/user-guide/*.md` | HelixQA | Markdown |
| Operator manual | `docs/nexus/operator-manual/*.md` | SRE | Markdown |
| Architecture diagrams | `docs/nexus/architecture/*.mmd` (Mermaid), exported PNG/SVG under `/diagrams/` | Architect | Mermaid |
| Sequence diagrams | `docs/nexus/sequences/*.mmd` | Architect | Mermaid |
| SQL schema reference | `docs/nexus/sql/*.sql` + Markdown cross-reference | DB | SQL + MD |
| API reference | generated from Go docstrings into `docs/nexus/api/` | Build | Markdown |
| Video course | `docs/nexus/video-course/<module>.md` (shot list, VO script, exercise file) + MP4 published | Content | Markdown + MP4 |
| Website | `website/src/nexus/` (VitePress) | Web | Vue / MD |
| CHANGELOG | `HelixQA/CHANGELOG.md` per release | HelixQA | Markdown |
| Runbooks | `docs/nexus/runbooks/*.md` (incident playbooks for LLM outage, iOS device-farm unreachable, Grafana down) | SRE | Markdown |
| Compliance reports | generated artefacts under `docs/reports/compliance/` | Compliance | PDF / Markdown |

Each user-visible feature earns an entry in **all** of: user guide,
operator manual, architecture diagram, API reference, CHANGELOG, website.
Missing an entry is a definition-of-done failure.

## 9. Web-research gaps to close during execution

The research document cites chromedp, go-rod, agouti, Appium, WinAppDriver,
XCUITest, axe-core, k6, GoCV, go-openai. During execution, three
additional items need quick vetting because they are underspecified in the
research:

1. **Wayland accessibility** — AT-SPI is the right bus, but Wayland's
   portal model matters; confirm `libatspi2` + AT-SPI-over-DBus proxy
   works on Fedora 41+ and Ubuntu 24.04. If not, fall back to X11
   ImageSearch.
2. **iOS real-device provisioning without Apple ID exposure** — WDA build
   steps are in the research; what is missing is the team-owned
   provisioning profile rotation. Look at tidbyt and appium/ios-runtime
   docs for battle-tested patterns.
3. **OpenClaw `browser-tool.ts` live source** — the research cites 725
   lines; verify the exact current shape at integration time. If the
   public repo has moved / been archived, pin to a vendored SHA under
   `third_party/openclaw/` (license-permitting) so our Go port is
   reviewable.

Each gap becomes a **Spike** task at the start of its relevant phase (P1,
P2, P3). Spikes are time-boxed to 1 day and produce a short note under
`docs/nexus/spikes/`.

## 10. Risk register

| Risk | Severity | Mitigation |
|---|---|---|
| chromedp/go-rod breaking changes during implementation | medium | Pin versions; run `go mod why` audit weekly |
| LLM provider price spike | medium | Budget cap env var; per-session hard abort; multi-provider fallback order |
| Appium + iOS provisioning hell | high | Spike P2-Spike-02 kills this before deep work |
| Windows / macOS device availability in dev | medium | MacStadium + Azure Windows VMs rented per week via operator manual |
| Flaky tests erode trust in new engine | high | Predictor + mandatory retry budget; red-flag script fails build on net new flake |
| Hook noise (e.g., Semgrep without auth) blocks iteration | low | Document opt-out; ship pre-commit + CI-less local equivalent |
| Scope creep (cloud offering, dashboard SaaS) | medium | Charter + YAGNI section enforced at PR review |

## 11. Success metrics (program-level)

- Snapshot → action round-trip **p50 < 250 ms**, **p95 < 500 ms** on a
  typical Linux workstation (target from Open-Clawed research §1.1).
- **Zero** visibly-blurry covers after cover-quality gate × **zero**
  accessibility violations at WCAG 2.2 AA on every shipped Catalogizer
  screen.
- **≥ 90%** of scenarios reach goal without human keystroke when AI
  navigation is enabled.
- **< 2%** false-positive rate on flake predictor.
- **100%** test categories green per feature before merge.
- Every feature has a user-guide page, operator-manual page, diagram,
  video module, website section, challenge, and bank case — no exception.

## 12. Rollout strategy

- **Phase 0 and Phase 1 ship behind a feature flag** (`NEXUS_ENABLED=true`).
  Existing Playwright adapter remains the default until Phase 1 passes
  soak tests.
- **Phase 2 onwards merges without a feature flag** because the new
  adapters sit alongside existing code and are selected by test-bank
  declaration.
- **Deprecation of the old adapters** only happens after two consecutive
  full QA campaigns pass on Nexus alone (earliest: after Phase 5).

## 13. Commit, push, and review discipline

- Per-task branches: `nexus/<phase>/<task-id>-short-name`.
- Commit style: Conventional Commits. Body always references the task id.
- Every push targets **all** applicable upstreams (HelixQA has 6 remotes;
  main repo has 7). Session script will automate the per-remote push.
- Docs and code land in the **same PR** — a code PR without doc updates
  fails review.
- The Semgrep hook is currently unauthenticated; run `/semgrep:setup-
  semgrep-plugin` once on the workstation so each PR is scanned locally
  before push.

## 14. Open questions for the team

1. Do we vendor OpenClaw's `browser-tool.ts` under `third_party/` (needs
   licence review) or do we port by specification alone?
2. Where does the evidence vault live — local MinIO or a cloud bucket? The
   operator manual needs this answer before Phase 5.
3. Do we pay for device-farm access (BrowserStack, Sauce Labs), or keep
   everything on owned devices + emulators?
4. Is AAA-level accessibility a must, or does AA suffice for all public
   Catalogizer screens? (Affects P5-02 scope.)
5. Who owns the video course recordings — HelixQA team or external
   contractor? Affects scheduling of P1-12, P2-12, P3-11, P4-11, P5-14.

The plan can execute without these answers for the first two phases.
Phases 3+ must resolve them before their kickoff.

---

**This plan is the source of truth for Helix Nexus execution.**

Subsequent implementation sessions should pick off tasks in order,
branching from main, completing the matching tests and docs, pushing to
all upstreams, and updating the status tracker at
`docs/nexus/status.md` as each task closes.
