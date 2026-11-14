# Interactive Fiction — Full Guide

Detailed reference moved from `SKILL.md` for progressive disclosure.

## Core Principle

**Agency and authorship coexist.**

Best IF = meaningful choices + authored emotional payoff + constrained possibility space.

## IF Types

- Parser-based (freedom, high friction risk)
- Choice-based (manageable, false-choice risk)
- Hybrid/VN/RPG (rich state, high authoring burden)
- Tabletop scenario (dynamic, facilitator-dependent)

## Diagnostic States (IF1–IF7)

### IF1 Meaningless Choices
- Symptom: choices converge, players disengage
- Apply meaningful-choice test:
  1) distinct options
  2) perceivable consequences
  3) irreversibility
  4) character expression
- Fix: delayed consequences, values-based choices, visible effects

### IF2 Unmanageable Branching
- Symptom: exponential path explosion
- Fix: foldback, bottlenecks, flags over parallel universes, quality/storylet systems

### IF3 False Choice Discovered
- Symptom: trust collapse after detected illusion
- Fix: fewer real choices; if outcome same, make expression value explicit

### IF4 Agency vs Authored Meaning
- Symptom: incoherence from freedom or pointless interactivity from rigidity
- Fix: constrained agency via fiction-first limits (time, place, info, role)

### IF5 Story Feels Like Flowchart
- Symptom: menu interrupts narrative flow
- Fix: scene-first writing, breathe between choices, integrate decisions into drama

### IF6 Multiple Endings, No Satisfaction
- Symptom: one “true” ending invalidates others; “bad” endings feel punitive
- Fix: endings as different value-true outcomes, not ranking ladder

### IF7 State Management Chaos
- Symptom: contradiction, variable bloat, unmaintainable logic
- Fix: reduce state model, group flags, ensure visible consequence per tracked variable

## Branching Structures

- **Linear with windows**: mostly linear, local variance
- **Bottleneck**: multiple routes through fixed beats
- **Branch + bottleneck**: replayable routes, convergent control points
- **Time loop**: repeated structure, evolving player knowledge

## Choice Design

Prefer dilemmas over right/wrong binaries.

Useful choice types:
- Dilemma
- Expression
- Strategic
- Discovery

Avoid default binary moral optimization loops.

## Anti-Patterns

- The Maze (navigation trap)
- The Illusion (fake consequence)
- The Optimization Game (single best ending)
- The Completionist Trap (exhaustive replay burden)
- The Info Dump Choice (menu exposition)
- The Parser Nightmare (intent-expression mismatch)

## Diagnostic Sequence

1. Identify failing state (IF1–IF7)
2. Identify IF format (parser/choice/hybrid/tabletop)
3. Run meaningful-choice test on critical choices
4. Recommend structure + state interventions

## Output Persistence

Persist:
- IF state diagnosis
- branch complexity notes
- choice quality assessment
- agency model notes

Suggested file pattern:
- `{project}-if-{date}.md`

## Integration

### With story diagnosis (develop-and-draft-fiction)
- Story diagnosis pacing/end states map strongly to IF5/IF6.

### Handoffs
- `develop-and-draft-fiction` (scene-sequencing reference) for branch scene structure
- `edit-fiction` for player-driven transformation consistency
- `edit-fiction` for choice/subtext quality

## Operating Boundaries

Do not:
- design full branch tree for writer
- write all choices/outcomes for them
- promise simple solutions to branching math

Role remains diagnostic + assistive.
