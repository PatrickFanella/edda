# Epic Boon feat source metadata (#298)

`internal/rules/srd521/testdata/character-epic-boon-feats.json` contains exactly
seven feat identities from the pinned English SRD 5.2.1 PDF page 88: Boon of
Combat Prowess, Boon of Dimensional Travel, Boon of Fate, Boon of Irresistible
Offense, Boon of Spell Recall, Boon of the Night Spirit and Boon of Truesight.
The 34-record graph uses the existing guarded character contract with no schema
or runtime changes. Its 27 owned records retain seven ability choices,
18 prerequisite/condition predicates and two single-use resources.

The source SHA-256 is
`8974902d109d6e63672d7c490bde9ccf052410503d9cfa768237154fbc5e3d87`.
Every record links to its containing feat's inventory identity at page 88.
The source object retains the artifact URL, size, page count, CC-BY-4.0 license,
attribution and adaptation notice. See [the conformance contract](srd-5.2.1-conformance.md).
The acquisition limit of one follows Parts of a Feat on page 87; none of these
seven feats has a Repeatable override.

All seven require character level 19 and increase one selected ability by 1,
capped at 30. Irresistible Offense restricts the choice to Strength or Dexterity;
Spell Recall restricts it to Intelligence, Wisdom or Charisma and also requires
the Spellcasting feature. The other five allow any of the six abilities.

The benefit metadata retains these source distinctions:

- Combat Prowess optionally replaces a missed attack roll with a hit, once until
  the start of the character's next turn.
- Dimensional Travel optionally teleports immediately after the character takes
  the Attack or Magic action, up to 30 feet to an unoccupied space they can see.
- Fate triggers on the character or another creature within 60 feet succeeding
  or failing a D20 Test. Its optional 2d4 total is a bonus or penalty to that d20
  roll. Its one use recovers on rolling Initiative or finishing a Short or Long Rest.
- Irresistible Offense always ignores Resistance for the character's Bludgeoning,
  Piercing and Slashing damage. A 20 on the attack's d20 optionally deals extra
  damage to that target, with the attack's damage type. The amount is the ability
  **score** increased by this feat. The predicate refers to that ability choice.
- Spell Recall rolls 1d4 whenever casting with a level 1–4 slot. The slot is not
  expended only when the die result equals its level. The match predicate refers
  to the casting-slot predicate, retaining the nested operand and numeric bounds.
- Night Spirit optionally gives the character the Invisible condition as a Bonus
  Action while in Dim Light or Darkness. It ends immediately after the character
  takes an action, Bonus Action or Reaction. In those light conditions, Shadowy
  Form gives Resistance to all damage except Psychic and Radiant.
- Truesight has a fixed range of 60 feet.

These facts remain opaque required mechanic keys, numeric parameters and owned
predicates, choices or resources. Validation evaluates none of them and grants
no feats, bonuses or actions. Unknown keys fail readiness. Creation, advancement,
coverage and Full Rules eligibility retain their existing code-owned boundaries.

The Go oracle independently transcribes every record and compares the complete
graph, including prerequisite bodies, operand references, ownership and empty
fields. Contrary-source fixtures mutate every mechanic and numeric field,
ability alternatives/counts, predicate bodies/operands, resource limits and
invented level scaling. They must pass the generic schema and fail source equality.
A reciprocal ownership transfer also passes schema validation but fails source
comparison. Separate guards cover strict decoding, trailing JSON, pin, license,
attribution, adaptation, duplicate IDs, source pages/families, ownership,
references, malformed keys and unknown-key readiness.

Composition appends each canonical advancement, background, class, subclass,
species and feat catalogue once, including Berserker. Reference subsets and
composition manifests are excluded to avoid duplication. Fixed Human Soldier
Fighter and Human Acolyte Cleric build, stats, sheet, feat, gold and restricted
creation regressions check that this metadata leaves runtime results unchanged.
Coverage is compared before and after and remains ineligible for Full Rules.

Verification commands (run from the repository root):

```sh
rtk proxy env GOMAXPROCS=2 go test -p 2 -short -race ./...
rtk proxy env GOMAXPROCS=2 go vet -p 2 ./...
rtk proxy env GOMAXPROCS=2 go test -p 2 -short -race ./internal/rules/srd521 -run EpicBoons
rtk proxy python scripts/verify_srd_epic_boon_feats.py /tmp/edda-wave1/SRD_CC_v5.2.1.pdf
rtk proxy gofmt -l internal/rules/srd521/catalogue_epic_boon_feats_test.go
rtk git diff --check
```

The standalone comparator checks the PDF hash, complete page-88 feat text,
page-87 default acquisition rule, exact issue-298 inventory identities/pages and
complete metadata graph. It also runs 255 contrary-source mutations against its
own comparison. It requires the local pinned PDF and `pdftotext`; Go fixtures
need neither the PDF nor network access.

No database, UI, inference or production checks apply to this metadata batch.
Independent review and hosted qualification belong to the parent. The active
Go 1.27.3 `golangci-lint` shim has no installed binary; host tooling is unchanged.
