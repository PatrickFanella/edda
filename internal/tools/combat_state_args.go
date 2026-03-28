package tools

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	"github.com/PatrickFanella/game-master/internal/combat"
)

type combatStateJSON struct {
	ID                        string          `json:"id"`
	CampaignID                string          `json:"campaign_id"`
	Combatants                []combatantJSON `json:"combatants"`
	InitiativeOrder           []string        `json:"initiative_order"`
	RoundNumber               int             `json:"round_number"`
	Environment               json.RawMessage `json:"environment"`
	Status                    string          `json:"status"`
	Narrative                 string          `json:"narrative"`
	InitiativeRerollEachRound bool            `json:"initiative_reroll_each_round"`
	TrackDeathSavingThrows    bool            `json:"track_death_saving_throws"`
	SurpriseRoundActive       bool            `json:"surprise_round_active"`
	ActiveEffects             json.RawMessage `json:"active_effects,omitempty"`
}

type combatantJSON struct {
	EntityID          string           `json:"entity_id"`
	EntityType        string           `json:"entity_type"`
	Name              string           `json:"name"`
	HP                int              `json:"hp"`
	MaxHP             int              `json:"max_hp"`
	Stats             json.RawMessage  `json:"stats"`
	Conditions        []conditionJSON  `json:"conditions"`
	Initiative        int              `json:"initiative"`
	Status            string           `json:"status"`
	Surprised         bool             `json:"surprised"`
	DeathSavingThrows *deathSavingJSON `json:"death_saving_throws,omitempty"`
}

type deathSavingJSON struct {
	Successes int `json:"successes"`
	Failures  int `json:"failures"`
}

type conditionJSON struct {
	Name           string `json:"name"`
	DurationRounds int    `json:"duration_rounds"`
}

type environmentJSON struct {
	LocationID  string          `json:"location_id,omitempty"`
	Description string          `json:"description"`
	Properties  json.RawMessage `json:"properties,omitempty"`
}

func parseCombatStateArg(args map[string]any, key string) (*combat.CombatState, error) {
	raw, ok := args[key]
	if !ok {
		return nil, fmt.Errorf("%s is required", key)
	}
	stateMap, ok := raw.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%s must be an object", key)
	}

	jsonBytes, err := json.Marshal(stateMap)
	if err != nil {
		return nil, fmt.Errorf("marshal %s: %w", key, err)
	}

	var csj combatStateJSON
	if err := json.Unmarshal(jsonBytes, &csj); err != nil {
		return nil, fmt.Errorf("parse %s: %w", key, err)
	}

	id, err := uuid.Parse(csj.ID)
	if err != nil {
		return nil, fmt.Errorf("%s.id must be a valid UUID", key)
	}
	campaignID, err := uuid.Parse(csj.CampaignID)
	if err != nil {
		return nil, fmt.Errorf("%s.campaign_id must be a valid UUID", key)
	}

	combatants := make([]combat.Combatant, 0, len(csj.Combatants))
	for i, cj := range csj.Combatants {
		entityID, parseErr := uuid.Parse(cj.EntityID)
		if parseErr != nil {
			return nil, fmt.Errorf("%s.combatants[%d].entity_id must be a valid UUID", key, i)
		}
		conditions := make([]combat.ActiveCondition, len(cj.Conditions))
		for j, condJ := range cj.Conditions {
			conditions[j] = combat.ActiveCondition{
				Name:           condJ.Name,
				DurationRounds: condJ.DurationRounds,
			}
		}
		c := combat.Combatant{
			EntityID:   entityID,
			EntityType: combat.CombatantType(cj.EntityType),
			Name:       cj.Name,
			HP:         cj.HP,
			MaxHP:      cj.MaxHP,
			Stats:      cj.Stats,
			Conditions: conditions,
			Initiative: cj.Initiative,
			Status:     combat.CombatantStatus(cj.Status),
			Surprised:  cj.Surprised,
		}
		if cj.DeathSavingThrows != nil {
			c.DeathSavingThrows = &combat.DeathSavingThrows{
				Successes: cj.DeathSavingThrows.Successes,
				Failures:  cj.DeathSavingThrows.Failures,
			}
		}
		combatants = append(combatants, c)
	}

	initiativeOrder := make([]uuid.UUID, 0, len(csj.InitiativeOrder))
	for i, s := range csj.InitiativeOrder {
		oid, parseErr := uuid.Parse(s)
		if parseErr != nil {
			return nil, fmt.Errorf("%s.initiative_order[%d] must be a valid UUID", key, i)
		}
		initiativeOrder = append(initiativeOrder, oid)
	}

	var env combat.Environment
	if csj.Environment != nil {
		var envObj environmentJSON
		if err := json.Unmarshal(csj.Environment, &envObj); err != nil {
			return nil, fmt.Errorf("%s.environment: %w", key, err)
		}
		env.Description = envObj.Description
		env.Properties = envObj.Properties
		if envObj.LocationID != "" {
			lid, parseErr := uuid.Parse(envObj.LocationID)
			if parseErr != nil {
				return nil, fmt.Errorf("%s.environment.location_id must be a valid UUID", key)
			}
			env.LocationID = &lid
		}
	}

	status := combat.CombatStatus(csj.Status)
	if status == "" {
		status = combat.CombatStatusActive
	}

	return &combat.CombatState{
		ID:                        id,
		CampaignID:                campaignID,
		Combatants:                combatants,
		InitiativeOrder:           initiativeOrder,
		InitiativeRerollEachRound: csj.InitiativeRerollEachRound,
		TrackDeathSavingThrows:    csj.TrackDeathSavingThrows,
		SurpriseRoundActive:       csj.SurpriseRoundActive,
		RoundNumber:               csj.RoundNumber,
		ActiveEffects:             csj.ActiveEffects,
		Environment:               env,
		Status:                    status,
		Narrative:                 csj.Narrative,
	}, nil
}

func combatStateToMap(state *combat.CombatState) map[string]any {
	combatants := make([]map[string]any, 0, len(state.Combatants))
	for _, c := range state.Combatants {
		conditions := make([]map[string]any, 0, len(c.Conditions))
		for _, cond := range c.Conditions {
			conditions = append(conditions, map[string]any{
				"name":            cond.Name,
				"duration_rounds": cond.DurationRounds,
			})
		}
		cm := map[string]any{
			"entity_id":   c.EntityID.String(),
			"entity_type": string(c.EntityType),
			"name":        c.Name,
			"hp":          c.HP,
			"max_hp":      c.MaxHP,
			"stats":       c.Stats,
			"conditions":  conditions,
			"initiative":  c.Initiative,
			"status":      string(c.Status),
			"surprised":   c.Surprised,
		}
		if c.DeathSavingThrows != nil {
			cm["death_saving_throws"] = map[string]any{
				"successes": c.DeathSavingThrows.Successes,
				"failures":  c.DeathSavingThrows.Failures,
			}
		}
		combatants = append(combatants, cm)
	}

	initOrder := make([]string, 0, len(state.InitiativeOrder))
	for _, id := range state.InitiativeOrder {
		initOrder = append(initOrder, id.String())
	}

	envMap := map[string]any{
		"description": state.Environment.Description,
	}
	if state.Environment.LocationID != nil {
		envMap["location_id"] = state.Environment.LocationID.String()
	}
	if state.Environment.Properties != nil {
		envMap["properties"] = state.Environment.Properties
	}

	m := map[string]any{
		"id":                           state.ID.String(),
		"campaign_id":                  state.CampaignID.String(),
		"combatants":                   combatants,
		"initiative_order":             initOrder,
		"round_number":                 state.RoundNumber,
		"environment":                  envMap,
		"status":                       string(state.Status),
		"narrative":                    state.Narrative,
		"initiative_reroll_each_round": state.InitiativeRerollEachRound,
		"track_death_saving_throws":    state.TrackDeathSavingThrows,
		"surprise_round_active":        state.SurpriseRoundActive,
	}
	if state.ActiveEffects != nil {
		m["active_effects"] = state.ActiveEffects
	}
	return m
}
