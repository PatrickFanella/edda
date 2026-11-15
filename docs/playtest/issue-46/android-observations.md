# Android device-session evidence

Observed 2026-11-15 UTC, approximately 01:32–01:44. Baseline frontend `b84d5dc8bde8438920c867c2dab4d64c50c001b5`; isolated HTTP/WebSocket fixtures on ports 39246/39247, no real backend.

- Device tool host: `local`; requested AVD: `T3_Pixel_API_35`; opened device: `emulator-5554`.
- Device: Android 15.0, emulator, not physical hardware. Device screenshot is 1080 × 2400 physical pixels; no claim that those are CSS viewport dimensions.
- Browser: installed `com.android.chrome`, default portrait. Exact Chrome version and device CSS viewport were not collected. No viewport metrics were inferred from screenshot pixels.
- The device was initially stopped and was booted by `device_open`. Every app launch, setting interaction, tap, text fill, scroll and screenshot used the exact returned `agent-device` command; `device_screenshot` was used for visual inspection. No adb UI automation or independent browser/device session was used.
- Opened `http://10.0.2.2:39247/login`; filled synthetic email/password, tapped Sign in, and observed the fixture campaign list. Opened `/play/alpha` and observed the three historical entries and Mira's vitals. Scrolling was needed to expose Your move and Send action.
- Filled `Android fixture action`, tapped Send action, and observed both that player entry and `You find a silver key.` in the accessibility snapshot. This is actual Android touch/fixture-result evidence.

## Keyboard finding: INCONCLUSIVE

The input field could focus and receive text, but no visible on-screen keyboard appeared. The tool's keyboard status after focus/fill and a later tap reported:

```text
Keyboard visible: no
Input type: text
Input owner: app
Input method: com.callstack.agentdevice.imehelper
Focused package: com.android.chrome
```

This is the agent helper IME, not Gboard keyboard occlusion/resizing evidence. [The captured device screenshot](android-helper-ime.png) shows the UI without a software keyboard.

The bounded investigation tried Android Settings through the device UI: searched for keyboard, opened Physical keyboard, and confirmed “Use on-screen keyboard” was already enabled. Opened On-screen keyboard, where Gboard and Agent Device Test IME were enabled, and toggled the helper IME off. After reopening Chrome and tapping Your move, keyboard status still identified the helper and reported no keyboard. Subsequent fixture text entry used the helper again. A visible Gboard session was not established; the test does not claim the setting change produced one. No shell-based UI workaround was attempted. Final Settings readback showed Gboard, Agent Device Test IME and Google Voice Typing all checked, matching the initial enable states; the on-screen-keyboard option was not changed.

The real visible-software-keyboard gate remains pending. A human reviewer must record a suitable visible device session with normal keyboard open/close, browser chrome changes, action/choice reachability, multiline draft retention, detail return/focus, and saves/export in portrait/landscape. This emulator run does not satisfy that gate, and no human sign-off occurred here.
