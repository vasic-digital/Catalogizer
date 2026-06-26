# visual_proof_challenge.sh

**Revision:** 1
**Last modified:** 2026-06-26T08:30:00Z

## Overview

`visual_proof_challenge.sh` is the §11.4.169 full-automation **Challenge**
wrapper around the committed §11.4.170 host-side rendered-UI visual-proof
oracle. It does NOT re-implement any vision logic — it orchestrates two
already-committed, already-tested scripts and turns them into a single
re-runnable PASS/FAIL/SKIP challenge with a structured evidence artefact:

- `catalogizer-androidtv/scripts/testing/visual_proof_layout_oracle.py`
  — the §11.4.170 OCR/vision layout oracle (cv2 + pytesseract). Exit `0`=PASS,
  `1`=FAIL, `2`=ERROR (honest tool-absent).
- `catalogizer-androidtv/scripts/testing/visual_proof_oracle_selftest/run_selftest.py`
  — the §11.4.107(10) self-validation (golden-good PASS, two golden-bad FAIL).

The challenge is anti-bluff by construction: it cannot report PASS unless the
oracle FIRST proves it can catch its own golden-bad fixtures, AND THEN every
real device frame matches its expected verdict.

## Prerequisites

- `bash`, `python3`.
- The oracle's runtime deps: `numpy`, `opencv-python` (`cv2`), `pytesseract`,
  and the `tesseract` binary. Their **absence is the honest SKIP path**
  (exit 2, §11.4.3) — never a fake PASS.
- The committed oracle + selftest + golden fixtures (shipped in-repo).

## Usage

```bash
# default: glob qa-results/device_qa_20260626/*.png, else fall back to the
# committed genuine MIBOX4 device captures
catalogizer-androidtv/scripts/testing/visual_proof_challenge.sh

# explicit device-capture directory
catalogizer-androidtv/scripts/testing/visual_proof_challenge.sh /path/to/device_qa_dir

# override interpreter
PYTHON=python3.11 catalogizer-androidtv/scripts/testing/visual_proof_challenge.sh
```

Exit codes (honest, §11.4.3): `0`=PASS, `1`=FAIL, `2`=SKIP/ERROR.

Evidence JSON is written to
`catalogizer-androidtv/qa-results/visual_proof_challenge/<UTC-timestamp>/result.json`
(§11.4.69): overall verdict, selftest verdict + output, `fixture_source`, and a
per-frame array of `{file, expected, actual, oracle_verdict, match, metrics,
findings}`.

## Internal behaviour

1. **Dep probe** — really runs the oracle once on a golden fixture. If it
   exits 2 (cv2/pytesseract/tesseract absent) the challenge exits 2 = SKIP with
   reason. No fake PASS.
2. **STEP 1 — self-validation** — runs `run_selftest.py`. If the oracle does
   not PASS its golden-good and FAIL both golden-bad fixtures it is a **bluff
   gate**; the challenge ABORTS with FAIL (exit 1) and still writes a
   `result.json` documenting the abort. The oracle's verdicts cannot be trusted
   on real frames if it cannot catch its own golden-bad.
3. **Frame gathering** — globs the device-fixture dir for `*.png`. When that
   directory is absent or empty it **honestly falls back** to the committed
   genuine MIBOX4 captures (the golden fixtures: clean home + giant-poster
   detail) and records `fixture_source=fallback_committed_device_captures`
   (§11.4.6 — the fallback is recorded as fact, never silently assumed).
4. **STEP 2 — expected-verdict map** — each frame is classified by filename:
   `*detail* | *poster* | *giant* | *tvshow*` ⇒ **FAIL expected** (the known
   §11.4.170 giant-poster defect); everything else ⇒ **PASS expected** (clean
   home / browse). The oracle is run on each frame; ACTUAL is compared to
   EXPECTED. The known-defect frame FAILing is the challenge's JOB — it surfaces
   the giant-poster screen.
5. **STEP 3 — aggregate** — overall PASS iff selftest PASSed AND every frame's
   actual verdict equals its expected verdict AND no frame ERRORed. A
   `actual != expected` frame is a challenge FAIL (a real regression OR oracle
   drift).
6. **Cleanup** — a `trap … EXIT INT TERM` removes the private temp dir on every
   exit path (§11.4.14). The challenge touches NO device, NO emulator, NO
   network.

## Edge cases

- **Imaging deps absent** → exit 2 SKIP, reason printed, no result.json claim
  of PASS.
- **Oracle self-test fails** → exit 1 FAIL (abort), `result.json` records the
  abort + selftest output.
- **No frames at all** (no device dir, fallback fixtures missing) → exit 1 FAIL
  ("no frames to evaluate") — absence of evidence is never a PASS.
- **Per-frame ERROR** (unreadable/corrupt PNG present in the device dir) →
  counts as a challenge FAIL (the frame could not be evaluated).
- **Filenames** are classified case-insensitively.

## Related scripts

- `catalogizer-androidtv/scripts/testing/visual_proof_layout_oracle.py`
- `catalogizer-androidtv/scripts/testing/visual_proof_oracle_selftest/run_selftest.py`
- `catalogizer-androidtv/scripts/testing/visual_proof_oracle_selftest/make_interior_blank_fixture.py`

## Last verified date

2026-06-26 — ran on host: selftest PASS (3/3 fixtures), 2 real MIBOX4 device
captures scored against the expected-verdict map (clean home PASS, giant-poster
detail FAIL), overall PASS (exit 0). SKIP path verified by shadowing `cv2`
(oracle exit 2 → challenge exit 2). FAIL path verified by a mismatched
expected-verdict frame (exit 1).
