---
id: HELIX-095
severity: high
category: content
platform: 
screen: androidtv-curiosity-003.png
status: open
found_date: 2026-03-27
---

# Server URL field contains a placeholder with sensitive/incorrect-looking content

The 'Server URL' field contains a placeholder text that appears to simulate a sensitive or confidential string (e.g., 'testpassword'), which could be misleading or harmful in test environments.

## Evidence

Placeholder text in 'Server URL' contains 'http://localhost:8080testpassword'.
