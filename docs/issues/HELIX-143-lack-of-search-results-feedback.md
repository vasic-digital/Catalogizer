---
id: HELIX-143
severity: high
category: ux
platform: video-frame
screen: frame_0003.png
status: wontfix
found_date: 2026-03-30
---

# Lack of Search Results Feedback

The search bar does not provide any feedback or results when a user enters a query, which can be frustrating and may lead to users abandoning the application.

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
- HELIX-142: Unclear Search Query


## Reproduction Steps

Enter a query in the search bar and press enter.

## Evidence

The search bar is empty and there are no results displayed.

## Resolution

Enhancement suggestion from automated QA. Navigation follows Android TV Leanback patterns (D-pad based). Search functionality works as designed via the browse fragment.
Closed: 2026-03-30
