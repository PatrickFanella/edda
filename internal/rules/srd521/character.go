package srd521

import (
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"strings"
)

// Build is an explicit read-only fixture contract, not a character creator.
// Stored beneath stats.srd521, it never interprets legacy free-form stats.
type Build struct {
	Advancement   *Advancement   `json:"advancement,omitempty"`
	Casting       *CastingBuild  `json:"casting,omitempty"`
	FightingStyle string         `json:"fighting_style"`
	Class         string         `json:"class"`
	Abilities     map[string]int `json:"abilities"`
	Skills        []string       `json:"skills"`
	Expertise     []string       `json:"expertise,omitempty"`
	Armor         string         `json:"armor"`
	Shield        bool           `json:"shield"`
	SecondWind    int            `json:"second_wind"`
	HitDice       int            `json:"hit_dice"`
}

type Term struct {
	Label    string `json:"label"`
	Value    int    `json:"value"`
	SourceID string `json:"source_id"`
}
type Value struct {
	Value      int    `json:"value"`
	Derivation []Term `json:"derivation"`
}
type Ability struct {
	Score    int   `json:"score"`
	Modifier Value `json:"modifier"`
}
type Resource struct {
	ID       string `json:"id"`
	Label    string `json:"label"`
	Current  int    `json:"current"`
	Maximum  int    `json:"maximum"`
	SourceID string `json:"source_id"`
}
type Capability struct {
	ID       string `json:"id"`
	Label    string `json:"label"`
	SourceID string `json:"source_id"`
	Playable bool   `json:"playable"`
}
type Progression struct {
	CurrentXP   int    `json:"current_xp"`
	LevelXP     int    `json:"level_xp"`
	NextLevelXP int    `json:"next_level_xp"`
	SourceID    string `json:"source_id"`
}

// CharacterSheet is derived on the server; clients display values and sources.
type CharacterSheet struct {
	Coverage      CoverageReport     `json:"coverage"`
	Concentration *Concentration     `json:"concentration,omitempty"`
	RulesVersion  string             `json:"rules_version"`
	Scope         string             `json:"scope"`
	Abilities     map[string]Ability `json:"abilities"`
	Proficiency   Value              `json:"proficiency"`
	Saves         map[string]Value   `json:"saves"`
	Skills        map[string]Value   `json:"skills"`
	HP            Value              `json:"hp"`
	MaxHP         Value              `json:"max_hp"`
	Defenses      map[string]Value   `json:"defenses"`
	Resources     []Resource         `json:"resources"`
	Capabilities  []Capability       `json:"capabilities"`
	Progression   Progression        `json:"progression"`
	Spellcasting  *Spellcasting      `json:"spellcasting,omitempty"`
}

var abilityNames = []string{"strength", "dexterity", "constitution", "intelligence", "wisdom", "charisma"}
var skillAbilities = map[string]string{
	"acrobatics": "dexterity", "animal-handling": "wisdom", "arcana": "intelligence", "athletics": "strength", "deception": "charisma", "history": "intelligence", "insight": "wisdom", "intimidation": "charisma", "investigation": "intelligence", "medicine": "wisdom", "nature": "intelligence", "perception": "wisdom", "performance": "charisma", "persuasion": "charisma", "religion": "intelligence", "sleight-of-hand": "dexterity", "stealth": "dexterity", "survival": "wisdom",
}

const prefix = "srd-5.2.1:content/"

func derived(terms ...Term) Value {
	v := Value{Derivation: terms}
	for _, t := range terms {
		v.Value += t.Value
	}
	return v
}
func term(label string, value int, source string) Term { return Term{label, value, prefix + source} }

// Project supports the level-one Fighter and unarmored Cleric casting fixtures.
// It rejects unsupported input rather than manufacturing unearned features.
func Project(stats []byte, level, hp, maxHP, xp int) (*CharacterSheet, error) {
	var envelope struct {
		Build *Build `json:"srd521"`
	}
	if err := json.Unmarshal(stats, &envelope); err != nil {
		return nil, fmt.Errorf("invalid SRD build data")
	}
	b := envelope.Build
	if b == nil || (level != 1 && (level != 2 || b.Class != "fighter" || b.Advancement == nil)) || (level == 1 && b.Advancement != nil) || (b.Class != "fighter" && b.Class != "cleric") || (b.Class == "fighter" && (b.FightingStyle != "defense" || (b.Armor != "chain-mail" && b.Armor != "none"))) || (b.Class == "cleric" && ((b.Armor != "none" && !(b.Armor == "chain-mail" && HeavyArmorTraining(b))) || b.FightingStyle != "" || b.SecondWind != 0)) {
		return nil, fmt.Errorf("supported projection: Fighter or Cleric with supported ordinary equipment and training")
	}
	if len(b.Abilities) != 6 {
		return nil, fmt.Errorf("six explicit ability scores required")
	}
	register, err := Load()
	if err != nil {
		return nil, fmt.Errorf("load SRD coverage: %w", err)
	}
	sheet := &CharacterSheet{Coverage: register.Coverage(), RulesVersion: "srd-5.2.1", Scope: "fighter-1-chain-mail-defense-read-only", Abilities: map[string]Ability{}, Saves: map[string]Value{}, Skills: map[string]Value{}, Defenses: map[string]Value{}}
	sheet.Proficiency = derived(term("Level 1 proficiency", 2, "character/proficiency"))
	for _, name := range abilityNames {
		score, ok := b.Abilities[name]
		if !ok || score < 1 || score > 20 {
			return nil, fmt.Errorf("invalid %s score", name)
		}
		modifier := int(math.Floor(float64(score-10) / 2))
		sheet.Abilities[name] = Ability{score, derived(term("Ability modifier", modifier, "character/ability-modifiers"))}
		terms := []Term{term(name, modifier, "character/ability-modifiers")}
		if (b.Class == "fighter" && (name == "strength" || name == "constitution")) || (b.Class == "cleric" && (name == "wisdom" || name == "charisma")) {
			terms = append(terms, term(classLabel(b.Class)+" saving throw proficiency", 2, b.Class+"/core-traits"))
		}
		sheet.Saves[name] = derived(terms...)
	}
	trained := map[string]bool{}
	for _, skill := range b.Skills {
		if _, ok := skillAbilities[skill]; !ok || trained[skill] {
			return nil, fmt.Errorf("invalid or duplicate skill %s", skill)
		}
		trained[skill] = true
	}
	expertise := map[string]bool{}
	for _, skill := range b.Expertise {
		if !trained[skill] || expertise[skill] {
			return nil, fmt.Errorf("expertise requires a distinct proficient skill: %s", skill)
		}
		expertise[skill] = true
	}
	for skill, ability := range skillAbilities {
		terms := []Term{term(ability, sheet.Abilities[ability].Modifier.Value, "character/ability-modifiers")}
		if expertise[skill] {
			terms = append(terms, term("Recorded skill expertise (2 × proficiency)", 2*sheet.Proficiency.Value, "character/expertise"))
		} else if trained[skill] {
			terms = append(terms, term("Recorded skill proficiency", 2, "character/skill-proficiencies"))
		}
		sheet.Skills[skill] = derived(terms...)
	}
	hitDie := 10
	if b.Class == "cleric" {
		hitDie = 8
	}
	sheet.MaxHP = derived(term(classLabel(b.Class)+" first-level Hit Die", hitDie, b.Class+"/core-traits"), term("constitution", sheet.Abilities["constitution"].Modifier.Value, "character/hit-points"))
	if level == 2 {
		a := b.Advancement
		if xp < 300 || (a.Choice != "fixed" && a.Choice != "roll") || a.Die < 1 || a.Die > 10 || (a.Choice == "fixed" && a.Die != 6) || a.HPGain != max(1, a.Die+sheet.Abilities["constitution"].Modifier.Value) {
			return nil, fmt.Errorf("invalid saved advancement")
		}
		sheet.MaxHP = derived(append(sheet.MaxHP.Derivation, term("Level 2 HP (saved choice)", a.HPGain, "character/level-advancement"))...)
		sheet.Scope = "fighter-2-chain-mail-defense-fixture"
	}
	if maxHP != sheet.MaxHP.Value || hp < 0 || hp > maxHP || xp < 0 || b.SecondWind < 0 || b.SecondWind > 2 || b.HitDice < 0 || b.HitDice > level {
		return nil, fmt.Errorf("invalid level-one vitals or resources")
	}
	sheet.HP = derived(term("Current hit points", hp, "character/hit-points"))
	if b.Class == "fighter" && b.Armor == "none" {
		sheet.Scope = fmt.Sprintf("fighter-%d-unarmored-defense-fixture", level)
	}
	ac := []Term{term("Unarmored", 10, "armor/unarmored"), term("dexterity", sheet.Abilities["dexterity"].Modifier.Value, "character/ability-modifiers")}
	if b.Armor == "chain-mail" {
		ac = []Term{term("Chain mail", 16, "armor/chain-mail")}
		if b.FightingStyle == "defense" {
			ac = append(ac, term("Defense fighting style", 1, "feats/defense"))
		}
	}
	if b.Shield {
		ac = append(ac, term("Shield", 2, "armor/shield"))
	}
	sheet.Defenses["armor-class"] = derived(ac...)
	sheet.Defenses["initiative"] = derived(term("dexterity", sheet.Abilities["dexterity"].Modifier.Value, "combat/initiative"))
	sheet.Resources = []Resource{{"second-wind", "Second Wind", b.SecondWind, 2, prefix + "fighter/second-wind"}, {"hit-dice", "Hit Dice (d10)", b.HitDice, level, prefix + "character/hit-dice"}}
	sheet.Capabilities = []Capability{{"attack", "Attack", prefix + "actions/attack", false}, {"second-wind", "Second Wind", prefix + "fighter/second-wind", false}, {"weapon-mastery", "Weapon Mastery", prefix + "fighter/weapon-mastery", false}}
	if b.Class == "cleric" {
		if b.Casting == nil || b.Casting.Slots < 0 || b.Casting.Slots > 2 {
			return nil, fmt.Errorf("explicit level-one cleric casting resources required")
		}
		sheet.Scope = "cleric-1-cure-wounds-self-fixture"
		sheet.Resources = []Resource{{"spell-slot-1", "Level 1 spell slots", b.Casting.Slots, 2, prefix + "cleric/spellcasting"}, {"hit-dice", "Hit Dice (d8)", b.HitDice, 1, prefix + "character/hit-dice"}}
		sheet.Capabilities = []Capability{}
		sheet.Spellcasting = &Spellcasting{Ability: "wisdom", Modifier: sheet.Abilities["wisdom"].Modifier, Spells: []Spell{{ID: "cure-wounds", Playable: true, Label: "Cure Wounds", Known: true, Prepared: b.Casting.Prepared, Components: []string{"verbal", "somatic"}, ComponentsAvailable: b.Casting.Verbal && b.Casting.Somatic, SlotLevels: []int{1}, SourceID: prefix + "spells/cure-wounds"}}}
	}
	if b.Class == "cleric" && b.Casting.Choices != nil {
		choices := b.Casting.Choices
		expected := (&SupportedBuild{Class: "Cleric"}).CastingChoices()
		expected.FeatUses = choices.FeatUses
		if choices.FeatUses < 0 || choices.FeatUses > 1 || !reflect.DeepEqual(choices, expected) {
			return nil, fmt.Errorf("unsupported caster choices")
		}
		sheet.Resources = append(sheet.Resources, Resource{"magic-initiate", "Magic Initiate (Detect Magic; effect unavailable)", choices.FeatUses, 1, prefix + "feats/magic-initiate"})
		sheet.Spellcasting.Choices = choices
		for _, group := range [][]string{choices.PreparedSpells, choices.Cantrips, choices.FeatCantrips, {choices.FeatSpell}} {
			for _, name := range group {
				if name == "Cure Wounds" {
					continue
				}
				sheet.Spellcasting.Spells = append(sheet.Spellcasting.Spells, Spell{ID: strings.ReplaceAll(strings.ToLower(name), " ", "-"), Label: name, Known: true, Prepared: true, SourceID: prefix + "spells/" + strings.ReplaceAll(strings.ToLower(name), " ", "-")})
			}
		}
	}

	if b.Class == "cleric" {
		if c := b.Casting.Concentration; c != nil {
			if c.ID == "" || c.Spell != "shield-of-faith" || c.StartedAt < 0 || c.ExpiresAt-c.StartedAt != 600 {
				return nil, fmt.Errorf("invalid saved concentration")
			}
			sheet.Concentration = c
			ac := sheet.Defenses["armor-class"]
			sheet.Defenses["armor-class"] = derived(append(ac.Derivation, term("Shield of Faith", 2, "spells/shield-of-faith"))...)
		}
		// The old Cure Wounds fixture can opt into the second spell explicitly.
		found := false
		for i := range sheet.Spellcasting.Spells {
			if sheet.Spellcasting.Spells[i].ID == "shield-of-faith" {
				found = true
				sheet.Spellcasting.Spells[i] = shieldSpell(b.Casting)
			}
		}
		if !found && b.Casting.ShieldPrepared {
			sheet.Spellcasting.Spells = append(sheet.Spellcasting.Spells, shieldSpell(b.Casting))
		}
	}
	sheet.Progression = Progression{xp, 0, 300, prefix + "character/level-advancement"}
	if level == 2 {
		sheet.Progression.LevelXP, sheet.Progression.NextLevelXP = 300, 900
		sheet.Resources = append(sheet.Resources, Resource{"action-surge", "Action Surge (effect unavailable)", 1, 1, prefix + "fighter/action-surge"})
		sheet.Capabilities = append(sheet.Capabilities, Capability{"action-surge", "Action Surge", prefix + "fighter/action-surge", false}, Capability{"tactical-mind", "Tactical Mind", prefix + "fighter/tactical-mind", false})
	}
	return sheet, nil
}

func classLabel(class string) string {
	if class == "cleric" {
		return "Cleric"
	}
	return "Fighter"
}

func shieldSpell(c *CastingBuild) Spell {
	return Spell{ID: "shield-of-faith", Label: "Shield of Faith", Known: true, Prepared: c.ShieldPrepared, Playable: true, Components: []string{"verbal", "somatic", "material (Holy Symbol focus)"}, ComponentsAvailable: c.Verbal && c.Somatic && c.Focus, SlotLevels: []int{1}, SourceID: prefix + "spells/shield-of-faith"}
}
