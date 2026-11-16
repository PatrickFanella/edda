# Remaining Edda roadmap and development plan

Snapshot: November 16, 2026 UTC (November 15 in America/Chicago).
Source baseline: `18dc2053ecfaeeafba13a4f619dc2eea207fe551`.
Tracker: [roadmap #47](https://github.com/subculture-collective/edda/issues/47).
The [issue snapshot](remaining-issues.json) records every open issue, its native
parent, blocker states, delivery group and dependency wave at capture time.
Re-read GitHub before starting work; this document is a dated execution plan.

## Goal and delivery order

First qualify the playable slice already implemented, including a visible phone
keyboard and the required human review. Then expand useful gameplay through
small, server-authoritative actions while building complete source catalogues
and resolving the remaining mechanics. Full SRD support requires independent
character, spell, item and world qualification.

Keep story-led play, the existing Art Deco presentation and choice-based setup.
Gameplay fits one default desktop or phone viewport; history and details scroll
internally. Accessible enlarged text may reflow. Every action previews without
mutation and commits only after explicit confirmation. Preserve legacy campaign
reads, player-private information and the shared web/TUI engine.

The snapshot contains **200 open issues: 11 trackers and 189 work issues**.
Tracking issues are coordination and exit checks; they are not additional PRs.

| Delivery group | Existing work issues | Exit |
| --- | ---: | --- |
| A. Qualify the current playable slice | 3 | #46, #78 and #89 satisfy their device/human and functional criteria |
| B. Add safe contracts and useful gameplay slices | 19 | Four validated contracts and 15 bounded runtime slices, each with its own evidence |
| C. Complete source catalogues | 132 | Every assigned record, variant, nested option and required mechanic key has a validated owner |
| D. Resolve remaining mechanics | 31 investigations, plus their future implementation children | Every unsupported source key has a bounded implementation owner and passing evidence |
| E. Qualify complete coverage | 4 | #183, #279, #280 and #337 pass; #54's full exit criteria hold |

These are delivery groups, not five serial barriers. Start the device gate and
safe contracts now. Start the ready effect investigations early to expose shared
mechanic decisions. Import each catalogue after its own contract, alongside
runtime work. Run a class investigation as soon as its exact source batches
merge. Native blocked-by links determine eligibility; the appendix gives the
actual dependency waves.

## A. Finish current-slice qualification first

| Order | Issue | Remaining deliverable |
| --- | --- | --- |
| 1 | [#46](https://github.com/subculture-collective/edda/issues/46) | Review the baseline findings and record a suitable phone session with a visible normal software keyboard, device/OS/browser, orientation, text size, failures and reviewer/date |
| 2 | [#78](https://github.com/subculture-collective/edda/issues/78) | Join current viewport regression results with the actual touch, keyboard, browser-chrome, focus and accessibility evidence; fix demonstrated defects |
| 3 | [#89](https://github.com/subculture-collective/edda/issues/89) | Qualify creation, sheet/text actions, once-only costs, combat, retry/reconnect, chapter reading/export and legacy/TUI reads on one exact head |

Existing evidence is in [the #46 report](playtest/issue-46/README.md) and
[the #78 report](playtest/issue-78/README.md). The helper IME did not display a
keyboard in the first investigation. Later device attempts reported an occupied
session. Neither result satisfies keyboard qualification. Do not reuse or close
another session to complete the checkpoint.

At capture time [PR #343](https://github.com/subculture-collective/edda/pull/343),
head `076562ce01303de4b05d97c80ea1b632029f7637`, is the existing qualification
candidate. Reuse its work. Its live draft flag is false, although its description
still says to keep it draft; neither the flag nor a merge establishes human
approval. Check its current head, checks and evidence before continuing.
[Issue #342](https://github.com/subculture-collective/edda/issues/342) is closed
through merged [PR #344](https://github.com/subculture-collective/edda/pull/344).
The exact supported Athletics text request prepares a preview; general SRD
free-text adjudication remains unavailable. Qualify that declared boundary.

Run the default viewport matrix at 1366×768, 1440×900, 1920×1080, 768×1024,
360×800 and 390×844. Cover long narrative/choices, details, pending/error states,
combat, saves/export, internal scroll ownership and retained draft/reading/focus.
Record visible keyboard open/close, portrait/landscape, browser chrome, actual
assistive technology and final player-experience review separately. Desktop
emulation or synthetic `visualViewport` resizing cannot discharge those gates.

Close #49 only after its remaining children and exit criteria pass. #48, #50 and
#51 have no open native work children, but remain open trackers at capture time;
review their recorded exit evidence during #89 before closing them. Repair stale
checkboxes from issue state without treating that bookkeeping as new verification.
The completed suggestion/chapter epics #52/#53 remain regression surfaces.

## B. Add contracts and demonstrate practical gameplay

Each row is one existing PR-sized issue. Blockers below are **open native
blockers at capture time**; satisfied historical prerequisites remain in the
snapshot. Contract PRs validate metadata and fail closed. A source record alone
cannot make content selectable or playable.

| Issue | Deliverable | Open prerequisites |
| --- | --- | --- |
| [#143](https://github.com/subculture-collective/edda/issues/143) | Monster schema with the Bandit record and unsupported-key refusal | None |
| [#194](https://github.com/subculture-collective/edda/issues/194) | Spell schema with an Acid Arrow record and exact component/target/effect keys | None |
| [#195](https://github.com/subculture-collective/edda/issues/195) | Item-family schema with explicit variants and a Potion of Healing record | None |
| [#284](https://github.com/subculture-collective/edda/issues/284) | Character-choice schema with one Fighter record and source/prerequisite links | None |
| [#338](https://github.com/subculture-collective/edda/issues/338) | Confirm the existing Fighter2's one Action Surge extra Action, excluding Magic | None |
| [#339](https://github.com/subculture-collective/edda/issues/339) | Offer Tactical Mind after one failed ability check, preserving its original roll and hidden DC | None |
| [#144](https://github.com/subculture-collective/edda/issues/144) | Confirm one hour of travel on an explicitly known route with a selected pace | None |
| [#145](https://github.com/subculture-collective/edda/issues/145) | Derive visibility from declared light, obstruction and one Darkvision fixture | None |
| [#146](https://github.com/subculture-collective/edda/issues/146) | Open one known ordinary unlocked door with durable state and spatial invalidation | None |
| [#147](https://github.com/subculture-collective/edda/issues/147) | Resolve one declared fall with capped damage and Prone | None |
| [#148](https://github.com/subculture-collective/edda/issues/148) | Resolve one Bandit weapon attack through the durable encounter path | #143 |
| [#149](https://github.com/subculture-collective/edda/issues/149) | Calculate encounter XP difficulty budgets without granting rewards | #143 |
| [#179](https://github.com/subculture-collective/edda/issues/179) | Resolve the Wolf's Pack Tactics and size-gated Bite Prone | #148 |
| [#180](https://github.com/subculture-collective/edda/issues/180) | Resolve Brown Bear Multiattack as one atomic action | #148 |
| [#263](https://github.com/subculture-collective/edda/issues/263) | Cast level-one Cure Wounds on one willing conscious adjacent ally | #194 |
| [#264](https://github.com/subculture-collective/edda/issues/264) | Cast Light on one known carried object with durable expiry | #194, #145 |
| [#265](https://github.com/subculture-collective/edda/issues/265) | Use the recorded Magic Initiate Detect Magic use with bounded sensing/concentration | #194 |
| [#266](https://github.com/subculture-collective/edda/issues/266) | Administer one basic Potion of Healing to a willing conscious ally | #195, #263 |
| [#267](https://github.com/subculture-collective/edda/issues/267) | Equip/remove owned chain mail or shield, deriving AC and don/doff time | #195 |

The default next queue is #46 alongside #143, #194, #195 and #284; then select
one runtime slice from each available domain. Prefer #338/#339 for the already
created Fighter2, #148 for source-backed encounters, and #263 → #266 for ally
support. #145 unlocks #264. These preferences add no artificial dependency edges.
Keep one implementation PR active per shared mutation area; integrate against
fresh main before another issue changes that transaction or projection.

## C. Import complete source catalogues in bounded batches

| Domain | Contract | Data issues | Count | Completion boundary |
| --- | --- | --- | ---: | --- |
| Monsters and spell-created creatures | #143 | #151–#178, #184 | 29 | 331 indexed records plus four spell-created blocks, retaining scaling/control keys |
| Spells | #194 | #196–#224 | 29 | 339 source spell families with exact components, targeting, duration and effect keys |
| Magic items | #195 | #225–#246 | 22 | 258 named families, with every listed variant explicit |
| Mundane equipment | #195 | #247–#262 | 16 | 156 source entries across weapons, armor, tools and adventuring gear |
| Characters and advancement | #284 | #285–#320 | 36 | 344 planning identities plus required nested choices, all class level rows and source sections |

The 1,432 planning records/families describe source coverage, not playable
mechanics. The totals are not a uniform estimate of effort: a class batch
includes all 20 progression rows, and item families can contain several variants.
Source inventories remain authoritative:

- [Character and advancement decision](investigations/issue-74-character-expansion.md)
  and [inventory](investigations/issue-74-character-inventory.json).
- [Spell/item/effect decision](investigations/issue-80-spell-item-expansion.md)
  and [inventory](investigations/issue-80-spell-item-inventory.json).
- [World decision](investigations/issue-82-world-expansion.md)
  and [inventory](investigations/issue-82-world-inventory.json).

Prioritize batches that unlock the next mechanic decision: #285 advancement,
#290 Fighter and #312 Champion for #325; #205 for familiar planning #181;
#196/#205/#207/#221 and #184 for the spell-created-creature decision #185.
Other batches can merge independently after their contract. Keep their record
ranges disjoint and reuse the contract fixture instead of importing a duplicate.

Each import preserves source IDs/pages, pin and CC-BY-4.0 attribution/adaptation
notice; schema fixtures check identity completeness, nested choices/variants and
unknown mechanic refusal. Data changes cannot activate actions or change player
resources. Split an oversized batch into exact follow-ups before widening it.

## D. Resolve mechanics with explicit implementation follow-ups

Start ready investigations #271 (effect lifetime/stacking) and #273 (casting
routes/components) early because other domains reuse their decisions. The other
ready investigations #268–#270, #272 and #274–#276 need no open prerequisites.
Their eligibility is independent of the catalogue imports. Limit discovery to
decisions that the next runtime slices can use, while retaining every required
source scope in the backlog.

| Investigation set | Issues | Required output |
| --- | --- | --- |
| Sensory, mental and bodily conditions | #268–#270 | Named fixture PRs for all listed conditions; shared effect ownership and disclosure rules |
| Effect lifetime, recovery and casting | #271–#273 | Exact stacking/triggers, concentration, rest/recharge, rituals, components and item/free-use routing children |
| Equipment economy and production | #274–#276 | Bounded mount/vehicle/cargo, prices/services/lifestyle and crafting children |
| World hazards and creatures | #150, #181, #182, #185 | Named hazard children, familiar/control slices, every residual monster mechanic and spell-created creature lifecycle |
| Residual spell and item mechanics | #277, #278 | Source-key-to-owner matrix after every corresponding data import, deduplicating shared effects and summons |
| Class mechanics | #321–#332 | One class per investigation: exact feature/choice/level-transition implementation children, including its subclass/options |
| Shared character choices | #333–#336 | Origin/creation, all feat effects, multiclass and shared advancement/higher-level-start children |

An investigation is one working session and one recorded decision PR. It must
identify each unresolved source key and publish bounded children before closing.
Every child names the source scope, user action, target, cost, timing, result,
negative cases, affected modules, exclusions and verification. Add the native
parent and real blockers, then read them back and check the graph for cycles.
Terminal implementation children must block the corresponding full-coverage
gate. A comment listing future work is insufficient.

Use the existing spell/item/effect/world owner when a class feature needs it.
Avoid a second concentration, recharge or summon-control implementation. Shared
mechanics must have one authoritative rule path; web and TUI consume that path.
Unknown timing, target or sensory facts produce an unavailable action or an
explicit clarification. The model cannot supply missing mechanical authority.

## E. Qualify full coverage only after implementation evidence exists

| Gate | Prerequisite boundary | Passing evidence |
| --- | --- | --- |
| [#183](https://github.com/subculture-collective/edda/issues/183) world | Catalogue, runtime fixtures, hazards, familiars, monster keys and summons, including all splitter children | Complete monster/exploration source coverage and durable mechanic fixtures |
| [#279](https://github.com/subculture-collective/edda/issues/279) spells | All spell records/keys, condition/lifecycle/casting children, named spell actions and creature-control children | Every declared spell/effect has supported resolver and persistence evidence |
| [#280](https://github.com/subculture-collective/edda/issues/280) items | Item variants/keys, equipment/economy/crafting children, potion/equip fixtures and #279 | Every declared item family/variant has source-complete executable evidence |
| [#337](https://github.com/subculture-collective/edda/issues/337) characters | All class/shared-choice investigations and terminal implementations, plus #338/#339 | Complete creation, 12 classes/subclasses, levels 1–20, options, feats and multiclass fixtures |

The existing known longest dependency chain has five open issues:
#194 → any spell data batch (for example #196) → #277 → #279 → #280.
The character chain #284 → #285 → #321 → #337 and the world chain
#143 → #148 → #179 → #183 have four. These are lower bounds: the 31 mechanics
investigations will add implementation prerequisites. No credible end date or
Full Rules completion estimate follows from the current issue count.

Keep [actual coverage](srd-5.2.1-coverage.md) and the
[conformance register](srd-5.2.1-conformance.md) aligned with evidence. Closed
issues, imported catalogues and changed status flags cannot open the Full Rules
gate. #54 closes only after all required domains and its campaign/viewport/legacy
criteria qualify, or the owner explicitly narrows the target and claims.

## Implementation and verification procedure

1. Read the issue and current native blockers, inspect the exact worktree and
   main, and reproduce the requested behavior with deterministic fixtures.
2. Put source-derived calculations and content validation in
   `internal/rules/srd521`; reuse `internal/game`, `internal/engine` and
   `internal/combat` for authoritative preparation/confirmation and encounters.
   Keep the existing private-to-player-safe projection boundary.
3. For mutations, extend the existing atomic action/history boundary. Confirm
   ownership and revision; inject dice; commit costs, effects and receipt once.
   Preview, decline, stale input and unauthorized requests change nothing.
4. Add the caller and availability projection in the same PR where applicable.
   Keep web controls usable within the viewport and explicitly represent any
   unsupported TUI action. Avoid speculative adapters or duplicate engines.
5. Prove rollback, retry, concurrent confirmation and reload with an isolated
   PostgreSQL fixture. Use model doubles for mechanics and browser contracts;
   paid inference and production credentials are unnecessary for these checks.
6. Run the checks relevant to the change below. Review the diff and exclusions,
   push a focused PR with `Closes #n`, and require green checks on its exact head
   plus a clean merge state. Keep incomplete human/device work explicitly open.

| Change | Local verification |
| --- | --- |
| Go rules/engine | `rtk go test -short -race ./...` and `rtk go vet ./...` |
| Persistence/handler mutation | `rtk go test -tags integration -race -run Integration -count=1 -timeout 10m ./...` with Docker available for disposable pgvector containers |
| Web caller or presentation | `rtk proxy pnpm --dir frontend test`, `rtk proxy pnpm --dir frontend test:browser`, `rtk proxy pnpm --dir frontend build`, changed-file lint and comparison against current lint baseline |
| Shared client behavior | `rtk go test -short ./tui/...` plus an explicit legacy-read/TUI fixture for the touched boundary |
| Source-only catalogue | Offline pin/schema/identity/variant/key-completeness fixtures for that exact batch; no inference or production data |
| Investigation | Recorded decision, published bounded children and readback of parent/blocker links, terminal gate coverage and acyclicity |
| Device/human gate | Suitable visible keyboard and accessibility session, exact build/device/browser, screenshots or recording and reviewer/date |

Choose a free `EDDA_FIXTURE_PORT` when the default is occupied; Playwright owns
its fixture server. Preserve unrelated processes. Check disk space before large
builds/downloads. Record existing broad lint failures separately and fix new
findings in changed files without absorbing unrelated cleanup.

Hosted `Mirror CI gate` joins Go tests, PostgreSQL integration, frontend image
build and frontend behavior. Passing local checks is separate from hosted CI.
The main-branch deployment workflow publishes the frontend; it does not prove
that a changed API is running. Deployment requires separate authorization,
exact backend/frontend artifact revisions, migration/rollback preparation and
live URL/service and relevant interaction checks.

## Capacity, readiness and refresh

The ready frontier contains 20 work issues: #46, #143–#147, #194, #195,
#268–#276, #284, #338 and #339. Choose by the next user-visible outcome and shared
code ownership. The recommendations above are priorities within this plan;
they do not apply automation labels, assign agents or start implementation.

The 132 source imports are known data PRs. The 19 contract/runtime slices are
known bounded changes. The 31 investigations add an unknown number of mechanic
PRs, and the four full gates may reveal additional defects. Retain existing
issue sizes as review estimates; do not convert them into delivery dates. Use
recent merged-PR throughput and the children actually published after discovery
to make a later forecast. Re-estimate at each domain's first completed runtime
slice and after its residual-key investigation.

After each merged issue, refresh checkboxes and the native ready frontier. After
each investigation, add its implementation children to the appropriate epic and
qualification gate before marking the decision complete. Close an epic only
after its children and explicit exit criteria pass. #47 remains the single
roadmap. Leave production rollout and out-of-scope product decisions separate.

To refresh the source list:

```sh
rtk proxy gh issue list --repo subculture-collective/edda --state open --limit 1000 \
  --json number,title,body,labels,url,updatedAt
rtk proxy gh pr list --repo subculture-collective/edda --state open \
  --json number,title,url,headRefOid,isDraft,statusCheckRollup
```

Fetch native `parent`, `blockedBy` and `subIssues` through GitHub GraphQL with
pagination; prose references are not dependency edges. Exclude
`Planning/Roadmap` and `Planning/Epic` issues when calculating work waves. Recheck
closed blockers too. At capture, #342 is a satisfied native blocker of #89,
#181/#185 have additional spell prerequisites, and #337 includes #338/#339.
The snapshot retains these links even where the original issue text is older.

## Complete remaining issue ledger

Every open issue appears exactly once below. **Wave 0** is the ready frontier;
later waves are dependency depth among open work issues. Items in one wave
have no blockers in that wave. Groups A–E express outcomes; waves express the
actual merge order. New children will change the waves and counts.

### Tracking issues

| Issue | Role | Parent | Open direct children |
| --- | --- | --- | --- |
| [#47](https://github.com/subculture-collective/edda/issues/47) [Roadmap] Story-led Edda gameplay with SRD 5.2.1 | Roadmap | — | 5/7 |
| [#48](https://github.com/subculture-collective/edda/issues/48) Make campaign actions private, durable and recoverable | Epic | [#47](https://github.com/subculture-collective/edda/issues/47) | 0/12 |
| [#49](https://github.com/subculture-collective/edda/issues/49) Keep story-led gameplay usable within one screen | Epic | [#47](https://github.com/subculture-collective/edda/issues/47) | 2/10 |
| [#50](https://github.com/subculture-collective/edda/issues/50) Play supported SRD actions from the interactive sheet | Epic | [#47](https://github.com/subculture-collective/edda/issues/47) | 0/13 |
| [#51](https://github.com/subculture-collective/edda/issues/51) Create and advance valid rules-versioned characters | Epic | [#47](https://github.com/subculture-collective/edda/issues/47) | 0/8 |
| [#54](https://github.com/subculture-collective/edda/issues/54) Complete and qualify the declared SRD 5.2.1 coverage | Epic | [#47](https://github.com/subculture-collective/edda/issues/47) | 49/53 |
| [#191](https://github.com/subculture-collective/edda/issues/191) Catalogue and qualify the remaining SRD spells | Epic | [#54](https://github.com/subculture-collective/edda/issues/54) | 35/35 |
| [#192](https://github.com/subculture-collective/edda/issues/192) Catalogue and qualify the remaining SRD equipment and magic items | Epic | [#54](https://github.com/subculture-collective/edda/issues/54) | 43/43 |
| [#193](https://github.com/subculture-collective/edda/issues/193) Resolve remaining spell and item effect interactions | Epic | [#54](https://github.com/subculture-collective/edda/issues/54) | 9/9 |
| [#282](https://github.com/subculture-collective/edda/issues/282) Catalogue remaining SRD character source requirements | Epic | [#54](https://github.com/subculture-collective/edda/issues/54) | 37/37 |
| [#283](https://github.com/subculture-collective/edda/issues/283) Resolve and qualify all SRD character mechanics | Epic | [#54](https://github.com/subculture-collective/edda/issues/54) | 19/19 |

### Dependency wave 0: 20 work issues

| Issue | Group / kind | Parent | Open blockers | Checkpoint |
| --- | --- | --- | --- | --- |
| [#46](https://github.com/subculture-collective/edda/issues/46) Assess narrow-screen campaign play before redesigning the HUD | A / investigation | [#49](https://github.com/subculture-collective/edda/issues/49) | None | human checkpoint |
| [#143](https://github.com/subculture-collective/edda/issues/143) Validate versioned SRD monster records and fail closed on unsupported mechanics | B / contract/runtime | [#54](https://github.com/subculture-collective/edda/issues/54) | None | agent |
| [#144](https://github.com/subculture-collective/edda/issues/144) Prepare and commit a one-hour SRD travel segment with explicit pace | B / contract/runtime | [#54](https://github.com/subculture-collective/edda/issues/54) | None | agent |
| [#145](https://github.com/subculture-collective/edda/issues/145) Adjudicate declared light and Darkvision for narrative target visibility | B / contract/runtime | [#54](https://github.com/subculture-collective/edda/issues/54) | None | agent |
| [#146](https://github.com/subculture-collective/edda/issues/146) Confirm opening an ordinary unlocked door through narrative-space facts | B / contract/runtime | [#54](https://github.com/subculture-collective/edda/issues/54) | None | agent |
| [#147](https://github.com/subculture-collective/edda/issues/147) Resolve one declared fall with inspectable damage and Prone state | B / contract/runtime | [#54](https://github.com/subculture-collective/edda/issues/54) | None | agent |
| [#194](https://github.com/subculture-collective/edda/issues/194) Validate versioned spell records and fail closed on unsupported effects | B / contract/runtime | [#191](https://github.com/subculture-collective/edda/issues/191) | None | agent |
| [#195](https://github.com/subculture-collective/edda/issues/195) Validate versioned item families and explicit variant identities | B / contract/runtime | [#192](https://github.com/subculture-collective/edda/issues/192) | None | agent |
| [#268](https://github.com/subculture-collective/edda/issues/268) Split Blinded, Deafened and Invisible effects into fixture PRs | D / investigation | [#193](https://github.com/subculture-collective/edda/issues/193) | None | agent |
| [#269](https://github.com/subculture-collective/edda/issues/269) Split Charmed, Frightened and incapacitating effects into fixture PRs | D / investigation | [#193](https://github.com/subculture-collective/edda/issues/193) | None | agent |
| [#270](https://github.com/subculture-collective/edda/issues/270) Split remaining bodily condition effects into fixture PRs | D / investigation | [#193](https://github.com/subculture-collective/edda/issues/193) | None | agent |
| [#271](https://github.com/subculture-collective/edda/issues/271) Split concentration, duration and stacking interactions | D / investigation | [#193](https://github.com/subculture-collective/edda/issues/193) | None | agent |
| [#272](https://github.com/subculture-collective/edda/issues/272) Split interrupted rest, recovery and recharge exceptions | D / investigation | [#193](https://github.com/subculture-collective/edda/issues/193) | None | agent |
| [#273](https://github.com/subculture-collective/edda/issues/273) Split ritual casting and costly or consumed components | D / investigation | [#193](https://github.com/subculture-collective/edda/issues/193) | None | agent |
| [#274](https://github.com/subculture-collective/edda/issues/274) Split mount, vehicle and cargo rules into bounded fixtures | D / investigation | [#193](https://github.com/subculture-collective/edda/issues/193) | None | agent |
| [#275](https://github.com/subculture-collective/edda/issues/275) Split equipment prices, services and lifestyle consequences | D / investigation | [#193](https://github.com/subculture-collective/edda/issues/193) | None | agent |
| [#276](https://github.com/subculture-collective/edda/issues/276) Split nonmagical crafting, potion brewing and scroll scribing | D / investigation | [#193](https://github.com/subculture-collective/edda/issues/193) | None | agent |
| [#284](https://github.com/subculture-collective/edda/issues/284) Validate source-linked character choices and fail closed on unsupported features | B / contract/runtime | [#282](https://github.com/subculture-collective/edda/issues/282) | None | agent |
| [#338](https://github.com/subculture-collective/edda/issues/338) Confirm one extra Fighter turn action from Action Surge | B / contract/runtime | [#283](https://github.com/subculture-collective/edda/issues/283) | None | agent |
| [#339](https://github.com/subculture-collective/edda/issues/339) Confirm Tactical Mind for one failed Fighter ability check | B / contract/runtime | [#283](https://github.com/subculture-collective/edda/issues/283) | None | agent |

### Dependency wave 1: 140 work issues

| Issue | Group / kind | Parent | Open blockers | Checkpoint |
| --- | --- | --- | --- | --- |
| [#78](https://github.com/subculture-collective/edda/issues/78) Assert gameplay viewport fit, scroll ownership and control reachability | A / contract/runtime | [#49](https://github.com/subculture-collective/edda/issues/49) | [#46](https://github.com/subculture-collective/edda/issues/46) | human checkpoint |
| [#148](https://github.com/subculture-collective/edda/issues/148) Run a Bandit attack from its pinned stat block through durable encounters | B / contract/runtime | [#54](https://github.com/subculture-collective/edda/issues/54) | [#143](https://github.com/subculture-collective/edda/issues/143) | agent |
| [#149](https://github.com/subculture-collective/edda/issues/149) Calculate SRD encounter XP budgets without inventing rewards | B / contract/runtime | [#54](https://github.com/subculture-collective/edda/issues/54) | [#143](https://github.com/subculture-collective/edda/issues/143) | agent |
| [#150](https://github.com/subculture-collective/edda/issues/150) Split the remaining named environmental hazards into bounded rules PRs | D / investigation | [#54](https://github.com/subculture-collective/edda/issues/54) | [#144](https://github.com/subculture-collective/edda/issues/144), [#147](https://github.com/subculture-collective/edda/issues/147) | agent |
| [#151](https://github.com/subculture-collective/edda/issues/151) Catalogue SRD stat blocks 01: Aboleth to Air Elemental | C / data | [#54](https://github.com/subculture-collective/edda/issues/54) | [#143](https://github.com/subculture-collective/edda/issues/143) | agent |
| [#152](https://github.com/subculture-collective/edda/issues/152) Catalogue SRD stat blocks 02: Allosaurus to Animals | C / data | [#54](https://github.com/subculture-collective/edda/issues/54) | [#143](https://github.com/subculture-collective/edda/issues/143) | agent |
| [#153](https://github.com/subculture-collective/edda/issues/153) Catalogue SRD stat blocks 03: Animated Armor to Axe Beak | C / data | [#54](https://github.com/subculture-collective/edda/issues/54) | [#143](https://github.com/subculture-collective/edda/issues/143) | agent |
| [#154](https://github.com/subculture-collective/edda/issues/154) Catalogue SRD stat blocks 04: Azer Sentinel to Berserker | C / data | [#54](https://github.com/subculture-collective/edda/issues/54) | [#143](https://github.com/subculture-collective/edda/issues/143) | agent |
| [#155](https://github.com/subculture-collective/edda/issues/155) Catalogue SRD stat blocks 05: Black Bear to Bugbear Stalker | C / data | [#54](https://github.com/subculture-collective/edda/issues/54) | [#143](https://github.com/subculture-collective/edda/issues/143) | agent |
| [#156](https://github.com/subculture-collective/edda/issues/156) Catalogue SRD stat blocks 06: Bugbear Warrior to Cockatrice | C / data | [#54](https://github.com/subculture-collective/edda/issues/54) | [#143](https://github.com/subculture-collective/edda/issues/143) | agent |
| [#157](https://github.com/subculture-collective/edda/issues/157) Catalogue SRD stat blocks 07: Commoner to Deva | C / data | [#54](https://github.com/subculture-collective/edda/issues/54) | [#143](https://github.com/subculture-collective/edda/issues/143) | agent |
| [#158](https://github.com/subculture-collective/edda/issues/158) Catalogue SRD stat blocks 08: Dire Wolf to Earth Elemental | C / data | [#54](https://github.com/subculture-collective/edda/issues/54) | [#143](https://github.com/subculture-collective/edda/issues/143) | agent |
| [#159](https://github.com/subculture-collective/edda/issues/159) Catalogue SRD stat blocks 09: Efreeti to Frost Giant | C / data | [#54](https://github.com/subculture-collective/edda/issues/54) | [#143](https://github.com/subculture-collective/edda/issues/143) | agent |
| [#160](https://github.com/subculture-collective/edda/issues/160) Catalogue SRD stat blocks 10: Gargoyle to Giant Crab | C / data | [#54](https://github.com/subculture-collective/edda/issues/54) | [#143](https://github.com/subculture-collective/edda/issues/143) | agent |
| [#161](https://github.com/subculture-collective/edda/issues/161) Catalogue SRD stat blocks 11: Giant Crocodile to Giant Scorpion | C / data | [#54](https://github.com/subculture-collective/edda/issues/54) | [#143](https://github.com/subculture-collective/edda/issues/143) | agent |
| [#162](https://github.com/subculture-collective/edda/issues/162) Catalogue SRD stat blocks 12: Giant Seahorse to Gladiator | C / data | [#54](https://github.com/subculture-collective/edda/issues/54) | [#143](https://github.com/subculture-collective/edda/issues/143) | agent |
| [#163](https://github.com/subculture-collective/edda/issues/163) Catalogue SRD stat blocks 13: Gnoll Warrior to Griffon | C / data | [#54](https://github.com/subculture-collective/edda/issues/54) | [#143](https://github.com/subculture-collective/edda/issues/143) | agent |
| [#164](https://github.com/subculture-collective/edda/issues/164) Catalogue SRD stat blocks 14: Grimlock to Hippopotamus | C / data | [#54](https://github.com/subculture-collective/edda/issues/54) | [#143](https://github.com/subculture-collective/edda/issues/143) | agent |
| [#165](https://github.com/subculture-collective/edda/issues/165) Catalogue SRD stat blocks 15: Hobgoblin Captain to Invisible Stalker | C / data | [#54](https://github.com/subculture-collective/edda/issues/54) | [#143](https://github.com/subculture-collective/edda/issues/143) | agent |
| [#166](https://github.com/subculture-collective/edda/issues/166) Catalogue SRD stat blocks 16: Iron Golem to Mage | C / data | [#54](https://github.com/subculture-collective/edda/issues/54) | [#143](https://github.com/subculture-collective/edda/issues/143) | agent |
| [#167](https://github.com/subculture-collective/edda/issues/167) Catalogue SRD stat blocks 17: Magma Mephit to Minotaur of Baphomet | C / data | [#54](https://github.com/subculture-collective/edda/issues/54) | [#143](https://github.com/subculture-collective/edda/issues/143) | agent |
| [#168](https://github.com/subculture-collective/edda/issues/168) Catalogue SRD stat blocks 18: Mule to Oni | C / data | [#54](https://github.com/subculture-collective/edda/issues/54) | [#143](https://github.com/subculture-collective/edda/issues/143) | agent |
| [#169](https://github.com/subculture-collective/edda/issues/169) Catalogue SRD stat blocks 19: Otyugh to Plesiosaurus | C / data | [#54](https://github.com/subculture-collective/edda/issues/54) | [#143](https://github.com/subculture-collective/edda/issues/143) | agent |
| [#170](https://github.com/subculture-collective/edda/issues/170) Catalogue SRD stat blocks 20: Polar Bear to Red Dragon Wyrmling | C / data | [#54](https://github.com/subculture-collective/edda/issues/54) | [#143](https://github.com/subculture-collective/edda/issues/143) | agent |
| [#171](https://github.com/subculture-collective/edda/issues/171) Catalogue SRD stat blocks 21: Reef Shark to Scorpion | C / data | [#54](https://github.com/subculture-collective/edda/issues/54) | [#143](https://github.com/subculture-collective/edda/issues/143) | agent |
| [#172](https://github.com/subculture-collective/edda/issues/172) Catalogue SRD stat blocks 22: Scout to Sphinx of Lore | C / data | [#54](https://github.com/subculture-collective/edda/issues/54) | [#143](https://github.com/subculture-collective/edda/issues/143) | agent |
| [#173](https://github.com/subculture-collective/edda/issues/173) Catalogue SRD stat blocks 23: Sphinx of Valor to Succubus | C / data | [#54](https://github.com/subculture-collective/edda/issues/54) | [#143](https://github.com/subculture-collective/edda/issues/143) | agent |
| [#174](https://github.com/subculture-collective/edda/issues/174) Catalogue SRD stat blocks 24: Swarm of Bats to Treant | C / data | [#54](https://github.com/subculture-collective/edda/issues/54) | [#143](https://github.com/subculture-collective/edda/issues/143) | agent |
| [#175](https://github.com/subculture-collective/edda/issues/175) Catalogue SRD stat blocks 25: Triceratops to Vulture | C / data | [#54](https://github.com/subculture-collective/edda/issues/54) | [#143](https://github.com/subculture-collective/edda/issues/143) | agent |
| [#176](https://github.com/subculture-collective/edda/issues/176) Catalogue SRD stat blocks 26: Warhorse to White Dragon Wyrmling | C / data | [#54](https://github.com/subculture-collective/edda/issues/54) | [#143](https://github.com/subculture-collective/edda/issues/143) | agent |
| [#177](https://github.com/subculture-collective/edda/issues/177) Catalogue SRD stat blocks 27: Wight to Young Bronze Dragon | C / data | [#54](https://github.com/subculture-collective/edda/issues/54) | [#143](https://github.com/subculture-collective/edda/issues/143) | agent |
| [#178](https://github.com/subculture-collective/edda/issues/178) Catalogue SRD stat blocks 28: Young Copper Dragon to Zombie | C / data | [#54](https://github.com/subculture-collective/edda/issues/54) | [#143](https://github.com/subculture-collective/edda/issues/143) | agent |
| [#184](https://github.com/subculture-collective/edda/issues/184) Catalogue spell-created Animated Object, Otherworldly Steed, Giant Insect and Draconic Spirit | C / data | [#54](https://github.com/subculture-collective/edda/issues/54) | [#143](https://github.com/subculture-collective/edda/issues/143) | agent |
| [#196](https://github.com/subculture-collective/edda/issues/196) Catalogue SRD spells 01: Acid Arrow to Antimagic Field | C / data | [#191](https://github.com/subculture-collective/edda/issues/191) | [#194](https://github.com/subculture-collective/edda/issues/194) | agent |
| [#197](https://github.com/subculture-collective/edda/issues/197) Catalogue SRD spells 02: Antipathy/Sympathy to Banishment | C / data | [#191](https://github.com/subculture-collective/edda/issues/191) | [#194](https://github.com/subculture-collective/edda/issues/194) | agent |
| [#198](https://github.com/subculture-collective/edda/issues/198) Catalogue SRD spells 03: Barkskin to Burning Hands | C / data | [#191](https://github.com/subculture-collective/edda/issues/191) | [#194](https://github.com/subculture-collective/edda/issues/194) | agent |
| [#199](https://github.com/subculture-collective/edda/issues/199) Catalogue SRD spells 04: Call Lightning to Color Spray | C / data | [#191](https://github.com/subculture-collective/edda/issues/191) | [#194](https://github.com/subculture-collective/edda/issues/194) | agent |
| [#200](https://github.com/subculture-collective/edda/issues/200) Catalogue SRD spells 05: Command to Conjure Minor Elementals | C / data | [#191](https://github.com/subculture-collective/edda/issues/191) | [#194](https://github.com/subculture-collective/edda/issues/194) | agent |
| [#201](https://github.com/subculture-collective/edda/issues/201) Catalogue SRD spells 06: Conjure Woodland Beings to Creation | C / data | [#191](https://github.com/subculture-collective/edda/issues/191) | [#194](https://github.com/subculture-collective/edda/issues/194) | agent |
| [#202](https://github.com/subculture-collective/edda/issues/202) Catalogue SRD spells 07: Cure Wounds to Detect Thoughts | C / data | [#191](https://github.com/subculture-collective/edda/issues/191) | [#194](https://github.com/subculture-collective/edda/issues/194) | agent |
| [#203](https://github.com/subculture-collective/edda/issues/203) Catalogue SRD spells 08: Dimension Door to Dominate Monster | C / data | [#191](https://github.com/subculture-collective/edda/issues/191) | [#194](https://github.com/subculture-collective/edda/issues/194) | agent |
| [#204](https://github.com/subculture-collective/edda/issues/204) Catalogue SRD spells 09: Dominate Person to Enthrall | C / data | [#191](https://github.com/subculture-collective/edda/issues/191) | [#194](https://github.com/subculture-collective/edda/issues/194) | agent |
| [#205](https://github.com/subculture-collective/edda/issues/205) Catalogue SRD spells 10: Etherealness to Find Traps | C / data | [#191](https://github.com/subculture-collective/edda/issues/191) | [#194](https://github.com/subculture-collective/edda/issues/194) | agent |
| [#206](https://github.com/subculture-collective/edda/issues/206) Catalogue SRD spells 11: Find the Path to Fly | C / data | [#191](https://github.com/subculture-collective/edda/issues/191) | [#194](https://github.com/subculture-collective/edda/issues/194) | agent |
| [#207](https://github.com/subculture-collective/edda/issues/207) Catalogue SRD spells 12: Fog Cloud to Glibness | C / data | [#191](https://github.com/subculture-collective/edda/issues/191) | [#194](https://github.com/subculture-collective/edda/issues/194) | agent |
| [#208](https://github.com/subculture-collective/edda/issues/208) Catalogue SRD spells 13: Globe of Invulnerability to Hallow | C / data | [#191](https://github.com/subculture-collective/edda/issues/191) | [#194](https://github.com/subculture-collective/edda/issues/194) | agent |
| [#209](https://github.com/subculture-collective/edda/issues/209) Catalogue SRD spells 14: Hallucinatory Terrain to Hold Monster | C / data | [#191](https://github.com/subculture-collective/edda/issues/191) | [#194](https://github.com/subculture-collective/edda/issues/194) | agent |
| [#210](https://github.com/subculture-collective/edda/issues/210) Catalogue SRD spells 15: Hold Person to Insect Plague | C / data | [#191](https://github.com/subculture-collective/edda/issues/191) | [#194](https://github.com/subculture-collective/edda/issues/194) | agent |
| [#211](https://github.com/subculture-collective/edda/issues/211) Catalogue SRD spells 16: Instant Summons to Locate Creature | C / data | [#191](https://github.com/subculture-collective/edda/issues/191) | [#194](https://github.com/subculture-collective/edda/issues/194) | agent |
| [#212](https://github.com/subculture-collective/edda/issues/212) Catalogue SRD spells 17: Locate Object to Mass Cure Wounds | C / data | [#191](https://github.com/subculture-collective/edda/issues/191) | [#194](https://github.com/subculture-collective/edda/issues/194) | agent |
| [#213](https://github.com/subculture-collective/edda/issues/213) Catalogue SRD spells 18: Mass Heal to Mirage Arcane | C / data | [#191](https://github.com/subculture-collective/edda/issues/191) | [#194](https://github.com/subculture-collective/edda/issues/194) | agent |
| [#214](https://github.com/subculture-collective/edda/issues/214) Catalogue SRD spells 19: Mirror Image to Phantom Steed | C / data | [#191](https://github.com/subculture-collective/edda/issues/191) | [#194](https://github.com/subculture-collective/edda/issues/194) | agent |
| [#215](https://github.com/subculture-collective/edda/issues/215) Catalogue SRD spells 20: Planar Ally to Prismatic Spray | C / data | [#191](https://github.com/subculture-collective/edda/issues/191) | [#194](https://github.com/subculture-collective/edda/issues/194) | agent |
| [#216](https://github.com/subculture-collective/edda/issues/216) Catalogue SRD spells 21: Prismatic Wall to Ray of Frost | C / data | [#191](https://github.com/subculture-collective/edda/issues/191) | [#194](https://github.com/subculture-collective/edda/issues/194) | agent |
| [#217](https://github.com/subculture-collective/edda/issues/217) Catalogue SRD spells 22: Ray of Sickness to Sanctuary | C / data | [#191](https://github.com/subculture-collective/edda/issues/191) | [#194](https://github.com/subculture-collective/edda/issues/194) | agent |
| [#218](https://github.com/subculture-collective/edda/issues/218) Catalogue SRD spells 23: Scorching Ray to Shield of Faith | C / data | [#191](https://github.com/subculture-collective/edda/issues/191) | [#194](https://github.com/subculture-collective/edda/issues/194) | agent |
| [#219](https://github.com/subculture-collective/edda/issues/219) Catalogue SRD spells 24: Shillelagh to Speak with Animals | C / data | [#191](https://github.com/subculture-collective/edda/issues/191) | [#194](https://github.com/subculture-collective/edda/issues/194) | agent |
| [#220](https://github.com/subculture-collective/edda/issues/220) Catalogue SRD spells 25: Speak with Dead to Suggestion | C / data | [#191](https://github.com/subculture-collective/edda/issues/191) | [#194](https://github.com/subculture-collective/edda/issues/194) | agent |
| [#221](https://github.com/subculture-collective/edda/issues/221) Catalogue SRD spells 26: Summon Dragon to Tiny Hut | C / data | [#191](https://github.com/subculture-collective/edda/issues/191) | [#194](https://github.com/subculture-collective/edda/issues/194) | agent |
| [#222](https://github.com/subculture-collective/edda/issues/222) Catalogue SRD spells 27: Tongues to Vitriolic Sphere | C / data | [#191](https://github.com/subculture-collective/edda/issues/191) | [#194](https://github.com/subculture-collective/edda/issues/194) | agent |
| [#223](https://github.com/subculture-collective/edda/issues/223) Catalogue SRD spells 28: Wall of Fire to Wind Wall | C / data | [#191](https://github.com/subculture-collective/edda/issues/191) | [#194](https://github.com/subculture-collective/edda/issues/194) | agent |
| [#224](https://github.com/subculture-collective/edda/issues/224) Catalogue SRD spells 29: Wish to Zone of Truth | C / data | [#191](https://github.com/subculture-collective/edda/issues/191) | [#194](https://github.com/subculture-collective/edda/issues/194) | agent |
| [#225](https://github.com/subculture-collective/edda/issues/225) Catalogue SRD magic-items 01: Adamantine Armor to Armor, +1, +2, or +3 | C / data | [#192](https://github.com/subculture-collective/edda/issues/192) | [#195](https://github.com/subculture-collective/edda/issues/195) | agent |
| [#226](https://github.com/subculture-collective/edda/issues/226) Catalogue SRD magic-items 02: Arrow-Catching Shield to Boots of Levitation | C / data | [#192](https://github.com/subculture-collective/edda/issues/192) | [#195](https://github.com/subculture-collective/edda/issues/195) | agent |
| [#227](https://github.com/subculture-collective/edda/issues/227) Catalogue SRD magic-items 03: Boots of Speed to Carpet of Flying | C / data | [#192](https://github.com/subculture-collective/edda/issues/192) | [#195](https://github.com/subculture-collective/edda/issues/195) | agent |
| [#228](https://github.com/subculture-collective/edda/issues/228) Catalogue SRD magic-items 04: Censer of Controlling Air Elementals to Crystal Ball of Mind Reading | C / data | [#192](https://github.com/subculture-collective/edda/issues/192) | [#195](https://github.com/subculture-collective/edda/issues/195) | agent |
| [#229](https://github.com/subculture-collective/edda/issues/229) Catalogue SRD magic-items 05: Crystal Ball of Telepathy to Dragon Orb | C / data | [#192](https://github.com/subculture-collective/edda/issues/192) | [#195](https://github.com/subculture-collective/edda/issues/195) | agent |
| [#230](https://github.com/subculture-collective/edda/issues/230) Catalogue SRD magic-items 06: Dragon Scale Mail to Elven Chain | C / data | [#192](https://github.com/subculture-collective/edda/issues/192) | [#195](https://github.com/subculture-collective/edda/issues/195) | agent |
| [#231](https://github.com/subculture-collective/edda/issues/231) Catalogue SRD magic-items 07: Energy Bow to Gem of Brightness | C / data | [#192](https://github.com/subculture-collective/edda/issues/192) | [#195](https://github.com/subculture-collective/edda/issues/195) | agent |
| [#232](https://github.com/subculture-collective/edda/issues/232) Catalogue SRD magic-items 08: Gem of Seeing to Headband of Intellect | C / data | [#192](https://github.com/subculture-collective/edda/issues/192) | [#195](https://github.com/subculture-collective/edda/issues/195) | agent |
| [#233](https://github.com/subculture-collective/edda/issues/233) Catalogue SRD magic-items 09: Helm of Brilliance to Ioun Stone | C / data | [#192](https://github.com/subculture-collective/edda/issues/192) | [#195](https://github.com/subculture-collective/edda/issues/195) | agent |
| [#234](https://github.com/subculture-collective/edda/issues/234) Catalogue SRD magic-items 10: Iron Bands to Manual of Golems | C / data | [#192](https://github.com/subculture-collective/edda/issues/192) | [#195](https://github.com/subculture-collective/edda/issues/195) | agent |
| [#235](https://github.com/subculture-collective/edda/issues/235) Catalogue SRD magic-items 11: Manual of Quickness of Action to Oil of Etherealness | C / data | [#192](https://github.com/subculture-collective/edda/issues/192) | [#195](https://github.com/subculture-collective/edda/issues/195) | agent |
| [#236](https://github.com/subculture-collective/edda/issues/236) Catalogue SRD magic-items 12: Oil of Sharpness to Potion of Animal Friendship | C / data | [#192](https://github.com/subculture-collective/edda/issues/192) | [#195](https://github.com/subculture-collective/edda/issues/195) | agent |
| [#237](https://github.com/subculture-collective/edda/issues/237) Catalogue SRD magic-items 13: Potion of Clairvoyance to Potion of Mind Reading | C / data | [#192](https://github.com/subculture-collective/edda/issues/192) | [#195](https://github.com/subculture-collective/edda/issues/195) | agent |
| [#238](https://github.com/subculture-collective/edda/issues/238) Catalogue SRD magic-items 14: Potion of Poison to Ring of Feather Falling | C / data | [#192](https://github.com/subculture-collective/edda/issues/192) | [#195](https://github.com/subculture-collective/edda/issues/195) | agent |
| [#239](https://github.com/subculture-collective/edda/issues/239) Catalogue SRD magic-items 15: Ring of Free Action to Ring of Telekinesis | C / data | [#192](https://github.com/subculture-collective/edda/issues/192) | [#195](https://github.com/subculture-collective/edda/issues/195) | agent |
| [#240](https://github.com/subculture-collective/edda/issues/240) Catalogue SRD magic-items 16: Ring of Three Wishes to Rod of Alertness | C / data | [#192](https://github.com/subculture-collective/edda/issues/192) | [#195](https://github.com/subculture-collective/edda/issues/195) | agent |
| [#241](https://github.com/subculture-collective/edda/issues/241) Catalogue SRD magic-items 17: Rod of Lordly Might to Shield of the Cavalier | C / data | [#192](https://github.com/subculture-collective/edda/issues/192) | [#195](https://github.com/subculture-collective/edda/issues/195) | agent |
| [#242](https://github.com/subculture-collective/edda/issues/242) Catalogue SRD magic-items 18: Shield, +1, +2, or +3 to Staff of Striking | C / data | [#192](https://github.com/subculture-collective/edda/issues/192) | [#195](https://github.com/subculture-collective/edda/issues/195) | agent |
| [#243](https://github.com/subculture-collective/edda/issues/243) Catalogue SRD magic-items 19: Staff of Swarming Insects to Sword of Wounding | C / data | [#192](https://github.com/subculture-collective/edda/issues/192) | [#195](https://github.com/subculture-collective/edda/issues/195) | agent |
| [#244](https://github.com/subculture-collective/edda/issues/244) Catalogue SRD magic-items 20: Talisman of Pure Good to Wand of Binding | C / data | [#192](https://github.com/subculture-collective/edda/issues/192) | [#195](https://github.com/subculture-collective/edda/issues/195) | agent |
| [#245](https://github.com/subculture-collective/edda/issues/245) Catalogue SRD magic-items 21: Wand of Enemy Detection to Wand of the War Mage, +1, +2, or +3 | C / data | [#192](https://github.com/subculture-collective/edda/issues/192) | [#195](https://github.com/subculture-collective/edda/issues/195) | agent |
| [#246](https://github.com/subculture-collective/edda/issues/246) Catalogue SRD magic-items 22: Weapon of Warning to Wings of Flying | C / data | [#192](https://github.com/subculture-collective/edda/issues/192) | [#195](https://github.com/subculture-collective/edda/issues/195) | agent |
| [#247](https://github.com/subculture-collective/edda/issues/247) Catalogue SRD weapons 01: Battleaxe to Hand Crossbow | C / data | [#192](https://github.com/subculture-collective/edda/issues/192) | [#195](https://github.com/subculture-collective/edda/issues/195) | agent |
| [#248](https://github.com/subculture-collective/edda/issues/248) Catalogue SRD weapons 02: Handaxe to Pike | C / data | [#192](https://github.com/subculture-collective/edda/issues/192) | [#195](https://github.com/subculture-collective/edda/issues/195) | agent |
| [#249](https://github.com/subculture-collective/edda/issues/249) Catalogue SRD weapons 03: Pistol to Warhammer | C / data | [#192](https://github.com/subculture-collective/edda/issues/192) | [#195](https://github.com/subculture-collective/edda/issues/195) | agent |
| [#250](https://github.com/subculture-collective/edda/issues/250) Catalogue SRD weapons 04: Whip to Whip | C / data | [#192](https://github.com/subculture-collective/edda/issues/192) | [#195](https://github.com/subculture-collective/edda/issues/195) | agent |
| [#251](https://github.com/subculture-collective/edda/issues/251) Catalogue SRD armor 01: Breastplate to Splint Armor | C / data | [#192](https://github.com/subculture-collective/edda/issues/192) | [#195](https://github.com/subculture-collective/edda/issues/195) | agent |
| [#252](https://github.com/subculture-collective/edda/issues/252) Catalogue SRD armor 02: Studded Leather Armor to Studded Leather Armor | C / data | [#192](https://github.com/subculture-collective/edda/issues/192) | [#195](https://github.com/subculture-collective/edda/issues/195) | agent |
| [#253](https://github.com/subculture-collective/edda/issues/253) Catalogue SRD tools 01: Alchemist’s Supplies to Herbalism Kit | C / data | [#192](https://github.com/subculture-collective/edda/issues/192) | [#195](https://github.com/subculture-collective/edda/issues/195) | agent |
| [#254](https://github.com/subculture-collective/edda/issues/254) Catalogue SRD tools 02: Jeweler’s Tools to Weaver’s Tools | C / data | [#192](https://github.com/subculture-collective/edda/issues/192) | [#195](https://github.com/subculture-collective/edda/issues/195) | agent |
| [#255](https://github.com/subculture-collective/edda/issues/255) Catalogue SRD tools 03: Woodcarver’s Tools to Woodcarver’s Tools | C / data | [#192](https://github.com/subculture-collective/edda/issues/192) | [#195](https://github.com/subculture-collective/edda/issues/195) | agent |
| [#256](https://github.com/subculture-collective/edda/issues/256) Catalogue SRD adventuring-gear 01: Acid to Blanket | C / data | [#192](https://github.com/subculture-collective/edda/issues/192) | [#195](https://github.com/subculture-collective/edda/issues/195) | agent |
| [#257](https://github.com/subculture-collective/edda/issues/257) Catalogue SRD adventuring-gear 02: Block and Tackle to Climber’s Kit | C / data | [#192](https://github.com/subculture-collective/edda/issues/192) | [#195](https://github.com/subculture-collective/edda/issues/195) | agent |
| [#258](https://github.com/subculture-collective/edda/issues/258) Catalogue SRD adventuring-gear 03: Clothes, Fine to Grappling Hook | C / data | [#192](https://github.com/subculture-collective/edda/issues/192) | [#195](https://github.com/subculture-collective/edda/issues/195) | agent |
| [#259](https://github.com/subculture-collective/edda/issues/259) Catalogue SRD adventuring-gear 04: Healer’s Kit to Lock | C / data | [#192](https://github.com/subculture-collective/edda/issues/192) | [#195](https://github.com/subculture-collective/edda/issues/195) | agent |
| [#260](https://github.com/subculture-collective/edda/issues/260) Catalogue SRD adventuring-gear 05: Magnifying Glass to Pot, Iron | C / data | [#192](https://github.com/subculture-collective/edda/issues/192) | [#195](https://github.com/subculture-collective/edda/issues/195) | agent |
| [#261](https://github.com/subculture-collective/edda/issues/261) Catalogue SRD adventuring-gear 06: Potion of Healing to Signal Whistle | C / data | [#192](https://github.com/subculture-collective/edda/issues/192) | [#195](https://github.com/subculture-collective/edda/issues/195) | agent |
| [#262](https://github.com/subculture-collective/edda/issues/262) Catalogue SRD adventuring-gear 07: Spell Scroll to Waterskin | C / data | [#192](https://github.com/subculture-collective/edda/issues/192) | [#195](https://github.com/subculture-collective/edda/issues/195) | agent |
| [#263](https://github.com/subculture-collective/edda/issues/263) Cast level-one Cure Wounds on one willing adjacent ally | B / contract/runtime | [#191](https://github.com/subculture-collective/edda/issues/191) | [#194](https://github.com/subculture-collective/edda/issues/194) | agent |
| [#264](https://github.com/subculture-collective/edda/issues/264) Cast Light on one known carried object with durable expiration | B / contract/runtime | [#191](https://github.com/subculture-collective/edda/issues/191) | [#145](https://github.com/subculture-collective/edda/issues/145), [#194](https://github.com/subculture-collective/edda/issues/194) | agent |
| [#265](https://github.com/subculture-collective/edda/issues/265) Confirm Detect Magic from one recorded Magic Initiate free use | B / contract/runtime | [#191](https://github.com/subculture-collective/edda/issues/191) | [#194](https://github.com/subculture-collective/edda/issues/194) | agent |
| [#267](https://github.com/subculture-collective/edda/issues/267) Confirm chain mail and shield equipment changes through the sheet | B / contract/runtime | [#192](https://github.com/subculture-collective/edda/issues/192) | [#195](https://github.com/subculture-collective/edda/issues/195) | agent |
| [#285](https://github.com/subculture-collective/edda/issues/285) Catalogue SRD advancement table source records | C / data | [#282](https://github.com/subculture-collective/edda/issues/282) | [#284](https://github.com/subculture-collective/edda/issues/284) | agent |
| [#286](https://github.com/subculture-collective/edda/issues/286) Catalogue SRD Barbarian class levels and feature keys | C / data | [#282](https://github.com/subculture-collective/edda/issues/282) | [#284](https://github.com/subculture-collective/edda/issues/284) | agent |
| [#287](https://github.com/subculture-collective/edda/issues/287) Catalogue SRD Bard class levels and feature keys | C / data | [#282](https://github.com/subculture-collective/edda/issues/282) | [#284](https://github.com/subculture-collective/edda/issues/284) | agent |
| [#288](https://github.com/subculture-collective/edda/issues/288) Catalogue SRD Cleric class levels and feature keys | C / data | [#282](https://github.com/subculture-collective/edda/issues/282) | [#284](https://github.com/subculture-collective/edda/issues/284) | agent |
| [#289](https://github.com/subculture-collective/edda/issues/289) Catalogue SRD Druid class levels and feature keys | C / data | [#282](https://github.com/subculture-collective/edda/issues/282) | [#284](https://github.com/subculture-collective/edda/issues/284) | agent |
| [#290](https://github.com/subculture-collective/edda/issues/290) Catalogue SRD Fighter class levels and feature keys | C / data | [#282](https://github.com/subculture-collective/edda/issues/282) | [#284](https://github.com/subculture-collective/edda/issues/284) | agent |
| [#291](https://github.com/subculture-collective/edda/issues/291) Catalogue SRD Monk class levels and feature keys | C / data | [#282](https://github.com/subculture-collective/edda/issues/282) | [#284](https://github.com/subculture-collective/edda/issues/284) | agent |
| [#292](https://github.com/subculture-collective/edda/issues/292) Catalogue SRD Paladin class levels and feature keys | C / data | [#282](https://github.com/subculture-collective/edda/issues/282) | [#284](https://github.com/subculture-collective/edda/issues/284) | agent |
| [#293](https://github.com/subculture-collective/edda/issues/293) Catalogue SRD Ranger class levels and feature keys | C / data | [#282](https://github.com/subculture-collective/edda/issues/282) | [#284](https://github.com/subculture-collective/edda/issues/284) | agent |
| [#294](https://github.com/subculture-collective/edda/issues/294) Catalogue SRD Rogue class levels and feature keys | C / data | [#282](https://github.com/subculture-collective/edda/issues/282) | [#284](https://github.com/subculture-collective/edda/issues/284) | agent |
| [#295](https://github.com/subculture-collective/edda/issues/295) Catalogue SRD Sorcerer class levels and feature keys | C / data | [#282](https://github.com/subculture-collective/edda/issues/282) | [#284](https://github.com/subculture-collective/edda/issues/284) | agent |
| [#296](https://github.com/subculture-collective/edda/issues/296) Catalogue SRD Warlock class levels and feature keys | C / data | [#282](https://github.com/subculture-collective/edda/issues/282) | [#284](https://github.com/subculture-collective/edda/issues/284) | agent |
| [#297](https://github.com/subculture-collective/edda/issues/297) Catalogue SRD Wizard class levels and feature keys | C / data | [#282](https://github.com/subculture-collective/edda/issues/282) | [#284](https://github.com/subculture-collective/edda/issues/284) | agent |
| [#298](https://github.com/subculture-collective/edda/issues/298) Catalogue SRD feats epic boon source records | C / data | [#282](https://github.com/subculture-collective/edda/issues/282) | [#284](https://github.com/subculture-collective/edda/issues/284) | agent |
| [#299](https://github.com/subculture-collective/edda/issues/299) Catalogue SRD feats fighting style source records | C / data | [#282](https://github.com/subculture-collective/edda/issues/282) | [#284](https://github.com/subculture-collective/edda/issues/284) | agent |
| [#300](https://github.com/subculture-collective/edda/issues/300) Catalogue SRD feats general source records | C / data | [#282](https://github.com/subculture-collective/edda/issues/282) | [#284](https://github.com/subculture-collective/edda/issues/284) | agent |
| [#301](https://github.com/subculture-collective/edda/issues/301) Catalogue SRD feats origin source records | C / data | [#282](https://github.com/subculture-collective/edda/issues/282) | [#284](https://github.com/subculture-collective/edda/issues/284) | agent |
| [#302](https://github.com/subculture-collective/edda/issues/302) Catalogue SRD options invocations 01 source records | C / data | [#282](https://github.com/subculture-collective/edda/issues/282) | [#284](https://github.com/subculture-collective/edda/issues/284) | agent |
| [#303](https://github.com/subculture-collective/edda/issues/303) Catalogue SRD options invocations 02 source records | C / data | [#282](https://github.com/subculture-collective/edda/issues/282) | [#284](https://github.com/subculture-collective/edda/issues/284) | agent |
| [#304](https://github.com/subculture-collective/edda/issues/304) Catalogue SRD options invocations 03 source records | C / data | [#282](https://github.com/subculture-collective/edda/issues/282) | [#284](https://github.com/subculture-collective/edda/issues/284) | agent |
| [#305](https://github.com/subculture-collective/edda/issues/305) Catalogue SRD options metamagic source records | C / data | [#282](https://github.com/subculture-collective/edda/issues/282) | [#284](https://github.com/subculture-collective/edda/issues/284) | agent |
| [#306](https://github.com/subculture-collective/edda/issues/306) Catalogue SRD origins backgrounds source records | C / data | [#282](https://github.com/subculture-collective/edda/issues/282) | [#284](https://github.com/subculture-collective/edda/issues/284) | agent |
| [#307](https://github.com/subculture-collective/edda/issues/307) Catalogue SRD origins species source records | C / data | [#282](https://github.com/subculture-collective/edda/issues/282) | [#284](https://github.com/subculture-collective/edda/issues/284) | agent |
| [#308](https://github.com/subculture-collective/edda/issues/308) Catalogue SRD Path of the Berserker subclass feature keys | C / data | [#282](https://github.com/subculture-collective/edda/issues/282) | [#284](https://github.com/subculture-collective/edda/issues/284) | agent |
| [#309](https://github.com/subculture-collective/edda/issues/309) Catalogue SRD College of Lore subclass feature keys | C / data | [#282](https://github.com/subculture-collective/edda/issues/282) | [#284](https://github.com/subculture-collective/edda/issues/284) | agent |
| [#310](https://github.com/subculture-collective/edda/issues/310) Catalogue SRD Life Domain subclass feature keys | C / data | [#282](https://github.com/subculture-collective/edda/issues/282) | [#284](https://github.com/subculture-collective/edda/issues/284) | agent |
| [#311](https://github.com/subculture-collective/edda/issues/311) Catalogue SRD Circle of the Land subclass feature keys | C / data | [#282](https://github.com/subculture-collective/edda/issues/282) | [#284](https://github.com/subculture-collective/edda/issues/284) | agent |
| [#312](https://github.com/subculture-collective/edda/issues/312) Catalogue SRD Champion subclass feature keys | C / data | [#282](https://github.com/subculture-collective/edda/issues/282) | [#284](https://github.com/subculture-collective/edda/issues/284) | agent |
| [#313](https://github.com/subculture-collective/edda/issues/313) Catalogue SRD Warrior of the Open Hand subclass feature keys | C / data | [#282](https://github.com/subculture-collective/edda/issues/282) | [#284](https://github.com/subculture-collective/edda/issues/284) | agent |
| [#314](https://github.com/subculture-collective/edda/issues/314) Catalogue SRD Oath of Devotion subclass feature keys | C / data | [#282](https://github.com/subculture-collective/edda/issues/282) | [#284](https://github.com/subculture-collective/edda/issues/284) | agent |
| [#315](https://github.com/subculture-collective/edda/issues/315) Catalogue SRD Hunter subclass feature keys | C / data | [#282](https://github.com/subculture-collective/edda/issues/282) | [#284](https://github.com/subculture-collective/edda/issues/284) | agent |
| [#316](https://github.com/subculture-collective/edda/issues/316) Catalogue SRD Thief subclass feature keys | C / data | [#282](https://github.com/subculture-collective/edda/issues/282) | [#284](https://github.com/subculture-collective/edda/issues/284) | agent |
| [#317](https://github.com/subculture-collective/edda/issues/317) Catalogue SRD Draconic Sorcery subclass feature keys | C / data | [#282](https://github.com/subculture-collective/edda/issues/282) | [#284](https://github.com/subculture-collective/edda/issues/284) | agent |
| [#318](https://github.com/subculture-collective/edda/issues/318) Catalogue SRD Fiend Patron subclass feature keys | C / data | [#282](https://github.com/subculture-collective/edda/issues/282) | [#284](https://github.com/subculture-collective/edda/issues/284) | agent |
| [#319](https://github.com/subculture-collective/edda/issues/319) Catalogue SRD Evoker subclass feature keys | C / data | [#282](https://github.com/subculture-collective/edda/issues/282) | [#284](https://github.com/subculture-collective/edda/issues/284) | agent |
| [#320](https://github.com/subculture-collective/edda/issues/320) Catalogue SRD multiclass prerequisites and shared progression tables | C / data | [#282](https://github.com/subculture-collective/edda/issues/282) | [#284](https://github.com/subculture-collective/edda/issues/284) | agent |

### Dependency wave 2: 25 work issues

| Issue | Group / kind | Parent | Open blockers | Checkpoint |
| --- | --- | --- | --- | --- |
| [#89](https://github.com/subculture-collective/edda/issues/89) Qualify the declared gameplay slice across desktop, phone and reconnects | A / qualification | [#54](https://github.com/subculture-collective/edda/issues/54) | [#78](https://github.com/subculture-collective/edda/issues/78) | human checkpoint |
| [#179](https://github.com/subculture-collective/edda/issues/179) Resolve Wolf Pack Tactics and size-gated Bite Prone through durable encounters | B / contract/runtime | [#54](https://github.com/subculture-collective/edda/issues/54) | [#148](https://github.com/subculture-collective/edda/issues/148) | agent |
| [#180](https://github.com/subculture-collective/edda/issues/180) Resolve the Brown Bear Multiattack as one atomic NPC action | B / contract/runtime | [#54](https://github.com/subculture-collective/edda/issues/54) | [#148](https://github.com/subculture-collective/edda/issues/148) | agent |
| [#181](https://github.com/subculture-collective/edda/issues/181) Bind one Find Familiar creature to its summoner and turn permissions | D / investigation | [#54](https://github.com/subculture-collective/edda/issues/54) | [#148](https://github.com/subculture-collective/edda/issues/148), [#194](https://github.com/subculture-collective/edda/issues/194), [#205](https://github.com/subculture-collective/edda/issues/205) | agent |
| [#182](https://github.com/subculture-collective/edda/issues/182) Split unimplemented monster mechanic keys into exact fixture PRs | D / investigation | [#54](https://github.com/subculture-collective/edda/issues/54) | [#151](https://github.com/subculture-collective/edda/issues/151), [#152](https://github.com/subculture-collective/edda/issues/152), [#153](https://github.com/subculture-collective/edda/issues/153), [#154](https://github.com/subculture-collective/edda/issues/154), [#155](https://github.com/subculture-collective/edda/issues/155), [#156](https://github.com/subculture-collective/edda/issues/156), [#157](https://github.com/subculture-collective/edda/issues/157), [#158](https://github.com/subculture-collective/edda/issues/158), [#159](https://github.com/subculture-collective/edda/issues/159), [#160](https://github.com/subculture-collective/edda/issues/160), [#161](https://github.com/subculture-collective/edda/issues/161), [#162](https://github.com/subculture-collective/edda/issues/162), [#163](https://github.com/subculture-collective/edda/issues/163), [#164](https://github.com/subculture-collective/edda/issues/164), [#165](https://github.com/subculture-collective/edda/issues/165), [#166](https://github.com/subculture-collective/edda/issues/166), [#167](https://github.com/subculture-collective/edda/issues/167), [#168](https://github.com/subculture-collective/edda/issues/168), [#169](https://github.com/subculture-collective/edda/issues/169), [#170](https://github.com/subculture-collective/edda/issues/170), [#171](https://github.com/subculture-collective/edda/issues/171), [#172](https://github.com/subculture-collective/edda/issues/172), [#173](https://github.com/subculture-collective/edda/issues/173), [#174](https://github.com/subculture-collective/edda/issues/174), [#175](https://github.com/subculture-collective/edda/issues/175), [#176](https://github.com/subculture-collective/edda/issues/176), [#177](https://github.com/subculture-collective/edda/issues/177), [#178](https://github.com/subculture-collective/edda/issues/178), [#184](https://github.com/subculture-collective/edda/issues/184) | agent |
| [#185](https://github.com/subculture-collective/edda/issues/185) Split remaining spell-created creature control and scaling into fixture PRs | D / investigation | [#54](https://github.com/subculture-collective/edda/issues/54) | [#184](https://github.com/subculture-collective/edda/issues/184), [#194](https://github.com/subculture-collective/edda/issues/194), [#196](https://github.com/subculture-collective/edda/issues/196), [#205](https://github.com/subculture-collective/edda/issues/205), [#207](https://github.com/subculture-collective/edda/issues/207), [#221](https://github.com/subculture-collective/edda/issues/221) | agent |
| [#266](https://github.com/subculture-collective/edda/issues/266) Administer one basic Potion of Healing to a willing ally | B / contract/runtime | [#192](https://github.com/subculture-collective/edda/issues/192) | [#195](https://github.com/subculture-collective/edda/issues/195), [#263](https://github.com/subculture-collective/edda/issues/263) | agent |
| [#277](https://github.com/subculture-collective/edda/issues/277) Assign every remaining spell mechanic key to bounded implementation PRs | D / investigation | [#191](https://github.com/subculture-collective/edda/issues/191) | [#196](https://github.com/subculture-collective/edda/issues/196), [#197](https://github.com/subculture-collective/edda/issues/197), [#198](https://github.com/subculture-collective/edda/issues/198), [#199](https://github.com/subculture-collective/edda/issues/199), [#200](https://github.com/subculture-collective/edda/issues/200), [#201](https://github.com/subculture-collective/edda/issues/201), [#202](https://github.com/subculture-collective/edda/issues/202), [#203](https://github.com/subculture-collective/edda/issues/203), [#204](https://github.com/subculture-collective/edda/issues/204), [#205](https://github.com/subculture-collective/edda/issues/205), [#206](https://github.com/subculture-collective/edda/issues/206), [#207](https://github.com/subculture-collective/edda/issues/207), [#208](https://github.com/subculture-collective/edda/issues/208), [#209](https://github.com/subculture-collective/edda/issues/209), [#210](https://github.com/subculture-collective/edda/issues/210), [#211](https://github.com/subculture-collective/edda/issues/211), [#212](https://github.com/subculture-collective/edda/issues/212), [#213](https://github.com/subculture-collective/edda/issues/213), [#214](https://github.com/subculture-collective/edda/issues/214), [#215](https://github.com/subculture-collective/edda/issues/215), [#216](https://github.com/subculture-collective/edda/issues/216), [#217](https://github.com/subculture-collective/edda/issues/217), [#218](https://github.com/subculture-collective/edda/issues/218), [#219](https://github.com/subculture-collective/edda/issues/219), [#220](https://github.com/subculture-collective/edda/issues/220), [#221](https://github.com/subculture-collective/edda/issues/221), [#222](https://github.com/subculture-collective/edda/issues/222), [#223](https://github.com/subculture-collective/edda/issues/223), [#224](https://github.com/subculture-collective/edda/issues/224) | agent |
| [#278](https://github.com/subculture-collective/edda/issues/278) Assign every remaining equipment and magic-item key to bounded implementation PRs | D / investigation | [#192](https://github.com/subculture-collective/edda/issues/192) | [#225](https://github.com/subculture-collective/edda/issues/225), [#226](https://github.com/subculture-collective/edda/issues/226), [#227](https://github.com/subculture-collective/edda/issues/227), [#228](https://github.com/subculture-collective/edda/issues/228), [#229](https://github.com/subculture-collective/edda/issues/229), [#230](https://github.com/subculture-collective/edda/issues/230), [#231](https://github.com/subculture-collective/edda/issues/231), [#232](https://github.com/subculture-collective/edda/issues/232), [#233](https://github.com/subculture-collective/edda/issues/233), [#234](https://github.com/subculture-collective/edda/issues/234), [#235](https://github.com/subculture-collective/edda/issues/235), [#236](https://github.com/subculture-collective/edda/issues/236), [#237](https://github.com/subculture-collective/edda/issues/237), [#238](https://github.com/subculture-collective/edda/issues/238), [#239](https://github.com/subculture-collective/edda/issues/239), [#240](https://github.com/subculture-collective/edda/issues/240), [#241](https://github.com/subculture-collective/edda/issues/241), [#242](https://github.com/subculture-collective/edda/issues/242), [#243](https://github.com/subculture-collective/edda/issues/243), [#244](https://github.com/subculture-collective/edda/issues/244), [#245](https://github.com/subculture-collective/edda/issues/245), [#246](https://github.com/subculture-collective/edda/issues/246), [#247](https://github.com/subculture-collective/edda/issues/247), [#248](https://github.com/subculture-collective/edda/issues/248), [#249](https://github.com/subculture-collective/edda/issues/249), [#250](https://github.com/subculture-collective/edda/issues/250), [#251](https://github.com/subculture-collective/edda/issues/251), [#252](https://github.com/subculture-collective/edda/issues/252), [#253](https://github.com/subculture-collective/edda/issues/253), [#254](https://github.com/subculture-collective/edda/issues/254), [#255](https://github.com/subculture-collective/edda/issues/255), [#256](https://github.com/subculture-collective/edda/issues/256), [#257](https://github.com/subculture-collective/edda/issues/257), [#258](https://github.com/subculture-collective/edda/issues/258), [#259](https://github.com/subculture-collective/edda/issues/259), [#260](https://github.com/subculture-collective/edda/issues/260), [#261](https://github.com/subculture-collective/edda/issues/261), [#262](https://github.com/subculture-collective/edda/issues/262) | agent |
| [#321](https://github.com/subculture-collective/edda/issues/321) Split remaining Barbarian features and levels into exact playable fixtures | D / investigation | [#283](https://github.com/subculture-collective/edda/issues/283) | [#285](https://github.com/subculture-collective/edda/issues/285), [#286](https://github.com/subculture-collective/edda/issues/286), [#308](https://github.com/subculture-collective/edda/issues/308) | agent |
| [#322](https://github.com/subculture-collective/edda/issues/322) Split remaining Bard features and levels into exact playable fixtures | D / investigation | [#283](https://github.com/subculture-collective/edda/issues/283) | [#285](https://github.com/subculture-collective/edda/issues/285), [#287](https://github.com/subculture-collective/edda/issues/287), [#309](https://github.com/subculture-collective/edda/issues/309) | agent |
| [#323](https://github.com/subculture-collective/edda/issues/323) Split remaining Cleric features and levels into exact playable fixtures | D / investigation | [#283](https://github.com/subculture-collective/edda/issues/283) | [#285](https://github.com/subculture-collective/edda/issues/285), [#288](https://github.com/subculture-collective/edda/issues/288), [#310](https://github.com/subculture-collective/edda/issues/310) | agent |
| [#324](https://github.com/subculture-collective/edda/issues/324) Split remaining Druid features and levels into exact playable fixtures | D / investigation | [#283](https://github.com/subculture-collective/edda/issues/283) | [#285](https://github.com/subculture-collective/edda/issues/285), [#289](https://github.com/subculture-collective/edda/issues/289), [#311](https://github.com/subculture-collective/edda/issues/311) | agent |
| [#325](https://github.com/subculture-collective/edda/issues/325) Split remaining Fighter features and levels into exact playable fixtures | D / investigation | [#283](https://github.com/subculture-collective/edda/issues/283) | [#285](https://github.com/subculture-collective/edda/issues/285), [#290](https://github.com/subculture-collective/edda/issues/290), [#312](https://github.com/subculture-collective/edda/issues/312) | agent |
| [#326](https://github.com/subculture-collective/edda/issues/326) Split remaining Monk features and levels into exact playable fixtures | D / investigation | [#283](https://github.com/subculture-collective/edda/issues/283) | [#285](https://github.com/subculture-collective/edda/issues/285), [#291](https://github.com/subculture-collective/edda/issues/291), [#313](https://github.com/subculture-collective/edda/issues/313) | agent |
| [#327](https://github.com/subculture-collective/edda/issues/327) Split remaining Paladin features and levels into exact playable fixtures | D / investigation | [#283](https://github.com/subculture-collective/edda/issues/283) | [#285](https://github.com/subculture-collective/edda/issues/285), [#292](https://github.com/subculture-collective/edda/issues/292), [#314](https://github.com/subculture-collective/edda/issues/314) | agent |
| [#328](https://github.com/subculture-collective/edda/issues/328) Split remaining Ranger features and levels into exact playable fixtures | D / investigation | [#283](https://github.com/subculture-collective/edda/issues/283) | [#285](https://github.com/subculture-collective/edda/issues/285), [#293](https://github.com/subculture-collective/edda/issues/293), [#315](https://github.com/subculture-collective/edda/issues/315) | agent |
| [#329](https://github.com/subculture-collective/edda/issues/329) Split remaining Rogue features and levels into exact playable fixtures | D / investigation | [#283](https://github.com/subculture-collective/edda/issues/283) | [#285](https://github.com/subculture-collective/edda/issues/285), [#294](https://github.com/subculture-collective/edda/issues/294), [#316](https://github.com/subculture-collective/edda/issues/316) | agent |
| [#330](https://github.com/subculture-collective/edda/issues/330) Split remaining Sorcerer features and levels into exact playable fixtures | D / investigation | [#283](https://github.com/subculture-collective/edda/issues/283) | [#285](https://github.com/subculture-collective/edda/issues/285), [#295](https://github.com/subculture-collective/edda/issues/295), [#305](https://github.com/subculture-collective/edda/issues/305), [#317](https://github.com/subculture-collective/edda/issues/317) | agent |
| [#331](https://github.com/subculture-collective/edda/issues/331) Split remaining Warlock features and levels into exact playable fixtures | D / investigation | [#283](https://github.com/subculture-collective/edda/issues/283) | [#285](https://github.com/subculture-collective/edda/issues/285), [#296](https://github.com/subculture-collective/edda/issues/296), [#302](https://github.com/subculture-collective/edda/issues/302), [#303](https://github.com/subculture-collective/edda/issues/303), [#304](https://github.com/subculture-collective/edda/issues/304), [#318](https://github.com/subculture-collective/edda/issues/318) | agent |
| [#332](https://github.com/subculture-collective/edda/issues/332) Split remaining Wizard features and levels into exact playable fixtures | D / investigation | [#283](https://github.com/subculture-collective/edda/issues/283) | [#285](https://github.com/subculture-collective/edda/issues/285), [#297](https://github.com/subculture-collective/edda/issues/297), [#319](https://github.com/subculture-collective/edda/issues/319) | agent |
| [#333](https://github.com/subculture-collective/edda/issues/333) Split remaining origin and creation choices into exact playable fixtures | D / investigation | [#283](https://github.com/subculture-collective/edda/issues/283) | [#306](https://github.com/subculture-collective/edda/issues/306), [#307](https://github.com/subculture-collective/edda/issues/307) | agent |
| [#334](https://github.com/subculture-collective/edda/issues/334) Split all remaining feat prerequisites and effects into exact playable fixtures | D / investigation | [#283](https://github.com/subculture-collective/edda/issues/283) | [#298](https://github.com/subculture-collective/edda/issues/298), [#299](https://github.com/subculture-collective/edda/issues/299), [#300](https://github.com/subculture-collective/edda/issues/300), [#301](https://github.com/subculture-collective/edda/issues/301) | agent |
| [#335](https://github.com/subculture-collective/edda/issues/335) Split multiclass advancement and feature interactions into exact fixture PRs | D / investigation | [#283](https://github.com/subculture-collective/edda/issues/283) | [#320](https://github.com/subculture-collective/edda/issues/320) | agent |
| [#336](https://github.com/subculture-collective/edda/issues/336) Split remaining advancement choices and higher-level starts into exact fixture PRs | D / investigation | [#283](https://github.com/subculture-collective/edda/issues/283) | [#285](https://github.com/subculture-collective/edda/issues/285) | agent |

### Dependency wave 3: 3 work issues

| Issue | Group / kind | Parent | Open blockers | Checkpoint |
| --- | --- | --- | --- | --- |
| [#183](https://github.com/subculture-collective/edda/issues/183) Qualify complete monster and exploration coverage before advertising full support | E / qualification | [#54](https://github.com/subculture-collective/edda/issues/54) | [#145](https://github.com/subculture-collective/edda/issues/145), [#146](https://github.com/subculture-collective/edda/issues/146), [#149](https://github.com/subculture-collective/edda/issues/149), [#150](https://github.com/subculture-collective/edda/issues/150), [#179](https://github.com/subculture-collective/edda/issues/179), [#180](https://github.com/subculture-collective/edda/issues/180), [#181](https://github.com/subculture-collective/edda/issues/181), [#182](https://github.com/subculture-collective/edda/issues/182), [#185](https://github.com/subculture-collective/edda/issues/185) | agent |
| [#279](https://github.com/subculture-collective/edda/issues/279) Qualify every declared SRD spell and effect before full coverage claims | E / qualification | [#191](https://github.com/subculture-collective/edda/issues/191) | [#181](https://github.com/subculture-collective/edda/issues/181), [#185](https://github.com/subculture-collective/edda/issues/185), [#263](https://github.com/subculture-collective/edda/issues/263), [#264](https://github.com/subculture-collective/edda/issues/264), [#265](https://github.com/subculture-collective/edda/issues/265), [#268](https://github.com/subculture-collective/edda/issues/268), [#269](https://github.com/subculture-collective/edda/issues/269), [#270](https://github.com/subculture-collective/edda/issues/270), [#271](https://github.com/subculture-collective/edda/issues/271), [#272](https://github.com/subculture-collective/edda/issues/272), [#273](https://github.com/subculture-collective/edda/issues/273), [#277](https://github.com/subculture-collective/edda/issues/277) | agent |
| [#337](https://github.com/subculture-collective/edda/issues/337) Qualify all SRD character origins classes levels and multiclass coverage | E / qualification | [#283](https://github.com/subculture-collective/edda/issues/283) | [#321](https://github.com/subculture-collective/edda/issues/321), [#322](https://github.com/subculture-collective/edda/issues/322), [#323](https://github.com/subculture-collective/edda/issues/323), [#324](https://github.com/subculture-collective/edda/issues/324), [#325](https://github.com/subculture-collective/edda/issues/325), [#326](https://github.com/subculture-collective/edda/issues/326), [#327](https://github.com/subculture-collective/edda/issues/327), [#328](https://github.com/subculture-collective/edda/issues/328), [#329](https://github.com/subculture-collective/edda/issues/329), [#330](https://github.com/subculture-collective/edda/issues/330), [#331](https://github.com/subculture-collective/edda/issues/331), [#332](https://github.com/subculture-collective/edda/issues/332), [#333](https://github.com/subculture-collective/edda/issues/333), [#334](https://github.com/subculture-collective/edda/issues/334), [#335](https://github.com/subculture-collective/edda/issues/335), [#336](https://github.com/subculture-collective/edda/issues/336), [#338](https://github.com/subculture-collective/edda/issues/338), [#339](https://github.com/subculture-collective/edda/issues/339) | agent |

### Dependency wave 4: 1 work issues

| Issue | Group / kind | Parent | Open blockers | Checkpoint |
| --- | --- | --- | --- | --- |
| [#280](https://github.com/subculture-collective/edda/issues/280) Qualify declared SRD item families and variants before full coverage claims | E / qualification | [#192](https://github.com/subculture-collective/edda/issues/192) | [#266](https://github.com/subculture-collective/edda/issues/266), [#267](https://github.com/subculture-collective/edda/issues/267), [#274](https://github.com/subculture-collective/edda/issues/274), [#275](https://github.com/subculture-collective/edda/issues/275), [#276](https://github.com/subculture-collective/edda/issues/276), [#278](https://github.com/subculture-collective/edda/issues/278), [#279](https://github.com/subculture-collective/edda/issues/279) | agent |

<!-- edda-271-followups:start -->
## Concentration, duration and stacking split (#271)

Decision: `docs/investigations/issue-271-effect-lifecycle-decision.md`; exact fixture/source/issue mapping: `docs/investigations/issue-271-effect-lifecycle-scope.json`.
Six bounded implementation PRs and two one-session source spikes attach natively under #193. Every listed issue directly blocks both #279 and #280; spikes must publish and wire their terminal implementation children before closure. Catalogue/condition/casting/class owners retain their scope. Isolated higher-level engine fixtures do not establish production eligibility. #271 remains open for its decision PR; no full-coverage claim or implementation completion is implied.

- [ ] #351 Track two Shield of Faith casters with independent recipient effects (L; task; after #194, #218)
- [ ] #352 Apply one level-one Bless cast to up to three explicit recipients (L; task; after #198, #351)
- [ ] #353 End the exact concentration cast on resolved damage and incapacitation events (M; task; after #351)
- [ ] #354 Resolve Acid Arrow initial damage and one durable next-turn end tick (L; task; after #196)
- [ ] #355 Dispel level-three-or-lower spells on one explicit creature (L; task; after #203, #352)
- [ ] #356 Persist one Counterspell reaction window around Cure Wounds casting (L; task; after #201)
- [ ] #358 Split Heroism start-turn and Spirit Guardians entry/end timing fixtures (M; spike; after existing bounded baseline)
- [ ] #359 Split Dispel Magic higher-level checks and object/effect targeting (M; spike; after #203, #355)
<!-- edda-271-followups:end -->
