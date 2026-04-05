---
id: HELIX-282
severity: cosmetic
category: visual
platform: 
screen: android-curiosity-003.png
status: open
found_date: 2026-04-04
---

# Focused app icon's highlight box overlaps adjacent element and text is misaligned.

The highlight box surrounding the focused 'Каталог Заяв' app icon slightly overlaps the adjacent 'RUTUBE' app icon to its right. Additionally, the text label 'Каталог Заяв' appears slightly lower than the labels of other apps in the 'Favorite Apps' row (e.g., 'IPTV', 'VLC'), creating a minor vertical misalignment and a somewhat crowded appearance.

## Related Issues

- HELIX-009: Misaligned Text
- HELIX-010: Logo and text are not aligned
- HELIX-017: Inconsistent font sizes
- HELIX-020: Inconsistent Font Sizes
- HELIX-039: Inconsistent button styles
- HELIX-048: Color scheme issue
- HELIX-053: Inconsistent Text Color
- HELIX-063: Inconsistent font size
- HELIX-077: Inconsistent Spacing Between Elements
- HELIX-100: Insufficient Contrast
- HELIX-110: Insufficient White Space
- HELIX-111: Missing background color
- HELIX-120: Inconsistent Icon Sizes
- HELIX-144: Button overlap
- HELIX-145: Button shadow overlaps text
- HELIX-176: Keyboard layout is not optimized for mobile devices


## Reproduction Steps

Navigate to the Home screen and observe the focus state when it's on the app labeled 'Каталог Заяв'.

## Evidence

The right edge of the black highlight box for 'Каталог Заяв' touches the left edge of the 'RUTUBE' icon. The baseline of the 'Каталог Заяв' text is lower than the baseline of 'IPTV' or 'VLC' texts.
