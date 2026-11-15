# Durable turn completion (#30)

A mechanical tool commits its database effects and exact applied-call receipt in
the same short transaction. A handler failure rolls that transaction back. Later
provider/choice/history failures can leave partial mechanical progress, but it is
tracked in turn_attempts and cannot be silently rerun. This is an explicit partial
completion contract, not a transaction held across a whole model-driven turn.

Migration 00039 adds campaign-scoped attempts with an action UUID, allocated turn
number, original input, pending_mechanics/pending_narration/completed/failed state,
applied calls, candidate narration, final narration and a recovered flag. A unique
pending-attempt index and the campaign turn lock prevent overlapping allocation.
Pool and direct-connection engines use a dedicated session advisory lock, not an open
transaction or query-pool slot across inference. Custom database adapters must
expose a pgx connection configuration; opaque transaction-only adapters fail at
construction rather than retaining an inference-spanning transaction.

The HTTP and WebSocket action payloads accept action_id. A completed ID replays
its stored outcome; the same ID with changed input or another campaign is rejected.
The browser reuses its ID after interruption/recovery and creates a fresh ID after
a result, so deliberately repeating the same text remains possible. Callers that
omit an ID remain compatible; they can recover the sole pending attempt by its
original input, but full completed-request idempotence requires an explicit ID.
The shared engine/TUI signature is unchanged; clients can use engine.WithActionID.

Pending attempts never enter normal session history, replay, context, automatic
summaries or auto-save. Normal completion atomically inserts the history row and
marks the attempt completed. A failed completion returns a recovery conflict with
the action ID. A retry with that ID reads the committed receipts, writes an
explicit mechanical-only fallback, and returns recovered=true without invoking
providers, tools or dice. A different action is refused until that pending work
is recovered. Original candidate text and receipts remain available for #59's
narration-retry work; this slice does not regenerate narration.

A commit error is treated as uncertain, stops all subsequent tool processing, and
requires a durable receipt read on retry. It never falls through to a model-led
tool retry. Failures before any receipt/candidate can safely retry mechanics.
Tool-created semantic memories retain their existing content/metadata/route;
their embedding callbacks run after the mechanical commit and remain best effort.
They are not a durable outbox; a process crash may omit an auxiliary embedding.
Start Over clears attempts atomically with history/saves/time. Campaign deletion
cascades attempt cleanup. No existing campaign data is converted.

Verification injects failures in receipt writes and final history writes, checks
HP rollback versus committed receipt recovery, same-ID replay, fresh-ID repeated
input, cross-campaign rejection, zero-tool and multiple-tool attempts, no provider
calls inside open transactions, and uncertain-commit early termination. Existing
HTTP/WebSocket, turn-order, ownership and reset integration suites remain enabled.
