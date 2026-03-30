---
id: HELIX-012
severity: low
category: ux
platform: 
screen: androidtv-003-api-endpoints.png
status: wontfix
found_date: 2026-03-30
---

# Unclear Server Connection Status

The server connection status is not clearly indicated, leaving users uncertain about the connection quality.

## Related Issues

- HELIX-001: Unclear Form Field Labels
- HELIX-002: Missing Password Masking
- HELIX-003: Lack of Form Validation Feedback
- HELIX-004: Inconsistent Button Styling
- HELIX-006: Lack of Feedback on Form Submission
- HELIX-008: Inconsistent Input Field Lengths
- HELIX-009: Lack of Clear Error Messages
- HELIX-010: Insufficient Feedback on Form Submission
- HELIX-011: Missing Password Visibility Toggle


## Reproduction Steps

None

## Evidence

The server connection status is represented by a URL, but its meaning is unclear to non-technical users.

## Resolution

Server connection URL field is intentional for multi-server support. Users configure their catalog-api server address during initial setup. UX improvements for this field are tracked in the product backlog.
Closed: 2026-03-30
