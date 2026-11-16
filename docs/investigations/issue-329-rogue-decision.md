# Rogue and Thief decision (#329)

FINAL sealed source/native publication proof PASS at 2026-11-16T20:08:42.485242+00:00; runtime remains unqualified; independent acceptance/CI are ROOT-owned.

This decision supplies a finite implementation roadmap and native qualification paths. It implements no Rogue runtime. Universal source identity/eligibility and typed-dispatch permission are distinct from qualified effect execution. Every unknown/unqualified output remains unavailable. Full production class qualification consumes both #279 and #280 as well as the Rogue feature/level cohorts; broad Rogue and Full Rules remain unavailable.

Pinned English SRD5.2.1: Rogue PDF61–63,Thief64,advancement23,multiclass24–26; consumed source locators appear in each terminal. SHA256 `8974902d109d6e63672d7c490bde9ccf052410503d9cfa768237154fbc5e3d87`. Complete canonical records retain CC-BY-4.0 attribution/adaptation in their source headers and `docs/srd-5.2.1-conformance.md`. The source contracts companion contains direct complete consumed spell/item contracts,not selected quotations. Ownership/fixtures are Edda adaptations.

## Native qualification architecture

Qualification #1093 is under #329 and remains blocked by #329,#279,#280,feature cohort #1094,level cohort #1095 and shared cohort #1096. Cohorts #1094/#1095 sit under #1093; shared #1096 sits under #192. Hide #1137 sits under #50; other shared items sit under #1096. Execution dependency gateway #1097 sits under #329,consumes only #61/#64/#65/#66/#67/#273/#353,and has an exact finite consumer roster. It supplies no runtime prerequisites or implementations. Own HP adapter consumes #968→#429; no new direct #429 edge.

No child is added under #283. One shared cohort edge is added to each of #279/#280,and one qualification edge to #337. Dependency and containment DAGs and all three100-link caps are checked separately. Every preexisting required path is retained; all owned paths remain mandatory after #329 closes. Four parent supplements are local marked artifacts for ROOT publication only. Accepted POST-RANGER prefixes were freshly matched and frozen; ROOT published all four exact sections and the fresh FINAL sealed proof passed.

## All twenty class rows

Class PB is source metadata,not the multiclass XP/PB operand. Canonical live advancement selects the complete row by authoritative post-transition TOTAL admitted class levels. Fighter4/Rogue1→Rogue2 reaches total6,14000XP/PB3; Fighter4→Rogue1 reaches total5,6500XP/PB3. Rogue feature eligibility,dice and resources remain Rogue-class-level. Reject totals>20,invalid/noninteger levels,nonadjacent applications and unqualified class vectors.

| Rogue level | Source PB | Sneak Attack | Applied feature owners | Application |
|---|---|---|---|---|
|1|2|1d6|expertise, sneak-attack, languages, mastery|#1149|
|2|2|1d6|cunning-action|#1150|
|3|2|2d6|subclass, steady-aim, fast-hands-check, fast-hands-object, climber, jumper|#1151|
|4|2|2d6|improvement|#1152|
|5|3|3d6|strike, poison, trip, withdraw, uncanny-dodge|#1153|
|6|3|3d6|expertise|#1154|
|7|3|4d6|evasion, reliable-talent|#1155|
|8|3|4d6|improvement|#1156|
|9|4|5d6|stealth-attack|#1157|
|10|4|5d6|improvement|#1158|
|11|4|6d6|improved-strike|#1159|
|12|4|6d6|improvement|#1160|
|13|5|7d6|attunement, charges, scroll-check, scroll-light, scroll-detect, scroll-aid|#1161|
|14|5|7d6|daze, knock-out, obscure|#1162|
|15|5|8d6|slippery-mind|#1163|
|16|5|8d6|improvement|#1164|
|17|6|9d6|reflexes|#1165|
|18|6|9d6|elusive|#1166|
|19|6|10d6|boon|#1167|
|20|6|10d6|luck|#1168|

Later-class first entry has its own application #1169 and actual own eligibility adapter. First-character branches consume creation traits/skills/equipment,without adding multiclass eligibility as a runtime prerequisite.

## Actual runtime evidence

- `internal/rules/srd521/creation.go`: SupportedBuildForConcept and NewSupportedMartialBuild admit only fixed Fighter/Cleric fixtures; no Rogue.
- `internal/rules/srd521/character.go`: Project class/level guard excludes Rogue; class label is insufficient.
- `internal/rules/srd521/catalogue.go`: ValidateReadiness requires implemented mechanic keys; metadata does not grant runtime.
- `internal/game/advancement_preview.go`: Fixed Fighter1->2 preview; no Rogue HP/eligibility adapter.
- `internal/game/spell_preview.go`: Fixed Cleric cast seams; no Thief scroll admission.
- `internal/engine/action_confirmation.go`: Durable receipt/transaction seam; no inferred Rogue feature execution.
- `internal/combat/resolver.go`: CombatResolver is a strategy interface,not Rogue arithmetic or source-key entitlement.
- `internal/combat/bandit.go`: BanditAttack validates only its explicit NPC profile and trusted spatial facts; it is not generic PLAYER Unarmed Strike779 or Rogue feature support.
- `internal/handlers/character_rules.go`: Shared sheet read route; projection guard still controls readiness.
- `frontend/src/components/start/CharacterGuidedForm.tsx`: Class selector labels do not qualify Rogue creation.
- `frontend/src/components/character/ActionPreviewPanel.tsx`: Shared confirmation UI does not qualify Rogue action options.
- `tui/app.go`: Trusted command path consumes shared engine; no separate Rogue entitlement.

## Exact one-PR publication roster

Every Task is Size M. The full body/source-key/prerequisite/parent/metadata receipt is `issue-329-link-verification.json`; the durable mutation journal is `issue-329-publication-journal.json`. Gate closure is qualification,not behavior implementation.

### #1098: Rogue: hp

PDF23/61: own D8/fixed5 adapter ONLY, qualification968 -> transaction429. First character8+CON; later level max(1,d8+CON) or max(1,5+CON). +2 CON fixed gives7; die1 CON-3 gives1; CON17->18 at total8 adds8 retroactively without healing. Reject D10/fixed6, reroll, first8 on later entry.

Native parent #1094; exact blockers: #58, #968.

### #1099: Rogue: traits

PDF61: DEX primary; DEX/INT saves; Simple and Martial with Finesse OR Light; Thieves’ Tools and Light armor. Reject Heavy armor, martial neither property, later-class saves.

Native parent #1094; exact blockers: #58, #333, #430.

### #1100: Rogue: skills

PDF61: choose exactly4 distinct Acrobatics/Athletics/Deception/Insight/Intimidation/Investigation/Perception/Persuasion/Sleight of Hand/Stealth. Positive first4; negative3/5/duplicate/Survival.

Native parent #1094; exact blockers: #384, #430.

### #1101: Rogue: equipment

PDF61: A Leather Armor,2 Daggers,Shortsword,Shortbow,20 Arrows,Quiver,Thieves’ Tools,Burglar’s Pack,8GP OR B100GP. Positive exact A or B; reject19arrows,9GP,mixed/repeated grants. Inventory grant only; no inferred item effects.

Native parent #1094; exact blockers: #195, #430.

### #1102: Rogue: multiclass

PDF24-25/61: own DEX13 and EVERY admitted current-class primary minimum13. DEX12 fails independently; existing Fighter STR12/DEX12 fails. Gain D8,one Rogue skill,Thieves’ Tools,Light armor and level1 features; no saves,weapon/equipment/four-skill regrant. Positive Fighter4 DEX13 -> Rogue1; reject unqualified vector.

Native parent #1094; exact blockers: #335, #430, #1098.

### #1103: Rogue: expertise

PDF61: exactly2 EXISTING skill proficiencies at1 and2 MORE at6; PB2 expertise adds4 not6, PB3 adds6. Reject tools,unproficient/duplicate skills,third selection,grant6 at5.

Native parent #1094; exact blockers: #61, #384, #430, #1097.

### #1104: Rogue: sneak attack

PDF61-62: once PER TURN creature hit,Finesse OR Ranged weapon; Advantage OR ally within5ft/not Incapacitated AND no Disadvantage. Same weapon damage type; ceil(Rogue level/2)d6. Positive Rogue5 3d6,own turn and different off-turn; reject same turn repeat,ally6ft/incapacitated,disadvantage,unarmed/nonqualifying weapon,noncreature. Inject3/4/5=12 extra; no off-hand conflation.

Native parent #1094; exact blockers: #487, #1097.

### #1105: Rogue: languages

PDF62/18: Thieves’ Cant plus1 from actual standard/rare language tables. Positive Common Sign Language; reject unknown/duplicate/two extra; preserve independent provenance.

Native parent #1094; exact blockers: #384, #430.

### #1106: Rogue: mastery

PDF62: two proficient weapon kinds; completed Long Rest may replace. Dagger/Shortbow positive; reject3,duplicate,unproficient,rest start/interruption. Consume selected property owner489-496,not label.

Native parent #1094; exact blockers: #384, #428, #430, #489, #490, #491, #492, #493, #494, #495, #496.

### #1107: Rogue: cunning action

PDF62: on OWN turn Bonus Action Dash/Disengage/Hide; one choice. Reject off-turn,spent Bonus Action,automatic extra action. Speed30 Dash adds30 movement; consumes780/781 and proposed hide output.

Native parent #1094; exact blockers: #780, #781, #1097, #1137.

### #1108: Rogue: subclass

PDF62/64: choose exactly Thief at Rogue3; no subclass at2,second choice or foreign subclass; feature grant class3/9/13/17 only.

Native parent #1094; exact blockers: #384, #1097.

### #1109: Rogue: steady aim

PDF62: Bonus Action only if NO movement this turn; Advantage NEXT attack on CURRENT turn; Speed0 remainder turn. Positive no move -> attack; reject moved1ft,second attack benefit,next turn carryover.

Native parent #1094; exact blockers: #1097.

### #1110: Rogue: improvement

PDF63/86-88: class4/8/10/12/16 ASI OR qualifying feat. +2one/+1two,max20. Positive18->20; reject21,duplicate,unqualified,wrong class level. NO blanket Spellcasting,encounter or Concentration prerequisite for ordinary selection.399 Spell Recall needs ACTUAL Spellcasting;386 Magic Initiate and scroll access do not supply it. Selected feat requires actual relevant487/427/428/778/779/962 execution,not269/272/270/268 planning closure.

Native parent #1094; exact blockers: #384, #385, #386, #387, #388, #389, #390, #391, #392, #393, #394, #395, #396, #397, #398, #399, #400, #401, #427, #428, #487, #778, #779, #962.

### #1111: Rogue: strike

PDF63: Rogue5 when Sneak Attack damage occurs choose1 effect; remove cost dice BEFORE rolling,resolve AFTER attack damage; DC8+DEX+livePB. DEX3/PB3 DC14,3d6 minus1=2d6. Reject miss,no Sneak Attack,negative dice,unsupported effect.

Native parent #1094; exact blockers: #1097, #1104.

### #1112: Rogue: poison

PDF63: cost1d6,Poisoner’s Kit ON PERSON; CON save failure Poisoned1minute; repeat end each target turn success ends own instance. DC14 fail13/succeed14; reject kit absent,client cure,clearing other poison,expiry61seconds.

Native parent #1094; exact blockers: #504, #727, #1097, #1111.

### #1113: Rogue: trip

PDF63: cost1d6 target Large OR SMALLER,DEX save fail Prone. DC14 roll13 vs14; Huge rejects,no cost on invalid target. Consume488 source-owned Prone.

Native parent #1094; exact blockers: #488, #1097, #1111.

### #1114: Rogue: withdraw

PDF63: cost1d6 immediately AFTER attack move up to HALF Speed without Opportunity Attacks; Speed30 allows15,not16; no teleport or next-turn movement credit.

Native parent #1094; exact blockers: #1097, #1111.

### #1115: Rogue: uncanny dodge

PDF63: Reaction when attacker YOU SEE hits with attack roll; halve attack damage round DOWN. 9->4; reject unseen,save-only damage,no reaction,miss,whole round resistance.

Native parent #1094; exact blockers: #1097.

### #1116: Rogue: evasion

PDF63: DEX save for HALF damage only; success0/fail HALF; cannot use Incapacitated. Damage11 ->0 or5. Reject CON save,no-half effect,incapacitated,ASI ownership.

Native parent #1094; exact blockers: #487, #1097.

### #1117: Rogue: reliable talent

PDF63: ability check using skill OR tool proficiency treats d20<=9 as10. Die9->10 plus own modifiers;10 stays10; reject unproficient,save,attack,change die11.

Native parent #1094; exact blockers: #61, #430, #1097.

### #1118: Rogue: improved strike

PDF63: Rogue11 up to2 strike effects,pay EACH before rolling;6d6 minus1minus1=4d6. Reject3 effects,grant at10,free second effect.

Native parent #1094; exact blockers: #1097, #1111, #1112, #1113, #1114.

### #1119: Rogue: daze

PDF63: Rogue14 cost2d6 CON save; failure NEXT turn only ONE of move/action/Bonus Action.7d6->5d6; reject two categories,next-next turn,wrong save. No invented Stunned condition.

Native parent #1094; exact blockers: #1097, #1111.

### #1120: Rogue: knock out

PDF63: Rogue14 cost6d6 CON fail Unconscious1minute OR ANY damage; repeat save end each turn success ends own instance.7d6->1d6; reject five dice affordability,damage not ending,foreign instance removal.

Native parent #1094; exact blockers: #487, #1097, #1111.

### #1121: Rogue: obscure

PDF63: Rogue14 cost3d6 DEX fail Blinded until END next target turn.7d6->4d6; reject CON save,early start-turn end,foreign condition removal.

Native parent #1094; exact blockers: #960, #963, #1097, #1111.

### #1122: Rogue: slippery mind

PDF63: Rogue15 adds WIS and CHA save proficiency using livePB. PB5 bonus5 not10; reject level14,duplicate stacking,INT replacement.

Native parent #1094; exact blockers: #430, #1097.

### #1123: Rogue: elusive

PDF63: Rogue18 no attack roll has Advantage AGAINST self unless Incapacitated. Positive unseen attacker loses advantage; negative Incapacitated allows normal calculation,save treated as attack.

Native parent #1094; exact blockers: #487, #1097.

### #1124: Rogue: boon

PDF63/87-88: Rogue19 Epic Boon OR other qualified feat,Night Spirit recommendation only. Positive chosen qualified boon; reject auto-Night-Spirit,level18,unqualified feat;399 requires real Spellcasting.

Native parent #1094; exact blockers: #384, #385, #386, #387, #388, #389, #390, #391, #392, #393, #394, #395, #396, #397, #398, #399, #400, #401, #427, #428, #487, #778, #779, #962.

### #1125: Rogue: luck

PDF63: Rogue20 FAILED D20 Test -> roll20; once until completed Short OR Long Rest. Positive failed die2 ->20 (then ordinary modifiers); reject success,unknown outcome,second use,rest start/interrupted rest,critical applied to ability check/save; changed attack roll20 resolves actual source critical rules.

Native parent #1094; exact blockers: #427, #428, #1097.

### #1126: Rogue: fast hands check

PDF64: Thief3 Bonus Action DEX(Sleight of Hand) lock/trap with Thieves’ Tools OR pocket. Check uses supplied trusted object DC15,14fails/15succeeds; reject no tools for lock/trap,unknown object facts,foreign-campaign pocket,spell substitution. Requires separately proposed bounded lock/trap/pocket outputs.

Native parent #1094; exact blockers: #1097, #1108, #1138.

### #1127: Rogue: fast hands object

PDF64: Thief3 Bonus Action Utilize OR Magic to USE ITEM requiring that action; not arbitrary spell Magic action. Positive named qualified item route only; reject wrong activation,unsupported item,spent Bonus Action; exact mundane/magic outputs prerequisite.

Native parent #1094; exact blockers: #1097, #1108, #1139, #1140.

### #1128: Rogue: climber

PDF64: Thief3 Climb Speed=current Speed; Speed30 ->30,Speed0->0. Reject flat30,Swim grant; dynamic modifiers retained.

Native parent #1094; exact blockers: #1097, #1108.

### #1129: Rogue: jumper

PDF64/183/185: Thief3 optional DEX rather than STR jump distance. STR10/DEX16 running long jump16 vs10; high jump3+DEXmod=6 vs3; reject adding both,flight,waiving distance/movement/run-up rules.

Native parent #1094; exact blockers: #1097, #1108.

### #1130: Rogue: stealth attack

PDF64: Thief9 Cunning Strike1d6; Hide’s Invisible survives THIS attack only if end turn behind Three-Quarters OR Total Cover.5d6->4d6; reject Half Cover,wrong source Invisible,no Hide,no end-turn check; save original Hide source and independently end on other triggers.

Native parent #1094; exact blockers: #962, #1097, #1108, #1111, #1137.

### #1131: Rogue: attunement

PDF64/102: Thief13 up to FOUR simultaneous,not five. Positive fourth eligible item; reject fifth,foreign/duplicate item,ignoring item-specific restrictions. No all-item activation.

Native parent #1094; exact blockers: #1097, #1108, #1141.

### #1132: Rogue: charges

PDF64: Thief13 on item PROPERTY expending charges roll1d6;6 skips charge spend,1-5 ordinary spend; preserve action cost and property effect. Positive6 keeps2charges,5 spends1of2; reject reroll/recharge/waived eligibility; generic typed hook over EVERY qualified charged property; Wand of Magic Detection is the deterministic witness,not an eligibility whitelist. Unknown property/effect stays unavailable.

Native parent #1094; exact blockers: #1097, #1108, #1140.

### #1133: Rogue: scroll check

PDF64: Thief13 any Spell Scroll,INT casting ability. Cantrip/level1 reliable; higher INT(Arcana) DC10+SPELL LEVEL; ALL levels2-9 DC12/13/14/15/16/17/18/19; for EACH level test DC-1 fails/DC succeeds with INT(Arcana),including expertise/Reliable Talent if separately granted. Level2 DC12 (11fail disintegrates,12casts),level9 DC19. No Spellcasting grant; no invented class-list restriction. Real spell effect must separately qualify; canonical339spell identity permission is universal; effect dispatch requires exact qualified owner,not all-spell execution. Unknown/unqualified effects unavailable BEFORE rolling/spending; committed qualified higher-scroll failed check disintegrates without effect.

Native parent #1094; exact blockers: #1097, #1108, #1142.

### #1134: Rogue: scroll light

PDF64/144/104-106: compose Thief scroll adapter with actual967 Light,INT; no learned/prepared/Spellcasting grant. Reject other-holder object,Huge,foreign scroll,slot spending or ritual/upcast. Own ongoing Light dismissal NO ACTION when not Incapacitated; Incapacitated alone does not end nonconcentration Light.

Native parent #1094; exact blockers: #967, #1097, #1133.

### #1135: Rogue: scroll detect

PDF64/123/104-106: compose scroll adapter with906 Detect Magic,INT and spell concentration/visibility/barriers only. Reject invisible-object aura disclosure; no slot spend or Ritual from consuming scroll. Prepared Ritual elsewhere normal time+10min,no slot,no Ritual Adept,NOritual upcast; >=1min cast Magic each turn/casting Concentration; interrupted no slot spend.

Native parent #1094; exact blockers: #906, #1097, #1133.

### #1136: Rogue: reflexes

PDF64: Thief17 TWO turns FIRST round every combat,normal Initiative and minus10. Initiative15 ->15and5; reject later round,reload granting third,legacy surprise gate; equal initiative ordering deterministic; each is actual distinct turn with action economy and per-turn features.

Native parent #1094; exact blockers: #1097, #1108.

### #1137: Rogue: hide

PDF183: DC15 DEX(Stealth) while Heavily Obscured or Three-Quarters/Total Cover,out of enemy sight; save check total as discovery DC,Invisible source. Ends on louder-than-whisper sound,enemy finds,attack roll,Verbal spell.14fails/15succeeds; Half Cover alone rejects. Supreme Sneak owns its one attack exception.

Native parent #50; exact blockers: #962, #1097.

### #1138: Rogue: object check

PDF98/94/64: ONLY owned Lock/Thieves’ Tools lock operation,one trusted mechanical trap disarm,and one trusted pocket fixture; retain DEX(Sleight of Hand) Fast Hands exception. DC15 14/15 controls; no arbitrary traps/catalogue inference; unknown DC/ownership unavailable.

Native parent #1096; exact blockers: #195, #430, #1097.

### #1139: Rogue: utilize

PDF99/191: ONLY owned Signal Whistle via Utilize, audible up to600ft. Positive trusted listener600ft; negative601ft,unknown acoustic facts,foreign whistle,spell action,invented consumption. Tinderbox exposed-fuel ignition is already Bonus Action onPDF100 and is NOT this Utilize fixture. Fast Hands supplies Bonus Action exception,not this owner.

Native parent #1096; exact blockers: #195, #1097.

### #1140: Rogue: magic item

PDF251/206 Wand of Magic Detection: holding wand,3charges,1charge casts actual906 Detect Magic with normal casting time; no components/slot. NOattunement and NOlast-charge destruction. Reject0charges,not held,foreign item,other property; separate dawn-recharge leaf owns recovery. Thief conserves charge only,never restores.

Native parent #1096; exact blockers: #195, #906, #1097.

### #1141: Rogue: attune item

PDF102: ONLY four distinct owned source identities: Cloak of Protection PDF216,Gauntlets of Ogre Power PDF223,Ring of Jumping PDF238,Ring of Protection PDF238. Attunement protocol ONLY,not their magical effects; Short Rest focus one item,normal cap3,Thief adapter4. Reject unknown restrictions,5items,foreign possession,rest interruption; reject duplicate Ring of Protection type and using the identification Short Rest as the attunement rest; each of four independent item-effect owners below is mandatory; protocol itself never executes effects. End attunement if prerequisites cease,item>100ft away24hours,death,other creature attunes,or own separate focused Short Rest (unless cursed). Do not infer item restrictions.

Native parent #1096; exact blockers: #195, #427, #1097.

### #1142: Rogue: scroll item

PDF99/244/206: generic canonical Spell Scroll identity and normal consumption protocol over ALL339 canonical spell identities and source legal spell levels0-9,not Light/Detect whitelist. Validate trusted inventory/source/version/spell identity/level; ordinary class-list permission or qualified Thief override. Typed effect dispatch requires actual qualified source-bound effect owner; unknown/unqualified effect unavailable before cost/check,without treating permission as execution. Normal cast time,no material components; interrupted casting does NOT consume scroll; completed cast disintegrates. Spell Scroll fixed DC/attack table by spell level:0-2=13/+5,3-4=15/+7,5-6=17/+9,7-8=18/+10,9=19/+11; Thief suppliesINT as source casting ability,does not turn Spell Scroll fixed DC/attack intoINT+PB. Default higher-than-normal-castable spell check uses spellcasting ability DC10+level; ordinary failure removes the spell from the scroll without other effect (do not invent physical destruction); Thief failure instead disintegrates its scroll perPDF64;ordinary failure removes spell with no effect (do not invent physical destruction),Thief override failure disintegrates scroll; Thief exact override ownsINT(Arcana),cantrip/1 reliable,2-9check. No slots,no ritual/upcast,learned spells or actual Spellcasting grant. Independent positive/negative each level0-9 and every source-identity family; unauthorized/unknown identity fails. Successful/failed qualified check,consumption,effect,history atomic.

Native parent #1096; exact blockers: #195, #1097.

### #1143: Rogue: scroll aid

PDF64/107/244/206: Thief13 level2 Aid scroll INT(Arcana)DC12;11fails and disintegrates NOeffect,12succeeds invokes actual614. Up to3creatures30ft,current/maxHP+5 for8hours; HP10/max20->15/25;31ft/fourtargets reject before costs; expiry removes own contribution,no recast/replay heal. No slot/material expenditure,no Cleric grant,no inferred upcast. Qualified effect required before check; rollback restores scroll/check/effect/history; same receipt replay never rerolls.

Native parent #1094; exact blockers: #614, #1097, #1133.

### #1144: Rogue: cloak protection

PDF216/102-103: ONLY Cloak of Protection,requires own valid attunement AND WORN; +1 AC and ALL saves from THIS source. AC14->15,save+2->+3; remove cloak or lose attunement removes own1 only; separate RingProtection can stack different item contributions,duplicate same item type cannot. Reject merely held/stowed/foreign/unattuned,attack/check bonus,permanent base-stat mutation. Recompute source-linked projection on equip/unequip/end events; no action cost invented for passive bonus.

Native parent #1096; exact blockers: #58, #195, #430, #1097, #1141.

### #1145: Rogue: gauntlets ogre

PDF223/102-103: ONLY Gauntlets of Ogre Power,attuned AND worn: STR becomes19 IF underlying STR<19. Base10->19/mod4;underlying19->19,20->20; removal restores actual base,not captured stale10. Reject +19 or lowering20 to19,unattuned/held,DEX substitution; recompute derived STR attacks/checks/saves/carrying through qualified seams; no permanent ability/HP grant.

Native parent #1096; exact blockers: #58, #195, #430, #1097, #1141.

### #1146: Rogue: ring jumping

PDF238/143/206: ONLY Ring of Jumping,attuned AND worn,casts actual1064 Jump on SELF ONLY. Normal Bonus Action,1minute,once each target turn jump<=30ft spending10ft movement;20ftjump costs10,second same-turn jump fails. No charges/slots/components,not old tripled-distance. Reject other target,unworn/unattuned,free movement; item removal prevents NEWcasts but ongoing spell ends only source-valid expiry/dismissal,not invented equip leash. Casting eligibility comes from item,not Ranger membership or Spellcasting grant.

Native parent #1096; exact blockers: #195, #487, #1064, #1097, #1141.

### #1147: Rogue: ring protection

PDF238/102-103: ONLY Ring of Protection,own attunement AND WORN:+1AC and ALL saves. AC14->15,save+2->+3; distinct Cloak yields16/+4. Reject held/unattuned,secondRingProtection attunement,attack/check bonus; remove own contribution only on unequip/attunement loss,retain Cloak. Typed provenance prevents permanent base-stat changes.

Native parent #1096; exact blockers: #58, #195, #430, #1097, #1141.

### #1148: Rogue: wand dawn

PDF251/206: ONLY Wand of Magic Detection,3charge cap; at EACH authoritative dawn regain1d3 EXPENDED charges. Start1 +die3 =>3,not4;start3 remains3,no extra charge;start0 +die2=>2. Noattunement,nolast-charge destruction. Reject rest-start/LongRest recovery,client clock,duplicate dawn,reroll,foreign wand. Plane without dawn requires trusted explicit GM recharge event,never fabricated time. Recorded event/die/revision/recharge atomic; replay/reload/concurrent event exactly once.

Native parent #1096; exact blockers: #195, #1097, #1140.

### #1149: Rogue: level 1

PDF62 row1: first-character Rogue1 creation ONLY. Exact progression={"feature_keys": [], "feature_refs": ["srd-5.2.1:character/feature/rogue/expertise", "srd-5.2.1:character/feature/rogue/sneak-attack", "srd-5.2.1:character/feature/rogue/thieves-cant", "srd-5.2.1:character/feature/rogue/weapon-mastery"], "proficiency_bonus": 2}. XP/livePB from complete canonical advancement by post-transition TOTAL admitted class levels,NOT class-table PB. Fighter4/Rogue1->Rogue2 total6=14000XP/PB3; Fighter4->Rogue1 total5=6500/PB3. Reject class-indexed300XP/PB2,noninteger/invalid levels,total21,skip,unqualified combination,duplicate grant; class dice/resources remain Rogue-level. Later-class variant MUST consume own multiclass eligibility; first1 consumes traits/skills/equipment. Independent feature ownership,not ASI fallback. Exact pure-class witness: Rogue0->Rogue1,post-total1,XP0,livePB2; reject XP-1 and any forged livePB; CON+2 later fixedHP+7 (first-character1 HP10).

Native parent #1095; exact blockers: #1098, #1099, #1100, #1101, #1102, #1103, #1104, #1105, #1106.

### #1150: Rogue: level 2

PDF62 row2: exact adjacent Rogue1->2 ONLY. Exact progression={"feature_keys": [], "feature_refs": ["srd-5.2.1:character/feature/rogue/cunning-action"], "proficiency_bonus": 2}. XP/livePB from complete canonical advancement by post-transition TOTAL admitted class levels,NOT class-table PB. Fighter4/Rogue1->Rogue2 total6=14000XP/PB3; Fighter4->Rogue1 total5=6500/PB3. Reject class-indexed300XP/PB2,noninteger/invalid levels,total21,skip,unqualified combination,duplicate grant; class dice/resources remain Rogue-level. Later-class variant MUST consume own multiclass eligibility; NO creation branch or creation trait/equipment regrant. Independent feature ownership,not ASI fallback. Exact pure-class witness: Rogue1->Rogue2,post-total2,XP300,livePB2; reject XP299 and any forged livePB; CON+2 later fixedHP+7 (first-character1 HP10). Later-class vector witness Fighter4/Rogue1->Rogue2:post-total6,XP14000,livePB3; reject XP13999,missing own admission and class-row XP/PB substitution.

Native parent #1095; exact blockers: #1098, #1102, #1107, #1149.

### #1151: Rogue: level 3

PDF62 row3: exact adjacent Rogue2->3 ONLY. Exact progression={"feature_keys": [], "feature_refs": ["srd-5.2.1:character/feature/rogue/rogue-subclass", "srd-5.2.1:character/feature/rogue/steady-aim"], "proficiency_bonus": 2}. XP/livePB from complete canonical advancement by post-transition TOTAL admitted class levels,NOT class-table PB. Fighter4/Rogue1->Rogue2 total6=14000XP/PB3; Fighter4->Rogue1 total5=6500/PB3. Reject class-indexed300XP/PB2,noninteger/invalid levels,total21,skip,unqualified combination,duplicate grant; class dice/resources remain Rogue-level. Later-class variant MUST consume own multiclass eligibility; NO creation branch or creation trait/equipment regrant. Independent feature ownership,not ASI fallback. Exact pure-class witness: Rogue2->Rogue3,post-total3,XP900,livePB2; reject XP899 and any forged livePB; CON+2 later fixedHP+7 (first-character1 HP10). Later-class vector witness Fighter4/Rogue2->Rogue3:post-total7,XP23000,livePB3; reject XP22999,missing own admission and class-row XP/PB substitution.

Native parent #1095; exact blockers: #1098, #1102, #1108, #1109, #1126, #1127, #1128, #1129, #1150.

### #1152: Rogue: level 4

PDF62 row4: exact adjacent Rogue3->4 ONLY. Exact progression={"feature_keys": [], "feature_refs": ["srd-5.2.1:character/feature/rogue/ability-score-improvement"], "proficiency_bonus": 2}. XP/livePB from complete canonical advancement by post-transition TOTAL admitted class levels,NOT class-table PB. Fighter4/Rogue1->Rogue2 total6=14000XP/PB3; Fighter4->Rogue1 total5=6500/PB3. Reject class-indexed300XP/PB2,noninteger/invalid levels,total21,skip,unqualified combination,duplicate grant; class dice/resources remain Rogue-level. Later-class variant MUST consume own multiclass eligibility; NO creation branch or creation trait/equipment regrant. Independent feature ownership,not ASI fallback. Exact pure-class witness: Rogue3->Rogue4,post-total4,XP2700,livePB2; reject XP2699 and any forged livePB; CON+2 later fixedHP+7 (first-character1 HP10). Later-class vector witness Fighter4/Rogue3->Rogue4:post-total8,XP34000,livePB3; reject XP33999,missing own admission and class-row XP/PB substitution.

Native parent #1095; exact blockers: #1098, #1102, #1110, #1151.

### #1153: Rogue: level 5

PDF62 row5: exact adjacent Rogue4->5 ONLY. Exact progression={"feature_keys": [], "feature_refs": ["srd-5.2.1:character/feature/rogue/cunning-strike", "srd-5.2.1:character/feature/rogue/uncanny-dodge"], "proficiency_bonus": 3}. XP/livePB from complete canonical advancement by post-transition TOTAL admitted class levels,NOT class-table PB. Fighter4/Rogue1->Rogue2 total6=14000XP/PB3; Fighter4->Rogue1 total5=6500/PB3. Reject class-indexed300XP/PB2,noninteger/invalid levels,total21,skip,unqualified combination,duplicate grant; class dice/resources remain Rogue-level. Later-class variant MUST consume own multiclass eligibility; NO creation branch or creation trait/equipment regrant. Independent feature ownership,not ASI fallback. Exact pure-class witness: Rogue4->Rogue5,post-total5,XP6500,livePB3; reject XP6499 and any forged livePB; CON+2 later fixedHP+7 (first-character1 HP10). Later-class vector witness Fighter4/Rogue4->Rogue5:post-total9,XP48000,livePB4; reject XP47999,missing own admission and class-row XP/PB substitution.

Native parent #1095; exact blockers: #1098, #1102, #1111, #1112, #1113, #1114, #1115, #1152.

### #1154: Rogue: level 6

PDF62 row6: exact adjacent Rogue5->6 ONLY. Exact progression={"feature_keys": [], "feature_refs": ["srd-5.2.1:character/feature/rogue/expertise"], "proficiency_bonus": 3}. XP/livePB from complete canonical advancement by post-transition TOTAL admitted class levels,NOT class-table PB. Fighter4/Rogue1->Rogue2 total6=14000XP/PB3; Fighter4->Rogue1 total5=6500/PB3. Reject class-indexed300XP/PB2,noninteger/invalid levels,total21,skip,unqualified combination,duplicate grant; class dice/resources remain Rogue-level. Later-class variant MUST consume own multiclass eligibility; NO creation branch or creation trait/equipment regrant. Independent feature ownership,not ASI fallback. Exact pure-class witness: Rogue5->Rogue6,post-total6,XP14000,livePB3; reject XP13999 and any forged livePB; CON+2 later fixedHP+7 (first-character1 HP10). Later-class vector witness Fighter4/Rogue5->Rogue6:post-total10,XP64000,livePB4; reject XP63999,missing own admission and class-row XP/PB substitution.

Native parent #1095; exact blockers: #1098, #1102, #1103, #1153.

### #1155: Rogue: level 7

PDF62 row7: exact adjacent Rogue6->7 ONLY. Exact progression={"feature_keys": [], "feature_refs": ["srd-5.2.1:character/feature/rogue/evasion", "srd-5.2.1:character/feature/rogue/reliable-talent"], "proficiency_bonus": 3}. XP/livePB from complete canonical advancement by post-transition TOTAL admitted class levels,NOT class-table PB. Fighter4/Rogue1->Rogue2 total6=14000XP/PB3; Fighter4->Rogue1 total5=6500/PB3. Reject class-indexed300XP/PB2,noninteger/invalid levels,total21,skip,unqualified combination,duplicate grant; class dice/resources remain Rogue-level. Later-class variant MUST consume own multiclass eligibility; NO creation branch or creation trait/equipment regrant. Independent feature ownership,not ASI fallback. Exact pure-class witness: Rogue6->Rogue7,post-total7,XP23000,livePB3; reject XP22999 and any forged livePB; CON+2 later fixedHP+7 (first-character1 HP10). Later-class vector witness Fighter4/Rogue6->Rogue7:post-total11,XP85000,livePB4; reject XP84999,missing own admission and class-row XP/PB substitution.

Native parent #1095; exact blockers: #1098, #1102, #1116, #1117, #1154.

### #1156: Rogue: level 8

PDF62 row8: exact adjacent Rogue7->8 ONLY. Exact progression={"feature_keys": [], "feature_refs": ["srd-5.2.1:character/feature/rogue/ability-score-improvement"], "proficiency_bonus": 3}. XP/livePB from complete canonical advancement by post-transition TOTAL admitted class levels,NOT class-table PB. Fighter4/Rogue1->Rogue2 total6=14000XP/PB3; Fighter4->Rogue1 total5=6500/PB3. Reject class-indexed300XP/PB2,noninteger/invalid levels,total21,skip,unqualified combination,duplicate grant; class dice/resources remain Rogue-level. Later-class variant MUST consume own multiclass eligibility; NO creation branch or creation trait/equipment regrant. Independent feature ownership,not ASI fallback. Exact pure-class witness: Rogue7->Rogue8,post-total8,XP34000,livePB3; reject XP33999 and any forged livePB; CON+2 later fixedHP+7 (first-character1 HP10). Later-class vector witness Fighter4/Rogue7->Rogue8:post-total12,XP100000,livePB4; reject XP99999,missing own admission and class-row XP/PB substitution.

Native parent #1095; exact blockers: #1098, #1102, #1110, #1155.

### #1157: Rogue: level 9

PDF62 row9: exact adjacent Rogue8->9 ONLY. Exact progression={"feature_keys": ["srd-5.2.1:mechanic/rogue/subclass-feature"], "feature_refs": [], "proficiency_bonus": 4}. XP/livePB from complete canonical advancement by post-transition TOTAL admitted class levels,NOT class-table PB. Fighter4/Rogue1->Rogue2 total6=14000XP/PB3; Fighter4->Rogue1 total5=6500/PB3. Reject class-indexed300XP/PB2,noninteger/invalid levels,total21,skip,unqualified combination,duplicate grant; class dice/resources remain Rogue-level. Later-class variant MUST consume own multiclass eligibility; NO creation branch or creation trait/equipment regrant. Independent feature ownership,not ASI fallback. Exact pure-class witness: Rogue8->Rogue9,post-total9,XP48000,livePB4; reject XP47999 and any forged livePB; CON+2 later fixedHP+7 (first-character1 HP10). Later-class vector witness Fighter4/Rogue8->Rogue9:post-total13,XP120000,livePB5; reject XP119999,missing own admission and class-row XP/PB substitution.

Native parent #1095; exact blockers: #1098, #1102, #1130, #1156.

### #1158: Rogue: level 10

PDF62 row10: exact adjacent Rogue9->10 ONLY. Exact progression={"feature_keys": [], "feature_refs": ["srd-5.2.1:character/feature/rogue/ability-score-improvement"], "proficiency_bonus": 4}. XP/livePB from complete canonical advancement by post-transition TOTAL admitted class levels,NOT class-table PB. Fighter4/Rogue1->Rogue2 total6=14000XP/PB3; Fighter4->Rogue1 total5=6500/PB3. Reject class-indexed300XP/PB2,noninteger/invalid levels,total21,skip,unqualified combination,duplicate grant; class dice/resources remain Rogue-level. Later-class variant MUST consume own multiclass eligibility; NO creation branch or creation trait/equipment regrant. Independent feature ownership,not ASI fallback. Exact pure-class witness: Rogue9->Rogue10,post-total10,XP64000,livePB4; reject XP63999 and any forged livePB; CON+2 later fixedHP+7 (first-character1 HP10). Later-class vector witness Fighter4/Rogue9->Rogue10:post-total14,XP140000,livePB5; reject XP139999,missing own admission and class-row XP/PB substitution.

Native parent #1095; exact blockers: #1098, #1102, #1110, #1157.

### #1159: Rogue: level 11

PDF62 row11: exact adjacent Rogue10->11 ONLY. Exact progression={"feature_keys": [], "feature_refs": ["srd-5.2.1:character/feature/rogue/improved-cunning-strike"], "proficiency_bonus": 4}. XP/livePB from complete canonical advancement by post-transition TOTAL admitted class levels,NOT class-table PB. Fighter4/Rogue1->Rogue2 total6=14000XP/PB3; Fighter4->Rogue1 total5=6500/PB3. Reject class-indexed300XP/PB2,noninteger/invalid levels,total21,skip,unqualified combination,duplicate grant; class dice/resources remain Rogue-level. Later-class variant MUST consume own multiclass eligibility; NO creation branch or creation trait/equipment regrant. Independent feature ownership,not ASI fallback. Exact pure-class witness: Rogue10->Rogue11,post-total11,XP85000,livePB4; reject XP84999 and any forged livePB; CON+2 later fixedHP+7 (first-character1 HP10). Later-class vector witness Fighter4/Rogue10->Rogue11:post-total15,XP165000,livePB5; reject XP164999,missing own admission and class-row XP/PB substitution.

Native parent #1095; exact blockers: #1098, #1102, #1118, #1158.

### #1160: Rogue: level 12

PDF62 row12: exact adjacent Rogue11->12 ONLY. Exact progression={"feature_keys": [], "feature_refs": ["srd-5.2.1:character/feature/rogue/ability-score-improvement"], "proficiency_bonus": 4}. XP/livePB from complete canonical advancement by post-transition TOTAL admitted class levels,NOT class-table PB. Fighter4/Rogue1->Rogue2 total6=14000XP/PB3; Fighter4->Rogue1 total5=6500/PB3. Reject class-indexed300XP/PB2,noninteger/invalid levels,total21,skip,unqualified combination,duplicate grant; class dice/resources remain Rogue-level. Later-class variant MUST consume own multiclass eligibility; NO creation branch or creation trait/equipment regrant. Independent feature ownership,not ASI fallback. Exact pure-class witness: Rogue11->Rogue12,post-total12,XP100000,livePB4; reject XP99999 and any forged livePB; CON+2 later fixedHP+7 (first-character1 HP10). Later-class vector witness Fighter4/Rogue11->Rogue12:post-total16,XP195000,livePB5; reject XP194999,missing own admission and class-row XP/PB substitution.

Native parent #1095; exact blockers: #1098, #1102, #1110, #1159.

### #1161: Rogue: level 13

PDF62 row13: exact adjacent Rogue12->13 ONLY. Exact progression={"feature_keys": ["srd-5.2.1:mechanic/rogue/subclass-feature"], "feature_refs": [], "proficiency_bonus": 5}. XP/livePB from complete canonical advancement by post-transition TOTAL admitted class levels,NOT class-table PB. Fighter4/Rogue1->Rogue2 total6=14000XP/PB3; Fighter4->Rogue1 total5=6500/PB3. Reject class-indexed300XP/PB2,noninteger/invalid levels,total21,skip,unqualified combination,duplicate grant; class dice/resources remain Rogue-level. Later-class variant MUST consume own multiclass eligibility; NO creation branch or creation trait/equipment regrant. Independent feature ownership,not ASI fallback. Exact pure-class witness: Rogue12->Rogue13,post-total13,XP120000,livePB5; reject XP119999 and any forged livePB; CON+2 later fixedHP+7 (first-character1 HP10). Later-class vector witness Fighter4/Rogue12->Rogue13:post-total17,XP225000,livePB6; reject XP224999,missing own admission and class-row XP/PB substitution.

Native parent #1095; exact blockers: #1098, #1102, #1131, #1132, #1133, #1134, #1135, #1143, #1160.

### #1162: Rogue: level 14

PDF62 row14: exact adjacent Rogue13->14 ONLY. Exact progression={"feature_keys": [], "feature_refs": ["srd-5.2.1:character/feature/rogue/devious-strikes"], "proficiency_bonus": 5}. XP/livePB from complete canonical advancement by post-transition TOTAL admitted class levels,NOT class-table PB. Fighter4/Rogue1->Rogue2 total6=14000XP/PB3; Fighter4->Rogue1 total5=6500/PB3. Reject class-indexed300XP/PB2,noninteger/invalid levels,total21,skip,unqualified combination,duplicate grant; class dice/resources remain Rogue-level. Later-class variant MUST consume own multiclass eligibility; NO creation branch or creation trait/equipment regrant. Independent feature ownership,not ASI fallback. Exact pure-class witness: Rogue13->Rogue14,post-total14,XP140000,livePB5; reject XP139999 and any forged livePB; CON+2 later fixedHP+7 (first-character1 HP10). Later-class vector witness Fighter4/Rogue13->Rogue14:post-total18,XP265000,livePB6; reject XP264999,missing own admission and class-row XP/PB substitution.

Native parent #1095; exact blockers: #1098, #1102, #1119, #1120, #1121, #1161.

### #1163: Rogue: level 15

PDF62 row15: exact adjacent Rogue14->15 ONLY. Exact progression={"feature_keys": [], "feature_refs": ["srd-5.2.1:character/feature/rogue/slippery-mind"], "proficiency_bonus": 5}. XP/livePB from complete canonical advancement by post-transition TOTAL admitted class levels,NOT class-table PB. Fighter4/Rogue1->Rogue2 total6=14000XP/PB3; Fighter4->Rogue1 total5=6500/PB3. Reject class-indexed300XP/PB2,noninteger/invalid levels,total21,skip,unqualified combination,duplicate grant; class dice/resources remain Rogue-level. Later-class variant MUST consume own multiclass eligibility; NO creation branch or creation trait/equipment regrant. Independent feature ownership,not ASI fallback. Exact pure-class witness: Rogue14->Rogue15,post-total15,XP165000,livePB5; reject XP164999 and any forged livePB; CON+2 later fixedHP+7 (first-character1 HP10). Later-class vector witness Fighter4/Rogue14->Rogue15:post-total19,XP305000,livePB6; reject XP304999,missing own admission and class-row XP/PB substitution.

Native parent #1095; exact blockers: #1098, #1102, #1122, #1162.

### #1164: Rogue: level 16

PDF62 row16: exact adjacent Rogue15->16 ONLY. Exact progression={"feature_keys": [], "feature_refs": ["srd-5.2.1:character/feature/rogue/ability-score-improvement"], "proficiency_bonus": 5}. XP/livePB from complete canonical advancement by post-transition TOTAL admitted class levels,NOT class-table PB. Fighter4/Rogue1->Rogue2 total6=14000XP/PB3; Fighter4->Rogue1 total5=6500/PB3. Reject class-indexed300XP/PB2,noninteger/invalid levels,total21,skip,unqualified combination,duplicate grant; class dice/resources remain Rogue-level. Later-class variant MUST consume own multiclass eligibility; NO creation branch or creation trait/equipment regrant. Independent feature ownership,not ASI fallback. Exact pure-class witness: Rogue15->Rogue16,post-total16,XP195000,livePB5; reject XP194999 and any forged livePB; CON+2 later fixedHP+7 (first-character1 HP10). Later-class vector witness Fighter4/Rogue15->Rogue16:post-total20,XP355000,livePB6; reject XP354999,missing own admission and class-row XP/PB substitution.

Native parent #1095; exact blockers: #1098, #1102, #1110, #1163.

### #1165: Rogue: level 17

PDF62 row17: exact adjacent Rogue16->17 ONLY. Exact progression={"feature_keys": ["srd-5.2.1:mechanic/rogue/subclass-feature"], "feature_refs": [], "proficiency_bonus": 6}. XP/livePB from complete canonical advancement by post-transition TOTAL admitted class levels,NOT class-table PB. Fighter4/Rogue1->Rogue2 total6=14000XP/PB3; Fighter4->Rogue1 total5=6500/PB3. Reject class-indexed300XP/PB2,noninteger/invalid levels,total21,skip,unqualified combination,duplicate grant; class dice/resources remain Rogue-level. Later-class variant MUST consume own multiclass eligibility; NO creation branch or creation trait/equipment regrant. Independent feature ownership,not ASI fallback. Exact pure-class witness: Rogue16->Rogue17,post-total17,XP225000,livePB6; reject XP224999 and any forged livePB; CON+2 later fixedHP+7 (first-character1 HP10). Later-class vector witness Fighter1/Rogue16->Rogue17:post-total18,XP265000,livePB6; reject XP264999,missing own admission and class-row XP/PB substitution.

Native parent #1095; exact blockers: #1098, #1102, #1136, #1164.

### #1166: Rogue: level 18

PDF62 row18: exact adjacent Rogue17->18 ONLY. Exact progression={"feature_keys": [], "feature_refs": ["srd-5.2.1:character/feature/rogue/elusive"], "proficiency_bonus": 6}. XP/livePB from complete canonical advancement by post-transition TOTAL admitted class levels,NOT class-table PB. Fighter4/Rogue1->Rogue2 total6=14000XP/PB3; Fighter4->Rogue1 total5=6500/PB3. Reject class-indexed300XP/PB2,noninteger/invalid levels,total21,skip,unqualified combination,duplicate grant; class dice/resources remain Rogue-level. Later-class variant MUST consume own multiclass eligibility; NO creation branch or creation trait/equipment regrant. Independent feature ownership,not ASI fallback. Exact pure-class witness: Rogue17->Rogue18,post-total18,XP265000,livePB6; reject XP264999 and any forged livePB; CON+2 later fixedHP+7 (first-character1 HP10). Later-class vector witness Fighter1/Rogue17->Rogue18:post-total19,XP305000,livePB6; reject XP304999,missing own admission and class-row XP/PB substitution.

Native parent #1095; exact blockers: #1098, #1102, #1123, #1165.

### #1167: Rogue: level 19

PDF62 row19: exact adjacent Rogue18->19 ONLY. Exact progression={"feature_keys": [], "feature_refs": ["srd-5.2.1:character/feature/rogue/epic-boon"], "proficiency_bonus": 6}. XP/livePB from complete canonical advancement by post-transition TOTAL admitted class levels,NOT class-table PB. Fighter4/Rogue1->Rogue2 total6=14000XP/PB3; Fighter4->Rogue1 total5=6500/PB3. Reject class-indexed300XP/PB2,noninteger/invalid levels,total21,skip,unqualified combination,duplicate grant; class dice/resources remain Rogue-level. Later-class variant MUST consume own multiclass eligibility; NO creation branch or creation trait/equipment regrant. Independent feature ownership,not ASI fallback. Exact pure-class witness: Rogue18->Rogue19,post-total19,XP305000,livePB6; reject XP304999 and any forged livePB; CON+2 later fixedHP+7 (first-character1 HP10). Later-class vector witness Fighter1/Rogue18->Rogue19:post-total20,XP355000,livePB6; reject XP354999,missing own admission and class-row XP/PB substitution.

Native parent #1095; exact blockers: #1098, #1102, #1124, #1166.

### #1168: Rogue: level 20

PDF62 row20: exact adjacent Rogue19->20 ONLY. Exact progression={"feature_keys": [], "feature_refs": ["srd-5.2.1:character/feature/rogue/stroke-of-luck"], "proficiency_bonus": 6}. XP/livePB from complete canonical advancement by post-transition TOTAL admitted class levels,NOT class-table PB. Fighter4/Rogue1->Rogue2 total6=14000XP/PB3; Fighter4->Rogue1 total5=6500/PB3. Reject class-indexed300XP/PB2,noninteger/invalid levels,total21,skip,unqualified combination,duplicate grant; class dice/resources remain Rogue-level. Later-class variant MUST consume own multiclass eligibility; NO creation branch or creation trait/equipment regrant. Independent feature ownership,not ASI fallback. Exact pure-class witness: Rogue19->Rogue20,post-total20,XP355000,livePB6; reject XP354999 and any forged livePB; CON+2 later fixedHP+7 (first-character1 HP10). Any other admitted class plusRogue20 exceeds20 and MUSTreject; no valid later-class20 vector exists.

Native parent #1095; exact blockers: #1098, #1102, #1125, #1167.

### #1169: Rogue: later 1

PDF24-25/61: Fighter4->Rogue1 TOTAL5 6500XP/PB3,own reduced traits and all level1 features,HPmax(1,5+CON) not8+CON; reject0XP/PB2,saves/weapons/equipment regrant,missing own eligibility,total21.

Native parent #1095; exact blockers: #1098, #1102, #1103, #1104, #1105, #1106.

### #1093: Qualify finite Rogue qualification contracts

Finite prerequisite qualification ONLY; no new runtime implementation or eligibility requirement. Every listed blocker remains mandatory after329closure. Exact consumer roster: . Full production class qualification also requires actual279spell/280item qualifications; universal scroll identity permission alone never qualifies all effects. Keep own feature and level cohorts mandatory.

Native parent #329; exact blockers: #279, #280, #329, #1094, #1095, #1096.

### #1094: Qualify finite Rogue features contracts

Finite prerequisite qualification ONLY; no new runtime implementation or eligibility requirement. Every listed blocker remains mandatory after329closure. Exact consumer roster: hp, traits, skills, equipment, multiclass, expertise, sneak-attack, languages, mastery, cunning-action, subclass, steady-aim, improvement, strike, poison, trip, withdraw, uncanny-dodge, evasion, reliable-talent, improved-strike, daze, knock-out, obscure, slippery-mind, elusive, boon, luck, fast-hands-check, fast-hands-object, climber, jumper, stealth-attack, attunement, charges, scroll-check, scroll-light, scroll-detect, reflexes, scroll-aid.

Native parent #1093; exact blockers: #1098, #1099, #1100, #1101, #1102, #1103, #1104, #1105, #1106, #1107, #1108, #1109, #1110, #1111, #1112, #1113, #1114, #1115, #1116, #1117, #1118, #1119, #1120, #1121, #1122, #1123, #1124, #1125, #1126, #1127, #1128, #1129, #1130, #1131, #1132, #1133, #1134, #1135, #1136, #1143.

### #1095: Qualify finite Rogue levels contracts

Finite prerequisite qualification ONLY; no new runtime implementation or eligibility requirement. Every listed blocker remains mandatory after329closure. Exact consumer roster: level-1, level-2, level-3, level-4, level-5, level-6, level-7, level-8, level-9, level-10, level-11, level-12, level-13, level-14, level-15, level-16, level-17, level-18, level-19, level-20, later-1.

Native parent #1093; exact blockers: #1149, #1150, #1151, #1152, #1153, #1154, #1155, #1156, #1157, #1158, #1159, #1160, #1161, #1162, #1163, #1164, #1165, #1166, #1167, #1168, #1169.

### #1096: Qualify finite Rogue shared contracts

Finite prerequisite qualification ONLY; no new runtime implementation or eligibility requirement. Every listed blocker remains mandatory after329closure. Exact consumer roster: hide, object-check, utilize, magic-item, attune-item, scroll-item, cloak-protection, gauntlets-ogre, ring-jumping, ring-protection, wand-dawn.

Native parent #192; exact blockers: #1137, #1138, #1139, #1140, #1141, #1142, #1144, #1145, #1146, #1147, #1148.

### #1097: Qualify finite Rogue execution contracts

Finite prerequisite qualification ONLY; no new runtime implementation or eligibility requirement. Every listed blocker remains mandatory after329closure. Exact consumer roster: expertise, sneak-attack, cunning-action, subclass, steady-aim, strike, poison, trip, withdraw, uncanny-dodge, evasion, reliable-talent, improved-strike, daze, knock-out, obscure, slippery-mind, elusive, luck, fast-hands-check, fast-hands-object, climber, jumper, stealth-attack, attunement, charges, scroll-check, scroll-light, scroll-detect, reflexes, hide, object-check, utilize, magic-item, attune-item, scroll-item, scroll-aid, cloak-protection, gauntlets-ogre, ring-jumping, ring-protection, wand-dawn. ONLY61check/save receipts,64inertprepare/durableconfirm/replay,65attackhit/damage,66actor/turn/actioncost events,67trustedtarget/range/sight/cover;273actual qualified casting-mode outputs (planning closure supplies NONE);353source-bound concentration start/maintain/end events ONLY when selected spell/mode requires them. Item/scroll routers and exact qualified effect owners must prove PDF206/244 modes: normal casting time,no slot,component exceptions,source access instead of class prepared eligibility; selected spell concentration remains with user. Do not import906class prepared/material requirements into item/scroll use. Preserve Spell Scroll fixed DC/attack table. Qualify actual outputs independently in each named consumer. No consumer can block this gateway; no effect/class/transaction/casting implementation owned here. No activeencounter/Spellcasting/Concentration prerequisite introduced.

Native parent #329; exact blockers: #61, #64, #65, #66, #67, #273, #353.

## Shared confirmation and client contract

Preview inert: no writes or dice. Explicit confirmation revalidates source,actor,campaign,revision,target,range,timing,cost and qualification. Injectable dice; atomic choice/effect/resource/history persistence. Retry/replay/reload/reconnect returns same receipt,conflicting payload rejects,concurrent confirmations commit ONCE. Inject failures for complete isolated PostgreSQL rollback. Web AND TUI expose supported choices/unavailable reasons; accessible cancel/confirm/reconnect; changed web requires desktop/phone fixtures,build and focused lint. Automatic effects belong to triggering confirmed event,not optional new activation. Unknown/private facts unavailable without leaking hidden reasons. Short tests are not integration-tag PostgreSQL/browser/device proof. No production/inference/runtime promotion. Class-total levels integer1-20,not booleans/fractions; each admitted class positive,post-transition total<=20; later-class variants require own multiclass eligibility,first-character variants do not acquire a multiclass prerequisite at runtime. Spell Recall needs actual Spellcasting; ordinary feats have no blanket casting/encounter requirement. Slot-cast one/turn applies only to slot expenditure TO CAST; authoritative confirmed invalid target spends slot/no effect with source-private filtering; unauthorized/unverifiable facts unavailable.

All leaves require positive and independently violated numeric/source/target/timing/eligibility controls,isolated PostgreSQL history/effect/choice rollback,durable same-ID replay/reload/concurrency and explicit shared web AND TUI availability. Short tests are separate from integration-tag PostgreSQL,desktop/phone/browser/device qualification. No production access,inference,deployment or runtime promotion follows from publishing this roadmap.

## Complete source-record execution matrix

Each canonical record,including every nested option/prerequisite/resource,is retained in `issue-329-rogue-scope.json`. Its `key_owners` maps EVERY required mechanic and option key; each row also names hashed engine/web/TUI evidence. All are currently metadata-only/unqualified.

| Complete source record | Exact execution owners |
|---|---|
| `srd-5.2.1:character/class/rogue` | hp (#1098), traits (#1099), skills (#1100), multiclass (#1102) |
| `srd-5.2.1:character/choice/rogue/skills` | skills (#1100) |
| `srd-5.2.1:character/choice/rogue/multiclass-skill` | multiclass (#1102) |
| `srd-5.2.1:character/choice/rogue/starting-equipment` | equipment (#1101) |
| `srd-5.2.1:character/level/rogue/1` | level-1 (#1149) |
| `srd-5.2.1:character/level/rogue/2` | level-2 (#1150) |
| `srd-5.2.1:character/level/rogue/3` | level-3 (#1151) |
| `srd-5.2.1:character/level/rogue/4` | level-4 (#1152) |
| `srd-5.2.1:character/level/rogue/5` | level-5 (#1153) |
| `srd-5.2.1:character/level/rogue/6` | level-6 (#1154) |
| `srd-5.2.1:character/level/rogue/7` | level-7 (#1155) |
| `srd-5.2.1:character/level/rogue/8` | level-8 (#1156) |
| `srd-5.2.1:character/level/rogue/9` | level-9 (#1157) |
| `srd-5.2.1:character/level/rogue/10` | level-10 (#1158) |
| `srd-5.2.1:character/level/rogue/11` | level-11 (#1159) |
| `srd-5.2.1:character/level/rogue/12` | level-12 (#1160) |
| `srd-5.2.1:character/level/rogue/13` | level-13 (#1161) |
| `srd-5.2.1:character/level/rogue/14` | level-14 (#1162) |
| `srd-5.2.1:character/level/rogue/15` | level-15 (#1163) |
| `srd-5.2.1:character/level/rogue/16` | level-16 (#1164) |
| `srd-5.2.1:character/level/rogue/17` | level-17 (#1165) |
| `srd-5.2.1:character/level/rogue/18` | level-18 (#1166) |
| `srd-5.2.1:character/level/rogue/19` | level-19 (#1167) |
| `srd-5.2.1:character/level/rogue/20` | level-20 (#1168) |
| `srd-5.2.1:character/feature/rogue/expertise` | expertise (#1103) |
| `srd-5.2.1:character/feature/rogue/sneak-attack` | sneak-attack (#1104) |
| `srd-5.2.1:character/feature/rogue/thieves-cant` | languages (#1105) |
| `srd-5.2.1:character/feature/rogue/weapon-mastery` | mastery (#1106) |
| `srd-5.2.1:character/feature/rogue/cunning-action` | cunning-action (#1107) |
| `srd-5.2.1:character/feature/rogue/rogue-subclass` | subclass (#1108) |
| `srd-5.2.1:character/feature/rogue/steady-aim` | steady-aim (#1109) |
| `srd-5.2.1:character/feature/rogue/ability-score-improvement` | improvement (#1110) |
| `srd-5.2.1:character/feature/rogue/cunning-strike` | strike (#1111), poison (#1112), trip (#1113), withdraw (#1114) |
| `srd-5.2.1:character/feature/rogue/uncanny-dodge` | uncanny-dodge (#1115) |
| `srd-5.2.1:character/feature/rogue/evasion` | evasion (#1116) |
| `srd-5.2.1:character/feature/rogue/reliable-talent` | reliable-talent (#1117) |
| `srd-5.2.1:character/feature/rogue/improved-cunning-strike` | improved-strike (#1118) |
| `srd-5.2.1:character/feature/rogue/devious-strikes` | daze (#1119), knock-out (#1120), obscure (#1121) |
| `srd-5.2.1:character/feature/rogue/slippery-mind` | slippery-mind (#1122) |
| `srd-5.2.1:character/feature/rogue/elusive` | elusive (#1123) |
| `srd-5.2.1:character/feature/rogue/epic-boon` | boon (#1124) |
| `srd-5.2.1:character/feature/rogue/stroke-of-luck` | luck (#1125) |
| `srd-5.2.1:character/choice/rogue/expertise-initial` | expertise (#1103) |
| `srd-5.2.1:character/prerequisite/rogue/expertise-initial` | expertise (#1103) |
| `srd-5.2.1:character/choice/rogue/expertise-level-6` | expertise (#1103) |
| `srd-5.2.1:character/prerequisite/rogue/expertise-level-6` | expertise (#1103) |
| `srd-5.2.1:character/choice/rogue/sneak-attack-eligibility` | sneak-attack (#1104) |
| `srd-5.2.1:character/prerequisite/rogue/sneak-attack-eligibility` | sneak-attack (#1104) |
| `srd-5.2.1:character/choice/rogue/additional-language` | languages (#1105) |
| `srd-5.2.1:character/choice/rogue/weapon-mastery` | mastery (#1106) |
| `srd-5.2.1:character/prerequisite/rogue/weapon-mastery-proficiency` | mastery (#1106) |
| `srd-5.2.1:character/choice/rogue/cunning-action` | cunning-action (#1107) |
| `srd-5.2.1:character/choice/rogue/rogue-subclass` | subclass (#1108) |
| `srd-5.2.1:character/prerequisite/rogue/steady-aim-unmoved` | steady-aim (#1109) |
| `srd-5.2.1:character/choice/rogue/ability-score-improvement` | improvement (#1110) |
| `srd-5.2.1:character/prerequisite/rogue/ability-score-improvement-qualification` | improvement (#1110) |
| `srd-5.2.1:character/choice/rogue/epic-boon` | boon (#1124) |
| `srd-5.2.1:character/prerequisite/rogue/epic-boon-qualification` | boon (#1124) |
| `srd-5.2.1:character/choice/rogue/cunning-strike` | strike (#1111), poison (#1112), trip (#1113), withdraw (#1114) |
| `srd-5.2.1:character/choice/rogue/devious-strikes` | daze (#1119), knock-out (#1120), obscure (#1121) |
| `srd-5.2.1:character/prerequisite/rogue/devious-strikes-options` | daze (#1119), knock-out (#1120), obscure (#1121) |
| `srd-5.2.1:character/choice/rogue/strike-poison` | poison (#1112) |
| `srd-5.2.1:character/prerequisite/rogue/strike-poison-selected` | poison (#1112) |
| `srd-5.2.1:character/prerequisite/rogue/strike-poison-numbers` | poison (#1112) |
| `srd-5.2.1:character/choice/rogue/strike-trip` | trip (#1113) |
| `srd-5.2.1:character/prerequisite/rogue/strike-trip-selected` | trip (#1113) |
| `srd-5.2.1:character/choice/rogue/strike-withdraw` | withdraw (#1114) |
| `srd-5.2.1:character/prerequisite/rogue/strike-withdraw-selected` | withdraw (#1114) |
| `srd-5.2.1:character/prerequisite/rogue/strike-withdraw-numbers` | withdraw (#1114) |
| `srd-5.2.1:character/choice/rogue/strike-daze` | daze (#1119) |
| `srd-5.2.1:character/prerequisite/rogue/strike-daze-selected` | daze (#1119) |
| `srd-5.2.1:character/choice/rogue/daze-restricted-turn` | daze (#1119) |
| `srd-5.2.1:character/choice/rogue/strike-knock-out` | knock-out (#1120) |
| `srd-5.2.1:character/prerequisite/rogue/strike-knock-out-selected` | knock-out (#1120) |
| `srd-5.2.1:character/prerequisite/rogue/strike-knock-out-numbers` | knock-out (#1120) |
| `srd-5.2.1:character/choice/rogue/strike-obscure` | obscure (#1121) |
| `srd-5.2.1:character/prerequisite/rogue/strike-obscure-selected` | obscure (#1121) |
| `srd-5.2.1:character/prerequisite/rogue/improved-cunning-strike-operands` | improved-strike (#1118) |
| `srd-5.2.1:character/resource/rogue/stroke-of-luck` | luck (#1125) |
| `srd-5.2.1:character/subclass/rogue/thief/identity` | subclass (#1108) |
| `srd-5.2.1:character/prerequisite/rogue/thief/identity-level` | subclass (#1108) |
| `srd-5.2.1:character/prerequisite/rogue/thief/identity-selection` | subclass (#1108) |
| `srd-5.2.1:character/feature/rogue/thief/fast-hands` | fast-hands-check (#1126), fast-hands-object (#1127) |
| `srd-5.2.1:character/prerequisite/rogue/thief/fast-hands-level` | fast-hands-check (#1126), fast-hands-object (#1127) |
| `srd-5.2.1:character/choice/rogue/thief/fast-hands` | fast-hands-check (#1126), fast-hands-object (#1127) |
| `srd-5.2.1:character/choice/rogue/thief/sleight-of-hand` | fast-hands-check (#1126) |
| `srd-5.2.1:character/prerequisite/rogue/thief/sleight-of-hand-check` | fast-hands-check (#1126) |
| `srd-5.2.1:character/choice/rogue/thief/use-an-object` | fast-hands-object (#1127) |
| `srd-5.2.1:character/prerequisite/rogue/thief/use-an-object-selection` | fast-hands-object (#1127) |
| `srd-5.2.1:character/prerequisite/rogue/thief/sleight-of-hand-selection` | fast-hands-check (#1126) |
| `srd-5.2.1:character/feature/rogue/thief/second-story-work` | climber (#1128), jumper (#1129) |
| `srd-5.2.1:character/prerequisite/rogue/thief/second-story-work-level` | climber (#1128), jumper (#1129) |
| `srd-5.2.1:character/feature/rogue/thief/supreme-sneak` | stealth-attack (#1130) |
| `srd-5.2.1:character/prerequisite/rogue/thief/supreme-sneak-level` | stealth-attack (#1130) |
| `srd-5.2.1:character/choice/rogue/thief/stealth-attack` | stealth-attack (#1130) |
| `srd-5.2.1:character/prerequisite/rogue/thief/stealth-attack-strike` | stealth-attack (#1130) |
| `srd-5.2.1:character/prerequisite/rogue/thief/stealth-attack-hide` | stealth-attack (#1130) |
| `srd-5.2.1:character/choice/rogue/thief/stealth-cover` | stealth-attack (#1130) |
| `srd-5.2.1:character/prerequisite/rogue/thief/stealth-cover-timing` | stealth-attack (#1130) |
| `srd-5.2.1:character/feature/rogue/thief/use-magic-device` | attunement (#1131), charges (#1132), scroll-check (#1133), scroll-light (#1134), scroll-detect (#1135), scroll-aid (#1143) |
| `srd-5.2.1:character/prerequisite/rogue/thief/use-magic-device-level` | attunement (#1131), charges (#1132), scroll-check (#1133), scroll-light (#1134), scroll-detect (#1135), scroll-aid (#1143) |
| `srd-5.2.1:character/choice/rogue/thief/scrolls` | scroll-check (#1133) |
| `srd-5.2.1:character/prerequisite/rogue/thief/scrolls-level` | scroll-check (#1133) |
| `srd-5.2.1:character/choice/rogue/thief/scroll-check` | scroll-check (#1133) |
| `srd-5.2.1:character/prerequisite/rogue/thief/scroll-check-condition` | scroll-check (#1133) |
| `srd-5.2.1:character/feature/rogue/thief/thiefs-reflexes` | reflexes (#1136) |
| `srd-5.2.1:character/prerequisite/rogue/thief/thiefs-reflexes-level` | reflexes (#1136) |

## Universal identity permission and conditional effects

The complete597-record inventory (339spells,258magic-item headings) is retained in `dispatch_inventory`,including source pages/IDs and every unqualified dispatch ID. Universal Spell Scroll identity permission does not grant Spellcasting,prepared spells,slots or effect execution. All0–9 scroll source levels have exact DC/attack-table contracts; Thief’s reliable0/1 and INT(Arcana) checks for EACH2–9 are separately tested. Unknown/unqualified effects are rejected before rolling/spending. Qualified failed higher-level Thief casts disintegrate the scroll with no effect. Normal interrupted item casting preserves it. Ordinary scroll failure removes its spell without invented physical destruction.

Concrete spell composition fixtures consume Light #967,Detect Magic #906 and Aid #614; Ring of Jumping consumes Jump #1064. These are actual existing owner contracts,not copies of their effects. Four distinct attunable item effects and Wand dawn recharge are independent leaves. Spell Recall #399 remains unavailable to pure Rogue,Magic Initiate or UMD-only actors; actual authoritative qualified Spellcasting is required. Ordinary feat selection adds no blanket caster,encounter or Concentration condition.

Remaining 587 source identities lack a qualified effect binding through this Rogue dispatcher; their COMPLETE names/IDs/source pages remain explicit in the JSON. This is a dispatch availability statement,not a claim that other shared decisions lack owners. Their source presence does not qualify them.

## Proof boundaries

The strict offline/live verifier checks complete source transcriptions,source-key ownership,full actual native contracts,Task/Project Size,all old required paths,all own paths after decision closure,exact finite gateway rosters,both DAGs and native caps. Independent semantic reconstruction and immutable seals reject coherently weakened own matrices/receipts. FINAL parent proof separately rejects missing release,original-only rollback,changed prefixes/sections and duplicate or altered duplicate markers,while accepting later unrelated parent appends. No current live body is automatically accepted as a released prefix.

## Dated publication and final seal

The four `-PENDING.md` files are immutable prepared artifacts. Their embedded PENDING wording records the earlier handoff stage; ROOT published their exact bytes on2026-11-16 at20:06:53–20:07:04UTC. The final sealed observation at 2026-11-16T20:08:42.485242+00:00 passed and supersedes that workflow status. The artifacts stay byte-for-byte unchanged for independent prefix/section/marker verification. This seals publication evidence,not runtime implementation.
