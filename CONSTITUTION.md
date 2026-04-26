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

### Article VIII: Device State Preservation (MANDATORY)

**A QA session MUST return every test device to the exact state it was in at session start. Modifying device-level settings as a side effect is forbidden.**

**§8.1 What must never change as a side effect**
- `system.font_scale` / accessibility text-scaling
- Display density (`wm density`) / screen resolution (`wm size`)
- Screen brightness and brightness mode
- Screen-off timeout
- Auto-rotation
- Any other `settings put system|secure|global …` key that outlives the app under test

**§8.2 Mechanism**
HelixQA captures the sensitive keys on every connected device at Phase 0b (after ADB reverse proxy setup) and registers a restoration deferred to the end of the pipeline. On normal completion, crash, timeout, or operator interrupt, the settings are written back to their captured values. See `HelixQA/pkg/autonomous/device_preserve.go`.

**§8.3 Root-cause over cosmetic**
If the LLM-driven curiosity phase navigates into device settings and flips a switch, the immediate fix is **both**:
1. Constrain curiosity navigation so it never targets device-settings areas that fall outside the app under test.
2. The preservation hook in §8.2 is defence in depth, not a permission slip.

**§8.4 Verification**
Every session's FINAL-REPORT.md includes a "Device state diff" section confirming every preserved key matches its captured value. A non-empty diff is a Constitution violation.

---

### Article IX: HelixQA Test Tool Hygiene (MANDATORY)

**HelixQA is testing infrastructure. It must never ask the operator to paper over its own limitations.**

**§9.1 No manual workarounds**
- No manual screenrecord loops. No scripts that substitute for HelixQA recording.
- No "run this script after the session to fix …" instructions.
- If the autonomous pipeline produces broken output, fix the Go code.

**§9.2 Recording must span the whole session**
Android's `screenrecord` caps at 180 s per invocation. HelixQA's recorder runs in a loop and concatenates segments via `ffmpeg -f concat -c copy`. A 2-hour session produces a 2-hour MP4. Partial recordings are a Constitution violation. See `HelixQA/pkg/video/scrcpy.go`.

**§9.3 Vision verification must tolerate provider variety**
Some vision providers (e.g., astica) return natural-language descriptions and don't emit the literal `VERIFIED: yes` marker. When the response is non-empty but doesn't parse, HelixQA falls back to action-success. See `HelixQA/pkg/autonomous/structured_executor.go` (`verifyOutcome`).

**§9.4 False-positive reporting is prohibited**
An orchestrator or pipeline that reports "✓ completed successfully" when the underlying command failed is a Constitution violation (see FIX-QA-2026-04-20-001/002). Every success log must be gated by a real passing assertion, never by `tee` exit codes or similar shell pipeline artefacts.

### Article X: No Sudo / No Root Execution (MANDATORY)

**ALL operations MUST run at local user level ONLY.**

**§X.1 Permanent Security Constraint**

This is a PERMANENT and NON-NEGOTIABLE security constraint:

- **NEVER** use `sudo` in ANY command
- **NEVER** use `su` in ANY command
- **NEVER** execute operations as `root` user
- **NEVER** elevate privileges for file operations
- **ALL** infrastructure commands MUST use user-level container runtimes (rootless podman/docker)
- **ALL** file operations MUST be within user-accessible directories
- **ALL** service management MUST be done via user systemd or local process management
- **ALL** builds, tests, and deployments MUST run as the current user

**§X.2 Container-Based Solutions**

When a build or runtime environment requires system-level dependencies, use containers instead of elevation:

- **Use the `Containers` submodule** (`https://github.com/vasic-digital/Containers`) for containerized build and runtime environments
- **Add the `Containers` submodule as a Git dependency** and configure it for local use within the project
- **Build and run inside containers** to avoid any need for privilege escalation
- **Rootless Podman/Docker** is the preferred container runtime

**§X.3 Rationale**

- **Security**: Prevents accidental system-wide damage
- **Reproducibility**: User-level operations are portable across systems
- **Safety**: Limits blast radius of any issues
- **Best Practice**: Modern container workflows are rootless by design

**§X.4 Violation Consequences**

Any use of `sudo`, `su`, or root-level execution in scripts, builds, or operations is a Constitution violation.

---

---

*Last Updated: 2026-04-21*
*Enforced by: Project Lead*


## Universal Mandatory Constraints

These rules are inherited from the cross-project Universal Mandatory Development Constraints (canonical source: `/tmp/UNIVERSAL_MANDATORY_RULES.md`, derived from the HelixAgent root `CLAUDE.md`). They are non-negotiable across every project, submodule, and sibling repository. Project-specific addenda are welcome but cannot weaken or override these.

### Hard Stops (permanent, non-negotiable)

1. **NO CI/CD pipelines.** No `.github/workflows/`, `.gitlab-ci.yml`, `Jenkinsfile`, `.travis.yml`, `.circleci/`, or any automated pipeline. No Git hooks either. All builds and tests run manually or via Makefile / script targets.
2. **NO HTTPS for Git.** SSH URLs only (`git@github.com:…`, `git@gitlab.com:…`, etc.) for clones, fetches, pushes, and submodule operations. Including for public repos. SSH keys are configured on every service.
3. **NO manual container commands.** Container orchestration is owned by the project's binary / orchestrator (e.g. `make build` → `./bin/<app>`). Direct `docker`/`podman start|stop|rm` and `docker-compose up|down` are prohibited as workflows. The orchestrator reads its configured `.env` and brings up everything.

### Mandatory Development Standards

1. **100% Test Coverage.** Every component MUST have unit, integration, E2E, automation, security/penetration, and benchmark tests. No false positives. Mocks/stubs ONLY in unit tests; all other test types use real data and live services.
2. **Challenge Coverage.** Every component MUST have Challenge scripts (`./challenges/scripts/`) validating real-life use cases. No false success — validate actual behavior, not return codes.
3. **Real Data.** Beyond unit tests, all components MUST use actual API calls, real databases, live services. No simulated success. Fallback chains tested with actual failures.
4. **Health & Observability.** Every service MUST expose health endpoints. Circuit breakers for all external dependencies. Prometheus / OpenTelemetry integration where applicable.
5. **Documentation & Quality.** Update `CLAUDE.md`, `AGENTS.md`, and relevant docs alongside code changes. Pass language-appropriate format/lint/security gates. Conventional Commits: `<type>(<scope>): <description>`.
6. **Validation Before Release.** Pass the project's full validation suite (`make ci-validate-all`-equivalent) plus all challenges (`./challenges/scripts/run_all_challenges.sh`).
7. **No Mocks or Stubs in Production.** Mocks, stubs, fakes, placeholder classes, TODO implementations are STRICTLY FORBIDDEN in production code. All production code is fully functional with real integrations. Only unit tests may use mocks/stubs.
8. **Comprehensive Verification.** Every fix MUST be verified from all angles: runtime testing (actual HTTP requests / real CLI invocations), compile verification, code structure checks, dependency existence checks, backward compatibility, and no false positives in tests or challenges. Grep-only validation is NEVER sufficient.
9. **Resource Limits for Tests & Challenges (CRITICAL).** ALL test and challenge execution MUST be strictly limited to 30-40% of host system resources. Use `GOMAXPROCS=2`, `nice -n 19`, `ionice -c 3`, `-p 1` for `go test`. Container limits required. The host runs mission-critical processes — exceeding limits causes system crashes.
10. **Bugfix Documentation.** All bug fixes MUST be documented in `docs/issues/fixed/BUGFIXES.md` (or the project's equivalent) with root cause analysis, affected files, fix description, and a link to the verification test/challenge.
11. **Real Infrastructure for All Non-Unit Tests.** Mocks/fakes/stubs/placeholders MAY be used ONLY in unit tests (files ending `_test.go` run under `go test -short`, equivalent for other languages). ALL other test types — integration, E2E, functional, security, stress, chaos, challenge, benchmark, runtime verification — MUST execute against the REAL running system with REAL containers, REAL databases, REAL services, and REAL HTTP calls. Non-unit tests that cannot connect to real services MUST skip (not fail).
12. **Reproduction-Before-Fix (CONST-032 — MANDATORY).** Every reported error, defect, or unexpected behavior MUST be reproduced by a Challenge script BEFORE any fix is attempted. Sequence: (1) Write the Challenge first. (2) Run it; confirm fail (it reproduces the bug). (3) Then write the fix. (4) Re-run; confirm pass. (5) Commit Challenge + fix together. The Challenge becomes the regression guard for that bug forever.
13. **Concurrent-Safe Containers (Go-specific, where applicable).** Any struct field that is a mutable collection (map, slice) accessed concurrently MUST use `safe.Store[K,V]` / `safe.Slice[T]` from `digital.vasic.concurrency/pkg/safe` (or the project's equivalent primitives). Bare `sync.Mutex + map/slice` combinations are prohibited for new code.

### Definition of Done (universal)

A change is NOT done because code compiles and tests pass. "Done" requires pasted terminal output from a real run, produced in the same session as the change.

- **No self-certification.** Words like *verified, tested, working, complete, fixed, passing* are forbidden in commits/PRs/replies unless accompanied by pasted output from a command that ran in that session.
- **Demo before code.** Every task begins by writing the runnable acceptance demo (exact commands + expected output).
- **Real system, every time.** Demos run against real artifacts.
- **Skips are loud.** `t.Skip` / `@Ignore` / `xit` / `describe.skip` without a trailing `SKIP-OK: #<ticket>` comment break validation.
- **Evidence in the PR.** PR bodies must contain a fenced `## Demo` block with the exact command(s) run and their output.
