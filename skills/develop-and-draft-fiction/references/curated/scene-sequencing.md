# Scene Sequencing

Curated from a previously merged skill. This file keeps the useful task guidance without preserving the old skill as an active nested skill.

## Extracted Resources

- `scripts/scene-sequencing/`

## Guidance

### From `scene-sequencing-skill.md`

_Source topic: scene-sequencing_

**Purpose:** Structure scenes and control pacing using scene-sequel rhythm. Use when individual scenes work but don't accumulate, when pacing feels off (too rushed or too slow), when transitions feel mechanical, or when readers can follow but aren't compelled forward. Based on Dwight Swain's Goal-Conflict-Disaster and Reaction-Dilemma-Decision structure.

# Scene Sequencing: Pacing Skill

You help writers structure scenes and control narrative pacing using the scene-sequel rhythm.

## Core Principle

**The fundamental unit of pacing is not the scene alone, but the scene-sequel pair.** Scenes create tension; sequels process it. The alternation creates peaks and valleys that make stories readable.

## Scene Structure: Goal → Conflict → Disaster

### Goal
- Specific and concrete
- Achievable within the scene
- Connected to larger story goals
- Clear to reader within first beats

### Conflict
- Another character with different agenda
- Environmental obstacle or time pressure
- Internal resistance (fear, doubt, values)

### Disaster
1. **Yes, but...** — Goal achieved, new problem created (strongest)
2. **No, and furthermore...** — Goal failed, situation worse
3. **No** — Goal failed, must try again
4. **Yes** — Goal achieved cleanly (use sparingly—kills tension)

## Sequel Structure: Reaction → Dilemma → Decision

### Reaction
- Process what happened
- Connect with character's emotional state
- Breathe between high-tension scenes

### Dilemma
- Closed some paths
...

# Get JSON output for further processing
deno run --allow-read scripts/analyze-scene.ts scene.txt --json
```

**What it detects:**
- Goal indicators (want, need, trying to)
- Conflict indicators (but, blocked, obstacle)
- Disaster indicators (failed, worse, trapped)
- Reaction indicators (felt, emotion, shock)
- Dilemma indicators (choice, either, what if)
- Decision indicators (decided, will, plan)

**Output includes:**
- Scene/sequel ratio assessment
- Pacing classification (action-heavy, balanced, reflective)
- Missing element warnings
- Specific recommendations

**When to use:**
- Quick diagnostic on a draft scene
- Identifying why a scene feels off
...


### From `MERGED.md`

_Source topic: MERGED_

# Merged skill: scene-sequencing
Source path: `skills/writing-craft/scene-sequencing`
Canonical skill: `skills/writing-craft/story-sense`
