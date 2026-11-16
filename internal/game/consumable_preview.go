package game

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"

	"git.subcult.tv/subculture-collective/edda/internal/dbutil"
	"git.subcult.tv/subculture-collective/edda/internal/domain"
	"git.subcult.tv/subculture-collective/edda/internal/rules/srd521"
	statedb "git.subcult.tv/subculture-collective/edda/internal/state/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func prepareConsumable(ctx context.Context, database statedb.DBTX, campaign domain.Campaign, pc domain.PlayerCharacter, intent ActionIntent, lock bool) (*StoredAction, error) {
	preview, err := PrepareCharacterAction(ctx, campaign, pc, ActionIntent{ActorID: intent.ActorID, Kind: srd521.AbilityCheck, Ability: "strength"})
	if err != nil {
		return nil, err
	}
	preview.Intent = intent
	result := &StoredAction{Preview: preview}
	if !preview.Available {
		return result, nil
	}
	preview.Available = false
	preview.Modifier = nil
	preview.Reason = "Only an owned Potion of Healing supports self-use or a prepared willing conscious ally."
	if intent.Ability != "" || intent.Skill != "" || intent.SlotLevel != 0 {
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
	var quantity int
	err = database.QueryRow(ctx, `SELECT properties,updated_at,quantity FROM items WHERE id=$1 AND campaign_id=$2 AND player_character_id=$3 AND item_type='consumable' AND quantity>0`+locking, dbutil.ToPgtype(itemID), dbutil.ToPgtype(campaign.ID), dbutil.ToPgtype(pc.ID)).Scan(&properties, &updated, &quantity)
	if err == pgx.ErrNoRows {
		return result, nil
	}
	if err != nil {
		return nil, err
	}
	var item struct {
		ID string `json:"srd521_id"`
	}
	if json.Unmarshal(properties, &item) != nil || item.ID != "potion-of-healing" {
		return result, nil
	}
	encoded, err := json.Marshal([]any{preview.ExpectedRevision, itemID, properties, updated, quantity})
	if err != nil {
		return nil, err
	}
	preview.ExpectedRevision = fmt.Sprintf("consumable-v1:%x", sha256.Sum256(encoded))
	result.Quantity = quantity
	preview.Available = true
	preview.Reason = ""
	preview.Targets = []ActionTarget{{ID: pc.ID, Label: pc.Name + " (self)"}}
	preview.Options = []string{fmt.Sprintf("Quantity: %d; regain 2d4 + 2 HP, capped at %d (current %d); self-use only", quantity, pc.MaxHP, pc.HP)}
	preview.Costs = []ActionCost{{ResourceID: "item:" + itemID.String(), Amount: 1}, {ResourceID: "bonus-action", Amount: 1}}
	preview.CostStatus = "fixture-only"
	preview.SourceIDs = []string{"srd-5.2.1:content/items/potion-of-healing", "srd-5.2.1:content/damage/healing"}
	if err := appendHealingTargets(ctx, database, campaign, pc, result, "potion-healing-ally-v1", lock); err != nil {
		return nil, err
	}
	if len(result.HealingTargets) > 0 {
		encoded, err := json.Marshal([]any{preview.ExpectedRevision, result.HealingTargets})
		if err != nil {
			return nil, err
		}
		preview.ExpectedRevision = fmt.Sprintf("potion-ally-v1:%x", sha256.Sum256(encoded))
		preview.Options = []string{fmt.Sprintf("Quantity: %d; regain 2d4 + 2 HP capped at target maximum HP; self or one prepared willing conscious ally within 5 feet; one potion and one Bonus Action; outside encounters only for allies", quantity)}
	}
	return result, nil
}
