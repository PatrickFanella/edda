# Charmed, Frightened and incapacitating condition ownership (#269)

All six named condition scopes already have bounded published implementation owners. This decision reuses **#487, #488, #727 and #777** and creates no duplicate implementation Task. The existing native paths require every owner from both spell/item qualification roots and character qualification. Those Tasks remain open; closing #269 completes source reconciliation and publication bookkeeping only.

## Source reconciliation

The pinned English SRD 5.2.1 PDF has SHA-256 `8974902d109d6e63672d7c490bde9ccf052410503d9cfa768237154fbc5e3d87`. Source-derived text retains the CC-BY-4.0 attribution and adaptation notice in [the source register](../srd-5.2.1-conformance.md).

| Condition | PDF | Existing execution owner | Source predicates retained by that contract |
|---|---:|---|---|
| Charmed | 178 | #488 | Cannot attack the charmer or target it with damaging abilities or magical effects. The charmer has Advantage on social ability checks against the recipient. This grants no universal obedience or control. |
| Frightened | 182 | #488 | Checks and attacks have Disadvantage while the fear source is in sight; the recipient cannot willingly approach that source. The line-of-sight qualifier applies to the roll penalties. Source identity and ending/immunity contributions remain distinct. |
| Incapacitated | 184 | #487 | No action, Bonus Action or Reaction; Concentration breaks; speech is unavailable; Initiative rolls have Disadvantage while Incapacitated. Other conditions consume these consequences instead of duplicating flags. |
| Stunned | 189 | #777 | Includes Incapacitated, automatic Strength/Dexterity save failure and incoming attack Advantage. Stunned alone adds neither Speed 0 nor automatic critical hits. Expiry removes only its source contribution. |
| Paralyzed | 186 | #727 | Includes Incapacitated; Speed 0 cannot increase; automatic Strength/Dexterity save failure; incoming attack Advantage; a hit is critical only when the attacker is within 5 feet. The attack still needs to hit. |
| Unconscious | 191 | #487 | Includes Incapacitated and Prone; held objects drop; Speed 0 cannot increase; the recipient is unaware; incoming attacks have Advantage; Strength/Dexterity saves automatically fail; hits within 5 feet are critical. Ending Unconscious leaves Prone. |

The published contracts require trusted source-bound state and preserve independent contributions. Source-specific spells, items, features, duration, immunity and cure events remain with their existing owners. #504 retains its exact trusted Paralyzed/Poisoned/Stunned endings; the Charmed/Frightened/Prone source-ending and immunity API remains in #488. No arbitrary client condition flag, default duration or generic spell-control framework is admitted by this decision.

## Existing publication and ownership

#487 is already a native child of #269. #488, #727 and #777 retain their accepted native parent #270 because their published scopes are shared across the condition groups. They are referenced rather than moved, copied or widened. #488 also owns Prone, which is reconciled separately by #270.

#269 retains native parent #193 and its closed prerequisite #80. The accompanying source/ownership matrix records every exact owner body hash, parent, blocker set and Task/Size. The native receipt records the concrete mandatory paths to #279, #280 and #337. Closing the decision changes none of those implementation dependencies.

The existing Tasks require authoritative source/actor/campaign state, inert previews, confirmation and revision checks, atomic state/history persistence, injected dice where a roll exists, rollback, same-action receipt replay, concurrent-confirmation protection, reload/reconnect, disclosure controls and consistent web/TUI availability. Those are acceptance requirements for the implementation PRs. They are not implemented by this documentation decision.

## Verification and limits

Source clauses were checked directly against PDF pages 178/182/184/186/189/191. A complete fresh native graph and focused Project readback checked the four exact published contracts and both native DAGs/caps. The source decision has no open hard prerequisite. The closure control keeps every implementation owner mandatory after #269 closes; removing an owner's complete incoming qualification paths fails that invariant.

No native issue creation, owner-body edit, hierarchy change or prerequisite addition is needed. Runtime implementation remains paused. This source/publication PR supplies no condition execution, class admission, database migration, UI behavior, grant, deployment or Full Rules qualification. Hosted CI applies to the decision head; the prior publication's full Go/vet evidence is retained without rerunning those expensive commands for documentation-only changes.
