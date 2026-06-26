#!/usr/bin/env bash
###############################################################################
# visual_proof_challenge.sh
#
# Purpose:
#   §11.4.169 / §11.4.170 full-automation "Challenge" wrapper around the
#   committed §11.4.170 host-side rendered-UI visual-proof oracle
#   (visual_proof_layout_oracle.py) and its §11.4.107(10) self-validation
#   (visual_proof_oracle_selftest/run_selftest.py).
#
#   It is anti-bluff (§11.4 / §11.4.6) by construction:
#     1. It FIRST runs the oracle's self-test. If the oracle cannot catch its
#        own golden-BAD fixtures it is a BLUFF GATE — its verdicts cannot be
#        trusted — so the whole challenge ABORTS with a clear FAIL.
#     2. It THEN runs the oracle against committed REAL device-capture frames
#        and scores each frame against an EXPECTED-verdict map (clean home =>
#        PASS expected; known giant-poster *detail* defect => FAIL expected).
#        The OVERALL challenge PASSes only when EVERY frame's ACTUAL verdict
#        matches its EXPECTED verdict — i.e. the oracle correctly PASSes good
#        screens AND correctly FAILs the known giant-poster defect. A frame
#        whose actual != expected is a challenge FAIL (a regression OR oracle
#        drift). The known-defect frame FAILing is the challenge's JOB — it
#        surfaces the §11.4.170 forensic giant-poster screen.
#     3. If the imaging deps (cv2 / pytesseract / tesseract) are ABSENT the
#        oracle exits 2; the challenge then exits 2 = SKIP-with-reason
#        "imaging deps absent" (§11.4.3) — NEVER a fake PASS.
#
# Usage:
#   visual_proof_challenge.sh [DEVICE_FIXTURE_DIR]
#     DEVICE_FIXTURE_DIR  optional. Directory glob'd for *.png REAL device
#                         captures. Default:
#                         <androidtv>/qa-results/device_qa_20260626
#                         When that directory is absent OR contains no *.png,
#                         the challenge HONESTLY falls back to the committed
#                         genuine MIBOX4 device captures shipped as the
#                         oracle's golden fixtures (clean home + giant poster),
#                         and records that the fallback was used (§11.4.6).
#
# Inputs:
#   - visual_proof_layout_oracle.py            (sibling — the §11.4.170 oracle)
#   - visual_proof_oracle_selftest/run_selftest.py (§11.4.107(10) self-test)
#   - *.png frames under DEVICE_FIXTURE_DIR (or the fallback fixtures)
#   - env PYTHON  (optional override of the python3 interpreter)
#
# Outputs:
#   - Structured human-readable verdict to stdout.
#   - Evidence JSON at:
#       <androidtv>/qa-results/visual_proof_challenge/<UTC-timestamp>/result.json
#     containing overall verdict, selftest verdict, per-frame
#     {file, expected, actual, match, key metrics}, and evidence_path (§11.4.69).
#   - Exit code (honest §11.4.3): 0 = PASS, 1 = FAIL, 2 = SKIP / ERROR.
#
# Side-effects:
#   - Creates the evidence directory + result.json under qa-results/ (not
#     git-tracked raw corpus per §11.4.128; curated evidence committed at
#     release-prep per §11.4.83).
#   - Creates a private temp dir under ${TMPDIR}; removed on every exit path
#     via a trap (§11.4.14). Touches NO device, NO emulator, NO network.
#
# Dependencies:
#   - bash (parses clean under sh -n / bash -n per §11.4.67)
#   - python3 (the oracle's interpreter)
#   - the oracle's own runtime deps: numpy, opencv-python (cv2), pytesseract,
#     and the tesseract binary. Their ABSENCE is the honest SKIP path, never a
#     fake PASS.
#
# Cross-references:
#   §11.4.169 (mandatory test-type coverage — this is the Challenge layer),
#   §11.4.170 (host-side rendered-UI visual-proof — the oracle this wraps),
#   §11.4.107(10) (self-validated golden-good/golden-bad analyzer — step 1),
#   §11.4.3 (SKIP-with-reason when deps/topology absent),
#   §11.4.6 (no-guessing — fallback recorded as fact, never silently assumed),
#   §11.4.14 (trap-based cleanup), §11.4.18 (this doc block + companion guide),
#   §11.4.67 (target-shell-parseable), §11.4.69 (captured-evidence path),
#   §11.4.83 (curated qa evidence), §11.4.128 (raw corpus git-ignored).
###############################################################################
set -u

PY="${PYTHON:-python3}"

# ---- resolve canonical paths (no cd; absolute) ------------------------------
SCRIPT_PATH="$0"
case "$SCRIPT_PATH" in
  /*) : ;;
  *)  SCRIPT_PATH="$(pwd)/$SCRIPT_PATH" ;;
esac
TESTING_DIR="$(dirname "$SCRIPT_PATH")"
TESTING_DIR="$(cd "$TESTING_DIR" && pwd)"
ANDROIDTV_ROOT="$(cd "$TESTING_DIR/../.." && pwd)"

ORACLE="$TESTING_DIR/visual_proof_layout_oracle.py"
SELFTEST="$TESTING_DIR/visual_proof_oracle_selftest/run_selftest.py"
SELFTEST_FIX="$TESTING_DIR/visual_proof_oracle_selftest/fixtures"

DEVICE_FIX_DIR="${1:-$ANDROIDTV_ROOT/qa-results/device_qa_20260626}"

TS="$(date -u +%Y%m%dT%H%M%SZ)"
EVIDENCE_DIR="$ANDROIDTV_ROOT/qa-results/visual_proof_challenge/$TS"
RESULT_JSON="$EVIDENCE_DIR/result.json"

# ---- temp dir + trap cleanup (§11.4.14) -------------------------------------
TMP="$(mktemp -d "${TMPDIR:-/tmp}/vpchallenge.XXXXXX")" || {
  echo "ERROR: cannot create temp dir" >&2
  exit 2
}
cleanup() { rm -rf "$TMP" 2>/dev/null || true; }
trap cleanup EXIT INT TERM

echo "==============================================================="
echo "§11.4.170 VISUAL-PROOF CHALLENGE  (§11.4.169 full-automation)"
echo "  oracle      : $ORACLE"
echo "  selftest    : $SELFTEST"
echo "  device dir  : $DEVICE_FIX_DIR"
echo "  evidence    : $RESULT_JSON"
echo "==============================================================="

# ---- sanity: required scripts present ---------------------------------------
if [ ! -f "$ORACLE" ] || [ ! -f "$SELFTEST" ]; then
  echo "ERROR: oracle and/or selftest script missing — cannot run challenge" >&2
  exit 2
fi

# ---- dep probe (§11.4.3): really run the oracle once; exit 2 => SKIP ---------
PROBE_IMG="$SELFTEST_FIX/golden_good_home.png"
if [ ! -f "$PROBE_IMG" ]; then
  echo "ERROR: probe fixture missing: $PROBE_IMG" >&2
  exit 2
fi
"$PY" "$ORACLE" "$PROBE_IMG" --json >"$TMP/probe.json" 2>"$TMP/probe.err"
PROBE_RC=$?
if [ "$PROBE_RC" -eq 2 ]; then
  echo
  echo "SKIP (exit 2): imaging deps absent — oracle exited 2 (§11.4.3 honest skip)"
  echo "  reason: $(cat "$TMP/probe.json" 2>/dev/null; cat "$TMP/probe.err" 2>/dev/null)"
  echo "  NOT a fake PASS: the visual-proof oracle could not run (cv2/pytesseract/tesseract)."
  exit 2
fi

# ---- STEP 1: oracle self-validation (§11.4.107(10)) -------------------------
echo
echo "--- STEP 1: oracle self-validation (golden-good PASS, golden-bad FAIL) ---"
"$PY" "$SELFTEST" >"$TMP/selftest.out" 2>&1
SELFTEST_RC=$?
sed 's/^/  /' "$TMP/selftest.out"
if [ "$SELFTEST_RC" -ne 0 ]; then
  echo
  echo "CHALLENGE ABORT — FAIL (exit 1): the oracle FAILED its own self-test."
  echo "  An oracle that cannot catch its golden-BAD fixtures is a BLUFF GATE;"
  echo "  its PASS/FAIL verdicts on real frames cannot be trusted (§11.4.107(10))."
  # still emit an evidence file documenting the abort
  mkdir -p "$EVIDENCE_DIR" 2>/dev/null || true
  "$PY" - "$TMP" "$SELFTEST_RC" "$TS" "$EVIDENCE_DIR" "ABORTED_SELFTEST" "$DEVICE_FIX_DIR" <<'PYEOF' || true
import sys, os, json, datetime
tmp, st_rc, ts, evdir, source, devdir = sys.argv[1:7]
st_out = ""
p = os.path.join(tmp, "selftest.out")
if os.path.exists(p):
    st_out = open(p).read()
res = {
    "challenge": "visual_proof_challenge",
    "covenant": ["11.4.169", "11.4.170", "11.4.107(10)", "11.4.3"],
    "timestamp_utc": ts,
    "overall": "FAIL",
    "abort_reason": "oracle self-test failed — bluff gate, verdicts untrusted",
    "selftest": {"verdict": "FAIL", "exit_code": int(st_rc), "output": st_out},
    "device_fixture_dir": devdir,
    "fixture_source": source,
    "frames": [],
    "evidence_path": os.path.join(evdir, "result.json"),
}
os.makedirs(evdir, exist_ok=True)
json.dump(res, open(os.path.join(evdir, "result.json"), "w"), indent=2)
print("  evidence written: %s" % os.path.join(evdir, "result.json"))
PYEOF
  exit 1
fi
echo "  STEP 1 OK — oracle is trustworthy (self-test PASS)."

# ---- gather REAL device-capture frames --------------------------------------
# Glob the device fixture dir; honestly fall back to the committed genuine
# MIBOX4 device captures (the oracle's golden fixtures) if that dir is empty.
FIXTURE_SOURCE="device_fixture_dir"
FRAMES=""
if [ -d "$DEVICE_FIX_DIR" ]; then
  for f in "$DEVICE_FIX_DIR"/*.png; do
    [ -f "$f" ] && FRAMES="$FRAMES
$f"
  done
fi
if [ -z "$FRAMES" ]; then
  FIXTURE_SOURCE="fallback_committed_device_captures"
  echo
  echo "  NOTE (§11.4.6): no *.png under $DEVICE_FIX_DIR —"
  echo "  falling back to the committed genuine MIBOX4 device captures"
  echo "  (golden_good_home = clean home; golden_bad_giant_poster = giant-poster detail)."
  for f in "$SELFTEST_FIX/golden_good_home.png" "$SELFTEST_FIX/golden_bad_giant_poster.png"; do
    [ -f "$f" ] && FRAMES="$FRAMES
$f"
  done
fi
if [ -z "$FRAMES" ]; then
  echo
  echo "FAIL (exit 1): no frames found to evaluate (no device captures, no fallback)."
  exit 1
fi

# ---- STEP 2: run oracle on each frame vs an EXPECTED-verdict map -------------
echo
echo "--- STEP 2: oracle vs committed REAL device frames (expected-verdict map) ---"
: > "$TMP/manifest.tsv"
IDX=0
printf '%s\n' "$FRAMES" | while IFS= read -r img; do
  [ -n "$img" ] || continue
  base="$(basename "$img")"
  lc="$(printf '%s' "$base" | tr '[:upper:]' '[:lower:]')"
  # Expected-verdict map: known giant-poster *detail* defect => FAIL expected;
  # clean home / browse screens => PASS expected (§11.4.170 forensic).
  case "$lc" in
    *detail*|*poster*|*giant*|*tvshow*) expected="FAIL" ;;
    *)                                  expected="PASS" ;;
  esac
  "$PY" "$ORACLE" "$img" --json >"$TMP/frame_$IDX.json" 2>"$TMP/frame_$IDX.err"
  rc=$?
  case "$rc" in
    0) actual="PASS" ;;
    1) actual="FAIL" ;;
    *) actual="ERROR" ;;
  esac
  if [ "$actual" = "$expected" ]; then mark="OK"; else mark="MISMATCH"; fi
  printf '[%s] %-44s expected=%s actual=%s\n' "$mark" "$base" "$expected" "$actual"
  printf '%s\t%s\t%s\t%s\n' "$IDX" "$img" "$expected" "$rc" >> "$TMP/manifest.tsv"
  IDX=$((IDX + 1))
done

# ---- STEP 3: aggregate -> structured stdout + result.json (§11.4.69) ---------
mkdir -p "$EVIDENCE_DIR" 2>/dev/null || true
"$PY" - "$TMP" "$SELFTEST_RC" "$TS" "$EVIDENCE_DIR" "$FIXTURE_SOURCE" "$DEVICE_FIX_DIR" <<'PYEOF'
import sys, os, json
tmp, st_rc, ts, evdir, source, devdir = sys.argv[1:7]

def metrics(j):
    out = {}
    ch = j.get("checks", {})
    g = ch.get("no_giant_unbounded_widget", {})
    if g:
        out["largest_featureless_fg_frac"] = g.get("largest_featureless_fg_frac")
    ib = ch.get("no_giant_interior_blank", {})
    if ib:
        out["largest_interior_featureless_frac"] = ib.get("largest_interior_featureless_frac")
    leg = ch.get("legibility", {})
    if leg:
        out["legible_words"] = leg.get("legible_words")
    return out

frames = []
all_match = True
any_error = False
man = os.path.join(tmp, "manifest.tsv")
rows = []
if os.path.exists(man):
    for line in open(man):
        line = line.rstrip("\n")
        if not line:
            continue
        rows.append(line.split("\t"))
rows.sort(key=lambda r: int(r[0]))
for idx, img, expected, rc in rows:
    jpath = os.path.join(tmp, "frame_%s.json" % idx)
    j = {}
    if os.path.exists(jpath):
        try:
            j = json.load(open(jpath))
        except Exception:
            j = {}
    rc = int(rc)
    actual = {0: "PASS", 1: "FAIL"}.get(rc, "ERROR")
    if actual == "ERROR":
        any_error = True
    match = (actual == expected)
    all_match = all_match and match
    frames.append({
        "file": os.path.basename(img),
        "path": img,
        "expected": expected,
        "actual": actual,
        "oracle_verdict": j.get("verdict"),
        "match": match,
        "metrics": metrics(j),
        "findings": j.get("findings", []),
    })

selftest_pass = (int(st_rc) == 0)
st_out = ""
p = os.path.join(tmp, "selftest.out")
if os.path.exists(p):
    st_out = open(p).read()

overall = "PASS" if (selftest_pass and frames and all_match and not any_error) else "FAIL"

res = {
    "challenge": "visual_proof_challenge",
    "covenant": ["11.4.169", "11.4.170", "11.4.107(10)", "11.4.3", "11.4.69"],
    "timestamp_utc": ts,
    "overall": overall,
    "selftest": {
        "verdict": "PASS" if selftest_pass else "FAIL",
        "exit_code": int(st_rc),
        "output": st_out,
    },
    "device_fixture_dir": devdir,
    "fixture_source": source,
    "frame_count": len(frames),
    "frames": frames,
    "evidence_path": os.path.join(evdir, "result.json"),
}
os.makedirs(evdir, exist_ok=True)
json.dump(res, open(os.path.join(evdir, "result.json"), "w"), indent=2)

print("")
print("--- STEP 3: structured verdict ---")
print("  selftest        : %s (exit %s)" % (res["selftest"]["verdict"], res["selftest"]["exit_code"]))
print("  fixture source  : %s" % source)
print("  frames evaluated: %d" % len(frames))
for fr in frames:
    print("    [%s] %-44s expected=%s actual=%s  %s"
          % ("OK" if fr["match"] else "MISMATCH", fr["file"],
             fr["expected"], fr["actual"], fr["metrics"]))
print("  evidence_path   : %s" % res["evidence_path"])
print("")
print("OVERALL: %s" % overall)
if overall != "PASS":
    if not selftest_pass:
        print("  cause: oracle self-test FAILED (bluff gate).")
    if any_error:
        print("  cause: one or more frames returned ERROR (oracle could not evaluate).")
    if frames and not all_match:
        print("  cause: a frame's actual verdict != expected (regression OR oracle drift).")
sys.exit(0 if overall == "PASS" else 1)
PYEOF
AGG_RC=$?

echo "==============================================================="
exit "$AGG_RC"
