# Chapter fidelity qualification (#87)

**Human review received; the reported foreshadowing concern is corrected.**
See [the recorded feedback and bounded correction](human-review.md).

This packet contains original synthetic transcripts and authored chapter examples.
They were created for this test, not copied from a production campaign or a human
play session. The chapter model double returns the example files verbatim. No
live model generated these examples. The owner reviewed the original examples;
the agent applied the one recorded editorial correction. A green
test run demonstrates the source/persistence contract and the fixed example
checks; it does not qualify a provider's prose.

## Evidence to inspect

The canonical original transcripts are in [campaigns.json](campaigns.json), under
each campaign's `source.source_turns`. Each turn has an original player input,
visible narration, stable log ID and completed action ID. `outcomes` cites a
specific source turn and the corresponding phrase in the chapter example.

| Campaign | Frozen range | Chapter example | Questions for the reviewer |
| --- | --- | --- | --- |
| `failed-plan` | 1–3, legacy/narrative | [Greywash](failed-plan.chapter.md) | Did the failed crossing stay failed, the rope stay lost and the two silver stay with Mara? Is waiting distinct from the offered upstream bridge? Does the rising water remain unresolved? |
| `corrected-misunderstanding` | 1–4, SRD 5.2.1/strict | [Bell Quay](corrected-misunderstanding.chapter.md) | Is theft only Tamsin's initial suspicion? Does finding the tally correct it? Are the apology, refused guidance and departure alone all retained? |
| `undiscovered-secret` | 1–3, SRD 5.2.1/light | [Ash Reach](undiscovered-secret.chapter.md) | Does Sera find no entrance? Is Noll's lack of knowledge retained? Do the marks and route remain unresolved without hinting at the hidden vault? |

The JSON's `hidden_lore` is reviewer/test-only, seeded as a world fact with
`player_known=false`. `tool_secret` goes into raw session-log tool data. `pending`
contains narration that is not committed; `later` is a completed turn beyond that
pending turn. None belongs in the frozen range or in the model request. These
are deliberately readable synthetic secrets, not credentials.

## Reproduce the mechanical evidence

From the repository root, using no application DSN or provider credentials:

```sh
rtk proxy go test -short -race ./...
rtk proxy go vet ./...
rtk proxy go test -tags integration -run Integration -count=1 -timeout 10m ./internal/journal ./internal/handlers ./internal/state/sqlc
rtk proxy pnpm --dir frontend install --frozen-lockfile
rtk proxy pnpm --dir frontend test
rtk proxy env EDDA_FIXTURE_PORT=4197 pnpm --dir frontend test:browser --workers=2
rtk proxy pnpm --dir frontend build
rtk proxy pnpm --dir frontend lint
```

The integration suites own and remove disposable pgvector PostgreSQL containers.
The fidelity test uses the real journal store and summarizer, captures the model
request, and compares source IDs, turn order/range, names, inputs, outcomes, rules
provenance and hash with the original fixture. It excludes private facts, raw tool
results, pending narration and later turns. Regeneration after a transcript edit
and hidden-fact change must still use the original frozen source. It checks that
the saved chapter equals the recorded model response.

The short tests validate the fixtures' cited names/outcomes and reject deliberately
broken examples: missing names/outcomes, contradicted results, added hidden lore,
tool markers and pending/later narration. Literal phrases are useful regression
anchors for these examples, not a general semantic validator. A model could
invent an event or paraphrase a secret without using any forbidden phrase; that
is why source comparison and the human review remain necessary.

## Human checkpoint

Read the original `source.source_turns` and all three chapter files, then consult
[human-review.md](human-review.md) for the recorded owner feedback and agent correction. For each example, record a plot-fidelity verdict
and a separate readability verdict, with concrete source turns and chapter
sentences. Record any omission, invented event, reversed outcome, premature
resolution or spoiler. The recorded human feedback is separate from the test results.

If a human needs provider-generated prose assessed, capture it as a new, separately
identified candidate with its model, prompt version, exact input and unedited
response; this packet contains deterministic examples only. No paid inference or
production campaign is required for this checkpoint.

A future failure gets a bounded corrective item in the review report: identify
the fixture, source turn, offending sentence, owning boundary, smallest fix and
specific rerun/human re-review. Source contamination belongs to journal capture;
misleading paraphrase belongs to chapter instructions or generation; reader
presentation belongs to the chapter reader. Do not absorb export, action,
suggestion or rules-engine work. Split any corrective work that exceeds a focused
PR. No corrective product failure is claimed before one has been observed.
