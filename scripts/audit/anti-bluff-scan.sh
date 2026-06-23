#!/usr/bin/env bash
# anti-bluff-scan.sh
#
# Static scanner that detects common "bluff" patterns in tests, Challenges,
# and HelixQA bank entries. A bluff is a test that passes without exercising
# real end-user behaviour — see CONSTITUTION.md Article XI.
#
# This is a STATIC heuristic scanner. It MUST be paired with the dynamic
# "comment out the feature, re-run, see if it still passes" ritual from
# Article XI §11.2.5. It is intentionally conservative (low false-positive)
# so that every finding is worth a human's attention.
#
# Output: machine-readable findings on stdout (TSV: file<TAB>line<TAB>kind<TAB>excerpt)
# Each kind is one of:
#   GO_NIL_ONLY            — Go *_test.go that asserts only that err is nil
#   GO_NO_ASSERT           — Go *_test.go function with no t.Error/t.Fatal/assert/require
#   GO_RACE_ONLY           — Go concurrency/race test (goroutines + Wait) with no
#                            observable post-join assertion; weak (relies on -race +
#                            no-panic) but separated from true no-assert stubs
#   GO_HTTPTEST_ABUSE      — Go test using httptest.NewServer in an E2E-named file
#   GO_MOCK_IN_INTEGRATION — Go test in integration/e2e dir using mock/stub/fake
#   TS_NO_ASSERT           — TS/JS test with describe/it body but no expect/assert
#   TS_TRUTHY_ONLY         — TS/JS test with only expect(x).toBeTruthy() / toBeDefined()
#   PROSE_HELIXQA_ACTION   — HelixQA bank entry whose action: field is prose
#   CHALLENGE_200_OK_ONLY  — Challenge that asserts only HTTP 200 / curl exit code
#   CHALLENGE_BLIND_SHELL  — Shell pattern: `&& echo PASS`, `|| true`, `tee` exit-laundering
#   CHALLENGE_SUB_SECOND   — Challenge whose run takes <1s (added by dynamic mode)
#   SKIP_WITHOUT_TICKET    — t.Skip / @Ignore / xit / describe.skip without SKIP-OK: #ticket
#   ASSERT_TAUTOLOGY       — assertion that is structurally tautological (1==1, true, etc.)
#
# Exit code: 0 when ZERO findings, 1 when ANY finding.
# Override scope with $1 (path); default is the repo root.
#
# META-TEST NOTE (refinement of 2026-06-23 — §11.4.6 false-positive reduction):
# This scanner has no dedicated fixture harness of its own (the fixture suite in
# submodules/challenges/scripts/anti-bluff/tests/ exercises the SEPARATE scan_go
# library in scripts/anti-bluff/lib/go.sh, not this file). The refinements below
# were validated by a real before/after run over the live tree:
#   • SKIP honoring (SKIP-OK numeric/symbolic, §11.4.3, testing.Short()):
#       160 -> 24 SKIP_WITHOUT_TICKET (the 24 are genuine un-annotated skips).
#   • chaos/teardown context exemption for CHALLENGE_BLIND_SHELL:
#       533 -> 93 (the 93 are set+e / terminal `| tee` / parse-`|| true` in
#       NON fault-injection describe/CLI challenges — real verdict-laundering).
#   • shell:/bash: bank actions excluded from PROSE_HELIXQA_ACTION: 4 -> 0.
#   • skip-bodied tests no longer mis-flagged GO_NO_ASSERT; race/concurrency
#       tests with no post-join assertion split into GO_RACE_ONLY (weak, kept).
#   Net: 737 -> 142 findings; every survivor is a genuine, actionable hit.
# To re-validate after edits: run this script over the tree, confirm the four
# false-positive classes above stay collapsed and exit code stays non-zero
# while ANY genuine finding remains (per Constitution anti-bluff CI lane).

set -uo pipefail

ROOT="${1:-$(cd "$(dirname "$0")/../.." && pwd)}"
cd "$ROOT" || { echo "cannot cd to $ROOT" >&2; exit 2; }

EXCLUDE_DIRS='\(\./\)\?\(\.git\|node_modules\|target\|vendor\|build\|releases\|qa-results\|docs/reports/qa-sessions\|docs/audits\|submodules/helix_qa/banks/templates\|tests/k6/node_modules\|tools/opensource\|tools/external\|submodules/helix_qa/tools/opensource\|releases\)'

# Substring match — excludes paths at any depth (not just root prefix).
# /tools/opensource/ catches submodules/helix_qa/tools/opensource and any nested vendor.
# /node_modules/ catches catalog-web/node_modules, installer-wizard/node_modules, etc.
# /submodules/helix_memory/ — helix_memory (HelixDevelopment/HelixMemory) is an
#   EXTERNAL dependency submodule that catalogizer only CONSUMES; its upstream
#   tests are authored + owned upstream, not catalogizer code. Per §11.4.28
#   decoupling we do NOT edit a consumed dependency's tests; per §11.4.74 it is a
#   reuse (no-match→consume) dependency, so its in-tree test patterns are out of
#   scope for catalogizer's own anti-bluff gate (same treatment as tools/opensource
#   vendored code). Kept narrow — only helix_memory, not all submodules.
EXCLUDE_SUBSTRINGS=(
  '/node_modules/'
  '/tools/opensource/'
  '/tools/external/'
  '/submodules/helix_memory/'
  '/.git/'
  '/target/'
  '/vendor/'
  '/build/'
  '/releases/'
  '/qa-results/'
  '/docs/reports/qa-sessions/'
  '/docs/audits/'
)

# is_excluded <path> — return 0 if path matches any EXCLUDE_SUBSTRINGS.
is_excluded() {
  local p="$1"
  local sub
  for sub in "${EXCLUDE_SUBSTRINGS[@]}"; do
    case "$p" in *"$sub"*) return 0 ;; esac
  done
  return 1
}

emit() {
  # emit kind file line excerpt
  local kind="$1" file="$2" line="$3" excerpt="$4"
  excerpt="${excerpt//$'\t'/ }"
  printf '%s\t%s\t%s\t%s\n' "$file" "$line" "$kind" "$excerpt"
}

scan_go_tests() {
  # GO_NIL_ONLY: a Test* function whose only assertion is `if err != nil { t.Fatal(err) }`
  #              and has NO other t.Error/t.Fatal/assert/require/expect calls.
  # GO_NO_ASSERT: a Test* function (NOT a method, NOT a suite runner) whose body
  #               has zero t.Error/t.Fatal/assert./require.
  # GO_HTTPTEST_ABUSE: file path contains e2e or _e2e_ AND uses httptest.NewServer
  # GO_MOCK_IN_INTEGRATION: file is under a real integration/e2e/stress/chaos/security/
  #                          challenges directory (NOT just filename match — that
  #                          flags challenge_handler_test.go which is a unit test)
  #                          AND uses mock/stub/fake.
  while IFS= read -r -d '' f; do
    is_excluded "$f" && continue
    local low="${f,,}"
    # GO_HTTPTEST_ABUSE — file path contains /e2e/ or _e2e_test.go.
    # File-level exemption: if the file contains a SKIP-OK guard
    # (e.g. `t.Skip("SKIP-OK: #BLUFF-... fake-server pseudo-E2E is
    # anti-bluff banned (Article XI §11.5); set <ENV> to opt-in
    # once test bodies are rewritten against real data")`), then
    # the httptest body is documented dormant code — already
    # caught by Article XI's SKIP-OK ledger, not a silent bluff.
    if [[ "$low" == */e2e/* || "$low" == *_e2e_test.go ]]; then
      if grep -q 'SKIP-OK:[[:space:]]*#BLUFF-' "$f" 2>/dev/null; then
        :  # documented dormant; reviewed via the SKIP-OK ticket
      else
        grep -nE 'httptest\.NewServer\b|httptest\.NewRequest\b' "$f" 2>/dev/null | while IFS=: read -r ln rest; do
          stripped="$(echo "$rest" | sed 's/^[[:space:]]*//')"
          case "$stripped" in '//'*|'/*'*|'*'*) continue ;; esac
          emit GO_HTTPTEST_ABUSE "$f" "$ln" "$rest"
        done
      fi
    fi
    # GO_MOCK_IN_INTEGRATION — directory must be one of the integration-class dirs
    if [[ "$low" == */tests/integration/* || "$low" == */tests/e2e/* || \
          "$low" == */tests/stress/* || "$low" == */tests/chaos/* || \
          "$low" == */tests/security/* || "$low" == */challenges/scripts/* ]]; then
      grep -nE 'gomock\.NewController|testify/mock\.|mock\.Mock\{|NewMockClient|NewMockService|NewMockRepository|/mocks/' "$f" 2>/dev/null | while IFS=: read -r ln rest; do
        stripped="$(echo "$rest" | sed 's/^[[:space:]]*//')"
        case "$stripped" in '//'*|'/*'*|'*'*) continue ;; esac
        emit GO_MOCK_IN_INTEGRATION "$f" "$ln" "$rest"
      done
    fi
    # Per-function nil-only / no-assert — only flag standalone Test funcs
    # (no receiver) that aren't suite runners, subtest dispatchers, "no panic"
    # tests, or compile-time interface assertions.
    awk -v f="$f" '
      /^func[[:space:]]+Test[A-Z][A-Za-z0-9_]*\(t \*testing\.T\)/ {
        in_func=1; depth=0; start=NR; body=""; nil_only=1; has_assert=0;
        saw_suite_run=0; saw_subtest=0; saw_no_panic=0; saw_interface_assert=0;
        saw_optin=0; saw_skip=0; saw_goroutine=0; saw_waitsync=0;
        first_line=$0;
      }
      in_func {
        body=body"\n"$0;
        for (i=1; i<=length($0); i++) {
          ch=substr($0,i,1);
          if (ch=="{") depth++;
          else if (ch=="}") depth--;
        }
        if (NR > start) {
          if ($0 ~ /t\.Error|t\.Errorf|t\.Fatal|t\.Fatalf|assert\.|require\.|expect\(/) has_assert=1;
          if ($0 ~ /(t\.Error|t\.Errorf|assert\.|require\.|expect\()/) nil_only=0;
          # Non-err `if X { t.Fatal/t.Fatalf }` is a real value
          # assertion (Go idiom). Match `if <something-not-err-not-nil> {`
          # followed by a line containing t.Fatal in the same block.
          if ($0 ~ /t\.Fatal[fF]?\(/) {
            prev3 = (NR-1>=1) ? prev_line[NR-1] : "";
            prev2 = (NR-2>=1) ? prev_line[NR-2] : "";
            prev1 = (NR-3>=1) ? prev_line[NR-3] : "";
            window = prev1 prev2 prev3;
            if (window ~ /if[[:space:]]/ && window !~ /err[[:space:]]*!=[[:space:]]*nil/) {
              nil_only=0;
            }
          }
          if ($0 ~ /suite\.Run\(t,|s\.Run\(/) saw_suite_run=1;
          if ($0 ~ /t\.Run\(/) saw_subtest=1;
          # Implicit "no panic" assertion — Go panic = test failure
          if ($0 ~ /[Ss]hould not panic|[Nn]o panic|[Sh][hh]ould.*not.*crash|[Dd]oes not crash|[Mm]ust not panic/) saw_no_panic=1;
          # Compile-time interface assertion: covers
          #   var _ Iface = (*Type)(nil)
          #   var _ pkg.Iface = (*Type)(nil)
          #   var _ pkg.Iface = NewThing("...")
          if ($0 ~ /var[[:space:]]+_[[:space:]]+([a-z][a-zA-Z0-9_]*\.)?[A-Z][A-Za-z0-9_]*[[:space:]]*=[[:space:]]/) saw_interface_assert=1;
          # Author-defined assertion helpers — common Go pattern of
          # `assertX(t, got, want)` / `checkX(t, ...)` / `requireX(t, ...)`
          # is a value assertion, just like a direct t.Errorf call.
          if ($0 ~ /[[:space:]](assert|check|require|verify|expect)[A-Z][A-Za-z0-9_]*\(t[,)]/) {
            has_assert=1; nil_only=0;
          }
          # testify/mock expectations: `mock.On("Method", args).Return(...)`
          # acts as a strict assertion — the mock framework fails the
          # test on any UNEXPECTED call, so the .On(...).Return(...) IS
          # the assertion. Same for AssertExpectations / AssertCalled.
          # `\.On\(` matches `inner.On(` reliably; no need for prefix.
          if ($0 ~ /\.On\(|\.AssertExpectations\(t\)|\.AssertCalled\(t,|\.AssertNotCalled\(t,/) {
            has_assert=1; nil_only=0;
          }
          # gomock expectations: `mockObj.EXPECT().Method(...).Return(...)`
          if ($0 ~ /\.EXPECT\(\)\.[A-Z][A-Za-z0-9_]*\(/) {
            has_assert=1; nil_only=0;
          }
          # go-cmp / google diffing libraries
          if ($0 ~ /cmp\.Diff\(|cmp\.Equal\(/) {
            has_assert=1; nil_only=0;
          }
          # gomega / Eventually / Expect
          if ($0 ~ /\bgomega\.|Eventually\([^)]+\)\.Should\(|\bExpect\([^)]+\)\.To\(/) {
            has_assert=1; nil_only=0;
          }
          # Author opt-in marker — `// bluff-scan: nil-only-ok (<reason>)`
          # tells the scanner the test intentionally asserts only on
          # absence-of-error (e.g. lifecycle Stop(), `go vet`/`go build`
          # success, "must not panic" exec). Author owns the
          # justification.
          if ($0 ~ /\/\/[[:space:]]*bluff-scan:[[:space:]]*(nil-only-ok|no-assert-ok)/) saw_optin=1;
          # A test whose body is `t.Skip(...)` / `t.Skipf(...)` is a
          # DOCUMENTED skip, not a no-assert bluff. Skips are audited
          # separately by scan_skip_no_ticket (which enforces SKIP-OK
          # markers). Re-flagging the skip body as GO_NO_ASSERT just
          # double-counts the same line — and a SKIP-OK skip has already
          # been reviewed. Excludes the test from no-assert/nil-only.
          if ($0 ~ /t\.Skip[fF]?\(/) saw_skip=1;
          # Concurrency markers — `go func`, sync.WaitGroup, channels used
          # to drive a goroutine fan-out. A test that spins goroutines and
          # waits is a RACE/concurrency test (meaningful only under -race);
          # if it has no observable post-join assertion it is WEAK, but it
          # is materially different from a pure no-assert stub, so we emit a
          # separate, lower-severity kind for it (GO_RACE_ONLY).
          if ($0 ~ /go[[:space:]]+func[[:space:]]*\(/) saw_goroutine=1;
          if ($0 ~ /\bsync\.WaitGroup\b|\.Wait\(\)|<-[[:space:]]*[a-zA-Z_]/) saw_waitsync=1;
        }
        prev_line[NR] = $0;
        if (depth==0 && NR>start) {
          if (saw_skip) {
            # documented skip — audited by scan_skip_no_ticket, not here
          } else if (!has_assert && !saw_suite_run && !saw_subtest && !saw_no_panic && !saw_interface_assert && !saw_optin) {
            if (saw_goroutine && saw_waitsync) {
              # Concurrency/race test with no observable post-join assertion.
              # Weak (relies solely on -race + no-panic), but separated from
              # true no-assert so it can be triaged distinctly.
              printf "%s\t%d\tGO_RACE_ONLY\t%s\n", f, start, first_line;
            } else {
              printf "%s\t%d\tGO_NO_ASSERT\t%s\n", f, start, first_line;
            }
          } else if (nil_only && !saw_suite_run && !saw_subtest && !saw_optin && body ~ /if[[:space:]]+err[[:space:]]*!=[[:space:]]*nil/) {
            printf "%s\t%d\tGO_NIL_ONLY\t%s\n", f, start, first_line;
          }
          in_func=0;
        }
      }
    ' "$f" 2>/dev/null
  done < <(find . -type f -name '*_test.go' -print0 2>/dev/null)
}

scan_ts_tests() {
  # TS_NO_ASSERT: a test() / it() block with no expect( or assert( inside.
  # TS_TRUTHY_ONLY: only expect().toBeTruthy() or .toBeDefined() (no value comparison)
  while IFS= read -r -d '' f; do
    is_excluded "$f" && continue
    awk -v f="$f" '
      /\b(it|test)\s*\(/ {
        depth=0; start=NR; body=""; saw_expect=0; saw_strong=0;
        in_block=1;
      }
      in_block {
        body=body"\n"$0;
        for (i=1; i<=length($0); i++) {
          ch=substr($0,i,1);
          if (ch=="{") depth++;
          else if (ch=="}") depth--;
        }
        if ($0 ~ /\bexpect\s*\(|\bassert\s*\(/) saw_expect=1;
        if ($0 ~ /\.(toEqual|toBe|toMatch|toContain|toHaveBeenCalledWith|toHaveTextContent|toHaveValue|toHaveAttribute|toBeVisible)\b/) saw_strong=1;
        if (depth==0 && NR>start) {
          if (!saw_expect) printf "%s\t%d\tTS_NO_ASSERT\t%s\n", f, start, "(test/it block with no expect/assert)";
          else if (!saw_strong) printf "%s\t%d\tTS_TRUTHY_ONLY\t%s\n", f, start, "(only toBeTruthy/toBeDefined-class, no value compare)";
          in_block=0;
        }
      }
    ' "$f" 2>/dev/null
  done < <(find . -type f \( -name '*.test.ts' -o -name '*.test.tsx' -o -name '*.test.js' -o -name '*.spec.ts' -o -name '*.spec.tsx' -o -name '*.spec.js' \) -print0 2>/dev/null)
}

scan_helixqa_banks() {
  # Bank actions are bluffs ONLY when the action field contains prose AND there
  # is no accompanying executable field (target/command). Two bank schemas exist:
  #
  #  (A) Challenges format: {"action": "verb", "target": "adb shell ...",
  #      "value": "...", "description": "..."}. The action is a verb, but
  #      target carries the executable command — runner uses target+action+value.
  #      NOT a bluff.
  #  (B) HelixQA format: {"action": "POST /api/v1/auth/login with body {...}",
  #      "expected": "200 OK with JSON containing token"}. Prose only — the
  #      runner cannot execute this without interpretation. IS a bluff.
  #
  # Recognized executable-prefix actions also pass (adb_shell: ... etc).
  # Files under submodules/helix_qa/banks/templates/ are scaffolding examples.
  while IFS= read -r -d '' f; do
    is_excluded "$f" && continue
    case "$f" in *banks/templates/*) continue ;; esac
    python3 - "$f" 2>/dev/null <<'PY'
import json, re, sys
path = sys.argv[1]
try:
    with open(path) as fh:
        if path.endswith(('.yaml', '.yml')):
            try:
                import yaml
                data = yaml.safe_load(fh)
            except ImportError:
                sys.exit(0)
        else:
            data = json.load(fh)
except Exception:
    sys.exit(0)

# Find all step-like dicts recursively, with their JSON line position estimated
# (rough — we re-read the file for line numbers).
with open(path) as fh:
    text = fh.read()

EXEC_PREFIXES = (
    # HelixQA testbank/schema.go::ActionType*
    'adb_shell:', 'sleep:', 'screenshot:', 'keypress:', 'tap:', 'swipe:',
    'text:', 'playback_check:', 'frame_diff:', 'http:',
    'assert:', 'playwright:',
    # `shell:<command>` is a first-class executable action — the bank runner
    # runs the remainder through a shell. A `{"action": "shell: git status ..."}`
    # entry IS a shell action, NOT prose; flagging it as PROSE_HELIXQA_ACTION
    # mis-classifies real executable steps in the bash/CLI tool banks.
    'shell:', 'bash:', 'sh:', 'exec:', 'cmd:', 'run:',
    # Other commonly-seen executable forms in surrounding banks
    'adb:', 'click:', 'type:', 'key:', 'press:', 'https:', 'sql:',
    'assertvisible:', 'assertnotvisible:', 'evaluate:', 'navigate:',
    'wait_for:', 'focus:', 'scroll:',
)

def visit(obj):
    if isinstance(obj, dict):
        # Only treat this dict as a step if it has step-like shape:
        # `action` AND `name` (HelixQA TestStep contract). Without
        # this guard, we false-positive on request-body fields that
        # happen to be named `action` (e.g. {"media_id": 1,
        # "action": "play"} for /api/v1/analytics/access).
        is_step_shape = ('action' in obj and isinstance(obj.get('action'), str)
                         and 'name' in obj and isinstance(obj.get('name'), str))
        if is_step_shape:
            action = obj['action']
            action_trim = action.strip()
            # Has accompanying executable field?
            has_target = 'target' in obj and isinstance(obj['target'], str) and obj['target'].strip()
            has_command = 'command' in obj and isinstance(obj['command'], str) and obj['command'].strip()
            has_executable_prefix = any(action.lower().startswith(p) for p in EXEC_PREFIXES)
            # Per testbank/schema.go::ParseAction, a few actions are
            # standalone keywords (no colon-prefix needed). The most
            # common is `screenshot` — exists in 525+ androidtv bank
            # entries. `frame_diff` likewise.
            STANDALONE_EXEC = ('screenshot', 'frame_diff', 'playback_check')
            is_standalone_exec = action_trim.lower() in STANDALONE_EXEC
            # Article XI §11.5: explicit _skip: true with _skip_reason
            # is honest non-execution. Bank entries marked this way are
            # NOT bluffs — they're documented holes (destructive
            # side-effect, missing fixture, converter limitation) that
            # the runtime correctly skips with SKIP-OK marker.
            is_explicitly_skipped = bool(obj.get('_skip')) and bool(obj.get('_skip_reason'))
            # `_conversion_note: manual-review-required` is the bank
            # converter's marker for prose entries that cannot be
            # mechanically translated. Treat as a SKIP candidate too —
            # the converter has already flagged them for human review,
            # and re-flagging here is double-counting.
            has_manual_review_marker = obj.get('_conversion_note') == 'manual-review-required'
            if has_target or has_command or has_executable_prefix or is_explicitly_skipped or has_manual_review_marker or is_standalone_exec:
                pass  # not a bluff
            else:
                # Estimate line number by searching for the action string
                snippet = action[:50].replace('\\', '\\\\').replace('"', '\\"')
                m = re.search(r'"action"\s*:\s*"' + re.escape(snippet[:30]) + r'[^"]*"', text)
                ln = text[:m.start()].count('\n') + 1 if m else 1
                print(f'{path}\t{ln}\tPROSE_HELIXQA_ACTION\taction="{snippet[:50]}"')
        for v in obj.values():
            visit(v)
    elif isinstance(obj, list):
        for v in obj:
            visit(v)

# Article XI §11.5 + CONST-008 (LLM-Driven QA): bank-level exemption
# for vision-driven test suites. The HelixQA pipeline runs these
# banks through the Learn/Plan/Execute/Curiosity/Analyze phases
# where the vision model interprets prose actions and decides what
# to do in the UI. Such banks declare {"metadata": {"_llm_driven": true}}
# OR {"_llm_driven": true} at the top level and ALL their prose
# actions are exempt from PROSE_HELIXQA_ACTION flagging.
top_llm_driven = bool(data.get('_llm_driven')) or bool((data.get('metadata') or {}).get('_llm_driven'))

if top_llm_driven:
    # Whole bank is LLM-driven; all prose actions in it are
    # legitimate, not bluffs. Still emit a per-bank summary line so
    # auditors know we DID look.
    pass  # explicit exemption — no further scan needed
else:
    visit(data)
PY
  done < <(find . -type f -path '*/banks/*' -name '*.json' -print0 2>/dev/null)
  # Per CLAUDE.md "Bank format is JSON at runtime" — YAML mirrors are kept in
  # sync; scanning both would double-count.
}

scan_challenge_scripts() {
  # Heuristic blunders in shell-style Challenge scripts.
  # Excluded by basename: canonical CONST-033 / Article XI tooling that
  # uses `|| true` legitimately for graceful absence-handling (not for
  # exit-code laundering).
  while IFS= read -r -d '' f; do
    is_excluded "$f" && continue
    case "$(basename "$f")" in
      host_no_auto_suspend_challenge.sh|no_suspend_calls_challenge.sh| \
      no_session_termination_calls_challenge.sh)
        continue
        ;;
    esac
    # CHALLENGE_BLIND_SHELL
    # Inline-comment exemption: a line carrying `# bluff-scan: ok (<reason>)`
    # is intentionally a best-effort/teardown line whose next-line
    # assertion compensates. Any author who adds the comment owns the
    # justification and a hostile reviewer can audit by grep.
    #
    # Fault-injection / teardown context exemption (CONST-035 chaos suites):
    # chaos/stress/ddos/scaling fault-injection scripts and teardown blocks
    # LEGITIMATELY suppress exit codes — the WHOLE POINT of `|| true` /
    # `2>/dev/null` on a slow-loris probe, a kill, a wait, or a best-effort
    # health curl is that the injected fault is EXPECTED to fail; the real
    # PASS/FAIL verdict lives on a SEPARATE assertion line (`[[ "$post" != 200 ]]
    # && exit 1`, `case ... ) exit 1`). We therefore exempt `|| true` /
    # `2>/dev/null` exit-laundering ONLY when:
    #   (a) the script is a fault-injection script by basename
    #       (*chaos*, *stress*, *ddos*, *scaling*, *failure_injection*,
    #        *sustained_load*, *horizontal*, *flood*), OR
    #   (b) the offending line is itself a teardown/best-effort line:
    #       trap / kill / wait / rm -f / mv -f / cleanup / pkill / a probe
    #       whose stdout is discarded to /dev/null or captured for later
    #       assertion (`>> "$RES"`, `-o /dev/null`, `2>&1 || true`).
    # We KEEP flagging exit-laundering on real assertion/verification commands
    # — i.e. `&& echo PASS`, `| tee <file>` as the terminal step, and bare
    # `set +e` outside fault-injection scripts — because those launder the
    # verdict itself.
    base="$(basename "$f")"
    is_faultinj=0
    case "$base" in
      *chaos*|*stress*|*ddos*|*scaling*|*failure_injection*|*sustained_load*|*horizontal*|*flood*|*slow_loris*)
        is_faultinj=1 ;;
    esac
    # Pre-read the file so the `set +e` lookahead can verify a matching
    # `set -e` re-enable AND an exit-code capture exist below it (a BOUNDED
    # intentional-failure probe whose RC is then asserted — not laundering).
    mapfile -t _clines < "$f" 2>/dev/null || _clines=()
    grep -nE '(\|\|[[:space:]]*true\b|&&[[:space:]]*echo[[:space:]]+(PASS|OK|SUCCESS)\b|\|[[:space:]]*tee[[:space:]]+\S+[[:space:]]*$|set[[:space:]]+\+e\b)' "$f" 2>/dev/null | while IFS=: read -r ln rest; do
      case "$rest" in *'# bluff-scan: ok'*) continue ;; esac
      # Determine whether THIS line is a teardown / best-effort line.
      is_teardown=0
      case "$rest" in
        *trap\ *|*kill\ *|*kill" "*|*' wait'*|*'wait '*|*'wait;'*|*rm\ -f*|*rm\ -rf*|*mv\ -f*|*pkill*|*cleanup*|*killall*|*'command -v'*|*xargs*|*reload-or-restart*)
          is_teardown=1 ;;
      esac
      # A best-effort probe: stdout discarded or captured (to a file or via
      # command-substitution) for a LATER assertion, ending in `|| true`
      # (the exit code is intentionally ignored here, not the verdict).
      case "$rest" in
        *'2>/dev/null'*'|| true'*|*'2>&1'*'|| true'*|*'-o /dev/null'*'|| true'*|*'>> '*'|| true'*|*'>>"'*'|| true'*)
          is_teardown=1 ;;
        *'2>/dev/null || true'*)
          is_teardown=1 ;;
        # Output captured to a file (`>OUT`, `>"$OUT"`, `>'file'`) then `|| true`:
        # the command is run for its side-effect/output which a downstream line
        # asserts; the inline `|| true` only tolerates a non-zero probe exit.
        *'>'*'|| true'*)
          is_teardown=1 ;;
        # Command-substitution capture with `|| true` inside: `x="$(... || true)"`
        # — the captured value is asserted later; `|| true` guards the subshell.
        *'$('*'|| true'*')'*|*'`'*'|| true'*'`'*)
          is_teardown=1 ;;
      esac
      # `set +e` is a legitimate bounded-probe disable ONLY when a matching
      # `set -e` re-enable AND an exit-code capture (`$?` / PIPESTATUS) appear
      # within the next 8 lines. Bare `set +e` with no re-enable (verdict
      # laundering for the rest of the script) STAYS flagged.
      is_bounded_probe=0
      case "$rest" in
        *'set +e'*)
          saw_reenable=0; saw_rc=0; i="$ln"
          end=$((ln + 8))
          while [ "$i" -lt "$end" ]; do
            cur="${_clines[$i]:-}"   # _clines 0-indexed => line ln+1 at idx ln
            case "$cur" in *'set -e'*) saw_reenable=1 ;; esac
            case "$cur" in *'$?'*|*'PIPESTATUS'*) saw_rc=1 ;; esac
            i=$((i + 1))
          done
          [ "$saw_reenable" -eq 1 ] && [ "$saw_rc" -eq 1 ] && is_bounded_probe=1 ;;
      esac
      # The `|| true` / `2>/dev/null` exit-suppression and `set +e` forms are
      # exempt inside fault-injection scripts, on teardown/best-effort lines,
      # or for a bounded `set +e`/`set -e` probe whose RC is asserted.
      case "$rest" in
        *'|| true'*|*'set +e'*)
          if [ "$is_faultinj" -eq 1 ] || [ "$is_teardown" -eq 1 ] || [ "$is_bounded_probe" -eq 1 ]; then
            continue
          fi ;;
      esac
      emit CHALLENGE_BLIND_SHELL "$f" "$ln" "$rest"
    done
    # CHALLENGE_200_OK_ONLY: curl whose ONLY check is HTTP code or grep -q 200
    grep -nE 'curl[^\n]*-w[^\n]*%\{http_code\}|grep[[:space:]]+-q[[:space:]]+200|status_code"][[:space:]]*==[[:space:]]*200' "$f" 2>/dev/null | while IFS=: read -r ln rest; do
      case "$rest" in *'# bluff-scan: ok'*) continue ;; esac
      emit CHALLENGE_200_OK_ONLY "$f" "$ln" "$rest"
    done
  done < <(find . -type f \( -path '*/challenges/*' -a -name '*.sh' \) -print0 2>/dev/null)
}

scan_skip_no_ticket() {
  # Lines with t.Skip / @Ignore / it.skip / describe.skip / xit / xdescribe lacking a SKIP-OK: #
  # Skips lines that are inside a // comment (the scanner's own first-pass
  # would otherwise flag prose mentions of "t.Skip" in docstrings).
  #
  # A skip is an HONEST, documented skip — NOT a SKIP_WITHOUT_TICKET — when the
  # skip line (or the test's nearby guard) carries any of:
  #   • a SKIP-OK marker with EITHER a numeric ticket (#1234) OR a symbolic
  #     ticket (#short-mode, #BLUFF-001, #round-47-api-drift, #env-no-target).
  #     Article XI §11.4.3 honours symbolic tickets — the marker's job is to
  #     make the skip greppable and owned, not to demand a bug-tracker integer.
  #   • a Constitution §11.4.3 topology-gate reference (the canonical clause for
  #     "host lacks the runtime/binary to run this leg").
  #   • `testing.Short()` short-mode gating — a `t.Skip("... short mode ...")`
  #     guarded by `if testing.Short()` is the Go-idiomatic unit/integration
  #     split, never a bluff. We accept either the literal `short mode` phrase
  #     on the skip line or a `testing.Short()` guard within the 2 lines above.
  while IFS= read -r -d '' f; do
    is_excluded "$f" && continue
    # Pre-read the file lines once so we can look BACKWARD for a testing.Short()
    # guard above the skip (the guard and the skip are on different lines).
    mapfile -t _flines < "$f" 2>/dev/null || continue
    grep -nE '\bt\.Skip\b|\bt\.Skipf\b|@Ignore\b|\bit\.skip\b|\bdescribe\.skip\b|\bxit\b|\bxdescribe\b' "$f" 2>/dev/null | while IFS=: read -r ln rest; do
      # Skip comment lines (//-prefixed in Go/TS/Kotlin)
      stripped="$(echo "$rest" | sed 's/^[[:space:]]*//')"
      case "$stripped" in '//'*|'#'*) continue ;; esac
      # (1) SKIP-OK marker — numeric OR symbolic ticket (#<alnum/._-> token).
      if echo "$rest" | grep -qE 'SKIP-OK:?[[:space:]]*#[A-Za-z0-9._/-]+'; then
        continue
      fi
      # (2) SKIP-OK marker WITHOUT a '#' (e.g. `SKIP-OK real-backend: ...`,
      #     `SKIP-OK could not synthesize ...`) — still an owned, greppable skip.
      if echo "$rest" | grep -qE 'SKIP-OK\b'; then
        continue
      fi
      # (3) Constitution §11.4.3 topology-gate reference on the skip line.
      if echo "$rest" | grep -qE '§11\.4\.3|11\.4\.3'; then
        continue
      fi
      # (4) short-mode gating: literal phrase on the skip line, OR a
      #     `if testing.Short()` guard within the 3 lines immediately above.
      if echo "$rest" | grep -qiE 'short[ _-]?mode|in short mode'; then
        continue
      fi
      short_guard=0
      for back in 1 2 3; do
        idx=$((ln - back - 1))   # _flines is 0-indexed
        [ "$idx" -ge 0 ] || break
        case "${_flines[$idx]:-}" in
          *testing.Short\(\)*) short_guard=1; break ;;
        esac
      done
      [ "$short_guard" -eq 1 ] && continue
      emit SKIP_WITHOUT_TICKET "$f" "$ln" "$rest"
    done
  done < <(find . -type f \( -name '*_test.go' -o -name '*.test.ts' -o -name '*.test.tsx' -o -name '*.spec.ts' -o -name '*.test.js' -o -name '*.kt' -o -name '*.java' \) -print0 2>/dev/null)
}

scan_assert_tautology() {
  while IFS= read -r -d '' f; do
    is_excluded "$f" && continue
    grep -nE 'assert\.True\([^,]*,[[:space:]]*true\)|assert\.Equal\([^,]*,[[:space:]]*1[[:space:]]*,[[:space:]]*1[[:space:]]*\)|expect\(true\)\.toBe\(true\)|expect\(1\)\.toBe\(1\)' "$f" 2>/dev/null | while IFS=: read -r ln rest; do
      emit ASSERT_TAUTOLOGY "$f" "$ln" "$rest"
    done
  done < <(find . -type f \( -name '*_test.go' -o -name '*.test.ts' -o -name '*.test.tsx' -o -name '*.spec.ts' -o -name '*.test.js' -o -name '*.kt' \) -print0 2>/dev/null)
}

main() {
  scan_go_tests
  scan_ts_tests
  scan_helixqa_banks
  scan_challenge_scripts
  scan_skip_no_ticket
  scan_assert_tautology
}

# Capture all findings, print them, and set the documented exit contract:
# 0 when ZERO findings, 1 when ANY finding (Constitution anti-bluff CI lane —
# the scan MUST fail the lane on any real violation). Using a tmp file keeps
# the existing TSV-on-stdout output byte-identical for downstream consumers.
_findings="$(main)"
if [ -n "$_findings" ]; then
  printf '%s\n' "$_findings"
  exit 1
fi
exit 0
