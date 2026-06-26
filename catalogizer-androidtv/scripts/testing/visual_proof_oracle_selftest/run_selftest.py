#!/usr/bin/env python3
"""
§11.4.107(10) self-validation for the §11.4.170 layout oracle.

An analyzer that PASSes its golden-BAD fixture is itself a bluff gate. This guard
runs visual_proof_layout_oracle.py against committed REAL device-capture fixtures
and asserts:
  - golden_good_home.png        -> PASS  (clean home, must not false-FAIL)
  - golden_bad_giant_poster.png -> FAIL  (the §11.4.170 giant-poster signature)

Both fixtures are genuine on-device captures from MIBOX4 (2026-06-26):
the giant-poster detail screen is the real forensic case §11.4.170 was written for.

Exit 0 = oracle is trustworthy; non-zero = oracle is a bluff gate / over-sensitive.
"""
import os, sys, subprocess, json

HERE = os.path.dirname(os.path.abspath(__file__))
ORACLE = os.path.join(HERE, "..", "visual_proof_layout_oracle.py")
FIX = os.path.join(HERE, "fixtures")

CASES = [
    ("golden_good_home.png", "PASS"),        # clean home -> must PASS
    ("golden_bad_giant_poster.png", "FAIL"), # giant poster -> must FAIL
]


def run(img):
    p = subprocess.run([sys.executable, ORACLE, os.path.join(FIX, img), "--json"],
                       capture_output=True, text=True)
    try:
        return json.loads(p.stdout), p.returncode
    except Exception:
        print("ORACLE ERROR on %s:\n%s\n%s" % (img, p.stdout, p.stderr))
        return {"verdict": "ERROR"}, p.returncode


def main():
    ok = True
    for img, expected in CASES:
        res, rc = run(img)
        got = res.get("verdict")
        good = (got == expected)
        ok = ok and good
        extra = ""
        if "checks" in res and "no_giant_unbounded_widget" in res["checks"]:
            extra = " giant_frac=%s" % res["checks"]["no_giant_unbounded_widget"]["largest_featureless_fg_frac"]
        print("[%s] %-28s expected=%s got=%s%s" %
              ("OK" if good else "BLUFF-GATE", img, expected, got, extra))
        if not good:
            print("   findings: %s" % res.get("findings"))
    if ok:
        print("SELFTEST PASS — oracle FAILs golden-bad and PASSes golden-good (§11.4.107(10))")
        sys.exit(0)
    print("SELFTEST FAIL — oracle is a bluff gate or over-sensitive (§11.4.107(10) violation)")
    sys.exit(1)


if __name__ == "__main__":
    main()
