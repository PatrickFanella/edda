# Live-event retention decision (#43)

Investigated on 2026-11-15, frontend baseline `28edd19` (includes #42's campaign
isolation). Decision: retention grows without a bound, but there is no measured
user-visible latency failure in this fixture. Implement consumption acknowledgments
in [#109](https://github.com/subculture-collective/edda/issues/109), a separate PR.

## Repeatable baseline

`frontend/tests/eventRetention.test.tsx` drives the real socket hook with ten
status frames and one 560-character narrative result per turn, committing each
frame in a separate React `act`. No API, server, database or model is used.
Run `pnpm --dir frontend exec vitest run tests/eventRetention.test.tsx
--reporter=verbose --disableConsoleIntercept` to print `RETENTION_BASELINE`.
Only timeout/Date APIs use the fake clock; durations use real `performance.now`.

| Completed turns | Retained frames | Serialized event bytes | Prior references copied by append | Local interval |
| --- | --- | --- | --- | --- |
| 100 | 1,100 | 400,785 | 604,450 | 50 ms for turns 1–100 |
| 1,000 | 11,000 | 4,009,787 | 60,494,500 | 383 ms for turns 101–1,000 |

Bytes are the UTF-8 JSON size of the retained parsed events, including their
envelope and decoded payload. They are a reproducible payload measure, **not a
heap measurement**. Object overhead, shared strings, GC and React scheduling
make actual memory different. Reference counts follow `N*(N-1)/2` for the
existing `[...current, event]` append. Durations are observations on this host,
include dispatch/render/test overhead, are not comparative benchmarks and have
no pass/fail threshold. Neither these timings nor the browser run establish
long-session phone usability; production heap/long-task profiling remains unqualified.

## What must remain and what can be released

The narrative reducer already converts results into durable transcript entries
and consumes transport with `processedEventCount`. Status uses only the latest
display state; consumed status frames, chunks and the duplicate result payloads
are not needed for rendering older entries. Same-campaign reconnect currently
keeps the lifetime list/cursor, while #42 now clears it on a campaign switch.
Reconnect does not provide server event IDs or replay guarantees.

Do not impose an arbitrary ring limit: slicing the existing array changes the
meaning of the reducer's index and can skip or replay events, especially if
React receives more than the limit in a batch. #109 requires acknowledging
consumed sequence ranges, releasing processed payload, an explicit suspended
consumer policy, and tests for bursts, errors/chunks/results, campaign switches,
history-versus-live ordering and reconnect. Keep durable narrative entries;
transcript virtualization belongs to #90, incremental model streaming to #41.
No retention implementation, protocol change or history deletion is hidden in
this spike. Its baseline assertion intentionally describes the current behavior
and must become a bounded assertion in #109.

## Counterevidence and checks

The Playwright retention fixture sends 1,100 frames / 100 results into a page
with restored history, reconnects, checks every result remains exactly once,
then completes another action. This confirms current history/reconnect rendering
for the fixture; it does not prove a future queue implementation safe.

- `pnpm --dir frontend test`: 10 tests passed, including the 1,000-turn probe.
- `EDDA_FIXTURE_PORT=18443 pnpm --dir frontend test:browser`: 6 journeys passed,
  including the 100-turn reconnect test. External service requests are blocked.
- `pnpm --dir frontend build`: passed with the existing bundle-size warning.
- Focused ESLint on the two new tests: passed. Full lint's 23 existing errors
  were measured on the same frontend for #42 and are unchanged by these tests.

No production profiling, deployment, browser/device human checkpoint or hosted
CI claim is made. The parent monitors the required integration gate.
