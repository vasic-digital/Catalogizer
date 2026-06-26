#!/usr/bin/env python3
"""
§11.4.170 dual-oracle (ii): OCR / vision LAYOUT oracle for host-rendered UI PNGs.

Reads a rendered PNG (host-side render OR on-device screencap) and asserts the
§11.4.162 layout invariants ON REAL PIXELS — the proof a value/token-equality
test can never give:
  - legibility:        OCR text runs present + per-word confidence >= floor
  - no giant/unbounded: no single featureless foreground region exceeds a
                        calibrated fraction of the canvas (the giant-button /
                        giant-poster-placeholder signature, §11.4.170 forensic)
  - no label overlap:  no two OCR text boxes overlap (label-over-label)
  - on-screen:         text boxes lie within the canvas (no clipping/off-screen)

Emits a structured JSON verdict (PASS/FAIL + per-check + metrics + pinpoint).
Thresholds are calibrated on THIS project's own fixtures (§11.4.6), recorded in
THRESHOLDS below with the calibrating frames.

Self-validation (§11.4.107(10)) lives in visual_proof_oracle_selftest.py:
a golden-GOOD frame MUST PASS and golden-BAD frames MUST FAIL, else the oracle
is itself a bluff gate.

Usage:
    python3 visual_proof_layout_oracle.py <image.png> [--json] [--min-conf N]
Exit: 0 = PASS, 1 = FAIL, 2 = ERROR (unreadable / tool absent — honest §11.4.3).

HONEST LIMITATIONS (§11.4.6 — an anti-bluff analyzer MUST name what it cannot catch;
confirmed by independent review 2026-06-26):
  L1 (giant-widget blind spot): the giant-unbounded-widget detector keys on a
     featureless foreground region BRIGHTER than the dark page background
     (gray > bg_brightness_max). A giant featureless placeholder rendered in the
     EXACT background color (gray <= bg_brightness_max) on an otherwise-legible
     screen EVADES this check (frac=0.0). Narrow in practice — real Compose
     unloaded/placeholder surfaces are mid-gray (>=45) and ARE caught — but a
     near-background-colored giant blank is not. FOLLOW-UP: add an interior-blank
     check (giant featureless region NOT touching the canvas border, even at bg
     color, distinguishes a centered placeholder pill from the true background).
  L2 (legibility near-vacuous): min_legible_words=1 makes legibility a weak floor;
     the giant-widget check carries the actual §11.4.170 proof, not legibility.
  L3 (overlap/clip not yet golden-bad-guarded): the no_label_overlap and on_screen
     checks are IMPLEMENTED and logically sound but ship WITHOUT a §11.4.107(10)
     golden-bad fixture (a flaky synthetic overlap fixture was honestly dropped
     rather than shipped as a bluff). FOLLOW-UP: capture a real overlap/clip defect
     frame as a golden-bad and bring these checks under the self-test.
"""
import sys, json, argparse

# Calibrated on this project's real device captures (qa-results/device_qa_20260626/):
#   golden-GOOD  catalogizer---androidtv-retry-005339.png        (clean home)
#   golden-BAD   catalogizer---androidtv-tvshow-detail.png       (Inception giant poster)
# The giant gray poster pill measured ~0.45-0.70 canvas-area; clean home's largest
# featureless foreground blob (a card) measured well under 0.25. Threshold set at 0.32.
THRESHOLDS = {
    "giant_region_area_frac_max": 0.32,  # > this featureless fg fraction => giant-unbounded FAIL
    "ocr_min_conf": 55,                  # per-word tesseract confidence floor
    "min_legible_words": 1,              # at least one legible text run expected
    "bg_brightness_max": 40,             # gray<=this is treated as page background (dark navy ~10-25)
    "featureless_std_max": 12.0,         # local-stddev <= this => featureless (no texture/text)
}


def _err(msg):
    print(json.dumps({"verdict": "ERROR", "reason": msg}), file=sys.stdout)
    sys.exit(2)


def analyze(path, min_conf):
    try:
        import numpy as np
        import cv2
    except Exception as e:                       # honest tool-absent, never fake PASS
        _err("imaging deps absent: %s" % e)
    try:
        import pytesseract
        from pytesseract import Output
    except Exception as e:
        _err("pytesseract absent: %s" % e)

    img = cv2.imread(path)
    if img is None:
        _err("unreadable image: %s" % path)
    h, w = img.shape[:2]
    canvas_area = float(h * w)
    gray = cv2.cvtColor(img, cv2.COLOR_BGR2GRAY)

    checks = {}
    findings = []

    # ---- (1) OCR legibility + collect text boxes ----
    try:
        data = pytesseract.image_to_data(img, output_type=Output.DICT)
    except Exception as e:
        _err("tesseract run failed: %s" % e)
    boxes = []
    legible = 0
    for i in range(len(data["text"])):
        txt = (data["text"][i] or "").strip()
        try:
            conf = float(data["conf"][i])
        except ValueError:
            conf = -1.0
        if txt and conf >= min_conf:
            legible += 1
            boxes.append((data["left"][i], data["top"][i],
                          data["width"][i], data["height"][i], txt, conf))
    checks["legibility"] = {
        "pass": legible >= THRESHOLDS["min_legible_words"],
        "legible_words": legible,
        "min_conf_used": min_conf,
    }
    if legible < THRESHOLDS["min_legible_words"]:
        findings.append("no legible text runs above conf floor (blank/garbled render)")

    # ---- (2) giant unbounded featureless foreground region ----
    # Foreground = brighter than the dark page background. Featureless = low local stddev
    # (no text, no texture) -> the giant gray poster pill / giant blank button.
    fg = (gray > THRESHOLDS["bg_brightness_max"]).astype("uint8")
    # local stddev via box filter on gray and gray^2
    g = gray.astype("float32")
    k = 25
    mean = cv2.blur(g, (k, k))
    mean_sq = cv2.blur(g * g, (k, k))
    var = np.clip(mean_sq - mean * mean, 0, None)
    std = np.sqrt(var)
    featureless = (std <= THRESHOLDS["featureless_std_max"]).astype("uint8")
    mask = (fg & featureless).astype("uint8")
    # morphological close to consolidate the filled pill, drop thin noise
    mask = cv2.morphologyEx(mask, cv2.MORPH_OPEN, np.ones((9, 9), "uint8"))
    mask = cv2.morphologyEx(mask, cv2.MORPH_CLOSE, np.ones((25, 25), "uint8"))
    n, labels, stats, _ = cv2.connectedComponentsWithStats(mask, connectivity=8)
    largest_frac = 0.0
    largest_box = None
    for lbl in range(1, n):
        area = stats[lbl, cv2.CC_STAT_AREA]
        frac = area / canvas_area
        if frac > largest_frac:
            largest_frac = frac
            largest_box = [int(stats[lbl, cv2.CC_STAT_LEFT]), int(stats[lbl, cv2.CC_STAT_TOP]),
                           int(stats[lbl, cv2.CC_STAT_WIDTH]), int(stats[lbl, cv2.CC_STAT_HEIGHT])]
    giant_ok = largest_frac <= THRESHOLDS["giant_region_area_frac_max"]
    checks["no_giant_unbounded_widget"] = {
        "pass": bool(giant_ok),
        "largest_featureless_fg_frac": round(largest_frac, 4),
        "threshold": THRESHOLDS["giant_region_area_frac_max"],
        "region_bbox_xywh": largest_box,
    }
    if not giant_ok:
        findings.append(
            "giant unbounded featureless region covers %.1f%% of canvas (> %.0f%%) at bbox %s "
            "(the §11.4.170 giant-poster/giant-button signature)"
            % (largest_frac * 100, THRESHOLDS["giant_region_area_frac_max"] * 100, largest_box))

    # ---- (3) label-over-label overlap ----
    overlaps = 0
    overlap_pairs = []
    for i in range(len(boxes)):
        ax, ay, aw, ah = boxes[i][:4]
        for j in range(i + 1, len(boxes)):
            bx, by, bw, bh = boxes[j][:4]
            ix = max(0, min(ax + aw, bx + bw) - max(ax, bx))
            iy = max(0, min(ay + ah, by + bh) - max(ay, by))
            inter = ix * iy
            if inter > 0:
                smaller = min(aw * ah, bw * bh)
                if smaller > 0 and inter / smaller > 0.5:   # >50% of the smaller box covered
                    overlaps += 1
                    if len(overlap_pairs) < 8:
                        overlap_pairs.append([boxes[i][4], boxes[j][4]])
    checks["no_label_overlap"] = {"pass": overlaps == 0, "overlapping_pairs": overlaps,
                                  "examples": overlap_pairs}
    if overlaps:
        findings.append("%d overlapping text boxes (label-over-label): %s" % (overlaps, overlap_pairs))

    # ---- (4) on-screen / no clipping ----
    clipped = 0
    for (x, y, bw, bh, txt, conf) in boxes:
        if x < 0 or y < 0 or x + bw > w or y + bh > h:
            clipped += 1
    checks["on_screen"] = {"pass": clipped == 0, "clipped_boxes": clipped}
    if clipped:
        findings.append("%d text boxes clipped/off-screen" % clipped)

    verdict = "PASS" if all(c["pass"] for c in checks.values()) else "FAIL"
    return {
        "verdict": verdict,
        "image": path,
        "canvas": [w, h],
        "checks": checks,
        "findings": findings,
    }


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("image")
    ap.add_argument("--json", action="store_true")
    ap.add_argument("--min-conf", type=int, default=THRESHOLDS["ocr_min_conf"])
    a = ap.parse_args()
    res = analyze(a.image, a.min_conf)
    if a.json:
        print(json.dumps(res, indent=2))
    else:
        print("%s  %s" % (res["verdict"], a.image))
        for f in res["findings"]:
            print("  - " + f)
    sys.exit(0 if res["verdict"] == "PASS" else 1)


if __name__ == "__main__":
    main()
