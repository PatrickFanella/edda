# Cleric and Life Domain implementation decision (#323)

This decision publishes bounded implementation work. It does not implement Cleric features, change production entitlements, promote coverage or qualify Full Rules. Keep #323 open until the parent merges its decision PR. Independent acceptance, hosted CI and merging belong to the parent.

The inspected baseline is `32252f64380eb49dd47f7d1f53171b6a8791624e`, including the accepted Barbarian decision #500 and multiclass metadata #320/#426 plus raw JSON repair #499. Closed #285, #288 and #310 supply advancement, Cleric and Life Domain metadata; they do not supply executable class behavior. The authoritative repository is GitHub; Gitea remains fetch-only.

The English SRD 5.2.1 PDF at `/tmp/edda-wave1/SRD_CC_v5.2.1.pdf` matches SHA-256 `8974902d109d6e63672d7c490bde9ccf052410503d9cfa768237154fbc5e3d87`. Direct source reading covers pp36–40, character creation/advancement/multiclass pp19–25, feats pp87–88, spellcasting pp104–106, named Life spell pages107/111/121/122/137/144/147/148/159, Wish p175, and relevant rest/condition glossary pp174–191. The catalogue comparators independently check all class table cells, nested choices, spell schedule and numeric improvements. Retain CC-BY-4.0 attribution and adaptation notices in `docs/srd-5.2.1-conformance.md`; descriptions here are paraphrases and implementation decisions are Edda additions.

## Current support and its limits

`creation.go` admits the fixed Human Acolyte Cleric, with fixed standard array, origin/feat/proficiency/equipment choices, Protector, three recorded class cantrips, four prepared spells, and Magic Initiate choices. This is not general Cleric creation. `character.go` admits Cleric level1 with explicit casting resources and rejects higher Cleric levels; recorded choice validation does not execute those choices. Its ordinary-equipment/training projection does not qualify all Cleric starting equipment or all Divine Order choices. `advancement.go` remains the narrow Fighter transition.

`internal/game/spell_preview.go` and `ally_spell.go` supply bounded Cure Wounds previews and confirmation contracts. The current coverage artefact reports Cure Wounds on self or one prepared willing conscious adjacent ally outside encounters, Shield of Faith on self, and bounded concentration/rest fixtures. #351 adds an isolated recipient effect fixture; #382 still owns production recipient eligibility. Neither catalogue import nor this decision establishes those grants.

The web `RulesSheet.tsx` names the supported Cleric fixture and unavailable effects; `ActionPreviewPanel.tsx` exposes bounded spells and rest choices. The TUI shares engine confirmation interfaces in `tui/app.go`, including explicit faith recipient preview/confirm, rather than implementing a second rules engine. This inspection is source evidence, not a new live browser or device qualification. Every new child requires explicit web and TUI availability or unavailable reason, confirmation/cancel/reconnect fixtures and changed web desktop/phone journeys before admission.

## Ownership and qualification

Cleric gate **#551 is a native child of #323**, not another child of #283. Every listed reused terminal and every ungrouped terminal directly blocks #551; fourteen Life spell/removal terminals reach #551 through the bounded Life spell gate. #551 directly blocks #337. The flattened 111 prerequisites become 98 direct blockers plus one explicit bounded group; no requirement is omitted. #323 itself is a prerequisite of #551. This separates the accepted decision from implementation and avoids overflowing #283's sub-issue limit. The gate consumes external world/spell/item qualification #183/#279/#280 as well as concrete effect terminals. The receipt checks the entire mandatory path, GitHub's 100-blocker and 100-sub-issue limits, prerequisite acyclicity and containment acyclicity separately.

**#382 remains a native child of #323**, blocked by #351 and directly blocking #279/#280/#337. Adding it as a prerequisite of #551 does not change its original parent or edges. Barbarian gate #498, Bard gate #501 and historical feat #334 direct gates remain intact. No class gate is treated as permission to remove older gates.

Reuse #427/#428 completed-rest events under #272; #429's reusable XP/PB/CON transaction under #336 with a separate Cleric D8 adapter; #430 proficiency provenance under #333; #487/#488 condition contracts under #269/#270; and source feat decision #334 plus exact selection/effect terminals #384–401. Cleric does not gain Weapon Mastery: #489–496 remain owned by #278 and are not unnecessary Cleric prerequisites.

Bard shared publication is consumed at its exact live scopes: #502 is the **Bard** casting/instrument adapter, so Cleric has its own Wisdom/Holy Symbol adapter; #503 owns Bard/Cleric/Druid/Wizard source-list membership; #504 ends trusted Paralyzed/Poisoned/Stunned instances; #505 owns Power Word Heal; #506 owns Power Word Kill. This decision creates no overlapping generic list, casting or named-spell owner. #505 is an exact reused Cleric level9-list effect; #506 remains available through its spell owner for Wish duplication, rather than gaining a Cleric-list identity.

List membership does not execute a spell. The Life spell schedule consumes nine new spell resolvers plus existing #352 Bless, with five bounded Greater Restoration removal adapters. Other selected spells require their exact #277-owned resolver and #279 qualification before availability. #273 retains ritual/costly-component ownership; #278/#280 retain items; #181/#185 and #183 retain spell-created creature/world lifecycle and qualification. Divine Intervention can select only a qualified effect; it cannot bypass an unavailable resolver. Wish duplication likewise requires the exact qualified effect contract, including world/item/feat dependencies. Missing broad spell coverage blocks qualification rather than being silently excluded.

Each mechanic key has one behavior owner in `issue-323-cleric-scope.json`; all source records include nested choices, options, prerequisites, numbers and capacities. A record can consume several owners. Level applications own once-only grants and consume behavior resolvers; they never implement the effects again. Repeated ASI grants are one resolver with four separate transition applications. High-level isolated test state grants no production admission.

## Source feature boundaries

Cleric entry distinguishes first class from later class: full saves/skills/weapons/equipment only at first class, reduced later-class armor/Hit Die/features with Wisdom13 and existing class ability requirements. Protector and Thaumaturge are exclusive, with Thaumaturge's extra cantrip separate from table counts and INT Arcana/Religion bonus minimum1. Cantrip replacement is at most one per Cleric level; prepared replacement is any number only after a completed Long Rest. Always-prepared domain spells stay outside the table count.

Channel Divinity owns one cost per chosen effect and exact capacity/recovery. Divine Spark healing and damage have separate bounded terminals, using the same scaling dice. Turn Undead targeting/saves and condition/movement/ending lifecycle are separate terminals. Sear rolls one Wisdom-scaled pool for failed targets inside the original Turn receipt, with a trusted exception preventing its own damage from ending Turn. Ordinary later damage still ends it.

Blessed Strikes has a separate exclusive choice and separate Divine Strike/Potent behaviors. Each selected branch has its own level14 improvement; no older-subclass stacking or simultaneous option grant. Divine Intervention casting and recovery are separate; Greater Intervention consumes nine spell-owned Wish terminals. Wish stress and its day-based recovery remain distinct from Cleric's 2d4 completed-Long-Rest lock.

Life Domain preparation, Disciple, Preserve Life, Blessed Healer and Supreme Healing each have one behavior terminal. Disciple is slot healing on casting turn; Blessed Healer is once after a slot spell heals others; neither applies to slotless Intervention or Channel healing. Preserve Life uses Bloodied targets, a pool of five times Cleric level and a half-maximum cap. Supreme Healing replaces spell/Channel healing dice with maxima; it does not maximize damage, Hit Dice or item healing.

## All twenty applications

Each row consumes the exact source table, XP/PB transaction, D8 adapter, preparation and slot contracts, plus prior-level state and newly eligible features. Table-only changes remain mandatory even on rows without named features.

|Cleric level|PB|Cantrips|Prepared|Channel capacity|Slots 1–9|Application|
|---|---|---|---|---|---|---|
|1|+2|3|4|0|2/0/0/0/0/0/0/0/0|#593|
|2|+2|3|5|2|3/0/0/0/0/0/0/0/0|#594|
|3|+2|3|6|2|4/2/0/0/0/0/0/0/0|#595|
|4|+2|4|7|2|4/3/0/0/0/0/0/0/0|#596|
|5|+3|4|9|2|4/3/2/0/0/0/0/0/0|#597|
|6|+3|4|10|3|4/3/3/0/0/0/0/0/0|#598|
|7|+3|4|11|3|4/3/3/1/0/0/0/0/0|#599|
|8|+3|4|12|3|4/3/3/2/0/0/0/0/0|#600|
|9|+4|4|14|3|4/3/3/3/1/0/0/0/0|#601|
|10|+4|5|15|3|4/3/3/3/2/0/0/0/0|#602|
|11|+4|5|16|3|4/3/3/3/2/1/0/0/0|#603|
|12|+4|5|16|3|4/3/3/3/2/1/0/0/0|#604|
|13|+5|5|17|3|4/3/3/3/2/1/1/0/0|#605|
|14|+5|5|17|3|4/3/3/3/2/1/1/0/0|#606|
|15|+5|5|18|3|4/3/3/3/2/1/1/1/0|#607|
|16|+5|5|18|3|4/3/3/3/2/1/1/1/0|#608|
|17|+6|5|19|3|4/3/3/3/2/1/1/1/1|#609|
|18|+6|5|20|4|4/3/3/3/3/1/1/1/1|#610|
|19|+6|5|21|4|4/3/3/3/3/2/1/1/1|#611|
|20|+6|5|22|4|4/3/3/3/3/2/2/1/1|#612|

## Published one-PR terminals

|Issue|Native parent|Bounded outcome|
|---|---|---|
|#552|#551|Apply Cleric D8 or fixed-five HP through shared advancement|
|#553|#551|Confirm first-class Cleric traits and two skills|
|#554|#551|Confirm either Cleric starting equipment package|
|#555|#551|Validate Cleric later-class Wisdom and reduced traits|
|#556|#551|Route Cleric Wisdom and Holy Symbol casts through authoritative receipts|
|#557|#551|Confirm Cleric cantrips and prepared-spell changes|
|#558|#551|Project all Cleric slot columns and Long Rest recovery|
|#559|#551|Confirm Protector or Thaumaturge at Cleric entry|
|#560|#551|Resolve Thaumaturge extra cantrip and Intelligence check bonus|
|#561|#551|Track Channel Divinity capacity and completed-rest recovery|
|#562|#551|Confirm Divine Spark healing on another visible creature|
|#563|#551|Confirm Divine Spark Necrotic or Radiant save damage|
|#564|#551|Confirm Turn Undead chosen targets and Wisdom saves|
|#565|#551|Resolve Turn Undead movement and exact ending events|
|#566|#551|Apply Sear Undead once within the Turn Undead receipt|
|#567|#551|Confirm Life Domain at Cleric level three|
|#568|#551|Offer qualified Cleric feats at four eight twelve sixteen|
|#569|#551|Confirm one Blessed Strikes option without legacy stacking|
|#570|#551|Confirm Divine Strike once on own-turn weapon hit|
|#571|#551|Add Wisdom to eligible Cleric cantrip damage|
|#572|#551|Apply improved Divine Strike two-die scaling|
|#573|#551|Confirm improved Potent vitality after Cleric cantrip damage|
|#574|#551|Confirm Divine Intervention eligible spell and same-action casting|
|#575|#551|Recover Divine Intervention on exact completed Long Rests|
|#576|#551|Offer Cleric nineteen Epic Boon or another qualified feat|
|#577|#551|Confirm Greater Divine Intervention Wish option|
|#578|#551|Apply exact Life Domain always-prepared spell schedule|
|#579|#551|Apply Disciple of Life only to slot healing on casting turn|
|#580|#551|Confirm Preserve Life allocation among Bloodied creatures|
|#581|#551|Apply Blessed Healer once after slot healing others|
|#582|#551|Replace spell and Channel healing dice at Life seventeen|
|#583|#191|Resolve Wish duplicate a qualified spell|
|#584|#191|Resolve Wish create one bounded mundane object|
|#585|#191|Resolve Wish restore health and Greater Restoration effects|
|#586|#191|Resolve Wish grant permanent chosen damage resistance|
|#587|#191|Resolve Wish grant eight-hour spell immunity|
|#588|#191|Resolve Wish replace one eligible feat|
|#589|#191|Resolve Wish confirm one recent roll redo|
|#590|#193|Resolve Wish record explicit GM Wish adjudication|
|#591|#193|Resolve Wish persist Wish stress and recovery|
|#593|#551|Confirm Cleric level one creation|
|#594|#551|Confirm Cleric level1 to2 transition|
|#595|#551|Confirm Cleric level2 to3 transition|
|#596|#551|Confirm Cleric level3 to4 transition|
|#597|#551|Confirm Cleric level4 to5 transition|
|#598|#551|Confirm Cleric level5 to6 transition|
|#599|#551|Confirm Cleric level6 to7 transition|
|#600|#551|Confirm Cleric level7 to8 transition|
|#601|#551|Confirm Cleric level8 to9 transition|
|#602|#551|Confirm Cleric level9 to10 transition|
|#603|#551|Confirm Cleric level10 to11 transition|
|#604|#551|Confirm Cleric level11 to12 transition|
|#605|#551|Confirm Cleric level12 to13 transition|
|#606|#551|Confirm Cleric level13 to14 transition|
|#607|#551|Confirm Cleric level14 to15 transition|
|#608|#551|Confirm Cleric level15 to16 transition|
|#609|#551|Confirm Cleric level16 to17 transition|
|#610|#551|Confirm Cleric level17 to18 transition|
|#611|#551|Confirm Cleric level18 to19 transition|
|#612|#551|Confirm Cleric level19 to20 transition|
|#613|#551|Confirm later-class Cleric first-level application|
|#614|#191|Resolve Aid|
|#615|#191|Resolve Cure Wounds beyond existing narrow targets|
|#616|#191|Resolve Lesser Restoration one named condition|
|#617|#191|Resolve Mass Healing Word visible recipients|
|#618|#191|Resolve Revivify one recent death|
|#619|#191|Resolve Aura of Life emanation and turn healing|
|#620|#191|Resolve Death Ward first-trigger replacement|
|#621|#191|Resolve Greater Restoration one selected effect|
|#622|#191|Resolve Mass Cure Wounds chosen sphere recipients|
|#623|#193|End Greater Restoration one Exhaustion level|
|#624|#193|End Greater Restoration Charmed or Petrified|
|#625|#193|End Greater Restoration one curse and cursed-item Attunement|
|#626|#193|End Greater Restoration one ability-score reduction|
|#627|#193|End Greater Restoration one maximum-HP reduction|
|#628|#551|Qualify fourteen exact Life spell and restoration terminal contracts|

## Parent serial application and verification

The parent has applied the four exact marked supplements in `issue-323-parent-supplements.json` serially to #47/#54/#283/#337, after fresh body reads. The sealed receipt includes each original prefix and its appended block. Future authorized supplements must preserve these exact sections; never replace a shared body from an old snapshot.

`issue-323-link-verification.json` is the final live receipt. `scripts/verify_cleric_scope.py` compares exact source snapshots and evidence hashes, every owned title/body, native parent/prerequisite set, Task/Project Size, all mandatory gate paths, legacy edges and both DAGs. Its corruption probes reject source changes, missing keys, bodies/titles, missing terminal and source edges, wrong parents/sizes and both cycle types. `test_verify_cleric_scope.py` parses query parameters and caps synthetic pagination at100 calls; repeated pages fail immediately. All live reads are bounded and have repeated-page/cursor guards.

The required Go suite is one full `GOMAXPROCS=2 go test -p 2 -short -race ./...`; vet runs with default limits. Source comparators cover Cleric, Life Domain, advancement, multiclass and the four feat families. The known unavailable linter shim is not repaired and no host/toolchain changes are made. Runtime/PostgreSQL/browser/device work belongs to implementation children; this decision does not claim it ran those journeys.
