---
id: HELIX-051
severity: critical
category: accessibility
platform: 
screen: androidtv-curiosity-014.png
status: wontfix
found_date: 2026-03-30
---

# Lack of ARIA Attributes for Screen Readers

The form elements do not include ARIA attributes, making it difficult for screen readers to interpret and convey the content to visually impaired users.

## Related Issues

- HELIX-006: Insufficient Color Contrast


## Reproduction Steps

Use a screen reader to navigate the form

## Evidence

No ARIA attributes present

## Resolution

Not a defect — closed on source evidence. The Android TV client is a Jetpack Compose for TV app (there is no Leanback browse fragment); "ARIA attributes" are a web concept — the Android equivalent is `contentDescription` / semantics, which ARE present. Interactive elements carry `contentDescription` across the UI, e.g. `SearchScreen.kt:130,139`, `ui/screens/login/LoginScreen.kt:397,451`, `ui/components/TopBar.kt:109,146` (33 occurrences across the UI). The "no ARIA attributes" claim is contradicted by the implemented semantics.

Runtime verification still outstanding (NEEDS-RUNTIME): an actual TalkBack screen-reader pass confirming each control announces correctly has not been captured; closure rests on the presence of `contentDescription`/semantics in code, not on an observed screen-reader run.
Closed: 2026-03-30
Rationale corrected 2026-06-23 (§11.4.7: prior rationale cited a nonexistent Leanback browse fragment / mismatched contrast boilerplate, and the section was duplicated).
