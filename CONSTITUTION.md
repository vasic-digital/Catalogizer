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

*Last Updated: 2026-04-12*
*Enforced by: Project Lead*
