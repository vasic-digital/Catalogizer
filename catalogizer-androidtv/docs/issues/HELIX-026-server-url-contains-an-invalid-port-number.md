---
id: HELIX-026
severity: high
category: functional
platform: video-frame
screen: frame_0003.png
status: open
found_date: 2026-04-08
---

# Server URL contains an invalid port number.

The specified Server URL `http://192.160.0.2:80000` includes port `80000`. Standard TCP/UDP port numbers range from 0 to 65535. An invalid port number like this will prevent a legitimate connection from being established, making the URL functionally incorrect.

## Related Issues

- HELIX-024: Server URL displays an invalid port number while indicating a successful connection.


## Evidence

The text 'http://192.160.0.2:80000' visible in the 'Server URL' input field.
