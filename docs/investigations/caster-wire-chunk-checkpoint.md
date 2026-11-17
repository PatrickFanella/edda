# Caster wiring request bound: halted checkpoint

The initial wiring run acknowledged `330:hp` and `330:predicates`, then returned an ambiguous HTTP 502 during the eleven-field `330:creation` request. Bounding serial GraphQL work is a timeout mitigation, not a proven diagnosis of that failure.

`Coordinator.wire` now sends at most four mutation fields per serial HTTP request. It retains body, parent and sorted prerequisite ordering, variable-based inputs, the existing 7.5-second interval and 480-request rolling-hour budget, nested result identity validation, durable request intents and immediate ambiguity halt. Operation labels include the alias and exact field names. The existing metadata read determines which exact body, parent and prerequisite operations remain; committed fields are skipped on restart. No retry or fallback was added.

## Fresh read-only reconciliation

A single complete read under the coordinator lock saved all 1,888 graph rows and complete metadata for all 1,888 issues before diagnostic prefix assertions. Observed at `2026-11-17T01:59:14.912189+00:00`:

- All 871 registered identities match the immutable IDS milestone and stable map. All are OPEN Task/M.
- All 1,017 old contracts and protected Project type/Size values match the baseline. All old typed identities also match the retained preflight snapshot. Only decisions 330/331/332 have the previously allowed state tolerance; all three are currently OPEN.
- New bodies are allocation or independently reproduced exact final bodies; parents are absent or approved; blockers are subsets of approved blockers. Graph and metadata agree. Both DAGs and all three caps pass.
- `330:hp` (#1173) has final body, parent 330 and blocker 968. `330:predicates` (#1174) has final body, parent 330 and no blockers. `330:creation` (#1175) has final body, parent 330 and blockers 430, 1173, 1176, 1177. Its remaining blockers are 1179, 1180, 1181, 1187, 1188.
- No other new node differs from allocation. The committed fields match the actual journal prefix: two acknowledged requests followed by the partially committed ambiguous request.
- Registry, journal, active halt, immutable IDS bytes, milestone sidecar and original 502 packet hashes were unchanged across reconciliation. No native mutations occurred.

Evidence is retained only in the dedicated authority diagnostics directory:
`/home/onnwee/.local/state/edda-overnight-219965e8/caster-publication/diagnostics/chunk-followup/`.
It contains `graph.json`, `metadata.json`, `snapshot.json`, `report.json`, `old-identity-report.json`, `validated-partial.json`, the diagnostic script and `hashes.json`.

| Artifact | SHA-256 |
| --- | --- |
| Full saved snapshot | `24d3408b4a2d3b640d1683e2a54c84de01799f1b2b1476db39968b2463fe2a5e` |
| Partial reconciliation report | `d8fbe7114457b325f62c796d4b156b0b4b21cbff0ae36567fff1ec7d4b685503` |
| Registry | `210bb5bac74c16cb41d8ff24fabd5e86744476ac50aa038504ff0a62871e26aa` |
| Journal | `e149f2b99f0ba253da9715148ffb98e3fc8f504032849f009c1189423e6b5aef` |
| Active halt | `b5c5cbe912a3c2afae2d5bb66bf5026d29bd5c83e1fb2577aeb051f552a5cc50` |
| Coordinator code | `947f99d5d5ffadd853f60f400e37e408a1a9e701e72018cacc6944ee28d51a18` |
| Unchanged pacing/transport guard | `d814238dca69aa480430405fec367f2dbcd5314832d0438999fda510c9eb2a31` |
| New offline chunk tests | `dea0bf12f6c66ad77b4ba7aaabcd4bf6b1608e5a4b20363649696407d8f0eac9` |

## Focused offline validation

- `python -m unittest discover -s scripts -p 'test_caster_*.py'`: 32 passed.
- `python -m unittest discover -s scripts -p 'test_sorcerer_native_binding.py'`: 7 passed.
- `git diff --check`: passed.

Tests use actual frozen source nodes with synthetic native identities and mocked HTTP. The 80-prerequisite node produces 21 requests for 82 ordered fields. Controls cover complete operation coverage, variable-safe body construction, durable labels, nested-ID rejection at every chunk, partial-state restart skipping, second-batch 502 stopping, active-halt refusal and unchanged old-root protection. Existing template, registry, primitive pacing and null-recovery controls remain passing. No Go, race, vet or source-audit suite was repeated.

This checkpoint is a draft code follow-up only. The original 502 and earlier null-create evidence remain intact. The native phase remains halted at WIRING, with no VERIFIED receipt. ROOT owns review, qualification and a separate release authorization before any actual wire invocation. Four appends/root-prefix work remains pending. Classes, runtime and FullRules remain UNQUALIFIED; source tracking issues 330/331/332 remain OPEN. No issue closure, merge, watch or deployment is part of this change.
