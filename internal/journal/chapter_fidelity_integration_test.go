//go:build integration

package journal

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"git.subcult.tv/subculture-collective/edda/internal/db"
	"git.subcult.tv/subculture-collective/edda/internal/llm"
	statedb "git.subcult.tv/subculture-collective/edda/internal/state/sqlc"
	"git.subcult.tv/subculture-collective/edda/pkg/api"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
)

type fidelityModel struct {
	messages [][]llm.Message
	content  string
}

func (m *fidelityModel) Complete(_ context.Context, messages []llm.Message, tools []llm.Tool) (*llm.Response, error) {
	if len(tools) != 0 {
		return nil, errors.New("chapter fixture must not invoke tools")
	}
	m.messages = append(m.messages, messages)
	return &llm.Response{Content: m.content}, nil
}

func (*fidelityModel) Stream(context.Context, []llm.Message, []llm.Tool) (<-chan llm.StreamChunk, error) {
	return nil, errors.New("chapter fixture must not stream")
}

func TestIntegrationCuratedChapterFidelity(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	container, err := tcpostgres.Run(ctx, "docker.io/pgvector/pgvector:pg16", tcpostgres.WithDatabase("fidelity"), tcpostgres.WithUsername("edda"), tcpostgres.WithPassword("fixture-only"), tcpostgres.BasicWaitStrategies())
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		cleanup, done := context.WithTimeout(context.Background(), 30*time.Second)
		defer done()
		if err := container.Terminate(cleanup); err != nil {
			t.Error(err)
		}
	}()
	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	migrationDB, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer migrationDB.Close()
	migrations, err := goose.NewProvider(goose.DialectPostgres, migrationDB, os.DirFS("../../migrations"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := migrations.Up(ctx); err != nil {
		t.Fatal(err)
	}
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	q := statedb.New(pool)
	user, err := q.CreateUser(ctx, "chapter-fidelity-fixture")
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range fidelityFixtures(t) {
		t.Run(f.ID, func(t *testing.T) {
			campaign, err := q.CreateCampaign(ctx, statedb.CreateCampaignParams{Name: f.ID, Status: "active", CreatedBy: user.ID, RulesVersion: f.Source.RulesVersion})
			if err != nil {
				t.Fatal(err)
			}
			cid := db.FromPgUUID(campaign.ID)
			exec := func(query string, args ...any) {
				t.Helper()
				if _, err := pool.Exec(ctx, query, args...); err != nil {
					t.Fatal(err)
				}
			}
			exec(`UPDATE campaigns SET rules_mode=$2 WHERE id=$1`, cid, f.Source.RulesMode)
			exec(`INSERT INTO world_facts(campaign_id,fact,category,source,player_known) VALUES($1,$2,'secret','fixture',false)`, cid, f.HiddenLore)
			toolJSON, err := json.Marshal([]map[string]string{{"private_result": f.ToolSecret}})
			if err != nil {
				t.Fatal(err)
			}
			all := append(append([]api.ChapterTurn{}, f.Source.SourceTurns...), f.Pending, f.Later)
			for _, turn := range all {
				exec(`INSERT INTO session_logs(id,campaign_id,turn_number,player_input,input_type,llm_response,tool_calls) VALUES($1,$2,$3,$4,'game_action',$5,$6)`, turn.LogID, cid, turn.TurnNumber, turn.PlayerInput, turn.Narrative, toolJSON)
				status := "completed"
				if turn.TurnNumber == f.Pending.TurnNumber {
					status = "pending_narration"
				}
				exec(`INSERT INTO turn_attempts(id,campaign_id,turn_number,player_input,status,final_narrative) VALUES($1,$2,$3,$4,$5,$6)`, turn.AttemptID, cid, turn.TurnNumber, turn.PlayerInput, status, turn.Narrative)
			}
			store := NewStore(pool)
			model := &fidelityModel{content: f.chapter(t)}
			service := NewSummarizer(model, store)
			revision, err := service.GenerateChapter(ctx, cid, uuid.New(), "")
			if err != nil || revision == nil || revision.Status != "completed" || revision.PromptVersion != chapterPromptVersion {
				t.Fatalf("generation=%+v %v", revision, err)
			}
			chapters, err := store.ListChapters(ctx, cid)
			if err != nil || len(chapters) != 1 {
				t.Fatalf("saved chapters=%+v %v", chapters, err)
			}
			got := chapters[0]
			_, hash, err := chapterSource(f.Source.SourceTurns, f.Source.RulesVersion, f.Source.RulesMode)
			if err != nil || got.SourceHash != hash || got.FromTurn != f.Source.FromTurn || got.ToTurn != f.Source.ToTurn || got.KnowledgeBoundary != f.Source.KnowledgeBoundary || got.RulesVersion != f.Source.RulesVersion || got.RulesMode != f.Source.RulesMode || !reflect.DeepEqual(got.SourceTurns, f.Source.SourceTurns) {
				t.Fatalf("saved source differs from original: %+v %v", got, err)
			}
			if len(got.Revisions) != 1 || got.Revisions[0].Content != model.content || revision.Content != model.content {
				t.Fatal("recorded chapter response changed in persistence")
			}
			if err := f.checkChapter(got.Revisions[0].Content); err != nil {
				t.Fatal(err)
			}
			// Retry against changed private state must still use the same source.
			exec(`UPDATE world_facts SET fact='CHANGED_PRIVATE_LORE' WHERE campaign_id=$1`, cid)
			exec(`UPDATE session_logs SET llm_response='EDITED_AFTER_CAPTURE' WHERE campaign_id=$1 AND turn_number=1`, cid)
			if _, err := service.GenerateChapter(ctx, cid, uuid.New(), got.ID); err != nil {
				t.Fatal(err)
			}
			if len(model.messages) != 2 {
				t.Fatalf("expected generation and regeneration, got %d calls", len(model.messages))
			}
			for _, messages := range model.messages {
				if len(messages) != 2 || messages[0].Role != llm.RoleSystem || messages[0].Content != faithfulChapterPrompt || messages[1].Role != llm.RoleUser {
					t.Fatal("unexpected chapter message contract")
				}
				var input api.ChapterResponse
				if err := json.Unmarshal([]byte(messages[1].Content), &input); err != nil || !reflect.DeepEqual(input.SourceTurns, f.Source.SourceTurns) || len(input.Revisions) != 0 {
					t.Fatalf("model did not receive exact frozen original: %v", err)
				}
				for _, forbidden := range []string{f.HiddenLore, f.ToolSecret, f.Pending.Narrative, f.Later.Narrative, "CHANGED_PRIVATE_LORE", "EDITED_AFTER_CAPTURE"} {
					for _, message := range messages {
						if strings.Contains(message.Content, forbidden) {
							t.Fatalf("private/out-of-range material entered model request: %s", forbidden)
						}
					}
				}
			}
		})
	}
}
