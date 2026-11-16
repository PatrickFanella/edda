# General feat source metadata (#300)

`internal/rules/srd521/testdata/character-general-feats.json` contains exactly
Ability Score Improvement and Grappler from PDF page 87 of the pinned English
SRD 5.2.1, with nine owned choice and prerequisite records. It is offline metadata.

The source SHA-256 is
`8974902d109d6e63672d7c490bde9ccf052410503d9cfa768237154fbc5e3d87`.
Every record links to its containing feat in the planning inventory at page 87.
The source object retains the artifact URL, size, page count, CC-BY-4.0 license,
attribution and adaptation notice. See [the conformance contract](srd-5.2.1-conformance.md).

The existing guarded character contract represents all requirements:

- Both feats are General feats with a minimum character level of 4.
- Ability Score Improvement is repeatable. Choose one of two increase patterns:
  one ability score by 2, or two distinct ability scores by 1. Each pattern owns
  a nested ability choice from all six abilities. A prerequisite predicate refers
  to the pattern choice and names the selected alternative. These are conditional
  alternatives; they are not simultaneous grants. Option quantities retain the
  increase per selected score; a root parameter retains the maximum score of 20.
- Grappler requires Strength **or** Dexterity of at least 13, in addition to the
  level prerequisite. Its choice increases either Strength or Dexterity by 1,
  capped at 20. It uses page 87's default single acquisition rule.
- Punch and Grab is optional when an Unarmed Strike hits a creature as part of
  the Attack action on the character's turn. It permits both the Damage and
  Grapple options, once per turn.
- Attack Advantage applies to attack rolls against a creature Grappled by the
  character. Fast Wrestler removes extra movement for moving a creature Grappled
  by the character only when that creature is the character's size or smaller.

All conditions and quantities are opaque mechanic keys, choice quantities or
prerequisite parameters. The catalogue does not evaluate them. No schema change
or runtime consumer is added. Unknown mechanics remain unavailable; validation
cannot grant bonuses, feats, actions, creation choices, coverage or Full Rules.

Composition tests append each canonical advancement, background, class,
subclass, species and feat catalogue once. They omit level-one reference subsets
and composition manifests, which would duplicate definitions. Tests compare the
complete graph to an independent source transcription and reject structurally
valid contrary values for category, repeatability, prerequisites, alternatives,
ability domains, selection counts, branch guards, increments, caps and benefits.
Schema tests cover source pin, license attribution/adaptation guards, ownership,
references, unknown fields and unknown-key readiness. Fixed Human Soldier Fighter
and Human Acolyte Cleric builds, stats, sheets and creation limits have regression
checks, along with unchanged coverage and unavailable Full Rules.

Local verification:

```sh
GOMAXPROCS=2 go test -p 2 -short -race ./...
GOMAXPROCS=2 go vet -p 2 ./...
python scripts/verify_srd_general_feats.py /tmp/edda-wave1/SRD_CC_v5.2.1.pdf
gofmt -l internal/rules/srd521/catalogue_general_feats_test.go
git diff --check
```

The offline comparator verifies the PDF hash, full General Feats source text,
exact two issue-300 inventory identities/pages, nested increase patterns,
prerequisites, ability domains, limits and conditional benefit keys. It requires
`pdftotext` and the pinned local PDF; Go tests need no PDF or network.

No database, UI, inference or production changes are involved. Hosted
qualification and independent review belong to the parent. The active Go 1.27.3
`golangci-lint` shim has no installed binary; host tooling is unchanged.
