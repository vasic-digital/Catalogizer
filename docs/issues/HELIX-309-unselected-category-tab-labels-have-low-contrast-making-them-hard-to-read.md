---
id: HELIX-309
severity: medium
category: visual
platform: 
screen: android-curiosity-027.png
status: open
found_date: 2026-04-04
---

# Unselected category tab labels have low contrast, making them hard to read.

The text labels for unselected content categories (e.g., 'Movies', 'Software', 'TV Shows', 'Books') are displayed in a light grey color against a dark background. This results in low contrast, making the labels difficult to read, especially for users with visual impairments. This also creates an inconsistent visual hierarchy, as the corresponding item counts (e.g., '174', '2') and the text in the selected 'Comics' tab are much brighter and more legible.

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


## Reproduction Steps

N/A (visible in screenshot)

## Evidence

The text 'Movies', 'Software', 'TV Shows', and 'Books' in the unselected tabs appears faded and less distinct than their corresponding numbers (174, 2, 2, 1) and the text 'Comics' in the selected tab.
