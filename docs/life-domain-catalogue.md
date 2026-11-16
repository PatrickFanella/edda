# Life Domain metadata (#310)

`internal/rules/srd521/testdata/character-life-domain.json` adapts exactly six
canonical identities from pinned SRD 5.2.1 PDF p40: Life Domain, Disciple of Life,
Life Domain Spells, Preserve Life, Blessed Healer and Supreme Healing. Seven
nested prerequisite records retain eligibility at Cleric levels 3, 6 and 17 and
Preserve Life's requirement to spend one canonical Channel Divinity use.
The composition manifest references the complete Cleric catalogue without
copying its records. Global advancement remains separate.

The minimal contract addition, `always_prepared`, retains four fixed table rows
at Cleric levels 3, 5, 7 and 9 and all ten spell references. These are fixed grants,
not choices. Every spell key is also a required mechanic and remains unavailable;
this batch imports no spell definitions. Validation checks feature ownership,
ordered levels, nonempty rows, unique required keys and unknown JSON fields.
Existing source, graph, count, numeric and readiness guards still apply.

Disciple of Life retains the spell-slot requirement, recipient, casting-turn
limit and 2 + slot level bonus. Preserve Life retains the Magic action, Holy
Symbol, Channel Divinity cost, five-times-Cleric-level pool, Bloodied targets
within 30 feet (including self), division among targets and half-maximum HP cap.
Blessed Healer retains slot casting, healing one or more other creatures,
immediate self-healing and 2 + slot level. Supreme Healing retains spells and
Channel Divinity, replacement of every healing die with its maximum, and the
source's 2d6 = 12 example. No older-edition restrictions are added.

The catalogue retains the pinned hash, CC-BY-4.0 attribution and adaptation
notice. Offline tests assert all identities/pages, predicates, mechanics,
numbers, spell rows, nested references, composition uniqueness and malformed
metadata rejection. The direct PDF comparator is:

```sh
rtk proxy python3 scripts/verify_srd_life_domain.py /tmp/edda-wave1/SRD_CC_v5.2.1.pdf
```

Validation grants no runtime support. Fixed Cleric creation and Cure Wounds
remain on their existing paths, with a regression check for ordinary dice,
healing and slot expenditure after metadata validation. There are no runtime
consumers, actions, automatic Life bonuses, coverage changes or Full Rules
promotion. Database, browser, inference and production checks are outside this
metadata-only batch.
