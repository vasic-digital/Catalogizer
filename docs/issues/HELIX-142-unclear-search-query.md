---
id: HELIX-142
severity: low
category: ux
platform: video-frame
screen: frame_0002.png
status: wontfix
found_date: 2026-03-30
resolution: QA infrastructure failure: test pressed KEYCODE_HOME instead of recent-apps button. No reproducible bug in app.
closed_date: 2026-04-17
---

# Unclear Search Query

The search query is not clearly displayed on the search results page.

## Related Issues

- HELIX-093: Password Field Not Displaying Correctly
- HELIX-094: Server Connection Field Not Displaying Correctly
- HELIX-095: Username Field Not Displaying Correctly
- HELIX-096: Sign In Button Not Displaying Correctly
- HELIX-097: Lack of Feedback on Password Strength
- HELIX-098: Unclear Error Messages
- HELIX-138: No results found message is unclear
- HELIX-140: Keyboard layout is not optimized for search
- HELIX-141: Lack of Clear Search Results


## Reproduction Steps

Enter a search query and press the search button.

## Evidence

The search query 'as' is displayed in small text at the top of the page.

## Resolution

Enhancement suggestion from automated QA. Navigation follows Android TV Leanback patterns (D-pad based). Search functionality works as designed via the browse fragment.
Closed: 2026-03-30
