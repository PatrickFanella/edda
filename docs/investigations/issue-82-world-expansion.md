# Monster and exploration expansion decision (#82)

Observed 2026-11-15 against Wave 9 main `2aa4f55` and the casting merge
`f88d2ac` (#142). This investigation delivers a source inventory, bounded
follow-up issues and dependency links. It does not implement those follow-ups.

## Source and inventory boundary

The English SRD 5.2.1 PDF was downloaded from the official pinned URL and its
SHA-256 matched `8974902d109d6e63672d7c490bde9ccf052410503d9cfa768237154fbc5e3d87`.
The PDF index on pp. 2–4 names **331 distinct entries**: 235 monster/template
entries and 96 animals. Every indexed name was found on its cited PDF page.
Ogre Zombie remains a monster despite sharing p. 344 with the Animals heading.
Half-Dragon is a template, not a fully instantiated ordinary creature.

The spell section also contains four creature blocks outside that index:
Animated Object (p. 109), Otherworldly Steed (p. 131), Giant Insect (p. 136), and
Draconic Spirit (p. 166). These bring the planning inventory to **335 entries**.
Find Familiar uses animal forms; its control and spell lifecycle still need
support even though those forms already appear in the index.

The companion `issue-82-world-inventory.json` contains source names/pages,
unsupported status and exact owning data-batch issue numbers. This is planning
metadata. It contains no imported stat values or descriptive prose, and does
not establish playable conformance. The inventory adapts source indexing;
retain [the pinned source attribution and CC-BY-4.0 notice](../srd-5.2.1-conformance.md).

## What the current code actually supports

| Boundary | Evidence | Remaining gap |
| --- | --- | --- |
| Encounter state | `internal/combat/encounter.go`, `internal/game/encounter_action.go`, `internal/engine/encounter.go`; deterministic and PostgreSQL encounter fixtures | Exactly two actors, ordinary hard-coded NPC AC/initiative/attack, 30-foot budget; no arbitrary stat-block execution, party ordering or general NPC policy |
| Narrative space | `internal/combat/space.go`, spatial handler integration tests | Exact symmetric distance, declared sight/cover, five-foot approach/retreat; no full light/senses/hiding, size/occupancy, obstacles, difficult terrain or movement modes |
| Reactions | `internal/combat/reaction.go`, reaction handler integration tests | One explicit longsword opportunity-attack trigger before a fixture withdrawal; no general monster reactions, legendary actions or arbitrary triggers |
| Monster content | `internal/rules/srd521/coverage.json`, `internal/game/stored_action.go` | No reviewed SRD stat-block catalogue; every named monster, template and animal remains unsupported. A generic NPC with a free-form name/stats is not a source-backed monster |
| Summons | No versioned creature-instance/control resolver; #68 only supports Cleric self-healing | All spell-created blocks, caster scaling, summon ownership, command/initiative, attack restrictions, dismissal and lifecycle; Find Familiar and transformations also remain unsupported |
| Travel and time | `internal/game/location_service.go`, `internal/tools/move_player.go`, `internal/tools/advance_time.go` | Connected-location movement and clock arithmetic exist; pace, distance/terrain, navigation, foraging, forced travel and resource/exposure consequences have no SRD evidence |
| Environment/toolbox | Rules register families `exploration` and `toolbox` remain unsupported | Vision/light, hiding, object interactions, hazards, weather, falling, food/water/suffocation, curses/contagions, fear/stress, poisons, traps and encounter-building budgets |
| Spell overlap | `docs/spell-actions.md`, #72, #77 and #80 | Casting a supported self-heal does not establish summoning, conditions, concentration, duration, recovery or all spell effects |

The existing guards in `internal/tools/rules_guard.go` keep model-driven legacy
mechanics from establishing versioned SRD rules authority. Keep that boundary.
Reuse the existing transaction, action identity, revision and player-known
projection contracts instead of creating a separate monster action engine.

## Batch decision

Separate **catalogue data**, **playable mechanics**, and **qualification**:

1. Introduce one validated stat-block contract with an initial Bandit fixture.
   Required mechanic keys fail closed. A parsed record is not playable unless
   every required trait/action/effect is supported.
2. Transcribe the exact index into independent data-only PRs of at most 12
   entries, plus one four-record spell-created batch. Preserve per-record source
   links, adaptation notice and required mechanic keys. Do not implement new
   behavior in these data PRs or promote whole-family coverage.
3. Deliver named runtime fixtures independently: Bandit attacks, Wolf Pack
   Tactics with size-gated Bite Prone (the 5.2.1 Wolf does **not** require a save),
   Brown Bear Multiattack, encounter budget calculation, a one-hour travel
   segment, declared light/Darkvision, an ordinary unlocked door, and falling.
4. Use bounded decision spikes for Find Familiar/Owl control, the four other
   spell-created creature lifecycles, remaining named hazards, and the final
   distinct unsupported mechanic-key inventory after data imports. These spikes
   must publish one-PR implementation children with exact source/fixture limits,
   attach native blockers to full qualification, and leave unknowns unavailable.
5. Complete full world-coverage qualification only after every required record
   and mechanic has evidence or an explicit owner-approved exclusion. More
   blockers than five reflect real independent requirements; do not manufacture
   a serial chain to conceal that fan-in.

No initial fixture or data import covers the whole SRD. The full catalogue,
monster abilities, exploration and toolbox families stay unsupported today.
The advanced-mechanic and hazard spikes must extend this same roadmap; closing
#82 or those splitters cannot close #54 by itself.

## Unknowns retained for the bounded spikes

The exact per-block trait/action dependency union is not yet encoded. Derive it
while transcribing records rather than guessing from names. Monster resistances,
immunities, condition interactions, save riders, grapples/swallowing, recharge,
limited-use recovery, spellcasting, bonus actions, reactions and legendary actions
need independent fixtures and may split further. General multi-creature geometry,
party turn ordering, templates and encounter outcomes/rewards also need explicit
follow-ups when their required keys are reconciled. Imported content must never
silently discard one of those fields to become playable.

Summon-specific spell prerequisites cannot be settled until #77/#80 publish their
contracts. Their spikes must add the real native edges; a parenthetical issue
reference does not implement or unblock those mechanics. The environmental
spike must enumerate actual named poisons/traps/contagions from pp. 193–201,
not treat a paragraph of generic hazard advice as one implementation PR.

## Tracking and qualification

All published children belong to epic #54 and roadmap #47. The native GitHub
blocked-by edges define the execution order below. #88 continues to report
actual supported coverage and block premature Full Rules claims; it does not
have to wait for every catalogue import merely to display unsupported scopes.
#89 qualifies the declared playable slice, and is distinct from full world
coverage. The new world qualification issue is the exit gate for this family.

The decision publishes **43 work issues**, all with native parent #54 and Size M
in the Edda Project: 29 data batches, nine contract/runtime slices, four bounded
spikes and one full-world qualification PR. No automation queue labels were added.

After #82 closes, the immediate frontier is #143 (catalogue contract), #144
(travel), #145 (light), #146 (door) and #147 (falling). The longest currently
specified additional chains have five issues including #82; for example
#82 → #143 → #148 → #179 → #183. The follow-up spikes can extend those chains.
The familiar and spell-created lifecycle spikes also depend on #80 rather than
pretending the unresolved spell/effect contract already exists.

| Issue | Unit | Direct blockers | Records |
| --- | --- | --- | --- |

| [#143](https://github.com/subculture-collective/edda/issues/143) | Validate versioned SRD monster records and fail closed on unsupported mechanics | #82 | — |
| [#148](https://github.com/subculture-collective/edda/issues/148) | Run a Bandit attack from its pinned stat block through durable encounters | #143 | — |
| [#179](https://github.com/subculture-collective/edda/issues/179) | Resolve Wolf Pack Tactics and size-gated Bite Prone through durable encounters | #148 | — |
| [#180](https://github.com/subculture-collective/edda/issues/180) | Resolve the Brown Bear Multiattack as one atomic NPC action | #148 | — |
| [#149](https://github.com/subculture-collective/edda/issues/149) | Calculate SRD encounter XP budgets without inventing rewards | #143 | — |
| [#144](https://github.com/subculture-collective/edda/issues/144) | Prepare and commit a one-hour SRD travel segment with explicit pace | #82 | — |
| [#145](https://github.com/subculture-collective/edda/issues/145) | Adjudicate declared light and Darkvision for narrative target visibility | #82 | — |
| [#146](https://github.com/subculture-collective/edda/issues/146) | Confirm opening an ordinary unlocked door through narrative-space facts | #82 | — |
| [#147](https://github.com/subculture-collective/edda/issues/147) | Resolve one declared fall with inspectable damage and Prone state | #82 | — |
| [#181](https://github.com/subculture-collective/edda/issues/181) | Bind one Find Familiar creature to its summoner and turn permissions | #80, #148 | — |
| [#150](https://github.com/subculture-collective/edda/issues/150) | Split the remaining named environmental hazards into bounded rules PRs | #144, #147 | — |
| [#182](https://github.com/subculture-collective/edda/issues/182) | Split unimplemented monster mechanic keys into exact fixture PRs | #151, #152, #153, #154, #155, #156, #157, #158, #159, #160, #161, #162, #163, #164, #165, #166, #167, #168, #169, #170, #171, #172, #173, #174, #175, #176, #177, #178, #184 | — |
| [#151](https://github.com/subculture-collective/edda/issues/151) | Catalogue SRD stat blocks 01: Aboleth to Air Elemental | #143 | 12 |
| [#152](https://github.com/subculture-collective/edda/issues/152) | Catalogue SRD stat blocks 02: Allosaurus to Animals | #143 | 12 |
| [#153](https://github.com/subculture-collective/edda/issues/153) | Catalogue SRD stat blocks 03: Animated Armor to Axe Beak | #143 | 12 |
| [#154](https://github.com/subculture-collective/edda/issues/154) | Catalogue SRD stat blocks 04: Azer Sentinel to Berserker | #143 | 12 |
| [#155](https://github.com/subculture-collective/edda/issues/155) | Catalogue SRD stat blocks 05: Black Bear to Bugbear Stalker | #143 | 12 |
| [#156](https://github.com/subculture-collective/edda/issues/156) | Catalogue SRD stat blocks 06: Bugbear Warrior to Cockatrice | #143 | 12 |
| [#157](https://github.com/subculture-collective/edda/issues/157) | Catalogue SRD stat blocks 07: Commoner to Deva | #143 | 12 |
| [#158](https://github.com/subculture-collective/edda/issues/158) | Catalogue SRD stat blocks 08: Dire Wolf to Earth Elemental | #143 | 12 |
| [#159](https://github.com/subculture-collective/edda/issues/159) | Catalogue SRD stat blocks 09: Efreeti to Frost Giant | #143 | 12 |
| [#160](https://github.com/subculture-collective/edda/issues/160) | Catalogue SRD stat blocks 10: Gargoyle to Giant Crab | #143 | 12 |
| [#161](https://github.com/subculture-collective/edda/issues/161) | Catalogue SRD stat blocks 11: Giant Crocodile to Giant Scorpion | #143 | 12 |
| [#162](https://github.com/subculture-collective/edda/issues/162) | Catalogue SRD stat blocks 12: Giant Seahorse to Gladiator | #143 | 12 |
| [#163](https://github.com/subculture-collective/edda/issues/163) | Catalogue SRD stat blocks 13: Gnoll Warrior to Griffon | #143 | 12 |
| [#164](https://github.com/subculture-collective/edda/issues/164) | Catalogue SRD stat blocks 14: Grimlock to Hippopotamus | #143 | 12 |
| [#165](https://github.com/subculture-collective/edda/issues/165) | Catalogue SRD stat blocks 15: Hobgoblin Captain to Invisible Stalker | #143 | 12 |
| [#166](https://github.com/subculture-collective/edda/issues/166) | Catalogue SRD stat blocks 16: Iron Golem to Mage | #143 | 12 |
| [#167](https://github.com/subculture-collective/edda/issues/167) | Catalogue SRD stat blocks 17: Magma Mephit to Minotaur of Baphomet | #143 | 12 |
| [#168](https://github.com/subculture-collective/edda/issues/168) | Catalogue SRD stat blocks 18: Mule to Oni | #143 | 12 |
| [#169](https://github.com/subculture-collective/edda/issues/169) | Catalogue SRD stat blocks 19: Otyugh to Plesiosaurus | #143 | 12 |
| [#170](https://github.com/subculture-collective/edda/issues/170) | Catalogue SRD stat blocks 20: Polar Bear to Red Dragon Wyrmling | #143 | 12 |
| [#171](https://github.com/subculture-collective/edda/issues/171) | Catalogue SRD stat blocks 21: Reef Shark to Scorpion | #143 | 12 |
| [#172](https://github.com/subculture-collective/edda/issues/172) | Catalogue SRD stat blocks 22: Scout to Sphinx of Lore | #143 | 12 |
| [#173](https://github.com/subculture-collective/edda/issues/173) | Catalogue SRD stat blocks 23: Sphinx of Valor to Succubus | #143 | 12 |
| [#174](https://github.com/subculture-collective/edda/issues/174) | Catalogue SRD stat blocks 24: Swarm of Bats to Treant | #143 | 12 |
| [#175](https://github.com/subculture-collective/edda/issues/175) | Catalogue SRD stat blocks 25: Triceratops to Vulture | #143 | 12 |
| [#176](https://github.com/subculture-collective/edda/issues/176) | Catalogue SRD stat blocks 26: Warhorse to White Dragon Wyrmling | #143 | 12 |
| [#177](https://github.com/subculture-collective/edda/issues/177) | Catalogue SRD stat blocks 27: Wight to Young Bronze Dragon | #143 | 12 |
| [#178](https://github.com/subculture-collective/edda/issues/178) | Catalogue SRD stat blocks 28: Young Copper Dragon to Zombie | #143 | 7 |
| [#184](https://github.com/subculture-collective/edda/issues/184) | Catalogue spell-created Animated Object, Otherworldly Steed, Giant Insect and Draconic Spirit | #143 | 4 |
| [#185](https://github.com/subculture-collective/edda/issues/185) | Split remaining spell-created creature control and scaling into fixture PRs | #80, #184 | — |
| [#183](https://github.com/subculture-collective/edda/issues/183) | Qualify complete monster and exploration coverage before advertising full support | #145, #146, #149, #150, #179, #180, #181, #182, #185 | — |

## Verification recorded for this decision

- The official PDF bytes matched the repository SHA-256 pin. Every one of the
  331 indexed names was found on its cited page; the four additional blocks were
  located in the spell section. The PDF and extracted prose are not committed.
- Offline manifest checks confirmed 335 unique IDs, one owner for every record,
  source pages within 1–364, 43 distinct issue units, at most 12 records per data
  PR, and an acyclic dependency graph. Catalogue and family statuses remain
  unsupported.
- Publication reads back issue titles/bodies, native parent/blocked-by links and
  Edda Project Size M. #47/#54 list every new issue; #80/#88 retain the overlap and
  claim boundaries. Transitive edges were removed from full qualification; its
  independent prerequisites remain visible.
- `go test -short -race ./internal/rules/srd521` and `git diff --check` pass for
  the register pointer/documentation change. This spike adds no runtime code or
  database mutation.
- Casting is separately delivered by merged #142 (`384c915` head, `f88d2ac`
  merge), with all five hosted checks green. Its hosted frontend job passed 60
  interaction tests and all 82 Chromium journeys. Local changed-file lint and
  all eight action-preview Chromium cases passed. Full local lint retains 22
  errors in untouched files; local full-browser attempts hit intermittent login
  fixture failures, while the hosted full run passed.
