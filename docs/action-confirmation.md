# Prepared action confirmation (#64)

`POST /api/v1/campaigns/{id}/action/confirm` accepts a required UUID `action_id`,
`expected_revision`, the preview's typed `intent`, and an optional `target_id`.
The route uses campaign ownership middleware. Unknown fields, client modifiers,
difficulty, advantage and resource quantities are rejected. The shared engine
method is available to TUI adapters without an HTTP dependency.

Confirmation holds the campaign turn lock and a short database transaction. It
reloads and locks the same campaign/actor snapshot as preview, checks availability
and revision, validates the target, and resolves the ordinary check with the
server SRD resolver. Saves must explicitly target the actor. Other supported
checks have no scene target. Combat, attacks, spells and consumables are excluded;
there are no direct resource costs for these checks. A DC or scene outcome has
not been adjudicated, so the receipt explicitly withholds success/failure. It
does not invent a difficulty, advance time or claim a scene effect.

The exact dice receipt, completed action result and history entry commit together.
An action ID bound to different input or campaign is rejected. Identical retries
return the saved result before checking the now-obsolete revision. A fresh action
needs a fresh preview after relevant actor/campaign changes or a completed check.
Pending narrative recovery blocks a new confirmation. No model inference or
second Roll prompt participates in confirmation.

The sheet retains uncertain request identities in session storage before sending;
retry sends the same ID and input. Remount/reload reads `GET /action?action_id=…`
to reconcile a committed result. Storage denial retains in-memory retries during
the panel lifetime. The sheet shows the saved inspectable receipt and publishes
the result to the active transcript; turn numbers deduplicate live/history/HTTP
reconciliation. Character caches refresh after confirmation. Other connected
clients obtain the saved receipt and source history on reload; this endpoint does
not broadcast to other clients. A WebSocket `confirm_action` message accepts the
same request and action ID; reconnect retries return the identical saved result.
The socket allows SRD confirmations while free-form SRD turns remain blocked by
the shared engine rules guard.

Database fixtures use injectable dice and isolated PostgreSQL to check stale HP,
actor/target/entry validation, changed-input conflicts, exact receipt replay after
engine recreation, history invalidation, a single roll and a single history row.
Frontend fixtures cover explicit confirmation, uncertain-delivery retry and reload
reconciliation. No live inference, production campaign or deployment is involved.
