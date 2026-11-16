package game

import (
	"context"
	"encoding/json"

	"git.subcult.tv/subculture-collective/edda/internal/domain"
	"git.subcult.tv/subculture-collective/edda/internal/rules/srd521"
	statedb "git.subcult.tv/subculture-collective/edda/internal/state/sqlc"
	"github.com/google/uuid"
)

// ReadSpellClock reads absolute campaign seconds. A pre-clock test/save starts
// at day 1 08:00, the same default used by campaign creation. Reading is inert.
func ReadSpellClock(ctx context.Context, database statedb.DBTX, campaignID uuid.UUID) (int, error) {
	var now int
	err := database.QueryRow(ctx, `SELECT COALESCE((SELECT (day-1)*86400+hour*3600+minute*60+second FROM campaign_time WHERE campaign_id=$1),28800)`, campaignID).Scan(&now)
	return now, err
}

func prepareLifecycle(ctx context.Context, campaign domain.Campaign, pc domain.PlayerCharacter, intent ActionIntent) (*StoredAction, error) {
	p, err := PrepareCharacterAction(ctx, campaign, pc, ActionIntent{ActorID: intent.ActorID, Kind: srd521.AbilityCheck, Ability: "wisdom"})
	if err != nil {
		return nil, err
	}
	p.Intent = intent
	r := &StoredAction{Preview: p}
	if !p.Available {
		return r, nil
	}
	p.Available = false
	p.Reason = "Choose a supported caster timing or rest action."
	if intent.Ability != "" || intent.Skill != "" || intent.SlotLevel != 0 {
		return r, nil
	}
	sheet, err := srd521.Project(pc.Stats, pc.Level, pc.HP, pc.MaxHP, pc.Experience)
	if err != nil {
		return r, nil
	}
	var envelope struct {
		Build srd521.Build `json:"srd521"`
	}
	if err := json.Unmarshal(pc.Stats, &envelope); err != nil {
		return nil, err
	}
	if sheet.Spellcasting == nil {
		if !srd521.SupportsActionSurge(pc.Stats, pc.Level) || intent.EntryID != "" || (intent.Kind != srd521.ShortRest && intent.Kind != srd521.LongRest) {
			return r, nil
		}
		p.Available, p.Reason, p.Modifier = true, "", nil
		p.Options = []string{"Confirm uninterrupted downtime: Short Rest one hour; Long Rest eight hours including six hours sleep. Recover Action Surge; Short Rest recovers one Second Wind, Long Rest recovers HP, Hit Dice and all Second Wind."}
		p.SourceIDs = []string{"srd-5.2.1:content/fighter/action-surge", "srd-5.2.1:content/rules-glossary/short-rest", "srd-5.2.1:content/rules-glossary/long-rest"}
		p.CostStatus = "fixture-only"
		return r, nil
	}
	switch intent.Kind {
	case srd521.EndConcentration:
		if intent.EntryID != "" || sheet.Concentration == nil {
			p.Reason = "No concentration effect to end."
			return r, nil
		}
		p.Options = []string{"End the current concentration effect; no action required."}
	case srd521.WaitMinute:
		if intent.EntryID != "" {
			return r, nil
		}
		p.Options = []string{"One minute of quiet downtime; no recovery."}
	case srd521.ShortRest:
		if intent.EntryID != "" && intent.EntryID != "hit-die" {
			return r, nil
		}
		if intent.EntryID == "hit-die" && envelope.Build.HitDice < 1 {
			p.Reason = "No d8 Hit Die remains."
			return r, nil
		}
		p.Options = []string{"Confirm one uninterrupted hour of light activity. No spell slots recovered."}
		if intent.EntryID == "hit-die" {
			p.Options = append(p.Options, "Spend one d8 Hit Die: d8 + Constitution healing capped at maximum HP.")
			p.Costs = []ActionCost{{ResourceID: "hit-dice", Amount: 1}}
		}
	case srd521.LongRest:
		if intent.EntryID != "" {
			return r, nil
		}
		p.Options = []string{"Confirm eight uninterrupted hours including six hours of sleep. Recover HP, one Hit Die, level 1 slots and the recorded Magic Initiate use; sleep ends concentration. Wait 16 hours after the previous Long Rest before starting another."}
	default:
		return r, nil
	}
	p.Available = true
	p.Reason = ""
	p.Modifier = nil
	p.CostStatus = "fixture-only"
	p.SourceIDs = []string{"srd-5.2.1:content/rules-glossary/concentration", "srd-5.2.1:content/rules-glossary/short-rest", "srd-5.2.1:content/rules-glossary/long-rest"}
	return r, nil
}
