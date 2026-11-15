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

Import `test`, `expect` and `signIn` from `browser/fixtures.ts`. The `game` fixture exposes sockets, submitted actions and API request records. Use `await page.setViewportSize({ width: 390, height: 844 })` before navigation, then `await signIn(page)` and `await page.goto('/play/alpha')`. Campaign beta has distinct history. Extend `fixtures/data.ts` for long transcripts or resource-rich scenes and map each extra endpoint explicitly. Production components and layout are unchanged by this harness. There are no viewport-fit assertions yet; #46/#78 own that assessment.

`Frontend behavior` runs unit/interaction checks, browser journeys and the build on GitHub branch pushes, and is required by `Mirror CI gate`. Traces remain in ignored `test-results/` on failures. Full lint currently reports 23 existing source errors on the starting revision b84d5dc (React hook/compiler and Fast Refresh rules); new harness files lint cleanly. That unrelated debt is not disabled or added to this CI gate.

## Campaign tabs

Tabs use automatic activation: Left/Right wrap through all eight sections, Home selects Narrative, and End selects Logs. Only the selected tab participates in sequential keyboard focus. Selection retains focus on the tab, including when a panel has an autofocus input. Each tab controls a stable panel ID and each panel is labelled by its tab; inactive panel shells stay hidden while their content remains unmounted. The selected panel is focusable.

The HUD's existing document order places panels before the bottom tab bar. Shift+Tab from Logs reaches its panel; Tab returns to the selected tab. This PR retains that order and visual layout. Interaction tests cover roving focus and bidirectional Tab traversal in the component; Chromium checks actual panel relationships, all sections, wraparound, Home/End, click activation and focus returning from the HUD panel. Screen-reader semantics are checked through roles/names/relationships; a real assistive-technology session remains unperformed.
