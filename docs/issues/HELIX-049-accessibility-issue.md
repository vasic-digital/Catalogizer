---
id: HELIX-049
severity: critical
category: accessibility
platform: 
screen: androidtv-curiosity-033.png
status: wontfix
found_date: 2026-03-30
---

# Accessibility issue

The application does not provide adequate accessibility features for users with disabilities, which can limit their ability to use the application effectively.

## Related Issues

- HELIX-012: Insufficient color contrast
- HELIX-022: Insufficient Color Contrast
- HELIX-027: Inadequate Color Contrast


## Evidence

The application does not provide adequate accessibility features, as evidenced by the lack of screen reader support and high contrast mode.

## Resolution

Not a defect — closed on source evidence. The Android TV client is a Jetpack Compose for TV app (there is no Leanback browse fragment). The claimed gaps are not borne out by the code: interactive elements carry `contentDescription` / semantics for screen readers, e.g. `SearchScreen.kt:130,139`, `ui/screens/login/LoginScreen.kt:397,451`, `ui/components/TopBar.kt:109,146` (33 occurrences across the UI), and the dark theme contrast is ~13:1 AAA (`ui/theme/Theme.kt:33-36,42-46`). The "no screen reader support / high contrast" claim is contradicted by the implemented semantics and theme.

Runtime verification still outstanding (NEEDS-RUNTIME): an actual TalkBack screen-reader pass confirming each control announces correctly has not been captured; closure rests on the presence of `contentDescription`/semantics in code, not on an observed screen-reader run.
Closed: 2026-03-30
Rationale corrected 2026-06-23 (§11.4.7: prior rationale cited a nonexistent Leanback browse fragment / generic contrast boilerplate).
