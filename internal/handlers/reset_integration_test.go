//go:build integration

package handlers_test

import (
	"context"
	"fmt"
	"testing"

	"git.subcult.tv/subculture-collective/edda/internal/dbutil"
	"git.subcult.tv/subculture-collective/edda/internal/saves"
	statedb "git.subcult.tv/subculture-collective/edda/internal/state/sqlc"
)

func TestIntegrationStartOverAtomic(t *testing.T) {
	for _, failure := range []string{"session_logs", "save_points", "campaign_time", ""} {
		name := failure
		if name == "" {
			name = "success"
		}
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			tx, err := testPool.Begin(ctx)
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = tx.Rollback(ctx) }()
			q := statedb.New(tx)
			camp, err := q.CreateCampaign(ctx, statedb.CreateCampaignParams{Name: "Reset test", Status: "active", CreatedBy: dbutil.ToPgtype(testUserID)})
			if err != nil {
				t.Fatal(err)
			}
			other, err := q.CreateCampaign(ctx, statedb.CreateCampaignParams{Name: "Other campaign", Status: "active", CreatedBy: dbutil.ToPgtype(testUserID)})
			if err != nil {
				t.Fatal(err)
			}
			for _, id := range []any{camp.ID, other.ID} {
				for _, query := range []string{
					`INSERT INTO session_logs (campaign_id, turn_number, player_input, input_type, llm_response) VALUES ($1, 1, 'wait', 'game_action', 'Still here')`,
					`INSERT INTO save_points (campaign_id, name, turn_number, is_auto) VALUES ($1, 'keep', 1, false)`,
					`INSERT INTO campaign_time (campaign_id, day, hour, minute) VALUES ($1, 7, 12, 30)`,
					`INSERT INTO locations (campaign_id, name) VALUES ($1, 'Keep world')`,
				} {
					if _, err := tx.Exec(ctx, query, id); err != nil {
						t.Fatal(err)
					}
				}
			}
			if failure != "" {
				_, err = tx.Exec(ctx, `CREATE FUNCTION pg_temp.reject_reset() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'injected reset failure'; END $$`)
				if err != nil {
					t.Fatal(err)
				}
				_, err = tx.Exec(ctx, fmt.Sprintf(`CREATE TRIGGER reject_reset BEFORE DELETE ON %s FOR EACH ROW EXECUTE FUNCTION pg_temp.reject_reset()`, failure))
				if err != nil {
					t.Fatal(err)
				}
			}
			// A savepoint lets us inspect rows after PostgreSQL aborts the statement.
			attempt, err := tx.Begin(ctx)
			if err != nil {
				t.Fatal(err)
			}
			err = saves.NewStore(attempt).StartOver(ctx, dbutil.FromPgtype(camp.ID))
			if failure != "" {
				if err == nil {
					t.Fatal("expected injected failure")
				}
				if err := attempt.Rollback(ctx); err != nil {
					t.Fatal(err)
				}
			} else {
				if err != nil {
					t.Fatal(err)
				}
				if err := attempt.Commit(ctx); err != nil {
					t.Fatal(err)
				}
			}
			for _, table := range []string{"session_logs", "save_points", "campaign_time", "locations"} {
				for _, target := range []struct {
					id   any
					want int
				}{{camp.ID, 1}, {other.ID, 1}} {
					if failure == "" && target.id == camp.ID && table != "locations" {
						target.want = 0
					}
					var count int
					if err := tx.QueryRow(ctx, "SELECT count(*) FROM "+table+" WHERE campaign_id = $1", target.id).Scan(&count); err != nil {
						t.Fatal(err)
					}
					if count != target.want {
						t.Errorf("%s rows = %d, want %d", table, count, target.want)
					}
				}
			}
		})
	}
}
