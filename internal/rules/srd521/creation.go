package srd521

import (
	"encoding/json"
	"fmt"
	"strings"
)

// SupportedBuild records fixed character-creation choices supported by
// the SRD 5.2.1 rules path. It names every choice that affects the projection.
type SupportedBuild struct {
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
func NewSupportedMartialBuild(origin, class, background string) (*SupportedBuild, error) {
	if strings.ToLower(strings.TrimSpace(origin)) != "human" ||
		strings.ToLower(strings.TrimSpace(class)) != "fighter" ||
		strings.ToLower(strings.TrimSpace(background)) != "soldier" {
		return nil, fmt.Errorf("SRD 5.2.1 currently supports the Human Soldier Fighter build")
	}
	return &SupportedBuild{
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
func (b *SupportedBuild) Stats() ([]byte, error) {
	if b == nil {
		return nil, fmt.Errorf("supported build is required")
	}
	mechanics := Build{Class: "fighter", FightingStyle: "defense", Abilities: b.Abilities, Skills: b.Proficiencies, Armor: "chain-mail", Shield: true, SecondWind: 2, HitDice: 1}
	if b.Class == "Cleric" {
		mechanics = Build{Class: "cleric", Abilities: b.Abilities, Skills: b.Proficiencies, Armor: "none", HitDice: 1, Casting: &CastingBuild{Prepared: true, Slots: 2, Verbal: true, Somatic: true, Choices: b.CastingChoices()}}
	}
	return json.Marshal(map[string]any{
		"srd521": mechanics,
		"srd521_creation": map[string]any{
			"origin": b.Origin, "background": b.Background, "origin_feat": b.OriginFeat,
			"species_feat": b.SpeciesFeat, "tool": b.Tool, "weapon_mastery": b.WeaponMastery,
			"equipment": b.Equipment, "remaining_gold": b.Gold,
		},
	})
}

// MartialBuildForConcept validates the guided concept before world generation.
func MartialBuildForConcept(concept, background string) (*SupportedBuild, error) {
	parts := strings.Fields(concept)
	if len(parts) != 2 {
		return nil, fmt.Errorf("SRD 5.2.1 requires the Human Fighter concept and Soldier background")
	}
	return NewSupportedMartialBuild(parts[0], parts[1], background)
}

// SupportedBuildForConcept uses the same fixed catalogue for web and TUI setup.
func SupportedBuildForConcept(concept, background string) (*SupportedBuild, error) {
	if strings.EqualFold(strings.TrimSpace(concept), "human cleric") && strings.EqualFold(strings.TrimSpace(background), "acolyte") {
		return &SupportedBuild{Origin: "Human", Class: "Cleric", Background: "Acolyte",
			Abilities:     map[string]int{"strength": 10, "dexterity": 12, "constitution": 14, "intelligence": 9, "wisdom": 16, "charisma": 14},
			Proficiencies: []string{"insight", "religion", "medicine", "persuasion", "perception", "athletics", "survival", "nature"},
			Equipment:     []string{"Calligrapher's Supplies", "Prayer Book", "Holy Symbol", "Parchment (10)", "Robe"},
			OriginFeat:    "Magic Initiate (Cleric; Wisdom)", SpeciesFeat: "Skilled (Athletics, Survival, Nature)", Tool: "Calligrapher's Supplies", Gold: 118,
		}, nil
	}
	build, err := MartialBuildForConcept(concept, background)
	if err != nil {
		return nil, fmt.Errorf("SRD 5.2.1 supports Human Soldier Fighter or Human Acolyte Cleric")
	}
	return build, nil
}

// CastingChoices records valid selections without advertising unimplemented actions.
func (b *SupportedBuild) CastingChoices() *CasterChoices {
	if b.Class != "Cleric" {
		return nil
	}
	return &CasterChoices{DivineOrder: "Protector", Cantrips: []string{"Guidance", "Sacred Flame", "Thaumaturgy"}, PreparedSpells: []string{"Bless", "Cure Wounds", "Guiding Bolt", "Shield of Faith"}, FeatCantrips: []string{"Light", "Mending"}, FeatSpell: "Detect Magic", FeatAbility: "wisdom", FeatUses: 1}
}
