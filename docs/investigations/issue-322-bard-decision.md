# Bard and College of Lore decision for #322

This decision assigns implementation work. It does not enable Bard creation,
advancement, features, spells, resource grants, coverage or Full Rules. Bard remains
unavailable and unqualified. Leave #322 open until this decision PR merges; its
implementation and qualification issues remain independent work.

The inspected checkout starts at `9ee2e4a6076fac030d4672835135b29ed2c713fa`.
GitHub is authoritative; Gitea is fetch-only. The parent owns independent acceptance,
hosted CI and merge. Concurrent multiclass raw-JSON validation repairs belong to the
parent and are outside this decision.

## Source and evidence

The English SRD 5.2.1 PDF SHA-256 was checked before direct reading:
`8974902d109d6e63672d7c490bde9ccf052410503d9cfa768237154fbc5e3d87`.
Read PDF pp31–35 for all Bard rows, feature prose, spell lists and College of Lore;
pp23–26 for advancement, higher-level starts and multiclass acquisition; pp104–107
for preparation, components, slots and casting; p154 for both Power Word spells;
and glossary rest/condition passages on pp184–189. Attribution, CC-BY-4.0 and the
adaptation notice remain in `docs/srd-5.2.1-conformance.md` and the matrix source pin.

The [matrix](issue-322-bard-scope.json) copies the canonical Bard batch's 69 records,
Lore's 15 records and the 20 global advancement rows exactly, including nested
conditions, choices, alternatives, numerical parameters, resource formulas and
all nine slot columns. It maps all 129 unique required Bard/Lore mechanic keys
to one implementation owner and every record to its unavailable state and actual
evidence. Context records in child bodies do not transfer shared effect ownership.

The existing source comparators independently compare the PDF with these catalogues.
Bard's source proof checks all 20 rows, 13 named identities and all slot columns;
Lore's proof checks all 15 records and rejects 76 contrary controls. Advancement's
proof checks every XP/PB identity and value on p23. These are metadata proofs.

`creation.go` admits fixed Human Soldier Fighter and Human Acolyte Cleric choices.
`character.go` projects their supported equipment/casting boundaries. `advancement.go`
and `game/advancement_preview.go` admit the bounded Fighter 1→2 fixture. None admits
Bard. Bard/Lore catalogue readiness tests explicitly reject production admission and
mechanic qualification. `engine/action_confirmation.go` supplies a transactional
receipt/replay pattern; it does not execute these Bard mechanics. Its existing
out-of-combat confirmation guard means combat reactions need their own authoritative
encounter path rather than a bypass of that guard.

The current web `ActionPreviewPanel` and TUI router/creation form use the shared
engine's supported actions. Their presence is not a Bard UI. Every new terminal
requires explicit availability or an unavailable reason on both clients, inert
preview, explicit confirmation/cancel, reload/reconnect receipt behavior, and
player-known disclosure. Web changes require desktop/phone confirmation journeys,
build and focused behavior/lint checks. Future mutations need isolated PostgreSQL
rollback, replay, reload and concurrent-confirmation tests; this decision runs no
production mutation or inference.

## Chosen ownership boundaries

| Work | Exact owners and boundary |
| --- | --- |
| Casting and list identity | #502 consumes #273's casting receipt contract; exact slot/free-use identity and instrument focus. #503 joins Bard/Cleric/Druid/Wizard source lists and preserves unsupported execution. |
| Shared effects | #504 ends only trusted Paralyzed/Poisoned/Stunned instances; #505 Power Word Heal, #506 Power Word Kill. Existing #487/#488 own their shared condition scopes; #351/#353 retain concentration/effect lifecycle. |
| HP and creation | #507 consumes #429's reusable XP/PB/CON transaction with d8/fixed-five. #508 first-class traits/equipment; #509 later-class CHA eligibility/reduced traits; #430 supplies proficiency provenance. |
| Inspiration | #510 capacity, die progression and completed-rest recharge; #511 confer and expire a recipient die; #512 consume it after a failed D20 Test. #427/#428 retain completed-rest semantics. |
| Spell preparation | #513 counts and per-Bard-level replacement; #514 slots and Long Rest recovery; #523 Magical Secrets; #524 Lore Magical Discoveries. Spell execution remains with exact #277 terminals. |
| Skills and subclass | #515 Expertise, #516 Jack of All Trades, #517 Lore selection, #518 Lore bonus skills, #519 Cutting Words. |
| Feats and higher features | #520 repeated improvement grants; #521 Font slot conversion; #522 Countercharm; #525 Peerless Skill; #526 Superior Inspiration; #527 level19 feat grant; #528 Words of Creation. |
| Application | #529–#548 each own exactly one first-class admission or adjacent Bard-level transition. #549 owns later-class Bard1 application. Feature children own rules; these consumers apply exact table deltas and grant identities through shared transactions. |

First-class Bard gets DEX/CHA saves, any three skills, three instrument proficiencies,
Simple weapons and Light armor. Equipment A includes Leather Armor, two Daggers,
one chosen instrument, an Entertainer's Pack and 19 GP; B is 90 GP. Equipment's
instrument choice is separate from instrument proficiency choices. Later-class
Bard gets one skill, one instrument and Light armor, with no first-class equipment,
weapon or save grants. CHA and every current class's primary abilities must meet
the actual multiclass prerequisites. #431's Barbarian STR adapter is not reused.

#333 remains the origin decision; #430 provides provenance/duplicate validation,
not background/species effects. Unsupported origins stay unavailable. #335 retains
combined slots, total-level PB and mixed-class interactions. The Bard adapter is a
terminal consumer of that contract, not blanket support for all combinations.
Unsupported mixed configurations stay unavailable pending their exact terminals.
Higher-level starts consume the ordinary admissions; source rows and isolated
fixture entitlements never grant production eligibility.

ASI grants occur at Bard4/8/12/16, with distinct grant identities; Bard19 offers an
Epic Boon or another qualified feat. Recommendations are optional. #384 owns
selection, #389 the shared ability updater/caps, and #399 Spell Recall's effect.
All seventeen exact feat effect terminals from #334 are prerequisites of the
Bard improvement/boon consumers and Bard gate. Their source prerequisites still
control selection; this does not make Fighting Style or other ineligible choices
available without the actual named feature. Each transition consumes those
contracts before exposing an option. There is
no new feat effect implementation here and no invented Fighter6/14 grant.

## Rules that must survive implementation

Inspiration capacity is `max(1, CHA modifier)`, not a class-level fixed number.
Dice are d6/d8/d10/d12 at Bard1/5/10/15. Conferring costs a Bonus Action and one use,
targets another creature within 60 feet that can see or hear the Bard, permits only
one held die, and lasts one hour. A held die is consumed when rolled after a failed
D20 Test, even if the test still fails. Store its source and absolute expiry.
Advancement or reconnect cannot reroll a committed die or silently restart expiry.

Before Font, completed Long Rest recovers all uses. At Bard5, completed Short or
Long Rest does so. Font separately spends any one spell slot, without an action,
to regain exactly one expended use. That conversion is not casting a spell.
Superior Inspiration at Bard18 recovers to two on Initiative when below two;
it neither adds two to one nor lowers three. Preserve the source-specific threshold
even with an ordinary capacity of one; do not silently clamp it to that capacity.

Expertise selects two already-proficient skills at Bard2 and two more at Bard9.
Jack of All Trades adds floor(PB/2) only to a skill check lacking proficiency and
not otherwise using PB. It does not apply to ordinary initiative or every d20 roll.
Lore3 adds three skills. Cutting Words uses a Reaction and one Inspiration use
against a visible creature within 60 feet on a damage roll or successful check/attack.
Do not invent a hearing, language or charm-immunity restriction.

Countercharm uses a Reaction after self or a creature within 30 feet fails a save
against a Charmed/Frightened effect; the reroll has Advantage. No Inspiration cost
or hearing/sight requirement is stated. Peerless Skill applies to the Bard's own
failed check or attack at Lore14; a continued failure retains the use, but the
rolled attempt must still be saved to prevent retry fishing. Neither feature
applies to arbitrary successful rolls or makes another character's choices silently.

Bard knows two cantrips initially, gains one at4 and10, and may replace one at
each Bard-level gain. Four level1 spells are prepared initially; table increases
add slot-eligible spells and each Bard level may replace one prepared spell.
Always-prepared spells do not consume the table count. Magical Secrets expands
new preparations and replacements to Bard/Cleric/Druid/Wizard at Bard10; it does
not grant a fixed pair. Lore6 Magical Discoveries separately chooses two spells
from Cleric/Druid/Wizard, including eligible cantrips, always prepared outside
the regular count; replace at most one on each Bard level.

Words of Creation always prepares Power Word Heal/Kill at Bard20, outside the
22 regular spells. An explicitly chosen second creature must be within 10 feet
of the first and satisfy the base spell's target predicates. Commit both effects
and one cast cost atomically. Heal restores all HP, ends its five named conditions,
and offers a Prone target its own optional Reaction to stand. Kill causes death
at HP≤100 or 12d12 Psychic damage above100; no saving throw or hidden-HP preview.

## Native graph and pending shared-body publication

GitHub permits at most 100 blockers per issue. The parent authorized the explicit
hierarchy: all 48 new terminals and 38 reused terminals block Bard gate #501;
#322 also blocks #501; #501 directly blocks #337. All historical #334 direct
blockers remain required. No terminal was dropped to make the graph fit.

#283 already contained 94 native children before this publication. Bard-specific
children therefore have native parent #501, itself under #283. Shared additions
#502/#504 retain parent #193; #503/#505/#506 retain parent #191. Their additional
direct gates are #279 and #280. Every shared reused terminal retains its parent
and body. Native containment and prerequisite dependencies are separate graphs:
a child contained by #501 while blocking #501 is expected, not a dependency cycle.

The [receipt](issue-322-link-verification.json) records exact titles, complete bodies,
body hashes, native parents/prerequisites, Task types, Edda Project sizes, source
prerequisite states, preserved owner snapshots and the full repository graph.
The verifier checks dependency and parent acyclicity separately and tests corruption
of both graphs, source data/edges, children, body hashes, prerequisite edges and gate
paths. Its API reader is bounded; synthetic pagination parses query parameters
exactly, including page10, rather than matching substrings.

The four `issue-322-parent-{47,54,283,337}-supplement.md` files are exact append-only
marked supplements for the parent to apply serially. This child does not edit those
live bodies. Receipt status `parent-body pending` is intentional and distinct from
verified native publication. Parent acceptance must append each section once,
read it back, and rerun the live verifier with `--require-supplements` before merge; it then records
`published`. External edits outside marked sections are not overwritten.

## Reproduction

Run these commands from this checkout, with the host's `rtk proxy` prefix:

```sh
python3 scripts/verify_srd_bard.py /tmp/edda-wave1/SRD_CC_v5.2.1.pdf
python3 scripts/verify_srd_college_of_lore.py /tmp/edda-wave1/SRD_CC_v5.2.1.pdf
python3 scripts/verify_srd_advancement.py /tmp/edda-wave1/SRD_CC_v5.2.1.pdf
python3 scripts/verify_bard_scope.py --probes --pdf /tmp/edda-wave1/SRD_CC_v5.2.1.pdf
python3 scripts/verify_bard_scope.py --live --probes --pdf /tmp/edda-wave1/SRD_CC_v5.2.1.pdf
GOMAXPROCS=2 go test -p 2 -short -race ./...
GOMAXPROCS=2 go vet -p 2 ./...
git diff --check
```

Live verification only reads GitHub and writes its local receipt after successful
validation. No host/toolchain changes or linter installation are included; the
installed golangci-lint shim is known unavailable. Local checks do not establish
hosted CI, implementation acceptance, playable Bard support or production readiness.

## All twenty source rows and application owners

| Bard level | PB | Die | Cantrips | Prepared | Slots 1–9 | Transition owner |
| --- | --- | --- | --- | --- | --- | --- |
| 1 | 2 | d6 | 2 | 4 | 2/0/0/0/0/0/0/0/0 | #529 |
| 2 | 2 | d6 | 2 | 5 | 3/0/0/0/0/0/0/0/0 | #530 |
| 3 | 2 | d6 | 2 | 6 | 4/2/0/0/0/0/0/0/0 | #531 |
| 4 | 2 | d6 | 3 | 7 | 4/3/0/0/0/0/0/0/0 | #532 |
| 5 | 3 | d8 | 3 | 9 | 4/3/2/0/0/0/0/0/0 | #533 |
| 6 | 3 | d8 | 3 | 10 | 4/3/3/0/0/0/0/0/0 | #534 |
| 7 | 3 | d8 | 3 | 11 | 4/3/3/1/0/0/0/0/0 | #535 |
| 8 | 3 | d8 | 3 | 12 | 4/3/3/2/0/0/0/0/0 | #536 |
| 9 | 4 | d8 | 3 | 14 | 4/3/3/3/1/0/0/0/0 | #537 |
| 10 | 4 | d10 | 4 | 15 | 4/3/3/3/2/0/0/0/0 | #538 |
| 11 | 4 | d10 | 4 | 16 | 4/3/3/3/2/1/0/0/0 | #539 |
| 12 | 4 | d10 | 4 | 16 | 4/3/3/3/2/1/0/0/0 | #540 |
| 13 | 5 | d10 | 4 | 17 | 4/3/3/3/2/1/1/0/0 | #541 |
| 14 | 5 | d10 | 4 | 17 | 4/3/3/3/2/1/1/0/0 | #542 |
| 15 | 5 | d12 | 4 | 18 | 4/3/3/3/2/1/1/1/0 | #543 |
| 16 | 5 | d12 | 4 | 18 | 4/3/3/3/2/1/1/1/0 | #544 |
| 17 | 6 | d12 | 4 | 19 | 4/3/3/3/2/1/1/1/1 | #545 |
| 18 | 6 | d12 | 4 | 20 | 4/3/3/3/3/1/1/1/1 | #546 |
| 19 | 6 | d12 | 4 | 21 | 4/3/3/3/3/2/1/1/1 | #547 |
| 20 | 6 | d12 | 4 | 22 | 4/3/3/3/3/2/2/1/1 | #548 |

Zero denotes no slot at that spell level. The exact feature references and table-only
improvements remain in each transition body and the machine-checkable matrix.

## Local checks recorded for this decision

The Bard, Lore and advancement PDF comparators passed against the pinned artifact.
The full `GOMAXPROCS=2 go test -p 2 -short -race ./...` run passed; the SRD package
reported 239.881 seconds. `GOMAXPROCS=2 go vet -p 2 ./...` passed with default limits.
The unchanged #334 offline verifier also passed its 17-feat, 146-key and historical
native-gate controls. The final publication receipt and new verifier report their
own exact graph/body/Project readback and corruption controls. No child runtime
implementation, database integration mutation or client qualification was performed.

Final native readback observed 2026-11-16T14:59:08Z: #501 has 87 prerequisites
(86 terminals plus #322); #337 has 49 prerequisites, including #501. The full
repository graph contains 406 issue nodes and is acyclic in each of the separate
containment and dependency checks. All 49 newly published issues have exact bodies,
Task types and Edda Project sizes. The receipt also reads Project metadata for the
38 reused terminals and seals 53 existing owners' bodies/parents/prerequisites.
Fourteen corruption probes passed, as did exact-query pagination through page10
and the 20-call pagination cap. Contrary PDF input was rejected before remote
access. The parent acceptance flag correctly rejects the four pending supplements.
