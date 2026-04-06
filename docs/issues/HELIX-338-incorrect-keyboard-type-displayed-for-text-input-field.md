---
id: HELIX-338
severity: high
category: ux
platform: 
screen: androidtv-001-loginform.png
status: resolved
resolution: fixed
fixed_date: 2026-04-06
found_date: 2026-04-05
---

# Incorrect keyboard type displayed for text input field

When the user is interacting with the search input field, which expects alphabetic characters (as demonstrated by the text 'movies.' already entered), the application displays a numeric/symbol-only keyboard. This forces users to manually switch keyboards to type common words or numeric values, leading to a frustrating and inefficient input experience. A text input field should default to an alphabetic keyboard (e.g., QWERTY) to facilitate typing words.

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
- HELIX-127: No feedback on search results
- HELIX-128: Inadequate Feedback on Loading Status
- HELIX-129: Lack of Clear Instruction
- HELIX-130: Lack of Progress Indicator
- HELIX-131: Unclear Error Message
- HELIX-132: No Option to Retry
- HELIX-133: Lack of User Feedback
- HELIX-134: Button Labeling
- HELIX-135: Indeterminate Progress Indicator
- HELIX-136: Lack of Clear Filtering Options
- HELIX-137: Inconsistent Button Labels
- HELIX-146: Button overlaps other elements on screen
- HELIX-147: Too many buttons in a row
- HELIX-148: Keyboard layout mismatch
- HELIX-149: Lack of clear input field label
- HELIX-150: No title for tabbed navigation items
- HELIX-151: Negative wording in a button label
- HELIX-152: Lack of clear search button
- HELIX-154: Non-responsive button state on login page.
- HELIX-170: Insufficient Call-to-Action
- HELIX-171: Unnecessary 'Back to Library' Button
- HELIX-172: Keyboard overlaying search results
- HELIX-173: No search history or suggestions
- HELIX-175: Search bar functionality is unclear
- HELIX-183: Lack of Visual Feedback
- HELIX-278: Icon's interactive area may be too small for optimal usability.
- HELIX-279: Icon displayed without a supporting text label.
- HELIX-281: Inconsistent languages detected on the home screen.
- HELIX-288: Tight horizontal spacing between action buttons
- HELIX-289: Duplicate 'Play' functionality and buttons displayed.
- HELIX-291: Video player area displays a blank black screen when paused at start.
- HELIX-292: Redundant Play buttons
- HELIX-293: Missing video progress bar
- HELIX-294: Missing time display (elapsed/remaining/total)
- HELIX-295: Multiple redundant play buttons displayed on screen.
- HELIX-297: Critical video playback control (progress bar/timeline) is missing.
- HELIX-302: Missing movie artwork or video player area
- HELIX-304: Lack of clear visual hierarchy for primary action ('Play' button)
- HELIX-308: Content card is partially truncated without clear scroll indication
- HELIX-310: "Recently Added Movies" section displays generic placeholder images instead of actual movie covers.
- HELIX-312: Generic placeholder cards displayed for 'Recently Added Movies'.
- HELIX-315: Incomplete language selection UI
- HELIX-323: Search button lacks clear visual affordance.
- HELIX-325: Call-to-action in instructions is slightly redundant and clunky.
- HELIX-327: Search suggestions show inconsistent relevance and character processing for ambiguous input.
- HELIX-329: Missing visual feedback for spacebar presses on virtual keyboard
- HELIX-331: Redundant elements for initiating a search action.
- HELIX-332: Inconvenient placement of the 'Search' button relative to the virtual keyboard (assuming mobile context).
- HELIX-336: Redundant search suggestion after no results are found.


## Reproduction Steps

1. Navigate to the search screen.
2. Tap on the search input field.

## Evidence

The screenshot shows the search input field containing 'movies.' while the on-screen keyboard below it displays only numbers (1-0) and symbols (@, #, $, -, +, etc.), with no alphabetic character keys visible. The highlighted key is also a period (.), indicating user interaction with a symbol.
