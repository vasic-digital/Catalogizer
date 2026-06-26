#!/usr/bin/env python3
"""
Regenerator for the §11.4.107(10) golden-BAD fixture that closes review finding L1
of the §11.4.170 layout oracle: golden_bad_interior_blank_bgcolor.png.

The fixture reproduces the L1 evasion EXACTLY:
  - a giant featureless rounded box (~60% of a 1920x1080 canvas), centered, filled
    with a DARK near-background navy (RGB 32,38,52 -> gray ~38) that is BELOW the
    oracle's foreground brightness cutoff (gray > 40), so the foreground giant-widget
    detector sees frac=0.0 and PASSes it (the bug);
  - drawn on the app's real dark-navy background (RGB 15,23,42 -> gray ~23);
  - PLUS legible title text ("Inception", "2010   Play Now") elsewhere so the
    legibility check passes — proving the screen is otherwise legible while a giant
    blank placeholder is present.
The patched oracle's interior-blank check (featureless region of ANY brightness that
sits in the canvas interior) FAILs it. Deterministic: identical bytes every run.

Usage: python3 make_interior_blank_fixture.py
"""
import os
from PIL import Image, ImageDraw, ImageFont

HERE = os.path.dirname(os.path.abspath(__file__))
OUT = os.path.join(HERE, "fixtures", "golden_bad_interior_blank_bgcolor.png")

W, H = 1920, 1080
BG = (15, 23, 42)        # real app dark navy, gray ~23
BOX = (32, 38, 52)       # gray ~38: BELOW the gray>40 foreground cutoff => evades fg check
BOX_W, BOX_H = int(W * 0.86), int(H * 0.70)   # 0.602 of the canvas


def _font(size):
    for p in ("/System/Library/Fonts/Supplemental/Arial.ttf",
              "/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf",
              "/Library/Fonts/Arial.ttf"):
        if os.path.exists(p):
            return ImageFont.truetype(p, size)
    return ImageFont.load_default()


def build():
    im = Image.new("RGB", (W, H), BG)
    d = ImageDraw.Draw(im)
    x0 = (W - BOX_W) // 2
    y0 = (H - BOX_H) // 2
    d.rounded_rectangle([x0, y0, x0 + BOX_W, y0 + BOX_H], radius=28, fill=BOX)
    fnt = _font(54)
    d.text((60, 30), "Inception", fill=(235, 235, 240), font=fnt)
    d.text((60, H - 90), "2010    Play Now", fill=(225, 225, 230), font=fnt)
    im.save(OUT)
    print("wrote %s (%dx%d, box %.3f of canvas)" % (OUT, W, H, (BOX_W * BOX_H) / (W * H)))


if __name__ == "__main__":
    build()
