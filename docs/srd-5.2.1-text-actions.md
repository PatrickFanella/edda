# One supported SRD text request (#342)

In an SRD-versioned campaign, enter **Make an Athletics check** in Play.
The request opens a prepared Athletics preview in Character. Inspect its
modifier and source, then explicitly confirm. Preparing does not roll dice,
spend resources, advance campaign time, add history or request model inference.

Only that exact literal is supported in this slice. Other wording, capitalization
or whitespace receives a clarification and remains editable in the composer.
Choose other supported actions directly from Character. This is a bounded
entry point, rather than general free-text SRD adjudication. Legacy campaign
text continues through its existing narrative WebSocket path.

## Shared authoritative contract

The client sends a read-only POST to the campaign character action-preview
resource with exactly two JSON fields: `actor_id` and `input`. The actor must
belong to the authenticated user's campaign; the input must be the literal
above. The body keeps entered prose out of URL query strings. GET previews for
existing typed intents retain their current contract.

The server maps the supported literal to the same canonical ability-check,
Athletics intent used by the sheet and calls the existing owned-snapshot and
stored-action preparation path. It returns the existing `ActionPreview`, with
its expected revision, modifier, source and availability. Unknown or mixed
parameters cannot specify a DC, outcome, resource or effect. Responses are not
cacheable. Unsupported builds remain unavailable.

Confirmation uses the existing durable action endpoint and identity. A stale
preview cannot roll or spend; duplicate confirmation or reconnect replays the
same committed receipt. An earlier unresolved prepared request must remain
available for its own retry/receipt recovery; a new text preview cannot replace
it or cause it to confirm under a misleading new selection.

## Qualification boundary

Backend integration fixtures compare text and typed previews and deterministic
receipts against disposable PostgreSQL, including refusal, ownership and retry
negative cases. Frontend fixtures verify preparation, draft retention and
explicit confirmation at desktop and phone sizes. Those browser responses are
client contract evidence; they are not substitutes for real database proof.

No legacy gameplay guard or WebSocket protocol is loosened. No broader build,
action, spell, target, skill parser, model account or deployment is added.
#89 still requires its separate phone-keyboard and final human experience
checkpoint, and Full Rules remains gated by #54 and the complete register.
