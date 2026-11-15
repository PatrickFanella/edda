# Grounded action suggestions (#75)

The owner-guarded `GET /api/v1/campaigns/{id}/action-suggestions` returns at most
four candidates and an opaque `suggestions-v1` snapshot hash. It selects up to two
current player-known facts categorized as clues and fills the remaining slots
with supported Perception, Investigation and Athletics check intents validated
through the same read-only preparation as the sheet. Each candidate includes a
player-safe reason and source identifiers. Text is bounded; no hidden facts,
superseded clues, enemy data or unimplemented spell/item capabilities are used.
Inactive, unconscious and paused characters receive no suggestions. Legacy
characters may receive known-clue prompts; unsupported SRD checks are omitted.

This first candidate source is deterministic. It calls no inference provider,
spends no resources and promises no outcome. A player-known clue can motivate
investigation without implying that investigation will succeed. Skill prompts
explicitly leave circumstances and outcomes for adjudication. It does not add
catalogue support or choose optimal tactics.

A known current location grounds the Perception prompt; unknown location labels
are never disclosed. Location identity is included in the version fingerprint.

The web composer fetches candidates when opened, on browser-window focus and after completed
turns. Relevant character/clue changes change the snapshot version. Typing uses
local word-prefix matching with no per-keystroke requests. Up to two matches and
their reasons appear in a bounded internal region; selecting a match fills the
editable draft and never submits it. Assistance errors remove candidates while
ordinary play remains available. Campaign query keys isolate responses during
campaign switches. Ghost text and Tab acceptance remain #84.

Pure fixtures cover deterministic versions, supported intents and unavailable
actors. An isolated PostgreSQL/API fixture proves owner isolation, hidden-clue
exclusion, knowledge/HP refresh and no history writes. Frontend and desktop/phone
browser fixtures cover local matching, draft-only selection, refresh after a
turn, bounded document layout and playable assistance failure.
