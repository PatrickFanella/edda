# Hunter source catalogue (#315)

The offline Hunter fixture adapts the six identities from SRD 5.2.1 PDF p61
into 21 unavailable metadata records: Hunter; Hunter’s Lore (3); Hunter’s Prey
(3); Defensive Tactics (7); Superior Hunter’s Prey (11); Superior Hunter’s
Defense (15). Source SHA-256 is
`8974902d109d6e63672d7c490bde9ccf052410503d9cfa768237154fbc5e3d87`.
The fixture retains CC-BY-4.0 attribution and an adaptation notice. The PDF is
not committed.

`character-hunter-composition.json` composes canonical Ranger externally once.
Hunter contains no class, resource, spell, feat or adjacent subclass definitions.
The guarded character contract is unchanged. Hunter’s Mark references Ranger’s
actual Favored Enemy feature. Superior Hunter’s Prey refers to the current
canonical spell extra damage and Foe Slayer improvement, without introducing a
fixed spell die, extra resource cost, recharge or invented scaling.

The two choose-one domains retain both source alternatives and replacement
with the other option whenever a Short or Long Rest finishes. Each conditional
option is a prerequisite record owned by its choice. Colossus Slayer retains
weapon hit, missing HP, 1d8 extra weapon damage to that target and once-per-turn
limits. Horde Breaker retains the same weapon, another creature, within 5 feet
of the original target and within weapon range, not attacked this turn, once
on each of your turns when making a weapon attack. Defensive options retain
Opportunity Attack disadvantage and the distinct creature-hit trigger and
rest-of-turn restriction for Multiattack Defense.

Lore records only source metadata about the player’s own marked creature’s
Immunities, Resistances and Vulnerabilities; no runtime disclosure exists.
Superior Prey retains once-per-turn damage to the marked creature, optional
spell extra damage to a different visible creature within 30 feet of the first.
Superior Defense retains optional Reaction on taking damage, Resistance to
that damage and further damage of the same type until the current turn ends.

## Verification

```sh
rtk proxy env EDDA_SRD_PDF=/tmp/edda-wave1/SRD_CC_v5.2.1.pdf GOMAXPROCS=2 go test -p 2 ./internal/rules/srd521 -run Hunter -count=1
rtk proxy python3 scripts/verify_srd_hunter.py /tmp/edda-wave1/SRD_CC_v5.2.1.pdf
rtk proxy env GOMAXPROCS=2 go test -p 2 -short -race ./...
rtk proxy env GOMAXPROCS=2 go vet -p 2 ./...
```

The readable Go oracle compares every graph field. The separate standalone
Python proof reads the pinned PDF directly, checks the source clauses, extracts
damage dice and both distances, constructs its own expected graph and compares
actual metadata. It also checks the six inventory identities/pages and the
external composition manifest and runs canonical Ranger’s PDF comparator.
An optional candidate JSON argument allows independent contradiction probes.
Neither proof relies on record digests or merely matching source anchors.

Structurally valid negative graphs must pass schema validation before both
semantic proofs reject them: existing wrong operands, missing or changed
options, choice count and rest replacement, omitted conditions, targets,
numbers, range, frequency, Reaction cost, duration and invented level scaling.
Strict decoding, source/schema pinning, unknown mechanic readiness, missing
Ranger and duplicate canonical/batch composition remain guarded. Fixed Fighter
and Cleric creation/projection and Cleric casting assertions remain checked;
Ranger and Hunter creation/projection remain unavailable. PDF tests are opt-in
via `EDDA_SRD_PDF`; graph proofs run without downloading the PDF.

There is no runtime consumer, grant, creation, advancement, private-lore
access, coverage change or Full Rules promotion. Parent independent review and
hosted final gates remain separate from local validation.
