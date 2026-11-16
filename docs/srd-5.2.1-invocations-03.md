# Invocation catalogue batch 03 (#304)

The four page-74 identities are Thirsting Blade, Visions of Distant Realms,
Whispers of the Grave and Witch Sight. The fixture retains the pinned English
SRD 5.2.1 source, CC-BY-4.0 attribution and adaptation notice. Feature IDs use
`feature/warlock/invocations/<source-slug>` and belong to the canonical
`choice/warlock/invocations`, under Eldritch Invocations and Warlock.

The composition manifest imports canonical Warlock, Origin feats and Invocation
batch 02 once each. This batch copies no class or other invocation records. Thirsting Blade requires Warlock level 5
and Pact of the Blade; its Extra Attack applies only to the pact weapon, twice
instead of once when taking the Attack action on your turn. Its prerequisite
references `srd-5.2.1:character/feature/warlock/invocations/pact-of-the-blade`, the stable
owner delivered by #303. The predicate remains unavailable metadata.
Visions requires level 9 and permits casting Arcane Eye without expending a
spell slot. Whispers requires level 7 and permits casting Speak with Dead
without expending a spell slot. Witch Sight requires level 15 and grants
Truesight at 30 feet. Spell definitions retain their separate owner; these
invocations add no casting-time override or rest-based pool. All four have
canonical selection prerequisites and default nonrepeatability. No further
nested option, resource or level scaling is specified for these four headings.

The Go oracle independently constructs every record, including all empty
fields, exact operands, conditions and parameters. The standalone Python
comparator checks every actual field against an independent transcription,
then ties the transcription to the pinned PDF hash, page-74 clauses, extracted
levels, attack quantity and sight range. It checks the inventory's four exact
identities and source envelope. Neither proof infers correctness from failed
readiness. Structurally valid contrary graphs must fail both semantic proofs
while remaining unavailable. Controls cover every key and numeric field,
existing wrong operands, omitted predicates and conditions, source-name
variants and invented recovery, cost, range or scaling. Strict decoding,
unique ownership, secondary-source bypass, duplicate composition, unknown-key,
fixed Fighter/Cleric and unavailable Warlock guards also run.

```sh
rtk proxy python3 scripts/verify_srd_invocations_03.py /tmp/edda-wave1/SRD_CC_v5.2.1.pdf
rtk proxy env GOMAXPROCS=2 go test -p 2 -short -race ./internal/rules/srd521 -run Invocations03 -count=1
rtk proxy env GOMAXPROCS=2 go test -p 2 -short -race ./...
rtk proxy env GOMAXPROCS=2 go vet -p 2 ./...
```

This is offline metadata. Catalogue validation enables no creation,
advancement, actions, grants, derived statistics, coverage or Full Rules support.
