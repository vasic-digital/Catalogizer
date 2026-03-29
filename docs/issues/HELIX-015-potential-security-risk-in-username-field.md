---
id: HELIX-015
severity: low
category: functional
platform: 
screen: androidtv-004-loginform.png
status: fixed
found_date: 2026-03-29
---

# Potential security risk in 'Username' field

The 'Username' field is set to 'media collection manager', which may be a common username for the application. This could potentially expose users with similar usernames to increased risk of unauthorized access or password guessing.

## Related Issues

- HELIX-006: Incorrect Default Login Credentials


## Reproduction Steps

Navigate to the login screen and observe the default username value.

## Evidence

The default 'Username' value is set to a common phrase for the application's intended use, which may not be secure in a production environment.
