---
id: HELIX-477
severity: low
category: accessibility
platform: 
screen: androidtv-curiosity-031.png
status: fixed
found_date: 2026-03-29
---

# Insufficient Alt Text on Images

The alt text provided for images in the catalog is minimal, which can create accessibility issues for users who rely on screen readers. This can prevent these users from fully understanding the content and may hinder their ability to navigate the catalog.

## Related Issues

- HELIX-002: Low contrast between text and background in the sidebar
- HELIX-004: Insufficient color contrast on toggle switches
- HELIX-005: No visible focus indicator for interactive elements
- HELIX-009: Small text size
- HELIX-012: Lack of accessible contrast in text and background
- HELIX-015: Missing clear focus indication
- HELIX-019: Low contrast text in Customize section description
- HELIX-020: Missing alt text or labels for app icons
- HELIX-027: Low contrast text in instructional sections
- HELIX-028: Missing alt text or labels for app icons
- HELIX-035: Low contrast text in Customize section description
- HELIX-036: Missing alt text or labels for app icons
- HELIX-044: Low contrast text in sections
- HELIX-045: Missing alt text or labels for app icons
- HELIX-053: Low contrast text in sections
- HELIX-058: Low contrast text in option descriptions
- HELIX-067: Low contrast text in option descriptions
- HELIX-073: Low contrast text in option descriptions
- HELIX-079: Low contrast text
- HELIX-083: No text alternative for loading animation
- HELIX-086: Insufficient contrast between text and background
- HELIX-090: Insufficient color contrast between text and background
- HELIX-091: Small input fields reduce usability
- HELIX-094: Input fields lack visible labels or instructions for assistive technologies
- HELIX-097: Lack of labeling for the 'Server URL' field
- HELIX-101: Low contrast for placeholder text in input fields
- HELIX-102: Missing labels or ARIA attributes for interactive elements
- HELIX-110: Low contrast for input field text and labels
- HELIX-111: Missing form field labels for screen readers
- HELIX-121: Low contrast for placeholder text in input fields
- HELIX-122: Missing labels or ARIA attributes for interactive elements
- HELIX-131: Low contrast for placeholder text in input fields
- HELIX-132: Missing labels or ARIA attributes for input fields
- HELIX-142: Low contrast for placeholder text in input fields
- HELIX-143: Missing labels or ARIA attributes for interactive elements
- HELIX-150: Low contrast for placeholder text and input borders
- HELIX-151: Missing labels or ARIA attributes for interactive elements
- HELIX-157: Missing focus indicators for keyboard navigation
- HELIX-160: Missing input field labels for screen readers
- HELIX-161: Low contrast for input field text
- HELIX-169: Low contrast between text and background
- HELIX-174: Low contrast between text and background
- HELIX-179: Low contrast between loading spinner and background
- HELIX-180: Missing loading state announcement for screen readers
- HELIX-185: Low contrast between icon and background
- HELIX-190: Lack of text contrast for file details
- HELIX-197: Insufficient contrast for error message text
- HELIX-202: Low contrast between text and background
- HELIX-207: Low contrast between text/button and background
- HELIX-212: Missing labels or ARIA attributes for input fields
- HELIX-213: Low contrast for placeholder text
- HELIX-221: Low contrast for input labels
- HELIX-222: No focus indicators on input fields
- HELIX-231: Low contrast for placeholder text
- HELIX-232: Missing form field labels
- HELIX-239: Low contrast for input field labels
- HELIX-240: No visible focus state for input fields
- HELIX-247: Low contrast loading indicator
- HELIX-248: Missing screen reader support for loading state
- HELIX-252: Question mark icon lacks alternative text or description
- HELIX-258: Low contrast for file size text
- HELIX-262: No alt text or labels for thumbnails
- HELIX-266: Low contrast and small icon size reduce accessibility
- HELIX-269: Missing screen reader support for empty state
- HELIX-272: Lack of text contrast for filenames and metadata
- HELIX-276: Missing alt text or labels for thumbnails
- HELIX-280: Low contrast between text and background
- HELIX-285: Low contrast between text and background
- HELIX-291: Low contrast for secondary text
- HELIX-298: Low contrast between text and background
- HELIX-304: Low contrast on file size text
- HELIX-305: Missing alt text or descriptions for thumbnails
- HELIX-313: Low contrast between text and background
- HELIX-319: Low contrast for input field text and labels
- HELIX-320: Missing form field labels for screen readers
- HELIX-329: Low contrast for input field text and labels
- HELIX-330: Missing labels or ARIA attributes for interactive elements
- HELIX-338: Missing accessibility labels for input fields
- HELIX-344: Low contrast for input field text
- HELIX-348: Lack of text contrast for file metadata
- HELIX-352: Missing alt text or labels for thumbnails
- HELIX-356: Low contrast in text and UI elements
- HELIX-364: Low contrast on text and icons
- HELIX-368: Missing alt text or labels for thumbnails
- HELIX-369: Low contrast between text and background
- HELIX-373: Missing alt text for key visual elements
- HELIX-375: Low contrast text in game title and metadata
- HELIX-379: Missing alt text or labels for interactive icons
- HELIX-381: Low contrast between text and background
- HELIX-385: No visible focus states for interactive elements
- HELIX-390: Low contrast text on movie cards
- HELIX-393: Missing alt text or labels for movie thumbnails
- HELIX-398: Low contrast for file metadata text
- HELIX-402: No alt text or labels for thumbnails
- HELIX-405: Insufficient color contrast for text and buttons
- HELIX-410: Insufficient color contrast for text
- HELIX-415: Low contrast between text and background
- HELIX-421: Lack of descriptive placeholder text in input fields
- HELIX-425: Server URL text is not easily visible
- HELIX-428: Low contrast between text and background
- HELIX-430: Missing labels for input fields
- HELIX-435: Password field lacks a visible label
- HELIX-436: Lack of text contrast in the UI
- HELIX-441: Missing descriptive username and password field labels
- HELIX-442: Missing accessible contrast ratio for text
- HELIX-446: Eye icon lacks descriptive alt text
- HELIX-454: Lack of Keyboard Navigation Support


## Reproduction Steps

Examining the images in the catalog with focus on the alt text descriptions.

## Evidence

The alt text for images is minimal, often just describing the image as a whole without providing specific details. This can make it difficult for screen-reader users to understand the content of individual images.
