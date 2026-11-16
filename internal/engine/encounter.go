package engine

import (
	"context"
	"encoding/json"
	"git.subcult.tv/subculture-collective/edda/internal/combat"
	"git.subcult.tv/subculture-collective/edda/internal/dbutil"
	"git.subcult.tv/subculture-collective/edda/internal/game"
	"git.subcult.tv/subculture-collective/edda/internal/rules/srd521"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// resolveEncounterNPC runs the fixture's NPC turn under the confirmation
// transaction. Both HP effects and dice receipts commit with the turn cursor.
func resolveEncounterNPC(ctx context.Context, tx pgx.Tx, campaignID uuid.UUID, encounter *combat.Encounter, sheet *srd521.CharacterSheet, dice srd521.Dice) ([]*srd521.AttackReceipt, []*srd521.EffectReceipt, error) {
	receipts := []*srd521.AttackReceipt{}
	effects := []*srd521.EffectReceipt{}
	if encounter.Status != "active" || encounter.Active().Player || encounter.Reaction.Pending != nil {
		return receipts, effects, nil
	}
	// An out-of-reach NPC holds position. Unknown facts pause resolution rather
	// than manufacturing a hit or a path through an obstacle.
	if encounter.Space == nil {
		return nil, nil, ErrActionRefresh
	}
	if err := encounter.Space.Validate(); err != nil {
		return nil, nil, ErrActionRefresh
	}
	if encounter.Bandit != nil {
		if err := combat.BanditAttack(encounter.Bandit, encounter.Space); err != nil {
			return nil, nil, ErrActionRefresh
		}
	} else if err := encounter.Space.Melee(); err != nil {
		encounter.Advance()
		return receipts, effects, nil
	}
	npc := encounter.Active()
	if err := encounter.Spend(npc.ID, srd521.MartialAttack); err != nil {
		return nil, nil, err
	}
	var player *combat.EncounterActor
	for i := range encounter.Actors {
		if encounter.Actors[i].Player {
			player = &encounter.Actors[i]
		}
	}
	var attack *srd521.AttackReceipt
	var err error
	ac := sheet.Defenses["armor-class"].Value + encounter.Space.CoverBonus()
	if encounter.Bandit != nil {
		attack, err = srd521.ResolveBanditAttack(*encounter.Bandit, ac, dice)
	} else {
		attack, err = srd521.ResolveFixtureUnarmed(ac, dice)
	}
	if err != nil {
		return nil, nil, err
	}
	attack.TargetID = player.ID.String()
	attack.Target = player.Name
	hp := max(0, player.HP-attack.Damage)
	if _, err := tx.Exec(ctx, `UPDATE player_characters SET hp=$1,updated_at=clock_timestamp() WHERE id=$2 AND campaign_id=$3`, hp, dbutil.ToPgtype(player.ID), dbutil.ToPgtype(campaignID)); err != nil {
		return nil, nil, err
	}
	if attack.Damage > 0 && sheet.Spellcasting != nil {
		var stats []byte
		if err := tx.QueryRow(ctx, `SELECT stats FROM player_characters WHERE id=$1`, player.ID).Scan(&stats); err != nil {
			return nil, nil, err
		}
		var envelope struct {
			Build srd521.Build `json:"srd521"`
		}
		if err := json.Unmarshal(stats, &envelope); err != nil {
			return nil, nil, err
		}
		now, err := game.ReadSpellClock(ctx, tx, campaignID)
		if err != nil {
			return nil, nil, err
		}
		damageSheet := *sheet
		damageSheet.HP.Value = player.HP
		effect, err := srd521.ConcentrationDamage(&envelope.Build, &damageSheet, attack.Damage, hp, now, dice)
		if err != nil {
			return nil, nil, err
		}
		if effect != nil {
			if err := saveCaster(ctx, tx, campaignID, player.ID, &envelope.Build, hp); err != nil {
				return nil, nil, err
			}
			effects = append(effects, effect)
		}
	}
	encounter.SetHP(player.ID, hp)
	receipts = append(receipts, attack)
	if encounter.Status == "active" && encounter.NPCTactic == "withdraw" && encounter.Space.Cover == "none" && *encounter.Space.DistanceFeet == 5 {
		// A voluntary step from five to ten feet is this fixture's one supported
		// opportunity trigger. Suspend BEFORE movement, after the NPC's attack.
		var weapon uuid.UUID
		err := tx.QueryRow(ctx, `SELECT id FROM items WHERE campaign_id=$1 AND player_character_id=$2 AND item_type='weapon' AND equipped AND quantity>0 AND properties->>'srd521_id'='longsword' ORDER BY id LIMIT 1 FOR UPDATE`, dbutil.ToPgtype(campaignID), dbutil.ToPgtype(player.ID)).Scan(&weapon)
		if err != nil && err != pgx.ErrNoRows {
			return nil, nil, err
		}
		if err == nil && encounter.Reaction.Available {
			encounter.Reaction.Pending = &combat.ReactionTrigger{ID: uuid.New(), TargetID: npc.ID, WeaponID: weapon, SpaceRevision: encounter.Space.Revision}
			return receipts, effects, nil
		}
		finishWithdrawal(encounter)
		return receipts, effects, nil
	}
	encounter.Advance()
	return receipts, effects, nil
}

// finishWithdrawal consumes the pending NPC movement once, then returns control
// to the player. A changed/unclear path cancels the step rather than guessing.
func finishWithdrawal(encounter *combat.Encounter) {
	encounter.Reaction.Pending = nil
	if encounter.Status != "active" {
		return
	}
	if err := encounter.Spend(encounter.Active().ID, combat.EncounterRetreat); err != nil {
		// Retaining position is safe when a new scene ruling blocks the path.
		encounter.Advance()
		return
	}
	encounter.Advance()
}
