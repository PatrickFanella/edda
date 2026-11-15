# Gameplay concept decision for #55

Status: selected by the owner on 2026-11-15 UTC: **A's layout with B's serif story text**. This investigation does not change the production UI. The interactive prototype opens with the selected combination.

Open [the interactive concepts](gameplay-concepts.html). Choose A or B, switch between desktop and phone arrangements, open Character/World/Journal, copy a suggestion into the editable draft, and try the reduced-height sketch. The menu demonstrates the proposed location of Saves, Export and campaign navigation.

## Recommendation and alternative

The selected combination uses **A: story with a context rail**, Marcellus headings, Josefin Sans controls, and Georgia serif story copy from B. Desktop keeps a 220-pixel context rail; phones replace it with a bounded detail view above the persistent composer. The transcript and details own their scrolling. This retains the current game's simultaneous character context without carrying all of its telemetry onto a phone.

**B: full reading column** removes the persistent desktop rail. A player opens details explicitly, temporarily replacing the transcript area. It uses Georgia for story copy, retaining Marcellus headings and Josefin Sans controls. The wider reading area gives long passages more space, but character state takes another action to inspect. The owner selected A's arrangement with this reading face.

Both show the same isolated fixture: Adventure alpha; Mira, level 1, HP 10/10; an open gate, the waiting keeper, one played action, and the silver-key narrative result. No unplayed events or hidden lore are used. Draft copy and navigation are demonstrations; Send makes no API request. The prototype adds no claim of implemented SRD mechanics.

## Design tokens and accessibility intent

The application owns its interface palette and typography; the identity-artwork pack explicitly does not replace them. These concepts retain the Art Deco application tokens from frontend/tailwind.config.ts and frontend/src/index.css:

- Obsidian #0a0a0a; charcoal #141414; champagne #f2f0e4; gold #d4af37.
- Jade #4ecca3 for vitals, sapphire #5b8fb9 for the player's action, ruby #ab2346 for Send. Muted reading text is lightened to #c7c3b8.
- Marcellus headings; Josefin Sans controls and A story; Georgia is B's optional reading face. The existing hosted font families have a font-host stylesheet. Georgia is a system serif; there is no added font package.
- Square corners, thin gold dividers, no card shadows or motion. Body text is 18 pixels with 1.6 line height; inputs are 17 pixels. Story width is capped at 64 characters (58 in B). Buttons have a minimum height of 44 pixels.

The dark theme and square Art Deco detailing retain the established application direction. The concepts do not introduce the marketing pack's separate purple palette. Each visible control has a text label, focus styling, and keyboard activation. Details can return to the triggering control; Escape closes the detail view. The modal menu uses the native dialog focus boundary. A live status announces draft acceptance; accepting it never submits. No animation needs a reduced-motion alternative.

## Bounded implementation handoff

- #60 owns the actual viewport frame: include the application credit within the height budget; use dynamic viewport sizing and nested min-height constraints; bound transcript/detail scrolling at the six roadmap sizes.
- #62 owns anchored input and choices: keep drafts when details open; provide an explicit, bounded choices region; do not make every turn increase the page height. Recheck an open phone keyboard instead of assuming a CSS resize proves it.
- #71 owns context navigation: retain Character (including inventory, skills and feats), World (including quests, NPCs and known facts) and Journal (including logs and later chapters). The campaign menu keeps Saves, Export and campaign-list access discoverable.
- #78 owns automated viewport, control reachability, zoom/reflow and real-device qualification. The reduced-height sketch here is not keyboard evidence.
- #90 owns transcript reading and retained position. Keep player actions distinct from GM prose with more than color.

No new follow-up issue is needed: these existing roadmap issues bound the implementation. The owner has recorded the arrangement and typography decision; implement that combination in wave 2. Human review of real phone/software-keyboard behavior remains in #46/#78.

## Validation and limits

On 2026-11-15, the HTML preview rendered at 728 and 390 CSS pixels without console errors. Scripted prototype checks exercised A/B, desktop/phone and normal/reduced-height states: Send remained inside the frame, the frame did not overflow internally, context opened/closed, and choice acceptance produced an editable draft without submission. These are prototype checks, not proof that the production application fits the viewport.

The current implementation assessment is separate in #46. Font appearance, comfort over long reading, native screen-reader behavior, increased text size and actual software-keyboard behavior still require the relevant human/device checks. No production build or deployment is included.
