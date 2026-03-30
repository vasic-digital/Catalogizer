---
id: HELIX-097
severity: low
category: ux
platform: video-frame
screen: frame_0003.png
status: wontfix
found_date: 2026-03-30
---

# Lack of Feedback on Password Strength

The application does not provide any feedback to the user about the strength of their password, which may lead to weak passwords being used.

## Related Issues

- HELIX-093: Password Field Not Displaying Correctly
- HELIX-094: Server Connection Field Not Displaying Correctly
- HELIX-095: Username Field Not Displaying Correctly
- HELIX-096: Sign In Button Not Displaying Correctly


## Reproduction Steps

Enter a weak password in the password field.

## Evidence

No password strength indicator is visible next to the password field.

## Resolution

Enhancement suggestion from automated QA vision analysis. Password fields follow standard Android TV Material Design patterns. Password visibility toggle is a backlog enhancement, not a defect.
Closed: 2026-03-30
