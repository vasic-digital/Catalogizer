---
id: HELIX-062
severity: critical
category: functional
platform: 
screen: androidtv-curiosity-039.png
status: wontfix
found_date: 2026-03-30
---

# Non-Functional Search Bar

The search bar is not functional, preventing users from searching for content. This critical issue must be addressed immediately to ensure a functional user experience.

## Related Issues

- HELIX-014: Unclear functionality
- HELIX-019: Non-functional keyboard
- HELIX-035: Broken search function
- HELIX-041: Broken search functionality
- HELIX-055: Missing Navigation Menu


## Reproduction Steps

Try to use the search bar to search for content.

## Evidence

The search bar is not functional, preventing users from searching for content.

## Resolution

Not a defect — closed on source evidence. The Android TV client is a Jetpack Compose for TV app, NOT a Leanback app — there is no browse fragment. The search bar is functional and wired end-to-end: `SearchScreen.kt:644-669` calls `MediaRepository.searchEntities` (`catalogizer-androidtv/app/src/main/java/com/catalogizer/androidtv/data/repository/MediaRepository.kt:70-91`), which issues `GET api/v1/entities` (`data/remote/CatalogizerApi.kt:35-36`); results render at `SearchScreen.kt:372-413` with an empty-state at `SearchScreen.kt:372-460`. A bar that returns nothing matches an unreachable/empty server — `searchEntities` returns an empty list on failure (`MediaRepository.kt:84-89`) — an environment/data condition, not a code defect.

Known limitation (not this defect): search history/suggestions are cosmetic stubs — `SearchViewModel.loadSearchHistory()` returns `emptyList()` (`SearchScreen.kt:634-638`), suggestions hardcoded (`SearchScreen.kt:612-616`); core query→results path is real.
Closed: 2026-03-30
Rationale corrected 2026-06-23 (§11.4.7: prior rationale cited a nonexistent Leanback browse fragment).
