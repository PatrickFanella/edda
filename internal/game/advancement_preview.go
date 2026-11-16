package game

import (
	"context"
	"git.subcult.tv/subculture-collective/edda/internal/domain"
	"git.subcult.tv/subculture-collective/edda/internal/rules/srd521"
)

func prepareAdvancement(ctx context.Context, campaign domain.Campaign, pc domain.PlayerCharacter, intent ActionIntent) (*StoredAction, error) {
	p, err := PrepareCharacterAction(ctx, campaign, pc, ActionIntent{ActorID: intent.ActorID, Kind: srd521.AbilityCheck, Ability: "strength"})
	if err != nil {
		return nil, err
	}
	p.Intent = intent
	result := &StoredAction{Preview: p}
	if !p.Available {
		return result, nil
	}
	p.Available = false
	p.Modifier = nil
	p.Reason = "Advance only a level 1 Fighter with 300 earned XP; choose fixed or rolled HP."
	sheet, err := srd521.Project(pc.Stats, pc.Level, pc.HP, pc.MaxHP, pc.Experience)
	if err != nil || sheet.Spellcasting != nil || pc.Level != 1 || pc.Experience < 300 || intent.Ability != "" || intent.Skill != "" || intent.SlotLevel != 0 || (intent.EntryID != "fixed" && intent.EntryID != "roll") {
		return result, nil
	}
	p.Available, p.Reason = true, ""
	p.Options = []string{"Fighter 1 to 2; retain Defense and three weapon mastery choices. Gain one d10 Hit Die; keep current HP and spent Second Wind uses. Action Surge (one use; extra Action except Magic on your encounter turn) is available. Tactical Mind effect remains unavailable."}
	if intent.EntryID == "fixed" {
		p.Options = append(p.Options, "Fixed HP: 6 + Constitution (minimum 1) added to maximum HP.")
	} else {
		p.Options = append(p.Options, "Rolled HP: d10 + Constitution (minimum 1); roll only on confirmation and save the result.")
	}
	p.SourceIDs = []string{"srd-5.2.1:content/character/level-advancement", "srd-5.2.1:content/fighter/action-surge", "srd-5.2.1:content/fighter/tactical-mind"}
	p.CostStatus = "fixture-only"
	return result, nil
}
