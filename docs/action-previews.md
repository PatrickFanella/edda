# Read-only action preparation (#63)

The Character sheet offers a preview selector for the supported SRD 5.2.1
ability checks, saving throws and skills. Previewing never performs an action.
The ordinary check modifier and source terms come from the same `CheckModifier`
validation used by `ResolveCheck`; no client or model computes them. Skills may
omit their ability in the request; the server returns the canonical pairing.

`GET /api/v1/campaigns/{id}/character/action-preview` accepts `actor_id`, `kind`
(`ability-check` or `saving-throw`), `ability`, `skill` and `entry_id`. It returns
the shared `game.ActionPreview` / `api.ActionPreview` with a typed intent,
`rules_version`, `expected_revision`, availability/reason, known targets,
options, direct resource costs and source identifiers. Unknown or repeated query
parameters are rejected, including difficulty, disclosure flags and modifiers.
Responses use `Cache-Control: no-store`.

The endpoint sits inside the existing campaign ownership middleware and repeats
the owner/actor/campaign filter in a single SQL statement. That statement reads
only campaign metadata and the selected character. It never reads NPCs, private
world facts, enemy stats, dice or hidden DCs. A foreign or absent actor returns
the same 404. Unsupported builds, inactive characters, unavailable sheet entries
and invalid checks return player-safe reasons without a modifier or targets.
Legacy character reads and legacy gameplay keep their existing behavior.

The snapshot fingerprint is an opaque `actor-v1:` SHA-256 over the trusted
campaign/actor inputs consumed by `game.PrepareCharacterAction`. It covers rules
version, campaign status/update time, actor identity, stats, HP, XP, status,
location and update time. Repeating a preview of the same snapshot returns the
same revision; a change to those inputs invalidates it. It is not an action ID,
reservation, whole-world revision, or authorization to commit. Confirmation
(#64) must reload and compare these same inputs in its protected action path,
and validate any additional scene/resource inputs it consumes. The fingerprint
must not be used alone to authorize an action.

Checks offer the ordinary roll option. Advantage/disadvantage circumstances,
DCs, scene targets, and contextual action/time costs require server adjudication
before a future confirmation; this slice does not let the player declare them.
An empty `costs` array means no direct resource cost for the check, not a promise
of free action economy or no elapsed time. Saving throws identify only the actor
as a known target. No scene target is inferred for an ability or skill check.
Attacks, Second Wind and other sheet entries return unavailable until their own
roadmap slices implement preparation/resolution. No catalogue coverage is raised.

Preparation has no dice, provider, persistence or time interface. The web caller
shows pending, ready, unavailable and retry states; changing selection clears
the previous result, and campaign/actor changes remount the preview. There is no
confirmation button. The pure Go boundary is available to TUI consumers without
an HTTP dependency. Integration fixtures compare the exact stored campaign,
character, time, history and turn-attempt snapshots before and after repeated
preparation; frontend tests inspect the selected request, terms, safe failures
and lack of a confirmation path.
