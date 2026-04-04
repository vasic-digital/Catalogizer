# [MAJOR] Android TV Search: Text input accumulates instead of replacing

**Platform**: Android TV (Mi Box 192.168.0.214:5555)
**Screen**: Search screen
**Severity**: MAJOR
**Discovered by**: HelixQA autonomous session (curiosity phase steps #17-#33)

## Description

When typing search queries on the Android TV search screen, the `clear` command via ADB does not properly clear the existing text. Subsequent `type` commands append to the existing garbled text instead of replacing it. After multiple search attempts, the search field shows accumulated text like `a-aMovies movies a200a1a-aaMovieseuii`.

## Reproduction Steps

1. Navigate to Search screen via DPAD
2. Type a search query using `adb shell input text "Matrix"`
3. Attempt to clear with `adb shell input keyevent KEYCODE_CLEAR` or select-all + delete
4. Type a new query
5. Observe: new text is appended to old text

## Expected Behavior

The `clear` action should fully clear the search input field. New `type` commands should replace any existing text.

## Actual Behavior

Text accumulates in the search field. Multiple clear+type attempts result in concatenated garbage text (`a-aMovies movies a200a1a-aaMovieseuii`).

## Evidence

- Screenshots: `androidtv-curiosity-020.png` through `androidtv-curiosity-036.png`
- Video: `androidtv-session.mp4`
- HelixQA log: Steps #17-#33 show repeated failed search attempts

## Root Cause (suspected)

The search field's `EditText` may not properly handle `KEYCODE_CLEAR` or the focus/selection state may prevent text replacement. The ADB `input text` command always appends at cursor position.

## Fix Suggestion

In `HelixQA/pkg/navigator/adb.go`, the `clear` action should use:
```
adb shell input keyevent KEYCODE_MOVE_HOME
adb shell input keyevent --longpress $(for i in $(seq 1 100); do echo -n "KEYCODE_DEL "; done)
```
Or in the Android TV app: implement proper `setText("")` handling for the search field when receiving clear intent.
