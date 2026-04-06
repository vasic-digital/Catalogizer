---
id: HELIX-311
severity: medium
category: visual
platform: 
screen: android-curiosity-029.png
status: resolved
resolution: fixed
fixed_date: 2026-04-06
found_date: 2026-04-04
---

# Low contrast and potential clipping for category names.

The text labels for categories (e.g., 'Movies', 'Comics', 'Software', 'TV Shows', 'Books') have significantly lower contrast and a thinner font weight compared to their corresponding numerical counts. This makes them less readable, especially for users with visual impairments. Additionally, the text for 'Movies', 'Comics', and 'Software' appears to be positioned too close to the bottom edge of their respective buttons/containers, causing the descenders of some letters (e.g., 's' in Movies, 'e' in Software) to be almost or partially clipped, indicating insufficient vertical padding or incorrect text alignment.

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


## Evidence

Observe the text 'Movies' next to '174', 'Comics' next to '10', and 'Software' next to '2'. The text is lighter and thinner than the numbers, and the bottom of the text is very close to the container's edge, particularly for 'Movies' and 'Software'.
