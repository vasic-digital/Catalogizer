---
id: HELIX-333
severity: medium
category: visual
platform: 
screen: androidtv-curiosity-016.png
status: open
found_date: 2026-04-04
---

# Highlighted keyboard key does not match the last character in the search input field.

The on-screen keyboard displays the '6' key as highlighted, which typically indicates it is currently selected or has just been pressed. However, the search input field clearly shows '5' as the last entered digit ('mcv cw 2. 8777775'). This visual inconsistency can confuse the user about the actual character being input or the current state of the keyboard interaction.

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
- HELIX-290: Inconsistent button design for playback controls.
- HELIX-296: Inconsistent design language for playback control buttons.
- HELIX-300: Inconsistent icon usage for action buttons
- HELIX-301: Inconsistent width of action buttons at the bottom
- HELIX-305: Inconsistent border radius among action buttons
- HELIX-306: Misaligned icon and text in 'Favorite' button
- HELIX-309: Unselected category tab labels have low contrast, making them hard to read.
- HELIX-311: Low contrast and potential clipping for category names.
- HELIX-313: 'Recently Added TV Shows' section is incomplete.
- HELIX-314: Inconsistent background highlight for settings icon.
- HELIX-316: Truncated description text in 'Channel Tap Behavior' section
- HELIX-318: Low contrast for category names in library filter tags
- HELIX-319: Last 'Recently Added Movies' item is partially cut off
- HELIX-321: Search button text has insufficient contrast with its background.
- HELIX-324: Search button is slightly misaligned with the search input field.
- HELIX-326: Search input text appears vertically misaligned within the search bar.


## Reproduction Steps

1. Navigate to the search bar and activate the on-screen keyboard.
2. Enter the string 'mcv cw 2. 8777775' into the search bar.
3. Observe that the '6' key on the virtual keyboard is highlighted while the input prominently ends with '5'.

## Evidence

The search input field displays 'mcv cw 2. 8777775', while simultaneously the '6' key on the virtual keyboard is highlighted.
