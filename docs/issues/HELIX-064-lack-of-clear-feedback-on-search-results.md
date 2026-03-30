---
id: HELIX-064
severity: medium
category: ux
platform: 
screen: androidtv-curiosity-047.png
status: wontfix
found_date: 2026-03-30
---

# Lack of clear feedback on search results

When the user types a query in the search bar and presses enter, there is no clear feedback on whether the search was successful or not. This may cause the user to wonder if the search is still processing or if the results are being displayed.

## Related Issues

- HELIX-001: Insecure Password Input
- HELIX-002: Lack of Feedback on Form Submission
- HELIX-003: Inconsistent Label Alignment
- HELIX-004: Lack of Placeholder Text
- HELIX-005: Inconsistent Button Styling
- HELIX-006: Password field does not have a show password option
- HELIX-007: Username and password fields do not have labels
- HELIX-008: Inconsistent Button Design
- HELIX-011: Lack of clear navigation
- HELIX-018: Lack of clear search functionality
- HELIX-021: Lack of Clear Call-to-Action
- HELIX-023: Unclear search functionality
- HELIX-024: Keyboard Layout Issue
- HELIX-025: Lack of Progress Indication
- HELIX-026: Insufficient Feedback
- HELIX-028: Lack of Clear Error Message
- HELIX-029: Insufficient Recovery Options
- HELIX-030: Lack of Visual Hierarchy
- HELIX-031: No results found for search query
- HELIX-032: Lack of clear instructions or guidance
- HELIX-033: Inconsistent button styling
- HELIX-034: Unclear navigation
- HELIX-042: Inconsistent Text Alignment
- HELIX-043: Insufficient Spacing Between Elements
- HELIX-044: Small Font Size
- HELIX-045: Inconsistent Button Styles
- HELIX-046: Keyboard overlap
- HELIX-047: Keyboard layout issue
- HELIX-050: Inadequate Error Message
- HELIX-051: Lack of Suggested Actions
- HELIX-052: Inconsistent Spelling
- HELIX-058: Search Bar Placement
- HELIX-059: Keyboard layout is not optimized for the search field
- HELIX-060: Unclear Search Functionality


## Reproduction Steps

Open the application, type a query in the search bar, and press enter. Observe the lack of clear feedback on the search results.

## Evidence

There is no loading animation or message indicating that the search is processing. The search results are displayed without any clear indication of whether the search was successful or not.

## Resolution

Enhancement suggestion from automated QA. Navigation follows Android TV Leanback patterns (D-pad based). Search functionality works as designed via the browse fragment.
Closed: 2026-03-30
