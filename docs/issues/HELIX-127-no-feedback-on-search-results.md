---
id: HELIX-127
severity: low
category: ux
platform: 
screen: androidtv-017-navigate.png
status: wontfix
found_date: 2026-03-30
resolution: QA infrastructure failure: screenshot showed on-screen keyboard instead of connection attempt to self-signed cert server. No reproducible bug in app.
closed_date: 2026-04-17
---

# No feedback on search results

When the user enters a search query and presses the search button, there is no feedback to indicate that the search is being processed or that no results were found. Adding a loading animation or a message to indicate that no results were found could improve the user experience.

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
- HELIX-064: Lack of clear feedback on search results
- HELIX-065: Inconsistent Image Size
- HELIX-066: Unclear Navigation
- HELIX-067: Lack of clear categorization
- HELIX-068: Insufficient information
- HELIX-069: Lack of filtering options
- HELIX-070: Inconsistent spacing
- HELIX-075: Inconsistent Input Field Widths
- HELIX-076: Lack of Placeholder Text in Input Fields
- HELIX-078: Keyboard Layout
- HELIX-079: Inconsistent Keyboard Layout
- HELIX-080: Unclear Call-to-Action
- HELIX-081: Insufficient Guidance
- HELIX-082: Lack of Feedback on Search Results
- HELIX-083: No Clear Call-to-Action
- HELIX-084: Inconsistent Search Results Display
- HELIX-085: Lack of Clear Search Results Indication
- HELIX-086: Unclear Search Results
- HELIX-087: Lack of Feedback on Search Input
- HELIX-088: Lack of Clear Input Field
- HELIX-089: No Feedback Mechanism
- HELIX-090: Lack of clear feedback
- HELIX-091: Search bar not responding to input
- HELIX-092: No clear call-to-action
- HELIX-099: Lack of Visible Buttons
- HELIX-103: Lack of feedback on search results
- HELIX-104: Unclear error message
- HELIX-105: Insufficient Error Handling
- HELIX-107: Lack of clear instructions on how to resolve the error
- HELIX-108: Unclear button functionality
- HELIX-109: Lack of Feedback
- HELIX-112: Unresponsive buttons
- HELIX-117: Inconsistent Typography
- HELIX-118: Inadequate Search Results Feedback
- HELIX-119: Lack of Suggested Alternatives
- HELIX-121: Lack of Clear File Type Indicators
- HELIX-123: Insufficient Error Recovery Options
- HELIX-124: Unclear Button Functionality
- HELIX-125: Lack of Clear Navigation
- HELIX-126: No clear call to action


## Evidence

The search bar is empty and there is no text or animation to indicate that the search is being processed or that no results were found.

## Resolution

Enhancement suggestion from automated QA. Navigation follows Android TV Leanback patterns (D-pad based). Search functionality works as designed via the browse fragment.
Closed: 2026-03-30
