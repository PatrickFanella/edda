# Frontend behavior baseline

Run from the repository root with Node 24 and the package-pinned pnpm:

```sh
cd frontend
corepack enable
pnpm install --frozen-lockfile
pnpm exec playwright install --with-deps chromium
pnpm test
pnpm test:browser
pnpm build
pnpm lint
cd ..
```

`test` type-checks the harness and runs Vitest interaction, narrative and socket tests in jsdom. `test:browser` starts and stops its own local Vite process on 127.0.0.1:4175 and drives Chromium. That port must be free. Each browser test has fresh storage, query state, HTTP fixtures and sockets. No backend, database, credentials or provider keys are used. API requests are fulfilled locally; unknown resources fail the test. External requests are blocked and sockets never connect to a server. Narrative mode fixtures have no dice or model calls.

The browser suite covers campaign and character interviews through world creation and play, a completed live turn, interrupted socket/reconnect, switching via the campaign list, replay next/previous controls, registration, login, restored sessions, sign-out and expired tokens. Socket interaction tests also check campaign-change and unmount cleanup. These checks establish current behavior, not backend persistence or live provider qualification. Same-mounted campaign-switch races belong to #42; this baseline exercises the existing list-to-play navigation.

## Reusing fixtures for viewport work

Import `test`, `expect` and `signIn` from `browser/fixtures.ts`. The `game` fixture exposes sockets, submitted actions and API request records. Use `await page.setViewportSize({ width: 390, height: 844 })` before navigation, then `await signIn(page)` and `await page.goto('/play/alpha')`. Campaign beta has distinct history. Extend `fixtures/data.ts` for long transcripts or resource-rich scenes and map each extra endpoint explicitly. The viewport checks in `browser/viewport.spec.ts` and `browser/qualification.spec.ts` cover the screen-fitting layout; the baseline investigation remains in `docs/playtest/issue-46/README.md`.

`Frontend behavior` runs unit/interaction checks, browser journeys and the build on GitHub branch pushes, and is required by `Mirror CI gate`. Traces remain in ignored `test-results/` on failures. Full lint reports 22 existing source errors on the #78 starting revision cf0a61f (React hook/compiler and Fast Refresh rules); new harness files lint cleanly. That unrelated debt is not disabled or added to this CI gate.

## Campaign tabs

Tabs use automatic activation: Left/Right wrap through Play, Character, World and Journal; Home selects Play and End selects Journal. Only the selected tab participates in sequential keyboard focus. Selection retains focus on the tab, including when a panel has an autofocus input. Each tab controls a stable panel ID and each panel is labelled by its tab; inactive panel shells stay hidden while the story remains mounted to retain its draft and reading position. The selected panel is focusable.

The HUD's existing document order places panels before the bottom tab bar. Shift+Tab from the selected tab reaches its panel; Tab returns to the selected tab. This PR retains that order and visual layout. Interaction tests cover roving focus and bidirectional Tab traversal in the component; Chromium checks actual panel relationships, all sections, wraparound, Home/End, click activation and focus returning from the HUD panel. Screen-reader semantics are checked through roles/names/relationships; a real assistive-technology session remains unperformed.

## Viewport qualification (#78)

`browser/qualification.spec.ts` uses the existing isolated `game` fixture at 1366×768, 1440×900, 1920×1080, 768×1024, 360×800 and 390×844. It covers long history, eight long choices, multiline drafts, pending/validation recovery, Character/World/Journal context, saves/export and the current fixture combat UI. The probe in `browser/layout.ts` combines document dimensions, visual-viewport and clipping-ancestor intersections, center hit testing, 44px targets and wheel-scroll ownership. Its negative control in `layout-guards.spec.ts` rejects clipped and covered controls even when the document fits.

Fixed action controls are measured before scrolling in normal play. Long choices, narrow horizontal navigation, details, saves, combat and enlarged-text layouts may scroll within their declared owners; tests expose each control before measuring it and require the document to stay at (0, 0). A tall choice needs a visible, hittable 44px portion; fixed controls must fit completely, allowing 1px for rounding. The 200% root-text cases include two phone sizes and a constrained 683×384 workspace; these are text/reflow checks, not native browser zoom or keyboard evidence.

Each qualification case attaches JSON geometry and writes screenshots under ignored `test-results/`. To save a machine-readable run, set `PLAYWRIGHT_JSON_OUTPUT_NAME` and use `--reporter=line,json`. `EDDA_FIXTURE_PORT` selects a free isolated server port when another agent is using 4175. Playwright owns and stops that server. Run only these cases with `pnpm --dir frontend test:browser qualification layout-guards --workers=2`.

Touch emulation (`hasTouch` and `isMobile`) is a separate test group. Synthetic visual-viewport resizing stays in `viewport.spec.ts`. Neither opens a phone software keyboard. See `docs/playtest/issue-78/README.md` for retained evidence and the still-required visible device/human checkpoint.
