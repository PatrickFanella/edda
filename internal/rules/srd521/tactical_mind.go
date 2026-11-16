package srd521

import "fmt"

// TacticalMind identifies the explicit post-failure choice, not a new check.
const TacticalMind = "tactical-mind"

// TacticalMindReceipt preserves the original roll and the inspected cost.
// The private difficulty never crosses this transport.
type TacticalMindReceipt struct {
	RulesVersion     string       `json:"rules_version"`
	Kind             string       `json:"kind"`
	CheckID          string       `json:"check_id"`
	Base             CheckReceipt `json:"base"`
	Die              int          `json:"die"`
	Total            int          `json:"total"`
	Outcome          string       `json:"outcome"`
	SecondWindBefore int          `json:"second_wind_before"`
	SecondWindAfter  int          `json:"second_wind_after"`
	SourceIDs        []string     `json:"source_ids"`
}

// ResolveTacticalMind adds one d10 without rerolling or healing. A failed
// attempt consumes the window but does not expend Second Wind (SRD p.48).
func ResolveTacticalMind(base CheckReceipt, dc, uses int, choice string, dice Dice) (*TacticalMindReceipt, error) {
	if base.Kind != AbilityCheck || base.Mode != "ordinary" || base.Outcome != "failure" || base.Total >= dc || uses < 1 || uses > 2 || (choice != "use" && choice != "decline") || len(base.Dice) != 1 || base.Dice[0].Sides != 20 || !base.Dice[0].Kept {
		return nil, fmt.Errorf("ordinary failed ability check and available Second Wind required")
	}
	r := &TacticalMindReceipt{RulesVersion: "srd-5.2.1", Kind: TacticalMind, Base: base, Total: base.Total, Outcome: "declined", SecondWindBefore: uses, SecondWindAfter: uses, SourceIDs: []string{prefix + "fighter/tactical-mind", prefix + "fighter/second-wind"}}
	if choice == "decline" {
		return r, nil
	}
	if dice == nil {
		return nil, fmt.Errorf("d10 required")
	}
	r.Die = dice.Intn(10) + 1
	if r.Die < 1 || r.Die > 10 {
		return nil, fmt.Errorf("invalid d10")
	}
	r.Total += r.Die
	r.Outcome = "failure"
	if r.Total >= dc {
		r.Outcome = "success"
		r.SecondWindAfter--
	}
	return r, nil
}
