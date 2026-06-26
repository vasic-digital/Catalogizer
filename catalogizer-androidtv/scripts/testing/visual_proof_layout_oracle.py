#!/usr/bin/env python3
"""
§11.4.170 dual-oracle (ii): OCR / vision LAYOUT oracle for host-rendered UI PNGs.

Reads a rendered PNG (host-side render OR on-device screencap) and asserts the
§11.4.162 layout invariants ON REAL PIXELS — the proof a value/token-equality
test can never give:
  - legibility:        OCR text runs present + per-word confidence >= floor
  - no giant/unbounded: no single featureless FOREGROUND region (brighter than
                        the dark page background) exceeds a calibrated fraction
                        of the canvas (the giant-button / giant-poster-placeholder
                        signature, §11.4.170 forensic)
  - no interior blank: no single featureless region OF ANY BRIGHTNESS that is
                        large AND sits in the canvas INTERIOR (does not hug the
                        canvas border the way the true page background does) —
                        catches a giant placeholder rendered AT/near the
                        background color that the foreground check cannot see
                        (closes review finding L1, §11.4.107(10))
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
  L1 (giant-widget at background color) — CLOSED 2026-06-26 by the interior-blank
     check (check #3 below, no_giant_interior_blank). The foreground giant-widget
     detector (#2) keys on a featureless region BRIGHTER than the dark page
     background (gray > bg_brightness_max), so a giant featureless placeholder
     rendered BELOW that cutoff (gray <= bg_brightness_max) used to EVADE it
     (frac=0.0). The interior-blank check now builds a featureless mask over ALL
     brightness and flags any large featureless component that sits in the canvas
     INTERIOR (low canvas-border contact) — the true page background hugs the
     border (contact ~1.0) and is NOT flagged, while a centered placeholder box
     (contact ~0.0) IS. Self-validated by the golden_bad_interior_blank_bgcolor
     fixture (a gray~38 box, BELOW the gray>40 fg cutoff, that PASSes the #2 check
     and FAILs the #3 check). RESIDUAL LIMIT (honest §11.4.6): a blank rendered at
     the EXACT background color/texture (color delta < ~the separating edge
     contrast, so no edge breaks the featureless connectivity) merges with the
     background into one border-touching component and is NOT separable — but such
     a region is also ~invisible to the end user (it IS the background); and a
     giant interior blank legitimately is a §11.4.170 defect, so a large intended
     uniform interior surface (e.g. a full-screen player fill) would also flag —
     by design, not a false positive for this mandate.
  L2 (legibility near-vacuous): min_legible_words=1 makes legibility a weak floor;
     the giant-widget check carries the actual §11.4.170 proof, not legibility.
  L3 (overlap/clip not yet golden-bad-guarded): the no_label_overlap and on_screen
     checks are IMPLEMENTED and logically sound but ship WITHOUT a §11.4.107(10)
     golden-bad fixture (a flaky synthetic overlap fixture was honestly dropped
     rather than shipped as a bluff). FOLLOW-UP: capture a real overlap/clip defect
     frame as a golden-bad and bring these checks under the self-test.
"""
import sys, json, argparse

# Calibrated on this project's real device captures (qa-results/device_qa_20260626/)
# plus one synthesized L1-evasion fixture (§11.4.6 — thresholds set from measured
# separation on THESE frames, recorded here with the measurements):
#   golden-GOOD  golden_good_home.png                   (clean home; real MIBOX4 capture)
#   golden-BAD   golden_bad_giant_poster.png            (Inception giant poster; real capture)
#   golden-BAD   golden_bad_interior_blank_bgcolor.png  (synth: gray~38 interior box BELOW
#                                                         the gray>40 fg cutoff — the L1 evasion)
# Measured (interior featureless mask, std<=6.0, OPEN 9x9, all-brightness, connected comps):
#   clean home   -> ONE featureless blob frac 0.82 hugging the border (contact 0.99) => NOT interior
#   giant poster -> a featureless interior blob frac 0.57 contact 0.00 (also caught by the fg check)
#   synth blank  -> a featureless interior blob frac 0.59 contact 0.00; fg check sees frac 0.00
# The giant gray poster pill measured ~0.45-0.70 canvas-area in the FG check; clean home's
# largest featureless FG blob (a card) measured well under 0.25 (fg threshold 0.32). For the
# interior check, the clean-home background's border contact is ~1.0 and the two defect blobs'
# is ~0.0, a wide margin -> interior_border_contact_max set at 0.10.
THRESHOLDS = {
    "giant_region_area_frac_max": 0.32,  # > this featureless fg fraction => giant-unbounded FAIL
    "ocr_min_conf": 55,                  # per-word tesseract confidence floor
    "min_legible_words": 1,              # at least one legible text run expected
    "bg_brightness_max": 40,             # gray<=this is treated as page background (dark navy ~10-25)
    "featureless_std_max": 12.0,         # local-stddev <= this => featureless (no texture/text)
    # --- interior-blank check (closes L1): featureless region of ANY brightness ---
    "interior_featureless_std_max": 6.0,    # lower std floor so a low-contrast box edge breaks
                                            # featureless connectivity & separates box from bg
    "interior_region_area_frac_max": 0.32,  # > this interior featureless fraction => blank FAIL
    "interior_border_contact_max": 0.10,    # component touching < this of the canvas border is
                                            # INTERIOR (not the true page background, contact ~1.0)
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

    # ---- (3) giant INTERIOR blank at/near background color (closes L1) ----
    # Foreground check (2) is blind to a featureless region rendered at/below the page
    # background brightness (gray <= bg_brightness_max). Build a featureless mask over ALL
    # brightness with a lower std floor (so a low-contrast box edge breaks featureless
    # connectivity and the box separates from the bg), connected-component it, and flag any
    # LARGE component whose canvas-border contact is LOW. The true page background hugs the
    # canvas border (contact ~1.0) and is exonerated; a centered placeholder pill (contact
    # ~0.0) is flagged even at background color.
    istd = np.sqrt(np.clip(mean_sq - mean * mean, 0, None))  # reuse mean/mean_sq from above
    ifeat = (istd <= THRESHOLDS["interior_featureless_std_max"]).astype("uint8")
    ifeat = cv2.morphologyEx(ifeat, cv2.MORPH_OPEN, np.ones((9, 9), "uint8"))
    inn, ilabels, istats, _ = cv2.connectedComponentsWithStats(ifeat, connectivity=8)
    border_lbls = np.concatenate([ilabels[0, :], ilabels[h - 1, :],
                                  ilabels[:, 0], ilabels[:, w - 1]])
    border_counts = np.bincount(border_lbls.ravel(), minlength=inn)
    perimeter = 2.0 * (w + h)
    interior_frac = 0.0
    interior_contact = 1.0
    interior_box = None
    for lbl in range(1, inn):
        frac = istats[lbl, cv2.CC_STAT_AREA] / canvas_area
        if frac < THRESHOLDS["interior_region_area_frac_max"]:
            continue
        contact = border_counts[lbl] / perimeter
        if contact < THRESHOLDS["interior_border_contact_max"] and frac > interior_frac:
            interior_frac = frac
            interior_contact = round(float(contact), 4)
            interior_box = [int(istats[lbl, cv2.CC_STAT_LEFT]), int(istats[lbl, cv2.CC_STAT_TOP]),
                            int(istats[lbl, cv2.CC_STAT_WIDTH]), int(istats[lbl, cv2.CC_STAT_HEIGHT])]
    interior_ok = interior_frac <= 0.0  # nothing large + interior found => OK
    checks["no_giant_interior_blank"] = {
        "pass": bool(interior_ok),
        "largest_interior_featureless_frac": round(interior_frac, 4),
        "border_contact": interior_contact if interior_box else None,
        "area_frac_threshold": THRESHOLDS["interior_region_area_frac_max"],
        "border_contact_threshold": THRESHOLDS["interior_border_contact_max"],
        "region_bbox_xywh": interior_box,
    }
    if not interior_ok:
        findings.append(
            "giant INTERIOR featureless blank covers %.1f%% of canvas (border-contact %.2f < %.2f) "
            "at bbox %s — a placeholder/blank at/near the background color the foreground check "
            "cannot see (§11.4.170 giant-unbounded signature, L1)"
            % (interior_frac * 100, interior_contact,
               THRESHOLDS["interior_border_contact_max"], interior_box))

    # ---- (4) label-over-label overlap ----
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

    # ---- (5) on-screen / no clipping ----
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
