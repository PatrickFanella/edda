# Character Naming

Curated from a previously merged skill. This file keeps the useful task guidance without preserving the old skill as an active nested skill.

## Extracted Resources

- `scripts/character-naming/`
- `templates/character-naming/`

## Guidance

### From `character-naming-skill.md`

_Source topic: character-naming_

**Purpose:** Break LLM name defaults with external entropy. Use when character names cluster around statistical medians (Chen, Patel, Maya, Marcus), when cast has collision risks, or when fantasy cultures need phonologically consistent naming.

# Character Naming: Breaking the Chen Proliferation

You help writers generate character names that escape LLM statistical defaults. Your role is to diagnose naming problems, provide external entropy for generation, and track cast coherence.

## Core Principle

**LLMs default to statistical medians. External entropy is the only cure.**

When asked for "diverse" names, LLMs produce whatever names appear most frequently in their training data for each perceived category. "Chen" appears repeatedly because it's the statistical center of "East Asian surname." When corrected, LLMs "median-hop"—switching to the next most common name from another ethnicity rather than providing genuine variety.

The solution: never let the LLM pick names. Use curated lists with true randomization.

## The States

### State CN1: No Context
- What's the genre and setting?
- What time period?
- What cultures are present in this world?
- How diverse should the cast be?

### State CN2: Chen Proliferation
- What's the actual cultural distribution of your setting?
- Have you defined which cultures are present and in what proportions?
- What names have you already used?
**Interventions:** Use cultural name lists with external randomization. Never let the LLM "suggest" names—always draw from entropy.

### State CN3: Cultural Incoherence
- Does this fictional culture have defined phonological rules?
- What's the naming convention (patronymic, descriptive, clan-based)?
- What real-world cultures, if any, inspired this fictional one?
**Interventions:** Use phoneme presets for consistent sound patterns. For complex cultures, consider design-fictional-languages (if available).

### State CN4: Cast Collision
- What names have already been used in this project?
- What initial sounds are overrepresented?
- What syllable patterns dominate the cast?
...

# With cast collision checking
deno run --allow-read scripts/character-name.ts --culture korean --cast project-cast.json
```

**Options:**
- `--culture <name>` — Use specific cultural pool (chinese, anglo, hispanic, etc.)
- `--pool <name>` — Use mixed pool (contemporary-american, etc.)
- `--fantasy <preset>` — Generate from phoneme preset (elvish-like, harsh-fantasy, neutral)
- `--gender <m|f|n>` — Filter for gendered lists where available
- `--count <n>` — Number of names to generate (default: 5)
- `--syllables <range>` — Syllable range for fantasy names (e.g., "2-3")
- `--cast <file>` — Path to cast tracker JSON for collision checking
- `--full-name` — Generate given + surname combination
- `--json` — Output as JSON

### cast-tracker.ts
Manages cast tracking for collision detection and distribution analysis.

```bash

# Get suggestions for underrepresented cultures
deno run --allow-read scripts/cast-tracker.ts suggest
```

## Anti-Patterns

### The Chen Again
**Problem:** Correcting "Chen" by picking "Kim" or "Patel" is still median-hopping. You're just cycling through the top name from each ethnicity cluster.
**Fix:** Never let the LLM suggest alternatives. Use the entropy script to draw from deep in the list.

### The Diversity Checkbox
**Problem:** Adding exactly one character of each ethnicity feels like tokenism. The cast reads like a diversity compliance spreadsheet.
**Fix:** Base cultural distribution on setting logic. A story set in Seoul shouldn't have one of every culture. A story set in London can justify real diversity.

### The Unpronounceable Fantasy Name
**Problem:** Generated fantasy names are hard to read or say. "Xzylthrix" breaks immersion.
**Fix:** Use phoneme presets with pronounceability constraints. Limit consonant clusters. Test by reading aloud.

### The Cast Collision
**Problem:** Readers confuse Mark and Mike, Sarah and Sara, Lee and Leigh. Similar sounds blur together.
**Fix:** Always run cast-tracker check before finalizing. Analyze sound profiles—vary initial consonants, syllable counts, stress patterns.
...
```bash
```

### Example 2: Fantasy Novel
**User:** "I need names for my elvish kingdom."

**Your approach:**
1. Ask about the aesthetic—high fantasy, dark, whimsical?

### From `MERGED.md`

_Source topic: MERGED_

# Merged skill: character-naming
Source path: `skills/worldbuilding-fiction/character-naming`
Canonical skill: `skills/worldbuilding-fiction/naming`

### From `naming-framework.md`

_Source topic: naming-framework_

# Naming Framework

A diagnostic framework for creating names that work—brands, products, characters, places, titles.

## Problem Statement

Names fail in predictable ways: they don't feel right, they don't fit together, they're forgettable, they send wrong signals. The difference between a name that works and one that doesn't often comes down to invisible patterns—sound, meaning, cultural resonance—that most people can sense but can't articulate.

## Core Insight: Names Operate on Multiple Layers

1. **Sound Layer** — How it sounds, feels in the mouth, and what sounds evoke
2. **Meaning Layer** — Denotation, connotation, metaphor, reference
3. **Cultural Layer** — What associations it triggers in specific audiences
4. **Functional Layer** — How it works in use (typing, speaking, remembering)

When layers align, names feel inevitable. When they conflict, names feel wrong even if no one can say why.

### State N2: Names Don't Belong Together

- Is there a consistent sound palette?
- Do syllable structures match?
- Is there a unifying pattern (prefix, suffix, rhythm)?

- No phoneme inventory defined
- Mixed linguistic origins without intent
- Ad-hoc naming without system

- Define phoneme inventory (which sounds are "in" this system)
- Establish syllable templates (CV, CVC, CVCV, etc.)
- Create naming conventions document
- Regenerate outliers to fit system

### State N4: Name Sends Wrong Signals

- What does this name sound like it should be?
- What category conventions is it following/breaking?
...
