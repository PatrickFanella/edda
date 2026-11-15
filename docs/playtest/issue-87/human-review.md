# Human chapter review — pending

Issue: [#87](https://github.com/subculture-collective/edda/issues/87), roadmap
[#47](https://github.com/subculture-collective/edda/issues/47).

Review status: **PENDING — no human review has been performed or recorded.**
Reviewer: not assigned. Review date: not performed. Reviewed commit: not recorded.
Candidate provenance: authored synthetic examples returned by the deterministic
model double; not live-provider output. Automated results are recorded separately
in [verification.md](verification.md).

## Instructions for the human reviewer

Replace the pending entries below with your name, date and the exact commit you
read. Read [campaigns.json](campaigns.json)'s `source.source_turns` against the
linked chapter for each row. Keep `hidden_lore`, `tool_secret`, `pending` and
`later` outside the story's allowed knowledge. Judge plot fidelity and readability
separately as pass / revise / fail. Cite at least one source turn and chapter
sentence per example; explain any concern with the specific wording. Record a
readability reason even when the plot is correct. Leave unreviewed rows pending.

| Example | Plot fidelity | Readability | Source turn → chapter sentence evidence | Reviewer findings |
| --- | --- | --- | --- | --- |
| [Failed plan](failed-plan.chapter.md), turns 1–3 | Pending | Pending | Not reviewed | Check failure, rope loss, retained payment, offered vs played route, unresolved water. |
| [Corrected misunderstanding](corrected-misunderstanding.chapter.md), turns 1–4 | Pending | Pending | Not reviewed | Check initial suspicion vs fact, later correction, apology and lost help. |
| [Undiscovered secret](undiscovered-secret.chapter.md), turns 1–3 | Pending | Pending | Not reviewed | Check failed search, unknown marks and no vault/hatch discovery or omniscient hint. |

Readability criteria: clear chronology and subjects; concise prose without erasing
consequences; natural transitions; no confusing pronouns or invented dialogue.
Fidelity criteria: names/places correct; intentions distinct from outcomes;
corrections retained; costs and failed plans retained; no unplayed choice or later
event; no hidden lore or invented resolution.

## Findings and bounded corrective follow-ups

Observed human findings: none recorded; review pending. No human pass is implied.

For each actual failure, record: finding ID; fixture and source turn/log ID;
offending chapter sentence; expected faithful meaning; severity (spoiler,
contradiction, omission or readability); smallest corrective change and owner;
tracking issue/PR; exact regression check and human re-review required. Do not
create speculative product issues from the intentionally broken test controls.

## Checkpoint decision

Decision: **PENDING**. All three fidelity and readability judgments are required.
Human reviewer/parent must record approval or corrective work here before marking
the PR ready or replacing `Refs #87` with a closing reference. Automated structural
checks alone do not satisfy the human acceptance criterion. This review covers
these candidate examples; it does not qualify every provider or future chapter.
