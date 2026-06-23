---
id: HELIX-122
severity: critical
category: functional
platform: 
screen: androidtv-curiosity-033.png
status: wontfix
found_date: 2026-03-30
resolution: QA infrastructure failure: test captured login screen before any login attempt was made. No slow network was actually simulated. No reproducible bug in app.
closed_date: 2026-04-17
---

# Missing Search Functionality

Despite having a search bar, there's no indication of how the search function works or what it searches for, which could severely hinder user productivity and satisfaction.

## Related Issues

- HELIX-014: Unclear functionality
- HELIX-019: Non-functional keyboard
- HELIX-035: Broken search function
- HELIX-041: Broken search functionality
- HELIX-055: Missing Navigation Menu
- HELIX-062: Non-Functional Search Bar
- HELIX-073: Broken Links
- HELIX-114: Non-functional search bar


## Reproduction Steps

Attempt to use the search bar by typing in a query and pressing enter or clicking the search button.

## Evidence

The search bar is present but lacks any functionality or feedback upon use.

## Resolution

Not a defect — closed on source evidence. The Android TV client is a Jetpack Compose for TV app, NOT a Leanback app — there is no browse fragment. Search is implemented and wired end-to-end: `SearchScreen.kt:644-669` calls `MediaRepository.searchEntities` (`catalogizer-androidtv/app/src/main/java/com/catalogizer/androidtv/data/repository/MediaRepository.kt:70-91`), which issues `GET api/v1/entities` (`data/remote/CatalogizerApi.kt:35-36`); results render at `SearchScreen.kt:372-413` with an empty-state at `SearchScreen.kt:372-460`. The original frontmatter note also stands: the screenshot captured the login screen before any login/search attempt. A search bar that shows nothing matches an unreachable/empty server — `searchEntities` returns an empty list on failure (`MediaRepository.kt:84-89`) — an environment/data condition, not a code defect.

Known limitation (not this defect): search history/suggestions are cosmetic stubs — `SearchViewModel.loadSearchHistory()` returns `emptyList()` (`SearchScreen.kt:634-638`), suggestions hardcoded (`SearchScreen.kt:612-616`); core query→results path is real.
Closed: 2026-03-30
Rationale corrected 2026-06-23 (§11.4.7: prior rationale cited a nonexistent Leanback browse fragment).
