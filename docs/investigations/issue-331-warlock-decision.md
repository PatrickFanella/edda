# Warlock source and implementation decision — publication pending

The decision draft reconciles every canonical Warlock, Fiend and Invocation record with the pinned source and current consumers. It proposes future implementation and qualification Tasks. It delivers no runtime effects, class grants, Warlock admission, coverage or FullRules promotion. ROOT approved the exact finite native reservation; Sorcerer alone publishes the combined Tasks. Actual native readback and ROOT’s accepted POST-SORCERER prefix/appends remain pending. See [the independent publication proof](issue-331-publication-proof.md).

The source count is computed:71 Warlock +17 Fiend +41/49/13 Invocation batches =191 records,588 contextual mechanic/option mappings and28 Invocation identities. The direct printed spell list contains72 identities. Fiend, named Invocations and Tome’s27cantrips/11level1Rituals produce117 consumed spell identities. The familiar domain is24pinnedCR0Beasts plus7Chainforms; CreateUndead has4separate named forms.

Required prepared base: `e7cd1ba8318ba296edd4fc5b6b84715ca11016f2`. GitHub origin is authoritative; Gitea push is disabled. SRD5.2.1 SHA256: `8974902d109d6e63672d7c490bde9ccf052410503d9cfa768237154fbc5e3d87`. CC-BY4 attribution and adaptations are retained in `docs/srd-5.2.1-conformance.md` and every source envelope.

## All20 source rows

|Warlock level|Printed single-class PB|Invocations|Cantrips|Prepared|Pact slots|Slot level|Named features|
|---|---|---|---|---|---|---|---|
|1|2|1|2|2|1|1|eldritch-invocations, pact-magic|
|2|2|3|2|3|2|1|magical-cunning|
|3|2|3|2|4|2|2|warlock-subclass|
|4|2|3|3|5|2|2|ability-score-improvement|
|5|3|5|3|6|2|3|no new named feature|
|6|3|5|3|7|2|3|srd-5.2.1:mechanic/warlock/subclass-feature|
|7|3|6|3|8|2|4|no new named feature|
|8|3|6|3|9|2|4|ability-score-improvement|
|9|4|7|3|10|2|5|contact-patron|
|10|4|7|4|10|2|5|srd-5.2.1:mechanic/warlock/subclass-feature|
|11|4|7|4|11|3|5|6|mystic-arcanum|
|12|4|8|4|11|3|5|ability-score-improvement|
|13|5|8|4|12|3|5|7|mystic-arcanum|
|14|5|8|4|12|3|5|srd-5.2.1:mechanic/warlock/subclass-feature|
|15|5|9|4|13|3|5|8|mystic-arcanum|
|16|5|9|4|13|3|5|ability-score-improvement|
|17|6|9|4|14|4|5|9|mystic-arcanum|
|18|6|10|4|14|4|5|no new named feature|
|19|6|10|4|15|4|5|epic-boon|
|20|6|10|4|15|4|5|eldritch-master|

XP and live PB use the complete canonical285 table at the authoritative sum of admitted post-transition class levels. The printed class PB is not a multiclass override. Fighter4→Warlock1 uses total5/6500XP/PB3 and later HP; Fighter4/Warlock1→Warlock2 uses total6/14000XP/PB3. All20 applications consume ownHP and exact source grants. First-character creation is level1 only. Ordinary first-class advancement does not gain an invented multiclass ability minimum.

## Actual consumers

- **rules** — `internal/rules/srd521/creation.go`, `SupportedBuildForConcept`: Fixed Fighter/Cleric builds only; Warlock is not admitted.
- **rules** — `internal/rules/srd521/equipment.go`, `ChangeEquipment`: Chain mail/shield and fixed Fighter/Cleric only; no general Leather/Orb/Book/weapon adapter.
- **game** — `internal/game/advancement_preview.go`, `prepareAdvancement`: Fixed Fighter1-to2 preview; no Warlock class vector or adjacent transition.
- **engine** — `internal/engine/action_confirmation.go`, `ConfirmAction`: Authoritative preview/revision gate and narrow spell targets; storage does not supply Warlock effects.
- **combat** — `internal/combat/encounter.go`, `SpendSpell`: One slot TO CAST turn budget; no Pact/Arcanum/Invocation resource ownership.
- **database** — `internal/state/sqlc/player_characters.sql.go`, `CreatePlayerCharacter`: Persists character state; no source grant/effect execution supplied by stored labels or JSON.
- **handlers** — `internal/handlers/advancement_integration_test.go`, `TestIntegrationFighterAdvancement`: Isolated PostgreSQL Fighter precedent only, not Warlock qualification.
- **handlers** — `internal/world/rules_character.go`, `ValidateRulesCharacter`: Guided/interview/persistence share supported-build validation.
- **web** — `frontend/src/components/start/CharacterGuidedForm.tsx`, `Warlock label`: Display label does not admit or execute Warlock.
- **tui** — `tui/campaign/char_form.go`, `NewCharFormModel`: Class labels from world.D5EClasses; no independent Warlock rules executor.
- **rules** — `internal/rules/srd521/catalogue_warlock_test.go`, `TestWarlockSourceAndBoundaries`: Explicit metadata-only and Warlock creation rejection.

The JSON companion preserves all191 complete source records and all588 per-record owner mappings. Contexts are retained when a mechanic key appears in multiple granting paths. Classification/provenance/storage does not execute effects. The independent oracle checks the actual candidate, original pinned comparators and unchanged required-base catalogue envelopes.

## Publication structure

The actual proposed native packet is281 nodes:100class/level/feature qualification contracts,70spell effect parts,70creature/control parts,14named item parts and27qualification nodes. Feature leaves use the bounded feature cohort;20level leaves use the level cohort. New spell/creature leaves use cohorts under277; named item leaves use278. No new283children or429consumers are added. HP depends on968 and the actual774die-window output, not directly429.

One finite own execution output consumes the exact old primitive contracts; three bounded consumer factors keep outgoing counts below100. Condition, casting and provenance outputs have explicit finite consumers. The four casting seams are Sorcerer-owned; their provisional contracts do not supply runtime support. Every shared leaf reaches the own shared output, then the exactly-three-output aggregate for279/280, and the own class gate for337. No shared output consumes a class admission gate.

The complete combined projection uses real305Sorcerer +281Warlock +285Wizard proposed nodes and1017observed native nodes:1888total, no dangling references, both DAGs acyclic. Allthree caps pass: incoming100,outgoing100,children95;279incoming100,280incoming95,337incoming59,283children95,429outgoing100,285outgoing82,67outgoing100. Old183is unchanged; Sorcerer’s world cohort is required by its shared output. This projection certificate is separate from the pinned ROOT approval and the pending actual native publication proof.

Existing full owner bodies/parents/prerequisites remain intact. The exact reuse table excludes substring false matches: Gate is not Wolf size-gating; Light is not ProduceFlame/Daylight/LightningBolt/Blight; Message is not Druidic messages; Resistance is not Wish/Hunter defense. FindFamiliar181is a decision. Dispel consumes355’s narrow fixture plus965/966; Light967andDetect906 retain their own exact modes. Unknown/unqualified variants remain unavailable.

## Class and level contracts

### Implement Warlock core traits (implementation)

Decision #331. One implementation PR; Task Size M.

PDF70first-character ONLY: Charismaprimary(noautomaticabilityincrease),Wisdom/Charismasaveproficiencies,Simpleweaponproficiency,Lightarmortraining;HitDie/HPownadapter,skill/equipmentchoicesseparate. PositivefirstWarlock1bothsaves/simple/light; rejectMartial/Medium/Shieldtraining,automaticCHA13orfirst-classsavesonlater-entry. Consume430wholeprovenance; do notgrantitemeffectsfromproficiency.

Prepare is inert: no dice, cost, grant or writes. Confirm revalidates actor, campaign, source, revision, entitlement, target, action, timing and operands. Persist choice, injected dice, resources, effects, derived sheet and history atomically in isolated PostgreSQL. Same-ID replay/reload/reconnect returns one receipt without reroll or regrant; conflicting payload, foreign state and stale revision reject. Concurrent confirms commit exactly once. Inject failure after cost/effect/history and prove full rollback. Web and TUI consume the same engine availability and explicit unavailable reasons. Unqualified effects stay unavailable; source metadata and closed planning issues grant nothing. SRD5.2.1 CC-BY4 attribution/adaptation is retained via docs/srd-5.2.1-conformance.md.

### Implement Warlock skills (implementation)

Decision #331. One implementation PR; Task Size M.

PDF70first-characterONLYchoose2distinctArcana/Deception/History/Intimidation/Investigation/Nature/Religion. Actualskillmodifiers/PBvia61/58with430sourceprovenance; overlappingothersourcegrantdoesnotdoublePB. PositiveArcana+ReligionINT2/PB2=4each; reject3selections,duplicatewithinchoice,unlistedPerception,PBtwiceornewskillsonlaterWarlockentry. Do notinventunprintedreplacementpermissions.

Prepare is inert: no dice, cost, grant or writes. Confirm revalidates actor, campaign, source, revision, entitlement, target, action, timing and operands. Persist choice, injected dice, resources, effects, derived sheet and history atomically in isolated PostgreSQL. Same-ID replay/reload/reconnect returns one receipt without reroll or regrant; conflicting payload, foreign state and stale revision reject. Concurrent confirms commit exactly once. Inject failure after cost/effect/history and prove full rollback. Web and TUI consume the same engine availability and explicit unavailable reasons. Unqualified effects stay unavailable; source metadata and closed planning issues grant nothing. SRD5.2.1 CC-BY4 attribution/adaptation is retained via docs/srd-5.2.1-conformance.md.

### Implement Warlock equipment (implementation)

Decision #331. One implementation PR; Task Size M.

PDF70initialA grantsLeatherArmor,Sickle,2Daggers,OrbArcaneFocus,occultBook,ScholarsPack,15GP; B grants100GPonly; chooseexactlyone,creationonly. ActualScholarPack99 addsBackpack,SECOND Book,Ink,InkPen,Lamp,10Oil,10Parchment,Tinderbox. PositiveA2Daggers/2Books/10Oil/15GP; reject1Booktotal,3Daggers,mixedA+B,purchasechargesatcreationorgrantingpackagainonlaterentry. Consumeactualnameditemoutputs; storage/identityalone doesnotexecutearmor/weapon/book/focus/light effects.

Prepare is inert: no dice, cost, grant or writes. Confirm revalidates actor, campaign, source, revision, entitlement, target, action, timing and operands. Persist choice, injected dice, resources, effects, derived sheet and history atomically in isolated PostgreSQL. Same-ID replay/reload/reconnect returns one receipt without reroll or regrant; conflicting payload, foreign state and stale revision reject. Concurrent confirms commit exactly once. Inject failure after cost/effect/history and prove full rollback. Web and TUI consume the same engine availability and explicit unavailable reasons. Unqualified effects stay unavailable; source metadata and closed planning issues grant nothing. SRD5.2.1 CC-BY4 attribution/adaptation is retained via docs/srd-5.2.1-conformance.md.

### Implement Warlock warlock list (implementation)

Decision #331. One implementation PR; Task Size M.

PDF74–76/105: exact72 printed Warlock spell identities with level/school/Ritual/Concentration/material flags; prepared Pact selection levels1–5 and known cantrip selection level0 use own list, not503. Mystic Arcanum levels6–9 separate; named Fiend/Invocation/Tome routes do not silently add all117 to ordinary list. PositiveMindSpike listed75/header149 eligibleWarlock2; rejectFireball ordinaryWarlocklist, SpeakwithDead ordinarylist, disputedSorcererMindSpike grant or header-onlyoverride. Unknown/unqualified effects unavailable even when source eligibility is true.

Prepare is inert: no dice, cost, grant or writes. Confirm revalidates actor, campaign, source, revision, entitlement, target, action, timing and operands. Persist choice, injected dice, resources, effects, derived sheet and history atomically in isolated PostgreSQL. Same-ID replay/reload/reconnect returns one receipt without reroll or regrant; conflicting payload, foreign state and stale revision reject. Concurrent confirms commit exactly once. Inject failure after cost/effect/history and prove full rollback. Web and TUI consume the same engine availability and explicit unavailable reasons. Unqualified effects stay unavailable; source metadata and closed planning issues grant nothing. SRD5.2.1 CC-BY4 attribution/adaptation is retained via docs/srd-5.2.1-conformance.md.

### Implement Warlock pact cross casting (implementation)

Decision #331. One implementation PR; Task Size M.

PDF25–26: actual PactMagic AND actualSpellcasting required for cross-casting. Pactslots may cast prepared Spellcasting-class spells; actualSpellcasting slots may cast preparedWarlockspells, normal slot>=spelllevel and sourcecastingability/provenance. Pactlevels excluded sharedSpellcasting-slot sum; Arcanum resources remain separate. Positive admittedWizard1/Warlock1 preparedMagicMissile usesPactslot1; preparedHex usesWizard1slot withCHA; rejectMagicInitiate-only feature, unpreparedArcanumasordinaryprepared, addingWarlocklevel tosharedsum or counting other slot resource spend as slotTOCAST.

Prepare is inert: no dice, cost, grant or writes. Confirm revalidates actor, campaign, source, revision, entitlement, target, action, timing and operands. Persist choice, injected dice, resources, effects, derived sheet and history atomically in isolated PostgreSQL. Same-ID replay/reload/reconnect returns one receipt without reroll or regrant; conflicting payload, foreign state and stale revision reject. Concurrent confirms commit exactly once. Inject failure after cost/effect/history and prove full rollback. Web and TUI consume the same engine availability and explicit unavailable reasons. Unqualified effects stay unavailable; source metadata and closed planning issues grant nothing. SRD5.2.1 CC-BY4 attribution/adaptation is retained via docs/srd-5.2.1-conformance.md.

### Implement Warlock hp (implementation)

Decision #331. One implementation PR; Task Size M.

PDF23/70: own D8/fixed5 HP adapter consumes968, never direct429. First-character level1 gets8+CON. Later entry and every adjacent row add max(1,D8+CON) or max(1,5+CON) and one D8 Hit Die; no healing or reroll on replay. Positive CON2 fixed later gain7; reject D6, fixed4, first-character8 on later entry and duplicate gain.

Prepare is inert: no dice, cost, grant or writes. Confirm revalidates actor, campaign, source, revision, entitlement, target, action, timing and operands. Persist choice, injected dice, resources, effects, derived sheet and history atomically in isolated PostgreSQL. Same-ID replay/reload/reconnect returns one receipt without reroll or regrant; conflicting payload, foreign state and stale revision reject. Concurrent confirms commit exactly once. Inject failure after cost/effect/history and prove full rollback. Web and TUI consume the same engine availability and explicit unavailable reasons. Unqualified effects stay unavailable; source metadata and closed planning issues grant nothing. SRD5.2.1 CC-BY4 attribution/adaptation is retained via docs/srd-5.2.1-conformance.md.

### Implement Warlock creation (implementation)

Decision #331. One implementation PR; Task Size M.

PDF70: first-character creation only at level1, Charisma primary, Wisdom/Charisma saves, simple weapons, light armor; choose2 distinct Arcana/Deception/History/Intimidation/Investigation/Nature/Religion. Equipment A: Leather Armor,Sickle,2 Daggers,Arcane Focus orb,occult Book,Scholars Pack,15GP; B:100GP. Positive A exactly2 Daggers/15GP; reject3 skills, duplicate skill, mixed A+B, medium armor. Inventory identity does not execute item effects.

Prepare is inert: no dice, cost, grant or writes. Confirm revalidates actor, campaign, source, revision, entitlement, target, action, timing and operands. Persist choice, injected dice, resources, effects, derived sheet and history atomically in isolated PostgreSQL. Same-ID replay/reload/reconnect returns one receipt without reroll or regrant; conflicting payload, foreign state and stale revision reject. Concurrent confirms commit exactly once. Inject failure after cost/effect/history and prove full rollback. Web and TUI consume the same engine availability and explicit unavailable reasons. Unqualified effects stay unavailable; source metadata and closed planning issues grant nothing. SRD5.2.1 CC-BY4 attribution/adaptation is retained via docs/srd-5.2.1-conformance.md.

### Implement Warlock later entry (implementation)

Decision #331. One implementation PR; Task Size M.

PDF24/70: add Warlock1 only to admitted existing class combination with Charisma13 and each existing class source minimum; grant D8 and light armor plus level1 features, no saves/weapons/skills/equipment. Fighter4->Warlock1 total5 uses6500XP/PB3 and later HP. Reject CHA12, boolean13, unqualified prior class, total>20 or first-character HP/equipment.

Prepare is inert: no dice, cost, grant or writes. Confirm revalidates actor, campaign, source, revision, entitlement, target, action, timing and operands. Persist choice, injected dice, resources, effects, derived sheet and history atomically in isolated PostgreSQL. Same-ID replay/reload/reconnect returns one receipt without reroll or regrant; conflicting payload, foreign state and stale revision reject. Concurrent confirms commit exactly once. Inject failure after cost/effect/history and prove full rollback. Web and TUI consume the same engine availability and explicit unavailable reasons. Unqualified effects stay unavailable; source metadata and closed planning issues grant nothing. SRD5.2.1 CC-BY4 attribution/adaptation is retained via docs/srd-5.2.1-conformance.md.

### Implement Warlock eldritch invocations (implementation)

Decision #331. One implementation PR; Task Size M.

PDF70–71: exact per-Warlock-level counts; level1 choose1; level2 total3; level5 total5. Learn only eligible options. Once per gained Warlock level replace at most1; cannot replace an Invocation required by another owned Invocation. Nonrepeatable default; only Agonizing Blast/Eldritch Spear/Repelling Blast/Lessons repeat with different eligible selection. Positive remove independent option; reject removal Pact Blade while Thirsting Blade retained, duplicate nonrepeatable, same repeatable cantrip and2 replacements.

Prepare is inert: no dice, cost, grant or writes. Confirm revalidates actor, campaign, source, revision, entitlement, target, action, timing and operands. Persist choice, injected dice, resources, effects, derived sheet and history atomically in isolated PostgreSQL. Same-ID replay/reload/reconnect returns one receipt without reroll or regrant; conflicting payload, foreign state and stale revision reject. Concurrent confirms commit exactly once. Inject failure after cost/effect/history and prove full rollback. Web and TUI consume the same engine availability and explicit unavailable reasons. Unqualified effects stay unavailable; source metadata and closed planning issues grant nothing. SRD5.2.1 CC-BY4 attribution/adaptation is retained via docs/srd-5.2.1-conformance.md.

### Implement Warlock pact magic (implementation)

Decision #331. One implementation PR; Task Size M.

PDF71/104–106: Charisma spell attack/DC; two Warlock cantrips, two prepared level1 spells initially; exact list adapter not503. Spell costs one distinct Pact slot of printed class slot level, not generic Spellcasting or Arcanum resource. Positive Warlock5 Charm Person cast at3 using one of2 slots; reject unprepared, wrong list, level4 spell atWarlock5, generic slot substitution without actual qualified multiclass routing. One SLOT TO CAST per turn; other slot resource expenditures are not casts. Confirm private invalid target spends confirmed slot/no effect without private leakage; unverifiable facts unavailable.

Prepare is inert: no dice, cost, grant or writes. Confirm revalidates actor, campaign, source, revision, entitlement, target, action, timing and operands. Persist choice, injected dice, resources, effects, derived sheet and history atomically in isolated PostgreSQL. Same-ID replay/reload/reconnect returns one receipt without reroll or regrant; conflicting payload, foreign state and stale revision reject. Concurrent confirms commit exactly once. Inject failure after cost/effect/history and prove full rollback. Web and TUI consume the same engine availability and explicit unavailable reasons. Unqualified effects stay unavailable; source metadata and closed planning issues grant nothing. SRD5.2.1 CC-BY4 attribution/adaptation is retained via docs/srd-5.2.1-conformance.md.

### Implement Warlock preparation (implementation)

Decision #331. One implementation PR; Task Size M.

PDF71–72: on each Warlock level replace at most1 class cantrip and1 prepared spell; additions exactly printed count deltas, chosen level no higher than own slot level. Always-prepared feature spells excluded from class prepared count; retain distinct source provenance. Positive level4 third cantrip/5 prepared; reject rest-based wholesale replacement, overcapacity, invalid level, removing another source grant.

Prepare is inert: no dice, cost, grant or writes. Confirm revalidates actor, campaign, source, revision, entitlement, target, action, timing and operands. Persist choice, injected dice, resources, effects, derived sheet and history atomically in isolated PostgreSQL. Same-ID replay/reload/reconnect returns one receipt without reroll or regrant; conflicting payload, foreign state and stale revision reject. Concurrent confirms commit exactly once. Inject failure after cost/effect/history and prove full rollback. Web and TUI consume the same engine availability and explicit unavailable reasons. Unqualified effects stay unavailable; source metadata and closed planning issues grant nothing. SRD5.2.1 CC-BY4 attribution/adaptation is retained via docs/srd-5.2.1-conformance.md.

### Implement Warlock slots (implementation)

Decision #331. One implementation PR; Task Size M.

PDF71/427/428: distinct Pact pool recovers all expended slots after completed Short or Long Rest. Capacity by own Warlock level1=1@1,2=2@1,11=3@5,17=4@5. Positive spent2/2->2 after rest; reject overcapacity, Arcanum recharge on short rest, interruption and duplicate receipt recharge.

Prepare is inert: no dice, cost, grant or writes. Confirm revalidates actor, campaign, source, revision, entitlement, target, action, timing and operands. Persist choice, injected dice, resources, effects, derived sheet and history atomically in isolated PostgreSQL. Same-ID replay/reload/reconnect returns one receipt without reroll or regrant; conflicting payload, foreign state and stale revision reject. Concurrent confirms commit exactly once. Inject failure after cost/effect/history and prove full rollback. Web and TUI consume the same engine availability and explicit unavailable reasons. Unqualified effects stay unavailable; source metadata and closed planning issues grant nothing. SRD5.2.1 CC-BY4 attribution/adaptation is retained via docs/srd-5.2.1-conformance.md.

### Implement Warlock magical cunning (implementation)

Decision #331. One implementation PR; Task Size M.

PDF72: perform1-minute rite; at end regain expended Pact slots up to ceil(max/2), once per Long Rest. Positive capacity3 spent3 recovers2, capacity1 recovers1; reject instant recovery, generic pool, second use, floor rounding, interruption/replay restoring again.

Prepare is inert: no dice, cost, grant or writes. Confirm revalidates actor, campaign, source, revision, entitlement, target, action, timing and operands. Persist choice, injected dice, resources, effects, derived sheet and history atomically in isolated PostgreSQL. Same-ID replay/reload/reconnect returns one receipt without reroll or regrant; conflicting payload, foreign state and stale revision reject. Concurrent confirms commit exactly once. Inject failure after cost/effect/history and prove full rollback. Web and TUI consume the same engine availability and explicit unavailable reasons. Unqualified effects stay unavailable; source metadata and closed planning issues grant nothing. SRD5.2.1 CC-BY4 attribution/adaptation is retained via docs/srd-5.2.1-conformance.md.

### Implement Warlock warlock subclass (implementation)

Decision #331. One implementation PR; Task Size M.

PDF72/76: choose one qualified subclass at Warlock3; Fiend grants3/6/10/14 features at own Warlock level. Positive Warlock3 selects Fiend; reject Warlock2, character-total3/Warlock1, second subclass and source-only subclass execution.

Prepare is inert: no dice, cost, grant or writes. Confirm revalidates actor, campaign, source, revision, entitlement, target, action, timing and operands. Persist choice, injected dice, resources, effects, derived sheet and history atomically in isolated PostgreSQL. Same-ID replay/reload/reconnect returns one receipt without reroll or regrant; conflicting payload, foreign state and stale revision reject. Concurrent confirms commit exactly once. Inject failure after cost/effect/history and prove full rollback. Web and TUI consume the same engine availability and explicit unavailable reasons. Unqualified effects stay unavailable; source metadata and closed planning issues grant nothing. SRD5.2.1 CC-BY4 attribution/adaptation is retained via docs/srd-5.2.1-conformance.md.

### Implement Warlock ability score improvement (implementation)

Decision #331. One implementation PR; Task Size M.

PDF72/87: independent feat grant at each Warlock4/8/12/16; choose ASI or another eligible feat, exact385–401 effect owners and384 selection/389 caps. Positive four distinct grant receipts; reject missed level8, double grant at4, absorbing Invocation/Arcanum/other feature. Spell Recall399 requires actual Spellcasting feature, not Pact Magic alone, Magic Initiate or scroll identity.

Prepare is inert: no dice, cost, grant or writes. Confirm revalidates actor, campaign, source, revision, entitlement, target, action, timing and operands. Persist choice, injected dice, resources, effects, derived sheet and history atomically in isolated PostgreSQL. Same-ID replay/reload/reconnect returns one receipt without reroll or regrant; conflicting payload, foreign state and stale revision reject. Concurrent confirms commit exactly once. Inject failure after cost/effect/history and prove full rollback. Web and TUI consume the same engine availability and explicit unavailable reasons. Unqualified effects stay unavailable; source metadata and closed planning issues grant nothing. SRD5.2.1 CC-BY4 attribution/adaptation is retained via docs/srd-5.2.1-conformance.md.

### Implement Warlock contact patron (implementation)

Decision #331. One implementation PR; Task Size M.

PDF72: at Warlock9 always prepare Contact Other Plane; special cast contacts patron, no slot, automatic save success, once per Long Rest; consume real effect owner. Positive eligible special cast with0 Pact slots; reject second free use, wrong contact recipient, generic spell-wide auto-success or lost always-prepared provenance.

Prepare is inert: no dice, cost, grant or writes. Confirm revalidates actor, campaign, source, revision, entitlement, target, action, timing and operands. Persist choice, injected dice, resources, effects, derived sheet and history atomically in isolated PostgreSQL. Same-ID replay/reload/reconnect returns one receipt without reroll or regrant; conflicting payload, foreign state and stale revision reject. Concurrent confirms commit exactly once. Inject failure after cost/effect/history and prove full rollback. Web and TUI consume the same engine availability and explicit unavailable reasons. Unqualified effects stay unavailable; source metadata and closed planning issues grant nothing. SRD5.2.1 CC-BY4 attribution/adaptation is retained via docs/srd-5.2.1-conformance.md.

### Qualify Warlock mystic arcanum (qualification)

Decision #331. One implementation PR; Task Size M.

PDF72: separate one-use resource for chosen level6/7/8/9 Warlock spell at own11/13/15/17; slot-free cast once per resource, all recover Long Rest. On every Warlock level replace at most1 arcanum with same-level Warlock spell. Positive own17 retains four independent resources; reject shared one-use counter, Pact expenditure, short rest recovery, wrong-level replacement,2 replacements.

Prepare is inert: no dice, cost, grant or writes. Confirm revalidates actor, campaign, source, revision, entitlement, target, action, timing and operands. Persist choice, injected dice, resources, effects, derived sheet and history atomically in isolated PostgreSQL. Same-ID replay/reload/reconnect returns one receipt without reroll or regrant; conflicting payload, foreign state and stale revision reject. Concurrent confirms commit exactly once. Inject failure after cost/effect/history and prove full rollback. Web and TUI consume the same engine availability and explicit unavailable reasons. Unqualified effects stay unavailable; source metadata and closed planning issues grant nothing. SRD5.2.1 CC-BY4 attribution/adaptation is retained via docs/srd-5.2.1-conformance.md.

### Implement Warlock epic boon (implementation)

Decision #331. One implementation PR; Task Size M.

PDF72/87–88: own Warlock19 grants one Epic Boon or other eligible feat; Boon of Fate recommendation is not mandatory. Positive admitted eligible feat; reject level18, duplicate receipt, ineligible Spell Recall via PactMagic name, replacing Invocation or ASI receipt.

Prepare is inert: no dice, cost, grant or writes. Confirm revalidates actor, campaign, source, revision, entitlement, target, action, timing and operands. Persist choice, injected dice, resources, effects, derived sheet and history atomically in isolated PostgreSQL. Same-ID replay/reload/reconnect returns one receipt without reroll or regrant; conflicting payload, foreign state and stale revision reject. Concurrent confirms commit exactly once. Inject failure after cost/effect/history and prove full rollback. Web and TUI consume the same engine availability and explicit unavailable reasons. Unqualified effects stay unavailable; source metadata and closed planning issues grant nothing. SRD5.2.1 CC-BY4 attribution/adaptation is retained via docs/srd-5.2.1-conformance.md.

### Implement Warlock eldritch master (implementation)

Decision #331. One implementation PR; Task Size M.

PDF72: own Warlock20 changes Magical Cunning recovery to all expended Pact slots; retain1-minute rite/once Long Rest. Positive4 spent restores4; reject regain2, generic slots, twice per rest or instant restore.

Prepare is inert: no dice, cost, grant or writes. Confirm revalidates actor, campaign, source, revision, entitlement, target, action, timing and operands. Persist choice, injected dice, resources, effects, derived sheet and history atomically in isolated PostgreSQL. Same-ID replay/reload/reconnect returns one receipt without reroll or regrant; conflicting payload, foreign state and stale revision reject. Concurrent confirms commit exactly once. Inject failure after cost/effect/history and prove full rollback. Web and TUI consume the same engine availability and explicit unavailable reasons. Unqualified effects stay unavailable; source metadata and closed planning issues grant nothing. SRD5.2.1 CC-BY4 attribution/adaptation is retained via docs/srd-5.2.1-conformance.md.

### Implement Warlock dark ones blessing (implementation)

Decision #331. One implementation PR; Task Size M.

PDF76: when you reduce enemy to0HP, or someone else reduces enemy within10ft of you to0HP, gain max(1,CHAmod+WarlockLevel) Temporary HP. Positive Warlock3 CHA2 yields5, other killer10ft eligible; reject11ft other killer, nonenemy, heal5, stacking Temporary HP.

Prepare is inert: no dice, cost, grant or writes. Confirm revalidates actor, campaign, source, revision, entitlement, target, action, timing and operands. Persist choice, injected dice, resources, effects, derived sheet and history atomically in isolated PostgreSQL. Same-ID replay/reload/reconnect returns one receipt without reroll or regrant; conflicting payload, foreign state and stale revision reject. Concurrent confirms commit exactly once. Inject failure after cost/effect/history and prove full rollback. Web and TUI consume the same engine availability and explicit unavailable reasons. Unqualified effects stay unavailable; source metadata and closed planning issues grant nothing. SRD5.2.1 CC-BY4 attribution/adaptation is retained via docs/srd-5.2.1-conformance.md.

### Implement Warlock fiend spells (implementation)

Decision #331. One implementation PR; Task Size M.

PDF76: own3 Burning Hands/Command/Scorching Ray/Suggestion;5 Fireball/Stinking Cloud;7 Fire Shield/Wall of Fire;9 Geas/Insect Plague always prepared without consuming class preparation capacity. Positive own9 ten spells retained; reject grant at character-total9/Warlock3, missing earlier rows, capacity subtraction, names as execution.

Prepare is inert: no dice, cost, grant or writes. Confirm revalidates actor, campaign, source, revision, entitlement, target, action, timing and operands. Persist choice, injected dice, resources, effects, derived sheet and history atomically in isolated PostgreSQL. Same-ID replay/reload/reconnect returns one receipt without reroll or regrant; conflicting payload, foreign state and stale revision reject. Concurrent confirms commit exactly once. Inject failure after cost/effect/history and prove full rollback. Web and TUI consume the same engine availability and explicit unavailable reasons. Unqualified effects stay unavailable; source metadata and closed planning issues grant nothing. SRD5.2.1 CC-BY4 attribution/adaptation is retained via docs/srd-5.2.1-conformance.md.

### Implement Warlock dark ones own luck (implementation)

Decision #331. One implementation PR; Task Size M.

PDF76: add1d10 to own ability check or save after seeing roll before effects; max1 per roll; max(1,CHAmod) uses per Long Rest. Positive CHA3 permits3 distinct rolls; reject attack roll, after effects, fourth use or two uses same roll. Undisclosed outcome alone does not forbid outcome-independent known-die windows.

Prepare is inert: no dice, cost, grant or writes. Confirm revalidates actor, campaign, source, revision, entitlement, target, action, timing and operands. Persist choice, injected dice, resources, effects, derived sheet and history atomically in isolated PostgreSQL. Same-ID replay/reload/reconnect returns one receipt without reroll or regrant; conflicting payload, foreign state and stale revision reject. Concurrent confirms commit exactly once. Inject failure after cost/effect/history and prove full rollback. Web and TUI consume the same engine availability and explicit unavailable reasons. Unqualified effects stay unavailable; source metadata and closed planning issues grant nothing. SRD5.2.1 CC-BY4 attribution/adaptation is retained via docs/srd-5.2.1-conformance.md.

### Implement Warlock fiendish resilience (implementation)

Decision #331. One implementation PR; Task Size M.

PDF76: at completed Short/Long Rest choose1 damage type except Force; Resistance until choose different. Positive Fire->Cold replaces; reject Force,2 types, arbitrary anytime change, expired on next rest without replacement.

Prepare is inert: no dice, cost, grant or writes. Confirm revalidates actor, campaign, source, revision, entitlement, target, action, timing and operands. Persist choice, injected dice, resources, effects, derived sheet and history atomically in isolated PostgreSQL. Same-ID replay/reload/reconnect returns one receipt without reroll or regrant; conflicting payload, foreign state and stale revision reject. Concurrent confirms commit exactly once. Inject failure after cost/effect/history and prove full rollback. Web and TUI consume the same engine availability and explicit unavailable reasons. Unqualified effects stay unavailable; source metadata and closed planning issues grant nothing. SRD5.2.1 CC-BY4 attribution/adaptation is retained via docs/srd-5.2.1-conformance.md.

### Qualify Warlock hurl through hell (qualification)

Decision #331. One implementation PR; Task Size M.

PDF76: once per turn on attack-roll hit creature, CHA save vs own spellDC; failure transports,8d10 Psychic unless Fiend, Incapacitated until end of your next turn; return prior or nearest unoccupied space. Once per Long Rest; no-action Pact slot restores use. Positive failed nonFiend takes8d10; Fiend still transports but0 Psychic. Reject hitless trigger, second same turn, generic slot restore, success transport,8d10 to Fiend. Split world transit/damage/487 lifecycle before publication.

Prepare is inert: no dice, cost, grant or writes. Confirm revalidates actor, campaign, source, revision, entitlement, target, action, timing and operands. Persist choice, injected dice, resources, effects, derived sheet and history atomically in isolated PostgreSQL. Same-ID replay/reload/reconnect returns one receipt without reroll or regrant; conflicting payload, foreign state and stale revision reject. Concurrent confirms commit exactly once. Inject failure after cost/effect/history and prove full rollback. Web and TUI consume the same engine availability and explicit unavailable reasons. Unqualified effects stay unavailable; source metadata and closed planning issues grant nothing. SRD5.2.1 CC-BY4 attribution/adaptation is retained via docs/srd-5.2.1-conformance.md.

### Implement Warlock agonizing blast (implementation)

Decision #331. One implementation PR; Task Size M.

PDF72: own2+, choose known damaging Warlock cantrip, add CHAmod to its damage rolls; repeat only different eligible cantrip. Positive CHA3 adds3 per damage roll; reject unknown/nonWarlock/non-damaging, duplicate selection or only once per cast.

Prepare is inert: no dice, cost, grant or writes. Confirm revalidates actor, campaign, source, revision, entitlement, target, action, timing and operands. Persist choice, injected dice, resources, effects, derived sheet and history atomically in isolated PostgreSQL. Same-ID replay/reload/reconnect returns one receipt without reroll or regrant; conflicting payload, foreign state and stale revision reject. Concurrent confirms commit exactly once. Inject failure after cost/effect/history and prove full rollback. Web and TUI consume the same engine availability and explicit unavailable reasons. Unqualified effects stay unavailable; source metadata and closed planning issues grant nothing. SRD5.2.1 CC-BY4 attribution/adaptation is retained via docs/srd-5.2.1-conformance.md.

### Implement Warlock armor of shadows (implementation)

Decision #331. One implementation PR; Task Size M.

PDF72: cast Mage Armor on self without slot. Positive0slots self; reject other target, copied effect or inferred preparation requirement.

Prepare is inert: no dice, cost, grant or writes. Confirm revalidates actor, campaign, source, revision, entitlement, target, action, timing and operands. Persist choice, injected dice, resources, effects, derived sheet and history atomically in isolated PostgreSQL. Same-ID replay/reload/reconnect returns one receipt without reroll or regrant; conflicting payload, foreign state and stale revision reject. Concurrent confirms commit exactly once. Inject failure after cost/effect/history and prove full rollback. Web and TUI consume the same engine availability and explicit unavailable reasons. Unqualified effects stay unavailable; source metadata and closed planning issues grant nothing. SRD5.2.1 CC-BY4 attribution/adaptation is retained via docs/srd-5.2.1-conformance.md.

### Implement Warlock ascendant step (implementation)

Decision #331. One implementation PR; Task Size M.

PDF72: own5+, cast Levitate on self without slot. Positive own5/0slots; reject own4 or other target; consume exact spell effect.

Prepare is inert: no dice, cost, grant or writes. Confirm revalidates actor, campaign, source, revision, entitlement, target, action, timing and operands. Persist choice, injected dice, resources, effects, derived sheet and history atomically in isolated PostgreSQL. Same-ID replay/reload/reconnect returns one receipt without reroll or regrant; conflicting payload, foreign state and stale revision reject. Concurrent confirms commit exactly once. Inject failure after cost/effect/history and prove full rollback. Web and TUI consume the same engine availability and explicit unavailable reasons. Unqualified effects stay unavailable; source metadata and closed planning issues grant nothing. SRD5.2.1 CC-BY4 attribution/adaptation is retained via docs/srd-5.2.1-conformance.md.

### Implement Warlock devils sight (implementation)

Decision #331. One implementation PR; Task Size M.

PDF72: own2+, normal vision in magical/nonmagical Dim Light/Darkness within120ft. Positive magical Darkness120; reject121ft, Truesight or seeing through obstruction.

Prepare is inert: no dice, cost, grant or writes. Confirm revalidates actor, campaign, source, revision, entitlement, target, action, timing and operands. Persist choice, injected dice, resources, effects, derived sheet and history atomically in isolated PostgreSQL. Same-ID replay/reload/reconnect returns one receipt without reroll or regrant; conflicting payload, foreign state and stale revision reject. Concurrent confirms commit exactly once. Inject failure after cost/effect/history and prove full rollback. Web and TUI consume the same engine availability and explicit unavailable reasons. Unqualified effects stay unavailable; source metadata and closed planning issues grant nothing. SRD5.2.1 CC-BY4 attribution/adaptation is retained via docs/srd-5.2.1-conformance.md.

### Implement Warlock devouring blade (implementation)

Decision #331. One implementation PR; Task Size M.

PDF72: own12+ and Thirsting Blade; its pact-weapon Extra Attack grants2 extra attacks (3 total). Positive3 pact attacks; reject own11, nonpact weapon,4 attacks or stacking independent ExtraAttack.

Prepare is inert: no dice, cost, grant or writes. Confirm revalidates actor, campaign, source, revision, entitlement, target, action, timing and operands. Persist choice, injected dice, resources, effects, derived sheet and history atomically in isolated PostgreSQL. Same-ID replay/reload/reconnect returns one receipt without reroll or regrant; conflicting payload, foreign state and stale revision reject. Concurrent confirms commit exactly once. Inject failure after cost/effect/history and prove full rollback. Web and TUI consume the same engine availability and explicit unavailable reasons. Unqualified effects stay unavailable; source metadata and closed planning issues grant nothing. SRD5.2.1 CC-BY4 attribution/adaptation is retained via docs/srd-5.2.1-conformance.md.

### Implement Warlock eldritch mind (implementation)

Decision #331. One implementation PR; Task Size M.

PDF72: Advantage on CON saves to maintain Concentration. Positive casting concentration damage save; reject ordinary CON save, other ability, general save advantage.

Prepare is inert: no dice, cost, grant or writes. Confirm revalidates actor, campaign, source, revision, entitlement, target, action, timing and operands. Persist choice, injected dice, resources, effects, derived sheet and history atomically in isolated PostgreSQL. Same-ID replay/reload/reconnect returns one receipt without reroll or regrant; conflicting payload, foreign state and stale revision reject. Concurrent confirms commit exactly once. Inject failure after cost/effect/history and prove full rollback. Web and TUI consume the same engine availability and explicit unavailable reasons. Unqualified effects stay unavailable; source metadata and closed planning issues grant nothing. SRD5.2.1 CC-BY4 attribution/adaptation is retained via docs/srd-5.2.1-conformance.md.

### Implement Warlock eldritch smite (implementation)

Decision #331. One implementation PR; Task Size M.

PDF72–73: own5+/PactBlade; once per turn hit creature with pact weapon, spend1 Pact slot, extra(1+slotLevel)d8 Force; optional Prone if Huge or smaller via488. Positive slot3=4d8 Huge Prone; reject Gargantuan Prone, miss, generic slot, second same turn. Slot resource spend is not slot TO CAST.

Prepare is inert: no dice, cost, grant or writes. Confirm revalidates actor, campaign, source, revision, entitlement, target, action, timing and operands. Persist choice, injected dice, resources, effects, derived sheet and history atomically in isolated PostgreSQL. Same-ID replay/reload/reconnect returns one receipt without reroll or regrant; conflicting payload, foreign state and stale revision reject. Concurrent confirms commit exactly once. Inject failure after cost/effect/history and prove full rollback. Web and TUI consume the same engine availability and explicit unavailable reasons. Unqualified effects stay unavailable; source metadata and closed planning issues grant nothing. SRD5.2.1 CC-BY4 attribution/adaptation is retained via docs/srd-5.2.1-conformance.md.

### Implement Warlock eldritch spear (implementation)

Decision #331. One implementation PR; Task Size M.

PDF73: own2+, known damaging Warlock cantrip printed range10+ft; on cast add30*WarlockLevel ft. Repeat different eligible cantrip. Positive own5 base10->160; reject base5, unknown, total-level scaling or duplicate cantrip.

Prepare is inert: no dice, cost, grant or writes. Confirm revalidates actor, campaign, source, revision, entitlement, target, action, timing and operands. Persist choice, injected dice, resources, effects, derived sheet and history atomically in isolated PostgreSQL. Same-ID replay/reload/reconnect returns one receipt without reroll or regrant; conflicting payload, foreign state and stale revision reject. Concurrent confirms commit exactly once. Inject failure after cost/effect/history and prove full rollback. Web and TUI consume the same engine availability and explicit unavailable reasons. Unqualified effects stay unavailable; source metadata and closed planning issues grant nothing. SRD5.2.1 CC-BY4 attribution/adaptation is retained via docs/srd-5.2.1-conformance.md.

### Implement Warlock fiendish vigor (implementation)

Decision #331. One implementation PR; Task Size M.

PDF73: own2+, cast False Life self without slot, maximum die result instead of roll. Positive base2d4+4=12THP; reject roll, other target, spell slot spending or HP healing. Consume exact effect/scaling owner.

Prepare is inert: no dice, cost, grant or writes. Confirm revalidates actor, campaign, source, revision, entitlement, target, action, timing and operands. Persist choice, injected dice, resources, effects, derived sheet and history atomically in isolated PostgreSQL. Same-ID replay/reload/reconnect returns one receipt without reroll or regrant; conflicting payload, foreign state and stale revision reject. Concurrent confirms commit exactly once. Inject failure after cost/effect/history and prove full rollback. Web and TUI consume the same engine availability and explicit unavailable reasons. Unqualified effects stay unavailable; source metadata and closed planning issues grant nothing. SRD5.2.1 CC-BY4 attribution/adaptation is retained via docs/srd-5.2.1-conformance.md.

### Qualify Warlock gaze of two minds (qualification)

Decision #331. One implementation PR; Task Size M.

PDF73: own5+, Bonus Action touch willing creature; perceive its senses through end next turn; Bonus Action subsequent turns maintains on same plane. Cast from either space only within60ft. Positive60ft cast origin; reject61ft origin, unwilling touch, cross-plane maintenance, missed maintenance. Split perception/maintenance/casting-origin contracts.

Prepare is inert: no dice, cost, grant or writes. Confirm revalidates actor, campaign, source, revision, entitlement, target, action, timing and operands. Persist choice, injected dice, resources, effects, derived sheet and history atomically in isolated PostgreSQL. Same-ID replay/reload/reconnect returns one receipt without reroll or regrant; conflicting payload, foreign state and stale revision reject. Concurrent confirms commit exactly once. Inject failure after cost/effect/history and prove full rollback. Web and TUI consume the same engine availability and explicit unavailable reasons. Unqualified effects stay unavailable; source metadata and closed planning issues grant nothing. SRD5.2.1 CC-BY4 attribution/adaptation is retained via docs/srd-5.2.1-conformance.md.

### Qualify Warlock gift of the depths (qualification)

Decision #331. One implementation PR; Task Size M.

PDF73: own5+, underwater breathing and SwimSpeed=Speed; Water Breathing slot-free once per Long Rest. Positive Speed30->Swim30; reject fixed40, second free cast or short-rest reset; separate passive movement and named effect mode.

Prepare is inert: no dice, cost, grant or writes. Confirm revalidates actor, campaign, source, revision, entitlement, target, action, timing and operands. Persist choice, injected dice, resources, effects, derived sheet and history atomically in isolated PostgreSQL. Same-ID replay/reload/reconnect returns one receipt without reroll or regrant; conflicting payload, foreign state and stale revision reject. Concurrent confirms commit exactly once. Inject failure after cost/effect/history and prove full rollback. Web and TUI consume the same engine availability and explicit unavailable reasons. Unqualified effects stay unavailable; source metadata and closed planning issues grant nothing. SRD5.2.1 CC-BY4 attribution/adaptation is retained via docs/srd-5.2.1-conformance.md.

### Qualify Warlock gift of the protectors (qualification)

Decision #331. One implementation PR; Task Size M.

PDF73: own9+/PactTome; page on conjured book, permission+creature action writes name, max(1,CHAmod) names; named creature reduced0HP not killed outright drops1HP; global once Long Rest, not per name; Magic action touch erases name. Positive CHA2 two names/one total rescue; reject third name, no permission, outright death rescue, second creature rescue before rest. Split inscription/trigger/recovery.

Prepare is inert: no dice, cost, grant or writes. Confirm revalidates actor, campaign, source, revision, entitlement, target, action, timing and operands. Persist choice, injected dice, resources, effects, derived sheet and history atomically in isolated PostgreSQL. Same-ID replay/reload/reconnect returns one receipt without reroll or regrant; conflicting payload, foreign state and stale revision reject. Concurrent confirms commit exactly once. Inject failure after cost/effect/history and prove full rollback. Web and TUI consume the same engine availability and explicit unavailable reasons. Unqualified effects stay unavailable; source metadata and closed planning issues grant nothing. SRD5.2.1 CC-BY4 attribution/adaptation is retained via docs/srd-5.2.1-conformance.md.

### Qualify Warlock investment of the chain master (qualification)

Decision #331. One implementation PR; Task Size M.

PDF73: own5+/PactChain, when casting Find Familiar grant Fly or Swim40; Bonus Action commands Attack action; optional B/P/S->Necrotic/Radiant per damage; familiar saveDC yours; Reaction Resistance against that damage. Positive40ft and actual familiar saveDC; reject both speeds, replacing Fire, free reaction, command without Bonus Action. Split five benefits and familiar infusion lifecycle.

Prepare is inert: no dice, cost, grant or writes. Confirm revalidates actor, campaign, source, revision, entitlement, target, action, timing and operands. Persist choice, injected dice, resources, effects, derived sheet and history atomically in isolated PostgreSQL. Same-ID replay/reload/reconnect returns one receipt without reroll or regrant; conflicting payload, foreign state and stale revision reject. Concurrent confirms commit exactly once. Inject failure after cost/effect/history and prove full rollback. Web and TUI consume the same engine availability and explicit unavailable reasons. Unqualified effects stay unavailable; source metadata and closed planning issues grant nothing. SRD5.2.1 CC-BY4 attribution/adaptation is retained via docs/srd-5.2.1-conformance.md.

### Implement Warlock lessons of the first ones (implementation)

Decision #331. One implementation PR; Task Size M.

PDF73: own2+, gain1 Origin feat via actual384/385–388 effects; repeat only different Origin feat. Positive Alert then Skilled; reject duplicate Alert, General feat, nonqualified feat effect or ignoring nested choices.

Prepare is inert: no dice, cost, grant or writes. Confirm revalidates actor, campaign, source, revision, entitlement, target, action, timing and operands. Persist choice, injected dice, resources, effects, derived sheet and history atomically in isolated PostgreSQL. Same-ID replay/reload/reconnect returns one receipt without reroll or regrant; conflicting payload, foreign state and stale revision reject. Concurrent confirms commit exactly once. Inject failure after cost/effect/history and prove full rollback. Web and TUI consume the same engine availability and explicit unavailable reasons. Unqualified effects stay unavailable; source metadata and closed planning issues grant nothing. SRD5.2.1 CC-BY4 attribution/adaptation is retained via docs/srd-5.2.1-conformance.md.

### Implement Warlock lifedrinker (implementation)

Decision #331. One implementation PR; Task Size M.

PDF73: own9+/PactBlade once per turn pact-weapon hit creature extra1d6 Necrotic/Psychic/Radiant; optionally spend one own Hit Die, heal max(1,dieRoll+CONmod). Positive D8roll4/CON2 heals6; reject zero HitDice healing, attack miss, second per turn, d6 forced for HitDie or wrong damage.

Prepare is inert: no dice, cost, grant or writes. Confirm revalidates actor, campaign, source, revision, entitlement, target, action, timing and operands. Persist choice, injected dice, resources, effects, derived sheet and history atomically in isolated PostgreSQL. Same-ID replay/reload/reconnect returns one receipt without reroll or regrant; conflicting payload, foreign state and stale revision reject. Concurrent confirms commit exactly once. Inject failure after cost/effect/history and prove full rollback. Web and TUI consume the same engine availability and explicit unavailable reasons. Unqualified effects stay unavailable; source metadata and closed planning issues grant nothing. SRD5.2.1 CC-BY4 attribution/adaptation is retained via docs/srd-5.2.1-conformance.md.

### Implement Warlock mask of many faces (implementation)

Decision #331. One implementation PR; Task Size M.

PDF73: own2+, Disguise Self without slot. Positive0slots normal spell targets; reject own1, fabricated effect or blanket self-only restriction absent source.

Prepare is inert: no dice, cost, grant or writes. Confirm revalidates actor, campaign, source, revision, entitlement, target, action, timing and operands. Persist choice, injected dice, resources, effects, derived sheet and history atomically in isolated PostgreSQL. Same-ID replay/reload/reconnect returns one receipt without reroll or regrant; conflicting payload, foreign state and stale revision reject. Concurrent confirms commit exactly once. Inject failure after cost/effect/history and prove full rollback. Web and TUI consume the same engine availability and explicit unavailable reasons. Unqualified effects stay unavailable; source metadata and closed planning issues grant nothing. SRD5.2.1 CC-BY4 attribution/adaptation is retained via docs/srd-5.2.1-conformance.md.

### Implement Warlock master of myriad forms (implementation)

Decision #331. One implementation PR; Task Size M.

PDF73: own5+, Alter Self without slot; consume appearance/aquatic/natural-weapons bounded owners reserved330. Positive0slots qualified mode; reject own4 or inferring unqualified natural weapons.

Prepare is inert: no dice, cost, grant or writes. Confirm revalidates actor, campaign, source, revision, entitlement, target, action, timing and operands. Persist choice, injected dice, resources, effects, derived sheet and history atomically in isolated PostgreSQL. Same-ID replay/reload/reconnect returns one receipt without reroll or regrant; conflicting payload, foreign state and stale revision reject. Concurrent confirms commit exactly once. Inject failure after cost/effect/history and prove full rollback. Web and TUI consume the same engine availability and explicit unavailable reasons. Unqualified effects stay unavailable; source metadata and closed planning issues grant nothing. SRD5.2.1 CC-BY4 attribution/adaptation is retained via docs/srd-5.2.1-conformance.md.

### Implement Warlock misty visions (implementation)

Decision #331. One implementation PR; Task Size M.

PDF74: own2+, Silent Image without slot; consume real effect. Positive0slots; reject own1, treating image as real object or copied spell owner.

Prepare is inert: no dice, cost, grant or writes. Confirm revalidates actor, campaign, source, revision, entitlement, target, action, timing and operands. Persist choice, injected dice, resources, effects, derived sheet and history atomically in isolated PostgreSQL. Same-ID replay/reload/reconnect returns one receipt without reroll or regrant; conflicting payload, foreign state and stale revision reject. Concurrent confirms commit exactly once. Inject failure after cost/effect/history and prove full rollback. Web and TUI consume the same engine availability and explicit unavailable reasons. Unqualified effects stay unavailable; source metadata and closed planning issues grant nothing. SRD5.2.1 CC-BY4 attribution/adaptation is retained via docs/srd-5.2.1-conformance.md.

### Implement Warlock one with shadows (implementation)

Decision #331. One implementation PR; Task Size M.

PDF74: own5+, while in DimLight/Darkness cast Invisibility self without slot. Positive DimLight self; reject BrightLight, own4, other target or inferred immobility ending from2014.

Prepare is inert: no dice, cost, grant or writes. Confirm revalidates actor, campaign, source, revision, entitlement, target, action, timing and operands. Persist choice, injected dice, resources, effects, derived sheet and history atomically in isolated PostgreSQL. Same-ID replay/reload/reconnect returns one receipt without reroll or regrant; conflicting payload, foreign state and stale revision reject. Concurrent confirms commit exactly once. Inject failure after cost/effect/history and prove full rollback. Web and TUI consume the same engine availability and explicit unavailable reasons. Unqualified effects stay unavailable; source metadata and closed planning issues grant nothing. SRD5.2.1 CC-BY4 attribution/adaptation is retained via docs/srd-5.2.1-conformance.md.

### Implement Warlock otherworldly leap (implementation)

Decision #331. One implementation PR; Task Size M.

PDF74: own2+, Jump self without slot. Positive self0slots; reject other target, own1 or general free Jump permission.

Prepare is inert: no dice, cost, grant or writes. Confirm revalidates actor, campaign, source, revision, entitlement, target, action, timing and operands. Persist choice, injected dice, resources, effects, derived sheet and history atomically in isolated PostgreSQL. Same-ID replay/reload/reconnect returns one receipt without reroll or regrant; conflicting payload, foreign state and stale revision reject. Concurrent confirms commit exactly once. Inject failure after cost/effect/history and prove full rollback. Web and TUI consume the same engine availability and explicit unavailable reasons. Unqualified effects stay unavailable; source metadata and closed planning issues grant nothing. SRD5.2.1 CC-BY4 attribution/adaptation is retained via docs/srd-5.2.1-conformance.md.

### Qualify Warlock pact of the blade (qualification)

Decision #331. One implementation PR; Task Size M.

PDF74: Bonus Action conjure chosen Simple/Martial Melee weapon in hand and bond, or touch magic weapon not attuned to someone else/not bonded to another Warlock. Grant proficiency/focus; attacks optionally CHA and normal/Necrotic/Psychic/Radiant. Bond ends reuse BA, death or >5ft for1minute; conjured weapon disappears. Positive distance6ft60s ends; reject5ft,59s, other attunement, ranged mundane conjure, stale old bond. Split conjure/bond, combat, focus/proficiency, lifecycle; no general item effects inferred.

Prepare is inert: no dice, cost, grant or writes. Confirm revalidates actor, campaign, source, revision, entitlement, target, action, timing and operands. Persist choice, injected dice, resources, effects, derived sheet and history atomically in isolated PostgreSQL. Same-ID replay/reload/reconnect returns one receipt without reroll or regrant; conflicting payload, foreign state and stale revision reject. Concurrent confirms commit exactly once. Inject failure after cost/effect/history and prove full rollback. Web and TUI consume the same engine availability and explicit unavailable reasons. Unqualified effects stay unavailable; source metadata and closed planning issues grant nothing. SRD5.2.1 CC-BY4 attribution/adaptation is retained via docs/srd-5.2.1-conformance.md.

### Qualify Warlock pact of the chain (qualification)

Decision #331. One implementation PR; Task Size M.

PDF74: learn Find Familiar; Magic-action slot-free casting; normal forms or Imp/Pseudodragon/Quasit/Skeleton/Sphinx of Wonder/Sprite/Venomous Snake. On own Attack action forgo1 own attack for familiar1 attack using its Reaction. Positive1 forfeited/1 familiar Reaction; reject free familiar attack, spent Reaction, multiattack substitution or unqualified form/statblock. Split spell mode/forms/attack exchange.

Prepare is inert: no dice, cost, grant or writes. Confirm revalidates actor, campaign, source, revision, entitlement, target, action, timing and operands. Persist choice, injected dice, resources, effects, derived sheet and history atomically in isolated PostgreSQL. Same-ID replay/reload/reconnect returns one receipt without reroll or regrant; conflicting payload, foreign state and stale revision reject. Concurrent confirms commit exactly once. Inject failure after cost/effect/history and prove full rollback. Web and TUI consume the same engine availability and explicit unavailable reasons. Unqualified effects stay unavailable; source metadata and closed planning issues grant nothing. SRD5.2.1 CC-BY4 attribution/adaptation is retained via docs/srd-5.2.1-conformance.md.

### Qualify Warlock pact of the tome (qualification)

Decision #331. One implementation PR; Task Size M.

PDF74: at end Short/Long Rest conjure book in hand; choose3 cantrips from finite27 and2 level1 Rituals from finite11, none already prepared. While book on person chosen spells prepared as Warlock spells; book focus. Only you access magic; new book/death disappears old. Positive3+2 distinct eligible unprepared; reject4 cantrips, level2 ritual, already prepared, absent-book grant or2014 transcription. Split conjuration/lifecycle, finite choices/provenance, focus; normal prepared Ritual cast+10minutes/no slot/no upcast, no extra RitualAdept.

Prepare is inert: no dice, cost, grant or writes. Confirm revalidates actor, campaign, source, revision, entitlement, target, action, timing and operands. Persist choice, injected dice, resources, effects, derived sheet and history atomically in isolated PostgreSQL. Same-ID replay/reload/reconnect returns one receipt without reroll or regrant; conflicting payload, foreign state and stale revision reject. Concurrent confirms commit exactly once. Inject failure after cost/effect/history and prove full rollback. Web and TUI consume the same engine availability and explicit unavailable reasons. Unqualified effects stay unavailable; source metadata and closed planning issues grant nothing. SRD5.2.1 CC-BY4 attribution/adaptation is retained via docs/srd-5.2.1-conformance.md.

### Implement Warlock repelling blast (implementation)

Decision #331. One implementation PR; Task Size M.

PDF74: own2+, known Warlock cantrip requiring attack roll; hit Large/smaller pushes up to10ft straight away. Repeat different eligible cantrip. Positive Large10ft; reject Huge,11ft, sideways push, saving-throw cantrip or duplicate selection.

Prepare is inert: no dice, cost, grant or writes. Confirm revalidates actor, campaign, source, revision, entitlement, target, action, timing and operands. Persist choice, injected dice, resources, effects, derived sheet and history atomically in isolated PostgreSQL. Same-ID replay/reload/reconnect returns one receipt without reroll or regrant; conflicting payload, foreign state and stale revision reject. Concurrent confirms commit exactly once. Inject failure after cost/effect/history and prove full rollback. Web and TUI consume the same engine availability and explicit unavailable reasons. Unqualified effects stay unavailable; source metadata and closed planning issues grant nothing. SRD5.2.1 CC-BY4 attribution/adaptation is retained via docs/srd-5.2.1-conformance.md.

### Implement Warlock thirsting blade (implementation)

Decision #331. One implementation PR; Task Size M.

PDF74/25: own5+/PactBlade; Attack action on own turn twice with pact weapon only. Positive2 attacks; reject nonpact, offturn, third attack, stacking independent ExtraAttack; DevouringBlade explicit exception.

Prepare is inert: no dice, cost, grant or writes. Confirm revalidates actor, campaign, source, revision, entitlement, target, action, timing and operands. Persist choice, injected dice, resources, effects, derived sheet and history atomically in isolated PostgreSQL. Same-ID replay/reload/reconnect returns one receipt without reroll or regrant; conflicting payload, foreign state and stale revision reject. Concurrent confirms commit exactly once. Inject failure after cost/effect/history and prove full rollback. Web and TUI consume the same engine availability and explicit unavailable reasons. Unqualified effects stay unavailable; source metadata and closed planning issues grant nothing. SRD5.2.1 CC-BY4 attribution/adaptation is retained via docs/srd-5.2.1-conformance.md.

### Implement Warlock visions of distant realms (implementation)

Decision #331. One implementation PR; Task Size M.

PDF74: own9+, Arcane Eye without slot, real qualified sensor/movement owners reserved330. Positive0slots; reject own8 or treating sensor observation as omniscience.

Prepare is inert: no dice, cost, grant or writes. Confirm revalidates actor, campaign, source, revision, entitlement, target, action, timing and operands. Persist choice, injected dice, resources, effects, derived sheet and history atomically in isolated PostgreSQL. Same-ID replay/reload/reconnect returns one receipt without reroll or regrant; conflicting payload, foreign state and stale revision reject. Concurrent confirms commit exactly once. Inject failure after cost/effect/history and prove full rollback. Web and TUI consume the same engine availability and explicit unavailable reasons. Unqualified effects stay unavailable; source metadata and closed planning issues grant nothing. SRD5.2.1 CC-BY4 attribution/adaptation is retained via docs/srd-5.2.1-conformance.md.

### Implement Warlock whispers of the grave (implementation)

Decision #331. One implementation PR; Task Size M.

PDF74: own7+, Speak with Dead without slot; exact dead-creature effect owner, no generic undead command. Positive own7/0slots; reject own6 or unqualified effect.

Prepare is inert: no dice, cost, grant or writes. Confirm revalidates actor, campaign, source, revision, entitlement, target, action, timing and operands. Persist choice, injected dice, resources, effects, derived sheet and history atomically in isolated PostgreSQL. Same-ID replay/reload/reconnect returns one receipt without reroll or regrant; conflicting payload, foreign state and stale revision reject. Concurrent confirms commit exactly once. Inject failure after cost/effect/history and prove full rollback. Web and TUI consume the same engine availability and explicit unavailable reasons. Unqualified effects stay unavailable; source metadata and closed planning issues grant nothing. SRD5.2.1 CC-BY4 attribution/adaptation is retained via docs/srd-5.2.1-conformance.md.

### Implement Warlock witch sight (implementation)

Decision #331. One implementation PR; Task Size M.

PDF74: own15+, Truesight30ft; qualified glossary perception including actual962 Invisible obligations. Positive30ft; reject31ft, own14, unlimited private disclosure.

Prepare is inert: no dice, cost, grant or writes. Confirm revalidates actor, campaign, source, revision, entitlement, target, action, timing and operands. Persist choice, injected dice, resources, effects, derived sheet and history atomically in isolated PostgreSQL. Same-ID replay/reload/reconnect returns one receipt without reroll or regrant; conflicting payload, foreign state and stale revision reject. Concurrent confirms commit exactly once. Inject failure after cost/effect/history and prove full rollback. Web and TUI consume the same engine availability and explicit unavailable reasons. Unqualified effects stay unavailable; source metadata and closed planning issues grant nothing. SRD5.2.1 CC-BY4 attribution/adaptation is retained via docs/srd-5.2.1-conformance.md.

### Implement Warlock blade bond (implementation)

Decision #331. One implementation PR; Task Size M.

PDF74: Bonus Action conjure chosen Simple/Martial Melee weapon in hand and bond OR touch magic weapon; reject other attunement/Warlock bond. Positive mundane Sickle in hand; reject Longbow conjure, remote magic weapon, second live bond. No copied magic-item effect or arbitrary item definition.

Prepare is inert: no dice, cost, grant or writes. Confirm revalidates actor, campaign, source, revision, entitlement, target, action, timing and operands. Persist choice, injected dice, resources, effects, derived sheet and history atomically in isolated PostgreSQL. Same-ID replay/reload/reconnect returns one receipt without reroll or regrant; conflicting payload, foreign state and stale revision reject. Concurrent confirms commit exactly once. Inject failure after cost/effect/history and prove full rollback. Web and TUI consume the same engine availability and explicit unavailable reasons. Unqualified effects stay unavailable; source metadata and closed planning issues grant nothing. SRD5.2.1 CC-BY4 attribution/adaptation is retained via docs/srd-5.2.1-conformance.md.

### Implement Warlock blade combat (implementation)

Decision #331. One implementation PR; Task Size M.

PDF74: bonded weapon attacks optionally use CHA for attack AND damage, choose normal/Necrotic/Psychic/Radiant per attack. Positive CHA3 normal weapon damage+3; reject CHA substitution on unbonded weapon, Force option or mixed attack/damage ability provenance.

Prepare is inert: no dice, cost, grant or writes. Confirm revalidates actor, campaign, source, revision, entitlement, target, action, timing and operands. Persist choice, injected dice, resources, effects, derived sheet and history atomically in isolated PostgreSQL. Same-ID replay/reload/reconnect returns one receipt without reroll or regrant; conflicting payload, foreign state and stale revision reject. Concurrent confirms commit exactly once. Inject failure after cost/effect/history and prove full rollback. Web and TUI consume the same engine availability and explicit unavailable reasons. Unqualified effects stay unavailable; source metadata and closed planning issues grant nothing. SRD5.2.1 CC-BY4 attribution/adaptation is retained via docs/srd-5.2.1-conformance.md.

### Implement Warlock blade proficiency focus (implementation)

Decision #331. One implementation PR; Task Size M.

PDF74: proficiency and SpellcastingFocus exist only until exact weapon bond ends; use430provenance. Positive bonded proficient/qualifiedfocus; reject global martial proficiency, unrelated item focus, removing other source proficiency on bond end.

Prepare is inert: no dice, cost, grant or writes. Confirm revalidates actor, campaign, source, revision, entitlement, target, action, timing and operands. Persist choice, injected dice, resources, effects, derived sheet and history atomically in isolated PostgreSQL. Same-ID replay/reload/reconnect returns one receipt without reroll or regrant; conflicting payload, foreign state and stale revision reject. Concurrent confirms commit exactly once. Inject failure after cost/effect/history and prove full rollback. Web and TUI consume the same engine availability and explicit unavailable reasons. Unqualified effects stay unavailable; source metadata and closed planning issues grant nothing. SRD5.2.1 CC-BY4 attribution/adaptation is retained via docs/srd-5.2.1-conformance.md.

### Implement Warlock blade ending (implementation)

Decision #331. One implementation PR; Task Size M.

PDF74: bond ends feature BonusAction reused, death, or weapon more than5ft for1minute; conjured weapon disappears then. Positive6ft60s; reject5ft/60s,6ft59s or disappearing nonconjured magic weapon; race new bond against old ending atomically.

Prepare is inert: no dice, cost, grant or writes. Confirm revalidates actor, campaign, source, revision, entitlement, target, action, timing and operands. Persist choice, injected dice, resources, effects, derived sheet and history atomically in isolated PostgreSQL. Same-ID replay/reload/reconnect returns one receipt without reroll or regrant; conflicting payload, foreign state and stale revision reject. Concurrent confirms commit exactly once. Inject failure after cost/effect/history and prove full rollback. Web and TUI consume the same engine availability and explicit unavailable reasons. Unqualified effects stay unavailable; source metadata and closed planning issues grant nothing. SRD5.2.1 CC-BY4 attribution/adaptation is retained via docs/srd-5.2.1-conformance.md.

### Implement Warlock chain cast mode (implementation)

Decision #331. One implementation PR; Task Size M.

PDF74: grant FindFamiliar identity; cast via Magic action without slot, retaining source-specific component exceptions only if stated. Positive0slots Magic action; reject generic1-hour casting for this mode or unrelated spell freecast. Named base effect owner mandatory; planning181/185/274 not sufficient.

Prepare is inert: no dice, cost, grant or writes. Confirm revalidates actor, campaign, source, revision, entitlement, target, action, timing and operands. Persist choice, injected dice, resources, effects, derived sheet and history atomically in isolated PostgreSQL. Same-ID replay/reload/reconnect returns one receipt without reroll or regrant; conflicting payload, foreign state and stale revision reject. Concurrent confirms commit exactly once. Inject failure after cost/effect/history and prove full rollback. Web and TUI consume the same engine availability and explicit unavailable reasons. Unqualified effects stay unavailable; source metadata and closed planning issues grant nothing. SRD5.2.1 CC-BY4 attribution/adaptation is retained via docs/srd-5.2.1-conformance.md.

### Implement Warlock chain form choice (implementation)

Decision #331. One implementation PR; Task Size M.

PDF74: choose normal FindFamiliar form OR exactly Imp/Pseudodragon/Quasit/Skeleton/SphinxofWonder/Sprite/VenomousSnake; each needs own real named statblock/actions/availability outputs. Positive qualifiedImp; reject unknownDragon, unqualifiedImp, copying2014forms, treating name as statblock.

Prepare is inert: no dice, cost, grant or writes. Confirm revalidates actor, campaign, source, revision, entitlement, target, action, timing and operands. Persist choice, injected dice, resources, effects, derived sheet and history atomically in isolated PostgreSQL. Same-ID replay/reload/reconnect returns one receipt without reroll or regrant; conflicting payload, foreign state and stale revision reject. Concurrent confirms commit exactly once. Inject failure after cost/effect/history and prove full rollback. Web and TUI consume the same engine availability and explicit unavailable reasons. Unqualified effects stay unavailable; source metadata and closed planning issues grant nothing. SRD5.2.1 CC-BY4 attribution/adaptation is retained via docs/srd-5.2.1-conformance.md.

### Implement Warlock chain attack exchange (implementation)

Decision #331. One implementation PR; Task Size M.

PDF74: on own Attack action forgo1 own attack to allow familiar1 attack using its Reaction. Positive actor2attacks forgo1/familiarReaction1; reject extra free attack, Reaction0, nonAttackaction, multipleattacks from one exchange. Consume exact familiar attack owner, never automaticMultiattack.

Prepare is inert: no dice, cost, grant or writes. Confirm revalidates actor, campaign, source, revision, entitlement, target, action, timing and operands. Persist choice, injected dice, resources, effects, derived sheet and history atomically in isolated PostgreSQL. Same-ID replay/reload/reconnect returns one receipt without reroll or regrant; conflicting payload, foreign state and stale revision reject. Concurrent confirms commit exactly once. Inject failure after cost/effect/history and prove full rollback. Web and TUI consume the same engine availability and explicit unavailable reasons. Unqualified effects stay unavailable; source metadata and closed planning issues grant nothing. SRD5.2.1 CC-BY4 attribution/adaptation is retained via docs/srd-5.2.1-conformance.md.

### Implement Warlock tome conjuration ending (implementation)

Decision #331. One implementation PR; Task Size M.

PDF74: end completed Short/LongRest conjure book in hand; old book disappears on another conjure or death. Positive completedShortRest onebook; reject arbitraryMagicAction conjure, interruptedrest, twoactivebooks or other creature access.

Prepare is inert: no dice, cost, grant or writes. Confirm revalidates actor, campaign, source, revision, entitlement, target, action, timing and operands. Persist choice, injected dice, resources, effects, derived sheet and history atomically in isolated PostgreSQL. Same-ID replay/reload/reconnect returns one receipt without reroll or regrant; conflicting payload, foreign state and stale revision reject. Concurrent confirms commit exactly once. Inject failure after cost/effect/history and prove full rollback. Web and TUI consume the same engine availability and explicit unavailable reasons. Unqualified effects stay unavailable; source metadata and closed planning issues grant nothing. SRD5.2.1 CC-BY4 attribution/adaptation is retained via docs/srd-5.2.1-conformance.md.

### Implement Warlock tome finite preparation (implementation)

Decision #331. One implementation PR; Task Size M.

PDF74: on book appearance choose3 distinct cantrips from exact27 and2 distinct level1Rituals from exact11; none already prepared. While on person prepared as Warlock spells independent of class count. Positive3+2; reject4cantrips, level2Ritual, alreadyprepared, absentbook cast provenance. Eligibility can include finite unqualified spells; execution remains unavailable.

Prepare is inert: no dice, cost, grant or writes. Confirm revalidates actor, campaign, source, revision, entitlement, target, action, timing and operands. Persist choice, injected dice, resources, effects, derived sheet and history atomically in isolated PostgreSQL. Same-ID replay/reload/reconnect returns one receipt without reroll or regrant; conflicting payload, foreign state and stale revision reject. Concurrent confirms commit exactly once. Inject failure after cost/effect/history and prove full rollback. Web and TUI consume the same engine availability and explicit unavailable reasons. Unqualified effects stay unavailable; source metadata and closed planning issues grant nothing. SRD5.2.1 CC-BY4 attribution/adaptation is retained via docs/srd-5.2.1-conformance.md.

### Implement Warlock tome focus (implementation)

Decision #331. One implementation PR; Task Size M.

PDF74: own conjured BookofShadows usable as SpellcastingFocus; book must be accessible under real focus rules, not generic component waiver. Positive ownbook qualifiedcasting; reject otherbook, consumed/costlymaterial waiver or copied itemcasting behavior.

Prepare is inert: no dice, cost, grant or writes. Confirm revalidates actor, campaign, source, revision, entitlement, target, action, timing and operands. Persist choice, injected dice, resources, effects, derived sheet and history atomically in isolated PostgreSQL. Same-ID replay/reload/reconnect returns one receipt without reroll or regrant; conflicting payload, foreign state and stale revision reject. Concurrent confirms commit exactly once. Inject failure after cost/effect/history and prove full rollback. Web and TUI consume the same engine availability and explicit unavailable reasons. Unqualified effects stay unavailable; source metadata and closed planning issues grant nothing. SRD5.2.1 CC-BY4 attribution/adaptation is retained via docs/srd-5.2.1-conformance.md.

### Implement Warlock gaze connection (implementation)

Decision #331. One implementation PR; Task Size M.

PDF73: BonusAction touch willing creature; perceive special senses until end nextturn; sameplane BonusAction each subsequentturn extends. Positive maintain nextturn; reject unwilling, crossplane, missedBA; actual perception owner cannot reveal private unavailable facts.

Prepare is inert: no dice, cost, grant or writes. Confirm revalidates actor, campaign, source, revision, entitlement, target, action, timing and operands. Persist choice, injected dice, resources, effects, derived sheet and history atomically in isolated PostgreSQL. Same-ID replay/reload/reconnect returns one receipt without reroll or regrant; conflicting payload, foreign state and stale revision reject. Concurrent confirms commit exactly once. Inject failure after cost/effect/history and prove full rollback. Web and TUI consume the same engine availability and explicit unavailable reasons. Unqualified effects stay unavailable; source metadata and closed planning issues grant nothing. SRD5.2.1 CC-BY4 attribution/adaptation is retained via docs/srd-5.2.1-conformance.md.

### Implement Warlock gaze cast origin (implementation)

Decision #331. One implementation PR; Task Size M.

PDF73: while connection active choose own or linkedcreature spellorigin only within60ft. Positive60ft alternateorigin; reject61ft, expiredconnection or target/range measured from bothorigins. Consume exact qualified effect, targeting and casting protocols.

Prepare is inert: no dice, cost, grant or writes. Confirm revalidates actor, campaign, source, revision, entitlement, target, action, timing and operands. Persist choice, injected dice, resources, effects, derived sheet and history atomically in isolated PostgreSQL. Same-ID replay/reload/reconnect returns one receipt without reroll or regrant; conflicting payload, foreign state and stale revision reject. Concurrent confirms commit exactly once. Inject failure after cost/effect/history and prove full rollback. Web and TUI consume the same engine availability and explicit unavailable reasons. Unqualified effects stay unavailable; source metadata and closed planning issues grant nothing. SRD5.2.1 CC-BY4 attribution/adaptation is retained via docs/srd-5.2.1-conformance.md.

### Implement Warlock depths passive (implementation)

Decision #331. One implementation PR; Task Size M.

PDF73: own5+ breathe underwater and SwimSpeed equals current Speed. PositiveSpeed30->Swim30, speed20->Swim20; reject fixed40 or no underwaterbreathing. Do not infer aquatic statblock or environmental immunity.

Prepare is inert: no dice, cost, grant or writes. Confirm revalidates actor, campaign, source, revision, entitlement, target, action, timing and operands. Persist choice, injected dice, resources, effects, derived sheet and history atomically in isolated PostgreSQL. Same-ID replay/reload/reconnect returns one receipt without reroll or regrant; conflicting payload, foreign state and stale revision reject. Concurrent confirms commit exactly once. Inject failure after cost/effect/history and prove full rollback. Web and TUI consume the same engine availability and explicit unavailable reasons. Unqualified effects stay unavailable; source metadata and closed planning issues grant nothing. SRD5.2.1 CC-BY4 attribution/adaptation is retained via docs/srd-5.2.1-conformance.md.

### Implement Warlock depths cast mode (implementation)

Decision #331. One implementation PR; Task Size M.

PDF73: own5+ cast WaterBreathing withoutslot onceLongRest. Positive uses1->0 at0slots; reject seconduse, ShortRestreset or copied spell effect. Exact spellowner mandatory.

Prepare is inert: no dice, cost, grant or writes. Confirm revalidates actor, campaign, source, revision, entitlement, target, action, timing and operands. Persist choice, injected dice, resources, effects, derived sheet and history atomically in isolated PostgreSQL. Same-ID replay/reload/reconnect returns one receipt without reroll or regrant; conflicting payload, foreign state and stale revision reject. Concurrent confirms commit exactly once. Inject failure after cost/effect/history and prove full rollback. Web and TUI consume the same engine availability and explicit unavailable reasons. Unqualified effects stay unavailable; source metadata and closed planning issues grant nothing. SRD5.2.1 CC-BY4 attribution/adaptation is retained via docs/srd-5.2.1-conformance.md.

### Implement Warlock protectors inscription (implementation)

Decision #331. One implementation PR; Task Size M.

PDF73: own9+/Tome, newpage on conjurebook; with permission creature action writes ownname, max(1,CHAmod)names; MagicActiontouch erases. PositiveCHA2 twonames; reject thirdname/no permission/freeerase. Actualbookidentity/lifecycle mandatory.

Prepare is inert: no dice, cost, grant or writes. Confirm revalidates actor, campaign, source, revision, entitlement, target, action, timing and operands. Persist choice, injected dice, resources, effects, derived sheet and history atomically in isolated PostgreSQL. Same-ID replay/reload/reconnect returns one receipt without reroll or regrant; conflicting payload, foreign state and stale revision reject. Concurrent confirms commit exactly once. Inject failure after cost/effect/history and prove full rollback. Web and TUI consume the same engine availability and explicit unavailable reasons. Unqualified effects stay unavailable; source metadata and closed planning issues grant nothing. SRD5.2.1 CC-BY4 attribution/adaptation is retained via docs/srd-5.2.1-conformance.md.

### Implement Warlock protectors rescue (implementation)

Decision #331. One implementation PR; Task Size M.

PDF73: namedcreature reduced0HP not killedoutright drops1HP instead; shared global trigger onceLongRest. Positive twonames firstrescue only; reject outrightdeath, secondname before rest, TempHP insteadHP or trigger on unnamedcreature. AtomicHP/lifecycle/ending owner required.

Prepare is inert: no dice, cost, grant or writes. Confirm revalidates actor, campaign, source, revision, entitlement, target, action, timing and operands. Persist choice, injected dice, resources, effects, derived sheet and history atomically in isolated PostgreSQL. Same-ID replay/reload/reconnect returns one receipt without reroll or regrant; conflicting payload, foreign state and stale revision reject. Concurrent confirms commit exactly once. Inject failure after cost/effect/history and prove full rollback. Web and TUI consume the same engine availability and explicit unavailable reasons. Unqualified effects stay unavailable; source metadata and closed planning issues grant nothing. SRD5.2.1 CC-BY4 attribution/adaptation is retained via docs/srd-5.2.1-conformance.md.

### Implement Warlock investment movement (implementation)

Decision #331. One implementation PR; Task Size M.

PDF73: own5+/Chain when castingFindFamiliar infuse summonedfamiliar; chooseFly40 OR Swim40. PositiveFly40; reject both, speed30, retroactive unrelatedcreature. Realformstatblock/lifecycle required.

Prepare is inert: no dice, cost, grant or writes. Confirm revalidates actor, campaign, source, revision, entitlement, target, action, timing and operands. Persist choice, injected dice, resources, effects, derived sheet and history atomically in isolated PostgreSQL. Same-ID replay/reload/reconnect returns one receipt without reroll or regrant; conflicting payload, foreign state and stale revision reject. Concurrent confirms commit exactly once. Inject failure after cost/effect/history and prove full rollback. Web and TUI consume the same engine availability and explicit unavailable reasons. Unqualified effects stay unavailable; source metadata and closed planning issues grant nothing. SRD5.2.1 CC-BY4 attribution/adaptation is retained via docs/srd-5.2.1-conformance.md.

### Implement Warlock investment command (implementation)

Decision #331. One implementation PR; Task Size M.

PDF73/257: BonusAction command familiar Attack action, consume actual familiar action economy/attack implementation. Source257Multiattack IS partAttackaction: qualifiedPseudodragon mayuseTWO Bites; this is distinct from ChainONEattackexchange. Positive1BA/twoactualBites; reject noBA, unrelatedcreature, unsupported attack, twoStings or Chainforgoneattackforfree.

Prepare is inert: no dice, cost, grant or writes. Confirm revalidates actor, campaign, source, revision, entitlement, target, action, timing and operands. Persist choice, injected dice, resources, effects, derived sheet and history atomically in isolated PostgreSQL. Same-ID replay/reload/reconnect returns one receipt without reroll or regrant; conflicting payload, foreign state and stale revision reject. Concurrent confirms commit exactly once. Inject failure after cost/effect/history and prove full rollback. Web and TUI consume the same engine availability and explicit unavailable reasons. Unqualified effects stay unavailable; source metadata and closed planning issues grant nothing. SRD5.2.1 CC-BY4 attribution/adaptation is retained via docs/srd-5.2.1-conformance.md.

### Implement Warlock investment damage (implementation)

Decision #331. One implementation PR; Task Size M.

PDF73: whenever familiar deals B/P/S optionally make thatdamage Necrotic or Radiant. PositivePiercing->Radiant; rejectFire->Radiant, Force choice, changing extra unrelateddamage.

Prepare is inert: no dice, cost, grant or writes. Confirm revalidates actor, campaign, source, revision, entitlement, target, action, timing and operands. Persist choice, injected dice, resources, effects, derived sheet and history atomically in isolated PostgreSQL. Same-ID replay/reload/reconnect returns one receipt without reroll or regrant; conflicting payload, foreign state and stale revision reject. Concurrent confirms commit exactly once. Inject failure after cost/effect/history and prove full rollback. Web and TUI consume the same engine availability and explicit unavailable reasons. Unqualified effects stay unavailable; source metadata and closed planning issues grant nothing. SRD5.2.1 CC-BY4 attribution/adaptation is retained via docs/srd-5.2.1-conformance.md.

### Implement Warlock investment save dc (implementation)

Decision #331. One implementation PR; Task Size M.

PDF73: if familiar forces save use your spellSaveDC. PositiveCHA3/PB3/DC14; reject creatureprintedDC, attackingwithDC or changing nonfamiliar save.

Prepare is inert: no dice, cost, grant or writes. Confirm revalidates actor, campaign, source, revision, entitlement, target, action, timing and operands. Persist choice, injected dice, resources, effects, derived sheet and history atomically in isolated PostgreSQL. Same-ID replay/reload/reconnect returns one receipt without reroll or regrant; conflicting payload, foreign state and stale revision reject. Concurrent confirms commit exactly once. Inject failure after cost/effect/history and prove full rollback. Web and TUI consume the same engine availability and explicit unavailable reasons. Unqualified effects stay unavailable; source metadata and closed planning issues grant nothing. SRD5.2.1 CC-BY4 attribution/adaptation is retained via docs/srd-5.2.1-conformance.md.

### Implement Warlock investment resistance (implementation)

Decision #331. One implementation PR; Task Size M.

PDF73: when familiar takesdamage use ownReaction to grant Resistance against thatdamage. PositiveReaction1/Fire9->4; reject Reaction0, unrelateddamagepacket or persistentResistance afterpacket.

Prepare is inert: no dice, cost, grant or writes. Confirm revalidates actor, campaign, source, revision, entitlement, target, action, timing and operands. Persist choice, injected dice, resources, effects, derived sheet and history atomically in isolated PostgreSQL. Same-ID replay/reload/reconnect returns one receipt without reroll or regrant; conflicting payload, foreign state and stale revision reject. Concurrent confirms commit exactly once. Inject failure after cost/effect/history and prove full rollback. Web and TUI consume the same engine availability and explicit unavailable reasons. Unqualified effects stay unavailable; source metadata and closed planning issues grant nothing. SRD5.2.1 CC-BY4 attribution/adaptation is retained via docs/srd-5.2.1-conformance.md.

### Implement Warlock hurl admission cost (implementation)

Decision #331. One implementation PR; Task Size M.

PDF76: own14+/Fiend onceperturn attackroll hitcreature, ownspellDC Charisma save; onceLongRest use, noactionPactslot restoresuse. PositivePactslot1 restores; reject miss, seconduse sameturn, genericslot restore, resource restoration treated slotTOCAST.

Prepare is inert: no dice, cost, grant or writes. Confirm revalidates actor, campaign, source, revision, entitlement, target, action, timing and operands. Persist choice, injected dice, resources, effects, derived sheet and history atomically in isolated PostgreSQL. Same-ID replay/reload/reconnect returns one receipt without reroll or regrant; conflicting payload, foreign state and stale revision reject. Concurrent confirms commit exactly once. Inject failure after cost/effect/history and prove full rollback. Web and TUI consume the same engine availability and explicit unavailable reasons. Unqualified effects stay unavailable; source metadata and closed planning issues grant nothing. SRD5.2.1 CC-BY4 attribution/adaptation is retained via docs/srd-5.2.1-conformance.md.

### Implement Warlock hurl transit ending (implementation)

Decision #331. One implementation PR; Task Size M.

PDF76: failedsave targetdisappears, Incapacitated via487 until end YOURnextturn, then return prior or nearestunoccupiedspace. Positive blockedprior usesnearest; reject successfulsave transport, ending targetnextturn, unknownworldoccupancy or unsupportedplane behavior. Real boundedworld/transit owner required.

Prepare is inert: no dice, cost, grant or writes. Confirm revalidates actor, campaign, source, revision, entitlement, target, action, timing and operands. Persist choice, injected dice, resources, effects, derived sheet and history atomically in isolated PostgreSQL. Same-ID replay/reload/reconnect returns one receipt without reroll or regrant; conflicting payload, foreign state and stale revision reject. Concurrent confirms commit exactly once. Inject failure after cost/effect/history and prove full rollback. Web and TUI consume the same engine availability and explicit unavailable reasons. Unqualified effects stay unavailable; source metadata and closed planning issues grant nothing. SRD5.2.1 CC-BY4 attribution/adaptation is retained via docs/srd-5.2.1-conformance.md.

### Implement Warlock hurl psychic (implementation)

Decision #331. One implementation PR; Task Size M.

PDF76: transported failedsave nonFiend takes8d10Psychic; Fiend takes0 but still transports. Positive8injected10s=80 beforequalifieddefenses; reject Psychic toFiend, success damage, wrongdice or omittingFiendtransport.

Prepare is inert: no dice, cost, grant or writes. Confirm revalidates actor, campaign, source, revision, entitlement, target, action, timing and operands. Persist choice, injected dice, resources, effects, derived sheet and history atomically in isolated PostgreSQL. Same-ID replay/reload/reconnect returns one receipt without reroll or regrant; conflicting payload, foreign state and stale revision reject. Concurrent confirms commit exactly once. Inject failure after cost/effect/history and prove full rollback. Web and TUI consume the same engine availability and explicit unavailable reasons. Unqualified effects stay unavailable; source metadata and closed planning issues grant nothing. SRD5.2.1 CC-BY4 attribution/adaptation is retained via docs/srd-5.2.1-conformance.md.

### Implement Warlock arcanum 6 (implementation)

Decision #331. One implementation PR; Task Size M.

PDF72: ownWarlock11+ choose1 level6 Warlockspell; separate1use slot-free resource, LongRest recovery; each gainedWarlocklevel replace atmost1 arcanum across wholefeature with samelevel. Positive own17 retains resources6/7/8/9 separately, this6resource spends1 only; reject othersspent, Pactslotcost, ShortRestrecharge, wronglevel replacement or2global replacements. Actual named effect and specialcasting mode mandatory.

Prepare is inert: no dice, cost, grant or writes. Confirm revalidates actor, campaign, source, revision, entitlement, target, action, timing and operands. Persist choice, injected dice, resources, effects, derived sheet and history atomically in isolated PostgreSQL. Same-ID replay/reload/reconnect returns one receipt without reroll or regrant; conflicting payload, foreign state and stale revision reject. Concurrent confirms commit exactly once. Inject failure after cost/effect/history and prove full rollback. Web and TUI consume the same engine availability and explicit unavailable reasons. Unqualified effects stay unavailable; source metadata and closed planning issues grant nothing. SRD5.2.1 CC-BY4 attribution/adaptation is retained via docs/srd-5.2.1-conformance.md.

### Implement Warlock arcanum 7 (implementation)

Decision #331. One implementation PR; Task Size M.

PDF72: ownWarlock13+ choose1 level7 Warlockspell; separate1use slot-free resource, LongRest recovery; each gainedWarlocklevel replace atmost1 arcanum across wholefeature with samelevel. Positive own17 retains resources6/7/8/9 separately, this7resource spends1 only; reject othersspent, Pactslotcost, ShortRestrecharge, wronglevel replacement or2global replacements. Actual named effect and specialcasting mode mandatory.

Prepare is inert: no dice, cost, grant or writes. Confirm revalidates actor, campaign, source, revision, entitlement, target, action, timing and operands. Persist choice, injected dice, resources, effects, derived sheet and history atomically in isolated PostgreSQL. Same-ID replay/reload/reconnect returns one receipt without reroll or regrant; conflicting payload, foreign state and stale revision reject. Concurrent confirms commit exactly once. Inject failure after cost/effect/history and prove full rollback. Web and TUI consume the same engine availability and explicit unavailable reasons. Unqualified effects stay unavailable; source metadata and closed planning issues grant nothing. SRD5.2.1 CC-BY4 attribution/adaptation is retained via docs/srd-5.2.1-conformance.md.

### Implement Warlock arcanum 8 (implementation)

Decision #331. One implementation PR; Task Size M.

PDF72: ownWarlock15+ choose1 level8 Warlockspell; separate1use slot-free resource, LongRest recovery; each gainedWarlocklevel replace atmost1 arcanum across wholefeature with samelevel. Positive own17 retains resources6/7/8/9 separately, this8resource spends1 only; reject othersspent, Pactslotcost, ShortRestrecharge, wronglevel replacement or2global replacements. Actual named effect and specialcasting mode mandatory.

Prepare is inert: no dice, cost, grant or writes. Confirm revalidates actor, campaign, source, revision, entitlement, target, action, timing and operands. Persist choice, injected dice, resources, effects, derived sheet and history atomically in isolated PostgreSQL. Same-ID replay/reload/reconnect returns one receipt without reroll or regrant; conflicting payload, foreign state and stale revision reject. Concurrent confirms commit exactly once. Inject failure after cost/effect/history and prove full rollback. Web and TUI consume the same engine availability and explicit unavailable reasons. Unqualified effects stay unavailable; source metadata and closed planning issues grant nothing. SRD5.2.1 CC-BY4 attribution/adaptation is retained via docs/srd-5.2.1-conformance.md.

### Implement Warlock arcanum 9 (implementation)

Decision #331. One implementation PR; Task Size M.

PDF72: ownWarlock17+ choose1 level9 Warlockspell; separate1use slot-free resource, LongRest recovery; each gainedWarlocklevel replace atmost1 arcanum across wholefeature with samelevel. Positive own17 retains resources6/7/8/9 separately, this9resource spends1 only; reject othersspent, Pactslotcost, ShortRestrecharge, wronglevel replacement or2global replacements. Actual named effect and specialcasting mode mandatory.

Prepare is inert: no dice, cost, grant or writes. Confirm revalidates actor, campaign, source, revision, entitlement, target, action, timing and operands. Persist choice, injected dice, resources, effects, derived sheet and history atomically in isolated PostgreSQL. Same-ID replay/reload/reconnect returns one receipt without reroll or regrant; conflicting payload, foreign state and stale revision reject. Concurrent confirms commit exactly once. Inject failure after cost/effect/history and prove full rollback. Web and TUI consume the same engine availability and explicit unavailable reasons. Unqualified effects stay unavailable; source metadata and closed planning issues grant nothing. SRD5.2.1 CC-BY4 attribution/adaptation is retained via docs/srd-5.2.1-conformance.md.

### Implement Warlock level 1 (implementation)

Decision #331. One implementation PR; Task Size M.

PDF23–26/70–72/76: Warlock1, exact source row {'invocations/count': 1, 'cantrips/count': 2, 'prepared-spells/count': 2, 'pact-slots/count': 1, 'pact-slots/level': 1}. First-character level1 OR distinct later-entry contract with actual eligibility; own HP adapter mandatory, slots/Invocation/preparation by Warlock level; XP/live PB by complete canonical285 table at authoritative sum of admitted post-transition class levels. A first-class Warlock lineage does not gain a blanket multiclass CHA13 prerequisite on ordinary advancement; only actual later-entry/multiclass variants consume their own source eligibility. Positive Fighter4/Warlock1->Warlock2 requires total6/14000XP/PB3; Fighter4->Warlock1 total5/6500XP/PB3/later HP. Reject booleans/decimals/zero/nonadjacent/total>20/unqualified combinations. Apply every granted feature, nested option and replacement exactly once; repeated ASIs4/8/12/16 and Arcanum11/13/15/17 retain independent grants, not a single initial receipt. No grant inferred from empty row features.

Prepare is inert: no dice, cost, grant or writes. Confirm revalidates actor, campaign, source, revision, entitlement, target, action, timing and operands. Persist choice, injected dice, resources, effects, derived sheet and history atomically in isolated PostgreSQL. Same-ID replay/reload/reconnect returns one receipt without reroll or regrant; conflicting payload, foreign state and stale revision reject. Concurrent confirms commit exactly once. Inject failure after cost/effect/history and prove full rollback. Web and TUI consume the same engine availability and explicit unavailable reasons. Unqualified effects stay unavailable; source metadata and closed planning issues grant nothing. SRD5.2.1 CC-BY4 attribution/adaptation is retained via docs/srd-5.2.1-conformance.md.

### Implement Warlock level 2 (implementation)

Decision #331. One implementation PR; Task Size M.

PDF23–26/70–72/76: Warlock2, exact source row {'invocations/count': 3, 'cantrips/count': 2, 'prepared-spells/count': 3, 'pact-slots/count': 2, 'pact-slots/level': 1}. Only exact adjacent Warlock1->2; no first-character creation; own HP adapter mandatory, slots/Invocation/preparation by Warlock level; XP/live PB by complete canonical285 table at authoritative sum of admitted post-transition class levels. A first-class Warlock lineage does not gain a blanket multiclass CHA13 prerequisite on ordinary advancement; only actual later-entry/multiclass variants consume their own source eligibility. Positive Fighter4/Warlock1->Warlock2 requires total6/14000XP/PB3; Fighter4->Warlock1 total5/6500XP/PB3/later HP. Reject booleans/decimals/zero/nonadjacent/total>20/unqualified combinations. Apply every granted feature, nested option and replacement exactly once; repeated ASIs4/8/12/16 and Arcanum11/13/15/17 retain independent grants, not a single initial receipt. No grant inferred from empty row features.

Prepare is inert: no dice, cost, grant or writes. Confirm revalidates actor, campaign, source, revision, entitlement, target, action, timing and operands. Persist choice, injected dice, resources, effects, derived sheet and history atomically in isolated PostgreSQL. Same-ID replay/reload/reconnect returns one receipt without reroll or regrant; conflicting payload, foreign state and stale revision reject. Concurrent confirms commit exactly once. Inject failure after cost/effect/history and prove full rollback. Web and TUI consume the same engine availability and explicit unavailable reasons. Unqualified effects stay unavailable; source metadata and closed planning issues grant nothing. SRD5.2.1 CC-BY4 attribution/adaptation is retained via docs/srd-5.2.1-conformance.md.

### Implement Warlock level 3 (implementation)

Decision #331. One implementation PR; Task Size M.

PDF23–26/70–72/76: Warlock3, exact source row {'invocations/count': 3, 'cantrips/count': 2, 'prepared-spells/count': 4, 'pact-slots/count': 2, 'pact-slots/level': 2}. Only exact adjacent Warlock2->3; no first-character creation; own HP adapter mandatory, slots/Invocation/preparation by Warlock level; XP/live PB by complete canonical285 table at authoritative sum of admitted post-transition class levels. A first-class Warlock lineage does not gain a blanket multiclass CHA13 prerequisite on ordinary advancement; only actual later-entry/multiclass variants consume their own source eligibility. Positive Fighter4/Warlock1->Warlock2 requires total6/14000XP/PB3; Fighter4->Warlock1 total5/6500XP/PB3/later HP. Reject booleans/decimals/zero/nonadjacent/total>20/unqualified combinations. Apply every granted feature, nested option and replacement exactly once; repeated ASIs4/8/12/16 and Arcanum11/13/15/17 retain independent grants, not a single initial receipt. No grant inferred from empty row features.

Prepare is inert: no dice, cost, grant or writes. Confirm revalidates actor, campaign, source, revision, entitlement, target, action, timing and operands. Persist choice, injected dice, resources, effects, derived sheet and history atomically in isolated PostgreSQL. Same-ID replay/reload/reconnect returns one receipt without reroll or regrant; conflicting payload, foreign state and stale revision reject. Concurrent confirms commit exactly once. Inject failure after cost/effect/history and prove full rollback. Web and TUI consume the same engine availability and explicit unavailable reasons. Unqualified effects stay unavailable; source metadata and closed planning issues grant nothing. SRD5.2.1 CC-BY4 attribution/adaptation is retained via docs/srd-5.2.1-conformance.md.

### Implement Warlock level 4 (implementation)

Decision #331. One implementation PR; Task Size M.

PDF23–26/70–72/76: Warlock4, exact source row {'invocations/count': 3, 'cantrips/count': 3, 'prepared-spells/count': 5, 'pact-slots/count': 2, 'pact-slots/level': 2}. Only exact adjacent Warlock3->4; no first-character creation; own HP adapter mandatory, slots/Invocation/preparation by Warlock level; XP/live PB by complete canonical285 table at authoritative sum of admitted post-transition class levels. A first-class Warlock lineage does not gain a blanket multiclass CHA13 prerequisite on ordinary advancement; only actual later-entry/multiclass variants consume their own source eligibility. Positive Fighter4/Warlock1->Warlock2 requires total6/14000XP/PB3; Fighter4->Warlock1 total5/6500XP/PB3/later HP. Reject booleans/decimals/zero/nonadjacent/total>20/unqualified combinations. Apply every granted feature, nested option and replacement exactly once; repeated ASIs4/8/12/16 and Arcanum11/13/15/17 retain independent grants, not a single initial receipt. No grant inferred from empty row features.

Prepare is inert: no dice, cost, grant or writes. Confirm revalidates actor, campaign, source, revision, entitlement, target, action, timing and operands. Persist choice, injected dice, resources, effects, derived sheet and history atomically in isolated PostgreSQL. Same-ID replay/reload/reconnect returns one receipt without reroll or regrant; conflicting payload, foreign state and stale revision reject. Concurrent confirms commit exactly once. Inject failure after cost/effect/history and prove full rollback. Web and TUI consume the same engine availability and explicit unavailable reasons. Unqualified effects stay unavailable; source metadata and closed planning issues grant nothing. SRD5.2.1 CC-BY4 attribution/adaptation is retained via docs/srd-5.2.1-conformance.md.

### Implement Warlock level 5 (implementation)

Decision #331. One implementation PR; Task Size M.

PDF23–26/70–72/76: Warlock5, exact source row {'invocations/count': 5, 'cantrips/count': 3, 'prepared-spells/count': 6, 'pact-slots/count': 2, 'pact-slots/level': 3}. Only exact adjacent Warlock4->5; no first-character creation; own HP adapter mandatory, slots/Invocation/preparation by Warlock level; XP/live PB by complete canonical285 table at authoritative sum of admitted post-transition class levels. A first-class Warlock lineage does not gain a blanket multiclass CHA13 prerequisite on ordinary advancement; only actual later-entry/multiclass variants consume their own source eligibility. Positive Fighter4/Warlock1->Warlock2 requires total6/14000XP/PB3; Fighter4->Warlock1 total5/6500XP/PB3/later HP. Reject booleans/decimals/zero/nonadjacent/total>20/unqualified combinations. Apply every granted feature, nested option and replacement exactly once; repeated ASIs4/8/12/16 and Arcanum11/13/15/17 retain independent grants, not a single initial receipt. No grant inferred from empty row features.

Prepare is inert: no dice, cost, grant or writes. Confirm revalidates actor, campaign, source, revision, entitlement, target, action, timing and operands. Persist choice, injected dice, resources, effects, derived sheet and history atomically in isolated PostgreSQL. Same-ID replay/reload/reconnect returns one receipt without reroll or regrant; conflicting payload, foreign state and stale revision reject. Concurrent confirms commit exactly once. Inject failure after cost/effect/history and prove full rollback. Web and TUI consume the same engine availability and explicit unavailable reasons. Unqualified effects stay unavailable; source metadata and closed planning issues grant nothing. SRD5.2.1 CC-BY4 attribution/adaptation is retained via docs/srd-5.2.1-conformance.md.

### Implement Warlock level 6 (implementation)

Decision #331. One implementation PR; Task Size M.

PDF23–26/70–72/76: Warlock6, exact source row {'invocations/count': 5, 'cantrips/count': 3, 'prepared-spells/count': 7, 'pact-slots/count': 2, 'pact-slots/level': 3}. Only exact adjacent Warlock5->6; no first-character creation; own HP adapter mandatory, slots/Invocation/preparation by Warlock level; XP/live PB by complete canonical285 table at authoritative sum of admitted post-transition class levels. A first-class Warlock lineage does not gain a blanket multiclass CHA13 prerequisite on ordinary advancement; only actual later-entry/multiclass variants consume their own source eligibility. Positive Fighter4/Warlock1->Warlock2 requires total6/14000XP/PB3; Fighter4->Warlock1 total5/6500XP/PB3/later HP. Reject booleans/decimals/zero/nonadjacent/total>20/unqualified combinations. Apply every granted feature, nested option and replacement exactly once; repeated ASIs4/8/12/16 and Arcanum11/13/15/17 retain independent grants, not a single initial receipt. No grant inferred from empty row features.

Prepare is inert: no dice, cost, grant or writes. Confirm revalidates actor, campaign, source, revision, entitlement, target, action, timing and operands. Persist choice, injected dice, resources, effects, derived sheet and history atomically in isolated PostgreSQL. Same-ID replay/reload/reconnect returns one receipt without reroll or regrant; conflicting payload, foreign state and stale revision reject. Concurrent confirms commit exactly once. Inject failure after cost/effect/history and prove full rollback. Web and TUI consume the same engine availability and explicit unavailable reasons. Unqualified effects stay unavailable; source metadata and closed planning issues grant nothing. SRD5.2.1 CC-BY4 attribution/adaptation is retained via docs/srd-5.2.1-conformance.md.

### Implement Warlock level 7 (implementation)

Decision #331. One implementation PR; Task Size M.

PDF23–26/70–72/76: Warlock7, exact source row {'invocations/count': 6, 'cantrips/count': 3, 'prepared-spells/count': 8, 'pact-slots/count': 2, 'pact-slots/level': 4}. Only exact adjacent Warlock6->7; no first-character creation; own HP adapter mandatory, slots/Invocation/preparation by Warlock level; XP/live PB by complete canonical285 table at authoritative sum of admitted post-transition class levels. A first-class Warlock lineage does not gain a blanket multiclass CHA13 prerequisite on ordinary advancement; only actual later-entry/multiclass variants consume their own source eligibility. Positive Fighter4/Warlock1->Warlock2 requires total6/14000XP/PB3; Fighter4->Warlock1 total5/6500XP/PB3/later HP. Reject booleans/decimals/zero/nonadjacent/total>20/unqualified combinations. Apply every granted feature, nested option and replacement exactly once; repeated ASIs4/8/12/16 and Arcanum11/13/15/17 retain independent grants, not a single initial receipt. No grant inferred from empty row features.

Prepare is inert: no dice, cost, grant or writes. Confirm revalidates actor, campaign, source, revision, entitlement, target, action, timing and operands. Persist choice, injected dice, resources, effects, derived sheet and history atomically in isolated PostgreSQL. Same-ID replay/reload/reconnect returns one receipt without reroll or regrant; conflicting payload, foreign state and stale revision reject. Concurrent confirms commit exactly once. Inject failure after cost/effect/history and prove full rollback. Web and TUI consume the same engine availability and explicit unavailable reasons. Unqualified effects stay unavailable; source metadata and closed planning issues grant nothing. SRD5.2.1 CC-BY4 attribution/adaptation is retained via docs/srd-5.2.1-conformance.md.

### Implement Warlock level 8 (implementation)

Decision #331. One implementation PR; Task Size M.

PDF23–26/70–72/76: Warlock8, exact source row {'invocations/count': 6, 'cantrips/count': 3, 'prepared-spells/count': 9, 'pact-slots/count': 2, 'pact-slots/level': 4}. Only exact adjacent Warlock7->8; no first-character creation; own HP adapter mandatory, slots/Invocation/preparation by Warlock level; XP/live PB by complete canonical285 table at authoritative sum of admitted post-transition class levels. A first-class Warlock lineage does not gain a blanket multiclass CHA13 prerequisite on ordinary advancement; only actual later-entry/multiclass variants consume their own source eligibility. Positive Fighter4/Warlock1->Warlock2 requires total6/14000XP/PB3; Fighter4->Warlock1 total5/6500XP/PB3/later HP. Reject booleans/decimals/zero/nonadjacent/total>20/unqualified combinations. Apply every granted feature, nested option and replacement exactly once; repeated ASIs4/8/12/16 and Arcanum11/13/15/17 retain independent grants, not a single initial receipt. No grant inferred from empty row features.

Prepare is inert: no dice, cost, grant or writes. Confirm revalidates actor, campaign, source, revision, entitlement, target, action, timing and operands. Persist choice, injected dice, resources, effects, derived sheet and history atomically in isolated PostgreSQL. Same-ID replay/reload/reconnect returns one receipt without reroll or regrant; conflicting payload, foreign state and stale revision reject. Concurrent confirms commit exactly once. Inject failure after cost/effect/history and prove full rollback. Web and TUI consume the same engine availability and explicit unavailable reasons. Unqualified effects stay unavailable; source metadata and closed planning issues grant nothing. SRD5.2.1 CC-BY4 attribution/adaptation is retained via docs/srd-5.2.1-conformance.md.

### Implement Warlock level 9 (implementation)

Decision #331. One implementation PR; Task Size M.

PDF23–26/70–72/76: Warlock9, exact source row {'invocations/count': 7, 'cantrips/count': 3, 'prepared-spells/count': 10, 'pact-slots/count': 2, 'pact-slots/level': 5}. Only exact adjacent Warlock8->9; no first-character creation; own HP adapter mandatory, slots/Invocation/preparation by Warlock level; XP/live PB by complete canonical285 table at authoritative sum of admitted post-transition class levels. A first-class Warlock lineage does not gain a blanket multiclass CHA13 prerequisite on ordinary advancement; only actual later-entry/multiclass variants consume their own source eligibility. Positive Fighter4/Warlock1->Warlock2 requires total6/14000XP/PB3; Fighter4->Warlock1 total5/6500XP/PB3/later HP. Reject booleans/decimals/zero/nonadjacent/total>20/unqualified combinations. Apply every granted feature, nested option and replacement exactly once; repeated ASIs4/8/12/16 and Arcanum11/13/15/17 retain independent grants, not a single initial receipt. No grant inferred from empty row features.

Prepare is inert: no dice, cost, grant or writes. Confirm revalidates actor, campaign, source, revision, entitlement, target, action, timing and operands. Persist choice, injected dice, resources, effects, derived sheet and history atomically in isolated PostgreSQL. Same-ID replay/reload/reconnect returns one receipt without reroll or regrant; conflicting payload, foreign state and stale revision reject. Concurrent confirms commit exactly once. Inject failure after cost/effect/history and prove full rollback. Web and TUI consume the same engine availability and explicit unavailable reasons. Unqualified effects stay unavailable; source metadata and closed planning issues grant nothing. SRD5.2.1 CC-BY4 attribution/adaptation is retained via docs/srd-5.2.1-conformance.md.

### Implement Warlock level 10 (implementation)

Decision #331. One implementation PR; Task Size M.

PDF23–26/70–72/76: Warlock10, exact source row {'invocations/count': 7, 'cantrips/count': 4, 'prepared-spells/count': 10, 'pact-slots/count': 2, 'pact-slots/level': 5}. Only exact adjacent Warlock9->10; no first-character creation; own HP adapter mandatory, slots/Invocation/preparation by Warlock level; XP/live PB by complete canonical285 table at authoritative sum of admitted post-transition class levels. A first-class Warlock lineage does not gain a blanket multiclass CHA13 prerequisite on ordinary advancement; only actual later-entry/multiclass variants consume their own source eligibility. Positive Fighter4/Warlock1->Warlock2 requires total6/14000XP/PB3; Fighter4->Warlock1 total5/6500XP/PB3/later HP. Reject booleans/decimals/zero/nonadjacent/total>20/unqualified combinations. Apply every granted feature, nested option and replacement exactly once; repeated ASIs4/8/12/16 and Arcanum11/13/15/17 retain independent grants, not a single initial receipt. No grant inferred from empty row features.

Prepare is inert: no dice, cost, grant or writes. Confirm revalidates actor, campaign, source, revision, entitlement, target, action, timing and operands. Persist choice, injected dice, resources, effects, derived sheet and history atomically in isolated PostgreSQL. Same-ID replay/reload/reconnect returns one receipt without reroll or regrant; conflicting payload, foreign state and stale revision reject. Concurrent confirms commit exactly once. Inject failure after cost/effect/history and prove full rollback. Web and TUI consume the same engine availability and explicit unavailable reasons. Unqualified effects stay unavailable; source metadata and closed planning issues grant nothing. SRD5.2.1 CC-BY4 attribution/adaptation is retained via docs/srd-5.2.1-conformance.md.

### Implement Warlock level 11 (implementation)

Decision #331. One implementation PR; Task Size M.

PDF23–26/70–72/76: Warlock11, exact source row {'invocations/count': 7, 'cantrips/count': 4, 'prepared-spells/count': 11, 'pact-slots/count': 3, 'pact-slots/level': 5, 'mystic-arcanum/spell-level': 6}. Only exact adjacent Warlock10->11; no first-character creation; own HP adapter mandatory, slots/Invocation/preparation by Warlock level; XP/live PB by complete canonical285 table at authoritative sum of admitted post-transition class levels. A first-class Warlock lineage does not gain a blanket multiclass CHA13 prerequisite on ordinary advancement; only actual later-entry/multiclass variants consume their own source eligibility. Positive Fighter4/Warlock1->Warlock2 requires total6/14000XP/PB3; Fighter4->Warlock1 total5/6500XP/PB3/later HP. Reject booleans/decimals/zero/nonadjacent/total>20/unqualified combinations. Apply every granted feature, nested option and replacement exactly once; repeated ASIs4/8/12/16 and Arcanum11/13/15/17 retain independent grants, not a single initial receipt. No grant inferred from empty row features.

Prepare is inert: no dice, cost, grant or writes. Confirm revalidates actor, campaign, source, revision, entitlement, target, action, timing and operands. Persist choice, injected dice, resources, effects, derived sheet and history atomically in isolated PostgreSQL. Same-ID replay/reload/reconnect returns one receipt without reroll or regrant; conflicting payload, foreign state and stale revision reject. Concurrent confirms commit exactly once. Inject failure after cost/effect/history and prove full rollback. Web and TUI consume the same engine availability and explicit unavailable reasons. Unqualified effects stay unavailable; source metadata and closed planning issues grant nothing. SRD5.2.1 CC-BY4 attribution/adaptation is retained via docs/srd-5.2.1-conformance.md.

### Implement Warlock level 12 (implementation)

Decision #331. One implementation PR; Task Size M.

PDF23–26/70–72/76: Warlock12, exact source row {'invocations/count': 8, 'cantrips/count': 4, 'prepared-spells/count': 11, 'pact-slots/count': 3, 'pact-slots/level': 5}. Only exact adjacent Warlock11->12; no first-character creation; own HP adapter mandatory, slots/Invocation/preparation by Warlock level; XP/live PB by complete canonical285 table at authoritative sum of admitted post-transition class levels. A first-class Warlock lineage does not gain a blanket multiclass CHA13 prerequisite on ordinary advancement; only actual later-entry/multiclass variants consume their own source eligibility. Positive Fighter4/Warlock1->Warlock2 requires total6/14000XP/PB3; Fighter4->Warlock1 total5/6500XP/PB3/later HP. Reject booleans/decimals/zero/nonadjacent/total>20/unqualified combinations. Apply every granted feature, nested option and replacement exactly once; repeated ASIs4/8/12/16 and Arcanum11/13/15/17 retain independent grants, not a single initial receipt. No grant inferred from empty row features.

Prepare is inert: no dice, cost, grant or writes. Confirm revalidates actor, campaign, source, revision, entitlement, target, action, timing and operands. Persist choice, injected dice, resources, effects, derived sheet and history atomically in isolated PostgreSQL. Same-ID replay/reload/reconnect returns one receipt without reroll or regrant; conflicting payload, foreign state and stale revision reject. Concurrent confirms commit exactly once. Inject failure after cost/effect/history and prove full rollback. Web and TUI consume the same engine availability and explicit unavailable reasons. Unqualified effects stay unavailable; source metadata and closed planning issues grant nothing. SRD5.2.1 CC-BY4 attribution/adaptation is retained via docs/srd-5.2.1-conformance.md.

### Implement Warlock level 13 (implementation)

Decision #331. One implementation PR; Task Size M.

PDF23–26/70–72/76: Warlock13, exact source row {'invocations/count': 8, 'cantrips/count': 4, 'prepared-spells/count': 12, 'pact-slots/count': 3, 'pact-slots/level': 5, 'mystic-arcanum/spell-level': 7}. Only exact adjacent Warlock12->13; no first-character creation; own HP adapter mandatory, slots/Invocation/preparation by Warlock level; XP/live PB by complete canonical285 table at authoritative sum of admitted post-transition class levels. A first-class Warlock lineage does not gain a blanket multiclass CHA13 prerequisite on ordinary advancement; only actual later-entry/multiclass variants consume their own source eligibility. Positive Fighter4/Warlock1->Warlock2 requires total6/14000XP/PB3; Fighter4->Warlock1 total5/6500XP/PB3/later HP. Reject booleans/decimals/zero/nonadjacent/total>20/unqualified combinations. Apply every granted feature, nested option and replacement exactly once; repeated ASIs4/8/12/16 and Arcanum11/13/15/17 retain independent grants, not a single initial receipt. No grant inferred from empty row features.

Prepare is inert: no dice, cost, grant or writes. Confirm revalidates actor, campaign, source, revision, entitlement, target, action, timing and operands. Persist choice, injected dice, resources, effects, derived sheet and history atomically in isolated PostgreSQL. Same-ID replay/reload/reconnect returns one receipt without reroll or regrant; conflicting payload, foreign state and stale revision reject. Concurrent confirms commit exactly once. Inject failure after cost/effect/history and prove full rollback. Web and TUI consume the same engine availability and explicit unavailable reasons. Unqualified effects stay unavailable; source metadata and closed planning issues grant nothing. SRD5.2.1 CC-BY4 attribution/adaptation is retained via docs/srd-5.2.1-conformance.md.

### Implement Warlock level 14 (implementation)

Decision #331. One implementation PR; Task Size M.

PDF23–26/70–72/76: Warlock14, exact source row {'invocations/count': 8, 'cantrips/count': 4, 'prepared-spells/count': 12, 'pact-slots/count': 3, 'pact-slots/level': 5}. Only exact adjacent Warlock13->14; no first-character creation; own HP adapter mandatory, slots/Invocation/preparation by Warlock level; XP/live PB by complete canonical285 table at authoritative sum of admitted post-transition class levels. A first-class Warlock lineage does not gain a blanket multiclass CHA13 prerequisite on ordinary advancement; only actual later-entry/multiclass variants consume their own source eligibility. Positive Fighter4/Warlock1->Warlock2 requires total6/14000XP/PB3; Fighter4->Warlock1 total5/6500XP/PB3/later HP. Reject booleans/decimals/zero/nonadjacent/total>20/unqualified combinations. Apply every granted feature, nested option and replacement exactly once; repeated ASIs4/8/12/16 and Arcanum11/13/15/17 retain independent grants, not a single initial receipt. No grant inferred from empty row features.

Prepare is inert: no dice, cost, grant or writes. Confirm revalidates actor, campaign, source, revision, entitlement, target, action, timing and operands. Persist choice, injected dice, resources, effects, derived sheet and history atomically in isolated PostgreSQL. Same-ID replay/reload/reconnect returns one receipt without reroll or regrant; conflicting payload, foreign state and stale revision reject. Concurrent confirms commit exactly once. Inject failure after cost/effect/history and prove full rollback. Web and TUI consume the same engine availability and explicit unavailable reasons. Unqualified effects stay unavailable; source metadata and closed planning issues grant nothing. SRD5.2.1 CC-BY4 attribution/adaptation is retained via docs/srd-5.2.1-conformance.md.

### Implement Warlock level 15 (implementation)

Decision #331. One implementation PR; Task Size M.

PDF23–26/70–72/76: Warlock15, exact source row {'invocations/count': 9, 'cantrips/count': 4, 'prepared-spells/count': 13, 'pact-slots/count': 3, 'pact-slots/level': 5, 'mystic-arcanum/spell-level': 8}. Only exact adjacent Warlock14->15; no first-character creation; own HP adapter mandatory, slots/Invocation/preparation by Warlock level; XP/live PB by complete canonical285 table at authoritative sum of admitted post-transition class levels. A first-class Warlock lineage does not gain a blanket multiclass CHA13 prerequisite on ordinary advancement; only actual later-entry/multiclass variants consume their own source eligibility. Positive Fighter4/Warlock1->Warlock2 requires total6/14000XP/PB3; Fighter4->Warlock1 total5/6500XP/PB3/later HP. Reject booleans/decimals/zero/nonadjacent/total>20/unqualified combinations. Apply every granted feature, nested option and replacement exactly once; repeated ASIs4/8/12/16 and Arcanum11/13/15/17 retain independent grants, not a single initial receipt. No grant inferred from empty row features.

Prepare is inert: no dice, cost, grant or writes. Confirm revalidates actor, campaign, source, revision, entitlement, target, action, timing and operands. Persist choice, injected dice, resources, effects, derived sheet and history atomically in isolated PostgreSQL. Same-ID replay/reload/reconnect returns one receipt without reroll or regrant; conflicting payload, foreign state and stale revision reject. Concurrent confirms commit exactly once. Inject failure after cost/effect/history and prove full rollback. Web and TUI consume the same engine availability and explicit unavailable reasons. Unqualified effects stay unavailable; source metadata and closed planning issues grant nothing. SRD5.2.1 CC-BY4 attribution/adaptation is retained via docs/srd-5.2.1-conformance.md.

### Implement Warlock level 16 (implementation)

Decision #331. One implementation PR; Task Size M.

PDF23–26/70–72/76: Warlock16, exact source row {'invocations/count': 9, 'cantrips/count': 4, 'prepared-spells/count': 13, 'pact-slots/count': 3, 'pact-slots/level': 5}. Only exact adjacent Warlock15->16; no first-character creation; own HP adapter mandatory, slots/Invocation/preparation by Warlock level; XP/live PB by complete canonical285 table at authoritative sum of admitted post-transition class levels. A first-class Warlock lineage does not gain a blanket multiclass CHA13 prerequisite on ordinary advancement; only actual later-entry/multiclass variants consume their own source eligibility. Positive Fighter4/Warlock1->Warlock2 requires total6/14000XP/PB3; Fighter4->Warlock1 total5/6500XP/PB3/later HP. Reject booleans/decimals/zero/nonadjacent/total>20/unqualified combinations. Apply every granted feature, nested option and replacement exactly once; repeated ASIs4/8/12/16 and Arcanum11/13/15/17 retain independent grants, not a single initial receipt. No grant inferred from empty row features.

Prepare is inert: no dice, cost, grant or writes. Confirm revalidates actor, campaign, source, revision, entitlement, target, action, timing and operands. Persist choice, injected dice, resources, effects, derived sheet and history atomically in isolated PostgreSQL. Same-ID replay/reload/reconnect returns one receipt without reroll or regrant; conflicting payload, foreign state and stale revision reject. Concurrent confirms commit exactly once. Inject failure after cost/effect/history and prove full rollback. Web and TUI consume the same engine availability and explicit unavailable reasons. Unqualified effects stay unavailable; source metadata and closed planning issues grant nothing. SRD5.2.1 CC-BY4 attribution/adaptation is retained via docs/srd-5.2.1-conformance.md.

### Implement Warlock level 17 (implementation)

Decision #331. One implementation PR; Task Size M.

PDF23–26/70–72/76: Warlock17, exact source row {'invocations/count': 9, 'cantrips/count': 4, 'prepared-spells/count': 14, 'pact-slots/count': 4, 'pact-slots/level': 5, 'mystic-arcanum/spell-level': 9}. Only exact adjacent Warlock16->17; no first-character creation; own HP adapter mandatory, slots/Invocation/preparation by Warlock level; XP/live PB by complete canonical285 table at authoritative sum of admitted post-transition class levels. A first-class Warlock lineage does not gain a blanket multiclass CHA13 prerequisite on ordinary advancement; only actual later-entry/multiclass variants consume their own source eligibility. Positive Fighter4/Warlock1->Warlock2 requires total6/14000XP/PB3; Fighter4->Warlock1 total5/6500XP/PB3/later HP. Reject booleans/decimals/zero/nonadjacent/total>20/unqualified combinations. Apply every granted feature, nested option and replacement exactly once; repeated ASIs4/8/12/16 and Arcanum11/13/15/17 retain independent grants, not a single initial receipt. No grant inferred from empty row features.

Prepare is inert: no dice, cost, grant or writes. Confirm revalidates actor, campaign, source, revision, entitlement, target, action, timing and operands. Persist choice, injected dice, resources, effects, derived sheet and history atomically in isolated PostgreSQL. Same-ID replay/reload/reconnect returns one receipt without reroll or regrant; conflicting payload, foreign state and stale revision reject. Concurrent confirms commit exactly once. Inject failure after cost/effect/history and prove full rollback. Web and TUI consume the same engine availability and explicit unavailable reasons. Unqualified effects stay unavailable; source metadata and closed planning issues grant nothing. SRD5.2.1 CC-BY4 attribution/adaptation is retained via docs/srd-5.2.1-conformance.md.

### Implement Warlock level 18 (implementation)

Decision #331. One implementation PR; Task Size M.

PDF23–26/70–72/76: Warlock18, exact source row {'invocations/count': 10, 'cantrips/count': 4, 'prepared-spells/count': 14, 'pact-slots/count': 4, 'pact-slots/level': 5}. Only exact adjacent Warlock17->18; no first-character creation; own HP adapter mandatory, slots/Invocation/preparation by Warlock level; XP/live PB by complete canonical285 table at authoritative sum of admitted post-transition class levels. A first-class Warlock lineage does not gain a blanket multiclass CHA13 prerequisite on ordinary advancement; only actual later-entry/multiclass variants consume their own source eligibility. Positive Fighter4/Warlock1->Warlock2 requires total6/14000XP/PB3; Fighter4->Warlock1 total5/6500XP/PB3/later HP. Reject booleans/decimals/zero/nonadjacent/total>20/unqualified combinations. Apply every granted feature, nested option and replacement exactly once; repeated ASIs4/8/12/16 and Arcanum11/13/15/17 retain independent grants, not a single initial receipt. No grant inferred from empty row features.

Prepare is inert: no dice, cost, grant or writes. Confirm revalidates actor, campaign, source, revision, entitlement, target, action, timing and operands. Persist choice, injected dice, resources, effects, derived sheet and history atomically in isolated PostgreSQL. Same-ID replay/reload/reconnect returns one receipt without reroll or regrant; conflicting payload, foreign state and stale revision reject. Concurrent confirms commit exactly once. Inject failure after cost/effect/history and prove full rollback. Web and TUI consume the same engine availability and explicit unavailable reasons. Unqualified effects stay unavailable; source metadata and closed planning issues grant nothing. SRD5.2.1 CC-BY4 attribution/adaptation is retained via docs/srd-5.2.1-conformance.md.

### Implement Warlock level 19 (implementation)

Decision #331. One implementation PR; Task Size M.

PDF23–26/70–72/76: Warlock19, exact source row {'invocations/count': 10, 'cantrips/count': 4, 'prepared-spells/count': 15, 'pact-slots/count': 4, 'pact-slots/level': 5}. Only exact adjacent Warlock18->19; no first-character creation; own HP adapter mandatory, slots/Invocation/preparation by Warlock level; XP/live PB by complete canonical285 table at authoritative sum of admitted post-transition class levels. A first-class Warlock lineage does not gain a blanket multiclass CHA13 prerequisite on ordinary advancement; only actual later-entry/multiclass variants consume their own source eligibility. Positive Fighter4/Warlock1->Warlock2 requires total6/14000XP/PB3; Fighter4->Warlock1 total5/6500XP/PB3/later HP. Reject booleans/decimals/zero/nonadjacent/total>20/unqualified combinations. Apply every granted feature, nested option and replacement exactly once; repeated ASIs4/8/12/16 and Arcanum11/13/15/17 retain independent grants, not a single initial receipt. No grant inferred from empty row features.

Prepare is inert: no dice, cost, grant or writes. Confirm revalidates actor, campaign, source, revision, entitlement, target, action, timing and operands. Persist choice, injected dice, resources, effects, derived sheet and history atomically in isolated PostgreSQL. Same-ID replay/reload/reconnect returns one receipt without reroll or regrant; conflicting payload, foreign state and stale revision reject. Concurrent confirms commit exactly once. Inject failure after cost/effect/history and prove full rollback. Web and TUI consume the same engine availability and explicit unavailable reasons. Unqualified effects stay unavailable; source metadata and closed planning issues grant nothing. SRD5.2.1 CC-BY4 attribution/adaptation is retained via docs/srd-5.2.1-conformance.md.

### Implement Warlock level 20 (implementation)

Decision #331. One implementation PR; Task Size M.

PDF23–26/70–72/76: Warlock20, exact source row {'invocations/count': 10, 'cantrips/count': 4, 'prepared-spells/count': 15, 'pact-slots/count': 4, 'pact-slots/level': 5}. Only exact adjacent Warlock19->20; no first-character creation; own HP adapter mandatory, slots/Invocation/preparation by Warlock level; XP/live PB by complete canonical285 table at authoritative sum of admitted post-transition class levels. A first-class Warlock lineage does not gain a blanket multiclass CHA13 prerequisite on ordinary advancement; only actual later-entry/multiclass variants consume their own source eligibility. Positive Fighter4/Warlock1->Warlock2 requires total6/14000XP/PB3; Fighter4->Warlock1 total5/6500XP/PB3/later HP. Reject booleans/decimals/zero/nonadjacent/total>20/unqualified combinations. Apply every granted feature, nested option and replacement exactly once; repeated ASIs4/8/12/16 and Arcanum11/13/15/17 retain independent grants, not a single initial receipt. No grant inferred from empty row features.

Prepare is inert: no dice, cost, grant or writes. Confirm revalidates actor, campaign, source, revision, entitlement, target, action, timing and operands. Persist choice, injected dice, resources, effects, derived sheet and history atomically in isolated PostgreSQL. Same-ID replay/reload/reconnect returns one receipt without reroll or regrant; conflicting payload, foreign state and stale revision reject. Concurrent confirms commit exactly once. Inject failure after cost/effect/history and prove full rollback. Web and TUI consume the same engine availability and explicit unavailable reasons. Unqualified effects stay unavailable; source metadata and closed planning issues grant nothing. SRD5.2.1 CC-BY4 attribution/adaptation is retained via docs/srd-5.2.1-conformance.md.
