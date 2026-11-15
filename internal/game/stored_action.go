package game

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"

	"git.subcult.tv/subculture-collective/edda/internal/combat"
	"git.subcult.tv/subculture-collective/edda/internal/dbutil"
	"git.subcult.tv/subculture-collective/edda/internal/domain"
	"git.subcult.tv/subculture-collective/edda/internal/rules/srd521"
	statedb "git.subcult.tv/subculture-collective/edda/internal/state/sqlc"
	"github.com/google/uuid"
)

// MartialTarget is a server-only, explicitly ordinary fixture target.
type MartialTarget struct {
	ID   uuid.UUID
	Name string
	HP   int
}

// StoredAction adds persisted equipment and legal fixture targets to preparation.
// Only its Preview is public. Targets deliberately exclude arbitrary NPC stats.
type StoredAction struct {
	Preview   *ActionPreview
	Targets   []MartialTarget
	Quantity  int
	Encounter *combat.Encounter
}

// PrepareStoredAction is shared by API preparation and transactional confirmation.
// The narrow fixture is a one-handed longsword against a colocated, previously
// encountered unarmored target explicitly assigned to this actor at five feet.
func prepareStoredActionBase(ctx context.Context, database statedb.DBTX, campaign domain.Campaign, pc domain.PlayerCharacter, intent ActionIntent, lock bool) (*StoredAction, error) {
	if intent.Kind == srd521.ConsumableUse {
		return prepareConsumable(ctx, database, campaign, pc, intent, lock)
	}
	baseIntent := intent
	if intent.Kind == srd521.MartialAttack {
		baseIntent = ActionIntent{ActorID: intent.ActorID, Kind: srd521.AbilityCheck, Ability: "strength"}
	}
	preview, err := PrepareCharacterAction(ctx, campaign, pc, baseIntent)
	if err != nil {
		return nil, err
	}
	result := &StoredAction{Preview: preview}
	if intent.Kind != srd521.MartialAttack {
		return result, nil
	}
	preview.Intent = intent
	if !preview.Available {
		return result, nil
	}
	preview.Available = false
	preview.Reason = "Only an equipped SRD longsword and a prepared ordinary fixture target are supported."
	if intent.Ability != "" || intent.Skill != "" {
		return result, nil
	}
	itemID, err := uuid.Parse(intent.EntryID)
	if err != nil {
		return result, nil
	}
	locking := ""
	if lock {
		locking = " FOR UPDATE"
	}
	var properties []byte
	var updated time.Time
	rows, err := database.Query(ctx, `SELECT properties,updated_at FROM items WHERE id=$1 AND campaign_id=$2 AND player_character_id=$3 AND item_type='weapon' AND equipped AND quantity>0`+locking, dbutil.ToPgtype(itemID), dbutil.ToPgtype(campaign.ID), dbutil.ToPgtype(pc.ID))
	if err != nil {
		return nil, err
	}
	found := rows.Next()
	if found {
		err = rows.Scan(&properties, &updated)
	}
	rows.Close()
	if err != nil {
		return nil, err
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	var item struct {
		ID string `json:"srd521_id"`
	}
	if !found || json.Unmarshal(properties, &item) != nil || item.ID != "longsword" {
		return result, nil
	}
	targets, err := database.Query(ctx, `SELECT n.id,n.name,n.hp FROM npcs n WHERE n.campaign_id=$1 AND n.alive AND n.hp>0 AND n.location_id=$2 AND n.properties->'srd521_target'=jsonb_build_object('profile','unarmored-fixture-v1','actor_id',$3::text,'distance_feet',5) AND EXISTS(SELECT 1 FROM session_logs l WHERE l.campaign_id=n.campaign_id AND n.id=ANY(l.npcs_involved)) ORDER BY n.id`+locking, dbutil.ToPgtype(campaign.ID), pc.CurrentLocationID, pc.ID.String())
	if err != nil {
		return nil, err
	}
	defer targets.Close()
	for targets.Next() {
		var target MartialTarget
		if err := targets.Scan(&target.ID, &target.Name, &target.HP); err != nil {
			return nil, err
		}
		result.Targets = append(result.Targets, target)
		preview.Targets = append(preview.Targets, ActionTarget{ID: target.ID, Label: target.Name})
	}
	if err := targets.Err(); err != nil {
		return nil, err
	}
	encoded, err := json.Marshal([]any{preview.ExpectedRevision, itemID, properties, updated, result.Targets})
	if err != nil {
		return nil, err
	}
	preview.ExpectedRevision = fmt.Sprintf("martial-v1:%x", sha256.Sum256(encoded))
	if len(result.Targets) == 0 {
		return result, nil
	}
	sheet, err := srd521.Project(pc.Stats, pc.Level, pc.HP, pc.MaxHP, pc.Experience)
	if err != nil {
		return nil, err
	}
	modifier := srd521.LongswordModifier(sheet)
	preview.Modifier = &modifier
	preview.Available = true
	preview.Reason = ""
	preview.Options = []string{"ordinary, one hand; 1d8 + Strength slashing; critical hit: 2d8 + Strength"}
	preview.Costs = []ActionCost{{ResourceID: "attack-action", Amount: 1}}
	preview.CostStatus = "fixture-only"
	preview.SourceIDs = []string{"srd-5.2.1:content/d20-tests/attack-rolls", "srd-5.2.1:content/weapons/longsword", "srd-5.2.1:content/damage/critical-hits"}
	return result, nil
}
