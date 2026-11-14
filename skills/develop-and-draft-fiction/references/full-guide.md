# Story Diagnosis Full Guide

Detailed diagnostic reference. Keep `SKILL.md` short and routing-focused.

## Decision Tree

```text
Is there anything on the page?
├── NO → premise-forge
└── YES → What's the dominant failure mode?
    ├── Feels generic → edit-fiction (trope-refresher mode)
    ├── World feels thin → fictional-worldbuilding
    ├── Non-humans feel fake → design-fictional-languages
    ├── Characters flat → edit-fiction (character-arc-doctor mode)
    ├── Pacing off → scene-sequencing
    ├── Dialogue wooden → edit-fiction (dialogue-doctor mode)
    ├── Ending weak → write-and-revise-nonfiction (endings reference)
    ├── Meaning unclear → fictional-worldbuilding (moral-parallax reference)
    ├── Draft not progressing → draft-unblocker
    ├── Prose flat → edit-fiction (line edit)
    └── Draft needs revision → write-and-revise-nonfiction (draft-reviser mode)
```

## State Notes (Extended)

### State 0: No Story

- Symptom: blank page.
- Entry intervention: `develop-and-draft-fiction` (premise-forge mode).

### State 1: Concept Without Foundation

- Symptom: idea exists but no stable premise/pressure.
- Intervention: `edit-fiction`, `key-moments`.

### State 2: World Without Life

- Symptom: setting exists but does not produce consequences.
- Intervention: `fictional-worldbuilding` and subsystem work (belief/economy/governance).

### State 3: Flat Non-Humans

- Symptom: species differ cosmetically only.
- Intervention: `design-fictional-languages` + species logic constraints.

### State 4: Characters Without Dimension

- Symptom: characters act as plot devices.
- Intervention: `edit-fiction` + agency pressure design.

### State 4.5: Plot Without Pacing

- Symptom: scenes function locally but momentum decays.
- Intervention: `scene-sequencing`.

### State 5: Plot Without Purpose

- Symptom: events stack but don’t generate meaning.
- Intervention: `fictional-worldbuilding` (moral-parallax reference), `key-moments`.

### State 5.5: Dialogue Feels Flat

- Symptom: same voice across characters, no subtext.
- Intervention: `edit-fiction`.

### State 5.75: Ending Doesn’t Land

- Symptom: setup strong, payoff weak.
- Intervention: `write-and-revise-nonfiction` (endings reference in the draft-reviser mode).

### State 5.85: Draft Not Progressing

- Symptom: planning complete, pages not produced.
- Intervention: `develop-and-draft-fiction` (draft-unblocker mode).

### State 5.9: Prose Feels Flat

- Symptom: structural success, sentence-level weakness.
- Intervention: `edit-fiction`.

### State 6: Draft Complete, Needs Revision

- Symptom: draft exists, revision paralysis.
- Intervention: `write-and-revise-nonfiction`.

### State 7: Ready for Evaluation

- Symptom: story complete, quality/risk uncertain.
- Intervention: `edit-fiction`, `write-and-revise-nonfiction` (story-analysis reference in the draft-reviser mode).

## Script Usage

### `scripts/entropy.ts`

Inject creative randomness from curated lists.

```bash
deno run --allow-read scripts/entropy.ts lies
deno run --allow-read scripts/entropy.ts disasters --count 3
deno run --allow-read scripts/entropy.ts --combo
```

Lists: `lies`, `ghosts`, `disasters`, `dilemmas`, `professions`, `locations`, `collisions`, `openings`.

### `scripts/functions.ts`

Generate characters from abstract story functions.

```bash
deno run --allow-read scripts/functions.ts
deno run --allow-read scripts/functions.ts --setting scifi
deno run --allow-read scripts/functions.ts healer --setting fantasy
```

Functions include: `healer`, `enforcer`, `keeper_of_secrets`, `maker`, `trader`, `guide`, `entertainer`, `death_worker`, `transgressor`.

## Anti-Patterns (Extended)

1. **Prescribing instead of diagnosing**
   - Fix: ask clarifying questions before selecting intervention.
2. **Framework overload**
   - Fix: choose one intervention, then reassess.
3. **Ignoring writer energy**
   - Fix: balance precision with momentum.
4. **Treating structure as story**
   - Fix: keep testing emotional truth (“does this feel right?”).
