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
	ID     uuid.UUID
	Name   string
	HP     int
	Space  combat.Space
	Tactic string
	Bandit *srd521.BanditProfile
}

// StoredAction adds persisted equipment and legal fixture targets to preparation.
// Only its Preview is public. Targets deliberately exclude arbitrary NPC stats.
type StoredAction struct {
	HealingTargets []HealingTarget
	Equipment      *srd521.EquipmentReceipt
	Build          *srd521.Build
	Clock          int
	Preview        *ActionPreview
	Targets        []MartialTarget
	Quantity       int
	Encounter      *combat.Encounter
}

// PrepareStoredAction is shared by API preparation and transactional confirmation.
// The narrow fixture is a one-handed longsword against a colocated, previously
// encountered unarmored target explicitly assigned to this actor at five feet.
func prepareStoredActionBase(ctx context.Context, database statedb.DBTX, campaign domain.Campaign, pc domain.PlayerCharacter, intent ActionIntent, lock bool) (*StoredAction, error) {
	if srd521.IsEquipment(intent.Kind) {
		return prepareEquipment(ctx, database, campaign, pc, intent, lock)
	}
	if intent.Kind == srd521.AdvanceLevel {
		return prepareAdvancement(ctx, campaign, pc, intent)
	}
	if srd521.IsLifecycle(intent.Kind) {
		return prepareLifecycle(ctx, campaign, pc, intent)
	}
	if intent.Kind == srd521.SpellCast {
		return prepareStoredSpell(ctx, database, campaign, pc, intent, lock)
	}
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
	result.Targets, err = readMartialTargets(ctx, database, campaign, pc, lock)
	if err != nil {
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
	if sheet.Spellcasting != nil {
		preview.Reason = "Martial attacks require the supported Fighter fixture."
		return result, nil
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

// fixtureSpace preserves the original explicitly ordinary five-foot contract.
// The spatial profile requires explicit facts; it never infers them from prose.
func fixtureSpace(data []byte) combat.Space {
	var props struct {
		Profile string `json:"profile"`
		combat.Space
	}
	if json.Unmarshal(data, &props) != nil {
		return combat.Space{}
	}
	space := props.Space
	if props.Profile == "unarmored-fixture-v1" {
		distance, sight := 5, true
		// Only the exact original fixture gets compatibility defaults.
		if space.DistanceFeet != nil && *space.DistanceFeet == 5 && space.LineOfSight == nil && space.Cover == "" {
			space.DistanceFeet, space.LineOfSight, space.Cover = &distance, &sight, "none"
		}
	}
	space.Revision = fmt.Sprintf("%x", sha256.Sum256(data))
	return space
}

func readMartialTargets(ctx context.Context, database statedb.DBTX, campaign domain.Campaign, pc domain.PlayerCharacter, lock bool) ([]MartialTarget, error) {
	locking := ""
	if lock {
		locking = " FOR UPDATE"
	}
	result := []MartialTarget{}
	targets, err := database.Query(ctx, `SELECT n.id,n.name,n.hp,n.properties->'srd521_target' FROM npcs n WHERE n.campaign_id=$1 AND n.alive AND n.hp>0 AND n.location_id=$2 AND n.properties->'srd521_target'->>'profile' IN ('unarmored-fixture-v1','ordinary-spatial-fixture-v1','bandit-v1') AND n.properties->'srd521_target'->>'actor_id'=$3::text AND COALESCE(n.properties->'srd521_target'->>'hidden','false')='false' AND EXISTS(SELECT 1 FROM session_logs l WHERE l.campaign_id=n.campaign_id AND n.id=ANY(l.npcs_involved)) ORDER BY n.id`+locking, dbutil.ToPgtype(campaign.ID), pc.CurrentLocationID, pc.ID.String())
	if err != nil {
		return nil, err
	}
	defer targets.Close()
	for targets.Next() {
		var target MartialTarget
		var facts []byte
		if err := targets.Scan(&target.ID, &target.Name, &target.HP, &facts); err != nil {
			return nil, err
		}
		target.Space = fixtureSpace(facts)
		var profile struct {
			Profile string                `json:"profile"`
			Bandit  *srd521.BanditProfile `json:"bandit"`
		}
		if err := json.Unmarshal(facts, &profile); err != nil {
			continue
		}
		if profile.Profile == "bandit-v1" {
			if profile.Bandit == nil {
				continue
			}
			if _, _, err := profile.Bandit.Record(); err != nil {
				continue
			}
			target.Bandit = profile.Bandit
		}
		var tactic struct {
			Tactic string `json:"tactic"`
		}
		if err := json.Unmarshal(facts, &tactic); err == nil && tactic.Tactic == "withdraw" {
			target.Tactic = "withdraw"
		}
		result = append(result, target)
	}
	if err := targets.Err(); err != nil {
		return nil, err
	}

	return result, nil
}
