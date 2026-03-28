---
id: HELIX-252
severity: high
category: accessibility
platform: 
screen: androidtv-curiosity-010.png
status: resolved
found_date: 2026-03-28
---

# Question mark icon lacks alternative text or description

The question mark icon does not have any associated alt text or ARIA label, making it inaccessible to screen reader users. This prevents users with visual impairments from understanding the purpose of the icon or the screen.

## Reproduction Steps

Use a screen reader to navigate to the empty state screen.

## Evidence

No visible or programmatically determinable text describing the question mark icon.
