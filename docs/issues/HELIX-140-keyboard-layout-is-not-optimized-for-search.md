---
id: HELIX-140
severity: medium
category: ux
platform: video-frame
screen: frame_0001.png
status: wontfix
found_date: 2026-03-30
resolution: QA infrastructure failure: screenshot showed generic search results with no evidence of voice search intent handling. No reproducible bug in app.
closed_date: 2026-04-17
---

# Keyboard layout is not optimized for search

The keyboard layout is not optimized for search, making it difficult for users to enter their search query quickly.

## Related Issues

- HELIX-093: Password Field Not Displaying Correctly
- HELIX-094: Server Connection Field Not Displaying Correctly
- HELIX-095: Username Field Not Displaying Correctly
- HELIX-096: Sign In Button Not Displaying Correctly
- HELIX-097: Lack of Feedback on Password Strength
- HELIX-098: Unclear Error Messages
- HELIX-138: No results found message is unclear


## Reproduction Steps

Open the keyboard and try to enter a search query.

## Evidence

The keyboard layout is not optimized for search.

## Resolution

Enhancement suggestion from automated QA. Navigation follows Android TV Leanback patterns (D-pad based). Search functionality works as designed via the browse fragment.
Closed: 2026-03-30
