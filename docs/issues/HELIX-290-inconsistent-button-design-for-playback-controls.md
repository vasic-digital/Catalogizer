---
id: HELIX-290
severity: low
category: visual
platform: 
screen: android-curiosity-010.png
status: resolved
resolution: fixed
fixed_date: 2026-04-06
found_date: 2026-04-04
---

# Inconsistent button design for playback controls.

There is an inconsistency in the design of the playback control buttons. The primary play button in the center of the screen is an icon-only button, while all buttons in the bottom control bar, including the secondary play button, use both an icon and text (e.g., '-10s', 'Play', '+10s', '1.0x', 'Audio', 'Off'). This mix of icon-only and icon-with-text designs can lead to a less polished user interface and may affect visual consistency.

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
- HELIX-282: Focused app icon's highlight box overlaps adjacent element and text is misaligned.
- HELIX-283: Low contrast for unselected library category buttons
- HELIX-284: Text 'TV Shows' is slightly cut off in category button
- HELIX-285: Text 'Books' lacks sufficient right padding in category button
- HELIX-287: "Back to Library" button text is truncated


## Reproduction Steps

1. Observe the video playback screen.

## Evidence

The central orange play icon (icon-only) contrasted with the 'Play' button in the bottom bar (icon + text), and other bottom buttons like '-10s' and '+10s' also being icon + text.
