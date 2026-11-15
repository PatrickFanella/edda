# On-demand campaign chapters

Journal → Chronicle → **Generate next chapter** freezes the next completed player-visible turn range and asks the configured `summary` model route for a concise chapter. It records the source before inference. No paid provider, production account or application database is needed for the regression tests.

A chapter has an immutable ID, exact source log IDs and completed action receipt IDs (when available), covered turn numbers, a SHA-256 source identity, rules version/mode at capture, and the knowledge boundary `player-visible-transcript-through-to-turn-v1`. The source snapshot contains only player input and final visible narration. It excludes raw tool calls/results, unrevealed world facts, current world state and pending narration. Modern engine logs already strip choice scaffolding; older logs may retain offered choices, which the generation instructions forbid treating as played actions. Legacy completed logs without action receipts remain eligible. Legacy summaries neither define chapter coverage nor become model input.

The next range starts after the last captured range belonging to the current transcript. It stops at a gap, an empty narration or a pending/uncompleted turn, and covers at most 100 consecutive turns per request. A failed first generation reserves its source for retry rather than skipping it. Completing more turns while generation runs cannot change that source. Start Over keeps earlier chapter snapshots; replacement logs have new IDs and begin their own coverage, even when turn numbers repeat. This does not change Start Over's existing transcript deletion behavior.

Each generation revision records a request ID, revision number, prompt version, status, prose or a sanitized failure message, and creation time. Explicit regeneration creates another revision over the same snapshot, retaining earlier prose. Replaying a completed request returns the saved revision. Replaying a failed request retries that revision. A retry in another process uses the same durable record. Journal retains uncertain request keys in tab-local session storage across transport failures, panel navigation and reloads (with an in-memory fallback when storage is blocked); persisted retry buttons also work after reload.

The provider receives only the frozen source and fidelity instructions: preserve played outcomes, costs and uncertainty; treat intentions as intentions; invent no events, dialogue or resolutions; never reroll or reinterpret mechanics. These are generation constraints, not a semantic proof that every possible model output is faithful. Source snapshots remain available through the API for comparison. Blank responses and tool-call responses fail without saving chapter prose.

Generation runs outside database transactions, with a 90-second request timeout and a two-minute persisted lease. Concurrent requests using the same key return the in-progress revision; competing keys for the same source receive 409. Expired leases can be reclaimed by retry. A lease token prevents a late worker from overwriting a recovered revision. Completion and next-range reservation use the same short campaign row lock, so completion cannot turn a next-range request into an unintended regeneration. A worker killed without cleanup may show `generating` until retried; Journal polls the durable status and exposes retry. Provider failures do not mutate gameplay, the transcript, saves or earlier revisions. Provider error text is not stored or returned.

## API

Both routes use the existing campaign authentication and ownership middleware. Chapter generation supports legacy and SRD campaigns because it does not invoke mechanics or mutate game state. Other rules compatibility guards remain in force.

- `GET /api/v1/campaigns/{id}/journal/chapters` returns an ordered array of chapters, their source snapshots and all revisions.
- `POST /api/v1/campaigns/{id}/journal/chapters` accepts `request_id` (a nonzero UUID) and optional `chapter_id`. Omit `chapter_id` for the next source range; supply an existing campaign-owned chapter ID for regeneration. Reuse `request_id` for retries. A key cannot be rebound to another chapter.
- Responses: 200 for a saved completed/failed revision or a message when no completed range exists, 202 for an in-progress retry, 400 for invalid input, 404 for an unavailable chapter/campaign, 409 for a competing generation or rebound key, 503 when no chapter provider is configured.

Migration 00041 creates two tables and does not rewrite existing logs, summaries or notes. Legacy summary reads and generation APIs remain compatible; Journal's new generation control uses chapters. Automatic legacy summarization scheduling is unchanged. The migration down operation removes only the new chapter tables, so back up their contents before rolling it back after use.

## Verification

`internal/handlers/chapter_generation_integration_test.go` exercises the production router and disposable Testcontainers PostgreSQL with deterministic model doubles. It covers committed range selection, legacy sources/summary compatibility, knowledge isolation, immutable regeneration, process restart, idempotency, concurrent keys, failures, lease expiry, stale-worker fencing, ownership and reset coverage. `internal/server/routing_integration_test.go` verifies the configured summary route for chapters. No dice or mechanics run during chapter generation.

Run:

```sh
go test -short -race ./...
go vet ./...
go test -tags integration -run Integration -count=1 -timeout 10m ./internal/handlers ./internal/state/sqlc ./internal/server
pnpm --dir frontend test
EDDA_FIXTURE_PORT=4187 pnpm --dir frontend test:browser --workers=2
pnpm --dir frontend build
pnpm --dir frontend lint
```

The isolated frontend tests cover failure, same-key retry, regeneration, reload and continued play through fixture sockets. The existing frontend lint baseline has 22 errors outside the chapter files; do not suppress those errors to qualify this change.

## Chapter reader (#76)

Journal displays one completed saved revision at a time in a bounded prose pane.
Chapters sort by source capture time, then turn range/identity, so retained
chapters precede newer chapters after turn numbers restart. Previous/next and
chapter selectors provide the same source-ordered reader on desktop and phone;
a revision selector exposes earlier completed prose without showing failed or
in-progress content as a finished chapter. Generation/retry controls and revision
status remain available below the reader.

Chapter/revision selection and relative reading offsets are retained per campaign
and revision in local storage. Navigating between chapters, leaving Journal and
reloading restores the saved position. Relative offsets adapt to layout changes;
there is no cross-device preference synchronization. If storage is denied, an
in-memory fallback retains position while the app remains loaded. Preferences
contain identities and offsets, never copied story prose.

Coverage displays the selected source range and player-knowledge boundary. A
fresh history read reports completed turns that have no completed saved chapter.
Turn-number coverage also requires the history entry to predate the chapter's
capture, so restarted play is not claimed as covered merely because its turn
numbers match an old chapter. Freshness failure is explicit and does not block
reading. Completing a turn invalidates history/Journal queries. Pending mechanical
work is not completed history and is not counted as chapter-ready play.

“Read source history” opens and focuses the chapter's immutable captured source
inside the reader. Source turns retain original player input, visible narration,
log IDs and action IDs. This is the exact provenance used for generation, including
when current campaign history has since been restored or restarted. Reading never
changes history or generates another outcome. The existing outer gameplay frame
retains its internal scroll ownership; the prose pane also owns its own scrolling.

Deterministic frontend tests cover ordering, navigation, revision filtering,
freshness failure, restarted-turn coverage and persisted preferences. Desktop and
phone browser fixtures exercise actual scroll restoration across chapter changes,
Journal unmounts and reloads, source inspection and document viewport bounds.

## Saved chapter export (#86)

Open **Export → Saved Chapters**, check the chapters to include, choose each saved
revision, then download Markdown or JSON. Only completed revisions are offered.
The revision currently selected in the reader is the default for that chapter;
other chapters default to their latest completed revision. Selection is explicit:
newer generated revisions cannot silently replace the requested text during a
download. Nothing is regenerated or rerolled.

Both formats retain the selected prose, chapter and revision identities, request
ID, capture/save timestamps, turn coverage, source log/action IDs, immutable
player-visible source history, source SHA-256, knowledge boundary, rules
version/mode and prompt version. They include the canonical SRD 5.2.1 attribution
and an Edda adaptation/generated-prose notice. Rules metadata comes from the
captured chapter, even if the campaign's current mode has changed. Captured
sources may predate restored or restarted play, exactly as in the reader.

`GET /api/v1/campaigns/{id}/export/chapters` uses the existing campaign
authentication and ownership middleware. Supply one or more repeated
`revision_id` UUID parameters (1–100 unique values), and optionally
`format=markdown` (default) or `format=json`. The UI selects one revision per
chapter; API callers may explicitly select multiple completed revisions of the
same chapter. Chapters use the reader's source-capture order, including ranges
retained from earlier play. Revisions keep their saved version order.

A missing, foreign-campaign, failed or generating revision returns 404 for the
entire selection; invalid selections/formats return 400. Anonymous requests
return 401 when authentication is enabled; another owner's campaign is 404.
Downloads are private/no-store. Empty campaigns show guidance to create a chapter
in Journal. Loading and download failures leave an explicit retry path.

This chapter-only export reads the same saved projection as the reader. It does
not add campaign descriptions, current world state, hidden lore, raw tool data,
pending narration, unselected prose, failed-attempt details, notes or legacy
summaries. Existing full-data, transcript and character exports keep their
existing contracts; this privacy boundary describes the Saved Chapters option.
Markdown retains raw saved text; JSON retains its exact string values for
machine comparison. Neither format certifies model prose semantically faithful
beyond the existing chapter-generation constraints.

Deterministic checks live in `internal/export/chapters_test.go`,
`internal/handlers/chapter_export_integration_test.go`,
`frontend/tests/chapterExport.test.tsx` and
`frontend/tests/browser/chapter-export.spec.ts`. The backend integration test uses
the production router, disposable Testcontainers PostgreSQL and a recording model
double; it compares exports with the reader's saved snapshot after regeneration
and transcript changes, tests JWT ownership/selection failures and poison-marker
exclusions, and checks that export makes no generation calls or revision writes.
Frontend fixtures cover exact revision selection, loading/download recovery,
empty campaigns and desktop/phone-sized downloads without external services.
