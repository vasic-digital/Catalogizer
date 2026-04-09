---
id: HELIX-024
severity: high
category: functional
platform: video-frame
screen: frame_0002.png
status: open
found_date: 2026-04-08
---

# Server URL displays an invalid port number while indicating a successful connection.

The 'Server URL' input field displays `http://192.160.0.2:80000`. Port numbers are reserved and can only range from 0 to 65535. The displayed port `80000` is outside this valid range. Despite this invalid port, the UI indicates a successful connection with a green checkmark icon and the 'Connected to server' text. This discrepancy is functionally misleading and suggests a potential issue with either URL validation, connection logic, or status reporting.

## Evidence

The text `http://192.160.0.2:80000` within the 'Server URL' field, accompanied by the green checkmark icon to its right and the 'Connected to server' text below it.
