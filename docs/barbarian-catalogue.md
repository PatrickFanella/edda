# Barbarian class metadata (#286)

`internal/rules/srd521/testdata/character-barbarian.json` adapts exactly 21
source identities from SRD 5.2.1 pp28–30: Barbarian and its 20 named class
feature headings. Improved Brutal Strike at level 13 and level 17 have separate
stable IDs, source identities, owners and grants despite sharing a printed name.
Referenced subclasses, feats, weapons, equipment and skills remain opaque keys.
No Berserker feature or adjacent class definition is imported.

All 20 class rows retain proficiency bonuses, Rage uses, Rage damage and weapon
mastery counts. Repeated ASI grants at 8/12/16 reuse the level-4 identity;
subclass slots at 6/10/14 remain unimplemented keys. Global character advancement
is independent. The existing parameter map now accepts positive numeric metadata
on a level record with class progression, in addition to classes/features/feats.
This stores the Rage damage bonus without inventing a consumable resource.
Existing required-key and numeric guards apply to those table parameters.

Core traits retain Strength, D12, both saves, armor/weapon training, all six
skill alternatives, the two equipment bundles and their quantities, plus the
separate first-class and multiclass trait grants described on p28. Rage records
Bonus Action entry, the Heavy armor gate, resistance types, Strength attack
bonus conditions, Strength advantage, casting and Concentration restrictions,
end/extension conditions, three extension alternatives, the ten-minute cap,
and one-use Short Rest/all-use Long Rest recovery. Persistent Rage separately
records Initiative recovery once per Long Rest and replaces round extension
and the Incapacitated end condition with its Unconscious end condition; Heavy
armor still ends it.

Nested choices and prerequisites retain Primal Knowledge skill options and its
conditional Strength substitutions, melee-only mastery and replacement, feat
qualification, and Brutal Strike selection and attack eligibility. Numeric
parameters retain base AC, attack count, movement, dice, distances, Relentless
Rage DC escalation/reset and replacement HP multiplier, Sundering Blow bonus,
and Primal Champion increases/caps. Brutal Strike retains Forceful/Hamstring
Blow conditions, movement and nonstacking behavior; the two improvements retain
Staggering/Sundering Blow and the level-17 two-different-effects/2d10 upgrade.
The base effect-choice count increases from one to two at level 17. Improvements
refer to the existing option domain rather than importing subclass rules.

`catalogue_barbarian_test.go` independently transcribes every table cell,
canonical source ID/page, option domain, equipment quantity, numeric feature
parameter and prerequisite operand. A complete-record digest also freezes every
required mechanic key, graph edge and condition; changing it requires reviewing
the source facts, not merely regenerating a snapshot. Negative fixtures reject
malformed source/owner/grant/count/numeric metadata and unknown JSON fields;
unknown mechanic keys remain valid metadata but fail readiness. Composition with
advancement, Fighter, Champion, Cleric, Life Domain, backgrounds and Fighting
Styles checks uniqueness and graph validation. Existing fixed-build/combat and
coverage regression fixtures remain part of the full short/race suite.

Run the offline PDF comparator with:

```sh
rtk proxy python3 scripts/verify_srd_barbarian.py /tmp/edda-wave1/SRD_CC_v5.2.1.pdf
```

It checks SHA-256
`8974902d109d6e63672d7c490bde9ccf052410503d9cfa768237154fbc5e3d87`,
the exact inventory identities/pages, all extracted progression cells and grants,
and source anchors for feature conditions/options/numbers. The JSON retains the
pinned source, CC-BY-4.0 attribution and adaptation notice.

This batch adds no runtime consumer, creation or advancement enablement, Rage
or combat action, automatic bonus/grant/inventory change, coverage promotion or
Full Rules qualification. No database, browser, production or inference check
is needed for this offline metadata scope.
