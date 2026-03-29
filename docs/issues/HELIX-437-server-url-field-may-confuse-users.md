---
id: HELIX-437
severity: medium
category: ux
platform: 
screen: androidtv-005-loginform.png
status: fixed
found_date: 2026-03-29
---

# Server URL field may confuse users

The 'Server URL' field currently provides 'http://localhost:8080' as the URL, which might confuse users if not applicable to their setup or if they are unfamiliar with localhost.

## Evidence

The Server URL field is pre-populated with potentially irrelevant data ('http://localhost:8080'), which might not suit all users' needs.
