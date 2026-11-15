# Character and advancement expansion decision (#74)

Observed November 15, 2026 against `c802194` after #73 and #83 merged.
This is a source inventory and executable batch decision under #47/#54.
Closing the decision does not implement its follow-ups or qualify Full Rules.

## Source and inventory

The English SRD 5.2.1 PDF matches the repository pin:
`8974902d109d6e63672d7c490bde9ccf052410503d9cfa768237154fbc5e3d87`.
The companion [inventory](issue-74-character-inventory.json) contains 344
planning identities extracted from the pinned PDF:

| Source family | Identities | PDF pages |
| --- | ---: | --- |
| Classes | 12 | 28–82 |
| Class feature headings | 174 | 28–79 |
| Subclasses | 12 | 30–82 |
| Subclass feature headings | 58 | 30–82 |
| Metamagic options | 10 | 66–67 |
| Eldritch Invocation options | 28 | 72–74 |
| Backgrounds | 4 | 83 |
| Species | 9 | 84–86 |
| Feats | 17 | 87–88 |
| Character levels | 20 | 23 |

These counts describe named headings and required level scopes, rather than
playable effects or every nested option. A feature heading that improves at
several later levels remains one heading; the class data owner must preserve
all 20 progression rows and every table-only change. Class spell-list membership
is a reference to the #80 spell catalogue, never a second spell/effect owner.
Lineages, ancestry, size, languages, proficiencies, ability assignments, equipment
packages, feat choices and feature options must become explicit nested records
and required mechanic keys during import. Planning IDs are deliberately separate
from runtime content IDs; source inventory does not register executable content.

`scripts/inventory_srd_characters.py` checks the PDF hash and reproduces the
name/page/owner-key inventory offline with Poppler. It handles wrapped headings,
class boundaries sharing a page, repeated feature names at different levels and
subclass boundaries. The manifest records one exact owning data issue for every
identity. Names/pages are adapted metadata; no rules text or numeric progression
catalogue is imported in this investigation.

Some requirements do not have independent named catalogue headings. The manifest
therefore also assigns exact source sections for creation, higher-level starts,
multiclassing, nested options and optional advancement beyond level 20. These
remain mandatory full-coverage requirements. They cannot disappear from the
qualification gate because they were not counted among the 344 headings/scopes.

## Reconciliation with the implementation

| Current evidence | Supported boundary | Remaining work |
| --- | --- | --- |
| `creation.go`, `creation_test.go`, martial/caster creation and interview integration fixtures | Fixed level-one Human Soldier Fighter and Human Acolyte Cleric; guided and interview paths share validation | Other builds, species/background choices, general ability assignment, proficiencies, feats, equipment packages and higher-level starts |
| `character.go`, `character_test.go`, `attack.go` and durable martial-action fixtures | Derived sheet, checks/saves and bounded ordinary longsword fixture; fixed Defense/armor choices | All 12 complete classes, subclasses, general feature options and conditional interactions |
| `advancement.go`, `advancement_test.go`, `advancement_integration_test.go` | Fighter 1→2 after earned XP with fixed/rolled HP, one saved decision and replay; new Action Surge/Tactical Mind are recorded read-only features | Other transitions, executable new features, feat/subclass decisions, broader XP/HP/recovery and multiclass progression |
| `spell.go`, `lifecycle.go` and casting/lifecycle integration fixtures | Narrow conscious-self Cure Wounds and Shield of Faith plus bounded rest/concentration state | Recorded prepared/cantrip/feat spells are not general executable spell support; #80 owns wider spell/effect catalogue |
| `coverage.json` | Broad character families, classes, subclasses and levels remain unsupported | Source-complete data, required mechanics and reviewed qualification evidence must all exist before promotion |

Creation records Savage Attacker, Skilled and Magic Initiate selections. Recording
a choice, mastery or spell is not proof that its full effect is executable.
Likewise, the Fighter level-two projection does not prove all Fighter levels or
all level-two characters. Existing legacy characters stay separate; no inferred
or bulk conversion is part of this decision.

## Batch decision and delivery boundaries

Tracking groups #282 (catalogue) and #283 (mechanics/qualification) beneath
#54 own 56 one-PR work units, #284–#339:

- One versioned character-record contract with one Fighter record fixture.
- 36 metadata-only data PRs: 12 classes, 12 subclasses, two origin groups,
  four feat categories, four option batches, one advancement table and one
  multiclass table/exception inventory. Invocation batches contain at most 12
  options; other named option groups contain at most 10. Each class batch is
  metadata for one class, including its 20 rows and feature references, without
  implementing those mechanics or transcribing prose.
- Two vertical runtime PRs activate the existing level-two Fighter’s Action
  Surge extra action and Tactical Mind post-failure check choice, each with
  deterministic costs, action restrictions and durable retry/reload fixtures.
- 16 bounded decisions: one per class, plus origin choices, feat effects,
  multiclass runtime and remaining advancement/higher-level starts. Each compares
  every assigned source key with actual code and publishes exact vertical
  implementation children before closing.
- One complete character qualification gate #337 joining all required implementations.

Each data record carries exact source IDs/pages, nested choice/prerequisite
references and required mechanic keys. Unknown or unimplemented keys keep it
unavailable. Importing metadata cannot change current creation/action availability.
If a data batch exceeds a reviewable PR, its owner must publish exact split
children before widening it. Spell, item, condition, summon and exploration
mechanics retain the owners established by #80/#82; new character children add
cross-domain prerequisites to those owners rather than duplicating them.

Every decision must publish source-key-specific implementation children: concrete
selection/target, cost, timing, source-derived result, stale-preview refusal,
prepare/confirm, once-only persistence, retry/reload and shared web/TUI behavior.
Class splitters own class-specific level transitions and features. The general
advancement decision owns shared XP/HP/choice validation and higher-level starts;
origin/feat decisions own common choices and effects. Shared prerequisites must
be explicit. Their terminal implementations must be native blockers of the full
character gate before each splitter closes. Closed decisions and imported data
alone cannot satisfy that gate.

The Warlock decision has six genuine independent catalogue prerequisites; the
final gate joins 16 independent decisions and two concrete runtime fixtures; their implementation blockers must
be added. Those graph warnings are accepted: a serial chain would obscure the
actual dependencies. The data-only class boundaries and bounded splitters keep
individual delivery reviewable without claiming the remaining rules fit in one
implementation PR.

#88 reports partial coverage while these scopes remain unimplemented. #89
qualifies only its declared playable slice. Complete spell/item/world gates
#279/#280/#183 remain separate; #54 stays open until every required domain and
its implementation/qualification children complete or the owner explicitly
narrows the target.

## Verification and attribution

Offline reproduction with the pinned local PDF:

```sh
python3 scripts/inventory_srd_characters.py --pdf /tmp/edda-srd521-wave11.pdf --check docs/investigations/issue-74-character-inventory.json
```

Tracker publication reads back all bodies, native parents, exact blocker sets,
acyclic dependency closure and Edda Project Size M. The manifest records the
published IDs, group ownership and execution order. The contract #284 and Fighter runtime fixtures #338/#339 form the first
frontier after #74; all data follows the contract. Decisions follow their exact source
batches, and the full gate follows the decisions and their required children.

This work includes material from the System Reference Document 5.2.1 (“SRD
5.2.1”) by Wizards of the Coast LLC, available at https://www.dndbeyond.com/srd.
The SRD 5.2.1 is licensed under the Creative Commons Attribution 4.0 International
License, available at https://creativecommons.org/licenses/by/4.0/legalcode.
Names/page references are adapted into Edda planning metadata. See the
[pin, license and attribution](../srd-5.2.1-conformance.md).
