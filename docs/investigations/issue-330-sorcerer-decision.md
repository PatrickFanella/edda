# Sorcerer / Metamagic / Draconic implementation decision (#330)

This decision reconciles **103 source records, 278 contextual mechanic/option occurrences and all 20 Sorcerer rows**. The approved publication contains **305 Sorcerer Task/M issues: 67 class/level implementations, 211 shared implementations and 27 qualification nodes**. All remain open. Their complete native bodies, identities, parents, prerequisites and Project Size were collected and verified as part of the 871-task caster publication at **2026-11-17 06:21 UTC**. The four parent supplements are published, and fresh strict native/PDF validation passed at **2026-11-17 06:33 UTC**. Final independent PR acceptance and CI remain pending. Sorcerer runtime remains unavailable.

The original source audit started at `e7cd1ba8318ba296edd4fc5b6b84715ca11016f2`. This acceptance follow-up starts from merged coordinator recovery `c8dd99e716f264054896b399bcac464326e93de8`. The earlier merged checkpoint did not complete the decision; #330 remains open until this follow-up meets its publication and acceptance gates. The [native publication record](issue-330-native-publication.json) binds the actual issue identities and immutable milestone provenance. GitHub is authoritative; its Gitea copy remains a fetch-only mirror.

Pinned English SRD5.2.1 PDF SHA256 `8974902d109d6e63672d7c490bde9ccf052410503d9cfa768237154fbc5e3d87`; directly extracted class64-66, Metamagic66-67, spell lists67-69, Draconic69-70, advancement/multiclass23-26, casting104-106, Summon Dragon/Draconic Spirit166 and future Tome73-74. The independent source comparators check all source operands against that pin. Source-derived records retain the CC-BY-4.0 attribution/adaptation in [the source register](../srd-5.2.1-conformance.md). No PDF or private lore is committed.

Live #330 is open, parent283, blocked by closed285/295/305/317. Their closures imported source metadata only. Counts computed from canonical JSON: **63+22+18=103 records**, **278 contextual mechanic/option occurrences**. The [JSON matrix](issue-330-sorcerer-scope.json) retains every record field and exact context, source-key ownership and findings across rules/game/engine/combat/handlers/database/web/TUI. Shared spell references in class records own grant/adaptation context only. The 211 shared source contracts have independent fixture/body expectations and retain their exact execution owners. A class preparation adapter never implements a referenced spell.

`catalogue.go` returns false from `qualifiedCharacterMechanic`; `creation.go` admits fixed Human Soldier Fighter / Human Acolyte Cleric only. `advancement.go` and game advancement preview admit Fighter1→2. `spell.go` explicitly retains a bounded Cleric Cure Wounds fixture. Generic JSON storage/transactions in state/sqlc do not implement source predicates. Frontend startup labels contain Sorcerer but grant no admission. TUI character view is a placeholder; its shared world flow supplies no Sorcerer mechanics. Unknown source variants remain unavailable.

## All twenty Sorcerer rows

Features/points/slots/preparations use authoritative **Sorcerer class level**. XP/livePB use the COMPLETE canonical285 advancement table selected by the sum of admitted class levels AFTER transition. The PB below is the printed single-class context. `—` preserves absent positive capacity rather than inventing a source zero.

| Class level | Printed PB | Points | Cantrips | Prepared | Metamagic options | Slots 1/2/3/4/5/6/7/8/9 | New features |
|---|---|---|---|---|---|---|---|
| 1 | +2 | — | 4 | 2 | — | 2/—/—/—/—/—/—/—/— | spellcasting, innate sorcery |
| 2 | +2 | 2 | 4 | 4 | 2 | 3/—/—/—/—/—/—/—/— | font of magic, metamagic |
| 3 | +2 | 3 | 4 | 6 | 2 | 4/2/—/—/—/—/—/—/— | sorcerer subclass |
| 4 | +2 | 4 | 5 | 7 | 2 | 4/3/—/—/—/—/—/—/— | ability score improvement |
| 5 | +3 | 5 | 5 | 9 | 2 | 4/3/2/—/—/—/—/—/— | sorcerous restoration |
| 6 | +3 | 6 | 5 | 10 | 2 | 4/3/3/—/—/—/—/—/— | subclass feature |
| 7 | +3 | 7 | 5 | 11 | 2 | 4/3/3/1/—/—/—/—/— | sorcery incarnate |
| 8 | +3 | 8 | 5 | 12 | 2 | 4/3/3/2/—/—/—/—/— | ability score improvement |
| 9 | +4 | 9 | 5 | 14 | 2 | 4/3/3/3/1/—/—/—/— | — |
| 10 | +4 | 10 | 6 | 15 | 4 | 4/3/3/3/2/—/—/—/— | metamagic |
| 11 | +4 | 11 | 6 | 16 | 4 | 4/3/3/3/2/1/—/—/— | — |
| 12 | +4 | 12 | 6 | 16 | 4 | 4/3/3/3/2/1/—/—/— | ability score improvement |
| 13 | +5 | 13 | 6 | 17 | 4 | 4/3/3/3/2/1/1/—/— | — |
| 14 | +5 | 14 | 6 | 17 | 4 | 4/3/3/3/2/1/1/—/— | subclass feature |
| 15 | +5 | 15 | 6 | 18 | 4 | 4/3/3/3/2/1/1/1/— | — |
| 16 | +5 | 16 | 6 | 18 | 4 | 4/3/3/3/2/1/1/1/— | ability score improvement |
| 17 | +6 | 17 | 6 | 19 | 6 | 4/3/3/3/2/1/1/1/1 | metamagic |
| 18 | +6 | 18 | 6 | 20 | 6 | 4/3/3/3/3/1/1/1/1 | subclass feature |
| 19 | +6 | 19 | 6 | 21 | 6 | 4/3/3/3/3/2/1/1/1 | epic boon |
| 20 | +6 | 20 | 6 | 22 | 6 | 4/3/3/3/3/2/2/1/1 | arcane apotheosis |

Unequal witnesses: Fighter4→Sorcerer1 selects class1 but total5/6500XP/PB3 and later-class D6/fixed4HP. Fighter4/Sorcerer1→Sorcerer2 selects class2 but total6/14000XP/PB3. Reject booleans, decimals, zero, nonadjacent changes, total>20 and unqualified class combinations. First-character creation is level1 only, and it has no blanket multiclass prerequisite. Later entry never uses first-character HP/proficiency/equipment grants. Own HP adapter consumes968→429, with no new direct429edge and no mutation to either old contract.

## Contracts and split boundaries

The67 class contracts include47 distinct feature/adapter leaves and20 application leaves. Each contains source predicates, deterministic positive and independently violated negative controls, exact costs/actions/target/timing and isolated entitlement boundaries. Every contract repeats prepare/confirm/retry/reload/concurrency/rollback/isolatedPostgreSQL and shared web+TUI availability requirements. Source-predicate execution is separate from feature behavior, creation/proficiencies/equipment/choices, HP/later-entry, subclass, improvements, resource/slots/preparation and rest recovery.

Font splits pool, slot→points, points→slot and created-slot ending. Innate splits activation, benefits and LongRest recovery. Metamagic splits selection, combination protocol and all ten actual2024 options. Restoration reset, Incarnate activation/two-option benefit and Apotheosis cost waiver are separate. Draconic splits HP/AC, spell grants, affinity choice/resistance/damage, wings activation/endings/recovery and companion casting/modification. Each level has its own HP/selector path and named granting dependencies; ASI cannot absorb unrelated mechanics. Source records and choices remain unavailable until actual exact outputs qualify.

The [reservation proposal](issue-330-shared-reservation-proposal.md) now names211 concrete shared runtime leaves:204spell parts for98names,3Petrified/Zombie seams and4bounded casting/lifecycle adapters. [All source fixtures](issue-330-shared-contracts.md) and complete candidate bodies have an independent readable oracle. Seven SummonDragon/statblock scopes retain exact166predicates and0HP/spell-ending distinction. The root preserves245source identities and three explicit list/header conflicts; no automatic header eligibility.

Reuse must compare complete old bodies/parents/prerequisites and supported modes. Known false overlaps: Slow494 is weapon mastery; Light1076 is Produce Flame. Polymorph requires698/724/725, not just its admission issue. Light967 extends264; DetectMagic906 extends265. Counterspell356 remains NPC-versus-CureWounds only. Dispel353 concentration ending is not355/965/966dispel execution. 503 is Bard/Cleric/Druid/Wizard membership only; own Sorcerer CHA/list/focus adapter remains separate. Existing Ranger978 has only finite Ranger consumers. No old spell/item/world/condition body or edge is changed.

## Published ownership and qualification

The [native receipt](issue-330-link-verification.json) records the complete **1,888-node graph**: 1,017 prior issues plus **305 Sorcerer, 281 Warlock and 285 Wizard** Tasks. Every new Task is OPEN and Size M. The native identity map is unchanged between the immutable IDS_ALLOCATED and VERIFIED milestones.

All Sorcerer feature and level outputs remain mandatory through class gate **#1475 → #337**. Shared spell, condition, creature and casting outputs also remain mandatory through their finite cohorts and **#1474 → #1477 → #279/#280**. The aggregate #1477 additionally consumes Warlock #1757 and Wizard #2033. The aggregate consumes shared outputs, not the class admission gates. The independent source/native proof checks every frozen body and required edge rather than trusting a publication flag or recorded body digest.

Only five old-root relations were added: each class qualification gate to #337, and the shared aggregate to #279 and #280. Every pre-existing body, title, parent and required edge is retained, including #183. The approved D6/fixed-4 HP adapter consumes #968 → #429; no new direct #429 consumer was added. No children were added under #283. Final native counts are:

| Native constraint | Observed count |
|---|---:|
| #279 incoming prerequisites | 100 |
| #280 incoming prerequisites | 95 |
| #337 incoming prerequisites | 59 |
| #429 outgoing consumers | 100 |
| #285 outgoing consumers | 82 |
| #283 children | 95 |
| Maximum incoming / outgoing / children | 100 / 100 / 95 |

Dependency and containment DAGs pass separately. Closing #330 cannot remove any implementation obligation. Source eligibility remains separate from executable effects, and explicitly unbound forms, spell variants and disputed list admissions remain unavailable.

## Parent supplements and acceptance

ROOT serially applied the four exact sections in [the append package](issue-330-parent-appends.json) to #47/#54/#283/#337 against the independently pinned accepted post-Rogue prefixes. Each complete result was read back exactly. The #47 section is 67 characters, producing a 65,047-character body; the other sections are 463/472/497 characters. The [parent publication receipt](issue-330-parent-publication.json) records the prior, section and result hashes. The read-only verifier independently pins the four published section hashes, permits unrelated later appends, and rejects prefix rollback, altered text, coherent manifest/receipt rewrites and duplicate markers.

The Warlock acceptance prefix is released only after accepted Sorcerer publication; the Wizard prefix follows accepted Warlock publication. Implementation Tasks remain open throughout. No runtime implementation, production grant, coverage promotion or deployment is supplied by these decision closures.

## Validation and evidence limits

The [validation record](issue-330-validation.json) retains the independent source comparators, semantic contrary cases, native evidence and historical recovery receipts. The source guards cover all 103 contextual owners, 211 shared fixtures and their complete bodies, 67 class leaves, 27 qualification nodes, typed transition witnesses and all three printed-list/header conflicts. The corrected Innate Sorcery witness uses CHA modifier 3 and PB 3: DC 14 → 15.

The original full `GOMAXPROCS=2 go test -p 2 -short -race ./...` passed, with the SRD package taking 209.746 seconds. The original `GOMAXPROCS=2 go vet -p 2 ./...` passed in 2.728 seconds. These full runs were not repeated for this evidence-only follow-up. Source/Python checks and final hosted CI qualify the changed artifacts. No fresh integration-tagged PostgreSQL or browser/device journey is claimed. The previously unavailable Go 1.27.3 golangci-lint binary and Ruff configuration were not installed or changed.

The [fresh sealed proof](issue-330-sealed-validation.json) passed across 1,888 nodes using 63 bounded read-only calls. All nine pinned comparators and source-conflict reconciliation passed. The focused suite passed 32 methods; eight coherent parent-text/marker rewrites reproduced acceptance under the prior verifier and were rejected by the repair. Final acceptance still requires independent review and green hosted checks on the final PR head. Historical allocation/recovery records remain dated evidence; they are superseded for current publication status by the immutable VERIFIED delivery.

Publication compatibility: old183body/parent/prerequisites remain EXACTLY unchanged. The world-cohort is an additional mandatory input of own shared-output, so its obligations remain on aggregate/337paths. The literal source fixtures, owner boundaries and305count match root archive20261116T211457Z; only this routing and corresponding body footer text changed. See `issue-330-root-review-delta.json`.
