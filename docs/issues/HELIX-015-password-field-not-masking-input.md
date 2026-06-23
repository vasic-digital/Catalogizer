---
id: HELIX-015
severity: critical
category: functional
platform: 
screen: androidtv-curiosity-006.png
status: wontfix
found_date: 2026-03-30
---

# Password Field Not Masking Input

The password field is not masking the user's input, which is a security risk as it allows others to see the password.

## Reproduction Steps

Enter a password and observe the field

## Evidence

The password is visible in plain text in the field

## Resolution

Not a defect — closed on source evidence. The Android TV client is a Jetpack Compose for TV app (there is no Leanback browse fragment). The password field is masked by default, NOT plaintext: it applies `PasswordVisualTransformation()` whenever `passwordVisible` is false (`catalogizer-androidtv/app/src/main/java/com/catalogizer/androidtv/ui/screens/login/LoginScreen.kt:384`), and `passwordVisible` is initialised to `false` (`LoginScreen.kt:88`). Masking is the default render; the optional show/hide toggle (`LoginScreen.kt:390-397`) lets the user reveal it deliberately. The "password visible in plain text" claim is contradicted by the default `passwordVisible=false` + `PasswordVisualTransformation()` wiring; the prior framing of this as a missing "visibility toggle enhancement" was also wrong — the toggle exists and masking is already on.

Runtime verification still outstanding (optional, NEEDS-RUNTIME): a zero-trust closure would capture a first-render screenshot confirming the field shows masking dots before any toggle interaction. Closed wontfix on code evidence; first-render capture remains optional belt-and-suspenders.
Closed: 2026-03-30
Rationale corrected 2026-06-23 (§11.4.7: prior rationale mischaracterised a masked-by-default field as a missing toggle enhancement).
