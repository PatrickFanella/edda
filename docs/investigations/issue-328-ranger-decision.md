# Ranger / Hunter source and implementation reconciliation (#328)

This is a decision and publication change. Ranger and Hunter remain unavailable.
No runtime, grants, coverage, fixed Fighter/Cleric fixtures or production state
are changed. Native publication and final sealed readback are complete against the accepted
Paladin release. The root has published all four exact parent appends; fresh live verification with required supplements passes.

Prepared base: `3042172c935bdefa606c008c8ddda277b7548fd8` (accepted Monk #843).
The pinned SRD5.2.1 PDF SHA256 is
`8974902d109d6e63672d7c490bde9ccf052410503d9cfa768237154fbc5e3d87`.
PDF57-59 Ranger,60 full spell list,61 Hunter were extracted directly. Eight
independent PDF comparators cover Ranger, Hunter, advancement, multiclass and
all four feat catalogues. CC-BY-4.0 attribution/adaptations follow
[the source register](../srd-5.2.1-conformance.md).

Canonical counts checked: **72 Ranger +21 Hunter =93 records**. The JSON matrix
retains each exact serialized source record and all216 contextual mechanic and
option key occurrences, with rules/game/engine/combat/database/handler/web/TUI
findings. Repeated shared keys retain their source contexts. Every named feature
is owned by its execution leaf, with a native prerequisite path from its granting
row; ASI and selectors do not substitute for behavior.

`creation.go` admits only fixed Human Soldier Fighter / Human Acolyte Cleric.
`advancement.go` and `advancement_preview.go` implement Fighter1→2, not Ranger.
Database storage and the web sheet cannot admit Ranger. The TUI character view
is a placeholder; its launcher enters the shared world flow, without Ranger key
consumption.
Source catalogue and label presence therefore remain metadata evidence only.

Class tables use **Ranger level** for features, preparations, slots and Favored
Enemy counts. XP/livePB independently select the COMPLETE canonical advancement
table by sum of authoritative admitted class levels AFTER transition. Table PB
is single-class context only. Fighter4+Ranger1 reaches total5/6500XP/PB3 with
later-class D10/fixed6 HP. Fighter4/Ranger1→Ranger2 reaches total6/14000XP/PB3
with class2 features. Unsupported combinations and totals>20 stay unavailable.

The published prerequisite gateway [#968](https://github.com/subculture-collective/edda/issues/968)
is Task/M, parent336, sole blocker429. Readback preserves429's full body,
prerequisites and99old outgoing edges;968 fills slot100. It qualifies the
existing protocol and owns no transaction, die or class eligibility. Own Ranger
HP adapter consumes968; every level/later-entry contract consumes ownHP.
Monk783 and Paladin851 retain their paths. No more direct429consumers are legal.

The exact source spell roster is48Ranger spells plus11Druid cantrips. Existing
bounded spell owners are reused; FOURCLASS membership503 is not Ranger membership.
New spell contracts split lifecycle, movement, damage and world disclosures.
Hunter's Mark uses attack-roll hits/Force damage, not the old weapon-only variant;
transfer requires target0HP before expiry and spends a BonusAction, not a new cast.
All feature fixtures require isolated entitlements and grant no production access.

## All twenty class rows

| Ranger level | Class-context PB | Prepared | Favored Enemy | Slots 1/2/3/4/5 | Features |
|---|---|---|---|---|---|
| 1 | +2 | 2 | 2 | 2/0/0/0/0 | Spellcasting, Favored Enemy, Weapon Mastery |
| 2 | +2 | 3 | 2 | 2/0/0/0/0 | Deft Explorer, Fighting Style |
| 3 | +2 | 4 | 2 | 3/0/0/0/0 | Ranger Subclass |
| 4 | +2 | 5 | 2 | 3/0/0/0/0 | Ability Score Improvement |
| 5 | +3 | 6 | 3 | 4/2/0/0/0 | Extra Attack |
| 6 | +3 | 6 | 3 | 4/2/0/0/0 | Roving |
| 7 | +3 | 7 | 3 | 4/3/0/0/0 | Hunter subclass feature |
| 8 | +3 | 7 | 3 | 4/3/0/0/0 | Ability Score Improvement |
| 9 | +4 | 9 | 4 | 4/3/2/0/0 | Expertise |
| 10 | +4 | 9 | 4 | 4/3/2/0/0 | Tireless |
| 11 | +4 | 10 | 4 | 4/3/3/0/0 | Hunter subclass feature |
| 12 | +4 | 10 | 4 | 4/3/3/0/0 | Ability Score Improvement |
| 13 | +5 | 11 | 5 | 4/3/3/1/0 | Relentless Hunter |
| 14 | +5 | 11 | 5 | 4/3/3/1/0 | Nature’s Veil |
| 15 | +5 | 12 | 5 | 4/3/3/2/0 | Hunter subclass feature |
| 16 | +5 | 12 | 5 | 4/3/3/2/0 | Ability Score Improvement |
| 17 | +6 | 14 | 6 | 4/3/3/3/1 | Precise Hunter |
| 18 | +6 | 14 | 6 | 4/3/3/3/1 | Feral Senses |
| 19 | +6 | 15 | 6 | 4/3/3/3/2 | Epic Boon |
| 20 | +6 | 15 | 6 | 4/3/3/3/2 | Foe Slayer |

## Published bounded owners and native routing

Published **123 Task/M issues**: protocol gateway968,113 implementation leaves
and9 finite qualification gates. The final published-supplement readback covers940issues.
Every owned title/body/parent/prerequisite and Task/Size matched; both DAGs and
all native caps passed. After the accepted Paladin merge952fab917329e00cceebf46a8bacd59cb1c30e99,
the root released four complete parent prefixes. Fresh live readback matched
all four released hashes and the final audited reused bodies. The final receipt
is `issue-328-link-verification.json`; the earlier unsealed summary is retained
only as checkpoint provenance.

| Role | Native issues |
|---|---|
| Protocol qualification | #968, parent336, sole blocker429 |
| Ranger class qualification | #970, parent328, blocks337 |
| Class and level cohorts | #971/#972, parent328 |
| Shared qualification | #973, parent328, blocks279/280 |
| Spell cohorts,15/15/15/12 | #974–977, parent277 |
| Exact execution dependency qualification | #978, parent328, blockers61/64/65/66/67/273/353 |
| Own HP and later entry | #979/#983 |
| All20 level applications | #1015–1034 |
| New bounded spell contracts | #1035–1091 |

The explicit source/owner matrix and exact native paths are in
`issue-328-ranger-scope.json`. Newly shared leaves route through a spell cohort
and973 to279/280, and through973→970→337. Class/level cohorts route through970
to337. Gateway968 is mandatory through ownHP979; no new direct429 consumer.
The execution dependency gate is build/qualification coupling and adds no
runtime casting/concentration/encounter precondition to ordinary leaves.
Individual leaf tests against actual qualified outputs remain mandatory.

- [#970](https://github.com/subculture-collective/edda/issues/970) `class-gate`: Qualify finite Ranger class gate; parent#328; exact blockers #328, #971, #972, #973.
- [#971](https://github.com/subculture-collective/edda/issues/971) `class-cohort`: Qualify finite Ranger class cohort; parent#328; exact blockers #979, #980, #981, #982, #983, #984, #985, #986, #987, #988, #989, #990, #991, #992, #993, #994, #995, #996, #997, #998, #999, #1000, #1001, #1002, #1003, #1004, #1005, #1006, #1007, #1008, #1009, #1010, #1011, #1012, #1013, #1014.
- [#972](https://github.com/subculture-collective/edda/issues/972) `levels-cohort`: Qualify finite Ranger levels cohort; parent#328; exact blockers #1015, #1016, #1017, #1018, #1019, #1020, #1021, #1022, #1023, #1024, #1025, #1026, #1027, #1028, #1029, #1030, #1031, #1032, #1033, #1034.
- [#973](https://github.com/subculture-collective/edda/issues/973) `shared-gate`: Qualify finite Ranger shared gate; parent#328; exact blockers #263, #265, #355, #614, #615, #616, #618, #621, #623, #624, #625, #626, #627, #682, #692, #693, #700, #906, #909, #923, #924, #927, #934, #941, #954, #955, #956, #958, #965, #966, #974, #975, #976, #977.
- [#974](https://github.com/subculture-collective/edda/issues/974) `spell-cohort-1`: Qualify finite Ranger spell cohort 1; parent#277; exact blockers #1035, #1036, #1037, #1038, #1039, #1040, #1041, #1042, #1043, #1044, #1045, #1046, #1047, #1048, #1049.
- [#975](https://github.com/subculture-collective/edda/issues/975) `spell-cohort-2`: Qualify finite Ranger spell cohort 2; parent#277; exact blockers #1050, #1051, #1052, #1053, #1054, #1055, #1056, #1057, #1058, #1059, #1060, #1061, #1062, #1063, #1064.
- [#976](https://github.com/subculture-collective/edda/issues/976) `spell-cohort-3`: Qualify finite Ranger spell cohort 3; parent#277; exact blockers #1065, #1066, #1067, #1068, #1069, #1070, #1071, #1072, #1073, #1074, #1075, #1076, #1077, #1078, #1079.
- [#977](https://github.com/subculture-collective/edda/issues/977) `spell-cohort-4`: Qualify finite Ranger spell cohort 4; parent#277; exact blockers #1080, #1081, #1082, #1083, #1084, #1085, #1086, #1087, #1088, #1089, #1090, #1091.
- [#978](https://github.com/subculture-collective/edda/issues/978) `execution-dependencies`: Qualify finite Ranger shared execution dependency contracts; parent#328; exact blockers #61, #64, #65, #66, #67, #273, #353.
- [#979](https://github.com/subculture-collective/edda/issues/979) `hp`: Adapt Ranger own D10 fixed6 HP to shared transaction; parent#971; exact blockers #968.
- [#980](https://github.com/subculture-collective/edda/issues/980) `core`: Confirm Ranger first-character core proficiencies; parent#971; exact blockers #430.
- [#981](https://github.com/subculture-collective/edda/issues/981) `skills`: Select Ranger three first-character skills; parent#971; exact blockers #430.
- [#982](https://github.com/subculture-collective/edda/issues/982) `equipment`: Choose Ranger exact starting equipment; parent#971; exact blockers .
- [#983](https://github.com/subculture-collective/edda/issues/983) `later-entry`: Admit exact later-class Ranger entry; parent#971; exact blockers #430, #979.
- [#984](https://github.com/subculture-collective/edda/issues/984) `spellcasting`: Prepare Ranger spells with Wisdom and Druidic Focus; parent#971; exact blockers #263, #265, #355, #428, #614, #615, #616, #618, #621, #623, #624, #625, #626, #627, #682, #692, #693, #700, #906, #909, #923, #924, #927, #934, #941, #965, #966, #978, #1002, #1003, #1035, #1036, #1037, #1038, #1039, #1040, #1041, #1042, #1043, #1044, #1045, #1046, #1047, #1048, #1054, #1055, #1056, #1057, #1058, #1059, #1060, #1061, #1062, #1063, #1064, #1065, #1066, #1067, #1068, #1069, #1071, #1072, #1073, #1074, #1078, #1080, #1081, #1082, #1083, #1084, #1086, #1087, #1088, #1089, #1090, #1091.
- [#985](https://github.com/subculture-collective/edda/issues/985) `favored-enemy`: Cast Favored Enemy exact 5.2.1 Hunter’s Mark; parent#971; exact blockers #428, #1062, #1063.
- [#986](https://github.com/subculture-collective/edda/issues/986) `weapon-mastery`: Select and replace two Ranger mastery kinds; parent#971; exact blockers #428, #489, #490, #491, #492, #493, #494, #495, #496.
- [#987](https://github.com/subculture-collective/edda/issues/987) `deft-explorer`: Apply Deft Explorer expertise and languages; parent#971; exact blockers #430.
- [#988](https://github.com/subculture-collective/edda/issues/988) `fighting-style`: Choose Ranger Fighting Style or Druidic Warrior; parent#971; exact blockers #384, #385, #386, #387, #388, #389, #390, #391, #392, #393, #394, #395, #396, #397, #398, #399, #400, #401, #1004.
- [#989](https://github.com/subculture-collective/edda/issues/989) `ranger-subclass`: Select Hunter and exact Ranger subclass timing; parent#971; exact blockers #1005.
- [#990](https://github.com/subculture-collective/edda/issues/990) `ability-score-improvement`: Select Ranger qualified ASI at four class levels; parent#971; exact blockers #384, #385, #386, #387, #388, #389, #390, #391, #392, #393, #394, #395, #396, #397, #398, #399, #400, #401, #427, #428, #487, #778, #779, #962, #984.
- [#991](https://github.com/subculture-collective/edda/issues/991) `extra-attack`: Resolve Ranger two own-turn Attack attacks; parent#971; exact blockers #978.
- [#992](https://github.com/subculture-collective/edda/issues/992) `roving`: Apply Roving nonheavy Speed climb and swim; parent#971; exact blockers #978.
- [#993](https://github.com/subculture-collective/edda/issues/993) `expertise`: Select two Ranger Expertise skills; parent#971; exact blockers #430.
- [#994](https://github.com/subculture-collective/edda/issues/994) `relentless-hunter`: Protect Hunter’s Mark concentration against damage only; parent#971; exact blockers #487, #978, #1062.
- [#995](https://github.com/subculture-collective/edda/issues/995) `natures-veil`: Spend Nature’s Veil uses for exact Invisible duration; parent#971; exact blockers #428, #962.
- [#996](https://github.com/subculture-collective/edda/issues/996) `precise-hunter`: Apply Precise Hunter marked-target attack advantage; parent#971; exact blockers #978, #1063.
- [#997](https://github.com/subculture-collective/edda/issues/997) `feral-senses`: Resolve Feral Senses exact thirty-foot Blindsight; parent#971; exact blockers #962, #978.
- [#998](https://github.com/subculture-collective/edda/issues/998) `epic-boon`: Select qualified Ranger Epic Boon at class nineteen; parent#971; exact blockers #384, #385, #386, #387, #388, #389, #390, #391, #392, #393, #394, #395, #396, #397, #398, #399, #400, #401, #427, #428, #487, #778, #779, #962, #984.
- [#999](https://github.com/subculture-collective/edda/issues/999) `foe-slayer`: Replace own Hunter’s Mark damage die with D10; parent#971; exact blockers #978, #1062.
- [#1000](https://github.com/subculture-collective/edda/issues/1000) `tireless-hp`: Resolve Tireless Magic action temporary HP; parent#971; exact blockers #428, #978.
- [#1001](https://github.com/subculture-collective/edda/issues/1001) `tireless-rest`: Resolve Tireless Short Rest Exhaustion reduction; parent#971; exact blockers #427, #623.
- [#1002](https://github.com/subculture-collective/edda/issues/1002) `slots`: Apply Ranger class slot counts and combined multiclass slot table; parent#971; exact blockers #428, #978.
- [#1003](https://github.com/subculture-collective/edda/issues/1003) `membership`: Validate exact Ranger source spell list membership; parent#971; exact blockers #293, #315.
- [#1004](https://github.com/subculture-collective/edda/issues/1004) `druidic-cantrips`: Prepare and replace two Druidic Warrior Ranger cantrips; parent#971; exact blockers #954, #955, #956, #958, #1003, #1049, #1050, #1051, #1052, #1053, #1070, #1075, #1076, #1077, #1079, #1085.
- [#1005](https://github.com/subculture-collective/edda/issues/1005) `hunter-identity`: Validate chosen Hunter class-three identity; parent#971; exact blockers .
- [#1006](https://github.com/subculture-collective/edda/issues/1006) `hunters-lore`: Disclose Hunter’s Lore exact own marked defenses; parent#971; exact blockers #1005, #1063.
- [#1007](https://github.com/subculture-collective/edda/issues/1007) `colossus-slayer`: Resolve Colossus Slayer once per turn extra weapon damage; parent#971; exact blockers #978, #1005.
- [#1008](https://github.com/subculture-collective/edda/issues/1008) `horde-breaker`: Resolve Horde Breaker same-weapon distinct nearby target; parent#971; exact blockers #978, #1005.
- [#1009](https://github.com/subculture-collective/edda/issues/1009) `escape-the-horde`: Apply Escape the Horde Opportunity Attack disadvantage; parent#971; exact blockers #978, #1005.
- [#1010](https://github.com/subculture-collective/edda/issues/1010) `multiattack-defense`: Apply Multiattack Defense same-creature remainder-turn disadvantage; parent#971; exact blockers #978, #1005.
- [#1011](https://github.com/subculture-collective/edda/issues/1011) `superior-hunters-prey`: Resolve Superior Hunter’s Prey current mark damage to nearby visible creature; parent#971; exact blockers #978, #999, #1005, #1062.
- [#1012](https://github.com/subculture-collective/edda/issues/1012) `superior-hunters-defense`: Resolve Superior Hunter’s Defense reaction resistance through turn end; parent#971; exact blockers #978, #1005.
- [#1013](https://github.com/subculture-collective/edda/issues/1013) `hunters-prey`: Choose or rest-replace Hunter’s Prey option; parent#971; exact blockers #427, #428, #1007, #1008.
- [#1014](https://github.com/subculture-collective/edda/issues/1014) `defensive-tactics`: Choose or rest-replace Defensive Tactics option; parent#971; exact blockers #427, #428, #1009, #1010.
- [#1015](https://github.com/subculture-collective/edda/issues/1015) `level-1`: Apply Ranger class level 1 first-character or later-class entry; parent#972; exact blockers #979, #980, #981, #982, #983, #984, #985, #986, #1002.
- [#1016](https://github.com/subculture-collective/edda/issues/1016) `level-2`: Apply Ranger class level 2 from class 1; parent#972; exact blockers #979, #983, #984, #985, #987, #988, #1002, #1015.
- [#1017](https://github.com/subculture-collective/edda/issues/1017) `level-3`: Apply Ranger class level 3 from class 2; parent#972; exact blockers #979, #983, #984, #985, #989, #1002, #1006, #1013, #1016.
- [#1018](https://github.com/subculture-collective/edda/issues/1018) `level-4`: Apply Ranger class level 4 from class 3; parent#972; exact blockers #979, #983, #984, #985, #990, #1002, #1017.
- [#1019](https://github.com/subculture-collective/edda/issues/1019) `level-5`: Apply Ranger class level 5 from class 4; parent#972; exact blockers #979, #983, #984, #985, #991, #1002, #1018.
- [#1020](https://github.com/subculture-collective/edda/issues/1020) `level-6`: Apply Ranger class level 6 from class 5; parent#972; exact blockers #979, #983, #984, #985, #992, #1002, #1019.
- [#1021](https://github.com/subculture-collective/edda/issues/1021) `level-7`: Apply Ranger class level 7 from class 6; parent#972; exact blockers #979, #983, #984, #985, #1002, #1014, #1020.
- [#1022](https://github.com/subculture-collective/edda/issues/1022) `level-8`: Apply Ranger class level 8 from class 7; parent#972; exact blockers #979, #983, #984, #985, #990, #1002, #1021.
- [#1023](https://github.com/subculture-collective/edda/issues/1023) `level-9`: Apply Ranger class level 9 from class 8; parent#972; exact blockers #979, #983, #984, #985, #993, #1002, #1022.
- [#1024](https://github.com/subculture-collective/edda/issues/1024) `level-10`: Apply Ranger class level 10 from class 9; parent#972; exact blockers #979, #983, #984, #985, #1000, #1001, #1002, #1023.
- [#1025](https://github.com/subculture-collective/edda/issues/1025) `level-11`: Apply Ranger class level 11 from class 10; parent#972; exact blockers #979, #983, #984, #985, #1002, #1011, #1024.
- [#1026](https://github.com/subculture-collective/edda/issues/1026) `level-12`: Apply Ranger class level 12 from class 11; parent#972; exact blockers #979, #983, #984, #985, #990, #1002, #1025.
- [#1027](https://github.com/subculture-collective/edda/issues/1027) `level-13`: Apply Ranger class level 13 from class 12; parent#972; exact blockers #979, #983, #984, #985, #994, #1002, #1026.
- [#1028](https://github.com/subculture-collective/edda/issues/1028) `level-14`: Apply Ranger class level 14 from class 13; parent#972; exact blockers #979, #983, #984, #985, #995, #1002, #1027.
- [#1029](https://github.com/subculture-collective/edda/issues/1029) `level-15`: Apply Ranger class level 15 from class 14; parent#972; exact blockers #979, #983, #984, #985, #1002, #1012, #1028.
- [#1030](https://github.com/subculture-collective/edda/issues/1030) `level-16`: Apply Ranger class level 16 from class 15; parent#972; exact blockers #979, #983, #984, #985, #990, #1002, #1029.
- [#1031](https://github.com/subculture-collective/edda/issues/1031) `level-17`: Apply Ranger class level 17 from class 16; parent#972; exact blockers #979, #983, #984, #985, #996, #1002, #1030.
- [#1032](https://github.com/subculture-collective/edda/issues/1032) `level-18`: Apply Ranger class level 18 from class 17; parent#972; exact blockers #979, #983, #984, #985, #997, #1002, #1031.
- [#1033](https://github.com/subculture-collective/edda/issues/1033) `level-19`: Apply Ranger class level 19 from class 18; parent#972; exact blockers #979, #983, #984, #985, #998, #1002, #1032.
- [#1034](https://github.com/subculture-collective/edda/issues/1034) `level-20`: Apply Ranger class level 20 from class 19; parent#972; exact blockers #979, #983, #984, #985, #999, #1002, #1033.
- [#1035](https://github.com/subculture-collective/edda/issues/1035) `spell-alarm`: Resolve Alarm effect; parent#974; exact blockers #978.
- [#1036](https://github.com/subculture-collective/edda/issues/1036) `spell-animal-friendship`: Resolve Animal Friendship effect; parent#974; exact blockers #488, #978.
- [#1037](https://github.com/subculture-collective/edda/issues/1037) `spell-animal-messenger-admission`: Resolve Animal Messenger admission; parent#974; exact blockers #978.
- [#1038](https://github.com/subculture-collective/edda/issues/1038) `spell-animal-messenger-journey`: Resolve Animal Messenger journey; parent#974; exact blockers #978, #1037.
- [#1039](https://github.com/subculture-collective/edda/issues/1039) `spell-barkskin`: Resolve Barkskin effect; parent#974; exact blockers #978.
- [#1040](https://github.com/subculture-collective/edda/issues/1040) `spell-commune-with-nature-facts`: Resolve Commune with Nature facts; parent#974; exact blockers #978.
- [#1041](https://github.com/subculture-collective/edda/issues/1041) `spell-conjure-animals-aura`: Resolve Conjure Animals aura; parent#974; exact blockers #978.
- [#1042](https://github.com/subculture-collective/edda/issues/1042) `spell-conjure-animals-movement`: Resolve Conjure Animals movement; parent#974; exact blockers #978, #1041.
- [#1043](https://github.com/subculture-collective/edda/issues/1043) `spell-conjure-animals-damage`: Resolve Conjure Animals damage; parent#974; exact blockers #978, #1041.
- [#1044](https://github.com/subculture-collective/edda/issues/1044) `spell-conjure-woodland-beings-aura`: Resolve Conjure Woodland Beings aura; parent#974; exact blockers #978.
- [#1045](https://github.com/subculture-collective/edda/issues/1045) `spell-conjure-woodland-beings-disengage`: Resolve Conjure Woodland Beings disengage; parent#974; exact blockers #781, #978, #1044.
- [#1046](https://github.com/subculture-collective/edda/issues/1046) `spell-darkvision`: Resolve Darkvision effect; parent#974; exact blockers #978.
- [#1047](https://github.com/subculture-collective/edda/issues/1047) `spell-dominate-beast-charm`: Resolve Dominate Beast charm; parent#974; exact blockers #488, #978.
- [#1048](https://github.com/subculture-collective/edda/issues/1048) `spell-dominate-beast-control`: Resolve Dominate Beast control; parent#974; exact blockers #488, #978, #1047.
- [#1049](https://github.com/subculture-collective/edda/issues/1049) `spell-druidcraft-weather`: Resolve Druidcraft weather; parent#974; exact blockers #978.
- [#1050](https://github.com/subculture-collective/edda/issues/1050) `spell-druidcraft-nature`: Resolve Druidcraft nature; parent#975; exact blockers #978, #1049.
- [#1051](https://github.com/subculture-collective/edda/issues/1051) `spell-elementalism-air-earth`: Resolve Elementalism air-earth; parent#975; exact blockers #978.
- [#1052](https://github.com/subculture-collective/edda/issues/1052) `spell-elementalism-fire-water`: Resolve Elementalism fire-water; parent#975; exact blockers #978, #1051.
- [#1053](https://github.com/subculture-collective/edda/issues/1053) `spell-elementalism-sculpt`: Resolve Elementalism sculpt; parent#975; exact blockers #978, #1051.
- [#1054](https://github.com/subculture-collective/edda/issues/1054) `spell-enhance-ability`: Resolve Enhance Ability effect; parent#975; exact blockers #978.
- [#1055](https://github.com/subculture-collective/edda/issues/1055) `spell-ensnaring-strike-hit`: Resolve Ensnaring Strike hit; parent#975; exact blockers #727, #978.
- [#1056](https://github.com/subculture-collective/edda/issues/1056) `spell-ensnaring-strike-tick-escape`: Resolve Ensnaring Strike tick-escape; parent#975; exact blockers #727, #978, #1055.
- [#1057](https://github.com/subculture-collective/edda/issues/1057) `spell-entangle`: Resolve Entangle effect; parent#975; exact blockers #727, #978.
- [#1058](https://github.com/subculture-collective/edda/issues/1058) `spell-find-traps`: Resolve Find Traps effect; parent#975; exact blockers #978.
- [#1059](https://github.com/subculture-collective/edda/issues/1059) `spell-goodberry-create`: Resolve Goodberry create; parent#975; exact blockers #978.
- [#1060](https://github.com/subculture-collective/edda/issues/1060) `spell-goodberry-eat`: Resolve Goodberry eat; parent#975; exact blockers #978, #1059.
- [#1061](https://github.com/subculture-collective/edda/issues/1061) `spell-gust-of-wind`: Resolve Gust of Wind effect; parent#975; exact blockers #978.
- [#1062](https://github.com/subculture-collective/edda/issues/1062) `spell-hunters-mark-cast`: Resolve Hunter’s Mark cast; parent#975; exact blockers #978.
- [#1063](https://github.com/subculture-collective/edda/issues/1063) `spell-hunters-mark-transfer`: Resolve Hunter’s Mark transfer; parent#975; exact blockers #978, #1062.
- [#1064](https://github.com/subculture-collective/edda/issues/1064) `spell-jump`: Resolve Jump effect; parent#975; exact blockers #978.
- [#1065](https://github.com/subculture-collective/edda/issues/1065) `spell-locate-animals-or-plants`: Resolve Locate Animals or Plants effect; parent#976; exact blockers #978.
- [#1066](https://github.com/subculture-collective/edda/issues/1066) `spell-longstrider`: Resolve Longstrider effect; parent#976; exact blockers #978.
- [#1067](https://github.com/subculture-collective/edda/issues/1067) `spell-meld-into-stone-entry`: Resolve Meld into Stone entry; parent#976; exact blockers #978.
- [#1068](https://github.com/subculture-collective/edda/issues/1068) `spell-meld-into-stone-inside`: Resolve Meld into Stone inside; parent#976; exact blockers #978, #1067.
- [#1069](https://github.com/subculture-collective/edda/issues/1069) `spell-meld-into-stone-expulsion`: Resolve Meld into Stone expulsion; parent#976; exact blockers #488, #978, #1067.
- [#1070](https://github.com/subculture-collective/edda/issues/1070) `spell-message`: Resolve Message effect; parent#976; exact blockers #978.
- [#1071](https://github.com/subculture-collective/edda/issues/1071) `spell-nondetection`: Resolve Nondetection effect; parent#976; exact blockers #978.
- [#1072](https://github.com/subculture-collective/edda/issues/1072) `spell-pass-without-trace`: Resolve Pass without Trace effect; parent#976; exact blockers #978.
- [#1073](https://github.com/subculture-collective/edda/issues/1073) `spell-plant-growth-overgrowth`: Resolve Plant Growth overgrowth; parent#976; exact blockers #978.
- [#1074](https://github.com/subculture-collective/edda/issues/1074) `spell-plant-growth-enrichment`: Resolve Plant Growth enrichment; parent#976; exact blockers #978, #1073.
- [#1075](https://github.com/subculture-collective/edda/issues/1075) `spell-poison-spray`: Resolve Poison Spray effect; parent#976; exact blockers #978.
- [#1076](https://github.com/subculture-collective/edda/issues/1076) `spell-produce-flame-light`: Resolve Produce Flame light; parent#976; exact blockers #978.
- [#1077](https://github.com/subculture-collective/edda/issues/1077) `spell-produce-flame-attack`: Resolve Produce Flame attack; parent#976; exact blockers #978, #1076.
- [#1078](https://github.com/subculture-collective/edda/issues/1078) `spell-protection-from-energy`: Resolve Protection from Energy effect; parent#976; exact blockers #978.
- [#1079](https://github.com/subculture-collective/edda/issues/1079) `spell-shillelagh`: Resolve Shillelagh effect; parent#976; exact blockers #978.
- [#1080](https://github.com/subculture-collective/edda/issues/1080) `spell-silence`: Resolve Silence effect; parent#977; exact blockers #961, #978.
- [#1081](https://github.com/subculture-collective/edda/issues/1081) `spell-speak-with-plants-facts`: Resolve Speak with Plants facts; parent#977; exact blockers #978.
- [#1082](https://github.com/subculture-collective/edda/issues/1082) `spell-speak-with-plants-terrain`: Resolve Speak with Plants terrain; parent#977; exact blockers #978, #1081.
- [#1083](https://github.com/subculture-collective/edda/issues/1083) `spell-spike-growth-terrain`: Resolve Spike Growth terrain; parent#977; exact blockers #978.
- [#1084](https://github.com/subculture-collective/edda/issues/1084) `spell-spike-growth-movement`: Resolve Spike Growth movement; parent#977; exact blockers #978, #1083.
- [#1085](https://github.com/subculture-collective/edda/issues/1085) `spell-starry-wisp`: Resolve Starry Wisp effect; parent#977; exact blockers #962, #978.
- [#1086](https://github.com/subculture-collective/edda/issues/1086) `spell-stoneskin`: Resolve Stoneskin effect; parent#977; exact blockers #978.
- [#1087](https://github.com/subculture-collective/edda/issues/1087) `spell-water-breathing`: Resolve Water Breathing effect; parent#977; exact blockers #978.
- [#1088](https://github.com/subculture-collective/edda/issues/1088) `spell-water-walk`: Resolve Water Walk effect; parent#977; exact blockers #978.
- [#1089](https://github.com/subculture-collective/edda/issues/1089) `spell-wind-wall-placement`: Resolve Wind Wall placement; parent#977; exact blockers #978.
- [#1090](https://github.com/subculture-collective/edda/issues/1090) `spell-wind-wall-damage`: Resolve Wind Wall damage; parent#977; exact blockers #978, #1089.
- [#1091](https://github.com/subculture-collective/edda/issues/1091) `spell-wind-wall-barrier`: Resolve Wind Wall barrier; parent#977; exact blockers #978, #1089.

Direct PDF104–106 audit preserves prepared ritual+10min/no slot, longer-cast
Concentration and interruption without slot expenditure, general non-Incapacitated
time-span dismissal, component/armor/clear-path rules and privately authoritative
invalid-target slot expenditure without disclosure. Missing or unauthorized
facts remain unavailable. No extra Ritual Adept feature is invented.

Tireless owns its distinct completed-Short-Rest adapter. #623 remains a trusted
Greater Restoration owner; no fake restoration event or spell cost is borrowed.
ASI/Epic composition consumes actual487/427/428/778/779/962 plus own casting984,
rather than treating planning parents as effect implementation. Existing owners
and the exact779 body/prerequisites remain unchanged.

## Verification and remaining release work

The one released `GOMAXPROCS=2 go test -p 2 -short -race ./...` and
`GOMAXPROCS=2 go vet -p 2 ./...` passed on this base. All eight pinned comparators
passed. Source comparators are rerunnable; the full suite will not be repeated
for narrow publication changes. Go1.27.3 lint shim lacks its binary; no install
or host changes were made. No database, browser/device or production journey
was run; runtime behavior has not changed.

Parent owns serial47/54/283/337 supplements, fresh independent review, final CI
and merge. The accepted post-Paladin prefixes are frozen in
`issue-328-accepted-parent-baselines.json`; the verifier hardcodes their hashes.
`python3 scripts/publish_ranger_supplements.py` is a read-only preflight. Only the
parent invokes `--apply`, which fresh-reads each target, preserves its entire
accepted prefix and later pure appends, and rejects partial/altered/duplicate
owned sections before readback. The Ranger child did not apply parent appends. Keep328 OPEN until its decision PR merges. Source/old
owner bodies remain untouched; no watch, polling, merge, deploy or closure.

Capacity projection from initial live graph plus proposed exact edges:279incoming98,
280incoming93,337incoming55,429outgoing100,283children95,328children5.
Existing primitive outgoing projections61=75,64=41,65=66,66=55,67=92,273=32,
353=68. These values also match the final sealed native readback. All four parent appends are published and sealed in the refreshed receipt. Decision closure is accepted without removing any implementation requirement; all616 owned edge-removal controls also reject after simulated closure.

The source proof also compares every one of the 57 bounded spell scopes against a readable independent oracle, including numeric/timing predicates, exact published envelopes and split-effect dependencies. All 93 source record contexts and 216 contextual keys have independent execution-owner expectations; Hunter nested alternatives remain required at their granting levels. The protocol gate independently requires its exact seven owners and 72 consumers, even if a contrary manifest also deletes path declarations. All 33 focused Python tests pass, including coherent contract/body/graph corruptions; no native contract or edge changed in these proof repairs.
