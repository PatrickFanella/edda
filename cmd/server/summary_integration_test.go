//go:build integration

package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/charmbracelet/log"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"

	"git.subcult.tv/subculture-collective/edda/internal/config"
	"git.subcult.tv/subculture-collective/edda/internal/journal"
	"git.subcult.tv/subculture-collective/edda/internal/llm"
	serverroutes "git.subcult.tv/subculture-collective/edda/internal/server"
	statedb "git.subcult.tv/subculture-collective/edda/internal/state/sqlc"
)

// Exercise the exact API engine constructor, routed providers, HTTP action
// handler and durable summary store without external inference or app data.
func TestIntegrationAPIAutomaticSummaries(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	container, err := tcpostgres.Run(ctx, "docker.io/pgvector/pgvector:pg16", tcpostgres.WithDatabase("summaries"), tcpostgres.WithUsername("edda"), tcpostgres.WithPassword("fixture-only"), tcpostgres.BasicWaitStrategies())
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

	for _, mode := range []string{"fallback", "routed", "failure", "shutdown"} {
		t.Run(mode, func(t *testing.T) {
			appCtx, stop := context.WithCancel(ctx)
			defer stop()
			var calls atomic.Int32
			summaryStarted := make(chan struct{}, 2)
			summaryStopped := make(chan struct{}, 2)
			modelServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method == http.MethodGet {
					w.Header().Set("Content-Type", "application/json")
					if _, err := w.Write([]byte(`{"models":[]}`)); err != nil {
						t.Error(err)
					}
					return
				}
				var req struct {
					Model    string `json:"model"`
					Messages []struct {
						Content string `json:"content"`
					} `json:"messages"`
				}
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					t.Error(err)
					return
				}
				isSummary := len(req.Messages) > 0 && strings.Contains(req.Messages[0].Content, "chronicle writer")
				if isSummary {
					want := "summary"
					if mode == "fallback" {
						want = "gm"
					}
					if req.Model != want {
						t.Errorf("summary model = %q, want %q", req.Model, want)
					}
					calls.Add(1)
					summaryStarted <- struct{}{}
					if mode == "failure" {
						http.Error(w, "fixture summary failure", 500)
						return
					}
					if mode == "shutdown" {
						<-r.Context().Done()
						summaryStopped <- struct{}{}
						return
					}
				} else if req.Model != "gm" {
					t.Errorf("turn model = %q", req.Model)
				}
				text := "The gate is closed.\nChoices:\n1. Inspect the gate.\n2. Wait."
				if isSummary {
					text = "The player inspected the closed gate."
				}
				w.Header().Set("Content-Type", "application/json")
				if err := json.NewEncoder(w).Encode(map[string]any{"message": map[string]string{"role": "assistant", "content": text}, "done": true}); err != nil {
					t.Error(err)
				}
			}))
			defer modelServer.Close()
			cfg := config.Config{LLM: config.LLMConfig{Provider: "ollama", Ollama: config.OllamaConfig{Endpoint: modelServer.URL, Model: "gm"}, MemoryEmbedding: &config.MemoryEmbeddingConfig{Enabled: false}}}
			if mode != "fallback" {
				cfg.LLM.Profiles = map[string]config.LLMProfileConfig{"summary": {Provider: "ollama", Model: "summary"}}
				cfg.LLM.Routes.Summary = "summary"
			}
			roster, err := llm.NewRoster(cfg)
			if err != nil {
				t.Fatal(err)
			}
			gameEngine, err := newAPIEngine(appCtx, pool, cfg, roster, nil)
			if err != nil {
				t.Fatal(err)
			}
			q := statedb.New(pool)
			user, err := q.CreateUser(ctx, "summary-"+mode)
			if err != nil {
				t.Fatal(err)
			}
			campaign, err := q.CreateCampaign(ctx, statedb.CreateCampaignParams{Name: "Summary fixture", Status: "active", CreatedBy: user.ID})
			if err != nil {
				t.Fatal(err)
			}
			cid := uuid.UUID(campaign.ID.Bytes)
			router := serverroutes.NewRouterWithProviders(log.New(io.Discard), gameEngine, q, roster.Provider(llm.FlowDefault), roster.Provider(llm.FlowWorldgen), roster.Provider(llm.FlowSummary), pool, uuid.UUID(user.ID.Bytes), cfg, nil)
			for turn := 1; turn <= 21; turn++ {
				reqCtx, reqCancel := context.WithTimeout(ctx, 5*time.Second)
				req := httptest.NewRequest(http.MethodPost, "/api/v1/campaigns/"+cid.String()+"/action", strings.NewReader(`{"input":"Inspect the gate"}`)).WithContext(reqCtx)
				rec := httptest.NewRecorder()
				router.ServeHTTP(rec, req)
				reqCancel() // Automatic summaries must outlive the completed request.
				if rec.Code != http.StatusOK {
					t.Fatalf("turn %d = %d %s", turn, rec.Code, rec.Body.String())
				}
				if turn == 9 && calls.Load() != 0 {
					t.Fatalf("early summary calls = %d", calls.Load())
				}
				if turn == 10 || turn == 20 {
					select {
					case <-summaryStarted:
					case <-time.After(5 * time.Second):
						t.Fatal("summary not triggered")
					}
				}
			}
			if got := calls.Load(); got != 2 {
				t.Fatalf("summary calls = %d, want 2", got)
			}
			store := journal.NewStore(pool)
			if mode == "fallback" || mode == "routed" {
				deadline := time.Now().Add(5 * time.Second)
				for {
					summaries, err := store.ListSummaries(ctx, cid)
					if err != nil {
						t.Fatal(err)
					}
					if len(summaries) == 2 {
						ranges := map[int]int{}
						for _, s := range summaries {
							ranges[s.FromTurn] = s.ToTurn
						}
						if ranges[1] != 10 || ranges[11] != 20 {
							t.Fatalf("summary ranges = %v", ranges)
						}
						break
					}
					if time.Now().After(deadline) {
						t.Fatalf("durable summaries = %d, want 2", len(summaries))
					}
					time.Sleep(10 * time.Millisecond)
				}
			} else {
				stop()
				if mode == "shutdown" {
					for i := 0; i < 2; i++ {
						select {
						case <-summaryStopped:
						case <-time.After(5 * time.Second):
							t.Fatal("summary survived shutdown")
						}
					}
				}
				summaries, err := store.ListSummaries(ctx, cid)
				if err != nil {
					t.Fatal(err)
				}
				if len(summaries) != 0 {
					t.Fatalf("failed summaries persisted: %v", summaries)
				}
			}
		})
	}
}
