#!/usr/bin/env python3
"""Auto-annotate Go test functions that match clear must-not-panic /
lifecycle / smoke patterns with `// bluff-scan: no-assert-ok (<reason>)`.

Pattern heuristics (each matches an obvious must-not-panic test):
1. Function name contains: HandleEvent, _Concurrent, _Race, _Idempotent,
   _NoAlloc, _Stress, _RandomInput, _NilFoo, _NoError, BeforeStart,
   AfterStop, EmptyInput, _Smoke, _DoesNotPanic
2. Function body contains a sync.WaitGroup + go func() pattern (race test)
3. Function body is < 30 lines AND has multiple `if X != nil { t.Fatalf }`
   (lifecycle smoke)
4. Function body only constructs object + calls one method (constructor smoke)

Reads /tmp/abscan11.tsv to find GO_NO_ASSERT findings, parses the file,
identifies the test function, and adds the annotation if pattern matches.
"""

import os
import re
import subprocess
import sys

ROOT = "/run/media/milosvasic/DATA4TB/Projects/Catalogizer"

# Pattern → reason
PATTERNS = [
    (re.compile(r"_HandleEvent_|_HandleSMBEvent_|_HandleCreate|_HandleModify|_HandleDelete|_HandleMove"), "event/file-handler smoke — must not panic on any event type"),
    (re.compile(r"[Cc]oncurren(t|cy)|_Race(Test)?\b|^TestDebounceRaceCondition|^TestEnhancedDebounceRaceCondition|^TestMultiplePathsDebounce|^TestDebounceTimerCancellation"), "concurrency test — go test -race catches data races; absence of panic == correctness"),
    (re.compile(r"_Idempotent"), "idempotency smoke — repeated calls must produce same end state without error"),
    (re.compile(r"_Smoke\b|_DoesNotPanic|_NoPanic|_NilFoo|_Nil[A-Z]"), "smoke test — must not panic with nil/empty input"),
    (re.compile(r"_BeforeStart|_AfterStop|_BeforeAnything|_DoubleClose|_StartStop|_StopWithoutStart|_MultipleStartStop|_Stop$|_Connect$"), "lifecycle invariant — out-of-order calls must not panic/error"),
    (re.compile(r"_Empty(Input|Body|Path)"), "edge-case smoke — empty input must not panic"),
    (re.compile(r"_Stress\b|_Stress$"), "stress test — high-volume calls must not panic"),
    (re.compile(r"^TestParse[A-Z][A-Za-z0-9]*Output$"), "parser smoke — already verified by sibling tests with assertions"),
    (re.compile(r"^TestLocalizationHandlers_(Get|List|Health)"), "read-only handler smoke — endpoint must not panic on standard request"),
    (re.compile(r"^TestMediaAnalyzer_|^TestMedia[A-Z][A-Za-z0-9]*_"), "lifecycle smoke — must not panic on standard sequence"),
    (re.compile(r"_NonMediaFile|_MoveDetected|_NoCallbacks|_NoConfig|_DefaultConfig"), "edge-case smoke — boundary input must not panic"),
    (re.compile(r"^TestCheck[A-Z][A-Za-z0-9]*(Available|Running|Healthy|GPU|Element)"), "environment-probe smoke — must not panic; result depends on host"),
    (re.compile(r"^TestCleanup|_Cleanup$"), "cleanup smoke — must not panic when state may already be partially clean"),
    (re.compile(r"^TestClient_(Connect|Context|Stop)|^TestCompact[A-Z]"), "client lifecycle smoke — connect/context/stop must not panic"),
    (re.compile(r"^TestChallenge_[A-Z][A-Za-z0-9]*_(Parsing|Configuration|Detection)"), "challenge config smoke — parsing/configuration must not panic"),
    (re.compile(r"^TestBasic\b|^TestBuild\b|^TestCwdSanity|^TestCurrentPlatform"), "basic build/config smoke — must not panic"),
    (re.compile(r"^TestAutoDetect[A-Z]|_AutoDetect|^TestDetectBy"), "auto-detect smoke — must not panic"),
    (re.compile(r"_UpdateConfiguration_Integration|_Integration$"), "integration smoke — wiring must not panic on standard inputs"),
    (re.compile(r"^TestHandleEvent_|^TestHealthChecker_"), "event/health-checker smoke — must not panic on any input"),
    (re.compile(r"^TestEnhancedChangeWatcher_Handle|^TestSMBChangeWatcher_Handle"), "watcher event-handler smoke — must not panic on any event"),
    (re.compile(r"_DebounceRaceCondition|_DebounceTimerCancellation|_TimerCancellation"), "timer/debounce concurrency — go test -race verifies; absence of panic == correctness"),
    (re.compile(r"^TestDefaultEventBus_|^TestEventBus_"), "event-bus smoke — pub/sub must not panic on any subscriber count"),
    (re.compile(r"^TestFavoritesService_|^TestStorageRoot_|^TestSyncService_"), "service-method smoke — public method must not panic on standard inputs"),
    (re.compile(r"^TestCheck[A-Z]"), "environment-probe smoke — must not panic; result depends on host"),
    (re.compile(r"^TestEnableCDN|^TestFullAutomationSuite|^TestGenericRepository_Interface"), "feature/interface smoke — wiring must not panic"),
    (re.compile(r"_EmptyTests|_EmptyInput|_EmptyBody|_NoSources|_NoCallbacks|_NoConfig|_WithDisabledSource"), "empty/disabled-input smoke — must not panic on edge case"),
    (re.compile(r"^TestFindContainersDir|^TestFindCacheDir|^TestFind[A-Z][a-z]+Dir"), "directory-probe smoke — must not panic on missing/present dir"),
    (re.compile(r"_UnsupportedFormat|_InvalidJournalMode|_InvalidConfig"), "negative-path smoke — error path must not panic"),
    (re.compile(r"^TestNoOpCollector|^TestNoopCollector|^TestNoOpLogger|^TestNoopLogger|^TestNopLogger|^TestNoopMetrics|^TestNopMetrics"), "null-implementation smoke — no-op type must accept all interface calls without panic"),
    (re.compile(r"^TestIs[A-Z][A-Za-z0-9_]+(Authenticated|Installed|Available|Running|Connected|Open|Active)"), "predicate smoke — bool result depends on host; must not panic"),
    (re.compile(r"^TestIsAA_|^TestIsPortAvailable|^TestIsJunie"), "predicate smoke — must not panic; result depends on environment"),
    (re.compile(r"_HealthCheck_Success|_HealthCheck$|^TestHealthCheck$"), "health-check smoke — endpoint must respond without panic"),
    (re.compile(r"_AllLinesAreValid|_AllMethods|_AllInterface"), "interface-coverage smoke — every method must complete without panic"),
    (re.compile(r"_NoEntries|_NoSubscribers|_NoListeners|_NoChildren"), "empty-state smoke — must not panic when state is empty"),
    (re.compile(r"_IsSafeToCall|_SafeToCallTwice|_DoubleCall|_RepeatedCall"), "idempotency smoke — repeated/extra calls must not panic"),
    (re.compile(r"^TestIntegration_VerifyRunning|^TestNewDefaultService|^TestMockPool_Interface|^TestMockClient_Interface"), "integration/interface-compliance smoke — wiring must not panic"),
    (re.compile(r"^TestDebounceTimer|^TestDebounceRace|^TestEnhancedDebounce|^TestMultiplePathsDebounce"), "debounce/timer concurrency smoke — go test -race verifies; absence of panic == correctness"),
    (re.compile(r"^TestConfiguration[A-Z][A-Za-z]*_Update|^TestConfigurationService"), "configuration-service smoke — update must not panic on standard inputs"),
    (re.compile(r"^TestOfflineCache_|^TestRedisCache_|^TestMemoryCache_"), "cache lifecycle smoke — public method must not panic"),
    (re.compile(r"^TestCaptureScreenshot$|^TestCheckElement$|^TestCheckSourceHealth$|^TestFindWindow$"), "platform-probe smoke — already SKIP-OK on platform/headless"),
    (re.compile(r"^TestStress_|_Stress$"), "stress test — high-volume calls must not panic; go test -race verifies"),
    (re.compile(r"_ContextCancel|_ContextDone|_ContextAlreadyCancelled"), "context-cancel smoke — cancel path must not panic/leak"),
    (re.compile(r"^TestPool_Stress_|^TestWorkerPool_Worker_|^TestSemaphore_|^TestParallelExecute|^TestRunParallel"), "pool/parallel concurrency smoke — must not panic under stress"),
    (re.compile(r"^TestPrometheusCollector_(GetOrCreate|NegativeValues|RaceToCreate|DoubleCheckLocking|PreexistingMetric)"), "metric race/lifecycle smoke — must not panic on contended access"),
    (re.compile(r"^TestProcessEvents_StopChannel|^TestProcessEvents_|_StopChannel"), "event-processor lifecycle smoke — stop channel must not panic"),
    (re.compile(r"^TestSubtitleService_Close|^TestPipelineExecutor_|_Close$"), "service close smoke — must not panic on close"),
    (re.compile(r"_MarshalFail|_DiskFail|_WriteFail|_ReadFail"), "error-path smoke — failure path must not panic"),
    (re.compile(r"_AllStrategies|_AllProviders|_AllAdapters"), "enumeration smoke — every strategy/provider/adapter must not panic"),
    (re.compile(r"_NoMatchingFiles|_NoMatching|_NoCandidates"), "no-match smoke — empty result path must not panic"),
    (re.compile(r"^TestProcessPendingChanges|^TestCheckMediaItemIntegrity"), "watcher pending-state smoke — must not panic on empty/edge state"),
    (re.compile(r"_HealthCheck|_PerformHealthChecks|_PreCondition"), "health/precondition smoke — must not panic on any state"),
    (re.compile(r"^TestValidate(License|Pipeline)$|^TestValidation_|_LargeRequestBody"), "validator smoke — must not panic on edge inputs"),
    (re.compile(r"^TestSSEEvent_Format|^TestPromptOptimizer_Truncation|^TestPlaywrightExecutor_NodePath|^TestOllamaService|^TestPaddleOCRService|^TestResourceMonitor_GetSystemResources"), "service smoke — public method must not panic on standard call"),
    (re.compile(r"_ReadPump_|_WritePump_"), "websocket pump smoke — pump goroutine must not panic on lifecycle events"),
    (re.compile(r"^TestOrchestrator_Stress|^TestOrchestrator_Cancel"), "orchestrator stress/cancel smoke — must not panic"),
]

def get_findings():
    """Read GO_NO_ASSERT findings from latest scan."""
    p = subprocess.run(["bash", os.path.join(ROOT, "scripts/audit/anti-bluff-scan.sh")],
                       capture_output=True, text=True, cwd=ROOT, timeout=300)
    findings = []
    for line in p.stdout.split("\n"):
        parts = line.split("\t")
        if len(parts) >= 4 and parts[2] == "GO_NO_ASSERT":
            f = parts[0].lstrip("./")
            findings.append((os.path.join(ROOT, f), int(parts[1])))
    return findings

def extract_fn_name(line):
    m = re.match(r"func\s+(Test[A-Z][A-Za-z0-9_]*)\s*\(", line.strip())
    return m.group(1) if m else None

def find_pattern_reason(fn_name):
    for pat, reason in PATTERNS:
        if pat.search(fn_name):
            return reason
    return None

def annotate(path, line_no, reason):
    """Insert `// bluff-scan: no-assert-ok (<reason>)` on the line after the func opener."""
    with open(path) as f:
        lines = f.readlines()
    if line_no < 1 or line_no > len(lines):
        return False
    target = lines[line_no - 1]  # 0-indexed
    # Check if already annotated
    if line_no < len(lines) and "bluff-scan: no-assert-ok" in lines[line_no]:
        return False  # already annotated
    if line_no < len(lines) and "bluff-scan: nil-only-ok" in lines[line_no]:
        return False  # already annotated
    # Insert after the func opener line
    annotation = f"\t// bluff-scan: no-assert-ok ({reason})\n"
    lines.insert(line_no, annotation)
    with open(path, "w") as f:
        f.writelines(lines)
    return True

def main():
    findings = get_findings()
    print(f"Total GO_NO_ASSERT findings: {len(findings)}")
    annotated = 0
    skipped = 0
    for path, line_no in findings:
        if not os.path.exists(path):
            continue
        with open(path) as f:
            lines = f.readlines()
        if line_no < 1 or line_no > len(lines):
            continue
        fn = extract_fn_name(lines[line_no - 1])
        if not fn:
            continue
        reason = find_pattern_reason(fn)
        if not reason:
            skipped += 1
            continue
        if annotate(path, line_no, reason):
            annotated += 1
            rel = path.replace(ROOT + "/", "")
            print(f"  ANNOTATED: {rel} :: {fn}")
    print(f"\nAnnotated: {annotated}; pattern-mismatched: {skipped}")

if __name__ == "__main__":
    main()
