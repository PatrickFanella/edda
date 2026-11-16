# Cleric decision validation

Baseline: `32252f64380eb49dd47f7d1f53171b6a8791624e`. This is decision/publication work, with no runtime changes or production grants.

Completed local checks:

- `GOMAXPROCS=2 go test -p 2 -short -race ./...`: passed the full suite, including the SRD package and TUI. Run once on this unchanged runtime tree.
- `go vet ./...`: passed with default limits.
- Direct pinned-PDF comparators `verify_srd_cleric.py`, `verify_srd_life_domain.py`, `verify_srd_advancement.py`, `verify_srd_multiclass.py`: passed. Each received `/tmp/edda-wave1/SRD_CC_v5.2.1.pdf`; the PDF SHA-256 matches the committed pin.
- Origin, General, Fighting Style and Epic Boon feat source comparators: passed; the Fighting Style comparator also rejected 255 contrary-source fixtures.
- Existing shared scope pagination regression: passed both fixtures.
- Cleric pagination tests parse `per_page` and `page` using `urllib.parse`, reject repeated pages, and enforce a 100-call ceiling. No synthetic fixture can make unbounded API calls.

The final publication receipt and offline verifier provide exact title/body, parent/prerequisite, Task/Project Size, source-key, legacy-edge and separate containment/dependency cycle checks. Corruption probes remove each mandatory gate prerequisite and mutate source, owner keys, evidence hashes, titles, bodies, parents, Project Size and cycle edges. The readable progression table is checked against all actual catalogue cells.

The known unavailable linter shim was not changed. No toolchain or host configuration changed. No implementation-child PostgreSQL, browser, device, production, inference or Full Rules qualification was attempted; those remain child acceptance work. The draft PR's one hosted snapshot is reported separately at handoff. The parent owns independent acceptance, CI qualification and merging.

Final publication readback: 483 issue nodes, 75 implementation terminals plus Cleric gate551 and bounded Life spell gate628. All77 newly published issues have exact Task/Project Size metadata. Gate551 has98 direct blockers; gate628 retains14 terminals plus decision323. Every required native path, source/body/parent/prerequisite guard and both DAGs passed. All six Cleric verifier tests and both existing pagination tests passed. Existing382 remains untyped and unsized in the Edda Project; its original parent, body and edges are unchanged. Existing351/352 retain SizeL. No inherited metadata was rewritten.

Parent publication seal: all four exact supplements were applied serially to #47/#54/#283/#337 and read back with their original prefixes preserved. The refreshed receipt records them in full. A focused regression accepts decision closure while still rejecting removal of a terminal gate prerequisite; the direct-parent verifier rejects that intended closed state. The six verifier tests, complete live validation and corruption probes passed after the correction. No Go runtime code changed.
