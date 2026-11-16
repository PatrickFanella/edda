package srd521

import "fmt"

const (
	EndConcentration = "concentration-end"
	WaitMinute       = "wait-minute"
	ShortRest        = "rest-short"
	LongRest         = "rest-long"
)

// Concentration is the single supported self effect. Absolute campaign seconds
// survive reloads; the action ID identifies precisely which effect was removed.
type Concentration struct {
	ID        string `json:"id"`
	Spell     string `json:"spell"`
	StartedAt int    `json:"started_at"`
	ExpiresAt int    `json:"expires_at"`
}

// EffectReceipt is a public, saved account of timing, costs and required checks.
type EffectReceipt struct {
	ActionSurgeBefore int            `json:"action_surge_before,omitempty"`
	ActionSurgeAfter  int            `json:"action_surge_after,omitempty"`
	RulesVersion      string         `json:"rules_version"`
	Kind              string         `json:"kind"`
	Reason            string         `json:"reason"`
	Before            *Concentration `json:"before,omitempty"`
	After             *Concentration `json:"after,omitempty"`
	TimeBefore        int            `json:"time_before"`
	TimeAfter         int            `json:"time_after"`
	SlotsBefore       int            `json:"slots_before"`
	SlotsAfter        int            `json:"slots_after"`
	HPBefore          int            `json:"hp_before"`
	HPAfter           int            `json:"hp_after"`
	HitDiceBefore     int            `json:"hit_dice_before"`
	HitDiceAfter      int            `json:"hit_dice_after"`
	HealingDie        int            `json:"healing_die,omitempty"`
	Check             *CheckReceipt  `json:"check,omitempty"`
	SourceIDs         []string       `json:"source_ids"`
}

func IsLifecycle(kind string) bool {
	return kind == EndConcentration || kind == WaitMinute || kind == ShortRest || kind == LongRest
}

// Lifecycle mutates a detached build under the caller's transaction. Rest means
// explicitly confirmed uninterrupted downtime; interruptions are not simulated.
func Lifecycle(b *Build, sheet *CharacterSheet, kind, entry, actionID string, now int, dice Dice) (*EffectReceipt, error) {
	if b != nil && b.Class == "fighter" {
		return RestActionSurge(b, sheet, kind, entry, now)
	}
	if b == nil || b.Class != "cleric" || b.Casting == nil || sheet == nil || sheet.Spellcasting == nil || sheet.HP.Value <= 0 || now < 0 {
		return nil, fmt.Errorf("conscious supported caster and campaign clock required")
	}
	c := b.Casting
	r := &EffectReceipt{RulesVersion: "srd-5.2.1", Kind: kind, Before: c.Concentration, TimeBefore: now, TimeAfter: now, SlotsBefore: c.Slots, SlotsAfter: c.Slots, HPBefore: sheet.HP.Value, HPAfter: sheet.HP.Value, HitDiceBefore: b.HitDice, HitDiceAfter: b.HitDice, SourceIDs: []string{prefix + "rules-glossary/concentration"}}
	switch kind {
	case SpellCast:
		if entry != "shield-of-faith" || !c.ShieldPrepared || !c.Focus || !c.Verbal || !c.Somatic || c.Slots <= 0 || actionID == "" {
			return nil, fmt.Errorf("Shield of Faith or slot unavailable")
		}
		c.Concentration = &Concentration{ID: actionID, Spell: "shield-of-faith", StartedAt: now, ExpiresAt: now + 600}
		c.Slots--
		r.Reason = "Shield of Faith on self; +2 AC; one level 1 slot and Bonus Action."
		r.SourceIDs = append(r.SourceIDs, prefix+"spells/shield-of-faith", prefix+"cleric/spellcasting")
	case EndConcentration:
		if entry != "" || c.Concentration == nil {
			return nil, fmt.Errorf("no concentration to end")
		}
		c.Concentration = nil
		r.Reason = "Concentration ended voluntarily; no action or slot cost."
	case WaitMinute:
		if entry != "" {
			return nil, fmt.Errorf("unsupported wait")
		}
		r.TimeAfter += 60
		r.Reason = "One minute of quiet downtime elapsed."
	case ShortRest:
		if entry != "" && entry != "hit-die" {
			return nil, fmt.Errorf("unsupported rest choice")
		}
		r.TimeAfter += 3600
		r.Reason = "Uninterrupted one-hour Short Rest; no spell slots recovered."
		r.SourceIDs = append(r.SourceIDs, prefix+"rules-glossary/short-rest")
		if entry == "hit-die" {
			if b.HitDice < 1 || dice == nil {
				return nil, fmt.Errorf("no Hit Die available")
			}
			v := dice.Intn(8)
			if v < 0 || v >= 8 {
				return nil, fmt.Errorf("invalid d8")
			}
			r.HealingDie = v + 1
			b.HitDice--
			r.HPAfter = min(sheet.MaxHP.Value, r.HPBefore+max(0, v+1+sheet.Abilities["constitution"].Modifier.Value))
		}
	case LongRest:
		if entry != "" || (c.LastLongRest != nil && now-*c.LastLongRest < 16*3600) {
			return nil, fmt.Errorf("wait 16 hours after the previous Long Rest before starting another")
		}
		r.TimeAfter += 8 * 3600
		c.LastLongRest = &r.TimeAfter
		c.Slots = 2
		b.HitDice = 1
		c.Concentration = nil
		if c.Choices != nil {
			c.Choices.FeatUses = 1
		}
		r.HPAfter = sheet.MaxHP.Value
		r.Reason = "Uninterrupted eight-hour Long Rest, including six hours of sleep; HP, Hit Die, Cleric slots and recorded Magic Initiate use recovered."
		r.SourceIDs = append(r.SourceIDs, prefix+"rules-glossary/long-rest", prefix+"cleric/spellcasting", prefix+"feats/magic-initiate")
	default:
		return nil, fmt.Errorf("unsupported lifecycle action")
	}
	if c.Concentration != nil && c.Concentration.ExpiresAt <= r.TimeAfter {
		c.Concentration = nil
		r.Reason += " Shield of Faith expired."
	}
	r.After = c.Concentration
	r.SlotsAfter = c.Slots
	r.HitDiceAfter = b.HitDice
	return r, nil
}

// ConcentrationDamage receives actual resolved damage, never client input.
// Zero HP ends concentration without rolling; each other damage source rolls
// exactly one Constitution save at min(30,max(10,floor(damage/2))).
func ConcentrationDamage(b *Build, sheet *CharacterSheet, damage, hp, now int, dice Dice) (*EffectReceipt, error) {
	if b.Casting == nil || b.Casting.Concentration == nil || damage <= 0 {
		return nil, nil
	}
	c := b.Casting
	r := &EffectReceipt{RulesVersion: "srd-5.2.1", Kind: "concentration-damage", Before: c.Concentration, TimeBefore: now, TimeAfter: now, HPBefore: sheet.HP.Value, HPAfter: hp, SlotsBefore: c.Slots, SlotsAfter: c.Slots, HitDiceBefore: b.HitDice, HitDiceAfter: b.HitDice, SourceIDs: []string{prefix + "rules-glossary/concentration"}}
	r.Reason = "Concentration ended at zero HP."
	if hp <= 0 {
		c.Concentration = nil
	} else {
		check, err := ResolveCheck(sheet, Check{Kind: SavingThrow, Ability: "constitution"}, Difficulty{Value: min(30, max(10, damage/2)), PlayerKnown: true, OutcomeKnown: true}, dice)
		if err != nil {
			return nil, err
		}
		receipt := check.PlayerReceipt()
		r.Check = &receipt
		r.Reason = "Concentration maintained after damage."
		if !check.Success() {
			c.Concentration = nil
			r.Reason = "Concentration lost after damage."
		}
	}
	r.After = c.Concentration
	return r, nil
}
