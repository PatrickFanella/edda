# Base procedure

Reviewed source clients: claude, codex, opencode.

Resolve command and resource paths from the owning skill root, unless a linked reference says otherwise.

# Story Diagnosis: Diagnostic Procedure

Diagnose story state, route to the right intervention, then reassess. This is a loop: Assess → Diagnose → Intervene → Reassess.

## Scope

- Use when writer is stuck, root cause unclear, or asks for diagnosis.
- Do not use when writer asks for direct prose generation (use the story-cowriter mode).
- Do not use when writer wants coaching-only questioning (use the story-mentor mode).
- Do not use for publishing/marketing positioning (use `write-and-adapt-copy`).

## Core Principle

There is no permanent “stuck.” There is only unresolved diagnosis or wrong intervention.

## Diagnostic Process

1. Listen for symptoms.
2. Ask clarifying questions.
3. Identify current state.
4. Name diagnosis clearly.
5. Recommend one intervention first.
6. Reassess after user applies it.

## The Story States

| State | Symptoms | Primary intervention |
|---|---|---|
| 0 No story | Blank page | `develop-and-draft-fiction` (premise-forge mode) |
| 1 Weak foundation | Idea exists, thin grounding | `edit-fiction`, `key-moments` |
| 2 World thin | Setting feels decorative | `fictional-worldbuilding` |
| 3 Non-humans flat | Species read as reskinned humans | `design-fictional-languages` |
| 4 Characters flat | Plot puppets, low agency | `edit-fiction` |
| 4.5 Pacing weak | Scenes don’t accumulate | `scene-sequencing` |
| 5 Meaning weak | Events lack thematic force | `fictional-worldbuilding` (moral-parallax reference), `key-moments` |
| 5.5 Dialogue flat | Same-voice / lifeless talk | `edit-fiction` |
| 5.75 Ending weak | Resolution underdelivers | `write-and-revise-nonfiction` (endings reference in the draft-reviser mode) |
| 5.85 Draft blocked | Planning done, no drafting | `develop-and-draft-fiction` (draft-unblocker mode) |
| 5.9 Prose flat | Functional but not vivid | `edit-fiction` |
| 6 Revision stage | Draft exists, edits overwhelming | `write-and-revise-nonfiction` |
| 7 Evaluation stage | Story complete, quality unknown | `edit-fiction`, `write-and-revise-nonfiction` (story-analysis reference in the draft-reviser mode) |

## Anti-Patterns

- Prescribing before diagnosis -> ask clarifiers first.
- Recommending too many frameworks at once -> start with one intervention.
- Ignoring writer energy -> pick the smallest energizing next move.
- Treating structure as sufficient -> verify emotional/narrative truth.

## Related Skills

Routes to all fiction skills based on diagnosed state.

## Resources

- `references/`
  - `references/full-guide.md` — detailed state notes, decision tree, and extended anti-pattern guidance.
- `scripts/`
  - `scripts/entropy.ts` — randomness injection for ideation prompts.
  - `scripts/functions.ts` — character generation by story function.
- `data/`
  - `data/functions-forms.json`
  - `data/genre-elements.json`

### Curated references

Former merged-skill dumps were distilled into these lookup files:

- `references/curated/flash-fiction.md` — Flash Fiction guidance.
- `references/curated/identity-denial.md` — Identity Denial guidance.
- `references/curated/key-moments.md` — Key Moments guidance.
- `references/curated/scene-sequencing.md` — Scene Sequencing guidance.

## Companion skills

Load [required steps and optional helpers](skill-routes.md) only when their stated condition applies.
