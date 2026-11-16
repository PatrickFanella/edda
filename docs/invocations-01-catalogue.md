# Invocation metadata, batch 01 (#302)

`character-invocations-01.json` records exactly twelve inventory identities:
Agonizing Blast, Armor of Shadows, Ascendant Step, Devil’s Sight, Devouring
Blade, Eldritch Mind and Eldritch Smite on PDF page 72; Eldritch Spear,
Fiendish Vigor, Gaze of Two Minds, Gift of the Depths and Gift of the Protectors
on page 73. The 41 records contain these identities and their nested choices
and unevaluated prerequisites. The source retains CC-BY-4.0 attribution,
the pinned 5.2.1 digest and an adaptation notice.

The composition manifest imports canonical Warlock once, then this batch.
No Warlock records are copied. The narrow schema-version-1 extension permits
an `invocations`-family feature only under the canonical Warlock invocation
choice, with an exact inventory-slug identity and Warlock class ancestry.
Nested ownership, matching source, references, cycles and duplicate guards
continue to apply. Other feature families retain their existing owners.

Every listed minimum Warlock level is a prerequisite using canonical Warlock.
Agonizing Blast and Eldritch Spear retain known damaging Warlock cantrip
choices, distinct eligible selections on repetition, and canonical cantrip
and Pact Magic operands. Spear additionally requires range of at least ten
feet and increases range by thirty feet per Warlock level. Agonizing Blast
adds Charisma modifier to the chosen spell’s damage rolls. Only these two
entries specify repeatability.

The self-only spell casts retain their spell references and no-slot costs.
Fiendish Vigor uses the highest temporary-hit-point die result without rolling;
it does not import an older edition’s roll. Devil’s Sight includes magical
and nonmagical Dim Light and Darkness within 120 feet. Devouring Blade
replaces Thirsting Blade’s one extra attack with two. Eldritch Mind applies
to Constitution saves made to maintain Concentration.

Eldritch Smite retains the once-per-turn pact-weapon hit trigger, one Pact
Magic slot cost, extra Force damage of 1d8 plus 1d8 per slot level, and optional
Prone only for Huge or smaller targets. The slot operands reference canonical
Pact Magic and its resource. Thirsting Blade, Pact of the Blade and Pact of the
Tome remain exact opaque invocation requirements owned by #303/#304; no
placeholder definition is imported. Their predicate operands identify the
existing invocation selection owner, while their keys retain the missing
specific requirement.

Gaze of Two Minds records a Bonus Action touch of a willing creature, duration
through the end of the next turn, subsequent Bonus Action maintenance on the
same plane, termination without maintenance, special senses, and a choice of
spell origin within sixty feet. Gift of the Depths records underwater
breathing, Swim Speed equal to Speed and one no-slot Water Breathing cast
recovered on finishing a Long Rest. Gift of the Protectors records the Book
of Shadows page, permission and action to write a creature’s own name, symbolic
Charisma modifier capacity with minimum one, the zero-HP but not outright death
condition, replacement with one HP, one shared trigger until the Warlock
finishes a Long Rest, and Magic action touch to erase a chosen name. Symbolic
capacity and recovery are mechanic metadata; no fixed capacity is invented.

The independent Go graph oracle asserts every record and field. The standalone
Python comparator reads the pinned PDF columns directly, checks the source
clauses and exact twelve inventory/page identities, and compares its separate
semantic transcription to the complete actual metadata. It does not call Go
or read the Go oracle. Contrary controls run through structural validation,
then must fail both semantic proofs and readiness: omitted choices and
prerequisites, missing conditions, changed numeric values, known wrong record
operands, changed costs/recovery, previous-edition variants and invented scaling.
Strict decoding, pin/schema checks, narrow ownership, duplicate composition,
unknown-key refusal, fixed Human Fighter/Cleric paths and unavailable Warlock
creation remain covered.

Run from the repository root:

```sh
rtk proxy python3 scripts/verify_srd_invocations_01.py /tmp/edda-wave1/SRD_CC_v5.2.1.pdf
rtk proxy sh -c 'GOMAXPROCS=2 go test -p 2 -short -race ./...'
rtk proxy go vet -p 2 ./...
```

These proofs establish this batch’s source metadata. They do not grant
invocations, cast spells, evaluate prerequisites, register Warlock creation,
advance characters, change derived statistics or promote coverage or Full Rules.
