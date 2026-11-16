# Chain mail and shield confirmation (#267)

This adds two intents to the existing shared game/engine contract:
`equipment-equip` and `equipment-remove`, with the owned item UUID in `entry_id`.
The web sheet exposes these through its existing preview/confirm menu. TUI
callers use the same engine methods and durable receipt contract.

The resolver uses ordinary chain mail and shield fixture identities, never
client AC or arbitrary properties. It requires exactly one owned common item
with only `srd521_id`, reconciled equipped inventory/build state, one suit and
one shield at most, and known hand use. The only supported held weapon is the
existing one-handed longsword. Unknown equipped weapons, magic/custom
properties, stacked equipment and inconsistent legacy saves stay readable but
cannot mutate through this path. No automatic inventory repair occurs.

Fighters have heavy armor and shield training. Clerics have shield training;
heavy armor requires the recorded Protector choice. Chain mail mutation requires
Strength 13 because reduced-speed armor is outside this fixture. Removing chain
mail derives unarmored AC from Dexterity and stops applying Defense; a shield
continues to contribute +2. Concentration contributes separately until expiry.
Equipment changes are quiet downtime outside supported encounters.

## Source and adaptation

The local PDF `/tmp/edda-wave1/SRD_CC_v5.2.1.pdf` matches the repository pin:
`8974902d109d6e63672d7c490bde9ccf052410503d9cfa768237154fbc5e3d87`.
SRD 5.2.1 p. 92 specifies chain mail AC 16, Strength 13, heavy armor training,
600 seconds to don and 300 seconds to doff, shield +2 and a Utilize action to
don/doff. The fixture accounts for a shield Utilize action as one six-second
interval outside combat; six seconds is Edda's action-time convention rather
than a separate duration stated by the armor table. Page 103's magic-item rules
do not authorize any magic/cursed item in this implementation.

Values and receipts are Edda adaptations. Source attribution and CC-BY-4.0
license: [SRD conformance register](../srd-5.2.1-conformance.md).

## Criterion to proof

- Preview and explicit confirm: `equipment.spec.ts` exercises desktop (1440px)
  and phone (390px), proves zero mutations before confirm, and inspects AC/time.
  `equipment_test.go` tests don/doff durations, AC, Defense, training and expiry.
- Atomic equipment/build/time/history: `TestIntegrationEquipmentAction` uses the
  existing disposable PostgreSQL harness, injects a history-write failure,
  verifies rollback, reloads receipts through a new engine, and replays without
  time advancement or dice. It also covers shield-triggered concentration expiry.
- Stale/negative/ownership: the same integration fixture changes item ownership,
  quantity, identity, magic properties, equipped state and another equipped
  item's timestamp; each rejects confirmation. An equipped unsupported
  two-handed weapon rejects unresolved hand use. Legacy read behavior remains
  covered by the existing projection/handler suites.
- Delivery loss/reload: browser fixtures save the action but abort its response,
  reload with the retained action ID, reconcile the saved receipt and equipment
  selector, and assert exactly one confirmation and no narrative action.

These are isolated deterministic fixtures. They do not establish production
runtime, physical-phone behavior, wider armor/magic rules, or hosted CI status.

## Local validation

- `go test -short -race ./...`: 2,215 tests passed on base `76c4ce6` (includes merged #346/#348 catalogue contracts).
- `go vet ./...`: passed.
- `go test -tags integration -run 'TestIntegration(EquipmentAction|ConsumableAction|ActionPreviewReadOnly)' -count=1 -race ./internal/handlers`: three tests passed in disposable PostgreSQL.
- Frontend production build and changed-file ESLint: passed; build retains the
  existing large-chunk warning.
- `pnpm test`: 66 tests passed across 17 files.
- Playwright equipment, action-preview and caster-lifecycle suites: 12 desktop/
  phone fixtures passed, including saved-response-loss reload recovery.
- Opt-in `TestPinnedArtifact`: passed with the specified local PDF.

Additional Go lint could not complete: no golangci-lint is installed for active
Go 1.27.3; the installed binary cannot read its export data. Running the installed
Go 1.27.0/linter pair instead panics inside Staticcheck's analysis of standard
library packages. This is a validation gap, not a passing lint result.
