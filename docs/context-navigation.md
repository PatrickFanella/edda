# Gameplay context navigation (#71)

The four destinations are Play, Character, World and Journal. Character contains
the sheet and inventory. World keeps Map/Facts/Codex/Relationships and includes
quests and NPCs. Journal keeps Chronicle/Notes and a disclosure for Logs.
Account controls, campaign list/replay and the saves/export menu keep their routes.

At 1280 CSS pixels and wider, context opens beside the mounted story. Below that,
context uses the workspace while Play remains mounted and hidden. Close details
or Escape returns to Play, restores its prior scroll offset, and returns focus
to the story control that opened context (or the Play tab if no control was saved).
Tab arrow keys retain the existing automatic activation and ARIA relationships.
Draft and textarea selection retention come from #62.

Browser fixtures cover four-destination keyboard navigation, nested capability
access, desktop/phone visibility, draft, focus, reading position, long-content
scroll owners and all six roadmap viewports. This does not claim screen-reader
or real software-keyboard device qualification.
