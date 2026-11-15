package game

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	"git.subcult.tv/subculture-collective/edda/internal/combat"
	"git.subcult.tv/subculture-collective/edda/internal/domain"
	"git.subcult.tv/subculture-collective/edda/internal/rules/srd521"
)

// ActionIntent identifies a sheet selection, never client-supplied modifiers,
// disclosure flags, dice, or model instructions.
type ActionIntent struct {
	ActorID   uuid.UUID `json:"actor_id"`
	Kind      string    `json:"kind"`
	Ability   string    `json:"ability,omitempty"`
	Skill     string    `json:"skill,omitempty"`
	EntryID   string    `json:"entry_id,omitempty"`
	SlotLevel int       `json:"slot_level,omitempty"`
}

// ActionTarget is an explicitly player-known target option.
type ActionTarget struct {
	ID                uuid.UUID `json:"id"`
	Label             string    `json:"label"`
	DistanceFeet      *int      `json:"distance_feet,omitempty"`
	ReachFeet         int       `json:"reach_feet,omitempty"`
	UnavailableReason string    `json:"unavailable_reason,omitempty"`
}

// ActionCost describes a known resource cost. Contextual action/time costs need
// adjudication; an empty list is not a promise of a free in-world action.
type ActionCost struct {
	ResourceID string `json:"resource_id"`
	Amount     int    `json:"amount"`
}

// ActionPreview is read-only preparation, not a reservation or confirmation token.
type ActionPreview struct {
	Encounter        *combat.Encounter   `json:"encounter,omitempty"`
	Intent           ActionIntent        `json:"intent"`
	RulesVersion     domain.RulesVersion `json:"rules_version"`
	ExpectedRevision string              `json:"expected_revision"`
	Available        bool                `json:"available"`
	Reason           string              `json:"reason,omitempty"`
	Modifier         *srd521.Value       `json:"modifier,omitempty"`
	Targets          []ActionTarget      `json:"targets"`
	Options          []string            `json:"options"`
	Costs            []ActionCost        `json:"costs"`
	CostStatus       string              `json:"cost_status"`
	SourceIDs        []string            `json:"source_ids"`
}

// PrepareCharacterAction describes the supported sheet checks from one trusted
// actor/campaign snapshot. It cannot roll, persist, call a model or advance time.
// ExpectedRevision fingerprints only these inputs; confirmation must reload them
// and additionally adjudicate scene, costs and targets before committing (#64).
func PrepareCharacterAction(ctx context.Context, campaign domain.Campaign, pc domain.PlayerCharacter, intent ActionIntent) (*ActionPreview, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	// The hash is opaque: raw stored stats and timestamps never cross this boundary.
	encoded, err := json.Marshal(struct {
		Campaign  domain.Campaign
		Character domain.PlayerCharacter
	}{campaign, pc})
	if err != nil {
		return nil, fmt.Errorf("encode action snapshot: %w", err)
	}
	revision := fmt.Sprintf("actor-v1:%x", sha256.Sum256(encoded))
	preview := &ActionPreview{Intent: intent, RulesVersion: campaign.RulesVersion, ExpectedRevision: revision, Targets: []ActionTarget{}, Options: []string{}, Costs: []ActionCost{}, SourceIDs: []string{}, CostStatus: "requires-adjudication"}
	unavailable := func(reason string) (*ActionPreview, error) { preview.Reason = reason; return preview, nil }
	if pc.ID == uuid.Nil || intent.ActorID != pc.ID || pc.CampaignID != campaign.ID {
		return unavailable("This character action is unavailable.")
	}
	if campaign.RulesVersion != domain.RulesVersionSRD521 {
		return unavailable("Action previews require an SRD 5.2.1 character.")
	}
	if campaign.Status != domain.CampaignStatusActive || pc.Status != "active" || pc.HP <= 0 {
		return unavailable("This character cannot prepare actions in its current state.")
	}
	sheet, err := srd521.Project(pc.Stats, pc.Level, pc.HP, pc.MaxHP, pc.Experience)
	if err != nil {
		return unavailable("Action previews are unavailable for this character build.")
	}
	if intent.EntryID != "" || intent.SlotLevel != 0 {
		return unavailable("This sheet entry does not support action previews yet.")
	}
	if intent.Kind == srd521.AbilityCheck && intent.Skill != "" && intent.Ability == "" {
		intent.Ability, _ = srd521.AbilityForSkill(intent.Skill)
		preview.Intent = intent
	}
	modifier, sources, err := srd521.CheckModifier(sheet, srd521.Check{Kind: intent.Kind, Ability: intent.Ability, Skill: intent.Skill})
	if err != nil {
		return unavailable("Choose a supported ability check, skill check or saving throw.")
	}
	preview.Available = true
	preview.Modifier = &modifier
	preview.SourceIDs = sources
	preview.Options = []string{"ordinary"}
	if intent.Kind == srd521.SavingThrow {
		preview.Targets = []ActionTarget{{ID: pc.ID, Label: pc.Name}}
	}
	return preview, nil
}
