package game

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"

	"git.subcult.tv/subculture-collective/edda/internal/combat"
	"git.subcult.tv/subculture-collective/edda/internal/dbutil"
	"git.subcult.tv/subculture-collective/edda/internal/domain"
	statedb "git.subcult.tv/subculture-collective/edda/internal/state/sqlc"
	"github.com/google/uuid"
)

// HealingTarget is a server-only trusted fixture, never arbitrary NPC statistics.
type HealingTarget struct {
	ID        uuid.UUID
	Name      string
	HP        int
	MaxHP     int
	Space     combat.Space
	UpdatedAt time.Time
}

// HealingTarget returns only a target revalidated in this preparation snapshot.
func (s *StoredAction) HealingTarget(id uuid.UUID) *HealingTarget {
	for i := range s.HealingTargets {
		if s.HealingTargets[i].ID == id {
			return &s.HealingTargets[i]
		}
	}
	return nil
}

// allyFacts deliberately has no compatibility defaults. Missing booleans and
// unknown condition/mechanic fields fail closed. Disposition/prose is not consent.
type allyFacts struct {
	Profile      string    `json:"profile"`
	ActorID      uuid.UUID `json:"actor_id"`
	LocationID   uuid.UUID `json:"location_id"`
	Allied       *bool     `json:"allied"`
	Willing      *bool     `json:"willing"`
	Conscious    *bool     `json:"conscious"`
	Hidden       *bool     `json:"hidden"`
	Conditions   string    `json:"conditions"`
	MaxHP        int       `json:"max_hp"`
	DistanceFeet *int      `json:"distance_feet"`
	LineOfSight  *bool     `json:"line_of_sight"`
	Cover        string    `json:"cover"`
}

func prepareStoredSpell(ctx context.Context, database statedb.DBTX, campaign domain.Campaign, pc domain.PlayerCharacter, intent ActionIntent, lock bool) (*StoredAction, error) {
	result, err := prepareSpell(ctx, campaign, pc, intent)
	if err != nil || !result.Preview.Available || intent.EntryID != "cure-wounds" {
		return result, err
	}
	locking := ""
	if lock {
		locking = " FOR UPDATE OF n"
	}
	rows, err := database.Query(ctx, `SELECT n.id,n.name,n.hp,n.updated_at,n.properties->'srd521_ally'
 FROM npcs n WHERE n.campaign_id=$1 AND n.location_id=$2 AND n.alive AND n.hp>0
 AND n.properties->'srd521_ally'->>'profile'='cure-wounds-ally-v1'
 AND n.properties->'srd521_ally'->>'actor_id'=$3
 AND EXISTS(SELECT 1 FROM session_logs l WHERE l.campaign_id=n.campaign_id AND n.id=ANY(l.npcs_involved))
 ORDER BY n.id`+locking, dbutil.ToPgtype(campaign.ID), pc.CurrentLocationID, pc.ID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	yes := func(v *bool) bool { return v != nil && *v }
	for rows.Next() {
		var target HealingTarget
		var data []byte
		if err := rows.Scan(&target.ID, &target.Name, &target.HP, &target.UpdatedAt, &data); err != nil {
			return nil, err
		}
		// Require the exact documented vocabulary before Go's case-insensitive
		// struct decoding. Every fixture field is explicit and non-null.
		var fields map[string]json.RawMessage
		if json.Unmarshal(data, &fields) != nil {
			continue
		}
		keys := []string{"profile", "actor_id", "location_id", "allied", "willing", "conscious", "hidden", "conditions", "max_hp", "distance_feet", "line_of_sight", "cover"}
		valid := len(fields) == len(keys)
		for _, key := range keys {
			value := bytes.TrimSpace(fields[key])
			if len(value) == 0 || bytes.Equal(value, []byte("null")) {
				valid = false
			}
		}
		var facts allyFacts
		if !valid || json.Unmarshal(data, &facts) != nil {
			continue
		}
		space := combat.Space{DistanceFeet: facts.DistanceFeet, LineOfSight: facts.LineOfSight, Cover: facts.Cover}
		if facts.Profile != "cure-wounds-ally-v1" || facts.ActorID != pc.ID || facts.LocationID == uuid.Nil || pc.CurrentLocationID == nil || facts.LocationID != *pc.CurrentLocationID || !yes(facts.Allied) || !yes(facts.Willing) || !yes(facts.Conscious) || facts.Hidden == nil || *facts.Hidden || facts.Conditions != "ordinary" || facts.MaxHP < target.HP || space.Melee() != nil || facts.Cover != "none" {
			continue
		}
		target.MaxHP, target.Space = facts.MaxHP, space
		target.Space.Revision = fmt.Sprintf("%x", sha256.Sum256(data))
		result.HealingTargets = append(result.HealingTargets, target)
		result.Preview.Targets = append(result.Preview.Targets, ActionTarget{ID: target.ID, Label: target.Name + " (willing conscious ally; touch)", DistanceFeet: target.Space.DistanceFeet, ReachFeet: 5})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	encoded, err := json.Marshal([]any{result.Preview.ExpectedRevision, result.HealingTargets})
	if err != nil {
		return nil, err
	}
	result.Preview.ExpectedRevision = fmt.Sprintf("cure-ally-v1:%x", sha256.Sum256(encoded))
	result.Preview.Options = []string{"Cure Wounds; prepared Cleric spell; Wisdom; verbal and somatic components; self or one explicitly willing conscious known allied creature within 5 feet with clear touch reach and ordinary conditions; instantaneous; 2d8 + Wisdom capped at target maximum HP; one level 1 slot. Other targets, conditions and slots unavailable."}
	return result, nil
}
