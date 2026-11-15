package srd521

import (
	"encoding/json"
	"fmt"
)

// Dice supplies zero-based random values. Implementations shared between turns
// must be safe for concurrent use, as math/rand/v2's top-level functions are.
type Dice interface {
	Intn(n int) int
}

// Check identifies a supported test. Circumstances are adjudicated by the caller;
// no free-form reason, hidden entity or world fact is copied into a receipt.
type Check struct {
	Kind         string
	Ability      string
	Skill        string
	Advantage    bool
	Disadvantage bool
}

// Supported test kinds exclude attacks and death saving throws.
const (
	AbilityCheck = "ability-check"
	SavingThrow  = "saving-throw"
)

// Difficulty is trusted server context, never a client-supplied disclosure flag.
// OutcomeKnown independently controls whether the result can be shown to players.
type Difficulty struct {
	Value        int
	PlayerKnown  bool
	OutcomeKnown bool
}

// CheckDie records each die, including the discarded die of a two-d20 test.
type CheckDie struct {
	Sides int  `json:"sides"`
	Value int  `json:"value"`
	Kept  bool `json:"kept"`
}

// CheckReceipt is the player-safe transport shared by sheet and TUI consumers.
// It has no margin, raw request, narrative, hidden fact or private outcome field.
type CheckReceipt struct {
	RulesVersion string     `json:"rules_version"`
	Kind         string     `json:"kind"`
	Ability      string     `json:"ability"`
	Skill        string     `json:"skill,omitempty"`
	Mode         string     `json:"mode"`
	Dice         []CheckDie `json:"dice"`
	Modifier     Value      `json:"modifier"`
	Total        int        `json:"total"`
	DC           *int       `json:"dc,omitempty"`
	Outcome      string     `json:"outcome"`
	SourceIDs    []string   `json:"source_ids"`
}

// CheckResolution retains a private outcome for server adjudication. JSON
// marshaling emits only PlayerReceipt, never the private outcome.
type CheckResolution struct {
	receipt CheckReceipt
	success bool
}

// Success returns the actual outcome for trusted server consumers.
func (r CheckResolution) Success() bool { return r.success }

// MarshalJSON defaults to the player-safe projection even when a caller tries
// to serialize the whole server resolution.
func (r CheckResolution) MarshalJSON() ([]byte, error) { return json.Marshal(r.PlayerReceipt()) }

// PlayerReceipt returns an independent copy with only explicitly public context.
func (r CheckResolution) PlayerReceipt() CheckReceipt {
	c := r.receipt
	c.Dice = append([]CheckDie(nil), c.Dice...)
	c.Modifier.Derivation = append([]Term(nil), c.Modifier.Derivation...)
	c.SourceIDs = append([]string(nil), c.SourceIDs...)
	if c.DC != nil {
		dc := *c.DC
		c.DC = &dc
	}
	return c
}

// ResolveCheck rolls against the server projection. Skill checks use their
// recorded ability, proficiency and expertise; saves use only save proficiency.
// SRD checks/saves have no automatic natural-1 failure or natural-20 success.
// This is not an attack, death save, reroll or situational-feature resolver.
func ResolveCheck(sheet *CharacterSheet, check Check, difficulty Difficulty, dice Dice) (*CheckResolution, error) {
	if sheet == nil || sheet.RulesVersion != "srd-5.2.1" {
		return nil, fmt.Errorf("SRD character sheet required")
	}
	if dice == nil {
		return nil, fmt.Errorf("dice source required")
	}
	if difficulty.Value < 0 {
		return nil, fmt.Errorf("nonnegative difficulty required")
	}
	ability, ok := sheet.Abilities[check.Ability]
	if !ok {
		return nil, fmt.Errorf("unsupported ability")
	}
	var modifier Value
	source := "d20-tests/ability-checks"
	switch check.Kind {
	case AbilityCheck:
		modifier = ability.Modifier
		if check.Skill != "" {
			if skillAbilities[check.Skill] != check.Ability {
				return nil, fmt.Errorf("unsupported skill and ability pairing")
			}
			modifier, ok = sheet.Skills[check.Skill]
			if !ok {
				return nil, fmt.Errorf("skill projection required")
			}
		}
	case SavingThrow:
		if check.Skill != "" {
			return nil, fmt.Errorf("skills do not apply to saving throws")
		}
		modifier, ok = sheet.Saves[check.Ability]
		if !ok {
			return nil, fmt.Errorf("save projection required")
		}
		source = "d20-tests/saving-throws"
	default:
		return nil, fmt.Errorf("unsupported test kind")
	}
	modifier.Derivation = append([]Term(nil), modifier.Derivation...)
	receipt := CheckReceipt{
		RulesVersion: sheet.RulesVersion, Kind: check.Kind, Ability: check.Ability,
		Skill: check.Skill, Mode: "ordinary", Modifier: modifier, Outcome: "unknown",
		SourceIDs: []string{prefix + source},
	}
	count := 1
	if check.Advantage || check.Disadvantage {
		receipt.SourceIDs = append(receipt.SourceIDs, prefix+"d20-tests/advantage-disadvantage")
	}
	if check.Advantage != check.Disadvantage {
		count = 2
		receipt.Mode = "advantage"
		if check.Disadvantage {
			receipt.Mode = "disadvantage"
		}
	}
	kept := 0
	for i := 0; i < count; i++ {
		roll := dice.Intn(20)
		if roll < 0 || roll >= 20 {
			return nil, fmt.Errorf("dice source returned an invalid d20 value")
		}
		receipt.Dice = append(receipt.Dice, CheckDie{Sides: 20, Value: roll + 1})
		if (receipt.Mode == "advantage" && receipt.Dice[i].Value > receipt.Dice[kept].Value) ||
			(receipt.Mode == "disadvantage" && receipt.Dice[i].Value < receipt.Dice[kept].Value) {
			kept = i
		}
	}
	receipt.Dice[kept].Kept = true
	receipt.Total = receipt.Dice[kept].Value + modifier.Value
	success := receipt.Total >= difficulty.Value
	if difficulty.PlayerKnown {
		dc := difficulty.Value
		receipt.DC = &dc
	}
	if difficulty.OutcomeKnown {
		receipt.Outcome = "failure"
		if success {
			receipt.Outcome = "success"
		}
	}
	return &CheckResolution{receipt: receipt, success: success}, nil
}
