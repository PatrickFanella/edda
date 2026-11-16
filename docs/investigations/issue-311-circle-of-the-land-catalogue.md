# Circle of the Land source catalogue

Issue #311 imports the six Circle of the Land inventory identities on PDF p46
of SRD 5.2.1. The pinned local PDF SHA-256 is
`8974902d109d6e63672d7c490bde9ccf052410503d9cfa768237154fbc5e3d87`.
Attribution, license and adaptation notice travel with the JSON source envelope;
see [the source policy](../srd-5.2.1-conformance.md).

The batch has 27 records: the subclass, five named features, four supporting
conditional land-table features, one land choice, two separate Natural Recovery
resources and 14 prerequisite predicates. Supporting records reuse the six
inventory headings rather than inventing additional source identities.

Composition loads `character-druid.json` first, then
`character-circle-of-the-land.json`. The canonical Druid's class, 20 progression
rows, Wild Shape and nine spell-slot resources are referenced without copying
or changing their ownership. This uses the existing character contract.

The land choice selects one of arid, polar, temperate and tropical at the end of
a Long Rest. Each land has a separate conditional table with rows at Druid
levels 3, 5, 7 and 9; the selected land's rows at or below the current level apply.
This retains all 24 spell references, including the cantrips, without making
four lands cumulative or treating table spells as optional selections.

Land's Aid records the Magic action, one Wild Shape use, 60-foot point range,
10-foot-radius Sphere, chosen creature Constitution saves against the caster's
spell save DC, Necrotic damage with half damage on success, and one chosen
creature's healing. Damage and healing use d6s: two at level 3, three at level 10
and four at level 14.

Natural Recovery keeps its two Long Rest limits separate: one prepared level
1+ Circle spell cast without a slot, and one Short Rest recovery of expended
slots. The latter's combined slot-level budget is half the Druid level rounded
up; no recovered slot can exceed level 5. The canonical nine slot records are
operands, while the level cap restricts recovery eligibility.

Nature's Ward retains Poisoned condition immunity and the current land's
resistance: arid Fire, polar Cold, temperate Lightning and tropical Poison.
Nature's Sanctuary retains its Magic action, one Wild Shape use, ground-only
15-foot Cube within 120 feet, one-minute duration ending on Incapacitated or
death, Half Cover for the Druid and allies in the area, and the current Ward
resistance for allies there. A Bonus Action can move the Cube up to 60 feet to
ground within 120 feet. Every feature and the subclass has its Druid-level
prerequisite; relevant conditions and canonical resource operands remain
unevaluated metadata.

## Semantic proof

`catalogue_circle_of_the_land_source_test.go` independently transcribes the
complete graph in readable Go expectations. It compares all CharacterRecord
fields, including absence of unrelated grants and scaling.

`scripts/verify_srd_circle_of_the_land.py` is a standalone Python/pdftotext
comparator. It checks the actual pinned PDF, independently transcribes the full
p46 semantic graph, compares every serialized field and verifies the exact six
inventory IDs/pages. It also runs the existing canonical Druid PDF comparator.
It does not use Go assertions, catalogue hashes, a network service or a PDF
text file checked into the repository.

The same 185 contrary catalogues run through both semantic proofs. The Go test
first requires every mutation to pass structural validation in composition.
Controls alter every mechanic, numeric parameter and predicate operand; choice
counts/options; recovery capacity; prepared-spell row levels and row membership;
and source heading ownership. A schema-valid source contradiction must fail
semantic equality. Unknown schema fields, duplicate composition and missing
canonical Druid are rejected separately.

Run locally with the pinned PDF available:

```sh
GOMAXPROCS=2 EDDA_SRD_PDF=/tmp/edda-wave1/SRD_CC_v5.2.1.pdf go test -p 2 ./internal/rules/srd521 -run CircleOfTheLand -count=1
python3 scripts/verify_srd_circle_of_the_land.py /tmp/edda-wave1/SRD_CC_v5.2.1.pdf
GOMAXPROCS=2 go test -p 2 -short -race ./...
GOMAXPROCS=2 go vet -p 2 ./...
```

Readiness remains unavailable for catalogue mechanics and unknown keys. Tests
check fixed-build creation rejection, fixed Cure Wounds behavior and unchanged
coverage/full-support eligibility. This batch supplies no runtime grants,
playable Druid support, coverage expansion, UI, database changes or deployment.
The PDF test skips when `EDDA_SRD_PDF` is absent; the standalone comparison was
run locally against the pin. The host's golangci-lint shim has no binary for
active Go 1.27.3, so lint execution is unavailable without a host tooling change.
Independent review and hosted CI qualification belong to the parent workflow.
