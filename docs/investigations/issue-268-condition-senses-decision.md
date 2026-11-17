# Blinded, Deafened and Invisible execution ownership (#268)

The three condition scopes already have bounded implementation Tasks: **Blinded #960, Deafened #961 and Invisible #962**. This decision reuses those published contracts and creates no duplicate implementation issue. All three are native children of #268 and remain required directly by both #279 and #280. Closing this planning decision does not implement the conditions or satisfy those qualification gates.

## Source and scope

The pinned English SRD 5.2.1 PDF has SHA-256 `8974902d109d6e63672d7c490bde9ccf052410503d9cfa768237154fbc5e3d87`. The source-derived decision retains the CC-BY-4.0 attribution and adaptation notice in [the source register](../srd-5.2.1-conformance.md).

| Condition and source | Published execution owner | Exact consequences and boundary |
|---|---|---|
| Blinded, PDF 177 | #960 | Sight is unavailable and sight-required ability checks automatically fail. The affected creature has Disadvantage on its attacks; incoming attacks have Advantage. Separately qualified senses remain distinct. This grants no Blindsight, source instance or invented duration. |
| Deafened, PDF 181 | #961 | Hearing is unavailable and hearing-required ability checks automatically fail. Unrelated checks and saving throws receive no invented penalty. Ending one source does not clear an independent contribution. |
| Invisible, PDF 184 | #962 | Initiative rolls have Advantage. Seen-target effects cannot affect the creature unless their creator can see it; worn/carried equipment is concealed. Attack Advantage/Disadvantage depends on whether the other creature can see the invisible creature. This grants neither Hide nor an undetectable position. Suppression retains its source and remaining duration. |

Trusted Blinded and Deafened ending adapters remain separately owned by #963/#964 under #193. Their callers own the action, target and payment. Spell, item and class sources retain their own admission, duration, expiry, cure, sense and suppression rules. This decision neither widens the existing ending owner #504 nor duplicates the source-specific effect implementations.

## Published contracts and dependencies

All three primary Tasks were published with the accepted Paladin decision and are Task/Size M. Their native prerequisites are:

| Task | Native parent | Exact blockers |
|---|---|---|
| #960 Blinded | #268 | #61, #65, #67, #268 |
| #961 Deafened | #268 | #61, #67, #268 |
| #962 Invisible | #268 | #61, #65, #66, #67, #268 |

#268 itself retains parent #193 and closed prerequisite #80. Its three implementation children remain open. Both qualification gates continue to require all three children after #268 closes. Existing titles, bodies, parents, prerequisites and Task/Size are retained; this PR supplies a decision and verification evidence only.

Each existing implementation contract requires inert previews, authoritative confirmation, source-bound state, injected dice, atomic cost/effect/history persistence, rollback, replay/reload/reconnect, concurrent-confirmation protection, disclosure controls and shared web/TUI availability. Those requirements remain work for the implementation PRs. Source catalogue presence, the published Tasks and closure of #268 establish no executable vision, hearing or invisibility support.

## Verification

The accompanying reconciliation and native verification receipts map each source clause to its existing owner. Source checks use the pinned PDF pages 177/181/184. The accepted Paladin native proof checks the exact published bodies, metadata, containment and prerequisite sets, including all six direct child-to-qualification relations. It also checks the complete dependency and containment graphs and all native caps.

The closure control changes only #268 to closed in memory and retains every implementation obligation. Independently deleting any one of the three children from either #279 or #280 must fail the native proof. No tracker body, hierarchy, prerequisite or implementation issue is changed by this decision.

No runtime files, grants, database behavior, UI or deployment changes are included. Full Go/vet and source execution evidence remains with the implementation and prior accepted publication PRs; this documentation follow-up uses focused source/native checks and its own hosted CI. Runtime implementation remains paused.
