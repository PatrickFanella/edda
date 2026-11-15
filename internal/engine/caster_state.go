package engine

import (
	"context"
	"encoding/json"
	"git.subcult.tv/subculture-collective/edda/internal/rules/srd521"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// saveCaster changes only the fixture's casting/Hit Dice fields, keeping other
// stored stats (including unknown metadata) intact. The caller owns the lock.
func saveCaster(ctx context.Context, tx pgx.Tx, campaign, actor uuid.UUID, b *srd521.Build, hp int) error {
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
