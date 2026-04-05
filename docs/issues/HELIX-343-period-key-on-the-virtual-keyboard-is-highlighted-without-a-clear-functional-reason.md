---
id: HELIX-343
severity: low
category: visual
platform: 
screen: androidtv-curiosity-001-after.png
status: open
found_date: 2026-04-05
---

# Period key on the virtual keyboard is highlighted without a clear functional reason.

The period key on the virtual keyboard is visually highlighted with a white background. This highlighting typically indicates a selected state, a recently pressed key, or a focused element. However, in this context, combined with the cursor issue, it's unclear why it's highlighted. If it's a 'last pressed' indicator, it persists for too long or conflicts with the general input state. If it's a focus state, it implies the keyboard key is focused rather than the input field, which is a UX problem.

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


## Reproduction Steps

N/A (observed state in screenshot)

## Evidence

The key for the period character ('.') on the virtual keyboard has a distinct white background, contrasting with the dark background of other keys.
