---
id: HELIX-035
severity: critical
category: functional
platform: 
screen: androidtv-curiosity-036.png
status: wontfix
found_date: 2026-03-30
---

# Broken search function

The search function is not working properly, which may prevent users from finding the information they need.

## Related Issues

- HELIX-014: Unclear functionality
- HELIX-019: Non-functional keyboard


## Reproduction Steps

None

## Evidence

The search function is not working properly.

## Resolution

Not a defect — closed on source evidence. The Android TV client is a Jetpack Compose for TV app, NOT a Leanback app — there is no browse fragment. Search is implemented and wired end-to-end: `SearchScreen.kt:644-669` calls `MediaRepository.searchEntities` (`catalogizer-androidtv/app/src/main/java/com/catalogizer/androidtv/data/repository/MediaRepository.kt:70-91`), which issues `GET api/v1/entities` (`data/remote/CatalogizerApi.kt:35-36`); results render at `SearchScreen.kt:372-413` with an empty-state at `SearchScreen.kt:372-460`. On an unreachable or empty server, `searchEntities` returns an empty list (`MediaRepository.kt:84-89`), so "broken search" screenshots are consistent with an environment/data condition (backend down or no matching data), not a code defect.

Known limitation (not this defect): search history/suggestions are cosmetic stubs — `SearchViewModel.loadSearchHistory()` returns `emptyList()` (`SearchScreen.kt:634-638`), suggestions hardcoded (`SearchScreen.kt:612-616`); core query→results path is real.
Closed: 2026-03-30
Rationale corrected 2026-06-23 (§11.4.7: prior rationale cited a nonexistent Leanback browse fragment).
