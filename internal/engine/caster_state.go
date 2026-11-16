package engine

import (
	"context"
	"encoding/json"
	"git.subcult.tv/subculture-collective/edda/internal/rules/srd521"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// saveCaster changes only the fixture's casting or Fighter recovery fields, keeping other
// stored stats (including unknown metadata) intact. The caller owns the lock.
func saveCaster(ctx context.Context, tx pgx.Tx, campaign, actor uuid.UUID, b *srd521.Build, hp int) error {
	if b.Class == "fighter" {
		fields, err := json.Marshal(map[string]any{"second_wind": b.SecondWind, "hit_dice": b.HitDice, "action_surge_spent": b.ActionSurgeSpent, "last_fighter_long_rest": b.LastFighterLongRest})
		if err != nil {
			return err
		}
		_, err = tx.Exec(ctx, `UPDATE player_characters SET hp=$4,stats=jsonb_set(stats,'{srd521}',(stats->'srd521')||$1::jsonb),updated_at=clock_timestamp() WHERE id=$2 AND campaign_id=$3`, fields, actor, campaign, hp)
		return err
	}
	// Remove links only for this caster's ended/replaced cast; retain cast identity.
	current := ""
	if b.Casting.Concentration != nil {
		current = b.Casting.Concentration.ID
	}
	if _, err := tx.Exec(ctx, `DELETE FROM concentration_recipients r USING concentration_casts c WHERE r.cast_id=c.id AND c.campaign_id=$1 AND c.caster_id=$2 AND c.id::text<>$3`, campaign, actor, current); err != nil {
		return err
	}
	if c := b.Casting.Concentration; c != nil && (c.RecipientID != "" || validCastID(c.ID)) {
		if _, err := tx.Exec(ctx, `INSERT INTO concentration_casts(id,campaign_id,caster_id,started_at,expires_at) VALUES($1,$2,$3,$4,$5) ON CONFLICT(id) DO NOTHING`, c.ID, campaign, actor, c.StartedAt, c.ExpiresAt); err != nil {
			return err
		}
		recipient := actor.String()
		if c.RecipientID != "" {
			recipient = c.RecipientID
		}
		if _, err := tx.Exec(ctx, `INSERT INTO concentration_recipients(cast_id,recipient_id) VALUES($1,$2) ON CONFLICT(cast_id) DO NOTHING`, c.ID, recipient); err != nil {
			return err
		}
	}
	casting, err := json.Marshal(b.Casting)
	if err != nil {
		return err
	}
	// jsonb concatenation preserves unknown casting fields. Set nil pointers
	// explicitly so ending an effect removes the former saved concentration.
	var fields map[string]any
	if err := json.Unmarshal(casting, &fields); err != nil {
		return err
	}
	fields["concentration"] = b.Casting.Concentration
	casting, err = json.Marshal(fields)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `UPDATE player_characters SET hp=$1,stats=jsonb_set(jsonb_set(stats,'{srd521,casting}',COALESCE(stats#>'{srd521,casting}','{}'::jsonb)||$2::jsonb),'{srd521,hit_dice}',to_jsonb($3::int)),updated_at=clock_timestamp() WHERE id=$4 AND campaign_id=$5`, hp, casting, b.HitDice, actor, campaign)
	return err
}

func saveSpellClock(ctx context.Context, tx pgx.Tx, campaign uuid.UUID, now int) error {
	_, err := tx.Exec(ctx, `INSERT INTO campaign_time(campaign_id,day,hour,minute,second) VALUES($1,$2,$3,$4,$5) ON CONFLICT(campaign_id) DO UPDATE SET day=EXCLUDED.day,hour=EXCLUDED.hour,minute=EXCLUDED.minute,second=EXCLUDED.second,updated_at=clock_timestamp()`, campaign, now/86400+1, (now/3600)%24, (now/60)%60, now%60)
	return err
}

// Old self-only saves accepted opaque effect IDs before normalized persistence.
func validCastID(id string) bool { _, err := uuid.Parse(id); return err == nil }
