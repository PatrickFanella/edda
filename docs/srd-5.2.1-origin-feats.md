# Origin feat source metadata (#301)

`internal/rules/srd521/testdata/character-origin-feats.json` records exactly
Alert, Magic Initiate, Savage Attacker and Skilled from PDF page 87 of the pinned
English SRD 5.2.1. Its four feat identities and seventeen nested records are offline
character metadata. No runtime consumer loads this file.

The source pin is
`8974902d109d6e63672d7c490bde9ccf052410503d9cfa768237154fbc5e3d87`.
Each record links to its containing feat in the planning inventory at page 87.
The source object retains the artifact URL, size, page count, CC-BY-4.0 license,
required attribution and adaptation notice. See [the conformance contract](srd-5.2.1-conformance.md)
for the source and licensing policy.

The metadata retains:

- Alert’s optional Proficiency Bonus on Initiative rolls and optional immediate
  Initiative exchange with one willing ally in the same combat. Separate
  predicates bar the exchange when either participant is Incapacitated.
- Magic Initiate’s single Cleric, Druid or Wizard list; Intelligence, Wisdom or
  Charisma choice at feat selection for its spells; two cantrips and one level 1
  spell from that list; always-prepared status; one slot-free cast recovered on
  finishing a Long Rest; casting with any owned spell slots; optional replacement
  of one feat spell on gaining a level, with the same level and selected list;
  and a different selected list on each repeat.
- Savage Attacker’s optional weapon-hit trigger, rolling the weapon’s damage
  dice twice and using either roll against that target, once per turn.
- Skilled’s proficiency in any combination of three skills or tools and
  unrestricted repeatability. Alert and Savage Attacker retain the page’s default
  single acquisition rule.

Mechanic keys identify opaque domains and requirements. For example, a selection
count of three with skill and tool domain keys means three proficiencies from
those domains, not three domain definitions. Predicate operands refer to the
feat instance and its choices; they are never evaluated by this catalogue.
Magic Initiate’s replacement selection is optional and level-triggered, not an
extra initial spell. Its repeatability operand denotes previous selections of
the same feat identity. No spell, skill, tool, item, class or adjacent feat is
imported here.

The existing schema suffices except that its guarded numeric `parameters` field
now also permits feat records. Savage Attacker uses it for two rolls and one use
per turn. The existing positive-value and required-key guards still apply.
Category and repeatability use required mechanic keys. Choices, resource capacity,
requirements and ownership retain their existing schema and readiness checks.
All character mechanic keys remain unavailable, including unknown keys.

Canonical composition combines this file once with the complete advancement,
background, Cleric, Fighter and Champion catalogues. Level-one reference subsets
are not appended as duplicate definitions. Existing runtime content IDs for
Savage Attacker and Magic Initiate remain separate from the character metadata
namespace. Fixed Human Soldier Fighter and Human Acolyte Cleric builds, their
saved stats, projected sheets and feat availability are covered by regression
fixtures. This import grants no automatic effects, actions, starting inventory,
derived statistics, new creation choices, coverage or Full Rules status.

Verification commands:

```sh
go test -short -race ./...
go vet ./...
python scripts/verify_srd_origin_feats.py /tmp/edda-wave1/SRD_CC_v5.2.1.pdf
git diff --check
```

The exact-source Go fixture independently transcribes the entire expected graph,
including every predicate key and operand, alternative, selection count, resource
capacity, numeric value and repeatability restriction. Negative tests reject bad
sources, ownership, references, fields, predicates, counts and numeric metadata;
unknown well-formed mechanics pass structural validation but fail readiness.
The offline PDF comparator checks the artifact hash, exact four inventory
identities/pages, source wording anchors and selection/numeric values.

No database, browser, inference or production work is required for this metadata
batch. Hosted checks and independent review belong to the parent. The host’s
active Go 1.27.3 has no configured `golangci-lint` binary; invoking it fails before
analysis. Host tooling has not been changed.
