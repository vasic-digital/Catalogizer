---
id: HELIX-298
severity: medium
category: content
platform: 
screen: android-curiosity-017-after.png
status: resolved
resolution: fixed
fixed_date: 2026-04-06
found_date: 2026-04-04
---

# Incorrect movie type/genre tag displayed

The tag displayed next to the year '1968' is 'MOVIL'. Assuming the application's primary language is English, this is likely a typo or an incorrect localization, as 'MOVIL' means 'mobile' in Spanish and is out of context for describing a film. It should likely be 'MOVIE' or a genre such as 'Sci-Fi'. This causes confusion for the user and indicates a data error or localization issue.

## Related Issues

- HELIX-016: Lack of content
- HELIX-036: Outdated content
- HELIX-057: Outdated Content
- HELIX-074: Typos and Grammatical Errors
- HELIX-116: Missing content
- HELIX-280: App label is truncated on the home screen.
- HELIX-286: Typo in movie category label


## Evidence

The text label 'MOVIL' next to '1968' below the movie title.
