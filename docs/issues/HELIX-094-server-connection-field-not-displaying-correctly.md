---
id: HELIX-094
severity: medium
category: ux
platform: video-frame
screen: frame_0002.png
status: wontfix
found_date: 2026-03-30
---

# Server Connection Field Not Displaying Correctly

The server connection field is not displaying correctly, with the URL not being fully visible. This makes it difficult for users to enter the correct server connection.

## Related Issues

- HELIX-093: Password Field Not Displaying Correctly


## Reproduction Steps

Open the login page and attempt to enter a server connection.

## Evidence

The URL is not fully visible in the server connection field.

## Resolution

Server connection URL field is intentional for multi-server support. Users configure their catalog-api server address during initial setup. UX improvements for this field are tracked in the product backlog.
Closed: 2026-03-30
