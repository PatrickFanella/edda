package game

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"git.subcult.tv/subculture-collective/edda/internal/combat"
	"git.subcult.tv/subculture-collective/edda/internal/dbutil"
	"git.subcult.tv/subculture-collective/edda/internal/domain"
	"git.subcult.tv/subculture-collective/edda/internal/rules/srd521"
	statedb "git.subcult.tv/subculture-collective/edda/internal/state/sqlc"
	"github.com/google/uuid"
)

// ReadEncounter restores the latest durable fixture, including completed state.
// The indexed campaign history query has no recent-turn cutoff.
func ReadEncounter(ctx context.Context, database statedb.DBTX, campaignID uuid.UUID) (*combat.Encounter, error) {
	var data []byte
	err := database.QueryRow(ctx, `SELECT COALESCE((SELECT call->'Result'->'encounter' FROM session_logs l CROSS JOIN LATERAL jsonb_array_elements(COALESCE(l.tool_calls,'[]'::jsonb)) WITH ORDINALITY AS t(call,position) WHERE l.campaign_id=$1 AND call->'Result' ? 'encounter' ORDER BY l.turn_number DESC,position DESC LIMIT 1),'null'::jsonb)`, dbutil.ToPgtype(campaignID)).Scan(&data)
	if err != nil {
		return nil, err
	}
	var encounter *combat.Encounter
	if err := json.Unmarshal(data, &encounter); err != nil {
		return nil, err
	}
	if encounter != nil && (len(encounter.Actors) != 2 || encounter.Turn < 0 || encounter.Turn >= 2 || encounter.Round < 1) {
		return nil, fmt.Errorf("invalid saved encounter")
	}
	return encounter, nil
}

// PrepareStoredAction adds a durable encounter fingerprint and enforces turn
// ownership on the same path used by confirmation. Preparation never rolls.
func PrepareStoredAction(ctx context.Context, database statedb.DBTX, campaign domain.Campaign, pc domain.PlayerCharacter, intent ActionIntent, lock bool) (*StoredAction, error) {
	base := intent
	switch intent.Kind {
	case combat.EncounterStart:
		base.Kind = srd521.MartialAttack
	case combat.EncounterEndTurn, combat.EncounterMove, combat.EncounterDash:
		base = ActionIntent{ActorID: intent.ActorID, Kind: srd521.AbilityCheck, Ability: "strength"}
	}
	result, err := prepareStoredActionBase(ctx, database, campaign, pc, base, lock)
	if err != nil {
		return nil, err
	}
	p := result.Preview
	if base != intent {
		p.Intent = intent
	}
	encounter, err := ReadEncounter(ctx, database, campaign.ID)
	if err != nil {
		return nil, err
	}
	result.Encounter = encounter
	p.Encounter = encounter
	if encounter != nil {
		encoded, err := json.Marshal([]any{p.ExpectedRevision, encounter})
		if err != nil {
			return nil, err
		}
		p.ExpectedRevision = fmt.Sprintf("encounter-v1:%x", sha256.Sum256(encoded))
	}
	unavailable := func(reason string) (*StoredAction, error) { p.Available = false; p.Reason = reason; return result, nil }
	if !p.Available {
		return result, nil
	}
	active := encounter != nil && encounter.Status == "active"
	if intent.Kind == combat.EncounterStart {
		if active {
			return unavailable("An encounter is already active.")
		}
		p.Options = []string{"Two-actor ordinary melee fixture; NPC Unarmed Strike +2, 1 bludgeoning; speed 30 feet; tied initiative: player first. Confirm to roll initiative."}
		p.Costs = []ActionCost{}
		p.CostStatus = "encounter"
		return result, nil
	}
	if !active {
		switch intent.Kind {
		case combat.EncounterEndTurn, combat.EncounterMove, combat.EncounterDash:
			return unavailable("No supported encounter is active.")
		}
		return result, nil
	}
	if encounter.Active().ID != pc.ID {
		return unavailable("Wait for the active actor's turn.")
	}
	// Do not pretend arbitrary checks are encounter actions in this slice.
	switch intent.Kind {
	case srd521.MartialAttack:
		allowed := []ActionTarget{}
		targets := []MartialTarget{}
		for _, t := range result.Targets {
			for _, a := range encounter.Actors {
				if !a.Player && a.ID == t.ID {
					allowed = append(allowed, ActionTarget{ID: t.ID, Label: t.Name})
					targets = append(targets, t)
				}
			}
		}
		result.Targets = targets
		p.Targets = allowed
		if len(targets) == 0 {
			return unavailable("The encounter target is no longer available.")
		}
	case srd521.ConsumableUse:
	case combat.EncounterMove, combat.EncounterDash, combat.EncounterEndTurn:
		if intent.EntryID != "" || intent.Ability != "" || intent.Skill != "" {
			return unavailable("Choose a supported encounter action.")
		}
		p.Modifier = nil
		p.Targets = []ActionTarget{}
		p.Costs = []ActionCost{}
		p.Options = []string{"End turn resolves the NPC turn through saved dice mechanics."}
		if intent.Kind == combat.EncounterMove {
			p.Costs = []ActionCost{{ResourceID: "movement-feet", Amount: 5}}
			p.Options = []string{"Spend 5 feet of movement; position and range changes require adjudication."}
		}
		if intent.Kind == combat.EncounterDash {
			p.Costs = []ActionCost{{ResourceID: "action", Amount: 1}}
			p.Options = []string{"Dash adds 30 feet of movement this turn."}
		}
	default:
		return unavailable("Use a supported encounter action or end your turn.")
	}
	// Spend only on a copy so previews remain read-only.
	copy := *encounter
	if err := copy.Spend(pc.ID, intent.Kind); err != nil {
		return unavailable(err.Error())
	}
	p.CostStatus = "encounter"
	p.SourceIDs = append(p.SourceIDs, "srd-5.2.1:content/combat/initiative", "srd-5.2.1:content/combat/your-turn")
	return result, nil
}
