# Durable turn completion and narration recovery (#30, #59)

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
the action ID. A retry reads the receipts and either reuses a validated narration
candidate or requests narration with no tools advertised or dispatched. Recovery
rejects returned tool calls, empty narration and known unbacked durable claims;
the existing claim audit is a heuristic, not a semantic fidelity guarantee. It
never runs post-turn mechanical extraction or rerolls dice. Completion writes one
final narration/history row and returns recovered=true. A different action is
refused until the pending work is recovered.

Migration 00040 adds candidate_result, narration_ready and narration_attempts. The first distinguishes
validated candidate text from the raw initial model response, which may precede
partial or failed tool processing. The second limits recovery narration inference
to three calls across requests and process restarts. An attempt is charged before
inference, including when the process dies during it. A validated candidate can
still be persisted after the limit without another model call. The prepared completion candidate stores generated choices and derived state before
final history persistence. Recovery reuses it without inference. If interrupted
before that preparation and the narration lacks choices, the existing choice
provider supplies them under the same three-call recovery budget. At exhaustion the attempt stays pending, with its receipt available and
an explicit attempt-limit error. No mechanical fallback is silently published.
Operator intervention is required after persistent failure; the public API does
not reset the counter or abandon committed effects.

Recovery is demand-driven through the existing HTTP/WebSocket turn lifecycle,
including after process restart. There is no background queue or startup inference.
The shared engine exposes GetActionReceipt; the campaign-owned API exposes
GET /api/v1/campaigns/{id}/action to discover pending work and the same route with
an action_id query parameter to read one recorded action. A missing receipt returns
JSON null. The response includes the original input, action ID, status, turn number,
attempt count/limit, and the existing player-facing resolution/state-change
projections. Unvalidated candidate prose and raw tool arguments are withheld.
Completed receipts also include the final TurnResponse. Reading never runs inference
or mechanics. To recover, POST the receipt's original input and action_id to the
existing /action endpoint (or send the same WebSocket action payload). Existing
clients can retain and retry an ID; automatic reconnect discovery UI is outside
this backend delivery.

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

Narration recovery verification seeds exact dice/effects to simulate a crash,
reconstructs the engine, retries through HTTP/shared engine/streamed result paths,
and checks unchanged receipts, HP, one history row and replay without inference.
It exercises provider failure, hallucinated recovery tools, durable attempt bounds,
pending discovery, completed reads and cross-owner denial. The completion-trigger
fixture covers failure after generation and before delivery, proving candidate
reuse without a further provider call. All inference uses deterministic doubles.
