# Wave 14 browser qualification (#89)

Observed against the Wave 13 merged coverage gate. This is deterministic
Chromium evidence using the isolated fixture API and socket; it has no
database, provider, deployed service, or physical-phone access.

## Combined journey

`campaign-qualification.spec.ts` runs at desktop 1440x900 and phone-sized
390x844 workspaces. It proves the visible partial-SRD statement, selects and
previews self-only Cure Wounds, refuses confirmation until the self target is
selected, and spends one slot on confirmation. Its SRD text socket then returns
the current public guard error; after reconnect, an explicit retry resends the
same action id and leaves the resource confirmation unchanged. That establishes
client recovery behavior only, not SRD text-turn support. The test then opens
the retained legacy campaign, reconnects its narrative socket, and submits one
ordinary action. Reconnecting never repeats the sheet confirmation. The
separate viewport qualification suite owns document-fit and scroll measurements.

It complements, rather than replaces, the focused browser suites:

- `martial-creation.spec.ts` covers guided Fighter and Cleric creation plus
  failed-generation retry.
- `action-preview.spec.ts`, `caster-lifecycle.spec.ts`, and `reactions.spec.ts`
  cover martial actions, potion cost, caster concentration/rest, and explicit
  reactions on desktop and phone-sized workspaces. The caster lifecycle fixture
  includes reload after its confirmed effect.
- `chapters.spec.ts`, `chapter-reader.spec.ts`, and `chapter-export.spec.ts`
  cover chapter generation/retry/reader position, fidelity fixtures, and saved
  revision export. `journeys.spec.ts` covers legacy campaign switching/replay.
- `qualification.spec.ts` and `layout-guards.spec.ts` cover six viewport sizes,
  internal scroll ownership, 44px reachability, touch emulation, and text reflow.

## Result boundary

Passing Chromium fixtures establish front-end contract behavior with explicit,
local responses. They do not establish PostgreSQL durability, a live model,
browser chrome, a visible software keyboard, native zoom, assistive technology,
or human player judgement. The SRD free-text guard remains a gameplay blocker;
the fixture retry does not resolve it. The available Android session and the
human review remain separate evidence; #78 and #46 retain that checkpoint.

## Final validation

- Full frontend unit/interaction suite: 64/64 PASS.
- Final full Chromium suite: 93/93 PASS, no skips, retries or flaky cases,
  using one worker and the normal assertions/timeouts.
- Production build: PASS with the existing chunk-size warning.
- Focused hook/reducer retry cases: 10/10 PASS; changed-file ESLint: PASS.

The final passing browser command was:

```sh
rtk proxy env EDDA_FIXTURE_PORT=39605 PLAYWRIGHT_JSON_OUTPUT_NAME=/tmp/edda-wave14-browser-final.json mise exec node@26.13.0 -- pnpm --dir frontend test:browser --workers=1 --reporter=line,json
```

An earlier run with default parallelism ended 65 passed/28 timed out on the
busy shared host. Observed load average was 91, with swap full. Reducing this
run to one worker, without changing assertions or timeouts, produced the
passing result above in 3.5 minutes. No unrelated process was stopped. This
is functional Chromium qualification, not a latency or capacity benchmark.

One initial unit run had a transient PlayerInput timing failure; the focused
rerun passed and the final complete unit suite passed 64/64. No expectation
was weakened to obtain those results.

Whole-tree ESLint was rerun after browser artifacts settled: 22 errors remain
in baseline source; changed-file lint passes. The temporary two retry-change
ref diagnostics were fixed, without suppressing rules or changing the baseline.
