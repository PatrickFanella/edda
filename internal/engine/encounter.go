package engine

import (
	"context"
	"git.subcult.tv/subculture-collective/edda/internal/combat"
	"git.subcult.tv/subculture-collective/edda/internal/dbutil"
	"git.subcult.tv/subculture-collective/edda/internal/rules/srd521"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// resolveEncounterNPC runs the fixture's NPC turn under the confirmation
// transaction. Both HP effects and dice receipts commit with the turn cursor.
func resolveEncounterNPC(ctx context.Context, tx pgx.Tx, campaignID uuid.UUID, encounter *combat.Encounter, sheet *srd521.CharacterSheet, dice srd521.Dice) ([]*srd521.AttackReceipt, error) {
	receipts := []*srd521.AttackReceipt{}
	if encounter.Status != "active" || encounter.Active().Player {
		return receipts, nil
	}
	npc := encounter.Active()
	if err := encounter.Spend(npc.ID, srd521.MartialAttack); err != nil {
		return nil, err
	}
	var player *combat.EncounterActor
	for i := range encounter.Actors {
		if encounter.Actors[i].Player {
			player = &encounter.Actors[i]
		}
	}
	attack, err := srd521.ResolveFixtureUnarmed(sheet.Defenses["armor-class"].Value, dice)
	if err != nil {
		return nil, err
	}
	attack.TargetID = player.ID.String()
	attack.Target = player.Name
	hp := max(0, player.HP-attack.Damage)
	if _, err := tx.Exec(ctx, `UPDATE player_characters SET hp=$1,updated_at=clock_timestamp() WHERE id=$2 AND campaign_id=$3`, hp, dbutil.ToPgtype(player.ID), dbutil.ToPgtype(campaignID)); err != nil {
		return nil, err
	}
	encounter.SetHP(player.ID, hp)
	receipts = append(receipts, attack)
	encounter.Advance()
	return receipts, nil
}
