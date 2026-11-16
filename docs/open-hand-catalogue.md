# Warrior of the Open Hand source metadata (#313)

The batch imports exactly Warrior of the Open Hand, Open Hand Technique (Monk
level 3), Wholeness of Body (6), Fleet Step (11), and Quivering Palm (17) from
pinned SRD 5.2.1 PDF p. 52. Seventeen records retain the five identities, class
level prerequisites, three optional Open Hand effects, and canonical operands.

The composition manifest loads canonical Monk once and this subclass once.
Flurry of Blows and Step of the Wind point to the existing Monk choice records;
Focus expenditure points to its Focus Points resource, and healing points to
Martial Arts for its die. The canonical Monk comparator checks its die progression,
free/paid Focus alternatives, costs, and recovery. The subclass does not copy
those definitions. Wholeness retains Wisdom-based healing and use limits,
minimums, Bonus Action timing, and complete Long Rest recovery. Fleet Step keeps
the exclusion of Step of the Wind as the triggering Bonus Action and the immediate
follow-up. Quivering Palm retains its optional hit trigger, four-point cost,
Monk-level duration in days, both damage termination actions, same-plane
restriction, Constitution save and half-damage alternative, 10d12 Force damage,
one-creature limit, and harmless termination without an action.

The existing guarded contract is unchanged. All keys remain unavailable; the
batch grants no runtime effects, creation or advancement support, and makes no
coverage or Full Rules claim. No adjacent definition is imported. The JSON carries
the pinned artifact, CC-BY-4.0 attribution, and adaptation notice; see
[srd-5.2.1-conformance.md](srd-5.2.1-conformance.md).

`catalogue_open_hand_source_test.go` contains a readable Go graph oracle.
`verify_srd_open_hand.py` separately transcribes the source facts, reads the pinned
PDF directly, checks all relevant clauses, and compares every actual record field.
It also runs the canonical Monk source comparator. These are semantic comparisons,
not digest or anchor-only checks. The Go controls first require structural validity
and unavailable readiness, then pass the same contrary graphs to both proofs.
They cover wrong existing operands, altered mechanics, missing alternatives and
conditions, costs, healing, recovery, class levels, range, damage, duration and
invented scaling. Composition duplicates, source/schema errors and unknown fields
and keys are rejected; fixed Fighter/Cleric runtime regressions also run.

```sh
rtk proxy python3 scripts/verify_srd_open_hand.py /tmp/edda-wave1/SRD_CC_v5.2.1.pdf
rtk proxy env GOMAXPROCS=2 EDDA_SRD_PDF=/tmp/edda-wave1/SRD_CC_v5.2.1.pdf go test -p 2 -short -race ./...
rtk proxy env GOMAXPROCS=2 go vet -p 2 ./...
```

Source transcription still needs independent human review. Hosted CI is separate
from local tests. The known Go 1.27.3 golangci-lint shim has no installed binary;
this task does not change host tooling.
