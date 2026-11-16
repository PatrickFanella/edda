# Sorcerer / Metamagic / Draconic implementation decision (#330)

**PENDING checkpoint.** Source/class audit and offline guards are available for review. Native shared publication, full shared execution semantics, final root POST-ROGUE seals and acceptance are incomplete. This draft cannot satisfy #330 closure or any implementation/FullRules qualification. The PR uses `Closes #330` only so a later accepted merge closes the decision; keep it draft until independent root acceptance.

The prepared clean branch `feat/sorcerer-source-decision` starts at required main `e7cd1ba8318ba296edd4fc5b6b84715ca11016f2`. Physical checkout `/home/shared/Work/Worktrees/edda/t3-1cd94a74-1dc9a0e2b4b7efd2`; commonGitdir `/home/shared/Work/Subcult/Games/edda/.git`. GitHub origin is `https://github.com/subculture-collective/edda.git`; Gitea push URL is disabled. Identity readback: commit `SUBCULT Agent <info@subcult.tv>`, GitHub `PatrickFanella`. No checkout/history replacement was needed.

Pinned English SRD5.2.1 PDF SHA256 `8974902d109d6e63672d7c490bde9ccf052410503d9cfa768237154fbc5e3d87`; directly extracted class64-66, Metamagic66-67, spell lists67-69, Draconic69-70, advancement/multiclass23-26, casting104-106, Summon Dragon/Draconic Spirit166 and future Tome73-74. The independent source comparators check all source operands against that pin. Source-derived records retain the CC-BY-4.0 attribution/adaptation in [the source register](../srd-5.2.1-conformance.md). No PDF or private lore is committed.

Live #330 is open, parent283, blocked by closed285/295/305/317. Their closures imported source metadata only. Counts computed from canonical JSON: **63+22+18=103 records**, **278 contextual mechanic/option occurrences**. The [JSON matrix](issue-330-sorcerer-scope.json) retains every record field and exact context, source-key ownership and findings across rules/game/engine/combat/handlers/database/web/TUI. Shared spell references in class records own grant/adaptation context only; exact spell execution owners still need the separate complete semantic audit. A class preparation adapter never implements a referenced spell.

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

The66 class contracts include46 distinct feature/adapter leaves and20 application leaves. Each contains source predicates, deterministic positive and independently violated negative controls, exact costs/actions/target/timing and isolated entitlement boundaries. Every contract repeats prepare/confirm/retry/reload/concurrency/rollback/isolatedPostgreSQL and shared web+TUI availability requirements. Source-predicate execution is separate from feature behavior, creation/proficiencies/equipment/choices, HP/later-entry, subclass, improvements, resource/slots/preparation and rest recovery.

Font splits pool, slot→points, points→slot and created-slot ending. Innate splits activation, benefits and LongRest recovery. Metamagic splits selection, combination protocol and all ten actual2024 options. Restoration reset, Incarnate activation/two-option benefit and Apotheosis cost waiver are separate. Draconic splits HP/AC, spell grants, affinity choice/resistance/damage, wings activation/endings/recovery and companion casting/modification. Each level has its own HP/selector path and named granting dependencies; ASI cannot absorb unrelated mechanics. Source records and choices remain unavailable until actual exact outputs qualify.

The [reservation proposal](issue-330-shared-reservation-proposal.md) names98 spell gaps with199 proposed bounded shared leaves. That roster is **not yet a complete execution contract set**. It includes seven SummonDragon/statblock owners: manifest/scaling; initiative/verbal commands/default Dodge; defenses/senses; shared resistance; Rend/Multiattack; Breath Weapon; zeroHP/spell ending. No185/274planning closure can execute any of them. The actual source says level5slot scaling, AC14+spelllevel, HP50+10perlevelabove5, PB=yourPB, exact senses/defenses, halfspelllevel roundeddown Rend attacks plus Breath, and disappearance at0HP/spellend. These named scopes still require their independent complete candidate oracle before publication.

Reuse must compare complete old bodies/parents/prerequisites and supported modes. Known false overlaps: Slow494 is weapon mastery; Light1076 is Produce Flame. Polymorph requires698/724/725, not just its admission issue. Light967 extends264; DetectMagic906 extends265. Counterspell356 remains NPC-versus-CureWounds only. Dispel353 concentration ending is not355/965/966dispel execution. 503 is Bard/Cleric/Druid/Wizard membership only; own Sorcerer CHA/list/focus adapter remains separate. Existing Ranger978 has only finite Ranger consumers. No old spell/item/world/condition body or edge is changed.

## Publication boundary and root handoff

Root reserved exactly one279/280relation for one qualification-only aggregate with three class-SHARED outputs330/331/332. It is **not approved**. Proposed structure projects279100/28095, retains283children95 and429outgoing100, and separates dependency/containment DAGs. The provisional282node projection leaves330at70children; it excludes incomplete shared execution prerequisites and is not final cap proof. Future-source audit finds142Sorcerer/117Warlock/217Wizard consumed identities, union245; Tome includes27cantrips and11level1Rituals. Future output prerequisites must expand only through finite source-bound, root-accepted contracts, never an empty placeholder or arbitrary mutation. Full details are in the proposal.

Native child publication awaits completed shared contracts, concrete approval and final dependency/cap proof. Four compact [PENDING root-only append artifacts](issue-330-parent-appends.json) are not live publication. Final immutable seals await root's accepted POST-ROGUE release after329merge/bookkeeping. No watch, merge, closure, runtime implementation, production mutation or grants belong to this child.

## Validation and remaining work

`python3 scripts/verify_sorcerer_scope.py --allow-pending` passes the source/class checkpoint. Default invocation deliberately rejects PENDING; even a coherently forged PUBLISHED flag/receipt is rejected until the final shared/native semantic oracle exists. Ten unittest methods exercise all103 coherent owner reroutes,66 actual fixture/body mutations,20 granting-path deletions, strict numeric types, full-prefix/append controls, all3caps/separateDAGs and bounded exact-query pagination. These checks do not establish the missing final oracle.

Nine pinned comparators passed: Sorcerer, Metamagic, Draconic, advancement, multiclass and four feat catalogues. One requested `GOMAXPROCS=2 go test -p 2 -short -race ./...` passed; srd521package209.746s, overall walltime was not instrumented. One requested `GOMAXPROCS=2 go vet -p 2 ./...` passed in2.728s. No timeouts/assertions or caches changed. This suite does not run integration-tagged PostgreSQL/browser/device evidence. Go1.27.3 golangci-lint shim has no binary and Ruff's mise shim has no configured version; no tool installation or host config changes are authorized here.

Remaining: whole-body exact-mode overlap audit;199 shared source-semantic contracts and independent actual-candidate oracle; finite future-output/cohort contracts; complete protocol prerequisites and cap/DAG projections; concrete parent approval; native Task/ProjectSize/body/title/parent/blocked-by readback; root serial appends and accepted POST-ROGUE seal; final strict live proof and one final-head hosted snapshot. The draft is not ready for acceptance.
