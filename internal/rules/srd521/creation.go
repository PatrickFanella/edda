package srd521

import (
	"encoding/json"
	"fmt"
	"strings"
)

// SupportedMartialBuild is the single character-creation slice supported by
// the SRD 5.2.1 rules path. It names every choice that affects the projection.
type SupportedMartialBuild struct {
	Origin        string
	Class         string
	Background    string
	Abilities     map[string]int
	Proficiencies []string
	Equipment     []string
	OriginFeat    string
	SpeciesFeat   string
	Tool          string
	WeaponMastery []string
	Gold          int
}

// NewSupportedMartialBuild validates the catalogue-backed Human Soldier
// Fighter fixture used during initial SRD character creation.
func NewSupportedMartialBuild(origin, class, background string) (*SupportedMartialBuild, error) {
	if strings.ToLower(strings.TrimSpace(origin)) != "human" ||
		strings.ToLower(strings.TrimSpace(class)) != "fighter" ||
		strings.ToLower(strings.TrimSpace(background)) != "soldier" {
		return nil, fmt.Errorf("SRD 5.2.1 currently supports the Human Soldier Fighter build")
	}
	return &SupportedMartialBuild{
		Origin: "Human", Class: "Fighter", Background: "Soldier",
		// Standard Array fighter assignment (15, 14, 13, 8, 10, 12), then
		// Soldier's +1 to each of Strength, Dexterity, and Constitution.
		Abilities: map[string]int{"strength": 16, "dexterity": 15, "constitution": 14, "intelligence": 8, "wisdom": 10, "charisma": 12},
		// Soldier grants Athletics and Intimidation; Fighter adds Perception
		// and Survival; Human's Skillful trait adds Insight. Human's Versatile
		// trait grants Skilled, selecting History, Medicine, and Nature.
		Proficiencies: []string{"athletics", "intimidation", "perception", "survival", "insight", "history", "medicine", "nature"},
		// Fighter package A plus the Soldier's 50 GP option pays for the
		// longsword and shield used by the narrow action fixture.
		Equipment:     []string{"Chain Mail", "Greatsword", "Flail", "Javelin (8)", "Dungeoneer's Pack", "Longsword", "Shield"},
		OriginFeat:    "Savage Attacker",
		SpeciesFeat:   "Skilled (History, Medicine, Nature)",
		Tool:          "Dice Set",
		WeaponMastery: []string{"Longsword", "Greatsword", "Javelin"},
		Gold:          29,
	}, nil
}

// Stats serializes the creation choices in the exact envelope used by Project.
func (b *SupportedMartialBuild) Stats() ([]byte, error) {
	if b == nil {
		return nil, fmt.Errorf("supported martial build is required")
	}
	return json.Marshal(map[string]any{
		"srd521": Build{Class: "fighter", FightingStyle: "defense", Abilities: b.Abilities, Skills: b.Proficiencies, Armor: "chain-mail", Shield: true, SecondWind: 2, HitDice: 1},
		"srd521_creation": map[string]any{
			"origin": b.Origin, "background": b.Background, "origin_feat": b.OriginFeat,
			"species_feat": b.SpeciesFeat, "tool": b.Tool, "weapon_mastery": b.WeaponMastery,
			"equipment": b.Equipment, "remaining_gold": b.Gold,
		},
	})
}

// MartialBuildForConcept validates the guided concept before world generation.
func MartialBuildForConcept(concept, background string) (*SupportedMartialBuild, error) {
	parts := strings.Fields(concept)
	if len(parts) != 2 {
		return nil, fmt.Errorf("SRD 5.2.1 requires the Human Fighter concept and Soldier background")
	}
	return NewSupportedMartialBuild(parts[0], parts[1], background)
}
