# Campaign-switch isolation decision (#42)

Investigated on 2026-11-15 against `faa9c110ff159b081c56967cb0de4fdba4182bf3`.
Decision: confirmed contamination; apply a small containment fix in the hooks.

## Reproduction and counterevidence

`frontend/tests/campaignSwitch.test.tsx` mounts the real narrative and socket
hooks, with a controllable socket and deferred history. Campaign A receives a
result/choice, starts combat and sends an unfinished action. The same hook then
switches directly to B. A's history resolves or rejects after B's history loads;
a frame from A is also delivered after its socket closes.

Before the fix both cases failed: B retained A's `Inspect the key` choice.
Source inspection explains the related failures: the reducer reset its cursor
while the socket retained events/status/combat/loading state, and both history
promise callbacks could still dispatch after a campaign change.

The existing Playwright journey switches through the campaign list (unmount and
remount) and passes. That path does not disprove a failure in the mounted hook.
The new tests cover both direct switching and a subsequent unmount/remount.

## Containment and limits

Reset narrative and transient socket state before committing a render for a new
campaign. Retain events across reconnection to the **same** campaign. Invalidate
both success and failure history callbacks when their effect is cleaned up.
The old socket already removes listeners and checks socket identity; that is
counterevidence for late frames from a closed connection reaching B.

Both delayed outcomes now leave only B's history, no old choices/result/status,
no active combat and no pending response after B connects. No transport protocol,
backend, data conversion or deployment changed. Requests are ignored after
cleanup rather than aborted; this prevents stale UI writes but does not claim
server-side work cancellation. History/live ordering within the same campaign
and replay/deduplication on reconnect need a separate protocol decision when
retention is implemented under #43. This fix does not change those contracts.

## Repeatable checks

- `pnpm --dir frontend test`: 9 tests passed, including both new switch cases.
- `EDDA_FIXTURE_PORT=18442 pnpm --dir frontend test:browser`: 5 isolated journeys
  passed, including reconnect, campaign-list switching and replay. The harness
  now accepts a per-run port and still blocks external/API traffic by default.
- `pnpm --dir frontend build`: passed; existing large-bundle warning remains.
- `pnpm --dir frontend lint`: 23 errors, matching the #46 baseline (including
  the pre-existing synchronous effect setter in `useWebSocket`). No new error
  was reported in the narrative hook or test.

These are deterministic local checks. They do not qualify real networking,
production data, phone keyboard behavior or hosted CI.
