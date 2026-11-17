# Caster publication recovery guard

Native publication remains HELD. This change permits no immediate native mutation and does not lift or delete `native-write-halt.json`. The previous one-use authorization remains consumed. Both null responses and original journal entries remain unchanged. No allocation milestone, wiring, five old-root additions, parent supplements, final prefix release or native qualification is claimed. All future Tasks remain OPEN; runtime remains unavailable and FullRules false.

## Read-only reconciliation

A complete GraphQL graph and complete metadata read finished at 2026-11-17T00:03:40.961124+00:00. It found 1731 issues: the original 1017 plus exactly 714 retained allocations (330:305, 331:281, 332:128). Every retained identity, title, exact allocation body, OPEN state, Task/M metadata, absent parent and empty prerequisite set matched. Every old body/title/identity/parent/edge and baseline Project contract matched; no unexpected drift or extra identity was found. The failed dancing-lights marker remained absent. Decisions 330/331/332 were OPEN after ROOT restored their states. A preceding complete capture saw the earlier CLOSED states; it is retained separately.

Durable evidence directory: `/home/onnwee/.local/state/edda-overnight-219965e8/caster-publication/recovery-followup-readonly/`.

- `post-reopen-graph.json`: SHA256 `11fd444d1a73b5bdab3c478142b7793b609757c9c22395036ec084837a37fee1`.
- `post-reopen-metadata.json`: SHA256 `7dc4a06bfbad44b64f3b70b3619ca01ac373369abb714360933d1751ee36a666`.
- Registry unchanged: SHA256 `9c4531386a04f3113cb5665fec248357e8fd2328b73e50908f5b83db030dd296`.
- Journal unchanged: SHA256 `7413c2be14c16720b378219dfddd5716cc7d820cc9dd43f6df1843b062d804bd`.
- Active halt unchanged: SHA256 `1c9c43ce8a87dda1778fcc74f88122f5d6b79b4593ac6c06b1676c6b753a0264`.

## Timing boundary

The serial transport requires at least 7.5 seconds between mutation starts and no more than 480 starts in a rolling hour. It reconstructs timing from the retained journal on restart, including unresolved INTENT entries. ROOT's known state-only PATCHes at 00:00:02Z,00:00:04Z and00:00:06Z on November17 count in that budget. Future external writers must be accounted for before using this lane; this lane must remain the sole native Task writer for all 871.

The earliest recovery time is November17 at 00:04:50Z, one hour after the last ambiguous publication create. This does not claim an account-wide hour without writes. Unresolved journal operations impose their own full-hour cooldown. Retry-After (seconds or HTTP date) and exhausted primary-reset headers can extend the deadline. Restrictions over 60 seconds fail closed for a later invocation instead of blocking an agent. Timing is hypothesis mitigation; neither null response proves a rate-limit error, and the successful state PATCHes do not prove createIssue recovery.

A refusal before the HTTP transport runs is recorded as `NOT_DISPATCHED`. It does not install a new halt, count as a mutation start, extend the cooldown or fence an otherwise unissued create. A persisted unresolved intent still retains the conservative restart fence. This distinction does not bypass an existing halt.

GitHub recommends serial requests with at least one second between mutations and generally limits content generation to 500 requests/hour across REST, GraphQL and the web UI: [GraphQL limits](https://docs.github.com/en/graphql/overview/rate-limits-and-query-limits-for-the-graphql-api), [request best practices](https://docs.github.com/en/rest/using-the-rest-api/best-practices-for-using-the-rest-api).

Future mutation records retain the exact request payload and exact GraphQL response body with only whitelisted response headers. Credentials and unfiltered headers/stderr are never retained by the transport.

Every mutation field must return its expected nested Issue or Project item with a nonempty string identity before acknowledgement. This applies to every field in a bundled wiring request; a null sibling or missing field halts the writer. Created issues must also return positive integer issue/database identities. The exact single-operation `UNPROCESSABLE` response `Content already exists in this project` retains the allocator's bounded readback path. It is recorded as a known rejection and requires the existing Project item to be read; other errors or null results remain ambiguous and halt publication.

## ROOT authorization protocol

ROOT must review this exact code and read-only evidence before delivering a new immutable JSON authorization and its separately supplied SHA256. The helper requires `--authorization` and `--authorization-sha256`; the old authorization hash is explicitly rejected. No new authorization is created by this PR.

Required authorization fields:

- `status`: `ROOT_AUTHORIZED_ONE_RECONCILED_ATTEMPT`; integer `allowed_attempts`:1; `not_a_milestone`:true; `preserve_active_halt`:true.
- `approval_sha256` and `renderer_contract_sha256`: the unchanged original pins; `alias`: `332:spell-guards-and-wards-dancing-lights`; `original_intent_request_sha256`: `463189eda52ff60234ecb2b3557beaba52a0c0a3d887af38c58cf2d60b7d95fe`.
- `code_sha256`: exact reviewed file-byte hashes under keys `coordinator`, `recovery`, `guard` for the three corresponding Python files.
- `retained_sha256`: exact file-byte hashes under relative keys `native-write-halt.json`, `registry.json`, `publication-journal.json`, `recovery-20261116-null-create/one-attempt-consumed.json`, and `recovery-20261116-null-create/one-attempt-result.json`.

Under the existing exclusive coordinator lock, the helper checks the original authorization and retained 714 map, source pins, old null entry, new authorization, reviewed code and active halt. It checks cooldown, obtains a new complete graph and metadata snapshot, then reads the latest 100 issues immediately before the action. Any changed issue count or marker disagreement stops it. The snapshot must match every old contract and every retained allocation, including Task/M and unwired states. Only decisions 330/331/332 may be OPEN or CLOSED; closure never removes required nodes, source paths or terminal edges.

The later metadata response must independently agree with the full graph for every identity, title, body, state, parent and prerequisite set. A contradiction in either capture stops recovery before authorization consumption, even if the earlier graph matched the approved contracts.

The helper uses O_EXCL plus file and directory fsync to consume the new authorization before adoption or create. An exact existing marker is adopted without create. Otherwise it journals the original exact create intent and sends at most one identical GraphQL request. Typed identity and exact title/body/OPEN/Task/project/unwired readback are mandatory before registry adoption. An adopted existing issue must already be Size M; a freshly created issue may still need the normal allocator's paced Size M step. No milestone is emitted. Any ambiguity retains the halt and records a new result separately; the original halt bytes and old evidence remain intact. Restart cannot reuse consumed authorization.

Even a confirmed recovery leaves normal allocation blocked by the existing halt. ROOT must separately review the result and control the subsequent continuation boundary. No REST fallback, probe issue or skipped alias is available.

## Offline validation

Focused fake-clock/API/temp-journal tests cover pacing, rolling budgets including ROOT's PATCHes, restart INTENT fences, full-hour cooldowns, Retry-After and primary-reset restrictions, sanitized response capture, halt preservation, typed identities, unique exact markers, one-use null recovery and adoption without create. Independent source proofs exercise CLOSED decisions while rejecting removed source and terminal edges. Relevant Rogue shared-root and Wizard milestone guards are included. No live mutation probe or Go/race/vet rerun was performed.

Initial checkpoint validation: 67 focused Python tests passed. ROOT's repair adds six test methods covering nested/bundled results, later metadata contradictions, unsent cooldown refusals and the captured Project auto-add rejection. Against the direct parent, the new defect controls reproduced 14 expected failures with zero errors. The repaired focused suite passed all 73 methods in 22.835 seconds (27 coordinator/recovery/pacing, 14 Wizard publication, 14 Warlock approved-native, 7 Sorcerer binding, 5 Rogue shared-root compatibility, 6 Wizard milestone). These are offline controls, not native publication acceptance.
