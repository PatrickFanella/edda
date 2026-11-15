# SRD 5.2.1 coverage status

Edda currently supports a partial, tested SRD 5.2.1 gameplay slice. It does not
claim Full Rules support. The character sheet gets the boundary from the server.
Campaign setup reads a checked-in JSON snapshot of that same report, and a Go
test rejects a change when the snapshot and server report differ. The declared
slice includes the two supported level-one Human builds, the narrow Fighter
level-two transition with Action Surge and Tactical Mind still read-only,
inspectable checks and saves, one-handed longsword attacks, named self-only
spells/items, one explicit opportunity-attack trigger, and the recorded rest
and concentration fixtures.

`internal/rules/srd521/coverage.json` is the register of source scopes. The
machine gate refuses a full-support claim unless every stable scope remains in
the register, each scope is implemented, and every fixture path validates. The
gate also remains closed while source inventories are planned but unqualified:
344 character and advancement records from #74, 335 world records from #82,
and 753 spell/item source families from #80. Planning data, parsed catalogue
records and a narrow fixture do not count as playable support.

The full target remains owned by epic #54. Its bounded follow-up issues and
qualification gates must close with evidence before a future release can change
this status.

The current inventory totals are deliberately fail-closed constants, checked
against their source manifests. A future full-coverage release must replace that
pending-inventory representation with source-complete qualification evidence;
changing register statuses or closing tracker issues alone cannot open the gate.

## Supported text entry point

The exact composer request **Make an Athletics check** prepares the same
Athletics intent as the character sheet for the supported build. It shows a
read-only preview first; the existing explicit Confirm action remains the only
mutation. Other text requests receive a clarification and stay in the draft.
Legacy campaigns keep their existing narrative WebSocket behavior. This
bounded entry point does not enable general SRD narrative adjudication, remove
the legacy mutation guard, or promote a whole rules family.
