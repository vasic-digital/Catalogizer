# Catalogizer Project Constitution
## Core Development & QA Principles

### Article I: Real-Time Log Monitoring (MANDATORY)

**§1.1 All QA Sessions Must Monitor Logs in Real-Time**

During ANY QA session execution (HelixQA or manual testing), real-time log monitoring is **MANDATORY** and **NON-NEGOTIABLE** for:

- **Android/Android TV Apps**: `adb logcat` must be actively monitored
- **Web Applications**: Browser console logs must be captured and monitored
- **Desktop Applications**: Application logs and system logs must be monitored
- **Backend Services**: Service logs, error logs, and access logs must be monitored
- **All Services**: Any running service or daemon must have its logs monitored

**§1.2 Purpose**

Real-time log monitoring ensures immediate detection of:
- Application Not Responding (ANR) errors
- Fatal crashes and exceptions
- Memory leaks and OOM errors
- Network timeouts and connection failures
- Security violations and unauthorized access attempts
- Performance degradation indicators

**§1.3 Implementation Requirements**

1. **HelixQA must capture and stream logs** for all tested platforms simultaneously
2. **Log analysis must happen in real-time**, not post-session
3. **ANR/Crash detection must trigger immediate alerts** and session pausing
4. **All log outputs must be saved** to the session directory for post-analysis
5. **Log monitoring is NOT optional** - no QA session is valid without it

**§1.4 Violation Consequences**

Any QA session conducted without real-time log monitoring is **INVALID** and must be repeated.

---

### Article II: Video Recording (MANDATORY)

**§2.1 All QA Sessions Must Record Video**

Video recording is **MANDATORY** for all UI/UX QA sessions. Screen captures alone are insufficient.

---

### Article III: Crash Fix Protocol

**§3.1 Immediate Action Required**

When ANRs or crashes are detected:
1. **STOP** the QA session immediately
2. **CAPTURE** all relevant logs and stack traces
3. **ANALYZE** the root cause
4. **FIX** the underlying issue in the application code
5. **VERIFY** the fix with a focused regression test
6. **RESUME** full QA only after verification passes

---

### Article IV: Universal Solution Principle (MANDATORY)

**§4.1 All Solutions Must Be Universal**

When implementing fixes, workarounds, or testing infrastructure, **ALL solutions MUST be UNIVERSAL** and work with **ANY application**, not just Catalogizer.

**§4.2 Scope**

This applies to:
- **QA Tools** (HelixQA, test frameworks, automation scripts)
- **Input Methods** (text input, navigation, gestures)
- **Detection Systems** (ANR detection, crash monitoring, performance metrics)
- **Workarounds** (temporary fixes, bypasses, fallbacks)
- **Testing Infrastructure** (video recording, screenshot capture, log monitoring)

**§4.3 Prohibited Practices**

The following are **STRICTLY PROHIBITED**:
1. Adding test-only code to the application under test (e.g., `QAInputReceiver` in app code)
2. Modifying the target application to make it "testable"
3. Creating app-specific testing hooks, endpoints, or bypasses
4. Hardcoding app-specific coordinates, IDs, or behaviors in testing tools

**§4.4 Correct Approach**

The fix MUST be implemented in the **testing tool/infrastructure** itself:
- If text input doesn't work → Fix the input method in HelixQA (e.g., on-screen keyboard navigation)
- If detection fails → Improve detection algorithms in the testing framework
- If navigation breaks → Enhance the navigation engine
- The target application should require **ZERO modifications** for testing

**§4.5 Rationale**

Universal solutions ensure:
- **Portability**: Testing tools work with any Android TV app without modification
- **Maintainability**: No app-specific test code to maintain
- **Validity**: Tests reflect real user interactions
- **Reusability**: Solutions benefit future projects and the broader community

**§4.6 Violation Consequences**

Any solution that modifies the application under test to facilitate testing is **INVALID** and must be reimplemented as a universal solution in the testing infrastructure.

---

### Article V: 100% Test Coverage Across All Categories (MANDATORY)

**§5.1 Required Test Categories**

Every component, service, and application MUST maintain **no less than 100% coverage** across **every** one of the following test categories. None of these categories may be skipped, deferred, or partially covered:

1. **Unit tests** — individual function / class behavior, pure logic
2. **Integration tests** — cross-module interactions, DB, cache, queues, filesystems
3. **End-to-end (E2E) tests** — full user journeys through the live system
4. **Full automation tests** — unattended, reproducible, CI-runnable versions of the E2E suite
5. **Stress tests** — resource saturation, high concurrency, large payloads, long-running sessions
6. **Security tests** — authn/z, input validation, injection (SQL/command/XSS), SSRF, secrets handling, CVE scans (`govulncheck`, `npm audit`, Semgrep, Gosec, Trivy)
7. **DDoS / rate-limit tests** — sustained floods, burst attacks, slowloris, amplification, connection exhaustion, verification that limiters and circuit breakers actually reject and recover
8. **Benchmarking** — latency / throughput / memory baselines with regression detection
9. **Challenges** — the `digital.vasic.challenges` framework must have a green registered challenge for every feature
10. **HelixQA QA testing** — autonomous LLM-driven sessions (Learn → Plan → Execute → Curiosity → Analyze) covering every screen, every flow, every edge case

**§5.2 "100%" Definition**

"100%" in this constitution means **every one** of:
- every public function / API endpoint / UI component has at least one test in each applicable category
- every branch (happy path, each error path, edge cases, adversarial input) is exercised
- every feature has an end-to-end flow test and a registered challenge and a HelixQA bank entry
- every fix has a regression test added to the **fixes-validation** bank before the ticket is closed

**§5.3 Mandatory Retesting Loop**

After any change — feature, fix, refactor, dependency bump — the full loop below MUST run until it reaches a clean state with **zero** failing tests, **zero** open tickets, and **zero** new issues surfaced by any category:

1. Rebuild affected binaries, containers, and deployments
2. Execute every category of tests from §5.1
3. Analyze results, videos, screenshots, and logs
4. Open a ticket for every defect with severity, evidence, and repro steps
5. Fix the root cause (never a workaround) and add a regression test to the fixes-validation bank
6. Return to step 1

The loop terminates only when every category passes and no new tickets are generated in a full pass.

**§5.4 Sequential Platform Coverage**

Coverage must be achieved **one platform at a time, sequentially**, across all platforms, services, and applications, with no platform left behind:

- catalog-api (Go backend)
- catalog-web (React frontend)
- catalogizer-desktop (Tauri)
- installer-wizard (Tauri)
- catalogizer-android (Android phone)
- catalogizer-androidtv (Android TV)
- catalogizer-api-client (TypeScript library)
- every Go submodule
- every TypeScript submodule
- HelixQA itself and every AI submodule

A platform is "complete" only when every category in §5.1 is green for it.

**§5.5 Violation Consequences**

Shipping code (merging, releasing, tagging, deploying) is **prohibited** while any category is incomplete or any ticket is open. No partial "we'll get to it later" releases are permitted.

---

### Article VI: Open-Points Closure Brief (MANDATORY)

`docs/OPEN_POINTS_CLOSURE.md` is a **permanent source of truth** for every unfinished item across the program. It enumerates exactly what the operator (human, credentials, hardware, external accounts, filming time) must supply so that the program reaches **zero open points**.

**§6.1 Scope**

The brief is the authoritative list for work that cannot be completed autonomously by Claude or any automated pipeline. It covers:

1. Credentials & secrets (API keys, vault entries, rotation schedule)
2. Hardware / test harnesses (devices, runners, NAS corpus)
3. Infra & deployments (DNS, live campaigns, training datasets, signing keys)
4. Human / creative (video filming, brand review, legal review)
5. Optional hardening (non-blocking improvements)
6. Definition of done + verification commands

**§6.2 Maintenance Duty**

- Every time an open point closes, the corresponding checkbox in the brief **must** be ticked in the same commit that lands the closure.
- Every time a new operator-action item is discovered, it **must** be appended to the brief in the same commit that discovers it. It is never acceptable to leave such an item only in chat, tickets, or ad-hoc notes.
- Every session that touches production-affecting code **must** end by running the verification commands in §7 of the brief and updating the Last refresh date at the top.
- The brief is reviewed by the Project Lead at every release gate. A release is **prohibited** if the brief is out of date relative to the working state.

**§6.3 Cleanup Cadence**

The brief drives the "cleaning up the depth" workstream. Every regular maintenance cycle picks the highest-impact unclosed checkbox, drives it to ticked, and commits the closure + brief update atomically.

**§6.4 Non-negotiability**

- The brief **must** be referenced from `CLAUDE.md` and `AGENTS.md` so every Claude session reads it at start-up.
- No competing "open items" list may be created elsewhere. Add to the brief or extend it; never fork it.
- Deleting an unclosed item from the brief without actually closing it is a Constitution violation on par with committing a TODO.

---

### Article VII: Full-QA Master Cycle (MANDATORY)

**Every production QA effort must follow this rigid loop. Partial execution is prohibited.**

**§7.1 Mandatory preconditions**
- All binaries + all containers **must** be rebuilt from a clean slate before the loop begins. No cached artefacts permitted to avoid stale-code false-positives.
- `.env` (at project root) **must** supply all LLMsVerifier-scored model keys so HelixQA runs with real vision models, not stubs. Missing keys = the loop cannot begin.
- Devices listed in `.devignore` (ATMOSphere) are **permanently excluded** from Android + Android TV testing. No exception, no fallback.
- `.devconnect` **must** list at least one non-ATMOSphere Android TV + one non-ATMOSphere Android phone when those platforms are in scope.

**§7.2 Execution order (do not reorder)**
1. Clean rebuild — all apps + services + containers
2. Unit + integration tests, every submodule
3. Challenges bank run, every registered challenge
4. HelixQA bank tests, every bank
5. Full autonomous QA per app per platform — `helixqa autonomous`
6. Video + screenshot post-session review — every frame scrutinised for UI/UX imperfections
7. Ticket creation for every defect with evidence attached
8. Root-cause fixes + regression tests added to the **Fixes Validation Tests Suite**
9. Full rebuild, re-run from step 1 until a clean pass
10. Version-code bump for every app/service; release artefacts archived to `releases/<platform>/<app>/<version>/`

**§7.3 Stop conditions (only these)**
- **FATAL BLOCKER** — a defect the operator must resolve (missing hardware, credential, network route).
- **SYSTEM BREAKS** — infrastructure collapse preventing further testing.
- **NOTHING LEFT** — no defects, no warnings, no missing features, 100% pass.

Any other termination is a Constitution violation.

**§7.4 HelixQA coverage contract**
Before every autonomous session, HelixQA **must** plan — in planning mode — every:
- happy path (login, browse, open detail, play, favourite, search, settings)
- standard flow variation
- screen of every app (including Android TV channels, Tauri windows, web routes)
- UI / UX component, widget, input, button, toggle, menu
- use case, edge case
- data-set combination (positive, faulty, wrong, boundary, internationalised, malicious)
- media type in the library (movie / tv_show / tv_season / tv_episode / music_artist / music_album / song / game / software / book / comic)

HelixQA **must** use the real catalogue content — titles it already knows about — when exercising browse/play/search. Scripted generic inputs are prohibited.

No feature, flow, screen, component, or use case may be omitted. A session that skips anything is non-compliant and its results are void.

**§7.5 Evidence and ticketing**
Every defect surfaced by the post-session review **must** be ticketed with:
- video reference (filename + MM:SS timestamp)
- screenshot references (before + after state)
- session ID and step number
- reproduction path (actions taken)
- suspected root cause (if identifiable)

Tickets without complete evidence are rejected (§ existing HelixQA MANDATORY rules).

**§7.6 Fixes Validation Tests Suite**
Every fix **must** land with:
- a unit or integration test asserting the fix
- a regression entry in `banks/fixes-validation-*.json`
- a HelixQA bank entry replaying the defect scenario
- a challenge registration

All four artefacts in the **same commit** that lands the fix. A fix without its four-artefact tail is a Constitution violation.

**§7.7 Live monitoring**
During the loop, the operator console **must** always show:
- current platform under test
- current app or service under test
- current test case ID + short description
- progress (e.g. step 12 / 34)
- running + final result

Comprehensive logs stream to `docs/reports/qa-sessions/<YYYY-MM-DD-THH-MM>/` with per-step timing, actions taken, LLM reasoning, evidence produced, and final verdicts.

**§7.8 Archiving**
Every session's artefacts are permanent. `docs/reports/qa-sessions/` is **not** gitignored. Each session directory contains:
- `FINAL-REPORT.md` — aggregated results
- `logs/` — per-run command logs
- `challenges/` — JSON results + summary
- `helixqa/` — bank + autonomous results
- `videos/` — MP4 recordings
- `screenshots/` — pre + post per action
- `tickets/` — markdown tickets with evidence
- `analysis/` — deep analysis, suggestions, conclusions for further improvements

**§7.9 Release promotion**
On a clean-pass exit (§7.3 "NOTHING LEFT"):
- version codes increment for every app + service
- debug + release builds land in `releases/<platform>/<app>/<version>/`
- release notes update `docs/releases/v<version>.md`
- submodule pointers bump in main repo
- all upstream remotes advanced

**§7.10 Extensibility mandate**
The testing systems (Challenges, HelixQA, and all of their dependency submodules) **must** be extended continuously — every session adds coverage, every iteration adds regression defences. OSS research is encouraged: if a cutting-edge open-source framework raises the bar, it is vendored and integrated.

**§7.11 Violation enforcement**
Shipping (merging, releasing, tagging, deploying) is **prohibited** while the master cycle has not completed at least one clean pass under §7.3 "NOTHING LEFT". No partial "we'll re-run next sprint" releases are permitted.

---

*Last Updated: 2026-04-18*
*Enforced by: Project Lead*
