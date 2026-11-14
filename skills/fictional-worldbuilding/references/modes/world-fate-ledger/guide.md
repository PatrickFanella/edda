# world-fate-ledger procedure

Reviewed source clients: claude, codex, opencode. Check the actual client tools before following this procedure. Resolve its scripts, references and assets from this directory, not from the outer skill root.

# World-Fates

Manage long-term destiny of characters/factions/locations so world keeps moving.

## Purpose

Turn static power into dynamic consequence with transparent probability + human approval gate.

## When to Use

- Powerful entities never fall
- Campaign world feels frozen
- You need probabilistic decline/death rules
- You want persistent world-level consequence tracking

## When Not to Use

- Single-scene stakes only (use scene-level skills)
- You need deterministic scripted outcomes only

## Workflow

1. **Assess entity state** (F1–F5)
   - F1 stable ascent
   - F2 peak power
   - F3 overextension
   - F4 precarious
   - F5 fall in progress
2. **Calculate fate pressure** using formula inputs
3. **Apply three-tier exit model**
   - Tier 1 voluntary NPC transition
   - Tier 2 fate-offered choice
   - Tier 3 fate-forced roll (last resort)
4. **Generate proposal** with roll/severity/consequences/alternatives
5. **Require human approval** before fictional-worldbuilding application

## Output Checklist

- [ ] Pressure inputs recorded (power, tenure, vulnerability, protection, fortune, exposure)
- [ ] Current state assigned (F1–F5)
- [ ] Tier path selected (1/2/3) with rationale
- [ ] Fate roll + severity captured when applicable
- [ ] Proposal includes consequences + alternatives
- [ ] Changes marked Proposed pending approval

## Guardrails

- Never apply shifts without human approval
- Never introduce death risk without pressure + danger + depleted fortune
- Same rules for protagonists and NPCs
- No pressure from nowhere; trace to events/choices

## Resources

- Full guide: `references/full-guide.md`
- Scripts:
  - `scripts/fate-pressure.ts`
  - `scripts/fate-roll.ts`
  - `scripts/fate-choice.ts`
  - `scripts/propose-shift.ts`
  - `scripts/exposure-log.ts`
- Data:
  - `data/fate-tracking.json`
  - `data/shift-types.json`
  - `data/fate-choices.json`
- Templates:
  - `templates/fate-tracking.md`

## Integration / Handoffs

- `run-rpg-sessions` for session boundary triggers + live Tier 2 moments
- `fictional-worldbuilding` for canonical world bible updates
- `edit-fiction` for vulnerability/fortune evolution
