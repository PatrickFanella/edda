package srd521

import "fmt"

const MartialAttack = "martial-attack"

// AttackReceipt describes only the ordinary one-handed longsword fixture.
// It never contains a target's private stats or raw stored properties.
type AttackReceipt struct {
	RulesVersion   string   `json:"rules_version"`
	Kind           string   `json:"kind"`
	Weapon         string   `json:"weapon"`
	TargetID       string   `json:"target_id"`
	Target         string   `json:"target"`
	Die            int      `json:"die"`
	Modifier       Value    `json:"modifier"`
	Total          int      `json:"total"`
	Outcome        string   `json:"outcome"`
	DamageDice     []int    `json:"damage_dice"`
	DamageModifier int      `json:"damage_modifier"`
	Damage         int      `json:"damage"`
	DamageType     string   `json:"damage_type"`
	SourceIDs      []string `json:"source_ids"`
}

// LongswordModifier uses Fighter martial proficiency and Strength.
func LongswordModifier(sheet *CharacterSheet) Value {
	return derived(term("Strength", sheet.Abilities["strength"].Modifier.Value, "character/ability-modifiers"), term("Fighter martial weapon proficiency", sheet.Proficiency.Value, "fighter/core-traits"))
}

// ResolveLongsword handles an ordinary attack with no mastery, conditions,
// resistance or other situational effects. Callers must validate that scope.
func ResolveLongsword(sheet *CharacterSheet, armorClass int, dice Dice) (*AttackReceipt, error) {
	if sheet == nil || sheet.RulesVersion != "srd-5.2.1" || armorClass < 1 || dice == nil {
		return nil, fmt.Errorf("supported sheet, target and dice required")
	}
	roll := dice.Intn(20)
	if roll < 0 || roll >= 20 {
		return nil, fmt.Errorf("invalid d20")
	}
	r := &AttackReceipt{RulesVersion: sheet.RulesVersion, Kind: MartialAttack, Weapon: "Longsword (one hand)", Die: roll + 1, Modifier: LongswordModifier(sheet), Outcome: "miss", DamageType: "slashing", DamageDice: []int{}, SourceIDs: []string{prefix + "d20-tests/attack-rolls", prefix + "weapons/longsword", prefix + "damage/critical-hits"}}
	r.Total = r.Die + r.Modifier.Value
	if r.Die == 1 || (r.Die != 20 && r.Total < armorClass) {
		return r, nil
	}
	r.Outcome = "hit"
	count := 1
	if r.Die == 20 {
		count = 2
		r.Outcome = "critical-hit"
	}
	r.DamageModifier = sheet.Abilities["strength"].Modifier.Value
	r.Damage = r.DamageModifier
	for i := 0; i < count; i++ {
		value := dice.Intn(8)
		if value < 0 || value >= 8 {
			return nil, fmt.Errorf("invalid d8")
		}
		r.DamageDice = append(r.DamageDice, value+1)
		r.Damage += value + 1
	}
	if r.Damage < 0 {
		r.Damage = 0
	}
	return r, nil
}
