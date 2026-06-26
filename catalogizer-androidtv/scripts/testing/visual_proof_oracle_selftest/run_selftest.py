#!/usr/bin/env python3
"""
§11.4.107(10) self-validation for the §11.4.170 layout oracle.

An analyzer that PASSes its golden-BAD fixture is itself a bluff gate. This guard
runs visual_proof_layout_oracle.py against committed fixtures and asserts:
  - golden_good_home.png                  -> PASS  (clean home, must not false-FAIL)
  - golden_bad_giant_poster.png           -> FAIL  (the §11.4.170 giant-poster signature)
  - golden_bad_interior_blank_bgcolor.png -> FAIL  (the L1 evasion: a giant blank box
                                                    BELOW the fg brightness cutoff that
                                                    the foreground check alone misses)
  - golden_bad_wide_top_placeholder.png   -> FAIL  (the L4 escape: a real MIBOX4 frame with a
                                                    full-width gray backdrop band across the
                                                    TOP whose AREA frac (~0.10) is below the
                                                    giant threshold and which hugs the top
                                                    border, so it evaded checks #2 AND #3)

golden_good_home, golden_bad_giant_poster and golden_bad_wide_top_placeholder are genuine
on-device captures from MIBOX4 (2026-06-26): the giant-poster detail screen is the real
forensic case §11.4.170 was written for, and golden_bad_wide_top_placeholder is the real
§11.4.138 operator-escape frame (a defect frame that PASSed the oracle — now caught by the
no_wide_featureless_band check). golden_bad_interior_blank_bgcolor is synthesized by
make_interior_blank_fixture.py to reproduce review finding L1 exactly — the foreground
giant-widget check sees frac=0.0 (would PASS) while the interior-blank check FAILs it,
proving each added check genuinely closes its gap (not by weakening the existing checks).

Exit 0 = oracle is trustworthy; non-zero = oracle is a bluff gate / over-sensitive.
"""
import os, sys, subprocess, json

HERE = os.path.dirname(os.path.abspath(__file__))
ORACLE = os.path.join(HERE, "..", "visual_proof_layout_oracle.py")
FIX = os.path.join(HERE, "fixtures")

CASES = [
    ("golden_good_home.png", "PASS"),                      # clean home -> must PASS
    ("golden_bad_giant_poster.png", "FAIL"),               # giant poster -> must FAIL
    ("golden_bad_interior_blank_bgcolor.png", "FAIL"),     # L1 evasion blank -> must FAIL
    ("golden_bad_wide_top_placeholder.png", "FAIL"),       # L4 escape: real MIBOX4 wide top gray band -> must FAIL
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
        ch = res.get("checks", {})
        if "no_giant_unbounded_widget" in ch:
            extra += " giant_fg_frac=%s" % ch["no_giant_unbounded_widget"]["largest_featureless_fg_frac"]
        if "no_giant_interior_blank" in ch:
            extra += " interior_frac=%s" % ch["no_giant_interior_blank"]["largest_interior_featureless_frac"]
        if "no_wide_featureless_band" in ch:
            extra += " band_wspan=%s band_hspan=%s" % (
                ch["no_wide_featureless_band"]["band_width_span"],
                ch["no_wide_featureless_band"]["band_height_span"])
        print("[%s] %-39s expected=%s got=%s%s" %
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
