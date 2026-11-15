# SRD 5.2.1 reference and conformance register

Edda pins the **English System Reference Document 5.2.1** from Wizards of the
Coast. The machine-readable register is
[`internal/rules/srd521/coverage.json`](../internal/rules/srd521/coverage.json).
It records coverage scopes, not an imported rules catalogue. No rules prose or
stat blocks are included in this slice.

All 59 initial entries are **unsupported**. This means Edda has no reviewed
SRD 5.2.1 conformance evidence for those scopes. Existing dice, skills, feats and
campaign behavior do not establish compliance. The legacy rules in
`internal/rules/skills.go` and `internal/rules/feats.go` retain their current
behavior; the new `srd521` package neither replaces them nor seeds campaign data.

## Pinned source

| Field | Value |
| --- | --- |
| Content version | SRD 5.2.1, English |
| Official landing page | https://www.dndbeyond.com/srd |
| Original artifact | https://media.dndbeyond.com/compendium-images/srd/5.2/SRD_CC_v5.2.1.pdf |
| SHA-256 | `8974902d109d6e63672d7c490bde9ccf052410503d9cfa768237154fbc5e3d87` |
| Size | 6,031,375 bytes |
| Pages | 364 |
| Retrieved | 2026-11-15 UTC |
| License | CC-BY-4.0 |

The PDF's page 1 gives the following attribution, retained verbatim in the
register and `srd521.Attribution` (with line wrapping removed):

> This work includes material from the System Reference Document 5.2.1 (“SRD 5.2.1”) by Wizards of the Coast LLC, available at https://www.dndbeyond.com/srd. The SRD 5.2.1 is licensed under the Creative Commons Attribution 4.0 International License, available at https://creativecommons.org/licenses/by/4.0/legalcode.

Coverage identifiers and notes are Edda additions. The repository's software
license does not replace the source material's CC-BY-4.0 license. Future imported
excerpts or fixtures must retain this source notice and identify changes.

## Reproduce the pin

From the repository root, download the original PDF to a new temporary directory
and check its bytes. This is an explicit maintenance command, never a CI download:

```sh
srd_dir=$(rtk proxy mktemp -d)
rtk proxy curl --fail --location --proto '=https' --proto-redir '=https' \
  --output "$srd_dir/SRD_CC_v5.2.1.pdf" \
  https://media.dndbeyond.com/compendium-images/srd/5.2/SRD_CC_v5.2.1.pdf
rtk proxy sh -c 'printf "%s  %s\n" \
  8974902d109d6e63672d7c490bde9ccf052410503d9cfa768237154fbc5e3d87 \
  "$1" | sha256sum --check --strict' sh "$srd_dir/SRD_CC_v5.2.1.pdf"
rtk proxy env EDDA_SRD_PDF="$srd_dir/SRD_CC_v5.2.1.pdf" \
  go test -count=1 -run '^TestPinnedArtifact$' ./internal/rules/srd521
```

`VerifyArtifact(io.Reader)` checks both size and SHA-256 against constants. A
different PDF at the same URL must fail; do not silently update the pin. The
original PDF and extracted text are not vendored. The opt-in artifact test skips
when `EDDA_SRD_PDF` is unset; all register validation tests run offline by default.

## Register contract (schema version 1)

`srd521.Load()` returns an independent decoded copy. It rejects unknown JSON
fields. `Register.Validate(repo fs.FS)` validates the source, entries and fixture
links. Tests use a filesystem rooted at the repository. Neither function performs
network requests or changes campaign state.

The top-level object contains `schema_version`, `source` and `entries`. The source
includes version, language, official URLs, artifact hash/size/page count,
retrieval date, license, exact attribution and a change notice. Each entry has:

| Field | Meaning |
| --- | --- |
| `id` | Stable, case-sensitive identifier with prefix `srd-5.2.1:` |
| `family` | Group used to select related coverage scopes |
| `name` | Human-readable scope being assessed |
| `source_pages` | Two integers: inclusive first and last PDF pages, starting at 1 |
| `status` | Exactly `unsupported` or `implemented` |
| `fixtures` | Explicit array of repository-relative Go test or `testdata` file paths |
| `follow_up` | GitHub issue number in `subculture-collective/edda`, or 0 when unassigned |
| `notes` | Evidence limits and the scope of the claim |

The initial inventory includes 15 rules/content families, all 12 classes and their
12 SRD subclasses, and character levels 1 through 20. Families include spells
(cantrips through level 9), ordinary equipment, magic items, monsters, animals,
exploration and the gameplay toolbox. Class scopes cover levels 1–20. Level
entries cover advancement at that character level; they do not imply that every
class feature at that level is implemented. Overlapping source ranges are
intentional because multiple sections can share a PDF page.

IDs use `family/<slug>`, `class/<class>`, `subclass/<class>/<subclass>`, and
`level/<number>`. Existing IDs must not be renamed or reused. Later slices may add
`content/<family>/<slug>` entries for individual features, spells, items or
monsters without importing a full catalogue. A different SRD content version needs
its own namespace and pin.

## Adding evidence in later slices

1. Add deterministic mechanical tests and fixtures for the exact named scope.
   Link actual nonempty files, using repository-relative paths without `..`.
   Imported source text must carry the attribution above and a change notice.
2. Keep containing families/classes/levels unsupported when only a smaller scope
   works. Add a narrower content entry for that support. Unsupported entries may
   link partial fixtures while their notes explain what remains unqualified.
3. Promote an entry to `implemented` only when tests substantiate its entire
   scope. Update the explicit reviewed-evidence list in
   `TestImplementationClaimsRequireReviewedEvidence` with that ID and its fixture
   paths in the same PR. The list is deliberately empty in this baseline.
4. Run the focused package tests and the repository checks below. Reviewers must
   assess the mechanical assertions: path validation alone cannot prove full
   SRD compliance.

```sh
rtk proxy go test -short -race -count=1 ./internal/rules/srd521
rtk proxy go test -short -race ./...
rtk proxy go vet ./...
```

Tests protect the stable initial inventory, pin and legal notice; reject malformed
IDs, duplicate IDs, invalid page ranges/statuses, unsupported full-support claims
and broken fixture paths; and ensure loaded copies cannot mutate the bundled
register. No database schema or query changes are required.
