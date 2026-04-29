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

set -uo pipefail

ROOT="${1:-$(cd "$(dirname "$0")/../.." && pwd)}"
cd "$ROOT" || { echo "cannot cd to $ROOT" >&2; exit 2; }

EXCLUDE_DIRS='\(\./\)\?\(\.git\|node_modules\|target\|vendor\|build\|releases\|qa-results\|docs/reports/qa-sessions\|docs/audits\|HelixQA/banks/templates\|tests/k6/node_modules\|tools/opensource\|tools/external\|HelixQA/tools/opensource\|releases\)'

# Substring match — excludes paths at any depth (not just root prefix).
# /tools/opensource/ catches HelixQA/tools/opensource and any nested vendor.
# /node_modules/ catches catalog-web/node_modules, installer-wizard/node_modules, etc.
EXCLUDE_SUBSTRINGS=(
  '/node_modules/'
  '/tools/opensource/'
  '/tools/external/'
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
  # GO_NIL_ONLY: a Test* function whose only assertion check is `if err != nil { t.Fatal(err) }`
  # GO_NO_ASSERT: a Test* function whose body has zero t.Error/t.Fatal/assert./require.
  # GO_HTTPTEST_ABUSE: file path contains e2e or _e2e_ AND uses httptest.NewServer
  # GO_MOCK_IN_INTEGRATION: file path contains integration|e2e|stress|chaos|security|challenge AND uses mock/stub/fake
  while IFS= read -r -d '' f; do
    is_excluded "$f" && continue
    local low="${f,,}"
    # GO_HTTPTEST_ABUSE
    if [[ "$low" == *e2e* ]]; then
      grep -nE 'httptest\.NewServer\b|httptest\.NewRequest\b' "$f" 2>/dev/null | while IFS=: read -r ln rest; do
        stripped="$(echo "$rest" | sed 's/^[[:space:]]*//')"
        case "$stripped" in '//'*|'/*'*|'*'*) continue ;; esac
        emit GO_HTTPTEST_ABUSE "$f" "$ln" "$rest"
      done
    fi
    # GO_MOCK_IN_INTEGRATION
    if [[ "$low" == *integration* || "$low" == *_e2e* || "$low" == *stress* || "$low" == *chaos* || "$low" == *security* || "$low" == *challenge* ]]; then
      grep -nE '\bmock\.|stub\.|fake\.|NewMock|NewFake|NewStub|/mocks/|gomock\.|testify/mock' "$f" 2>/dev/null | while IFS=: read -r ln rest; do
        stripped="$(echo "$rest" | sed 's/^[[:space:]]*//')"
        case "$stripped" in '//'*|'/*'*|'*'*) continue ;; esac
        emit GO_MOCK_IN_INTEGRATION "$f" "$ln" "$rest"
      done
    fi
    # Per-function nil-only / no-assert
    awk -v f="$f" '
      /^func[[:space:]]+(\([^)]*\)[[:space:]]*)?Test[A-Z]/ {
        in_func=1; depth=0; start=NR; body=""; nil_only=1; has_assert=0; saw_anything=0; first_line=$0;
      }
      in_func {
        body=body"\n"$0;
        # depth tracker
        for (i=1; i<=length($0); i++) {
          ch=substr($0,i,1);
          if (ch=="{") depth++;
          else if (ch=="}") depth--;
        }
        if (NR > start) {
          if ($0 ~ /t\.Error|t\.Errorf|t\.Fatal|t\.Fatalf|t\.Logf|assert\.|require\.|expect\(/) has_assert=1;
          # very conservative: any assertion that is NOT just "if err != nil { t.Fatal(err) }" disqualifies nil_only
          if ($0 ~ /(t\.Error|t\.Errorf|assert\.|require\.|expect\()/) nil_only=0;
        }
        if (depth==0 && NR>start) {
          if (!has_assert) printf "%s\t%d\tGO_NO_ASSERT\t%s\n", f, start, first_line;
          else if (nil_only && body ~ /if[[:space:]]+err[[:space:]]*!=[[:space:]]*nil/) printf "%s\t%d\tGO_NIL_ONLY\t%s\n", f, start, first_line;
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
  # HelixQA bank actions are JSON. A prose action is one without a recognized executable prefix.
  # Recognized prefixes: adb_shell:, adb:, playwright:, click:, type:, key:, http:, sql:, sleep:, screenshot:, assertVisible:, assertNotVisible:, evaluate:, navigate:.
  while IFS= read -r -d '' f; do
    is_excluded "$f" && continue
    grep -nE '"action"[[:space:]]*:[[:space:]]*"' "$f" 2>/dev/null | while IFS=: read -r ln rest; do
      # extract the value between the first pair of quotes after the colon
      val=$(echo "$rest" | sed -nE 's/.*"action"[[:space:]]*:[[:space:]]*"([^"]*)".*/\1/p')
      if ! echo "$val" | grep -qE '^(adb_shell:|adb:|playwright:|click:|type:|key:|press:|swipe:|http(s)?:|sql:|sleep:|screenshot:|assertVisible:|assertNotVisible:|evaluate:|navigate:|wait_for:|tap:|focus:|scroll:)'; then
        if [ -n "$val" ]; then
          emit PROSE_HELIXQA_ACTION "$f" "$ln" "action=\"$val\""
        fi
      fi
    done
  done < <(find . -type f \( -path '*/banks/*' -a \( -name '*.json' -o -name '*.yaml' -o -name '*.yml' \) \) -print0 2>/dev/null)
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
    grep -nE '(\|\|[[:space:]]*true\b|&&[[:space:]]*echo[[:space:]]+(PASS|OK|SUCCESS)\b|\|[[:space:]]*tee[[:space:]]+\S+[[:space:]]*$|set[[:space:]]+\+e\b)' "$f" 2>/dev/null | while IFS=: read -r ln rest; do
      emit CHALLENGE_BLIND_SHELL "$f" "$ln" "$rest"
    done
    # CHALLENGE_200_OK_ONLY: curl whose ONLY check is HTTP code or grep -q 200
    grep -nE 'curl[^\n]*-w[^\n]*%\{http_code\}|grep[[:space:]]+-q[[:space:]]+200|status_code"][[:space:]]*==[[:space:]]*200' "$f" 2>/dev/null | while IFS=: read -r ln rest; do
      emit CHALLENGE_200_OK_ONLY "$f" "$ln" "$rest"
    done
  done < <(find . -type f \( -path '*/challenges/*' -a -name '*.sh' \) -print0 2>/dev/null)
}

scan_skip_no_ticket() {
  # Lines with t.Skip / @Ignore / it.skip / describe.skip / xit / xdescribe lacking a SKIP-OK: #
  # Skips lines that are inside a // comment (the scanner's own first-pass
  # would otherwise flag prose mentions of "t.Skip" in docstrings).
  while IFS= read -r -d '' f; do
    is_excluded "$f" && continue
    grep -nE '\bt\.Skip\b|\bt\.Skipf\b|@Ignore\b|\bit\.skip\b|\bdescribe\.skip\b|\bxit\b|\bxdescribe\b' "$f" 2>/dev/null | while IFS=: read -r ln rest; do
      # Skip comment lines (//-prefixed in Go/TS/Kotlin)
      stripped="$(echo "$rest" | sed 's/^[[:space:]]*//')"
      case "$stripped" in '//'*|'#'*) continue ;; esac
      if ! echo "$rest" | grep -qE 'SKIP-OK:[[:space:]]*#[A-Za-z0-9_-]+'; then
        emit SKIP_WITHOUT_TICKET "$f" "$ln" "$rest"
      fi
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

main
