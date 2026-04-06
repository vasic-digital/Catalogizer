---
id: HELIX-160
severity: critical
category: functional
platform: androidtv
screen: Layout
status: resolved
resolution: fixed
fixed_date: 2026-04-06
found_date: 2026-03-30
---

# App crash detected during test: Security Test

Stack trace: 03-30 20:54:45.022 27870 22736 E AndroidRuntime: FATAL EXCEPTION: Thread-107
03-30 20:54:45.022 27870 22736 E AndroidRuntime: java.lang.NullPointerException: Attempt to invoke virtual method 'int java.lang.String.length()' on a null object reference

Log entries: [03-30 20:54:45.022 27870 22736 E AndroidRuntime: FATAL EXCEPTION: Thread-107 03-30 20:54:45.022 27870 22736 E AndroidRuntime: java.lang.NullPointerException: Attempt to invoke virtual method 'int java.lang.String.length()' on a null object reference]

## Related Issues

- HELIX-155: App crash detected during test: Login Form Test
- HELIX-156: App crash detected during test: Register Form Test
- HELIX-157: App crash detected during test: Layout Test
- HELIX-158: App crash detected during test: Navigation Test
- HELIX-159: App crash detected during test: Entity Detail Test


