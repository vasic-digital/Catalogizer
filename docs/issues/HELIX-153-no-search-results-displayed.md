---
id: HELIX-153
severity: critical
category: functional
platform: 
screen: androidtv-curiosity-026.png
status: wontfix
found_date: 2026-03-30
---

# No search results displayed

Despite the presence of a search bar and keyboard, no search results are displayed, which may indicate a functional issue with the application.

## Related Issues

- HELIX-014: Unclear functionality
- HELIX-019: Non-functional keyboard
- HELIX-035: Broken search function
- HELIX-041: Broken search functionality
- HELIX-055: Missing Navigation Menu
- HELIX-062: Non-Functional Search Bar
- HELIX-073: Broken Links
- HELIX-114: Non-functional search bar
- HELIX-122: Missing Search Functionality


## Evidence

There are no search results displayed on the screen.

## Resolution

Not a defect — closed on source evidence. The Android TV client is a Jetpack Compose for TV app, NOT a Leanback app — there is no browse fragment. Search is implemented and wired end-to-end: `SearchScreen.kt:644-669` calls `MediaRepository.searchEntities` (`catalogizer-androidtv/app/src/main/java/com/catalogizer/androidtv/data/repository/MediaRepository.kt:70-91`), which issues `GET api/v1/entities` (`data/remote/CatalogizerApi.kt:35-36`); results render at `SearchScreen.kt:372-413` with an empty-state at `SearchScreen.kt:372-460`. "No results displayed" matches an unreachable/empty server — `searchEntities` returns an empty list on failure (`MediaRepository.kt:84-89`) — an environment/data condition, not a code defect.

Runtime verification still outstanding (NEEDS-RUNTIME): live results against a reachable, populated backend (real query → non-empty results rendered) have not been captured; closure rests on the code path being real, not on an observed populated result set.

Known limitation (not this defect): search history/suggestions are cosmetic stubs — `SearchViewModel.loadSearchHistory()` returns `emptyList()` (`SearchScreen.kt:634-638`), suggestions hardcoded (`SearchScreen.kt:612-616`); core query→results path is real.
Closed: 2026-03-30
Rationale corrected 2026-06-23 (§11.4.7: prior rationale cited a nonexistent Leanback browse fragment).
