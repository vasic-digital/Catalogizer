---
id: HELIX-356
severity: medium
category: visual
platform: 
screen: androidtv-curiosity-011.png
status: resolved
resolution: fixed
fixed_date: 2026-04-06
found_date: 2026-04-05
---

# Unexpected diacritic characters in search input field

The text entered in the search bar, intended to be 'shows', displays with diacritic marks as 'šhõws'. Specifically, the 's' character appears with a caron (š) and the 'o' character appears with a tilde (õ). This can be a font rendering issue or an issue with the input method incorrectly converting standard characters, leading to a potentially incorrect or unsearchable query and user confusion.

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
- HELIX-333: Highlighted keyboard key does not match the last character in the search input field.
- HELIX-337: Keyboard key highlight inconsistent with search result display.
- HELIX-343: Period key on the virtual keyboard is highlighted without a clear functional reason.
- HELIX-345: The 'Search' button appears partially clipped on the right edge.
- HELIX-349: Inconsistent background styling for interactive buttons.
- HELIX-353: Suggested word buttons display inconsistent styling.


## Reproduction Steps

1. Navigate to the search screen.
2. Type 'shows' into the search input field.
3. Observe the characters displayed.

## Evidence

The text 'šhõws' visible within the search input field at the top of the screen.
