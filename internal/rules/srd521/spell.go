package srd521

import "fmt"

// SpellCast identifies the explicit supported casting slices.
const SpellCast = "spell-cast"

// CastingBuild stores only the resources and component rulings for this fixture.
// Prepared refers to Cure Wounds; it is not a general spell catalogue.
type CastingBuild struct {
	IsolatedRecipientEffects bool           `json:"isolated_recipient_effects,omitempty"`
	Concentration            *Concentration `json:"concentration,omitempty"`
	LastLongRest             *int           `json:"last_long_rest,omitempty"`
	ShieldPrepared           bool           `json:"shield_prepared"`
	Focus                    bool           `json:"focus"`
	Choices                  *CasterChoices `json:"choices,omitempty"`
	Prepared                 bool           `json:"prepared"`
	Slots                    int            `json:"slots"`
	Verbal                   bool           `json:"verbal"`
	Somatic                  bool           `json:"somatic"`
}

// Spell is a player-visible supported spell and its casting requirements.
type Spell struct {
	Playable            bool     `json:"playable"`
	ID                  string   `json:"id"`
	Label               string   `json:"label"`
	Known               bool     `json:"known"`
	Prepared            bool     `json:"prepared"`
	Components          []string `json:"components"`
	ComponentsAvailable bool     `json:"components_available"`
	SlotLevels          []int    `json:"slot_levels"`
	SourceID            string   `json:"source_id"`
}

// Spellcasting exposes server-derived casting ability and the supported spell.
type Spellcasting struct {
	Choices  *CasterChoices `json:"choices,omitempty"`
	Ability  string         `json:"ability"`
	Modifier Value          `json:"modifier"`
	Spells   []Spell        `json:"spells"`
}

// SpellReceipt records the healing roll and exact slot cost for durable replay.
type SpellReceipt struct {
	RulesVersion  string   `json:"rules_version"`
	Kind          string   `json:"kind"`
	Spell         string   `json:"spell"`
	TargetID      string   `json:"target_id"`
	Target        string   `json:"target"`
	SlotLevel     int      `json:"slot_level"`
	SlotsBefore   int      `json:"slots_before"`
	SlotsAfter    int      `json:"slots_after"`
	Dice          []int    `json:"dice"`
	Modifier      Value    `json:"modifier"`
	RolledHealing int      `json:"rolled_healing"`
	Healing       int      `json:"healing"`
	HPBefore      int      `json:"hp_before"`
	HPAfter       int      `json:"hp_after"`
	SourceIDs     []string `json:"source_ids"`
}

// ResolveCureWounds resolves a level-one slot on a conscious self target.
func ResolveCureWounds(sheet *CharacterSheet, dice Dice) (*SpellReceipt, error) {
	if sheet == nil {
		return nil, fmt.Errorf("supported conscious caster required")
	}
	return ResolveCureWoundsTarget(sheet, sheet.HP.Value, sheet.MaxHP.Value, dice)
}

// ResolveCureWoundsTarget rolls healing for an independently validated conscious
// target. Scene eligibility and target identity belong to the shared engine.
func ResolveCureWoundsTarget(sheet *CharacterSheet, hp, maxHP int, dice Dice) (*SpellReceipt, error) {
	if hp <= 0 || maxHP < hp || sheet == nil || sheet.Spellcasting == nil || sheet.Scope != "cleric-1-cure-wounds-self-fixture" || len(sheet.Resources) == 0 || sheet.Resources[0].ID != "spell-slot-1" || sheet.HP.Value <= 0 || sheet.HP.Value > sheet.MaxHP.Value || dice == nil {
		return nil, fmt.Errorf("supported conscious caster required")
	}
	spell := sheet.Spellcasting.Spells[0]
	if !spell.Prepared || !spell.ComponentsAvailable || sheet.Resources[0].Current <= 0 {
		return nil, fmt.Errorf("spell or slot unavailable")
	}
	r := &SpellReceipt{RulesVersion: "srd-5.2.1", Kind: SpellCast, Spell: "Cure Wounds", SlotLevel: 1, SlotsBefore: sheet.Resources[0].Current, SlotsAfter: sheet.Resources[0].Current - 1, Modifier: sheet.Spellcasting.Modifier, HPBefore: hp, SourceIDs: []string{prefix + "spells/cure-wounds", prefix + "cleric/spellcasting", prefix + "damage/healing"}}
	r.RolledHealing = r.Modifier.Value
	for i := 0; i < 2; i++ {
		v := dice.Intn(8)
		if v < 0 || v >= 8 {
			return nil, fmt.Errorf("invalid d8")
		}
		r.Dice = append(r.Dice, v+1)
		r.RolledHealing += v + 1
	}
	r.HPAfter = min(maxHP, r.HPBefore+max(0, r.RolledHealing))
	r.Healing = r.HPAfter - r.HPBefore
	return r, nil
}

// CasterChoices persists the fixed creation catalogue, including recorded but
// unavailable spell effects. It is not inferred from descriptive profile text.
type CasterChoices struct {
	DivineOrder    string   `json:"divine_order"`
	Cantrips       []string `json:"cantrips"`
	PreparedSpells []string `json:"prepared_spells"`
	FeatCantrips   []string `json:"magic_initiate_cantrips"`
	FeatSpell      string   `json:"magic_initiate_spell"`
	FeatAbility    string   `json:"magic_initiate_ability"`
	FeatUses       int      `json:"magic_initiate_uses"`
}
