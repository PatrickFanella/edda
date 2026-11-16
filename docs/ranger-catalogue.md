# Ranger source catalogue (#293)

The offline fixture `internal/rules/srd521/testdata/character-ranger.json`
adapts exactly the 18 source identities in #293 into 72 graph records. The
source is SRD 5.2.1 PDF pp57–59, SHA-256
`8974902d109d6e63672d7c490bde9ccf052410503d9cfa768237154fbc5e3d87`.
Attribution, CC-BY-4.0 license and adaptation notice are in the fixture.

All 20 class rows retain proficiency, feature references, ASI repeats at
4/8/12/16, subclass slots at 7/11/15, prepared-spell counts, Favored Enemy
uses and all five spell-slot columns. Unavailable slots have no positive
capacity entry. Favored Enemy explicitly records the 5.2.1 always-prepared
Hunter’s Mark and slot-free casting variant. Relentless Hunter, Precise Hunter
and Foe Slayer retain their distinct concentration, target and damage-die
conditions and references to that feature.

Nested records retain the eight-skill domain, conditional one-skill multiclass
selection, equipment alternatives and quantities, two-language choice, Deft
Explorer and Expertise eligibility, mastery proficiency and replacement,
Fighting Style versus Druidic Warrior with its conditional cantrip choice,
prepared-spell eligibility, qualified feats and opaque subclass domains.
Tireless records Magic action temporary HP, 1d8 plus Wisdom with minimum one,
its separate use limit and Long Rest recovery, and Short Rest Exhaustion
reduction. Nature’s Veil records Bonus Action, Invisible until end of next
turn, Wisdom minimum-one uses and Long Rest recovery. Roving, Extra Attack,
Feral Senses and all other named features retain numbers and conditions.

The sole contract addition accepts the symbolic capacity key
`ranger/wisdom-modifier-minimum-one`. It remains unavailable; negative fixtures
reject unknown formulas, fixed capacity or level counts combined with a formula,
missing required keys and capacity on other record kinds. No formula is evaluated.

## Verification

```sh
rtk proxy env GOMAXPROCS=2 go test -p 2 ./internal/rules/srd521 -run Ranger -count=1
rtk proxy python3 scripts/verify_srd_ranger.py /tmp/edda-wave1/SRD_CC_v5.2.1.pdf
rtk proxy env GOMAXPROCS=2 go test -p 2 -short -race ./...
rtk proxy env GOMAXPROCS=2 go vet -p 2 ./...
```

The readable Go oracle freezes every serialized graph field and reports field
paths and expected values. The standalone Python comparator reads the pinned
PDF directly, extracts every progression cell and numeric prose values, and
maps source clauses to the entire expected graph independently of the Go
oracle. It accepts an optional candidate JSON path. Neither proof uses record
digests or stops at finding source anchors.

Both reject eight contrary-source controls: an existing wrong operand,
condition, recovery, cost, range, table count, Favored Enemy variant and invented
scaling. Go first confirms these controls are structurally valid. Additional
checks cover exact identity/page ownership, strict decoding, readiness unknown
keys, composition with each existing canonical catalogue once, fixed Fighter
and Cleric creation/projection assertions, and unavailable Ranger creation.

This imports no Hunter, spell, feat or adjacent class definitions. It adds no
runtime consumer, grants, derived statistics, creation, advancement, coverage
promotion or Full Rules enablement. Parent review and hosted final gates remain
separate from these local checks. The source PDF is not committed.
