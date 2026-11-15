# Issue #46: viewport and interaction investigation

Observed 2026-11-15 UTC against `b84d5dc8bde8438920c867c2dab4d64c50c001b5`. Related: [#46](https://github.com/subculture-collective/edda/issues/46), [roadmap #47](https://github.com/subculture-collective/edda/issues/47).

**Decision: the current UI fails the default viewport contract. Proceed with the existing bounded implementation issues; keep #46 open for human review of the viewport/device evidence and an actual visible software-keyboard check.** This investigation changes no application layout, backend, provider routing, rules, or campaign data. It does not qualify production. The owner selected A's layout with B's serif story text in [#55](https://github.com/subculture-collective/edda/issues/55#issuecomment-6887405453), recorded in [draft PR #98](https://github.com/subculture-collective/edda/pull/98). That choice is the implementation direction; this report supplies baseline evidence, not a competing concept.

## Evidence and isolation

- [Root viewport measurements](root-viewport-measurements.json): an unchanged copy supplied by the root agent from its separate browser session, with isolated fetch/WebSocket fixtures, after one turn. Source copy SHA-256: `2cd3ce2404d5aea91b0acc23376b875c173d68763bc17e80a96830182c7c5cd8`. The root checkout, tab and server were not changed.
- [Browser observations](browser-observations.json): this investigation's separate shared-browser tab, using the actual baseline frontend with a local HTTP/WebSocket fixture service. Its initial history has three visible entries, so its heights differ from the root fixture. Do not combine those heights as though they were the same campaign state.
- [Fixture server](fixture-server.mjs): synthetic Adventure alpha/Mira data derived from [PR #95's fixture data and routes at `8b63e10`](https://github.com/subculture-collective/edda/blob/8b63e101e85c738ea7c59dd113804d9bcefba0c5/frontend/tests/browser/fixtures.ts), plus explicit saves/export, long-choice and validation-error responses. It returns 501 for unmapped API paths and never forwards requests. This is a disposable investigation tool, not a second regression framework. The normal frontend may still fetch its public Google Fonts resources.
- Local Vite used port 39247; its `/api` proxy targeted only fixture port 39246. Android Chrome used `http://10.0.2.2:39247`. Dummy account `player@example.test` used only synthetic credentials. No production backend, real account, database or paid inference was used.
- [Desktop long-choice screenshot](desktop-long-choices.png), [phone after-turn screenshot](phone-after-turn.png), [downloaded fixture JSON](fixture-export.json), and [Android observations](android-observations.md) retain supporting evidence. A screenshot is a moment in a sequence, not proof that every control is usable.

## Default viewport reproduction

Root measurements, in CSS pixels, with document scroll reset to the top. Coordinates below are document coordinates rounded to the nearest pixel. “Tabs bottom” describes the tablist, not all horizontally clipped tab buttons.

| Viewport | Document | Vertical excess | Send bottom | Tabs bottom | Result |
| --- | --- | ---: | ---: | ---: | --- |
| 1366 × 768 | 1366 × 844 | 76 | 669 | 742 | FAIL: credit footer below viewport |
| 1440 × 900 | 1440 × 976 | 76 | 801 | 874 | FAIL: credit footer below viewport |
| 1920 × 1080 | 1920 × 1156 | 76 | 981 | 1054 | FAIL: credit footer below viewport |
| 768 × 1024 | 768 × 1752 | 728 | 972 | 1650 | FAIL: navigation below viewport |
| 360 × 800 | 360 × 2016 | 1216 | 1236 | 1914 | FAIL: action and navigation below viewport |
| 390 × 844 | 390 × 1992 | 1148 | 1212 | 1890 | FAIL: action and navigation below viewport |
| 844 × 390 (supplemental landscape) | 844 × 1736 | 1346 | 956 | 1634 | FAIL: rotating does not establish fit |

There was no document-wide horizontal overflow in these measurements. That is useful counterevidence, but horizontal scrolling within the eight-tab strip still hides destinations initially. At 390 px, Narrative and most of Character are visible within the strip; the remaining tabs require horizontal travel. “Below the initial viewport” does not mean permanently unreachable: the narrow-screen document can scroll to the action area and tab strip.

## Interaction assessment

| Surface | Reproduction and observed result | Bound and implication |
| --- | --- | --- |
| Free-text action | At 390 × 844, enter `Look around` and press Enter. Fixture API logs the action and the UI adds `You find a silver key.` | PASS for isolated dispatch/result; FAIL for anchored reachability. This fixture's document grew from 1824 to 2201 px. |
| Choice action | Click `Inspect the key` after the result. Fixture API logs that exact input once in this sequence; a new result renders. | PASS for the observed choice dispatch. Existing choice buttons submit immediately; no new confirmation behavior was introduced. |
| Long choices and history | Submit `long choices`: 180 repeated narrative sentences and eight choices, each with four repeated descriptive clauses. At 390 × 844, the log expands to 3866 px and the document to 7766 px with Saves open. | FAIL: history and choices grow the document. Send bottom ≈6986 and tablist bottom ≈7664 in that particular state. These stress values are fixture-specific, not ordinary campaign estimates. |
| Desktop clipping | Reload at 1366 × 768 and submit `long choices`. Log collapses to 32 px, choices occupy 836 px, input starts at y≈1067, Send starts at y≈1195. Its narrative ancestor ends at y≈687 with `overflow-y:hidden`. Document remains 844 px tall. | FAIL: hiding overflow masks a clipped composer. Check ancestor clipping and hit testing as well as document dimensions. See the screenshot and `desktopClipping` in the JSON. |
| Error | Submit `fixture error`. The synthetic server validation error renders; action input becomes available again. | PASS for error display/recovery in this sequence; FAIL for bounded layout. This was a synthetic server error, not rules validation qualification. |
| Every top-level section | Click Narrative, Character, Inventory, Quests, NPCs, World, Journal and Logs at 390 × 844, 360 × 800, 844 × 390 and 768 × 1024. Each selection renders its matching fixture/empty state. | PASS for selection via scrolling/automation; FAIL for viewport fit. Browser automation scrolls targets into view; it does not establish touch discoverability. Empty lists do not qualify populated item/NPC/quest details. |
| World and Journal | At 390 × 844, open Map, Facts, Codex and Relationships, and Journal Chronicle/Notes. All render their expected empty states or note editor. | PASS for those destinations. Codex opens its Languages category; the other Codex categories were not individually exercised. |
| Draft, reading position and focus | Type `Keep this draft`, switch away and return: draft becomes empty. In a separate desktop probe, set log scrollTop to 100, switch Character → Narrative: immediate scrollTop becomes 0 and focus is the new textarea. | FAIL for retained context. The remounted log may then auto-follow; it does not restore the recorded reading position. Text selection and assistive-technology behavior remain unqualified. |
| Saves | Toggle Saves at 390 × 844. `Fixture checkpoint` renders in a bounded save-list wrapper. Opening it adds 98 px in this short-list state. | PASS for opening/reading. Creating a save through the native prompt was inconclusive: the browser evaluate call timed out on the prompt, and no matching save POST was observed. No durable save/restore claim. |
| Export | Open Export from the user menu at 390 × 844, 360 × 800, 768 × 1024 and 844 × 390. All three formats and Close are present. A 624-byte fixture JSON downloads and parses at tablet size. | Portrait controls fit. Landscape dialog has 422 px scroll content in a 346 px client area; Close starts at y393, then moves to y317 after internal scrollTop=76 and can be clicked. Counterevidence: landscape Close is scroll-reachable. Other export formats and a real campaign export were not qualified. |
| Phone software keyboard | Booted Android 15 emulator, signed into the fixture, scrolled to input, filled and tapped Send; fixture result rendered. The session input method stayed `com.callstack.agentdevice.imehelper`, reporting `Keyboard visible: no`. | PASS for this emulator touch/fixture-action path. INCONCLUSIVE for visible keyboard resizing/occlusion. A hidden test IME is not software-keyboard evidence. |

## Source investigation and counterevidence

Links below pin the inspected source so later implementation changes do not silently change this report's meaning.

1. [AppShell lines 18–43](https://github.com/subculture-collective/edda/blob/b84d5dc8bde8438920c867c2dab4d64c50c001b5/frontend/src/components/layout/AppShell.tsx#L18-L43) supplies viewport height and hidden overflow only at `xl`; narrower sizes use minimum height. [CampaignPlayPage lines 284–315](https://github.com/subculture-collective/edda/blob/b84d5dc8bde8438920c867c2dab4d64c50c001b5/frontend/src/pages/CampaignPlayPage.tsx#L284-L315) stacks narrative, the still-present sidebar and navigation below that breakpoint. This explains document growth; a one-column fallback already exists, but it does not satisfy the owner's contract.
2. The [credit footer outside `#root`](https://github.com/subculture-collective/edda/blob/b84d5dc8bde8438920c867c2dab4d64c50c001b5/frontend/index.html#L32-L34) follows a full-viewport main. [Credit CSS](https://github.com/subculture-collective/edda/blob/b84d5dc8bde8438920c867c2dab4d64c50c001b5/frontend/public/brand/edda/credit.css#L1) adds 32 px vertical padding around a 44 px link. That explains the measured 76 px desktop excess. Preserve the credit while budgeting its space; removing identity content is not this spike's decision.
3. [ContainedTabPanel and NarrativeTab](https://github.com/subculture-collective/edda/blob/b84d5dc8bde8438920c867c2dab4d64c50c001b5/frontend/src/pages/CampaignPlayPage.tsx#L392-L473) already request internal scrolling and put choices/input outside the transcript. These are useful existing boundaries. They fail when their ancestor lacks a bounded narrow-screen height, or when long non-shrinking content exceeds the desktop frame. Merely adding `overflow-hidden` would make the desktop clipping failure worse.
4. [PlayerInput](https://github.com/subculture-collective/edda/blob/b84d5dc8bde8438920c867c2dab4d64c50c001b5/frontend/src/components/narrative/PlayerInput.tsx#L22-L61) owns its draft in local component state; switching the conditional content unmounts it. It already prevents blank submissions, supports Enter/Shift+Enter and clears only accepted dispatches. Preserve those semantics while moving draft ownership/presentation under #62/#71.
5. [TabBar](https://github.com/subculture-collective/edda/blob/b84d5dc8bde8438920c867c2dab4d64c50c001b5/frontend/src/components/layout/TabBar.tsx#L18-L49) already scrolls horizontally and marks the selected tab. Arrow-key/roving-focus/tabpanel work belongs to #44, not this documentation PR. Existing tab targets measure about 31 px high and Send about 33 px; shrinking them further is not a solution.
6. [NarrativePanel](https://github.com/subculture-collective/edda/blob/b84d5dc8bde8438920c867c2dab4d64c50c001b5/frontend/src/components/narrative/NarrativePanel.tsx#L64-L139) already detects an 80 px near-bottom threshold and shows `New response ↓` when a new response arrives while the reader is away. Retain and test that behavior under #90. The investigation did not prove continuous auto-follow across every streaming/reconnect state. Its `scrollIntoView` can affect ancestors; qualification must assert document scroll stays stable too.

## Bounded requirements for existing issues

| Owner | Requirement and reproducible check |
| --- | --- |
| [#60](https://github.com/subculture-collective/edda/issues/60) frame | Implement the selected A layout, with all outer chrome including software credit inside an explicit available-height budget at the six default sizes. Declare scroll owners for transcript, detail pane and overflow content. Phone uses a detail view instead of stacking the full desktop rail before navigation. Keep readable text/touch targets. At increased zoom/text size, allow accessible reflow rather than clipping controls. Document the browser-chrome/visual-viewport strategy and verify it with a visible keyboard later. |
| [#62](https://github.com/subculture-collective/edda/issues/62) composer | Keep draft, selected action, submission and pending/error status reachable inside the current visual viewport. Bound long choices and multiline drafts with a discoverable internal scroll/expand path. Reproduce the eight-long-choice desktop clipping case and phone growth case. Keep legacy free-text/choice actions usable and preserve Enter/Shift+Enter, accepted-dispatch clearing, and duplicate-submission protection. Retain an unsent draft while opening details. |
| [#71](https://github.com/subculture-collective/edda/issues/71) context/navigation | Expose Play, Character, World and Journal; retain inventory under Character, and quests/NPCs/codex under World. Provide a reachable route for Logs, account, campaign list/replay, saves and export. Closing details restores the same draft, reading anchor and sensible focus without scrolling the document. Keep populated panels and modal overflow internally scrollable; do not rebuild the domain panels here. #44 remains the owner of tab semantics. |
| [#78](https://github.com/subculture-collective/edda/issues/78) qualification | Extend #45's harness after its merge; do not add a parallel harness here. At all six default sizes assert document width/height within the viewport (allow at most 1 px rounding), no document movement during history/detail scrolling, and visible/hittable enabled input/submission/navigation. Check long transcript, eight long choices, expanded details, errors/pending state, saves/export and combat. Assert intersection with the visual viewport and every clipping ancestor; a bounded document alone gives a false pass. Separately record touch navigation, focus/draft/reading-position restoration, visible software keyboard open/close and browser-chrome changes. Keep keyboard and human gates explicit. |
| [#90](https://github.com/subculture-collective/edda/issues/90) reading | Apply the selected serif narrative text within A's layout. Reduce repeated badges/borders/timestamps while keeping player/GM identity and mechanics distinguishable. Long history scrolls internally, stable entry IDs remain stable, reading older content survives new responses and detail navigation, and return-to-latest stays visible when appropriate. Reuse the existing near-bottom/unread logic where sound; this issue does not implement token streaming or chapters. |

No new tickets are needed. Combat, increased text size/browser zoom, populated detail panels, screen readers, real provider output and iOS Safari remain qualification cases, not completed tests in this spike.

## Reproduction and checks

Use a clean baseline checkout and two terminals. The fixture dependency is installed only into a disposable directory; the frontend lockfile is unchanged. Stop only these foreground processes when finished.

```sh
rtk proxy mkdir -p /tmp/edda-issue46
rtk proxy npm install --prefix /tmp/edda-issue46 --no-audit --no-fund ws@8.23.0
rtk proxy cp docs/playtest/issue-46/fixture-server.mjs /tmp/edda-issue46/server.mjs
rtk proxy node /tmp/edda-issue46/server.mjs
```

```sh
rtk proxy pnpm --dir frontend install --frozen-lockfile
rtk proxy env VITE_API_BASE=/api/v1 VITE_PROXY_TARGET=http://127.0.0.1:39246 pnpm --dir frontend dev --host 0.0.0.0 --port 39247 --strictPort
```

Open `/login`, sign in with `player@example.test` and `fixture-password`, then open Adventure alpha. The fixture accepts this dummy login. Submit `Look around`, click the returned choice, type a draft, switch every section and return. Reload between stress scenarios; the service does not persist played turns. Submit `long choices` and `fixture error` for the stress cases. Open Saves and user-menu Export. Resize the browser to the matrix above, reset document scroll to top for coordinates, and inspect element rectangles, clipping ancestors, scroll dimensions and focus. A shortened desktop viewport is not an actual software-keyboard check.

| Check | Result |
| --- | --- |
| `pnpm --dir frontend build` | PASS against unchanged baseline; existing >500 kB chunk warning |
| `pnpm --dir frontend lint` | FAIL: 23 existing errors; [captured diagnostics](lint-baseline.json). No frontend source or dependency edits in this PR. |
| `node --check docs/playtest/issue-46/fixture-server.mjs` | PASS |
| JSON parsing, relative artifact links, PNG signatures and measurement arithmetic | Checked before commit |
| Hosted PR checks | Read on the exact PR head at handoff; not implied by local checks |
| Human/device approval | PENDING; do not close #46 |

## Exact remaining human checkpoint

A human must review these baseline findings and perform or review a suitable visible phone session with an actual on-screen keyboard: focus input, type a multiline draft, observe browser chrome and viewport resizing, reach Send and choices while the keyboard is visible, open/close details and restore draft/reading position/focus, dismiss/reopen the keyboard, and inspect saves/export in portrait and landscape. Record device, OS/browser, orientation, text-size/zoom, observed failures and the reviewer/date on #46. This run's Android emulator touch result is supporting evidence; its hidden helper IME cannot satisfy the keyboard portion. No human completion was recorded during this investigation. Post-overhaul acceptance belongs to #78.
