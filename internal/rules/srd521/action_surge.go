package srd521

import (
	"encoding/json"
	"fmt"
)

// ActionSurge is the bounded level-two Fighter activation, not an Action cost.
const ActionSurge = "action-surge"

// ActionSurgeReceipt describes the committed single use without private inputs.
type ActionSurgeReceipt struct {
	RulesVersion      string   `json:"rules_version"`
	Kind              string   `json:"kind"`
	UsesBefore        int      `json:"uses_before"`
	UsesAfter         int      `json:"uses_after"`
	AdditionalActions int      `json:"additional_actions"`
	ExcludedAction    string   `json:"excluded_action"`
	SourceIDs         []string `json:"source_ids"`
}

// SupportsActionSurge limits execution to the existing created Human Soldier.
func SupportsActionSurge(stats []byte, level int) bool {
	var e struct {
		Build    *Build `json:"srd521"`
		Creation struct {
			Origin     string `json:"origin"`
			Background string `json:"background"`
		} `json:"srd521_creation"`
	}
	return json.Unmarshal(stats, &e) == nil && level == 2 && e.Build != nil && e.Build.Class == "fighter" && e.Build.Advancement != nil && e.Creation.Origin == "Human" && e.Creation.Background == "Soldier"
}

// ActionSurgeUses treats historical level-two builds as initially unspent.
func ActionSurgeUses(b *Build) int {
	if b == nil || b.Class != "fighter" || b.Advancement == nil || b.ActionSurgeSpent {
		return 0
	}
	return 1
}

// UseActionSurge spends only the build resource. The encounter owns the grant.
func UseActionSurge(b *Build) (*ActionSurgeReceipt, error) {
	if ActionSurgeUses(b) != 1 {
		return nil, fmt.Errorf("Action Surge unavailable")
	}
	b.ActionSurgeSpent = true
	return &ActionSurgeReceipt{RulesVersion: "srd-5.2.1", Kind: ActionSurge, UsesBefore: 1, UsesAfter: 0, AdditionalActions: 1, ExcludedAction: "Magic", SourceIDs: []string{prefix + "fighter/action-surge"}}, nil
}

// RestActionSurge implements only the downtime needed to recover this feature.
// It recovers the existing Fighter resources without simulating interruptions.
func RestActionSurge(b *Build, sheet *CharacterSheet, kind, entry string, now int) (*EffectReceipt, error) {
	if b == nil || b.Class != "fighter" || b.Advancement == nil || sheet == nil || sheet.HP.Value <= 0 || now < 0 || entry != "" || (kind != ShortRest && kind != LongRest) {
		return nil, fmt.Errorf("unsupported Fighter recovery")
	}
	duration := 3600
	if kind == LongRest {
		if b.LastFighterLongRest != nil && now-*b.LastFighterLongRest < 16*3600 {
			return nil, fmt.Errorf("wait 16 hours after the previous Long Rest")
		}
		duration = 8 * 3600
	}
	r := &EffectReceipt{RulesVersion: "srd-5.2.1", Kind: kind, Reason: "Uninterrupted rest; Action Surge and source-valid Fighter resources recovered.", TimeBefore: now, TimeAfter: now + duration, HPBefore: sheet.HP.Value, HPAfter: sheet.HP.Value, HitDiceBefore: b.HitDice, HitDiceAfter: b.HitDice, SourceIDs: []string{prefix + "fighter/action-surge", prefix + "rules-glossary/" + map[string]string{ShortRest: "short-rest", LongRest: "long-rest"}[kind]}}
	r.ActionSurgeBefore = ActionSurgeUses(b)
	b.ActionSurgeSpent = false
	b.SecondWind = min(2, b.SecondWind+1)
	r.ActionSurgeAfter = ActionSurgeUses(b)
	if kind == LongRest {
		b.LastFighterLongRest = &r.TimeAfter
		b.SecondWind = 2
		b.HitDice = 2
		r.HitDiceAfter = 2
		r.HPAfter = sheet.MaxHP.Value
	}
	return r, nil
}
