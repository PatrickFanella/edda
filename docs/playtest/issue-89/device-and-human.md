# Device and human qualification status

Observed November 15, 2026 UTC on the `c0383a2` application baseline.
The visible normal software-keyboard session and final human review are
**INCONCLUSIVE / not performed**. No device pass is claimed.

## Fresh Android attempt

`device_list` reported one local Android target, `T3_Pixel_API_35`, Android 15,
initially not booted. It reported no available iOS platform: this Linux host
cannot run iOS Simulators. `device_open` booted the target and supplied a new
command bound to this thread's own session. Its first Chrome launch failed:

```text
Error (DEVICE_IN_USE): Device is already in use by session "t3-923a17297ea61926f89c3f06".
Diagnostic ID: mwif82nc-54fd774f
```

The suggested session-list inspection confirmed that exact occupying session
on `emulator-5554`. That is another session, so it was neither reused nor
closed. No UI input, Settings changes, keyboard interaction, device screenshot
or recording was performed. No adb, simctl or SSH UI workaround was used.
The tool's successful device-open response alone is not a keyboard result.

The earlier #78 ownership blocker is therefore still present in this fresh
attempt. Keep #46/#78 open. Desktop Chromium touch emulation, synthetic viewport
changes and root-text reflow do not replace this checkpoint.

## Shared browser attempt

A fresh incognito shared-browser tab reached a temporary local fixture login,
logged in with synthetic credentials and rendered the saved transcript at
390×844. Initial stale references followed development-server reloads; the
server was then restarted with watching/HMR disabled. The temporary fixture
omitted current versioned encounter-preview endpoints and showed 501 errors.
This establishes neither real mechanics nor an end-to-end manual playthrough.
Those fixture errors are not recorded as production application defects.

The browser tab was closed. The temporary API process had exited; the recorded
Vite process started for this attempt was stopped by its own PID. No other
browser tab, process or device session was closed. The reproducible successful
browser evidence is the Playwright suite described in [browser.md](browser.md).

## Remaining human checkpoint

No human reviewer/date is recorded for this candidate. The following work
remains explicitly unqualified:

- A visible normal phone keyboard: focus the composer, type a multiline draft,
  reach Send and choices with the keyboard open, dismiss and reopen it, and
  inspect browser chrome and safe-area changes.
- Context retention: open and close Character/World/Journal; check draft,
  selection, reading position and focus while the keyboard is open.
- Portrait and landscape saves/export, text size, browser zoom and controls.
- A real assistive-technology session, including spoken labels, focus order and
  recovery announcements; semantic role assertions alone are not that session.
- Final human assessment of the declared supported player journey, including
  the functional qualification findings in [README.md](README.md).

The device/OS/browser, orientation, keyboard, zoom/text settings, outcome and
reviewer/date must accompany any later completion claim. Existing curated
chapter feedback from #87 remains valid for its named examples; it does not
supply a new human verdict for this candidate or arbitrary generated prose.
