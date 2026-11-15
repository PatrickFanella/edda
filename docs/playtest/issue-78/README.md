# Issue #78: viewport regression qualification

Observed 2026-11-15 UTC, starting from `cf0a61f` with this PR's small touch-target/combat-input fixes. Related: [#78](https://github.com/subculture-collective/edda/issues/78), [#46 baseline report](../issue-46/README.md), [roadmap #47](https://github.com/subculture-collective/edda/issues/47).

**Automated viewport qualification passes. The actual software-keyboard and human checkpoint remains blocked; keep #78 and #46 open.** This is a draft implementation PR, not production or rules qualification. No production account, database, inference provider or deployment was used. #59's narration recovery and #61's rules mechanics are outside this change.

## Automated evidence

[Browser results and geometry](browser-results.json) retain all 46 test outcomes and 51 measurements from the passing full Chromium run. There were no retries, skips or flaky results. The new qualification cases extend #45's existing HTTP/WebSocket fixture harness. Unknown API routes fail and external traffic is blocked. The combat test uses `light` rules mode and the existing mock combatants; it proves control layout/dispatch, not authentic combat state or spell resolution.

The checks combine document dimensions (1px rounding allowance), intersection with `visualViewport` and every ancestor that clips either axis, center hit testing, enabled/disabled states and at least 44×44px control rectangles. Fixed narrative input/Send are measured before any actionability scroll. A negative control proves the geometry probe rejects a control hidden behind an overflow ancestor or another element even when the document fits.

| Default viewport | Observed document | Document scroll | Send bottom, long choices | Send target | Result |
| --- | --- | --- | ---: | --- | --- |
| 1366×768 | 1366×768 | (0, 0) | 620 | 185.5×44, hittable | PASS |
| 1440×900 | 1440×900 | (0, 0) | 752 | 185.5×44, hittable | PASS |
| 1920×1080 | 1920×1080 | (0, 0) | 932 | 185.5×44, hittable | PASS |
| 768×1024 | 768×1024 | (0, 0) | 876 | 185.5×44, hittable | PASS |
| 360×800 | 360×800 | (0, 0) | 652 | 304×44, hittable | PASS |
| 390×844 | 390×844 | (0, 0) | 696 | 334×44, hittable | PASS |

At every default size the matrix covers:

- Long transcript and eight choices, each with twenty repeated descriptive clauses, alongside a multiline draft. Wheel input changes the transcript/choice owner's scroll position while all its ancestors and the document remain stationary. The final choice can be exposed and hit; input stays at least 17px and Send remains reachable.
- Pending dispatch with input/Send disabled but visible and navigation usable, followed by an oversized validation error inside its own scrolling feedback region and a corrected action/result.
- Character, World and Journal details; expanded Journal Logs supply long detail content. Closing each detail restores the draft, text selection, input focus and exact transcript scroll position. Detail scrolling leaves its ancestors stationary.
- Twenty saves in their bounded list, reachable Send with saves expanded, all three export-format controls, Close and a fixture JSON download. These are read/download paths; durable save creation/restoration and real exports are not qualified.
- The current combat presentation, all five action buttons and an expanded spell input with Send dispatching exactly `I cast Light` to the isolated fixture.

Long choices can be taller than their scroll owner, so the test requires a visible, hittable 44px portion after scrolling. Other controls must fit completely, allowing at most 1px rounding. Horizontal phone navigation deliberately scrolls; each destination is exposed and checked individually. Saves, combat and short/enlarged-text workspaces also use internal scrolling. This report does not claim that every control stays simultaneously visible in those states. The document stays stationary throughout.

Screenshots retained from the passing run:

- [1366×768 long choices](desktop-long-choices.png)
- [390×844 long choices](phone-long-choices.png)
- [390×844 expanded combat input](phone-combat.png)
- [390×844 export](phone-export.png)
- [200% text in a 683×384 workspace](enlarged-text.png)

The screenshots were visually inspected. They show the viewport at one point in each test; the assertions and measurements supply reachability evidence across the sequence. Font requests are blocked in the automated harness, so these use local fallbacks. Real font loading is not established by these screenshots.

## Minimal production fixes

The new probe measured Close details at 32.78125px high before the fix. It is now at least 44px. Combat action/Send/Cancel and export Close receive the same minimum target size. Expanded spell/item input gets a full row on phones and can shrink within the desktop row; wrapping avoids squeezing the buttons beyond the narrow pane. Its text increases from `text-sm` to `text-base`. No text or target is reduced to obtain fit. These are presentation changes; mechanical dispatch semantics are unchanged.

## Touch, focus and text reflow

Two separate Chromium touch-emulation cases (`hasTouch`/`isMobile`, at 360 and 390px) tap input, World, Close details, Send and a suggested choice. They check retained multiline draft, restored input focus and exact dispatched actions. **These are desktop touch emulation, not actual device or keyboard evidence.** The six-size context matrix separately verifies focus, draft, selection and reading-position restoration.

At 200% root text, both phone sizes and a constrained 683×384 workspace retain reachable input, Send and every navigation destination through internal scrolling. Textarea content may occupy only part of the short workspace but retains a hittable region of at least 44×44px and its own text scrolling. Send must fit completely. These checks establish enlarged-text reflow; they do not establish native browser zoom, pinch panning, assistive-technology behavior or OS text scaling.

The earlier synthetic visual-viewport checks remain in `viewport.spec.ts`, including resize/cleanup and a 480px workspace with saves open. Synthetic height changes cannot qualify keyboard occlusion, browser chrome or keyboard dismissal/reopening.

## Actual Android session: blocked before app interaction

The requested target was `hostId=local`, `id=emulator-5554`, T3 Pixel API 35, Android 15. `device_open` succeeded and returned a command bound to this thread's own session. The first app launch through that exact command failed:

```text
Error (DEVICE_IN_USE): Device is already in use by session "t3-923a17297ea61926f89c3f06".
Diagnostic ID: mwha0x1e-1a0a9bd9
```

The tool suggested inspecting `agent-device session list`; that read confirmed the occupying session. The task explicitly prohibits reusing or closing another session, so neither was attempted. No app/input/Settings interaction, recording or screenshot was captured from the occupied session. No adb/simctl UI workaround was used. This run supplies **no actual visible-keyboard result**. #46's older hidden-helper-IME attempt also remains inconclusive.

A separate shared-browser attempt reached the isolated login page, but ref-based actions repeatedly returned `This element ref is stale` after fresh snapshots. That tab was closed. The automated Playwright run is the successful browser evidence; the shared-browser attempt is not recorded as a passed login or play session. The disposable fixture API (PID 3775940) and this worktree's Vite process were stopped; no other session, server or process was stopped.

## Recommended resolution for #46

Keep #46 open for human review of its baseline findings and a visible session with a normal on-screen keyboard. The desktop/narrow layout concerns now have passing post-overhaul regression coverage, but that does not discharge its human gate. Keep this PR in draft and do not use a closing keyword for #78 until that checkpoint is satisfied.

When the emulator is released by its owner, or another suitable phone session is available, use a fresh `device_open` session and its returned command for all interactions, with screenshots through `device_screenshot`. A human must perform or review: focus input, enter a multiline draft with a **visible normal keyboard**, inspect browser-chrome/viewport changes, reach Send and choices while the keyboard is open, open/close context and verify draft/reading/focus, dismiss/reopen the keyboard, and inspect saves/export in portrait and landscape. Record device/OS/browser, orientation, text size/zoom, failures, screenshots/recording and reviewer/date on #46. Fix any resulting defects and rerun affected checks. Only that recorded review can support resolving #46 and completing #78.

## Reproduction and validation

From the repository root, with the lockfile-pinned dependencies installed:

```sh
rtk proxy pnpm --dir frontend install --frozen-lockfile
rtk proxy pnpm --dir frontend test
rtk proxy env EDDA_FIXTURE_PORT=39388 PLAYWRIGHT_JSON_OUTPUT_NAME=/tmp/edda-78-browser-results.json pnpm --dir frontend test:browser --workers=2 --reporter=line,json
rtk proxy pnpm --dir frontend build
rtk proxy pnpm --dir frontend lint
```

Playwright starts/stops its own fixture frontend. The chosen port must be free. One attempted run aborted safely because another server occupied default port 4175; no existing server was reused or stopped. Fresh runs write screenshots under ignored `frontend/test-results/`; JSON reporter attachments contain the raw geometry used for this report.

| Check | Outcome |
| --- | --- |
| `pnpm --dir frontend test` | PASS: TypeScript harness check, 7 files / 18 tests |
| Full browser suite | PASS: 46 tests, no skips/retries/flaky results; 62.24 seconds |
| `pnpm --dir frontend build` | PASS; existing >500kB chunk warning |
| `pnpm --dir frontend lint` | FAIL: 22 errors in 9 unchanged files; [baseline proof](lint-baseline.json) records rules and SHA-256 hashes, each checked against `cf0a61f` |
| ESLint on new tests and all changed components | PASS |
| Review | Self-review of source/test diff and representative screenshots; no independent review claimed |
| Hosted CI | Separate from local results; read on the PR's exact head at handoff |
| Visible keyboard / browser chrome / human review | BLOCKED; no completion claim |

No Go files changed. Backend persistence, SRD conformance, production, iOS Safari and real assistive technology remain outside this evidence.
