package combat

import (
	"fmt"
	"github.com/google/uuid"
)

const (
	EncounterReact   = "combat-react"
	EncounterDecline = "combat-decline-reaction"
)

// ReactionTrigger suspends the NPC immediately before its voluntary withdrawal
// leaves five-foot reach. No attack or reaction cost occurs until player consent.
type ReactionTrigger struct {
	ID            uuid.UUID `json:"id"`
	TargetID      uuid.UUID `json:"target_id"`
	WeaponID      uuid.UUID `json:"weapon_id"`
	SpaceRevision string    `json:"space_revision"`
}

// ReactionState tracks the player's reaction separately from the active budget.
type ReactionState struct {
	Available bool             `json:"available"`
	Pending   *ReactionTrigger `json:"pending,omitempty"`
}

// SpendReaction requires the exact pending ordinary opportunity-attack trigger.
// The shared engine validates equipment, current scene and target as well.
func (e *Encounter) SpendReaction() error {
	if e.Status != "active" || e.Active().Player || e.Reaction == nil || !e.Reaction.Available || e.Reaction.Pending == nil || e.Space == nil {
		return fmt.Errorf("No eligible reaction is pending.")
	}
	if e.Space.Revision != e.Reaction.Pending.SpaceRevision {
		return fmt.Errorf("The trigger's spatial ruling changed; refresh or decline.")
	}
	if err := e.Space.Melee(); err != nil {
		return err
	}
	e.Reaction.Available = false
	return nil
}
