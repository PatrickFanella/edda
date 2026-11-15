//go:build integration

package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"git.subcult.tv/subculture-collective/edda/internal/auth"
	"git.subcult.tv/subculture-collective/edda/internal/dbutil"
	"git.subcult.tv/subculture-collective/edda/internal/handlers"
	"git.subcult.tv/subculture-collective/edda/internal/llm"
	"git.subcult.tv/subculture-collective/edda/internal/rules"
	statedb "git.subcult.tv/subculture-collective/edda/internal/state/sqlc"
	"git.subcult.tv/subculture-collective/edda/internal/world"
	"git.subcult.tv/subculture-collective/edda/pkg/api"
	"github.com/charmbracelet/log"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

const creationSkeleton = `{"factions":[{"name":"Guild"}],"locations":[{"name":"Gate","location_type":"city"}],"npcs":[{"name":"Guard","faction":"Guild","location":"Gate"}],"world_facts":[{"fact":"The roads are old.","category":"history"}],"starting_location":"Gate"}`
const creationScene = `{"narrative":"You reach the gate.","choices":["Enter","Wait"]}`

type creationProvider struct {
	llm.Provider
	t      *testing.T
	calls  int
	failAt int
	cancel context.CancelFunc
}

func (p *creationProvider) Complete(ctx context.Context, _ []llm.Message, _ []llm.Tool) (*llm.Response, error) {
	p.t.Helper()
	// Model calls must finish before any connection is held by creation.
	if n := testPool.Stat().AcquiredConns(); n != 0 {
		p.t.Fatalf("model called while %d database connections were held", n)
	}
	p.calls++
	if p.calls == p.failAt {
		return nil, errors.New("injected model failure")
	}
	if p.cancel != nil && p.calls == 2 {
		p.cancel()
		return nil, ctx.Err()
	}
	if p.calls == 1 {
		return &llm.Response{Content: creationSkeleton}, nil
	}
	return &llm.Response{Content: creationScene}, nil
}

func creationInput(userID uuid.UUID) world.OrchestratorInput {
	return world.OrchestratorInput{
		Name: "Atomic creation", UserID: userID,
		Profile:          &world.CampaignProfile{Genre: "fantasy", Tone: "hopeful", Themes: []string{}},
		CharacterProfile: &world.CharacterProfile{Name: "Scout", Concept: "ranger"},
		SpawnPackage: &world.CharacterSpawnPackage{
			Items:         []world.StarterItem{{Name: "Lantern", ItemType: "misc", Quantity: 1}},
			KnownFacts:    []world.StarterKnownFact{{Fact: "I know this road.", Category: "history"}},
			Relationships: []world.StarterRelationship{{TargetEntityType: "npc", TargetEntityID: uuid.New(), RelationshipType: "knows"}},
		},
	}
}

func assertNoCreationRows(t *testing.T, userID uuid.UUID) {
	t.Helper()
	var count int
	if err := testPool.QueryRow(context.Background(), "SELECT count(*) FROM campaigns WHERE created_by = $1", userID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("failed creation left %d campaigns", count)
	}
	for _, table := range []string{"factions", "locations", "npcs", "world_facts", "player_characters", "items", "entity_relationships", "session_logs"} {
		// Every owned row must still have an existing campaign, even after rollback.
		sql := fmt.Sprintf("SELECT count(*) FROM %s c LEFT JOIN campaigns p ON p.id = c.campaign_id WHERE p.id IS NULL", table)
		if err := testPool.QueryRow(context.Background(), sql).Scan(&count); err != nil {
			t.Fatal(err)
		}
		if count != 0 {
			t.Fatalf("%s has %d orphan rows", table, count)
		}
	}
}

func TestIntegrationWorldCreationRollbackAndRetry(t *testing.T) {
	ctx := context.Background()
	q := statedb.New(testPool)
	user, err := q.CreateUser(ctx, "creation-"+uuid.NewString())
	if err != nil {
		t.Fatal(err)
	}
	userID := dbutil.FromPgtype(user.ID)
	input := creationInput(userID)
	// Each write boundary fails separately, including the second world-fact
	// write for the spawn package and the relationship awareness update.
	stages := []struct {
		name, table, event, when string
		deferred                 bool
	}{
		{name: "campaign", table: "campaigns", event: "INSERT"},
		{name: "faction", table: "factions", event: "INSERT"},
		{name: "location", table: "locations", event: "INSERT"},
		{name: "npc", table: "npcs", event: "INSERT"},
		{name: "world fact", table: "world_facts", event: "INSERT", when: "WHEN (NEW.source = 'world_generation')"},
		{name: "character", table: "player_characters", event: "INSERT"},
		{name: "item", table: "items", event: "INSERT"},
		{name: "spawn fact", table: "world_facts", event: "INSERT", when: "WHEN (NEW.source = 'character_spawn')"},
		{name: "relationship", table: "entity_relationships", event: "INSERT"},
		{name: "relationship awareness", table: "entity_relationships", event: "UPDATE"},
		{name: "scene", table: "session_logs", event: "INSERT"},
		{name: "commit", table: "session_logs", event: "INSERT", deferred: true},
	}
	_, err = testPool.Exec(ctx, `CREATE FUNCTION fail_creation_write() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'injected creation failure'; END $$`)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if _, err := testPool.Exec(ctx, "DROP FUNCTION fail_creation_write() CASCADE"); err != nil {
			t.Error(err)
		}
	}()
	for _, stage := range stages {
		t.Run(stage.name, func(t *testing.T) {
			timing, kind, deferral := "BEFORE", "", ""
			if stage.deferred {
				timing, kind, deferral = "AFTER", "CONSTRAINT ", "DEFERRABLE INITIALLY DEFERRED"
			}
			sql := fmt.Sprintf("CREATE %sTRIGGER fail_creation %s %s ON %s %s FOR EACH ROW %s EXECUTE FUNCTION fail_creation_write()", kind, timing, stage.event, stage.table, deferral, stage.when)
			if _, err := testPool.Exec(ctx, sql); err != nil {
				t.Fatal(err)
			}
			result, err := world.NewOrchestrator(&creationProvider{t: t}, q).Run(ctx, input, nil)
			if _, dropErr := testPool.Exec(ctx, "DROP TRIGGER fail_creation ON "+stage.table); dropErr != nil {
				t.Fatal(dropErr)
			}
			if err == nil || result != nil {
				t.Fatalf("expected failed creation, got result=%v err=%v", result, err)
			}
			assertNoCreationRows(t, userID)
			// Retry from the same wizard input creates exactly one complete aggregate.
			result, err = world.NewOrchestrator(&creationProvider{t: t}, q).Run(ctx, input, nil)
			if err != nil {
				t.Fatal(err)
			}
			campaigns, err := q.ListCampaignsByUser(ctx, user.ID)
			if err != nil || len(campaigns) != 1 {
				t.Fatalf("retry campaigns=%d err=%v", len(campaigns), err)
			}
			for table, want := range map[string]int{"factions": 1, "locations": 1, "npcs": 1, "world_facts": 2, "player_characters": 1, "items": 1, "entity_relationships": 1, "session_logs": 1} {
				var count int
				if err := testPool.QueryRow(ctx, "SELECT count(*) FROM "+table+" WHERE campaign_id = $1", result.Campaign.ID).Scan(&count); err != nil {
					t.Fatal(err)
				}
				if count != want {
					t.Fatalf("%s count=%d want=%d", table, count, want)
				}
			}
			if result.Campaign.Status != "active" {
				t.Fatalf("status=%s", result.Campaign.Status)
			}
			if err := q.DeleteCampaign(ctx, result.Campaign.ID); err != nil {
				t.Fatal(err)
			}
		})
	}
	for _, failAt := range []int{1, 2} {
		t.Run(fmt.Sprintf("model failure %d", failAt), func(t *testing.T) {
			result, err := world.NewOrchestrator(&creationProvider{t: t, failAt: failAt}, q).Run(ctx, input, nil)
			if err == nil || result != nil {
				t.Fatalf("result=%v err=%v", result, err)
			}
			assertNoCreationRows(t, userID)
		})
	}
	t.Run("cancelled generation", func(t *testing.T) {
		cancelledCtx, cancel := context.WithCancel(ctx)
		defer cancel()
		result, err := world.NewOrchestrator(&creationProvider{t: t, cancel: cancel}, q).Run(cancelledCtx, input, nil)
		if !errors.Is(err, context.Canceled) || result != nil {
			t.Fatalf("result=%v err=%v", result, err)
		}
		assertNoCreationRows(t, userID)
	})
	t.Run("cancelled persistence", func(t *testing.T) {
		cancelledCtx, cancel := context.WithCancel(ctx)
		defer cancel()
		result, err := world.NewOrchestrator(&creationProvider{t: t}, q).Run(cancelledCtx, input, func(stage string) {
			if stage == "Bringing your character to life…" {
				cancel()
			}
		})
		if err == nil || result != nil {
			t.Fatalf("result=%v err=%v", result, err)
		}
		assertNoCreationRows(t, userID)
	})
}

func TestIntegrationWorldCreationRequiredSetup(t *testing.T) {
	ctx := context.Background()
	q := statedb.New(testPool)
	user, err := q.CreateUser(ctx, "setup-"+uuid.NewString())
	if err != nil {
		t.Fatal(err)
	}
	userID := dbutil.FromPgtype(user.ID)
	input := creationInput(userID)
	request := func(t *testing.T, mode string) *httptest.ResponseRecorder {
		t.Helper()
		// Required setup uses the transaction's database, even without a separate pool.
		h := handlers.NewStartupHandlers(&creationProvider{t: t}, q, log.Default(), nil)
		router := chi.NewRouter()
		router.Use(auth.NewNoOpMiddleware(userID).Authenticate)
		router.Post("/world", h.BuildWorld)
		body, err := json.Marshal(api.WorldBuildRequest{Name: input.Name, Profile: input.Profile, CharacterProfile: input.CharacterProfile, RulesMode: mode})
		if err != nil {
			t.Fatal(err)
		}
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/world", bytes.NewReader(body)))
		return rec
	}
	_, err = testPool.Exec(ctx, `CREATE FUNCTION fail_required_setup() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'injected required setup failure'; END $$`)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if _, err := testPool.Exec(ctx, "DROP FUNCTION fail_required_setup() CASCADE"); err != nil {
			t.Error(err)
		}
	}()
	for _, stage := range []struct{ name, table, event string }{
		{"clock", "campaign_time", "INSERT"},
		{"rules", "campaigns", "UPDATE"},
		{"feats", "feat_definitions", "INSERT"},
		{"skills", "skill_definitions", "INSERT"},
	} {
		t.Run(stage.name, func(t *testing.T) {
			sql := fmt.Sprintf("CREATE TRIGGER fail_setup BEFORE %s ON %s FOR EACH ROW EXECUTE FUNCTION fail_required_setup()", stage.event, stage.table)
			if _, err := testPool.Exec(ctx, sql); err != nil {
				t.Fatal(err)
			}
			rec := request(t, "crunch")
			if _, err := testPool.Exec(ctx, "DROP TRIGGER fail_setup ON "+stage.table); err != nil {
				t.Fatal(err)
			}
			if rec.Code != http.StatusInternalServerError {
				t.Fatalf("got %d: %s", rec.Code, rec.Body.String())
			}
			if strings.Contains(rec.Body.String(), `"campaign"`) {
				t.Fatalf("failed setup returned a campaign: %s", rec.Body.String())
			}
			assertNoCreationRows(t, userID)
			for _, table := range []string{"campaign_time", "feat_definitions", "skill_definitions"} {
				var count int
				if err := testPool.QueryRow(ctx, "SELECT count(*) FROM "+table+" c LEFT JOIN campaigns p ON p.id=c.campaign_id WHERE p.id IS NULL").Scan(&count); err != nil {
					t.Fatal(err)
				}
				if count != 0 {
					t.Fatalf("%s has %d orphan rows", table, count)
				}
			}
		})
	}
	for _, mode := range []string{"", "narrative", "crunch"} {
		t.Run("success "+mode, func(t *testing.T) {
			rec := request(t, mode)
			if rec.Code != http.StatusOK {
				t.Fatalf("got %d: %s", rec.Code, rec.Body.String())
			}
			var resp api.WorldBuildResponse
			if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
				t.Fatal(err)
			}
			id, err := uuid.Parse(resp.Campaign.ID)
			if err != nil {
				t.Fatal(err)
			}
			saved, err := q.GetCampaignByID(ctx, dbutil.ToPgtype(id))
			if err != nil {
				t.Fatal(err)
			}
			want := mode
			if want == "" {
				want = "narrative"
			}
			if resp.Campaign.RulesMode != want || saved.RulesMode != want {
				t.Fatalf("response mode=%s saved=%s want=%s", resp.Campaign.RulesMode, saved.RulesMode, want)
			}
			var day, hour, minute int
			if err := testPool.QueryRow(ctx, "SELECT day,hour,minute FROM campaign_time WHERE campaign_id=$1", id).Scan(&day, &hour, &minute); err != nil {
				t.Fatal(err)
			}
			if day != 1 || hour != 8 || minute != 0 {
				t.Fatalf("clock=%d/%d:%d", day, hour, minute)
			}
			for table, wantCount := range map[string]int{"player_characters": 1, "session_logs": 1, "feat_definitions": 0, "skill_definitions": 0} {
				if want == "crunch" && table == "feat_definitions" {
					wantCount = len(rules.DefaultFeats)
				}
				if want == "crunch" && table == "skill_definitions" {
					wantCount = len(rules.DefaultSkills)
				}
				var count int
				if err := testPool.QueryRow(ctx, "SELECT count(*) FROM "+table+" WHERE campaign_id=$1", id).Scan(&count); err != nil {
					t.Fatal(err)
				}
				if count != wantCount {
					t.Fatalf("%s count=%d want=%d", table, count, wantCount)
				}
			}
			if err := q.DeleteCampaign(ctx, saved.ID); err != nil {
				t.Fatal(err)
			}
		})
	}
}
