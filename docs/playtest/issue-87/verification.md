# Mechanical verification and review report

Observation date: 2026-11-15 UTC. Base: `e6fd912` (chapter reader PR #130).
Scope: original synthetic campaign fixtures, deterministic chapter examples,
journal structural tests and playtest documentation for #87. No application,
schema, export or frontend source changed. No production campaign or paid model
was used.

## Results

| Check | Observed result |
| --- | --- |
| `rtk proxy go test -short -race ./...` | Passed across all packages. Includes the three curated fixture cases and 27 deliberately broken example controls. |
| `rtk proxy go vet ./...` | Passed. |
| `rtk proxy go test -tags integration -run TestIntegrationCuratedChapterFidelity -race -count=1 -timeout 5m -v ./internal/journal` | Passed all three campaigns against a fresh `pgvector/pgvector:pg16` Testcontainers database. Container terminated by the test. |
| `rtk proxy go test -tags integration -run Integration -count=1 -timeout 10m ./internal/journal ./internal/handlers ./internal/state/sqlc` | Passed all three packages against their disposable PostgreSQL databases. |
| `rtk proxy pnpm --dir frontend install --frozen-lockfile` | Passed with the repository's pinned pnpm 10.33.0. |
| `rtk proxy pnpm --dir frontend test` | Passed: 13 files, 43 tests. |
| `rtk proxy env EDDA_FIXTURE_PORT=4197 pnpm --dir frontend test:browser --workers=2` | Passed: 56 checks, including chapter generation/retry and chapter source/reading-position behavior at desktop and phone widths. Browser emulation, not a human device review. |
| `rtk proxy pnpm --dir frontend build` | Passed. Existing bundle-size warning for the main chunk above 500 kB. |
| `rtk proxy pnpm --dir frontend lint` | Failed: 22 existing errors outside this change. No frontend diff against the base; the same baseline is documented in `docs/chapters.md`. |
| `rtk proxy golangci-lint run ./internal/journal/...` | Tool unavailable through the active Go 1.27.3 mise shim. Direct invocation of the installed Go 1.27.0 linter failed to decode the newer Go export data (version 5 vs supported 4). No successful golangci-lint result is claimed. |
| `rtk proxy git diff --check` | Passed. |

Go used `go1.27.3 linux/amd64`; Node used `v24.22.1`. PostgreSQL suites have their
own disposable databases; no application connection string was supplied.

## Self-review

The frozen source was compared with all ten original turns, including exact
names, actions and visible outcomes. Both the first generation request and
regeneration after edits retained the original source. The integration checks
excluded hidden facts, raw tool results, pending narration and completed turns
beyond the pending gap. Chapter persistence retained the deterministic response.

The fixed-example oracle detects missing names/outcomes and the enumerated
contradictions and private markers. It cannot judge arbitrary prose or a
paraphrased spoiler. The examples are explicitly labeled as authored doubles,
and the owner feedback is recorded separately in human-review.md. No corrective
chapter failure was observed in the initial mechanical runs. The later human
review identified suggestive repetition in one closing clause; F87-1 records its
removal and the new negative control. Intentionally rejected controls are test
evidence, not observed product defects.

The frontend lint failures are confined to `TurnNotifications.tsx`, `MapPanel.tsx`,
`ReplayNarrative.tsx`, `CampaignAttributesForm.tsx`, `CharacterGuidedForm.tsx`,
`AuthContext.tsx`, `CampaignContext.tsx`, `useAudio.ts` and `useDiscoveryBadge.ts`.
If assigned, correct that existing React hooks/refresh lint debt in a separate PR
and rerun lint plus frontend behavior checks. Toolchain maintenance can separately
install a linter compatible with Go 1.27.3 and rerun the journal lint check.
Neither is folded into chapter fidelity work.

## Human acceptance evidence

**The owner reviewed the three curated examples and gave overall favorable
feedback with one specific wording concern.**
[human-review.md](human-review.md) records that feedback without inventing separate
human row-by-row ratings, along with the agent source comparison and bounded
correction. The revised clause has not received a second human sign-off.
Hosted CI status must be read separately from these local results; this report
does not claim hosted CI, merge, deployment or live-provider qualification.

## Follow-up after owner feedback

The owner reviewed candidate `b937b5e` in the Wave 7 conversation. F87-1 removes
the redundant “no route beneath the kiln” clause from the chapter without changing
the frozen source transcript. A separate suggestive-phrase negative control
guards this editorial correction. The existing failed-search and cart/waiting
anchors still verify the played outcomes. Focused journal race tests, vet and the
curated PostgreSQL integration race test passed after this correction.
