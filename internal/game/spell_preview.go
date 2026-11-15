package game

import (
	"context"

	"git.subcult.tv/subculture-collective/edda/internal/domain"
	"git.subcult.tv/subculture-collective/edda/internal/rules/srd521"
)

func prepareSpell(ctx context.Context, campaign domain.Campaign, pc domain.PlayerCharacter, intent ActionIntent) (*StoredAction, error) {
	preview, err := PrepareCharacterAction(ctx, campaign, pc, ActionIntent{ActorID: intent.ActorID, Kind: srd521.AbilityCheck, Ability: "wisdom"})
	if err != nil {
		return nil, err
	}
	preview.Intent = intent
	result := &StoredAction{Preview: preview}
	if !preview.Available {
		return result, nil
	}
	preview.Available = false
	preview.Reason = "Choose Cure Wounds with a level 1 slot on the supported Cleric fixture."
	if intent.Ability != "" || intent.Skill != "" || intent.EntryID != "cure-wounds" || intent.SlotLevel != 1 {
		return result, nil
	}
	sheet, err := srd521.Project(pc.Stats, pc.Level, pc.HP, pc.MaxHP, pc.Experience)
	if err != nil || sheet.Spellcasting == nil {
		return result, nil
	}
	preview.Modifier = &sheet.Spellcasting.Modifier
	preview.Targets = []ActionTarget{{ID: pc.ID, Label: pc.Name + " (self; touch)"}}
	preview.Options = []string{"Cure Wounds; prepared Cleric spell; Wisdom; verbal and somatic components; touch (self only); instantaneous; 2d8 + Wisdom healing capped at maximum HP; level 1 slot"}
	preview.Costs = []ActionCost{{ResourceID: "spell-slot-1", Amount: 1}, {ResourceID: "magic-action", Amount: 1}}
	preview.CostStatus = "fixture-only"
	preview.SourceIDs = []string{"srd-5.2.1:content/spells/cure-wounds", "srd-5.2.1:content/cleric/spellcasting"}
	spell := sheet.Spellcasting.Spells[0]
	switch {
	case !spell.Prepared:
		preview.Reason = "Cure Wounds is not prepared."
	case !spell.ComponentsAvailable:
		preview.Reason = "Cure Wounds requires available verbal and somatic components."
	case sheet.Resources[0].Current <= 0:
		preview.Reason = "No level 1 spell slot remains."
	default:
		preview.Available = true
		preview.Reason = ""
	}
	return result, nil
}
