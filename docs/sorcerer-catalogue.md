# Sorcerer source catalogue (#295)

The offline fixture `internal/rules/srd521/testdata/character-sorcerer.json`
records exactly the 11 source identities owned by #295 as 63 graph records.
Its source is SRD 5.2.1 PDF pp64–66, SHA-256
`8974902d109d6e63672d7c490bde9ccf052410503d9cfa768237154fbc5e3d87`.
The fixture retains CC-BY-4.0 attribution and an adaptation notice.

All 20 progression rows retain proficiency and feature references, including
ASI repeats at 4/8/12/16, Metamagic improvements at 10/17, and subclass slots
at 6/14/18. Nested choices and resources retain all cantrip, prepared-spell,
Sorcery Point, Metamagic selection and nine spell-slot column counts.
A dash means no positive capacity at that level; no zero capacity is invented.
The class source heading is p64; its progression table continues on p65.

Conditional nested first-class and multiclass entry records distinguish all
core traits from the multiclass Hit Point Die and level-one features.
The two-skill domain and both starting-equipment alternatives belong to
first-class entry. The multiclass slot rule remains an opaque requirement.
Spellcasting retains the Sorcerer lists, initial level-one spell restriction,
slot eligibility, table-driven additions, level-gain replacement counts,
recommendations, always-prepared exclusion, Charisma and Arcane Focus.

Font of Magic retains the capped point pool and Long Rest recovery, no-action
slot-to-point conversion equal to the expended slot level, and Bonus Action
point-to-slot conversion. All five conversion costs and minimum Sorcerer levels
are explicit parameters. Created slots cannot exceed level five and vanish
on a Long Rest. Innate Sorcery retains two uses, Long Rest recovery, Bonus
Action, one minute, +1 Sorcerer spell save DC and Advantage on Sorcerer spell
attacks. Restoration retains optional recovery of expended points on a Short
Rest up to half Sorcerer level rounded down, once until a Long Rest.
Incarnate retains its two-point activation alternative only with no Innate
Sorcery uses left and the two-option benefit while Innate Sorcery is active.
Apotheosis retains one waived Metamagic option cost on each own turn while
Innate Sorcery is active.

Metamagic retains its opaque option domain, chosen-option cost, one option per
cast spell with option-defined exceptions, and replacement with an unknown
option on level gain. Named option definitions belong to #305. Draconic
Sorcery remains an opaque subclass reference. Neither adjacent batch is copied.
The existing guarded schema expresses this entire batch without an extension.
No runtime consumer, grants, derived statistics, creation, advancement,
coverage promotion or Full Rules support is added. Readiness stays unavailable.

## Verification

```sh
rtk proxy env GOMAXPROCS=2 go test -p 2 ./internal/rules/srd521 -run Sorcerer -count=1
rtk proxy python3 scripts/verify_srd_sorcerer.py /tmp/edda-wave1/SRD_CC_v5.2.1.pdf
rtk proxy env GOMAXPROCS=2 go test -p 2 -short -race ./...
rtk proxy env GOMAXPROCS=2 go vet -p 2 ./...
```

The readable Go oracle freezes every graph field, operand and table cell and
reports differences by record and field. The standalone Python comparator
reads the pinned PDF, extracts numeric progression and conversion cells, and
binds prose clauses to the entire expected graph independently of the Go oracle.
It accepts an optional candidate JSON path and needs Poppler `pdftotext`.
Neither check relies on record digests or source anchors alone.

Both reject eleven semantic contradictions: an existing wrong operand, missing
equipment alternative, wrong active condition, wrong recovery, activation cost,
duration, conversion cost, conversion minimum, spell-slot cell, invented scaling,
and a missing active condition. Go first checks that each mutation remains
structurally valid. Additional checks cover exact inventory identities/pages,
strict decoding, schema/source guards, duplicate records, unknown readiness
keys, canonical composition once per batch, and fixed Fighter/Cleric creation
and projection boundaries. Sorcerer creation remains unavailable.

The Go 1.27.3 linter shim has no installed binary, so that local check cannot
run. Parent review and hosted final gates remain separate from local validation.
The PDF is not committed. No UI, deployment or production data changes belong
to this batch.
