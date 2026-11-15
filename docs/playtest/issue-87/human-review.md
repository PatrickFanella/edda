# Human chapter review

Issue: [#87](https://github.com/subculture-collective/edda/issues/87), roadmap
[#47](https://github.com/subculture-collective/edda/issues/47).

Reviewer: repository owner, in the T3 conversation requesting Wave 7 completion.
Review received: 2026-11-15 UTC (November 14 Central).
Reviewed candidate commit: `b937b5e37e63ea52e27039cf12ec582331201297`.
Candidate provenance: authored synthetic examples returned by a deterministic
model double. These are not live-provider outputs.

The owner's overall assessment was “i think it's all pretty good”, with one
specific concern: “and she had found no route beneath the kiln” could give away
that a route exists. This records the owner's actual feedback, not a separate
human verdict invented for each row. The source-linked assessment below is the
implementing agent's analysis of that feedback.

| Example | Human feedback | Agent plot-fidelity assessment | Agent readability assessment and source evidence |
| --- | --- | --- | --- |
| [Failed plan](failed-plan.chapter.md), turns 1–3 | Overall favorable; no specific objection | Pass for this example. Turn 1's failed crossing and lost rope remain failed/lost; turn 2's refused payment stays with Mara; turn 3 remains a wait. | Clear sequence. “The crossing failed” and “Neither of them had crossed” retain the outcome. The offered upstream bridge does not become a journey. |
| [Corrected misunderstanding](corrected-misunderstanding.chapter.md), turns 1–4 | Overall favorable; no specific objection | Pass for this example. Turn 1's suspicion stays a suspicion; turn 2 finds the tally under Tamsin's map; turns 3–4 retain apology, refusal and departure alone. | Clear correction and consequence: “found the tally beneath her own folded map” and “She left alone for the east stair”. No invented confession or renewed guidance. |
| [Undiscovered secret](undiscovered-secret.chapter.md), turns 1–3 | Overall favorable with a specific concern about the final clause suggesting a route exists | Revised. The source itself contains the flagged clause in turn 3, so this was suggestive repetition, not an invented factual discovery. The source remains unchanged. | Removed that redundant closing clause. The chapter now ends “The marks remained unexplained.” Turn 1 already states “she found no entrance”; turn 2 retains Noll's uncertainty; turn 3's cart-covering and waiting remain. |

## Finding F87-1 and bounded correction

- Boundary: authored chapter example, final sentence; source turn 3, log
  `87000000-0000-4000-8000-000300000003`.
- Concern: repeating a failed search as “no route beneath the kiln” can sound
  like an omniscient hint, even though the source uses that wording.
- Correction: remove the redundant clause from the example. Retain the original
  transcript exactly, the earlier failed search and the unresolved marks.
- Regression: the fixed-example negative controls now reject reintroducing that
  clause. The turn-3 outcome anchor cites the actual cart-covering/waiting action;
  the failed-search anchor remains in turn 1.
- Agent re-review: the revised two paragraphs retain all played events and add
  no route, hatch, vault, decoding or other discovery. Structural and database
  checks are recorded separately in verification.md.

The owner has not separately reviewed the revised sentence. This small editorial
correction responds directly to the reported concern; no additional human
sign-off is claimed. The issue's human-review checkpoint is recorded above.
This qualifies the curated examples and source boundary, not arbitrary models,
future prose, real-phone behavior or production deployment.
