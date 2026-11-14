---
name: improve-game-design
description: "Game design theory: apply MDA, player psychology, balance and progression principles to explain or improve why a game is fun. Not for building the game (build-web-games)."
metadata:
  blacktower-pack: creative
  blacktower-kind: workflow
  blacktower-tier: standard
  blacktower-capabilities: games-interactive,game,design,theory
  version: "2.0.0"
---

# Game Design Theory

Use for system design, balance tuning, and player motivation diagnostics.

## Activation Map

Use this skill when task asks for any of:

- mechanics/dynamics/aesthetics alignment
- core loop design or loop quality diagnosis
- difficulty/flow curve tuning
- reward schedule design
- progression, economy, or PvP fairness balancing

## Core Model

1. Define target **aesthetic** outcomes (emotion/experience).
2. Model **dynamics** that should create those outcomes.
3. Implement **mechanics** that generate those dynamics.
4. Validate via playtests and telemetry.

Shortcut: `Mechanics -> Dynamics -> Aesthetics`.

## Diagnostic Heuristics

- **Boredom** -> challenge too low / weak goals / slow feedback.
- **Frustration/churn** -> spikes, unclear feedback, unfair outcomes.
- **Dominant strategy** -> missing counter-play or option asymmetry.

Fix in order:

1. challenge curve,
2. feedback clarity,
3. reward timing,
4. option balance/counters.

## Player Motivation Lens

Check support for:

- autonomy (choice/control)
- competence (mastery)
- relatedness (social connection)

Cover Bartle preferences where relevant: achiever/explorer/socializer/killer.

## Delivery Contract

When applying skill, return:

- target player experience
- proposed loop/mechanics changes
- expected dynamic effects
- balance risks + mitigations
- test plan and success signals

## Resources

- `references/full-guide.md` — full theory, troubleshooting patterns, detailed checklist.
- `references/DESIGN_GUIDE.md` — focused reference deep-dive.
- `scripts/design_analyzer.py` — deterministic analysis helper.
- `assets/design_config.yaml` — static config for analysis script.

Load deeper resources only when needed; keep this file activation-first.

Load [companion procedures](references/skill-routes.md) only when their stated condition applies.
