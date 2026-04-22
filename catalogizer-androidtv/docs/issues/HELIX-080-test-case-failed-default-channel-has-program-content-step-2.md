---
id: HELIX-080
severity: critical
category: functional
platform: androidtv
screen: 
status: open
found_date: 2026-04-22
---

# Test Case Failed: Default channel has program content - Step 2

Step: Verify programs
Action: keypress: KEYCODE_DPAD_RIGHT
Expected: Program cards with media titles and images displayed (up to 30 items)
Actual: ACTUAL: The image shows a home screen with various apps and media titles, but it does not display up to 30 program cards with media titles and images.

## Related Issues

- HELIX-029: Test Case Failed: Catalogizer Picks channel appears on TV home - Step 1
- HELIX-030: Test Case Failed: Default channel has program content - Step 1
- HELIX-031: Test Case Failed: Deep link opens correct media detail - Step 1
- HELIX-032: Test Case Failed: Channels cleaned up on logout - Step 1
- HELIX-033: Test Case Failed: Deep link without auth redirects to login - Step 1
- HELIX-034: Test Case Failed: Deep link while app is in background - Step 2


## Acceptance Criteria

Step 'Verify programs' executes successfully and produces the expected outcome: Program cards with media titles and images displayed (up to 30 items)

