"""Readable independent SRD5.2.1 Warlock decision expectations (CC-BY-4.0).

These expectations do not read the candidate manifest or native receipts.
Source record expectations are supplied by the independently pinned comparators.
"""

COMMON = ('Prepare is inert: no dice, cost, grant or writes. Confirm revalidates actor, '
          'campaign, source, revision, entitlement, target, action, timing and operands. '
          'Persist choice, injected dice, resources, effects, derived sheet and history '
          'atomically in isolated PostgreSQL. Same-ID replay/reload/reconnect returns '
          'one receipt without reroll or regrant; conflicting payload, foreign state '
          'and stale revision reject. Concurrent confirms commit exactly once. Inject '
          'failure after cost/effect/history and prove full rollback. Web and TUI '
          'consume the same engine availability and explicit unavailable reasons. '
          'Unqualified effects stay unavailable; source metadata and closed planning '
          'issues grant nothing. SRD5.2.1 CC-BY4 attribution/adaptation is retained '
          'via docs/srd-5.2.1-conformance.md.')

CONSUMER_EVIDENCE = [
 dict(area='rules', path='internal/rules/srd521/creation.go', symbol='SupportedBuildForConcept', finding='Fixed Fighter/Cleric builds only; Warlock is not admitted.'),
 dict(area='rules', path='internal/rules/srd521/equipment.go', symbol='ChangeEquipment', finding='Chain mail/shield and fixed Fighter/Cleric only; no general Leather/Orb/Book/weapon adapter.'),
 dict(area='game', path='internal/game/advancement_preview.go', symbol='prepareAdvancement', finding='Fixed Fighter1-to2 preview; no Warlock class vector or adjacent transition.'),
 dict(area='engine', path='internal/engine/action_confirmation.go', symbol='ConfirmAction', finding='Authoritative preview/revision gate and narrow spell targets; storage does not supply Warlock effects.'),
 dict(area='combat', path='internal/combat/encounter.go', symbol='SpendSpell', finding='One slot TO CAST turn budget; no Pact/Arcanum/Invocation resource ownership.'),
 dict(area='database', path='internal/state/sqlc/player_characters.sql.go', symbol='CreatePlayerCharacter', finding='Persists character state; no source grant/effect execution supplied by stored labels or JSON.'),
 dict(area='handlers', path='internal/handlers/advancement_integration_test.go', symbol='TestIntegrationFighterAdvancement', finding='Isolated PostgreSQL Fighter precedent only, not Warlock qualification.'),
 dict(area='handlers', path='internal/world/rules_character.go', symbol='ValidateRulesCharacter', finding='Guided/interview/persistence share supported-build validation.'),
 dict(area='web', path='frontend/src/components/start/CharacterGuidedForm.tsx', symbol='Warlock label', finding='Display label does not admit or execute Warlock.'),
 dict(area='tui', path='tui/campaign/char_form.go', symbol='NewCharFormModel', finding='Class labels from world.D5EClasses; no independent Warlock rules executor.'),
 dict(area='rules', path='internal/rules/srd521/catalogue_warlock_test.go', symbol='TestWarlockSourceAndBoundaries', finding='Explicit metadata-only and Warlock creation rejection.'),
]

CONTRACTS = {
 'core-traits': 'PDF70first-character ONLY: Charismaprimary(noautomaticabilityincrease),Wisdom/Charismasaveproficiencies,Simpleweaponproficiency,Lightarmortraining;HitDie/HPownadapter,skill/equipmentchoicesseparate. PositivefirstWarlock1bothsaves/simple/light; rejectMartial/Medium/Shieldtraining,automaticCHA13orfirst-classsavesonlater-entry. Consume430wholeprovenance; do notgrantitemeffectsfromproficiency.',
 'skills': 'PDF70first-characterONLYchoose2distinctArcana/Deception/History/Intimidation/Investigation/Nature/Religion. Actualskillmodifiers/PBvia61/58with430sourceprovenance; overlappingothersourcegrantdoesnotdoublePB. PositiveArcana+ReligionINT2/PB2=4each; reject3selections,duplicatewithinchoice,unlistedPerception,PBtwiceornewskillsonlaterWarlockentry. Do notinventunprintedreplacementpermissions.',
 'equipment': 'PDF70initialA grantsLeatherArmor,Sickle,2Daggers,OrbArcaneFocus,occultBook,ScholarsPack,15GP; B grants100GPonly; chooseexactlyone,creationonly. ActualScholarPack99 addsBackpack,SECOND Book,Ink,InkPen,Lamp,10Oil,10Parchment,Tinderbox. PositiveA2Daggers/2Books/10Oil/15GP; reject1Booktotal,3Daggers,mixedA+B,purchasechargesatcreationorgrantingpackagainonlaterentry. Consumeactualnameditemoutputs; storage/identityalone doesnotexecutearmor/weapon/book/focus/light effects.',
 'warlock-list': 'PDF74–76/105: exact72 printed Warlock spell identities with level/school/Ritual/Concentration/material flags; prepared Pact selection levels1–5 and known cantrip selection level0 use own list, not503. Mystic Arcanum levels6–9 separate; named Fiend/Invocation/Tome routes do not silently add all117 to ordinary list. PositiveMindSpike listed75/header149 eligibleWarlock2; rejectFireball ordinaryWarlocklist, SpeakwithDead ordinarylist, disputedSorcererMindSpike grant or header-onlyoverride. Unknown/unqualified effects unavailable even when source eligibility is true.',
 'pact-cross-casting': 'PDF25–26: actual PactMagic AND actualSpellcasting required for cross-casting. Pactslots may cast prepared Spellcasting-class spells; actualSpellcasting slots may cast preparedWarlockspells, normal slot>=spelllevel and sourcecastingability/provenance. Pactlevels excluded sharedSpellcasting-slot sum; Arcanum resources remain separate. Positive admittedWizard1/Warlock1 preparedMagicMissile usesPactslot1; preparedHex usesWizard1slot withCHA; rejectMagicInitiate-only feature, unpreparedArcanumasordinaryprepared, addingWarlocklevel tosharedsum or counting other slot resource spend as slotTOCAST.',
 'hp': 'PDF23/70: own D8/fixed5 HP adapter consumes968, never direct429. First-character level1 gets8+CON. Later entry and every adjacent row add max(1,D8+CON) or max(1,5+CON) and one D8 Hit Die; no healing or reroll on replay. Positive CON2 fixed later gain7; reject D6, fixed4, first-character8 on later entry and duplicate gain.',
 'creation': 'PDF70: first-character creation only at level1, Charisma primary, Wisdom/Charisma saves, simple weapons, light armor; choose2 distinct Arcana/Deception/History/Intimidation/Investigation/Nature/Religion. Equipment A: Leather Armor,Sickle,2 Daggers,Arcane Focus orb,occult Book,Scholars Pack,15GP; B:100GP. Positive A exactly2 Daggers/15GP; reject3 skills, duplicate skill, mixed A+B, medium armor. Inventory identity does not execute item effects.',
 'later-entry': 'PDF24/70: add Warlock1 only to admitted existing class combination with Charisma13 and each existing class source minimum; grant D8 and light armor plus level1 features, no saves/weapons/skills/equipment. Fighter4->Warlock1 total5 uses6500XP/PB3 and later HP. Reject CHA12, boolean13, unqualified prior class, total>20 or first-character HP/equipment.',
 'eldritch-invocations': 'PDF70–71: exact per-Warlock-level counts; level1 choose1; level2 total3; level5 total5. Learn only eligible options. Once per gained Warlock level replace at most1; cannot replace an Invocation required by another owned Invocation. Nonrepeatable default; only Agonizing Blast/Eldritch Spear/Repelling Blast/Lessons repeat with different eligible selection. Positive remove independent option; reject removal Pact Blade while Thirsting Blade retained, duplicate nonrepeatable, same repeatable cantrip and2 replacements.',
 'pact-magic': 'PDF71/104–106: Charisma spell attack/DC; two Warlock cantrips, two prepared level1 spells initially; exact list adapter not503. Spell costs one distinct Pact slot of printed class slot level, not generic Spellcasting or Arcanum resource. Positive Warlock5 Charm Person cast at3 using one of2 slots; reject unprepared, wrong list, level4 spell atWarlock5, generic slot substitution without actual qualified multiclass routing. One SLOT TO CAST per turn; other slot resource expenditures are not casts. Confirm private invalid target spends confirmed slot/no effect without private leakage; unverifiable facts unavailable.',
 'preparation': 'PDF71–72: on each Warlock level replace at most1 class cantrip and1 prepared spell; additions exactly printed count deltas, chosen level no higher than own slot level. Always-prepared feature spells excluded from class prepared count; retain distinct source provenance. Positive level4 third cantrip/5 prepared; reject rest-based wholesale replacement, overcapacity, invalid level, removing another source grant.',
 'slots': 'PDF71/427/428: distinct Pact pool recovers all expended slots after completed Short or Long Rest. Capacity by own Warlock level1=1@1,2=2@1,11=3@5,17=4@5. Positive spent2/2->2 after rest; reject overcapacity, Arcanum recharge on short rest, interruption and duplicate receipt recharge.',
 'magical-cunning': 'PDF72: perform1-minute rite; at end regain expended Pact slots up to ceil(max/2), once per Long Rest. Positive capacity3 spent3 recovers2, capacity1 recovers1; reject instant recovery, generic pool, second use, floor rounding, interruption/replay restoring again.',
 'warlock-subclass': 'PDF72/76: choose one qualified subclass at Warlock3; Fiend grants3/6/10/14 features at own Warlock level. Positive Warlock3 selects Fiend; reject Warlock2, character-total3/Warlock1, second subclass and source-only subclass execution.',
 'ability-score-improvement': 'PDF72/87: independent feat grant at each Warlock4/8/12/16; choose ASI or another eligible feat, exact385–401 effect owners and384 selection/389 caps. Positive four distinct grant receipts; reject missed level8, double grant at4, absorbing Invocation/Arcanum/other feature. Spell Recall399 requires actual Spellcasting feature, not Pact Magic alone, Magic Initiate or scroll identity.',
 'contact-patron': 'PDF72: at Warlock9 always prepare Contact Other Plane; special cast contacts patron, no slot, automatic save success, once per Long Rest; consume real effect owner. Positive eligible special cast with0 Pact slots; reject second free use, wrong contact recipient, generic spell-wide auto-success or lost always-prepared provenance.',
 'mystic-arcanum': 'PDF72: separate one-use resource for chosen level6/7/8/9 Warlock spell at own11/13/15/17; slot-free cast once per resource, all recover Long Rest. On every Warlock level replace at most1 arcanum with same-level Warlock spell. Positive own17 retains four independent resources; reject shared one-use counter, Pact expenditure, short rest recovery, wrong-level replacement,2 replacements.',
 'epic-boon': 'PDF72/87–88: own Warlock19 grants one Epic Boon or other eligible feat; Boon of Fate recommendation is not mandatory. Positive admitted eligible feat; reject level18, duplicate receipt, ineligible Spell Recall via PactMagic name, replacing Invocation or ASI receipt.',
 'eldritch-master': 'PDF72: own Warlock20 changes Magical Cunning recovery to all expended Pact slots; retain1-minute rite/once Long Rest. Positive4 spent restores4; reject regain2, generic slots, twice per rest or instant restore.',
 'dark-ones-blessing': 'PDF76: when you reduce enemy to0HP, or someone else reduces enemy within10ft of you to0HP, gain max(1,CHAmod+WarlockLevel) Temporary HP. Positive Warlock3 CHA2 yields5, other killer10ft eligible; reject11ft other killer, nonenemy, heal5, stacking Temporary HP.',
 'fiend-spells': 'PDF76: own3 Burning Hands/Command/Scorching Ray/Suggestion;5 Fireball/Stinking Cloud;7 Fire Shield/Wall of Fire;9 Geas/Insect Plague always prepared without consuming class preparation capacity. Positive own9 ten spells retained; reject grant at character-total9/Warlock3, missing earlier rows, capacity subtraction, names as execution.',
 'dark-ones-own-luck': 'PDF76: add1d10 to own ability check or save after seeing roll before effects; max1 per roll; max(1,CHAmod) uses per Long Rest. Positive CHA3 permits3 distinct rolls; reject attack roll, after effects, fourth use or two uses same roll. Undisclosed outcome alone does not forbid outcome-independent known-die windows.',
 'fiendish-resilience': 'PDF76: at completed Short/Long Rest choose1 damage type except Force; Resistance until choose different. Positive Fire->Cold replaces; reject Force,2 types, arbitrary anytime change, expired on next rest without replacement.',
 'hurl-through-hell': 'PDF76: once per turn on attack-roll hit creature, CHA save vs own spellDC; failure transports,8d10 Psychic unless Fiend, Incapacitated until end of your next turn; return prior or nearest unoccupied space. Once per Long Rest; no-action Pact slot restores use. Positive failed nonFiend takes8d10; Fiend still transports but0 Psychic. Reject hitless trigger, second same turn, generic slot restore, success transport,8d10 to Fiend. Split world transit/damage/487 lifecycle before publication.',
 'agonizing-blast': 'PDF72: own2+, choose known damaging Warlock cantrip, add CHAmod to its damage rolls; repeat only different eligible cantrip. Positive CHA3 adds3 per damage roll; reject unknown/nonWarlock/non-damaging, duplicate selection or only once per cast.',
 'armor-of-shadows': 'PDF72: cast Mage Armor on self without slot. Positive0slots self; reject other target, copied effect or inferred preparation requirement.',
 'ascendant-step': 'PDF72: own5+, cast Levitate on self without slot. Positive own5/0slots; reject own4 or other target; consume exact spell effect.',
 'devils-sight': 'PDF72: own2+, normal vision in magical/nonmagical Dim Light/Darkness within120ft. Positive magical Darkness120; reject121ft, Truesight or seeing through obstruction.',
 'devouring-blade': 'PDF72: own12+ and Thirsting Blade; its pact-weapon Extra Attack grants2 extra attacks (3 total). Positive3 pact attacks; reject own11, nonpact weapon,4 attacks or stacking independent ExtraAttack.',
 'eldritch-mind': 'PDF72: Advantage on CON saves to maintain Concentration. Positive casting concentration damage save; reject ordinary CON save, other ability, general save advantage.',
 'eldritch-smite': 'PDF72–73: own5+/PactBlade; once per turn hit creature with pact weapon, spend1 Pact slot, extra(1+slotLevel)d8 Force; optional Prone if Huge or smaller via488. Positive slot3=4d8 Huge Prone; reject Gargantuan Prone, miss, generic slot, second same turn. Slot resource spend is not slot TO CAST.',
 'eldritch-spear': 'PDF73: own2+, known damaging Warlock cantrip printed range10+ft; on cast add30*WarlockLevel ft. Repeat different eligible cantrip. Positive own5 base10->160; reject base5, unknown, total-level scaling or duplicate cantrip.',
 'fiendish-vigor': 'PDF73: own2+, cast False Life self without slot, maximum die result instead of roll. Positive base2d4+4=12THP; reject roll, other target, spell slot spending or HP healing. Consume exact effect/scaling owner.',
 'gaze-of-two-minds': 'PDF73: own5+, Bonus Action touch willing creature; perceive its senses through end next turn; Bonus Action subsequent turns maintains on same plane. Cast from either space only within60ft. Positive60ft cast origin; reject61ft origin, unwilling touch, cross-plane maintenance, missed maintenance. Split perception/maintenance/casting-origin contracts.',
 'gift-of-the-depths': 'PDF73: own5+, underwater breathing and SwimSpeed=Speed; Water Breathing slot-free once per Long Rest. Positive Speed30->Swim30; reject fixed40, second free cast or short-rest reset; separate passive movement and named effect mode.',
 'gift-of-the-protectors': 'PDF73: own9+/PactTome; page on conjured book, permission+creature action writes name, max(1,CHAmod) names; named creature reduced0HP not killed outright drops1HP; global once Long Rest, not per name; Magic action touch erases name. Positive CHA2 two names/one total rescue; reject third name, no permission, outright death rescue, second creature rescue before rest. Split inscription/trigger/recovery.',
 'investment-of-the-chain-master': 'PDF73: own5+/PactChain, when casting Find Familiar grant Fly or Swim40; Bonus Action commands Attack action; optional B/P/S->Necrotic/Radiant per damage; familiar saveDC yours; Reaction Resistance against that damage. Positive40ft and actual familiar saveDC; reject both speeds, replacing Fire, free reaction, command without Bonus Action. Split five benefits and familiar infusion lifecycle.',
 'lessons-of-the-first-ones': 'PDF73: own2+, gain1 Origin feat via actual384/385–388 effects; repeat only different Origin feat. Positive Alert then Skilled; reject duplicate Alert, General feat, nonqualified feat effect or ignoring nested choices.',
 'lifedrinker': 'PDF73: own9+/PactBlade once per turn pact-weapon hit creature extra1d6 Necrotic/Psychic/Radiant; optionally spend one own Hit Die, heal max(1,dieRoll+CONmod). Positive D8roll4/CON2 heals6; reject zero HitDice healing, attack miss, second per turn, d6 forced for HitDie or wrong damage.',
 'mask-of-many-faces': 'PDF73: own2+, Disguise Self without slot. Positive0slots normal spell targets; reject own1, fabricated effect or blanket self-only restriction absent source.',
 'master-of-myriad-forms': 'PDF73: own5+, Alter Self without slot; consume appearance/aquatic/natural-weapons bounded owners reserved330. Positive0slots qualified mode; reject own4 or inferring unqualified natural weapons.',
 'misty-visions': 'PDF74: own2+, Silent Image without slot; consume real effect. Positive0slots; reject own1, treating image as real object or copied spell owner.',
 'one-with-shadows': 'PDF74: own5+, while in DimLight/Darkness cast Invisibility self without slot. Positive DimLight self; reject BrightLight, own4, other target or inferred immobility ending from2014.',
 'otherworldly-leap': 'PDF74: own2+, Jump self without slot. Positive self0slots; reject other target, own1 or general free Jump permission.',
 'pact-of-the-blade': 'PDF74: Bonus Action conjure chosen Simple/Martial Melee weapon in hand and bond, or touch magic weapon not attuned to someone else/not bonded to another Warlock. Grant proficiency/focus; attacks optionally CHA and normal/Necrotic/Psychic/Radiant. Bond ends reuse BA, death or >5ft for1minute; conjured weapon disappears. Positive distance6ft60s ends; reject5ft,59s, other attunement, ranged mundane conjure, stale old bond. Split conjure/bond, combat, focus/proficiency, lifecycle; no general item effects inferred.',
 'pact-of-the-chain': 'PDF74: learn Find Familiar; Magic-action slot-free casting; normal forms or Imp/Pseudodragon/Quasit/Skeleton/Sphinx of Wonder/Sprite/Venomous Snake. On own Attack action forgo1 own attack for familiar1 attack using its Reaction. Positive1 forfeited/1 familiar Reaction; reject free familiar attack, spent Reaction, multiattack substitution or unqualified form/statblock. Split spell mode/forms/attack exchange.',
 'pact-of-the-tome': 'PDF74: at end Short/Long Rest conjure book in hand; choose3 cantrips from finite27 and2 level1 Rituals from finite11, none already prepared. While book on person chosen spells prepared as Warlock spells; book focus. Only you access magic; new book/death disappears old. Positive3+2 distinct eligible unprepared; reject4 cantrips, level2 ritual, already prepared, absent-book grant or2014 transcription. Split conjuration/lifecycle, finite choices/provenance, focus; normal prepared Ritual cast+10minutes/no slot/no upcast, no extra RitualAdept.',
 'repelling-blast': 'PDF74: own2+, known Warlock cantrip requiring attack roll; hit Large/smaller pushes up to10ft straight away. Repeat different eligible cantrip. Positive Large10ft; reject Huge,11ft, sideways push, saving-throw cantrip or duplicate selection.',
 'thirsting-blade': 'PDF74/25: own5+/PactBlade; Attack action on own turn twice with pact weapon only. Positive2 attacks; reject nonpact, offturn, third attack, stacking independent ExtraAttack; DevouringBlade explicit exception.',
 'visions-of-distant-realms': 'PDF74: own9+, Arcane Eye without slot, real qualified sensor/movement owners reserved330. Positive0slots; reject own8 or treating sensor observation as omniscience.',
 'whispers-of-the-grave': 'PDF74: own7+, Speak with Dead without slot; exact dead-creature effect owner, no generic undead command. Positive own7/0slots; reject own6 or unqualified effect.',
 'witch-sight': 'PDF74: own15+, Truesight30ft; qualified glossary perception including actual962 Invisible obligations. Positive30ft; reject31ft, own14, unlimited private disclosure.',
}

TOME_CANTRIPS = 'Acid Splash|Chill Touch|Dancing Lights|Druidcraft|Eldritch Blast|Elementalism|Fire Bolt|Guidance|Light|Mage Hand|Mending|Message|Minor Illusion|Poison Spray|Prestidigitation|Produce Flame|Ray of Frost|Resistance|Sacred Flame|Shillelagh|Shocking Grasp|Sorcerous Burst|Spare the Dying|Starry Wisp|Thaumaturgy|True Strike|Vicious Mockery'.split('|')
TOME_RITUALS = 'Alarm|Comprehend Languages|Detect Magic|Detect Poison and Disease|Find Familiar|Floating Disk|Identify|Illusory Script|Purify Food and Drink|Speak with Animals|Unseen Servant'.split('|')
FIEND_SPELLS = 'Burning Hands|Command|Scorching Ray|Suggestion|Fireball|Stinking Cloud|Fire Shield|Wall of Fire|Geas|Insect Plague'.split('|')
NAMED_SPELLS = 'Mage Armor|Levitate|False Life|Water Breathing|Disguise Self|Alter Self|Silent Image|Invisibility|Jump|Find Familiar|Arcane Eye|Speak with Dead|Contact Other Plane'.split('|')

# Large feature entries above become finite qualification contracts over these
# leaves. They retain all source obligations but own no runtime implementation.
SPLITS = {
 'pact-of-the-blade': {
  'blade-bond': 'PDF74: Bonus Action conjure chosen Simple/Martial Melee weapon in hand and bond OR touch magic weapon; reject other attunement/Warlock bond. Positive mundane Sickle in hand; reject Longbow conjure, remote magic weapon, second live bond. No copied magic-item effect or arbitrary item definition.',
  'blade-combat': 'PDF74: bonded weapon attacks optionally use CHA for attack AND damage, choose normal/Necrotic/Psychic/Radiant per attack. Positive CHA3 normal weapon damage+3; reject CHA substitution on unbonded weapon, Force option or mixed attack/damage ability provenance.',
  'blade-proficiency-focus': 'PDF74: proficiency and SpellcastingFocus exist only until exact weapon bond ends; use430provenance. Positive bonded proficient/qualifiedfocus; reject global martial proficiency, unrelated item focus, removing other source proficiency on bond end.',
  'blade-ending': 'PDF74: bond ends feature BonusAction reused, death, or weapon more than5ft for1minute; conjured weapon disappears then. Positive6ft60s; reject5ft/60s,6ft59s or disappearing nonconjured magic weapon; race new bond against old ending atomically.',
 },
 'pact-of-the-chain': {
  'chain-cast-mode': 'PDF74: grant FindFamiliar identity; cast via Magic action without slot, retaining source-specific component exceptions only if stated. Positive0slots Magic action; reject generic1-hour casting for this mode or unrelated spell freecast. Named base effect owner mandatory; planning181/185/274 not sufficient.',
  'chain-form-choice': 'PDF74: choose normal FindFamiliar form OR exactly Imp/Pseudodragon/Quasit/Skeleton/SphinxofWonder/Sprite/VenomousSnake; each needs own real named statblock/actions/availability outputs. Positive qualifiedImp; reject unknownDragon, unqualifiedImp, copying2014forms, treating name as statblock.',
  'chain-attack-exchange': 'PDF74: on own Attack action forgo1 own attack to allow familiar1 attack using its Reaction. Positive actor2attacks forgo1/familiarReaction1; reject extra free attack, Reaction0, nonAttackaction, multipleattacks from one exchange. Consume exact familiar attack owner, never automaticMultiattack.',
 },
 'pact-of-the-tome': {
  'tome-conjuration-ending': 'PDF74: end completed Short/LongRest conjure book in hand; old book disappears on another conjure or death. Positive completedShortRest onebook; reject arbitraryMagicAction conjure, interruptedrest, twoactivebooks or other creature access.',
  'tome-finite-preparation': 'PDF74: on book appearance choose3 distinct cantrips from exact27 and2 distinct level1Rituals from exact11; none already prepared. While on person prepared as Warlock spells independent of class count. Positive3+2; reject4cantrips, level2Ritual, alreadyprepared, absentbook cast provenance. Eligibility can include finite unqualified spells; execution remains unavailable.',
  'tome-focus': 'PDF74: own conjured BookofShadows usable as SpellcastingFocus; book must be accessible under real focus rules, not generic component waiver. Positive ownbook qualifiedcasting; reject otherbook, consumed/costlymaterial waiver or copied itemcasting behavior.',
 },
 'gaze-of-two-minds': {
  'gaze-connection': 'PDF73: BonusAction touch willing creature; perceive special senses until end nextturn; sameplane BonusAction each subsequentturn extends. Positive maintain nextturn; reject unwilling, crossplane, missedBA; actual perception owner cannot reveal private unavailable facts.',
  'gaze-cast-origin': 'PDF73: while connection active choose own or linkedcreature spellorigin only within60ft. Positive60ft alternateorigin; reject61ft, expiredconnection or target/range measured from bothorigins. Consume exact qualified effect, targeting and casting protocols.',
 },
 'gift-of-the-depths': {
  'depths-passive': 'PDF73: own5+ breathe underwater and SwimSpeed equals current Speed. PositiveSpeed30->Swim30, speed20->Swim20; reject fixed40 or no underwaterbreathing. Do not infer aquatic statblock or environmental immunity.',
  'depths-cast-mode': 'PDF73: own5+ cast WaterBreathing withoutslot onceLongRest. Positive uses1->0 at0slots; reject seconduse, ShortRestreset or copied spell effect. Exact spellowner mandatory.',
 },
 'gift-of-the-protectors': {
  'protectors-inscription': 'PDF73: own9+/Tome, newpage on conjurebook; with permission creature action writes ownname, max(1,CHAmod)names; MagicActiontouch erases. PositiveCHA2 twonames; reject thirdname/no permission/freeerase. Actualbookidentity/lifecycle mandatory.',
  'protectors-rescue': 'PDF73: namedcreature reduced0HP not killedoutright drops1HP instead; shared global trigger onceLongRest. Positive twonames firstrescue only; reject outrightdeath, secondname before rest, TempHP insteadHP or trigger on unnamedcreature. AtomicHP/lifecycle/ending owner required.',
 },
 'investment-of-the-chain-master': {
  'investment-movement': 'PDF73: own5+/Chain when castingFindFamiliar infuse summonedfamiliar; chooseFly40 OR Swim40. PositiveFly40; reject both, speed30, retroactive unrelatedcreature. Realformstatblock/lifecycle required.',
  'investment-command': 'PDF73/257: BonusAction command familiar Attack action, consume actual familiar action economy/attack implementation. Source257Multiattack IS partAttackaction: qualifiedPseudodragon mayuseTWO Bites; this is distinct from ChainONEattackexchange. Positive1BA/twoactualBites; reject noBA, unrelatedcreature, unsupported attack, twoStings or Chainforgoneattackforfree.',
  'investment-damage': 'PDF73: whenever familiar deals B/P/S optionally make thatdamage Necrotic or Radiant. PositivePiercing->Radiant; rejectFire->Radiant, Force choice, changing extra unrelateddamage.',
  'investment-save-dc': 'PDF73: if familiar forces save use your spellSaveDC. PositiveCHA3/PB3/DC14; reject creatureprintedDC, attackingwithDC or changing nonfamiliar save.',
  'investment-resistance': 'PDF73: when familiar takesdamage use ownReaction to grant Resistance against thatdamage. PositiveReaction1/Fire9->4; reject Reaction0, unrelateddamagepacket or persistentResistance afterpacket.',
 },
 'hurl-through-hell': {
  'hurl-admission-cost': 'PDF76: own14+/Fiend onceperturn attackroll hitcreature, ownspellDC Charisma save; onceLongRest use, noactionPactslot restoresuse. PositivePactslot1 restores; reject miss, seconduse sameturn, genericslot restore, resource restoration treated slotTOCAST.',
  'hurl-transit-ending': 'PDF76: failedsave targetdisappears, Incapacitated via487 until end YOURnextturn, then return prior or nearestunoccupiedspace. Positive blockedprior usesnearest; reject successfulsave transport, ending targetnextturn, unknownworldoccupancy or unsupportedplane behavior. Real boundedworld/transit owner required.',
  'hurl-psychic': 'PDF76: transported failedsave nonFiend takes8d10Psychic; Fiend takes0 but still transports. Positive8injected10s=80 beforequalifieddefenses; reject Psychic toFiend, success damage, wrongdice or omittingFiendtransport.',
 },
 'mystic-arcanum': {
  **{f'arcanum-{spell}': f'PDF72: ownWarlock{level}+ choose1 level{spell} Warlockspell; separate1use slot-free resource, LongRest recovery; each gainedWarlocklevel replace atmost1 arcanum across wholefeature with samelevel. Positive own17 retains resources6/7/8/9 separately, this{spell}resource spends1 only; reject othersspent, Pactslotcost, ShortRestrecharge, wronglevel replacement or2global replacements. Actual named effect and specialcasting mode mandatory.' for level, spell in [(11,6),(13,7),(15,8),(17,9)]},
 },
}


def body(key, fixture):
    return f'Decision #331. One implementation PR; Task Size M.\n\n{fixture}\n\n{COMMON}\n'


TRANSITION_WITNESSES = [
    dict(before={'fighter': 4}, after={'fighter': 4, 'warlock': 1}, warlock_row=1,
         total_row=5, experience_points=6500, proficiency_bonus=3, hp_mode='later-D8-or-fixed5',
         pact_slots=1, pact_slot_level=1, invocations=1, cantrips=2, prepared=2),
    dict(before={'fighter': 4, 'warlock': 1}, after={'fighter': 4, 'warlock': 2}, warlock_row=2,
         total_row=6, experience_points=14000, proficiency_bonus=3, hp_mode='later-D8-or-fixed5',
         pact_slots=2, pact_slot_level=1, invocations=3, cantrips=2, prepared=3),
]


def level_fixture(level, row):
    # row supplied only after independent PDF comparator verifies it.
    values = {k.rsplit('/', 2)[-2] + '/' + k.rsplit('/', 1)[-1]: v for k, v in row['parameters'].items()}
    return (f'PDF23–26/70–72/76: Warlock{level}, exact source row {values}. '
            + ('First-character level1 OR distinct later-entry contract with actual eligibility; ' if level == 1 else f'Only exact adjacent Warlock{level-1}->{level}; no first-character creation; ')
            + 'own HP adapter mandatory, slots/Invocation/preparation by Warlock level; XP/live PB by complete canonical285 table at authoritative sum of admitted post-transition class levels. '
            'A first-class Warlock lineage does not gain a blanket multiclass CHA13 prerequisite on ordinary advancement; only actual later-entry/multiclass variants consume their own source eligibility. '
            'Positive Fighter4/Warlock1->Warlock2 requires total6/14000XP/PB3; Fighter4->Warlock1 total5/6500XP/PB3/later HP. '
            'Reject booleans/decimals/zero/nonadjacent/total>20/unqualified combinations. Apply every granted feature, nested option and replacement exactly once; '
            'repeated ASIs4/8/12/16 and Arcanum11/13/15/17 retain independent grants, not a single initial receipt. No grant inferred from empty row features.')
