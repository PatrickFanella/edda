package world

import (
	"fmt"
	"strings"
)

// CharacterAttributes holds the selections from the guided character creation form.
type CharacterAttributes struct {
	Name       string
	Race       string   // e.g. "Human", "Elf", "Dwarf"
	Class      string   // e.g. "Fighter", "Wizard", "Rogue"
	Background string   // e.g. "Noble", "Outlander", "Criminal"
	Alignment  string   // e.g. "Chaotic Good", "Lawful Neutral"
	Traits     []string // 1-3 personality traits
}

// D5ERaces lists the standard D&D 5e races for character creation.
var D5ERaces = []string{
	"Human", "Elf", "Dwarf", "Halfling", "Gnome",
	"Half-Elf", "Half-Orc", "Tiefling", "Dragonborn",
}

// D5EClasses lists the standard D&D 5e classes.
var D5EClasses = []string{
	"Barbarian", "Bard", "Cleric", "Druid", "Fighter", "Monk",
	"Paladin", "Ranger", "Rogue", "Sorcerer", "Warlock", "Wizard",
}

// D5EBackgrounds lists common D&D 5e backgrounds.
var D5EBackgrounds = []string{
	"Acolyte", "Charlatan", "Criminal", "Entertainer", "Folk Hero",
	"Guild Artisan", "Hermit", "Noble", "Outlander", "Sage",
	"Sailor", "Soldier", "Urchin",
}

// D5EAlignments lists the 9 standard alignments.
var D5EAlignments = []string{
	"Lawful Good", "Neutral Good", "Chaotic Good",
	"Lawful Neutral", "True Neutral", "Chaotic Neutral",
	"Lawful Evil", "Neutral Evil", "Chaotic Evil",
}

// classMotivations maps a D&D class to typical character motivations.
var classMotivations = map[string][]string{
	"Barbarian": {"Prove strength in battle", "Protect the tribe"},
	"Bard":      {"Collect epic tales", "Inspire others through art"},
	"Cleric":    {"Serve the divine", "Heal the wounded"},
	"Druid":     {"Preserve the natural order", "Guard the wilds"},
	"Fighter":   {"Protect the weak", "Master the art of combat"},
	"Monk":      {"Achieve inner perfection", "Uphold monastic traditions"},
	"Paladin":   {"Uphold a sacred oath", "Smite the wicked"},
	"Ranger":    {"Guard the frontier", "Hunt dangerous beasts"},
	"Rogue":     {"Seek fortune and thrills", "Uncover hidden truths"},
	"Sorcerer":  {"Understand innate power", "Prove worthy of the gift"},
	"Warlock":   {"Fulfill the pact", "Gain forbidden knowledge"},
	"Wizard":    {"Seek arcane knowledge", "Unlock the mysteries of magic"},
}

// raceStrengths maps a D&D race to inherent strengths.
var raceStrengths = map[string][]string{
	"Human":      {"Versatile adaptability"},
	"Elf":        {"Keen perception", "Grace under pressure"},
	"Dwarf":      {"Unyielding resilience", "Stonework intuition"},
	"Halfling":   {"Remarkable luck", "Nimble evasion"},
	"Gnome":      {"Inventive ingenuity", "Arcane curiosity"},
	"Half-Elf":   {"Social adaptability", "Dual heritage insight"},
	"Half-Orc":   {"Savage endurance", "Intimidating presence"},
	"Tiefling":   {"Infernal resilience", "Force of personality"},
	"Dragonborn": {"Draconic breath weapon", "Commanding presence"},
}

// classStrengths maps a D&D class to trained strengths.
var classStrengths = map[string][]string{
	"Barbarian": {"Raw physical power"},
	"Bard":      {"Silver tongue"},
	"Cleric":    {"Divine channeling"},
	"Druid":     {"Nature magic"},
	"Fighter":   {"Combat expertise"},
	"Monk":      {"Unarmed discipline"},
	"Paladin":   {"Righteous resolve"},
	"Ranger":    {"Wilderness tracking"},
	"Rogue":     {"Stealth and cunning"},
	"Sorcerer":  {"Innate spellcasting"},
	"Warlock":   {"Eldritch invocations"},
	"Wizard":    {"Scholarly spellcraft"},
}

// backgroundWeaknesses maps a D&D background to character weaknesses.
var backgroundWeaknesses = map[string][]string{
	"Acolyte":       {"Naïve idealism", "Blind faith"},
	"Charlatan":     {"Compulsive dishonesty", "Difficulty forming trust"},
	"Criminal":      {"Trust issues", "Haunted by past deeds"},
	"Entertainer":   {"Craves attention", "Fear of being forgotten"},
	"Folk Hero":     {"Stubborn pride", "Unrealistic expectations"},
	"Guild Artisan": {"Obsessive perfectionism", "Material attachment"},
	"Hermit":        {"Social awkwardness", "Distrust of authority"},
	"Noble":         {"Sense of entitlement", "Sheltered worldview"},
	"Outlander":     {"Distrust of civilization", "Blunt to a fault"},
	"Sage":          {"Absent-minded", "Overthinks simple problems"},
	"Sailor":        {"Restless on land", "Rough manners"},
	"Soldier":       {"Rigid discipline", "Haunted by war"},
	"Urchin":        {"Deep-seated insecurity", "Hoarding instinct"},
}

// defaultMotivations is used when the class is not in classMotivations.
var defaultMotivations = []string{"Seek adventure", "Find a place in the world"}

// defaultStrengths is used when race or class is unrecognized.
var defaultStrengths = []string{"Determined spirit"}

// defaultWeaknesses is used when the background is unrecognized.
var defaultWeaknesses = []string{"Untested resolve"}

// BuildCharacterFromAttributes maps D&D character selections to a
// CharacterProfile without any LLM involvement. Unknown race, class, or
// background values fall back to generic defaults. An empty name is left
// empty — the caller is responsible for validation.
func BuildCharacterFromAttributes(attrs CharacterAttributes) *CharacterProfile {
	concept := fmt.Sprintf("%s %s", strings.ToLower(attrs.Race), strings.ToLower(attrs.Class))

	personality := attrs.Alignment
	if len(attrs.Traits) > 0 {
		personality = personality + "; " + strings.Join(attrs.Traits, "; ")
	}

	motivations := lookupStrings(classMotivations, attrs.Class, defaultMotivations)
	strengths := mergeStrengths(attrs.Race, attrs.Class)
	weaknesses := lookupStrings(backgroundWeaknesses, attrs.Background, defaultWeaknesses)

	return &CharacterProfile{
		Name:        attrs.Name,
		Concept:     concept,
		Background:  attrs.Background,
		Personality: personality,
		Motivations: motivations,
		Strengths:   strengths,
		Weaknesses:  weaknesses,
	}
}

// lookupStrings returns the slice for key in m, or the fallback if not found.
func lookupStrings(m map[string][]string, key string, fallback []string) []string {
	if v, ok := m[key]; ok {
		return v
	}
	return fallback
}

// mergeStrengths combines race-derived and class-derived strengths, falling
// back to defaultStrengths when either lookup misses.
func mergeStrengths(race, class string) []string {
	rs := lookupStrings(raceStrengths, race, defaultStrengths)
	cs := lookupStrings(classStrengths, class, defaultStrengths)

	merged := make([]string, 0, len(rs)+len(cs))
	merged = append(merged, rs...)
	merged = append(merged, cs...)
	return merged
}
