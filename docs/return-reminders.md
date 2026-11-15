# Returning to play (#85)

The gameplay header offers **Where was I?** to open a short reminder.
**Dismiss reminder** or Escape from the reminder closes it. Its content scrolls inside a bounded popover;
the action composer and transcript stay mounted. Opening or dismissing does not
change a draft, selection or reading position, submit an action, or advance time.
It is available in the narrative and existing combat views.

## Sources and API

`GET /api/v1/campaigns/{id}/journal/reminder` is behind the existing campaign
ownership guard. It returns optional `scene` and `objective` items plus a
`known_facts` array. Each item has `text` and `source_id`.

A single database statement reads one consistent committed snapshot:

- The latest character's current location label, only if the location belongs to
  the campaign and is explicitly player-known or visited.
- The most recently updated active quest's title, already exposed in the quest
  log. It does not read quest descriptions or future objective text.
- At most three current player-known world facts, preferring clues and then
  recent facts. Hidden and superseded facts are excluded.

Every text field is trimmed and capped at 200 Unicode code points plus an
ellipsis. Source identifiers use `known-location:`, `quest:` or `known-fact:`.
Missing scene/objective/facts are shown as missing; the reminder never invents a
forgotten event or suggests that a quest has succeeded.

This is a deterministic projection, with no provider, dice, writes, migrations
or extra engine interface. Automatic summaries remain available in Journal, but
the reminder does not reuse generated prose that may be stale or make claims
beyond current player knowledge. It is a current-situation aid, not a chapter
or a chronological session recap.

## Availability and retention

The reminder request is independent of the action path. Pending and failed reads
explicitly leave play available, with a retry control for failure. Campaign-scoped
React Query keys prevent cross-campaign responses; existing journal invalidation
refreshes the reminder after a completed turn and the query refreshes on window
focus/reopening. Closing keeps the composer and transcript intact.

The component uses the existing charcoal/champagne/pewter/gold palette, current
body font, square borders and focus ring. No new fonts, shadows or motion are
introduced. Its content has an internal height cap of 24dvh or 12rem, with a
separate reachable toggle. The popover adds no height to the story/composer stack.

## Verification

`internal/handlers/return_reminder_integration_test.go` uses isolated PostgreSQL
and the real ownership-guarded router with no provider or engine. It covers
empty/legacy context, other-owner denial, unknown/visited locations, hidden and
superseded facts, inactive quests, generated-summary exclusion, bounded Unicode
text, refreshed knowledge and absence of history/action/save writes.

`ReturnReminder.test.tsx` covers contents, dismiss/reopen, delayed reads, retry,
empty state and safe errors. `browser/reminder.spec.ts` uses isolated HTTP/socket
fixtures at 360, 390 and 1366px to check retained drafts/selections/scroll,
reachable Send, document fit, playable delayed/failing requests, turn refresh and
campaign isolation. The full browser regression suite also covers combat and
six default viewport sizes. These are Chromium fixture checks, not live account,
real phone keyboard, screen-reader or human qualification evidence.
