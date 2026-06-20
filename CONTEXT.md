# Edda architecture context

## Domain terms

- **Campaign** — a playable run.
- **Startup wizard** — the flow that creates or resumes a Campaign.
- **World creation** — the Module that turns setup inputs into a persisted playable world.
- **Campaign play** — the play session view for an active Campaign.
- **Narrative turn** — one player action through streamed GM response to final choices/status.
- **Quest** — a tracked goal in the Campaign.
- **Objective** — a quest step or subgoal.
- **World fact** — persisted lore or state fact about the world.
- **Location** — a place in the world map.
- **NPC** — a non-player character.
- **Save point** — a persisted checkpoint for reload/resume.
- **Campaign time** — the campaign clock advanced by gameplay.
- **Rules mode** — the ruleset/profile controlling gameplay behaviour.
- **Combat** — the tactical presentation/state inside Campaign play.
- **Session log** — the durable record of player and GM activity.

## Architecture words

- **Module** — anything with an interface and an implementation.
- **Interface** — everything a caller must know: types, invariants, error modes, ordering, config.
- **Implementation** — the code inside.
- **Depth** — leverage at the interface.
- **Deep/Shallow** — high leverage behind a small interface / interface nearly as complex as implementation.
- **Seam** — where an interface lives.
- **Adapter** — concrete thing satisfying an interface at a seam.
- **Leverage** — what callers get from depth.
- **Locality** — what maintainers get from depth.
- **Deletion test** — if a Module were removed, would complexity reappear at the call sites?
- **Interface as test surface** — tests should exercise the Module's interface, not its hidden internals.
- **One adapter hypothetical / two adapters real** — keep a seam only when more than one real adapter exists or is likely soon.
