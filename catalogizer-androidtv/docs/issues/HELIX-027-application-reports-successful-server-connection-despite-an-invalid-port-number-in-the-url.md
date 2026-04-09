---
id: HELIX-027
severity: critical
category: functional
platform: video-frame
screen: frame_0003.png
status: open
found_date: 2026-04-08
---

# Application reports successful server connection despite an invalid port number in the URL.

Despite the 'Server URL' displaying an invalid port number (`80000`), the application shows a green checkmark icon next to the URL and the message 'Connected to server' in green text. This provides false positive feedback to the user, implying a connection has been established when it is functionally impossible with the given port. This can lead to significant user confusion and inability to use the application.

## Related Issues

- HELIX-024: Server URL displays an invalid port number while indicating a successful connection.
- HELIX-026: Server URL contains an invalid port number.


## Evidence

The text 'http://192.160.0.2:80000' in the 'Server URL' field, combined with the green checkmark icon and the green text 'Connected to server' at the bottom of the screen.
