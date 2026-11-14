# World-Fates — Full Guide

Detailed reference moved from `SKILL.md` for progressive disclosure.

## Core Principle

**Fortune is consequence of choice across time.**

Goal: keep world dynamic across campaigns. No permanent unchallenged power.

## Operational Mode

Hybrid loop:
1. Scripts compute pressure + weighted rolls
2. Model generates fate-shift proposals
3. Human approves/modifies/rejects before fictional-worldbuilding changes

## Fate States (F1–F5)

- **F1 Stable Ascent**: low pressure, monitor
- **F2 Peak Power**: moderate pressure, surface vulnerabilities/rivals
- **F3 Overextension**: high pressure, cascade small failures
- **F4 Precarious Position**: critical pressure, active fate-choice windows
- **F5 Fall in Progress**: terminal pressure, manage transition

## Three-Tier Exit System

### Tier 1 — Voluntary NPC Transition
Always available graceful player exit.

### Tier 2 — Fate-Offered Choice
At pressure > 0.6, offer dramatic branching:
- path to transformation/NPC transition
- path to risk roll

### Tier 3 — Fate-Forced Outcome
Roll consequence only when Tier 2 refused or impossible.

Design order:
1) visible pressure
2) dramatic choice
3) rolled consequence if needed

## Fate Pressure Formula

```text
Fate Pressure = (Power Level × Tenure Modifier × Vulnerability Score / 10)
                ÷ (Protection Factor × max(Fortune Buffer, 0.5))
                + Risk Exposure Accumulation
```

### Inputs
- Power level (1–10)
- Tenure modifier (0.5x to 2.5x)
- Vulnerability score (0–10)
- Protection factor (1–10)
- Fortune buffer (0–5, depletable)
- Exposure accumulation (+ per risky event)

## Wheel of Fate Mechanics

Trigger assessments at:
- session boundary
- arc completion
- time skip
- major world event
- manual trigger

Roll model:

```text
Roll d100
If roll < (fate_pressure × 100): fate-shift triggers
```

Severity bands: minor / moderate / major / catastrophic.

## Fate-Shift Types

- Characters: death, fall, corruption, exile, transformation, diminishment
- Factions: collapse, schism, absorption, decline, reformation, defeat
- Locations: destruction, conquest, disaster, decline, transformation, isolation

## Death Mechanics Gate

Death only enters probability pool when all true:
1. active danger
2. pressure > 0.5
3. fortune buffer < 1

Threshold:

```text
death_threshold = (fate_pressure - 0.5) × 2 × (1 - fortune_buffer / 5)
```

## Proposal Format

Include:
- trigger + pressure snapshot
- vulnerability analysis
- roll result + threshold + severity
- proposed shift + consequences
- alternatives
- approval checkbox/status

## World Bible Integration

Read from world bible for tenure/power/vulnerability/protection.

Write as **Proposed** changes first:
1. history event
2. entity status updates
3. approval gate -> **Established**

Track per-entity fate metadata:
- pressure
- fortune buffer
- vulnerabilities/protections
- exposure log

## Anti-Patterns

- Immortal emperor (no consequence)
- Random death (no buildup)
- Unequal rules for protagonists vs NPCs
- Instant collapse without cascade
- Foregone outcomes (no uncertainty)
- World-breaking shift without approval gate
- Cheap Tier 2 exits with no cost

## Tooling Reference

- `scripts/fate-pressure.ts`
- `scripts/fate-roll.ts`
- `scripts/fate-choice.ts`
- `scripts/propose-shift.ts`
- `scripts/exposure-log.ts`

Data/templates:
- `data/fate-tracking.json`
- `data/shift-types.json`
- `data/fate-choices.json`
- `templates/fate-tracking.md`

## Integration

- `fictional-worldbuilding`: canonical state surface
- `run-rpg-sessions`: boundary triggers and hook injection
- `edit-fiction`: vulnerability/fortune impacts
- `develop-and-draft-fiction`: genre-specific fate expectations

Handoff is two-way between `run-rpg-sessions` and `fictional-worldbuilding` at session boundaries.

## Operating Boundaries

Do not:
- apply fate shifts without human approval
- fabricate pressure without causal events
- bypass uncertainty with deterministic outcomes
- make Tier 2 choice trivial
