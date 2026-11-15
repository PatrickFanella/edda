# Rules-backed character interviews

Issue #83 enables the existing character interview for the SRD 5.2.1 slice.
The request names its rules version; legacy interviews keep their free-form
concepts. Changing the rules mode resets the character flow so an old interview
cannot be reused under a different contract.

The versioned tool proposes structured `rules_choices`: origin, class and
background. Human Soldier Fighter and Human Acolyte Cleric select exactly the
fixed builds documented in [character creation](character-projection.md) and
[caster creation](spell-actions.md). Required feats, skill choices, ability
scores, equipment and spells come from that catalogue. The interview cannot
invent or override these mechanics. The player reviews the same detailed
confirmation as guided creation and explicitly begins generation.

Missing or unsupported choices, inconsistent displayed choices and unexpected
mechanical fields produce a correction message without completing the interview.
Personality, strengths, weaknesses and backstory remain descriptive and do not
grant abilities. Both the HTTP build boundary and shared world orchestrator
validate the same catalogue before inference or writes; persistence uses its
existing derived contract. An interview cannot bypass this by authoring stats,
abilities, equipment or spells.

Provider failures restore the last successful server transcript. Retrying does
not duplicate the player message. Back/Continue retains the active client
transcript and server session. Sessions remain process-local: browser reload or
server restart recovery is not added here. A completed interview is followed by
confirmation; editing its mechanical choices requires another interview or the
guided form. No live-provider or production qualification is claimed.

Deterministic tests cover both classes, corrections, missing choices, invented
abilities, provider failure and retry. PostgreSQL tests compare actual persisted
sheets from guided and structured interview profiles through the HTTP boundary.
Desktop/phone browser fixtures cover correction, failure, Back/Continue resume,
explicit confirmation and generation. This is an adaptation of the pinned
SRD 5.2.1 fixed creation builds; source pages, attribution and CC-BY-4.0 license
are linked from the creation documentation.
