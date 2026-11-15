//go:build integration

package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/charmbracelet/log"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"

	"git.subcult.tv/subculture-collective/edda/internal/config"
	"git.subcult.tv/subculture-collective/edda/internal/db"
	"git.subcult.tv/subculture-collective/edda/internal/llm"
	statedb "git.subcult.tv/subculture-collective/edda/internal/state/sqlc"
)

// Exercise the production router with real configured providers, local recording
// endpoints and disposable PostgreSQL. No external model service is contacted.
func TestIntegrationWorkflowRoutes(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	container, err := tcpostgres.Run(ctx, "docker.io/pgvector/pgvector:pg16", tcpostgres.WithDatabase("routing"), tcpostgres.WithUsername("edda"), tcpostgres.WithPassword("fixture-only"), tcpostgres.BasicWaitStrategies())
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cleanupCancel()
		if err := container.Terminate(cleanupCtx); err != nil {
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
	queries := statedb.New(pool)

	for _, explicit := range []bool{false, true} {
		name := "fallback"
		if explicit {
			name = "explicit"
		}
		t.Run(name, func(t *testing.T) {
			var mu sync.Mutex
			var calls []string
			modelServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method == http.MethodGet {
					w.Header().Set("Content-Type", "application/json")
					_, _ = w.Write([]byte(`{"models":[]}`))
					return
				}
				var req struct {
					Model string `json:"model"`
				}
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					t.Error(err)
					http.Error(w, "invalid request", 400)
					return
				}
				mu.Lock()
				calls = append(calls, req.Model)
				mu.Unlock()
				w.Header().Set("Content-Type", "application/json")
				// Valid for journal summaries; deliberately invalid skeleton content makes
				// world generation stop after the selected provider is called.
				if err := json.NewEncoder(w).Encode(map[string]any{"message": map[string]string{"role": "assistant", "content": "A faithful fixture summary."}, "done": true}); err != nil {
					t.Error(err)
				}
			}))
			defer modelServer.Close()
			cfg := config.Config{LLM: config.LLMConfig{Provider: "ollama", Ollama: config.OllamaConfig{Endpoint: modelServer.URL, Model: "default"}}}
			if explicit {
				cfg.LLM.Profiles = map[string]config.LLMProfileConfig{"world": {Provider: "ollama", Model: "world"}, "summary": {Provider: "ollama", Model: "summary"}}
				cfg.LLM.Routes = config.LLMRoutesConfig{Worldgen: "world", Summary: "summary"}
			}
			roster, err := llm.NewRoster(cfg)
			if err != nil {
				t.Fatal(err)
			}
			user, err := queries.CreateUser(ctx, "route-"+name)
			if err != nil {
				t.Fatal(err)
			}
			campaign, err := queries.CreateCampaign(ctx, statedb.CreateCampaignParams{Name: "route fixture", Status: "active", CreatedBy: user.ID})
			if err != nil {
				t.Fatal(err)
			}
			cid := db.FromPgUUID(campaign.ID)
			if _, err := pool.Exec(ctx, "INSERT INTO session_logs(campaign_id,turn_number,player_input,input_type,llm_response) VALUES($1,1,'Inspect the gate','game_action','The gate is closed')", cid); err != nil {
				t.Fatal(err)
			}
			router := NewRouterWithProviders(log.New(io.Discard), nil, queries, roster.Provider(llm.FlowDefault), roster.Provider(llm.FlowWorldgen), roster.Provider(llm.FlowSummary), pool, uuid.UUID(user.ID.Bytes), cfg, nil)
			for _, body := range []string{`{"from_turn":1,"to_turn":1}`, `{}`} {
				// Cover explicit and unsummarized ranges; the turn scheduler is
				// intentionally not enabled by this route-wiring issue (#38).
				if _, err := pool.Exec(ctx, "DELETE FROM session_summaries WHERE campaign_id=$1", cid); err != nil {
					t.Fatal(err)
				}
				rec := httptest.NewRecorder()
				router.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/v1/campaigns/"+cid.String()+"/journal/summarize", strings.NewReader(body)))
				if rec.Code != http.StatusCreated || !strings.Contains(rec.Body.String(), "A faithful fixture summary.") {
					t.Fatalf("summary = %d %s", rec.Code, rec.Body.String())
				}
			}
			recChapter := httptest.NewRecorder()
			router.ServeHTTP(recChapter, httptest.NewRequest(http.MethodPost, "/api/v1/campaigns/"+cid.String()+"/journal/chapters", strings.NewReader(`{"request_id":"`+uuid.NewString()+`"}`)))
			if recChapter.Code != http.StatusOK || !strings.Contains(recChapter.Body.String(), "A faithful fixture summary.") {
				t.Fatalf("chapter = %d %s", recChapter.Code, recChapter.Body.String())
			}
			mu.Lock()
			summaryCalls := append([]string(nil), calls...)
			calls = nil
			mu.Unlock()
			want := "default"
			if explicit {
				want = "summary"
			}
			if len(summaryCalls) != 3 {
				t.Fatalf("summary calls=%v", summaryCalls)
			}
			for _, model := range summaryCalls {
				if model != want {
					t.Errorf("summary model=%s want %s", model, want)
				}
			}
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/v1/campaigns/start/world", strings.NewReader(`{"name":"Fixture world","profile":{"genre":"fantasy"},"character_profile":{"name":"Ada"}}`)))
			if rec.Code != http.StatusInternalServerError {
				t.Fatalf("expected invalid skeleton rejection, got %d %s", rec.Code, rec.Body.String())
			}
			mu.Lock()
			worldCalls := append([]string(nil), calls...)
			mu.Unlock()
			want = "default"
			if explicit {
				want = "world"
			}
			if len(worldCalls) == 0 {
				t.Fatal("world provider never called")
			}
			for _, model := range worldCalls {
				if model != want {
					t.Errorf("world model=%s want %s", model, want)
				}
			}
		})
	}
}
