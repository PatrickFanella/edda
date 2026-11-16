# Wizard/Evoker source decision (#332)

This is an audit checkpoint. Wizard remains unavailable in the engine. Native
publication, the combined capacity proof and accepted POST-WARLOCK prefix release
are pending. This document does not satisfy issue #332's closure criteria yet.

The prepared clean base was `e7cd1ba8318ba296edd4fc5b6b84715ca11016f2`, branch
`feat/wizard-evoker-roadmap`, physical path
`/home/shared/Work/Worktrees/edda/t3-18f52b24-e7041b6e3b02824b`, common Git directory
`/home/shared/Work/Subcult/Games/edda/.git`. Origin is GitHub; Gitea push is disabled.

The source is SRD 5.2.1, pinned SHA-256
`8974902d109d6e63672d7c490bde9ccf052410503d9cfa768237154fbc5e3d87`.
Directly audited pages: advancement/multiclass 23–26, Wizard 77–81, Evoker 82,
casting 104–106 and each referenced residual spell's exact page. Source-derived
transcriptions and execution boundaries are adaptations. CC-BY-4.0 attribution
and license: [conformance register](../srd-5.2.1-conformance.md).

`issue-332-wizard-scope.json` retains all 106 complete Wizard/Evoker source records,
all 20 Wizard rows, every mechanic/option context and its execution owner. Catalogue
and source comparators are metadata evidence only. The independent readable owner
oracle includes nested prerequisite records and splits large source features among
actual implementation contracts. It does not import the generator's classifier.
The composite multiclass Hit Die plus level-1-feature key belongs to later-entry,
which requires the separate HP adapter. It does not belong to HP by substring.

## Observed execution boundaries

| Boundary | Actual behavior at the prepared base | Consequence |
|---|---|---|
| `internal/rules/srd521/character.go:97` | Project admits only fixed Fighter/Cleric fixtures | No Wizard sheet or production entitlement |
| `internal/rules/srd521/advancement.go:29` | AdvanceFighter implements only Fighter 1→2 | All Wizard admissions and transitions need separate execution |
| `internal/engine/action_confirmation.go:330` | Advancement confirmation invokes AdvanceFighter | Generic durable confirmations do not implement Wizard |
| `internal/game/spell_preview.go:21` | Preview describes the supported Cleric Cure Wounds fixture | No Wizard casting access |
| `internal/game/advancement_preview.go:21` | Preview describes Fighter advancement | No Wizard level applications |
| `internal/combat` | Shared combat primitives; no Wizard/Evoker route found by targeted source search | Named modifiers and source timing require implementation |
| `frontend/src/lib/startupWorkflow.ts` | Wizard descriptive labels | Labels provide no feature or level execution |
| `tui/character` | Consumes shared character state | A generic sheet does not qualify Wizard choices |

No database, runtime, browser, device or Full Rules qualification was performed.
The future children require isolated PostgreSQL atomicity, replay/reload, concurrency,
and shared web/TUI confirmation and availability tests before playable admission.

## Source scope and routing

The independent column extraction reconstructs all 217 printed Wizard identities,
with level, school, special flags and page, in `issue-332-wizard-spell-list.json`.
133 identities overlap Sorcerer's consumed142, 95 overlap Warlock's consumed117.
The 56 outside both are identities to audit, not automatically new effect owners.
Six reuse existing whole-body owners: Acid Arrow #354, Gentle Repose #922,
Locate Creature #941, Locate Object #923, Longstrider #1066 and Nondetection #1071.
Acid Arrow #354 is level-2-only; its separate higher-slot adapter retains #354's
exact exclusion and consumes its initial/delayed damage outputs.

The candidate contains 60 class/level contracts, 179 residual shared effect parts and six bounded generic
form/snapshot contracts, with 40 qualification nodes. These counts are provisional. Spellbook
identity, six-spell initialization, learning, found/own copying, lost-book recovery,
preparation, rest replacement, slots, multiclass slot/preparation separation,
ritual modes, recovery, Mastery selection/casting/replacement, Signature independent
resources and the Evoker alternatives have distinct bounded contracts. Every
individual level application reaches its required execution alternatives.
Class rows and HP use Wizard level; XP/live PB use the complete canonical admitted
TOTAL level. Unequal Wizard1/total5 and Wizard2/total6 witnesses and bool/fractional/
total21 controls are retained. ASI owns feat admission; Spell Recall #399 requires
actual Spellcasting without imposing caster or active-encounter prerequisites on
ordinary feats.

The root reserved one shared-output head under #332 and ten printed-level cohorts
0–9 beneath it. Each lists its exact source identities and whole output owners;
finite output slices are bounded at 80 incoming prerequisites. Shared outputs
route to the reserved exactly-three-class aggregate under #191, then #279/#280,
and through the own Wizard class gate under #332 to #337. No Wizard direct edge
to #279/#280, no extra #283 child or #429 consumer is proposed. Own D6/fixed4 HP
uses #968, which qualifies the existing #429 transaction and owns no die/class
implementation. Shared casting adapters reuse #330; Find Familiar forms/control
reuse #331. Planning #181/#185/#273/#274/#277 never supplies runtime execution.

## Source conflicts

| Printed list locator | Header locator | Conflict and admission policy |
|---|---|---|
| Sorcerer level 2, PDF68 | Mind Spike, PDF149 | Header includes Sorcerer; printed list omits it. No list admission. Effect owner remains Warlock #331. |
| Sorcerer level 2, PDF68 | Phantasmal Force, PDF151 | Header includes Sorcerer; printed list omits it. No list admission. Unallocated outside current union. |
| Wizard level 2, PDF80 | Phantasmal Force, PDF151 | Header includes Wizard; printed list omits it. No list admission. Unallocated outside current union. |

Contrary controls include Wizard Mind Spike, whose list80/header149 agree, and
Warlock Mind Spike, whose list75/header149 agree. All 339 spell headers are compared
against the complete three printed lists; exactly these three header-only pairs
are expected. An explicit feature grant requires its own source/qualification
route. General Wish/scroll/any-spell permissions do not qualify every effect.
The reserved union remains 245; Phantasmal Force is not a silently added 246th.

## Remaining audit and publication gates

- The bounded form reservation retains all 235 canonical typed monster dispatch
  identities and 26 explicitly proposed generic form bindings. Skeleton statistics
  and attacks reuse #331; Animate Dead creation, commands and 24-hour control remain
  Wizard-owned. Own Riding Horse and Commoner outputs provide concrete Beast and
  Humanoid witnesses. Shapechange and Simulacrum permission does not execute unbound
  statblocks. Named interoperability owners are explicit candidate dependencies.
- The complete combined proposal currently passes all three caps and both DAGs:
  1,017 native + 305 Sorcerer + 282 Warlock + 285 Wizard = 1,889 nodes. The
  proposal receipt records exact input hashes. #285 has 82 outgoing, #67 has 100,
  #279 has 100 incoming, #280 has 95, #283 has 95 children and #429 has 100
  outgoing. Recheck changed inputs and obtain root publication approval. This is
  a proposed graph, not a native publication receipt.
- Publish only approved bounded actual children; read back exact bodies, titles,
  parent, prerequisites, Task and Project Size. No placeholders qualify.
- Root alone applies the four compact supplements. Each remains a local PENDING
  artifact. #47 append is at most70 characters, other appends at most750.
- Freeze immutable parent prefixes only after the root's accepted POST-WARLOCK
  release and real four appends. No current parent hash is a final prefix seal.

## Verification

`python scripts/verify_wizard_scope.py --pdf /tmp/edda-wave1/SRD_CC_v5.2.1.pdf`
checks the ACTUAL candidate fixtures and source texts, complete metadata graphs,
217-entry roster, 339 headers and source conflicts. `--final` deliberately refuses
this unpublished checkpoint. It must not produce a final native proof.

`python -m unittest discover -s scripts -p test_wizard_scope.py` rejects coherent
context rerouting, actual fixture/body weakening, changed numeric spell operands,
missing alternatives, multiclass HP dependency removal, adjacent transition removal
and blanket caster gating of ordinary feats. These are document/plan checks.

One authorized serial full run passed on the unchanged runtime tree:
`GOMAXPROCS=2 go test -p2 -short -race ./...` in 234.060 seconds and
`GOMAXPROCS=2 go vet -p2 ./...` in 2.924 seconds, both exit0. Logs and timing are in
the local validation receipt. No toolchain changes or unavailable linter installs.
Hosted CI, independent acceptance, native publication and final seal remain separate.
