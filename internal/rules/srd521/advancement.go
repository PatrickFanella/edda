package srd521

import "fmt"

// AdvanceLevel is the bounded Fighter 1 to 2 transition. EntryID selects HP.
const AdvanceLevel = "advance-level"

// Advancement records the HP decision and newly acquired, read-only features.
type Advancement struct {
	Choice string `json:"choice"`
	Die    int    `json:"die"`
	HPGain int    `json:"hp_gain"`
}

// AdvancementReceipt is saved with the action so a retry cannot reroll HP.
type AdvancementReceipt struct {
	RulesVersion string `json:"rules_version"`
	Kind         string `json:"kind"`
	Advancement
	LevelBefore int      `json:"level_before"`
	LevelAfter  int      `json:"level_after"`
	MaxHPBefore int      `json:"max_hp_before"`
	MaxHPAfter  int      `json:"max_hp_after"`
	SourceIDs   []string `json:"source_ids"`
}

// AdvanceFighter applies only the supported transition; XP is earned separately.
// It retains Defense and the three recorded mastery choices, without healing.
func AdvanceFighter(b *Build, sheet *CharacterSheet, choice string, dice Dice) (*AdvancementReceipt, error) {
	if b == nil || sheet == nil || b.Class != "fighter" || b.Advancement != nil || sheet.Progression.CurrentXP < 300 || (choice != "fixed" && choice != "roll") {
		return nil, fmt.Errorf("fighter level 1, 300 earned XP and an explicit HP choice required")
	}
	die := 6
	if choice == "roll" {
		if dice == nil {
			return nil, fmt.Errorf("HP dice required")
		}
		die = dice.Intn(10) + 1
		if die < 1 || die > 10 {
			return nil, fmt.Errorf("invalid HP die")
		}
	}
	a := &Advancement{Choice: choice, Die: die, HPGain: max(1, die+sheet.Abilities["constitution"].Modifier.Value)}
	b.Advancement = a
	b.HitDice++
	return &AdvancementReceipt{RulesVersion: "srd-5.2.1", Kind: AdvanceLevel, Advancement: *a, LevelBefore: 1, LevelAfter: 2, MaxHPBefore: sheet.MaxHP.Value, MaxHPAfter: sheet.MaxHP.Value + a.HPGain, SourceIDs: []string{prefix + "character/level-advancement", prefix + "fighter/action-surge", prefix + "fighter/tactical-mind"}}, nil
}
