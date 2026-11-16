# Cleric class metadata (#288)

`testdata/character-cleric.json` contains exactly 12 canonical source identities:
Cleric and Spellcasting on p36; Divine Order, Channel Divinity, Cleric Subclass,
Ability Score Improvement and Sear Undead on p37; Blessed Strikes, Divine
Intervention, Improved Blessed Strikes, Epic Boon and Greater Divine Intervention
on p38. Nested records reuse their containing heading rather than adding source
identities. `character-cleric1.json` is a reference manifest into that catalogue.

All 20 p36 progression rows retain proficiency, repeated ASI grants, subclass
feature slots at levels 6 and 17, Channel Divinity uses, cantrip/prepared-spell
counts and all nine spell-slot columns. Absent slots have no positive capacity
entry. Divine Spark dice improvements are numeric feature parameters; they are
not consumable resources. Spell-list domains refer to pp38–40 without importing
spell records. Feat and Life Domain references remain opaque domains.

The only contract extension permits a choice to own conditional subchoices.
Protector/Thaumaturge, Divine Spark and Divine Strike retain parent choice links
and unevaluated selection prerequisites. The existing reciprocal links, source
ownership, acyclicity, numeric/count and readiness checks apply unchanged.
Thaumaturge's extra cantrip is separate from the class-table count. Numeric
parameters retain the Wisdom minimums, damage dice, distances, temporary-HP
multiplier and Wish recovery dice. Feat eligibility, prepared-spell slot limits,
Divine Intervention's non-Reaction limit and conditional improvements remain
unimplemented predicates or required keys.

The pinned PDF was read directly on pp36–38, with referenced spell lists on
pp38–40, multiclass rules on pp24–25, feats on pp87–88, equipment on pp97/99,
spellcasting on pp104–106 and Wish on p175. Referenced definitions are not
imported. CC-BY-4.0 attribution and the adaptation notice remain in `source`.

`catalogue_cleric_test.go` independently transcribes every table cell and checks
exact canonical IDs/pages, nested choice domains, quantities, prerequisites,
numeric improvements, canonical composition and rejection of malformed or
unknown executable support. Run the direct offline comparison with:

```sh
rtk proxy python3 scripts/verify_srd_cleric.py /tmp/edda-wave1/SRD_CC_v5.2.1.pdf
```

The comparator verifies the PDF hash, exact identity/page inventory, all 20
extracted table rows and source anchors for conditional options. This is metadata
proof only: no creation/advancement grants, general spellcasting, Divine
Intervention, action consumer, coverage promotion or Full Rules qualification.
The fixed Cleric compatibility path is unchanged and retains its existing
regression fixtures. No database or browser checks are required by this batch.
