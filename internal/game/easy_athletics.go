package game

import (
	"context"

	"git.subcult.tv/subculture-collective/edda/internal/domain"
	"git.subcult.tv/subculture-collective/edda/internal/rules/srd521"
	statedb "git.subcult.tv/subculture-collective/edda/internal/state/sqlc"
)

// EasyAthleticsEntry is Edda's explicitly selected, fixed adjudication fixture.
// SRD p.6 defines Easy as DC 10; it does not adjudicate an arbitrary scene.
const EasyAthleticsEntry = "easy-strength-athletics-dc10"

// EasyAthleticsDC is disclosed in preparation and the saved check receipt.
const EasyAthleticsDC = 10

func prepareEasyAthletics(ctx context.Context, database statedb.DBTX, campaign domain.Campaign, pc domain.PlayerCharacter, intent ActionIntent) (*StoredAction, error) {
	p, err := PrepareCharacterAction(ctx, campaign, pc, ActionIntent{ActorID: intent.ActorID, Kind: srd521.AbilityCheck, Ability: "strength", Skill: "athletics"})
	if err != nil {
		return nil, err
	}
	result := &StoredAction{Preview: p}
	p.Intent = intent
	if !p.Available {
		return result, nil
	}
	p.Available = false
	p.Reason = "Edda's Easy Strength (Athletics) DC 10 fixture requires the owned existing Human Soldier Fighter 2 outside encounters; no custom fields."
	if intent.Kind != srd521.AbilityCheck || intent.EntryID != EasyAthleticsEntry || intent.Ability != "" || intent.Skill != "" || intent.SlotLevel != 0 || !srd521.SupportsActionSurge(pc.Stats, pc.Level) || pc.UserID != campaign.CreatedBy {
		return result, nil
	}
	encounter, err := ReadEncounter(ctx, database, campaign.ID)
	if err != nil {
		return nil, err
	}
	if encounter != nil && encounter.Status == "active" {
		return result, nil
	}
	p.Available = true
	p.Reason = ""
	p.CostStatus = "fixture-only"
	p.Options = []string{"Edda bounded adjudication fixture: Easy Strength (Athletics), disclosed fixed DC 10; one ordinary d20. Failure opens an explicit Tactical Mind choice; no d10 or Second Wind spent by this check. No scene effect or time cost."}
	return result, nil
}
